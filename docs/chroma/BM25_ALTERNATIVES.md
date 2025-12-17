# BM25 Alternatives for Chroma

## Overview

**Chroma does not support native BM25 full-text search.** It is designed as a pure vector database focused on embeddings and similarity search.

This document outlines alternative approaches for achieving BM25-like functionality when using Chroma.

---

## Why No BM25?

Chroma's design philosophy:
- **Embedding-first**: Focus on semantic search via vector embeddings
- **Simplicity**: Lightweight API without complex text indexing
- **Portability**: Pure Python implementation (no heavy NLP dependencies)
- **Developer-friendly**: Easy to embed in applications without infrastructure

For applications requiring BM25, consider these alternatives within Chroma's ecosystem or migrate to hybrid databases.

---

## Alternative Approaches

### Option 1: Metadata Text Filtering

Use Chroma's metadata filtering to approximate keyword search.

**How It Works**:
```bash
# Create collection with rich metadata
weave collection create Docs --text

# Add documents (Chroma stores text in metadata)
weave document create Docs document.txt

# Filter by metadata keywords
weave query metadata Docs --filter "keywords CONTAINS machine learning"
```

**Pros**:
- Available today in weave-cli
- Simple implementation
- Good for categorical filtering

**Cons**:
- No term frequency scoring (not true BM25)
- Exact/prefix matching only
- Limited query expressiveness

---

### Option 2: Pre-filter with External Full-Text Search

Combine Chroma with an external full-text search engine.

**Architecture**:
1. **PostgreSQL/SQLite Full-Text Search**: Initial keyword filtering
2. **Chroma**: Semantic reranking of top results

**Example Workflow**:
```sql
-- Step 1: BM25-like search in PostgreSQL
SELECT id, ts_rank(tsv, query) AS rank
FROM documents, to_tsquery('machine & learning') query
WHERE tsv @@ query
ORDER BY rank DESC
LIMIT 100;

-- Step 2: Get IDs and rerank with Chroma
weave query semantic Docs "machine learning" --filter "id IN (1,5,12,...)"
```

**Pros**:
- True full-text search capabilities
- Leverages PostgreSQL's mature text search
- Can use existing database infrastructure

**Cons**:
- Requires running PostgreSQL/SQLite alongside Chroma
- Data synchronization complexity
- Additional operational overhead

---

### Option 3: Client-Side BM25 Scoring

Implement BM25 scoring in your application before querying Chroma.

**Approach**:
```python
from rank_bm25 import BM25Okapi
import chromadb

# Load documents and create BM25 index
corpus = [doc['text'] for doc in documents]
tokenized_corpus = [doc.split() for doc in corpus]
bm25 = BM25Okapi(tokenized_corpus)

# Query: Get top-K from BM25
query = "machine learning"
bm25_scores = bm25.get_scores(query.split())
top_k_ids = np.argsort(bm25_scores)[-100:][::-1]

# Rerank with Chroma (semantic)
chroma_client = chromadb.Client()
collection = chroma_client.get_collection("docs")
results = collection.query(
    query_texts=[query],
    where={"id": {"$in": top_k_ids.tolist()}},
    n_results=10
)
```

**Pros**:
- Full control over BM25 scoring
- No additional services needed
- Can tune BM25 parameters

**Cons**:
- Requires loading full corpus into memory
- Client-side processing overhead
- Not practical for large datasets (>100K docs)

---

### Option 4: Switch to Hybrid Database

If BM25 is critical, consider databases with native hybrid search:

| Database | BM25 | Vector | Hybrid | Setup Complexity |
|----------|------|--------|--------|------------------|
| **Weaviate** | ✅ | ✅ | ✅ | Low (Docker) |
| **Elasticsearch** | ✅ | ✅ | ✅ | Medium |
| **Supabase (pgvector)** | ✅ | ✅ | ⚠️ | Low (Cloud) |
| **Milvus** | ✅ | ✅ | ✅ | Medium |

**Migration Path**:
```bash
# Export from Chroma
weave collection export MyDocs --chroma-local > data.json

# Import to Weaviate
export VECTOR_DB_TYPE=weaviate
weave collection import MyDocs data.json

# Now use hybrid search
weave query hybrid MyDocs "machine learning" --alpha 0.5
```

---

## Use Case Recommendations

### When Chroma Alone Is Sufficient

✅ **Use Chroma if**:
- Semantic search is primary requirement
- Small to medium datasets (< 1M vectors)
- Simplicity and ease of deployment are priorities
- Running locally or embedded in applications
- macOS development (Chroma's primary platform in weave-cli)

### When You Need BM25 Alternatives

⚠️ **Consider alternatives if**:
- Users search with specific keywords or phrases
- Domain jargon requires exact matching
- Combining keyword + semantic improves relevance
- Dataset is large (> 1M docs) needing efficient text search

### When to Switch Databases

❌ **Switch from Chroma if**:
- BM25 is a core requirement (not nice-to-have)
- You need cross-platform support (Chroma is macOS-only in weave-cli)
- Production scale requires enterprise features
- Multi-modal search (text, image, audio) is needed

---

## Platform Considerations

⚠️ **Important**: Chroma in weave-cli is currently **macOS-only** due to CGO dependencies.

**For cross-platform BM25 needs**, consider:
- **Weaviate**: Pure Go, all platforms
- **Qdrant**: Rust-based, all platforms
- **Supabase (pgvector)**: Cloud-based, platform-independent

---

## Comparison: Chroma vs Hybrid Databases

| Feature | Chroma | Weaviate | Elasticsearch |
|---------|--------|----------|---------------|
| **BM25 Native** | ❌ | ✅ | ✅ |
| **Vector Search** | ✅ | ✅ | ✅ |
| **Hybrid Search** | ❌ | ✅ | ✅ |
| **Setup Complexity** | Low | Low | Medium |
| **Cross-Platform** | ⚠️ (macOS only) | ✅ | ✅ |
| **Embedded Use** | ✅ | ⚠️ | ❌ |

---

## Practical Examples

### Example 1: Metadata Filtering (Available Today)

```bash
# Create collection
weave collection create KnowledgeBase --text

# Add documents with metadata
weave document create KnowledgeBase doc1.txt doc2.txt

# Search with metadata filter (keyword approximation)
weave query semantic KnowledgeBase "neural networks" \
  --metadata "category=tutorial,topic=ai"
```

### Example 2: Chroma + SQLite FTS5 (Hybrid Architecture)

```python
import chromadb
import sqlite3

# Setup
chroma_client = chromadb.PersistentClient(path="./chroma_db")
sqlite_conn = sqlite3.connect("docs.db")

# Create FTS5 table in SQLite
sqlite_conn.execute("""
    CREATE VIRTUAL TABLE docs_fts USING fts5(id, text);
""")

# Insert documents to both systems
for doc in documents:
    # Chroma (semantic)
    collection.add(ids=[doc.id], documents=[doc.text])

    # SQLite (keyword)
    sqlite_conn.execute(
        "INSERT INTO docs_fts VALUES (?, ?)",
        (doc.id, doc.text)
    )

# Query: BM25-like + semantic
# 1. SQLite full-text search
cursor = sqlite_conn.execute("""
    SELECT id FROM docs_fts
    WHERE docs_fts MATCH ?
    ORDER BY rank LIMIT 100
""", ("machine learning",))
candidate_ids = [row[0] for row in cursor]

# 2. Chroma semantic reranking
results = collection.query(
    query_texts=["machine learning"],
    where={"id": {"$in": candidate_ids}},
    n_results=10
)
```

---

## Future Enhancements

**Planned for weave-cli**:
- [ ] Built-in client-side BM25 scoring
- [ ] Hybrid scoring utilities (BM25 + semantic merger)
- [ ] Integration guides for Chroma + PostgreSQL FTS
- [ ] Cross-platform Chroma support (Linux/Windows)

**Tracking Issue**: [#TBD]

---

## Additional Resources

- **Chroma Documentation**: [https://docs.trychroma.com/](https://docs.trychroma.com/)
- **Chroma Setup**: [SETUP.md](SETUP.md)
- **Weaviate Setup (Hybrid Alternative)**: [../weaviate/SETUP.md](../weaviate/SETUP.md)
- **PostgreSQL FTS**: [https://www.postgresql.org/docs/current/textsearch.html](https://www.postgresql.org/docs/current/textsearch.html)

---

## Summary

**Key Takeaways**:
1. Chroma is **embedding-focused** - no native BM25 support
2. **Metadata filtering** is the simplest available option today
3. **External FTS** (PostgreSQL, SQLite) can provide true BM25
4. **Client-side BM25** works for small datasets (< 100K docs)
5. For critical BM25 needs, consider **Weaviate, Elasticsearch, or Supabase**

**Recommendation**: Evaluate whether semantic search alone meets your needs, or if the added complexity of hybrid approaches justifies switching to a natively hybrid database.
