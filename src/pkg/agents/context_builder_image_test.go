// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package agents

import (
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExtractContent_ImageWithFullMetadata tests image extraction with complete metadata
func TestExtractContent_ImageWithFullMetadata(t *testing.T) {
	config := &CustomAgentConfig{
		Response: CustomAgentResponseConfig{
			MaxContextChunks:   10,
			MinRelevanceScore:  0.0,
			DeduplicateSources: false,
			SortByRelevance:    false,
		},
		Output: CustomAgentOutputConfig{
			IncludeMetadata: true,
		},
	}

	cb := NewContextBuilder(config)

	doc := vectordb.Document{
		ID:    "img-001",
		Image: "https://cdn.example.com/vintage-car.jpg",
		Metadata: map[string]interface{}{
			"ocr_text":    "1967 Ford Mustang",
			"description": "Vintage red Mustang in excellent condition",
			"alt_text":    "Classic 1967 Ford Mustang automobile",
			"caption":     "Restored vintage Mustang from the 1960s",
			"tags":        []interface{}{"vintage", "car", "mustang", "1967"},
		},
	}

	content := cb.extractContent(doc)

	// Should contain all metadata fields
	assert.Contains(t, content, "Text in image: 1967 Ford Mustang")
	assert.Contains(t, content, "Description: Vintage red Mustang in excellent condition")
	assert.Contains(t, content, "Alt text: Classic 1967 Ford Mustang automobile")
	assert.Contains(t, content, "Caption: Restored vintage Mustang from the 1960s")
	assert.Contains(t, content, "Tags: vintage, car, mustang, 1967")
	assert.Contains(t, content, "Image URL: https://cdn.example.com/vintage-car.jpg")

	// Should NOT contain empty document marker
	assert.NotContains(t, content, "[Empty document]")
}

// TestExtractContent_ImageWithOCROnly tests image with only OCR text
func TestExtractContent_ImageWithOCROnly(t *testing.T) {
	config := &CustomAgentConfig{
		Response: CustomAgentResponseConfig{
			MaxContextChunks:   10,
			MinRelevanceScore:  0.0,
			DeduplicateSources: false,
			SortByRelevance:    false,
		},
	}

	cb := NewContextBuilder(config)

	doc := vectordb.Document{
		ID:    "img-002",
		Image: "https://cdn.example.com/document.jpg",
		Metadata: map[string]interface{}{
			"ocr_text": "Important document text extracted via OCR",
		},
	}

	content := cb.extractContent(doc)

	assert.Contains(t, content, "Text in image: Important document text extracted via OCR")
	assert.Contains(t, content, "Image URL: https://cdn.example.com/document.jpg")
	assert.NotContains(t, content, "[Empty document]")
}

// TestExtractContent_ImageWithDescriptionOnly tests image with only description
func TestExtractContent_ImageWithDescriptionOnly(t *testing.T) {
	config := &CustomAgentConfig{
		Response: CustomAgentResponseConfig{
			MaxContextChunks:   10,
			MinRelevanceScore:  0.0,
			DeduplicateSources: false,
			SortByRelevance:    false,
		},
	}

	cb := NewContextBuilder(config)

	doc := vectordb.Document{
		ID:    "img-003",
		Image: "https://cdn.example.com/artwork.jpg",
		Metadata: map[string]interface{}{
			"description": "Abstract painting with vibrant colors",
		},
	}

	content := cb.extractContent(doc)

	assert.Contains(t, content, "Description: Abstract painting with vibrant colors")
	assert.Contains(t, content, "Image URL: https://cdn.example.com/artwork.jpg")
	assert.NotContains(t, content, "[Empty document]")
}

// TestExtractContent_ImageWithTagsAsStringSlice tests tags as []string
func TestExtractContent_ImageWithTagsAsStringSlice(t *testing.T) {
	config := &CustomAgentConfig{
		Response: CustomAgentResponseConfig{
			MaxContextChunks:   10,
			MinRelevanceScore:  0.0,
			DeduplicateSources: false,
			SortByRelevance:    false,
		},
	}

	cb := NewContextBuilder(config)

	doc := vectordb.Document{
		ID:    "img-004",
		Image: "https://cdn.example.com/product.jpg",
		Metadata: map[string]interface{}{
			"tags": []string{"product", "blue", "widget"},
		},
	}

	content := cb.extractContent(doc)

	assert.Contains(t, content, "Tags: product, blue, widget")
	assert.Contains(t, content, "Image URL: https://cdn.example.com/product.jpg")
}

// TestExtractContent_ImageWithoutMetadata tests image without any metadata
func TestExtractContent_ImageWithoutMetadata(t *testing.T) {
	config := &CustomAgentConfig{
		Response: CustomAgentResponseConfig{
			MaxContextChunks:   10,
			MinRelevanceScore:  0.0,
			DeduplicateSources: false,
			SortByRelevance:    false,
		},
	}

	cb := NewContextBuilder(config)

	doc := vectordb.Document{
		ID:       "img-005",
		Image:    "https://cdn.example.com/photo.jpg",
		Metadata: map[string]interface{}{},
	}

	content := cb.extractContent(doc)

	// Should fall back to just showing image URL
	assert.Equal(t, "Image URL: https://cdn.example.com/photo.jpg", content)
	assert.NotContains(t, content, "[Empty document]")
}

// TestExtractContent_ImageDataOnly tests base64 image data without URL
func TestExtractContent_ImageDataOnly(t *testing.T) {
	config := &CustomAgentConfig{
		Response: CustomAgentResponseConfig{
			MaxContextChunks:   10,
			MinRelevanceScore:  0.0,
			DeduplicateSources: false,
			SortByRelevance:    false,
		},
	}

	cb := NewContextBuilder(config)

	doc := vectordb.Document{
		ID:        "img-006",
		ImageData: "base64encodedimagedata...",
		Metadata:  map[string]interface{}{},
	}

	content := cb.extractContent(doc)

	// Should indicate it's base64 image data
	assert.Equal(t, "[Image data (base64)]", content)
	assert.NotContains(t, content, "[Empty document]")
}

// TestExtractContent_TextPriorityOverImage tests that text content takes priority
func TestExtractContent_TextPriorityOverImage(t *testing.T) {
	config := &CustomAgentConfig{
		Response: CustomAgentResponseConfig{
			MaxContextChunks:   10,
			MinRelevanceScore:  0.0,
			DeduplicateSources: false,
			SortByRelevance:    false,
		},
	}

	cb := NewContextBuilder(config)

	doc := vectordb.Document{
		ID:      "doc-001",
		Content: "This is text content",
		Image:   "https://cdn.example.com/image.jpg",
		Metadata: map[string]interface{}{
			"ocr_text": "OCR text from image",
		},
	}

	content := cb.extractContent(doc)

	// Text content should take priority over image
	assert.Equal(t, "This is text content", content)
	assert.NotContains(t, content, "OCR text from image")
}

// TestBuildContext_WithImageDocuments tests full context building with images
func TestBuildContext_WithImageDocuments(t *testing.T) {
	config := &CustomAgentConfig{
		Response: CustomAgentResponseConfig{
			MaxContextChunks:   10,
			MinRelevanceScore:  0.0,
			DeduplicateSources: false,
			SortByRelevance:    true,
		},
		Output: CustomAgentOutputConfig{
			IncludeMetadata: true,
		},
	}

	cb := NewContextBuilder(config)

	results := []*vectordb.QueryResult{
		{
			Document: vectordb.Document{
				ID:    "img-001",
				Image: "https://cdn.example.com/car1.jpg",
				Metadata: map[string]interface{}{
					"ocr_text":    "1967 Mustang",
					"description": "Red vintage Mustang",
					"_collection": "ProductImages",
					"_vdb":        "weaviate-cloud",
				},
			},
			Score: 0.89,
		},
		{
			Document: vectordb.Document{
				ID:      "text-001",
				Content: "Ford Mustang history and specifications",
				Metadata: map[string]interface{}{
					"_collection": "ProductDocs",
					"_vdb":        "mongodb-cloud",
				},
			},
			Score: 0.85,
		},
		{
			Document: vectordb.Document{
				ID:    "img-002",
				Image: "https://cdn.example.com/car2.jpg",
				Metadata: map[string]interface{}{
					"ocr_text":    "1969 Camaro",
					"description": "Blue vintage Camaro",
					"_collection": "ProductImages",
					"_vdb":        "weaviate-cloud",
				},
			},
			Score: 0.82,
		},
	}

	context, err := cb.BuildContext("vintage cars", results)
	require.NoError(t, err)

	// Should have all 3 sources
	assert.Equal(t, 3, len(context.Sources))

	// First source (highest score) should be image with OCR
	assert.Equal(t, 0.89, context.Sources[0].Score)
	assert.Contains(t, context.Sources[0].Content, "Text in image: 1967 Mustang")
	assert.Contains(t, context.Sources[0].Content, "Description: Red vintage Mustang")

	// Second source should be text
	assert.Equal(t, 0.85, context.Sources[1].Score)
	assert.Equal(t, "Ford Mustang history and specifications", context.Sources[1].Content)

	// Third source should be image
	assert.Equal(t, 0.82, context.Sources[2].Score)
	assert.Contains(t, context.Sources[2].Content, "Text in image: 1969 Camaro")
}

// TestExtractContent_ImageWithEmptyStringMetadata tests handling of empty string metadata
func TestExtractContent_ImageWithEmptyStringMetadata(t *testing.T) {
	config := &CustomAgentConfig{
		Response: CustomAgentResponseConfig{
			MaxContextChunks:   10,
			MinRelevanceScore:  0.0,
			DeduplicateSources: false,
			SortByRelevance:    false,
		},
	}

	cb := NewContextBuilder(config)

	doc := vectordb.Document{
		ID:    "img-007",
		Image: "https://cdn.example.com/empty.jpg",
		Metadata: map[string]interface{}{
			"ocr_text":    "",
			"description": "",
			"alt_text":    "",
			"tags":        []interface{}{},
		},
	}

	content := cb.extractContent(doc)

	// Should fall back to image URL since all metadata is empty
	assert.Equal(t, "Image URL: https://cdn.example.com/empty.jpg", content)
}

// TestExtractContent_ImageWithMixedEmptyMetadata tests partial metadata
func TestExtractContent_ImageWithMixedEmptyMetadata(t *testing.T) {
	config := &CustomAgentConfig{
		Response: CustomAgentResponseConfig{
			MaxContextChunks:   10,
			MinRelevanceScore:  0.0,
			DeduplicateSources: false,
			SortByRelevance:    false,
		},
	}

	cb := NewContextBuilder(config)

	doc := vectordb.Document{
		ID:    "img-008",
		Image: "https://cdn.example.com/partial.jpg",
		Metadata: map[string]interface{}{
			"ocr_text":    "",
			"description": "Valid description",
			"alt_text":    "",
		},
	}

	content := cb.extractContent(doc)

	// Should only include non-empty fields
	assert.Contains(t, content, "Description: Valid description")
	assert.Contains(t, content, "Image URL: https://cdn.example.com/partial.jpg")
	assert.NotContains(t, content, "Text in image:")
	assert.NotContains(t, content, "Alt text:")
}
