# Backup Performance Profiling Results - March 15, 2026

## Quick Summary

**Key Finding**: 🚀 **Batch size 200 is 2x faster than batch size 100!**

- **Baseline** (batch=100): 184 docs/sec
- **Optimized** (batch=200): 376 docs/sec
- **Improvement**: +104% (2x faster)

## Files in This Directory

### 1. `perf-profile-backup.sh`
**Profiling script** that tests backup performance across:
- 3 collections (38, 79, 301 documents)
- Different batch sizes (50, 100, 200)
- Compressed vs uncompressed backups

**How to Run**:
```bash
chmod +x perf-profile-backup.sh
./perf-profile-backup.sh
```

**Output**: Creates backup files in `/tmp/backup-perf-results/`

### 2. `performance-analysis.md`
**Detailed profiling analysis** including:
- Test results summary
- Key findings (batch size impact, compression overhead)
- Bottleneck identification (VDB query latency, startup overhead)
- Optimization opportunities
- Performance projections for v0.12.0

### 3. `perf-baseline-metrics.md`
**Baseline metrics summary** with:
- Performance numbers for each test
- Batch size impact comparison
- File sizes and compression ratios
- Optimization roadmap (Quick Wins, medium-term, long-term)

### 4. `optimization-opportunities.md`
**Comprehensive optimization roadmap** with:
- Quick Wins (5 min - 2 hours each)
- Medium-term optimizations (4-8 hours each)
- Long-term optimizations (8+ hours each)
- Code examples and implementation details
- Prioritized roadmap with effort estimates

## Key Results

### Batch Size Impact (301 documents)
| Batch Size | Time | Throughput | vs Baseline |
|------------|------|------------|-------------|
| 50 | 2.69s | 112 docs/sec | -51% ❌ |
| 100 (old default) | 1.64s | 184 docs/sec | baseline |
| **200 (new default)** | **0.80s** | **376 docs/sec** | **+104%** ✅ |

### Bottlenecks Identified
1. **VDB Query Latency** (PRIMARY)
   - Each batch = 1 VDB query
   - Batch 200 = 2 queries vs Batch 100 = 4 queries
   - 50% fewer queries = 2x faster

2. **Startup Overhead** (~3-3.5s fixed cost)
   - Config loading, VDB connection
   - Dominates for small collections

3. **Single-threaded Processing**
   - Batches fetched sequentially
   - No parallelism

## Optimization Roadmap

### ✅ Quick Win #1: Change Default Batch Size to 200
- **Impact**: 2x faster (implemented March 16, 2026)
- **Effort**: 5 minutes
- **Risk**: Very low

### 🚀 Sprint 3 Optimizations (Next Week)
1. **Parallel batch processing** → 2-3x improvement
2. **Connection pooling** → 10-20% improvement
3. **Streaming JSON/compression** → 5-10% improvement

**Expected v0.12.0 Result**: 950 docs/sec (exceeds 500+ target by 90%)

## Performance Progress

| Stage | Throughput | vs Baseline | Status |
|-------|------------|-------------|--------|
| v0.11.3 (batch=100) | 184 docs/sec | - | ✅ Measured |
| **Quick Win (batch=200)** | **376 docs/sec** | **2.0x** | **✅ Implemented** |
| v0.12.0 Target | 500+ docs/sec | 2.7x | 🎯 Goal |
| Projected v0.12.0 | 950 docs/sec | 5.2x | 🚀 Exceeds! |

## References

- **Main Report**: `../PROFILING_MARCH_15_2026.md`
- **Tomorrow's Tasks**: `../../planning/TOMORROW_MARCH_16_2026.md`
- **PLAN.md**: `../../../PLAN.md`
- **BACKUP_RESTORE.md**: `../../../guides/BACKUP_RESTORE.md` (Performance section)

## Test Environment

- **Platform**: macOS (Darwin)
- **Version**: v0.11.3-5-g2f7dc35
- **Date**: March 15, 2026
- **Collections Tested**:
  - DemoDocs (38 docs, 1024-dim vectors)
  - WeaveDocs (79 docs, 1536-dim vectors)
  - AuctionsImages (301 docs, 1536-dim vectors)

---

**Profiling by**: Claude Code
**Status**: ✅ Complete
**Next**: Implement Quick Win #1 (March 16, 2026)
