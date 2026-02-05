// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"golang.org/x/term"
)

// InteractiveFixer handles interactive fixing of configuration issues
type InteractiveFixer struct {
	issues  []ConfigIssue
	results []FixResult
	reader  *bufio.Reader
}

// NewInteractiveFixer creates a new interactive fixer
func NewInteractiveFixer(issues []ConfigIssue) *InteractiveFixer {
	return &InteractiveFixer{
		issues:  issues,
		results: make([]FixResult, 0, len(issues)),
		reader:  bufio.NewReader(os.Stdin),
	}
}

// Run executes the interactive fix flow
func (f *InteractiveFixer) Run() ([]FixResult, error) {
	// Check if running in a terminal
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return nil, fmt.Errorf("interactive mode requires a terminal")
	}

	// Show summary
	f.showSummary()

	// Ask if user wants to proceed
	if !f.confirmProceed() {
		return nil, fmt.Errorf("user cancelled")
	}

	// Process each issue
	for i, issue := range f.issues {
		fmt.Println()
		f.showSeparator()
		f.showIssueHeader(i+1, len(f.issues), issue)
		f.showSeparator()

		// Get fix action from user
		result, err := f.promptFixAction(issue)
		if err != nil {
			return nil, err
		}

		// Check if user wants to quit
		if result.Action == FixActionQuit {
			color.Yellow("\n⏸️  Stopping without applying changes")
			return nil, fmt.Errorf("user quit")
		}

		f.results = append(f.results, result)

		// Show result
		f.showActionResult(result)
	}

	// Show summary and confirm
	return f.showSummaryAndConfirm()
}

// showSummary shows initial summary of issues
func (f *InteractiveFixer) showSummary() {
	errorCount := 0
	warnCount := 0
	for _, issue := range f.issues {
		if issue.Type == IssueTypeError {
			errorCount++
		} else {
			warnCount++
		}
	}

	fmt.Println()
	color.Cyan("🔍 Found %d issue(s) in config.yaml", len(f.issues))
	if errorCount > 0 {
		color.Red("   • %d error(s)", errorCount)
	}
	if warnCount > 0 {
		color.Yellow("   • %d warning(s)", warnCount)
	}
	fmt.Println()
}

// confirmProceed asks user if they want to proceed
func (f *InteractiveFixer) confirmProceed() bool {
	color.Cyan("Would you like to fix these issues? (Y/n): ")
	response, err := f.reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "" || response == "y" || response == "yes"
}

// showSeparator shows a visual separator
func (f *InteractiveFixer) showSeparator() {
	color.New(color.FgCyan).Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

// showIssueHeader shows the issue header
func (f *InteractiveFixer) showIssueHeader(current, total int, issue ConfigIssue) {
	if issue.Type == IssueTypeError {
		color.Red("Error %d/%d: %s missing %s", current, total, issue.DatabaseName, issue.Field)
	} else {
		color.Yellow("Warning %d/%d: %s missing %s", current, total, issue.DatabaseName, issue.Field)
	}

	fmt.Println()

	if issue.DatabaseName != "" {
		color.White("Database: %s (databases.vector_databases[%d])", issue.DatabaseName, issue.DatabaseIdx)
	}
	color.White("Field: %s", issue.Field)
	color.White("Issue: %s", issue.Message)

	if issue.Hint != "" {
		fmt.Println()
		color.New(color.FgHiBlack).Printf("💡 Tip: %s\n", issue.Hint)
	}

	// Show example
	example := issue.GetFieldExample()
	fmt.Println()
	color.New(color.FgHiBlack).Printf("   Example: %s\n", example)
}

// promptFixAction prompts the user for a fix action
func (f *InteractiveFixer) promptFixAction(issue ConfigIssue) (FixResult, error) {
	fmt.Println()
	color.White("Options:")

	options := issue.GetFixPromptOptions()
	for i, opt := range options {
		fmt.Printf("  %d. %s\n", i+1, opt)
	}

	fmt.Println()
	color.Cyan("Your choice (1-%d): ", len(options))

	// Read choice
	input, err := f.reader.ReadString('\n')
	if err != nil {
		return FixResult{}, fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(input)
	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(options) {
		color.Red("❌ Invalid choice, skipping this issue")
		return FixResult{
			Issue:  issue,
			Action: FixActionSkip,
		}, nil
	}

	// Map choice to action
	action := MapOptionToAction(choice, issue)

	// Handle action
	switch action {
	case FixActionSetValue:
		return f.promptForValue(issue)
	case FixActionSkip:
		return FixResult{
			Issue:  issue,
			Action: FixActionSkip,
		}, nil
	case FixActionRemove:
		return FixResult{
			Issue:  issue,
			Action: FixActionRemove,
		}, nil
	case FixActionDisable:
		return FixResult{
			Issue:  issue,
			Action: FixActionDisable,
		}, nil
	case FixActionQuit:
		return FixResult{
			Issue:  issue,
			Action: FixActionQuit,
		}, nil
	default:
		return FixResult{
			Issue:  issue,
			Action: FixActionSkip,
		}, nil
	}
}

// promptForValue prompts user to enter a value
func (f *InteractiveFixer) promptForValue(issue ConfigIssue) (FixResult, error) {
	fmt.Println()

	// Allow up to 3 attempts for validation
	maxAttempts := 3
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		var value string

		if issue.IsSecretField() {
			color.Cyan("Enter %s (input hidden): ", issue.Field)

			// Read password (hidden)
			fd := int(os.Stdin.Fd())
			password, readErr := term.ReadPassword(fd)
			if readErr != nil {
				return FixResult{}, fmt.Errorf("failed to read secret: %w", readErr)
			}
			fmt.Println() // New line after password input

			value = strings.TrimSpace(string(password))
		} else {
			color.Cyan("Enter %s: ", issue.Field)

			input, readErr := f.reader.ReadString('\n')
			if readErr != nil {
				return FixResult{}, fmt.Errorf("failed to read input: %w", readErr)
			}

			value = strings.TrimSpace(input)
		}

		// Check for empty
		if value == "" {
			color.Yellow("⚠️  Empty value, skipping")
			return FixResult{
				Issue:  issue,
				Action: FixActionSkip,
			}, nil
		}

		// Show smart suggestions before validation
		suggestions := GetFieldSuggestions(issue.Field, value)
		if len(suggestions) > 0 {
			fmt.Println()
			for _, suggestion := range suggestions {
				color.Cyan(suggestion)
			}
			fmt.Println()
		}

		// Validate the value
		if validationErr := ValidateFieldValue(issue.Field, value); validationErr != nil {
			color.Red("❌ Validation failed: %s", validationErr.Error())

			// Show helpful hint
			if hint := GetValidationHint(issue.Field, validationErr); hint != "" {
				fmt.Println(hint)
			}

			if attempt < maxAttempts {
				color.Yellow("\nTry again (%d/%d attempts remaining)...", maxAttempts-attempt, maxAttempts)
				fmt.Println()
				continue
			} else {
				color.Yellow("\n⚠️  Max attempts reached, skipping")
				return FixResult{
					Issue:  issue,
					Action: FixActionSkip,
				}, nil
			}
		}

		// Show positive feedback for good values
		if len(suggestions) == 0 {
			// If no suggestions were shown and validation passed, it's likely a good value
			fmt.Println()
		}

		// Validation passed
		return FixResult{
			Issue:  issue,
			Action: FixActionSetValue,
			Value:  value,
		}, nil
	}

	// Should not reach here
	return FixResult{
		Issue:  issue,
		Action: FixActionSkip,
	}, nil
}

// showActionResult shows the result of an action
func (f *InteractiveFixer) showActionResult(result FixResult) {
	switch result.Action {
	case FixActionSetValue:
		color.Green("✅ Value set")
	case FixActionSkip:
		color.Yellow("⏭️  Skipped")
	case FixActionRemove:
		color.Yellow("🗑️  Will remove database from config")
	case FixActionDisable:
		color.Yellow("🚫 Will disable database (comment out)")
	}
}

// showSummaryAndConfirm shows summary and asks for confirmation
func (f *InteractiveFixer) showSummaryAndConfirm() ([]FixResult, error) {
	fmt.Println()
	f.showSeparator()
	color.New(color.FgCyan, color.Bold).Println("📋 Summary")
	f.showSeparator()
	fmt.Println()

	color.White("Changes to be made:")
	for _, result := range f.results {
		issue := result.Issue
		switch result.Action {
		case FixActionSetValue:
			displayValue := result.Value
			if issue.IsSecretField() {
				displayValue = maskSecret(displayValue)
			}
			color.Green("  ✅ %s.%s = %s", issue.DatabaseName, issue.Field, displayValue)
		case FixActionSkip:
			color.Yellow("  ⏭️  %s.%s (skipped)", issue.DatabaseName, issue.Field)
		case FixActionRemove:
			color.Red("  🗑️  Remove database: %s", issue.DatabaseName)
		case FixActionDisable:
			color.Yellow("  🚫 Disable database: %s", issue.DatabaseName)
		}
	}

	fmt.Println()
	color.New(color.FgYellow, color.Bold).Print("Apply changes to config.yaml? (Y/n): ")

	response, err := f.reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read confirmation: %w", err)
	}

	response = strings.ToLower(strings.TrimSpace(response))
	if response != "" && response != "y" && response != "yes" {
		color.Yellow("\n❌ Changes not applied")
		return nil, fmt.Errorf("user cancelled")
	}

	return f.results, nil
}
