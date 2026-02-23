// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package stack

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestGenerateHelmValues(t *testing.T) {
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
				Resources: ResourceRequirements{
					Requests: ResourceLimits{
						Memory: "8Gi",
						CPU:    "2",
					},
					Limits: ResourceLimits{
						Memory: "12Gi",
						CPU:    "4",
					},
				},
				Kubernetes: &KubernetesResources{
					Deployment: DeploymentConfig{
						Replicas:     1,
						StorageClass: "standard",
						PVCSize:      "50Gi",
					},
				},
				Health: &HealthCheck{
					Liveness: &Probe{
						HTTPGet: &HTTPGetAction{
							Path: "/healthz",
							Port: 9091,
						},
						InitialDelaySeconds: 30,
						PeriodSeconds:       30,
						FailureThreshold:    3,
					},
					Readiness: &Probe{
						HTTPGet: &HTTPGetAction{
							Path: "/healthz",
							Port: 9091,
						},
						InitialDelaySeconds: 10,
						PeriodSeconds:       10,
					},
				},
			},
		},
	}

	values, err := GenerateHelmValues(config)
	require.NoError(t, err)
	assert.NotNil(t, values)

	// Check stack metadata
	assert.Equal(t, "test-stack", values.Stack.Name)
	assert.Equal(t, "1.0", values.Stack.Version)

	// Check VectorDB config
	assert.Equal(t, "milvus", values.VectorDB.Type)
	assert.Equal(t, "2.3.0", values.VectorDB.Version)
	assert.Equal(t, 1, values.VectorDB.Replicas)

	// Check resources
	assert.Equal(t, "8Gi", values.VectorDB.Resources.Requests.Memory)
	assert.Equal(t, "2", values.VectorDB.Resources.Requests.CPU)
	assert.Equal(t, "12Gi", values.VectorDB.Resources.Limits.Memory)
	assert.Equal(t, "4", values.VectorDB.Resources.Limits.CPU)

	// Check storage
	assert.Equal(t, "standard", values.VectorDB.Storage.Class)
	assert.Equal(t, "50Gi", values.VectorDB.Storage.Size)

	// Check service
	assert.Equal(t, "ClusterIP", values.VectorDB.Service.Type)
	assert.Equal(t, 19530, values.VectorDB.Service.Port)
	assert.Equal(t, 9091, values.VectorDB.Service.MetricsPort)

	// Check health checks
	assert.Equal(t, "/healthz", values.VectorDB.Health.Liveness.Path)
	assert.Equal(t, 9091, values.VectorDB.Health.Liveness.Port)
	assert.Equal(t, 30, values.VectorDB.Health.Liveness.InitialDelaySeconds)
	assert.Equal(t, 30, values.VectorDB.Health.Liveness.PeriodSeconds)
	assert.Equal(t, 3, values.VectorDB.Health.Liveness.FailureThreshold)

	assert.Equal(t, "/healthz", values.VectorDB.Health.Readiness.Path)
	assert.Equal(t, 9091, values.VectorDB.Health.Readiness.Port)
	assert.Equal(t, 10, values.VectorDB.Health.Readiness.InitialDelaySeconds)
	assert.Equal(t, 10, values.VectorDB.Health.Readiness.PeriodSeconds)
}

func TestGenerateHelmValuesWithImageStorage(t *testing.T) {
	config := &StackConfig{
		Version: "1.0",
		Name:    "multimodal-stack",
		Infrastructure: InfrastructureConfig{
			VectorDB: VectorDBConfig{
				Type:    "milvus",
				Version: "2.3.0",
				Resources: ResourceRequirements{
					Requests: ResourceLimits{Memory: "8Gi", CPU: "2"},
					Limits:   ResourceLimits{Memory: "12Gi", CPU: "4"},
				},
			},
			ImageStorage: &ImageStorageConfig{
				Type:   "minio",
				Bucket: "weave-images",
			},
		},
	}

	values, err := GenerateHelmValues(config)
	require.NoError(t, err)
	assert.NotNil(t, values)

	// Check image storage
	require.NotNil(t, values.ImageStorage)
	assert.True(t, values.ImageStorage.Enabled)
	assert.Equal(t, "minio", values.ImageStorage.Type)
	assert.Equal(t, 1, values.ImageStorage.Replicas)
}

func TestSaveHelmValues(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "kubernetes", "values.yaml")

	values := &HelmValues{
		Stack: HelmStackMeta{
			Name:    "test-stack",
			Version: "1.0",
		},
		VectorDB: HelmVectorDB{
			Type:     "milvus",
			Version:  "2.3.0",
			Replicas: 1,
			Resources: HelmResources{
				Requests: HelmResourceLimits{Memory: "8Gi", CPU: "2"},
				Limits:   HelmResourceLimits{Memory: "12Gi", CPU: "4"},
			},
			Storage: HelmStorage{
				Class: "standard",
				Size:  "50Gi",
			},
			Service: HelmService{
				Type:        "ClusterIP",
				Port:        19530,
				MetricsPort: 9091,
			},
		},
	}

	err := SaveHelmValues(values, outputPath)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(outputPath)
	require.NoError(t, err)

	// Load and verify content
	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	var loaded HelmValues
	err = yaml.Unmarshal(data, &loaded)
	require.NoError(t, err)

	assert.Equal(t, values.Stack.Name, loaded.Stack.Name)
	assert.Equal(t, values.VectorDB.Type, loaded.VectorDB.Type)
	assert.Equal(t, values.VectorDB.Version, loaded.VectorDB.Version)
}

func TestGenerateHelmChart(t *testing.T) {
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "helm-chart")

	config := &StackConfig{
		Version: "1.0",
		Name:    "test-stack",
		Infrastructure: InfrastructureConfig{
			VectorDB: VectorDBConfig{
				Type:    "milvus",
				Version: "2.3.0",
				Resources: ResourceRequirements{
					Requests: ResourceLimits{Memory: "8Gi", CPU: "2"},
					Limits:   ResourceLimits{Memory: "12Gi", CPU: "4"},
				},
			},
		},
	}

	err := GenerateHelmChart(config, outputDir)
	require.NoError(t, err)

	// Verify values.yaml exists
	valuesPath := filepath.Join(outputDir, "values.yaml")
	_, err = os.Stat(valuesPath)
	require.NoError(t, err)
}
