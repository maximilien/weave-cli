# Monday Quick Start - February 17, 2026

## 🎯 Today's Mission
**Complete Issue #31 parallel processing implementation (4-5 hours)**

---

## ✅ Weekend Recap (What Got Done)

### v0.9.24 Released
- ✅ **Issue #32 CLOSED**: Fixed critical query embedding bug
  - All VDBs now use collection's configured embedding model
  - No more dimension mismatches (768 vs 1536)
  - Client0 fully unblocked for OSS queries

- ✅ **Issue #31 Infrastructure (60% complete)**:
  - Rate limiter package (11 tests passing)
  - Worker pool package (10 tests passing)
  - `--workers` flag added to CLI

**Commits**:
- `770c4d7` - fix(qdrant): use collection's embedding model
- `9d09ffb` - fix(mongodb): use collection's embedding model
- `6fe36a0` - fix(neo4j,pinecone): use collection's embedding model
- `ed362fd` - feat: rate limiter package
- `4d8d177` - feat: worker pool package
- `0376470` - feat: add --workers flag to CLI

**Version Check**:
```bash
./bin/weave -V
# Should show: Weave CLI 0.9.24
```

---

## 📋 Today's Tasks (Priority Order)

### 1. Wire Worker Pool to Document Ingestion (2-3 hours)
**File**: `src/cmd/utils/document_utils.go` (likely)

**Tasks**:
- [ ] Modify `CreateDocument()` to detect `--workers > 1`
- [ ] Create worker pool with config (workers, embedding model)
- [ ] Convert single file ingestion to task submission
- [ ] Handle results aggregation
- [ ] Implement progress tracking across workers
- [ ] Add error handling for partial failures

**Pattern**:
```go
// If workers > 1, use parallel mode
if workers > 1 {
    pool := worker.NewPool(worker.Config{
        Workers:        workers,
        EmbeddingModel: embeddingModel,
        QueueSize:      workers * 2,
    })
    pool.Start()
    defer pool.Stop()

    // Submit tasks...
    // Collect results...
}
```

### 2. Add Glob Pattern Support (1 hour)
**File**: `src/cmd/document/create.go` or utils

**Tasks**:
- [ ] Check if `filePath` contains glob patterns (`*`, `?`, `[...]`)
- [ ] Enumerate matching files using `filepath.Glob()`
- [ ] Submit each file as separate task to worker pool
- [ ] Track per-file progress and errors

**Pattern**:
```go
files, err := filepath.Glob(filePath)
for _, file := range files {
    pool.Submit(worker.Task{FilePath: file, ...})
}
```

### 3. Progress Aggregation (30 min)
**Files**: Progress tracking across workers

**Tasks**:
- [ ] Aggregate progress from multiple workers
- [ ] Update progress bar with total/completed/failed
- [ ] Handle concurrent progress updates (thread-safe)

### 4. Integration Tests (1 hour)
**File**: Create new test file or extend existing

**Tasks**:
- [ ] Test parallel ingestion with 3 workers, 10 small files
- [ ] Test error handling (1 file fails, others succeed)
- [ ] Test cancellation (Ctrl+C should stop gracefully)
- [ ] Test rate limiting with OpenAI model

### 5. Documentation (30 min)
**Files**: `src/cmd/document/create.go` help text, README

**Tasks**:
- [ ] Update command help with parallel examples
- [ ] Add performance notes (expected speedup)
- [ ] Document rate limiting behavior
- [ ] Add troubleshooting tips

---

## 🧪 Testing Checklist

### Unit Tests
```bash
go test ./src/pkg/worker/... -v
go test ./src/pkg/ratelimit/... -v
```

### Integration Tests (Create These)
```bash
# Sequential (baseline)
./bin/weave docs create TestCol test_pdfs/*.pdf

# Parallel (3 workers)
./bin/weave docs create TestCol test_pdfs/*.pdf --workers 3

# Verify speedup
# Expected: 2-3x faster with 3 workers
```

### Performance Test (Manual)
```bash
# 10 PDFs, sequential
time ./bin/weave docs create TestCol test_pdfs/*.pdf

# 10 PDFs, 3 workers
time ./bin/weave docs create TestCol test_pdfs/*.pdf --workers 3

# Compare results
```

---

## 🚨 Gotchas to Watch For

1. **Rate Limiting**: Ensure OpenAI models are rate-limited, OSS models are not
2. **Progress Bar**: Must be thread-safe when updated by multiple workers
3. **Error Handling**: One file failure shouldn't crash entire batch
4. **Context Cancellation**: Ctrl+C should gracefully stop all workers
5. **Memory**: Don't load all files into memory at once (stream processing)
6. **File Ordering**: Results may not be in same order as input (async)

---

## 🎯 Success Criteria

By end of day Monday:
- [ ] `--workers 3` successfully ingests 10 files in parallel
- [ ] All unit tests passing
- [ ] At least 1 integration test passing
- [ ] No regressions (sequential mode still works)
- [ ] Ready for v0.9.25 release (Tuesday)

---

## 🔍 Files to Focus On

### Primary (Implementation)
1. `src/cmd/utils/document_utils.go` - Main ingestion logic
2. `src/cmd/document/create.go` - CLI flag handling (already done)
3. `src/pkg/worker/pool.go` - Worker pool (already done)
4. `src/pkg/ratelimit/ratelimit.go` - Rate limiting (already done)

### Secondary (Testing)
1. `src/cmd/utils/document_utils_test.go` - Add parallel tests
2. `src/pkg/worker/pool_test.go` - Already has 10 tests
3. `src/pkg/ratelimit/ratelimit_test.go` - Already has 11 tests

### Tertiary (Documentation)
1. `README.md` - Add parallel ingestion examples
2. `src/cmd/document/create.go` - Update help text

---

## 🐛 Known Issues to Address

### Issue #31 Remaining Work
- CLI wiring (main task for today)
- Glob pattern handling
- Progress aggregation
- Integration tests

### Issue #29 (Milvus 65KB Limit)
- External storage already implemented (v0.9.21-23)
- Need to verify it works and close issue

---

## 💡 Quick Commands

### Build
```bash
./build.sh
```

### Test
```bash
# All tests
go test ./... -v

# Specific package
go test ./src/pkg/worker/... -v
go test ./src/pkg/ratelimit/... -v
```

### Lint
```bash
./lint.sh
```

### Git Status
```bash
git status
git log --oneline -5
```

---

## 📊 Expected Timeline

| Time | Task | Duration |
|------|------|----------|
| 9:00 AM | Wire worker pool to ingestion | 2-3 hours |
| 12:00 PM | Add glob pattern support | 1 hour |
| 1:00 PM | Progress aggregation | 30 min |
| 1:30 PM | Integration tests | 1 hour |
| 2:30 PM | Documentation | 30 min |
| 3:00 PM | Testing & bug fixes | 1 hour |
| **4:00 PM** | **v0.9.25 ready** | **DONE** |

---

## 🎉 Tuesday Goals (After Monday Success)

- Final testing of parallel ingestion
- Release v0.9.25
- Close Issue #31
- Check Client0 feedback on v0.9.24
- Start stretch goals (video demos or external storage tests)

---

## 📝 Notes

- Keep commits small and incremental
- Test after each major change
- Don't break sequential mode (workers=1)
- Rate limiting is already done, just use it
- Worker pool is already done, just wire it up

---

**Status**: Ready to code
**Focus**: Issue #31 completion
**Goal**: v0.9.25 by end of day
**Mindset**: Ship it! 🚀
