// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package supabase

import (
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/assert"
)

func TestFactory_GetSupportedTypes(t *testing.T) {
	factory := NewFactory()
	types := factory.GetSupportedTypes()

	assert.Len(t, types, 3)
	assert.Contains(t, types, vectordb.VectorDBTypeSupabase)
	assert.Contains(t, types, vectordb.VectorDBTypeSupabaseCloud)
	assert.Contains(t, types, vectordb.VectorDBTypeSupabaseLocal)
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
			name: "Valid cloud config",
			config: &vectordb.Config{
				Type:        vectordb.VectorDBTypeSupabase,
				DatabaseURL: "postgresql://user:pass@db.test.supabase.co:5432/postgres",
				DatabaseKey: "test-anon-key",
			},
			wantErr: false,
		},
		{
			name: "Valid local config",
			config: &vectordb.Config{
				Type:        vectordb.VectorDBTypeSupabaseLocal,
				DatabaseURL: "postgresql://user:pass@localhost:5432/postgres",
			},
			wantErr: false,
		},
		{
			name: "Missing database key for cloud",
			config: &vectordb.Config{
				Type:        vectordb.VectorDBTypeSupabase,
				DatabaseURL: "postgresql://user:pass@db.test.supabase.co:5432/postgres",
			},
			wantErr: true,
			errMsg:  "database key",
		},
		{
			name: "Missing database URL",
			config: &vectordb.Config{
				Type:        vectordb.VectorDBTypeSupabase,
				DatabaseKey: "test-anon-key",
			},
			wantErr: true,
			errMsg:  "database URL is required",
		},
		{
			name: "Invalid type",
			config: &vectordb.Config{
				Type: vectordb.VectorDBTypeMilvusLocal,
			},
			wantErr: true,
			errMsg:  "unsupported Supabase type",
		},
		{
			name: "Invalid database URL format",
			config: &vectordb.Config{
				Type:        vectordb.VectorDBTypeSupabase,
				DatabaseURL: "http://test.supabase.co",
				DatabaseKey: "test-anon-key",
			},
			wantErr: true,
			errMsg:  "must be a valid PostgreSQL connection string",
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
