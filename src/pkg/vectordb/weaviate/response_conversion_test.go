// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package weaviate

import (
	"context"
	"errors"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

func TestParseQueryResults(t *testing.T) {
	client := &Client{}
	result := map[string]interface{}{
		"Get": map[string]interface{}{
			"Documents": []interface{}{
				map[string]interface{}{
					"content":     "certain",
					"metadata":    map[string]interface{}{"source": "test"},
					"image_data":  "encoded",
					"_additional": map[string]interface{}{"id": "one", "certainty": 0.8},
				},
				map[string]interface{}{
					"content":     "distant",
					"_additional": map[string]interface{}{"id": "two", "distance": 0.4},
				},
				map[string]interface{}{
					"content":     "too distant",
					"_additional": map[string]interface{}{"id": "three", "distance": 3.0},
				},
				map[string]interface{}{
					"content":     "ranked",
					"_additional": map[string]interface{}{"id": "four", "score": 4.0},
				},
				map[string]interface{}{
					"content":     "default",
					"_additional": map[string]interface{}{"id": "five"},
				},
				"skip non-map",
				map[string]interface{}{"content": "skip without additional"},
			},
		},
	}

	results, err := client.parseQueryResults(result, "content")
	if err != nil {
		t.Fatalf("parseQueryResults() error = %v", err)
	}
	if len(results) != 5 {
		t.Fatalf("len(results) = %d, want 5", len(results))
	}
	if results[0].ID != "one" || results[0].Content != "certain" || math.Abs(results[0].Score-0.64) > 0.0001 {
		t.Fatalf("certainty result = %#v", results[0])
	}
	if results[0].Metadata["image_base64"] != "encoded" || results[0].Metadata["source"] != "test" {
		t.Fatalf("image metadata = %#v", results[0].Metadata)
	}
	if math.Abs(results[1].Score-0.64) > 0.0001 || results[2].Score != 0 || results[3].Score != 1 || results[4].Score != 0.25 {
		t.Fatalf("scores = %v, %v, %v, %v", results[1].Score, results[2].Score, results[3].Score, results[4].Score)
	}
}

func TestParseQueryResultsReflectionAndErrors(t *testing.T) {
	type response struct {
		Data map[string]interface{}
	}
	client := &Client{}
	reflected := &response{Data: map[string]interface{}{
		"Get": map[string]interface{}{
			"Docs": []interface{}{map[string]interface{}{
				"text":        "hello",
				"image_data":  "encoded",
				"_additional": map[string]interface{}{"id": "doc"},
			}},
		},
	}}
	results, err := client.parseQueryResults(reflected, "text")
	if err != nil || len(results) != 1 || results[0].Metadata["image_base64"] != "encoded" {
		t.Fatalf("parseQueryResults(reflection) = %#v, %v", results, err)
	}

	tests := []struct {
		name   string
		result interface{}
		want   string
	}{
		{name: "nil", want: "received nil result"},
		{name: "invalid struct", result: &struct{ Other string }{}, want: "invalid result format"},
		{name: "missing Get", result: map[string]interface{}{}, want: "missing Get data"},
		{name: "no collection array", result: map[string]interface{}{"Get": map[string]interface{}{"Docs": "bad"}}, want: "no results found"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.parseQueryResults(tt.result, "text")
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func TestQueryResultHelpers(t *testing.T) {
	for _, tt := range []struct {
		raw  float64
		want float64
	}{{-1, 0}, {0.5, 0.25}, {2, 1}} {
		if got := normalizeScore(tt.raw); got != tt.want {
			t.Fatalf("normalizeScore(%v) = %v, want %v", tt.raw, got, tt.want)
		}
	}

	type response struct{ Errors []string }
	if !hasGraphQLErrors(nil) {
		t.Fatal("hasGraphQLErrors(nil) = false")
	}
	if hasGraphQLErrors(&response{}) {
		t.Fatal("hasGraphQLErrors(empty) = true")
	}
	if !hasGraphQLErrors(&response{Errors: []string{"failed"}}) {
		t.Fatal("hasGraphQLErrors(errors) = false")
	}
	if hasGraphQLErrors(&struct{ Value string }{}) {
		t.Fatal("hasGraphQLErrors(no field) = true")
	}

	type objectMap map[string]interface{}
	converted := convertJSONObjectToMap(objectMap{"one": 1, "two": "second"})
	if converted["one"] != 1 || converted["two"] != "second" {
		t.Fatalf("convertJSONObjectToMap() = %#v", converted)
	}
	if got := convertJSONObjectToMap("not a map"); len(got) != 0 {
		t.Fatalf("convertJSONObjectToMap(non-map) = %#v", got)
	}
}

func TestAdapterConversions(t *testing.T) {
	adapter := &Adapter{}
	if adapter.convertDocument(nil) != nil || adapter.convertDocumentFromWeaviate(nil) != nil {
		t.Fatal("nil document conversion returned a value")
	}
	doc := &vectordb.Document{
		ID: "one", Text: "text", Content: "content", Image: "image", ImageData: "data",
		URL: "url", Metadata: map[string]interface{}{"key": "value"},
	}
	converted := adapter.convertDocument(doc)
	converted.Embedding = []float64{1, 2}
	roundTrip := adapter.convertDocumentFromWeaviate(converted)
	if roundTrip.ID != doc.ID || roundTrip.ImageData != doc.ImageData || len(roundTrip.Embedding) != 2 {
		t.Fatalf("document round trip = %#v", roundTrip)
	}

	schema := adapter.convertSchemaFromWeaviate(&CollectionSchema{
		Class: "Docs", Vectorizer: "none",
		Properties: []SchemaProperty{{
			Name: "metadata", DataType: []string{"object"}, Description: "metadata",
			JSONSchema:       map[string]interface{}{"type": "object"},
			NestedProperties: []SchemaProperty{{Name: "author", DataType: []string{"text"}, NestedProperties: []SchemaProperty{{Name: "name", DataType: []string{"text"}}}}},
		}},
	})
	if schema.Class != "Docs" || schema.Properties[0].NestedProperties[0].Name != "author" || schema.Properties[0].NestedProperties[0].NestedProperties[0].Name != "name" {
		t.Fatalf("schema conversion = %#v", schema)
	}
	if adapter.convertNestedPropertiesFromWeaviate(nil) != nil {
		t.Fatal("nil nested properties returned a value")
	}

	for _, tt := range []struct {
		message string
		code    vectordb.ErrorType
	}{{"connection refused", vectordb.ErrorTypeConnection}, {"no such host", vectordb.ErrorTypeConnection}, {"timeout", vectordb.ErrorTypeConnection}, {"unauthorized", vectordb.ErrorTypeAuthentication}, {"forbidden", vectordb.ErrorTypeAuthentication}, {"not found", vectordb.ErrorTypeNotFound}, {"other", vectordb.ErrorTypeInternal}} {
		err := adapter.wrapError(errors.New(tt.message), "query")
		var vdbErr *vectordb.VectorDBError
		if !errors.As(err, &vdbErr) || vdbErr.Type != tt.code {
			t.Fatalf("wrapError(%q) = %#v", tt.message, err)
		}
	}
	if adapter.wrapError(nil, "query") != nil || adapter.Close() != nil {
		t.Fatal("nil error or Close returned an error")
	}
}

func TestClientHealthWithFakeHTTPServer(t *testing.T) {
	t.Run("healthy", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/v1/meta" {
				t.Errorf("path = %s", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"hostname":"fake","version":"1.27.0","modules":{}}`)
		}))
		t.Cleanup(server.Close)
		client, err := NewClient(&Config{URL: server.URL, Timeout: 1})
		if err != nil {
			t.Fatal(err)
		}
		if err := client.Health(context.Background()); err != nil {
			t.Fatalf("Health() error = %v", err)
		}
	})

	t.Run("authentication response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = io.WriteString(w, "Unauthorized")
		}))
		t.Cleanup(server.Close)
		client, err := NewClient(&Config{URL: server.URL, Timeout: 1})
		if err != nil {
			t.Fatal(err)
		}
		err = client.Health(context.Background())
		if err == nil || !strings.Contains(err.Error(), "Authentication error") {
			t.Fatalf("Health() error = %v", err)
		}
	})

	t.Run("connection refused", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
		url := server.URL
		server.Close()
		client, err := NewClient(&Config{URL: url, Timeout: 1})
		if err != nil {
			t.Fatal(err)
		}
		err = client.Health(context.Background())
		if err == nil || !strings.Contains(err.Error(), "Connection refused") {
			t.Fatalf("Health() error = %v", err)
		}
	})
}
