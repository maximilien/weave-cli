# MCP AI Tools API

This document describes the AI-powered tools available via the Weave MCP (Model Context Protocol) server.

## Overview

The Weave MCP server exposes two AI tools for schema and chunking analysis:
- `suggest_schema` - Analyze documents and suggest optimal collection schemas
- `suggest_chunking` - Analyze documents and suggest optimal chunking configurations

These tools enable AI-assisted configuration through the MCP interface, making them accessible from REPL mode and other MCP clients.

## Starting the MCP Server

```bash
# Start MCP HTTP server on port 8030
cd /path/to/weave-mcp
./bin/weave-mcp -port 8030

# Check server health
curl http://localhost:8030/health
```

## Available Tools

### List All Tools

```bash
curl http://localhost:8030/mcp/tools/list | jq '.tools[] | {name, description}'
```

## suggest_schema Tool

Analyzes sample documents and suggests an optimal vector database schema using AI.

### Input Schema

```json
{
  "source_path": "string (required)",      // Path to documents directory
  "collection_name": "string (required)",  // Name for the collection
  "requirements": "string (optional)",     // User requirements
  "vdb_type": "string (optional)",         // Target VDB type (default: weaviate)
  "max_samples": "integer (optional)"      // Max files to analyze (default: 50)
}
```

### Example Request

```bash
curl -X POST http://localhost:8030/mcp/tools/call \
  -H "Content-Type: application/json" \
  -d '{
    "name": "suggest_schema",
    "arguments": {
      "source_path": "docs",
      "collection_name": "TestDocs",
      "vdb_type": "weaviate",
      "max_samples": 2
    }
  }' | jq .
```

### Example Response

```json
{
  "result": {
    "output": "🔍 Scanning for sample files in: docs\n📊 Found 4 sample files\n🤖 Analyzing documents with AI...\n\n🎯 Confidence: 85.0%\n\n🔧 Suggested Schema:\n   Collection: TestDocs\n   Vector Dimensions: 1536\n   Similarity Metric: cosine\n\n📊 Fields (1):\n   • content              text [indexed] [required]\n..."
  }
}
```

## suggest_chunking Tool

Analyzes sample documents and suggests optimal chunking configuration using AI.

### Input Schema

```json
{
  "source_path": "string (required)",      // Path to documents directory
  "collection_name": "string (required)",  // Name for the collection
  "requirements": "string (optional)",     // User requirements
  "vdb_type": "string (optional)",         // Target VDB type (default: weaviate)
  "max_samples": "integer (optional)"      // Max files to analyze (default: 50)
}
```

### Example Request

```bash
curl -X POST http://localhost:8030/mcp/tools/call \
  -H "Content-Type: application/json" \
  -d '{
    "name": "suggest_chunking",
    "arguments": {
      "source_path": "docs",
      "collection_name": "TestDocs",
      "vdb_type": "weaviate",
      "max_samples": 2
    }
  }' | jq .
```

### Example Response

```json
{
  "result": {
    "output": "🔍 Scanning for sample files in: docs\n📊 Found 4 sample files\n🤖 Analyzing documents with AI...\n\n🎯 Confidence: 85.0%\n\n📊 Recommended Configuration:\n   Chunk Size: 4000 characters (~1000 tokens)\n   Size Range: 2800 - 4800 characters\n   Overlap: 600 characters (~15%)\n   Document Type: mixed\n..."
  }
}
```

## Using from Command Line

### Quick Test with jq

```bash
# Schema suggestion
curl -s -X POST http://localhost:8030/mcp/tools/call \
  -H "Content-Type: application/json" \
  -d '{
    "name": "suggest_schema",
    "arguments": {
      "source_path": "docs",
      "collection_name": "MyDocs",
      "max_samples": 5
    }
  }' | jq -r '.result.output'

# Chunking suggestion
curl -s -X POST http://localhost:8030/mcp/tools/call \
  -H "Content-Type: application/json" \
  -d '{
    "name": "suggest_chunking",
    "arguments": {
      "source_path": "docs",
      "collection_name": "MyDocs",
      "max_samples": 5
    }
  }' | jq -r '.result.output'
```

## Configuration

The AI tools use settings from `weave-agents.yaml`:

```yaml
llm:
  provider: openai
  default_model: "gpt-4o"
  default_temperature: 0.3

schema_agent:
  max_samples: 50

chunking_agent:
  default_chunk_size: 1000
```

Create config with:
```bash
weave config agents         # Current directory
weave config agents --global  # Global ~/.weave-cli
```

## Requirements

- OpenAI API key in `.env`: `OPENAI_API_KEY=sk-...`
- MCP server running on port 8030 (configurable)
- Weave CLI binary in PATH

## Performance

- **Timeout**: 60 seconds for AI operations
- **Max Samples**: 50 files by default (configurable)
- **Concurrency**: 10 concurrent files (configurable)

## See Also

- [AI Agents Configuration](../../configs/weave-agents.yaml)
- [User Guide - AI Features](../USER_GUIDE.md#ai-features)
- [weave-mcp Repository](https://github.com/maximilien/weave-mcp)
