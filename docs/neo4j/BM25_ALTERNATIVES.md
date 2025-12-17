# BM25 Alternatives for Neo4j

## Overview

**Neo4j does not support native BM25 full-text search** in its vector search implementation. Neo4j is primarily a **graph database** with added vector search capabilities via indexes.

This document outlines alternative approaches for achieving BM25-like functionality when using Neo4j's vector search features.

---

## Why No BM25?

Neo4j's architecture and priorities:
- **Graph-first**: Optimized for relationships and graph traversal
- **Cypher Query Language**: Declarative graph pattern matching
- **Vector Search Add-on**: Vectors are a newer feature, not core functionality
- **Text Search Alternative**: Cypher has text matching (but not BM25 scoring)

Neo4j focuses on combining **graph relationships** with **vector similarity**, rather than traditional text search ranking algorithms.

---

## Alternative Approaches

### Option 1: Cypher Text Search (Native)

Neo4j provides **Cypher text matching** for keyword search - not BM25, but useful for filtering.

**How It Works**:
```cypher
// Create full-text index (one-time setup)
CREATE FULLTEXT INDEX documentText FOR (n:Document) ON EACH [n.text];

// Search with text matching
CALL db.index.fulltext.queryNodes("documentText", "machine learning")
YIELD node, score
RETURN node.id, node.text, score
ORDER BY score DESC
LIMIT 10;
```

**Via weave-cli**:
```bash
# Create collection with text
weave collection create Articles --text

# Add documents (stored as graph nodes)
weave document create Articles ml-guide.txt ai-tutorial.txt

# Semantic search + Cypher text filter
weave query semantic Articles "neural networks"
# (Note: Full-text index queries require direct Cypher currently)
```

**Capabilities**:
- Lucene-based full-text search
- Term matching with scoring
- Supports wildcards and fuzzy search
- Can combine with graph traversal

**Limitations**:
- Scoring is Lucene-based, not pure BM25
- Weave CLI doesn't expose full-text index queries yet
- Requires direct Cypher for advanced usage

---

### Option 2: Graph + Vector Hybrid Search

Leverage Neo4j's unique strength: combining graph relationships with vector search.

**Approach**:
```cypher
// Find semantically similar documents within graph context
MATCH (doc:Document {category: 'tutorial'})-[:RELATES_TO]->(topic:Topic)
WHERE topic.name = 'Machine Learning'
CALL db.index.vector.queryNodes('documentVectors', 5, $queryVector)
YIELD node, score
WHERE node IN [doc]
RETURN node.text, score
ORDER BY score DESC
LIMIT 10;
```

**Use Case**: Text search within a specific part of the graph (e.g., documents related to a user or topic).

**Pros**:
- Leverages Neo4j's graph strengths
- Combines relationships + semantic search
- More powerful than keyword search alone

**Cons**:
- Requires Cypher knowledge
- Not traditional BM25 keyword ranking
- Weave CLI doesn't support complex Cypher yet

---

### Option 3: External Search Engine + Neo4j

Combine Neo4j with an external full-text search engine for true hybrid search.

**Architecture**:
1. **Elasticsearch/OpenSearch**: BM25 keyword search
2. **Neo4j**: Graph-aware vector search and relationship filtering

**Workflow**:
```bash
# Step 1: BM25 search in Elasticsearch
curl -X POST "localhost:9200/docs/_search" -d '{
  "query": {"match": {"text": "machine learning"}},
  "size": 100
}' | jq '.hits.hits[]._id' > candidate_ids.txt

# Step 2: Neo4j graph + vector search
# (Using Cypher directly)
MATCH (doc:Document)
WHERE doc.id IN $candidateIds
CALL db.index.vector.queryNodes('documentVectors', 10, $queryVector)
YIELD node, score
WHERE node = doc
RETURN doc, score
ORDER BY score DESC;
```

**Pros**:
- True BM25 + graph + vector hybrid
- Production-grade approach
- Leverages unique strengths of each system

**Cons**:
- High operational complexity (2 databases)
- Data synchronization overhead
- Increased latency

---

### Option 4: Switch to Hybrid Database

If BM25 is critical and graph features are not essential, consider hybrid databases:

| Database | BM25 | Vector | Graph | Hybrid Search |
|----------|------|--------|-------|---------------|
| **Weaviate** | ✅ | ✅ | ❌ | ✅ |
| **Elasticsearch** | ✅ | ✅ | ❌ | ✅ |
| **OpenSearch** | ✅ | ✅ | ❌ | ✅ |
| **Neo4j** | ⚠️ | ✅ | ✅ | ⚠️ |

**Migration Consideration**:
If you're **not using graph relationships**, a pure vector database with BM25 may be simpler:

```bash
# Export from Neo4j
weave collection export KnowledgeBase --neo4j-local > data.json

# Import to Weaviate
export VECTOR_DB_TYPE=weaviate
weave collection import KnowledgeBase data.json

# Use hybrid search
weave query hybrid KnowledgeBase "machine learning" --alpha 0.5
```

**Important**: Only migrate if **graph relationships are not essential** to your use case.

---

## Use Case Recommendations

### When Neo4j Is the Right Choice

✅ **Use Neo4j if**:
- **Graph relationships** are core to your data model
- You need to combine graph traversal + vector search
- Cypher text search meets your keyword needs
- Knowledge graphs, recommendation systems, or network analysis
- Semantic search within graph contexts (e.g., user's network)

### When Neo4j Might Not Fit

⚠️ **Consider alternatives if**:
- You don't have significant graph relationships
- BM25 ranking is critical for relevance
- Pure vector search is the primary use case
- Simpler architecture is preferred

### When to Switch Databases

❌ **Switch from Neo4j if**:
- Graph features are unused (overkill for flat documents)
- BM25 is non-negotiable and Cypher text search insufficient
- Team lacks Cypher/graph database expertise
- Cost/complexity of Neo4j not justified by graph value

---

## Comparison: Neo4j vs Other Options

| Feature | Neo4j | Weaviate | Elasticsearch |
|---------|-------|----------|---------------|
| **BM25 Native** | ⚠️ (Cypher FTS) | ✅ | ✅ |
| **Vector Search** | ✅ | ✅ | ✅ |
| **Graph Database** | ✅ | ❌ | ❌ |
| **Hybrid Search** | ⚠️ (Graph + Vector) | ✅ (BM25 + Vector) | ✅ (BM25 + Vector) |
| **Complexity** | High | Low | Medium |
| **Best For** | Graph + Vector | Pure Hybrid | Enterprise Scale |

---

## Practical Examples

### Example 1: Cypher Full-Text Search (Neo4j Native)

```cypher
// Create full-text index
CREATE FULLTEXT INDEX articleText IF NOT EXISTS
FOR (a:Article) ON EACH [a.title, a.content];

// Query with text search
CALL db.index.fulltext.queryNodes("articleText", "machine learning")
YIELD node, score
WHERE score > 0.5
RETURN node.title, node.content, score
ORDER BY score DESC
LIMIT 10;
```

```bash
# Via weave-cli (limited support currently)
weave collection create Articles --text
weave document create Articles ml-tutorial.txt

# Semantic search (vector-based)
weave query semantic Articles "deep learning concepts"
```

### Example 2: Graph-Context Vector Search

```cypher
// Find similar documents within user's reading history
MATCH (u:User {id: $userId})-[:READ]->(doc:Document)
CALL db.index.vector.queryNodes('docVectors', 5, $queryEmbedding)
YIELD node AS similarDoc, score
WHERE similarDoc <> doc AND similarDoc.category = doc.category
RETURN similarDoc.title, score, COUNT(doc) AS contextRelevance
ORDER BY score DESC, contextRelevance DESC
LIMIT 10;
```

**Use Case**: Personalized recommendations using graph + vectors

### Example 3: Hybrid Architecture (Neo4j + Elasticsearch)

```python
# Hybrid search combining BM25, vector, and graph
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
        LIMIT $topK
    """

    neo4j_results = session.run(
        cypher,
        userId=user_id,
        candidateIds=candidate_ids,
        queryVector=get_embedding(query_text),
        topK=top_k
    )

    return list(neo4j_results)
```

---

## Future Enhancements

**Planned for weave-cli**:
- [ ] Expose Neo4j full-text index queries
- [ ] Cypher text search integration in weave CLI
- [ ] Graph-aware hybrid search utilities
- [ ] Multi-hop graph + vector search patterns

**Tracking Issue**: [#TBD]

---

## Additional Resources

- **Neo4j Vector Search**: [https://neo4j.com/docs/cypher-manual/current/indexes-for-vector-search/](https://neo4j.com/docs/cypher-manual/current/indexes-for-vector-search/)
- **Neo4j Full-Text Indexes**: [https://neo4j.com/docs/cypher-manual/current/indexes-for-full-text-search/](https://neo4j.com/docs/cypher-manual/current/indexes-for-full-text-search/)
- **Neo4j Setup**: [README.md](README.md)
- **Graph + Vector Patterns**: [https://neo4j.com/developer/graph-data-science/](https://neo4j.com/developer/graph-data-science/)

---

## Summary

**Key Takeaways**:
1. Neo4j is **graph-first** with vector search add-on
2. **Cypher full-text search** provides Lucene-based text matching (not pure BM25)
3. **Graph + vector hybrid** is Neo4j's unique strength
4. For **pure BM25 + vector**, simpler databases (Weaviate, Elasticsearch) may be better
5. Use Neo4j when **graph relationships add significant value** to your search

**Recommendation**: Evaluate whether your use case truly benefits from graph relationships. If yes, Neo4j's Cypher text search + graph-aware vector search is powerful. If no, a simpler hybrid database (Weaviate, Elasticsearch) may be more appropriate.
