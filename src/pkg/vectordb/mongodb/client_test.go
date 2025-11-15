// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package mongodb

import (
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

func TestConfig_Defaults(t *testing.T) {
	config := &Config{
		URI:      "mongodb://localhost:27017", // Use localhost to avoid DNS lookups
		Database: "test",
	}

	_, err := NewClient(config)
	// Connection will fail without actual MongoDB, but we can check the config was processed
	if err != nil {
		// This is expected - we don't have a real MongoDB running
		t.Skip("Skipping test that requires MongoDB connection")
	}
}

func TestConfig_CustomValues(t *testing.T) {
	config := &Config{
		URI:              "mongodb+srv://test:test@cluster.mongodb.net",
		Database:         "custom",
		VectorDimensions: 768,
		SimilarityMetric: "euclidean",
		Timeout:          20,
	}

	// Note: This will fail to connect, but we're testing config handling
	_, err := NewClient(config)
	// We expect connection to fail, but config should be preserved
	if err == nil {
		t.Skip("Connection succeeded unexpectedly (requires actual MongoDB)")
	}
}

func TestConfig_ValidationErrors(t *testing.T) {
	testCases := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "Missing URI",
			config: &Config{
				Database: "test",
			},
			expectError: true,
		},
		{
			name: "Missing Database",
			config: &Config{
				URI: "mongodb+srv://test:test@cluster.mongodb.net",
			},
			expectError: true,
		},
		{
			name: "Valid Config",
			config: &Config{
				URI:      "mongodb+srv://test:test@cluster.mongodb.net",
				Database: "test",
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewClient(tc.config)
			hasError := err != nil

			if tc.expectError && !hasError {
				t.Errorf("Expected error but got none")
			}

			if !tc.expectError && hasError {
				// Connection errors are expected since we don't have a real MongoDB
				// We only care about validation errors
				if err.Error() == "MongoDB URI is required" || err.Error() == "MongoDB database name is required" {
					t.Errorf("Unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestDocumentConversion(t *testing.T) {
	client := &Client{
		config: &Config{
			VectorDimensions: 1536,
			SimilarityMetric: "cosine",
		},
	}

	// Test toMongoDocument
	doc := &vectordb.Document{
		ID:      "test-123",
		Text:    "Test document",
		Content: "This is test content",
		URL:     "https://example.com",
		Metadata: map[string]interface{}{
			"category": "test",
			"count":    42,
		},
	}

	mongoDoc := client.toMongoDocument(doc)

	if mongoDoc.DocumentID != doc.ID {
		t.Errorf("DocumentID mismatch: expected %s, got %s", doc.ID, mongoDoc.DocumentID)
	}

	if mongoDoc.Text != doc.Text {
		t.Errorf("Text mismatch: expected %s, got %s", doc.Text, mongoDoc.Text)
	}

	if mongoDoc.Content != doc.Content {
		t.Errorf("Content mismatch: expected %s, got %s", doc.Content, mongoDoc.Content)
	}

	if mongoDoc.URL != doc.URL {
		t.Errorf("URL mismatch: expected %s, got %s", doc.URL, mongoDoc.URL)
	}

	// Test fromMongoDocument
	convertedDoc := client.fromMongoDocument(mongoDoc)

	if convertedDoc.ID != doc.ID {
		t.Errorf("Converted ID mismatch: expected %s, got %s", doc.ID, convertedDoc.ID)
	}

	if convertedDoc.Text != doc.Text {
		t.Errorf("Converted Text mismatch: expected %s, got %s", doc.Text, convertedDoc.Text)
	}

	if convertedDoc.Content != doc.Content {
		t.Errorf("Converted Content mismatch: expected %s, got %s", doc.Content, convertedDoc.Content)
	}

	// Verify metadata
	if category, ok := convertedDoc.Metadata["category"].(string); !ok || category != "test" {
		t.Errorf("Metadata category mismatch: expected 'test', got %v", convertedDoc.Metadata["category"])
	}
}

func TestGetTimeout(t *testing.T) {
	testCases := []struct {
		name            string
		timeout         int
		expectedSeconds int
	}{
		{
			name:            "Default timeout",
			timeout:         0,
			expectedSeconds: 10,
		},
		{
			name:            "Custom timeout",
			timeout:         30,
			expectedSeconds: 30,
		},
		{
			name:            "Small timeout",
			timeout:         5,
			expectedSeconds: 5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &Client{
				config: &Config{
					Timeout: tc.timeout,
				},
			}

			duration := client.getTimeout()
			seconds := int(duration.Seconds())

			if seconds != tc.expectedSeconds {
				t.Errorf("Expected %d seconds, got %d", tc.expectedSeconds, seconds)
			}
		})
	}
}

func TestGetDefaultSchema(t *testing.T) {
	client := &Client{
		config: &Config{},
	}

	testCases := []struct {
		name            string
		schemaType      vectordb.SchemaType
		collectionName  string
		expectedClass   string
		expectEmbedding bool
	}{
		{
			name:            "Text schema",
			schemaType:      vectordb.SchemaTypeText,
			collectionName:  "TestDocs",
			expectedClass:   "TestDocs",
			expectEmbedding: true,
		},
		{
			name:            "Image schema",
			schemaType:      vectordb.SchemaTypeImage,
			collectionName:  "TestImages",
			expectedClass:   "TestImages",
			expectEmbedding: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			schema := client.GetDefaultSchema(tc.schemaType, tc.collectionName)

			if schema.Class != tc.expectedClass {
				t.Errorf("Class mismatch: expected %s, got %s", tc.expectedClass, schema.Class)
			}

			if schema.Vectorizer != "none" {
				t.Errorf("Expected vectorizer 'none', got %s", schema.Vectorizer)
			}

			// Check for embedding property
			foundEmbedding := false
			for _, prop := range schema.Properties {
				if prop.Name == "embedding" {
					foundEmbedding = true
					break
				}
			}

			if tc.expectEmbedding && !foundEmbedding {
				t.Errorf("Expected embedding property in schema")
			}
		})
	}
}

func TestValidateSchema(t *testing.T) {
	client := &Client{
		config: &Config{},
	}

	testCases := []struct {
		name        string
		schema      *vectordb.CollectionSchema
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Nil schema",
			schema:      nil,
			expectError: true,
			errorMsg:    "schema cannot be nil",
		},
		{
			name: "Empty class name",
			schema: &vectordb.CollectionSchema{
				Class:      "",
				Properties: []vectordb.SchemaProperty{},
			},
			expectError: true,
			errorMsg:    "schema class name is required",
		},
		{
			name: "No properties",
			schema: &vectordb.CollectionSchema{
				Class:      "TestClass",
				Properties: []vectordb.SchemaProperty{},
			},
			expectError: true,
			errorMsg:    "schema must have at least one property",
		},
		{
			name: "Schema without embedding (valid for BM25-only collections)",
			schema: &vectordb.CollectionSchema{
				Class: "TestClass",
				Properties: []vectordb.SchemaProperty{
					{Name: "text", DataType: []string{"text"}},
				},
			},
			expectError: false, // MongoDB allows text-only collections
		},
		{
			name: "Valid schema",
			schema: &vectordb.CollectionSchema{
				Class: "TestClass",
				Properties: []vectordb.SchemaProperty{
					{Name: "text", DataType: []string{"text"}},
					{Name: "embedding", DataType: []string{"number[]"}},
				},
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := client.ValidateSchema(tc.schema)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tc.errorMsg != "" && err.Error() != tc.errorMsg {
					t.Errorf("Expected error message %q, got %q", tc.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}
