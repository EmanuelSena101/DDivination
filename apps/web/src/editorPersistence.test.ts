import { describe, expect, it } from "vitest";
import {
  reconcileSavedAdventure,
  sameAdventureContent,
} from "./editorPersistence";
import type { AdventureDocument } from "./types";

describe("editor persistence reconciliation", () => {
  it("accepts the authoritative metadata when saved content still matches", () => {
    const current = document("Local", 1);
    const saved = { ...current, version: 2, updatedAt: "2026-07-30T15:00:00Z" };
    const result = reconcileSavedAdventure(current, saved);
    expect(result.dirty).toBe(false);
    expect(result.document.version).toBe(2);
  });

  it("keeps newer local content and rebases it on the saved version", () => {
    const submitted = document("Submitted", 1);
    const saved = { ...submitted, version: 2, updatedAt: "2026-07-30T15:00:00Z" };
    const newer = { ...submitted, name: { "pt-BR": "Mais novo", "en-US": "Newer" } };
    const result = reconcileSavedAdventure(newer, saved);
    expect(result.dirty).toBe(true);
    expect(result.document.name["pt-BR"]).toBe("Mais novo");
    expect(result.document.version).toBe(2);
  });

  it("ignores persistence metadata when comparing content", () => {
    expect(sameAdventureContent(document("A", 1), document("A", 9))).toBe(true);
  });
});

function document(name: string, version: number): AdventureDocument {
  return {
    id: "adventure",
    schemaVersion: "1.0.0",
    generatorVersion: "test",
    version,
    seed: 1,
    name: { "pt-BR": name, "en-US": name },
    spec: {
      partySize: 4,
      partyLevel: 1,
      durationHours: 2,
      difficulty: "medium",
      theme: "test",
      biome: "test",
      floorCount: 1,
      objective: "test",
      antagonist: "test",
      structureStyle: "linear",
      treasureQuality: "standard",
      useAI: false,
    },
    summary: { "pt-BR": "Resumo", "en-US": "Summary" },
    narrative: {
      hook: { "pt-BR": "Gancho", "en-US": "Hook" },
      objective: { "pt-BR": "Objetivo", "en-US": "Objective" },
      antagonist: { "pt-BR": "Antagonista", "en-US": "Antagonist" },
      atmosphere: { "pt-BR": "Atmosfera", "en-US": "Atmosphere" },
    },
    floors: [],
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
      totalRooms: 0,
      totalFloors: 0,
      criticalPath: [],
      deadEnds: [],
      estimatedDifficulty: "medium",
      invariants: [],
    },
    attributions: [],
    createdAt: "2026-07-30T12:00:00Z",
    updatedAt: "2026-07-30T12:00:00Z",
  };
}
