package session

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/EmanuelSena101/DDivination/apps/server/internal/domain"
	"github.com/EmanuelSena101/DDivination/apps/server/internal/generator"
	"github.com/EmanuelSena101/DDivination/apps/server/internal/testdb"
)

func TestDiceExpressions(t *testing.T) {
	for _, expression := range []string{"1d4", "2d6+3", "1d8-2", "3d10", "1d12", "1d20", "1d100"} {
		roll, err := rollDice(expression)
		if err != nil {
			t.Fatalf("%s: %v", expression, err)
		}
		if roll.Total < -2 {
			t.Fatalf("%s produced an impossible total: %d", expression, roll.Total)
		}
	}
	for _, expression := range []string{"d20", "21d6", "1d7", "1d20+1000", "attack"} {
		if _, err := rollDice(expression); err == nil {
			t.Fatalf("expected %q to be rejected", expression)
		}
	}
}

func TestPlayerCannotMoveUnassignedToken(t *testing.T) {
	state := domain.SessionState{
		Revision:       0,
		TokenOwners:    map[string]string{"token-party": "someone-else"},
		TokenPositions: map[string]domain.GridPosition{"token-party": {X: 1, Z: 1}},
		TokenFloors:    map[string]string{"token-party": "floor-1"},
		RevealedCells:  map[string][]domain.GridPosition{"floor-1": {{X: 2, Z: 1}}},
	}
	adventure := domain.AdventureDocument{Floors: []domain.FloorMap{{
		ID:    "floor-1",
		Tiles: []domain.Tile{{X: 2, Z: 1, Walkable: true}},
	}}}
	participant := domain.Participant{ID: "player-1", Role: "player"}
	command := domain.SessionCommand{
		Type: "token.move",
		Payload: map[string]any{
			"tokenId": "token-party",
			"floorId": "floor-1",
			"x":       2,
			"z":       1,
		},
	}
	if _, err := applyCommand(&state, adventure, participant, command); err != ErrUnauthorized {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestPlayerCannotMoveIntoFog(t *testing.T) {
	state := domain.SessionState{
		TokenOwners:    map[string]string{"token-party": "player-1"},
		TokenPositions: map[string]domain.GridPosition{"token-party": {X: 1, Z: 1}},
		TokenFloors:    map[string]string{"token-party": "floor-1"},
		RevealedCells:  map[string][]domain.GridPosition{"floor-1": {}},
	}
	adventure := domain.AdventureDocument{Floors: []domain.FloorMap{{
		ID:    "floor-1",
		Tiles: []domain.Tile{{X: 2, Z: 1, Walkable: true}},
	}}}
	participant := domain.Participant{ID: "player-1", Role: "player"}
	command := domain.SessionCommand{
		Type: "token.move",
		Payload: map[string]any{
			"tokenId": "token-party",
			"floorId": "floor-1",
			"x":       2,
			"z":       1,
		},
	}
	if _, err := applyCommand(&state, adventure, participant, command); err == nil {
		t.Fatal("expected movement into fog to be rejected")
	}
}

func TestJoinCodeComparison(t *testing.T) {
	if !sameCode("123456", " 123456 ") {
		t.Fatal("trimmed matching codes should compare equal")
	}
	if sameCode("123456", "123457") || sameCode("123456", "12345") {
		t.Fatal("different join codes compared equal")
	}
}

func TestPlayerSnapshotDoesNotRevealSecretMetadata(t *testing.T) {
	doc, err := generator.Generate(domain.DefaultAdventureSpec(), 123, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	var secretFloorID string
	var secretCell domain.GridPosition
	for _, floor := range doc.Floors {
		for _, tile := range floor.Tiles {
			if roomIsSecret(floor, tile.RoomID) {
				secretFloorID = floor.ID
				secretCell = domain.GridPosition{X: tile.X, Z: tile.Z}
				break
			}
		}
		if secretFloorID != "" {
			break
		}
	}
	if secretFloorID == "" {
		t.Fatal("fixture has no secret cell")
	}
	state := domain.SessionState{
		Rolls:          []domain.DiceRoll{},
		RevealedCells:  map[string][]domain.GridPosition{secretFloorID: {secretCell}},
		TokenPositions: map[string]domain.GridPosition{"secret-token": secretCell},
		TokenFloors:    map[string]string{"secret-token": secretFloorID},
		TokenOwners:    map[string]string{"secret-token": "gm"},
		Initiative: domain.InitiativeState{
			Entries: []domain.InitiativeEntry{{TokenID: "secret-token", Name: "Secret", Score: 20}}, Round: 1,
		},
	}
	player := domain.Participant{ID: "player", Role: "player"}
	visibleState, visible := filteredSnapshot(state, doc, player)
	roomCount := 0
	for _, floor := range visible.Floors {
		roomCount += len(floor.Rooms)
		for _, room := range floor.Rooms {
			if room.Secret {
				t.Fatalf("secret room %s leaked to the player", room.ID)
			}
		}
		for _, entity := range floor.Entities {
			if entity.Hidden {
				t.Fatalf("hidden entity %s leaked to the player", entity.ID)
			}
		}
	}
	if visible.Analysis.TotalRooms != roomCount {
		t.Fatalf("analysis leaks hidden room count: got %d, visible %d", visible.Analysis.TotalRooms, roomCount)
	}
	if len(visible.Analysis.DeadEnds) != 0 {
		t.Fatal("secret dead-end identifiers leaked to the player")
	}
	if len(visible.Progression.SecretRoomIDs) != 0 {
		t.Fatal("secret progression identifiers leaked to the player")
	}
	if len(visible.Encounters) != 0 || len(visible.Treasures) != 0 || len(visible.Puzzles) != 0 || len(visible.Traps) != 0 || len(visible.RestPoints) != 0 {
		t.Fatal("GM-only adventure content leaked to the player")
	}
	if visible.Analysis.ContentCounts != (domain.ContentCounts{}) || visible.Analysis.EncounterBudgetXP != 0 || visible.Analysis.TreasureValueGP != 0 {
		t.Fatal("GM-only content analysis leaked to the player")
	}
	if len(visibleState.RevealedCells[secretFloorID]) != 0 || len(visibleState.TokenPositions) != 0 || len(visibleState.Initiative.Entries) != 0 {
		t.Fatalf("secret runtime state leaked to the player: %#v", visibleState)
	}
	for _, eventType := range []string{"fog.changed", "map.pinged", "token.moved"} {
		projected := projectEvent(player, domain.SessionEvent{
			Revision: 3, Type: eventType, ActorID: "gm", OccurredAt: time.Now().UTC(),
			Payload: map[string]any{"tokenId": "secret-token", "floorId": secretFloorID, "x": secretCell.X, "z": secretCell.Z},
		}, doc)
		if projected.Type != "session.revision" || len(projected.Payload) != 0 {
			t.Fatalf("%s leaked a secret cell: %#v", eventType, projected)
		}
	}
}

func TestConfirmedSessionRestoresAfterServerRestart(t *testing.T) {
	database := testdb.Open(t)
	doc, err := generator.Generate(domain.DefaultAdventureSpec(), 777, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := database.SaveAdventure(ctx, doc, nil, "generated"); err != nil {
		t.Fatal(err)
	}
	hub := NewHub(database)
	created, err := hub.Create(ctx, doc, "GM")
	if err != nil {
		t.Fatal(err)
	}
	revision := created.State.Revision
	apply := func(id, commandType string, payload map[string]any) {
		t.Helper()
		event, commandErr := hub.HandleCommand(ctx, created.SessionID, created.Token, domain.SessionCommand{
			ID: id, ExpectedRevision: revision, Type: commandType, Payload: payload,
		})
		if commandErr != nil {
			t.Fatalf("apply %s before restart: %v", commandType, commandErr)
		}
		revision = event.Revision
	}
	floor := doc.Floors[0]
	target := floor.Tiles[0]
	for _, tile := range floor.Tiles {
		occupied := false
		for _, entity := range floor.Entities {
			if (entity.Kind == "token" || entity.Kind == "boss") && entity.Position.X == tile.X && entity.Position.Z == tile.Z {
				occupied = true
				break
			}
		}
		if tile.Walkable && !occupied {
			target = tile
			break
		}
	}
	apply("persist-fog", "fog.reveal", map[string]any{"floorId": floor.ID, "x": target.X, "z": target.Z})
	apply("persist-token", "token.move", map[string]any{"tokenId": "token-party", "floorId": floor.ID, "x": target.X, "z": target.Z})
	initiative := domain.InitiativeState{Entries: []domain.InitiativeEntry{{TokenID: "token-party", Name: "Party", Score: 17}}, ActiveIndex: 0, Round: 2}
	apply("persist-initiative", "initiative.set", map[string]any{"initiative": initiative})
	apply("persist-roll", "dice.roll", map[string]any{"expression": "1d20+2", "visibility": "public"})

	restoredHub := NewHub(database)
	subscription, snapshot, err := restoredHub.Subscribe(created.SessionID, created.Token)
	if err != nil {
		t.Fatal(err)
	}
	defer restoredHub.Unsubscribe(created.SessionID, subscription)
	if snapshot.State.ID != created.SessionID || snapshot.Adventure.ID != doc.ID {
		t.Fatal("restored session does not match the persisted snapshot")
	}
	if snapshot.State.TokenPositions["token-party"] != (domain.GridPosition{X: target.X, Z: target.Z}) {
		t.Fatalf("token position was not restored: %#v", snapshot.State.TokenPositions["token-party"])
	}
	if !cellRevealed(snapshot.State.RevealedCells[floor.ID], domain.GridPosition{X: target.X, Z: target.Z}) {
		t.Fatal("fog state was not restored")
	}
	if snapshot.State.Initiative.Round != 2 || len(snapshot.State.Initiative.Entries) != 1 {
		t.Fatalf("initiative was not restored: %#v", snapshot.State.Initiative)
	}
	if len(snapshot.State.Rolls) != 1 || snapshot.State.Rolls[0].ID != "persist-roll" {
		t.Fatalf("roll history was not restored: %#v", snapshot.State.Rolls)
	}
	if _, err := restoredHub.Join(ctx, created.SessionID, created.Code, "After restart", "player"); err != nil {
		t.Fatalf("persisted join code was not restored: %v", err)
	}
}

func TestCommandRetryAndReplayPreserveAuthoritativeRevision(t *testing.T) {
	hub, created := newAdministrativeTestHub(t)
	command := domain.SessionCommand{
		ID:               "idempotent-ping",
		ExpectedRevision: created.State.Revision,
		Type:             "ping",
		Payload: map[string]any{
			"floorId": created.State.ActiveFloorID,
			"x":       1,
			"z":       1,
		},
	}
	first, err := hub.HandleCommand(context.Background(), created.SessionID, created.Token, command)
	if err != nil {
		t.Fatal(err)
	}
	repeated, err := hub.HandleCommand(context.Background(), created.SessionID, created.Token, command)
	if err != nil || repeated.Revision != first.Revision || repeated.Type != first.Type {
		t.Fatalf("repeated command diverged: first=%#v repeated=%#v err=%v", first, repeated, err)
	}
	lastRevision := int64(0)
	sub, messages, err := hub.SubscribeFrom(created.SessionID, created.Token, &lastRevision)
	if err != nil {
		t.Fatal(err)
	}
	defer hub.Unsubscribe(created.SessionID, sub)
	if len(messages) < 2 {
		t.Fatalf("expected pending event replay, got %#v", messages)
	}
	latest := int64(0)
	for _, message := range messages {
		event, ok := message.(domain.SessionEvent)
		if !ok {
			t.Fatalf("recent reconnect unexpectedly received a snapshot: %#v", message)
		}
		if event.Revision <= latest {
			t.Fatalf("replay revisions are not monotonic: %d after %d", event.Revision, latest)
		}
		latest = event.Revision
	}
	if latest < first.Revision {
		t.Fatalf("replay did not reach the confirmed revision: %d", latest)
	}
}

func TestReplayRedactsRestrictedEventsWithoutCreatingRevisionGaps(t *testing.T) {
	hub, created := newAdministrativeTestHub(t)
	player, err := hub.Join(context.Background(), created.SessionID, created.Code, "Player", "player")
	if err != nil {
		t.Fatal(err)
	}
	beforeRoll := player.State.Revision
	roll, err := hub.HandleCommand(context.Background(), created.SessionID, created.Token, domain.SessionCommand{
		ID:               "gm-secret-roll",
		ExpectedRevision: beforeRoll,
		Type:             "dice.roll",
		Payload: map[string]any{
			"expression": "1d20",
			"visibility": "gm",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	sub, messages, err := hub.SubscribeFrom(created.SessionID, player.Token, &beforeRoll)
	if err != nil {
		t.Fatal(err)
	}
	defer hub.Unsubscribe(created.SessionID, sub)
	foundRedactedRevision := false
	for _, message := range messages {
		event, ok := message.(domain.SessionEvent)
		if !ok {
			t.Fatalf("recent replay unexpectedly used a snapshot: %#v", message)
		}
		if event.Revision == roll.Revision {
			if event.Type != "session.revision" || len(event.Payload) != 0 {
				t.Fatalf("restricted roll leaked through replay: %#v", event)
			}
			foundRedactedRevision = true
		}
	}
	if !foundRedactedRevision {
		t.Fatalf("restricted revision disappeared from replay: %#v", messages)
	}
}

func newAdministrativeTestHub(t *testing.T) (*Hub, Created) {
	t.Helper()
	database := testdb.Open(t)
	doc, err := generator.Generate(domain.DefaultAdventureSpec(), 811, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if err := database.SaveAdventure(context.Background(), doc, nil, "generated"); err != nil {
		t.Fatal(err)
	}
	hub := NewHub(database)
	created, err := hub.Create(context.Background(), doc, "GM")
	if err != nil {
		t.Fatal(err)
	}
	return hub, created
}

func TestDisplayIsStrictlyReadOnly(t *testing.T) {
	hub, created := newAdministrativeTestHub(t)
	joined, err := hub.Join(context.Background(), created.SessionID, created.Code, "Display", "display")
	if err != nil {
		t.Fatal(err)
	}
	for _, commandType := range []string{"ping", "dice.roll", "token.move", "fog.reveal", "initiative.set"} {
		command := domain.SessionCommand{ID: commandType, ExpectedRevision: joined.State.Revision, Type: commandType, Payload: map[string]any{}}
		if _, err := hub.HandleCommand(context.Background(), created.SessionID, joined.Token, command); !errors.Is(err, ErrUnauthorized) {
			t.Fatalf("display command %s: expected unauthorized, got %v", commandType, err)
		}
	}
}

func TestGMControlsPermissionsRolesTokensAndRemoval(t *testing.T) {
	hub, created := newAdministrativeTestHub(t)
	player, err := hub.Join(context.Background(), created.SessionID, created.Code, "Player", "player")
	if err != nil {
		t.Fatal(err)
	}
	revision := player.State.Revision
	event, err := hub.HandleCommand(context.Background(), created.SessionID, created.Token, domain.SessionCommand{
		ID: "permissions", ExpectedRevision: revision, Type: "permissions.set",
		Payload: map[string]any{"permissions": domain.SessionPermissions{}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := hub.HandleCommand(context.Background(), created.SessionID, player.Token, domain.SessionCommand{ID: "ping", ExpectedRevision: event.Revision, Type: "ping", Payload: map[string]any{"floorId": created.State.ActiveFloorID, "x": 1, "z": 1}}); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("disabled player ping: expected unauthorized, got %v", err)
	}
	event, err = hub.HandleCommand(context.Background(), created.SessionID, created.Token, domain.SessionCommand{ID: "assign", ExpectedRevision: event.Revision, Type: "token.assign", Payload: map[string]any{"tokenId": "token-party", "participantId": player.ParticipantID}})
	if err != nil {
		t.Fatal(err)
	}
	event, err = hub.HandleCommand(context.Background(), created.SessionID, created.Token, domain.SessionCommand{ID: "role", ExpectedRevision: event.Revision, Type: "participant.role.set", Payload: map[string]any{"participantId": player.ParticipantID, "role": "display"}})
	if err != nil {
		t.Fatal(err)
	}
	s, _ := hub.find(created.SessionID)
	if _, assigned := s.state.TokenOwners["token-party"]; assigned {
		t.Fatal("display retained token assignment")
	}
	sub, snapshot, err := hub.Subscribe(created.SessionID, player.Token)
	if err != nil {
		t.Fatal(err)
	}
	defer hub.Unsubscribe(created.SessionID, sub)
	if _, err := hub.HandleCommand(context.Background(), created.SessionID, created.Token, domain.SessionCommand{ID: "remove", ExpectedRevision: snapshot.State.Revision, Type: "participant.remove", Payload: map[string]any{"participantId": player.ParticipantID}}); err != nil {
		t.Fatal(err)
	}
	select {
	case <-sub.Revoked:
	case <-time.After(time.Second):
		t.Fatal("removed participant connection was not revoked")
	}
	if _, _, err := hub.Subscribe(created.SessionID, player.Token); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("removed credential remained valid: %v", err)
	}
}

func TestOptionalAdmissionApprovalFlow(t *testing.T) {
	hub, created := newAdministrativeTestHub(t)
	event, err := hub.HandleCommand(context.Background(), created.SessionID, created.Token, domain.SessionCommand{ID: "settings", ExpectedRevision: 0, Type: "admission.set", Payload: map[string]any{"joinOpen": true, "approvalRequired": true}})
	if err != nil {
		t.Fatal(err)
	}
	pending, err := hub.Join(context.Background(), created.SessionID, created.Code, "Waiting", "player")
	if err != nil || pending.Status != "pending" {
		t.Fatalf("expected pending join, got %#v, %v", pending, err)
	}
	if _, _, err := hub.Subscribe(created.SessionID, pending.Token); !errors.Is(err, ErrUnauthorized) {
		t.Fatal("pending participant subscribed before approval")
	}
	event, err = hub.HandleCommand(context.Background(), created.SessionID, created.Token, domain.SessionCommand{ID: "approve", ExpectedRevision: event.Revision + 1, Type: "admission.approve", Payload: map[string]any{"requestId": pending.RequestID}})
	if err != nil {
		t.Fatal(err)
	}
	status, err := hub.AdmissionStatus(created.SessionID, pending.RequestID, pending.Token)
	if err != nil || status.Status != "joined" || status.State.Revision != event.Revision {
		t.Fatalf("approved status mismatch: %#v, %v", status, err)
	}
}
