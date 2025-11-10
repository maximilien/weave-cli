// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package supabase

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// CreateDocument creates a new document in the specified collection
func (a *Adapter) CreateDocument(ctx context.Context, collectionName string, document *vectordb.Document) error {
	exists, err := a.CollectionExists(ctx, collectionName)
	if err != nil {
		return err
	}
	if !exists {
		return vectordb.ErrNotFound("collection", collectionName)
	}

	tableName := a.getTableName(collectionName)

	// Convert metadata to JSON
	metadataJSON, err := a.convertMetadataToJSON(document.Metadata)
	if err != nil {
		return a.wrapError(err, "convert metadata to JSON")
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (id, content, text, image, image_data, url, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			content = EXCLUDED.content,
			text = EXCLUDED.text,
			image = EXCLUDED.image,
			image_data = EXCLUDED.image_data,
			url = EXCLUDED.url,
			metadata = EXCLUDED.metadata
	`, tableName)

	_, err = a.db.ExecContext(ctx, query,
		document.ID,
		document.Content,
		document.Text,
		document.Image,
		document.ImageData,
		document.URL,
		metadataJSON,
	)

	if err != nil {
		return a.wrapError(err, "create document")
	}

	return nil
}

// CreateDocuments creates multiple documents in batch
func (a *Adapter) CreateDocuments(ctx context.Context, collectionName string, documents []*vectordb.Document) error {
	if len(documents) == 0 {
		return nil
	}

	exists, err := a.CollectionExists(ctx, collectionName)
	if err != nil {
		return err
	}
	if !exists {
		return vectordb.ErrNotFound("collection", collectionName)
	}

	tableName := a.getTableName(collectionName)

	// Use a transaction for batch insert
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return a.wrapError(err, "begin transaction")
	}
	defer func() {
		_ = tx.Rollback() // Ignore rollback error (expected if commit succeeded)
	}()

	query := fmt.Sprintf(`
		INSERT INTO %s (id, content, text, image, image_data, url, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			content = EXCLUDED.content,
			text = EXCLUDED.text,
			image = EXCLUDED.image,
			image_data = EXCLUDED.image_data,
			url = EXCLUDED.url,
			metadata = EXCLUDED.metadata
	`, tableName)

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return a.wrapError(err, "prepare statement")
	}
	defer stmt.Close()

	for _, document := range documents {
		metadataJSON, err := a.convertMetadataToJSON(document.Metadata)
		if err != nil {
			return a.wrapError(err, "convert metadata to JSON")
		}

		_, err = stmt.ExecContext(ctx,
			document.ID,
			document.Content,
			document.Text,
			document.Image,
			document.ImageData,
			document.URL,
			metadataJSON,
		)
		if err != nil {
			return a.wrapError(err, "execute batch insert")
		}
	}

	if err := tx.Commit(); err != nil {
		return a.wrapError(err, "commit transaction")
	}

	return nil
}

// GetDocument retrieves a document by ID
func (a *Adapter) GetDocument(ctx context.Context, collectionName, documentID string) (*vectordb.Document, error) {
	exists, err := a.CollectionExists(ctx, collectionName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, vectordb.ErrNotFound("collection", collectionName)
	}

	tableName := a.getTableName(collectionName)
	query := fmt.Sprintf(`
		SELECT id, content, text, image, image_data, url, metadata
		FROM %s WHERE id = $1
	`, tableName)

	row := a.db.QueryRowContext(ctx, query, documentID)

	var id, content, text, image, imageData, url, metadataJSON string
	err = row.Scan(&id, &content, &text, &image, &imageData, &url, &metadataJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, vectordb.ErrNotFound("document", documentID)
		}
		return nil, a.wrapError(err, "get document")
	}

	metadata, err := a.convertJSONToMetadata(metadataJSON)
	if err != nil {
		return nil, a.wrapError(err, "convert JSON to metadata")
	}

	return &vectordb.Document{
		ID:        id,
		Content:   content,
		Text:      text,
		Image:     image,
		ImageData: imageData,
		URL:       url,
		Metadata:  metadata,
	}, nil
}

// UpdateDocument updates an existing document
func (a *Adapter) UpdateDocument(ctx context.Context, collectionName string, document *vectordb.Document) error {
	exists, err := a.CollectionExists(ctx, collectionName)
	if err != nil {
		return err
	}
	if !exists {
		return vectordb.ErrNotFound("collection", collectionName)
	}

	tableName := a.getTableName(collectionName)

	metadataJSON, err := a.convertMetadataToJSON(document.Metadata)
	if err != nil {
		return a.wrapError(err, "convert metadata to JSON")
	}

	query := fmt.Sprintf(`
		UPDATE %s SET
			content = $2,
			text = $3,
			image = $4,
			image_data = $5,
			url = $6,
			metadata = $7
		WHERE id = $1
	`, tableName)

	result, err := a.db.ExecContext(ctx, query,
		document.ID,
		document.Content,
		document.Text,
		document.Image,
		document.ImageData,
		document.URL,
		metadataJSON,
	)

	if err != nil {
		return a.wrapError(err, "update document")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return a.wrapError(err, "get rows affected")
	}

	if rowsAffected == 0 {
		return vectordb.ErrNotFound("document", document.ID)
	}

	return nil
}

// DeleteDocument deletes a document by ID
func (a *Adapter) DeleteDocument(ctx context.Context, collectionName, documentID string) error {
	exists, err := a.CollectionExists(ctx, collectionName)
	if err != nil {
		return err
	}
	if !exists {
		return vectordb.ErrNotFound("collection", collectionName)
	}

	tableName := a.getTableName(collectionName)
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", tableName)

	result, err := a.db.ExecContext(ctx, query, documentID)
	if err != nil {
		return a.wrapError(err, "delete document")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return a.wrapError(err, "get rows affected")
	}

	if rowsAffected == 0 {
		return vectordb.ErrNotFound("document", documentID)
	}

	return nil
}

// DeleteDocuments deletes multiple documents by IDs
func (a *Adapter) DeleteDocuments(ctx context.Context, collectionName string, documentIDs []string) error {
	if len(documentIDs) == 0 {
		return nil
	}

	exists, err := a.CollectionExists(ctx, collectionName)
	if err != nil {
		return err
	}
	if !exists {
		return vectordb.ErrNotFound("collection", collectionName)
	}

	tableName := a.getTableName(collectionName)

	// Build placeholders for IN clause
	placeholders := make([]string, len(documentIDs))
	args := make([]interface{}, len(documentIDs))
	for i, id := range documentIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE id IN (%s)", tableName, strings.Join(placeholders, ", "))

	_, err = a.db.ExecContext(ctx, query, args...)
	if err != nil {
		return a.wrapError(err, "delete documents")
	}

	return nil
}

// DeleteDocumentsByMetadata deletes documents matching metadata criteria
func (a *Adapter) DeleteDocumentsByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}) error {
	exists, err := a.CollectionExists(ctx, collectionName)
	if err != nil {
		return err
	}
	if !exists {
		return vectordb.ErrNotFound("collection", collectionName)
	}

	tableName := a.getTableName(collectionName)

	// Build WHERE clause for metadata matching
	var conditions []string
	var args []interface{}
	argIndex := 1

	for key, value := range metadata {
		conditions = append(conditions, fmt.Sprintf("metadata->>'%s' = $%d", key, argIndex))
		args = append(args, fmt.Sprintf("%v", value))
		argIndex++
	}

	if len(conditions) == 0 {
		return vectordb.ErrInvalidConfig("no metadata criteria provided")
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE %s", tableName, strings.Join(conditions, " AND "))

	_, err = a.db.ExecContext(ctx, query, args...)
	if err != nil {
		return a.wrapError(err, "delete documents by metadata")
	}

	return nil
}

// ListDocuments returns a list of documents in a collection
func (a *Adapter) ListDocuments(ctx context.Context, collectionName string, limit int, offset int) ([]*vectordb.Document, error) {
	exists, err := a.CollectionExists(ctx, collectionName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, vectordb.ErrNotFound("collection", collectionName)
	}

	tableName := a.getTableName(collectionName)
	query := fmt.Sprintf(`
		SELECT id, content, text, image, image_data, url, metadata
		FROM %s
		ORDER BY id
		LIMIT $1 OFFSET $2
	`, tableName)

	rows, err := a.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, a.wrapError(err, "list documents")
	}
	defer rows.Close()

	var documents []*vectordb.Document
	for rows.Next() {
		var id, content, text, image, imageData, url, metadataJSON string
		if err := rows.Scan(&id, &content, &text, &image, &imageData, &url, &metadataJSON); err != nil {
			continue
		}

		metadata, err := a.convertJSONToMetadata(metadataJSON)
		if err != nil {
			continue
		}

		documents = append(documents, &vectordb.Document{
			ID:        id,
			Content:   content,
			Text:      text,
			Image:     image,
			ImageData: imageData,
			URL:       url,
			Metadata:  metadata,
		})
	}

	return documents, nil
}
