"""API routes for dungeon generation and analysis."""
from __future__ import annotations

from fastapi import APIRouter, HTTPException

from app.analysis.analyzer import analyze_dungeon
from app.data.store import load_sync_status
from app.generation.dungeon_generator import generate_dungeon
from app.models import Dungeon, DungeonConfig
from app.tactical.layout import (
    TacticalRoomLayout,
    generate_all_tactical_layouts,
    generate_tactical_layout,
)

router = APIRouter(prefix="/api/dungeon", tags=["dungeon"])


@router.post("/generate", response_model=Dungeon)
async def api_generate_dungeon(config: DungeonConfig) -> Dungeon:
    """Generate a dungeon from configuration."""
    status = load_sync_status()
    if not status.is_synced:
        raise HTTPException(
            status_code=400,
            detail="Data not yet synced. Please sync D&D 5e API data first via POST /api/sync/start",
        )

    dungeon = generate_dungeon(config)
    dungeon.analysis = analyze_dungeon(dungeon)
    return dungeon


@router.post("/quick-generate", response_model=Dungeon)
async def api_quick_generate(
    party_size: int = 4,
    party_level: int = 5,
) -> Dungeon:
    """Quick generation with minimal inputs."""
    status = load_sync_status()
    if not status.is_synced:
        raise HTTPException(
            status_code=400,
            detail="Data not yet synced. Please sync D&D 5e API data first via POST /api/sync/start",
        )

    config = DungeonConfig(party_size=party_size, party_level=party_level)
    dungeon = generate_dungeon(config)
    dungeon.analysis = analyze_dungeon(dungeon)
    return dungeon


@router.post("/tactical", response_model=list[TacticalRoomLayout])
async def api_tactical_layouts(config: DungeonConfig) -> list[TacticalRoomLayout]:
    """Generate tactical battle grid layouts for all rooms in a dungeon."""
    status = load_sync_status()
    if not status.is_synced:
        raise HTTPException(
            status_code=400,
            detail="Data not yet synced.",
        )

    dungeon = generate_dungeon(config)
    dungeon.analysis = analyze_dungeon(dungeon)
    return generate_all_tactical_layouts(dungeon)


@router.get("/options")
async def api_get_options() -> dict:
    """Get available options for the dungeon builder."""
    from app.models import (
        Biome,
        Difficulty,
        DungeonSize,
        StructureStyle,
        Theme,
        TrapDensity,
        TreasureQuality,
    )

    return {
        "themes": [{"value": t.value, "label": t.value.replace("_", " ").title()} for t in Theme],
        "biomes": [{"value": b.value, "label": b.value.replace("_", " ").title()} for b in Biome],
        "structure_styles": [{"value": s.value, "label": s.value.replace("_", " ").title()} for s in StructureStyle],
        "difficulties": [{"value": d.value, "label": d.value.replace("_", " ").title()} for d in Difficulty],
        "dungeon_sizes": [
            {"value": s.value, "label": s.value.replace("_", " ").title(),
             "description": {"small": "5-7 rooms", "medium": "8-12 rooms", "large": "13-18 rooms", "epic": "19-25 rooms"}.get(s.value, "")}
            for s in DungeonSize
        ],
        "trap_densities": [{"value": t.value, "label": t.value.replace("_", " ").title()} for t in TrapDensity],
        "treasure_qualities": [{"value": t.value, "label": t.value.replace("_", " ").title()} for t in TreasureQuality],
    }
