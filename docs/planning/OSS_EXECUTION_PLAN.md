# OSS Features - Execution Plan
*Client0 - Vintage Camera Auction RAG Platform*

## 🎯 Quick Summary

**Client Use Case**: Fully OSS AI stack validation (3 weeks)
- sentence-transformers + Ollama + Milvus Local
- 11 PDFs → 3,518 docs, 2,945 images
- Testing 3-5 different embedding models

**Highest ROI Feature**: Batch re-embedding (saves 10-20 hours)

---

## 📋 Feature Prioritization

| # | Feature | Priority | Time | ROI | When |
|---|---------|----------|------|-----|------|
| **#2** | **Auto-Detect Dimensions** | ⭐⭐ HIGH | 3-4h | HIGH | **Thu/Fri** |
| **#1** | **Batch Re-Embedding** | ⭐⭐⭐ HIGHEST | 6-8h | VERY HIGH | **Next Week** |
| #3 | Embedding Comparison | ⭐⭐ MED-HIGH | 6-8h | MEDIUM | Week After |
| #5 | Ollama Auto-Discovery | ⭐ LOW-MED | 4-5h | LOW | Week After |
| #4 | OSS Quickstart | ⭐ MEDIUM | 8-10h | MEDIUM | Future |

---

## 🚀 Phase 1: Quick Win (Thu/Fri - 4 hours)

### #2 - Auto-Detect Embedding Dimensions ✅

**Why First**:
- ✅ Easiest to implement (3-4 hours)
- ✅ Immediate UX improvement
- ✅ Foundation for re-embedding feature
- ✅ Can ship as v0.9.16 this week

**Implementation**:
```bash
# Create model registry
src/pkg/embeddings/model_registry.go

# Known models (15+)
sentence-transformers/all-mpnet-base-v2     → 768
sentence-transformers/all-MiniLM-L6-v2      → 384
text-embedding-3-small                      → 1536
text-embedding-3-large                      → 3072
ollama/nomic-embed-text                     → 768
```

**Deliverable**:
```bash
# Before
weave collection create MyCol \
  --embedding sentence-transformers/all-mpnet-base-v2 \
  --dimensions 768  # ← Manual

# After
weave collection create MyCol \
  --embedding sentence-transformers/all-mpnet-base-v2
# Auto-detects 768 dimensions ✨
```

**Tasks**:
1. Create `src/pkg/embeddings/model_registry.go` (1h)
2. Add 15+ models to registry (30min)
3. Integrate into collection create (1h)
4. Tests + documentation (1-1.5h)
5. Build, test, commit (30min)

**Ship**: v0.9.16 with auto-detection

---

## 🔄 Phase 2: Planning (Weekend - 2 hours)

### Saturday (1 hour) - Design Re-Embedding

**Tasks**:
- Design batch re-embedding architecture
- Identify reusable components from pipeline
- Create technical spec document
- Plan testing strategy (with client's 11 PDFs)

**Output**: `docs/planning/BATCH_REEMBEDDING_SPEC.md`

### Sunday (1 hour) - Prototype Core

**Tasks**:
- Implement collection reader (paginated)
- Test reading existing docs from Milvus
- Create progress tracking structure
- Validate approach with sample data

**Output**: Working prototype (not merged)

---

## 💪 Phase 3: High Impact (Next Week - 6-8 hours)

### #1 - Batch Re-Embedding Command ✅

**Why**: Saves client 10-20 hours during validation

**Implementation**:
```bash
weave collection re-embed SourceCollection \
  --new-embedding sentence-transformers/all-mpnet-base-v2 \
  --output TargetCollection

# Progress tracking
Re-embedding SourceCollection → TargetCollection
Embedding model: sentence-transformers/all-mpnet-base-v2
Dimensions: 768 (auto-detected)

Progress: [██████████████░░░░░░] 70% (2,463/3,518 docs)
Speed: 234 docs/min
ETA: 4m 30s
```

**Timeline**:
- **Mon-Tue**: Core implementation (4-5h)
  - Collection reader (reuse from weekend)
  - Embedding pipeline
  - Batch upsert
  - Progress UI
- **Wed**: Testing + refinement (2h)
  - Test with sample data
  - Performance optimization
  - Error handling
- **Thu**: Documentation + release (1h)
  - Update README
  - Add examples
  - Release v0.9.17

**Reusable Components**:
- ✅ `src/pkg/pipeline` - Document pipeline
- ✅ `src/pkg/llm` - Embedding generation
- ✅ `src/pkg/vectordb` - Collection operations
- ✅ `src/pkg/progress` - Progress tracking

**New Files**:
- `src/cmd/collection/re_embed.go`
- `src/cmd/collection/re_embed_test.go`

**Ship**: v0.9.17 with batch re-embedding

---

## 📊 Phase 4: Validation Tools (Week After - 10-13 hours)

### #3 - Embedding Comparison Reports (6-8h)

**Command**:
```bash
weave embeddings compare \
  --collections "Col_OpenAI,Col_OSS,Col_Nomic" \
  --queries "query1,query2,query3" \
  --output comparison.md
```

**Deliverables**:
- Markdown reports with metrics
- Precision@k, recall@k, latency
- Cost comparison (OpenAI vs OSS)
- Side-by-side results

### #5 - Ollama Auto-Discovery (4-5h)

**Enhancement**:
```bash
weave config agents

Select LLM:
  2. Ollama - Locally Installed:
     • llama3.1:8b (4.7GB)
     • kimi-k2 (15GB)
     • nomic-embed-text (274MB)
```

**Ship**: v0.9.18

---

## 🎁 Phase 5: Onboarding (Future - 8-10 hours)

### #4 - OSS Quickstart Template

**Command**:
```bash
weave quickstart oss \
  --embedding sentence-transformers/all-mpnet-base-v2 \
  --llm ollama:llama3.1:8b
```

**Generates**:
- `docker-compose.yml`
- `.env`
- `config.yaml`
- `oss-pipeline.yaml`
- `README-OSS.md`

**Ship**: v0.10.0 (major release - OSS focus)

---

## 📅 Timeline Summary

```
Week 1 (This Week):
  Thu/Fri    → #2 Auto-detect dimensions (3-4h) → v0.9.16
  Weekend    → Planning + prototype (2h)

Week 2 (Next Week):
  Mon-Thu    → #1 Batch re-embedding (6-8h) → v0.9.17

Week 3 (Following Week):
  Mon-Thu    → #3 Comparison + #5 Discovery (10-13h) → v0.9.18

Future:
  TBD        → #4 OSS Quickstart (8-10h) → v0.10.0
```

**Total effort**: ~29-35 hours over 3 weeks

---

## 🎯 Success Metrics

**For Client0**:
- ✅ Reduce embedding test time: 5h → 15min (20x faster)
- ✅ Complete 3-week OSS validation on schedule
- ✅ Test 3-5 embedding models efficiently
- ✅ Generate comparison reports for stakeholders

**For OSS Community**:
- ✅ Auto-detection reduces config errors by ~80%
- ✅ Re-embedding enables faster experimentation
- ✅ Comparison reports enable data-driven decisions
- ✅ Quickstart reduces setup time: 2h → 15min

---

## 🤝 Client Communication

**Beta Testing**: Client volunteered to test features
- Share early builds of #2 (auto-detect)
- Get feedback on #1 (re-embedding) during validation
- Request metrics/results for case study

**Check-ins**:
- After Phase 1: Share v0.9.16 (auto-detect)
- During Phase 3: Beta test v0.9.17 (re-embedding)
- End of validation: Request metrics, testimonial

---

## 🚀 Next Steps (Immediate)

1. ✅ Commit planning docs + GitHub issues
2. ✅ **Start Phase 1**: Auto-detect dimensions (3-4h)
3. ✅ Ship v0.9.16 this week
4. 📋 Weekend: Plan + prototype re-embedding
5. 📋 Next week: Ship v0.9.17 (batch re-embedding)

**LFG!** 🚀
