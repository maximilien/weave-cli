# Getting Started with Weave Stack

**Quick Start Guide for Production RAG Deployment**

---

## Prerequisites

Before you begin, ensure you have:

- **kubectl** - Kubernetes CLI
- **helm** - Package manager for Kubernetes
- **kind** - Local Kubernetes cluster (recommended for development)
- **OpenAI API key** - For embeddings (or use OSS template)
- **8GB+ RAM** available on your system

### Installation

**macOS**:

```bash
brew install kubectl helm kind
```

**Linux**:

- kubectl: https://kubernetes.io/docs/tasks/tools/
- helm: https://helm.sh/docs/intro/install/
- kind: https://kind.sigs.k8s.io/docs/user/quick-start/

---

## 5-Minute Quick Start

### Step 1: Install Weave CLI

```bash
# Clone repository
git clone https://github.com/maximilien/weave-cli
cd weave-cli

# Build
./build.sh

# Verify installation
./bin/weave --version
```

### Step 2: Initialize Your Stack

```bash
# Create new directory for your project
mkdir my-rag-project
cd my-rag-project

# Initialize with quickstart template
weave stack init --template quickstart --runtime kind
```

This creates:
- `weave-stack.yaml` - Stack configuration
- `kubernetes/` - Helm chart directory
- `.gitignore` - Git ignore rules

### Step 3: Deploy to Kubernetes

```bash
# Deploy the stack
weave stack up --runtime kind
```

This will:
1. Create a Kind cluster
2. Generate Helm charts
3. Deploy Milvus vector database
4. Wait for pods to be ready

**Expected output**:

```
✅ Cluster created: weave-stack
✅ Generated Helm chart: kubernetes/
✅ Helm chart installed successfully!
✅ All pods are ready!
🎉 Stack deployment initiated!
```

### Step 4: Check Status

```bash
# Verify deployment
weave stack status
```

### Step 5: Ingest Data

```bash
# Create sample data
mkdir -p data
echo "Weave is a production-ready RAG CLI tool" > data/sample.txt

# Set your OpenAI API key
export OPENAI_API_KEY="your-key-here"

# Ingest data
weave stack ingest Documents data/
```

### Step 6: Query Your Data

```bash
# Port-forward to Milvus
weave stack port-forward milvus 19530:19530 &

# Wait a moment for port-forward to establish
sleep 3

# Query
weave cols query Documents "RAG tool" --vector-db-type milvus-local
```

### Step 7: Clean Up

```bash
# Stop the stack
weave stack down
```

---

## Available Templates

### 1. Quickstart (Default)

Minimal stack for quick testing:

```bash
weave stack init --template quickstart --runtime kind
```

**Includes**:
- Milvus vector database
- 1 text collection
- OpenAI embeddings
- Local Kubernetes (Kind)

**Best for**: Testing, proof-of-concept

### 2. Production

Full-featured production stack:

```bash
weave stack init --template production --runtime kind
```

**Includes**:
- Everything in quickstart
- Ingestion pipeline with retry/checkpoint
- PM2 dashboard
- Health monitoring
- Enhanced error handling

**Best for**: Production deployments

### 3. Multimodal

Text + image processing:

```bash
weave stack init --template multimodal --runtime kind
```

**Includes**:
- Text and image collections
- PDF image extraction
- MinIO object storage
- OCR capabilities

**Best for**: Document processing with images

### 4. OSS (Open Source)

No API keys required:

```bash
weave stack init --template oss --runtime kind
```

**Includes**:
- Ollama for LLM
- Sentence-transformers for embeddings
- All open-source components

**Best for**: Privacy-first, no external APIs

---

## Common Commands

### Stack Management

```bash
# Initialize
weave stack init [--template <name>] [--runtime kind|minikube]

# Deploy
weave stack up --runtime kind

# Check status
weave stack status

# Stop
weave stack down

# Validate config
weave stack validate
```

### Data Operations

```bash
# Ingest data
weave stack ingest <collection-name> <data-path>

# With options
weave stack ingest Documents data/ \
  --chunk-size 1000 \
  --parallel-workers 4 \
  --batch-size 20
```

### Debugging

```bash
# View logs
weave stack logs milvus --tail 100 --follow

# Port forward
weave stack port-forward milvus 19530:19530

# Run kubectl commands
weave stack kubectl -- get pods
weave stack kubectl -- describe pod <pod-name>
```

---

## Troubleshooting

### Pods Not Starting

**Symptom**: Pods stuck in Pending or CrashLoopBackOff

**Check**:

```bash
kubectl --context kind-weave-stack get pods
kubectl --context kind-weave-stack describe pod <pod-name>
```

**Common fixes**:

1. **Insufficient memory**: Check node resources

   ```bash
   kubectl --context kind-weave-stack get nodes -o custom-columns=NAME:.metadata.name,MEMORY:.status.allocatable.memory
   ```

   Default template uses 2Gi memory. If needed, edit `weave-stack.yaml`:

   ```yaml
   infrastructure:
     vectordb:
       resources:
         requests:
           memory: "1Gi"  # Reduce if needed
   ```

2. **Image pull errors**: Check internet connection and image availability

### Port Forward Fails

**Symptom**: `error forwarding port`

**Fix**: Ensure pod is running first

```bash
weave stack status
# Wait for pods to be ready, then retry
```

### Ingestion Fails

**Symptom**: `no active stack found`

**Fix**: Deploy stack first

```bash
weave stack up --runtime kind
```

**Symptom**: `OPENAI_API_KEY not set`

**Fix**: Export your API key

```bash
export OPENAI_API_KEY="your-key-here"
```

### Cluster Already Exists

**Symptom**: `cluster 'weave-stack' already exists`

**Fix**: Delete existing cluster first

```bash
weave stack down
# Or manually:
kind delete cluster --name weave-stack
```

---

## Configuration

### Edit weave-stack.yaml

The stack configuration file controls all aspects of your deployment:

```yaml
version: "1.0"
name: my-rag-stack

runtime:
  kubernetes:
    provider: kind  # or minikube, eks, gke
    kind:
      name: weave-stack
      nodes: 1

  container_runtime: podman  # or docker

infrastructure:
  vectordb:
    type: milvus
    version: 2.3.0
    resources:
      requests:
        memory: "2Gi"  # Adjust based on your needs
        cpu: "1"

  llm:
    provider: openai  # or ollama
    models:
      embedding: text-embedding-3-small
      chat: gpt-4o

collections:
  - name: Documents
    type: text
    schema:
      vector_dimensions: 1536
```

### Environment Variables

Create `.env` file:

```bash
OPENAI_API_KEY=your-key-here
MILVUS_HOST=localhost
MILVUS_PORT=19530
```

---

## Next Steps

### 1. Customize Your Stack

Edit `weave-stack.yaml` to:
- Add more collections
- Adjust resource limits
- Configure different LLM providers
- Set up dashboards

### 2. Ingest Your Data

```bash
# Add your documents to data/
cp /path/to/your/docs/*.pdf data/

# Ingest
weave stack ingest YourCollection data/
```

### 3. Build Applications

Use the deployed Milvus instance in your applications:

```python
from pymilvus import connections, Collection

# Connect (via port-forward)
connections.connect(host="localhost", port="19530")

# Query
collection = Collection("Documents")
results = collection.search(...)
```

### 4. Deploy to Production

When ready for production:

```bash
# Use production template
weave stack init --template production --runtime kind

# Or for cloud (Phase 2 - coming soon)
weave stack init --template production --runtime eks  # AWS
weave stack init --template production --runtime gke  # GCP
```

---

## Support & Resources

### Documentation

- Main README: `README.md`
- Stack Guide: `docs/guides/WEAVE_STACK_QUICKSTART.md`
- Issue Tracker: https://github.com/maximilien/weave-cli/issues

### Getting Help

1. Check troubleshooting section above
2. Review logs: `weave stack logs milvus`
3. Check pod status: `weave stack status`
4. Open an issue on GitHub

### Version Information

```bash
weave --version
```

**Current version**: v0.10.2 (Production Ready)

---

## Known Limitations

1. **Local Development Only** (currently)
   - Kind and Minikube supported
   - EKS/GKE coming in Phase 2

2. **Single Node Clusters**
   - Default templates use 1 node
   - Multi-node support available (edit weave-stack.yaml)

3. **Memory Requirements**
   - Minimum 8GB RAM recommended
   - Milvus requires 2-4GB depending on data size

4. **Minikube Constraints**
   - Requires Docker Desktop >= 20.10.0
   - Podman machines may need memory adjustment
   - **Recommendation**: Use Kind for more reliable local development

---

## Quick Reference

```bash
# Complete workflow
weave stack init --template quickstart --runtime kind
weave stack up --runtime kind
export OPENAI_API_KEY="..."
weave stack ingest Documents data/
weave stack port-forward milvus 19530:19530 &
weave cols query Documents "search term"
weave stack down

# Debugging
weave stack status
weave stack logs milvus
weave stack kubectl -- get pods
weave stack kubectl -- describe pod <name>

# Cleanup
weave stack down
kind delete cluster --name weave-stack
```

---

**Happy RAG building!** 🚀
