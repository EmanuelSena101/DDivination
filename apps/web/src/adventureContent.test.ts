import { describe, expect, it } from "vitest";
import { buildAdventureContentView } from "./adventureContent";
import type { AdventureDocument } from "./types";

describe("buildAdventureContentView", () => {
  it("localizes and filters semantic content by floor", () => {
    const localized = { "pt-BR": "Runas", "en-US": "Runes" };
    const adventure = {
      encounters: [{
        id: "enc-1", floorId: "floor-1", difficulty: "medium", totalXp: 700,
        budgetXp: 900, budgetTier: "moderate", creatures: [{ count: 2, name: localized }],
      }],
      puzzles: [{ id: "puzzle-1", floorId: "floor-1", name: localized, prompt: localized, solution: localized, checkDc: 14 }],
      traps: [], treasures: [], restPoints: [],
    } as unknown as AdventureDocument;

    const items = buildAdventureContentView(adventure, "pt-BR", "floor-1");
    expect(items).toHaveLength(2);
    expect(items[0]).toMatchObject({ kind: "encounter", summary: "2× Runas", meta: "700/900 XP · moderada" });
    expect(items[1]).toMatchObject({ kind: "puzzle", name: "Runas", meta: "DC 14" });
    expect(buildAdventureContentView(adventure, "en-US", "floor-2")).toEqual([]);
  });
});
