"""DDivination FastAPI application."""
from __future__ import annotations

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from app.routers.dungeon import router as dungeon_router
from app.routers.sync import router as sync_router

app = FastAPI(
    title="DDivination",
    description="A guided dungeon builder for D&D Game Masters",
    version="0.1.0",
)

# CORS - allow all origins for local development
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(sync_router)
app.include_router(dungeon_router)


@app.get("/api/health")
async def health() -> dict:
    """Health check endpoint."""
    return {"status": "ok", "service": "ddivination"}
