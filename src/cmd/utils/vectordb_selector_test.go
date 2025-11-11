// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package utils

import (
	"context"
	"os"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/config"
)

// TestSelectDefaultDatabase tests the SelectDefaultDatabase function
func TestSelectDefaultDatabase(t *testing.T) {
	tests := []struct {
		name          string
		configs       []config.VectorDBConfig
		envValue      string
		configDefault string
		expectedType  config.VectorDBType
		expectNil     bool
	}{
		{
			name: "selects from env variable",
			configs: []config.VectorDBConfig{
				{Name: "cloud", Type: config.VectorDBTypeCloud},
				{Name: "local", Type: config.VectorDBTypeLocal},
			},
			envValue:     "weaviate-cloud",
			expectedType: config.VectorDBTypeCloud,
		},
		{
			name: "selects from config default when no env",
			configs: []config.VectorDBConfig{
				{Name: "cloud", Type: config.VectorDBTypeCloud},
				{Name: "local", Type: config.VectorDBTypeLocal},
			},
			configDefault: "weaviate-local",
			expectedType:  config.VectorDBTypeLocal,
		},
		{
			name: "env takes precedence over config",
			configs: []config.VectorDBConfig{
				{Name: "cloud", Type: config.VectorDBTypeCloud},
				{Name: "local", Type: config.VectorDBTypeLocal},
			},
			envValue:      "weaviate-cloud",
			configDefault: "weaviate-local",
			expectedType:  config.VectorDBTypeCloud,
		},
		{
			name: "returns nil when no default set",
			configs: []config.VectorDBConfig{
				{Name: "cloud", Type: config.VectorDBTypeCloud},
			},
			expectNil: true,
		},
		{
			name: "returns nil when default doesn't match any config",
			configs: []config.VectorDBConfig{
				{Name: "cloud", Type: config.VectorDBTypeCloud},
			},
			envValue:  "supabase",
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			if tt.envValue != "" {
				os.Setenv("VECTOR_DB_TYPE", tt.envValue)
				defer os.Unsetenv("VECTOR_DB_TYPE")
			} else {
				os.Unsetenv("VECTOR_DB_TYPE")
			}

			// Setup config
			cfg := &config.Config{}
			if tt.configDefault != "" {
				cfg.Databases.Default = tt.configDefault
			}

			// Test
			result := SelectDefaultDatabase(tt.configs, cfg)

			// Verify
			if tt.expectNil {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
			} else {
				if result == nil {
					t.Errorf("Expected non-nil result")
					return
				}
				if result.Type != tt.expectedType {
					t.Errorf("Expected type %s, got %s", tt.expectedType, result.Type)
				}
			}
		})
	}
}

// TestAreAllWeaviateDatabases tests the AreAllWeaviateDatabases function
func TestAreAllWeaviateDatabases(t *testing.T) {
	tests := []struct {
		name     string
		configs  []config.VectorDBConfig
		expected bool
	}{
		{
			name: "all weaviate cloud",
			configs: []config.VectorDBConfig{
				{Type: config.VectorDBTypeCloud},
				{Type: config.VectorDBTypeCloud},
			},
			expected: true,
		},
		{
			name: "all weaviate local",
			configs: []config.VectorDBConfig{
				{Type: config.VectorDBTypeLocal},
				{Type: config.VectorDBTypeLocal},
			},
			expected: true,
		},
		{
			name: "mixed weaviate types",
			configs: []config.VectorDBConfig{
				{Type: config.VectorDBTypeCloud},
				{Type: config.VectorDBTypeLocal},
			},
			expected: true,
		},
		{
			name: "contains non-weaviate",
			configs: []config.VectorDBConfig{
				{Type: config.VectorDBTypeCloud},
				{Type: config.VectorDBTypeSupabase},
			},
			expected: false,
		},
		{
			name: "only supabase",
			configs: []config.VectorDBConfig{
				{Type: config.VectorDBTypeSupabase},
			},
			expected: false,
		},
		{
			name: "only mock",
			configs: []config.VectorDBConfig{
				{Type: config.VectorDBTypeMock},
			},
			expected: false,
		},
		{
			name:     "empty configs",
			configs:  []config.VectorDBConfig{},
			expected: true, // vacuously true
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AreAllWeaviateDatabases(tt.configs)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestSelectSingleDatabase tests the SelectSingleDatabase function
func TestSelectSingleDatabase(t *testing.T) {
	tests := []struct {
		name           string
		configs        []config.VectorDBConfig
		envValue       string
		collectionName string
		expectedType   config.VectorDBType
		expectNil      bool
	}{
		{
			name: "single config returns that config",
			configs: []config.VectorDBConfig{
				{Name: "cloud", Type: config.VectorDBTypeCloud},
			},
			expectedType: config.VectorDBTypeCloud,
		},
		{
			name: "uses default from env when multiple configs",
			configs: []config.VectorDBConfig{
				{Name: "cloud", Type: config.VectorDBTypeCloud},
				{Name: "local", Type: config.VectorDBTypeLocal},
			},
			envValue:     "weaviate-local",
			expectedType: config.VectorDBTypeLocal,
		},
		{
			name: "returns nil when multiple non-weaviate and no default",
			configs: []config.VectorDBConfig{
				{Name: "cloud", Type: config.VectorDBTypeCloud},
				{Name: "supabase", Type: config.VectorDBTypeSupabase},
			},
			expectNil: true,
		},
		{
			name: "returns nil when multiple configs no default and no collection",
			configs: []config.VectorDBConfig{
				{Name: "cloud", Type: config.VectorDBTypeCloud},
				{Name: "local", Type: config.VectorDBTypeLocal},
			},
			collectionName: "", // no collection name to search for
			expectNil:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			if tt.envValue != "" {
				os.Setenv("VECTOR_DB_TYPE", tt.envValue)
				defer os.Unsetenv("VECTOR_DB_TYPE")
			} else {
				os.Unsetenv("VECTOR_DB_TYPE")
			}

			cfg := &config.Config{}
			ctx := context.Background()

			// Test
			result := SelectSingleDatabase(ctx, tt.configs, cfg, tt.collectionName)

			// Verify
			if tt.expectNil {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
			} else {
				if result == nil {
					t.Errorf("Expected non-nil result")
					return
				}
				if result.Type != tt.expectedType {
					t.Errorf("Expected type %s, got %s", tt.expectedType, result.Type)
				}
			}
		})
	}
}

// TestGetVectorDBTypeFromConfig tests the GetVectorDBTypeFromConfig function
func TestGetVectorDBTypeFromConfig(t *testing.T) {
	tests := []struct {
		name     string
		dbConfig config.VectorDBConfig
		expected string
	}{
		{
			name:     "weaviate cloud",
			dbConfig: config.VectorDBConfig{Type: config.VectorDBTypeCloud},
			expected: "weaviate-cloud",
		},
		{
			name:     "weaviate local",
			dbConfig: config.VectorDBConfig{Type: config.VectorDBTypeLocal},
			expected: "weaviate-local",
		},
		{
			name:     "supabase",
			dbConfig: config.VectorDBConfig{Type: config.VectorDBTypeSupabase},
			expected: "supabase",
		},
		{
			name:     "mock",
			dbConfig: config.VectorDBConfig{Type: config.VectorDBTypeMock},
			expected: "mock",
		},
		{
			name:     "unknown type",
			dbConfig: config.VectorDBConfig{Type: config.VectorDBType("unknown")},
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetVectorDBTypeFromConfig(&tt.dbConfig)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestValidateDatabaseSelection tests the ValidateDatabaseSelection function
func TestValidateDatabaseSelection(t *testing.T) {
	tests := []struct {
		name        string
		selection   *VectorDBSelection
		opType      OperationType
		expectError bool
	}{
		{
			name:        "nil selection returns error",
			selection:   nil,
			opType:      OperationTypeWrite,
			expectError: true,
		},
		{
			name: "empty configs returns error",
			selection: &VectorDBSelection{
				Configs: []config.VectorDBConfig{},
			},
			opType:      OperationTypeWrite,
			expectError: true,
		},
		{
			name: "single config for write operation succeeds",
			selection: &VectorDBSelection{
				Configs: []config.VectorDBConfig{
					{Type: config.VectorDBTypeCloud},
				},
			},
			opType:      OperationTypeWrite,
			expectError: false,
		},
		{
			name: "multiple configs for write operation fails",
			selection: &VectorDBSelection{
				Configs: []config.VectorDBConfig{
					{Type: config.VectorDBTypeCloud},
					{Type: config.VectorDBTypeLocal},
				},
			},
			opType:      OperationTypeWrite,
			expectError: true,
		},
		{
			name: "multiple configs for delete operation fails",
			selection: &VectorDBSelection{
				Configs: []config.VectorDBConfig{
					{Type: config.VectorDBTypeCloud},
					{Type: config.VectorDBTypeLocal},
				},
			},
			opType:      OperationTypeDelete,
			expectError: true,
		},
		{
			name: "multiple configs for read operation succeeds",
			selection: &VectorDBSelection{
				Configs: []config.VectorDBConfig{
					{Type: config.VectorDBTypeCloud},
					{Type: config.VectorDBTypeLocal},
				},
			},
			opType:      OperationTypeRead,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDatabaseSelection(tt.selection, tt.opType, "test operation")
			if tt.expectError && err == nil {
				t.Errorf("Expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

// TestGetAllConfiguredDatabases tests the getAllConfiguredDatabases function
func TestGetAllConfiguredDatabases(t *testing.T) {
	tests := []struct {
		name          string
		config        *config.Config
		expectedCount int
		expectError   bool
	}{
		{
			name: "returns configured databases",
			config: &config.Config{
				Databases: config.DatabasesConfig{
					VectorDatabases: []config.VectorDBConfig{
						{Name: "cloud", Type: config.VectorDBTypeCloud},
						{Name: "local", Type: config.VectorDBTypeLocal},
					},
				},
			},
			expectedCount: 2,
		},
		{
			name: "returns error when no databases configured",
			config: &config.Config{
				Databases: config.DatabasesConfig{
					VectorDatabases: []config.VectorDBConfig{},
				},
			},
			expectError: true,
		},
		{
			name: "single database returns one",
			config: &config.Config{
				Databases: config.DatabasesConfig{
					VectorDatabases: []config.VectorDBConfig{
						{Name: "mock", Type: config.VectorDBTypeMock},
					},
				},
			},
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getAllConfiguredDatabases(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
				return
			}

			if len(result.Configs) != tt.expectedCount {
				t.Errorf("Expected %d configs, got %d", tt.expectedCount, len(result.Configs))
			}
		})
	}
}

// TestGetWeaviateConfigs tests the getWeaviateConfigs function
func TestGetWeaviateConfigs(t *testing.T) {
	tests := []struct {
		name          string
		config        *config.Config
		expectedCount int
		expectError   bool
	}{
		{
			name: "returns only weaviate configs",
			config: &config.Config{
				Databases: config.DatabasesConfig{
					VectorDatabases: []config.VectorDBConfig{
						{Name: "cloud", Type: config.VectorDBTypeCloud},
						{Name: "local", Type: config.VectorDBTypeLocal},
						{Name: "supabase", Type: config.VectorDBTypeSupabase},
					},
				},
			},
			expectedCount: 2,
		},
		{
			name: "returns error when no weaviate configs",
			config: &config.Config{
				Databases: config.DatabasesConfig{
					VectorDatabases: []config.VectorDBConfig{
						{Name: "supabase", Type: config.VectorDBTypeSupabase},
						{Name: "mock", Type: config.VectorDBTypeMock},
					},
				},
			},
			expectError: true,
		},
		{
			name: "returns both cloud and local",
			config: &config.Config{
				Databases: config.DatabasesConfig{
					VectorDatabases: []config.VectorDBConfig{
						{Name: "cloud", Type: config.VectorDBTypeCloud},
						{Name: "local", Type: config.VectorDBTypeLocal},
					},
				},
			},
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getWeaviateConfigs(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
				return
			}

			if len(result) != tt.expectedCount {
				t.Errorf("Expected %d configs, got %d", tt.expectedCount, len(result))
			}

			// Verify all returned configs are Weaviate
			for _, cfg := range result {
				if cfg.Type != config.VectorDBTypeCloud && cfg.Type != config.VectorDBTypeLocal {
					t.Errorf("Expected Weaviate config, got %s", cfg.Type)
				}
			}
		})
	}
}
