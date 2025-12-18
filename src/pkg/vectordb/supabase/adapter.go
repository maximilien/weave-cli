// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package supabase

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/maximilien/weave-cli/src/pkg/llm"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/supabase-community/supabase-go"
)

// Adapter wraps the Supabase client to implement the VectorDBClient interface
type Adapter struct {
	client    *supabase.Client
	db        *sql.DB
	config    *vectordb.Config
	llmClient *llm.OpenAIClient
}

// NewAdapter creates a new Supabase adapter
func NewAdapter(config *vectordb.Config) (*Adapter, error) {
	var client *supabase.Client
	var err error

	// For supabase-local, we don't need the HTTP client - just use direct PostgreSQL
	if config.Type != vectordb.VectorDBTypeSupabaseLocal {
		// For Supabase Cloud, we need the HTTP URL (SUPABASE_PROJECT_URL) not the database URL
		// Try to extract project URL from database URL if not provided separately
		projectURL := config.URL
		if projectURL == "" {
			// Extract from database URL: postgresql://...@db.PROJECT.supabase.co -> https://PROJECT.supabase.co
			if strings.Contains(config.DatabaseURL, "supabase.co") {
				parts := strings.Split(config.DatabaseURL, "@")
				if len(parts) >= 2 {
					hostPart := strings.Split(parts[1], ":")[0]
					hostPart = strings.Replace(hostPart, "db.", "", 1)
					hostPart = strings.Split(hostPart, "/")[0]
					projectURL = "https://" + hostPart
				}
			}
		}

		// Create Supabase HTTP client (only for cloud)
		client, err = supabase.NewClient(projectURL, config.DatabaseKey, nil)
		if err != nil {
			return nil, vectordb.ErrConnectionFailed("failed to create Supabase client", err)
		}
	}

	// Try to create direct database connection (optional - falls back to REST API)
	var db *sql.DB
	if config.DatabaseURL != "" {
		connStr, err := prepareConnectionString(config.DatabaseURL)
		if err == nil {
			db, err = sql.Open("postgres", connStr)
			if err == nil {
				// Test the connection (don't fail if this doesn't work)
				if pingErr := db.Ping(); pingErr != nil {
					// Connection failed, will use REST API fallback
					db.Close()
					db = nil
				}
			}
		}
	}

	// Create LLM client for embeddings (optional, only if OpenAI API key is available)
	var llmClient *llm.OpenAIClient
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey != "" {
		var err error
		llmClient, err = llm.NewOpenAIClient(openaiKey)
		if err != nil {
			// Just log warning, don't fail - embeddings won't work but other operations will
			fmt.Fprintf(os.Stderr, "Warning: Failed to create OpenAI client for embeddings: %v\n", err)
		}
	}

	return &Adapter{
		client:    client,
		db:        db,
		config:    config,
		llmClient: llmClient,
	}, nil
}

// Close closes the Supabase adapter and cleans up resources
func (a *Adapter) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

// prepareConnectionString prepares the connection string with IPv4 resolution
func prepareConnectionString(dbURL string) (string, error) {
	// Parse the database URL
	parsedURL, err := url.Parse(dbURL)
	if err != nil {
		return "", fmt.Errorf("Supabase: invalid database URL: %w", err)
	}

	// Extract hostname
	hostname := parsedURL.Hostname()
	port := parsedURL.Port()
	if port == "" {
		port = "5432"
	}

	// Determine SSL mode based on hostname
	// For local development (localhost, 127.0.0.1), disable SSL
	// For cloud services (supabase.co, etc.), require SSL
	sslMode := "require"
	if hostname == "localhost" || hostname == "127.0.0.1" || hostname == "::1" {
		sslMode = "disable"
	}

	// Try to resolve to IPv4 address
	ips, err := net.LookupIP(hostname)
	if err != nil {
		// If resolution fails, use original URL with connection parameters
		connStr := dbURL
		if !strings.Contains(connStr, "?") {
			connStr += "?"
		} else {
			connStr += "&"
		}
		connStr += fmt.Sprintf("sslmode=%s&connect_timeout=10", sslMode)
		return connStr, nil
	}

	// Find first IPv4 address
	var ipv4 string
	for _, ip := range ips {
		if ip.To4() != nil {
			ipv4 = ip.String()
			break
		}
	}

	// If we found an IPv4 address, replace hostname with it
	if ipv4 != "" {
		parsedURL.Host = fmt.Sprintf("%s:%s", ipv4, port)
	}

	// Add connection parameters
	query := parsedURL.Query()
	query.Set("sslmode", sslMode)
	query.Set("connect_timeout", "10")
	parsedURL.RawQuery = query.Encode()

	return parsedURL.String(), nil
}

// Health checks the health of the Supabase instance
func (a *Adapter) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, a.getTimeoutFor(vectordb.OperationTypeHealth))
	defer cancel()

	if a.db == nil {
		// No direct database connection, but HTTP client should work
		// We can't do a full health check without SQL access
		return vectordb.ErrConnectionFailed("Supabase database connection not available (IPv6 only or connection failed)",
			fmt.Errorf("Supabase: direct PostgreSQL connection failed - only REST API available (limited functionality)"))
	}
	if err := a.db.PingContext(ctx); err != nil {
		errMsg := err.Error()
		// Provide helpful troubleshooting hints for common errors
		if strings.Contains(errMsg, "connection refused") || strings.Contains(errMsg, "connect:") {
			return fmt.Errorf("Supabase health check failed: %w\n\nConnection refused. Common causes:\n"+
				"  1. Incorrect database URL (should be db.[project-ref].supabase.co)\n"+
				"  2. Database not started or paused (check Supabase dashboard)\n"+
				"  3. For local Supabase: verify docker containers running\n"+
				"  → Verify connection URL at https://supabase.com/dashboard/project/[project]/settings/database", err)
		}
		if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline") {
			return fmt.Errorf("Supabase health check failed: %w\n\nConnection timeout. Common causes:\n"+
				"  1. Firewall blocking PostgreSQL port 5432 or 6543 (pooler)\n"+
				"  2. Project paused due to inactivity (free tier)\n"+
				"  3. Network connectivity issues\n"+
				"  → Check project status in Supabase dashboard", err)
		}
		if strings.Contains(errMsg, "authentication") || strings.Contains(errMsg, "password") || strings.Contains(errMsg, "pq: password") {
			return fmt.Errorf("Supabase health check failed: %w\n\nAuthentication error. Common causes:\n"+
				"  1. Invalid database password (not the API key)\n"+
				"  2. Using service_role key instead of database password\n"+
				"  3. Password reset required\n"+
				"  → Get correct password from Settings > Database in Supabase dashboard", err)
		}
		return vectordb.ErrConnectionFailed("Supabase health check failed", err)
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
	// Detect cloud: SupabaseLocal is local, otherwise cloud
	isCloud := a.config.Type != vectordb.VectorDBTypeSupabaseLocal
	// Config timeout is already in seconds
	return vectordb.GetTimeoutForOperation(opType, isCloud, a.config.Timeout)
}

// hasDBConnection returns true if direct database connection is available
func (a *Adapter) hasDBConnection() bool {
	return a.db != nil
}

// requireDBConnection returns an error if database connection is not available
func (a *Adapter) requireDBConnection(operation string) error {
	if !a.hasDBConnection() {
		return vectordb.ErrUnsupported(fmt.Sprintf(
			"%s: direct PostgreSQL connection not available (IPv6-only endpoint or connection failed). "+
				"Please check SUPABASE_DATABASE_URL configuration or enable IPv6 connectivity",
			operation))
	}
	return nil
}

// wrapError wraps Supabase errors in vectordb errors
func (a *Adapter) wrapError(err error, operation string) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	// Connection errors
	if strings.Contains(errMsg, "connection refused") ||
		strings.Contains(errMsg, "no such host") ||
		strings.Contains(errMsg, "timeout") ||
		strings.Contains(errMsg, "connection reset") {
		return vectordb.ErrConnectionFailed(fmt.Sprintf("%s failed", operation), err)
	}

	// Authentication errors
	if strings.Contains(errMsg, "authentication failed") ||
		strings.Contains(errMsg, "invalid api key") ||
		strings.Contains(errMsg, "unauthorized") ||
		strings.Contains(errMsg, "permission denied") {
		return vectordb.ErrAuthenticationFailed(fmt.Sprintf("%s failed: %s", operation, errMsg))
	}

	// Not found errors
	if strings.Contains(errMsg, "not found") ||
		strings.Contains(errMsg, "does not exist") {
		return vectordb.ErrNotFound("resource", "unknown")
	}

	// Invalid configuration errors
	if strings.Contains(errMsg, "invalid") ||
		strings.Contains(errMsg, "malformed") {
		return vectordb.ErrInvalidConfig(fmt.Sprintf("%s failed: %s", operation, errMsg))
	}

	// Default to internal error
	return vectordb.ErrInternal(fmt.Sprintf("%s failed", operation), err)
}

// getTableName returns the table name for a collection
// Preserves original casing and characters (except spaces)
func (a *Adapter) getTableName(collectionName string) string {
	// Only replace spaces with underscores for valid PostgreSQL identifiers
	tableName := strings.ReplaceAll(collectionName, " ", "_")
	return fmt.Sprintf("collection_%s", tableName)
}

// quoteIdentifier quotes a PostgreSQL identifier to preserve case and special characters
func quoteIdentifier(identifier string) string {
	// Escape any double quotes in the identifier
	escaped := strings.ReplaceAll(identifier, `"`, `""`)
	return fmt.Sprintf(`"%s"`, escaped)
}

// ensureVectorExtension ensures that the pgvector extension is enabled
func (a *Adapter) ensureVectorExtension(ctx context.Context) error {
	query := "CREATE EXTENSION IF NOT EXISTS vector;"
	_, err := a.db.ExecContext(ctx, query)
	if err != nil {
		return a.wrapError(err, "ensure vector extension")
	}
	return nil
}

// convertMetadataToJSON converts metadata map to JSON string
func (a *Adapter) convertMetadataToJSON(metadata map[string]interface{}) (string, error) {
	if metadata == nil {
		return "{}", nil
	}

	jsonBytes, err := json.Marshal(metadata)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// convertJSONToMetadata converts JSON string to metadata map
func (a *Adapter) convertJSONToMetadata(jsonStr string) (map[string]interface{}, error) {
	if jsonStr == "" {
		return make(map[string]interface{}), nil
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &metadata); err != nil {
		return nil, err
	}

	return metadata, nil
}
