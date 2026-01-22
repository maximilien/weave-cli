// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package mock

import (
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

func TestNewFactory(t *testing.T) {
	factory := NewFactory()
	if factory == nil {
		t.Fatal("Expected non-nil factory")
	}
}

func TestGetSupportedTypes(t *testing.T) {
	factory := NewFactory()
	types := factory.GetSupportedTypes()

	if len(types) != 1 {
		t.Errorf("Expected 1 supported type, got %d", len(types))
	}

	if types[0] != vectordb.VectorDBTypeMock {
		t.Errorf("Expected VectorDBTypeMock, got %v", types[0])
	}
}

func TestValidateConfig(t *testing.T) {
	factory := NewFactory()

	tests := []struct {
		name        string
		config      *vectordb.Config
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "Nil config",
			config:      nil,
			shouldError: true,
			errorMsg:    "config cannot be nil",
		},
		{
			name: "Valid config",
			config: &vectordb.Config{
				Type:               vectordb.VectorDBTypeMock,
				EmbeddingDimension: 384,
			},
			shouldError: false,
		},
		{
			name: "Wrong database type",
			config: &vectordb.Config{
				Type:               vectordb.VectorDBTypeWeaviateCloud,
				EmbeddingDimension: 384,
			},
			shouldError: true,
			errorMsg:    "unsupported database type",
		},
		{
			name: "Negative embedding dimension",
			config: &vectordb.Config{
				Type:               vectordb.VectorDBTypeMock,
				EmbeddingDimension: -1,
			},
			shouldError: true,
			errorMsg:    "embedding dimension cannot be negative",
		},
		{
			name: "Zero embedding dimension (valid)",
			config: &vectordb.Config{
				Type:               vectordb.VectorDBTypeMock,
				EmbeddingDimension: 0,
			},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := factory.ValidateConfig(tt.config)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error containing %q, got nil", tt.errorMsg)
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

func TestCreateClient(t *testing.T) {
	factory := NewFactory()

	config := &vectordb.Config{
		Type:               vectordb.VectorDBTypeMock,
		Enabled:            true,
		EmbeddingDimension: 384,
		SimulateEmbeddings: true,
	}

	client, err := factory.CreateClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	// Verify it's an Adapter
	adapter, ok := client.(*Adapter)
	if !ok {
		t.Fatalf("Expected *Adapter, got %T", client)
	}

	if adapter.client == nil {
		t.Error("Expected non-nil mock client")
	}

	if adapter.config != config {
		t.Error("Expected adapter to store config reference")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
