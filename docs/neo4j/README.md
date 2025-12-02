# Neo4j Vector Database Integration

## Overview

Neo4j is a graph database with powerful vector search capabilities added in version 5.11+. This integration allows you to store documents with vector embeddings and perform semantic search while leveraging Neo4j's graph capabilities for complex relationship queries.

**Key Features:**
- Native graph database with vector search
- Combines graph relationships with semantic search
- High-performance vector similarity search using HNSW indexes
- Support for cosine similarity metric
- Flexible property-based metadata filtering
- ACID transactions for data consistency

**Limitations:**
- BM25 keyword search not supported
- Hybrid search not supported
- Requires Neo4j 5.11 or later for vector support
- Vector dimensions must be consistent across all documents in a collection

## Version Requirements

- **Neo4j**: 5.11+ (for vector search support)
- **Weave CLI**: 0.7.1+
- **Go Driver**: neo4j-go-driver/v5

## Quick Start

### 1. Start Neo4j Locally

Using the included management script:

```bash
# Start Neo4j with Docker/Podman
./tools/vdb/local/neo4j.sh start

# Check status
./tools/vdb/local/neo4j.sh status

# View logs
./tools/vdb/local/neo4j.sh logs

# Stop Neo4j
./tools/vdb/local/neo4j.sh stop

# Clean up data (WARNING: destroys all data)
./tools/vdb/local/neo4j.sh clean
```

### 2. Configure Environment Variables

Add to your `.env` file:

```bash
# Neo4j Local Configuration
NEO4J_URL=bolt://localhost:7687
NEO4J_USERNAME=neo4j
NEO4J_PASSWORD=weaveneo4j
NEO4J_DATABASE=neo4j
```

**Important:** Change the default password for production use!

### 3. Verify Connection

```bash
# Check Neo4j health
weave health check --neo4j-local

# Expected output:
# ✅ Database connection is healthy!
# ✅ Successfully connected to neo4j-local at bolt://localhost:7687
```

### 4. Create Your First Collection

```bash
# Create a text collection
weave cols create MyDocs --text --json-metadata --neo4j-local

# List collections
weave cols list --neo4j-local
```

### 5. Add Documents

```bash
# Add a single document
weave docs create MyDocs document.txt --neo4j-local

# Add documents from a directory
weave docs batch MyDocs ./documents --neo4j-local

# List documents
weave docs list MyDocs --neo4j-local
```

## Architecture

### How Vectors Are Stored

Neo4j stores documents as nodes with properties:
- `id`: Document identifier (string)
- `content`: Document text content (string)
- `embedding`: Vector representation (array of float32)
- Custom properties: Metadata fields (various types)

Example Cypher representation:
```cypher
CREATE (d:MyDocs {
  id: 'doc-123',
  content: 'This is my document',
  embedding: [0.1, 0.2, 0.3, ...],
  author: 'John Doe',
  category: 'technical'
})
```

### Vector Indexes

Vector search is powered by HNSW (Hierarchical Navigable Small World) indexes:

```cypher
CREATE VECTOR INDEX myDocsVectorIndex
FOR (n:MyDocs) ON (n.embedding)
OPTIONS {
  indexConfig: {
    `vector.dimensions`: 1536,
    `vector.similarity_function`: 'cosine'
  }
}
```

The Weave CLI automatically creates these indexes when you create a collection.

### Embedding Generation

Unlike some vector databases, Neo4j does not generate embeddings automatically. The Weave CLI handles this by:

1. Using OpenAI's API to generate embeddings (requires `OPENAI_API_KEY`)
2. Using the `text-embedding-3-small` model (1536 dimensions) by default
3. Storing both the original text content and the generated vector

This client-side approach gives you full control over the embedding model and ensures consistency.

## Configuration

### Complete Configuration Example

```yaml
databases:
  default: neo4j
  vector_databases:
    - name: neo4j
      type: neo4j-local
      url: ${NEO4J_URL:-bolt://localhost:7687}
      username: ${NEO4J_USERNAME:-neo4j}
      password: ${NEO4J_PASSWORD}
      database: ${NEO4J_DATABASE:-neo4j}
      timeout: 30
      vector_dimensions: 1536
      similarity_metric: cosine
```

### Configuration Options

| Option | Description | Default | Required |
|--------|-------------|---------|----------|
| `url` | Bolt protocol URL | `bolt://localhost:7687` | Yes |
| `username` | Neo4j username | `neo4j` | Yes |
| `password` | Neo4j password | - | Yes |
| `database` | Database name | `neo4j` | No |
| `timeout` | Connection timeout (seconds) | `30` | No |
| `vector_dimensions` | Embedding dimensions | `1536` | No |
| `similarity_metric` | Similarity function | `cosine` | No |

**Similarity Metrics:**
- `cosine`: Cosine similarity (recommended)
- `euclidean`: Euclidean distance

### Environment Variables

Priority order: Command flags > `--env file` > `.env` file > Shell environment

```bash
# Connection
NEO4J_URL=bolt://localhost:7687
NEO4J_USERNAME=neo4j
NEO4J_PASSWORD=your-password-here
NEO4J_DATABASE=neo4j

# Embedding generation (required)
OPENAI_API_KEY=sk-...
```

## Basic Operations

### Collection Management

```bash
# Create collection with default settings
weave cols create Articles --text --json-metadata --neo4j-local

# List all collections
weave cols list --neo4j-local

# Delete collection and all documents
weave cols delete Articles --neo4j-local
```

### Document Operations

```bash
# Create single document
weave docs create Articles article.txt --neo4j-local

# Create with custom embedding model
weave docs create Articles article.txt \
  --embedding text-embedding-3-large \
  --neo4j-local

# List documents
weave docs list Articles --neo4j-local
weave docs list Articles --limit 10 --neo4j-local

# Show specific document
weave docs show Articles <document-id> --neo4j-local

# Count documents
weave docs count Articles --neo4j-local

# Delete document
weave docs delete Articles <document-id> --neo4j-local

# Delete all documents
weave docs delete-all Articles --neo4j-local
```

### Batch Operations

```bash
# Process directory of documents
weave docs batch Articles ./documents --neo4j-local

# Process with specific file types
weave docs batch Articles ./documents \
  --pattern "*.pdf" \
  --neo4j-local

# Process with custom chunk size
weave docs batch Articles ./documents \
  --chunk-size 1000 \
  --neo4j-local
```

## Advanced Usage

### Semantic Search

Semantic search uses vector similarity to find relevant documents:

```go
// Example using the Go client
import "github.com/maximilien/weave-cli/src/pkg/vectordb/neo4j"

// Create adapter
adapter := neo4j.NewAdapter(client, llmClient)

// Perform semantic search
results, err := adapter.SearchSemantic(ctx, "Articles", "machine learning", &vectordb.QueryOptions{
    TopK: 10,
})

for _, result := range results {
    fmt.Printf("Score: %.4f - %s\n", result.Score, result.Document.Text)
}
```

**Note:** CLI semantic search command coming in future release.

### Metadata Filtering

Search documents by metadata properties:

```go
// Search by metadata
results, err := adapter.SearchByMetadata(ctx, "Articles", map[string]interface{}{
    "category": "technology",
    "year": 2024,
}, &vectordb.QueryOptions{
    TopK: 20,
})
```

### Combined Graph + Vector Queries

Leverage Neo4j's graph capabilities alongside vector search:

```cypher
// Find similar documents connected by relationships
MATCH (d:Articles)
WHERE d.author = 'John Doe'
CALL db.index.vector.queryNodes('articlesVectorIndex', 10, d.embedding)
YIELD node, score
RETURN node.content, score
```

## Performance Tuning

### Vector Index Optimization

HNSW index parameters affect search performance:

```cypher
CREATE VECTOR INDEX articlesVectorIndex
FOR (n:Articles) ON (n.embedding)
OPTIONS {
  indexConfig: {
    `vector.dimensions`: 1536,
    `vector.similarity_function`: 'cosine',
    `vector.hnsw.m`: 16,              // Number of connections per layer
    `vector.hnsw.ef_construction`: 64  // Construction-time search width
  }
}
```

- Higher `m`: Better recall, more memory, slower writes
- Higher `ef_construction`: Better index quality, slower construction
- Defaults are good for most use cases

### Batch Insertion

For large datasets, use batch operations:

```bash
# Batch process with progress reporting
weave docs batch LargeDataset ./data \
  --batch-size 100 \
  --neo4j-local
```

### Connection Pooling

The Neo4j driver automatically manages connection pooling. For high-concurrency workloads, adjust pool size in code:

```go
config := &neo4j.Config{
    MaxConnectionPoolSize: 50,
    ConnectionTimeout: 30 * time.Second,
}
```

## Troubleshooting

### Common Issues

**1. Authentication Failed**
```
Neo4jError: Neo.ClientError.Security.Unauthorized
```

**Solution:** Verify credentials in `.env`:
```bash
# Check environment variables
source .env
echo $NEO4J_PASSWORD

# Test connection
weave health check --neo4j-local
```

**2. Vector Index Not Found**
```
Failed to create vector index for collection
```

**Solution:** Ensure Neo4j version is 5.11+:
```bash
# Check Neo4j version in browser UI
open http://localhost:7474

# Or via Cypher
CALL dbms.components() YIELD name, versions
```

**3. Dimension Mismatch**
```
Vector dimension mismatch
```

**Solution:** All documents in a collection must use the same embedding model and dimensions. Delete and recreate the collection if you need to change dimensions.

**4. Connection Timeout**
```
Failed to connect to bolt://localhost:7687
```

**Solution:** Verify Neo4j is running:
```bash
./tools/vdb/local/neo4j.sh status

# Start if not running
./tools/vdb/local/neo4j.sh start
```

**5. Out of Memory**
```
Java heap space error
```

**Solution:** Increase Neo4j memory allocation in `neo4j.sh`:
```bash
# Edit docker run command in tools/vdb/local/neo4j.sh
NEO4J_HEAP_INITIAL_SIZE=2G
NEO4J_HEAP_MAX_SIZE=4G
```

### Debug Mode

Enable verbose logging:

```bash
# Run commands with verbose flag
weave health check --neo4j-local -v

# Check Neo4j logs
./tools/vdb/local/neo4j.sh logs
```

### Performance Issues

**Slow Queries:**
1. Verify vector index exists: `SHOW INDEXES`
2. Check index status: `SHOW INDEX STATUS`
3. Increase connection timeout
4. Reduce `TopK` for searches

**High Memory Usage:**
1. Limit batch sizes
2. Reduce vector dimensions (requires recreating collections)
3. Use connection pooling appropriately

## Testing

### Unit Tests

The Neo4j integration includes comprehensive tests:

```bash
# Run Neo4j tests only
./test.sh integration --neo4j

# Skip Neo4j tests
./test.sh integration --skip neo4j

# Run all integration tests
./test.sh integration
```

### Test Coverage

- ✅ Basic connectivity
- ✅ Collection operations (create, list, delete)
- ✅ Document operations (CRUD, batch)
- ✅ Vector search
- ✅ Metadata filtering
- ✅ Vector search with filters

### Manual Testing Checklist

```bash
# 1. Connection test
weave health check --neo4j-local

# 2. Collection lifecycle
weave cols create TestCol --text --json-metadata --neo4j-local
weave cols list --neo4j-local
weave cols delete TestCol --neo4j-local

# 3. Document operations
echo "Test document" > test.txt
weave docs create TestCol test.txt --neo4j-local
weave docs list TestCol --neo4j-local
weave docs count TestCol --neo4j-local

# 4. Cleanup
weave cols delete TestCol --neo4j-local
```

## Security Best Practices

1. **Change Default Password:** Never use `weaveneo4j` in production
2. **Use Environment Variables:** Don't hardcode credentials
3. **Enable TLS:** Use `bolt+s://` for encrypted connections
4. **Restrict Network Access:** Use firewall rules to limit access
5. **Regular Updates:** Keep Neo4j updated for security patches
6. **Audit Logs:** Enable Neo4j audit logging for compliance

## Limitations

### Current Limitations

- ❌ **BM25 Search:** Not supported (Neo4j focuses on graph and vector)
- ❌ **Hybrid Search:** Not available in this integration
- ❌ **Auto-embedding:** Requires OpenAI API key for embedding generation
- ⚠️ **Schema Updates:** Cannot change vector dimensions after creation
- ⚠️ **Cloud Support:** Neo4j Aura (cloud) not yet tested

### Comparison with Other Vector Databases

| Feature | Neo4j | Weaviate | Qdrant | Milvus |
|---------|-------|----------|--------|--------|
| Vector Search | ✅ | ✅ | ✅ | ✅ |
| BM25 Search | ❌ | ✅ | ❌ | ❌ |
| Graph Queries | ✅ | ❌ | ❌ | ❌ |
| ACID Transactions | ✅ | ❌ | ❌ | ❌ |
| Native Embedding | ❌ | ✅ | ❌ | ❌ |
| Scalability | 🟡 | 🟢 | 🟢 | 🟢 |

**Legend:** ✅ Full support | 🟡 Limited | ❌ Not supported | 🟢 Excellent | 🟡 Good

## Resources

### Documentation

- [Neo4j Vector Search Guide](https://neo4j.com/docs/cypher-manual/current/indexes-for-vector-search/)
- [Neo4j Go Driver Docs](https://neo4j.com/docs/go-manual/current/)
- [Weave CLI Neo4j Implementation](../../src/pkg/vectordb/neo4j/)

### Tools

- [Neo4j Browser](http://localhost:7474) - Web UI for local instance
- [Neo4j Desktop](https://neo4j.com/download/) - Desktop application
- [Cypher Query Language](https://neo4j.com/docs/cypher-manual/current/)

### Examples

See the integration tests for complete examples:
- [Basic Connectivity](../../tests/neo4j_basic_test.go)
- [Collection Operations](../../tests/neo4j_collection_test.go)
- [Document Operations](../../tests/neo4j_document_test.go)
- [Vector Search](../../tests/neo4j_query_test.go)

## Support

### Getting Help

- **GitHub Issues:** [Report bugs or request features](https://github.com/maximilien/weave-cli/issues)
- **Integration Tests:** Run `./test.sh integration --neo4j` to verify setup
- **Neo4j Community:** [Neo4j Community Forum](https://community.neo4j.com/)

### Contributing

Contributions welcome! See the main [CONTRIBUTING.md](../../CONTRIBUTING.md) for guidelines.

---

**Version:** 0.7.1+
**Last Updated:** 2025-12-02
**Status:** ✅ Production Ready (Local), ⚠️ Cloud Untested
