// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package agents

// GetRAGTemplate returns a RAG agent template configuration
func GetRAGTemplate() *CustomAgentConfig {
	return &CustomAgentConfig{
		Name:        "template",
		Type:        "rag",
		Description: "General-purpose RAG agent that provides detailed answers with citations from retrieved sources",
		Version:     "1.0.0",
		Author:      "",
		LLM: CustomAgentLLMConfig{
			Provider:    "openai",
			Model:       "gpt-4o",
			Temperature: 0.7,
			MaxTokens:   2000,
			TopP:        1.0,
		},
		SystemPrompt: `You are a helpful and knowledgeable AI assistant that provides accurate,
well-researched answers based on the retrieved sources.

Your role is to:
1. Carefully read and understand all provided sources
2. Answer the user's question comprehensively using information from
   the sources
3. Always cite your sources using the numeric citation format
   [1], [2], etc.
4. When sources come from different collections (indicated by _collection
   in metadata), include the collection name in your citations to provide
   full context
5. Synthesize information from multiple sources when relevant
6. Be clear when sources conflict or when information is uncertain
7. If sources don't fully answer the question, acknowledge what's missing

Guidelines:
- Provide detailed, informative responses
- Use clear, professional language
- Structure your answer logically with paragraphs for complex topics
- Cite sources for every factual claim with collection context when available
- Be honest about limitations in the available information`,
		Response: CustomAgentResponseConfig{
			IncludeReferences:  true,
			CitationFormat:     "numeric",
			MaxContextChunks:   5,
			MinRelevanceScore:  0.3,
			DeduplicateSources: true,
			SortByRelevance:    true,
			StrictMode:         false,
		},
		Output: CustomAgentOutputConfig{
			Format:          "markdown",
			IncludeMetadata: true,
			ShowConfidence:  false,
			ShowSources:     true,
			TruncateSources: 500,
		},
		Performance: CustomAgentPerformanceConfig{
			TimeoutSeconds: 60,
			MaxRetries:     2,
			CacheResponses: false,
		},
	}
}

// GetQATemplate returns a QA agent template configuration
func GetQATemplate() *CustomAgentConfig {
	return &CustomAgentConfig{
		Name:        "template",
		Type:        "qa",
		Description: "Precise Q&A agent that provides direct, fact-based answers strictly from sources",
		Version:     "1.0.0",
		Author:      "",
		LLM: CustomAgentLLMConfig{
			Provider:    "openai",
			Model:       "gpt-4o",
			Temperature: 0.3,
			MaxTokens:   1000,
			TopP:        0.95,
		},
		SystemPrompt: `You are a precise, fact-focused Q&A assistant that provides direct
answers based strictly on the provided sources.

Your role is to:
1. Answer the question directly and concisely
2. Use ONLY information explicitly stated in the sources
3. Cite every fact with source numbers [1], [2], etc.
4. If the sources don't contain the answer, clearly state:
   "The provided sources do not contain information to answer this
   question."
5. Avoid speculation or inference beyond what's in the sources
6. Be brief but complete in your answer

Guidelines:
- Prioritize accuracy over completeness
- Start with the direct answer, then add supporting details
- Use the exact terminology from the sources when possible
- If sources conflict, note the discrepancy
- Keep responses under 200 words unless more detail is essential
- Never add information from your general knowledge`,
		UserPromptTemplate: `Question: {{query}}

Retrieved Sources ({{source_count}} found):
{{sources}}

Provide a direct answer based ONLY on the sources above.
Cite all facts with [1], [2], etc.`,
		Response: CustomAgentResponseConfig{
			IncludeReferences:  true,
			CitationFormat:     "numeric",
			MaxContextChunks:   5,
			MinRelevanceScore:  0.5,
			DeduplicateSources: true,
			SortByRelevance:    true,
			StrictMode:         true,
		},
		Output: CustomAgentOutputConfig{
			Format:          "text",
			IncludeMetadata: true,
			ShowConfidence:  true,
			ShowSources:     true,
			TruncateSources: 300,
		},
		Features: CustomAgentFeaturesConfig{
			FactChecking: true,
		},
		Performance: CustomAgentPerformanceConfig{
			TimeoutSeconds: 30,
			MaxRetries:     3,
			CacheResponses: true,
		},
	}
}

// GetSummarizeTemplate returns a summarization agent template configuration
func GetSummarizeTemplate() *CustomAgentConfig {
	return &CustomAgentConfig{
		Name:        "template",
		Type:        "summarize",
		Description: "Document summarization agent that creates concise overviews",
		Version:     "1.0.0",
		Author:      "",
		LLM: CustomAgentLLMConfig{
			Provider:    "openai",
			Model:       "gpt-4o",
			Temperature: 0.5,
			MaxTokens:   1500,
			TopP:        1.0,
		},
		SystemPrompt: `You are a skilled summarization assistant that creates clear,
concise overviews of documents and information.

Your role is to:
1. Identify and extract the main points from the sources
2. Create a well-structured summary with key takeaways
3. Organize information logically (chronological, thematic, etc.)
4. Preserve important details while removing redundancy
5. Maintain the original meaning and context
6. Use clear, accessible language

Guidelines:
- Start with a brief overview (1-2 sentences)
- Use bullet points or numbered lists for key points
- Include relevant statistics, dates, or specific details
- Indicate if sources provide conflicting information
- Keep summaries concise but comprehensive
- Adapt length based on source material complexity`,
		Response: CustomAgentResponseConfig{
			IncludeReferences:  true,
			CitationFormat:     "numeric",
			MaxContextChunks:   10,
			MinRelevanceScore:  0.3,
			DeduplicateSources: true,
			SortByRelevance:    true,
			StrictMode:         false,
		},
		Output: CustomAgentOutputConfig{
			Format:          "markdown",
			IncludeMetadata: false,
			ShowConfidence:  false,
			ShowSources:     true,
			TruncateSources: 400,
		},
		Performance: CustomAgentPerformanceConfig{
			TimeoutSeconds: 45,
			MaxRetries:     2,
			CacheResponses: false,
		},
	}
}

// GetCustomTemplate returns a minimal custom agent template
func GetCustomTemplate() *CustomAgentConfig {
	return &CustomAgentConfig{
		Name:        "template",
		Type:        "custom",
		Description: "Custom agent configuration",
		Version:     "1.0.0",
		Author:      "",
		LLM: CustomAgentLLMConfig{
			Provider:    "openai",
			Model:       "gpt-4o",
			Temperature: 0.7,
			MaxTokens:   2000,
			TopP:        1.0,
		},
		SystemPrompt: "You are a helpful AI assistant.",
		Response: CustomAgentResponseConfig{
			IncludeReferences:  true,
			CitationFormat:     "numeric",
			MaxContextChunks:   5,
			MinRelevanceScore:  0.3,
			DeduplicateSources: true,
			SortByRelevance:    true,
			StrictMode:         false,
		},
		Output: CustomAgentOutputConfig{
			Format:          "markdown",
			IncludeMetadata: true,
			ShowConfidence:  false,
			ShowSources:     true,
			TruncateSources: 500,
		},
		Performance: CustomAgentPerformanceConfig{
			TimeoutSeconds: 30,
			MaxRetries:     2,
			CacheResponses: false,
		},
	}
}
