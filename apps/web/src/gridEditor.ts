import type {
  AdventureDocument,
  FloorMap,
  GridPosition,
  Tile,
  WallEdge,
} from "./types";

export type TileEditorTool =
  | "tile-floor"
  | "tile-corridor"
  | "tile-stairs"
  | "tile-water"
  | "tile-lava"
  | "tile-erase";

export type WallEditorTool =
  | "edge-wall"
  | "edge-door"
  | "edge-secret-door"
  | "edge-erase";

export type GridEditorTool = TileEditorTool | WallEditorTool;
export type GridEdgeDirection = WallEdge["direction"];

export interface GridEdit {
  floorId: string;
  tool: GridEditorTool;
  position: GridPosition;
  direction?: GridEdgeDirection;
}

export interface GridEditResult {
  document: AdventureDocument;
  changed: boolean;
  rejection?: "floor-not-found" | "out-of-bounds" | "occupied" | "missing-tile" | "missing-direction";
}

const TILE_KINDS: Record<Exclude<TileEditorTool, "tile-erase">, Tile["kind"]> = {
  "tile-floor": "floor",
  "tile-corridor": "corridor",
  "tile-stairs": "stairs",
  "tile-water": "water",
  "tile-lava": "lava",
};

const EDGE_KINDS: Record<Exclude<WallEditorTool, "edge-erase">, WallEdge["kind"]> = {
  "edge-wall": "wall",
  "edge-door": "door",
  "edge-secret-door": "secret-door",
};

export function isWallEditorTool(tool: GridEditorTool): tool is WallEditorTool {
  return tool.startsWith("edge-");
}

export function applyGridEdit(
  document: AdventureDocument,
  edit: GridEdit,
): GridEditResult {
  const floorIndex = document.floors.findIndex((floor) => floor.id === edit.floorId);
  if (floorIndex < 0) return reject(document, "floor-not-found");

  const floor = document.floors[floorIndex];
  if (!insideFloor(floor, edit.position)) return reject(document, "out-of-bounds");

  const floorResult = isWallEditorTool(edit.tool)
    ? editWall(floor, { ...edit, tool: edit.tool })
    : editTile(floor, { ...edit, tool: edit.tool });
  if (!floorResult.changed) {
    return {
      document,
      changed: false,
      rejection: floorResult.rejection,
    };
  }

  const floors = [...document.floors];
  floors[floorIndex] = floorResult.floor;
  return {
    document: {
      ...document,
      floors,
      updatedAt: new Date().toISOString(),
    },
    changed: true,
  };
}

function editTile(
  floor: FloorMap,
  edit: GridEdit & { tool: TileEditorTool },
): { floor: FloorMap; changed: boolean; rejection?: GridEditResult["rejection"] } {
  const tileIndex = findTileIndex(floor, edit.position);

  if (edit.tool === "tile-erase") {
    if (tileIndex < 0) return { floor, changed: false, rejection: "missing-tile" };
    if (cellIsOccupied(floor, edit.position)) {
      return { floor, changed: false, rejection: "occupied" };
    }
    return {
      floor: {
        ...floor,
        tiles: floor.tiles.filter((_, index) => index !== tileIndex),
        walls: floor.walls.filter((wall) => !wallTouchesCell(wall, edit.position)),
      },
      changed: true,
    };
  }

  const kind = TILE_KINDS[edit.tool];
  const walkable = kind !== "water" && kind !== "lava";
  const current = floor.tiles[tileIndex];
  if (current?.kind === kind && current.walkable === walkable) {
    return { floor, changed: false };
  }

  const tile: Tile = {
    ...(current ?? edit.position),
    kind,
    walkable,
  };
  const tiles =
    tileIndex < 0
      ? sortedTiles([...floor.tiles, tile])
      : floor.tiles.map((item, index) => (index === tileIndex ? tile : item));
  return { floor: { ...floor, tiles }, changed: true };
}

function editWall(
  floor: FloorMap,
  edit: GridEdit & { tool: WallEditorTool },
): { floor: FloorMap; changed: boolean; rejection?: GridEditResult["rejection"] } {
  if (!edit.direction) return { floor, changed: false, rejection: "missing-direction" };
  if (findTileIndex(floor, edit.position) < 0) {
    return { floor, changed: false, rejection: "missing-tile" };
  }

  const wallIndex = floor.walls.findIndex(
    (wall) =>
      wall.x === edit.position.x &&
      wall.z === edit.position.z &&
      wall.direction === edit.direction,
  );

  if (edit.tool === "edge-erase") {
    if (wallIndex < 0) return { floor, changed: false };
    return {
      floor: {
        ...floor,
        walls: floor.walls.filter((_, index) => index !== wallIndex),
      },
      changed: true,
    };
  }

  const kind = EDGE_KINDS[edit.tool];
  const current = floor.walls[wallIndex];
  if (current?.kind === kind) return { floor, changed: false };
  const wall: WallEdge = {
    x: edit.position.x,
    z: edit.position.z,
    direction: edit.direction,
    kind,
    open: false,
    locked: false,
  };
  const walls =
    wallIndex < 0
      ? sortedWalls([...floor.walls, wall])
      : floor.walls.map((item, index) => (index === wallIndex ? wall : item));
  return { floor: { ...floor, walls }, changed: true };
}

function cellIsOccupied(floor: FloorMap, position: GridPosition): boolean {
  if (
    floor.entities.some(
      (entity) => entity.position.x === position.x && entity.position.z === position.z,
    )
  ) {
    return true;
  }
  return floor.portals.some(
    (portal) =>
      (portal.fromFloorId === floor.id &&
        portal.from.x === position.x &&
        portal.from.z === position.z) ||
      (portal.toFloorId === floor.id &&
        portal.to.x === position.x &&
        portal.to.z === position.z),
  );
}

function wallTouchesCell(wall: WallEdge, position: GridPosition): boolean {
  if (wall.x === position.x && wall.z === position.z) return true;
  const neighbor = neighboringCell(wall);
  return neighbor.x === position.x && neighbor.z === position.z;
}

function neighboringCell(wall: WallEdge): GridPosition {
  switch (wall.direction) {
    case "north":
      return { x: wall.x, z: wall.z - 1 };
    case "south":
      return { x: wall.x, z: wall.z + 1 };
    case "east":
      return { x: wall.x + 1, z: wall.z };
    case "west":
      return { x: wall.x - 1, z: wall.z };
  }
}

function findTileIndex(floor: FloorMap, position: GridPosition): number {
  return floor.tiles.findIndex(
    (tile) => tile.x === position.x && tile.z === position.z,
  );
}

function insideFloor(floor: FloorMap, position: GridPosition): boolean {
  return (
    Number.isInteger(position.x) &&
    Number.isInteger(position.z) &&
    position.x >= 0 &&
    position.z >= 0 &&
    position.x < floor.width &&
    position.z < floor.height
  );
}

function sortedTiles(tiles: Tile[]): Tile[] {
  return tiles.sort((left, right) => left.z - right.z || left.x - right.x);
}

function sortedWalls(walls: WallEdge[]): WallEdge[] {
  return walls.sort(
    (left, right) =>
      left.z - right.z ||
      left.x - right.x ||
      left.direction.localeCompare(right.direction),
  );
}

function reject(
  document: AdventureDocument,
  rejection: GridEditResult["rejection"],
): GridEditResult {
  return { document, changed: false, rejection };
}
