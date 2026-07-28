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
	store  *store.Store
	hub    *session.Hub
	lan    LANEnabler
	logger *slog.Logger
	webDir string
	webFS  fs.FS
	assets *asset.Manager
}

type Handlers struct {
	Local http.Handler
	LAN   http.Handler
}

func New(s *store.Store, hub *session.Hub, lan LANEnabler, logger *slog.Logger, webDir string, webFS fs.FS, assetsDir string) *Server {
	return &Server{
		store: s, hub: hub, lan: lan, logger: logger, webDir: webDir, webFS: webFS,
		assets: asset.NewManager(assetsDir, s),
	}
}

func (s *Server) Handlers() Handlers {
	localMux := http.NewServeMux()
	config := huma.DefaultConfig("DDivination API", "1.0.0-alpha.1")
	config.Info.Description = "Local-first 3D VTT and deterministic 5E-compatible adventure generator."
	localAPI := humago.New(localMux, config)
	s.registerREST(localAPI)
	s.registerShared(localMux)
	localMux.HandleFunc("GET /api/v1/assets", s.listAssets)
	localMux.HandleFunc("POST /api/v1/assets", s.importAsset)
	localMux.HandleFunc("POST /api/v1/ai/enrich", s.enrichWithAI)
	localMux.Handle("/", s.staticHandler())

	lanMux := http.NewServeMux()
	lanMux.HandleFunc("GET /api/v1/health", s.rawHealth)
	lanMux.HandleFunc("POST /api/v1/sessions/{id}/join", s.rawJoinSession)
	lanMux.HandleFunc("GET /api/v1/sessions/{id}/stream", s.streamSession)
	lanMux.Handle("/", s.staticHandler())
	return Handlers{Local: securityHeaders(localMux), LAN: securityHeaders(lanMux)}
}

func (s *Server) registerREST(api huma.API) {
	huma.Get(api, "/api/v1/health", func(_ context.Context, _ *struct{}) (*healthOutput, error) {
		return &healthOutput{Body: healthBody{
			Status:           "ok",
			Version:          "1.0.0-alpha.1",
			SchemaVersion:    domain.SchemaVersion,
			GeneratorVersion: domain.GeneratorVersion,
		}}, nil
	})

	huma.Get(api, "/api/v1/catalog", func(_ context.Context, _ *struct{}) (*catalogOutput, error) {
		return &catalogOutput{Body: starterCatalog()}, nil
	})

	huma.Register(api, huma.Operation{
		Method:        http.MethodPost,
		Path:          "/api/v1/generation-runs",
		OperationID:   "create-generation-run",
		Summary:       "Generate and persist an adventure",
		DefaultStatus: http.StatusAccepted,
	}, s.createGeneration)

	huma.Get(api, "/api/v1/generation-runs/{id}", s.getGeneration)
	huma.Get(api, "/api/v1/adventures", s.listAdventures)
	huma.Get(api, "/api/v1/adventures/{id}", s.getAdventure)
	huma.Put(api, "/api/v1/adventures/{id}", s.updateAdventure)
	huma.Post(api, "/api/v1/adventures/{id}/checkpoints", s.checkpointAdventure)
	huma.Delete(api, "/api/v1/adventures/{id}", s.deleteAdventure)
	huma.Register(api, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/api/v1/sessions",
		OperationID: "create-session",
		Summary:     "Open a local network VTT session",
	}, s.createSession)
}

func (s *Server) registerShared(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/sessions/{id}/join", s.rawJoinSession)
	mux.HandleFunc("GET /api/v1/sessions/{id}/stream", s.streamSession)
	mux.HandleFunc("GET /api/v1/packages/{id}", s.exportPackage)
	mux.HandleFunc("POST /api/v1/packages", s.importPackage)
	mux.HandleFunc("GET /api/v1/adventures/{id}/export.md", s.exportMarkdown)
	mux.HandleFunc("GET /api/v1/adventures/{id}/print", s.exportHTML)
	mux.HandleFunc("DELETE /api/v1/sessions/{id}", s.closeSession)
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

type CatalogItem struct {
	Index   string               `json:"index"`
	Name    domain.LocalizedText `json:"name"`
	Kind    string               `json:"kind"`
	CR      float64              `json:"cr,omitempty"`
	Source  string               `json:"source"`
	License string               `json:"license"`
}

type Catalog struct {
	Version     string             `json:"version"`
	Ruleset     string             `json:"ruleset"`
	Items       []CatalogItem      `json:"items"`
	Attribution domain.Attribution `json:"attribution"`
}

type catalogOutput struct {
	Body Catalog
}

type generationInput struct {
	OpenAIKey string `header:"X-OpenAI-API-Key"`
	Body      struct {
		Spec domain.AdventureSpec `json:"spec"`
		Seed *uint64              `json:"seed,omitempty" maximum:"9007199254740991"`
	}
}

type generationResult struct {
	Run       domain.GenerationRun     `json:"run"`
	Adventure domain.AdventureDocument `json:"adventure"`
}

type generationOutput struct {
	Body generationResult
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
		Status:           "running",
		Stage:            "building-topology",
		Progress:         20,
		Seed:             seed,
		GeneratorVersion: domain.GeneratorVersion,
		Spec:             input.Body.Spec,
		Diagnostics:      []string{"procedural-mode", "offline-ready"},
		CreatedAt:        now,
	}
	if err := s.store.SaveGenerationRun(ctx, run); err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, "could not persist generation run")
	}
	doc, err := generator.Generate(input.Body.Spec, seed, now)
	if err != nil {
		run.Status = "failed"
		run.Stage = "validation-failed"
		run.Diagnostics = append(run.Diagnostics, err.Error())
		completed := time.Now().UTC()
		run.CompletedAt = &completed
		_ = s.store.SaveGenerationRun(ctx, run)
		return nil, huma.NewError(http.StatusUnprocessableEntity, err.Error())
	}
	if input.Body.Spec.UseAI {
		if strings.TrimSpace(input.OpenAIKey) == "" {
			run.Diagnostics = append(run.Diagnostics, "ai-skipped:key-missing", "procedural-fallback-used")
		} else {
			result, enrichErr := ai.NewOpenAI(input.OpenAIKey, "").Enrich(ctx, input.Body.Spec)
			if enrichErr != nil {
				run.Diagnostics = append(run.Diagnostics, "ai-failed:"+enrichErr.Error(), "procedural-fallback-used")
			} else {
				doc.Narrative = result.Enrichment
				run.Diagnostics = append(
					run.Diagnostics,
					fmt.Sprintf("ai-provider:%s", result.Diagnostics.Provider),
					fmt.Sprintf("ai-model:%s", result.Diagnostics.Model),
					fmt.Sprintf("ai-input-tokens:%d", result.Diagnostics.InputTokens),
					fmt.Sprintf("ai-output-tokens:%d", result.Diagnostics.OutputTokens),
					fmt.Sprintf("ai-latency-ms:%d", result.Diagnostics.Latency.Milliseconds()),
				)
			}
		}
	}
	if err := s.store.SaveAdventure(ctx, doc, nil, "generated"); err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, "could not persist generated adventure")
	}
	completed := time.Now().UTC()
	run.Status = "completed"
	run.Stage = "completed"
	run.Progress = 100
	run.AdventureID = doc.ID
	run.CompletedAt = &completed
	if err := s.store.SaveGenerationRun(ctx, run); err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, "could not finalize generation run")
	}
	return &generationOutput{Body: generationResult{Run: run, Adventure: doc}}, nil
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
	} else if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, "could not update adventure")
	}
	return &adventureOutput{ETag: fmt.Sprintf(`"%d"`, input.Body.Version), Body: input.Body}, nil
}

func (s *Server) checkpointAdventure(ctx context.Context, input *idInput) (*adventureOutput, error) {
	doc, err := s.store.GetAdventure(ctx, input.ID)
	if errors.Is(err, store.ErrNotFound) {
		return nil, huma.NewError(http.StatusNotFound, "adventure not found")
	}
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, "could not load adventure")
	}
	if err := s.store.SaveAdventure(ctx, doc, &doc.Version, "manual-checkpoint"); err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, "could not create checkpoint")
	}
	return &adventureOutput{ETag: fmt.Sprintf(`"%d"`, doc.Version), Body: doc}, nil
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
	Code string `json:"code"`
	Name string `json:"name"`
	Role string `json:"role"`
}

func (s *Server) rawJoinSession(w http.ResponseWriter, r *http.Request) {
	var request joinRequest
	if err := decodeJSON(w, r, &request, 1<<20); err != nil {
		writeProblem(w, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}
	joined, err := s.hub.Join(r.Context(), r.PathValue("id"), request.Code, request.Name, request.Role)
	if err != nil {
		status := http.StatusUnauthorized
		if errors.Is(err, session.ErrSessionNotFound) {
			status = http.StatusNotFound
		}
		writeProblem(w, status, "Could not join session", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, joined)
}

func (s *Server) streamSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")
	token := r.URL.Query().Get("token")
	sub, snapshot, err := s.hub.Subscribe(sessionID, token)
	if err != nil {
		writeProblem(w, http.StatusUnauthorized, "WebSocket authorization failed", err.Error())
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
	outbound := make(chan any, 128)
	readErrors := make(chan error, 1)
	outbound <- snapshot

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

func (s *Server) exportPackage(w http.ResponseWriter, r *http.Request) {
	doc, err := s.store.GetAdventure(r.Context(), r.PathValue("id"))
	if errors.Is(err, store.ErrNotFound) {
		writeProblem(w, http.StatusNotFound, "Adventure not found", "")
		return
	}
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Export failed", err.Error())
		return
	}
	content, err := packageio.Export(doc)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Export failed", err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/vnd.ddivination+zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.ddivination"`, safeName(doc.Name.ENUS)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}

func (s *Server) importPackage(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, packageio.MaxPackageSize)
	content, err := io.ReadAll(r.Body)
	if err != nil {
		writeProblem(w, http.StatusRequestEntityTooLarge, "Package too large", err.Error())
		return
	}
	doc, err := packageio.Import(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		writeProblem(w, http.StatusUnprocessableEntity, "Invalid package", err.Error())
		return
	}
	if err := s.store.SaveAdventure(r.Context(), doc, nil, "imported"); err != nil {
		writeProblem(w, http.StatusInternalServerError, "Import failed", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, doc)
}

func (s *Server) exportMarkdown(w http.ResponseWriter, r *http.Request) {
	doc, err := s.store.GetAdventure(r.Context(), r.PathValue("id"))
	if errors.Is(err, store.ErrNotFound) {
		writeProblem(w, http.StatusNotFound, "Adventure not found", "")
		return
	}
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Export failed", err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.md"`, safeName(doc.Name.ENUS)))
	_, _ = w.Write(exporter.Markdown(doc))
}

func (s *Server) exportHTML(w http.ResponseWriter, r *http.Request) {
	doc, err := s.store.GetAdventure(r.Context(), r.PathValue("id"))
	if errors.Is(err, store.ErrNotFound) {
		writeProblem(w, http.StatusNotFound, "Adventure not found", "")
		return
	}
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Export failed", err.Error())
		return
	}
	content, err := exporter.HTML(doc)
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Export failed", err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(content)
}

func (s *Server) closeSession(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	if token == "" {
		writeProblem(w, http.StatusUnauthorized, "Missing session token", "")
		return
	}
	if err := s.hub.Close(r.Context(), r.PathValue("id"), token); errors.Is(err, session.ErrSessionNotFound) {
		writeProblem(w, http.StatusNotFound, "Session not found", "")
		return
	} else if err != nil {
		writeProblem(w, http.StatusForbidden, "Could not close session", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listAssets(w http.ResponseWriter, r *http.Request) {
	assets, err := s.store.ListAssets(r.Context())
	if err != nil {
		writeProblem(w, http.StatusInternalServerError, "Could not list assets", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, assets)
}

func (s *Server) importAsset(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, asset.MaxGLBSize+(1<<20))
	if err := r.ParseMultipartForm(asset.MaxGLBSize); err != nil {
		writeProblem(w, http.StatusBadRequest, "Invalid asset upload", err.Error())
		return
	}
	defer r.MultipartForm.RemoveAll()
	file, header, err := r.FormFile("file")
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Missing asset file", err.Error())
		return
	}
	defer file.Close()
	ref, err := s.assets.Import(
		r.Context(),
		header.Filename,
		file,
		r.FormValue("creator"),
		r.FormValue("license"),
	)
	if err != nil {
		writeProblem(w, http.StatusUnprocessableEntity, "Invalid asset", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, ref)
}

type aiEnrichRequest struct {
	Spec  domain.AdventureSpec `json:"spec"`
	Model string               `json:"model,omitempty"`
}

func (s *Server) enrichWithAI(w http.ResponseWriter, r *http.Request) {
	var request aiEnrichRequest
	if err := decodeJSON(w, r, &request, 1<<20); err != nil {
		writeProblem(w, http.StatusBadRequest, "Invalid AI enrichment request", err.Error())
		return
	}
	key := strings.TrimSpace(r.Header.Get("X-OpenAI-API-Key"))
	if key == "" {
		writeProblem(w, http.StatusBadRequest, "Missing API key", "The key is required for this request and is not persisted.")
		return
	}
	result, err := ai.NewOpenAI(key, request.Model).Enrich(r.Context(), request.Spec)
	if err != nil {
		writeProblem(w, http.StatusBadGateway, "AI enrichment failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
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

func starterCatalog() Catalog {
	attribution := domain.Attribution{
		Title:   "System Reference Document 5.2.1",
		Creator: "Wizards of the Coast LLC",
		Source:  "https://www.dndbeyond.com/srd",
		License: "CC-BY-4.0",
		Notice:  "This work includes material from the System Reference Document 5.2.1 by Wizards of the Coast LLC.",
	}
	return Catalog{
		Version: "starter-2024.1",
		Ruleset: "5E 2024 / SRD 5.2.1",
		Items: []CatalogItem{
			{Index: "skeleton", Name: domain.LocalizedText{PTBR: "Esqueleto", ENUS: "Skeleton"}, Kind: "monster", CR: .25, Source: attribution.Title, License: attribution.License},
			{Index: "goblin-warrior", Name: domain.LocalizedText{PTBR: "Guerreiro Goblin", ENUS: "Goblin Warrior"}, Kind: "monster", CR: .25, Source: attribution.Title, License: attribution.License},
			{Index: "ogre", Name: domain.LocalizedText{PTBR: "Ogro", ENUS: "Ogre"}, Kind: "monster", CR: 2, Source: attribution.Title, License: attribution.License},
		},
		Attribution: attribution,
	}
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
	writeJSON(w, status, map[string]any{
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
