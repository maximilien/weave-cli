---
marp: true
theme: default
class: lead
paginate: true
backgroundColor: #fff
backgroundImage: url('https://marp.app/assets/hero-background.svg')
---

# Weave CLI
## Universal Vector Database Management Made Simple

**A powerful CLI for managing 10 vector databases, Kubernetes RAG stacks, and agent evaluations**

**Maximilien.ai** | **v0.11.5+** | https://github.com/maximilien/weave-cli

---

# What is Weave CLI?

- 🤖 **AI-Powered** - Natural language queries with GPT-4o multi-agent system
- 🗄️ **10 Vector Databases** - Unified interface across all major VDBs
- ☸️ **Kubernetes Stack** - Deploy RAG stacks with one command
- 🧪 **Evaluations** - Test and benchmark RAG agents with datasets
- 💾 **Backup & Restore** - Cross-VDB migration with `.weavebak` format
- 🌟 **OSS Embeddings** - 100% free local alternatives to OpenAI
- 📦 **Batch Processing** - Parallel ingestion with checkpointing
- 📄 **PDF & Image Support** - Extract text and images with intelligent chunking
- 🔄 **Interactive REPL** - AI-powered interactive mode
- 🎭 **Mock Database** - Built-in testing and development support

---

# Supported Vector Databases (10!)

| Database | Status | Modes | Best For |
|----------|--------|-------|----------|
| **Weaviate** | Stable | cloud, local | Full-featured, recommended |
| **Qdrant** | Stable | cloud, local | Rust performance |
| **Milvus** | Stable | cloud, local | High-performance at scale |
| **Chroma** | Stable | cloud, local | Simple setup |
| **Supabase** | Stable | cloud, local | PostgreSQL + pgvector |
| **Neo4j** | Stable | cloud, local | Graph + vector |
| **MongoDB** | Stable | cloud | Atlas Vector Search |
| **Pinecone** | Beta | cloud | Serverless |
| **OpenSearch** | Stable | cloud, local | AWS, hybrid search |
| **Elasticsearch** | Beta | cloud, local | Elastic Cloud |

---

# Installation & Setup

```bash
# Clone and build
git clone https://github.com/maximilien/weave-cli.git
cd weave-cli
./setup.sh && ./build.sh

# Configure interactively
weave config create --env

# Test your setup
weave health check

# Start interactive REPL mode
weave
```

---

# Core Commands

```bash
# Health check
weave health check
weave health check --cloud

# Collections
weave cols ls
weave cols create DemoDocs --text --json-metadata
weave cols show DemoDocs --schema --expand-metadata
weave cols query DemoDocs "search term" --top_k 3

# Documents
weave docs create DemoDocs ./README.md
weave docs create DemoDocs ./docs/PRESENTATION.md
weave docs batch DemoDocs ./docs/
weave docs ls DemoDocs -w -S

# Statistics
weave stats DemoDocs
```

---

# Document Processing

## Text, Images, and PDFs

```bash
# Text documents
weave docs create DemoDocs ./README.md

# Image documents
weave docs create DemoImages ./tests/images/dog.png

# PDF with text + image extraction
weave docs create DemoDocs ~/Desktop/weave-cli.pdf \
  --image-col DemoImages

# Batch ingestion with progress tracking
weave pipeline ingest ./docs/ --collection DemoDocs

# Inspect before uploading
weave docs inspect ~/Desktop/weave-cli.pdf
```

---

# Weave Stack - Kubernetes RAG

## Deploy a complete RAG stack to Kubernetes in minutes

```bash
# Initialize with a template
weave stack init --template quickstart --runtime kind

# 4 templates available:
#   quickstart  - Minimal RAG stack
#   production  - Full stack with dashboard & evaluations
#   multimodal  - Text + image collections
#   oss         - 100% open source (Ollama + sentence-transformers)

# Validate configuration
weave stack validate

# Deploy!
weave stack up --runtime kind
```

---

# Stack Management

```bash
# Check status
weave stack status

# View logs
weave stack logs milvus --tail 20
weave stack logs milvus --follow

# kubectl passthrough (auto-injects context)
weave stack kubectl -- get pods
weave stack kubectl -- get services

# Port forwarding
weave stack port-forward milvus 19530:19530

# Ingest data into stack
weave stack ingest ClientDocs ./data/

# Backup stack collections
weave stack backup ClientDocs --output backup.weavebak

# Dashboard (production template)
weave stack dashboard start
weave stack dashboard status

# Tear down
weave stack down
```

---

# Backup & Restore

## Portable `.weavebak` format with cross-VDB migration

```bash
# Create backup (with compression)
weave backup create DemoDocs --output /tmp/demo.weavebak

# List backups
weave backup list /tmp/

# Validate integrity
weave backup validate /tmp/demo.weavebak

# Restore (even to a different VDB!)
weave backup restore /tmp/demo.weavebak
```

**Features:** Compression, complete data preservation (embeddings, metadata, images), cross-VDB migration, remote storage (S3, MinIO)

**Performance:** 2,636 docs in <2 minutes, ~50-60MB compressed

---

# Evaluations - Getting Started

## Test your RAG agents with structured datasets

```bash
# Explore available datasets
weave eval datasets list
weave eval datasets show baseline

# Run an evaluation
weave eval run --agent rag-agent --dataset baseline

# View results
weave eval list
weave eval show <RUN_ID>
```

**Built-in metrics:** Accuracy, Citation quality, Hallucination detection

**Built-in datasets:** baseline, simple-qa, medical-qa, technical-docs, multi-collection

---

# Evaluations - Custom Datasets & Evaluators

## Create domain-specific tests and scoring

```bash
# Create a dataset from template
weave eval datasets create my-qa --template simple-qa

# Create interactively
weave eval datasets create my-qa --interactive

# Create custom evaluators
weave eval create-evaluator technical_accuracy --type llm_judge
weave eval create-evaluator url_checker --type regex

# Validate
weave eval validate-evaluator evals/evaluators/technical_accuracy.yaml

# List all evaluators
weave eval list-evaluators
```

**Evaluator types:** `llm_judge`, `regex`, `exact_match`, `contains`

---

# Evaluations - Benchmarking & Observability

## Compare agents side-by-side with production monitoring

```bash
# Benchmark multiple agents on same dataset
weave eval benchmark \
  --agents rag-agent,qa-agent,summarize-agent \
  --dataset baseline

# Use Opik for production observability
weave eval run --agent rag-agent --dataset baseline --use-opik
```

**Opik integration provides:**
- Rich dashboard visualization
- Detailed trace of each evaluation
- Cost breakdown and tracking
- Historical trends
- Export to CSV/JSON

---

# Agent Management

```bash
# List available agents
weave agents list
weave agents show rag-agent

# Create from template
weave agents create my-agent

# Copy and customize
weave agents copy rag-agent my-custom-agent

# Validate configuration
weave agents validate evals/agents/my-agent.yaml
```

---

# OSS Embedding Providers

## 100% Free, Local Embeddings

| Provider | Type | Cost | Quality vs OpenAI |
|----------|------|------|-------------------|
| **OpenAI** | Cloud API | $0.02/1M tokens | Baseline (100%) |
| **sentence-transformers** | Local Python | **FREE** | 90-95% |
| **Ollama** | Local HTTP | **FREE** | 90-95% |

```bash
# Re-embed with OSS provider
weave cols re-embed MyCollection \
  --new-embedding sentence-transformers/all-mpnet-base-v2

# Compare quality
weave cols compare MyCollection MyCollection_OSS \
  --query "test query"
```

---

# AI-Powered REPL

```bash
# Enter interactive mode
weave

# Natural language queries
> list my collections
> list my empty collections
> how many documents in DemoDocs?
> search DemoDocs for "vector database"
> show me collection statistics
> create a backup of DemoDocs
```

**Features:** Multi-agent system, cost tracking, dry-run mode, Opik observability

---

# Configuration

```bash
weave config show                    # Current configuration
weave config list                    # All configured databases
weave config list --details          # With connection details
weave config list-schemas            # Configured schemas
weave config show-schema WeaveDocs   # Schema details
weave config show-schema WeaveDocs --yaml
```

## Priority (highest to lowest)
1. CLI flags (`--weaviate-cloud`, `--vdb milvus`)
2. `--env` file
3. `.env` file
4. Environment variables
5. `config.yaml`
6. Defaults

---

# Architecture Overview

```
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│  CLI / REPL  │  │  Eval Runner │  │  Stack Mgr   │
└──────┬───────┘  └──────┬───────┘  └──────┬───────┘
       │                 │                  │
       └────────┬────────┘──────────────────┘
                │
       ┌────────▼────────┐
       │   Unified VDB   │   10 database adapters
       │    Interface     │   3 embedding providers
       └────────┬────────┘   Backup/restore engine
                │
   ┌────────────┼────────────┐
   │   ┌────┐ ┌────┐ ┌────┐ │
   │   │Wvt │ │Qdrt│ │Mlvs│ │  ... +7 more
   │   └────┘ └────┘ └────┘ │
   └─────────────────────────┘
```

---

# Quick Reference

| Feature | Command |
|---------|---------|
| Health check | `weave health check` |
| Kubernetes stack | `weave stack init / up / down` |
| Collections | `weave cols ls / create / query` |
| Documents | `weave docs create / batch / ls` |
| Backup & restore | `weave backup create / restore` |
| Evaluations | `weave eval run --agent X --dataset Y` |
| Benchmarking | `weave eval benchmark --agents X,Y` |
| Custom evaluators | `weave eval create-evaluator NAME` |
| Agent management | `weave agents list / create` |
| OSS embeddings | `weave cols re-embed COL` |
| Statistics | `weave stats COL` |
| AI REPL | `weave` (no args) |

---

# Getting Started

## Resources
- 📖 **User Guide** - `docs/USER_GUIDE.md`
- 🎬 **Demo Scripts** - `demos/` directory (9 automated demos)
- 📋 **Demo Script** - `demos/DEMO_SCRIPT.md` (manual walkthrough)
- ☸️ **Stack Guide** - `docs/guides/WEAVE_STACK_QUICKSTART.md`
- 🧪 **Eval Datasets** - `evals/datasets/`

## Links
- 🐙 **GitHub**: https://github.com/maximilien/weave-cli
- 📋 **Issues**: https://github.com/maximilien/weave-cli/issues

---

# Thank You!

**Weave CLI** - Universal vector database management, Kubernetes RAG stacks, and agent evaluations in a single binary.

```bash
git clone https://github.com/maximilien/weave-cli.git
cd weave-cli && ./setup.sh && ./build.sh
weave health check
weave
```

**MIT License** | **Maximilien.ai**
