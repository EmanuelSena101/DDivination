"""History API: list / fetch / update / delete saved dungeons."""
from __future__ import annotations

from fastapi import APIRouter, HTTPException, Query

from app.data import dungeon_repo
from app.models import Dungeon, DungeonListItem, DungeonUpdate

router = APIRouter(prefix="/api/dungeons", tags=["history"])


@router.get("", response_model=list[DungeonListItem])
async def api_list_dungeons(
    favorites_only: bool = Query(False),
    limit: int = Query(50, ge=1, le=200),
    offset: int = Query(0, ge=0),
) -> list[DungeonListItem]:
    return dungeon_repo.list_dungeons(
        favorites_only=favorites_only, limit=limit, offset=offset
    )


@router.get("/{dungeon_id}", response_model=Dungeon)
async def api_get_dungeon(dungeon_id: str) -> Dungeon:
    d = dungeon_repo.get_dungeon(dungeon_id)
    if d is None:
        raise HTTPException(status_code=404, detail="Dungeon not found")
    return d


@router.patch("/{dungeon_id}", response_model=Dungeon)
async def api_update_dungeon(dungeon_id: str, patch: DungeonUpdate) -> Dungeon:
    d = dungeon_repo.update_dungeon(
        dungeon_id,
        favorite=patch.favorite,
        notes=patch.notes,
        name=patch.name,
    )
    if d is None:
        raise HTTPException(status_code=404, detail="Dungeon not found")
    return d


@router.delete("/{dungeon_id}", status_code=204)
async def api_delete_dungeon(dungeon_id: str) -> None:
    if not dungeon_repo.delete_dungeon(dungeon_id):
        raise HTTPException(status_code=404, detail="Dungeon not found")
