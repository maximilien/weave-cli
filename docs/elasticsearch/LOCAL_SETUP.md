# Elasticsearch Local Setup

Complete guide for running Elasticsearch locally with Weave CLI.

## Prerequisites

- Docker or Podman (recommended) OR
- Homebrew (macOS) OR
- Elasticsearch binary download
- 2GB+ RAM available
- OpenAI API key

## Quick Start (Docker)

**Easiest option** - Single command to get started:

```bash
# Start Elasticsearch (no security)
docker run -d \
  --name elasticsearch \
  -p 9200:9200 \
  -p 9300:9300 \
  -e "discovery.type=single-node" \
  -e "xpack.security.enabled=false" \
  -e "ES_JAVA_OPTS=-Xms2g -Xmx2g" \
  docker.elastic.co/elasticsearch/elasticsearch:8.11.0

# Wait for startup (30-60 seconds)
sleep 30

# Verify Elasticsearch is running
curl http://localhost:9200

# Expected output: {"name":"...", "cluster_name":"docker-cluster", ...}
```

## Setup Options

Choose the method that works best for you:

### Option 1: Docker (Recommended)

**Pros:** Easiest, consistent, isolated
**Cons:** Requires Docker

```bash
# 1. Start Elasticsearch
docker run -d \
  --name elasticsearch \
  -p 9200:9200 \
  -e "discovery.type=single-node" \
  -e "xpack.security.enabled=false" \
  -e "ES_JAVA_OPTS=-Xms2g -Xmx2g" \
  docker.elastic.co/elasticsearch/elasticsearch:8.11.0

# 2. Verify
curl http://localhost:9200

# 3. View logs
docker logs elasticsearch

# 4. Stop
docker stop elasticsearch

# 5. Start again
docker start elasticsearch

# 6. Remove completely
docker rm -f elasticsearch
```

### Option 2: Homebrew (macOS)

**Pros:** Native installation
**Cons:** System-wide, harder to clean up

```bash
# 1. Install
brew install elastic/tap/elasticsearch-full

# 2. Configure (disable security for dev)
echo "xpack.security.enabled: false" >> \
  /opt/homebrew/etc/elasticsearch/elasticsearch.yml

# 3. Start service
brew services start elastic/tap/elasticsearch-full

# 4. Verify
curl http://localhost:9200

# 5. Stop service
brew services stop elastic/tap/elasticsearch-full

# 6. Uninstall
brew uninstall elastic/tap/elasticsearch-full
```

### Option 3: Binary Download

**Pros:** No dependencies
**Cons:** Manual management

```bash
# 1. Download
curl -O https://artifacts.elastic.co/downloads/elasticsearch/elasticsearch-8.11.0-darwin-aarch64.tar.gz

# 2. Extract
tar -xzf elasticsearch-8.11.0-darwin-aarch64.tar.gz
cd elasticsearch-8.11.0

# 3. Configure (disable security)
echo "xpack.security.enabled: false" >> config/elasticsearch.yml

# 4. Start
./bin/elasticsearch

# 5. In another terminal, verify
curl http://localhost:9200
```

## Configure Weave CLI

After Elasticsearch is running:

### 1. Set Environment Variables

```bash
# Required
export ELASTICSEARCH_LOCAL_ADDRESS="http://localhost:9200"
export OPENAI_API_KEY="sk-..."

# Optional (if security enabled)
export ELASTICSEARCH_LOCAL_USERNAME="elastic"
export ELASTICSEARCH_LOCAL_PASSWORD="changeme"
```

### 2. Create Configuration

```bash
# Interactive config creation
weave config create --env

# Select: elasticsearch-local
# Address: http://localhost:9200
# OpenAI API key: sk-...
```

Or create manually at `configs/config.elasticsearch-local.yaml`:

```yaml
databases:
  default: elasticsearch
  vector_databases:
    - name: elasticsearch
      type: elasticsearch-local
      address: http://localhost:9200
      timeout: 30
      vector_dimensions: 1536
      similarity_metric: cosine
```

### 3. Verify Setup

```bash
# Test connection
weave health check

# Should show:
# ✓ Elasticsearch connection successful
# ✓ Cluster status: yellow (single node) or green
```

## First Operations

```bash
# Create a collection
weave cols create MyDocs --text

# Add a document
echo "Hello from Elasticsearch!" > test.txt
weave docs create MyDocs test.txt

# Search
weave cols q MyDocs "hello"

# List documents
weave docs ls MyDocs
```

## Troubleshooting

### Elasticsearch Won't Start

**Check if port is in use:**
```bash
lsof -i :9200
# Kill process if needed
```

**Check Docker memory:**
```bash
docker stats elasticsearch
# Ensure at least 2GB allocated
```

### Connection Refused

**Verify Elasticsearch is running:**
```bash
curl http://localhost:9200
```

**Check Docker container:**
```bash
docker ps | grep elasticsearch
docker logs elasticsearch
```

### Out of Memory Errors

**Increase heap size:**
```bash
docker run -e "ES_JAVA_OPTS=-Xms4g -Xmx4g" ...
```

**Or reduce if limited RAM:**
```bash
docker run -e "ES_JAVA_OPTS=-Xms1g -Xmx1g" ...
```

### Cluster Status Yellow

**Normal for single-node:** Yellow means replicas aren't assigned (expected with one node)

**Check:**
```bash
curl http://localhost:9200/_cluster/health?pretty
```

### Security Errors (403)

**If you enabled security:**
```bash
# Get the auto-generated password
docker logs elasticsearch | grep "Password for the elastic user"

# Or reset it
docker exec -it elasticsearch \
  bin/elasticsearch-reset-password -u elastic
```

## Performance Tuning

### For Development

```bash
docker run \
  -e "ES_JAVA_OPTS=-Xms1g -Xmx1g" \
  -e "bootstrap.memory_lock=false" \
  ...
```

### For Testing (Large Datasets)

```bash
docker run \
  -e "ES_JAVA_OPTS=-Xms4g -Xmx4g" \
  -e "thread_pool.write.queue_size=500" \
  ...
```

## Persistent Storage

### Save data between restarts

```bash
docker run \
  -v es-data:/usr/share/elasticsearch/data \
  ...

# Data persists even after:
docker rm -f elasticsearch
docker run -v es-data:/usr/share/elasticsearch/data ...
```

## Clean Up

### Remove all data

```bash
# Stop and remove container
docker rm -f elasticsearch

# Remove data volume
docker volume rm es-data
```

## Next Steps

- [SETUP.md](SETUP.md) - General configuration guide
- [README.md](README.md) - Elasticsearch features overview
- [CLOUD_SETUP.md](CLOUD_SETUP.md) - Elastic Cloud setup

## Resources

- [Elasticsearch Docker Docs](https://www.elastic.co/guide/en/elasticsearch/reference/current/docker.html)
- [Elasticsearch on Homebrew](https://www.elastic.co/guide/en/elasticsearch/reference/current/brew.html)
- [Performance Tuning](https://www.elastic.co/guide/en/elasticsearch/reference/current/tune-for-indexing-speed.html)
