// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package qdrant

import (
	"context"
	"fmt"
	"os"

	"github.com/maximilien/weave-cli/src/pkg/llm"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// Adapter wraps the Qdrant client to implement the vectordb.VectorDBClient interface
type Adapter struct {
	*Client
	llmClient *llm.OpenAIClient
}

// NewAdapter creates a new Qdrant adapter from the vectordb.Config
func NewAdapter(config *vectordb.Config) (*Adapter, error) {
	// Parse address or URL to extract host and port
	host, port := parseAddress(config)

	qdrantConfig := &Config{
		Host:             host,
		Port:             port,
		APIKey:           config.APIKey,
		UseTLS:           config.Type == vectordb.VectorDBTypeQdrantCloud,
		VectorDimensions: config.VectorDimensions,
		SimilarityMetric: config.SimilarityMetric,
		Timeout:          config.Timeout,
	}

	client, err := NewClient(qdrantConfig)
	if err != nil {
		return nil, err
	}

	// Create LLM client for embeddings (optional, only if OpenAI API key is available)
	var llmClient *llm.OpenAIClient
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey != "" {
		var err error
		llmClient, err = llm.NewOpenAIClient(openaiKey)
		if err != nil {
			// Just log warning, don't fail - embeddings won't work but other operations will
			fmt.Fprintf(os.Stderr, "Warning: Failed to create OpenAI client for embeddings: %v\n", err)
		}
	}

	return &Adapter{
		Client:    client,
		llmClient: llmClient,
	}, nil
}

// CreateCollection creates a new collection with the given schema
func (a *Adapter) CreateCollection(ctx context.Context, name string, schema *vectordb.CollectionSchema) error {
	// Use config defaults or schema settings
	vectorDims := a.config.VectorDimensions
	similarityMetric := a.config.SimilarityMetric

	return a.Client.CreateCollection(ctx, name, vectorDims, similarityMetric)
}

// ListCollections returns a list of all collections
func (a *Adapter) ListCollections(ctx context.Context) ([]vectordb.CollectionInfo, error) {
	names, err := a.Client.ListCollections(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to CollectionInfo
	collections := make([]vectordb.CollectionInfo, len(names))
	for i, name := range names {
		count, err := a.Client.GetCollectionCount(ctx, name)
		if err != nil {
			count = 0 // If we can't get count, just use 0
		}
		collections[i] = vectordb.CollectionInfo{
			Name:  name,
			Count: count,
		}
	}

	return collections, nil
}

// CreateDocument creates a new document in the specified collection
func (a *Adapter) CreateDocument(ctx context.Context, collectionName string, document *vectordb.Document) error {
	// Convert vectordb.Document to Qdrant Document
	doc := &Document{
		ID:       document.ID,
		Content:  document.Text, // Use Text field as content
		Metadata: document.Metadata,
	}

	// If no vector is provided and we have an LLM client, generate embedding
	if doc.Vector == nil && a.llmClient != nil && doc.Content != "" {
		embedding, err := a.llmClient.GenerateEmbedding(ctx, doc.Content, "text-embedding-3-small")
		if err != nil {
			return fmt.Errorf("failed to generate embedding: %w", err)
		}
		// Convert []float64 to []float32
		doc.Vector = make([]float32, len(embedding))
		for i, v := range embedding {
			doc.Vector[i] = float32(v)
		}
	}

	return a.Client.CreateDocument(ctx, collectionName, doc)
}

// CreateDocuments creates multiple documents in batch
func (a *Adapter) CreateDocuments(ctx context.Context, collectionName string, documents []*vectordb.Document) error {
	// Convert vectordb.Documents to Qdrant Documents
	docs := make([]*Document, len(documents))
	for i, document := range documents {
		doc := &Document{
			ID:       document.ID,
			Content:  document.Text,
			Metadata: document.Metadata,
		}

		// If no vector is provided and we have an LLM client, generate embedding
		if doc.Vector == nil && a.llmClient != nil && doc.Content != "" {
			embedding, err := a.llmClient.GenerateEmbedding(ctx, doc.Content, "text-embedding-3-small")
			if err != nil {
				return fmt.Errorf("failed to generate embedding for document %d: %w", i, err)
			}
			// Convert []float64 to []float32
			doc.Vector = make([]float32, len(embedding))
			for j, v := range embedding {
				doc.Vector[j] = float32(v)
			}
		}

		docs[i] = doc
	}

	return a.Client.CreateDocuments(ctx, collectionName, docs)
}

// GetDocument retrieves a document by ID
func (a *Adapter) GetDocument(ctx context.Context, collectionName, documentID string) (*vectordb.Document, error) {
	doc, err := a.Client.GetDocument(ctx, collectionName, documentID)
	if err != nil {
		return nil, err
	}

	// Convert Qdrant Document to vectordb.Document
	return &vectordb.Document{
		ID:       doc.ID,
		Text:     doc.Content,
		Metadata: doc.Metadata,
	}, nil
}

// UpdateDocument updates an existing document
func (a *Adapter) UpdateDocument(ctx context.Context, collectionName string, document *vectordb.Document) error {
	// Convert vectordb.Document to Qdrant Document
	doc := &Document{
		ID:       document.ID,
		Content:  document.Text,
		Metadata: document.Metadata,
	}

	// If no vector is provided and we have an LLM client, generate embedding
	if doc.Vector == nil && a.llmClient != nil && doc.Content != "" {
		embedding, err := a.llmClient.GenerateEmbedding(ctx, doc.Content, "text-embedding-3-small")
		if err != nil {
			return fmt.Errorf("failed to generate embedding: %w", err)
		}
		// Convert []float64 to []float32
		doc.Vector = make([]float32, len(embedding))
		for i, v := range embedding {
			doc.Vector[i] = float32(v)
		}
	}

	return a.Client.UpdateDocument(ctx, collectionName, doc)
}

// DeleteDocumentsByMetadata deletes documents matching metadata criteria
func (a *Adapter) DeleteDocumentsByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}) error {
	// Build filter from metadata
	filter, err := BuildFilter(metadata)
	if err != nil {
		return fmt.Errorf("failed to build filter: %w", err)
	}

	// Search for matching documents
	docs, err := a.Client.SearchByMetadata(ctx, collectionName, filter, 1000)
	if err != nil {
		return fmt.Errorf("failed to search for documents: %w", err)
	}

	// Extract IDs and delete
	if len(docs) > 0 {
		ids := make([]string, len(docs))
		for i, doc := range docs {
			ids[i] = doc.ID
		}
		return a.Client.DeleteDocuments(ctx, collectionName, ids)
	}

	return nil
}

// ListDocuments returns a list of documents in a collection
func (a *Adapter) ListDocuments(ctx context.Context, collectionName string, limit int, offset int) ([]*vectordb.Document, error) {
	// Qdrant doesn't have native offset support in Scroll API
	// For now, we'll just use limit and ignore offset
	docs, err := a.Client.ListDocuments(ctx, collectionName, limit)
	if err != nil {
		return nil, err
	}

	// Convert Qdrant Documents to vectordb.Documents
	result := make([]*vectordb.Document, len(docs))
	for i, doc := range docs {
		result[i] = &vectordb.Document{
			ID:       doc.ID,
			Text:     doc.Content,
			Metadata: doc.Metadata,
		}
	}

	return result, nil
}

// SearchSemantic performs semantic search using vector embeddings
func (a *Adapter) SearchSemantic(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	if a.llmClient == nil {
		return nil, fmt.Errorf("OpenAI client not configured - cannot generate embeddings")
	}

	// Generate embedding for query
	embedding64, err := a.llmClient.GenerateEmbedding(ctx, query, "text-embedding-3-small")
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Convert []float64 to []float32
	embedding := make([]float32, len(embedding64))
	for i, v := range embedding64 {
		embedding[i] = float32(v)
	}

	// Get TopK from options
	topK := 10
	if options != nil && options.TopK > 0 {
		topK = options.TopK
	}

	// Search
	results, err := a.Client.SearchSemantic(ctx, collectionName, embedding, topK)
	if err != nil {
		return nil, err
	}

	// Convert to QueryResults
	queryResults := make([]*vectordb.QueryResult, len(results))
	for i, result := range results {
		queryResults[i] = &vectordb.QueryResult{
			Document: vectordb.Document{
				ID:       result.Document.ID,
				Text:     result.Document.Content,
				Metadata: result.Document.Metadata,
			},
			Score: float64(result.Score),
		}
	}

	return queryResults, nil
}

// SearchBM25 performs keyword-based search using BM25
func (a *Adapter) SearchBM25(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	// Qdrant doesn't support BM25 out of the box
	return nil, fmt.Errorf("Qdrant does not support BM25 search")
}

// SearchHybrid performs hybrid search combining vector and keyword search
func (a *Adapter) SearchHybrid(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	// For now, just use semantic search
	// In the future, we could support filtering via metadata
	return a.SearchSemantic(ctx, collectionName, query, options)
}

// SearchByMetadata searches documents by metadata fields
func (a *Adapter) SearchByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	// Build filter from metadata
	filter, err := BuildFilter(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to build filter: %w", err)
	}

	// Get limit from options
	limit := 100
	if options != nil && options.TopK > 0 {
		limit = options.TopK
	}

	// Search
	docs, err := a.Client.SearchByMetadata(ctx, collectionName, filter, limit)
	if err != nil {
		return nil, err
	}

	// Convert to QueryResults
	queryResults := make([]*vectordb.QueryResult, len(docs))
	for i, doc := range docs {
		queryResults[i] = &vectordb.QueryResult{
			Document: vectordb.Document{
				ID:       doc.ID,
				Text:     doc.Content,
				Metadata: doc.Metadata,
			},
			Score: 1.0, // Metadata search doesn't have a score
		}
	}

	return queryResults, nil
}

// GetSchema retrieves the schema for a collection
func (a *Adapter) GetSchema(ctx context.Context, collectionName string) (*vectordb.CollectionSchema, error) {
	// Qdrant doesn't have a schema concept like Weaviate
	// Return a minimal schema
	return &vectordb.CollectionSchema{
		Class: collectionName,
	}, nil
}

// GetDefaultSchema returns a default schema for the given type
func (a *Adapter) GetDefaultSchema(schemaType vectordb.SchemaType, collectionName string) *vectordb.CollectionSchema {
	vectorizer := "text-embedding-3-small"
	return &vectordb.CollectionSchema{
		Class:      collectionName,
		Vectorizer: vectorizer,
	}
}

// ValidateSchema validates a schema definition
func (a *Adapter) ValidateSchema(schema *vectordb.CollectionSchema) error {
	// Qdrant doesn't require complex schema validation
	return nil
}

// UpdateSchema updates a collection's schema
func (a *Adapter) UpdateSchema(ctx context.Context, collectionName string, schema *vectordb.CollectionSchema) error {
	// Qdrant doesn't support schema updates after collection creation
	return fmt.Errorf("Qdrant does not support schema updates after collection creation")
}

// parseAddress extracts host and port from config address or URL
func parseAddress(config *vectordb.Config) (string, int) {
	host := "localhost"
	port := 6334

	// Check environment variables first
	if envHost := os.Getenv("QDRANT_HOST"); envHost != "" {
		host = envHost
	}
	if envPort := os.Getenv("QDRANT_GRPC_PORT"); envPort != "" {
		_, _ = fmt.Sscanf(envPort, "%d", &port)
	}

	// Config takes precedence over environment
	if config.Address != "" {
		// Parse address (e.g., "localhost:6334" or just "localhost")
		parts := splitHostPort(config.Address)
		if len(parts) > 0 {
			host = parts[0]
		}
		if len(parts) > 1 {
			_, _ = fmt.Sscanf(parts[1], "%d", &port)
		}
	} else if config.URL != "" {
		// Parse URL to extract host and port
		parts := splitHostPort(config.URL)
		if len(parts) > 0 {
			host = parts[0]
		}
		if len(parts) > 1 {
			_, _ = fmt.Sscanf(parts[1], "%d", &port)
		}
	}

	return host, port
}

// splitHostPort splits a host:port string, handling cases with or without port
func splitHostPort(addr string) []string {
	// Remove any protocol prefix (https://, http://)
	if idx := findProtocol(addr); idx >= 0 {
		addr = addr[idx:]
	}

	// Split on last colon to handle IPv6 addresses
	lastColon := -1
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			lastColon = i
			break
		}
	}

	if lastColon >= 0 {
		return []string{addr[:lastColon], addr[lastColon+1:]}
	}
	return []string{addr}
}

// findProtocol returns the index after the protocol, or -1 if not found
func findProtocol(addr string) int {
	if len(addr) > 8 && addr[:8] == "https://" {
		return 8
	}
	if len(addr) > 7 && addr[:7] == "http://" {
		return 7
	}
	return -1
}
