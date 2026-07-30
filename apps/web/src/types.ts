export type Language = "pt-BR" | "en-US";

export interface LocalizedText {
  "pt-BR": string;
  "en-US": string;
}

export interface AdventureSpec {
  name?: LocalizedText;
  partySize: number;
  partyLevel: number;
  durationHours: number;
  difficulty: "easy" | "medium" | "hard" | "deadly";
  theme: string;
  biome: string;
  floorCount: number;
  objective: string;
  antagonist: string;
  structureStyle: "linear" | "branching" | "labyrinthine";
  treasureQuality: "poor" | "standard" | "rich" | "legendary";
  useAI: boolean;
}

export interface GridPosition {
  x: number;
  z: number;
}

export interface Tile extends GridPosition {
  kind: "floor" | "corridor" | "stairs" | "water" | "lava";
  roomId?: string;
  walkable: boolean;
}

export interface WallEdge extends GridPosition {
  direction: "north" | "east" | "south" | "west";
  kind: "wall" | "door" | "secret-door";
  open: boolean;
  locked: boolean;
}

export interface Portal {
  id: string;
  fromFloorId: string;
  from: GridPosition;
  toFloorId: string;
  to: GridPosition;
  kind: "stairs-up" | "stairs-down" | "portal";
  requiredKeyId?: string;
}

export interface Room {
  id: string;
  name: LocalizedText;
  role: string;
  description: LocalizedText;
  center: GridPosition;
  secret: boolean;
  mandatory: boolean;
}

export interface SceneEntity {
  id: string;
  kind: "prop" | "light" | "trap" | "token" | "marker" | "key" | "boss";
  name: LocalizedText;
  position: GridPosition;
  assetId?: string;
  blocksMovement: boolean;
  hidden: boolean;
  roomId?: string;
  ownerId?: string;
}

export interface FloorMap {
  id: string;
  index: number;
  name: LocalizedText;
  width: number;
  height: number;
  tiles: Tile[];
  walls: WallEdge[];
  rooms: Room[];
  entities: SceneEntity[];
  portals: Portal[];
}

export interface AdventureDocument {
  id: string;
  schemaVersion: string;
  generatorVersion: string;
  version: number;
  seed: number;
  name: LocalizedText;
  spec: AdventureSpec;
  summary: LocalizedText;
  narrative: {
    hook: LocalizedText;
    objective: LocalizedText;
    antagonist: LocalizedText;
    atmosphere: LocalizedText;
  };
  floors: FloorMap[];
  encounters: Array<{
    id: string;
    floorId: string;
    roomId: string;
    difficulty: string;
    totalXp: number;
  }>;
  analysis: {
    totalRooms: number;
    totalFloors: number;
    criticalPath: string[];
    deadEnds: string[];
    estimatedDifficulty: string;
    invariants: string[];
  };
  attributions: Array<{
    title: string;
    creator: string;
    source: string;
    license: string;
    notice: string;
  }>;
  createdAt: string;
  updatedAt: string;
}

export interface AdventureSnapshotSummary {
  id: string;
  adventureId: string;
  version: number;
  reason: string;
  name: LocalizedText;
  createdAt: string;
}

export interface GenerationResult {
  run: {
    id: string;
    status: string;
    progress: number;
    seed: number;
    adventureId: string;
  };
  adventure: AdventureDocument;
}

export interface Participant {
  id: string;
  name: string;
  role: "gm" | "player" | "display";
  joinedAt: string;
}

export interface DiceRoll {
  id: string;
  actorId: string;
  expression: string;
  values: number[];
  modifier: number;
  total: number;
  visibility: "public" | "gm" | "private";
  targetId?: string;
  createdAt: string;
}

export interface SessionState {
  id: string;
  adventureId: string;
  revision: number;
  activeFloorId: string;
  participants: Record<string, Participant>;
  tokenPositions: Record<string, GridPosition>;
  tokenFloors: Record<string, string>;
  tokenOwners: Record<string, string>;
  revealedCells: Record<string, GridPosition[]>;
  initiative: {
    entries: Array<{ tokenId: string; name: string; score: number }>;
    activeIndex: number;
    round: number;
  };
  rolls: DiceRoll[];
  open: boolean;
  createdAt: string;
}

export interface CreatedSession {
  session: {
    sessionId: string;
    code: string;
    expiresAt: string;
    token: string;
    state: SessionState;
    adventure: AdventureDocument;
  };
  joinUrls: string[];
}

export interface JoinedSession {
  sessionId: string;
  participantId: string;
  token: string;
  state: SessionState;
  adventure: AdventureDocument;
}

export interface SessionEvent {
  revision: number;
  type: string;
  actorId: string;
  occurredAt: string;
  payload: Record<string, unknown>;
}

export interface SessionCommand {
  id: string;
  expectedRevision: number;
  type: string;
  payload: Record<string, unknown>;
}

export const DEFAULT_SPEC: AdventureSpec = {
  partySize: 4,
  partyLevel: 5,
  durationHours: 4,
  difficulty: "medium",
  theme: "forgotten temple",
  biome: "underground",
  floorCount: 2,
  objective: "stop the awakening ritual",
  antagonist: "serpent cult",
  structureStyle: "branching",
  treasureQuality: "standard",
  useAI: false,
};
