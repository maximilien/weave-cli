// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package config

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestCheckRequiredEnvVars(t *testing.T) {
	// Save original env vars
	originalURL := os.Getenv("WEAVIATE_URL")
	originalAPIKey := os.Getenv("WEAVIATE_API_KEY")
	originalOpenAIKey := os.Getenv("OPENAI_API_KEY")

	// Clean up after test
	defer func() {
		os.Setenv("WEAVIATE_URL", originalURL)
		os.Setenv("WEAVIATE_API_KEY", originalAPIKey)
		os.Setenv("OPENAI_API_KEY", originalOpenAIKey)
	}()

	tests := []struct {
		name            string
		envVars         map[string]string
		expectError     bool
		expectedMissing []string
	}{
		{
			name: "All required vars set",
			envVars: map[string]string{
				"WEAVIATE_URL":     "https://test.weaviate.cloud",
				"WEAVIATE_API_KEY": "test-key",
				"OPENAI_API_KEY":   "sk-test",
			},
			expectError: false,
		},
		{
			name:            "All vars missing",
			envVars:         map[string]string{},
			expectError:     true,
			expectedMissing: []string{"WEAVIATE_URL", "WEAVIATE_API_KEY", "OPENAI_API_KEY"},
		},
		{
			name: "Only URL set",
			envVars: map[string]string{
				"WEAVIATE_URL": "https://test.weaviate.cloud",
			},
			expectError:     true,
			expectedMissing: []string{"WEAVIATE_API_KEY", "OPENAI_API_KEY"},
		},
		{
			name: "URL and API key set",
			envVars: map[string]string{
				"WEAVIATE_URL":     "https://test.weaviate.cloud",
				"WEAVIATE_API_KEY": "test-key",
			},
			expectError:     true,
			expectedMissing: []string{"OPENAI_API_KEY"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all env vars
			os.Unsetenv("WEAVIATE_URL")
			os.Unsetenv("WEAVIATE_API_KEY")
			os.Unsetenv("OPENAI_API_KEY")

			// Set test env vars
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Check required env vars
			err := CheckRequiredEnvVars()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}

				var configErr *ConfigError
				if !errors.As(err, &configErr) {
					t.Errorf("Expected *ConfigError but got %T", err)
					return
				}

				// Check missing variables
				if len(configErr.MissingVars) != len(tt.expectedMissing) {
					t.Errorf("Expected %d missing vars but got %d: %v",
						len(tt.expectedMissing), len(configErr.MissingVars), configErr.MissingVars)
					return
				}

				for _, expected := range tt.expectedMissing {
					found := false
					for _, missing := range configErr.MissingVars {
						if missing == expected {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected missing var %s not found in %v", expected, configErr.MissingVars)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestFormatConfigError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectEnhanced bool
		expectedParts  []string
	}{
		{
			name:           "Nil error",
			err:            nil,
			expectEnhanced: false,
		},
		{
			name: "ConfigError",
			err: func() error {
				return &ConfigError{
					Message:     "Missing required configuration",
					MissingVars: []string{"WEAVIATE_URL"},
				}
			}(),
			expectEnhanced: true,
			expectedParts: []string{
				"Configuration Error",
				"Missing environment variables",
				"WEAVIATE_URL",
				"How to fix this",
			},
		},
		{
			name:           "Regular error",
			err:            os.ErrNotExist,
			expectEnhanced: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatted := FormatConfigError(tt.err)

			if tt.err == nil {
				if formatted != "" {
					t.Errorf("Expected empty string for nil error but got: %s", formatted)
				}
				return
			}

			if tt.expectEnhanced {
				for _, part := range tt.expectedParts {
					if !strings.Contains(formatted, part) {
						t.Errorf("Expected formatted error to contain %q but it didn't.\nFormatted error: %s",
							part, formatted)
					}
				}
			} else {
				if formatted != tt.err.Error() {
					t.Errorf("Expected unmodified error message but got: %s", formatted)
				}
			}
		})
	}
}

func TestMaskSecret(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Short secret",
			input:    "abc",
			expected: "***",
		},
		{
			name:     "Medium secret",
			input:    "abcdefgh",
			expected: "***",
		},
		{
			name:     "Long secret",
			input:    "sk-proj-1234567890abcdef",
			expected: "sk-p...cdef",
		},
		{
			name:     "API key",
			input:    "1234567890abcdefghijklmnop",
			expected: "1234...mnop",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskSecret(tt.input)
			if result != tt.expected {
				t.Errorf("maskSecret(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetEnvVarExample(t *testing.T) {
	tests := []struct {
		name     string
		varName  string
		expected string
	}{
		{
			name:     "WEAVIATE_URL",
			varName:  "WEAVIATE_URL",
			expected: "https://your-cluster-id.weaviate.cloud",
		},
		{
			name:     "WEAVIATE_API_KEY",
			varName:  "WEAVIATE_API_KEY",
			expected: "your-weaviate-api-key",
		},
		{
			name:     "OPENAI_API_KEY",
			varName:  "OPENAI_API_KEY",
			expected: "sk-proj-your-openai-api-key",
		},
		{
			name:     "VECTOR_DB_TYPE",
			varName:  "VECTOR_DB_TYPE",
			expected: "weaviate-cloud",
		},
		{
			name:     "Unknown var",
			varName:  "UNKNOWN_VAR",
			expected: "your-value-here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getEnvVarExample(tt.varName)
			if result != tt.expected {
				t.Errorf("getEnvVarExample(%q) = %q, want %q", tt.varName, result, tt.expected)
			}
		})
	}
}

func TestLoadEnvFileSimple(t *testing.T) {
	// Create a temporary .env file
	tmpFile := ".env.test"
	content := `# Test env file
WEAVIATE_URL="https://test.weaviate.cloud"
WEAVIATE_API_KEY=test-key
OPENAI_API_KEY='sk-test'

# Comment line
EMPTY_VAR=
`
	err := os.WriteFile(tmpFile, []byte(content), 0600)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	// Load the file
	values := loadEnvFileSimple(tmpFile)

	// Verify values
	expected := map[string]string{
		"WEAVIATE_URL":     "https://test.weaviate.cloud",
		"WEAVIATE_API_KEY": "test-key",
		"OPENAI_API_KEY":   "sk-test",
		"EMPTY_VAR":        "",
	}

	for key, expectedVal := range expected {
		if val, ok := values[key]; !ok {
			t.Errorf("Expected key %s not found", key)
		} else if val != expectedVal {
			t.Errorf("For key %s: got %q, want %q", key, val, expectedVal)
		}
	}
}

func TestCreateSimpleEnvFile(t *testing.T) {
	tmpFile := ".env.test.create"
	defer os.Remove(tmpFile)

	values := map[string]string{
		"WEAVIATE_URL":     "https://test.weaviate.cloud",
		"WEAVIATE_API_KEY": "test-key",
		"OPENAI_API_KEY":   "sk-test",
	}

	err := createSimpleEnvFile(tmpFile, values)
	if err != nil {
		t.Fatalf("Failed to create env file: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Fatalf("File was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)
	for key, val := range values {
		if !strings.Contains(contentStr, key) {
			t.Errorf("File does not contain key %s", key)
		}
		if !strings.Contains(contentStr, val) {
			t.Errorf("File does not contain value %s", val)
		}
	}
}
