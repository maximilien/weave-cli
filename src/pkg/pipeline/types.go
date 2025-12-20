package pipeline

import (
	"time"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// FileType represents the type of file
type FileType string

const (
	FileTypePDF     FileType = "pdf"
	FileTypeTXT     FileType = "txt"
	FileTypeMD      FileType = "md"
	FileTypeJSON    FileType = "json"
	FileTypeYAML    FileType = "yaml"
	FileTypeImage   FileType = "image"
	FileTypeUnknown FileType = "unknown"
)

// FileInfo represents a discovered file
type FileInfo struct {
	Path    string
	Size    int64
	ModTime time.Time
	Type    FileType
	Hash    string // SHA256 for deduplication
}

// IngestOptions configures the ingestion pipeline
type IngestOptions struct {
	// Source configuration
	Source    string   // File or directory path
	Glob      string   // Glob pattern (e.g., "**/*.pdf")
	Exclude   []string // Exclusion patterns
	Recursive bool     // Recursive directory scan (default: true)

	// Target configuration
	Collection string // Target collection name
	VDBType    string // Vector database type

	// Processing configuration
	BatchSize int               // Documents per batch (default: 100)
	Workers   int               // Concurrent workers (default: 4)
	Metadata  map[string]string // Additional metadata for all docs

	// Operation modes
	DryRun    bool   // Preview without executing
	Resume    bool   // Resume from previous state
	StateFile string // State file path for resume

	// Output configuration
	OutputFormat string // json, yaml, table
	ReportFile   string // Report output file
	Quiet        bool   // Suppress progress output
}

// ProcessingResult represents the result of processing a file
type ProcessingResult struct {
	File          FileInfo
	Documents     []*vectordb.Document
	DocumentCount int
	Error         error
	Duration      time.Duration
}

// ErrorDetail contains information about a processing error
type ErrorDetail struct {
	File      string    `json:"file"`
	Error     string    `json:"error"`
	Timestamp time.Time `json:"timestamp"`
}

// IngestReport is the final summary
type IngestReport struct {
	Status    string    `json:"status"` // success, partial, failure
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Duration  float64   `json:"duration_seconds"`

	// File statistics
	FilesScanned   int `json:"files_scanned"`
	FilesProcessed int `json:"files_processed"`
	FilesSkipped   int `json:"files_skipped"`
	FilesFailed    int `json:"files_failed"`

	// Document statistics
	DocumentsCreated int `json:"documents_created"`
	DocumentsFailed  int `json:"documents_failed"`

	// Performance metrics
	ThroughputFiles float64 `json:"throughput_files_per_sec"`
	ThroughputDocs  float64 `json:"throughput_docs_per_sec"`

	// Error details
	Errors []ErrorDetail `json:"errors,omitempty"`

	// Configuration
	Collection string `json:"collection"`
	VDBType    string `json:"vdb_type"`
	BatchSize  int    `json:"batch_size"`
	Workers    int    `json:"workers"`
}

// IngestState tracks processing state for resume capability
type IngestState struct {
	ProcessedFiles map[string]bool `json:"processed_files"`
	LastUpdated    time.Time       `json:"last_updated"`
}
