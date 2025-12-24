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

// ChunkingAgent analyzes documents and suggests optimal chunking strategies
type ChunkingAgent struct {
	llmClient *llm.OpenAIClient
	config    *AgentConfig
}

// ChunkingAnalysisInput represents input for chunking analysis
type ChunkingAnalysisInput struct {
	SampleFiles    []string // Paths to sample documents
	CollectionName string   // Target collection name
	VDBType        string   // Target VDB type
	Requirements   string   // User requirements (optional)
	MaxSamples     int      // Max files to analyze (default: 50)
}

// ChunkingAnalysisOutput represents AI-generated chunking recommendations
type ChunkingAnalysisOutput struct {
	Recommendation ChunkingRecommendation `json:"chunking_advice"`
	Confidence     float64                `json:"confidence"`
	Warnings       []string               `json:"warnings,omitempty"`
	Metrics        ChunkingMetrics        `json:"metrics,omitempty"`
}

// ChunkingRecommendation represents AI-powered chunking size recommendations
type ChunkingRecommendation struct {
	RecommendedSize int      `json:"recommended_size" yaml:"recommended_size"`
	MinSize         int      `json:"min_size" yaml:"min_size"`
	MaxSize         int      `json:"max_size" yaml:"max_size"`
	OverlapSize     int      `json:"overlap_size" yaml:"overlap_size"`
	Reasoning       string   `json:"reasoning" yaml:"reasoning"`
	DocumentType    string   `json:"document_type" yaml:"document_type"`
	Considerations  []string `json:"considerations,omitempty" yaml:"considerations,omitempty"`
}

// ChunkingMetrics represents analyzed chunking characteristics
type ChunkingMetrics struct {
	AvgContentLength int      `json:"avg_content_length" yaml:"avg_content_length"`
	AvgParagraphs    int      `json:"avg_paragraphs" yaml:"avg_paragraphs"`
	AvgSections      int      `json:"avg_sections" yaml:"avg_sections"`
	AvgParagraphLen  int      `json:"avg_paragraph_len" yaml:"avg_paragraph_len"`
	ContentDensity   string   `json:"content_density" yaml:"content_density"`
	FileTypes        []string `json:"file_types" yaml:"file_types"`
}

// NewChunkingAgent creates a new chunking analysis agent
func NewChunkingAgent(llmClient *llm.OpenAIClient) *ChunkingAgent {
	// Load agent config (use defaults if not found)
	config, err := LoadAgentConfig()
	if err != nil {
		// Fall back to defaults on error
		config = defaultConfig()
	}
	return &ChunkingAgent{
		llmClient: llmClient,
		config:    config,
	}
}

// Name returns the agent name
func (a *ChunkingAgent) Name() string {
	return "chunking-agent"
}

// Execute analyzes documents and suggests chunking strategy
func (a *ChunkingAgent) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	analysisInput, ok := input.(*ChunkingAnalysisInput)
	if !ok {
		return nil, fmt.Errorf("invalid input type for ChunkingAgent")
	}

	// Extract samples from files
	schemaAgent := &SchemaAgent{} // Reuse sample extraction
	samples := schemaAgent.extractSamples(analysisInput.SampleFiles, analysisInput.MaxSamples)

	if len(samples) == 0 {
		return nil, fmt.Errorf("no valid samples found for chunking analysis")
	}

	// Analyze chunking metrics
	metrics := a.analyzeChunkingMetrics(samples)

	// Build prompt for LLM
	prompt := a.buildChunkingPrompt(metrics, analysisInput)

	// Call LLM for recommendations (use config values)
	response, err := a.llmClient.Complete(ctx, prompt,
		llm.WithModel(a.config.GetModel("chunking")),
		llm.WithTemperature(a.config.GetTemperature("chunking")),
		llm.WithMaxTokens(a.config.GetMaxTokens("chunking")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze chunking strategy: %w", err)
	}

	// Parse JSON response
	var output ChunkingAnalysisOutput
	if err := json.Unmarshal([]byte(response), &output); err != nil {
		// Try to extract JSON from response
		jsonStart := strings.Index(response, "{")
		jsonEnd := strings.LastIndex(response, "}")
		if jsonStart >= 0 && jsonEnd > jsonStart {
			jsonStr := response[jsonStart : jsonEnd+1]
			if err := json.Unmarshal([]byte(jsonStr), &output); err != nil {
				return nil, fmt.Errorf("failed to parse chunking analysis output: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to parse chunking analysis output: %w", err)
		}
	}

	// Add metrics to output
	output.Metrics = metrics

	return &output, nil
}

// analyzeChunkingMetrics analyzes document samples for chunking characteristics
func (a *ChunkingAgent) analyzeChunkingMetrics(samples []DocumentSample) ChunkingMetrics {
	metrics := ChunkingMetrics{}

	typeMap := make(map[string]bool)
	totalContentLen := 0
	totalParagraphLen := 0
	paragraphCount := 0
	totalParagraphs := 0
	totalSections := 0

	for _, sample := range samples {
		typeMap[sample.Type] = true
		totalContentLen += int(sample.Size)

		// Get content
		content := sample.Preview
		if contentField, ok := sample.Fields["content"]; ok {
			if contentStr, ok := contentField.(string); ok {
				content = contentStr
			}
		}

		// Count paragraphs (separated by double newline or single newline for simple cases)
		paragraphs := strings.Split(content, "\n\n")
		if len(paragraphs) == 1 {
			// Fallback to single newline if no double newlines
			paragraphs = strings.Split(content, "\n")
		}
		totalParagraphs += len(paragraphs)

		// Calculate paragraph lengths
		for _, p := range paragraphs {
			p = strings.TrimSpace(p)
			if len(p) > 0 {
				totalParagraphLen += len(p)
				paragraphCount++
			}
		}

		// Count sections (markdown headers or similar patterns)
		sections := strings.Count(content, "#")
		if sections == 0 {
			// Try other section markers
			sections = strings.Count(content, "===") + strings.Count(content, "---")
		}
		totalSections += sections
	}

	// Calculate averages
	if len(samples) > 0 {
		metrics.AvgContentLength = totalContentLen / len(samples)
		metrics.AvgParagraphs = totalParagraphs / len(samples)
		metrics.AvgSections = totalSections / len(samples)
	}
	if paragraphCount > 0 {
		metrics.AvgParagraphLen = totalParagraphLen / paragraphCount
	}

	// Determine content density
	if metrics.AvgContentLength < 1000 {
		metrics.ContentDensity = "sparse"
	} else if metrics.AvgContentLength < 5000 {
		metrics.ContentDensity = "medium"
	} else {
		metrics.ContentDensity = "dense"
	}

	for fileType := range typeMap {
		metrics.FileTypes = append(metrics.FileTypes, fileType)
	}

	return metrics
}

// buildChunkingPrompt builds the LLM prompt for chunking analysis
func (a *ChunkingAgent) buildChunkingPrompt(metrics ChunkingMetrics, input *ChunkingAnalysisInput) string {
	return fmt.Sprintf(`You are an expert at document chunking strategies for vector databases and RAG systems. Analyze the following document characteristics and suggest an optimal chunking configuration.

Collection: %s
Vector Database: %s
Number of Samples Analyzed: %d
User Requirements: %s

Document Characteristics:
- File Types: %v
- Average Content Length: %d bytes
- Average Paragraphs per Document: %d
- Average Sections per Document: %d
- Average Paragraph Length: %d characters
- Content Density: %s

CHUNKING STRATEGY GUIDELINES:

1. Technical Documentation (code, APIs, specs):
   - Chunk size: 500-1000 tokens (2000-4000 characters)
   - Preserves code blocks and technical explanations
   - Higher overlap (15-20%%) for context continuity
   - Natural boundaries: code blocks, function definitions

2. Narrative Content (articles, stories, books):
   - Chunk size: 1000-2000 tokens (4000-8000 characters)
   - Preserves story flow and narrative context
   - Moderate overlap (10-15%%)
   - Natural boundaries: paragraphs, scene breaks

3. Mixed Content (wikis, knowledge bases):
   - Chunk size: 700-1200 tokens (2800-4800 characters)
   - Balances both technical and narrative
   - Standard overlap (10-15%%)
   - Natural boundaries: sections, headings

4. General Considerations:
   - Use natural boundaries (paragraphs, sections, headers)
   - Balance retrieval accuracy vs context preservation
   - Consider query patterns (specific vs exploratory)
   - Overlap prevents context loss at chunk boundaries
   - Token estimation: ~4 characters per token for English text

IMPORTANT SIZING RULES:
- Recommended size should be in CHARACTERS (not tokens)
- For embedding models with 8k token limits: max ~30,000 characters
- For typical use: 2000-5000 characters is optimal range
- Min/Max should give flexibility for edge cases

Based on the metrics above, provide:
- Optimal chunk size in characters (multiply tokens × 4)
- Minimum and maximum chunk sizes for flexibility
- Overlap size for context continuity (10-20%% of recommended)
- Document type classification
- Specific considerations for this content type

Return ONLY a valid JSON object (no markdown, no explanations outside JSON):
{
  "chunking_advice": {
    "recommended_size": 4000,
    "min_size": 2000,
    "max_size": 6000,
    "overlap_size": 400,
    "reasoning": "Based on average paragraph length of X and medium content density, recommend 4000 characters (~1000 tokens) per chunk. This preserves context while maintaining good retrieval accuracy.",
    "document_type": "technical|narrative|mixed",
    "considerations": [
      "Preserve code blocks",
      "Maintain context across chunks",
      "Use natural paragraph boundaries"
    ]
  },
  "confidence": 0.85,
  "warnings": ["Add warnings if metrics are unusual or concerning"]
}`,
		input.CollectionName,
		input.VDBType,
		len(metrics.FileTypes),
		input.Requirements,
		metrics.FileTypes,
		metrics.AvgContentLength,
		metrics.AvgParagraphs,
		metrics.AvgSections,
		metrics.AvgParagraphLen,
		metrics.ContentDensity,
	)
}
