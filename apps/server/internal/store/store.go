package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/EmanuelSena101/DDivination/apps/server/internal/domain"
	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

var (
	ErrNotFound         = errors.New("not found")
	ErrConflict         = errors.New("version conflict")
	ErrRevisionConflict = errors.New("session revision conflict")
	ErrMigrationDrift   = errors.New("migration checksum mismatch")
	ErrCorruptSession   = errors.New("corrupt durable session")
)

type Store struct {
	db                     *sql.DB
	sessionReplayRetention int64
}

type AdventureSummary struct {
	ID        string               `json:"id"`
	Version   int64                `json:"version"`
	Name      domain.LocalizedText `json:"name"`
	Seed      uint64               `json:"seed"`
	UpdatedAt time.Time            `json:"updatedAt"`
}

func Open(databaseURL string) (*Store, error) {
	if strings.TrimSpace(databaseURL) == "" {
		return nil, errors.New("DATABASE_URL is required")
	}
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	db.SetMaxOpenConns(envInt("DDIVINATION_DB_MAX_CONNS", 10))
	db.SetMaxIdleConns(envInt("DDIVINATION_DB_MAX_IDLE_CONNS", 5))
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(30 * time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	s := &Store{
		db:                     db,
		sessionReplayRetention: int64(envInt("DDIVINATION_SESSION_EVENT_RETENTION", 500)),
	}
	if err := s.migrate(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	closedSessionRetention := envDuration("DDIVINATION_CLOSED_SESSION_RETENTION", 24*time.Hour)
	if closedSessionRetention > 0 {
		if _, err := s.PruneClosedSessions(ctx, time.Now().UTC().Add(-closedSessionRetention)); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("prune expired sessions: %w", err)
		}
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate(ctx context.Context) error {
	connection, err := s.db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("acquire migration connection: %w", err)
	}
	defer connection.Close()
	if _, err := connection.ExecContext(ctx, `SELECT pg_advisory_lock(hashtext('ddivination-schema-migrations'))`); err != nil {
		return fmt.Errorf("lock migrations: %w", err)
	}
	defer connection.ExecContext(context.Background(), `SELECT pg_advisory_unlock(hashtext('ddivination-schema-migrations'))`)
	if _, err := connection.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, checksum TEXT NOT NULL, applied_at TIMESTAMPTZ NOT NULL)`); err != nil {
		return fmt.Errorf("bootstrap migrations: %w", err)
	}
	entries, err := fs.ReadDir(migrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		content, err := migrationFiles.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return err
		}
		checksumBytes := sha256.Sum256(content)
		checksum := hex.EncodeToString(checksumBytes[:])
		var appliedChecksum string
		err = connection.QueryRowContext(ctx, `SELECT checksum FROM schema_migrations WHERE version = $1`, entry.Name()).Scan(&appliedChecksum)
		if err == nil {
			if appliedChecksum != checksum {
				return fmt.Errorf("%w: %s", ErrMigrationDrift, entry.Name())
			}
			continue
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		tx, err := connection.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, string(content)); err == nil {
			_, err = tx.ExecContext(ctx, `INSERT INTO schema_migrations(version, checksum, applied_at) VALUES($1, $2, $3)`, entry.Name(), checksum, time.Now().UTC())
		}
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("apply migration %s: %w", entry.Name(), err)
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

func envInt(name string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(os.Getenv(name)))
	if err != nil || value < 1 {
		return fallback
	}
	return value
}

func envDuration(name string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}
	if raw == "0" {
		return 0
	}
	value, err := time.ParseDuration(raw)
	if err != nil || value < 0 {
		return fallback
	}
	return value
}

func (s *Store) SaveAdventure(ctx context.Context, doc domain.AdventureDocument, expectedVersion *int64, reason string) error {
	if err := domain.ValidateAdventure(doc); err != nil {
		return err
	}
	payload, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("encode adventure: %w", err)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if expectedVersion != nil {
		var current int64
		if err := tx.QueryRowContext(ctx, `SELECT version FROM adventures WHERE id = $1`, doc.ID).Scan(&current); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrNotFound
			}
			return err
		}
		if current != *expectedVersion {
			return ErrConflict
		}
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO adventures(id, version, name_pt, name_en, seed, generator_version, document_json, created_at, updated_at)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT(id) DO UPDATE SET
			version=EXCLUDED.version,
			name_pt=EXCLUDED.name_pt,
			name_en=EXCLUDED.name_en,
			seed=EXCLUDED.seed,
			generator_version=EXCLUDED.generator_version,
			document_json=EXCLUDED.document_json,
			updated_at=EXCLUDED.updated_at
	`, doc.ID, doc.Version, doc.Name.PTBR, doc.Name.ENUS, strconv.FormatUint(doc.Seed, 10), doc.GeneratorVersion, string(payload), doc.CreatedAt, doc.UpdatedAt)
	if err != nil {
		return fmt.Errorf("save adventure: %w", err)
	}
	if reason != "" {
		if _, err = insertAdventureSnapshot(ctx, tx, doc, payload, reason); err != nil {
			return fmt.Errorf("save snapshot: %w", err)
		}
	}
	return tx.Commit()
}

func (s *Store) GetAdventure(ctx context.Context, id string) (domain.AdventureDocument, error) {
	var payload []byte
	if err := s.db.QueryRowContext(ctx, `SELECT document_json FROM adventures WHERE id = $1`, id).Scan(&payload); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.AdventureDocument{}, ErrNotFound
		}
		return domain.AdventureDocument{}, err
	}
	var doc domain.AdventureDocument
	if err := json.Unmarshal(payload, &doc); err != nil {
		return domain.AdventureDocument{}, fmt.Errorf("decode adventure: %w", err)
	}
	return doc, nil
}

func (s *Store) ListAdventures(ctx context.Context, limit int) ([]AdventureSummary, error) {
	if limit < 1 || limit > 200 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, version, name_pt, name_en, seed, updated_at
		FROM adventures ORDER BY updated_at DESC LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]AdventureSummary, 0)
	for rows.Next() {
		var item AdventureSummary
		var seed string
		if err := rows.Scan(&item.ID, &item.Version, &item.Name.PTBR, &item.Name.ENUS, &seed, &item.UpdatedAt); err != nil {
			return nil, err
		}
		item.Seed, _ = strconv.ParseUint(seed, 10, 64)
		result = append(result, item)
	}
	return result, rows.Err()
}

func (s *Store) CreateAdventureSnapshot(ctx context.Context, adventureID, reason string) (domain.AdventureSnapshotSummary, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.AdventureSnapshotSummary{}, err
	}
	defer tx.Rollback()
	var payload []byte
	if err := tx.QueryRowContext(ctx, `SELECT document_json FROM adventures WHERE id = $1`, adventureID).Scan(&payload); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.AdventureSnapshotSummary{}, ErrNotFound
		}
		return domain.AdventureSnapshotSummary{}, err
	}
	var document domain.AdventureDocument
	if err := json.Unmarshal(payload, &document); err != nil {
		return domain.AdventureSnapshotSummary{}, fmt.Errorf("decode adventure snapshot: %w", err)
	}
	snapshot, err := insertAdventureSnapshot(ctx, tx, document, payload, reason)
	if err != nil {
		return domain.AdventureSnapshotSummary{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.AdventureSnapshotSummary{}, err
	}
	return snapshot, nil
}

func (s *Store) ListAdventureSnapshots(ctx context.Context, adventureID string, limit int) ([]domain.AdventureSnapshotSummary, error) {
	if limit < 1 || limit > 200 {
		limit = 50
	}
	var exists int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM adventures WHERE id = $1`, adventureID).Scan(&exists); err != nil {
		return nil, err
	}
	if exists == 0 {
		return nil, ErrNotFound
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, version, reason, document_json, created_at
		FROM adventure_snapshots
		WHERE adventure_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, adventureID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]domain.AdventureSnapshotSummary, 0)
	for rows.Next() {
		var summary domain.AdventureSnapshotSummary
		var payload []byte
		if err := rows.Scan(&summary.ID, &summary.Version, &summary.Reason, &payload, &summary.CreatedAt); err != nil {
			return nil, err
		}
		var document domain.AdventureDocument
		if err := json.Unmarshal(payload, &document); err != nil {
			return nil, fmt.Errorf("decode adventure snapshot: %w", err)
		}
		summary.AdventureID = adventureID
		summary.Name = document.Name
		result = append(result, summary)
	}
	return result, rows.Err()
}

func (s *Store) GetAdventureSnapshot(ctx context.Context, adventureID, snapshotID string) (domain.AdventureSnapshot, error) {
	var snapshot domain.AdventureSnapshot
	var payload []byte
	if err := s.db.QueryRowContext(ctx, `
		SELECT id, version, reason, document_json, created_at
		FROM adventure_snapshots
		WHERE adventure_id = $1 AND id = $2
	`, adventureID, snapshotID).Scan(
		&snapshot.ID,
		&snapshot.Version,
		&snapshot.Reason,
		&payload,
		&snapshot.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.AdventureSnapshot{}, ErrNotFound
		}
		return domain.AdventureSnapshot{}, err
	}
	if err := json.Unmarshal(payload, &snapshot.Document); err != nil {
		return domain.AdventureSnapshot{}, fmt.Errorf("decode adventure snapshot: %w", err)
	}
	snapshot.AdventureID = adventureID
	snapshot.Name = snapshot.Document.Name
	return snapshot, nil
}

func (s *Store) DeleteAdventure(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM adventures WHERE id = $1`, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func insertAdventureSnapshot(
	ctx context.Context,
	tx *sql.Tx,
	document domain.AdventureDocument,
	payload []byte,
	reason string,
) (domain.AdventureSnapshotSummary, error) {
	createdAt := time.Now().UTC()
	snapshot := domain.AdventureSnapshotSummary{
		ID:          fmt.Sprintf("%s-v%d-%d", document.ID, document.Version, createdAt.UnixNano()),
		AdventureID: document.ID,
		Version:     document.Version,
		Reason:      reason,
		Name:        document.Name,
		CreatedAt:   createdAt,
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO adventure_snapshots(id, adventure_id, version, reason, document_json, created_at)
		VALUES($1, $2, $3, $4, $5, $6)
	`, snapshot.ID, document.ID, document.Version, reason, string(payload), createdAt)
	return snapshot, err
}

func (s *Store) SaveGenerationRun(ctx context.Context, run domain.GenerationRun) error {
	payload, err := json.Marshal(run)
	if err != nil {
		return err
	}
	var completed any
	if run.CompletedAt != nil {
		completed = *run.CompletedAt
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO generation_runs(id, status, run_json, created_at, completed_at)
		VALUES($1, $2, $3, $4, $5)
		ON CONFLICT(id) DO UPDATE SET status=EXCLUDED.status, run_json=EXCLUDED.run_json, completed_at=EXCLUDED.completed_at
	`, run.ID, run.Status, string(payload), run.CreatedAt, completed)
	return err
}

func (s *Store) GetGenerationRun(ctx context.Context, id string) (domain.GenerationRun, error) {
	var payload []byte
	if err := s.db.QueryRowContext(ctx, `SELECT run_json FROM generation_runs WHERE id = $1`, id).Scan(&payload); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.GenerationRun{}, ErrNotFound
		}
		return domain.GenerationRun{}, err
	}
	var run domain.GenerationRun
	if err := json.Unmarshal(payload, &run); err != nil {
		return domain.GenerationRun{}, err
	}
	return run, nil
}

func (s *Store) ListGenerationRuns(ctx context.Context, limit int) ([]domain.GenerationRun, error) {
	if limit < 1 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT run_json
		FROM generation_runs
		ORDER BY created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	runs := make([]domain.GenerationRun, 0)
	for rows.Next() {
		var payload []byte
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}
		var run domain.GenerationRun
		if err := json.Unmarshal(payload, &run); err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	return runs, rows.Err()
}

const (
	sessionSnapshotInterval int64 = 100
)

func (s *Store) InitializeSession(ctx context.Context, initialization SessionInitialization) error {
	state := initialization.State
	payload, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("encode initial session: %w", err)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	now := time.Now().UTC()
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO session_heads(session_id, adventure_id, revision, state_json, join_code_hash, join_code_expires_at, updated_at)
		VALUES($1, $2, $3, $4, $5, $6, $7)
	`, state.ID, state.AdventureID, state.Revision, string(payload), initialization.JoinCodeHash, initialization.JoinCodeExpiresAt, now); err != nil {
		return fmt.Errorf("initialize session head: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO session_snapshots(session_id, revision, state_json, created_at)
		VALUES($1, $2, $3, $4)
	`, state.ID, state.Revision, string(payload), now); err != nil {
		return fmt.Errorf("initialize session snapshot: %w", err)
	}
	if initialization.CredentialParticipantID != "" && initialization.CredentialToken != "" {
		hash := sha256.Sum256([]byte(initialization.CredentialToken))
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO session_credentials(session_id, participant_id, token_hash, created_at)
			VALUES($1, $2, $3, $4)
		`, state.ID, initialization.CredentialParticipantID, hex.EncodeToString(hash[:]), now); err != nil {
			return fmt.Errorf("initialize session credential: %w", err)
		}
	}
	return tx.Commit()
}

func (s *Store) CommitSession(ctx context.Context, commit SessionCommit) (SessionCommitResult, error) {
	if commit.SessionID == "" || commit.State.ID != commit.SessionID {
		return SessionCommitResult{}, errors.New("invalid session commit")
	}
	eventPayload, err := json.Marshal(commit.Event)
	if err != nil {
		return SessionCommitResult{}, fmt.Errorf("encode session event: %w", err)
	}
	statePayload, err := json.Marshal(commit.State)
	if err != nil {
		return SessionCommitResult{}, fmt.Errorf("encode session state: %w", err)
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return SessionCommitResult{}, err
	}
	defer tx.Rollback()

	if commit.CommandID != "" {
		if existing, found, err := sessionEventByCommand(ctx, tx, commit.SessionID, commit.CommandID); err != nil {
			return SessionCommitResult{}, err
		} else if found {
			return SessionCommitResult{Event: existing, Duplicate: true}, nil
		}
	}

	var currentRevision int64
	if err := tx.QueryRowContext(ctx, `SELECT revision FROM session_heads WHERE session_id = $1 FOR UPDATE`, commit.SessionID).Scan(&currentRevision); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SessionCommitResult{}, ErrNotFound
		}
		return SessionCommitResult{}, err
	}
	if commit.CommandID != "" {
		if existing, found, err := sessionEventByCommand(ctx, tx, commit.SessionID, commit.CommandID); err != nil {
			return SessionCommitResult{}, err
		} else if found {
			return SessionCommitResult{Event: existing, Duplicate: true}, nil
		}
	}
	if currentRevision != commit.ExpectedRevision || commit.Event.Revision != currentRevision+1 || commit.State.Revision != commit.Event.Revision {
		return SessionCommitResult{}, ErrRevisionConflict
	}

	var commandID any
	if commit.CommandID != "" {
		commandID = commit.CommandID
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO session_events(session_id, revision, command_id, event_json, created_at)
		VALUES($1, $2, $3, $4, $5)
	`, commit.SessionID, commit.Event.Revision, commandID, string(eventPayload), commit.Event.OccurredAt); err != nil {
		return SessionCommitResult{}, fmt.Errorf("append session event: %w", err)
	}
	if commit.CommandID != "" {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO session_commands(session_id, command_id, revision, event_json, created_at)
			VALUES($1, $2, $3, $4, $5)
		`, commit.SessionID, commit.CommandID, commit.Event.Revision, string(eventPayload), commit.Event.OccurredAt); err != nil {
			return SessionCommitResult{}, fmt.Errorf("record session command: %w", err)
		}
	}
	if commit.Access == nil {
		_, err = tx.ExecContext(ctx, `
			UPDATE session_heads SET revision=$2, state_json=$3, updated_at=$4 WHERE session_id=$1
		`, commit.SessionID, commit.State.Revision, string(statePayload), time.Now().UTC())
	} else {
		_, err = tx.ExecContext(ctx, `
			UPDATE session_heads
			SET revision=$2, state_json=$3, join_code_hash=$4, join_code_expires_at=$5, updated_at=$6
			WHERE session_id=$1
		`, commit.SessionID, commit.State.Revision, string(statePayload), commit.Access.CodeHash, commit.Access.ExpiresAt, time.Now().UTC())
	}
	if err != nil {
		return SessionCommitResult{}, fmt.Errorf("advance session head: %w", err)
	}
	if commit.ForceSnapshot || commit.Event.Revision%sessionSnapshotInterval == 0 {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO session_snapshots(session_id, revision, state_json, created_at)
			VALUES($1, $2, $3, $4)
			ON CONFLICT(session_id, revision) DO NOTHING
		`, commit.SessionID, commit.State.Revision, string(statePayload), time.Now().UTC()); err != nil {
			return SessionCommitResult{}, fmt.Errorf("checkpoint session: %w", err)
		}
		if commit.Event.Revision > s.sessionReplayRetention {
			if _, err := tx.ExecContext(ctx, `
				DELETE FROM session_events
				WHERE session_id=$1 AND revision <= $2
			`, commit.SessionID, commit.Event.Revision-s.sessionReplayRetention); err != nil {
				return SessionCommitResult{}, fmt.Errorf("compact session events: %w", err)
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return SessionCommitResult{}, err
	}
	return SessionCommitResult{Event: commit.Event}, nil
}

func sessionEventByCommand(ctx context.Context, query interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, sessionID, commandID string) (domain.SessionEvent, bool, error) {
	var payload []byte
	err := query.QueryRowContext(ctx, `
		SELECT event_json FROM session_commands WHERE session_id=$1 AND command_id=$2
	`, sessionID, commandID).Scan(&payload)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.SessionEvent{}, false, nil
	}
	if err != nil {
		return domain.SessionEvent{}, false, err
	}
	var event domain.SessionEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return domain.SessionEvent{}, false, fmt.Errorf("decode session event: %w", err)
	}
	return event, true, nil
}

func (s *Store) LoadSessionEventByCommand(ctx context.Context, sessionID, commandID string) (domain.SessionEvent, error) {
	event, found, err := sessionEventByCommand(ctx, s.db, sessionID, commandID)
	if err != nil {
		return domain.SessionEvent{}, err
	}
	if !found {
		return domain.SessionEvent{}, ErrNotFound
	}
	return event, nil
}

func (s *Store) LoadSession(ctx context.Context, sessionID string) (SessionRecord, error) {
	var record SessionRecord
	var payload []byte
	var revision int64
	var adventureID string
	if err := s.db.QueryRowContext(ctx, `
		SELECT revision, adventure_id, state_json, join_code_hash, join_code_expires_at
		FROM session_heads WHERE session_id=$1
	`, sessionID).Scan(&revision, &adventureID, &payload, &record.JoinCodeHash, &record.JoinCodeExpiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SessionRecord{}, ErrNotFound
		}
		return SessionRecord{}, err
	}
	state, err := decodeSessionHead(sessionID, adventureID, revision, payload)
	if err != nil {
		return SessionRecord{}, err
	}
	record.State = state
	return record, nil
}

func decodeSessionHead(sessionID, adventureID string, revision int64, payload []byte) (domain.SessionState, error) {
	var state domain.SessionState
	if err := json.Unmarshal(payload, &state); err != nil {
		return domain.SessionState{}, fmt.Errorf("%w: decode session head: %v", ErrCorruptSession, err)
	}
	if state.ID != sessionID || state.AdventureID != adventureID || state.Revision != revision {
		return domain.SessionState{}, fmt.Errorf(
			"%w: head columns and state disagree for %s",
			ErrCorruptSession,
			sessionID,
		)
	}
	return state, nil
}

func (s *Store) LoadSessionReplay(ctx context.Context, sessionID string, afterRevision int64, limit int) (SessionReplay, error) {
	if limit < 1 || limit > 1000 {
		limit = 500
	}
	var replay SessionReplay
	if err := s.db.QueryRowContext(ctx, `SELECT revision FROM session_heads WHERE session_id=$1`, sessionID).Scan(&replay.CurrentRevision); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SessionReplay{}, ErrNotFound
		}
		return SessionReplay{}, err
	}
	var oldest sql.NullInt64
	if err := s.db.QueryRowContext(ctx, `SELECT MIN(revision) FROM session_events WHERE session_id=$1`, sessionID).Scan(&oldest); err != nil {
		return SessionReplay{}, err
	}
	if oldest.Valid {
		replay.OldestRevision = oldest.Int64
	} else {
		replay.OldestRevision = replay.CurrentRevision + 1
	}
	if afterRevision < replay.OldestRevision-1 {
		return replay, nil
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT revision, event_json FROM session_events
		WHERE session_id=$1 AND revision>$2
		ORDER BY revision ASC LIMIT $3
	`, sessionID, afterRevision, limit+1)
	if err != nil {
		return SessionReplay{}, err
	}
	defer rows.Close()
	expectedRevision := afterRevision + 1
	for rows.Next() {
		var revision int64
		var payload []byte
		if err := rows.Scan(&revision, &payload); err != nil {
			return SessionReplay{}, err
		}
		var event domain.SessionEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			return SessionReplay{}, fmt.Errorf("%w: decode replay event: %v", ErrCorruptSession, err)
		}
		if revision != expectedRevision || event.Revision != revision {
			return SessionReplay{}, fmt.Errorf("%w: non-contiguous event log for %s", ErrCorruptSession, sessionID)
		}
		replay.Events = append(replay.Events, event)
		expectedRevision++
	}
	if err := rows.Err(); err != nil {
		return SessionReplay{}, err
	}
	if len(replay.Events) > limit {
		replay.Events = nil
		return replay, nil
	}
	lastRevision := afterRevision
	if len(replay.Events) > 0 {
		lastRevision = replay.Events[len(replay.Events)-1].Revision
	}
	replay.Complete = lastRevision == replay.CurrentRevision
	return replay, nil
}

func (s *Store) CompactSessionEvents(ctx context.Context, sessionID string, retainAfterRevision int64) (int64, error) {
	result, err := s.db.ExecContext(ctx, `
		DELETE FROM session_events
		WHERE session_id=$1 AND revision<$2
		  AND EXISTS (
			SELECT 1 FROM session_snapshots
			WHERE session_id=$1 AND revision >= $2
		  )
	`, sessionID, retainAfterRevision)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// PruneClosedSessions removes only sessions that were explicitly closed and
// have remained untouched before the cutoff. Open tables are never selected.
func (s *Store) PruneClosedSessions(ctx context.Context, cutoff time.Time) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	for _, table := range []string{"session_credentials", "session_events", "session_commands", "session_snapshots"} {
		if _, err := tx.ExecContext(ctx, fmt.Sprintf(`
			DELETE FROM %s WHERE session_id IN (
				SELECT session_id FROM session_heads
				WHERE state_json->>'open' = 'false' AND updated_at < $1
			)
		`, table), cutoff); err != nil {
			return 0, fmt.Errorf("prune %s: %w", table, err)
		}
	}
	result, err := tx.ExecContext(ctx, `
		DELETE FROM session_heads
		WHERE state_json->>'open' = 'false' AND updated_at < $1
	`, cutoff)
	if err != nil {
		return 0, fmt.Errorf("prune session heads: %w", err)
	}
	deleted, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return deleted, nil
}

func (s *Store) SaveSessionCredential(ctx context.Context, sessionID, participantID, token string) error {
	hash := sha256.Sum256([]byte(token))
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO session_credentials(session_id, participant_id, token_hash, created_at)
		VALUES($1, $2, $3, $4)
		ON CONFLICT(session_id, participant_id) DO UPDATE SET token_hash=EXCLUDED.token_hash
	`, sessionID, participantID, hex.EncodeToString(hash[:]), time.Now().UTC())
	return err
}

func (s *Store) LoadSessionCredentials(ctx context.Context, sessionID string) (map[string]string, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT token_hash, participant_id FROM session_credentials WHERE session_id = $1
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make(map[string]string)
	for rows.Next() {
		var hash, participantID string
		if err := rows.Scan(&hash, &participantID); err != nil {
			return nil, err
		}
		result["sha256:"+hash] = participantID
	}
	return result, rows.Err()
}

func (s *Store) DeleteSessionCredential(ctx context.Context, sessionID, participantID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM session_credentials WHERE session_id = $1 AND participant_id = $2`, sessionID, participantID)
	return err
}

func (s *Store) SaveAsset(ctx context.Context, ref domain.AssetRef, relativePath string) error {
	metadata, err := json.Marshal(ref)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO assets(id, sha256, metadata_json, relative_path, created_at)
		VALUES($1, $2, $3, $4, $5)
		ON CONFLICT(sha256) DO UPDATE SET
			metadata_json=EXCLUDED.metadata_json,
			relative_path=EXCLUDED.relative_path
	`, ref.ID, ref.SHA256, string(metadata), relativePath, time.Now().UTC())
	return err
}

func (s *Store) ListAssets(ctx context.Context) ([]domain.AssetRef, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT metadata_json FROM assets ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	assets := make([]domain.AssetRef, 0)
	for rows.Next() {
		var payload []byte
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}
		var ref domain.AssetRef
		if err := json.Unmarshal(payload, &ref); err != nil {
			return nil, err
		}
		assets = append(assets, ref)
	}
	return assets, rows.Err()
}
