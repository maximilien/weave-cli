// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package document

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/maximilien/weave-cli/src/cmd/utils"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/spf13/cobra"
)

// ProcessedFileStatus represents the status of a processed file
type ProcessedFileStatus struct {
	FilePath         string    `json:"file_path"`
	ProcessedAt      time.Time `json:"processed_at"`
	Success          bool      `json:"success"`
	Error            string    `json:"error,omitempty"`
	TextChunks       int       `json:"text_chunks"`
	Images           int       `json:"images"`
	ProcessingTimeMs int64     `json:"processing_time_ms"`
	FileSize         int64     `json:"file_size"`
	RetryCount       int       `json:"retry_count"`
	ChunksFailed     int       `json:"chunks_failed,omitempty"`
	ImagesFailed     int       `json:"images_failed,omitempty"`
}

// BatchProgress tracks overall batch processing progress
type BatchProgress struct {
	TotalFiles         int
	ProcessedFiles     int
	SuccessFiles       int
	FailedFiles        int
	SkippedFiles       int
	TotalChunks        int
	TotalImages        int
	StartTime          time.Time
	EstimatedRemaining time.Duration
	CurrentFile        string
}

// BatchCmd represents the batch command
var BatchCmd = &cobra.Command{
	Use:     "batch",
	Aliases: []string{"b"},
	Short:   "Batch create documents from a directory",
	Long: `Batch create documents from all files in a directory.

Supported file types:
- Text files (.txt, .md, .json, etc.)
- Image files (.jpg, .jpeg, .png, .gif, etc.)
- PDF files (.pdf) with text and image extraction

Features:
- Parallel processing with --parallel flag
- Automatic retry on failure with --retry flag
- Skip already processed files (tracks in .processed files)
- Visual progress indicators with emojis and colors
- Time estimation and progress tracking
- Optional CSV report generation with --create-report

Examples:
  weave docs batch --directory ./docs --collection MyCollection
  weave docs batch --dir ./docs --collection MyCollection --parallel 3
  weave docs batch --dir ./docs --collection MyCollection --create-report
  weave docs batch --dir ./docs --collection MyCollection --retry 3 --parallel 5`,
	Run: runBatchCreate,
}

func init() {
	// BatchCmd is registered in document.go, not here
	// This allows it to be a direct subcommand of 'docs'

	// Required flags
	BatchCmd.Flags().StringP("directory", "d", "", "Directory containing documents to process (required)")
	BatchCmd.Flags().StringP("collection", "c", "", "Collection name for documents (required)")
	if err := BatchCmd.MarkFlagRequired("directory"); err != nil {
		panic(fmt.Sprintf("failed to mark 'directory' flag as required: %v", err))
	}
	if err := BatchCmd.MarkFlagRequired("collection"); err != nil {
		panic(fmt.Sprintf("failed to mark 'collection' flag as required: %v", err))
	}

	// Processing options
	BatchCmd.Flags().IntP("parallel", "p", 1, "Number of parallel workers (default: 1)")
	BatchCmd.Flags().Int("retry", 2, "Number of retry attempts for failed documents (default: 2)")
	BatchCmd.Flags().IntP("chunk-size", "s", 5000, "Chunk size for text content (default: 5000 characters)")

	// PDF-specific options
	BatchCmd.Flags().String("image-collection", "", "Collection name for extracted PDF images (default: same as main collection)")
	BatchCmd.Flags().Bool("skip-small-images", true, "Skip small images when extracting from PDFs (default: true)")
	BatchCmd.Flags().Int("min-image-size", 5120, "Minimum image size in bytes (default: 5120 = 5KB)")
	BatchCmd.Flags().Int("batch-size", 10, "Number of images to process before pausing for memory cleanup (default: 10)")

	// Reporting options
	BatchCmd.Flags().String("create-report", "", "Create a new CSV report of processing results (default: batch-report.csv)")
	BatchCmd.Flags().Bool("force-reprocess", false, "Force reprocess all files, ignore .processed files (default: false)")
	BatchCmd.Flags().Bool("json", false, "Output in JSON format")
}

func runBatchCreate(cmd *cobra.Command, args []string) {
	// Get flags
	directory, _ := cmd.Flags().GetString("directory")
	collectionName, _ := cmd.Flags().GetString("collection")
	parallel, _ := cmd.Flags().GetInt("parallel")
	retryCount, _ := cmd.Flags().GetInt("retry")
	chunkSize, _ := cmd.Flags().GetInt("chunk-size")
	imageCollection, _ := cmd.Flags().GetString("image-collection")
	skipSmallImages, _ := cmd.Flags().GetBool("skip-small-images")
	minImageSize, _ := cmd.Flags().GetInt("min-image-size")
	batchSize, _ := cmd.Flags().GetInt("batch-size")
	createReport, _ := cmd.Flags().GetString("create-report")
	forceReprocess, _ := cmd.Flags().GetBool("force-reprocess")

	// Validate directory
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		utils.PrintError(fmt.Sprintf("Directory not found: %s", directory))
		os.Exit(1)
	}

	// Load configuration
	cfg, err := utils.LoadConfigWithInteractiveHelp()
	if err != nil {
		// Error already formatted and displayed by LoadConfigWithInteractiveHelp
		os.Exit(1)
	}

	// Get selected databases
	selection, err := utils.GetSelectedVectorDBs(cmd, cfg)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to get database selection: %v", err))
		os.Exit(1)
	}

	ctx := context.Background()

	// Smart database selection for single-database operations
	dbConfig := utils.HandleSingleDatabaseSelection(ctx, selection, cfg, collectionName, fmt.Sprintf("weave docs batch --directory %s --collection %s", directory, collectionName))

	// Default report path
	if createReport == "" && cmd.Flags().Lookup("create-report").Changed {
		createReport = "batch-report.csv"
	}

	// Print batch configuration
	printBatchConfiguration(directory, collectionName, parallel, retryCount, chunkSize, imageCollection, skipSmallImages, minImageSize, batchSize, createReport)

	// Scan directory for files
	files, err := scanDirectory(directory)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to scan directory: %v", err))
		os.Exit(1)
	}

	if len(files) == 0 {
		utils.PrintWarning("No supported files found in directory")
		return
	}

	utils.PrintInfo(fmt.Sprintf("Found %d file(s) to process", len(files)))
	fmt.Println()

	// Filter out already processed files unless force-reprocess is enabled
	var filesToProcess []string
	if forceReprocess {
		filesToProcess = files
		utils.PrintInfo("Force reprocess enabled - processing all files")
	} else {
		filesToProcess = filterProcessedFiles(files)
		skippedCount := len(files) - len(filesToProcess)
		if skippedCount > 0 {
			utils.PrintInfo(fmt.Sprintf("Skipping %d already processed file(s)", skippedCount))
		}
	}

	if len(filesToProcess) == 0 {
		utils.PrintSuccess("All files already processed! Use --force-reprocess to reprocess.")
		return
	}

	fmt.Println()
	utils.PrintInfo(fmt.Sprintf("Processing %d file(s)...", len(filesToProcess)))
	fmt.Println()

	// Initialize batch progress
	progress := &BatchProgress{
		TotalFiles:     len(filesToProcess),
		ProcessedFiles: 0,
		SuccessFiles:   0,
		FailedFiles:    0,
		SkippedFiles:   0,
		TotalChunks:    0,
		TotalImages:    0,
		StartTime:      time.Now(),
	}

	// Process files
	processedStatuses := processBatchFiles(ctx, dbConfig, filesToProcess, collectionName, parallel, retryCount, chunkSize, imageCollection, skipSmallImages, minImageSize, batchSize, progress)

	// Print final summary
	fmt.Println()
	printBatchSummary(progress, processedStatuses)

	// Write report if requested
	if createReport != "" {
		if err := writeBatchReport(createReport, processedStatuses); err != nil {
			utils.PrintWarning(fmt.Sprintf("Failed to write batch report: %v", err))
		} else {
			utils.PrintSuccess(fmt.Sprintf("Batch report written to: %s", createReport))
		}
	}
}

// scanDirectory scans directory for supported files
func scanDirectory(directory string) ([]string, error) {
	var files []string
	supportedExts := map[string]bool{
		".txt": true, ".md": true, ".json": true, ".yaml": true, ".yml": true,
		".pdf": true,
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true, ".webp": true,
	}

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if supportedExts[ext] {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

// filterProcessedFiles filters out files that have already been processed
func filterProcessedFiles(files []string) []string {
	var unprocessed []string
	for _, file := range files {
		processedFile := file + ".processed"
		if _, err := os.Stat(processedFile); os.IsNotExist(err) {
			unprocessed = append(unprocessed, file)
		} else {
			// Check if processed file is valid
			status, err := loadProcessedStatus(processedFile)
			if err != nil || !status.Success {
				// If can't load or failed, reprocess
				unprocessed = append(unprocessed, file)
			}
		}
	}
	return unprocessed
}

// processBatchFiles processes files in parallel
func processBatchFiles(ctx context.Context, dbConfig *config.VectorDBConfig, files []string, collectionName string, parallel int, retryCount int, chunkSize int, imageCollection string, skipSmallImages bool, minImageSize int, batchSize int, progress *BatchProgress) []ProcessedFileStatus {
	var wg sync.WaitGroup
	var mu sync.Mutex
	results := make([]ProcessedFileStatus, 0, len(files))

	// Create worker pool
	fileChan := make(chan string, len(files))
	for _, file := range files {
		fileChan <- file
	}
	close(fileChan)

	// Start workers
	for i := 0; i < parallel; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for file := range fileChan {
				status := processFileWithRetry(ctx, dbConfig, file, collectionName, retryCount, chunkSize, imageCollection, skipSmallImages, minImageSize, batchSize, progress, &mu)
				mu.Lock()
				results = append(results, status)
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()
	return results
}

// processFileWithRetry processes a single file with retry logic
func processFileWithRetry(ctx context.Context, dbConfig *config.VectorDBConfig, filePath string, collectionName string, maxRetries int, chunkSize int, imageCollection string, skipSmallImages bool, minImageSize int, batchSize int, progress *BatchProgress, mu *sync.Mutex) ProcessedFileStatus {
	startTime := time.Now()
	fileInfo, _ := os.Stat(filePath)
	var fileSize int64
	if fileInfo != nil {
		fileSize = fileInfo.Size()
	}

	status := ProcessedFileStatus{
		FilePath:   filePath,
		FileSize:   fileSize,
		RetryCount: 0,
	}

	// Update current file in progress (thread-safe)
	mu.Lock()
	progress.CurrentFile = filepath.Base(filePath)
	mu.Unlock()

	// Try processing with retries
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		status.RetryCount = attempt

		// Create temporary report for this file
		tempReport := &utils.ProcessingReport{
			FilePath:   filePath,
			Collection: collectionName,
			Timestamp:  time.Now(),
			TextChunks: []utils.ChunkReport{},
			Images:     []utils.ImageReport{},
			ReportPath: "",
			ReportMode: "",
		}

		err := utils.CreateDocument(ctx, dbConfig, collectionName, filePath, chunkSize, imageCollection, skipSmallImages, minImageSize, batchSize, "", "", "")
		if err == nil {
			// Success
			status.Success = true
			status.ProcessedAt = time.Now()
			status.ProcessingTimeMs = time.Since(startTime).Milliseconds()
			status.TextChunks = len(tempReport.TextChunks)
			status.Images = len(tempReport.Images)

			// Count failures
			for _, chunk := range tempReport.TextChunks {
				if !chunk.Success {
					status.ChunksFailed++
				}
			}
			for _, img := range tempReport.Images {
				if !img.Success {
					status.ImagesFailed++
				}
			}

			// Update progress
			updateProgress(progress, &status, mu)

			// Print success message
			printFileSuccess(filePath, &status, progress)

			// Save processed status
			if err := saveProcessedStatus(filePath+".processed", &status); err != nil {
				// Log error but don't fail the processing
				fmt.Fprintf(os.Stderr, "Warning: failed to save processed status for %s: %v\n", filePath, err)
			}
			return status
		}

		lastErr = err
		if attempt < maxRetries {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}

	// All attempts failed
	status.Success = false
	status.Error = lastErr.Error()
	status.ProcessedAt = time.Now()
	status.ProcessingTimeMs = time.Since(startTime).Milliseconds()

	// Update progress
	updateProgress(progress, &status, mu)

	// Print failure message
	printFileFailure(filePath, &status, progress)

	// Save failed status
	if err := saveProcessedStatus(filePath+".processed", &status); err != nil {
		// Log error but don't fail the processing
		fmt.Fprintf(os.Stderr, "Warning: failed to save processed status for %s: %v\n", filePath, err)
	}
	return status
}

// updateProgress updates the batch progress (thread-safe)
func updateProgress(progress *BatchProgress, status *ProcessedFileStatus, mu *sync.Mutex) {
	mu.Lock()
	defer mu.Unlock()

	progress.ProcessedFiles++
	if status.Success {
		progress.SuccessFiles++
		progress.TotalChunks += status.TextChunks
		progress.TotalImages += status.Images
	} else {
		progress.FailedFiles++
	}

	// Update time estimation
	if progress.ProcessedFiles > 0 {
		elapsed := time.Since(progress.StartTime)
		avgTimePerFile := elapsed / time.Duration(progress.ProcessedFiles)
		remainingFiles := progress.TotalFiles - progress.ProcessedFiles
		progress.EstimatedRemaining = avgTimePerFile * time.Duration(remainingFiles)
	}
}

// printFileSuccess prints a success message for a processed file
func printFileSuccess(filePath string, status *ProcessedFileStatus, progress *BatchProgress) {
	fileName := filepath.Base(filePath)
	fileType := getFileType(filePath)

	green := color.New(color.FgGreen).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	var emoji string
	switch fileType {
	case "pdf":
		emoji = "📄"
	case "image":
		emoji = "🖼️"
	default:
		emoji = "📝"
	}

	msg := fmt.Sprintf("%s %s [%d/%d] %s", emoji, green("✓"), progress.ProcessedFiles, progress.TotalFiles, fileName)

	// Add details
	details := []string{}
	if status.TextChunks > 0 {
		details = append(details, fmt.Sprintf("%d chunks", status.TextChunks))
	}
	if status.Images > 0 {
		details = append(details, fmt.Sprintf("%d images", status.Images))
	}
	if len(details) > 0 {
		msg += fmt.Sprintf(" (%s)", strings.Join(details, ", "))
	}

	// Add timing
	msg += cyan(fmt.Sprintf(" - %.2fs", float64(status.ProcessingTimeMs)/1000))

	// Add ETA if available
	if progress.EstimatedRemaining > 0 {
		msg += cyan(fmt.Sprintf(" - ETA: %s", formatDuration(progress.EstimatedRemaining)))
	}

	fmt.Println(msg)
}

// printFileFailure prints a failure message for a processed file
func printFileFailure(filePath string, status *ProcessedFileStatus, progress *BatchProgress) {
	fileName := filepath.Base(filePath)
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	msg := fmt.Sprintf("❌ %s [%d/%d] %s", red("✗"), progress.ProcessedFiles, progress.TotalFiles, fileName)
	msg += yellow(fmt.Sprintf(" - %s", status.Error))

	if status.RetryCount > 0 {
		msg += yellow(fmt.Sprintf(" (after %d retries)", status.RetryCount))
	}

	fmt.Println(msg)
}

// getFileType determines file type from extension
func getFileType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".pdf":
		return "pdf"
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp":
		return "image"
	default:
		return "text"
	}
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
}

// printBatchConfiguration prints the batch processing configuration
func printBatchConfiguration(directory, collection string, parallel, retry, chunkSize int, imageCollection string, skipSmallImages bool, minImageSize, batchSize int, reportPath string) {
	blue := color.New(color.FgBlue, color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	fmt.Println(blue("Batch Processing Configuration:"))
	fmt.Printf("  Directory:        %s\n", cyan(directory))
	fmt.Printf("  Collection:       %s\n", cyan(collection))
	fmt.Printf("  Parallel workers: %s\n", cyan(fmt.Sprintf("%d", parallel)))
	fmt.Printf("  Retry attempts:   %s\n", cyan(fmt.Sprintf("%d", retry)))
	fmt.Printf("  Chunk size:       %s\n", cyan(fmt.Sprintf("%d chars", chunkSize)))

	if imageCollection != "" {
		fmt.Printf("  Image collection: %s\n", cyan(imageCollection))
	}

	fmt.Printf("  PDF options:\n")
	fmt.Printf("    Skip small images: %s\n", cyan(fmt.Sprintf("%v", skipSmallImages)))
	fmt.Printf("    Min image size:    %s\n", cyan(fmt.Sprintf("%d bytes", minImageSize)))
	fmt.Printf("    Batch size:        %s\n", cyan(fmt.Sprintf("%d", batchSize)))

	if reportPath != "" {
		fmt.Printf("  Report path:      %s\n", cyan(reportPath))
	}

	fmt.Println()
}

// printBatchSummary prints the final batch processing summary
func printBatchSummary(progress *BatchProgress, statuses []ProcessedFileStatus) {
	blue := color.New(color.FgBlue, color.Bold).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	elapsed := time.Since(progress.StartTime)

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println(blue("Batch Processing Summary"))
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Total files:      %d\n", progress.TotalFiles)
	fmt.Printf("Successful:       %s\n", green(fmt.Sprintf("%d", progress.SuccessFiles)))
	fmt.Printf("Failed:           %s\n", red(fmt.Sprintf("%d", progress.FailedFiles)))
	fmt.Printf("Text chunks:      %d\n", progress.TotalChunks)
	fmt.Printf("Images extracted: %d\n", progress.TotalImages)
	fmt.Printf("Total time:       %s\n", formatDuration(elapsed))

	if progress.ProcessedFiles > 0 {
		avgTime := elapsed / time.Duration(progress.ProcessedFiles)
		fmt.Printf("Avg time/file:    %s\n", formatDuration(avgTime))
	}

	// List failed files if any
	if progress.FailedFiles > 0 {
		fmt.Println()
		fmt.Println(yellow("Failed files:"))
		for _, status := range statuses {
			if !status.Success {
				fmt.Printf("  - %s: %s\n", filepath.Base(status.FilePath), status.Error)
			}
		}
	}

	fmt.Println(strings.Repeat("=", 60))
}

// saveProcessedStatus saves the processed status to a .processed file
func saveProcessedStatus(filePath string, status *ProcessedFileStatus) error {
	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

// loadProcessedStatus loads the processed status from a .processed file
func loadProcessedStatus(filePath string) (*ProcessedFileStatus, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var status ProcessedFileStatus
	err = json.Unmarshal(data, &status)
	return &status, err
}

// writeBatchReport writes a CSV report of all processed files
func writeBatchReport(reportPath string, statuses []ProcessedFileStatus) error {
	// Use existing CSV writing infrastructure from utils
	file, err := os.Create(reportPath)
	if err != nil {
		return fmt.Errorf("failed to create report file: %w", err)
	}
	defer file.Close()

	// Write header
	header := "Timestamp,File,Status,TextChunks,Images,ProcessingTimeMs,FileSize,RetryCount,ChunksFailed,ImagesFailed,Error\n"
	if _, err := file.WriteString(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write each status
	for _, status := range statuses {
		statusStr := "success"
		if !status.Success {
			statusStr = "failed"
		}

		line := fmt.Sprintf("%s,%s,%s,%d,%d,%d,%d,%d,%d,%d,%s\n",
			status.ProcessedAt.Format("2006-01-02 15:04:05"),
			filepath.Base(status.FilePath),
			statusStr,
			status.TextChunks,
			status.Images,
			status.ProcessingTimeMs,
			status.FileSize,
			status.RetryCount,
			status.ChunksFailed,
			status.ImagesFailed,
			status.Error,
		)

		if _, err := file.WriteString(line); err != nil {
			return fmt.Errorf("failed to write status line: %w", err)
		}
	}

	return nil
}
