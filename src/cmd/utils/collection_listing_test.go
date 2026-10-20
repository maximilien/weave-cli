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
	"path/filepath"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/weaviate"
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

	queryOutput := captureUtilsOutput(t, func() {
		QueryWeaviateCollection(context.Background(), cfg, "Docs", "fixture question", weaviate.QueryOptions{TopK: 2})
	})
	if !strings.Contains(queryOutput, "Found 2 results") {
		t.Fatalf("Weaviate query output did not report fixture result count: %s", queryOutput)
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

	summary := captureUtilsOutput(t, func() {
		ShowWeaviateCollection(context.Background(), cfg, "Docs", 1, false, false, true, true, false, false, false, "", "", false)
	})
	if !strings.Contains(summary, "first") || !strings.Contains(summary, "from 2 documents") {
		t.Fatalf("collection summary = %q", summary)
	}
	fullSummary := captureUtilsOutput(t, func() {
		ShowWeaviateCollection(context.Background(), cfg, "Docs", 3, true, false, false, false, false, false, false, "", "", false)
	})
	if !strings.Contains(fullSummary, "first") {
		t.Fatalf("untruncated collection summary = %q", fullSummary)
	}

	export, err := ExportCollectionSchemaAndMetadata(context.Background(), client, "Docs", true, true, true)
	if err != nil {
		t.Fatalf("ExportCollectionSchemaAndMetadata() error = %v", err)
	}
	if export.Name != "Docs" || export.Schema == nil || len(export.Metadata) == 0 {
		t.Fatalf("schema and metadata export = %#v", export)
	}
	for key, info := range export.Metadata {
		if info.Occurrences != 0 || info.Sample != nil {
			t.Errorf("compact metadata %q retained runtime details: %#v", key, info)
		}
	}
	fullExport, err := ExportCollectionSchemaAndMetadata(context.Background(), client, "Docs", true, false, false)
	if err != nil {
		t.Fatalf("ExportCollectionSchemaAndMetadata(noncompact) error = %v", err)
	}
	for key, info := range fullExport.Metadata {
		if info.Occurrences == 0 {
			t.Errorf("noncompact metadata %q has no occurrence count: %#v", key, info)
		}
	}
}

func TestWeaviateCollectionMaintenanceWithFakeServer(t *testing.T) {
	var deleted []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/meta":
			_, _ = io.WriteString(w, `{"hostname":"fake","version":"1.27.0","modules":{}}`)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/schema":
			_, _ = io.WriteString(w, `{"classes":[{"class":"Docs"},{"class":"DocsArchive"},{"class":"Images"}]}`)
		case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/v1/schema/"):
			deleted = append(deleted, strings.TrimPrefix(r.URL.Path, "/v1/schema/"))
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	cfg := &config.VectorDBConfig{Type: config.VectorDBTypeLocal, URL: server.URL}
	ctx := context.Background()

	if count, err := CountWeaviateCollections(ctx, cfg); err != nil || count != 3 {
		t.Fatalf("CountWeaviateCollections() = (%d, %v), want (3, nil)", count, err)
	}
	if err := DeleteWeaviateCollectionsByPattern(ctx, cfg, "Docs*"); err != nil {
		t.Fatalf("DeleteWeaviateCollectionsByPattern() error = %v", err)
	}
	if len(deleted) != 2 || deleted[0] != "Docs" || deleted[1] != "DocsArchive" {
		t.Fatalf("pattern deletions = %#v", deleted)
	}
	if err := DeleteWeaviateCollections(ctx, cfg, []string{"Images"}); err != nil {
		t.Fatalf("DeleteWeaviateCollections() error = %v", err)
	}
	if err := DeleteWeaviateCollectionSchemasByPattern(ctx, cfg, "^Images$"); err != nil {
		t.Fatalf("DeleteWeaviateCollectionSchemasByPattern() error = %v", err)
	}
	if err := DeleteWeaviateCollectionSchema(ctx, cfg, "Docs"); err != nil {
		t.Fatalf("DeleteWeaviateCollectionSchema() error = %v", err)
	}
	if len(deleted) != 5 || deleted[2] != "Images" || deleted[3] != "Images" || deleted[4] != "Docs" {
		t.Fatalf("all schema deletions = %#v", deleted)
	}
}

func TestDeleteAllWeaviateCollectionsClearsDocuments(t *testing.T) {
	var deletedIDs []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/meta":
			_, _ = io.WriteString(w, `{"hostname":"fake","version":"1.27.0","modules":{}}`)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/schema":
			_, _ = io.WriteString(w, `{"classes":[{"class":"Docs","properties":[{"name":"content","dataType":["text"]}]}]}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/graphql":
			_, _ = io.WriteString(w, `{"data":{"Get":{"Docs":[{"_additional":{"id":"doc-1","vector":[0.1,0.2]},"content":"fixture"}]}}}`)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/objects/Docs/doc-1":
			_, _ = io.WriteString(w, `{"id":"doc-1","class":"Docs","properties":{"content":"fixture"}}`)
		case r.Method == http.MethodDelete && r.URL.Path == "/v1/objects/Docs/doc-1":
			deletedIDs = append(deletedIDs, "doc-1")
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	cfg := &config.VectorDBConfig{Type: config.VectorDBTypeLocal, URL: server.URL}

	DeleteAllWeaviateCollections(context.Background(), cfg)
	if len(deletedIDs) != 1 || deletedIDs[0] != "doc-1" {
		t.Fatalf("deleted document IDs = %#v", deletedIDs)
	}
}

func TestWeaviateCollectionCreationPathsWithFakeServer(t *testing.T) {
	var created []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/meta":
			_, _ = io.WriteString(w, `{"hostname":"fake","version":"1.27.0","modules":{}}`)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/schema":
			_, _ = io.WriteString(w, `{"classes":[]}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/schema":
			var schema struct {
				Class string `json:"class"`
			}
			if err := json.NewDecoder(r.Body).Decode(&schema); err != nil {
				t.Errorf("decode schema request: %v", err)
			}
			created = append(created, schema.Class)
			w.WriteHeader(http.StatusCreated)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	cfg := &config.VectorDBConfig{Type: config.VectorDBTypeLocal, URL: server.URL}
	ctx := context.Background()

	if err := CreateWeaviateCollection(ctx, cfg, "BasicDocs", "text-embedding-3-small", nil); err != nil {
		t.Fatalf("CreateWeaviateCollection() error = %v", err)
	}
	if err := CreateWeaviateCollection(ctx, cfg, "CustomDocs", "", nil); err != nil {
		t.Fatalf("CreateWeaviateCollection(custom) error = %v", err)
	}

	yamlPath := filepath.Join(t.TempDir(), "schema.yaml")
	if err := os.WriteFile(yamlPath, []byte("name: FileDocs\nschema:\n  class: FileDocs\n  properties:\n    - name: content\n      datatype: [text]\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := CreateWeaviateCollectionFromSchemaFile(ctx, cfg, "RenamedDocs", yamlPath); err != nil {
		t.Fatalf("CreateWeaviateCollectionFromSchemaFile() error = %v", err)
	}
	appConfig := &config.Config{Databases: config.DatabasesConfig{Schemas: []config.SchemaDefinition{{
		Name: "NamedSchema",
		Schema: map[string]interface{}{
			"class":      "NamedSchema",
			"properties": []interface{}{map[string]interface{}{"name": "content", "datatype": []interface{}{"text"}}},
		},
	}}}}
	if err := CreateWeaviateCollectionFromConfigSchema(ctx, appConfig, cfg, "ConfiguredDocs", "NamedSchema", "nested"); err != nil {
		t.Fatalf("CreateWeaviateCollectionFromConfigSchema() error = %v", err)
	}
	if len(created) != 4 || created[0] != "BasicDocs" || created[1] != "CustomDocs" || created[2] != "RenamedDocs" || created[3] != "ConfiguredDocs" {
		t.Fatalf("created collection names = %#v", created)
	}
}

func TestVectorDBTypeUtilityFunctions(t *testing.T) {
	if types := GetSupportedVectorDBTypes(); len(types) == 0 {
		t.Fatal("GetSupportedVectorDBTypes() returned no registered types")
	}
	if !IsVectorDBTypeSupported("mock") || IsVectorDBTypeSupported("unsupported") {
		t.Fatal("IsVectorDBTypeSupported() returned unexpected results")
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
