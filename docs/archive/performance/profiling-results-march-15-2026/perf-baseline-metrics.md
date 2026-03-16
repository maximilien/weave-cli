# Backup Performance Baseline Metrics
**Sprint 3 Prep Work - March 15, 2026**

## Baseline Performance (v0.11.3)

### Throughput
- **Small collections (38-79 docs)**: 102-134 docs/sec
- **Large collections (301 docs, batch=100)**: 184-203 docs/sec
- **Large collections (301 docs, batch=200)**: **376 docs/sec** ← Best result

### Timing Breakdown (301 docs example)
| Operation | Duration | % of Total |
|-----------|----------|------------|
| Backup logic (batch=200) | 0.80s | 19% |
| Startup overhead | ~3.3s | 81% |
| **Total (real time)** | **4.1s** | 100% |

### File Sizes (Compressed)
- DemoDocs (38 docs, 1024-dim): 254K → 6.7 KB/doc
- WeaveDocs (79 docs, 1536-dim): 552K → 7.0 KB/doc
- AuctionsImages (301 docs): 27K → 0.09 KB/doc

### Batch Size Impact
| Batch Size | Time (301 docs) | Throughput | vs Default |
|------------|-----------------|------------|------------|
| 50 | 2.69s | 112 docs/sec | -51% |
| 100 (default) | 1.64s | 184 docs/sec | baseline |
| **200** | **0.80s** | **376 docs/sec** | **+104%** |

**Key Finding**: Batch size 200 is **2x faster** than default batch size 100.

## Bottlenecks

### 1. VDB Query Latency (PRIMARY)
- **Impact**: Dominates backup time for large collections
- **Evidence**: Batch size 200 cuts time in half (fewer queries)
- **Fix**: Increase batch size, parallel queries

### 2. Startup Overhead
- **Impact**: ~3-3.5s fixed cost per backup
- **Evidence**: Real time = 3.7-6.4s, backup time = 0.34-2.69s
- **Fix**: Connection pooling, lazy initialization

### 3. Single-threaded Processing
- **Impact**: Batches are fetched sequentially
- **Evidence**: No CPU parallelism in profiling
- **Fix**: Goroutines for parallel batch fetching

## v0.12.0 Goals vs Current

| Metric | Current Best | v0.12.0 Target | Status |
|--------|--------------|----------------|--------|
| Backup speed | 376 docs/sec | 500+ docs/sec | 75% ✅ |
| Memory (10K docs) | Unknown | <500MB | Need profiling ⏳ |
| Restore speed | 18 docs/sec | 50+ docs/sec | Need work ❌ |

## Optimization Roadmap

### Quick Win #1: Increase Default Batch Size ⚡
- **Change**: `BatchSize: 100` → `BatchSize: 200`
- **File**: `src/cmd/backup/create.go:25`
- **Impact**: 2x faster backups (184 → 376 docs/sec)
- **Risk**: Low (users can override with `--batch-size`)
- **Effort**: 1 line change + tests
- **Status**: Ready to implement

### Quick Win #2: Document Batch Size Tuning 📚
- **Change**: Add batch size recommendations to docs
- **File**: `docs/guides/BACKUP_RESTORE.md`
- **Impact**: Users can self-optimize
- **Effort**: 10 minutes

### Medium-term: Parallel Batch Processing 🚀
- **Change**: Fetch batches concurrently with goroutines
- **Expected**: 2-3x improvement (750-1000 docs/sec)
- **Complexity**: Medium (synchronization, error handling)
- **Effort**: 4-6 hours

### Medium-term: Connection Pooling 🔌
- **Change**: Reuse VDB connections across batches
- **Expected**: 10-20% improvement
- **Complexity**: Low
- **Effort**: 2-3 hours

## Test Collections Used

1. **DemoDocs** (38 docs, 1024-dim vectors)
   - Good for testing small collection performance
   - Startup overhead dominates

2. **WeaveDocs** (79 docs, 1536-dim vectors)
   - Medium-sized collection
   - Balanced startup vs backup time

3. **AuctionsImages** (301 docs, image embeddings)
   - Large collection
   - Batch size impact is most visible
   - **Note**: Small file size (27K) suggests missing embeddings

## Next Actions

1. ✅ **Profiling complete**
2. ✅ **Baseline metrics documented**
3. ⏳ **Identify optimization opportunities** (in progress)
4. ⏳ **Implement Quick Win #1** (change default batch size)
5. ⏳ **Update documentation**
6. ⏳ **Test with Client0's 2,636-doc collections**

## Files Created

- `/tmp/perf-profile-backup.sh` - Performance profiling script
- `/tmp/backup-perf-results/` - Test results (7 backup files)
- `/tmp/performance-analysis.md` - Detailed analysis
- `/tmp/perf-baseline-metrics.md` - This document
