// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package providers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewOllamaProvider(t *testing.T) {
	tests := []struct {
		name      string
		modelName string
		expected  string
	}{
		{
			name:      "basic model name",
			modelName: "nomic-embed-text",
			expected:  "nomic-embed-text",
		},
		{
			name:      "model with ollama prefix",
			modelName: "ollama/nomic-embed-text",
			expected:  "nomic-embed-text",
		},
		{
			name:      "model with version tag",
			modelName: "mxbai-embed-large:latest",
			expected:  "mxbai-embed-large:latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewOllamaProvider(tt.modelName)
			if err != nil {
				t.Fatalf("Failed to create provider: %v", err)
			}

			if provider.modelName != tt.expected {
				t.Errorf("Expected model name %s, got %s", tt.expected, provider.modelName)
			}

			if provider.GetModelName() != tt.expected {
				t.Errorf("Expected GetModelName() %s, got %s", tt.expected, provider.GetModelName())
			}

			if provider.GetProvider() != "ollama" {
				t.Errorf("Expected provider 'ollama', got %s", provider.GetProvider())
			}

			if provider.baseURL != "http://localhost:11434" {
				t.Errorf("Expected baseURL 'http://localhost:11434', got %s", provider.baseURL)
			}
		})
	}
}

func TestOllamaProvider_GenerateEmbedding(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/embeddings" {
			t.Errorf("Expected path /api/embeddings, got %s", r.URL.Path)
		}

		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		// Parse request
		var req OllamaEmbeddingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Validate request
		if req.Model != "nomic-embed-text" {
			t.Errorf("Expected model 'nomic-embed-text', got %s", req.Model)
		}

		if req.Prompt == "" {
			t.Error("Expected non-empty prompt")
		}

		// Return mock embedding
		resp := OllamaEmbeddingResponse{
			Embedding: []float64{0.1, 0.2, 0.3, 0.4, 0.5},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider, _ := NewOllamaProvider("nomic-embed-text")
	provider.baseURL = server.URL

	ctx := context.Background()
	embedding, err := provider.GenerateEmbedding(ctx, "test text")

	if err != nil {
		t.Fatalf("GenerateEmbedding failed: %v", err)
	}

	if len(embedding) != 5 {
		t.Errorf("Expected 5 dimensions, got %d", len(embedding))
	}

	expected := []float64{0.1, 0.2, 0.3, 0.4, 0.5}
	for i, v := range embedding {
		if v != expected[i] {
			t.Errorf("Expected embedding[%d] = %f, got %f", i, expected[i], v)
		}
	}
}

func TestOllamaProvider_GenerateEmbedding_EmptyText(t *testing.T) {
	provider, _ := NewOllamaProvider("nomic-embed-text")

	ctx := context.Background()
	_, err := provider.GenerateEmbedding(ctx, "")

	if err == nil {
		t.Error("Expected error for empty text, got nil")
	}

	if !strings.Contains(err.Error(), "cannot be empty") {
		t.Errorf("Expected error about empty text, got: %s", err.Error())
	}
}

func TestOllamaProvider_GenerateEmbedding_ServerError(t *testing.T) {
	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "model not found"}`))
	}))
	defer server.Close()

	provider, _ := NewOllamaProvider("unknown-model")
	provider.baseURL = server.URL

	ctx := context.Background()
	_, err := provider.GenerateEmbedding(ctx, "test text")

	if err == nil {
		t.Error("Expected error for server error, got nil")
	}

	if !strings.Contains(err.Error(), "Ollama API error") {
		t.Errorf("Expected Ollama API error, got: %s", err.Error())
	}
}

func TestOllamaProvider_GenerateEmbeddings(t *testing.T) {
	// Create mock server
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		resp := OllamaEmbeddingResponse{
			Embedding: []float64{float64(callCount), float64(callCount + 1), float64(callCount + 2)},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider, _ := NewOllamaProvider("nomic-embed-text")
	provider.baseURL = server.URL

	ctx := context.Background()
	texts := []string{"text1", "text2", "text3"}
	embeddings, err := provider.GenerateEmbeddings(ctx, texts)

	if err != nil {
		t.Fatalf("GenerateEmbeddings failed: %v", err)
	}

	if len(embeddings) != 3 {
		t.Errorf("Expected 3 embeddings, got %d", len(embeddings))
	}

	if callCount != 3 {
		t.Errorf("Expected 3 API calls, got %d", callCount)
	}
}

func TestOllamaProvider_GenerateEmbeddings_EmptyTexts(t *testing.T) {
	provider, _ := NewOllamaProvider("nomic-embed-text")

	ctx := context.Background()
	embeddings, err := provider.GenerateEmbeddings(ctx, []string{})

	if err != nil {
		t.Errorf("Expected no error for empty texts, got: %v", err)
	}

	if len(embeddings) != 0 {
		t.Errorf("Expected 0 embeddings, got %d", len(embeddings))
	}
}

func TestOllamaProvider_GenerateEmbeddings_WithEmptyText(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := OllamaEmbeddingResponse{
			Embedding: []float64{0.1, 0.2, 0.3},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider, _ := NewOllamaProvider("nomic-embed-text")
	provider.baseURL = server.URL

	ctx := context.Background()
	texts := []string{"text1", "", "text3"} // Middle text is empty
	embeddings, err := provider.GenerateEmbeddings(ctx, texts)

	if err != nil {
		t.Fatalf("GenerateEmbeddings failed: %v", err)
	}

	if len(embeddings) != 3 {
		t.Errorf("Expected 3 embeddings, got %d", len(embeddings))
	}

	// Empty text should produce empty embedding
	if len(embeddings[1]) != 0 {
		t.Errorf("Expected empty embedding for empty text, got %d dimensions", len(embeddings[1]))
	}
}

func TestOllamaProvider_IsAvailable(t *testing.T) {
	tests := []struct {
		name      string
		modelName string
		handler   http.HandlerFunc
		expectErr bool
		errMsg    string
	}{
		{
			name:      "model available",
			modelName: "nomic-embed-text",
			handler: func(w http.ResponseWriter, r *http.Request) {
				resp := struct {
					Models []struct {
						Name string `json:"name"`
					} `json:"models"`
				}{
					Models: []struct {
						Name string `json:"name"`
					}{
						{Name: "nomic-embed-text:latest"},
						{Name: "llama2:latest"},
					},
				}
				json.NewEncoder(w).Encode(resp)
			},
			expectErr: false,
		},
		{
			name:      "model with version tag",
			modelName: "nomic-embed-text:latest",
			handler: func(w http.ResponseWriter, r *http.Request) {
				resp := struct {
					Models []struct {
						Name string `json:"name"`
					} `json:"models"`
				}{
					Models: []struct {
						Name string `json:"name"`
					}{
						{Name: "nomic-embed-text:latest"},
					},
				}
				json.NewEncoder(w).Encode(resp)
			},
			expectErr: false,
		},
		{
			name:      "model not found",
			modelName: "unknown-model",
			handler: func(w http.ResponseWriter, r *http.Request) {
				resp := struct {
					Models []struct {
						Name string `json:"name"`
					} `json:"models"`
				}{
					Models: []struct {
						Name string `json:"name"`
					}{
						{Name: "llama2:latest"},
					},
				}
				json.NewEncoder(w).Encode(resp)
			},
			expectErr: true,
			errMsg:    "not found in Ollama",
		},
		{
			name:      "server error",
			modelName: "nomic-embed-text",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectErr: true,
			errMsg:    "returned status 500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			provider, _ := NewOllamaProvider(tt.modelName)
			provider.baseURL = server.URL

			ctx := context.Background()
			err := provider.IsAvailable(ctx)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error, got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing '%s', got: %s", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

func TestOllamaProvider_IsAvailable_ServerNotRunning(t *testing.T) {
	provider, _ := NewOllamaProvider("nomic-embed-text")
	provider.baseURL = "http://localhost:99999" // Invalid port

	ctx := context.Background()
	err := provider.IsAvailable(ctx)

	if err == nil {
		t.Error("Expected error for server not running, got nil")
	}

	if !strings.Contains(err.Error(), "not running") {
		t.Errorf("Expected error about server not running, got: %s", err.Error())
	}
}
