package session

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/EmanuelSena101/DDivination/apps/server/internal/domain"
	"github.com/EmanuelSena101/DDivination/apps/server/internal/store"
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrInvalidCode     = errors.New("invalid or expired join code")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrRevision        = errors.New("session revision conflict")
	ErrInvalidCommand  = errors.New("invalid session command")
)

type Snapshot struct {
	Type      string                   `json:"type"`
	State     domain.SessionState      `json:"state"`
	Adventure domain.AdventureDocument `json:"adventure"`
}

type Created struct {
	SessionID string                   `json:"sessionId"`
	Code      string                   `json:"code"`
	ExpiresAt time.Time                `json:"expiresAt"`
	Token     string                   `json:"token"`
	State     domain.SessionState      `json:"state"`
	Adventure domain.AdventureDocument `json:"adventure"`
}

type Joined struct {
	SessionID     string                   `json:"sessionId"`
	ParticipantID string                   `json:"participantId"`
	Token         string                   `json:"token"`
	State         domain.SessionState      `json:"state"`
	Adventure     domain.AdventureDocument `json:"adventure"`
}

type Subscriber struct {
	Participant domain.Participant
	Events      chan domain.SessionEvent
}

type persistence interface {
	SaveSessionEvent(context.Context, string, domain.SessionEvent) error
	SaveSessionSnapshot(context.Context, domain.SessionState) error
	LoadSessionSnapshot(context.Context, string) (domain.SessionState, error)
	SaveSessionCredential(context.Context, string, string, string) error
	LoadSessionCredentials(context.Context, string) (map[string]string, error)
	GetAdventure(context.Context, string) (domain.AdventureDocument, error)
}

type liveSession struct {
	mu          sync.RWMutex
	state       domain.SessionState
	adventure   domain.AdventureDocument
	code        string
	codeExpires time.Time
	byToken     map[string]string
	subscribers map[*Subscriber]struct{}
}

type Hub struct {
	mu       sync.RWMutex
	sessions map[string]*liveSession
	store    persistence
}

func NewHub(s *store.Store) *Hub {
	return &Hub{sessions: make(map[string]*liveSession), store: s}
}

func (h *Hub) Create(ctx context.Context, adventure domain.AdventureDocument, gmName string) (Created, error) {
	sessionID, err := randomToken(9)
	if err != nil {
		return Created{}, err
	}
	gmToken, err := randomToken(24)
	if err != nil {
		return Created{}, err
	}
	codeValue, err := randomDigits(6)
	if err != nil {
		return Created{}, err
	}
	now := time.Now().UTC()
	gm := domain.Participant{ID: "gm", Name: cleanName(gmName, "Game Master"), Role: "gm", Token: gmToken, JoinedAt: now}
	state := domain.SessionState{
		ID:             sessionID,
		AdventureID:    adventure.ID,
		Revision:       0,
		ActiveFloorID:  adventure.Floors[0].ID,
		Participants:   map[string]domain.Participant{gm.ID: gm},
		TokenPositions: make(map[string]domain.GridPosition),
		TokenFloors:    make(map[string]string),
		TokenOwners:    make(map[string]string),
		RevealedCells:  make(map[string][]domain.GridPosition),
		Initiative:     domain.InitiativeState{Entries: make([]domain.InitiativeEntry, 0), Round: 1},
		Rolls:          make([]domain.DiceRoll, 0),
		Open:           true,
		CreatedAt:      now,
	}
	for _, floor := range adventure.Floors {
		for _, entity := range floor.Entities {
			if entity.Kind == "token" || entity.Kind == "boss" {
				state.TokenPositions[entity.ID] = entity.Position
				state.TokenFloors[entity.ID] = floor.ID
			}
		}
	}
	if entrance, ok := state.TokenPositions["token-party"]; ok {
		floorID := state.TokenFloors["token-party"]
		for x := entrance.X - 2; x <= entrance.X+2; x++ {
			for z := entrance.Z - 2; z <= entrance.Z+2; z++ {
				if walkable(adventure, floorID, x, z) {
					state.RevealedCells[floorID] = append(
						state.RevealedCells[floorID],
						domain.GridPosition{X: x, Z: z},
					)
				}
			}
		}
	}
	if _, ok := state.TokenPositions["token-party"]; ok {
		state.TokenOwners["token-party"] = gm.ID
	}
	s := &liveSession{
		state:       state,
		adventure:   adventure,
		code:        codeValue,
		codeExpires: now.Add(15 * time.Minute),
		byToken:     map[string]string{gmToken: gm.ID},
		subscribers: make(map[*Subscriber]struct{}),
	}
	h.mu.Lock()
	h.sessions[sessionID] = s
	h.mu.Unlock()
	if err := h.store.SaveSessionSnapshot(ctx, state); err != nil {
		return Created{}, err
	}
	if err := h.store.SaveSessionCredential(ctx, sessionID, gm.ID, gmToken); err != nil {
		return Created{}, err
	}
	return Created{
		SessionID: sessionID,
		Code:      codeValue,
		ExpiresAt: s.codeExpires,
		Token:     gmToken,
		State:     state,
		Adventure: adventure,
	}, nil
}

func (h *Hub) Join(ctx context.Context, sessionID, codeValue, name, role string) (Joined, error) {
	s, err := h.find(sessionID)
	if err != nil {
		return Joined{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.state.Open || time.Now().UTC().After(s.codeExpires) || !sameCode(s.code, codeValue) {
		return Joined{}, ErrInvalidCode
	}
	if role != "player" && role != "display" {
		role = "player"
	}
	token, err := randomToken(24)
	if err != nil {
		return Joined{}, err
	}
	participantID, err := randomToken(6)
	if err != nil {
		return Joined{}, err
	}
	participant := domain.Participant{
		ID:       participantID,
		Name:     cleanName(name, "Player"),
		Role:     role,
		Token:    token,
		JoinedAt: time.Now().UTC(),
	}
	s.state.Participants[participant.ID] = participant
	s.byToken[token] = participant.ID
	if err := h.store.SaveSessionCredential(ctx, sessionID, participant.ID, token); err != nil {
		return Joined{}, err
	}
	if role == "player" {
		if owner, ok := s.state.TokenOwners["token-party"]; !ok || owner == "gm" {
			s.state.TokenOwners["token-party"] = participant.ID
		}
	}
	s.state.Revision++
	event := domain.SessionEvent{
		Revision:   s.state.Revision,
		Type:       "participant.joined",
		ActorID:    participant.ID,
		OccurredAt: time.Now().UTC(),
		Payload: map[string]any{
			"participant": map[string]any{
				"id":       participant.ID,
				"name":     participant.Name,
				"role":     participant.Role,
				"joinedAt": participant.JoinedAt,
			},
		},
	}
	if err := h.store.SaveSessionEvent(ctx, sessionID, event); err != nil {
		return Joined{}, err
	}
	if err := h.store.SaveSessionSnapshot(ctx, s.state); err != nil {
		return Joined{}, err
	}
	for sub := range s.subscribers {
		select {
		case sub.Events <- event:
		default:
		}
	}
	state, adventure := filteredSnapshot(s.state, s.adventure, participant)
	return Joined{SessionID: sessionID, ParticipantID: participantID, Token: token, State: state, Adventure: adventure}, nil
}

func (h *Hub) Subscribe(sessionID, token string) (*Subscriber, Snapshot, error) {
	s, err := h.find(sessionID)
	if err != nil {
		return nil, Snapshot{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	participant, ok := participantByToken(s, token)
	if !ok {
		return nil, Snapshot{}, ErrUnauthorized
	}
	sub := &Subscriber{Participant: participant, Events: make(chan domain.SessionEvent, 128)}
	s.subscribers[sub] = struct{}{}
	state, adventure := filteredSnapshot(s.state, s.adventure, participant)
	return sub, Snapshot{Type: "session.snapshot", State: state, Adventure: adventure}, nil
}

func (h *Hub) Unsubscribe(sessionID string, sub *Subscriber) {
	s, err := h.find(sessionID)
	if err != nil {
		return
	}
	s.mu.Lock()
	delete(s.subscribers, sub)
	close(sub.Events)
	s.mu.Unlock()
}

func (h *Hub) HandleCommand(ctx context.Context, sessionID, token string, command domain.SessionCommand) (domain.SessionEvent, error) {
	s, err := h.find(sessionID)
	if err != nil {
		return domain.SessionEvent{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	participant, ok := participantByToken(s, token)
	if !ok {
		return domain.SessionEvent{}, ErrUnauthorized
	}
	if command.ExpectedRevision != s.state.Revision {
		return domain.SessionEvent{}, ErrRevision
	}
	payload, err := applyCommand(&s.state, s.adventure, participant, command)
	if err != nil {
		return domain.SessionEvent{}, err
	}
	s.state.Revision++
	event := domain.SessionEvent{
		Revision:   s.state.Revision,
		Type:       eventType(command.Type),
		ActorID:    participant.ID,
		OccurredAt: time.Now().UTC(),
		Payload:    payload,
	}
	if err := h.store.SaveSessionEvent(ctx, sessionID, event); err != nil {
		return domain.SessionEvent{}, err
	}
	if err := h.store.SaveSessionSnapshot(ctx, s.state); err != nil {
		return domain.SessionEvent{}, err
	}
	for sub := range s.subscribers {
		if canReceive(sub.Participant, event, s.adventure) {
			select {
			case sub.Events <- event:
			default:
				// Slow clients will reconnect and receive a fresh snapshot.
			}
		}
	}
	return event, nil
}

func (h *Hub) Close(ctx context.Context, sessionID, token string) error {
	s, err := h.find(sessionID)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	participant, ok := participantByToken(s, token)
	if !ok || participant.Role != "gm" {
		return ErrUnauthorized
	}
	s.state.Open = false
	s.state.Revision++
	event := domain.SessionEvent{
		Revision:   s.state.Revision,
		Type:       "session.closed",
		ActorID:    participant.ID,
		OccurredAt: time.Now().UTC(),
		Payload:    map[string]any{},
	}
	if err := h.store.SaveSessionEvent(ctx, sessionID, event); err != nil {
		return err
	}
	if err := h.store.SaveSessionSnapshot(ctx, s.state); err != nil {
		return err
	}
	for sub := range s.subscribers {
		select {
		case sub.Events <- event:
		default:
		}
	}
	return nil
}

func (h *Hub) find(id string) (*liveSession, error) {
	h.mu.RLock()
	s := h.sessions[id]
	h.mu.RUnlock()
	if s != nil {
		return s, nil
	}
	ctx := context.Background()
	state, err := h.store.LoadSessionSnapshot(ctx, id)
	if err != nil {
		return nil, ErrSessionNotFound
	}
	adventure, err := h.store.GetAdventure(ctx, state.AdventureID)
	if err != nil {
		return nil, ErrSessionNotFound
	}
	credentials, err := h.store.LoadSessionCredentials(ctx, id)
	if err != nil {
		return nil, ErrSessionNotFound
	}
	restored := &liveSession{
		state:       state,
		adventure:   adventure,
		codeExpires: time.Time{},
		byToken:     credentials,
		subscribers: make(map[*Subscriber]struct{}),
	}
	h.mu.Lock()
	if existing := h.sessions[id]; existing != nil {
		restored = existing
	} else {
		h.sessions[id] = restored
	}
	h.mu.Unlock()
	return restored, nil
}

func participantByToken(s *liveSession, token string) (domain.Participant, bool) {
	id, ok := s.byToken[token]
	if !ok {
		hash := sha256.Sum256([]byte(token))
		id, ok = s.byToken["sha256:"+hex.EncodeToString(hash[:])]
	}
	if !ok {
		return domain.Participant{}, false
	}
	participant, ok := s.state.Participants[id]
	return participant, ok
}

func applyCommand(state *domain.SessionState, adventure domain.AdventureDocument, participant domain.Participant, command domain.SessionCommand) (map[string]any, error) {
	switch command.Type {
	case "token.move":
		if participant.Role == "display" {
			return nil, ErrUnauthorized
		}
		tokenID, _ := command.Payload["tokenId"].(string)
		floorID, _ := command.Payload["floorId"].(string)
		x, okX := intValue(command.Payload["x"])
		z, okZ := intValue(command.Payload["z"])
		if tokenID == "" || floorID == "" || !okX || !okZ || !walkable(adventure, floorID, x, z) {
			return nil, ErrInvalidCommand
		}
		if participant.Role != "gm" && state.TokenOwners[tokenID] != participant.ID {
			return nil, ErrUnauthorized
		}
		if participant.Role != "gm" && !cellRevealed(state.RevealedCells[floorID], domain.GridPosition{X: x, Z: z}) {
			return nil, fmt.Errorf("%w: target cell is hidden by fog", ErrUnauthorized)
		}
		for otherID, position := range state.TokenPositions {
			if otherID != tokenID && state.TokenFloors[otherID] == floorID && position.X == x && position.Z == z {
				return nil, fmt.Errorf("%w: target cell occupied", ErrInvalidCommand)
			}
		}
		state.TokenPositions[tokenID] = domain.GridPosition{X: x, Z: z}
		state.TokenFloors[tokenID] = floorID
		return map[string]any{"tokenId": tokenID, "floorId": floorID, "x": x, "z": z}, nil

	case "fog.reveal", "fog.hide":
		if participant.Role != "gm" {
			return nil, ErrUnauthorized
		}
		floorID, _ := command.Payload["floorId"].(string)
		x, okX := intValue(command.Payload["x"])
		z, okZ := intValue(command.Payload["z"])
		if floorID == "" || !okX || !okZ {
			return nil, ErrInvalidCommand
		}
		position := domain.GridPosition{X: x, Z: z}
		reveal := command.Type == "fog.reveal"
		state.RevealedCells[floorID] = updateCell(state.RevealedCells[floorID], position, reveal)
		return map[string]any{"floorId": floorID, "x": x, "z": z, "revealed": reveal}, nil

	case "floor.set":
		if participant.Role != "gm" {
			return nil, ErrUnauthorized
		}
		floorID, _ := command.Payload["floorId"].(string)
		if !floorExists(adventure, floorID) {
			return nil, ErrInvalidCommand
		}
		state.ActiveFloorID = floorID
		return map[string]any{"floorId": floorID}, nil

	case "initiative.set":
		if participant.Role != "gm" {
			return nil, ErrUnauthorized
		}
		raw, err := json.Marshal(command.Payload["initiative"])
		if err != nil {
			return nil, ErrInvalidCommand
		}
		var initiative domain.InitiativeState
		if err := json.Unmarshal(raw, &initiative); err != nil || initiative.Round < 1 {
			return nil, ErrInvalidCommand
		}
		state.Initiative = initiative
		return map[string]any{"initiative": initiative}, nil

	case "dice.roll":
		expression, _ := command.Payload["expression"].(string)
		visibility, _ := command.Payload["visibility"].(string)
		targetID, _ := command.Payload["targetId"].(string)
		if visibility == "" {
			visibility = "public"
		}
		if visibility != "public" && visibility != "gm" && visibility != "private" {
			return nil, ErrInvalidCommand
		}
		roll, err := rollDice(expression)
		if err != nil {
			return nil, err
		}
		roll.ID = command.ID
		roll.ActorID = participant.ID
		roll.Visibility = visibility
		roll.TargetID = targetID
		roll.CreatedAt = time.Now().UTC()
		state.Rolls = append(state.Rolls, roll)
		if len(state.Rolls) > 100 {
			state.Rolls = state.Rolls[len(state.Rolls)-100:]
		}
		payload := map[string]any{}
		encoded, _ := json.Marshal(roll)
		_ = json.Unmarshal(encoded, &payload)
		return payload, nil

	case "ping":
		floorID, _ := command.Payload["floorId"].(string)
		x, okX := intValue(command.Payload["x"])
		z, okZ := intValue(command.Payload["z"])
		if floorID == "" || !okX || !okZ {
			return nil, ErrInvalidCommand
		}
		return map[string]any{"floorId": floorID, "x": x, "z": z}, nil
	default:
		return nil, ErrInvalidCommand
	}
}

func filteredSnapshot(state domain.SessionState, adventure domain.AdventureDocument, participant domain.Participant) (domain.SessionState, domain.AdventureDocument) {
	if participant.Role == "gm" {
		return state, adventure
	}
	encodedState, _ := json.Marshal(state)
	var cleanState domain.SessionState
	_ = json.Unmarshal(encodedState, &cleanState)
	for id, p := range cleanState.Participants {
		p.Token = ""
		cleanState.Participants[id] = p
	}
	cleanState.Rolls = filterRolls(cleanState.Rolls, participant)

	encodedAdventure, _ := json.Marshal(adventure)
	var cleanAdventure domain.AdventureDocument
	_ = json.Unmarshal(encodedAdventure, &cleanAdventure)
	for fi := range cleanAdventure.Floors {
		floor := &cleanAdventure.Floors[fi]
		secretRooms := make(map[string]bool)
		visibleRooms := floor.Rooms[:0]
		for _, room := range floor.Rooms {
			if room.Secret {
				secretRooms[room.ID] = true
				continue
			}
			visibleRooms = append(visibleRooms, room)
		}
		floor.Rooms = visibleRooms
		visibleTiles := floor.Tiles[:0]
		for _, tile := range floor.Tiles {
			if !secretRooms[tile.RoomID] {
				visibleTiles = append(visibleTiles, tile)
			}
		}
		floor.Tiles = visibleTiles
		visibleEntities := floor.Entities[:0]
		for _, entity := range floor.Entities {
			if !entity.Hidden && !secretRooms[entity.RoomID] {
				visibleEntities = append(visibleEntities, entity)
			}
		}
		floor.Entities = visibleEntities
	}
	cleanAdventure.Analysis.TotalRooms = 0
	for _, floor := range cleanAdventure.Floors {
		cleanAdventure.Analysis.TotalRooms += len(floor.Rooms)
	}
	cleanAdventure.Analysis.DeadEnds = nil
	cleanAdventure.Progression.SecretRoomIDs = nil
	visibleInvariants := cleanAdventure.Analysis.Invariants[:0]
	for _, invariant := range cleanAdventure.Analysis.Invariants {
		if invariant != "secrets-are-optional" {
			visibleInvariants = append(visibleInvariants, invariant)
		}
	}
	cleanAdventure.Analysis.Invariants = visibleInvariants
	return cleanState, cleanAdventure
}

func canReceive(participant domain.Participant, event domain.SessionEvent, adventure domain.AdventureDocument) bool {
	if participant.Role != "gm" && event.Type == "token.moved" {
		tokenID, _ := event.Payload["tokenId"].(string)
		for _, floor := range adventure.Floors {
			for _, entity := range floor.Entities {
				if entity.ID != tokenID {
					continue
				}
				if entity.Hidden || roomIsSecret(floor, entity.RoomID) {
					return false
				}
			}
		}
	}
	if event.Type != "dice.rolled" {
		return true
	}
	visibility, _ := event.Payload["visibility"].(string)
	targetID, _ := event.Payload["targetId"].(string)
	switch visibility {
	case "gm":
		return participant.Role == "gm" || participant.ID == event.ActorID
	case "private":
		return participant.Role == "gm" || participant.ID == event.ActorID || participant.ID == targetID
	default:
		return true
	}
}

func roomIsSecret(floor domain.FloorMap, roomID string) bool {
	for _, room := range floor.Rooms {
		if room.ID == roomID {
			return room.Secret
		}
	}
	return false
}

func cellRevealed(cells []domain.GridPosition, target domain.GridPosition) bool {
	for _, cell := range cells {
		if cell == target {
			return true
		}
	}
	return false
}

func filterRolls(rolls []domain.DiceRoll, participant domain.Participant) []domain.DiceRoll {
	result := make([]domain.DiceRoll, 0, len(rolls))
	for _, roll := range rolls {
		if roll.Visibility == "public" || participant.Role == "gm" || roll.ActorID == participant.ID || roll.TargetID == participant.ID {
			result = append(result, roll)
		}
	}
	return result
}

var diceExpression = regexp.MustCompile(`(?i)^(\d{1,2})d(4|6|8|10|12|20|100)([+-]\d{1,3})?$`)

func rollDice(expression string) (domain.DiceRoll, error) {
	expression = strings.ReplaceAll(strings.TrimSpace(expression), " ", "")
	match := diceExpression.FindStringSubmatch(expression)
	if match == nil {
		return domain.DiceRoll{}, fmt.Errorf("%w: unsupported dice expression", ErrInvalidCommand)
	}
	count, _ := strconv.Atoi(match[1])
	sides, _ := strconv.Atoi(match[2])
	modifier := 0
	if match[3] != "" {
		modifier, _ = strconv.Atoi(match[3])
	}
	if count < 1 || count > 20 {
		return domain.DiceRoll{}, fmt.Errorf("%w: dice count out of range", ErrInvalidCommand)
	}
	values := make([]int, count)
	total := modifier
	for i := range values {
		value, err := cryptoInt(sides)
		if err != nil {
			return domain.DiceRoll{}, err
		}
		values[i] = value
		total += value
	}
	return domain.DiceRoll{Expression: strings.ToLower(expression), Values: values, Modifier: modifier, Total: total}, nil
}

func eventType(commandType string) string {
	return map[string]string{
		"token.move":     "token.moved",
		"fog.reveal":     "fog.changed",
		"fog.hide":       "fog.changed",
		"floor.set":      "floor.changed",
		"initiative.set": "initiative.changed",
		"dice.roll":      "dice.rolled",
		"ping":           "map.pinged",
	}[commandType]
}

func walkable(adventure domain.AdventureDocument, floorID string, x, z int) bool {
	for _, floor := range adventure.Floors {
		if floor.ID != floorID {
			continue
		}
		for _, tile := range floor.Tiles {
			if tile.X == x && tile.Z == z {
				return tile.Walkable
			}
		}
	}
	return false
}

func floorExists(adventure domain.AdventureDocument, floorID string) bool {
	for _, floor := range adventure.Floors {
		if floor.ID == floorID {
			return true
		}
	}
	return false
}

func updateCell(cells []domain.GridPosition, target domain.GridPosition, reveal bool) []domain.GridPosition {
	result := make([]domain.GridPosition, 0, len(cells)+1)
	found := false
	for _, cell := range cells {
		if cell == target {
			found = true
			if !reveal {
				continue
			}
		}
		result = append(result, cell)
	}
	if reveal && !found {
		result = append(result, target)
	}
	return result
}

func intValue(value any) (int, bool) {
	switch v := value.(type) {
	case float64:
		return int(v), v == float64(int(v))
	case int:
		return v, true
	case json.Number:
		i, err := v.Int64()
		return int(i), err == nil
	default:
		return 0, false
	}
}

func randomToken(bytesCount int) (string, error) {
	buffer := make([]byte, bytesCount)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func randomDigits(count int) (string, error) {
	limit := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(count)), nil)
	value, err := rand.Int(rand.Reader, limit)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%0*d", count, value.Int64()), nil
}

func cryptoInt(sides int) (int, error) {
	value, err := rand.Int(rand.Reader, big.NewInt(int64(sides)))
	if err != nil {
		return 0, err
	}
	return int(value.Int64()) + 1, nil
}

func cleanName(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	if len(value) > 60 {
		return value[:60]
	}
	return value
}

func sameCode(expected, actual string) bool {
	expected = strings.TrimSpace(expected)
	actual = strings.TrimSpace(actual)
	return len(expected) == len(actual) &&
		subtle.ConstantTimeCompare([]byte(expected), []byte(actual)) == 1
}
