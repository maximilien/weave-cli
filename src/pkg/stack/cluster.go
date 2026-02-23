// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package stack

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// CreateKindCluster creates a Kind cluster with the specified configuration
func CreateKindCluster(config *StackConfig) (*ClusterInfo, error) {
	if config.Runtime.Kubernetes.Provider != "kind" {
		return nil, fmt.Errorf("invalid provider for Kind cluster: %s", config.Runtime.Kubernetes.Provider)
	}

	// Get Kind config
	kindConfig := config.Runtime.Kubernetes.Kind
	if kindConfig == nil {
		return nil, fmt.Errorf("kind configuration is missing")
	}

	// Detect container runtime
	runtime, err := GetRuntimeCommand(config.Runtime.ContainerRuntime)
	if err != nil {
		return nil, fmt.Errorf("failed to detect container runtime: %w", err)
	}

	// Check if cluster already exists
	if clusterExists(kindConfig.Name) {
		return nil, fmt.Errorf("cluster '%s' already exists (use 'weave stack down' to remove it first)", kindConfig.Name)
	}

	// Generate Kind config file if custom config specified
	var configPath string
	if kindConfig.Config != "" {
		configPath = filepath.Join(DefaultStateDir, "kind-config.yaml")
		if err := os.WriteFile(configPath, []byte(kindConfig.Config), 0644); err != nil {
			return nil, fmt.Errorf("failed to write kind config: %w", err)
		}
	} else {
		// Generate default Kind config
		configPath, err = generateDefaultKindConfig(kindConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to generate kind config: %w", err)
		}
	}

	// Build kind create command
	args := []string{
		"create", "cluster",
		"--name", kindConfig.Name,
		"--config", configPath,
	}

	cmd := exec.Command("kind", args...)

	// Set environment for podman if detected
	if runtime == RuntimePodman {
		cmd.Env = append(os.Environ(), "KIND_EXPERIMENTAL_PROVIDER=podman")
	}

	// Run command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to create kind cluster: %w\nOutput: %s", err, string(output))
	}

	// Get kubectl context
	context := fmt.Sprintf("kind-%s", kindConfig.Name)

	// Create cluster info
	info := &ClusterInfo{
		Name:             kindConfig.Name,
		Provider:         "kind",
		Context:          context,
		ContainerRuntime: string(runtime),
		CreatedAt:        time.Now(),
		Status:           "active",
	}

	// Save cluster info
	if err := SaveClusterInfo(info); err != nil {
		return nil, fmt.Errorf("failed to save cluster info: %w", err)
	}

	return info, nil
}

// CreateMinikubeCluster creates a Minikube cluster with the specified configuration
func CreateMinikubeCluster(config *StackConfig) (*ClusterInfo, error) {
	if config.Runtime.Kubernetes.Provider != "minikube" {
		return nil, fmt.Errorf("invalid provider for Minikube cluster: %s", config.Runtime.Kubernetes.Provider)
	}

	minikubeConfig := config.Runtime.Kubernetes.Minikube
	if minikubeConfig == nil {
		return nil, fmt.Errorf("minikube configuration is missing")
	}

	// Detect container runtime
	runtime, err := GetRuntimeCommand(config.Runtime.ContainerRuntime)
	if err != nil {
		return nil, fmt.Errorf("failed to detect container runtime: %w", err)
	}

	// Build minikube start command
	args := []string{
		"start",
		"--driver", minikubeConfig.Driver,
		"--cpus", fmt.Sprintf("%d", minikubeConfig.CPUs),
		"--memory", minikubeConfig.Memory,
	}

	// Add addons
	for _, addon := range minikubeConfig.Addons {
		args = append(args, "--addons", addon)
	}

	cmd := exec.Command("minikube", args...)

	// Run command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to create minikube cluster: %w\nOutput: %s", err, string(output))
	}

	// Get kubectl context
	context := "minikube"

	// Create cluster info
	info := &ClusterInfo{
		Name:             "minikube",
		Provider:         "minikube",
		Context:          context,
		ContainerRuntime: string(runtime),
		CreatedAt:        time.Now(),
		Status:           "active",
	}

	// Save cluster info
	if err := SaveClusterInfo(info); err != nil {
		return nil, fmt.Errorf("failed to save cluster info: %w", err)
	}

	return info, nil
}

// DeleteCluster deletes the cluster based on provider
func DeleteCluster(info *ClusterInfo) error {
	switch info.Provider {
	case "kind":
		return deleteKindCluster(info.Name)
	case "minikube":
		return deleteMinikubeCluster()
	default:
		return fmt.Errorf("unsupported provider: %s", info.Provider)
	}
}

// clusterExists checks if a Kind cluster exists
func clusterExists(name string) bool {
	cmd := exec.Command("kind", "get", "clusters")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	clusters := string(output)
	return contains(clusters, name)
}

// deleteKindCluster deletes a Kind cluster
func deleteKindCluster(name string) error {
	cmd := exec.Command("kind", "delete", "cluster", "--name", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete kind cluster: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// deleteMinikubeCluster deletes a Minikube cluster
func deleteMinikubeCluster() error {
	cmd := exec.Command("minikube", "delete")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete minikube cluster: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// generateDefaultKindConfig generates a default Kind configuration file
func generateDefaultKindConfig(kindConfig *KindConfig) (string, error) {
	// Ensure state directory exists
	if err := os.MkdirAll(DefaultStateDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create state directory: %w", err)
	}

	configPath := filepath.Join(DefaultStateDir, "kind-config.yaml")

	// Generate Kind config YAML
	config := fmt.Sprintf(`kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
`)

	// Add control plane node
	config += `  - role: control-plane
`

	// Add worker nodes
	for i := 0; i < kindConfig.Nodes-1; i++ {
		config += `  - role: worker
`
	}

	// Write config file
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return "", fmt.Errorf("failed to write kind config: %w", err)
	}

	return configPath, nil
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsInMiddle(s, substr)))
}

func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// GetClusterStatus gets the current status of the cluster
func GetClusterStatus(info *ClusterInfo) (string, error) {
	switch info.Provider {
	case "kind":
		return getKindClusterStatus(info.Name)
	case "minikube":
		return getMinikubeClusterStatus()
	default:
		return "", fmt.Errorf("unsupported provider: %s", info.Provider)
	}
}

// getKindClusterStatus gets the status of a Kind cluster
func getKindClusterStatus(name string) (string, error) {
	if !clusterExists(name) {
		return "stopped", nil
	}

	// Check if cluster is accessible via kubectl
	cmd := exec.Command("kubectl", "cluster-info", "--context", fmt.Sprintf("kind-%s", name))
	err := cmd.Run()
	if err != nil {
		return "error", nil
	}

	return "active", nil
}

// getMinikubeClusterStatus gets the status of a Minikube cluster
func getMinikubeClusterStatus() (string, error) {
	cmd := exec.Command("minikube", "status", "--format", "{{.Host}}")
	output, err := cmd.Output()
	if err != nil {
		return "stopped", nil
	}

	status := string(output)
	if contains(status, "Running") {
		return "active", nil
	}

	return "stopped", nil
}
