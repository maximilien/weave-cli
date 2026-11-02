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
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/weaviate"
)

func TestAgentExecutionWithLocalCompletionFixture(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("OPENAI_API_KEY", "fixture-key")
	t.Chdir(root)
	dir := filepath.Join(root, "configs", "agents")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"coverage-fixture-alpha"} {
		config := "name: " + name + "\ntype: rag\nllm:\n  model: gpt-4o\nsystem_prompt: Answer from sources.\noutput:\n  format: text\n"
		if err := os.WriteFile(filepath.Join(dir, name+".yaml"), []byte(config), 0600); err != nil {
			t.Fatal(err)
		}
	}
	requests := make(chan map[string]interface{}, 8)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/chat/completions" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer fixture-key" {
			t.Error("missing fixture authentication")
		}
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
		}
		requests <- req
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"fixture","object":"chat.completion","model":"gpt-4o","choices":[{"index":0,"message":{"role":"assistant","content":"Fixture answer [1]"},"finish_reason":"stop"}],"usage":{"prompt_tokens":10,"completion_tokens":4,"total_tokens":14}}`)
	}))
	defer server.Close()
	t.Setenv("OPENAI_BASE_URL", server.URL+"/v1/")
	results := []*vectordb.QueryResult{{Document: vectordb.Document{ID: "source-1", Content: "Fixture source text", Metadata: map[string]interface{}{"filename": "source.txt"}}, Score: 0.9}}
	for _, tc := range []struct {
		name      string
		agents    []string
		format    string
		progress  bool
		wantCalls int
	}{
		{name: "single text", agents: []string{"coverage-fixture-alpha"}, format: "text", wantCalls: 1},
		{name: "single JSON with progress", agents: []string{"coverage-fixture-alpha"}, format: "json", progress: true, wantCalls: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, diagnostics := captureCollectionListing(t, func() {
				ExecuteQueryWithAgent(context.Background(), tc.agents, "What is the answer?", results, tc.format, tc.progress)
			})
			if !strings.Contains(out, "Fixture answer") {
				t.Fatalf("answer missing: %s %s", out, diagnostics)
			}
			if len(requests) != tc.wantCalls {
				t.Fatalf("completion calls=%d; want %d", len(requests), tc.wantCalls)
			}
			for len(requests) > 0 {
				req := <-requests
				if req["model"] != "gpt-4o" {
					t.Fatalf("model = %#v", req)
				}
				encoded, _ := json.Marshal(req["messages"])
				if !strings.Contains(string(encoded), "What is the answer?") {
					t.Fatalf("query missing in prompt: %s", encoded)
				}
			}
		})
	}
	cfg, _ := documentOperationsFixture(t, "none", false, false, false, false)
	out, diagnostics := captureCollectionListing(t, func() {
		QueryCollectionWithAgent(context.Background(), cfg, "Docs", "What is the answer?", weaviate.QueryOptions{TopK: 2}, []string{"coverage-fixture-alpha"}, "text", false)
	})
	if !strings.Contains(out, "Fixture answer") || len(requests) != 1 {
		t.Fatalf("query agent output=%s diagnostics=%s calls=%d", out, diagnostics, len(requests))
	}
	<-requests
	for _, tc := range []struct {
		name string
		run  func()
	}{
		{"weaviate query", func() {
			QueryWeaviateCollectionWithAgent(context.Background(), cfg, "Docs", "What is the answer?", weaviate.QueryOptions{TopK: 2}, []string{"coverage-fixture-alpha"}, "text", false)
		}},
		{"multiple collections", func() {
			QueryMultipleCollectionsWithAgent(context.Background(), cfg, []string{"Docs"}, "What is the answer?", weaviate.QueryOptions{TopK: 2}, []string{"coverage-fixture-alpha"}, "text", false)
		}},
		{"cross database", func() {
			QueryMultipleCollectionsWithAgentCrossVDB(context.Background(), []CollectionSpec{{Name: "Docs", VDBKey: "fixture"}}, map[string]*config.VectorDBConfig{"Docs": cfg}, "What is the answer?", weaviate.QueryOptions{TopK: 2}, []string{"coverage-fixture-alpha"}, "text", false)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, diagnostics := captureCollectionListing(t, tc.run)
			if !strings.Contains(out, "Fixture answer") || len(requests) != 1 {
				t.Fatalf("output=%s diagnostics=%s calls=%d", out, diagnostics, len(requests))
			}
			<-requests
		})
	}
}
