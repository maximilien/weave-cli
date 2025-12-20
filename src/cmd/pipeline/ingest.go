package pipeline

import (
	"context"
	"fmt"
	"os"

	"github.com/maximilien/weave-cli/src/cmd/utils"
	"github.com/maximilien/weave-cli/src/pkg/llm"
	"github.com/maximilien/weave-cli/src/pkg/output"
	"github.com/maximilien/weave-cli/src/pkg/pipeline"
	"github.com/spf13/cobra"
)

var (
	// Ingest flags
	collection   string
	glob         string
	exclude      []string
	recursive    bool
	batchSize    int
	workers      int
	metadata     map[string]string
	dryRun       bool
	resume       bool
	stateFile    string
	outputFormat string
	reportFile   string
	quiet        bool
)

var ingestCmd = &cobra.Command{
	Use:   "ingest SOURCE",
	Short: "Ingest documents from a directory into a vector database",
	Long: `Ingest documents from a directory into a vector database.

This command scans a directory for files matching the specified criteria,
processes them in batches, and creates documents in the target vector database.

Features:
- Supports multiple file types (PDF, TXT, MD, JSON, YAML)
- Concurrent processing with configurable workers
- Batch document creation for efficiency
- Resume capability for interrupted operations
- Dry-run mode to preview operations
- Custom metadata for all documents
- Multiple output formats (JSON, YAML, Table)

Examples:
  # Basic ingestion
  weave pipeline ingest ./docs --collection documents --vdb qdrant-cloud

  # With glob pattern
  weave pipeline ingest ./docs --glob "**/*.pdf" --collection pdfs

  # With exclusions
  weave pipeline ingest ./docs --exclude "*.tmp" --exclude "drafts/**" --collection docs

  # Custom batch size and workers
  weave pipeline ingest ./docs --collection docs --batch-size 50 --workers 8

  # Add custom metadata
  weave pipeline ingest ./docs --collection docs --metadata project=myapp --metadata version=1.0

  # Dry run
  weave pipeline ingest ./docs --collection docs --dry-run --output json

  # Resume interrupted ingestion
  weave pipeline ingest ./docs --collection docs --resume --state-file .weave-state.json
`,
	Args: cobra.ExactArgs(1),
	RunE: runIngest,
}

func init() {
	ingestCmd.Flags().StringVar(&collection, "collection", "", "Target collection name (required)")
	ingestCmd.MarkFlagRequired("collection")

	ingestCmd.Flags().StringVar(&glob, "glob", "", "Glob pattern for file matching (e.g., '**/*.pdf')")
	ingestCmd.Flags().StringSliceVar(&exclude, "exclude", []string{}, "Exclusion patterns")
	ingestCmd.Flags().BoolVar(&recursive, "recursive", true, "Recursively scan directories")

	ingestCmd.Flags().IntVar(&batchSize, "batch-size", 100, "Documents per batch")
	ingestCmd.Flags().IntVar(&workers, "workers", 4, "Number of concurrent workers")
	ingestCmd.Flags().StringToStringVar(&metadata, "metadata", nil, "Additional metadata for all documents")

	ingestCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without executing")
	ingestCmd.Flags().BoolVar(&resume, "resume", false, "Resume from previous state")
	ingestCmd.Flags().StringVar(&stateFile, "state-file", ".weave-state.json", "State file path for resume")

	ingestCmd.Flags().StringVar(&outputFormat, "output", "table", "Output format (json, yaml, table)")
	ingestCmd.Flags().StringVar(&reportFile, "report-file", "", "Report output file")
	ingestCmd.Flags().BoolVar(&quiet, "quiet", false, "Suppress progress output")
}

func runIngest(cmd *cobra.Command, args []string) error {
	source := args[0]

	// Validate source path
	if _, err := os.Stat(source); err != nil {
		return fmt.Errorf("source path does not exist: %s", source)
	}

	ctx := context.Background()

	// Load config
	cfg, err := utils.LoadConfigWithOverrides()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get VDB selection
	selection, err := utils.GetSelectedVectorDBs(cmd, cfg)
	if err != nil {
		return fmt.Errorf("failed to get vector database selection: %w", err)
	}

	// Validate single database required for write operation
	if err := utils.ValidateDatabaseSelection(selection, utils.OperationTypeWrite, "pipeline ingest"); err != nil {
		return err
	}

	// Get the database config
	dbConfig := utils.HandleSingleDatabaseSelection(ctx, selection, cfg, collection,
		fmt.Sprintf("weave pipeline ingest %s --collection %s", source, collection))

	// Create VDB client
	vdbClient, err := utils.CreateVectorDBClient(dbConfig)
	if err != nil {
		return fmt.Errorf("failed to create VDB client: %w", err)
	}

	vdbType := string(dbConfig.Type)

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
		Source:       source,
		Glob:         glob,
		Exclude:      exclude,
		Recursive:    recursive,
		Collection:   collection,
		VDBType:      vdbType,
		BatchSize:    batchSize,
		Workers:      workers,
		Metadata:     metadata,
		DryRun:       dryRun,
		Resume:       resume,
		StateFile:    stateFile,
		OutputFormat: outputFormat,
		ReportFile:   reportFile,
		Quiet:        quiet,
	}

	// Phase 1: Scan files
	if !quiet {
		fmt.Printf("🔍 Scanning files in %s...\n", source)
	}

	scanner := pipeline.NewFileScanner(source, glob, exclude, recursive)
	files, err := scanner.Scan(ctx)
	if err != nil {
		return fmt.Errorf("failed to scan files: %w", err)
	}

	if !quiet {
		fmt.Printf("✅ Found %d files\n", len(files))
	}

	if len(files) == 0 {
		fmt.Println("No files found matching criteria")
		return nil
	}

	// Phase 2: Process files
	if !quiet {
		fmt.Printf("⚙️  Processing files (batch_size=%d, workers=%d)...\n", batchSize, workers)
	}

	processor := pipeline.NewProcessor(vdbClient, llmClient, options)
	report, err := processor.ProcessFiles(ctx, files)
	if err != nil {
		return fmt.Errorf("failed to process files: %w", err)
	}

	// Phase 3: Output report
	var writer = os.Stdout
	if reportFile != "" {
		f, err := os.Create(reportFile)
		if err != nil {
			return fmt.Errorf("failed to create report file: %w", err)
		}
		defer f.Close()
		writer = f
	}

	// Determine output format
	format := output.Format(outputFormat)
	if err := output.FormatIngestReport(report, format, writer); err != nil {
		return fmt.Errorf("failed to format report: %w", err)
	}

	// Print summary if not quiet and not writing to file
	if !quiet && reportFile == "" {
		fmt.Println()
		fmt.Printf("📊 Ingestion %s\n", report.Status)
		fmt.Printf("   Files processed: %d/%d\n", report.FilesProcessed, report.FilesScanned)
		fmt.Printf("   Documents created: %d\n", report.DocumentsCreated)
		fmt.Printf("   Duration: %.2fs\n", report.Duration)
		if report.ThroughputFiles > 0 {
			fmt.Printf("   Throughput: %.2f files/sec, %.2f docs/sec\n", report.ThroughputFiles, report.ThroughputDocs)
		}
		if report.FilesFailed > 0 {
			fmt.Printf("   ⚠️  Failed: %d files\n", report.FilesFailed)
		}
	}

	return nil
}
