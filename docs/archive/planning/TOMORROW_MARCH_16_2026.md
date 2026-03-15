# Tasks for Tomorrow - March 16, 2026

## Quick Summary

**Yesterday's Work**: ✅ Performance profiling complete
**Key Discovery**: 🚀 Batch size 200 is 2x faster than batch size 100
**Today's Goal**: Implement Quick Win for 2x performance improvement

---

## Priority 1: Quick Win Implementation (1-2 hours) ⚡

### Task 1: Change Default Batch Size to 200 (5 min)

**File**: `src/cmd/backup/create.go:25`

**Current Code**:
```go
var (
	createOpts = &backuppkg.CreateOptions{
		BatchSize: 100,  // ← Change this
		Compress:  true,
	}
```

**New Code**:
```go
var (
	createOpts = &backuppkg.CreateOptions{
		BatchSize: 200,  // 2x faster based on v0.11.3 profiling (Mar 15, 2026)
		Compress:  true,
	}
```

---

### Task 2: Update Flag Help Text (5 min)

**File**: `src/cmd/backup/create.go` (around line 96)

**Current**:
```go
CreateCmd.Flags().IntVar(&createOpts.BatchSize, "batch-size", 100, "Documents per batch")
```

**New**:
```go
CreateCmd.Flags().IntVar(&createOpts.BatchSize, "batch-size", 200, "Documents per batch (default optimized for performance)")
```

---

### Task 3: Run Tests (10 min)

```bash
# Run backup tests
go test ./src/pkg/backup/... -v

# Run backup command tests
go test ./src/cmd/backup/... -v

# Run linting
./lint.sh
```

**Expected**: All tests should pass (no new tests needed, flag already supported)

---

### Task 4: Re-run Profiling (20 min)

**Validate the 2x improvement with new default**:

```bash
# Re-run profiling script
/tmp/perf-profile-backup.sh

# Compare results (should see batch=200 performance as default now)
ls -lh /tmp/backup-perf-results/
```

**Expected Results**:
- AuctionsImages backup: ~0.80s (was 1.64s with batch=100)
- Throughput: ~376 docs/sec (was 184 docs/sec)

---

### Task 5: Commit Changes (20 min)

**Stage changes**:
```bash
git add src/cmd/backup/create.go
```

**Commit message**:
```bash
git commit -m "$(cat <<'EOF'
perf: change default batch size from 100 to 200 (2x faster)

Based on performance profiling (March 15, 2026), batch size 200
is 2x faster than batch size 100 for backup operations.

Performance Impact:
- Before (batch=100): 184 docs/sec
- After (batch=200): 376 docs/sec
- Improvement: +104% (2x faster)

Profiling Results (301 documents):
- Batch 50: 2.69s (112 docs/sec)
- Batch 100: 1.64s (184 docs/sec)
- Batch 200: 0.80s (376 docs/sec) ← New default

Primary Bottleneck Addressed:
- VDB query latency (fewer batches = fewer queries)
- Batch 100: 4 queries
- Batch 200: 2 queries (50% reduction)

This change gets us 75% of the way to v0.12.0 performance goal
(376/500 docs/sec) with zero code complexity.

Users can still override with --batch-size flag if needed.

See: docs/archive/performance/PROFILING_MARCH_15_2026.md

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
EOF
)"
```

**Update CHANGELOG.md** (add to unreleased or v0.11.5):
```markdown
## [Unreleased]

### Performance

- **2x faster backups**: Changed default batch size from 100 to 200
  - Based on comprehensive profiling (March 15, 2026)
  - Throughput improved from 184 to 376 docs/sec
  - Users can override with `--batch-size` flag
  - See profiling report: `docs/archive/performance/PROFILING_MARCH_15_2026.md`
```

---

## Priority 2: Sprint 3 Planning (1 hour)

### Task 6: Design Parallel Batch Processing (30 min)

**Goal**: Sketch out goroutine-based parallel batch fetching

**Key Questions**:
1. How many concurrent goroutines? (start with 4-8)
2. How to handle errors from goroutines?
3. How to maintain document order?
4. How to limit memory usage?

**Sketch Architecture**:
```go
// Pseudo-code for parallel batch processing
type batchResult struct {
    offset int
    docs   []vectordb.Document
    err    error
}

// Create worker pool
numWorkers := 4  // Tune based on VDB type
results := make(chan batchResult, numBatches)
var wg sync.WaitGroup

// Launch workers
for offset := 0; offset < totalDocs; offset += batchSize {
    wg.Add(1)
    go func(off int) {
        defer wg.Done()
        docs, err := vdbClient.ListDocuments(ctx, collection, batchSize, off)
        results <- batchResult{off, docs, err}
    }(offset)
}

// Close results channel when done
go func() {
    wg.Wait()
    close(results)
}()

// Collect results (need to sort by offset)
batches := make([]batchResult, 0, numBatches)
for result := range results {
    if result.err != nil {
        return err
    }
    batches = append(batches, result)
}

// Sort by offset to maintain order
sort.Slice(batches, func(i, j int) bool {
    return batches[i].offset < batches[j].offset
})

// Append documents in order
for _, batch := range batches {
    backup.Documents = append(backup.Documents, batch.docs...)
}
```

**Create design document**: `docs/design/PARALLEL_BATCH_PROCESSING.md`

---

### Task 7: Research Go Concurrency Patterns (30 min)

**Topics to Research**:
1. Worker pool patterns
2. Error collection from goroutines
3. Context cancellation
4. Rate limiting concurrent requests

**Useful Resources**:
- Go Blog: Concurrency Patterns
- Effective Go: Concurrency
- errgroup package for error handling

**Create notes**: `docs/design/CONCURRENCY_PATTERNS.md`

---

## Priority 3: Documentation (30 min)

### Task 8: Update PLAN.md (15 min)

**Updates**:
- Mark Saturday (Mar 15) profiling as complete ✅
- Add Quick Win implementation status
- Update Sprint 3 goals with profiling insights
- Adjust v0.12.0 projections (now 950 docs/sec expected)

---

### Task 9: Create Sprint 3 Task Breakdown (15 min)

**File**: `docs/sprints/SPRINT_3_TASKS.md`

**Tasks to Define**:
1. Parallel batch processing (6-8 hours)
   - Design worker pool architecture
   - Implement goroutine-based fetching
   - Add error handling and synchronization
   - Test with real VDBs

2. Connection pooling (2-3 hours)
   - Add connection reuse to VectorDBClient interface
   - Implement pooling for Weaviate, Milvus, Qdrant
   - Test connection lifecycle

3. Streaming JSON/compression (4-5 hours)
   - Replace in-memory JSON marshaling with streaming
   - Implement streaming compression
   - Test memory usage improvements

**Estimated Total**: 12-16 hours (Sprint 3 week)

---

## Expected Outcomes

After completing today's tasks:

✅ **Performance**: 2x faster backups (376 docs/sec vs 184 docs/sec)
✅ **Code Changes**: Minimal (2 lines changed)
✅ **Testing**: All tests passing, profiling validated
✅ **Documentation**: CHANGELOG updated, code committed
✅ **Sprint 3**: Ready to start with clear plan

---

## Quick Reference

### Files to Modify Today
1. `src/cmd/backup/create.go` - Change default batch size
2. `CHANGELOG.md` - Add performance improvement entry
3. `docs/PLAN.md` - Update task status

### Files to Create Today
1. `docs/design/PARALLEL_BATCH_PROCESSING.md` - Parallel processing design
2. `docs/design/CONCURRENCY_PATTERNS.md` - Go concurrency research notes
3. `docs/sprints/SPRINT_3_TASKS.md` - Sprint 3 task breakdown

### Commands to Run Today
```bash
# Change code (5 min)
vim src/cmd/backup/create.go

# Run tests (10 min)
go test ./src/pkg/backup/... -v
go test ./src/cmd/backup/... -v
./lint.sh

# Re-run profiling (20 min)
/tmp/perf-profile-backup.sh

# Commit changes (20 min)
git add src/cmd/backup/create.go CHANGELOG.md
git commit -m "perf: change default batch size from 100 to 200 (2x faster)"
```

---

## Performance Metrics (for reference)

### Current Performance (v0.11.3, batch=100)
- DemoDocs (38 docs): 112 docs/sec
- WeaveDocs (79 docs): 134 docs/sec
- AuctionsImages (301 docs): 184 docs/sec

### After Quick Win (batch=200)
- Expected: ~376 docs/sec for large collections
- Improvement: 2x (104%)

### v0.12.0 Target (after Sprint 3)
- Goal: 500+ docs/sec
- Projected with all optimizations: 950 docs/sec
- Exceeds target by 90%!

---

## Profiling Reference

**Profiling Report**: `docs/archive/performance/PROFILING_MARCH_15_2026.md`
**Profiling Script**: `/tmp/perf-profile-backup.sh`
**Test Results**: `/tmp/backup-perf-results/`

**Yesterday's Commit**: `b2006ea - docs: performance profiling results and batch size optimization discovery`

---

**Ready to start!** 🚀

Total estimated time: 2.5-3.5 hours
Impact: 2x performance improvement + Sprint 3 ready to go
