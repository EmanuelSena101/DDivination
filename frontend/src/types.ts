// DDivination type definitions

export interface DungeonConfig {
  party_size: number;
  party_level: number;
  dungeon_size: string;
  difficulty: string;
  theme: string;
  biome: string;
  structure_style: string;
  trap_density: string;
  treasure_quality: string;
  boss_type: string | null;
  seed: number | null;
  name: string | null;
}

export interface Monster {
  index: string;
  name: string;
  size: string;
  monster_type: string;
  alignment: string;
  challenge_rating: number;
  hit_points: number;
  armor_class: number;
  xp: number;
  combat_role: string;
  theme_tags: string[];
  biome_tags: string[];
  boss_suitable: boolean;
}

export interface MagicItem {
  index: string;
  name: string;
  rarity: string;
  category: string;
  description: string;
  reward_weight: number;
}

export interface Equipment {
  index: string;
  name: string;
  category: string;
  cost_gp: number;
  weight: number;
}

export interface Trap {
  name: string;
  description: string;
  damage_dice: string;
  save_dc: number;
  save_type: string;
  danger_level: string;
}

export interface EncounterMonster {
  monster: Monster;
  count: number;
}

export interface Encounter {
  monsters: EncounterMonster[];
  total_xp: number;
  adjusted_xp: number;
  difficulty_rating: string;
  description: string;
}

export interface RoomTreasure {
  gold: number;
  items: MagicItem[];
  equipment: Equipment[];
  description: string;
}

export interface Room {
  room_id: number;
  name: string;
  role: string;
  description: string;
  encounter: Encounter | null;
  trap: Trap | null;
  treasure: RoomTreasure | null;
  flavor_text: string;
  difficulty_score: number;
  is_boss_room: boolean;
  is_secret: boolean;
}

export interface DungeonEdge {
  from_room: number;
  to_room: number;
  description: string;
  is_locked: boolean;
  is_hidden: boolean;
}

export interface DungeonAnalysis {
  total_rooms: number;
  total_encounters: number;
  total_traps: number;
  total_xp: number;
  estimated_difficulty: string;
  difficulty_by_room: Record<number, number>;
  difficulty_progression: number[];
  critical_path: number[];
  critical_path_length: number;
  dead_ends: number[];
  branching_factor: number;
  total_gold: number;
  total_magic_items: number;
  risk_reward_by_room: Record<number, { risk: number; reward: number }>;
  risk_reward_balance: string;
  has_boss: boolean;
  boss_room_id: number | null;
  avg_room_difficulty: number;
  max_room_difficulty: number;
  pacing_notes: string[];
}

export interface Dungeon {
  name: string;
  seed: number;
  config: DungeonConfig;
  summary: string;
  rooms: Room[];
  edges: DungeonEdge[];
  analysis: DungeonAnalysis | null;
  narrative_intro: string;
  narrative_hook: string;
}

export interface SyncStatus {
  monsters_count: number;
  magic_items_count: number;
  equipment_count: number;
  last_sync: string | null;
  is_synced: boolean;
}

export interface SelectOption {
  value: string;
  label: string;
  description?: string;
}

export interface BuilderOptions {
  themes: SelectOption[];
  biomes: SelectOption[];
  structure_styles: SelectOption[];
  difficulties: SelectOption[];
  dungeon_sizes: SelectOption[];
  trap_densities: SelectOption[];
  treasure_qualities: SelectOption[];
}

// Tactical battle grid types

export interface GridPosition {
  x: number;
  y: number;
}

export interface TacticalEnemy {
  name: string;
  count: number;
  combat_role: string;
  challenge_rating: number;
  hit_points: number;
  armor_class: number;
  size: string;
  is_boss: boolean;
  positions: GridPosition[];
}

export interface TacticalObstacle {
  kind: string;
  x: number;
  y: number;
  width: number;
  height: number;
  blocks_movement: boolean;
  provides_cover: boolean;
}

export interface TacticalDoor {
  x: number;
  y: number;
  wall: string;
  is_locked: boolean;
  is_hidden: boolean;
}

export interface TacticalTrap {
  name: string;
  x: number;
  y: number;
  radius: number;
  save_dc: number;
  damage_dice: string;
}

export interface TacticalRoomLayout {
  room_id: number;
  room_name: string;
  room_role: string;
  grid_width: number;
  grid_height: number;
  enemies: TacticalEnemy[];
  obstacles: TacticalObstacle[];
  doors: TacticalDoor[];
  traps: TacticalTrap[];
  is_boss_room: boolean;
  difficulty_score: number;
  description: string;
}

export const DEFAULT_CONFIG: DungeonConfig = {
  party_size: 4,
  party_level: 5,
  dungeon_size: "medium",
  difficulty: "medium",
  theme: "crypt",
  biome: "underground",
  structure_style: "branching",
  trap_density: "medium",
  treasure_quality: "standard",
  boss_type: null,
  seed: null,
  name: null,
};
