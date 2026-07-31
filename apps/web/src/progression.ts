import type { AdventureDocument, Language } from "./types";

export interface ProgressionViewItem {
  order: number;
  roomName: string;
  beat: string;
  kind: AdventureDocument["progression"]["steps"][number]["kind"];
  grantedKeys: string[];
  requiredKeys: string[];
  lockKind?: "door" | "portal";
}

export function buildProgressionView(
  adventure: Pick<AdventureDocument, "floors" | "progression">,
  language: Language,
  floorId: string,
): ProgressionViewItem[] {
  if (!adventure.progression?.steps) return [];
  const rooms = new Map(
    adventure.floors.flatMap((floor) =>
      floor.rooms.map((room) => [room.id, room] as const),
    ),
  );
  const entities = new Map(
    adventure.floors.flatMap((floor) =>
      floor.entities.map((entity) => [entity.id, entity] as const),
    ),
  );
  const locks = new Map(
    adventure.progression.locks.map((lock) => [lock.toRoomId, lock] as const),
  );
  return adventure.progression.steps
    .filter((step) => step.floorId === floorId)
    .map((step) => ({
      order: step.order,
      roomName: rooms.get(step.roomId)?.name[language] || step.roomId,
      beat: step.beat[language],
      kind: step.kind,
      grantedKeys: step.grantsKeyIds.map(
        (keyId) => entities.get(keyId)?.name[language] || keyId,
      ),
      requiredKeys: step.requiresKeyIds.map(
        (keyId) => entities.get(keyId)?.name[language] || keyId,
      ),
      lockKind: locks.get(step.roomId)?.kind,
    }));
}
