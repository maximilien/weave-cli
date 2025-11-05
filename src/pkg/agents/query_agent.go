// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/maximilien/weave-cli/src/pkg/llm"
)

// QueryAgent validates and normalizes user queries
type QueryAgent struct {
	llmClient llm.Client
	mcpTools  []string
}

// NewQueryAgent creates a new QueryAgent
func NewQueryAgent(llmClient llm.Client) *QueryAgent {
	return &QueryAgent{
		llmClient: llmClient,
		mcpTools:  []string{},
	}
}

// SetMCPTools sets the available MCP tools for validation
func (a *QueryAgent) SetMCPTools(tools []string) {
	a.mcpTools = tools
}

// Name returns the agent's name
func (a *QueryAgent) Name() string {
	return "QueryAgent"
}

// Execute processes the query and determines if it's weave-related
func (a *QueryAgent) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	queryInput, ok := input.(*QueryAgentInput)
	if !ok {
		return nil, fmt.Errorf("invalid input type for QueryAgent")
	}

	prompt := a.buildPrompt(queryInput.Query)

	var output QueryAgentOutput
	_, err := a.llmClient.CompleteStructured(
		ctx,
		prompt,
		&output,
		llm.WithSystemMessage(a.getSystemMessage()),
		llm.WithTemperature(0.1),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to analyze query: %w", err)
	}

	return &output, nil
}

// getSystemMessage returns the system message for the LLM
func (a *QueryAgent) getSystemMessage() string {
	return `You are a query validation agent for weave-cli, a vector database management tool.

Your task is to analyze user queries and determine if they are related to weave-cli operations.

Weave-cli supports:
- Health checks (check database connectivity and status)
- Collection management (list, create, delete, show, count)
- Document management (create, update, delete, list, query, show, count)
- PDF processing (text extraction, image extraction, CMYK conversion)
- Batch operations (process directories with parallel workers)
- Semantic search (query documents with natural language)

You must respond with valid JSON only.`
}

// buildPrompt builds the prompt for the LLM
func (a *QueryAgent) buildPrompt(query string) string {
	toolsList := ""
	if len(a.mcpTools) > 0 {
		toolsList = fmt.Sprintf("\n\nAvailable weave-cli MCP tools: %s", strings.Join(a.mcpTools, ", "))
	}

	return fmt.Sprintf(`Analyze the following user query and determine if it's related to weave-cli operations.

User query: "%s"%s

Determine:
1. Is this query about weave-cli operations? (health checks, vector database management, collections, documents, PDFs)
2. If yes, fix any grammar/spelling errors and clarify ambiguities
3. Determine the primary intent (health, list, create, delete, update, query, show, count, batch, convert)
4. Provide a confidence score (0.0 to 1.0)
5. Explain your reasoning

Examples of weave-cli queries:
- "check health"
- "is the database connected?"
- "show me all my collections"
- "add documents from /tmp/docs to MyCollection"
- "find empty collections"
- "convert CMYK PDF to RGB"
- "search for documents about AI in MyDocs"

Examples of non-weave-cli queries:
- "how do I bake a cake?"
- "what's the weather today?"
- "explain quantum physics"

Respond with JSON in this exact format:
{
  "is_weave_query": true or false,
  "fixed_query": "the corrected query text",
  "intent": "health|list|create|delete|update|query|show|count|batch|convert",
  "confidence": 0.95,
  "reason": "explanation of why this is or isn't a weave query"
}`, query, toolsList)
}

// ParseJSONResponse is a helper to parse JSON responses
func ParseJSONResponse(response string, target interface{}) error {
	return json.Unmarshal([]byte(response), target)
}
