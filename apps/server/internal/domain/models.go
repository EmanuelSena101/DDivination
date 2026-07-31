package domain

import "time"

const (
	SchemaVersion    = "1.0.0"
	GeneratorVersion = "go-v1-alpha.2"
)

type LocalizedText struct {
	PTBR string `json:"pt-BR" minLength:"1"`
	ENUS string `json:"en-US" minLength:"1"`
}

type AdventureSpec struct {
	Name            *LocalizedText `json:"name,omitempty"`
	PartySize       int            `json:"partySize" minimum:"1" maximum:"8"`
	PartyLevel      int            `json:"partyLevel" minimum:"1" maximum:"20"`
	DurationHours   int            `json:"durationHours" minimum:"1" maximum:"12"`
	Difficulty      string         `json:"difficulty" enum:"easy,medium,hard,deadly"`
	Theme           string         `json:"theme" minLength:"1" maxLength:"40"`
	Biome           string         `json:"biome" minLength:"1" maxLength:"40"`
	FloorCount      int            `json:"floorCount" minimum:"1" maximum:"5"`
	Objective       string         `json:"objective" minLength:"1" maxLength:"120"`
	Antagonist      string         `json:"antagonist" minLength:"1" maxLength:"120"`
	StructureStyle  string         `json:"structureStyle" enum:"linear,branching,labyrinthine"`
	TreasureQuality string         `json:"treasureQuality" enum:"poor,standard,rich,legendary"`
	UseAI           bool           `json:"useAI"`
}

func DefaultAdventureSpec() AdventureSpec {
	return AdventureSpec{
		PartySize:       4,
		PartyLevel:      5,
		DurationHours:   4,
		Difficulty:      "medium",
		Theme:           "forgotten temple",
		Biome:           "underground",
		FloorCount:      2,
		Objective:       "stop the awakening ritual",
		Antagonist:      "serpent cult",
		StructureStyle:  "branching",
		TreasureQuality: "standard",
	}
}

type GridPosition struct {
	X int `json:"x"`
	Z int `json:"z"`
}

type Tile struct {
	X        int    `json:"x"`
	Z        int    `json:"z"`
	Kind     string `json:"kind" enum:"floor,corridor,stairs,water,lava"`
	RoomID   string `json:"roomId,omitempty"`
	Walkable bool   `json:"walkable"`
}

type WallEdge struct {
	ID            string `json:"id,omitempty"`
	X             int    `json:"x"`
	Z             int    `json:"z"`
	Direction     string `json:"direction" enum:"north,east,south,west"`
	Kind          string `json:"kind" enum:"wall,door,secret-door"`
	Open          bool   `json:"open"`
	Locked        bool   `json:"locked"`
	RequiredKeyID string `json:"requiredKeyId,omitempty"`
}

type Portal struct {
	ID            string       `json:"id"`
	FromFloorID   string       `json:"fromFloorId"`
	From          GridPosition `json:"from"`
	ToFloorID     string       `json:"toFloorId"`
	To            GridPosition `json:"to"`
	Kind          string       `json:"kind" enum:"stairs-up,stairs-down,portal"`
	RequiredKeyID string       `json:"requiredKeyId,omitempty"`
}

type Room struct {
	ID          string        `json:"id"`
	Name        LocalizedText `json:"name"`
	Role        string        `json:"role"`
	Description LocalizedText `json:"description"`
	Center      GridPosition  `json:"center"`
	Secret      bool          `json:"secret"`
	Mandatory   bool          `json:"mandatory"`
}

type SceneEntity struct {
	ID             string        `json:"id"`
	Kind           string        `json:"kind" enum:"prop,light,trap,token,marker,key,boss"`
	Name           LocalizedText `json:"name"`
	Position       GridPosition  `json:"position"`
	AssetID        string        `json:"assetId,omitempty"`
	BlocksMovement bool          `json:"blocksMovement"`
	Hidden         bool          `json:"hidden"`
	RoomID         string        `json:"roomId,omitempty"`
	OwnerID        string        `json:"ownerId,omitempty"`
}

type FloorMap struct {
	ID       string        `json:"id"`
	Index    int           `json:"index"`
	Name     LocalizedText `json:"name"`
	Width    int           `json:"width"`
	Height   int           `json:"height"`
	Tiles    []Tile        `json:"tiles"`
	Walls    []WallEdge    `json:"walls"`
	Rooms    []Room        `json:"rooms"`
	Entities []SceneEntity `json:"entities"`
	Portals  []Portal      `json:"portals"`
}

type EncounterCreature struct {
	Index string        `json:"index"`
	Name  LocalizedText `json:"name"`
	Count int           `json:"count"`
	CR    float64       `json:"cr"`
	XP    int           `json:"xp"`
}

type Encounter struct {
	ID         string              `json:"id"`
	FloorID    string              `json:"floorId"`
	RoomID     string              `json:"roomId"`
	Difficulty string              `json:"difficulty"`
	Creatures  []EncounterCreature `json:"creatures"`
	TotalXP    int                 `json:"totalXp"`
}

type DungeonAnalysis struct {
	TotalRooms          int      `json:"totalRooms"`
	TotalFloors         int      `json:"totalFloors"`
	CriticalPath        []string `json:"criticalPath"`
	DeadEnds            []string `json:"deadEnds"`
	EstimatedDifficulty string   `json:"estimatedDifficulty"`
	Invariants          []string `json:"invariants"`
}

type ProgressionStep struct {
	Order          int           `json:"order" minimum:"1"`
	FloorID        string        `json:"floorId"`
	RoomID         string        `json:"roomId"`
	Kind           string        `json:"kind" enum:"entrance,exploration,key,transition,climax"`
	Beat           LocalizedText `json:"beat"`
	GrantsKeyIDs   []string      `json:"grantsKeyIds"`
	RequiresKeyIDs []string      `json:"requiresKeyIds"`
}

type ProgressionLock struct {
	ID         string `json:"id"`
	Kind       string `json:"kind" enum:"door,portal"`
	TargetID   string `json:"targetId"`
	FloorID    string `json:"floorId"`
	FromRoomID string `json:"fromRoomId"`
	ToRoomID   string `json:"toRoomId"`
	KeyID      string `json:"keyId"`
}

type DungeonProgression struct {
	EntryRoomID     string            `json:"entryRoomId"`
	ObjectiveRoomID string            `json:"objectiveRoomId"`
	ClimaxRoomID    string            `json:"climaxRoomId"`
	Steps           []ProgressionStep `json:"steps"`
	Locks           []ProgressionLock `json:"locks"`
	SecretRoomIDs   []string          `json:"secretRoomIds"`
	Solvable        bool              `json:"solvable"`
}

type AdventureNarrative struct {
	Hook       LocalizedText `json:"hook"`
	Objective  LocalizedText `json:"objective"`
	Antagonist LocalizedText `json:"antagonist"`
	Atmosphere LocalizedText `json:"atmosphere"`
}

type AdventureDocument struct {
	ID               string             `json:"id"`
	SchemaVersion    string             `json:"schemaVersion"`
	GeneratorVersion string             `json:"generatorVersion"`
	Version          int64              `json:"version"`
	Seed             uint64             `json:"seed"`
	Name             LocalizedText      `json:"name"`
	Spec             AdventureSpec      `json:"spec"`
	Summary          LocalizedText      `json:"summary"`
	Narrative        AdventureNarrative `json:"narrative"`
	Floors           []FloorMap         `json:"floors"`
	Encounters       []Encounter        `json:"encounters"`
	Progression      DungeonProgression `json:"progression"`
	Analysis         DungeonAnalysis    `json:"analysis"`
	Attributions     []Attribution      `json:"attributions"`
	CreatedAt        time.Time          `json:"createdAt"`
	UpdatedAt        time.Time          `json:"updatedAt"`
}

type AdventureSnapshot struct {
	ID          string            `json:"id"`
	AdventureID string            `json:"adventureId"`
	Version     int64             `json:"version"`
	Reason      string            `json:"reason"`
	Name        LocalizedText     `json:"name"`
	CreatedAt   time.Time         `json:"createdAt"`
	Document    AdventureDocument `json:"document"`
}

type AdventureSnapshotSummary struct {
	ID          string        `json:"id"`
	AdventureID string        `json:"adventureId"`
	Version     int64         `json:"version"`
	Reason      string        `json:"reason"`
	Name        LocalizedText `json:"name"`
	CreatedAt   time.Time     `json:"createdAt"`
}

type Attribution struct {
	Title   string `json:"title"`
	Creator string `json:"creator"`
	Source  string `json:"source"`
	License string `json:"license"`
	Notice  string `json:"notice"`
}

type GenerationStage struct {
	Name       string    `json:"name"`
	Progress   int       `json:"progress" minimum:"0" maximum:"100"`
	OccurredAt time.Time `json:"occurredAt"`
}

type GenerationRun struct {
	ID               string            `json:"id"`
	Status           string            `json:"status" enum:"queued,running,completed,failed,cancelled"`
	Stage            string            `json:"stage"`
	Progress         int               `json:"progress" minimum:"0" maximum:"100"`
	Seed             uint64            `json:"seed"`
	GeneratorVersion string            `json:"generatorVersion"`
	Spec             AdventureSpec     `json:"spec"`
	AdventureID      string            `json:"adventureId,omitempty"`
	Diagnostics      []string          `json:"diagnostics"`
	Stages           []GenerationStage `json:"stages"`
	CreatedAt        time.Time         `json:"createdAt"`
	UpdatedAt        time.Time         `json:"updatedAt"`
	CompletedAt      *time.Time        `json:"completedAt,omitempty"`
}

type AssetRef struct {
	ID        string      `json:"id"`
	SHA256    string      `json:"sha256"`
	FileName  string      `json:"fileName"`
	MediaType string      `json:"mediaType"`
	Size      int64       `json:"size"`
	Kind      string      `json:"kind"`
	License   Attribution `json:"license"`
}

type Participant struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Role     string    `json:"role" enum:"gm,player,display"`
	Token    string    `json:"-"`
	JoinedAt time.Time `json:"joinedAt"`
}

type InitiativeEntry struct {
	TokenID string `json:"tokenId"`
	Name    string `json:"name"`
	Score   int    `json:"score"`
}

type InitiativeState struct {
	Entries     []InitiativeEntry `json:"entries"`
	ActiveIndex int               `json:"activeIndex"`
	Round       int               `json:"round"`
}

type DiceRoll struct {
	ID         string    `json:"id"`
	ActorID    string    `json:"actorId"`
	Expression string    `json:"expression"`
	Values     []int     `json:"values"`
	Modifier   int       `json:"modifier"`
	Total      int       `json:"total"`
	Visibility string    `json:"visibility" enum:"public,gm,private"`
	TargetID   string    `json:"targetId,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}

type SessionState struct {
	ID             string                    `json:"id"`
	AdventureID    string                    `json:"adventureId"`
	Revision       int64                     `json:"revision"`
	ActiveFloorID  string                    `json:"activeFloorId"`
	Participants   map[string]Participant    `json:"participants"`
	TokenPositions map[string]GridPosition   `json:"tokenPositions"`
	TokenFloors    map[string]string         `json:"tokenFloors"`
	TokenOwners    map[string]string         `json:"tokenOwners"`
	RevealedCells  map[string][]GridPosition `json:"revealedCells"`
	Initiative     InitiativeState           `json:"initiative"`
	Rolls          []DiceRoll                `json:"rolls"`
	Open           bool                      `json:"open"`
	CreatedAt      time.Time                 `json:"createdAt"`
}

type SessionCommand struct {
	ID               string         `json:"id"`
	ExpectedRevision int64          `json:"expectedRevision"`
	Type             string         `json:"type"`
	Payload          map[string]any `json:"payload"`
}

type SessionEvent struct {
	Revision   int64          `json:"revision"`
	Type       string         `json:"type"`
	ActorID    string         `json:"actorId"`
	OccurredAt time.Time      `json:"occurredAt"`
	Payload    map[string]any `json:"payload"`
}
