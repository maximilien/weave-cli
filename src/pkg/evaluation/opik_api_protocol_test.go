// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package evaluation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/llm"
)

func TestOpikAPIClientLifecycle(t *testing.T) {
	created := false
	requests := make([]string, 0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		if r.Header.Get("Authorization") != "fixture-key" || r.Header.Get("Comet-Workspace") != "fixture-workspace" {
			t.Errorf("unexpected Opik headers: %#v", r.Header)
		}
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/private/datasets/retrieve":
			if !created {
				http.Error(w, "missing", http.StatusNotFound)
				return
			}
			_, _ = fmt.Fprint(w, `{"id":"dataset-1","name":"fixture-dataset","dataset_items_count":1,"latest_version":{"id":"version-1","version_name":"v1"}}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/private/datasets":
			created = true
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodGet && r.URL.Path == "/v1/private/datasets/dataset-1/items":
			_, _ = fmt.Fprint(w, `{"content":[{"dataset_item_id":"item-1","data":{"weave_test_case_id":"case-1"}}]}`)
		case r.Method == http.MethodPut && r.URL.Path == "/v1/private/datasets/items":
			var payload map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Errorf("decode dataset items: %v", err)
			}
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/private/experiments":
			w.WriteHeader(http.StatusCreated)
		case r.Method == http.MethodPut && r.URL.Path == "/v1/private/experiments/items/bulk":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodPatch && strings.HasPrefix(r.URL.Path, "/v1/private/experiments/"):
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/private/traces":
			w.WriteHeader(http.StatusCreated)
		case r.Method == http.MethodPatch && strings.HasPrefix(r.URL.Path, "/v1/private/traces/"):
			var payload map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Errorf("decode trace update: %v", err)
			}
			if _, ok := payload["error_info"]; !ok {
				t.Error("trace error details were not sent")
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "unexpected request", http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := &OpikAPIClient{
		baseURL: server.URL, apiKey: "fixture-key", workspace: "fixture-workspace",
		projectName: "fixture-project", httpClient: server.Client(),
	}
	dataset := &Dataset{
		Name: "fixture-dataset", Description: "fixture", Tags: []string{"unit"},
		TestCases: []TestCase{{
			ID: "case-1", Query: "question", ExpectedAnswer: "answer",
			ExpectedCitations: []string{"[1]"}, RequiredConcepts: []string{"concept"},
			RetrievedContext: []string{"context"}, Collection: "Docs", MustCite: true,
			MinRelevanceScore: 0.5,
		}},
	}

	summary, err := client.UploadDataset(context.Background(), dataset)
	if err != nil {
		t.Fatal(err)
	}
	if summary.ID != "dataset-1" || summary.ItemCount != 1 || summary.URL != "https://www.comet.com/fixture-workspace/fixture-project" {
		t.Fatalf("unexpected dataset summary: %#v", summary)
	}

	run := &EvaluationRun{
		ID: "run-1", DatasetName: dataset.Name, AgentName: "agent", Collection: "Docs",
		Config:  EvaluationConfig{Parameters: map[string]string{"evaluator_provider": "local"}},
		Summary: EvaluationSummary{PassRate: 100, AvgAccuracy: 0.9, AvgCitation: 0.8},
		Results: []TestCaseResult{{
			TestCaseID: "case-1", Query: "question", ActualAnswer: "answer", ActualCitations: []string{"[1]"},
			Passed: true, AccuracyScore: 0.9, CitationScore: 0.8, HallucinationScore: 0.1,
			ContextRelevanceScore: 0.7, FaithfulnessScore: 0.95, ResponseTime: 12,
			CustomScores: map[string]float64{"style": 0.75}, Details: map[string]interface{}{"source": "fixture"},
		}, {
			TestCaseID: "unmatched", Query: "ignored",
		}},
	}
	syncResult, err := client.SyncEvaluationRun(context.Background(), dataset, run)
	if err != nil {
		t.Fatal(err)
	}
	if syncResult.Experiment.ID == "" || syncResult.Experiment.Name != "fixture-dataset-agent-run-1" {
		t.Fatalf("unexpected sync result: %#v", syncResult)
	}

	trace, err := client.CreateTrace(context.Background(), "00112233445566778899aabbccddeeff", "query", time.Unix(10, 0), map[string]interface{}{"q": "x"}, nil)
	if err != nil || trace.ID != "00112233-4455-6677-8899-aabbccddeeff" {
		t.Fatalf("CreateTrace() = (%#v, %v)", trace, err)
	}
	if err := client.UpdateTrace(context.Background(), trace.ID, time.Unix(20, 0), map[string]interface{}{"answer": "y"}, nil, errors.New("fixture failure")); err != nil {
		t.Fatal(err)
	}
	if len(requests) < 12 {
		t.Fatalf("expected complete lifecycle, got requests: %v", requests)
	}
}

func TestOpikAPIClientValidationAndErrors(t *testing.T) {
	if _, err := NewOpikAPIClient(nil); err == nil {
		t.Fatal("expected nil configuration error")
	}
	if _, err := NewOpikAPIClient(&llm.OpikConfig{}); err == nil {
		t.Fatal("expected missing API key error")
	}

	t.Setenv("OPIK_API_BASE_URL", "https://override.invalid/root/")
	client, err := NewOpikAPIClient(&llm.OpikConfig{APIKey: "key", Workspace: "workspace", ProjectName: "project"})
	if err != nil || client.baseURL != "https://override.invalid/root" {
		t.Fatalf("NewOpikAPIClient() = (%#v, %v)", client, err)
	}
	if client.projectURL() != "https://www.comet.com/workspace/project" {
		t.Fatalf("projectURL() = %q", client.projectURL())
	}
	client.workspace = ""
	if client.projectURL() != "https://www.comet.com" {
		t.Fatalf("fallback projectURL() = %q", client.projectURL())
	}

	baseURLTests := map[string]string{
		"": defaultOpikAPIBaseURL,
		"https://example.test/custom/api/v1/private/otel/v1/traces": "https://example.test/custom/api",
		"https://example.test/custom/v1/private/otel/v1/traces":     defaultOpikAPIBaseURL,
		"https://example.test/unrelated":                            defaultOpikAPIBaseURL,
		"://bad":                                                    defaultOpikAPIBaseURL,
	}
	for endpoint, want := range baseURLTests {
		if got := deriveOpikAPIBaseURL(endpoint); got != want {
			t.Errorf("deriveOpikAPIBaseURL(%q) = %q, want %q", endpoint, got, want)
		}
	}
	if _, err := formatTraceIDAsUUID("short"); err == nil {
		t.Fatal("expected invalid trace ID error")
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/failure":
			http.Error(w, "denied", http.StatusForbidden)
		case "/invalid-json":
			_, _ = fmt.Fprint(w, "{")
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer server.Close()
	client.baseURL = server.URL
	client.httpClient = server.Client()
	if err := client.doJSON(context.Background(), http.MethodGet, "/failure", nil, nil); err == nil || !strings.Contains(err.Error(), "HTTP 403") {
		t.Fatalf("expected HTTP error, got %v", err)
	}
	var decoded map[string]interface{}
	if err := client.doJSON(context.Background(), http.MethodGet, "/invalid-json", nil, &decoded); err == nil || !strings.Contains(err.Error(), "decode") {
		t.Fatalf("expected decode error, got %v", err)
	}
	if err := client.doJSON(context.Background(), http.MethodPost, "/ok", make(chan int), nil); err == nil || !strings.Contains(err.Error(), "marshal") {
		t.Fatalf("expected marshal error, got %v", err)
	}
	client.baseURL = "://bad"
	if err := client.doJSON(context.Background(), http.MethodGet, "/request", nil, nil); err == nil || !strings.Contains(err.Error(), "create Opik request") {
		t.Fatalf("expected request construction error, got %v", err)
	}
	client.baseURL = "http://127.0.0.1:1"
	client.httpClient = &http.Client{Timeout: 50 * time.Millisecond}
	if err := client.doJSON(context.Background(), http.MethodGet, "/unavailable", nil, nil); err == nil || !strings.Contains(err.Error(), "request failed") {
		t.Fatalf("expected transport error, got %v", err)
	}
}
