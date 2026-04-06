"""Local JSON-based data store for synced D&D 5e API data."""
from __future__ import annotations

import json
import os
from datetime import datetime, timezone
from pathlib import Path
from typing import Optional

from app.models import Equipment, MagicItem, Monster, SyncStatus

DATA_DIR = Path(os.environ.get("DDIVINATION_DATA_DIR", Path(__file__).parent / "local"))


def _ensure_dir() -> None:
    DATA_DIR.mkdir(parents=True, exist_ok=True)


def _path(name: str) -> Path:
    return DATA_DIR / f"{name}.json"


def _read_json(name: str) -> list[dict]:
    p = _path(name)
    if not p.exists():
        return []
    with open(p, "r") as f:
        return json.load(f)


def _write_json(name: str, data: list[dict]) -> None:
    _ensure_dir()
    with open(_path(name), "w") as f:
        json.dump(data, f, indent=2)


# ── Monsters ──────────────────────────────────────────────────────────────

def save_monsters(monsters: list[dict]) -> None:
    _write_json("monsters", monsters)


def load_monsters_raw() -> list[dict]:
    return _read_json("monsters")


def load_monsters() -> list[Monster]:
    raw = load_monsters_raw()
    result = []
    for m in raw:
        try:
            result.append(Monster(**m))
        except Exception:
            continue
    return result


def get_monster_by_index(index: str) -> Optional[Monster]:
    for m in load_monsters():
        if m.index == index:
            return m
    return None


# ── Magic Items ───────────────────────────────────────────────────────────

def save_magic_items(items: list[dict]) -> None:
    _write_json("magic_items", items)


def load_magic_items_raw() -> list[dict]:
    return _read_json("magic_items")


def load_magic_items() -> list[MagicItem]:
    raw = load_magic_items_raw()
    result = []
    for m in raw:
        try:
            result.append(MagicItem(**m))
        except Exception:
            continue
    return result


# ── Equipment ─────────────────────────────────────────────────────────────

def save_equipment(items: list[dict]) -> None:
    _write_json("equipment", items)


def load_equipment_raw() -> list[dict]:
    return _read_json("equipment")


def load_equipment() -> list[Equipment]:
    raw = load_equipment_raw()
    result = []
    for e in raw:
        try:
            result.append(Equipment(**e))
        except Exception:
            continue
    return result


# ── Sync metadata ─────────────────────────────────────────────────────────

def save_sync_status(status: SyncStatus) -> None:
    _write_json("sync_status", [status.model_dump()])


def load_sync_status() -> SyncStatus:
    raw = _read_json("sync_status")
    if raw:
        return SyncStatus(**raw[0])
    return SyncStatus()


def mark_synced() -> None:
    monsters = load_monsters_raw()
    items = load_magic_items_raw()
    equip = load_equipment_raw()
    status = SyncStatus(
        monsters_count=len(monsters),
        magic_items_count=len(items),
        equipment_count=len(equip),
        last_sync=datetime.now(timezone.utc).isoformat(),
        is_synced=True,
    )
    save_sync_status(status)
