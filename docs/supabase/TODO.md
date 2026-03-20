# Supabase Support - Remaining TODOs

Prioritized list of remaining Supabase limitations to address (easiest to hardest).

## Current Limitations

From `docs/VDB_SUPPORT.md`:
1. **Vectorizers**: Currently only supports OpenAI embeddings and manual embeddings
2. **Collection Names**: Automatically normalizes names (underscores → hyphens)
3. **Performance**: BM25 search could be faster with GIN indexes (optional optimization)

---

## TODO List (Easiest → Hardest)

### 1. ~~Fix Collection Name Normalization~~ (DONE) ⭐

**Priority**: High — **FIXED**
**Effort**: 1-2 hours
**Impact**: User experience (eliminates unexpected naming behavior)

**Problem:**
- Supabase adapter normalizes collection names: `my_collection` → `my-collection`
- Users create `TestCollection_1` but it becomes `testcollection-1`
- Inconsistent with Weaviate behavior

**Why Easy:**
- Simple code change in `getTableName()` function
- No schema migration needed
- Just preserve original naming

**Solution:**
```go
// Current: src/pkg/vectordb/supabase/adapter.go:213-219
func (a *Adapter) getTableName(collectionName string) string {
    tableName := strings.ToLower(collectionName)
    tableName = strings.ReplaceAll(tableName, "-", "_")  // REMOVE this line
    tableName = strings.ReplaceAll(tableName, " ", "_")
    return fmt.Sprintf("collection_%s", tableName)
}

// Better: Keep original casing and characters (PostgreSQL supports it)
func (a *Adapter) getTableName(collectionName string) string {
    // Only replace spaces with underscores for valid table names
    tableName := strings.ReplaceAll(collectionName, " ", "_")
    // Quote table name to preserve case and special chars
    return fmt.Sprintf("collection_%s", tableName)
}
```

**Testing:**
- Create collection `MyTest_Collection`
- Verify it's stored as `collection_MyTest_Collection`
- Verify list/get/delete all work with original name

**Files to Change:**
- `src/pkg/vectordb/supabase/adapter.go` - Update `getTableName()`
- `src/pkg/vectordb/supabase/collections.go` - Ensure SQL uses quoted identifiers
- `tests/supabase_integration_test.go` - Add test for name preservation

---

### 2. Add More Embedding Provider Support (MEDIUM) ⭐⭐

**Priority**: Medium
**Effort**: 4-6 hours
**Impact**: Feature parity with Weaviate

**Problem:**
- Supabase only supports OpenAI embeddings
- Weaviate supports: Cohere, Hugging Face, Google PaLM, AWS Bedrock, Jina AI
- Users can't use their preferred embedding provider

**Why Medium:**
- Need to add multiple LLM client implementations
- Need to configure vectorizers in collection schema
- Need to handle different embedding dimensions
- Testing requires API keys for each provider

**Solution Approach:**

1. **Add LLM Client Support** (in `src/pkg/llm/`):
   ```go
   // Add clients for each provider
   - cohere_client.go
   - huggingface_client.go
   - palm_client.go
   - bedrock_client.go
   - jina_client.go
   ```

2. **Update Supabase Adapter**:
   ```go
   // src/pkg/vectordb/supabase/adapter.go
   type Adapter struct {
       client    *supabase.Client
       db        *sql.DB
       config    *vectordb.Config
       llmClients map[string]llm.Client  // Map of provider → client
   }

   // Initialize based on vectorizer config
   func (a *Adapter) getLLMClient(vectorizer string) (llm.Client, error) {
       switch vectorizer {
       case "text2vec-openai":
           return a.llmClients["openai"], nil
       case "text2vec-cohere":
           return a.llmClients["cohere"], nil
       // ... etc
       }
   }
   ```

3. **Update Document Creation**:
   - Check collection's vectorizer setting
   - Use appropriate LLM client for embeddings
   - Store embedding in `embedding` column

**Implementation Steps:**
1. Start with Cohere (simplest, similar to OpenAI)
2. Add Hugging Face (no API key needed for some models)
3. Add Google PaLM, AWS Bedrock, Jina AI
4. Update collection schema to track vectorizer per collection
5. Update embedding tests for each provider

**Files to Change:**
- `src/pkg/llm/cohere_client.go` (new)
- `src/pkg/llm/huggingface_client.go` (new)
- `src/pkg/llm/palm_client.go` (new)
- `src/pkg/llm/bedrock_client.go` (new)
- `src/pkg/llm/jina_client.go` (new)
- `src/pkg/vectordb/supabase/adapter.go` - Add multi-client support
- `src/pkg/vectordb/supabase/documents.go` - Use correct client per vectorizer
- `src/pkg/vectordb/supabase/collections.go` - Store vectorizer in schema
- `tests/supabase_integration_test.go` - Add tests for each provider

**Testing:**
- Create collections with different vectorizers
- Verify embeddings are generated correctly
- Test semantic search with each embedding type

---

### 3. Optimize BM25 Performance with Indexes (HARDEST) ⭐⭐⭐

**Priority**: Low (works well now, this is optimization only)
**Effort**: 6-8 hours
**Impact**: Performance (10-100x faster on large datasets)

**Problem:**
- BM25 search computes `to_tsvector()` on every query
- No GIN indexes for fast full-text search
- Slow for large collections (>10k documents)

**Why Hardest:**
- Requires schema migration for existing collections
- Need to handle backward compatibility
- Need to detect and use indexes when available
- More complex SQL with generated columns

**Solution:**
See detailed plan in `docs/SUPABASE_BM25_IMPROVEMENT.md`

**Key Steps:**
1. Add `content_tsv` tsvector column (generated)
2. Add GIN index on `content_tsv`
3. Add `doc_length` column for BM25 normalization
4. Update queries to use pre-computed columns
5. Create migration script for existing collections
6. Add fallback for collections without optimization

**Implementation:**
```sql
-- Add to collection creation
ALTER TABLE collection_table
ADD COLUMN content_tsv tsvector
GENERATED ALWAYS AS (
  to_tsvector('english', COALESCE(content, '') || ' ' || COALESCE(text, ''))
) STORED;

CREATE INDEX idx_content_tsv ON collection_table USING GIN(content_tsv);
```

```go
// Update searchByFullText to use optimized column when available
func (a *Adapter) searchByFullText(...) {
    // Check if optimized columns exist
    hasOptimization, _ := a.hasTextSearchOptimization(tableName)

    if hasOptimization {
        // Use pre-computed tsvector column (fast)
        sqlQuery = `SELECT ..., ts_rank_cd(content_tsv, ...) FROM ... WHERE content_tsv @@ ...`
    } else {
        // Fall back to current implementation (slower but works)
        sqlQuery = `SELECT ..., ts_rank_cd(to_tsvector(...), ...) FROM ... WHERE to_tsvector(...) @@ ...`
    }
}
```

**Files to Change:**
- `src/pkg/vectordb/supabase/collections.go` - Add optimization columns on creation
- `src/pkg/vectordb/supabase/queries.go` - Use optimized queries when available
- `src/pkg/vectordb/supabase/migrations.go` (new) - Migration utilities
- `src/cmd/supabase/optimize.go` (new) - CLI command to optimize existing collections
- `tests/supabase_integration_test.go` - Test both optimized and non-optimized paths

**Testing:**
- Create collection with optimization
- Create collection without optimization
- Verify both work correctly
- Benchmark performance difference

---

## Summary

| Task | Priority | Effort | Impact | Complexity |
|------|----------|--------|--------|------------|
| 1. Fix Collection Names | High | 1-2h | UX | Easy ⭐ |
| 2. Add Embedding Providers | Medium | 4-6h | Features | Medium ⭐⭐ |
| 3. Optimize BM25 Performance | Low | 6-8h | Performance | Hard ⭐⭐⭐ |

**Recommended Order:**
1. **Start with #1** (Fix Collection Names) - Quick win, improves UX
2. **Then do #2** (Add Embedding Providers) - Achieves feature parity
3. **Save #3** (BM25 Optimization) for later - Current implementation works well, this is just for scale

**Total Estimated Effort:** 11-16 hours to complete all three

---

## Notes

- All three are independent and can be done in any order
- #1 is the quickest win and most user-visible
- #2 is the most impactful for feature parity
- #3 is optional optimization (current BM25 works fine for most use cases)

After completing #1 and #2, Supabase support will be essentially feature-complete! 🎉
