# Qdrant Setup Guide

> **⚠️ Experimental Feature**: Qdrant integration is currently experimental. Core functionality is implemented and tested, but real-world production validation is still in progress.

## Overview

[Qdrant](https://qdrant.tech/) is a high-performance vector database written in Rust. It provides:

- **HNSW Indexing**: Fast and accurate similarity search
- **gRPC API**: High-performance communication
- **Flexible Filtering**: Advanced payload-based filtering
- **Quantization**: Memory-efficient vector storage
- **Local & Cloud**: Docker/Podman deployment or managed cloud service

This guide covers both **Qdrant Local** (self-hosted) and **Qdrant Cloud** (managed service) setup.

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Qdrant Local Setup](#qdrant-local-setup)
3. [Qdrant Cloud Setup](#qdrant-cloud-setup)
4. [Configuration](#configuration)
5. [Usage Examples](#usage-examples)
6. [Troubleshooting](#troubleshooting)
7. [Known Limitations](#known-limitations)

---

## Prerequisites

### Required

- **weave-cli** v0.7.0 or later
- **OpenAI API Key** (for automatic embedding generation)

### For Local Development

- **Docker** or **Podman** (for running Qdrant locally)

### For Cloud

- **Qdrant Cloud Account** (free tier available at [cloud.qdrant.io](https://cloud.qdrant.io))

---

## Qdrant Local Setup

### 1. Start Qdrant with Docker

```bash
# Using Docker
docker run -d \
  --name qdrant \
  -p 6333:6333 \
  -p 6334:6334 \
  -v $(pwd)/qdrant_storage:/qdrant/storage:z \
  qdrant/qdrant

# Using Podman
podman run -d \
  --name qdrant \
  -p 6333:6333 \
  -p 6334:6334 \
  -v $(pwd)/qdrant_storage:/qdrant/storage:z \
  qdrant/qdrant
```

**Ports:**
- `6333`: REST API
- `6334`: gRPC API (used by weave-cli)

### 2. Verify Qdrant is Running

```bash
# Check container status
docker ps | grep qdrant

# Test REST API
curl http://localhost:6333/

# Expected response:
# {"title":"qdrant - vector search engine","version":"..."}
```

### 3. Set Environment Variables

```bash
# Required for automatic embeddings
export OPENAI_API_KEY="sk-..."

# Optional (defaults shown)
export QDRANT_HOST="localhost"
export QDRANT_GRPC_PORT="6334"
```

### 4. Test with weave-cli

```bash
# Health check
./bin/weave health check --qdrant-local

# List collections
./bin/weave --qdrant-local cols ls

# Create a collection
./bin/weave --qdrant-local cols create \
  --name test_collection \
  --vector-dimensions 1536 \
  --similarity-metric Cosine
```

---

## Qdrant Cloud Setup

### 1. Create a Qdrant Cloud Account

1. Go to [cloud.qdrant.io](https://cloud.qdrant.io)
2. Sign up for a free account
3. Create a new cluster

### 2. Get Your Credentials

From your Qdrant Cloud dashboard:

1. **Cluster URL**: e.g., `https://xyz.cloud.qdrant.io:6334`
2. **API Key**: Found in cluster settings

### 3. Set Environment Variables

```bash
# Qdrant Cloud credentials
export QDRANT_URL="https://xyz.cloud.qdrant.io:6334"
export QDRANT_API_KEY="your-api-key-here"

# OpenAI for embeddings
export OPENAI_API_KEY="sk-..."
```

### 4. Test with weave-cli

```bash
# Health check
./bin/weave health check --qdrant-cloud

# List collections
./bin/weave --qdrant-cloud cols ls

# Create a collection
./bin/weave --qdrant-cloud cols create \
  --name test_collection \
  --vector-dimensions 1536 \
  --similarity-metric Cosine
```

---

## Configuration

### Option 1: Generate Config with weave-cli (Easiest)

weave-cli can automatically generate configuration files for you.

#### Generate .env file

```bash
# For local Qdrant
./bin/weave config create --env --vdb qdrant-local

# For Qdrant Cloud
./bin/weave config create --env --vdb qdrant-cloud
```

This creates a `.env` file with the necessary environment variables:

```bash
# For qdrant-local
QDRANT_HOST=localhost
QDRANT_GRPC_PORT=6334
OPENAI_API_KEY=your-openai-api-key

# For qdrant-cloud
QDRANT_URL=https://xyz.cloud.qdrant.io:6334
QDRANT_API_KEY=your-qdrant-api-key
OPENAI_API_KEY=your-openai-api-key
```

#### Generate config.yaml file

```bash
# For local Qdrant
./bin/weave config create --config-yaml --vdb qdrant-local

# For Qdrant Cloud
./bin/weave config create --config-yaml --vdb qdrant-cloud
```

This creates a `config.yaml` file ready to use.

### Option 2: Use Pre-made Config Files

#### Local Configuration

Create or use `configs/config.qdrant-local.yaml`:

```yaml
databases:
  default: qdrant
  vector_databases:
    - name: qdrant
      type: qdrant-local
      host: ${QDRANT_HOST:-localhost}
      port: ${QDRANT_GRPC_PORT:-6334}
      vector_dimensions: 1536
      similarity_metric: Cosine
      timeout: 30
```

Usage:

```bash
./bin/weave --config configs/config.qdrant-local.yaml cols ls
```

#### Cloud Configuration

Create or use `configs/config.qdrant-cloud.yaml`:

```yaml
databases:
  default: qdrant
  vector_databases:
    - name: qdrant
      type: qdrant-cloud
      url: ${QDRANT_URL}
      api_key: ${QDRANT_API_KEY}
      vector_dimensions: 1536
      similarity_metric: Cosine
      timeout: 30
```

Usage:

```bash
./bin/weave --config configs/config.qdrant-cloud.yaml cols ls
```

### Option 3: Use CLI Flags

```bash
# Local
./bin/weave --qdrant-local <command>

# Cloud
./bin/weave --qdrant-cloud <command>
```

### Similarity Metrics

Qdrant supports the following distance metrics:

- **Cosine**: Cosine similarity (default, good for normalized vectors)
- **Dot**: Dot product (for unnormalized vectors)
- **Euclidean**: Euclidean distance (L2 distance)

---

## Usage Examples

### Collection Management

```bash
# List all collections
./bin/weave --qdrant-local cols ls

# Create a collection
./bin/weave --qdrant-local cols create \
  --name my_docs \
  --vector-dimensions 1536 \
  --similarity-metric Cosine

# Delete a collection
./bin/weave --qdrant-local cols delete --name my_docs

# Get collection schema
./bin/weave --qdrant-local cols schema --name my_docs
```

### Document Operations

```bash
# Create a document (auto-generates embedding)
./bin/weave --qdrant-local docs create \
  --collection my_docs \
  --content "This is a test document about machine learning." \
  --metadata '{"category": "AI", "date": "2025-11-28"}'

# Get a document
./bin/weave --qdrant-local docs get \
  --collection my_docs \
  --id <document-id>

# Update a document
./bin/weave --qdrant-local docs update \
  --collection my_docs \
  --id <document-id> \
  --content "Updated content about deep learning."

# List documents
./bin/weave --qdrant-local docs ls \
  --collection my_docs \
  --limit 10

# Delete a document
./bin/weave --qdrant-local docs delete \
  --collection my_docs \
  --id <document-id>
```

### Search Operations

```bash
# Semantic search
./bin/weave --qdrant-local search semantic \
  --collection my_docs \
  --query "machine learning algorithms" \
  --top-k 5

# Metadata search
./bin/weave --qdrant-local search metadata \
  --collection my_docs \
  --filters '{"category": "AI"}' \
  --limit 10
```

---

## Troubleshooting

### Connection Issues

#### Local: "connection refused"

```bash
# Check if Qdrant is running
docker ps | grep qdrant

# Check logs
docker logs qdrant

# Restart Qdrant
docker restart qdrant
```

#### Cloud: "authentication failed"

```bash
# Verify credentials
echo $QDRANT_URL
echo $QDRANT_API_KEY

# Make sure URL includes port 6334 (gRPC)
# Correct: https://xyz.cloud.qdrant.io:6334
# Wrong: https://xyz.cloud.qdrant.io
```

### Port Conflicts

If port 6333 or 6334 is already in use:

```bash
# Find process using the port
lsof -i :6334

# Stop Qdrant
docker stop qdrant

# Remove container
docker rm qdrant

# Start with different ports
docker run -d \
  --name qdrant \
  -p 7333:6333 \
  -p 7334:6334 \
  -v $(pwd)/qdrant_storage:/qdrant/storage:z \
  qdrant/qdrant

# Update environment variable
export QDRANT_GRPC_PORT="7334"
```

### Embedding Generation Errors

```bash
# Verify OpenAI API key
echo $OPENAI_API_KEY

# Test OpenAI connection
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer $OPENAI_API_KEY"
```

### Performance Issues

- **Slow queries**: Ensure collection is indexed (Qdrant indexes automatically)
- **Memory usage**: Consider using quantization for large datasets
- **Network latency**: For cloud, ensure good network connectivity

---

## Known Limitations

### Current Status: Experimental

- ✅ Core functionality implemented and tested
- 🧪 Real-world production validation in progress
- 🧪 Performance benchmarking pending

### Feature Limitations

- **No BM25 Search**: Qdrant does not support keyword-based BM25 search natively
- **Hybrid Search Fallback**: Hybrid search falls back to vector-only search
- **Float32 Only**: Embeddings are automatically converted from float64 to float32

### Integration Testing

- Integration tests are written but need validation against:
  - Live local Qdrant instance
  - Qdrant Cloud cluster

---

## Next Steps

1. **Test with Real Data**: Run the integration tests with live Qdrant instances
2. **Production Validation**: Test at scale with production workloads
3. **Performance Benchmarking**: Compare with other vector databases
4. **Advanced Features**: Explore quantization and multi-vector support

---

## Resources

- [Qdrant Documentation](https://qdrant.tech/documentation/)
- [Qdrant Cloud](https://cloud.qdrant.io)
- [Qdrant GitHub](https://github.com/qdrant/qdrant)
- [weave-cli VDB Support Matrix](../VDB_SUPPORT.md)

---

## Feedback

Since Qdrant integration is experimental, we welcome feedback:

- Found a bug? [Open an issue](https://github.com/maximilien/weave-cli/issues)
- Have a suggestion? [Start a discussion](https://github.com/maximilien/weave-cli/discussions)
- Production use case? Share your experience!

---

**Last Updated**: 2025-11-28
**Version**: v0.7.0 (Experimental)
