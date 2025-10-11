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
