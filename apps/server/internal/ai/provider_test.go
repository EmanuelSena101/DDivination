package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EmanuelSena101/DDivination/apps/server/internal/domain"
)

func TestOpenAIEnrichmentUsesStructuredOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer secret" {
			t.Error("missing bearer authentication")
		}
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		text := request["text"].(map[string]any)
		format := text["format"].(map[string]any)
		if format["type"] != "json_schema" || format["strict"] != true || request["store"] != false {
			t.Fatalf("unexpected request: %#v", request)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"output": [{"type":"message","content":[{"type":"output_text","text":"{\"hook\":{\"pt-BR\":\"Chamado\",\"en-US\":\"Call\"},\"objective\":{\"pt-BR\":\"Objetivo\",\"en-US\":\"Objective\"},\"antagonist\":{\"pt-BR\":\"Vilão\",\"en-US\":\"Villain\"},\"atmosphere\":{\"pt-BR\":\"Sombria\",\"en-US\":\"Dark\"}}"}]}],
			"usage": {"input_tokens": 120, "output_tokens": 40}
		}`))
	}))
	defer server.Close()

	provider := NewOpenAI("secret", "test-model")
	provider.baseURL = server.URL
	provider.client = server.Client()
	result, err := provider.Enrich(context.Background(), domain.DefaultAdventureSpec())
	if err != nil {
		t.Fatal(err)
	}
	if result.Enrichment.Hook.PTBR != "Chamado" || result.Diagnostics.OutputTokens != 40 {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestOpenAIRequiresInMemoryKey(t *testing.T) {
	_, err := NewOpenAI("", "").Enrich(context.Background(), domain.DefaultAdventureSpec())
	if err == nil {
		t.Fatal("expected missing key to fail without an HTTP request")
	}
}
