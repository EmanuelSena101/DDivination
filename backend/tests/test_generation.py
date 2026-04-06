"""Tests for dungeon generation and analysis logic."""
from __future__ import annotations

import json
import os
import tempfile

import pytest

# Set up test data dir before imports
_test_data_dir = tempfile.mkdtemp()
os.environ["DDIVINATION_DATA_DIR"] = _test_data_dir

from app.analysis.analyzer import analyze_dungeon
from app.data.store import save_equipment, save_magic_items, save_monsters, mark_synced
from app.enrichment.tagger import enrich_monster, enrich_magic_item
from app.generation.dungeon_generator import generate_dungeon
from app.models import (
    Biome,
    Difficulty,
    DungeonConfig,
    DungeonSize,
    StructureStyle,
    Theme,
    TrapDensity,
    TreasureQuality,
)


# ── Test fixtures ─────────────────────────────────────────────────────────

def _create_test_monsters() -> list[dict]:
    """Create a realistic set of test monsters."""
    monsters = [
        {"index": "zombie", "name": "Zombie", "size": "Medium", "monster_type": "undead", "alignment": "neutral evil", "challenge_rating": 0.25, "hit_points": 22, "armor_class": 8, "xp": 50},
        {"index": "skeleton", "name": "Skeleton", "size": "Medium", "monster_type": "undead", "alignment": "lawful evil", "challenge_rating": 0.25, "hit_points": 13, "armor_class": 13, "xp": 50},
        {"index": "ghoul", "name": "Ghoul", "size": "Medium", "monster_type": "undead", "alignment": "chaotic evil", "challenge_rating": 1, "hit_points": 22, "armor_class": 12, "xp": 200},
        {"index": "ogre", "name": "Ogre", "size": "Large", "monster_type": "giant", "alignment": "chaotic evil", "challenge_rating": 2, "hit_points": 59, "armor_class": 11, "xp": 450},
        {"index": "owlbear", "name": "Owlbear", "size": "Large", "monster_type": "beast", "alignment": "unaligned", "challenge_rating": 3, "hit_points": 59, "armor_class": 13, "xp": 700},
        {"index": "wraith", "name": "Wraith", "size": "Medium", "monster_type": "undead", "alignment": "neutral evil", "challenge_rating": 5, "hit_points": 67, "armor_class": 13, "xp": 1800},
        {"index": "mummy", "name": "Mummy", "size": "Medium", "monster_type": "undead", "alignment": "lawful evil", "challenge_rating": 3, "hit_points": 58, "armor_class": 11, "xp": 700},
        {"index": "vampire-spawn", "name": "Vampire Spawn", "size": "Medium", "monster_type": "undead", "alignment": "neutral evil", "challenge_rating": 5, "hit_points": 82, "armor_class": 15, "xp": 1800},
        {"index": "young-black-dragon", "name": "Young Black Dragon", "size": "Large", "monster_type": "dragon", "alignment": "chaotic evil", "challenge_rating": 7, "hit_points": 127, "armor_class": 18, "xp": 2900},
        {"index": "lich", "name": "Lich", "size": "Medium", "monster_type": "undead", "alignment": "any evil alignment", "challenge_rating": 21, "hit_points": 135, "armor_class": 17, "xp": 33000},
        {"index": "goblin", "name": "Goblin", "size": "Small", "monster_type": "humanoid", "alignment": "neutral evil", "challenge_rating": 0.25, "hit_points": 7, "armor_class": 15, "xp": 50},
        {"index": "wolf", "name": "Wolf", "size": "Medium", "monster_type": "beast", "alignment": "unaligned", "challenge_rating": 0.25, "hit_points": 11, "armor_class": 13, "xp": 50},
        {"index": "fire-elemental", "name": "Fire Elemental", "size": "Large", "monster_type": "elemental", "alignment": "neutral", "challenge_rating": 5, "hit_points": 102, "armor_class": 13, "xp": 1800},
        {"index": "adult-red-dragon", "name": "Adult Red Dragon", "size": "Huge", "monster_type": "dragon", "alignment": "chaotic evil", "challenge_rating": 17, "hit_points": 256, "armor_class": 19, "xp": 18000},
        {"index": "gelatinous-cube", "name": "Gelatinous Cube", "size": "Large", "monster_type": "ooze", "alignment": "unaligned", "challenge_rating": 2, "hit_points": 84, "armor_class": 6, "xp": 450},
    ]
    return [enrich_monster(m) for m in monsters]


def _create_test_magic_items() -> list[dict]:
    """Create test magic items."""
    items = [
        {"index": "potion-of-healing", "name": "Potion of Healing", "rarity": "common", "category": "potion", "description": "Heals 2d4+2 hit points."},
        {"index": "bag-of-holding", "name": "Bag of Holding", "rarity": "uncommon", "category": "wondrous item", "description": "This bag has an interior space considerably larger than its outside dimensions."},
        {"index": "flame-tongue", "name": "Flame Tongue", "rarity": "rare", "category": "weapon", "description": "A sword that bursts into flame on command."},
        {"index": "ring-of-protection", "name": "Ring of Protection", "rarity": "rare", "category": "ring", "description": "+1 bonus to AC and saving throws."},
        {"index": "staff-of-power", "name": "Staff of Power", "rarity": "very rare", "category": "staff", "description": "A powerful staff for spellcasters."},
    ]
    return [enrich_magic_item(i) for i in items]


def _create_test_equipment() -> list[dict]:
    """Create test equipment."""
    return [
        {"index": "longsword", "name": "Longsword", "category": "martial melee weapons", "cost_gp": 15.0, "weight": 3.0},
        {"index": "chain-mail", "name": "Chain Mail", "category": "heavy armor", "cost_gp": 75.0, "weight": 55.0},
        {"index": "shield", "name": "Shield", "category": "armor", "cost_gp": 10.0, "weight": 6.0},
    ]


@pytest.fixture(autouse=True)
def setup_test_data():
    """Set up test data in the local store before each test."""
    save_monsters(_create_test_monsters())
    save_magic_items(_create_test_magic_items())
    save_equipment(_create_test_equipment())
    mark_synced()
    yield


# ── Generation tests ──────────────────────────────────────────────────────

class TestDungeonGeneration:
    def test_basic_generation(self):
        """Test that a dungeon can be generated with default config."""
        config = DungeonConfig(seed=42)
        dungeon = generate_dungeon(config)

        assert dungeon.name
        assert dungeon.seed == 42
        assert len(dungeon.rooms) > 0
        assert len(dungeon.edges) > 0
        assert dungeon.narrative_intro
        assert dungeon.narrative_hook

    def test_seed_reproducibility(self):
        """Same seed should produce the same dungeon."""
        config = DungeonConfig(seed=12345)
        dungeon1 = generate_dungeon(config)
        dungeon2 = generate_dungeon(config)

        assert dungeon1.name == dungeon2.name
        assert len(dungeon1.rooms) == len(dungeon2.rooms)
        for r1, r2 in zip(dungeon1.rooms, dungeon2.rooms):
            assert r1.name == r2.name
            assert r1.role == r2.role

    def test_different_seeds_produce_different_dungeons(self):
        """Different seeds should produce different dungeons."""
        config1 = DungeonConfig(seed=100)
        config2 = DungeonConfig(seed=200)
        d1 = generate_dungeon(config1)
        d2 = generate_dungeon(config2)

        # At least the rooms or layout should differ
        names1 = [r.name for r in d1.rooms]
        names2 = [r.name for r in d2.rooms]
        assert names1 != names2 or len(d1.rooms) != len(d2.rooms)

    def test_dungeon_has_entrance_and_boss(self):
        """Every dungeon should have an entrance and a boss room."""
        config = DungeonConfig(seed=42)
        dungeon = generate_dungeon(config)

        roles = [r.role.value for r in dungeon.rooms]
        assert roles[0] == "entrance"
        assert roles[-1] == "boss_room"

    def test_small_dungeon_size(self):
        """Small dungeon should have 5-7 rooms."""
        config = DungeonConfig(seed=42, dungeon_size=DungeonSize.SMALL)
        dungeon = generate_dungeon(config)
        assert 5 <= len(dungeon.rooms) <= 7

    def test_large_dungeon_size(self):
        """Large dungeon should have 13-18 rooms."""
        config = DungeonConfig(seed=42, dungeon_size=DungeonSize.LARGE)
        dungeon = generate_dungeon(config)
        assert 13 <= len(dungeon.rooms) <= 18

    def test_epic_dungeon_size(self):
        """Epic dungeon should have 19-25 rooms."""
        config = DungeonConfig(seed=42, dungeon_size=DungeonSize.EPIC)
        dungeon = generate_dungeon(config)
        assert 19 <= len(dungeon.rooms) <= 25

    def test_linear_structure(self):
        """Linear structure should have sequential edges."""
        config = DungeonConfig(seed=42, structure_style=StructureStyle.LINEAR, dungeon_size=DungeonSize.SMALL)
        dungeon = generate_dungeon(config)

        # In linear layout, edges should form a chain
        assert len(dungeon.edges) == len(dungeon.rooms) - 1

    def test_labyrinthine_structure_has_more_edges(self):
        """Labyrinthine structure should have more connections than linear."""
        config_linear = DungeonConfig(seed=42, structure_style=StructureStyle.LINEAR, dungeon_size=DungeonSize.MEDIUM)
        config_lab = DungeonConfig(seed=42, structure_style=StructureStyle.LABYRINTHINE, dungeon_size=DungeonSize.MEDIUM)
        d_linear = generate_dungeon(config_linear)
        d_lab = generate_dungeon(config_lab)

        assert len(d_lab.edges) >= len(d_linear.edges)

    def test_different_themes(self):
        """Different themes should produce thematically different dungeons."""
        config_crypt = DungeonConfig(seed=42, theme=Theme.CRYPT)
        config_cave = DungeonConfig(seed=42, theme=Theme.CAVE)
        d_crypt = generate_dungeon(config_crypt)
        d_cave = generate_dungeon(config_cave)

        assert "crypt" in d_crypt.name.lower() or "tomb" in d_crypt.name.lower() or "sepulcher" in d_crypt.name.lower() or "mausoleum" in d_crypt.name.lower()
        assert "cave" in d_cave.name.lower() or "cavern" in d_cave.name.lower() or "depth" in d_cave.name.lower()

    def test_boss_room_has_encounter(self):
        """The boss room should always have an encounter."""
        config = DungeonConfig(seed=42)
        dungeon = generate_dungeon(config)
        boss_room = next(r for r in dungeon.rooms if r.is_boss_room)
        assert boss_room.encounter is not None
        assert len(boss_room.encounter.monsters) > 0

    def test_no_traps_density(self):
        """With trap density none, there should be no traps outside trap rooms."""
        config = DungeonConfig(seed=42, trap_density=TrapDensity.NONE, dungeon_size=DungeonSize.SMALL)
        dungeon = generate_dungeon(config)
        non_trap_rooms = [r for r in dungeon.rooms if r.role.value != "trap_room"]
        traps_in_other_rooms = sum(1 for r in non_trap_rooms if r.trap)
        assert traps_in_other_rooms == 0

    def test_summary_generated(self):
        """Dungeon should have a summary."""
        config = DungeonConfig(seed=42)
        dungeon = generate_dungeon(config)
        assert dungeon.summary
        assert str(config.party_size) in dungeon.summary
        assert str(config.party_level) in dungeon.summary


# ── Analysis tests ────────────────────────────────────────────────────────

class TestDungeonAnalysis:
    def test_basic_analysis(self):
        """Test that analysis produces valid results."""
        config = DungeonConfig(seed=42)
        dungeon = generate_dungeon(config)
        analysis = analyze_dungeon(dungeon)

        assert analysis.total_rooms == len(dungeon.rooms)
        assert analysis.total_rooms > 0
        assert analysis.estimated_difficulty in ["easy", "medium", "hard", "deadly"]

    def test_critical_path_exists(self):
        """Analysis should find a critical path."""
        config = DungeonConfig(seed=42)
        dungeon = generate_dungeon(config)
        analysis = analyze_dungeon(dungeon)

        assert len(analysis.critical_path) > 0
        assert analysis.critical_path[0] == 0  # Starts at entrance
        assert analysis.critical_path_length > 0

    def test_difficulty_progression(self):
        """Analysis should include difficulty for each room."""
        config = DungeonConfig(seed=42)
        dungeon = generate_dungeon(config)
        analysis = analyze_dungeon(dungeon)

        assert len(analysis.difficulty_progression) == len(dungeon.rooms)
        assert len(analysis.difficulty_by_room) == len(dungeon.rooms)

    def test_risk_reward_analysis(self):
        """Analysis should include risk/reward data."""
        config = DungeonConfig(seed=42)
        dungeon = generate_dungeon(config)
        analysis = analyze_dungeon(dungeon)

        assert len(analysis.risk_reward_by_room) == len(dungeon.rooms)
        assert analysis.risk_reward_balance in ["balanced", "risk-heavy (low rewards for the danger)", "reward-heavy (generous loot for the difficulty)"]

    def test_pacing_notes_generated(self):
        """Analysis should generate pacing notes for GMs."""
        config = DungeonConfig(seed=42)
        dungeon = generate_dungeon(config)
        analysis = analyze_dungeon(dungeon)

        assert len(analysis.pacing_notes) > 0

    def test_boss_detection(self):
        """Analysis should detect boss presence."""
        config = DungeonConfig(seed=42)
        dungeon = generate_dungeon(config)
        analysis = analyze_dungeon(dungeon)

        assert analysis.has_boss is True
        assert analysis.boss_room_id is not None

    def test_branching_factor(self):
        """Branching factor should differ between linear and labyrinthine."""
        config_linear = DungeonConfig(seed=42, structure_style=StructureStyle.LINEAR, dungeon_size=DungeonSize.MEDIUM)
        config_lab = DungeonConfig(seed=42, structure_style=StructureStyle.LABYRINTHINE, dungeon_size=DungeonSize.MEDIUM)

        d1 = generate_dungeon(config_linear)
        d2 = generate_dungeon(config_lab)

        a1 = analyze_dungeon(d1)
        a2 = analyze_dungeon(d2)

        assert a2.branching_factor >= a1.branching_factor


# ── Enrichment tests ──────────────────────────────────────────────────────

class TestEnrichment:
    def test_monster_enrichment(self):
        """Monsters should get theme and biome tags."""
        monster = {"index": "zombie", "name": "Zombie", "size": "Medium",
                   "monster_type": "undead", "alignment": "neutral evil",
                   "challenge_rating": 0.25, "hit_points": 22, "armor_class": 8, "xp": 50}
        enriched = enrich_monster(monster)

        assert "crypt" in enriched["theme_tags"]
        assert enriched["combat_role"] == "minion"
        assert enriched["source"] == "dnd5eapi"

    def test_boss_suitability(self):
        """High CR monsters should be boss-suitable."""
        dragon = {"index": "adult-red-dragon", "name": "Adult Red Dragon", "size": "Huge",
                  "monster_type": "dragon", "alignment": "chaotic evil",
                  "challenge_rating": 17, "hit_points": 256, "armor_class": 19, "xp": 18000}
        enriched = enrich_monster(dragon)

        assert enriched["boss_suitable"] is True
        assert enriched["combat_role"] == "boss"

    def test_magic_item_enrichment(self):
        """Magic items should get reward weights based on rarity."""
        item = {"index": "flame-tongue", "name": "Flame Tongue", "rarity": "rare",
                "category": "weapon", "description": "A sword of flame."}
        enriched = enrich_magic_item(item)

        assert enriched["reward_weight"] == 4.0
        assert enriched["source"] == "dnd5eapi"
