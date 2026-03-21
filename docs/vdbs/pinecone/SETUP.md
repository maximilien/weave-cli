# Pinecone Setup Guide

## Overview

[Pinecone](https://www.pinecone.io/) is a fully managed, serverless vector database designed for high-performance similarity search at scale. It provides:

- **Serverless Architecture**: Auto-scaling with no infrastructure management
- **Fast Vector Search**: Optimized for low-latency similarity queries
- **Metadata Filtering**: Filter search results by metadata
- **Cloud-Only**: No local deployment (managed service only)
- **Automatic Embedding**: Integrated with OpenAI for text embedding generation

This guide covers **Pinecone Cloud** setup (managed serverless service).

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Pinecone Cloud Setup](#pinecone-cloud-setup)
3. [Configuration](#configuration)
4. [Usage Examples](#usage-examples)
5. [Troubleshooting](#troubleshooting)
6. [Known Limitations](#known-limitations)

---

## Prerequisites

### Required

- **weave-cli** v0.7.6 or later
- **Pinecone Account** (free tier available at [pinecone.io](https://www.pinecone.io/))
- **OpenAI API Key** (for automatic embedding generation)

---

## Pinecone Cloud Setup

### 1. Create Pinecone Account

1. Go to [https://app.pinecone.io/](https://app.pinecone.io/)
2. Sign up for a free account (no credit card required for starter plan)
3. Verify your email address

### 2. Get API Key

1. Log in to Pinecone Console
2. Navigate to **API Keys** in the left sidebar
3. Click **Create API Key**
4. Copy the API key (starts with `pcsk_...`)
5. **Important**: Save this key securely - it won't be shown again

### 3. Set Environment Variables

```bash
# Required for Pinecone authentication
export PINECONE_API_KEY="pcsk_..."

# Required for automatic embeddings
export OPENAI_API_KEY="sk-..."
```

### 4. Verify Setup with weave-cli

```bash
# Interactive configuration (easiest)
weave config create --env
# Select: pinecone
# Enter your PINECONE_API_KEY and OPENAI_API_KEY

# Verify connection
weave health check --pinecone

# Expected output:
# ✓ Pinecone: Connected
```

---

## Configuration

### Environment Variables

| Variable | Required | Description | Example |
|----------|----------|-------------|---------|
| `PINECONE_API_KEY` | Yes | API key from Pinecone Console | `pcsk_...` |
| `OPENAI_API_KEY` | Yes | OpenAI API key for embeddings | `sk-...` |
| `PINECONE_HOST` | No | Custom host (optional, for non-serverless) | `https://...` |

### Configuration Methods

#### Method 1: Interactive Setup (Recommended)

```bash
weave config create --env
```

Follow the prompts to enter your keys.

#### Method 2: Manual `.env` File

Create or edit `.env` in your project root:

```bash
VECTOR_DB_TYPE=pinecone
PINECONE_API_KEY=pcsk_your_api_key_here
OPENAI_API_KEY=sk_your_openai_key_here
```

#### Method 3: Export Environment Variables

```bash
export VECTOR_DB_TYPE=pinecone
export PINECONE_API_KEY="pcsk_..."
export OPENAI_API_KEY="sk-..."
```

---

## Usage Examples

### Create an Index (Collection)

```bash
# Create a serverless index for text embeddings (1536 dimensions)
weave collection create MyDocuments --text

# Pinecone will automatically create a serverless index
# with appropriate settings (cosine similarity, 1536 dims)
```

**Note**: Pinecone automatically configures indexes based on your first upsert operation.

### Add Documents

```bash
# Add a single document
weave document create MyDocuments document.txt

# Add multiple documents
weave document create MyDocuments doc1.txt doc2.txt doc3.txt

# Documents are automatically embedded using OpenAI
```

### Search

```bash
# Semantic search
weave query semantic MyDocuments "find documents about machine learning"

# Metadata filtering
weave query metadata MyDocuments --filter "type=tutorial"
```

### List Indexes

```bash
# List all Pinecone indexes
weave collection list --pinecone
```

### Delete an Index

```bash
# Delete an index (CAUTION: Cannot be undone)
weave collection delete MyDocuments --pinecone
```

---

## Troubleshooting

### Connection Issues

**Problem**: `PINECONE_API_KEY is required` error

**Solution**:
```bash
# Verify environment variable is set
echo $PINECONE_API_KEY

# If empty, set it:
export PINECONE_API_KEY="pcsk_..."

# Or use config create
weave config create --env
```

**Problem**: `failed to create Pinecone client: invalid API key`

**Solutions**:
1. Verify API key is correct (copy-paste from Pinecone Console)
2. Check for extra spaces or newlines in the key
3. Regenerate API key in Pinecone Console if needed

### Embedding Issues

**Problem**: `Warning: Failed to create OpenAI client for embeddings`

**Solution**:
```bash
# Set OpenAI API key
export OPENAI_API_KEY="sk-..."

# Verify it's set
echo $OPENAI_API_KEY
```

### Rate Limiting

**Problem**: `429 Too Many Requests` errors

**Solutions**:
1. Free tier has rate limits - wait a few seconds between requests
2. Upgrade to paid plan for higher limits
3. Use batch operations instead of individual upserts

### Index Not Found

**Problem**: `Index not found` error

**Solutions**:
1. Verify index name is correct (case-sensitive)
2. Check index exists: `weave collection list --pinecone`
3. Create the index: `weave collection create IndexName --text`

---

## Known Limitations

### No Local Deployment

- **Pinecone is cloud-only** - No local/self-hosted option
- All data is stored in Pinecone's cloud infrastructure
- Requires internet connectivity for all operations

**Alternatives for Local Development**:
- Use **Weaviate** (supports local Docker deployment)
- Use **Qdrant** (supports local Docker deployment)
- Use **Milvus** (supports local Docker deployment)

### No Native BM25 Search

- Pinecone does not support BM25 full-text search
- Only vector (semantic) search is available
- For BM25 + vector hybrid search, consider:
  - **Elasticsearch** (native kNN + BM25)
  - **OpenSearch** (native kNN + BM25)
  - **Weaviate** (native BM25 + vector)

**Workaround**: See [BM25_ALTERNATIVES.md](BM25_ALTERNATIVES.md) for sparse-dense hybrid approaches.

### Serverless-Only Architecture

- Pinecone uses serverless indexes (auto-scaling)
- No control over infrastructure or pod configuration
- Cold start latency possible for inactive indexes (< 1 second)

### Dimension Limits

- Maximum vector dimensions: 20,000
- Recommended: 1536 (OpenAI text-embedding-3-small) or 3072 (text-embedding-3-large)
- Starter plan has storage limits (check current pricing)

### Metadata Limitations

- Metadata values are limited in size
- Complex nested metadata may require flattening
- Metadata filtering performance depends on cardinality

---

## Best Practices

1. **Use Batch Operations**: Insert documents in batches for better performance
2. **Monitor Costs**: Track usage in Pinecone Console (free tier is generous but limited)
3. **Index Naming**: Use descriptive names (Pinecone indexes are globally namespaced in your account)
4. **Dimension Consistency**: Always use the same embedding dimension within an index
5. **Metadata Design**: Keep metadata simple and flat for optimal filtering performance

---

## Additional Resources

- **Pinecone Documentation**: [https://docs.pinecone.io/](https://docs.pinecone.io/)
- **Pinecone Console**: [https://app.pinecone.io/](https://app.pinecone.io/)
- **Pricing**: [https://www.pinecone.io/pricing/](https://www.pinecone.io/pricing/)
- **Weave CLI Docs**: [../../README.md](../../README.md)
- **BM25 Alternatives**: [BM25_ALTERNATIVES.md](BM25_ALTERNATIVES.md)

---

## Getting Help

- **Weave CLI Issues**: [github.com/maximilien/weave-cli/issues](https://github.com/maximilien/weave-cli/issues)
- **Pinecone Support**: [https://support.pinecone.io/](https://support.pinecone.io/)
- **Community**: Pinecone Slack community (link in Pinecone Console)
