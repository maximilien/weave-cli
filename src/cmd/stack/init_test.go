// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package stack

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	stackpkg "github.com/maximilien/weave-cli/src/pkg/stack"
)

func inTempWorkingDir(t *testing.T) string {
	t.Helper()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("change working directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})

	return tempDir
}

func TestGenerateStackFromTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		runtime  string
		check    func(*testing.T, *stackpkg.StackConfig)
	}{
		{
			name:     "quickstart kind",
			template: "quickstart",
			runtime:  "kind",
			check: func(t *testing.T, config *stackpkg.StackConfig) {
				if config.Name != "my-rag-stack" || config.Runtime.Kubernetes.Kind == nil {
					t.Fatalf("unexpected quickstart config: %#v", config)
				}
				if len(config.Collections) != 1 || config.Collections[0].Schema.VectorDimensions != 1536 {
					t.Fatalf("unexpected quickstart collection: %#v", config.Collections)
				}
			},
		},
		{
			name:     "production eks",
			template: "production",
			runtime:  "eks",
			check: func(t *testing.T, config *stackpkg.StackConfig) {
				if config.Runtime.Kubernetes.EKS == nil || config.Ingestion == nil || config.Dashboard == nil {
					t.Fatalf("production template is incomplete: %#v", config)
				}
				if !config.Ingestion.Checkpoint.Enabled || !config.Dashboard.Enabled {
					t.Fatalf("production features are not enabled: %#v", config)
				}
			},
		},
		{
			name:     "multimodal gke",
			template: "multimodal",
			runtime:  "gke",
			check: func(t *testing.T, config *stackpkg.StackConfig) {
				if config.Runtime.Kubernetes.GKE == nil || len(config.Collections) != 2 {
					t.Fatalf("multimodal template is incomplete: %#v", config)
				}
				if config.Collections[1].Type != "image" || config.Infrastructure.ImageStorage == nil {
					t.Fatalf("image configuration is missing: %#v", config)
				}
			},
		},
		{
			name:     "oss minikube",
			template: "oss",
			runtime:  "minikube",
			check: func(t *testing.T, config *stackpkg.StackConfig) {
				if config.Runtime.Kubernetes.Minikube == nil || config.Runtime.ContainerRuntime != "docker" {
					t.Fatalf("unexpected minikube defaults: %#v", config.Runtime)
				}
				if config.Infrastructure.LLM.Provider != "ollama" || config.Collections[0].Schema.VectorDimensions != 768 {
					t.Fatalf("unexpected OSS providers: %#v", config)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config, err := generateStackFromTemplate(test.template, test.runtime)
			if err != nil {
				t.Fatalf("generate template: %v", err)
			}
			if config.Runtime.Kubernetes.Provider != test.runtime {
				t.Fatalf("provider = %q, want %q", config.Runtime.Kubernetes.Provider, test.runtime)
			}
			test.check(t, config)
		})
	}
}

func TestGenerateStackFromTemplateRejectsUnknownTemplate(t *testing.T) {
	config, err := generateStackFromTemplate("unknown", "kind")
	if err == nil || config != nil || !strings.Contains(err.Error(), "unknown template") {
		t.Fatalf("generateStackFromTemplate() = (%#v, %v), want unknown-template error", config, err)
	}
}

func TestSetRuntimeDefaultsLeavesUnknownProviderUnconfigured(t *testing.T) {
	config := generateQuickstartTemplate("custom")
	if config.Runtime.Kubernetes.Provider != "custom" {
		t.Fatalf("provider = %q, want custom", config.Runtime.Kubernetes.Provider)
	}
	if config.Runtime.Kubernetes.Kind != nil || config.Runtime.Kubernetes.Minikube != nil ||
		config.Runtime.Kubernetes.EKS != nil || config.Runtime.Kubernetes.GKE != nil {
		t.Fatalf("unknown provider received runtime defaults: %#v", config.Runtime.Kubernetes)
	}
}

func TestRunInitCreatesStackFiles(t *testing.T) {
	tempDir := inTempWorkingDir(t)

	oldTemplate, oldRuntime, oldForce := initTemplate, initRuntime, initForce
	t.Cleanup(func() {
		initTemplate, initRuntime, initForce = oldTemplate, oldRuntime, oldForce
	})
	initTemplate, initRuntime, initForce = "quickstart", "kind", false

	if err := runInit(nil, nil); err != nil {
		t.Fatalf("runInit() error = %v", err)
	}

	for _, path := range []string{
		stackpkg.DefaultStackFile,
		".gitignore",
		filepath.Join("kubernetes", "templates"),
	} {
		if _, err := os.Stat(filepath.Join(tempDir, path)); err != nil {
			t.Errorf("expected %s: %v", path, err)
		}
	}

	config, err := stackpkg.LoadStackConfig("")
	if err != nil {
		t.Fatalf("load generated stack: %v", err)
	}
	if config.Name != "my-rag-stack" {
		t.Fatalf("generated stack name = %q", config.Name)
	}
}

func TestRunInitExistingAndInvalidConfigurations(t *testing.T) {
	inTempWorkingDir(t)

	oldTemplate, oldRuntime, oldForce := initTemplate, initRuntime, initForce
	t.Cleanup(func() {
		initTemplate, initRuntime, initForce = oldTemplate, oldRuntime, oldForce
	})

	if err := os.WriteFile(stackpkg.DefaultStackFile, []byte("existing"), 0o600); err != nil {
		t.Fatalf("write existing stack: %v", err)
	}
	initTemplate, initRuntime, initForce = "quickstart", "kind", false
	if err := runInit(nil, nil); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("runInit() error = %v, want existing-stack error", err)
	}

	initTemplate, initForce = "unknown", true
	if err := runInit(nil, nil); err == nil || !strings.Contains(err.Error(), "unknown template") {
		t.Fatalf("runInit() error = %v, want unknown-template error", err)
	}
}

func TestCreateGitignore(t *testing.T) {
	inTempWorkingDir(t)

	if err := createGitignore(); err != nil {
		t.Fatalf("createGitignore() error = %v", err)
	}
	data, err := os.ReadFile(".gitignore")
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if !strings.Contains(string(data), ".weave-state/") || !strings.Contains(string(data), "*.kubeconfig") {
		t.Fatalf("unexpected .gitignore contents: %s", data)
	}
}
