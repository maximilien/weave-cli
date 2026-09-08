// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package stack

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnhanceHelmError(t *testing.T) {
	if EnhanceHelmError(nil, "chart", "release") != nil {
		t.Fatal("EnhanceHelmError(nil) returned an error")
	}
	tests := []struct {
		message string
		want    string
	}{
		{"cannot re-use a name that is still in use", "helm uninstall release"},
		{"has no deployed releases", "normal for first deployment"},
		{"timed out waiting for the condition", "kubectl get pods"},
		{"connection refused", "Cannot connect to Kubernetes"},
		{"unable to connect to cluster", "kind get clusters"},
		{"chart not found", "Helm chart not found at: chart"},
	}
	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			base := errors.New(tt.message)
			got := EnhanceHelmError(base, "chart", "release")
			if !errors.Is(got, base) || !strings.Contains(got.Error(), tt.want) {
				t.Fatalf("EnhanceHelmError() = %v", got)
			}
		})
	}
	base := errors.New("unknown")
	if EnhanceHelmError(base, "chart", "release") != base {
		t.Fatal("unknown Helm error was replaced")
	}
}

func TestEnhancePodError(t *testing.T) {
	if EnhancePodError(nil, "pod", "ctx") != nil {
		t.Fatal("EnhancePodError(nil) returned an error")
	}
	tests := []struct {
		message string
		want    string
	}{
		{"ImagePullBackOff", "docker pull"},
		{"ErrImagePull", "image pull secrets"},
		{"CrashLoopBackOff", "kubectl --context ctx logs pod"},
		{"Pending", "Check PVC status"},
		{"pod not found", "get pods --all-namespaces"},
	}
	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			base := errors.New(tt.message)
			got := EnhancePodError(base, "pod", "ctx")
			if !errors.Is(got, base) || !strings.Contains(got.Error(), tt.want) {
				t.Fatalf("EnhancePodError() = %v", got)
			}
		})
	}
	base := errors.New("unknown")
	if EnhancePodError(base, "pod", "ctx") != base {
		t.Fatal("unknown pod error was replaced")
	}
}

func TestEnhanceClusterError(t *testing.T) {
	if EnhanceClusterError(nil, "kind") != nil {
		t.Fatal("EnhanceClusterError(nil) returned an error")
	}
	tests := []struct {
		message string
		want    string
	}{
		{"cluster already exists", "kind delete cluster"},
		{"command not found", "kind not installed"},
		{"executable file not found", "brew install kind"},
		{"failed to create cluster", "podman machine start"},
	}
	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			base := errors.New(tt.message)
			got := EnhanceClusterError(base, "kind")
			if !errors.Is(got, base) || !strings.Contains(got.Error(), tt.want) {
				t.Fatalf("EnhanceClusterError() = %v", got)
			}
		})
	}
	base := errors.New("unknown")
	if EnhanceClusterError(base, "kind") != base {
		t.Fatal("unknown cluster error was replaced")
	}
}

func TestCheckDependencies(t *testing.T) {
	t.Run("reports all missing tools", func(t *testing.T) {
		t.Setenv("PATH", t.TempDir())
		err := CheckDependencies("kind")
		if err == nil || !strings.Contains(err.Error(), "kubectl, helm, kind") || !strings.Contains(err.Error(), "--runtime kind") {
			t.Fatalf("CheckDependencies() error = %v", err)
		}
	})

	t.Run("runtime-specific tool", func(t *testing.T) {
		bin := t.TempDir()
		writeExecutable(t, bin, "kubectl")
		writeExecutable(t, bin, "helm")
		t.Setenv("PATH", bin)
		err := CheckDependencies("minikube")
		if err == nil || !strings.Contains(err.Error(), "minikube") {
			t.Fatalf("CheckDependencies() error = %v", err)
		}
	})

	t.Run("all present", func(t *testing.T) {
		bin := t.TempDir()
		for _, command := range []string{"kubectl", "helm", "kind"} {
			writeExecutable(t, bin, command)
		}
		t.Setenv("PATH", bin)
		if err := CheckDependencies("kind"); err != nil {
			t.Fatalf("CheckDependencies() error = %v", err)
		}
	})
}

func writeExecutable(t *testing.T, dir, name string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0700); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}
