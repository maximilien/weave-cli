// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package milvus

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// CreateDocument creates a new document in the specified collection
func (a *Adapter) CreateDocument(ctx context.Context, collectionName string, document *vectordb.Document) error {
	ctx, cancel := context.WithTimeout(ctx, a.getTimeout())
	defer cancel()

	mdoc := a.toMilvusDocument(document)

	// CRITICAL FIX (v0.9.7): Truncate Image field if it exceeds VARCHAR limit
	// The Image field stores data URLs like "data:image/jpeg;base64,..."
	// which can be 15KB-96KB for typical PDF images, but Milvus VARCHAR
	// limit is 2048 chars. Store only URL reference in Image field.
	// The full base64 data is stored in ImageData JSON field (64KB limit).
	const milvusImageVarCharLimit = 2048
	if len(mdoc.Image) > milvusImageVarCharLimit {
		// Store URL reference instead of full data URL
		mdoc.Image = mdoc.URL

		// If URL is also too long (edge case), truncate it
		if len(mdoc.Image) > milvusImageVarCharLimit {
			mdoc.Image = mdoc.Image[:milvusImageVarCharLimit]
		}
	}

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

	// If no embedding was generated, create zero vector with correct dimensions
	if mdoc.Embedding == nil {
		mdoc.Embedding = make([]float32, a.config.VectorDimensions)
	}

	// Prepare image data as JSON (supports large base64 strings unlike VARCHAR)
	// NOTE: Milvus JSON fields have a 65,536 byte (64KB) limit
	// For images larger than this, we store empty ImageData and rely on Image URL
	imageDataJSON := map[string]string{"data": mdoc.ImageData}
	imageDataBytes := mustMarshalJSON(imageDataJSON)

	// Check if image data exceeds Milvus JSON limit (64KB)
	const milvusJSONLimit = 65536
	if len(imageDataBytes) > milvusJSONLimit {
		// Store empty image data and log warning
		imageDataJSON = map[string]string{"data": ""}
		imageDataBytes = mustMarshalJSON(imageDataJSON)
		fmt.Fprintf(os.Stderr, "Warning: Image data for document %s exceeds Milvus JSON limit (%d bytes > %d bytes). Storing URL only.\n",
			mdoc.DocumentID, len(imageDataBytes), milvusJSONLimit)
	}

	// Prepare column data for insertion
	columns := []entity.Column{
		entity.NewColumnVarChar(FieldDocumentID, []string{mdoc.DocumentID}),
		entity.NewColumnVarChar(FieldText, []string{mdoc.Text}),
		entity.NewColumnVarChar(FieldContent, []string{mdoc.Content}),
		entity.NewColumnVarChar(FieldImage, []string{mdoc.Image}),
		entity.NewColumnJSONBytes(FieldImageData, [][]byte{imageDataBytes}),
		entity.NewColumnVarChar(FieldURL, []string{mdoc.URL}),
		entity.NewColumnFloatVector(FieldEmbedding, a.config.VectorDimensions, [][]float32{mdoc.Embedding}),
		entity.NewColumnJSONBytes(FieldMetadata, [][]byte{mustMarshalJSON(mdoc.Metadata)}),
		entity.NewColumnInt64(FieldCreatedAt, []int64{mdoc.CreatedAt}),
		entity.NewColumnInt64(FieldUpdatedAt, []int64{mdoc.UpdatedAt}),
	}

	// Insert data
	_, err := a.client.Insert(ctx, collectionName, "", columns...)
	if err != nil {
		return fmt.Errorf("Milvus: failed to create document: %w", err)
	}

	// Flush to make data available for search
	err = a.client.Flush(ctx, collectionName, false)
	if err != nil {
		return fmt.Errorf("Milvus: failed to flush collection: %w", err)
	}

	return nil
}

// CreateDocuments creates multiple documents in batch
func (a *Adapter) CreateDocuments(ctx context.Context, collectionName string, documents []*vectordb.Document) error {
	ctx, cancel := context.WithTimeout(ctx, a.getTimeoutFor(vectordb.OperationTypeBulk))
	defer cancel()

	if len(documents) == 0 {
		return nil
	}

	// Prepare data arrays for batch insertion
	documentIDs := make([]string, len(documents))
	texts := make([]string, len(documents))
	contents := make([]string, len(documents))
	images := make([]string, len(documents))
	imageDatas := make([][]byte, len(documents)) // JSON field for large base64 images
	urls := make([]string, len(documents))
	embeddings := make([][]float32, len(documents))
	metadatas := make([][]byte, len(documents))
	createdAts := make([]int64, len(documents))
	updatedAts := make([]int64, len(documents))

	for i, doc := range documents {
		mdoc := a.toMilvusDocument(doc)

		// CRITICAL FIX (v0.9.7): Truncate Image field if it exceeds VARCHAR limit
		const milvusImageVarCharLimit = 2048
		if len(mdoc.Image) > milvusImageVarCharLimit {
			// Store URL reference instead of full data URL
			mdoc.Image = mdoc.URL

			// If URL is also too long (edge case), truncate it
			if len(mdoc.Image) > milvusImageVarCharLimit {
				mdoc.Image = mdoc.Image[:milvusImageVarCharLimit]
			}
		}

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

		// If no embedding was generated, create zero vector with correct dimensions
		if mdoc.Embedding == nil {
			mdoc.Embedding = make([]float32, a.config.VectorDimensions)
		}

		documentIDs[i] = mdoc.DocumentID
		texts[i] = mdoc.Text
		contents[i] = mdoc.Content
		images[i] = mdoc.Image

		// Check if image data exceeds Milvus JSON limit (64KB)
		imageDataJSON := map[string]string{"data": mdoc.ImageData}
		imageDataBytes := mustMarshalJSON(imageDataJSON)
		const milvusJSONLimit = 65536
		if len(imageDataBytes) > milvusJSONLimit {
			// Store empty image data and log warning
			imageDataJSON = map[string]string{"data": ""}
			imageDataBytes = mustMarshalJSON(imageDataJSON)
			fmt.Fprintf(os.Stderr, "Warning: Image data for document %s exceeds Milvus JSON limit (%d bytes > %d bytes). Storing URL only.\n",
				mdoc.DocumentID, len(imageDataBytes), milvusJSONLimit)
		}
		imageDatas[i] = imageDataBytes

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
		entity.NewColumnJSONBytes(FieldImageData, imageDatas), // JSON field supports large base64 images
		entity.NewColumnVarChar(FieldURL, urls),
		entity.NewColumnFloatVector(FieldEmbedding, a.config.VectorDimensions, embeddings),
		entity.NewColumnJSONBytes(FieldMetadata, metadatas),
		entity.NewColumnInt64(FieldCreatedAt, createdAts),
		entity.NewColumnInt64(FieldUpdatedAt, updatedAts),
	}

	// Insert data
	_, err := a.client.Insert(ctx, collectionName, "", columns...)
	if err != nil {
		return fmt.Errorf("Milvus: failed to create documents: %w", err)
	}

	// Flush to make data available for search
	err = a.client.Flush(ctx, collectionName, false)
	if err != nil {
		return fmt.Errorf("Milvus: failed to flush collection: %w", err)
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
		return nil, fmt.Errorf("Milvus: failed to get document: %w", err)
	}

	if result.Len() == 0 {
		return nil, fmt.Errorf("Milvus: document not found: %s", documentID)
	}

	// Extract fields from result
	docID := result.GetColumn(FieldDocumentID).(*entity.ColumnVarChar).Data()[0]
	text := result.GetColumn(FieldText).(*entity.ColumnVarChar).Data()[0]
	content := result.GetColumn(FieldContent).(*entity.ColumnVarChar).Data()[0]
	image := result.GetColumn(FieldImage).(*entity.ColumnVarChar).Data()[0]

	// Extract image data from JSON field
	var imageData string
	if imageDataCol := result.GetColumn(FieldImageData); imageDataCol != nil {
		imageDataBytes := imageDataCol.(*entity.ColumnJSONBytes).Data()[0]
		imageDataMap := mustUnmarshalJSON(imageDataBytes)
		if data, ok := imageDataMap["data"].(string); ok {
			imageData = data
		}
	}

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
		return fmt.Errorf("Milvus: failed to delete document: %w", err)
	}

	// Flush to ensure deletion is applied
	err = c.client.Flush(ctx, collectionName, false)
	if err != nil {
		return fmt.Errorf("Milvus: failed to flush after deletion: %w", err)
	}

	return nil
}

// DeleteDocuments deletes multiple documents by IDs
func (c *Client) DeleteDocuments(ctx context.Context, collectionName string, documentIDs []string) error {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeoutFor(vectordb.OperationTypeBulk))
	defer cancel()

	if len(documentIDs) == 0 {
		return nil
	}

	// Delete by primary keys
	err := c.client.DeleteByPks(ctx, collectionName, "", entity.NewColumnVarChar(FieldDocumentID, documentIDs))
	if err != nil {
		return fmt.Errorf("Milvus: failed to delete documents: %w", err)
	}

	// Flush to ensure deletions are applied
	err = c.client.Flush(ctx, collectionName, false)
	if err != nil {
		return fmt.Errorf("Milvus: failed to flush after deletion: %w", err)
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
		return fmt.Errorf("Milvus: no metadata criteria specified")
	}

	// Query to find matching documents
	outputFields := []string{FieldDocumentID}
	result, err := c.client.Query(ctx, collectionName, nil, expr, outputFields)
	if err != nil {
		return fmt.Errorf("Milvus: failed to query documents by metadata: %w", err)
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
		return fmt.Errorf("Milvus: failed to delete documents: %w", err)
	}

	// Flush to ensure deletions are applied
	err = c.client.Flush(ctx, collectionName, false)
	if err != nil {
		return fmt.Errorf("Milvus: failed to flush after deletion: %w", err)
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

	// Milvus requires limit when using empty expression
	// Use a large default if limit is 0
	queryLimit := limit
	if queryLimit <= 0 {
		queryLimit = 10000 // Default max limit for listing
	}

	// Use QueryByPks alternative or query with limit option
	// For empty expression, we need to use the QueryOption with limit
	result, err := c.client.Query(
		ctx,
		collectionName,
		nil, // partition names
		"",  // empty expression to get all
		outputFields,
		client.WithLimit(int64(queryLimit)),
		client.WithOffset(int64(offset)),
	)
	if err != nil {
		return nil, fmt.Errorf("Milvus: failed to list documents: %w", err)
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

		// Extract image data from JSON field
		var imageData string
		if imageDataCol := result.GetColumn(FieldImageData); imageDataCol != nil {
			imageDataBytes := imageDataCol.(*entity.ColumnJSONBytes).Data()[i]
			imageDataMap := mustUnmarshalJSON(imageDataBytes)
			if data, ok := imageDataMap["data"].(string); ok {
				imageData = data
			}
		}

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
	data, err := json.Marshal(v)
	if err != nil {
		return []byte("{}")
	}
	return data
}

// Helper function to unmarshal JSON metadata
func mustUnmarshalJSON(data []byte) map[string]interface{} {
	if len(data) == 0 {
		return make(map[string]interface{})
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return make(map[string]interface{})
	}
	return result
}
