# Weave CLI: A Unified CLI for RAG Across 11 Vector Databases

**TL;DR**: Weave CLI is a Go-based command-line tool that provides a unified interface for managing collections, documents, and RAG queries across 11 vector databases. It includes a multi-agent AI system, a Kubernetes deployment stack, pluggable evaluation with Opik, and full observability via OpenTelemetry. Built for RAG developers who want one tool instead of eleven.

---

## The Problem

Building RAG applications means picking a vector database and committing to its API, CLI, and quirks. But what if you need to compare databases? Or migrate from one to another? Or run the same pipeline across multiple providers?

Every VDB has its own:
- API style (GraphQL, gRPC, REST, SQL)
- Schema model (schemaless, typed, hybrid)
- Search capabilities (semantic only, BM25, hybrid)
- Deployment model (cloud, local, embedded)

You end up writing adapter code, custom scripts, and glue logic for each one. Testing across databases means maintaining parallel toolchains.

We built Weave CLI to solve this.

---

## What is Weave CLI?

Weave CLI provides a **single interface** that works across 11 vector databases:

| Database | Type | BM25 | Hybrid | Status |
|----------|------|------|--------|--------|
| Weaviate | Cloud + Local | Yes | Yes | Stable |
| Qdrant | Cloud + Local | No | HNSW | Stable |
| Milvus | Cloud + Local | Planned | No | Stable |
| Chroma | Cloud + Local | No | No | Stable |
| Supabase | Cloud + Local | Yes (tsquery) | Partial | Stable |
| MongoDB Atlas | Cloud | No | No | Stable |
| Neo4j | Cloud + Local | Lucene FTS | Graph+Vector | Stable |
| Pinecone | Cloud | No | Sparse-dense | Beta |
| Elasticsearch | Cloud + Local | Yes | Yes (RRF) | Beta |
| OpenSearch | Cloud + Local | Yes | Yes (RRF) | Stable |

The same commands work everywhere:

```bash
# Works with any VDB
weave cols create MyDocs --text --weaviate-local
weave cols create MyDocs --text --milvus-local
weave cols create MyDocs --text --qdrant-local

# Ingest documents
weave docs create MyDocs data/papers/ --embedding text-embedding-3-small

# Semantic search
weave cols query MyDocs "what is retrieval augmented generation?"

# Hybrid search (where supported)
weave query hybrid MyDocs "RAG architecture" --alpha 0.5
```

### How It Works: The Adapter Pattern

Under the hood, a `VectorDBClient` interface defines the contract:

```
VectorDBClient
├── CollectionOperations  (create, delete, list, count)
├── DocumentOperations    (create, get, update, delete, batch)
├── QueryOperations       (semantic, BM25, hybrid, metadata)
└── SchemaOperations      (get, update, validate)
```

Each database gets its own adapter that translates this interface into native API calls. A factory/registry provides dynamic client creation. You never touch provider-specific code.

---

## Weave Stack: One Command to Deploy Everything

For local development and production stacks, `weave stack` deploys your entire RAG infrastructure on Kubernetes:

```bash
# Deploy Milvus + MinIO + monitoring on Kind
weave stack up --runtime kind

# Check what's running
weave stack status

# Ingest documents directly into the stack
weave stack ingest Documents data/pdfs/ --embedding text-embedding-3-small

# Ingest images with OSS embeddings
weave stack ingest Images data/catalogs/ \
  --type image \
  --embedding sentence-transformers/all-mpnet-base-v2

# Ingest all collections from config
weave stack ingest --all

# Tear down when done
weave stack down
```

The stack handles:
- Kubernetes cluster provisioning (Kind, Minikube)
- Helm-based VDB deployment
- Port-forwarding and connection management
- Checkpointing and resume for large ingestion jobs
- Auto-restart on OOM with `--restart-every N`
- Multi-collection parallel ingestion with `--all`

---

## 10 Built-in Agents for AI-Powered Operations

Weave CLI includes a multi-agent system that turns natural language into executed database operations:

```bash
# Start the interactive REPL
weave

> show me all collections with more than 100 documents
> ingest the PDFs in data/ into my TechDocs collection
> compare search results between Weaviate and Qdrant for "machine learning"
```

Behind the scenes, 10 specialized agents collaborate through a 7-step orchestration pipeline:

1. **QueryAgent** -- validates and classifies intent
2. **PlanningAgent** -- creates a step-by-step execution plan
3. **User Confirm** -- shows the plan, waits for approval
4. **WeaveAgent** / **BashAgent** -- executes weave commands and shell operations
5. **RAGAgent** -- retrieval-augmented generation with citations
6. **ReportAgent** -- generates structured operation reports
7. **EvalAgent** -- tracks metrics and evaluates success

Additional agents handle chunking strategy (ChunkingAgent), schema analysis (SchemaAgent), and formatted output (OutputAgent).

Custom agents are supported via YAML configuration:

```bash
# Create a custom agent
weave agents create my-rag-agent --type rag --config agent.yaml

# List all agents
weave agents list

# Use a custom agent
weave eval run --agent my-rag-agent --dataset baseline
```

---

## Monitoring and Observability with Opik

Every agent operation, LLM call, and database query is traced via OpenTelemetry and exported to [Opik](https://www.comet.com/site/products/opik/) for full observability.

### What Gets Traced

The 7-step orchestration naturally produces rich traces with 5+ spans each:

- **Query analysis** -- input parsing, intent classification
- **Plan generation** -- LLM call for execution planning
- **Tool execution** -- each weave/bash command as a child span
- **RAG retrieval** -- VDB queries, context building, deduplication
- **LLM generation** -- prompt, completion, token counts
- **Evaluation** -- metric scoring for each response

### What You See in Opik

- **Trace waterfall** -- full request lifecycle with latency breakdown
- **Token usage** -- prompt, completion, and total tokens per operation
- **Cost tracking** -- per-trace and aggregate cost with color-coded display
- **Metadata** -- collection name, VDB type, embedding model, query parameters
- **Error correlation** -- failed spans linked to root cause

### How It Helped

Opik tracing was instrumental in diagnosing a critical production bug (Issue #57): documents appeared to ingest successfully but never persisted to Milvus. The trace waterfall showed that Flush operations were timing out silently, leading to the fix of using dedicated timeout contexts for the flush step.

Cost tracking also drove a key architectural decision: benchmarking showed that OSS embeddings (sentence-transformers) scored 11% higher on quality, ran 240x faster, and cost nothing compared to OpenAI -- data that was only visible through structured tracing.

### Setup

```bash
# Configure Opik
export OPIK_API_KEY=your-key
export OPIK_WORKSPACE=your-workspace
export OPIK_PROJECT_NAME=weave-cli

# Tracing is automatic -- every REPL query generates traces
weave
> search TechDocs for "kubernetes deployment patterns"
# -> Trace appears in Opik dashboard with all spans
```

---

## Evaluation with Opik

Weave CLI includes a pluggable evaluation harness with 4 LLM-judge evaluators, integrated with Opik for experiment tracking and comparison.

### Built-in Evaluators

| Evaluator | What It Measures |
|-----------|-----------------|
| **Accuracy** | Semantic correctness of the answer |
| **Faithfulness** | Is the answer grounded in retrieved context? |
| **Hallucination** | Does the answer contain unsupported claims? |
| **Context Relevance** | Are the retrieved documents relevant to the query? |

### Running Evaluations

```bash
# Run evaluation with Opik tracking
weave eval run --agent rag-agent --dataset baseline --use-opik

# Run against multiple datasets
weave eval run --agent rag-agent --dataset technical-docs --use-opik
weave eval run --agent rag-agent --dataset medical-qa --use-opik

# Compare across configurations
weave eval run --agent rag-agent --dataset baseline --use-opik \
  --vdb-type weaviate-local
weave eval run --agent rag-agent --dataset baseline --use-opik \
  --vdb-type qdrant-local
```

### What You See in Opik

- **Experiment dashboard** -- side-by-side comparison of runs
- **Per-sample scores** -- accuracy, faithfulness, hallucination, relevance for each query
- **Aggregate metrics** -- mean, median, distribution across datasets
- **Drift detection** -- how metrics change across experiments
- **Dataset versioning** -- track which dataset version produced which results

Datasets are YAML-based and live in `evals/datasets/`:

```yaml
# evals/datasets/baseline.yaml
name: baseline
description: Core RAG capability test
questions:
  - query: "What is vector similarity search?"
    expected: "Vector similarity search finds..."
    context_collection: TechDocs
```

The evaluation provider is pluggable -- you can run evaluations locally (no API needed) or with Opik for the full dashboard experience.

---

## Embedding Providers

Weave CLI supports 5 embedding providers with 15+ models:

| Provider | Models | Cost |
|----------|--------|------|
| **OpenAI** | text-embedding-3-small, text-embedding-3-large, ada-002 | Paid |
| **sentence-transformers** | all-mpnet-base-v2, all-MiniLM-L6-v2, + 4 more | Free |
| **Ollama** | nomic-embed-text, mxbai-embed-large, snowflake-arctic | Free (local) |
| **Cohere** | embed-english-v3.0, embed-multilingual-v3.0 | Paid |
| **Voyage AI** | voyage-2, voyage-large-2 | Paid |

OSS models are recommended as the default after benchmarking showed superior quality-to-cost ratios on production datasets.

---

## Architecture

```
CLI Layer (Cobra)
    |
    v
Executor -- 7-Step Orchestration
    |
    +-- Agent Layer (10 Built-in Agents)
    |       |
    |       +-- LLM Integration (OpenAI GPT-4o)
    |       +-- VectorDBClient Interface
    |               |
    |               +-- 11 VDB Adapters
    |
    +-- Ingestion Pipeline
    |       |
    |       +-- FileScanner (glob, exclude, SHA256 dedup)
    |       +-- Processor (PDF, text, JSON, YAML, image)
    |       +-- Batch Creator (100 docs/batch)
    |       +-- Embedding Provider (5 providers)
    |
    +-- Observability
            |
            +-- OpenTelemetry Tracing -> Opik Dashboard
            +-- Evaluation Harness (4 LLM Judges)
```

Key design decisions:
- **Go** -- single binary, cross-platform, fast startup, strong concurrency
- **Adapter pattern** -- each VDB adapter implements the same interface
- **Factory/registry** -- dynamic client creation from config
- **Pluggable providers** -- embeddings, evaluators, and LLM clients are all swappable

---

## Getting Started

```bash
# Clone and build
git clone https://github.com/maximilien/weave-cli.git
cd weave-cli
./build.sh

# Configure
./bin/weave config create --env

# Deploy a local stack
weave stack up --runtime kind

# Ingest documents
weave stack ingest Documents data/pdfs/

# Start querying
weave
> search Documents for "machine learning best practices"
```

**Resources:**
- [User Guide](https://github.com/maximilien/weave-cli/blob/main/docs/USER_GUIDE.md)
- [Architecture](https://github.com/maximilien/weave-cli/blob/main/docs/ARCHITECTURE.md)
- [VDB Support Matrix](https://github.com/maximilien/weave-cli/blob/main/docs/VDB_SUPPORT_MATRIX.md)
- [GitHub](https://github.com/maximilien/weave-cli)

---

**Built by [dr.max](https://github.com/maximilien) | MIT License | v0.11.5**
