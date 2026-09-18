// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package pinecone

import (
	"context"
	"errors"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	pcsdk "github.com/pinecone-io/go-pinecone/pinecone"
	"google.golang.org/protobuf/types/known/structpb"
)

type fakeVectorExecutor struct {
	index       vectorIndex
	describeErr error
	indexErr    error
}

func (f *fakeVectorExecutor) DescribeIndex(context.Context, string) (*pcsdk.Index, error) {
	if f.describeErr != nil {
		return nil, f.describeErr
	}
	return &pcsdk.Index{Name: "docs", Host: "data.fixture", Dimension: 3}, nil
}

func (f *fakeVectorExecutor) Index(pcsdk.NewIndexConnParams) (vectorIndex, error) {
	if f.indexErr != nil {
		return nil, f.indexErr
	}
	return f.index, nil
}

type fakeVectorIndex struct {
	closed       int
	upserted     []*pcsdk.Vector
	deletedIDs   []string
	deleteFilter *pcsdk.MetadataFilter
	fetch        *pcsdk.FetchVectorsResponse
	list         *pcsdk.ListVectorsResponse
	query        *pcsdk.QueryVectorsResponse
	err          error
}

func (f *fakeVectorIndex) Close() error { f.closed++; return nil }
func (f *fakeVectorIndex) UpsertVectors(_ context.Context, vectors []*pcsdk.Vector) (uint32, error) {
	f.upserted = vectors
	return uint32(len(vectors)), f.err
}
func (f *fakeVectorIndex) FetchVectors(context.Context, []string) (*pcsdk.FetchVectorsResponse, error) {
	return f.fetch, f.err
}
func (f *fakeVectorIndex) ListVectors(context.Context, *pcsdk.ListVectorsRequest) (*pcsdk.ListVectorsResponse, error) {
	return f.list, f.err
}
func (f *fakeVectorIndex) QueryByVectorValues(context.Context, *pcsdk.QueryByVectorValuesRequest) (*pcsdk.QueryVectorsResponse, error) {
	return f.query, f.err
}
func (f *fakeVectorIndex) DeleteVectorsById(_ context.Context, ids []string) error {
	f.deletedIDs = append([]string(nil), ids...)
	return f.err
}
func (f *fakeVectorIndex) DeleteVectorsByFilter(_ context.Context, filter *pcsdk.MetadataFilter) error {
	f.deleteFilter = filter
	return f.err
}

func pineconeDataAdapter(index *fakeVectorIndex) *Adapter {
	return &Adapter{
		executor: &fakeVectorExecutor{index: index},
		config:   &vectordb.Config{Timeout: 1},
	}
}

func TestPineconeDocumentDataPlane(t *testing.T) {
	id1, id2 := "one", "two"
	meta1, _ := structpb.NewStruct(map[string]interface{}{"content": "first", "kind": "match"})
	meta2, _ := structpb.NewStruct(map[string]interface{}{"content": "second", "kind": "other"})
	index := &fakeVectorIndex{
		fetch: &pcsdk.FetchVectorsResponse{Vectors: map[string]*pcsdk.Vector{
			id1: {Id: id1, Metadata: meta1},
			id2: {Id: id2, Metadata: meta2},
		}},
		list: &pcsdk.ListVectorsResponse{VectorIds: []*string{&id1, nil, &id2}},
	}
	adapter := pineconeDataAdapter(index)
	ctx := context.Background()

	doc, err := adapter.GetDocument(ctx, "docs", id1)
	if err != nil || doc.ID != id1 || doc.Content != "first" || doc.Metadata["kind"] != "match" {
		t.Fatalf("GetDocument() = (%#v, %v)", doc, err)
	}
	docs, err := adapter.ListDocuments(ctx, "docs", 10, 0)
	if err != nil || len(docs) != 2 || docs[1].Content != "second" {
		t.Fatalf("ListDocuments() = (%#v, %v)", docs, err)
	}
	matched, err := adapter.GetDocumentsByMetadata(ctx, "docs", map[string]interface{}{"kind": "match"}, 1)
	if err != nil || len(matched) != 1 || matched[0].ID != id1 {
		t.Fatalf("GetDocumentsByMetadata() = (%#v, %v)", matched, err)
	}
	results, err := adapter.SearchByMetadata(ctx, "docs", map[string]interface{}{"kind": "match"}, &vectordb.QueryOptions{TopK: 1})
	if err != nil || len(results) != 1 || results[0].Score != 1 {
		t.Fatalf("SearchByMetadata() = (%#v, %v)", results, err)
	}

	if err := adapter.DeleteDocument(ctx, "docs", id1); err != nil {
		t.Fatalf("DeleteDocument() error = %v", err)
	}
	if err := adapter.DeleteDocuments(ctx, "docs", []string{id1, id2}); err != nil {
		t.Fatalf("DeleteDocuments() error = %v", err)
	}
	if len(index.deletedIDs) != 2 {
		t.Fatalf("deleted IDs = %#v", index.deletedIDs)
	}
	if err := adapter.DeleteDocuments(ctx, "docs", nil); err != nil {
		t.Fatalf("DeleteDocuments(empty) error = %v", err)
	}
	if err := adapter.DeleteDocumentsByMetadata(ctx, "docs", map[string]interface{}{"kind": "match"}); err != nil {
		t.Fatalf("DeleteDocumentsByMetadata() error = %v", err)
	}
	if index.deleteFilter == nil || index.deleteFilter.AsMap()["kind"].(map[string]interface{})["$eq"] != "match" {
		t.Fatalf("delete filter = %#v", index.deleteFilter)
	}

	if err := adapter.CreateDocument(ctx, "docs", &vectordb.Document{}); err == nil {
		t.Fatal("CreateDocument(empty) expected error")
	}
	if err := adapter.CreateDocument(ctx, "docs", &vectordb.Document{ID: "new", Content: "body"}); err == nil {
		t.Fatal("CreateDocument(no embedding client) expected error")
	}
	if err := adapter.UpdateDocument(ctx, "docs", &vectordb.Document{ID: "new", Text: "body"}); err == nil {
		t.Fatal("UpdateDocument(no embedding client) expected error")
	}
	if err := adapter.CreateDocuments(ctx, "docs", nil); err != nil {
		t.Fatalf("CreateDocuments(empty) error = %v", err)
	}
	if err := adapter.CreateDocuments(ctx, "docs", []*vectordb.Document{{ID: "new", Text: "body"}}); err == nil {
		t.Fatal("CreateDocuments(no embedding client) expected error")
	}
}

func TestPineconeDataPlaneErrorsAndEmptyResults(t *testing.T) {
	ctx := context.Background()
	index := &fakeVectorIndex{
		fetch: &pcsdk.FetchVectorsResponse{Vectors: map[string]*pcsdk.Vector{}},
		list:  &pcsdk.ListVectorsResponse{},
	}
	adapter := pineconeDataAdapter(index)
	if _, err := adapter.GetDocument(ctx, "docs", "missing"); err == nil {
		t.Fatal("GetDocument(missing) expected error")
	}
	index.fetch.Vectors["nil"] = nil
	if _, err := adapter.GetDocument(ctx, "docs", "nil"); err == nil {
		t.Fatal("GetDocument(nil vector) expected error")
	}
	if docs, err := adapter.ListDocuments(ctx, "docs", 10, 0); err != nil || len(docs) != 0 {
		t.Fatalf("ListDocuments(empty) = (%#v, %v)", docs, err)
	}

	boom := errors.New("fixture failure")
	for name, run := range map[string]func(*Adapter) error{
		"get":     func(a *Adapter) error { _, err := a.GetDocument(ctx, "docs", "id"); return err },
		"list":    func(a *Adapter) error { _, err := a.ListDocuments(ctx, "docs", 10, 0); return err },
		"delete":  func(a *Adapter) error { return a.DeleteDocument(ctx, "docs", "id") },
		"deletes": func(a *Adapter) error { return a.DeleteDocuments(ctx, "docs", []string{"id"}) },
		"metadata": func(a *Adapter) error {
			return a.DeleteDocumentsByMetadata(ctx, "docs", map[string]interface{}{"kind": "x"})
		},
	} {
		t.Run(name+" describe", func(t *testing.T) {
			a := pineconeDataAdapter(index)
			a.executor = &fakeVectorExecutor{describeErr: boom}
			if err := run(a); err == nil {
				t.Fatal("expected describe error")
			}
		})
		t.Run(name+" connect", func(t *testing.T) {
			a := pineconeDataAdapter(index)
			a.executor = &fakeVectorExecutor{indexErr: boom}
			if err := run(a); err == nil {
				t.Fatal("expected index connection error")
			}
		})
	}

	index.err = boom
	if _, err := adapter.GetDocument(ctx, "docs", "id"); err == nil {
		t.Fatal("GetDocument(fetch error) expected error")
	}
	if _, err := adapter.ListDocuments(ctx, "docs", 10, 0); err == nil {
		t.Fatal("ListDocuments(list error) expected error")
	}
	if err := adapter.DeleteDocument(ctx, "docs", "id"); err == nil {
		t.Fatal("DeleteDocument(delete error) expected error")
	}
	if err := adapter.DeleteDocumentsByMetadata(ctx, "docs", map[string]interface{}{"kind": "x"}); err == nil {
		t.Fatal("DeleteDocumentsByMetadata(delete error) expected error")
	}
}
