# Elasticsearch Integration

Elasticsearch is a powerful search and analytics engine with native support for dense vector search (kNN) and BM25 full-text search. Weave CLI supports both local Elasticsearch deployments and Elastic Cloud.

## Status

🚧 **In Progress** (71% complete - 5/7 phases)
- ✅ Phase 1-2: Core infrastructure + Collection operations
- ✅ Phase 3-5: Document, Query, Schema operations
- ⏳ Phase 6: Integration tests (pending)
- ⏳ Phase 7: Documentation (in progress)

## Features

| Feature | Support | Notes |
|---------|---------|-------|
| **Vector Search** | ✅ | Dense vectors with HNSW indexing |
| **BM25 Full-Text** | ✅ | Native Elasticsearch BM25 |
| **Hybrid Search** | ✅ | Combined kNN + BM25 queries |
| **Metadata Filtering** | ✅ | Term and range queries |
| **Batch Operations** | ✅ | BulkIndexer for performance |
| **Collection Management** | ✅ | Index create/delete/list |
| **Schema Management** | ✅ | Mappings and settings |
| **Auto Embeddings** | ✅ | OpenAI integration |

## Quick Start

### Local Elasticsearch

```bash
# 1. Start Elasticsearch
docker run -d \
  --name elasticsearch \
  -p 9200:9200 \
  -e "discovery.type=single-node" \
  -e "xpack.security.enabled=false" \
  docker.elastic.co/elasticsearch/elasticsearch:8.11.0

# 2. Create config
weave config create --env
# Select: elasticsearch-local
# ELASTICSEARCH_LOCAL_ADDRESS: http://localhost:9200
# OPENAI_API_KEY: sk-...

# 3. Verify
weave health check

# 4. Create collection
weave cols create MyDocs --text
```

### Elastic Cloud

```bash
# 1. Get Cloud ID and API key from Elastic Cloud console

# 2. Create config
weave config create --env
# Select: elasticsearch-cloud
# ELASTICSEARCH_CLOUD_ID: deployment:base64string
# ELASTICSEARCH_CLOUD_API_KEY: your-api-key
# OPENAI_API_KEY: sk-...

# 3. Verify
weave health check

# 4. Use it
weave cols ls
```

## Setup Guides

- **[Local Setup](LOCAL_SETUP.md)** - Docker, Homebrew, binary
- **[Cloud Setup](CLOUD_SETUP.md)** - Elastic Cloud configuration

## Architecture

Weave CLI uses the official `go-elasticsearch` v9 client with:
- **TypedClient API** - Type-safe operations
- **Dense vector fields** - HNSW indexing for kNN search
- **BulkIndexer** - High-performance batch operations
- **No CGO dependencies** - Pure Go implementation

## Similarity Metrics

Elasticsearch supports multiple similarity metrics for vector search:

| Metric | Config Value | Use Case |
|--------|--------------|----------|
| Cosine | `cosine` | Normalized embeddings (recommended) |
| Dot Product | `dot_product` | Pre-normalized vectors |
| L2 Norm | `l2_norm` | Euclidean distance |

## Configuration

Example config file (`configs/config.elasticsearch-local.yaml`):

```yaml
databases:
  default: elasticsearch
  vector_databases:
    - name: elasticsearch
      type: elasticsearch-local
      address: http://localhost:9200
      timeout: 30
      vector_dimensions: 1536  # OpenAI text-embedding-3-small
      similarity_metric: cosine
```

See [SETUP.md](SETUP.md) for detailed configuration options.

## Common Operations

```bash
# Collections
weave cols create MyCollection --text
weave cols ls --elasticsearch-local
weave cols rm MyCollection

# Documents
weave docs create MyCollection document.txt
weave docs get MyCollection doc-id
weave docs update MyCollection doc-id --content "updated text"
weave docs rm MyCollection doc-id

# Search
weave cols q MyCollection "search query"  # Semantic search
weave cols q MyCollection "query" --bm25  # BM25 full-text
weave cols q MyCollection "query" --hybrid # Hybrid (kNN + BM25)
```

## Limitations

- **Schema updates**: Cannot modify vector dimensions after index creation
- **Batch operations**: Large batches may require increased timeout values
- **Memory**: Elasticsearch requires adequate heap size for vector search (2GB+ recommended)

## Resources

- [Elasticsearch Documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/index.html)
- [Vector Search Guide](https://www.elastic.co/guide/en/elasticsearch/reference/current/knn-search.html)
- [go-elasticsearch Client](https://github.com/elastic/go-elasticsearch)
- [Elastic Cloud](https://cloud.elastic.co/)

## Support

- Implementation status: See [ARCHITECTURE.md](ARCHITECTURE.md)
- Issues: [GitHub Issues](https://github.com/maximilien/weave-cli/issues)
- Research notes: [RESEARCH.md](RESEARCH.md)
