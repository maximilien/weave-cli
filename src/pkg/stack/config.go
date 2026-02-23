// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package stack

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	// DefaultStackFile is the default name for stack configuration
	DefaultStackFile = "weave-stack.yaml"

	// DefaultStateDir is the directory for stack state
	DefaultStateDir = ".weave-state"
)

// LoadStackConfig loads and parses weave-stack.yaml
func LoadStackConfig(path string) (*StackConfig, error) {
	if path == "" {
		path = DefaultStackFile
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read stack config: %w", err)
	}

	var config StackConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse stack config: %w", err)
	}

	// Set defaults
	if err := setDefaults(&config); err != nil {
		return nil, fmt.Errorf("failed to set defaults: %w", err)
	}

	// Validate
	if err := ValidateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid stack config: %w", err)
	}

	return &config, nil
}

// SaveStackConfig saves configuration to YAML file
func SaveStackConfig(config *StackConfig, path string) error {
	if path == "" {
		path = DefaultStackFile
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// setDefaults sets default values for missing fields
func setDefaults(config *StackConfig) error {
	// Runtime defaults
	if config.Runtime.ContainerRuntime == "" {
		config.Runtime.ContainerRuntime = "podman" // OSS-first!
	}

	if config.Runtime.Kubernetes.Provider == "" {
		config.Runtime.Kubernetes.Provider = "kind" // Default to Kind for local dev
	}

	// Kind defaults
	if config.Runtime.Kubernetes.Provider == "kind" && config.Runtime.Kubernetes.Kind == nil {
		config.Runtime.Kubernetes.Kind = &KindConfig{
			Name:  "weave-stack",
			Nodes: 1,
		}
	}

	// Minikube defaults
	if config.Runtime.Kubernetes.Provider == "minikube" && config.Runtime.Kubernetes.Minikube == nil {
		config.Runtime.Kubernetes.Minikube = &MinikubeConfig{
			Driver: config.Runtime.ContainerRuntime, // Use configured runtime
			CPUs:   4,
			Memory: "16384", // 16GB
		}
	}

	// VectorDB defaults
	if config.Infrastructure.VectorDB.Resources.Requests.Memory == "" {
		config.Infrastructure.VectorDB.Resources.Requests.Memory = "8Gi"
	}
	if config.Infrastructure.VectorDB.Resources.Requests.CPU == "" {
		config.Infrastructure.VectorDB.Resources.Requests.CPU = "2"
	}
	if config.Infrastructure.VectorDB.Resources.Limits.Memory == "" {
		config.Infrastructure.VectorDB.Resources.Limits.Memory = "12Gi"
	}
	if config.Infrastructure.VectorDB.Resources.Limits.CPU == "" {
		config.Infrastructure.VectorDB.Resources.Limits.CPU = "4"
	}

	// K8s defaults for VectorDB
	if config.Infrastructure.VectorDB.Kubernetes == nil {
		config.Infrastructure.VectorDB.Kubernetes = &KubernetesResources{
			Deployment: DeploymentConfig{
				Replicas:     1,
				StorageClass: "standard",
				PVCSize:      "50Gi",
			},
		}
	}

	// Health check defaults
	if config.Infrastructure.VectorDB.Health == nil {
		config.Infrastructure.VectorDB.Health = &HealthCheck{
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
		}
	}

	return nil
}

// ValidateConfig validates the stack configuration
func ValidateConfig(config *StackConfig) error {
	// Required fields
	if config.Version == "" {
		return fmt.Errorf("version is required")
	}

	if config.Name == "" {
		return fmt.Errorf("name is required")
	}

	// Runtime validation
	validRuntimes := map[string]bool{"podman": true, "docker": true}
	if !validRuntimes[config.Runtime.ContainerRuntime] {
		return fmt.Errorf("invalid container_runtime: %s (must be podman or docker)", config.Runtime.ContainerRuntime)
	}

	validProviders := map[string]bool{"kind": true, "minikube": true, "eks": true, "gke": true}
	if !validProviders[config.Runtime.Kubernetes.Provider] {
		return fmt.Errorf("invalid kubernetes provider: %s", config.Runtime.Kubernetes.Provider)
	}

	// VectorDB validation
	if config.Infrastructure.VectorDB.Type == "" {
		return fmt.Errorf("infrastructure.vectordb.type is required")
	}

	validVDBTypes := map[string]bool{"milvus": true, "qdrant": true, "weaviate": true}
	if !validVDBTypes[config.Infrastructure.VectorDB.Type] {
		return fmt.Errorf("invalid vectordb type: %s", config.Infrastructure.VectorDB.Type)
	}

	if config.Infrastructure.VectorDB.Version == "" {
		return fmt.Errorf("infrastructure.vectordb.version is required")
	}

	return nil
}

// LoadClusterInfo loads cluster information from state directory
func LoadClusterInfo() (*ClusterInfo, error) {
	stateFile := filepath.Join(DefaultStateDir, "cluster.json")

	data, err := os.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No cluster info yet
		}
		return nil, fmt.Errorf("failed to read cluster info: %w", err)
	}

	var info ClusterInfo
	if err := yaml.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("failed to parse cluster info: %w", err)
	}

	return &info, nil
}

// SaveClusterInfo saves cluster information to state directory
func SaveClusterInfo(info *ClusterInfo) error {
	// Ensure state directory exists
	if err := os.MkdirAll(DefaultStateDir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	stateFile := filepath.Join(DefaultStateDir, "cluster.json")

	data, err := yaml.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal cluster info: %w", err)
	}

	if err := os.WriteFile(stateFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write cluster info: %w", err)
	}

	return nil
}

// StackExists checks if weave-stack.yaml exists in current directory
func StackExists() bool {
	_, err := os.Stat(DefaultStackFile)
	return err == nil
}
