// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package redis

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/logging"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/redis/go-redis/v9"
)

const (
	// keyPrefix is prepended to all Redis keys for namespace isolation.
	keyPrefix = "weave"

	// defaultVectorField is the HASH field name for the vector blob.
	defaultVectorField = "vector"

	// defaultContentField is the HASH field name for document text.
	defaultContentField = "content"
)

// Client wraps a go-redis client with vector-search helpers.
type Client struct {
	rdb    *redis.Client
	config *Config
}

// Config holds Redis client configuration.
type Config struct {
	Addr             string // host:port (e.g. "localhost:6379")
	Password         string
	Username         string
	UseTLS           bool
	DB               int    // Redis database number
	VectorDimensions int    // e.g. 1536
	SimilarityMetric string // COSINE, L2, IP
	Timeout          int    // seconds
}

// NewClient creates a new Redis client.
func NewClient(cfg *Config) (*Client, error) {
	if cfg.Addr == "" {
		return nil, fmt.Errorf("Redis address is required")
	}
	if cfg.VectorDimensions == 0 {
		cfg.VectorDimensions = 1536
	}
	if cfg.SimilarityMetric == "" {
		cfg.SimilarityMetric = "COSINE"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 10
	}

	opts := &redis.Options{
		Addr:        cfg.Addr,
		Password:    cfg.Password,
		Username:    cfg.Username,
		DB:          cfg.DB,
		DialTimeout: time.Duration(cfg.Timeout) * time.Second,
	}
	if cfg.UseTLS {
		opts.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	rdb := redis.NewClient(opts)
	logging.Debug("Redis client created: addr=%s tls=%v db=%d", cfg.Addr, cfg.UseTLS, cfg.DB)

	return &Client{rdb: rdb, config: cfg}, nil
}

// Close closes the Redis connection.
func (c *Client) Close() error {
	return c.rdb.Close()
}

// Health checks if the Redis server is reachable and has the search module.
func (c *Client) Health(ctx context.Context) error {
	timeout := c.getTimeoutFor(vectordb.OperationTypeHealth)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := c.rdb.Ping(ctx).Err(); err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			return fmt.Errorf("Redis health check failed: %w\n\n"+
				"Connection refused. Common causes:\n"+
				"  1. Redis not running (start with: docker run -d -p 6379:6379 redis/redis-stack:latest)\n"+
				"  2. Wrong host or port in configuration\n"+
				"  → Verify connection details in config", err)
		}
		return fmt.Errorf("Redis health check failed: %w", err)
	}

	// Verify RediSearch module is loaded by trying FT._LIST
	if err := c.rdb.Do(ctx, "FT._LIST").Err(); err != nil {
		return fmt.Errorf("Redis health check failed: RediSearch module not loaded.\n"+
			"  Use Redis Stack: docker run -d -p 6379:6379 redis/redis-stack:latest\n"+
			"  Error: %w", err)
	}

	logging.Debug("Redis health check passed: addr=%s", c.config.Addr)
	return nil
}

// getTimeout returns the configured timeout as time.Duration.
func (c *Client) getTimeout() time.Duration {
	return time.Duration(c.config.Timeout) * time.Second
}

// getTimeoutFor returns an operation-specific timeout.
func (c *Client) getTimeoutFor(opType vectordb.OperationType) time.Duration {
	return vectordb.GetTimeoutForOperation(opType, c.config.UseTLS, c.config.Timeout)
}

// indexName returns the FT index name for a collection.
func indexName(collection string) string {
	return fmt.Sprintf("%s:%s", keyPrefix, collection)
}

// docKey returns the Redis key for a document in a collection.
func docKey(collection, docID string) string {
	return fmt.Sprintf("%s:%s:%s", keyPrefix, collection, docID)
}

// docKeyPrefix returns the key prefix for all docs in a collection.
func docKeyPrefix(collection string) string {
	return fmt.Sprintf("%s:%s:", keyPrefix, collection)
}
