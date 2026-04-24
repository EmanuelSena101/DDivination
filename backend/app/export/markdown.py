"""Render a Dungeon to a single Markdown document (GM-facing)."""
from __future__ import annotations

from io import StringIO

from app.models import Dungeon, Room

ROLE_SYMBOLS: dict[str, str] = {
    "entrance": "E",
    "corridor": "C",
    "shrine": "S",
    "lair": "L",
    "vault": "V",
    "trap_room": "T",
    "secret_room": "?",
    "boss_room": "B",
    "rest_area": "R",
    "guard_post": "G",
    "puzzle_room": "P",
    "armory": "A",
}


def _title(s: str) -> str:
    return s.replace("_", " ").title()


def _room_heading(room: Room, index: int) -> str:
    tags: list[str] = []
    if room.is_boss_room:
        tags.append("**BOSS**")
    if room.is_secret:
        tags.append("**SECRET**")
    tag_str = f" · {' · '.join(tags)}" if tags else ""
    return f"### Room {index + 1}: {room.name}{tag_str}"


def _render_encounter(buf: StringIO, room: Room) -> None:
    enc = room.encounter
    if not enc:
        return
    buf.write(f"**Encounter** — _{enc.difficulty_rating.upper()}_  \n")
    if enc.description:
        buf.write(f"_{enc.description}_\n\n")
    for em in enc.monsters:
        m = em.monster
        buf.write(
            f"- {em.count}× **{m.name}** — CR {m.challenge_rating}, "
            f"{m.combat_role.value}, HP {m.hit_points}, AC {m.armor_class}, "
            f"{m.xp * em.count} XP\n"
        )
    buf.write(f"\n_Total XP: {enc.total_xp} · Adjusted: {enc.adjusted_xp}_\n\n")


def _render_trap(buf: StringIO, room: Room) -> None:
    t = room.trap
    if not t:
        return
    buf.write(f"**Trap — {t.name}** _(danger: {t.danger_level.value})_  \n")
    buf.write(f"{t.description}  \n")
    buf.write(f"Damage `{t.damage_dice}` · Save DC {t.save_dc} {t.save_type}\n\n")


def _render_treasure(buf: StringIO, room: Room) -> None:
    tr = room.treasure
    if not tr:
        return
    if tr.gold == 0 and not tr.items and not tr.equipment:
        return
    buf.write("**Treasure**\n")
    if tr.gold:
        buf.write(f"- {tr.gold} gp\n")
    for item in tr.items:
        buf.write(f"- {item.name} _({item.rarity})_\n")
    for eq in tr.equipment:
        buf.write(f"- {eq.name}\n")
    buf.write("\n")


def render_markdown(dungeon: Dungeon) -> str:
    """Render a complete dungeon to a GM-facing Markdown string."""
    buf = StringIO()
    cfg = dungeon.config

    # Header
    buf.write(f"# {dungeon.name}\n\n")
    if dungeon.summary:
        buf.write(f"> {dungeon.summary}\n\n")

    # Meta
    buf.write("## At a Glance\n\n")
    buf.write(f"- **Seed:** `{dungeon.seed}`\n")
    buf.write(f"- **Party:** {cfg.party_size} players, level {cfg.party_level}\n")
    buf.write(f"- **Theme:** {_title(cfg.theme.value)} · **Biome:** {_title(cfg.biome.value)}\n")
    buf.write(
        f"- **Size:** {_title(cfg.dungeon_size.value)} · "
        f"**Layout:** {_title(cfg.structure_style.value)} · "
        f"**Difficulty:** {_title(cfg.difficulty.value)}\n"
    )
    buf.write(
        f"- **Trap Density:** {_title(cfg.trap_density.value)} · "
        f"**Treasure:** {_title(cfg.treasure_quality.value)}\n"
    )
    if cfg.boss_type:
        buf.write(f"- **Boss Type:** {cfg.boss_type}\n")
    buf.write("\n")

    # Narrative
    if dungeon.narrative_intro or dungeon.narrative_hook:
        buf.write("## Narrative\n\n")
        if dungeon.narrative_hook:
            buf.write(f"**Hook.** {dungeon.narrative_hook}\n\n")
        if dungeon.narrative_intro:
            buf.write(f"**Setting.** {dungeon.narrative_intro}\n\n")

    # Map (graph as text list)
    buf.write("## Map\n\n")
    buf.write("| # | Room | Role | Symbol |\n|---|------|------|:------:|\n")
    for r in dungeon.rooms:
        sym = ROLE_SYMBOLS.get(r.role.value, "X")
        buf.write(f"| {r.room_id} | {r.name} | {_title(r.role.value)} | {sym} |\n")
    buf.write("\n")

    if dungeon.edges:
        buf.write("### Connections\n\n")
        for e in dungeon.edges:
            flags = []
            if e.is_locked:
                flags.append("locked")
            if e.is_hidden:
                flags.append("hidden")
            suffix = f" _({', '.join(flags)})_" if flags else ""
            buf.write(f"- Room {e.from_room} ↔ Room {e.to_room} — {e.description}{suffix}\n")
        buf.write("\n")

    # Rooms
    buf.write("## Rooms\n\n")
    for i, room in enumerate(dungeon.rooms):
        buf.write(f"{_room_heading(room, i)}\n\n")
        buf.write(f"_{_title(room.role.value)} · Difficulty {room.difficulty_score:.1f}/10_\n\n")
        if room.description:
            buf.write(f"{room.description}\n\n")
        if room.flavor_text:
            buf.write(f"> {room.flavor_text}\n\n")
        _render_encounter(buf, room)
        _render_trap(buf, room)
        _render_treasure(buf, room)
        buf.write("---\n\n")

    # Analysis
    a = dungeon.analysis
    if a:
        buf.write("## Analysis\n\n")
        buf.write(f"- **Overall difficulty:** {a.estimated_difficulty.upper()}\n")
        buf.write(f"- **Rooms:** {a.total_rooms} · **Encounters:** {a.total_encounters} · **Traps:** {a.total_traps}\n")
        buf.write(f"- **Total XP:** {a.total_xp:,} · **Total Gold:** {a.total_gold:,} · **Magic Items:** {a.total_magic_items}\n")
        buf.write(f"- **Avg room difficulty:** {a.avg_room_difficulty} · **Max:** {a.max_room_difficulty}\n")
        buf.write(f"- **Branching factor:** {a.branching_factor}\n")
        buf.write(f"- **Dead ends:** {len(a.dead_ends)} · **Critical path:** {a.critical_path_length} rooms\n")
        buf.write(f"- **Risk/reward:** {a.risk_reward_balance}\n\n")

        if a.critical_path:
            path = " → ".join(f"R{rid}" for rid in a.critical_path)
            buf.write(f"**Critical path:** {path}\n\n")

        if a.pacing_notes:
            buf.write("### GM Pacing Notes\n\n")
            for note in a.pacing_notes:
                buf.write(f"- {note}\n")
            buf.write("\n")

    buf.write("_Generated by DDivination — data from the D&D 5e SRD._\n")
    return buf.getvalue()
