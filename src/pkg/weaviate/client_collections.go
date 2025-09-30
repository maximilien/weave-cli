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
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	collections, err := c.client.Schema().Getter().Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get collections: %w", err)
	}

	var collectionNames []string
	for _, collection := range collections.Classes {
		collectionNames = append(collectionNames, collection.Class)
	}

	return collectionNames, nil
}

// DeleteCollection deletes all objects from a collection
func (c *Client) DeleteCollection(ctx context.Context, collectionName string) error {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// Use the WeaveClient which has better REST API support
	weaveClient, err := NewWeaveClient(c.config)
	if err != nil {
		return fmt.Errorf("failed to create weave client: %w", err)
	}

	return weaveClient.DeleteCollection(ctx, collectionName)
}

// DeleteCollectionSchema deletes the collection schema completely
func (c *Client) DeleteCollectionSchema(ctx context.Context, collectionName string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Use the WeaveClient which has better REST API support
	weaveClient, err := NewWeaveClient(c.config)
	if err != nil {
		return fmt.Errorf("failed to create weave client: %w", err)
	}

	return weaveClient.DeleteCollectionSchema(ctx, collectionName)
}

// CreateCollection creates a new collection with the specified schema
func (c *Client) CreateCollection(ctx context.Context, collectionName, embeddingModel string, customFields []FieldDefinition) error {
	return c.CreateCollectionWithSchema(ctx, collectionName, embeddingModel, customFields, "")
}

// CreateCollectionWithSchema creates a new collection with the specified schema type
func (c *Client) CreateCollectionWithSchema(ctx context.Context, collectionName, embeddingModel string, customFields []FieldDefinition, schemaType string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Check if collection already exists
	collections, err := c.ListCollections(ctx)
	if err != nil {
		return fmt.Errorf("failed to check existing collections: %w", err)
	}

	for _, existingCollection := range collections {
		if existingCollection == collectionName {
			return fmt.Errorf("collection '%s' already exists", collectionName)
		}
	}

	// Create the collection using Weaviate's REST API
	err = c.createCollectionViaREST(ctx, collectionName, embeddingModel, customFields, schemaType)
	if err != nil {
		return fmt.Errorf("failed to create collection '%s': %w", collectionName, err)
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

	// Configure vectorizer based on collection type
	if isImage {
		// For image collections, disable vectorization to avoid issues with large base64 data
		classSchema["vectorizer"] = "none"
	} else {
		// For text collections, use text2vec-openai
		classSchema["vectorizer"] = "text2vec-openai"
		classSchema["moduleConfig"] = map[string]interface{}{
			"text2vec-openai": map[string]interface{}{
				"model": embeddingModel,
			},
		}
	}

	// Create schema based on type
	if isImage {
		// RagMeImages schema: url, image, metadata, image_data
		classSchema["properties"] = []map[string]interface{}{
			{
				"name":     "url",
				"dataType": []string{"text"},
			},
			{
				"name":     "image",
				"dataType": []string{"text"},
			},
			{
				"name":     "image_data",
				"dataType": []string{"text"},
			},
			{
				"name":     "metadata",
				"dataType": []string{"object"},
				"nestedProperties": []map[string]interface{}{
					{
						"name":     "filename",
						"dataType": []string{"text"},
					},
					{
						"name":     "file_size",
						"dataType": []string{"number"},
					},
					{
						"name":     "content_type",
						"dataType": []string{"text"},
					},
					{
						"name":     "date_added",
						"dataType": []string{"text"},
					},
					{
						"name":     "image_index",
						"dataType": []string{"int"},
					},
					{
						"name":     "source_document",
						"dataType": []string{"text"},
					},
					{
						"name":     "is_extracted_from_document",
						"dataType": []string{"boolean"},
					},
					{
						"name":     "ocr_content",
						"dataType": []string{"text"},
					},
					{
						"name":     "has_ocr_text",
						"dataType": []string{"boolean"},
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

	// Add custom fields if provided
	for _, field := range customFields {
		property := map[string]interface{}{
			"name":     field.Name,
			"dataType": []string{mapWeaviateDataType(field.Type)},
		}
		classSchema["properties"] = append(classSchema["properties"].([]map[string]interface{}), property)
	}

	// Convert to JSON
	jsonData, err := json.Marshal(classSchema)
	if err != nil {
		return fmt.Errorf("failed to marshal class schema: %w", err)
	}

	// Create HTTP request
	url := c.config.URL + "/v1/schema"
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
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
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create collection: HTTP %d - %s", resp.StatusCode, string(body))
	}

	return nil
}
