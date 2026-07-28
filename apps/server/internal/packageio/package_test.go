package packageio

import (
	"archive/zip"
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/EmanuelSena101/DDivination/apps/server/internal/domain"
	"github.com/EmanuelSena101/DDivination/apps/server/internal/generator"
)

func TestRoundTrip(t *testing.T) {
	doc, err := generator.Generate(domain.DefaultAdventureSpec(), 99, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	content, err := Export(doc)
	if err != nil {
		t.Fatal(err)
	}
	imported, err := Import(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		t.Fatal(err)
	}
	if imported.ID != doc.ID || imported.Seed != doc.Seed || imported.Version != doc.Version+1 {
		t.Fatalf("round-trip changed identity: %#v", imported)
	}
}

func TestRejectsPathTraversal(t *testing.T) {
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	entry, err := writer.Create("../adventure.json")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = entry.Write([]byte("{}"))
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	_, err = Import(bytes.NewReader(buffer.Bytes()), int64(buffer.Len()))
	if !errors.Is(err, ErrInvalidPackage) {
		t.Fatalf("expected ErrInvalidPackage, got %v", err)
	}
}
