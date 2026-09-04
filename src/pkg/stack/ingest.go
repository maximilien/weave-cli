// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package stack

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	AutoRestart      bool                 // Automatically restart on failures
	MaxRetries       int                  // Maximum retry attempts (with auto-restart)
	ResumeFrom       string               // Resume from specific file
	CheckpointEvery  int                  // Save checkpoint every N files
	CheckpointFile   string               // Path to checkpoint file
	SkipFailed       bool                 // Skip files that fail after max retries
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

	// Load checkpoint if checkpoint features are enabled
	var checkpoint *IngestCheckpoint
	if cfg.AutoRestart || cfg.ResumeFrom != "" || cfg.CheckpointEvery > 0 {
		checkpoint, err = LoadCheckpoint(cfg.CheckpointFile)
		if err != nil {
			return fmt.Errorf("failed to load checkpoint: %w", err)
		}

		if checkpoint != nil {
			utils.PrintInfo(fmt.Sprintf("📋 Loaded checkpoint from %s", cfg.CheckpointFile))
			utils.PrintInfo(fmt.Sprintf("   Completed: %d files", len(checkpoint.CompletedFiles)))
			if len(checkpoint.FailedFiles) > 0 {
				utils.PrintWarning(fmt.Sprintf("   Failed: %d files", len(checkpoint.FailedFiles)))
			}
			fmt.Println()
		}
	}

	// Filter files based on checkpoint and --resume-from
	filePathsToProcess := make([]string, len(files))
	for i, f := range files {
		filePathsToProcess[i] = f.Path
	}

	filteredPaths, skippedCount, err := FilterFilesFromCheckpoint(filePathsToProcess, checkpoint, cfg.ResumeFrom)
	if err != nil {
		return fmt.Errorf("failed to filter files: %w", err)
	}

	if skippedCount > 0 {
		utils.PrintInfo(fmt.Sprintf("⏭️  Skipped %d already completed files", skippedCount))
		fmt.Println()
	}

	if len(filteredPaths) == 0 {
		utils.PrintSuccess("✅ All files already processed!")
		return nil
	}

	// Rebuild files list with filtered paths
	filteredFiles := []pipeline.FileInfo{}
	for _, path := range filteredPaths {
		for _, f := range files {
			if f.Path == path {
				filteredFiles = append(filteredFiles, f)
				break
			}
		}
	}

	files = filteredFiles
	utils.PrintInfo(fmt.Sprintf("Processing %d files", len(files)))
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

	// Initialize checkpoint if enabled
	if (cfg.AutoRestart || cfg.ResumeFrom != "" || cfg.CheckpointEvery > 0) && checkpoint == nil {
		filePathsForCheckpoint := make([]string, len(files))
		for i, f := range files {
			filePathsForCheckpoint[i] = f.Path
		}
		checkpoint = InitCheckpoint(cfg.CollectionName, filePathsForCheckpoint, nil)
	}

	// Create processor
	processor := pipeline.NewProcessor(vdbClient, llmClient, options, progress)

	// Process files with auto-restart or checkpoint support
	var report *pipeline.IngestReport
	if cfg.AutoRestart {
		// Process with auto-restart and retry logic
		report, err = processFilesWithAutoRestart(ctx, processor, files, cfg, vdbClient, llmClient, options, progress, checkpoint)
		if err != nil {
			return fmt.Errorf("processing with auto-restart failed: %w", err)
		}
	} else if cfg.RestartEvery > 0 {
		// Process files in batches with scheduled restarts
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

	// Poll for document persistence with retry backoff (Issue #57)
	// Milvus flush operations timeout during batch inserts but return success,
	// relying on async flush. We poll the collection count to verify persistence.
	fmt.Println()
	utils.PrintInfo("⏳ Waiting for documents to persist (polling with retry)...")

	expectedCount := int64(report.DocumentsCreated)
	maxRetries := 6 // Total wait: ~63 seconds (1+2+4+8+16+32)
	var actualCount int64
	var verifyErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Exponential backoff: 1s, 2s, 4s, 8s, 16s, 32s
		waitDuration := time.Duration(1<<uint(attempt)) * time.Second
		if attempt > 0 {
			utils.PrintInfo(fmt.Sprintf("   Retry %d/%d (waiting %ds)...", attempt, maxRetries-1, int(waitDuration.Seconds())))
		}
		time.Sleep(waitDuration)

		actualCount, verifyErr = vdbClient.GetCollectionCount(ctx, cfg.CollectionName)
		if verifyErr != nil {
			// Can't verify - keep retrying
			continue
		}

		// Check if documents have appeared
		if actualCount >= expectedCount {
			// Success!
			break
		}

		// Partial success or still waiting
		if actualCount > 0 {
			utils.PrintInfo(fmt.Sprintf("   Found %d/%d documents so far...", actualCount, expectedCount))
		}
	}

	// Final verification
	if verifyErr != nil {
		utils.PrintWarning(fmt.Sprintf("⚠️  Could not verify document persistence: %v", verifyErr))
		utils.PrintWarning("   Documents may have been created but verification failed")
	} else {
		if actualCount == 0 && expectedCount > 0 {
			fmt.Println()
			utils.PrintError("❌ VERIFICATION FAILED: No documents found in collection!")
			utils.PrintError(fmt.Sprintf("   Expected: %d documents", expectedCount))
			utils.PrintError("   Actual: 0 documents")
			utils.PrintError(fmt.Sprintf("   Waited: ~%d seconds with %d retries", (1<<uint(maxRetries))-1, maxRetries))
			fmt.Println()
			utils.PrintError("This is Issue #57: Documents created but not persisted to Milvus")
			utils.PrintError("Possible causes:")
			utils.PrintError("  • Milvus async flush failed (check Milvus logs)")
			utils.PrintError("  • Connection dropped before flush completed")
			utils.PrintError("  • Milvus memory/resource constraints (OOM)")
			utils.PrintError("  • Milvus pod restart during operation")
			fmt.Println()
			utils.PrintError("Workaround: Use 'weave docs create' instead of 'weave stack ingest'")
			fmt.Println()
			return fmt.Errorf("document persistence verification failed: 0/%d documents persisted after %d retries", expectedCount, maxRetries)
		} else if actualCount < expectedCount {
			fmt.Println()
			utils.PrintWarning(fmt.Sprintf("⚠️  PARTIAL PERSISTENCE: Only %d/%d documents verified in collection", actualCount, expectedCount))
			utils.PrintWarning(fmt.Sprintf("   Missing: %d documents", expectedCount-actualCount))
			utils.PrintWarning(fmt.Sprintf("   Waited: ~%d seconds with %d retries", (1<<uint(maxRetries))-1, maxRetries))
			utils.PrintWarning("   Some documents may not have been persisted (flush timeout or failure)")
		} else {
			fmt.Println()
			utils.PrintSuccess(fmt.Sprintf("✅ Verified: %d documents persisted to collection", actualCount))
		}
	}

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

// processFilesWithAutoRestart processes files with auto-restart on failures
func processFilesWithAutoRestart(ctx context.Context, processor *pipeline.Processor, files []pipeline.FileInfo, cfg IngestConfig, vdbClient vectordb.VectorDBClient, llmClient *llm.OpenAIClient, options *pipeline.IngestOptions, progress *pipeline.ProgressTracker, checkpoint *IngestCheckpoint) (*pipeline.IngestReport, error) {
	totalFiles := len(files)
	totalProcessed := 0
	totalCreated := 0
	totalFailed := 0
	startTime := time.Now()

	utils.PrintInfo(fmt.Sprintf("🔄 Auto-restart enabled: Max %d retries per file", cfg.MaxRetries))
	if cfg.CheckpointEvery > 0 {
		utils.PrintInfo(fmt.Sprintf("💾 Checkpoint enabled: Saving every %d files to %s", cfg.CheckpointEvery, cfg.CheckpointFile))
	}
	fmt.Println()

	// Process files one by one with retry logic
	for fileIdx, file := range files {
		fileNum := fileIdx + 1
		utils.PrintInfo(fmt.Sprintf("[%d/%d] Processing: %s", fileNum, totalFiles, filepath.Base(file.Path)))

		// Check if file previously failed
		var previousFailure *FailedFile
		if checkpoint != nil {
			previousFailure = checkpoint.GetFailedFile(file.Path)
		}

		retryCount := 0
		if previousFailure != nil {
			retryCount = previousFailure.Attempts
		}

		var lastError error
		success := false

		// Retry loop
		for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
			if attempt > 0 {
				utils.PrintWarning(fmt.Sprintf("   Retry %d/%d", attempt, cfg.MaxRetries))
			}

			// Process single file
			report, err := processor.ProcessFiles(ctx, []pipeline.FileInfo{file})

			if err == nil && report.FilesFailed == 0 {
				// Success!
				success = true
				totalProcessed++
				totalCreated += report.DocumentsCreated
				utils.PrintSuccess(fmt.Sprintf("   ✅ Completed: %d documents created", report.DocumentsCreated))

				// Mark file as completed in checkpoint
				if checkpoint != nil {
					checkpoint.MarkFileCompleted(file.Path)
				}
				break
			}

			// Failure - check if it's a connection error
			lastError = err
			if err != nil {
				utils.PrintError(fmt.Sprintf("   ❌ Error: %v", err))
			}

			// Check if we should restart and retry
			if attempt < cfg.MaxRetries {
				utils.PrintWarning("   🔄 Restarting Milvus and retrying...")

				// Stop current port-forward
				if cfg.PortForwardCtx != nil && *cfg.PortForwardCtx != nil {
					(*cfg.PortForwardCtx).Stop()
				}

				// Restart Milvus pod
				if err := restartMilvus(cfg.ClusterInfo); err != nil {
					return nil, fmt.Errorf("failed to restart Milvus: %w", err)
				}

				// Wait for restart
				utils.PrintInfo(fmt.Sprintf("   ⏳ Waiting %d seconds for Milvus to restart...", cfg.RestartDelay))
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

				utils.PrintSuccess("   ✅ Milvus restarted and reconnected")
			}
		}

		// Handle failure after all retries
		if !success {
			totalFailed++
			retryCount += cfg.MaxRetries + 1

			if cfg.SkipFailed {
				utils.PrintWarning(fmt.Sprintf("   ⚠️  Skipped after %d failed attempts", cfg.MaxRetries+1))

				// Mark file as failed in checkpoint
				if checkpoint != nil {
					errMsg := "unknown error"
					if lastError != nil {
						errMsg = lastError.Error()
					}
					checkpoint.MarkFileFailed(file.Path, retryCount, errMsg)
				}
			} else {
				// Abort entire job
				if checkpoint != nil {
					SaveCheckpoint(checkpoint, cfg.CheckpointFile)
				}
				return nil, fmt.Errorf("file failed after %d attempts: %s (use --skip-failed to continue)", cfg.MaxRetries+1, file.Path)
			}
		}

		// Save checkpoint every N files
		if checkpoint != nil && cfg.CheckpointEvery > 0 && fileNum%cfg.CheckpointEvery == 0 {
			if err := SaveCheckpoint(checkpoint, cfg.CheckpointFile); err != nil {
				utils.PrintWarning(fmt.Sprintf("Failed to save checkpoint: %v", err))
			} else {
				utils.PrintInfo(fmt.Sprintf("💾 Checkpoint saved (%d files completed)", len(checkpoint.CompletedFiles)))
			}
		}

		fmt.Println()
	}

	// Save final checkpoint
	if checkpoint != nil {
		if err := SaveCheckpoint(checkpoint, cfg.CheckpointFile); err != nil {
			utils.PrintWarning(fmt.Sprintf("Failed to save final checkpoint: %v", err))
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
