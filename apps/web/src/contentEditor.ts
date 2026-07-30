import type {
  AdventureDocument,
  LocalizedText,
  SceneEntity,
} from "./types";

export interface AdventureContentEdit {
  floorId: string;
  name: LocalizedText;
  summary: LocalizedText;
  hook: LocalizedText;
  objective: LocalizedText;
  antagonist: LocalizedText;
  atmosphere: LocalizedText;
  floorName: LocalizedText;
}

export type ContentEditRejection =
  | "floor-not-found"
  | "entity-not-found"
  | "duplicate-entity"
  | "out-of-bounds"
  | "missing-tile";

export interface ContentEditResult {
  document: AdventureDocument;
  changed: boolean;
  rejection?: ContentEditRejection;
}

export function applyAdventureContentEdit(
  document: AdventureDocument,
  edit: AdventureContentEdit,
): ContentEditResult {
  const floorIndex = document.floors.findIndex((floor) => floor.id === edit.floorId);
  if (floorIndex < 0) return reject(document, "floor-not-found");

  const floor = document.floors[floorIndex];
  const nextFloor = { ...floor, name: normalizedText(edit.floorName) };
  const floors = [...document.floors];
  floors[floorIndex] = nextFloor;
  const next = touch({
    ...document,
    name: normalizedText(edit.name),
    summary: normalizedText(edit.summary),
    narrative: {
      hook: normalizedText(edit.hook),
      objective: normalizedText(edit.objective),
      antagonist: normalizedText(edit.antagonist),
      atmosphere: normalizedText(edit.atmosphere),
    },
    floors,
  });
  return unchanged(document, next)
    ? { document, changed: false }
    : { document: next, changed: true };
}

export function addSceneEntity(
  document: AdventureDocument,
  floorId: string,
  entity: SceneEntity,
): ContentEditResult {
  const floorIndex = document.floors.findIndex((floor) => floor.id === floorId);
  if (floorIndex < 0) return reject(document, "floor-not-found");
  if (
    document.floors.some((floor) =>
      floor.entities.some((candidate) => candidate.id === entity.id),
    )
  ) {
    return reject(document, "duplicate-entity");
  }
  const positionRejection = validatePosition(
    document.floors[floorIndex],
    entity.position.x,
    entity.position.z,
  );
  if (positionRejection) return reject(document, positionRejection);

  return replaceFloor(document, floorIndex, {
    ...document.floors[floorIndex],
    entities: [
      ...document.floors[floorIndex].entities,
      normalizedEntity(entity),
    ],
  });
}

export function updateSceneEntity(
  document: AdventureDocument,
  floorId: string,
  entity: SceneEntity,
): ContentEditResult {
  const floorIndex = document.floors.findIndex((floor) => floor.id === floorId);
  if (floorIndex < 0) return reject(document, "floor-not-found");
  const floor = document.floors[floorIndex];
  const entityIndex = floor.entities.findIndex(
    (candidate) => candidate.id === entity.id,
  );
  if (entityIndex < 0) return reject(document, "entity-not-found");
  const positionRejection = validatePosition(
    floor,
    entity.position.x,
    entity.position.z,
  );
  if (positionRejection) return reject(document, positionRejection);

  const nextEntity = normalizedEntity(entity);
  if (JSON.stringify(floor.entities[entityIndex]) === JSON.stringify(nextEntity)) {
    return { document, changed: false };
  }
  return replaceFloor(document, floorIndex, {
    ...floor,
    entities: floor.entities.map((candidate, index) =>
      index === entityIndex ? nextEntity : candidate,
    ),
  });
}

export function removeSceneEntity(
  document: AdventureDocument,
  floorId: string,
  entityId: string,
): ContentEditResult {
  const floorIndex = document.floors.findIndex((floor) => floor.id === floorId);
  if (floorIndex < 0) return reject(document, "floor-not-found");
  const floor = document.floors[floorIndex];
  if (!floor.entities.some((entity) => entity.id === entityId)) {
    return reject(document, "entity-not-found");
  }
  return replaceFloor(document, floorIndex, {
    ...floor,
    entities: floor.entities.filter((entity) => entity.id !== entityId),
  });
}

export function defaultAssetForKind(kind: SceneEntity["kind"]): string {
  switch (kind) {
    case "light":
      return "procedural-brazier";
    case "trap":
      return "procedural-marker-trap";
    case "marker":
      return "procedural-marker";
    case "token":
      return "procedural-party-token";
    case "boss":
      return "procedural-boss-token";
    case "key":
      return "procedural-chest";
    case "prop":
      return "procedural-crate";
  }
}

function validatePosition(
  floor: AdventureDocument["floors"][number],
  x: number,
  z: number,
): ContentEditRejection | undefined {
  if (
    !Number.isInteger(x) ||
    !Number.isInteger(z) ||
    x < 0 ||
    z < 0 ||
    x >= floor.width ||
    z >= floor.height
  ) {
    return "out-of-bounds";
  }
  if (!floor.tiles.some((tile) => tile.x === x && tile.z === z)) {
    return "missing-tile";
  }
  return undefined;
}

function normalizedEntity(entity: SceneEntity): SceneEntity {
  return {
    ...entity,
    name: normalizedText(entity.name),
    assetId: entity.assetId?.trim() || undefined,
  };
}

function normalizedText(value: LocalizedText): LocalizedText {
  return {
    "pt-BR": value["pt-BR"].trim(),
    "en-US": value["en-US"].trim(),
  };
}

function replaceFloor(
  document: AdventureDocument,
  floorIndex: number,
  floor: AdventureDocument["floors"][number],
): ContentEditResult {
  const floors = [...document.floors];
  floors[floorIndex] = floor;
  return {
    document: touch({ ...document, floors }),
    changed: true,
  };
}

function touch(document: AdventureDocument): AdventureDocument {
  return { ...document, updatedAt: new Date().toISOString() };
}

function unchanged(
  before: AdventureDocument,
  after: AdventureDocument,
): boolean {
  return JSON.stringify({ ...before, updatedAt: "" }) ===
    JSON.stringify({ ...after, updatedAt: "" });
}

function reject(
  document: AdventureDocument,
  rejection: ContentEditRejection,
): ContentEditResult {
  return { document, changed: false, rejection };
}
