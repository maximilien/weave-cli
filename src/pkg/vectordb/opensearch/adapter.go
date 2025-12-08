// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package opensearch

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	opensearch "github.com/opensearch-project/opensearch-go/v4"
	opensearchapi "github.com/opensearch-project/opensearch-go/v4/opensearchapi"

	"github.com/maximilien/weave-cli/src/pkg/llm"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// Adapter wraps the OpenSearch client to implement the vectordb.VectorDBClient interface
type Adapter struct {
	client    *opensearchapi.Client
	config    *vectordb.Config
	llmClient *llm.OpenAIClient
}

// NewAdapter creates a new OpenSearch adapter from the vectordb.Config
func NewAdapter(config *vectordb.Config) (*Adapter, error) {
	// Determine URL/address
	url := config.URL
	if url == "" {
		url = config.Address
	}
	if url == "" {
		url = "http://localhost:9200"
	}

	// Parse addresses (support comma-separated)
	addresses := strings.Split(url, ",")
	for i, addr := range addresses {
		addresses[i] = strings.TrimSpace(addr)
	}

	// Create HTTP transport with timeout
	timeout := time.Duration(config.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.Type == vectordb.VectorDBTypeOpenSearchLocal,
		},
	}

	// Create OpenSearch client config
	osConfig := opensearch.Config{
		Addresses: addresses,
		Transport: transport,
	}

	// Add authentication for cloud
	if config.Type == vectordb.VectorDBTypeOpenSearchCloud {
		if config.Username != "" && config.Password != "" {
			osConfig.Username = config.Username
			osConfig.Password = config.Password
		} else if config.APIKey != "" {
			// API key authentication
			osConfig.Header = http.Header{
				"Authorization": []string{"ApiKey " + config.APIKey},
			}
		}
		// TODO: Add AWS Signature V4 support for AWS OpenSearch Service
	}

	// Create the OpenSearch API client
	client, err := opensearchapi.NewClient(opensearchapi.Config{
		Client: osConfig,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenSearch client: %w", err)
	}

	// Create LLM client for embeddings (optional)
	var llmClient *llm.OpenAIClient
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey != "" {
		llmClient, err = llm.NewOpenAIClient(openaiKey)
		if err != nil {
			// Just log warning, don't fail
			fmt.Fprintf(os.Stderr, "Warning: Failed to create OpenAI client for embeddings: %v\n", err)
		}
	}

	return &Adapter{
		client:    client,
		config:    config,
		llmClient: llmClient,
	}, nil
}

// Health checks the health of the OpenSearch cluster
func (a *Adapter) Health(ctx context.Context) error {
	resp, err := a.client.Cluster.Health(ctx, nil)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	// Check cluster status
	if resp.Status == "red" {
		return fmt.Errorf("cluster status is red")
	}

	return nil
}

// CreateDocument indexes a document with vector
func (a *Adapter) CreateDocument(ctx context.Context, collectionName string, document *vectordb.Document) error {
	// TODO: Implement document indexing
	return fmt.Errorf("OpenSearch adapter not yet fully implemented")
}

// CreateDocuments bulk indexes multiple documents
func (a *Adapter) CreateDocuments(ctx context.Context, collectionName string, documents []*vectordb.Document) error {
	// TODO: Implement bulk indexing
	return fmt.Errorf("OpenSearch adapter not yet fully implemented")
}

// GetDocument retrieves a document by ID
func (a *Adapter) GetDocument(ctx context.Context, collectionName, documentID string) (*vectordb.Document, error) {
	// TODO: Implement document retrieval
	return nil, fmt.Errorf("OpenSearch adapter not yet fully implemented")
}

// UpdateDocument updates a document
func (a *Adapter) UpdateDocument(ctx context.Context, collectionName string, document *vectordb.Document) error {
	// TODO: Implement document update
	return fmt.Errorf("OpenSearch adapter not yet fully implemented")
}

// DeleteDocument deletes a document by ID
func (a *Adapter) DeleteDocument(ctx context.Context, collectionName, documentID string) error {
	// TODO: Implement document deletion
	return fmt.Errorf("OpenSearch adapter not yet fully implemented")
}

// DeleteDocuments deletes multiple documents
func (a *Adapter) DeleteDocuments(ctx context.Context, collectionName string, documentIDs []string) error {
	// TODO: Implement bulk deletion
	return fmt.Errorf("OpenSearch adapter not yet fully implemented")
}

// DeleteDocumentsByMetadata deletes documents matching metadata filters
func (a *Adapter) DeleteDocumentsByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}) error {
	// TODO: Implement delete by query
	return fmt.Errorf("OpenSearch adapter not yet fully implemented")
}

// ListDocuments lists all documents in a collection
func (a *Adapter) ListDocuments(ctx context.Context, collectionName string, limit int, offset int) ([]*vectordb.Document, error) {
	// TODO: Implement document listing
	return nil, fmt.Errorf("OpenSearch adapter not yet fully implemented")
}

// SearchSemantic performs k-NN vector search
func (a *Adapter) SearchSemantic(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	// TODO: Implement k-NN search
	return nil, fmt.Errorf("OpenSearch adapter not yet fully implemented")
}

// SearchBM25 performs BM25 text search
func (a *Adapter) SearchBM25(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	// TODO: Implement BM25 search
	return nil, fmt.Errorf("OpenSearch adapter not yet fully implemented")
}

// SearchHybrid performs hybrid search (vector + BM25)
func (a *Adapter) SearchHybrid(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	// TODO: Implement hybrid search
	return nil, fmt.Errorf("OpenSearch adapter not yet fully implemented")
}

// SearchByMetadata searches documents by metadata filters
func (a *Adapter) SearchByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	// TODO: Implement metadata search
	return nil, fmt.Errorf("OpenSearch adapter not yet fully implemented")
}

// GetSchema returns the schema for a collection
func (a *Adapter) GetSchema(ctx context.Context, collectionName string) (*vectordb.CollectionSchema, error) {
	// TODO: Implement schema retrieval
	return nil, fmt.Errorf("OpenSearch adapter not yet fully implemented")
}

// UpdateSchema updates the schema for a collection
func (a *Adapter) UpdateSchema(ctx context.Context, collectionName string, schema *vectordb.CollectionSchema) error {
	// TODO: Implement schema update
	return fmt.Errorf("OpenSearch adapter not yet fully implemented")
}

// GetDefaultSchema returns a default schema for the given type
func (a *Adapter) GetDefaultSchema(schemaType vectordb.SchemaType, collectionName string) *vectordb.CollectionSchema {
	// TODO: Implement default schema generation
	return nil
}

// ValidateSchema validates if a schema is compatible
func (a *Adapter) ValidateSchema(schema *vectordb.CollectionSchema) error {
	// TODO: Implement schema validation
	return fmt.Errorf("OpenSearch adapter not yet fully implemented")
}

// Close closes the OpenSearch client connection
func (a *Adapter) Close() error {
	// TODO: Implement cleanup if needed
	return nil
}
