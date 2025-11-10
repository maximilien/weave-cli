// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package supabase

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/supabase-community/supabase-go"
)

// Adapter wraps the Supabase client to implement the VectorDBClient interface
type Adapter struct {
	client *supabase.Client
	db     *sql.DB
	config *vectordb.Config
}

// NewAdapter creates a new Supabase adapter
func NewAdapter(config *vectordb.Config) (*Adapter, error) {
	// Create Supabase client
	client, err := supabase.NewClient(config.DatabaseURL, config.DatabaseKey, nil)
	if err != nil {
		return nil, vectordb.ErrConnectionFailed("failed to create Supabase client", err)
	}

	// Create direct database connection for vector operations
	db, err := sql.Open("postgres", config.DatabaseURL)
	if err != nil {
		return nil, vectordb.ErrConnectionFailed("failed to connect to Supabase database", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, vectordb.ErrConnectionFailed("failed to ping Supabase database", err)
	}

	return &Adapter{
		client: client,
		db:     db,
		config: config,
	}, nil
}

// Health checks the health of the Supabase instance
func (a *Adapter) Health(ctx context.Context) error {
	if err := a.db.PingContext(ctx); err != nil {
		return vectordb.ErrConnectionFailed("Supabase health check failed", err)
	}
	return nil
}

// convertDocument converts vectordb.Document to a format suitable for Supabase
func (a *Adapter) convertDocument(doc *vectordb.Document) map[string]interface{} {
	if doc == nil {
		return nil
	}

	result := map[string]interface{}{
		"id":      doc.ID,
		"content": doc.Content,
	}

	// Add text field if present
	if doc.Text != "" {
		result["text"] = doc.Text
	}

	// Add image fields if present
	if doc.Image != "" {
		result["image"] = doc.Image
	}
	if doc.ImageData != "" {
		result["image_data"] = doc.ImageData
	}

	// Add URL if present
	if doc.URL != "" {
		result["url"] = doc.URL
	}

	// Add metadata as JSONB
	if doc.Metadata != nil {
		metadataJSON, _ := json.Marshal(doc.Metadata)
		result["metadata"] = string(metadataJSON)
	}

	return result
}

// convertDocumentFromSupabase converts Supabase result to vectordb.Document
func (a *Adapter) convertDocumentFromSupabase(row map[string]interface{}) *vectordb.Document {
	if row == nil {
		return nil
	}

	doc := &vectordb.Document{}

	// Extract ID
	if id, ok := row["id"].(string); ok {
		doc.ID = id
	}

	// Extract content
	if content, ok := row["content"].(string); ok {
		doc.Content = content
	}

	// Extract text
	if text, ok := row["text"].(string); ok {
		doc.Text = text
	}

	// Extract image fields
	if image, ok := row["image"].(string); ok {
		doc.Image = image
	}
	if imageData, ok := row["image_data"].(string); ok {
		doc.ImageData = imageData
	}

	// Extract URL
	if url, ok := row["url"].(string); ok {
		doc.URL = url
	}

	// Extract metadata
	if metadataStr, ok := row["metadata"].(string); ok && metadataStr != "" {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(metadataStr), &metadata); err == nil {
			doc.Metadata = metadata
		}
	}

	return doc
}

// wrapError wraps Supabase errors in vectordb errors
func (a *Adapter) wrapError(err error, operation string) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	// Connection errors
	if strings.Contains(errMsg, "connection refused") ||
		strings.Contains(errMsg, "no such host") ||
		strings.Contains(errMsg, "timeout") ||
		strings.Contains(errMsg, "connection reset") {
		return vectordb.ErrConnectionFailed(fmt.Sprintf("%s failed", operation), err)
	}

	// Authentication errors
	if strings.Contains(errMsg, "authentication failed") ||
		strings.Contains(errMsg, "invalid api key") ||
		strings.Contains(errMsg, "unauthorized") ||
		strings.Contains(errMsg, "permission denied") {
		return vectordb.ErrAuthenticationFailed(fmt.Sprintf("%s failed: %s", operation, errMsg))
	}

	// Not found errors
	if strings.Contains(errMsg, "not found") ||
		strings.Contains(errMsg, "does not exist") {
		return vectordb.ErrNotFound("resource", "unknown")
	}

	// Invalid configuration errors
	if strings.Contains(errMsg, "invalid") ||
		strings.Contains(errMsg, "malformed") {
		return vectordb.ErrInvalidConfig(fmt.Sprintf("%s failed: %s", operation, errMsg))
	}

	// Default to internal error
	return vectordb.ErrInternal(fmt.Sprintf("%s failed", operation), err)
}

// getTableName returns the table name for a collection
func (a *Adapter) getTableName(collectionName string) string {
	// Sanitize collection name to be a valid PostgreSQL table name
	tableName := strings.ToLower(collectionName)
	tableName = strings.ReplaceAll(tableName, "-", "_")
	tableName = strings.ReplaceAll(tableName, " ", "_")
	return fmt.Sprintf("collection_%s", tableName)
}

// ensureVectorExtension ensures that the pgvector extension is enabled
func (a *Adapter) ensureVectorExtension(ctx context.Context) error {
	query := "CREATE EXTENSION IF NOT EXISTS vector;"
	_, err := a.db.ExecContext(ctx, query)
	if err != nil {
		return a.wrapError(err, "ensure vector extension")
	}
	return nil
}

// convertMetadataToJSON converts metadata map to JSON string
func (a *Adapter) convertMetadataToJSON(metadata map[string]interface{}) (string, error) {
	if metadata == nil {
		return "{}", nil
	}

	jsonBytes, err := json.Marshal(metadata)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// convertJSONToMetadata converts JSON string to metadata map
func (a *Adapter) convertJSONToMetadata(jsonStr string) (map[string]interface{}, error) {
	if jsonStr == "" {
		return make(map[string]interface{}), nil
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &metadata); err != nil {
		return nil, err
	}

	return metadata, nil
}
