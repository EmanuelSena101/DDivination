"""API routes for D&D 5e API data synchronization."""
from __future__ import annotations

from fastapi import APIRouter

from app.data.store import load_sync_status
from app.models import SyncStatus
from app.sync.dnd5e_client import sync_all

router = APIRouter(prefix="/api/sync", tags=["sync"])


@router.get("/status", response_model=SyncStatus)
async def api_sync_status() -> SyncStatus:
    """Get current sync status."""
    return load_sync_status()


@router.post("/start", response_model=SyncStatus)
async def api_start_sync() -> SyncStatus:
    """Start syncing data from D&D 5e API."""
    return await sync_all()
