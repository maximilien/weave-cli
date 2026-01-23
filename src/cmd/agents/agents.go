// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package agents

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/maximilien/weave-cli/src/pkg/agents"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// NewAgentsCommand creates the agents command
func NewAgentsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "agents",
		Aliases: []string{"agent"},
		Short:   "Manage and inspect RAG agents",
		Long: `Manage and inspect RAG (Retrieval-Augmented Generation) agents.

Agents process query results to provide comprehensive answers, summaries, or
precise question-answering with citations. Use agents with the --agent flag
in collection query commands.

Available Commands:
  list      List all available agents
  show      Show detailed information about an agent
  validate  Validate an agent configuration file
  create    Create a new agent configuration from template
  delete    Delete an agent configuration
  edit      Edit an agent configuration
  copy      Copy an agent configuration

Examples:
  # List all available agents
  weave agents list

  # Show detailed info about an agent
  weave agents show rag-agent

  # Validate an agent config file
  weave agents validate configs/agents/my-agent.yaml

  # Create a new agent
  weave agents create my-agent --type rag

  # Delete an agent
  weave agents delete my-agent

  # Edit an agent
  weave agents edit my-agent

  # Copy an agent
  weave agents copy rag-agent my-custom-agent`,
	}

	cmd.AddCommand(NewListCommand())
	cmd.AddCommand(NewShowCommand())
	cmd.AddCommand(NewValidateCommand())
	cmd.AddCommand(NewCreateCommand())
	cmd.AddCommand(NewDeleteCommand())
	cmd.AddCommand(NewEditCommand())
	cmd.AddCommand(NewCopyCommand())

	return cmd
}

// NewListCommand creates the list subcommand
func NewListCommand() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all available agents",
		Long: `List all available agents from all search paths.

Agents are loaded from:
  1. configs/agents/ (current directory)
  2. ~/.weave-cli/agents/ (user home)
  3. /etc/weave-cli/agents/ (system-wide)

Examples:
  weave agents list
  weave agents ls
  weave agents list --output json`,
		Run: func(cmd *cobra.Command, args []string) {
			runListAgents(outputFormat)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "Output format: text, json, yaml")

	return cmd
}

// NewShowCommand creates the show subcommand
func NewShowCommand() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "show AGENT_NAME",
		Short: "Show detailed information about an agent",
		Long: `Show detailed information about a specific agent including configuration,
capabilities, and metadata.

Examples:
  weave agents show rag-agent
  weave agents show qa-agent --output json`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runShowAgent(args[0], outputFormat)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "Output format: text, json, yaml")

	return cmd
}

// NewValidateCommand creates the validate subcommand
func NewValidateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate AGENT_FILE",
		Short: "Validate an agent configuration file",
		Long: `Validate an agent configuration file to ensure it's correctly formatted
and contains all required fields.

Examples:
  weave agents validate configs/agents/my-agent.yaml
  weave agents validate ~/.weave-cli/agents/custom-agent.yaml`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runValidateAgent(args[0])
		},
	}

	return cmd
}

// runListAgents lists all available agents
func runListAgents(outputFormat string) {
	registry := agents.GetDefaultAgentRegistry()

	agentList, err := registry.ListAgents()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing agents: %v\n", err)
		os.Exit(1)
	}

	if len(agentList) == 0 {
		fmt.Println("No agents found in search paths.")
		fmt.Println("\nSearch paths:")
		loader := agents.GetDefaultAgentLoader()
		for _, path := range loader.GetSearchPaths() {
			fmt.Printf("  - %s\n", path)
		}
		return
	}

	switch outputFormat {
	case "json":
		data, _ := json.MarshalIndent(agentList, "", "  ")
		fmt.Println(string(data))
	case "yaml":
		data, _ := yaml.Marshal(agentList)
		fmt.Println(string(data))
	default:
		// Text format with table
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tTYPE\tVERSION\tDESCRIPTION")
		fmt.Fprintln(w, "----\t----\t-------\t-----------")
		for _, agent := range agentList {
			desc := agent.Description
			if len(desc) > 60 {
				desc = desc[:57] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", agent.Name, agent.Type, agent.Version, desc)
		}
		w.Flush()

		fmt.Printf("\nTotal: %d agents\n", len(agentList))
	}
}

// runShowAgent shows detailed info about a specific agent
func runShowAgent(agentName string, outputFormat string) {
	registry := agents.GetDefaultAgentRegistry()

	agentInfo, err := registry.GetAgentInfo(agentName)
	if err != nil {
		if agents.IsAgentNotFoundError(err) {
			fmt.Fprintf(os.Stderr, "Agent '%s' not found.\n", agentName)
			fmt.Fprintf(os.Stderr, "Run 'weave agents list' to see available agents.\n")
		} else {
			fmt.Fprintf(os.Stderr, "Error loading agent: %v\n", err)
		}
		os.Exit(1)
	}

	// Load full config
	config, err := agents.LoadAgent(agentName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading agent config: %v\n", err)
		os.Exit(1)
	}

	switch outputFormat {
	case "json":
		data, _ := json.MarshalIndent(config, "", "  ")
		fmt.Println(string(data))
	case "yaml":
		data, _ := yaml.Marshal(config)
		fmt.Println(string(data))
	default:
		// Text format with detailed info
		fmt.Printf("Agent: %s\n", config.Name)
		fmt.Printf("Type: %s\n", config.Type)
		fmt.Printf("Version: %s\n", config.Version)
		if config.Author != "" {
			fmt.Printf("Author: %s\n", config.Author)
		}
		fmt.Printf("Description: %s\n", config.Description)
		fmt.Printf("File Path: %s\n", agentInfo.FilePath)

		fmt.Println("\nLLM Configuration:")
		fmt.Printf("  Provider: %s\n", config.LLM.Provider)
		fmt.Printf("  Model: %s\n", config.LLM.Model)
		fmt.Printf("  Temperature: %.2f\n", config.LLM.Temperature)
		fmt.Printf("  Max Tokens: %d\n", config.LLM.MaxTokens)

		fmt.Println("\nResponse Configuration:")
		fmt.Printf("  Include References: %v\n", config.Response.IncludeReferences)
		fmt.Printf("  Citation Format: %s\n", config.Response.CitationFormat)
		fmt.Printf("  Max Context Chunks: %d\n", config.Response.MaxContextChunks)
		fmt.Printf("  Min Relevance Score: %.2f\n", config.Response.MinRelevanceScore)
		fmt.Printf("  Strict Mode: %v\n", config.Response.StrictMode)

		fmt.Println("\nOutput Configuration:")
		fmt.Printf("  Format: %s\n", config.Output.Format)
		fmt.Printf("  Show Sources: %v\n", config.Output.ShowSources)
		fmt.Printf("  Include Metadata: %v\n", config.Output.IncludeMetadata)
		fmt.Printf("  Show Confidence: %v\n", config.Output.ShowConfidence)

		fmt.Println("\nSystem Prompt:")
		fmt.Printf("%s\n", config.SystemPrompt)
	}
}

// runValidateAgent validates an agent configuration file
func runValidateAgent(filePath string) {
	config, err := agents.LoadCustomAgentConfig(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Validation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Agent configuration is valid!")
	fmt.Printf("\nAgent Details:\n")
	fmt.Printf("  Name: %s\n", config.Name)
	fmt.Printf("  Type: %s\n", config.Type)
	fmt.Printf("  Version: %s\n", config.Version)
	fmt.Printf("  Model: %s\n", config.LLM.Model)
}

// NewCreateCommand creates the create subcommand
func NewCreateCommand() *cobra.Command {
	var agentType string
	var interactive bool
	var outputDir string

	cmd := &cobra.Command{
		Use:   "create NAME",
		Short: "Create a new agent configuration",
		Long: `Create a new agent configuration from template.

Available agent types:
  rag       - General-purpose RAG agent with citations
  qa        - Precise question-answering agent
  summarize - Document summarization agent
  custom    - Start with minimal template

Examples:
  weave agents create my-medical-agent --type rag
  weave agents create legal-qa --type qa --interactive
  weave agents create custom-agent --type custom --output ~/.weave-cli/agents`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runCreateAgent(args[0], agentType, interactive, outputDir)
		},
	}

	cmd.Flags().StringVar(&agentType, "type", "rag", "Agent type: rag, qa, summarize, custom")
	cmd.Flags().BoolVar(&interactive, "interactive", false, "Interactive wizard for configuration")
	cmd.Flags().StringVar(&outputDir, "output", "configs/agents", "Output directory")

	return cmd
}

// NewDeleteCommand creates the delete subcommand
func NewDeleteCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete an agent configuration",
		Long: `Delete an agent configuration file.

Examples:
  weave agents delete my-old-agent
  weave agents delete test-agent --force`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runDeleteAgent(args[0], force)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")

	return cmd
}

// NewEditCommand creates the edit subcommand
func NewEditCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit NAME",
		Short: "Edit an agent configuration",
		Long: `Edit an agent configuration in your default editor.

The editor is determined by the EDITOR environment variable, defaulting to vim.

After editing, the configuration is automatically validated.

Examples:
  weave agents edit rag-agent
  weave agents edit my-custom-agent`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runEditAgent(args[0])
		},
	}

	return cmd
}

// NewCopyCommand creates the copy subcommand
func NewCopyCommand() *cobra.Command {
	var outputDir string

	cmd := &cobra.Command{
		Use:   "copy SOURCE TARGET",
		Short: "Copy an agent configuration",
		Long: `Copy an existing agent configuration to a new name.

This is useful for creating variations of agents for A/B testing or experimentation.

Examples:
  weave agents copy rag-agent my-rag-variant
  weave agents copy qa-agent experimental-qa --output ~/.weave-cli/agents`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			runCopyAgent(args[0], args[1], outputDir)
		},
	}

	cmd.Flags().StringVar(&outputDir, "output", "configs/agents", "Output directory")

	return cmd
}

// runCreateAgent creates a new agent from template
func runCreateAgent(name, agentType string, interactive bool, outputDir string) {
	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	outputPath := filepath.Join(outputDir, name+".yaml")

	// Check if file already exists
	if _, err := os.Stat(outputPath); err == nil {
		fmt.Fprintf(os.Stderr, "Error: Agent '%s' already exists at %s\n", name, outputPath)
		fmt.Fprintf(os.Stderr, "Use 'weave agents edit %s' to modify it or choose a different name.\n", name)
		os.Exit(1)
	}

	var templateName string
	switch agentType {
	case "rag":
		templateName = "rag-agent"
	case "qa":
		templateName = "qa-agent"
	case "summarize":
		templateName = "summarize-agent"
	case "custom":
		templateName = ""
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown agent type '%s'\n", agentType)
		fmt.Fprintf(os.Stderr, "Valid types: rag, qa, summarize, custom\n")
		os.Exit(1)
	}

	var config *agents.CustomAgentConfig
	var err error

	if templateName != "" {
		// Load template from existing agent
		config, err = agents.LoadAgent(templateName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading template '%s': %v\n", templateName, err)
			os.Exit(1)
		}

		// Update name and metadata
		config.Name = name
		config.Version = "1.0.0"
		config.Author = ""
	} else {
		// Create minimal custom template
		config = &agents.CustomAgentConfig{
			Name:        name,
			Type:        "custom",
			Version:     "1.0.0",
			Description: "Custom agent configuration",
			LLM: agents.CustomAgentLLMConfig{
				Provider:    "openai",
				Model:       "gpt-4o",
				Temperature: 0.7,
				MaxTokens:   2000,
				TopP:        1.0,
			},
			SystemPrompt: "You are a helpful AI assistant.",
			Response: agents.CustomAgentResponseConfig{
				IncludeReferences:  true,
				CitationFormat:     "numeric",
				MaxContextChunks:   5,
				MinRelevanceScore:  0.3,
				DeduplicateSources: true,
				SortByRelevance:    true,
				StrictMode:         false,
			},
			Output: agents.CustomAgentOutputConfig{
				Format:          "markdown",
				IncludeMetadata: true,
				ShowConfidence:  false,
				ShowSources:     true,
				TruncateSources: 500,
			},
		}
	}

	// Interactive mode
	if interactive {
		config = promptForConfig(config)
	}

	// Write config to file
	data, err := yaml.Marshal(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling config: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing config file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Created agent '%s' at %s\n", name, outputPath)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  1. Edit the agent: weave agents edit %s\n", name)
	fmt.Printf("  2. Validate: weave agents validate %s\n", outputPath)
	fmt.Printf("  3. Use it: weave collections query <collection> --agent %s\n", name)
}

// runDeleteAgent deletes an agent
func runDeleteAgent(name string, force bool) {
	registry := agents.GetDefaultAgentRegistry()

	agentInfo, err := registry.GetAgentInfo(name)
	if err != nil {
		if agents.IsAgentNotFoundError(err) {
			fmt.Fprintf(os.Stderr, "Error: Agent '%s' not found.\n", name)
			fmt.Fprintf(os.Stderr, "Run 'weave agents list' to see available agents.\n")
		} else {
			fmt.Fprintf(os.Stderr, "Error loading agent: %v\n", err)
		}
		os.Exit(1)
	}

	// Check if it's a built-in agent (in configs/agents/)
	if strings.Contains(agentInfo.FilePath, "configs/agents/") {
		builtInAgents := []string{"rag-agent", "qa-agent", "summarize-agent"}
		for _, builtIn := range builtInAgents {
			if name == builtIn {
				fmt.Fprintf(os.Stderr, "Error: Cannot delete built-in agent '%s'\n", name)
				fmt.Fprintf(os.Stderr, "Built-in agents: %v\n", builtInAgents)
				os.Exit(1)
			}
		}
	}

	// Confirm deletion
	if !force {
		fmt.Printf("Delete agent '%s' at %s? (y/N): ", name, agentInfo.FilePath)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "y" && response != "yes" {
			fmt.Println("Deletion cancelled.")
			return
		}
	}

	// Delete the file
	if err := os.Remove(agentInfo.FilePath); err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting agent: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Deleted agent '%s'\n", name)
}

// runEditAgent opens an agent config in the editor
func runEditAgent(name string) {
	registry := agents.GetDefaultAgentRegistry()

	agentInfo, err := registry.GetAgentInfo(name)
	if err != nil {
		if agents.IsAgentNotFoundError(err) {
			fmt.Fprintf(os.Stderr, "Error: Agent '%s' not found.\n", name)
			fmt.Fprintf(os.Stderr, "Run 'weave agents list' to see available agents.\n")
		} else {
			fmt.Fprintf(os.Stderr, "Error loading agent: %v\n", err)
		}
		os.Exit(1)
	}

	// Get editor from environment
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	// Open editor
	cmd := exec.Command(editor, agentInfo.FilePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running editor: %v\n", err)
		os.Exit(1)
	}

	// Validate after editing
	fmt.Println("\nValidating edited configuration...")
	config, err := agents.LoadCustomAgentConfig(agentInfo.FilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Validation failed: %v\n", err)
		fmt.Fprintf(os.Stderr, "Please fix the errors and try again.\n")
		os.Exit(1)
	}

	fmt.Println("✅ Agent configuration is valid!")
	fmt.Printf("  Name: %s\n", config.Name)
	fmt.Printf("  Type: %s\n", config.Type)
	fmt.Printf("  Model: %s\n", config.LLM.Model)
}

// runCopyAgent copies an agent configuration
func runCopyAgent(sourceName, targetName string, outputDir string) {
	// Load source agent
	sourceConfig, err := agents.LoadAgent(sourceName)
	if err != nil {
		if agents.IsAgentNotFoundError(err) {
			fmt.Fprintf(os.Stderr, "Error: Agent '%s' not found.\n", sourceName)
			fmt.Fprintf(os.Stderr, "Run 'weave agents list' to see available agents.\n")
		} else {
			fmt.Fprintf(os.Stderr, "Error loading agent: %v\n", err)
		}
		os.Exit(1)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	outputPath := filepath.Join(outputDir, targetName+".yaml")

	// Check if target already exists
	if _, err := os.Stat(outputPath); err == nil {
		fmt.Fprintf(os.Stderr, "Error: Agent '%s' already exists at %s\n", targetName, outputPath)
		os.Exit(1)
	}

	// Update config with new name
	sourceConfig.Name = targetName
	sourceConfig.Version = "1.0.0"
	sourceConfig.Author = ""

	// Write to new file
	data, err := yaml.Marshal(sourceConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling config: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing config file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Copied agent '%s' to '%s' at %s\n", sourceName, targetName, outputPath)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  1. Edit the agent: weave agents edit %s\n", targetName)
	fmt.Printf("  2. Customize it for your use case\n")
}

// promptForConfig interactively prompts for config values
func promptForConfig(config *agents.CustomAgentConfig) *agents.CustomAgentConfig {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n🔧 Interactive Agent Configuration")
	fmt.Println("Press Enter to keep default values shown in [brackets]\n")

	// Description
	fmt.Printf("Description [%s]: ", config.Description)
	if desc, _ := reader.ReadString('\n'); strings.TrimSpace(desc) != "" {
		config.Description = strings.TrimSpace(desc)
	}

	// Model
	fmt.Printf("LLM Model [%s]: ", config.LLM.Model)
	if model, _ := reader.ReadString('\n'); strings.TrimSpace(model) != "" {
		config.LLM.Model = strings.TrimSpace(model)
	}

	// Temperature
	fmt.Printf("Temperature [%.2f]: ", config.LLM.Temperature)
	if temp, _ := reader.ReadString('\n'); strings.TrimSpace(temp) != "" {
		var tempVal float64
		if _, err := fmt.Sscanf(strings.TrimSpace(temp), "%f", &tempVal); err == nil {
			config.LLM.Temperature = tempVal
		}
	}

	// Max tokens
	fmt.Printf("Max Tokens [%d]: ", config.LLM.MaxTokens)
	if tokens, _ := reader.ReadString('\n'); strings.TrimSpace(tokens) != "" {
		var tokensVal int
		if _, err := fmt.Sscanf(strings.TrimSpace(tokens), "%d", &tokensVal); err == nil {
			config.LLM.MaxTokens = tokensVal
		}
	}

	fmt.Println("\n✅ Configuration complete!")
	return config
}
