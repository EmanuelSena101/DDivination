import { describe, expect, it } from "vitest";
import {
  TOKEN_INSTANCE_LAYERS,
  sceneSemanticLoad,
  selectSceneQualityProfile,
} from "./scenePerformance";
import { createSceneBenchmarkFloor } from "./testFixtures/sceneBenchmark";

describe("scene performance plan", () => {
  it("uses three draw-call layers regardless of token count", () => {
    expect(TOKEN_INSTANCE_LAYERS).toBe(3);
    expect(createSceneBenchmarkFloor({ size: 64 }).entities.filter((item) => item.kind === "token")).toHaveLength(100);
  });

  it("selects the balanced profile for the 64×64 target", () => {
    const floor = createSceneBenchmarkFloor({ size: 64 });

    expect(selectSceneQualityProfile(floor)).toMatchObject({
      name: "balanced",
      maxDpr: 1.35,
      shadows: true,
      shadowMapSize: 1024,
    });
    expect(sceneSemanticLoad(floor)).toBe(4_952);
  });

  it("selects the performance profile for the 128×128 target", () => {
    const floor = createSceneBenchmarkFloor({ size: 128 });

    expect(selectSceneQualityProfile(floor)).toMatchObject({
      name: "performance",
      maxDpr: 1,
      shadows: false,
      starCount: 0,
    });
    expect(sceneSemanticLoad(floor)).toBe(17_496);
  });

  it("builds deterministic fixtures", () => {
    const first = createSceneBenchmarkFloor({ size: 64 });
    const second = createSceneBenchmarkFloor({ size: 64 });

    expect(first).toEqual(second);
  });
});
