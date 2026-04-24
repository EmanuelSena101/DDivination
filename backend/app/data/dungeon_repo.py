"""CRUD for persisted Dungeons.

Storage layout: one row per dungeon in `dungeons`. The full Dungeon is serialized
as JSON in `dungeon_json` so the user can always return to *exactly* what they saw
at generation time, even if the procedural generator evolves.

A handful of columns (name, seed, theme, biome, party_size, party_level,
summary, total_rooms, estimated_difficulty, favorite, created_at) are
denormalised for cheap list queries without JSON parsing.
"""
from __future__ import annotations

import json
import secrets
from datetime import datetime, timezone
from typing import Optional

from app.data.db import connection
from app.models import Dungeon, DungeonListItem

# Retention: keep all favorites forever, prune oldest non-favorites beyond this.
NON_FAVORITE_RETENTION = 100


def _new_id() -> str:
    # 10 URL-safe chars (~60 bits of entropy) — collisions are effectively impossible
    # at any scale this single-user GM tool will hit.
    return secrets.token_urlsafe(7)[:10]


def _now_iso() -> str:
    return datetime.now(timezone.utc).isoformat(timespec="seconds")


def save_dungeon(dungeon: Dungeon) -> Dungeon:
    """Persist a dungeon and return it with id/created_at populated.

    If the dungeon already has an id, behave as an upsert (same id is re-used).
    """
    dungeon_id = dungeon.id or _new_id()
    created_at = dungeon.created_at or _now_iso()

    stored = dungeon.model_copy(update={"id": dungeon_id, "created_at": created_at})

    with connection() as conn:
        conn.execute(
            """
            INSERT INTO dungeons
                (id, name, seed, config_json, dungeon_json, summary,
                 theme, biome, party_size, party_level, total_rooms,
                 estimated_difficulty, favorite, notes, created_at)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
            ON CONFLICT(id) DO UPDATE SET
                name = excluded.name,
                seed = excluded.seed,
                config_json = excluded.config_json,
                dungeon_json = excluded.dungeon_json,
                summary = excluded.summary,
                theme = excluded.theme,
                biome = excluded.biome,
                party_size = excluded.party_size,
                party_level = excluded.party_level,
                total_rooms = excluded.total_rooms,
                estimated_difficulty = excluded.estimated_difficulty,
                favorite = excluded.favorite,
                notes = COALESCE(excluded.notes, dungeons.notes)
            """,
            (
                dungeon_id,
                stored.name,
                stored.seed,
                stored.config.model_dump_json(),
                stored.model_dump_json(),
                stored.summary,
                stored.config.theme.value,
                stored.config.biome.value,
                stored.config.party_size,
                stored.config.party_level,
                stored.analysis.total_rooms if stored.analysis else len(stored.rooms),
                stored.analysis.estimated_difficulty if stored.analysis else None,
                1 if stored.favorite else 0,
                stored.notes,
                created_at,
            ),
        )

    _prune_non_favorites()
    return stored


def get_dungeon(dungeon_id: str) -> Optional[Dungeon]:
    with connection() as conn:
        row = conn.execute(
            "SELECT dungeon_json, favorite, notes, created_at FROM dungeons WHERE id = ?",
            (dungeon_id,),
        ).fetchone()
    if row is None:
        return None
    data = json.loads(row["dungeon_json"])
    # Re-hydrate flags from denormalised columns in case the stored JSON is older.
    data["id"] = dungeon_id
    data["favorite"] = bool(row["favorite"])
    data["notes"] = row["notes"]
    data["created_at"] = row["created_at"]
    return Dungeon.model_validate(data)


def list_dungeons(
    *, favorites_only: bool = False, limit: int = 50, offset: int = 0
) -> list[DungeonListItem]:
    where = "WHERE favorite = 1" if favorites_only else ""
    with connection() as conn:
        rows = conn.execute(
            f"""
            SELECT id, name, seed, created_at, favorite, theme, biome,
                   party_size, party_level, summary, estimated_difficulty, total_rooms
            FROM dungeons
            {where}
            ORDER BY favorite DESC, created_at DESC
            LIMIT ? OFFSET ?
            """,
            (limit, offset),
        ).fetchall()
    return [
        DungeonListItem(
            id=r["id"],
            name=r["name"],
            seed=r["seed"],
            created_at=r["created_at"],
            favorite=bool(r["favorite"]),
            theme=r["theme"],
            biome=r["biome"],
            party_size=r["party_size"],
            party_level=r["party_level"],
            summary=r["summary"] or "",
            estimated_difficulty=r["estimated_difficulty"],
            total_rooms=r["total_rooms"] or 0,
        )
        for r in rows
    ]


def update_dungeon(
    dungeon_id: str,
    *,
    favorite: Optional[bool] = None,
    notes: Optional[str] = None,
    name: Optional[str] = None,
) -> Optional[Dungeon]:
    sets: list[str] = []
    params: list[object] = []
    if favorite is not None:
        sets.append("favorite = ?")
        params.append(1 if favorite else 0)
    if notes is not None:
        sets.append("notes = ?")
        params.append(notes)
    if name is not None and name.strip():
        sets.append("name = ?")
        params.append(name.strip())
    if not sets:
        return get_dungeon(dungeon_id)

    params.append(dungeon_id)
    with connection() as conn:
        cur = conn.execute(
            f"UPDATE dungeons SET {', '.join(sets)} WHERE id = ?",
            params,
        )
        if cur.rowcount == 0:
            return None
    return get_dungeon(dungeon_id)


def delete_dungeon(dungeon_id: str) -> bool:
    with connection() as conn:
        cur = conn.execute("DELETE FROM dungeons WHERE id = ?", (dungeon_id,))
        return cur.rowcount > 0


def count_dungeons() -> int:
    with connection() as conn:
        row = conn.execute("SELECT COUNT(*) AS c FROM dungeons").fetchone()
    return int(row["c"])


def _prune_non_favorites(keep: int = NON_FAVORITE_RETENTION) -> int:
    """Drop oldest non-favorite dungeons beyond the retention cap.

    Favorites are never pruned. Returns number of rows deleted.
    """
    with connection() as conn:
        cur = conn.execute(
            """
            DELETE FROM dungeons
            WHERE favorite = 0
              AND id IN (
                  SELECT id FROM dungeons
                  WHERE favorite = 0
                  ORDER BY created_at DESC
                  LIMIT -1 OFFSET ?
              )
            """,
            (keep,),
        )
        return cur.rowcount
