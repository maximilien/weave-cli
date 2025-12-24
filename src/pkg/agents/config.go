// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package agents

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// AgentConfig represents the configuration for AI agents
type AgentConfig struct {
	LLM           LLMConfig           `yaml:"llm"`
	SchemaAgent   SchemaAgentConfig   `yaml:"schema_agent"`
	ChunkingAgent ChunkingAgentConfig `yaml:"chunking_agent"`
	QueryAgent    QueryAgentConfig    `yaml:"query_agent"`
	PlanningAgent PlanningAgentConfig `yaml:"planning_agent"`
	ReportAgent   ReportAgentConfig   `yaml:"report_agent"`
	EvalAgent     EvalAgentConfig     `yaml:"eval_agent"`
	Output        AgentOutputConfig   `yaml:"output"`
	Cache         CacheConfig         `yaml:"cache"`
	Performance   PerformanceConfig   `yaml:"performance"`
	Features      FeaturesConfig      `yaml:"features"`
}

// LLMConfig represents LLM provider configuration
type LLMConfig struct {
	Provider           string  `yaml:"provider"`
	APIKeyEnv          string  `yaml:"api_key_env"`
	DefaultModel       string  `yaml:"default_model"`
	DefaultTemperature float64 `yaml:"default_temperature"`
	DefaultMaxTokens   int     `yaml:"default_max_tokens"`
}

// SchemaAgentConfig represents schema agent configuration
type SchemaAgentConfig struct {
	Enabled                 bool    `yaml:"enabled"`
	Model                   string  `yaml:"model"`
	Temperature             float64 `yaml:"temperature"`
	MaxTokens               int     `yaml:"max_tokens"`
	MaxSamples              int     `yaml:"max_samples"`
	DefaultVectorDimensions int     `yaml:"default_vector_dimensions"`
	MinConfidence           float64 `yaml:"min_confidence"`
	WarnBelowConfidence     float64 `yaml:"warn_below_confidence"`
}

// ChunkingAgentConfig represents chunking agent configuration
type ChunkingAgentConfig struct {
	Enabled               bool    `yaml:"enabled"`
	Model                 string  `yaml:"model"`
	Temperature           float64 `yaml:"temperature"`
	MaxTokens             int     `yaml:"max_tokens"`
	DefaultChunkSize      int     `yaml:"default_chunk_size"`
	MinChunkSize          int     `yaml:"min_chunk_size"`
	MaxChunkSize          int     `yaml:"max_chunk_size"`
	DefaultOverlapPercent int     `yaml:"default_overlap_percent"`
	MinOverlapPercent     int     `yaml:"min_overlap_percent"`
	MaxOverlapPercent     int     `yaml:"max_overlap_percent"`
	TechnicalChunkSize    int     `yaml:"technical_chunk_size"`
	NarrativeChunkSize    int     `yaml:"narrative_chunk_size"`
	MixedChunkSize        int     `yaml:"mixed_chunk_size"`
}

// QueryAgentConfig represents query agent configuration
type QueryAgentConfig struct {
	Enabled     bool    `yaml:"enabled"`
	Model       string  `yaml:"model"`
	Temperature float64 `yaml:"temperature"`
	MaxTokens   int     `yaml:"max_tokens"`
}

// PlanningAgentConfig represents planning agent configuration
type PlanningAgentConfig struct {
	Enabled     bool    `yaml:"enabled"`
	Model       string  `yaml:"model"`
	Temperature float64 `yaml:"temperature"`
	MaxTokens   int     `yaml:"max_tokens"`
}

// ReportAgentConfig represents report agent configuration
type ReportAgentConfig struct {
	Enabled     bool    `yaml:"enabled"`
	Temperature float64 `yaml:"temperature"`
	MaxTokens   int     `yaml:"max_tokens"`
}

// EvalAgentConfig represents eval agent configuration
type EvalAgentConfig struct {
	Enabled     bool    `yaml:"enabled"`
	Temperature float64 `yaml:"temperature"`
	MaxTokens   int     `yaml:"max_tokens"`
}

// AgentOutputConfig represents output settings for AI agents
type AgentOutputConfig struct {
	DefaultFormat  string `yaml:"default_format"`
	Verbosity      string `yaml:"verbosity"`
	ShowConfidence bool   `yaml:"show_confidence"`
	ShowReasoning  bool   `yaml:"show_reasoning"`
	ShowWarnings   bool   `yaml:"show_warnings"`
}

// CacheConfig represents cache settings
type CacheConfig struct {
	Enabled    bool `yaml:"enabled"`
	TTLSeconds int  `yaml:"ttl_seconds"`
	MaxSizeMB  int  `yaml:"max_size_mb"`
}

// PerformanceConfig represents performance settings
type PerformanceConfig struct {
	LLMTimeout         int `yaml:"llm_timeout"`
	MaxConcurrentFiles int `yaml:"max_concurrent_files"`
	MaxRetries         int `yaml:"max_retries"`
	RetryDelayMS       int `yaml:"retry_delay_ms"`
}

// FeaturesConfig represents feature flags
type FeaturesConfig struct {
	Experimental bool `yaml:"experimental"`
	Debug        bool `yaml:"debug"`
	Telemetry    bool `yaml:"telemetry"`
}

// defaultConfig returns the default agent configuration
func defaultConfig() *AgentConfig {
	return &AgentConfig{
		LLM: LLMConfig{
			Provider:           "openai",
			APIKeyEnv:          "OPENAI_API_KEY",
			DefaultModel:       "gpt-4o",
			DefaultTemperature: 0.3,
			DefaultMaxTokens:   2000,
		},
		SchemaAgent: SchemaAgentConfig{
			Enabled:                 true,
			Model:                   "gpt-4o",
			Temperature:             0.3,
			MaxTokens:               2000,
			MaxSamples:              50,
			DefaultVectorDimensions: 1536,
			MinConfidence:           0.5,
			WarnBelowConfidence:     0.7,
		},
		ChunkingAgent: ChunkingAgentConfig{
			Enabled:               true,
			Model:                 "gpt-4o",
			Temperature:           0.3,
			MaxTokens:             1000,
			DefaultChunkSize:      1000,
			MinChunkSize:          200,
			MaxChunkSize:          4000,
			DefaultOverlapPercent: 15,
			MinOverlapPercent:     5,
			MaxOverlapPercent:     30,
			TechnicalChunkSize:    800,
			NarrativeChunkSize:    1500,
			MixedChunkSize:        1000,
		},
		QueryAgent: QueryAgentConfig{
			Enabled:     true,
			Model:       "gpt-4o",
			Temperature: 0.1,
			MaxTokens:   4096,
		},
		PlanningAgent: PlanningAgentConfig{
			Enabled:     true,
			Model:       "gpt-4o",
			Temperature: 0.1,
			MaxTokens:   4096,
		},
		ReportAgent: ReportAgentConfig{
			Enabled:     true,
			Temperature: 0.3,
			MaxTokens:   500,
		},
		EvalAgent: EvalAgentConfig{
			Enabled:     true,
			Temperature: 0.0,
			MaxTokens:   10,
		},
		Output: AgentOutputConfig{
			DefaultFormat:  "text",
			Verbosity:      "normal",
			ShowConfidence: true,
			ShowReasoning:  true,
			ShowWarnings:   true,
		},
		Cache: CacheConfig{
			Enabled:    false,
			TTLSeconds: 3600,
			MaxSizeMB:  100,
		},
		Performance: PerformanceConfig{
			LLMTimeout:         30,
			MaxConcurrentFiles: 10,
			MaxRetries:         2,
			RetryDelayMS:       1000,
		},
		Features: FeaturesConfig{
			Experimental: false,
			Debug:        false,
			Telemetry:    false,
		},
	}
}

// LoadAgentConfig loads agent configuration from file or returns defaults
func LoadAgentConfig() (*AgentConfig, error) {
	// Try to load from current directory first
	configPath := "weave-agents.yaml"
	if _, err := os.Stat(configPath); err != nil {
		// Try configs directory
		configPath = "configs/weave-agents.yaml"
		if _, err := os.Stat(configPath); err != nil {
			// Try home directory
			home, err := os.UserHomeDir()
			if err == nil {
				configPath = filepath.Join(home, ".weave-cli", "weave-agents.yaml")
				if _, err := os.Stat(configPath); err != nil {
					// No config file found, use defaults
					return defaultConfig(), nil
				}
			} else {
				// Can't determine home directory, use defaults
				return defaultConfig(), nil
			}
		}
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read agent config: %w", err)
	}

	// Parse YAML
	config := defaultConfig()
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse agent config: %w", err)
	}

	return config, nil
}

// GetModel returns the model to use for an agent, falling back to defaults
func (c *AgentConfig) GetModel(agentType string) string {
	switch agentType {
	case "schema":
		if c.SchemaAgent.Model != "" {
			return c.SchemaAgent.Model
		}
	case "chunking":
		if c.ChunkingAgent.Model != "" {
			return c.ChunkingAgent.Model
		}
	case "query":
		if c.QueryAgent.Model != "" {
			return c.QueryAgent.Model
		}
	case "planning":
		if c.PlanningAgent.Model != "" {
			return c.PlanningAgent.Model
		}
	}
	return c.LLM.DefaultModel
}

// GetTemperature returns the temperature to use for an agent
func (c *AgentConfig) GetTemperature(agentType string) float64 {
	switch agentType {
	case "schema":
		return c.SchemaAgent.Temperature
	case "chunking":
		return c.ChunkingAgent.Temperature
	case "query":
		return c.QueryAgent.Temperature
	case "planning":
		return c.PlanningAgent.Temperature
	case "report":
		return c.ReportAgent.Temperature
	case "eval":
		return c.EvalAgent.Temperature
	}
	return c.LLM.DefaultTemperature
}

// GetMaxTokens returns the max tokens to use for an agent
func (c *AgentConfig) GetMaxTokens(agentType string) int {
	switch agentType {
	case "schema":
		return c.SchemaAgent.MaxTokens
	case "chunking":
		return c.ChunkingAgent.MaxTokens
	case "query":
		return c.QueryAgent.MaxTokens
	case "planning":
		return c.PlanningAgent.MaxTokens
	case "report":
		return c.ReportAgent.MaxTokens
	case "eval":
		return c.EvalAgent.MaxTokens
	}
	return c.LLM.DefaultMaxTokens
}
