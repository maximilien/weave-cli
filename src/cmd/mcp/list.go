// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/maximilien/weave-cli/src/pkg/mcp"
	"github.com/spf13/cobra"
)

var (
	listOutputFormat string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available MCP tools",
	Long: `List all available tools from an MCP server.

Examples:
  # List tools from HTTP server
  weave mcp list --server http://localhost:8030

  # List tools from stdio server
  weave mcp list --server ./bin/weave-mcp-stdio --transport stdio

  # Output as JSON
  weave mcp list --server localhost:8030 --output json`,
	RunE: runMCPList,
}

func init() {
	listCmd.Flags().StringVar(&listOutputFormat, "output", "table", "Output format (table, json)")
}

func runMCPList(cmd *cobra.Command, args []string) error {
	if serverURL == "" {
		return fmt.Errorf("--server is required")
	}

	// Create MCP client config
	config := &mcp.Config{
		Name:      "cli-server",
		ServerURL: serverURL,
		Transport: mcp.Transport(transport),
		Timeout:   time.Duration(timeout) * time.Second,
	}

	// Create client
	client, err := mcp.NewMCPClient(config)
	if err != nil {
		return fmt.Errorf("failed to create MCP client: %w", err)
	}
	defer client.Close()

	// Connect
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to MCP server: %w", err)
	}

	// List tools
	tools, err := client.ListTools(ctx)
	if err != nil {
		return fmt.Errorf("failed to list tools: %w", err)
	}

	// Display results
	if listOutputFormat == "json" {
		data, err := json.MarshalIndent(tools, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
	} else {
		displayToolsTable(tools)
	}

	return nil
}

func displayToolsTable(tools []mcp.Tool) {
	blue := color.New(color.FgBlue, color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	fmt.Println(blue("\nAvailable MCP Tools:"))
	fmt.Println(blue("==================="))

	if len(tools) == 0 {
		fmt.Println("No tools available")
		return
	}

	for i, tool := range tools {
		fmt.Printf("\n%s %s\n", green(fmt.Sprintf("%d.", i+1)), cyan(tool.Name))
		if tool.Description != "" {
			fmt.Printf("   %s\n", tool.Description)
		}

		// Show input schema if available
		if tool.InputSchema != nil {
			if props, ok := tool.InputSchema["properties"].(map[string]interface{}); ok {
				fmt.Printf("   %s\n", "Parameters:")
				for name, propData := range props {
					if propMap, ok := propData.(map[string]interface{}); ok {
						propType := "any"
						if t, ok := propMap["type"].(string); ok {
							propType = t
						}
						desc := ""
						if d, ok := propMap["description"].(string); ok {
							desc = " - " + d
						}
						fmt.Printf("     • %s (%s)%s\n", name, propType, desc)
					}
				}
			}
		}
	}

	fmt.Println()
}
