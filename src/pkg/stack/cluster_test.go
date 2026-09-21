// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package stack

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateDefaultKindConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Change to temp dir
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldWd)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	tests := []struct {
		name       string
		kindConfig *KindConfig
		wantNodes  int
	}{
		{
			name: "single node cluster",
			kindConfig: &KindConfig{
				Name:  "test-cluster",
				Nodes: 1,
			},
			wantNodes: 1,
		},
		{
			name: "multi-node cluster",
			kindConfig: &KindConfig{
				Name:  "test-cluster",
				Nodes: 3,
			},
			wantNodes: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath, err := generateDefaultKindConfig(tt.kindConfig)
			require.NoError(t, err)
			assert.NotEmpty(t, configPath)

			// Verify file exists
			_, err = os.Stat(configPath)
			require.NoError(t, err)

			// Read and verify content
			content, err := os.ReadFile(configPath)
			require.NoError(t, err)
			assert.Contains(t, string(content), "kind: Cluster")
			assert.Contains(t, string(content), "apiVersion: kind.x-k8s.io/v1alpha4")

			// Count nodes in config
			// Should have 1 control-plane + (Nodes-1) workers
			contentStr := string(content)
			assert.Contains(t, contentStr, "role: control-plane")
		})
	}
}

func TestClusterExists(t *testing.T) {
	// This test checks if clusterExists works correctly
	// We can't assume any specific cluster exists, but we can test the function logic

	tests := []struct {
		name        string
		clusterName string
		// We can't assert specific values since it depends on system state
	}{
		{"check nonexistent cluster", "nonexistent-cluster-12345"},
		{"check with special chars", "test-cluster-!@#"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists := clusterExists(tt.clusterName)
			// Just verify the function doesn't panic
			t.Logf("Cluster %s exists: %v", tt.clusterName, exists)
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		substr string
		want   bool
	}{
		{"exact match", "test", "test", true},
		{"substring at start", "test-cluster", "test", true},
		{"substring at end", "kind-test", "test", true},
		{"substring in middle", "my-test-cluster", "test", true},
		{"no match", "cluster", "test", false},
		{"empty string", "", "test", false},
		{"empty substring", "test", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := contains(tt.s, tt.substr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCreateKindCluster_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  *StackConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "invalid provider",
			config: &StackConfig{
				Runtime: RuntimeConfig{
					Kubernetes: KubernetesConfig{
						Provider: "minikube",
					},
				},
			},
			wantErr: true,
			errMsg:  "invalid provider for Kind cluster",
		},
		{
			name: "missing kind config",
			config: &StackConfig{
				Runtime: RuntimeConfig{
					Kubernetes: KubernetesConfig{
						Provider: "kind",
					},
				},
			},
			wantErr: true,
			errMsg:  "kind configuration is missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CreateKindCluster(tt.config)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				// Skip actual cluster creation in unit tests
				t.Skip("Skipping actual cluster creation in unit test")
			}
		})
	}
}

func TestCreateMinikubeCluster_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  *StackConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "invalid provider",
			config: &StackConfig{
				Runtime: RuntimeConfig{
					Kubernetes: KubernetesConfig{
						Provider: "kind",
					},
				},
			},
			wantErr: true,
			errMsg:  "invalid provider for Minikube cluster",
		},
		{
			name: "missing minikube config",
			config: &StackConfig{
				Runtime: RuntimeConfig{
					Kubernetes: KubernetesConfig{
						Provider: "minikube",
					},
				},
			},
			wantErr: true,
			errMsg:  "minikube configuration is missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CreateMinikubeCluster(tt.config)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				// Skip actual cluster creation in unit tests
				t.Skip("Skipping actual cluster creation in unit test")
			}
		})
	}
}

func TestDeleteCluster(t *testing.T) {
	tests := []struct {
		name    string
		info    *ClusterInfo
		wantErr bool
		errMsg  string
	}{
		{
			name: "unsupported provider",
			info: &ClusterInfo{
				Provider: "eks",
			},
			wantErr: true,
			errMsg:  "unsupported provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := DeleteCluster(tt.info)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			}
		})
	}
}

func TestClusterStatusAndDeleteCommandPaths(t *testing.T) {
	binDir := t.TempDir()
	scripts := map[string]string{
		"kind": `#!/bin/sh
if [ "$1" = "get" ]; then
  if [ "$KIND_MODE" = "missing" ]; then echo other; else echo weave-test; fi
  exit 0
fi
if [ "$KIND_MODE" = "fail" ]; then echo failed; exit 1; fi
`,
		"kubectl": `#!/bin/sh
if [ "$KUBE_FAIL" = "true" ]; then exit 1; fi
`,
		"minikube": `#!/bin/sh
if [ "$1" = "status" ]; then
  if [ "$MINI_MODE" = "fail" ]; then exit 1; fi
  if [ "$MINI_MODE" = "running" ]; then echo Running; else echo Stopped; fi
  exit 0
fi
if [ "$MINI_MODE" = "fail" ]; then echo failed; exit 1; fi
`,
	}
	for name, script := range scripts {
		if err := os.WriteFile(filepath.Join(binDir, name), []byte(script), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", binDir)

	status, err := GetClusterStatus(&ClusterInfo{Provider: "kind", Name: "weave-test"})
	assert.NoError(t, err)
	assert.Equal(t, "active", status)
	t.Setenv("KUBE_FAIL", "true")
	status, err = GetClusterStatus(&ClusterInfo{Provider: "kind", Name: "weave-test"})
	assert.NoError(t, err)
	assert.Equal(t, "error", status)
	t.Setenv("KIND_MODE", "missing")
	status, err = GetClusterStatus(&ClusterInfo{Provider: "kind", Name: "weave-test"})
	assert.NoError(t, err)
	assert.Equal(t, "stopped", status)

	for _, test := range []struct {
		mode string
		want string
	}{
		{mode: "running", want: "active"},
		{mode: "stopped", want: "stopped"},
		{mode: "fail", want: "stopped"},
	} {
		t.Setenv("MINI_MODE", test.mode)
		status, err = GetClusterStatus(&ClusterInfo{Provider: "minikube"})
		assert.NoError(t, err)
		assert.Equal(t, test.want, status)
	}
	_, err = GetClusterStatus(&ClusterInfo{Provider: "unsupported"})
	assert.Error(t, err)

	t.Setenv("KIND_MODE", "ok")
	assert.NoError(t, DeleteCluster(&ClusterInfo{Provider: "kind", Name: "weave-test"}))
	t.Setenv("MINI_MODE", "running")
	assert.NoError(t, DeleteCluster(&ClusterInfo{Provider: "minikube"}))
	t.Setenv("KIND_MODE", "fail")
	assert.Error(t, DeleteCluster(&ClusterInfo{Provider: "kind", Name: "weave-test"}))
	t.Setenv("MINI_MODE", "fail")
	assert.Error(t, DeleteCluster(&ClusterInfo{Provider: "minikube"}))
}
