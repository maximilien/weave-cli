// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package weaviate

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/auth"
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
		return nil, fmt.Errorf("failed to create Weaviate client: %w", err)
	}

	return &Client{
		client: client,
		config: config,
	}, nil
}

// Health checks the health of the Weaviate instance
func (c *Client) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Try to get the meta information
	meta, err := c.client.Misc().MetaGetter().Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Weaviate meta: %w", err)
	}

	if meta == nil {
		return fmt.Errorf("received nil meta from Weaviate")
	}

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
