# RAG Agents Configuration

This directory contains RAG (Retrieval-Augmented Generation) agent
configurations for the Weave CLI. Agents process vector database query results
to provide comprehensive answers, summaries, or precise question-answering with
citations.

## Overview

Agents enhance query results by:

- Providing natural language answers with citations
- Summarizing retrieved documents
- Answering specific questions based on retrieved context
- Formatting responses in various output formats (Markdown, text, JSON)

## Available Agents

### rag-agent

General-purpose RAG agent that provides detailed, comprehensive answers with citations.

**Best for:**

- Exploratory questions
- Complex topics requiring synthesis
- When you want detailed explanations

**Features:**

- Numeric citations [1], [2], etc.
- Synthesizes information from multiple sources
- Provides detailed responses
- Markdown output format

**Usage:**

```bash
weave cols query MyDocs "What is machine learning?" --agent rag-agent
weave cols query MyDocs "Explain neural networks" --agent rag-agent --top_k 10
```

### qa-agent

Precise question-answering agent with strict source adherence.

**Best for:**

- Specific factual questions
- When you need concise, direct answers
- Fact-checking
- Quick lookups

**Features:**

- Strict mode: only uses information from sources
- Brief, direct answers
- Higher relevance threshold (0.5)
- Text output format
- Confidence scoring

**Usage:**

```bash
weave cols query MyDocs "What year was it founded?" --agent qa-agent
weave cols q MyDocs "Who is the CEO?" --agent qa-agent --top_k 5
```

### summarize-agent

Document summarization agent for creating concise overviews.

**Best for:**

- Long documents
- Topic overviews
- Research synthesis
- Quick scanning

**Features:**

- Handles more sources (up to 10)
- Lower relevance threshold (0.2)
- Structured bullet points
- No individual citations (summary format)
- Markdown output

**Usage:**

```bash
weave cols query MyDocs "AI research trends" --agent summarize-agent --top_k 10
weave cols q MyDocs "Recent developments" --agent summarize-agent
```

## Agent Configuration Format

Agents are configured using YAML files with the following structure:

```yaml
---
name: my-agent                    # Agent name (lowercase, alphanumeric, hyphens)
type: rag                         # Agent type: rag, summarize, qa, custom
description: "Agent description"
version: "1.0.0"
author: "Your Name"

# LLM Configuration
llm:
  provider: openai               # LLM provider
  model: "gpt-4o"                # Model to use
  temperature: 0.7               # Temperature (0.0-2.0)
  max_tokens: 2000               # Max response tokens
  top_p: 1.0                     # Top-p sampling

# System Prompt
system_prompt: |
  You are a helpful assistant...

# Optional: Custom user prompt template
user_prompt_template: |
  Query: {{query}}
  Sources: {{sources}}
  Count: {{source_count}}

# Response Configuration
response:
  include_references: true       # Include source citations
  citation_format: "numeric"     # Citation format: numeric, author-year, footnote
  max_context_chunks: 5          # Max sources to use
  min_relevance_score: 0.3       # Min relevance threshold
  deduplicate_sources: true      # Remove duplicate sources
  sort_by_relevance: true        # Sort by relevance score
  strict_mode: false             # Only use source information

# Output Configuration
output:
  format: "markdown"             # Output format: markdown, text, json
  include_metadata: false        # Include metadata in output
  show_confidence: false         # Show confidence score
  show_sources: true             # Show source list
  truncate_sources: 500          # Truncate source content (chars)

# Features
features:
  streaming: false               # Stream responses
  multi_turn: false              # Multi-turn conversations
  fact_checking: false           # Enable fact checking

# Performance
performance:
  timeout_seconds: 60            # Query timeout
  max_retries: 2                 # Max retry attempts
  cache_responses: false         # Cache responses
```

## Creating Custom Agents

### 1. Create Agent YAML File

Create a new YAML file in one of the search paths:

- `configs/agents/` (project-specific)
- `~/.weave-cli/agents/` (user-specific)
- `/etc/weave-cli/agents/` (system-wide)

### 2. Define Agent Configuration

```yaml
---
name: my-custom-agent
type: custom
description: "My custom RAG agent"
version: "1.0.0"

llm:
  model: "gpt-4o"
  temperature: 0.7

system_prompt: |
  You are a specialized assistant for...

response:
  include_references: true
  max_context_chunks: 5
```

### 3. Validate Configuration

```bash
weave agents validate configs/agents/my-custom-agent.yaml
```

### 4. Test Your Agent

```bash
weave cols query MyDocs "test query" --agent my-custom-agent
```

## Agent Management Commands

### List Available Agents

```bash
# List all agents
weave agents list

# List in JSON format
weave agents list --output json
```

### Show Agent Details

```bash
# Show detailed configuration
weave agents show rag-agent

# Show in YAML format
weave agents show qa-agent --output yaml
```

### Validate Agent

```bash
# Validate configuration file
weave agents validate configs/agents/my-agent.yaml
```

## Search Paths

Agents are loaded from these locations in order:

1. `configs/agents/` - Project-specific agents
2. `~/.weave-cli/agents/` - User-specific agents
3. `/etc/weave-cli/agents/` - System-wide agents

If an agent with the same name exists in multiple locations, the first one
found is used.

## Environment Requirements

To use agents, you must set:

```bash
export OPENAI_API_KEY="your-api-key-here"
```

Or add to your `.env` file:

```bash
OPENAI_API_KEY=your-api-key-here
```

## Configuration Reference

### Agent Types

- `rag` - General-purpose retrieval-augmented generation
- `summarize` - Document summarization
- `qa` - Precise question-answering
- `custom` - Custom agent behavior

### Citation Formats

- `numeric` - [1], [2], [3]
- `author-year` - (Smith, 2020), (Jones, 2021)
- `footnote` - ¹, ², ³

### Output Formats

- `markdown` - Formatted markdown with headers
- `text` - Plain text output
- `json` - Structured JSON output

## Best Practices

1. **Choose the Right Agent**
   - Use `rag-agent` for comprehensive answers
   - Use `qa-agent` for quick facts
   - Use `summarize-agent` for overviews

2. **Tune Parameters**
   - Increase `top_k` for better context
   - Adjust `min_relevance_score` to filter results
   - Set `max_context_chunks` based on response length needs

3. **Use Strict Mode**
   - Enable `strict_mode` for factual accuracy
   - Disable for creative synthesis

4. **Customize System Prompts**
   - Tailor prompts to your domain
   - Include specific instructions
   - Define output format expectations

## Examples

### Complex Research Query

```bash
weave cols query Research "impact of climate change on agriculture" \
  --agent rag-agent \
  --top_k 10
```

### Quick Fact Lookup

```bash
weave cols query Docs "What is the API endpoint?" \
  --agent qa-agent \
  --top_k 3
```

### Document Summary

```bash
weave cols query Papers "recent AI developments" \
  --agent summarize-agent \
  --top_k 15
```

### Custom Agent with JSON Output

Create `custom-json-agent.yaml`:

```yaml
name: custom-json-agent
type: custom
description: "Agent that outputs JSON"

llm:
  model: "gpt-4o"
  temperature: 0.3

system_prompt: "You are a precise assistant that outputs structured JSON."

response:
  include_references: true
  strict_mode: true

output:
  format: "json"
  include_metadata: true
  show_confidence: true
```

Use it:

```bash
weave cols query MyDocs "query" --agent custom-json-agent
```

## Troubleshooting

### Agent Not Found

```bash
# List available agents
weave agents list

# Check search paths
ls configs/agents/
ls ~/.weave-cli/agents/
```

### Validation Errors

```bash
# Validate your agent config
weave agents validate configs/agents/my-agent.yaml
```

### API Key Issues

```bash
# Check if API key is set
echo $OPENAI_API_KEY

# Or check .env file
cat .env | grep OPENAI_API_KEY
```

## Additional Resources

- [Agent Planning Document](../../docs/planning/RAG_AGENT_FEATURE.md)
- [Vector Database Configuration](../README.md)
- [Weave CLI Documentation](../../README.md)

## Support

For issues or questions:

- GitHub Issues: <https://github.com/maximilien/weave-cli/issues>
- Documentation: <https://github.com/maximilien/weave-cli/tree/main/docs>
