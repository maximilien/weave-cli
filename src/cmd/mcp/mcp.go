// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package mcp

import (
	"github.com/spf13/cobra"
)

// MCPCmd represents the mcp command group
var MCPCmd = &cobra.Command{
	Use:   "mcp",
	Short: "MCP (Model Context Protocol) client commands",
	Long: `MCP (Model Context Protocol) client commands.

Connect to external MCP servers to access tools and services.
Supports both HTTP and stdio transports.

Examples:
  # List tools from an MCP server
  weave mcp list --server http://localhost:8030

  # Call an MCP tool
  weave mcp call extract_entities --server localhost:8030 --arg text="Sample text"

  # Test MCP server connection
  weave mcp test --server stdio:./bin/weave-mcp-stdio`,
}

var (
	// Global MCP flags
	serverURL string
	transport string
	timeout   int
)

func init() {
	// Add persistent flags for all mcp subcommands
	MCPCmd.PersistentFlags().StringVar(&serverURL, "server", "", "MCP server URL (HTTP) or command (stdio)")
	MCPCmd.PersistentFlags().StringVar(&transport, "transport", "http", "Transport protocol (http or stdio)")
	MCPCmd.PersistentFlags().IntVar(&timeout, "timeout", 30, "Request timeout in seconds")

	// Add subcommands
	MCPCmd.AddCommand(listCmd)
	MCPCmd.AddCommand(callCmd)
	MCPCmd.AddCommand(testCmd)
}
