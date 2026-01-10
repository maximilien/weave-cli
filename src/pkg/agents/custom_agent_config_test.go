// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package agents

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCustomAgentConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *CustomAgentConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid RAG agent config",
			config: &CustomAgentConfig{
				Name: "test-agent",
				Type: "rag",
				LLM: CustomAgentLLMConfig{
					Model:       "gpt-4o",
					Temperature: 0.7,
					MaxTokens:   2000,
				},
				SystemPrompt: "You are a helpful assistant.",
			},
			wantErr: false,
		},
		{
			name: "Missing name",
			config: &CustomAgentConfig{
				Type: "rag",
				LLM: CustomAgentLLMConfig{
					Model: "gpt-4o",
				},
				SystemPrompt: "Test",
			},
			wantErr: true,
			errMsg:  "agent name is required",
		},
		{
			name: "Missing type",
			config: &CustomAgentConfig{
				Name: "test-agent",
				LLM: CustomAgentLLMConfig{
					Model: "gpt-4o",
				},
				SystemPrompt: "Test",
			},
			wantErr: true,
			errMsg:  "agent type is required",
		},
		{
			name: "Missing model",
			config: &CustomAgentConfig{
				Name:         "test-agent",
				Type:         "rag",
				SystemPrompt: "Test",
			},
			wantErr: true,
			errMsg:  "llm.model is required",
		},
		{
			name: "Missing system prompt",
			config: &CustomAgentConfig{
				Name: "test-agent",
				Type: "rag",
				LLM: CustomAgentLLMConfig{
					Model: "gpt-4o",
				},
			},
			wantErr: true,
			errMsg:  "system_prompt is required",
		},
		{
			name: "Invalid name format - uppercase",
			config: &CustomAgentConfig{
				Name: "TestAgent",
				Type: "rag",
				LLM: CustomAgentLLMConfig{
					Model: "gpt-4o",
				},
				SystemPrompt: "Test",
			},
			wantErr: true,
			errMsg:  "agent name must be lowercase alphanumeric",
		},
		{
			name: "Invalid name format - spaces",
			config: &CustomAgentConfig{
				Name: "test agent",
				Type: "rag",
				LLM: CustomAgentLLMConfig{
					Model: "gpt-4o",
				},
				SystemPrompt: "Test",
			},
			wantErr: true,
			errMsg:  "agent name must be lowercase alphanumeric",
		},
		{
			name: "Invalid type",
			config: &CustomAgentConfig{
				Name: "test-agent",
				Type: "invalid",
				LLM: CustomAgentLLMConfig{
					Model: "gpt-4o",
				},
				SystemPrompt: "Test",
			},
			wantErr: true,
			errMsg:  "agent type must be one of",
		},
		{
			name: "Invalid temperature - too low",
			config: &CustomAgentConfig{
				Name: "test-agent",
				Type: "rag",
				LLM: CustomAgentLLMConfig{
					Model:       "gpt-4o",
					Temperature: -0.1,
				},
				SystemPrompt: "Test",
			},
			wantErr: true,
			errMsg:  "llm.temperature must be between",
		},
		{
			name: "Invalid temperature - too high",
			config: &CustomAgentConfig{
				Name: "test-agent",
				Type: "rag",
				LLM: CustomAgentLLMConfig{
					Model:       "gpt-4o",
					Temperature: 2.1,
				},
				SystemPrompt: "Test",
			},
			wantErr: true,
			errMsg:  "llm.temperature must be between",
		},
		{
			name: "Invalid max tokens - explicitly set too high",
			config: &CustomAgentConfig{
				Name: "test-agent",
				Type: "rag",
				LLM: CustomAgentLLMConfig{
					Model:     "gpt-4o",
					MaxTokens: 35000, // Too high
				},
				SystemPrompt: "Test",
			},
			wantErr: true,
			errMsg:  "llm.max_tokens must be between",
		},
		{
			name: "Invalid citation format",
			config: &CustomAgentConfig{
				Name: "test-agent",
				Type: "rag",
				LLM: CustomAgentLLMConfig{
					Model: "gpt-4o",
				},
				SystemPrompt: "Test",
				Response: CustomAgentResponseConfig{
					CitationFormat: "invalid",
				},
			},
			wantErr: true,
			errMsg:  "response.citation_format must be one of",
		},
		{
			name: "Invalid output format",
			config: &CustomAgentConfig{
				Name: "test-agent",
				Type: "rag",
				LLM: CustomAgentLLMConfig{
					Model: "gpt-4o",
				},
				SystemPrompt: "Test",
				Output: CustomAgentOutputConfig{
					Format: "invalid",
				},
			},
			wantErr: true,
			errMsg:  "output.format must be one of",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Apply defaults first (like LoadCustomAgentConfig does)
			tt.config.applyDefaults()

			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					// Check if error contains expected message
					if len(err.Error()) < len(tt.errMsg) || err.Error()[:len(tt.errMsg)] != tt.errMsg {
						t.Errorf("Validate() error = %v, want error containing %v", err, tt.errMsg)
					}
				}
			}
		})
	}
}

func TestCustomAgentConfig_ApplyDefaults(t *testing.T) {
	config := &CustomAgentConfig{
		Name: "test-agent",
		Type: "rag",
		LLM: CustomAgentLLMConfig{
			Model: "gpt-4o",
		},
		SystemPrompt: "Test",
	}

	config.applyDefaults()

	// Test LLM defaults
	if config.LLM.Provider != "openai" {
		t.Errorf("Expected provider 'openai', got '%s'", config.LLM.Provider)
	}
	if config.LLM.Temperature != 0.7 {
		t.Errorf("Expected temperature 0.7, got %f", config.LLM.Temperature)
	}
	if config.LLM.MaxTokens != 2000 {
		t.Errorf("Expected max_tokens 2000, got %d", config.LLM.MaxTokens)
	}
	if config.LLM.TopP != 1.0 {
		t.Errorf("Expected top_p 1.0, got %f", config.LLM.TopP)
	}

	// Test Response defaults
	if config.Response.MaxContextChunks != 5 {
		t.Errorf("Expected max_context_chunks 5, got %d", config.Response.MaxContextChunks)
	}
	if config.Response.MinRelevanceScore != 0.3 {
		t.Errorf("Expected min_relevance_score 0.3, got %f", config.Response.MinRelevanceScore)
	}
	if config.Response.CitationFormat != "numeric" {
		t.Errorf("Expected citation_format 'numeric', got '%s'", config.Response.CitationFormat)
	}

	// Test Output defaults
	if config.Output.Format != "markdown" {
		t.Errorf("Expected format 'markdown', got '%s'", config.Output.Format)
	}
	if config.Output.TruncateSources != 200 {
		t.Errorf("Expected truncate_sources 200, got %d", config.Output.TruncateSources)
	}

	// Test Performance defaults
	if config.Performance.TimeoutSeconds != 30 {
		t.Errorf("Expected timeout_seconds 30, got %d", config.Performance.TimeoutSeconds)
	}
	if config.Performance.MaxRetries != 2 {
		t.Errorf("Expected max_retries 2, got %d", config.Performance.MaxRetries)
	}

	// Test Version default
	if config.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", config.Version)
	}
}

func TestLoadCustomAgentConfig(t *testing.T) {
	// Create a temporary test config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-agent.yaml")

	validConfig := `---
name: test-agent
type: rag
description: "Test RAG agent"

llm:
  model: "gpt-4o"
  temperature: 0.7
  max_tokens: 2000

system_prompt: |
  You are a helpful assistant.
  Always cite sources.

response:
  include_references: true
  citation_format: "numeric"
  max_context_chunks: 5

output:
  format: "markdown"
  include_metadata: true
`

	if err := os.WriteFile(configPath, []byte(validConfig), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Test loading valid config
	config, err := LoadCustomAgentConfig(configPath)
	if err != nil {
		t.Fatalf("LoadCustomAgentConfig() failed: %v", err)
	}

	// Verify loaded values
	if config.Name != "test-agent" {
		t.Errorf("Expected name 'test-agent', got '%s'", config.Name)
	}
	if config.Type != "rag" {
		t.Errorf("Expected type 'rag', got '%s'", config.Type)
	}
	if config.LLM.Model != "gpt-4o" {
		t.Errorf("Expected model 'gpt-4o', got '%s'", config.LLM.Model)
	}
	if config.LLM.Temperature != 0.7 {
		t.Errorf("Expected temperature 0.7, got %f", config.LLM.Temperature)
	}
	if !config.Response.IncludeReferences {
		t.Error("Expected include_references to be true")
	}
	if config.Output.Format != "markdown" {
		t.Errorf("Expected format 'markdown', got '%s'", config.Output.Format)
	}

	// Test loading non-existent file
	_, err = LoadCustomAgentConfig(filepath.Join(tmpDir, "nonexistent.yaml"))
	if err == nil {
		t.Error("Expected error loading non-existent file, got nil")
	}

	// Test loading invalid YAML
	invalidPath := filepath.Join(tmpDir, "invalid.yaml")
	invalidYAML := "invalid: yaml: ::::"
	if err := os.WriteFile(invalidPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}
	_, err = LoadCustomAgentConfig(invalidPath)
	if err == nil {
		t.Error("Expected error loading invalid YAML, got nil")
	}

	// Test loading config with missing required field
	missingFieldPath := filepath.Join(tmpDir, "missing-field.yaml")
	missingFieldYAML := `---
name: test-agent
type: rag
system_prompt: "Test"
# Missing llm.model
`
	if err := os.WriteFile(missingFieldPath, []byte(missingFieldYAML), 0644); err != nil {
		t.Fatalf("Failed to write missing field config: %v", err)
	}
	_, err = LoadCustomAgentConfig(missingFieldPath)
	if err == nil {
		t.Error("Expected validation error for missing field, got nil")
	}
}

func TestCustomAgentConfig_GetterMethods(t *testing.T) {
	config := &CustomAgentConfig{
		Name: "test-agent",
		Type: "rag",
		LLM: CustomAgentLLMConfig{
			Model:       "gpt-4o",
			Temperature: 0.5,
			MaxTokens:   1500,
		},
		SystemPrompt: "You are a helpful assistant.",
		Response: CustomAgentResponseConfig{
			IncludeReferences: true,
		},
	}

	if model := config.GetLLMModel(); model != "gpt-4o" {
		t.Errorf("GetLLMModel() = %s, want 'gpt-4o'", model)
	}

	if temp := config.GetLLMTemperature(); temp != 0.5 {
		t.Errorf("GetLLMTemperature() = %f, want 0.5", temp)
	}

	if tokens := config.GetLLMMaxTokens(); tokens != 1500 {
		t.Errorf("GetLLMMaxTokens() = %d, want 1500", tokens)
	}

	if prompt := config.GetSystemPrompt(); prompt != "You are a helpful assistant." {
		t.Errorf("GetSystemPrompt() = %s, want 'You are a helpful assistant.'", prompt)
	}

	if !config.IsRAGType() {
		t.Error("IsRAGType() = false, want true")
	}

	// Test non-RAG type
	config.Type = "summarize"
	config.Response.IncludeReferences = false
	if config.IsRAGType() {
		t.Error("IsRAGType() = true, want false for summarize without references")
	}
}
