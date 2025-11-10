// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package supabase

import (
	"context"
	"fmt"
	"strings"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// GetSchema retrieves the schema for a collection
func (a *Adapter) GetSchema(ctx context.Context, collectionName string) (*vectordb.CollectionSchema, error) {
	exists, err := a.CollectionExists(ctx, collectionName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, vectordb.ErrNotFound("collection", collectionName)
	}

	tableName := a.getTableName(collectionName)

	// Query PostgreSQL information_schema to get column information
	query := `
		SELECT column_name, data_type, is_nullable, column_default
		FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = $1
		ORDER BY ordinal_position
	`

	rows, err := a.db.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, a.wrapError(err, "get schema")
	}
	defer rows.Close()

	var properties []vectordb.SchemaProperty
	for rows.Next() {
		var columnName, dataType, isNullable string
		var columnDefault *string

		if err := rows.Scan(&columnName, &dataType, &isNullable, &columnDefault); err != nil {
			continue
		}

		// Skip standard columns
		if columnName == "id" || columnName == "content" || columnName == "text" ||
			columnName == "image" || columnName == "image_data" || columnName == "url" ||
			columnName == "metadata" || columnName == "embedding" {
			continue
		}

		// Convert PostgreSQL data type to vectordb data type
		vectordbDataType := a.convertPostgreSQLToDataType(dataType)

		properties = append(properties, vectordb.SchemaProperty{
			Name:     columnName,
			DataType: []string{vectordbDataType},
		})
	}

	return &vectordb.CollectionSchema{
		Class:      collectionName,
		Vectorizer: "text2vec-openai", // Default vectorizer
		Properties: properties,
	}, nil
}

// UpdateSchema updates the schema for a collection
func (a *Adapter) UpdateSchema(ctx context.Context, collectionName string, schema *vectordb.CollectionSchema) error {
	exists, err := a.CollectionExists(ctx, collectionName)
	if err != nil {
		return err
	}
	if !exists {
		return vectordb.ErrNotFound("collection", collectionName)
	}

	if schema == nil {
		return vectordb.ErrInvalidConfig("schema cannot be nil")
	}

	tableName := a.getTableName(collectionName)

	// Get current schema
	currentSchema, err := a.GetSchema(ctx, collectionName)
	if err != nil {
		return err
	}

	// Find new properties to add
	currentProps := make(map[string]bool)
	for _, prop := range currentSchema.Properties {
		currentProps[prop.Name] = true
	}

	// Add new columns
	for _, prop := range schema.Properties {
		if !currentProps[prop.Name] {
			columnType := a.convertDataTypeToPostgreSQL(prop.DataType)
			if columnType != "" {
				alterQuery := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableName, prop.Name, columnType)
				_, err := a.db.ExecContext(ctx, alterQuery)
				if err != nil {
					return a.wrapError(err, "add column")
				}
			}
		}
	}

	return nil
}

// GetDefaultSchema returns a default schema for the given type
func (a *Adapter) GetDefaultSchema(schemaType vectordb.SchemaType, collectionName string) *vectordb.CollectionSchema {
	switch schemaType {
	case vectordb.SchemaTypeText:
		return &vectordb.CollectionSchema{
			Class:      collectionName,
			Vectorizer: "text2vec-openai",
			Properties: []vectordb.SchemaProperty{
				{
					Name:     "content",
					DataType: []string{"text"},
				},
				{
					Name:     "title",
					DataType: []string{"text"},
				},
				{
					Name:     "author",
					DataType: []string{"text"},
				},
				{
					Name:     "created_at",
					DataType: []string{"date"},
				},
			},
		}
	case vectordb.SchemaTypeImage:
		return &vectordb.CollectionSchema{
			Class:      collectionName,
			Vectorizer: "img2vec-neural",
			Properties: []vectordb.SchemaProperty{
				{
					Name:     "image",
					DataType: []string{"text"},
				},
				{
					Name:     "image_data",
					DataType: []string{"text"},
				},
				{
					Name:     "description",
					DataType: []string{"text"},
				},
				{
					Name:     "tags",
					DataType: []string{"text"},
				},
				{
					Name:     "created_at",
					DataType: []string{"date"},
				},
			},
		}
	default:
		return &vectordb.CollectionSchema{
			Class:      collectionName,
			Vectorizer: "text2vec-openai",
			Properties: []vectordb.SchemaProperty{
				{
					Name:     "content",
					DataType: []string{"text"},
				},
			},
		}
	}
}

// ValidateSchema validates a schema definition
func (a *Adapter) ValidateSchema(schema *vectordb.CollectionSchema) error {
	if schema == nil {
		return vectordb.ErrInvalidConfig("schema cannot be nil")
	}

	if schema.Class == "" {
		return vectordb.ErrInvalidConfig("schema class name cannot be empty")
	}

	// Validate class name (must be valid PostgreSQL table name)
	if !a.isValidTableName(schema.Class) {
		return vectordb.ErrInvalidConfig("invalid class name: must be a valid PostgreSQL table name")
	}

	// Validate properties
	if len(schema.Properties) == 0 {
		return vectordb.ErrInvalidConfig("schema must have at least one property")
	}

	propertyNames := make(map[string]bool)
	for _, prop := range schema.Properties {
		// Check for duplicate property names
		if propertyNames[prop.Name] {
			return vectordb.ErrInvalidConfig(fmt.Sprintf("duplicate property name: %s", prop.Name))
		}
		propertyNames[prop.Name] = true

		// Validate property name
		if prop.Name == "" {
			return vectordb.ErrInvalidConfig("property name cannot be empty")
		}

		if !a.isValidColumnName(prop.Name) {
			return vectordb.ErrInvalidConfig(fmt.Sprintf("invalid property name: %s", prop.Name))
		}

		// Validate data type
		if len(prop.DataType) == 0 {
			return vectordb.ErrInvalidConfig(fmt.Sprintf("property %s must have at least one data type", prop.Name))
		}

		for _, dataType := range prop.DataType {
			if !a.isValidDataType(dataType) {
				return vectordb.ErrInvalidConfig(fmt.Sprintf("invalid data type: %s", dataType))
			}
		}
	}

	return nil
}

// convertPostgreSQLToDataType converts PostgreSQL data types to vectordb data types
func (a *Adapter) convertPostgreSQLToDataType(pgType string) string {
	pgType = strings.ToLower(pgType)
	switch {
	case strings.Contains(pgType, "text") || strings.Contains(pgType, "varchar") || strings.Contains(pgType, "char"):
		return "text"
	case strings.Contains(pgType, "integer") || strings.Contains(pgType, "int"):
		return "int"
	case strings.Contains(pgType, "real") || strings.Contains(pgType, "float") || strings.Contains(pgType, "double"):
		return "float"
	case strings.Contains(pgType, "boolean") || strings.Contains(pgType, "bool"):
		return "boolean"
	case strings.Contains(pgType, "timestamp") || strings.Contains(pgType, "date"):
		return "date"
	case strings.Contains(pgType, "uuid"):
		return "uuid"
	case strings.Contains(pgType, "json"):
		return "json"
	default:
		return "text"
	}
}

// isValidTableName checks if a string is a valid PostgreSQL table name
func (a *Adapter) isValidTableName(name string) bool {
	if name == "" || len(name) > 63 {
		return false
	}

	// Must start with letter or underscore
	if !((name[0] >= 'a' && name[0] <= 'z') || (name[0] >= 'A' && name[0] <= 'Z') || name[0] == '_') {
		return false
	}

	// Can contain letters, digits, underscores
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_') {
			return false
		}
	}

	return true
}

// isValidColumnName checks if a string is a valid PostgreSQL column name
func (a *Adapter) isValidColumnName(name string) bool {
	return a.isValidTableName(name) // Same rules as table names
}

// isValidDataType checks if a data type is supported
func (a *Adapter) isValidDataType(dataType string) bool {
	validTypes := map[string]bool{
		"text":     true,
		"string":   true,
		"int":      true,
		"integer":  true,
		"float":    true,
		"number":   true,
		"boolean":  true,
		"bool":     true,
		"date":     true,
		"datetime": true,
		"uuid":     true,
		"json":     true,
		"object":   true,
	}

	return validTypes[strings.ToLower(dataType)]
}
