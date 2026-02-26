#!/bin/bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2025 dr.max

# Client0 Demo Script - 5-Minute RAG Workflow
# Version: v0.10.2
# Purpose: Demonstrate production-ready weave stack

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Demo settings
DEMO_DIR="/tmp/client0-demo"
COLLECTION="ClientDocuments"

# Use local weave binary if available, otherwise use installed
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WEAVE_CLI_ROOT="$(dirname "$SCRIPT_DIR")"
if [ -f "$WEAVE_CLI_ROOT/bin/weave" ]; then
    WEAVE="$WEAVE_CLI_ROOT/bin/weave"
    echo -e "${GREEN}Using local weave binary: $WEAVE${NC}"
else
    WEAVE="weave"
    echo -e "${YELLOW}Using installed weave binary${NC}"
fi
echo

echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  Weave Stack Demo for Client0${NC}"
echo -e "${BLUE}  Version: v0.10.2 (Production Ready)${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo

# Check prerequisites
echo -e "${GREEN}▶ Checking prerequisites...${NC}"
command -v kubectl >/dev/null 2>&1 || { echo "kubectl required"; exit 1; }
command -v helm >/dev/null 2>&1 || { echo "helm required"; exit 1; }
command -v kind >/dev/null 2>&1 || { echo "kind required"; exit 1; }

if [ -z "$OPENAI_API_KEY" ]; then
    echo -e "${YELLOW}⚠️  OPENAI_API_KEY not set${NC}"
    echo "Set it with: export OPENAI_API_KEY='your-key-here'"
    exit 1
fi

echo -e "${GREEN}✓ All prerequisites met${NC}"
echo

# Clean up any existing demo
if [ -d "$DEMO_DIR" ]; then
    echo -e "${YELLOW}Cleaning up previous demo...${NC}"
    cd "$DEMO_DIR"
    "$WEAVE" stack down --quiet-config 2>/dev/null || true
    rm -rf "$DEMO_DIR"
fi

# Create demo directory
mkdir -p "$DEMO_DIR"
cd "$DEMO_DIR"

echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  Step 1: Initialize Stack${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo
echo -e "${GREEN}▶ weave stack init --template quickstart --runtime kind${NC}"
echo

"$WEAVE" stack init --template quickstart --runtime kind --quiet-config

echo
echo -e "${GREEN}✓ Stack initialized${NC}"
echo -e "  Created: weave-stack.yaml, kubernetes/, .gitignore"
echo

# Brief pause
sleep 2

echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  Step 2: Deploy to Kubernetes${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo
echo -e "${GREEN}▶ weave stack up --runtime kind${NC}"
echo
echo -e "${YELLOW}This will:${NC}"
echo "  1. Create Kind cluster"
echo "  2. Generate Helm charts"
echo "  3. Deploy Milvus vector database"
echo "  4. Wait for pods to be ready"
echo

"$WEAVE" stack up --runtime kind --quiet-config

echo
echo -e "${GREEN}✓ Stack deployed successfully${NC}"
echo

# Pause
sleep 2

echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  Step 3: Verify Deployment${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo
echo -e "${GREEN}▶ weave stack status${NC}"
echo

"$WEAVE" stack status --quiet-config

echo

# Pause
sleep 2

echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  Step 4: Create Sample Data${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo

mkdir -p data

cat > data/product-guide.txt <<'EOF'
Weave CLI is a production-ready RAG (Retrieval-Augmented Generation) tool.
It provides seamless integration with multiple vector databases including Milvus, Qdrant, Weaviate, and Chroma.
The tool supports batch ingestion with checkpointing and parallel processing.
For local development, use Kind or Minikube. For production, deploy to EKS or GKE.
EOF

cat > data/features.txt <<'EOF'
Key features of Weave Stack:
- Kubernetes-based deployment with Helm charts
- Support for 4 stack templates: quickstart, production, multimodal, and OSS
- PM2 dashboard for monitoring and health checks
- Automatic port forwarding and kubectl passthrough
- Data ingestion with progress tracking
- Integration with OpenAI and Ollama for embeddings
EOF

echo -e "${GREEN}✓ Created sample documents:${NC}"
echo "  - data/product-guide.txt"
echo "  - data/features.txt"
echo

# Pause
sleep 2

echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  Step 5: Setup Port-Forward${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo
echo -e "${GREEN}▶ Setting up port-forward to Milvus...${NC}"
echo

# Start port-forward in background
"$WEAVE" stack port-forward milvus 19530:19530 --quiet-config >/dev/null 2>&1 &
PF_PID=$!

# Wait for port-forward to establish
sleep 3

echo -e "${GREEN}✓ Port-forward established (localhost:19530)${NC}"
echo

# Create .env file with Milvus connection
cat > .env <<EOF
MILVUS_HOST=localhost
MILVUS_PORT=19530
EOF

# Pause
sleep 2

echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  Step 6: Ingest Data${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo
echo -e "${GREEN}▶ weave stack ingest $COLLECTION data/ --vector-db-type milvus-local${NC}"
echo

"$WEAVE" stack ingest "$COLLECTION" data/ --vector-db-type milvus-local --quiet-config

echo
echo -e "${GREEN}✓ Data ingestion complete${NC}"
echo

# Pause
sleep 2

echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  Step 7: Query Your Data${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo
echo -e "${GREEN}▶ Querying: 'RAG tool features'${NC}"
echo

"$WEAVE" cols query "$COLLECTION" "RAG tool features" \
    --vector-db-type milvus-local \
    --quiet-config \
    --no-tips

echo
echo -e "${GREEN}✓ Query complete${NC}"
echo

# Kill port-forward
kill $PF_PID 2>/dev/null || true

# Pause
sleep 2

echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  Step 8: Clean Up${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo
echo -e "${GREEN}▶ weave stack down${NC}"
echo

"$WEAVE" stack down --quiet-config

echo
echo -e "${GREEN}✓ Stack stopped and cleaned up${NC}"
echo

# Final summary
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  Demo Complete! 🎉${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo
echo -e "${GREEN}What you just saw:${NC}"
echo "  ✅ Full RAG workflow in < 5 minutes"
echo "  ✅ Local Kubernetes deployment with Kind"
echo "  ✅ Milvus vector database"
echo "  ✅ Data ingestion with progress tracking"
echo "  ✅ Semantic search queries"
echo "  ✅ Clean teardown"
echo
echo -e "${BLUE}Production Features (v0.10.2):${NC}"
echo "  ✅ Works from any directory"
echo "  ✅ Optimized resource allocation"
echo "  ✅ 4 templates: quickstart, production, multimodal, oss"
echo "  ✅ PM2 dashboard (production template)"
echo "  ✅ kubectl passthrough"
echo "  ✅ Port forwarding"
echo "  ✅ Log streaming"
echo
echo -e "${BLUE}Next Steps:${NC}"
echo "  📖 Read: docs/CLIENT0_GETTING_STARTED.md"
echo "  📖 Read: docs/guides/WEAVE_STACK_QUICKSTART.md"
echo "  🚀 Try: weave stack init --template production"
echo "  🚀 Try: weave stack init --template multimodal"
echo
echo -e "${BLUE}Coming in Phase 2:${NC}"
echo "  🔜 AWS EKS support"
echo "  🔜 GCP GKE support"
echo "  🔜 TLS/SSL certificates"
echo "  🔜 Secrets management"
echo "  🔜 Monitoring & observability"
echo "  🔜 Backup & restore"
echo
echo -e "${GREEN}Thank you for your time!${NC}"
echo
