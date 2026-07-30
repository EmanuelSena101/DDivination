package generation

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/EmanuelSena101/DDivination/apps/server/internal/domain"
	"github.com/EmanuelSena101/DDivination/apps/server/internal/store"
)

var ErrInvalidTransition = errors.New("invalid generation run transition")

type Reporter func(stage string, progress int, diagnostics ...string) error

type Executor func(ctx context.Context, report Reporter) (adventureID string, err error)

type Manager struct {
	store  *store.Store
	logger *slog.Logger

	stateMu sync.Mutex
	mu      sync.Mutex
	cancels map[string]context.CancelFunc
	streams map[string]map[chan domain.GenerationRun]struct{}
}

func NewManager(database *store.Store, logger *slog.Logger) *Manager {
	return &Manager{
		store:   database,
		logger:  logger,
		cancels: make(map[string]context.CancelFunc),
		streams: make(map[string]map[chan domain.GenerationRun]struct{}),
	}
}

func (m *Manager) Recover(ctx context.Context) error {
	runs, err := m.store.ListGenerationRuns(ctx, 500)
	if err != nil {
		return err
	}
	for _, run := range runs {
		if run.Status != "queued" && run.Status != "running" {
			continue
		}
		if _, err := m.finish(
			ctx,
			run.ID,
			"failed",
			"interrupted",
			run.Progress,
			"",
			"generation-interrupted:server-restarted",
		); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) Enqueue(
	ctx context.Context,
	run domain.GenerationRun,
	executor Executor,
) (domain.GenerationRun, error) {
	now := time.Now().UTC()
	run.Status = "queued"
	run.Stage = "queued"
	run.Progress = 0
	run.CreatedAt = now
	run.UpdatedAt = now
	run.CompletedAt = nil
	run.Stages = []domain.GenerationStage{{
		Name:       "queued",
		Progress:   0,
		OccurredAt: now,
	}}
	if run.Diagnostics == nil {
		run.Diagnostics = []string{}
	}
	if err := m.store.SaveGenerationRun(ctx, run); err != nil {
		return domain.GenerationRun{}, err
	}

	jobCtx, cancel := context.WithCancel(context.Background())
	m.mu.Lock()
	m.cancels[run.ID] = cancel
	m.mu.Unlock()
	m.broadcast(run)

	go m.execute(jobCtx, run.ID, executor)
	return run, nil
}

func (m *Manager) execute(ctx context.Context, runID string, executor Executor) {
	defer func() {
		m.mu.Lock()
		delete(m.cancels, runID)
		m.mu.Unlock()
	}()

	if _, err := m.update(ctx, runID, "running", "starting", 1); err != nil {
		m.logger.Error("could not start generation run", "run_id", runID, "error", err)
		return
	}
	adventureID, err := executor(ctx, func(stage string, progress int, diagnostics ...string) error {
		_, updateErr := m.updateWithDiagnostics(ctx, runID, "running", stage, progress, diagnostics...)
		return updateErr
	})
	if errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
		if _, finishErr := m.finish(
			context.Background(),
			runID,
			"cancelled",
			"cancelled",
			0,
			"",
			"generation-cancelled",
		); finishErr != nil && !errors.Is(finishErr, ErrInvalidTransition) {
			m.logger.Error("could not cancel generation run", "run_id", runID, "error", finishErr)
		}
		return
	}
	if err != nil {
		if _, finishErr := m.finish(
			context.Background(),
			runID,
			"failed",
			"failed",
			0,
			"",
			err.Error(),
		); finishErr != nil {
			m.logger.Error("could not fail generation run", "run_id", runID, "error", finishErr)
		}
		return
	}
	if _, err := m.finish(
		context.Background(),
		runID,
		"completed",
		"completed",
		100,
		adventureID,
	); err != nil {
		m.logger.Error("could not complete generation run", "run_id", runID, "error", err)
	}
}

func (m *Manager) Cancel(ctx context.Context, runID string) (domain.GenerationRun, error) {
	run, err := m.store.GetGenerationRun(ctx, runID)
	if err != nil {
		return domain.GenerationRun{}, err
	}
	if terminal(run.Status) {
		return run, nil
	}

	m.mu.Lock()
	cancel := m.cancels[runID]
	m.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	return m.finish(
		context.Background(),
		runID,
		"cancelled",
		"cancelled",
		run.Progress,
		"",
		"generation-cancelled:user-request",
	)
}

func (m *Manager) Subscribe(
	ctx context.Context,
	runID string,
) (<-chan domain.GenerationRun, func(), error) {
	run, err := m.store.GetGenerationRun(ctx, runID)
	if err != nil {
		return nil, nil, err
	}
	stream := make(chan domain.GenerationRun, 1)
	stream <- run

	m.mu.Lock()
	if m.streams[runID] == nil {
		m.streams[runID] = make(map[chan domain.GenerationRun]struct{})
	}
	m.streams[runID][stream] = struct{}{}
	m.mu.Unlock()

	unsubscribe := func() {
		m.mu.Lock()
		delete(m.streams[runID], stream)
		if len(m.streams[runID]) == 0 {
			delete(m.streams, runID)
		}
		m.mu.Unlock()
	}
	return stream, unsubscribe, nil
}

func (m *Manager) update(
	ctx context.Context,
	runID, status, stage string,
	progress int,
) (domain.GenerationRun, error) {
	return m.updateWithDiagnostics(ctx, runID, status, stage, progress)
}

func (m *Manager) updateWithDiagnostics(
	ctx context.Context,
	runID, status, stage string,
	progress int,
	diagnostics ...string,
) (domain.GenerationRun, error) {
	m.stateMu.Lock()
	defer m.stateMu.Unlock()

	run, err := m.store.GetGenerationRun(ctx, runID)
	if err != nil {
		return domain.GenerationRun{}, err
	}
	if terminal(run.Status) {
		return run, ErrInvalidTransition
	}
	if progress < run.Progress {
		progress = run.Progress
	}
	if progress > 99 {
		progress = 99
	}
	now := time.Now().UTC()
	run.Status = status
	run.Stage = stage
	run.Progress = progress
	run.UpdatedAt = now
	run.Diagnostics = append(run.Diagnostics, diagnostics...)
	if len(run.Stages) == 0 || run.Stages[len(run.Stages)-1].Name != stage {
		run.Stages = append(run.Stages, domain.GenerationStage{
			Name:       stage,
			Progress:   progress,
			OccurredAt: now,
		})
	}
	if err := m.store.SaveGenerationRun(ctx, run); err != nil {
		return domain.GenerationRun{}, err
	}
	m.broadcast(run)
	return run, nil
}

func (m *Manager) finish(
	ctx context.Context,
	runID, status, stage string,
	progress int,
	adventureID string,
	diagnostics ...string,
) (domain.GenerationRun, error) {
	m.stateMu.Lock()
	defer m.stateMu.Unlock()

	run, err := m.store.GetGenerationRun(ctx, runID)
	if err != nil {
		return domain.GenerationRun{}, err
	}
	if terminal(run.Status) {
		return run, nil
	}
	now := time.Now().UTC()
	if status == "completed" {
		progress = 100
	} else if progress < run.Progress {
		progress = run.Progress
	}
	run.Status = status
	run.Stage = stage
	run.Progress = progress
	run.AdventureID = adventureID
	run.UpdatedAt = now
	run.CompletedAt = &now
	run.Diagnostics = append(run.Diagnostics, diagnostics...)
	run.Stages = append(run.Stages, domain.GenerationStage{
		Name:       stage,
		Progress:   progress,
		OccurredAt: now,
	})
	if err := m.store.SaveGenerationRun(ctx, run); err != nil {
		return domain.GenerationRun{}, err
	}
	m.broadcast(run)
	return run, nil
}

func (m *Manager) broadcast(run domain.GenerationRun) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for stream := range m.streams[run.ID] {
		select {
		case stream <- run:
		default:
			select {
			case <-stream:
			default:
			}
			select {
			case stream <- run:
			default:
			}
		}
	}
}

func terminal(status string) bool {
	return status == "completed" || status == "failed" || status == "cancelled"
}
