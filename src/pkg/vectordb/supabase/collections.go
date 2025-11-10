// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package supabase

import (
	"context"
	"fmt"
	"strings"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// CreateCollection creates a new collection with the given schema
func (a *Adapter) CreateCollection(ctx context.Context, name string, schema *vectordb.CollectionSchema) error {
	if err := a.ensureVectorExtension(ctx); err != nil {
		return err
	}

	tableName := a.getTableName(name)

	// Build CREATE TABLE query based on schema
	var columns []string
	columns = append(columns, "id TEXT PRIMARY KEY")
	columns = append(columns, "content TEXT")
	columns = append(columns, "text TEXT")
	columns = append(columns, "image TEXT")
	columns = append(columns, "image_data TEXT")
	columns = append(columns, "url TEXT")
	columns = append(columns, "metadata JSONB")
	columns = append(columns, "embedding vector(1536)") // Default OpenAI embedding dimension

	// Add custom properties from schema
	if schema != nil {
		for _, prop := range schema.Properties {
			columnType := a.convertDataTypeToPostgreSQL(prop.DataType)
			if columnType != "" {
				columns = append(columns, fmt.Sprintf("%s %s", prop.Name, columnType))
			}
		}
	}

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", tableName, strings.Join(columns, ", "))

	_, err := a.db.ExecContext(ctx, query)
	if err != nil {
		return a.wrapError(err, "create collection")
	}

	// Create indexes for better performance
	indexQueries := []string{
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s_embedding_idx ON %s USING ivfflat (embedding vector_cosine_ops)", tableName, tableName),
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s_metadata_idx ON %s USING gin (metadata)", tableName, tableName),
	}

	for _, indexQuery := range indexQueries {
		_, err := a.db.ExecContext(ctx, indexQuery)
		if err != nil {
			// Log warning but don't fail - indexes are optional
			continue
		}
	}

	return nil
}

// DeleteCollection deletes a collection and all its documents
func (a *Adapter) DeleteCollection(ctx context.Context, name string) error {
	tableName := a.getTableName(name)
	query := fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)

	_, err := a.db.ExecContext(ctx, query)
	if err != nil {
		return a.wrapError(err, "delete collection")
	}

	return nil
}

// ListCollections returns a list of all collections
func (a *Adapter) ListCollections(ctx context.Context) ([]vectordb.CollectionInfo, error) {
	query := `
		SELECT table_name, 
		       COALESCE(obj_description(c.oid), '') as description,
		       COALESCE(n_tup_ins, 0) as count
		FROM information_schema.tables t
		LEFT JOIN pg_class c ON c.relname = t.table_name
		LEFT JOIN pg_stat_user_tables s ON s.relname = t.table_name
		WHERE t.table_schema = 'public' 
		  AND t.table_name LIKE 'collection_%'
		  AND t.table_type = 'BASE TABLE'
	`

	rows, err := a.db.QueryContext(ctx, query)
	if err != nil {
		return nil, a.wrapError(err, "list collections")
	}
	defer rows.Close()

	var collections []vectordb.CollectionInfo
	for rows.Next() {
		var tableName, description string
		var count int64

		if err := rows.Scan(&tableName, &description, &count); err != nil {
			continue
		}

		// Convert table name back to collection name
		collectionName := strings.TrimPrefix(tableName, "collection_")
		collectionName = strings.ReplaceAll(collectionName, "_", "-")

		collections = append(collections, vectordb.CollectionInfo{
			Name:        collectionName,
			Description: description,
			Count:       count,
		})
	}

	return collections, nil
}

// CollectionExists checks if a collection exists
func (a *Adapter) CollectionExists(ctx context.Context, name string) (bool, error) {
	tableName := a.getTableName(name)
	query := `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_schema = 'public' 
			  AND table_name = $1
		)
	`

	var exists bool
	err := a.db.QueryRowContext(ctx, query, tableName).Scan(&exists)
	if err != nil {
		return false, a.wrapError(err, "check collection exists")
	}

	return exists, nil
}

// GetCollectionCount returns the number of documents in a collection
func (a *Adapter) GetCollectionCount(ctx context.Context, name string) (int64, error) {
	exists, err := a.CollectionExists(ctx, name)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, vectordb.ErrNotFound("collection", name)
	}

	tableName := a.getTableName(name)
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)

	var count int64
	err = a.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, a.wrapError(err, "get collection count")
	}

	return count, nil
}

// convertDataTypeToPostgreSQL converts vectordb data types to PostgreSQL types
func (a *Adapter) convertDataTypeToPostgreSQL(dataTypes []string) string {
	if len(dataTypes) == 0 {
		return ""
	}

	dataType := strings.ToLower(dataTypes[0])
	switch dataType {
	case "text", "string":
		return "TEXT"
	case "int", "integer":
		return "INTEGER"
	case "float", "number":
		return "REAL"
	case "boolean", "bool":
		return "BOOLEAN"
	case "date", "datetime":
		return "TIMESTAMP"
	case "uuid":
		return "UUID"
	case "json", "object":
		return "JSONB"
	default:
		return "TEXT" // Default to TEXT for unknown types
	}
}
