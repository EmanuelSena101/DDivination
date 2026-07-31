import { describe, expect, it } from "vitest";
import { applyGridEdit } from "./gridEditor";
import { createSceneBenchmarkFloor } from "./testFixtures/sceneBenchmark";
import type { AdventureDocument } from "./types";

describe("grid editor", () => {
  it("paints and replaces a tile immutably", () => {
    const before = document();
    const result = applyGridEdit(before, {
      floorId: "benchmark-64",
      tool: "tile-lava",
      position: { x: 2, z: 3 },
    });

    expect(result.changed).toBe(true);
    expect(tile(result.document, 2, 3)).toMatchObject({
      kind: "lava",
      walkable: false,
    });
    expect(tile(before, 2, 3)?.kind).not.toBe("lava");
  });

  it("adds, replaces and removes an edge", () => {
    const withoutWalls = {
      ...document(),
      floors: [{ ...document().floors[0], walls: [] }],
    };
    const wall = applyGridEdit(withoutWalls, {
      floorId: "benchmark-64",
      tool: "edge-wall",
      position: { x: 4, z: 4 },
      direction: "east",
    });
    const door = applyGridEdit(wall.document, {
      floorId: "benchmark-64",
      tool: "edge-door",
      position: { x: 4, z: 4 },
      direction: "east",
    });
    const removed = applyGridEdit(door.document, {
      floorId: "benchmark-64",
      tool: "edge-erase",
      position: { x: 4, z: 4 },
      direction: "east",
    });

    expect(wall.document.floors[0].walls).toHaveLength(1);
    expect(door.document.floors[0].walls[0].kind).toBe("door");
    expect(removed.document.floors[0].walls).toHaveLength(0);
  });

  it("protects cells occupied by an entity or portal", () => {
    const withEntity = applyGridEdit(document(), {
      floorId: "benchmark-64",
      tool: "tile-erase",
      position: { x: 3, z: 5 },
    });
    expect(withEntity.changed).toBe(false);
    expect(withEntity.rejection).toBe("occupied");

    const withPortal = document();
    withPortal.floors[0].portals = [
      {
        id: "portal-1",
        fromFloorId: "benchmark-64",
        from: { x: 8, z: 8 },
        toFloorId: "floor-2",
        to: { x: 1, z: 1 },
        kind: "stairs-down",
      },
    ];
    const portalResult = applyGridEdit(withPortal, {
      floorId: "benchmark-64",
      tool: "tile-erase",
      position: { x: 8, z: 8 },
    });
    expect(portalResult.rejection).toBe("occupied");
  });
});

function tile(adventure: AdventureDocument, x: number, z: number) {
  return adventure.floors[0].tiles.find((item) => item.x === x && item.z === z);
}

function document(): AdventureDocument {
  return {
    id: "adventure-1",
    schemaVersion: "1.0.0",
    generatorVersion: "test",
    version: 1,
    seed: 1,
    name: localized("Adventure"),
    spec: {
      partySize: 4,
      partyLevel: 5,
      durationHours: 4,
      difficulty: "medium",
      theme: "test",
      biome: "test",
      floorCount: 1,
      objective: "test",
      antagonist: "test",
      structureStyle: "branching",
      treasureQuality: "standard",
      useAI: false,
    },
    summary: localized("Summary"),
    narrative: {
      hook: localized("Hook"),
      objective: localized("Objective"),
      antagonist: localized("Antagonist"),
      atmosphere: localized("Atmosphere"),
    },
    floors: [createSceneBenchmarkFloor({ size: 64, tokenCount: 1, propCount: 0 })],
    encounters: [],
    treasures: [],
    puzzles: [],
    traps: [],
    restPoints: [],
    progression: {
      entryRoomId: "",
      objectiveRoomId: "",
      climaxRoomId: "",
      steps: [],
      locks: [],
      secretRoomIds: [],
      solvable: false,
    },
    analysis: {
      totalRooms: 1,
      totalFloors: 1,
      criticalPath: [],
      deadEnds: [],
      estimatedDifficulty: "medium",
      encounterBudgetXp: 0,
      encounterTotalXp: 0,
      treasureValueGp: 0,
      contentCounts: { encounters: 0, treasures: 0, puzzles: 0, traps: 0, restPoints: 0 },
      invariants: [],
    },
    attributions: [],
    createdAt: "2026-07-30T00:00:00.000Z",
    updatedAt: "2026-07-30T00:00:00.000Z",
  };
}

function localized(value: string) {
  return { "pt-BR": value, "en-US": value };
}
