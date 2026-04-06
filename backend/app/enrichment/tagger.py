"""Enrichment layer: tags, combat roles, boss suitability, reward weighting.

This module adds semantic meaning on top of raw D&D 5e API data so the
procedural generator can make thematic, balanced decisions.
"""
from __future__ import annotations


# ── Theme tag rules (keyword -> tag) ──────────────────────────────────────

_THEME_KEYWORDS: dict[str, list[str]] = {
    "temple": ["celestial", "fiend", "angel", "devil", "demon", "cleric", "priest",
               "holy", "divine", "altar", "radiant", "necrotic"],
    "crypt": ["undead", "skeleton", "zombie", "ghoul", "ghost", "wight", "wraith",
              "lich", "mummy", "vampire", "bone", "tomb", "grave", "death"],
    "cave": ["beast", "ooze", "spider", "bat", "bear", "wolf", "worm",
             "cave", "rock", "stone", "earth", "burrow"],
    "ruins": ["construct", "golem", "elemental", "ancient", "ruin", "guardian",
              "gargoyle", "animated", "statue"],
    "fortress": ["humanoid", "soldier", "guard", "knight", "orc", "goblin",
                 "hobgoblin", "bugbear", "bandit", "captain", "war"],
    "cult": ["fiend", "cultist", "warlock", "demon", "devil", "acolyte",
             "fanatic", "ritual", "sacrifice", "dark", "shadow"],
    "sewer": ["ooze", "rat", "swarm", "slime", "gelatinous", "otyugh",
              "crocodile", "aboleth", "filth"],
    "mine": ["dwarf", "earth", "elemental", "umber", "hulk", "xorn",
             "galeb", "duhr", "crystal", "gem", "ore"],
}

_BIOME_KEYWORDS: dict[str, list[str]] = {
    "swamp": ["swamp", "marsh", "bog", "lizard", "toad", "frog", "hag",
              "shambling", "mound", "vine", "plant", "fungus"],
    "aquatic": ["aquatic", "water", "sea", "fish", "shark", "merfolk", "sahuagin",
                "aboleth", "kraken", "swim", "amphibious", "crab"],
    "forest": ["forest", "wood", "tree", "elf", "fey", "dryad", "treant",
               "owl", "wolf", "boar", "deer", "sprite", "pixie"],
    "volcanic": ["fire", "flame", "magma", "lava", "salamander", "efreeti",
                 "red dragon", "hell", "infernal", "burn", "ash"],
    "shadow": ["shadow", "dark", "underdark", "drow", "wraith", "specter",
               "shadow", "nightmare", "gloom", "void"],
    "underground": ["underground", "underdark", "drow", "duergar", "myconid",
                    "cave", "tunnel", "deep", "dark", "mind flayer", "illithid"],
    "arctic": ["cold", "ice", "frost", "winter", "yeti", "remorhaz",
               "white dragon", "polar", "snow", "frozen"],
    "desert": ["desert", "sand", "scorpion", "mummy", "dust", "djinn",
               "blue dragon", "brass dragon", "dry", "arid"],
}

_COMBAT_ROLE_RULES = {
    # (min_cr, max_cr, size_hint, type_hint) -> role
    "boss": {"min_cr": 5.0, "large_or_bigger": True, "legendary": True},
}


def _match_tags(text: str, keyword_map: dict[str, list[str]]) -> list[str]:
    """Match text against keyword map and return matching tags."""
    text_lower = text.lower()
    tags = []
    for tag, keywords in keyword_map.items():
        for kw in keywords:
            if kw in text_lower:
                tags.append(tag)
                break
    return tags


def _determine_combat_role(monster: dict) -> str:
    """Determine the combat role of a monster based on its stats."""
    cr = monster.get("challenge_rating", 0)
    hp = monster.get("hit_points", 1)
    ac = monster.get("armor_class", 10)
    size = monster.get("size", "Medium").lower()
    name = monster.get("name", "").lower()
    mtype = monster.get("monster_type", "").lower()

    # Boss: high CR, large+, or named creatures
    if cr >= 10 or (cr >= 5 and size in ("large", "huge", "gargantuan")):
        return "boss"

    # Brute: high HP, medium+ size
    if hp > 50 and size in ("large", "huge", "gargantuan"):
        return "brute"

    # Controller: spellcasters and special types
    controller_hints = ["mage", "wizard", "sorcerer", "beholder", "mind flayer",
                        "aboleth", "naga", "hag", "lich"]
    if any(h in name or h in mtype for h in controller_hints):
        return "controller"

    # Sniper: ranged attackers
    sniper_hints = ["archer", "scout", "assassin", "sniper"]
    if any(h in name for h in sniper_hints):
        return "sniper"

    # Skirmisher: medium CR, mobile
    if cr >= 2 and size in ("medium", "small"):
        return "skirmisher"

    # Minion: low CR fodder
    return "minion"


def _is_boss_suitable(monster: dict) -> bool:
    """Determine if a monster is suitable as a boss encounter."""
    cr = monster.get("challenge_rating", 0)
    name = monster.get("name", "").lower()
    boss_hints = ["dragon", "lich", "beholder", "vampire", "demon lord",
                  "devil", "archmage", "death knight", "mind flayer",
                  "aboleth", "kraken", "hydra", "giant", "golem"]
    if cr >= 5:
        return True
    if any(h in name for h in boss_hints):
        return True
    return False


def enrich_monster(monster: dict) -> dict:
    """Add enrichment fields to a normalized monster dict."""
    searchable = f"{monster.get('name', '')} {monster.get('monster_type', '')} {monster.get('alignment', '')}"

    monster["theme_tags"] = _match_tags(searchable, _THEME_KEYWORDS)
    monster["biome_tags"] = _match_tags(searchable, _BIOME_KEYWORDS)
    monster["combat_role"] = _determine_combat_role(monster)
    monster["boss_suitable"] = _is_boss_suitable(monster)
    monster["source"] = "dnd5eapi"

    return monster


def enrich_magic_item(item: dict) -> dict:
    """Add enrichment fields to a normalized magic item dict."""
    searchable = f"{item.get('name', '')} {item.get('description', '')} {item.get('category', '')}"

    item["theme_tags"] = _match_tags(searchable, _THEME_KEYWORDS)

    # Reward weight based on rarity
    rarity = item.get("rarity", "common").lower()
    weight_map = {
        "common": 1.0,
        "uncommon": 2.0,
        "rare": 4.0,
        "very rare": 8.0,
        "legendary": 16.0,
        "artifact": 32.0,
    }
    item["reward_weight"] = weight_map.get(rarity, 1.0)
    item["source"] = "dnd5eapi"

    return item


def enrich_equipment(equip: dict) -> dict:
    """Add enrichment fields to normalized equipment."""
    equip["source"] = "dnd5eapi"
    return equip
