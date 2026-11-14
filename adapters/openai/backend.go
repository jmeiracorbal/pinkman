package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jmeiracorbal/pinkman"
)

type Config struct {
	ID           string
	BaseURL      string
	APIKey       string
	Capabilities []pinkman.Capability
	Priority     int
	HTTPClient   *http.Client
}

type Backend struct {
	config Config
	client *http.Client
}

func New(config Config) (*Backend, error) {
	if config.ID == "" {
		return nil, fmt.Errorf("openai adapter: id is required")
	}
	if config.BaseURL == "" {
		return nil, fmt.Errorf("openai adapter: base URL is required")
	}
	client := config.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	return &Backend{config: config, client: client}, nil
}

func (b *Backend) ID() string { return b.config.ID }

func (b *Backend) Discover(ctx context.Context) ([]pinkman.Target, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(b.config.BaseURL, "/")+"/models", nil)
	if err != nil {
		return nil, err
	}
	b.authorize(req)
	res, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		return nil, fmt.Errorf("openai adapter: models endpoint returned %s: %s", res.Status, strings.TrimSpace(string(body)))
	}
	var payload struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("openai adapter: decode models: %w", err)
	}
	targets := make([]pinkman.Target, 0, len(payload.Data))
	for _, model := range payload.Data {
		if strings.TrimSpace(model.ID) == "" {
			continue
		}
		targets = append(targets, pinkman.Target{
			Backend: b.config.ID, Model: model.ID, Transport: pinkman.TransportOpenAI,
			Capabilities: append([]pinkman.Capability(nil), b.config.Capabilities...), Priority: b.config.Priority,
		})
	}
	return targets, nil
}

func (b *Backend) Execute(ctx context.Context, invocation pinkman.Invocation) (pinkman.Result, error) {
	payload, err := json.Marshal(map[string]any{"model": invocation.Target.Model, "input": invocation.Input})
	if err != nil {
		return pinkman.Result{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(b.config.BaseURL, "/")+"/responses", bytes.NewReader(payload))
	if err != nil {
		return pinkman.Result{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	b.authorize(req)
	res, err := b.client.Do(req)
	if err != nil {
		return pinkman.Result{}, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return pinkman.Result{}, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return pinkman.Result{}, fmt.Errorf("openai adapter: responses endpoint returned %s: %s", res.Status, strings.TrimSpace(string(body)))
	}
	return pinkman.Result{Output: extractOutput(body), Raw: body}, nil
}

func (b *Backend) authorize(req *http.Request) {
	if b.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+b.config.APIKey)
	}
}

func extractOutput(body []byte) string {
	var payload struct {
		OutputText string `json:"output_text"`
	}
	if json.Unmarshal(body, &payload) == nil && payload.OutputText != "" {
		return payload.OutputText
	}
	return string(body)
}
