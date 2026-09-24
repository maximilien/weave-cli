// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package utils

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/config"
)

func TestListMockCollectionsOutputs(t *testing.T) {
	ctx := context.Background()
	collectionConfig := &config.VectorDBConfig{
		Type:    config.VectorDBTypeMock,
		Enabled: true,
		Collections: []config.Collection{
			{Name: "WeaveImages", Type: "image"},
			{Name: "WeaveDocs", Type: "text"},
			{Name: "Empty", Type: "text"},
		},
	}

	textOutput := captureUtilsOutput(t, func() {
		ListMockCollections(ctx, collectionConfig, 0, true, false)
	})
	for _, want := range []string{"Empty", "WeaveDocs", "6 items", "WeaveImages", "3 items", "Virtual structure"} {
		if !strings.Contains(textOutput, want) {
			t.Errorf("text output missing %q: %s", want, textOutput)
		}
	}
	if strings.Index(textOutput, "Empty") > strings.Index(textOutput, "WeaveDocs") {
		t.Fatalf("collections were not sorted: %s", textOutput)
	}

	jsonOutput := captureUtilsOutput(t, func() {
		ListMockCollections(ctx, collectionConfig, 2, false, true)
	})
	var result struct {
		Collections []struct {
			Name          string `json:"name"`
			DocumentCount int    `json:"document_count"`
			IsImage       bool   `json:"is_image"`
		} `json:"collections"`
		Total int `json:"total"`
	}
	if err := json.Unmarshal([]byte(jsonOutput), &result); err != nil {
		t.Fatalf("invalid JSON output: %v: %q", err, jsonOutput)
	}
	if result.Total != 2 || len(result.Collections) != 2 || result.Collections[0].Name != "Empty" || result.Collections[1].Name != "WeaveDocs" || result.Collections[1].DocumentCount != 6 {
		t.Fatalf("unexpected limited JSON output: %+v", result)
	}

	emptyOutput := captureUtilsOutput(t, func() {
		ListMockCollections(ctx, &config.VectorDBConfig{Type: config.VectorDBTypeMock}, 0, false, true)
	})
	if !strings.Contains(emptyOutput, `"total": 0`) {
		t.Fatalf("empty JSON output = %q", emptyOutput)
	}
}

func TestListWeaviateCollectionsOutputs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/schema":
			_, _ = io.WriteString(w, `{"classes":[{"class":"WeaveImages","vectorizer":"none","properties":[{"name":"image","dataType":["text"]},{"name":"image_data","dataType":["text"]}]},{"class":"Docs","properties":[{"name":"content","dataType":["text"]}]}]}`)
		case "/v1/graphql":
			body, _ := io.ReadAll(r.Body)
			if strings.Contains(string(body), "Get") {
				_, _ = io.WriteString(w, `{"data":{"Get":{"Docs":[{"_additional":{"id":"doc-one"},"content":"first"},{"_additional":{"id":"doc-two"},"content":"second"}]}}}`)
			} else {
				_, _ = io.WriteString(w, `{"data":{"Aggregate":{"Docs":[{"meta":{"count":4}}],"WeaveImages":[{"meta":{"count":0}}]}}}`)
			}
		case "/v1/objects/Docs/doc-one":
			_, _ = io.WriteString(w, `{"id":"doc-one","class":"Docs","properties":{"content":"first","author":"Ada","topic":"search"}}`)
		case "/v1/objects/Docs/doc-two":
			_, _ = io.WriteString(w, `{"id":"doc-two","class":"Docs","properties":{"content":"second","author":"Grace"}}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	cfg := &config.VectorDBConfig{Type: config.VectorDBTypeLocal, URL: server.URL}

	output := captureUtilsOutput(t, func() {
		ListWeaviateCollections(context.Background(), cfg, 0, true, false)
	})
	for _, want := range []string{"Docs", "4 items", "WeaveImages", "empty", "Virtual structure"} {
		if !strings.Contains(output, want) {
			t.Errorf("list output missing %q: %s", want, output)
		}
	}

	jsonOutput := captureUtilsOutput(t, func() {
		ListWeaviateCollections(context.Background(), cfg, 1, false, true)
	})
	var result struct {
		Collections []struct {
			Name          string `json:"name"`
			DocumentCount int    `json:"document_count"`
		} `json:"collections"`
		Total int `json:"total"`
	}
	if err := json.Unmarshal([]byte(jsonOutput), &result); err != nil {
		t.Fatalf("invalid JSON output: %v: %q", err, jsonOutput)
	}
	if result.Total != 1 || len(result.Collections) != 1 || result.Collections[0].Name != "Docs" || result.Collections[0].DocumentCount != 4 {
		t.Fatalf("unexpected limited JSON output: %+v", result)
	}

	client, err := CreateWeaviateClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	schema, err := client.GetFullCollectionSchema(context.Background(), "WeaveImages")
	if err != nil || schema.Vectorizer != "none" || len(schema.Properties) != 2 || schema.Properties[1].Name != "image_data" {
		t.Fatalf("image schema = %+v, %v", schema, err)
	}
	ShowCollectionSchema(context.Background(), client, "WeaveImages")
	metadataOutput := captureUtilsOutput(t, func() {
		ShowCollectionMetadata(context.Background(), client, "Docs", true)
	})
	if !strings.Contains(metadataOutput, "from 2 documents") || !strings.Contains(metadataOutput, "Sample:") {
		t.Fatalf("metadata output = %q", metadataOutput)
	}
}

func captureUtilsOutput(t *testing.T, run func()) string {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	previous := os.Stdout
	os.Stdout = writer
	defer func() {
		os.Stdout = previous
		_ = writer.Close()
		_ = reader.Close()
	}()
	run()
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = previous
	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	return string(output)
}
