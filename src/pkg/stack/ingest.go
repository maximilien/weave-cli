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
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
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
	RestartEvery     int                  // Restart Milvus every N files (0 = no restart)
	RestartDelay     int                  // Seconds to wait after restart
	ClusterInfo      *ClusterInfo         // Needed for restart
	PortForwardCtx   **PortForwardContext // Pointer to pointer so we can update it
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

	// Start port forwarding in background
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start port forwarding: %w", err)
	}

	// Wait for port forwarding to be ready
	time.Sleep(2 * time.Second)

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
	vdbClient, err := utils.CreateVectorDBClient(milvusConfig)
	if err != nil {
		return fmt.Errorf("failed to create VDB client: %w", err)
	}

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

	// Ensure collection exists (auto-create if needed)
	collectionExists, err := vdbClient.CollectionExists(ctx, cfg.CollectionName)
	if err != nil {
		return fmt.Errorf("failed to check collection existence: %w", err)
	}

	if !collectionExists {
		utils.PrintInfo(fmt.Sprintf("Collection '%s' does not exist, creating...", cfg.CollectionName))
		schema := &vectordb.CollectionSchema{
			Class:      cfg.CollectionName,
			Vectorizer: cfg.EmbeddingModel,
		}
		if err := vdbClient.CreateCollection(ctx, cfg.CollectionName, schema); err != nil {
			return fmt.Errorf("failed to create collection: %w", err)
		}
		utils.PrintInfo(fmt.Sprintf("✅ Created collection: %s", cfg.CollectionName))
	}

	// Create processor
	processor := pipeline.NewProcessor(vdbClient, llmClient, options, progress)

	// Process files with restart support
	var report *pipeline.IngestReport
	if cfg.RestartEvery > 0 {
		// Process files in batches with restarts
		report, err = processFilesWithRestarts(ctx, processor, files, cfg, vdbClient, llmClient, options, progress)
		if err != nil {
			return fmt.Errorf("processing with restarts failed: %w", err)
		}
	} else {
		// Normal processing without restarts
		report, err = processor.ProcessFiles(ctx, files)
		if err != nil {
			return fmt.Errorf("processing failed: %w", err)
		}
	}

	// Show summary
	fmt.Println()
	utils.PrintInfo(fmt.Sprintf("Files processed: %d/%d", report.FilesProcessed, report.FilesScanned))
	utils.PrintInfo(fmt.Sprintf("Documents created: %d", report.DocumentsCreated))
	utils.PrintInfo(fmt.Sprintf("Duration: %.2fs", report.Duration))

	if report.FilesFailed > 0 {
		fmt.Println()
		utils.PrintWarning(fmt.Sprintf("Failed files: %d", report.FilesFailed))
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

// processFilesWithRestarts processes files in batches, restarting Milvus periodically
func processFilesWithRestarts(ctx context.Context, processor *pipeline.Processor, files []pipeline.FileInfo, cfg IngestConfig, vdbClient vectordb.VectorDBClient, llmClient *llm.OpenAIClient, options *pipeline.IngestOptions, progress *pipeline.ProgressTracker) (*pipeline.IngestReport, error) {
	totalFiles := len(files)
	totalProcessed := 0
	totalCreated := 0
	totalFailed := 0
	startTime := time.Now()

	utils.PrintInfo(fmt.Sprintf("Processing %d files in batches of %d", totalFiles, cfg.RestartEvery))
	fmt.Println()

	// Process files in batches
	for batchStart := 0; batchStart < totalFiles; batchStart += cfg.RestartEvery {
		batchEnd := batchStart + cfg.RestartEvery
		if batchEnd > totalFiles {
			batchEnd = totalFiles
		}

		batch := files[batchStart:batchEnd]
		batchNum := (batchStart / cfg.RestartEvery) + 1

		utils.PrintInfo(fmt.Sprintf("📦 Batch %d: Processing files %d-%d of %d", batchNum, batchStart+1, batchEnd, totalFiles))

		// Process this batch
		report, err := processor.ProcessFiles(ctx, batch)
		if err != nil {
			return nil, fmt.Errorf("batch %d processing failed: %w", batchNum, err)
		}

		// Accumulate results
		totalProcessed += report.FilesProcessed
		totalCreated += report.DocumentsCreated
		totalFailed += report.FilesFailed

		utils.PrintSuccess(fmt.Sprintf("✅ Batch %d complete: %d files, %d documents", batchNum, report.FilesProcessed, report.DocumentsCreated))

		// Restart Milvus if not the last batch
		if batchEnd < totalFiles {
			fmt.Println()
			utils.PrintWarning(fmt.Sprintf("🔄 Restarting Milvus after %d files (prevents OOM)...", batchEnd))

			// Stop current port-forward
			if cfg.PortForwardCtx != nil && *cfg.PortForwardCtx != nil {
				(*cfg.PortForwardCtx).Stop()
			}

			// Restart Milvus pod
			if err := restartMilvus(cfg.ClusterInfo); err != nil {
				return nil, fmt.Errorf("failed to restart Milvus: %w", err)
			}

			// Wait for restart
			utils.PrintInfo(fmt.Sprintf("⏳ Waiting %d seconds for Milvus to restart...", cfg.RestartDelay))
			time.Sleep(time.Duration(cfg.RestartDelay) * time.Second)

			// Restart port-forward
			newPortForward, err := StartMilvusPortForward(cfg.ClusterInfo)
			if err != nil {
				return nil, fmt.Errorf("failed to restart port forwarding: %w", err)
			}

			// Update the port-forward context
			if cfg.PortForwardCtx != nil {
				*cfg.PortForwardCtx = newPortForward
			}

			utils.PrintSuccess("✅ Milvus restarted and reconnected")
			fmt.Println()

			// Recreate VDB client with new port
			milvusConfig := &config.VectorDBConfig{
				Type:             "milvus-local",
				Name:             "stack-milvus",
				Address:          fmt.Sprintf("localhost:%d", newPortForward.LocalPort),
				VectorDimensions: getEmbeddingDimensions(cfg.EmbeddingModel),
				Timeout:          30,
			}

			newVDBClient, err := utils.CreateVectorDBClient(milvusConfig)
			if err != nil {
				return nil, fmt.Errorf("failed to recreate VDB client: %w", err)
			}

			// Recreate processor with new client
			processor = pipeline.NewProcessor(newVDBClient, llmClient, options, progress)
		}
	}

	// Return aggregate report
	return &pipeline.IngestReport{
		Status:           "success",
		StartTime:        startTime,
		EndTime:          time.Now(),
		Duration:         time.Since(startTime).Seconds(),
		FilesScanned:     totalFiles,
		FilesProcessed:   totalProcessed,
		DocumentsCreated: totalCreated,
		FilesFailed:      totalFailed,
	}, nil
}

// restartMilvus restarts the Milvus pod in the stack
func restartMilvus(clusterInfo *ClusterInfo) error {
	// Load stack config to get Helm release name
	stackConfig, err := LoadStackConfig("")
	if err != nil {
		return fmt.Errorf("failed to load stack config: %w", err)
	}

	// Deployment pattern: {helmReleaseName}-weave-stack-milvus
	deploymentName := fmt.Sprintf("%s-weave-stack-milvus", stackConfig.Name)

	// Restart using kubectl rollout restart
	cmd := exec.Command("kubectl", "rollout", "restart",
		fmt.Sprintf("deployment/%s", deploymentName),
		"--context", clusterInfo.Context)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to restart deployment: %w (output: %s)", err, string(output))
	}

	return nil
}
