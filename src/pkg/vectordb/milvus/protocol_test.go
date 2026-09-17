// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package milvus

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

type milvusFixture struct {
	client.Client
	exists          bool
	collections     []*entity.Collection
	described       *entity.Collection
	queryResults    []client.ResultSet
	searchResults   []client.SearchResult
	fail            map[string]error
	createdSchema   *entity.Schema
	insertedColumns []entity.Column
	deletedIDs      []string
	closed          bool
}

func newMilvusAdapter(fixture *milvusFixture) *Adapter {
	return &Adapter{Client: &Client{
		client: fixture,
		config: &Config{
			Address:          "fixture:19530",
			Database:         "default",
			Timeout:          1,
			VectorDimensions: 3,
			SimilarityMetric: "COSINE",
		},
	}}
}

func (f *milvusFixture) failure(operation string) error {
	if f.fail == nil {
		return nil
	}
	return f.fail[operation]
}

func (f *milvusFixture) ListDatabases(context.Context) ([]entity.Database, error) {
	return []entity.Database{{Name: "default"}}, f.failure("ListDatabases")
}

func (f *milvusFixture) Close() error {
	f.closed = true
	return f.failure("Close")
}

func (f *milvusFixture) HasCollection(context.Context, string) (bool, error) {
	return f.exists, f.failure("HasCollection")
}

func (f *milvusFixture) CreateCollection(_ context.Context, schema *entity.Schema, _ int32, _ ...client.CreateCollectionOption) error {
	f.createdSchema = schema
	f.exists = true
	return f.failure("CreateCollection")
}

func (f *milvusFixture) CreateIndex(context.Context, string, string, entity.Index, bool, ...client.IndexOption) error {
	return f.failure("CreateIndex")
}

func (f *milvusFixture) LoadCollection(context.Context, string, bool, ...client.LoadCollectionOption) error {
	return f.failure("LoadCollection")
}

func (f *milvusFixture) DropCollection(context.Context, string, ...client.DropCollectionOption) error {
	f.exists = false
	return f.failure("DropCollection")
}

func (f *milvusFixture) ListCollections(context.Context, ...client.ListCollectionOption) ([]*entity.Collection, error) {
	return f.collections, f.failure("ListCollections")
}

func (f *milvusFixture) GetCollectionStatistics(context.Context, string) (map[string]string, error) {
	return map[string]string{"row_count": "2"}, f.failure("GetCollectionStatistics")
}

func (f *milvusFixture) DescribeCollection(context.Context, string) (*entity.Collection, error) {
	if err := f.failure("DescribeCollection"); err != nil {
		return nil, err
	}
	if f.described != nil {
		return f.described, nil
	}
	return fixtureCollection("docs", "text-embedding-3-small", 3), nil
}

func (f *milvusFixture) DescribeIndex(context.Context, string, string, ...client.IndexOption) ([]entity.Index, error) {
	if err := f.failure("DescribeIndex"); err != nil {
		return nil, err
	}
	index, err := entity.NewIndexIvfFlat(entity.COSINE, 128)
	if err != nil {
		return nil, err
	}
	return []entity.Index{index}, nil
}

func (f *milvusFixture) Insert(_ context.Context, _ string, _ string, columns ...entity.Column) (entity.Column, error) {
	f.insertedColumns = columns
	return entity.NewColumnVarChar(FieldDocumentID, []string{"inserted"}), f.failure("Insert")
}

func (f *milvusFixture) Flush(context.Context, string, bool, ...client.FlushOption) error {
	return f.failure("Flush")
}

func (f *milvusFixture) DeleteByPks(_ context.Context, _ string, _ string, ids entity.Column) error {
	if values, ok := ids.(*entity.ColumnVarChar); ok {
		f.deletedIDs = append([]string(nil), values.Data()...)
	}
	return f.failure("DeleteByPks")
}

func (f *milvusFixture) Query(context.Context, string, []string, string, []string, ...client.SearchQueryOptionFunc) (client.ResultSet, error) {
	if err := f.failure("Query"); err != nil {
		return nil, err
	}
	if len(f.queryResults) == 0 {
		return nil, nil
	}
	result := f.queryResults[0]
	f.queryResults = f.queryResults[1:]
	return result, nil
}

func (f *milvusFixture) Search(context.Context, string, []string, string, []string, []entity.Vector, string, entity.MetricType, int, entity.SearchParam, ...client.SearchQueryOptionFunc) ([]client.SearchResult, error) {
	return f.searchResults, f.failure("Search")
}

func fixtureCollection(name, vectorizer string, dimensions int) *entity.Collection {
	description := "fixture collection"
	if vectorizer != "" {
		description += " | vectorizer=" + vectorizer
	}
	return &entity.Collection{Name: name, Schema: &entity.Schema{
		CollectionName: name,
		Description:    description,
		Fields: []*entity.Field{entity.NewField().
			WithName(FieldEmbedding).
			WithDataType(entity.FieldTypeFloatVector).
			WithDim(int64(dimensions))},
	}}
}

func fixtureResultSet(ids ...string) client.ResultSet {
	texts := make([]string, len(ids))
	contents := make([]string, len(ids))
	images := make([]string, len(ids))
	imageData := make([][]byte, len(ids))
	imageURLs := make([]string, len(ids))
	urls := make([]string, len(ids))
	metadata := make([][]byte, len(ids))
	embeddings := make([][]float32, len(ids))
	for i, id := range ids {
		texts[i] = "text " + id
		contents[i] = "content " + id
		imageData[i] = []byte(`{"data":"base64-` + id + `"}`)
		imageURLs[i] = "https://images.example/" + id
		urls[i] = "https://example/" + id
		metadata[i] = []byte(`{"kind":"guide"}`)
		embeddings[i] = []float32{1, 2, 3}
	}
	return client.ResultSet{
		entity.NewColumnVarChar(FieldDocumentID, ids),
		entity.NewColumnVarChar(FieldText, texts),
		entity.NewColumnVarChar(FieldContent, contents),
		entity.NewColumnVarChar(FieldImage, images),
		entity.NewColumnJSONBytes(FieldImageData, imageData),
		entity.NewColumnVarChar(FieldImageURL, imageURLs),
		entity.NewColumnVarChar(FieldURL, urls),
		entity.NewColumnJSONBytes(FieldMetadata, metadata),
		entity.NewColumnFloatVector(FieldEmbedding, 3, embeddings),
	}
}

func TestMilvusCollectionProtocolOperations(t *testing.T) {
	fixture := &milvusFixture{}
	adapter := newMilvusAdapter(fixture)
	ctx := context.Background()

	if err := adapter.Health(ctx); err != nil {
		t.Fatalf("Health() error = %v", err)
	}
	schema := &vectordb.CollectionSchema{Vectorizer: "sentence-transformers/all-MiniLM-L6-v2"}
	if err := adapter.CreateCollection(ctx, "docs", schema); err != nil {
		t.Fatalf("CreateCollection() error = %v", err)
	}
	if fixture.createdSchema == nil || fixture.createdSchema.CollectionName != "docs" || !strings.Contains(fixture.createdSchema.Description, schema.Vectorizer) {
		t.Fatalf("created schema = %#v", fixture.createdSchema)
	}
	vectorField := fixture.createdSchema.Fields[6]
	if vectorField.TypeParams["dim"] != "384" {
		t.Fatalf("vector dimensions = %q", vectorField.TypeParams["dim"])
	}

	exists, err := adapter.CollectionExists(ctx, "docs")
	if err != nil || !exists {
		t.Fatalf("CollectionExists() = %v, %v", exists, err)
	}
	fixture.collections = []*entity.Collection{fixtureCollection("docs", "text-embedding-3-small", 1536)}
	collections, err := adapter.ListCollections(ctx)
	if err != nil || len(collections) != 1 || collections[0].Count != 2 || collections[0].Vectorizer != "text-embedding-3-small" {
		t.Fatalf("ListCollections() = %#v, %v", collections, err)
	}
	count, err := adapter.GetCollectionCount(ctx, "docs")
	if err != nil || count != 2 {
		t.Fatalf("GetCollectionCount() = %d, %v", count, err)
	}
	gotSchema, err := adapter.GetSchema(ctx, "docs")
	if err != nil || gotSchema.Vectorizer != "text-embedding-3-small" {
		t.Fatalf("GetSchema() = %#v, %v", gotSchema, err)
	}
	if err := adapter.DeleteCollection(ctx, "docs"); err != nil {
		t.Fatalf("DeleteCollection() error = %v", err)
	}
	if fixture.exists {
		t.Fatal("DeleteCollection() did not drop collection")
	}
	if err := adapter.Close(); err != nil || !fixture.closed {
		t.Fatalf("Close() = %v, closed=%v", err, fixture.closed)
	}
}

func TestMilvusDocumentProtocolOperations(t *testing.T) {
	fixture := &milvusFixture{}
	adapter := newMilvusAdapter(fixture)
	ctx := context.Background()

	doc := &vectordb.Document{ID: "doc-1", Text: "hello", Metadata: map[string]interface{}{"kind": "guide"}}
	if err := adapter.CreateDocument(ctx, "docs", doc); err != nil {
		t.Fatalf("CreateDocument() error = %v", err)
	}
	if len(fixture.insertedColumns) != 13 {
		t.Fatalf("CreateDocument() inserted %d columns", len(fixture.insertedColumns))
	}
	if vector := fixture.insertedColumns[6].(*entity.ColumnFloatVector).Data()[0]; len(vector) != 1536 {
		t.Fatalf("generated fallback vector length = %d", len(vector))
	}
	if err := adapter.CreateDocuments(ctx, "docs", []*vectordb.Document{doc, {ID: "doc-2", Content: "world"}}); err != nil {
		t.Fatalf("CreateDocuments() error = %v", err)
	}
	if err := adapter.CreateDocuments(ctx, "docs", nil); err != nil {
		t.Fatalf("CreateDocuments(empty) error = %v", err)
	}

	fixture.queryResults = []client.ResultSet{fixtureResultSet("doc-1")}
	got, err := adapter.GetDocument(ctx, "docs", "doc-1")
	if err != nil || got.ID != "doc-1" || got.Metadata["kind"] != "guide" || got.Metadata["image_url"] == nil {
		t.Fatalf("GetDocument() = %#v, %v", got, err)
	}
	if err := adapter.DeleteDocument(ctx, "docs", "doc-1"); err != nil {
		t.Fatalf("DeleteDocument() error = %v", err)
	}
	if err := adapter.DeleteDocuments(ctx, "docs", []string{"doc-1", "doc-2"}); err != nil {
		t.Fatalf("DeleteDocuments() error = %v", err)
	}
	if err := adapter.DeleteDocuments(ctx, "docs", nil); err != nil {
		t.Fatalf("DeleteDocuments(empty) error = %v", err)
	}
	fixture.queryResults = []client.ResultSet{fixtureResultSet("doc-1", "doc-2")}
	if err := adapter.DeleteDocumentsByMetadata(ctx, "docs", map[string]interface{}{"kind": "guide"}); err != nil {
		t.Fatalf("DeleteDocumentsByMetadata() error = %v", err)
	}
	if len(fixture.deletedIDs) != 2 {
		t.Fatalf("deleted IDs = %v", fixture.deletedIDs)
	}
	if err := adapter.DeleteDocumentsByMetadata(ctx, "docs", nil); err == nil {
		t.Fatal("DeleteDocumentsByMetadata() expected empty-filter error")
	}

	fixture.queryResults = []client.ResultSet{fixtureResultSet("doc-1", "doc-2")}
	docs, err := adapter.ListDocuments(ctx, "docs", 0, 1)
	if err != nil || len(docs) != 2 || len(docs[0].Embedding) != 3 || docs[0].Metadata["image_base64"] == nil {
		t.Fatalf("ListDocuments() = %#v, %v", docs, err)
	}
}

func TestMilvusQueryProtocolOperations(t *testing.T) {
	resultSet := fixtureResultSet("doc-1")
	fixture := &milvusFixture{
		queryResults:  []client.ResultSet{resultSet, resultSet},
		searchResults: []client.SearchResult{{ResultCount: 1, Fields: resultSet, Scores: []float32{0.75}}},
	}
	adapter := newMilvusAdapter(fixture)
	ctx := context.Background()
	opts := &vectordb.QueryOptions{TopK: 3}

	bm25, err := adapter.SearchBM25(ctx, "docs", "hello", opts)
	if err != nil || len(bm25) != 1 || bm25[0].Score != 1 {
		t.Fatalf("SearchBM25() = %#v, %v", bm25, err)
	}
	metadata, err := adapter.SearchByMetadata(ctx, "docs", map[string]interface{}{"kind": "guide"}, opts)
	if err != nil || len(metadata) != 1 || metadata[0].Document.ID != "doc-1" {
		t.Fatalf("SearchByMetadata() = %#v, %v", metadata, err)
	}
	parsed, err := adapter.parseSearchResults(fixture.searchResults)
	if err != nil || len(parsed) != 1 || parsed[0].Score != 0.75 || parsed[0].Document.Content != "content doc-1" {
		t.Fatalf("parseSearchResults() = %#v, %v", parsed, err)
	}
	if empty, err := adapter.parseSearchResults(nil); err != nil || len(empty) != 0 {
		t.Fatalf("parseSearchResults(nil) = %#v, %v", empty, err)
	}
}

func TestMilvusProtocolErrors(t *testing.T) {
	tests := []struct {
		name string
		fail string
		call func(*Adapter) error
		want string
	}{
		{"health", "ListDatabases", func(a *Adapter) error { return a.Health(context.Background()) }, "fixture failure"},
		{"create collection", "HasCollection", func(a *Adapter) error { return a.CreateCollection(context.Background(), "docs", nil) }, "failed to check collection"},
		{"get document", "Query", func(a *Adapter) error { _, err := a.GetDocument(context.Background(), "docs", "x"); return err }, "failed to get document"},
		{"delete document", "DeleteByPks", func(a *Adapter) error { return a.DeleteDocument(context.Background(), "docs", "x") }, "failed to delete document"},
		{"BM25", "Query", func(a *Adapter) error {
			_, err := a.SearchBM25(context.Background(), "docs", "x", &vectordb.QueryOptions{TopK: 1})
			return err
		}, "BM25 search failed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := newMilvusAdapter(&milvusFixture{fail: map[string]error{tt.fail: errors.New("fixture failure")}})
			err := tt.call(adapter)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want substring %q", err, tt.want)
			}
		})
	}
}
