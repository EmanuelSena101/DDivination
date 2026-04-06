"""Dungeon analysis layer.

Analyzes a generated dungeon for difficulty, pathing, dead ends,
risk/reward balance, and pacing.
"""
from __future__ import annotations

import networkx as nx

from app.models import Dungeon, DungeonAnalysis, Room


def analyze_dungeon(dungeon: Dungeon) -> DungeonAnalysis:
    """Perform complete analysis on a generated dungeon."""
    rooms = dungeon.rooms
    edges = dungeon.edges

    if not rooms:
        return DungeonAnalysis()

    # Build networkx graph
    G = nx.Graph()
    for room in rooms:
        G.add_node(room.room_id, role=room.role.value, difficulty=room.difficulty_score)

    for edge in edges:
        G.add_edge(edge.from_room, edge.to_room, description=edge.description)

    # Basic counts
    total_rooms = len(rooms)
    total_encounters = sum(1 for r in rooms if r.encounter)
    total_traps = sum(1 for r in rooms if r.trap)
    total_xp = sum(
        r.encounter.total_xp for r in rooms if r.encounter
    )

    # Find boss room
    boss_room = next((r for r in rooms if r.is_boss_room), None)
    boss_room_id = boss_room.room_id if boss_room else None

    # Difficulty by room
    difficulty_by_room = {r.room_id: r.difficulty_score for r in rooms}
    difficulty_progression = [r.difficulty_score for r in rooms]

    # Average and max difficulty
    scores = [r.difficulty_score for r in rooms if r.difficulty_score > 0]
    avg_difficulty = sum(scores) / len(scores) if scores else 0.0
    max_difficulty = max(scores) if scores else 0.0

    # Path analysis
    critical_path: list[int] = []
    critical_path_length = 0
    dead_ends: list[int] = []
    branching_factor = 0.0

    if G.number_of_nodes() > 0:
        # Critical path: shortest path from entrance (0) to boss room
        entrance_id = 0
        target_id = boss_room_id if boss_room_id is not None else len(rooms) - 1

        try:
            if nx.has_path(G, entrance_id, target_id):
                critical_path = list(nx.shortest_path(G, entrance_id, target_id))
                critical_path_length = len(critical_path)
        except (nx.NetworkXError, nx.NodeNotFound):
            critical_path = list(range(total_rooms))
            critical_path_length = total_rooms

        # Dead ends: nodes with degree 1 (except entrance if it's the only connection)
        dead_ends = [n for n in G.nodes() if G.degree(n) == 1 and n != entrance_id]

        # Branching factor: average degree
        degrees = [G.degree(n) for n in G.nodes()]
        branching_factor = sum(degrees) / len(degrees) if degrees else 0.0

    # Treasure analysis
    total_gold = sum(
        r.treasure.gold for r in rooms if r.treasure
    )
    total_magic_items = sum(
        len(r.treasure.items) for r in rooms if r.treasure
    )

    # Risk/reward by room
    risk_reward_by_room: dict[int, dict[str, float]] = {}
    for room in rooms:
        risk = room.difficulty_score
        reward = 0.0
        if room.treasure:
            reward += room.treasure.gold / 100.0
            reward += len(room.treasure.items) * 5.0
            reward += len(room.treasure.equipment) * 1.0
        risk_reward_by_room[room.room_id] = {"risk": risk, "reward": reward}

    # Risk/reward balance
    total_risk = sum(v["risk"] for v in risk_reward_by_room.values())
    total_reward = sum(v["reward"] for v in risk_reward_by_room.values())
    if total_risk > 0 and total_reward > 0:
        ratio = total_reward / total_risk
        if ratio < 0.5:
            risk_reward_balance = "risk-heavy (low rewards for the danger)"
        elif ratio > 2.0:
            risk_reward_balance = "reward-heavy (generous loot for the difficulty)"
        else:
            risk_reward_balance = "balanced"
    else:
        risk_reward_balance = "balanced"

    # Estimated overall difficulty
    estimated_difficulty = _estimate_overall_difficulty(rooms, dungeon.config.party_level, dungeon.config.party_size)

    # Pacing notes
    pacing_notes = _generate_pacing_notes(rooms, critical_path, dead_ends, branching_factor)

    return DungeonAnalysis(
        total_rooms=total_rooms,
        total_encounters=total_encounters,
        total_traps=total_traps,
        total_xp=total_xp,
        estimated_difficulty=estimated_difficulty,
        difficulty_by_room=difficulty_by_room,
        difficulty_progression=difficulty_progression,
        critical_path=critical_path,
        critical_path_length=critical_path_length,
        dead_ends=dead_ends,
        branching_factor=round(branching_factor, 2),
        total_gold=total_gold,
        total_magic_items=total_magic_items,
        risk_reward_by_room=risk_reward_by_room,
        risk_reward_balance=risk_reward_balance,
        has_boss=boss_room is not None and boss_room.encounter is not None,
        boss_room_id=boss_room_id,
        avg_room_difficulty=round(avg_difficulty, 2),
        max_room_difficulty=round(max_difficulty, 2),
        pacing_notes=pacing_notes,
    )


def _estimate_overall_difficulty(rooms: list[Room], party_level: int, party_size: int) -> str:
    """Estimate the overall dungeon difficulty."""
    encounter_difficulties = []
    for room in rooms:
        if room.encounter:
            encounter_difficulties.append(room.encounter.difficulty_rating)

    if not encounter_difficulties:
        return "easy"

    diff_scores = {"easy": 1, "medium": 2, "hard": 3, "deadly": 4}
    avg_score = sum(diff_scores.get(d, 2) for d in encounter_difficulties) / len(encounter_difficulties)

    if avg_score >= 3.5:
        return "deadly"
    elif avg_score >= 2.5:
        return "hard"
    elif avg_score >= 1.5:
        return "medium"
    else:
        return "easy"


def _generate_pacing_notes(
    rooms: list[Room],
    critical_path: list[int],
    dead_ends: list[int],
    branching_factor: float,
) -> list[str]:
    """Generate GM-facing pacing notes about the dungeon."""
    notes: list[str] = []

    # Encounter density
    encounter_rooms = sum(1 for r in rooms if r.encounter)
    density = encounter_rooms / len(rooms) if rooms else 0
    if density > 0.7:
        notes.append("High combat density: Consider allowing short rests between encounters.")
    elif density < 0.3:
        notes.append("Low combat density: The dungeon emphasizes exploration over combat.")
    else:
        notes.append("Balanced combat pacing with exploration breaks between encounters.")

    # Rest opportunities
    rest_rooms = sum(1 for r in rooms if r.role.value == "rest_area")
    if rest_rooms == 0:
        notes.append("No designated rest areas. The party may need to find creative solutions for resting.")
    elif rest_rooms == 1:
        notes.append("One rest area available. Timing the rest will be a strategic decision.")

    # Dead end value
    if dead_ends:
        dead_end_rooms = [r for r in rooms if r.room_id in dead_ends]
        has_treasure = any(r.treasure for r in dead_end_rooms)
        if has_treasure:
            notes.append(f"{len(dead_ends)} dead end(s) with hidden rewards encourage exploration.")
        else:
            notes.append(f"{len(dead_ends)} dead end(s) may slow the party down without payoff.")

    # Branching
    if branching_factor > 2.5:
        notes.append("Highly branching layout: The party will face meaningful choices about which path to take.")
    elif branching_factor < 1.5:
        notes.append("Mostly linear layout: Straightforward progression toward the boss.")

    # Difficulty escalation
    if len(rooms) >= 3:
        first_third = rooms[:len(rooms) // 3]
        last_third = rooms[-(len(rooms) // 3):]
        avg_early = sum(r.difficulty_score for r in first_third) / len(first_third) if first_third else 0
        avg_late = sum(r.difficulty_score for r in last_third) / len(last_third) if last_third else 0
        if avg_late > avg_early * 1.5:
            notes.append("Good difficulty escalation: The dungeon gets progressively harder.")
        elif avg_late < avg_early:
            notes.append("Note: Later rooms are easier than early ones. Consider rerolling encounters.")

    # Critical path
    if critical_path:
        path_rooms = [r for r in rooms if r.room_id in critical_path]
        path_encounters = sum(1 for r in path_rooms if r.encounter)
        notes.append(f"Critical path: {len(critical_path)} rooms, {path_encounters} mandatory encounters.")

    return notes
