// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package elasticsearch

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// Client wraps the Elasticsearch TypedClient with vector database functionality
type Client struct {
	client *elasticsearch.TypedClient
	config *Config
}

// NewClient creates a new Elasticsearch client
func NewClient(config *Config) (*Client, error) {
	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Set defaults
	config.SetDefaults()

	// Build Elasticsearch config
	esConfig := elasticsearch.Config{}

	// Configure connection method
	if config.CloudID != "" {
		// Elastic Cloud connection
		esConfig.CloudID = config.CloudID
		esConfig.APIKey = config.APIKey
	} else {
		// Self-hosted connection
		esConfig.Addresses = config.Addresses

		// Authentication
		if config.APIKey != "" {
			esConfig.APIKey = config.APIKey
		} else {
			esConfig.Username = config.Username
			esConfig.Password = config.Password
		}
	}

	// Optional settings
	if config.CertFingerprint != "" {
		esConfig.CertificateFingerprint = config.CertFingerprint
	}
	esConfig.MaxRetries = config.MaxRetries

	// Create typed client
	typedClient, err := elasticsearch.NewTypedClient(esConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	// Test connection with ping
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(config.Timeout)*time.Second)
	defer cancel()

	if _, err := typedClient.Ping().Do(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping Elasticsearch: %w", err)
	}

	return &Client{
		client: typedClient,
		config: config,
	}, nil
}

// Health checks the health of the Elasticsearch cluster
func (c *Client) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeoutFor(vectordb.OperationTypeHealth))
	defer cancel()

	_, err := c.client.Ping().Do(ctx)
	if err != nil {
		errMsg := err.Error()
		// Provide helpful troubleshooting hints for common errors
		if strings.Contains(errMsg, "connection refused") || strings.Contains(errMsg, "connect:") {
			return fmt.Errorf("Elasticsearch health check failed: %w\n\nConnection refused. Common causes:\n"+
				"  1. Elasticsearch not running (start with: docker run -p 9200:9200 -e \"discovery.type=single-node\" docker.elastic.co/elasticsearch/elasticsearch:8.11.0)\n"+
				"  2. Wrong URL/port in configuration (default: http://localhost:9200)\n"+
				"  3. For Elastic Cloud: verify cloud ID and API key\n"+
				"  → Check Elasticsearch is running and verify connection details", err)
		}
		if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline") {
			return fmt.Errorf("Elasticsearch health check failed: %w\n\nConnection timeout. Common causes:\n"+
				"  1. Elasticsearch starting up (can take 30-60s, check logs: docker logs <container>)\n"+
				"  2. Insufficient resources (needs 2GB+ RAM)\n"+
				"  3. Network/firewall blocking port 9200\n"+
				"  → Wait for startup or check system resources", err)
		}
		if strings.Contains(errMsg, "401") || strings.Contains(errMsg, "403") || strings.Contains(errMsg, "Unauthorized") {
			return fmt.Errorf("Elasticsearch health check failed: %w\n\nAuthentication error. Common causes:\n"+
				"  1. Missing or invalid credentials for secured cluster\n"+
				"  2. For Elastic Cloud: check API key or cloud credentials\n"+
				"  3. Security enabled but credentials not provided\n"+
				"  → Verify username/password or API key configuration", err)
		}
		return fmt.Errorf("Elasticsearch health check failed: %w", err)
	}

	return nil
}

// Close closes the Elasticsearch client connection
func (c *Client) Close(ctx context.Context) error {
	// TypedClient doesn't have explicit close method
	// Connection will be closed when client is garbage collected
	return nil
}

// getTimeout returns the timeout duration for operations
func (c *Client) getTimeout() time.Duration {
	return time.Duration(c.config.Timeout) * time.Second
}

// getTimeoutFor returns an operation-specific timeout based on deployment type
func (c *Client) getTimeoutFor(opType vectordb.OperationType) time.Duration {
	// Detect cloud: CloudID indicates Elastic Cloud deployment
	isCloud := c.config.CloudID != ""
	return vectordb.GetTimeoutForOperation(opType, isCloud, c.config.Timeout)
}
