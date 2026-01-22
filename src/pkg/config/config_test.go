// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package config

import (
	"os"
	"path/filepath"
	"testing"

	yaml "gopkg.in/yaml.v3"
)

func TestLoadSchemasFromDirectory(t *testing.T) {
	// Create a temporary directory for test schemas
	tmpDir := t.TempDir()

	// Create test schema files
	schema1 := SchemaDefinition{
		Name: "TestSchema1",
		Schema: map[string]interface{}{
			"class":      "TestClass1",
			"vectorizer": "text2vec-weaviate",
			"properties": []interface{}{
				map[string]interface{}{
					"name":        "field1",
					"datatype":    []interface{}{"text"},
					"description": "test field 1",
				},
			},
		},
	}

	schema2 := SchemaDefinition{
		Name: "TestSchema2",
		Schema: map[string]interface{}{
			"class":      "TestClass2",
			"vectorizer": "text2vec-transformers",
		},
	}

	// Write schema files
	schema1Data, _ := yaml.Marshal(schema1)
	schema2Data, _ := yaml.Marshal(schema2)

	if err := os.WriteFile(filepath.Join(tmpDir, "schema1.yaml"), schema1Data, 0644); err != nil {
		t.Fatalf("Failed to write schema1.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "schema2.yml"), schema2Data, 0644); err != nil {
		t.Fatalf("Failed to write schema2.yml: %v", err)
	}

	// Create config with schemas_dir
	config := &Config{
		SchemasDir: tmpDir,
		Databases: DatabasesConfig{
			Schemas: []SchemaDefinition{},
		},
	}

	// Load schemas from directory
	err := config.loadSchemasFromDirectory()
	if err != nil {
		t.Fatalf("Failed to load schemas from directory: %v", err)
	}

	// Verify schemas were loaded
	if len(config.Databases.Schemas) != 2 {
		t.Errorf("Expected 2 schemas, got %d", len(config.Databases.Schemas))
	}

	// Verify schema names
	schemaNames := make(map[string]bool)
	for _, schema := range config.Databases.Schemas {
		schemaNames[schema.Name] = true
	}

	if !schemaNames["TestSchema1"] {
		t.Error("TestSchema1 not found in loaded schemas")
	}
	if !schemaNames["TestSchema2"] {
		t.Error("TestSchema2 not found in loaded schemas")
	}
}

func TestLoadSchemasFromDirectory_Precedence(t *testing.T) {
	// Create a temporary directory for test schemas
	tmpDir := t.TempDir()

	// Create a schema file that will be overridden
	dirSchema := SchemaDefinition{
		Name: "OverrideTest",
		Schema: map[string]interface{}{
			"class":      "FromDirectory",
			"vectorizer": "text2vec-weaviate",
		},
	}

	dirSchemaData, _ := yaml.Marshal(dirSchema)
	if err := os.WriteFile(filepath.Join(tmpDir, "override.yaml"), dirSchemaData, 0644); err != nil {
		t.Fatalf("Failed to write override.yaml: %v", err)
	}

	// Create config with inline schema with same name
	inlineSchema := SchemaDefinition{
		Name: "OverrideTest",
		Schema: map[string]interface{}{
			"class":      "FromConfig",
			"vectorizer": "text2vec-transformers",
		},
	}

	config := &Config{
		SchemasDir: tmpDir,
		Databases: DatabasesConfig{
			Schemas: []SchemaDefinition{inlineSchema},
		},
	}

	// Load schemas from directory
	err := config.loadSchemasFromDirectory()
	if err != nil {
		t.Fatalf("Failed to load schemas from directory: %v", err)
	}

	// Verify only one schema exists (inline took precedence)
	if len(config.Databases.Schemas) != 1 {
		t.Errorf("Expected 1 schema (inline precedence), got %d", len(config.Databases.Schemas))
	}

	// Verify it's the inline schema (class should be "FromConfig")
	schema := config.Databases.Schemas[0]
	if schemaMap, ok := schema.Schema["schema"].(map[string]interface{}); ok {
		if class, ok := schemaMap["class"].(string); ok && class != "FromConfig" {
			t.Errorf("Expected class 'FromConfig', got '%s' - inline schema should take precedence", class)
		}
	} else if class, ok := schema.Schema["class"].(string); !ok || class != "FromConfig" {
		t.Errorf("Expected class 'FromConfig', got '%v' - inline schema should take precedence", schema.Schema["class"])
	}
}

func TestLoadSchemasFromDirectory_NonExistent(t *testing.T) {
	// Create config with non-existent schemas_dir
	config := &Config{
		SchemasDir: "/non/existent/path",
		Databases: DatabasesConfig{
			Schemas: []SchemaDefinition{},
		},
	}

	// Should not error, just skip loading
	err := config.loadSchemasFromDirectory()
	if err != nil {
		t.Errorf("Expected no error for non-existent directory, got: %v", err)
	}

	// Should have no schemas
	if len(config.Databases.Schemas) != 0 {
		t.Errorf("Expected 0 schemas, got %d", len(config.Databases.Schemas))
	}
}

func TestGetSchema(t *testing.T) {
	config := &Config{
		Databases: DatabasesConfig{
			Schemas: []SchemaDefinition{
				{
					Name: "TestSchema",
					Schema: map[string]interface{}{
						"class": "TestClass",
					},
				},
			},
		},
	}

	// Test getting existing schema
	schema, err := config.GetSchema("TestSchema")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if schema == nil {
		t.Error("Expected schema, got nil")
		return
	}
	if schema.Name != "TestSchema" {
		t.Errorf("Expected schema name 'TestSchema', got '%s'", schema.Name)
	}

	// Test getting non-existent schema
	_, err = config.GetSchema("NonExistent")
	if err == nil {
		t.Error("Expected error for non-existent schema, got nil")
	}
}

func TestListSchemas(t *testing.T) {
	config := &Config{
		Databases: DatabasesConfig{
			Schemas: []SchemaDefinition{
				{Name: "Schema1"},
				{Name: "Schema2"},
				{Name: "Schema3"},
			},
		},
	}

	names := config.ListSchemas()
	if len(names) != 3 {
		t.Errorf("Expected 3 schema names, got %d", len(names))
	}

	expected := map[string]bool{
		"Schema1": true,
		"Schema2": true,
		"Schema3": true,
	}

	for _, name := range names {
		if !expected[name] {
			t.Errorf("Unexpected schema name: %s", name)
		}
	}
}

func TestGetAllSchemas(t *testing.T) {
	schemas := []SchemaDefinition{
		{Name: "Schema1"},
		{Name: "Schema2"},
	}

	config := &Config{
		Databases: DatabasesConfig{
			Schemas: schemas,
		},
	}

	allSchemas := config.GetAllSchemas()
	if len(allSchemas) != 2 {
		t.Errorf("Expected 2 schemas, got %d", len(allSchemas))
	}

	if allSchemas[0].Name != "Schema1" || allSchemas[1].Name != "Schema2" {
		t.Error("Schema names don't match expected values")
	}
}
func TestParseTimeoutToSeconds(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "Empty string returns default",
			input:    "",
			expected: DefaultTimeout,
		},
		{
			name:     "Duration string 5s",
			input:    "5s",
			expected: 5,
		},
		{
			name:     "Duration string 10s",
			input:    "10s",
			expected: 10,
		},
		{
			name:     "Duration string 1m",
			input:    "1m",
			expected: 60,
		},
		{
			name:     "Duration string 2m30s",
			input:    "2m30s",
			expected: 150,
		},
		{
			name:     "Integer seconds as string",
			input:    "30",
			expected: 30,
		},
		{
			name:     "Integer seconds large value",
			input:    "120",
			expected: 120,
		},
		{
			name:     "Zero duration returns default",
			input:    "0s",
			expected: DefaultTimeout,
		},
		{
			name:     "Negative duration returns default",
			input:    "-5s",
			expected: DefaultTimeout,
		},
		{
			name:     "Invalid string returns default",
			input:    "invalid",
			expected: DefaultTimeout,
		},
		{
			name:     "Zero integer returns default",
			input:    "0",
			expected: DefaultTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTimeoutToSeconds(tt.input)
			if result != tt.expected {
				t.Errorf("parseTimeoutToSeconds(%q) = %d, expected %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestInterpolateString(t *testing.T) {
	// Set up environment variables for testing
	os.Setenv("TEST_VAR", "test_value")
	os.Setenv("PORT", "8080")
	os.Setenv("HOST", "localhost")
	defer func() {
		os.Unsetenv("TEST_VAR")
		os.Unsetenv("PORT")
		os.Unsetenv("HOST")
	}()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No interpolation",
			input:    "plain string",
			expected: "plain string",
		},
		{
			name:     "Simple variable",
			input:    "${TEST_VAR}",
			expected: "test_value",
		},
		{
			name:     "Variable with default (var exists)",
			input:    "${TEST_VAR:-default}",
			expected: "test_value",
		},
		{
			name:     "Variable with default (var doesn't exist)",
			input:    "${NONEXISTENT:-default_value}",
			expected: "default_value",
		},
		{
			name:     "Variable in middle of string",
			input:    "http://${HOST}:${PORT}/api",
			expected: "http://localhost:8080/api",
		},
		{
			name:     "Multiple variables",
			input:    "${HOST} and ${PORT} and ${TEST_VAR}",
			expected: "localhost and 8080 and test_value",
		},
		{
			name:     "Empty default value",
			input:    "${NONEXISTENT:-}",
			expected: "",
		},
		{
			name:     "No closing brace",
			input:    "${TEST_VAR",
			expected: "${TEST_VAR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := InterpolateString(tt.input)
			if result != tt.expected {
				t.Errorf("InterpolateString(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetDefaultDatabase(t *testing.T) {
	tests := []struct {
		name         string
		config       *Config
		expectedName string
		shouldError  bool
	}{
		{
			name: "No databases configured",
			config: &Config{
				Databases: DatabasesConfig{
					VectorDatabases: []VectorDBConfig{},
				},
			},
			shouldError: true,
		},
		{
			name: "Single database (no default specified)",
			config: &Config{
				Databases: DatabasesConfig{
					VectorDatabases: []VectorDBConfig{
						{Name: "weaviate-cloud"},
					},
				},
			},
			expectedName: "weaviate-cloud",
			shouldError:  false,
		},
		{
			name: "Multiple databases with default specified",
			config: &Config{
				Databases: DatabasesConfig{
					Default: "milvus-cloud",
					VectorDatabases: []VectorDBConfig{
						{Name: "weaviate-cloud"},
						{Name: "milvus-cloud"},
						{Name: "qdrant-cloud"},
					},
				},
			},
			expectedName: "milvus-cloud",
			shouldError:  false,
		},
		{
			name: "Default not found (returns first)",
			config: &Config{
				Databases: DatabasesConfig{
					Default: "nonexistent",
					VectorDatabases: []VectorDBConfig{
						{Name: "weaviate-cloud"},
						{Name: "milvus-cloud"},
					},
				},
			},
			expectedName: "weaviate-cloud",
			shouldError:  false,
		},
		{
			name: "Empty default string (returns first)",
			config: &Config{
				Databases: DatabasesConfig{
					Default: "",
					VectorDatabases: []VectorDBConfig{
						{Name: "qdrant-cloud"},
					},
				},
			},
			expectedName: "qdrant-cloud",
			shouldError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := tt.config.GetDefaultDatabase()

			if tt.shouldError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if db == nil {
				t.Error("Expected database, got nil")
				return
			}

			if db.Name != tt.expectedName {
				t.Errorf("Expected database name %q, got %q", tt.expectedName, db.Name)
			}
		})
	}
}

func TestGetDatabase(t *testing.T) {
	config := &Config{
		Databases: DatabasesConfig{
			VectorDatabases: []VectorDBConfig{
				{Name: "weaviate-cloud"},
				{Name: "weaviate-local"},
				{Name: "milvus-cloud"},
				{Name: "qdrant-local"},
			},
		},
	}

	tests := []struct {
		name         string
		searchName   string
		expectedName string
		shouldError  bool
	}{
		{
			name:         "Exact match",
			searchName:   "weaviate-cloud",
			expectedName: "weaviate-cloud",
			shouldError:  false,
		},
		{
			name:         "Shortcut to cloud variant",
			searchName:   "weaviate",
			expectedName: "weaviate-cloud",
			shouldError:  false,
		},
		{
			name:         "Shortcut to cloud variant (milvus)",
			searchName:   "milvus",
			expectedName: "milvus-cloud",
			shouldError:  false,
		},
		{
			name:        "Not found",
			searchName:  "chroma",
			shouldError: true,
		},
		{
			name:         "Local variant (exact match takes precedence)",
			searchName:   "weaviate-local",
			expectedName: "weaviate-local",
			shouldError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := config.GetDatabase(tt.searchName)

			if tt.shouldError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if db == nil {
				t.Error("Expected database, got nil")
				return
			}

			if db.Name != tt.expectedName {
				t.Errorf("Expected database name %q, got %q", tt.expectedName, db.Name)
			}
		})
	}
}

func TestListDatabases(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected []string
	}{
		{
			name: "No databases",
			config: &Config{
				Databases: DatabasesConfig{
					VectorDatabases: []VectorDBConfig{},
				},
			},
			expected: []string{},
		},
		{
			name: "Single database",
			config: &Config{
				Databases: DatabasesConfig{
					VectorDatabases: []VectorDBConfig{
						{Name: "weaviate-cloud"},
					},
				},
			},
			expected: []string{"weaviate-cloud"},
		},
		{
			name: "Multiple databases",
			config: &Config{
				Databases: DatabasesConfig{
					VectorDatabases: []VectorDBConfig{
						{Name: "weaviate-cloud"},
						{Name: "milvus-cloud"},
						{Name: "qdrant-local"},
					},
				},
			},
			expected: []string{"weaviate-cloud", "milvus-cloud", "qdrant-local"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.ListDatabases()

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d databases, got %d", len(tt.expected), len(result))
				return
			}

			for i, name := range result {
				if name != tt.expected[i] {
					t.Errorf("Expected database[%d] = %q, got %q", i, tt.expected[i], name)
				}
			}
		})
	}
}
