CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    applied_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS adventures (
    id TEXT PRIMARY KEY,
    version INTEGER NOT NULL,
    name_pt TEXT NOT NULL,
    name_en TEXT NOT NULL,
    seed TEXT NOT NULL,
    generator_version TEXT NOT NULL,
    document_json BLOB NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_adventures_updated
    ON adventures(updated_at DESC);

CREATE TABLE IF NOT EXISTS adventure_snapshots (
    id TEXT PRIMARY KEY,
    adventure_id TEXT NOT NULL REFERENCES adventures(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    reason TEXT NOT NULL,
    document_json BLOB NOT NULL,
    created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_adventure_snapshots_adventure
    ON adventure_snapshots(adventure_id, created_at DESC);

CREATE TABLE IF NOT EXISTS generation_runs (
    id TEXT PRIMARY KEY,
    status TEXT NOT NULL,
    run_json BLOB NOT NULL,
    created_at TEXT NOT NULL,
    completed_at TEXT
);

CREATE TABLE IF NOT EXISTS session_events (
    session_id TEXT NOT NULL,
    revision INTEGER NOT NULL,
    event_json BLOB NOT NULL,
    created_at TEXT NOT NULL,
    PRIMARY KEY(session_id, revision)
);

CREATE TABLE IF NOT EXISTS session_snapshots (
    session_id TEXT PRIMARY KEY,
    revision INTEGER NOT NULL,
    state_json BLOB NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS assets (
    id TEXT PRIMARY KEY,
    sha256 TEXT NOT NULL UNIQUE,
    metadata_json BLOB NOT NULL,
    relative_path TEXT NOT NULL,
    created_at TEXT NOT NULL
);
