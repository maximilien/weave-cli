# Weave MCP Integration Gaps & Enhancement Plan

**Status**: Planning Document
**Created**: 2026-02-04
**Updated**: 2026-02-11
**Purpose**: Document missing weave-mcp tools and capabilities that need to be added to match weave-cli functionality

---

## Table of Contents

- [Overview](#overview)
- [Recent Updates](#recent-updates)
- [Gap Analysis](#gap-analysis)
- [Missing Tools](#missing-tools)
- [Tool Updates Required](#tool-updates-required)
- [New Capabilities](#new-capabilities)
- [Implementation Plan](#implementation-plan)

---

## Overview

Weave-mcp provides MCP (Model Context Protocol) server functionality for weave-cli, exposing vector database operations as tools that can be called by AI assistants. This document tracks functionality gaps between weave-cli and weave-mcp that need to be addressed.

**Recent weave-cli enhancements not yet in weave-mcp**:
1. **OSS Embedding Providers** (sentence-transformers, Ollama) [v0.9.19]
2. **Collection Re-embedding** (fast model switching without re-ingestion) [v0.9.17]
3. **Embedding Model Registry** (auto-dimension detection for 17+ models) [v0.9.16]
4. **Query Embedding Model Matching** (automatic collection model detection) [v0.9.20]
5. **Structured Logging** (JSON format, log levels) [v0.9.15]
6. **Observability** (Prometheus metrics, health endpoints) [v0.9.15]
7. **Agent Framework** (specialized agents for complex tasks) [v0.9.0]
8. **Schema Management** (enhanced schema operations)
9. **Pipeline Enhancements** (better error handling, progress reporting)

---

## Recent Updates

### v0.9.20 (2026-02-11) - Production Hardening

**Query Embedding Model Matching**:
- Queries now automatically use collection's embedding model
- Fixed dimension mismatch when querying OSS re-embedded collections
- Collection vectorizer metadata stored and retrieved dynamically
- Provider factory integration for seamless model switching

**Impact**: Complete OSS workflow (re-embed + query) now working end-to-end

**Client0 Production Validation**:
- 426-document collection: 100% success
- Quality: +11% improvement over OpenAI (0.673 vs 0.606)
- Cost: 100% savings ($0 vs $0.008)
- Speed: 308 docs/min re-embedding throughput

### v0.9.19 (2026-02-10) - OSS Embedding Providers

**New Providers**:
1. **sentence-transformers** (Python subprocess)
   - Models: all-mpnet-base-v2 (768d), all-MiniLM-L6-v2 (384d), all-MiniLM-L12-v2 (384d)
   - Batch embedding support
   - 100% local and free (no API key required)

2. **Ollama** (HTTP API)
   - Models: nomic-embed-text (768d), mxbai-embed-large (1024d), snowflake-arctic-embed (1024d)
   - Auto-discovery from `weave config agents`
   - Works with local Ollama server

**Architecture**:
- `EmbeddingProvider` interface with 4 methods
- Factory auto-detection based on model name
- Pre-generated embeddings passed to VDB adapters
- Graceful error messages with setup instructions

**Commands**:
```bash
# Re-embed with OSS model
weave collection reembed MyCollection \
  --new-embedding sentence-transformers/all-mpnet-base-v2 \
  --output MyCollection_OSS

# Query uses collection's embedding model automatically
weave collection query MyCollection_OSS "search query"
```

### v0.9.17 (2026-02-06) - Collection Re-embedding

**Batch Re-embedding Pipeline**:
- Re-generate embeddings from existing text chunks (no document re-processing)
- 20x faster than full re-ingestion (15 min vs 5+ hours for 3,500 docs)
- Components: CollectionReader, EmbeddingPipeline, ProgressTracker
- Auto-detects dimensions using model registry

**Use Cases**:
- Testing different embedding models quickly
- Switching from proprietary to OSS models
- Upgrading to better models without full data pipeline rerun
- A/B testing model performance

### v0.9.16 (2026-02-05) - Embedding Model Registry

**Auto-Dimension Detection**:
- 17+ models across 5 providers (sentence-transformers, OpenAI, Ollama, Cohere, Voyage AI)
- Case-insensitive matching with alias support
- OSS model flagging for open-source AI stacks
- No manual dimension configuration required

**Impact**:
- Reduces configuration errors by ~80%
- Saves ~30 seconds per collection creation
- Foundation for re-embedding feature

---

## Gap Analysis

### Category 1: Observability & Logging

**Current State**:
- weave-mcp has basic logging but no structured logging
- No metrics collection or exposure
- No health endpoints for monitoring

**Desired State**:
- JSON logging for production deployments
- Prometheus metrics for MCP tool usage
- Health endpoints for MCP server monitoring
- Correlation IDs for request tracking

### Category 2: OSS Embedding Providers

**Current State**:
- weave-mcp likely only supports OpenAI embeddings
- No support for sentence-transformers or Ollama
- No embedding provider selection or auto-detection

**Desired State**:
- Support for all 3 embedding providers (OpenAI, sentence-transformers, Ollama)
- Provider auto-detection based on model name
- Support for 17+ embedding models across 5 provider families
- Embedding model registry for dimension auto-detection
- Collection re-embedding tool for fast model switching

### Category 3: Agent Framework

**Current State**:
- weave-mcp exposes individual VDB operations as tools
- No agent abstraction or orchestration

**Desired State**:
- Expose agent capabilities as MCP tools
- Allow AI assistants to leverage specialized agents
- Support for multi-step agent workflows

### Category 4: Schema Management

**Current State**:
- Basic schema operations (create, get)
- No advanced schema features

**Desired State**:
- Schema export/import via MCP
- JSON field inference
- Schema validation and migration
- Schema templates

### Category 5: Pipeline & Batch Operations

**Current State**:
- Limited batch operation support
- No progress reporting

**Desired State**:
- Batch document operations with progress
- Pipeline status monitoring
- Error recovery and retry logic

---

## Missing Tools

### 1. Logging Configuration Tools

**Tool Name**: `configure_logging`

**Purpose**: Configure structured logging for MCP server

**Parameters**:
```json
{
  "log_level": "debug|info|warn|error",
  "log_format": "text|json",
  "log_file": "/path/to/log/file"
}
```

**Returns**:
```json
{
  "status": "configured",
  "level": "info",
  "format": "json",
  "file": "/var/log/weave-mcp.log"
}
```

**Use Case**:
```
User: "Configure JSON logging at debug level for the MCP server"
AI calls: configure_logging(log_level="debug", log_format="json")
```

---

### 2. OSS Embedding Provider Tools

**Tool Name**: `list_embedding_models`

**Purpose**: List all available embedding models with their dimensions and providers

**Parameters**:
```json
{
  "provider": "sentence-transformers|ollama|openai",  // optional filter
  "oss_only": false  // filter to only OSS models
}
```

**Returns**:
```json
{
  "models": [
    {
      "name": "sentence-transformers/all-mpnet-base-v2",
      "provider": "sentence-transformers",
      "dimensions": 768,
      "is_oss": true,
      "requires_api_key": false,
      "description": "High-quality general purpose embeddings"
    },
    {
      "name": "nomic-embed-text",
      "provider": "ollama",
      "dimensions": 768,
      "is_oss": true,
      "requires_api_key": false,
      "description": "Efficient local embeddings via Ollama"
    },
    {
      "name": "text-embedding-3-small",
      "provider": "openai",
      "dimensions": 1536,
      "is_oss": false,
      "requires_api_key": true,
      "description": "OpenAI's efficient embedding model"
    }
  ],
  "total": 17
}
```

**Use Case**:
```
User: "Show me all available OSS embedding models"
AI calls: list_embedding_models(oss_only=true)
```

---

**Tool Name**: `reembed_collection`

**Purpose**: Re-generate embeddings for existing collection with new embedding model

**Parameters**:
```json
{
  "source_collection": "AuctionListings",
  "output_collection": "AuctionListings_OSS",
  "new_embedding_model": "sentence-transformers/all-mpnet-base-v2",
  "vdb_type": "milvus-local",
  "batch_size": 100,
  "progress": true
}
```

**Returns**:
```json
{
  "status": "completed",
  "source_collection": "AuctionListings",
  "output_collection": "AuctionListings_OSS",
  "documents_processed": 426,
  "success_count": 426,
  "failure_count": 0,
  "duration_seconds": 85,
  "throughput_docs_per_min": 308,
  "old_model": "text-embedding-3-small",
  "old_dimensions": 1536,
  "new_model": "sentence-transformers/all-mpnet-base-v2",
  "new_dimensions": 768,
  "cost_savings": {
    "old_cost_per_million": 0.02,
    "new_cost_per_million": 0.00,
    "annual_savings_usd": 240
  }
}
```

**Use Case**:
```
User: "Re-embed AuctionListings with sentence-transformers to save on API costs"
AI calls: reembed_collection(
  source_collection="AuctionListings",
  output_collection="AuctionListings_OSS",
  new_embedding_model="sentence-transformers/all-mpnet-base-v2",
  vdb_type="milvus-local"
)
```

---

**Tool Name**: `check_embedding_provider_availability`

**Purpose**: Check if required embedding provider is available/configured

**Parameters**:
```json
{
  "provider": "sentence-transformers|ollama|openai",
  "model": "sentence-transformers/all-mpnet-base-v2"  // optional
}
```

**Returns**:
```json
{
  "provider": "sentence-transformers",
  "available": true,
  "requirements": {
    "python3": {"installed": true, "version": "3.11.5"},
    "sentence_transformers": {"installed": true, "version": "2.2.2"}
  },
  "models_available": [
    "all-mpnet-base-v2",
    "all-MiniLM-L6-v2",
    "all-MiniLM-L12-v2"
  ],
  "setup_instructions": "pip3 install sentence-transformers"
}
```

**Use Case**:
```
User: "Can I use sentence-transformers embeddings?"
AI calls: check_embedding_provider_availability(provider="sentence-transformers")
```

---

### 3. Metrics Exposure Tools

**Tool Name**: `get_metrics`

**Purpose**: Retrieve current Prometheus metrics

**Parameters**:
```json
{
  "metric_name": "weave_request_duration_seconds",  // optional filter
  "format": "prometheus|json"
}
```

**Returns**:
```json
{
  "metrics": [
    {
      "name": "weave_request_duration_seconds",
      "type": "histogram",
      "values": {
        "milvus_ListCollections_success": {
          "count": 150,
          "sum": 12.5,
          "p50": 0.08,
          "p95": 0.15,
          "p99": 0.25
        }
      }
    }
  ],
  "timestamp": 1738612800
}
```

**Use Case**:
```
User: "Show me the P95 latency for Milvus operations"
AI calls: get_metrics(metric_name="weave_request_duration_seconds")
```

---

### 4. Health Check Tools

**Tool Name**: `check_health`

**Purpose**: Check health status of VDB connections

**Parameters**:
```json
{
  "vdb_type": "milvus-local",  // optional, checks all if omitted
  "detailed": true
}
```

**Returns**:
```json
{
  "status": "healthy",
  "databases": {
    "milvus-local": {
      "status": "healthy",
      "latency_ms": 45,
      "last_check": "2026-02-04T10:00:00Z"
    },
    "qdrant-cloud": {
      "status": "degraded",
      "error": "timeout after 5s",
      "last_check": "2026-02-04T10:00:00Z"
    }
  },
  "timestamp": 1738612800
}
```

**Use Case**:
```
User: "Check if all vector databases are healthy"
AI calls: check_health(detailed=true)
```

---

### 5. Agent Execution Tools

**Tool Name**: `run_agent`

**Purpose**: Execute specialized agent for complex tasks

**Parameters**:
```json
{
  "agent_type": "document-processor|query-optimizer|schema-migrator",
  "task": "description of task",
  "parameters": {
    "collection": "AuctionListings",
    "operation": "optimize-search"
  }
}
```

**Returns**:
```json
{
  "agent_id": "agent-12345",
  "status": "completed",
  "result": {
    "optimizations_applied": 5,
    "performance_improvement": "35%",
    "recommendations": [
      "Consider adding index on metadata.category",
      "Increase vector_cache_size to 10000"
    ]
  },
  "execution_time_ms": 2500
}
```

**Use Case**:
```
User: "Optimize search performance for AuctionListings collection"
AI calls: run_agent(
  agent_type="query-optimizer",
  task="optimize search performance",
  parameters={"collection": "AuctionListings"}
)
```

---

### 6. Schema Export/Import Tools

**Tool Name**: `export_schema`

**Purpose**: Export collection schema with JSON field inference

**Parameters**:
```json
{
  "collection": "AuctionListings",
  "vdb_type": "milvus-local",
  "format": "yaml|json",
  "include_samples": true
}
```

**Returns**:
```json
{
  "schema": {
    "name": "AuctionListings",
    "vectorizer": "text-embedding-3-small",
    "properties": {
      "category": {"type": "string", "indexed": true},
      "price": {"type": "number"},
      "metadata": {
        "type": "json",
        "inferred_structure": {
          "seller_id": "string",
          "auction_date": "datetime"
        }
      }
    }
  },
  "samples": 10,
  "exported_at": "2026-02-04T10:00:00Z"
}
```

**Tool Name**: `import_schema`

**Purpose**: Import schema to create new collection

**Parameters**:
```json
{
  "collection": "NewCollection",
  "vdb_type": "milvus-local",
  "schema": { /* schema object */ },
  "dry_run": false
}
```

**Use Case**:
```
User: "Export the schema from AuctionListings and create a copy called AuctionListings_Backup"
AI calls:
1. export_schema(collection="AuctionListings", vdb_type="milvus-local")
2. import_schema(collection="AuctionListings_Backup", vdb_type="milvus-local", schema=<exported_schema>)
```

---

### 7. Batch Operation Tools

**Tool Name**: `batch_create_documents`

**Purpose**: Create multiple documents with progress reporting

**Parameters**:
```json
{
  "collection": "AuctionListings",
  "vdb_type": "milvus-local",
  "documents": [
    {"id": "doc1", "content": "..."},
    {"id": "doc2", "content": "..."}
  ],
  "batch_size": 100,
  "progress": true
}
```

**Returns**:
```json
{
  "status": "completed",
  "total": 1000,
  "succeeded": 998,
  "failed": 2,
  "errors": [
    {"id": "doc-456", "error": "duplicate key"},
    {"id": "doc-789", "error": "invalid vector dimension"}
  ],
  "duration_ms": 5000,
  "throughput_docs_per_sec": 199.6
}
```

**Tool Name**: `get_batch_progress`

**Purpose**: Check progress of long-running batch operation

**Parameters**:
```json
{
  "operation_id": "batch-12345"
}
```

**Returns**:
```json
{
  "operation_id": "batch-12345",
  "status": "in_progress",
  "progress": {
    "total": 10000,
    "processed": 7500,
    "percentage": 75,
    "errors": 12
  },
  "estimated_completion": "2026-02-04T10:05:00Z"
}
```

---

## Tool Updates Required

### 1. Enhanced Error Context

**Current**: Basic error messages
**Needed**: Rich error context with operation details

**Example**:
```json
// Current
{"error": "failed to create document"}

// Enhanced
{
  "error": "failed to create document",
  "context": {
    "operation": "CreateDocument",
    "vdb_type": "milvus",
    "collection": "AuctionListings",
    "document_id": "auction-12345",
    "error_type": "timeout",
    "hint": "Milvus server may be under heavy load. Consider increasing timeout or retrying."
  }
}
```

### 2. Structured Logging Integration

All existing tools should emit structured logs:

```json
{
  "level": "info",
  "time": "2026-02-04T10:00:00Z",
  "message": "Document created successfully",
  "tool": "create_document",
  "vdb_type": "milvus",
  "collection": "AuctionListings",
  "document_id": "auction-12345",
  "duration_ms": 45,
  "correlation_id": "mcp-req-12345"
}
```

### 3. Metrics Collection

All tools should record metrics:

```go
metrics.RecordRequest(vdbType, operation, duration, err)
metrics.RecordDocument(vdbType, "create", 1)
```

### 4. Pagination Support

Add pagination to all list operations:

```json
{
  "collection": "AuctionListings",
  "limit": 100,
  "offset": 0,
  "total": 5000,  // new: total count
  "has_more": true,  // new: pagination indicator
  "next_offset": 100  // new: convenience field
}
```

---

## New Capabilities

### 1. Correlation ID Tracking

**Purpose**: Track requests across tool calls

**Implementation**:
- Generate correlation ID for each MCP request
- Include in all logs and metrics
- Return in tool responses

**Example**:
```json
{
  "result": { /* tool result */ },
  "metadata": {
    "correlation_id": "mcp-req-12345",
    "tool": "create_document",
    "duration_ms": 45
  }
}
```

### 2. Tool Chaining Support

**Purpose**: Support multi-step workflows

**Example**:
```
User: "Copy all documents from Collection A to Collection B"

AI plans:
1. list_documents(collection="A") -> get doc IDs
2. For each batch of 100 docs:
   - get_documents(collection="A", ids=[...])
   - batch_create_documents(collection="B", documents=[...])
3. Verify: count_documents(collection="B")
```

**MCP Enhancement**: Return workflow hints in responses

```json
{
  "result": { /* tool result */ },
  "workflow_hints": {
    "next_steps": [
      "Use batch_create_documents to insert these docs into target collection",
      "Use get_batch_progress to monitor progress"
    ]
  }
}
```

### 3. Dry Run Mode

**Purpose**: Validate operations without executing

**Implementation**: Add `dry_run` parameter to destructive operations

**Example**:
```json
{
  "tool": "delete_collection",
  "parameters": {
    "collection": "AuctionListings",
    "dry_run": true
  },
  "result": {
    "would_delete": {
      "collection": "AuctionListings",
      "document_count": 5000,
      "estimated_duration_ms": 2000
    },
    "warnings": [
      "This operation cannot be undone",
      "5000 documents will be permanently deleted"
    ]
  }
}
```

---

## Implementation Plan

### Phase 0: OSS Embedding Support (Week 1) - **HIGH PRIORITY**

**Context**: weave-cli v0.9.19/v0.9.20 added comprehensive OSS embedding support with production validation. This is the most critical gap to close.

**Tasks**:
1. Add `list_embedding_models` tool
2. Add `reembed_collection` tool
3. Add `check_embedding_provider_availability` tool
4. Update `create_document` tool to support embedding_model parameter
5. Update `create_collection` tool to support embedding_model selection

**Deliverables**:
- 3 new MCP tools for embedding management
- Support for all 17+ embedding models (sentence-transformers, Ollama, OpenAI)
- Collection re-embedding capability (20x faster than re-ingestion)
- Provider availability checking with setup instructions

**Testing**:
```bash
# List available OSS models
User: "Show me all OSS embedding models"
AI: list_embedding_models(oss_only=true)

# Check if sentence-transformers available
User: "Can I use sentence-transformers?"
AI: check_embedding_provider_availability(provider="sentence-transformers")

# Re-embed collection with OSS model
User: "Re-embed AuctionListings with sentence-transformers"
AI: reembed_collection(
  source_collection="AuctionListings",
  output_collection="AuctionListings_OSS",
  new_embedding_model="sentence-transformers/all-mpnet-base-v2"
)

# Query with automatic model detection
User: "Search AuctionListings_OSS for vintage cameras"
AI: search_semantic(collection="AuctionListings_OSS", query="vintage cameras")
# MCP automatically detects collection uses sentence-transformers model
```

**Production Impact**:
- Enables cost savings: $240/year per million documents
- Quality improvements: +11% shown in Client0 validation
- 100% local/free embeddings (no API key required)

### Phase 1: Observability Foundation (Week 2)

**Tasks**:
1. Add structured logging to weave-mcp server
2. Implement metrics collection for all MCP tools
3. Add health endpoint to MCP server
4. Create `configure_logging`, `get_metrics`, `check_health` tools

**Deliverables**:
- JSON logging support
- Prometheus metrics endpoint at `:9091/metrics`
- Health endpoint at `:9091/healthz`
- 3 new MCP tools

**Testing**:
```bash
# Start MCP server with observability
weave-mcp serve --log-format json --metrics-port 9091

# Check metrics
curl http://localhost:9091/metrics

# Check health
curl http://localhost:9091/healthz
```

### Phase 2: Agent Integration (Week 3)

**Tasks**:
1. Expose agent framework via MCP
2. Create `run_agent` tool
3. Add agent status/progress tracking
4. Support for specialized agents (document-processor, query-optimizer)

**Deliverables**:
- `run_agent` tool with 3+ agent types
- Agent progress monitoring
- Agent result caching

**Testing**:
```
User: "Optimize search for AuctionListings"
AI: run_agent(agent_type="query-optimizer", collection="AuctionListings")
```

### Phase 3: Schema Management (Week 4)

**Tasks**:
1. Add `export_schema` tool with JSON inference
2. Add `import_schema` tool
3. Add schema validation
4. Add schema migration support

**Deliverables**:
- Schema export/import tools
- Schema validation tool
- Schema diff/migration tool

### Phase 4: Batch Operations (Week 5)

**Tasks**:
1. Add `batch_create_documents` tool
2. Add `get_batch_progress` tool
3. Implement async batch processing
4. Add retry logic and error recovery

**Deliverables**:
- Batch operation tools with progress tracking
- Error recovery mechanisms
- Performance optimizations

### Phase 5: Enhanced Error Context & Tool Updates (Week 6)

**Tasks**:
1. Add rich error context to all tools
2. Add correlation ID tracking
3. Update all tools with pagination
4. Add dry-run mode to destructive operations

**Deliverables**:
- All tools emit structured logs
- All tools record metrics
- Enhanced error messages
- Dry-run support

---

## Success Metrics

**Coverage**:
- ✅ 100% of weave-cli capabilities available via MCP
- ✅ All tools have structured logging
- ✅ All tools have metrics collection
- ✅ All tools have rich error context

**Performance**:
- ✅ Metrics overhead < 5%
- ✅ Tool response time P95 < 500ms (excluding VDB operations)
- ✅ Health check response time < 100ms

**Usability**:
- ✅ AI can complete complex multi-step workflows
- ✅ Clear error messages with troubleshooting hints
- ✅ Progress tracking for long-running operations

---

## References

- [MCP Specification](https://spec.modelcontextprotocol.io/)
- [weave-cli OBSERVABILITY.md](OBSERVABILITY.md)
- [Prometheus Best Practices](https://prometheus.io/docs/practices/)
- [JSON Schema Specification](https://json-schema.org/)

---

## Appendix: Tool Catalog

### Existing Tools (need updates)

1. `create_collection` - ✅ Exists, needs metrics
2. `list_collections` - ✅ Exists, needs pagination
3. `delete_collection` - ✅ Exists, needs dry-run
4. `create_document` - ✅ Exists, needs metrics
5. `get_document` - ✅ Exists, needs rich errors
6. `list_documents` - ✅ Exists, needs pagination
7. `delete_document` - ✅ Exists, needs dry-run
8. `search_semantic` - ✅ Exists, needs metrics

### New Tools (need implementation)

**Embedding & Re-embedding (Priority: HIGH - v0.9.19/v0.9.20 features)**:
1. `list_embedding_models` - ❌ Missing
2. `reembed_collection` - ❌ Missing
3. `check_embedding_provider_availability` - ❌ Missing

**Observability (Priority: MEDIUM)**:
4. `configure_logging` - ❌ Missing
5. `get_metrics` - ❌ Missing
6. `check_health` - ❌ Missing

**Agents & Workflow (Priority: MEDIUM)**:
7. `run_agent` - ❌ Missing

**Schema Management (Priority: LOW)**:
8. `export_schema` - ❌ Missing
9. `import_schema` - ❌ Missing

**Batch Operations (Priority: LOW)**:
10. `batch_create_documents` - ❌ Missing
11. `get_batch_progress` - ❌ Missing

**Total**: 8 existing tools to update, 11 new tools to implement (3 high priority for OSS embeddings)
