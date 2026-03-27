#!/bin/bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2026 dr.max

# =================================================================
# AI Alliance Demo — Weave CLI v0.12.2
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
DIM='\033[2m'
BOLD='\033[1m'
NC='\033[0m'

# Step tracking
CURRENT_STEP=0
TOTAL_STEPS=19
SKIP_SECTION=false

page() {
    CURRENT_STEP=$((CURRENT_STEP + 1))
    SKIP_SECTION=false
    echo ""
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BOLD}  [$CURRENT_STEP/$TOTAL_STEPS] $1${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
}

# Format a command for display (quote args with spaces)
_fmt_cmd() {
    local display=""
    for arg in "$@"; do
        if [[ "$arg" == *" "* ]]; then
            display="$display \"$arg\""
        else
            display="$display $arg"
        fi
    done
    echo "$display"
}

run() {
    if $SKIP_SECTION; then return 0; fi
    _LAST_CMD=("$@")
    local display
    display=$(_fmt_cmd "$@")
    echo -e "${GREEN}\$${display}${NC}"
    "$@"
    echo ""
}

# Interactive pause with controls:
#   Enter  — continue to next step
#   Esc    — skip remaining commands in this section
#   e      — edit and re-run the last command
pause() {
    if $SKIP_SECTION; then return 0; fi
    echo -e "${DIM}  [$CURRENT_STEP/$TOTAL_STEPS] Enter=continue | Esc=skip section | e=edit | q=quit${NC}"
    while true; do
        IFS= read -r -s -n1 key
        case "$key" in
            "") # Enter
                clear
                return 0
                ;;
            $'\x1b') # Esc
                SKIP_SECTION=true
                clear
                return 0
                ;;
            e|E)
                _edit_last
                ;;
            q|Q)
                echo ""
                echo "Demo ended early."
                exit 0
                ;;
        esac
    done
}

# Let the user edit and re-run the last displayed command
_LAST_CMD=()
_edit_last() {
    if [ ${#_LAST_CMD[@]} -eq 0 ]; then
        echo "  (no command to edit)"
        return
    fi
    local orig
    orig=$(_fmt_cmd "${_LAST_CMD[@]}")
    echo -ne "${YELLOW}  Edit command: ${NC}"
    read -r -e -i "$orig" edited
    if [ -n "$edited" ]; then
        echo -e "${GREEN}\$${edited}${NC}"
        eval "$edited" || true
        echo ""
    fi
}

DEMO_COL="AllianceDemo"
DEMO_IMG_COL="AllianceDemoImages"
DOCS="$SHARED_DIR/docs"

# =================================================================
page "AI Alliance Demo — Weave CLI v0.12.2"
echo "AI-powered CLI for managing 11+ vector databases."
echo "Built in Go. Single binary. Open source."
echo ""
echo "Today: install -> diagnose -> ingest -> search -> agents"
echo "       -> evals -> backup, across multiple VDBs, live."
pause

# =================================================================
# PART 1: Installation & Setup
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
# PART 2: weave doctor
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
# PART 3: Configuration & Health
# =================================================================
page "Part 3: Configuration & VDB Health"

run "$WEAVE" config list
pause

run "$WEAVE" health check --cloud
pause

run "$WEAVE" vdb list
pause

# =================================================================
# PART 4: Collections & Documents
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
# PART 5: Semantic Search
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
# PART 6: Statistics & Schema
# =================================================================
page "Part 6: Statistics & Schema"

run "$WEAVE" stats $DEMO_COL --milvus-local
pause

run "$WEAVE" cols show $DEMO_COL \
    --schema --weaviate-cloud
pause

# =================================================================
# PART 7: Backup & Restore
# =================================================================
page "Part 7: Backup & Restore"

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
# PART 8: Batch Ingestion
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
# PART 9: Multi-VDB Support
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
pause

# =================================================================
# PART 10: Embeddings
# =================================================================
page "Part 10: Embedding Models"

run "$WEAVE" embeddings list
pause

# =================================================================
# PART 11: Agents
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
# PART 12: Evaluations
# =================================================================
page "Part 12: Evaluations"

echo "Run evaluation against a dataset:"
run "$WEAVE" eval run --agent rag-agent --dataset baseline
pause

echo "Benchmark multiple agents side-by-side:"
run "$WEAVE" eval benchmark --agents rag-agent,qa-agent --dataset baseline
pause

# =================================================================
# PART 13: REPL
# =================================================================
page "Part 13: Interactive AI REPL"

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
run "$WEAVE" cols del $DEMO_COL --force --milvus-local || true
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
