// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package weaviate

import (
	"context"
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// Adapter wraps the existing Weaviate client to implement the VectorDBClient interface
type Adapter struct {
	client      *Client
	weaveClient *WeaveClient
	config      *vectordb.Config
}

// NewAdapter creates a new Weaviate adapter
func NewAdapter(config *vectordb.Config) (*Adapter, error) {
	// Convert vectordb.Config to weaviate.Config
	weaviateConfig := &Config{
		URL:          config.URL,
		APIKey:       config.APIKey,
		OpenAIAPIKey: config.OpenAIAPIKey,
		Timeout:      config.Timeout,
	}

	// Create the original Weaviate client
	client, err := NewClient(weaviateConfig)
	if err != nil {
		return nil, vectordb.ErrConnectionFailed("failed to create Weaviate client", err)
	}

	// Create the enhanced WeaveClient
	weaveClient, err := NewWeaveClient(weaviateConfig)
	if err != nil {
		return nil, vectordb.ErrConnectionFailed("failed to create WeaveClient", err)
	}

	return &Adapter{
		client:      client,
		weaveClient: weaveClient,
		config:      config,
	}, nil
}

// Health checks the health of the Weaviate instance
func (a *Adapter) Health(ctx context.Context) error {
	if err := a.client.Health(ctx); err != nil {
		return vectordb.ErrConnectionFailed("Weaviate health check failed", err)
	}
	return nil
}

// Close closes the Weaviate adapter and cleans up resources
// Note: Weaviate Go SDK uses HTTP/REST client which doesn't require explicit cleanup
// HTTP connections are managed by the Go HTTP transport and closed when no longer needed
func (a *Adapter) Close() error {
	// No cleanup required for HTTP client
	return nil
}

// convertDocument converts between vectordb.Document and weaviate.Document
func (a *Adapter) convertDocument(doc *vectordb.Document) *Document {
	if doc == nil {
		return nil
	}
	return &Document{
		ID:        doc.ID,
		Text:      doc.Text,
		Content:   doc.Content,
		Image:     doc.Image,
		ImageData: doc.ImageData,
		URL:       doc.URL,
		Metadata:  doc.Metadata,
	}
}

// convertDocumentFromWeaviate converts from weaviate.Document to vectordb.Document
func (a *Adapter) convertDocumentFromWeaviate(doc *Document) *vectordb.Document {
	if doc == nil {
		return nil
	}
	return &vectordb.Document{
		ID:        doc.ID,
		Text:      doc.Text,
		Content:   doc.Content,
		Image:     doc.Image,
		ImageData: doc.ImageData,
		URL:       doc.URL,
		Metadata:  doc.Metadata,
		Embedding: doc.Embedding, // CRITICAL for backup/restore
	}
}

// convertDocumentsFromWeaviate converts a slice of weaviate.Document to vectordb.Document
func (a *Adapter) convertDocumentsFromWeaviate(docs []*Document) []*vectordb.Document {
	if docs == nil {
		return nil
	}
	result := make([]*vectordb.Document, len(docs))
	for i, doc := range docs {
		result[i] = a.convertDocumentFromWeaviate(doc)
	}
	return result
}

// convertQueryOptions converts vectordb.QueryOptions to weaviate.QueryOptions
func (a *Adapter) convertQueryOptions(opts *vectordb.QueryOptions) *QueryOptions {
	if opts == nil {
		return &QueryOptions{}
	}
	return &QueryOptions{
		TopK:           opts.TopK,
		Distance:       opts.Distance,
		SearchMetadata: opts.SearchMetadata,
		NoTruncate:     opts.NoTruncate,
		UseBM25:        opts.UseBM25,
	}
}

// convertSchemaFromWeaviate converts from weaviate.CollectionSchema to vectordb.CollectionSchema
func (a *Adapter) convertSchemaFromWeaviate(schema *CollectionSchema) *vectordb.CollectionSchema {
	if schema == nil {
		return nil
	}

	properties := make([]vectordb.SchemaProperty, len(schema.Properties))
	for i, prop := range schema.Properties {
		properties[i] = vectordb.SchemaProperty{
			Name:             prop.Name,
			DataType:         prop.DataType,
			Description:      prop.Description,
			NestedProperties: a.convertNestedPropertiesFromWeaviate(prop.NestedProperties),
			JSONSchema:       prop.JSONSchema,
		}
	}

	return &vectordb.CollectionSchema{
		Class:      schema.Class,
		Vectorizer: schema.Vectorizer,
		Properties: properties,
	}
}

// convertNestedPropertiesFromWeaviate converts nested properties from Weaviate format
func (a *Adapter) convertNestedPropertiesFromWeaviate(props []SchemaProperty) []vectordb.SchemaProperty {
	if props == nil {
		return nil
	}

	result := make([]vectordb.SchemaProperty, len(props))
	for i, prop := range props {
		result[i] = vectordb.SchemaProperty{
			Name:             prop.Name,
			DataType:         prop.DataType,
			Description:      prop.Description,
			NestedProperties: a.convertNestedPropertiesFromWeaviate(prop.NestedProperties),
			JSONSchema:       prop.JSONSchema,
		}
	}
	return result
}

// wrapError wraps Weaviate errors in vectordb errors
func (a *Adapter) wrapError(err error, operation string) error {
	if err == nil {
		return nil
	}

	// Try to categorize the error based on the message
	errMsg := err.Error()

	// Connection errors
	if fmt.Sprintf("%v", err) == "connection refused" ||
		fmt.Sprintf("%v", err) == "no such host" ||
		fmt.Sprintf("%v", err) == "timeout" {
		return vectordb.ErrConnectionFailed(fmt.Sprintf("%s failed", operation), err)
	}

	// Authentication errors
	if fmt.Sprintf("%v", err) == "unauthorized" ||
		fmt.Sprintf("%v", err) == "forbidden" {
		return vectordb.ErrAuthenticationFailed(fmt.Sprintf("%s failed: %s", operation, errMsg))
	}

	// Not found errors
	if fmt.Sprintf("%v", err) == "not found" {
		return vectordb.ErrNotFound("resource", "unknown")
	}

	// Default to internal error
	return vectordb.ErrInternal(fmt.Sprintf("%s failed", operation), err)
}
