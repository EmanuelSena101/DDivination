import { describe, expect, it } from "vitest";
import {
  addSceneEntity,
  applyAdventureContentEdit,
  removeSceneEntity,
  updateSceneEntity,
} from "./contentEditor";
import { createSceneBenchmarkFloor } from "./testFixtures/sceneBenchmark";
import type { AdventureDocument, SceneEntity } from "./types";

describe("content editor", () => {
  it("updates bilingual adventure and floor content immutably", () => {
    const before = document();
    const result = applyAdventureContentEdit(before, {
      floorId: "benchmark-64",
      name: localized("Ruínas Renascidas", "Ruins Reborn"),
      summary: localized("Resumo novo", "New summary"),
      hook: localized("Gancho novo", "New hook"),
      objective: localized("Objetivo novo", "New objective"),
      antagonist: localized("Rival novo", "New rival"),
      atmosphere: localized("Atmosfera nova", "New atmosphere"),
      floorName: localized("Cripta", "Crypt"),
    });

    expect(result.changed).toBe(true);
    expect(result.document.name["pt-BR"]).toBe("Ruínas Renascidas");
    expect(result.document.floors[0].name["en-US"]).toBe("Crypt");
    expect(before.name["pt-BR"]).toBe("Adventure");
  });

  it("adds, moves and removes a scene entity", () => {
    const entity = sceneEntity();
    const added = addSceneEntity(document(), "benchmark-64", entity);
    const moved = updateSceneEntity(added.document, "benchmark-64", {
      ...entity,
      position: { x: 12, z: 13 },
      assetId: " procedural-column ",
    });
    const removed = removeSceneEntity(
      moved.document,
      "benchmark-64",
      entity.id,
    );

    expect(added.changed).toBe(true);
    expect(moved.document.floors[0].entities.at(-1)).toMatchObject({
      position: { x: 12, z: 13 },
      assetId: "procedural-column",
    });
    expect(removed.document.floors[0].entities).toHaveLength(0);
  });

  it("rejects positions outside the map or without a tile", () => {
    const outside = addSceneEntity(document(), "benchmark-64", {
      ...sceneEntity(),
      position: { x: 99, z: 1 },
    });
    expect(outside.rejection).toBe("out-of-bounds");

    const withHole = document();
    withHole.floors[0].tiles = withHole.floors[0].tiles.filter(
      (tile) => tile.x !== 8 || tile.z !== 8,
    );
    const missingTile = addSceneEntity(withHole, "benchmark-64", {
      ...sceneEntity(),
      position: { x: 8, z: 8 },
    });
    expect(missingTile.rejection).toBe("missing-tile");
  });
});

function sceneEntity(): SceneEntity {
  return {
    id: "draft-entity",
    kind: "prop",
    name: localized("Caixa", "Crate"),
    position: { x: 10, z: 10 },
    assetId: "procedural-crate",
    blocksMovement: true,
    hidden: false,
  };
}

function document(): AdventureDocument {
  return {
    id: "adventure-1",
    schemaVersion: "1.0.0",
    generatorVersion: "test",
    version: 1,
    seed: 1,
    name: localized("Adventure", "Adventure"),
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
    summary: localized("Summary", "Summary"),
    narrative: {
      hook: localized("Hook", "Hook"),
      objective: localized("Objective", "Objective"),
      antagonist: localized("Antagonist", "Antagonist"),
      atmosphere: localized("Atmosphere", "Atmosphere"),
    },
    floors: [
      createSceneBenchmarkFloor({ size: 64, tokenCount: 0, propCount: 0 }),
    ],
    encounters: [],
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
      invariants: [],
    },
    attributions: [],
    createdAt: "2026-07-30T00:00:00.000Z",
    updatedAt: "2026-07-30T00:00:00.000Z",
  };
}

function localized(pt: string, en: string) {
  return { "pt-BR": pt, "en-US": en };
}
