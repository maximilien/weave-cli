// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package weaviate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ListCollections returns a list of all collections
func (c *Client) ListCollections(ctx context.Context) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	collections, err := c.client.Schema().Getter().Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("Weaviate: failed to get collections: %w", err)
	}

	var collectionNames []string
	for _, collection := range collections.Classes {
		collectionNames = append(collectionNames, collection.Class)
	}

	return collectionNames, nil
}

// DeleteCollection deletes all objects from a collection
func (c *Client) DeleteCollection(ctx context.Context, collectionName string) error {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	// Use the WeaveClient which has better REST API support
	weaveClient, err := NewWeaveClient(c.config)
	if err != nil {
		return fmt.Errorf("Weaviate: failed to create weave client: %w", err)
	}

	return weaveClient.DeleteCollection(ctx, collectionName)
}

// DeleteCollectionSchema deletes the collection schema completely
func (c *Client) DeleteCollectionSchema(ctx context.Context, collectionName string) error {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	// Use the WeaveClient which has better REST API support
	weaveClient, err := NewWeaveClient(c.config)
	if err != nil {
		return fmt.Errorf("Weaviate: failed to create weave client: %w", err)
	}

	return weaveClient.DeleteCollectionSchema(ctx, collectionName)
}

// CreateCollection creates a new collection with the specified schema
func (c *Client) CreateCollection(ctx context.Context, collectionName, embeddingModel string, customFields []FieldDefinition) error {
	return c.CreateCollectionWithSchema(ctx, collectionName, embeddingModel, customFields, "")
}

// CreateCollectionWithSchema creates a new collection with the specified schema type
func (c *Client) CreateCollectionWithSchema(ctx context.Context, collectionName, embeddingModel string, customFields []FieldDefinition, schemaType string) error {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	// Check if collection already exists
	collections, err := c.ListCollections(ctx)
	if err != nil {
		return fmt.Errorf("Weaviate: failed to check existing collections: %w", err)
	}

	for _, existingCollection := range collections {
		if existingCollection == collectionName {
			return fmt.Errorf("Weaviate: collection '%s' already exists", collectionName)
		}
	}

	// Create the collection using Weaviate's REST API
	err = c.createCollectionViaREST(ctx, collectionName, embeddingModel, customFields, schemaType)
	if err != nil {
		return fmt.Errorf("Weaviate: failed to create collection '%s': %w", collectionName, err)
	}

	return nil
}

// createCollectionViaREST creates a collection using Weaviate's REST API
func (c *Client) createCollectionViaREST(ctx context.Context, collectionName, embeddingModel string, customFields []FieldDefinition, schemaType string) error {
	// Determine if this is an image collection
	var isImage bool
	if schemaType != "" {
		// Use explicit schema type if provided
		isImage = (schemaType == "image")
	} else {
		// Fall back to name-based detection for backward compatibility
		isImage = isImageCollection(collectionName)
	}

	// Build the class schema
	classSchema := map[string]interface{}{
		"class": collectionName,
	}

	// Configure vectorizer using named vectors (required in Weaviate v1.25+)
	if isImage {
		// Image collections: Use dual named vectors for text + visual search
		// 1. text_vector: Vectorizes text metadata (captions, OCR, filenames) via text2vec-openai
		// 2. image_vector: Vectorizes actual image pixels via multi2vec-clip
		classSchema["vectorConfig"] = map[string]interface{}{
			"text_vector": map[string]interface{}{
				"vectorizer": map[string]interface{}{
					"text2vec-openai": map[string]interface{}{
						"model": embeddingModel,
					},
				},
				"vectorIndexType": "hnsw",
			},
			"image_vector": map[string]interface{}{
				"vectorizer": map[string]interface{}{
					"multi2vec-clip": map[string]interface{}{
						"imageFields": []string{"image_data"}, // Vectorize the actual image
						"textFields":  []string{"url"},        // Include filename for context
						"weights": map[string]interface{}{
							"imageFields": []float64{0.9}, // Prioritize visual content
							"textFields":  []float64{0.1}, // Minor boost from filename
						},
					},
				},
				"vectorIndexType": "hnsw",
			},
		}
	} else {
		// Text collections: Single named vector for text search
		// Check if vectorizer is disabled
		if embeddingModel == "none" {
			classSchema["vectorizer"] = "none"
		} else {
			classSchema["vectorConfig"] = map[string]interface{}{
				"default": map[string]interface{}{
					"vectorizer": map[string]interface{}{
						"text2vec-openai": map[string]interface{}{
							"model": embeddingModel,
						},
					},
					"vectorIndexType": "hnsw",
				},
			}
		}
	}

	// Create schema based on type
	if isImage {
		// RagMeImages schema: url, image, metadata, image_data
		classSchema["properties"] = []map[string]interface{}{
			{
				"name":     "url",
				"dataType": []string{"text"},
				"moduleConfig": map[string]interface{}{
					"text2vec-openai": map[string]interface{}{
						"skip": false, // Vectorize URL for text search
					},
					"multi2vec-clip": map[string]interface{}{
						"skip": false, // Include filename in CLIP vector
					},
				},
			},
			{
				"name":     "image",
				"dataType": []string{"text"},
				"moduleConfig": map[string]interface{}{
					"text2vec-openai": map[string]interface{}{
						"skip": true, // Skip binary image data
					},
				},
			},
			{
				"name":     "image_data",
				"dataType": []string{"text"}, // Base64 encoded image data
				"moduleConfig": map[string]interface{}{
					"text2vec-openai": map[string]interface{}{
						"skip": true, // Skip for text vectorization
					},
					"multi2vec-clip": map[string]interface{}{
						"skip": false, // Use for visual vectorization
					},
				},
			},
			{
				"name":     "metadata",
				"dataType": []string{"object"},
				"nestedProperties": []map[string]interface{}{
					{
						"name":     "filename",
						"dataType": []string{"text"},
						"moduleConfig": map[string]interface{}{
							"text2vec-openai": map[string]interface{}{
								"skip": false, // Vectorize filename for searchability
							},
						},
					},
					{
						"name":     "file_size",
						"dataType": []string{"number"},
					},
					{
						"name":     "content_type",
						"dataType": []string{"text"},
						"moduleConfig": map[string]interface{}{
							"text2vec-openai": map[string]interface{}{
								"skip": true, // Skip MIME types
							},
						},
					},
					{
						"name":     "date_added",
						"dataType": []string{"text"},
						"moduleConfig": map[string]interface{}{
							"text2vec-openai": map[string]interface{}{
								"skip": true, // Skip timestamps
							},
						},
					},
					{
						"name":     "image_index",
						"dataType": []string{"int"},
					},
					{
						"name":     "source_document",
						"dataType": []string{"text"},
						"moduleConfig": map[string]interface{}{
							"text2vec-openai": map[string]interface{}{
								"skip": false, // Vectorize source document name
							},
						},
					},
					{
						"name":     "is_extracted_from_document",
						"dataType": []string{"boolean"},
					},
					{
						"name":     "ocr_content",
						"dataType": []string{"text"},
						"moduleConfig": map[string]interface{}{
							"text2vec-openai": map[string]interface{}{
								"skip": false, // Vectorize OCR text for semantic search
							},
						},
					},
					{
						"name":     "has_ocr_text",
						"dataType": []string{"boolean"},
					},
					{
						"name":     "caption",
						"dataType": []string{"text"},
						"moduleConfig": map[string]interface{}{
							"text2vec-openai": map[string]interface{}{
								"skip": false, // Vectorize captions for semantic search
							},
						},
					},
					{
						"name":     "surrounding_text",
						"dataType": []string{"text"},
						"moduleConfig": map[string]interface{}{
							"text2vec-openai": map[string]interface{}{
								"skip": false, // Vectorize surrounding text for semantic search
							},
						},
					},
					{
						"name":     "section_heading",
						"dataType": []string{"text"},
						"moduleConfig": map[string]interface{}{
							"text2vec-openai": map[string]interface{}{
								"skip": false, // Vectorize section headings for semantic search
							},
						},
					},
					{
						"name":     "pdf_page",
						"dataType": []string{"int"},
					},
				},
			},
		}
	} else {
		// RagMeDocs schema: url, text, metadata
		classSchema["properties"] = []map[string]interface{}{
			{
				"name":     "url",
				"dataType": []string{"text"},
			},
			{
				"name":     "text",
				"dataType": []string{"text"},
			},
			{
				"name":     "metadata",
				"dataType": []string{"text"}, // Store as JSON string for compatibility
			},
		}
	}

	// Add custom fields if provided (skip fields that already exist in default schema)
	existingProps := make(map[string]bool)
	for _, prop := range classSchema["properties"].([]map[string]interface{}) {
		existingProps[prop["name"].(string)] = true
	}

	for _, field := range customFields {
		// Skip if property already exists in the default schema
		if existingProps[field.Name] {
			continue
		}
		property := map[string]interface{}{
			"name":     field.Name,
			"dataType": []string{mapWeaviateDataType(field.Type)},
		}
		classSchema["properties"] = append(classSchema["properties"].([]map[string]interface{}), property)
	}

	// Convert to JSON
	jsonData, err := json.Marshal(classSchema)
	if err != nil {
		return fmt.Errorf("Weaviate: failed to marshal class schema: %w", err)
	}

	// Create HTTP request
	url := c.config.URL + "/v1/schema"
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("Weaviate: failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	if c.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}

	// Make the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Weaviate: failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Weaviate: failed to read response: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Weaviate: failed to create collection: HTTP %d - %s", resp.StatusCode, string(body))
	}

	return nil
}
