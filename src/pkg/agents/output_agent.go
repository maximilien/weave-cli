// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package agents

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
)

// OutputAgent formats and displays information to users
type OutputAgent struct {
	noColor      bool
	verbose      bool
	quiet        bool
	outputFormat string
}

// OutputConfig configures the OutputAgent
type OutputConfig struct {
	NoColor      bool
	Verbose      bool
	Quiet        bool
	OutputFormat string // "text", "json", "yaml"
}

// NewOutputAgent creates a new OutputAgent
func NewOutputAgent(config OutputConfig) *OutputAgent {
	return &OutputAgent{
		noColor:      config.NoColor,
		verbose:      config.Verbose,
		quiet:        config.Quiet,
		outputFormat: config.OutputFormat,
	}
}

// Name returns the agent's name
func (a *OutputAgent) Name() string {
	return "OutputAgent"
}

// Execute formats and displays output (not typically used directly)
func (a *OutputAgent) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	// OutputAgent is primarily used through its formatting methods
	return input, nil
}

// PrintPlan displays the execution plan to the user
func (a *OutputAgent) PrintPlan(plan *ExecutionPlan) {
	if a.quiet {
		return
	}

	a.printHeader("Query Plan")
	fmt.Println()

	fmt.Printf("Intent: %s\n", plan.Summary)
	fmt.Printf("Estimated time: %s\n", plan.Estimations.Duration)
	fmt.Printf("Risk level: %s\n", a.colorizeRisk(plan.Estimations.Risk))
	fmt.Println()

	fmt.Println("Steps:")
	for i, step := range plan.Steps {
		a.printStep(i+1, &step)
	}
	fmt.Println()

	if len(plan.Warnings) > 0 {
		a.printWarnings(plan.Warnings)
		fmt.Println()
	}
}

// PrintStepProgress prints progress for a step
func (a *OutputAgent) PrintStepProgress(stepNum int, step *ExecutionStep, status string) {
	if a.quiet {
		return
	}

	icon := "⏳"
	fmt.Printf("%s Step %d: %s\n", icon, stepNum, step.Description)
}

// PrintStepCompletion prints step completion with duration
func (a *OutputAgent) PrintStepCompletion(stepNum int, duration time.Duration) {
	if a.quiet {
		return
	}

	durationStr := FormatDuration(duration)
	if a.noColor {
		fmt.Printf("✓ Step %d completed (%s)\n", stepNum, durationStr)
	} else {
		color.New(color.FgGreen).Printf("✓ Step %d completed (%s)\n", stepNum, durationStr)
	}
}

// PrintStepError prints step error
func (a *OutputAgent) PrintStepError(stepNum int, errMsg string) {
	if a.quiet {
		return
	}

	if a.noColor {
		fmt.Printf("✗ Step %d failed: %s\n", stepNum, errMsg)
	} else {
		color.New(color.FgRed).Printf("✗ Step %d failed: %s\n", stepNum, errMsg)
	}
}

// PrintSuccess prints a success message
func (a *OutputAgent) PrintSuccess(message string) {
	if a.noColor {
		fmt.Printf("✅ %s\n", message)
	} else {
		color.New(color.FgGreen).Printf("✅ %s\n", message)
	}
}

// PrintError prints an error message
func (a *OutputAgent) PrintError(message string) {
	if a.noColor {
		fmt.Printf("❌ %s\n", message)
	} else {
		color.New(color.FgRed).Printf("❌ %s\n", message)
	}
}

// PrintWarning prints a warning message
func (a *OutputAgent) PrintWarning(message string) {
	if a.noColor {
		fmt.Printf("⚠️  %s\n", message)
	} else {
		color.New(color.FgYellow).Printf("⚠️  %s\n", message)
	}
}

// PrintInfo prints an info message
func (a *OutputAgent) PrintInfo(message string) {
	if a.quiet {
		return
	}
	fmt.Printf("ℹ️  %s\n", message)
}

// PrintCommandOutput prints the output from a command execution
func (a *OutputAgent) PrintCommandOutput(output string) {
	if a.quiet {
		return
	}

	// Try to extract text content from MCP response format
	cleaned := a.extractMCPContent(output)

	// Colorize JSON if not in no-color mode
	if !a.noColor && isJSON(cleaned) {
		cleaned = colorizeJSON(cleaned)
	}

	// Indent the output for better readability
	lines := strings.Split(strings.TrimSpace(cleaned), "\n")
	for _, line := range lines {
		if line != "" {
			fmt.Printf("  %s\n", line)
		}
	}
}

// isJSON checks if a string is valid JSON
func isJSON(s string) bool {
	s = strings.TrimSpace(s)
	return (strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}")) ||
		(strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]"))
}

// colorizeJSON adds syntax highlighting to JSON output
func colorizeJSON(jsonStr string) string {
	// Simple JSON colorization
	lines := strings.Split(jsonStr, "\n")
	for i, line := range lines {
		// Color keys (words before :)
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				// Color the key in cyan
				key := parts[0]
				value := parts[1]

				// Check if value is a string, number, boolean, or null
				trimmedValue := strings.TrimSpace(value)
				coloredValue := value

				if strings.HasPrefix(trimmedValue, "\"") {
					// String value - green
					coloredValue = color.GreenString(value)
				} else if trimmedValue == "true" || trimmedValue == "false" {
					// Boolean - yellow
					coloredValue = color.YellowString(value)
				} else if trimmedValue == "null" || trimmedValue == "null," {
					// Null - red
					coloredValue = color.RedString(value)
				} else if len(trimmedValue) > 0 && (trimmedValue[0] >= '0' && trimmedValue[0] <= '9') {
					// Number - magenta
					coloredValue = color.MagentaString(value)
				}

				lines[i] = color.CyanString(key) + ":" + coloredValue
			}
		} else if strings.Contains(line, "\"") && !strings.Contains(line, ":") {
			// Array string values - green
			lines[i] = color.GreenString(line)
		}
	}
	return strings.Join(lines, "\n")
}

// extractMCPContent extracts the text content from MCP response format
func (a *OutputAgent) extractMCPContent(output string) string {
	// MCP responses come as: [map[text:{json} type:text]]
	// We want to extract just the JSON content

	// Look for patterns like: text:{...}
	if strings.Contains(output, "text:{") {
		// Find the start and end of JSON content
		start := strings.Index(output, "text:{")
		if start >= 0 {
			start += 5 // Move past "text:"

			// Find the matching closing brace
			bracketCount := 0
			for i := start; i < len(output); i++ {
				if output[i] == '{' {
					bracketCount++
				} else if output[i] == '}' {
					bracketCount--
					if bracketCount == 0 {
						return output[start : i+1]
					}
				}
			}
		}
	}

	// If we can't extract, return the original
	return output
}

// CreateProgressBar creates a progress bar for tracking operations
func (a *OutputAgent) CreateProgressBar(max int, description string) *progressbar.ProgressBar {
	if a.quiet {
		return progressbar.DefaultSilent(int64(max))
	}

	return progressbar.NewOptions(max,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWidth(40),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)
}

// PrintCommandResult prints the result of a command execution
func (a *OutputAgent) PrintCommandResult(report *CommandReport) {
	if a.quiet && report.Success {
		return
	}

	if report.Success {
		if a.verbose {
			fmt.Printf("  ✓ %s (%v)\n", report.Command, report.Duration)
			if report.Output != "" {
				fmt.Printf("    Output: %s\n", truncate(report.Output, 100))
			}
		}
	} else {
		a.PrintError(fmt.Sprintf("%s failed", report.Command))
		if report.Error != "" {
			fmt.Printf("    Error: %s\n", report.Error)
		}
	}
}

// PrintMetrics prints evaluation metrics
func (a *OutputAgent) PrintMetrics(metrics *EvaluationMetrics) {
	if a.quiet {
		return
	}

	// Always show basic metrics (not just with --verbose)
	fmt.Println()
	a.printHeader("AI Metrics")

	// LLM usage summary
	if metrics.LLMInvocations > 0 {
		if a.noColor {
			fmt.Printf("  LLM calls: %d\n", metrics.LLMInvocations)
			fmt.Printf("  Tokens: %s (prompt: %s, completion: %s)\n",
				formatNumber(metrics.TotalTokens),
				formatNumber(metrics.PromptTokens),
				formatNumber(metrics.CompletionTokens))
			fmt.Printf("  Cost: $%.4f\n", metrics.TotalCost)
		} else {
			color.New(color.FgCyan).Printf("  LLM calls: ")
			fmt.Printf("%d\n", metrics.LLMInvocations)

			color.New(color.FgCyan).Printf("  Tokens: ")
			fmt.Printf("%s ", formatNumber(metrics.TotalTokens))
			color.New(color.FgHiBlack).Printf("(prompt: %s, completion: %s)\n",
				formatNumber(metrics.PromptTokens),
				formatNumber(metrics.CompletionTokens))

			color.New(color.FgCyan).Printf("  Cost: ")
			if metrics.TotalCost < 0.01 {
				color.New(color.FgGreen).Printf("$%.4f\n", metrics.TotalCost)
			} else if metrics.TotalCost < 0.10 {
				color.New(color.FgYellow).Printf("$%.4f\n", metrics.TotalCost)
			} else {
				color.New(color.FgRed).Printf("$%.4f\n", metrics.TotalCost)
			}
		}
	}

	// Show Opik/OpenTelemetry link
	if a.noColor {
		fmt.Println("  💡 View detailed traces in Opik dashboard: https://www.comet.com/opik")
	} else {
		color.New(color.FgHiBlack).Println("  💡 View detailed traces in Opik dashboard: https://www.comet.com/opik")
	}

	// Show detailed breakdown with --verbose
	if a.verbose {
		fmt.Println()
		if a.noColor {
			fmt.Println("  Detailed Metrics:")
		} else {
			color.New(color.FgCyan, color.Bold).Println("  Detailed Metrics:")
		}
		fmt.Printf("    Query ID: %s\n", metrics.QueryID)
		fmt.Printf("    Success: %v\n", metrics.Success)
		fmt.Printf("    Intent Matched: %v\n", metrics.IntentMatched)
		fmt.Printf("    Latency: %v\n", metrics.Latency)
		fmt.Printf("    Error Rate: %.1f%%\n", metrics.ErrorRate*100)

		if metrics.UserSatisfaction != nil {
			fmt.Printf("    User Satisfaction: %.1f/5.0\n", *metrics.UserSatisfaction)
		}

		// Token breakdown
		if metrics.TotalTokens > 0 {
			promptPct := float64(metrics.PromptTokens) / float64(metrics.TotalTokens) * 100
			completionPct := float64(metrics.CompletionTokens) / float64(metrics.TotalTokens) * 100
			fmt.Printf("    Token Distribution: %.1f%% prompt, %.1f%% completion\n", promptPct, completionPct)
		}
	}
}

// formatNumber formats large numbers with commas for readability
func formatNumber(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 1000000 {
		return fmt.Sprintf("%d,%03d", n/1000, n%1000)
	}
	return fmt.Sprintf("%d,%03d,%03d", n/1000000, (n%1000000)/1000, n%1000)
}

// PrintRejectionMessage prints a message when query is rejected
func (a *OutputAgent) PrintRejectionMessage(reason string) {
	a.PrintError("Unable to process query")
	fmt.Println()
	fmt.Println(reason)
	fmt.Println()
	fmt.Println("Weave CLI helps you:")
	fmt.Println("  • Manage collections")
	fmt.Println("  • Create and query documents")
	fmt.Println("  • Process PDFs and images")
	fmt.Println("  • Search vector databases")
	fmt.Println()
	fmt.Println("Try rephrasing your query or see `weave help` for available commands.")
}

// AskConfirmation asks the user for confirmation
func (a *OutputAgent) AskConfirmation(message string) (bool, error) {
	fmt.Printf("\n%s [Y/n]: ", message)

	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		// If no input (just Enter), default to Yes
		return true, nil
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "" || response == "y" || response == "yes", nil
}

// Private helper methods

func (a *OutputAgent) printHeader(message string) {
	if a.noColor {
		fmt.Printf("📋 %s\n", message)
	} else {
		color.New(color.FgBlue, color.Bold).Printf("📋 %s\n", message)
	}
}

func (a *OutputAgent) printStep(num int, step *ExecutionStep) {
	typeIcon := "🔧"
	if step.Type == "bash" {
		typeIcon = "⚙️"
	} else if step.Type == "weave" {
		typeIcon = "🔍"
	}

	destructiveMarker := ""
	if step.Destructive {
		destructiveMarker = " ⚠️"
	}

	fmt.Printf("  %d. [%s %s] %s%s\n", num, typeIcon, step.Type, step.Description, destructiveMarker)
}

func (a *OutputAgent) printWarnings(warnings []string) {
	for _, warning := range warnings {
		a.PrintWarning(warning)
	}
}

func (a *OutputAgent) colorizeRisk(risk string) string {
	if a.noColor {
		return risk
	}

	switch strings.ToLower(risk) {
	case "low":
		return color.GreenString(risk)
	case "medium":
		return color.YellowString(risk)
	case "high":
		return color.RedString(risk)
	default:
		return risk
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// FormatDuration formats a duration for display
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%.1fm", d.Minutes())
}
