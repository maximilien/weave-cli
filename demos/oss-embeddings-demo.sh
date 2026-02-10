#!/bin/bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2026 dr.max

# OSS Embeddings Demo - sentence-transformers and Ollama
# Demonstrates batch re-embedding with free, local embedding providers

set -e  # Exit on error

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Configuration
WEAVE_BIN="${WEAVE_BIN:-./bin/weave}"
TEST_COLLECTION="OSS_Demo_Collection"
VDB_FLAG="--milvus-local"  # Change to your VDB

echo -e "${BLUE}======================================${NC}"
echo -e "${BLUE}  OSS Embeddings Demo${NC}"
echo -e "${BLUE}  weave-cli v0.9.19+${NC}"
echo -e "${BLUE}======================================${NC}"
echo ""

# Step 1: Check prerequisites
echo -e "${BLUE}[1/8]${NC} Checking prerequisites..."
echo ""

# Check weave binary
if [ ! -f "$WEAVE_BIN" ]; then
    echo -e "${RED}Error: weave binary not found at $WEAVE_BIN${NC}"
    echo "Please build first: ./build.sh"
    exit 1
fi

# Check Python
if ! command -v python3 &> /dev/null; then
    echo -e "${RED}Error: python3 not found${NC}"
    echo "Please install Python 3.8+"
    exit 1
fi

echo -e "${GREEN}✓ weave binary found${NC}"
echo -e "${GREEN}✓ python3 found: $(python3 --version)${NC}"

# Check sentence-transformers
if ! python3 -c "import sentence_transformers" 2>/dev/null; then
    echo -e "${YELLOW}⚠ sentence-transformers not installed${NC}"
    echo ""
    echo "Would you like to install it now? (y/n)"
    read -r response
    if [[ "$response" =~ ^[Yy]$ ]]; then
        echo "Installing sentence-transformers..."
        pip install sentence-transformers
        echo -e "${GREEN}✓ sentence-transformers installed${NC}"
    else
        echo -e "${RED}Skipping sentence-transformers demo${NC}"
        SKIP_SENTRANS=true
    fi
else
    echo -e "${GREEN}✓ sentence-transformers installed${NC}"
fi

# Check Ollama (optional)
if command -v ollama &> /dev/null; then
    echo -e "${GREEN}✓ Ollama installed${NC}"
    OLLAMA_AVAILABLE=true
else
    echo -e "${YELLOW}⚠ Ollama not installed (optional)${NC}"
    echo "  Install from: https://ollama.ai"
    OLLAMA_AVAILABLE=false
fi

echo ""
sleep 2

# Step 2: Create test collection with OpenAI
echo -e "${BLUE}[2/8]${NC} Creating test collection with OpenAI embeddings..."
echo ""

# Check if collection exists
if $WEAVE_BIN cols ls $VDB_FLAG | grep -q "$TEST_COLLECTION"; then
    echo -e "${YELLOW}Collection $TEST_COLLECTION already exists${NC}"
    echo "Delete it? (y/n)"
    read -r response
    if [[ "$response" =~ ^[Yy]$ ]]; then
        $WEAVE_BIN collection delete "$TEST_COLLECTION" $VDB_FLAG
        echo -e "${GREEN}✓ Deleted existing collection${NC}"
    else
        echo "Using existing collection"
    fi
fi

# Create collection if it doesn't exist
if ! $WEAVE_BIN cols ls $VDB_FLAG | grep -q "$TEST_COLLECTION"; then
    $WEAVE_BIN collection create "$TEST_COLLECTION" \
        --embedding text-embedding-3-small \
        $VDB_FLAG

    echo -e "${GREEN}✓ Created collection: $TEST_COLLECTION${NC}"

    # Add sample documents
    echo "Adding sample documents..."

    # Create temp files
    cat > /tmp/doc1.txt << 'EOF'
Vintage Leica M3 camera from 1954 in excellent condition.
This classic rangefinder camera is highly sought after by collectors.
Includes original leather case and lens cap.
EOF

    cat > /tmp/doc2.txt << 'EOF'
Nikon F series professional SLR cameras.
Known for their durability and reliability in photojournalism.
Used by photographers worldwide since 1959.
EOF

    cat > /tmp/doc3.txt << 'EOF'
Hasselblad 500C medium format camera system.
Famous for being used in NASA space missions.
Produces stunning 6x6cm square format images.
EOF

    $WEAVE_BIN docs create "$TEST_COLLECTION" /tmp/doc1.txt $VDB_FLAG
    $WEAVE_BIN docs create "$TEST_COLLECTION" /tmp/doc2.txt $VDB_FLAG
    $WEAVE_BIN docs create "$TEST_COLLECTION" /tmp/doc3.txt $VDB_FLAG

    echo -e "${GREEN}✓ Added 3 sample documents${NC}"
else
    echo -e "${GREEN}✓ Using existing collection${NC}"
fi

echo ""
sleep 2

# Step 3: Query with OpenAI embeddings (baseline)
echo -e "${BLUE}[3/8]${NC} Testing search with OpenAI embeddings (baseline)..."
echo ""

$WEAVE_BIN collection search "$TEST_COLLECTION" \
    --query "vintage rangefinder cameras" \
    --top-k 3 \
    $VDB_FLAG

echo ""
echo -e "${GREEN}✓ Baseline search completed${NC}"
echo ""
sleep 2

# Step 4: Re-embed with sentence-transformers
if [ "$SKIP_SENTRANS" != true ]; then
    echo -e "${BLUE}[4/8]${NC} Re-embedding with sentence-transformers..."
    echo ""
    echo "This will:"
    echo "  1. Download model (first time only, ~400MB)"
    echo "  2. Generate embeddings locally (no API calls)"
    echo "  3. Create new collection with OSS embeddings"
    echo ""

    SENTRANS_COLLECTION="${TEST_COLLECTION}_OSS"

    # Delete if exists
    if $WEAVE_BIN cols ls $VDB_FLAG | grep -q "$SENTRANS_COLLECTION"; then
        $WEAVE_BIN collection delete "$SENTRANS_COLLECTION" $VDB_FLAG
    fi

    $WEAVE_BIN collection reembed "$TEST_COLLECTION" \
        --new-embedding sentence-transformers/all-mpnet-base-v2 \
        --output "$SENTRANS_COLLECTION" \
        $VDB_FLAG

    echo ""
    echo -e "${GREEN}✓ Re-embedded with sentence-transformers${NC}"
    echo ""
    sleep 2

    # Step 5: Compare results
    echo -e "${BLUE}[5/8]${NC} Comparing OpenAI vs sentence-transformers..."
    echo ""

    $WEAVE_BIN collection compare \
        "$TEST_COLLECTION" \
        "$SENTRANS_COLLECTION" \
        --query "vintage rangefinder cameras" \
        --query "professional SLR cameras" \
        --query "medium format photography" \
        --top-k 3 \
        --report /tmp/oss-comparison.md \
        $VDB_FLAG

    echo ""
    echo -e "${GREEN}✓ Comparison complete${NC}"
    echo ""
    echo "Report saved to: /tmp/oss-comparison.md"
    echo ""
    echo "Preview:"
    head -30 /tmp/oss-comparison.md
    echo ""
    sleep 3
else
    echo -e "${YELLOW}[4/8] Skipping sentence-transformers (not installed)${NC}"
    echo -e "${YELLOW}[5/8] Skipping comparison${NC}"
    echo ""
fi

# Step 6: Ollama demo (optional)
if [ "$OLLAMA_AVAILABLE" = true ]; then
    echo -e "${BLUE}[6/8]${NC} Testing Ollama embeddings (optional)..."
    echo ""
    echo "Check if nomic-embed-text model is available..."

    if ollama list | grep -q "nomic-embed-text"; then
        echo -e "${GREEN}✓ nomic-embed-text found${NC}"

        OLLAMA_COLLECTION="${TEST_COLLECTION}_Ollama"

        # Delete if exists
        if $WEAVE_BIN cols ls $VDB_FLAG | grep -q "$OLLAMA_COLLECTION"; then
            $WEAVE_BIN collection delete "$OLLAMA_COLLECTION" $VDB_FLAG
        fi

        $WEAVE_BIN collection reembed "$TEST_COLLECTION" \
            --new-embedding nomic-embed-text \
            --output "$OLLAMA_COLLECTION" \
            $VDB_FLAG

        echo ""
        echo -e "${GREEN}✓ Re-embedded with Ollama${NC}"
        echo ""
    else
        echo -e "${YELLOW}⚠ nomic-embed-text not found${NC}"
        echo "To install: ollama pull nomic-embed-text"
        echo ""
    fi
else
    echo -e "${YELLOW}[6/8] Skipping Ollama demo (not installed)${NC}"
    echo ""
fi

# Step 7: Show cost comparison
echo -e "${BLUE}[7/8]${NC} Cost Comparison"
echo ""
echo -e "${BLUE}Provider             | Cost per 1M tokens | 3 docs cost${NC}"
echo "--------------------------------------------------------"
echo -e "OpenAI               | \$0.02              | ~\$0.0001"
echo -e "${GREEN}sentence-transformers | \$0.00 (FREE)      | \$0.00${NC}"
echo -e "${GREEN}Ollama               | \$0.00 (FREE)      | \$0.00${NC}"
echo ""
echo "💰 Annual savings (4 re-embeddings × 1000 docs): ~\$0.80"
echo ""
sleep 3

# Step 8: Summary
echo -e "${BLUE}[8/8]${NC} Summary"
echo ""
echo -e "${GREEN}✓ Demo Complete!${NC}"
echo ""
echo "What you learned:"
echo "  • How to install OSS embedding providers"
echo "  • How to re-embed collections with OSS models"
echo "  • How to compare OpenAI vs OSS quality"
echo "  • Cost savings: 100% free local embeddings"
echo ""
echo "Next steps:"
echo "  1. Review comparison report: /tmp/oss-comparison.md"
echo "  2. Test with your own collections"
echo "  3. Choose best provider for your use case"
echo ""
echo "Resources:"
echo "  • Guide: docs/guides/OSS_EMBEDDING_TESTING_TIPS.md"
echo "  • CHANGELOG: v0.9.19 release notes"
echo "  • GitHub: github.com/maximilien/weave-cli"
echo ""
echo -e "${BLUE}======================================${NC}"
echo -e "${BLUE}  Thank you for trying OSS embeddings!${NC}"
echo -e "${BLUE}======================================${NC}"
