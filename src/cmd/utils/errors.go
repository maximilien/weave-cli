// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package utils

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// WeaviateError represents a structured error from Weaviate
type WeaviateError struct {
	Error []struct {
		Message string `json:"message"`
	} `json:"error"`
}

// SimplifyError takes a verbose error message and makes it more human-friendly
func SimplifyError(err error) string {
	if err == nil {
		return ""
	}

	errMsg := err.Error()

	// Try to extract JSON error message from Weaviate responses
	if strings.Contains(errMsg, `{"error":[`) {
		if cleanMsg := extractWeaviateError(errMsg); cleanMsg != "" {
			return cleanMsg
		}
	}

	// Remove redundant "failed to create collection" prefixes
	errMsg = removeRedundantPrefixes(errMsg)

	// Simplify common error patterns
	errMsg = simplifyCommonErrors(errMsg)

	return errMsg
}

// extractWeaviateError extracts and formats Weaviate API error messages
func extractWeaviateError(errMsg string) string {
	// Find JSON portion
	jsonStart := strings.Index(errMsg, `{"error":[`)
	if jsonStart == -1 {
		return ""
	}
	jsonPart := errMsg[jsonStart:]

	// Parse JSON
	var weaviateErr WeaviateError
	if err := json.Unmarshal([]byte(jsonPart), &weaviateErr); err != nil {
		return ""
	}

	// Extract error messages
	if len(weaviateErr.Error) == 0 {
		return ""
	}

	var messages []string
	for _, e := range weaviateErr.Error {
		if msg := formatWeaviateMessage(e.Message); msg != "" {
			messages = append(messages, msg)
		}
	}

	if len(messages) == 0 {
		return ""
	}

	return strings.Join(messages, "\n")
}

// formatWeaviateMessage formats a single Weaviate error message
func formatWeaviateMessage(msg string) string {
	// Handle OpenAI model errors
	if strings.Contains(msg, "wrong OpenAI model name") {
		return formatOpenAIModelError(msg)
	}

	// Handle module errors
	if strings.Contains(msg, "module '") {
		msg = strings.ReplaceAll(msg, "module '", "")
		msg = strings.ReplaceAll(msg, "':", ":")
	}

	return msg
}

// formatOpenAIModelError formats OpenAI model-related errors
func formatOpenAIModelError(msg string) string {
	// Extract available models
	re := regexp.MustCompile(`available model names are: \[(.*?)\]`)
	matches := re.FindStringSubmatch(msg)

	if len(matches) > 1 {
		models := strings.Split(matches[1], " ")
		var cleanModels []string
		for _, m := range models {
			m = strings.TrimSpace(m)
			if m != "" {
				cleanModels = append(cleanModels, m)
			}
		}

		// Build friendly message
		return fmt.Sprintf("Invalid OpenAI embedding model. Available models:\n%s\n\nUse --embedding-model flag to specify a valid model.", formatList(cleanModels))
	}

	return msg
}

// formatList formats a list of items in a readable way
func formatList(items []string) string {
	var result strings.Builder
	for i, item := range items {
		if i < len(items)-1 {
			result.WriteString(fmt.Sprintf("  • %s\n", item))
		} else {
			result.WriteString(fmt.Sprintf("  • %s", item))
		}
	}
	return result.String()
}

// removeRedundantPrefixes removes redundant error prefixes
func removeRedundantPrefixes(errMsg string) string {
	patterns := []string{
		"failed to create collection",
		"failed to create document",
		"failed to delete collection",
		"failed to delete document",
	}

	for _, pattern := range patterns {
		count := strings.Count(errMsg, pattern)

		if count > 1 {
			// Remove all occurrences except the first
			parts := strings.Split(errMsg, pattern)
			if len(parts) <= 1 {
				continue
			}

			// Keep the first part (before first occurrence) and everything after
			result := parts[0]
			for i := 1; i < len(parts); i++ {
				if i == 1 {
					// After the first pattern, keep it
					result += pattern + parts[i]
				} else {
					// Skip the pattern for subsequent occurrences
					result += parts[i]
				}
			}
			errMsg = result
		}
	}

	return errMsg
}

// simplifyCommonErrors simplifies common error patterns
func simplifyCommonErrors(errMsg string) string {
	// Remove HTTP status codes that are already implied
	errMsg = regexp.MustCompile(`HTTP \d+ - `).ReplaceAllString(errMsg, "")

	// Simplify nested error wrapping
	errMsg = strings.ReplaceAll(errMsg, ": failed to", ":")

	// Simplify Milvus "can't find collection" errors
	if strings.Contains(errMsg, "can't find collection") {
		// Extract collection name from pattern like [collection=WeaveDocs]
		re := regexp.MustCompile(`\[collection=([^\]]+)\]`)
		if matches := re.FindStringSubmatch(errMsg); len(matches) > 1 {
			return fmt.Sprintf("Collection '%s' not found", matches[1])
		}
		return "Collection not found"
	}

	// Clean up excessive colons and spaces
	errMsg = regexp.MustCompile(`: +:`).ReplaceAllString(errMsg, ":")
	errMsg = regexp.MustCompile(`: +`).ReplaceAllString(errMsg, ": ")

	return strings.TrimSpace(errMsg)
}

// FormatCreationError formats collection/document creation errors
func FormatCreationError(resourceType string, err error) string {
	simplified := SimplifyError(err)

	// If error starts with "collection", remove it to avoid redundancy
	if strings.HasPrefix(simplified, "collection") {
		simplified = strings.TrimPrefix(simplified, "collection ")
		simplified = strings.TrimPrefix(simplified, "'")
	}

	return fmt.Sprintf("Failed to create %s: %s", resourceType, simplified)
}

// FormatDeletionError formats collection/document deletion errors
func FormatDeletionError(resourceType string, err error) string {
	simplified := SimplifyError(err)
	return fmt.Sprintf("Failed to delete %s: %s", resourceType, simplified)
}

// FormatQueryError formats query errors
func FormatQueryError(err error) string {
	simplified := SimplifyError(err)

	// Handle common query errors
	if strings.Contains(simplified, "collection") && strings.Contains(simplified, "not found") {
		return simplified
	}

	return fmt.Sprintf("Query failed: %s", simplified)
}
