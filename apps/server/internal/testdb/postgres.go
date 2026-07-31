package testdb

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/EmanuelSena101/DDivination/apps/server/internal/store"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Open gives a test an isolated PostgreSQL schema. Tests that need persistence
// are skipped when TEST_DATABASE_URL is absent; CI has a dedicated PostgreSQL
// integration job where it is always configured.
func Open(t testing.TB) *store.Store {
	t.Helper()
	database, _ := OpenWithURL(t)
	return database
}

// OpenWithURL also returns the schema-scoped URL so migration/restart tests can
// open an independent pool against exactly the same isolated database state.
func OpenWithURL(t testing.TB) (*store.Store, string) {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not configured")
	}
	parsed, err := url.Parse(databaseURL)
	if err != nil || (parsed.Scheme != "postgres" && parsed.Scheme != "postgresql") {
		t.Fatalf("TEST_DATABASE_URL must be a PostgreSQL URL: %v", err)
	}
	suffix := make([]byte, 8)
	if _, err := rand.Read(suffix); err != nil {
		t.Fatal(err)
	}
	schema := "ddivination_test_" + hex.EncodeToString(suffix)

	admin, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if _, err := admin.ExecContext(ctx, fmt.Sprintf(`CREATE SCHEMA "%s"`, schema)); err != nil {
		_ = admin.Close()
		t.Fatalf("create test schema: %v", err)
	}
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	database, err := store.Open(parsed.String())
	if err != nil {
		_, _ = admin.ExecContext(context.Background(), fmt.Sprintf(`DROP SCHEMA "%s" CASCADE`, schema))
		_ = admin.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = database.Close()
		dropCtx, dropCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer dropCancel()
		_, _ = admin.ExecContext(dropCtx, fmt.Sprintf(`DROP SCHEMA "%s" CASCADE`, schema))
		_ = admin.Close()
	})
	return database, parsed.String()
}
