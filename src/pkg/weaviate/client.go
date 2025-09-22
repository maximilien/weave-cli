package weaviate

import (
	"context"
	"fmt"
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

	if config.APIKey != "" {
		// Use API key authentication for Weaviate Cloud
		client, err = weaviate.NewClient(weaviate.Config{
			Host:   config.URL,
			Scheme: "https",
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
			Host:   config.URL,
			Scheme: "http",
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
	// Use the basic method that works reliably
	return c.listDocumentsBasic(ctx, collectionName, limit)
}

// listDocumentsBasic fetches documents with actual properties
func (c *Client) listDocumentsBasic(ctx context.Context, collectionName string, limit int) ([]Document, error) {
	// First, get the schema to know what fields are available
	properties, err := c.GetCollectionSchema(ctx, collectionName)
	if err != nil {
		// If we can't get schema, fall back to a simple ID-only query
		return c.listDocumentsSimple(ctx, collectionName, limit)
	}

	// Build a query with the actual properties from the schema
	query := fmt.Sprintf(`
		{
			Get {
				%s(limit: %d) {
					_additional {
						id
					}
	`, collectionName, limit)

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

	// Delete the specific object using GraphQL
	query := fmt.Sprintf(`
		mutation {
			delete {
				%s(where: {
					path: ["_additional", "id"]
					operator: Equal
					valueString: "%s"
				}) {
					successful
					failed
				}
			}
		}
	`, collectionName, documentID)

	result, err := c.client.GraphQL().Raw().WithQuery(query).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete document %s from collection %s: %w", documentID, collectionName, err)
	}

	// Check if deletion was successful
	if data, ok := result.Data["delete"].(map[string]interface{}); ok {
		if collectionData, ok := data[collectionName].(map[string]interface{}); ok {
			if successful, ok := collectionData["successful"].(float64); ok && successful > 0 {
				return nil
			}
		}
	}

	return fmt.Errorf("failed to delete document %s from collection %s: document not found", documentID, collectionName)
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
