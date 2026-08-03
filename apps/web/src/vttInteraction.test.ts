import { describe, expect, it } from "vitest";
import {
  POINTER_DRAG_THRESHOLD,
  idlePointerGesture,
  isMapActivation,
  pointerPurpose,
  reducePointerGesture,
  validateDiceExpression,
} from "./vttInteraction";

describe("VTT pointer gesture machine", () => {
  it("assigns one responsibility to each mouse button", () => {
    expect(pointerPurpose(0)).toBe("map-action");
    expect(pointerPurpose(1)).toBe("camera-pan");
    expect(pointerPurpose(2)).toBe("camera-orbit");
    expect(pointerPurpose(4)).toBe("ignored");
  });

  it("classifies a left click as a map action but suppresses a drag", () => {
    let state = reducePointerGesture(idlePointerGesture, {
      type: "pointer-down",
      pointerId: 7,
      button: 0,
      x: 10,
      y: 10,
    });
    state = reducePointerGesture(state, {
      type: "pointer-move",
      pointerId: 7,
      x: 10 + POINTER_DRAG_THRESHOLD + 1,
      y: 10,
    });
    expect(state.purpose).toBe("map-action");
    expect(state.dragged).toBe(true);
    expect(isMapActivation(0, state.movement)).toBe(false);
  });

  it("never lets camera buttons activate the map", () => {
    expect(isMapActivation(1, 0)).toBe(false);
    expect(isMapActivation(2, 0)).toBe(false);
    expect(isMapActivation(0, POINTER_DRAG_THRESHOLD)).toBe(true);
  });
});

describe("dice expression validation", () => {
  it.each(["1d4", "2D20 + 3", "20d100-100"])("accepts %s", (expression) => {
    expect(validateDiceExpression(expression).valid).toBe(true);
  });

  it.each(["d20", "1d3", "21d6", "1d20+1000", "hello"])("rejects %s", (expression) => {
    expect(validateDiceExpression(expression).valid).toBe(false);
  });
});
