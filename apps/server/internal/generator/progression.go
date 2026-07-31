package generator

import (
	"fmt"
	"sort"

	"github.com/EmanuelSena101/DDivination/apps/server/internal/domain"
)

func buildProgression(floors []domain.FloorMap, paths [][]string) (domain.DungeonProgression, []domain.FloorMap) {
	grants := make(map[string][]string, len(floors))
	requires := make(map[string][]string, len(floors))
	keyRooms := make(map[string]bool, len(floors))
	locks := make([]domain.ProgressionLock, 0, len(floors))
	secretRoomIDs := make([]string, 0, len(floors))

	for floorIndex := range floors {
		floor := &floors[floorIndex]
		path := paths[floorIndex]
		for _, room := range floor.Rooms {
			if room.Secret {
				secretRoomIDs = append(secretRoomIDs, room.ID)
			}
		}

		keyRoomID := path[min(1, len(path)-1)]
		keyID := fmt.Sprintf("progression-key-%d", floorIndex+1)
		keyRooms[keyRoomID] = true
		grants[keyRoomID] = append(grants[keyRoomID], keyID)
		keyRoom := roomByID(*floor, keyRoomID)
		floor.Entities = append(floor.Entities, domain.SceneEntity{
			ID:       keyID,
			Kind:     "key",
			Name:     progressionKeyName(floorIndex),
			Position: freeRoomPosition(*floor, keyRoom),
			AssetID:  "prop-key",
			RoomID:   keyRoomID,
		})

		lock := domain.ProgressionLock{
			ID:         fmt.Sprintf("progression-lock-%d", floorIndex+1),
			FloorID:    floor.ID,
			FromRoomID: path[len(path)-1],
			KeyID:      keyID,
		}
		if floorIndex < len(floors)-1 {
			lock.Kind = "portal"
			lock.ToRoomID = paths[floorIndex+1][0]
			for portalIndex := range floor.Portals {
				portal := &floor.Portals[portalIndex]
				if portal.Kind != "stairs-down" {
					continue
				}
				portal.RequiredKeyID = keyID
				lock.TargetID = portal.ID
				break
			}
		} else {
			lock.Kind = "door"
			lock.FromRoomID = path[len(path)-2]
			lock.ToRoomID = path[len(path)-1]
			lock.TargetID = addProgressionDoor(floor, lock.ID, keyID, lock.FromRoomID, lock.ToRoomID)
		}
		requires[lock.ToRoomID] = append(requires[lock.ToRoomID], keyID)
		locks = append(locks, lock)
	}

	steps := make([]domain.ProgressionStep, 0)
	for floorIndex, path := range paths {
		for roomIndex, roomID := range path {
			room := roomByID(floors[floorIndex], roomID)
			kind := "exploration"
			switch {
			case floorIndex == 0 && roomIndex == 0:
				kind = "entrance"
			case floorIndex == len(paths)-1 && roomIndex == len(path)-1:
				kind = "climax"
			case keyRooms[roomID]:
				kind = "key"
			case floorIndex < len(paths)-1 && roomIndex == len(path)-1:
				kind = "transition"
			}
			steps = append(steps, domain.ProgressionStep{
				Order:          len(steps) + 1,
				FloorID:        floors[floorIndex].ID,
				RoomID:         roomID,
				Kind:           kind,
				Beat:           progressionBeat(kind, room),
				GrantsKeyIDs:   nonNilStrings(grants[roomID]),
				RequiresKeyIDs: nonNilStrings(requires[roomID]),
			})
		}
	}

	return domain.DungeonProgression{
		EntryRoomID:     steps[0].RoomID,
		ObjectiveRoomID: steps[len(steps)-1].RoomID,
		ClimaxRoomID:    steps[len(steps)-1].RoomID,
		Steps:           steps,
		Locks:           locks,
		SecretRoomIDs:   secretRoomIDs,
		Solvable:        true,
	}, floors
}

func addProgressionDoor(floor *domain.FloorMap, lockID, keyID, fromRoomID, toRoomID string) string {
	from := roomByID(*floor, fromRoomID).Center
	to := roomByID(*floor, toRoomID).Center
	direction := "east"
	if abs(to.Z-from.Z) > abs(to.X-from.X) {
		if to.Z < from.Z {
			direction = "north"
		} else {
			direction = "south"
		}
	} else if to.X < from.X {
		direction = "west"
	}
	targetID := lockID + "-door"
	floor.Walls = append(floor.Walls, domain.WallEdge{
		ID:            targetID,
		X:             from.X,
		Z:             from.Z,
		Direction:     direction,
		Kind:          "door",
		Locked:        true,
		RequiredKeyID: keyID,
	})
	sort.Slice(floor.Walls, func(i, j int) bool {
		if floor.Walls[i].Z != floor.Walls[j].Z {
			return floor.Walls[i].Z < floor.Walls[j].Z
		}
		if floor.Walls[i].X != floor.Walls[j].X {
			return floor.Walls[i].X < floor.Walls[j].X
		}
		return floor.Walls[i].Direction < floor.Walls[j].Direction
	})
	return targetID
}

func freeRoomPosition(floor domain.FloorMap, room domain.Room) domain.GridPosition {
	occupied := make(map[domain.GridPosition]bool, len(floor.Entities))
	for _, entity := range floor.Entities {
		occupied[entity.Position] = true
	}
	candidates := []domain.GridPosition{
		{X: room.Center.X + 1, Z: room.Center.Z},
		{X: room.Center.X, Z: room.Center.Z + 1},
		{X: room.Center.X - 1, Z: room.Center.Z},
		room.Center,
	}
	for _, candidate := range candidates {
		for _, tile := range floor.Tiles {
			if tile.X == candidate.X && tile.Z == candidate.Z && tile.RoomID == room.ID && !occupied[candidate] {
				return candidate
			}
		}
	}
	return room.Center
}

func roomByID(floor domain.FloorMap, roomID string) domain.Room {
	for _, room := range floor.Rooms {
		if room.ID == roomID {
			return room
		}
	}
	return domain.Room{}
}

func progressionKeyName(floorIndex int) domain.LocalizedText {
	return domain.LocalizedText{
		PTBR: fmt.Sprintf("Chave do limiar %d", floorIndex+1),
		ENUS: fmt.Sprintf("Threshold key %d", floorIndex+1),
	}
}

func progressionBeat(kind string, room domain.Room) domain.LocalizedText {
	switch kind {
	case "entrance":
		return domain.LocalizedText{PTBR: "A expedição começa e apresenta o primeiro indício.", ENUS: "The expedition begins and reveals its first clue."}
	case "key":
		return domain.LocalizedText{PTBR: "O grupo conquista uma chave necessária para avançar.", ENUS: "The party earns a key required to advance."}
	case "transition":
		return domain.LocalizedText{PTBR: "A passagem protegida conduz ao próximo andar.", ENUS: "The guarded passage leads to the next floor."}
	case "climax":
		return domain.LocalizedText{PTBR: "O caminho culmina no confronto com o antagonista.", ENUS: "The path culminates in a confrontation with the antagonist."}
	default:
		return room.Description
	}
}

func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
