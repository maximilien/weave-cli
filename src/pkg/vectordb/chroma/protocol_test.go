//go:build (darwin && amd64) || (darwin && arm64)

// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package chroma

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/stretchr/testify/require"
)

const (
	testCollectionName = "test-collection"
	testCollectionID   = "8ecf0f7e-e806-47f8-96a1-4732ef42359e"
)

type chromaRequest struct {
	method string
	path   string
	body   string
}

func newChromaProtocolServer(t *testing.T) (*httptest.Server, *[]chromaRequest) {
	t.Helper()

	requests := make([]chromaRequest, 0)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		requests = append(requests, chromaRequest{method: r.Method, path: r.URL.Path, body: string(body)})
		w.Header().Set("Content-Type", "application/json")

		collectionBase := "/api/v2/tenants/default_tenant/databases/default_database/collections"
		collectionModel := map[string]interface{}{
			"id":        testCollectionID,
			"name":      testCollectionName,
			"tenant":    "default_tenant",
			"database":  "default_database",
			"metadata":  map[string]interface{}{"vectorizer": "none"},
			"dimension": 3,
		}

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/heartbeat":
			_, err = io.WriteString(w, `{"nanosecond heartbeat":123}`)
		case r.Method == http.MethodGet && r.URL.Path == "/api/v2/pre-flight-checks":
			_, err = io.WriteString(w, `{"max_batch_size":100}`)
		case r.Method == http.MethodPost && r.URL.Path == collectionBase:
			err = json.NewEncoder(w).Encode(collectionModel)
		case r.Method == http.MethodGet && r.URL.Path == collectionBase:
			err = json.NewEncoder(w).Encode([]interface{}{collectionModel})
		case r.Method == http.MethodGet && r.URL.Path == collectionBase+"/"+testCollectionName:
			err = json.NewEncoder(w).Encode(collectionModel)
		case r.Method == http.MethodDelete && r.URL.Path == collectionBase+"/"+testCollectionName:
			_, err = io.WriteString(w, `true`)
		case r.Method == http.MethodGet && r.URL.Path == collectionBase+"/"+testCollectionID+"/count":
			_, err = io.WriteString(w, `3`)
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/add"):
			_, err = io.WriteString(w, `true`)
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/update"):
			_, err = io.WriteString(w, `true`)
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/delete"):
			_, err = io.WriteString(w, `true`)
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/get"):
			_, err = io.WriteString(w, `{"ids":["first","second","third"],"documents":["first body","second body","third body"],"metadatas":[{"url":"https://example.com/1","filename":"one.txt"},{"image":"image.png","type":"image"},null],"include":["documents","metadatas"]}`)
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/query"):
			_, err = io.WriteString(w, `{"ids":[["first","second"]],"documents":[["first body","second body"]],"metadatas":[[{"url":"https://example.com/1","filename":"one.txt"},{"image":"image.png","type":"image"}]],"distances":[[0.25,1]],"include":["documents","metadatas","distances"]}`)
		default:
			http.Error(w, "unexpected request", http.StatusNotFound)
			return
		}
		require.NoError(t, err)
	})

	server := httptest.NewServer(handler)
	return server, &requests
}

func newProtocolClient(t *testing.T, serverURL string) *Client {
	t.Helper()
	client, err := NewClient(&Config{URL: serverURL, VectorDimensions: 3, Timeout: 1})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.Close(context.Background())) })
	return client
}

func TestClientCollectionProtocols(t *testing.T) {
	server, requests := newChromaProtocolServer(t)
	defer server.Close()
	client := newProtocolClient(t, server.URL)
	ctx := context.Background()

	require.NoError(t, client.Health(ctx))
	require.NoError(t, client.CreateCollection(ctx, testCollectionName, nil))

	collections, err := client.ListCollections(ctx)
	require.NoError(t, err)
	require.Equal(t, []vectordb.CollectionInfo{{Name: testCollectionName, Count: 3, Vectorizer: "none"}}, collections)

	exists, err := client.CollectionExists(ctx, testCollectionName)
	require.NoError(t, err)
	require.True(t, exists)

	count, err := client.GetCollectionCount(ctx, testCollectionName)
	require.NoError(t, err)
	require.EqualValues(t, 3, count)

	require.NoError(t, client.DeleteCollection(ctx, testCollectionName))
	require.NotEmpty(t, *requests)
}

func TestClientDocumentProtocols(t *testing.T) {
	server, requests := newChromaProtocolServer(t)
	defer server.Close()
	client := newProtocolClient(t, server.URL)
	ctx := context.Background()

	doc := &vectordb.Document{
		ID:      "first",
		Content: "first body",
		URL:     "https://example.com/1",
		Image:   "image.png",
		Metadata: map[string]interface{}{
			"filename": "one.txt",
			"rank":     2,
			"ignored":  []string{"complex"},
		},
	}
	require.NoError(t, client.CreateDocument(ctx, testCollectionName, doc))

	fetched, err := client.GetDocument(ctx, testCollectionName, "first")
	require.NoError(t, err)
	require.Equal(t, "first", fetched.ID)
	require.Equal(t, "first body", fetched.Content)
	require.Equal(t, "https://example.com/1", fetched.URL)
	require.Equal(t, "one.txt", fetched.Metadata["filename"])

	doc.Content = "updated body"
	require.NoError(t, client.UpdateDocument(ctx, testCollectionName, doc))

	docs, err := client.ListDocuments(ctx, testCollectionName, 1, 1)
	require.NoError(t, err)
	require.Len(t, docs, 1)
	require.Equal(t, "second", docs[0].ID)
	require.Equal(t, "image.png", docs[0].Image)

	require.NoError(t, client.CreateDocuments(ctx, testCollectionName, []*vectordb.Document{
		{ID: "second", Text: "second body", Metadata: map[string]interface{}{"kind": "text"}},
		{ID: "third", Content: "third body"},
	}))
	require.NoError(t, client.CreateDocuments(ctx, testCollectionName, nil))
	require.NoError(t, client.DeleteDocument(ctx, testCollectionName, "first"))
	require.NoError(t, client.DeleteDocuments(ctx, testCollectionName, []string{"second", "third"}))
	require.NoError(t, client.DeleteDocuments(ctx, testCollectionName, nil))
	require.NoError(t, client.DeleteDocumentsByMetadata(ctx, testCollectionName, map[string]interface{}{
		"kind": "text", "rank": 2, "score": 1.5, "active": true,
	}))

	var sawFilteredCreate bool
	for _, request := range *requests {
		if strings.HasSuffix(request.path, "/add") && strings.Contains(request.body, `"ids":["first"]`) {
			sawFilteredCreate = !strings.Contains(request.body, "ignored")
		}
	}
	require.True(t, sawFilteredCreate)
}

func TestClientSearchProtocols(t *testing.T) {
	server, _ := newChromaProtocolServer(t)
	defer server.Close()
	client := newProtocolClient(t, server.URL)
	ctx := context.Background()

	results, err := client.SearchSemantic(ctx, testCollectionName, "query", &vectordb.QueryOptions{TopK: 2})
	require.NoError(t, err)
	require.Len(t, results, 2)
	require.Equal(t, "first", results[0].Document.ID)
	require.InDelta(t, 0.8, results[0].Score, 0.0001)
	require.Equal(t, "one.txt", results[0].Document.Metadata["filename"])

	hybrid, err := client.SearchHybrid(ctx, testCollectionName, "query", nil)
	require.NoError(t, err)
	require.Len(t, hybrid, 2)

	metadata, err := client.SearchByMetadata(ctx, testCollectionName, map[string]interface{}{
		"kind": "text", "rank": 2, "score": 1.5, "active": true,
	}, &vectordb.QueryOptions{TopK: 2})
	require.NoError(t, err)
	require.Len(t, metadata, 2)
	require.Equal(t, "https://example.com/1", metadata[0].Document.URL)

	_, err = client.SearchBM25(ctx, testCollectionName, "query", nil)
	require.ErrorContains(t, err, "not supported")
}

func TestClientSchemaBehavior(t *testing.T) {
	client := &Client{config: &Config{}}
	ctx := context.Background()

	schema, err := client.GetSchema(ctx, "documents")
	require.NoError(t, err)
	require.Equal(t, "documents", schema.Class)
	require.NoError(t, client.UpdateSchema(ctx, "documents", schema))

	textSchema := client.GetDefaultSchema(vectordb.SchemaTypeText, "text")
	imageSchema := client.GetDefaultSchema(vectordb.SchemaTypeImage, "images")
	defaultSchema := client.GetDefaultSchema(vectordb.SchemaType("unknown"), "fallback")
	require.Equal(t, "text", textSchema.Class)
	require.Equal(t, "images", imageSchema.Class)
	require.Equal(t, "fallback", defaultSchema.Class)
}
