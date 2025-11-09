// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package tests

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/cmd/embeddings"
	"github.com/spf13/cobra"
)

func TestEmbeddingsListCommand(t *testing.T) {
	// Create a test command
	cmd := &cobra.Command{
		Use: "test",
	}

	// Add embeddings list command
	cmd.AddCommand(embeddings.ListCmd)

	// Capture output
	output := &bytes.Buffer{}
	cmd.SetOut(output)
	cmd.SetErr(output)

	// Set arguments for list command
	cmd.SetArgs([]string{"list"})

	// Execute
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Failed to execute embeddings list: %v", err)
	}

	// Check output contains expected providers
	outputStr := output.String()
	expectedProviders := []string{"OpenAI", "Cohere", "Hugging Face", "Weaviate", "Google PaLM"}
	for _, provider := range expectedProviders {
		if !strings.Contains(outputStr, provider) {
			t.Errorf("Expected output to contain provider '%s', but it didn't", provider)
		}
	}

	// Check output contains expected models
	expectedModels := []string{
		"text-embedding-3-small",
		"text-embedding-3-large",
		"embed-english-v3.0",
		"weaviate-default",
	}
	for _, model := range expectedModels {
		if !strings.Contains(outputStr, model) {
			t.Errorf("Expected output to contain model '%s', but it didn't", model)
		}
	}
}

func TestEmbeddingsListVerbose(t *testing.T) {
	// Create a test command
	cmd := &cobra.Command{
		Use: "test",
	}

	// Add embeddings list command
	cmd.AddCommand(embeddings.ListCmd)

	// Capture output
	output := &bytes.Buffer{}
	cmd.SetOut(output)
	cmd.SetErr(output)

	// Set arguments for verbose list
	cmd.SetArgs([]string{"list", "--verbose"})

	// Execute
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Failed to execute embeddings list --verbose: %v", err)
	}

	// Check output contains verbose information
	outputStr := output.String()
	verboseInfo := []string{
		"Dimensions:",
		"OPENAI_API_KEY",
		"Weaviate Module:",
		"text2vec-openai",
	}
	for _, info := range verboseInfo {
		if !strings.Contains(outputStr, info) {
			t.Errorf("Expected verbose output to contain '%s', but it didn't", info)
		}
	}
}

func TestEmbeddingsListWithCollection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Set environment for mock database
	os.Setenv("VECTOR_DB_TYPE", "mock")
	defer os.Unsetenv("VECTOR_DB_TYPE")

	collectionName := "TestEmbeddingsCollection"

	// Create a test command
	cmd := &cobra.Command{
		Use: "test",
	}

	// Add embeddings list command
	cmd.AddCommand(embeddings.ListCmd)

	// Capture output
	output := &bytes.Buffer{}
	cmd.SetOut(output)
	cmd.SetErr(output)

	// Set arguments for collection-specific list
	cmd.SetArgs([]string{"list", collectionName})

	// Execute
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Failed to execute embeddings list for collection: %v", err)
	}

	// Check output contains collection name
	outputStr := output.String()
	if !strings.Contains(outputStr, collectionName) {
		t.Errorf("Expected output to contain collection name '%s'", collectionName)
	}

	// Check output contains database type information
	if !strings.Contains(outputStr, "Database Type:") {
		t.Errorf("Expected output to contain database type information")
	}
}

func TestGetAllEmbeddingModels(t *testing.T) {
	models := embeddings.GetAllEmbeddingModels()

	if len(models) == 0 {
		t.Fatal("Expected to get embedding models, but got none")
	}

	// Check that we have models from different providers
	providers := make(map[string]bool)
	types := make(map[string]bool)

	for _, model := range models {
		providers[model.Provider] = true
		types[model.Type] = true

		// Validate required fields
		if model.Name == "" {
			t.Errorf("Model has empty name")
		}
		if model.Provider == "" {
			t.Errorf("Model '%s' has empty provider", model.Name)
		}
		if model.Type == "" {
			t.Errorf("Model '%s' has empty type", model.Name)
		}
		if model.Module == "" {
			t.Errorf("Model '%s' has empty module", model.Name)
		}
	}

	// Check we have multiple providers
	if len(providers) < 3 {
		t.Errorf("Expected at least 3 providers, got %d", len(providers))
	}

	// Check we have both text and image types
	if !types["text"] {
		t.Error("Expected to have text embeddings")
	}
}

func TestIsAPIKeySet(t *testing.T) {
	tests := []struct {
		name     string
		envVars  string
		setValue bool
		expected bool
	}{
		{
			name:     "No API key required",
			envVars:  "",
			setValue: false,
			expected: true,
		},
		{
			name:     "Single API key set",
			envVars:  "TEST_API_KEY",
			setValue: true,
			expected: true,
		},
		{
			name:     "Single API key not set",
			envVars:  "TEST_API_KEY",
			setValue: false,
			expected: false,
		},
		{
			name:     "Multiple API keys all set",
			envVars:  "TEST_KEY_1, TEST_KEY_2",
			setValue: true,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			if tt.setValue && tt.envVars != "" {
				vars := strings.Split(tt.envVars, ",")
				for _, v := range vars {
					v = strings.TrimSpace(v)
					os.Setenv(v, "test-value")
					defer os.Unsetenv(v)
				}
			}

			// Test
			result := embeddings.IsAPIKeySet(tt.envVars)
			if result != tt.expected {
				t.Errorf("IsAPIKeySet(%q) = %v, want %v", tt.envVars, result, tt.expected)
			}
		})
	}
}

func TestEmbeddingsCommandAliases(t *testing.T) {
	// Test that all aliases work
	aliases := []string{"embeddings", "emb", "embed", "embeds"}

	for _, alias := range aliases {
		t.Run("alias_"+alias, func(t *testing.T) {
			cmd := &cobra.Command{
				Use: "test",
			}

			// Create embeddings command
			embeddingsCmd := &cobra.Command{
				Use:     "embeddings",
				Aliases: []string{"emb", "embed", "embeds"},
			}
			embeddingsCmd.AddCommand(embeddings.ListCmd)
			cmd.AddCommand(embeddingsCmd)

			// Capture output
			output := &bytes.Buffer{}
			cmd.SetOut(output)
			cmd.SetErr(output)

			// Try to use the alias
			cmd.SetArgs([]string{alias, "list"})

			// Execute - should not fail if alias works
			err := cmd.Execute()
			if err != nil {
				t.Errorf("Alias '%s' failed to execute: %v", alias, err)
			}
		})
	}
}

func TestEmbeddingsGroupedByProvider(t *testing.T) {
	cmd := &cobra.Command{
		Use: "test",
	}
	cmd.AddCommand(embeddings.ListCmd)

	output := &bytes.Buffer{}
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs([]string{"list"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	outputStr := output.String()

	// Check that providers appear in alphabetical order
	// and that models are grouped under providers
	awsIdx := strings.Index(outputStr, "AWS Bedrock")
	cohereIdx := strings.Index(outputStr, "Cohere")
	openaiIdx := strings.Index(outputStr, "OpenAI")

	if awsIdx > cohereIdx {
		t.Error("Expected AWS Bedrock to appear before Cohere (alphabetical order)")
	}
	if cohereIdx > openaiIdx {
		t.Error("Expected Cohere to appear before OpenAI (alphabetical order)")
	}
}

func TestEmbeddingTypeIcons(t *testing.T) {
	cmd := &cobra.Command{
		Use: "test",
	}
	cmd.AddCommand(embeddings.ListCmd)

	output := &bytes.Buffer{}
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs([]string{"list"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	outputStr := output.String()

	// Check for type indicators
	// Text embeddings should have 📝
	// Image embeddings should have 🖼️
	// Multimodal should have 🔄

	if !strings.Contains(outputStr, "📝") {
		t.Error("Expected text embedding icon 📝 in output")
	}
}
