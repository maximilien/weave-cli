// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package agents

import (
	"encoding/json"
	"fmt"
	"os"
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

Examples:
  # List all available agents
  weave agents list

  # Show detailed info about an agent
  weave agents show rag-agent

  # Validate an agent config file
  weave agents validate configs/agents/my-agent.yaml`,
	}

	cmd.AddCommand(NewListCommand())
	cmd.AddCommand(NewShowCommand())
	cmd.AddCommand(NewValidateCommand())

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
