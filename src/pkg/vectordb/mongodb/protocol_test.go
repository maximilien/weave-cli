// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package mongodb

import (
	"context"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

func runMongoMock(t *testing.T, name string, test func(*mtest.T, *Adapter)) {
	t.Helper()
	mt := mtest.New(t)
	mt.RunOpts(name, mtest.NewOptions().ClientType(mtest.Mock).DatabaseName("weave"), func(mt *mtest.T) {
		client := &Client{
			client:   mt.Client,
			database: mt.DB,
			config: &Config{
				URI:              "mongodb://fixture",
				Database:         "weave",
				Timeout:          1,
				VectorDimensions: 1536,
				SimilarityMetric: "cosine",
			},
		}
		test(mt, &Adapter{Client: client})
	})
}

func mongoFixtureDocument() bson.D {
	return bson.D{
		{Key: "document_id", Value: "doc-1"},
		{Key: "text", Value: "hello world"},
		{Key: "content", Value: "full content"},
		{Key: "url", Value: "https://example.com"},
		{Key: "metadata", Value: bson.D{{Key: "kind", Value: "guide"}}},
		{Key: "textScore", Value: 2.5},
	}
}

func TestMongoDocumentProtocolOperations(t *testing.T) {
	runMongoMock(t, "document operations", func(mt *mtest.T, adapter *Adapter) {
		ctx := context.Background()
		mt.AddMockResponses(
			mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 1}),
			mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 2}),
			mtest.CreateCursorResponse(0, "weave.docs", mtest.FirstBatch, mongoFixtureDocument()),
			mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 1}, bson.E{Key: "nModified", Value: 1}),
			mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 1}),
			mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 2}),
			mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 2}),
			mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 3}),
			mtest.CreateCursorResponse(0, "weave.docs", mtest.FirstBatch, mongoFixtureDocument()),
		)

		doc := &vectordb.Document{ID: "doc-1", Text: "hello", Metadata: map[string]interface{}{"kind": "guide"}}
		if err := adapter.CreateDocument(ctx, "docs", doc); err != nil {
			mt.Fatalf("CreateDocument() error = %v", err)
		}
		if err := adapter.CreateDocuments(ctx, "docs", []*vectordb.Document{doc, {ID: "doc-2", Text: "world"}}); err != nil {
			mt.Fatalf("CreateDocuments() error = %v", err)
		}
		if err := adapter.CreateDocuments(ctx, "docs", nil); err != nil {
			mt.Fatalf("CreateDocuments(empty) error = %v", err)
		}

		got, err := adapter.GetDocument(ctx, "docs", "doc-1")
		if err != nil || got.ID != "doc-1" || got.Content != "full content" || got.Metadata["kind"] != "guide" {
			mt.Fatalf("GetDocument() = %#v, %v", got, err)
		}
		if err := adapter.UpdateDocument(ctx, "docs", doc); err != nil {
			mt.Fatalf("UpdateDocument() error = %v", err)
		}
		if err := adapter.DeleteDocument(ctx, "docs", "doc-1"); err != nil {
			mt.Fatalf("DeleteDocument() error = %v", err)
		}
		if err := adapter.DeleteDocuments(ctx, "docs", []string{"doc-1", "doc-2"}); err != nil {
			mt.Fatalf("DeleteDocuments() error = %v", err)
		}
		if err := adapter.DeleteDocuments(ctx, "docs", nil); err != nil {
			mt.Fatalf("DeleteDocuments(empty) error = %v", err)
		}
		if err := adapter.DeleteDocumentsByMetadata(ctx, "docs", map[string]interface{}{"kind": "guide"}); err != nil {
			mt.Fatalf("DeleteDocumentsByMetadata() error = %v", err)
		}
		if err := adapter.DeleteAllDocuments(ctx, "docs"); err != nil {
			mt.Fatalf("DeleteAllDocuments() error = %v", err)
		}

		docs, err := adapter.ListDocuments(ctx, "docs", 5, 2)
		if err != nil || len(docs) != 1 || docs[0].ID != "doc-1" {
			mt.Fatalf("ListDocuments() = %#v, %v", docs, err)
		}
	})
}

func TestMongoSearchProtocolOperations(t *testing.T) {
	runMongoMock(t, "search operations", func(mt *mtest.T, adapter *Adapter) {
		ctx := context.Background()
		mt.AddMockResponses(
			mtest.CreateCursorResponse(0, "weave.docs", mtest.FirstBatch, mongoFixtureDocument()),
			mtest.CreateCursorResponse(0, "weave.docs", mtest.FirstBatch, mongoFixtureDocument()),
			mtest.CreateCursorResponse(0, "weave.docs", mtest.FirstBatch, mongoFixtureDocument()),
		)

		opts := &vectordb.QueryOptions{TopK: 4}
		bm25, err := adapter.SearchBM25(ctx, "docs", "hello", opts)
		if err != nil || len(bm25) != 1 || bm25[0].Document.ID != "doc-1" || bm25[0].Score != 2.5 {
			mt.Fatalf("SearchBM25() = %#v, %v", bm25, err)
		}
		hybrid, err := adapter.SearchHybrid(ctx, "docs", "hello", opts)
		if err != nil || len(hybrid) != 1 || hybrid[0].Document.Text != "hello world" {
			mt.Fatalf("SearchHybrid() = %#v, %v", hybrid, err)
		}
		metadata, err := adapter.SearchByMetadata(ctx, "docs", map[string]interface{}{"kind": "guide"}, opts)
		if err != nil || len(metadata) != 1 || metadata[0].Score != 1 {
			mt.Fatalf("SearchByMetadata() = %#v, %v", metadata, err)
		}

		started := mt.GetAllStartedEvents()
		if len(started) != 3 || started[0].CommandName != "aggregate" || started[2].CommandName != "find" {
			mt.Fatalf("commands = %#v", started)
		}
	})
}

func TestMongoCollectionAndSchemaProtocolOperations(t *testing.T) {
	runMongoMock(t, "collection operations", func(mt *mtest.T, adapter *Adapter) {
		ctx := context.Background()
		mt.AddMockResponses(
			mtest.CreateSuccessResponse(),
			mtest.CreateSuccessResponse(),
			mtest.CreateSuccessResponse(),
			mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 1}, bson.E{Key: "nModified", Value: 1}),
			mtest.CreateCursorResponse(0, "weave.$cmd.listCollections", mtest.FirstBatch, bson.D{{Key: "name", Value: "docs"}}),
			mtest.CreateCursorResponse(0, "weave.docs", mtest.FirstBatch, bson.D{{Key: "n", Value: 2}}),
			mtest.CreateCursorResponse(0, "weave.docs", mtest.FirstBatch, bson.D{{Key: "document_id", Value: "_weave_metadata"}, {Key: "vector_dimensions", Value: 1536}}),
			mtest.CreateSuccessResponse(),
		)

		if err := adapter.CreateCollection(ctx, "docs", adapter.GetDefaultSchema(vectordb.SchemaTypeText, "docs")); err != nil {
			mt.Fatalf("CreateCollection() error = %v", err)
		}
		exists, err := adapter.CollectionExists(ctx, "docs")
		if err != nil || !exists {
			mt.Fatalf("CollectionExists() = %v, %v", exists, err)
		}
		count, err := adapter.GetCollectionCount(ctx, "docs")
		if err != nil || count != 2 {
			mt.Fatalf("GetCollectionCount() = %d, %v", count, err)
		}
		schema, err := adapter.GetSchema(ctx, "docs")
		if err != nil || schema.Class != "docs" || schema.Vectorizer != "text-embedding-3-small" {
			mt.Fatalf("GetSchema() = %#v, %v", schema, err)
		}
		if err := adapter.UpdateSchema(ctx, "docs", schema); err != nil {
			mt.Fatalf("UpdateSchema() error = %v", err)
		}
		if err := adapter.DeleteCollection(ctx, "docs"); err != nil {
			mt.Fatalf("DeleteCollection() error = %v", err)
		}
	})
}

func TestMongoProtocolErrors(t *testing.T) {
	tests := []struct {
		name string
		call func(*Adapter) error
		want string
	}{
		{"create", func(a *Adapter) error {
			return a.CreateDocument(context.Background(), "docs", &vectordb.Document{ID: "x"})
		}, "failed to create document"},
		{"get", func(a *Adapter) error { _, err := a.GetDocument(context.Background(), "docs", "x"); return err }, "failed to get document"},
		{"list", func(a *Adapter) error { _, err := a.ListDocuments(context.Background(), "docs", 1, 0); return err }, "failed to list documents"},
		{"BM25", func(a *Adapter) error {
			_, err := a.SearchBM25(context.Background(), "docs", "x", &vectordb.QueryOptions{TopK: 1})
			return err
		}, "BM25 search failed"},
	}
	for _, tt := range tests {
		runMongoMock(t, tt.name, func(mt *mtest.T, adapter *Adapter) {
			mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{Code: 42, Message: "fixture failure"}))
			err := tt.call(adapter)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				mt.Fatalf("error = %v, want substring %q", err, tt.want)
			}
		})
	}
}
