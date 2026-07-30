import type { FloorMap, SessionState } from "./types";

export const TELEMETRY_SCHEMA_VERSION = "vtt-telemetry/v1";
export const LONG_FRAME_THRESHOLD_MS = 50;
export const FRAME_SAMPLE_CAPACITY = 240;

export interface FrameTelemetry {
  fps: number;
  averageFrameMs: number;
  p95FrameMs: number;
  longFrames: number;
  sampleCount: number;
  sampleWindowMs: number;
}

export interface RendererTelemetry {
  drawCalls: number;
  triangles: number;
  points: number;
  lines: number;
  geometries: number;
  textures: number;
}

export interface RenderTelemetry {
  frames: FrameTelemetry;
  renderer: RendererTelemetry;
}

export interface SceneTelemetry {
  width: number;
  height: number;
  tiles: number;
  walls: number;
  rooms: number;
  portals: number;
  entities: number;
  props: number;
  tokens: number;
  fogCells: number;
}

export type ConnectionStatus = "idle" | "connecting" | "open" | "reconnecting" | "closed" | "error";

export interface ConnectionTelemetry {
  status: ConnectionStatus;
  lastRevision: number;
  commandsSent: number;
  eventsReceived: number;
  snapshotsReceived: number;
  rejectedCommands: number;
  reconnectAttempts: number;
  lastEventLatencyMs: number | null;
}

export type ConnectionTelemetryEvent =
  | { type: "connect" }
  | { type: "reconnect" }
  | { type: "open" }
  | { type: "closed" }
  | { type: "error" }
  | { type: "command-sent" }
  | { type: "snapshot"; revision: number }
  | { type: "event"; revision: number; latencyMs: number }
  | { type: "rejected" };

export interface VTTTelemetryReport {
  schemaVersion: typeof TELEMETRY_SCHEMA_VERSION;
  capturedAt: string;
  viewport: {
    width: number;
    height: number;
    devicePixelRatio: number;
  };
  frames: FrameTelemetry;
  renderer: RendererTelemetry;
  scene: SceneTelemetry;
  connection: ConnectionTelemetry;
}

declare global {
  interface Window {
    __DDIVINATION_TELEMETRY__?: VTTTelemetryReport;
  }
}

const EMPTY_FRAMES: FrameTelemetry = {
  fps: 0,
  averageFrameMs: 0,
  p95FrameMs: 0,
  longFrames: 0,
  sampleCount: 0,
  sampleWindowMs: 0,
};

const EMPTY_RENDERER: RendererTelemetry = {
  drawCalls: 0,
  triangles: 0,
  points: 0,
  lines: 0,
  geometries: 0,
  textures: 0,
};

export function emptyRenderTelemetry(): RenderTelemetry {
  return {
    frames: { ...EMPTY_FRAMES },
    renderer: { ...EMPTY_RENDERER },
  };
}

export class FrameSampler {
  private readonly samples: number[] = [];

  constructor(private readonly capacity = FRAME_SAMPLE_CAPACITY) {
    if (!Number.isInteger(capacity) || capacity < 1) {
      throw new Error("frame sample capacity must be a positive integer");
    }
  }

  record(frameMs: number): void {
    if (!Number.isFinite(frameMs) || frameMs <= 0 || frameMs > 1000) return;
    this.samples.push(frameMs);
    if (this.samples.length > this.capacity) {
      this.samples.splice(0, this.samples.length - this.capacity);
    }
  }

  clear(): void {
    this.samples.length = 0;
  }

  snapshot(): FrameTelemetry {
    if (this.samples.length === 0) return { ...EMPTY_FRAMES };
    const total = this.samples.reduce((sum, value) => sum + value, 0);
    const average = total / this.samples.length;
    return {
      fps: round(1000 / average),
      averageFrameMs: round(average),
      p95FrameMs: round(percentile(this.samples, 0.95)),
      longFrames: this.samples.filter((value) => value > LONG_FRAME_THRESHOLD_MS).length,
      sampleCount: this.samples.length,
      sampleWindowMs: round(total),
    };
  }
}

export function percentile(values: readonly number[], quantile: number): number {
  if (values.length === 0) return 0;
  if (!Number.isFinite(quantile) || quantile < 0 || quantile > 1) {
    throw new Error("quantile must be between 0 and 1");
  }
  const sorted = [...values].sort((left, right) => left - right);
  const index = Math.max(0, Math.ceil(quantile * sorted.length) - 1);
  return sorted[index];
}

export function rendererTelemetry(info: {
  render: { calls: number; triangles: number; points: number; lines: number };
  memory: { geometries: number; textures: number };
}): RendererTelemetry {
  return {
    drawCalls: info.render.calls,
    triangles: info.render.triangles,
    points: info.render.points,
    lines: info.render.lines,
    geometries: info.memory.geometries,
    textures: info.memory.textures,
  };
}

export function sceneTelemetry(
  floor: FloorMap,
  session: SessionState | null,
  role: "gm" | "player" | "display",
): SceneTelemetry {
  const visibleEntities = floor.entities.filter((entity) => !entity.hidden || role === "gm");
  const tokens = visibleEntities.filter((entity) => entity.kind === "token" || entity.kind === "boss");
  const revealed = session?.revealedCells[floor.id] ?? floor.tiles;
  const revealedKeys = new Set(revealed.map((cell) => `${cell.x}:${cell.z}`));
  const fogCells = session
    ? floor.tiles.filter((tile) => !revealedKeys.has(`${tile.x}:${tile.z}`)).length
    : 0;

  return {
    width: floor.width,
    height: floor.height,
    tiles: floor.tiles.length,
    walls: floor.walls.length,
    rooms: floor.rooms.length,
    portals: floor.portals.length,
    entities: visibleEntities.length,
    props: visibleEntities.length - tokens.length,
    tokens: tokens.length,
    fogCells,
  };
}

export function createConnectionTelemetry(): ConnectionTelemetry {
  return {
    status: "idle",
    lastRevision: 0,
    commandsSent: 0,
    eventsReceived: 0,
    snapshotsReceived: 0,
    rejectedCommands: 0,
    reconnectAttempts: 0,
    lastEventLatencyMs: null,
  };
}

export function reduceConnectionTelemetry(
  current: ConnectionTelemetry,
  event: ConnectionTelemetryEvent,
): ConnectionTelemetry {
  switch (event.type) {
    case "connect":
      return { ...current, status: "connecting" };
    case "reconnect":
      return {
        ...current,
        status: "reconnecting",
        reconnectAttempts: current.reconnectAttempts + 1,
      };
    case "open":
      return { ...current, status: "open" };
    case "closed":
      return { ...current, status: "closed" };
    case "error":
      return { ...current, status: "error" };
    case "command-sent":
      return { ...current, commandsSent: current.commandsSent + 1 };
    case "snapshot":
      return {
        ...current,
        snapshotsReceived: current.snapshotsReceived + 1,
        lastRevision: event.revision,
      };
    case "event":
      return {
        ...current,
        eventsReceived: current.eventsReceived + 1,
        lastRevision: event.revision,
        lastEventLatencyMs: Math.max(0, round(event.latencyMs)),
      };
    case "rejected":
      return {
        ...current,
        rejectedCommands: current.rejectedCommands + 1,
      };
  }
}

export function createTelemetryReport(input: {
  render: RenderTelemetry;
  scene: SceneTelemetry;
  connection: ConnectionTelemetry;
  capturedAt?: Date;
  viewport?: { width: number; height: number; devicePixelRatio: number };
}): VTTTelemetryReport {
  const viewport = input.viewport ?? {
    width: window.innerWidth,
    height: window.innerHeight,
    devicePixelRatio: window.devicePixelRatio,
  };
  return {
    schemaVersion: TELEMETRY_SCHEMA_VERSION,
    capturedAt: (input.capturedAt ?? new Date()).toISOString(),
    viewport: {
      width: viewport.width,
      height: viewport.height,
      devicePixelRatio: round(viewport.devicePixelRatio),
    },
    frames: { ...input.render.frames },
    renderer: { ...input.render.renderer },
    scene: { ...input.scene },
    connection: { ...input.connection },
  };
}

export function downloadTelemetryReport(report: VTTTelemetryReport): void {
  const blob = new Blob([`${JSON.stringify(report, null, 2)}\n`], {
    type: "application/json",
  });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = `ddivination-vtt-telemetry-${report.capturedAt.replaceAll(":", "-")}.json`;
  link.click();
  URL.revokeObjectURL(url);
}

function round(value: number): number {
  return Math.round(value * 100) / 100;
}
