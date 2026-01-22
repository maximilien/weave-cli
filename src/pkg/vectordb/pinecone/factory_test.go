// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package pinecone

import (
	"strings"
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

	if types[0] != vectordb.VectorDBTypePinecone {
		t.Errorf("Expected VectorDBTypePinecone, got %v", types[0])
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
				Type:   vectordb.VectorDBTypePinecone,
				APIKey: "test-api-key",
			},
			shouldError: false,
		},
		{
			name: "Valid config with URL",
			config: &vectordb.Config{
				Type:   vectordb.VectorDBTypePinecone,
				APIKey: "test-api-key",
				URL:    "https://my-index-12345.svc.pinecone.io",
			},
			shouldError: false,
		},
		{
			name: "Wrong database type",
			config: &vectordb.Config{
				Type:   vectordb.VectorDBTypeWeaviateCloud,
				APIKey: "test-key",
			},
			shouldError: true,
			errorMsg:    "invalid config type",
		},
		{
			name: "Missing API key",
			config: &vectordb.Config{
				Type:   vectordb.VectorDBTypePinecone,
				APIKey: "",
			},
			shouldError: true,
			errorMsg:    "API key is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := factory.ValidateConfig(tt.config)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error containing %q, got nil", tt.errorMsg)
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
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

	tests := []struct {
		name        string
		config      *vectordb.Config
		shouldError bool
		errorMsg    string
	}{
		{
			name: "Valid config",
			config: &vectordb.Config{
				Type:   vectordb.VectorDBTypePinecone,
				APIKey: "test-api-key",
			},
			shouldError: false,
		},
		{
			name: "Wrong type",
			config: &vectordb.Config{
				Type:   vectordb.VectorDBTypeMock,
				APIKey: "test-key",
			},
			shouldError: true,
			errorMsg:    "invalid config type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := factory.CreateClient(tt.config)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error containing %q, got nil", tt.errorMsg)
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
				if client != nil {
					t.Error("Expected nil client on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if client == nil {
					t.Error("Expected non-nil client")
				}
			}
		})
	}
}
