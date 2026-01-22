// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package llm

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestDefaultCompletionOptions(t *testing.T) {
	opts := DefaultCompletionOptions()

	if opts == nil {
		t.Fatal("Expected non-nil options")
	}

	if opts.Model != "gpt-4o" {
		t.Errorf("Expected default model 'gpt-4o', got '%s'", opts.Model)
	}

	if opts.Temperature != 0.1 {
		t.Errorf("Expected default temperature 0.1, got %f", opts.Temperature)
	}

	if opts.MaxTokens != 4096 {
		t.Errorf("Expected default max tokens 4096, got %d", opts.MaxTokens)
	}

	if opts.SystemMsg != "" {
		t.Errorf("Expected empty system message, got '%s'", opts.SystemMsg)
	}
}

func TestWithModel(t *testing.T) {
	opts := DefaultCompletionOptions()
	opt := WithModel("gpt-3.5-turbo")
	opt(opts)

	if opts.Model != "gpt-3.5-turbo" {
		t.Errorf("Expected model 'gpt-3.5-turbo', got '%s'", opts.Model)
	}
}

func TestWithTemperature(t *testing.T) {
	opts := DefaultCompletionOptions()
	opt := WithTemperature(0.8)
	opt(opts)

	if opts.Temperature != 0.8 {
		t.Errorf("Expected temperature 0.8, got %f", opts.Temperature)
	}
}

func TestWithMaxTokens(t *testing.T) {
	opts := DefaultCompletionOptions()
	opt := WithMaxTokens(2048)
	opt(opts)

	if opts.MaxTokens != 2048 {
		t.Errorf("Expected max tokens 2048, got %d", opts.MaxTokens)
	}
}

func TestWithSystemMessage(t *testing.T) {
	opts := DefaultCompletionOptions()
	opt := WithSystemMessage("You are a helpful assistant")
	opt(opts)

	if opts.SystemMsg != "You are a helpful assistant" {
		t.Errorf("Expected system message 'You are a helpful assistant', got '%s'", opts.SystemMsg)
	}
}

func TestMultipleOptions(t *testing.T) {
	opts := DefaultCompletionOptions()

	// Apply multiple options
	WithModel("gpt-4")(opts)
	WithTemperature(0.5)(opts)
	WithMaxTokens(1024)(opts)
	WithSystemMessage("Test system")(opts)

	if opts.Model != "gpt-4" {
		t.Errorf("Expected model 'gpt-4', got '%s'", opts.Model)
	}
	if opts.Temperature != 0.5 {
		t.Errorf("Expected temperature 0.5, got %f", opts.Temperature)
	}
	if opts.MaxTokens != 1024 {
		t.Errorf("Expected max tokens 1024, got %d", opts.MaxTokens)
	}
	if opts.SystemMsg != "Test system" {
		t.Errorf("Expected system message 'Test system', got '%s'", opts.SystemMsg)
	}
}

func TestCleanJSONResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Clean JSON without fences",
			input:    `{"key": "value"}`,
			expected: `{"key": "value"}`,
		},
		{
			name:     "JSON with markdown code fences",
			input:    "```json\n{\"key\": \"value\"}\n```",
			expected: `{"key": "value"}`,
		},
		{
			name:     "JSON with backticks at start",
			input:    "`{\"key\": \"value\"}`",
			expected: `{"key": "value"}`,
		},
		{
			name:     "JSON with nested objects",
			input:    "```\n{\"outer\": {\"inner\": \"value\"}}\n```",
			expected: `{"outer": {"inner": "value"}}`,
		},
		{
			name:     "Empty JSON object",
			input:    "```json\n{}\n```",
			expected: `{}`,
		},
		{
			name:     "JSON array",
			input:    "```json\n{\"items\": [1, 2, 3]}\n```",
			expected: `{"items": [1, 2, 3]}`,
		},
		{
			name:     "Plain text without JSON",
			input:    "plain text",
			expected: "plain text",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanJSONResponse(tt.input)
			if result != tt.expected {
				t.Errorf("cleanJSONResponse(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCalculateCost(t *testing.T) {
	tests := []struct {
		name             string
		model            string
		promptTokens     int
		completionTokens int
		expectedCost     float64
	}{
		{
			name:             "GPT-4o default pricing",
			model:            "gpt-4o",
			promptTokens:     1000,
			completionTokens: 500,
			expectedCost:     (1000.0/1000000)*2.50 + (500.0/1000000)*10.0,
		},
		{
			name:             "GPT-4o-2024-11-20",
			model:            "gpt-4o-2024-11-20",
			promptTokens:     1000,
			completionTokens: 500,
			expectedCost:     (1000.0/1000000)*2.50 + (500.0/1000000)*10.0,
		},
		{
			name:             "GPT-4 Turbo",
			model:            "gpt-4-turbo",
			promptTokens:     1000,
			completionTokens: 500,
			expectedCost:     (1000.0/1000000)*10.0 + (500.0/1000000)*30.0,
		},
		{
			name:             "GPT-4",
			model:            "gpt-4",
			promptTokens:     1000,
			completionTokens: 500,
			expectedCost:     (1000.0/1000000)*30.0 + (500.0/1000000)*60.0,
		},
		{
			name:             "GPT-3.5 Turbo",
			model:            "gpt-3.5-turbo",
			promptTokens:     1000,
			completionTokens: 500,
			expectedCost:     (1000.0/1000000)*0.5 + (500.0/1000000)*1.5,
		},
		{
			name:             "Unknown model defaults to GPT-4o",
			model:            "unknown-model",
			promptTokens:     1000,
			completionTokens: 500,
			expectedCost:     (1000.0/1000000)*2.50 + (500.0/1000000)*10.0,
		},
		{
			name:             "Zero tokens",
			model:            "gpt-4o",
			promptTokens:     0,
			completionTokens: 0,
			expectedCost:     0.0,
		},
		{
			name:             "Large token count",
			model:            "gpt-4o",
			promptTokens:     1000000,
			completionTokens: 500000,
			expectedCost:     (1000000.0/1000000)*2.50 + (500000.0/1000000)*10.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := calculateCost(tt.model, tt.promptTokens, tt.completionTokens)
			if cost != tt.expectedCost {
				t.Errorf("calculateCost(%q, %d, %d) = %f, expected %f",
					tt.model, tt.promptTokens, tt.completionTokens, cost, tt.expectedCost)
			}
		})
	}
}

func TestNewOpenAIClient(t *testing.T) {
	tests := []struct {
		name        string
		apiKey      string
		shouldError bool
	}{
		{
			name:        "Valid API key",
			apiKey:      "test-api-key",
			shouldError: false,
		},
		{
			name:        "Empty API key",
			apiKey:      "",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewOpenAIClient(tt.apiKey)

			if tt.shouldError {
				if err == nil {
					t.Error("Expected error for empty API key, got nil")
				}
				if client != nil {
					t.Error("Expected nil client for error case")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if client == nil {
				t.Error("Expected non-nil client")
				return
			}

			// Verify metrics are initialized
			metrics := client.GetMetrics()
			if metrics == nil {
				t.Error("Expected non-nil metrics")
				return
			}

			if metrics.Invocations != 0 {
				t.Errorf("Expected 0 invocations, got %d", metrics.Invocations)
			}
			if metrics.TotalTokens != 0 {
				t.Errorf("Expected 0 total tokens, got %d", metrics.TotalTokens)
			}
			if metrics.PromptTokens != 0 {
				t.Errorf("Expected 0 prompt tokens, got %d", metrics.PromptTokens)
			}
			if metrics.CompletionTokens != 0 {
				t.Errorf("Expected 0 completion tokens, got %d", metrics.CompletionTokens)
			}
			if metrics.TotalCost != 0.0 {
				t.Errorf("Expected 0.0 total cost, got %f", metrics.TotalCost)
			}
		})
	}
}

func TestNewOpenAIClientWithHTTP(t *testing.T) {
	tests := []struct {
		name        string
		apiKey      string
		httpClient  *http.Client
		shouldError bool
	}{
		{
			name:        "Valid API key with custom HTTP client",
			apiKey:      "test-api-key",
			httpClient:  &http.Client{},
			shouldError: false,
		},
		{
			name:        "Empty API key with custom HTTP client",
			apiKey:      "",
			httpClient:  &http.Client{},
			shouldError: true,
		},
		{
			name:        "Valid API key with nil HTTP client",
			apiKey:      "test-api-key",
			httpClient:  nil,
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewOpenAIClientWithHTTP(tt.apiKey, tt.httpClient)

			if tt.shouldError {
				if err == nil {
					t.Error("Expected error for empty API key, got nil")
				}
				if client != nil {
					t.Error("Expected nil client for error case")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if client == nil {
				t.Error("Expected non-nil client")
				return
			}

			// Verify metrics are initialized
			metrics := client.GetMetrics()
			if metrics == nil {
				t.Error("Expected non-nil metrics")
			}
		})
	}
}

func TestMetrics(t *testing.T) {
	client, err := NewOpenAIClient("test-api-key")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Get initial metrics
	metrics := client.GetMetrics()
	if metrics == nil {
		t.Fatal("Expected non-nil metrics")
	}

	// Verify initial state
	if metrics.Invocations != 0 {
		t.Errorf("Expected 0 invocations, got %d", metrics.Invocations)
	}
	if metrics.TotalTokens != 0 {
		t.Errorf("Expected 0 total tokens, got %d", metrics.TotalTokens)
	}
	if metrics.PromptTokens != 0 {
		t.Errorf("Expected 0 prompt tokens, got %d", metrics.PromptTokens)
	}
	if metrics.CompletionTokens != 0 {
		t.Errorf("Expected 0 completion tokens, got %d", metrics.CompletionTokens)
	}
	if metrics.TotalCost != 0.0 {
		t.Errorf("Expected 0.0 total cost, got %f", metrics.TotalCost)
	}
}

func TestLoadOpikConfig(t *testing.T) {
	// Save original env vars
	originalEnabled := os.Getenv("OPIK_ENABLED")
	originalAPIKey := os.Getenv("OPIK_API_KEY")
	originalEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	originalWorkspace := os.Getenv("OPIK_WORKSPACE")
	originalProject := os.Getenv("OPIK_PROJECT_NAME")

	defer func() {
		// Restore original env vars
		os.Setenv("OPIK_ENABLED", originalEnabled)
		os.Setenv("OPIK_API_KEY", originalAPIKey)
		os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", originalEndpoint)
		os.Setenv("OPIK_WORKSPACE", originalWorkspace)
		os.Setenv("OPIK_PROJECT_NAME", originalProject)
	}()

	tests := []struct {
		name              string
		envEnabled        string
		envAPIKey         string
		envEndpoint       string
		envWorkspace      string
		envProject        string
		expectedEnabled   bool
		expectedAPIKey    string
		expectedEndpoint  string
		expectedWorkspace string
		expectedProject   string
	}{
		{
			name:              "All defaults (no env vars)",
			envEnabled:        "",
			envAPIKey:         "",
			envEndpoint:       "",
			envWorkspace:      "",
			envProject:        "",
			expectedEnabled:   false,
			expectedAPIKey:    "",
			expectedEndpoint:  "https://www.comet.com/opik/api/v1/private/otel/v1/traces",
			expectedWorkspace: "default",
			expectedProject:   "weave-cli-queries",
		},
		{
			name:              "API key enables Opik",
			envEnabled:        "",
			envAPIKey:         "test-api-key",
			envEndpoint:       "",
			envWorkspace:      "",
			envProject:        "",
			expectedEnabled:   true,
			expectedAPIKey:    "test-api-key",
			expectedEndpoint:  "https://www.comet.com/opik/api/v1/private/otel/v1/traces",
			expectedWorkspace: "default",
			expectedProject:   "weave-cli-queries",
		},
		{
			name:              "Custom configuration",
			envEnabled:        "true",
			envAPIKey:         "custom-key",
			envEndpoint:       "https://custom.endpoint.com/traces",
			envWorkspace:      "custom-workspace",
			envProject:        "custom-project",
			expectedEnabled:   true,
			expectedAPIKey:    "custom-key",
			expectedEndpoint:  "https://custom.endpoint.com/traces",
			expectedWorkspace: "custom-workspace",
			expectedProject:   "custom-project",
		},
		{
			name:              "Explicitly disabled with API key",
			envEnabled:        "false",
			envAPIKey:         "test-key",
			envEndpoint:       "",
			envWorkspace:      "",
			envProject:        "",
			expectedEnabled:   true, // API key overrides disabled flag
			expectedAPIKey:    "test-key",
			expectedEndpoint:  "https://www.comet.com/opik/api/v1/private/otel/v1/traces",
			expectedWorkspace: "default",
			expectedProject:   "weave-cli-queries",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all env vars first
			os.Unsetenv("OPIK_ENABLED")
			os.Unsetenv("OPIK_API_KEY")
			os.Unsetenv("OTEL_EXPORTER_OTLP_ENDPOINT")
			os.Unsetenv("OPIK_WORKSPACE")
			os.Unsetenv("OPIK_PROJECT_NAME")

			// Set test env vars
			if tt.envEnabled != "" {
				os.Setenv("OPIK_ENABLED", tt.envEnabled)
			}
			if tt.envAPIKey != "" {
				os.Setenv("OPIK_API_KEY", tt.envAPIKey)
			}
			if tt.envEndpoint != "" {
				os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", tt.envEndpoint)
			}
			if tt.envWorkspace != "" {
				os.Setenv("OPIK_WORKSPACE", tt.envWorkspace)
			}
			if tt.envProject != "" {
				os.Setenv("OPIK_PROJECT_NAME", tt.envProject)
			}

			config := LoadOpikConfig()

			if config.Enabled != tt.expectedEnabled {
				t.Errorf("Expected Enabled=%v, got %v", tt.expectedEnabled, config.Enabled)
			}
			if config.APIKey != tt.expectedAPIKey {
				t.Errorf("Expected APIKey=%q, got %q", tt.expectedAPIKey, config.APIKey)
			}
			if config.Endpoint != tt.expectedEndpoint {
				t.Errorf("Expected Endpoint=%q, got %q", tt.expectedEndpoint, config.Endpoint)
			}
			if config.Workspace != tt.expectedWorkspace {
				t.Errorf("Expected Workspace=%q, got %q", tt.expectedWorkspace, config.Workspace)
			}
			if config.ProjectName != tt.expectedProject {
				t.Errorf("Expected ProjectName=%q, got %q", tt.expectedProject, config.ProjectName)
			}
		})
	}
}

func TestWrapHTTPClient(t *testing.T) {
	tests := []struct {
		name   string
		client *http.Client
	}{
		{
			name:   "Wrap custom HTTP client",
			client: &http.Client{},
		},
		{
			name:   "Wrap nil client (use default)",
			client: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapped := WrapHTTPClient(tt.client)

			if wrapped == nil {
				t.Error("Expected non-nil wrapped client")
				return
			}

			// Verify transport is wrapped
			if wrapped.Transport == nil {
				t.Error("Expected non-nil transport")
			}
		})
	}
}

func TestInitOpikTracing_Disabled(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name   string
		config *OpikConfig
	}{
		{
			name: "Disabled Opik",
			config: &OpikConfig{
				Enabled: false,
				APIKey:  "",
			},
		},
		{
			name: "No API key",
			config: &OpikConfig{
				Enabled: true,
				APIKey:  "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp, err := InitOpikTracing(ctx, tt.config)

			if err != nil {
				t.Errorf("Expected no error for disabled Opik, got: %v", err)
			}

			if tp != nil {
				t.Error("Expected nil tracer provider for disabled Opik")
			}
		})
	}
}

func TestInitOpikTracing_InvalidURL(t *testing.T) {
	ctx := context.Background()

	config := &OpikConfig{
		Enabled:  true,
		APIKey:   "test-key",
		Endpoint: "://invalid-url",
	}

	_, err := InitOpikTracing(ctx, config)

	if err == nil {
		t.Error("Expected error for invalid URL")
	}

	if err != nil && !strings.Contains(err.Error(), "invalid OTEL endpoint URL") {
		t.Errorf("Expected 'invalid OTEL endpoint URL' error, got: %v", err)
	}
}

func TestShutdownOpikTracing(t *testing.T) {
	ctx := context.Background()

	// Test with nil tracer provider
	err := ShutdownOpikTracing(ctx, nil)
	if err != nil {
		t.Errorf("Expected no error for nil tracer provider, got: %v", err)
	}
}
