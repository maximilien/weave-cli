// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package elasticsearch

import (
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/assert"
)

func TestFactory_GetSupportedTypes(t *testing.T) {
	factory := NewFactory()
	types := factory.GetSupportedTypes()

	assert.Len(t, types, 2)
	assert.Contains(t, types, vectordb.VectorDBTypeElasticsearchLocal)
	assert.Contains(t, types, vectordb.VectorDBTypeElasticsearchCloud)
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
			errMsg:  "unsupported Elasticsearch type",
		},
		{
			name: "Missing URL, Address, and APIKey",
			config: &vectordb.Config{
				Type: vectordb.VectorDBTypeElasticsearchLocal,
			},
			wantErr: true,
			errMsg:  "URL, Address, or APIKey is required",
		},
		{
			name: "Negative timeout",
			config: &vectordb.Config{
				Type:    vectordb.VectorDBTypeElasticsearchLocal,
				URL:     "http://localhost:9200",
				Timeout: -1,
			},
			wantErr: true,
			errMsg:  "timeout cannot be negative",
		},
		{
			name: "Negative vector dimensions",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeElasticsearchLocal,
				URL:              "http://localhost:9200",
				VectorDimensions: -100,
			},
			wantErr: true,
			errMsg:  "vector dimensions cannot be negative",
		},
		{
			name: "Invalid similarity metric",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeElasticsearchLocal,
				URL:              "http://localhost:9200",
				SimilarityMetric: "invalid",
			},
			wantErr: true,
			errMsg:  "invalid similarity metric",
		},
		{
			name: "Valid cosine metric",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeElasticsearchLocal,
				URL:              "http://localhost:9200",
				SimilarityMetric: "cosine",
			},
			wantErr: false,
		},
		{
			name: "Valid dot_product metric",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeElasticsearchLocal,
				URL:              "http://localhost:9200",
				SimilarityMetric: "dot_product",
			},
			wantErr: false,
		},
		{
			name: "Valid l2_norm metric",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeElasticsearchLocal,
				URL:              "http://localhost:9200",
				SimilarityMetric: "l2_norm",
			},
			wantErr: false,
		},
		{
			name: "Valid config with URL",
			config: &vectordb.Config{
				Type: vectordb.VectorDBTypeElasticsearchLocal,
				URL:  "http://localhost:9200",
			},
			wantErr: false,
		},
		{
			name: "Valid config with Address",
			config: &vectordb.Config{
				Type:    vectordb.VectorDBTypeElasticsearchLocal,
				Address: "localhost:9200",
			},
			wantErr: false,
		},
		{
			name: "Valid config with APIKey",
			config: &vectordb.Config{
				Type:   vectordb.VectorDBTypeElasticsearchCloud,
				APIKey: "test-api-key",
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
