// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package evaluation

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDatasetCustomEvaluatorThresholdsAndErrors(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".weave-cli", "evals", "evaluators")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, pattern string
		required      bool
	}{
		{"match", `\d+`, true},
		{"optional", `missing`, false},
		{"required", `missing`, true},
		{"bad_regex", `[`, true},
	} {
		body := fmt.Sprintf("name: %s\nprompt: placeholder\nrequired: %t\nscoring:\n  type: regex\n  pattern: '%s'\n  threshold: 0.8\n", tc.name, tc.required, tc.pattern)
		if err := os.WriteFile(filepath.Join(dir, tc.name+".yaml"), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		name       string
		evaluators []string
		passed     bool
		wantScore  map[string]float64
		wantErr    string
	}{
		{"success", []string{"match"}, true, map[string]float64{"match": 1}, ""},
		{"optional below threshold", []string{"optional"}, true, map[string]float64{"optional": 0}, "Custom evaluator optional: score 0.00 below threshold 0.80"},
		{"required below threshold", []string{"required"}, false, map[string]float64{"required": 0}, "Custom evaluator required: score 0.00 below threshold 0.80"},
		{"missing definition", []string{"missing", "match"}, true, map[string]float64{"match": 1}, "Failed to load custom evaluator missing"},
		{"invalid regex", []string{"bad_regex", "match"}, true, map[string]float64{"match": 1}, "Custom evaluator bad_regex failed: invalid regex pattern"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			testCase := &TestCase{ID: "case", Query: "What is the value?", ExpectedAnswer: "123"}
			result, err := EvaluateTestCaseWithDataset(context.Background(), testCase, "Value 123", nil, NewLocalProvider(&MockLLMClient{response: "0.85"}), &Dataset{CustomEvaluators: tc.evaluators})
			if err != nil {
				t.Fatal(err)
			}
			if result.Passed != tc.passed || result.TestCaseID != "case" || result.Details["evaluator_provider"] != "local" {
				t.Fatalf("result = %#v", result)
			}
			if len(result.CustomScores) != len(tc.wantScore) {
				t.Fatalf("scores = %#v", result.CustomScores)
			}
			for name, want := range tc.wantScore {
				if result.CustomScores[name] != want || result.Details["custom_"+name+"_score"] != want {
					t.Fatalf("scores/details = %#v %#v", result.CustomScores, result.Details)
				}
			}
			if tc.wantErr == "" && len(result.Errors) != 0 {
				t.Fatalf("unexpected errors: %#v", result.Errors)
			}
			if tc.wantErr != "" && !strings.Contains(strings.Join(result.Errors, "\n"), tc.wantErr) {
				t.Fatalf("errors = %#v; want %q", result.Errors, tc.wantErr)
			}
		})
	}
}
