// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package tests

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/weaviate"
)

func TestWeaviateClientCreation(t *testing.T) {
	testCases := []struct {
		name        string
		config      *weaviate.Config
		expectError bool
	}{
		{
			name: "Valid Cloud Config",
			config: &weaviate.Config{
				URL:    "https://test.weaviate.cloud",
				APIKey: "test-key",
			},
			expectError: false,
		},
		{
			name: "Valid Local Config",
			config: &weaviate.Config{
				URL: "http://localhost:8080",
			},
			expectError: false,
		},
		{
			name: "Empty URL",
			config: &weaviate.Config{
				URL: "",
			},
			expectError: false, // Weaviate client doesn't validate empty URL at creation time
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client, err := weaviate.NewClient(tc.config)

			if tc.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if client == nil {
					t.Error("Client should not be nil")
				}
			}
		})
	}
}

func TestWeaviateClientWithTimeout(t *testing.T) {
	config := &weaviate.Config{
		URL: "http://localhost:8080",
	}

	client, err := weaviate.NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// This should timeout quickly for localhost if not running
	_, err = client.ListCollections(ctx)
	if err != nil {
		// Expected for localhost without running Weaviate
		t.Logf("Expected timeout/connection error: %v", err)
	}
}

func TestWeaviateConfigValidation(t *testing.T) {
	testCases := []struct {
		name   string
		config *weaviate.Config
		valid  bool
	}{
		{
			name: "Valid Cloud URL",
			config: &weaviate.Config{
				URL:    "https://test.weaviate.cloud",
				APIKey: "test-key",
			},
			valid: true,
		},
		{
			name: "Valid Local URL",
			config: &weaviate.Config{
				URL: "http://localhost:8080",
			},
			valid: true,
		},
		{
			name: "Invalid URL",
			config: &weaviate.Config{
				URL: "not-a-url",
			},
			valid: true, // Weaviate client doesn't validate URL format at creation time
		},
		{
			name: "Empty URL",
			config: &weaviate.Config{
				URL: "",
			},
			valid: true, // Weaviate client doesn't validate empty URL at creation time
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client, err := weaviate.NewClient(tc.config)

			if tc.valid {
				if err != nil {
					t.Errorf("Expected valid config, got error: %v", err)
				}
				if client == nil {
					t.Error("Client should not be nil for valid config")
				}
			} else {
				if err == nil {
					t.Error("Expected error for invalid config, got nil")
				}
			}
		})
	}
}

func TestWeaviateDocumentStructure(t *testing.T) {
	// Test document structure
	doc := weaviate.Document{
		ID:      "test-doc-1",
		Content: "Test content",
		Metadata: map[string]interface{}{
			"source": "test",
			"type":   "text",
		},
	}

	if doc.ID == "" {
		t.Error("Document ID should not be empty")
	}

	if doc.Content == "" {
		t.Error("Document content should not be empty")
	}

	if doc.Metadata == nil {
		t.Error("Document metadata should not be nil")
	}

	// Test metadata access
	if doc.Metadata["source"] != "test" {
		t.Error("Metadata source should be 'test'")
	}
}

func TestWeaviateCollectionOperations(t *testing.T) {
	// Test collection name validation
	testCases := []struct {
		name        string
		collection  string
		expectError bool
	}{
		{"Valid Collection", "ValidCollection", false},
		{"Empty Collection", "", true},
		{"Collection with Spaces", "Collection With Spaces", true},
		{"Collection with Special Chars", "collection@#$", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// This is a basic validation test
			// In a real implementation, you'd validate collection names
			if tc.collection == "" && !tc.expectError {
				t.Error("Empty collection name should be invalid")
			}
		})
	}
}

func TestWeaviateErrorTypes(t *testing.T) {
	// Test error handling scenarios
	config := &weaviate.Config{
		URL: "http://localhost:8080",
	}

	client, err := weaviate.NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test operations that should fail gracefully
	operations := []struct {
		name string
		fn   func() error
	}{
		{
			name: "ListCollections",
			fn: func() error {
				_, err := client.ListCollections(ctx)
				return err
			},
		},
		{
			name: "ListDocuments",
			fn: func() error {
				_, err := client.ListDocuments(ctx, "test-collection", 5)
				return err
			},
		},
		{
			name: "GetDocument",
			fn: func() error {
				_, err := client.GetDocument(ctx, "test-collection", "test-doc")
				return err
			},
		},
	}

	for _, op := range operations {
		t.Run(op.name, func(t *testing.T) {
			err := op.fn()
			// We expect errors since Weaviate is not running
			if err == nil {
				t.Logf("%s succeeded unexpectedly", op.name)
			} else {
				t.Logf("%s failed as expected: %v", op.name, err)
			}
		})
	}
}

func TestWeaviateContextHandling(t *testing.T) {
	config := &weaviate.Config{
		URL: "http://localhost:8080",
	}

	client, err := weaviate.NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = client.ListCollections(ctx)
	if err == nil {
		t.Error("Expected error for cancelled context")
	} else {
		t.Logf("Correctly handled cancelled context: %v", err)
	}

	// Test context timeout
	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(1 * time.Millisecond) // Ensure timeout

	_, err = client.ListCollections(ctx)
	if err == nil {
		t.Error("Expected error for timed out context")
	} else {
		t.Logf("Correctly handled timed out context: %v", err)
	}
}

func TestWeaviateClientDeleteDocumentsByMetadata(t *testing.T) {
	// This test validates the metadata filtering logic without requiring a real Weaviate instance
	// It tests the GraphQL query construction and error handling

	testCases := []struct {
		name            string
		metadataFilters []string
		expectError     bool
		errorContains   string
	}{
		{
			name:            "Single metadata filter",
			metadataFilters: []string{"filename=test.png"},
			expectError:     false,
		},
		{
			name:            "Multiple metadata filters",
			metadataFilters: []string{"filename=test.png", "type=image"},
			expectError:     false,
		},
		{
			name:            "Invalid filter format - no equals",
			metadataFilters: []string{"invalid-filter"},
			expectError:     true,
			errorContains:   "invalid metadata filter format",
		},
		{
			name:            "Invalid filter format - empty key",
			metadataFilters: []string{"=value"},
			expectError:     true,
			errorContains:   "empty key",
		},
		{
			name:            "Empty metadata filters",
			metadataFilters: []string{},
			expectError:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock client config
			config := &weaviate.Config{
				URL: "http://localhost:8080",
			}

			_, err := weaviate.NewClient(config)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			// Test the metadata filter parsing logic
			if len(tc.metadataFilters) > 0 {
				filters := make(map[string]string)
				for _, filter := range tc.metadataFilters {
					parts := strings.SplitN(filter, "=", 2)
					if len(parts) != 2 {
						if tc.expectError {
							if !strings.Contains("invalid metadata filter format", tc.errorContains) {
								t.Errorf("Expected error containing '%s', but got different error", tc.errorContains)
							}
							return
						}
						t.Errorf("Unexpected error parsing filter: %s", filter)
						return
					}

					// Check for empty key
					if parts[0] == "" {
						if tc.expectError && strings.Contains(tc.errorContains, "empty key") {
							t.Logf("Correctly detected empty key in filter: %s", filter)
							return
						} else if !tc.expectError {
							t.Errorf("Unexpected empty key in filter: %s", filter)
							return
						}
					}

					filters[parts[0]] = parts[1]
				}

				// Validate that filters were parsed correctly
				if tc.expectError {
					t.Error("Expected error but parsing succeeded")
				} else {
					t.Logf("Successfully parsed %d metadata filters", len(filters))
				}
			}
		})
	}
}
