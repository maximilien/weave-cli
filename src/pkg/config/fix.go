// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package config

import (
	"fmt"
	"strings"
)

// IssueType represents the severity of a configuration issue
type IssueType string

const (
	IssueTypeError   IssueType = "error"
	IssueTypeWarning IssueType = "warning"
)

// ConfigIssue represents a configuration validation issue that can be fixed
type ConfigIssue struct {
	Type         IssueType // error or warning
	Path         string    // Full path like "databases.vector_databases[2].database_url"
	DatabaseName string    // Database name like "mongodb-cloud"
	DatabaseIdx  int       // Index in vector_databases array (-1 if not applicable)
	Field        string    // Field name like "database_url"
	Message      string    // Human-readable description
	Hint         string    // Fix suggestion
	CurrentValue string    // Current value (empty if missing)
}

// FixAction represents what to do to fix an issue
type FixAction string

const (
	FixActionSetValue FixAction = "set_value"
	FixActionSkip     FixAction = "skip"
	FixActionRemove   FixAction = "remove"
	FixActionDisable  FixAction = "disable"
	FixActionQuit     FixAction = "quit"
)

// FixResult tracks the outcome of fixing an issue
type FixResult struct {
	Issue  ConfigIssue
	Action FixAction
	Value  string // New value (if action is set_value)
	Error  error
}

// ParseValidationOutput parses validation error/warning output into ConfigIssue list
// Expected format from validation:
//
//	"⚠️  Configuration Errors:"
//	"  • databases.vector_databases[2] (mongodb-cloud).database_url: MongoDB database_url is required"
//	"    💡 Set 'database_url' field with MongoDB Atlas connection string"
func ParseValidationOutput(output string) ([]ConfigIssue, error) {
	var issues []ConfigIssue
	lines := strings.Split(output, "\n")

	var currentIssueType IssueType
	var currentIssue *ConfigIssue

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Check for section headers
		if strings.Contains(line, "Configuration Errors:") {
			currentIssueType = IssueTypeError
			continue
		}
		if strings.Contains(line, "Configuration Warnings:") {
			currentIssueType = IssueTypeWarning
			continue
		}

		// Skip empty lines
		if line == "" {
			continue
		}

		// Parse issue line: "  • databases.vector_databases[2] (mongodb-cloud).database_url: MongoDB database_url is required"
		if strings.HasPrefix(line, "•") {
			line = strings.TrimPrefix(line, "•")
			line = strings.TrimSpace(line)

			// Split by ": " to separate path from message
			parts := strings.SplitN(line, ": ", 2)
			if len(parts) != 2 {
				continue
			}

			pathPart := parts[0]
			message := parts[1]

			// Parse path: "databases.vector_databases[2] (mongodb-cloud).database_url"
			issue := ConfigIssue{
				Type:        currentIssueType,
				Path:        "",
				DatabaseIdx: -1,
				Message:     message,
			}

			// Extract database name from (name) pattern
			if startIdx := strings.Index(pathPart, "("); startIdx != -1 {
				endIdx := strings.Index(pathPart[startIdx:], ")")
				if endIdx != -1 {
					issue.DatabaseName = pathPart[startIdx+1 : startIdx+endIdx]
				}
			}

			// Extract array index [N]
			if startIdx := strings.Index(pathPart, "["); startIdx != -1 {
				endIdx := strings.Index(pathPart[startIdx:], "]")
				if endIdx != -1 {
					var idx int
					fmt.Sscanf(pathPart[startIdx+1:startIdx+endIdx], "%d", &idx)
					issue.DatabaseIdx = idx
				}
			}

			// Extract field name (last part after final dot)
			if dotIdx := strings.LastIndex(pathPart, "."); dotIdx != -1 {
				issue.Field = pathPart[dotIdx+1:]
			}

			// Full path for YAML navigation
			issue.Path = pathPart

			currentIssue = &issue
			issues = append(issues, issue)
		}

		// Parse hint line: "    💡 Set 'database_url' field with MongoDB Atlas connection string"
		if strings.Contains(line, "💡") && currentIssue != nil {
			// Extract hint text after 💡
			if idx := strings.Index(line, "💡"); idx != -1 {
				hint := strings.TrimSpace(line[idx+len("💡"):])
				issues[len(issues)-1].Hint = hint
			}
		}
	}

	return issues, nil
}

// String returns a human-readable representation of the issue
func (i ConfigIssue) String() string {
	var sb strings.Builder

	if i.Type == IssueTypeError {
		sb.WriteString("❌ Error: ")
	} else {
		sb.WriteString("⚠️  Warning: ")
	}

	if i.DatabaseName != "" {
		sb.WriteString(i.DatabaseName)
		sb.WriteString(" - ")
	}

	sb.WriteString(i.Message)

	return sb.String()
}

// GetFixPromptOptions returns the available fix options for this issue
func (i ConfigIssue) GetFixPromptOptions() []string {
	options := []string{
		"Enter value now",
		"Skip (fix later)",
	}

	// For database-level issues, offer remove/disable
	if i.DatabaseIdx >= 0 {
		options = append(options, "Remove this database from config")
		options = append(options, "Disable this database (comment out)")
	}

	options = append(options, "Quit")

	return options
}

// MapOptionToAction maps user's numeric choice to FixAction
func MapOptionToAction(choice int, issue ConfigIssue) FixAction {
	if issue.DatabaseIdx >= 0 {
		// Options: 1=Enter, 2=Skip, 3=Remove, 4=Disable, 5=Quit
		switch choice {
		case 1:
			return FixActionSetValue
		case 2:
			return FixActionSkip
		case 3:
			return FixActionRemove
		case 4:
			return FixActionDisable
		case 5:
			return FixActionQuit
		default:
			return FixActionSkip
		}
	} else {
		// Options: 1=Enter, 2=Skip, 3=Quit
		switch choice {
		case 1:
			return FixActionSetValue
		case 2:
			return FixActionSkip
		case 3:
			return FixActionQuit
		default:
			return FixActionSkip
		}
	}
}

// IsSecretField returns true if the field should be treated as a secret
func (i ConfigIssue) IsSecretField() bool {
	secretKeywords := []string{
		"api_key",
		"apikey",
		"token",
		"secret",
		"password",
		"passwd",
		"credentials",
	}

	fieldLower := strings.ToLower(i.Field)
	for _, keyword := range secretKeywords {
		if strings.Contains(fieldLower, keyword) {
			return true
		}
	}

	return false
}

// GetFieldExample returns an example value for this field
func (i ConfigIssue) GetFieldExample() string {
	fieldLower := strings.ToLower(i.Field)

	examples := map[string]string{
		"url":          "https://your-instance.cloud-provider.com",
		"database_url": "mongodb+srv://user:pass@cluster.mongodb.net/db",
		"host":         "localhost",
		"port":         "8080",
		"api_key":      "your-api-key-here",
		"token":        "your-token-here",
		"tenant":       "your-tenant-id",
		"database":     "your-database-name",
		"username":     "your-username",
		"password":     "your-password",
	}

	// Check for exact match
	if example, ok := examples[fieldLower]; ok {
		return example
	}

	// Check for partial match
	for key, example := range examples {
		if strings.Contains(fieldLower, key) {
			return example
		}
	}

	return "your-value-here"
}
