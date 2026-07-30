import type {
  AdventureDocument,
  AdventureSnapshotSummary,
  AdventureSpec,
  CreatedSession,
  GenerationResult,
  JoinedSession,
} from "./types";

export class APIError extends Error {
  constructor(
    public readonly status: number,
    message: string,
  ) {
    super(message);
    this.name = "APIError";
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, {
    ...init,
    headers: {
      ...(init?.body instanceof FormData ? {} : { "Content-Type": "application/json" }),
      ...init?.headers,
    },
  });
  if (!response.ok) {
    const problem = await response.json().catch(() => null);
    throw new APIError(
      response.status,
      problem?.detail || problem?.title || `HTTP ${response.status}`,
    );
  }
  return response.json() as Promise<T>;
}

export function generateAdventure(spec: AdventureSpec, seed?: number): Promise<GenerationResult> {
  return request("/api/v1/generation-runs", {
    method: "POST",
    body: JSON.stringify({ spec, ...(seed == null ? {} : { seed }) }),
  });
}

export function getAdventure(id: string): Promise<AdventureDocument> {
  return request(`/api/v1/adventures/${encodeURIComponent(id)}`);
}

export function updateAdventure(document: AdventureDocument): Promise<AdventureDocument> {
  return request(`/api/v1/adventures/${encodeURIComponent(document.id)}`, {
    method: "PUT",
    headers: { "If-Match": `"${document.version}"` },
    body: JSON.stringify(document),
  });
}

export function listAdventureCheckpoints(
  adventureId: string,
): Promise<AdventureSnapshotSummary[]> {
  return request(`/api/v1/adventures/${encodeURIComponent(adventureId)}/checkpoints`);
}

export function createAdventureCheckpoint(
  adventureId: string,
): Promise<AdventureSnapshotSummary> {
  return request(`/api/v1/adventures/${encodeURIComponent(adventureId)}/checkpoints`, {
    method: "POST",
  });
}

export function restoreAdventureCheckpoint(
  adventureId: string,
  checkpointId: string,
  expectedVersion: number,
): Promise<AdventureDocument> {
  return request(
    `/api/v1/adventures/${encodeURIComponent(adventureId)}/checkpoints/${encodeURIComponent(checkpointId)}/restore`,
    {
      method: "POST",
      headers: { "If-Match": `"${expectedVersion}"` },
    },
  );
}

export function createSession(adventureId: string, gmName: string): Promise<CreatedSession> {
  return request("/api/v1/sessions", {
    method: "POST",
    body: JSON.stringify({ adventureId, gmName }),
  });
}

export function joinSession(sessionId: string, code: string, name: string, role: "player" | "display"): Promise<JoinedSession> {
  return request(`/api/v1/sessions/${encodeURIComponent(sessionId)}/join`, {
    method: "POST",
    body: JSON.stringify({ code, name, role }),
  });
}

export async function closeSession(sessionId: string, token: string): Promise<void> {
  const response = await fetch(`/api/v1/sessions/${encodeURIComponent(sessionId)}`, {
    method: "DELETE",
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!response.ok) {
    const problem = await response.json().catch(() => null);
    throw new Error(problem?.detail || problem?.title || `HTTP ${response.status}`);
  }
}

export function packageURL(adventureId: string): string {
  return `/api/v1/packages/${encodeURIComponent(adventureId)}`;
}

export function markdownURL(adventureId: string): string {
  return `/api/v1/adventures/${encodeURIComponent(adventureId)}/export.md`;
}

export function printURL(adventureId: string): string {
  return `/api/v1/adventures/${encodeURIComponent(adventureId)}/print`;
}

export function sessionWebSocketURL(sessionId: string, token: string): string {
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  return `${protocol}//${window.location.host}/api/v1/sessions/${encodeURIComponent(sessionId)}/stream?token=${encodeURIComponent(token)}`;
}
