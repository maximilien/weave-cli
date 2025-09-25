package weaviate

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/auth"
)

// Client wraps the Weaviate client with additional functionality
type Client struct {
	client *weaviate.Client
	config *Config
}

// Config holds Weaviate client configuration
type Config struct {
	URL    string
	APIKey string
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
				"X-OpenAI-Api-Key": config.APIKey,
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

	// Delete all objects in the collection using GraphQL
	query := fmt.Sprintf(`
		mutation {
			delete {
				%s(where: {
					path: ["id"]
					operator: Like
					valueString: "*"
				}) {
					successful
					failed
				}
			}
		}
	`, collectionName)

	result, err := c.client.GraphQL().Raw().WithQuery(query).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete collection %s: %w", collectionName, err)
	}

	// Check if deletion was successful
	if data, ok := result.Data["delete"].(map[string]interface{}); ok {
		if collectionData, ok := data[collectionName].(map[string]interface{}); ok {
			if successful, ok := collectionData["successful"].(float64); ok && successful > 0 {
				return nil
			}
		}
	}

	return fmt.Errorf("failed to delete collection %s: no objects deleted", collectionName)
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

// listDocumentsOptimized fetches documents with ID and essential fields for virtual aggregation
// Excludes large base64 image fields for better performance
func (c *Client) listDocumentsOptimized(ctx context.Context, collectionName string, limit int) ([]Document, error) {
	// Build a query that includes ID and specific fields needed for virtual document aggregation
	// Excludes 'image', 'image_data', and 'base64_data' fields to avoid performance issues
	query := fmt.Sprintf(`
		{
			Get {
				%s(limit: %d) {
					_additional {
						id
					}
					url
					metadata
					content_type
					date_added
					file_size
					filename
					source_document
					pdf_filename
					source_type
					processed_by
				}
			}
		}
	`, collectionName, limit)

	result, err := c.client.GraphQL().Raw().WithQuery(query).Do(ctx)
	if err != nil {
		// If the optimized query fails, fall back to simple query
		return c.listDocumentsSimple(ctx, collectionName, limit)
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

					// Extract essential fields for virtual document aggregation
					doc.Metadata = make(map[string]interface{})
					doc.Metadata["id"] = doc.ID

					// Add fields that help with virtual document grouping
					if url, ok := itemMap["url"].(string); ok {
						doc.Metadata["url"] = url
					}
					if filename, ok := itemMap["filename"].(string); ok {
						doc.Metadata["filename"] = filename
					}
					if sourceDocument, ok := itemMap["source_document"].(string); ok {
						doc.Metadata["source_document"] = sourceDocument
					}
					if pdfFilename, ok := itemMap["pdf_filename"].(string); ok {
						doc.Metadata["pdf_filename"] = pdfFilename
					}
					if sourceType, ok := itemMap["source_type"].(string); ok {
						doc.Metadata["source_type"] = sourceType
					}
					if contentType, ok := itemMap["content_type"].(string); ok {
						doc.Metadata["content_type"] = contentType
					}
					if dateAdded, ok := itemMap["date_added"].(string); ok {
						doc.Metadata["date_added"] = dateAdded
					}
					if fileSize, ok := itemMap["file_size"]; ok {
						doc.Metadata["file_size"] = fileSize
					}
					if processedBy, ok := itemMap["processed_by"].(string); ok {
						doc.Metadata["processed_by"] = processedBy
					}

					// Add metadata field if present (may contain classification info)
					if metadata, ok := itemMap["metadata"]; ok {
						doc.Metadata["metadata"] = metadata
					}

					// Add placeholders for excluded large fields to indicate they exist
					doc.Metadata["image"] = "[base64 data excluded for performance]"
					doc.Metadata["image_data"] = "[base64 data excluded for performance]"

					doc.Content = fmt.Sprintf("Document ID: %s", doc.ID)
					documents = append(documents, doc)
				}
			}
		}
	}

	return documents, nil
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
		return c.listDocumentsSimple(ctx, collectionName, limit)
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
		return nil, fmt.Errorf("failed to query documents: %w", err)
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
		query += fmt.Sprintf("\n\t\t\t\t%s", prop)
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
	ID       string                 `json:"id"`
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata"`
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
