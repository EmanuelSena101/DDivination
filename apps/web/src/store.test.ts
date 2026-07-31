import { describe, expect, it } from "vitest";
import { applySessionEvent, useAppStore } from "./store";
import { createSceneBenchmarkFloor } from "./testFixtures/sceneBenchmark";
import type { AdventureDocument, SessionEvent, SessionState } from "./types";

function state(): SessionState {
  return {
    id: "session-1",
    adventureId: "adventure-1",
    revision: 0,
    activeFloorId: "floor-1",
    participants: {},
    tokenPositions: { party: { x: 1, z: 1 } },
    tokenFloors: { party: "floor-1" },
    tokenOwners: {},
    revealedCells: { "floor-1": [] },
    initiative: { entries: [], activeIndex: 0, round: 1 },
    rolls: [],
    open: true,
    createdAt: "2026-07-28T12:00:00Z",
  };
}

function event(type: string, payload: Record<string, unknown>): SessionEvent {
  return {
    revision: 1,
    type,
    actorId: "gm",
    occurredAt: "2026-07-28T12:00:01Z",
    payload,
  };
}

describe("session event projection", () => {
  it("moves a token without mutating the prior snapshot", () => {
    const before = state();
    const after = applySessionEvent(
      before,
      event("token.moved", { tokenId: "party", floorId: "floor-2", x: 8, z: 5 }),
    );
    expect(after.revision).toBe(1);
    expect(after.tokenPositions.party).toEqual({ x: 8, z: 5 });
    expect(after.tokenFloors.party).toBe("floor-2");
    expect(before.tokenPositions.party).toEqual({ x: 1, z: 1 });
  });

  it("applies idempotent fog reveal and hide events", () => {
    const revealed = applySessionEvent(
      state(),
      event("fog.changed", { floorId: "floor-1", x: 4, z: 7, revealed: true }),
    );
    const duplicate = applySessionEvent(
      revealed,
      event("fog.changed", { floorId: "floor-1", x: 4, z: 7, revealed: true }),
    );
    expect(duplicate.revealedCells["floor-1"]).toEqual([{ x: 4, z: 7 }]);
    const hidden = applySessionEvent(
      duplicate,
      event("fog.changed", { floorId: "floor-1", x: 4, z: 7, revealed: false }),
    );
    expect(hidden.revealedCells["floor-1"]).toEqual([]);
  });
});

describe("local editor history", () => {
  it("undoes and redoes grid changes exactly", () => {
    const before = adventure();
    useAppStore.getState().setAdventure(before);
    useAppStore.getState().editGrid({
      floorId: "benchmark-64",
      tool: "tile-lava",
      position: { x: 4, z: 4 },
    });
    const edited = useAppStore.getState().adventure!;

    expect(useAppStore.getState().editorDirty).toBe(true);
    expect(tileKind(edited, 4, 4)).toBe("lava");

    useAppStore.getState().undoGridEdit();
    expect(useAppStore.getState().adventure).toEqual(before);
    expect(useAppStore.getState().editorDirty).toBe(false);

    useAppStore.getState().redoGridEdit();
    expect(useAppStore.getState().adventure).toEqual(edited);
    expect(useAppStore.getState().editorDirty).toBe(true);
  });

  it("shares history between content and entity changes", () => {
    const before = adventure();
    useAppStore.getState().setAdventure(before);
    useAppStore.getState().editContent({
      floorId: "benchmark-64",
      name: localized("Aventura revisada"),
      summary: before.summary,
      hook: before.narrative.hook,
      objective: before.narrative.objective,
      antagonist: before.narrative.antagonist,
      atmosphere: before.narrative.atmosphere,
      floorName: before.floors[0].name,
    });
    useAppStore.getState().addEntity("benchmark-64", {
      id: "draft-prop",
      kind: "prop",
      name: localized("Nova caixa"),
      position: { x: 10, z: 10 },
      assetId: "procedural-crate",
      blocksMovement: true,
      hidden: false,
    });

    expect(useAppStore.getState().editorPast).toHaveLength(2);
    expect(useAppStore.getState().adventure?.floors[0].entities).toHaveLength(1);

    useAppStore.getState().undoGridEdit();
    expect(useAppStore.getState().adventure?.floors[0].entities).toHaveLength(0);
    expect(useAppStore.getState().adventure?.name["pt-BR"]).toBe(
      "Aventura revisada",
    );

    useAppStore.getState().undoGridEdit();
    expect(useAppStore.getState().adventure).toEqual(before);
    expect(useAppStore.getState().editorDirty).toBe(false);
  });

  it("reconciles an autosave without losing a newer local edit", () => {
    const before = adventure();
    useAppStore.getState().setAdventure(before);
    useAppStore.getState().editContent({
      floorId: "benchmark-64",
      name: { "pt-BR": "Enviado", "en-US": "Submitted" },
      summary: before.summary,
      hook: before.narrative.hook,
      objective: before.narrative.objective,
      antagonist: before.narrative.antagonist,
      atmosphere: before.narrative.atmosphere,
      floorName: before.floors[0].name,
    });
    const submitted = useAppStore.getState().adventure!;
    useAppStore.getState().editContent({
      floorId: "benchmark-64",
      name: { "pt-BR": "Mais novo", "en-US": "Newer" },
      summary: before.summary,
      hook: before.narrative.hook,
      objective: before.narrative.objective,
      antagonist: before.narrative.antagonist,
      atmosphere: before.narrative.atmosphere,
      floorName: before.floors[0].name,
    });

    useAppStore.getState().acceptEditorSave(
      {
        ...submitted,
        version: 2,
        updatedAt: "2026-07-30T15:00:00Z",
      },
      submitted,
    );

    expect(useAppStore.getState().adventure?.name["pt-BR"]).toBe("Mais novo");
    expect(useAppStore.getState().adventure?.version).toBe(2);
    expect(useAppStore.getState().editorDirty).toBe(true);
  });
});

function tileKind(document: AdventureDocument, x: number, z: number) {
  return document.floors[0].tiles.find((tile) => tile.x === x && tile.z === z)?.kind;
}

function adventure(): AdventureDocument {
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
    floors: [createSceneBenchmarkFloor({ size: 64, tokenCount: 0, propCount: 0 })],
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
