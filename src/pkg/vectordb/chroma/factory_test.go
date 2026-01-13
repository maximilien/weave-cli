//go:build (darwin && amd64) || (darwin && arm64)

// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package chroma

import (
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/assert"
)

func TestFactory_GetSupportedTypes(t *testing.T) {
	factory := NewFactory()
	types := factory.GetSupportedTypes()

	assert.Len(t, types, 2, "Should support 2 types")
	assert.Contains(t, types, vectordb.VectorDBTypeChromaLocal)
	assert.Contains(t, types, vectordb.VectorDBTypeChromaCloud)
}

func TestFactory_ValidateConfig(t *testing.T) {
	factory := NewFactory()

	tests := []struct {
		name    string
		config  *vectordb.Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Nil config",
			config:  nil,
			wantErr: true,
			errMsg:  "config cannot be nil",
		},
		{
			name: "Wrong database type",
			config: &vectordb.Config{
				Type: vectordb.VectorDBTypeMilvusLocal,
			},
			wantErr: true,
			errMsg:  "unsupported Chroma type",
		},
		{
			name: "Missing URL",
			config: &vectordb.Config{
				Type: vectordb.VectorDBTypeChromaLocal,
			},
			wantErr: true,
			errMsg:  "Chroma URL is required",
		},
		{
			name: "Cloud without API key",
			config: &vectordb.Config{
				Type: vectordb.VectorDBTypeChromaCloud,
				URL:  "https://api.trychroma.com",
			},
			wantErr: true,
			errMsg:  "API key is required for Chroma Cloud",
		},
		{
			name: "Negative timeout",
			config: &vectordb.Config{
				Type:    vectordb.VectorDBTypeChromaLocal,
				URL:     "http://localhost:8000",
				Timeout: -1,
			},
			wantErr: true,
			errMsg:  "timeout cannot be negative",
		},
		{
			name: "Negative vector dimensions",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeChromaLocal,
				URL:              "http://localhost:8000",
				VectorDimensions: -100,
			},
			wantErr: true,
			errMsg:  "vector dimensions cannot be negative",
		},
		{
			name: "Invalid similarity metric",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeChromaLocal,
				URL:              "http://localhost:8000",
				SimilarityMetric: "invalid",
			},
			wantErr: true,
			errMsg:  "invalid similarity metric",
		},
		{
			name: "Valid cosine metric",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeChromaLocal,
				URL:              "http://localhost:8000",
				SimilarityMetric: "cosine",
			},
			wantErr: false,
		},
		{
			name: "Valid l2 metric",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeChromaLocal,
				URL:              "http://localhost:8000",
				SimilarityMetric: "l2",
			},
			wantErr: false,
		},
		{
			name: "Valid ip metric",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeChromaLocal,
				URL:              "http://localhost:8000",
				SimilarityMetric: "ip",
			},
			wantErr: false,
		},
		{
			name: "Valid cloud config",
			config: &vectordb.Config{
				Type:   vectordb.VectorDBTypeChromaCloud,
				URL:    "https://api.trychroma.com",
				APIKey: "test-api-key",
			},
			wantErr: false,
		},
		{
			name: "Valid local config minimal",
			config: &vectordb.Config{
				Type: vectordb.VectorDBTypeChromaLocal,
				URL:  "http://localhost:8000",
			},
			wantErr: false,
		},
		{
			name: "Valid local config full",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeChromaLocal,
				URL:              "http://localhost:8000",
				Tenant:           "test-tenant",
				Database:         "test-db",
				VectorDimensions: 1536,
				SimilarityMetric: "cosine",
				Timeout:          30,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := factory.ValidateConfig(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewFactory(t *testing.T) {
	factory := NewFactory()
	assert.NotNil(t, factory)
}

func TestFactory_CreateClient_Defaults(t *testing.T) {
	t.Skip("Requires Chroma connection - needs mock-based approach")

	factory := NewFactory()
	config := &vectordb.Config{
		Type: vectordb.VectorDBTypeChromaLocal,
		URL:  "http://localhost:8000",
	}

	_, err := factory.CreateClient(config)
	if err != nil {
		t.Skip("Skipping test that requires Chroma connection")
	}
}
