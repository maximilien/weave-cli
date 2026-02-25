# Weekend & Next Week Plan - Post v0.9.17 🚀

## ✅ Completed Today (Feb 6, 2026)

### Morning Session
- Completed Phase 3B: EmbeddingPipeline (127 lines + 267 test lines)
- Completed Phase 3C: Integration Tests (296 lines, 5 scenarios)
- Completed Phase 3D: CLI Command (255 lines)
- All 28 tests passing

### Afternoon Session
- **Debugged CLI registration mystery** - Found missing `ReEmbedCmd` in `src/cmd/collection.go`
- **Fixed issue #28** - One line fix resolved CLI visibility
- **Released v0.9.17** - Production-ready batch re-embedding
- Updated CHANGELOG, closed issue, tagged and pushed release

**Total Time**: ~4 hours from "tested OK" to v0.9.17 shipped

---

## 🎯 Weekend Plan (Optional - Client0 Prep)

### Saturday (2-3 hours) - OSS Embedding Providers

**Goal**: Add sentence-transformers support to enable Client0 testing

#### Task 1: sentence-transformers Provider (1.5 hours)
```bash
# File: src/pkg/reembedding/providers/sentence_transformers.go
```

**Implementation:**
- Python bridge to sentence-transformers library
- Batch embedding generation (100+ docs)
- Support for models:
  - `sentence-transformers/all-mpnet-base-v2` (768d)
  - `sentence-transformers/all-minilm-l6-v2` (384d)
  - `sentence-transformers/all-minilm-l12-v2` (384d)

**Integration:**
- Update `EmbeddingPipeline.ProcessBatch()` to detect provider
- Add provider-specific embedding generation
- Test with mock documents

**Deliverable:** Commit "feat: add sentence-transformers provider for re-embedding"

#### Task 2: Provider Tests (30-45 min)
- Unit tests for sentence-transformers provider
- Integration test with re-embed workflow
- Error handling (missing Python, missing library)

**Deliverable:** Commit "test: add sentence-transformers provider tests"

---

### Sunday (2-3 hours) - Ollama Provider

**Goal**: Add Ollama local embedding support

#### Task 1: Ollama Provider (1.5 hours)
```bash
# File: src/pkg/reembedding/providers/ollama.go
```

**Implementation:**
- HTTP client for Ollama API
- Batch embedding support
- Models:
  - `nomic-embed-text` (768d)
  - `mxbai-embed-large` (1024d)
  - `snowflake-arctic-embed` (1024d)

**Integration:**
- Update `EmbeddingPipeline.ProcessBatch()` for Ollama
- Connection validation (check Ollama server running)
- Graceful error messages

**Deliverable:** Commit "feat: add Ollama provider for re-embedding"

#### Task 2: Documentation (30-45 min)
- Update `docs/BATCH_REEMBEDDING_SPEC.md` with provider status
- Add provider setup instructions
- Update examples in CHANGELOG

**Deliverable:** Commit "docs: update re-embedding provider documentation"

---

## 📋 Next Week Priority Tasks

### Monday - Client0 Production Testing

**Morning Session (2-3 hours)**
1. **Client0 Dataset Preparation**
   - Ensure 3,518 docs ingested with OpenAI embeddings
   - Verify source collection health
   - Document baseline metrics

2. **First Re-Embedding Test (OpenAI → OpenAI)**
   ```bash
   # Sanity check - same dimensions
   weave collection reembed Client0_PDFs \
     --new-embedding text-embedding-3-large \
     --output Client0_PDFs_Large \
     --milvus-local
   ```
   - Expected: 3,518 docs, ~15-18 minutes
   - Validate: Dimension change 1536d → 3072d
   - Compare search results quality

**Afternoon Session (2-3 hours)**
3. **OSS Model Testing (if sentence-transformers ready)**
   ```bash
   weave collection reembed Client0_PDFs \
     --new-embedding sentence-transformers/all-mpnet-base-v2 \
     --output Client0_PDFs_OSS \
     --milvus-local
   ```
   - Expected: 3,518 docs, ~15-18 minutes
   - Validate: Dimension change 1536d → 768d
   - **Critical**: Compare search quality vs OpenAI

---

### Tuesday - Performance Optimization

**Goal**: Ensure 200+ docs/min throughput

1. **Profiling** (1-2 hours)
   - Add benchmark tests
   - Profile CPU/memory during re-embedding
   - Identify bottlenecks (API calls, VDB writes, etc.)

2. **Batch Size Tuning** (1-2 hours)
   - Test batch sizes: 50, 100, 200, 500
   - Measure throughput at each size
   - Document optimal settings per provider

3. **Concurrency Exploration** (2-3 hours)
   - Parallel batch processing (goroutines)
   - Concurrent embedding API calls
   - Rate limiting for API quotas

**Deliverable:** Commit "perf: optimize batch re-embedding throughput"

---

### Wednesday - Error Recovery & Resilience

**Goal**: Production-grade error handling

1. **Checkpoint System** (2-3 hours)
   ```bash
   # Resume failed re-embedding
   weave collection reembed Client0_PDFs \
     --new-embedding text-embedding-3-large \
     --output Client0_PDFs_Large \
     --resume
   ```
   - Save progress every N batches
   - Detect partial completion
   - Resume from last checkpoint

2. **Retry Logic** (1-2 hours)
   - Exponential backoff for API failures
   - Configurable retry limits
   - Per-batch error tracking

**Deliverable:** Commit "feat: add checkpoint and retry for re-embedding"

---

### Thursday - Multi-VDB Support

**Goal**: Test re-embedding across all VDBs

1. **VDB Compatibility Testing** (3-4 hours)
   - Test with Weaviate (Cloud + Local)
   - Test with Milvus (Cloud + Local)
   - Test with Qdrant, MongoDB, Chroma
   - Document any VDB-specific quirks

2. **Cross-VDB Re-Embedding** (1-2 hours)
   ```bash
   # Re-embed from Weaviate to Milvus
   weave collection reembed Docs --weaviate-cloud \
     --new-embedding text-embedding-3-large \
     --output Docs_Large --milvus-cloud
   ```

**Deliverable:** Commit "feat: support cross-VDB re-embedding workflows"

---

### Friday - Client0 Handoff & Documentation

**Goal**: Production-ready for Client0

1. **Client0 Validation Report** (2-3 hours)
   - Document all model comparisons
   - Search quality metrics
   - Performance benchmarks
   - Cost analysis (OSS vs proprietary)

2. **User Documentation** (2-3 hours)
   - Update README with re-embedding examples
   - Create `docs/REEMBEDDING_GUIDE.md`
   - Add troubleshooting section
   - Video walkthrough (optional)

3. **v0.9.18 Release** (1 hour)
   - Tag release with OSS provider support
   - Update CHANGELOG
   - Announce to Client0

**Deliverable:** v0.9.18 shipped with full OSS support

---

## 📊 Metrics to Track

### Performance Targets
- ✅ Throughput: 200+ docs/min (met in tests)
- ⏳ 3,518 docs: <20 minutes target
- ⏳ API errors: <1% failure rate
- ⏳ Memory usage: <2GB peak

### Client0 Success Criteria
1. Successfully re-embed 3,518 docs with 3+ models
2. Search quality comparable or better with OSS models
3. Total validation time: <3 days (vs 3 weeks without re-embedding)
4. Cost reduction: 80%+ by switching to OSS

---

## 🚧 Known Blockers & Risks

### Weekend Risks
- **Python bridge complexity**: sentence-transformers requires Python
  - Mitigation: Use subprocess or embed Python interpreter
  - Fallback: Docker container with Python + sentence-transformers

- **Ollama setup**: Users need Ollama server running
  - Mitigation: Clear setup docs, health check in CLI
  - Fallback: Skip Ollama, focus on sentence-transformers

### Next Week Risks
- **API rate limits**: OpenAI may throttle during 3,518 doc tests
  - Mitigation: Batch size tuning, exponential backoff
  - Fallback: Test with smaller dataset first

- **VDB performance**: Milvus/Weaviate may be slow for batch inserts
  - Mitigation: Profile and optimize batch sizes
  - Fallback: Document recommended batch sizes per VDB

---

## 🎯 Success Definition

**Weekend:**
- OSS providers (sentence-transformers or Ollama) working
- Tests passing
- Ready for Client0 Monday testing

**Next Week:**
- Client0 successfully tests 3+ embedding models
- Performance meets 200+ docs/min target
- v0.9.18 released with OSS support
- Client0 saves 15-25 hours during validation

---

## 📝 Optional Enhancements (Nice to Have)

### Low Priority
1. **Embedding Comparison Tool**
   ```bash
   weave collection compare Col1 Col2 --query "test query"
   # Shows search results side-by-side
   ```

2. **Cost Estimator**
   ```bash
   weave collection estimate-cost Client0_PDFs \
     --new-embedding text-embedding-3-large
   # Shows: API cost, time estimate, token usage
   ```

3. **Automatic Model Selection**
   ```bash
   weave collection suggest-embedding Client0_PDFs
   # Analyzes collection, suggests best model
   ```

---

## 📚 Reference

### Key Files
- `src/pkg/reembedding/` - Core implementation
- `src/cmd/collection/re_embed.go` - CLI command
- `docs/planning/BATCH_REEMBEDDING_SPEC.md` - Technical spec
- `CHANGELOG.md` - v0.9.17 release notes

### Key Commits (Today)
- `8c38608` - feat: complete batch re-embedding (Phase 3B-C)
- `0480b69` - docs: prepare release v0.9.17
- `fb6155e` - fix: register ReEmbedCmd (THE FIX!)
- `2c6a441` - docs: remove known issues from CHANGELOG

### GitHub
- Issue #28: CLI registration (CLOSED)
- Tag: v0.9.17 (SHIPPED)
- Next: v0.9.18 with OSS providers

---

**Session Time Today**: 4 hours
**Weekend Estimate**: 4-6 hours (optional)
**Next Week Estimate**: 12-15 hours
**Total to Production**: 16-21 hours remaining

**Status**: v0.9.17 shipped and ready for Client0! 🎉
