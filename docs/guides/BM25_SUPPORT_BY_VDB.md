# BM25 Support by Vector Database

A consolidated guide covering BM25 full-text search support across all vector databases supported by weave-cli.

---

## Summary Table

| Vector Database | BM25 Support | Notes |
|----------------|-------------|-------|
| **Weaviate** | Native | Built-in BM25 + hybrid search |
| **Elasticsearch** | Native | Mature BM25 + kNN hybrid via RRF |
| **OpenSearch** | Native | BM25 + kNN hybrid (Elasticsearch fork) |
| **Supabase** | Implemented via tsquery | PostgreSQL `ts_rank()` with BM25-like scoring; see [improvement plan](../vdbs/supabase/BM25_IMPROVEMENT.md) |
| **Milvus** | Planned for 2.4+ | Sparse vector support in progress |
| **Chroma** | No native support | Embedding-first design; see [alternatives below](#chroma) |
| **Qdrant** | No native support | Vector-optimized (HNSW); see [alternatives below](#qdrant) |
| **Neo4j** | No native support | Graph-first with Lucene-based FTS; see [alternatives below](#neo4j) |
| **Pinecone** | No native support | Vector-only (cloud); see [alternatives below](#pinecone) |
| **MongoDB** | No native support | Text search exists but not BM25-scored |

---

## VDBs Without Native BM25

The sections below cover alternative approaches for each database that lacks native BM25 support.

---

<a id="chroma"></a>
## Chroma

**Setup guide**: [../vdbs/chroma/SETUP.md](../vdbs/chroma/SETUP.md)

Chroma does not support native BM25 full-text search. It is designed as a pure vector database focused on embeddings and similarity search.

**Why no BM25?**
- Embedding-first: Focus on semantic search via vector embeddings
- Simplicity: Lightweight API without complex text indexing
- Portability: Pure Python implementation (no heavy NLP dependencies)
- Developer-friendly: Easy to embed in applications without infrastructure

### Alternative Approaches

**Option 1: Metadata Text Filtering**

Use Chroma's metadata filtering to approximate keyword search.

```bash
# Create collection with rich metadata
weave collection create Docs --text

# Add documents (Chroma stores text in metadata)
weave document create Docs document.txt

# Filter by metadata keywords
weave query metadata Docs --filter "keywords CONTAINS machine learning"
```

Pros: Available today, simple. Cons: No term frequency scoring, exact matching only.

**Option 2: Pre-filter with External Full-Text Search**

Combine Chroma with PostgreSQL/SQLite for initial BM25 filtering, then rerank with Chroma semantic search.

```sql
-- Step 1: BM25-like search in PostgreSQL
SELECT id, ts_rank(tsv, query) AS rank
FROM documents, to_tsquery('machine & learning') query
WHERE tsv @@ query
ORDER BY rank DESC
LIMIT 100;
```

```bash
# Step 2: Rerank with Chroma
weave query semantic Docs "machine learning" --filter "id IN (1,5,12,...)"
```

**Option 3: Client-Side BM25 Scoring**

```python
from rank_bm25 import BM25Okapi

corpus = [doc['text'] for doc in documents]
tokenized_corpus = [doc.split() for doc in corpus]
bm25 = BM25Okapi(tokenized_corpus)

bm25_scores = bm25.get_scores("machine learning".split())
top_k_ids = np.argsort(bm25_scores)[-100:][::-1]

# Rerank with Chroma semantic search
results = collection.query(
    query_texts=["machine learning"],
    where={"id": {"$in": top_k_ids.tolist()}},
    n_results=10
)
```

Works for small datasets (<100K docs). Not practical for large corpora.

**Option 4: Switch to Hybrid Database**

| Database | BM25 | Vector | Hybrid | Setup |
|----------|------|--------|--------|-------|
| **Weaviate** | Yes | Yes | Yes | Low |
| **Elasticsearch** | Yes | Yes | Yes | Medium |
| **Supabase** | Yes | Yes | Partial | Low |

```bash
# Migration path
weave collection export MyDocs --chroma-local > data.json
export VECTOR_DB_TYPE=weaviate
weave collection import MyDocs data.json
weave query hybrid MyDocs "machine learning" --alpha 0.5
```

**Platform note**: Chroma in weave-cli is currently macOS-only due to CGO dependencies.

---

<a id="qdrant"></a>
## Qdrant

**Setup guide**: [../vdbs/qdrant/SETUP.md](../vdbs/qdrant/SETUP.md)

Qdrant does not support native BM25 full-text search. It is a high-performance vector database optimized for similarity search using HNSW indexing, written in Rust.

**Why no BM25?**
- Vector-optimized: HNSW indexing for fast similarity search
- Payload filtering: Rich metadata filtering (but not full-text ranked search)
- Performance: Rust implementation for speed and memory efficiency
- Scalability: Distributed architecture for horizontal scaling

### Alternative Approaches

**Option 1: Payload Full-Text Filtering**

```bash
weave query semantic Articles "neural networks" \
  --metadata "text MATCH 'machine learning'"
```

Qdrant REST API example:
```json
{
  "vector": [0.1, 0.2, "..."],
  "filter": {
    "must": [{"key": "text", "match": {"text": "machine learning"}}]
  },
  "limit": 10
}
```

Supports exact and prefix matching, but no BM25 term frequency scoring.

**Option 2: External Search Engine + Qdrant Reranking**

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

**Option 3: Qdrant Sparse Vectors (Experimental)**

Qdrant has experimental sparse vector support that can approximate BM25.

```python
from qdrant_client.models import SparseVector

client.upsert(
    collection_name="articles",
    points=[{
        "id": "doc-1",
        "vector": {
            "dense": [0.1, 0.2, "..."],
            "sparse": SparseVector(
                indices=[45, 132, 689],
                values=[0.5, 0.3, 0.7]
            )
        }
    }]
)
```

Status: Experimental, not yet available in weave-cli.

**Option 4: Dual-Database Architecture**

```python
def hybrid_search(query_text, top_k=10):
    es_results = elasticsearch.search(
        index="docs",
        body={"query": {"match": {"text": query_text}}},
        size=100
    )
    qdrant_results = qdrant_client.search(
        collection_name="docs",
        query_vector=get_embedding(query_text),
        limit=100
    )
    return merge_results(es_results, qdrant_results, k=60)  # RRF
```

---

<a id="neo4j"></a>
## Neo4j

**Setup guide**: [../vdbs/neo4j/README.md](../vdbs/neo4j/README.md)

Neo4j does not support native BM25 full-text search in its vector search implementation. It is primarily a graph database with added vector search capabilities.

**Why no BM25?**
- Graph-first: Optimized for relationships and graph traversal
- Vector Search Add-on: Vectors are a newer feature, not core functionality
- Text Search Alternative: Cypher has text matching (Lucene-based, not BM25 scoring)

### Alternative Approaches

**Option 1: Cypher Full-Text Search (Native)**

Neo4j provides Lucene-based full-text search via Cypher:

```cypher
-- Create full-text index (one-time setup)
CREATE FULLTEXT INDEX documentText FOR (n:Document) ON EACH [n.text];

-- Search with text matching
CALL db.index.fulltext.queryNodes("documentText", "machine learning")
YIELD node, score
RETURN node.id, node.text, score
ORDER BY score DESC
LIMIT 10;
```

Supports wildcards, fuzzy search, and term matching. Scoring is Lucene-based, not pure BM25.

**Option 2: Graph + Vector Hybrid Search**

Neo4j's unique strength: combining graph relationships with vector search.

```cypher
MATCH (doc:Document {category: 'tutorial'})-[:RELATES_TO]->(topic:Topic)
WHERE topic.name = 'Machine Learning'
CALL db.index.vector.queryNodes('documentVectors', 5, $queryVector)
YIELD node, score
WHERE node IN [doc]
RETURN node.text, score
ORDER BY score DESC;
```

**Option 3: Neo4j + Elasticsearch**

```python
def hybrid_graph_search(query_text, user_id, top_k=10):
    # 1. BM25 search (Elasticsearch)
    es_results = elasticsearch.search(
        index="documents",
        body={"query": {"match": {"text": query_text}}},
        size=100
    )
    candidate_ids = [hit['_id'] for hit in es_results['hits']['hits']]

    # 2. Graph + vector search (Neo4j)
    cypher = """
        MATCH (u:User {id: $userId})-[:INTERESTED_IN]->(topic:Topic)
              <-[:ABOUT]-(doc:Document)
        WHERE doc.id IN $candidateIds
        CALL db.index.vector.queryNodes('docVectors', $topK, $queryVector)
        YIELD node, score
        WHERE node = doc
        RETURN doc, score, COUNT(topic) AS topicRelevance
        ORDER BY score DESC, topicRelevance DESC
    """
    return session.run(cypher, userId=user_id, candidateIds=candidate_ids,
                       queryVector=get_embedding(query_text), topK=top_k)
```

**When to use Neo4j**: When graph relationships are core to your data model (knowledge graphs, recommendations, network analysis). If you don't need graph features, a simpler hybrid database like Weaviate may be better.

---

<a id="pinecone"></a>
## Pinecone

**Setup guide**: [../vdbs/pinecone/SETUP.md](../vdbs/pinecone/SETUP.md)

Pinecone does not support native BM25 full-text search. It is purely a vector database optimized for dense vector similarity search.

**Why no BM25?**
- Vector-only: Focuses on dense vector embeddings and similarity search
- Serverless architecture optimized for vector operations
- No built-in text indexing or tokenization capabilities

### Alternative Approaches

**Option 1: Sparse-Dense Hybrid Embeddings (Recommended)**

Pinecone supports sparse-dense vectors for hybrid search:

```python
from pinecone_text.sparse import BM25Encoder

bm25_encoder = BM25Encoder()
bm25_encoder.fit(documents)

dense_vector = get_openai_embedding(text)
sparse_vector = bm25_encoder.encode_queries(text)

index.upsert(vectors=[("doc-1", dense_vector, sparse_vector, metadata)])

results = index.query(
    vector=query_dense,
    sparse_vector=query_sparse,
    top_k=10
)
```

Status: Requires `pinecone-text` library. Not yet available in weave-cli.

**Option 2: Metadata Text Filtering**

```bash
weave query semantic TechDocs "neural networks" --metadata "category=tutorial"
```

Available today but limited to exact/prefix matching (no term frequency scoring).

**Option 3: External BM25 Service + Vector Reranking**

```bash
# 1. BM25 search in Elasticsearch
curl -X POST "localhost:9200/docs/_search" -d '{
  "query": {"match": {"text": "machine learning"}}
}' > bm25_results.json

# 2. Vector search in Pinecone
weave query semantic MyIndex "machine learning" > vector_results.json

# 3. Merge and rerank
python merge_results.py bm25_results.json vector_results.json
```

**Option 4: Switch to Hybrid-Native Database**

| Database | BM25 | Vector | Hybrid | Serverless |
|----------|------|--------|--------|------------|
| **Elasticsearch** | Yes | Yes (HNSW) | Yes (RRF) | Elastic Cloud |
| **OpenSearch** | Yes | Yes (HNSW) | Yes (RRF) | AWS |
| **Weaviate** | Yes | Yes (HNSW) | Yes | Cloud option |

```bash
weave collection export MyIndex --pinecone > data.json
export VECTOR_DB_TYPE=elasticsearch-cloud
weave collection import MyIndex data.json
```

Note: Pinecone is cloud-only. For self-hosted needs, consider Weaviate or Elasticsearch.

---

## Cross-Database Comparison

| Feature | Weaviate | Elasticsearch | Qdrant | Neo4j | Pinecone | Chroma |
|---------|----------|---------------|--------|-------|----------|--------|
| **BM25 Native** | Yes | Yes | No | Lucene FTS | No | No |
| **Vector Search** | Excellent | Good | Excellent | Good | Excellent | Good |
| **Hybrid Search** | Native | RRF | Sparse (exp.) | Graph+Vector | Sparse-dense | No |
| **Self-Hosted** | Yes | Yes | Yes | Yes | No | Yes |
| **Setup** | Easy | Medium | Easy | Medium | Easy | Easy |

---

## Additional Resources

- **Weaviate Setup**: [../vdbs/weaviate/SETUP.md](../vdbs/weaviate/SETUP.md)
- **Elasticsearch Setup**: [../vdbs/elasticsearch/SETUP.md](../vdbs/elasticsearch/SETUP.md)
- **Supabase BM25 Improvement**: [../vdbs/supabase/BM25_IMPROVEMENT.md](../vdbs/supabase/BM25_IMPROVEMENT.md)
- **VDB Support Matrix**: [../VDB_SUPPORT_MATRIX.md](../VDB_SUPPORT_MATRIX.md)
