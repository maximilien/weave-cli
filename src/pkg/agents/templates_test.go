// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package agents

import (
	"testing"
)

func TestGetRAGTemplate(t *testing.T) {
	t.Run("Returns valid RAG template", func(t *testing.T) {
		config := GetRAGTemplate()

		// Check type
		if config.Type != "rag" {
			t.Errorf("Expected type 'rag', got '%s'", config.Type)
		}

		// Check required fields
		if config.LLM.Model == "" {
			t.Error("Expected LLM model to be set")
		}

		if config.SystemPrompt == "" {
			t.Error("Expected system prompt to be set")
		}

		// Check RAG-specific settings
		if !config.Response.IncludeReferences {
			t.Error("Expected RAG template to include references")
		}

		if config.Response.CitationFormat != "numeric" {
			t.Errorf("Expected citation format 'numeric', got '%s'", config.Response.CitationFormat)
		}

		// Validate the config
		if err := config.Validate(); err != nil {
			t.Errorf("Template validation failed: %v", err)
		}
	})

	t.Run("Has reasonable defaults", func(t *testing.T) {
		config := GetRAGTemplate()

		if config.LLM.Temperature <= 0 || config.LLM.Temperature > 2.0 {
			t.Errorf("Unexpected temperature: %f", config.LLM.Temperature)
		}

		if config.LLM.MaxTokens <= 0 {
			t.Errorf("Unexpected max tokens: %d", config.LLM.MaxTokens)
		}

		if config.Response.MaxContextChunks <= 0 {
			t.Errorf("Unexpected max context chunks: %d", config.Response.MaxContextChunks)
		}
	})
}

func TestGetQATemplate(t *testing.T) {
	t.Run("Returns valid QA template", func(t *testing.T) {
		config := GetQATemplate()

		// Check type
		if config.Type != "qa" {
			t.Errorf("Expected type 'qa', got '%s'", config.Type)
		}

		// Check required fields
		if config.LLM.Model == "" {
			t.Error("Expected LLM model to be set")
		}

		if config.SystemPrompt == "" {
			t.Error("Expected system prompt to be set")
		}

		// Check QA-specific settings
		if !config.Response.StrictMode {
			t.Error("Expected QA template to have strict mode enabled")
		}

		// Lower temperature for precision
		if config.LLM.Temperature >= 0.7 {
			t.Errorf("Expected lower temperature for QA, got %f", config.LLM.Temperature)
		}

		// Validate the config
		if err := config.Validate(); err != nil {
			t.Errorf("Template validation failed: %v", err)
		}
	})

	t.Run("Has user prompt template", func(t *testing.T) {
		config := GetQATemplate()

		if config.UserPromptTemplate == "" {
			t.Error("Expected QA template to have user prompt template")
		}
	})
}

func TestGetSummarizeTemplate(t *testing.T) {
	t.Run("Returns valid Summarize template", func(t *testing.T) {
		config := GetSummarizeTemplate()

		// Check type
		if config.Type != "summarize" {
			t.Errorf("Expected type 'summarize', got '%s'", config.Type)
		}

		// Check required fields
		if config.LLM.Model == "" {
			t.Error("Expected LLM model to be set")
		}

		if config.SystemPrompt == "" {
			t.Error("Expected system prompt to be set")
		}

		// Check summarize-specific settings
		if config.Response.MaxContextChunks < 5 {
			t.Errorf("Expected higher max context chunks for summarization, got %d", config.Response.MaxContextChunks)
		}

		// Validate the config
		if err := config.Validate(); err != nil {
			t.Errorf("Template validation failed: %v", err)
		}
	})

	t.Run("Has reasonable defaults for summarization", func(t *testing.T) {
		config := GetSummarizeTemplate()

		if config.Output.Format != "markdown" {
			t.Errorf("Expected markdown output format, got '%s'", config.Output.Format)
		}

		// Should have more context for better summaries
		if config.Response.MaxContextChunks <= 5 {
			t.Error("Expected more context chunks for summarization")
		}
	})
}

func TestGetCustomTemplate(t *testing.T) {
	t.Run("Returns valid Custom template", func(t *testing.T) {
		config := GetCustomTemplate()

		// Check type
		if config.Type != "custom" {
			t.Errorf("Expected type 'custom', got '%s'", config.Type)
		}

		// Check required fields
		if config.LLM.Model == "" {
			t.Error("Expected LLM model to be set")
		}

		if config.SystemPrompt == "" {
			t.Error("Expected system prompt to be set")
		}

		// Validate the config
		if err := config.Validate(); err != nil {
			t.Errorf("Template validation failed: %v", err)
		}
	})

	t.Run("Has minimal but complete configuration", func(t *testing.T) {
		config := GetCustomTemplate()

		// Should have basic defaults
		if config.LLM.Temperature == 0 {
			t.Error("Expected temperature to be set")
		}

		if config.LLM.MaxTokens == 0 {
			t.Error("Expected max tokens to be set")
		}

		if config.Response.MaxContextChunks == 0 {
			t.Error("Expected max context chunks to be set")
		}
	})
}

func TestAllTemplatesAreValid(t *testing.T) {
	templates := map[string]*CustomAgentConfig{
		"rag":       GetRAGTemplate(),
		"qa":        GetQATemplate(),
		"summarize": GetSummarizeTemplate(),
		"custom":    GetCustomTemplate(),
	}

	for name, template := range templates {
		t.Run(name, func(t *testing.T) {
			// Set a valid name (required for validation)
			template.Name = "test-agent"

			if err := template.Validate(); err != nil {
				t.Errorf("Template '%s' failed validation: %v", name, err)
			}
		})
	}
}

func TestTemplateImmutability(t *testing.T) {
	t.Run("Templates return new instances", func(t *testing.T) {
		// Get two instances
		config1 := GetRAGTemplate()
		config2 := GetRAGTemplate()

		// They should have the same content but be different objects
		if config1 == config2 {
			t.Error("Expected different instances, got same pointer")
		}

		// Modify one
		config1.Name = "modified"

		// Other should be unchanged
		if config2.Name == "modified" {
			t.Error("Template instances should be independent")
		}
	})
}

func TestTemplateFieldValues(t *testing.T) {
	t.Run("RAG template has appropriate values", func(t *testing.T) {
		config := GetRAGTemplate()

		if config.LLM.Provider != "openai" {
			t.Errorf("Expected provider 'openai', got '%s'", config.LLM.Provider)
		}

		if config.Output.Format != "markdown" {
			t.Errorf("Expected markdown format, got '%s'", config.Output.Format)
		}

		if config.Response.CitationFormat != "numeric" {
			t.Errorf("Expected numeric citation format, got '%s'", config.Response.CitationFormat)
		}
	})

	t.Run("QA template has strict settings", func(t *testing.T) {
		config := GetQATemplate()

		if !config.Response.StrictMode {
			t.Error("Expected strict mode to be enabled for QA")
		}

		if !config.Features.FactChecking {
			t.Error("Expected fact checking to be enabled for QA")
		}

		if config.Performance.CacheResponses != true {
			t.Error("Expected caching to be enabled for QA")
		}
	})
}
