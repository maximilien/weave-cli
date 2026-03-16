# Backup Performance Optimization Opportunities
**Based on profiling results from March 15, 2026**

## Executive Summary

**Current Performance**: 184 docs/sec (batch=100), 376 docs/sec (batch=200)
**v0.12.0 Target**: 500+ docs/sec
**Gap**: 1.3x improvement needed

**Recommendation**: Implement Quick Win #1 (change default batch size to 200) immediately for 2x improvement with near-zero effort.

---

## Quick Wins (1-2 hours each)

### 1. Increase Default Batch Size to 200 ⚡ **HIGHEST IMPACT**

**Impact**: 2x faster backups (184 → 376 docs/sec)
**Effort**: 5 minutes
**Risk**: Very low
**Files**: `src/cmd/backup/create.go:25`

**Change**:
```go
// Before
createOpts = &backuppkg.CreateOptions{
    BatchSize: 100,  // ← Change this
    Compress:  true,
}

// After
createOpts = &backuppkg.CreateOptions{
    BatchSize: 200,  // 2x faster based on profiling
    Compress:  true,
}
```

**Testing**:
- Re-run profiling script to validate
- Test with Client0's 2,636-doc collections
- Ensure no memory issues with larger batches

**Documentation**:
- Update `--batch-size` flag help text
- Add batch size recommendations to BACKUP_RESTORE.md

---

### 2. Document Batch Size Best Practices 📚

**Impact**: Users can self-optimize performance
**Effort**: 15 minutes
**Risk**: None
**Files**: `docs/guides/BACKUP_RESTORE.md`

**Add section**:
```markdown
## Performance Tuning

### Batch Size Optimization

The `--batch-size` flag controls how many documents are fetched per VDB query:

- **Default**: 200 (optimized for most use cases)
- **Small collections (<100 docs)**: Use default
- **Large collections (1000+ docs)**: Try 500-1000 for best performance
- **Memory constrained**: Use 50-100

**Example**:
\`\`\`bash
# Fast backup for large collections
weave backup create MyCollection --output backup.weavebak --batch-size 500

# Memory-efficient backup
weave backup create MyCollection --output backup.weavebak --batch-size 50
\`\`\`

**Performance impact** (301 docs test):
- Batch 50: 112 docs/sec
- Batch 100: 184 docs/sec
- Batch 200: 376 docs/sec ← **2x faster!**
```

---

### 3. Connection Pooling for VDB Clients 🔌

**Impact**: 10-20% improvement
**Effort**: 2-3 hours
**Risk**: Low
**Files**: `src/pkg/vectordb/*.go`

**Current**: New connection per batch (expensive)
**Proposed**: Reuse connection across batches

**Implementation**:
1. Add connection reuse to VectorDBClient interface
2. Cache connections in backup process
3. Close connection after all batches complete

**Complexity**: Low (most VDB clients already support pooling)

---

## Medium-term Optimizations (4-8 hours each)

### 4. Parallel Batch Processing 🚀 **HIGHEST IMPACT**

**Impact**: 2-3x improvement (750-1000 docs/sec)
**Effort**: 6-8 hours
**Risk**: Medium (concurrency bugs)
**Files**: `src/pkg/backup/format.go`, `src/cmd/backup/create.go`

**Current**: Sequential batch fetching
```go
for offset := 0; offset < totalDocs; offset += batchSize {
    docs := vdbClient.ListDocuments(...)  // ← Sequential
    backup.Documents = append(...)
}
```

**Proposed**: Concurrent batch fetching with goroutines
```go
type batchResult struct {
    offset int
    docs   []vectordb.Document
    err    error
}

results := make(chan batchResult, numBatches)
var wg sync.WaitGroup

for offset := 0; offset < totalDocs; offset += batchSize {
    wg.Add(1)
    go func(off int) {
        defer wg.Done()
        docs, err := vdbClient.ListDocuments(ctx, collection, batchSize, off)
        results <- batchResult{off, docs, err}
    }(offset)
}

go func() {
    wg.Wait()
    close(results)
}()

// Collect results (may need sorting by offset)
for result := range results {
    if result.err != nil {
        return err
    }
    backup.Documents = append(backup.Documents, result.docs...)
}
```

**Challenges**:
- Document ordering (need to sort by offset)
- Error handling across goroutines
- Rate limiting (don't overwhelm VDB)
- Memory management (large collections)

**Testing**:
- Unit tests with mock VDB
- Integration tests with real VDBs
- Load testing with 10K+ docs

---

### 5. Streaming JSON Serialization 📤

**Impact**: Lower memory footprint, 5-10% faster
**Effort**: 4-5 hours
**Risk**: Low
**Files**: `src/pkg/backup/format.go:WriteBackup()`

**Current**: Build entire JSON in memory, then write
```go
data, err := json.Marshal(backup)  // ← Entire backup in memory
if err != nil {
    return err
}
file.Write(data)
```

**Proposed**: Stream JSON output
```go
encoder := json.NewEncoder(file)
encoder.SetIndent("", "  ")
if err := encoder.Encode(backup); err != nil {
    return err
}
```

**Benefits**:
- Lower memory usage (important for 10K+ docs)
- Faster startup (don't wait for entire JSON)
- Compatible with streaming compression

---

### 6. Streaming Compression 🗜️

**Impact**: 5-10% faster, lower memory
**Effort**: 3-4 hours
**Risk**: Low
**Files**: `src/pkg/backup/format.go`

**Current**: Compress entire file after writing
```go
WriteBackup(backup, file)  // ← Write full file
CompressFile(file)         // ← Then compress
```

**Proposed**: Compress while writing
```go
gzWriter := gzip.NewWriter(file)
defer gzWriter.Close()

encoder := json.NewEncoder(gzWriter)  // ← Compress during write
encoder.Encode(backup)
```

**Benefits**:
- No intermediate uncompressed file
- Lower disk I/O
- Lower memory usage

---

## Long-term Optimizations (8+ hours each)

### 7. Adaptive Batch Sizing 🎯

**Impact**: Optimal performance per VDB type
**Effort**: 8-10 hours
**Risk**: Medium

**Concept**: Automatically adjust batch size based on VDB type and collection size

**Implementation**:
```go
func GetOptimalBatchSize(vdbType string, totalDocs int) int {
    switch vdbType {
    case "milvus":
        if totalDocs > 10000 {
            return 500  // Milvus handles large batches well
        }
        return 200
    case "weaviate":
        return 100  // Weaviate prefers smaller batches
    case "qdrant":
        return 300
    default:
        return 200
    }
}
```

**Testing**:
- Benchmark each VDB type with varying batch sizes
- Build lookup table of optimal sizes
- Allow user override with --batch-size

---

### 8. Incremental Backups 🔄

**Impact**: 10-100x faster for daily backups
**Effort**: 16-20 hours
**Risk**: High (complexity)

**Concept**: Backup only changed documents since last backup

**Implementation**:
- Track last backup timestamp in metadata
- Query VDB for documents modified since timestamp
- Store incremental backup separately
- Merge incremental with full backup on restore

**Challenges**:
- Not all VDBs support timestamp queries
- Complex merge logic
- Validation and consistency

**Deferred**: Sprint 13 (v0.13.0)

---

## Prioritized Roadmap

### This Week (Sprint 3 Prep)
1. ⚡ **Quick Win #1**: Change default batch size to 200 (5 min)
2. 📚 **Quick Win #2**: Document batch size tuning (15 min)
3. 🔬 **Validation**: Test with Client0's datasets

### Next Week (Sprint 3)
4. 🚀 **Parallel batch processing** (6-8 hours) ← **Highest impact**
5. 🔌 **Connection pooling** (2-3 hours)
6. 📤 **Streaming JSON serialization** (4-5 hours)

### Week After (Sprint 3 cont.)
7. 🗜️ **Streaming compression** (3-4 hours)
8. 📊 **Performance documentation** (2 hours)
9. ✅ **v0.12.0 Release** (backup speed: 750-1000 docs/sec)

### Future (Sprint 13)
10. 🎯 **Adaptive batch sizing** (8-10 hours)
11. 🔄 **Incremental backups** (16-20 hours)

---

## Expected Performance Gains

| Optimization | Current | After | Improvement |
|--------------|---------|-------|-------------|
| Baseline (batch=100) | 184 docs/sec | - | - |
| **Quick Win #1** (batch=200) | 184 docs/sec | **376 docs/sec** | **2.0x** |
| + Parallel processing | 376 docs/sec | **750-1000 docs/sec** | **2.0-2.7x** |
| + Connection pooling | 750 docs/sec | **825-900 docs/sec** | **1.1-1.2x** |
| + Streaming JSON/compression | 825 docs/sec | **870-950 docs/sec** | **1.05-1.1x** |
| **Total (v0.12.0)** | **184 docs/sec** | **950 docs/sec** | **5.2x** |

**v0.12.0 Target**: 500+ docs/sec ✅ (will achieve 950 docs/sec)

---

## Risk Assessment

| Optimization | Risk Level | Mitigation |
|--------------|------------|------------|
| Batch size 200 | 🟢 Low | Users can override, memory unlikely issue |
| Documentation | 🟢 None | - |
| Connection pooling | 🟢 Low | Revert if issues, good test coverage |
| Parallel processing | 🟡 Medium | Extensive testing, limit concurrency |
| Streaming JSON | 🟢 Low | Well-tested Go stdlib, backwards compatible |
| Streaming compression | 🟢 Low | Same as above |
| Adaptive batch sizing | 🟡 Medium | Extensive VDB testing needed |
| Incremental backups | 🔴 High | Complex logic, deferred to v0.13.0 |

---

## Files Modified (for Quick Win #1)

### src/cmd/backup/create.go
```diff
 var (
     createOpts = &backuppkg.CreateOptions{
-        BatchSize: 100,
+        BatchSize: 200,  // 2x faster based on v0.11.3 profiling (Mar 15, 2026)
         Compress:  true,
     }
```

### src/cmd/backup/create.go (flag help text)
```diff
-CreateCmd.Flags().IntVar(&createOpts.BatchSize, "batch-size", 100, "Documents per batch")
+CreateCmd.Flags().IntVar(&createOpts.BatchSize, "batch-size", 200, "Documents per batch (default optimized for performance)")
```

### docs/guides/BACKUP_RESTORE.md
Add new section: "Performance Tuning" (see Quick Win #2 above)

---

## Validation Plan

1. **Unit tests**: Verify batch size handling
2. **Integration tests**: Test with real VDBs (Milvus, Weaviate, Qdrant)
3. **Performance tests**: Re-run profiling with batch=200 as default
4. **Client0 validation**: Test with 2,636-doc production datasets
5. **Memory profiling**: Ensure no memory issues with larger batches
6. **Edge cases**: Empty collections, single-doc collections, very large batches (1000+)

---

## Success Metrics

- ✅ Backup speed: 500+ docs/sec (v0.12.0 goal)
- ✅ Memory usage: <500MB for 10K docs
- ✅ Test coverage: 95%+ for backup module
- ✅ Documentation: Complete batch size tuning guide
- ✅ Client0 feedback: Positive performance validation

---

**Next Step**: Implement Quick Win #1 (change default batch size to 200) and re-run profiling to validate 2x improvement.
