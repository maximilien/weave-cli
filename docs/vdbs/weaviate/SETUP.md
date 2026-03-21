# Weaviate Setup Guide

Weaviate is the default and recommended vector database for Weave CLI.
It's production-ready, fully tested, and supports both cloud and local
deployments.

## Status

✅ **Stable** - Production-ready, recommended for all use cases

## Prerequisites

- Weave CLI installed ([installation guide](../../README.md#installation))
- OpenAI API key for automatic embeddings (optional but recommended)

## Option 1: Weaviate Cloud (Recommended)

### 1. Create Weaviate Cloud Account

1. Go to [console.weaviate.cloud](https://console.weaviate.cloud)
2. Create a free account
3. Create a new cluster
4. Note your:
   - Cluster URL (e.g., `https://your-cluster.weaviate.network`)
   - API Key (Admin key from Details tab)

### 2. Configure Weave CLI

**Interactive Setup (Recommended)**:

```bash
# Configure only Weaviate Cloud variables (smart filtering)
weave config create --env --weaviate-cloud

# OR configure all databases
weave config create --env

# Follow prompts to enter:
# - WEAVIATE_URL: https://your-cluster.weaviate.network
# - WEAVIATE_API_KEY: your-admin-api-key
# - OPENAI_API_KEY: sk-...
```

**Manual Setup**:

```bash
# Add to .env file or export
export WEAVIATE_URL="https://your-cluster.weaviate.network"
export WEAVIATE_API_KEY="your-admin-api-key"
export OPENAI_API_KEY="sk-..."  # Optional but recommended
```

### 3. Verify Connection

```bash
weave health check --weaviate-cloud
```

## Option 2: Weaviate Local

### 1. Start Weaviate Locally

**Using Weave CLI Helper Script (Recommended)**:

```bash
# Start Weaviate (uses podman or docker automatically)
./tools/vdb/local/weaviate.sh start

# Check status
./tools/vdb/local/weaviate.sh status

# View logs
./tools/vdb/local/weaviate.sh logs

# Stop Weaviate
./tools/vdb/local/weaviate.sh stop
```

The script automatically:
- Uses OPENAI_API_KEY from .env if available
- Creates persistent storage in `weaviate_storage/`
- Configures text2vec-openai module
- Waits for health check to pass

**Using Docker (Manual)**:

```bash
docker run -d \
  -p 8080:8080 \
  -e AUTHENTICATION_ANONYMOUS_ACCESS_ENABLED=true \
  -e PERSISTENCE_DATA_PATH=/var/lib/weaviate \
  -e ENABLE_MODULES='text2vec-openai' \
  -e OPENAI_APIKEY=$OPENAI_API_KEY \
  semitechnologies/weaviate:latest
```

**Using Docker Compose**:

```yaml
# docker-compose.yml
version: '3.4'
services:
  weaviate:
    image: semitechnologies/weaviate:latest
    ports:
      - 8080:8080
    environment:
      AUTHENTICATION_ANONYMOUS_ACCESS_ENABLED: 'true'
      PERSISTENCE_DATA_PATH: '/var/lib/weaviate'
      QUERY_DEFAULTS_LIMIT: 25
      ENABLE_MODULES: 'text2vec-openai'
      OPENAI_APIKEY: ${OPENAI_API_KEY}
    volumes:
      - weaviate_data:/var/lib/weaviate

volumes:
  weaviate_data:
```

Start:

```bash
docker-compose up -d
```

### 2. Configure Weave CLI

**Interactive Setup (Recommended)**:

```bash
# Configure only Weaviate Local variables (smart filtering)
weave config create --env --weaviate-local

# Follow prompts to enter:
# - Leave WEAVIATE_URL as http://localhost:8080 (or press Enter)
# - OPENAI_API_KEY: sk-...
```

**Manual Setup**:

```bash
# Set environment variables
export VECTOR_DB_TYPE="weaviate-local"
export OPENAI_API_KEY="sk-..."  # Required for embeddings

# URL defaults to http://localhost:8080
```

**Using config.yaml** (already configured in example):

The weaviate-local section in config.yaml.example is already set up for local
development.

### 3. Verify Connection

```bash
weave health check --weaviate-local
```

## Usage Examples

### Create Collection

```bash
# Create a text collection
weave cols create MyDocs

# Create with specific schema
weave cols create MyDocs --schema my-schema.yaml
```

### Add Documents

```bash
# Add a single document
weave docs create MyDocs ./document.txt

# Add with metadata
weave docs create MyDocs ./document.txt --metadata '{"author":"John"}'

# Batch add
weave docs create MyDocs ./docs/*.pdf
```

### Search

```bash
# Vector search
weave cols query MyDocs "machine learning concepts"

# With filters
weave cols query MyDocs "AI" --limit 5 --where '{"author":"John"}'
```

### List and Manage

```bash
# List collections
weave cols ls --weaviate-cloud

# List documents
weave docs ls MyDocs

# Delete collection
weave cols delete MyDocs
```

## Advanced Configuration

### Custom Vectorizer

Edit `config.yaml`:

```yaml
databases:
  vectordatabases:
    - name: weaviate-cloud
      type: weaviate-cloud
      url: ${WEAVIATE_URL}
      apikey: ${WEAVIATE_API_KEY}
      vectorizer: text2vec-openai  # or text2vec-cohere, etc.
```

### Authentication

For production deployments, use API keys:

```bash
export WEAVIATE_API_KEY="your-admin-key"
```

## Troubleshooting

### Connection Issues

1. **Check URL format**: Must include `https://` or `http://`
2. **Verify API key**: Ensure it's the Admin key, not Read-only
3. **Test connectivity**: `curl $WEAVIATE_URL/v1/meta`

### Common Errors

**"connection refused"**:

- For local: Ensure Docker container is running
- For cloud: Check cluster status in console

**"authentication failed"**:

- Verify API key is correct
- Check key permissions in Weaviate console

**"module not enabled"**:

- Ensure ENABLE_MODULES includes required vectorizer
- Restart Weaviate after configuration changes

## OSS Embedding Providers (v0.9.19+)

Weaviate works with **all embedding providers** since embeddings are generated
client-side and passed as vectors.

### Quick Start

```bash
# List available providers
weave embeddings list

# OpenAI (default)
weave docs create MyCollection data/ --embedding text-embedding-3-small --weaviate

# sentence-transformers (OSS, FREE)
weave docs create MyCollection data/ \
  --embedding sentence-transformers/all-mpnet-base-v2 \
  --weaviate

# Ollama (local, FREE)
weave docs create MyCollection data/ --embedding nomic-embed-text --weaviate
```

### Cost Optimization Workflow

```bash
# 1. Quick start with OpenAI
weave docs create QuickTest data.pdf --embedding text-embedding-3-small --weaviate

# 2. Re-embed to OSS (20x faster than re-ingestion)
weave collection reembed QuickTest \
  --new-embedding sentence-transformers/all-mpnet-base-v2 \
  --output Production \
  --weaviate

# 3. Compare quality
weave collection compare QuickTest Production --query "test" --weaviate

# Result: 90%+ quality retention, $0 embedding costs
```

See [OSS Embedding Testing Tips](../guides/OSS_EMBEDDING_TESTING_TIPS.md) for
detailed setup and benchmarks.

## Resources

- [Weaviate Documentation](https://weaviate.io/developers/weaviate)
- [Weaviate Cloud Console](https://console.weaviate.cloud)
- [Weaviate Slack Community](https://weaviate.io/slack)
- [OSS Embeddings Guide](../../guides/OSS_EMBEDDING_TESTING_TIPS.md) - Free embedding providers

## Next Steps

- See [User Guide](../../USER_GUIDE.md) for more commands
- Check [VDB Support Matrix](../../VDB_SUPPORT_MATRIX.md) for feature comparison
- Review [Architecture](../../ARCHITECTURE.md) for system design
