// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package stack

import (
	"fmt"
	"os"

	"github.com/maximilien/weave-cli/src/cmd/utils"
	stackpkg "github.com/maximilien/weave-cli/src/pkg/stack"
	"github.com/spf13/cobra"
)

var (
	initTemplate string
	initRuntime  string
	initForce    bool
)

// InitCmd represents the stack init command
var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new weave stack",
	Long: `Initialize a new weave stack by generating weave-stack.yaml.

This command creates a new stack configuration file with sensible defaults
for your chosen runtime environment (Kind, Minikube, EKS, GKE).

Templates:
  quickstart  - Minimal stack (1 collection, Milvus, local K8s)
  production  - Full stack (collections, evals, dashboard)
  multimodal  - Text + image collections
  oss         - Ollama + sentence-transformers (no API keys)

Examples:
  # Initialize with quickstart template for Kind
  weave stack init --template quickstart --runtime kind

  # Initialize for AWS EKS production
  weave stack init --template production --runtime eks

  # Initialize with defaults (quickstart + kind)
  weave stack init`,
	RunE: runInit,
}

func init() {
	InitCmd.Flags().StringVar(&initTemplate, "template", "quickstart", "Stack template to use")
	InitCmd.Flags().StringVar(&initRuntime, "runtime", "kind", "Kubernetes runtime (kind, minikube, eks, gke)")
	InitCmd.Flags().BoolVar(&initForce, "force", false, "Overwrite existing weave-stack.yaml")
}

func runInit(cmd *cobra.Command, args []string) error {
	// Check if stack already exists
	if stackpkg.StackExists() && !initForce {
		return fmt.Errorf("weave-stack.yaml already exists (use --force to overwrite)")
	}

	utils.PrintInfo("Initializing weave stack...")
	utils.PrintInfo(fmt.Sprintf("Template: %s", initTemplate))
	utils.PrintInfo(fmt.Sprintf("Runtime: %s", initRuntime))
	fmt.Println()

	// Generate stack config from template
	config, err := generateStackFromTemplate(initTemplate, initRuntime)
	if err != nil {
		return fmt.Errorf("failed to generate stack config: %w", err)
	}

	// Save to weave-stack.yaml
	if err := stackpkg.SaveStackConfig(config, ""); err != nil {
		return fmt.Errorf("failed to save stack config: %w", err)
	}

	utils.PrintSuccess("✅ Generated weave-stack.yaml")

	// Create kubernetes/ directory for Helm chart
	if err := os.MkdirAll("kubernetes/templates", 0755); err != nil {
		return fmt.Errorf("failed to create kubernetes directory: %w", err)
	}

	utils.PrintSuccess("✅ Created kubernetes/ directory")

	// Create .gitignore
	if err := createGitignore(); err != nil {
		return fmt.Errorf("failed to create .gitignore: %w", err)
	}

	utils.PrintSuccess("✅ Created .gitignore")

	fmt.Println()
	utils.PrintSuccess("🎉 Stack initialized successfully!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Review weave-stack.yaml and adjust as needed")
	fmt.Println("  2. Add your data files to data/")
	fmt.Println("  3. Deploy: weave stack up --runtime", initRuntime)

	return nil
}

// generateStackFromTemplate creates a StackConfig from a template
func generateStackFromTemplate(template, runtime string) (*stackpkg.StackConfig, error) {
	switch template {
	case "quickstart":
		return generateQuickstartTemplate(runtime), nil
	case "production":
		return generateProductionTemplate(runtime), nil
	case "multimodal":
		return generateMultimodalTemplate(runtime), nil
	case "oss":
		return generateOSSTemplate(runtime), nil
	default:
		return nil, fmt.Errorf("unknown template: %s", template)
	}
}

// generateQuickstartTemplate creates a minimal quickstart stack
func generateQuickstartTemplate(runtime string) *stackpkg.StackConfig {
	config := &stackpkg.StackConfig{
		Version:     "1.0",
		Name:        "my-rag-stack",
		Description: "Quickstart RAG stack",
		Runtime: stackpkg.RuntimeConfig{
			Kubernetes: stackpkg.KubernetesConfig{
				Provider: runtime,
			},
			ContainerRuntime: "podman", // OSS-first
		},
		Infrastructure: stackpkg.InfrastructureConfig{
			VectorDB: stackpkg.VectorDBConfig{
				Type:    "milvus",
				Version: "2.3.0",
				Resources: stackpkg.ResourceRequirements{
					Requests: stackpkg.ResourceLimits{
						Memory: "2Gi", // Reduced for local K8s (Kind/Minikube typically have 8GB)
						CPU:    "1",
					},
					Limits: stackpkg.ResourceLimits{
						Memory: "4Gi", // Reasonable limit for development
						CPU:    "2",
					},
				},
			},
			LLM: &stackpkg.LLMConfig{
				Provider: "openai",
				Models: stackpkg.LLMModels{
					Embedding: "text-embedding-3-small",
					Chat:      "gpt-4o",
				},
			},
		},
		Collections: []stackpkg.CollectionConfig{
			{
				Name:        "Documents",
				Type:        "text",
				Description: "Text documents",
				Schema: stackpkg.SchemaConfig{
					VectorDimensions: 1536, // OpenAI text-embedding-3-small
				},
				Sources: []stackpkg.DataSource{
					{
						Pattern: "data/**/*.pdf",
						Type:    "pdf",
					},
				},
				Chunking: &stackpkg.ChunkingConfig{
					Strategy:  "semantic",
					ChunkSize: 500,
				},
				Embedding: &stackpkg.EmbeddingConfig{
					Model:    "text-embedding-3-small",
					Provider: "openai",
				},
			},
		},
	}

	// Set runtime-specific config
	setRuntimeDefaults(config, runtime)

	return config
}

// generateProductionTemplate creates a full production stack
func generateProductionTemplate(runtime string) *stackpkg.StackConfig {
	config := generateQuickstartTemplate(runtime)
	config.Name = "production-rag-stack"
	config.Description = "Production RAG stack with evaluations and dashboard"

	// Add ingestion config
	config.Ingestion = &stackpkg.IngestionConfig{
		ParallelWorkers: 1,
		BatchSize:       10,
		Retry: &stackpkg.RetryConfig{
			MaxAttempts: 3,
			Backoff:     "exponential",
		},
		Checkpoint: &stackpkg.CheckpointConfig{
			Enabled: true,
			Path:    ".weave-state/ingestion.json",
		},
		Monitoring: &stackpkg.MonitoringConfig{
			VDBHealthCheck: true,
			RestartOnOOM:   true,
		},
	}

	// Add dashboard
	config.Dashboard = &stackpkg.DashboardConfig{
		Enabled: true,
		Type:    "web",
		Web: &stackpkg.WebDashboard{
			Framework: "nextjs",
			Language:  "typescript",
			Version:   "14.x",
			Features:  []string{"search_interface", "health_monitoring"},
			API: []stackpkg.APIRoute{
				{
					Path:       "/api/search",
					Collection: "Documents",
					Methods:    []string{"GET", "POST"},
				},
			},
		},
	}

	return config
}

// generateMultimodalTemplate creates a stack with text + image collections
func generateMultimodalTemplate(runtime string) *stackpkg.StackConfig {
	config := generateQuickstartTemplate(runtime)
	config.Name = "multimodal-rag-stack"
	config.Description = "Multimodal RAG stack (text + images)"

	// Add image collection
	config.Collections = append(config.Collections, stackpkg.CollectionConfig{
		Name:        "Images",
		Type:        "image",
		Description: "Product images with OCR",
		Schema: stackpkg.SchemaConfig{
			VectorDimensions: 1536,
		},
		Sources: []stackpkg.DataSource{
			{
				Pattern: "data/**/*.pdf",
				Type:    "pdf",
				Mode:    "image-extraction",
			},
		},
		Embedding: &stackpkg.EmbeddingConfig{
			Model:    "text-embedding-3-small",
			Provider: "openai",
		},
	})

	// Add MinIO for image storage
	config.Infrastructure.ImageStorage = &stackpkg.ImageStorageConfig{
		Type:     "minio",
		Bucket:   "weave-images",
		Endpoint: "minio:9000",
	}

	return config
}

// generateOSSTemplate creates a stack with OSS components (no API keys)
func generateOSSTemplate(runtime string) *stackpkg.StackConfig {
	config := generateQuickstartTemplate(runtime)
	config.Name = "oss-rag-stack"
	config.Description = "OSS RAG stack (Ollama + sentence-transformers)"

	// Use Ollama for LLM
	config.Infrastructure.LLM = &stackpkg.LLMConfig{
		Provider: "ollama",
		Models: stackpkg.LLMModels{
			Embedding: "nomic-embed-text",
			Chat:      "llama3.1:8b",
		},
	}

	// Use sentence-transformers for embeddings
	config.Collections[0].Embedding = &stackpkg.EmbeddingConfig{
		Model:    "sentence-transformers/all-mpnet-base-v2",
		Provider: "sentence-transformers",
	}
	config.Collections[0].Schema.VectorDimensions = 768 // sentence-transformers dimension

	return config
}

// setRuntimeDefaults sets runtime-specific configuration
func setRuntimeDefaults(config *stackpkg.StackConfig, runtime string) {
	switch runtime {
	case "kind":
		config.Runtime.Kubernetes.Kind = &stackpkg.KindConfig{
			Name:  "weave-stack",
			Nodes: 1,
		}
	case "minikube":
		config.Runtime.Kubernetes.Minikube = &stackpkg.MinikubeConfig{
			Driver: "docker", // More reliable than podman for minikube
			CPUs:   2,
			Memory: "4096", // 4GB - works on most systems (podman machines often default to 8GB)
			Addons: []string{"ingress", "metrics-server"},
		}
		// Note: For podman driver, ensure your podman machine has sufficient memory:
		//   podman machine set --memory 8192 podman-machine-default
		// For Docker Desktop, ensure version >= 20.10.0
		config.Runtime.ContainerRuntime = "docker"
	case "eks":
		config.Runtime.Kubernetes.EKS = &stackpkg.EKSConfig{
			Region: "us-west-2",
			NodeGroups: []stackpkg.EKSNodeGroup{
				{
					Name:         "weave-workers",
					InstanceType: "m5.xlarge",
					DesiredSize:  3,
					MinSize:      1,
					MaxSize:      5,
				},
			},
		}
	case "gke":
		config.Runtime.Kubernetes.GKE = &stackpkg.GKEConfig{
			Zone: "us-central1-a",
			NodePools: []stackpkg.GKENodePool{
				{
					Name:        "weave-workers",
					MachineType: "n1-standard-4",
					NodeCount:   3,
					AutoScaling: &stackpkg.AutoScaling{
						MinNodeCount: 1,
						MaxNodeCount: 5,
					},
				},
			},
		}
	}
}

// createGitignore creates a .gitignore file for the stack
func createGitignore() error {
	gitignore := `# Weave Stack
.weave-state/
kubernetes/values.yaml
.env

# IDE
.vscode/
.idea/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Logs
logs/
*.log

# Kubernetes
kubeconfig
*.kubeconfig
`

	return os.WriteFile(".gitignore", []byte(gitignore), 0644)
}
