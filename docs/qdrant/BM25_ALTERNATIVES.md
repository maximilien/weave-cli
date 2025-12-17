# BM25 Alternatives for Qdrant

## Overview

**Qdrant does not support native BM25 full-text search.** It is a high-performance vector database optimized for similarity search using HNSW indexing, written in Rust.

This document outlines alternative approaches for achieving BM25-like functionality when using Qdrant.

---

## Why No BM25?

Qdrant's design priorities:
- **Vector-optimized**: HNSW indexing for fast similarity search
- **Payload filtering**: Rich metadata filtering (but not full-text ranked search)
- **Performance**: Rust implementation for speed and memory efficiency
- **Scalability**: Distributed architecture for horizontal scaling

Qdrant focuses on **dense vector search** with powerful payload (metadata) filtering, but does not include text indexing or BM25 scoring.

---

## Alternative Approaches

### Option 1: Payload Full-Text Filtering

Qdrant supports **payload text filtering** - not BM25, but useful for keyword matching.

**How It Works**:
```bash
# Create collection with rich payload
weave collection create Articles --text

# Add documents (payload includes text fields)
weave document create Articles article1.txt article2.txt

# Filter by payload text (keyword matching)
weave query semantic Articles "neural networks" \
  --metadata "text MATCH 'machine learning'"
```

**Capabilities**:
- Exact text matching
- Prefix matching
- Combine with vector search for hybrid-like behavior

**Limitations**:
- No BM25 term frequency scoring
- No ranking by text relevance
- Exact/prefix only (no fuzzy matching)

**Qdrant Query Example** (using REST API):
```json
{
  "vector": [0.1, 0.2, ...],
  "filter": {
    "must": [
      {
        "key": "text",
        "match": {
          "text": "machine learning"
        }
      }
    ]
  },
  "limit": 10
}
```

---

### Option 2: External Search Engine + Qdrant Reranking

Combine Qdrant with an external full-text search engine for true hybrid search.

**Architecture**:
1. **Elasticsearch/OpenSearch**: BM25 search for keyword matching
2. **Qdrant**: Semantic reranking of top results

**Workflow**:
```bash
# Step 1: BM25 search in Elasticsearch
curl -X POST "localhost:9200/docs/_search" -d '{
  "query": {"match": {"text": "machine learning"}},
  "size": 100
}' | jq '.hits.hits[]._id' > candidate_ids.txt

# Step 2: Rerank with Qdrant (semantic)
weave query semantic Articles "machine learning" \
  --filter "id IN (id1, id2, id3, ...)"
```

**Pros**:
- True BM25 + vector hybrid search
- Production-grade approach used by many systems
- Leverages strengths of both databases

**Cons**:
- Requires running multiple services (complexity)
- Data synchronization overhead
- Increased latency (two database calls)

---

### Option 3: Qdrant Sparse Vectors (Experimental)

Qdrant has **experimental support for sparse vectors**, which can approximate BM25.

**Concept**:
- Dense vectors: Semantic embeddings (e.g., OpenAI 1536 dims)
- Sparse vectors: Term frequencies (BM25-like)
- Qdrant searches both simultaneously

**Status**: ⚠️ Experimental feature, not yet available in weave-cli

**Future Implementation**:
```python
# Example (Python SDK, not available in weave-cli yet)
from qdrant_client import QdrantClient
from qdrant_client.models import SparseVector

client = QdrantClient(url="http://localhost:6334")

# Upsert with dense + sparse vectors
client.upsert(
    collection_name="articles",
    points=[{
        "id": "doc-1",
        "vector": {
            "dense": [0.1, 0.2, ...],  # OpenAI embedding
            "sparse": SparseVector(
                indices=[45, 132, 689],  # Token IDs
                values=[0.5, 0.3, 0.7]   # BM25-like scores
            )
        },
        "payload": {"text": "..."}
    }]
)
```

**Weave CLI Support**: 🚧 Planned for future release

---

### Option 4: Switch to Hybrid Database

If BM25 is critical, consider databases with native hybrid search:

| Database | BM25 | Vector | Hybrid | Qdrant Alternative |
|----------|------|--------|--------|-------------------|
| **Weaviate** | ✅ | ✅ | ✅ | Similar simplicity, adds BM25 |
| **Elasticsearch** | ✅ | ✅ | ✅ | Mature ecosystem, more complex |
| **OpenSearch** | ✅ | ✅ | ✅ | Open-source Elasticsearch fork |
| **Milvus** | ✅ | ✅ | ✅ | Similar performance focus |

**Migration Path**:
```bash
# Export from Qdrant
weave collection export Articles --qdrant-local > data.json

# Import to Weaviate
export VECTOR_DB_TYPE=weaviate
weave collection import Articles data.json

# Use hybrid search
weave query hybrid Articles "machine learning" --alpha 0.5
```

---

## Use Case Recommendations

### When Qdrant Alone Is Sufficient

✅ **Use Qdrant if**:
- Semantic search is primary use case
- Payload filtering meets keyword needs
- Performance and scalability are critical
- Rust-based performance is valued
- Cross-platform support needed (unlike Chroma)

### When You Need BM25 Alternatives

⚠️ **Consider alternatives if**:
- Users frequently search with exact keywords
- BM25 ranking significantly improves relevance
- You're already running Elasticsearch/OpenSearch
- Hybrid search is a core product requirement

### When to Switch Databases

❌ **Switch from Qdrant if**:
- BM25 is non-negotiable (e.g., compliance, UX requirements)
- You need simpler hybrid search (Weaviate's native BM25)
- Operational complexity of dual systems is too high
- Budget/resources don't allow multiple services

---

## Comparison: Qdrant vs Hybrid Databases

| Feature | Qdrant | Weaviate | Elasticsearch |
|---------|--------|----------|---------------|
| **BM25 Native** | ❌ | ✅ | ✅ |
| **Vector Search** | ✅ (Excellent) | ✅ (Excellent) | ✅ (Good) |
| **Hybrid Search** | ⚠️ (Sparse vectors) | ✅ | ✅ |
| **Performance** | Excellent | Very Good | Good |
| **Setup** | Easy (Docker) | Easy (Docker) | Medium |
| **Scalability** | ✅ Distributed | ✅ Clustering | ✅ Clustering |

---

## Practical Examples

### Example 1: Payload Text Filtering (Available Today)

```bash
# Create collection
weave collection create TechDocs --text

# Add documents
weave document create TechDocs ml-guide.txt ai-tutorial.txt

# Semantic search + payload filter
weave query semantic TechDocs "deep learning" \
  --metadata "category=tutorial,tags CONTAINS python"
```

### Example 2: Qdrant + PostgreSQL FTS (Hybrid)

```sql
-- PostgreSQL full-text search setup
CREATE TABLE documents (
    id SERIAL PRIMARY KEY,
    text TEXT,
    tsv tsvector GENERATED ALWAYS AS (to_tsvector('english', text)) STORED
);

CREATE INDEX tsv_idx ON documents USING GIN(tsv);

-- Query: BM25-like search
SELECT id, ts_rank(tsv, query) AS score
FROM documents, to_tsquery('machine & learning') query
WHERE tsv @@ query
ORDER BY score DESC
LIMIT 100;
```

```bash
# Then rerank with Qdrant
weave query semantic TechDocs "machine learning" \
  --filter "id IN (1,5,23,45,...)"
```

### Example 3: Dual-Database Architecture (Production)

```yaml
# docker-compose.yml
version: '3.8'
services:
  qdrant:
    image: qdrant/qdrant:latest
    ports:
      - "6333:6333"
      - "6334:6334"
    volumes:
      - ./qdrant_storage:/qdrant/storage

  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.11.0
    ports:
      - "9200:9200"
    environment:
      - discovery.type=single-node
```

```python
# Application: Merge results from both
def hybrid_search(query_text, top_k=10):
    # 1. BM25 search (Elasticsearch)
    es_results = elasticsearch.search(
        index="docs",
        body={"query": {"match": {"text": query_text}}},
        size=100
    )

    # 2. Vector search (Qdrant)
    qdrant_results = qdrant_client.search(
        collection_name="docs",
        query_vector=get_embedding(query_text),
        limit=100
    )

    # 3. Reciprocal Rank Fusion (RRF)
    return merge_results(es_results, qdrant_results, k=60)
```

---

## Future Enhancements

**Planned for weave-cli**:
- [ ] Sparse vector support (when Qdrant stabilizes feature)
- [ ] Built-in BM25 encoder integration
- [ ] Hybrid scoring utilities (RRF merger)
- [ ] Multi-DB federation (Qdrant + Elasticsearch queries)

**Tracking Issue**: [#TBD]

---

## Additional Resources

- **Qdrant Documentation**: [https://qdrant.tech/documentation/](https://qdrant.tech/documentation/)
- **Qdrant Payload Filtering**: [https://qdrant.tech/documentation/concepts/filtering/](https://qdrant.tech/documentation/concepts/filtering/)
- **Qdrant Setup**: [SETUP.md](SETUP.md)
- **Weaviate Setup (Hybrid Alternative)**: [../weaviate/SETUP.md](../weaviate/SETUP.md)
- **Sparse Vectors (Experimental)**: [https://qdrant.tech/articles/sparse-vectors/](https://qdrant.tech/articles/sparse-vectors/)

---

## Summary

**Key Takeaways**:
1. Qdrant is **vector-optimized** - no native BM25 support
2. **Payload text filtering** provides keyword matching (not ranked)
3. **Sparse vectors** are experimental - not yet production-ready
4. **External FTS** (Elasticsearch, PostgreSQL) enables true hybrid search
5. For critical BM25 needs, consider **Weaviate** (simplest) or **Elasticsearch** (most mature)

**Recommendation**: Qdrant excels at vector search. If BM25 is important, evaluate whether payload filtering suffices, or if the operational complexity of a hybrid architecture is justified versus switching to a natively hybrid database like Weaviate.
