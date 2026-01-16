// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package agents

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// QueryContext represents context built from query results
type QueryContext struct {
	Query        string          `json:"query"`
	Sources      []SourceContext `json:"sources"`
	TotalResults int             `json:"total_results"`
	FilteredBy   FilterInfo      `json:"filtered_by,omitempty"`
}

// SourceContext represents a single source document with relevance
type SourceContext struct {
	Index    int                    `json:"index"`    // Citation index (1-based)
	Content  string                 `json:"content"`  // Document content
	Score    float64                `json:"score"`    // Relevance score
	DocID    string                 `json:"doc_id"`   // Document ID
	Metadata map[string]interface{} `json:"metadata"` // Document metadata
}

// FilterInfo describes what filters were applied
type FilterInfo struct {
	MinRelevanceScore float64 `json:"min_relevance_score,omitempty"`
	MaxContextChunks  int     `json:"max_context_chunks,omitempty"`
	Deduplicated      bool    `json:"deduplicated,omitempty"`
	SortedByRelevance bool    `json:"sorted_by_relevance,omitempty"`
}

// ContextBuilder builds agent context from vector database query results
type ContextBuilder struct {
	config *CustomAgentConfig
}

// NewContextBuilder creates a new context builder with the given agent config
func NewContextBuilder(config *CustomAgentConfig) *ContextBuilder {
	return &ContextBuilder{
		config: config,
	}
}

// BuildContext converts vector database query results into structured agent context
func (cb *ContextBuilder) BuildContext(query string, results []*vectordb.QueryResult) (*QueryContext, error) {
	if cb.config == nil {
		return nil, fmt.Errorf("context builder requires a valid agent config")
	}

	context := &QueryContext{
		Query:        query,
		Sources:      make([]SourceContext, 0),
		TotalResults: len(results),
		FilteredBy: FilterInfo{
			MinRelevanceScore: cb.config.Response.MinRelevanceScore,
			MaxContextChunks:  cb.config.Response.MaxContextChunks,
			Deduplicated:      cb.config.Response.DeduplicateSources,
			SortedByRelevance: cb.config.Response.SortByRelevance,
		},
	}

	// Filter by minimum relevance score
	filteredResults := cb.filterByRelevance(results)

	// Sort by relevance FIRST if enabled (before deduplication)
	// This ensures we keep the highest-scored version of each document
	// when querying the same collection across multiple VDBs
	if cb.config.Response.SortByRelevance {
		cb.sortByRelevance(filteredResults)
	}

	// Deduplicate AFTER sorting (so highest-scored duplicates are kept)
	if cb.config.Response.DeduplicateSources {
		filteredResults = cb.deduplicate(filteredResults)
	}

	// Limit to max context chunks
	if len(filteredResults) > cb.config.Response.MaxContextChunks {
		filteredResults = filteredResults[:cb.config.Response.MaxContextChunks]
	}

	// Build source contexts
	for i, result := range filteredResults {
		source := cb.buildSourceContext(i+1, result)
		context.Sources = append(context.Sources, source)
	}

	return context, nil
}

// BuildContextFromQueryResults is a convenience function that creates a context builder and builds context
func BuildContextFromQueryResults(query string, results []*vectordb.QueryResult, config *CustomAgentConfig) (*QueryContext, error) {
	builder := NewContextBuilder(config)
	return builder.BuildContext(query, results)
}

// filterByRelevance filters results by minimum relevance score
func (cb *ContextBuilder) filterByRelevance(results []*vectordb.QueryResult) []*vectordb.QueryResult {
	if cb.config.Response.MinRelevanceScore <= 0 {
		return results
	}

	filtered := make([]*vectordb.QueryResult, 0, len(results))
	for _, result := range results {
		if result.Score >= cb.config.Response.MinRelevanceScore {
			filtered = append(filtered, result)
		}
	}
	return filtered
}

// deduplicate removes duplicate documents based on content similarity
func (cb *ContextBuilder) deduplicate(results []*vectordb.QueryResult) []*vectordb.QueryResult {
	if len(results) <= 1 {
		return results
	}

	seen := make(map[string]bool)
	deduplicated := make([]*vectordb.QueryResult, 0, len(results))

	for _, result := range results {
		// Use document ID as primary deduplication key
		key := result.Document.ID
		if key == "" {
			// Fallback to content hash if no ID
			key = cb.contentHash(result.Document.Content)
		}

		if !seen[key] {
			seen[key] = true
			deduplicated = append(deduplicated, result)
		}
	}

	return deduplicated
}

// contentHash creates a simple hash of content for deduplication
func (cb *ContextBuilder) contentHash(content string) string {
	// Simple hash: first 100 chars + length
	normalized := strings.TrimSpace(content)
	if len(normalized) > 100 {
		return fmt.Sprintf("%s:%d", normalized[:100], len(normalized))
	}
	return normalized
}

// sortByRelevance sorts results by score in descending order
// For scores within epsilon (approximately equal), randomly shuffles them
// to avoid bias toward first VDB in aggregation
func (cb *ContextBuilder) sortByRelevance(results []*vectordb.QueryResult) {
	if len(results) <= 1 {
		return
	}

	// Epsilon for floating-point score comparison
	// Scores within this threshold are considered "equal" for fairness
	const epsilon = 0.001

	// Create random tiebreaker values for each result
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	tiebreakers := make([]float64, len(results))
	for i := range tiebreakers {
		tiebreakers[i] = rng.Float64()
	}

	// Sort by score descending, then by random tiebreaker for approximately equal scores
	sort.SliceStable(results, func(i, j int) bool {
		scoreI := results[i].Score
		scoreJ := results[j].Score

		// Calculate difference
		diff := scoreI - scoreJ

		// If scores differ by more than epsilon, sort by score
		if diff > epsilon {
			return true // i is significantly higher
		}
		if diff < -epsilon {
			return false // j is significantly higher
		}

		// Scores are approximately equal (within epsilon): shuffle randomly
		// This prevents bias toward whichever VDB was queried first
		return tiebreakers[i] > tiebreakers[j]
	})
}

// buildSourceContext creates a SourceContext from a QueryResult
func (cb *ContextBuilder) buildSourceContext(index int, result *vectordb.QueryResult) SourceContext {
	content := cb.extractContent(result.Document)

	// Truncate content if configured
	if cb.config.Output.TruncateSources > 0 && len(content) > cb.config.Output.TruncateSources {
		content = content[:cb.config.Output.TruncateSources] + "..."
	}

	return SourceContext{
		Index:    index,
		Content:  content,
		Score:    result.Score,
		DocID:    result.Document.ID,
		Metadata: result.Document.Metadata,
	}
}

// extractContent extracts the primary content from a document
func (cb *ContextBuilder) extractContent(doc vectordb.Document) string {
	// Priority: Content > Text > Image Metadata > Image URL > URL

	// 1. Text content (existing)
	if doc.Content != "" {
		return doc.Content
	}
	if doc.Text != "" {
		return doc.Text
	}

	// 2. Image metadata (NEW - for multi-modal support)
	if doc.Image != "" || doc.ImageData != "" {
		return cb.extractImageContent(doc)
	}

	// 3. URL fallback (existing)
	if doc.URL != "" {
		return fmt.Sprintf("[Document URL: %s]", doc.URL)
	}

	return "[Empty document]"
}

// extractImageContent extracts content from image documents
// This enables multi-modal RAG by using OCR text, descriptions, and metadata
func (cb *ContextBuilder) extractImageContent(doc vectordb.Document) string {
	var parts []string

	// Check for OCR text in metadata (most important for searchability)
	if ocrText, ok := doc.Metadata["ocr_text"].(string); ok && ocrText != "" {
		parts = append(parts, fmt.Sprintf("Text in image: %s", ocrText))
	}

	// Check for human-provided description
	if desc, ok := doc.Metadata["description"].(string); ok && desc != "" {
		parts = append(parts, fmt.Sprintf("Description: %s", desc))
	}

	// Check for alt text (accessibility)
	if alt, ok := doc.Metadata["alt_text"].(string); ok && alt != "" {
		parts = append(parts, fmt.Sprintf("Alt text: %s", alt))
	}

	// Check for caption
	if caption, ok := doc.Metadata["caption"].(string); ok && caption != "" {
		parts = append(parts, fmt.Sprintf("Caption: %s", caption))
	}

	// Check for tags (important for categorization)
	if tags, ok := doc.Metadata["tags"].([]interface{}); ok && len(tags) > 0 {
		tagStrs := make([]string, 0, len(tags))
		for _, tag := range tags {
			if tagStr, ok := tag.(string); ok {
				tagStrs = append(tagStrs, tagStr)
			}
		}
		if len(tagStrs) > 0 {
			parts = append(parts, fmt.Sprintf("Tags: %s", strings.Join(tagStrs, ", ")))
		}
	}

	// Also handle tags as string slice (some VDBs may store it this way)
	if tagsSlice, ok := doc.Metadata["tags"].([]string); ok && len(tagsSlice) > 0 {
		parts = append(parts, fmt.Sprintf("Tags: %s", strings.Join(tagsSlice, ", ")))
	}

	// Include image URL for reference
	if doc.Image != "" {
		parts = append(parts, fmt.Sprintf("Image URL: %s", doc.Image))
	}

	// If we have content, join it
	if len(parts) > 0 {
		return strings.Join(parts, "\n")
	}

	// Fallback: just indicate it's an image
	if doc.Image != "" {
		return fmt.Sprintf("[Image: %s]", doc.Image)
	}
	if doc.ImageData != "" {
		return "[Image data (base64)]"
	}

	return "[Image document without metadata]"
}

// FormatContextForPrompt formats the context into a string suitable for LLM prompts
func (cb *ContextBuilder) FormatContextForPrompt(context *QueryContext) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Query: %s\n\n", context.Query))

	if len(context.Sources) == 0 {
		builder.WriteString("No relevant sources found.\n")
		return builder.String()
	}

	builder.WriteString(fmt.Sprintf("Retrieved %d relevant sources:\n\n", len(context.Sources)))

	for _, source := range context.Sources {
		builder.WriteString(fmt.Sprintf("[%d] (Score: %.3f)\n", source.Index, source.Score))
		builder.WriteString(fmt.Sprintf("%s\n\n", source.Content))
	}

	return builder.String()
}

// FormatContextForPromptWithCitations formats the context with citation instructions
func (cb *ContextBuilder) FormatContextForPromptWithCitations(context *QueryContext) string {
	var builder strings.Builder

	// Add query
	builder.WriteString(fmt.Sprintf("User Query: %s\n\n", context.Query))

	// Add sources
	if len(context.Sources) == 0 {
		builder.WriteString("No relevant sources found. Please inform the user that no information was found for their query.\n")
		return builder.String()
	}

	builder.WriteString("Retrieved Sources:\n")
	builder.WriteString("==================\n\n")

	for _, source := range context.Sources {
		builder.WriteString(fmt.Sprintf("Source [%d] (Relevance: %.1f%%):\n", source.Index, source.Score*100))
		builder.WriteString(fmt.Sprintf("%s\n", source.Content))

		// Add metadata if configured
		if cb.config.Output.IncludeMetadata && len(source.Metadata) > 0 {
			builder.WriteString(fmt.Sprintf("Metadata: %v\n", source.Metadata))
		}

		builder.WriteString("\n")
	}

	// Add citation instructions based on format
	builder.WriteString("\nInstructions:\n")
	builder.WriteString("=============\n")
	builder.WriteString("When answering the user's query, cite sources using the following format:\n")

	switch cb.config.Response.CitationFormat {
	case "numeric":
		builder.WriteString("- Use [1], [2], etc. to reference sources\n")
		builder.WriteString("- Example: \"According to the documentation [1], the feature was added in version 2.0.\"\n")
	case "author-year":
		builder.WriteString("- Use (Author, Year) format to reference sources\n")
		builder.WriteString("- Example: \"The study (Smith, 2020) found that...\"\n")
	case "footnote":
		builder.WriteString("- Use superscript numbers¹, ², etc. to reference sources\n")
		builder.WriteString("- Example: \"The research¹ indicates that...\"\n")
	}

	// Add strict mode warning if enabled
	if cb.config.Response.StrictMode {
		builder.WriteString("\nSTRICT MODE ENABLED:\n")
		builder.WriteString("- Only use information from the provided sources\n")
		builder.WriteString("- Do not add information from your general knowledge\n")
		builder.WriteString("- If the sources don't contain the answer, clearly state that\n")
		builder.WriteString("- Always cite your sources for every claim\n")
	}

	return builder.String()
}

// GetSourceByIndex retrieves a source by its citation index
func (cb *ContextBuilder) GetSourceByIndex(context *QueryContext, index int) (*SourceContext, error) {
	if index < 1 || index > len(context.Sources) {
		return nil, fmt.Errorf("invalid source index %d (valid range: 1-%d)", index, len(context.Sources))
	}
	return &context.Sources[index-1], nil
}

// GetSourceCount returns the number of sources in the context
func (cb *ContextBuilder) GetSourceCount(context *QueryContext) int {
	return len(context.Sources)
}
