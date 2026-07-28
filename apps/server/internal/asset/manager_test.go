package asset

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/EmanuelSena101/DDivination/apps/server/internal/domain"
)

type memoryStore struct {
	ref  domain.AssetRef
	path string
}

func (s *memoryStore) SaveAsset(_ context.Context, ref domain.AssetRef, path string) error {
	s.ref, s.path = ref, path
	return nil
}

func TestImportsPNGByContentHash(t *testing.T) {
	var content bytes.Buffer
	source := image.NewRGBA(image.Rect(0, 0, 4, 4))
	source.Set(1, 1, color.RGBA{R: 120, G: 80, B: 240, A: 255})
	if err := png.Encode(&content, source); err != nil {
		t.Fatal(err)
	}
	database := &memoryStore{}
	manager := NewManager(t.TempDir(), database)
	ref, err := manager.Import(context.Background(), "token.png", &content, "Test Artist", "CC0-1.0")
	if err != nil {
		t.Fatal(err)
	}
	if ref.Kind != "image" || ref.MediaType != "image/png" || ref.SHA256 == "" {
		t.Fatalf("unexpected asset metadata: %#v", ref)
	}
	if database.path == "" {
		t.Fatal("asset path was not persisted")
	}
}

func TestRejectsUnsupportedFile(t *testing.T) {
	manager := NewManager(t.TempDir(), &memoryStore{})
	_, err := manager.Import(context.Background(), "payload.exe", bytes.NewBufferString("not an asset"), "", "")
	if err == nil {
		t.Fatal("expected unsupported file to be rejected")
	}
}
