# VectorDB Timeout Strategy Guide

This guide explains the operation-specific timeout strategy implemented in `timeout.go`.

## Overview

The timeout strategy provides intelligent, operation-specific timeouts
that adapt based on:

1. **Deployment type** (local vs cloud)
2. **Operation type** (health check, bulk operation, query, etc.)

This prevents false timeouts on slow operations while maintaining fast
feedback for quick operations.

## Timeout Values

### Local Deployments

| Operation Type | Timeout | Rationale |
|----------------|---------|-----------|
| Health Check | 10s | Expect fast local response |
| Document (single) | 15s | Quick single-doc operations |
| Collection Management | 20s | Schema operations, listing |
| Query/Search | 20s | Vector search is fast locally |
| Schema Operations | 15s | Schema retrieval/validation |
| **Bulk Operations** | **120s (2 min)** | Large batches need time |

### Cloud Deployments

| Operation Type | Timeout | Rationale |
|----------------|---------|-----------|
| Health Check | 20s | Allow for network latency |
| Document (single) | 30s | Network latency + processing |
| Collection Management | 40s | Schema operations can be slow |
| Query/Search | 40s | Complex queries + network |
| Schema Operations | 30s | Schema operations |
| **Bulk Operations** | **300s (5 min)** | Large batches + network |

## Usage Pattern

### 1. Add `getTimeoutFor` Method to Your Adapter/Client

```go
// getTimeoutFor returns an operation-specific timeout based on deployment type
func (a *Adapter) getTimeoutFor(opType vectordb.OperationType) time.Duration {
    isCloud := a.config.Type == vectordb.VectorDBTypeYourDBCloud
    return vectordb.GetTimeoutForOperation(opType, isCloud, a.config.Timeout)
}
```

### 2. Use Operation-Specific Timeouts

**Health Check** (shorter timeout):

```go
func (a *Adapter) Health(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, a.getTimeoutFor(vectordb.OperationTypeHealth))
    defer cancel()
    // ... health check logic
}
```

**Bulk Operations** (extended timeout):

```go
func (a *Adapter) CreateDocuments(ctx context.Context, collectionName string, documents []*vectordb.Document) error {
    ctx, cancel := context.WithTimeout(ctx, a.getTimeoutFor(vectordb.OperationTypeBulk))
    defer cancel()
    // ... bulk operation logic
}
```

**Standard Operations** (use existing `getTimeout()` or update):

```go
func (a *Adapter) CreateCollection(ctx context.Context, name string, schema *CollectionSchema) error {
    ctx, cancel := context.WithTimeout(ctx, a.getTimeoutFor(vectordb.OperationTypeCollection))
    defer cancel()
    // ... collection operation logic
}
```

## Operation Types

Available operation types from `vectordb.OperationType`:

- `OperationTypeHealth` - Health checks, ping operations
- `OperationTypeDocument` - Single document CRUD
  (Create, Get, Update, Delete)
- `OperationTypeCollection` - Collection management
  (Create, Delete, List, Exists, Count)
- `OperationTypeQuery` - Search/query operations
  (Semantic, BM25, Hybrid, Metadata)
- `OperationTypeSchema` - Schema operations (Get, Update, Validate)
- `OperationTypeBulk` - Bulk operations
  (CreateDocuments, DeleteDocuments)

## Custom Timeout Override

Users can still override timeouts via configuration:

```yaml
type: qdrant-local
url: http://localhost:6333
timeout: 60  # Custom timeout in seconds - overrides smart defaults
```

When a custom timeout is set, `GetTimeoutForOperation` uses it for
all operations.

## Implementation Status

### Implemented (OpenSearch - Reference Implementation)

- ✅ Health check (10s local, 20s cloud)
- ✅ CreateDocuments bulk (120s local, 300s cloud)
- ✅ DeleteDocuments bulk (120s local, 300s cloud)

### To Be Implemented (Other VDBs)

- Follow the OpenSearch pattern to add operation-specific timeouts
- Priority: Health checks and bulk operations (highest impact)
- Optional: Migrate other operations gradually

## Benefits

1. **No False Timeouts**: Bulk operations get sufficient time
2. **Fast Feedback**: Health checks fail quickly on connection issues
3. **Network Tolerance**: Cloud deployments account for latency
4. **User Override**: Custom timeouts still respected
5. **Backward Compatible**: Existing `getTimeout()` methods still work

## Migration Guide

### Step 1: Add Helper Method

Add `getTimeoutFor()` method to your adapter (see Usage Pattern above).

### Step 2: Update Critical Operations

Update in priority order:

1. Health checks → `OperationTypeHealth`
2. Bulk operations → `OperationTypeBulk`
3. Collection operations → `OperationTypeCollection`
4. Query operations → `OperationTypeQuery`
5. Document operations → `OperationTypeDocument`
6. Schema operations → `OperationTypeSchema`

### Step 3: Test

```bash
go build ./src/pkg/vectordb/yourdb/...
```

## Future Enhancements

- Add timeout metrics/logging
- Add per-VDB timeout profiles (some VDBs are naturally faster)
- Add timeout auto-tuning based on observed latencies
- Add timeout warnings in CLI output

## Questions?

See `timeout.go` source code for implementation details.
