export const POINTER_DRAG_THRESHOLD = 6;

export type VTTInteractionTool =
  | "select"
  | "fog"
  | "ping"
  | "measure"
  | "edit-grid";

export type PointerPurpose =
  | "idle"
  | "map-action"
  | "camera-orbit"
  | "camera-pan"
  | "ignored";

export interface PointerGestureState {
  purpose: PointerPurpose;
  pointerId: number | null;
  origin: { x: number; y: number } | null;
  movement: number;
  dragged: boolean;
}

export type PointerGestureEvent =
  | { type: "pointer-down"; pointerId: number; button: number; x: number; y: number }
  | { type: "pointer-move"; pointerId: number; x: number; y: number }
  | { type: "pointer-up"; pointerId: number }
  | { type: "cancel" };

export const idlePointerGesture: PointerGestureState = {
  purpose: "idle",
  pointerId: null,
  origin: null,
  movement: 0,
  dragged: false,
};

export function pointerPurpose(button: number): PointerPurpose {
  if (button === 0) return "map-action";
  if (button === 1) return "camera-pan";
  if (button === 2) return "camera-orbit";
  return "ignored";
}

export function reducePointerGesture(
  state: PointerGestureState,
  event: PointerGestureEvent,
): PointerGestureState {
  if (event.type === "cancel" || event.type === "pointer-up") {
    if (event.type === "pointer-up" && event.pointerId !== state.pointerId) return state;
    return idlePointerGesture;
  }
  if (event.type === "pointer-down") {
    return {
      purpose: pointerPurpose(event.button),
      pointerId: event.pointerId,
      origin: { x: event.x, y: event.y },
      movement: 0,
      dragged: false,
    };
  }
  if (state.pointerId !== event.pointerId || !state.origin) return state;
  const movement = Math.hypot(event.x - state.origin.x, event.y - state.origin.y);
  return {
    ...state,
    movement: Math.max(state.movement, movement),
    dragged: state.dragged || movement > POINTER_DRAG_THRESHOLD,
  };
}

export function isMapActivation(button: number, movement: number): boolean {
  return button === 0 && movement <= POINTER_DRAG_THRESHOLD;
}

export function cursorForTool(tool: VTTInteractionTool, hasSelectedToken: boolean): string {
  if (tool === "fog") return "crosshair";
  if (tool === "ping") return "cell";
  if (tool === "measure") return "crosshair";
  if (tool === "edit-grid") return "copy";
  return hasSelectedToken ? "crosshair" : "default";
}

export type DiceValidation =
  | { valid: true; normalized: string }
  | { valid: false; reason: "format" | "count" };

const diceExpression = /^(\d{1,2})d(4|6|8|10|12|20|100)([+-]\d{1,3})?$/i;

export function validateDiceExpression(expression: string): DiceValidation {
  const normalized = expression.trim().replaceAll(" ", "").toLowerCase();
  const match = diceExpression.exec(normalized);
  if (!match) return { valid: false, reason: "format" };
  const count = Number(match[1]);
  if (count < 1 || count > 20) return { valid: false, reason: "count" };
  return { valid: true, normalized };
}
