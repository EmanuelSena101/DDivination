package api

import (
	"fmt"
	"net/http"

	"github.com/EmanuelSena101/DDivination/apps/server/internal/asset"
	"github.com/EmanuelSena101/DDivination/apps/server/internal/packageio"
	"github.com/danielgtaylor/huma/v2"
)

type endpointExposure string

const (
	exposureLocal endpointExposure = "local"
	exposureLAN   endpointExposure = "lan"
)

type endpointContract struct {
	Method      string
	Path        string
	OperationID string
	Exposure    endpointExposure
	Protocol    string
}

// endpointContracts is the auditable boundary for HTTP and WebSocket access.
// REST operations are registered with Huma below. LAN-only raw handlers are
// tested against this matrix to prevent administrative routes from leaking.
var endpointContracts = []endpointContract{
	{http.MethodGet, "/api/v1/health", "getHealth", exposureLAN, "rest"},
	{http.MethodGet, "/api/v1/catalog", "getCatalog", exposureLocal, "rest"},
	{http.MethodPost, "/api/v1/generation-runs", "createGenerationRun", exposureLocal, "rest"},
	{http.MethodGet, "/api/v1/generation-runs", "listGenerationRuns", exposureLocal, "rest"},
	{http.MethodGet, "/api/v1/generation-runs/{id}", "getGenerationRun", exposureLocal, "rest"},
	{http.MethodDelete, "/api/v1/generation-runs/{id}", "cancelGenerationRun", exposureLocal, "rest"},
	{http.MethodGet, "/api/v1/generation-runs/{id}/stream", "streamGenerationRun", exposureLocal, "websocket"},
	{http.MethodGet, "/api/v1/adventures", "listAdventures", exposureLocal, "rest"},
	{http.MethodGet, "/api/v1/adventures/{id}", "getAdventure", exposureLocal, "rest"},
	{http.MethodPut, "/api/v1/adventures/{id}", "updateAdventure", exposureLocal, "rest"},
	{http.MethodDelete, "/api/v1/adventures/{id}", "deleteAdventure", exposureLocal, "rest"},
	{http.MethodGet, "/api/v1/adventures/{id}/checkpoints", "listAdventureCheckpoints", exposureLocal, "rest"},
	{http.MethodPost, "/api/v1/adventures/{id}/checkpoints", "checkpointAdventure", exposureLocal, "rest"},
	{http.MethodPost, "/api/v1/adventures/{id}/checkpoints/{checkpointId}/restore", "restoreAdventureCheckpoint", exposureLocal, "rest"},
	{http.MethodGet, "/api/v1/adventures/{id}/export.md", "exportAdventureMarkdown", exposureLocal, "rest"},
	{http.MethodGet, "/api/v1/adventures/{id}/print", "printAdventure", exposureLocal, "rest"},
	{http.MethodPost, "/api/v1/sessions", "createSession", exposureLocal, "rest"},
	{http.MethodPost, "/api/v1/sessions/{id}/join", "joinSession", exposureLAN, "rest"},
	{http.MethodDelete, "/api/v1/sessions/{id}", "closeSession", exposureLocal, "rest"},
	{http.MethodGet, "/api/v1/sessions/{id}/stream", "streamSession", exposureLAN, "websocket"},
	{http.MethodGet, "/api/v1/packages/{id}", "exportPackage", exposureLocal, "rest"},
	{http.MethodPost, "/api/v1/packages", "importPackage", exposureLocal, "rest"},
	{http.MethodGet, "/api/v1/assets", "listAssets", exposureLocal, "rest"},
	{http.MethodPost, "/api/v1/assets", "importAsset", exposureLocal, "rest"},
	{http.MethodPost, "/api/v1/ai/enrich", "enrichAdventure", exposureLocal, "rest"},
}

func apiOperation(method, path, operationID, summary string, tags []string, status int, errors ...int) huma.Operation {
	return huma.Operation{
		Method:        method,
		Path:          path,
		OperationID:   operationID,
		Summary:       summary,
		Tags:          tags,
		DefaultStatus: status,
		Errors:        errors,
	}
}

func (s *Server) registerREST(api huma.API) {
	huma.Register(api, apiOperation(
		http.MethodGet, "/api/v1/health", "getHealth", "Report server and schema versions",
		[]string{"System"}, http.StatusOK,
	), s.getHealth)
	huma.Register(api, apiOperation(
		http.MethodGet, "/api/v1/catalog", "getCatalog", "Get the bundled 5E-compatible catalog",
		[]string{"Catalog"}, http.StatusOK,
	), s.getCatalog)
	huma.Register(api, apiOperation(
		http.MethodPost, "/api/v1/generation-runs", "createGenerationRun", "Queue an adventure generation run",
		[]string{"Generation"}, http.StatusAccepted,
		http.StatusBadRequest, http.StatusUnprocessableEntity, http.StatusInternalServerError,
	), s.createGeneration)
	huma.Register(api, apiOperation(
		http.MethodGet, "/api/v1/generation-runs", "listGenerationRuns", "List generation runs",
		[]string{"Generation"}, http.StatusOK,
		http.StatusInternalServerError,
	), s.listGenerationRuns)
	huma.Register(api, apiOperation(
		http.MethodGet, "/api/v1/generation-runs/{id}", "getGenerationRun", "Get a generation run",
		[]string{"Generation"}, http.StatusOK,
		http.StatusNotFound, http.StatusInternalServerError,
	), s.getGeneration)
	huma.Register(api, apiOperation(
		http.MethodDelete, "/api/v1/generation-runs/{id}", "cancelGenerationRun", "Cancel an active generation run",
		[]string{"Generation"}, http.StatusOK,
		http.StatusNotFound, http.StatusInternalServerError,
	), s.cancelGeneration)
	huma.Register(api, apiOperation(
		http.MethodGet, "/api/v1/adventures", "listAdventures", "List adventures",
		[]string{"Adventures"}, http.StatusOK,
		http.StatusInternalServerError,
	), s.listAdventures)
	huma.Register(api, apiOperation(
		http.MethodGet, "/api/v1/adventures/{id}", "getAdventure", "Get an adventure",
		[]string{"Adventures"}, http.StatusOK,
		http.StatusNotFound, http.StatusInternalServerError,
	), s.getAdventure)
	huma.Register(api, apiOperation(
		http.MethodPut, "/api/v1/adventures/{id}", "updateAdventure", "Replace an adventure using optimistic locking",
		[]string{"Adventures"}, http.StatusOK,
		http.StatusBadRequest, http.StatusNotFound, http.StatusConflict, http.StatusInternalServerError,
	), s.updateAdventure)
	huma.Register(api, apiOperation(
		http.MethodDelete, "/api/v1/adventures/{id}", "deleteAdventure", "Delete an adventure",
		[]string{"Adventures"}, http.StatusNoContent,
		http.StatusNotFound, http.StatusInternalServerError,
	), s.deleteAdventure)
	huma.Register(api, apiOperation(
		http.MethodGet, "/api/v1/adventures/{id}/checkpoints", "listAdventureCheckpoints", "List immutable adventure checkpoints",
		[]string{"Adventures"}, http.StatusOK,
		http.StatusNotFound, http.StatusInternalServerError,
	), s.listAdventureCheckpoints)
	huma.Register(api, apiOperation(
		http.MethodPost, "/api/v1/adventures/{id}/checkpoints", "checkpointAdventure", "Create an immutable adventure checkpoint",
		[]string{"Adventures"}, http.StatusCreated,
		http.StatusNotFound, http.StatusInternalServerError,
	), s.checkpointAdventure)
	huma.Register(api, apiOperation(
		http.MethodPost, "/api/v1/adventures/{id}/checkpoints/{checkpointId}/restore", "restoreAdventureCheckpoint", "Restore a checkpoint as a new adventure version",
		[]string{"Adventures"}, http.StatusOK,
		http.StatusBadRequest, http.StatusNotFound, http.StatusConflict, http.StatusUnprocessableEntity, http.StatusInternalServerError,
	), s.restoreAdventureCheckpoint)
	huma.Register(api, apiOperation(
		http.MethodGet, "/api/v1/adventures/{id}/export.md", "exportAdventureMarkdown", "Export an adventure as Markdown",
		[]string{"Portability"}, http.StatusOK,
		http.StatusNotFound, http.StatusInternalServerError,
	), s.exportMarkdown)
	huma.Register(api, apiOperation(
		http.MethodGet, "/api/v1/adventures/{id}/print", "printAdventure", "Render a printable adventure as HTML",
		[]string{"Portability"}, http.StatusOK,
		http.StatusNotFound, http.StatusInternalServerError,
	), s.exportHTML)
	huma.Register(api, apiOperation(
		http.MethodPost, "/api/v1/sessions", "createSession", "Open a local network VTT session",
		[]string{"Sessions"}, http.StatusOK,
		http.StatusNotFound, http.StatusInternalServerError,
	), s.createSession)
	huma.Register(api, apiOperation(
		http.MethodPost, "/api/v1/sessions/{id}/join", "joinSession", "Exchange a temporary code for a session token",
		[]string{"Sessions"}, http.StatusOK,
		http.StatusBadRequest, http.StatusUnauthorized, http.StatusNotFound,
	), s.joinSession)
	huma.Register(api, apiOperation(
		http.MethodDelete, "/api/v1/sessions/{id}", "closeSession", "Close a VTT session",
		[]string{"Sessions"}, http.StatusNoContent,
		http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound,
	), s.closeSession)
	huma.Register(api, apiOperation(
		http.MethodGet, "/api/v1/packages/{id}", "exportPackage", "Export a portable DDivination package",
		[]string{"Portability"}, http.StatusOK,
		http.StatusNotFound, http.StatusInternalServerError,
	), s.exportPackage)
	importPackageOperation := apiOperation(
		http.MethodPost, "/api/v1/packages", "importPackage", "Import a portable DDivination package",
		[]string{"Portability"}, http.StatusCreated,
		http.StatusRequestEntityTooLarge, http.StatusUnprocessableEntity, http.StatusInternalServerError,
	)
	importPackageOperation.MaxBodyBytes = packageio.MaxPackageSize
	huma.Register(api, importPackageOperation, s.importPackage)
	huma.Register(api, apiOperation(
		http.MethodGet, "/api/v1/assets", "listAssets", "List imported assets",
		[]string{"Assets"}, http.StatusOK,
		http.StatusInternalServerError,
	), s.listAssets)
	importAssetOperation := apiOperation(
		http.MethodPost, "/api/v1/assets", "importAsset", "Import a PNG, WebP, or self-contained GLB",
		[]string{"Assets"}, http.StatusCreated,
		http.StatusBadRequest, http.StatusUnprocessableEntity, http.StatusInternalServerError,
	)
	importAssetOperation.MaxBodyBytes = asset.MaxGLBSize + (1 << 20)
	huma.Register(api, importAssetOperation, s.importAsset)
	huma.Register(api, apiOperation(
		http.MethodPost, "/api/v1/ai/enrich", "enrichAdventure", "Enrich an adventure specification with optional AI",
		[]string{"AI"}, http.StatusOK,
		http.StatusBadRequest, http.StatusBadGateway,
	), s.enrichWithAI)

	setResponseContentType(api, http.MethodGet, "/api/v1/packages/{id}", http.StatusOK, "application/vnd.ddivination+zip")
	setResponseContentType(api, http.MethodGet, "/api/v1/adventures/{id}/export.md", http.StatusOK, "text/markdown")
	setResponseContentType(api, http.MethodGet, "/api/v1/adventures/{id}/print", http.StatusOK, "text/html")
	api.OpenAPI().Paths["/api/v1/packages/{id}"].Get.Responses[statusCode(http.StatusOK)].
		Content["application/vnd.ddivination+zip"].Schema = &huma.Schema{Type: "string", Format: "binary"}
	api.OpenAPI().Paths["/api/v1/assets"].Post.RequestBody.Required = true
}

func setResponseContentType(api huma.API, method, path string, status int, mediaType string) {
	pathItem := api.OpenAPI().Paths[path]
	var operation *huma.Operation
	switch method {
	case http.MethodGet:
		operation = pathItem.Get
	case http.MethodPost:
		operation = pathItem.Post
	case http.MethodPut:
		operation = pathItem.Put
	case http.MethodDelete:
		operation = pathItem.Delete
	}
	if operation == nil {
		panic("operation not registered: " + method + " " + path)
	}
	response := operation.Responses[http.StatusText(status)]
	if response == nil {
		response = operation.Responses[statusCode(status)]
	}
	if response == nil {
		panic("response not registered: " + method + " " + path)
	}
	content := response.Content["application/json"]
	delete(response.Content, "application/json")
	response.Content[mediaType] = content
}

func statusCode(status int) string {
	return fmt.Sprintf("%d", status)
}
