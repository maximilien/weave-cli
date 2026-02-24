# Weave Stack Quick Start Guide

**Version**: v0.10.0 (Phase 1)
**Last Updated**: 2025-02-24

This guide walks through deploying your first RAG stack with Weave Stack on local Kubernetes.

## What is Weave Stack?

Weave Stack orchestrates complete RAG (Retrieval-Augmented Generation) deployments including:
- **Vector Databases** (Milvus, Qdrant, Weaviate)
- **Data Ingestion Pipelines**
- **Web Dashboards** (Next.js/TypeScript)
- **Kubernetes Orchestration** (Kind, Minikube, EKS, GKE)

All managed via declarative YAML configuration and Helm charts.

## Prerequisites

### Install Required Tools

**macOS:**
```bash
brew install kubectl helm kind
```

**Linux:**
- kubectl: https://kubernetes.io/docs/tasks/tools/
- helm: https://helm.sh/docs/intro/install/
- kind: https://kind.sigs.k8s.io/docs/user/quick-start/

### Verify Installation

```bash
kubectl version --client
helm version
kind version
```

## 10-Minute Walkthrough

### 1. Initialize Stack

```bash
# Create new directory for your stack
mkdir my-rag-stack
cd my-rag-stack

# Initialize with quickstart template
weave stack init --template quickstart --runtime kind
```

**What this creates:**
- `weave-stack.yaml` - Stack configuration
- `kubernetes/` - Directory for Helm charts
- `.gitignore` - Git ignore rules

### 2. Review Configuration

```bash
cat weave-stack.yaml
```

**Example quickstart configuration:**
```yaml
---
version: "1.0"
name: my-rag-stack
description: Quickstart RAG stack

runtime:
  kubernetes:
    provider: kind
    kind:
      name: weave-stack
      nodes: 1
  container_runtime: podman

infrastructure:
  vectordb:
    type: milvus
    version: "2.3.0"
    resources:
      requests:
        memory: 8Gi
        cpu: "2"
      limits:
        memory: 12Gi
        cpu: "4"

  llm:
    provider: openai
    models:
      embedding: text-embedding-3-small
      chat: gpt-4o

collections:
  - name: Documents
    type: text
    description: Text documents
    schema:
      vector_dimensions: 1536  # OpenAI text-embedding-3-small
    sources:
      - pattern: "data/**/*.pdf"
        type: pdf
    chunking:
      strategy: semantic
      chunk_size: 500
    embedding:
      model: text-embedding-3-small
      provider: openai
```

### 3. Validate Configuration

```bash
weave stack validate
```

**Expected output:**
```
✅ Configuration valid

Summary:
  Name: my-rag-stack
  Runtime: kind
  Vector DB: milvus (2.3.0)
  Collections: 1
  Dashboard: disabled
```

### 4. Deploy to Kubernetes

```bash
weave stack up --runtime kind
```

**What happens:**
1. Creates Kind cluster (if needed)
2. Generates Helm charts from weave-stack.yaml
3. Deploys Milvus to Kubernetes
4. Waits for pods to be ready
5. Saves cluster info to `.weave-state/cluster.json`

**Expected output:**
```
📦 Creating Kind cluster: weave-stack
✅ Cluster created: kind-weave-stack

📝 Generating Helm charts...
✅ Charts generated: kubernetes/

🚀 Deploying stack...
NAME: my-rag-stack
NAMESPACE: default
STATUS: deployed
REVISION: 1

✅ Stack deployed successfully!

Next steps:
  1. Check status: weave stack status
  2. View logs: weave stack logs milvus --follow
  3. Port forward: weave stack port-forward milvus 19530:19530
```

**Deployment time:** ~2-3 minutes

### 5. Check Stack Status

```bash
weave stack status
```

**Example output:**
```
📊 Stack Status: my-rag-stack

Cluster:
  Provider: kind
  Context: kind-weave-stack
  Status: Running

Components:
  ✅ my-rag-stack-milvus-standalone  (Running)
     Pod: my-rag-stack-milvus-standalone-xyz
     Age: 2m30s
     Restarts: 0

Resources:
  CPU: 0.5/2.0 cores (25%)
  Memory: 4.2Gi/8.0Gi (52%)
```

### 6. View Logs

```bash
# Stream logs from Milvus
weave stack logs milvus --follow

# Show last 50 lines
weave stack logs milvus --tail 50

# Logs from all pods
weave stack logs --follow
```

### 7. Access Services

```bash
# Forward Milvus gRPC port (19530)
weave stack port-forward milvus 19530:19530
```

**In another terminal:**
```bash
# Now you can connect to Milvus at localhost:19530
weave vdb info milvus-local

# Or use kubectl directly
weave stack kubectl -- get pods
weave stack kubectl -- describe svc my-rag-stack-milvus
```

### 8. Stop Stack

```bash
# Stop and cleanup (deletes cluster)
weave stack down
```

**Expected output:**
```
ℹ️  Stopping weave stack...

📋 Uninstalling Helm release: my-rag-stack
✅ Helm release uninstalled

🗑️  Deleting Kind cluster: weave-stack
✅ Cluster deleted

🧹 Cleaning up state...
✅ Stack stopped successfully
```

## Stack Templates

### Quickstart (Default)

Minimal RAG stack for learning and testing.

```bash
weave stack init --template quickstart --runtime kind
```

**Includes:**
- Milvus vector database
- 1 text collection
- OpenAI embeddings

### Production

Full production stack with evaluations and dashboard.

```bash
weave stack init --template production --runtime kind
```

**Includes:**
- Milvus vector database
- Multiple collections
- Data ingestion pipeline
- Next.js dashboard
- Health monitoring
- Checkpointing

### Multimodal

Text + image RAG system.

```bash
weave stack init --template multimodal --runtime kind
```

**Includes:**
- Milvus vector database
- Text collection (PDFs)
- Image collection (PDF image extraction)
- MinIO for image storage
- OpenAI embeddings

### OSS (Open Source)

100% open source stack with no API keys required.

```bash
weave stack init --template oss --runtime kind
```

**Includes:**
- Milvus vector database
- Ollama for LLMs
- sentence-transformers for embeddings
- No external API dependencies

## Stack Management Commands

### Core Commands

```bash
# Initialize
weave stack init [--template <name>] [--runtime <provider>]

# Validate
weave stack validate

# Deploy
weave stack up --runtime <provider> [--timeout 5m]

# Status
weave stack status

# Logs
weave stack logs [service] [--follow] [--tail <lines>]

# Stop
weave stack down
```

### Kubectl Integration

```bash
# Auto-injects cluster context
weave stack kubectl -- <kubectl-args>

# Examples
weave stack kubectl -- get pods
weave stack kubectl -- describe pod my-rag-stack-milvus-xyz
weave stack kubectl -- get services
weave stack kubectl -- logs my-rag-stack-milvus-xyz --follow
```

### Port Forwarding

```bash
# Forward service port to localhost
weave stack port-forward <service> <local-port>:<remote-port>

# Examples
weave stack port-forward milvus 19530:19530
weave stack port-forward dashboard 3000:3000
```

**Service shortcuts:**
- `milvus`, `vectordb`, `vector-db` → Milvus service
- `dashboard`, `ui`, `web` → Dashboard service

### Dashboard Management (PM2)

If your stack includes a dashboard with PM2 runtime:

```bash
# Start dashboard
weave stack dashboard start

# Check status
weave stack dashboard status

# View logs
weave stack dashboard logs [--follow]

# Restart
weave stack dashboard restart

# Stop
weave stack dashboard stop
```

## Common Workflows

### Deploy and Query

```bash
# 1. Deploy stack
weave stack init
weave stack up --runtime kind

# 2. Wait for ready
weave stack status

# 3. Connect to Milvus
weave stack port-forward milvus 19530:19530

# 4. In another terminal, ingest data
weave docs create MyDocs data/documents.pdf --milvus-local

# 5. Query
weave cols query MyDocs "search query" --top-k 5
```

### Multi-Template Testing

```bash
# Test quickstart
mkdir quickstart-test && cd quickstart-test
weave stack init --template quickstart --runtime kind
weave stack up --runtime kind
weave stack status
weave stack down

# Test production
cd .. && mkdir production-test && cd production-test
weave stack init --template production --runtime kind
weave stack up --runtime kind
weave stack status
weave stack down
```

### Development Workflow

```bash
# 1. Initialize stack
weave stack init --template production

# 2. Customize weave-stack.yaml
vim weave-stack.yaml

# 3. Validate changes
weave stack validate

# 4. Deploy with longer timeout
weave stack up --runtime kind --timeout 10m

# 5. Monitor deployment
weave stack logs milvus --follow

# 6. Develop and test
# ... work on your application ...

# 7. Check health
weave stack status

# 8. Cleanup
weave stack down
```

## Troubleshooting

### Common Issues

**Cluster creation fails:**
```bash
# Check dependencies
kubectl version
helm version
kind version

# Check container runtime
podman machine start  # or docker
```

**Pods stuck in Pending:**
```bash
# Check node resources
weave stack kubectl -- get nodes
weave stack kubectl -- describe pod <pod-name>

# Check events
weave stack kubectl -- get events --sort-by='.lastTimestamp'
```

**Port forwarding fails:**
```bash
# Check service exists
weave stack kubectl -- get svc

# Check pod is running
weave stack status
```

**Helm install fails:**
```bash
# Check existing releases
helm list

# Uninstall if needed
helm uninstall my-rag-stack

# Or use stack down and retry
weave stack down
weave stack up --runtime kind
```

### Helpful Commands

```bash
# Get cluster info
cat .weave-state/cluster.json

# Check Helm release
helm list
helm status my-rag-stack

# Get pod details
weave stack kubectl -- get pods -o wide
weave stack kubectl -- describe pod <pod-name>

# Check resource usage
weave stack kubectl -- top nodes
weave stack kubectl -- top pods
```

## Next Steps

1. **Customize Configuration**: Edit `weave-stack.yaml` to add collections, change resources, enable dashboard
2. **Add Data Sources**: Configure data ingestion pipelines
3. **Deploy Dashboard**: Add Next.js dashboard for RAG interface
4. **Production Deployment**: Switch to EKS/GKE for cloud deployment (Phase 2)
5. **Monitoring**: Add Prometheus/Grafana for observability (Phase 3)

## Resources

- **[Weave Stack Design Doc](../planning/WEAVE_STACK_DESIGN.md)** - Architecture details
- **[Integration Tests](../../test-stack-basic.sh)** - Automated testing
- **[Main README](../../README.md#weave-stack---kubernetes-rag-management)** - Quick reference
- **[User Guide](../USER_GUIDE.md)** - Complete CLI documentation

## Support

For issues or questions:
- **GitHub Issues**: https://github.com/maximilien/weave-cli/issues
- **Discussions**: https://github.com/maximilien/weave-cli/discussions

## Version History

- **v0.10.0 (2025-02-24)**: Phase 1 release
  - Stack init/validate/up/down
  - Template system (quickstart, production, multimodal, oss)
  - kubectl/logs/port-forward integration
  - PM2 dashboard support
  - Error handling and dependency checks
