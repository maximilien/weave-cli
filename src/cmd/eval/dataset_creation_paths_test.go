// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package eval

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/evaluation"
)

func TestCreateDatasetFromTemplateAndExample(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	if err := os.MkdirAll(filepath.Join("evals", "datasets"), 0o755); err != nil {
		t.Fatal(err)
	}
	template := &evaluation.Dataset{
		Name: "sample", Version: "1.0.0",
		TestCases: []evaluation.TestCase{{ID: "test-001", Query: "question", ExpectedAnswer: "answer"}},
	}
	templatePath := filepath.Join("evals", "datasets", "sample.yaml")
	if err := evaluation.SaveDataset(template, templatePath); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, template, example string }{
		{name: "from-template", template: "sample"},
		{name: "from-example", example: templatePath},
	} {
		out := filepath.Join(root, tc.name+".yaml")
		createDataset(tc.name, tc.template, tc.example, false, out)
		got, err := evaluation.LoadDataset(out)
		if err != nil || got.Name != tc.name || len(got.TestCases) != 1 {
			t.Fatalf("created dataset %q: %#v, %v", tc.name, got, err)
		}
	}
}

func TestCreateInteractiveDatasetDefaults(t *testing.T) {
	input, err := os.CreateTemp(t.TempDir(), "empty-input")
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	previous := os.Stdin
	os.Stdin = input
	t.Cleanup(func() { os.Stdin = previous })

	got := createInteractiveDataset("interactive", true)
	if got.Name != "interactive" || got.Config.DefaultAgent != "rag-agent" ||
		len(got.TestCases) != 1 || got.TestCases[0].ID != "test-001" {
		t.Fatalf("interactive defaults: %#v", got)
	}
	minimal := createInteractiveDataset("minimal", false)
	if len(minimal.TestCases) != 1 || minimal.TestCases[0].Query == "" {
		t.Fatalf("minimal dataset: %#v", minimal)
	}
}
