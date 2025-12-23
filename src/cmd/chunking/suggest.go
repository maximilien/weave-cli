// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package chunking

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/agents"
	"github.com/maximilien/weave-cli/src/pkg/llm"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	collectionName string
	outputFile     string
	requirements   string
	maxSamples     int
	vdbType        string
	globPattern    string
	outputFormat   string
)

var suggestCmd = &cobra.Command{
	Use:   "suggest SOURCE",
	Short: "Suggest optimal chunking strategy by analyzing sample documents",
	Long: `Analyze sample documents and suggest an optimal chunking strategy using AI.

The AI agent examines document structure, content density, paragraph distribution,
and other characteristics to recommend ideal chunk sizes and overlap strategies.

Examples:
  # Analyze samples and display recommendations
  weave chunking suggest ./samples --collection docs

  # Save recommendations to YAML file
  weave chunking suggest ./docs --collection articles --output chunking.yaml

  # Include custom requirements
  weave chunking suggest ./samples --collection tech-docs \
    --requirements "Code-heavy technical documentation" \
    --output chunking-config.yaml

  # Analyze specific file types
  weave chunking suggest ./samples --glob "**/*.{pdf,md}" \
    --collection documentation`,
	Args: cobra.ExactArgs(1),
	RunE: runChunkingSuggest,
}

func init() {
	suggestCmd.Flags().StringVar(&collectionName, "collection", "", "Target collection name (required)")
	suggestCmd.MarkFlagRequired("collection")

	suggestCmd.Flags().StringVar(&outputFile, "output", "", "Output file for recommendations (YAML format)")
	suggestCmd.Flags().StringVar(&requirements, "requirements", "", "Custom requirements for chunking strategy")
	suggestCmd.Flags().IntVar(&maxSamples, "max-samples", 50, "Maximum number of files to analyze")
	suggestCmd.Flags().StringVar(&vdbType, "vdb", "weaviate", "Target vector database type")
	suggestCmd.Flags().StringVar(&globPattern, "glob", "", "Glob pattern for file filtering")
	suggestCmd.Flags().StringVarP(&outputFormat, "format", "f", "text", "Output format: text, json, yaml")
}

func runChunkingSuggest(cmd *cobra.Command, args []string) error {
	source := args[0]
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Validate source path
	if _, err := os.Stat(source); err != nil {
		return fmt.Errorf("source path does not exist: %s", source)
	}

	// Create LLM client
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("OPENAI_API_KEY required for AI chunking analysis")
	}

	llmClient, err := llm.NewOpenAIClient(apiKey)
	if err != nil {
		return fmt.Errorf("failed to create LLM client: %w", err)
	}

	// Scan for sample files
	fmt.Printf("🔍 Scanning for sample files in: %s\n", source)
	files, err := scanFiles(source, globPattern)
	if err != nil {
		return fmt.Errorf("failed to scan files: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no files found in source directory")
	}

	fmt.Printf("📊 Found %d sample files\n", len(files))

	// Create chunking agent
	agent := agents.NewChunkingAgent(llmClient)

	// Prepare input
	input := &agents.ChunkingAnalysisInput{
		SampleFiles:    files,
		CollectionName: collectionName,
		Requirements:   requirements,
		VDBType:        vdbType,
		MaxSamples:     maxSamples,
	}

	// Execute analysis
	fmt.Printf("🤖 Analyzing documents with AI...\n")
	result, err := agent.Execute(ctx, input)
	if err != nil {
		return fmt.Errorf("chunking analysis failed: %w", err)
	}

	output, ok := result.(*agents.ChunkingAnalysisOutput)
	if !ok {
		return fmt.Errorf("invalid output from chunking agent")
	}

	// Display results
	displayChunkingAnalysis(output)

	// Save to file if requested
	if outputFile != "" {
		if err := saveChunkingToFile(output, outputFile); err != nil {
			return fmt.Errorf("failed to save chunking recommendations: %w", err)
		}
		fmt.Printf("\n✅ Chunking recommendations saved to: %s\n", outputFile)
	}

	return nil
}

// scanFiles scans directory for files matching pattern
func scanFiles(root string, pattern string) ([]string, error) {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Apply glob pattern if specified
		if pattern != "" {
			matched, err := filepath.Match(pattern, filepath.Base(path))
			if err != nil {
				return err
			}
			if !matched {
				return nil
			}
		}

		// Include common document types
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".pdf" || ext == ".txt" || ext == ".md" || ext == ".json" || ext == ".yaml" || ext == ".yml" {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// displayChunkingAnalysis displays the analysis results
func displayChunkingAnalysis(output *agents.ChunkingAnalysisOutput) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("📏 Chunking Strategy Recommendations\n")
	fmt.Println(strings.Repeat("=", 60))

	fmt.Printf("\n🎯 Confidence: %.1f%%\n", output.Confidence*100)

	rec := output.Recommendation

	fmt.Printf("\n📊 Recommended Configuration:\n")
	fmt.Printf("   Chunk Size: %d characters (~%d tokens)\n",
		rec.RecommendedSize, rec.RecommendedSize/4)
	fmt.Printf("   Size Range: %d - %d characters\n",
		rec.MinSize, rec.MaxSize)
	fmt.Printf("   Overlap: %d characters (~%.0f%%)\n",
		rec.OverlapSize,
		float64(rec.OverlapSize)/float64(rec.RecommendedSize)*100)
	fmt.Printf("   Document Type: %s\n", rec.DocumentType)

	fmt.Printf("\n💡 Reasoning:\n   %s\n", rec.Reasoning)

	if len(rec.Considerations) > 0 {
		fmt.Printf("\n📝 Key Considerations:\n")
		for _, consideration := range rec.Considerations {
			fmt.Printf("   • %s\n", consideration)
		}
	}

	if len(output.Metrics.FileTypes) > 0 {
		fmt.Printf("\n📈 Document Metrics:\n")
		fmt.Printf("   File Types: %v\n", output.Metrics.FileTypes)
		fmt.Printf("   Avg Content Length: %d bytes\n", output.Metrics.AvgContentLength)
		fmt.Printf("   Avg Paragraphs: %d\n", output.Metrics.AvgParagraphs)
		fmt.Printf("   Avg Paragraph Length: %d chars\n", output.Metrics.AvgParagraphLen)
		fmt.Printf("   Content Density: %s\n", output.Metrics.ContentDensity)
	}

	if len(output.Warnings) > 0 {
		fmt.Printf("\n⚠️  Warnings:\n")
		for _, warning := range output.Warnings {
			fmt.Printf("   • %s\n", warning)
		}
	}

	fmt.Println(strings.Repeat("=", 60))
}

// saveChunkingToFile saves chunking recommendations to a file
func saveChunkingToFile(output *agents.ChunkingAnalysisOutput, filename string) error {
	data, err := yaml.Marshal(output)
	if err != nil {
		return fmt.Errorf("failed to marshal to YAML: %w", err)
	}

	// Add comment header
	header := fmt.Sprintf(`# Chunking Strategy Recommendations
# Generated by Weave CLI at %s
# Confidence: %.1f%%

`, time.Now().Format(time.RFC3339), output.Confidence*100)

	content := header + string(data)

	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
