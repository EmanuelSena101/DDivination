CREATE TABLE IF NOT EXISTS session_credentials (
    session_id TEXT NOT NULL,
    participant_id TEXT NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    created_at TEXT NOT NULL,
    PRIMARY KEY(session_id, participant_id)
);

CREATE INDEX IF NOT EXISTS idx_session_credentials_session
    ON session_credentials(session_id);
