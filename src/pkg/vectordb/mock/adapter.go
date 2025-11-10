// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package mock

import (
	"context"

	"github.com/maximilien/weave-cli/src/pkg/config"
	mockClient "github.com/maximilien/weave-cli/src/pkg/mock"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// Adapter wraps the existing mock client to implement the VectorDBClient interface
type Adapter struct {
	client *mockClient.Client
	config *vectordb.Config
}

// NewAdapter creates a new mock adapter
func NewAdapter(vdbConfig *vectordb.Config) (*Adapter, error) {
	// Convert vectordb.Config to config.MockConfig
	mockConfig := &config.MockConfig{
		Enabled:            vdbConfig.Enabled,
		SimulateEmbeddings: vdbConfig.SimulateEmbeddings,
		EmbeddingDimension: vdbConfig.EmbeddingDimension,
		Collections:        []config.MockCollection{}, // Will be populated as needed
	}

	client := mockClient.NewClient(mockConfig)

	return &Adapter{
		client: client,
		config: vdbConfig,
	}, nil
}

// Health checks the health of the mock instance (always returns nil)
func (a *Adapter) Health(ctx context.Context) error {
	return nil
}

// convertDocument converts between vectordb.Document and mockClient.Document
func (a *Adapter) convertDocument(doc *vectordb.Document) *mockClient.Document {
	if doc == nil {
		return nil
	}
	return &mockClient.Document{
		ID:        doc.ID,
		Content:   doc.Content,
		Image:     doc.Image,
		ImageData: doc.ImageData,
		URL:       doc.URL,
		Metadata:  doc.Metadata,
	}
}

// convertDocumentFromMock converts from mockClient.Document to vectordb.Document
func (a *Adapter) convertDocumentFromMock(doc *mockClient.Document) *vectordb.Document {
	if doc == nil {
		return nil
	}
	return &vectordb.Document{
		ID:        doc.ID,
		Text:      doc.Content, // Mock uses Content field for text
		Content:   doc.Content,
		Image:     doc.Image,
		ImageData: doc.ImageData,
		URL:       doc.URL,
		Metadata:  doc.Metadata,
	}
}

// convertDocumentsFromMock converts a slice of mockClient.Document to vectordb.Document
func (a *Adapter) convertDocumentsFromMock(docs []*mockClient.Document) []*vectordb.Document {
	if docs == nil {
		return nil
	}
	result := make([]*vectordb.Document, len(docs))
	for i, doc := range docs {
		result[i] = a.convertDocumentFromMock(doc)
	}
	return result
}

// convertQueryOptions converts vectordb.QueryOptions to mock query parameters
func (a *Adapter) convertQueryOptions(opts *vectordb.QueryOptions) (int, float64) {
	if opts == nil {
		return 10, 0.0 // Default values
	}
	topK := opts.TopK
	if topK <= 0 {
		topK = 10
	}
	return topK, opts.Distance
}
