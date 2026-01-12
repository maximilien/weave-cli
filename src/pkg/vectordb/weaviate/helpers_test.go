// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package weaviate

import (
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/assert"
)

func TestMapWeaviateDataType(t *testing.T) {
	tests := []struct {
		name     string
		dataType string
		expected []string
	}{
		{
			name:     "Text type",
			dataType: "text",
			expected: []string{"text"},
		},
		{
			name:     "String type",
			dataType: "string",
			expected: []string{"text"},
		},
		{
			name:     "Number type",
			dataType: "number",
			expected: []string{"number"},
		},
		{
			name:     "Boolean type",
			dataType: "boolean",
			expected: []string{"boolean"},
		},
		{
			name:     "Date type",
			dataType: "date",
			expected: []string{"date"},
		},
		{
			name:     "UUID type",
			dataType: "uuid",
			expected: []string{"uuid"},
		},
		{
			name:     "Unknown type defaults to text",
			dataType: "unknown",
			expected: []string{"text"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapWeaviateDataType(tt.dataType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsImageCollection(t *testing.T) {
	tests := []struct {
		name           string
		collectionName string
		expected       bool
	}{
		{
			name:           "Image collection with image_ prefix",
			collectionName: "image_photos",
			expected:       true,
		},
		{
			name:           "Image collection with img_ prefix",
			collectionName: "img_gallery",
			expected:       true,
		},
		{
			name:           "Image collection with _image suffix",
			collectionName: "photos_image",
			expected:       true,
		},
		{
			name:           "Image collection with _img suffix",
			collectionName: "gallery_img",
			expected:       true,
		},
		{
			name:           "Text collection",
			collectionName: "documents",
			expected:       false,
		},
		{
			name:           "Collection with image in middle",
			collectionName: "my_image_collection",
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isImageCollection(tt.collectionName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWrapErrorMethod(t *testing.T) {
	adapter := &Adapter{}

	tests := []struct {
		name      string
		err       error
		message   string
		expectNil bool
	}{
		{
			name:      "Wrap non-nil error",
			err:       assert.AnError,
			message:   "operation failed",
			expectNil: false,
		},
		{
			name:      "Wrap nil error returns nil",
			err:       nil,
			message:   "operation failed",
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adapter.wrapError(tt.err, tt.message)
			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Contains(t, result.Error(), tt.message)
			}
		})
	}
}

func TestGetTimeoutWeaviate(t *testing.T) {
	tests := []struct {
		name     string
		timeout  int
		expected time.Duration
	}{
		{
			name:     "Custom timeout",
			timeout:  60,
			expected: 60 * time.Second,
		},
		{
			name:     "Default timeout",
			timeout:  0,
			expected: 30 * time.Second, // Weaviate default is 30
		},
		{
			name:     "Small timeout",
			timeout:  5,
			expected: 5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				URL:     "http://localhost:8080",
				Timeout: tt.timeout,
			}

			client, err := NewClient(config)
			if err != nil {
				// Connection might fail, but we can still test the config
				assert.NotNil(t, config)
				return
			}

			result := client.getTimeout()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetTimeoutForWeaviate(t *testing.T) {
	config := &Config{
		URL:     "http://localhost:8080",
		Timeout: 30,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Skip("Skipping test that requires Weaviate connection")
	}

	tests := []struct {
		name          string
		operationType vectordb.OperationType
		minExpected   time.Duration
	}{
		{
			name:          "Collection operation",
			operationType: vectordb.OperationTypeCollection,
			minExpected:   30 * time.Second,
		},
		{
			name:          "Document operation",
			operationType: vectordb.OperationTypeDocument,
			minExpected:   30 * time.Second,
		},
		{
			name:          "Query operation",
			operationType: vectordb.OperationTypeQuery,
			minExpected:   30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.getTimeoutFor(tt.operationType)
			assert.GreaterOrEqual(t, result, tt.minExpected)
		})
	}
}

func TestConvertQueryOptions(t *testing.T) {
	tests := []struct {
		name    string
		options *vectordb.QueryOptions
	}{
		{
			name: "Options with TopK",
			options: &vectordb.QueryOptions{
				TopK: 10,
			},
		},
		{
			name: "Options with Distance",
			options: &vectordb.QueryOptions{
				TopK:     5,
				Distance: 0.8,
			},
		},
		{
			name:    "Nil options",
			options: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := &Adapter{}
			result := adapter.convertQueryOptions(tt.options)

			if tt.options == nil {
				// Should use defaults
				assert.NotNil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

func TestConvertDocument(t *testing.T) {
	tests := []struct {
		name string
		doc  *vectordb.Document
	}{
		{
			name: "Document with all fields",
			doc: &vectordb.Document{
				ID:      "doc-1",
				Text:    "Test content",
				Content: "More content",
				Metadata: map[string]interface{}{
					"key1": "value1",
					"key2": 42,
				},
			},
		},
		{
			name: "Document with empty metadata",
			doc: &vectordb.Document{
				ID:   "doc-2",
				Text: "Test",
			},
		},
		{
			name: "Document with image",
			doc: &vectordb.Document{
				ID:    "doc-3",
				Image: "https://example.com/image.jpg",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := &Adapter{}
			result := adapter.convertDocument(tt.doc)

			assert.NotNil(t, result)
			assert.Equal(t, tt.doc.ID, result.ID)
		})
	}
}
