// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package mongodb

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// Client wraps the MongoDB client with vector database functionality
type Client struct {
	client   *mongo.Client
	database *mongo.Database
	config   *Config
}

// Config holds MongoDB client configuration
type Config struct {
	// Connection settings
	URI      string // MongoDB connection URI (mongodb+srv://...)
	Database string // Database name
	Timeout  int    // Timeout in seconds for operations (default: 10)

	// Vector search settings
	VectorDimensions int    // Embedding vector dimensions (e.g., 1536 for OpenAI)
	SimilarityMetric string // Similarity metric: "cosine", "euclidean", or "dotProduct"
}

// containsAny checks if a string contains any of the given substrings
func containsAny(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}

// NewClient creates a new MongoDB Atlas Vector Search client
func NewClient(config *Config) (*Client, error) {
	if config.URI == "" {
		return nil, fmt.Errorf("MongoDB URI is required")
	}
	if config.Database == "" {
		return nil, fmt.Errorf("MongoDB database name is required")
	}

	// Set defaults
	if config.VectorDimensions == 0 {
		config.VectorDimensions = 1536 // OpenAI ada-002 default
	}
	if config.SimilarityMetric == "" {
		config.SimilarityMetric = "cosine"
	}

	// Create MongoDB client options
	clientOptions := options.Client().ApplyURI(config.URI)

	// Set timeout for connection
	timeout := time.Duration(config.Timeout) * time.Second
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	clientOptions.SetConnectTimeout(timeout)
	clientOptions.SetServerSelectionTimeout(timeout)

	// Enable direct connection for better reliability with Atlas
	clientOptions.SetDirect(false)

	// Set TLS configuration - MongoDB Atlas requires TLS
	// Note: We don't set a custom TLS config here as the MongoDB driver
	// handles TLS automatically when using mongodb+srv:// URIs.
	// Setting a custom config without proper root CAs can cause issues.
	// The driver will use system certificates by default.

	// Create context with timeout for connection
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		// Provide helpful error message for common TLS issues
		errMsg := err.Error()
		if containsAny(errMsg, "tls:", "x509:", "certificate") {
			return nil, fmt.Errorf("failed to ping MongoDB: %w\n\nTLS connection error. Common causes:\n"+
				"  1. IP address not whitelisted in MongoDB Atlas Network Access\n"+
				"  2. Invalid credentials in connection string\n"+
				"  3. Using old MongoDB driver version\n"+
				"  → Check MongoDB Atlas dashboard: Network Access and Database Access", err)
		}
		if containsAny(errMsg, "timeout", "deadline") {
			return nil, fmt.Errorf("failed to ping MongoDB: %w\n\nConnection timeout. Common causes:\n"+
				"  1. IP address not whitelisted in MongoDB Atlas (most common)\n"+
				"  2. Firewall blocking outbound connections\n"+
				"  3. Incorrect connection string\n"+
				"  → Add your IP to Network Access in MongoDB Atlas dashboard", err)
		}
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(config.Database)

	return &Client{
		client:   client,
		database: database,
		config:   config,
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
	// Detect cloud: mongodb+srv or mongodb.net indicates Atlas cloud deployment
	isCloud := strings.Contains(c.config.URI, "+srv") || strings.Contains(c.config.URI, "mongodb.net")
	return vectordb.GetTimeoutForOperation(opType, isCloud, c.config.Timeout)
}

// Health checks the health of the MongoDB instance
func (c *Client) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeoutFor(vectordb.OperationTypeHealth))
	defer cancel()

	// Ping the database
	if err := c.client.Ping(ctx, nil); err != nil {
		errMsg := err.Error()
		// Provide helpful troubleshooting hints for common errors
		if containsAny(errMsg, "tls:", "x509:", "certificate") {
			return fmt.Errorf("MongoDB health check failed: %w\n\nTLS connection error. Common causes:\n"+
				"  1. IP address not whitelisted in MongoDB Atlas Network Access\n"+
				"  2. Invalid credentials in connection string\n"+
				"  3. Using old MongoDB driver version\n"+
				"  → Check MongoDB Atlas dashboard: Network Access and Database Access", err)
		}
		if containsAny(errMsg, "timeout", "deadline") {
			return fmt.Errorf("MongoDB health check failed: %w\n\nConnection timeout. Common causes:\n"+
				"  1. IP address not whitelisted in MongoDB Atlas (most common)\n"+
				"  2. Firewall blocking outbound connections\n"+
				"  3. Incorrect connection string or cluster paused\n"+
				"  → Add your IP to Network Access in MongoDB Atlas dashboard", err)
		}
		if containsAny(errMsg, "authentication", "credentials", "unauthorized") {
			return fmt.Errorf("MongoDB health check failed: %w\n\nAuthentication error. Common causes:\n"+
				"  1. Invalid username or password in connection string\n"+
				"  2. User not created in MongoDB Atlas Database Access\n"+
				"  3. User lacks permissions for specified database\n"+
				"  → Verify credentials in MongoDB Atlas: Database Access", err)
		}
		return fmt.Errorf("MongoDB health check failed: %w", err)
	}

	return nil
}

// Close closes the MongoDB client connection
func (c *Client) Close(ctx context.Context) error {
	if c.client != nil {
		return c.client.Disconnect(ctx)
	}
	return nil
}

// getCollection returns a MongoDB collection handle
func (c *Client) getCollection(name string) *mongo.Collection {
	return c.database.Collection(name)
}

// mongoDocument represents the MongoDB document structure with vector embeddings
type mongoDocument struct {
	ID         primitive.ObjectID     `bson:"_id,omitempty"`
	DocumentID string                 `bson:"document_id"`
	Text       string                 `bson:"text"`
	Content    string                 `bson:"content"`
	Image      string                 `bson:"image,omitempty"`
	ImageData  string                 `bson:"image_data,omitempty"`
	URL        string                 `bson:"url,omitempty"`
	Embedding  []float64              `bson:"embedding"`
	Metadata   map[string]interface{} `bson:"metadata,omitempty"`
	CreatedAt  time.Time              `bson:"created_at"`
	UpdatedAt  time.Time              `bson:"updated_at"`
}

// toMongoDocument converts a vectordb.Document to a mongoDocument
func (c *Client) toMongoDocument(doc *vectordb.Document) *mongoDocument {
	now := time.Now()
	return &mongoDocument{
		DocumentID: doc.ID,
		Text:       doc.Text,
		Content:    doc.Content,
		Image:      doc.Image,
		ImageData:  doc.ImageData,
		URL:        doc.URL,
		Metadata:   doc.Metadata,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// fromMongoDocument converts a mongoDocument to a vectordb.Document
func (c *Client) fromMongoDocument(mdoc *mongoDocument) *vectordb.Document {
	return &vectordb.Document{
		ID:        mdoc.DocumentID,
		Text:      mdoc.Text,
		Content:   mdoc.Content,
		Image:     mdoc.Image,
		ImageData: mdoc.ImageData,
		URL:       mdoc.URL,
		Metadata:  mdoc.Metadata,
	}
}
