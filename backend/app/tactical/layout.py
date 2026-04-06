"""Tactical room layout generator.

Generates grid-based tactical data for rooms including:
- Room dimensions and grid
- Enemy placements on the grid
- Obstacle positions (pillars, rubble, altars, etc.)
- Door positions along walls
- Chokepoints and cover positions
"""
from __future__ import annotations

import random

from pydantic import BaseModel, Field

from app.models import Dungeon, DungeonEdge, Room


# ── Tactical Models ──────────────────────────────────────────────────────

class GridPosition(BaseModel):
    x: int
    y: int


class TacticalEnemy(BaseModel):
    name: str
    count: int
    combat_role: str
    challenge_rating: float
    hit_points: int
    armor_class: int
    size: str
    is_boss: bool = False
    positions: list[GridPosition] = Field(default_factory=list)


class TacticalObstacle(BaseModel):
    kind: str  # pillar, rubble, altar, pool, statue, crate, table, web, pit, fire
    x: int
    y: int
    width: int = 1
    height: int = 1
    blocks_movement: bool = True
    provides_cover: bool = True


class TacticalDoor(BaseModel):
    x: int
    y: int
    wall: str  # north, south, east, west
    is_locked: bool = False
    is_hidden: bool = False


class TacticalTrap(BaseModel):
    name: str
    x: int
    y: int
    radius: int = 1
    save_dc: int = 12
    damage_dice: str = "1d6"


class TacticalRoomLayout(BaseModel):
    room_id: int
    room_name: str
    room_role: str
    grid_width: int
    grid_height: int
    enemies: list[TacticalEnemy] = Field(default_factory=list)
    obstacles: list[TacticalObstacle] = Field(default_factory=list)
    doors: list[TacticalDoor] = Field(default_factory=list)
    traps: list[TacticalTrap] = Field(default_factory=list)
    is_boss_room: bool = False
    difficulty_score: float = 0.0
    description: str = ""


# ── Room Dimension Rules ─────────────────────────────────────────────────

ROLE_DIMENSIONS: dict[str, tuple[int, int, int, int]] = {
    # (min_w, max_w, min_h, max_h) in grid squares (5ft each)
    "entrance": (6, 8, 6, 8),
    "corridor": (4, 5, 8, 14),
    "shrine": (6, 10, 6, 10),
    "lair": (8, 12, 8, 12),
    "vault": (5, 8, 5, 8),
    "trap_room": (6, 10, 6, 10),
    "secret_room": (5, 6, 5, 6),
    "boss_room": (10, 16, 10, 16),
    "rest_area": (5, 8, 5, 8),
    "guard_post": (5, 8, 6, 10),
    "puzzle_room": (6, 10, 6, 10),
    "armory": (5, 8, 5, 8),
}

# Obstacle types by room role
ROLE_OBSTACLES: dict[str, list[tuple[str, float]]] = {
    "entrance": [("pillar", 0.3), ("crate", 0.2)],
    "corridor": [("pillar", 0.4), ("rubble", 0.3)],
    "shrine": [("altar", 1.0), ("pillar", 0.6), ("statue", 0.3)],
    "lair": [("rubble", 0.4), ("pool", 0.3), ("web", 0.2)],
    "vault": [("crate", 0.5), ("pillar", 0.3)],
    "trap_room": [("pillar", 0.3), ("pit", 0.4), ("rubble", 0.2)],
    "secret_room": [("crate", 0.4), ("statue", 0.3)],
    "boss_room": [("pillar", 0.7), ("statue", 0.4), ("fire", 0.3), ("pool", 0.2)],
    "rest_area": [("table", 0.5), ("crate", 0.3)],
    "guard_post": [("table", 0.4), ("crate", 0.3), ("pillar", 0.2)],
    "puzzle_room": [("pillar", 0.5), ("statue", 0.4), ("altar", 0.3)],
    "armory": [("crate", 0.6), ("table", 0.3)],
}

# Size map for monster sizes to grid squares
SIZE_GRID: dict[str, int] = {
    "Tiny": 1,
    "Small": 1,
    "Medium": 1,
    "Large": 2,
    "Huge": 3,
    "Gargantuan": 4,
}


def _get_room_dimensions(role: str, rng: random.Random) -> tuple[int, int]:
    dims = ROLE_DIMENSIONS.get(role, (6, 10, 6, 10))
    w = rng.randint(dims[0], dims[1])
    h = rng.randint(dims[2], dims[3])
    return w, h


def _place_obstacles(
    role: str,
    width: int,
    height: int,
    occupied: set[tuple[int, int]],
    rng: random.Random,
) -> list[TacticalObstacle]:
    obstacles: list[TacticalObstacle] = []
    role_obs = ROLE_OBSTACLES.get(role, [("pillar", 0.3)])

    for kind, chance in role_obs:
        if rng.random() > chance:
            continue

        # Determine count based on room size
        area = width * height
        max_obs = max(1, area // 20)
        count = rng.randint(1, max_obs)

        for _ in range(count):
            # Place obstacles away from edges (1 tile margin)
            if width <= 4 or height <= 4:
                continue
            for _attempt in range(20):
                ox = rng.randint(2, max(2, width - 3))
                oy = rng.randint(2, max(2, height - 3))
                if (ox, oy) not in occupied:
                    occupied.add((ox, oy))
                    blocks = kind not in ("pool", "fire", "web")
                    provides_cover = kind in ("pillar", "crate", "table", "statue", "altar", "rubble")
                    obstacles.append(TacticalObstacle(
                        kind=kind, x=ox, y=oy,
                        blocks_movement=blocks,
                        provides_cover=provides_cover,
                    ))
                    break

    return obstacles


def _place_doors(
    room: Room,
    width: int,
    height: int,
    edges: list[DungeonEdge],
    rng: random.Random,
) -> list[TacticalDoor]:
    doors: list[TacticalDoor] = []
    walls = ["north", "south", "east", "west"]

    # Find connections to this room
    connections = [e for e in edges if e.from_room == room.room_id or e.to_room == room.room_id]

    for i, edge in enumerate(connections):
        wall = walls[i % len(walls)]
        if wall == "north":
            dx, dy = width // 2, 0
        elif wall == "south":
            dx, dy = width // 2, height - 1
        elif wall == "west":
            dx, dy = 0, height // 2
        else:
            dx, dy = width - 1, height // 2

        doors.append(TacticalDoor(
            x=dx, y=dy, wall=wall,
            is_locked=edge.is_locked,
            is_hidden=edge.is_hidden,
        ))

    return doors


def _place_enemies(
    room: Room,
    width: int,
    height: int,
    occupied: set[tuple[int, int]],
    rng: random.Random,
) -> list[TacticalEnemy]:
    enemies: list[TacticalEnemy] = []
    if not room.encounter:
        return enemies

    for em in room.encounter.monsters:
        monster = em.monster
        grid_size = SIZE_GRID.get(monster.size, 1)
        positions: list[GridPosition] = []

        for _ in range(em.count):
            placed = False
            for _attempt in range(50):
                # Boss monsters go toward the back of the room
                max_x = max(1, width - 2 - grid_size)
                max_y = max(1, height - 2 - grid_size)
                if room.is_boss_room and monster.boss_suitable:
                    ex = rng.randint(max(1, width // 4), min(max_x, 3 * width // 4))
                    ey = rng.randint(max(1, height // 2), max_y)
                else:
                    ex = rng.randint(1, max_x)
                    ey = rng.randint(1, max_y)

                # Check all cells for this unit
                cells_clear = True
                for gx in range(grid_size):
                    for gy in range(grid_size):
                        if (ex + gx, ey + gy) in occupied:
                            cells_clear = False
                            break
                    if not cells_clear:
                        break

                if cells_clear:
                    for gx in range(grid_size):
                        for gy in range(grid_size):
                            occupied.add((ex + gx, ey + gy))
                    positions.append(GridPosition(x=ex, y=ey))
                    placed = True
                    break

            if not placed:
                # Fallback: place anywhere not occupied
                positions.append(GridPosition(x=width // 2, y=height // 2))

        enemies.append(TacticalEnemy(
            name=monster.name,
            count=em.count,
            combat_role=monster.combat_role,
            challenge_rating=monster.challenge_rating,
            hit_points=monster.hit_points,
            armor_class=monster.armor_class,
            size=monster.size,
            is_boss=room.is_boss_room and monster.boss_suitable,
            positions=positions,
        ))

    return enemies


def _place_traps(
    room: Room,
    width: int,
    height: int,
    occupied: set[tuple[int, int]],
    rng: random.Random,
) -> list[TacticalTrap]:
    traps: list[TacticalTrap] = []
    if not room.trap:
        return traps

    if width <= 4 or height <= 4:
        return traps

    for _attempt in range(30):
        tx = rng.randint(2, max(2, width - 3))
        ty = rng.randint(2, max(2, height - 3))
        if (tx, ty) not in occupied:
            occupied.add((tx, ty))
            traps.append(TacticalTrap(
                name=room.trap.name,
                x=tx, y=ty,
                save_dc=room.trap.save_dc,
                damage_dice=room.trap.damage_dice,
            ))
            break

    return traps


def generate_tactical_layout(
    room: Room,
    edges: list[DungeonEdge],
    seed: int,
) -> TacticalRoomLayout:
    """Generate a tactical grid layout for a single room."""
    rng = random.Random(seed + room.room_id * 1000)

    role = room.role if isinstance(room.role, str) else room.role.value
    width, height = _get_room_dimensions(role, rng)

    occupied: set[tuple[int, int]] = set()

    # Mark edges as occupied
    for x in range(width):
        occupied.add((x, 0))
        occupied.add((x, height - 1))
    for y in range(height):
        occupied.add((0, y))
        occupied.add((width - 1, y))

    obstacles = _place_obstacles(role, width, height, occupied, rng)
    doors = _place_doors(room, width, height, edges, rng)
    traps = _place_traps(room, width, height, occupied, rng)
    enemies = _place_enemies(room, width, height, occupied, rng)

    return TacticalRoomLayout(
        room_id=room.room_id,
        room_name=room.name,
        room_role=role,
        grid_width=width,
        grid_height=height,
        enemies=enemies,
        obstacles=obstacles,
        doors=doors,
        traps=traps,
        is_boss_room=room.is_boss_room,
        difficulty_score=room.difficulty_score,
        description=room.description,
    )


def generate_all_tactical_layouts(dungeon: Dungeon) -> list[TacticalRoomLayout]:
    """Generate tactical layouts for all rooms with encounters in a dungeon."""
    layouts: list[TacticalRoomLayout] = []
    for room in dungeon.rooms:
        layout = generate_tactical_layout(room, dungeon.edges, dungeon.seed)
        layouts.append(layout)
    return layouts
