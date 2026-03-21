# Milvus Setup Guide

Milvus is a high-performance vector database built for scale.
It supports both local development and cloud deployment (Zilliz).

## Status

🟢 **Beta** - Feature complete, tested, recommended for development and testing

## Quick Start

### Local Setup

See [LOCAL_SETUP.md](./LOCAL_SETUP.md) for detailed local installation.

Quick start:

```bash
# Using the helper script
./tools/vdb/local/milvus.sh start

# Or manually with docker
docker-compose -f tools/vdb/local/milvus-standalone.yml up -d

# Configure
export MILVUS_ADDRESS="localhost:19530"
export OPENAI_API_KEY="sk-..."

# Verify
weave health check --milvus-local
```

### Cloud Setup (Zilliz)

See [CLOUD_SETUP.md](./CLOUD_SETUP.md) for detailed cloud setup.

Quick start:

```bash
# Get credentials from Zilliz Cloud
export MILVUS_CLOUD_ADDRESS="your-cluster.cloud.zilliz.com:443"
export MILVUS_CLOUD_TOKEN="your-token"
export OPENAI_API_KEY="sk-..."

# Verify
weave health check --milvus-cloud
```

## Resources

- [Local Setup Guide](./LOCAL_SETUP.md)
- [Cloud Setup Guide](./CLOUD_SETUP.md)
- [Milvus Documentation](https://milvus.io/docs)
- [Zilliz Cloud](https://zilliz.com/cloud)
