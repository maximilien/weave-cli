// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package repl

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/maximilien/weave-cli/src/pkg/mcp"
)

// handleMCPCommand handles /mcp subcommands
func (r *REPL) handleMCPCommand(args []string) {
	if len(args) == 0 {
		r.displayMCPHelp()
		return
	}

	subcommand := strings.ToLower(args[0])
	subargs := args[1:]

	switch subcommand {
	case "list":
		r.mcpList()
	case "call":
		r.mcpCall(subargs)
	case "status":
		r.mcpStatus()
	case "help":
		r.displayMCPHelp()
	default:
		color.Red("Unknown MCP subcommand: %s", subcommand)
		r.displayMCPHelp()
	}
}

// handleCollectionCommand handles /collection subcommands
func (r *REPL) handleCollectionCommand(args []string) {
	if len(args) == 0 {
		r.collectionList()
		return
	}

	subcommand := strings.ToLower(args[0])
	subargs := args[1:]

	switch subcommand {
	case "list", "ls":
		r.collectionList()
	case "info":
		r.collectionInfo()
	case "use":
		r.collectionUse(subargs)
	case "create":
		r.collectionCreate(subargs)
	case "delete", "del":
		r.collectionDelete(subargs)
	case "help":
		r.displayCollectionHelp()
	default:
		color.Red("Unknown collection subcommand: %s", subcommand)
		r.displayCollectionHelp()
	}
}

// handleSearchCommand handles /search command
func (r *REPL) handleSearchCommand(args []string) {
	if len(args) == 0 {
		color.Yellow("Usage: /search <query> [--top N]")
		return
	}

	// Parse query and options
	query := strings.Join(args, " ")

	// Execute semantic search
	searchQuery := fmt.Sprintf("search for '%s' in %s", query, r.getCurrentCollectionOrPrompt())
	r.executeQuery(searchQuery)
}

// handleStatsCommand handles /stats command
func (r *REPL) handleStatsCommand(args []string) {
	if r.currentCollection == "" {
		color.Yellow("No collection selected. Use /use <collection> first")
		return
	}

	statsQuery := fmt.Sprintf("show statistics for collection %s", r.currentCollection)
	r.executeQuery(statsQuery)
}

// handleUseCommand handles /use command
func (r *REPL) handleUseCommand(args []string) {
	if len(args) == 0 {
		color.Yellow("Usage: /use <collection-name>")
		return
	}

	r.currentCollection = args[0]
	color.Green("✓ Now using collection: %s", r.currentCollection)

	// Update prompt to show current collection
	if r.rl != nil {
		prompt := fmt.Sprintf("\033[36m%s>\033[0m ", r.currentCollection)
		r.rl.SetPrompt(prompt)
	}
}

// displayStatus displays current REPL status
func (r *REPL) displayStatus() {
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	fmt.Println()
	fmt.Println(cyan("Current Status:"))
	fmt.Println()

	// VDB connection
	if r.currentVDB != "" {
		fmt.Printf("  VDB:        %s\n", green(r.currentVDB))
	} else {
		fmt.Printf("  VDB:        %s\n", yellow("not connected"))
	}

	// Collection
	if r.currentCollection != "" {
		fmt.Printf("  Collection: %s\n", green(r.currentCollection))
	} else {
		fmt.Printf("  Collection: %s\n", yellow("none selected"))
	}

	// MCP status
	if r.mcpEnabled && r.mcpClient != nil {
		fmt.Printf("  MCP:        %s\n", green("connected"))
		config := r.mcpClient.Config()
		fmt.Printf("  MCP Server: %s\n", cyan(config.ServerURL))
	} else {
		fmt.Printf("  MCP:        %s\n", yellow("not configured"))
	}

	fmt.Println()
}

// MCP command implementations
func (r *REPL) mcpList() {
	// Use direct MCP client if enabled
	if r.mcpEnabled && r.mcpClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		tools, err := r.mcpClient.ListTools(ctx)
		if err != nil {
			color.Red("❌ Failed to list MCP tools: %v", err)
			return
		}

		r.displayToolsList(tools)
		return
	}

	// Fallback to LLM-based execution
	color.Yellow("⚠️  No MCP server configured. Use --mcp-server flag when starting REPL")
	fmt.Println("   Falling back to LLM-based MCP tools...")
	r.executeQuery("list all available MCP tools")
}

func (r *REPL) mcpCall(args []string) {
	if len(args) == 0 {
		color.Yellow("Usage: /mcp call <tool-name> [args...]")
		return
	}

	toolName := args[0]

	// Use direct MCP client if enabled
	if r.mcpEnabled && r.mcpClient != nil {
		// Parse remaining args as JSON key=value pairs
		toolArgs := make(map[string]interface{})
		for _, arg := range args[1:] {
			parts := strings.SplitN(arg, "=", 2)
			if len(parts) == 2 {
				// Try to parse as JSON value
				var val interface{}
				if err := json.Unmarshal([]byte(parts[1]), &val); err != nil {
					// If not valid JSON, treat as string
					toolArgs[parts[0]] = parts[1]
				} else {
					toolArgs[parts[0]] = val
				}
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		fmt.Printf("→ Calling MCP tool: %s\n", toolName)
		result, err := r.mcpClient.CallTool(ctx, toolName, toolArgs)
		if err != nil {
			color.Red("❌ Failed to call MCP tool: %v", err)
			return
		}

		// Display result
		resultJSON, _ := json.MarshalIndent(result, "", "  ")
		color.Green("✓ Tool Result:")
		fmt.Println(string(resultJSON))
		return
	}

	// Fallback to LLM-based execution
	color.Yellow("⚠️  No MCP server configured. Use --mcp-server flag when starting REPL")
	fmt.Println("   Falling back to LLM-based MCP tools...")
	toolArgs := strings.Join(args[1:], " ")
	query := fmt.Sprintf("call MCP tool %s with arguments: %s", toolName, toolArgs)
	r.executeQuery(query)
}

func (r *REPL) mcpStatus() {
	if r.mcpEnabled && r.mcpClient != nil {
		// Try to ping the MCP server
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := r.mcpClient.Ping(ctx)
		if err != nil {
			color.Red("❌ MCP connection: inactive")
			fmt.Printf("   Error: %v\n", err)
			return
		}

		color.Green("✓ MCP connection: active")
		config := r.mcpClient.Config()
		fmt.Printf("   Server: %s\n", config.ServerURL)
		fmt.Printf("   Transport: %s\n", config.Transport)
		fmt.Println("   Type '/mcp list' to see available tools")
	} else {
		color.Yellow("⚠️  No MCP server configured")
		fmt.Println("   Use --mcp-server flag when starting REPL to enable direct MCP calls")
		fmt.Println("   Example: weave repl --mcp-server http://localhost:8030")
	}
}

func (r *REPL) displayMCPHelp() {
	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Println()
	fmt.Println("MCP Commands:")
	fmt.Println()
	fmt.Println("  " + cyan("/mcp list") + "              - List available MCP tools")
	fmt.Println("  " + cyan("/mcp call <tool> [args]") + " - Call an MCP tool")
	fmt.Println("  " + cyan("/mcp status") + "            - Show MCP connection status")
	fmt.Println()
}

// Collection command implementations
func (r *REPL) collectionList() {
	r.executeQuery("list all collections")
}

func (r *REPL) collectionInfo() {
	if r.currentCollection == "" {
		color.Yellow("No collection selected. Use /use <collection> first")
		return
	}

	query := fmt.Sprintf("show detailed information about collection %s", r.currentCollection)
	r.executeQuery(query)
}

func (r *REPL) collectionUse(args []string) {
	if len(args) == 0 {
		color.Yellow("Usage: /collection use <collection-name>")
		return
	}

	r.handleUseCommand(args)
}

func (r *REPL) collectionCreate(args []string) {
	if len(args) == 0 {
		color.Yellow("Usage: /collection create <collection-name>")
		return
	}

	collectionName := args[0]
	query := fmt.Sprintf("create a new text collection named %s", collectionName)
	r.executeQuery(query)
}

func (r *REPL) collectionDelete(args []string) {
	if len(args) == 0 {
		color.Yellow("Usage: /collection delete <collection-name>")
		return
	}

	collectionName := args[0]
	query := fmt.Sprintf("delete collection %s", collectionName)
	r.executeQuery(query)
}

func (r *REPL) displayCollectionHelp() {
	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Println()
	fmt.Println("Collection Commands:")
	fmt.Println()
	fmt.Println("  " + cyan("/collection list") + "          - List all collections (alias: /col list)")
	fmt.Println("  " + cyan("/collection use <name>") + "   - Set active collection")
	fmt.Println("  " + cyan("/collection create <name>") + " - Create new collection")
	fmt.Println("  " + cyan("/collection delete <name>") + " - Delete collection")
	fmt.Println("  " + cyan("/collection info") + "         - Show current collection info")
	fmt.Println()
}

// Helper functions
func (r *REPL) getCurrentCollectionOrPrompt() string {
	if r.currentCollection != "" {
		return fmt.Sprintf("collection %s", r.currentCollection)
	}
	return "all collections"
}

// displayToolsList displays MCP tools in a formatted list
func (r *REPL) displayToolsList(tools []mcp.Tool) {
	blue := color.New(color.FgBlue, color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	fmt.Println()
	fmt.Println(blue("Available MCP Tools:"))
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
				for paramName, paramData := range props {
					if paramMap, ok := paramData.(map[string]interface{}); ok {
						paramType := "any"
						if t, ok := paramMap["type"].(string); ok {
							paramType = t
						}
						desc := ""
						if d, ok := paramMap["description"].(string); ok {
							desc = " - " + d
						}
						fmt.Printf("     • %s (%s)%s\n", paramName, paramType, desc)
					}
				}
			}
		}
	}

	fmt.Println()
}
