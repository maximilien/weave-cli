// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package eval

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/evaluation"
)

func TestDatasetLifecycleHelpers(t *testing.T) {
	root := setupEvalWorkspace(t)
	listDatasets("text")

	createdPath := filepath.Join(root, "evals", "datasets", "created.yaml")
	createDataset("created", "", "", false, createdPath)
	created, err := evaluation.LoadDataset(createdPath)
	if err != nil {
		t.Fatalf("load created dataset: %v", err)
	}
	if created.Name != "created" || len(created.TestCases) != 1 {
		t.Fatalf("created dataset = %#v", created)
	}

	for _, format := range []string{"text", "json", "yaml"} {
		listDatasets(format)
		showDataset("created", format)
	}
	validateDataset(createdPath)

	copiedPath := filepath.Join(root, "evals", "datasets", "copied.yaml")
	createDataset("copied", "", createdPath, false, copiedPath)
	if copied, err := evaluation.LoadDataset(copiedPath); err != nil || copied.Name != "copied" {
		t.Fatalf("copied dataset = %#v, %v", copied, err)
	}

	templatePath := filepath.Join(root, "evals", "datasets", "template.yaml")
	if err := evaluation.SaveDataset(created, templatePath); err != nil {
		t.Fatalf("save template: %v", err)
	}
	if template := getTemplateDataset("template"); template == nil {
		t.Fatal("getTemplateDataset() did not load template")
	}
	if template := getTemplateDataset("missing"); template != nil {
		t.Fatalf("getTemplateDataset(missing) = %#v", template)
	}

	fromTemplatePath := filepath.Join(root, "evals", "datasets", "from-template.yaml")
	createDataset("from-template", "template", "", false, fromTemplatePath)
	if dataset, err := evaluation.LoadDataset(fromTemplatePath); err != nil || dataset.Name != "from-template" {
		t.Fatalf("template dataset = %#v, %v", dataset, err)
	}
}

func TestInteractiveDatasetAndPromptHelpers(t *testing.T) {
	dataset := createInteractiveDataset("minimal", false)
	if dataset.Name != "minimal" || dataset.Version != "1.0.0" || len(dataset.TestCases) != 1 {
		t.Fatalf("createInteractiveDataset() = %#v", dataset)
	}

	withEvalStdin(t, "custom\n", func() {
		if got := promptString("Name", "default"); got != "custom" {
			t.Fatalf("promptString() = %q", got)
		}
	})
	withEvalStdin(t, "\n", func() {
		if got := promptString("Name", "default"); got != "default" {
			t.Fatalf("promptString(default) = %q", got)
		}
	})
	withEvalStdin(t, "7\n", func() {
		if got := promptInt("Count", 2); got != 7 {
			t.Fatalf("promptInt() = %d", got)
		}
	})
	withEvalStdin(t, "\n", func() {
		if got := promptInt("Count", 2); got != 2 {
			t.Fatalf("promptInt(default) = %d", got)
		}
	})
	withEvalStdin(t, "0.75\n", func() {
		if got := promptFloat("Score", 0.5); got != 0.75 {
			t.Fatalf("promptFloat() = %f", got)
		}
	})
	withEvalStdin(t, "\n", func() {
		if got := promptFloat("Score", 0.5); got != 0.5 {
			t.Fatalf("promptFloat(default) = %f", got)
		}
	})
	for _, test := range []struct {
		input        string
		defaultValue bool
		want         bool
	}{
		{input: "yes\n", want: true},
		{input: "n\n", defaultValue: true, want: false},
		{input: "\n", defaultValue: true, want: true},
	} {
		withEvalStdin(t, test.input, func() {
			if got := promptBool("Enabled", test.defaultValue); got != test.want {
				t.Fatalf("promptBool(%q) = %t, want %t", test.input, got, test.want)
			}
		})
	}
}

func TestCustomEvaluatorLifecycle(t *testing.T) {
	root := setupEvalWorkspace(t)
	if err := runListEvaluators(nil, nil); err != nil {
		t.Fatalf("runListEvaluators(empty) error: %v", err)
	}

	for _, scoringType := range []string{"llm_judge", "regex", "exact_match", "contains"} {
		name := strings.ReplaceAll(scoringType, "_", "-")
		if err := runCreateEvaluator(name, scoringType); err != nil {
			t.Fatalf("runCreateEvaluator(%s) error: %v", scoringType, err)
		}
		path := filepath.Join(root, "evals", "evaluators", name+".yaml")
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("created evaluator %s: %v", path, err)
		}
		if err := runValidateEvaluator(nil, []string{path}); err != nil {
			t.Fatalf("runValidateEvaluator(%s) error: %v", scoringType, err)
		}
		if template := generateEvaluatorTemplate(name, scoringType); !strings.Contains(template, "name: "+name) {
			t.Fatalf("generated template = %q", template)
		}
	}

	if err := runListEvaluators(nil, nil); err != nil {
		t.Fatalf("runListEvaluators() error: %v", err)
	}
	if err := runCreateEvaluator("regex", "regex"); err == nil {
		t.Fatal("runCreateEvaluator() overwrote existing evaluator")
	}
	if err := runValidateEvaluator(nil, []string{filepath.Join(root, "missing.yaml")}); err == nil {
		t.Fatal("runValidateEvaluator() accepted missing file")
	}

	invalidYAML := filepath.Join(root, "invalid.yaml")
	if err := os.WriteFile(invalidYAML, []byte("name: ["), 0o600); err != nil {
		t.Fatalf("write invalid YAML: %v", err)
	}
	if err := runValidateEvaluator(nil, []string{invalidYAML}); err == nil {
		t.Fatal("runValidateEvaluator() accepted invalid YAML")
	}

	invalidDefinition := filepath.Join(root, "invalid-definition.yaml")
	if err := os.WriteFile(invalidDefinition, []byte("name: incomplete\n"), 0o600); err != nil {
		t.Fatalf("write invalid definition: %v", err)
	}
	if err := runValidateEvaluator(nil, []string{invalidDefinition}); err == nil {
		t.Fatal("runValidateEvaluator() accepted invalid definition")
	}
}

func setupEvalWorkspace(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	t.Chdir(root)
	t.Setenv("HOME", root)
	for _, directory := range []string{
		filepath.Join(root, ".git"),
		filepath.Join(root, "configs", "agents"),
		filepath.Join(root, "evals", "datasets"),
		filepath.Join(root, "evals", "evaluators"),
	} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatalf("create %s: %v", directory, err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "configs", "agents", "rag-agent.yaml"), nil, 0o600); err != nil {
		t.Fatalf("write development marker: %v", err)
	}
	return root
}

func withEvalStdin(t *testing.T, input string, run func()) {
	t.Helper()
	readEnd, writeEnd, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdin pipe: %v", err)
	}
	oldStdin := os.Stdin
	os.Stdin = readEnd
	t.Cleanup(func() {
		os.Stdin = oldStdin
		_ = readEnd.Close()
	})
	if _, err := writeEnd.WriteString(input); err != nil {
		t.Fatalf("write stdin: %v", err)
	}
	if err := writeEnd.Close(); err != nil {
		t.Fatalf("close stdin writer: %v", err)
	}
	run()
}
