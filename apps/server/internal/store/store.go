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
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/EmanuelSena101/DDivination/apps/server/internal/domain"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("version conflict")
)

type Store struct {
	db *sql.DB
}

type AdventureSummary struct {
	ID        string               `json:"id"`
	Version   int64                `json:"version"`
	Name      domain.LocalizedText `json:"name"`
	Seed      uint64               `json:"seed"`
	UpdatedAt time.Time            `json:"updatedAt"`
}

func Open(path string) (*Store, error) {
	if err := ensureParent(path); err != nil {
		return nil, err
	}
	dsn := "file:" + filepath.ToSlash(path) + "?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db}
	if err := s.migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, applied_at TEXT NOT NULL)`); err != nil {
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
		var exists int
		err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, entry.Name()).Scan(&exists)
		if err != nil {
			return err
		}
		if exists > 0 {
			continue
		}
		content, err := migrationFiles.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return err
		}
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, string(content)); err == nil {
			_, err = tx.ExecContext(ctx, `INSERT INTO schema_migrations(version, applied_at) VALUES(?, ?)`, entry.Name(), time.Now().UTC().Format(time.RFC3339Nano))
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
		if err := tx.QueryRowContext(ctx, `SELECT version FROM adventures WHERE id = ?`, doc.ID).Scan(&current); err != nil {
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
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			version=excluded.version,
			name_pt=excluded.name_pt,
			name_en=excluded.name_en,
			seed=excluded.seed,
			generator_version=excluded.generator_version,
			document_json=excluded.document_json,
			updated_at=excluded.updated_at
	`, doc.ID, doc.Version, doc.Name.PTBR, doc.Name.ENUS, strconv.FormatUint(doc.Seed, 10), doc.GeneratorVersion, payload, doc.CreatedAt.Format(time.RFC3339Nano), doc.UpdatedAt.Format(time.RFC3339Nano))
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
	if err := s.db.QueryRowContext(ctx, `SELECT document_json FROM adventures WHERE id = ?`, id).Scan(&payload); err != nil {
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
		FROM adventures ORDER BY updated_at DESC LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]AdventureSummary, 0)
	for rows.Next() {
		var item AdventureSummary
		var seed, updated string
		if err := rows.Scan(&item.ID, &item.Version, &item.Name.PTBR, &item.Name.ENUS, &seed, &updated); err != nil {
			return nil, err
		}
		item.Seed, _ = strconv.ParseUint(seed, 10, 64)
		item.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updated)
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
	if err := tx.QueryRowContext(ctx, `SELECT document_json FROM adventures WHERE id = ?`, adventureID).Scan(&payload); err != nil {
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
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM adventures WHERE id = ?`, adventureID).Scan(&exists); err != nil {
		return nil, err
	}
	if exists == 0 {
		return nil, ErrNotFound
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, version, reason, document_json, created_at
		FROM adventure_snapshots
		WHERE adventure_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`, adventureID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]domain.AdventureSnapshotSummary, 0)
	for rows.Next() {
		var summary domain.AdventureSnapshotSummary
		var payload []byte
		var created string
		if err := rows.Scan(&summary.ID, &summary.Version, &summary.Reason, &payload, &created); err != nil {
			return nil, err
		}
		var document domain.AdventureDocument
		if err := json.Unmarshal(payload, &document); err != nil {
			return nil, fmt.Errorf("decode adventure snapshot: %w", err)
		}
		summary.AdventureID = adventureID
		summary.Name = document.Name
		summary.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
		result = append(result, summary)
	}
	return result, rows.Err()
}

func (s *Store) GetAdventureSnapshot(ctx context.Context, adventureID, snapshotID string) (domain.AdventureSnapshot, error) {
	var snapshot domain.AdventureSnapshot
	var payload []byte
	var created string
	if err := s.db.QueryRowContext(ctx, `
		SELECT id, version, reason, document_json, created_at
		FROM adventure_snapshots
		WHERE adventure_id = ? AND id = ?
	`, adventureID, snapshotID).Scan(
		&snapshot.ID,
		&snapshot.Version,
		&snapshot.Reason,
		&payload,
		&created,
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
	snapshot.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
	return snapshot, nil
}

func (s *Store) DeleteAdventure(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM adventures WHERE id = ?`, id)
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
		VALUES(?, ?, ?, ?, ?, ?)
	`, snapshot.ID, document.ID, document.Version, reason, payload, createdAt.Format(time.RFC3339Nano))
	return snapshot, err
}

func (s *Store) SaveGenerationRun(ctx context.Context, run domain.GenerationRun) error {
	payload, err := json.Marshal(run)
	if err != nil {
		return err
	}
	var completed any
	if run.CompletedAt != nil {
		completed = run.CompletedAt.Format(time.RFC3339Nano)
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO generation_runs(id, status, run_json, created_at, completed_at)
		VALUES(?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET status=excluded.status, run_json=excluded.run_json, completed_at=excluded.completed_at
	`, run.ID, run.Status, payload, run.CreatedAt.Format(time.RFC3339Nano), completed)
	return err
}

func (s *Store) GetGenerationRun(ctx context.Context, id string) (domain.GenerationRun, error) {
	var payload []byte
	if err := s.db.QueryRowContext(ctx, `SELECT run_json FROM generation_runs WHERE id = ?`, id).Scan(&payload); err != nil {
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
		LIMIT ?
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

func (s *Store) SaveSessionEvent(ctx context.Context, sessionID string, event domain.SessionEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT OR IGNORE INTO session_events(session_id, revision, event_json, created_at)
		VALUES(?, ?, ?, ?)
	`, sessionID, event.Revision, payload, event.OccurredAt.Format(time.RFC3339Nano))
	return err
}

func (s *Store) SaveSessionSnapshot(ctx context.Context, state domain.SessionState) error {
	payload, err := json.Marshal(state)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO session_snapshots(session_id, revision, state_json, updated_at)
		VALUES(?, ?, ?, ?)
		ON CONFLICT(session_id) DO UPDATE SET revision=excluded.revision, state_json=excluded.state_json, updated_at=excluded.updated_at
	`, state.ID, state.Revision, payload, time.Now().UTC().Format(time.RFC3339Nano))
	return err
}

func (s *Store) LoadSessionSnapshot(ctx context.Context, sessionID string) (domain.SessionState, error) {
	var payload []byte
	if err := s.db.QueryRowContext(ctx, `SELECT state_json FROM session_snapshots WHERE session_id = ?`, sessionID).Scan(&payload); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.SessionState{}, ErrNotFound
		}
		return domain.SessionState{}, err
	}
	var state domain.SessionState
	if err := json.Unmarshal(payload, &state); err != nil {
		return domain.SessionState{}, err
	}
	return state, nil
}

func (s *Store) SaveSessionCredential(ctx context.Context, sessionID, participantID, token string) error {
	hash := sha256.Sum256([]byte(token))
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO session_credentials(session_id, participant_id, token_hash, created_at)
		VALUES(?, ?, ?, ?)
		ON CONFLICT(session_id, participant_id) DO UPDATE SET token_hash=excluded.token_hash
	`, sessionID, participantID, hex.EncodeToString(hash[:]), time.Now().UTC().Format(time.RFC3339Nano))
	return err
}

func (s *Store) LoadSessionCredentials(ctx context.Context, sessionID string) (map[string]string, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT token_hash, participant_id FROM session_credentials WHERE session_id = ?
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

func (s *Store) SaveAsset(ctx context.Context, ref domain.AssetRef, relativePath string) error {
	metadata, err := json.Marshal(ref)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO assets(id, sha256, metadata_json, relative_path, created_at)
		VALUES(?, ?, ?, ?, ?)
		ON CONFLICT(sha256) DO UPDATE SET
			metadata_json=excluded.metadata_json,
			relative_path=excluded.relative_path
	`, ref.ID, ref.SHA256, metadata, relativePath, time.Now().UTC().Format(time.RFC3339Nano))
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

func ensureParent(path string) error {
	parent := filepath.Dir(path)
	if parent == "." || parent == "" {
		return nil
	}
	return ensureDir(parent)
}
