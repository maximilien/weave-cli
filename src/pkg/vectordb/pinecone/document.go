// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package pinecone

import (
	"context"
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/pinecone-io/go-pinecone/pinecone"
	"google.golang.org/protobuf/types/known/structpb"
)

// CreateDocument creates a new document (vector) in Pinecone
func (a *Adapter) CreateDocument(ctx context.Context, collectionName string, doc *vectordb.Document) error {
	if a.client == nil {
		return fmt.Errorf("pinecone client not initialized")
	}

	// Get index connection
	idx, err := a.client.DescribeIndex(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to describe index: %w", err)
	}

	idxConn, err := a.client.Index(pinecone.NewIndexConnParams{Host: idx.Host})
	if err != nil {
		return fmt.Errorf("failed to connect to index: %w", err)
	}
	defer idxConn.Close()

	// Generate embedding for document content
	if doc.Content == "" && doc.Text == "" {
		return fmt.Errorf("document must have content or text")
	}

	// Use Content first, fallback to Text
	contentToEmbed := doc.Content
	if contentToEmbed == "" {
		contentToEmbed = doc.Text
	}

	var embedding []float32
	if a.llmClient != nil {
		embeddingFloat64, err := a.llmClient.GenerateEmbedding(ctx, contentToEmbed, "text-embedding-3-small")
		if err != nil {
			return fmt.Errorf("failed to generate embedding: %w", err)
		}
		// Convert []float64 to []float32
		embedding = make([]float32, len(embeddingFloat64))
		for i, v := range embeddingFloat64 {
			embedding[i] = float32(v)
		}
	} else {
		return fmt.Errorf("OpenAI client not configured for embedding generation")
	}

	// Convert metadata to Pinecone format
	metadataMap := make(map[string]interface{})
	if doc.Metadata != nil {
		for k, v := range doc.Metadata {
			metadataMap[k] = v
		}
	}
	// Add content to metadata
	metadataMap["content"] = contentToEmbed

	metadata, err := structpb.NewStruct(metadataMap)
	if err != nil {
		return fmt.Errorf("failed to create metadata: %w", err)
	}

	// Create vector
	vector := &pinecone.Vector{
		Id:       doc.ID,
		Values:   embedding,
		Metadata: metadata,
	}

	// Upsert the vector
	_, err = idxConn.UpsertVectors(ctx, []*pinecone.Vector{vector})
	if err != nil {
		return fmt.Errorf("failed to upsert vector: %w", err)
	}

	return nil
}

// GetDocument retrieves a document by ID from Pinecone
func (a *Adapter) GetDocument(ctx context.Context, collectionName, id string) (*vectordb.Document, error) {
	if a.client == nil {
		return nil, fmt.Errorf("pinecone client not initialized")
	}

	// Get index connection
	idx, err := a.client.DescribeIndex(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to describe index: %w", err)
	}

	idxConn, err := a.client.Index(pinecone.NewIndexConnParams{Host: idx.Host})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to index: %w", err)
	}
	defer idxConn.Close()

	// Fetch the vector
	resp, err := idxConn.FetchVectors(ctx, []string{id})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch vector: %w", err)
	}

	// Check if vector exists
	if len(resp.Vectors) == 0 {
		return nil, fmt.Errorf("document not found")
	}

	vector := resp.Vectors[id]
	if vector == nil {
		return nil, fmt.Errorf("document not found")
	}

	// Convert to vectordb.Document
	doc := &vectordb.Document{
		ID: vector.Id,
	}

	// Extract metadata and content
	if vector.Metadata != nil {
		metadataMap := vector.Metadata.AsMap()
		doc.Metadata = make(map[string]interface{})
		for k, v := range metadataMap {
			if k == "content" {
				if content, ok := v.(string); ok {
					doc.Content = content
				}
			} else {
				doc.Metadata[k] = v
			}
		}
	}

	return doc, nil
}

// UpdateDocument updates an existing document in Pinecone
func (a *Adapter) UpdateDocument(ctx context.Context, collectionName string, doc *vectordb.Document) error {
	// In Pinecone, update is done via upsert - just reuse CreateDocument
	return a.CreateDocument(ctx, collectionName, doc)
}

// DeleteDocument deletes a document by ID from Pinecone
func (a *Adapter) DeleteDocument(ctx context.Context, collectionName, id string) error {
	if a.client == nil {
		return fmt.Errorf("pinecone client not initialized")
	}

	// Get index connection
	idx, err := a.client.DescribeIndex(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to describe index: %w", err)
	}

	idxConn, err := a.client.Index(pinecone.NewIndexConnParams{Host: idx.Host})
	if err != nil {
		return fmt.Errorf("failed to connect to index: %w", err)
	}
	defer idxConn.Close()

	// Delete the vector
	err = idxConn.DeleteVectorsById(ctx, []string{id})
	if err != nil {
		return fmt.Errorf("failed to delete vector: %w", err)
	}

	return nil
}

// ListDocuments lists documents in a Pinecone index
func (a *Adapter) ListDocuments(ctx context.Context, collectionName string, limit int, offset int) ([]*vectordb.Document, error) {
	if a.client == nil {
		return nil, fmt.Errorf("pinecone client not initialized")
	}

	// Get index connection
	idx, err := a.client.DescribeIndex(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to describe index: %w", err)
	}

	idxConn, err := a.client.Index(pinecone.NewIndexConnParams{Host: idx.Host})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to index: %w", err)
	}
	defer idxConn.Close()

	// List vector IDs
	limit32 := uint32(limit)
	listResp, err := idxConn.ListVectors(ctx, &pinecone.ListVectorsRequest{
		Limit: &limit32,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list vectors: %w", err)
	}

	if len(listResp.VectorIds) == 0 {
		return []*vectordb.Document{}, nil
	}

	// Fetch full vectors
	vectorIDs := make([]string, 0, len(listResp.VectorIds))
	for _, vecID := range listResp.VectorIds {
		if vecID != nil {
			vectorIDs = append(vectorIDs, *vecID)
		}
	}

	fetchResp, err := idxConn.FetchVectors(ctx, vectorIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch vectors: %w", err)
	}

	// Convert to documents
	docs := make([]*vectordb.Document, 0, len(vectorIDs))
	if fetchResp.Vectors != nil {
		for _, id := range vectorIDs {
			if vector, ok := fetchResp.Vectors[id]; ok && vector != nil {
				doc := &vectordb.Document{
					ID: vector.Id,
				}

				// Extract metadata and content
				if vector.Metadata != nil {
					metadataMap := vector.Metadata.AsMap()
					doc.Metadata = make(map[string]interface{})
					for k, v := range metadataMap {
						if k == "content" {
							if content, ok := v.(string); ok {
								doc.Content = content
							}
						} else {
							doc.Metadata[k] = v
						}
					}
				}

				docs = append(docs, doc)
			}
		}
	}

	return docs, nil
}

// CreateDocuments batch creates multiple documents in Pinecone
func (a *Adapter) CreateDocuments(ctx context.Context, collectionName string, docs []*vectordb.Document) error {
	if a.client == nil {
		return fmt.Errorf("pinecone client not initialized")
	}

	if len(docs) == 0 {
		return nil
	}

	// Get index connection
	idx, err := a.client.DescribeIndex(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to describe index: %w", err)
	}

	idxConn, err := a.client.Index(pinecone.NewIndexConnParams{Host: idx.Host})
	if err != nil {
		return fmt.Errorf("failed to connect to index: %w", err)
	}
	defer idxConn.Close()

	// Prepare vectors
	vectors := make([]*pinecone.Vector, 0, len(docs))
	for _, doc := range docs {
		// Use Content first, fallback to Text
		contentToEmbed := doc.Content
		if contentToEmbed == "" {
			contentToEmbed = doc.Text
		}

		if contentToEmbed == "" {
			return fmt.Errorf("document %s must have content or text", doc.ID)
		}

		// Generate embedding
		var embedding []float32
		if a.llmClient != nil {
			embeddingFloat64, err := a.llmClient.GenerateEmbedding(ctx, contentToEmbed, "text-embedding-3-small")
			if err != nil {
				return fmt.Errorf("failed to generate embedding for doc %s: %w", doc.ID, err)
			}
			embedding = make([]float32, len(embeddingFloat64))
			for i, v := range embeddingFloat64 {
				embedding[i] = float32(v)
			}
		} else {
			return fmt.Errorf("OpenAI client not configured for embedding generation")
		}

		// Prepare metadata
		metadataMap := make(map[string]interface{})
		if doc.Metadata != nil {
			for k, v := range doc.Metadata {
				metadataMap[k] = v
			}
		}
		// Store content in metadata
		metadataMap["content"] = contentToEmbed

		metadata, err := structpb.NewStruct(metadataMap)
		if err != nil {
			return fmt.Errorf("failed to create metadata for doc %s: %w", doc.ID, err)
		}

		vectors = append(vectors, &pinecone.Vector{
			Id:       doc.ID,
			Values:   embedding,
			Metadata: metadata,
		})
	}

	// Batch upsert
	_, err = idxConn.UpsertVectors(ctx, vectors)
	if err != nil {
		return fmt.Errorf("failed to upsert vectors: %w", err)
	}

	return nil
}

// DeleteDocuments batch deletes multiple documents from Pinecone
func (a *Adapter) DeleteDocuments(ctx context.Context, collectionName string, ids []string) error {
	if a.client == nil {
		return fmt.Errorf("pinecone client not initialized")
	}

	if len(ids) == 0 {
		return nil
	}

	// Get index connection
	idx, err := a.client.DescribeIndex(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to describe index: %w", err)
	}

	idxConn, err := a.client.Index(pinecone.NewIndexConnParams{Host: idx.Host})
	if err != nil {
		return fmt.Errorf("failed to connect to index: %w", err)
	}
	defer idxConn.Close()

	// Delete vectors
	err = idxConn.DeleteVectorsById(ctx, ids)
	if err != nil {
		return fmt.Errorf("failed to delete vectors: %w", err)
	}

	return nil
}

// DeleteDocumentsByMetadata deletes documents by metadata filter
func (a *Adapter) DeleteDocumentsByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}) error {
	if a.client == nil {
		return fmt.Errorf("pinecone client not initialized")
	}

	// Get index connection
	idx, err := a.client.DescribeIndex(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to describe index: %w", err)
	}

	idxConn, err := a.client.Index(pinecone.NewIndexConnParams{Host: idx.Host})
	if err != nil {
		return fmt.Errorf("failed to connect to index: %w", err)
	}
	defer idxConn.Close()

	// Convert metadata to Pinecone filter
	filterMap := make(map[string]interface{})
	for k, v := range metadata {
		filterMap[k] = map[string]interface{}{"$eq": v}
	}

	filter, err := structpb.NewStruct(filterMap)
	if err != nil {
		return fmt.Errorf("failed to create filter: %w", err)
	}

	// Delete by filter
	err = idxConn.DeleteVectorsByFilter(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete by filter: %w", err)
	}

	return nil
}

// GetDocumentsByMetadata retrieves documents by metadata filter
func (a *Adapter) GetDocumentsByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}, limit int) ([]*vectordb.Document, error) {
	// For Pinecone, we need to use query with metadata filter
	// Since we don't have a query vector, we'll use a dummy approach:
	// List all vectors and filter client-side (not efficient, but works for small datasets)
	docs, err := a.ListDocuments(ctx, collectionName, limit*10, 0) // Get more to filter
	if err != nil {
		return nil, err
	}

	// Filter by metadata
	filtered := make([]*vectordb.Document, 0)
	for _, doc := range docs {
		matches := true
		for k, v := range metadata {
			if doc.Metadata == nil || doc.Metadata[k] != v {
				matches = false
				break
			}
		}
		if matches {
			filtered = append(filtered, doc)
			if len(filtered) >= limit {
				break
			}
		}
	}

	return filtered, nil
}
