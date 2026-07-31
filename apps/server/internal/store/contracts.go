package store

import (
	"context"
	"time"

	"github.com/EmanuelSena101/DDivination/apps/server/internal/domain"
)

// Persistence is the durable boundary used by the API and application
// services. Implementations may use a managed PostgreSQL provider, but callers
// never depend on a provider SDK.
type Persistence interface {
	AdventurePersistence
	GenerationPersistence
	SessionPersistence
	AssetPersistence
}

type AdventurePersistence interface {
	SaveAdventure(context.Context, domain.AdventureDocument, *int64, string) error
	GetAdventure(context.Context, string) (domain.AdventureDocument, error)
	ListAdventures(context.Context, int) ([]AdventureSummary, error)
	CreateAdventureSnapshot(context.Context, string, string) (domain.AdventureSnapshotSummary, error)
	ListAdventureSnapshots(context.Context, string, int) ([]domain.AdventureSnapshotSummary, error)
	GetAdventureSnapshot(context.Context, string, string) (domain.AdventureSnapshot, error)
	DeleteAdventure(context.Context, string) error
}

type GenerationPersistence interface {
	SaveGenerationRun(context.Context, domain.GenerationRun) error
	GetGenerationRun(context.Context, string) (domain.GenerationRun, error)
	ListGenerationRuns(context.Context, int) ([]domain.GenerationRun, error)
}

type AssetPersistence interface {
	SaveAsset(context.Context, domain.AssetRef, string) error
	ListAssets(context.Context) ([]domain.AssetRef, error)
}

type SessionPersistence interface {
	InitializeSession(context.Context, SessionInitialization) error
	CommitSession(context.Context, SessionCommit) (SessionCommitResult, error)
	LoadSession(context.Context, string) (SessionRecord, error)
	LoadSessionEventByCommand(context.Context, string, string) (domain.SessionEvent, error)
	LoadSessionReplay(context.Context, string, int64, int) (SessionReplay, error)
	CompactSessionEvents(context.Context, string, int64) (int64, error)
	SaveSessionCredential(context.Context, string, string, string) error
	DeleteSessionCredential(context.Context, string, string) error
	LoadSessionCredentials(context.Context, string) (map[string]string, error)
}

type SessionInitialization struct {
	State                   domain.SessionState
	JoinCodeHash            string
	JoinCodeExpiresAt       time.Time
	CredentialParticipantID string
	CredentialToken         string
}

type SessionAccessUpdate struct {
	CodeHash  string
	ExpiresAt time.Time
}

type SessionCommit struct {
	SessionID        string
	CommandID        string
	ExpectedRevision int64
	Event            domain.SessionEvent
	State            domain.SessionState
	ForceSnapshot    bool
	Access           *SessionAccessUpdate
}

type SessionCommitResult struct {
	Event     domain.SessionEvent
	Duplicate bool
}

type SessionRecord struct {
	State             domain.SessionState
	JoinCodeHash      string
	JoinCodeExpiresAt time.Time
}

type SessionReplay struct {
	Events          []domain.SessionEvent
	CurrentRevision int64
	OldestRevision  int64
	Complete        bool
}
