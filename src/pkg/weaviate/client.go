package weaviate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/auth"
)

// FieldDefinition represents a field in a collection
type FieldDefinition struct {
	Name string
	Type string
}

// Client wraps the Weaviate client with additional functionality
type Client struct {
	client *weaviate.Client
	config *Config
}

// Config holds Weaviate client configuration
type Config struct {
	URL          string
	APIKey       string
	OpenAIAPIKey string
}

// NewClient creates a new Weaviate client
func NewClient(config *Config) (*Client, error) {
	var client *weaviate.Client
	var err error

	// Parse URL to extract host and scheme
	host := config.URL
	scheme := "http"

	// Remove protocol if present
	if strings.HasPrefix(host, "http://") {
		host = strings.TrimPrefix(host, "http://")
		scheme = "http"
	} else if strings.HasPrefix(host, "https://") {
		host = strings.TrimPrefix(host, "https://")
		scheme = "https"
	}

	if config.APIKey != "" {
		// Use API key authentication for Weaviate Cloud
		client, err = weaviate.NewClient(weaviate.Config{
			Host:   host,
			Scheme: scheme,
			AuthConfig: auth.ApiKey{
				Value: config.APIKey,
			},
			Headers: map[string]string{
				"X-Openai-Api-Key": config.OpenAIAPIKey,
			},
		})
	} else {
		// Use no authentication for local Weaviate
		client, err = weaviate.NewClient(weaviate.Config{
			Host:   host,
			Scheme: scheme,
		})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create Weaviate client: %w", err)
	}

	return &Client{
		client: client,
		config: config,
	}, nil
}

// Health checks the health of the Weaviate instance
func (c *Client) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Try to get the meta information
	meta, err := c.client.Misc().MetaGetter().Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Weaviate meta: %w", err)
	}

	if meta == nil {
		return fmt.Errorf("received nil meta from Weaviate")
	}

	return nil
}

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
	err = c.createCollectionViaREST(ctx, collectionName, embeddingModel, customFields)
	if err != nil {
		return fmt.Errorf("failed to create collection '%s': %w", collectionName, err)
	}

	return nil
}

// createCollectionViaREST creates a collection using Weaviate's REST API
func (c *Client) createCollectionViaREST(ctx context.Context, collectionName, embeddingModel string, customFields []FieldDefinition) error {
	// Build the class schema
	classSchema := map[string]interface{}{
		"class":      collectionName,
		"vectorizer": "text2vec-openai",
		"moduleConfig": map[string]interface{}{
			"text2vec-openai": map[string]interface{}{
				"model": embeddingModel,
			},
		},
		"properties": []map[string]interface{}{
			{
				"name":     "content",
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
						"name":     "chunk_index",
						"dataType": []string{"int"},
					},
					{
						"name":     "chunk_size",
						"dataType": []string{"int"},
					},
					{
						"name":     "total_chunks",
						"dataType": []string{"int"},
					},
					{
						"name":     "source_document",
						"dataType": []string{"text"},
					},
					{
						"name":     "processed_by",
						"dataType": []string{"text"},
					},
					{
						"name":     "processing_time",
						"dataType": []string{"int"},
					},
					{
						"name":     "is_extracted_from_document",
						"dataType": []string{"boolean"},
					},
					{
						"name":     "file_extension",
						"dataType": []string{"text"},
					},
				},
			},
		},
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

// mapWeaviateDataType maps our field types to Weaviate data types
func mapWeaviateDataType(fieldType string) string {
	switch fieldType {
	case "text":
		return "text"
	case "int":
		return "int"
	case "float":
		return "number"
	case "bool":
		return "boolean"
	case "date":
		return "date"
	case "object":
		return "object"
	default:
		return "text" // Default to text
	}
}

// ListDocuments returns a list of documents in a collection
// Note: Currently shows document IDs only. To show actual document content/metadata,
// we would need to implement dynamic schema discovery for each collection.
func (c *Client) ListDocuments(ctx context.Context, collectionName string, limit int) ([]Document, error) {
	// For image collections, use basic method with excluded fields for performance
	if isImageCollection(collectionName) {
		// Use basic method which dynamically discovers schema and excludes large fields
		return c.listDocumentsBasic(ctx, collectionName, limit)
	}

	// Use the basic method that works reliably for text collections
	return c.listDocumentsBasic(ctx, collectionName, limit)
}

// CountDocuments efficiently counts documents in a collection without fetching content
// This is much faster than ListDocuments for large collections with heavy data
func (c *Client) CountDocuments(ctx context.Context, collectionName string) (int, error) {
	// Use Weaviate's aggregation API to count documents efficiently
	// This doesn't fetch the actual document content, just counts them
	query := fmt.Sprintf(`
		query {
			Aggregate {
				%s {
					meta {
						count
					}
				}
			}
		}
	`, collectionName)

	result, err := c.client.GraphQL().Raw().WithQuery(query).Do(ctx)
	if err != nil {
		// Check for common connection errors and provide better messages
		if strings.Contains(err.Error(), "connection reset") || strings.Contains(err.Error(), "status code: -1") {
			return 0, fmt.Errorf("collection %s not found, check database configuration", collectionName)
		}
		return 0, fmt.Errorf("failed to count documents in collection %s: %w", collectionName, err)
	}

	// Extract count from the result
	if len(result.Errors) > 0 {
		// Parse GraphQL errors to provide user-friendly messages
		for _, err := range result.Errors {
			if err.Message != "" {
				// Check for common error patterns and provide better messages
				if strings.Contains(err.Message, "class") && strings.Contains(err.Message, "not found") {
					return 0, fmt.Errorf("collection %s does not exist", collectionName)
				}
				if strings.Contains(err.Message, "Unknown class") {
					return 0, fmt.Errorf("collection %s does not exist", collectionName)
				}
				// Check for "Did you mean" suggestions and extract them
				if strings.Contains(err.Message, "Did you mean") {
					// Extract the suggestion part from the error message
					parts := strings.Split(err.Message, "Did you mean")
					if len(parts) > 1 {
						suggestion := strings.TrimSpace(parts[1])
						// Remove trailing question mark and clean up
						suggestion = strings.TrimSuffix(suggestion, "?")
						return 0, fmt.Errorf("collection %s does not exist. Did you mean %s?", collectionName, suggestion)
					}
				}
				return 0, fmt.Errorf("graphql error: %s", err.Message)
			}
		}
		return 0, fmt.Errorf("graphql errors: %v", result.Errors)
	}

	// Parse the aggregation result
	if result.Data != nil {
		if aggregateData, ok := result.Data["Aggregate"].(map[string]interface{}); ok {
			if collectionData, ok := aggregateData[collectionName].([]interface{}); ok && len(collectionData) > 0 {
				if metaData, ok := collectionData[0].(map[string]interface{}); ok {
					if countData, ok := metaData["meta"].(map[string]interface{}); ok {
						if count, ok := countData["count"].(float64); ok {
							return int(count), nil
						}
					}
				}
			}
		}
	}

	return 0, fmt.Errorf("failed to parse count result for collection %s", collectionName)
}

// isImageCollection checks if a collection name suggests it contains images
func isImageCollection(collectionName string) bool {
	imageKeywords := []string{"image", "img", "photo", "picture", "visual"}
	name := strings.ToLower(collectionName)
	for _, keyword := range imageKeywords {
		if strings.Contains(name, keyword) {
			return true
		}
	}
	return false
}

// listDocumentsBasic fetches documents with actual properties (excluding large fields)
func (c *Client) listDocumentsBasic(ctx context.Context, collectionName string, limit int) ([]Document, error) {
	// First, get the schema to know what fields are available
	properties, err := c.GetCollectionSchema(ctx, collectionName)
	if err != nil {
		// If we can't get schema, fall back to a simple ID-only query
		return c.listDocumentsSimple(ctx, collectionName, limit)
	}

	// Filter out large fields that cause performance issues
	excludedFields := map[string]bool{
		"image":       true, // Base64 image data can be very large
		"image_data":  true, // Another image field name
		"base64_data": true, // Alternative image field name
		"content":     true, // Large text content
	}

	// Build a query with the actual properties from the schema (excluding large fields)
	query := fmt.Sprintf(`
		{
			Get {
				%s(limit: %d) {
					_additional {
						id
					}
	`, collectionName, limit)

	// Add available properties to the query, excluding large fields
	for _, prop := range properties {
		if !excludedFields[prop] {
			if prop == "metadata" {
				// Dynamically discover metadata schema and build appropriate query
				metadataQuery, err := c.buildMetadataQuery(ctx, collectionName)
				if err != nil {
					// If we can't discover the schema, use simple field
					query += fmt.Sprintf("\n\t\t\t\t%s", prop)
				} else {
					query += metadataQuery
				}
			} else {
				query += fmt.Sprintf("\n\t\t\t\t%s", prop)
			}
		}
	}

	query += `
				}
			}
		}
	`

	result, err := c.client.GraphQL().Raw().WithQuery(query).Do(ctx)
	if err != nil {
		// Check for metadata field type mismatch error
		if strings.Contains(err.Error(), "must not have a sub selection") && strings.Contains(err.Error(), "metadata") {
			// Retry with simple metadata field (for old collections with string metadata)
			return c.listDocumentsWithSimpleMetadata(ctx, collectionName, limit, properties, excludedFields)
		}
		// Check for common connection errors and provide better messages
		if strings.Contains(err.Error(), "connection reset") || strings.Contains(err.Error(), "status code: -1") {
			return nil, fmt.Errorf("collection %s not found, check database configuration", collectionName)
		}
		// If the schema-based query fails, fall back to simple query
		return c.listDocumentsSimple(ctx, collectionName, limit)
	}

	// Check for GraphQL errors
	if len(result.Errors) > 0 {
		// Parse GraphQL errors to provide user-friendly messages
		for _, err := range result.Errors {
			if err.Message != "" {
				// Check for common error patterns and provide better messages
				if strings.Contains(err.Message, "class") && strings.Contains(err.Message, "not found") {
					return nil, fmt.Errorf("collection %s does not exist", collectionName)
				}
				if strings.Contains(err.Message, "Unknown class") {
					return nil, fmt.Errorf("collection %s does not exist", collectionName)
				}
				// Check for "Did you mean" suggestions and extract them
				if strings.Contains(err.Message, "Did you mean") {
					// Extract the suggestion part from the error message
					parts := strings.Split(err.Message, "Did you mean")
					if len(parts) > 1 {
						suggestion := strings.TrimSpace(parts[1])
						// Remove trailing question mark and clean up
						suggestion = strings.TrimSuffix(suggestion, "?")
						return nil, fmt.Errorf("collection %s does not exist. Did you mean %s?", collectionName, suggestion)
					}
				}
				return nil, fmt.Errorf("graphql error: %s", err.Message)
			}
		}
		return nil, fmt.Errorf("graphql errors: %v", result.Errors)
	}

	var documents []Document
	if data, ok := result.Data["Get"].(map[string]interface{}); ok {
		if collectionData, ok := data[collectionName].([]interface{}); ok {
			for _, item := range collectionData {
				if itemMap, ok := item.(map[string]interface{}); ok {
					doc := Document{}

					// Extract ID
					if additional, ok := itemMap["_additional"].(map[string]interface{}); ok {
						if id, ok := additional["id"].(string); ok {
							doc.ID = id
						}
					}

					// Extract all properties as metadata
					doc.Metadata = make(map[string]interface{})
					doc.Metadata["id"] = doc.ID

					// Extract content from common field names
					contentFields := []string{"text", "content", "body", "description", "title", "name", "chunk", "pageContent", "document"}
					doc.Content = ""

					for key, value := range itemMap {
						if key != "_additional" {
							doc.Metadata[key] = value

							// Try to find content in common field names
							for _, field := range contentFields {
								if key == field {
									if str, ok := value.(string); ok && str != "" {
										doc.Content = str
										break
									}
								}
							}
						}
					}

					// Add placeholders for excluded large fields to indicate they exist
					if isImageCollection(collectionName) {
						doc.Metadata["image"] = "[base64 data excluded for performance]"
						doc.Metadata["base64_data"] = "[base64 data excluded for performance]"
					}
					if doc.Metadata["content"] == nil {
						doc.Metadata["content"] = "[large content excluded for performance]"
					}

					// If no content found, create a summary
					if doc.Content == "" {
						doc.Content = fmt.Sprintf("Document ID: %s", doc.ID)
					}

					documents = append(documents, doc)
				}
			}
		}
	}

	return documents, nil
}

// buildMetadataQuery dynamically discovers the metadata schema and builds the appropriate GraphQL query
func (c *Client) buildMetadataQuery(ctx context.Context, collectionName string) (string, error) {
	// Get the collection schema via REST API to understand the metadata structure
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/v1/schema/%s", c.config.URL, collectionName), nil)
	if err != nil {
		return "\n\t\t\t\tmetadata", nil
	}

	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	if c.config.OpenAIAPIKey != "" {
		req.Header.Set("X-Openai-Api-Key", c.config.OpenAIAPIKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "\n\t\t\t\tmetadata", nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "\n\t\t\t\tmetadata", nil
	}

	var schema struct {
		Properties []struct {
			Name             string   `json:"name"`
			DataType         []string `json:"dataType"`
			NestedProperties []struct {
				Name string `json:"name"`
			} `json:"nestedProperties"`
		} `json:"properties"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&schema); err != nil {
		return "\n\t\t\t\tmetadata", nil
	}

	// Find the metadata property
	for _, prop := range schema.Properties {
		if prop.Name == "metadata" {
			// Check if it's an object type
			if len(prop.DataType) > 0 && prop.DataType[0] == "object" {
				// Build query with available nested properties
				if len(prop.NestedProperties) > 0 {
					var nestedFields []string
					for _, nested := range prop.NestedProperties {
						nestedFields = append(nestedFields, nested.Name)
					}

					query := "\n\t\t\t\tmetadata {\n"
					for _, field := range nestedFields {
						query += fmt.Sprintf("\t\t\t\t\t%s\n", field)
					}
					query += "\t\t\t\t}"
					return query, nil
				}
			}
			// If it's not an object or has no nested properties, use simple metadata
			return "\n\t\t\t\tmetadata", nil
		}
	}

	// If metadata property not found, use simple metadata
	return "\n\t\t\t\tmetadata", nil
}

// listDocumentsWithSimpleMetadata handles collections with string metadata (old format)
func (c *Client) listDocumentsWithSimpleMetadata(ctx context.Context, collectionName string, limit int, properties []string, excludedFields map[string]bool) ([]Document, error) {
	// Build a query with simple metadata field (no sub-selection)
	query := fmt.Sprintf(`
		{
			Get {
				%s(limit: %d) {
					_additional {
						id
					}
	`, collectionName, limit)

	// Add available properties to the query, excluding large fields
	for _, prop := range properties {
		if !excludedFields[prop] {
			query += fmt.Sprintf("\n\t\t\t\t%s", prop)
		}
	}

	query += `
				}
			}
		}
	`

	result, err := c.client.GraphQL().Raw().WithQuery(query).Do(ctx)
	if err != nil {
		// If this also fails, fall back to simple query
		return c.listDocumentsSimple(ctx, collectionName, limit)
	}

	// Check for GraphQL errors
	if len(result.Errors) > 0 {
		// Parse GraphQL errors to provide user-friendly messages
		for _, err := range result.Errors {
			if err.Message != "" {
				// Check for common error patterns and provide better messages
				if strings.Contains(err.Message, "class") && strings.Contains(err.Message, "not found") {
					return nil, fmt.Errorf("collection %s does not exist", collectionName)
				}
			}
		}
		return nil, fmt.Errorf("graphql error: %s", result.Errors[0].Message)
	}

	var documents []Document
	if data, ok := result.Data["Get"].(map[string]interface{}); ok {
		if collectionData, ok := data[collectionName].([]interface{}); ok {
			for _, item := range collectionData {
				if itemMap, ok := item.(map[string]interface{}); ok {
					doc := Document{}

					// Extract ID
					if additional, ok := itemMap["_additional"].(map[string]interface{}); ok {
						if id, ok := additional["id"].(string); ok {
							doc.ID = id
						}
					}

					// Extract properties
					for _, prop := range properties {
						if !excludedFields[prop] {
							if value, exists := itemMap[prop]; exists {
								if doc.Metadata == nil {
									doc.Metadata = make(map[string]interface{})
								}
								doc.Metadata[prop] = value
							}
						}
					}

					documents = append(documents, doc)
				}
			}
		}
	}

	return documents, nil
}

// listDocumentsSimple is a fallback method that only gets IDs
func (c *Client) listDocumentsSimple(ctx context.Context, collectionName string, limit int) ([]Document, error) {
	query := fmt.Sprintf(`
		{
			Get {
				%s(limit: %d) {
					_additional {
						id
					}
				}
			}
		}
	`, collectionName, limit)

	result, err := c.client.GraphQL().Raw().WithQuery(query).Do(ctx)
	if err != nil {
		// Check for common connection errors and provide better messages
		if strings.Contains(err.Error(), "connection reset") || strings.Contains(err.Error(), "status code: -1") {
			return nil, fmt.Errorf("collection %s not found, check database configuration", collectionName)
		}
		return nil, fmt.Errorf("failed to query documents: %w", err)
	}

	// Check for GraphQL errors
	if len(result.Errors) > 0 {
		// Parse GraphQL errors to provide user-friendly messages
		for _, err := range result.Errors {
			if err.Message != "" {
				// Check for common error patterns and provide better messages
				if strings.Contains(err.Message, "class") && strings.Contains(err.Message, "not found") {
					return nil, fmt.Errorf("collection %s does not exist", collectionName)
				}
				if strings.Contains(err.Message, "Unknown class") {
					return nil, fmt.Errorf("collection %s does not exist", collectionName)
				}
				// Check for "Did you mean" suggestions and extract them
				if strings.Contains(err.Message, "Did you mean") {
					// Extract the suggestion part from the error message
					parts := strings.Split(err.Message, "Did you mean")
					if len(parts) > 1 {
						suggestion := strings.TrimSpace(parts[1])
						// Remove trailing question mark and clean up
						suggestion = strings.TrimSuffix(suggestion, "?")
						return nil, fmt.Errorf("collection %s does not exist. Did you mean %s?", collectionName, suggestion)
					}
				}
				return nil, fmt.Errorf("graphql error: %s", err.Message)
			}
		}
		return nil, fmt.Errorf("graphql errors: %v", result.Errors)
	}

	var documents []Document
	if data, ok := result.Data["Get"].(map[string]interface{}); ok {
		if collectionData, ok := data[collectionName].([]interface{}); ok {
			for _, item := range collectionData {
				if itemMap, ok := item.(map[string]interface{}); ok {
					doc := Document{}

					// Extract ID
					if additional, ok := itemMap["_additional"].(map[string]interface{}); ok {
						if id, ok := additional["id"].(string); ok {
							doc.ID = id
						}
					}

					// For counting purposes, we just need the ID
					doc.Metadata = make(map[string]interface{})
					doc.Metadata["id"] = doc.ID
					doc.Content = fmt.Sprintf("Document ID: %s", doc.ID)

					documents = append(documents, doc)
				}
			}
		}
	}

	return documents, nil
}

// GetDocument retrieves a specific document by ID
func (c *Client) GetDocument(ctx context.Context, collectionName, documentID string) (*Document, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// First, get the schema to know what fields are available
	properties, err := c.GetCollectionSchema(ctx, collectionName)
	if err != nil {
		// If we can't get schema, fall back to a simple ID-only query
		return c.getDocumentSimple(ctx, collectionName, documentID)
	}

	// Build a query with the actual properties from the schema
	query := fmt.Sprintf(`
		{
			Get {
				%s(where: {
					path: ["id"]
					operator: Equal
					valueString: "%s"
				}) {
					_additional {
						id
					}
	`, collectionName, documentID)

	// Add all available properties to the query
	for _, prop := range properties {
		if prop == "metadata" {
			// Dynamically discover metadata schema and build appropriate query
			metadataQuery, err := c.buildMetadataQuery(ctx, collectionName)
			if err != nil {
				// If we can't discover the schema, use simple field
				query += fmt.Sprintf("\n\t\t\t\t%s", prop)
			} else {
				query += metadataQuery
			}
		} else {
			query += fmt.Sprintf("\n\t\t\t\t%s", prop)
		}
	}

	query += `
				}
			}
		}
	`

	result, err := c.client.GraphQL().Raw().WithQuery(query).Do(ctx)
	if err != nil {
		// If the schema-based query fails, fall back to simple query
		return c.getDocumentSimple(ctx, collectionName, documentID)
	}

	var document *Document
	if data, ok := result.Data["Get"].(map[string]interface{}); ok {
		if collectionData, ok := data[collectionName].([]interface{}); ok {
			if len(collectionData) > 0 {
				if itemMap, ok := collectionData[0].(map[string]interface{}); ok {
					doc := Document{}

					// Extract ID
					if additional, ok := itemMap["_additional"].(map[string]interface{}); ok {
						if id, ok := additional["id"].(string); ok {
							doc.ID = id
						}
					}

					// Extract all properties as metadata
					doc.Metadata = make(map[string]interface{})
					doc.Metadata["id"] = doc.ID

					// Extract content from common field names
					contentFields := []string{"text", "content", "body", "description", "title", "name", "chunk", "pageContent", "document"}
					doc.Content = ""

					for key, value := range itemMap {
						if key != "_additional" {
							doc.Metadata[key] = value

							// Try to find content in common field names
							for _, field := range contentFields {
								if key == field {
									if str, ok := value.(string); ok && str != "" {
										doc.Content = str
										break
									}
								}
							}
						}
					}

					// If no content found, create a summary
					if doc.Content == "" {
						doc.Content = fmt.Sprintf("Document ID: %s", doc.ID)
					}

					document = &doc
				}
			}
		}
	}

	if document == nil {
		return nil, fmt.Errorf("document with ID %s not found in collection %s", documentID, collectionName)
	}

	return document, nil
}

// getDocumentSimple is a fallback method that only gets IDs
func (c *Client) getDocumentSimple(ctx context.Context, collectionName, documentID string) (*Document, error) {
	query := fmt.Sprintf(`
		{
			Get {
				%s(where: {
					path: ["id"]
					operator: Equal
					valueString: "%s"
				}) {
					_additional {
						id
					}
				}
			}
		}
	`, collectionName, documentID)

	result, err := c.client.GraphQL().Raw().WithQuery(query).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query document: %w", err)
	}

	var document *Document
	if data, ok := result.Data["Get"].(map[string]interface{}); ok {
		if collectionData, ok := data[collectionName].([]interface{}); ok {
			if len(collectionData) > 0 {
				if itemMap, ok := collectionData[0].(map[string]interface{}); ok {
					doc := Document{}

					// Extract ID
					if additional, ok := itemMap["_additional"].(map[string]interface{}); ok {
						if id, ok := additional["id"].(string); ok {
							doc.ID = id
						}
					}

					// For fallback, we just have the ID
					doc.Metadata = make(map[string]interface{})
					doc.Metadata["id"] = doc.ID
					doc.Content = fmt.Sprintf("Document ID: %s", doc.ID)

					document = &doc
				}
			}
		}
	}

	if document == nil {
		return nil, fmt.Errorf("document with ID %s not found in collection %s", documentID, collectionName)
	}

	return document, nil
}

// DeleteDocument deletes a specific document by ID
func (c *Client) DeleteDocument(ctx context.Context, collectionName, documentID string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Use REST API directly since the client's delete method has issues
	// Ensure URL doesn't have trailing slash
	baseURL := strings.TrimSuffix(c.config.URL, "/")
	url := fmt.Sprintf("%s/v1/objects/%s/%s", baseURL, collectionName, documentID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	// Add authorization header
	if c.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete document %s from collection %s: %w", documentID, collectionName, err)
	}
	defer resp.Body.Close()

	// Read response body to check for errors
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check response status
	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK {
		return nil
	}

	// If we get a 404, it means the document was not found
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("failed to delete document %s from collection %s: document not found", documentID, collectionName)
	}

	return fmt.Errorf("failed to delete document %s from collection %s: HTTP %d - %s", documentID, collectionName, resp.StatusCode, string(body))
}

// DeleteDocumentsByMetadata deletes documents matching metadata filters using REST API
func (c *Client) DeleteDocumentsByMetadata(ctx context.Context, collectionName string, metadataFilters []string) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// Parse metadata filters
	filters := make(map[string]string)
	for _, filter := range metadataFilters {
		parts := strings.SplitN(filter, "=", 2)
		if len(parts) != 2 {
			return 0, fmt.Errorf("invalid metadata filter format: %s (expected key=value)", filter)
		}
		filters[parts[0]] = parts[1]
	}

	// First, query for documents matching the metadata filters
	documents, err := c.queryDocumentsByMetadata(ctx, collectionName, filters)
	if err != nil {
		return 0, fmt.Errorf("failed to query documents by metadata: %w", err)
	}

	if len(documents) == 0 {
		return 0, nil // No documents found matching the filters
	}

	// Delete each document individually using REST API
	deletedCount := 0
	for _, doc := range documents {
		if err := c.DeleteDocument(ctx, collectionName, doc.ID); err != nil {
			// Log error but continue with other documents
			fmt.Printf("Warning: Failed to delete document %s: %v\n", doc.ID, err)
			continue
		}
		deletedCount++
	}

	return deletedCount, nil
}

// queryDocumentsByMetadata queries for documents matching metadata filters using GraphQL
func (c *Client) queryDocumentsByMetadata(ctx context.Context, collectionName string, filters map[string]string) ([]Document, error) {
	// Build the where clause for metadata filtering
	var whereClauses []string
	for key, value := range filters {
		if key == "filename" {
			// For filename, we need to search within the JSON string in the metadata field
			// Use Like operator to search for the filename within the JSON string
			whereClauses = append(whereClauses, fmt.Sprintf(`{
				path: ["metadata"]
				operator: Like
				valueString: "*filename\": \"%s\"*"
			}`, value))
		} else if key == "original_filename" {
			// For original_filename, we need to search within the JSON string in the metadata field
			// Use Like operator to search for the original_filename within the JSON string
			whereClauses = append(whereClauses, fmt.Sprintf(`{
				path: ["metadata"]
				operator: Like
				valueString: "*original_filename\": \"%s\"*"
			}`, value))
		} else if key == "url" {
			// For URL, use Like operator to allow partial matching
			whereClauses = append(whereClauses, fmt.Sprintf(`{
				path: ["%s"]
				operator: Like
				valueString: "*%s*"
			}`, key, value))
		} else {
			// For other fields, use direct path with exact matching
			whereClauses = append(whereClauses, fmt.Sprintf(`{
				path: ["%s"]
				operator: Equal
				valueString: "%s"
			}`, key, value))
		}
	}

	// Combine multiple filters with AND
	var whereClause string
	if len(whereClauses) == 1 {
		whereClause = whereClauses[0]
	} else {
		whereClause = fmt.Sprintf(`{
			operator: And
			operands: [%s]
		}`, strings.Join(whereClauses, ", "))
	}

	// Create GraphQL query to get documents
	query := fmt.Sprintf(`
		query {
			Get {
				%s(where: %s) {
					_additional {
						id
					}
				}
			}
		}
	`, collectionName, whereClause)

	result, err := c.client.GraphQL().Raw().WithQuery(query).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query documents by metadata: %w", err)
	}

	// Extract documents
	var documents []Document
	if data, ok := result.Data["Get"].(map[string]interface{}); ok {
		if collectionData, ok := data[collectionName].([]interface{}); ok {
			for _, item := range collectionData {
				if docMap, ok := item.(map[string]interface{}); ok {
					doc := Document{}

					// Get the document ID from _additional
					if additional, ok := docMap["_additional"].(map[string]interface{}); ok {
						if id, ok := additional["id"].(string); ok {
							doc.ID = id
						}
					}

					if doc.ID != "" {
						documents = append(documents, doc)
					}
				}
			}
		}
	}

	return documents, nil
}

// GetDocumentsByMetadata gets documents matching metadata filters
func (c *Client) GetDocumentsByMetadata(ctx context.Context, collectionName string, metadataFilters []string) ([]Document, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// Parse metadata filters
	filters := make(map[string]string)
	for _, filter := range metadataFilters {
		parts := strings.SplitN(filter, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid metadata filter format: %s (expected key=value)", filter)
		}
		filters[parts[0]] = parts[1]
	}

	// Query for documents matching the metadata filters
	documents, err := c.queryDocumentsByMetadata(ctx, collectionName, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to query documents by metadata: %w", err)
	}

	// Get full document details for each document
	var fullDocuments []Document
	for _, doc := range documents {
		fullDoc, err := c.GetDocument(ctx, collectionName, doc.ID)
		if err != nil {
			// Log error but continue with other documents
			fmt.Printf("Warning: Failed to get document %s: %v\n", doc.ID, err)
			continue
		}
		fullDocuments = append(fullDocuments, *fullDoc)
	}

	return fullDocuments, nil
}

// Document represents a document in Weaviate
type Document struct {
	ID        string                 `json:"id"`
	Content   string                 `json:"content"`
	Image     string                 `json:"image"`
	ImageData string                 `json:"image_data"`
	URL       string                 `json:"url"`
	Metadata  map[string]interface{} `json:"metadata"`
}

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
