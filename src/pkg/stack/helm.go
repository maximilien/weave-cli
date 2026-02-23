// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package stack

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// HelmValues represents the Helm values.yaml structure
type HelmValues struct {
	Stack        HelmStackMeta     `yaml:"stack"`
	VectorDB     HelmVectorDB      `yaml:"vectordb"`
	ImageStorage *HelmImageStorage `yaml:"imageStorage,omitempty"`
}

// HelmStackMeta represents stack metadata
type HelmStackMeta struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

// HelmVectorDB represents VectorDB configuration for Helm
type HelmVectorDB struct {
	Type      string          `yaml:"type"`
	Version   string          `yaml:"version"`
	Replicas  int             `yaml:"replicas"`
	Resources HelmResources   `yaml:"resources"`
	Storage   HelmStorage     `yaml:"storage"`
	Service   HelmService     `yaml:"service"`
	Health    HelmHealthCheck `yaml:"health"`
}

// HelmResources represents resource requests/limits
type HelmResources struct {
	Requests HelmResourceLimits `yaml:"requests"`
	Limits   HelmResourceLimits `yaml:"limits"`
}

// HelmResourceLimits represents CPU/memory limits
type HelmResourceLimits struct {
	Memory string `yaml:"memory"`
	CPU    string `yaml:"cpu"`
}

// HelmStorage represents storage configuration
type HelmStorage struct {
	Class string `yaml:"class"`
	Size  string `yaml:"size"`
}

// HelmService represents service configuration
type HelmService struct {
	Type        string `yaml:"type"`
	Port        int    `yaml:"port"`
	MetricsPort int    `yaml:"metricsPort"`
}

// HelmHealthCheck represents health check configuration
type HelmHealthCheck struct {
	Liveness  HelmProbe `yaml:"liveness"`
	Readiness HelmProbe `yaml:"readiness"`
}

// HelmProbe represents a probe configuration
type HelmProbe struct {
	Path                string `yaml:"path"`
	Port                int    `yaml:"port"`
	InitialDelaySeconds int    `yaml:"initialDelaySeconds"`
	PeriodSeconds       int    `yaml:"periodSeconds"`
	FailureThreshold    int    `yaml:"failureThreshold,omitempty"`
}

// HelmImageStorage represents image storage configuration
type HelmImageStorage struct {
	Enabled   bool          `yaml:"enabled"`
	Type      string        `yaml:"type"`
	Replicas  int           `yaml:"replicas"`
	Resources HelmResources `yaml:"resources"`
	Storage   HelmStorage   `yaml:"storage"`
}

// GenerateHelmValues generates Helm values.yaml from StackConfig
func GenerateHelmValues(config *StackConfig) (*HelmValues, error) {
	values := &HelmValues{
		Stack: HelmStackMeta{
			Name:    config.Name,
			Version: config.Version,
		},
		VectorDB: HelmVectorDB{
			Type:     config.Infrastructure.VectorDB.Type,
			Version:  config.Infrastructure.VectorDB.Version,
			Replicas: 1, // Default
			Resources: HelmResources{
				Requests: HelmResourceLimits{
					Memory: config.Infrastructure.VectorDB.Resources.Requests.Memory,
					CPU:    config.Infrastructure.VectorDB.Resources.Requests.CPU,
				},
				Limits: HelmResourceLimits{
					Memory: config.Infrastructure.VectorDB.Resources.Limits.Memory,
					CPU:    config.Infrastructure.VectorDB.Resources.Limits.CPU,
				},
			},
			Storage: HelmStorage{
				Class: "standard", // Default
				Size:  "50Gi",     // Default
			},
			Service: HelmService{
				Type:        "ClusterIP",
				Port:        19530, // Milvus default
				MetricsPort: 9091,  // Milvus metrics default
			},
		},
	}

	// Set replicas from K8s config if available
	if config.Infrastructure.VectorDB.Kubernetes != nil {
		values.VectorDB.Replicas = config.Infrastructure.VectorDB.Kubernetes.Deployment.Replicas
		values.VectorDB.Storage.Class = config.Infrastructure.VectorDB.Kubernetes.Deployment.StorageClass
		values.VectorDB.Storage.Size = config.Infrastructure.VectorDB.Kubernetes.Deployment.PVCSize
	}

	// Set health checks if available
	if config.Infrastructure.VectorDB.Health != nil {
		if config.Infrastructure.VectorDB.Health.Liveness != nil {
			liveness := config.Infrastructure.VectorDB.Health.Liveness
			values.VectorDB.Health.Liveness = HelmProbe{
				Path:                liveness.HTTPGet.Path,
				Port:                liveness.HTTPGet.Port,
				InitialDelaySeconds: liveness.InitialDelaySeconds,
				PeriodSeconds:       liveness.PeriodSeconds,
				FailureThreshold:    liveness.FailureThreshold,
			}
		}

		if config.Infrastructure.VectorDB.Health.Readiness != nil {
			readiness := config.Infrastructure.VectorDB.Health.Readiness
			values.VectorDB.Health.Readiness = HelmProbe{
				Path:                readiness.HTTPGet.Path,
				Port:                readiness.HTTPGet.Port,
				InitialDelaySeconds: readiness.InitialDelaySeconds,
				PeriodSeconds:       readiness.PeriodSeconds,
			}
		}
	}

	// Add image storage if configured
	if config.Infrastructure.ImageStorage != nil {
		values.ImageStorage = &HelmImageStorage{
			Enabled:  true,
			Type:     config.Infrastructure.ImageStorage.Type,
			Replicas: 1,
			Resources: HelmResources{
				Requests: HelmResourceLimits{
					Memory: "1Gi",
					CPU:    "500m",
				},
				Limits: HelmResourceLimits{
					Memory: "2Gi",
					CPU:    "1",
				},
			},
			Storage: HelmStorage{
				Class: "standard",
				Size:  "20Gi",
			},
		}
	}

	return values, nil
}

// SaveHelmValues saves Helm values to a file
func SaveHelmValues(values *HelmValues, outputPath string) error {
	data, err := yaml.Marshal(values)
	if err != nil {
		return fmt.Errorf("failed to marshal Helm values: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write Helm values: %w", err)
	}

	return nil
}

// GenerateHelmChart generates complete Helm chart from StackConfig
func GenerateHelmChart(config *StackConfig, outputDir string) error {
	// Generate values.yaml
	values, err := GenerateHelmValues(config)
	if err != nil {
		return fmt.Errorf("failed to generate Helm values: %w", err)
	}

	// Save values.yaml
	valuesPath := filepath.Join(outputDir, "values.yaml")
	if err := SaveHelmValues(values, valuesPath); err != nil {
		return fmt.Errorf("failed to save Helm values: %w", err)
	}

	// Copy template files from templates/helm/weave-stack/
	if err := CopyHelmTemplates(outputDir); err != nil {
		return fmt.Errorf("failed to copy Helm templates: %w", err)
	}

	return nil
}

// CopyHelmTemplates copies Helm chart templates to output directory
func CopyHelmTemplates(outputDir string) error {
	// Get weave-cli root directory (assume we're in src/pkg/stack/)
	// Navigate to templates/helm/weave-stack/
	templateDir := "templates/helm/weave-stack"

	// Copy Chart.yaml
	if err := copyFile(
		filepath.Join(templateDir, "Chart.yaml"),
		filepath.Join(outputDir, "Chart.yaml"),
	); err != nil {
		return fmt.Errorf("failed to copy Chart.yaml: %w", err)
	}

	// Copy templates directory
	templatesDir := filepath.Join(templateDir, "templates")
	outputTemplatesDir := filepath.Join(outputDir, "templates")

	if err := os.MkdirAll(outputTemplatesDir, 0755); err != nil {
		return fmt.Errorf("failed to create templates directory: %w", err)
	}

	// Copy all template files
	templateFiles := []string{
		"_helpers.tpl",
		"milvus-deployment.yaml",
		"milvus-service.yaml",
		"vectordb-pvc.yaml",
	}

	for _, file := range templateFiles {
		src := filepath.Join(templatesDir, file)
		dst := filepath.Join(outputTemplatesDir, file)
		if err := copyFile(src, dst); err != nil {
			return fmt.Errorf("failed to copy %s: %w", file, err)
		}
	}

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// Ensure destination directory exists
	dir := filepath.Dir(dst)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(dst, data, 0644)
}

// HelmInstall installs a Helm chart
func HelmInstall(chartPath, releaseName, namespace string, timeout string, context string) error {
	args := []string{
		"install",
		releaseName,
		chartPath,
		"--namespace", namespace,
		"--create-namespace",
		"--wait",
		"--timeout", timeout,
	}

	// Add kubeconfig context if provided
	if context != "" {
		args = append(args, "--kube-context", context)
	}

	cmd := exec.Command("helm", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("helm install failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// HelmUninstall uninstalls a Helm release
func HelmUninstall(releaseName, namespace string, context string) error {
	args := []string{
		"uninstall",
		releaseName,
		"--namespace", namespace,
	}

	if context != "" {
		args = append(args, "--kube-context", context)
	}

	cmd := exec.Command("helm", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("helm uninstall failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}
