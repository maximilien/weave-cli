# OSS Stack Feature Requests - Client0
*Vintage Camera Auction RAG Platform*

## Client Context

**Use Case**: Fully OSS AI stack for RAG platform
- **Embeddings**: sentence-transformers (all-mpnet-base-v2, etc.)
- **LLMs**: Ollama (llama3.1:8b, kimi-k2)
- **Vector DB**: Milvus Local
- **Dataset**: 11 PDF catalogs → 3,518 docs, 2,945 images
- **Goal**: 3-week OSS validation period

**Pain Point**: Testing different embeddings requires full re-ingestion (5+ hours for 11 PDFs)

---

## Feature Requests (Prioritized by ROI)

### #1 - Batch Re-Embedding Command ⭐⭐⭐
**Priority**: HIGHEST (saves 10-20 hours during OSS validation)

**Problem**:
- Currently takes 5+ hours to re-ingest 11 PDFs when testing different embeddings
- Need to re-compute vectors from existing text chunks (without re-ingesting PDFs)
- Target: ~15 minutes for re-embedding (vs 5+ hours for full re-ingestion)

**Proposed Command**:
```bash
weave collection re-embed MyCollection \
  --new-embedding sentence-transformers/all-mpnet-base-v2 \
  --output MyCollection_OSS
```

**What it does**:
1. Read existing text chunks from source collection
2. Generate new embeddings using different model
3. Create new collection with new vectors
4. Preserve all metadata, document structure

**ROI**:
- Saves 4-5 hours per embedding test
- Client needs to test 3-5 embedding models → 15-20 hours saved
- Critical for OSS validation timeline

**Technical Approach**:
- Read documents from source collection (paginated batches)
- Extract text content (skip re-chunking)
- Generate embeddings with new model
- Upsert to target collection
- Progress tracking + estimated time remaining

**Implementation Estimate**: 6-8 hours
- Collection reader (2h)
- Embedding pipeline (2h)
- Progress tracking (1h)
- Testing (2h)
- Documentation (1h)

---

### #2 - Auto-Detect Embedding Dimensions ⭐⭐
**Priority**: HIGH (improves UX, reduces errors)

**Problem**:
- Users must manually specify vector dimensions
- Common mistake: wrong dimensions → collection creation fails
- sentence-transformers models have well-known dimensions

**Proposed Command**:
```bash
weave collection create MyCol \
  --embedding sentence-transformers/all-mpnet-base-v2
# Auto-detects 768 dimensions
```

**What it does**:
1. Parse embedding model name
2. Look up known dimensions from registry
3. Auto-fill `--dimensions` parameter
4. Allow manual override if needed

**Known Models Registry**:
```
sentence-transformers/all-mpnet-base-v2     → 768
sentence-transformers/all-MiniLM-L6-v2      → 384
sentence-transformers/all-MiniLM-L12-v2     → 384
text-embedding-3-small (OpenAI)             → 1536
text-embedding-3-large (OpenAI)             → 3072
text-embedding-ada-002 (OpenAI)             → 1536
ollama/nomic-embed-text                     → 768
```

**Technical Approach**:
- Create `src/pkg/embeddings/model_registry.go`
- Add lookup function: `GetModelDimensions(modelName string) (int, error)`
- Integrate into collection creation flow
- Add `--auto-detect-dimensions` flag (default: true)

**Implementation Estimate**: 3-4 hours
- Model registry (1h)
- Integration (1h)
- Testing (1h)
- Documentation (1h)

---

### #3 - Embedding Comparison Reports ⭐⭐
**Priority**: MEDIUM-HIGH (validation tool)

**Problem**:
- No systematic way to compare embedding quality
- Manual testing is time-consuming and inconsistent
- Need reproducible comparison metrics

**Proposed Command**:
```bash
weave embeddings compare \
  --collections "MyCol_OpenAI,MyCol_OSS,MyCol_Nomic" \
  --queries "query1,query2,query3" \
  --output comparison.md
```

**What it does**:
1. Run same queries against multiple collections
2. Measure: precision@k, recall@k, latency, relevance
3. Generate markdown report with side-by-side results
4. Include: top-k results, similarity scores, timing

**Report Format** (comparison.md):
```markdown
# Embedding Model Comparison

## Test Configuration
- Collections: MyCol_OpenAI (1536d), MyCol_OSS (768d), MyCol_Nomic (768d)
- Queries: 3 test queries
- Date: 2026-02-05

## Results Summary

| Model | Avg Precision@5 | Avg Latency | Cost |
|-------|----------------|-------------|------|
| OpenAI (text-3-large) | 0.92 | 45ms | $0.13/1M |
| sentence-transformers | 0.87 | 12ms | $0 (OSS) |
| Nomic | 0.85 | 18ms | $0 (OSS) |

## Query 1: "vintage Leica M3"
...
```

**Technical Approach**:
- Create `src/cmd/embeddings/compare.go`
- Query executor (parallel queries)
- Metrics calculator (precision, recall, latency)
- Report generator (markdown template)

**Implementation Estimate**: 6-8 hours
- Query execution (2h)
- Metrics calculation (2h)
- Report generation (2h)
- Testing (1h)
- Documentation (1h)

---

### #4 - OSS Quick-Start Template ⭐
**Priority**: MEDIUM (onboarding tool)

**Problem**:
- Setting up OSS stack requires multiple manual steps
- Docker Compose + config files + environment setup
- High barrier for OSS users

**Proposed Command**:
```bash
weave quickstart oss \
  --embedding sentence-transformers/all-mpnet-base-v2 \
  --llm ollama:llama3.1:8b \
  --vdb milvus-local
```

**What it does**:
1. Generate `docker-compose.yml` for OSS stack
2. Create `.env` with OSS-specific settings
3. Create `config.yaml` with recommended defaults
4. Generate sample pipeline YAML
5. Provide next-step instructions

**Generated Files**:
```
project/
├── docker-compose.yml    # Milvus + Ollama containers
├── .env                  # OSS configuration
├── config.yaml           # Weave config
├── oss-pipeline.yaml     # Sample pipeline
└── README-OSS.md         # Getting started guide
```

**Technical Approach**:
- Create `src/cmd/quickstart.go`
- Template system for docker-compose, configs
- Ollama model pull instructions
- sentence-transformers setup guide

**Implementation Estimate**: 8-10 hours
- Template creation (3h)
- CLI command (2h)
- Docker Compose configs (2h)
- Testing (2h)
- Documentation (1h)

---

### #5 - Ollama Model Auto-Discovery ⭐
**Priority**: LOW-MEDIUM (nice-to-have)

**Problem**:
- Users must manually type Ollama model names
- Easy to make typos (llama3.1:8b vs llama3:8b)
- Interactive wizard doesn't show available models

**Proposed Enhancement**:
```bash
weave config agents

# Interactive wizard shows:
Select LLM:
  1. OpenAI (gpt-4, gpt-3.5-turbo)
  2. Ollama - Locally Installed:
     • llama3.1:8b
     • kimi-k2
     • nomic-embed-text
  3. Custom model name
```

**What it does**:
1. Call `ollama list` to get installed models
2. Parse output and show in interactive wizard
3. Validate model exists before proceeding
4. Show model size, last used timestamp

**Technical Approach**:
- Create `src/pkg/llm/ollama/discovery.go`
- Execute `ollama list` command
- Parse model names, sizes, timestamps
- Integrate into `weave config agents` wizard

**Implementation Estimate**: 4-5 hours
- Ollama discovery (2h)
- Integration (1h)
- Testing (1h)
- Documentation (1h)

---

## Implementation Plan

### Phase 1: Quick Wins (Thu/Fri - 4 hours)
**Goal**: Deliver highest ROI feature for client validation

✅ **#2 - Auto-Detect Embedding Dimensions** (3-4 hours)
- Easiest to implement
- Immediate UX improvement
- Reduces configuration errors
- Foundation for other features

**Deliverables**:
- Model registry with 10+ known models
- Auto-detection in collection creation
- Tests + documentation
- Release as v0.9.16

---

### Phase 2: High Impact (Weekend - 2 hours total)
**Goal**: Plan and start #1 (batch re-embedding)

**Saturday (1 hour)**:
- Design batch re-embedding architecture
- Create detailed technical spec
- Identify reusable components from existing pipeline

**Sunday (1 hour)**:
- Implement collection reader
- Test reading existing documents from Milvus
- Create progress tracking structure

---

### Phase 3: Core Feature (Next Week - 6-8 hours)
**Goal**: Complete #1 - Batch Re-Embedding

✅ **#1 - Batch Re-Embedding Command** (6-8 hours)
- Collection reader (reuse from Phase 2)
- Embedding pipeline
- Progress tracking
- Testing with client dataset
- Documentation

**Timeline**:
- Mon-Tue: Implementation (4-5 hours)
- Wed: Testing + refinement (2 hours)
- Thu: Documentation + release (1 hour)

**Target**: v0.9.17 with batch re-embedding

---

### Phase 4: Validation Tools (Following Week - 6-8 hours)
**Goal**: Deliver comparison and discovery features

✅ **#3 - Embedding Comparison Reports** (6-8 hours)
✅ **#5 - Ollama Model Auto-Discovery** (4-5 hours)

**Target**: v0.9.18

---

### Phase 5: Onboarding (Future - 8-10 hours)
**Goal**: Lower barrier for OSS users

✅ **#4 - OSS Quick-Start Template** (8-10 hours)

**Target**: v0.10.0 (major release with OSS focus)

---

## Success Metrics

**For Client0**:
- ✅ Reduce embedding test time: 5 hours → 15 minutes (20x faster)
- ✅ Complete 3-week OSS validation on schedule
- ✅ Test 3-5 embedding models efficiently
- ✅ Generate comparison reports for stakeholders

**For OSS Community**:
- ✅ Auto-detection reduces config errors by ~80%
- ✅ Quickstart template reduces setup time: 2 hours → 15 minutes
- ✅ Ollama discovery improves DX for local LLM users

---

## Technical Notes

### Reusable Components
From existing codebase:
- ✅ Document pipeline (`src/pkg/pipeline`)
- ✅ Embedding generation (`src/pkg/llm`)
- ✅ Collection operations (`src/pkg/vectordb`)
- ✅ Progress tracking (`src/pkg/progress`)

### New Components Needed
- `src/pkg/embeddings/model_registry.go` - Model dimension lookup
- `src/cmd/collection/re_embed.go` - Batch re-embedding command
- `src/cmd/embeddings/compare.go` - Comparison reports
- `src/cmd/quickstart.go` - OSS template generator
- `src/pkg/llm/ollama/discovery.go` - Model auto-discovery

### Testing Strategy
- Unit tests for each new component
- Integration tests with Milvus local
- Manual testing with client's use case (11 PDFs)
- Performance benchmarks (re-embedding speed)

---

## Client Communication

**Beta Testing Offer**: "Happy to beta test any of these features!"
- Provide early access to #1 and #2
- Get feedback during 3-week validation period
- Iterate based on real-world usage

**Metrics Request**: "Will share results/metrics after validation"
- Request: embedding comparison results
- Request: time savings from batch re-embedding
- Use for case study (anonymized as "Client0")

---

## Next Steps (Immediate)

1. ✅ Create GitHub issues (sanitized, no client names)
2. ✅ Start Phase 1: Auto-detect embedding dimensions (3-4 hours)
3. ✅ Update CHANGELOG.md with upcoming features
4. ✅ Create v0.9.16 release with auto-detection
5. 📋 Schedule check-in with client after Phase 1 delivery

**Target Timeline**:
- Thu/Fri: Deliver #2 (auto-detection)
- Weekend: Plan #1 (re-embedding)
- Next week: Deliver #1 (batch re-embedding)
- Following week: Deliver #3, #5
- Future: Deliver #4 (quickstart)

LFG! 🚀
