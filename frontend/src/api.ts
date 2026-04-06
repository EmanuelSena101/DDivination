// DDivination API client

import type { BuilderOptions, Dungeon, DungeonConfig, SyncStatus, TacticalRoomLayout } from "./types";

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8000";

async function fetchJSON<T>(url: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`${API_URL}${url}`, {
    headers: { "Content-Type": "application/json" },
    ...options,
  });
  if (!response.ok) {
    const body = await response.text();
    throw new Error(`API error ${response.status}: ${body}`);
  }
  return response.json();
}

export async function getSyncStatus(): Promise<SyncStatus> {
  return fetchJSON<SyncStatus>("/api/sync/status");
}

export async function startSync(): Promise<SyncStatus> {
  return fetchJSON<SyncStatus>("/api/sync/start", { method: "POST" });
}

export async function generateDungeon(config: DungeonConfig): Promise<Dungeon> {
  return fetchJSON<Dungeon>("/api/dungeon/generate", {
    method: "POST",
    body: JSON.stringify(config),
  });
}

export async function getBuilderOptions(): Promise<BuilderOptions> {
  return fetchJSON<BuilderOptions>("/api/dungeon/options");
}

export async function getTacticalLayouts(config: DungeonConfig): Promise<TacticalRoomLayout[]> {
  return fetchJSON<TacticalRoomLayout[]>("/api/dungeon/tactical", {
    method: "POST",
    body: JSON.stringify(config),
  });
}
