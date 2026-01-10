// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package agents

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

// CustomAgentConfig represents a custom agent loaded from YAML
type CustomAgentConfig struct {
	// Metadata
	Name        string `yaml:"name"`
	Type        string `yaml:"type"` // rag, summarize, qa, custom
	Description string `yaml:"description"`
	Version     string `yaml:"version,omitempty"`
	Author      string `yaml:"author,omitempty"`

	// LLM Configuration
	LLM CustomAgentLLMConfig `yaml:"llm"`

	// Prompts
	SystemPrompt         string `yaml:"system_prompt"`
	UserPromptTemplate   string `yaml:"user_prompt_template,omitempty"`
	AssistantPromptHint  string `yaml:"assistant_prompt_hint,omitempty"`

	// Response Configuration
	Response CustomAgentResponseConfig `yaml:"response,omitempty"`

	// Output Configuration
	Output CustomAgentOutputConfig `yaml:"output,omitempty"`

	// Features
	Features CustomAgentFeaturesConfig `yaml:"features,omitempty"`

	// Performance
	Performance CustomAgentPerformanceConfig `yaml:"performance,omitempty"`
}

// CustomAgentLLMConfig represents LLM settings for a custom agent
type CustomAgentLLMConfig struct {
	Provider          string  `yaml:"provider,omitempty"` // openai, anthropic
	Model             string  `yaml:"model"`
	Temperature       float64 `yaml:"temperature,omitempty"`
	MaxTokens         int     `yaml:"max_tokens,omitempty"`
	TopP              float64 `yaml:"top_p,omitempty"`
	FrequencyPenalty  float64 `yaml:"frequency_penalty,omitempty"`
	PresencePenalty   float64 `yaml:"presence_penalty,omitempty"`
}

// CustomAgentResponseConfig represents response formatting configuration
type CustomAgentResponseConfig struct {
	IncludeReferences   bool    `yaml:"include_references,omitempty"`
	CitationFormat      string  `yaml:"citation_format,omitempty"` // numeric, author-year, footnote
	MaxContextChunks    int     `yaml:"max_context_chunks,omitempty"`
	MinRelevanceScore   float64 `yaml:"min_relevance_score,omitempty"`
	DeduplicateSources  bool    `yaml:"deduplicate_sources,omitempty"`
	SortByRelevance     bool    `yaml:"sort_by_relevance,omitempty"`
	StrictMode          bool    `yaml:"strict_mode,omitempty"`
}

// CustomAgentOutputConfig represents output formatting configuration
type CustomAgentOutputConfig struct {
	Format           string `yaml:"format,omitempty"` // markdown, text, json
	IncludeMetadata  bool   `yaml:"include_metadata,omitempty"`
	ShowConfidence   bool   `yaml:"show_confidence,omitempty"`
	ShowSources      bool   `yaml:"show_sources,omitempty"`
	TruncateSources  int    `yaml:"truncate_sources,omitempty"` // characters
}

// CustomAgentFeaturesConfig represents feature flags
type CustomAgentFeaturesConfig struct {
	Streaming    bool `yaml:"streaming,omitempty"`
	MultiTurn    bool `yaml:"multi_turn,omitempty"`
	FactChecking bool `yaml:"fact_checking,omitempty"`
}

// CustomAgentPerformanceConfig represents performance settings
type CustomAgentPerformanceConfig struct {
	TimeoutSeconds  int  `yaml:"timeout_seconds,omitempty"`
	MaxRetries      int  `yaml:"max_retries,omitempty"`
	CacheResponses  bool `yaml:"cache_responses,omitempty"`
}

// LoadCustomAgentConfig loads a custom agent configuration from a YAML file
func LoadCustomAgentConfig(filepath string) (*CustomAgentConfig, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read agent config file: %w", err)
	}

	var config CustomAgentConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse agent config YAML: %w", err)
	}

	// Apply defaults
	config.applyDefaults()

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid agent configuration: %w", err)
	}

	return &config, nil
}

// applyDefaults applies default values to missing configuration fields
func (c *CustomAgentConfig) applyDefaults() {
	// LLM defaults
	if c.LLM.Provider == "" {
		c.LLM.Provider = "openai"
	}
	if c.LLM.Temperature == 0 {
		c.LLM.Temperature = 0.7
	}
	if c.LLM.MaxTokens == 0 {
		c.LLM.MaxTokens = 2000
	}
	if c.LLM.TopP == 0 {
		c.LLM.TopP = 1.0
	}

	// Response defaults
	if c.Response.MaxContextChunks == 0 {
		c.Response.MaxContextChunks = 5
	}
	if c.Response.MinRelevanceScore == 0 {
		c.Response.MinRelevanceScore = 0.3
	}
	if c.Response.CitationFormat == "" {
		c.Response.CitationFormat = "numeric"
	}

	// Output defaults
	if c.Output.Format == "" {
		c.Output.Format = "markdown"
	}
	if c.Output.TruncateSources == 0 {
		c.Output.TruncateSources = 200
	}

	// Performance defaults
	if c.Performance.TimeoutSeconds == 0 {
		c.Performance.TimeoutSeconds = 30
	}
	if c.Performance.MaxRetries == 0 {
		c.Performance.MaxRetries = 2
	}

	// Version default
	if c.Version == "" {
		c.Version = "1.0.0"
	}
}

// Validate validates the agent configuration
func (c *CustomAgentConfig) Validate() error {
	// Required fields
	if c.Name == "" {
		return fmt.Errorf("agent name is required")
	}
	if c.Type == "" {
		return fmt.Errorf("agent type is required")
	}
	if c.LLM.Model == "" {
		return fmt.Errorf("llm.model is required")
	}
	if c.SystemPrompt == "" {
		return fmt.Errorf("system_prompt is required")
	}

	// Validate name format (lowercase, alphanumeric, hyphens only)
	nameRegex := regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)
	if !nameRegex.MatchString(c.Name) {
		return fmt.Errorf("agent name must be lowercase alphanumeric with hyphens, start and end with alphanumeric")
	}

	// Validate type
	validTypes := map[string]bool{
		"rag":       true,
		"summarize": true,
		"qa":        true,
		"custom":    true,
	}
	if !validTypes[c.Type] {
		return fmt.Errorf("agent type must be one of: rag, summarize, qa, custom")
	}

	// Validate LLM settings
	if c.LLM.Temperature < 0 || c.LLM.Temperature > 2.0 {
		return fmt.Errorf("llm.temperature must be between 0.0 and 2.0")
	}
	if c.LLM.MaxTokens < 1 || c.LLM.MaxTokens > 32000 {
		return fmt.Errorf("llm.max_tokens must be between 1 and 32000")
	}
	if c.LLM.TopP < 0 || c.LLM.TopP > 1.0 {
		return fmt.Errorf("llm.top_p must be between 0.0 and 1.0")
	}

	// Validate citation format
	validCitationFormats := map[string]bool{
		"numeric":     true,
		"author-year": true,
		"footnote":    true,
	}
	if !validCitationFormats[c.Response.CitationFormat] {
		return fmt.Errorf("response.citation_format must be one of: numeric, author-year, footnote")
	}

	// Validate output format
	validOutputFormats := map[string]bool{
		"markdown": true,
		"text":     true,
		"json":     true,
	}
	if !validOutputFormats[c.Output.Format] {
		return fmt.Errorf("output.format must be one of: markdown, text, json")
	}

	// Validate ranges
	if c.Response.MaxContextChunks < 1 || c.Response.MaxContextChunks > 50 {
		return fmt.Errorf("response.max_context_chunks must be between 1 and 50")
	}
	if c.Response.MinRelevanceScore < 0 || c.Response.MinRelevanceScore > 1.0 {
		return fmt.Errorf("response.min_relevance_score must be between 0.0 and 1.0")
	}
	if c.Performance.TimeoutSeconds < 1 || c.Performance.TimeoutSeconds > 300 {
		return fmt.Errorf("performance.timeout_seconds must be between 1 and 300")
	}
	if c.Performance.MaxRetries < 0 || c.Performance.MaxRetries > 10 {
		return fmt.Errorf("performance.max_retries must be between 0 and 10")
	}

	return nil
}

// GetLLMModel returns the LLM model to use
func (c *CustomAgentConfig) GetLLMModel() string {
	return c.LLM.Model
}

// GetLLMTemperature returns the LLM temperature setting
func (c *CustomAgentConfig) GetLLMTemperature() float64 {
	return c.LLM.Temperature
}

// GetLLMMaxTokens returns the max tokens setting
func (c *CustomAgentConfig) GetLLMMaxTokens() int {
	return c.LLM.MaxTokens
}

// GetSystemPrompt returns the system prompt
func (c *CustomAgentConfig) GetSystemPrompt() string {
	return c.SystemPrompt
}

// IsRAGType returns true if this is a RAG-type agent
func (c *CustomAgentConfig) IsRAGType() bool {
	return c.Type == "rag" || c.Response.IncludeReferences
}
