// DDivination API client

import type { BuilderOptions, Dungeon, DungeonConfig, DungeonListItem, DungeonUpdate, SyncStatus, TacticalRoomLayout } from "./types";

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

async function downloadExport(path: string, dungeon: Dungeon, fallbackExt: string): Promise<void> {
  const response = await fetch(`${API_URL}${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(dungeon),
  });
  if (!response.ok) {
    const body = await response.text();
    throw new Error(`API error ${response.status}: ${body}`);
  }

  const disposition = response.headers.get("Content-Disposition") || "";
  const utfMatch = disposition.match(/filename\*=UTF-8''([^;]+)/i);
  const plainMatch = disposition.match(/filename="([^"]+)"/i);
  let filename: string;
  if (utfMatch) {
    try { filename = decodeURIComponent(utfMatch[1]); } catch { filename = plainMatch?.[1] ?? `dungeon.${fallbackExt}`; }
  } else {
    filename = plainMatch?.[1] ?? `dungeon.${fallbackExt}`;
  }

  const blob = await response.blob();
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  link.remove();
  setTimeout(() => URL.revokeObjectURL(url), 1000);
}

export async function exportDungeonPdf(dungeon: Dungeon): Promise<void> {
  return downloadExport("/api/export/pdf", dungeon, "pdf");
}

export async function exportDungeonMarkdown(dungeon: Dungeon): Promise<void> {
  return downloadExport("/api/export/markdown", dungeon, "md");
}

// ── History / persistence ─────────────────────────────────────────────

export async function listDungeons(opts?: { favoritesOnly?: boolean; limit?: number; offset?: number }): Promise<DungeonListItem[]> {
  const q = new URLSearchParams();
  if (opts?.favoritesOnly) q.set("favorites_only", "true");
  if (opts?.limit != null) q.set("limit", String(opts.limit));
  if (opts?.offset != null) q.set("offset", String(opts.offset));
  const suffix = q.toString() ? `?${q}` : "";
  return fetchJSON<DungeonListItem[]>(`/api/dungeons${suffix}`);
}

export async function getStoredDungeon(id: string): Promise<Dungeon> {
  return fetchJSON<Dungeon>(`/api/dungeons/${encodeURIComponent(id)}`);
}

export async function updateStoredDungeon(id: string, patch: DungeonUpdate): Promise<Dungeon> {
  return fetchJSON<Dungeon>(`/api/dungeons/${encodeURIComponent(id)}`, {
    method: "PATCH",
    body: JSON.stringify(patch),
  });
}

export async function deleteStoredDungeon(id: string): Promise<void> {
  const response = await fetch(`${API_URL}/api/dungeons/${encodeURIComponent(id)}`, {
    method: "DELETE",
  });
  if (!response.ok && response.status !== 204) {
    throw new Error(`API error ${response.status}: ${await response.text()}`);
  }
}

export function buildDungeonPermalink(id: string): string {
  const url = new URL(window.location.href);
  url.searchParams.set("id", id);
  url.hash = "";
  return url.toString();
}
