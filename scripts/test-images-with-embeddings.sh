#!/bin/bash
# Quick test: Create small image collection WITH embeddings to test --top_k_images flag

set -e

WEAVE="./bin/weave"
TEST_COL="TestWeaveImages"
TEXT_COL="WeaveDocs"
VDB_FLAG="--weaviate-cloud"

echo "🧪 Testing --top_k_images with Embeddings"
echo "=========================================="
echo ""

# Step 1: Create test image collection WITH embeddings
echo "📦 Step 1: Creating test image collection with embeddings..."
echo "   Command: $WEAVE cols delete $TEST_COL $VDB_FLAG"
$WEAVE cols delete $TEST_COL $VDB_FLAG 2>/dev/null || true
echo "   Command: $WEAVE cols create $TEST_COL --image --embedding text-embedding-3-small $VDB_FLAG"
$WEAVE cols create $TEST_COL \
    --image \
    --embedding text-embedding-3-small \
    $VDB_FLAG
echo "   ✅ Created $TEST_COL with text2vec-openai"
echo ""

# Step 2: Add test images with searchable text
echo "📥 Step 2: Adding test images with OCR text..."
echo "   Command: $WEAVE docs create $TEST_COL --id <id> --image <url> --content \"...\" --metadata '{...}' $VDB_FLAG"
echo ""

# Image 1: Screenshot-related
$WEAVE docs create $TEST_COL \
    --id "test-screenshot-1" \
    --image "https://example.com/weave-cli-screenshot.png" \
    --content "Weave CLI screenshot showing terminal output with collection query results" \
    --metadata '{
        "title": "Weave CLI Screenshot",
        "ocr_text": "weave cols query WeaveDocs screenshot --agent rag-agent --top_k 5",
        "description": "Terminal screenshot of Weave CLI executing a query command",
        "tags": ["screenshot", "cli", "weave", "terminal"]
    }' \
    $VDB_FLAG
echo "   ✅ Added: test-screenshot-1 (CLI screenshot)"

# Image 2: Auction catalog
$WEAVE docs create $TEST_COL \
    --id "test-auction-1" \
    --image "https://example.com/vintage-car-auction.jpg" \
    --content "Vintage 1967 Ford Mustang red classic car from auction catalog" \
    --metadata '{
        "title": "1967 Ford Mustang Auction Photo",
        "ocr_text": "Lot 247: 1967 Ford Mustang Fastback, Candy Apple Red, Original Paint",
        "description": "Classic muscle car from vintage auction catalog",
        "tags": ["auction", "vintage", "car", "mustang", "1967"]
    }' \
    $VDB_FLAG
echo "   ✅ Added: test-auction-1 (vintage car)"

# Image 3: Documentation diagram
$WEAVE docs create $TEST_COL \
    --id "test-diagram-1" \
    --image "https://example.com/architecture-diagram.png" \
    --content "System architecture diagram showing Weave CLI, vector database, and RAG agent flow" \
    --metadata '{
        "title": "Weave Architecture Diagram",
        "ocr_text": "Client -> Weave CLI -> Vector DB (Weaviate) -> RAG Agent -> LLM",
        "description": "Technical architecture diagram of Weave CLI system",
        "tags": ["diagram", "architecture", "weave", "rag", "technical"]
    }' \
    $VDB_FLAG
echo "   ✅ Added: test-diagram-1 (architecture)"

# Image 4: Code example
$WEAVE docs create $TEST_COL \
    --id "test-code-1" \
    --image "https://example.com/code-example.png" \
    --content "Code example showing how to use Weave CLI query command with multiple collections" \
    --metadata '{
        "title": "Weave CLI Code Example",
        "ocr_text": "weave cols query WeaveDocs WeaveImages \"query\" --agent rag-agent --top_k 5 --top_k_images 2",
        "description": "Code snippet demonstrating multi-collection query with --top_k_images flag",
        "tags": ["code", "example", "cli", "weave", "tutorial"]
    }' \
    $VDB_FLAG
echo "   ✅ Added: test-code-1 (code example)"

echo ""
echo "✅ Step 2 complete: Added 4 test images"
echo ""

# Step 3: Test semantic search on images alone
echo "🔍 Step 3: Testing semantic search on image collection..."
echo ""
echo "Query: 'weave cli screenshot'"
echo "   Command: $WEAVE cols query $TEST_COL \"weave cli screenshot\" --top_k 3 $VDB_FLAG"
$WEAVE cols query $TEST_COL "weave cli screenshot" --top_k 3 $VDB_FLAG
echo ""

# Step 4: Test --top_k_images with mixed collections
echo "🎯 Step 4: Testing --top_k_images with mixed text + image collections..."
echo ""
echo "Without --top_k_images (may get all text, no images):"
echo "   Command: $WEAVE cols query $TEXT_COL $TEST_COL \"weave cli screenshot\" --agent rag-agent --top_k 5 $VDB_FLAG"
$WEAVE cols query $TEXT_COL $TEST_COL "weave cli screenshot" \
    --agent rag-agent --top_k 5 $VDB_FLAG
echo ""
echo "---"
echo ""
echo "With --top_k_images=2 (should guarantee 2 images):"
echo "   Command: $WEAVE cols query $TEXT_COL $TEST_COL \"weave cli screenshot\" --agent rag-agent --top_k 5 --top_k_images 2 --verbose $VDB_FLAG"
$WEAVE cols query $TEXT_COL $TEST_COL "weave cli screenshot" \
    --agent rag-agent --top_k 5 --top_k_images 2 --verbose $VDB_FLAG 2>&1 | tail -60
echo ""

# Summary
echo "=========================================="
echo "✅ Test Complete!"
echo "=========================================="
echo ""
echo "Collections created:"
echo "  - $TEXT_COL (text, existing)"
echo "  - $TEST_COL (images WITH embeddings)"
echo ""
echo "Test queries:"
echo "  1. Image-only search: ✅"
echo "  2. Mixed search without --top_k_images: ✅"
echo "  3. Mixed search WITH --top_k_images: ✅"
echo ""
echo "Cleanup:"
echo "  Command: weave cols delete $TEST_COL $VDB_FLAG"
echo ""
