// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/maximilien/weave-cli/src/pkg/mcp"
	"github.com/spf13/cobra"
)

var (
	callArgs         []string
	callOutputFormat string
)

var callCmd = &cobra.Command{
	Use:   "call TOOL_NAME",
	Short: "Call an MCP tool",
	Long: `Call an MCP tool with specified arguments.

Arguments are provided as key=value pairs using --arg flag.

Examples:
  # Call extract_entities tool
  weave mcp call extract_entities --server localhost:8030 \
    --arg text="Sample text to analyze"

  # Call with multiple arguments
  weave mcp call summarize --server localhost:8030 \
    --arg text="Long text..." \
    --arg max_length=100

  # Output as JSON
  weave mcp call classify --server localhost:8030 \
    --arg text="Text to classify" \
    --output json`,
	Args: cobra.ExactArgs(1),
	RunE: runMCPCall,
}

func init() {
	callCmd.Flags().StringSliceVar(&callArgs, "arg", []string{}, "Tool arguments as key=value pairs")
	callCmd.Flags().StringVar(&callOutputFormat, "output", "text", "Output format (text, json)")
}

func runMCPCall(cmd *cobra.Command, args []string) error {
	toolName := args[0]

	if serverURL == "" {
		return fmt.Errorf("--server is required")
	}

	// Parse arguments
	toolArgs := make(map[string]interface{})
	for _, arg := range callArgs {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid argument format: %s (expected key=value)", arg)
		}

		key := parts[0]
		value := parts[1]

		// Try to parse as JSON first, otherwise treat as string
		var jsonValue interface{}
		if err := json.Unmarshal([]byte(value), &jsonValue); err == nil {
			toolArgs[key] = jsonValue
		} else {
			toolArgs[key] = value
		}
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

	// Call tool
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue, color.Bold).SprintFunc()

	fmt.Printf("%s Calling tool: %s\n", green("→"), blue(toolName))

	result, err := client.CallTool(ctx, toolName, toolArgs)
	if err != nil {
		return fmt.Errorf("tool call failed: %w", err)
	}

	// Display result
	if callOutputFormat == "json" {
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
	} else {
		displayToolResult(result)
	}

	return nil
}

func displayToolResult(result interface{}) {
	green := color.New(color.FgGreen, color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	fmt.Printf("\n%s\n", green("Tool Result:"))
	fmt.Println(green("============"))

	// Handle different result types
	switch v := result.(type) {
	case string:
		fmt.Println(v)
	case []interface{}:
		// Handle content array (MCP standard format)
		for _, item := range v {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if contentType, ok := itemMap["type"].(string); ok && contentType == "text" {
					if text, ok := itemMap["text"].(string); ok {
						fmt.Println(text)
					}
				} else {
					// Print as JSON for non-text content
					data, _ := json.MarshalIndent(item, "", "  ")
					fmt.Println(string(data))
				}
			}
		}
	case map[string]interface{}:
		// Pretty print map
		for key, val := range v {
			fmt.Printf("%s: ", cyan(key))
			switch tv := val.(type) {
			case string:
				fmt.Println(tv)
			default:
				data, _ := json.MarshalIndent(val, "  ", "  ")
				fmt.Println(string(data))
			}
		}
	default:
		// Fallback to JSON
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Printf("%v\n", result)
		} else {
			fmt.Println(string(data))
		}
	}

	fmt.Println()
}
