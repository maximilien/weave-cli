// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package schema

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/agents"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type suggestTransport func(*http.Request) (*http.Response, error)

func (f suggestTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestRunSchemaSuggestWithRecordedCompletion(t *testing.T) {
	for _, mode := range []string{"saved", "display only", "save failure", "analysis failure", "empty", "bad glob", "viper key"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			t.Chdir(root)
			t.Setenv("HOME", root)
			t.Setenv("OPENAI_API_KEY", "fixture-key")
			t.Setenv("OPENAI_BASE_URL", "https://fixture.invalid/v1")
			viper.Reset()
			t.Cleanup(viper.Reset)
			if mode == "viper key" {
				t.Setenv("OPENAI_API_KEY", "")
				viper.Set("OPENAI_API_KEY", "fixture-key")
			}
			oldCollection, oldOutput, oldRequirements, oldMax, oldVDB, oldGlob := collectionName, outputFile, requirements, maxSamples, vdbType, globPattern
			defer func() {
				collectionName, outputFile, requirements, maxSamples, vdbType, globPattern = oldCollection, oldOutput, oldRequirements, oldMax, oldVDB, oldGlob
			}()
			oldInteractive, oldApply := interactive, apply
			defer func() { interactive, apply = oldInteractive, oldApply }()
			collectionName, requirements, maxSamples, vdbType, globPattern = "fixture-docs", "Preserve technical sections", 3, "weaviate", "*.md"
			interactive, apply = true, true
			samples := filepath.Join(root, "samples")
			if err := os.Mkdir(samples, 0700); err != nil {
				t.Fatal(err)
			}
			if mode != "empty" {
				if err := os.WriteFile(filepath.Join(samples, "guide.md"), []byte("# Technical guide\n\nDocument content for local analysis."), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if mode == "bad glob" {
				globPattern = "["
			}
			outputFile = filepath.Join(root, "result.yaml")
			if mode == "display only" {
				outputFile = ""
			}
			if mode == "save failure" {
				outputFile = filepath.Join(root, "missing", "result.yaml")
			}
			calls := 0
			previous := http.DefaultTransport
			http.DefaultTransport = suggestTransport(func(r *http.Request) (*http.Response, error) {
				calls++
				if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/chat/completions") {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
				if r.Header.Get("Authorization") != "Bearer fixture-key" {
					t.Error("fixture API key not used")
				}
				var body struct{ Messages []struct{ Content string } }
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}
				prompt, err := json.Marshal(body)
				if err != nil {
					t.Fatal(err)
				}
				for _, want := range []string{"fixture-docs", "Preserve technical sections"} {
					if !strings.Contains(string(prompt), want) {
						t.Errorf("prompt missing %q", want)
					}
				}
				content := `{"schema":{"collection_name":"fixture-docs","vector_dimensions":384,"similarity_metric":"cosine","fields":[{"name":"content","type":"text"}]},"confidence":0.91,"reasoning":"preserve content"}`
				if mode == "analysis failure" {
					content = "invalid completion"
				}
				data, err := json.Marshal(map[string]interface{}{"id": "fixture", "choices": []interface{}{map[string]interface{}{"index": 0, "message": map[string]string{"role": "assistant", "content": content}, "finish_reason": "stop"}}})
				if err != nil {
					t.Fatal(err)
				}
				return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(string(data))), Request: r}, nil
			})
			t.Cleanup(func() { http.DefaultTransport = previous })
			err := runSchemaSuggest(nil, []string{samples})
			wantErr := map[string]string{"empty": "no files found", "bad glob": "failed to scan files", "analysis failure": "analysis failed", "save failure": "failed to save"}[mode]
			if wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), wantErr) {
					t.Fatalf("error=%v want %q", err, wantErr)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			wantCalls := 1
			if mode == "empty" || mode == "bad glob" {
				wantCalls = 0
			}
			if calls != wantCalls {
				t.Fatalf("completion calls=%d want %d", calls, wantCalls)
			}
			if wantErr != "" || outputFile == "" {
				return
			}
			data, err := os.ReadFile(outputFile)
			if err != nil {
				t.Fatal(err)
			}
			var saved agents.SchemaConfig
			if err := yaml.Unmarshal(data, &saved); err != nil {
				t.Fatal(err)
			}
			if saved.CollectionName != "fixture-docs" || saved.VectorDimensions != 384 || len(saved.Fields) != 1 || saved.Fields[0].Name != "content" {
				t.Fatalf("saved schema=%+v", saved)
			}
		})
	}
}
