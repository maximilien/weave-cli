// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package stack

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/maximilien/weave-cli/src/cmd/utils"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/llm"
	"github.com/maximilien/weave-cli/src/pkg/pipeline"
)

// IngestConfig holds configuration for stack ingestion
type IngestConfig struct {
	CollectionName   string
	DataPath         string
	Type             string
	EmbeddingModel   string
	ChunkSize        int
	ParallelWorkers  int
	BatchSize        int
	MilvusLocalPort  int
	ProgressCallback func(string)
}

// PortForwardContext holds port forwarding context
type PortForwardContext struct {
	LocalPort int
	Cmd       *exec.Cmd
	Cancel    context.CancelFunc
}

// Stop stops the port forwarding
func (p *PortForwardContext) Stop() {
	if p.Cancel != nil {
		p.Cancel()
	}
	if p.Cmd != nil && p.Cmd.Process != nil {
		_ = p.Cmd.Process.Kill()
		_ = p.Cmd.Wait()
	}
}

// StartMilvusPortForward starts port forwarding to Milvus
func StartMilvusPortForward(clusterInfo *ClusterInfo) (*PortForwardContext, error) {
	localPort := 19530

	// Load stack config to get Helm release name
	stackConfig, err := LoadStackConfig("")
	if err != nil {
		return nil, fmt.Errorf("failed to load stack config: %w", err)
	}

	// Service pattern: {helmReleaseName}-weave-stack-milvus
	serviceName := fmt.Sprintf("%s-weave-stack-milvus", stackConfig.Name)

	// Create context with cancel
	ctx, cancel := context.WithCancel(context.Background())

	// Build kubectl port-forward command
	cmd := exec.CommandContext(ctx, "kubectl", "port-forward",
		fmt.Sprintf("svc/%s", serviceName),
		fmt.Sprintf("%d:19530", localPort),
		"--context", clusterInfo.Context)

	// Capture stderr to see if port-forward fails
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Printf("DEBUG: Starting port-forward: kubectl port-forward svc/%s %d:19530 --context %s\n", serviceName, localPort, clusterInfo.Context)

	// Start port forwarding in background
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start port forwarding: %w", err)
	}

	// Wait for port forwarding to be ready
	fmt.Printf("DEBUG: Waiting for port-forward to establish...\n")
	time.Sleep(2 * time.Second)
	fmt.Printf("DEBUG: Port-forward should be ready on localhost:%d\n", localPort)

	return &PortForwardContext{
		LocalPort: localPort,
		Cmd:       cmd,
		Cancel:    cancel,
	}, nil
}

// IngestToStack runs ingestion pipeline targeting stack Milvus
func IngestToStack(cfg IngestConfig) error {
	ctx := context.Background()

	// Create Milvus config for local port-forwarded connection
	milvusConfig := &config.VectorDBConfig{
		Type:             "milvus-local",
		Name:             "stack-milvus",
		Address:          fmt.Sprintf("localhost:%d", cfg.MilvusLocalPort),
		VectorDimensions: getEmbeddingDimensions(cfg.EmbeddingModel),
		Timeout:          30,
	}

	// Create VDB client
	fmt.Printf("DEBUG: Creating VDB client with config:\n")
	fmt.Printf("  Type: %s\n", milvusConfig.Type)
	fmt.Printf("  Name: %s\n", milvusConfig.Name)
	fmt.Printf("  Address: %s\n", milvusConfig.Address)
	fmt.Printf("  VectorDimensions: %d\n", milvusConfig.VectorDimensions)

	vdbClient, err := utils.CreateVectorDBClient(milvusConfig)
	if err != nil {
		return fmt.Errorf("failed to create VDB client: %w", err)
	}
	fmt.Printf("DEBUG: VDB client created successfully (type: %T)\n", vdbClient)

	// Create LLM client for embeddings
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}

	llmClient, err := llm.NewOpenAIClient(apiKey)
	if err != nil {
		return fmt.Errorf("failed to create LLM client: %w", err)
	}

	// Create ingestion options
	options := &pipeline.IngestOptions{
		Source:     cfg.DataPath,
		Collection: cfg.CollectionName,
		VDBType:    string(milvusConfig.Type),
		BatchSize:  cfg.BatchSize,
		Workers:    cfg.ParallelWorkers,
		Recursive:  true,
		Quiet:      false,
	}

	// Create progress tracker
	progress := pipeline.NewProgressTracker(true, false)

	// Scan files
	progress.StartScanning()

	scanner := pipeline.NewFileScanner(cfg.DataPath, "", []string{}, true)
	files, err := scanner.Scan(ctx)
	if err != nil {
		return fmt.Errorf("failed to scan files: %w", err)
	}

	progress.FinishScanning(len(files))

	if len(files) == 0 {
		return fmt.Errorf("no files found in: %s", cfg.DataPath)
	}

	utils.PrintInfo(fmt.Sprintf("Found %d files to process", len(files)))
	fmt.Println()

	// Create processor
	fmt.Printf("DEBUG: Creating processor with collection: %s\n", options.Collection)
	processor := pipeline.NewProcessor(vdbClient, llmClient, options, progress)

	// Process files
	fmt.Printf("DEBUG: Processing %d files...\n", len(files))
	report, err := processor.ProcessFiles(ctx, files)
	if err != nil {
		return fmt.Errorf("processing failed: %w", err)
	}
	fmt.Printf("DEBUG: Processing complete. Documents created: %d\n", report.DocumentsCreated)

	// Show summary
	fmt.Println()
	utils.PrintInfo(fmt.Sprintf("Files processed: %d/%d", report.FilesProcessed, report.FilesScanned))
	utils.PrintInfo(fmt.Sprintf("Documents created: %d", report.DocumentsCreated))
	utils.PrintInfo(fmt.Sprintf("Duration: %.2fs", report.Duration))

	if report.FilesFailed > 0 {
		fmt.Println()
		utils.PrintWarning(fmt.Sprintf("Failed files: %d", report.FilesFailed))
		for _, e := range report.Errors {
			utils.PrintError(fmt.Sprintf("  %s: %s", e.File, e.Error))
		}
	}

	// Give Milvus time to flush and persist data before stopping port-forward
	time.Sleep(3 * time.Second)

	return nil
}

// getEmbeddingDimensions returns dimensions for common embedding models
func getEmbeddingDimensions(model string) int {
	switch model {
	case "text-embedding-3-small":
		return 1536
	case "text-embedding-3-large":
		return 3072
	case "text-embedding-ada-002":
		return 1536
	case "sentence-transformers/all-mpnet-base-v2":
		return 768
	case "sentence-transformers/all-MiniLM-L6-v2":
		return 384
	case "nomic-embed-text":
		return 768
	case "mxbai-embed-large":
		return 1024
	default:
		return 1536 // Default to OpenAI small
	}
}
