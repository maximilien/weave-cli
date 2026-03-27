#!/bin/bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2026 dr.max

# =================================================================
# Opik Demo — Weave CLI v0.12.3
# Duration: ~45 minutes (leave 15 min for Q&A)
#
# Uses config.yaml + .env in this directory.
# VDBs: weaviate-cloud (default) + milvus-local.
#
# Usage: Run manually command by command, or press Enter between
#        pages if running the script directly.
# =================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SHARED_DIR="$SCRIPT_DIR/../shared"

# Use local bin/ if present, otherwise system weave
if [ -x "$SCRIPT_DIR/bin/weave" ]; then
    WEAVE="$SCRIPT_DIR/bin/weave"
else
    WEAVE="weave"
fi

# Run from demo directory so config.yaml and .env are picked up
cd "$SCRIPT_DIR"

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
    "$@"
    echo ""
}

pause() {
    echo -e "${YELLOW}[Press Enter]${NC}"
    read -r
    clear
}

DEMO_COL="OpikDemo"
DEMO_IMG_COL="OpikDemoImages"
DOCS="$SHARED_DIR/docs"

# =================================================================
page "Opik Demo — Weave CLI v0.12.3"
echo "AI-powered CLI for managing 11+ vector databases."
echo "Built in Go. Single binary. Open source."
echo ""
echo "Today: install -> diagnose -> ingest -> search -> agents"
echo "       -> evals -> backup, across multiple VDBs, live."
pause

# =================================================================
# PART 1: Installation & Setup (~3 min)
# =================================================================
page "Part 1: Installation & Setup"

echo "Install via Homebrew:"
echo ""
echo "  brew install Maximilien-ai/weave-cli/weave-cli"
echo ""
echo "(Already installed — let's verify)"
echo ""

run "$WEAVE" --version
pause

# =================================================================
# PART 2: weave doctor (~3 min)
# =================================================================
page "Part 2: weave doctor — Unified Diagnostics"

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
# PART 3: Configuration & Health (~3 min)
# =================================================================
page "Part 3: Configuration & VDB Health"

run "$WEAVE" config list
pause

run "$WEAVE" health check --cloud
pause

run "$WEAVE" vdb list
pause

# =================================================================
# PART 4: Collections & Documents (~10 min)
# =================================================================
page "Part 4: Collections & Documents"

echo "List existing collections:"
run "$WEAVE" cols ls
pause

echo "Create text collection on Milvus (local):"
run "$WEAVE" cols create $DEMO_COL \
    --embedding text-embedding-3-small --milvus-local || true
pause

echo "Create image collection on Weaviate (cloud):"
run "$WEAVE" cols create $DEMO_IMG_COL \
    --image --weaviate-cloud || true
pause

# --- Ingest text documents (Milvus) ---
page "Ingest Text Documents (Milvus Local)"

echo "Ingest the architecture doc:"
run "$WEAVE" docs create $DEMO_COL \
    "$DOCS/ARCHITECTURE.md" --milvus-local
pause

echo "Ingest the README:"
run "$WEAVE" docs create $DEMO_COL \
    "$DOCS/README.md" --milvus-local
pause

echo "List documents on Milvus:"
run "$WEAVE" docs ls $DEMO_COL --milvus-local
pause

# --- Ingest PDF (Weaviate) ---
page "Ingest PDF (Weaviate Cloud)"

echo "Ingest a PDF (text + image extraction):"
run "$WEAVE" docs create $DEMO_COL \
    "$DOCS/ragme-io.pdf" \
    --image-col $DEMO_IMG_COL --weaviate-cloud || true
pause

# --- Ingest images (Weaviate) ---
page "Ingest Images (Weaviate Cloud)"

echo "Add test image:"
run "$WEAVE" docs create $DEMO_IMG_COL \
    "$DOCS/dog.png" --weaviate-cloud || true
pause

run "$WEAVE" docs ls $DEMO_IMG_COL --weaviate-cloud
pause

# =================================================================
# PART 5: Semantic Search (~5 min)
# =================================================================
page "Part 5: Semantic Search (RAG)"

echo "Search text on Milvus local:"
echo ""

run "$WEAVE" cols query $DEMO_COL \
    "how does the vector database abstraction work?" \
    --top-k 3 --milvus-local
pause

run "$WEAVE" cols query $DEMO_COL \
    "what embedding models are supported?" \
    --top-k 3 --milvus-local
pause

echo "Same query on Weaviate cloud (cross-VDB — same syntax!):"
run "$WEAVE" cols query $DEMO_COL \
    "kubernetes deployment" \
    --top-k 2 --weaviate-cloud
pause

echo "Image search on Weaviate cloud:"
run "$WEAVE" cols query $DEMO_IMG_COL \
    "dog" --top-k 2 --weaviate-cloud
pause

# =================================================================
# PART 6: Statistics & Schema (~3 min)
# =================================================================
page "Part 6: Statistics & Schema"

run "$WEAVE" stats $DEMO_COL --milvus-local
pause

run "$WEAVE" cols show $DEMO_COL \
    --schema --weaviate-cloud
pause

# =================================================================
# PART 7: Backup & Restore (~5 min)
# =================================================================
page "Part 7: Backup & Restore"

echo "Create a portable backup (cross-VDB compatible):"
run "$WEAVE" backup create $DEMO_COL \
    --output /tmp/opik-demo.weavebak --weaviate-cloud
pause

echo "Inspect the backup:"
run "$WEAVE" backup list /tmp/
pause

run "$WEAVE" backup validate /tmp/opik-demo.weavebak
pause

# =================================================================
# PART 8: Batch Ingestion (~5 min)
# =================================================================
page "Part 8: Batch Ingestion"

echo "Batch ingest a directory of docs:"
run "$WEAVE" pipeline ingest "$DOCS/vdbs/redis/" \
    --collection $DEMO_COL --weaviate-cloud || true
pause

echo "Updated document count:"
run "$WEAVE" cols count --weaviate-cloud
pause

# =================================================================
# PART 9: Multi-VDB Support (~3 min)
# =================================================================
page "Part 9: 11 Vector Databases Supported"

echo "Weave CLI supports 11 VDBs with a unified interface."
echo "Same commands work across all providers:"
echo ""
echo "  Stable:       Weaviate, Qdrant, Milvus, Chroma,"
echo "                Supabase, Neo4j, MongoDB, OpenSearch"
echo "  Beta:         Pinecone, Elasticsearch"
echo "  Experimental: Redis"
echo ""
echo "Switch databases with a single flag:"
echo "  --weaviate-cloud, --milvus-local, --qdrant-local, etc."
echo ""
echo "Or set in config.yaml and forget about it."
pause

# =================================================================
# PART 10: Embeddings (~3 min)
# =================================================================
page "Part 10: Embedding Models"

run "$WEAVE" embeddings list
pause

echo "Supports OpenAI, sentence-transformers, Ollama, Cohere,"
echo "Voyage AI — including 100% free local models."
echo ""
echo "OSS embeddings scored 11% higher than OpenAI on our"
echo "production dataset, ran 240x faster, and cost nothing."
pause

# =================================================================
# PART 11: Agents (~5 min)
# =================================================================
page "Part 11: Agents"

echo "List available agents:"
run "$WEAVE" agents list
pause

echo "Show agent details:"
run "$WEAVE" agents show rag-agent
pause

echo "RAG agent query:"
run "$WEAVE" cols query $DEMO_COL \
    "explain the vector database abstraction layer" \
    --agent rag-agent --milvus-local
pause

echo "Summarizer agent query:"
run "$WEAVE" cols query $DEMO_COL \
    "summarize the architecture document" \
    --agent summarize-agent --milvus-local
pause

echo "QA agent query:"
run "$WEAVE" cols query $DEMO_COL \
    "what embedding models are supported?" \
    --agent qa-agent --milvus-local
pause

# =================================================================
# PART 12: Evaluations (~3 min)
# =================================================================
page "Part 12: Evaluations"

echo "Built-in evaluation harness with LLM-as-judge:"
echo ""

echo "Run evaluation against a dataset:"
run "$WEAVE" eval run --agent rag-agent --dataset baseline
pause

echo "Benchmark multiple agents side-by-side:"
run "$WEAVE" eval benchmark --agents rag-agent,qa-agent --dataset baseline
pause

echo "Opik integration for production observability:"
echo "  traces, cost tracking, experiment comparison."
pause

# =================================================================
# PART 13: REPL (~2 min)
# =================================================================
page "Part 13: Interactive AI REPL"

echo "Start the REPL with just 'weave' (no args)."
echo "Natural language interface to all operations:"
echo ""
echo "  > list my collections"
echo "  > how many documents in OpikDemo?"
echo "  > search OpikDemo for vector database architecture"
echo ""
echo "(Demo live if time permits)"
pause

# =================================================================
# CLEANUP
# =================================================================
page "Cleanup"

echo "Delete demo collections:"
run "$WEAVE" cols del $DEMO_COL --force --milvus-local || true
run "$WEAVE" docs da $DEMO_COL --no-confirm --weaviate-cloud || true
run "$WEAVE" docs da $DEMO_IMG_COL --no-confirm --weaviate-cloud || true
run "$WEAVE" cols ds $DEMO_COL --no-confirm --weaviate-cloud || true
run "$WEAVE" cols ds $DEMO_IMG_COL --no-confirm --weaviate-cloud || true
echo ""
echo "Backup still available at: /tmp/opik-demo.weavebak"
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
