// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	configpkg "github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/ollama"
	"github.com/spf13/cobra"
)

// agentsCmd represents the config agents command
var agentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "Create AI agents configuration file",
	Long: `Create weave-agents.yaml configuration file for AI features.

This command helps you set up AI agent configuration for:
- Schema suggestions (weave schema suggest)
- Chunking recommendations (weave chunking suggest)
- LLM model selection and parameters
- Temperature and token limits
- Performance tuning

The command will:
1. Check if weave-agents.yaml already exists
2. Copy from configs/weave-agents.yaml template
3. Allow you to choose location (current directory or global ~/.weave-cli)

Examples:
  # Create weave-agents.yaml in current directory
  weave config agents

  # Create in global ~/.weave-cli directory
  weave config agents --global

  # Preview the template without creating
  weave config agents --show-template`,
	Run: runConfigAgents,
}

var (
	agentsGlobal       bool
	agentsShowTemplate bool
)

func init() {
	agentsCmd.Flags().BoolVar(&agentsGlobal, "global", false, "Create weave-agents.yaml in ~/.weave-cli directory (default: current directory)")
	agentsCmd.Flags().BoolVar(&agentsShowTemplate, "show-template", false, "Display the template without creating file")
}

func runConfigAgents(cmd *cobra.Command, args []string) {
	// If --show-template, just display the template and exit
	if agentsShowTemplate {
		showAgentsTemplate()
		return
	}

	// Determine target directory
	var targetDir string
	var locationDesc string
	if agentsGlobal {
		globalDir, err := configpkg.GetGlobalConfigDir()
		if err != nil {
			printError(fmt.Sprintf("Failed to get global config directory: %v", err))
			os.Exit(1)
		}

		// Ensure global directory exists
		created, err := configpkg.EnsureGlobalConfigDir()
		if err != nil {
			printError(fmt.Sprintf("Failed to create global config directory: %v", err))
			os.Exit(1)
		}

		if created {
			color.New(color.FgGreen).Printf("✓ Created global config directory: %s\n", globalDir)
			fmt.Println()
		}

		targetDir = globalDir
		locationDesc = fmt.Sprintf("global directory (%s)", globalDir)
	} else {
		targetDir = "."
		locationDesc = "current directory"
	}

	// Display welcome message
	printHeader("🤖 Weave AI Agents Configuration")
	fmt.Println()
	color.New(color.FgCyan).Println("This tool will create a weave-agents.yaml configuration file.")
	color.New(color.FgCyan).Printf("Target location: %s\n", locationDesc)
	fmt.Println()

	if err := createAgentsFileInDir(targetDir); err != nil {
		printError(fmt.Sprintf("Failed to create weave-agents.yaml: %v", err))
		os.Exit(1)
	}

	// Discover Ollama models
	discoverOllamaModels()

	// Success message
	fmt.Println()
	color.New(color.FgGreen, color.Bold).Println("✅ AI agents configuration created!")
	fmt.Println()
	color.New(color.FgCyan).Println("📚 Configuration details:")
	fmt.Println("  • LLM Provider: OpenAI (gpt-4o)")
	fmt.Println("  • Schema Agent: Enabled (max 50 samples)")
	fmt.Println("  • Chunking Agent: Enabled (1000 default chunk size)")
	fmt.Println("  • Query Agent: Enabled")
	fmt.Println("  • Planning Agent: Enabled")
	fmt.Println()
	color.New(color.FgCyan).Println("💡 Next steps:")
	fmt.Println("  • Edit the file to customize LLM models and parameters")
	fmt.Println("  • Set OPENAI_API_KEY in your .env file")
	fmt.Println("  • Run 'weave schema suggest ./docs --collection MyDocs'")
	fmt.Println("  • Run 'weave chunking suggest ./docs --collection MyDocs'")
}

// createAgentsFileInDir creates weave-agents.yaml in the specified directory
func createAgentsFileInDir(targetDir string) error {
	agentsFile := filepath.Join(targetDir, "weave-agents.yaml")

	// Try to find template in multiple locations
	templatePaths := []string{
		"configs/weave-agents.yaml",                                         // From project root
		"../configs/weave-agents.yaml",                                      // From bin directory
		filepath.Join(os.Getenv("HOME"), ".weave-cli", "weave-agents.yaml"), // Global template
	}

	var templateFile string
	for _, path := range templatePaths {
		if fileExists(path) {
			templateFile = path
			break
		}
	}

	if templateFile == "" {
		// If no template found, use embedded default
		return createDefaultAgentsFile(agentsFile)
	}

	// Check if weave-agents.yaml already exists
	if fileExists(agentsFile) {
		color.New(color.FgYellow, color.Bold).Printf("⚠️  WARNING: %s already exists!\n", agentsFile)
		fmt.Println()
		color.New(color.FgYellow).Print("Do you want to overwrite it? (y/N): ")

		response, err := readLine()
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		if response != "y" && response != "yes" {
			color.New(color.FgBlue).Printf("→ Skipping %s creation\n", agentsFile)
			return nil
		}
		fmt.Println()
	}

	color.New(color.FgGreen).Printf("✓ Found template: %s\n", templateFile)
	fmt.Println()

	// Show info about the file
	color.New(color.FgCyan).Println("📝 About weave-agents.yaml:")
	fmt.Println("  • Configure AI models (GPT-4, Claude, etc.)")
	fmt.Println("  • Customize temperatures and token limits")
	fmt.Println("  • Tune chunking defaults (size, overlap)")
	fmt.Println("  • Enable/disable specific agents")
	fmt.Println("  • Adjust confidence thresholds")
	fmt.Println()

	// Copy template
	color.New(color.FgCyan).Printf("Creating %s from template...\n", agentsFile)
	if err := copyFile(templateFile, agentsFile); err != nil {
		return fmt.Errorf("failed to copy template: %w", err)
	}

	color.New(color.FgGreen, color.Bold).Printf("✅ Successfully created %s\n", agentsFile)

	return nil
}

// showAgentsTemplate displays the template content
func showAgentsTemplate() {
	templateFile := "configs/weave-agents.yaml"

	printHeader("🤖 Weave AI Agents Configuration Template")
	fmt.Println()

	if !fileExists(templateFile) {
		printError(fmt.Sprintf("Template not found: %s", templateFile))
		os.Exit(1)
	}

	content, err := os.ReadFile(templateFile)
	if err != nil {
		printError(fmt.Sprintf("Failed to read template: %v", err))
		os.Exit(1)
	}

	color.New(color.FgCyan).Printf("Template location: %s\n", templateFile)
	fmt.Println()
	fmt.Println(string(content))
	fmt.Println()
	color.New(color.FgCyan).Println("💡 To create this file:")
	fmt.Println("  weave config agents              # Current directory")
	fmt.Println("  weave config agents --global     # Global ~/.weave-cli")
}

// createDefaultAgentsFile creates weave-agents.yaml with embedded default config
func createDefaultAgentsFile(agentsFile string) error {
	color.New(color.FgYellow).Println("⚠️  No template found, using embedded defaults")
	fmt.Println()

	// Check if file already exists
	if fileExists(agentsFile) {
		color.New(color.FgYellow, color.Bold).Printf("⚠️  WARNING: %s already exists!\n", agentsFile)
		fmt.Println()
		color.New(color.FgYellow).Print("Do you want to overwrite it? (y/N): ")

		response, err := readLine()
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		if response != "y" && response != "yes" {
			color.New(color.FgBlue).Printf("→ Skipping %s creation\n", agentsFile)
			return nil
		}
		fmt.Println()
	}

	defaultConfig := `---
# Weave AI Agents Configuration
# This file configures the AI agents used by Weave CLI for schema suggestions,
# chunking recommendations, and other AI-powered features.

# LLM Configuration
llm:
  provider: openai
  api_key_env: "OPENAI_API_KEY"
  default_model: "gpt-4o"
  default_temperature: 0.3
  default_max_tokens: 2000

# Schema Agent Configuration
schema_agent:
  enabled: true
  model: "gpt-4o"
  temperature: 0.3
  max_tokens: 2000
  max_samples: 50
  default_vector_dimensions: 1536
  min_confidence: 0.5
  warn_below_confidence: 0.7

# Chunking Agent Configuration
chunking_agent:
  enabled: true
  model: "gpt-4o"
  temperature: 0.3
  max_tokens: 1000
  default_chunk_size: 1000
  min_chunk_size: 200
  max_chunk_size: 4000
  default_overlap_percent: 15
  min_overlap_percent: 5
  max_overlap_percent: 30
  technical_chunk_size: 800
  narrative_chunk_size: 1500
  mixed_chunk_size: 1000

# Query Agent Configuration (for REPL mode)
query_agent:
  enabled: true
  model: "gpt-4o"
  temperature: 0.1
  max_tokens: 4096

# Planning Agent Configuration (for REPL mode)
planning_agent:
  enabled: true
  model: "gpt-4o"
  temperature: 0.1
  max_tokens: 4096

# Report Agent Configuration
report_agent:
  enabled: true
  temperature: 0.3
  max_tokens: 500

# Eval Agent Configuration
eval_agent:
  enabled: true
  temperature: 0.0
  max_tokens: 10

# Output Settings
output:
  default_format: "text"
  verbosity: "normal"
  show_confidence: true
  show_reasoning: true
  show_warnings: true

# Cache Settings
cache:
  enabled: false
  ttl_seconds: 3600
  max_size_mb: 100

# Performance Settings
performance:
  llm_timeout: 30
  max_concurrent_files: 10
  max_retries: 2
  retry_delay_ms: 1000

# Feature Flags
features:
  experimental: false
  debug: false
  telemetry: false
`

	// Write the file
	if err := os.WriteFile(agentsFile, []byte(defaultConfig), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	color.New(color.FgGreen, color.Bold).Printf("✅ Successfully created %s\n", agentsFile)
	return nil
}

// discoverOllamaModels discovers available Ollama embedding models
func discoverOllamaModels() {
	fmt.Println()
	color.New(color.FgCyan).Println("🔍 Discovering local Ollama models...")

	client := ollama.NewClient("")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Check if Ollama is available
	if !client.IsAvailable(ctx) {
		color.New(color.FgYellow).Println("⚠️  Ollama not detected (server not running at http://localhost:11434)")
		fmt.Println()
		color.New(color.FgCyan).Println("💡 To use local Ollama embeddings:")
		fmt.Println("  1. Install Ollama: https://ollama.ai")
		fmt.Println("  2. Start Ollama server")
		fmt.Println("  3. Pull embedding model: ollama pull nomic-embed-text")
		return
	}

	// Discover embedding models
	embeddings, err := client.DiscoverEmbeddingModels(ctx)
	if err != nil {
		color.New(color.FgYellow).Printf("⚠️  Failed to discover Ollama models: %v\n", err)
		return
	}

	if len(embeddings) == 0 {
		color.New(color.FgYellow).Println("⚠️  No embedding models found")
		fmt.Println()
		color.New(color.FgCyan).Println("💡 To add embedding models:")
		fmt.Println("  ollama pull nomic-embed-text    # 768 dimensions, fast")
		fmt.Println("  ollama pull mxbai-embed-large   # 1024 dimensions, accurate")
		return
	}

	// Display discovered models
	color.New(color.FgGreen).Printf("✓ Found %d local embedding model(s):\n", len(embeddings))
	for _, em := range embeddings {
		fmt.Printf("  • %s (%dd)\n", em.Name, em.Dimensions)
	}

	fmt.Println()
	color.New(color.FgCyan).Println("💡 To use Ollama embeddings in re-embedding:")
	if len(embeddings) > 0 {
		fmt.Printf("  weave collection reembed SOURCE --new-embedding %s --output TARGET\n", embeddings[0].Name)
	}
}
