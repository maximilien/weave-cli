# MongoDB Atlas Vector Search Integration

> **🚧 Experimental**: MongoDB integration is currently experimental. BM25 search, document operations, and collection management are fully functional. Vector search embedding integration is pending.

This guide covers using MongoDB Atlas Vector Search with weave-cli for semantic search and RAG applications.

## Status

**What Works:**
- ✅ Connection and health checks
- ✅ Collection management (create, list, delete)
- ✅ Document CRUD operations
- ✅ BM25 keyword search (fully functional)
- ✅ Metadata search and filtering

**What's Pending:**
- 🚧 Vector search (embedding generation integration needed)
- 🚧 True hybrid search (depends on vector search)
- 🚧 Integration tests with Atlas

## Overview

MongoDB Atlas Vector Search provides native vector search capabilities within MongoDB, combining the familiarity of a document database with powerful vector similarity search.

### Why MongoDB Atlas?

- **Native Integration**: Vector search built into MongoDB, no separate vector database needed
- **Flexible Schema**: Store documents, metadata, and vectors together
- **Hybrid Search**: Combine vector search with traditional MongoDB queries
- **Cost-Effective**: Free M0 tier available for development and testing
- **Scalability**: Scales from free tier to enterprise deployments
- **Familiar Tools**: Use MongoDB Compass, Atlas UI, and standard MongoDB drivers

### Comparison with Weaviate

| Feature | MongoDB Atlas | Weaviate |
|---------|---------------|----------|
| **Vector Search** | ✅ $vectorSearch aggregation | ✅ Native GraphQL API |
| **BM25 Search** | ✅ Text indexes | ✅ Built-in |
| **Hybrid Search** | ✅ RRF combination | ✅ Native hybrid |
| **Free Tier** | ✅ M0 cluster | ✅ Sandbox |
| **Metadata Filtering** | ✅ Pre-filtering | ✅ Where filters |
| **Schema Flexibility** | ✅ Document model | ⚠️ Schema-based |
| **Learning Curve** | ✅ MongoDB familiarity | ⚠️ New concepts |

## Quick Start

### 1. Setup

See [ATLAS_SETUP.md](ATLAS_SETUP.md) for detailed setup instructions.

**TL;DR:**
1. Create MongoDB Atlas account and cluster (free M0 tier)
2. Create database user and whitelist IP
3. Create vector search index on `embedding` field
4. Configure weave-cli with connection string

### 2. Configure weave-cli

Add to `.env`:

```bash
VECTOR_DB_TYPE=mongodb
MONGODB_URI="mongodb+srv://username:password@cluster.mongodb.net/?appName=weave-cli"
MONGODB_DATABASE="weave-cli"
OPENAI_API_KEY="sk-..."
```

### 3. Test Connection

```bash
# Health check
weave health check

# Create collection
weave collection create WeaveDocs --type text

# List collections
weave collection list
```

## Usage

### Document Operations

```bash
# Create a document
weave document create WeaveDocs \
  --text "MongoDB is a document database with vector search capabilities" \
  --metadata '{"category": "database", "source": "docs"}'

# Create from file
weave document create WeaveDocs --file README.md

# List documents
weave document list WeaveDocs

# Get a specific document
weave document get WeaveDocs <document-id>

# Delete a document
weave document delete WeaveDocs <document-id>
```

### Search Operations

#### Semantic Search (Vector Search)

**Note**: Vector search requires the vector search index to be created in Atlas UI.

```bash
# Semantic search using embeddings
weave query semantic WeaveDocs "What is MongoDB vector search?" --top-k 5

# With metadata filtering (if filter indexes are configured)
weave query semantic WeaveDocs "database features" \
  --metadata '{"category": "database"}'
```

#### BM25 Keyword Search

BM25 search uses MongoDB text indexes (automatically created).

```bash
# BM25 keyword search
weave query bm25 WeaveDocs "MongoDB vector search" --top-k 5

# Works immediately without vector index setup
weave query bm25 WeaveDocs "document database"
```

#### Hybrid Search

Combines vector and keyword search using Reciprocal Rank Fusion (RRF).

```bash
# Hybrid search (best of both worlds)
weave query hybrid WeaveDocs "MongoDB features" --top-k 5
```

### Collection Management

```bash
# Create collections
weave collection create WeaveDocs --type text
weave collection create WeaveImages --type image

# List all collections
weave collection list

# Get collection info
weave collection info WeaveDocs

# Delete collection
weave collection delete WeaveDocs
```

## Architecture

### Document Structure

Documents in MongoDB are stored with this structure:

```json
{
  "_id": ObjectId("..."),
  "document_id": "unique-id",
  "text": "Document text content",
  "content": "Full document content",
  "image": "image-url",
  "image_data": "base64-encoded-data",
  "url": "source-url",
  "embedding": [0.123, -0.456, ...],  // 1536 dimensions for ada-002
  "metadata": {
    "category": "example",
    "source": "upload",
    "custom_field": "value"
  },
  "created_at": ISODate("..."),
  "updated_at": ISODate("...")
}
```

### Vector Search Pipeline

When you perform semantic search, weave-cli:

1. **Generates embedding** using OpenAI API
2. **Executes $vectorSearch** aggregation:
   ```javascript
   db.collection.aggregate([
     {
       $vectorSearch: {
         index: "vector_index",
         path: "embedding",
         queryVector: [...],
         numCandidates: 50,
         limit: 5
       }
     },
     {
       $project: {
         document_id: 1,
         text: 1,
         content: 1,
         metadata: 1,
         score: { $meta: "vectorSearchScore" }
       }
     }
   ])
   ```
3. **Returns results** with similarity scores

### BM25 Search

BM25 uses MongoDB text indexes:

```javascript
db.collection.find(
  { $text: { $search: "query" } },
  { score: { $meta: "textScore" } }
).sort({ score: { $meta: "textScore" } })
```

### Hybrid Search

Combines both approaches using RRF:

1. Get top-k results from vector search
2. Get top-k results from BM25 search
3. Combine using Reciprocal Rank Fusion formula
4. Re-rank and return top-k

## Configuration

### Vector Dimensions

Match your embedding model:

```yaml
vector_databases:
  - name: mongodb
    vector_dimensions: 1536  # OpenAI ada-002
    # vector_dimensions: 3072  # OpenAI text-embedding-3-large
    # vector_dimensions: 512   # OpenAI text-embedding-3-small
```

### Similarity Metrics

```yaml
vector_databases:
  - name: mongodb
    similarity_metric: cosine     # Recommended for normalized embeddings
    # similarity_metric: euclidean  # For unnormalized vectors
    # similarity_metric: dotProduct # Fastest, requires normalized vectors
```

### Timeout Configuration

```yaml
vector_databases:
  - name: mongodb
    timeout: 10  # Seconds for database operations
```

## Performance Optimization

### Index Configuration

**For better performance**, create compound indexes:

```javascript
// Text search performance
db.WeaveDocs.createIndex({ text: "text", content: "text" })

// Metadata filtering
db.WeaveDocs.createIndex({ "metadata.category": 1 })
db.WeaveDocs.createIndex({ "metadata.source": 1 })

// Document ID lookups
db.WeaveDocs.createIndex({ document_id: 1 }, { unique: true })
```

### Vector Search Optimization

- **numCandidates**: Set to 10-20x your desired limit
- **Metadata pre-filtering**: Use filter fields in index definition
- **Smaller embeddings**: Use text-embedding-3-small (512D) for faster searches

### Cluster Tiers

| Tier | Use Case | Performance | Cost |
|------|----------|-------------|------|
| M0 | Development, testing | Limited | Free |
| M2 | Small apps, prototypes | Better | ~$9/month |
| M10 | Production apps | Good | ~$60/month |
| M30+ | Enterprise | Excellent | Varies |

## Limitations

### Free Tier (M0)

- Max 3 vector search indexes
- Max 10M vector dimensions per index
- Shared resources (variable performance)
- No SLA guarantees

### Vector Search

- **Index creation**: Manual process via Atlas UI (not automated via API)
- **Index updates**: Require recreation for dimension/metric changes
- **numCandidates limit**: Performance degrades with very large values

### General

- **Network latency**: Atlas is cloud-hosted (use region close to your app)
- **Connection limits**: Free tier has connection limits
- **Storage limits**: Free tier limited to 512 MB

## Troubleshooting

### Vector Search Not Working

```bash
# Error: index not found
```

**Solution**: Create vector search index in Atlas UI (see [ATLAS_SETUP.md](ATLAS_SETUP.md))

### Connection Timeout

```bash
# Error: connection timeout
```

**Solutions**:
- Check IP whitelist in Atlas Network Access
- Verify connection string format
- Check firewall settings

### Slow Queries

**Solutions**:
- Upgrade to M2+ tier
- Reduce `numCandidates` in vector search
- Create appropriate indexes
- Use metadata pre-filtering

### Dimensions Mismatch

```bash
# Error: dimensions don't match
```

**Solution**: Ensure vector search index dimensions match your embedding model:
- ada-002: 1536
- text-embedding-3-large: 3072
- text-embedding-3-small: configurable (512-1536)

## Best Practices

### 1. Start with BM25

BM25 search works immediately without vector index setup. Use it to test your setup:

```bash
weave query bm25 WeaveDocs "test query"
```

### 2. Create Vector Index Before Vector Search

Vector search requires manual index creation in Atlas UI. Do this before attempting semantic search.

### 3. Use Hybrid Search for Best Results

Hybrid search combines the precision of vector search with the recall of BM25:

```bash
weave query hybrid WeaveDocs "your query"
```

### 4. Monitor Performance

Use MongoDB Atlas monitoring:
- Query performance
- Index usage
- Connection counts
- Resource utilization

### 5. Implement Retry Logic

For production apps, implement retry logic for transient failures.

## Migration Guide

### From Weaviate to MongoDB

1. **Export Weaviate data**:
   ```bash
   weave document list WeaveDocs > documents.json
   ```

2. **Update configuration**:
   ```bash
   # Change VECTOR_DB_TYPE
   export VECTOR_DB_TYPE=mongodb
   ```

3. **Create MongoDB collections and indexes**

4. **Import documents**:
   ```bash
   # Re-create documents (embeddings will be regenerated)
   weave document create WeaveDocs --file documents.json
   ```

### From MongoDB to Weaviate

Same process in reverse.

## Advanced Features

### Custom Metadata Schemas

Store any metadata structure:

```bash
weave document create WeaveDocs \
  --file doc.txt \
  --metadata '{
    "category": "technical",
    "tags": ["mongodb", "vector-search"],
    "author": "Alice",
    "version": "1.0",
    "custom": {
      "nested": "data"
    }
  }'
```

### Batch Operations

```bash
# Batch create from directory
weave document create WeaveDocs --directory ./docs

# Batch delete by metadata
weave document delete WeaveDocs --metadata '{"category": "old"}'
```

## Resources

- [MongoDB Atlas Vector Search Docs](https://www.mongodb.com/docs/atlas/atlas-vector-search/)
- [Vector Search Tutorial](https://www.mongodb.com/docs/atlas/atlas-vector-search/tutorials/)
- [weave-cli MongoDB Setup](ATLAS_SETUP.md)
- [MongoDB Atlas Pricing](https://www.mongodb.com/pricing)

## Support

For issues or questions:
- Check [Troubleshooting](#troubleshooting) section
- Review [MongoDB Atlas documentation](https://www.mongodb.com/docs/atlas/)
- File an issue on [GitHub](https://github.com/maximilien/weave-cli/issues)
