// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package neo4j

import (
	"context"
	"fmt"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	config2 "github.com/neo4j/neo4j-go-driver/v5/neo4j/config"
)

// Client wraps the Neo4j driver to provide vector database operations
type Client struct {
	driver neo4j.DriverWithContext
	config *Config
}

// NewClient creates a new Neo4j client
func NewClient(config *Config) (*Client, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Set defaults if not provided
	if config.Database == "" {
		config.Database = "neo4j"
	}
	if config.MaxConnections == 0 {
		config.MaxConnections = 50
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * 1000000000 // 30 seconds in nanoseconds
	}
	if config.VectorDimensions == 0 {
		config.VectorDimensions = 1536
	}
	if config.SimilarityMetric == "" {
		config.SimilarityMetric = "cosine"
	}

	// Create driver
	driver, err := neo4j.NewDriverWithContext(
		config.URI,
		neo4j.BasicAuth(config.Username, config.Password, ""),
		func(c *config2.Config) {
			c.MaxConnectionPoolSize = config.MaxConnections
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Neo4j driver: %w", err)
	}

	return &Client{
		driver: driver,
		config: config,
	}, nil
}

// Close closes the Neo4j driver
func (c *Client) Close(ctx context.Context) error {
	return c.driver.Close(ctx)
}

// Health checks the connection to Neo4j
func (c *Client) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	// Verify connectivity
	err := c.driver.VerifyConnectivity(ctx)
	if err != nil {
		return fmt.Errorf("Neo4j health check failed: %w", err)
	}
	return nil
}

// executeQuery is a helper to execute Cypher queries
func (c *Client) executeQuery(ctx context.Context, query string, params map[string]interface{}) (*neo4j.EagerResult, error) {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeout())
	defer cancel()

	result, err := neo4j.ExecuteQuery(ctx, c.driver, query, params,
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase(c.config.Database),
	)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// getTimeout returns the configured timeout
func (c *Client) getTimeout() time.Duration {
	return c.config.Timeout
}
