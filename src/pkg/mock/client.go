package mock

import (
	"context"
	"fmt"
	"strings"
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
	_, cancel := context.WithTimeout(ctx, 1*time.Second)
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
	_, cancel := context.WithTimeout(ctx, 1*time.Second)
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
	_, cancel := context.WithTimeout(ctx, 1*time.Second)
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

// DeleteDocumentsByMetadata deletes documents matching metadata filters
func (c *Client) DeleteDocumentsByMetadata(ctx context.Context, collectionName string, metadataFilters []string) (int, error) {
	_, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	c.mutex.Lock()
	defer c.mutex.Unlock()

	documents, exists := c.collections[collectionName]
	if !exists {
		return 0, fmt.Errorf("collection %s does not exist", collectionName)
	}

	// Parse metadata filters
	filters := make(map[string]string)
	for _, filter := range metadataFilters {
		parts := strings.SplitN(filter, "=", 2)
		if len(parts) != 2 {
			return 0, fmt.Errorf("invalid metadata filter format: %s (expected key=value)", filter)
		}
		filters[parts[0]] = parts[1]
	}

	// Find documents matching the filters
	var matchingIndices []int
	for i, doc := range documents {
		matches := true
		for key, value := range filters {
			if doc.Metadata == nil {
				matches = false
				break
			}
			if docValue, exists := doc.Metadata[key]; !exists || fmt.Sprintf("%v", docValue) != value {
				matches = false
				break
			}
		}
		if matches {
			matchingIndices = append(matchingIndices, i)
		}
	}

	// Delete matching documents (in reverse order to maintain indices)
	deletedCount := 0
	for i := len(matchingIndices) - 1; i >= 0; i-- {
		idx := matchingIndices[i]
		c.collections[collectionName] = append(documents[:idx], documents[idx+1:]...)
		documents = c.collections[collectionName] // Update slice reference
		deletedCount++
	}

	return deletedCount, nil
}

// AddDocument adds a document to a collection (for testing purposes)
func (c *Client) AddDocument(ctx context.Context, collectionName string, doc Document) error {
	_, cancel := context.WithTimeout(ctx, 1*time.Second)
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
	_, cancel := context.WithTimeout(ctx, 1*time.Second)
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

// GetDocumentsByMetadata gets documents matching metadata filters
func (c *Client) GetDocumentsByMetadata(ctx context.Context, collectionName string, metadataFilters []string) ([]Document, error) {
	_, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	c.mutex.Lock()
	defer c.mutex.Unlock()

	documents, exists := c.collections[collectionName]
	if !exists {
		return nil, fmt.Errorf("collection %s does not exist", collectionName)
	}

	// Parse metadata filters
	filters := make(map[string]string)
	for _, filter := range metadataFilters {
		parts := strings.SplitN(filter, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid metadata filter format: %s (expected key=value)", filter)
		}
		filters[parts[0]] = parts[1]
	}

	// Find documents matching the filters
	var matchingDocuments []Document
	for _, doc := range documents {
		matches := true
		for key, value := range filters {
			if doc.Metadata == nil {
				matches = false
				break
			}

			// Check if the metadata contains the key-value pair
			if doc.Metadata[key] != value {
				matches = false
				break
			}
		}

		if matches {
			matchingDocuments = append(matchingDocuments, doc)
		}
	}

	return matchingDocuments, nil
}
