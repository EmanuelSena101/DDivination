CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    checksum TEXT NOT NULL,
    applied_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS adventures (
    id TEXT PRIMARY KEY,
    version BIGINT NOT NULL,
    name_pt TEXT NOT NULL,
    name_en TEXT NOT NULL,
    seed TEXT NOT NULL,
    generator_version TEXT NOT NULL,
    document_json JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_adventures_updated
    ON adventures(updated_at DESC);

CREATE TABLE IF NOT EXISTS adventure_snapshots (
    id TEXT PRIMARY KEY,
    adventure_id TEXT NOT NULL REFERENCES adventures(id) ON DELETE CASCADE,
    version BIGINT NOT NULL,
    reason TEXT NOT NULL,
    document_json JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_adventure_snapshots_adventure
    ON adventure_snapshots(adventure_id, created_at DESC);

CREATE TABLE IF NOT EXISTS generation_runs (
    id TEXT PRIMARY KEY,
    status TEXT NOT NULL,
    run_json JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS session_events (
    session_id TEXT NOT NULL,
    revision BIGINT NOT NULL,
    command_id TEXT,
    event_json JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY(session_id, revision)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_session_events_command
    ON session_events(session_id, command_id)
    WHERE command_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS session_commands (
    session_id TEXT NOT NULL,
    command_id TEXT NOT NULL,
    revision BIGINT NOT NULL,
    event_json JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY(session_id, command_id)
);

CREATE TABLE IF NOT EXISTS session_heads (
    session_id TEXT PRIMARY KEY,
    adventure_id TEXT NOT NULL REFERENCES adventures(id) ON DELETE CASCADE,
    revision BIGINT NOT NULL,
    state_json JSONB NOT NULL,
    join_code_hash TEXT NOT NULL,
    join_code_expires_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS session_snapshots (
    session_id TEXT NOT NULL REFERENCES session_heads(session_id) ON DELETE CASCADE,
    revision BIGINT NOT NULL,
    state_json JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY(session_id, revision)
);

CREATE TABLE IF NOT EXISTS assets (
    id TEXT PRIMARY KEY,
    sha256 TEXT NOT NULL UNIQUE,
    metadata_json JSONB NOT NULL,
    relative_path TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);
