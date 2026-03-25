// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package redis

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/maximilien/weave-cli/src/pkg/llm"
	"github.com/maximilien/weave-cli/src/pkg/reembedding/providers"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// Adapter wraps the Redis Client to implement vectordb.VectorDBClient.
type Adapter struct {
	*Client
	llmClient *llm.OpenAIClient
}

// NewAdapter creates a new Redis adapter from vectordb.Config.
func NewAdapter(config *vectordb.Config) (*Adapter, error) {
	addr := parseRedisAddress(config)

	password := config.Password
	if password == "" {
		password = config.APIKey // Redis Cloud uses API key as password
	}

	redisCfg := &Config{
		Addr:             addr,
		Password:         password,
		Username:         config.Username,
		UseTLS:           config.Type == vectordb.VectorDBTypeRedisCloud,
		DB:               0,
		VectorDimensions: config.VectorDimensions,
		SimilarityMetric: config.SimilarityMetric,
		Timeout:          config.Timeout,
	}

	client, err := NewClient(redisCfg)
	if err != nil {
		return nil, err
	}

	// Create LLM client for embeddings (optional)
	var llmClient *llm.OpenAIClient
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey != "" {
		llmClient, err = llm.NewOpenAIClient(openaiKey)
		if err != nil {
			fmt.Fprintf(os.Stderr,
				"Warning: Failed to create OpenAI client for embeddings: %v\n", err)
		}
	}

	return &Adapter{Client: client, llmClient: llmClient}, nil
}

// --- Collection operations ---

// CreateCollection creates a new collection.
func (a *Adapter) CreateCollection(ctx context.Context, name string, schema *vectordb.CollectionSchema) error {
	return a.Client.CreateCollection(ctx, name, a.config.VectorDimensions, a.config.SimilarityMetric)
}

// DeleteCollection deletes a collection.
func (a *Adapter) DeleteCollection(ctx context.Context, name string) error {
	return a.Client.DeleteCollection(ctx, name)
}

// ListCollections returns all collections.
func (a *Adapter) ListCollections(ctx context.Context) ([]vectordb.CollectionInfo, error) {
	names, err := a.Client.ListCollections(ctx)
	if err != nil {
		return nil, err
	}

	collections := make([]vectordb.CollectionInfo, len(names))
	for i, name := range names {
		count, _ := a.Client.GetCollectionCount(ctx, name)
		collections[i] = vectordb.CollectionInfo{Name: name, Count: count}
	}
	return collections, nil
}

// CollectionExists checks if a collection exists.
func (a *Adapter) CollectionExists(ctx context.Context, name string) (bool, error) {
	return a.Client.CollectionExists(ctx, name)
}

// GetCollectionCount returns the document count for a collection.
func (a *Adapter) GetCollectionCount(ctx context.Context, name string) (int64, error) {
	return a.Client.GetCollectionCount(ctx, name)
}

// --- Document operations ---

// CreateDocument creates a single document, auto-generating embeddings if needed.
func (a *Adapter) CreateDocument(ctx context.Context, collectionName string, document *vectordb.Document) error {
	if err := a.ensureEmbedding(ctx, document); err != nil {
		return err
	}
	return a.Client.CreateDocument(ctx, collectionName, document)
}

// CreateDocuments creates multiple documents in batch.
func (a *Adapter) CreateDocuments(ctx context.Context, collectionName string, documents []*vectordb.Document) error {
	for _, doc := range documents {
		if err := a.ensureEmbedding(ctx, doc); err != nil {
			return err
		}
	}
	return a.Client.CreateDocuments(ctx, collectionName, documents)
}

// GetDocument retrieves a document by ID.
func (a *Adapter) GetDocument(ctx context.Context, collectionName, documentID string) (*vectordb.Document, error) {
	return a.Client.GetDocument(ctx, collectionName, documentID)
}

// UpdateDocument updates an existing document.
func (a *Adapter) UpdateDocument(ctx context.Context, collectionName string, document *vectordb.Document) error {
	if err := a.ensureEmbedding(ctx, document); err != nil {
		return err
	}
	return a.Client.UpdateDocument(ctx, collectionName, document)
}

// DeleteDocument deletes a single document by ID.
func (a *Adapter) DeleteDocument(ctx context.Context, collectionName, documentID string) error {
	return a.Client.DeleteDocument(ctx, collectionName, documentID)
}

// DeleteDocuments deletes multiple documents by IDs.
func (a *Adapter) DeleteDocuments(ctx context.Context, collectionName string, documentIDs []string) error {
	return a.Client.DeleteDocuments(ctx, collectionName, documentIDs)
}

// DeleteDocumentsByMetadata deletes documents matching metadata criteria.
func (a *Adapter) DeleteDocumentsByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}) error {
	results, err := a.Client.SearchByMetadata(ctx, collectionName, metadata, 1000)
	if err != nil {
		return fmt.Errorf("failed to search for documents: %w", err)
	}
	if len(results) == 0 {
		return nil
	}
	ids := make([]string, len(results))
	for i, r := range results {
		ids[i] = r.Document.ID
	}
	return a.Client.DeleteDocuments(ctx, collectionName, ids)
}

// ListDocuments returns documents with pagination.
func (a *Adapter) ListDocuments(ctx context.Context, collectionName string, limit, offset int) ([]*vectordb.Document, error) {
	return a.Client.ListDocuments(ctx, collectionName, limit, offset)
}

// --- Search operations ---

// SearchSemantic performs vector similarity search.
func (a *Adapter) SearchSemantic(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	embedding, err := a.generateQueryEmbedding(ctx, query)
	if err != nil {
		return nil, err
	}

	topK := 10
	if options != nil && options.TopK > 0 {
		topK = options.TopK
	}

	results, err := a.Client.SearchSemantic(ctx, collectionName, embedding, topK)
	if err != nil {
		return nil, err
	}

	return toQueryResults(results), nil
}

// SearchBM25 performs keyword-based search using Redis full-text search.
func (a *Adapter) SearchBM25(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	topK := 10
	if options != nil && options.TopK > 0 {
		topK = options.TopK
	}

	results, err := a.Client.SearchFullText(ctx, collectionName, query, topK)
	if err != nil {
		return nil, err
	}

	return toQueryResults(results), nil
}

// SearchHybrid performs hybrid search (vector + text filter).
func (a *Adapter) SearchHybrid(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	// Delegate to semantic search (Redis KNN + filter can be combined,
	// but for now semantic is the primary path)
	return a.SearchSemantic(ctx, collectionName, query, options)
}

// SearchByMetadata searches documents by metadata fields.
func (a *Adapter) SearchByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	limit := 100
	if options != nil && options.TopK > 0 {
		limit = options.TopK
	}

	results, err := a.Client.SearchByMetadata(ctx, collectionName, metadata, limit)
	if err != nil {
		return nil, err
	}

	return toQueryResults(results), nil
}

// --- Schema operations ---

// GetSchema returns the schema for a collection.
func (a *Adapter) GetSchema(ctx context.Context, collectionName string) (*vectordb.CollectionSchema, error) {
	vectorizer := inferEmbeddingModelFromDimensions(a.config.VectorDimensions)
	return &vectordb.CollectionSchema{
		Class:      collectionName,
		Vectorizer: vectorizer,
	}, nil
}

// UpdateSchema is not supported — Redis indexes are immutable.
func (a *Adapter) UpdateSchema(ctx context.Context, collectionName string, schema *vectordb.CollectionSchema) error {
	return fmt.Errorf("Redis does not support schema updates after index creation")
}

// GetDefaultSchema returns a default schema.
func (a *Adapter) GetDefaultSchema(schemaType vectordb.SchemaType, collectionName string) *vectordb.CollectionSchema {
	return &vectordb.CollectionSchema{
		Class:      collectionName,
		Vectorizer: "text-embedding-3-small",
	}
}

// ValidateSchema validates a schema definition.
func (a *Adapter) ValidateSchema(schema *vectordb.CollectionSchema) error {
	return nil
}

// --- internal helpers ---

// ensureEmbedding generates an embedding if the document doesn't have one.
func (a *Adapter) ensureEmbedding(ctx context.Context, doc *vectordb.Document) error {
	if doc.Embedding != nil || doc.Text == "" {
		return nil
	}
	if a.llmClient == nil {
		return nil // No embedding client — skip silently
	}

	embedding, err := a.llmClient.GenerateEmbedding(ctx, doc.Text, "text-embedding-3-small")
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}
	doc.Embedding = embedding
	return nil
}

// generateQueryEmbedding generates a query embedding vector.
func (a *Adapter) generateQueryEmbedding(ctx context.Context, query string) ([]float32, error) {
	model := inferEmbeddingModelFromDimensions(a.config.VectorDimensions)

	isOpenAI := strings.HasPrefix(model, "text-embedding-")
	var embedding64 []float64
	var err error

	if isOpenAI {
		if a.llmClient == nil {
			return nil, fmt.Errorf("Redis: SearchSemantic requires OPENAI_API_KEY for OpenAI embedding models")
		}
		embedding64, err = a.llmClient.GenerateEmbedding(ctx, query, model)
	} else {
		var provider providers.EmbeddingProvider
		provider, err = providers.CreateProvider(ctx, model)
		if err != nil {
			return nil, fmt.Errorf("Redis: failed to create embedding provider for %q: %w", model, err)
		}
		embedding64, err = provider.GenerateEmbedding(ctx, query)
	}
	if err != nil {
		return nil, fmt.Errorf("Redis: failed to generate query embedding: %w", err)
	}

	// Convert float64 → float32
	vec := make([]float32, len(embedding64))
	for i, v := range embedding64 {
		vec[i] = float32(v)
	}
	return vec, nil
}

// toQueryResults converts SearchResults to vectordb.QueryResults.
func toQueryResults(results []SearchResult) []*vectordb.QueryResult {
	qr := make([]*vectordb.QueryResult, len(results))
	for i, r := range results {
		qr[i] = &vectordb.QueryResult{
			Document: *r.Document,
			Score:    r.Score,
		}
	}
	return qr
}

// parseRedisAddress extracts host:port from config.
func parseRedisAddress(config *vectordb.Config) string {
	if config.Address != "" {
		return config.Address
	}
	if config.URL != "" {
		// Strip redis:// prefix if present
		addr := config.URL
		for _, prefix := range []string{"redis://", "rediss://", "redis-stack://"} {
			if strings.HasPrefix(addr, prefix) {
				addr = addr[len(prefix):]
				break
			}
		}
		// Strip path/query
		if idx := strings.Index(addr, "/"); idx >= 0 {
			addr = addr[:idx]
		}
		// Strip userinfo
		if idx := strings.Index(addr, "@"); idx >= 0 {
			addr = addr[idx+1:]
		}
		return addr
	}
	return "localhost:6379"
}

// inferEmbeddingModelFromDimensions maps dimension count to a likely model.
func inferEmbeddingModelFromDimensions(dims int) string {
	switch dims {
	case 768:
		return "sentence-transformers/all-mpnet-base-v2"
	case 384:
		return "sentence-transformers/all-MiniLM-L6-v2"
	case 1536:
		return "text-embedding-3-small"
	case 3072:
		return "text-embedding-3-large"
	case 1024:
		return "nomic-embed-text"
	default:
		return "text-embedding-3-small"
	}
}
