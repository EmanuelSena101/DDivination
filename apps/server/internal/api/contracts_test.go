package api

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/EmanuelSena101/DDivination/apps/server/internal/domain"
	"github.com/EmanuelSena101/DDivination/apps/server/internal/generator"
	"github.com/EmanuelSena101/DDivination/apps/server/internal/session"
	"github.com/EmanuelSena101/DDivination/apps/server/internal/store"
)

func newContractTestServer(t *testing.T) (*Server, Handlers) {
	t.Helper()
	root := t.TempDir()
	database, err := store.Open(filepath.Join(root, "test.sqlite3"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Errorf("close database: %v", err)
		}
	})
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := New(database, session.NewHub(database), nil, logger, "", nil, filepath.Join(root, "assets"))
	return server, server.Handlers()
}

func TestOpenAPIContainsEveryRESTContractWithStableOperationID(t *testing.T) {
	_, handlers := newContractTestServer(t)
	response := httptest.NewRecorder()
	handlers.Local.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/openapi.json", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected OpenAPI 200, got %d: %s", response.Code, response.Body.String())
	}

	var document struct {
		Paths map[string]map[string]struct {
			OperationID string `json:"operationId"`
			Responses   map[string]struct {
				Content map[string]any `json:"content"`
			} `json:"responses"`
		} `json:"paths"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &document); err != nil {
		t.Fatal(err)
	}

	expectedCount := 0
	seenIDs := make(map[string]string)
	for _, contract := range endpointContracts {
		if contract.Protocol != "rest" {
			continue
		}
		expectedCount++
		operation, ok := document.Paths[contract.Path][strings.ToLower(contract.Method)]
		if !ok {
			t.Errorf("OpenAPI missing %s %s", contract.Method, contract.Path)
			continue
		}
		if operation.OperationID != contract.OperationID {
			t.Errorf("%s %s: expected operationId %q, got %q", contract.Method, contract.Path, contract.OperationID, operation.OperationID)
		}
		if previous, duplicate := seenIDs[operation.OperationID]; duplicate {
			t.Errorf("duplicate operationId %q on %s and %s %s", operation.OperationID, previous, contract.Method, contract.Path)
		}
		seenIDs[operation.OperationID] = contract.Method + " " + contract.Path

		hasProblemResponse := false
		for status, candidate := range operation.Responses {
			if strings.HasPrefix(status, "2") {
				continue
			}
			if _, ok := candidate.Content["application/problem+json"]; ok {
				hasProblemResponse = true
				break
			}
		}
		if !hasProblemResponse {
			t.Errorf("%s %s has no application/problem+json error response", contract.Method, contract.Path)
		}
	}

	actualCount := 0
	for _, pathItem := range document.Paths {
		for method := range pathItem {
			switch method {
			case "get", "post", "put", "patch", "delete":
				actualCount++
			}
		}
	}
	if actualCount != expectedCount {
		t.Fatalf("expected %d documented REST operations, got %d", expectedCount, actualCount)
	}
}

func TestLANAllowlistRejectsEveryLocalOnlyEndpoint(t *testing.T) {
	_, handlers := newContractTestServer(t)

	health := httptest.NewRecorder()
	handlers.LAN.ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/api/v1/health", nil))
	if health.Code != http.StatusOK {
		t.Fatalf("LAN health: expected 200, got %d", health.Code)
	}

	for _, contract := range endpointContracts {
		if contract.Exposure != exposureLocal {
			continue
		}
		requestPath := strings.ReplaceAll(contract.Path, "{id}", "missing")
		response := httptest.NewRecorder()
		handlers.LAN.ServeHTTP(response, httptest.NewRequest(contract.Method, requestPath, nil))
		if response.Code != http.StatusNotFound {
			t.Errorf("LAN leaked %s %s: expected 404, got %d", contract.Method, contract.Path, response.Code)
			continue
		}
		assertProblemContentType(t, response)
		if !strings.Contains(response.Body.String(), "Endpoint not found") {
			t.Errorf("LAN %s %s reached a handler: %s", contract.Method, contract.Path, response.Body.String())
		}
	}
}

func TestLocalAndLANFailuresUseProblemJSON(t *testing.T) {
	_, handlers := newContractTestServer(t)
	testCases := []struct {
		name    string
		handler http.Handler
		request *http.Request
		status  int
	}{
		{
			name:    "local Huma error",
			handler: handlers.Local,
			request: httptest.NewRequest(http.MethodGet, "/api/v1/adventures/missing", nil),
			status:  http.StatusNotFound,
		},
		{
			name:    "LAN malformed join",
			handler: handlers.LAN,
			request: httptest.NewRequest(http.MethodPost, "/api/v1/sessions/missing/join", strings.NewReader("{")),
			status:  http.StatusBadRequest,
		},
		{
			name:    "LAN unauthorized WebSocket",
			handler: handlers.LAN,
			request: httptest.NewRequest(http.MethodGet, "/api/v1/sessions/missing/stream", nil),
			status:  http.StatusUnauthorized,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			testCase.handler.ServeHTTP(response, testCase.request)
			if response.Code != testCase.status {
				t.Fatalf("expected %d, got %d: %s", testCase.status, response.Code, response.Body.String())
			}
			assertProblemContentType(t, response)
		})
	}
}

func TestPortableExportsPreserveTheirMediaTypes(t *testing.T) {
	server, handlers := newContractTestServer(t)
	spec := domain.DefaultAdventureSpec()
	document, err := generator.Generate(spec, 42, time.Date(2026, time.July, 28, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if err := server.store.SaveAdventure(t.Context(), document, nil, "test"); err != nil {
		t.Fatal(err)
	}

	for _, testCase := range []struct {
		path        string
		contentType string
	}{
		{"/api/v1/packages/" + document.ID, "application/vnd.ddivination+zip"},
		{"/api/v1/adventures/" + document.ID + "/export.md", "text/markdown; charset=utf-8"},
		{"/api/v1/adventures/" + document.ID + "/print", "text/html; charset=utf-8"},
	} {
		response := httptest.NewRecorder()
		handlers.Local.ServeHTTP(response, httptest.NewRequest(http.MethodGet, testCase.path, nil))
		if response.Code != http.StatusOK {
			t.Fatalf("%s: expected 200, got %d: %s", testCase.path, response.Code, response.Body.String())
		}
		if actual := response.Header().Get("Content-Type"); actual != testCase.contentType {
			t.Errorf("%s: expected %q, got %q", testCase.path, testCase.contentType, actual)
		}
		if response.Body.Len() == 0 {
			t.Errorf("%s returned an empty body", testCase.path)
		}
	}
}

func TestGenerationRunIsAsynchronousObservableAndPersisted(t *testing.T) {
	_, handlers := newContractTestServer(t)
	response := serveJSON(
		t,
		handlers.Local,
		http.MethodPost,
		"/api/v1/generation-runs",
		map[string]any{
			"spec": domain.DefaultAdventureSpec(),
			"seed": 808,
		},
		nil,
	)
	if response.Code != http.StatusAccepted {
		t.Fatalf("create run: expected 202, got %d: %s", response.Code, response.Body.String())
	}
	var queued domain.GenerationRun
	if err := json.Unmarshal(response.Body.Bytes(), &queued); err != nil {
		t.Fatal(err)
	}
	if queued.Status != "queued" || queued.Progress != 0 || len(queued.Stages) != 1 {
		t.Fatalf("expected queued run, got %#v", queued)
	}

	var completed domain.GenerationRun
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		get := httptest.NewRecorder()
		handlers.Local.ServeHTTP(
			get,
			httptest.NewRequest(http.MethodGet, "/api/v1/generation-runs/"+queued.ID, nil),
		)
		if get.Code != http.StatusOK {
			t.Fatalf("get run: expected 200, got %d: %s", get.Code, get.Body.String())
		}
		if err := json.Unmarshal(get.Body.Bytes(), &completed); err != nil {
			t.Fatal(err)
		}
		if completed.Status == "completed" {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if completed.Status != "completed" || completed.AdventureID == "" {
		t.Fatalf("run did not complete: %#v", completed)
	}

	adventure := httptest.NewRecorder()
	handlers.Local.ServeHTTP(
		adventure,
		httptest.NewRequest(http.MethodGet, "/api/v1/adventures/"+completed.AdventureID, nil),
	)
	if adventure.Code != http.StatusOK {
		t.Fatalf("persisted adventure: expected 200, got %d: %s", adventure.Code, adventure.Body.String())
	}

	list := httptest.NewRecorder()
	handlers.Local.ServeHTTP(
		list,
		httptest.NewRequest(http.MethodGet, "/api/v1/generation-runs?limit=10", nil),
	)
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), queued.ID) {
		t.Fatalf("list runs: expected run %s, got %d: %s", queued.ID, list.Code, list.Body.String())
	}
}

func TestAdventureEditingCheckpointsAndRestore(t *testing.T) {
	server, handlers := newContractTestServer(t)
	document, err := generator.Generate(
		domain.DefaultAdventureSpec(),
		91,
		time.Date(2026, time.July, 30, 12, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := server.store.SaveAdventure(t.Context(), document, nil, "generated"); err != nil {
		t.Fatal(err)
	}

	document.Name.PTBR = "Edição persistida"
	response := serveJSON(t, handlers.Local, http.MethodPut, "/api/v1/adventures/"+document.ID, document, map[string]string{
		"If-Match": `"1"`,
	})
	if response.Code != http.StatusOK {
		t.Fatalf("update: expected 200, got %d: %s", response.Code, response.Body.String())
	}
	var saved domain.AdventureDocument
	if err := json.Unmarshal(response.Body.Bytes(), &saved); err != nil {
		t.Fatal(err)
	}
	if saved.Version != 2 || response.Header().Get("ETag") != `"2"` {
		t.Fatalf("expected version 2 and ETag, got version=%d etag=%q", saved.Version, response.Header().Get("ETag"))
	}

	conflict := serveJSON(t, handlers.Local, http.MethodPut, "/api/v1/adventures/"+document.ID, document, map[string]string{
		"If-Match": `"1"`,
	})
	if conflict.Code != http.StatusConflict {
		t.Fatalf("stale update: expected 409, got %d: %s", conflict.Code, conflict.Body.String())
	}

	checkpointResponse := serveJSON(t, handlers.Local, http.MethodPost, "/api/v1/adventures/"+document.ID+"/checkpoints", nil, nil)
	if checkpointResponse.Code != http.StatusCreated {
		t.Fatalf("checkpoint: expected 201, got %d: %s", checkpointResponse.Code, checkpointResponse.Body.String())
	}
	var checkpoint domain.AdventureSnapshotSummary
	if err := json.Unmarshal(checkpointResponse.Body.Bytes(), &checkpoint); err != nil {
		t.Fatal(err)
	}
	if checkpoint.Version != 2 || checkpoint.Reason != "manual-checkpoint" {
		t.Fatalf("unexpected checkpoint: %#v", checkpoint)
	}

	saved.Name.PTBR = "Nova edição"
	response = serveJSON(t, handlers.Local, http.MethodPut, "/api/v1/adventures/"+document.ID, saved, map[string]string{
		"If-Match": `"2"`,
	})
	if response.Code != http.StatusOK {
		t.Fatalf("second update: expected 200, got %d: %s", response.Code, response.Body.String())
	}
	if err := json.Unmarshal(response.Body.Bytes(), &saved); err != nil {
		t.Fatal(err)
	}

	restorePath := "/api/v1/adventures/" + document.ID + "/checkpoints/" + checkpoint.ID + "/restore"
	response = serveJSON(t, handlers.Local, http.MethodPost, restorePath, nil, map[string]string{
		"If-Match": `"3"`,
	})
	if response.Code != http.StatusOK {
		t.Fatalf("restore: expected 200, got %d: %s", response.Code, response.Body.String())
	}
	var restored domain.AdventureDocument
	if err := json.Unmarshal(response.Body.Bytes(), &restored); err != nil {
		t.Fatal(err)
	}
	if restored.Version != 4 || restored.Name.PTBR != "Edição persistida" {
		t.Fatalf("checkpoint was not restored as version 4: %#v", restored)
	}

	list := httptest.NewRecorder()
	handlers.Local.ServeHTTP(list, httptest.NewRequest(http.MethodGet, "/api/v1/adventures/"+document.ID+"/checkpoints", nil))
	if list.Code != http.StatusOK {
		t.Fatalf("list checkpoints: expected 200, got %d: %s", list.Code, list.Body.String())
	}
	var snapshots []domain.AdventureSnapshotSummary
	if err := json.Unmarshal(list.Body.Bytes(), &snapshots); err != nil {
		t.Fatal(err)
	}
	if len(snapshots) != 5 {
		t.Fatalf("expected five immutable snapshots, got %d", len(snapshots))
	}

	restored.Floors[0].Entities[0].Position = domain.GridPosition{X: -1, Z: -1}
	invalid := serveJSON(t, handlers.Local, http.MethodPut, "/api/v1/adventures/"+document.ID, restored, map[string]string{
		"If-Match": `"4"`,
	})
	if invalid.Code != http.StatusUnprocessableEntity {
		t.Fatalf("invalid update: expected 422, got %d: %s", invalid.Code, invalid.Body.String())
	}
}

func serveJSON(
	t *testing.T,
	handler http.Handler,
	method string,
	path string,
	body any,
	headers map[string]string,
) *httptest.ResponseRecorder {
	t.Helper()
	var payload io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		payload = bytes.NewReader(encoded)
	}
	request := httptest.NewRequest(method, path, payload)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	for name, value := range headers {
		request.Header.Set(name, value)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func assertProblemContentType(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	if actual := response.Header().Get("Content-Type"); actual != "application/problem+json" {
		t.Errorf("expected application/problem+json, got %q", actual)
	}
}
