// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package mock

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/weaviate"
)

// Client represents a mock vector database client
type Client struct {
	config      *config.MockConfig
	collections map[string][]Document
	mutex       sync.RWMutex
}

// Document represents a mock document
type Document struct {
	ID        string                 `json:"id"`
	Content   string                 `json:"content"`
	Image     string                 `json:"image"`
	ImageData string                 `json:"image_data"`
	URL       string                 `json:"url"`
	Metadata  map[string]interface{} `json:"metadata"`
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

	// Add some test data for demonstration
	client.addTestData()

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

	// Actually delete the collection from the map
	delete(c.collections, collectionName)
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

// CountDocuments returns the number of documents in a collection
func (c *Client) CountDocuments(ctx context.Context, collectionName string) (int, error) {
	_, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	documents, exists := c.collections[collectionName]
	if !exists {
		return 0, fmt.Errorf("collection %s does not exist", collectionName)
	}

	return len(documents), nil
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

// addTestData adds sample documents for demonstration purposes
func (c *Client) addTestData() {
	// Add test documents to WeaveDocs collection
	if _, exists := c.collections["WeaveDocs"]; exists {
		testDocs := []Document{
			{
				ID:      "doc1-chunk1",
				Content: "This is the first chunk of a document about machine learning. It covers the basics of supervised learning algorithms.",
				Metadata: map[string]interface{}{
					"metadata": `{"original_filename": "ml_guide.pdf", "is_chunked": true, "chunk_index": 0, "total_chunks": 3}`,
				},
			},
			{
				ID:      "doc1-chunk2",
				Content: "This is the second chunk discussing neural networks and deep learning architectures including CNNs and RNNs.",
				Metadata: map[string]interface{}{
					"metadata": `{"original_filename": "ml_guide.pdf", "is_chunked": true, "chunk_index": 1, "total_chunks": 3}`,
				},
			},
			{
				ID:      "doc1-chunk3",
				Content: "This is the final chunk covering practical applications, examples, and best practices for machine learning projects.",
				Metadata: map[string]interface{}{
					"metadata": `{"original_filename": "ml_guide.pdf", "is_chunked": true, "chunk_index": 2, "total_chunks": 3}`,
				},
			},
			{
				ID: "doc2-single",
				Content: `This is a single document without chunks. It contains information about data preprocessing techniques and feature engineering.

Data preprocessing is a crucial step in the machine learning pipeline. It involves cleaning, transforming, and organizing raw data into a format that is suitable for analysis and modeling.

Key steps in data preprocessing include:
1. Data cleaning - removing or correcting errors, inconsistencies, and missing values
2. Data transformation - converting data into appropriate formats
3. Feature selection - choosing relevant features for the model
4. Feature scaling - normalizing or standardizing features
5. Data splitting - dividing data into training, validation, and test sets

Proper data preprocessing can significantly improve model performance and reduce overfitting. It's important to understand the domain and the specific requirements of your machine learning task when applying these techniques.`,
				Metadata: map[string]interface{}{
					"author": "Test Author",
					"topic":  "Data Preprocessing",
					"year":   "2024",
				},
			},
			{
				ID:      "doc3-chunk1",
				Content: "First chunk of another document about data science methodologies and statistical analysis techniques.",
				Metadata: map[string]interface{}{
					"metadata": `{"original_filename": "data_science.pdf", "is_chunked": true, "chunk_index": 0, "total_chunks": 2}`,
				},
			},
			{
				ID:      "doc3-chunk2",
				Content: "Second chunk covering data analysis techniques, visualization methods, and reporting best practices.",
				Metadata: map[string]interface{}{
					"metadata": `{"original_filename": "data_science.pdf", "is_chunked": true, "chunk_index": 1, "total_chunks": 2}`,
				},
			},
		}
		c.collections["WeaveDocs"] = testDocs
	}

	// Add test images to WeaveImages collection
	if _, exists := c.collections["WeaveImages"]; exists {
		testImages := []Document{
			{
				ID:      "img1-page1",
				Content: "Image extracted from page 1 of document.pdf",
				Metadata: map[string]interface{}{
					"metadata": `{"original_filename": "document.pdf", "page_number": 1, "image_type": "chart"}`,
				},
			},
			{
				ID:      "img1-page2",
				Content: "Image extracted from page 2 of document.pdf",
				Metadata: map[string]interface{}{
					"metadata": `{"original_filename": "document.pdf", "page_number": 2, "image_type": "diagram"}`,
				},
			},
			{
				ID:      "img2-single",
				Content: "Standalone image from presentation.pptx",
				Metadata: map[string]interface{}{
					"metadata": `{"original_filename": "presentation.pptx", "slide_number": 5, "image_type": "screenshot"}`,
				},
			},
		}
		c.collections["WeaveImages"] = testImages
	}
}

// CreateCollection creates a new collection in the mock database
func (c *Client) CreateCollection(ctx context.Context, collectionName, embeddingModel string, customFields []weaviate.FieldDefinition) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Check if collection already exists
	if _, exists := c.collections[collectionName]; exists {
		return fmt.Errorf("collection '%s' already exists", collectionName)
	}

	// Create empty collection
	c.collections[collectionName] = []Document{}

	return nil
}

// CreateDocument creates a new document in the specified collection
func (c *Client) CreateDocument(ctx context.Context, collectionName string, document Document) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Check if collection exists
	if _, exists := c.collections[collectionName]; !exists {
		return fmt.Errorf("collection '%s' does not exist", collectionName)
	}

	// Check if document ID already exists
	for _, existingDoc := range c.collections[collectionName] {
		if existingDoc.ID == document.ID {
			return fmt.Errorf("document with ID '%s' already exists in collection '%s'", document.ID, collectionName)
		}
	}

	// Add document to collection
	c.collections[collectionName] = append(c.collections[collectionName], document)
	return nil
}

// DeleteAllDocuments deletes all documents in a collection
func (c *Client) DeleteAllDocuments(ctx context.Context, collectionName string) error {
	_, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, exists := c.collections[collectionName]; !exists {
		return fmt.Errorf("collection %s does not exist", collectionName)
	}

	// Clear all documents from the collection
	c.collections[collectionName] = []Document{}
	return nil
}

// Query performs semantic search on a collection (mock implementation)
func (c *Client) Query(ctx context.Context, collectionName, queryText string, options weaviate.QueryOptions) ([]weaviate.QueryResult, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// Check if collection exists
	documents, exists := c.collections[collectionName]
	if !exists {
		return nil, fmt.Errorf("collection '%s' does not exist", collectionName)
	}

	// Simple mock semantic search based on keyword matching
	var results []weaviate.QueryResult
	queryWords := strings.Fields(strings.ToLower(queryText))

	for _, doc := range documents {
		score := c.CalculateMockScore(doc, queryWords)
		if score > 0 {
			results = append(results, weaviate.QueryResult{
				ID:       doc.ID,
				Content:  doc.Content,
				Metadata: doc.Metadata,
				Score:    score,
			})
		}
	}

	// Sort by score (highest first)
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Score < results[j].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// Limit results to top_k
	if options.TopK > 0 && len(results) > options.TopK {
		results = results[:options.TopK]
	}

	return results, nil
}

// CalculateMockScore calculates a mock similarity score based on keyword matching
func (c *Client) CalculateMockScore(doc Document, queryWords []string) float64 {
	if len(queryWords) == 0 {
		return 0.0
	}

	content := strings.ToLower(doc.Content)
	matches := 0.0
	totalWords := float64(len(queryWords))

	for _, word := range queryWords {
		wordLower := strings.ToLower(word)
		if strings.Contains(content, wordLower) {
			matches++
		}
	}

	// Return score as percentage of matched words
	return matches / totalWords
}
