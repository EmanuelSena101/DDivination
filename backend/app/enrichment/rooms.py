"""Local room archetype definitions and narrative templates."""
from __future__ import annotations

from app.models import RoomRole

# ── Room archetypes with name templates and flavor ────────────────────────

ROOM_ARCHETYPES: dict[RoomRole, dict[str, list[str]]] = {
    RoomRole.ENTRANCE: {
        "names": ["Grand Entrance", "Gateway Chamber", "Threshold Hall", "Entry Vestibule", "Foyer of Shadows"],
        "descriptions": [
            "The entrance to the dungeon yawns before you, its darkness promising both danger and glory.",
            "Worn steps lead down into the earth. The air grows cold and stale.",
            "A massive archway frames the passage ahead, carved with ancient symbols.",
            "The entrance is flanked by crumbling pillars, remnants of a grander age.",
        ],
    },
    RoomRole.CORRIDOR: {
        "names": ["Winding Passage", "Long Corridor", "Narrow Tunnel", "Dark Hallway", "Connecting Gallery"],
        "descriptions": [
            "A long corridor stretches ahead, the walls close and oppressive.",
            "This passage twists and turns, making it difficult to maintain direction.",
            "The corridor is lined with faded murals depicting forgotten battles.",
            "Water drips from the ceiling of this damp, narrow passage.",
        ],
    },
    RoomRole.SHRINE: {
        "names": ["Forgotten Shrine", "Desecrated Chapel", "Prayer Chamber", "Sacred Sanctum", "Altar Room"],
        "descriptions": [
            "An altar stands at the center of this room, still faintly radiating divine energy.",
            "Broken prayer beads and candle stubs litter the floor of this abandoned shrine.",
            "The walls are covered in religious iconography, some defaced by unknown hands.",
            "A sense of lingering holiness pervades this chamber despite its decay.",
        ],
    },
    RoomRole.LAIR: {
        "names": ["Beast's Lair", "Nesting Chamber", "Den of Horrors", "Creature's Domain", "Hunter's Den"],
        "descriptions": [
            "Bones and refuse are scattered across the floor. Something lives here.",
            "The stench of a predator's lair fills your nostrils. Claw marks score the walls.",
            "This chamber has been claimed by a creature that has made it its territory.",
            "Tattered remains of previous victims hang from crude hooks along the walls.",
        ],
    },
    RoomRole.VAULT: {
        "names": ["Treasure Vault", "Ancient Treasury", "Sealed Vault", "Hoard Chamber", "Repository"],
        "descriptions": [
            "Behind a heavy door lies a chamber that once held great wealth.",
            "Dust-covered chests and shelves line this secure chamber.",
            "This vault's locks have been broken, but some treasures remain hidden.",
            "The gleam of gold and gems catches the torchlight in this fortified room.",
        ],
    },
    RoomRole.TRAP_ROOM: {
        "names": ["Gauntlet Chamber", "Testing Hall", "Trial Room", "Danger Room", "Perilous Passage"],
        "descriptions": [
            "Something about this room feels wrong. The floor tiles are uneven, the walls too smooth.",
            "Scorch marks and broken mechanisms hint at the dangers this room conceals.",
            "An ominous clicking sound echoes through this chamber as you enter.",
            "The remains of less fortunate adventurers serve as a warning of what lies ahead.",
        ],
    },
    RoomRole.SECRET_ROOM: {
        "names": ["Hidden Chamber", "Secret Study", "Concealed Alcove", "Shadow Cache", "Forgotten Room"],
        "descriptions": [
            "Behind a concealed door lies a chamber that has remained undisturbed for ages.",
            "This hidden room contains items its builders clearly wanted to keep safe.",
            "A secret passage opens into a small chamber, its existence known to few.",
            "Dust covers everything in this sealed room, untouched by time.",
        ],
    },
    RoomRole.BOSS_ROOM: {
        "names": ["Throne Room", "Inner Sanctum", "Final Chamber", "Grand Hall", "Heart of Darkness"],
        "descriptions": [
            "This grand chamber radiates an aura of power. The master of this domain awaits.",
            "The largest room in the dungeon, clearly the seat of whatever evil resides here.",
            "Every corridor, every trap, every guardian has been leading to this moment.",
            "Dark energy pulses through this chamber. The final confrontation is at hand.",
        ],
    },
    RoomRole.REST_AREA: {
        "names": ["Abandoned Barracks", "Safe Haven", "Quiet Chamber", "Resting Alcove", "Calm Grotto"],
        "descriptions": [
            "This room feels oddly peaceful. Perhaps a brief rest could be taken here.",
            "Old bedrolls and cold fire pits suggest this was once a resting place.",
            "A natural spring provides fresh water in this relatively safe chamber.",
            "The air is cleaner here, and the dangers of the dungeon feel momentarily distant.",
        ],
    },
    RoomRole.GUARD_POST: {
        "names": ["Guard Post", "Watchtower Chamber", "Sentinel Room", "Checkpoint", "Patrol Station"],
        "descriptions": [
            "This room serves as a defensive position, with arrow slits and barricades.",
            "Guards have been posted here to prevent intruders from advancing further.",
            "Weapon racks and watchtower-like structures fill this defensive chamber.",
            "An alarm mechanism sits ready to alert the deeper dungeon of intruders.",
        ],
    },
    RoomRole.PUZZLE_ROOM: {
        "names": ["Enigma Chamber", "Puzzle Hall", "Riddle Room", "Mechanism Chamber", "Logic Gate"],
        "descriptions": [
            "Intricate mechanisms and strange symbols cover the walls. A puzzle awaits.",
            "The door ahead is sealed. Strange rotating discs on the wall hint at how to open it.",
            "This room is a test of intellect rather than might. The solution is hidden in plain sight.",
            "Ancient gears and levers suggest a mechanical puzzle must be solved to proceed.",
        ],
    },
    RoomRole.ARMORY: {
        "names": ["Old Armory", "Weapon Store", "Equipment Cache", "Arsenal Chamber", "War Room"],
        "descriptions": [
            "Weapon racks line the walls, though many have been emptied or rusted beyond use.",
            "This was once a well-stocked armory. Some useful equipment may remain.",
            "Suits of armor stand like silent sentinels in this military supply room.",
            "The smithing forge in the corner has long grown cold, but the weapons it made endure.",
        ],
    },
}

# ── Narrative templates ───────────────────────────────────────────────────

THEME_NARRATIVES: dict[str, dict[str, list[str]]] = {
    "temple": {
        "intro": [
            "Deep within the {biome}, an ancient temple stands forgotten by all but the darkest powers.",
            "The temple of a fallen god rises from the {biome}, its sacred halls now corrupted.",
            "Once a beacon of divine light, this temple in the {biome} has been claimed by darkness.",
        ],
        "hook": [
            "Villagers report strange lights and chanting emanating from the temple at night.",
            "A holy relic has been stolen and brought to this desecrated temple.",
            "The temple's corruption is spreading, tainting the surrounding {biome}.",
        ],
    },
    "crypt": {
        "intro": [
            "Beneath the {biome}, an ancient crypt holds the restless dead and their forgotten treasures.",
            "The crypt of a powerful necromancer has been unsealed in the heart of the {biome}.",
            "Generations of the dead lie in this sprawling crypt beneath the {biome}.",
        ],
        "hook": [
            "The dead have begun to walk, emerging from the crypt to terrorize nearby settlements.",
            "A powerful artifact is buried with a long-dead king in the deepest vault.",
            "A necromancer has taken residence in the crypt, raising an army of undead.",
        ],
    },
    "cave": {
        "intro": [
            "A natural cave system in the {biome} has become home to dangerous creatures.",
            "The caverns beneath the {biome} are vast and unexplored, hiding ancient secrets.",
            "Miners stumbled upon these caves in the {biome} before vanishing without a trace.",
        ],
        "hook": [
            "Travelers have gone missing along the road near the cave entrance.",
            "Strange roaring echoes from deep within the caves at night.",
            "A rare mineral has been discovered in the caves, attracting both miners and monsters.",
        ],
    },
    "ruins": {
        "intro": [
            "The ruins of an ancient civilization lie half-buried in the {biome}.",
            "Once a great fortress of a forgotten empire, these ruins in the {biome} now crumble.",
            "Overgrown and weathered, these ruins in the {biome} still hold magical defenses.",
        ],
        "hook": [
            "Archaeologists have uncovered a sealed chamber and need protection to explore it.",
            "Ancient constructs have reactivated, attacking anyone who approaches the ruins.",
            "A map found in a dusty library points to treasure hidden within these ruins.",
        ],
    },
    "fortress": {
        "intro": [
            "A military fortress in the {biome} has been seized by hostile forces.",
            "This imposing fortress guards a strategic pass through the {biome}.",
            "Once a bastion of order, this fortress in the {biome} has fallen to chaos.",
        ],
        "hook": [
            "The fortress garrison has stopped responding to communications.",
            "Enemy forces have fortified this position and must be dislodged.",
            "Political prisoners are being held in the fortress dungeons.",
        ],
    },
    "cult": {
        "intro": [
            "A sinister cult has established a hidden base of operations in the {biome}.",
            "Dark rituals echo through a cult compound concealed within the {biome}.",
            "The Cult of the Void has chosen the {biome} as the site of their ultimate summoning.",
        ],
        "hook": [
            "Townspeople have been disappearing, taken for dark sacrificial rituals.",
            "The cult is close to completing a summoning that could unleash a great evil.",
            "A former cult member has escaped and begs for help stopping the ritual.",
        ],
    },
    "sewer": {
        "intro": [
            "Beneath the city streets, the sewers have become a labyrinth of danger.",
            "The old sewer system connects to something far older and more sinister.",
            "What lurks in the sewers has grown bold enough to drag victims below.",
        ],
        "hook": [
            "City workers have vanished in the sewers, and something is blocking the drains.",
            "A thieves' guild operates from the sewers, and their activities have turned violent.",
            "Strange creatures are emerging from the sewer grates at night.",
        ],
    },
    "mine": {
        "intro": [
            "An abandoned mine in the {biome} holds secrets deeper than anyone expected.",
            "The miners dug too deep, and what they awakened still dwells below.",
            "Rich veins of ore drew miners here, but something drove them all away.",
        ],
        "hook": [
            "The mining company wants the mine cleared so operations can resume.",
            "Earthquakes centered on the mine suggest something massive stirs within.",
            "Rare gems of immense value are said to be found in the deepest shafts.",
        ],
    },
}


BIOME_FLAVOR: dict[str, list[str]] = {
    "swamp": ["The air is thick with humidity and the smell of decay.", "Moss and vines creep along every surface.", "Murky water pools in the lower sections."],
    "aquatic": ["Water drips constantly from the ceiling.", "Bioluminescent algae provides an eerie blue glow.", "The sound of rushing water echoes through the corridors."],
    "forest": ["Roots and vines have invaded the structure.", "The scent of earth and growing things fills the air.", "Shafts of green-filtered light pierce through gaps above."],
    "volcanic": ["The air shimmers with heat.", "Cracks in the floor glow with molten rock beneath.", "The acrid smell of sulfur permeates everything."],
    "shadow": ["Shadows seem to move with a will of their own.", "The darkness here is oppressive, resisting torchlight.", "A chill runs down your spine as whispers seem to come from nowhere."],
    "underground": ["The weight of the earth above is a constant presence.", "Stalactites and stalagmites create natural columns.", "The silence is broken only by the drip of water."],
    "arctic": ["Frost covers every surface in a thin crystalline layer.", "Your breath mists in the frigid air.", "Ice formations create glittering but treacherous terrain."],
    "desert": ["Fine sand has infiltrated every crack and crevice.", "The air is dry and scorching even underground.", "Carved sandstone walls bear the marks of ancient desert winds."],
}
