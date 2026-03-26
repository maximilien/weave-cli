# Weave CLI Demo — AI Alliance (v0.12.0)

> **Duration**: ~45 minutes (leave 15 min for Q&A)
>
> **Automated script**: `demos/ai-alliance-demo.sh`
>
> Run each command manually for a live demo, or use the script
> with Enter-to-advance paging.

---

## Prerequisites

```bash
# Option 1: Homebrew (recommended for demos)
brew install Maximilien-ai/weave-cli/weave-cli

# Option 2: Build from source
./build.sh

# Verify
weave --version
```

Ensure your `config.yaml` and `.env` are set up with at least
`weaviate-cloud`, `milvus-local`, and `OPENAI_API_KEY`.

Start Milvus locally before the demo:

```bash
weave vdb local start milvus
# or: docker compose up -d milvus
```

---

## Part 1: Installation and Setup (~3 min)

**Goal**: Show how easy it is to get started.

```bash
# Homebrew install (one line!)
brew install Maximilien-ai/weave-cli/weave-cli

# Verify version
weave --version
```

**Talking points**:

- Single Go binary, no runtime dependencies
- 11 vector databases supported
- Open source, MIT licensed

---

## Part 2: weave doctor — Unified Diagnostics (~3 min)

**Goal**: Show the new diagnostic command (v0.12.0).

```bash
# Full scan
weave doctor

# Or by section
weave doctor --section system --verbose
weave doctor --section config
weave doctor --section vdb
```

**Talking points**:

- Checks system deps, config, env vars, VDB connectivity, LLM,
  embeddings, stack, and Opik
- Color-coded `[OK]`, `[WARN]`, `[FAIL]`, `[SKIP]` output
- `--json` for CI, `--fix` for remediation suggestions
- Replaces manual checks across `health check`, `config show`,
  env var inspection

---

## Part 3: Configuration and Health (~3 min)

**Goal**: Show multi-VDB configuration and health.

```bash
# List all configured databases
weave config list

# Health check (cloud databases)
weave health check --cloud

# VDB info
weave vdb list
```

**Talking points**:

- Single config.yaml for all 11 VDBs
- Environment variable interpolation (`${WEAVIATE_URL}`)
- Switch databases with `--weaviate-cloud`, `--milvus-local`, etc.
- Show milvus-local in config alongside weaviate-cloud

---

## Part 4: Collections and Documents (~10 min)

**Goal**: Create collections, ingest text/PDF/images, list docs.

### List existing collections

```bash
weave cols ls
```

### Create collections

```bash
# Text collection on milvus-local (show multi-VDB!)
weave cols create AllianceDemo \
    --embedding text-embedding-3-small --milvus-local

# Image collection on weaviate-cloud
weave cols create AllianceDemoImages \
    --image --weaviate-cloud
```

### Ingest text documents (Milvus)

Use milvus-local for text/markdown to show ingesting into a
different VDB than weaviate-cloud.

```bash
# Architecture doc from the repo itself
weave docs create AllianceDemo \
    docs/ARCHITECTURE.md --milvus-local

# README
weave docs create AllianceDemo \
    README.md --milvus-local

# List what we ingested
weave docs ls AllianceDemo --milvus-local
```

### Ingest a PDF (Weaviate — text + image extraction)

```bash
weave docs create AllianceDemo \
    tests/fixtures/ragme-io.pdf \
    --image-col AllianceDemoImages --weaviate-cloud
```

### Ingest images (Weaviate)

```bash
weave docs create AllianceDemoImages \
    tests/images/dog.png --weaviate-cloud

weave docs ls AllianceDemoImages --weaviate-cloud
```

**Talking points**:

- Automatic chunking for text
- PDF text + image extraction (tesseract OCR)
- Auto-generated OpenAI embeddings
- Images stored with base64 or external storage (S3/MinIO)

---

## Part 5: Semantic Search (~5 min)

**Goal**: Live queries showing RAG retrieval quality.

```bash
# Query text on milvus-local
weave cols query AllianceDemo \
    "how does the vector database abstraction work?" \
    --top-k 3 --milvus-local

weave cols query AllianceDemo \
    "what embedding models are supported?" \
    --top-k 3 --milvus-local

# Same query on weaviate-cloud (cross-VDB — same syntax!)
weave cols query AllianceDemo \
    "kubernetes deployment" \
    --top-k 2 --weaviate-cloud

# Image search on weaviate-cloud
weave cols query AllianceDemoImages \
    "dog" --top-k 2 --weaviate-cloud
```

**Talking points**:

- Vector similarity search using OpenAI embeddings
- Top-K results with relevance scores
- Same query syntax works across all 11 VDBs — just swap the flag
- Search text on Milvus and images on Weaviate in the same demo

---

## Part 6: Statistics and Schema (~3 min)

**Goal**: Inspect what's inside a collection.

```bash
# Stats on milvus-local
weave stats AllianceDemo --milvus-local

# Schema on weaviate-cloud
weave cols show AllianceDemo \
    --schema --weaviate-cloud
```

---

## Part 7: Backup and Restore (~5 min)

**Goal**: Portable backups, cross-VDB migration.

```bash
# Create a backup
weave backup create AllianceDemo \
    --output /tmp/alliance-demo.weavebak --weaviate-cloud

# List backups
weave backup list /tmp/

# Validate integrity
weave backup validate /tmp/alliance-demo.weavebak
```

**Talking points**:

- `.weavebak` format: gzip compressed, 65-95% size reduction
- Preserves embeddings, metadata, images
- Restore to any VDB (cross-VDB migration)
- S3/MinIO remote storage support

---

## Part 8: Batch Ingestion (~5 min)

**Goal**: Ingest a whole directory at once.

```bash
# Pipeline ingest
weave pipeline ingest docs/vdbs/redis/ \
    --collection AllianceDemo --weaviate-cloud

# Check updated count
weave cols count --weaviate-cloud
```

**Talking points**:

- Progress tracking, checkpointing, resume on failure
- Parallel workers, configurable batch size
- File type detection (PDF, text, markdown, images)

---

## Part 9: Multi-VDB Support (~3 min)

**Goal**: Show breadth of database support.

No commands — just talk through the table:

| VDB | Status | Modes |
| --- | --- | --- |
| Weaviate | Stable | cloud, local |
| Qdrant | Stable | cloud, local |
| Milvus | Stable | cloud, local |
| Chroma | Stable | cloud, local |
| Supabase | Stable | cloud, local |
| Neo4j | Stable | cloud, local |
| MongoDB | Stable | cloud |
| Pinecone | Beta | cloud |
| OpenSearch | Stable | cloud, local |
| Elasticsearch | Beta | cloud, local |
| Redis | Experimental | cloud, local |

**Talking points**:

- Unified `VectorDBClient` interface — same commands everywhere
- Adapter pattern: each VDB has its own adapter + factory
- Switch with a flag or set default in config

---

## Part 10: Embeddings (~3 min)

```bash
weave embeddings list
```

**Talking points**:

- OpenAI, sentence-transformers, Ollama, Cohere, Voyage AI
- OSS models: free, local, private — scored 11% higher than
  OpenAI on production dataset
- Auto-dimension detection from model registry

---

## Part 11: Agents (~5 min)

**Goal**: Show built-in agents creating a RAG solution.

```bash
# List available agents
weave agents list

# RAG agent — retrieval-augmented generation
weave agents run rag-agent \
    --collection AllianceDemo --milvus-local \
    "explain the vector database abstraction layer"

# Summarizer agent
weave agents run summarizer \
    --collection AllianceDemo --milvus-local \
    "summarize the architecture document"

# QA agent
weave agents run qa-agent \
    --collection AllianceDemo --milvus-local \
    "what embedding models are supported?"
```

**Talking points**:

- 10 built-in agents + custom YAML agents
- Each agent composes retrieval, prompting, and response
- Same `--milvus-local` / `--weaviate-cloud` flag works here too

---

## Part 12: Evaluations (~3 min)

**Goal**: Show evaluation harness — focus on eval quality.

```bash
# Run evaluation against a dataset
weave eval run --agent rag-agent --dataset baseline

# Benchmark multiple agents side-by-side
weave eval benchmark --agents rag,qa --dataset baseline
```

**Talking points**:

- LLM-as-judge evaluators (accuracy, faithfulness,
  hallucination, context relevance)
- Compare agents head-to-head on the same dataset
- Opik integration for production observability
- Traces, cost tracking, experiment comparison

---

## Part 13: Interactive REPL (~2 min)

```bash
# Start REPL (just run weave with no args)
weave
```

Example queries in the REPL:

```text
> list my collections
> how many documents in AllianceDemo?
> search AllianceDemo for vector database architecture
```

---

## Cleanup

```bash
# Milvus
weave docs da AllianceDemo --no-confirm --milvus-local
weave cols ds AllianceDemo --no-confirm --milvus-local

# Weaviate
weave docs da AllianceDemo --no-confirm --weaviate-cloud
weave docs da AllianceDemoImages --no-confirm --weaviate-cloud
weave cols ds AllianceDemo --no-confirm --weaviate-cloud
weave cols ds AllianceDemoImages --no-confirm --weaviate-cloud
```

---

## Quick Reference

| Feature | Command |
| --- | --- |
| Install | `brew install Maximilien-ai/weave-cli/weave-cli` |
| Diagnose | `weave doctor` |
| Health | `weave health check` |
| Collections | `weave cols ls / create / show / query` |
| Documents | `weave docs create / ls / show / del` |
| Batch ingest | `weave pipeline ingest DIR --collection COL` |
| Backup | `weave backup create / restore / validate` |
| Embeddings | `weave embeddings list` |
| Evals | `weave eval run --agent X --dataset Y` |
| Agents | `weave agents list / show / create` |
| REPL | `weave` (no args) |
| Statistics | `weave stats COL` |

---

**GitHub**: <https://github.com/maximilien/weave-cli>
**Install**: `brew install Maximilien-ai/weave-cli/weave-cli`
