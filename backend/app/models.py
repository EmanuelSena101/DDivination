"""Domain models for DDivination."""
from __future__ import annotations

from enum import Enum
from typing import Optional

from pydantic import BaseModel, Field


# ── Enums ──────────────────────────────────────────────────────────────────

class Theme(str, Enum):
    TEMPLE = "temple"
    CRYPT = "crypt"
    CAVE = "cave"
    RUINS = "ruins"
    FORTRESS = "fortress"
    CULT = "cult"
    SEWER = "sewer"
    MINE = "mine"


class Biome(str, Enum):
    SWAMP = "swamp"
    AQUATIC = "aquatic"
    FOREST = "forest"
    VOLCANIC = "volcanic"
    SHADOW = "shadow"
    UNDERGROUND = "underground"
    ARCTIC = "arctic"
    DESERT = "desert"


class StructureStyle(str, Enum):
    LINEAR = "linear"
    BRANCHING = "branching"
    LABYRINTHINE = "labyrinthine"


class Difficulty(str, Enum):
    EASY = "easy"
    MEDIUM = "medium"
    HARD = "hard"
    DEADLY = "deadly"


class TrapDensity(str, Enum):
    NONE = "none"
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"


class TreasureQuality(str, Enum):
    POOR = "poor"
    STANDARD = "standard"
    RICH = "rich"
    LEGENDARY = "legendary"


class CombatRole(str, Enum):
    MINION = "minion"
    BRUTE = "brute"
    SKIRMISHER = "skirmisher"
    CONTROLLER = "controller"
    SNIPER = "sniper"
    BOSS = "boss"


class RoomRole(str, Enum):
    ENTRANCE = "entrance"
    CORRIDOR = "corridor"
    SHRINE = "shrine"
    LAIR = "lair"
    VAULT = "vault"
    TRAP_ROOM = "trap_room"
    SECRET_ROOM = "secret_room"
    BOSS_ROOM = "boss_room"
    REST_AREA = "rest_area"
    GUARD_POST = "guard_post"
    PUZZLE_ROOM = "puzzle_room"
    ARMORY = "armory"


class DungeonSize(str, Enum):
    SMALL = "small"      # 5-7 rooms
    MEDIUM = "medium"    # 8-12 rooms
    LARGE = "large"      # 13-18 rooms
    EPIC = "epic"        # 19-25 rooms


# ── Request / Config Models ───────────────────────────────────────────────

class DungeonConfig(BaseModel):
    """Input configuration for dungeon generation."""
    party_size: int = Field(default=4, ge=1, le=8, description="Number of party members")
    party_level: int = Field(default=5, ge=1, le=20, description="Average party level")
    dungeon_size: DungeonSize = Field(default=DungeonSize.MEDIUM)
    difficulty: Difficulty = Field(default=Difficulty.MEDIUM)
    theme: Theme = Field(default=Theme.CRYPT)
    biome: Biome = Field(default=Biome.UNDERGROUND)
    structure_style: StructureStyle = Field(default=StructureStyle.BRANCHING)
    trap_density: TrapDensity = Field(default=TrapDensity.MEDIUM)
    treasure_quality: TreasureQuality = Field(default=TreasureQuality.STANDARD)
    boss_type: Optional[str] = Field(default=None, description="Monster family for the boss (e.g. 'dragon', 'undead')")
    seed: Optional[int] = Field(default=None, description="Random seed for reproducibility")
    name: Optional[str] = Field(default=None, description="Optional dungeon name")


# ── Monster / Item domain models ──────────────────────────────────────────

class Monster(BaseModel):
    """Normalized monster from D&D 5e API + enrichment."""
    index: str
    name: str
    size: str
    monster_type: str
    alignment: str
    challenge_rating: float
    hit_points: int
    armor_class: int
    xp: int
    # Enrichment fields
    combat_role: CombatRole = CombatRole.MINION
    theme_tags: list[str] = Field(default_factory=list)
    biome_tags: list[str] = Field(default_factory=list)
    boss_suitable: bool = False
    source: str = "dnd5eapi"


class MagicItem(BaseModel):
    """Normalized magic item from D&D 5e API + enrichment."""
    index: str
    name: str
    rarity: str
    category: str
    description: str = ""
    # Enrichment
    theme_tags: list[str] = Field(default_factory=list)
    reward_weight: float = 1.0
    source: str = "dnd5eapi"


class Equipment(BaseModel):
    """Normalized equipment from D&D 5e API."""
    index: str
    name: str
    category: str
    cost_gp: float = 0.0
    weight: float = 0.0
    description: str = ""
    source: str = "dnd5eapi"


# ── Trap model ────────────────────────────────────────────────────────────

class Trap(BaseModel):
    """Local trap definition."""
    name: str
    description: str
    damage_dice: str
    save_dc: int
    save_type: str = "DEX"
    danger_level: Difficulty = Difficulty.MEDIUM
    theme_tags: list[str] = Field(default_factory=list)
    biome_tags: list[str] = Field(default_factory=list)


# ── Room / Encounter / Dungeon output models ─────────────────────────────

class EncounterMonster(BaseModel):
    """A monster placed in an encounter with a count."""
    monster: Monster
    count: int = 1


class Encounter(BaseModel):
    """A combat encounter in a room."""
    monsters: list[EncounterMonster] = Field(default_factory=list)
    total_xp: int = 0
    adjusted_xp: int = 0
    difficulty_rating: str = "easy"
    description: str = ""


class RoomTreasure(BaseModel):
    """Treasure placed in a room."""
    gold: int = 0
    items: list[MagicItem] = Field(default_factory=list)
    equipment: list[Equipment] = Field(default_factory=list)
    description: str = ""


class Room(BaseModel):
    """A room in the dungeon."""
    room_id: int
    name: str
    role: RoomRole
    description: str = ""
    encounter: Optional[Encounter] = None
    trap: Optional[Trap] = None
    treasure: Optional[RoomTreasure] = None
    flavor_text: str = ""
    difficulty_score: float = 0.0
    is_boss_room: bool = False
    is_secret: bool = False


class DungeonEdge(BaseModel):
    """A connection between two rooms."""
    from_room: int
    to_room: int
    description: str = "passage"
    is_locked: bool = False
    is_hidden: bool = False


class DungeonAnalysis(BaseModel):
    """Analysis results for a generated dungeon."""
    total_rooms: int = 0
    total_encounters: int = 0
    total_traps: int = 0
    total_xp: int = 0
    estimated_difficulty: str = "medium"
    difficulty_by_room: dict[int, float] = Field(default_factory=dict)
    difficulty_progression: list[float] = Field(default_factory=list)
    critical_path: list[int] = Field(default_factory=list)
    critical_path_length: int = 0
    dead_ends: list[int] = Field(default_factory=list)
    branching_factor: float = 0.0
    total_gold: int = 0
    total_magic_items: int = 0
    risk_reward_by_room: dict[int, dict[str, float]] = Field(default_factory=dict)
    risk_reward_balance: str = "balanced"
    has_boss: bool = False
    boss_room_id: Optional[int] = None
    avg_room_difficulty: float = 0.0
    max_room_difficulty: float = 0.0
    pacing_notes: list[str] = Field(default_factory=list)


class Dungeon(BaseModel):
    """A complete generated dungeon."""
    id: Optional[str] = Field(default=None, description="Persistent short id assigned by the server on save")
    name: str
    seed: int
    config: DungeonConfig
    summary: str = ""
    rooms: list[Room] = Field(default_factory=list)
    edges: list[DungeonEdge] = Field(default_factory=list)
    analysis: Optional[DungeonAnalysis] = None
    narrative_intro: str = ""
    narrative_hook: str = ""
    created_at: Optional[str] = Field(default=None, description="ISO-8601 timestamp when persisted")
    favorite: bool = Field(default=False, description="Marked as a favorite by the GM")
    notes: Optional[str] = Field(default=None, description="GM-only notes attached after play")


class DungeonListItem(BaseModel):
    """Lightweight summary row for history list views."""
    id: str
    name: str
    seed: int
    created_at: str
    favorite: bool = False
    theme: str
    biome: str
    party_size: int
    party_level: int
    summary: str = ""
    estimated_difficulty: Optional[str] = None
    total_rooms: int = 0


class DungeonUpdate(BaseModel):
    """PATCH payload for history rows."""
    favorite: Optional[bool] = None
    notes: Optional[str] = None
    name: Optional[str] = None


# ── Sync status ───────────────────────────────────────────────────────────

class SyncStatus(BaseModel):
    """Status of data synchronization."""
    monsters_count: int = 0
    magic_items_count: int = 0
    equipment_count: int = 0
    last_sync: Optional[str] = None
    is_synced: bool = False
