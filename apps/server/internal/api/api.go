package api

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/EmanuelSena101/DDivination/apps/server/internal/ai"
	"github.com/EmanuelSena101/DDivination/apps/server/internal/asset"
	"github.com/EmanuelSena101/DDivination/apps/server/internal/domain"
	"github.com/EmanuelSena101/DDivination/apps/server/internal/exporter"
	generationjobs "github.com/EmanuelSena101/DDivination/apps/server/internal/generation"
	"github.com/EmanuelSena101/DDivination/apps/server/internal/generator"
	"github.com/EmanuelSena101/DDivination/apps/server/internal/packageio"
	"github.com/EmanuelSena101/DDivination/apps/server/internal/session"
	"github.com/EmanuelSena101/DDivination/apps/server/internal/store"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
)

type LANEnabler interface {
	Enable(http.Handler) ([]string, error)
}

type Server struct {
	store      store.Persistence
	hub        *session.Hub
	lan        LANEnabler
	logger     *slog.Logger
	webDir     string
	webFS      fs.FS
	assets     *asset.Manager
	generation *generationjobs.Manager
}

type Handlers struct {
	Local http.Handler
	LAN   http.Handler
}

func New(s store.Persistence, hub *session.Hub, lan LANEnabler, logger *slog.Logger, webDir string, webFS fs.FS, assetsDir string) *Server {
	server := &Server{
		store: s, hub: hub, lan: lan, logger: logger, webDir: webDir, webFS: webFS,
		assets:     asset.NewManager(assetsDir, s),
		generation: generationjobs.NewManager(s, logger),
	}
	if err := server.generation.Recover(context.Background()); err != nil {
		logger.Error("could not recover interrupted generation runs", "error", err)
	}
	return server
}

// NewContractServer registers the complete HTTP contract without requiring a
// live database. It is used only to render OpenAPI; handlers are not executed.
func NewContractServer(logger *slog.Logger) *Server {
	return &Server{logger: logger}
}

func (s *Server) Handlers() Handlers {
	localMux := http.NewServeMux()
	config := huma.DefaultConfig("DDivination API", "1.0.0-alpha.1")
	config.Info.Description = "Online-first 3D VTT and deterministic 5E-compatible adventure generator."
	localAPI := humago.New(localMux, config)
	s.registerREST(localAPI)
	localMux.HandleFunc("GET /api/v1/generation-runs/{id}/stream", s.streamGeneration)
	localMux.HandleFunc("GET /api/v1/sessions/{id}/stream", s.streamSession)
	localMux.Handle("/", s.staticHandler())

	lanMux := http.NewServeMux()
	lanMux.HandleFunc("GET /api/v1/health", s.rawHealth)
	lanMux.HandleFunc("POST /api/v1/sessions/{id}/join", s.rawJoinSession)
	lanMux.HandleFunc("GET /api/v1/sessions/{id}/join/{requestId}", s.rawJoinStatus)
	lanMux.HandleFunc("GET /api/v1/sessions/{id}/stream", s.streamSession)
	lanMux.Handle("/", s.staticHandler())
	return Handlers{Local: securityHeaders(localMux), LAN: securityHeaders(lanMux)}
}

type healthBody struct {
	Status           string `json:"status"`
	Version          string `json:"version"`
	SchemaVersion    string `json:"schemaVersion"`
	GeneratorVersion string `json:"generatorVersion"`
}

type healthOutput struct {
	Body healthBody
}

func (s *Server) getHealth(_ context.Context, _ *struct{}) (*healthOutput, error) {
	return &healthOutput{Body: healthBody{
		Status:           "ok",
		Version:          "1.0.0-alpha.1",
		SchemaVersion:    domain.SchemaVersion,
		GeneratorVersion: domain.GeneratorVersion,
	}}, nil
}

type catalogOutput struct {
	Body domain.Catalog
}

func (s *Server) getCatalog(_ context.Context, _ *struct{}) (*catalogOutput, error) {
	return &catalogOutput{Body: starterCatalog()}, nil
}

type generationInput struct {
	OpenAIKey string `header:"X-OpenAI-API-Key"`
	Body      struct {
		Spec domain.AdventureSpec `json:"spec"`
		Seed *uint64              `json:"seed,omitempty" maximum:"9007199254740991"`
	}
}

type generationOutput struct {
	Body domain.GenerationRun
}

func (s *Server) createGeneration(ctx context.Context, input *generationInput) (*generationOutput, error) {
	seed := randomSeed()
	if input.Body.Seed != nil {
		seed = *input.Body.Seed
	}
	runID, err := randomID(9)
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, "could not create generation run")
	}
	now := time.Now().UTC()
	run := domain.GenerationRun{
		ID:               runID,
		Seed:             seed,
		GeneratorVersion: domain.GeneratorVersion,
		Spec:             input.Body.Spec,
		Diagnostics:      []string{"procedural-mode", "offline-ready"},
	}
	queued, err := s.generation.Enqueue(
		ctx,
		run,
		s.generationExecutor(run, input.OpenAIKey, now),
	)
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, "could not persist generation run")
	}
	return &generationOutput{Body: queued}, nil
}

func (s *Server) generationExecutor(
	run domain.GenerationRun,
	openAIKey string,
	startedAt time.Time,
) generationjobs.Executor {
	return func(ctx context.Context, report generationjobs.Reporter) (string, error) {
		if err := report("validating-spec", 5); err != nil {
			return "", err
		}
		var reportErr error
		doc, err := generator.GenerateContext(
			ctx,
			run.Spec,
			run.Seed,
			startedAt,
			func(completed, total int) {
				if reportErr != nil {
					return
				}
				progress := 10 + completed*55/total
				reportErr = report(
					fmt.Sprintf("building-floor-%d-of-%d", completed, total),
					progress,
				)
			},
		)
		if err != nil {
			return "", err
		}
		if reportErr != nil {
			return "", reportErr
		}
		if err := report("procedural-complete", 72); err != nil {
			return "", err
		}
		if run.Spec.UseAI {
			if err := report("enriching-narrative", 78); err != nil {
				return "", err
			}
			if strings.TrimSpace(openAIKey) == "" {
				if err := report(
					"procedural-fallback",
					84,
					"ai-skipped:key-missing",
					"procedural-fallback-used",
				); err != nil {
					return "", err
				}
			} else {
				result, enrichErr := ai.NewOpenAI(openAIKey, "").Enrich(ctx, run.Spec)
				if enrichErr != nil {
					if err := report(
						"procedural-fallback",
						84,
						"ai-failed:"+enrichErr.Error(),
						"procedural-fallback-used",
					); err != nil {
						return "", err
					}
				} else {
					doc.Narrative = result.Enrichment
					if err := report(
						"ai-complete",
						84,
						fmt.Sprintf("ai-provider:%s", result.Diagnostics.Provider),
						fmt.Sprintf("ai-model:%s", result.Diagnostics.Model),
						fmt.Sprintf("ai-input-tokens:%d", result.Diagnostics.InputTokens),
						fmt.Sprintf("ai-output-tokens:%d", result.Diagnostics.OutputTokens),
						fmt.Sprintf("ai-latency-ms:%d", result.Diagnostics.Latency.Milliseconds()),
					); err != nil {
						return "", err
					}
				}
			}
		}
		if err := report("validating-document", 90); err != nil {
			return "", err
		}
		if err := domain.ValidateAdventure(doc); err != nil {
			return "", err
		}
		if err := ctx.Err(); err != nil {
			return "", err
		}
		if err := report("persisting-adventure", 96); err != nil {
			return "", err
		}
		if err := ctx.Err(); err != nil {
			return "", err
		}
		if err := s.store.SaveAdventure(ctx, doc, nil, "generated"); err != nil {
			return "", err
		}
		return doc.ID, nil
	}
}

type idInput struct {
	ID string `path:"id"`
}

type generationRunOutput struct {
	Body domain.GenerationRun
}

func (s *Server) getGeneration(ctx context.Context, input *idInput) (*generationRunOutput, error) {
	run, err := s.store.GetGenerationRun(ctx, input.ID)
	if errors.Is(err, store.ErrNotFound) {
		return nil, huma.NewError(http.StatusNotFound, "generation run not found")
	}
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, "could not load generation run")
	}
	return &generationRunOutput{Body: run}, nil
}

type listGenerationRunsInput struct {
	Limit int `query:"limit" minimum:"1" maximum:"200" default:"50"`
}

type listGenerationRunsOutput struct {
	Body []domain.GenerationRun
}

func (s *Server) listGenerationRuns(
	ctx context.Context,
	input *listGenerationRunsInput,
) (*listGenerationRunsOutput, error) {
	runs, err := s.store.ListGenerationRuns(ctx, input.Limit)
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, "could not list generation runs")
	}
	return &listGenerationRunsOutput{Body: runs}, nil
}

func (s *Server) cancelGeneration(
	ctx context.Context,
	input *idInput,
) (*generationRunOutput, error) {
	run, err := s.generation.Cancel(ctx, input.ID)
	if errors.Is(err, store.ErrNotFound) {
		return nil, huma.NewError(http.StatusNotFound, "generation run not found")
	}
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, "could not cancel generation run")
	}
	return &generationRunOutput{Body: run}, nil
}

func (s *Server) streamGeneration(w http.ResponseWriter, r *http.Request) {
	runID := r.PathValue("id")
	updates, unsubscribe, err := s.generation.Subscribe(r.Context(), runID)
	if errors.Is(err, store.ErrNotFound) {
		writeProblem(w, http.StatusNotFound, "Generation run not found", "")
		return
	}
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Could not subscribe to generation run", err.Error())
		return
	}
	defer unsubscribe()

	connection, err := websocket.Accept(
		w,
		r,
		&websocket.AcceptOptions{OriginPatterns: allowedOrigins(r)},
	)
	if err != nil {
		return
	}
	defer connection.Close(websocket.StatusNormalClosure, "generation run closed")
	for {
		select {
		case <-r.Context().Done():
			return
		case run := <-updates:
			if err := wsjson.Write(r.Context(), connection, run); err != nil {
				return
			}
			if run.Status == "completed" || run.Status == "failed" || run.Status == "cancelled" {
				return
			}
		}
	}
}

type listAdventuresInput struct {
	Limit int `query:"limit" minimum:"1" maximum:"200" default:"50"`
}

type listAdventuresOutput struct {
	Body []store.AdventureSummary
}

func (s *Server) listAdventures(ctx context.Context, input *listAdventuresInput) (*listAdventuresOutput, error) {
	items, err := s.store.ListAdventures(ctx, input.Limit)
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, "could not list adventures")
	}
	return &listAdventuresOutput{Body: items}, nil
}

type adventureOutput struct {
	ETag string `header:"ETag"`
	Body domain.AdventureDocument
}

func (s *Server) getAdventure(ctx context.Context, input *idInput) (*adventureOutput, error) {
	doc, err := s.store.GetAdventure(ctx, input.ID)
	if errors.Is(err, store.ErrNotFound) {
		return nil, huma.NewError(http.StatusNotFound, "adventure not found")
	}
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, "could not load adventure")
	}
	return &adventureOutput{ETag: fmt.Sprintf(`"%d"`, doc.Version), Body: doc}, nil
}

type updateAdventureInput struct {
	ID      string `path:"id"`
	IfMatch string `header:"If-Match" required:"true"`
	Body    domain.AdventureDocument
}

func (s *Server) updateAdventure(ctx context.Context, input *updateAdventureInput) (*adventureOutput, error) {
	expected, err := etagVersion(input.IfMatch)
	if err != nil || expected == nil {
		return nil, huma.NewError(http.StatusBadRequest, "invalid If-Match version")
	}
	if input.Body.ID != "" && input.Body.ID != input.ID {
		return nil, huma.NewError(http.StatusBadRequest, "document ID does not match URL")
	}
	input.Body.ID = input.ID
	input.Body.Version = *expected + 1
	input.Body.UpdatedAt = time.Now().UTC()
	if err := s.store.SaveAdventure(ctx, input.Body, expected, "edited"); errors.Is(err, store.ErrConflict) {
		return nil, huma.NewError(http.StatusConflict, "adventure was changed by another editor")
	} else if errors.Is(err, store.ErrNotFound) {
		return nil, huma.NewError(http.StatusNotFound, "adventure not found")
	} else if errors.Is(err, domain.ErrInvalidAdventure) {
		return nil, huma.NewError(http.StatusUnprocessableEntity, err.Error())
	} else if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, "could not update adventure")
	}
	return &adventureOutput{ETag: fmt.Sprintf(`"%d"`, input.Body.Version), Body: input.Body}, nil
}

type listAdventureCheckpointsInput struct {
	ID    string `path:"id"`
	Limit int    `query:"limit" minimum:"1" maximum:"200" default:"50"`
}

type adventureCheckpointsOutput struct {
	Body []domain.AdventureSnapshotSummary
}

func (s *Server) listAdventureCheckpoints(ctx context.Context, input *listAdventureCheckpointsInput) (*adventureCheckpointsOutput, error) {
	snapshots, err := s.store.ListAdventureSnapshots(ctx, input.ID, input.Limit)
	if errors.Is(err, store.ErrNotFound) {
		return nil, huma.NewError(http.StatusNotFound, "adventure not found")
	}
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, "could not list checkpoints")
	}
	return &adventureCheckpointsOutput{Body: snapshots}, nil
}

type adventureCheckpointOutput struct {
	Body domain.AdventureSnapshotSummary
}

func (s *Server) checkpointAdventure(ctx context.Context, input *idInput) (*adventureCheckpointOutput, error) {
	snapshot, err := s.store.CreateAdventureSnapshot(ctx, input.ID, "manual-checkpoint")
	if errors.Is(err, store.ErrNotFound) {
		return nil, huma.NewError(http.StatusNotFound, "adventure not found")
	}
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, "could not create checkpoint")
	}
	return &adventureCheckpointOutput{Body: snapshot}, nil
}

type restoreAdventureCheckpointInput struct {
	ID           string `path:"id"`
	CheckpointID string `path:"checkpointId"`
	IfMatch      string `header:"If-Match" required:"true"`
}

func (s *Server) restoreAdventureCheckpoint(ctx context.Context, input *restoreAdventureCheckpointInput) (*adventureOutput, error) {
	expected, err := etagVersion(input.IfMatch)
	if err != nil || expected == nil {
		return nil, huma.NewError(http.StatusBadRequest, "invalid If-Match version")
	}
	snapshot, err := s.store.GetAdventureSnapshot(ctx, input.ID, input.CheckpointID)
	if errors.Is(err, store.ErrNotFound) {
		return nil, huma.NewError(http.StatusNotFound, "checkpoint not found")
	}
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, "could not load checkpoint")
	}
	document := snapshot.Document
	document.ID = input.ID
	document.Version = *expected + 1
	document.UpdatedAt = time.Now().UTC()
	if err := s.store.SaveAdventure(ctx, document, expected, "checkpoint-restored"); errors.Is(err, store.ErrConflict) {
		return nil, huma.NewError(http.StatusConflict, "adventure was changed by another editor")
	} else if errors.Is(err, store.ErrNotFound) {
		return nil, huma.NewError(http.StatusNotFound, "adventure not found")
	} else if errors.Is(err, domain.ErrInvalidAdventure) {
		return nil, huma.NewError(http.StatusUnprocessableEntity, err.Error())
	} else if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, "could not restore checkpoint")
	}
	return &adventureOutput{ETag: fmt.Sprintf(`"%d"`, document.Version), Body: document}, nil
}

type emptyOutput struct{}

func (s *Server) deleteAdventure(ctx context.Context, input *idInput) (*emptyOutput, error) {
	if err := s.store.DeleteAdventure(ctx, input.ID); errors.Is(err, store.ErrNotFound) {
		return nil, huma.NewError(http.StatusNotFound, "adventure not found")
	} else if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, "could not delete adventure")
	}
	return &emptyOutput{}, nil
}

type createSessionInput struct {
	Body struct {
		AdventureID string `json:"adventureId" minLength:"1"`
		GMName      string `json:"gmName" maxLength:"60"`
	}
}

type createSessionResponse struct {
	Session  session.Created `json:"session"`
	JoinURLs []string        `json:"joinUrls"`
}

type createSessionOutput struct {
	Body createSessionResponse
}

func (s *Server) createSession(ctx context.Context, input *createSessionInput) (*createSessionOutput, error) {
	doc, err := s.store.GetAdventure(ctx, input.Body.AdventureID)
	if errors.Is(err, store.ErrNotFound) {
		return nil, huma.NewError(http.StatusNotFound, "adventure not found")
	}
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, "could not load adventure")
	}
	created, err := s.hub.Create(ctx, doc, input.Body.GMName)
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, "could not create session")
	}
	urls := []string{"http://127.0.0.1:8080"}
	if s.lan != nil {
		handlers := s.Handlers()
		if enabled, enableErr := s.lan.Enable(handlers.LAN); enableErr == nil && len(enabled) > 0 {
			urls = enabled
		} else if enableErr != nil {
			s.logger.Warn("LAN listener unavailable", "error", enableErr)
		}
	}
	for i := range urls {
		urls[i] = fmt.Sprintf("%s/?session=%s&code=%s", strings.TrimSuffix(urls[i], "/"), created.SessionID, created.Code)
	}
	return &createSessionOutput{Body: createSessionResponse{Session: created, JoinURLs: urls}}, nil
}

type joinRequest struct {
	Code string `json:"code" minLength:"1"`
	Name string `json:"name" maxLength:"60"`
	Role string `json:"role" enum:"player,display"`
}

type joinSessionInput struct {
	ID   string `path:"id"`
	Body joinRequest
}

type joinSessionOutput struct {
	Body session.Joined
}

func (s *Server) joinSession(ctx context.Context, input *joinSessionInput) (*joinSessionOutput, error) {
	joined, err := s.hub.Join(ctx, input.ID, input.Body.Code, input.Body.Name, input.Body.Role)
	if errors.Is(err, session.ErrSessionNotFound) {
		return nil, huma.NewError(http.StatusNotFound, "session not found")
	}
	if errors.Is(err, session.ErrSessionCorrupt) {
		return nil, huma.NewError(http.StatusInternalServerError, "session recovery failed", err)
	}
	if err != nil {
		return nil, huma.NewError(http.StatusUnauthorized, "could not join session", err)
	}
	return &joinSessionOutput{Body: joined}, nil
}

func (s *Server) rawJoinSession(w http.ResponseWriter, r *http.Request) {
	var request joinRequest
	if err := decodeJSON(w, r, &request, 1<<20); err != nil {
		writeProblem(w, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}
	output, err := s.joinSession(r.Context(), &joinSessionInput{ID: r.PathValue("id"), Body: request})
	if err != nil {
		status := http.StatusUnauthorized
		if statusError, ok := err.(huma.StatusError); ok {
			status = statusError.GetStatus()
		} else if errors.Is(err, session.ErrSessionNotFound) {
			status = http.StatusNotFound
		}
		writeProblem(w, status, "Could not join session", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, output.Body)
}

type joinStatusInput struct {
	ID            string `path:"id"`
	RequestID     string `path:"requestId"`
	Authorization string `header:"Authorization"`
}

type joinStatusOutput struct{ Body session.Joined }

func (s *Server) joinStatus(_ context.Context, input *joinStatusInput) (*joinStatusOutput, error) {
	token := bearerToken(input.Authorization)
	if token == "" {
		return nil, huma.NewError(http.StatusUnauthorized, "missing admission token")
	}
	joined, err := s.hub.AdmissionStatus(input.ID, input.RequestID, token)
	if errors.Is(err, session.ErrSessionNotFound) {
		return nil, huma.NewError(http.StatusNotFound, "session not found")
	}
	if errors.Is(err, session.ErrSessionCorrupt) {
		return nil, huma.NewError(http.StatusInternalServerError, "session recovery failed", err)
	}
	if err != nil {
		return nil, huma.NewError(http.StatusForbidden, "could not inspect admission", err)
	}
	return &joinStatusOutput{Body: joined}, nil
}

func (s *Server) rawJoinStatus(w http.ResponseWriter, r *http.Request) {
	output, err := s.joinStatus(r.Context(), &joinStatusInput{ID: r.PathValue("id"), RequestID: r.PathValue("requestId"), Authorization: r.Header.Get("Authorization")})
	if err != nil {
		status := http.StatusForbidden
		if statusError, ok := err.(huma.StatusError); ok {
			status = statusError.GetStatus()
		}
		writeProblem(w, status, "Could not inspect admission", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, output.Body)
}

func (s *Server) streamSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	token := r.URL.Query().Get("token")
	var lastRevision *int64
	if raw := strings.TrimSpace(r.URL.Query().Get("lastRevision")); raw != "" {
		parsed, parseErr := strconv.ParseInt(raw, 10, 64)
		if parseErr != nil || parsed < 0 {
			writeProblem(w, http.StatusBadRequest, "Invalid replay revision", "lastRevision must be a non-negative integer")
			return
		}
		lastRevision = &parsed
	}
	sub, initialMessages, err := s.hub.SubscribeFrom(sessionID, token, lastRevision)
	if err != nil {
		if errors.Is(err, session.ErrSessionCorrupt) {
			writeProblem(w, http.StatusInternalServerError, "Session recovery failed", err.Error())
		} else {
			writeProblem(w, http.StatusUnauthorized, "WebSocket authorization failed", err.Error())
		}
		return
	}
	connection, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: allowedOrigins(r)})
	if err != nil {
		s.hub.Unsubscribe(sessionID, sub)
		return
	}
	defer connection.Close(websocket.StatusNormalClosure, "session closed")
	defer s.hub.Unsubscribe(sessionID, sub)
	connection.SetReadLimit(1 << 20)

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	heartbeat := time.NewTicker(20 * time.Second)
	defer heartbeat.Stop()
	outbound := make(chan any, 128)
	readErrors := make(chan error, 1)
	for _, message := range initialMessages {
		outbound <- message
	}

	go func() {
		for {
			var command domain.SessionCommand
			if err := wsjson.Read(ctx, connection, &command); err != nil {
				readErrors <- err
				return
			}
			if _, err := s.hub.HandleCommand(ctx, sessionID, token, command); err != nil {
				outbound <- map[string]any{
					"type":    "command.rejected",
					"command": command.ID,
					"detail":  err.Error(),
				}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-readErrors:
			return
		case <-sub.Revoked:
			_ = connection.Close(websocket.StatusPolicyViolation, "session access revoked")
			return
		case <-heartbeat.C:
			pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
			err := connection.Ping(pingCtx)
			pingCancel()
			if err != nil {
				return
			}
		case event := <-sub.Events:
			if err := wsjson.Write(ctx, connection, event); err != nil {
				return
			}
		case message := <-outbound:
			if err := wsjson.Write(ctx, connection, message); err != nil {
				return
			}
		}
	}
}

type exportPackageOutput struct {
	ContentType        string `header:"Content-Type"`
	ContentDisposition string `header:"Content-Disposition"`
	Body               []byte
}

func (s *Server) exportPackage(ctx context.Context, input *idInput) (*exportPackageOutput, error) {
	doc, err := s.store.GetAdventure(ctx, input.ID)
	if errors.Is(err, store.ErrNotFound) {
		return nil, huma.NewError(http.StatusNotFound, "adventure not found")
	}
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, "could not export adventure")
	}
	content, err := packageio.Export(doc)
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, "could not export adventure", err)
	}
	return &exportPackageOutput{
		ContentType:        "application/vnd.ddivination+zip",
		ContentDisposition: fmt.Sprintf(`attachment; filename="%s.ddivination"`, safeName(doc.Name.ENUS)),
		Body:               content,
	}, nil
}

type importPackageInput struct {
	RawBody []byte `contentType:"application/vnd.ddivination+zip"`
}

type importPackageOutput struct {
	Body domain.AdventureDocument
}

func (s *Server) importPackage(ctx context.Context, input *importPackageInput) (*importPackageOutput, error) {
	doc, err := packageio.Import(bytes.NewReader(input.RawBody), int64(len(input.RawBody)))
	if err != nil {
		return nil, huma.NewError(http.StatusUnprocessableEntity, "invalid package", err)
	}
	if err := s.store.SaveAdventure(ctx, doc, nil, "imported"); err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, "could not import package", err)
	}
	return &importPackageOutput{Body: doc}, nil
}

type textExportOutput struct {
	ContentType        string `header:"Content-Type"`
	ContentDisposition string `header:"Content-Disposition"`
	Body               []byte
}

func (s *Server) exportMarkdown(ctx context.Context, input *idInput) (*textExportOutput, error) {
	doc, err := s.store.GetAdventure(ctx, input.ID)
	if errors.Is(err, store.ErrNotFound) {
		return nil, huma.NewError(http.StatusNotFound, "adventure not found")
	}
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, "could not export adventure")
	}
	return &textExportOutput{
		ContentType:        "text/markdown; charset=utf-8",
		ContentDisposition: fmt.Sprintf(`attachment; filename="%s.md"`, safeName(doc.Name.ENUS)),
		Body:               []byte(exporter.Markdown(doc)),
	}, nil
}

func (s *Server) exportHTML(ctx context.Context, input *idInput) (*textExportOutput, error) {
	doc, err := s.store.GetAdventure(ctx, input.ID)
	if errors.Is(err, store.ErrNotFound) {
		return nil, huma.NewError(http.StatusNotFound, "adventure not found")
	}
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, "could not render adventure")
	}
	content, err := exporter.HTML(doc)
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, "could not render adventure", err)
	}
	return &textExportOutput{ContentType: "text/html; charset=utf-8", Body: content}, nil
}

type closeSessionInput struct {
	ID            string `path:"id"`
	Authorization string `header:"Authorization"`
}

type rotateSessionCodeOutput struct{ Body session.CodeStatus }

func (s *Server) rotateSessionCode(ctx context.Context, input *closeSessionInput) (*rotateSessionCodeOutput, error) {
	token := bearerToken(input.Authorization)
	if token == "" {
		return nil, huma.NewError(http.StatusUnauthorized, "missing session token")
	}
	status, err := s.hub.RotateCode(ctx, input.ID, token)
	if errors.Is(err, session.ErrSessionNotFound) {
		return nil, huma.NewError(http.StatusNotFound, "session not found")
	}
	if errors.Is(err, session.ErrSessionCorrupt) {
		return nil, huma.NewError(http.StatusInternalServerError, "session recovery failed", err)
	}
	if err != nil {
		return nil, huma.NewError(http.StatusForbidden, "could not rotate session code", err)
	}
	return &rotateSessionCodeOutput{Body: status}, nil
}

func (s *Server) closeSession(ctx context.Context, input *closeSessionInput) (*emptyOutput, error) {
	token := bearerToken(input.Authorization)
	if token == "" {
		return nil, huma.NewError(http.StatusUnauthorized, "missing session token")
	}
	if err := s.hub.Close(ctx, input.ID, token); errors.Is(err, session.ErrSessionNotFound) {
		return nil, huma.NewError(http.StatusNotFound, "session not found")
	} else if errors.Is(err, session.ErrSessionCorrupt) {
		return nil, huma.NewError(http.StatusInternalServerError, "session recovery failed", err)
	} else if err != nil {
		return nil, huma.NewError(http.StatusForbidden, "could not close session", err)
	}
	return &emptyOutput{}, nil
}

func bearerToken(value string) string { return strings.TrimSpace(strings.TrimPrefix(value, "Bearer ")) }

type listAssetsOutput struct {
	Body []domain.AssetRef
}

func (s *Server) listAssets(ctx context.Context, _ *struct{}) (*listAssetsOutput, error) {
	assets, err := s.store.ListAssets(ctx)
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, "could not list assets", err)
	}
	return &listAssetsOutput{Body: assets}, nil
}

type assetUploadForm struct {
	File    huma.FormFile `form:"file" required:"true"`
	Creator string        `form:"creator" required:"false"`
	License string        `form:"license" required:"false"`
}

type importAssetInput struct {
	RawBody huma.MultipartFormFiles[assetUploadForm]
}

type importAssetOutput struct {
	Body domain.AssetRef
}

func (s *Server) importAsset(ctx context.Context, input *importAssetInput) (*importAssetOutput, error) {
	form := input.RawBody.Data()
	defer form.File.Close()
	ref, err := s.assets.Import(
		ctx,
		form.File.Filename,
		form.File,
		form.Creator,
		form.License,
	)
	if err != nil {
		return nil, huma.NewError(http.StatusUnprocessableEntity, "invalid asset", err)
	}
	return &importAssetOutput{Body: ref}, nil
}

type aiEnrichRequest struct {
	Spec  domain.AdventureSpec `json:"spec"`
	Model string               `json:"model,omitempty"`
}

type aiEnrichInput struct {
	OpenAIKey string `header:"X-OpenAI-API-Key"`
	Body      aiEnrichRequest
}

type aiEnrichOutput struct {
	Body ai.Result
}

func (s *Server) enrichWithAI(ctx context.Context, input *aiEnrichInput) (*aiEnrichOutput, error) {
	key := strings.TrimSpace(input.OpenAIKey)
	if key == "" {
		return nil, huma.NewError(http.StatusBadRequest, "missing API key: the key is required and is not persisted")
	}
	result, err := ai.NewOpenAI(key, input.Body.Model).Enrich(ctx, input.Body.Spec)
	if err != nil {
		return nil, huma.NewError(http.StatusBadGateway, "AI enrichment failed", err)
	}
	return &aiEnrichOutput{Body: result}, nil
}

func (s *Server) rawHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthBody{Status: "ok", Version: "1.0.0-alpha.1", SchemaVersion: domain.SchemaVersion, GeneratorVersion: domain.GeneratorVersion})
}

func (s *Server) staticHandler() http.Handler {
	if s.webDir == "" && s.webFS == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				writeProblem(w, http.StatusNotFound, "Endpoint not found", "")
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = io.WriteString(w, "<!doctype html><title>DDivination</title><h1>DDivination API is running</h1><p>Start the Vite frontend or set DDIVINATION_WEB_DIR.</p>")
		})
	}
	if s.webDir != "" {
		root, err := filepath.Abs(s.webDir)
		if err != nil {
			root = s.webDir
		}
		return spaHandler(http.Dir(root), root)
	}
	return embeddedSPAHandler(s.webFS)
}

func spaHandler(fileSystem http.FileSystem, root string) http.Handler {
	files := http.FileServer(fileSystem)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			writeProblem(w, http.StatusNotFound, "Endpoint not found", "")
			return
		}
		name := cleanURLPath(r.URL.Path)
		requested := filepath.Join(root, filepath.FromSlash(name))
		if info, err := os.Stat(requested); err == nil && !info.IsDir() {
			files.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(root, "index.html"))
	})
}

func embeddedSPAHandler(fileSystem fs.FS) http.Handler {
	files := http.FileServer(http.FS(fileSystem))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			writeProblem(w, http.StatusNotFound, "Endpoint not found", "")
			return
		}
		name := cleanURLPath(r.URL.Path)
		if name != "" {
			if info, err := fs.Stat(fileSystem, name); err == nil && !info.IsDir() {
				files.ServeHTTP(w, r)
				return
			}
		}
		index, err := fs.ReadFile(fileSystem, "index.html")
		if err != nil {
			http.Error(w, "embedded frontend unavailable", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(index)
	})
}

func cleanURLPath(value string) string {
	cleaned := path.Clean("/" + strings.TrimLeft(value, "/"))
	return strings.TrimPrefix(cleaned, "/")
}

func starterCatalog() domain.Catalog {
	return domain.BundledCatalog()
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any, limit int64) error {
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeProblem(w http.ResponseWriter, status int, title, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"type":   "about:blank",
		"title":  title,
		"status": status,
		"detail": detail,
	})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; connect-src 'self' ws: wss:; img-src 'self' data: blob:; style-src 'self' 'unsafe-inline'; script-src 'self' 'wasm-unsafe-eval'; worker-src 'self' blob:")
		next.ServeHTTP(w, r)
	})
}

func allowedOrigins(r *http.Request) []string {
	host := r.Host
	if host == "" {
		return nil
	}
	return []string{host, "http://" + host, "https://" + host}
}

func randomID(size int) (string, error) {
	value := make([]byte, size)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func randomSeed() uint64 {
	value := make([]byte, 8)
	if _, err := rand.Read(value); err != nil {
		return uint64(time.Now().UnixNano())
	}
	var seed uint64
	for _, b := range value {
		seed = seed<<8 | uint64(b)
	}
	// JSON/JavaScript can represent integers exactly only through 2^53 - 1.
	return seed & ((1 << 53) - 1)
}

func safeName(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			builder.WriteRune(r)
		case r == ' ', r == '-', r == '_':
			builder.WriteByte('-')
		}
	}
	result := strings.Trim(builder.String(), "-")
	if result == "" {
		return "adventure"
	}
	return result
}

func etagVersion(value string) (*int64, error) {
	value = strings.Trim(value, `" `)
	if value == "" {
		return nil, nil
	}
	version, err := strconv.ParseInt(value, 10, 64)
	return &version, err
}
