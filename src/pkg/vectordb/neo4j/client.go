// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package neo4j

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
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
		func(c *neo4j.Config) {
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
	// Verify connectivity
	err := c.driver.VerifyConnectivity(ctx)
	if err != nil {
		return fmt.Errorf("Neo4j health check failed: %w", err)
	}
	return nil
}

// executeQuery is a helper to execute Cypher queries
func (c *Client) executeQuery(ctx context.Context, query string, params map[string]interface{}) (*neo4j.EagerResult, error) {
	result, err := neo4j.ExecuteQuery(ctx, c.driver, query, params,
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase(c.config.Database),
	)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// executeWrite is a helper for write transactions
func (c *Client) executeWrite(ctx context.Context, work func(tx neo4j.ManagedTransaction) (interface{}, error)) (interface{}, error) {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: c.config.Database,
	})
	defer session.Close(ctx)

	return session.ExecuteWrite(ctx, work)
}

// executeRead is a helper for read transactions
func (c *Client) executeRead(ctx context.Context, work func(tx neo4j.ManagedTransaction) (interface{}, error)) (interface{}, error) {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: c.config.Database,
	})
	defer session.Close(ctx)

	return session.ExecuteRead(ctx, work)
}
