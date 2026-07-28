package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/EmanuelSena101/DDivination/apps/server/internal/api"
	"github.com/EmanuelSena101/DDivination/apps/server/internal/network"
	"github.com/EmanuelSena101/DDivination/apps/server/internal/session"
	"github.com/EmanuelSena101/DDivination/apps/server/internal/store"
	"github.com/EmanuelSena101/DDivination/apps/server/internal/webapp"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	dataDir := applicationDataDir()
	dbPath := filepath.Join(dataDir, "ddivination.sqlite3")
	db, err := store.Open(dbPath)
	if err != nil {
		logger.Error("database startup failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	hub := session.NewHub(db)
	lan := network.NewManager(8080, logger)
	defer lan.Close()
	embeddedWeb, err := webapp.FS()
	if err != nil {
		logger.Error("embedded frontend startup failed", "error", err)
		os.Exit(1)
	}
	serverAPI := api.New(
		db,
		hub,
		lan,
		logger,
		os.Getenv("DDIVINATION_WEB_DIR"),
		embeddedWeb,
		filepath.Join(dataDir, "assets"),
	)
	handlers := serverAPI.Handlers()

	httpServer := &http.Server{
		Addr:              "127.0.0.1:8080",
		Handler:           handlers.Local,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errs := make(chan error, 1)
	go func() {
		logger.Info("DDivination started", "url", "http://127.0.0.1:8080", "dataDir", dataDir)
		errs <- httpServer.ListenAndServe()
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	select {
	case signal := <-signals:
		logger.Info("shutdown requested", "signal", signal.String())
	case serveErr := <-errs:
		if !errors.Is(serveErr, http.ErrServerClosed) {
			logger.Error("server stopped unexpectedly", "error", serveErr)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(ctx)
}

func applicationDataDir() string {
	if configured := os.Getenv("DDIVINATION_DATA_DIR"); configured != "" {
		_ = os.MkdirAll(configured, 0o755)
		return configured
	}
	base, err := os.UserConfigDir()
	if err != nil {
		base = "."
	}
	path := filepath.Join(base, "DDivination")
	if err := os.MkdirAll(path, 0o755); err != nil {
		panic(fmt.Sprintf("create application data directory: %v", err))
	}
	return path
}
