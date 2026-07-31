package domain

import (
	"fmt"
	"math"
	"strings"
)

func validateRulesContent(document AdventureDocument, roomFloors map[string]string, rooms map[string]Room) error {
	if document.RulesVersion != RulesVersion {
		return invalidAdventure("unsupported rules version %q", document.RulesVersion)
	}
	if !hasSRDAttribution(document.Attributions) {
		return invalidAdventure("rules content requires the SRD 5.2.1 attribution")
	}

	encounterIDs := make(map[string]struct{}, len(document.Encounters))
	budgetTotal, xpTotal := 0, 0
	climaxEncounters := 0
	for _, encounter := range document.Encounters {
		if strings.TrimSpace(encounter.ID) == "" {
			return invalidAdventure("encounter id is required")
		}
		if _, duplicate := encounterIDs[encounter.ID]; duplicate {
			return invalidAdventure("duplicate encounter id %q", encounter.ID)
		}
		encounterIDs[encounter.ID] = struct{}{}
		if roomFloors[encounter.RoomID] != encounter.FloorID {
			return invalidAdventure("encounter %q references an invalid room", encounter.ID)
		}
		if !oneOfValue(encounter.Difficulty, "easy", "medium", "hard", "deadly") {
			return invalidAdventure("encounter %q has unsupported difficulty", encounter.ID)
		}
		expectedBudget, expectedTier, ok := EncounterBudget(document.Spec.PartyLevel, document.Spec.PartySize, encounter.Difficulty)
		if !ok || encounter.BudgetXP != expectedBudget || encounter.BudgetTier != expectedTier {
			return invalidAdventure("encounter %q budget does not match SRD 5.2.1", encounter.ID)
		}
		if len(encounter.Creatures) == 0 {
			return invalidAdventure("encounter %q has no creatures", encounter.ID)
		}
		actualXP := 0
		for _, creature := range encounter.Creatures {
			catalogItem, exists := CatalogItemByIndex(creature.Index)
			if !exists || catalogItem.Kind != "monster" {
				return invalidAdventure("encounter %q references unknown creature %q", encounter.ID, creature.Index)
			}
			if creature.Count < 1 || creature.Count > 8 || creature.XP != catalogItem.XP ||
				math.Abs(creature.CR-catalogItem.CR) > .0001 {
				return invalidAdventure("encounter %q has invalid catalog data for %q", encounter.ID, creature.Index)
			}
			if err := validateLocalized("encounter creature "+creature.Index, creature.Name); err != nil {
				return err
			}
			actualXP += creature.XP * creature.Count
		}
		if actualXP != encounter.TotalXP || actualXP > encounter.BudgetXP {
			return invalidAdventure("encounter %q exceeds or misreports its XP budget", encounter.ID)
		}
		if encounter.RoomID == document.Progression.ClimaxRoomID {
			climaxEncounters++
			if encounter.Difficulty != "deadly" {
				return invalidAdventure("climax encounter must be deadly")
			}
		} else if encounter.Difficulty == "deadly" {
			return invalidAdventure("deadly difficulty is reserved for the climax")
		}
		budgetTotal += encounter.BudgetXP
		xpTotal += encounter.TotalXP
	}
	if climaxEncounters != 1 {
		return invalidAdventure("document requires exactly one climax encounter")
	}

	if len(document.Treasures) != len(document.Floors) || len(document.Puzzles) != len(document.Floors) || len(document.Traps) != len(document.Floors) {
		return invalidAdventure("rules content requires one treasure, puzzle, and trap per floor")
	}
	treasureTotal := 0
	seenTreasures := make(map[string]struct{})
	for _, treasure := range document.Treasures {
		if err := validateContentLocation("treasure", treasure.ID, treasure.FloorID, treasure.RoomID, roomFloors, seenTreasures); err != nil {
			return err
		}
		if !rooms[treasure.RoomID].Secret {
			return invalidAdventure("treasure %q must reward an optional secret room", treasure.ID)
		}
		if err := validateLocalized("treasure "+treasure.ID+" name", treasure.Name); err != nil {
			return err
		}
		if err := validateLocalized("treasure "+treasure.ID+" description", treasure.Description); err != nil {
			return err
		}
		if !oneOfValue(treasure.Quality, "poor", "standard", "rich", "legendary") || treasure.Quality != document.Spec.TreasureQuality || treasure.ValueGP < 1 || len(treasure.Contents) == 0 {
			return invalidAdventure("treasure %q has invalid budget or quality", treasure.ID)
		}
		for index, content := range treasure.Contents {
			if err := validateLocalized(fmt.Sprintf("treasure %s content %d", treasure.ID, index), content); err != nil {
				return err
			}
		}
		if strings.TrimSpace(treasure.Source) == "" {
			return invalidAdventure("treasure %q requires a source", treasure.ID)
		}
		treasureTotal += treasure.ValueGP
	}

	seenPuzzles := make(map[string]struct{})
	for _, puzzle := range document.Puzzles {
		if err := validateContentLocation("puzzle", puzzle.ID, puzzle.FloorID, puzzle.RoomID, roomFloors, seenPuzzles); err != nil {
			return err
		}
		for field, value := range map[string]LocalizedText{"name": puzzle.Name, "prompt": puzzle.Prompt, "solution": puzzle.Solution, "hint": puzzle.Hint} {
			if err := validateLocalized("puzzle "+puzzle.ID+" "+field, value); err != nil {
				return err
			}
		}
		if puzzle.CheckDC < 5 || puzzle.CheckDC > 30 || strings.TrimSpace(puzzle.Source) == "" {
			return invalidAdventure("puzzle %q has invalid rules", puzzle.ID)
		}
	}

	seenTraps := make(map[string]struct{})
	for _, trap := range document.Traps {
		if err := validateContentLocation("trap", trap.ID, trap.FloorID, trap.RoomID, roomFloors, seenTraps); err != nil {
			return err
		}
		catalogItem, exists := CatalogItemByIndex(trap.CatalogIndex)
		if !exists || catalogItem.Kind != "trap" || trap.Source != catalogItem.Source || trap.License != catalogItem.License {
			return invalidAdventure("trap %q has an invalid catalog reference", trap.ID)
		}
		if !rooms[trap.RoomID].Secret || !trap.Hidden || trap.LevelTier != LevelTier(document.Spec.PartyLevel) {
			return invalidAdventure("trap %q has invalid secrecy or level tier", trap.ID)
		}
		if err := validateLocalized("trap "+trap.ID+" name", trap.Name); err != nil {
			return err
		}
		if err := validateLocalized("trap "+trap.ID+" trigger", trap.Trigger); err != nil {
			return err
		}
		if err := validateLocalized("trap "+trap.ID+" damage", trap.Damage); err != nil {
			return err
		}
		if !oneOfValue(trap.Severity, "nuisance", "deadly") || trap.DetectionDC < 5 || trap.DetectionDC > 30 || trap.DisableDC < 5 || trap.DisableDC > 30 || trap.SaveDC < 5 || trap.SaveDC > 30 {
			return invalidAdventure("trap %q has invalid rules", trap.ID)
		}
	}

	expectedRestPoints := 0
	if document.Spec.DurationHours >= 3 {
		expectedRestPoints = len(document.Floors)
	}
	if len(document.RestPoints) != expectedRestPoints {
		return invalidAdventure("rest point count does not match adventure duration")
	}
	seenRestPoints := make(map[string]struct{})
	for _, rest := range document.RestPoints {
		if err := validateContentLocation("rest point", rest.ID, rest.FloorID, rest.RoomID, roomFloors, seenRestPoints); err != nil {
			return err
		}
		if rest.Kind != "short" {
			return invalidAdventure("rest point %q has unsupported kind", rest.ID)
		}
		if err := validateLocalized("rest point "+rest.ID+" name", rest.Name); err != nil {
			return err
		}
		if err := validateLocalized("rest point "+rest.ID+" description", rest.Description); err != nil {
			return err
		}
		if strings.TrimSpace(rest.Source) == "" {
			return invalidAdventure("rest point %q requires a source", rest.ID)
		}
	}

	expectedCounts := ContentCounts{Encounters: len(document.Encounters), Treasures: len(document.Treasures), Puzzles: len(document.Puzzles), Traps: len(document.Traps), RestPoints: len(document.RestPoints)}
	if document.Analysis.EncounterBudgetXP != budgetTotal || document.Analysis.EncounterTotalXP != xpTotal ||
		document.Analysis.TreasureValueGP != treasureTotal || document.Analysis.ContentCounts != expectedCounts {
		return invalidAdventure("content analysis does not match document")
	}
	return nil
}

func validateContentLocation(kind, id, floorID, roomID string, roomFloors map[string]string, seen map[string]struct{}) error {
	if strings.TrimSpace(id) == "" {
		return invalidAdventure("%s id is required", kind)
	}
	if _, duplicate := seen[id]; duplicate {
		return invalidAdventure("duplicate %s id %q", kind, id)
	}
	seen[id] = struct{}{}
	if roomFloors[roomID] != floorID {
		return invalidAdventure("%s %q references an invalid room", kind, id)
	}
	return nil
}

func hasSRDAttribution(attributions []Attribution) bool {
	for _, attribution := range attributions {
		if attribution.Title == SRDTitle && attribution.Source == SRDSource && attribution.License == SRDLicense && strings.Contains(attribution.Notice, "SRD 5.2.1") {
			return true
		}
	}
	return false
}
