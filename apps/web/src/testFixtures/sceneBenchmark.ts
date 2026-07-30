import type { FloorMap, SceneEntity, Tile, WallEdge } from "../types";

export interface SceneBenchmarkOptions {
  size: 64 | 128;
  tokenCount?: number;
  propCount?: number;
}

export function createSceneBenchmarkFloor({
  size,
  tokenCount = 100,
  propCount = 500,
}: SceneBenchmarkOptions): FloorMap {
  return {
    id: `benchmark-${size}`,
    index: 0,
    name: localized(`Benchmark ${size}×${size}`),
    width: size,
    height: size,
    tiles: createTiles(size),
    walls: createPerimeterWalls(size),
    rooms: [
      {
        id: "benchmark-room",
        name: localized("Benchmark room"),
        role: "benchmark",
        description: localized("Deterministic scene performance fixture"),
        center: { x: Math.floor(size / 2), z: Math.floor(size / 2) },
        secret: false,
        mandatory: true,
      },
    ],
    entities: [
      ...createEntities("token", tokenCount, size),
      ...createEntities("prop", propCount, size),
    ],
    portals: [],
  };
}

function createTiles(size: number): Tile[] {
  return Array.from({ length: size * size }, (_, index) => ({
    x: index % size,
    z: Math.floor(index / size),
    kind: index % 11 === 0 ? "corridor" : "floor",
    roomId: "benchmark-room",
    walkable: true,
  }));
}

function createPerimeterWalls(size: number): WallEdge[] {
  const walls: WallEdge[] = [];
  for (let index = 0; index < size; index += 1) {
    walls.push(
      wall(index, 0, "north"),
      wall(index, size - 1, "south"),
      wall(0, index, "west"),
      wall(size - 1, index, "east"),
    );
  }
  return walls;
}

function wall(
  x: number,
  z: number,
  direction: WallEdge["direction"],
): WallEdge {
  return {
    x,
    z,
    direction,
    kind: "wall",
    open: false,
    locked: false,
  };
}

function createEntities(
  kind: "token" | "prop",
  count: number,
  size: number,
): SceneEntity[] {
  return Array.from({ length: count }, (_, index) => ({
    id: `benchmark-${kind}-${String(index).padStart(3, "0")}`,
    kind,
    name: localized(`${kind} ${index + 1}`),
    position: {
      x: (index * 17 + (kind === "token" ? 3 : 7)) % size,
      z: (index * 29 + (kind === "token" ? 5 : 11)) % size,
    },
    assetId:
      kind === "prop"
        ? index % 3 === 0
          ? "base-column"
          : "base-crate"
        : "base-token",
    blocksMovement: kind === "prop",
    hidden: false,
    roomId: "benchmark-room",
  }));
}

function localized(value: string) {
  return {
    "pt-BR": value,
    "en-US": value,
  };
}
