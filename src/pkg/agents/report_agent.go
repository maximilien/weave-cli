// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package agents

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/llm"
)

// ReportAgent creates comprehensive operation reports
type ReportAgent struct {
	outputAgent *OutputAgent
	llmClient   llm.Client
}

// NewReportAgent creates a new ReportAgent
func NewReportAgent(outputAgent *OutputAgent, llmClient llm.Client) *ReportAgent {
	return &ReportAgent{
		outputAgent: outputAgent,
		llmClient:   llmClient,
	}
}

// Name returns the agent's name
func (a *ReportAgent) Name() string {
	return "ReportAgent"
}

// Execute generates a report from the operation results
func (a *ReportAgent) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	report, ok := input.(*OperationReport)
	if !ok {
		return nil, fmt.Errorf("invalid input type for ReportAgent")
	}

	// Enhance the report with analysis
	a.enhanceReport(report)

	return report, nil
}

// PrintReport displays the operation report
func (a *ReportAgent) PrintReport(report *OperationReport) {
	fmt.Println()
	a.outputAgent.printHeader("Results")
	fmt.Println()

	// Print summary
	fmt.Printf("Query: %s\n", report.QueryIntent)
	fmt.Printf("Operations: %d commands executed\n", report.ExecutedSteps)
	fmt.Printf("Duration: %v\n", FormatDuration(report.Duration))

	successRate := float64(report.SuccessfulSteps) / float64(report.ExecutedSteps) * 100
	fmt.Printf("Success rate: %.1f%%\n", successRate)
	fmt.Println()

	// Print command results if verbose or if there were failures
	if report.FailedSteps > 0 || a.outputAgent.verbose {
		fmt.Println("Command Results:")
		for i, cmd := range report.Commands {
			fmt.Printf("  %d. %s", i+1, cmd.Command)
			if cmd.Success {
				fmt.Printf(" ✓ (%v)\n", FormatDuration(cmd.Duration))
				if cmd.Output != "" && a.outputAgent.verbose {
					fmt.Printf("     Output: %s\n", truncateMultiline(cmd.Output, 200))
				}
			} else {
				fmt.Printf(" ✗\n")
				if cmd.Error != "" {
					fmt.Printf("     Error: %s\n", cmd.Error)
				}
			}
		}
		fmt.Println()
	}

	// Print summary if available
	if report.Summary != "" {
		fmt.Println(report.Summary)
		fmt.Println()
	}

	// Print recommendations
	if len(report.Recommendations) > 0 {
		fmt.Println("💡 Recommendations:")
		for _, rec := range report.Recommendations {
			fmt.Printf("  • %s\n", rec)
		}
		fmt.Println()
	}

	// Print next steps
	if len(report.NextSteps) > 0 {
		fmt.Println("Next Steps:")
		for i, step := range report.NextSteps {
			fmt.Printf("  %d. %s\n", i+1, step)
		}
		fmt.Println()
	}

	// Print final status
	if report.FailedSteps == 0 {
		a.outputAgent.PrintSuccess(fmt.Sprintf("Query completed (%v)", FormatDuration(report.Duration)))
	} else {
		a.outputAgent.PrintWarning(fmt.Sprintf("Query completed with %d failures", report.FailedSteps))
	}
}

// enhanceReport adds analysis and recommendations to the report
func (a *ReportAgent) enhanceReport(report *OperationReport) {
	// Generate summary if not present
	if report.Summary == "" {
		report.Summary = a.generateSummary(report)
	}

	// Add recommendations based on results
	report.Recommendations = a.generateRecommendations(report)

	// Add next steps
	report.NextSteps = a.generateNextSteps(report)
}

// generateSummary creates a summary of the operation
func (a *ReportAgent) generateSummary(report *OperationReport) string {
	var summary strings.Builder

	if report.SuccessfulSteps == report.ExecutedSteps {
		summary.WriteString(fmt.Sprintf("Successfully completed all %d operations.", report.ExecutedSteps))
	} else {
		summary.WriteString(fmt.Sprintf("Completed %d of %d operations successfully.",
			report.SuccessfulSteps, report.ExecutedSteps))
		if report.FailedSteps > 0 {
			summary.WriteString(fmt.Sprintf(" %d operations failed.", report.FailedSteps))
		}
	}

	return summary.String()
}

// generateRecommendations generates recommendations based on the report using LLM
func (a *ReportAgent) generateRecommendations(report *OperationReport) []string {
	// Build context for LLM
	ctx := context.Background()
	prompt := a.buildRecommendationsPrompt(report)

	// Use LLM to generate recommendations
	response, err := a.llmClient.Complete(ctx, prompt,
		llm.WithSystemMessage("You are a helpful assistant that provides concise, actionable recommendations for weave-cli users."),
		llm.WithTemperature(0.3),
		llm.WithMaxTokens(500))

	if err != nil {
		// Fallback to simple recommendations if LLM fails
		return a.generateFallbackRecommendations(report)
	}

	// Parse recommendations from response (one per line starting with •)
	recommendations := []string{}
	for _, line := range strings.Split(response, "\n") {
		line = strings.TrimSpace(line)
		// Remove bullet points and clean up
		line = strings.TrimPrefix(line, "•")
		line = strings.TrimPrefix(line, "-")
		line = strings.TrimPrefix(line, "*")
		line = strings.TrimSpace(line)

		if line != "" && len(line) > 10 { // Skip very short lines
			recommendations = append(recommendations, line)
		}
	}

	// Limit to 3 recommendations
	if len(recommendations) > 3 {
		recommendations = recommendations[:3]
	}

	return recommendations
}

// buildRecommendationsPrompt builds the prompt for generating recommendations
func (a *ReportAgent) buildRecommendationsPrompt(report *OperationReport) string {
	commandsSummary := ""
	for i, cmd := range report.Commands {
		status := "✓"
		if !cmd.Success {
			status = "✗"
		}
		commandsSummary += fmt.Sprintf("\n  %d. %s %s", i+1, status, cmd.Command)
		if !cmd.Success {
			commandsSummary += fmt.Sprintf(" (Error: %s)", cmd.Error)
		}
	}

	return fmt.Sprintf(`Based on this weave-cli operation, provide 2-3 concise, actionable recommendations:

Query Intent: %s
Success Rate: %.0f%% (%d/%d steps succeeded)
Duration: %v
Commands:%s

Guidelines:
- If the user asked to count documents but only got a list, suggest how to count (use weave docs count COLLECTION)
- If the user asked about empty collections, suggest checking each collection
- Keep recommendations practical and specific to weave-cli commands
- Format as bullet points (one per line)
- Focus on what the user can do next
- Include example commands when helpful

Provide ONLY the recommendations (2-3 bullet points), nothing else:`,
		report.QueryIntent,
		float64(report.SuccessfulSteps)/float64(report.ExecutedSteps)*100,
		report.SuccessfulSteps,
		report.ExecutedSteps,
		FormatDuration(report.Duration),
		commandsSummary)
}

// generateFallbackRecommendations provides simple recommendations if LLM fails
func (a *ReportAgent) generateFallbackRecommendations(report *OperationReport) []string {
	recommendations := []string{}

	// Check for failures
	if report.FailedSteps > 0 {
		recommendations = append(recommendations,
			"Review the error messages and fix any issues")
	}

	// Success recommendations
	if report.SuccessfulSteps == report.ExecutedSteps {
		intent := strings.ToLower(report.QueryIntent)

		if strings.Contains(intent, "list") {
			recommendations = append(recommendations,
				"Use `weave cols show COLLECTION` for detailed collection information.")

			// Check if user wanted counts but we only listed
			if strings.Contains(intent, "count") {
				recommendations = append(recommendations,
					"To count documents in each collection, use `weave docs count COLLECTION`")
			}
		}
	}

	return recommendations
}

// generateNextSteps generates next steps based on the report
func (a *ReportAgent) generateNextSteps(report *OperationReport) []string {
	nextSteps := []string{}

	// If there were failures, suggest retry or investigation
	if report.FailedSteps > 0 {
		nextSteps = append(nextSteps, "Review the error messages and fix any issues")
		nextSteps = append(nextSteps, "Retry the failed operations with `--verbose` for more details")
	}

	// Intent-specific next steps
	intent := strings.ToLower(report.QueryIntent)

	if strings.Contains(intent, "list") || strings.Contains(intent, "empty") {
		if report.SuccessfulSteps > 0 {
			nextSteps = append(nextSteps, "Delete empty collections with `weave cols del COLLECTION`")
		}
	}

	if strings.Contains(intent, "create") || strings.Contains(intent, "add") {
		if report.SuccessfulSteps > 0 {
			nextSteps = append(nextSteps, "Query the documents with `weave cols q COLLECTION \"query\"`")
			nextSteps = append(nextSteps, "View document statistics with `weave docs ls COLLECTION -S`")
		}
	}

	return nextSteps
}

// CreateReport creates an operation report from execution results
func CreateReport(intent string, startTime time.Time, commands []CommandReport) *OperationReport {
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	successfulSteps := 0
	failedSteps := 0

	for _, cmd := range commands {
		if cmd.Success {
			successfulSteps++
		} else {
			failedSteps++
		}
	}

	return &OperationReport{
		QueryIntent:     intent,
		ExecutedSteps:   len(commands),
		SuccessfulSteps: successfulSteps,
		FailedSteps:     failedSteps,
		StartTime:       startTime,
		EndTime:         endTime,
		Duration:        duration,
		Commands:        commands,
	}
}

// truncateMultiline truncates a multiline string while preserving line structure
func truncateMultiline(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	// Take first maxLen characters and add ellipsis
	truncated := s[:maxLen]

	// Try to end at a newline for cleaner display
	if lastNewline := strings.LastIndex(truncated, "\n"); lastNewline > maxLen/2 {
		truncated = truncated[:lastNewline]
	}

	return truncated + "..."
}
