# Vector Database Embedding Architecture Fix

**Status:** Planning
**Priority:** High
**Created:** 2026-01-12
**Related:** Multi-VDB Agent Support (d108d5e), VDB Compatibility (7d0a829)

---

## Problem Statement

Collections don't store metadata about which embedding model they use, causing query failures when:
- Collection was created with one embedding model (e.g., `sentence-transformers` with 384 dimensions)
- Queries use a different model (e.g., OpenAI `text-embedding-3-small` with 1536 dimensions)

### Current Errors

**Chroma:**
```
Error (400) InvalidDimension: Embedding dimension 1536 does not match collection dimensionality 384
```

**Milvus:**
```
metric type not match: invalid parameter[expected=COSINE][actual=L2]
```

### Root Cause

The Weave CLI architecture has a fundamental gap:

1. **Collection Creation**: Uses current config to determine embedding model/dimensions/metric
2. **Document Ingestion**: Uses current config to generate embeddings
3. **Queries**: Uses current config (which may be different!)
4. **No Persistence**: Collection metadata doesn't store embedding model information

This means if you:
- Create a collection with config A
- Change config to B
- Query the collection → **FAILS** (dimension/metric mismatch)

---

## Partial Fixes (Completed)

### ✅ Milvus Metric Type (Commit 7d0a829)

**What was fixed:**
- Added `getCollectionMetricType()` to retrieve metric from collection index
- Queries now use the collection's actual metric type (L2, IP, or COSINE)

**Impact:**
- Fixes Milvus "metric type not match" errors ✅
- Works for all Milvus collections (old and new)

### ⚠️ Chroma Dimensions (Commit 7d0a829)

**What was added:**
- Store `vector_dimensions` in collection metadata on creation
- Metadata structure in place for future auto-detection

**Limitations:**
- Old collections don't have this metadata
- Metadata is not yet retrieved/used for queries
- Still requires manual config matching

---

## Complete Solution Architecture

### Phase 1: Embedding Model Metadata Storage

**Goal:** Store complete embedding model information in collection metadata

**Implementation:**

1. **Define Embedding Model Metadata Structure**
```yaml
embedding_config:
  model: "text-embedding-3-small"  # or "sentence-transformers/all-MiniLM-L6-v2"
  provider: "openai"                # or "huggingface", "cohere", etc.
  dimensions: 1536
  normalization: true
  created_at: "2026-01-12T10:00:00Z"
```

2. **Store on Collection Creation**
```go
// In CreateCollection():
metadata := map[string]interface{}{
    "embedding_model": config.EmbeddingModel,
    "embedding_provider": config.EmbeddingProvider,
    "vector_dimensions": config.VectorDimensions,
    "similarity_metric": config.SimilarityMetric,
}
```

3. **Retrieve Before Querying**
```go
// In SearchSemantic():
collectionMeta := getCollectionMetadata(ctx, collectionName)
embeddingModel := collectionMeta.EmbeddingModel
dimensions := collectionMeta.VectorDimensions
metric := collectionMeta.SimilarityMetric
```

### Phase 2: Auto-Detection and Validation

**Features:**

1. **Query-Time Auto-Detection**
   - Retrieve collection's embedding model from metadata
   - Override query config with collection config
   - Generate embeddings using correct model

2. **Validation on Document Add**
   - Check if embedding model matches collection
   - Warn user if mismatch detected
   - Option to force with `--force-model`

3. **Migration Helper**
   - `weave cols migrate <name>` - adds metadata to existing collections
   - Prompts user for embedding model info
   - Updates all collections in batch

### Phase 3: VDB-Specific Implementations

Each VDB needs custom implementation:

#### Chroma
- ✅ Already stores metadata in collection properties
- ✅ Supports custom metadata fields
- ⏭️ Need to retrieve and use metadata in queries

#### Milvus
- ⚠️ No native collection-level metadata storage
- ✅ Can infer metric from index
- ⏭️ Store embedding info in collection description (JSON)
- ⏭️ Parse description to get model info

#### Weaviate
- ✅ Schema-based, explicitly defines vectorizer
- ✅ Already handles this correctly
- ✅ No changes needed

#### Qdrant
- ✅ Supports collection-level configuration
- ⏭️ Can store in payload schema
- ⏭️ Retrieve from collection info

#### Supabase/PostgreSQL
- ⚠️ pgvector doesn't have collection metadata
- ⏭️ Create separate metadata table
- ⏭️ Join on queries to get model info

#### MongoDB
- ✅ Document-based, supports metadata
- ⏭️ Store in collection properties
- ⏭️ Retrieve from collection stats

#### Neo4j
- ⚠️ Graph database, different paradigm
- ⏭️ Store as graph properties
- ⏭️ Retrieve from index configuration

#### Pinecone
- ✅ Supports metadata at index level
- ⏭️ Use index metadata field
- ⏭️ Retrieve before queries

---

## Implementation Plan

### Step 1: Config Extension (1-2 days)

**Tasks:**
- [ ] Add `EmbeddingModel` and `EmbeddingProvider` to `config.VectorDBConfig`
- [ ] Update config parsing to read embedding model settings
- [ ] Add validation for embedding model configurations
- [ ] Update config documentation

**Files:**
- `src/pkg/config/config.go`
- `configs/config.yaml`
- `.env.example`

### Step 2: Metadata Storage (2-3 days)

**Tasks:**
- [ ] Implement metadata storage for each VDB:
  - [ ] Chroma - use collection metadata ✅ (partial)
  - [ ] Milvus - use collection description
  - [ ] Qdrant - use payload schema
  - [ ] Supabase - create metadata table
  - [ ] MongoDB - use collection properties
  - [ ] Neo4j - use graph properties
  - [ ] Pinecone - use index metadata
- [ ] Test metadata persistence across restarts
- [ ] Handle metadata versioning

**Files:**
- `src/pkg/vectordb/*/client.go` (each VDB)
- `src/pkg/vectordb/types.go` (shared metadata types)

### Step 3: Auto-Detection (2-3 days)

**Tasks:**
- [ ] Implement `GetCollectionEmbeddingConfig()` for each VDB
- [ ] Update `SearchSemantic()` to use collection config
- [ ] Add override flags (`--force-model`, `--force-dimensions`)
- [ ] Add validation warnings on document creation
- [ ] Update error messages with helpful suggestions

**Files:**
- `src/pkg/vectordb/*/search.go`
- `src/pkg/vectordb/*/client.go`
- `src/cmd/utils/collection.go`

### Step 4: Migration Tools (1-2 days)

**Tasks:**
- [ ] Create `weave cols migrate` command
- [ ] Interactive prompts for embedding model selection
- [ ] Batch migration for all collections
- [ ] Validation and dry-run mode
- [ ] Migration rollback capability

**Files:**
- `src/cmd/collection/migrate.go`
- `src/cmd/utils/migration.go`

### Step 5: Testing (2-3 days)

**Tasks:**
- [ ] Unit tests for metadata storage/retrieval
- [ ] Integration tests with real VDBs
- [ ] Test dimension mismatches are caught
- [ ] Test metric mismatches are caught
- [ ] Test migration tools
- [ ] Update VDB testing plan

**Files:**
- `tests/vdb_metadata_test.go`
- `docs/planning/VDB_AGENT_TESTING_PLAN.md`

### Step 6: Documentation (1 day)

**Tasks:**
- [ ] Update USER_GUIDE with embedding model configuration
- [ ] Document migration process for existing collections
- [ ] Add troubleshooting guide for dimension/metric errors
- [ ] Update CHANGELOG
- [ ] Create migration announcement/blog post

**Files:**
- `docs/USER_GUIDE.md`
- `docs/TROUBLESHOOTING.md`
- `CHANGELOG.md`

---

## Timeline

**Total Estimated Time:** 10-14 days

- **Week 1:** Config extension, metadata storage, auto-detection
- **Week 2:** Migration tools, testing, documentation

---

## Success Criteria

### Functional Requirements

- [ ] Collections store embedding model metadata on creation
- [ ] Queries automatically use correct embedding model
- [ ] Dimension mismatches detected and reported clearly
- [ ] Metric mismatches detected and reported clearly
- [ ] Migration tool updates existing collections
- [ ] Works across all 10+ supported VDBs

### User Experience

- [ ] No manual configuration needed for queries
- [ ] Clear error messages when mismatches occur
- [ ] One-command migration for existing collections
- [ ] Backward compatible with existing workflows

### Quality

- [ ] All tests passing
- [ ] Comprehensive test coverage (>80%)
- [ ] Documentation complete and accurate
- [ ] No breaking changes for existing users

---

## Workarounds (Until Fix is Complete)

### For Dimension Mismatches

1. **Check collection dimensions:**
   ```bash
   # For Chroma
   weave cols show <collection> --chroma-local

   # For Milvus
   weave cols show <collection> --milvus-local
   ```

2. **Configure matching dimensions in `.env`:**
   ```bash
   # If collection uses 384 dimensions
   CHROMA_VECTOR_DIMENSIONS=384
   MILVUS_VECTOR_DIMENSIONS=384
   ```

3. **Or recreate collection with correct config**

### For Metric Mismatches

1. **Milvus Fix (7d0a829):** ✅ Already retrieves metric from index

2. **For other VDBs:** Configure metric in `.env`:
   ```bash
   MILVUS_SIMILARITY_METRIC=COSINE  # or L2, IP
   QDRANT_SIMILARITY_METRIC=cosine   # or euclidean, dot
   ```

---

## Related Issues

- [Multi-VDB Agent Support](./AGENT_VDB_SUPPORT_AND_PROGRESS.md) - Completed (d108d5e)
- [VDB Compatibility Fixes](https://github.com/maximilien/weave-cli/commit/7d0a829) - Partial fix
- [VDB Testing Plan](./VDB_AGENT_TESTING_PLAN.md) - Waiting for this fix

---

## References

- [OpenAI Embeddings](https://platform.openai.com/docs/guides/embeddings)
- [Sentence Transformers](https://www.sbert.net/)
- [Vector Database Comparison](https://weaviate.io/blog/vector-databases-comparison)
