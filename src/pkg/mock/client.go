package mock

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/config"
)

// Client represents a mock vector database client
type Client struct {
	config      *config.MockConfig
	collections map[string][]Document
	mutex       sync.RWMutex
}

// Document represents a mock document
type Document struct {
	ID       string                 `json:"id"`
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata"`
}

// NewClient creates a new mock client
func NewClient(config *config.MockConfig) *Client {
	client := &Client{
		config:      config,
		collections: make(map[string][]Document),
	}

	// Initialize collections from config
	for _, collection := range config.Collections {
		client.collections[collection.Name] = []Document{}
	}

	return client
}

// Health checks the health of the mock database
func (c *Client) Health(ctx context.Context) error {
	_, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	// Simulate a quick health check
	time.Sleep(100 * time.Millisecond)
	return nil
}

// ListCollections returns a list of all collections
func (c *Client) ListCollections(ctx context.Context) ([]string, error) {
	_, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	var collections []string
	for name := range c.collections {
		collections = append(collections, name)
	}

	return collections, nil
}

// DeleteCollection deletes all objects from a collection
func (c *Client) DeleteCollection(ctx context.Context, collectionName string) error {
	_, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, exists := c.collections[collectionName]; !exists {
		return fmt.Errorf("collection %s does not exist", collectionName)
	}

	c.collections[collectionName] = []Document{}
	return nil
}

// ListDocuments returns a list of documents in a collection
func (c *Client) ListDocuments(ctx context.Context, collectionName string, limit int) ([]Document, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	documents, exists := c.collections[collectionName]
	if !exists {
		return nil, fmt.Errorf("collection %s does not exist", collectionName)
	}

	// Apply limit
	if limit > 0 && limit < len(documents) {
		documents = documents[:limit]
	}

	return documents, nil
}

// GetDocument retrieves a specific document by ID
func (c *Client) GetDocument(ctx context.Context, collectionName, documentID string) (*Document, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	documents, exists := c.collections[collectionName]
	if !exists {
		return nil, fmt.Errorf("collection %s does not exist", collectionName)
	}

	for _, doc := range documents {
		if doc.ID == documentID {
			return &doc, nil
		}
	}

	return nil, fmt.Errorf("document with ID %s not found in collection %s", documentID, collectionName)
}

// DeleteDocument deletes a specific document by ID
func (c *Client) DeleteDocument(ctx context.Context, collectionName, documentID string) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	c.mutex.Lock()
	defer c.mutex.Unlock()

	documents, exists := c.collections[collectionName]
	if !exists {
		return fmt.Errorf("collection %s does not exist", collectionName)
	}

	// Find and remove the document
	for i, doc := range documents {
		if doc.ID == documentID {
			c.collections[collectionName] = append(documents[:i], documents[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("document with ID %s not found in collection %s", documentID, collectionName)
}

// AddDocument adds a document to a collection (for testing purposes)
func (c *Client) AddDocument(ctx context.Context, collectionName string, doc Document) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, exists := c.collections[collectionName]; !exists {
		return fmt.Errorf("collection %s does not exist", collectionName)
	}

	c.collections[collectionName] = append(c.collections[collectionName], doc)
	return nil
}

// GetCollectionStats returns statistics about a collection
func (c *Client) GetCollectionStats(ctx context.Context, collectionName string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	documents, exists := c.collections[collectionName]
	if !exists {
		return nil, fmt.Errorf("collection %s does not exist", collectionName)
	}

	stats := map[string]interface{}{
		"name":                collectionName,
		"document_count":      len(documents),
		"embedding_dimension": c.config.EmbeddingDimension,
		"simulate_embeddings": c.config.SimulateEmbeddings,
	}

	return stats, nil
}
