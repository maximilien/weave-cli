// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package elasticsearch

import (
	"fmt"
	"strings"
)

// Config holds Elasticsearch client configuration
type Config struct {
	// Connection settings (choose one method)
	// Method 1: Elastic Cloud
	CloudID string // Cloud ID from Elastic Cloud console
	APIKey  string // API key for authentication

	// Method 2: Self-hosted
	Addresses []string // List of Elasticsearch addresses (e.g., ["https://localhost:9200"])

	// Authentication (for self-hosted)
	Username string
	Password string

	// Optional settings
	CertFingerprint string // SHA256 fingerprint for cert validation
	Timeout         int    // Timeout in seconds (default: 10)
	MaxRetries      int    // Max retry attempts (default: 3)
	EnableDebugLog  bool   // Enable detailed logging

	// Vector search settings
	VectorDimensions int    // Embedding vector dimensions (e.g., 1536 for OpenAI)
	SimilarityMetric string // Similarity: "cosine", "dot_product", "l2_norm"
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Must have either CloudID or Addresses
	if c.CloudID == "" && len(c.Addresses) == 0 {
		return fmt.Errorf("either CloudID or Addresses must be provided")
	}

	// Cannot have both CloudID and Addresses
	if c.CloudID != "" && len(c.Addresses) > 0 {
		return fmt.Errorf("cannot specify both CloudID and Addresses, choose one connection method")
	}

	// CloudID requires APIKey
	if c.CloudID != "" && c.APIKey == "" {
		return fmt.Errorf("CloudID requires APIKey for authentication")
	}

	// Addresses requires authentication
	if len(c.Addresses) > 0 {
		hasAuth := (c.Username != "" && c.Password != "") || c.APIKey != ""
		if !hasAuth {
			return fmt.Errorf("self-hosted Elasticsearch requires authentication (Username+Password or APIKey)")
		}
	}

	// Validate addresses format
	for _, addr := range c.Addresses {
		if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
			return fmt.Errorf("invalid address format: %s (must start with http:// or https://)", addr)
		}
	}

	// Validate timeout
	if c.Timeout < 0 {
		return fmt.Errorf("timeout cannot be negative")
	}

	// Validate max retries
	if c.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}

	// Validate vector dimensions
	if c.VectorDimensions < 0 {
		return fmt.Errorf("vector dimensions cannot be negative")
	}

	// Validate similarity metric if specified
	if c.SimilarityMetric != "" {
		validMetrics := map[string]bool{
			"cosine":      true,
			"dot_product": true,
			"l2_norm":     true,
		}
		if !validMetrics[c.SimilarityMetric] {
			return fmt.Errorf("invalid similarity metric: %s (must be cosine, dot_product, or l2_norm)", c.SimilarityMetric)
		}
	}

	return nil
}

// SetDefaults sets default values for optional fields
func (c *Config) SetDefaults() {
	if c.VectorDimensions == 0 {
		c.VectorDimensions = 1536 // OpenAI ada-002 default
	}
	if c.SimilarityMetric == "" {
		c.SimilarityMetric = "cosine"
	}
	if c.Timeout == 0 {
		c.Timeout = 10
	}
	if c.MaxRetries == 0 {
		c.MaxRetries = 3
	}
}
