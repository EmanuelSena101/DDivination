package domain

import (
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidAdventure = errors.New("invalid adventure document")

func ValidateAdventure(document AdventureDocument) error {
	if strings.TrimSpace(document.ID) == "" {
		return invalidAdventure("id is required")
	}
	if document.SchemaVersion != SchemaVersion {
		return invalidAdventure("unsupported schema version %q", document.SchemaVersion)
	}
	if document.Version < 1 {
		return invalidAdventure("version must be positive")
	}
	if err := validateLocalized("name", document.Name); err != nil {
		return err
	}
	if err := validateLocalized("summary", document.Summary); err != nil {
		return err
	}
	for field, value := range map[string]LocalizedText{
		"narrative.hook":       document.Narrative.Hook,
		"narrative.objective":  document.Narrative.Objective,
		"narrative.antagonist": document.Narrative.Antagonist,
		"narrative.atmosphere": document.Narrative.Atmosphere,
	} {
		if err := validateLocalized(field, value); err != nil {
			return err
		}
	}
	if len(document.Floors) < 1 || len(document.Floors) > 5 {
		return invalidAdventure("floor count must be between 1 and 5")
	}

	floors := make(map[string]FloorMap, len(document.Floors))
	roomFloors := make(map[string]string)
	rooms := make(map[string]Room)
	entities := make(map[string]SceneEntity)
	portals := make(map[string]Portal)
	walls := make(map[string]WallEdge)
	wallFloors := make(map[string]string)
	indices := make(map[int]struct{})
	for _, floor := range document.Floors {
		if strings.TrimSpace(floor.ID) == "" {
			return invalidAdventure("floor id is required")
		}
		if _, duplicate := floors[floor.ID]; duplicate {
			return invalidAdventure("duplicate floor id %q", floor.ID)
		}
		if _, duplicate := indices[floor.Index]; duplicate {
			return invalidAdventure("duplicate floor index %d", floor.Index)
		}
		if floor.Width < 1 || floor.Width > 128 || floor.Height < 1 || floor.Height > 128 {
			return invalidAdventure("floor %q dimensions must be between 1 and 128", floor.ID)
		}
		if err := validateLocalized("floor "+floor.ID+" name", floor.Name); err != nil {
			return err
		}
		if len(floor.Tiles) == 0 {
			return invalidAdventure("floor %q has no tiles", floor.ID)
		}
		floors[floor.ID] = floor
		indices[floor.Index] = struct{}{}

		tiles := make(map[GridPosition]Tile, len(floor.Tiles))
		for _, tile := range floor.Tiles {
			position := GridPosition{X: tile.X, Z: tile.Z}
			if !inside(floor, position) {
				return invalidAdventure("tile %d,%d is outside floor %q", tile.X, tile.Z, floor.ID)
			}
			if _, duplicate := tiles[position]; duplicate {
				return invalidAdventure("duplicate tile %d,%d on floor %q", tile.X, tile.Z, floor.ID)
			}
			if !oneOfValue(tile.Kind, "floor", "corridor", "stairs", "water", "lava") {
				return invalidAdventure("unsupported tile kind %q", tile.Kind)
			}
			tiles[position] = tile
		}

		edges := make(map[string]struct{}, len(floor.Walls))
		for _, wall := range floor.Walls {
			position := GridPosition{X: wall.X, Z: wall.Z}
			if _, ok := tiles[position]; !ok {
				return invalidAdventure("wall at %d,%d has no tile on floor %q", wall.X, wall.Z, floor.ID)
			}
			if !oneOfValue(wall.Direction, "north", "east", "south", "west") ||
				!oneOfValue(wall.Kind, "wall", "door", "secret-door") {
				return invalidAdventure("unsupported wall on floor %q", floor.ID)
			}
			key := fmt.Sprintf("%d:%d:%s", wall.X, wall.Z, wall.Direction)
			if _, duplicate := edges[key]; duplicate {
				return invalidAdventure("duplicate wall %s on floor %q", key, floor.ID)
			}
			edges[key] = struct{}{}
			if wall.ID != "" {
				if _, duplicate := walls[wall.ID]; duplicate {
					return invalidAdventure("duplicate wall id %q", wall.ID)
				}
				walls[wall.ID] = wall
				wallFloors[wall.ID] = floor.ID
			}
			if wall.RequiredKeyID != "" && (wall.ID == "" || wall.Kind != "door" || !wall.Locked) {
				return invalidAdventure("keyed wall on floor %q must be an identified locked door", floor.ID)
			}
		}

		for _, room := range floor.Rooms {
			if strings.TrimSpace(room.ID) == "" {
				return invalidAdventure("room id is required on floor %q", floor.ID)
			}
			if _, duplicate := roomFloors[room.ID]; duplicate {
				return invalidAdventure("duplicate room id %q", room.ID)
			}
			if _, ok := tiles[room.Center]; !ok {
				return invalidAdventure("room %q center has no tile", room.ID)
			}
			if err := validateLocalized("room "+room.ID+" name", room.Name); err != nil {
				return err
			}
			if err := validateLocalized("room "+room.ID+" description", room.Description); err != nil {
				return err
			}
			roomFloors[room.ID] = floor.ID
			rooms[room.ID] = room
		}
		for _, entity := range floor.Entities {
			if strings.TrimSpace(entity.ID) == "" {
				return invalidAdventure("entity id is required on floor %q", floor.ID)
			}
			if _, duplicate := entities[entity.ID]; duplicate {
				return invalidAdventure("duplicate entity id %q", entity.ID)
			}
			if _, ok := tiles[entity.Position]; !ok {
				return invalidAdventure("entity %q has no tile at %d,%d", entity.ID, entity.Position.X, entity.Position.Z)
			}
			if !oneOfValue(entity.Kind, "prop", "light", "trap", "token", "marker", "key", "boss") {
				return invalidAdventure("unsupported entity kind %q", entity.Kind)
			}
			if err := validateLocalized("entity "+entity.ID+" name", entity.Name); err != nil {
				return err
			}
			if entity.RoomID != "" && roomFloors[entity.RoomID] != floor.ID {
				return invalidAdventure("entity %q references an invalid room", entity.ID)
			}
			entities[entity.ID] = entity
		}
		for _, portal := range floor.Portals {
			if strings.TrimSpace(portal.ID) == "" {
				return invalidAdventure("portal id is required on floor %q", floor.ID)
			}
			if _, duplicate := portals[portal.ID]; duplicate {
				return invalidAdventure("duplicate portal id %q", portal.ID)
			}
			if portal.FromFloorID != floor.ID {
				return invalidAdventure("portal %q source floor mismatch", portal.ID)
			}
			if _, ok := tiles[portal.From]; !ok {
				return invalidAdventure("portal %q source has no tile", portal.ID)
			}
			portals[portal.ID] = portal
		}
		if err := validateWalkableRooms(floor, tiles); err != nil {
			return err
		}
	}

	floorLinks := make(map[string]map[string]struct{}, len(floors))
	for _, floor := range document.Floors {
		for _, portal := range floor.Portals {
			target, ok := floors[portal.ToFloorID]
			if !ok {
				return invalidAdventure("portal %q targets unknown floor %q", portal.ID, portal.ToFloorID)
			}
			if !hasTile(target, portal.To) {
				return invalidAdventure("portal %q target has no tile", portal.ID)
			}
			if floorLinks[floor.ID] == nil {
				floorLinks[floor.ID] = make(map[string]struct{})
			}
			floorLinks[floor.ID][portal.ToFloorID] = struct{}{}
		}
	}
	if err := validateFloorConnectivity(document.Floors[0].ID, floors, floorLinks); err != nil {
		return err
	}
	if err := validatePortalPairs(portals); err != nil {
		return err
	}
	for _, encounter := range document.Encounters {
		if actualFloor, ok := roomFloors[encounter.RoomID]; !ok || actualFloor != encounter.FloorID {
			return invalidAdventure("encounter %q references an invalid room", encounter.ID)
		}
	}
	if document.Analysis.TotalFloors != len(document.Floors) {
		return invalidAdventure("analysis totalFloors does not match document")
	}
	if document.Analysis.TotalRooms != len(roomFloors) {
		return invalidAdventure("analysis totalRooms does not match document")
	}
	for _, roomID := range append(append([]string{}, document.Analysis.CriticalPath...), document.Analysis.DeadEnds...) {
		if _, ok := roomFloors[roomID]; !ok {
			return invalidAdventure("analysis references unknown room %q", roomID)
		}
	}
	if err := validateProgression(document, roomFloors, rooms, entities, portals, walls, wallFloors); err != nil {
		return err
	}
	if document.RulesVersion != "" {
		if err := validateRulesContent(document, roomFloors, rooms); err != nil {
			return err
		}
	}
	return nil
}

func validatePortalPairs(portals map[string]Portal) error {
	for _, portal := range portals {
		paired := false
		for _, candidate := range portals {
			if candidate.FromFloorID == portal.ToFloorID && candidate.ToFloorID == portal.FromFloorID &&
				candidate.From == portal.To && candidate.To == portal.From {
				paired = true
				break
			}
		}
		if !paired {
			return invalidAdventure("portal %q has no coherent return pair", portal.ID)
		}
	}
	return nil
}

func validateWalkableRooms(floor FloorMap, tiles map[GridPosition]Tile) error {
	var start GridPosition
	found := false
	for _, room := range floor.Rooms {
		if room.Mandatory {
			start = room.Center
			found = true
			break
		}
	}
	if !found {
		return nil
	}
	if tile := tiles[start]; !tile.Walkable {
		return invalidAdventure("mandatory room on floor %q starts on a non-walkable tile", floor.ID)
	}
	seen := map[GridPosition]struct{}{start: {}}
	queue := []GridPosition{start}
	directions := []GridPosition{{X: 1}, {X: -1}, {Z: 1}, {Z: -1}}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, direction := range directions {
			next := GridPosition{X: current.X + direction.X, Z: current.Z + direction.Z}
			tile, ok := tiles[next]
			if !ok || !tile.Walkable {
				continue
			}
			if _, visited := seen[next]; visited {
				continue
			}
			seen[next] = struct{}{}
			queue = append(queue, next)
		}
	}
	for _, room := range floor.Rooms {
		if !room.Mandatory {
			continue
		}
		if _, ok := seen[room.Center]; !ok {
			return invalidAdventure("mandatory room %q is unreachable", room.ID)
		}
	}
	return nil
}

func validateFloorConnectivity(start string, floors map[string]FloorMap, links map[string]map[string]struct{}) error {
	seen := map[string]struct{}{start: {}}
	queue := []string{start}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for next := range links[current] {
			if _, visited := seen[next]; visited {
				continue
			}
			seen[next] = struct{}{}
			queue = append(queue, next)
		}
	}
	if len(seen) != len(floors) {
		return invalidAdventure("not every floor is reachable from the first floor")
	}
	return nil
}

func validateLocalized(field string, value LocalizedText) error {
	if strings.TrimSpace(value.PTBR) == "" || strings.TrimSpace(value.ENUS) == "" {
		return invalidAdventure("%s requires pt-BR and en-US", field)
	}
	return nil
}

func inside(floor FloorMap, position GridPosition) bool {
	return position.X >= 0 && position.Z >= 0 && position.X < floor.Width && position.Z < floor.Height
}

func hasTile(floor FloorMap, position GridPosition) bool {
	for _, tile := range floor.Tiles {
		if tile.X == position.X && tile.Z == position.Z {
			return true
		}
	}
	return false
}

func oneOfValue(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}

func invalidAdventure(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidAdventure, fmt.Sprintf(format, args...))
}
