# Milvus Local Setup Guide

This guide covers setting up Milvus for local development and testing.

## Prerequisites

You need one of the following container runtimes:

- **Podman** (recommended): Better security, rootless operation
- **Docker**: Widely available alternative

### Installing Podman (Recommended)

**macOS:**
```bash
brew install podman
podman machine init
podman machine start
```

**Linux:**
```bash
# Debian/Ubuntu
sudo apt-get install podman

# Fedora/RHEL
sudo dnf install podman

# Arch Linux
sudo pacman -S podman
```

**Windows:**
```bash
# Install via Chocolatey
choco install podman

# Or download from GitHub releases
# https://github.com/containers/podman/releases
```

### Installing Docker (Alternative)

**macOS:**
```bash
brew install --cask docker
```

**Linux:**
```bash
# Follow instructions at https://docs.docker.com/engine/install/
```

**Windows:**
```bash
# Download Docker Desktop from https://www.docker.com/products/docker-desktop
```

## Quick Start

### 1. Start Milvus

```bash
# Start Milvus (auto-detects podman/docker)
./tools/vdb/local/milvus.sh start

# This will:
# - Detect your container runtime (podman preferred)
# - Use appropriate compose file
# - Start 3 containers: milvus-standalone, etcd, minio
# - Wait for Milvus to be ready
```

### 2. Verify Installation

```bash
# Check status
./tools/vdb/local/milvus.sh status

# Test connection
./tools/vdb/health.sh milvus

# View logs
./tools/vdb/local/milvus.sh logs
```

### 3. Use with Weave

```bash
# Set OpenAI API key for embeddings
export OPENAI_API_KEY="your-openai-api-key"

# Create a collection
weave --milvus-local collections create test_collection

# Add documents
weave --milvus-local documents create test_collection \
  --text "Milvus is a vector database" \
  --metadata '{"source":"test"}'

# Search
weave --milvus-local search vector test_collection \
  --query "vector database" \
  --limit 5
```

## Detailed Setup

### Directory Structure

```
weave-cli/
├── local/
│   └── milvus/
│       ├── docker-compose.yml       # Docker compose file
│       ├── podman-compose.yml       # Podman compose file
│       └── README.md                # Local setup notes
└── tools/
    └── vdb/
        ├── container/
        │   ├── detect.sh            # Runtime detection
        │   └── run.sh               # Unified interface
        ├── local/
        │   ├── milvus.sh            # Milvus management
        │   └── manager.sh           # Multi-VDB manager
        └── health.sh                # Health checks
```

### Container Components

The local Milvus setup consists of 3 containers:

#### 1. milvus-etcd
- **Purpose**: Metadata storage and service discovery
- **Image**: quay.io/coreos/etcd:v3.5.5
- **Ports**: 2379 (internal)
- **Volume**: milvus_etcd

#### 2. milvus-minio
- **Purpose**: Object storage for vector data
- **Image**: minio/minio:RELEASE.2023-03-20T20-16-18Z
- **Ports**: 9000 (internal), 9001 (console)
- **Volume**: milvus_minio
- **Credentials**: minioadmin/minioadmin

#### 3. milvus-standalone
- **Purpose**: Main Milvus service
- **Image**: milvusdb/milvus:v2.4.2
- **Ports**: 19530 (gRPC), 9091 (metrics)
- **Volume**: milvus_data
- **Depends on**: etcd, minio

### Container Runtime Differences

#### Podman
```yaml
volumes:
  - ./volumes/milvus:/var/lib/milvus:Z  # :Z for SELinux labeling
  - ./volumes/etcd:/etcd:Z
  - ./volumes/minio:/minio_data:Z
```

#### Docker
```yaml
volumes:
  - ./volumes/milvus:/var/lib/milvus
  - ./volumes/etcd:/etcd
  - ./volumes/minio:/minio_data
```

## Management Scripts

### milvus.sh Commands

```bash
# Start Milvus
./tools/vdb/local/milvus.sh start

# Stop Milvus
./tools/vdb/local/milvus.sh stop

# Restart Milvus
./tools/vdb/local/milvus.sh restart

# Check status
./tools/vdb/local/milvus.sh status

# View logs (all services)
./tools/vdb/local/milvus.sh logs

# View logs (specific service)
./tools/vdb/local/milvus.sh logs milvus-standalone

# Clean up (removes volumes and data)
./tools/vdb/local/milvus.sh clean
```

### manager.sh Commands

For managing multiple vector databases:

```bash
# Start Milvus
./tools/vdb/local/manager.sh start milvus

# Stop Milvus
./tools/vdb/local/manager.sh stop milvus

# List all available VDBs
./tools/vdb/local/manager.sh list

# Start all VDBs
./tools/vdb/local/manager.sh start-all

# Stop all VDBs
./tools/vdb/local/manager.sh stop-all
```

## Configuration

### Default Configuration

The local setup uses these defaults:

```yaml
address: localhost:19530
database: default
timeout: 10
vector_dimensions: 1536
similarity_metric: L2
```

### Custom Configuration

Create `config.yaml`:

```yaml
vectordb:
  type: milvus-local
  address: localhost:19530
  database: my_database
  timeout: 30
  vector_dimensions: 1536
  similarity_metric: COSINE  # or L2, IP
```

Use with weave:

```bash
weave --config config.yaml --milvus-local collections list
```

### Environment Variables

Set connection parameters via environment:

```bash
export MILVUS_LOCAL_ADDRESS="localhost:19530"
export MILVUS_LOCAL_DATABASE="default"
export OPENAI_API_KEY="sk-..."  # For automatic embeddings
```

## Data Persistence

### Volume Management

Milvus data is stored in named volumes:

```bash
# List volumes (podman)
podman volume ls | grep milvus

# List volumes (docker)
docker volume ls | grep milvus

# Inspect volume
podman volume inspect milvus_data

# Remove volumes (WARNING: deletes all data)
podman volume rm milvus_data milvus_etcd milvus_minio
```

### Backup and Restore

**Backup:**
```bash
# Stop Milvus first
./tools/vdb/local/milvus.sh stop

# Create backup directory
mkdir -p backups/milvus

# Copy volumes (adjust paths for your system)
cp -r volumes/milvus backups/milvus/data
cp -r volumes/etcd backups/milvus/etcd
cp -r volumes/minio backups/milvus/minio
```

**Restore:**
```bash
# Stop Milvus
./tools/vdb/local/milvus.sh stop

# Restore volumes
cp -r backups/milvus/data volumes/milvus
cp -r backups/milvus/etcd volumes/etcd
cp -r backups/milvus/minio volumes/minio

# Start Milvus
./tools/vdb/local/milvus.sh start
```

## Troubleshooting

### Connection Refused

**Problem:** Cannot connect to Milvus

**Solutions:**
```bash
# Check if containers are running
podman ps | grep milvus
# or
docker ps | grep milvus

# Check logs for errors
./tools/vdb/local/milvus.sh logs

# Restart Milvus
./tools/vdb/local/milvus.sh restart

# Clean start (removes data)
./tools/vdb/local/milvus.sh clean
./tools/vdb/local/milvus.sh start
```

### Port Conflicts

**Problem:** Port 19530 already in use

**Solutions:**

Edit `local/milvus/podman-compose.yml` or `docker-compose.yml`:

```yaml
services:
  standalone:
    ports:
      - "19531:19530"  # Use different external port
```

Update config:
```yaml
vectordb:
  address: localhost:19531
```

### Permission Denied (Podman)

**Problem:** Permission errors on Linux

**Solutions:**

1. **Use rootless mode:**
```bash
podman machine init --rootful=false
podman machine start
```

2. **Fix volume permissions:**
```bash
chmod -R 755 volumes/
```

3. **Add SELinux labels:**
```bash
# Volumes in compose already have :Z flag
# If mounting manually:
podman run -v ./data:/data:Z ...
```

### Container Won't Start

**Problem:** Containers fail to start

**Solutions:**

```bash
# Check logs
./tools/vdb/local/milvus.sh logs

# Common issues:

# 1. Insufficient memory
# - Increase container memory limits

# 2. etcd or minio not ready
# - Wait for dependencies to start
# - Check logs: ./tools/vdb/local/milvus.sh logs milvus-etcd

# 3. Corrupted volumes
# - Clean and restart: ./tools/vdb/local/milvus.sh clean && ./tools/vdb/local/milvus.sh start
```

### Slow Performance

**Problem:** Queries are slow

**Solutions:**

1. **Ensure collection is loaded:**
```bash
# Collection must be loaded into memory for fast searches
# This happens automatically on collection creation
```

2. **Check system resources:**
```bash
# Podman
podman stats

# Docker
docker stats

# Increase memory if needed (edit compose file)
```

3. **Optimize index:**
```bash
# Default IVF_FLAT with nlist=128 is a good start
# For larger datasets, consider tuning nlist = sqrt(N)
```

## Advanced Configuration

### Custom Index Types

Modify `src/pkg/vectordb/milvus/collection.go` to use different indexes:

```go
// HNSW for high recall
index, err := entity.NewIndexHNSW(entity.L2, 16, 200)

// IVF_SQ8 for memory efficiency
index, err := entity.NewIndexIvfSQ8(entity.L2, 128)

// IVF_PQ for large datasets
index, err := entity.NewIndexIvfPQ(entity.L2, 128, 16, 8)
```

### Multiple Databases

```bash
# Create database
podman exec -it milvus-standalone \
  milvus-cli create database --database my_db

# Use in config
vectordb:
  database: my_db
```

### Resource Limits

Edit compose file to set resource limits:

```yaml
services:
  standalone:
    deploy:
      resources:
        limits:
          cpus: '4'
          memory: 8G
        reservations:
          cpus: '2'
          memory: 4G
```

## Testing

### Integration Tests

```bash
# Run local tests
go test ./src/pkg/vectordb/milvus/... -v

# Run with local Milvus
export MILVUS_LOCAL_ADDRESS="localhost:19530"
go test ./tests/integration/milvus/... -v
```

### Manual Testing

```bash
# Create test collection
weave --milvus-local collections create test_$(date +%s)

# Add test documents
for i in {1..100}; do
  weave --milvus-local documents create test_collection \
    --text "Test document $i" \
    --metadata "{\"id\":$i}"
done

# Test search
weave --milvus-local search vector test_collection \
  --query "test" \
  --limit 10

# Clean up
weave --milvus-local collections delete test_collection
```

## OSS Embedding Providers (v0.9.19+)

Milvus works with **all embedding providers** since embeddings are generated
client-side and passed as vectors.

### Supported Providers

```bash
# List available providers
weave embeddings list

# OpenAI (default)
weave docs create MyCollection data/ --embedding text-embedding-3-small

# sentence-transformers (OSS, FREE)
weave docs create MyCollection data/ \
  --embedding sentence-transformers/all-mpnet-base-v2

# Ollama (local, FREE)
weave docs create MyCollection data/ --embedding nomic-embed-text
```

### Cost Optimization

Re-embed existing collections with OSS models to eliminate embedding costs:

```bash
# Phase 1: Quick start with OpenAI
weave docs create QuickTest data.pdf --embedding text-embedding-3-small

# Phase 2: Re-embed to OSS for production (20x faster than re-ingestion)
weave collection reembed QuickTest \
  --new-embedding sentence-transformers/all-mpnet-base-v2 \
  --output Production

# Phase 3: Compare quality
weave collection compare QuickTest Production --query "test queries"

# Result: 90%+ quality, $0 embedding costs
```

### Performance Notes

- **sentence-transformers**: 90-95% quality vs OpenAI, FREE, Python required
- **Ollama**: 90-95% quality vs OpenAI, FREE, great for local LLMs
- **Re-embedding**: 200+ docs/min (vs 10 docs/min full re-ingestion)

See [OSS Embedding Testing Tips](../guides/OSS_EMBEDDING_TESTING_TIPS.md)
for detailed setup and testing guide.

## Next Steps

- [Cloud Setup Guide](CLOUD_SETUP.md) - Deploy to Zilliz Cloud
- [Milvus Documentation](https://milvus.io/docs) - Official Milvus docs
- [Performance Tuning](https://milvus.io/docs/performance_faq.md) - Optimize for production
- [OSS Embeddings Guide](../guides/OSS_EMBEDDING_TESTING_TIPS.md) - Free embedding providers
