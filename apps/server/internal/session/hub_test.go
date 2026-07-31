package session

import (
	"context"
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
