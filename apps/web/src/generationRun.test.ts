import { describe, expect, it } from "vitest";
import {
  generationStageTranslationKey,
  isGenerationTerminal,
  reconcileGenerationRun,
} from "./generationRun";
import type { GenerationRun } from "./types";

const run = (overrides: Partial<GenerationRun> = {}): GenerationRun => ({
  id: "run-1",
  status: "running",
  stage: "building-floor-1-of-2",
  progress: 30,
  seed: 42,
  generatorVersion: "go-v1-alpha.2",
  spec: {
    partySize: 4,
    partyLevel: 5,
    durationHours: 4,
    difficulty: "medium",
    theme: "temple",
    biome: "underground",
    floorCount: 2,
    objective: "stop the ritual",
    antagonist: "cult",
    structureStyle: "branching",
    treasureQuality: "standard",
    useAI: false,
  },
  diagnostics: [],
  stages: [],
  createdAt: "2026-07-30T12:00:00Z",
  updatedAt: "2026-07-30T12:00:01Z",
  ...overrides,
});

describe("generation run reconciliation", () => {
  it("rejects stale updates and keeps progress monotonic", () => {
    const current = run({ progress: 60, updatedAt: "2026-07-30T12:00:03Z" });
    expect(
      reconcileGenerationRun(
        current,
        run({ progress: 20, updatedAt: "2026-07-30T12:00:02Z" }),
      ),
    ).toBe(current);
    expect(
      reconcileGenerationRun(
        current,
        run({ progress: 50, updatedAt: "2026-07-30T12:00:04Z" }),
      ).progress,
    ).toBe(60);
  });

  it("does not reopen a terminal execution", () => {
    const completed = run({ status: "completed", progress: 100 });
    expect(reconcileGenerationRun(completed, run())).toBe(completed);
    expect(isGenerationTerminal("completed")).toBe(true);
    expect(isGenerationTerminal("cancelled")).toBe(true);
  });

  it("normalizes dynamic floor stages for translation", () => {
    expect(generationStageTranslationKey("building-floor-2-of-4")).toBe(
      "generationStage_buildingFloor",
    );
    expect(generationStageTranslationKey("validating-document")).toBe(
      "generationStage_validating_document",
    );
  });
});
