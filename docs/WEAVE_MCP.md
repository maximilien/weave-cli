# Weave MCP Integration Gaps & Enhancement Plan

**Status**: Planning Document
**Created**: 2026-02-04
**Purpose**: Document missing weave-mcp tools and capabilities that need to be added to match weave-cli functionality

---

## Table of Contents

- [Overview](#overview)
- [Gap Analysis](#gap-analysis)
- [Missing Tools](#missing-tools)
- [Tool Updates Required](#tool-updates-required)
- [New Capabilities](#new-capabilities)
- [Implementation Plan](#implementation-plan)

---

## Overview

Weave-mcp provides MCP (Model Context Protocol) server functionality for weave-cli, exposing vector database operations as tools that can be called by AI assistants. This document tracks functionality gaps between weave-cli and weave-mcp that need to be addressed.

**Recent weave-cli enhancements not yet in weave-mcp**:
1. **Structured Logging** (JSON format, log levels)
2. **Observability** (Prometheus metrics, health endpoints)
3. **Agent Framework** (specialized agents for complex tasks)
4. **Schema Management** (enhanced schema operations)
5. **Pipeline Enhancements** (better error handling, progress reporting)

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

### Category 2: Agent Framework

**Current State**:
- weave-mcp exposes individual VDB operations as tools
- No agent abstraction or orchestration

**Desired State**:
- Expose agent capabilities as MCP tools
- Allow AI assistants to leverage specialized agents
- Support for multi-step agent workflows

### Category 3: Schema Management

**Current State**:
- Basic schema operations (create, get)
- No advanced schema features

**Desired State**:
- Schema export/import via MCP
- JSON field inference
- Schema validation and migration
- Schema templates

### Category 4: Pipeline & Batch Operations

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

### 2. Metrics Exposure Tools

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

### 3. Health Check Tools

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

### 4. Agent Execution Tools

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

### 5. Schema Export/Import Tools

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

### 6. Batch Operation Tools

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

### Phase 1: Observability Foundation (Week 1)

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

### Phase 2: Agent Integration (Week 2)

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

### Phase 3: Schema Management (Week 3)

**Tasks**:
1. Add `export_schema` tool with JSON inference
2. Add `import_schema` tool
3. Add schema validation
4. Add schema migration support

**Deliverables**:
- Schema export/import tools
- Schema validation tool
- Schema diff/migration tool

### Phase 4: Batch Operations (Week 4)

**Tasks**:
1. Add `batch_create_documents` tool
2. Add `get_batch_progress` tool
3. Implement async batch processing
4. Add retry logic and error recovery

**Deliverables**:
- Batch operation tools with progress tracking
- Error recovery mechanisms
- Performance optimizations

### Phase 5: Enhanced Error Context & Tool Updates (Week 5)

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

1. `configure_logging` - ❌ Missing
2. `get_metrics` - ❌ Missing
3. `check_health` - ❌ Missing
4. `run_agent` - ❌ Missing
5. `export_schema` - ❌ Missing
6. `import_schema` - ❌ Missing
7. `batch_create_documents` - ❌ Missing
8. `get_batch_progress` - ❌ Missing

**Total**: 8 existing tools to update, 8 new tools to implement
