// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/maximilien/weave-cli/src/pkg/llm"
	"github.com/maximilien/weave-cli/src/pkg/pdf"
)

// SchemaAgent analyzes documents and suggests collection schemas
type SchemaAgent struct {
	llmClient *llm.OpenAIClient
	config    *AgentConfig
}

// SchemaAnalysisInput represents input for schema analysis
type SchemaAnalysisInput struct {
	SampleFiles    []string // Paths to sample documents
	CollectionName string   // Target collection name
	Requirements   string   // User requirements (optional)
	VDBType        string   // Target VDB type
	MaxSamples     int      // Max files to analyze (default: 50)
}

// SchemaAnalysisOutput represents AI-generated schema suggestion
type SchemaAnalysisOutput struct {
	Schema         SchemaConfig            `json:"schema"`
	Reasoning      string                  `json:"reasoning"`
	FieldAnalysis  []FieldSuggestion       `json:"field_analysis"`
	Confidence     float64                 `json:"confidence"`
	Warnings       []string                `json:"warnings"`
	Alternatives   []SchemaConfig          `json:"alternatives,omitempty"`
	ChunkingAdvice *ChunkingRecommendation `json:"chunking_advice,omitempty"`
}

// Note: ChunkingRecommendation is defined in chunking_agent.go

// FieldSuggestion represents analysis for a single field
type FieldSuggestion struct {
	Name        string        `json:"name"`
	Type        string        `json:"type"`
	Indexed     bool          `json:"indexed"`
	Filterable  bool          `json:"filterable"`
	Required    bool          `json:"required"`
	Examples    []interface{} `json:"examples"`
	Frequency   float64       `json:"frequency"`
	Cardinality int           `json:"cardinality"`
	Reasoning   string        `json:"reasoning"`
}

// SchemaConfig represents a complete schema configuration
type SchemaConfig struct {
	CollectionName   string                 `yaml:"collection_name" json:"collection_name"`
	VectorDimensions int                    `yaml:"vector_dimensions" json:"vector_dimensions"`
	SimilarityMetric string                 `yaml:"similarity_metric" json:"similarity_metric"`
	Fields           []FieldConfig          `yaml:"fields" json:"fields"`
	Indexes          []IndexConfig          `yaml:"indexes" json:"indexes"`
	Metadata         map[string]interface{} `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

// FieldConfig represents a field configuration
type FieldConfig struct {
	Name        string `yaml:"name" json:"name"`
	Type        string `yaml:"type" json:"type"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Indexed     bool   `yaml:"indexed" json:"indexed"`
	Filterable  bool   `yaml:"filterable" json:"filterable"`
	Required    bool   `yaml:"required" json:"required"`
}

// IndexConfig represents an index configuration
type IndexConfig struct {
	Name   string   `yaml:"name" json:"name"`
	Fields []string `yaml:"fields" json:"fields"`
	Type   string   `yaml:"type" json:"type"` // "vector", "bm25", "composite"
}

// DocumentStructure represents analyzed document structure
type DocumentStructure struct {
	Samples        []DocumentSample
	FileTypes      []string
	CommonFields   map[string]int    // field -> frequency
	FieldTypes     map[string]string // field -> inferred type
	ContentLengths []int
	Languages      []string
	// Chunking analysis fields
	ParagraphCounts []int  // paragraph count per document
	SectionCounts   []int  // section count per document
	AvgParagraphLen int    // average paragraph length
	ContentDensity  string // "sparse", "medium", "dense"
}

// DocumentSample represents a sampled document
type DocumentSample struct {
	Path    string
	Type    string
	Size    int64
	Fields  map[string]interface{}
	Preview string // First 1000 chars
}

// NewSchemaAgent creates a new schema analysis agent
func NewSchemaAgent(llmClient *llm.OpenAIClient) *SchemaAgent {
	// Load agent config (use defaults if not found)
	config, err := LoadAgentConfig()
	if err != nil {
		// Fall back to defaults on error
		config = defaultConfig()
	}
	return &SchemaAgent{
		llmClient: llmClient,
		config:    config,
	}
}

// Name returns the agent name
func (a *SchemaAgent) Name() string {
	return "schema-agent"
}

// Execute analyzes documents and suggests a schema
func (a *SchemaAgent) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	analysisInput, ok := input.(*SchemaAnalysisInput)
	if !ok {
		return nil, fmt.Errorf("invalid input type for SchemaAgent")
	}

	// Step 1: Sample and extract document content
	samples := a.extractSamples(analysisInput.SampleFiles, analysisInput.MaxSamples)
	if len(samples) == 0 {
		return nil, fmt.Errorf("no valid samples found")
	}

	// Step 2: Analyze document structure
	structure := a.analyzeStructure(samples)

	// Step 3: Use LLM to suggest schema
	prompt := a.buildAnalysisPrompt(structure, analysisInput)

	// Call LLM with JSON mode for structured output (use config values)
	response, err := a.llmClient.Complete(ctx, prompt,
		llm.WithModel(a.config.GetModel("schema")),
		llm.WithTemperature(a.config.GetTemperature("schema")),
		llm.WithMaxTokens(a.config.GetMaxTokens("schema")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze schema: %w", err)
	}

	// Parse JSON response
	var output SchemaAnalysisOutput
	if err := json.Unmarshal([]byte(response), &output); err != nil {
		// If JSON parsing fails, try to extract JSON from response
		jsonStart := strings.Index(response, "{")
		jsonEnd := strings.LastIndex(response, "}")
		if jsonStart >= 0 && jsonEnd > jsonStart {
			jsonStr := response[jsonStart : jsonEnd+1]
			if err := json.Unmarshal([]byte(jsonStr), &output); err != nil {
				return nil, fmt.Errorf("failed to parse schema analysis output: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to parse schema analysis output: %w", err)
		}
	}

	return &output, nil
}

// extractSamples extracts samples from files
func (a *SchemaAgent) extractSamples(files []string, maxSamples int) []DocumentSample {
	var samples []DocumentSample
	count := 0

	for _, file := range files {
		if count >= maxSamples {
			break
		}

		sample := a.extractDocumentSample(file)
		if sample != nil {
			samples = append(samples, *sample)
			count++
		}
	}

	return samples
}

// extractDocumentSample extracts a sample from a single file
func (a *SchemaAgent) extractDocumentSample(filePath string) *DocumentSample {
	fileType := detectFileType(filePath)

	switch fileType {
	case "pdf":
		return a.extractPDFSample(filePath)
	case "json":
		return a.extractJSONSample(filePath)
	case "md", "txt":
		return a.extractTextSample(filePath)
	default:
		return nil
	}
}

// extractPDFSample extracts sample from PDF
func (a *SchemaAgent) extractPDFSample(filePath string) *DocumentSample {
	// Use existing PDF extractor
	textData, _, err := pdf.ExtractPDFContent(filePath, 1000, false, 0, false)
	if err != nil || len(textData) == 0 {
		return nil
	}

	stat, _ := os.Stat(filePath)

	preview := ""
	if len(textData) > 0 {
		preview = textData[0].Content
		if len(preview) > 1000 {
			preview = preview[:1000]
		}
	}

	return &DocumentSample{
		Path:    filePath,
		Type:    "pdf",
		Size:    stat.Size(),
		Fields:  map[string]interface{}{"content": preview},
		Preview: preview,
	}
}

// extractJSONSample extracts sample from JSON
func (a *SchemaAgent) extractJSONSample(filePath string) *DocumentSample {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}

	var fields map[string]interface{}
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil
	}

	stat, _ := os.Stat(filePath)

	preview := string(data)
	if len(preview) > 1000 {
		preview = preview[:1000]
	}

	return &DocumentSample{
		Path:    filePath,
		Type:    "json",
		Size:    stat.Size(),
		Fields:  fields,
		Preview: preview,
	}
}

// extractTextSample extracts sample from text file
func (a *SchemaAgent) extractTextSample(filePath string) *DocumentSample {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}

	stat, _ := os.Stat(filePath)

	preview := string(data)
	if len(preview) > 1000 {
		preview = preview[:1000]
	}

	fileType := "txt"
	if strings.HasSuffix(filePath, ".md") {
		fileType = "md"
	}

	return &DocumentSample{
		Path:    filePath,
		Type:    fileType,
		Size:    stat.Size(),
		Fields:  map[string]interface{}{"content": preview},
		Preview: preview,
	}
}

// analyzeStructure analyzes document structure
func (a *SchemaAgent) analyzeStructure(samples []DocumentSample) DocumentStructure {
	structure := DocumentStructure{
		Samples:      samples,
		CommonFields: make(map[string]int),
		FieldTypes:   make(map[string]string),
	}

	// Analyze file types
	typeMap := make(map[string]bool)
	totalParagraphLen := 0
	paragraphCount := 0

	for _, sample := range samples {
		typeMap[sample.Type] = true
		structure.ContentLengths = append(structure.ContentLengths, int(sample.Size))

		// Count field occurrences
		for fieldName, fieldValue := range sample.Fields {
			structure.CommonFields[fieldName]++

			// Infer field type
			if _, exists := structure.FieldTypes[fieldName]; !exists {
				structure.FieldTypes[fieldName] = inferType(fieldValue)
			}
		}

		// Analyze chunking characteristics
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
		structure.ParagraphCounts = append(structure.ParagraphCounts, len(paragraphs))

		// Calculate average paragraph length
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
		structure.SectionCounts = append(structure.SectionCounts, sections)
	}

	for fileType := range typeMap {
		structure.FileTypes = append(structure.FileTypes, fileType)
	}

	// Calculate average paragraph length
	if paragraphCount > 0 {
		structure.AvgParagraphLen = totalParagraphLen / paragraphCount
	}

	// Determine content density
	avgContentLen := avgLength(structure.ContentLengths)
	if avgContentLen < 1000 {
		structure.ContentDensity = "sparse"
	} else if avgContentLen < 5000 {
		structure.ContentDensity = "medium"
	} else {
		structure.ContentDensity = "dense"
	}

	return structure
}

// buildAnalysisPrompt builds the LLM prompt for schema analysis
func (a *SchemaAgent) buildAnalysisPrompt(structure DocumentStructure, input *SchemaAnalysisInput) string {
	avgParagraphs := 0
	if len(structure.ParagraphCounts) > 0 {
		sum := 0
		for _, c := range structure.ParagraphCounts {
			sum += c
		}
		avgParagraphs = sum / len(structure.ParagraphCounts)
	}

	avgSections := 0
	if len(structure.SectionCounts) > 0 {
		sum := 0
		for _, c := range structure.SectionCounts {
			sum += c
		}
		avgSections = sum / len(structure.SectionCounts)
	}

	return fmt.Sprintf(`You are an expert at designing vector database schemas and document chunking strategies. Analyze the following document samples and suggest an optimal schema AND chunking configuration for a %s collection.

Collection Name: %s
Number of Samples: %d
User Requirements: %s

Document Structure Analysis:
- File Types: %v
- Common Fields: %v
- Field Types: %v
- Average Content Length: %d bytes
- Average Paragraphs per Document: %d
- Average Sections per Document: %d
- Average Paragraph Length: %d characters
- Content Density: %s

Please suggest:
1. A schema configuration that optimizes for search and retrieval
2. A chunking strategy that balances context preservation and retrieval accuracy
3. Appropriate indexes for vector and keyword search
4. Field types that handle the observed data efficiently
5. Vector dimensions appropriate for the content type (1536 for OpenAI embeddings)

CHUNKING CONSIDERATIONS:
- For technical documentation: 500-1000 tokens (preserves code blocks and explanations)
- For narrative content: 1000-2000 tokens (preserves story flow and context)
- For mixed content: 700-1200 tokens (balances both)
- Overlap: 10-20%% of chunk size for context continuity
- Consider document structure (paragraphs, sections) for natural boundaries

Return ONLY a valid JSON object (no markdown, no explanations outside JSON) with this exact structure:
{
  "schema": {
    "collection_name": "%s",
    "vector_dimensions": 1536,
    "similarity_metric": "cosine",
    "fields": [
      {"name": "field_name", "type": "text|number|boolean|datetime", "description": "...", "indexed": true, "filterable": false, "required": true}
    ],
    "indexes": [
      {"name": "index_name", "fields": ["field1"], "type": "vector|bm25|composite"}
    ],
    "metadata": {"key": "value"}
  },
  "reasoning": "Explanation of schema choices...",
  "field_analysis": [
    {"name": "field_name", "type": "text", "indexed": true, "filterable": false, "required": true, "examples": ["ex1"], "frequency": 1.0, "cardinality": 10, "reasoning": "Why this config..."}
  ],
  "confidence": 0.85,
  "warnings": ["Warning messages if any"],
  "alternatives": [],
  "chunking_advice": {
    "recommended_size": 1000,
    "min_size": 500,
    "max_size": 1500,
    "overlap_size": 100,
    "reasoning": "Based on average paragraph length and content density...",
    "document_type": "technical|narrative|mixed",
    "considerations": ["Preserve code blocks", "Maintain context", "Balance retrieval accuracy"]
  }
}`,
		input.VDBType,
		input.CollectionName,
		len(structure.Samples),
		input.Requirements,
		structure.FileTypes,
		structure.CommonFields,
		structure.FieldTypes,
		avgLength(structure.ContentLengths),
		avgParagraphs,
		avgSections,
		structure.AvgParagraphLen,
		structure.ContentDensity,
		input.CollectionName,
	)
}

// detectFileType detects file type from extension
func detectFileType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".pdf":
		return "pdf"
	case ".json":
		return "json"
	case ".md":
		return "md"
	case ".txt":
		return "txt"
	case ".yaml", ".yml":
		return "yaml"
	default:
		return "unknown"
	}
}

// inferType infers field type from value
func inferType(value interface{}) string {
	switch v := value.(type) {
	case string:
		if isDate(v) {
			return "datetime"
		}
		if isURL(v) {
			return "url"
		}
		return "text"
	case int, int64, float64:
		return "number"
	case bool:
		return "boolean"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		return "unknown"
	}
}

// isDate checks if string is a date
func isDate(s string) bool {
	// Simple date pattern check
	return strings.Contains(s, "-") && (len(s) == 10 || len(s) == 19 || len(s) == 20)
}

// isURL checks if string is a URL
func isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

// avgLength calculates average length
func avgLength(lengths []int) int {
	if len(lengths) == 0 {
		return 0
	}
	sum := 0
	for _, l := range lengths {
		sum += l
	}
	return sum / len(lengths)
}
