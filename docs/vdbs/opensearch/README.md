# OpenSearch Integration Guide

**Status**: ✅ Production-ready (v0.9.14+)

This guide explains how to configure and use OpenSearch (local and cloud) with weave-cli.

## Overview

OpenSearch is an enterprise-grade, open-source search and analytics engine with powerful vector search capabilities. Weave-CLI provides full production support for both local OpenSearch instances and cloud deployments (AWS OpenSearch Service).

## Features (v0.9.14+)

### ✅ Fully Supported

- **Local & Cloud Deployment**: Single-node development or AWS OpenSearch Service
- **AWS Signature V4**: Auto-detection and authentication for AWS domains
- **k-NN Vector Search**: High-performance HNSW algorithm
- **Native BM25**: Full-text keyword search
- **Hybrid Search**: Vector + BM25 with Reciprocal Rank Fusion (RRF)
- **Collection Management**: Create, delete, list, exists, count (with accurate stats)
- **Document CRUD**: Full create, get, update, delete operations
- **Parallel Bulk Operations**: 10x faster ingestion with controlled concurrency
- **Metadata Filtering**: Complex queries with bool DSL
- **Auto-Embeddings**: Automatic OpenAI embedding generation
- **Multiple Distance Metrics**: l2, cosinesimil, innerproduct, l1, linf

## Quick Start

### Local OpenSearch

1. **Start Local OpenSearch Server**:
   ```bash
   ./tools/vdb/local/opensearch.sh start
   ```

2. **Configure Environment**:
   ```bash
   export OPENSEARCH_LOCAL_ADDRESS="http://localhost:9200"
   export VECTOR_DB_TYPE="opensearch-local"
   export OPENAI_API_KEY="your-openai-api-key"  # For embeddings
   ```

3. **Verify Connection**:
   ```bash
   weave health check --opensearch-local
   ```

### Cloud OpenSearch (AWS OpenSearch Service)

1. **Configure AWS Credentials** (v0.9.14+ with auto-detection):
   ```bash
   # AWS OpenSearch Service - uses AWS Signature V4 (auto-detected)
   export OPENSEARCH_CLOUD_ADDRESS="https://search-my-domain.us-east-1.es.amazonaws.com"
   export AWS_REGION="us-east-1"  # Or AWS_DEFAULT_REGION
   export VECTOR_DB_TYPE="opensearch-cloud"

   # AWS credentials from AWS CLI profile or environment
   # Option 1: Use AWS CLI profile (recommended)
   aws configure

   # Option 2: Set environment variables
   export AWS_ACCESS_KEY_ID="your-access-key"
   export AWS_SECRET_ACCESS_KEY="your-secret-key"

   # Option 3: Use IAM role (for EC2/ECS/Lambda)
   # No additional configuration needed
   ```

2. **Verify Connection**:
   ```bash
   weave health check --opensearch-cloud
   ```

**Note**: The adapter auto-detects AWS domains (`.amazonaws.com`, `.aoss.`) and automatically uses AWS Signature V4 authentication. No username/password needed for AWS OpenSearch Service.

## Configuration Options

### Environment Variables

#### OpenSearch Local
- `OPENSEARCH_LOCAL_ADDRESS`: Server address (default: `http://localhost:9200`)

#### OpenSearch Cloud
- `OPENSEARCH_CLOUD_ADDRESS`: Cloud endpoint URL (required)
- `OPENSEARCH_CLOUD_USERNAME`: Username for basic auth (optional)
- `OPENSEARCH_CLOUD_PASSWORD`: Password for basic auth (optional)
- `OPENSEARCH_CLOUD_API_KEY`: API key for authentication (alternative to username/password)

### Interactive Configuration

Use the interactive configuration tool:

```bash
# For local OpenSearch
weave config create --env --opensearch-local

# For cloud OpenSearch
weave config create --env --opensearch-cloud
```

## Usage Examples

### Collection Management

```bash
# List all collections (indices)
weave cols ls --opensearch-local

# Create a new collection
weave cols create my_documents --opensearch-local

# Show collection details
weave cols show my_documents --opensearch-local

# Delete a collection
weave cols ds my_documents --opensearch-local
```

### Document Operations

```bash
# Create documents from file
weave docs create my_documents document.txt --opensearch-local

# List documents in collection
weave docs ls my_documents --opensearch-local

# Show specific document
weave docs show my_documents doc_123 --opensearch-local

# Update document
weave docs update my_documents doc_123 --text "Updated content" --opensearch-local

# Delete document by ID
weave docs del my_documents doc_123 --opensearch-local

# Delete documents by metadata filter
weave docs del my_documents --filter '{"category": "outdated"}' --opensearch-local
```

### Search Operations (v0.9.14+)

```bash
# Vector similarity search (k-NN)
weave search semantic my_documents --query "machine learning algorithms" --limit 5 --opensearch-local

# BM25 keyword search
weave search bm25 my_documents --query "vector database" --limit 5 --opensearch-local

# Hybrid search (vector + BM25 with RRF)
weave search hybrid my_documents --query "distributed search" --limit 5 --opensearch-local

# Metadata filtering
weave search metadata my_documents --filter '{"year": 2024, "category": "research"}' --opensearch-local
```

### Batch Ingestion (v0.9.14+ with parallel processing)

```bash
# Ingest PDFs with parallel bulk operations (10x faster)
weave pipeline ingest \
  --collection my_documents \
  --path ./documents \
  --file-types pdf,txt,md \
  --batch-size 100 \
  --workers 10 \
  --opensearch-local

# With resume capability (state management)
weave pipeline ingest \
  --collection my_documents \
  --path ./documents \
  --resume \
  --state-file ./ingest-state.json \
  --opensearch-local
```

## Local OpenSearch Management

### Server Operations

```bash
# Start OpenSearch
./tools/vdb/local/opensearch.sh start

# Stop OpenSearch
./tools/vdb/local/opensearch.sh stop

# Restart OpenSearch
./tools/vdb/local/opensearch.sh restart

# Check status
./tools/vdb/local/opensearch.sh status

# View logs
./tools/vdb/local/opensearch.sh logs

# Health check
./tools/vdb/local/opensearch.sh health
```

### Data Management

```bash
# Remove all data (reset)
./tools/vdb/local/opensearch.sh reset

# Remove container and data
./tools/vdb/local/opensearch.sh clean
```

## Index Configuration

OpenSearch indices created by weave-cli use the following configuration:

### Vector Field Settings
- **Type**: `knn_vector`
- **Dimension**: 384 (default, configurable via VECTOR_DIMENSIONS)
- **Algorithm**: HNSW (Hierarchical Navigable Small World)
- **Engine**: NMSLIB
- **Space Type**: l2 (Euclidean distance, configurable)
- **Parameters**:
  - `ef_construction`: 128
  - `m`: 16
  - `ef_search`: 100

### Distance Metrics

Supported distance metrics for similarity search:
- `l2`: Euclidean distance (default)
- `cosinesimil`: Cosine similarity
- `innerproduct`: Inner product (dot product)
- `l1`: Manhattan distance
- `linf`: Chebyshev distance

Configure via environment:
```bash
export SIMILARITY_METRIC="cosinesimil"
```

## Docker Configuration

The local OpenSearch server runs in Docker/Podman with:
- **Image**: `opensearchproject/opensearch:latest`
- **HTTP Port**: 9200
- **Performance Port**: 9600
- **Security**: Disabled for local development
- **Discovery**: Single-node mode
- **Storage**: Persistent at `./local/storage/opensearch_storage`

### System Requirements

⚠️ **Memory Requirements**: OpenSearch requires significant memory to run reliably:
- **Minimum**: 2GB total system RAM (1.5GB for OpenSearch + 512MB for OS/other processes)
- **Recommended**: 4GB+ total system RAM for stable operation
- **Production**: 8GB+ total system RAM

**Known Limitation**: On systems with < 2GB RAM, OpenSearch may exit with code 137 (OOM killed) even with reduced JVM heap settings. For such systems, consider using cloud OpenSearch or alternative VDBs (Weaviate, Qdrant, Milvus, Chroma).

## Troubleshooting

### Connection Issues

1. **Verify OpenSearch is running**:
   ```bash
   curl http://localhost:9200
   ```

2. **Check container status**:
   ```bash
   docker ps | grep opensearch
   # OR
   podman ps | grep opensearch
   ```

3. **View logs**:
   ```bash
   ./tools/vdb/local/opensearch.sh logs
   ```

### Common Errors

#### "Connection refused"
- Ensure OpenSearch is running: `./tools/vdb/local/opensearch.sh start`
- Verify port 9200 is not blocked by firewall

#### "Cluster status is red"
- Check cluster health: `curl http://localhost:9200/_cluster/health`
- Review logs for errors: `./tools/vdb/local/opensearch.sh logs`

#### "Index creation failed"
- Verify k-NN plugin is enabled
- Check index mapping configuration
- Ensure sufficient disk space

#### Container exits with code 137 (OOM killed)
- **Cause**: System has insufficient memory (< 2GB total RAM)
- **Check system memory**: `docker info | grep "Total Memory"`
- **Solutions**:
  1. Use cloud OpenSearch instead of local (recommended)
  2. Increase system memory to at least 2GB
  3. Use alternative VDB (Weaviate, Qdrant, Milvus) with lower memory requirements
  4. Close other Docker containers: `docker stop $(docker ps -q)`
- **Note**: OpenSearch requires ~1.5GB RAM minimum; systems with < 2GB total RAM cannot run it reliably

## AWS OpenSearch Service

### Setup Steps

1. **Create OpenSearch Domain** in AWS Console:
   - Choose instance type (t3.small.search for dev)
   - Enable fine-grained access control
   - Create master user credentials
   - Configure VPC settings (optional)

2. **Configure Access**:
   - Use master username/password for authentication
   - Or create IAM role with AWS Signature V4 (coming soon)

3. **Get Endpoint URL**:
   - Copy domain endpoint from AWS Console
   - Format: `https://your-domain.us-east-1.es.amazonaws.com`

4. **Test Connection**:
   ```bash
   weave health check --opensearch-cloud
   ```

## Integration Tests

Run OpenSearch integration tests:

```bash
# Start local OpenSearch first
./tools/vdb/local/opensearch.sh start

# Run tests
go test -v -tags=integration ./tests -run="TestOpenSearch"
```

## What's New in v0.9.14

### Production Hardening Features

- ✅ **AWS Signature V4 Authentication**: Auto-detection and signing for AWS OpenSearch Service
- ✅ **Parallel Bulk Operations**: 10x performance improvement with controlled concurrency (semaphore pattern)
- ✅ **Accurate Statistics**: Real collection counts using `Indices.Stats` API
- ✅ **Complete Document Parsing**: Proper RawMessage unmarshaling for all fields
- ✅ **UpdateDocument Support**: Full CRUD operations with upsert
- ✅ **Metadata Filter Deletes**: Delete documents by metadata criteria with bool queries
- ✅ **k-NN Vector Search**: Fully functional with HNSW indexing
- ✅ **BM25 Search**: Native full-text keyword search
- ✅ **Hybrid Search**: RRF fusion of vector and keyword results

### Known Limitations

- **SSL Configuration**: Self-signed certificates for local setup (use HTTP or disable SSL verification)
- **Float32 Vectors**: Embeddings converted from float64 to float32 (standard for most vector DBs)
- **Memory Requirements**: Local OpenSearch requires ~2GB+ RAM

### Future Enhancements

- 🚧 Advanced index templates and aliases
- 🚧 Index lifecycle management (ILM)
- 🚧 Custom analyzers and tokenizers
- 🚧 Geo-spatial search support

## Resources

- [OpenSearch Documentation](https://opensearch.org/docs/latest/)
- [k-NN Plugin Guide](https://opensearch.org/docs/latest/search-plugins/knn/index/)
- [AWS OpenSearch Service](https://docs.aws.amazon.com/opensearch-service/)
- [opensearch-go Client](https://github.com/opensearch-project/opensearch-go)

## Support

For issues or questions:
1. Check logs: `./tools/vdb/local/opensearch.sh logs`
2. Verify configuration: `weave config show`
3. Test health: `weave health check --opensearch-local`
4. Review [GitHub Issues](https://github.com/maximilien/weave-cli/issues)
