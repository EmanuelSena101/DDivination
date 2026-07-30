import type { SceneEntity, Tile, WallEdge } from "./types";

export type ProceduralPropFamily =
  | "column"
  | "crate"
  | "brazier"
  | "chest"
  | "marker"
  | "generic";

export interface SurfacePalette {
  base: string;
  detail: string;
  emissive: string;
  emissiveIntensity: number;
}

const TILE_PALETTES: Record<Tile["kind"], SurfacePalette> = {
  floor: {
    base: "#757b8f",
    detail: "#a2a8bb",
    emissive: "#171a24",
    emissiveIntensity: 0.35,
  },
  corridor: {
    base: "#565e70",
    detail: "#80899c",
    emissive: "#11151d",
    emissiveIntensity: 0.3,
  },
  stairs: {
    base: "#866f4d",
    detail: "#d3a55d",
    emissive: "#2c2113",
    emissiveIntensity: 0.45,
  },
  water: {
    base: "#20556f",
    detail: "#46a7d0",
    emissive: "#0b405b",
    emissiveIntensity: 0.75,
  },
  lava: {
    base: "#772d20",
    detail: "#ff7b38",
    emissive: "#9c260e",
    emissiveIntensity: 1.25,
  },
};

export function tilePalette(kind: Tile["kind"]): SurfacePalette {
  return TILE_PALETTES[kind];
}

export function wallColor(kind: WallEdge["kind"]): string {
  switch (kind) {
    case "door":
      return "#8b643b";
    case "secret-door":
      return "#75688d";
    case "wall":
      return "#626775";
  }
}

export function proceduralPropFamily(entity: SceneEntity): ProceduralPropFamily {
  const asset = entity.assetId?.toLowerCase() ?? "";
  if (asset.includes("column") || asset.includes("pillar")) return "column";
  if (asset.includes("brazier") || entity.kind === "light") return "brazier";
  if (asset.includes("chest") || entity.kind === "key") return "chest";
  if (asset.includes("crate") || asset.includes("barrel")) return "crate";
  if (entity.kind === "marker" || entity.kind === "trap") return "marker";
  return "generic";
}

export function propColor(family: ProceduralPropFamily, entity: SceneEntity): string {
  if (entity.kind === "key") return "#d8b45e";
  switch (family) {
    case "column":
      return "#7d818d";
    case "crate":
      return "#765236";
    case "brazier":
      return "#4d4240";
    case "chest":
      return "#8d6238";
    case "marker":
      return entity.kind === "trap" ? "#99484d" : "#7b63c3";
    case "generic":
      return "#6f5970";
  }
}
