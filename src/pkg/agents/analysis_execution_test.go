// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package agents

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/llm"
)

func TestAnalysisAgentsWithCompletionFixture(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	t.Setenv("HOME", root)
	path := filepath.Join(root, "sample.md")
	if err := os.WriteFile(path, []byte("# Heading\n\nFirst paragraph.\n\nSecond paragraph."), 0600); err != nil {
		t.Fatal(err)
	}
	var response string
	var fail bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if fail {
			http.Error(w, "fixture failure", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": "fixture", "object": "chat.completion", "model": "gpt-4o",
			"choices": []any{map[string]any{"index": 0, "message": map[string]any{"role": "assistant", "content": response}, "finish_reason": "stop"}},
		})
	}))
	t.Cleanup(server.Close)
	t.Setenv("OPENAI_BASE_URL", server.URL+"/v1/")
	client, err := llm.NewOpenAIClient("fixture-key")
	if err != nil {
		t.Fatal(err)
	}
	schema := &SchemaAgent{llmClient: client, config: defaultConfig()}
	chunking := &ChunkingAgent{llmClient: client, config: defaultConfig()}
	schemaInput := &SchemaAnalysisInput{SampleFiles: []string{path}, MaxSamples: 1, CollectionName: "Docs"}
	chunkInput := &ChunkingAnalysisInput{SampleFiles: []string{path}, MaxSamples: 1, CollectionName: "Docs"}

	for _, tc := range []struct {
		name  string
		input any
		run   func(context.Context, interface{}) (interface{}, error)
		valid string
	}{
		{"schema", schemaInput, schema.Execute, `{"schema":{"collection_name":"Docs"},"confidence":0.9}`},
		{"chunking", chunkInput, chunking.Execute, `{"chunking_advice":{"recommended_size":800},"confidence":0.8}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, prefix := range []string{"", "Result: "} {
				response = prefix + tc.valid
				got, err := tc.run(context.Background(), tc.input)
				if err != nil {
					t.Fatal(err)
				}
				switch out := got.(type) {
				case *SchemaAnalysisOutput:
					if out.Schema.CollectionName != "Docs" || out.Confidence != 0.9 {
						t.Fatalf("schema: %+v", out)
					}
				case *ChunkingAnalysisOutput:
					if out.Recommendation.RecommendedSize != 800 || out.Metrics.AvgParagraphs != 3 {
						t.Fatalf("chunking: %+v", out)
					}
				default:
					t.Fatalf("unexpected result: %T", got)
				}
			}
			for _, invalid := range []string{"no json", "result {broken}"} {
				response = invalid
				if _, err := tc.run(context.Background(), tc.input); err == nil || !strings.Contains(err.Error(), "parse") {
					t.Fatalf("invalid response %q: %v", invalid, err)
				}
			}
			fail = true
			if _, err := tc.run(context.Background(), tc.input); err == nil || !strings.Contains(err.Error(), "failed to analyze") {
				t.Fatalf("service failure: %v", err)
			}
			fail = false
		})
	}
}
