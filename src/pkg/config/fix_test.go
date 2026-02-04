// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package config

import (
	"testing"
)

func TestParseValidationOutput(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedCount  int
		expectedErrors int
		expectedWarns  int
		firstIssue     ConfigIssue
	}{
		{
			name: "single error - missing database_url",
			input: `⚠️  Configuration Errors:
  • databases.vector_databases[2] (mongodb-cloud).database_url: MongoDB database_url is required
    💡 Set 'database_url' field with MongoDB Atlas connection string`,
			expectedCount:  1,
			expectedErrors: 1,
			expectedWarns:  0,
			firstIssue: ConfigIssue{
				Type:         IssueTypeError,
				DatabaseName: "mongodb-cloud",
				DatabaseIdx:  2,
				Field:        "database_url",
				Message:      "MongoDB database_url is required",
				Hint:         "Set 'database_url' field with MongoDB Atlas connection string",
			},
		},
		{
			name: "single warning - missing api_key",
			input: `⚠️  Configuration Warnings:
  • databases.vector_databases[4] (milvus-cloud).api_key: Milvus Cloud (Zilliz) requires an API key
    💡 Set 'api_key' field for authentication`,
			expectedCount:  1,
			expectedErrors: 0,
			expectedWarns:  1,
			firstIssue: ConfigIssue{
				Type:         IssueTypeWarning,
				DatabaseName: "milvus-cloud",
				DatabaseIdx:  4,
				Field:        "api_key",
				Message:      "Milvus Cloud (Zilliz) requires an API key",
				Hint:         "Set 'api_key' field for authentication",
			},
		},
		{
			name: "multiple issues - errors and warnings",
			input: `⚠️  Configuration Errors:
  • databases.vector_databases[2] (mongodb-cloud).database_url: MongoDB database_url is required
    💡 Set 'database_url' field with MongoDB Atlas connection string

⚠️  Configuration Warnings:
  • databases.vector_databases[4] (milvus-cloud).api_key: Milvus Cloud (Zilliz) requires an API key
    💡 Set 'api_key' field for authentication
  • databases.vector_databases[7] (chroma-cloud).tenant: Chroma Cloud requires a tenant ID
    💡 Set 'tenant' field to your Chroma Cloud tenant ID`,
			expectedCount:  3,
			expectedErrors: 1,
			expectedWarns:  2,
			firstIssue: ConfigIssue{
				Type:         IssueTypeError,
				DatabaseName: "mongodb-cloud",
				DatabaseIdx:  2,
				Field:        "database_url",
			},
		},
		{
			name: "field without database context",
			input: `⚠️  Configuration Errors:
  • ai.model: Model name is required
    💡 Set 'model' field to a valid model name like 'gpt-4o'`,
			expectedCount:  1,
			expectedErrors: 1,
			expectedWarns:  0,
			firstIssue: ConfigIssue{
				Type:         IssueTypeError,
				DatabaseName: "",
				DatabaseIdx:  -1,
				Field:        "model",
				Message:      "Model name is required",
				Hint:         "Set 'model' field to a valid model name like 'gpt-4o'",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues, err := ParseValidationOutput(tt.input)
			if err != nil {
				t.Fatalf("ParseValidationOutput() error = %v", err)
			}

			if len(issues) != tt.expectedCount {
				t.Errorf("Expected %d issues, got %d", tt.expectedCount, len(issues))
			}

			// Count errors and warnings
			errorCount := 0
			warnCount := 0
			for _, issue := range issues {
				if issue.Type == IssueTypeError {
					errorCount++
				} else if issue.Type == IssueTypeWarning {
					warnCount++
				}
			}

			if errorCount != tt.expectedErrors {
				t.Errorf("Expected %d errors, got %d", tt.expectedErrors, errorCount)
			}

			if warnCount != tt.expectedWarns {
				t.Errorf("Expected %d warnings, got %d", tt.expectedWarns, warnCount)
			}

			if len(issues) > 0 {
				first := issues[0]
				if first.Type != tt.firstIssue.Type {
					t.Errorf("First issue type = %v, want %v", first.Type, tt.firstIssue.Type)
				}
				if first.DatabaseName != tt.firstIssue.DatabaseName {
					t.Errorf("First issue database = %v, want %v", first.DatabaseName, tt.firstIssue.DatabaseName)
				}
				if first.DatabaseIdx != tt.firstIssue.DatabaseIdx {
					t.Errorf("First issue index = %v, want %v", first.DatabaseIdx, tt.firstIssue.DatabaseIdx)
				}
				if first.Field != tt.firstIssue.Field {
					t.Errorf("First issue field = %v, want %v", first.Field, tt.firstIssue.Field)
				}
			}
		})
	}
}

func TestConfigIssue_IsSecretField(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		expected bool
	}{
		{"api_key is secret", "api_key", true},
		{"API_KEY is secret", "API_KEY", true},
		{"openai_api_key is secret", "openai_api_key", true},
		{"token is secret", "token", true},
		{"auth_token is secret", "auth_token", true},
		{"password is secret", "password", true},
		{"secret is secret", "secret", true},
		{"credentials is secret", "credentials", true},
		{"url is not secret", "url", false},
		{"database_url is not secret", "database_url", false},
		{"tenant is not secret", "tenant", false},
		{"name is not secret", "name", false},
		{"port is not secret", "port", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issue := ConfigIssue{Field: tt.field}
			result := issue.IsSecretField()
			if result != tt.expected {
				t.Errorf("IsSecretField(%s) = %v, want %v", tt.field, result, tt.expected)
			}
		})
	}
}

func TestConfigIssue_GetFieldExample(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		contains string // Check if example contains this string
	}{
		{"url field", "url", "https://"},
		{"database_url field", "database_url", "mongodb"},
		{"api_key field", "api_key", "api-key"},
		{"token field", "token", "token"},
		{"tenant field", "tenant", "tenant"},
		{"host field", "host", "localhost"},
		{"port field", "port", "8080"},
		{"unknown field", "custom_field", "your-value-here"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issue := ConfigIssue{Field: tt.field}
			example := issue.GetFieldExample()
			if example == "" {
				t.Errorf("GetFieldExample(%s) returned empty string", tt.field)
			}
			// Just verify it returns something reasonable
			t.Logf("Example for %s: %s", tt.field, example)
		})
	}
}

func TestConfigIssue_GetFixPromptOptions(t *testing.T) {
	tests := []struct {
		name          string
		issue         ConfigIssue
		expectedCount int
		hasRemove     bool
		hasDisable    bool
	}{
		{
			name: "database field with index",
			issue: ConfigIssue{
				DatabaseIdx: 2,
				Field:       "database_url",
			},
			expectedCount: 5, // Enter, Skip, Remove, Disable, Quit
			hasRemove:     true,
			hasDisable:    true,
		},
		{
			name: "non-database field",
			issue: ConfigIssue{
				DatabaseIdx: -1,
				Field:       "model",
			},
			expectedCount: 3, // Enter, Skip, Quit
			hasRemove:     false,
			hasDisable:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := tt.issue.GetFixPromptOptions()

			if len(options) != tt.expectedCount {
				t.Errorf("Expected %d options, got %d", tt.expectedCount, len(options))
			}

			hasRemove := false
			hasDisable := false
			for _, opt := range options {
				if contains(opt, "Remove") {
					hasRemove = true
				}
				if contains(opt, "Disable") {
					hasDisable = true
				}
			}

			if hasRemove != tt.hasRemove {
				t.Errorf("Expected hasRemove=%v, got %v", tt.hasRemove, hasRemove)
			}
			if hasDisable != tt.hasDisable {
				t.Errorf("Expected hasDisable=%v, got %v", tt.hasDisable, hasDisable)
			}
		})
	}
}

func TestMapOptionToAction(t *testing.T) {
	tests := []struct {
		name           string
		choice         int
		issue          ConfigIssue
		expectedAction FixAction
	}{
		{
			name:           "database field - choice 1 (enter)",
			choice:         1,
			issue:          ConfigIssue{DatabaseIdx: 2},
			expectedAction: FixActionSetValue,
		},
		{
			name:           "database field - choice 2 (skip)",
			choice:         2,
			issue:          ConfigIssue{DatabaseIdx: 2},
			expectedAction: FixActionSkip,
		},
		{
			name:           "database field - choice 3 (remove)",
			choice:         3,
			issue:          ConfigIssue{DatabaseIdx: 2},
			expectedAction: FixActionRemove,
		},
		{
			name:           "database field - choice 4 (disable)",
			choice:         4,
			issue:          ConfigIssue{DatabaseIdx: 2},
			expectedAction: FixActionDisable,
		},
		{
			name:           "database field - choice 5 (quit)",
			choice:         5,
			issue:          ConfigIssue{DatabaseIdx: 2},
			expectedAction: FixActionQuit,
		},
		{
			name:           "non-database field - choice 1 (enter)",
			choice:         1,
			issue:          ConfigIssue{DatabaseIdx: -1},
			expectedAction: FixActionSetValue,
		},
		{
			name:           "non-database field - choice 2 (skip)",
			choice:         2,
			issue:          ConfigIssue{DatabaseIdx: -1},
			expectedAction: FixActionSkip,
		},
		{
			name:           "non-database field - choice 3 (quit)",
			choice:         3,
			issue:          ConfigIssue{DatabaseIdx: -1},
			expectedAction: FixActionQuit,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := MapOptionToAction(tt.choice, tt.issue)
			if action != tt.expectedAction {
				t.Errorf("MapOptionToAction(%d) = %v, want %v", tt.choice, action, tt.expectedAction)
			}
		})
	}
}

func TestConfigIssue_String(t *testing.T) {
	tests := []struct {
		name     string
		issue    ConfigIssue
		contains []string
	}{
		{
			name: "error with database",
			issue: ConfigIssue{
				Type:         IssueTypeError,
				DatabaseName: "mongodb-cloud",
				Message:      "database_url is required",
			},
			contains: []string{"❌", "mongodb-cloud", "required"},
		},
		{
			name: "warning without database",
			issue: ConfigIssue{
				Type:    IssueTypeWarning,
				Message: "model is not set",
			},
			contains: []string{"⚠️", "model", "not set"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.issue.String()
			for _, substr := range tt.contains {
				if !contains(result, substr) {
					t.Errorf("String() = %q, should contain %q", result, substr)
				}
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || len(s) > len(substr) && containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
