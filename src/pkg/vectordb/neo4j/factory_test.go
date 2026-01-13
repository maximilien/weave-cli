// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package neo4j

import (
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/assert"
)

func TestFactory_GetSupportedTypes(t *testing.T) {
	factory := NewFactory()
	types := factory.GetSupportedTypes()

	assert.Len(t, types, 2)
	assert.Contains(t, types, vectordb.VectorDBTypeNeo4jLocal)
	assert.Contains(t, types, vectordb.VectorDBTypeNeo4jCloud)
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
			name: "Valid local config",
			config: &vectordb.Config{
				Type:     vectordb.VectorDBTypeNeo4jLocal,
				Username: "neo4j",
				Password: "password",
			},
			wantErr: false,
		},
		{
			name: "Valid cloud config",
			config: &vectordb.Config{
				Type:     vectordb.VectorDBTypeNeo4jCloud,
				Username: "neo4j",
				Password: "password",
			},
			wantErr: false,
		},
		{
			name: "Invalid type",
			config: &vectordb.Config{
				Type: vectordb.VectorDBTypeMilvusLocal,
			},
			wantErr: true,
			errMsg:  "invalid type",
		},
		{
			name: "Invalid similarity metric",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeNeo4jLocal,
				SimilarityMetric: "invalid",
			},
			wantErr: true,
			errMsg:  "invalid similarity metric",
		},
		{
			name: "Valid cosine metric",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeNeo4jLocal,
				SimilarityMetric: "cosine",
			},
			wantErr: false,
		},
		{
			name: "Valid euclidean metric",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeNeo4jLocal,
				SimilarityMetric: "euclidean",
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
