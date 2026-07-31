package main

import (
	"io"
	"log/slog"
	"net/http/httptest"
	"os"

	"github.com/EmanuelSena101/DDivination/apps/server/internal/api"
)

func main() {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := api.NewContractServer(logger)
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
