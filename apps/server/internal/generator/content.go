package generator

import (
	"fmt"
	"sort"

	"github.com/EmanuelSena101/DDivination/apps/server/internal/domain"
)

const originalContentSource = "DDivination original template"

func buildAdventureContent(
	spec domain.AdventureSpec,
	rng *RNG,
	floors *[]domain.FloorMap,
	progression domain.DungeonProgression,
) ([]domain.Encounter, []domain.Treasure, []domain.Puzzle, []domain.Trap, []domain.RestPoint) {
	encounters := make([]domain.Encounter, 0, len(*floors)*2)
	treasures := make([]domain.Treasure, 0, len(*floors))
	puzzles := make([]domain.Puzzle, 0, len(*floors))
	traps := make([]domain.Trap, 0, len(*floors))
	restPoints := make([]domain.RestPoint, 0, len(*floors))

	for floorIndex := range *floors {
		floor := &(*floors)[floorIndex]
		secret := firstRoom(floor.Rooms, func(room domain.Room) bool { return room.Secret })
		boss := firstRoom(floor.Rooms, func(room domain.Room) bool { return room.ID == progression.ClimaxRoomID })
		combat := firstRoom(floor.Rooms, func(room domain.Room) bool {
			return room.Mandatory && room.ID != progression.EntryRoomID && room.ID != progression.ClimaxRoomID &&
				(room.Role == "guard" || room.Role == "lair" || room.Role == "shrine")
		})
		if combat.ID == "" {
			combat = firstRoom(floor.Rooms, func(room domain.Room) bool {
				return room.Mandatory && room.ID != progression.EntryRoomID && room.ID != progression.ClimaxRoomID && room.Role != "stairs"
			})
		}
		if combat.ID != "" {
			difficulty := spec.Difficulty
			if difficulty == "deadly" {
				difficulty = "hard"
			}
			encounter := composeEncounter(spec, difficulty, floor.ID, combat.ID, fmt.Sprintf("encounter-%d", floorIndex+1), rng)
			encounters = append(encounters, encounter)
			addEncounterMarker(floor, encounter, combat.Center)
		}
		if boss.ID != "" {
			encounter := composeEncounter(spec, "deadly", floor.ID, boss.ID, "encounter-final", rng)
			encounters = append(encounters, encounter)
		}

		if secret.ID != "" {
			treasures = append(treasures, buildTreasure(spec, floor.ID, secret.ID, floorIndex))
			trap := buildTrap(spec, floor.ID, secret.ID, floorIndex, rng)
			traps = append(traps, trap)
			floor.Entities = append(floor.Entities, domain.SceneEntity{
				ID: fmt.Sprintf("trap-%d-marker", floorIndex+1), Kind: "trap", Name: trap.Name,
				Position: secret.Center, AssetID: "prop-trap-runes", Hidden: true, RoomID: secret.ID,
			})
		}

		puzzleRoom := firstRoom(floor.Rooms, func(room domain.Room) bool { return room.Mandatory && room.Role == "puzzle" })
		if puzzleRoom.ID == "" {
			puzzleRoom = firstRoom(floor.Rooms, func(room domain.Room) bool {
				return room.Mandatory && room.ID != progression.EntryRoomID && room.ID != progression.ClimaxRoomID && room.Role != "stairs"
			})
		}
		if puzzleRoom.ID != "" {
			puzzles = append(puzzles, buildPuzzle(spec, floor.ID, puzzleRoom.ID, floorIndex, rng))
		}

		if spec.DurationHours >= 3 {
			restRoom := firstRoom(floor.Rooms, func(room domain.Room) bool { return room.Mandatory && room.Role == "rest" })
			if restRoom.ID == "" {
				restRoom = firstRoom(floor.Rooms, func(room domain.Room) bool {
					return room.Mandatory && room.ID != progression.EntryRoomID && room.ID != progression.ClimaxRoomID && room.Role != "stairs"
				})
			}
			if restRoom.ID != "" {
				restPoints = append(restPoints, domain.RestPoint{
					ID: fmt.Sprintf("rest-%d", floorIndex+1), FloorID: floor.ID, RoomID: restRoom.ID, Kind: "short",
					Name: domain.LocalizedText{PTBR: "Refúgio para descanso", ENUS: "Rest refuge"},
					Description: domain.LocalizedText{
						PTBR: "Uma área defensável permite um Descanso Curto de 1 hora, se o grupo a mantiver segura.",
						ENUS: "A defensible area allows a 1-hour Short Rest if the party keeps it secure.",
					},
					Source: domain.SRDTitle,
				})
			}
		}
	}
	return encounters, treasures, puzzles, traps, restPoints
}

func composeEncounter(spec domain.AdventureSpec, difficulty, floorID, roomID, id string, rng *RNG) domain.Encounter {
	budget, tier, _ := domain.EncounterBudget(spec.PartyLevel, spec.PartySize, difficulty)
	available := make([]domain.CatalogItem, 0)
	for _, creature := range domain.MonsterCatalogItems() {
		if creature.XP <= budget {
			available = append(available, creature)
		}
	}
	sort.Slice(available, func(i, j int) bool { return available[i].XP < available[j].XP })
	windowStart := max(0, len(available)-4)
	selected := available[windowStart+rng.Intn(len(available)-windowStart)]
	count := clamp(budget/selected.XP, 1, 8)
	creature := domain.EncounterCreature{Index: selected.Index, Name: selected.Name, Count: count, CR: selected.CR, XP: selected.XP}
	return domain.Encounter{
		ID: id, FloorID: floorID, RoomID: roomID, Difficulty: difficulty,
		Creatures: []domain.EncounterCreature{creature}, BudgetXP: budget, BudgetTier: tier,
		TotalXP: selected.XP * count,
	}
}

func addEncounterMarker(floor *domain.FloorMap, encounter domain.Encounter, position domain.GridPosition) {
	creature := encounter.Creatures[0]
	floor.Entities = append(floor.Entities, domain.SceneEntity{
		ID: "token-" + encounter.ID, Kind: "token", Name: creature.Name, Position: position,
		AssetID: "mini-hostile", BlocksMovement: true, RoomID: encounter.RoomID,
	})
}

func buildTreasure(spec domain.AdventureSpec, floorID, roomID string, floorIndex int) domain.Treasure {
	multiplier := map[string]int{"poor": 25, "standard": 50, "rich": 100, "legendary": 200}[spec.TreasureQuality]
	value := multiplier * spec.PartyLevel * spec.PartySize
	return domain.Treasure{
		ID: fmt.Sprintf("treasure-%d", floorIndex+1), FloorID: floorID, RoomID: roomID,
		Name: domain.LocalizedText{PTBR: "Reserva oculta", ENUS: "Hidden cache"},
		Description: domain.LocalizedText{
			PTBR: "Uma recompensa opcional ligada às pistas e riscos deste andar.",
			ENUS: "An optional reward tied to this floor's clues and risks.",
		},
		Quality: spec.TreasureQuality, ValueGP: value,
		Contents: []domain.LocalizedText{
			{PTBR: fmt.Sprintf("Moedas e objetos de arte no valor total de %d PO", value), ENUS: fmt.Sprintf("Coins and art objects worth %d GP in total", value)},
			{PTBR: "Uma relíquia original relacionada ao tema da aventura", ENUS: "An original relic tied to the adventure theme"},
		},
		Source: originalContentSource,
	}
}

func buildPuzzle(spec domain.AdventureSpec, floorID, roomID string, floorIndex int, rng *RNG) domain.Puzzle {
	templates := []struct{ name, prompt, solution, hint domain.LocalizedText }{
		{
			name:     domain.LocalizedText{PTBR: "Sequência das runas", ENUS: "Rune sequence"},
			prompt:   domain.LocalizedText{PTBR: "Quatro runas repetem o ciclo do tema em uma ordem incompleta.", ENUS: "Four runes repeat the theme's cycle with one step missing."},
			solution: domain.LocalizedText{PTBR: "Ativar as runas na ordem indicada pelas marcas mais antigas para as mais recentes.", ENUS: "Activate the runes from the oldest marks to the newest."},
			hint:     domain.LocalizedText{PTBR: "O desgaste revela qual símbolo foi tocado primeiro.", ENUS: "Wear reveals which symbol was touched first."},
		},
		{
			name:     domain.LocalizedText{PTBR: "Altar em equilíbrio", ENUS: "Balanced altar"},
			prompt:   domain.LocalizedText{PTBR: "Três pratos respondem a peso, luz e silêncio.", ENUS: "Three plates answer to weight, light, and silence."},
			solution: domain.LocalizedText{PTBR: "Equilibrar os pratos e apagar as fontes de luz antes de tocar o selo central.", ENUS: "Balance the plates and extinguish the lights before touching the central seal."},
			hint:     domain.LocalizedText{PTBR: "A inscrição associa cada prato a uma ausência.", ENUS: "The inscription links each plate to an absence."},
		},
	}
	selected := templates[rng.Intn(len(templates))]
	return domain.Puzzle{
		ID: fmt.Sprintf("puzzle-%d", floorIndex+1), FloorID: floorID, RoomID: roomID,
		Name: selected.name, Prompt: selected.prompt, Solution: selected.solution, Hint: selected.hint,
		CheckDC: clamp(10+spec.PartyLevel/2, 10, 20), Source: originalContentSource,
	}
}

func buildTrap(spec domain.AdventureSpec, floorID, roomID string, floorIndex int, rng *RNG) domain.Trap {
	index := []string{"collapsing-roof", "poisoned-needle"}[rng.Intn(2)]
	item, _ := domain.CatalogItemByIndex(index)
	tier := domain.LevelTier(spec.PartyLevel)
	damage, saveDC, severity := trapScale(index, tier)
	trigger := domain.LocalizedText{PTBR: "Uma criatura cruza o mecanismo oculto.", ENUS: "A creature crosses the hidden mechanism."}
	if index == "poisoned-needle" {
		trigger = domain.LocalizedText{PTBR: "Uma criatura abre a fechadura incorretamente.", ENUS: "A creature opens the lock incorrectly."}
	}
	return domain.Trap{
		ID: fmt.Sprintf("trap-%d", floorIndex+1), FloorID: floorID, RoomID: roomID,
		CatalogIndex: index, Name: item.Name, Severity: severity, LevelTier: tier, Trigger: trigger,
		DetectionDC: 15, DisableDC: 15, SaveDC: saveDC, Damage: damage,
		Source: item.Source, License: item.License, Hidden: true,
	}
}

func trapScale(index, tier string) (domain.LocalizedText, int, string) {
	if index == "poisoned-needle" {
		values := map[string]struct {
			damage domain.LocalizedText
			dc     int
		}{
			"1-4":   {damage: domain.LocalizedText{PTBR: "5 (1d10) de veneno", ENUS: "5 (1d10) poison"}, dc: 11},
			"5-10":  {damage: domain.LocalizedText{PTBR: "11 (2d10) de veneno", ENUS: "11 (2d10) poison"}, dc: 13},
			"11-16": {damage: domain.LocalizedText{PTBR: "22 (4d10) de veneno", ENUS: "22 (4d10) poison"}, dc: 15},
			"17-20": {damage: domain.LocalizedText{PTBR: "55 (10d10) de veneno", ENUS: "55 (10d10) poison"}, dc: 17},
		}
		value := values[tier]
		return value.damage, value.dc, "nuisance"
	}
	values := map[string]struct {
		damage domain.LocalizedText
		dc     int
	}{
		"1-4":   {damage: domain.LocalizedText{PTBR: "11 (2d10) de concussão", ENUS: "11 (2d10) bludgeoning"}, dc: 13},
		"5-10":  {damage: domain.LocalizedText{PTBR: "22 (4d10) de concussão", ENUS: "22 (4d10) bludgeoning"}, dc: 15},
		"11-16": {damage: domain.LocalizedText{PTBR: "55 (10d10) de concussão", ENUS: "55 (10d10) bludgeoning"}, dc: 17},
		"17-20": {damage: domain.LocalizedText{PTBR: "99 (18d10) de concussão", ENUS: "99 (18d10) bludgeoning"}, dc: 19},
	}
	value := values[tier]
	return value.damage, value.dc, "deadly"
}

func firstRoom(rooms []domain.Room, predicate func(domain.Room) bool) domain.Room {
	for _, room := range rooms {
		if predicate(room) {
			return room
		}
	}
	return domain.Room{}
}

func sumEncounterBudgets(encounters []domain.Encounter) int {
	total := 0
	for _, encounter := range encounters {
		total += encounter.BudgetXP
	}
	return total
}

func sumEncounterXP(encounters []domain.Encounter) int {
	total := 0
	for _, encounter := range encounters {
		total += encounter.TotalXP
	}
	return total
}

func sumTreasureValue(treasures []domain.Treasure) int {
	total := 0
	for _, treasure := range treasures {
		total += treasure.ValueGP
	}
	return total
}
