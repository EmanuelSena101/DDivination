import { create } from "zustand";
import { sessionWebSocketURL } from "./api";
import {
  addSceneEntity,
  applyAdventureContentEdit,
  removeSceneEntity,
  updateSceneEntity,
  type AdventureContentEdit,
  type ContentEditRejection,
} from "./contentEditor";
import { applyGridEdit, type GridEdit } from "./gridEditor";
import {
  reconcileSavedAdventure,
  sameAdventureContent,
  withPersistenceMetadata,
} from "./editorPersistence";
import {
  createConnectionTelemetry,
  reduceConnectionTelemetry,
  type ConnectionTelemetry,
} from "./telemetry";
import type {
  AdventureDocument,
  DiceRoll,
  GridPosition,
  Language,
  SceneEntity,
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
  editorPast: AdventureDocument[];
  editorFuture: AdventureDocument[];
  editorDirty: boolean;
  editorBaseline: AdventureDocument | null;
  connectionTelemetry: ConnectionTelemetry;
  error: string | null;
  socket: WebSocket | null;
  setLanguage: (language: Language) => void;
  setAdventure: (adventure: AdventureDocument) => void;
  clearAdventure: () => void;
  setFloor: (floorId: string) => void;
  setSelectedToken: (tokenId: string | null) => void;
  editGrid: (edit: GridEdit) => void;
  editContent: (edit: AdventureContentEdit) => void;
  addEntity: (floorId: string, entity: SceneEntity) => void;
  updateEntity: (floorId: string, entity: SceneEntity) => void;
  removeEntity: (floorId: string, entityId: string) => void;
  undoGridEdit: () => void;
  redoGridEdit: () => void;
  discardGridEdits: () => void;
  acceptEditorSave: (
    saved: AdventureDocument,
    submitted: AdventureDocument,
  ) => void;
  rebaseEditorAgainst: (remote: AdventureDocument) => void;
  connect: (args: {
    sessionId: string;
    participantId: string;
    token: string;
    role: "gm" | "player" | "display";
    state: SessionState;
    adventure: AdventureDocument;
    reconnecting?: boolean;
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
  editorPast: [],
  editorFuture: [],
  editorDirty: false,
  editorBaseline: null,
  connectionTelemetry: createConnectionTelemetry(),
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
      editorPast: [],
      editorFuture: [],
      editorDirty: false,
      editorBaseline: adventure,
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
      editorPast: [],
      editorFuture: [],
      editorDirty: false,
      editorBaseline: null,
      connectionTelemetry: createConnectionTelemetry(),
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
  editGrid: (edit) => {
    const current = get();
    if (!current.adventure) return;
    if (current.session) {
      set({ error: "Grid editing is unavailable while a table is open" });
      return;
    }
    const result = applyGridEdit(current.adventure, edit);
    if (!result.changed) {
      if (result.rejection) {
        set({ error: gridEditError(result.rejection) });
      }
      return;
    }
    set(editorChange(current, result.document));
  },
  editContent: (edit) => {
    const current = get();
    if (!current.adventure) return;
    if (current.session) {
      set({ error: "Content editing is unavailable while a table is open" });
      return;
    }
    const result = applyAdventureContentEdit(current.adventure, edit);
    if (!result.changed) {
      if (result.rejection) set({ error: contentEditError(result.rejection) });
      return;
    }
    set(editorChange(current, result.document));
  },
  addEntity: (floorId, entity) => {
    const current = get();
    if (!current.adventure) return;
    if (current.session) {
      set({ error: "Entity editing is unavailable while a table is open" });
      return;
    }
    const result = addSceneEntity(current.adventure, floorId, entity);
    if (!result.changed) {
      if (result.rejection) set({ error: contentEditError(result.rejection) });
      return;
    }
    set(editorChange(current, result.document));
  },
  updateEntity: (floorId, entity) => {
    const current = get();
    if (!current.adventure) return;
    if (current.session) {
      set({ error: "Entity editing is unavailable while a table is open" });
      return;
    }
    const result = updateSceneEntity(current.adventure, floorId, entity);
    if (!result.changed) {
      if (result.rejection) set({ error: contentEditError(result.rejection) });
      return;
    }
    set(editorChange(current, result.document));
  },
  removeEntity: (floorId, entityId) => {
    const current = get();
    if (!current.adventure) return;
    if (current.session) {
      set({ error: "Entity editing is unavailable while a table is open" });
      return;
    }
    const result = removeSceneEntity(current.adventure, floorId, entityId);
    if (!result.changed) {
      if (result.rejection) set({ error: contentEditError(result.rejection) });
      return;
    }
    set(editorChange(current, result.document));
  },
  undoGridEdit: () => {
    const current = get();
    if (current.session || !current.adventure || current.editorPast.length === 0) return;
    const previous = current.editorPast.at(-1)!;
    set({
      adventure: previous,
      editorPast: current.editorPast.slice(0, -1),
      editorFuture: [current.adventure, ...current.editorFuture].slice(0, 40),
      editorDirty: !sameAdventureContent(previous, current.editorBaseline),
      error: null,
    });
  },
  redoGridEdit: () => {
    const current = get();
    if (current.session || !current.adventure || current.editorFuture.length === 0) return;
    const [next, ...remaining] = current.editorFuture;
    set({
      adventure: next,
      editorPast: [...current.editorPast, current.adventure].slice(-40),
      editorFuture: remaining,
      editorDirty: !sameAdventureContent(next, current.editorBaseline),
      error: null,
    });
  },
  discardGridEdits: () => {
    const current = get();
    if (current.session || !current.editorBaseline) return;
    set({
      adventure: current.editorBaseline,
      editorPast: [],
      editorFuture: [],
      editorDirty: false,
      error: null,
      selectedTokenId: null,
    });
  },
  acceptEditorSave: (saved, submitted) => {
    const current = get();
    if (!current.adventure || current.adventure.id !== saved.id) return;
    const reconciled = reconcileSavedAdventure(current.adventure, {
      ...submitted,
      version: saved.version,
      createdAt: saved.createdAt,
      updatedAt: saved.updatedAt,
    });
    set({
      adventure: reconciled.document,
      editorBaseline: saved,
      editorPast: current.editorPast.map((item) =>
        withPersistenceMetadata(item, saved),
      ),
      editorFuture: current.editorFuture.map((item) =>
        withPersistenceMetadata(item, saved),
      ),
      editorDirty: reconciled.dirty,
      error: null,
    });
  },
  rebaseEditorAgainst: (remote) => {
    const current = get();
    if (!current.adventure || current.adventure.id !== remote.id) return;
    set({
      adventure: withPersistenceMetadata(current.adventure, remote),
      editorBaseline: remote,
      editorPast: current.editorPast.map((item) =>
        withPersistenceMetadata(item, remote),
      ),
      editorFuture: current.editorFuture.map((item) =>
        withPersistenceMetadata(item, remote),
      ),
      editorDirty: !sameAdventureContent(current.adventure, remote),
      error: null,
    });
  },
  connect: ({ sessionId, participantId, token, role, state, adventure, reconnecting = false }) => {
    get().socket?.close();
    const socket = new WebSocket(sessionWebSocketURL(sessionId, token));
    const initialConnectionTelemetry = reconnecting
      ? reduceConnectionTelemetry(get().connectionTelemetry, { type: "connect" })
      : reduceConnectionTelemetry(createConnectionTelemetry(), { type: "connect" });
    socket.onopen = () =>
      set((current) => ({
        connected: true,
        error: null,
        connectionTelemetry: reduceConnectionTelemetry(current.connectionTelemetry, { type: "open" }),
      }));
    socket.onclose = () => {
      if (get().socket !== socket) return;
      set((current) => ({
        connected: false,
        connectionTelemetry: reduceConnectionTelemetry(current.connectionTelemetry, {
          type: "reconnect",
        }),
      }));
      window.setTimeout(() => {
        const current = get();
        if (
          current.socket === socket &&
          current.sessionId === sessionId &&
          current.token === token
        ) {
          current.connect({
            sessionId,
            participantId,
            token,
            role,
            state,
            adventure,
            reconnecting: true,
          });
        }
      }, 1500);
    };
    socket.onerror = () =>
      set((current) => ({
        error: "WebSocket connection failed",
        connectionTelemetry: reduceConnectionTelemetry(current.connectionTelemetry, {
          type: "error",
        }),
      }));
    socket.onmessage = (message) => {
      const data = JSON.parse(String(message.data)) as SessionSnapshotMessage | SessionEvent | {
        type: "command.rejected";
        detail: string;
      };
      if (data.type === "session.snapshot" && "state" in data && "adventure" in data) {
        set((current) => ({
          adventure: data.adventure,
          session: data.state,
          floorId: data.state.activeFloorId,
          editorPast: [],
          editorFuture: [],
          editorDirty: false,
          editorBaseline: adventure,
          connectionTelemetry: reduceConnectionTelemetry(current.connectionTelemetry, {
            type: "snapshot",
            revision: data.state.revision,
          }),
        }));
        return;
      }
      if (data.type === "command.rejected" && "detail" in data) {
        set((current) => ({
          error: data.detail,
          connectionTelemetry: reduceConnectionTelemetry(current.connectionTelemetry, {
            type: "rejected",
          }),
        }));
        return;
      }
      const event = data as SessionEvent;
      const occurredAt = Date.parse(event.occurredAt);
      const latencyMs = Number.isFinite(occurredAt) ? Date.now() - occurredAt : 0;
      set((current) => ({
        connectionTelemetry: reduceConnectionTelemetry(current.connectionTelemetry, {
          type: "event",
          revision: event.revision,
          latencyMs,
        }),
      }));
      if (event.type === "session.closed") {
        socket.close();
        set((current) => ({
          socket: null,
          connected: false,
          session: null,
          sessionId: null,
          participantId: null,
          token: null,
          error: "The game master closed this table",
          connectionTelemetry: reduceConnectionTelemetry(current.connectionTelemetry, {
            type: "closed",
          }),
        }));
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
      editorPast: [],
      editorFuture: [],
      editorDirty: false,
      editorBaseline: null,
      connected: false,
      connectionTelemetry: initialConnectionTelemetry,
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
      connectionTelemetry: createConnectionTelemetry(),
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
    set((current) => ({
      connectionTelemetry: reduceConnectionTelemetry(current.connectionTelemetry, {
        type: "command-sent",
      }),
    }));
  },
  clearError: () => set({ error: null }),
}));

function gridEditError(rejection: NonNullable<ReturnType<typeof applyGridEdit>["rejection"]>): string {
  switch (rejection) {
    case "occupied":
      return "A tile containing an entity or portal cannot be removed";
    case "missing-tile":
      return "Add a tile before editing this edge";
    case "out-of-bounds":
      return "The selected cell is outside this floor";
    case "missing-direction":
      return "Select a valid grid edge";
    case "floor-not-found":
      return "The selected floor no longer exists";
  }
}

function editorChange(
  current: AppState,
  adventure: AdventureDocument,
): Partial<AppState> {
  return {
    adventure,
    editorPast: [...current.editorPast, current.adventure!].slice(-40),
    editorFuture: [],
    editorDirty: !sameAdventureContent(adventure, current.editorBaseline),
    error: null,
    selectedTokenId: null,
  };
}

function contentEditError(rejection: ContentEditRejection): string {
  switch (rejection) {
    case "floor-not-found":
      return "The selected floor no longer exists";
    case "entity-not-found":
      return "The selected entity no longer exists";
    case "duplicate-entity":
      return "An entity with this identifier already exists";
    case "out-of-bounds":
      return "Entity coordinates are outside the active floor";
    case "missing-tile":
      return "Entities must be placed on an existing tile";
  }
}
