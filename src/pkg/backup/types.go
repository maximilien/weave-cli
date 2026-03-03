// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package backup

import (
	"time"
)

// BackupFormat represents the structure of a .weavebak backup file
type BackupFormat struct {
	Version   string           `json:"version"`
	Metadata  BackupMetadata   `json:"metadata"`
	Schema    BackupSchema     `json:"schema,omitempty"`
	Documents []BackupDocument `json:"documents"`
}

// BackupMetadata contains metadata about the backup
type BackupMetadata struct {
	Collection       string    `json:"collection"`
	VDBType          string    `json:"vdb_type"`
	EmbeddingModel   string    `json:"embedding_model"`
	VectorDimensions int       `json:"vector_dimensions"`
	CreatedAt        time.Time `json:"created_at"`
	WeaveVersion     string    `json:"weave_version"`
	TotalDocuments   int       `json:"total_documents"`
	BackupSizeBytes  int64     `json:"backup_size_bytes"`
}

// BackupSchema contains the collection schema
type BackupSchema struct {
	VectorDimensions int           `json:"vector_dimensions"`
	SimilarityMetric string        `json:"similarity_metric,omitempty"`
	Fields           []SchemaField `json:"fields,omitempty"`
}

// SchemaField represents a field in the collection schema
type SchemaField struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// BackupDocument represents a single document in the backup
type BackupDocument struct {
	ID        string                 `json:"id"`
	Content   string                 `json:"content,omitempty"`
	Text      string                 `json:"text,omitempty"`
	Embedding []float64              `json:"embedding"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`

	// Image-specific fields
	Image     string `json:"image,omitempty"`
	ImageData string `json:"image_data,omitempty"`
	URL       string `json:"url,omitempty"`

	// External storage fields (v0.10.0+)
	ImageThumbnail string                 `json:"image_thumbnail,omitempty"`
	ImageURL       string                 `json:"image_url,omitempty"`
	ImageMetadata  map[string]interface{} `json:"image_metadata,omitempty"`
}

// CreateOptions holds options for backup creation
type CreateOptions struct {
	Collection string
	OutputFile string
	VDBType    string
	Compress   bool
	BatchSize  int
	Quiet      bool
}

// RestoreOptions holds options for backup restoration
type RestoreOptions struct {
	BackupFile string
	Collection string // Optional: override collection name
	VDBType    string
	Overwrite  bool
	Quiet      bool
}

// ValidationResult holds the result of backup validation
type ValidationResult struct {
	Valid           bool
	Version         string
	Collection      string
	TotalDocuments  int
	BackupSizeBytes int64
	Errors          []string
	Warnings        []string
}

// BackupInfo holds information about a backup file
type BackupInfo struct {
	FilePath        string
	FileName        string
	Collection      string
	VDBType         string
	TotalDocuments  int
	BackupSizeBytes int64
	CreatedAt       time.Time
	WeaveVersion    string
	Compressed      bool
}
