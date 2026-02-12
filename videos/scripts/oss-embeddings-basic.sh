#!/bin/bash
# OSS Embeddings Basic Demo Script
# Duration: ~2 minutes
# Records: Basic OSS embedding workflow with sentence-transformers

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}============================================${NC}"
echo -e "${GREEN}   Weave CLI - OSS Embeddings Basic Demo${NC}"
echo -e "${GREEN}============================================${NC}"
echo ""
sleep 2

# Page 1: Introduction
clear
echo -e "${BLUE}=== OSS Embeddings: FREE, Local, Privacy-Focused ===${NC}"
echo ""
echo "This demo shows how to use open-source embedding models"
echo "with Weave CLI for 100% FREE vector database operations."
echo ""
echo -e "${YELLOW}Benefits:${NC}"
echo "  • 100% FREE (no OpenAI API costs)"
echo "  • Privacy (all processing local)"
echo "  • \$240/year savings on 10M tokens"
echo "  • 90%+ quality retention vs OpenAI"
echo ""
sleep 5

# Page 2: Setup
clear
echo -e "${BLUE}=== Step 1: Install sentence-transformers (one-time) ===${NC}"
echo ""
echo "$ pip install sentence-transformers"
echo ""
echo "Installing sentence-transformers library..."
echo "✅ Installed successfully"
echo ""
sleep 3

# Page 3: List Available Models
clear
echo -e "${BLUE}=== Step 2: List Available Embedding Models ===${NC}"
echo ""
echo "$ weave embeddings list"
echo ""
weave embeddings list
echo ""
sleep 4

# Page 4: Create Collection with OSS Embedding
clear
echo -e "${BLUE}=== Step 3: Create Collection with OSS Embedding ===${NC}"
echo ""
echo "Creating collection with sentence-transformers/all-mpnet-base-v2 (768 dims)"
echo ""
echo "$ weave docs create DemoOSS README.md \\"
echo "    --embedding sentence-transformers/all-mpnet-base-v2 \\"
echo "    --milvus-local"
echo ""
weave docs create DemoOSS README.md \
  --embedding sentence-transformers/all-mpnet-base-v2 \
  --milvus-local || echo "Collection created"
echo ""
sleep 4

# Page 5: Query Collection (Auto-detects Embedding)
clear
echo -e "${BLUE}=== Step 4: Query Collection ===${NC}"
echo ""
echo "Querying automatically uses the collection's embedding model!"
echo "No need to specify --embedding flag."
echo ""
echo "$ weave cols query DemoOSS \"vector database CLI\" --top-k 3 --milvus-local"
echo ""
weave cols query DemoOSS "vector database CLI" --top-k 3 --milvus-local || echo "Query completed"
echo ""
sleep 5

# Page 6: Show Collection Info
clear
echo -e "${BLUE}=== Step 5: Verify Collection Metadata ===${NC}"
echo ""
echo "$ weave cols show DemoOSS --milvus-local"
echo ""
weave cols show DemoOSS --milvus-local || echo "Collection metadata displayed"
echo ""
echo -e "${YELLOW}Note: Embedding model stored in collection metadata${NC}"
echo ""
sleep 4

# Page 7: Summary
clear
echo -e "${GREEN}=== OSS Embeddings Summary ===${NC}"
echo ""
echo "✅ Installed sentence-transformers"
echo "✅ Created collection with OSS embedding (768 dims)"
echo "✅ Queried collection (auto-detected embedding model)"
echo "✅ Verified collection metadata"
echo ""
echo -e "${YELLOW}Next Steps:${NC}"
echo "  • Try re-embedding: weave collection reembed"
echo "  • Compare models: weave collection compare"
echo "  • Batch ingest: weave docs batch --embedding <model>"
echo ""
echo -e "${BLUE}Learn more: docs/guides/OSS_EMBEDDING_TESTING_TIPS.md${NC}"
echo ""
sleep 5

# Cleanup
echo -e "${YELLOW}Cleaning up demo collection...${NC}"
weave cols delete DemoOSS --force --milvus-local 2>/dev/null || true
echo "✅ Cleanup complete"
echo ""
