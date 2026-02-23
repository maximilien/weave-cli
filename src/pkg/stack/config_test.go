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

func TestLoadStackConfig(t *testing.T) {
	tests := []struct {
		name        string
		yaml        string
		wantErr     bool
		errContains string
	}{
		{
			name: "valid minimal config",
			yaml: `version: "1.0"
name: test-stack
runtime:
  kubernetes:
    provider: kind
  container_runtime: podman
infrastructure:
  vectordb:
    type: milvus
    version: 2.3.0
`,
			wantErr: false,
		},
		{
			name: "missing version",
			yaml: `name: test-stack
runtime:
  kubernetes:
    provider: kind
infrastructure:
  vectordb:
    type: milvus
    version: 2.3.0
`,
			wantErr:     true,
			errContains: "version is required",
		},
		{
			name: "missing name",
			yaml: `version: "1.0"
runtime:
  kubernetes:
    provider: kind
infrastructure:
  vectordb:
    type: milvus
    version: 2.3.0
`,
			wantErr:     true,
			errContains: "name is required",
		},
		{
			name: "missing vectordb type",
			yaml: `version: "1.0"
name: test-stack
runtime:
  kubernetes:
    provider: kind
infrastructure:
  vectordb:
    version: 2.3.0
`,
			wantErr:     true,
			errContains: "infrastructure.vectordb.type is required",
		},
		{
			name: "invalid vectordb type",
			yaml: `version: "1.0"
name: test-stack
runtime:
  kubernetes:
    provider: kind
infrastructure:
  vectordb:
    type: invalid-db
    version: 2.3.0
`,
			wantErr:     true,
			errContains: "invalid vectordb type",
		},
		{
			name: "invalid kubernetes provider",
			yaml: `version: "1.0"
name: test-stack
runtime:
  kubernetes:
    provider: invalid
infrastructure:
  vectordb:
    type: milvus
    version: 2.3.0
`,
			wantErr:     true,
			errContains: "invalid kubernetes provider",
		},
		{
			name: "invalid container runtime",
			yaml: `version: "1.0"
name: test-stack
runtime:
  kubernetes:
    provider: kind
  container_runtime: containerd
infrastructure:
  vectordb:
    type: milvus
    version: 2.3.0
`,
			wantErr:     true,
			errContains: "invalid container_runtime",
		},
		{
			name:        "invalid yaml syntax",
			yaml:        `invalid: yaml: syntax: [`,
			wantErr:     true,
			errContains: "failed to parse stack config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "weave-stack.yaml")
			err := os.WriteFile(configPath, []byte(tt.yaml), 0644)
			require.NoError(t, err)

			// Load config
			config, err := LoadStackConfig(configPath)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, config)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, config)
			}
		})
	}
}

func TestSetDefaults(t *testing.T) {
	tests := []struct {
		name   string
		config StackConfig
		check  func(t *testing.T, cfg *StackConfig)
	}{
		{
			name: "sets podman as default container runtime",
			config: StackConfig{
				Version: "1.0",
				Name:    "test",
				Runtime: RuntimeConfig{
					Kubernetes: KubernetesConfig{
						Provider: "kind",
					},
				},
				Infrastructure: InfrastructureConfig{
					VectorDB: VectorDBConfig{
						Type:    "milvus",
						Version: "2.3.0",
					},
				},
			},
			check: func(t *testing.T, cfg *StackConfig) {
				assert.Equal(t, "podman", cfg.Runtime.ContainerRuntime)
			},
		},
		{
			name: "sets kind as default provider",
			config: StackConfig{
				Version: "1.0",
				Name:    "test",
				Runtime: RuntimeConfig{
					ContainerRuntime: "podman",
				},
				Infrastructure: InfrastructureConfig{
					VectorDB: VectorDBConfig{
						Type:    "milvus",
						Version: "2.3.0",
					},
				},
			},
			check: func(t *testing.T, cfg *StackConfig) {
				assert.Equal(t, "kind", cfg.Runtime.Kubernetes.Provider)
			},
		},
		{
			name: "sets kind config when provider is kind",
			config: StackConfig{
				Version: "1.0",
				Name:    "test",
				Runtime: RuntimeConfig{
					ContainerRuntime: "podman",
					Kubernetes: KubernetesConfig{
						Provider: "kind",
					},
				},
				Infrastructure: InfrastructureConfig{
					VectorDB: VectorDBConfig{
						Type:    "milvus",
						Version: "2.3.0",
					},
				},
			},
			check: func(t *testing.T, cfg *StackConfig) {
				require.NotNil(t, cfg.Runtime.Kubernetes.Kind)
				assert.Equal(t, "weave-stack", cfg.Runtime.Kubernetes.Kind.Name)
				assert.Equal(t, 1, cfg.Runtime.Kubernetes.Kind.Nodes)
			},
		},
		{
			name: "sets minikube config when provider is minikube",
			config: StackConfig{
				Version: "1.0",
				Name:    "test",
				Runtime: RuntimeConfig{
					ContainerRuntime: "podman",
					Kubernetes: KubernetesConfig{
						Provider: "minikube",
					},
				},
				Infrastructure: InfrastructureConfig{
					VectorDB: VectorDBConfig{
						Type:    "milvus",
						Version: "2.3.0",
					},
				},
			},
			check: func(t *testing.T, cfg *StackConfig) {
				require.NotNil(t, cfg.Runtime.Kubernetes.Minikube)
				assert.Equal(t, "podman", cfg.Runtime.Kubernetes.Minikube.Driver)
				assert.Equal(t, 4, cfg.Runtime.Kubernetes.Minikube.CPUs)
				assert.Equal(t, "16384", cfg.Runtime.Kubernetes.Minikube.Memory)
			},
		},
		{
			name: "sets vectordb resource defaults",
			config: StackConfig{
				Version: "1.0",
				Name:    "test",
				Runtime: RuntimeConfig{
					ContainerRuntime: "podman",
					Kubernetes: KubernetesConfig{
						Provider: "kind",
					},
				},
				Infrastructure: InfrastructureConfig{
					VectorDB: VectorDBConfig{
						Type:    "milvus",
						Version: "2.3.0",
					},
				},
			},
			check: func(t *testing.T, cfg *StackConfig) {
				assert.Equal(t, "8Gi", cfg.Infrastructure.VectorDB.Resources.Requests.Memory)
				assert.Equal(t, "2", cfg.Infrastructure.VectorDB.Resources.Requests.CPU)
				assert.Equal(t, "12Gi", cfg.Infrastructure.VectorDB.Resources.Limits.Memory)
				assert.Equal(t, "4", cfg.Infrastructure.VectorDB.Resources.Limits.CPU)
			},
		},
		{
			name: "sets kubernetes deployment defaults",
			config: StackConfig{
				Version: "1.0",
				Name:    "test",
				Runtime: RuntimeConfig{
					ContainerRuntime: "podman",
					Kubernetes: KubernetesConfig{
						Provider: "kind",
					},
				},
				Infrastructure: InfrastructureConfig{
					VectorDB: VectorDBConfig{
						Type:    "milvus",
						Version: "2.3.0",
					},
				},
			},
			check: func(t *testing.T, cfg *StackConfig) {
				require.NotNil(t, cfg.Infrastructure.VectorDB.Kubernetes)
				assert.Equal(t, 1, cfg.Infrastructure.VectorDB.Kubernetes.Deployment.Replicas)
				assert.Equal(t, "standard", cfg.Infrastructure.VectorDB.Kubernetes.Deployment.StorageClass)
				assert.Equal(t, "50Gi", cfg.Infrastructure.VectorDB.Kubernetes.Deployment.PVCSize)
			},
		},
		{
			name: "sets health check defaults",
			config: StackConfig{
				Version: "1.0",
				Name:    "test",
				Runtime: RuntimeConfig{
					ContainerRuntime: "podman",
					Kubernetes: KubernetesConfig{
						Provider: "kind",
					},
				},
				Infrastructure: InfrastructureConfig{
					VectorDB: VectorDBConfig{
						Type:    "milvus",
						Version: "2.3.0",
					},
				},
			},
			check: func(t *testing.T, cfg *StackConfig) {
				require.NotNil(t, cfg.Infrastructure.VectorDB.Health)
				require.NotNil(t, cfg.Infrastructure.VectorDB.Health.Liveness)
				require.NotNil(t, cfg.Infrastructure.VectorDB.Health.Liveness.HTTPGet)
				assert.Equal(t, "/healthz", cfg.Infrastructure.VectorDB.Health.Liveness.HTTPGet.Path)
				assert.Equal(t, 9091, cfg.Infrastructure.VectorDB.Health.Liveness.HTTPGet.Port)
				assert.Equal(t, 30, cfg.Infrastructure.VectorDB.Health.Liveness.InitialDelaySeconds)

				require.NotNil(t, cfg.Infrastructure.VectorDB.Health.Readiness)
				require.NotNil(t, cfg.Infrastructure.VectorDB.Health.Readiness.HTTPGet)
				assert.Equal(t, "/healthz", cfg.Infrastructure.VectorDB.Health.Readiness.HTTPGet.Path)
				assert.Equal(t, 9091, cfg.Infrastructure.VectorDB.Health.Readiness.HTTPGet.Port)
				assert.Equal(t, 10, cfg.Infrastructure.VectorDB.Health.Readiness.InitialDelaySeconds)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.config
			err := setDefaults(&cfg)
			require.NoError(t, err)
			tt.check(t, &cfg)
		})
	}
}

func TestValidateConfig(t *testing.T) {
	validConfig := &StackConfig{
		Version: "1.0",
		Name:    "test-stack",
		Runtime: RuntimeConfig{
			ContainerRuntime: "podman",
			Kubernetes: KubernetesConfig{
				Provider: "kind",
			},
		},
		Infrastructure: InfrastructureConfig{
			VectorDB: VectorDBConfig{
				Type:    "milvus",
				Version: "2.3.0",
			},
		},
	}

	tests := []struct {
		name        string
		modify      func(*StackConfig)
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid config",
			modify:  func(c *StackConfig) {},
			wantErr: false,
		},
		{
			name: "missing version",
			modify: func(c *StackConfig) {
				c.Version = ""
			},
			wantErr:     true,
			errContains: "version is required",
		},
		{
			name: "missing name",
			modify: func(c *StackConfig) {
				c.Name = ""
			},
			wantErr:     true,
			errContains: "name is required",
		},
		{
			name: "invalid container runtime",
			modify: func(c *StackConfig) {
				c.Runtime.ContainerRuntime = "containerd"
			},
			wantErr:     true,
			errContains: "invalid container_runtime",
		},
		{
			name: "invalid kubernetes provider",
			modify: func(c *StackConfig) {
				c.Runtime.Kubernetes.Provider = "openshift"
			},
			wantErr:     true,
			errContains: "invalid kubernetes provider",
		},
		{
			name: "missing vectordb type",
			modify: func(c *StackConfig) {
				c.Infrastructure.VectorDB.Type = ""
			},
			wantErr:     true,
			errContains: "infrastructure.vectordb.type is required",
		},
		{
			name: "invalid vectordb type",
			modify: func(c *StackConfig) {
				c.Infrastructure.VectorDB.Type = "invalid"
			},
			wantErr:     true,
			errContains: "invalid vectordb type",
		},
		{
			name: "missing vectordb version",
			modify: func(c *StackConfig) {
				c.Infrastructure.VectorDB.Version = ""
			},
			wantErr:     true,
			errContains: "infrastructure.vectordb.version is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy of valid config
			cfg := *validConfig
			tt.modify(&cfg)

			err := ValidateConfig(&cfg)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSaveStackConfig(t *testing.T) {
	config := &StackConfig{
		Version: "1.0",
		Name:    "test-stack",
		Runtime: RuntimeConfig{
			ContainerRuntime: "podman",
			Kubernetes: KubernetesConfig{
				Provider: "kind",
			},
		},
		Infrastructure: InfrastructureConfig{
			VectorDB: VectorDBConfig{
				Type:    "milvus",
				Version: "2.3.0",
			},
		},
	}

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "weave-stack.yaml")

	// Save config
	err := SaveStackConfig(config, configPath)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(configPath)
	require.NoError(t, err)

	// Load it back and verify
	loaded, err := LoadStackConfig(configPath)
	require.NoError(t, err)
	assert.Equal(t, config.Version, loaded.Version)
	assert.Equal(t, config.Name, loaded.Name)
	assert.Equal(t, config.Infrastructure.VectorDB.Type, loaded.Infrastructure.VectorDB.Type)
}

func TestStackExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Change to temp dir
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldWd)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Should not exist initially
	assert.False(t, StackExists())

	// Create stack file
	err = os.WriteFile(DefaultStackFile, []byte("test"), 0644)
	require.NoError(t, err)

	// Should exist now
	assert.True(t, StackExists())
}

func TestLoadClusterInfo(t *testing.T) {
	tmpDir := t.TempDir()

	// Change to temp dir
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldWd)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Should return nil when no cluster info exists
	info, err := LoadClusterInfo()
	require.NoError(t, err)
	assert.Nil(t, info)

	// Create cluster info
	testInfo := &ClusterInfo{
		Name:             "test-cluster",
		Provider:         "kind",
		Context:          "kind-test",
		ContainerRuntime: "podman",
		Status:           "active",
	}

	err = SaveClusterInfo(testInfo)
	require.NoError(t, err)

	// Load it back
	loaded, err := LoadClusterInfo()
	require.NoError(t, err)
	require.NotNil(t, loaded)
	assert.Equal(t, testInfo.Name, loaded.Name)
	assert.Equal(t, testInfo.Provider, loaded.Provider)
	assert.Equal(t, testInfo.Context, loaded.Context)
}
