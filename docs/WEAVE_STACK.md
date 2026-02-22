# Weave Stack: Zero-to-Dashboard in Under One Day

**Status**: 📋 Proposal
**Goal**: Reduce client onboarding from weeks to <1 day
**Inspired by**: Docker Compose, Kubernetes Helm, Rails/Laravel stack management
**Version**: 1.0 (Draft)

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Problem Statement](#problem-statement)
3. [Architecture Overview](#architecture-overview)
4. [Core Commands](#core-commands)
5. [Stack File Format](#stack-file-format)
6. [Component Lifecycle](#component-lifecycle)
7. [Implementation Examples](#implementation-examples)
8. [Day Two Operations](#day-two-operations)
9. [Best Practices from Industry](#best-practices-from-industry)
10. [Implementation Plan](#implementation-plan)

---

## Executive Summary

**Current Reality** (Client0 Experience):
- **Weeks 1-2**: VDB selection, Docker setup, weave-cli installation, config debugging
- **Weeks 3-4**: Data ingestion iteration (memory issues, glob patterns, retry logic)
- **Weeks 5-6**: Evaluation setup, dashboard integration, production hardening
- **Total**: 6+ weeks from data directory to running dashboard

**Weave Stack Vision**:
```bash
# Day 1 (Morning): Define your stack
weave stack init

# Day 1 (Afternoon): Run interactive wizard to configure
weave stack wizard

# Day 1 (Evening): Deploy everything
weave stack up

# Result: Running dashboard with ingested data and evaluations
```

**Key Innovation**: Declarative YAML + Automated Orchestration + Self-Healing Monitoring

---

## Problem Statement

### Client0 Pain Points (Analyzed from `/Users/maximilien/github/auctionsmax-ai`)

| Phase | Manual Steps | Time | Failure Points |
|-------|--------------|------|----------------|
| **Setup** | Docker install, VDB selection, config files, .env | 1-2 days | Docker networking, port conflicts, memory limits |
| **Ingestion** | Write bash scripts, handle retries, monitor progress | 5-10 days | Milvus OOM, glob patterns, PDF routing, checkpoint logic |
| **Evaluation** | Install eval tools, create test cases, run baselines | 3-5 days | Dependency conflicts, API rate limits, metric interpretation |
| **Dashboard** | Frontend setup, API integration, health checks | 3-5 days | CORS, auth, state management, error handling |

**Root Cause**: Each client re-invents the same orchestration, retry logic, and monitoring.

**Client0 Created** (to work around gaps):
- `run.sh` - 240 lines of Docker + health check orchestration
- `reingest-pdf-monitored.sh` - 255 lines of retry + checkpoint + Milvus restart logic
- `ingestion-status.sh` - 100 lines of progress monitoring
- `test.sh` - 850 lines of end-to-end testing
- Custom `dev.sh`, collection scripts, eval runners, etc.

**Total LOC**: ~2,000+ lines of bespoke glue code that should be standardized.

---

## Architecture Overview

### Philosophy

**Inspired by**:
- **Docker Compose**: Simple YAML, single command up/down, service dependencies
- **Kubernetes Helm**: Templating, values.yaml for config, lifecycle hooks
- **Terraform**: Declarative infra, plan/apply/destroy, state management
- **Rails migrations**: Version-tracked schema changes, rollback support

**Weave Stack Principles**:
1. **Declarative over imperative** - Define what you want, not how to get it
2. **Self-healing** - Auto-restart services, retry failed ingestions
3. **Observable** - Real-time progress, health checks, failure alerts
4. **Reproducible** - Same stack.yaml on dev, staging, prod
5. **Composable** - Mix and match VDBs, LLMs, evals, dashboards

### Stack Architecture

```
weave-stack/
├── weave-stack.yaml         # Main stack definition
├── .env                     # Secrets (API keys)
├── config.yaml              # weave-cli config (auto-generated from stack)
├── data/                    # Source documents
│   ├── pdfs/
│   ├── docs/
│   └── ...
├── docker-compose.yaml      # Auto-generated from stack.yaml
├── .weave-state/            # Stack runtime state
│   ├── services.json        # Service health status
│   ├── ingestion.json       # Ingestion checkpoints
│   ├── evaluations.json     # Eval results
│   └── dashboard.json       # Dashboard status
└── logs/                    # Centralized logs
    ├── vdb/
    ├── ingestion/
    ├── evals/
    └── dashboard/
```

---

## Core Commands

### Lifecycle Commands

```bash
# Initialize a new stack (creates weave-stack.yaml template)
weave stack init [--template <name>]

# Run interactive wizard to configure stack
weave stack wizard

# Validate stack configuration
weave stack validate

# Show what will be deployed (dry-run)
weave stack plan

# Deploy all services + ingest data + run evals + start dashboard
weave stack up [--detach] [--skip-ingestion] [--skip-evals]

# Show real-time status of all components
weave stack status [--watch] [--json]

# Monitor ingestion/eval progress
weave stack monitor [--component <name>]

# View logs from all services
weave stack logs [--service <name>] [--follow]

# Update stack after changing weave-stack.yaml
weave stack update [--component <name>]

# Restart specific component
weave stack restart <component>

# Stop all services (preserve data)
weave stack down

# Destroy everything (including data)
weave stack destroy [--force]
```

### Day Two Operations

```bash
# Add new data source
weave stack add-data <glob-pattern>

# Re-ingest specific collection
weave stack reingest <collection> [--resume]

# Switch VDB (e.g., Milvus → Qdrant)
weave stack migrate --from milvus-local --to qdrant-local

# Run evaluations
weave stack eval [--collection <name>]

# Health check all services
weave stack health

# Export stack state for backup
weave stack export --output backup.tar.gz

# Import stack state
weave stack import backup.tar.gz
```

---

## Stack File Format

### `weave-stack.yaml` (Full Example)

```yaml
version: "1.0"
name: "auctionsmax-ai"
description: "Rare camera auction RAG system"

# ========================================
# Infrastructure
# ========================================
infrastructure:
  # Vector Database
  vectordb:
    type: milvus-local  # Options: milvus-local, milvus-cloud, qdrant-local, weaviate-local, etc.
    version: "2.3.0"
    resources:
      memory: "12G"      # Learned from Client0: image ingestion needs 8-10G
      cpu: "4"
    config:
      similarity_metric: COSINE
      database: default

    # Health checks (for auto-restart)
    health:
      endpoint: http://localhost:9091/healthz
      interval: 30s
      retries: 10
      startup_timeout: 60s

  # LLM Provider
  llm:
    provider: openai
    models:
      embedding: text-embedding-3-small
      chat: gpt-4o
    config:
      temperature: 0.7
      max_tokens: 4096

  # Image Storage (for large images ≥65KB)
  image_storage:
    type: minio         # Options: minio, s3, local
    bucket: weave-images
    endpoint: localhost:9000
    resources:
      memory: "2G"

  # Observability (optional)
  monitoring:
    enabled: false      # Future: Prometheus + Grafana
    metrics_port: 9090

# ========================================
# Data Collections
# ========================================
collections:
  - name: AuctionListings
    type: text
    description: "Auction lot descriptions and estimates from catalogs"
    schema:
      vector_dimensions: 1536
      fields:
        - name: lot_number
          type: string
        - name: estimate
          type: string
        - name: description
          type: text
        - name: auction_year
          type: int

    # Data sources
    sources:
      - pattern: "data/tamarkin/*-catalogue.pdf"
        type: pdf
        mode: text-only

    # Chunking strategy
    chunking:
      strategy: semantic
      chunk_size: 500
      chunk_overlap: 50

    # Embedding
    embedding:
      model: text-embedding-3-small
      provider: openai

  - name: AuctionImages
    type: image
    description: "Product images with OCR from auction catalogs"
    schema:
      vector_dimensions: 1536
      fields:
        - name: image_url
          type: string
        - name: page_number
          type: int
        - name: ocr_text
          type: text

    sources:
      - pattern: "data/tamarkin/*-catalogue.pdf"
        type: pdf
        mode: image-extraction
        options:
          extract_images: true
          skip_small_images: true
          min_image_size_kb: 10
          store_pdf: true  # Store original PDF in MinIO

    embedding:
      model: text-embedding-3-small
      provider: openai

    # Image-specific settings
    image_storage:
      threshold_kb: 65   # Images ≥65KB go to external storage
      formats: [jpg, png, webp]

  - name: AuctionResults
    type: text
    description: "Final sale prices and results"
    schema:
      vector_dimensions: 1536

    sources:
      - pattern: "data/tamarkin/Results-sheet-*.pdf"
        type: pdf
        mode: text-only

    chunking:
      strategy: fixed
      chunk_size: 300

    embedding:
      model: text-embedding-3-small
      provider: openai

# ========================================
# Ingestion Pipeline
# ========================================
ingestion:
  # Global settings
  parallel_workers: 1        # Default: sequential (safer for memory)
  batch_size: 10             # Files per batch
  retry:
    max_attempts: 3
    backoff: exponential     # 30s, 60s, 120s
    base_delay: 30s

  timeout:
    per_file: 30m            # Learned from Client0: large PDFs can take 20+ min
    total: 8h                # Safety limit for full ingestion

  # Checkpointing (for resume on failure)
  checkpoint:
    enabled: true
    path: .weave-state/ingestion.json
    save_interval: 1         # Save after each file

  # Health monitoring
  monitoring:
    vdb_health_check: true
    restart_on_oom: true     # Auto-restart Milvus if OOM detected
    memory_threshold: "90%"  # Restart if memory usage >90%

  # Phase ordering (text-only before images)
  phases:
    - name: text-collections
      collections: [AuctionListings, AuctionResults]
      parallel: false        # Sequential for safety

    - name: image-collections
      collections: [AuctionImages]
      parallel: false
      restart_vdb_between_files: true  # Prevent memory buildup

# ========================================
# Evaluations
# ========================================
evaluations:
  enabled: true

  # Eval suites
  suites:
    - name: baseline-quality
      type: rag-quality
      collections: [AuctionListings]
      metrics:
        - faithfulness
        - answer_relevance
        - context_recall

      test_queries:
        - "Leica M3 camera"
        - "black paint Leica M6"
        - "Canadian Midland"
        - "single stroke advance"

      baseline:
        min_faithfulness: 0.7
        min_relevance: 0.6

    - name: image-retrieval
      type: multimodal
      collections: [AuctionImages]
      metrics:
        - image_relevance
        - ocr_accuracy

      test_queries:
        - "Leica M3 camera front view"
        - "camera with black paint finish"

  # Reporting
  output:
    format: json
    path: .weave-state/evaluations.json
    dashboard_webhook: null  # Future: POST results to dashboard

# ========================================
# Dashboard (Optional)
# ========================================
dashboard:
  enabled: true
  type: web              # Options: web, cli, none

  web:
    framework: nextjs    # What Client0 uses
    port: 3000
    host: localhost

    # Auto-generated API routes
    api:
      - path: /api/search
        collection: AuctionListings
      - path: /api/search-images
        collection: AuctionImages
      - path: /api/health
        type: healthcheck

    # Environment
    env_file: .env

    # Deployment
    build:
      command: npm run build
      output: .next

    start:
      command: npm run dev
      hot_reload: true

# ========================================
# Operational Settings
# ========================================
operations:
  # Logging
  logging:
    level: info          # debug, info, warn, error
    format: json
    outputs:
      - type: file
        path: logs/stack.log
      - type: stdout

  # Cleanup policies
  cleanup:
    remove_checkpoints_on_success: false
    retain_logs_days: 30
    compress_old_logs: true

  # Backup
  backup:
    enabled: false       # Future: auto-backup to S3
    schedule: "0 2 * * *"  # Daily at 2am
    retention_days: 7
```

### Minimal Stack Example (Quickstart)

```yaml
version: "1.0"
name: "quickstart"

infrastructure:
  vectordb:
    type: milvus-local
  llm:
    provider: openai
    models:
      embedding: text-embedding-3-small

collections:
  - name: Documents
    type: text
    sources:
      - pattern: "data/**/*.pdf"
        type: pdf

ingestion:
  retry:
    max_attempts: 3

# That's it! Sensible defaults for everything else
```

---

## Component Lifecycle

### State Machine

Each component (VDB, collection, eval suite, dashboard) has a lifecycle:

```
┌──────────┐
│ DEFINED  │ (in weave-stack.yaml)
└────┬─────┘
     │ weave stack up
     ▼
┌──────────┐
│ PLANNED  │ (validated, resources allocated)
└────┬─────┘
     │ deploy
     ▼
┌──────────┐
│ STARTING │ (Docker containers launching)
└────┬─────┘
     │ health check passes
     ▼
┌──────────┐
│ HEALTHY  │ ◄─── Self-healing restarts if unhealthy
└────┬─────┘
     │ weave stack down
     ▼
┌──────────┐
│ STOPPED  │
└────┬─────┘
     │ weave stack destroy
     ▼
┌──────────┐
│ DESTROYED│
└──────────┘
```

### Self-Healing Logic

**VDB Health Monitoring**:
```bash
# Every 30 seconds
1. HTTP GET {vdb.health.endpoint}
2. If 200 OK: HEALTHY
3. If connection refused:
   - Check Docker container status
   - If exited (OOM): restart container + wait 60s
   - Retry health check
   - If still failing after 10 retries: ALERT + PAUSE ingestion
```

**Ingestion Checkpoint Logic**:
```bash
# After each file
1. Save to .weave-state/ingestion.json:
   - completed_files: [...]
   - failed_files: [{file, error, attempts}]
   - current_phase: "image-collections"

2. On next `weave stack up --resume`:
   - Skip files in completed_files
   - Retry files in failed_files (up to max_attempts)
   - Resume from current_phase
```

---

## Implementation Examples

### Example 1: Client0 Onboarding in <1 Day

**Current**: 6+ weeks
**With Weave Stack**: 6 hours

#### Morning (2 hours): Setup

```bash
# 1. Clone project template
git clone https://github.com/maximilien/weave-stack-template.git auctionsmax-ai
cd auctionsmax-ai

# 2. Run wizard
weave stack wizard

# Interactive prompts:
#
# 📦 Project name: auctionsmax-ai
# 📄 Data location: data/tamarkin/
#
# 🗄️  Vector Database:
#   1. Milvus (local)
#   2. Qdrant (local)
#   3. Weaviate (local)
#   Choice: 1
#
# 🤖 LLM Provider:
#   1. OpenAI
#   2. Ollama (local)
#   3. Anthropic
#   Choice: 1
#   API Key: sk-...
#
# 📊 Enable evaluations? (y/n): y
# 🌐 Enable dashboard? (y/n): y
#
# ✅ Generated weave-stack.yaml
# ✅ Generated .env
# ✅ Generated config.yaml

# 3. Review generated files
cat weave-stack.yaml
nano weave-stack.yaml  # Tweak if needed
```

#### Afternoon (3 hours): Deploy

```bash
# 4. Deploy everything
weave stack up

# Output:
#
# 🚀 Weave Stack Deployment
# ════════════════════════════════════════
#
# [1/5] 🐳 Starting Docker services...
#   ✅ milvus-standalone (2.3.0)
#   ✅ minio (latest)
#   ⏳ Waiting for health checks... (60s)
#   ✅ All services healthy
#
# [2/5] 📦 Creating collections...
#   ✅ AuctionListings (1536d, COSINE)
#   ✅ AuctionImages (1536d, COSINE)
#   ✅ AuctionResults (1536d, COSINE)
#
# [3/5] 📥 Ingesting data...
#   Phase: text-collections
#     ⏳ AuctionListings: 9 files
#       [████████████████████] 9/9 (100%) | ETA: 0s
#       ✅ 267 chunks created (12m 33s)
#     ⏳ AuctionResults: 6 files
#       [████████████████████] 6/6 (100%)
#       ✅ 127 chunks created (3m 42s)
#
#   Phase: image-collections
#     ⏳ AuctionImages: 9 files
#       [████░░░░░░░░░░░░░░░░] 2/9 (22%) | ETA: 1h 23m
#       ... (real-time progress)
#       ✅ 6591 images created (2h 15m)
#
# [4/5] 🧪 Running evaluations...
#   ✅ baseline-quality: 4/4 queries passed
#      - faithfulness: 0.82 (target: 0.7)
#      - relevance: 0.71 (target: 0.6)
#   ✅ image-retrieval: 2/2 queries passed
#
# [5/5] 🌐 Starting dashboard...
#   ✅ Next.js dev server on http://localhost:3000
#
# ════════════════════════════════════════
# 🎉 Stack deployment complete!
#
# 📊 Status: weave stack status
# 🔍 Logs:   weave stack logs --follow
# 🌐 Dashboard: http://localhost:3000
#
# Next steps:
#   - Visit http://localhost:3000
#   - Try a search: curl http://localhost:3000/api/search?q=Leica+M3
#   - Run evals: weave stack eval
```

#### Evening (1 hour): Test & Iterate

```bash
# 5. Monitor status
weave stack status

# Output:
# Component       Status    Health  Uptime   Resources
# ─────────────────────────────────────────────────────
# milvus          HEALTHY   100%    2h 15m   8.2G / 12G
# minio           HEALTHY   100%    2h 15m   0.3G / 2G
# AuctionListings READY     -       -        267 docs
# AuctionImages   READY     -       -        6591 docs
# AuctionResults  READY     -       -        127 docs
# evals           PASSED    -       -        6/6 suites
# dashboard       HEALTHY   100%    15m      http://localhost:3000

# 6. Test a query
curl "http://localhost:3000/api/search?q=Leica%20M3%20black%20paint"

# 7. View ingestion logs
weave stack logs --service ingestion --tail 100

# 8. Check eval results
weave stack eval --report

# Done! Running system in <6 hours
```

### Example 2: Adding New Data (Day 2)

**Scenario**: Client0 gets 2026 auction catalog

```bash
# 1. Add new PDF to data directory
cp ~/Downloads/2026-tamarkin-auction-catalogue.pdf data/tamarkin/

# 2. Update stack (auto-detects new file via glob pattern)
weave stack update --component AuctionListings

# Output:
# 🔄 Updating AuctionListings...
#
# Detected new files:
#   + data/tamarkin/2026-tamarkin-auction-catalogue.pdf
#
# ⏳ Ingesting 1 new file...
#   [████████████████████] 1/1 (100%)
#   ✅ 31 new chunks created (14m 22s)
#
# ✅ Collection updated: 267 → 298 docs

# 3. Re-run evals
weave stack eval --collection AuctionListings

# Done! New data ingested in 15 minutes
```

### Example 3: Switching VDB (Milvus → Qdrant)

```bash
# 1. Edit weave-stack.yaml
# Change:
#   vectordb:
#     type: milvus-local
# To:
#   vectordb:
#     type: qdrant-local

# 2. Migrate
weave stack migrate --from milvus-local --to qdrant-local

# Output:
# 🔄 Migration Plan
# ════════════════════════════════════════
# From: milvus-local
# To:   qdrant-local
#
# Collections to migrate:
#   - AuctionListings (267 docs)
#   - AuctionImages (6591 docs)
#   - AuctionResults (127 docs)
#
# Estimated time: ~45 minutes
#
# Proceed? (y/n): y
#
# [1/4] 🐳 Starting qdrant-local...
#   ✅ qdrant-standalone (v1.7.0)
#
# [2/4] 📦 Creating collections in Qdrant...
#   ✅ AuctionListings
#   ✅ AuctionImages
#   ✅ AuctionResults
#
# [3/4] 📥 Migrating data...
#   ⏳ AuctionListings: [████████████] 267/267 (100%)
#   ⏳ AuctionImages: [████████████] 6591/6591 (100%)
#   ⏳ AuctionResults: [████████████] 127/127 (100%)
#
# [4/4] ✅ Updating weave-stack.yaml...
#
# 🎉 Migration complete!
#
# Old VDB (milvus-local) is still running.
# To remove: weave stack down --service milvus-local

# 3. Verify
weave stack status

# Component       Status    Health  VDB
# ─────────────────────────────────────────────
# qdrant          HEALTHY   100%    qdrant-local
# AuctionListings READY     -       267 docs
# ...
```

---

## Day Two Operations

### Common Workflows

#### 1. Re-ingestion After Failure

**Scenario**: Ingestion crashed at file 5/9 due to OOM

```bash
# Resume from checkpoint
weave stack up --resume

# Output:
# 🔄 Resuming ingestion from checkpoint...
#
# Already completed: 4/9 files
# Remaining: 5/9 files
# Failed (will retry): 1 file
#
# ⏳ Resuming phase: image-collections
#   ⏳ AuctionImages: 5 files (+ 1 retry)
#     [██████░░░░░░░░░░░░░░] 3/6 (50%) | ETA: 42m
```

#### 2. Testing New Embedding Model

**Scenario**: Try `all-mpnet-base-v2` (OSS) vs OpenAI

```bash
# 1. Duplicate collection in stack.yaml
# Add:
# - name: AuctionListings_OSS
#   type: text
#   sources:
#     - pattern: "data/tamarkin/*-catalogue.pdf"
#       type: pdf
#   embedding:
#     model: sentence-transformers/all-mpnet-base-v2
#     provider: sentence-transformers
#   schema:
#     vector_dimensions: 768  # Different from OpenAI!

# 2. Update stack
weave stack update --component AuctionListings_OSS

# Output:
# 🔄 Creating new collection: AuctionListings_OSS
#
# ⏳ Ingesting 9 files with OSS embeddings...
#   [████████████████████] 9/9 (100%)
#   ✅ 267 chunks created (8m 15s)  # Faster! No API calls
#
# 💰 Cost comparison:
#   OpenAI (text-embedding-3-small): $0.008
#   OSS (sentence-transformers):      $0.000
#   Savings: 100%

# 3. Compare quality
weave stack eval --collections AuctionListings,AuctionListings_OSS

# Output:
# 📊 Evaluation Comparison
# ════════════════════════════════════════
# Collection             Faithfulness  Relevance
# AuctionListings        0.82          0.71
# AuctionListings_OSS    0.89          0.74
#
# 🎉 AuctionListings_OSS is better!
```

#### 3. Monitoring Long-Running Ingestion

```bash
# Start ingestion in background
weave stack up --detach

# Monitor in separate terminal
weave stack monitor --component ingestion --watch

# Output (refreshes every 5s):
# ════════════════════════════════════════
# ⏳ Ingestion Progress (Live)
# ════════════════════════════════════════
#
# Phase: image-collections
# File: 2023-tamarkin-auction-catalogue.pdf (6/9)
#
# Progress: [████████░░░░░░░░░░░░] 67% (6/9 files)
# Chunks:   1,234 / ~2,000 estimated
# Images:   892 / ~1,200 estimated
# Elapsed:  2h 14m
# ETA:      ~45min remaining
#
# VDB Health: ✅ HEALTHY (7.8G / 12G memory)
#
# Recent logs:
#   14:23:15 [INFO] Processing image 892/1200
#   14:23:18 [INFO] Image uploaded to MinIO: uuid.jpg
#   14:23:21 [WARN] Image 893 exceeds 65KB, storing URL only
```

#### 4. Backup & Restore

```bash
# Backup entire stack state
weave stack export --output backup-2026-02-22.tar.gz

# Includes:
# - weave-stack.yaml
# - .env
# - .weave-state/ (checkpoints, eval results)
# - VDB snapshots (Milvus collections export)
# - Logs (last 7 days)

# Restore on new machine
weave stack import backup-2026-02-22.tar.gz
weave stack up

# Output:
# 🔄 Restoring from backup...
#
# [1/3] 📦 Recreating collections...
# [2/3] 📥 Importing VDB data...
#   ⏳ AuctionListings: 267 docs
#   ⏳ AuctionImages: 6591 docs
# [3/3] 🌐 Starting dashboard...
#
# ✅ Restore complete!
```

---

## Best Practices from Industry

### Inspiration Sources

| Platform | What We Adopted | Why |
|----------|-----------------|-----|
| **Docker Compose** | YAML services, `up`/`down`, health checks | Familiar to devs, simple orchestration |
| **Kubernetes Helm** | Values templating, lifecycle hooks, rollback | Production-grade deployment patterns |
| **Terraform** | `plan` before `apply`, state management, destroy | Infra-as-code best practices |
| **Rails** | Migrations (versioned, rollback), generators | Developer experience, convention over config |
| **Laravel Artisan** | CLI wizard (`php artisan make:*`), task runners | Guided setup, automation |
| **Airflow DAGs** | Task dependencies, retry logic, monitoring | Pipeline orchestration |
| **Prometheus** | Health checks, self-healing, alerting | Production observability |

### Weave Stack Design Decisions

#### 1. Declarative YAML (not imperative scripts)

**Why**: Client0 wrote 2,000+ LOC of bash scripts to orchestrate ingestion, retries, monitoring. Every client will need the same logic.

**How**: Define *what* you want in `weave-stack.yaml`, weave-cli figures out *how*.

**Example**:
```yaml
# Declarative (Weave Stack)
ingestion:
  retry:
    max_attempts: 3
    backoff: exponential

# vs Imperative (Client0's bash)
for attempt in 1 2 3; do
  if weave docs create ...; then
    break
  else
    sleep $((30 * 2 ** (attempt - 1)))
  fi
done
```

#### 2. Self-Healing (Kubernetes-inspired)

**Why**: Milvus OOM crashes are common (Client0 experienced this repeatedly). Manual intervention breaks flow.

**How**: Health checks + auto-restart + checkpoint resume.

**Example**:
```yaml
infrastructure:
  vectordb:
    health:
      endpoint: http://localhost:9091/healthz
      interval: 30s
      retries: 10

ingestion:
  monitoring:
    restart_on_oom: true  # Auto-restart Milvus if OOM detected
    resume_on_restart: true
```

#### 3. Wizard-Driven Setup (Rails/Laravel-inspired)

**Why**: Reduces barrier to entry. Non-experts can get started.

**How**: Interactive prompts → generated YAML.

**Example**:
```bash
$ weave stack wizard

📦 Project Setup
─────────────────
Name: my-rag-app
Description: Knowledge base for company docs

🗄️  Vector Database
─────────────────
1. Milvus (recommended, good for images)
2. Qdrant (lightweight, easy setup)
3. Weaviate (cloud-native, GraphQL API)
Choice [1-3]: 1

🤖 LLM Provider
─────────────────
1. OpenAI (best quality, paid)
2. Ollama (free, local, privacy)
Choice [1-2]: 2

✅ Installing Ollama... (requires sudo)
✅ Pulling nomic-embed-text model...

📁 Data Location
─────────────────
Where are your documents? [./data]: ./company-docs

✅ Found 127 files:
   - 45 PDFs
   - 67 DOCX
   - 15 TXT

📊 Evaluations
─────────────────
Enable automatic quality testing? [y/n]: y

🌐 Dashboard
─────────────────
Enable web dashboard? [y/n]: y

✅ Generated weave-stack.yaml
✅ Generated .env (add API keys if needed)
✅ Generated docker-compose.yaml

Next: weave stack up
```

#### 4. Checkpoint-Based Resume (Airflow-inspired)

**Why**: Client0's ingestion took 2+ hours. If it crashed at file 7/9, starting over wastes time.

**How**: Save state after each file. Resume skips completed files.

**Example**:
```json
// .weave-state/ingestion.json
{
  "started_at": "2026-02-22T10:00:00Z",
  "phase": "image-collections",
  "completed_files": [
    "2017-catalogue.pdf",
    "2018-catalogue.pdf",
    "2019-catalogue.pdf"
  ],
  "failed_files": [
    {
      "file": "2020-catalogue.pdf",
      "error": "context deadline exceeded",
      "attempts": 2
    }
  ],
  "current_file": "2021-catalogue.pdf",
  "progress": {
    "total_files": 9,
    "completed": 3,
    "failed": 1,
    "remaining": 5
  }
}
```

#### 5. State Management (Terraform-inspired)

**Why**: Know what's deployed, what changed, what will happen.

**How**: Store runtime state in `.weave-state/`, show plan before apply.

**Example**:
```bash
$ weave stack plan

Weave Stack Plan
════════════════════════════════════════

Terraform-style diff:

+ infrastructure.vectordb.milvus-local
  + resources.memory: 12G
  + health.endpoint: http://localhost:9091/healthz

+ collections.AuctionListings
  + sources: 9 files (*.pdf)
  + schema.dimensions: 1536

~ collections.AuctionImages
  - sources: 8 files
  + sources: 9 files (1 new)

Will create:
  - 3 collections
  - 2 Docker services
  - 1 dashboard

Will update:
  - AuctionImages (1 new file)

Proceed? (y/n):
```

---

## Implementation Plan

### Phase 1: Core Stack Management (Week 1-2)

**Goal**: `weave stack up` deploys VDB + ingests data

**Deliverables**:
- [ ] `weave stack init` - Generate template `weave-stack.yaml`
- [ ] `weave stack validate` - Parse and validate YAML
- [ ] `weave stack plan` - Show deployment plan (dry-run)
- [ ] `weave stack up` - Deploy VDB + create collections + ingest data
- [ ] `weave stack down` - Stop services
- [ ] `weave stack status` - Show component health

**Files**:
- `src/cmd/stack/init.go`
- `src/cmd/stack/validate.go`
- `src/cmd/stack/plan.go`
- `src/cmd/stack/up.go`
- `src/cmd/stack/down.go`
- `src/cmd/stack/status.go`
- `src/pkg/stack/config.go` - Parse weave-stack.yaml
- `src/pkg/stack/orchestrator.go` - Deployment logic
- `src/pkg/stack/health.go` - Health checks

**Testing**:
- Unit tests for YAML parsing
- Integration test: Deploy Milvus + ingest 3 PDFs
- Verify checkpoint/resume works

### Phase 2: Self-Healing & Monitoring (Week 3)

**Goal**: Auto-restart on failure + real-time progress

**Deliverables**:
- [ ] `weave stack monitor` - Real-time progress dashboard
- [ ] `weave stack logs` - Centralized log viewer
- [ ] Health check auto-restart
- [ ] OOM detection + Milvus restart
- [ ] Checkpoint save/resume

**Files**:
- `src/cmd/stack/monitor.go`
- `src/cmd/stack/logs.go`
- `src/pkg/stack/health.go` (extend)
- `src/pkg/stack/checkpoint.go`
- `.weave-state/` state directory structure

**Testing**:
- Simulate Milvus crash → verify auto-restart
- Kill ingestion mid-way → verify resume
- Monitor UI shows real-time progress

### Phase 3: Wizard & Templates (Week 4)

**Goal**: `weave stack wizard` generates stack from prompts

**Deliverables**:
- [ ] `weave stack wizard` - Interactive setup
- [ ] Stack templates (quickstart, production, multimodal)
- [ ] Ollama auto-install
- [ ] API key validation

**Files**:
- `src/cmd/stack/wizard.go`
- `src/pkg/stack/templates/` - YAML templates
- `src/pkg/stack/prompts.go` - Interactive prompts

**Templates**:
- `quickstart.yaml` - Minimal stack (1 collection, local Milvus)
- `production.yaml` - Full stack (3+ collections, evals, dashboard)
- `multimodal.yaml` - Image + text collections
- `oss.yaml` - Ollama + sentence-transformers (no API keys)

**Testing**:
- Run wizard → verify generated YAML is valid
- Deploy from template → verify works

### Phase 4: Day Two Operations (Week 5)

**Goal**: Update, migrate, backup

**Deliverables**:
- [ ] `weave stack update` - Add new data / change config
- [ ] `weave stack migrate` - Switch VDB
- [ ] `weave stack export/import` - Backup/restore
- [ ] `weave stack reingest` - Re-ingest collection

**Files**:
- `src/cmd/stack/update.go`
- `src/cmd/stack/migrate.go`
- `src/cmd/stack/export.go`
- `src/cmd/stack/import.go`
- `src/pkg/stack/migration.go`

**Testing**:
- Migrate Milvus → Qdrant → verify data intact
- Export → destroy → import → verify restored
- Add new file → update → verify ingested

### Phase 5: Evaluations & Dashboard (Week 6)

**Goal**: Integrated evals + dashboard deployment

**Deliverables**:
- [ ] `weave stack eval` - Run evaluation suites
- [ ] Dashboard auto-deploy (Next.js)
- [ ] API route generation
- [ ] Eval results in dashboard

**Files**:
- `src/cmd/stack/eval.go`
- `src/pkg/stack/evaluations.go`
- `src/pkg/stack/dashboard.go`
- `templates/dashboard/` - Next.js starter

**Testing**:
- Run evals → verify metrics collected
- Deploy dashboard → verify search works
- Check eval results in UI

---

## Open Questions

1. **Docker Compose vs Kubernetes**:
   - Start with Docker Compose for simplicity?
   - Add Kubernetes support later for production?

2. **Dashboard Framework**:
   - Generate Next.js boilerplate?
   - Support other frameworks (Streamlit, Gradio)?
   - Or just expose API, let users choose UI?

3. **Cloud Deployment**:
   - Support cloud VDBs (Milvus Cloud, Qdrant Cloud)?
   - Generate Terraform for AWS/GCP/Azure?
   - Or focus on local-first?

4. **Observability**:
   - Integrate Prometheus + Grafana?
   - Or build custom monitoring?

5. **Multi-Tenancy**:
   - Support multiple stacks on same machine?
   - Namespace collision prevention?

---

## Success Metrics

**Quantitative**:
- Client onboarding time: 6 weeks → <1 day (6+ weeks → 6 hours)
- Lines of custom glue code: 2,000+ → 0 (all in weave-stack.yaml)
- Time to add new data: 2 hours → 15 minutes
- VDB migration time: 1 week → 1 hour

**Qualitative**:
- Non-technical users can deploy RAG systems
- Consistent deployments (dev/staging/prod)
- Self-healing reduces ops burden
- Easy to experiment (new VDBs, embeddings)

---

## Next Steps

1. **Review & Feedback** (You!)
   - Is the vision clear?
   - Any missing workflows?
   - Concerns about complexity?

2. **Prototype** (Week 1-2)
   - Minimal `weave stack up` (VDB + ingest)
   - Test with Client0's data
   - Validate 6-hour onboarding claim

3. **Dogfood** (Week 3-4)
   - Use Weave Stack for Client1 onboarding
   - Iterate based on real feedback

4. **GA Release** (Week 5-6)
   - Ship v0.10.0 with `weave stack`
   - Documentation + video demo
   - Blog post: "Zero to Dashboard in 6 Hours"

---

**LFG! Let's make RAG accessible to everyone.** 🚀

---

**Document Status**: 📋 Proposal (awaiting review)
**Author**: Claude + dr.max collaboration
**Date**: 2026-02-21
**Next Review**: After user feedback
