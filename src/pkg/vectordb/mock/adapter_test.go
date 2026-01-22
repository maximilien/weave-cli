// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package mock

import (
	"context"
	"testing"

	mockClient "github.com/maximilien/weave-cli/src/pkg/mock"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

func TestNewAdapter(t *testing.T) {
	config := &vectordb.Config{
		Type:               vectordb.VectorDBTypeMock,
		Enabled:            true,
		EmbeddingDimension: 384,
		SimulateEmbeddings: true,
	}

	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	if adapter == nil {
		t.Fatal("Expected non-nil adapter")
	}

	if adapter.client == nil {
		t.Error("Expected non-nil mock client")
	}

	if adapter.config != config {
		t.Error("Expected adapter to store config reference")
	}
}

func TestHealth(t *testing.T) {
	config := &vectordb.Config{
		Type:               vectordb.VectorDBTypeMock,
		EmbeddingDimension: 384,
	}

	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	ctx := context.Background()
	err = adapter.Health(ctx)
	if err != nil {
		t.Errorf("Expected no error from Health, got: %v", err)
	}
}

func TestConvertDocument(t *testing.T) {
	adapter := createTestAdapter(t)

	tests := []struct {
		name     string
		input    *vectordb.Document
		expected *mockClient.Document
	}{
		{
			name:     "Nil document",
			input:    nil,
			expected: nil,
		},
		{
			name: "Basic document",
			input: &vectordb.Document{
				ID:      "doc1",
				Content: "test content",
			},
			expected: &mockClient.Document{
				ID:      "doc1",
				Content: "test content",
			},
		},
		{
			name: "Document with all fields",
			input: &vectordb.Document{
				ID:        "doc2",
				Content:   "full content",
				Image:     "image.png",
				ImageData: "base64encodeddata",
				URL:       "http://example.com",
				Metadata: map[string]interface{}{
					"key": "value",
				},
			},
			expected: &mockClient.Document{
				ID:        "doc2",
				Content:   "full content",
				Image:     "image.png",
				ImageData: "base64encodeddata",
				URL:       "http://example.com",
				Metadata: map[string]interface{}{
					"key": "value",
				},
			},
		},
		{
			name: "Document with empty metadata",
			input: &vectordb.Document{
				ID:       "doc3",
				Content:  "content",
				Metadata: map[string]interface{}{},
			},
			expected: &mockClient.Document{
				ID:       "doc3",
				Content:  "content",
				Metadata: map[string]interface{}{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adapter.convertDocument(tt.input)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, got %+v", result)
				}
				return
			}

			if result == nil {
				t.Fatal("Expected non-nil result")
			}

			if result.ID != tt.expected.ID {
				t.Errorf("ID: expected %q, got %q", tt.expected.ID, result.ID)
			}
			if result.Content != tt.expected.Content {
				t.Errorf("Content: expected %q, got %q", tt.expected.Content, result.Content)
			}
			if result.Image != tt.expected.Image {
				t.Errorf("Image: expected %q, got %q", tt.expected.Image, result.Image)
			}
			if result.URL != tt.expected.URL {
				t.Errorf("URL: expected %q, got %q", tt.expected.URL, result.URL)
			}
		})
	}
}

func TestConvertDocumentFromMock(t *testing.T) {
	adapter := createTestAdapter(t)

	tests := []struct {
		name     string
		input    *mockClient.Document
		expected *vectordb.Document
	}{
		{
			name:     "Nil document",
			input:    nil,
			expected: nil,
		},
		{
			name: "Basic document",
			input: &mockClient.Document{
				ID:      "doc1",
				Content: "test content",
			},
			expected: &vectordb.Document{
				ID:      "doc1",
				Text:    "test content",
				Content: "test content",
			},
		},
		{
			name: "Document with all fields",
			input: &mockClient.Document{
				ID:        "doc2",
				Content:   "full content",
				Image:     "image.png",
				ImageData: "base64encodeddata",
				URL:       "http://example.com",
				Metadata: map[string]interface{}{
					"key": "value",
				},
			},
			expected: &vectordb.Document{
				ID:        "doc2",
				Text:      "full content",
				Content:   "full content",
				Image:     "image.png",
				ImageData: "base64encodeddata",
				URL:       "http://example.com",
				Metadata: map[string]interface{}{
					"key": "value",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adapter.convertDocumentFromMock(tt.input)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, got %+v", result)
				}
				return
			}

			if result == nil {
				t.Fatal("Expected non-nil result")
			}

			if result.ID != tt.expected.ID {
				t.Errorf("ID: expected %q, got %q", tt.expected.ID, result.ID)
			}
			if result.Text != tt.expected.Text {
				t.Errorf("Text: expected %q, got %q", tt.expected.Text, result.Text)
			}
			if result.Content != tt.expected.Content {
				t.Errorf("Content: expected %q, got %q", tt.expected.Content, result.Content)
			}
			if result.Image != tt.expected.Image {
				t.Errorf("Image: expected %q, got %q", tt.expected.Image, result.Image)
			}
			if result.URL != tt.expected.URL {
				t.Errorf("URL: expected %q, got %q", tt.expected.URL, result.URL)
			}
		})
	}
}

func TestConvertDocumentsFromMock(t *testing.T) {
	adapter := createTestAdapter(t)

	tests := []struct {
		name          string
		input         []*mockClient.Document
		expectedCount int
	}{
		{
			name:          "Nil slice",
			input:         nil,
			expectedCount: 0,
		},
		{
			name:          "Empty slice",
			input:         []*mockClient.Document{},
			expectedCount: 0,
		},
		{
			name: "Single document",
			input: []*mockClient.Document{
				{ID: "doc1", Content: "content1"},
			},
			expectedCount: 1,
		},
		{
			name: "Multiple documents",
			input: []*mockClient.Document{
				{ID: "doc1", Content: "content1"},
				{ID: "doc2", Content: "content2"},
				{ID: "doc3", Content: "content3"},
			},
			expectedCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adapter.convertDocumentsFromMock(tt.input)

			if tt.input == nil {
				if result != nil {
					t.Errorf("Expected nil for nil input, got slice with %d elements", len(result))
				}
				return
			}

			if len(result) != tt.expectedCount {
				t.Errorf("Expected %d documents, got %d", tt.expectedCount, len(result))
				return
			}

			// Verify each document was converted correctly
			for i, doc := range result {
				if doc.ID != tt.input[i].ID {
					t.Errorf("Document[%d] ID: expected %q, got %q", i, tt.input[i].ID, doc.ID)
				}
				if doc.Content != tt.input[i].Content {
					t.Errorf("Document[%d] Content: expected %q, got %q", i, tt.input[i].Content, doc.Content)
				}
			}
		})
	}
}

func TestConvertQueryOptions(t *testing.T) {
	adapter := createTestAdapter(t)

	tests := []struct {
		name             string
		input            *vectordb.QueryOptions
		expectedTopK     int
		expectedDistance float64
	}{
		{
			name:             "Nil options",
			input:            nil,
			expectedTopK:     10,
			expectedDistance: 0.0,
		},
		{
			name: "Custom TopK and Distance",
			input: &vectordb.QueryOptions{
				TopK:     5,
				Distance: 0.8,
			},
			expectedTopK:     5,
			expectedDistance: 0.8,
		},
		{
			name: "Zero TopK (use default)",
			input: &vectordb.QueryOptions{
				TopK:     0,
				Distance: 0.5,
			},
			expectedTopK:     10,
			expectedDistance: 0.5,
		},
		{
			name: "Negative TopK (use default)",
			input: &vectordb.QueryOptions{
				TopK:     -1,
				Distance: 0.3,
			},
			expectedTopK:     10,
			expectedDistance: 0.3,
		},
		{
			name: "Large TopK",
			input: &vectordb.QueryOptions{
				TopK:     100,
				Distance: 0.9,
			},
			expectedTopK:     100,
			expectedDistance: 0.9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			topK, distance := adapter.convertQueryOptions(tt.input)

			if topK != tt.expectedTopK {
				t.Errorf("TopK: expected %d, got %d", tt.expectedTopK, topK)
			}
			if distance != tt.expectedDistance {
				t.Errorf("Distance: expected %f, got %f", tt.expectedDistance, distance)
			}
		})
	}
}

// Helper function to create a test adapter
func createTestAdapter(t *testing.T) *Adapter {
	config := &vectordb.Config{
		Type:               vectordb.VectorDBTypeMock,
		Enabled:            true,
		EmbeddingDimension: 384,
		SimulateEmbeddings: true,
	}

	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	return adapter
}
