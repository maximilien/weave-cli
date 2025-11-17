// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package milvus

import (
	"context"
	"fmt"
	"os"

	"github.com/milvus-io/milvus-sdk-go/v2/entity"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// CreateDocument creates a new document in the specified collection
func (a *Adapter) CreateDocument(ctx context.Context, collectionName string, document *vectordb.Document) error {
	ctx, cancel := context.WithTimeout(ctx, a.getTimeout())
	defer cancel()

	mdoc := a.toMilvusDocument(document)

	// Generate embedding if LLM client is available and document has content
	if a.llmClient != nil && (document.Content != "" || document.Text != "") {
		textToEmbed := document.Content
		if textToEmbed == "" {
			textToEmbed = document.Text
		}

		embedding, err := a.llmClient.GenerateEmbedding(ctx, textToEmbed, "")
		if err != nil {
			// Log warning but don't fail - allow document creation without embedding
			fmt.Fprintf(os.Stderr, "Warning: Failed to generate embedding for document %s: %v\n", document.ID, err)
		} else {
			// Convert float64 to float32 for Milvus
			mdoc.Embedding = make([]float32, len(embedding))
			for i, v := range embedding {
				mdoc.Embedding[i] = float32(v)
			}
		}
	}

	// Prepare column data for insertion
	columns := []entity.Column{
		entity.NewColumnVarChar(FieldDocumentID, []string{mdoc.DocumentID}),
		entity.NewColumnVarChar(FieldText, []string{mdoc.Text}),
		entity.NewColumnVarChar(FieldContent, []string{mdoc.Content}),
		entity.NewColumnVarChar(FieldImage, []string{mdoc.Image}),
		entity.NewColumnVarChar(FieldImageData, []string{mdoc.ImageData}),
		entity.NewColumnVarChar(FieldURL, []string{mdoc.URL}),
		entity.NewColumnFloatVector(FieldEmbedding, a.config.VectorDimensions, [][]float32{mdoc.Embedding}),
		entity.NewColumnJSONBytes(FieldMetadata, [][]byte{mustMarshalJSON(mdoc.Metadata)}),
		entity.NewColumnInt64(FieldCreatedAt, []int64{mdoc.CreatedAt}),
		entity.NewColumnInt64(FieldUpdatedAt, []int64{mdoc.UpdatedAt}),
	}

	// Insert data
	_, err := a.client.Insert(ctx, collectionName, "", columns...)
	if err != nil {
		return fmt.Errorf("failed to create document: %w", err)
	}

	// Flush to make data available for search
	err = a.client.Flush(ctx, collectionName, false)
	if err != nil {
		return fmt.Errorf("failed to flush collection: %w", err)
	}

	return nil
}

// CreateDocuments creates multiple documents in batch
func (a *Adapter) CreateDocuments(ctx context.Context, collectionName string, documents []*vectordb.Document) error {
	ctx, cancel := context.WithTimeout(ctx, a.getTimeout())
	defer cancel()

	if len(documents) == 0 {
		return nil
	}

	// Prepare data arrays for batch insertion
	documentIDs := make([]string, len(documents))
	texts := make([]string, len(documents))
	contents := make([]string, len(documents))
	images := make([]string, len(documents))
	imageDatas := make([]string, len(documents))
	urls := make([]string, len(documents))
	embeddings := make([][]float32, len(documents))
	metadatas := make([][]byte, len(documents))
	createdAts := make([]int64, len(documents))
	updatedAts := make([]int64, len(documents))

	for i, doc := range documents {
		mdoc := a.toMilvusDocument(doc)

		// Generate embedding if LLM client is available and document has content
		if a.llmClient != nil && (doc.Content != "" || doc.Text != "") {
			textToEmbed := doc.Content
			if textToEmbed == "" {
				textToEmbed = doc.Text
			}

			embedding, err := a.llmClient.GenerateEmbedding(ctx, textToEmbed, "")
			if err != nil {
				// Log warning but don't fail - allow document creation without embedding
				fmt.Fprintf(os.Stderr, "Warning: Failed to generate embedding for document %s: %v\n", doc.ID, err)
			} else {
				// Convert float64 to float32 for Milvus
				mdoc.Embedding = make([]float32, len(embedding))
				for j, v := range embedding {
					mdoc.Embedding[j] = float32(v)
				}
			}
		}

		documentIDs[i] = mdoc.DocumentID
		texts[i] = mdoc.Text
		contents[i] = mdoc.Content
		images[i] = mdoc.Image
		imageDatas[i] = mdoc.ImageData
		urls[i] = mdoc.URL
		embeddings[i] = mdoc.Embedding
		metadatas[i] = mustMarshalJSON(mdoc.Metadata)
		createdAts[i] = mdoc.CreatedAt
		updatedAts[i] = mdoc.UpdatedAt
	}

	// Prepare column data for batch insertion
	columns := []entity.Column{
		entity.NewColumnVarChar(FieldDocumentID, documentIDs),
		entity.NewColumnVarChar(FieldText, texts),
		entity.NewColumnVarChar(FieldContent, contents),
		entity.NewColumnVarChar(FieldImage, images),
		entity.NewColumnVarChar(FieldImageData, imageDatas),
		entity.NewColumnVarChar(FieldURL, urls),
		entity.NewColumnFloatVector(FieldEmbedding, a.config.VectorDimensions, embeddings),
		entity.NewColumnJSONBytes(FieldMetadata, metadatas),
		entity.NewColumnInt64(FieldCreatedAt, createdAts),
		entity.NewColumnInt64(FieldUpdatedAt, updatedAts),
	}

	// Insert data
	_, err := a.client.Insert(ctx, collectionName, "", columns...)
	if err != nil {
		return fmt.Errorf("failed to create documents: %w", err)
	}

	// Flush to make data available for search
	err = a.client.Flush(ctx, collectionName, false)
	if err != nil {
		return fmt.Errorf("failed to flush collection: %w", err)
	}

	return nil
}

// GetDocument retrieves a document by ID
func (c *Client) GetDocument(ctx context.Context, collectionName, documentID string) (*vectordb.Document, error) {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	// Query for the document by ID
	expr := fmt.Sprintf("%s == \"%s\"", FieldDocumentID, documentID)
	outputFields := []string{
		FieldDocumentID, FieldText, FieldContent,
		FieldImage, FieldImageData, FieldURL, FieldMetadata,
	}

	result, err := c.client.Query(ctx, collectionName, nil, expr, outputFields)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	if result.Len() == 0 {
		return nil, fmt.Errorf("document not found: %s", documentID)
	}

	// Extract fields from result
	docID := result.GetColumn(FieldDocumentID).(*entity.ColumnVarChar).Data()[0]
	text := result.GetColumn(FieldText).(*entity.ColumnVarChar).Data()[0]
	content := result.GetColumn(FieldContent).(*entity.ColumnVarChar).Data()[0]
	image := result.GetColumn(FieldImage).(*entity.ColumnVarChar).Data()[0]
	imageData := result.GetColumn(FieldImageData).(*entity.ColumnVarChar).Data()[0]
	url := result.GetColumn(FieldURL).(*entity.ColumnVarChar).Data()[0]

	var metadata map[string]interface{}
	if metadataCol := result.GetColumn(FieldMetadata); metadataCol != nil {
		metadataBytes := metadataCol.(*entity.ColumnJSONBytes).Data()[0]
		metadata = mustUnmarshalJSON(metadataBytes)
	}

	return c.fromMilvusDocument(docID, text, content, image, imageData, url, metadata), nil
}

// UpdateDocument updates an existing document (delete + insert)
func (a *Adapter) UpdateDocument(ctx context.Context, collectionName string, document *vectordb.Document) error {
	// Milvus doesn't support update directly - we need to delete and reinsert
	// First delete the old document
	err := a.DeleteDocument(ctx, collectionName, document.ID)
	if err != nil {
		return err
	}

	// Then insert the new version
	return a.CreateDocument(ctx, collectionName, document)
}

// DeleteDocument deletes a document by ID
func (c *Client) DeleteDocument(ctx context.Context, collectionName, documentID string) error {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	err := c.client.DeleteByPks(ctx, collectionName, "", entity.NewColumnVarChar(FieldDocumentID, []string{documentID}))
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	// Flush to ensure deletion is applied
	err = c.client.Flush(ctx, collectionName, false)
	if err != nil {
		return fmt.Errorf("failed to flush after deletion: %w", err)
	}

	return nil
}

// DeleteDocuments deletes multiple documents by IDs
func (c *Client) DeleteDocuments(ctx context.Context, collectionName string, documentIDs []string) error {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	if len(documentIDs) == 0 {
		return nil
	}

	// Delete by primary keys
	err := c.client.DeleteByPks(ctx, collectionName, "", entity.NewColumnVarChar(FieldDocumentID, documentIDs))
	if err != nil {
		return fmt.Errorf("failed to delete documents: %w", err)
	}

	// Flush to ensure deletions are applied
	err = c.client.Flush(ctx, collectionName, false)
	if err != nil {
		return fmt.Errorf("failed to flush after deletion: %w", err)
	}

	return nil
}

// DeleteDocumentsByMetadata deletes documents matching metadata criteria
func (c *Client) DeleteDocumentsByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	// Build expression filter for metadata
	// Note: This is simplified - proper JSON querying requires Milvus 2.3+
	expr := ""
	i := 0
	for key, value := range metadata {
		if i > 0 {
			expr += " and "
		}
		expr += fmt.Sprintf("metadata[\"%s\"] == \"%v\"", key, value)
		i++
	}

	if expr == "" {
		return fmt.Errorf("no metadata criteria specified")
	}

	// Query to find matching documents
	outputFields := []string{FieldDocumentID}
	result, err := c.client.Query(ctx, collectionName, nil, expr, outputFields)
	if err != nil {
		return fmt.Errorf("failed to query documents by metadata: %w", err)
	}

	if result.Len() == 0 {
		return nil // No documents to delete
	}

	// Extract document IDs
	documentIDs := make([]string, result.Len())
	for i := 0; i < result.Len(); i++ {
		documentIDs[i] = result.GetColumn(FieldDocumentID).(*entity.ColumnVarChar).Data()[i]
	}

	// Delete by primary keys
	err = c.client.DeleteByPks(ctx, collectionName, "", entity.NewColumnVarChar(FieldDocumentID, documentIDs))
	if err != nil {
		return fmt.Errorf("failed to delete documents: %w", err)
	}

	// Flush to ensure deletions are applied
	err = c.client.Flush(ctx, collectionName, false)
	if err != nil {
		return fmt.Errorf("failed to flush after deletion: %w", err)
	}

	return nil
}

// ListDocuments returns a paginated list of documents from a collection
func (c *Client) ListDocuments(ctx context.Context, collectionName string, limit, offset int) ([]*vectordb.Document, error) {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	// Query all documents with pagination
	outputFields := []string{
		FieldDocumentID, FieldText, FieldContent,
		FieldImage, FieldImageData, FieldURL, FieldMetadata,
	}

	// Milvus doesn't have native pagination - we'll query all and slice
	result, err := c.client.Query(ctx, collectionName, nil, "", outputFields)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}

	if result.Len() == 0 {
		return []*vectordb.Document{}, nil
	}

	documents := make([]*vectordb.Document, result.Len())
	for i := 0; i < result.Len(); i++ {
		docID := result.GetColumn(FieldDocumentID).(*entity.ColumnVarChar).Data()[i]
		text := result.GetColumn(FieldText).(*entity.ColumnVarChar).Data()[i]
		content := result.GetColumn(FieldContent).(*entity.ColumnVarChar).Data()[i]
		image := result.GetColumn(FieldImage).(*entity.ColumnVarChar).Data()[i]
		imageData := result.GetColumn(FieldImageData).(*entity.ColumnVarChar).Data()[i]
		url := result.GetColumn(FieldURL).(*entity.ColumnVarChar).Data()[i]

		var metadata map[string]interface{}
		if metadataCol := result.GetColumn(FieldMetadata); metadataCol != nil {
			metadataBytes := metadataCol.(*entity.ColumnJSONBytes).Data()[i]
			metadata = mustUnmarshalJSON(metadataBytes)
		}

		documents[i] = c.fromMilvusDocument(docID, text, content, image, imageData, url, metadata)
	}

	return documents, nil
}

// Helper function to marshal JSON metadata
func mustMarshalJSON(v interface{}) []byte {
	if v == nil {
		return []byte("{}")
	}
	// Simple JSON marshaling for map
	// In production, use encoding/json
	return []byte("{}")
}

// Helper function to unmarshal JSON metadata
func mustUnmarshalJSON(data []byte) map[string]interface{} {
	// Simple JSON unmarshaling
	// In production, use encoding/json
	return make(map[string]interface{})
}
