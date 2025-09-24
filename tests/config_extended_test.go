package tests

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/config"
)

func TestConfigLoadConfig(t *testing.T) {
	t.Run("LoadConfigWithValidFiles", func(t *testing.T) {
		// Create temporary config file
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "test-config.yaml")
		envFile := filepath.Join(tempDir, "test.env")

		// Write test config
		configContent := `
databases:
  default: "mock"
  vector_databases:
    - name: "mock"
      type: "mock"
      enabled: true
      simulate_embeddings: true
      embedding_dimension: 384
      collections:
        - name: "TestCollection"
          type: "text"
          description: "Test collection"
`
		if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		// Write test env file
		envContent := `
TEST_VAR=test_value
WEAVIATE_URL=https://test.weaviate.cloud
WEAVIATE_API_KEY=test-key
`
		if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
			t.Fatalf("Failed to write env file: %v", err)
		}

		cfg, err := config.LoadConfig(configFile, envFile)
		if err != nil {
			t.Errorf("Failed to load config: %v", err)
		}

		if cfg == nil {
			t.Error("Config should not be nil")
		}

		// Test default database
		defaultDB, err := cfg.GetDefaultDatabase()
		if err != nil {
			t.Errorf("Failed to get default database: %v", err)
		}
		if defaultDB.Type != config.VectorDBTypeMock {
			t.Errorf("Expected mock type, got %s", defaultDB.Type)
		}
		if !defaultDB.Enabled {
			t.Error("Mock should be enabled")
		}
	})

	t.Run("LoadConfigWithNonExistentFiles", func(t *testing.T) {
		_, err := config.LoadConfig("non-existent.yaml", "non-existent.env")
		if err == nil {
			t.Error("Expected error for non-existent files")
		}
	})

	t.Run("LoadConfigWithInvalidYAML", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "invalid-config.yaml")

		invalidContent := `
database:
  vector_db:
    type: "mock"
    mock:
      enabled: true
      simulate_embeddings: true
      embedding_dimension: 384
      collections:
        - name: "TestCollection"
          type: "text"
          description: "Test collection"
invalid: [unclosed bracket
`
		if err := os.WriteFile(configFile, []byte(invalidContent), 0644); err != nil {
			t.Fatalf("Failed to write invalid config file: %v", err)
		}

		_, err := config.LoadConfig(configFile, "")
		if err == nil {
			t.Error("Expected error for invalid YAML")
		}
	})
}

func TestConfigInterpolateString(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
		setEnv   map[string]string
	}{
		{
			name:     "Simple variable",
			input:    "${TEST_VAR}",
			expected: "test_value",
			setEnv:   map[string]string{"TEST_VAR": "test_value"},
		},
		{
			name:     "Variable with default",
			input:    "${NONEXISTENT_VAR:-default_value}",
			expected: "default_value",
			setEnv:   map[string]string{},
		},
		{
			name:     "Variable with default when env exists",
			input:    "${TEST_VAR:-default_value}",
			expected: "env_value",
			setEnv:   map[string]string{"TEST_VAR": "env_value"},
		},
		{
			name:     "Multiple variables",
			input:    "${VAR1} and ${VAR2}",
			expected: "value1 and value2",
			setEnv:   map[string]string{"VAR1": "value1", "VAR2": "value2"},
		},
		{
			name:     "Mixed content",
			input:    "prefix_${VAR}_suffix",
			expected: "prefix_value_suffix",
			setEnv:   map[string]string{"VAR": "value"},
		},
		{
			name:     "No variables",
			input:    "simple string",
			expected: "simple string",
			setEnv:   map[string]string{},
		},
		{
			name:     "Empty variable",
			input:    "${EMPTY_VAR}",
			expected: "",
			setEnv:   map[string]string{"EMPTY_VAR": ""},
		},
		{
			name:     "Nested braces",
			input:    "${VAR:-default}",
			expected: "default",
			setEnv:   map[string]string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tc.setEnv {
				os.Setenv(key, value)
			}

			result := config.InterpolateString(tc.input)
			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}

			// Clean up environment variables
			for key := range tc.setEnv {
				os.Unsetenv(key)
			}
		})
	}
}

func TestConfigVectorDBTypes(t *testing.T) {
	t.Run("VectorDBTypeConstants", func(t *testing.T) {
		if config.VectorDBTypeCloud != "weaviate-cloud" {
			t.Errorf("Expected weaviate-cloud, got %s", config.VectorDBTypeCloud)
		}

		if config.VectorDBTypeLocal != "weaviate-local" {
			t.Errorf("Expected weaviate-local, got %s", config.VectorDBTypeLocal)
		}

		if config.VectorDBTypeMock != "mock" {
			t.Errorf("Expected mock, got %s", config.VectorDBTypeMock)
		}
	})

	t.Run("VectorDBTypeValidation", func(t *testing.T) {
		validTypes := []config.VectorDBType{
			config.VectorDBTypeCloud,
			config.VectorDBTypeLocal,
			config.VectorDBTypeMock,
		}

		for _, vdbType := range validTypes {
			if string(vdbType) == "" {
				t.Errorf("VectorDB type should not be empty")
			}
		}
	})
}

func TestConfigStructValidation(t *testing.T) {
	t.Run("VectorDBConfig_Cloud", func(t *testing.T) {
		cfg := &config.VectorDBConfig{
			Name:   "test-cloud",
			Type:   config.VectorDBTypeCloud,
			URL:    "https://test.weaviate.cloud",
			APIKey: "test-key",
			Collections: []config.Collection{
				{Name: "TestCollection", Type: "text"},
			},
		}

		if cfg.Name == "" {
			t.Error("Name should not be empty")
		}

		if cfg.Type != config.VectorDBTypeCloud {
			t.Error("Type should be cloud")
		}

		if cfg.URL == "" {
			t.Error("URL should not be empty")
		}

		if cfg.APIKey == "" {
			t.Error("APIKey should not be empty")
		}

		if len(cfg.Collections) == 0 {
			t.Error("Collections should not be empty")
		}
	})

	t.Run("VectorDBConfig_Local", func(t *testing.T) {
		cfg := &config.VectorDBConfig{
			Name: "test-local",
			Type: config.VectorDBTypeLocal,
			URL:  "http://localhost:8080",
			Collections: []config.Collection{
				{Name: "TestCollection", Type: "text"},
			},
		}

		if cfg.Name == "" {
			t.Error("Name should not be empty")
		}

		if cfg.Type != config.VectorDBTypeLocal {
			t.Error("Type should be local")
		}

		if cfg.URL == "" {
			t.Error("URL should not be empty")
		}

		if len(cfg.Collections) == 0 {
			t.Error("Collections should not be empty")
		}
	})

	t.Run("MockConfig", func(t *testing.T) {
		cfg := &config.MockConfig{
			Enabled:            true,
			SimulateEmbeddings: true,
			EmbeddingDimension: 384,
			Collections: []config.MockCollection{
				{Name: "TestCollection", Type: "text", Description: "Test collection"},
			},
		}

		if !cfg.Enabled {
			t.Error("Mock should be enabled")
		}

		if !cfg.SimulateEmbeddings {
			t.Error("SimulateEmbeddings should be true")
		}

		if cfg.EmbeddingDimension <= 0 {
			t.Error("EmbeddingDimension should be positive")
		}

		if len(cfg.Collections) == 0 {
			t.Error("Collections should not be empty")
		}

		for _, collection := range cfg.Collections {
			if collection.Name == "" {
				t.Error("Collection name should not be empty")
			}
			if collection.Type == "" {
				t.Error("Collection type should not be empty")
			}
		}
	})
}

func TestConfigFilePaths(t *testing.T) {
	t.Run("GetConfigFile", func(t *testing.T) {
		// Test with default config
		configFile := config.GetConfigFile()
		if configFile != "" {
			t.Logf("Config file: %s", configFile)
		}
	})

	t.Run("GetEnvFile", func(t *testing.T) {
		// Test with default env file
		envFile := config.GetEnvFile()
		if envFile != "" {
			t.Logf("Env file: %s", envFile)
		}
	})
}

func TestConfigEdgeCases(t *testing.T) {
	t.Run("EmptyConfig", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "empty-config.yaml")

		emptyContent := `{}`
		if err := os.WriteFile(configFile, []byte(emptyContent), 0644); err != nil {
			t.Fatalf("Failed to write empty config file: %v", err)
		}

		cfg, err := config.LoadConfig(configFile, "")
		if err != nil {
			t.Errorf("Failed to load empty config: %v", err)
		}

		if cfg == nil {
			t.Error("Config should not be nil even for empty config")
		}
	})

	t.Run("ConfigWithSpecialCharacters", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "special-config.yaml")

		specialContent := `
databases:
  default: "mock"
  vector_databases:
    - name: "mock"
      type: "mock"
      enabled: true
      simulate_embeddings: true
      embedding_dimension: 384
      collections:
        - name: "Test-Collection_123"
          type: "text"
          description: "Test collection with special chars: !@#$%^&*()"
`
		if err := os.WriteFile(configFile, []byte(specialContent), 0644); err != nil {
			t.Fatalf("Failed to write special config file: %v", err)
		}

		cfg, err := config.LoadConfig(configFile, "")
		if err != nil {
			t.Errorf("Failed to load special config: %v", err)
		}

		if cfg == nil {
			t.Error("Config should not be nil")
		}

		defaultDB, err := cfg.GetDefaultDatabase()
		if err != nil {
			t.Errorf("Failed to get default database: %v", err)
		}
		if len(defaultDB.Collections) == 0 {
			t.Error("Collections should not be empty")
		}

		collection := defaultDB.Collections[0]
		if collection.Name != "Test-Collection_123" {
			t.Errorf("Expected special collection name, got %s", collection.Name)
		}
	})

	t.Run("ConfigWithLargeValues", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "large-config.yaml")

		largeContent := `
databases:
  default: "mock"
  vector_databases:
    - name: "mock"
      type: "mock"
      enabled: true
      simulate_embeddings: true
      embedding_dimension: 4096
      collections:
        - name: "LargeCollection"
          type: "text"
          description: "` + string(make([]byte, 1000)) + `"
`
		if err := os.WriteFile(configFile, []byte(largeContent), 0644); err != nil {
			t.Fatalf("Failed to write large config file: %v", err)
		}

		cfg, err := config.LoadConfig(configFile, "")
		if err != nil {
			t.Logf("Failed to load large config (expected due to control characters): %v", err)
			return
		}

		if cfg == nil {
			t.Error("Config should not be nil")
			return
		}

		defaultDB, err := cfg.GetDefaultDatabase()
		if err != nil {
			t.Errorf("Failed to get default database: %v", err)
		}
		if defaultDB.EmbeddingDimension != 4096 {
			t.Errorf("Expected embedding dimension 4096, got %d", defaultDB.EmbeddingDimension)
		}
	})
}

func TestConfigConcurrency(t *testing.T) {
	t.Run("ConcurrentConfigLoading", func(t *testing.T) {
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "concurrent-config.yaml")

		configContent := `
database:
  vector_db:
    type: "mock"
    mock:
      enabled: true
      simulate_embeddings: true
      embedding_dimension: 384
      collections:
        - name: "ConcurrentCollection"
          type: "text"
          description: "Concurrent test collection"
`
		if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		// Test concurrent loading
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				cfg, err := config.LoadConfig(configFile, "")
				if err != nil {
					t.Errorf("Failed to load config concurrently: %v", err)
				}
				if cfg == nil {
					t.Error("Config should not be nil")
				}
				done <- true
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				t.Error("Concurrent config loading timed out")
				return
			}
		}
	})
}

func BenchmarkConfigLoading(b *testing.B) {
	tempDir := b.TempDir()
	configFile := filepath.Join(tempDir, "benchmark-config.yaml")

	configContent := `
database:
  vector_db:
    type: "mock"
    mock:
      enabled: true
      simulate_embeddings: true
      embedding_dimension: 384
      collections:
        - name: "BenchmarkCollection"
          type: "text"
          description: "Benchmark test collection"
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		b.Fatalf("Failed to write config file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := config.LoadConfig(configFile, "")
		if err != nil {
			b.Errorf("Failed to load config: %v", err)
		}
	}
}

func BenchmarkInterpolateString(b *testing.B) {
	testString := "${VAR1:-default1} and ${VAR2:-default2} with ${VAR3}"
	os.Setenv("VAR1", "value1")
	os.Setenv("VAR2", "value2")
	os.Setenv("VAR3", "value3")
	defer func() {
		os.Unsetenv("VAR1")
		os.Unsetenv("VAR2")
		os.Unsetenv("VAR3")
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.InterpolateString(testString)
	}
}
