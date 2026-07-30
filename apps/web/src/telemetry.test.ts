import { describe, expect, it } from "vitest";
import {
  createConnectionTelemetry,
  createTelemetryReport,
  FrameSampler,
  percentile,
  reduceConnectionTelemetry,
  rendererTelemetry,
  sceneTelemetry,
} from "./telemetry";
import type { FloorMap, SessionState } from "./types";

const floor: FloorMap = {
  id: "floor-secret-id",
  index: 0,
  name: { "pt-BR": "Segredo editorial", "en-US": "Editorial secret" },
  width: 4,
  height: 3,
  tiles: [
    { x: 0, z: 0, kind: "floor", walkable: true },
    { x: 1, z: 0, kind: "floor", walkable: true },
    { x: 2, z: 0, kind: "floor", walkable: true },
  ],
  walls: [{ x: 0, z: 0, direction: "north", kind: "wall", open: false, locked: false }],
  rooms: [
    {
      id: "room-secret",
      name: { "pt-BR": "Sala secreta", "en-US": "Secret room" },
      role: "goal",
      description: { "pt-BR": "Narrativa privada", "en-US": "Private narrative" },
      center: { x: 1, z: 0 },
      secret: true,
      mandatory: false,
    },
  ],
  entities: [
    {
      id: "visible-token",
      kind: "token",
      name: { "pt-BR": "Herói", "en-US": "Hero" },
      position: { x: 0, z: 0 },
      blocksMovement: true,
      hidden: false,
    },
    {
      id: "hidden-trap",
      kind: "trap",
      name: { "pt-BR": "Armadilha secreta", "en-US": "Secret trap" },
      position: { x: 2, z: 0 },
      blocksMovement: false,
      hidden: true,
    },
  ],
  portals: [],
};

function session(): SessionState {
  return {
    id: "session-secret",
    adventureId: "adventure-secret",
    revision: 4,
    activeFloorId: floor.id,
    participants: {},
    tokenPositions: {},
    tokenFloors: {},
    tokenOwners: {},
    revealedCells: { [floor.id]: [{ x: 0, z: 0 }] },
    initiative: { entries: [], activeIndex: 0, round: 1 },
    rolls: [],
    open: true,
    createdAt: "2026-07-30T12:00:00Z",
  };
}

describe("VTT telemetry", () => {
  it("keeps a bounded frame window and calculates representative metrics", () => {
    const sampler = new FrameSampler(4);
    [10, 20, 30, 60, 40].forEach((frame) => sampler.record(frame));
    sampler.record(Number.NaN);
    const snapshot = sampler.snapshot();
    expect(snapshot.sampleCount).toBe(4);
    expect(snapshot.averageFrameMs).toBe(37.5);
    expect(snapshot.fps).toBe(26.67);
    expect(snapshot.p95FrameMs).toBe(60);
    expect(snapshot.longFrames).toBe(1);
    expect(percentile([30, 10, 20], 0.5)).toBe(20);
  });

  it("reads renderer counters without retaining the renderer object", () => {
    expect(
      rendererTelemetry({
        render: { calls: 12, triangles: 450, points: 3, lines: 7 },
        memory: { geometries: 6, textures: 2 },
      }),
    ).toEqual({
      drawCalls: 12,
      triangles: 450,
      points: 3,
      lines: 7,
      geometries: 6,
      textures: 2,
    });
  });

  it("only counts scene content visible to the current role", () => {
    const player = sceneTelemetry(floor, session(), "player");
    const gm = sceneTelemetry(floor, session(), "gm");
    expect(player).toMatchObject({ entities: 1, tokens: 1, props: 0, fogCells: 2 });
    expect(gm).toMatchObject({ entities: 2, tokens: 1, props: 1, fogCells: 2 });
  });

  it("tracks connection activity through an explicit reducer", () => {
    let telemetry = createConnectionTelemetry();
    telemetry = reduceConnectionTelemetry(telemetry, { type: "connect" });
    telemetry = reduceConnectionTelemetry(telemetry, { type: "open" });
    telemetry = reduceConnectionTelemetry(telemetry, { type: "command-sent" });
    telemetry = reduceConnectionTelemetry(telemetry, { type: "event", revision: 7, latencyMs: -4 });
    telemetry = reduceConnectionTelemetry(telemetry, { type: "rejected" });
    telemetry = reduceConnectionTelemetry(telemetry, { type: "reconnect" });
    expect(telemetry).toMatchObject({
      status: "reconnecting",
      lastRevision: 7,
      commandsSent: 1,
      eventsReceived: 1,
      rejectedCommands: 1,
      reconnectAttempts: 1,
      lastEventLatencyMs: 0,
    });
    expect(reduceConnectionTelemetry(telemetry, { type: "closed" }).status).toBe("closed");
  });

  it("builds a versioned allowlisted report without editorial or session secrets", () => {
    const report = createTelemetryReport({
      render: {
        frames: new FrameSampler().snapshot(),
        renderer: rendererTelemetry({
          render: { calls: 1, triangles: 2, points: 0, lines: 0 },
          memory: { geometries: 3, textures: 4 },
        }),
      },
      scene: sceneTelemetry(floor, session(), "player"),
      connection: createConnectionTelemetry(),
      capturedAt: new Date("2026-07-30T12:00:00Z"),
      viewport: { width: 1280, height: 720, devicePixelRatio: 1.5 },
    });
    const serialized = JSON.stringify(report);
    expect(report.schemaVersion).toBe("vtt-telemetry/v1");
    expect(report.viewport).toEqual({ width: 1280, height: 720, devicePixelRatio: 1.5 });
    expect(serialized).not.toContain("secret");
    expect(serialized).not.toContain("Narrativa");
    expect(serialized).not.toContain("Hero");
  });
});
