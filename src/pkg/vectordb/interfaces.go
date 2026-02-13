// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package vectordb

import (
	"context"
)

// Document represents a document in the vector database
type Document struct {
	ID        string                 `json:"id"`
	Text      string                 `json:"text"`
	Content   string                 `json:"content"`
	Image     string                 `json:"image"`
	ImageData string                 `json:"image_data"`
	URL       string                 `json:"url"`
	Metadata  map[string]interface{} `json:"metadata"`
	Embedding []float64              `json:"embedding,omitempty"` // Pre-generated embedding (optional)

	// External storage fields (for images exceeding VDB limits, e.g., Milvus 65KB)
	ImageThumbnail string                 `json:"image_thumbnail,omitempty"` // Small thumbnail (base64, <47KB for Milvus)
	ImageURL       string                 `json:"image_url,omitempty"`       // Full-resolution image URL (S3/MinIO/etc)
	ImageMetadata  map[string]interface{} `json:"image_metadata,omitempty"`  // Image metadata (size, format, dimensions)
}

// QueryOptions represents options for querying the vector database
type QueryOptions struct {
	TopK           int     `json:"top_k"`
	Distance       float64 `json:"distance"`
	SearchMetadata bool    `json:"search_metadata"`
	NoTruncate     bool    `json:"no_truncate"`
	UseBM25        bool    `json:"use_bm25"`
	IncludeImages  bool    `json:"include_images"` // Include base64 image data in results
}

// QueryResult represents a search result with score
type QueryResult struct {
	Document Document `json:"document"`
	Score    float64  `json:"score"`
}

// FieldDefinition represents a field in a collection schema
type FieldDefinition struct {
	Name string
	Type string
}

// SchemaProperty represents a property in a collection schema
type SchemaProperty struct {
	Name             string                 `json:"name" yaml:"name"`
	DataType         []string               `json:"dataType" yaml:"datatype"`
	Description      string                 `json:"description,omitempty" yaml:"description,omitempty"`
	NestedProperties []SchemaProperty       `json:"nestedProperties,omitempty" yaml:"nestedproperties,omitempty"`
	JSONSchema       map[string]interface{} `json:"json_schema,omitempty" yaml:"json_schema,omitempty"`
}

// CollectionSchema represents a collection schema
type CollectionSchema struct {
	Class      string           `json:"class"`
	Vectorizer string           `json:"vectorizer,omitempty"`
	Properties []SchemaProperty `json:"properties"`
}

// SchemaType represents the type of collection schema
type SchemaType string

const (
	SchemaTypeText  SchemaType = "text"
	SchemaTypeImage SchemaType = "image"
)

// CollectionInfo represents information about a collection
type CollectionInfo struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Count       int64  `json:"count"`
	Vectorizer  string `json:"vectorizer,omitempty"`
}

// VectorDBClient defines the interface that all vector database implementations must satisfy
type VectorDBClient interface {
	// Health checks the health of the vector database instance
	Health(ctx context.Context) error

	// Collection operations
	CollectionOperations

	// Document operations
	DocumentOperations

	// Query operations
	QueryOperations

	// Schema operations
	SchemaOperations
}

// CollectionOperations defines collection management operations
type CollectionOperations interface {
	// CreateCollection creates a new collection with the given schema
	CreateCollection(ctx context.Context, name string, schema *CollectionSchema) error

	// DeleteCollection deletes a collection and all its documents
	DeleteCollection(ctx context.Context, name string) error

	// ListCollections returns a list of all collections
	ListCollections(ctx context.Context) ([]CollectionInfo, error)

	// CollectionExists checks if a collection exists
	CollectionExists(ctx context.Context, name string) (bool, error)

	// GetCollectionCount returns the number of documents in a collection
	GetCollectionCount(ctx context.Context, name string) (int64, error)
}

// DocumentOperations defines document management operations
type DocumentOperations interface {
	// CreateDocument creates a new document in the specified collection
	CreateDocument(ctx context.Context, collectionName string, document *Document) error

	// CreateDocuments creates multiple documents in batch
	CreateDocuments(ctx context.Context, collectionName string, documents []*Document) error

	// GetDocument retrieves a document by ID
	GetDocument(ctx context.Context, collectionName, documentID string) (*Document, error)

	// UpdateDocument updates an existing document
	UpdateDocument(ctx context.Context, collectionName string, document *Document) error

	// DeleteDocument deletes a document by ID
	DeleteDocument(ctx context.Context, collectionName, documentID string) error

	// DeleteDocuments deletes multiple documents by IDs
	DeleteDocuments(ctx context.Context, collectionName string, documentIDs []string) error

	// DeleteDocumentsByMetadata deletes documents matching metadata criteria
	DeleteDocumentsByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}) error

	// ListDocuments returns a list of documents in a collection
	ListDocuments(ctx context.Context, collectionName string, limit int, offset int) ([]*Document, error)
}

// QueryOperations defines search and query operations
type QueryOperations interface {
	// SearchSemantic performs semantic search using vector embeddings
	SearchSemantic(ctx context.Context, collectionName, query string, options *QueryOptions) ([]*QueryResult, error)

	// SearchBM25 performs keyword-based search using BM25
	SearchBM25(ctx context.Context, collectionName, query string, options *QueryOptions) ([]*QueryResult, error)

	// SearchHybrid performs hybrid search combining vector and keyword search
	SearchHybrid(ctx context.Context, collectionName, query string, options *QueryOptions) ([]*QueryResult, error)

	// SearchByMetadata searches documents by metadata fields
	SearchByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}, options *QueryOptions) ([]*QueryResult, error)
}

// SchemaOperations defines schema management operations
type SchemaOperations interface {
	// GetSchema retrieves the schema for a collection
	GetSchema(ctx context.Context, collectionName string) (*CollectionSchema, error)

	// UpdateSchema updates the schema for a collection
	UpdateSchema(ctx context.Context, collectionName string, schema *CollectionSchema) error

	// GetDefaultSchema returns a default schema for the given type
	GetDefaultSchema(schemaType SchemaType, collectionName string) *CollectionSchema

	// ValidateSchema validates a schema definition
	ValidateSchema(schema *CollectionSchema) error
}
