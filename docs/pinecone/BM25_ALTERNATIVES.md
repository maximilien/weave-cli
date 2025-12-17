# BM25 Alternatives for Pinecone

## Overview

**Pinecone does not support native BM25 full-text search.** It is purely a vector database optimized for dense vector similarity search (semantic search).

This document outlines alternative approaches for achieving BM25-like functionality or hybrid search when using Pinecone.

---

## Why No BM25?

Pinecone is designed as a **vector-only** database:
- Focuses on dense vector embeddings and similarity search
- Serverless architecture optimized for vector operations
- No built-in text indexing or tokenization capabilities

For applications requiring BM25, consider using a database with native support:
- **Elasticsearch** - Native kNN + BM25 hybrid
- **OpenSearch** - Native kNN + BM25 hybrid
- **Weaviate** - Native BM25 + vector hybrid
- **MongoDB Atlas** - Native text search + vector

---

## Alternative Approaches

### Option 1: Sparse-Dense Hybrid Embeddings (Recommended)

Pinecone supports **sparse-dense vectors** for hybrid search that approximates BM25 + semantic behavior.

**Concept**:
- Dense vectors (1536 dims): Semantic meaning (e.g., OpenAI embeddings)
- Sparse vectors: Term frequencies (BM25-like keyword matching)
- Pinecone combines both at query time

**Implementation**:

```python
# Example using Pinecone Python SDK
from pinecone_text.sparse import BM25Encoder

# Create BM25 encoder
bm25_encoder = BM25Encoder()
bm25_encoder.fit(documents)  # Train on your corpus

# Create hybrid vectors
dense_vector = get_openai_embedding(text)  # [1536 dims]
sparse_vector = bm25_encoder.encode_queries(text)  # Sparse {index: value}

# Upsert to Pinecone
index.upsert(vectors=[(
    "doc-1",
    dense_vector,
    sparse_vector,  # Pinecone's hybrid field
    metadata
)])

# Query with hybrid search
results = index.query(
    vector=query_dense,
    sparse_vector=query_sparse,
    top_k=10,
    include_metadata=True
)
```

**Limitations**:
- Requires `pinecone-text` library (not available in weave-cli Go SDK)
- Weave CLI currently does not support sparse vectors
- Would require custom implementation

**Status in weave-cli**: ❌ Not yet implemented

---

### Option 2: Metadata Text Filtering

Use Pinecone's metadata filtering to approximate keyword search.

**How It Works**:
1. Store important keywords in metadata
2. Use metadata filters to narrow search results
3. Combine with semantic search for ranking

**Example**:

```bash
# Store documents with keyword metadata
weave document create MyIndex doc.txt

# Metadata automatically includes text snippets
# Then filter by keywords:
weave query metadata MyIndex --filter "keywords=machine learning"
```

**Pros**:
- Available in weave-cli today
- Simple to implement
- Good for category/tag filtering

**Cons**:
- Not true BM25 (no term frequency scoring)
- Requires manual keyword extraction
- Limited to exact or prefix matches

---

### Option 3: External BM25 Service + Vector Reranking

Run BM25 search externally, then use Pinecone for semantic reranking.

**Architecture**:
1. **BM25 Retrieval**: Use Elasticsearch/OpenSearch/Solr for initial keyword search
2. **Vector Reranking**: Query Pinecone with top results for semantic similarity
3. **Hybrid Scoring**: Combine BM25 scores with vector scores

**Example Workflow**:

```bash
# Step 1: BM25 search in Elasticsearch
curl -X POST "localhost:9200/docs/_search" -d '{
  "query": {"match": {"text": "machine learning"}}
}' | jq '.hits.hits[]._id' > ids.txt

# Step 2: Get embeddings for query
export QUERY_EMBEDDING=$(weave embed "machine learning")

# Step 3: Rerank with Pinecone (semantic)
weave query semantic MyIndex "machine learning" --top-k 100

# Step 4: Merge results (custom script)
```

**Pros**:
- True BM25 + semantic hybrid
- Leverages strengths of both systems
- Production-grade approach

**Cons**:
- Requires running multiple services
- Increased complexity and latency
- Data synchronization needed

---

### Option 4: Switch to a Hybrid-Native Database

If BM25 is critical for your use case, consider using a database with native hybrid search:

| Database | BM25 Support | Vector Support | Hybrid Search |
|----------|--------------|----------------|---------------|
| **Elasticsearch** | ✅ Native | ✅ HNSW | ✅ Native RRF |
| **OpenSearch** | ✅ Native | ✅ HNSW | ✅ Native RRF |
| **Weaviate** | ✅ Native | ✅ HNSW | ✅ Native |
| **MongoDB Atlas** | ✅ Text Search | ✅ Vector | ✅ $search |

**Migration Path**:
```bash
# Export from Pinecone
weave collection export MyIndex --pinecone > data.json

# Import to Elasticsearch
export VECTOR_DB_TYPE=elasticsearch-cloud
weave collection import MyIndex data.json
```

---

## Use Case Recommendations

### When Pinecone Alone Is Sufficient

✅ **Use Pinecone if**:
- Semantic search is primary use case
- Keyword/exact match is not critical
- Simplified serverless architecture preferred
- Fast semantic retrieval is key requirement

### When You Need BM25 Alternatives

⚠️ **Consider alternatives if**:
- Users search with exact phrases or keywords
- Domain-specific terminology needs exact matching
- Hybrid search improves relevance significantly
- Compliance requires explainable keyword search

### When to Switch Databases

❌ **Switch from Pinecone if**:
- BM25 is a core requirement (not optional)
- You're already running Elasticsearch/OpenSearch
- Local deployment is required (Pinecone is cloud-only)
- Cost of running dual systems is prohibitive

---

## Future Enhancements

**Planned for weave-cli**:
- [ ] Sparse vector support for Pinecone hybrid search
- [ ] Built-in BM25 encoder integration
- [ ] Hybrid scoring utilities (BM25 + semantic)
- [ ] Multi-DB query federation (Pinecone + Elasticsearch)

**Tracking Issue**: [#TBD]

---

## Comparison: Pinecone vs Hybrid Databases

| Feature | Pinecone | Elasticsearch | Weaviate |
|---------|----------|---------------|----------|
| **BM25 Native** | ❌ | ✅ | ✅ |
| **Vector Search** | ✅ (Excellent) | ✅ (Good) | ✅ (Excellent) |
| **Hybrid Search** | ⚠️ (Sparse-dense) | ✅ (RRF) | ✅ (Native) |
| **Serverless** | ✅ | ❌ (Elastic Cloud) | ⚠️ (Cloud option) |
| **Self-Hosted** | ❌ | ✅ | ✅ |
| **Complexity** | Low | Medium | Medium |

---

## Additional Resources

- **Pinecone Hybrid Search**: [https://docs.pinecone.io/docs/hybrid-search](https://docs.pinecone.io/docs/hybrid-search)
- **pinecone-text Library**: [https://github.com/pinecone-io/pinecone-text](https://github.com/pinecone-io/pinecone-text)
- **Elasticsearch Setup**: [../elasticsearch/SETUP.md](../elasticsearch/SETUP.md)
- **Weaviate Setup**: [../weaviate/SETUP.md](../weaviate/SETUP.md)

---

## Examples

### Example 1: Metadata Keyword Filtering (Available Today)

```bash
# Create collection with metadata
weave collection create TechDocs --text

# Add documents (metadata auto-extracted)
weave document create TechDocs ml-guide.txt ai-tutorial.txt

# Search with metadata filters (keyword approximation)
weave query semantic TechDocs "neural networks" --metadata "category=tutorial"

# This combines:
# - Semantic search for "neural networks"
# - Metadata filter for category=tutorial
```

### Example 2: Dual-Database Architecture (Advanced)

```yaml
# docker-compose.yml - Hybrid Architecture
version: '3.8'
services:
  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.11.0
    environment:
      - discovery.type=single-node
    ports:
      - "9200:9200"

  # Application uses:
  # - Elasticsearch for BM25 search
  # - Pinecone (cloud) for vector search
  # - Custom merger for hybrid results
```

```bash
# Search workflow
# 1. BM25 search
curl -X POST "localhost:9200/docs/_search" -d '{...}' > bm25_results.json

# 2. Vector search (Pinecone)
weave query semantic MyIndex "query text" > vector_results.json

# 3. Merge and rerank
python merge_results.py bm25_results.json vector_results.json
```

---

## Summary

**Key Takeaways**:
1. Pinecone is **vector-only** - no native BM25 support
2. **Sparse-dense hybrid** is the closest alternative (requires pinecone-text)
3. **Metadata filtering** is available today in weave-cli but limited
4. For critical BM25 needs, consider **Elasticsearch, OpenSearch, or Weaviate**
5. **Dual-database architecture** is production-viable for complex hybrid search

**Recommendation**: If BM25 is important, evaluate whether Pinecone's sparse-dense approach or a hybrid database better fits your needs before committing to architecture.
