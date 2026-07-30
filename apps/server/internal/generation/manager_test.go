package generation

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/EmanuelSena101/DDivination/apps/server/internal/domain"
	"github.com/EmanuelSena101/DDivination/apps/server/internal/store"
)

func TestManagerCompletesWithMonotonicPersistedStages(t *testing.T) {
	database := openTestStore(t)
	manager := NewManager(database, testLogger())
	run := domain.GenerationRun{
		ID:               "run-complete",
		Seed:             42,
		GeneratorVersion: domain.GeneratorVersion,
		Spec:             domain.DefaultAdventureSpec(),
	}
	queued, err := manager.Enqueue(t.Context(), run, func(
		_ context.Context,
		report Reporter,
	) (string, error) {
		if err := report("planning", 12); err != nil {
			return "", err
		}
		if err := report("building", 68); err != nil {
			return "", err
		}
		return "adv-42", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if queued.Status != "queued" || queued.Progress != 0 {
		t.Fatalf("expected queued response, got %#v", queued)
	}

	completed := waitForStatus(t, database, run.ID, "completed")
	if completed.AdventureID != "adv-42" || completed.Progress != 100 {
		t.Fatalf("unexpected completion: %#v", completed)
	}
	previous := -1
	for _, stage := range completed.Stages {
		if stage.Progress < previous {
			t.Fatalf("progress regressed from %d to %d: %#v", previous, stage.Progress, completed.Stages)
		}
		previous = stage.Progress
	}
}

func TestManagerCancellationIsTerminalAndIdempotent(t *testing.T) {
	database := openTestStore(t)
	manager := NewManager(database, testLogger())
	started := make(chan struct{})
	run := domain.GenerationRun{
		ID:               "run-cancel",
		Seed:             43,
		GeneratorVersion: domain.GeneratorVersion,
		Spec:             domain.DefaultAdventureSpec(),
	}
	if _, err := manager.Enqueue(t.Context(), run, func(
		ctx context.Context,
		report Reporter,
	) (string, error) {
		if err := report("building", 35); err != nil {
			return "", err
		}
		close(started)
		<-ctx.Done()
		return "", ctx.Err()
	}); err != nil {
		t.Fatal(err)
	}
	<-started

	cancelled, err := manager.Cancel(t.Context(), run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled.Status != "cancelled" || cancelled.CompletedAt == nil {
		t.Fatalf("unexpected cancellation: %#v", cancelled)
	}
	again, err := manager.Cancel(t.Context(), run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if again.Status != "cancelled" || again.Progress != cancelled.Progress {
		t.Fatalf("idempotent cancellation changed the run: %#v", again)
	}
}

func TestManagerRecoversInterruptedRuns(t *testing.T) {
	database := openTestStore(t)
	now := time.Now().UTC()
	interrupted := domain.GenerationRun{
		ID:               "run-interrupted",
		Status:           "running",
		Stage:            "building",
		Progress:         55,
		Seed:             44,
		GeneratorVersion: domain.GeneratorVersion,
		Spec:             domain.DefaultAdventureSpec(),
		Diagnostics:      []string{},
		Stages: []domain.GenerationStage{{
			Name:       "building",
			Progress:   55,
			OccurredAt: now,
		}},
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := database.SaveGenerationRun(t.Context(), interrupted); err != nil {
		t.Fatal(err)
	}

	manager := NewManager(database, testLogger())
	if err := manager.Recover(t.Context()); err != nil {
		t.Fatal(err)
	}
	recovered, err := database.GetGenerationRun(t.Context(), interrupted.ID)
	if err != nil {
		t.Fatal(err)
	}
	if recovered.Status != "failed" || recovered.Stage != "interrupted" {
		t.Fatalf("unexpected recovered run: %#v", recovered)
	}
	if !slices.Contains(recovered.Diagnostics, "generation-interrupted:server-restarted") {
		t.Fatalf("missing interruption diagnostic: %#v", recovered.Diagnostics)
	}
}

func openTestStore(t *testing.T) *store.Store {
	t.Helper()
	database, err := store.Open(filepath.Join(t.TempDir(), "generation.sqlite3"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Errorf("close store: %v", err)
		}
	})
	return database
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func waitForStatus(
	t *testing.T,
	database *store.Store,
	runID, status string,
) domain.GenerationRun {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		run, err := database.GetGenerationRun(t.Context(), runID)
		if err == nil && run.Status == status {
			return run
		}
		time.Sleep(5 * time.Millisecond)
	}
	run, err := database.GetGenerationRun(t.Context(), runID)
	t.Fatalf("run did not reach %s: run=%#v err=%v", status, run, err)
	return domain.GenerationRun{}
}
