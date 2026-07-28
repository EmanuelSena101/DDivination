import { create } from "zustand";
import { sessionWebSocketURL } from "./api";
import type {
  AdventureDocument,
  DiceRoll,
  GridPosition,
  Language,
  SessionCommand,
  SessionEvent,
  SessionState,
} from "./types";

interface SessionSnapshotMessage {
  type: "session.snapshot";
  state: SessionState;
  adventure: AdventureDocument;
}

interface AppState {
  language: Language;
  adventure: AdventureDocument | null;
  floorId: string | null;
  session: SessionState | null;
  sessionId: string | null;
  participantId: string | null;
  token: string | null;
  role: "gm" | "player" | "display";
  connected: boolean;
  selectedTokenId: string | null;
  latestRoll: DiceRoll | null;
  latestPing: (GridPosition & { floorId: string; revision: number }) | null;
  error: string | null;
  socket: WebSocket | null;
  setLanguage: (language: Language) => void;
  setAdventure: (adventure: AdventureDocument) => void;
  clearAdventure: () => void;
  setFloor: (floorId: string) => void;
  setSelectedToken: (tokenId: string | null) => void;
  connect: (args: {
    sessionId: string;
    participantId: string;
    token: string;
    role: "gm" | "player" | "display";
    state: SessionState;
    adventure: AdventureDocument;
  }) => void;
  disconnect: () => void;
  send: (type: string, payload: Record<string, unknown>) => void;
  clearError: () => void;
}

export function applySessionEvent(state: SessionState, event: SessionEvent): SessionState {
  const next: SessionState = {
    ...state,
    revision: event.revision,
    tokenPositions: { ...state.tokenPositions },
    tokenFloors: { ...state.tokenFloors },
    revealedCells: { ...state.revealedCells },
    rolls: [...state.rolls],
  };
  switch (event.type) {
    case "token.moved": {
      const tokenId = String(event.payload.tokenId);
      next.tokenPositions[tokenId] = {
        x: Number(event.payload.x),
        z: Number(event.payload.z),
      };
      next.tokenFloors[tokenId] = String(event.payload.floorId);
      break;
    }
    case "fog.changed": {
      const floorId = String(event.payload.floorId);
      const position = { x: Number(event.payload.x), z: Number(event.payload.z) };
      const cells = [...(next.revealedCells[floorId] || [])];
      const index = cells.findIndex((cell) => cell.x === position.x && cell.z === position.z);
      if (event.payload.revealed && index < 0) cells.push(position);
      if (!event.payload.revealed && index >= 0) cells.splice(index, 1);
      next.revealedCells[floorId] = cells;
      break;
    }
    case "floor.changed":
      next.activeFloorId = String(event.payload.floorId);
      break;
    case "initiative.changed":
      next.initiative = event.payload.initiative as SessionState["initiative"];
      break;
    case "participant.joined": {
      const participant = event.payload.participant as SessionState["participants"][string];
      next.participants = { ...state.participants, [participant.id]: participant };
      break;
    }
    case "dice.rolled":
      next.rolls.push(event.payload as unknown as DiceRoll);
      if (next.rolls.length > 100) next.rolls = next.rolls.slice(-100);
      break;
  }
  return next;
}

export const useAppStore = create<AppState>((set, get) => ({
  language:
    (typeof localStorage !== "undefined"
      ? (localStorage.getItem("ddivination-language") as Language)
      : null) || "pt-BR",
  adventure: null,
  floorId: null,
  session: null,
  sessionId: null,
  participantId: null,
  token: null,
  role: "gm",
  connected: false,
  selectedTokenId: null,
  latestRoll: null,
  latestPing: null,
  error: null,
  socket: null,

  setLanguage: (language) => {
    if (typeof localStorage !== "undefined") {
      localStorage.setItem("ddivination-language", language);
    }
    set({ language });
  },
  setAdventure: (adventure) =>
    set({
      adventure,
      floorId: adventure.floors[0]?.id ?? null,
      session: null,
      latestRoll: null,
      latestPing: null,
    }),
  clearAdventure: () => {
    get().socket?.close();
    set({
      adventure: null,
      floorId: null,
      session: null,
      sessionId: null,
      participantId: null,
      token: null,
      role: "gm",
      connected: false,
      selectedTokenId: null,
      latestRoll: null,
      latestPing: null,
      socket: null,
    });
  },
  setFloor: (floorId) => {
    set({ floorId });
    if (get().session && get().role === "gm") {
      get().send("floor.set", { floorId });
    }
  },
  setSelectedToken: (selectedTokenId) => set({ selectedTokenId }),
  connect: ({ sessionId, participantId, token, role, state, adventure }) => {
    get().socket?.close();
    const socket = new WebSocket(sessionWebSocketURL(sessionId, token));
    socket.onopen = () => set({ connected: true, error: null });
    socket.onclose = () => {
      if (get().socket !== socket) return;
      set({ connected: false });
      window.setTimeout(() => {
        const current = get();
        if (
          current.socket === socket &&
          current.sessionId === sessionId &&
          current.token === token
        ) {
          current.connect({ sessionId, participantId, token, role, state, adventure });
        }
      }, 1500);
    };
    socket.onerror = () => set({ error: "WebSocket connection failed" });
    socket.onmessage = (message) => {
      const data = JSON.parse(String(message.data)) as SessionSnapshotMessage | SessionEvent | {
        type: "command.rejected";
        detail: string;
      };
      if (data.type === "session.snapshot" && "state" in data && "adventure" in data) {
        set({
          adventure: data.adventure,
          session: data.state,
          floorId: data.state.activeFloorId,
        });
        return;
      }
      if (data.type === "command.rejected" && "detail" in data) {
        set({ error: data.detail });
        return;
      }
      const event = data as SessionEvent;
      if (event.type === "session.closed") {
        socket.close();
        set({
          socket: null,
          connected: false,
          session: null,
          sessionId: null,
          participantId: null,
          token: null,
          error: "The game master closed this table",
        });
        return;
      }
      set((current) => {
        if (!current.session) return current;
        const sessionState = applySessionEvent(current.session, event);
        return {
          session: sessionState,
          floorId: event.type === "floor.changed" ? sessionState.activeFloorId : current.floorId,
          latestRoll: event.type === "dice.rolled" ? (event.payload as unknown as DiceRoll) : current.latestRoll,
          latestPing:
            event.type === "map.pinged"
              ? {
                  floorId: String(event.payload.floorId),
                  x: Number(event.payload.x),
                  z: Number(event.payload.z),
                  revision: event.revision,
                }
              : current.latestPing,
        };
      });
    };
    set({
      socket,
      sessionId,
      participantId,
      token,
      role,
      session: state,
      adventure,
      floorId: state.activeFloorId || adventure.floors[0]?.id || null,
      connected: false,
      error: null,
    });
  },
  disconnect: () => {
    get().socket?.close();
    set({
      socket: null,
      session: null,
      sessionId: null,
      participantId: null,
      token: null,
      role: "gm",
      connected: false,
      selectedTokenId: null,
    });
  },
  send: (type, payload) => {
    const { socket, session } = get();
    if (!socket || socket.readyState !== WebSocket.OPEN || !session) {
      set({ error: "Session is not connected" });
      return;
    }
    const command: SessionCommand = {
      id: crypto.randomUUID(),
      expectedRevision: session.revision,
      type,
      payload,
    };
    socket.send(JSON.stringify(command));
  },
  clearError: () => set({ error: null }),
}));
