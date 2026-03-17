# Weave CLI Demo Script (Manual)

> **Duration**: ~20-25 minutes | **Version**: v0.11.5+
>
> Step-by-step script for live demos. Run each command manually.
> For automated demos, see the shell scripts in this directory.

---

## Prerequisites

```bash
# Build weave CLI
./build.sh

# Verify it works
weave -h

# Ensure environment is configured
export OPENAI_API_KEY="your-key"

# Suppress config validation warnings during demo
# Add to your config.yaml (or use --quiet-config flag):
#   quiet-config: true
```

---

## Part 1: Core CLI Overview (~3 min)

### Help and Discovery

```bash
weave -h
```

### Health Check

```bash
weave health check
weave health check --cloud
```

### Configuration

```bash
weave config -h
weave config show
weave config list
weave config list --details
weave config list-schemas
weave config show-schema WeaveDocs
weave config show-schema WeaveDocs --yaml
```

---

## Part 2: Collections and Documents (~5 min)

### Collections

```bash
weave cols -h
weave cols ls

# Create text collection
weave cols create DemoDocs \
  --text --json-metadata --weaviate-cloud

# Create image collection
weave cols create DemoImages \
  --image --json-metadata --weaviate-cloud

# Inspect collections
weave cols show DemoDocs \
  --schema --expand-metadata --weaviate-cloud
weave cols show DemoImages \
  --schema --expand-metadata --json --weaviate-cloud
```

### Documents -- Text

```bash
weave docs -h
weave docs ls DemoDocs

# Add text documents
weave docs create DemoDocs ./docs/PRESENTATION.md
weave docs create DemoDocs ./docs/ARCHITECTURE.md

weave docs ls DemoDocs
weave docs ls DemoDocs -w -S

# Inspect a document (use ID from ls output)
weave docs show DemoDocs <ID> \
  --schema --expand-metadata

# Delete a document
weave docs del DemoDocs --name "PRESENTATION.md"
```

### Documents -- Images and PDFs

```bash
weave docs ls DemoImages

# Add image document
weave docs create DemoImages ./tests/images/dog.png

# Add PDF with text + image extraction
weave docs create DemoDocs \
  ~/Desktop/weave-cli.pdf --image-col DemoImages

# Inspect a file before uploading
weave docs inspect ~/Desktop/weave-cli.pdf

weave docs ls DemoDocs -w -S
weave docs ls DemoImages
weave docs ls DemoImages -w -S

# Inspect image document
weave docs show DemoImages <ID> \
  --schema --expand-metadata

weave docs del DemoImages --name "dog.png"
```

### Query

```bash
weave cols query DemoDocs "golang"
weave cols query DemoDocs "RAGme.io" --top_k 3
```

### Batch Ingestion

```bash
# Batch add all files from a directory
weave docs batch DemoDocs ./docs/

# Pipeline ingestion with progress tracking
weave pipeline ingest ./docs/ --collection DemoDocs
```

### Statistics

```bash
weave stats DemoDocs
```

---

## Part 3: Backup and Restore (~3 min)

```bash
# Create a backup
weave backup create DemoDocs \
  --output /tmp/demo-backup.weavebak

# List backups
weave backup list /tmp/

# Validate backup integrity
weave backup validate /tmp/demo-backup.weavebak

# Restore to a new collection (cross-VDB migration!)
weave backup restore /tmp/demo-backup.weavebak
```

---

## Part 4: Weave Stack -- Kubernetes RAG (~5 min)

### Initialize

```bash
weave stack -h

# Initialize with quickstart template
weave stack init --template quickstart --runtime kind

# Other templates: production, multimodal, oss
# Other runtimes: minikube, eks, gke
```

### Validate and Deploy

```bash
# Validate configuration
weave stack validate

# Deploy to local Kind cluster
weave stack up --runtime kind
```

### Monitor and Manage

```bash
# Check status
weave stack status

# View logs
weave stack logs milvus --tail 20

# kubectl passthrough
weave stack kubectl -- get pods
weave stack kubectl -- get services

# Port forwarding
weave stack port-forward milvus 19530:19530
```

### Stack Data Operations

```bash
# Ingest data into stack
weave stack ingest ClientDocs ./data/

# Backup stack collection
weave stack backup ClientDocs \
  --output /tmp/stack-backup.weavebak

# View stack collections
weave stack collections
```

### Dashboard

```bash
# Start monitoring dashboard (production template)
weave stack dashboard start
weave stack dashboard status
weave stack dashboard logs
weave stack dashboard stop
```

### Cleanup

```bash
weave stack down
```

---

## Part 5: Evaluations -- Getting Started (~3 min)

> Run a basic eval to see how agent quality is measured.

### Explore Datasets

```bash
weave eval -h

# List available datasets
weave eval datasets list

# Show dataset details
weave eval datasets show baseline
```

### Run a Basic Evaluation

```bash
# Run eval against baseline dataset
weave eval run --agent rag-agent --dataset baseline

# View results
weave eval list
weave eval show <RUN_ID>
```

> **What just happened?** The eval runner loaded 5 test
> cases, ran the rag-agent on each, and scored accuracy,
> citation quality, and hallucination using LLM-as-judge
> evaluators.

---

## Part 6: Evaluations -- Custom Datasets (~3 min)

> Create your own test datasets and domain-specific
> evaluators.

### Create a Custom Dataset

```bash
# Create from template
weave eval datasets create my-qa --template simple-qa

# Or create interactively
weave eval datasets create my-qa --interactive

# Validate it
weave eval datasets validate \
  evals/datasets/my-qa.yaml

# Show what's inside
weave eval datasets show my-qa
```

### Create a Custom Evaluator

```bash
# List existing evaluators
weave eval list-evaluators

# Create a new LLM-judge evaluator
weave eval create-evaluator technical_accuracy \
  --type llm_judge

# Create a regex-based evaluator
weave eval create-evaluator url_checker --type regex

# Validate it
weave eval validate-evaluator \
  evals/evaluators/technical_accuracy.yaml
```

### Run with Custom Evaluators

```bash
# Datasets can reference custom evaluators in YAML:
#   custom_evaluators:
#     - technical_accuracy
#
# Then run as normal:
weave eval run \
  --agent rag-agent --dataset custom-eval-demo
```

---

## Part 7: Evaluations -- Benchmarking (~3 min)

> Compare multiple agents side-by-side and use production
> observability.

### Benchmark Multiple Agents

```bash
# Compare agents on the same dataset
weave eval benchmark \
  --agents rag-agent,qa-agent \
  --dataset baseline

# Save results
weave eval benchmark \
  --agents rag-agent,qa-agent,summarize-agent \
  --dataset baseline \
  --output benchmark-results.json
```

### Use Opik for Production Observability

```bash
# Set up Opik
export OPIK_API_KEY="your-opik-key"
export OPIK_WORKSPACE="your-workspace"
export OPIK_PROJECT_NAME="weave-cli-evals"

# Run eval with Opik evaluators + dashboard
weave eval run \
  --agent rag-agent --dataset baseline --use-opik

# View results in Opik dashboard:
#   - Detailed trace of each evaluation
#   - Cost breakdown
#   - Historical trends
#   - Export to CSV/JSON
```

> **Opik benefits**: Better LLM-as-judge evaluators,
> rich dashboard visualization, cost tracking, and
> production monitoring.

---

## Part 8: Agents (~2 min)

### Agent Management

```bash
weave agents list
weave agents show rag-agent

# Create from template
weave agents create my-agent

# Validate configuration
weave agents validate evals/agents/my-agent.yaml

# Copy and customize
weave agents copy rag-agent my-custom-agent
```

---

## Part 9: AI Agent Mode (REPL) (~2 min)

```bash
# Enter interactive REPL
weave

# Natural language queries:
> list my collections
> list my empty collections
> how many documents in DemoDocs?
> search DemoDocs for "vector database"
> show me collection statistics
```

---

## Part 10: Cleanup

```bash
# Delete documents
weave docs delete-all DemoDocs
weave docs delete-all DemoImages

# Delete collection schemas
weave cols delete-schema DemoDocs
weave cols delete-schema DemoImages
```

---

## Quick Reference: Feature Highlights

| Feature | Command |
| ------- | ------- |
| 10 vector databases | `weave config list` |
| Kubernetes stack | `weave stack init/up/down` |
| Batch ingestion | `weave docs batch COL DIR` |
| Backup and restore | `weave backup create/restore` |
| Cross-VDB migration | `weave backup restore FILE` |
| Evaluations | `weave eval run --agent X --dataset Y` |
| Benchmarking | `weave eval benchmark --agents X,Y` |
| Custom evaluators | `weave eval create-evaluator NAME` |
| OSS embeddings | `weave cols re-embed COL` |
| Embedding comparison | `weave cols compare COL` |
| Agent management | `weave agents list/show/create` |
| AI REPL | `weave` (no args) |
| Statistics | `weave stats COL` |
| Schema suggestions | `weave schema suggest ./data/` |
| Chunking analysis | `weave chunking suggest ./data/` |
| MCP tools | `weave mcp list/call/test` |

---

## Supported Vector Databases

| Database | Status | Modes |
| -------- | ------ | ----- |
| Weaviate | Stable | cloud, local |
| Qdrant | Stable | cloud, local |
| Milvus | Stable | cloud, local |
| Chroma | Stable | cloud, local |
| Supabase | Stable | cloud, local |
| Neo4j | Stable | cloud, local |
| MongoDB | Stable | cloud only |
| Pinecone | Beta | cloud only |
| OpenSearch | Stable | cloud, local |
| Elasticsearch | Beta | cloud, local |
