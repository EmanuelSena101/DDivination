package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/EmanuelSena101/DDivination/apps/server/internal/domain"
	"github.com/EmanuelSena101/DDivination/apps/server/internal/generator"
)

func TestValidateAdventureAcceptsGeneratedDocument(t *testing.T) {
	document, err := generator.Generate(domain.DefaultAdventureSpec(), 77, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if err := domain.ValidateAdventure(document); err != nil {
		t.Fatalf("generated document should be valid: %v", err)
	}
}

func TestValidateAdventureRejectsInvalidEntityAndDisconnectedFloor(t *testing.T) {
	document, err := generator.Generate(domain.DefaultAdventureSpec(), 77, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	document.Floors[0].Entities[0].Position = domain.GridPosition{X: -1, Z: -1}
	if err := domain.ValidateAdventure(document); !errors.Is(err, domain.ErrInvalidAdventure) {
		t.Fatalf("expected invalid entity rejection, got %v", err)
	}

	document, err = generator.Generate(domain.DefaultAdventureSpec(), 77, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	for index := range document.Floors[0].Tiles {
		tile := &document.Floors[0].Tiles[index]
		if tile.X == document.Floors[0].Rooms[1].Center.X &&
			tile.Z == document.Floors[0].Rooms[1].Center.Z {
			tile.Walkable = false
			break
		}
	}
	if err := domain.ValidateAdventure(document); !errors.Is(err, domain.ErrInvalidAdventure) {
		t.Fatalf("expected disconnected room rejection, got %v", err)
	}
}
