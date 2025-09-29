// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max


package tests

import (
	"github.com/maximilien/weave-cli/src/pkg/config"
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Test loading config with default files
	cfg, err := config.LoadConfig("", "")
	if err != nil {
		t.Logf("Config loading failed (expected if no config files): %v", err)
		return
	}

	if cfg == nil {
		t.Error("Config should not be nil")
	}

	// Test that we can get the default database
	if cfg != nil {
		defaultDB, err := cfg.GetDefaultDatabase()
		if err != nil {
			t.Logf("Failed to get default database (expected if no databases configured): %v", err)
		} else if defaultDB == nil {
			t.Error("Default database should not be nil")
		}
	}
}

func TestInterpolateEnvVars(t *testing.T) {
	// Set test environment variable
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	testCases := []struct {
		input    string
		expected string
	}{
		{"${TEST_VAR:-default}", "test_value"},
		{"${NONEXISTENT_VAR:-default}", "default"},
		{"simple string", "simple string"},
		{"${TEST_VAR}", "test_value"},
		{"${NONEXISTENT_VAR}", ""},
	}

	for _, tc := range testCases {
		result := config.InterpolateString(tc.input)
		if result != tc.expected {
			t.Errorf("Expected %s, got %s", tc.expected, result)
		}
	}
}

func TestVectorDBTypes(t *testing.T) {
	// Test vector database type constants
	if config.VectorDBTypeCloud != "weaviate-cloud" {
		t.Errorf("Expected weaviate-cloud, got %s", config.VectorDBTypeCloud)
	}

	if config.VectorDBTypeLocal != "weaviate-local" {
		t.Errorf("Expected weaviate-local, got %s", config.VectorDBTypeLocal)
	}

	if config.VectorDBTypeMock != "mock" {
		t.Errorf("Expected mock, got %s", config.VectorDBTypeMock)
	}
}

func TestMultipleDatabases(t *testing.T) {
	// Create a test config with multiple databases
	cfg := &config.Config{
		Databases: config.DatabasesConfig{
			Default: "default",
			VectorDatabases: []config.VectorDBConfig{
				{
					Name:    "default",
					Type:    config.VectorDBTypeMock,
					Enabled: true,
					Collections: []config.Collection{
						{Name: "test", Type: "text"},
					},
				},
				{
					Name: "production",
					Type: config.VectorDBTypeCloud,
					URL:  "https://prod.example.com",
					Collections: []config.Collection{
						{Name: "prod", Type: "text"},
					},
				},
			},
		},
	}

	// Test GetDefaultDatabase
	defaultDB, err := cfg.GetDefaultDatabase()
	if err != nil {
		t.Errorf("Failed to get default database: %v", err)
	}
	if defaultDB.Type != config.VectorDBTypeMock {
		t.Errorf("Expected mock type for default database, got %s", defaultDB.Type)
	}

	// Test GetDatabase
	prodDB, err := cfg.GetDatabase("production")
	if err != nil {
		t.Errorf("Failed to get production database: %v", err)
	}
	if prodDB.Type != config.VectorDBTypeCloud {
		t.Errorf("Expected cloud type for production database, got %s", prodDB.Type)
	}

	// Test ListDatabases
	databaseNames := cfg.ListDatabases()
	if len(databaseNames) != 2 {
		t.Errorf("Expected 2 databases, got %d", len(databaseNames))
	}

	// Test GetDatabaseNames
	dbNames := cfg.GetDatabaseNames()
	if len(dbNames) != 2 {
		t.Errorf("Expected 2 database names, got %d", len(dbNames))
	}
	if dbNames["default"] != config.VectorDBTypeMock {
		t.Errorf("Expected mock type for default, got %s", dbNames["default"])
	}
	if dbNames["production"] != config.VectorDBTypeCloud {
		t.Errorf("Expected cloud type for production, got %s", dbNames["production"])
	}
}
