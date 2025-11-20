# Milvus Integration

Weave CLI supports Milvus as a vector database backend, providing powerful vector search capabilities with support for both local development and cloud deployment (Zilliz).

## Overview

Milvus is an open-source vector database built for AI applications, offering:

- **High Performance**: Optimized for large-scale vector similarity search
- **Multiple Index Types**: IVF_FLAT, IVF_SQ8, IVF_PQ, HNSW, and more
- **Hybrid Search**: Combines vector similarity and BM25 text search with RRF
- **Flexible Deployment**: Run locally via Docker/Podman or use managed Zilliz cloud
- **Production Ready**: ACID transactions, high availability, horizontal scaling

## Quick Start

### Local Development

```bash
# Start Milvus using podman (preferred)
./tools/vdb/local/milvus.sh start

# Or use the manager for multi-VDB scenarios
./tools/vdb/local/manager.sh start milvus

# Verify it's running
./tools/vdb/health.sh milvus

# Use with weave CLI
weave --milvus-local collections list
```

### Cloud Deployment (Zilliz)

```bash
# Set environment variables
export MILVUS_CLOUD_ADDRESS="your-cluster.zilliz.com:19530"
export MILVUS_CLOUD_USERNAME="your-username"
export MILVUS_CLOUD_PASSWORD="your-password"

# Use with weave CLI
weave --milvus-cloud collections list
```

## Configuration

### Local Configuration

Copy `configs/config.milvus-local.yaml` to use as a starting point, or create
your own:

```yaml
vectordb:
  type: milvus-local
  address: localhost:19530
  database: default
  timeout: 10
  vector_dimensions: 1536
  similarity_metric: L2
```

### Cloud Configuration

Copy `configs/config.milvus-cloud.yaml` to use as a starting point, or create
your own:

```yaml
vectordb:
  type: milvus-cloud
  address: ${MILVUS_CLOUD_ADDRESS}
  username: ${MILVUS_CLOUD_USERNAME}
  password: ${MILVUS_CLOUD_PASSWORD}
  database: default
  timeout: 30
  vector_dimensions: 1536
  similarity_metric: COSINE
```

## Features

### Document Operations

```bash
# Create collection
weave --milvus-local collections create my_collection

# Add documents
weave --milvus-local documents create my_collection \
  --text "Machine learning powers modern AI" \
  --metadata '{"category":"ai","tags":["ml","ai"]}'

# List documents
weave --milvus-local documents list my_collection

# Delete documents
weave --milvus-local documents delete my_collection doc-id
```

### Search Operations

```bash
# Vector search
weave --milvus-local search vector my_collection \
  --query "artificial intelligence" \
  --limit 5

# BM25 text search
weave --milvus-local search bm25 my_collection \
  --query "machine learning" \
  --limit 5

# Hybrid search (combines vector + BM25 with RRF)
weave --milvus-local search hybrid my_collection \
  --query "deep learning applications" \
  --limit 10
```

### Similarity Metrics

Milvus supports three similarity metrics:

- **L2** (Euclidean distance): Measures straight-line distance between vectors
- **IP** (Inner product): Measures vector alignment (not normalized)
- **COSINE**: Measures angular similarity (normalized inner product)

## Architecture

### Schema Design

Milvus uses explicit schemas with typed fields:

- `document_id`: Primary key (VARCHAR, max 256)
- `text`: Short text content (VARCHAR, max 65535)
- `content`: Full document content (VARCHAR, max 65535)
- `image`: Image URL (VARCHAR, max 512)
- `image_data`: Base64 encoded image (VARCHAR, max 65535)
- `url`: Document URL (VARCHAR, max 512)
- `embedding`: Vector field (FLOAT_VECTOR, 1536 dimensions)
- `metadata`: Additional metadata (JSON)
- `created_at`: Creation timestamp (INT64)
- `updated_at`: Update timestamp (INT64)

### Indexing

Collections use IVF_FLAT index by default:
- **Index Type**: IVF_FLAT (Inverted File with Flat compression)
- **nlist**: 128 (number of cluster units)
- **Search Efficiency**: Good balance of speed and accuracy

### Automatic Embeddings

When `OPENAI_API_KEY` is set, Weave automatically generates embeddings:
- **Model**: text-embedding-3-small
- **Dimensions**: 1536
- **Conversion**: float64 → float32 for Milvus compatibility

## Advanced Topics

### Reciprocal Rank Fusion (RRF)

Hybrid search combines vector and BM25 results using RRF:

```
RRF_score = Σ (1 / (k + rank_i))
```

Where:
- `k = 60` (standard RRF constant)
- `rank_i` is the rank in each result set

### Container Runtime

The local setup automatically detects and uses the best available runtime:

1. **Podman** (preferred): Better security, rootless mode
2. **Docker** (fallback): Widely available, good compatibility

Detection is handled by `tools/vdb/container/detect.sh`.

### Data Persistence

Local Milvus data is stored in named volumes:
- `milvus_etcd`: Metadata storage
- `milvus_minio`: Object storage for vectors
- `milvus_data`: Main data directory

## Documentation

- [Local Setup Guide](LOCAL_SETUP.md) - Detailed local development setup
- [Cloud Setup Guide](CLOUD_SETUP.md) - Zilliz cloud deployment
- [Official Milvus Docs](https://milvus.io/docs) - Complete Milvus documentation

## Troubleshooting

### Connection Issues

```bash
# Check if Milvus is running
./tools/vdb/local/milvus.sh status

# View logs
./tools/vdb/local/milvus.sh logs

# Restart Milvus
./tools/vdb/local/milvus.sh restart
```

### Common Problems

**Error: "failed to connect to Milvus"**
- Verify address is correct (default: localhost:19530)
- Check if Milvus container is running
- Ensure ports are not blocked by firewall

**Error: "collection already exists"**
- Use different collection name
- Or delete existing collection first

**Error: "dimension mismatch"**
- Ensure all embeddings have same dimensions (1536 for text-embedding-3-small)
- Check VectorDimensions in config matches your embedding model

## Performance Tips

1. **Batch Operations**: Use bulk insert for multiple documents
2. **Index Tuning**: Adjust nlist based on dataset size (sqrt(N) is a good starting point)
3. **Memory Management**: Load only frequently accessed collections
4. **Metric Selection**: Use COSINE for normalized vectors, L2 for absolute distance

## References

- [Milvus Official Site](https://milvus.io/)
- [Milvus GitHub](https://github.com/milvus-io/milvus)
- [Zilliz Cloud](https://zilliz.com/)
- [Go SDK Documentation](https://github.com/milvus-io/milvus-sdk-go)
