"""Procedural dungeon generation engine.

This is the core of DDivination. It takes a DungeonConfig and produces a
complete Dungeon with rooms, encounters, traps, treasure, and narrative.
"""
from __future__ import annotations

import random
from typing import Optional

import networkx as nx

from app.data.store import load_equipment, load_magic_items, load_monsters
from app.enrichment.rooms import (
    BIOME_FLAVOR,
    ROOM_ARCHETYPES,
    THEME_NARRATIVES,
)
from app.enrichment.traps import get_trap_for_difficulty, get_traps_for_theme
from app.models import (
    CombatRole,
    Difficulty,
    Dungeon,
    DungeonConfig,
    DungeonEdge,
    DungeonSize,
    Encounter,
    EncounterMonster,
    Equipment,
    MagicItem,
    Monster,
    Room,
    RoomRole,
    RoomTreasure,
    StructureStyle,
    Trap,
    TrapDensity,
    TreasureQuality,
)


# ── Size -> room count mapping ────────────────────────────────────────────

SIZE_ROOM_RANGES: dict[DungeonSize, tuple[int, int]] = {
    DungeonSize.SMALL: (5, 7),
    DungeonSize.MEDIUM: (8, 12),
    DungeonSize.LARGE: (13, 18),
    DungeonSize.EPIC: (19, 25),
}

# ── XP thresholds by level (per character) ────────────────────────────────

XP_THRESHOLDS: dict[int, dict[str, int]] = {
    1: {"easy": 25, "medium": 50, "hard": 75, "deadly": 100},
    2: {"easy": 50, "medium": 100, "hard": 150, "deadly": 200},
    3: {"easy": 75, "medium": 150, "hard": 225, "deadly": 400},
    4: {"easy": 125, "medium": 250, "hard": 375, "deadly": 500},
    5: {"easy": 250, "medium": 500, "hard": 750, "deadly": 1100},
    6: {"easy": 300, "medium": 600, "hard": 900, "deadly": 1400},
    7: {"easy": 350, "medium": 750, "hard": 1100, "deadly": 1700},
    8: {"easy": 450, "medium": 900, "hard": 1400, "deadly": 2100},
    9: {"easy": 550, "medium": 1100, "hard": 1600, "deadly": 2400},
    10: {"easy": 600, "medium": 1200, "hard": 1900, "deadly": 2800},
    11: {"easy": 800, "medium": 1600, "hard": 2400, "deadly": 3600},
    12: {"easy": 1000, "medium": 2000, "hard": 3000, "deadly": 4500},
    13: {"easy": 1100, "medium": 2200, "hard": 3400, "deadly": 5100},
    14: {"easy": 1250, "medium": 2500, "hard": 3800, "deadly": 5700},
    15: {"easy": 1400, "medium": 2800, "hard": 4300, "deadly": 6400},
    16: {"easy": 1600, "medium": 3200, "hard": 4800, "deadly": 7200},
    17: {"easy": 2000, "medium": 3900, "hard": 5900, "deadly": 8800},
    18: {"easy": 2100, "medium": 4200, "hard": 6300, "deadly": 9500},
    19: {"easy": 2400, "medium": 4900, "hard": 7300, "deadly": 10900},
    20: {"easy": 2800, "medium": 5700, "hard": 8500, "deadly": 12700},
}

# Encounter multipliers by number of monsters
ENCOUNTER_MULTIPLIERS = {
    1: 1.0, 2: 1.5, 3: 2.0, 4: 2.0, 5: 2.0, 6: 2.0,
    7: 2.5, 8: 2.5, 9: 2.5, 10: 2.5, 11: 3.0, 12: 3.0,
    13: 3.0, 14: 3.0, 15: 4.0,
}


def _get_multiplier(count: int) -> float:
    if count <= 0:
        return 1.0
    if count >= 15:
        return 4.0
    return ENCOUNTER_MULTIPLIERS.get(count, 2.0)


def _get_xp_threshold(level: int, party_size: int, diff: str) -> int:
    lvl = max(1, min(20, level))
    thresholds = XP_THRESHOLDS[lvl]
    return thresholds.get(diff, thresholds["medium"]) * party_size


def _classify_encounter_difficulty(adjusted_xp: int, level: int, party_size: int) -> str:
    for diff in ["deadly", "hard", "medium", "easy"]:
        if adjusted_xp >= _get_xp_threshold(level, party_size, diff):
            return diff
    return "easy"


# ── Room role distribution ────────────────────────────────────────────────

def _plan_room_roles(num_rooms: int, rng: random.Random, config: DungeonConfig) -> list[RoomRole]:
    """Plan which role each room will have."""
    roles: list[RoomRole] = []

    # First room is always entrance
    roles.append(RoomRole.ENTRANCE)
    # Last room is always boss room
    # We'll add it at the end

    remaining = num_rooms - 2  # minus entrance and boss

    # Mandatory rooms
    mandatory = []
    if remaining > 2:
        mandatory.append(RoomRole.GUARD_POST)
    if remaining > 3:
        mandatory.append(RoomRole.REST_AREA)

    for r in mandatory:
        roles.append(r)
        remaining -= 1

    # Fill remaining with weighted random selection
    weighted_roles = [
        (RoomRole.CORRIDOR, 3),
        (RoomRole.LAIR, 3),
        (RoomRole.SHRINE, 2),
        (RoomRole.TRAP_ROOM, 2),
        (RoomRole.GUARD_POST, 2),
        (RoomRole.VAULT, 1),
        (RoomRole.PUZZLE_ROOM, 1),
        (RoomRole.ARMORY, 1),
    ]

    # Add secret room chance
    if remaining > 2 and rng.random() < 0.4:
        roles.append(RoomRole.SECRET_ROOM)
        remaining -= 1

    pool = []
    for role, weight in weighted_roles:
        pool.extend([role] * weight)

    for _ in range(remaining):
        role = rng.choice(pool)
        roles.append(role)

    # Shuffle middle rooms (keep entrance first)
    entrance = roles[0]
    middle = roles[1:]
    rng.shuffle(middle)
    roles = [entrance] + middle + [RoomRole.BOSS_ROOM]

    return roles


# ── Graph layout generation ──────────────────────────────────────────────

def _build_linear_graph(num_rooms: int, rng: random.Random) -> list[tuple[int, int]]:
    """Build a linear dungeon graph."""
    edges = []
    for i in range(num_rooms - 1):
        edges.append((i, i + 1))
    return edges


def _build_branching_graph(num_rooms: int, rng: random.Random) -> list[tuple[int, int]]:
    """Build a branching dungeon graph with main path and side branches."""
    edges: list[tuple[int, int]] = []

    # Main path: ~60-70% of rooms
    main_path_len = max(3, int(num_rooms * 0.65))
    main_path = list(range(main_path_len))
    for i in range(len(main_path) - 1):
        edges.append((main_path[i], main_path[i + 1]))

    # Branch rooms off the main path
    remaining_rooms = list(range(main_path_len, num_rooms))
    rng.shuffle(remaining_rooms)

    for room_id in remaining_rooms:
        # Pick a random room on the main path (not the last one) to branch from
        branch_point = rng.choice(main_path[1:-1]) if len(main_path) > 2 else main_path[0]
        edges.append((branch_point, room_id))

    return edges


def _build_labyrinthine_graph(num_rooms: int, rng: random.Random) -> list[tuple[int, int]]:
    """Build a labyrinthine dungeon with many interconnections."""
    # Start with a spanning tree
    edges = _build_branching_graph(num_rooms, rng)
    edge_set = set((min(a, b), max(a, b)) for a, b in edges)

    # Add extra edges for loops (20-30% extra connections)
    extra_count = max(1, int(num_rooms * 0.25))
    attempts = 0
    added = 0
    while added < extra_count and attempts < extra_count * 5:
        a = rng.randint(0, num_rooms - 1)
        b = rng.randint(0, num_rooms - 1)
        if a != b:
            key = (min(a, b), max(a, b))
            if key not in edge_set:
                edges.append((a, b))
                edge_set.add(key)
                added += 1
        attempts += 1

    return edges


# ── Monster selection ─────────────────────────────────────────────────────

def _filter_monsters(
    monsters: list[Monster],
    theme: str,
    biome: str,
    min_cr: float = 0,
    max_cr: float = 30,
    role: Optional[CombatRole] = None,
    boss_only: bool = False,
    boss_type: Optional[str] = None,
) -> list[Monster]:
    """Filter monsters by theme, biome, CR range, role, etc."""
    result = []
    for m in monsters:
        # CR filter
        if m.challenge_rating < min_cr or m.challenge_rating > max_cr:
            continue

        # Boss filter
        if boss_only and not m.boss_suitable:
            continue

        # Boss type filter
        if boss_type and boss_type.lower() not in m.monster_type.lower() and boss_type.lower() not in m.name.lower():
            continue

        # Role filter
        if role and m.combat_role != role.value:
            continue

        # Theme/biome scoring
        score = 0
        if theme in m.theme_tags:
            score += 3
        if biome in m.biome_tags:
            score += 2
        # Always include some monsters even without tags
        if not m.theme_tags and not m.biome_tags:
            score += 0.5

        if score > 0:
            result.append(m)

    return result


def _select_encounter_monsters(
    monsters: list[Monster],
    target_xp: int,
    party_size: int,
    rng: random.Random,
    is_boss: bool = False,
) -> list[EncounterMonster]:
    """Select monsters for an encounter targeting a specific XP budget."""
    if not monsters:
        return []

    if is_boss:
        # For boss encounters, pick one strong monster
        boss_candidates = [m for m in monsters if m.boss_suitable]
        if not boss_candidates:
            boss_candidates = sorted(monsters, key=lambda m: m.challenge_rating, reverse=True)[:5]
        if boss_candidates:
            boss = rng.choice(boss_candidates[:min(5, len(boss_candidates))])
            result = [EncounterMonster(monster=boss, count=1)]
            # Maybe add minions
            remaining_xp = target_xp - boss.xp
            if remaining_xp > 0:
                minions = [m for m in monsters if m.challenge_rating <= boss.challenge_rating / 3 and m.xp > 0]
                if minions:
                    minion = rng.choice(minions)
                    count = min(4, max(1, int(remaining_xp / (minion.xp * 1.5))))
                    if count > 0:
                        result.append(EncounterMonster(monster=minion, count=count))
            return result

    # Regular encounter: pick 1-4 monster types
    sorted_monsters = sorted(monsters, key=lambda m: abs(m.xp - target_xp / 2))
    candidates = sorted_monsters[:max(10, len(sorted_monsters) // 3)]

    if not candidates:
        return []

    result: list[EncounterMonster] = []
    budget = target_xp
    max_types = rng.randint(1, min(3, len(candidates)))

    for _ in range(max_types):
        if budget <= 0 or not candidates:
            break
        monster = rng.choice(candidates)
        max_count = max(1, int(budget / max(1, monster.xp)))
        count = rng.randint(1, min(max_count, 4))
        result.append(EncounterMonster(monster=monster, count=count))
        budget -= monster.xp * count

    return result


# ── Treasure generation ──────────────────────────────────────────────────

GOLD_BY_LEVEL: dict[int, tuple[int, int]] = {
    1: (10, 50), 2: (20, 80), 3: (30, 120), 4: (50, 200),
    5: (100, 500), 6: (150, 600), 7: (200, 800), 8: (300, 1000),
    9: (400, 1500), 10: (500, 2000), 11: (700, 3000), 12: (1000, 4000),
    13: (1500, 5000), 14: (2000, 7000), 15: (3000, 10000), 16: (4000, 12000),
    17: (5000, 15000), 18: (7000, 20000), 19: (10000, 25000), 20: (15000, 40000),
}

RARITY_BY_QUALITY: dict[str, list[str]] = {
    "poor": ["common"],
    "standard": ["common", "uncommon"],
    "rich": ["uncommon", "rare"],
    "legendary": ["rare", "very rare", "legendary"],
}


def _generate_treasure(
    level: int,
    quality: TreasureQuality,
    magic_items: list[MagicItem],
    equipment: list[Equipment],
    theme: str,
    rng: random.Random,
    is_boss: bool = False,
    is_vault: bool = False,
) -> RoomTreasure:
    """Generate treasure for a room."""
    lvl = max(1, min(20, level))
    min_gold, max_gold = GOLD_BY_LEVEL[lvl]

    # Quality multiplier
    quality_mult = {"poor": 0.5, "standard": 1.0, "rich": 2.0, "legendary": 4.0}
    mult = quality_mult.get(quality.value, 1.0)
    if is_boss:
        mult *= 2.5
    if is_vault:
        mult *= 1.8

    gold = int(rng.randint(min_gold, max_gold) * mult)

    # Magic items
    allowed_rarities = RARITY_BY_QUALITY.get(quality.value, ["common"])
    eligible_items = [i for i in magic_items if i.rarity in allowed_rarities]

    # Prefer theme-matching items
    themed_items = [i for i in eligible_items if theme in i.theme_tags]
    if themed_items and rng.random() < 0.6:
        item_pool = themed_items
    else:
        item_pool = eligible_items

    num_items = 0
    if is_boss:
        num_items = rng.randint(1, 3)
    elif is_vault:
        num_items = rng.randint(1, 2)
    elif rng.random() < 0.3 * mult:
        num_items = 1

    selected_items: list[MagicItem] = []
    if item_pool and num_items > 0:
        selected_items = rng.sample(item_pool, min(num_items, len(item_pool)))

    # Equipment
    selected_equip: list[Equipment] = []
    if equipment and rng.random() < 0.4:
        selected_equip = rng.sample(equipment, min(rng.randint(1, 2), len(equipment)))

    desc_parts = []
    if gold:
        desc_parts.append(f"{gold} gold pieces")
    if selected_items:
        desc_parts.append(", ".join(i.name for i in selected_items))
    if selected_equip:
        desc_parts.append(", ".join(e.name for e in selected_equip))

    return RoomTreasure(
        gold=gold,
        items=selected_items,
        equipment=selected_equip,
        description=". ".join(desc_parts) if desc_parts else "A small amount of scattered coins.",
    )


# ── Room difficulty scoring ──────────────────────────────────────────────

def _score_room_difficulty(room: Room) -> float:
    """Score a room's difficulty from 0-10."""
    score = 0.0
    if room.encounter and room.encounter.monsters:
        # Based on encounter difficulty rating
        diff_scores = {"easy": 2.0, "medium": 4.0, "hard": 6.0, "deadly": 8.0}
        score += diff_scores.get(room.encounter.difficulty_rating, 3.0)
    if room.trap:
        trap_scores = {"easy": 1.0, "medium": 2.0, "hard": 3.0, "deadly": 4.0}
        score += trap_scores.get(room.trap.danger_level.value, 2.0)
    if room.is_boss_room:
        score += 2.0
    return min(10.0, score)


# ── Main generation function ─────────────────────────────────────────────

def generate_dungeon(config: DungeonConfig) -> Dungeon:
    """Generate a complete dungeon from configuration."""
    seed = config.seed if config.seed is not None else random.randint(0, 2**31)
    rng = random.Random(seed)

    # Load data
    all_monsters = load_monsters()
    all_magic_items = load_magic_items()
    all_equipment = load_equipment()

    theme = config.theme.value
    biome = config.biome.value

    # 1. Determine room count
    min_rooms, max_rooms = SIZE_ROOM_RANGES[config.dungeon_size]
    num_rooms = rng.randint(min_rooms, max_rooms)

    # 2. Plan room roles
    room_roles = _plan_room_roles(num_rooms, rng, config)

    # 3. Build graph layout
    if config.structure_style == StructureStyle.LINEAR:
        edges_raw = _build_linear_graph(num_rooms, rng)
    elif config.structure_style == StructureStyle.LABYRINTHINE:
        edges_raw = _build_labyrinthine_graph(num_rooms, rng)
    else:
        edges_raw = _build_branching_graph(num_rooms, rng)

    # 4. Filter monsters for this dungeon's theme/biome
    max_cr = config.party_level + 3
    themed_monsters = _filter_monsters(all_monsters, theme, biome, max_cr=max_cr)
    if len(themed_monsters) < 5:
        # Fallback: include all monsters in CR range
        themed_monsters = [m for m in all_monsters if m.challenge_rating <= max_cr]

    boss_monsters = _filter_monsters(
        all_monsters, theme, biome,
        min_cr=max(1, config.party_level - 2),
        max_cr=config.party_level + 5,
        boss_only=True,
        boss_type=config.boss_type,
    )
    if not boss_monsters:
        boss_monsters = sorted(themed_monsters, key=lambda m: m.challenge_rating, reverse=True)[:10]

    # 5. Get traps for this theme
    available_traps = get_traps_for_theme(theme, biome)
    if not available_traps:
        from app.enrichment.traps import TRAP_TEMPLATES
        available_traps = TRAP_TEMPLATES

    difficulty_traps = get_trap_for_difficulty(available_traps, config.difficulty.value)
    if not difficulty_traps:
        difficulty_traps = available_traps

    # 6. Determine trap placement
    trap_chance = {"none": 0.0, "low": 0.15, "medium": 0.3, "high": 0.5}
    trap_prob = trap_chance.get(config.trap_density.value, 0.3)

    # 7. Build rooms
    rooms: list[Room] = []
    difficulty_progression = _plan_difficulty_curve(num_rooms, config.difficulty.value, rng)

    for idx, role in enumerate(room_roles):
        is_boss = role == RoomRole.BOSS_ROOM
        is_vault = role == RoomRole.VAULT
        room_diff = difficulty_progression[idx] if idx < len(difficulty_progression) else "medium"

        # Get archetype data
        archetype = ROOM_ARCHETYPES.get(role, ROOM_ARCHETYPES[RoomRole.CORRIDOR])
        name = rng.choice(archetype["names"])
        description = rng.choice(archetype["descriptions"])

        # Add biome flavor
        biome_flavors = BIOME_FLAVOR.get(biome, [])
        flavor = rng.choice(biome_flavors) if biome_flavors else ""

        # Generate encounter for combat rooms
        encounter = None
        if role in (RoomRole.LAIR, RoomRole.GUARD_POST, RoomRole.BOSS_ROOM, RoomRole.SHRINE) or \
           (role == RoomRole.CORRIDOR and rng.random() < 0.3):
            target_xp = _get_xp_threshold(config.party_level, config.party_size, room_diff)
            if is_boss:
                target_xp = _get_xp_threshold(config.party_level, config.party_size, "deadly")
                monster_pool = boss_monsters if boss_monsters else themed_monsters
            else:
                monster_pool = themed_monsters

            enc_monsters = _select_encounter_monsters(
                monster_pool, target_xp, config.party_size, rng, is_boss=is_boss
            )

            if enc_monsters:
                total_xp = sum(em.monster.xp * em.count for em in enc_monsters)
                total_count = sum(em.count for em in enc_monsters)
                multiplier = _get_multiplier(total_count)
                adjusted_xp = int(total_xp * multiplier)
                diff_rating = _classify_encounter_difficulty(
                    adjusted_xp, config.party_level, config.party_size
                )

                monster_names = ", ".join(
                    f"{em.count}x {em.monster.name}" for em in enc_monsters
                )
                encounter = Encounter(
                    monsters=enc_monsters,
                    total_xp=total_xp,
                    adjusted_xp=adjusted_xp,
                    difficulty_rating=diff_rating,
                    description=f"Encounter: {monster_names}",
                )

        # Generate trap
        trap = None
        if role == RoomRole.TRAP_ROOM:
            trap = rng.choice(difficulty_traps) if difficulty_traps else None
        elif role not in (RoomRole.ENTRANCE, RoomRole.REST_AREA, RoomRole.BOSS_ROOM) and rng.random() < trap_prob:
            trap = rng.choice(difficulty_traps) if difficulty_traps else None

        # Generate treasure
        treasure = None
        if is_boss or is_vault or role == RoomRole.ARMORY:
            treasure = _generate_treasure(
                config.party_level, config.treasure_quality,
                all_magic_items, all_equipment, theme, rng,
                is_boss=is_boss, is_vault=is_vault,
            )
        elif role == RoomRole.SECRET_ROOM:
            treasure = _generate_treasure(
                config.party_level, TreasureQuality.RICH,
                all_magic_items, all_equipment, theme, rng,
            )
        elif role not in (RoomRole.ENTRANCE, RoomRole.CORRIDOR, RoomRole.TRAP_ROOM) and rng.random() < 0.35:
            treasure = _generate_treasure(
                config.party_level, config.treasure_quality,
                all_magic_items, all_equipment, theme, rng,
            )

        room = Room(
            room_id=idx,
            name=name,
            role=role,
            description=description,
            encounter=encounter,
            trap=trap,
            treasure=treasure,
            flavor_text=flavor,
            is_boss_room=is_boss,
            is_secret=role == RoomRole.SECRET_ROOM,
        )
        room.difficulty_score = _score_room_difficulty(room)
        rooms.append(room)

    # 8. Build edges
    edges = []
    for from_id, to_id in edges_raw:
        is_locked = False
        is_hidden = False
        desc = "passage"

        if rooms[to_id].role == RoomRole.SECRET_ROOM:
            is_hidden = True
            desc = "hidden passage"
        elif rooms[to_id].role == RoomRole.BOSS_ROOM:
            is_locked = rng.random() < 0.5
            desc = "heavy door" if is_locked else "ornate archway"
        elif rooms[to_id].role == RoomRole.VAULT:
            is_locked = True
            desc = "locked iron door"
        elif rng.random() < 0.2:
            desc = rng.choice(["narrow corridor", "stone archway", "wooden door", "crumbling passage"])

        edges.append(DungeonEdge(
            from_room=from_id,
            to_room=to_id,
            description=desc,
            is_locked=is_locked,
            is_hidden=is_hidden,
        ))

    # 9. Generate narrative
    dungeon_name = config.name or _generate_dungeon_name(theme, biome, rng)
    narratives = THEME_NARRATIVES.get(theme, THEME_NARRATIVES["cave"])
    intro = rng.choice(narratives["intro"]).format(biome=biome)
    hook = rng.choice(narratives["hook"]).format(biome=biome)

    summary = _generate_summary(config, rooms, rng)

    dungeon = Dungeon(
        name=dungeon_name,
        seed=seed,
        config=config,
        summary=summary,
        rooms=rooms,
        edges=edges,
        narrative_intro=intro,
        narrative_hook=hook,
    )

    return dungeon


def _plan_difficulty_curve(num_rooms: int, target_diff: str, rng: random.Random) -> list[str]:
    """Plan a difficulty curve that generally escalates toward the boss."""
    diff_levels = ["easy", "medium", "hard", "deadly"]
    target_idx = diff_levels.index(target_diff)

    curve = []
    for i in range(num_rooms):
        progress = i / max(1, num_rooms - 1)

        if i == 0:
            # Entrance is easy
            curve.append("easy")
        elif i == num_rooms - 1:
            # Boss room is always at or above target
            curve.append(diff_levels[min(target_idx + 1, 3)])
        else:
            # Escalating with some variation
            base_idx = int(progress * (target_idx + 1))
            variation = rng.randint(-1, 1)
            final_idx = max(0, min(3, base_idx + variation))
            curve.append(diff_levels[final_idx])

    return curve


def _generate_dungeon_name(theme: str, biome: str, rng: random.Random) -> str:
    """Generate a thematic dungeon name."""
    prefixes = {
        "temple": ["The Lost Temple", "The Desecrated Shrine", "The Sunken Sanctuary", "The Forsaken Cathedral"],
        "crypt": ["The Forgotten Crypt", "The Tomb of Shadows", "The Whispering Sepulcher", "The Cursed Mausoleum"],
        "cave": ["The Deep Caverns", "The Echoing Depths", "The Sunless Caves", "The Howling Caverns"],
        "ruins": ["The Ancient Ruins", "The Crumbling Citadel", "The Shattered Halls", "The Fallen Keep"],
        "fortress": ["The Iron Fortress", "The Dark Bastion", "The War Keep", "The Siege Tower"],
        "cult": ["The Cult's Sanctum", "The Dark Circle", "The Ritual Chambers", "The Profane Hall"],
        "sewer": ["The Drowned Tunnels", "The Festering Warren", "The Undercity Depths", "The Rat King's Domain"],
        "mine": ["The Abandoned Mine", "The Deep Shaft", "The Crystal Tunnels", "The Lost Excavation"],
    }

    suffixes = {
        "swamp": "of the Mire",
        "aquatic": "of the Deep",
        "forest": "of the Wild",
        "volcanic": "of Ash and Fire",
        "shadow": "of Eternal Night",
        "underground": "of the Underdark",
        "arctic": "of the Frozen Wastes",
        "desert": "of the Burning Sands",
    }

    prefix = rng.choice(prefixes.get(theme, ["The Dungeon"]))
    suffix = suffixes.get(biome, "")
    return f"{prefix} {suffix}".strip()


def _generate_summary(config: DungeonConfig, rooms: list[Room], rng: random.Random) -> str:
    """Generate a high-level adventure summary."""
    encounter_count = sum(1 for r in rooms if r.encounter)
    trap_count = sum(1 for r in rooms if r.trap)
    treasure_count = sum(1 for r in rooms if r.treasure)
    boss_room = next((r for r in rooms if r.is_boss_room), None)
    boss_name = ""
    if boss_room and boss_room.encounter and boss_room.encounter.monsters:
        boss_name = boss_room.encounter.monsters[0].monster.name

    parts = [
        f"A {config.difficulty.value}-difficulty {config.theme.value} dungeon",
        f"set in a {config.biome.value} environment.",
        f"Designed for {config.party_size} adventurers of level {config.party_level}.",
        f"The dungeon contains {len(rooms)} rooms with a {config.structure_style.value} layout,",
        f"featuring {encounter_count} combat encounters, {trap_count} traps, and {treasure_count} treasure hoards.",
    ]
    if boss_name:
        parts.append(f"The final challenge is a {boss_name} awaiting in the {boss_room.name if boss_room else 'final chamber'}.")

    return " ".join(parts)
