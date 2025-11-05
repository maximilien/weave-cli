// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/maximilien/weave-cli/src/pkg/llm"
	"github.com/maximilien/weave-cli/src/pkg/mcp"
)

// PlanningAgent creates execution plans for queries
type PlanningAgent struct {
	llmClient    llm.Client
	mcpTools     []string
	mcpToolsInfo []mcp.Tool
}

// NewPlanningAgent creates a new PlanningAgent with tool names only
func NewPlanningAgent(llmClient llm.Client, mcpTools []string) *PlanningAgent {
	return &PlanningAgent{
		llmClient: llmClient,
		mcpTools:  mcpTools,
	}
}

// NewPlanningAgentWithTools creates a new PlanningAgent with full tool information
func NewPlanningAgentWithTools(llmClient llm.Client, tools []mcp.Tool) *PlanningAgent {
	toolNames := make([]string, len(tools))
	for i, tool := range tools {
		toolNames[i] = tool.Name
	}
	return &PlanningAgent{
		llmClient:    llmClient,
		mcpTools:     toolNames,
		mcpToolsInfo: tools,
	}
}

// Name returns the agent's name
func (a *PlanningAgent) Name() string {
	return "PlanningAgent"
}

// Execute creates an execution plan for the query
func (a *PlanningAgent) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	planInput, ok := input.(*PlanningAgentInput)
	if !ok {
		return nil, fmt.Errorf("invalid input type for PlanningAgent")
	}

	prompt := a.buildPrompt(planInput.FixedQuery, planInput.Intent)

	var plan ExecutionPlan
	_, err := a.llmClient.CompleteStructured(
		ctx,
		prompt,
		&plan,
		llm.WithSystemMessage(a.getSystemMessage()),
		llm.WithTemperature(0.1),
		llm.WithMaxTokens(4096),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create execution plan: %w", err)
	}

	return &plan, nil
}

// getSystemMessage returns the system message for the LLM
func (a *PlanningAgent) getSystemMessage() string {
	return `You are a planning agent for weave-cli operations.

Your task is to create detailed execution plans for user queries.

You must create a sequence of steps that can be executed to fulfill the user's request.
Each step should be either a bash command or a weave-mcp tool call.

You must respond with valid JSON only.`
}

// buildPrompt builds the prompt for the LLM
func (a *PlanningAgent) buildPrompt(query string, intent string) string {
	var mcpToolsInfo string
	if len(a.mcpToolsInfo) > 0 {
		// Build detailed tool information with schemas
		var toolsBuilder strings.Builder
		for _, tool := range a.mcpToolsInfo {
			toolsBuilder.WriteString(fmt.Sprintf("\n- %s: %s\n", tool.Name, tool.Description))
			if tool.InputSchema != nil {
				schemaJSON, _ := json.MarshalIndent(tool.InputSchema, "  ", "  ")
				toolsBuilder.WriteString(fmt.Sprintf("  Parameters: %s\n", string(schemaJSON)))
			}
		}
		mcpToolsInfo = toolsBuilder.String()
	} else {
		// Fallback to simple tool list
		mcpToolsInfo = strings.Join(a.mcpTools, ", ")
	}

	safeBashCommands := strings.Join(getSafeBashCommands(), ", ")

	return fmt.Sprintf(`Create an execution plan for the following weave-cli query.

Query: "%s"
Intent: "%s"

Available MCP tools:%s

Safe bash commands: %s

Instructions:
1. PREFER SIMPLE BASH COMMANDS: Always use simple bash commands with weave CLI when possible
2. For adding documents: Use 'weave docs create COLLECTION FILE' - DO NOT read files with cat
3. For listing/querying: Use 'weave cols ls', 'weave docs ls COLLECTION' directly
4. For DELETE operations: ALWAYS use MCP tools (delete_collection, delete_document) - NOT bash commands
5. Break down the query into a sequence of executable steps
6. Each step must be either "bash" or "weave" type
7. For bash steps, use weave CLI commands (e.g., 'weave cols ls', 'weave docs create'), NOT MCP tool names
8. For weave steps, specify the MCP tool name and ALL required parameters from the tool's schema
9. IMPORTANT: Each MCP tool has specific required parameters - you MUST include them in the "params" field
10. Look at the "Parameters" schema above for each tool to see what fields are required
11. Identify dependencies between steps (use depends_on array with step indices)
12. Mark destructive operations (delete, batch operations) as destructive: true
13. Estimate duration and risk level
14. Add warnings for potentially dangerous operations
15. IMPORTANT: In bash scripts, use CLI commands: 'weave cols ls', 'weave docs count', etc.
16. IMPORTANT: In weave steps, use MCP tool names with all required parameters
17. Non-standard bash commands will require user approval before execution
18. IMPORTANT: The 'weave' CLI supports --json flag for JSON output - ALWAYS use it in bash scripts with jq
19. Format: 'weave cols ls --json' NOT 'weave list_collections --json'

Example plan structure:
{
  "steps": [
    {
      "type": "weave",
      "command": "list_collections",
      "description": "List all collections",
      "params": {},
      "depends_on": [],
      "optional": false,
      "destructive": false
    },
    {
      "type": "bash",
      "command": "for col in $(weave cols ls --json | jq -r '.collections[].name'); do count=$(weave docs count $col --json 2>/dev/null | jq -r '.count' || echo 0); echo \"$col: $count docs\"; done",
      "description": "Count documents in each collection",
      "args": [],
      "depends_on": [0],
      "optional": false,
      "destructive": false
    }
  ],
  "summary": "List all collections and count documents in each",
  "warnings": ["This will query all collections and count documents"],
  "estimations": {
    "duration": "10-20 seconds",
    "risk": "low"
  }
}

IMPORTANT: For bash scripts with loops/pipes, put the ENTIRE script in the "command" field.
For simple commands like "ls", you can use "command": "ls" with "args": ["-la", "path"].
For complex scripts, use "command": "for col in ...; do ...; done" with "args": [].
The 'weave' binary is available in PATH - use 'weave' NOT './bin/weave'.
Always use --json flag when using weave commands in bash scripts with jq parsing.

Common patterns:

1. Health checks:
   - For health/status checks, use bash: 'weave health check'
   - Example: "check health" → bash step with command "weave health check"
   - Intent "health" should ALWAYS map to this bash command

2. Document creation (adding files):
   - ALWAYS use bash: 'weave docs create COLLECTION FILE_PATH'
   - Example: "add README.md to TestDocs" → bash step: 'weave docs create TestDocs ./README.md'
   - DO NOT use cat/read file content and then create_document MCP tool
   - The CLI handles file reading automatically

3. Simple operations (single collection/document):
   - Use bash commands with weave CLI for simplicity
   - Example: "show collection details" → bash: 'weave cols show COLLECTION'
   - Example: "list documents" → bash: 'weave docs ls COLLECTION'

4. Bulk operations (multiple collections/documents):
   - Use bash scripts with for loops to iterate
   - Call weave CLI commands for each item
   - User will approve the script before execution
   - Examples:
     * Count documents in all collections
     * Delete all empty collections
     * List documents in multiple collections
     * Batch create/update operations

5. Pattern for iteration operations:
   - Step 1: Get list of items (collections, files, etc.)
   - Step 2: Bash script with for loop that:
     a) Iterates over items from step 1
     b) Calls weave CLI for each item
     c) Processes/displays results
   - Example: for col in $(weave cols ls --json | jq -r '.collections[].name'); do weave docs count $col --json; done

6. File operations:
   - Step 1: bash to list/find files
   - Step 2: bash to classify or filter
   - Step 3: bash loop to process each file with weave
   - Example: Process all PDFs in a directory

7. Conditional operations:
   - Use bash scripts with if/then logic
   - Example: Find and delete empty collections
   - Script can check conditions and take actions

KEY PRINCIPLES:
- When user asks for operation on "all" or "each" → use bash iteration
- When user asks for single item operation → use MCP tool directly
- Bash scripts give flexibility for complex multi-step operations
- Always explain what the script will do in warnings
- User approves non-standard bash commands before execution

Respond with JSON in the exact format shown above.`, query, intent, mcpToolsInfo, safeBashCommands)
}

// getSafeBashCommands returns list of safe bash commands
func getSafeBashCommands() []string {
	return []string{
		"ls",
		"file",
		"stat",
		"wc",
		"grep",
		"find",
		"cat",
		"head",
		"tail",
	}
}
