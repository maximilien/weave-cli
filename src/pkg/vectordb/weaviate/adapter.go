// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package weaviate

import (
	"context"
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	weaviateClient "github.com/maximilien/weave-cli/src/pkg/weaviate"
)

// Adapter wraps the existing Weaviate client to implement the VectorDBClient interface
type Adapter struct {
	client      *weaviateClient.Client
	weaveClient *weaviateClient.WeaveClient
	config      *vectordb.Config
}

// NewAdapter creates a new Weaviate adapter
func NewAdapter(config *vectordb.Config) (*Adapter, error) {
	// Convert vectordb.Config to weaviate.Config
	weaviateConfig := &weaviateClient.Config{
		URL:          config.URL,
		APIKey:       config.APIKey,
		OpenAIAPIKey: config.OpenAIAPIKey,
		Timeout:      config.Timeout,
	}

	// Create the original Weaviate client
	client, err := weaviateClient.NewClient(weaviateConfig)
	if err != nil {
		return nil, vectordb.ErrConnectionFailed("failed to create Weaviate client", err)
	}

	// Create the enhanced WeaveClient
	weaveClient, err := weaviateClient.NewWeaveClient(weaviateConfig)
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

// convertDocument converts between vectordb.Document and weaviate.Document
func (a *Adapter) convertDocument(doc *vectordb.Document) *weaviateClient.Document {
	if doc == nil {
		return nil
	}
	return &weaviateClient.Document{
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
func (a *Adapter) convertDocumentFromWeaviate(doc *weaviateClient.Document) *vectordb.Document {
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
	}
}

// convertDocuments converts a slice of vectordb.Document to weaviate.Document
func (a *Adapter) convertDocuments(docs []*vectordb.Document) []*weaviateClient.Document {
	if docs == nil {
		return nil
	}
	result := make([]*weaviateClient.Document, len(docs))
	for i, doc := range docs {
		result[i] = a.convertDocument(doc)
	}
	return result
}

// convertDocumentsFromWeaviate converts a slice of weaviate.Document to vectordb.Document
func (a *Adapter) convertDocumentsFromWeaviate(docs []*weaviateClient.Document) []*vectordb.Document {
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
func (a *Adapter) convertQueryOptions(opts *vectordb.QueryOptions) *weaviateClient.QueryOptions {
	if opts == nil {
		return &weaviateClient.QueryOptions{}
	}
	return &weaviateClient.QueryOptions{
		TopK:           opts.TopK,
		Distance:       opts.Distance,
		SearchMetadata: opts.SearchMetadata,
		NoTruncate:     opts.NoTruncate,
		UseBM25:        opts.UseBM25,
	}
}

// convertQueryResults converts weaviate query results to vectordb query results
func (a *Adapter) convertQueryResults(results []*weaviateClient.Document, scores []float64) []*vectordb.QueryResult {
	if results == nil {
		return nil
	}

	queryResults := make([]*vectordb.QueryResult, len(results))
	for i, doc := range results {
		score := 0.0
		if i < len(scores) {
			score = scores[i]
		}
		queryResults[i] = &vectordb.QueryResult{
			Document: *a.convertDocumentFromWeaviate(doc),
			Score:    score,
		}
	}
	return queryResults
}

// convertSchema converts between vectordb.CollectionSchema and weaviate.CollectionSchema
func (a *Adapter) convertSchema(schema *vectordb.CollectionSchema) *weaviateClient.CollectionSchema {
	if schema == nil {
		return nil
	}

	properties := make([]weaviateClient.SchemaProperty, len(schema.Properties))
	for i, prop := range schema.Properties {
		properties[i] = weaviateClient.SchemaProperty{
			Name:             prop.Name,
			DataType:         prop.DataType,
			Description:      prop.Description,
			NestedProperties: a.convertNestedProperties(prop.NestedProperties),
			JSONSchema:       prop.JSONSchema,
		}
	}

	return &weaviateClient.CollectionSchema{
		Class:      schema.Class,
		Vectorizer: schema.Vectorizer,
		Properties: properties,
	}
}

// convertSchemaFromWeaviate converts from weaviate.CollectionSchema to vectordb.CollectionSchema
func (a *Adapter) convertSchemaFromWeaviate(schema *weaviateClient.CollectionSchema) *vectordb.CollectionSchema {
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

// convertNestedProperties converts nested properties to Weaviate format
func (a *Adapter) convertNestedProperties(props []vectordb.SchemaProperty) []weaviateClient.SchemaProperty {
	if props == nil {
		return nil
	}

	result := make([]weaviateClient.SchemaProperty, len(props))
	for i, prop := range props {
		result[i] = weaviateClient.SchemaProperty{
			Name:             prop.Name,
			DataType:         prop.DataType,
			Description:      prop.Description,
			NestedProperties: a.convertNestedProperties(prop.NestedProperties),
			JSONSchema:       prop.JSONSchema,
		}
	}
	return result
}

// convertNestedPropertiesFromWeaviate converts nested properties from Weaviate format
func (a *Adapter) convertNestedPropertiesFromWeaviate(props []weaviateClient.SchemaProperty) []vectordb.SchemaProperty {
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
