# OpenSearch Integration Setup

This guide explains how to configure and use OpenSearch (local and cloud) with weave-cli.

## Overview

OpenSearch is an open-source search and analytics engine with vector search capabilities using k-NN (k-nearest neighbors) plugin. Weave-CLI supports both local OpenSearch instances and cloud deployments (AWS OpenSearch Service, OpenSearch Cloud).

## Features

- ✅ Local OpenSearch support (single-node development)
- ✅ Cloud OpenSearch support (AWS OpenSearch Service, managed OpenSearch)
- ✅ k-NN vector search using HNSW algorithm
- ✅ Multiple distance metrics (l2, cosinesimil, innerproduct, l1, linf)
- ✅ Collection/index management
- ✅ Document CRUD operations
- 🚧 Advanced search features (k-NN, BM25, hybrid) - coming soon

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

1. **Configure Cloud Credentials**:
   ```bash
   # Using basic auth (username/password)
   export OPENSEARCH_CLOUD_ADDRESS="https://your-domain.us-east-1.es.amazonaws.com"
   export OPENSEARCH_CLOUD_USERNAME="admin"
   export OPENSEARCH_CLOUD_PASSWORD="your-secure-password"
   export VECTOR_DB_TYPE="opensearch-cloud"

   # OR using API key
   export OPENSEARCH_CLOUD_ADDRESS="https://your-domain.us-east-1.es.amazonaws.com"
   export OPENSEARCH_CLOUD_API_KEY="your-api-key"
   export VECTOR_DB_TYPE="opensearch-cloud"
   ```

2. **Verify Connection**:
   ```bash
   weave health check --opensearch-cloud
   ```

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

# Delete document
weave docs del my_documents doc_123 --opensearch-local
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

## Limitations and Roadmap

### Current Limitations
- Search operations (k-NN, BM25, hybrid) are stubbed - implementation coming soon
- Document parsing for RawMessage responses needs enhancement
- Bulk operations use sequential processing (not optimized)
- AWS Signature V4 authentication not yet implemented

### Planned Features
- ✅ Basic collection and document operations
- 🚧 k-NN vector search with embeddings
- 🚧 BM25 text search
- 🚧 Hybrid search (vector + BM25)
- 🚧 Metadata filtering
- 🚧 Optimized bulk operations
- 🚧 AWS IAM authentication (Signature V4)
- 🚧 Advanced index settings
- 🚧 Index templates and aliases

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
