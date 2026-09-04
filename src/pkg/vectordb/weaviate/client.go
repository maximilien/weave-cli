// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package weaviate

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/logging"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/weaviate/weaviate-go-client/v5/weaviate"
	"github.com/weaviate/weaviate-go-client/v5/weaviate/auth"
)

// FieldDefinition represents a field in a collection
type FieldDefinition struct {
	Name string
	Type string
}

// CollectionSchema represents a collection schema
type CollectionSchema struct {
	Class      string           `json:"class"`
	Vectorizer string           `json:"vectorizer,omitempty"`
	Properties []SchemaProperty `json:"properties"`
}

// SchemaProperty represents a property in a collection schema
type SchemaProperty struct {
	Name             string                 `json:"name" yaml:"name"`
	DataType         []string               `json:"dataType" yaml:"datatype"`
	Description      string                 `json:"description,omitempty" yaml:"description,omitempty"`
	NestedProperties []SchemaProperty       `json:"nestedProperties,omitempty" yaml:"nestedproperties,omitempty"`
	JSONSchema       map[string]interface{} `json:"json_schema,omitempty" yaml:"json_schema,omitempty"`
}

// Client wraps the Weaviate client with additional functionality
type Client struct {
	client *weaviate.Client
	config *Config
}

// Config holds Weaviate client configuration
type Config struct {
	URL          string
	APIKey       string
	OpenAIAPIKey string
	Timeout      int // Timeout in seconds for operations (default: 10)
}

// SchemaType represents the type of collection schema
type SchemaType string

const (
	SchemaTypeText  SchemaType = "text"
	SchemaTypeImage SchemaType = "image"
)

// NewClient creates a new Weaviate client
func NewClient(config *Config) (*Client, error) {
	var client *weaviate.Client
	var err error

	// Parse URL to extract host and scheme
	host := config.URL
	scheme := "http"

	// Remove protocol if present
	if strings.HasPrefix(host, "http://") {
		host = strings.TrimPrefix(host, "http://")
		scheme = "http"
	} else if strings.HasPrefix(host, "https://") {
		host = strings.TrimPrefix(host, "https://")
		scheme = "https"
	}

	if config.APIKey != "" {
		// Use API key authentication for Weaviate Cloud
		headers := map[string]string{
			"X-Openai-Api-Key": config.OpenAIAPIKey,
		}

		// Add cluster URL header for Weaviate Cloud Serverless (text2vec-weaviate vectorizer)
		// The cluster URL is the full URL including scheme
		if scheme == "https" {
			headers["X-Weaviate-Cluster-Url"] = scheme + "://" + host
		}

		client, err = weaviate.NewClient(weaviate.Config{
			Host:   host,
			Scheme: scheme,
			AuthConfig: auth.ApiKey{
				Value: config.APIKey,
			},
			Headers: headers,
		})
	} else {
		// Use no authentication for local Weaviate
		client, err = weaviate.NewClient(weaviate.Config{
			Host:   host,
			Scheme: scheme,
		})
	}

	if err != nil {
		return nil, logging.WrapError(err, "NewClient", "weaviate", config.URL)
	}

	logging.Debug("Weaviate client created successfully: url=%s", config.URL)

	return &Client{
		client: client,
		config: config,
	}, nil
}

// getTimeout returns the timeout duration for operations
func (c *Client) getTimeout() time.Duration {
	timeout := c.config.Timeout
	if timeout == 0 {
		timeout = 10 // default 10 seconds
	}
	return time.Duration(timeout) * time.Second
}

// getTimeoutFor returns an operation-specific timeout based on deployment type
func (c *Client) getTimeoutFor(opType vectordb.OperationType) time.Duration {
	// Detect cloud: APIKey indicates Weaviate Cloud deployment
	isCloud := c.config.APIKey != ""
	return vectordb.GetTimeoutForOperation(opType, isCloud, c.config.Timeout)
}

// Health checks the health of the Weaviate instance
func (c *Client) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.getTimeoutFor(vectordb.OperationTypeHealth))
	defer cancel()

	// Try to get the meta information
	meta, err := c.client.Misc().MetaGetter().Do(ctx)
	if err != nil {
		errMsg := err.Error()
		context := make(map[string]interface{})
		context["timeout"] = c.getTimeoutFor(vectordb.OperationTypeHealth).String()

		// Provide helpful troubleshooting hints for common errors
		if strings.Contains(errMsg, "connection refused") || strings.Contains(errMsg, "connect:") {
			context["hint"] = "connection_refused"
			wrappedErr := logging.WrapErrorWithContext(err, "Health", "weaviate", c.config.URL, context)
			return fmt.Errorf("%w\n\nConnection refused. Common causes:\n"+
				"  1. Weaviate server not running (start with: docker run -p 8080:8080 -p 50051:50051 semitechnologies/weaviate:latest)\n"+
				"  2. Wrong URL in configuration (check host and port)\n"+
				"  3. For Weaviate Cloud: verify cluster URL and API key\n"+
				"  → Verify connection details in config", wrappedErr)
		}
		if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline") {
			context["hint"] = "timeout"
			wrappedErr := logging.WrapErrorWithContext(err, "Health", "weaviate", c.config.URL, context)
			return fmt.Errorf("%w\n\nConnection timeout. Common causes:\n"+
				"  1. Weaviate server not responding (check logs: docker logs <container>)\n"+
				"  2. Network/firewall issues blocking port 8080\n"+
				"  3. For Weaviate Cloud: verify cluster is running\n"+
				"  → Check Weaviate status and network connectivity", wrappedErr)
		}
		if strings.Contains(errMsg, "401") || strings.Contains(errMsg, "403") || strings.Contains(errMsg, "Unauthorized") {
			context["hint"] = "auth_error"
			wrappedErr := logging.WrapErrorWithContext(err, "Health", "weaviate", c.config.URL, context)
			return fmt.Errorf("%w\n\nAuthentication error. Common causes:\n"+
				"  1. Invalid or missing API key for Weaviate Cloud\n"+
				"  2. API key not set in WEAVIATE_API_KEY environment variable\n"+
				"  3. Wrong authentication scheme (API key vs OIDC)\n"+
				"  → Verify API key at https://console.weaviate.cloud", wrappedErr)
		}
		return logging.WrapErrorWithContext(err, "Health", "weaviate", c.config.URL, context)
	}

	if meta == nil {
		return logging.WrapError(fmt.Errorf("received nil meta from Weaviate"), "Health", "weaviate", c.config.URL)
	}

	logging.Debug("Weaviate health check passed: url=%s version=%s", c.config.URL, meta.Version)
	return nil
}

// mapWeaviateDataType maps our field types to Weaviate data types
func mapWeaviateDataType(fieldType string) string {
	switch fieldType {
	case "text":
		return "text"
	case "int":
		return "int"
	case "float":
		return "number"
	case "bool":
		return "boolean"
	case "date":
		return "date"
	case "object":
		return "object"
	default:
		return "text" // Default to text
	}
}

// isImageCollection checks if a collection name suggests it contains images
func isImageCollection(collectionName string) bool {
	imageKeywords := []string{"image", "img", "photo", "picture", "visual"}
	name := strings.ToLower(collectionName)
	for _, keyword := range imageKeywords {
		if strings.Contains(name, keyword) {
			return true
		}
	}
	return false
}
