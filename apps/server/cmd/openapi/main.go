package main

import (
	"io"
	"log/slog"
	"net/http/httptest"
	"os"

	"github.com/EmanuelSena101/DDivination/apps/server/internal/api"
	"github.com/EmanuelSena101/DDivination/apps/server/internal/session"
	"github.com/EmanuelSena101/DDivination/apps/server/internal/store"
)

func main() {
	temp, err := os.CreateTemp("", "ddivination-openapi-*.sqlite3")
	if err != nil {
		panic(err)
	}
	path := temp.Name()
	if err := temp.Close(); err != nil {
		panic(err)
	}
	if err := os.Remove(path); err != nil {
		panic(err)
	}
	defer os.Remove(path)

	database, err := store.Open(path)
	if err != nil {
		panic(err)
	}
	defer database.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := api.New(database, session.NewHub(database), nil, logger, "", nil, os.TempDir())
	request := httptest.NewRequest("GET", "/openapi.json", nil)
	response := httptest.NewRecorder()
	server.Handlers().Local.ServeHTTP(response, request)
	if response.Code != 200 {
		panic(response.Body.String())
	}
	if _, err := os.Stdout.Write(response.Body.Bytes()); err != nil {
		panic(err)
	}
}
