package pipeline

import (
	"github.com/spf13/cobra"
)

// PipelineCmd represents the pipeline command
var PipelineCmd = &cobra.Command{
	Use:     "pipeline",
	Aliases: []string{"pipe", "p"},
	Short:   "Pipeline commands for batch document ingestion",
	Long: `Pipeline commands for batch document ingestion.

The pipeline command provides tools for ingesting large numbers of documents
from directories into vector databases. It supports:

- Recursive directory scanning
- Glob pattern matching
- Batch processing with configurable workers
- Multiple file types (PDF, TXT, MD, JSON, YAML)
- Resume capability for interrupted operations
- Dry-run mode for previewing operations
- JSON/YAML/Table output formats

Examples:
  # Ingest all PDFs from a directory
  weave pipeline ingest ./docs --glob "**/*.pdf" --collection documents

  # Ingest with custom batch size and workers
  weave pipeline ingest ./docs --collection docs --batch-size 50 --workers 8

  # Dry run to preview what will be ingested
  weave pipeline ingest ./docs --collection docs --dry-run

  # Resume interrupted ingestion
  weave pipeline ingest ./docs --collection docs --resume --state-file .weave-state.json
`,
}

func init() {
	// Add subcommands
	PipelineCmd.AddCommand(ingestCmd)
}
