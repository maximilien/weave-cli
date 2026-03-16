# Parallel Batch Processing Design

**Author**: Claude Code
**Date**: March 16, 2026
**Status**: Draft / Design Phase
**Sprint**: Sprint 3
**Related**: Performance profiling (March 15, 2026)

---

## Executive Summary

Design document for implementing parallel batch processing in backup operations to achieve 2-3x performance improvement beyond the Quick Win (batch size 200).

**Expected Impact**: 750-1000 docs/sec (from current 376 docs/sec)

---

## Problem Statement

### Current Implementation (Sequential)

```go
for offset := 0; offset < totalDocs; offset += batchSize {
    docs, err := vdbClient.ListDocuments(ctx, collection, batchSize, offset)
    if err != nil {
        return err
    }
    backup.Documents = append(backup.Documents, docs...)
}
```

**Bottleneck**: Batches are fetched sequentially, one at a time.

For 301 documents with batch=200:
- Batch 1: offset=0, fetch 200 docs → wait
- Batch 2: offset=200, fetch 101 docs → wait
- **Total**: 2 sequential queries

**Opportunity**: Fetch both batches concurrently!

---

## Proposed Solution

### High-Level Architecture

```go
// 1. Calculate all batch offsets upfront
numBatches := (totalDocs + batchSize - 1) / batchSize
batches := make([]int, numBatches)
for i := 0; i < numBatches; i++ {
    batches[i] = i * batchSize
}

// 2. Fetch batches concurrently with worker pool
results := make(chan batchResult, numBatches)
var wg sync.WaitGroup

for _, offset := range batches {
    wg.Add(1)
    go func(off int) {
        defer wg.Done()
        docs, err := vdbClient.ListDocuments(ctx, collection, batchSize, off)
        results <- batchResult{offset: off, docs: docs, err: err}
    }(offset)
}

// 3. Close results channel when all workers done
go func() {
    wg.Wait()
    close(results)
}()

// 4. Collect results (need to sort by offset to maintain order)
batchResults := make([]batchResult, 0, numBatches)
for result := range results {
    if result.err != nil {
        return fmt.Errorf("batch fetch failed at offset %d: %w", result.offset, result.err)
    }
    batchResults = append(batchResults, result)
}

// 5. Sort by offset to maintain document order
sort.Slice(batchResults, func(i, j int) bool {
    return batchResults[i].offset < batchResults[j].offset
})

// 6. Append documents in order
for _, batch := range batchResults {
    backup.Documents = append(backup.Documents, batch.docs...)
}
```

---

## Detailed Design

### 1. Data Structures

```go
// batchResult holds the result of a single batch fetch
type batchResult struct {
    offset int                    // Starting offset of this batch
    docs   []vectordb.Document    // Documents fetched
    err    error                  // Error if fetch failed
}
```

### 2. Worker Pool Pattern

**Option A: Simple goroutines (recommended for MVP)**
- Launch one goroutine per batch
- Simple, easy to understand
- Works well for <100 batches
- Suitable for most use cases

```go
for _, offset := range batches {
    wg.Add(1)
    go fetchBatch(ctx, vdbClient, collection, batchSize, offset, results, &wg)
}
```

**Option B: Fixed worker pool (future optimization)**
- Limit concurrent goroutines to N workers
- Better for very large collections (1000+ batches)
- More complex error handling
- Deferred to Sprint 4+

### 3. Error Handling

**Strategy**: Fail fast on first error

```go
var firstErr error
var errMutex sync.Mutex

for result := range results {
    if result.err != nil {
        errMutex.Lock()
        if firstErr == nil {
            firstErr = result.err
        }
        errMutex.Unlock()
        // Continue draining channel to avoid goroutine leaks
    } else {
        batchResults = append(batchResults, result)
    }
}

if firstErr != nil {
    return fmt.Errorf("parallel batch fetch failed: %w", firstErr)
}
```

**Alternative**: Collect all errors and return aggregate error
- More complex implementation
- Provides better debugging info
- Deferred to Sprint 4+

### 4. Context Cancellation

**Requirement**: Support context cancellation for long-running operations

```go
func fetchBatch(ctx context.Context, client VectorDBClient, ...) {
    // Check context before expensive operation
    select {
    case <-ctx.Done():
        results <- batchResult{err: ctx.Err()}
        return
    default:
    }

    // Fetch with context
    docs, err := client.ListDocuments(ctx, collection, batchSize, offset)
    results <- batchResult{offset: offset, docs: docs, err: err}
}
```

**Benefit**: User can Ctrl+C to cancel long backups

### 5. Memory Management

**Challenge**: Large collections with many batches may consume significant memory

**Solution**: Process batches in waves (future optimization)

```go
// Current: Fetch all batches at once (simple, works for <10K docs)
// Future: Fetch in waves of N batches (for 10K+ docs)
waveSize := 10  // Fetch 10 batches at a time
for i := 0; i < len(batches); i += waveSize {
    end := min(i+waveSize, len(batches))
    fetchWave(batches[i:end])
}
```

**Deferred**: Sprint 4+ (not needed for v0.12.0 goal)

### 6. Progress Tracking

**Challenge**: Progress tracking is more complex with concurrent fetches

**Solution**: Atomic counter for completed batches

```go
var completed int32

go func(off int) {
    defer wg.Done()
    docs, err := vdbClient.ListDocuments(ctx, collection, batchSize, off)

    // Update progress
    count := atomic.AddInt32(&completed, 1)
    if !quiet {
        progress := float64(count) / float64(numBatches) * 100
        fmt.Printf("\rProgress: %d/%d batches (%.1f%%)", count, numBatches, progress)
    }

    results <- batchResult{offset: off, docs: docs, err: err}
}(offset)
```

---

## Implementation Plan

### Phase 1: Core Parallel Fetching (4 hours)

**File**: `src/cmd/backup/create.go`

1. Define `batchResult` struct
2. Implement parallel fetch logic
3. Add sorting to maintain document order
4. Basic error handling (fail fast)

### Phase 2: Context & Cancellation (1 hour)

1. Thread context through all goroutines
2. Test Ctrl+C cancellation
3. Ensure no goroutine leaks

### Phase 3: Progress Tracking (1 hour)

1. Add atomic counter for completed batches
2. Update progress reporting
3. Test with quiet mode

### Phase 4: Testing (2 hours)

1. Unit tests with mock VDB client
2. Integration tests with real VDBs (Weaviate, Milvus)
3. Load testing with 301+ docs
4. Error scenario testing

**Total Estimated Effort**: 8 hours

---

## Performance Expectations

### Current Performance (Sequential, batch=200)

- AuctionsImages (301 docs): 0.71s backup time
- Throughput: 424 docs/sec
- Batches: 2 sequential queries

### Expected Performance (Parallel, batch=200)

**Assumption**: 2 concurrent queries complete in ~same time as 1 query

- Expected time: ~0.35-0.40s (50% reduction)
- Expected throughput: 750-860 docs/sec
- **Improvement**: 2x faster than current

### Conservative Estimate

- Time: ~0.50s (30% reduction due to overhead)
- Throughput: 600 docs/sec
- **Improvement**: 1.4x faster

**Goal**: Exceed v0.12.0 target of 500+ docs/sec ✅

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Goroutine leaks | Medium | High | Proper defer, context cancellation, channel closure |
| Race conditions | Low | High | Use sync primitives, careful channel usage |
| Memory explosion | Low | Medium | Limit to reasonable batch counts (<100) |
| VDB rate limiting | Medium | Medium | Add concurrency limit flag (future) |
| Document ordering bugs | Medium | High | Thorough testing of sort logic |

---

## Testing Strategy

### Unit Tests

```go
func TestParallelBatchFetch(t *testing.T) {
    tests := []struct {
        name      string
        totalDocs int
        batchSize int
        wantErr   bool
    }{
        {"small collection", 10, 5, false},
        {"exact multiple", 100, 20, false},
        {"not exact multiple", 301, 200, false},
        {"single batch", 50, 100, false},
        {"error in batch", 100, 20, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Integration Tests

1. **Real VDB Test**: Weaviate local with DemoDocs (38 docs)
2. **Large Collection Test**: AuctionsImages (301 docs)
3. **Error Scenario**: Force network error mid-fetch
4. **Cancellation Test**: Cancel context during fetch

### Performance Benchmarks

```go
func BenchmarkSequentialFetch(b *testing.B) {
    // Current implementation
}

func BenchmarkParallelFetch(b *testing.B) {
    // New implementation
}
```

**Expected**: Parallel should be ~2x faster

---

## Rollout Plan

### v0.12.0-alpha (Testing)

- Feature flag: `--parallel-fetch` (opt-in)
- Default: sequential (safe)
- Allows testing without risk

### v0.12.0-beta (Validation)

- Default: parallel (new behavior)
- Fallback flag: `--sequential-fetch` (escape hatch)
- Gather performance metrics

### v0.12.0 (Release)

- Default: parallel
- Remove `--sequential-fetch` flag
- Document performance improvement

---

## Future Optimizations (Post-v0.12.0)

### 1. Fixed Worker Pool (Sprint 4+)

- Limit concurrent goroutines to N workers
- Prevents overwhelming VDB with too many connections
- Configurable: `--max-workers=8`

### 2. Adaptive Concurrency (Sprint 5+)

- Start with low concurrency (2-4 workers)
- Increase if queries complete quickly
- Decrease if errors or timeouts occur
- Self-tuning based on VDB performance

### 3. Connection Pooling (Sprint 3)

- Reuse VDB connections across batches
- Reduce connection overhead
- 10-20% improvement expected

### 4. Streaming Results (Sprint 4+)

- Don't wait for all batches to complete
- Stream results to disk as they arrive
- Lower memory footprint
- Faster perceived performance

---

## Alternatives Considered

### Alternative 1: Keep Sequential (Rejected)

- **Pros**: Simple, safe, no risk
- **Cons**: Leaves 2-3x performance on the table
- **Decision**: Rejected, performance improvement too significant

### Alternative 2: Async/Await Pattern (Rejected)

- **Pros**: Familiar to JavaScript/TypeScript developers
- **Cons**: Not idiomatic Go, more complex than goroutines
- **Decision**: Rejected, goroutines are simpler and more idiomatic

### Alternative 3: Channel-based Worker Pool (Deferred)

- **Pros**: Better control over concurrency
- **Cons**: More complex than needed for MVP
- **Decision**: Deferred to Sprint 4+ if needed

---

## Success Criteria

- ✅ Backup speed: 600+ docs/sec (exceeds v0.12.0 target of 500+)
- ✅ All unit tests passing
- ✅ Integration tests passing with real VDBs
- ✅ No goroutine leaks (validated with race detector)
- ✅ Proper error handling (no silent failures)
- ✅ Context cancellation working (Ctrl+C stops immediately)
- ✅ Document ordering preserved (validated with test cases)

---

## References

- **Profiling Report**: `docs/archive/performance/PROFILING_MARCH_15_2026.md`
- **Quick Win #1**: Default batch size 200 (implemented)
- **Go Blog**: [Go Concurrency Patterns](https://go.dev/blog/pipelines)
- **Effective Go**: [Concurrency](https://go.dev/doc/effective_go#concurrency)

---

**Next Steps**:
1. Review this design document
2. Get feedback on approach
3. Implement Phase 1 (core parallel fetching)
4. Test with real collections
5. Iterate based on results

**Status**: ✅ Design complete, ready for implementation
