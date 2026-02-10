// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OllamaProvider implements EmbeddingProvider for Ollama models
type OllamaProvider struct {
	modelName string
	baseURL   string
	client    *http.Client
}

// OllamaEmbeddingRequest represents the request to Ollama API
type OllamaEmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// OllamaEmbeddingResponse represents the response from Ollama API
type OllamaEmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(modelName string) (*OllamaProvider, error) {
	// Strip ollama/ prefix if present
	shortName := strings.TrimPrefix(modelName, "ollama/")

	return &OllamaProvider{
		modelName: shortName,
		baseURL:   "http://localhost:11434", // Default Ollama server
		client: &http.Client{
			Timeout: 30 * time.Second, // Longer timeout for embeddings
		},
	}, nil
}

// GenerateEmbedding generates an embedding for a single text
func (p *OllamaProvider) GenerateEmbedding(ctx context.Context, text string) ([]float64, error) {
	if text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	// Prepare request
	reqBody := OllamaEmbeddingRequest{
		Model:  p.modelName,
		Prompt: text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to Ollama: %w\n\nMake sure Ollama is running:\n  ollama serve\n\nOr check if model is installed:\n  ollama list", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama API error (status %d): %s\n\nMake sure model '%s' is installed:\n  ollama pull %s", resp.StatusCode, string(body), p.modelName, p.modelName)
	}

	// Parse response
	var ollamaResp OllamaEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(ollamaResp.Embedding) == 0 {
		return nil, fmt.Errorf("no embedding returned from Ollama")
	}

	return ollamaResp.Embedding, nil
}

// GenerateEmbeddings generates embeddings for multiple texts (batch)
func (p *OllamaProvider) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float64, error) {
	if len(texts) == 0 {
		return [][]float64{}, nil
	}

	embeddings := make([][]float64, len(texts))

	// Process each text individually
	// Note: Ollama doesn't support batch embeddings in a single request
	for i, text := range texts {
		if text == "" {
			// Create zero vector for empty texts
			embeddings[i] = []float64{}
			continue
		}

		embedding, err := p.GenerateEmbedding(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("failed to generate embedding for text %d: %w", i, err)
		}
		embeddings[i] = embedding
	}

	return embeddings, nil
}

// GetModelName returns the model name
func (p *OllamaProvider) GetModelName() string {
	return p.modelName
}

// GetProvider returns the provider name
func (p *OllamaProvider) GetProvider() string {
	return "ollama"
}

// IsAvailable checks if Ollama is running and the model is available
func (p *OllamaProvider) IsAvailable(ctx context.Context) error {
	// Check if Ollama server is running
	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("Ollama server not running at %s\n\nTo start Ollama:\n  ollama serve\n\nOr install from: https://ollama.ai", p.baseURL)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ollama server returned status %d", resp.StatusCode)
	}

	// Parse response to check if model exists
	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to parse Ollama response: %w", err)
	}

	// Check if our model is in the list
	modelFound := false
	for _, model := range result.Models {
		// Match exact name or name with :latest tag
		if model.Name == p.modelName || model.Name == p.modelName+":latest" || strings.HasPrefix(model.Name, p.modelName+":") {
			modelFound = true
			break
		}
	}

	if !modelFound {
		return fmt.Errorf("model '%s' not found in Ollama\n\nTo install:\n  ollama pull %s\n\nSee available models: ollama list", p.modelName, p.modelName)
	}

	return nil
}
