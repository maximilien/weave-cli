// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package schema

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
	interactive    bool
	apply          bool
	globPattern    string
)

var suggestCmd = &cobra.Command{
	Use:   "suggest SOURCE",
	Short: "Suggest collection schema by analyzing sample documents",
	Long: `Analyze sample documents and suggest an optimal collection schema using AI.

The AI agent examines document structure, field types, content patterns, and
metadata to recommend a schema configuration tailored to your data.

Examples:
  # Analyze samples and output suggested schema
  weave schema suggest ./samples --collection docs --output schema.yaml

  # Interactive mode with AI explanations
  weave schema suggest ./samples --collection products --interactive

  # Include custom requirements
  weave schema suggest ./samples --collection articles \
    --requirements "Support multi-language search, enable date filtering" \
    --output articles-schema.yaml

  # Analyze specific file types
  weave schema suggest ./samples --glob "**/*.{pdf,md}" \
    --collection documentation --output docs-schema.yaml`,
	Args: cobra.ExactArgs(1),
	RunE: runSchemaSuggest,
}

func init() {
	suggestCmd.Flags().StringVar(&collectionName, "collection", "", "Target collection name (required)")
	suggestCmd.MarkFlagRequired("collection")

	suggestCmd.Flags().StringVar(&outputFile, "output", "", "Output file for schema (YAML format)")
	suggestCmd.Flags().StringVar(&requirements, "requirements", "", "Custom requirements for schema")
	suggestCmd.Flags().IntVar(&maxSamples, "max-samples", 50, "Maximum number of files to analyze")
	suggestCmd.Flags().StringVar(&vdbType, "vdb", "weaviate", "Target vector database type")
	suggestCmd.Flags().BoolVar(&interactive, "interactive", false, "Interactive mode with refinement")
	suggestCmd.Flags().BoolVar(&apply, "apply", false, "Apply schema to VDB immediately")
	suggestCmd.Flags().StringVar(&globPattern, "glob", "", "Glob pattern for file filtering")
}

func runSchemaSuggest(cmd *cobra.Command, args []string) error {
	source := args[0]
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Validate source path
	if _, err := os.Stat(source); err != nil {
		return fmt.Errorf("source path does not exist: %s", source)
	}

	// Create LLM client
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("OPENAI_API_KEY required for AI schema suggestion")
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

	// Create schema agent
	agent := agents.NewSchemaAgent(llmClient)

	// Prepare input
	input := &agents.SchemaAnalysisInput{
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
		return fmt.Errorf("schema analysis failed: %w", err)
	}

	output, ok := result.(*agents.SchemaAnalysisOutput)
	if !ok {
		return fmt.Errorf("invalid output from schema agent")
	}

	// Display results
	displaySchemaAnalysis(output)

	// Interactive mode
	if interactive {
		fmt.Println("\n💬 Interactive mode enabled")
		fmt.Println("   Would you like to refine the schema? (y/n)")
		// TODO: Implement interactive refinement
	}

	// Save to file
	if outputFile != "" {
		if err := saveSchemaToYAML(output.Schema, outputFile); err != nil {
			return fmt.Errorf("failed to save schema: %w", err)
		}
		fmt.Printf("\n✅ Schema saved to: %s\n", outputFile)
	}

	// Apply to VDB
	if apply {
		fmt.Printf("\n🚀 Applying schema to %s...\n", vdbType)
		// TODO: Implement schema application
		fmt.Println("⚠️  Schema application not yet implemented")
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

// displaySchemaAnalysis displays the analysis results
func displaySchemaAnalysis(output *agents.SchemaAnalysisOutput) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("📋 Schema Analysis for '%s'\n", output.Schema.CollectionName)
	fmt.Println(strings.Repeat("=", 60))

	fmt.Printf("\n🎯 Confidence: %.1f%%\n", output.Confidence*100)

	fmt.Printf("\n📝 Reasoning:\n%s\n", output.Reasoning)

	fmt.Printf("\n🔧 Suggested Schema:\n")
	fmt.Printf("   Collection: %s\n", output.Schema.CollectionName)
	fmt.Printf("   Vector Dimensions: %d\n", output.Schema.VectorDimensions)
	fmt.Printf("   Similarity Metric: %s\n", output.Schema.SimilarityMetric)

	if len(output.Schema.Fields) > 0 {
		fmt.Printf("\n📊 Fields (%d):\n", len(output.Schema.Fields))
		for _, field := range output.Schema.Fields {
			indexed := ""
			if field.Indexed {
				indexed = " [indexed]"
			}
			filterable := ""
			if field.Filterable {
				filterable = " [filterable]"
			}
			required := ""
			if field.Required {
				required = " [required]"
			}
			fmt.Printf("   • %-20s %s%s%s%s\n",
				field.Name,
				field.Type,
				indexed,
				filterable,
				required,
			)
			if field.Description != "" {
				fmt.Printf("     %s\n", field.Description)
			}
		}
	}

	if len(output.Schema.Indexes) > 0 {
		fmt.Printf("\n🔍 Indexes (%d):\n", len(output.Schema.Indexes))
		for _, index := range output.Schema.Indexes {
			fmt.Printf("   • %s (%s): %v\n", index.Name, index.Type, index.Fields)
		}
	}

	if len(output.Warnings) > 0 {
		fmt.Printf("\n⚠️  Warnings:\n")
		for _, warning := range output.Warnings {
			fmt.Printf("   • %s\n", warning)
		}
	}

	if len(output.FieldAnalysis) > 0 {
		fmt.Printf("\n📈 Field Analysis:\n")
		for _, fa := range output.FieldAnalysis {
			fmt.Printf("   • %s (frequency: %.1f%%, cardinality: %d)\n",
				fa.Name,
				fa.Frequency*100,
				fa.Cardinality,
			)
			fmt.Printf("     %s\n", fa.Reasoning)
		}
	}

	if len(output.Alternatives) > 0 {
		fmt.Printf("\n🔄 Alternative Schemas Available: %d\n", len(output.Alternatives))
	}

	// Display chunking recommendations
	if output.ChunkingAdvice != nil {
		fmt.Printf("\n📏 Chunking Recommendations:\n")
		fmt.Printf("   Recommended Size: %d characters (~%d tokens)\n",
			output.ChunkingAdvice.RecommendedSize,
			output.ChunkingAdvice.RecommendedSize/4)
		fmt.Printf("   Size Range: %d - %d characters\n",
			output.ChunkingAdvice.MinSize,
			output.ChunkingAdvice.MaxSize)
		fmt.Printf("   Overlap Size: %d characters (~%.0f%%)\n",
			output.ChunkingAdvice.OverlapSize,
			float64(output.ChunkingAdvice.OverlapSize)/float64(output.ChunkingAdvice.RecommendedSize)*100)
		fmt.Printf("   Document Type: %s\n", output.ChunkingAdvice.DocumentType)
		fmt.Printf("\n   💡 Reasoning:\n   %s\n", output.ChunkingAdvice.Reasoning)

		if len(output.ChunkingAdvice.Considerations) > 0 {
			fmt.Printf("\n   📝 Considerations:\n")
			for _, consideration := range output.ChunkingAdvice.Considerations {
				fmt.Printf("      • %s\n", consideration)
			}
		}
	}

	fmt.Println(strings.Repeat("=", 60))
}

// saveSchemaToYAML saves schema configuration to YAML file
func saveSchemaToYAML(schema agents.SchemaConfig, filename string) error {
	data, err := yaml.Marshal(schema)
	if err != nil {
		return fmt.Errorf("failed to marshal schema to YAML: %w", err)
	}

	// Add comment header
	header := fmt.Sprintf(`# Vector Database Schema Configuration
# Generated by Weave CLI at %s
# Collection: %s

`, time.Now().Format(time.RFC3339), schema.CollectionName)

	content := header + string(data)

	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
