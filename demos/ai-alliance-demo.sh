#!/bin/bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2026 dr.max

# =================================================================
# AI Alliance Demo — Weave CLI v0.12.0
# Duration: ~45 minutes (leave 15 min for Q&A)
# Date: March 26, 2026
#
# Uses current config.yaml VDBs (weaviate-cloud default).
# Data: weave-cli repo docs + test fixtures + Desktop PDFs.
#
# Usage: Run manually command by command, or press Enter between
#        pages if running the script directly.
# =================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
WEAVE="$PROJECT_ROOT/bin/weave"

# Colors
CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BOLD='\033[1m'
NC='\033[0m'

page() {
    echo ""
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BOLD}  $1${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
}

run() {
    echo -e "${GREEN}\$ $*${NC}"
    eval "$@"
    echo ""
}

pause() {
    echo -e "${YELLOW}[Press Enter]${NC}"
    read -r
    clear
}

DEMO_COL="AllianceDemo"
DEMO_IMG_COL="AllianceDemoImages"

# =================================================================
page "AI Alliance Demo — Weave CLI v0.12.0"
echo "AI-powered CLI for managing 11+ vector databases."
echo "Built in Go. Single binary. Open source."
echo ""
echo "Today: install → diagnose → ingest → search → backup"
echo "       across real cloud VDBs, live."
pause

# =================================================================
# PART 1: Installation & Setup (~3 min)
# =================================================================
page "Part 1: Installation & Setup"

echo "Install via Homebrew (new in v0.12.0):"
echo ""
echo "  brew install Maximilien-ai/weave-cli/weave-cli"
echo ""
echo "(Already installed — let's verify)"
echo ""

run "$WEAVE" --version
pause

# =================================================================
page "weave doctor — Unified Diagnostics (NEW)"

echo "One command checks everything: system, config, env,"
echo "VDB connectivity, LLM, embeddings, stack, and Opik."
echo ""

run "$WEAVE" doctor --section system --verbose
pause

run "$WEAVE" doctor --section config
pause

run "$WEAVE" doctor --section vdb
pause

# =================================================================
# PART 2: Configuration & Health (~3 min)
# =================================================================
page "Part 2: Configuration & VDB Health"

run "$WEAVE" config list
pause

run "$WEAVE" health check --cloud
pause

run "$WEAVE" vdb list
pause

# =================================================================
# PART 3: Collections & Documents (~10 min)
# =================================================================
page "Part 3: Collections & Documents"

echo "List existing collections:"
run "$WEAVE" cols ls
pause

echo "Create a demo collection for text documents:"
run "$WEAVE" cols create $DEMO_COL \
    --embedding text-embedding-3-small --weaviate-cloud || true
pause

echo "Create a demo collection for images:"
run "$WEAVE" cols create $DEMO_IMG_COL \
    --image --weaviate-cloud || true
pause

# --- Ingest text documents ---
page "Ingest Text Documents"

echo "Ingest the weave-cli architecture doc:"
run "$WEAVE" docs create $DEMO_COL \
    "$PROJECT_ROOT/docs/ARCHITECTURE.md" --weaviate-cloud
pause

echo "Ingest the README:"
run "$WEAVE" docs create $DEMO_COL \
    "$PROJECT_ROOT/README.md" --weaviate-cloud
pause

echo "Ingest a PDF (text + image extraction):"
run "$WEAVE" docs create $DEMO_COL \
    "$PROJECT_ROOT/tests/fixtures/ragme-io.pdf" \
    --image-col $DEMO_IMG_COL --weaviate-cloud || true
pause

echo "List documents:"
run "$WEAVE" docs ls $DEMO_COL --weaviate-cloud
pause

# --- Ingest images ---
page "Ingest Images"

echo "Add test images:"
run "$WEAVE" docs create $DEMO_IMG_COL \
    "$PROJECT_ROOT/tests/images/dog.png" --weaviate-cloud || true
pause

run "$WEAVE" docs ls $DEMO_IMG_COL --weaviate-cloud
pause

# =================================================================
# PART 4: Semantic Search (~5 min)
# =================================================================
page "Part 4: Semantic Search (RAG)"

echo "Search across ingested documents using natural language:"
echo ""

run "$WEAVE" cols query $DEMO_COL \
    "how does the vector database abstraction work?" \
    --top-k 3 --weaviate-cloud
pause

run "$WEAVE" cols query $DEMO_COL \
    "what embedding models are supported?" \
    --top-k 3 --weaviate-cloud
pause

run "$WEAVE" cols query $DEMO_COL \
    "kubernetes deployment" \
    --top-k 2 --weaviate-cloud
pause

# =================================================================
# PART 5: Statistics & Schema (~3 min)
# =================================================================
page "Part 5: Statistics & Schema"

run "$WEAVE" stats $DEMO_COL --weaviate-cloud
pause

run "$WEAVE" cols show $DEMO_COL \
    --schema --weaviate-cloud
pause

# =================================================================
# PART 6: Backup & Restore (~5 min)
# =================================================================
page "Part 6: Backup & Restore"

echo "Create a portable backup (cross-VDB compatible):"
run "$WEAVE" backup create $DEMO_COL \
    --output /tmp/alliance-demo.weavebak --weaviate-cloud
pause

echo "Inspect the backup:"
run "$WEAVE" backup list /tmp/
pause

run "$WEAVE" backup validate /tmp/alliance-demo.weavebak
pause

# =================================================================
# PART 7: Batch Ingestion & Pipeline (~5 min)
# =================================================================
page "Part 7: Batch Ingestion"

echo "Batch ingest an entire directory of docs:"
run "$WEAVE" pipeline ingest "$PROJECT_ROOT/docs/vdbs/redis/" \
    --collection $DEMO_COL --weaviate-cloud || true
pause

echo "Updated document count:"
run "$WEAVE" cols count --weaviate-cloud
pause

# =================================================================
# PART 8: Multi-VDB Support (~3 min)
# =================================================================
page "Part 8: 11 Vector Databases Supported"

echo "Weave CLI supports 11 VDBs with a unified interface."
echo "Same commands work across all providers:"
echo ""
echo "  Stable:       Weaviate, Qdrant, Milvus, Chroma,"
echo "                Supabase, Neo4j, MongoDB, OpenSearch"
echo "  Beta:         Pinecone, Elasticsearch"
echo "  Experimental: Redis (NEW in v0.12.0)"
echo ""
echo "Switch databases with a single flag:"
echo "  --weaviate-cloud, --milvus-local, --qdrant-local, etc."
echo ""
echo "Or set in config.yaml and forget about it."
pause

# =================================================================
# PART 9: Embeddings (~3 min)
# =================================================================
page "Part 9: Embedding Models"

run "$WEAVE" embeddings list
pause

echo "Supports OpenAI, sentence-transformers, Ollama, Cohere,"
echo "Voyage AI — including 100% free local models."
echo ""
echo "OSS embeddings scored 11% higher than OpenAI on our"
echo "production dataset, ran 240x faster, and cost nothing."
pause

# =================================================================
# PART 10: Evals & Agents (overview) (~3 min)
# =================================================================
page "Part 10: Evaluations & Agents"

echo "Built-in evaluation harness with LLM-as-judge:"
echo ""
echo "  weave eval run --agent rag-agent --dataset baseline"
echo "  weave eval benchmark --agents rag,qa --dataset baseline"
echo ""
echo "10 built-in agents: QueryAgent, PlanningAgent, WeaveAgent,"
echo "BashAgent, RAGAgent, ChunkingAgent, SchemaAgent,"
echo "OutputAgent, ReportAgent, EvalAgent"
echo ""
echo "Custom agents via YAML config."
echo ""
echo "Opik integration for production observability:"
echo "  traces, cost tracking, experiment comparison."
pause

# =================================================================
# PART 11: REPL (~2 min)
# =================================================================
page "Part 11: Interactive AI REPL"

echo "Start the REPL with just 'weave' (no args)."
echo "Natural language interface to all operations:"
echo ""
echo "  > list my collections"
echo "  > how many documents in AllianceDemo?"
echo "  > search AllianceDemo for vector database architecture"
echo ""
echo "(Demo live if time permits)"
pause

# =================================================================
# CLEANUP
# =================================================================
page "Cleanup"

echo "Delete demo collections:"
run "$WEAVE" docs da $DEMO_COL --no-confirm --weaviate-cloud || true
run "$WEAVE" docs da $DEMO_IMG_COL --no-confirm --weaviate-cloud || true
run "$WEAVE" cols ds $DEMO_COL --no-confirm --weaviate-cloud || true
run "$WEAVE" cols ds $DEMO_IMG_COL --no-confirm --weaviate-cloud || true
echo ""
echo "Backup still available at: /tmp/alliance-demo.weavebak"
pause

# =================================================================
page "Thank You!"
echo ""
echo "  GitHub:   github.com/maximilien/weave-cli"
echo "  Install:  brew install Maximilien-ai/weave-cli/weave-cli"
echo "  Docs:     weave -h | weave doctor"
echo ""
echo "  11 vector databases, 1 CLI."
echo ""
