# Chroma Setup Guide

This guide covers setting up Chroma for use with weave-cli.

## Local Development Setup

### Prerequisites

- Docker or Podman installed
- weave-cli built (`./build.sh`)

### Starting Chroma Server

The simplest way to run Chroma locally is with Docker/Podman:

```bash
# Using Podman (recommended)
podman run -d --name chromadb -p 8000:8000 chromadb/chroma:0.6.2

# Or using Docker
docker run -d --name chromadb -p 8000:8000 chromadb/chroma:0.6.2
```

> **Note**: We recommend Chroma v0.6.2 for compatibility with the Go SDK.

### Verify Chroma is Running

```bash
curl http://localhost:8000/api/v1/heartbeat
```

### Environment Configuration

Set the following environment variable:

```bash
export CHROMA_URL=http://localhost:8000
```

Or add to your `.env` file:

```bash
CHROMA_URL=http://localhost:8000
```

### Test the Connection

```bash
./bin/weave health check --chroma-local
```

## Cloud Setup (Chroma Cloud)

### Prerequisites

- Chroma Cloud account
- API key from Chroma Cloud dashboard

### Environment Configuration

Set the following environment variables:

```bash
export CHROMA_CLOUD_URL=https://your-instance.chroma.cloud
export CHROMA_CLOUD_API_KEY=your-api-key
```

Or add to your `.env` file:

```bash
CHROMA_CLOUD_URL=https://your-instance.chroma.cloud
CHROMA_CLOUD_API_KEY=your-api-key
```

### Test the Connection

```bash
./bin/weave health check --chroma-cloud
```

## Configuration File Setup

### Local Configuration

Add to your `config.yaml`:

```yaml
databases:
  default: chroma-local
  vector_databases:
    - name: chroma-local
      type: chroma-local
      url: ${CHROMA_URL}
      database: default_database
      vector_dimensions: 1536
      similarity_metric: cosine
      timeout: 30
```

### Cloud Configuration

```yaml
databases:
  default: chroma-cloud
  vector_databases:
    - name: chroma-cloud
      type: chroma-cloud
      url: ${CHROMA_CLOUD_URL}
      api_key: ${CHROMA_CLOUD_API_KEY}
      database: default_database
      vector_dimensions: 1536
      similarity_metric: cosine
      timeout: 60
```

## Basic Usage

### List Collections

```bash
./bin/weave cols ls --chroma-local
```

### Create a Collection

```bash
./bin/weave cols create my_collection --chroma-local
```

### Add Documents

```bash
./bin/weave docs create my_collection document.txt --chroma-local
```

### List Documents

```bash
./bin/weave docs ls my_collection --chroma-local
```

### Search Documents

```bash
./bin/weave search semantic my_collection "your query" --chroma-local
```

### Delete Collection

```bash
./bin/weave cols del my_collection --chroma-local
```

## Similarity Metrics

Chroma supports three similarity metrics:

- **cosine** (default) - Angular distance, good for text embeddings
- **l2** - Euclidean distance
- **ip** - Inner product

Configure in your config.yaml:

```yaml
similarity_metric: cosine  # or l2, ip
```

## Limitations

### No BM25 Search

Chroma does not support keyword-based BM25 search. Use semantic (vector) search instead:

```bash
# This will return an error
./bin/weave search bm25 my_collection "query" --chroma-local

# Use this instead
./bin/weave search semantic my_collection "query" --chroma-local
```

### Hybrid Search

Hybrid search in Chroma falls back to pure vector search (no keyword component).

## Troubleshooting

### Connection Refused

If you get "connection refused" errors:

1. Verify Chroma is running:
   ```bash
   podman ps | grep chromadb
   # or
   docker ps | grep chromadb
   ```

2. Check the URL is correct:
   ```bash
   echo $CHROMA_URL
   ```

3. Test with curl:
   ```bash
   curl http://localhost:8000/api/v1/heartbeat
   ```

### API Version Mismatch

If you see API version errors, ensure you're using Chroma v0.6.2 or later:

```bash
podman stop chromadb && podman rm chromadb
podman run -d --name chromadb -p 8000:8000 chromadb/chroma:0.6.2
```

> **Note**: weave-cli uses Chroma Go SDK v2 API which requires Chroma server
> v0.6.0+

### Collection Not Found

Collections are case-sensitive. Verify the exact name:

```bash
./bin/weave cols ls --chroma-local
```

## Running Integration Tests

```bash
# Set environment
export CHROMA_URL=http://localhost:8000

# Run Chroma tests only
./test.sh --chroma

# Run all tests (Chroma will be included)
./test.sh integration

# Skip Chroma tests
./test.sh integration --skip chroma
```

## Container Management

### Stop Chroma

```bash
podman stop chromadb
# or
docker stop chromadb
```

### Remove Container

```bash
podman rm chromadb
# or
docker rm chromadb
```

### View Logs

```bash
podman logs chromadb
# or
docker logs chromadb
```

### Restart Chroma

```bash
podman restart chromadb
# or
docker restart chromadb
```

## Data Persistence

By default, Chroma stores data in the container. For persistence:

```bash
podman run -d --name chromadb \
  -p 8000:8000 \
  -v chroma-data:/chroma/chroma \
  chromadb/chroma:0.6.2
```

## Related Documentation

- [VDB Support Matrix](../VDB_SUPPORT.md) - Feature comparison across databases
- [User Guide](../USER_GUIDE.md) - General weave-cli usage
