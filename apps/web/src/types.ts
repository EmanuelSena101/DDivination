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
  id?: string;
  direction: "north" | "east" | "south" | "west";
  kind: "wall" | "door" | "secret-door";
  open: boolean;
  locked: boolean;
  requiredKeyId?: string;
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
  rulesVersion?: string;
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
  encounters: Encounter[];
  treasures: Treasure[];
  puzzles: Puzzle[];
  traps: Trap[];
  restPoints: RestPoint[];
  progression: DungeonProgression;
  analysis: {
    totalRooms: number;
    totalFloors: number;
    criticalPath: string[];
    deadEnds: string[];
    estimatedDifficulty: string;
    encounterBudgetXp: number;
    encounterTotalXp: number;
    treasureValueGp: number;
    contentCounts: ContentCounts;
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

export interface EncounterCreature {
  index: string;
  name: LocalizedText;
  count: number;
  cr: number;
  xp: number;
}

export interface Encounter {
  id: string;
  floorId: string;
  roomId: string;
  difficulty: "easy" | "medium" | "hard" | "deadly";
  creatures: EncounterCreature[];
  budgetXp: number;
  budgetTier: "low" | "moderate" | "high";
  totalXp: number;
}

export interface Treasure {
  id: string;
  floorId: string;
  roomId: string;
  name: LocalizedText;
  description: LocalizedText;
  quality: "poor" | "standard" | "rich" | "legendary";
  valueGp: number;
  contents: LocalizedText[];
  source: string;
}

export interface Puzzle {
  id: string;
  floorId: string;
  roomId: string;
  name: LocalizedText;
  prompt: LocalizedText;
  solution: LocalizedText;
  hint: LocalizedText;
  checkDc: number;
  source: string;
}

export interface Trap {
  id: string;
  floorId: string;
  roomId: string;
  catalogIndex: string;
  name: LocalizedText;
  severity: "nuisance" | "deadly";
  levelTier: "1-4" | "5-10" | "11-16" | "17-20";
  trigger: LocalizedText;
  detectionDc: number;
  disableDc: number;
  saveDc: number;
  damage: LocalizedText;
  source: string;
  license: string;
  hidden: boolean;
}

export interface RestPoint {
  id: string;
  floorId: string;
  roomId: string;
  kind: "short";
  name: LocalizedText;
  description: LocalizedText;
  source: string;
}

export interface ContentCounts {
  encounters: number;
  treasures: number;
  puzzles: number;
  traps: number;
  restPoints: number;
}

export interface ProgressionStep {
  order: number;
  floorId: string;
  roomId: string;
  kind: "entrance" | "exploration" | "key" | "transition" | "climax";
  beat: LocalizedText;
  grantsKeyIds: string[];
  requiresKeyIds: string[];
}

export interface ProgressionLock {
  id: string;
  kind: "door" | "portal";
  targetId: string;
  floorId: string;
  fromRoomId: string;
  toRoomId: string;
  keyId: string;
}

export interface DungeonProgression {
  entryRoomId: string;
  objectiveRoomId: string;
  climaxRoomId: string;
  steps: ProgressionStep[];
  locks: ProgressionLock[];
  secretRoomIds: string[];
  solvable: boolean;
}

export interface AdventureSnapshotSummary {
  id: string;
  adventureId: string;
  version: number;
  reason: string;
  name: LocalizedText;
  createdAt: string;
}

export interface GenerationStage {
  name: string;
  progress: number;
  occurredAt: string;
}

export interface GenerationRun {
  id: string;
  status: "queued" | "running" | "completed" | "failed" | "cancelled";
  stage: string;
  progress: number;
  seed: number;
  generatorVersion: string;
  spec: AdventureSpec;
  adventureId?: string;
  diagnostics: string[];
  stages: GenerationStage[];
  createdAt: string;
  updatedAt: string;
  completedAt?: string;
}

export interface Participant {
  id: string;
  name: string;
  role: "gm" | "player" | "display";
  connected: boolean;
  joinedAt: string;
  lastSeenAt: string;
}

export interface SessionPermissions {
  playerCanRevealFog: boolean;
  playerCanPing: boolean;
  playerCanRollDice: boolean;
  playerCanManageInitiative: boolean;
}

export interface AdmissionRequest {
  id: string;
  name: string;
  role: "player" | "display";
  status: "pending" | "approved" | "denied" | "expired";
  createdAt: string;
  expiresAt: string;
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
  joinOpen: boolean;
  approvalRequired: boolean;
  permissions: SessionPermissions;
  admissions?: Record<string, AdmissionRequest>;
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
  status: "joined" | "pending" | "denied" | "expired";
  requestId?: string;
  expiresAt?: string;
  state: SessionState;
  adventure: AdventureDocument;
}

export interface SessionCodeStatus {
  code: string;
  expiresAt: string;
  joinOpen: boolean;
  approvalRequired: boolean;
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
