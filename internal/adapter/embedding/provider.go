package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/zk-org/zk/internal/core"
)

// Provider embeds user text into dense vectors.
type Provider interface {
	EmbedTexts(ctx context.Context, texts []string) ([][]float32, error)
	EmbedQuery(ctx context.Context, text string) ([]float32, error)
}

// NewProvider creates a new embedding provider from config.
func NewProvider(cfg core.EmbeddingConfig) (Provider, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	switch cfg.Provider {
	case "openai":
		return newOpenAIProvider(cfg, "https://api.openai.com/v1")
	case "local":
		if cfg.Endpoint == "" {
			return nil, fmt.Errorf("embedding.endpoint is required when embedding.provider = \"local\"")
		}
		return newOpenAIProvider(cfg, cfg.Endpoint)
	case "googleai":
		return newGoogleAIProvider(cfg)
	default:
		return nil, fmt.Errorf("%s: unknown embedding.provider", cfg.Provider)
	}
}

func defaultAPIKeyEnv(provider string) string {
	switch provider {
	case "openai":
		return "ZK_EMBEDDING_OPENAI_API_KEY"
	case "googleai":
		return "ZK_EMBEDDING_GOOGLE_API_KEY"
	default:
		return ""
	}
}

type openAIProvider struct {
	endpoint   string
	model      string
	dimensions int
	apiKey     string
	client     *http.Client
}

func newOpenAIProvider(cfg core.EmbeddingConfig, endpoint string) (Provider, error) {
	apiKey := ""
	keyEnv := cfg.APIKeyEnv
	if keyEnv == "" {
		keyEnv = defaultAPIKeyEnv(cfg.Provider)
	}
	if keyEnv != "" {
		apiKey = os.Getenv(keyEnv)
		if apiKey == "" {
			return nil, fmt.Errorf("embedding API key env %s is not set", keyEnv)
		}
	}

	return &openAIProvider{
		endpoint:   strings.TrimRight(endpoint, "/"),
		model:      cfg.Model,
		dimensions: cfg.Dimensions,
		apiKey:     apiKey,
		client:     &http.Client{Timeout: 60 * time.Second},
	}, nil
}

func (p *openAIProvider) EmbedTexts(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	type reqBody struct {
		Model      string   `json:"model"`
		Input      []string `json:"input"`
		Dimensions int      `json:"dimensions,omitempty"`
	}
	body := reqBody{Model: p.model, Input: texts}
	if p.dimensions > 0 {
		body.Dimensions = p.dimensions
	}

	var resp struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error,omitempty"`
	}

	err := p.doJSON(ctx, "POST", p.endpoint+"/embeddings", body, &resp)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("embedding request failed: %s", resp.Error.Message)
	}

	vectors := make([][]float32, 0, len(resp.Data))
	for _, item := range resp.Data {
		vectors = append(vectors, item.Embedding)
	}
	if len(vectors) != len(texts) {
		return nil, fmt.Errorf("embedding response length mismatch: got %d vectors for %d texts", len(vectors), len(texts))
	}

	return vectors, nil
}

func (p *openAIProvider) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	res, err := p.EmbedTexts(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, fmt.Errorf("embedding response is empty")
	}
	return res[0], nil
}

func (p *openAIProvider) doJSON(ctx context.Context, method, url string, body any, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var e struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&e); err == nil && e.Error.Message != "" {
			return fmt.Errorf("%s: %s", resp.Status, e.Error.Message)
		}
		return fmt.Errorf("embedding request failed: %s", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return err
	}
	return nil
}

type googleAIProvider struct {
	model  string
	apiKey string
	client *http.Client
}

func newGoogleAIProvider(cfg core.EmbeddingConfig) (Provider, error) {
	keyEnv := cfg.APIKeyEnv
	if keyEnv == "" {
		keyEnv = defaultAPIKeyEnv("googleai")
	}
	apiKey := os.Getenv(keyEnv)
	if apiKey == "" {
		return nil, fmt.Errorf("embedding API key env %s is not set", keyEnv)
	}

	return &googleAIProvider{
		model:  cfg.Model,
		apiKey: apiKey,
		client: &http.Client{Timeout: 60 * time.Second},
	}, nil
}

func (p *googleAIProvider) EmbedTexts(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	vectors := make([][]float32, 0, len(texts))
	for _, text := range texts {
		vec, err := p.EmbedQuery(ctx, text)
		if err != nil {
			return nil, err
		}
		vectors = append(vectors, vec)
	}
	return vectors, nil
}

func (p *googleAIProvider) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	type reqBody struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	}
	var body reqBody
	body.Content.Parts = []struct {
		Text string `json:"text"`
	}{{Text: text}}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:embedContent?key=%s", p.model, p.apiKey)
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("embedding request failed: %s", resp.Status)
	}

	var out struct {
		Embedding struct {
			Values []float32 `json:"values"`
		} `json:"embedding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if len(out.Embedding.Values) == 0 {
		return nil, fmt.Errorf("embedding response is empty")
	}
	return out.Embedding.Values, nil
}
