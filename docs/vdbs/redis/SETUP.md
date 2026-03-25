# Redis Vector Database Setup Guide

## Overview

[Redis Stack](https://redis.io/docs/stack/) extends Redis with vector
search capabilities via the RediSearch module. It provides:

- **HNSW Indexing**: Fast approximate nearest neighbor search
- **Full-Text Search**: TF-IDF based text search on same index
- **Hybrid Search**: Combine vector KNN with text/metadata filters
- **Low Latency**: Sub-millisecond operations for cached data
- **Local & Cloud**: Docker deployment or Redis Cloud managed service

Weave CLI supports both **redis-local** (self-hosted Redis Stack) and
**redis-cloud** (Redis Cloud with vector search).

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Redis Local Setup](#redis-local-setup)
3. [Redis Cloud Setup](#redis-cloud-setup)
4. [Configuration](#configuration)
5. [Usage Examples](#usage-examples)
6. [Troubleshooting](#troubleshooting)
7. [Known Limitations](#known-limitations)

---

## Prerequisites

### Required

- **weave-cli** v0.11.6 or later
- **OpenAI API Key** (for automatic embedding generation)

### For Local Development

- **Docker** or **Podman**

---

## Redis Local Setup

### 1. Start Redis Stack

```bash
# Redis Stack includes RediSearch (vector search module)
docker run -d \
  --name redis-stack \
  -p 6379:6379 \
  redis/redis-stack:latest
```

Verify it's running:

```bash
docker exec redis-stack redis-cli PING
# Should return: PONG

# Verify RediSearch module is loaded
docker exec redis-stack redis-cli MODULE LIST
# Should include "search" module
```

### 2. Configure weave-cli

Add to your `config.yaml`:

```yaml
databases:
  default: redis-local
  vector_databases:
    - name: redis-local
      type: redis-local
      url: redis://localhost:6379
      vector_dimensions: 1536
      similarity_metric: COSINE
      enabled: true
```

Or set environment variables:

```bash
export REDIS_URL=redis://localhost:6379
```

### 3. Verify Connection

```bash
weave health check --redis-local
weave doctor --section vdb
```

---

## Redis Cloud Setup

### 1. Create a Redis Cloud Account

1. Go to [Redis Cloud](https://redis.com/try-free/)
2. Create a database with the **Search and Query** module enabled
3. Note your endpoint and password

### 2. Configure weave-cli

```yaml
databases:
  vector_databases:
    - name: redis-cloud
      type: redis-cloud
      url: redis-12345.c1.us-east-1-2.ec2.cloud.redislabs.com:12345
      api_key: your-redis-cloud-password
      vector_dimensions: 1536
      similarity_metric: COSINE
      enabled: true
```

---

## Configuration

### Config Fields

| Field | Required | Description |
| --- | --- | --- |
| `name` | Yes | Display name for this database |
| `type` | Yes | `redis-local` or `redis-cloud` |
| `url` | Yes | Redis connection URL or host:port |
| `api_key` | Cloud only | Redis Cloud password |
| `vector_dimensions` | Recommended | Embedding dimensions (default: 1536) |
| `similarity_metric` | No | `COSINE`, `L2`, or `IP` (default: COSINE) |
| `enabled` | No | Enable/disable this database |

### Similarity Metrics

| Metric | Description | Best For |
| --- | --- | --- |
| `COSINE` | Cosine similarity (default) | Most text embeddings |
| `L2` | Euclidean distance | Spatial data |
| `IP` | Inner product | Normalized vectors |

---

## Usage Examples

```bash
# Create a collection
weave cols create MyDocs --redis-local

# Ingest documents
weave docs create MyDocs data/document.pdf --redis-local

# Search
weave cols query MyDocs "search query" --redis-local

# List collections
weave cols ls --redis-local

# Health check
weave health check redis-local
```

---

## Troubleshooting

### "RediSearch module not loaded"

You're using plain Redis instead of Redis Stack. Use the Redis
Stack image:

```bash
docker run -d -p 6379:6379 redis/redis-stack:latest
```

### "Connection refused"

Check that Redis is running and the port is correct:

```bash
docker ps | grep redis
redis-cli -h localhost -p 6379 PING
```

### "WRONGTYPE Operation against a key holding wrong kind of value"

This can happen if non-weave keys conflict with the index prefix.
Weave uses the `weave:` key prefix to avoid conflicts.

---

## Known Limitations

1. **Schema updates**: RediSearch indexes are immutable after
   creation. To change fields, delete and recreate the collection.
2. **BM25**: Redis uses TF-IDF scoring (similar but not identical
   to BM25). Results are comparable for most use cases.
3. **Large metadata**: Redis HASH fields have practical size
   limits. Use external storage for large binary data.
4. **No native collections**: Collections are simulated using
   RediSearch indexes with key prefixes (`weave:{collection}:`).
