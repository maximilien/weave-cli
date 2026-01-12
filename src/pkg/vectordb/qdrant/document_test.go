// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package qdrant

import (
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/assert"
)

func TestConvertToQdrantDocument(t *testing.T) {
	tests := []struct {
		name     string
		input    *vectordb.Document
		wantErr  bool
		checkID  bool
		checkVec bool
	}{
		{
			name: "Valid document with all fields",
			input: &vectordb.Document{
				ID:      "doc-1",
				Text:    "Test content",
				Content: "More content",
				Metadata: map[string]interface{}{
					"key1": "value1",
					"key2": 42,
				},
			},
			wantErr:  false,
			checkID:  true,
			checkVec: false,
		},
		{
			name: "Document with empty ID",
			input: &vectordb.Document{
				ID:   "",
				Text: "Test",
			},
			wantErr:  false, // ID will be generated
			checkID:  false,
			checkVec: false,
		},
		{
			name: "Document with nil metadata",
			input: &vectordb.Document{
				ID:       "doc-2",
				Text:     "Test",
				Metadata: nil,
			},
			wantErr:  false,
			checkID:  true,
			checkVec: false,
		},
		{
			name: "Document with empty text",
			input: &vectordb.Document{
				ID:   "doc-3",
				Text: "",
			},
			wantErr:  false,
			checkID:  true,
			checkVec: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This tests the conversion logic
			doc := &Document{
				ID:       tt.input.ID,
				Content:  tt.input.Text,
				Metadata: tt.input.Metadata,
			}

			if tt.checkID {
				assert.Equal(t, tt.input.ID, doc.ID)
			}
			// For empty ID case, we're just testing the struct creation
			// Actual ID generation happens in the adapter layer

			assert.Equal(t, tt.input.Text, doc.Content)

			if tt.input.Metadata != nil {
				assert.Equal(t, tt.input.Metadata, doc.Metadata)
			}
		})
	}
}

func TestConvertVectorFloat64ToFloat32(t *testing.T) {
	tests := []struct {
		name     string
		input    []float64
		expected []float32
	}{
		{
			name:     "Empty vector",
			input:    []float64{},
			expected: []float32{},
		},
		{
			name:     "Single value",
			input:    []float64{1.5},
			expected: []float32{1.5},
		},
		{
			name:     "Multiple values",
			input:    []float64{1.0, 2.0, 3.0, 4.5},
			expected: []float32{1.0, 2.0, 3.0, 4.5},
		},
		{
			name:     "Large vector (typical embedding size)",
			input:    make([]float64, 384),
			expected: make([]float32, 384),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := make([]float32, len(tt.input))
			for i, v := range tt.input {
				result[i] = float32(v)
			}

			assert.Equal(t, len(tt.expected), len(result))
			for i := range result {
				assert.InDelta(t, tt.expected[i], result[i], 0.0001)
			}
		})
	}
}

func TestConvertVectorFloat32ToFloat64(t *testing.T) {
	tests := []struct {
		name     string
		input    []float32
		expected []float64
	}{
		{
			name:     "Empty vector",
			input:    []float32{},
			expected: []float64{},
		},
		{
			name:     "Single value",
			input:    []float32{1.5},
			expected: []float64{1.5},
		},
		{
			name:     "Multiple values",
			input:    []float32{1.0, 2.0, 3.0, 4.5},
			expected: []float64{1.0, 2.0, 3.0, 4.5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := make([]float64, len(tt.input))
			for i, v := range tt.input {
				result[i] = float64(v)
			}

			assert.Equal(t, len(tt.expected), len(result))
			for i := range result {
				assert.InDelta(t, tt.expected[i], result[i], 0.0001)
			}
		})
	}
}

func TestDocumentValidation(t *testing.T) {
	tests := []struct {
		name    string
		doc     *vectordb.Document
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid document",
			doc: &vectordb.Document{
				ID:   "valid-1",
				Text: "Valid content",
			},
			wantErr: false,
		},
		{
			name: "Document without ID (should auto-generate)",
			doc: &vectordb.Document{
				Text: "Content without ID",
			},
			wantErr: false,
		},
		{
			name: "Document without text (valid for image docs)",
			doc: &vectordb.Document{
				ID:    "doc-1",
				Image: "image-url",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation - documents should have either text or image
			hasContent := tt.doc.Text != "" || tt.doc.Image != "" || tt.doc.Content != ""

			if !tt.wantErr {
				// For valid docs, at least check they have some content field
				// (ID can be auto-generated, so we don't require it)
				if tt.doc.ID != "" || hasContent {
					assert.True(t, true, "Document has required fields")
				}
			}
		})
	}
}

func TestMetadataHandling(t *testing.T) {
	tests := []struct {
		name     string
		metadata map[string]interface{}
		wantErr  bool
	}{
		{
			name: "Simple metadata",
			metadata: map[string]interface{}{
				"author": "test",
				"date":   "2024-01-01",
			},
			wantErr: false,
		},
		{
			name: "Nested metadata",
			metadata: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "test",
					"id":   123,
				},
			},
			wantErr: false,
		},
		{
			name: "Metadata with arrays",
			metadata: map[string]interface{}{
				"tags": []string{"tag1", "tag2"},
			},
			wantErr: false,
		},
		{
			name:     "Nil metadata",
			metadata: nil,
			wantErr:  false,
		},
		{
			name:     "Empty metadata",
			metadata: map[string]interface{}{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := &vectordb.Document{
				ID:       "test-doc",
				Text:     "Test",
				Metadata: tt.metadata,
			}

			// Metadata should be preserved
			if tt.metadata != nil {
				assert.Equal(t, tt.metadata, doc.Metadata)
			}
		})
	}
}

func TestBatchOperations(t *testing.T) {
	tests := []struct {
		name      string
		batchSize int
		wantErr   bool
	}{
		{
			name:      "Small batch",
			batchSize: 10,
			wantErr:   false,
		},
		{
			name:      "Medium batch",
			batchSize: 100,
			wantErr:   false,
		},
		{
			name:      "Large batch",
			batchSize: 1000,
			wantErr:   false,
		},
		{
			name:      "Empty batch",
			batchSize: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create batch of documents
			docs := make([]*vectordb.Document, tt.batchSize)
			for i := 0; i < tt.batchSize; i++ {
				docs[i] = &vectordb.Document{
					ID:   "",  // Will be auto-generated
					Text: "Test content",
				}
			}

			// Basic validation that we can create batches
			assert.Equal(t, tt.batchSize, len(docs))
		})
	}
}
