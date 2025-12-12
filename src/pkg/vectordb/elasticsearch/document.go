// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v9/esutil"
	"github.com/elastic/go-elasticsearch/v9/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// elasticsearchDocument represents the Elasticsearch document structure
type elasticsearchDocument struct {
	DocumentID  string                 `json:"document_id"`
	Text        string                 `json:"text"`
	Content     string                 `json:"content"`
	Image       string                 `json:"image,omitempty"`
	ImageData   string                 `json:"image_data,omitempty"`
	URL         string                 `json:"url,omitempty"`
	VectorField []float64              `json:"vector_field,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// CreateDocument creates a single document in the specified index
func (a *Adapter) CreateDocument(ctx context.Context, collectionName string, document *vectordb.Document) error {
	ctx, cancel := context.WithTimeout(ctx, a.getTimeout())
	defer cancel()

	// Generate embedding if LLM client is available and text is provided
	var embedding []float64
	if a.llmClient != nil && document.Content != "" {
		var err error
		embedding, err = a.llmClient.GenerateEmbedding(ctx, document.Content, "")
		if err != nil {
			return fmt.Errorf("failed to create embedding: %w", err)
		}
	}

	// Convert to Elasticsearch document
	now := time.Now()
	esDoc := &elasticsearchDocument{
		DocumentID:  document.ID,
		Text:        document.Text,
		Content:     document.Content,
		Image:       document.Image,
		ImageData:   document.ImageData,
		URL:         document.URL,
		VectorField: embedding,
		Metadata:    document.Metadata,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Index the document
	res, err := a.client.Index(collectionName).
		Id(document.ID).
		Document(esDoc).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("failed to index document: %w", err)
	}

	// Check if indexing was successful (created or updated)
	_ = res // Result is validated by Do() returning no error

	return nil
}

// CreateDocuments creates multiple documents in batch using BulkIndexer
func (a *Adapter) CreateDocuments(ctx context.Context, collectionName string, documents []*vectordb.Document) error {
	if len(documents) == 0 {
		return nil
	}

	// Create bulk indexer
	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Client: a.client,
		Index:  collectionName,
	})
	if err != nil {
		return fmt.Errorf("failed to create bulk indexer: %w", err)
	}

	// Add documents to bulk indexer
	for _, doc := range documents {
		// Generate embedding if LLM client is available
		var embedding []float64
		if a.llmClient != nil && doc.Content != "" {
			embedding, err = a.llmClient.GenerateEmbedding(ctx, doc.Content, "")
			if err != nil {
				return fmt.Errorf("failed to create embedding for doc %s: %w", doc.ID, err)
			}
		}

		// Convert to Elasticsearch document
		now := time.Now()
		esDoc := &elasticsearchDocument{
			DocumentID:  doc.ID,
			Text:        doc.Text,
			Content:     doc.Content,
			Image:       doc.Image,
			ImageData:   doc.ImageData,
			URL:         doc.URL,
			VectorField: embedding,
			Metadata:    doc.Metadata,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		// Encode document
		data, err := json.Marshal(esDoc)
		if err != nil {
			return fmt.Errorf("failed to encode document %s: %w", doc.ID, err)
		}

		// Add to bulk indexer
		err = bi.Add(ctx, esutil.BulkIndexerItem{
			Action:     "index",
			DocumentID: doc.ID,
			Body:       strings.NewReader(string(data)),
			OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
				if err != nil {
					fmt.Printf("ERROR: Failed to index document %s: %v\n", item.DocumentID, err)
				} else {
					fmt.Printf("ERROR: Failed to index document %s: %s\n", item.DocumentID, res.Error.Reason)
				}
			},
		})
		if err != nil {
			return fmt.Errorf("failed to add document %s to bulk: %w", doc.ID, err)
		}
	}

	// Close bulk indexer to flush remaining items
	if err := bi.Close(ctx); err != nil {
		return fmt.Errorf("failed to close bulk indexer: %w", err)
	}

	// Check for errors
	stats := bi.Stats()
	if stats.NumFailed > 0 {
		return fmt.Errorf("bulk indexing failed: %d documents failed out of %d", stats.NumFailed, len(documents))
	}

	return nil
}

// GetDocument retrieves a document by ID
func (a *Adapter) GetDocument(ctx context.Context, collectionName, documentID string) (*vectordb.Document, error) {
	ctx, cancel := context.WithTimeout(ctx, a.getTimeout())
	defer cancel()

	res, err := a.client.Get(collectionName, documentID).Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	if !res.Found {
		return nil, fmt.Errorf("document %s not found in index %s", documentID, collectionName)
	}

	// Decode source
	var esDoc elasticsearchDocument
	if err := json.Unmarshal(res.Source_, &esDoc); err != nil {
		return nil, fmt.Errorf("failed to decode document: %w", err)
	}

	// Convert to vectordb.Document
	doc := &vectordb.Document{
		ID:        esDoc.DocumentID,
		Text:      esDoc.Text,
		Content:   esDoc.Content,
		Image:     esDoc.Image,
		ImageData: esDoc.ImageData,
		URL:       esDoc.URL,
		Metadata:  esDoc.Metadata,
	}

	return doc, nil
}

// UpdateDocument updates an existing document
func (a *Adapter) UpdateDocument(ctx context.Context, collectionName string, document *vectordb.Document) error {
	ctx, cancel := context.WithTimeout(ctx, a.getTimeout())
	defer cancel()

	// Generate embedding if content changed and LLM client available
	var embedding []float64
	if a.llmClient != nil && document.Content != "" {
		var err error
		embedding, err = a.llmClient.GenerateEmbedding(ctx, document.Content, "")
		if err != nil {
			return fmt.Errorf("failed to create embedding: %w", err)
		}
	}

	// Prepare update document
	updateDoc := map[string]interface{}{
		"document_id": document.ID,
		"text":        document.Text,
		"content":     document.Content,
		"image":       document.Image,
		"image_data":  document.ImageData,
		"url":         document.URL,
		"metadata":    document.Metadata,
		"updated_at":  time.Now(),
	}

	if len(embedding) > 0 {
		updateDoc["vector_field"] = embedding
	}

	// Update the document using fluent API
	res, err := a.client.Update(collectionName, document.ID).
		Doc(updateDoc).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	// Check if update was successful
	_ = res // Result is validated by Do() returning no error

	return nil
}

// DeleteDocument deletes a document by ID
func (a *Adapter) DeleteDocument(ctx context.Context, collectionName, documentID string) error {
	ctx, cancel := context.WithTimeout(ctx, a.getTimeout())
	defer cancel()

	res, err := a.client.Delete(collectionName, documentID).Do(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	// Check if delete was successful
	_ = res // Result is validated by Do() returning no error

	return nil
}

// DeleteDocuments deletes multiple documents by IDs
func (a *Adapter) DeleteDocuments(ctx context.Context, collectionName string, documentIDs []string) error {
	if len(documentIDs) == 0 {
		return nil
	}

	// Create bulk indexer for deletions
	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Client: a.client,
		Index:  collectionName,
	})
	if err != nil {
		return fmt.Errorf("failed to create bulk indexer: %w", err)
	}

	// Add delete operations
	for _, docID := range documentIDs {
		err = bi.Add(ctx, esutil.BulkIndexerItem{
			Action:     "delete",
			DocumentID: docID,
			OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
				if err != nil {
					fmt.Printf("ERROR: Failed to delete document %s: %v\n", item.DocumentID, err)
				}
			},
		})
		if err != nil {
			return fmt.Errorf("failed to add delete operation for %s: %w", docID, err)
		}
	}

	// Close bulk indexer
	if err := bi.Close(ctx); err != nil {
		return fmt.Errorf("failed to close bulk indexer: %w", err)
	}

	return nil
}

// DeleteDocumentsByMetadata deletes documents matching metadata criteria
func (a *Adapter) DeleteDocumentsByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, a.getTimeout())
	defer cancel()

	// Build query for metadata matching
	var filters []types.Query
	for key, value := range metadata {
		fieldName := fmt.Sprintf("metadata.%s", key)
		termQuery := types.Query{
			Term: map[string]types.TermQuery{
				fieldName: {Value: value},
			},
		}
		filters = append(filters, termQuery)
	}

	query := &types.Query{
		Bool: &types.BoolQuery{
			Must: filters,
		},
	}

	// Use DeleteByQuery API
	_, err := a.client.DeleteByQuery(collectionName).
		Query(query).
		Do(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete documents by metadata: %w", err)
	}

	return nil
}

// ListDocuments returns a paginated list of documents
func (a *Adapter) ListDocuments(ctx context.Context, collectionName string, limit int, offset int) ([]*vectordb.Document, error) {
	ctx, cancel := context.WithTimeout(ctx, a.getTimeout())
	defer cancel()

	// Default limit
	if limit <= 0 {
		limit = 10
	}

	// Build search request
	req := &search.Request{
		Query: &types.Query{
			MatchAll: &types.MatchAllQuery{},
		},
		Size: &limit,
		From: &offset,
	}

	res, err := a.client.Search().
		Index(collectionName).
		Request(req).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}

	// Convert hits to documents
	var documents []*vectordb.Document
	for _, hit := range res.Hits.Hits {
		var esDoc elasticsearchDocument
		if err := json.Unmarshal(hit.Source_, &esDoc); err != nil {
			continue // Skip malformed documents
		}

		doc := &vectordb.Document{
			ID:        esDoc.DocumentID,
			Text:      esDoc.Text,
			Content:   esDoc.Content,
			Image:     esDoc.Image,
			ImageData: esDoc.ImageData,
			URL:       esDoc.URL,
			Metadata:  esDoc.Metadata,
		}
		documents = append(documents, doc)
	}

	return documents, nil
}
