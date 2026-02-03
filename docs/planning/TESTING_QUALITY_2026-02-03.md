# Testing & Quality Plan - Option C

**Status**: Planned (Future)
**Priority**: Medium (Long-term stability)
**Total Effort**: 25-35 hours (incremental approach)
**Best Time**: After major features stabilize

---

## Overview

Increase test coverage from current ~50% (integration tests only) to 70%+
by adding comprehensive unit tests, E2E tests, and chaos engineering.

---

## Current State

**Integration Tests**: ✅ Excellent
- 10/10 VDBs covered
- Comprehensive CRUD operations
- Real database connections
- ~90% of integration scenarios covered

**Unit Tests**: ❌ Missing
- No package-level unit tests
- All tests in `tests/` directory
- No mocked dependencies
- 0% unit test coverage

**E2E Tests**: ❌ Missing
- No end-to-end workflow tests
- No CLI integration tests
- No multi-step scenarios

**Chaos Tests**: ❌ Missing
- No failure injection
- No resilience testing
- No timeout scenarios

---

## Area 1: Unit Test Coverage (20-25 hours)

### Strategy

Add unit tests to all `src/pkg/vectordb/*/` packages with mocked
dependencies. Focus on error paths, edge cases, timeout handling, input
validation, and data transformations.

### 1.1 Mock Framework Setup (2h)

**Create mock interfaces**:

```go
// tests/mocks/vectordb_client.go
package mocks

import (
    "context"
    "github.com/stretchr/testify/mock"
    "github.com/maximilien/weave-cli/src/pkg/vectordb"
)

type MockVectorDBClient struct {
    mock.Mock
}

func (m *MockVectorDBClient) Health(ctx context.Context) error {
    args := m.Called(ctx)
    return args.Error(0)
}

func (m *MockVectorDBClient) CreateDocument(
    ctx context.Context,
    collection string,
    doc *vectordb.Document,
) error {
    args := m.Called(ctx, collection, doc)
    return args.Error(0)
}

// ... all interface methods
```

**Mock LLM client**:

```go
type MockLLMClient struct {
    mock.Mock
}

func (m *MockLLMClient) GenerateEmbedding(
    ctx context.Context,
    text string,
    model string,
) ([]float32, error) {
    args := m.Called(ctx, text, model)
    return args.Get(0).([]float32), args.Error(1)
}
```

### 1.2 Per-VDB Package Tests (2-3h each × 10 VDBs)

**Test Structure** (example: Qdrant):

```go
// src/pkg/vectordb/qdrant/adapter_test.go
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

    err := adapter.CreateCollection(context.Background(), "test", nil)

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

func TestAdapter_SearchSemantic_ErrorHandling(t *testing.T) {
    mockClient := new(MockQdrantClient)
    mockLLM := new(MockLLMClient)
    adapter := &Adapter{
        client:    mockClient,
        llmClient: mockLLM,
        config:    &vectordb.Config{Timeout: 30},
    }

    // Simulate embedding generation failure
    mockLLM.On("GenerateEmbedding", mock.Anything, "test query", "").
        Return([]float32(nil), errors.New("API rate limit"))

    _, err := adapter.SearchSemantic(
        context.Background(),
        "test",
        "test query",
        nil,
    )

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "API rate limit")
}
```

**Test Categories** (per VDB):

1. **Adapter Initialization**
   - Valid config
   - Invalid config
   - Missing credentials
   - Connection errors

2. **Collection Operations**
   - Create (valid, invalid, timeout)
   - List (empty, multiple, error)
   - Delete (exists, not exists, error)
   - Get info (valid, invalid)

3. **Document Operations**
   - Create (with/without vector, with/without metadata)
   - Get (exists, not exists, error)
   - Update (full, partial, not exists)
   - Delete (exists, not exists, batch)
   - List (pagination, filtering)

4. **Query Operations**
   - Semantic search (valid, empty, error)
   - BM25 search (if supported)
   - Hybrid search (if supported)
   - Metadata filtering

5. **Error Handling**
   - Timeout scenarios
   - Network errors
   - Invalid input
   - API errors

6. **Data Transformations**
   - Document conversion (to/from VDB format)
   - Result parsing
   - Metadata handling
   - Vector normalization

**Implementation Order** (by priority):

1. **Week 1**: Weaviate, Qdrant, Milvus (most used)
2. **Week 2**: Pinecone, Chroma, Supabase
3. **Week 3**: Elasticsearch, OpenSearch, Neo4j, MongoDB

---

## Area 2: E2E Testing (3-5 hours)

### Goal

Test complete workflows from CLI invocation to VDB operations.

### 2.1 E2E Framework Setup (1h)

```go
// tests/e2e/framework.go
package e2e

import (
    "os/exec"
    "testing"
)

func runCLI(t *testing.T, args ...string) string {
    cmd := exec.Command("./bin/weave", args...)
    output, err := cmd.CombinedOutput()
    if err != nil {
        t.Fatalf("CLI failed: %v\nOutput: %s", err, output)
    }
    return string(output)
}

func runCLIExpectError(t *testing.T, args ...string) (string, error) {
    cmd := exec.Command("./bin/weave", args...)
    output, err := cmd.CombinedOutput()
    return string(output), err
}

func randomID() string {
    return fmt.Sprintf("%d", time.Now().UnixNano())
}
```

### 2.2 E2E Test Scenarios (2-4h)

**Scenario 1: Document Lifecycle**

```go
func TestE2E_DocumentLifecycle(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping E2E test in short mode")
    }

    vdbType := "qdrant-cloud"
    collection := "e2e-test-" + randomID()

    // Create collection
    runCLI(t, "cols", "create", collection, "--vdb", vdbType)

    // Create document
    docID := runCLI(t, "docs", "create", collection, "test.txt", "--vdb", vdbType)

    // Search for document
    results := runCLI(t, "search", collection, "test query", "--vdb", vdbType)
    assert.Contains(t, results, docID)

    // Get document
    doc := runCLI(t, "docs", "get", collection, docID, "--vdb", vdbType)
    assert.NotEmpty(t, doc)

    // Delete document
    runCLI(t, "docs", "delete", collection, docID, "--vdb", vdbType)

    // Verify deletion
    _, err := runCLIExpectError(t, "docs", "get", collection, docID, "--vdb", vdbType)
    assert.Error(t, err)

    // Cleanup
    runCLI(t, "cols", "delete", collection, "--vdb", vdbType)
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
    output := runCLI(t, "pipeline", "ingest", testDir,
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
    runCLI(t, "cols", "delete", collection, "--vdb", "qdrant-cloud")
}
```

**Scenario 3: Multi-VDB Operations**

```go
func TestE2E_MultiVDBReplication(t *testing.T) {
    // Create collection in Qdrant
    collection := "multi-vdb-" + randomID()
    runCLI(t, "cols", "create", collection, "--vdb", "qdrant-cloud")

    // Add documents to Qdrant
    for i := 0; i < 10; i++ {
        runCLI(t, "docs", "create", collection,
            fmt.Sprintf("doc-%d.txt", i),
            "--vdb", "qdrant-cloud")
    }

    // Replicate to Weaviate
    runCLI(t, "replicate", collection,
        "--from", "qdrant-cloud",
        "--to", "weaviate-cloud")

    // Verify in Weaviate
    results := runCLI(t, "docs", "ls", collection, "--vdb", "weaviate-cloud")
    assert.Contains(t, results, "10 documents")

    // Cleanup both
    runCLI(t, "cols", "delete", collection, "--vdb", "qdrant-cloud")
    runCLI(t, "cols", "delete", collection, "--vdb", "weaviate-cloud")
}
```

---

## Area 3: Chaos Engineering (2-3 hours)

### Goal

Test system behavior under failure conditions.

### 3.1 Network Failure Tests (1h)

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

### 3.2 Partial Failure Tests (0.5h)

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
    // Should continue processing valid docs
}
```

### 3.3 Resource Exhaustion Tests (0.5h)

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

---

## Implementation Timeline

**Week 1**: Unit Tests (Tier 1 VDBs)
- Days 1-2: Qdrant, Weaviate unit tests (5h)
- Days 3-4: Milvus unit tests (2.5h)

**Week 2**: Unit Tests (Tier 2 VDBs)
- Days 1-2: Pinecone, Chroma unit tests (5h)
- Days 3-4: Supabase unit tests (2.5h)

**Week 3**: Unit Tests (Tier 3 VDBs) + E2E
- Days 1-2: Elasticsearch, OpenSearch unit tests (5h)
- Days 3-4: Neo4j, MongoDB unit tests (5h)
- Day 5: E2E tests (3-5h)

**Week 4**: Chaos + Cleanup
- Days 1-2: Chaos engineering tests (2-3h)
- Days 3-5: Fix any issues, documentation

---

## Success Metrics

- ✅ Unit test coverage: 70%+ in all VDB packages
- ✅ Integration tests: Still passing (maintain quality)
- ✅ E2E tests: 5+ critical workflows covered
- ✅ Chaos tests: 10+ failure scenarios tested
- ✅ CI pipeline: All tests passing
- ✅ Test execution time: < 10 minutes for full suite

---

## Tools & Dependencies

```bash
# Testing frameworks
go get github.com/stretchr/testify/assert
go get github.com/stretchr/testify/mock
go get github.com/stretchr/testify/suite

# Coverage tools
go install github.com/axw/gocov/gocov@latest
go install github.com/AlekSi/gocov-xml@latest
```

---

## References

- [Option 4 Detailed Plan](OPTION_4_TESTING_QUALITY.md)
- [Go Testing Best Practices](https://go.dev/doc/tutorial/add-a-test)
- [Testify Documentation](https://github.com/stretchr/testify)
