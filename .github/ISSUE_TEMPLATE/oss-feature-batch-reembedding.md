---
name: OSS Feature - Batch Re-Embedding
about: Add batch re-embedding command to test different embedding models efficiently
title: 'feat: Add batch re-embedding command for collections'
labels: enhancement, oss, high-priority
assignees: ''
---

## Feature Request: Batch Re-Embedding Command

**Priority**: ⭐⭐⭐ HIGHEST

### Problem
When testing different embedding models with OSS stacks (sentence-transformers, Ollama), users must re-ingest entire datasets to change embeddings. For large datasets, this can take hours.

**Current workflow** (inefficient):
1. Ingest 11 PDFs → 3,518 documents (5+ hours)
2. Test embedding model A
3. Delete collection
4. Re-ingest same 11 PDFs with embedding model B (5+ hours again)
5. Repeat for each model (15-20 hours total for 3-5 models)

**Desired workflow** (efficient):
1. Ingest once (5 hours)
2. Re-embed with model B (15 minutes)
3. Re-embed with model C (15 minutes)
4. Compare results

**Time savings**: 5 hours → 15 minutes per embedding test (20x faster)

### Proposed Solution

Add `weave collection re-embed` command:

```bash
weave collection re-embed SourceCollection \
  --new-embedding sentence-transformers/all-mpnet-base-v2 \
  --output TargetCollection
```

**What it does**:
1. Read existing text chunks from source collection
2. Generate new embeddings using different model
3. Create new collection with new vectors
4. Preserve all metadata, document structure

### Use Case

**OSS AI Stack Testing**:
- Embeddings: sentence-transformers (all-mpnet-base-v2, all-MiniLM-L6-v2, etc.)
- LLMs: Ollama (llama3.1:8b, kimi-k2)
- Vector DB: Milvus Local, Weaviate Local
- Dataset: PDF catalogs → thousands of documents

**Validation workflow**:
1. Test 3-5 different embedding models
2. Generate comparison reports
3. Select best model for production

### Technical Approach

**Components**:
- Collection reader (paginated batches)
- Embedding pipeline (reuse existing from `src/pkg/llm`)
- Progress tracking (show % complete, ETA)
- Batch upsert to target collection

**Example output**:
```
Re-embedding SourceCollection → TargetCollection
Embedding model: sentence-transformers/all-mpnet-base-v2
Dimensions: 768 (auto-detected)

Progress: [██████████████░░░░░░] 70% (2,463/3,518 docs)
Speed: 234 docs/min
Estimated time remaining: 4m 30s
```

### Implementation Checklist

- [ ] Create `src/cmd/collection/re_embed.go`
- [ ] Collection reader (paginated)
- [ ] Embedding generation pipeline
- [ ] Progress tracking UI
- [ ] Batch upsert to target collection
- [ ] Handle errors gracefully (resume on failure)
- [ ] Unit tests
- [ ] Integration tests with Milvus/Weaviate
- [ ] Documentation
- [ ] Example in README

### Success Metrics

- ✅ Reduce re-embedding time: 5 hours → 15 minutes (20x faster)
- ✅ Support all embedding models (OpenAI, sentence-transformers, Ollama)
- ✅ Handle large collections (10,000+ documents)
- ✅ Resume on failure (checkpoint progress)

### Related Features

- #TBD - Auto-detect embedding dimensions
- #TBD - Embedding comparison reports

### Community Impact

**Who benefits**:
- OSS AI stack users (sentence-transformers, Ollama)
- Researchers testing different embedding models
- Teams migrating between embedding providers
- Cost-conscious users (avoid re-ingestion compute costs)

**Estimated time saved per user**: 10-20 hours during model selection phase
