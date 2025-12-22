// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package tests

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/agents"
	"github.com/maximilien/weave-cli/src/pkg/llm"
	"gopkg.in/yaml.v3"
)

// TestSchemaAgentInitialization tests SchemaAgent creation
func TestSchemaAgentInitialization(t *testing.T) {
	// Skip if no OpenAI API key (CI environments)
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping: OPENAI_API_KEY not set")
	}

	t.Run("CreateAgent", func(t *testing.T) {
		llmClient, err := llm.NewOpenAIClient(apiKey)
		if err != nil {
			t.Fatalf("Failed to create LLM client: %v", err)
		}

		agent := agents.NewSchemaAgent(llmClient)
		if agent == nil {
			t.Fatal("SchemaAgent should be created")
		}

		if agent.Name() != "schema-agent" {
			t.Errorf("Expected agent name 'schema-agent', got '%s'", agent.Name())
		}
	})

	t.Run("InvalidAPIKey", func(t *testing.T) {
		_, err := llm.NewOpenAIClient("")
		if err == nil {
			t.Fatal("Should fail with empty API key")
		}
	})
}

// TestSchemaAnalysisWithSamples tests schema analysis with document samples
func TestSchemaAnalysisWithSamples(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping: OPENAI_API_KEY not set")
	}

	// Create temporary directory with test files
	tmpDir := t.TempDir()

	// Create sample JSON files
	samples := []struct {
		filename string
		content  map[string]interface{}
	}{
		{
			"product1.json",
			map[string]interface{}{
				"id":          "p001",
				"name":        "Test Product 1",
				"price":       99.99,
				"category":    "Electronics",
				"in_stock":    true,
				"created_at":  "2025-01-01T10:00:00Z",
				"description": "A test product for schema analysis",
			},
		},
		{
			"product2.json",
			map[string]interface{}{
				"id":          "p002",
				"name":        "Test Product 2",
				"price":       149.99,
				"category":    "Books",
				"in_stock":    false,
				"created_at":  "2025-01-02T12:00:00Z",
				"description": "Another test product",
			},
		},
	}

	var sampleFiles []string
	for _, sample := range samples {
		data, err := json.MarshalIndent(sample.content, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal JSON: %v", err)
		}

		path := filepath.Join(tmpDir, sample.filename)
		err = os.WriteFile(path, data, 0644)
		if err != nil {
			t.Fatalf("Failed to write sample file: %v", err)
		}

		sampleFiles = append(sampleFiles, path)
	}

	t.Run("AnalyzeSamples", func(t *testing.T) {
		llmClient, err := llm.NewOpenAIClient(apiKey)
		if err != nil {
			t.Fatalf("Failed to create LLM client: %v", err)
		}

		agent := agents.NewSchemaAgent(llmClient)

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		input := &agents.SchemaAnalysisInput{
			SampleFiles:    sampleFiles,
			CollectionName: "test_products",
			Requirements:   "Support filtering by price and category",
			VDBType:        "weaviate",
			MaxSamples:     10,
		}

		result, err := agent.Execute(ctx, input)
		if err != nil {
			t.Fatalf("Schema analysis failed: %v", err)
		}

		output, ok := result.(*agents.SchemaAnalysisOutput)
		if !ok {
			t.Fatal("Invalid output type from schema agent")
		}

		// Verify schema properties
		if output.Schema.CollectionName != "test_products" {
			t.Errorf("Expected collection name 'test_products', got '%s'", output.Schema.CollectionName)
		}

		if output.Schema.VectorDimensions != 1536 {
			t.Errorf("Expected 1536 vector dimensions, got %d", output.Schema.VectorDimensions)
		}

		if output.Schema.SimilarityMetric != "cosine" {
			t.Errorf("Expected cosine similarity, got '%s'", output.Schema.SimilarityMetric)
		}

		// Verify fields were detected
		if len(output.Schema.Fields) == 0 {
			t.Fatal("Schema should have detected fields")
		}

		// Verify confidence score
		if output.Confidence <= 0 || output.Confidence > 1 {
			t.Errorf("Confidence should be between 0 and 1, got %f", output.Confidence)
		}

		// Verify reasoning is provided
		if output.Reasoning == "" {
			t.Error("Schema analysis should provide reasoning")
		}

		// Log results for manual inspection
		t.Logf("Detected %d fields", len(output.Schema.Fields))
		t.Logf("Confidence: %.2f%%", output.Confidence*100)
		t.Logf("Reasoning: %s", output.Reasoning)
	})
}

// TestSchemaYAMLOutput tests YAML schema output generation
func TestSchemaYAMLOutput(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a sample schema config
	schema := agents.SchemaConfig{
		CollectionName:   "test_collection",
		VectorDimensions: 1536,
		SimilarityMetric: "cosine",
		Fields: []agents.FieldConfig{
			{
				Name:        "id",
				Type:        "text",
				Description: "Unique identifier",
				Indexed:     true,
				Filterable:  true,
				Required:    true,
			},
			{
				Name:        "content",
				Type:        "text",
				Description: "Document content",
				Indexed:     false,
				Filterable:  false,
				Required:    true,
			},
			{
				Name:        "price",
				Type:        "number",
				Description: "Product price",
				Indexed:     true,
				Filterable:  true,
				Required:    false,
			},
		},
		Indexes: []agents.IndexConfig{
			{
				Name:   "vector_index",
				Fields: []string{"content"},
				Type:   "vector",
			},
			{
				Name:   "bm25_index",
				Fields: []string{"content"},
				Type:   "bm25",
			},
		},
		Metadata: map[string]string{
			"version": "1.0",
			"created": "2025-01-01",
		},
	}

	t.Run("MarshalToYAML", func(t *testing.T) {
		data, err := yaml.Marshal(schema)
		if err != nil {
			t.Fatalf("Failed to marshal schema to YAML: %v", err)
		}

		if len(data) == 0 {
			t.Fatal("YAML output should not be empty")
		}

		// Verify YAML contains expected fields
		yamlStr := string(data)
		expectedFields := []string{
			"collection_name: test_collection",
			"vector_dimensions: 1536",
			"similarity_metric: cosine",
		}

		for _, field := range expectedFields {
			if !stringContains(yamlStr, field) {
				t.Errorf("YAML should contain '%s'", field)
			}
		}
	})

	t.Run("WriteYAMLFile", func(t *testing.T) {
		outputFile := filepath.Join(tmpDir, "test-schema.yaml")

		data, err := yaml.Marshal(schema)
		if err != nil {
			t.Fatalf("Failed to marshal: %v", err)
		}

		// Add header like the actual implementation
		header := "# Vector Database Schema Configuration\n# Collection: test_collection\n\n"
		content := header + string(data)

		err = os.WriteFile(outputFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to write YAML file: %v", err)
		}

		// Verify file exists and is readable
		readData, err := os.ReadFile(outputFile)
		if err != nil {
			t.Fatalf("Failed to read YAML file: %v", err)
		}

		if len(readData) == 0 {
			t.Fatal("Written YAML file should not be empty")
		}

		// Verify header is present
		if !stringContains(string(readData), "# Vector Database Schema Configuration") {
			t.Error("YAML file should contain header comment")
		}
	})

	t.Run("UnmarshalFromYAML", func(t *testing.T) {
		// Marshal and then unmarshal to verify round-trip
		data, err := yaml.Marshal(schema)
		if err != nil {
			t.Fatalf("Failed to marshal: %v", err)
		}

		var unmarshaled agents.SchemaConfig
		err = yaml.Unmarshal(data, &unmarshaled)
		if err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		// Verify key fields
		if unmarshaled.CollectionName != schema.CollectionName {
			t.Errorf("Collection name mismatch after unmarshal")
		}

		if unmarshaled.VectorDimensions != schema.VectorDimensions {
			t.Errorf("Vector dimensions mismatch after unmarshal")
		}

		if len(unmarshaled.Fields) != len(schema.Fields) {
			t.Errorf("Expected %d fields, got %d after unmarshal", len(schema.Fields), len(unmarshaled.Fields))
		}
	})
}

// TestSchemaDocumentSampling tests document sampling functionality
func TestSchemaDocumentSampling(t *testing.T) {
	tmpDir := t.TempDir()

	// Create various file types
	testFiles := map[string]string{
		"doc1.txt":  "This is a plain text document for testing schema analysis.",
		"doc2.md":   "# Markdown Document\n\nThis is a markdown file with content.",
		"data.json": `{"id": "123", "name": "Test", "value": 42}`,
	}

	for filename, content := range testFiles {
		path := filepath.Join(tmpDir, filename)
		err := os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}
	}

	t.Run("ScanDirectory", func(t *testing.T) {
		// Scan for all supported file types
		files, err := filepath.Glob(filepath.Join(tmpDir, "*"))
		if err != nil {
			t.Fatalf("Failed to scan directory: %v", err)
		}

		if len(files) != len(testFiles) {
			t.Errorf("Expected %d files, found %d", len(testFiles), len(files))
		}
	})

	t.Run("FilterByExtension", func(t *testing.T) {
		// Scan for JSON files only
		jsonFiles, err := filepath.Glob(filepath.Join(tmpDir, "*.json"))
		if err != nil {
			t.Fatalf("Failed to scan for JSON files: %v", err)
		}

		if len(jsonFiles) != 1 {
			t.Errorf("Expected 1 JSON file, found %d", len(jsonFiles))
		}
	})
}

// Helper function to check if string contains substring
func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
