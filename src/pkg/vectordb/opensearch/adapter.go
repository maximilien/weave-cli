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

	"github.com/aws/aws-sdk-go/aws/session"
	opensearch "github.com/opensearch-project/opensearch-go/v4"
	opensearchapi "github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	requestsigner "github.com/opensearch-project/opensearch-go/v4/signer/aws"

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
	_ = timeout // Will be used when we implement context timeouts

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
		// Check if this is AWS OpenSearch Service (domain contains amazonaws.com)
		isAWS := false
		awsRegion := os.Getenv("AWS_REGION")
		if awsRegion == "" {
			awsRegion = os.Getenv("AWS_DEFAULT_REGION")
		}

		for _, addr := range addresses {
			if strings.Contains(addr, "amazonaws.com") || strings.Contains(addr, "aoss.") {
				isAWS = true
				// Extract region from domain if not set (e.g., search-domain.us-east-1.es.amazonaws.com)
				if awsRegion == "" && strings.Contains(addr, ".es.amazonaws.com") {
					parts := strings.Split(addr, ".")
					for i, part := range parts {
						if i > 0 && parts[i-1] != "search" && strings.Contains(parts[i+1], "es") {
							awsRegion = part
							break
						}
					}
				}
				break
			}
		}

		if isAWS && awsRegion != "" {
			// AWS OpenSearch Service with Signature V4
			// Create AWS signer with default credentials and region
			signer, err := requestsigner.NewSignerWithService(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}, requestsigner.OpenSearchService)
			if err != nil {
				return nil, fmt.Errorf("OpenSearch: failed to create AWS signer: %w", err)
			}

			// Add AWS signer to OpenSearch config
			osConfig.Signer = signer
		} else if config.Username != "" && config.Password != "" {
			osConfig.Username = config.Username
			osConfig.Password = config.Password
		} else if config.APIKey != "" {
			// API key authentication
			osConfig.Header = http.Header{
				"Authorization": []string{"ApiKey " + config.APIKey},
			}
		}
	}

	// Create the OpenSearch API client
	client, err := opensearchapi.NewClient(opensearchapi.Config{
		Client: osConfig,
	})
	if err != nil {
		return nil, fmt.Errorf("OpenSearch: failed to create OpenSearch client: %w", err)
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
	ctx, cancel := context.WithTimeout(ctx, a.getTimeoutFor(vectordb.OperationTypeHealth))
	defer cancel()

	resp, err := a.client.Cluster.Health(ctx, nil)
	if err != nil {
		errMsg := err.Error()
		// Provide helpful troubleshooting hints for common errors
		if strings.Contains(errMsg, "connection refused") || strings.Contains(errMsg, "connect:") {
			return fmt.Errorf("OpenSearch health check failed: %w\n\nConnection refused. Common causes:\n"+
				"  1. OpenSearch not running (start with: docker run -p 9200:9200 -p 9600:9600 opensearchproject/opensearch:latest)\n"+
				"  2. Wrong URL/port in configuration (default: http://localhost:9200)\n"+
				"  3. For AWS OpenSearch: verify domain endpoint and VPC settings\n"+
				"  → Check OpenSearch is running and verify connection details", err)
		}
		if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline") {
			return fmt.Errorf("OpenSearch health check failed: %w\n\nConnection timeout. Common causes:\n"+
				"  1. OpenSearch starting up (can take 30-60s, needs 2GB+ RAM)\n"+
				"  2. Insufficient system resources (check logs: docker logs <container>)\n"+
				"  3. Network/firewall blocking port 9200\n"+
				"  → Wait for startup or check system resources", err)
		}
		if strings.Contains(errMsg, "401") || strings.Contains(errMsg, "403") || strings.Contains(errMsg, "Unauthorized") {
			return fmt.Errorf("OpenSearch health check failed: %w\n\nAuthentication error. Common causes:\n"+
				"  1. Invalid username or password (default: admin/admin for local)\n"+
				"  2. For AWS OpenSearch: check IAM credentials or fine-grained access control\n"+
				"  3. Security plugin enabled but credentials not provided\n"+
				"  → Verify credentials in configuration", err)
		}
		return fmt.Errorf("OpenSearch: health check failed: %w", err)
	}

	// Check cluster status
	if resp.Status == "red" {
		return fmt.Errorf("OpenSearch: cluster status is red - check cluster health and shard allocation")
	}

	return nil
}

// getTimeout returns the configured timeout as a time.Duration
func (a *Adapter) getTimeout() time.Duration {
	if a.config.Timeout > 0 {
		return time.Duration(a.config.Timeout) * time.Second
	}
	return 30 * time.Second // Default timeout
}

// getTimeoutFor returns an operation-specific timeout based on deployment type
func (a *Adapter) getTimeoutFor(opType vectordb.OperationType) time.Duration {
	isCloud := a.config.Type == vectordb.VectorDBTypeOpenSearchCloud
	return vectordb.GetTimeoutForOperation(opType, isCloud, a.config.Timeout)
}

// GetSchema returns the schema for a collection
func (a *Adapter) GetSchema(ctx context.Context, collectionName string) (*vectordb.CollectionSchema, error) {
	ctx, cancel := context.WithTimeout(ctx, a.getTimeoutFor(vectordb.OperationTypeSchema))
	defer cancel()

	// Check if collection exists first
	exists, err := a.CollectionExists(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("OpenSearch: failed to check collection existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("OpenSearch: collection not found: %s", collectionName)
	}

	// Return basic schema
	// OpenSearch schemas are defined at index creation time
	// For simplicity, return the default schema structure
	schema := &vectordb.CollectionSchema{
		Class:      collectionName,
		Vectorizer: "text-embedding-3-small", // Default vectorizer
		Properties: []vectordb.SchemaProperty{
			{
				Name:     "text",
				DataType: []string{"text"},
			},
			{
				Name:     "content",
				DataType: []string{"text"},
			},
			{
				Name:     "vector",
				DataType: []string{"knn_vector"},
			},
			{
				Name:     "metadata",
				DataType: []string{"object"},
			},
		},
	}

	return schema, nil
}

// UpdateSchema updates the schema for a collection
func (a *Adapter) UpdateSchema(ctx context.Context, collectionName string, schema *vectordb.CollectionSchema) error {
	// OpenSearch allows adding new fields via Put Mapping API
	// However, existing fields cannot be modified (similar to Elasticsearch)
	// For simplicity, return error suggesting recreation
	return fmt.Errorf("OpenSearch does not support schema updates for existing fields - delete and recreate the collection instead")
}

// GetDefaultSchema returns a default schema for the given type
func (a *Adapter) GetDefaultSchema(schemaType vectordb.SchemaType, collectionName string) *vectordb.CollectionSchema {
	return &vectordb.CollectionSchema{
		Class: collectionName,
	}
}

// ValidateSchema validates if a schema is compatible
func (a *Adapter) ValidateSchema(schema *vectordb.CollectionSchema) error {
	if schema.Class == "" {
		return fmt.Errorf("OpenSearch: collection name is required")
	}
	return nil
}

// Close closes the OpenSearch client connection
func (a *Adapter) Close() error {
	// TODO: Implement cleanup if needed
	return nil
}
