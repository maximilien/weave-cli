// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package collection

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/maximilien/weave-cli/src/cmd/utils"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/spf13/cobra"
)

// CompareCmd represents the collection compare command
var CompareCmd = &cobra.Command{
	Use:   "compare COLLECTION1 COLLECTION2 [COLLECTION3...]",
	Short: "Compare embedding models across collections",
	Long: `Compare search results and quality metrics across collections with different embedding models.

This command runs test queries against multiple collections and generates a comparison report
showing relevance scores, document rankings, and performance metrics for each model.

Perfect for:
- Evaluating different embedding models (OpenAI vs OSS vs local)
- Comparing search quality after re-embedding
- Making data-driven model selection decisions
- Cost/performance analysis

Examples:
  # Compare two collections
  weave collection compare Client0_PDFs Client0_PDFs_OSS \
    --query "Leica M3 vintage camera" \
    --report comparison.md

  # Compare multiple models
  weave collection compare \
    Client0_OpenAI \
    Client0_OSS_768 \
    Client0_OSS_384 \
    Client0_Ollama \
    --query "auction results 2024" \
    --top-k 10 \
    --report models_comparison.md

  # Multiple queries
  weave collection compare Col1 Col2 \
    --query "query 1" \
    --query "query 2" \
    --query "query 3" \
    --report report.md`,
	Args: cobra.MinimumNArgs(2),
	Run:  runCompare,
}

func init() {
	CompareCmd.Flags().StringSlice("query", []string{}, "Test queries to run (can specify multiple)")
	CompareCmd.Flags().IntP("top-k", "k", 5, "Number of results to retrieve per query")
	CompareCmd.Flags().StringP("report", "r", "", "Output report file (markdown format)")
	CompareCmd.Flags().StringP("format", "f", "markdown", "Report format: markdown, json")
	CompareCmd.MarkFlagRequired("query")
}

func runCompare(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	collections := args
	queries, _ := cmd.Flags().GetStringSlice("query")
	topK, _ := cmd.Flags().GetInt("top-k")
	reportFile, _ := cmd.Flags().GetString("report")
	format, _ := cmd.Flags().GetString("format")

	// Load config
	cfg, err := utils.LoadConfigWithInteractiveHelp()
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to load config: %v", err))
		os.Exit(1)
	}

	// Get VDB client
	selection, err := utils.GetSelectedVectorDBs(cmd, cfg)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to determine vector databases: %v", err))
		os.Exit(1)
	}

	if len(selection.Configs) == 0 {
		utils.PrintError("No vector database configured")
		os.Exit(1)
	}

	dbConfig := &selection.Configs[0]
	client, err := utils.CreateVectorDBClient(dbConfig)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Failed to create VDB client: %v", err))
		os.Exit(1)
	}

	// Validate collections exist
	utils.PrintInfo(fmt.Sprintf("🔍 Validating %d collections...", len(collections)))
	for _, col := range collections {
		exists, err := client.CollectionExists(ctx, col)
		if err != nil || !exists {
			utils.PrintError(fmt.Sprintf("Collection '%s' not found", col))
			os.Exit(1)
		}
	}
	utils.PrintSuccess(fmt.Sprintf("✓ All %d collections found", len(collections)))

	// Run comparison
	utils.PrintInfo(fmt.Sprintf("📊 Running %d queries across %d collections...", len(queries), len(collections)))
	report := generateComparison(ctx, client, collections, queries, topK)

	// Output report
	var output string
	if format == "json" {
		output = report.ToJSON()
	} else {
		output = report.ToMarkdown()
	}

	if reportFile != "" {
		if err := os.WriteFile(reportFile, []byte(output), 0644); err != nil {
			utils.PrintError(fmt.Sprintf("Failed to write report: %v", err))
			os.Exit(1)
		}
		utils.PrintSuccess(fmt.Sprintf("✓ Report saved to: %s", reportFile))
	} else {
		fmt.Println(output)
	}
}

// ComparisonReport contains the comparison results
type ComparisonReport struct {
	Collections []string
	Queries     []QueryComparison
	Generated   time.Time
}

// QueryComparison contains results for a single query
type QueryComparison struct {
	Query       string
	Collections []CollectionResults
}

// CollectionResults contains results for one collection
type CollectionResults struct {
	Collection string
	TopResults []Result
	AvgScore   float32
	Latency    time.Duration
}

// Result represents a single search result
type Result struct {
	DocID string
	Score float32
	Rank  int
}

func generateComparison(ctx context.Context, client vectordb.VectorDBClient, collections []string, queries []string, topK int) *ComparisonReport {
	report := &ComparisonReport{
		Collections: collections,
		Queries:     make([]QueryComparison, 0, len(queries)),
		Generated:   time.Now(),
	}

	for _, query := range queries {
		qc := QueryComparison{
			Query:       query,
			Collections: make([]CollectionResults, 0, len(collections)),
		}

		for _, col := range collections {
			start := time.Now()

			// Perform search
			options := &vectordb.QueryOptions{
				TopK: topK,
			}
			results, err := client.SearchSemantic(ctx, col, query, options)
			latency := time.Since(start)

			if err != nil {
				utils.PrintWarning(fmt.Sprintf("⚠️  Query '%s' failed for collection '%s': %v", query, col, err))
				continue
			}

			// Convert results
			colResults := CollectionResults{
				Collection: col,
				TopResults: make([]Result, 0, len(results)),
				Latency:    latency,
			}

			var scoreSum float64
			for i, r := range results {
				colResults.TopResults = append(colResults.TopResults, Result{
					DocID: r.Document.ID,
					Score: float32(r.Score),
					Rank:  i + 1,
				})
				scoreSum += r.Score
			}

			if len(results) > 0 {
				colResults.AvgScore = float32(scoreSum / float64(len(results)))
			}

			qc.Collections = append(qc.Collections, colResults)
		}

		report.Queries = append(report.Queries, qc)
	}

	return report
}

// ToMarkdown generates a markdown report
func (r *ComparisonReport) ToMarkdown() string {
	var sb strings.Builder

	sb.WriteString("# Embedding Model Comparison Report\n\n")
	sb.WriteString(fmt.Sprintf("**Generated**: %s\n\n", r.Generated.Format("2006-01-02 15:04:05")))

	sb.WriteString("## Collections Compared\n\n")
	for i, col := range r.Collections {
		sb.WriteString(fmt.Sprintf("%d. `%s`\n", i+1, col))
	}
	sb.WriteString("\n")

	// Summary table
	sb.WriteString("## Summary\n\n")
	sb.WriteString("| Collection | Avg Score | Avg Latency |\n")
	sb.WriteString("|------------|-----------|-------------|\n")

	// Calculate averages across all queries
	avgScores := make(map[string]float32)
	avgLatencies := make(map[string]time.Duration)
	counts := make(map[string]int)

	for _, qc := range r.Queries {
		for _, cr := range qc.Collections {
			avgScores[cr.Collection] += cr.AvgScore
			avgLatencies[cr.Collection] += cr.Latency
			counts[cr.Collection]++
		}
	}

	for _, col := range r.Collections {
		if counts[col] > 0 {
			avgScore := avgScores[col] / float32(counts[col])
			avgLatency := avgLatencies[col] / time.Duration(counts[col])
			sb.WriteString(fmt.Sprintf("| %s | %.3f | %s |\n", col, avgScore, avgLatency))
		}
	}
	sb.WriteString("\n")

	// Detailed query results
	for _, qc := range r.Queries {
		sb.WriteString(fmt.Sprintf("## Query: \"%s\"\n\n", qc.Query))

		// Sort collections by average score (descending)
		colResults := make([]CollectionResults, len(qc.Collections))
		copy(colResults, qc.Collections)
		sort.Slice(colResults, func(i, j int) bool {
			return colResults[i].AvgScore > colResults[j].AvgScore
		})

		// Results by collection
		for rank, cr := range colResults {
			sb.WriteString(fmt.Sprintf("### %d. %s\n\n", rank+1, cr.Collection))
			sb.WriteString(fmt.Sprintf("- **Average Score**: %.3f\n", cr.AvgScore))
			sb.WriteString(fmt.Sprintf("- **Latency**: %s\n\n", cr.Latency))

			sb.WriteString("**Top Results**:\n\n")
			sb.WriteString("| Rank | Doc ID | Score |\n")
			sb.WriteString("|------|--------|-------|\n")
			for _, res := range cr.TopResults {
				sb.WriteString(fmt.Sprintf("| %d | `%s` | %.3f |\n", res.Rank, res.DocID, res.Score))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// ToJSON generates a JSON report
func (r *ComparisonReport) ToJSON() string {
	// Simple JSON serialization
	var sb strings.Builder
	sb.WriteString("{\n")
	sb.WriteString(fmt.Sprintf("  \"generated\": \"%s\",\n", r.Generated.Format(time.RFC3339)))
	sb.WriteString("  \"collections\": [\n")
	for i, col := range r.Collections {
		sb.WriteString(fmt.Sprintf("    \"%s\"", col))
		if i < len(r.Collections)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	sb.WriteString("  ],\n")
	sb.WriteString("  \"queries\": [\n")
	for i, qc := range r.Queries {
		sb.WriteString("    {\n")
		sb.WriteString(fmt.Sprintf("      \"query\": \"%s\",\n", qc.Query))
		sb.WriteString("      \"results\": [\n")
		for j, cr := range qc.Collections {
			sb.WriteString("        {\n")
			sb.WriteString(fmt.Sprintf("          \"collection\": \"%s\",\n", cr.Collection))
			sb.WriteString(fmt.Sprintf("          \"avg_score\": %.3f,\n", cr.AvgScore))
			sb.WriteString(fmt.Sprintf("          \"latency_ms\": %d\n", cr.Latency.Milliseconds()))
			sb.WriteString("        }")
			if j < len(qc.Collections)-1 {
				sb.WriteString(",")
			}
			sb.WriteString("\n")
		}
		sb.WriteString("      ]\n")
		sb.WriteString("    }")
		if i < len(r.Queries)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	sb.WriteString("  ]\n")
	sb.WriteString("}\n")
	return sb.String()
}
