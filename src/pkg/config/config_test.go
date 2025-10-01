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
