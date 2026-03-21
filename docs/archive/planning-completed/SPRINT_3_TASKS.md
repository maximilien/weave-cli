# Sprint 3: Performance Optimizations

**Sprint Duration**: March 17-21, 2026 (5 days)
**Sprint Goal**: Achieve 750-1000 docs/sec backup performance
**v0.12.0 Target**: 500+ docs/sec

---

## Sprint Overview

### Starting Point
- ✅ Quick Win #1 complete: Default batch size 200
- ✅ Current performance: 376-424 docs/sec
- ✅ 75% to v0.12.0 goal

### Sprint Goal
- Parallel batch processing → 2-3x improvement
- Connection pooling → 10-20% improvement
- Streaming optimizations → 5-10% improvement
- **Target**: 750-1000 docs/sec (exceeds v0.12.0 goal by 50-100%)

### Deliverables
1. Parallel batch processing implementation
2. Connection pooling for VDB clients
3. Streaming JSON/compression
4. Performance benchmarks and documentation
5. v0.12.0 release prep

---

## Task Breakdown

### Monday (Mar 17): Parallel Batch Processing - Core (8 hours)

#### Task 1.1: Implement batchResult structure (30 min)
**File**: `src/cmd/backup/create.go`

**Subtasks**:
- [ ] Define `batchResult` struct with offset, docs, error fields
- [ ] Add helper function to calculate batch offsets
- [ ] Add unit tests for offset calculation

**Acceptance Criteria**:
- Struct defined and documented
- Offset calculation handles edge cases (not exact multiples)
- Tests passing

---

#### Task 1.2: Implement parallel fetch logic (2 hours)
**File**: `src/cmd/backup/create.go`

**Subtasks**:
- [ ] Create goroutine-based batch fetching
- [ ] Implement results channel pattern
- [ ] Add WaitGroup for goroutine synchronization
- [ ] Implement result collection logic

**Code Template**:
```go
results := make(chan batchResult, numBatches)
var wg sync.WaitGroup

for _, offset := range batches {
    wg.Add(1)
    go func(off int) {
        defer wg.Done()
        // Fetch logic here
        results <- batchResult{...}
    }(offset)
}

go func() {
    wg.Wait()
    close(results)
}()
```

**Acceptance Criteria**:
- Parallel fetching working
- No goroutine leaks
- Channel properly closed

---

#### Task 1.3: Implement document ordering (1 hour)
**File**: `src/cmd/backup/create.go`

**Subtasks**:
- [ ] Collect all batch results
- [ ] Sort results by offset
- [ ] Append documents in order
- [ ] Unit tests for ordering logic

**Acceptance Criteria**:
- Documents maintain original order
- Tests cover various batch sizes and collection sizes

---

#### Task 1.4: Error handling (1.5 hours)
**File**: `src/cmd/backup/create.go`

**Subtasks**:
- [ ] Implement fail-fast on first error
- [ ] Add error mutex to prevent races
- [ ] Drain results channel to avoid goroutine leaks
- [ ] Add error tests

**Acceptance Criteria**:
- First error is captured and returned
- No goroutine leaks on error
- Clear error messages with batch offset info

---

#### Task 1.5: Context cancellation (1 hour)
**File**: `src/cmd/backup/create.go`

**Subtasks**:
- [ ] Thread context through goroutines
- [ ] Add context checks before expensive operations
- [ ] Test Ctrl+C cancellation
- [ ] Ensure proper cleanup on cancellation

**Acceptance Criteria**:
- Context cancellation stops all goroutines
- No resource leaks
- Clear cancellation message to user

---

#### Task 1.6: Progress tracking (1 hour)
**File**: `src/cmd/backup/create.go`

**Subtasks**:
- [ ] Add atomic counter for completed batches
- [ ] Update progress reporting for parallel fetches
- [ ] Test progress output
- [ ] Ensure quiet mode still works

**Acceptance Criteria**:
- Progress shows "X/Y batches completed"
- Atomic operations prevent race conditions
- Quiet mode suppresses progress

---

#### Task 1.7: Integration testing (1 hour)
**Files**: `src/cmd/backup/create_test.go` (new), `src/pkg/backup/parallel_test.go` (new)

**Subtasks**:
- [ ] Test with mock VDB client
- [ ] Test with real Weaviate local (DemoDocs 38 docs)
- [ ] Test with AuctionsImages (301 docs)
- [ ] Test error scenarios
- [ ] Test cancellation

**Acceptance Criteria**:
- All tests passing
- Performance improvement validated (2x faster)

---

### Tuesday (Mar 18): Connection Pooling (3 hours)

#### Task 2.1: Design connection pooling interface (1 hour)
**File**: `src/pkg/vectordb/client.go`

**Subtasks**:
- [ ] Add `Connect()` and `Disconnect()` methods to VectorDBClient interface
- [ ] Document connection lifecycle
- [ ] Plan implementation for each VDB type

**Acceptance Criteria**:
- Interface changes backward compatible
- Clear documentation

---

#### Task 2.2: Implement connection pooling for Weaviate (1 hour)
**File**: `src/pkg/vectordb/weaviate/client.go`

**Subtasks**:
- [ ] Cache Weaviate client connection
- [ ] Reuse connection across batch queries
- [ ] Add connection cleanup on backup complete
- [ ] Test connection reuse

**Acceptance Criteria**:
- Connection reused across batches
- No connection leaks
- 10-20% performance improvement

---

#### Task 2.3: Implement connection pooling for Milvus (1 hour)
**File**: `src/pkg/vectordb/milvus/client.go`

**Subtasks**:
- [ ] Same as Weaviate
- [ ] Test with Milvus local

**Acceptance Criteria**:
- Same as Weaviate

---

### Wednesday (Mar 19): Streaming Optimizations (4 hours)

#### Task 3.1: Implement streaming JSON serialization (2 hours)
**File**: `src/pkg/backup/format.go`

**Subtasks**:
- [ ] Replace `json.Marshal()` with `json.NewEncoder()`
- [ ] Stream JSON directly to file writer
- [ ] Add tests for large backups
- [ ] Measure memory usage improvement

**Code Template**:
```go
// Current
data, err := json.Marshal(backup)
file.Write(data)

// New
encoder := json.NewEncoder(file)
encoder.SetIndent("", "  ")
encoder.Encode(backup)
```

**Acceptance Criteria**:
- Lower memory footprint (measure with profiling)
- Backward compatible (same JSON output)
- Tests passing

---

#### Task 3.2: Implement streaming compression (2 hours)
**File**: `src/pkg/backup/format.go`

**Subtasks**:
- [ ] Replace two-phase (write→compress) with streaming gzip
- [ ] Use `gzip.NewWriter()` wrapping file writer
- [ ] Add tests for compressed output
- [ ] Measure performance improvement

**Code Template**:
```go
// Current
WriteBackup(backup, file)
CompressFile(file)

// New
gzWriter := gzip.NewWriter(file)
defer gzWriter.Close()
encoder := json.NewEncoder(gzWriter)
encoder.Encode(backup)
```

**Acceptance Criteria**:
- No intermediate uncompressed file
- Same compression ratio
- 5-10% faster

---

### Thursday (Mar 20): Testing & Benchmarking (4 hours)

#### Task 4.1: Comprehensive integration testing (2 hours)
**Files**: Various test files

**Subtasks**:
- [ ] Test all optimizations together
- [ ] Test with multiple VDB types
- [ ] Test with various collection sizes
- [ ] Load testing with 1000+ docs

**Acceptance Criteria**:
- All tests passing
- No regressions
- Performance validated

---

#### Task 4.2: Performance benchmarking (2 hours)
**File**: `src/pkg/backup/benchmark_test.go` (new)

**Subtasks**:
- [ ] Create Go benchmarks for backup operations
- [ ] Benchmark sequential vs parallel
- [ ] Benchmark with/without connection pooling
- [ ] Benchmark with/without streaming
- [ ] Document results

**Expected Results**:
- Parallel: 2x faster than sequential
- With pooling: 1.1-1.2x faster
- With streaming: 1.05-1.1x faster
- **Combined**: 750-1000 docs/sec

**Acceptance Criteria**:
- Benchmarks running successfully
- Results documented in profiling report

---

### Friday (Mar 21): Documentation & Release Prep (3 hours)

#### Task 5.1: Update documentation (1.5 hours)
**Files**: `docs/guides/BACKUP_RESTORE.md`, `CHANGELOG.md`, `docs/PLAN.md`

**Subtasks**:
- [ ] Update Performance section with new metrics
- [ ] Document parallel processing feature
- [ ] Update CHANGELOG.md for v0.12.0
- [ ] Update PLAN.md with Sprint 3 completion

**Acceptance Criteria**:
- Documentation accurate and complete
- CHANGELOG ready for release

---

#### Task 5.2: Create Sprint 3 retrospective (30 min)
**File**: `docs/sprints/SPRINT_3_RETROSPECTIVE.md`

**Subtasks**:
- [ ] Document what went well
- [ ] Document what could be improved
- [ ] Document lessons learned
- [ ] Plan Sprint 4 goals

**Acceptance Criteria**:
- Retrospective complete
- Sprint 4 scoped

---

#### Task 5.3: v0.12.0 release preparation (1 hour)
**Tasks**:
- [ ] Update version number
- [ ] Create release notes
- [ ] Create GitHub release draft
- [ ] Tag release (if ready)

**Acceptance Criteria**:
- Release notes complete
- Ready to ship v0.12.0

---

## Task Summary by Priority

### P0 - Must Have (v0.12.0 blockers)
1. ✅ Parallel batch processing (Monday)
2. ✅ Basic testing and validation (Monday/Thursday)
3. ✅ Documentation updates (Friday)

### P1 - Should Have (nice to have for v0.12.0)
4. Connection pooling (Tuesday)
5. Streaming JSON/compression (Wednesday)
6. Comprehensive benchmarking (Thursday)

### P2 - Could Have (can defer to v0.12.1)
7. Advanced error handling (collect all errors)
8. Fixed worker pool implementation
9. Adaptive concurrency

---

## Time Estimates

| Day | Tasks | Estimated Hours | Priority |
|-----|-------|-----------------|----------|
| Monday | Parallel processing | 8 hours | P0 |
| Tuesday | Connection pooling | 3 hours | P1 |
| Wednesday | Streaming optimizations | 4 hours | P1 |
| Thursday | Testing & benchmarking | 4 hours | P0/P1 |
| Friday | Docs & release prep | 3 hours | P0 |
| **Total** | | **22 hours** | |

**Sprint Capacity**: 5 days × 4-5 hours/day = 20-25 hours
**Status**: ✅ Sprint is feasible

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Goroutine leaks | Use `go test -race`, defer cleanup, context cancellation |
| Performance regression | Benchmark before/after, feature flag for rollback |
| VDB compatibility | Test with multiple VDB types early |
| Time overrun | P1 tasks can be deferred to v0.12.1 if needed |

---

## Success Criteria

### Sprint 3 Success
- ✅ Parallel batch processing working
- ✅ Backup speed: 600+ docs/sec (exceeds v0.12.0 goal)
- ✅ All tests passing
- ✅ No regressions
- ✅ Documentation complete

### v0.12.0 Success
- ✅ Backup speed: 500+ docs/sec (goal)
- ✅ Projected: 750-1000 docs/sec (exceeds by 50-100%)
- ✅ Stable, production-ready
- ✅ Client0 validation successful

---

## Dependencies

- ✅ Quick Win #1 complete (default batch size 200)
- ✅ Profiling complete (baseline metrics)
- ✅ Design document complete (parallel processing)
- ⏳ Client0 feedback on v0.11.4 (optional, nice to have)

---

## Next Steps (Monday AM)

1. Review this task breakdown
2. Start Task 1.1 (batchResult structure)
3. Work through tasks sequentially
4. Update PLAN.md daily with progress
5. Ship v0.12.0 by Friday!

---

**Sprint Owner**: @maximilien
**Status**: ✅ Sprint 3 planned and ready to start
**Created**: March 16, 2026
