// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package ollama

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client is a lightweight client for discovering Ollama models
type Client struct {
	baseURL string
	client  *http.Client
}

// Model represents an Ollama model
type Model struct {
	Name       string    `json:"name"`
	ModifiedAt time.Time `json:"modified_at"`
	Size       int64     `json:"size"`
	Digest     string    `json:"digest"`
	Details    Details   `json:"details"`
}

// Details contains model metadata
type Details struct {
	Format            string   `json:"format"`
	Family            string   `json:"family"`
	Families          []string `json:"families"`
	ParameterSize     string   `json:"parameter_size"`
	QuantizationLevel string   `json:"quantization_level"`
}

// TagsResponse is the response from /api/tags endpoint
type TagsResponse struct {
	Models []Model `json:"models"`
}

// EmbeddingModel represents a discovered embedding model
type EmbeddingModel struct {
	Name       string
	Dimensions int
	Available  bool
}

// NewClient creates a new Ollama client
func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	return &Client{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// IsAvailable checks if Ollama server is running
func (c *Client) IsAvailable(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/tags", nil)
	if err != nil {
		return false
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// ListModels lists all available Ollama models
func (c *Client) ListModels(ctx context.Context) ([]Model, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ollama returned status %d", resp.StatusCode)
	}

	var tagsResp TagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tagsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return tagsResp.Models, nil
}

// DiscoverEmbeddingModels finds available embedding models
func (c *Client) DiscoverEmbeddingModels(ctx context.Context) ([]EmbeddingModel, error) {
	models, err := c.ListModels(ctx)
	if err != nil {
		return nil, err
	}

	// Known Ollama embedding models and their dimensions
	knownEmbeddings := map[string]int{
		"nomic-embed-text":       768,
		"mxbai-embed-large":      1024,
		"snowflake-arctic-embed": 1024,
		"all-minilm":             384,
		"nomic-embed":            768, // Alias for nomic-embed-text
	}

	var embeddingModels []EmbeddingModel
	for _, model := range models {
		// Check if model name contains known embedding model names
		for knownName, dims := range knownEmbeddings {
			if contains(model.Name, knownName) {
				embeddingModels = append(embeddingModels, EmbeddingModel{
					Name:       model.Name,
					Dimensions: dims,
					Available:  true,
				})
				break
			}
		}
	}

	return embeddingModels, nil
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
				findSubstring(s, substr))))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
