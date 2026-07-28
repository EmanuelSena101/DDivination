import { describe, expect, it } from "vitest";
import { applySessionEvent } from "./store";
import type { SessionEvent, SessionState } from "./types";

function state(): SessionState {
  return {
    id: "session-1",
    adventureId: "adventure-1",
    revision: 0,
    activeFloorId: "floor-1",
    participants: {},
    tokenPositions: { party: { x: 1, z: 1 } },
    tokenFloors: { party: "floor-1" },
    tokenOwners: {},
    revealedCells: { "floor-1": [] },
    initiative: { entries: [], activeIndex: 0, round: 1 },
    rolls: [],
    open: true,
    createdAt: "2026-07-28T12:00:00Z",
  };
}

function event(type: string, payload: Record<string, unknown>): SessionEvent {
  return {
    revision: 1,
    type,
    actorId: "gm",
    occurredAt: "2026-07-28T12:00:01Z",
    payload,
  };
}

describe("session event projection", () => {
  it("moves a token without mutating the prior snapshot", () => {
    const before = state();
    const after = applySessionEvent(
      before,
      event("token.moved", { tokenId: "party", floorId: "floor-2", x: 8, z: 5 }),
    );
    expect(after.revision).toBe(1);
    expect(after.tokenPositions.party).toEqual({ x: 8, z: 5 });
    expect(after.tokenFloors.party).toBe("floor-2");
    expect(before.tokenPositions.party).toEqual({ x: 1, z: 1 });
  });

  it("applies idempotent fog reveal and hide events", () => {
    const revealed = applySessionEvent(
      state(),
      event("fog.changed", { floorId: "floor-1", x: 4, z: 7, revealed: true }),
    );
    const duplicate = applySessionEvent(
      revealed,
      event("fog.changed", { floorId: "floor-1", x: 4, z: 7, revealed: true }),
    );
    expect(duplicate.revealedCells["floor-1"]).toEqual([{ x: 4, z: 7 }]);
    const hidden = applySessionEvent(
      duplicate,
      event("fog.changed", { floorId: "floor-1", x: 4, z: 7, revealed: false }),
    );
    expect(hidden.revealedCells["floor-1"]).toEqual([]);
  });
});
