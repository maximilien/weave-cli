// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package qdrant

import (
	"context"
	"errors"
	"strings"
	"testing"

	qdrant "github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
)

type qdrantFixtureState struct {
	fail         map[string]error
	create       *qdrant.CreateCollection
	deleteCol    *qdrant.DeleteCollection
	upserts      []*qdrant.UpsertPoints
	deletes      []*qdrant.DeletePoints
	get          *qdrant.GetPoints
	scroll       *qdrant.ScrollPoints
	search       *qdrant.SearchPoints
	getResult    []*qdrant.RetrievedPoint
	scrollResult []*qdrant.RetrievedPoint
	searchResult []*qdrant.ScoredPoint
}

type qdrantControlFixture struct {
	qdrant.QdrantClient
	state *qdrantFixtureState
}

func (f *qdrantControlFixture) HealthCheck(context.Context, *qdrant.HealthCheckRequest, ...grpc.CallOption) (*qdrant.HealthCheckReply, error) {
	if err := f.state.fail["health"]; err != nil {
		return nil, err
	}
	return &qdrant.HealthCheckReply{}, nil
}

type qdrantCollectionsFixture struct {
	qdrant.CollectionsClient
	state *qdrantFixtureState
}

func (f *qdrantCollectionsFixture) Create(_ context.Context, req *qdrant.CreateCollection, _ ...grpc.CallOption) (*qdrant.CollectionOperationResponse, error) {
	f.state.create = req
	return &qdrant.CollectionOperationResponse{}, f.state.fail["createCollection"]
}

func (f *qdrantCollectionsFixture) Delete(_ context.Context, req *qdrant.DeleteCollection, _ ...grpc.CallOption) (*qdrant.CollectionOperationResponse, error) {
	f.state.deleteCol = req
	return &qdrant.CollectionOperationResponse{}, f.state.fail["deleteCollection"]
}

func (f *qdrantCollectionsFixture) List(context.Context, *qdrant.ListCollectionsRequest, ...grpc.CallOption) (*qdrant.ListCollectionsResponse, error) {
	if err := f.state.fail["listCollections"]; err != nil {
		return nil, err
	}
	return &qdrant.ListCollectionsResponse{Collections: []*qdrant.CollectionDescription{{Name: "docs"}, {Name: "notes"}}}, nil
}

func (f *qdrantCollectionsFixture) CollectionExists(context.Context, *qdrant.CollectionExistsRequest, ...grpc.CallOption) (*qdrant.CollectionExistsResponse, error) {
	if err := f.state.fail["collectionExists"]; err != nil {
		return nil, err
	}
	return &qdrant.CollectionExistsResponse{Result: &qdrant.CollectionExists{Exists: true}}, nil
}

func (f *qdrantCollectionsFixture) Get(context.Context, *qdrant.GetCollectionInfoRequest, ...grpc.CallOption) (*qdrant.GetCollectionInfoResponse, error) {
	if err := f.state.fail["getCollection"]; err != nil {
		return nil, err
	}
	count := uint64(2)
	return &qdrant.GetCollectionInfoResponse{Result: &qdrant.CollectionInfo{
		PointsCount: &count,
		Config: &qdrant.CollectionConfig{Params: &qdrant.CollectionParams{VectorsConfig: &qdrant.VectorsConfig{
			Config: &qdrant.VectorsConfig_Params{Params: &qdrant.VectorParams{Size: 3, Distance: qdrant.Distance_Cosine}},
		}}},
	}}, nil
}

type qdrantPointsFixture struct {
	qdrant.PointsClient
	state *qdrantFixtureState
}

func (f *qdrantPointsFixture) Upsert(_ context.Context, req *qdrant.UpsertPoints, _ ...grpc.CallOption) (*qdrant.PointsOperationResponse, error) {
	f.state.upserts = append(f.state.upserts, req)
	return &qdrant.PointsOperationResponse{}, f.state.fail["upsert"]
}

func (f *qdrantPointsFixture) Get(_ context.Context, req *qdrant.GetPoints, _ ...grpc.CallOption) (*qdrant.GetResponse, error) {
	f.state.get = req
	if err := f.state.fail["getPoint"]; err != nil {
		return nil, err
	}
	return &qdrant.GetResponse{Result: f.state.getResult}, nil
}

func (f *qdrantPointsFixture) Delete(_ context.Context, req *qdrant.DeletePoints, _ ...grpc.CallOption) (*qdrant.PointsOperationResponse, error) {
	f.state.deletes = append(f.state.deletes, req)
	return &qdrant.PointsOperationResponse{}, f.state.fail["deletePoint"]
}

func (f *qdrantPointsFixture) Scroll(_ context.Context, req *qdrant.ScrollPoints, _ ...grpc.CallOption) (*qdrant.ScrollResponse, error) {
	f.state.scroll = req
	if err := f.state.fail["scroll"]; err != nil {
		return nil, err
	}
	return &qdrant.ScrollResponse{Result: f.state.scrollResult}, nil
}

func (f *qdrantPointsFixture) Search(_ context.Context, req *qdrant.SearchPoints, _ ...grpc.CallOption) (*qdrant.SearchResponse, error) {
	f.state.search = req
	if err := f.state.fail["search"]; err != nil {
		return nil, err
	}
	return &qdrant.SearchResponse{Result: f.state.searchResult}, nil
}

func newQdrantFixtureClient(state *qdrantFixtureState) *Client {
	if state.fail == nil {
		state.fail = make(map[string]error)
	}
	return &Client{
		qdrantClient:      &qdrantControlFixture{state: state},
		collectionsClient: &qdrantCollectionsFixture{state: state},
		pointsClient:      &qdrantPointsFixture{state: state},
		config:            &Config{Host: "fixture", Port: 6334, Timeout: 1, VectorDimensions: 8, SimilarityMetric: "Dot"},
	}
}

func fixtureVectors(values ...float32) *qdrant.VectorsOutput {
	return &qdrant.VectorsOutput{VectorsOptions: &qdrant.VectorsOutput_Vector{Vector: &qdrant.VectorOutput{
		Vector: &qdrant.VectorOutput_Dense{Dense: &qdrant.DenseVector{Data: values}},
	}}}
}

func fixturePoint() *qdrant.RetrievedPoint {
	return &qdrant.RetrievedPoint{
		Id: &qdrant.PointId{PointIdOptions: &qdrant.PointId_Uuid{Uuid: "9b3bdce4-41f2-4fb6-8b8b-f47f1a6e240d"}},
		Payload: map[string]*qdrant.Value{
			"_original_id": {Kind: &qdrant.Value_StringValue{StringValue: "doc-1"}},
			"content":      {Kind: &qdrant.Value_StringValue{StringValue: "hello"}},
			"rank":         {Kind: &qdrant.Value_IntegerValue{IntegerValue: 7}},
		},
		Vectors: fixtureVectors(1, 2, 3),
	}
}

func fixtureScoredPoint() *qdrant.ScoredPoint {
	point := fixturePoint()
	return &qdrant.ScoredPoint{Id: point.Id, Payload: point.Payload, Vectors: point.Vectors, Score: 0.75}
}

func TestQdrantCollectionProtocolOperations(t *testing.T) {
	state := &qdrantFixtureState{}
	client := newQdrantFixtureClient(state)
	ctx := context.Background()

	if err := client.Health(ctx); err != nil {
		t.Fatalf("Health() error = %v", err)
	}
	client.SetCollection("docs")
	if got := client.GetCollection(); got != "docs" {
		t.Fatalf("GetCollection() = %q", got)
	}
	if err := client.CreateCollection(ctx, "docs", 3, "cosine"); err != nil {
		t.Fatalf("CreateCollection() error = %v", err)
	}
	params := state.create.GetVectorsConfig().GetParams()
	if state.create.GetCollectionName() != "docs" || params.GetSize() != 3 || params.GetDistance() != qdrant.Distance_Cosine {
		t.Fatalf("CreateCollection request = %#v", state.create)
	}
	collections, err := client.ListCollections(ctx)
	if err != nil || len(collections) != 2 || collections[1] != "notes" {
		t.Fatalf("ListCollections() = %v, %v", collections, err)
	}
	exists, err := client.CollectionExists(ctx, "docs")
	if err != nil || !exists {
		t.Fatalf("CollectionExists() = %v, %v", exists, err)
	}
	count, err := client.GetCollectionCount(ctx, "docs")
	if err != nil || count != 2 {
		t.Fatalf("GetCollectionCount() = %d, %v", count, err)
	}
	dimensions, _ := client.getCollectionDimensions(ctx, "docs")
	distance, _ := client.getCollectionDistance(ctx, "docs")
	if dimensions != 3 || distance != qdrant.Distance_Cosine {
		t.Fatalf("collection vector config = %d, %v", dimensions, distance)
	}
	if err := client.DeleteCollection(ctx, "docs"); err != nil || state.deleteCol.GetCollectionName() != "docs" {
		t.Fatalf("DeleteCollection() error = %v, request = %#v", err, state.deleteCol)
	}

	for _, metric := range []string{"Cosine", "dot", "Euclidean", "L2"} {
		if _, err := mapDistanceMetric(metric); err != nil {
			t.Errorf("mapDistanceMetric(%q) error = %v", metric, err)
		}
	}
	if _, err := mapDistanceMetric("taxicab"); err == nil {
		t.Fatal("mapDistanceMetric(taxicab) expected error")
	}
}

func TestQdrantDocumentProtocolOperations(t *testing.T) {
	point := fixturePoint()
	state := &qdrantFixtureState{getResult: []*qdrant.RetrievedPoint{point}, scrollResult: []*qdrant.RetrievedPoint{point}}
	client := newQdrantFixtureClient(state)
	ctx := context.Background()
	doc := &Document{ID: "doc-1", Vector: []float32{1, 2, 3}, Content: "hello", Metadata: map[string]interface{}{
		"string": "value", "int": 2, "int64": int64(3), "float32": float32(1.5), "float64": 2.5,
		"bool": true, "ints": []int{1, 2}, "strings": []string{"a", "b"}, "mixed": []interface{}{1, "two", false},
	}}
	if err := client.CreateDocument(ctx, "docs", doc); err != nil {
		t.Fatalf("CreateDocument() error = %v", err)
	}
	generated := &Document{Vector: []float32{4, 5, 6}}
	if err := client.CreateDocuments(ctx, "docs", []*Document{doc, generated}); err != nil {
		t.Fatalf("CreateDocuments() error = %v", err)
	}
	if generated.ID == "" || len(state.upserts) != 2 || len(state.upserts[1].GetPoints()) != 2 {
		t.Fatalf("batch upserts = %#v, generated ID = %q", state.upserts, generated.ID)
	}
	if err := client.CreateDocuments(ctx, "docs", nil); err != nil || len(state.upserts) != 2 {
		t.Fatalf("empty CreateDocuments() changed state: %v", err)
	}
	got, err := client.GetDocument(ctx, "docs", "doc-1")
	if err != nil || got.ID != "doc-1" || got.Content != "hello" || len(got.Vector) != 3 || got.Metadata["rank"] != int64(7) {
		t.Fatalf("GetDocument() = %#v, %v", got, err)
	}
	if err := client.UpdateDocument(ctx, "docs", doc); err != nil {
		t.Fatalf("UpdateDocument() error = %v", err)
	}
	if err := client.DeleteDocument(ctx, "docs", "doc-1"); err != nil {
		t.Fatalf("DeleteDocument() error = %v", err)
	}
	if err := client.DeleteDocuments(ctx, "docs", []string{"doc-1", "doc-2"}); err != nil {
		t.Fatalf("DeleteDocuments() error = %v", err)
	}
	deleteCount := len(state.deletes)
	if err := client.DeleteDocuments(ctx, "docs", nil); err != nil || len(state.deletes) != deleteCount {
		t.Fatalf("empty DeleteDocuments() changed state: %v", err)
	}
	docs, err := client.ListDocuments(ctx, "docs", 0)
	if err != nil || len(docs) != 1 || state.scroll.GetLimit() != 100 {
		t.Fatalf("ListDocuments() = %#v, %v, limit %d", docs, err, state.scroll.GetLimit())
	}
	if err := client.DeleteAllDocuments(ctx, "docs"); err != nil {
		t.Fatalf("DeleteAllDocuments() error = %v", err)
	}
	if len(state.deletes) != deleteCount+1 {
		t.Fatalf("delete requests = %d", len(state.deletes))
	}

	if _, err := documentToPoint(&Document{ID: "bad", Metadata: map[string]interface{}{"bad": make(chan int)}}); err == nil {
		t.Fatal("documentToPoint() expected unsupported metadata error")
	}
	if _, err := interfaceToValue([]interface{}{make(chan int)}); err == nil {
		t.Fatal("interfaceToValue() expected nested value error")
	}
	for _, value := range []*qdrant.Value{
		{Kind: &qdrant.Value_DoubleValue{DoubleValue: 2.5}},
		{Kind: &qdrant.Value_BoolValue{BoolValue: true}},
		{Kind: &qdrant.Value_ListValue{ListValue: nil}},
		{},
	} {
		_ = valueToInterface(value)
	}
	valid := "9b3bdce4-41f2-4fb6-8b8b-f47f1a6e240d"
	if got, _ := stringToUUID(valid); got != valid {
		t.Fatalf("stringToUUID(valid) = %q", got)
	}
}

func TestQdrantQueryProtocolOperations(t *testing.T) {
	point := fixturePoint()
	state := &qdrantFixtureState{
		scrollResult: []*qdrant.RetrievedPoint{point},
		searchResult: []*qdrant.ScoredPoint{fixtureScoredPoint()},
	}
	client := newQdrantFixtureClient(state)
	ctx := context.Background()

	filter, err := BuildFilter(map[string]interface{}{"kind": "guide", "rank": 3, "age": int64(4), "active": true})
	if err != nil || len(filter.Must) != 4 {
		t.Fatalf("BuildFilter() = %#v, %v", filter, err)
	}
	if _, err := BuildFilter(map[string]interface{}{"score": 1.5}); err == nil {
		t.Fatal("BuildFilter() expected unsupported value error")
	}
	results, err := client.SearchSemantic(ctx, "docs", []float32{1, 2, 3}, 0)
	if err != nil || len(results) != 1 || results[0].Score != 0.75 || state.search.GetLimit() != 10 {
		t.Fatalf("SearchSemantic() = %#v, %v", results, err)
	}
	docs, err := client.SearchByMetadata(ctx, "docs", filter, 0)
	if err != nil || len(docs) != 1 || state.scroll.GetLimit() != 100 {
		t.Fatalf("SearchByMetadata() = %#v, %v", docs, err)
	}
	results, err = client.SearchHybrid(ctx, "docs", []float32{1, 2, 3}, filter, 4)
	if err != nil || len(results) != 1 || state.search.GetFilter() != filter || state.search.GetLimit() != 4 {
		t.Fatalf("SearchHybrid() = %#v, %v", results, err)
	}
}

func TestQdrantProtocolErrorsAndFallbacks(t *testing.T) {
	ctx := context.Background()
	boom := errors.New("fixture failure")

	for _, tc := range []struct {
		name string
		op   string
		call func(*Client) error
	}{
		{"health connection", "health", func(c *Client) error { return c.Health(ctx) }},
		{"create collection", "createCollection", func(c *Client) error { return c.CreateCollection(ctx, "docs", 3, "cosine") }},
		{"delete collection", "deleteCollection", func(c *Client) error { return c.DeleteCollection(ctx, "docs") }},
		{"list collections", "listCollections", func(c *Client) error { _, err := c.ListCollections(ctx); return err }},
		{"collection exists", "collectionExists", func(c *Client) error { _, err := c.CollectionExists(ctx, "docs"); return err }},
		{"get point", "getPoint", func(c *Client) error { _, err := c.GetDocument(ctx, "docs", "id"); return err }},
		{"upsert", "upsert", func(c *Client) error { return c.CreateDocument(ctx, "docs", &Document{ID: "id"}) }},
		{"delete point", "deletePoint", func(c *Client) error { return c.DeleteDocument(ctx, "docs", "id") }},
		{"scroll", "scroll", func(c *Client) error { _, err := c.ListDocuments(ctx, "docs", 1); return err }},
		{"search", "search", func(c *Client) error { _, err := c.SearchSemantic(ctx, "docs", nil, 1); return err }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := &qdrantFixtureState{fail: map[string]error{tc.op: boom}}
			if err := tc.call(newQdrantFixtureClient(state)); err == nil {
				t.Fatal("expected error")
			}
		})
	}

	state := &qdrantFixtureState{}
	client := newQdrantFixtureClient(state)
	if _, err := client.GetDocument(ctx, "docs", "missing"); err == nil {
		t.Fatal("GetDocument() expected not found error")
	}
	state.fail["getCollection"] = boom
	dimensions, err := client.getCollectionDimensions(ctx, "docs")
	distance, distErr := client.getCollectionDistance(ctx, "docs")
	if err != nil || distErr != nil || dimensions != 8 || distance != qdrant.Distance_Dot {
		t.Fatalf("fallbacks = %d, %v, %v, %v", dimensions, distance, err, distErr)
	}

	for _, message := range []string{"connection refused", "deadline exceeded", "401 Unauthorized", "plain failure"} {
		state := &qdrantFixtureState{fail: map[string]error{"health": errors.New(message)}}
		err := newQdrantFixtureClient(state).Health(ctx)
		if err == nil || !strings.Contains(err.Error(), message) {
			t.Fatalf("Health(%q) error = %v", message, err)
		}
	}
}
