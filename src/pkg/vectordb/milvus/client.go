// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package milvus

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// isCloudEndpoint checks if the address is a cloud endpoint that requires TLS
func isCloudEndpoint(address string) bool {
	cloudDomains := []string{
		"zilliz.com",
		"zillizcloud.com",
		"cloud.zilliz",
		"serverless",
	}
	lowerAddr := strings.ToLower(address)
	for _, domain := range cloudDomains {
		if strings.Contains(lowerAddr, domain) {
			return true
		}
	}
	return false
}

// Client wraps the Milvus client with vector database functionality
type Client struct {
	client client.Client
	config *Config
}

// Config holds Milvus client configuration
type Config struct {
	// Connection settings
	Address  string // Milvus server address (e.g., "localhost:19530")
	Username string // Username for authentication (optional)
	Password string // Password for authentication (optional)
	APIKey   string // API key/token for Zilliz Cloud (optional)
	Database string // Database name (default: "default")
	Timeout  int    // Timeout in seconds for operations (default: 10)

	// Vector search settings
	VectorDimensions int    // Embedding vector dimensions (e.g., 1536 for OpenAI)
	SimilarityMetric string // Similarity metric: "L2", "IP" (inner product), or "COSINE"
}

// NewClient creates a new Milvus client
func NewClient(config *Config) (*Client, error) {
	if config.Address == "" {
		return nil, fmt.Errorf("Milvus: Milvus address is required")
	}

	// Set defaults
	if config.Database == "" {
		config.Database = "default"
	}
	if config.VectorDimensions == 0 {
		config.VectorDimensions = 1536 // OpenAI text-embedding-3-small default
	}
	if config.SimilarityMetric == "" {
		config.SimilarityMetric = "L2" // Milvus default
	}
	if config.Timeout == 0 {
		config.Timeout = 10 // default 10 seconds
	}

	// Create context with timeout for connection
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Timeout)*time.Second)
	defer cancel()

	// Build connection config
	connConfig := client.Config{
		Address: config.Address,
	}

	// Enable TLS for cloud endpoints (Zilliz Cloud requires TLS)
	// Detect cloud by checking if address contains cloud provider domains
	isCloud := isCloudEndpoint(config.Address)
	if isCloud {
		connConfig.EnableTLSAuth = true
	}

	// Add authentication
	// For Zilliz Cloud, prefer APIKey/Token if available
	if config.APIKey != "" {
		// Direct API key/token (Zilliz serverless)
		connConfig.APIKey = config.APIKey
	} else if isCloud && config.Username != "" && config.Password != "" {
		// Zilliz Cloud can use APIKey or username/password
		// Try APIKey format first (more reliable for serverless)
		connConfig.APIKey = config.Username + ":" + config.Password
	} else if config.Username != "" && config.Password != "" {
		// Standard username/password for non-cloud or explicit config
		connConfig.Username = config.Username
		connConfig.Password = config.Password
	}

	// Set database if not default
	if config.Database != "default" {
		connConfig.DBName = config.Database
	}

	// Connect to Milvus
	c, err := client.NewClient(ctx, connConfig)
	if err != nil {
		return nil, fmt.Errorf("Milvus: failed to connect to Milvus: %w", err)
	}

	return &Client{
		client: c,
		config: config,
	}, nil
}

// getTimeout returns the timeout duration for operations
func (c *Client) getTimeout() time.Duration {
	timeout := c.config.Timeout
	if timeout == 0 {
		timeout = 10 // default 10 seconds
	}
	return time.Duration(timeout) * time.Second
}

// getTimeoutFor returns an operation-specific timeout based on deployment type
func (c *Client) getTimeoutFor(opType vectordb.OperationType) time.Duration {
	// Detect cloud: check if address contains cloud/zilliz domains
	isCloud := isCloudEndpoint(c.config.Address)
	return vectordb.GetTimeoutForOperation(opType, isCloud, c.config.Timeout)
}

// Health checks the health of the Milvus instance
func (c *Client) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeoutFor(vectordb.OperationTypeHealth))
	defer cancel()

	// Check if client is alive by listing databases
	_, err := c.client.ListDatabases(ctx)
	if err != nil {
		return fmt.Errorf("Milvus: Milvus health check failed: %w", err)
	}

	return nil
}

// Close closes the Milvus client connection
func (c *Client) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// milvusDocument represents the data structure for Milvus documents
// Milvus uses explicit schemas, so this is used for dynamic field mapping
type milvusDocument struct {
	DocumentID string                 // Primary key field
	Text       string                 // Text content
	Content    string                 // Full content
	Image      string                 // Image URL
	ImageData  string                 // Base64 encoded image data
	URL        string                 // Document URL
	Embedding  []float32              // Vector embedding (Milvus uses float32)
	Metadata   map[string]interface{} // Additional metadata as JSON
	CreatedAt  int64                  // Unix timestamp
	UpdatedAt  int64                  // Unix timestamp
}

// toMilvusDocument converts a vectordb.Document to a milvusDocument
func (c *Client) toMilvusDocument(doc *vectordb.Document) *milvusDocument {
	now := time.Now().Unix()

	return &milvusDocument{
		DocumentID: doc.ID,
		Text:       doc.Text,
		Content:    doc.Content,
		Image:      doc.Image,
		ImageData:  doc.ImageData,
		URL:        doc.URL,
		Embedding:  nil, // Embeddings will be populated by the adapter
		Metadata:   doc.Metadata,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// fromMilvusDocument converts Milvus result data to a vectordb.Document
func (c *Client) fromMilvusDocument(
	documentID string,
	text string,
	content string,
	image string,
	imageData string,
	url string,
	metadata map[string]interface{},
) *vectordb.Document {
	return &vectordb.Document{
		ID:        documentID,
		Text:      text,
		Content:   content,
		Image:     image,
		ImageData: imageData,
		URL:       url,
		Metadata:  metadata,
	}
}

// getMetricType converts similarity metric string to Milvus metric type
func (c *Client) getMetricType() entity.MetricType {
	switch c.config.SimilarityMetric {
	case "IP":
		return entity.IP // Inner product
	case "COSINE":
		return entity.COSINE // Cosine similarity
	case "L2":
		fallthrough
	default:
		return entity.L2 // Euclidean distance (default)
	}
}
