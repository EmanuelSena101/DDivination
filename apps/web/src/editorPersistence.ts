import type { AdventureDocument } from "./types";

export function sameAdventureContent(
  left: AdventureDocument | null,
  right: AdventureDocument | null,
): boolean {
  if (!left || !right) return left === right;
  return fingerprint(left) === fingerprint(right);
}

export function withPersistenceMetadata(
  document: AdventureDocument,
  source: AdventureDocument,
): AdventureDocument {
  return {
    ...document,
    id: source.id,
    schemaVersion: source.schemaVersion,
    generatorVersion: source.generatorVersion,
    version: source.version,
    createdAt: source.createdAt,
    updatedAt: source.updatedAt,
  };
}

export function reconcileSavedAdventure(
  current: AdventureDocument,
  saved: AdventureDocument,
): { document: AdventureDocument; dirty: boolean } {
  if (sameAdventureContent(current, saved)) {
    return { document: saved, dirty: false };
  }
  return {
    document: withPersistenceMetadata(current, saved),
    dirty: true,
  };
}

function fingerprint(document: AdventureDocument): string {
  const { version: _version, updatedAt: _updatedAt, ...content } = document;
  return JSON.stringify(content);
}
