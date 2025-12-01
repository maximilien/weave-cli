// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package neo4j

import (
	"context"
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/llm"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// Adapter wraps the Neo4j client to implement the vectordb.VectorDBClient interface
type Adapter struct {
	client    *Client
	llmClient *llm.OpenAIClient
}

// Health checks the health of the Neo4j instance
func (a *Adapter) Health(ctx context.Context) error {
	return a.client.Health(ctx)
}

// CreateCollection creates a new collection with the given schema
func (a *Adapter) CreateCollection(ctx context.Context, name string, schema *vectordb.CollectionSchema) error {
	// Use config defaults or schema settings
	vectorDims := a.client.config.VectorDimensions
	similarityMetric := a.client.config.SimilarityMetric

	return a.client.CreateCollection(ctx, name, vectorDims, similarityMetric)
}

// DeleteCollection deletes a collection and all its documents
func (a *Adapter) DeleteCollection(ctx context.Context, name string) error {
	return a.client.DeleteCollection(ctx, name, true)
}

// ListCollections returns a list of all collections
func (a *Adapter) ListCollections(ctx context.Context) ([]vectordb.CollectionInfo, error) {
	names, err := a.client.ListCollections(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to CollectionInfo
	collections := make([]vectordb.CollectionInfo, len(names))
	for i, name := range names {
		count, err := a.client.GetCollectionCount(ctx, name)
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

// CollectionExists checks if a collection exists
func (a *Adapter) CollectionExists(ctx context.Context, name string) (bool, error) {
	return a.client.CollectionExists(ctx, name)
}

// GetCollectionCount returns the number of documents in a collection
func (a *Adapter) GetCollectionCount(ctx context.Context, name string) (int64, error) {
	return a.client.GetCollectionCount(ctx, name)
}

// CreateDocument creates a new document in the specified collection
func (a *Adapter) CreateDocument(ctx context.Context, collectionName string, document *vectordb.Document) error {
	// Convert vectordb.Document to Neo4j Document
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

	return a.client.CreateDocument(ctx, collectionName, doc)
}

// CreateDocuments creates multiple documents in batch
func (a *Adapter) CreateDocuments(ctx context.Context, collectionName string, documents []*vectordb.Document) error {
	// Convert documents
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

	return a.client.BatchCreateDocuments(ctx, collectionName, docs)
}

// GetDocument retrieves a document by ID
func (a *Adapter) GetDocument(ctx context.Context, collectionName, documentID string) (*vectordb.Document, error) {
	doc, err := a.client.GetDocument(ctx, collectionName, documentID)
	if err != nil {
		return nil, err
	}

	return &vectordb.Document{
		ID:       doc.ID,
		Text:     doc.Content,
		Metadata: doc.Metadata,
	}, nil
}

// UpdateDocument updates an existing document
func (a *Adapter) UpdateDocument(ctx context.Context, collectionName string, document *vectordb.Document) error {
	doc := &Document{
		ID:       document.ID,
		Content:  document.Text,
		Metadata: document.Metadata,
	}

	// Generate embedding if needed
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

	return a.client.UpdateDocument(ctx, collectionName, doc)
}

// DeleteDocument deletes a document by ID
func (a *Adapter) DeleteDocument(ctx context.Context, collectionName, documentID string) error {
	return a.client.DeleteDocument(ctx, collectionName, documentID)
}

// DeleteDocuments deletes multiple documents by IDs
func (a *Adapter) DeleteDocuments(ctx context.Context, collectionName string, documentIDs []string) error {
	return a.client.DeleteDocuments(ctx, collectionName, documentIDs)
}

// DeleteDocumentsByMetadata deletes documents matching metadata criteria
func (a *Adapter) DeleteDocumentsByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}) error {
	// Neo4j doesn't have a direct metadata-based delete, so we need to:
	// 1. Find documents matching the criteria
	// 2. Delete them by ID

	// Build WHERE clause
	whereClause := "WHERE "
	conditions := []string{}
	for key := range metadata {
		conditions = append(conditions, fmt.Sprintf("n.%s = $meta_%s", key, key))
	}
	whereClause += conditions[0]
	for i := 1; i < len(conditions); i++ {
		whereClause += " AND " + conditions[i]
	}

	query := fmt.Sprintf(`
		MATCH (n:%s)
		%s
		DETACH DELETE n
	`, collectionName, whereClause)

	params := make(map[string]interface{})
	for key, value := range metadata {
		params[fmt.Sprintf("meta_%s", key)] = value
	}

	_, err := a.client.executeQuery(ctx, query, params)
	return err
}

// ListDocuments returns a list of documents in a collection
func (a *Adapter) ListDocuments(ctx context.Context, collectionName string, limit int, offset int) ([]*vectordb.Document, error) {
	// Neo4j's ListDocuments doesn't support offset, so we'll fetch with limit and skip in the query
	query := fmt.Sprintf(`
		MATCH (d:%s)
		RETURN d.id as id, d.content as content, d.embedding as embedding, properties(d) as props
		SKIP $offset
		LIMIT $limit
	`, collectionName)

	if limit <= 0 {
		limit = 100
	}

	params := map[string]interface{}{
		"limit":  limit,
		"offset": offset,
	}

	result, err := a.client.executeQuery(ctx, query, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}

	var docs []*vectordb.Document
	for _, record := range result.Records {
		doc := &vectordb.Document{
			Metadata: make(map[string]interface{}),
		}

		if val, ok := record.Get("id"); ok {
			if id, ok := val.(string); ok {
				doc.ID = id
			}
		}
		if val, ok := record.Get("content"); ok {
			if content, ok := val.(string); ok {
				doc.Text = content
			}
		}
		// Skip vector - not included in vectordb.Document
		if val, ok := record.Get("props"); ok {
			if props, ok := val.(map[string]interface{}); ok {
				for k, v := range props {
					if k != "id" && k != "content" && k != "embedding" {
						doc.Metadata[k] = v
					}
				}
			}
		}

		docs = append(docs, doc)
	}

	return docs, nil
}

// SearchSemantic performs semantic search using vector embeddings
func (a *Adapter) SearchSemantic(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	// Generate embedding for the query
	if a.llmClient == nil {
		return nil, fmt.Errorf("LLM client not available for generating query embeddings")
	}

	embedding, err := a.llmClient.GenerateEmbedding(ctx, query, "text-embedding-3-small")
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Convert []float64 to []float32
	embeddingFloat32 := make([]float32, len(embedding))
	for i, v := range embedding {
		embeddingFloat32[i] = float32(v)
	}

	limit := 10
	if options != nil && options.TopK > 0 {
		limit = options.TopK
	}

	// Perform vector search
	results, err := a.client.VectorSearch(ctx, collectionName, embeddingFloat32, limit)
	if err != nil {
		return nil, err
	}

	// Convert to vectordb.QueryResult
	queryResults := make([]*vectordb.QueryResult, len(results))
	for i, result := range results {
		queryResults[i] = &vectordb.QueryResult{
			Document: vectordb.Document{
				ID:       result.Document.ID,
				Text:     result.Document.Content,
				Metadata: result.Document.Metadata,
			},
			Score: result.Score,
		}
	}

	return queryResults, nil
}

// SearchBM25 performs keyword-based search using BM25 (not supported by Neo4j)
func (a *Adapter) SearchBM25(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	return nil, fmt.Errorf("BM25 search is not supported by Neo4j")
}

// SearchHybrid performs hybrid search (not supported by Neo4j)
func (a *Adapter) SearchHybrid(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	return nil, fmt.Errorf("hybrid search is not supported by Neo4j")
}

// SearchByMetadata searches documents by metadata fields
func (a *Adapter) SearchByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	// Build WHERE clause
	whereClause := "WHERE "
	conditions := []string{}
	for key := range metadata {
		conditions = append(conditions, fmt.Sprintf("n.%s = $meta_%s", key, key))
	}
	whereClause += conditions[0]
	for i := 1; i < len(conditions); i++ {
		whereClause += " AND " + conditions[i]
	}

	limit := 10
	if options != nil && options.TopK > 0 {
		limit = options.TopK
	}

	query := fmt.Sprintf(`
		MATCH (n:%s)
		%s
		RETURN n.id as id, n.content as content, properties(n) as props
		LIMIT $limit
	`, collectionName, whereClause)

	params := make(map[string]interface{})
	params["limit"] = limit
	for key, value := range metadata {
		params[fmt.Sprintf("meta_%s", key)] = value
	}

	result, err := a.client.executeQuery(ctx, query, params)
	if err != nil {
		return nil, fmt.Errorf("failed to search by metadata: %w", err)
	}

	var results []*vectordb.QueryResult
	for _, record := range result.Records {
		doc := vectordb.Document{
			Metadata: make(map[string]interface{}),
		}

		if val, ok := record.Get("id"); ok {
			if id, ok := val.(string); ok {
				doc.ID = id
			}
		}
		if val, ok := record.Get("content"); ok {
			if content, ok := val.(string); ok {
				doc.Text = content
			}
		}
		// Skip vector - not included in vectordb.Document
		if val, ok := record.Get("props"); ok {
			if props, ok := val.(map[string]interface{}); ok {
				for k, v := range props {
					if k != "id" && k != "content" && k != "embedding" {
						doc.Metadata[k] = v
					}
				}
			}
		}

		results = append(results, &vectordb.QueryResult{
			Document: doc,
			Score:    1.0, // Metadata search doesn't have a score
		})
	}

	return results, nil
}

// GetSchema retrieves the schema for a collection
func (a *Adapter) GetSchema(ctx context.Context, collectionName string) (*vectordb.CollectionSchema, error) {
	_, err := a.client.GetCollectionInfo(ctx, collectionName)
	if err != nil {
		return nil, err
	}

	return &vectordb.CollectionSchema{
		Class: collectionName,
	}, nil
}

// UpdateSchema updates the schema for a collection (not fully supported)
func (a *Adapter) UpdateSchema(ctx context.Context, collectionName string, schema *vectordb.CollectionSchema) error {
	return fmt.Errorf("schema updates are not supported by Neo4j - delete and recreate the collection instead")
}

// GetDefaultSchema returns a default schema for the given type
func (a *Adapter) GetDefaultSchema(schemaType vectordb.SchemaType, collectionName string) *vectordb.CollectionSchema {
	return &vectordb.CollectionSchema{
		Class: collectionName,
	}
}

// ValidateSchema validates a schema definition
func (a *Adapter) ValidateSchema(schema *vectordb.CollectionSchema) error {
	if schema.Class == "" {
		return fmt.Errorf("collection name is required")
	}
	return nil
}
