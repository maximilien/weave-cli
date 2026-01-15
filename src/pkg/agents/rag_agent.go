// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/maximilien/weave-cli/src/pkg/llm"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// RAGAgent implements RAG (Retrieval-Augmented Generation) pattern
type RAGAgent struct {
	config         *CustomAgentConfig
	llmClient      llm.Client
	contextBuilder *ContextBuilder
}

// RAGInput is input for RAG agent
type RAGInput struct {
	Query   string                  `json:"query"`
	Results []*vectordb.QueryResult `json:"results"`
}

// RAGOutput is output from RAG agent
type RAGOutput struct {
	Answer     string           `json:"answer"`
	Sources    []SourceCitation `json:"sources,omitempty"`
	Confidence float64          `json:"confidence,omitempty"`
	Metadata   *RAGMetadata     `json:"metadata,omitempty"`
}

// SourceCitation represents a cited source
type SourceCitation struct {
	Index    int                    `json:"index"`
	Content  string                 `json:"content,omitempty"`
	Score    float64                `json:"score"`
	DocID    string                 `json:"doc_id"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// RAGMetadata contains metadata about the RAG response
type RAGMetadata struct {
	TotalSources int     `json:"total_sources"`
	SourcesUsed  int     `json:"sources_used"`
	MinRelevance float64 `json:"min_relevance"`
	MaxRelevance float64 `json:"max_relevance"`
	AvgRelevance float64 `json:"avg_relevance"`
	Model        string  `json:"model"`
	Temperature  float64 `json:"temperature"`
}

// NewRAGAgent creates a new RAG agent
func NewRAGAgent(config *CustomAgentConfig, llmClient llm.Client) (*RAGAgent, error) {
	if config == nil {
		return nil, fmt.Errorf("config is required")
	}
	if llmClient == nil {
		return nil, fmt.Errorf("llm client is required")
	}

	// Validate config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &RAGAgent{
		config:         config,
		llmClient:      llmClient,
		contextBuilder: NewContextBuilder(config),
	}, nil
}

// Name returns the agent's name
func (a *RAGAgent) Name() string {
	return a.config.Name
}

// Execute processes the RAG request
func (a *RAGAgent) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	ragInput, ok := input.(*RAGInput)
	if !ok {
		return nil, fmt.Errorf("invalid input type for RAGAgent, expected *RAGInput")
	}

	// Build context from query results
	queryContext, err := a.contextBuilder.BuildContext(ragInput.Query, ragInput.Results)
	if err != nil {
		return nil, fmt.Errorf("failed to build context: %w", err)
	}

	// Generate response using LLM
	answer, err := a.generateResponse(ctx, queryContext)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}

	// Build output
	output := a.buildOutput(answer, queryContext)

	return output, nil
}

// generateResponse generates the RAG response using the LLM
func (a *RAGAgent) generateResponse(ctx context.Context, queryContext *QueryContext) (string, error) {
	// Build the prompt
	var prompt string
	if a.config.UserPromptTemplate != "" {
		// Use custom template
		prompt = a.buildCustomPrompt(queryContext)
	} else {
		// Use default formatting with citations
		prompt = a.contextBuilder.FormatContextForPromptWithCitations(queryContext)
	}

	// Add assistant hint if configured
	if a.config.AssistantPromptHint != "" {
		prompt = fmt.Sprintf("%s\n\nAssistant guidance: %s", prompt, a.config.AssistantPromptHint)
	}

	// Call LLM
	response, err := a.llmClient.Complete(
		ctx,
		prompt,
		llm.WithSystemMessage(a.config.SystemPrompt),
		llm.WithModel(a.config.LLM.Model),
		llm.WithTemperature(a.config.LLM.Temperature),
		llm.WithMaxTokens(a.config.LLM.MaxTokens),
	)

	if err != nil {
		return "", fmt.Errorf("llm completion failed: %w", err)
	}

	return strings.TrimSpace(response), nil
}

// buildCustomPrompt builds a prompt using the custom template
func (a *RAGAgent) buildCustomPrompt(queryContext *QueryContext) string {
	// Simple template variable replacement
	template := a.config.UserPromptTemplate

	// Replace {{query}}
	template = strings.ReplaceAll(template, "{{query}}", queryContext.Query)

	// Replace {{sources}} with formatted sources
	sourcesText := a.formatSourcesForTemplate(queryContext)
	template = strings.ReplaceAll(template, "{{sources}}", sourcesText)

	// Replace {{source_count}}
	template = strings.ReplaceAll(template, "{{source_count}}", fmt.Sprintf("%d", len(queryContext.Sources)))

	return template
}

// formatSourcesForTemplate formats sources for template insertion
func (a *RAGAgent) formatSourcesForTemplate(queryContext *QueryContext) string {
	if len(queryContext.Sources) == 0 {
		return "No sources available."
	}

	var builder strings.Builder
	for _, source := range queryContext.Sources {
		builder.WriteString(fmt.Sprintf("[%d] %s (Score: %.3f)\n", source.Index, source.Content, source.Score))
	}
	return builder.String()
}

// buildOutput constructs the RAGOutput from the LLM response and context
func (a *RAGAgent) buildOutput(answer string, queryContext *QueryContext) *RAGOutput {
	output := &RAGOutput{
		Answer: answer,
	}

	// Add sources if configured
	if a.config.Response.IncludeReferences {
		output.Sources = a.buildSourceCitations(queryContext)
	}

	// Add metadata if configured
	if a.config.Output.IncludeMetadata {
		output.Metadata = a.buildMetadata(queryContext)
	}

	// Add confidence if configured
	if a.config.Output.ShowConfidence {
		output.Confidence = a.calculateConfidence(queryContext)
	}

	return output
}

// buildSourceCitations builds the source citations list
func (a *RAGAgent) buildSourceCitations(queryContext *QueryContext) []SourceCitation {
	citations := make([]SourceCitation, len(queryContext.Sources))

	for i, source := range queryContext.Sources {
		citation := SourceCitation{
			Index: source.Index,
			Score: source.Score,
			DocID: source.DocID,
		}

		// Include content if ShowSources is enabled
		if a.config.Output.ShowSources {
			citation.Content = source.Content
		}

		// Include metadata if configured
		if a.config.Output.IncludeMetadata && len(source.Metadata) > 0 {
			citation.Metadata = source.Metadata
		}

		citations[i] = citation
	}

	return citations
}

// buildMetadata builds the RAG metadata
func (a *RAGAgent) buildMetadata(queryContext *QueryContext) *RAGMetadata {
	metadata := &RAGMetadata{
		TotalSources: queryContext.TotalResults,
		SourcesUsed:  len(queryContext.Sources),
		Model:        a.config.LLM.Model,
		Temperature:  a.config.LLM.Temperature,
	}

	// Calculate relevance statistics
	if len(queryContext.Sources) > 0 {
		minScore := 1.0
		maxScore := 0.0
		totalScore := 0.0

		for _, source := range queryContext.Sources {
			if source.Score < minScore {
				minScore = source.Score
			}
			if source.Score > maxScore {
				maxScore = source.Score
			}
			totalScore += source.Score
		}

		metadata.MinRelevance = minScore
		metadata.MaxRelevance = maxScore
		metadata.AvgRelevance = totalScore / float64(len(queryContext.Sources))
	}

	return metadata
}

// calculateConfidence calculates a confidence score for the response
func (a *RAGAgent) calculateConfidence(queryContext *QueryContext) float64 {
	if len(queryContext.Sources) == 0 {
		return 0.0
	}

	// Simple confidence calculation based on:
	// 1. Number of sources (more is better, up to max_context_chunks)
	// 2. Average relevance score
	// 3. Ratio of sources used vs total results

	sourceCountFactor := float64(len(queryContext.Sources)) / float64(a.config.Response.MaxContextChunks)
	if sourceCountFactor > 1.0 {
		sourceCountFactor = 1.0
	}

	// Calculate average relevance
	totalScore := 0.0
	for _, source := range queryContext.Sources {
		totalScore += source.Score
	}
	avgRelevance := totalScore / float64(len(queryContext.Sources))

	// Combined confidence (weighted average)
	confidence := (sourceCountFactor * 0.3) + (avgRelevance * 0.7)

	return confidence
}

// FormatOutput formats the RAG output based on the configured format
func (a *RAGAgent) FormatOutput(output *RAGOutput) (string, error) {
	switch a.config.Output.Format {
	case "json":
		return a.formatJSON(output)
	case "markdown":
		return a.formatMarkdown(output)
	case "text":
		return a.formatText(output)
	default:
		return "", fmt.Errorf("unsupported output format: %s", a.config.Output.Format)
	}
}

// formatJSON formats output as JSON
func (a *RAGAgent) formatJSON(output *RAGOutput) (string, error) {
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(data), nil
}

// formatMarkdown formats output as Markdown
func (a *RAGAgent) formatMarkdown(output *RAGOutput) (string, error) {
	var builder strings.Builder

	// Answer section
	builder.WriteString("## Answer\n\n")
	builder.WriteString(output.Answer)
	builder.WriteString("\n\n")

	// Sources section
	if len(output.Sources) > 0 && a.config.Output.ShowSources {
		builder.WriteString("## Sources\n\n")
		for _, source := range output.Sources {
			// Extract collection name and VDB info if available
			collectionName := ""
			vdbName := ""
			if source.Metadata != nil {
				if col, ok := source.Metadata["_collection"].(string); ok {
					collectionName = col
				}
				if vdb, ok := source.Metadata["_vdb"].(string); ok {
					vdbName = vdb
				}
			}

			// Format citation with collection name and VDB
			if collectionName != "" && vdbName != "" {
				builder.WriteString(fmt.Sprintf("**[%d]** %s (%s) - Score: %.1f%%\n", source.Index, collectionName, vdbName, source.Score*100))
			} else if collectionName != "" {
				builder.WriteString(fmt.Sprintf("**[%d]** %s (Score: %.1f%%)\n", source.Index, collectionName, source.Score*100))
			} else {
				builder.WriteString(fmt.Sprintf("**[%d]** (Score: %.1f%%)\n", source.Index, source.Score*100))
			}

			if source.Content != "" {
				builder.WriteString(fmt.Sprintf("- %s\n", source.Content))
			}
			if source.DocID != "" {
				builder.WriteString(fmt.Sprintf("- Document ID: `%s`\n", source.DocID))
			}
			builder.WriteString("\n")
		}
	}

	// Metadata section
	if output.Metadata != nil {
		builder.WriteString("## Metadata\n\n")
		builder.WriteString(fmt.Sprintf("- **Model**: %s\n", output.Metadata.Model))
		builder.WriteString(fmt.Sprintf("- **Sources Used**: %d / %d\n", output.Metadata.SourcesUsed, output.Metadata.TotalSources))
		if output.Metadata.AvgRelevance > 0 {
			builder.WriteString(fmt.Sprintf("- **Avg Relevance**: %.1f%%\n", output.Metadata.AvgRelevance*100))
		}
		if output.Confidence > 0 {
			builder.WriteString(fmt.Sprintf("- **Confidence**: %.1f%%\n", output.Confidence*100))
		}
	}

	return builder.String(), nil
}

// formatText formats output as plain text
func (a *RAGAgent) formatText(output *RAGOutput) (string, error) {
	var builder strings.Builder

	// Answer
	builder.WriteString(output.Answer)
	builder.WriteString("\n\n")

	// Sources
	if len(output.Sources) > 0 && a.config.Output.ShowSources {
		builder.WriteString("Sources:\n")
		for _, source := range output.Sources {
			// Extract collection name and VDB info if available
			collectionName := ""
			vdbName := ""
			if source.Metadata != nil {
				if col, ok := source.Metadata["_collection"].(string); ok {
					collectionName = col
				}
				if vdb, ok := source.Metadata["_vdb"].(string); ok {
					vdbName = vdb
				}
			}

			// Format with collection name and VDB
			if collectionName != "" && vdbName != "" {
				builder.WriteString(fmt.Sprintf("[%d] %s (%s) - Score: %.1f%%", source.Index, collectionName, vdbName, source.Score*100))
			} else if collectionName != "" {
				builder.WriteString(fmt.Sprintf("[%d] %s - Score: %.1f%%", source.Index, collectionName, source.Score*100))
			} else {
				builder.WriteString(fmt.Sprintf("[%d] Score: %.1f%%", source.Index, source.Score*100))
			}

			if source.DocID != "" {
				builder.WriteString(fmt.Sprintf(" (ID: %s)", source.DocID))
			}
			builder.WriteString("\n")
		}
	}

	return builder.String(), nil
}

// GetConfig returns the agent configuration
func (a *RAGAgent) GetConfig() *CustomAgentConfig {
	return a.config
}

// GetType returns the agent type
func (a *RAGAgent) GetType() string {
	return a.config.Type
}

// IsRAGType returns true if this is a RAG-type agent
func (a *RAGAgent) IsRAGType() bool {
	return a.config.IsRAGType()
}
