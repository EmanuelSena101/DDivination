package session

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/EmanuelSena101/DDivination/apps/server/internal/domain"
	"github.com/EmanuelSena101/DDivination/apps/server/internal/generator"
	"github.com/EmanuelSena101/DDivination/apps/server/internal/store"
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
	state := domain.SessionState{Rolls: []domain.DiceRoll{}}
	_, visible := filteredSnapshot(state, doc, domain.Participant{ID: "player", Role: "player"})
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
}

func TestConfirmedSessionRestoresAfterServerRestart(t *testing.T) {
	database, err := store.Open(filepath.Join(t.TempDir(), "session.sqlite3"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	doc, err := generator.Generate(domain.DefaultAdventureSpec(), 777, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := database.SaveAdventure(ctx, doc, nil, "generated"); err != nil {
		t.Fatal(err)
	}
	created, err := NewHub(database).Create(ctx, doc, "GM")
	if err != nil {
		t.Fatal(err)
	}
	restoredHub := NewHub(database)
	subscription, snapshot, err := restoredHub.Subscribe(created.SessionID, created.Token)
	if err != nil {
		t.Fatal(err)
	}
	defer restoredHub.Unsubscribe(created.SessionID, subscription)
	if snapshot.State.ID != created.SessionID || snapshot.Adventure.ID != doc.ID {
		t.Fatal("restored session does not match the persisted snapshot")
	}
}

func newAdministrativeTestHub(t *testing.T) (*Hub, Created) {
	t.Helper()
	database, err := store.Open(filepath.Join(t.TempDir(), "session.sqlite3"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
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
