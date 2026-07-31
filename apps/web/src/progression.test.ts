import { describe, expect, it } from "vitest";
import { buildProgressionView } from "./progression";
import type { AdventureDocument } from "./types";

describe("buildProgressionView", () => {
  it("localizes the active floor and connects keys to locks", () => {
    const adventure = {
      floors: [{
        id: "floor-1",
        rooms: [{ id: "room-key", name: { "pt-BR": "Cofre", "en-US": "Vault" } }],
        entities: [{ id: "key-1", name: { "pt-BR": "Chave solar", "en-US": "Sun key" } }],
      }],
      progression: {
        solvable: true,
        steps: [{
          order: 2,
          floorId: "floor-1",
          roomId: "room-key",
          kind: "key",
          beat: { "pt-BR": "A chave é conquistada.", "en-US": "The key is earned." },
          grantsKeyIds: ["key-1"],
          requiresKeyIds: [],
        }],
        locks: [{
          id: "lock-1",
          kind: "portal",
          targetId: "portal-1",
          floorId: "floor-1",
          fromRoomId: "room-key",
          toRoomId: "room-next",
          keyId: "key-1",
        }],
      },
    } as unknown as Pick<AdventureDocument, "floors" | "progression">;

    expect(buildProgressionView(adventure, "pt-BR", "floor-1")).toEqual([{
      order: 2,
      roomName: "Cofre",
      beat: "A chave é conquistada.",
      kind: "key",
      grantedKeys: ["Chave solar"],
      requiredKeys: [],
      lockKind: undefined,
    }]);
    expect(buildProgressionView(adventure, "en-US", "other-floor")).toEqual([]);
  });
});
