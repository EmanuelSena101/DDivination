package domain

import "fmt"

func validateProgression(
	document AdventureDocument,
	roomFloors map[string]string,
	rooms map[string]Room,
	entities map[string]SceneEntity,
	portals map[string]Portal,
	walls map[string]WallEdge,
	wallFloors map[string]string,
) error {
	progression := document.Progression
	if len(progression.Steps) == 0 {
		if document.GeneratorVersion == GeneratorVersion {
			return invalidAdventure("generator %q requires semantic progression", document.GeneratorVersion)
		}
		return nil
	}
	if !progression.Solvable {
		return invalidAdventure("progression must be marked solvable after validation")
	}
	if progression.EntryRoomID == "" || progression.ObjectiveRoomID == "" || progression.ClimaxRoomID == "" {
		return invalidAdventure("progression entry, objective, and climax are required")
	}
	if progression.ObjectiveRoomID != progression.ClimaxRoomID {
		return invalidAdventure("progression objective must resolve at the climax in generator v1")
	}

	stepsByRoom := make(map[string]int, len(progression.Steps))
	heldKeys := make(map[string]bool)
	grantedAt := make(map[string]int)
	for index, step := range progression.Steps {
		if step.Order != index+1 {
			return invalidAdventure("progression step order must be contiguous from 1")
		}
		if !oneOfValue(step.Kind, "entrance", "exploration", "key", "transition", "climax") {
			return invalidAdventure("unsupported progression step kind %q", step.Kind)
		}
		if roomFloors[step.RoomID] != step.FloorID {
			return invalidAdventure("progression step %d references an invalid room", step.Order)
		}
		room := rooms[step.RoomID]
		if room.Secret || !room.Mandatory {
			return invalidAdventure("progression step %d must reference a mandatory non-secret room", step.Order)
		}
		if _, duplicate := stepsByRoom[step.RoomID]; duplicate {
			return invalidAdventure("progression visits room %q more than once", step.RoomID)
		}
		stepsByRoom[step.RoomID] = index
		if err := validateLocalized(fmt.Sprintf("progression step %d beat", step.Order), step.Beat); err != nil {
			return err
		}
		for _, keyID := range step.RequiresKeyIDs {
			if !heldKeys[keyID] {
				return invalidAdventure("progression step %d requires key %q before it is acquired", step.Order, keyID)
			}
		}
		for _, keyID := range step.GrantsKeyIDs {
			entity, ok := entities[keyID]
			if !ok || entity.Kind != "key" || entity.RoomID != step.RoomID {
				return invalidAdventure("progression step %d grants invalid key %q", step.Order, keyID)
			}
			if heldKeys[keyID] {
				return invalidAdventure("progression grants key %q more than once", keyID)
			}
			heldKeys[keyID] = true
			grantedAt[keyID] = index
		}
	}

	first := progression.Steps[0]
	last := progression.Steps[len(progression.Steps)-1]
	if first.RoomID != progression.EntryRoomID || first.Kind != "entrance" || rooms[first.RoomID].Role != "entrance" {
		return invalidAdventure("progression must begin at the entrance room")
	}
	if last.RoomID != progression.ClimaxRoomID || last.Kind != "climax" || rooms[last.RoomID].Role != "boss" {
		return invalidAdventure("progression must end at the boss climax")
	}
	lastFloorIndex := -1
	for _, floor := range document.Floors {
		if floor.Index > lastFloorIndex {
			lastFloorIndex = floor.Index
		}
	}
	climaxFloorID := roomFloors[progression.ClimaxRoomID]
	for _, floor := range document.Floors {
		if floor.ID == climaxFloorID && floor.Index != lastFloorIndex {
			return invalidAdventure("climax must be on the final floor")
		}
	}
	if !hasBossInRoom(entities, progression.ClimaxRoomID) {
		return invalidAdventure("climax room must contain the boss entity")
	}

	if len(document.Analysis.CriticalPath) != len(progression.Steps) {
		return invalidAdventure("critical path must match progression steps")
	}
	for index, roomID := range document.Analysis.CriticalPath {
		if roomID != progression.Steps[index].RoomID {
			return invalidAdventure("critical path diverges from progression at step %d", index+1)
		}
	}

	locksByTarget := make(map[string]bool, len(progression.Locks))
	lockIDs := make(map[string]bool, len(progression.Locks))
	for _, lock := range progression.Locks {
		if lock.ID == "" || lockIDs[lock.ID] || lock.TargetID == "" || locksByTarget[lock.TargetID] {
			return invalidAdventure("progression lock id and target must be unique")
		}
		lockIDs[lock.ID] = true
		locksByTarget[lock.TargetID] = true
		if !oneOfValue(lock.Kind, "door", "portal") {
			return invalidAdventure("unsupported progression lock kind %q", lock.Kind)
		}
		fromIndex, fromOK := stepsByRoom[lock.FromRoomID]
		toIndex, toOK := stepsByRoom[lock.ToRoomID]
		if !fromOK || !toOK || toIndex != fromIndex+1 {
			return invalidAdventure("lock %q must protect consecutive progression steps", lock.ID)
		}
		if lock.FloorID != roomFloors[lock.FromRoomID] {
			return invalidAdventure("lock %q source floor mismatch", lock.ID)
		}
		grantIndex, granted := grantedAt[lock.KeyID]
		if !granted || grantIndex >= toIndex || !containsString(progression.Steps[toIndex].RequiresKeyIDs, lock.KeyID) {
			return invalidAdventure("lock %q is not solved by a previously acquired key", lock.ID)
		}
		switch lock.Kind {
		case "portal":
			target, ok := portals[lock.TargetID]
			if !ok || target.FromFloorID != lock.FloorID || target.ToFloorID != roomFloors[lock.ToRoomID] || target.RequiredKeyID != lock.KeyID {
				return invalidAdventure("lock %q references an invalid keyed portal", lock.ID)
			}
		case "door":
			target, ok := walls[lock.TargetID]
			if !ok || wallFloors[lock.TargetID] != lock.FloorID || target.Kind != "door" || !target.Locked || target.RequiredKeyID != lock.KeyID {
				return invalidAdventure("lock %q references an invalid keyed door", lock.ID)
			}
		}
	}
	for id, portal := range portals {
		if portal.RequiredKeyID != "" && !locksByTarget[id] {
			return invalidAdventure("keyed portal %q is missing from progression locks", id)
		}
	}
	for id, wall := range walls {
		if wall.RequiredKeyID != "" && !locksByTarget[id] {
			return invalidAdventure("keyed door %q is missing from progression locks", id)
		}
	}

	secretIDs := make(map[string]bool, len(progression.SecretRoomIDs))
	for _, roomID := range progression.SecretRoomIDs {
		room, ok := rooms[roomID]
		if !ok || !room.Secret || room.Mandatory || secretIDs[roomID] {
			return invalidAdventure("progression references invalid secret room %q", roomID)
		}
		if _, mandatory := stepsByRoom[roomID]; mandatory {
			return invalidAdventure("secret room %q is on the mandatory path", roomID)
		}
		if !hasPurposeInRoom(entities, roomID) {
			return invalidAdventure("secret dead end %q must contain a reward, clue, or purpose", roomID)
		}
		secretIDs[roomID] = true
	}
	for roomID, room := range rooms {
		if room.Secret && !secretIDs[roomID] {
			return invalidAdventure("secret room %q is missing from progression analysis", roomID)
		}
	}
	return nil
}

func hasBossInRoom(entities map[string]SceneEntity, roomID string) bool {
	for _, entity := range entities {
		if entity.Kind == "boss" && entity.RoomID == roomID {
			return true
		}
	}
	return false
}

func hasPurposeInRoom(entities map[string]SceneEntity, roomID string) bool {
	for _, entity := range entities {
		if entity.RoomID == roomID && oneOfValue(entity.Kind, "prop", "marker", "key", "trap") {
			return true
		}
	}
	return false
}

func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
