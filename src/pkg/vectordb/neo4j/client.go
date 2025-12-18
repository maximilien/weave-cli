// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package neo4j

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
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
	ctx, cancel := context.WithTimeout(ctx, c.getTimeoutFor(vectordb.OperationTypeHealth))
	defer cancel()

	// Verify connectivity
	err := c.driver.VerifyConnectivity(ctx)
	if err != nil {
		errMsg := err.Error()
		// Provide helpful troubleshooting hints for common errors
		if strings.Contains(errMsg, "connection refused") || strings.Contains(errMsg, "connect:") {
			return fmt.Errorf("Neo4j health check failed: %w\n\nConnection refused. Common causes:\n"+
				"  1. Neo4j server not running (start with: docker run -p 7474:7474 -p 7687:7687 neo4j:latest)\n"+
				"  2. Wrong URI or port in configuration (Bolt: 7687, HTTP: 7474)\n"+
				"  3. For Neo4j Aura: verify cluster URI and credentials\n"+
				"  → Check Neo4j is running and verify connection details", err)
		}
		if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline") {
			return fmt.Errorf("Neo4j health check failed: %w\n\nConnection timeout. Common causes:\n"+
				"  1. Neo4j starting up (check logs: docker logs <container>)\n"+
				"  2. Network/firewall blocking port 7687\n"+
				"  3. For Neo4j Aura: verify cluster is running and not paused\n"+
				"  → Wait for startup or check network connectivity", err)
		}
		if strings.Contains(errMsg, "authentication") || strings.Contains(errMsg, "credentials") || strings.Contains(errMsg, "Unauthorized") {
			return fmt.Errorf("Neo4j health check failed: %w\n\nAuthentication error. Common causes:\n"+
				"  1. Invalid username or password\n"+
				"  2. Default credentials not changed (neo4j/neo4j requires password change)\n"+
				"  3. For Neo4j Aura: verify generated password from console\n"+
				"  → Check credentials at https://console.neo4j.io", err)
		}
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

// getTimeoutFor returns an operation-specific timeout based on deployment type
func (c *Client) getTimeoutFor(opType vectordb.OperationType) time.Duration {
	// Detect cloud by URI (Aura uses neo4j+s:// or neo4j+ssc://)
	isCloud := strings.Contains(c.config.URI, "neo4j+s") || strings.Contains(c.config.URI, "aura")
	// Convert time.Duration to seconds for GetTimeoutForOperation
	timeoutSecs := int(c.config.Timeout.Seconds())
	return vectordb.GetTimeoutForOperation(opType, isCloud, timeoutSecs)
}
