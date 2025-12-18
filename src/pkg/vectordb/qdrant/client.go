// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package qdrant

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	qdrant "github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client wraps the Qdrant gRPC client with vector database functionality
type Client struct {
	qdrantClient      qdrant.QdrantClient
	collectionsClient qdrant.CollectionsClient
	pointsClient      qdrant.PointsClient
	conn              *grpc.ClientConn
	config            *Config
	collection        string // Default collection name
}

// Config holds Qdrant client configuration
type Config struct {
	// Connection settings
	Host   string // Qdrant server host (e.g., "localhost")
	Port   int    // Qdrant server port (default: 6334 for gRPC)
	APIKey string // API key for Qdrant Cloud
	UseTLS bool   // Whether to use TLS (true for cloud, false for local)

	// Vector settings
	VectorDimensions int    // Embedding vector dimensions (e.g., 1536 for OpenAI)
	SimilarityMetric string // Distance metric: "Cosine", "Dot", or "Euclidean"

	// Connection settings
	Timeout int // Timeout in seconds for operations (default: 10)
}

// NewClient creates a new Qdrant client
func NewClient(config *Config) (*Client, error) {
	if config.Host == "" {
		return nil, fmt.Errorf("Qdrant host is required")
	}

	// Set defaults
	if config.Port == 0 {
		config.Port = 6334 // Default gRPC port
	}
	if config.VectorDimensions == 0 {
		config.VectorDimensions = 1536 // OpenAI text-embedding-3-small default
	}
	if config.SimilarityMetric == "" {
		config.SimilarityMetric = "Cosine"
	}
	if config.Timeout == 0 {
		config.Timeout = 10
	}

	// Create gRPC connection
	var opts []grpc.DialOption
	if config.UseTLS {
		// For Qdrant Cloud, we would use TLS credentials
		// For now, using insecure for local development
		// TODO: Implement proper TLS configuration for cloud
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Connect to Qdrant using new grpc.NewClient API
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Qdrant client for %s: %w", addr, err)
	}

	// Create Qdrant clients
	qdrantClient := qdrant.NewQdrantClient(conn)
	collectionsClient := qdrant.NewCollectionsClient(conn)
	pointsClient := qdrant.NewPointsClient(conn)

	return &Client{
		qdrantClient:      qdrantClient,
		collectionsClient: collectionsClient,
		pointsClient:      pointsClient,
		conn:              conn,
		config:            config,
	}, nil
}

// Close closes the Qdrant client connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Health checks if the Qdrant server is healthy and accessible
func (c *Client) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeoutFor(vectordb.OperationTypeHealth))
	defer cancel()

	// Use the HealthCheck gRPC method
	req := &qdrant.HealthCheckRequest{}
	_, err := c.qdrantClient.HealthCheck(ctx, req)
	if err != nil {
		errMsg := err.Error()
		// Provide helpful troubleshooting hints for common errors
		if strings.Contains(errMsg, "connection refused") || strings.Contains(errMsg, "connect:") {
			return fmt.Errorf("Qdrant health check failed: %w\n\nConnection refused. Common causes:\n"+
				"  1. Qdrant server not running (start with: docker run -p 6333:6333 -p 6334:6334 qdrant/qdrant)\n"+
				"  2. Wrong host or port in configuration\n"+
				"  3. For Qdrant Cloud: check API endpoint and API key\n"+
				"  → Verify connection details in config", err)
		}
		if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline") {
			return fmt.Errorf("Qdrant health check failed: %w\n\nConnection timeout. Common causes:\n"+
				"  1. Qdrant server not responding (check logs: docker logs <container>)\n"+
				"  2. Network/firewall issues\n"+
				"  3. For Qdrant Cloud: verify cluster is running\n"+
				"  → Check Qdrant status and network connectivity", err)
		}
		if strings.Contains(errMsg, "401") || strings.Contains(errMsg, "403") || strings.Contains(errMsg, "Unauthorized") {
			return fmt.Errorf("Qdrant health check failed: %w\n\nAuthentication error. Common causes:\n"+
				"  1. Invalid or missing API key for Qdrant Cloud\n"+
				"  2. API key not configured in environment\n"+
				"  → Set QDRANT_API_KEY for cloud deployments", err)
		}
		return fmt.Errorf("Qdrant health check failed: %w", err)
	}
	return nil
}

// SetCollection sets the default collection name for operations
func (c *Client) SetCollection(name string) {
	c.collection = name
}

// GetCollection returns the current default collection name
func (c *Client) GetCollection() string {
	return c.collection
}

// getTimeout returns the configured timeout as a time.Duration
func (c *Client) getTimeout() time.Duration {
	return time.Duration(c.config.Timeout) * time.Second
}

// getTimeoutFor returns an operation-specific timeout based on deployment type
func (c *Client) getTimeoutFor(opType vectordb.OperationType) time.Duration {
	isCloud := c.config.UseTLS // UseTLS indicates cloud deployment
	return vectordb.GetTimeoutForOperation(opType, isCloud, c.config.Timeout)
}
