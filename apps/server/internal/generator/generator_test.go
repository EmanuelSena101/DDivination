package generator

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/EmanuelSena101/DDivination/apps/server/internal/domain"
)

func TestGenerateContextHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := GenerateContext(
		ctx,
		domain.DefaultAdventureSpec(),
		1,
		time.Now(),
		nil,
	); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}

func TestGenerateIsDeterministic(t *testing.T) {
	now := time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)
	spec := domain.DefaultAdventureSpec()
	spec.FloorCount = 3
	spec.StructureStyle = "labyrinthine"

	first, err := Generate(spec, 424242, now)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Generate(spec, 424242, now)
	if err != nil {
		t.Fatal(err)
	}
	a, _ := json.Marshal(first)
	b, _ := json.Marshal(second)
	if string(a) != string(b) {
		t.Fatal("same seed, version, spec and clock produced different documents")
	}
}

func TestGeneratedDungeonInvariants(t *testing.T) {
	now := time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)
	styles := []string{"linear", "branching", "labyrinthine"}
	for seed := uint64(0); seed < 1000; seed++ {
		spec := domain.DefaultAdventureSpec()
		spec.FloorCount = 1 + int(seed%5)
		spec.DurationHours = 2 + int(seed%7)
		spec.StructureStyle = styles[seed%uint64(len(styles))]
		doc, err := Generate(spec, seed, now)
		if err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
		if len(doc.Floors) != spec.FloorCount {
			t.Fatalf("seed %d: got %d floors", seed, len(doc.Floors))
		}
		if !doc.Progression.Solvable || len(doc.Progression.Locks) != spec.FloorCount {
			t.Fatalf("seed %d: invalid progression summary", seed)
		}
		if doc.Progression.Steps[0].RoomID != doc.Progression.EntryRoomID ||
			doc.Progression.Steps[len(doc.Progression.Steps)-1].RoomID != doc.Progression.ClimaxRoomID {
			t.Fatalf("seed %d: progression endpoints mismatch", seed)
		}
		for floorIndex, floor := range doc.Floors {
			walkable := make(map[domain.GridPosition]bool, len(floor.Tiles))
			for _, tile := range floor.Tiles {
				walkable[domain.GridPosition{X: tile.X, Z: tile.Z}] = tile.Walkable
			}
			start := floor.Rooms[0].Center
			reached := flood(start, walkable)
			for _, room := range floor.Rooms {
				if !reached[room.Center] {
					t.Fatalf("seed %d floor %d: room %s is unreachable", seed, floorIndex, room.ID)
				}
			}
			for _, portal := range floor.Portals {
				if portal.ToFloorID == "" {
					t.Fatalf("seed %d floor %d: portal %s has no destination", seed, floorIndex, portal.ID)
				}
				if !walkable[portal.From] {
					t.Fatalf("seed %d floor %d: portal %s is not on a walkable tile", seed, floorIndex, portal.ID)
				}
			}
		}
	}
}

func flood(start domain.GridPosition, walkable map[domain.GridPosition]bool) map[domain.GridPosition]bool {
	reached := map[domain.GridPosition]bool{start: true}
	queue := []domain.GridPosition{start}
	directions := []domain.GridPosition{{X: 1}, {X: -1}, {Z: 1}, {Z: -1}}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, direction := range directions {
			next := domain.GridPosition{X: current.X + direction.X, Z: current.Z + direction.Z}
			if walkable[next] && !reached[next] {
				reached[next] = true
				queue = append(queue, next)
			}
		}
	}
	return reached
}

func TestRejectsOutOfRangeSpec(t *testing.T) {
	spec := domain.DefaultAdventureSpec()
	spec.FloorCount = 6
	if _, err := Generate(spec, 1, time.Now()); err == nil {
		t.Fatal("expected invalid floor count to be rejected")
	}
}
