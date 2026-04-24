"""SQLite connection helper and schema bootstrap for persistent storage."""
from __future__ import annotations

import sqlite3
from contextlib import contextmanager
from pathlib import Path
from typing import Iterator

from app.data.store import DATA_DIR

_SCHEMA = """
CREATE TABLE IF NOT EXISTS dungeons (
    id              TEXT PRIMARY KEY,
    name            TEXT NOT NULL,
    seed            INTEGER NOT NULL,
    config_json     TEXT NOT NULL,
    dungeon_json    TEXT NOT NULL,
    summary         TEXT NOT NULL DEFAULT '',
    theme           TEXT NOT NULL DEFAULT '',
    biome           TEXT NOT NULL DEFAULT '',
    party_size      INTEGER NOT NULL DEFAULT 0,
    party_level     INTEGER NOT NULL DEFAULT 0,
    total_rooms     INTEGER NOT NULL DEFAULT 0,
    estimated_difficulty TEXT,
    favorite        INTEGER NOT NULL DEFAULT 0,
    notes           TEXT,
    created_at      TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_dungeons_created_at
    ON dungeons(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_dungeons_favorite_created
    ON dungeons(favorite DESC, created_at DESC);
"""


def db_path() -> Path:
    DATA_DIR.mkdir(parents=True, exist_ok=True)
    return DATA_DIR / "ddivination.sqlite3"


def init_db() -> None:
    """Create the schema if it doesn't exist. Idempotent — safe on every startup."""
    with connection() as conn:
        conn.executescript(_SCHEMA)


@contextmanager
def connection() -> Iterator[sqlite3.Connection]:
    """Context-managed SQLite connection with row_factory set and FK enforcement on."""
    conn = sqlite3.connect(str(db_path()), detect_types=sqlite3.PARSE_DECLTYPES)
    conn.row_factory = sqlite3.Row
    conn.execute("PRAGMA foreign_keys = ON")
    conn.execute("PRAGMA journal_mode = WAL")
    try:
        yield conn
        conn.commit()
    except Exception:
        conn.rollback()
        raise
    finally:
        conn.close()
