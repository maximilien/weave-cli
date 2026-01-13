// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package opensearch

import (
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/assert"
)

func TestFactory_GetSupportedTypes(t *testing.T) {
	factory := NewFactory()
	types := factory.GetSupportedTypes()

	assert.Len(t, types, 2)
	assert.Contains(t, types, vectordb.VectorDBTypeOpenSearchLocal)
	assert.Contains(t, types, vectordb.VectorDBTypeOpenSearchCloud)
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
			errMsg:  "unsupported OpenSearch type",
		},
		{
			name: "Missing URL and Address",
			config: &vectordb.Config{
				Type: vectordb.VectorDBTypeOpenSearchLocal,
			},
			wantErr: true,
			errMsg:  "OpenSearch URL or address is required",
		},
		{
			name: "Negative timeout",
			config: &vectordb.Config{
				Type:    vectordb.VectorDBTypeOpenSearchLocal,
				URL:     "http://localhost:9200",
				Timeout: -1,
			},
			wantErr: true,
			errMsg:  "timeout cannot be negative",
		},
		{
			name: "Negative vector dimensions",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeOpenSearchLocal,
				URL:              "http://localhost:9200",
				VectorDimensions: -100,
			},
			wantErr: true,
			errMsg:  "vector dimensions cannot be negative",
		},
		{
			name: "Invalid similarity metric",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeOpenSearchLocal,
				URL:              "http://localhost:9200",
				SimilarityMetric: "invalid",
			},
			wantErr: true,
			errMsg:  "invalid similarity metric",
		},
		{
			name: "Valid l2 metric",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeOpenSearchLocal,
				URL:              "http://localhost:9200",
				SimilarityMetric: "l2",
			},
			wantErr: false,
		},
		{
			name: "Valid cosinesimil metric",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeOpenSearchLocal,
				URL:              "http://localhost:9200",
				SimilarityMetric: "cosinesimil",
			},
			wantErr: false,
		},
		{
			name: "Valid innerproduct metric",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeOpenSearchLocal,
				URL:              "http://localhost:9200",
				SimilarityMetric: "innerproduct",
			},
			wantErr: false,
		},
		{
			name: "Valid l1 metric",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeOpenSearchLocal,
				URL:              "http://localhost:9200",
				SimilarityMetric: "l1",
			},
			wantErr: false,
		},
		{
			name: "Valid linf metric",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeOpenSearchLocal,
				URL:              "http://localhost:9200",
				SimilarityMetric: "linf",
			},
			wantErr: false,
		},
		{
			name: "Valid config with URL",
			config: &vectordb.Config{
				Type: vectordb.VectorDBTypeOpenSearchLocal,
				URL:  "http://localhost:9200",
			},
			wantErr: false,
		},
		{
			name: "Valid config with Address",
			config: &vectordb.Config{
				Type:    vectordb.VectorDBTypeOpenSearchLocal,
				Address: "localhost:9200",
			},
			wantErr: false,
		},
		{
			name: "Valid cloud config",
			config: &vectordb.Config{
				Type:     vectordb.VectorDBTypeOpenSearchCloud,
				URL:      "https://my-cluster.us-east-1.es.amazonaws.com",
				Username: "admin",
				Password: "password",
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
