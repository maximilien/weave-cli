// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/maximilien/weave-cli/src/pkg/mcp"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test MCP server connection",
	Long: `Test connection to an MCP server and verify it's working.

This command connects to the server, performs a health check,
and lists available tools to verify functionality.

Examples:
  # Test HTTP server
  weave mcp test --server http://localhost:8030

  # Test stdio server
  weave mcp test --server ./bin/weave-mcp-stdio --transport stdio`,
	RunE: runMCPTest,
}

func runMCPTest(cmd *cobra.Command, args []string) error {
	if serverURL == "" {
		return fmt.Errorf("--server is required")
	}

	green := color.New(color.FgGreen, color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed, color.Bold).SprintFunc()

	fmt.Printf("\n%s Testing MCP server connection\n", cyan("→"))
	fmt.Printf("  Server: %s\n", yellow(serverURL))
	fmt.Printf("  Transport: %s\n", yellow(transport))
	fmt.Println()

	// Create MCP client config
	config := &mcp.Config{
		Name:      "cli-test",
		ServerURL: serverURL,
		Transport: mcp.Transport(transport),
		Timeout:   time.Duration(timeout) * time.Second,
	}

	// Test 1: Create client
	fmt.Printf("%s Creating MCP client...\n", cyan("1."))
	client, err := mcp.NewMCPClient(config)
	if err != nil {
		fmt.Printf("  %s Failed to create client: %v\n\n", red("✗"), err)
		return err
	}
	defer client.Close()
	fmt.Printf("  %s Client created successfully\n", green("✓"))

	// Test 2: Connect
	fmt.Printf("\n%s Connecting to server...\n", cyan("2."))
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		fmt.Printf("  %s Connection failed: %v\n\n", red("✗"), err)
		return err
	}
	fmt.Printf("  %s Connected successfully\n", green("✓"))

	// Test 3: Ping
	fmt.Printf("\n%s Testing ping/health check...\n", cyan("3."))
	if err := client.Ping(ctx); err != nil {
		fmt.Printf("  %s Ping failed: %v\n\n", red("✗"), err)
		return err
	}
	fmt.Printf("  %s Server is responding\n", green("✓"))

	// Test 4: List tools
	fmt.Printf("\n%s Listing available tools...\n", cyan("4."))
	tools, err := client.ListTools(ctx)
	if err != nil {
		fmt.Printf("  %s Failed to list tools: %v\n\n", red("✗"), err)
		return err
	}
	fmt.Printf("  %s Found %d tools\n", green("✓"), len(tools))

	// Display tool names
	if len(tools) > 0 {
		fmt.Printf("\n%s Available tools:\n", green("→"))
		for _, tool := range tools {
			fmt.Printf("  • %s", yellow(tool.Name))
			if tool.Description != "" {
				fmt.Printf(" - %s", tool.Description)
			}
			fmt.Println()
		}
	}

	// Summary
	fmt.Printf("\n%s\n", green("═══════════════════════════════════════"))
	fmt.Printf("%s MCP server is operational\n", green("✓"))
	fmt.Printf("%s\n\n", green("═══════════════════════════════════════"))

	return nil
}
