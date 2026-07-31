package store_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/EmanuelSena101/DDivination/apps/server/internal/domain"
	"github.com/EmanuelSena101/DDivination/apps/server/internal/generator"
	"github.com/EmanuelSena101/DDivination/apps/server/internal/store"
	"github.com/EmanuelSena101/DDivination/apps/server/internal/testdb"
)

func TestAdventurePersistenceAndOptimisticLock(t *testing.T) {
	database := testdb.Open(t)

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
	if err := database.SaveAdventure(ctx, doc, &wrongVersion, "manual"); !errors.Is(err, store.ErrConflict) {
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
	if _, err := database.GetAdventure(ctx, doc.ID); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("expected deleted adventure to be absent, got %v", err)
	}
}

func TestPostgresMigrationsAreIdempotentAcrossPools(t *testing.T) {
	database, databaseURL := testdb.OpenWithURL(t)
	ctx := context.Background()
	doc, err := generator.Generate(domain.DefaultAdventureSpec(), 17, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if err := database.SaveAdventure(ctx, doc, nil, "migration-fixture"); err != nil {
		t.Fatal(err)
	}

	reopened, err := store.Open(databaseURL)
	if err != nil {
		t.Fatalf("reopen and revalidate migrations: %v", err)
	}
	defer reopened.Close()
	loaded, err := reopened.GetAdventure(ctx, doc.ID)
	if err != nil || loaded.ID != doc.ID {
		t.Fatalf("fixture did not survive a second migration pass: %#v, %v", loaded, err)
	}
}

func TestSessionEventLogIsIdempotentReplayableAndCompactable(t *testing.T) {
	database := testdb.Open(t)
	ctx := context.Background()
	doc, err := generator.Generate(domain.DefaultAdventureSpec(), 88, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if err := database.SaveAdventure(ctx, doc, nil, "generated"); err != nil {
		t.Fatal(err)
	}
	state := domain.SessionState{
		ID:             "session-log",
		AdventureID:    doc.ID,
		Participants:   map[string]domain.Participant{},
		TokenPositions: map[string]domain.GridPosition{},
		TokenFloors:    map[string]string{},
		TokenOwners:    map[string]string{},
		RevealedCells:  map[string][]domain.GridPosition{},
		Admissions:     map[string]domain.AdmissionRequest{},
		Rolls:          []domain.DiceRoll{},
		Open:           true,
		CreatedAt:      time.Now().UTC(),
	}
	if err := database.InitializeSession(ctx, store.SessionInitialization{State: state, JoinCodeHash: "join-code-hash", JoinCodeExpiresAt: time.Now().UTC().Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}

	var last store.SessionCommitResult
	for revision := int64(1); revision <= 1000; revision++ {
		state.Revision = revision
		event := domain.SessionEvent{
			Revision:   revision,
			Type:       "ping.created",
			ActorID:    "gm",
			OccurredAt: time.Now().UTC(),
			Payload:    map[string]any{"revision": revision},
		}
		last, err = database.CommitSession(ctx, store.SessionCommit{
			SessionID:        state.ID,
			CommandID:        fmt.Sprintf("command-%d", revision),
			ExpectedRevision: revision - 1,
			Event:            event,
			State:            state,
		})
		if err != nil || last.Duplicate {
			t.Fatalf("commit revision %d: %#v, %v", revision, last, err)
		}
	}

	duplicate, err := database.CommitSession(ctx, store.SessionCommit{
		SessionID:        state.ID,
		CommandID:        "command-1000",
		ExpectedRevision: 999,
		Event:            last.Event,
		State:            state,
	})
	if err != nil || !duplicate.Duplicate || duplicate.Event.Revision != 1000 {
		t.Fatalf("idempotent command mismatch: %#v, %v", duplicate, err)
	}

	conflictingState := state
	conflictingState.Revision = 1001
	_, err = database.CommitSession(ctx, store.SessionCommit{
		SessionID:        state.ID,
		CommandID:        "conflict",
		ExpectedRevision: 999,
		Event: domain.SessionEvent{
			Revision: 1001, Type: "ping.created", ActorID: "gm", OccurredAt: time.Now().UTC(), Payload: map[string]any{},
		},
		State: conflictingState,
	})
	if !errors.Is(err, store.ErrRevisionConflict) {
		t.Fatalf("expected revision conflict, got %v", err)
	}

	replay, err := database.LoadSessionReplay(ctx, state.ID, 995, 100)
	if err != nil || !replay.Complete || len(replay.Events) != 5 || replay.CurrentRevision != 1000 {
		t.Fatalf("recent replay mismatch: %#v, %v", replay, err)
	}
	deleted, err := database.CompactSessionEvents(ctx, state.ID, 901)
	if err != nil || deleted != 400 {
		t.Fatalf("compaction mismatch: deleted=%d, err=%v", deleted, err)
	}
	oldReplay, err := database.LoadSessionReplay(ctx, state.ID, 0, 1000)
	if err != nil || oldReplay.Complete || len(oldReplay.Events) != 0 || oldReplay.OldestRevision != 901 {
		t.Fatalf("old replay should require a snapshot: %#v, %v", oldReplay, err)
	}
	restored, err := database.LoadSession(ctx, state.ID)
	if err != nil || restored.State.Revision != 1000 || restored.JoinCodeHash != "join-code-hash" {
		t.Fatalf("session head mismatch: %#v, %v", restored, err)
	}
}

func TestSessionCommitFailureRecoveryAndConcurrentIsolation(t *testing.T) {
	database := testdb.Open(t)
	ctx := context.Background()
	doc, err := generator.Generate(domain.DefaultAdventureSpec(), 144, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if err := database.SaveAdventure(ctx, doc, nil, "generated"); err != nil {
		t.Fatal(err)
	}

	newState := func(id string) domain.SessionState {
		return domain.SessionState{
			ID:             id,
			AdventureID:    doc.ID,
			Participants:   map[string]domain.Participant{},
			TokenPositions: map[string]domain.GridPosition{},
			TokenFloors:    map[string]string{},
			TokenOwners:    map[string]string{},
			RevealedCells:  map[string][]domain.GridPosition{},
			Admissions:     map[string]domain.AdmissionRequest{},
			Rolls:          []domain.DiceRoll{},
			Open:           true,
			CreatedAt:      time.Now().UTC(),
		}
	}
	states := []domain.SessionState{newState("isolated-a"), newState("isolated-b")}
	for _, state := range states {
		if err := database.InitializeSession(ctx, store.SessionInitialization{State: state, JoinCodeHash: "hash-" + state.ID, JoinCodeExpiresAt: time.Now().UTC().Add(time.Hour)}); err != nil {
			t.Fatal(err)
		}
	}

	failedState := states[0]
	failedState.Revision = 1
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	_, err = database.CommitSession(canceled, store.SessionCommit{
		SessionID:        failedState.ID,
		CommandID:        "canceled-before-commit",
		ExpectedRevision: 0,
		Event: domain.SessionEvent{
			Revision: 1, Type: "map.pinged", ActorID: "gm", OccurredAt: time.Now().UTC(), Payload: map[string]any{},
		},
		State: failedState,
	})
	if err == nil {
		t.Fatal("a canceled transaction unexpectedly committed")
	}
	unchanged, err := database.LoadSession(ctx, failedState.ID)
	if err != nil || unchanged.State.Revision != 0 {
		t.Fatalf("failed commit changed the durable head: %#v, %v", unchanged, err)
	}

	results := make(chan error, len(states))
	var wg sync.WaitGroup
	for index := range states {
		state := states[index]
		wg.Add(1)
		go func() {
			defer wg.Done()
			state.Revision = 1
			event := domain.SessionEvent{
				Revision: 1, Type: "map.pinged", ActorID: "gm", OccurredAt: time.Now().UTC(), Payload: map[string]any{"session": state.ID},
			}
			_, commitErr := database.CommitSession(ctx, store.SessionCommit{
				SessionID: state.ID, CommandID: "same-command-id", ExpectedRevision: 0, Event: event, State: state,
			})
			results <- commitErr
		}()
	}
	wg.Wait()
	close(results)
	for commitErr := range results {
		if commitErr != nil {
			t.Fatalf("isolated session commit failed: %v", commitErr)
		}
	}
	for _, state := range states {
		restored, err := database.LoadSession(ctx, state.ID)
		if err != nil || restored.State.Revision != 1 {
			t.Fatalf("session %s was not independently committed: %#v, %v", state.ID, restored, err)
		}
	}

	// Simulate losing the response after PostgreSQL committed. Retrying the same
	// command returns the original event without advancing the revision.
	retryState := states[0]
	retryState.Revision = 1
	retry, err := database.CommitSession(ctx, store.SessionCommit{
		SessionID: retryState.ID, CommandID: "same-command-id", ExpectedRevision: 0,
		Event: domain.SessionEvent{Revision: 1, Type: "map.pinged", ActorID: "gm", OccurredAt: time.Now().UTC()},
		State: retryState,
	})
	if err != nil || !retry.Duplicate || retry.Event.Revision != 1 {
		t.Fatalf("post-commit retry was not idempotent: %#v, %v", retry, err)
	}
}

func TestClosedSessionRetentionNeverPrunesOpenTables(t *testing.T) {
	database := testdb.Open(t)
	ctx := context.Background()
	doc, err := generator.Generate(domain.DefaultAdventureSpec(), 233, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if err := database.SaveAdventure(ctx, doc, nil, "generated"); err != nil {
		t.Fatal(err)
	}
	base := domain.SessionState{
		AdventureID: doc.ID, Participants: map[string]domain.Participant{}, TokenPositions: map[string]domain.GridPosition{},
		TokenFloors: map[string]string{}, TokenOwners: map[string]string{}, RevealedCells: map[string][]domain.GridPosition{},
		Admissions: map[string]domain.AdmissionRequest{}, Rolls: []domain.DiceRoll{}, Open: true, CreatedAt: time.Now().UTC(),
	}
	openState := base
	openState.ID = "retention-open"
	closedState := base
	closedState.ID = "retention-closed"
	for _, state := range []domain.SessionState{openState, closedState} {
		if err := database.InitializeSession(ctx, store.SessionInitialization{State: state, JoinCodeHash: "hash", JoinCodeExpiresAt: time.Now().UTC().Add(time.Hour)}); err != nil {
			t.Fatal(err)
		}
	}
	closedState.Open = false
	closedState.Revision = 1
	if _, err := database.CommitSession(ctx, store.SessionCommit{
		SessionID: closedState.ID, CommandID: "close", ExpectedRevision: 0,
		Event: domain.SessionEvent{Revision: 1, Type: "session.closed", ActorID: "gm", OccurredAt: time.Now().UTC(), Payload: map[string]any{}},
		State: closedState, ForceSnapshot: true,
	}); err != nil {
		t.Fatal(err)
	}
	deleted, err := database.PruneClosedSessions(ctx, time.Now().UTC().Add(time.Hour))
	if err != nil || deleted != 1 {
		t.Fatalf("closed retention mismatch: deleted=%d, err=%v", deleted, err)
	}
	if _, err := database.LoadSession(ctx, closedState.ID); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("closed session was not pruned: %v", err)
	}
	if _, err := database.LoadSession(ctx, openState.ID); err != nil {
		t.Fatalf("open session was pruned: %v", err)
	}
}
