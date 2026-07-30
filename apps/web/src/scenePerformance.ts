import type { FloorMap } from "./types";

export const TOKEN_INSTANCE_LAYERS = 3;

export type SceneQualityProfileName = "quality" | "balanced" | "performance";

export interface SceneQualityProfile {
  name: SceneQualityProfileName;
  maxDpr: number;
  shadowMapSize: 512 | 1024 | 2048;
  shadows: boolean;
  starCount: number;
}

const QUALITY_PROFILES: Record<SceneQualityProfileName, SceneQualityProfile> = {
  quality: {
    name: "quality",
    maxDpr: 1.75,
    shadowMapSize: 2048,
    shadows: true,
    starCount: 700,
  },
  balanced: {
    name: "balanced",
    maxDpr: 1.35,
    shadowMapSize: 1024,
    shadows: true,
    starCount: 320,
  },
  performance: {
    name: "performance",
    maxDpr: 1,
    shadowMapSize: 512,
    shadows: false,
    starCount: 0,
  },
};

export function sceneSemanticLoad(floor: FloorMap): number {
  return floor.tiles.length + floor.walls.length + floor.entities.length;
}

export function selectSceneQualityProfile(floor: FloorMap): SceneQualityProfile {
  const area = floor.width * floor.height;
  const load = sceneSemanticLoad(floor);

  if (area >= 128 * 128 || load >= 12_000) {
    return QUALITY_PROFILES.performance;
  }
  if (area >= 64 * 64 || load >= 4_000) {
    return QUALITY_PROFILES.balanced;
  }
  return QUALITY_PROFILES.quality;
}
