"""Client for fetching data from the D&D 5e API and normalizing it locally."""
from __future__ import annotations

import httpx

from app.data.store import (
    load_sync_status,
    mark_synced,
    save_equipment,
    save_magic_items,
    save_monsters,
)
from app.enrichment.tagger import enrich_equipment, enrich_magic_item, enrich_monster
from app.models import SyncStatus

BASE_URL = "https://www.dnd5eapi.co/api/2024"
LEGACY_BASE_URL = "https://www.dnd5eapi.co/api"


async def _fetch_list(client: httpx.AsyncClient, endpoint: str, base: str = BASE_URL) -> list[dict]:
    """Fetch all items from a list endpoint, handling pagination."""
    items: list[dict] = []
    url = f"{base}{endpoint}"
    while url:
        resp = await client.get(url, timeout=30.0)
        resp.raise_for_status()
        data = resp.json()
        if "results" in data:
            items.extend(data["results"])
            url = data.get("next")
        elif isinstance(data, list):
            items.extend(data)
            url = None
        else:
            items.append(data)
            url = None
    return items


async def _fetch_detail(client: httpx.AsyncClient, url: str) -> dict:
    """Fetch a single resource detail."""
    if url.startswith("/"):
        url = f"https://www.dnd5eapi.co{url}"
    resp = await client.get(url, timeout=30.0)
    resp.raise_for_status()
    return resp.json()


def _normalize_monster(raw: dict) -> dict:
    """Normalize a raw monster from the API into our domain model shape."""
    ac = 10
    ac_field = raw.get("armor_class")
    if isinstance(ac_field, list) and ac_field:
        first = ac_field[0]
        if isinstance(first, dict):
            ac = first.get("value", 10)
        else:
            ac = int(first)
    elif isinstance(ac_field, (int, float)):
        ac = int(ac_field)

    cr = raw.get("challenge_rating", 0)
    if isinstance(cr, str):
        if "/" in cr:
            parts = cr.split("/")
            cr = float(parts[0]) / float(parts[1])
        else:
            cr = float(cr)

    xp = raw.get("xp", 0)
    if not xp:
        xp = _cr_to_xp(cr)

    return {
        "index": raw.get("index", ""),
        "name": raw.get("name", ""),
        "size": raw.get("size", "Medium"),
        "monster_type": raw.get("type", "beast"),
        "alignment": raw.get("alignment", "unaligned"),
        "challenge_rating": cr,
        "hit_points": raw.get("hit_points", 1),
        "armor_class": ac,
        "xp": xp,
    }


def _cr_to_xp(cr: float) -> int:
    """Convert CR to XP value."""
    cr_xp_table = {
        0: 10, 0.125: 25, 0.25: 50, 0.5: 100,
        1: 200, 2: 450, 3: 700, 4: 1100, 5: 1800,
        6: 2300, 7: 2900, 8: 3900, 9: 5000, 10: 5900,
        11: 7200, 12: 8400, 13: 10000, 14: 11500, 15: 13000,
        16: 15000, 17: 18000, 18: 20000, 19: 22000, 20: 25000,
        21: 33000, 22: 41000, 23: 50000, 24: 62000, 25: 75000,
        26: 90000, 27: 105000, 28: 120000, 29: 135000, 30: 155000,
    }
    return cr_xp_table.get(cr, int(cr * 1000))


def _normalize_magic_item(raw: dict) -> dict:
    """Normalize a raw magic item."""
    rarity = "common"
    rarity_field = raw.get("rarity")
    if isinstance(rarity_field, dict):
        rarity = rarity_field.get("name", "common").lower()
    elif isinstance(rarity_field, str):
        rarity = rarity_field.lower()

    category = "wondrous item"
    cat_field = raw.get("equipment_category")
    if isinstance(cat_field, dict):
        category = cat_field.get("name", "wondrous item").lower()
    elif isinstance(cat_field, str):
        category = cat_field.lower()

    desc_parts = raw.get("desc", [])
    if isinstance(desc_parts, list):
        desc = " ".join(str(d) for d in desc_parts[:2])
    else:
        desc = str(desc_parts)

    return {
        "index": raw.get("index", ""),
        "name": raw.get("name", ""),
        "rarity": rarity,
        "category": category,
        "description": desc[:500],
    }


def _normalize_equipment(raw: dict) -> dict:
    """Normalize a raw equipment item."""
    category = ""
    cat_field = raw.get("equipment_category")
    if isinstance(cat_field, dict):
        category = cat_field.get("name", "").lower()
    elif isinstance(cat_field, str):
        category = cat_field.lower()

    cost = raw.get("cost", {})
    cost_gp = 0.0
    if isinstance(cost, dict):
        quantity = cost.get("quantity", 0)
        unit = cost.get("unit", "gp")
        multipliers = {"cp": 0.01, "sp": 0.1, "ep": 0.5, "gp": 1.0, "pp": 10.0}
        cost_gp = quantity * multipliers.get(unit, 1.0)

    return {
        "index": raw.get("index", ""),
        "name": raw.get("name", ""),
        "category": category,
        "cost_gp": cost_gp,
        "weight": raw.get("weight", 0),
        "description": "",
    }


async def sync_all(progress_callback=None) -> SyncStatus:
    """Sync all data from D&D 5e API to local store."""
    async with httpx.AsyncClient() as client:
        # 1. Fetch monsters
        if progress_callback:
            progress_callback("Fetching monster list...")
        monster_list = await _fetch_list(client, "/monsters")

        normalized_monsters = []
        for i, entry in enumerate(monster_list):
            try:
                url = entry.get("url", "")
                if not url:
                    continue
                detail = await _fetch_detail(client, url)
                normalized = _normalize_monster(detail)
                enriched = enrich_monster(normalized)
                normalized_monsters.append(enriched)
            except Exception:
                continue

        save_monsters(normalized_monsters)
        if progress_callback:
            progress_callback(f"Synced {len(normalized_monsters)} monsters")

        # 2. Fetch magic items
        if progress_callback:
            progress_callback("Fetching magic items...")
        try:
            item_list = await _fetch_list(client, "/magic-items", base=LEGACY_BASE_URL)
            normalized_items = []
            for entry in item_list:
                try:
                    url = entry.get("url", "")
                    if not url:
                        continue
                    detail = await _fetch_detail(client, url)
                    normalized = _normalize_magic_item(detail)
                    enriched = enrich_magic_item(normalized)
                    normalized_items.append(enriched)
                except Exception:
                    continue
            save_magic_items(normalized_items)
            if progress_callback:
                progress_callback(f"Synced {len(normalized_items)} magic items")
        except Exception:
            save_magic_items([])

        # 3. Fetch equipment
        if progress_callback:
            progress_callback("Fetching equipment...")
        try:
            equip_list = await _fetch_list(client, "/equipment", base=LEGACY_BASE_URL)
            normalized_equip = []
            for entry in equip_list:
                try:
                    url = entry.get("url", "")
                    if not url:
                        continue
                    detail = await _fetch_detail(client, url)
                    normalized = _normalize_equipment(detail)
                    enriched = enrich_equipment(normalized)
                    normalized_equip.append(enriched)
                except Exception:
                    continue
            save_equipment(normalized_equip)
            if progress_callback:
                progress_callback(f"Synced {len(normalized_equip)} equipment items")
        except Exception:
            save_equipment([])

        mark_synced()
        return load_sync_status()
