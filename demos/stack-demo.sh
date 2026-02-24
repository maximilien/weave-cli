#!/bin/bash
# Weave Stack Demo - Phase 1 (v0.10.0)
#
# This demo showcases the complete weave stack workflow:
# 1. Initialize stack configuration
# 2. Validate configuration
# 3. Deploy to Kind cluster
# 4. View status and logs
# 5. Access services via port forwarding
# 6. Cleanup
#
# Prerequisites:
# - kubectl, helm, kind installed
# - Docker or Podman running

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Demo directory
DEMO_DIR="/tmp/weave-stack-demo-$$"

# Function to print section headers
section() {
    echo ""
    echo -e "${CYAN}═══════════════════════════════════════════════════════${NC}"
    echo -e "${CYAN}$1${NC}"
    echo -e "${CYAN}═══════════════════════════════════════════════════════${NC}"
    echo ""
}

# Function to print commands before running
run_cmd() {
    echo -e "${YELLOW}→ $1${NC}"
    eval "$1"
    echo ""
}

# Function to pause for user
pause() {
    echo -e "${GREEN}Press Enter to continue...${NC}"
    read -r
}

# Cleanup function
cleanup() {
    section "Cleanup"
    echo "Removing demo directory: $DEMO_DIR"
    cd /
    rm -rf "$DEMO_DIR"
    echo "Cleanup complete"
}

# Trap cleanup on exit
trap cleanup EXIT

# Introduction
clear
echo -e "${GREEN}"
cat << "EOF"
╦ ╦┌─┐┌─┐┬  ┬┌─┐  ╔═╗┌┬┐┌─┐┌─┐┬┌─  ╔╦╗┌─┐┌┬┐┌─┐
║║║├┤ ├─┤└┐┌┘├┤   ╚═╗ │ ├─┤│  ├┴┐   ║║├┤ ││││ │
╚╩╝└─┘┴ ┴ └┘ └─┘  ╚═╝ ┴ ┴ ┴└─┘┴ ┴  ═╩╝└─┘┴ ┴└─┘

Kubernetes RAG Stack Management - Phase 1 (v0.10.0)
EOF
echo -e "${NC}"

echo "This demo will:"
echo "  1. Initialize a new RAG stack"
echo "  2. Deploy to local Kubernetes (Kind)"
echo "  3. Show stack management commands"
echo "  4. Cleanup everything"
echo ""
echo -e "${YELLOW}Prerequisites:${NC}"
echo "  - kubectl, helm, kind installed"
echo "  - Docker or Podman running"
echo ""

pause

# Check prerequisites
section "1. Checking Prerequisites"

run_cmd "kubectl version --client --short 2>/dev/null || kubectl version --client"
run_cmd "helm version --short"
run_cmd "kind version"

echo -e "${GREEN}✓ All prerequisites met${NC}"
pause

# Create demo directory
section "2. Creating Demo Directory"

run_cmd "mkdir -p $DEMO_DIR"
run_cmd "cd $DEMO_DIR && pwd"

pause

# Initialize stack - Quickstart
section "3. Initialize Stack (Quickstart Template)"

echo "The quickstart template creates a minimal RAG stack with:"
echo "  - Milvus vector database"
echo "  - 1 text collection"
echo "  - OpenAI embeddings"
echo ""

run_cmd "weave stack init --template quickstart --runtime kind --quiet-config"

echo "Files created:"
run_cmd "ls -la"

pause

# Show configuration
section "4. Review Stack Configuration"

echo "weave-stack.yaml defines your complete RAG stack:"
echo ""

run_cmd "cat weave-stack.yaml | head -50"

echo ""
echo "(Configuration continues...)"
pause

# Validate configuration
section "5. Validate Configuration"

run_cmd "weave stack validate --quiet-config"

pause

# Show other templates
section "6. Other Available Templates"

echo "Weave Stack provides 4 templates:"
echo ""
echo -e "${GREEN}quickstart${NC}  - Minimal RAG stack (current)"
echo -e "${GREEN}production${NC}  - Full stack with dashboard & evaluations"
echo -e "${GREEN}multimodal${NC}  - Text + image collections"
echo -e "${GREEN}oss${NC}         - 100% open source (Ollama + sentence-transformers)"
echo ""
echo "Try them with:"
echo "  weave stack init --template production --runtime kind"
echo ""

pause

# Deploy stack
section "7. Deploy to Kubernetes (Kind)"

echo -e "${YELLOW}This will take 2-3 minutes...${NC}"
echo ""
echo "The deployment will:"
echo "  1. Create Kind cluster"
echo "  2. Generate Helm charts"
echo "  3. Deploy Milvus to Kubernetes"
echo "  4. Wait for pods to be ready"
echo ""

pause

run_cmd "weave stack up --runtime kind --quiet-config 2>&1 | tail -20"

echo ""
echo -e "${GREEN}✓ Stack deployed successfully${NC}"
pause

# Check status
section "8. Check Stack Status"

run_cmd "weave stack status --quiet-config"

pause

# Show kubectl integration
section "9. kubectl Integration"

echo "weave stack kubectl auto-injects the cluster context:"
echo ""

run_cmd "weave stack kubectl -- get pods"

echo ""
run_cmd "weave stack kubectl -- get services"

pause

# Show logs
section "10. View Component Logs"

echo "View last 20 lines of Milvus logs:"
echo ""

run_cmd "weave stack logs milvus --tail 20 --quiet-config 2>&1 | head -30"

echo ""
echo "(Use --follow to stream logs in real-time)"
pause

# Port forwarding demo
section "11. Port Forwarding"

echo "Forward Milvus gRPC port (19530) to localhost:"
echo ""
echo -e "${YELLOW}weave stack port-forward milvus 19530:19530${NC}"
echo ""
echo "This allows you to connect to Milvus at localhost:19530"
echo "from your local machine."
echo ""
echo "(Skipping actual port-forward in demo - use Ctrl+C to stop)"

pause

# Show cluster info
section "12. Cluster Information"

echo "Cluster state is saved in .weave-state/:"
echo ""

run_cmd "cat .weave-state/cluster.json"

pause

# Stop stack
section "13. Stop and Cleanup Stack"

echo "This will:"
echo "  1. Uninstall Helm release"
echo "  2. Delete Kind cluster"
echo "  3. Clean up state files"
echo ""

pause

run_cmd "weave stack down --quiet-config"

echo ""
echo -e "${GREEN}✓ Stack stopped and cleaned up${NC}"
pause

# Summary
section "Demo Complete!"

echo -e "${GREEN}Summary:${NC}"
echo ""
echo "You learned how to:"
echo "  ✓ Initialize a RAG stack with templates"
echo "  ✓ Deploy to local Kubernetes (Kind)"
echo "  ✓ Check stack status and logs"
echo "  ✓ Use kubectl integration"
echo "  ✓ Forward ports to services"
echo "  ✓ Stop and cleanup stack"
echo ""
echo -e "${CYAN}Next Steps:${NC}"
echo "  1. Try other templates: production, multimodal, oss"
echo "  2. Customize weave-stack.yaml for your use case"
echo "  3. Add data sources and collections"
echo "  4. Deploy dashboard for RAG interface"
echo ""
echo -e "${CYAN}Resources:${NC}"
echo "  - Quick Start Guide: docs/guides/WEAVE_STACK_QUICKSTART.md"
echo "  - Design Doc: docs/planning/WEAVE_STACK_DESIGN.md"
echo "  - Integration Tests: test-stack-basic.sh"
echo ""
echo -e "${GREEN}Thank you for trying Weave Stack!${NC}"
echo ""
