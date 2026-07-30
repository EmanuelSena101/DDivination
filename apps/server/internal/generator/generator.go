package generator

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/EmanuelSena101/DDivination/apps/server/internal/domain"
)

var ErrInvalidSpec = errors.New("invalid adventure specification")

type rect struct {
	x, z, w, h int
}

func (r rect) center() domain.GridPosition {
	return domain.GridPosition{X: r.x + r.w/2, Z: r.z + r.h/2}
}

func Generate(spec domain.AdventureSpec, seed uint64, now time.Time) (domain.AdventureDocument, error) {
	return GenerateContext(context.Background(), spec, seed, now, nil)
}

type FloorProgress func(completed, total int)

func GenerateContext(
	ctx context.Context,
	spec domain.AdventureSpec,
	seed uint64,
	now time.Time,
	onFloor FloorProgress,
) (domain.AdventureDocument, error) {
	spec = normalize(spec)
	if err := validate(spec); err != nil {
		return domain.AdventureDocument{}, err
	}
	if err := ctx.Err(); err != nil {
		return domain.AdventureDocument{}, err
	}

	rng := NewRNG(seed)
	adventureID := fmt.Sprintf("adv-%016x", seed)
	floors := make([]domain.FloorMap, 0, spec.FloorCount)
	criticalPath := make([]string, 0)
	deadEnds := make([]string, 0)
	encounters := make([]domain.Encounter, 0)

	var previousPortal *domain.Portal
	for floorIndex := 0; floorIndex < spec.FloorCount; floorIndex++ {
		if err := ctx.Err(); err != nil {
			return domain.AdventureDocument{}, err
		}
		floor, path, secretID, encounter := generateFloor(spec, rng, adventureID, floorIndex)
		criticalPath = append(criticalPath, path...)
		if secretID != "" {
			deadEnds = append(deadEnds, secretID)
		}
		if encounter != nil {
			encounters = append(encounters, *encounter)
		}

		if previousPortal != nil {
			previousPortal.ToFloorID = floor.ID
			previousPortal.To = floor.Rooms[0].Center
			floor.Portals = append(floor.Portals, domain.Portal{
				ID:          previousPortal.ID + "-return",
				FromFloorID: floor.ID,
				From:        floor.Rooms[0].Center,
				ToFloorID:   previousPortal.FromFloorID,
				To:          previousPortal.From,
				Kind:        "stairs-up",
			})
			floor.Tiles = markTile(floor.Tiles, floor.Rooms[0].Center, "stairs")
		}

		if floorIndex < spec.FloorCount-1 {
			exitRoom := floor.Rooms[len(path)-1]
			portal := domain.Portal{
				ID:          fmt.Sprintf("portal-%d-%d", floorIndex, floorIndex+1),
				FromFloorID: floor.ID,
				From:        exitRoom.Center,
				Kind:        "stairs-down",
			}
			floor.Portals = append(floor.Portals, portal)
			floor.Tiles = markTile(floor.Tiles, exitRoom.Center, "stairs")
			previousPortal = &floor.Portals[len(floor.Portals)-1]
		} else {
			previousPortal = nil
		}

		floors = append(floors, floor)
		if floorIndex > 0 {
			// Complete the prior floor's outgoing portal after the destination exists.
			prior := &floors[floorIndex-1]
			for i := range prior.Portals {
				if prior.Portals[i].Kind == "stairs-down" && prior.Portals[i].ToFloorID == "" {
					prior.Portals[i].ToFloorID = floor.ID
					prior.Portals[i].To = floor.Rooms[0].Center
				}
			}
		}
		if onFloor != nil {
			onFloor(floorIndex+1, spec.FloorCount)
		}
	}

	if err := ctx.Err(); err != nil {
		return domain.AdventureDocument{}, err
	}
	name := generatedName(spec, rng)
	doc := domain.AdventureDocument{
		ID:               adventureID,
		SchemaVersion:    domain.SchemaVersion,
		GeneratorVersion: domain.GeneratorVersion,
		Version:          1,
		Seed:             seed,
		Name:             name,
		Spec:             spec,
		Summary: domain.LocalizedText{
			PTBR: fmt.Sprintf("Uma aventura %s de %d andares para %d personagens de nível %d, ambientada em %s.", difficultyPT(spec.Difficulty), spec.FloorCount, spec.PartySize, spec.PartyLevel, spec.Biome),
			ENUS: fmt.Sprintf("A %s %d-floor adventure for %d level-%d characters, set in %s.", spec.Difficulty, spec.FloorCount, spec.PartySize, spec.PartyLevel, spec.Biome),
		},
		Narrative: domain.AdventureNarrative{
			Hook: domain.LocalizedText{
				PTBR: fmt.Sprintf("Sinais de %s levam o grupo a %s.", spec.Antagonist, spec.Biome),
				ENUS: fmt.Sprintf("Signs of %s lead the party into %s.", spec.Antagonist, spec.Biome),
			},
			Objective: domain.LocalizedText{PTBR: spec.Objective, ENUS: spec.Objective},
			Antagonist: domain.LocalizedText{
				PTBR: fmt.Sprintf("%s move seus planos nas sombras.", title(spec.Antagonist)),
				ENUS: fmt.Sprintf("%s advances its plans from the shadows.", title(spec.Antagonist)),
			},
			Atmosphere: domain.LocalizedText{
				PTBR: fmt.Sprintf("Um ambiente %s com o tema %s.", spec.Biome, spec.Theme),
				ENUS: fmt.Sprintf("A %s environment shaped by %s.", spec.Biome, spec.Theme),
			},
		},
		Floors:     floors,
		Encounters: encounters,
		Analysis: domain.DungeonAnalysis{
			TotalRooms:          countRooms(floors),
			TotalFloors:         len(floors),
			CriticalPath:        criticalPath,
			DeadEnds:            deadEnds,
			EstimatedDifficulty: spec.Difficulty,
			Invariants: []string{
				"entrance-reaches-objective",
				"portals-are-paired",
				"secrets-are-optional",
				"critical-path-is-walkable",
				"boss-is-on-final-floor",
			},
		},
		Attributions: []domain.Attribution{srdAttribution()},
		CreatedAt:    now.UTC(),
		UpdatedAt:    now.UTC(),
	}
	return doc, nil
}

func normalize(spec domain.AdventureSpec) domain.AdventureSpec {
	defaults := domain.DefaultAdventureSpec()
	if spec.PartySize == 0 {
		spec.PartySize = defaults.PartySize
	}
	if spec.PartyLevel == 0 {
		spec.PartyLevel = defaults.PartyLevel
	}
	if spec.DurationHours == 0 {
		spec.DurationHours = defaults.DurationHours
	}
	if spec.Difficulty == "" {
		spec.Difficulty = defaults.Difficulty
	}
	if strings.TrimSpace(spec.Theme) == "" {
		spec.Theme = defaults.Theme
	}
	if strings.TrimSpace(spec.Biome) == "" {
		spec.Biome = defaults.Biome
	}
	if spec.FloorCount == 0 {
		spec.FloorCount = defaults.FloorCount
	}
	if strings.TrimSpace(spec.Objective) == "" {
		spec.Objective = defaults.Objective
	}
	if strings.TrimSpace(spec.Antagonist) == "" {
		spec.Antagonist = defaults.Antagonist
	}
	if spec.StructureStyle == "" {
		spec.StructureStyle = defaults.StructureStyle
	}
	if spec.TreasureQuality == "" {
		spec.TreasureQuality = defaults.TreasureQuality
	}
	return spec
}

func validate(spec domain.AdventureSpec) error {
	if spec.PartySize < 1 || spec.PartySize > 8 ||
		spec.PartyLevel < 1 || spec.PartyLevel > 20 ||
		spec.DurationHours < 1 || spec.DurationHours > 12 ||
		spec.FloorCount < 1 || spec.FloorCount > 5 {
		return fmt.Errorf("%w: party, duration, level, or floor count out of range", ErrInvalidSpec)
	}
	if !oneOf(spec.Difficulty, "easy", "medium", "hard", "deadly") ||
		!oneOf(spec.StructureStyle, "linear", "branching", "labyrinthine") ||
		!oneOf(spec.TreasureQuality, "poor", "standard", "rich", "legendary") {
		return fmt.Errorf("%w: unsupported option", ErrInvalidSpec)
	}
	return nil
}

func generateFloor(spec domain.AdventureSpec, rng *RNG, adventureID string, floorIndex int) (domain.FloorMap, []string, string, *domain.Encounter) {
	requiredRooms := clamp(4+spec.DurationHours/2, 4, 8)
	totalRooms := requiredRooms + 1 // one optional secret branch
	columns := 3
	rows := (totalRooms + columns - 1) / columns
	width := 64
	height := max(48, rows*20+4)
	floorID := fmt.Sprintf("%s-floor-%d", adventureID, floorIndex+1)
	floor := domain.FloorMap{
		ID:     floorID,
		Index:  floorIndex,
		Name:   floorName(floorIndex),
		Width:  width,
		Height: height,
	}

	cells := make(map[domain.GridPosition]domain.Tile, width*height/2)
	roomRects := make([]rect, 0, totalRooms)
	pathIDs := make([]string, 0, requiredRooms)

	for i := 0; i < totalRooms; i++ {
		col, row := i%columns, i/columns
		w, h := 8+rng.Intn(5), 8+rng.Intn(5)
		r := rect{x: 3 + col*20 + rng.Intn(4), z: 3 + row*20 + rng.Intn(4), w: w, h: h}
		roomRects = append(roomRects, r)
		roomID := fmt.Sprintf("%s-room-%d", floorID, i+1)
		secret := i == totalRooms-1
		role := roomRole(floorIndex, spec.FloorCount, i, requiredRooms, rng)
		room := domain.Room{
			ID:          roomID,
			Name:        roomName(role, i+1),
			Role:        role,
			Description: roomDescription(role, spec),
			Center:      r.center(),
			Secret:      secret,
			Mandatory:   !secret,
		}
		floor.Rooms = append(floor.Rooms, room)
		if !secret {
			pathIDs = append(pathIDs, roomID)
		}
		carveRect(cells, r, roomID)
	}

	for i := 1; i < requiredRooms; i++ {
		carveCorridor(cells, roomRects[i-1].center(), roomRects[i].center(), rng.Bool(50))
	}
	secretIndex := totalRooms - 1
	branchFrom := 1 + rng.Intn(max(1, requiredRooms-2))
	carveCorridor(cells, roomRects[branchFrom].center(), roomRects[secretIndex].center(), rng.Bool(50))
	if spec.StructureStyle == "labyrinthine" && requiredRooms > 4 {
		carveCorridor(cells, roomRects[1].center(), roomRects[requiredRooms-2].center(), rng.Bool(50))
	}

	floor.Tiles = sortedTiles(cells)
	floor.Walls = deriveWalls(cells)
	addDecorations(&floor, rng)

	if floorIndex == 0 {
		entrance := floor.Rooms[0]
		floor.Entities = append(floor.Entities, domain.SceneEntity{
			ID:       "token-party",
			Kind:     "token",
			Name:     domain.LocalizedText{PTBR: "Grupo", ENUS: "Party"},
			Position: entrance.Center,
			RoomID:   entrance.ID,
			AssetID:  "mini-party",
		})
	}

	var encounter *domain.Encounter
	if floorIndex == spec.FloorCount-1 {
		bossRoom := floor.Rooms[requiredRooms-1]
		floor.Entities = append(floor.Entities, domain.SceneEntity{
			ID:             "token-boss",
			Kind:           "boss",
			Name:           domain.LocalizedText{PTBR: title(spec.Antagonist), ENUS: title(spec.Antagonist)},
			Position:       bossRoom.Center,
			RoomID:         bossRoom.ID,
			AssetID:        "mini-boss",
			BlocksMovement: true,
		})
		e := bossEncounter(spec, floor.ID, bossRoom.ID)
		encounter = &e
	}

	return floor, pathIDs, floor.Rooms[secretIndex].ID, encounter
}

func carveRect(cells map[domain.GridPosition]domain.Tile, r rect, roomID string) {
	for x := r.x; x < r.x+r.w; x++ {
		for z := r.z; z < r.z+r.h; z++ {
			p := domain.GridPosition{X: x, Z: z}
			cells[p] = domain.Tile{X: x, Z: z, Kind: "floor", RoomID: roomID, Walkable: true}
		}
	}
}

func carveCorridor(cells map[domain.GridPosition]domain.Tile, from, to domain.GridPosition, horizontalFirst bool) {
	carve := func(x, z int) {
		for dx := -1; dx <= 1; dx++ {
			for dz := -1; dz <= 1; dz++ {
				p := domain.GridPosition{X: x + dx, Z: z + dz}
				if _, exists := cells[p]; !exists {
					cells[p] = domain.Tile{X: p.X, Z: p.Z, Kind: "corridor", Walkable: true}
				}
			}
		}
	}
	x, z := from.X, from.Z
	if horizontalFirst {
		for x != to.X {
			carve(x, z)
			x += sign(to.X - x)
		}
		for z != to.Z {
			carve(x, z)
			z += sign(to.Z - z)
		}
	} else {
		for z != to.Z {
			carve(x, z)
			z += sign(to.Z - z)
		}
		for x != to.X {
			carve(x, z)
			x += sign(to.X - x)
		}
	}
	carve(to.X, to.Z)
}

func deriveWalls(cells map[domain.GridPosition]domain.Tile) []domain.WallEdge {
	directions := []struct {
		name   string
		dx, dz int
	}{{"north", 0, -1}, {"east", 1, 0}, {"south", 0, 1}, {"west", -1, 0}}
	walls := make([]domain.WallEdge, 0)
	for pos := range cells {
		for _, d := range directions {
			if _, ok := cells[domain.GridPosition{X: pos.X + d.dx, Z: pos.Z + d.dz}]; !ok {
				walls = append(walls, domain.WallEdge{X: pos.X, Z: pos.Z, Direction: d.name, Kind: "wall"})
			}
		}
	}
	sort.Slice(walls, func(i, j int) bool {
		if walls[i].Z != walls[j].Z {
			return walls[i].Z < walls[j].Z
		}
		if walls[i].X != walls[j].X {
			return walls[i].X < walls[j].X
		}
		return walls[i].Direction < walls[j].Direction
	})
	return walls
}

func sortedTiles(cells map[domain.GridPosition]domain.Tile) []domain.Tile {
	tiles := make([]domain.Tile, 0, len(cells))
	for _, tile := range cells {
		tiles = append(tiles, tile)
	}
	sort.Slice(tiles, func(i, j int) bool {
		if tiles[i].Z == tiles[j].Z {
			return tiles[i].X < tiles[j].X
		}
		return tiles[i].Z < tiles[j].Z
	})
	return tiles
}

func addDecorations(floor *domain.FloorMap, rng *RNG) {
	for i, room := range floor.Rooms {
		if room.Role == "entrance" || room.Role == "boss" {
			continue
		}
		offset := domain.GridPosition{X: room.Center.X + 2 - rng.Intn(5), Z: room.Center.Z + 2 - rng.Intn(5)}
		kind := "prop"
		asset := rng.Pick([]string{"prop-crate", "prop-column", "prop-brazier"})
		name := domain.LocalizedText{PTBR: "Objeto de cenário", ENUS: "Scenery prop"}
		if room.Secret {
			kind = "key"
			asset = "prop-chest"
			name = domain.LocalizedText{PTBR: "Tesouro oculto", ENUS: "Hidden treasure"}
		}
		floor.Entities = append(floor.Entities, domain.SceneEntity{
			ID:       fmt.Sprintf("%s-entity-%d", floor.ID, i+1),
			Kind:     kind,
			Name:     name,
			Position: offset,
			AssetID:  asset,
			Hidden:   room.Secret,
			RoomID:   room.ID,
		})
	}
}

func bossEncounter(spec domain.AdventureSpec, floorID, roomID string) domain.Encounter {
	cr := float64(clamp(spec.PartyLevel+1, 1, 20))
	xp := 200 * spec.PartyLevel * spec.PartySize
	return domain.Encounter{
		ID:         "encounter-final",
		FloorID:    floorID,
		RoomID:     roomID,
		Difficulty: "deadly",
		Creatures: []domain.EncounterCreature{{
			Index: "generated-antagonist",
			Name:  domain.LocalizedText{PTBR: title(spec.Antagonist), ENUS: title(spec.Antagonist)},
			Count: 1,
			CR:    cr,
			XP:    xp,
		}},
		TotalXP: xp,
	}
}

func generatedName(spec domain.AdventureSpec, rng *RNG) domain.LocalizedText {
	if spec.Name != nil {
		return *spec.Name
	}
	ptPrefixes := []string{"O Santuário", "As Profundezas", "A Fortaleza", "O Labirinto"}
	enPrefixes := []string{"The Sanctuary", "The Depths", "The Fortress", "The Labyrinth"}
	i := rng.Intn(len(ptPrefixes))
	return domain.LocalizedText{
		PTBR: fmt.Sprintf("%s de %s", ptPrefixes[i], title(spec.Antagonist)),
		ENUS: fmt.Sprintf("%s of %s", enPrefixes[i], title(spec.Antagonist)),
	}
}

func floorName(index int) domain.LocalizedText {
	return domain.LocalizedText{
		PTBR: fmt.Sprintf("Andar %d", index+1),
		ENUS: fmt.Sprintf("Floor %d", index+1),
	}
}

func roomRole(floorIndex, floorCount, index, requiredRooms int, rng *RNG) string {
	if index == 0 && floorIndex == 0 {
		return "entrance"
	}
	if index == requiredRooms {
		return "secret"
	}
	if floorIndex == floorCount-1 && index == requiredRooms-1 {
		return "boss"
	}
	if index == requiredRooms-1 {
		return "stairs"
	}
	roles := []string{"lair", "puzzle", "guard", "shrine", "vault", "rest"}
	return roles[rng.Intn(len(roles))]
}

func roomName(role string, index int) domain.LocalizedText {
	names := map[string]domain.LocalizedText{
		"entrance": {PTBR: "Entrada Esquecida", ENUS: "Forgotten Entrance"},
		"boss":     {PTBR: "Câmara do Ritual", ENUS: "Ritual Chamber"},
		"secret":   {PTBR: "Câmara Oculta", ENUS: "Hidden Chamber"},
		"stairs":   {PTBR: "Passagem Descendente", ENUS: "Descending Passage"},
		"lair":     {PTBR: "Covil", ENUS: "Lair"},
		"puzzle":   {PTBR: "Salão dos Enigmas", ENUS: "Hall of Riddles"},
		"guard":    {PTBR: "Posto da Guarda", ENUS: "Guard Post"},
		"shrine":   {PTBR: "Santuário Profanado", ENUS: "Desecrated Shrine"},
		"vault":    {PTBR: "Cofre Antigo", ENUS: "Ancient Vault"},
		"rest":     {PTBR: "Refúgio Silencioso", ENUS: "Silent Refuge"},
	}
	name := names[role]
	if name.PTBR == "" {
		return domain.LocalizedText{PTBR: fmt.Sprintf("Sala %d", index), ENUS: fmt.Sprintf("Room %d", index)}
	}
	return name
}

func roomDescription(role string, spec domain.AdventureSpec) domain.LocalizedText {
	return domain.LocalizedText{
		PTBR: fmt.Sprintf("Uma área %s marcada pelo tema %s e pelos sinais de %s.", role, spec.Theme, spec.Antagonist),
		ENUS: fmt.Sprintf("A %s area shaped by the %s theme and signs of %s.", role, spec.Theme, spec.Antagonist),
	}
}

func markTile(tiles []domain.Tile, pos domain.GridPosition, kind string) []domain.Tile {
	for i := range tiles {
		if tiles[i].X == pos.X && tiles[i].Z == pos.Z {
			tiles[i].Kind = kind
			return tiles
		}
	}
	return tiles
}

func srdAttribution() domain.Attribution {
	return domain.Attribution{
		Title:   "System Reference Document 5.2.1",
		Creator: "Wizards of the Coast LLC",
		Source:  "https://www.dndbeyond.com/srd",
		License: "CC-BY-4.0",
		Notice:  "This work includes material from the System Reference Document 5.2.1 by Wizards of the Coast LLC, available at https://www.dndbeyond.com/srd and licensed under CC-BY-4.0.",
	}
}

func oneOf(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}

func countRooms(floors []domain.FloorMap) int {
	total := 0
	for _, floor := range floors {
		total += len(floor.Rooms)
	}
	return total
}

func difficultyPT(value string) string {
	return map[string]string{"easy": "fácil", "medium": "média", "hard": "difícil", "deadly": "mortal"}[value]
}

func title(value string) string {
	return strings.Title(strings.TrimSpace(value))
}

func clamp(value, low, high int) int {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func sign(value int) int {
	if value < 0 {
		return -1
	}
	if value > 0 {
		return 1
	}
	return 0
}
