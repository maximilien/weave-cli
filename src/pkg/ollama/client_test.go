// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		expectedURL string
	}{
		{
			name:        "default URL",
			baseURL:     "",
			expectedURL: "http://localhost:11434",
		},
		{
			name:        "custom URL",
			baseURL:     "http://custom:8080",
			expectedURL: "http://custom:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.baseURL)
			if client.baseURL != tt.expectedURL {
				t.Errorf("Expected baseURL %s, got %s", tt.expectedURL, client.baseURL)
			}
		})
	}
}

func TestIsAvailable(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		expected   bool
	}{
		{
			name:       "server available",
			statusCode: http.StatusOK,
			expected:   true,
		},
		{
			name:       "server unavailable",
			statusCode: http.StatusInternalServerError,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(TagsResponse{Models: []Model{}})
			}))
			defer server.Close()

			client := NewClient(server.URL)
			ctx := context.Background()

			available := client.IsAvailable(ctx)
			if available != tt.expected {
				t.Errorf("Expected IsAvailable %v, got %v", tt.expected, available)
			}
		})
	}
}

func TestListModels(t *testing.T) {
	mockModels := []Model{
		{
			Name:       "nomic-embed-text:latest",
			ModifiedAt: time.Now(),
			Size:       274301056,
			Digest:     "0a109f422b47a6acf013916cafa4f3f2e0f3f0a5da64c82ebd91a82e3ff0fbb4",
			Details: Details{
				Format:            "gguf",
				Family:            "nomic-bert",
				ParameterSize:     "137M",
				QuantizationLevel: "F16",
			},
		},
		{
			Name:       "llama2:latest",
			ModifiedAt: time.Now(),
			Size:       3826793677,
			Digest:     "78e26419b4469263f75331927a00a0284ef6544c1975b826b15abdaef17bb962",
			Details: Details{
				Format:            "gguf",
				Family:            "llama",
				ParameterSize:     "7B",
				QuantizationLevel: "Q4_0",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tags" {
			t.Errorf("Expected /api/tags, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(TagsResponse{Models: mockModels})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	ctx := context.Background()

	models, err := client.ListModels(ctx)
	if err != nil {
		t.Fatalf("ListModels failed: %v", err)
	}

	if len(models) != 2 {
		t.Errorf("Expected 2 models, got %d", len(models))
	}

	if models[0].Name != "nomic-embed-text:latest" {
		t.Errorf("Expected first model nomic-embed-text:latest, got %s", models[0].Name)
	}
}

func TestDiscoverEmbeddingModels(t *testing.T) {
	mockModels := []Model{
		{Name: "nomic-embed-text:latest"},
		{Name: "mxbai-embed-large:latest"},
		{Name: "llama2:latest"}, // Not an embedding model
		{Name: "snowflake-arctic-embed:latest"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(TagsResponse{Models: mockModels})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	ctx := context.Background()

	embeddings, err := client.DiscoverEmbeddingModels(ctx)
	if err != nil {
		t.Fatalf("DiscoverEmbeddingModels failed: %v", err)
	}

	// Should find 3 embedding models (nomic-embed-text, mxbai-embed-large, snowflake-arctic-embed)
	// llama2 should be filtered out
	if len(embeddings) != 3 {
		t.Errorf("Expected 3 embedding models, got %d", len(embeddings))
	}

	// Check nomic-embed-text
	found := false
	for _, em := range embeddings {
		if em.Name == "nomic-embed-text:latest" && em.Dimensions == 768 {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find nomic-embed-text:latest with 768 dimensions")
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{"exact match", "nomic-embed-text", "nomic-embed-text", true},
		{"contains at start", "nomic-embed-text:latest", "nomic-embed-text", true},
		{"contains at end", "model-nomic-embed", "nomic-embed", true},
		{"does not contain", "llama2", "nomic-embed", false},
		{"substring in middle", "my-nomic-embed-text-model", "nomic-embed-text", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.s, tt.substr)
			if result != tt.expected {
				t.Errorf("contains(%s, %s) = %v, expected %v", tt.s, tt.substr, result, tt.expected)
			}
		})
	}
}
