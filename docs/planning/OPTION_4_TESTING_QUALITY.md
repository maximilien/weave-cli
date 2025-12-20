# Option 4: Testing & Quality - Detailed Implementation Plan

**Status**: Planning
**Priority**: ⭐ Medium (quality focus)
**Total Effort**: 25-35 hours
**Target**: Weeks 6-7

---

## Overview

Increase test coverage from current ~50% (integration tests only) to 70%+ by adding unit tests, E2E tests, and chaos engineering.

**Current State:**
- ✅ Integration tests: Excellent (10/10 VDBs, comprehensive)
- ❌ Unit tests: None (all in `tests/`, not in package dirs)
- ❌ E2E tests: None
- ❌ Chaos tests: None

**Goal State:**
- ✅ Integration tests: Maintain current quality
- ✅ Unit tests: 70%+ coverage in package directories
- ✅ E2E tests: Critical workflows covered
- ✅ Chaos tests: Failure scenarios tested

---

## Area 1: Unit Test Coverage (20-25 hours)

### Goal
Add unit tests to all `src/pkg/vectordb/*/` packages with mocked dependencies.

### Strategy
Focus on:
1. Error paths
2. Edge cases
3. Timeout handling
4. Input validation
5. Data transformations

### Implementation Pattern

**Mock VectorDBClient:**
```go
// tests/mocks/vectordb_client.go
type MockVectorDBClient struct {
    mock.Mock
}

func (m *MockVectorDBClient) Health(ctx context.Context) error {
    args := m.Called(ctx)
    return args.Error(0)
}

func (m *MockVectorDBClient) CreateDocument(ctx context.Context, collection string, doc *vectordb.Document) error {
    args := m.Called(ctx, collection, doc)
    return args.Error(0)
}

// ... implement all interface methods
```

**Example Unit Tests:**

```go
// src/pkg/vectordb/qdrant/collection_test.go
package qdrant

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestAdapter_CreateCollection_ValidInput(t *testing.T) {
    mockClient := new(MockQdrantClient)
    adapter := &Adapter{
        client: mockClient,
        config: &vectordb.Config{Timeout: 30},
    }

    mockClient.On("CreateCollection", mock.Anything, "test", mock.Anything).
        Return(nil)

    err := adapter.CreateCollection(context.Background(), "test", nil)
    assert.NoError(t, err)
    mockClient.AssertExpectations(t)
}

func TestAdapter_CreateCollection_EmptyName(t *testing.T) {
    adapter := &Adapter{}
    
    err := adapter.CreateCollection(context.Background(), "", nil)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "collection name is required")
}

func TestAdapter_CreateCollection_Timeout(t *testing.T) {
    mockClient := new(MockQdrantClient)
    adapter := &Adapter{
        client: mockClient,
        config: &vectordb.Config{Timeout: 1}, // 1 second
    }

    // Simulate slow operation
    mockClient.On("CreateCollection", mock.Anything, "test", mock.Anything).
        Run(func(args mock.Arguments) {
            time.Sleep(2 * time.Second)
        }).
        Return(nil)

    ctx := context.Background()
    err := adapter.CreateCollection(ctx, "test", nil)
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestAdapter_CreateDocument_GeneratesEmbedding(t *testing.T) {
    mockClient := new(MockQdrantClient)
    mockLLM := new(MockLLMClient)
    adapter := &Adapter{
        client:    mockClient,
        llmClient: mockLLM,
        config:    &vectordb.Config{Timeout: 30},
    }

    doc := &vectordb.Document{
        ID:   "doc-1",
        Text: "Test document",
        // No vector provided
    }

    expectedEmbedding := []float32{0.1, 0.2, 0.3}
    mockLLM.On("GenerateEmbedding", mock.Anything, "Test document", "").
        Return(expectedEmbedding, nil)
    mockClient.On("Upsert", mock.Anything, mock.Anything).
        Return(nil)

    err := adapter.CreateDocument(context.Background(), "test", doc)
    assert.NoError(t, err)
    assert.Equal(t, expectedEmbedding, doc.Vector)
}
```

### Test Coverage by Package

**Per VDB Package (2-3 hours each):**
- Adapter tests (initialization, config)
- Collection tests (CRUD, validation)
- Document tests (CRUD, embedding generation)
- Query tests (search, filtering, result conversion)
- Schema tests (validation, defaults)
- Error handling tests
- Timeout tests

**10 VDBs × 2.5 hours = 25 hours**

### Success Criteria
- ✅ 70%+ coverage in each VDB package
- ✅ All error paths tested
- ✅ Timeout scenarios covered
- ✅ Input validation tested
- ✅ Mock objects properly isolated

---

## Area 2: E2E Testing (3-5 hours)

### Goal
Test complete workflows from CLI invocation to VDB operations.

### Test Scenarios

**Scenario 1: Document Lifecycle**
```go
func TestE2E_DocumentLifecycle(t *testing.T) {
    // Setup
    vdbType := "qdrant-cloud"
    collection := "test-" + randomID()
    
    // Create collection
    runCLI(t, "weave", "cols", "create", collection, "--vdb", vdbType)
    
    // Create document
    docID := runCLI(t, "weave", "docs", "create", collection, "test.txt", "--vdb", vdbType)
    
    // Search for document
    results := runCLI(t, "weave", "search", collection, "test query", "--vdb", vdbType)
    assert.Contains(t, results, docID)
    
    // Get document
    doc := runCLI(t, "weave", "docs", "get", collection, docID, "--vdb", vdbType)
    assert.NotEmpty(t, doc)
    
    // Delete document
    runCLI(t, "weave", "docs", "delete", collection, docID, "--vdb", vdbType)
    
    // Verify deletion
    _, err := runCLIExpectError(t, "weave", "docs", "get", collection, docID, "--vdb", vdbType)
    assert.Error(t, err)
    
    // Cleanup
    runCLI(t, "weave", "cols", "delete", collection, "--vdb", vdbType)
}
```

**Scenario 2: Pipeline Ingestion**
```go
func TestE2E_PipelineIngestion(t *testing.T) {
    // Create test directory with documents
    testDir := createTestDocuments(t, 100)
    defer os.RemoveAll(testDir)
    
    collection := "pipeline-test-" + randomID()
    
    // Run pipeline ingestion
    output := runCLI(t, "weave", "pipeline", "ingest", testDir,
        "--collection", collection,
        "--vdb", "qdrant-cloud",
        "--output", "json")
    
    // Parse report
    var report IngestReport
    json.Unmarshal([]byte(output), &report)
    
    // Verify
    assert.Equal(t, 100, report.FilesProcessed)
    assert.Greater(t, report.DocumentsCreated, 0)
    
    // Cleanup
    runCLI(t, "weave", "cols", "delete", collection, "--vdb", "qdrant-cloud")
}
```

**Scenario 3: Multi-VDB Replication**
```go
func TestE2E_MultiVDBReplication(t *testing.T) {
    // Create collection in Qdrant
    collection := "multi-vdb-" + randomID()
    runCLI(t, "weave", "cols", "create", collection, "--vdb", "qdrant-cloud")
    
    // Add documents to Qdrant
    for i := 0; i < 10; i++ {
        runCLI(t, "weave", "docs", "create", collection,
            fmt.Sprintf("doc-%d.txt", i),
            "--vdb", "qdrant-cloud")
    }
    
    // Replicate to Weaviate
    runCLI(t, "weave", "replicate", collection,
        "--from", "qdrant-cloud",
        "--to", "weaviate-cloud")
    
    // Verify in Weaviate
    results := runCLI(t, "weave", "docs", "ls", collection, "--vdb", "weaviate-cloud")
    assert.Contains(t, results, "10 documents")
    
    // Cleanup both
    runCLI(t, "weave", "cols", "delete", collection, "--vdb", "qdrant-cloud")
    runCLI(t, "weave", "cols", "delete", collection, "--vdb", "weaviate-cloud")
}
```

### Effort: 3-5 hours
- E2E framework setup: 1 hour
- Test scenarios (3-5): 2-4 hours

---

## Area 3: Chaos Engineering (2-3 hours)

### Goal
Test system behavior under failure conditions.

### Test Scenarios

**Network Failures:**
```go
func TestChaos_NetworkTimeout(t *testing.T) {
    // Configure aggressive timeout
    os.Setenv("QDRANT_TIMEOUT", "1")
    
    // Simulate slow network with proxy
    proxy := StartSlowProxy(t, "localhost:6333", 5*time.Second)
    defer proxy.Stop()
    
    // Attempt operation
    err := client.CreateDocument(ctx, collection, doc)
    
    // Verify graceful timeout
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "timeout")
}

func TestChaos_ConnectionRefused(t *testing.T) {
    // Point to non-existent server
    os.Setenv("QDRANT_URL", "http://localhost:9999")
    
    client, err := qdrant.NewAdapter(config)
    assert.NoError(t, err)
    
    // Attempt health check
    err = client.Health(ctx)
    
    // Verify error message contains troubleshooting hints
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "connection refused")
    assert.Contains(t, err.Error(), "Common causes")
}
```

**Partial Failures:**
```go
func TestChaos_PartialBatchFailure(t *testing.T) {
    // Create batch with some invalid documents
    docs := []*vectordb.Document{
        {ID: "valid-1", Text: "Valid document"},
        {ID: "", Text: "Invalid - no ID"}, // Will fail
        {ID: "valid-2", Text: "Valid document"},
    }
    
    // Execute batch create
    err := client.CreateDocuments(ctx, collection, docs)
    
    // Verify partial success handling
    // (Implementation should continue processing valid docs)
}
```

**Resource Exhaustion:**
```go
func TestChaos_LargePayload(t *testing.T) {
    // Create very large document
    largeText := strings.Repeat("x", 10*1024*1024) // 10MB
    
    doc := &vectordb.Document{
        ID:   "large-doc",
        Text: largeText,
    }
    
    // Attempt to create
    err := client.CreateDocument(ctx, collection, doc)
    
    // Verify appropriate error
    assert.Error(t, err)
}
```

### Effort: 2-3 hours
- Network failure tests: 1 hour
- Partial failure tests: 0.5 hour
- Resource tests: 0.5 hour
- Framework setup: 0.5 hour

---

## Implementation Timeline

**Week 6: Unit Tests (20-25 hours)**
- Days 1-2: Qdrant, Weaviate unit tests (5h)
- Days 3-4: Milvus, Pinecone, Chroma unit tests (7.5h)
- Day 5: Elasticsearch, OpenSearch unit tests (5h)

**Week 7: Unit Tests + E2E + Chaos (8-13 hours)**
- Days 1-2: Neo4j, MongoDB, Supabase unit tests (7.5h)
- Day 3-4: E2E tests (3-5h)
- Day 5: Chaos engineering tests (2-3h)

**Total: 25-35 hours across 2 weeks**

---

## Success Metrics

- ✅ Unit test coverage: 70%+ in all VDB packages
- ✅ Integration tests: Still passing (maintain quality)
- ✅ E2E tests: 5+ critical workflows covered
- ✅ Chaos tests: 10+ failure scenarios tested
- ✅ CI pipeline: All tests passing
- ✅ Test execution time: < 10 minutes for full suite
