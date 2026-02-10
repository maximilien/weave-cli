#!/bin/bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2026 dr.max

# Embedding Comparison Demo - 3-Way Performance Analysis
# Compares OpenAI vs sentence-transformers vs Ollama embeddings
# Focus: Quality metrics, performance benchmarks, cost analysis

set -e  # Exit on error

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Configuration
WEAVE_BIN="${WEAVE_BIN:-./bin/weave}"
BASE_COLLECTION="${BASE_COLLECTION:-EmbeddingTest}"
VDB_FLAG="--milvus-local"
REPORT_DIR="/tmp/weave-comparison-$(date +%Y%m%d-%H%M%S)"

# Test queries for comparison
TEST_QUERIES=(
    "vintage Leica cameras from 1950s"
    "professional photography equipment"
    "medium format film cameras"
    "rangefinder camera systems"
)

echo -e "${BOLD}${BLUE}========================================${NC}"
echo -e "${BOLD}${BLUE}  Embedding Comparison Benchmark${NC}"
echo -e "${BOLD}${BLUE}  3-Way Performance Analysis${NC}"
echo -e "${BOLD}${BLUE}========================================${NC}"
echo ""
echo "This demo compares three embedding providers:"
echo "  • OpenAI (Cloud API, baseline)"
echo "  • sentence-transformers (Local Python, OSS)"
echo "  • Ollama (Local HTTP API, OSS)"
echo ""
echo -e "${CYAN}Report directory: $REPORT_DIR${NC}"
mkdir -p "$REPORT_DIR"
echo ""

# Step 1: Prerequisites check
echo -e "${BOLD}${BLUE}[Step 1/7]${NC} Checking prerequisites...${NC}"
echo ""

PROVIDERS_AVAILABLE=()
PROVIDERS_SKIPPED=()

# Check weave
if [ ! -f "$WEAVE_BIN" ]; then
    echo -e "${RED}✗ weave binary not found at $WEAVE_BIN${NC}"
    exit 1
fi
echo -e "${GREEN}✓ weave binary found${NC}"

# Check OpenAI API key
if [ -n "$OPENAI_API_KEY" ]; then
    echo -e "${GREEN}✓ OpenAI API key configured${NC}"
    PROVIDERS_AVAILABLE+=("openai")
else
    echo -e "${YELLOW}⚠ OpenAI API key not set (will skip OpenAI)${NC}"
    PROVIDERS_SKIPPED+=("openai")
fi

# Check sentence-transformers
if python3 -c "import sentence_transformers" 2>/dev/null; then
    echo -e "${GREEN}✓ sentence-transformers installed${NC}"
    PROVIDERS_AVAILABLE+=("sentence-transformers")
else
    echo -e "${YELLOW}⚠ sentence-transformers not installed (will skip)${NC}"
    echo "  Install: pip install sentence-transformers"
    PROVIDERS_SKIPPED+=("sentence-transformers")
fi

# Check Ollama
if command -v ollama &> /dev/null && ollama list | grep -q "nomic-embed-text"; then
    echo -e "${GREEN}✓ Ollama with nomic-embed-text installed${NC}"
    PROVIDERS_AVAILABLE+=("ollama")
else
    echo -e "${YELLOW}⚠ Ollama or nomic-embed-text not found (will skip)${NC}"
    echo "  Install Ollama: https://ollama.ai"
    echo "  Pull model: ollama pull nomic-embed-text"
    PROVIDERS_SKIPPED+=("ollama")
fi

echo ""
echo -e "${CYAN}Providers available: ${#PROVIDERS_AVAILABLE[@]}/3${NC}"

if [ ${#PROVIDERS_AVAILABLE[@]} -lt 2 ]; then
    echo -e "${RED}Error: Need at least 2 providers for comparison${NC}"
    echo "Available: ${PROVIDERS_AVAILABLE[*]}"
    echo "Skipped: ${PROVIDERS_SKIPPED[*]}"
    exit 1
fi

echo ""
sleep 2

# Step 2: Create or verify base collection
echo -e "${BOLD}${BLUE}[Step 2/7]${NC} Preparing base collection...${NC}"
echo ""

if $WEAVE_BIN cols ls $VDB_FLAG | grep -q "^$BASE_COLLECTION$"; then
    echo -e "${CYAN}Using existing collection: $BASE_COLLECTION${NC}"
    DOC_COUNT=$($WEAVE_BIN docs ls "$BASE_COLLECTION" $VDB_FLAG | wc -l)
    echo "Documents: $DOC_COUNT"
else
    echo "Creating base collection with sample documents..."

    $WEAVE_BIN collection create "$BASE_COLLECTION" \
        --embedding text-embedding-3-small \
        $VDB_FLAG

    # Create sample documents
    cat > /tmp/camera1.txt << 'EOF'
Leica M3 Rangefinder Camera (1954-1966)
The Leica M3 is considered one of the finest rangefinder cameras ever made.
Features include a bright 0.91x viewfinder, bayonet M-mount, and exceptional build quality.
Highly sought after by collectors and street photographers.
Price range: $800-2000 depending on condition.
EOF

    cat > /tmp/camera2.txt << 'EOF'
Nikon F Professional SLR Camera (1959-1974)
Revolutionary professional 35mm SLR used by photojournalists worldwide.
Features interchangeable viewfinders, motor drive capability, and legendary durability.
Known for its role in documenting major historical events.
Price range: $200-600 depending on condition and accessories.
EOF

    cat > /tmp/camera3.txt << 'EOF'
Hasselblad 500C Medium Format Camera (1957-1970)
Professional medium format camera system producing 6x6cm square negatives.
Famous for being used on NASA Apollo moon missions.
Features Carl Zeiss lenses and modular design with interchangeable backs.
Price range: $800-1500 depending on condition and lens.
EOF

    cat > /tmp/camera4.txt << 'EOF'
Canon AE-1 35mm SLR Camera (1976-1984)
One of the first microprocessor-controlled SLR cameras.
Known for its affordability and reliability, introduced many to photography.
Features aperture priority auto-exposure and FD lens mount.
Price range: $50-150 depending on condition.
EOF

    cat > /tmp/camera5.txt << 'EOF'
Pentax 67 Medium Format SLR (1969-1989)
Large format SLR producing 6x7cm negatives on 120 film.
Popular with landscape and portrait photographers for exceptional image quality.
Features TTL metering, interchangeable lenses, and robust construction.
Price range: $600-1200 depending on condition and lens.
EOF

    for i in {1..5}; do
        $WEAVE_BIN docs create "$BASE_COLLECTION" "/tmp/camera$i.txt" $VDB_FLAG
    done

    echo -e "${GREEN}✓ Created base collection with 5 documents${NC}"
fi

echo ""
sleep 2

# Step 3: Re-embed with all available providers
echo -e "${BOLD}${BLUE}[Step 3/7]${NC} Re-embedding with all providers...${NC}"
echo ""

COLLECTIONS=()
START_TIMES=()
END_TIMES=()

for provider in "${PROVIDERS_AVAILABLE[@]}"; do
    case $provider in
        "openai")
            COLLECTION="${BASE_COLLECTION}_OpenAI"
            MODEL="text-embedding-3-small"
            echo -e "${CYAN}Re-embedding with OpenAI ($MODEL)...${NC}"
            ;;
        "sentence-transformers")
            COLLECTION="${BASE_COLLECTION}_OSS"
            MODEL="sentence-transformers/all-mpnet-base-v2"
            echo -e "${CYAN}Re-embedding with sentence-transformers...${NC}"
            ;;
        "ollama")
            COLLECTION="${BASE_COLLECTION}_Ollama"
            MODEL="nomic-embed-text"
            echo -e "${CYAN}Re-embedding with Ollama ($MODEL)...${NC}"
            ;;
    esac

    # Delete if exists
    $WEAVE_BIN collection delete "$COLLECTION" $VDB_FLAG 2>/dev/null || true

    # Time the re-embedding
    START_TIME=$(date +%s)
    START_TIMES+=("$START_TIME")

    $WEAVE_BIN collection reembed "$BASE_COLLECTION" \
        --new-embedding "$MODEL" \
        --output "$COLLECTION" \
        $VDB_FLAG

    END_TIME=$(date +%s)
    END_TIMES+=("$END_TIME")

    DURATION=$((END_TIME - START_TIME))
    echo -e "${GREEN}✓ Completed in ${DURATION}s${NC}"
    echo ""

    COLLECTIONS+=("$COLLECTION")
done

echo ""
sleep 2

# Step 4: Run test queries
echo -e "${BOLD}${BLUE}[Step 4/7]${NC} Running test queries...${NC}"
echo ""

for query in "${TEST_QUERIES[@]}"; do
    echo -e "${CYAN}Query: \"$query\"${NC}"

    for i in "${!COLLECTIONS[@]}"; do
        collection="${COLLECTIONS[$i]}"
        provider="${PROVIDERS_AVAILABLE[$i]}"

        echo -e "  ${provider}: Searching..."
        $WEAVE_BIN collection search "$collection" \
            --query "$query" \
            --top-k 3 \
            $VDB_FLAG > /dev/null
    done
    echo ""
done

echo -e "${GREEN}✓ Test queries completed${NC}"
echo ""
sleep 2

# Step 5: Generate comparison report
echo -e "${BOLD}${BLUE}[Step 5/7]${NC} Generating comparison report...${NC}"
echo ""

REPORT_FILE="$REPORT_DIR/comparison-report.md"

# Build collection list for compare command
COMPARE_COLLECTIONS="${COLLECTIONS[*]}"

# Build query arguments
QUERY_ARGS=""
for query in "${TEST_QUERIES[@]}"; do
    QUERY_ARGS="$QUERY_ARGS --query \"$query\""
done

echo "Generating comprehensive comparison report..."
eval $WEAVE_BIN collection compare $COMPARE_COLLECTIONS \
    $QUERY_ARGS \
    --top-k 5 \
    --report "$REPORT_FILE" \
    $VDB_FLAG

echo -e "${GREEN}✓ Report generated: $REPORT_FILE${NC}"
echo ""
sleep 2

# Step 6: Performance analysis
echo -e "${BOLD}${BLUE}[Step 6/7]${NC} Performance Analysis${NC}"
echo ""

PERF_FILE="$REPORT_DIR/performance-metrics.txt"

cat > "$PERF_FILE" << EOF
================================
Embedding Performance Benchmark
================================
Date: $(date)
Base Collection: $BASE_COLLECTION
Documents: $(wc -l < /tmp/camera1.txt 2>/dev/null || echo "5")

Re-embedding Times:
EOF

for i in "${!PROVIDERS_AVAILABLE[@]}"; do
    provider="${PROVIDERS_AVAILABLE[$i]}"
    duration=$((END_TIMES[$i] - START_TIMES[$i]))

    case $provider in
        "openai")
            cat >> "$PERF_FILE" << EOF
  • OpenAI (text-embedding-3-small):
    Duration: ${duration}s
    Dimensions: 1536
    Cost: ~\$0.0001 (for this demo)
EOF
            ;;
        "sentence-transformers")
            cat >> "$PERF_FILE" << EOF
  • sentence-transformers (all-mpnet-base-v2):
    Duration: ${duration}s
    Dimensions: 768
    Cost: \$0 (FREE)
EOF
            ;;
        "ollama")
            cat >> "$PERF_FILE" << EOF
  • Ollama (nomic-embed-text):
    Duration: ${duration}s
    Dimensions: 768
    Cost: \$0 (FREE)
EOF
            ;;
    esac
done

cat >> "$PERF_FILE" << EOF

Quality Metrics:
  See comparison report for detailed relevance scores.

Cost Analysis:
EOF

if [[ " ${PROVIDERS_AVAILABLE[*]} " =~ " openai " ]]; then
    cat >> "$PERF_FILE" << EOF
  • OpenAI: ~\$0.02 per 1M tokens
  • OSS Providers: \$0.00 (100% free)
  • Savings: 100% cost reduction
EOF
else
    cat >> "$PERF_FILE" << EOF
  • OSS Providers: \$0.00 (100% free)
  • No OpenAI baseline for cost comparison
EOF
fi

cat "$PERF_FILE"
echo ""
echo -e "${GREEN}✓ Performance metrics saved: $PERF_FILE${NC}"
echo ""
sleep 2

# Step 7: Summary and recommendations
echo -e "${BOLD}${BLUE}[Step 7/7]${NC} Summary & Recommendations${NC}"
echo ""

echo -e "${BOLD}Results Summary:${NC}"
echo ""
echo "📊 Reports generated in: $REPORT_DIR"
echo "   • comparison-report.md - Quality comparison"
echo "   • performance-metrics.txt - Speed & cost analysis"
echo ""

if [ ${#PROVIDERS_AVAILABLE[@]} -eq 3 ]; then
    echo -e "${GREEN}✓ All 3 providers tested successfully${NC}"
elif [ ${#PROVIDERS_AVAILABLE[@]} -eq 2 ]; then
    echo -e "${YELLOW}⚠ 2 of 3 providers tested${NC}"
    echo "  Skipped: ${PROVIDERS_SKIPPED[*]}"
fi

echo ""
echo -e "${BOLD}Next Steps:${NC}"
echo ""
echo "1. Review comparison report:"
echo "   cat $REPORT_FILE"
echo ""
echo "2. Review performance metrics:"
echo "   cat $PERF_FILE"
echo ""
echo "3. Test with your own collections:"
echo "   weave collection reembed YourCollection --new-embedding MODEL"
echo ""
echo "4. Choose best provider for your use case:"
echo "   • OpenAI: Best quality, costs money"
echo "   • sentence-transformers: 90%+ quality, FREE"
echo "   • Ollama: 90%+ quality, FREE, good for LLM integration"
echo ""

echo -e "${BOLD}Resources:${NC}"
echo "• Testing Guide: docs/guides/OSS_EMBEDDING_TESTING_TIPS.md"
echo "• CHANGELOG: v0.9.19 release notes"
echo "• GitHub: github.com/maximilien/weave-cli"
echo ""

echo -e "${BOLD}${BLUE}========================================${NC}"
echo -e "${BOLD}${BLUE}  Benchmark Complete!${NC}"
echo -e "${BOLD}${BLUE}========================================${NC}"
echo ""

# Offer to open report
echo "Open comparison report now? (y/n)"
read -r response
if [[ "$response" =~ ^[Yy]$ ]]; then
    if command -v less &> /dev/null; then
        less "$REPORT_FILE"
    else
        cat "$REPORT_FILE"
    fi
fi
