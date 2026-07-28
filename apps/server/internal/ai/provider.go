package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/EmanuelSena101/DDivination/apps/server/internal/domain"
)

var (
	ErrUnavailable     = errors.New("AI provider unavailable")
	ErrInvalidResponse = errors.New("AI provider returned an invalid response")
)

type Enrichment = domain.AdventureNarrative

type Diagnostics struct {
	Provider     string        `json:"provider"`
	Model        string        `json:"model"`
	InputTokens  int           `json:"inputTokens"`
	OutputTokens int           `json:"outputTokens"`
	Latency      time.Duration `json:"latency"`
}

type Result struct {
	Enrichment  Enrichment  `json:"enrichment"`
	Diagnostics Diagnostics `json:"diagnostics"`
}

type Provider interface {
	Enrich(context.Context, domain.AdventureSpec) (Result, error)
}

type OpenAI struct {
	key     string
	model   string
	baseURL string
	client  *http.Client
}

func NewOpenAI(key, model string) *OpenAI {
	if strings.TrimSpace(model) == "" {
		model = "gpt-5.6-terra"
	}
	return &OpenAI{
		key:     strings.TrimSpace(key),
		model:   model,
		baseURL: "https://api.openai.com/v1",
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

func (provider *OpenAI) Enrich(ctx context.Context, spec domain.AdventureSpec) (Result, error) {
	if provider.key == "" {
		return Result{}, fmt.Errorf("%w: API key is missing", ErrUnavailable)
	}
	specJSON, err := json.Marshal(spec)
	if err != nil {
		return Result{}, err
	}
	body := map[string]any{
		"model": provider.model,
		"store": false,
		"reasoning": map[string]any{
			"effort": "low",
		},
		"input": []map[string]any{
			{
				"role": "developer",
				"content": []map[string]string{{
					"type": "input_text",
					"text": "Enrich this 5E-compatible adventure specification. Return concise original narrative text in both pt-BR and en-US. Do not add copyrighted setting lore or change mechanical facts.",
				}},
			},
			{
				"role": "user",
				"content": []map[string]string{{
					"type": "input_text",
					"text": string(specJSON),
				}},
			},
		},
		"text": map[string]any{
			"verbosity": "low",
			"format": map[string]any{
				"type":   "json_schema",
				"name":   "adventure_enrichment",
				"strict": true,
				"schema": enrichmentSchema(),
			},
		},
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		return Result{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, provider.baseURL+"/responses", bytes.NewReader(encoded))
	if err != nil {
		return Result{}, err
	}
	request.Header.Set("Authorization", "Bearer "+provider.key)
	request.Header.Set("Content-Type", "application/json")
	started := time.Now()
	response, err := provider.client.Do(request)
	if err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer response.Body.Close()
	content, err := io.ReadAll(io.LimitReader(response.Body, 2<<20))
	if err != nil {
		return Result{}, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return Result{}, fmt.Errorf("%w: HTTP %d", ErrUnavailable, response.StatusCode)
	}
	var decoded responseEnvelope
	if err := json.Unmarshal(content, &decoded); err != nil {
		return Result{}, fmt.Errorf("%w: malformed envelope", ErrInvalidResponse)
	}
	outputText := decoded.OutputText()
	var enrichment Enrichment
	if outputText == "" || json.Unmarshal([]byte(outputText), &enrichment) != nil ||
		!validLocalized(enrichment.Hook) ||
		!validLocalized(enrichment.Objective) ||
		!validLocalized(enrichment.Antagonist) ||
		!validLocalized(enrichment.Atmosphere) {
		return Result{}, ErrInvalidResponse
	}
	return Result{
		Enrichment: enrichment,
		Diagnostics: Diagnostics{
			Provider:     "openai",
			Model:        provider.model,
			InputTokens:  decoded.Usage.InputTokens,
			OutputTokens: decoded.Usage.OutputTokens,
			Latency:      time.Since(started),
		},
	}, nil
}

type responseEnvelope struct {
	Output []struct {
		Type    string `json:"type"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func (response responseEnvelope) OutputText() string {
	for _, output := range response.Output {
		if output.Type != "message" {
			continue
		}
		for _, content := range output.Content {
			if content.Type == "output_text" {
				return content.Text
			}
		}
	}
	return ""
}

func enrichmentSchema() map[string]any {
	localized := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"pt-BR": map[string]any{"type": "string"},
			"en-US": map[string]any{"type": "string"},
		},
		"required": []string{"pt-BR", "en-US"},
	}
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"hook":       localized,
			"objective":  localized,
			"antagonist": localized,
			"atmosphere": localized,
		},
		"required": []string{"hook", "objective", "antagonist", "atmosphere"},
	}
}

func validLocalized(value domain.LocalizedText) bool {
	return strings.TrimSpace(value.PTBR) != "" && strings.TrimSpace(value.ENUS) != ""
}
