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

func TestValidateAdventureRejectsUnsolvableProgression(t *testing.T) {
	document := generatedAdventure(t)
	firstLock := document.Progression.Locks[0]
	for index := range document.Progression.Steps {
		step := &document.Progression.Steps[index]
		if contains(step.GrantsKeyIDs, firstLock.KeyID) {
			step.GrantsKeyIDs = nil
			break
		}
	}
	last := &document.Progression.Steps[len(document.Progression.Steps)-1]
	last.GrantsKeyIDs = append(last.GrantsKeyIDs, firstLock.KeyID)

	if err := domain.ValidateAdventure(document); !errors.Is(err, domain.ErrInvalidAdventure) {
		t.Fatalf("expected key-after-lock rejection, got %v", err)
	}
}

func TestValidateAdventureRejectsInvalidProgressionTarget(t *testing.T) {
	document := generatedAdventure(t)
	document.Progression.Locks[0].TargetID = "missing-target"
	if err := domain.ValidateAdventure(document); !errors.Is(err, domain.ErrInvalidAdventure) {
		t.Fatalf("expected invalid lock target rejection, got %v", err)
	}
}

func TestValidateAdventureRejectsSecretOnMandatoryPath(t *testing.T) {
	document := generatedAdventure(t)
	secretID := document.Progression.SecretRoomIDs[0]
	var secretFloorID string
	for _, floor := range document.Floors {
		for _, room := range floor.Rooms {
			if room.ID == secretID {
				secretFloorID = floor.ID
			}
		}
	}
	document.Progression.Steps[1].RoomID = secretID
	document.Progression.Steps[1].FloorID = secretFloorID
	document.Analysis.CriticalPath[1] = secretID
	if err := domain.ValidateAdventure(document); !errors.Is(err, domain.ErrInvalidAdventure) {
		t.Fatalf("expected secret mandatory step rejection, got %v", err)
	}
}

func generatedAdventure(t *testing.T) domain.AdventureDocument {
	t.Helper()
	document, err := generator.Generate(domain.DefaultAdventureSpec(), 1701, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	return document
}

func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
