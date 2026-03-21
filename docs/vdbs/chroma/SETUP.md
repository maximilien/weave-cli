# Chroma Setup Guide

This guide covers setting up Chroma for use with weave-cli.

## Platform Requirements

⚠️ **Important**: Chroma support is **macOS only (AMD64/ARM64)**.

**Why?** The chroma-go v0.2.5 SDK has a CGO dependency (libtokenizers) that does
not support Linux or Windows build targets.

**Alternatives for Linux/Windows:**
- Weaviate (Cloud or Local)
- Milvus (Cloud or Local)
- Qdrant (Cloud or Local)
- MongoDB Atlas
- Supabase

If you are on macOS, continue with the setup below.

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

- Chroma Cloud account ([sign up](https://www.trychroma.com/))
- API key from Chroma Cloud dashboard
- Team UUID (tenant ID)
- Database name

### Getting Your Credentials

1. Sign in to [Chroma Cloud](https://www.trychroma.com/)
2. Go to your team settings
3. Create an API key (dev-test or production)
4. Note your **Team UUID** (tenant) and **Database name**

### Environment Configuration

Set the following environment variables:

```bash
# API Key - supports both variable names
export CHROMA_CLOUD_API_KEY=your-api-key
# or
export CHROMA_API_KEY=your-api-key

# Tenant and Database (required for cloud)
export CHROMA_TENANT=your-team-uuid
export CHROMA_DATABASE=your-database-name

# URL (optional - defaults to https://api.trychroma.com)
export CHROMA_CLOUD_URL=https://api.trychroma.com
```

Or add to your `.env` file:

```bash
CHROMA_CLOUD_API_KEY=ck-xxxxx...
CHROMA_TENANT=158d67be-ae7e-484d-8489-424fac7b2e61
CHROMA_DATABASE=weave-cli
CHROMA_CLOUD_URL=https://api.trychroma.com
```

### Test the Connection

```bash
./bin/weave health check --chroma-cloud
```

### Cloud vs Local Client Implementation

**Important:** weave-cli automatically selects the correct Chroma client based
on your configuration:

- **Local**: Uses `NewHTTPClient` from Chroma Go SDK
  - Connects to self-hosted Chroma server
  - Only requires `CHROMA_URL`

- **Cloud**: Uses `NewCloudClient` from Chroma Go SDK
  - Automatically connects to `https://api.trychroma.com`
  - Requires `CHROMA_API_KEY` (or `CHROMA_CLOUD_API_KEY`)
  - Requires `CHROMA_TENANT` and `CHROMA_DATABASE`
  - URL is optional (defaults to Chroma Cloud endpoint)

This automatic switching ensures proper authentication and connection handling
for each environment.

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

### 403 Permission Denied (Cloud Only)

If you get "403 Permission Denied" or "Forbidden" errors with Chroma Cloud:

**Root Cause:** This typically happens when using the wrong client type
(HTTPClient instead of CloudClient) or incorrect credentials.

**Solution:**

1. Verify all required environment variables are set:
   ```bash
   echo "API Key: ${CHROMA_CLOUD_API_KEY:0:20}..."
   echo "Tenant: $CHROMA_TENANT"
   echo "Database: $CHROMA_DATABASE"
   echo "URL: $CHROMA_CLOUD_URL"
   ```

2. Ensure you're using an API key (not database credentials):
   - API key format: `ck-xxxxx...`
   - Get from Chroma Cloud dashboard, not connection string

3. Verify tenant and database are NOT default values:
   - ❌ Don't use: `default_tenant` or `default_database`
   - ✅ Use: Your actual team UUID and database name from dashboard

4. Test API access directly:
   ```bash
   curl -H "Authorization: Bearer $CHROMA_CLOUD_API_KEY" \
     https://api.trychroma.com/api/v1/heartbeat
   ```

5. Regenerate API key if needed:
   - Old/expired keys may cause 403 errors
   - Create new key from Chroma Cloud dashboard

**Note:** weave-cli automatically uses `NewCloudClient` when `CHROMA_API_KEY`
or `CHROMA_CLOUD_API_KEY` is present, which ensures proper authentication.

### Quota Exceeded (Cloud Free Tier)

If you see "Quota exceeded" errors:

- **Free tier limit**: 300 documents per GET request
- **Workaround**: Use smaller page sizes or upgrade plan
- **Expected behavior**: Some integration tests fail on free tier

### Connection Refused (Local Only)

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

### API Version Mismatch (Local Only)

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
# or
./bin/weave cols ls --chroma-cloud
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
