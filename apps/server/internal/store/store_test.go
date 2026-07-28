package store

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/EmanuelSena101/DDivination/apps/server/internal/domain"
	"github.com/EmanuelSena101/DDivination/apps/server/internal/generator"
)

func TestAdventurePersistenceAndOptimisticLock(t *testing.T) {
	database, err := Open(filepath.Join(t.TempDir(), "test.sqlite3"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	ctx := context.Background()
	doc, err := generator.Generate(domain.DefaultAdventureSpec(), 7, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if err := database.SaveAdventure(ctx, doc, nil, "generated"); err != nil {
		t.Fatal(err)
	}
	loaded, err := database.GetAdventure(ctx, doc.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Seed != doc.Seed || loaded.Name != doc.Name {
		t.Fatal("persisted adventure did not round-trip")
	}
	wrongVersion := int64(42)
	if err := database.SaveAdventure(ctx, doc, &wrongVersion, "manual"); !errors.Is(err, ErrConflict) {
		t.Fatalf("expected optimistic lock conflict, got %v", err)
	}
	if err := database.DeleteAdventure(ctx, doc.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := database.GetAdventure(ctx, doc.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected deleted adventure to be absent, got %v", err)
	}
}
