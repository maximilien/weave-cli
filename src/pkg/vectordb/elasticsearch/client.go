// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package elasticsearch

import (
	"context"
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v9"
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
	timeout := time.Duration(c.config.Timeout) * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	_, err := c.client.Ping().Do(ctx)
	if err != nil {
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

// getClient returns the underlying TypedClient for advanced operations
func (c *Client) getClient() *elasticsearch.TypedClient {
	return c.client
}
