"""Local trap definitions organized by theme and difficulty."""
from __future__ import annotations

from app.models import Difficulty, Trap

# ── Trap templates organized by theme ─────────────────────────────────────

TRAP_TEMPLATES: list[Trap] = [
    # Temple traps
    Trap(name="Sacred Flame Glyph", description="A divine glyph etched into the floor erupts with sacred flame when stepped upon.", damage_dice="2d8", save_dc=13, save_type="DEX", danger_level=Difficulty.EASY, theme_tags=["temple"], biome_tags=[]),
    Trap(name="Collapsing Altar", description="The ancient altar crumbles, sending heavy stone blocks tumbling down.", damage_dice="3d10", save_dc=15, save_type="DEX", danger_level=Difficulty.HARD, theme_tags=["temple", "ruins"], biome_tags=[]),
    Trap(name="Holy Water Deluge", description="Blessed water pours from the ceiling, burning undead and fiendish creatures.", damage_dice="2d6", save_dc=12, save_type="DEX", danger_level=Difficulty.EASY, theme_tags=["temple"], biome_tags=["aquatic"]),

    # Crypt traps
    Trap(name="Necrotic Rune", description="A dark rune pulses with deathly energy, draining life force from the living.", damage_dice="3d8", save_dc=14, save_type="CON", danger_level=Difficulty.MEDIUM, theme_tags=["crypt"], biome_tags=["shadow"]),
    Trap(name="Coffin Spike Trap", description="A coffin lid springs open, launching rusted iron spikes.", damage_dice="2d10", save_dc=13, save_type="DEX", danger_level=Difficulty.MEDIUM, theme_tags=["crypt"], biome_tags=[]),
    Trap(name="Bone Shard Explosion", description="Piled bones explode outward in a spray of sharp fragments.", damage_dice="4d6", save_dc=14, save_type="DEX", danger_level=Difficulty.MEDIUM, theme_tags=["crypt"], biome_tags=[]),

    # Cave traps
    Trap(name="Falling Stalactites", description="Vibrations cause stalactites to break free and crash down.", damage_dice="3d6", save_dc=12, save_type="DEX", danger_level=Difficulty.EASY, theme_tags=["cave", "mine"], biome_tags=["underground"]),
    Trap(name="Sinkhole", description="The floor gives way to a deep sinkhole concealed by loose earth.", damage_dice="2d10", save_dc=14, save_type="DEX", danger_level=Difficulty.MEDIUM, theme_tags=["cave"], biome_tags=["underground", "swamp"]),
    Trap(name="Toxic Spore Cloud", description="Disturbed fungi release a cloud of toxic spores.", damage_dice="2d8", save_dc=13, save_type="CON", danger_level=Difficulty.MEDIUM, theme_tags=["cave"], biome_tags=["underground", "forest", "swamp"]),

    # Fortress traps
    Trap(name="Crossbow Turret", description="A hidden crossbow mechanism fires bolts at intruders.", damage_dice="2d10", save_dc=14, save_type="DEX", danger_level=Difficulty.MEDIUM, theme_tags=["fortress"], biome_tags=[]),
    Trap(name="Portcullis Slam", description="A heavy portcullis crashes down, trapping and crushing those beneath.", damage_dice="3d10", save_dc=15, save_type="DEX", danger_level=Difficulty.HARD, theme_tags=["fortress"], biome_tags=[]),
    Trap(name="Oil and Fire", description="Oil pours from hidden vents and ignites, engulfing the corridor in flame.", damage_dice="4d6", save_dc=14, save_type="DEX", danger_level=Difficulty.HARD, theme_tags=["fortress"], biome_tags=["volcanic"]),

    # Ruins traps
    Trap(name="Crumbling Floor", description="Ancient stonework gives way underfoot, dropping victims into the level below.", damage_dice="2d6", save_dc=12, save_type="DEX", danger_level=Difficulty.EASY, theme_tags=["ruins", "temple"], biome_tags=[]),
    Trap(name="Guardian Glyph", description="An arcane ward left by ancient builders unleashes a blast of force.", damage_dice="3d8", save_dc=15, save_type="WIS", danger_level=Difficulty.HARD, theme_tags=["ruins"], biome_tags=[]),

    # Cult traps
    Trap(name="Blood Circle", description="Stepping into a ritual circle causes it to activate, draining blood.", damage_dice="3d6", save_dc=14, save_type="CON", danger_level=Difficulty.MEDIUM, theme_tags=["cult"], biome_tags=["shadow"]),
    Trap(name="Demonic Binding", description="A summoning circle tries to bind intruders, holding them in place.", damage_dice="1d8", save_dc=15, save_type="CHA", danger_level=Difficulty.MEDIUM, theme_tags=["cult"], biome_tags=[]),

    # Sewer traps
    Trap(name="Acid Pool", description="A pool of dissolved waste that is actually a potent acid.", damage_dice="3d6", save_dc=13, save_type="DEX", danger_level=Difficulty.MEDIUM, theme_tags=["sewer"], biome_tags=["swamp", "aquatic"]),
    Trap(name="Gas Pocket", description="A pocket of flammable sewer gas ignites when exposed to flame.", damage_dice="4d6", save_dc=14, save_type="DEX", danger_level=Difficulty.HARD, theme_tags=["sewer"], biome_tags=[]),

    # Mine traps
    Trap(name="Cave-In", description="Weakened supports give way, causing a section of tunnel to collapse.", damage_dice="4d8", save_dc=15, save_type="DEX", danger_level=Difficulty.HARD, theme_tags=["mine", "cave"], biome_tags=["underground"]),
    Trap(name="Mine Cart Rush", description="A loaded mine cart comes barreling down the tracks.", damage_dice="3d10", save_dc=14, save_type="DEX", danger_level=Difficulty.HARD, theme_tags=["mine"], biome_tags=[]),

    # Generic / multi-theme
    Trap(name="Pit Trap", description="A concealed pit opens beneath the unwary.", damage_dice="2d6", save_dc=12, save_type="DEX", danger_level=Difficulty.EASY, theme_tags=[], biome_tags=[]),
    Trap(name="Poison Dart Wall", description="Tiny holes in the wall fire poisoned darts triggered by a pressure plate.", damage_dice="1d4", save_dc=13, save_type="CON", danger_level=Difficulty.EASY, theme_tags=[], biome_tags=[]),
    Trap(name="Swinging Blade", description="A massive blade swings across the corridor on a pendulum.", damage_dice="3d8", save_dc=14, save_type="DEX", danger_level=Difficulty.MEDIUM, theme_tags=[], biome_tags=[]),
    Trap(name="Crushing Walls", description="The walls begin to close in, threatening to crush everything between them.", damage_dice="4d10", save_dc=16, save_type="STR", danger_level=Difficulty.DEADLY, theme_tags=[], biome_tags=[]),
]


def get_traps_for_theme(theme: str, biome: str) -> list[Trap]:
    """Get traps that match the given theme and/or biome."""
    result = []
    for trap in TRAP_TEMPLATES:
        score = 0
        if theme in trap.theme_tags:
            score += 2
        if biome in trap.biome_tags:
            score += 1
        if not trap.theme_tags and not trap.biome_tags:
            score += 0.5  # Generic traps are always somewhat relevant
        if score > 0:
            result.append((score, trap))
    result.sort(key=lambda x: x[0], reverse=True)
    return [t for _, t in result]


def get_trap_for_difficulty(traps: list[Trap], difficulty: str) -> list[Trap]:
    """Filter traps by difficulty level."""
    diff_map = {"easy": [Difficulty.EASY], "medium": [Difficulty.EASY, Difficulty.MEDIUM],
                "hard": [Difficulty.MEDIUM, Difficulty.HARD], "deadly": [Difficulty.HARD, Difficulty.DEADLY]}
    allowed = diff_map.get(difficulty, [Difficulty.MEDIUM])
    return [t for t in traps if t.danger_level in allowed]
