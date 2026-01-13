// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package milvus

import (
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/assert"
)

func TestFactory_GetSupportedTypes(t *testing.T) {
	factory := NewFactory()
	types := factory.GetSupportedTypes()

	assert.Len(t, types, 2, "Should support 2 types")
	assert.Contains(t, types, vectordb.VectorDBTypeMilvusLocal)
	assert.Contains(t, types, vectordb.VectorDBTypeMilvusCloud)
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
				Type: vectordb.VectorDBTypeWeaviateCloud,
			},
			wantErr: true,
			errMsg:  "unsupported Milvus type",
		},
		{
			name: "Missing address",
			config: &vectordb.Config{
				Type: vectordb.VectorDBTypeMilvusLocal,
			},
			wantErr: true,
			errMsg:  "Milvus address is required",
		},
		{
			name: "Negative timeout",
			config: &vectordb.Config{
				Type:    vectordb.VectorDBTypeMilvusLocal,
				Address: "localhost:19530",
				Timeout: -1,
			},
			wantErr: true,
			errMsg:  "timeout cannot be negative",
		},
		{
			name: "Negative vector dimensions",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeMilvusLocal,
				Address:          "localhost:19530",
				VectorDimensions: -100,
			},
			wantErr: true,
			errMsg:  "vector dimensions cannot be negative",
		},
		{
			name: "Invalid similarity metric",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeMilvusLocal,
				Address:          "localhost:19530",
				SimilarityMetric: "invalid",
			},
			wantErr: true,
			errMsg:  "invalid similarity metric",
		},
		{
			name: "Valid L2 metric",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeMilvusLocal,
				Address:          "localhost:19530",
				SimilarityMetric: "L2",
			},
			wantErr: false,
		},
		{
			name: "Valid IP metric",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeMilvusLocal,
				Address:          "localhost:19530",
				SimilarityMetric: "IP",
			},
			wantErr: false,
		},
		{
			name: "Valid COSINE metric",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeMilvusLocal,
				Address:          "localhost:19530",
				SimilarityMetric: "COSINE",
			},
			wantErr: false,
		},
		{
			name: "Cloud without authentication",
			config: &vectordb.Config{
				Type:    vectordb.VectorDBTypeMilvusCloud,
				Address: "https://cloud.zilliz.com:19530",
			},
			wantErr: true,
			errMsg:  "Milvus cloud requires either APIKey or username and password",
		},
		{
			name: "Cloud with API key",
			config: &vectordb.Config{
				Type:    vectordb.VectorDBTypeMilvusCloud,
				Address: "https://cloud.zilliz.com:19530",
				APIKey:  "test-api-key",
			},
			wantErr: false,
		},
		{
			name: "Cloud with username and password",
			config: &vectordb.Config{
				Type:     vectordb.VectorDBTypeMilvusCloud,
				Address:  "https://cloud.zilliz.com:19530",
				Username: "testuser",
				Password: "testpass",
			},
			wantErr: false,
		},
		{
			name: "Cloud with only username (missing password)",
			config: &vectordb.Config{
				Type:     vectordb.VectorDBTypeMilvusCloud,
				Address:  "https://cloud.zilliz.com:19530",
				Username: "testuser",
			},
			wantErr: true,
			errMsg:  "Milvus cloud requires either APIKey or username and password",
		},
		{
			name: "Valid local config minimal",
			config: &vectordb.Config{
				Type:    vectordb.VectorDBTypeMilvusLocal,
				Address: "localhost:19530",
			},
			wantErr: false,
		},
		{
			name: "Valid local config full",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeMilvusLocal,
				Address:          "localhost:19530",
				Database:         "testdb",
				VectorDimensions: 1536,
				SimilarityMetric: "L2",
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
