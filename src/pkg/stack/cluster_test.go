// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package stack

import (
	"os"
	"path/filepath"
	"strings"
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

func TestCreateKindClusterWithLocalCommandFixtures(t *testing.T) {
	t.Run("creates and records a cluster", func(t *testing.T) {
		workDir := t.TempDir()
		t.Chdir(workDir)
		binDir := filepath.Join(workDir, "bin")
		if err := os.Mkdir(binDir, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(DefaultStateDir, 0o700); err != nil {
			t.Fatal(err)
		}
		writeExecutable(t, binDir, "podman")
		if err := os.WriteFile(filepath.Join(binDir, "kind"), []byte("#!/bin/sh\nif [ \"$1\" = get ]; then printf 'unrelated\\n'; fi\nexit 0\n"), 0o700); err != nil {
			t.Fatal(err)
		}
		t.Setenv("PATH", binDir)

		info, err := CreateKindCluster(&StackConfig{Runtime: RuntimeConfig{
			ContainerRuntime: "podman",
			Kubernetes:       KubernetesConfig{Provider: "kind", Kind: &KindConfig{Name: "fixture", Nodes: 1}},
		}})
		if err != nil {
			t.Fatalf("CreateKindCluster() error = %v", err)
		}
		if info.Name != "fixture" || info.Context != "kind-fixture" || info.ContainerRuntime != "podman" || info.Status != "active" {
			t.Fatalf("unexpected cluster info: %#v", info)
		}
		stored, err := LoadClusterInfo()
		if err != nil || stored == nil || stored.Name != info.Name {
			t.Fatalf("LoadClusterInfo() = (%#v, %v)", stored, err)
		}
	})

	t.Run("reports cluster creation failure output", func(t *testing.T) {
		workDir := t.TempDir()
		t.Chdir(workDir)
		binDir := filepath.Join(workDir, "bin")
		if err := os.Mkdir(binDir, 0o700); err != nil {
			t.Fatal(err)
		}
		writeExecutable(t, binDir, "podman")
		if err := os.WriteFile(filepath.Join(binDir, "kind"), []byte("#!/bin/sh\nif [ \"$1\" = get ]; then exit 0; fi\necho fixture-failure >&2\nexit 1\n"), 0o700); err != nil {
			t.Fatal(err)
		}
		t.Setenv("PATH", binDir)

		info, err := CreateKindCluster(&StackConfig{Runtime: RuntimeConfig{
			ContainerRuntime: "podman",
			Kubernetes:       KubernetesConfig{Provider: "kind", Kind: &KindConfig{Name: "fixture", Nodes: 1}},
		}})
		if info != nil || err == nil || !strings.Contains(err.Error(), "fixture-failure") {
			t.Fatalf("CreateKindCluster() = (%#v, %v), want command failure", info, err)
		}
	})

	t.Run("uses a custom kind config", func(t *testing.T) {
		workDir := t.TempDir()
		t.Chdir(workDir)
		binDir := filepath.Join(workDir, "bin")
		if err := os.Mkdir(binDir, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(DefaultStateDir, 0o700); err != nil {
			t.Fatal(err)
		}
		writeExecutable(t, binDir, "podman")
		if err := os.WriteFile(filepath.Join(binDir, "kind"), []byte("#!/bin/sh\nif [ \"$1\" = get ]; then exit 0; fi\nexit 0\n"), 0o700); err != nil {
			t.Fatal(err)
		}
		t.Setenv("PATH", binDir)
		customYAML := "kind: Cluster\napiVersion: kind.x-k8s.io/v1alpha4\n"
		info, err := CreateKindCluster(&StackConfig{Runtime: RuntimeConfig{
			ContainerRuntime: "podman",
			Kubernetes: KubernetesConfig{Provider: "kind", Kind: &KindConfig{
				Name: "custom", Nodes: 1, Config: customYAML,
			}},
		}})
		if err != nil || info == nil {
			t.Fatalf("CreateKindCluster(custom config) = (%#v, %v)", info, err)
		}
		data, err := os.ReadFile(filepath.Join(DefaultStateDir, "kind-config.yaml"))
		if err != nil || string(data) != customYAML {
			t.Fatalf("custom kind config = %q, %v", data, err)
		}
	})

	t.Run("detects an existing cluster before writing config", func(t *testing.T) {
		workDir := t.TempDir()
		t.Chdir(workDir)
		binDir := filepath.Join(workDir, "bin")
		if err := os.Mkdir(binDir, 0o700); err != nil {
			t.Fatal(err)
		}
		writeExecutable(t, binDir, "podman")
		if err := os.WriteFile(filepath.Join(binDir, "kind"), []byte("#!/bin/sh\nprintf 'fixture\\n'\n"), 0o700); err != nil {
			t.Fatal(err)
		}
		t.Setenv("PATH", binDir)
		config := &StackConfig{Runtime: RuntimeConfig{
			ContainerRuntime: "podman",
			Kubernetes: KubernetesConfig{Provider: "kind", Kind: &KindConfig{
				Name: "fixture", Nodes: 1, Config: "kind: Cluster\napiVersion: kind.x-k8s.io/v1alpha4\n",
			}},
		}}
		info, err := CreateKindCluster(config)
		if info != nil || err == nil || !strings.Contains(err.Error(), "already exists") {
			t.Fatalf("CreateKindCluster() = (%#v, %v), want existing-cluster error", info, err)
		}
		if _, err := os.Stat(filepath.Join(DefaultStateDir, "kind-config.yaml")); !os.IsNotExist(err) {
			t.Fatalf("custom config should not be written after existing-cluster check, stat error = %v", err)
		}
	})
}

func TestCreateMinikubeClusterWithLocalCommandFixture(t *testing.T) {
	workDir := t.TempDir()
	t.Chdir(workDir)
	binDir := filepath.Join(workDir, "bin")
	if err := os.Mkdir(binDir, 0o700); err != nil {
		t.Fatal(err)
	}
	writeExecutable(t, binDir, "docker")
	writeExecutable(t, binDir, "minikube")
	t.Setenv("PATH", binDir)

	info, err := CreateMinikubeCluster(&StackConfig{Runtime: RuntimeConfig{
		ContainerRuntime: "docker",
		Kubernetes: KubernetesConfig{Provider: "minikube", Minikube: &MinikubeConfig{
			Driver: "docker", CPUs: 2, Memory: "4g", Addons: []string{"ingress", "metrics-server"},
		}},
	}})
	if err != nil {
		t.Fatalf("CreateMinikubeCluster() error = %v", err)
	}
	if info.Name != "minikube" || info.Context != "minikube" || info.ContainerRuntime != "docker" {
		t.Fatalf("unexpected cluster info: %#v", info)
	}
	stored, err := LoadClusterInfo()
	if err != nil || stored == nil || stored.Provider != "minikube" {
		t.Fatalf("LoadClusterInfo() = (%#v, %v)", stored, err)
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
