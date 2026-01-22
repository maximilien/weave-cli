// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package utils

import (
	"errors"
	"strings"
	"testing"
)

func TestSimplifyError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "Nil error",
			err:      nil,
			expected: "",
		},
		{
			name:     "Simple error",
			err:      errors.New("something went wrong"),
			expected: "something went wrong",
		},
		{
			name:     "Redundant prefixes",
			err:      errors.New("failed to create collection: failed to create collection: failed to create collection 'WeaveDocs': HTTP 422"),
			expected: "failed to create collection: 'WeaveDocs'",
		},
		{
			name:     "Weaviate OpenAI model error",
			err:      errors.New(`failed to create collection: failed to create collection 'WeaveDocs': failed to create collection: HTTP 422 - {"error":[{"message":"module 'text2vec-openai': wrong OpenAI model name, available model names are: [ada babbage curie davinci text-embedding-3-small text-embedding-3-large]"}]}`),
			expected: "Invalid OpenAI embedding model. Available models:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SimplifyError(tt.err)

			if tt.err == nil {
				if result != "" {
					t.Errorf("SimplifyError() with nil = %v, want empty string", result)
				}
				return
			}

			if !strings.Contains(result, tt.expected) {
				t.Errorf("SimplifyError() = %v, want to contain %v", result, tt.expected)
			}
		})
	}
}

func TestExtractWeaviateError(t *testing.T) {
	tests := []struct {
		name        string
		errMsg      string
		shouldFind  bool
		contains    []string
		notContains []string
	}{
		{
			name:       "OpenAI model error",
			errMsg:     `failed to create collection: HTTP 422 - {"error":[{"message":"module 'text2vec-openai': wrong OpenAI model name, available model names are: [ada babbage curie davinci text-embedding-3-small text-embedding-3-large]"}]}`,
			shouldFind: true,
			contains: []string{
				"Invalid OpenAI embedding model",
				"ada",
				"text-embedding-3-small",
				"--embedding-model",
			},
			notContains: []string{
				"HTTP 422",
				"module",
			},
		},
		{
			name:       "No JSON error",
			errMsg:     "simple error message",
			shouldFind: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractWeaviateError(tt.errMsg)

			if tt.shouldFind {
				if result == "" {
					t.Error("extractWeaviateError() returned empty string, expected error message")
				}

				for _, expected := range tt.contains {
					if !strings.Contains(result, expected) {
						t.Errorf("extractWeaviateError() = %v, want to contain %v", result, expected)
					}
				}

				for _, notExpected := range tt.notContains {
					if strings.Contains(result, notExpected) {
						t.Errorf("extractWeaviateError() = %v, should not contain %v", result, notExpected)
					}
				}
			} else {
				if result != "" {
					t.Errorf("extractWeaviateError() = %v, want empty string", result)
				}
			}
		})
	}
}

func TestFormatCreationError(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		err          error
		expected     string
	}{
		{
			name:         "Simple error",
			resourceType: "collection",
			err:          errors.New("something failed"),
			expected:     "Failed to create collection: something failed",
		},
		{
			name:         "Weaviate JSON error",
			resourceType: "collection",
			err:          errors.New(`failed to create collection 'Test': HTTP 422 - {"error":[{"message":"module 'text2vec-openai': wrong OpenAI model name, available model names are: [ada babbage]"}]}`),
			expected:     "Failed to create collection: Invalid OpenAI embedding model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatCreationError(tt.resourceType, tt.err)

			if !strings.Contains(result, tt.expected) {
				t.Errorf("FormatCreationError() = %v, want to contain %v", result, tt.expected)
			}
		})
	}
}

func TestFormatDeletionError(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		err          error
		expected     string
	}{
		{
			name:         "Simple deletion error",
			resourceType: "collection",
			err:          errors.New("not found"),
			expected:     "Failed to delete collection: not found",
		},
		{
			name:         "Complex deletion error",
			resourceType: "document",
			err:          errors.New("failed to delete document: failed to delete: connection refused"),
			expected:     "Failed to delete document:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDeletionError(tt.resourceType, tt.err)

			if !strings.Contains(result, tt.expected) {
				t.Errorf("FormatDeletionError() = %v, want to contain %v", result, tt.expected)
			}
		})
	}
}

func TestRemoveRedundantPrefixes(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected string
	}{
		{
			name:     "No redundancy",
			errMsg:   "failed to create collection: something failed",
			expected: "failed to create collection: something failed",
		},
		{
			name:     "Double redundancy",
			errMsg:   "failed to create collection: failed to create collection 'Test': error",
			expected: "failed to create collection:  'Test': error",
		},
		{
			name:     "Triple redundancy",
			errMsg:   "failed to create collection: failed to create collection: failed to create collection 'Test': error",
			expected: "failed to create collection: :  'Test': error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeRedundantPrefixes(tt.errMsg)

			if result != tt.expected {
				t.Errorf("removeRedundantPrefixes() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFormatList(t *testing.T) {
	tests := []struct {
		name     string
		items    []string
		expected []string
	}{
		{
			name:     "Single item",
			items:    []string{"item1"},
			expected: []string{"•", "item1"},
		},
		{
			name:     "Multiple items",
			items:    []string{"item1", "item2", "item3"},
			expected: []string{"•", "item1", "item2", "item3"},
		},
		{
			name:     "Empty list",
			items:    []string{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatList(tt.items)

			for _, expected := range tt.expected {
				if !strings.Contains(result, expected) {
					t.Errorf("formatList() = %v, want to contain %v", result, expected)
				}
			}
		})
	}
}

func TestSimplifyCommonErrors(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected string
	}{
		{
			name:     "HTTP status code removal",
			errMsg:   "HTTP 422 - Invalid request",
			expected: "Invalid request",
		},
		{
			name:     "Nested error wrapping",
			errMsg:   "operation: failed to: failed to connect",
			expected: "operation:: connect",
		},
		{
			name:     "Milvus collection not found with name",
			errMsg:   "can't find collection [collection=WeaveDocs]",
			expected: "Collection 'WeaveDocs' not found",
		},
		{
			name:     "Milvus collection not found generic",
			errMsg:   "can't find collection",
			expected: "Collection not found",
		},
		{
			name:     "Excessive colons with spaces",
			errMsg:   "error: : :  something failed",
			expected: "error: : something failed",
		},
		{
			name:     "Multiple HTTP codes",
			errMsg:   "HTTP 500 - Internal error: HTTP 422 - Bad request",
			expected: "Internal error: Bad request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := simplifyCommonErrors(tt.errMsg)

			if result != tt.expected {
				t.Errorf("simplifyCommonErrors() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFormatWeaviateMessage(t *testing.T) {
	tests := []struct {
		name     string
		msg      string
		expected string
	}{
		{
			name:     "OpenAI model error",
			msg:      "module 'text2vec-openai': wrong OpenAI model name, available model names are: [ada babbage]",
			expected: "Invalid OpenAI embedding model",
		},
		{
			name:     "Module error with quotes",
			msg:      "module 'text2vec-cohere': invalid API key",
			expected: "text2vec-cohere: invalid API key",
		},
		{
			name:     "Regular message",
			msg:      "Collection already exists",
			expected: "Collection already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatWeaviateMessage(tt.msg)

			if !strings.Contains(result, tt.expected) {
				t.Errorf("formatWeaviateMessage() = %q, want to contain %q", result, tt.expected)
			}
		})
	}
}

func TestFormatOpenAIModelError(t *testing.T) {
	tests := []struct {
		name     string
		msg      string
		contains []string
	}{
		{
			name: "Model error with available models",
			msg:  "wrong OpenAI model name, available model names are: [ada babbage curie davinci text-embedding-3-small text-embedding-3-large]",
			contains: []string{
				"Invalid OpenAI embedding model",
				"ada",
				"babbage",
				"text-embedding-3-small",
				"--embedding-model",
			},
		},
		{
			name:     "Model error without models list",
			msg:      "wrong OpenAI model name",
			contains: []string{"wrong OpenAI model name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatOpenAIModelError(tt.msg)

			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("formatOpenAIModelError() = %q, want to contain %q", result, expected)
				}
			}
		})
	}
}

func TestFormatQueryError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "Collection not found",
			err:      errors.New("collection 'WeaveDocs' not found"),
			expected: "collection 'WeaveDocs' not found",
		},
		{
			name:     "Generic query error",
			err:      errors.New("connection timeout"),
			expected: "Query failed: connection timeout",
		},
		{
			name:     "Complex error",
			err:      errors.New("failed to execute query: HTTP 500 - Internal server error"),
			expected: "Query failed:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatQueryError(tt.err)

			if !strings.Contains(result, tt.expected) {
				t.Errorf("FormatQueryError() = %q, want to contain %q", result, tt.expected)
			}
		})
	}
}

// Real-world error scenarios
func TestErrorFormatting_RealWorldScenarios(t *testing.T) {
	t.Run("AuctionsMax.ai Weaviate OpenAI model error", func(t *testing.T) {
		// This was a real error encountered during AuctionsMax.ai development
		err := errors.New(`failed to create collection: HTTP 422 - {"error":[{"message":"module 'text2vec-openai': wrong OpenAI model name, available model names are: [ada babbage curie davinci text-embedding-3-small text-embedding-3-large text-embedding-ada-002]"}]}`)

		result := SimplifyError(err)

		// Should extract and format the OpenAI model error
		if !strings.Contains(result, "Invalid OpenAI embedding model") {
			t.Errorf("Expected 'Invalid OpenAI embedding model', got: %s", result)
		}
		if !strings.Contains(result, "text-embedding-3-small") {
			t.Errorf("Expected available models list, got: %s", result)
		}
		if !strings.Contains(result, "--embedding-model") {
			t.Errorf("Expected usage hint, got: %s", result)
		}

		// Should NOT contain HTTP codes or module quotes
		if strings.Contains(result, "HTTP 422") {
			t.Errorf("Should not contain HTTP status code, got: %s", result)
		}
		if strings.Contains(result, "module '") {
			t.Errorf("Should not contain module quotes, got: %s", result)
		}
	})

	t.Run("Milvus collection not found error", func(t *testing.T) {
		err := errors.New("failed to query collection: can't find collection [collection=WeaveDocs]")

		result := SimplifyError(err)

		expected := "Collection 'WeaveDocs' not found"
		if result != expected {
			t.Errorf("Expected %q, got %q", expected, result)
		}
	})

	t.Run("Nested redundant prefixes", func(t *testing.T) {
		err := errors.New("failed to create collection: failed to create collection: failed to create collection 'TestCol': permission denied")

		result := SimplifyError(err)

		// Should keep only one "failed to create collection"
		count := strings.Count(result, "failed to create collection")
		if count > 1 {
			t.Errorf("Expected at most 1 'failed to create collection', got %d in: %s", count, result)
		}
	})
}
