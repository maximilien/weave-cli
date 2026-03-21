# Supabase BM25 Search Improvement Plan

## Current State

The Supabase adapter currently uses PostgreSQL's basic `ts_rank()` function with `to_tsvector()` for full-text search. While functional, this has limitations:

### Current Implementation

```sql
SELECT id, content, text, ...,
       ts_rank(to_tsvector('english', COALESCE(content, '') || ' ' || COALESCE(text, '')),
               plainto_tsquery('english', $1)) as score
FROM collection_table
WHERE to_tsvector('english', COALESCE(content, '') || ' ' || COALESCE(text, '')) @@ plainto_tsquery('english', $1)
ORDER BY score DESC
LIMIT $2
```

### Limitations

1. **Not True BM25**: Uses TF-IDF-based `ts_rank()`, not BM25 algorithm
2. **Performance**: Computes `to_tsvector()` on every query (very slow for large datasets)
3. **No Indexing**: No GIN/GiST indexes on text search vectors
4. **Limited Ranking**: Doesn't account for document length normalization (key BM25 feature)

## PostgreSQL BM25 Options

### Option 1: ParadeDB Extension (Recommended)

[ParadeDB](https://github.com/paradedb/paradedb) provides true BM25 search with Tantivy integration.

**Pros:**
- True BM25 algorithm
- High performance (Rust-based Tantivy)
- Modern, well-maintained
- Works with Supabase (self-hosted)

**Cons:**
- Requires extension installation (may not be available on Supabase Cloud)
- Newer project

**Implementation:**

```sql
-- Install extension
CREATE EXTENSION IF NOT EXISTS paradedb;

-- Create BM25 index
CALL paradedb.create_bm25(
  index_name => 'collection_search_idx',
  table_name => 'collection_table',
  key_field => 'id',
  text_fields => '{
    "content": {"tokenizer": {"type": "en_stem"}},
    "text": {"tokenizer": {"type": "en_stem"}}
  }'
);

-- Search with BM25
SELECT * FROM collection_search_idx.search(
  'search query',
  limit_rows => 10
);
```

### Option 2: pg_search Extension

Provides BM25 ranking function for PostgreSQL.

**Pros:**
- Simpler installation
- Direct BM25 implementation

**Cons:**
- Less actively maintained
- May not be available on Supabase Cloud

### Option 3: Enhanced Native PostgreSQL (No Extension Required)

Improve the native implementation without requiring extensions.

**Pros:**
- Works on any PostgreSQL/Supabase instance
- No extension dependencies
- Good performance with proper indexing

**Cons:**
- Not true BM25 (uses ts_rank with custom normalization)
- Requires manual BM25-like scoring implementation

## Recommended Approach: Multi-Tier Implementation

Implement all three approaches with automatic fallback:

```
1. Try ParadeDB (if available) → Best performance, true BM25
2. Try pg_search (if available) → Good BM25 implementation
3. Fall back to Enhanced Native → Works everywhere
```

## Enhanced Native PostgreSQL Implementation

Since Supabase Cloud may not have ParadeDB/pg_search, we should significantly improve the native implementation:

### Schema Changes

```sql
-- Add tsvector column for pre-computed full-text search
ALTER TABLE collection_table
ADD COLUMN content_tsv tsvector
GENERATED ALWAYS AS (
  to_tsvector('english', COALESCE(content, '') || ' ' || COALESCE(text, ''))
) STORED;

-- Create GIN index for fast full-text search
CREATE INDEX idx_content_tsv ON collection_table USING GIN(content_tsv);

-- Add document length column (for BM25-like normalization)
ALTER TABLE collection_table
ADD COLUMN doc_length INTEGER
GENERATED ALWAYS AS (
  length(COALESCE(content, '') || ' ' || COALESCE(text, ''))
) STORED;
```

### BM25-Like Scoring Function

```sql
-- Custom BM25-like ranking function
CREATE OR REPLACE FUNCTION bm25_rank(
  doc_tsv tsvector,
  query_tsquery tsquery,
  doc_length INTEGER,
  avg_doc_length FLOAT DEFAULT 1000,
  k1 FLOAT DEFAULT 1.5,
  b FLOAT DEFAULT 0.75
) RETURNS FLOAT AS $$
DECLARE
  tf FLOAT;
  idf FLOAT;
  doc_norm FLOAT;
  bm25_score FLOAT;
BEGIN
  -- Get term frequency (simplified - uses ts_rank as proxy)
  tf := ts_rank(doc_tsv, query_tsquery, 0);

  -- Document length normalization (BM25 key feature)
  doc_norm := 1.0 - b + b * (doc_length / avg_doc_length);

  -- BM25-like score
  bm25_score := (tf * (k1 + 1.0)) / (tf + k1 * doc_norm);

  RETURN bm25_score;
END;
$$ LANGUAGE plpgsql IMMUTABLE;
```

### Optimized Query

```sql
-- Pre-calculate average document length for collection
WITH avg_length AS (
  SELECT AVG(doc_length) as avg_len FROM collection_table
)
SELECT
  id, content, text, image, image_data, url, metadata,
  bm25_rank(
    content_tsv,
    plainto_tsquery('english', $1),
    doc_length,
    (SELECT avg_len FROM avg_length)
  ) as score
FROM collection_table, avg_length
WHERE content_tsv @@ plainto_tsquery('english', $1)
ORDER BY score DESC
LIMIT $2;
```

## Performance Comparison

### Before (Current Implementation)

```sql
-- Computes tsvector on every row during query
WHERE to_tsvector('english', ...) @@ plainto_tsquery(...)
```

- **No index utilization**
- **O(n) computation** on every query
- **Slow for large datasets** (>10k documents)

### After (Enhanced Implementation)

```sql
-- Uses pre-computed tsvector with GIN index
WHERE content_tsv @@ plainto_tsquery(...)
```

- **GIN index utilization** (O(log n) lookup)
- **Pre-computed vectors** (no query-time computation)
- **Fast for large datasets** (>1M documents)

## Implementation Checklist

- [ ] Add extension detection (ParadeDB, pg_search)
- [ ] Implement ParadeDB BM25 search (if available)
- [ ] Implement pg_search BM25 (if available)
- [ ] Enhance native implementation:
  - [ ] Add tsvector column to collection creation
  - [ ] Add GIN index creation
  - [ ] Add doc_length tracking
  - [ ] Implement BM25-like scoring function
  - [ ] Update SearchBM25 to use enhanced query
- [ ] Update collection creation to include new columns/indexes
- [ ] Add migration support for existing collections
- [ ] Update integration tests
- [ ] Update VDB_SUPPORT.md to reflect improvements

## Migration Path for Existing Collections

```sql
-- For collections created before this enhancement
DO $$
DECLARE
  collection_table TEXT;
BEGIN
  FOR collection_table IN
    SELECT tablename FROM pg_tables
    WHERE tablename LIKE 'collection_%'
  LOOP
    -- Add tsvector column if not exists
    EXECUTE format('
      ALTER TABLE %I
      ADD COLUMN IF NOT EXISTS content_tsv tsvector
      GENERATED ALWAYS AS (
        to_tsvector(''english'', COALESCE(content, '''') || '' '' || COALESCE(text, ''''))
      ) STORED
    ', collection_table);

    -- Create GIN index if not exists
    EXECUTE format('
      CREATE INDEX IF NOT EXISTS idx_%I_content_tsv
      ON %I USING GIN(content_tsv)
    ', collection_table, collection_table);

    -- Add doc_length if not exists
    EXECUTE format('
      ALTER TABLE %I
      ADD COLUMN IF NOT EXISTS doc_length INTEGER
      GENERATED ALWAYS AS (
        length(COALESCE(content, '''') || '' '' || COALESCE(text, ''''))
      ) STORED
    ', collection_table);
  END LOOP;
END $$;
```

## Expected Improvements

### Performance

- **10-100x faster** for BM25 search on large collections
- **Index-based lookups** instead of full table scans
- **Sub-second queries** even with millions of documents

### Quality

- **Better ranking** with BM25-like scoring
- **Document length normalization** (shorter docs don't unfairly dominate)
- **Improved relevance** for keyword-based searches

### Parity with Weaviate

After implementation:
- ✅ Fast BM25 search
- ✅ Proper text ranking
- ✅ Index-optimized queries
- ✅ Document length normalization
- ⚠️ Still not as sophisticated as Weaviate's native BM25 (but very close)

## References

- [PostgreSQL Full Text Search](https://www.postgresql.org/docs/current/textsearch.html)
- [ParadeDB](https://github.com/paradedb/paradedb)
- [BM25 Algorithm](https://en.wikipedia.org/wiki/Okapi_BM25)
- [Supabase Vector](https://supabase.com/docs/guides/ai/vector-indexes)
