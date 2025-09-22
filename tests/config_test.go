package tests

import (
	"os"
	"testing"
	"github.com/maximilien/weave-cli/src/pkg/config"
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