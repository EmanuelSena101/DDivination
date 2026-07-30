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
	checkpoint, err := database.CreateAdventureSnapshot(ctx, doc.ID, "manual-checkpoint")
	if err != nil {
		t.Fatal(err)
	}
	snapshots, err := database.ListAdventureSnapshots(ctx, doc.ID, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshots) != 2 || snapshots[0].ID != checkpoint.ID {
		t.Fatalf("expected manual and generated snapshots, got %#v", snapshots)
	}
	loadedSnapshot, err := database.GetAdventureSnapshot(ctx, doc.ID, checkpoint.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loadedSnapshot.Document.ID != doc.ID || loadedSnapshot.Version != doc.Version {
		t.Fatalf("snapshot did not round-trip: %#v", loadedSnapshot)
	}
	if err := database.DeleteAdventure(ctx, doc.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := database.GetAdventure(ctx, doc.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected deleted adventure to be absent, got %v", err)
	}
}
