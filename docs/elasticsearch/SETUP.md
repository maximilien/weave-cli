# Elasticsearch Setup Guide

Complete setup instructions for using Elasticsearch with Weave CLI.

## Prerequisites

- Elasticsearch 8.x+ (local or Elastic Cloud)
- OpenAI API key for embeddings
- Docker (for local setup) or Elastic Cloud account

## Choose Your Setup

Select the deployment type that matches your needs:

### Local Elasticsearch

Best for: Development, testing, learning

- ✅ **Free and open source**
- ✅ **Full control over configuration**
- ✅ **No cloud costs**
- ❌ Requires local resources (2GB+ RAM recommended)
- ❌ Manual maintenance and updates

**Setup Guide**: [LOCAL_SETUP.md](LOCAL_SETUP.md)

### Elastic Cloud

Best for: Production, managed service, auto-scaling

- ✅ **Managed service** - No infrastructure management
- ✅ **Auto-scaling** - Handles load automatically
- ✅ **Built-in monitoring** - Kibana included
- ✅ **Global deployment** - Multiple regions
- ❌ Costs money (free trial available)

**Setup Guide**: [CLOUD_SETUP.md](CLOUD_SETUP.md)

## Configuration Options

All configuration is done through YAML files in the `configs/` directory.

### Common Settings

```yaml
databases:
  default: elasticsearch
  vector_databases:
    - name: elasticsearch
      type: elasticsearch-local  # or elasticsearch-cloud

      # Vector configuration
      vector_dimensions: 1536      # Must match embedding model
      similarity_metric: cosine    # cosine, dot_product, or l2_norm

      # Timeout (seconds)
      timeout: 30  # local: 30s, cloud: 60s
```

### Similarity Metrics

Choose based on your embedding model:

| Metric | When to Use | Config Value |
|--------|-------------|--------------|
| **Cosine** | Normalized embeddings (OpenAI) | `cosine` |
| **Dot Product** | Pre-normalized vectors | `dot_product` |
| **L2 Norm** | Euclidean distance | `l2_norm` |

**Recommendation**: Use `cosine` for OpenAI embeddings.

### Vector Dimensions

Must match your embedding model:

| Model | Dimensions | Config |
|-------|------------|--------|
| text-embedding-3-small | 1536 | `vector_dimensions: 1536` |
| text-embedding-3-large | 3072 | `vector_dimensions: 3072` |
| text-embedding-ada-002 | 1536 | `vector_dimensions: 1536` |

## Environment Variables

### Local Elasticsearch

```bash
# Required
export ELASTICSEARCH_LOCAL_ADDRESS="http://localhost:9200"
export OPENAI_API_KEY="sk-..."

# Optional (if security enabled)
export ELASTICSEARCH_LOCAL_USERNAME="elastic"
export ELASTICSEARCH_LOCAL_PASSWORD="your-password"
```

### Elastic Cloud

```bash
# Required
export ELASTICSEARCH_CLOUD_ID="deployment:dXMtZWFzdC0xLmF3cy5mb3VuZC5pbyQxMjM0NTY="
export OPENAI_API_KEY="sk-..."

# Authentication (choose one method)
# Method 1: API Key (recommended)
export ELASTICSEARCH_CLOUD_API_KEY="your-base64-api-key"

# Method 2: Basic Auth (alternative)
export ELASTICSEARCH_CLOUD_USERNAME="elastic"
export ELASTICSEARCH_CLOUD_PASSWORD="your-password"
```

## Configuration Creation

Use the interactive config creator:

```bash
weave config create --env
```

This will:
1. Prompt for your Elasticsearch deployment type
2. Ask for connection details
3. Request OpenAI API key
4. Create the config file with proper formatting
5. Validate the configuration

## Verification

After setup, verify your configuration:

```bash
# Check Elasticsearch health
weave health check

# Should show:
# ✓ Elasticsearch connection successful
# ✓ Cluster status: green (or yellow for single-node)
# ✓ OpenAI API key valid
```

## Next Steps

Once configured:

1. **Create a collection**:
   ```bash
   weave cols create MyDocs --text
   ```

2. **Add documents**:
   ```bash
   weave docs create MyDocs document.txt
   ```

3. **Search**:
   ```bash
   weave cols q MyDocs "your search query"
   ```

## Troubleshooting

### Connection Failed

**Local**: Check Elasticsearch is running:
```bash
curl http://localhost:9200
```

**Cloud**: Verify Cloud ID and API key are correct.

### Authentication Errors

Check environment variables are set:
```bash
echo $ELASTICSEARCH_CLOUD_ID
echo $ELASTICSEARCH_CLOUD_API_KEY
```

### Timeout Errors

Increase timeout in config:
```yaml
timeout: 60  # or higher for large operations
```

### Memory Issues

Elasticsearch requires adequate heap size:
```bash
# Docker: Set memory limit
docker run -e "ES_JAVA_OPTS=-Xms2g -Xmx2g" ...
```

## Resources

- [LOCAL_SETUP.md](LOCAL_SETUP.md) - Local Elasticsearch setup
- [CLOUD_SETUP.md](CLOUD_SETUP.md) - Elastic Cloud setup
- [README.md](README.md) - Elasticsearch integration overview
- [Elasticsearch Docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/index.html)
