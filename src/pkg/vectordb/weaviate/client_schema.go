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

// GetCollectionSchema returns the schema for a collection
func (c *Client) GetCollectionSchema(ctx context.Context, collectionName string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Get the schema using the REST API
	schema, err := c.client.Schema().Getter().Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema: %w", err)
	}

	var properties []string
	for _, class := range schema.Classes {
		if class.Class == collectionName {
			for _, prop := range class.Properties {
				properties = append(properties, prop.Name)
			}
			break
		}
	}

	return properties, nil
}

// GetFullCollectionSchema returns the full schema for a collection
func (c *Client) GetFullCollectionSchema(ctx context.Context, collectionName string) (*CollectionSchema, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Get the schema using the REST API
	schema, err := c.client.Schema().Getter().Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema: %w", err)
	}

	for _, class := range schema.Classes {
		if class.Class == collectionName {
			// Convert to our schema format
			result := &CollectionSchema{
				Class:      class.Class,
				Vectorizer: class.Vectorizer,
				Properties: make([]SchemaProperty, len(class.Properties)),
			}

			for i, prop := range class.Properties {
				result.Properties[i] = SchemaProperty{
					Name:        prop.Name,
					DataType:    prop.DataType,
					Description: prop.Description,
				}

				// Convert nested properties if available
				if len(prop.NestedProperties) > 0 {
					result.Properties[i].NestedProperties = make([]SchemaProperty, len(prop.NestedProperties))
					for j, nested := range prop.NestedProperties {
						result.Properties[i].NestedProperties[j] = SchemaProperty{
							Name:     nested.Name,
							DataType: nested.DataType,
						}
					}
				}
			}

			return result, nil
		}
	}

	return nil, fmt.Errorf("collection '%s' not found in schema", collectionName)
}

// CreateCollectionFromSchema creates a collection from a CollectionSchema object
func (c *Client) CreateCollectionFromSchema(ctx context.Context, schema *CollectionSchema) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Check if collection already exists
	collections, err := c.ListCollections(ctx)
	if err != nil {
		return fmt.Errorf("failed to check existing collections: %w", err)
	}

	for _, existingCollection := range collections {
		if existingCollection == schema.Class {
			return fmt.Errorf("collection '%s' already exists", schema.Class)
		}
	}

	// Build the schema payload
	classSchema := map[string]interface{}{
		"class": schema.Class,
	}

	// Add vectorizer if specified
	if schema.Vectorizer != "" {
		classSchema["vectorizer"] = schema.Vectorizer
	}

	// Convert properties to the format expected by Weaviate API
	if len(schema.Properties) > 0 {
		properties := make([]map[string]interface{}, len(schema.Properties))
		for i, prop := range schema.Properties {
			property := map[string]interface{}{
				"name":     prop.Name,
				"dataType": prop.DataType,
			}

			if prop.Description != "" {
				property["description"] = prop.Description
			}

			// Handle nested properties
			if len(prop.NestedProperties) > 0 {
				nestedProps := make([]map[string]interface{}, len(prop.NestedProperties))
				for j, nested := range prop.NestedProperties {
					nestedProps[j] = map[string]interface{}{
						"name":     nested.Name,
						"dataType": nested.DataType,
					}
					if nested.Description != "" {
						nestedProps[j]["description"] = nested.Description
					}
				}
				property["nestedProperties"] = nestedProps
			}

			properties[i] = property
		}
		classSchema["properties"] = properties
	}

	// Marshal to JSON
	schemaJSON, err := json.Marshal(classSchema)
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	// Parse URL to extract host and scheme
	host := c.config.URL
	scheme := "http"
	if strings.HasPrefix(host, "http://") {
		host = strings.TrimPrefix(host, "http://")
		scheme = "http"
	} else if strings.HasPrefix(host, "https://") {
		host = strings.TrimPrefix(host, "https://")
		scheme = "https"
	}

	// Send request to Weaviate
	url := fmt.Sprintf("%s://%s/v1/schema", scheme, host)
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(schemaJSON)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.config.APIKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	}

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create collection: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}
