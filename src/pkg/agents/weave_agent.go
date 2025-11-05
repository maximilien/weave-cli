// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package agents

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/mcp"
)

// WeaveAgent executes weave-cli commands via MCP
type WeaveAgent struct {
	mcpClient *mcp.Client
	verbose   bool
}

// NewWeaveAgent creates a new WeaveAgent
func NewWeaveAgent(mcpClient *mcp.Client) *WeaveAgent {
	return &WeaveAgent{
		mcpClient: mcpClient,
		verbose:   false,
	}
}

// SetVerbose sets verbose mode for debug logging
func (a *WeaveAgent) SetVerbose(verbose bool) {
	a.verbose = verbose
}

// Name returns the agent's name
func (a *WeaveAgent) Name() string {
	return "WeaveAgent"
}

// Execute executes a weave command via MCP
func (a *WeaveAgent) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	cmd, ok := input.(*WeaveAgentCommand)
	if !ok {
		return nil, fmt.Errorf("invalid input type for WeaveAgent")
	}

	startTime := time.Now()
	retries := 0
	maxRetries := 3

	var result interface{}
	var err error

	// Retry loop
	for retries <= maxRetries {
		// Set timeout if specified
		cmdCtx := ctx
		if cmd.Timeout > 0 {
			var cancel context.CancelFunc
			cmdCtx, cancel = context.WithTimeout(ctx, cmd.Timeout)
			defer cancel()
		}

		// Call the MCP tool
		result, err = a.mcpClient.CallTool(cmdCtx, cmd.Tool, cmd.Arguments)

		// Debug logging in verbose mode
		if a.verbose {
			fmt.Fprintf(os.Stderr, "\n[DEBUG] MCP Tool: %s\n", cmd.Tool)
			fmt.Fprintf(os.Stderr, "[DEBUG] MCP Arguments: %+v\n", cmd.Arguments)
			fmt.Fprintf(os.Stderr, "[DEBUG] MCP Result: %+v\n", result)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[DEBUG] MCP Error: %v\n", err)
			}
			fmt.Fprintln(os.Stderr)
		}

		if err == nil {
			break
		}

		retries++
		if retries <= maxRetries {
			// Exponential backoff
			backoff := time.Duration(retries) * time.Second
			select {
			case <-time.After(backoff):
				continue
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	duration := time.Since(startTime)

	weaveResult := &WeaveAgentResult{
		Success:  err == nil,
		Output:   result,
		Duration: duration,
		Retries:  retries,
	}

	if err != nil {
		weaveResult.Error = err.Error()
	}

	return weaveResult, nil
}

// ValidateArguments validates that required arguments are present
func (a *WeaveAgent) ValidateArguments(tool string, args map[string]interface{}) error {
	requiredArgs := getRequiredArguments(tool)

	for _, required := range requiredArgs {
		if _, ok := args[required]; !ok {
			return fmt.Errorf("missing required argument: %s for tool: %s", required, tool)
		}
	}

	return nil
}

// getRequiredArguments returns required arguments for each tool
func getRequiredArguments(tool string) []string {
	toolRequirements := map[string][]string{
		"list_collections":  {},
		"create_collection": {"name", "type"},
		"delete_collection": {"name"},
		"list_documents":    {"collection"},
		"create_document":   {"collection", "url", "text"},
		"get_document":      {"collection", "document_id"},
		"update_document":   {"collection", "document_id"},
		"delete_document":   {"collection", "document_id"},
		"count_documents":   {"collection"},
		"query_documents":   {"collection", "query"},
	}

	if requirements, ok := toolRequirements[tool]; ok {
		return requirements
	}

	return []string{}
}

// InferMissingArguments attempts to infer missing arguments from context
func (a *WeaveAgent) InferMissingArguments(tool string, args map[string]interface{}, context map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range args {
		result[k] = v
	}

	// Try to infer from context
	required := getRequiredArguments(tool)
	for _, req := range required {
		if _, ok := result[req]; !ok {
			// Try to find in context
			if val, exists := context[req]; exists {
				result[req] = val
			}
		}
	}

	return result
}
