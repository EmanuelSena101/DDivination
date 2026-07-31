package store

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/EmanuelSena101/DDivination/apps/server/internal/domain"
)

func TestDecodeSessionHeadRejectsSemanticCorruption(t *testing.T) {
	valid := domain.SessionState{ID: "session-a", AdventureID: "adventure-a", Revision: 7}
	payload, err := json.Marshal(valid)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := decodeSessionHead(valid.ID, valid.AdventureID, valid.Revision, payload); err != nil {
		t.Fatalf("valid durable head was rejected: %v", err)
	}

	checks := []struct {
		name        string
		sessionID   string
		adventureID string
		revision    int64
		payload     []byte
	}{
		{name: "invalid json", sessionID: valid.ID, adventureID: valid.AdventureID, revision: valid.Revision, payload: []byte("{")},
		{name: "session mismatch", sessionID: "session-b", adventureID: valid.AdventureID, revision: valid.Revision, payload: payload},
		{name: "adventure mismatch", sessionID: valid.ID, adventureID: "adventure-b", revision: valid.Revision, payload: payload},
		{name: "revision mismatch", sessionID: valid.ID, adventureID: valid.AdventureID, revision: 8, payload: payload},
	}
	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			_, err := decodeSessionHead(check.sessionID, check.adventureID, check.revision, check.payload)
			if !errors.Is(err, ErrCorruptSession) {
				t.Fatalf("expected ErrCorruptSession, got %v", err)
			}
		})
	}
}
