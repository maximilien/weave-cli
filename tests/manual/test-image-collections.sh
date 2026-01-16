#!/bin/bash
# Manual test script for multi-modal RAG image collections
# Tests Phase 1 implementation of image content extraction

set -e

echo "🖼️  Multi-Modal RAG Manual Test Script"
echo "======================================"
echo ""

# Check if weave binary exists
if [ ! -f "./bin/weave" ]; then
    echo "❌ Error: ./bin/weave not found. Please build first:"
    echo "   ./build.sh"
    exit 1
fi

WEAVE="./bin/weave"

# Function to create test collection
create_image_collection() {
    local collection_name=$1
    local vdb_flag=$2

    echo "📦 Creating image collection: $collection_name"
    $WEAVE cols create "$collection_name" \
        --schema-type image \
        $vdb_flag 2>/dev/null || echo "   (Collection may already exist)"
}

# Function to add image document
add_image_doc() {
    local collection=$1
    local id=$2
    local image_url=$3
    local ocr_text=$4
    local description=$5
    local tags=$6
    local vdb_flag=$7

    echo "   Adding: $id"

    # Build metadata JSON
    local metadata
    metadata=$(cat <<EOF
{
  "ocr_text": "$ocr_text",
  "description": "$description",
  "tags": $tags
}
EOF
)

    $WEAVE docs create "$collection" \
        --id "$id" \
        --image "$image_url" \
        --metadata "$metadata" \
        $vdb_flag 2>/dev/null || echo "     (Document may already exist)"
}

# Test 1: Weaviate Cloud Image Collection
echo ""
echo "Test 1: Weaviate Cloud - Vintage Cars"
echo "--------------------------------------"

if [ -n "$WEAVIATE_CLOUD_API_KEY" ]; then
    create_image_collection "TestVintageCars" "--weaviate-cloud"

    add_image_doc "TestVintageCars" \
        "car-001" \
        "https://cdn.example.com/1967-mustang.jpg" \
        "1967 Ford Mustang" \
        "Vintage red Mustang in excellent condition" \
        '["vintage", "car", "mustang", "1967", "classic"]' \
        "--weaviate-cloud"

    add_image_doc "TestVintageCars" \
        "car-002" \
        "https://cdn.example.com/1969-camaro.jpg" \
        "1969 Chevrolet Camaro SS" \
        "Blue Camaro with original paint and chrome details" \
        '["vintage", "car", "camaro", "1969", "muscle-car"]' \
        "--weaviate-cloud"

    add_image_doc "TestVintageCars" \
        "car-003" \
        "https://cdn.example.com/1970-challenger.jpg" \
        "1970 Dodge Challenger" \
        "Orange Dodge Challenger R/T with black stripes" \
        '["vintage", "car", "challenger", "1970", "mopar"]' \
        "--weaviate-cloud"

    echo "✅ Test collection created in Weaviate Cloud"
    echo ""
    echo "🔍 Test Query:"
    echo "   $WEAVE cols query TestVintageCars \"vintage cars from 1960s\" \\"
    echo "     --agent rag-agent --top_k 3 --weaviate-cloud"
else
    echo "⏭️  Skipping (WEAVIATE_CLOUD_API_KEY not set)"
fi

# Test 2: Mixed Text + Image Collections
echo ""
echo "Test 2: Mixed Text + Image Collections"
echo "---------------------------------------"

if [ -n "$WEAVIATE_CLOUD_API_KEY" ]; then
    # Create text collection
    echo "📦 Creating text collection: TestCarDocs"
    $WEAVE cols create "TestCarDocs" --weaviate-cloud 2>/dev/null || \
        echo "   (Collection may already exist)"

    # Add text documents
    echo "   Adding text documents..."
    $WEAVE docs create "TestCarDocs" \
        --id "doc-001" \
        --content "Ford Mustang: Introduced in 1964, the Ford Mustang created the pony car class. The 1967 model was the first major redesign, featuring a larger body and more powerful engine options." \
        --weaviate-cloud 2>/dev/null || echo "     (Document may already exist)"

    $WEAVE docs create "TestCarDocs" \
        --id "doc-002" \
        --content "Chevrolet Camaro: Launched in 1966 as a competitor to the Mustang. The 1969 Camaro SS featured a 396 cubic inch V8 engine and became an icon of American muscle." \
        --weaviate-cloud 2>/dev/null || echo "     (Document may already exist)"

    echo "✅ Mixed collections ready"
    echo ""
    echo "🔍 Test Query (Multi-Collection):"
    echo "   $WEAVE cols query TestCarDocs TestVintageCars \\"
    echo "     \"vintage muscle cars\" --agent rag-agent --top_k 5 --weaviate-cloud"
else
    echo "⏭️  Skipping (WEAVIATE_CLOUD_API_KEY not set)"
fi

# Test 3: Images with Different Metadata Levels
echo ""
echo "Test 3: Weaviate Local - Various Metadata Levels"
echo "-------------------------------------------------"

if command -v docker &> /dev/null || command -v podman &> /dev/null; then
    create_image_collection "TestImageMetadata" "--weaviate-local"

    # Full metadata
    add_image_doc "TestImageMetadata" \
        "full-001" \
        "https://cdn.example.com/full-metadata.jpg" \
        "Complete OCR text from image" \
        "Full description with all details" \
        '["tag1", "tag2", "tag3"]' \
        "--weaviate-local"

    # OCR only
    echo "   Adding: ocr-only-001"
    $WEAVE docs create "TestImageMetadata" \
        --id "ocr-only-001" \
        --image "https://cdn.example.com/ocr-only.jpg" \
        --metadata '{"ocr_text": "Only OCR text available"}' \
        --weaviate-local 2>/dev/null || echo "     (Document may already exist)"

    # Description only
    echo "   Adding: desc-only-001"
    $WEAVE docs create "TestImageMetadata" \
        --id "desc-only-001" \
        --image "https://cdn.example.com/desc-only.jpg" \
        --metadata '{"description": "Only description provided"}' \
        --weaviate-local 2>/dev/null || echo "     (Document may already exist)"

    # No metadata (just URL)
    echo "   Adding: no-meta-001"
    $WEAVE docs create "TestImageMetadata" \
        --id "no-meta-001" \
        --image "https://cdn.example.com/no-metadata.jpg" \
        --weaviate-local 2>/dev/null || echo "     (Document may already exist)"

    echo "✅ Test collection created in Weaviate Local"
    echo ""
    echo "🔍 Test Query:"
    echo "   $WEAVE cols query TestImageMetadata \"test\" \\"
    echo "     --agent rag-agent --top_k 5 --weaviate-local"
else
    echo "⏭️  Skipping (Docker/Podman not available)"
fi

# Test 4: Cross-VDB Image Collections
echo ""
echo "Test 4: Cross-VDB Image Query"
echo "------------------------------"

if [ -n "$WEAVIATE_CLOUD_API_KEY" ] && [ -n "$MONGODB_ATLAS_CONNECTION_STRING" ]; then
    # Create MongoDB image collection
    create_image_collection "TestImagesDB" "--mongodb-cloud"

    add_image_doc "TestImagesDB" \
        "mongo-img-001" \
        "https://cdn.example.com/mongodb-image.jpg" \
        "Image stored in MongoDB" \
        "Testing MongoDB image storage" \
        '["mongodb", "test", "image"]' \
        "--mongodb-cloud"

    echo "✅ Cross-VDB collections ready"
    echo ""
    echo "🔍 Test Query (Cross-VDB):"
    echo "   $WEAVE cols query \\"
    echo "     TestVintageCars:weaviate-cloud \\"
    echo "     TestImagesDB:mongodb-cloud \\"
    echo "     \"vintage cars\" --agent rag-agent --top_k 5"
else
    echo "⏭️  Skipping (Cloud VDB credentials not set)"
fi

echo ""
echo "=================================="
echo "🎯 Test Collections Created!"
echo "=================================="
echo ""
echo "Available Test Collections:"
echo ""
echo "1. TestVintageCars (Weaviate Cloud)"
echo "   - 3 vintage car images with full metadata"
echo "   - OCR text, descriptions, tags"
echo ""
echo "2. TestCarDocs + TestVintageCars (Mixed)"
echo "   - Text documents about cars"
echo "   - Image documents with car photos"
echo "   - Test multi-modal query"
echo ""
echo "3. TestImageMetadata (Weaviate Local)"
echo "   - Various metadata levels"
echo "   - Full, OCR-only, desc-only, no-metadata"
echo ""
echo "4. Cross-VDB (Weaviate + MongoDB)"
echo "   - Same collection across different VDBs"
echo ""
echo "📋 Example Queries:"
echo ""
echo "# Single image collection:"
echo "$WEAVE cols query TestVintageCars \"vintage muscle cars\" \\"
echo "  --agent rag-agent --top_k 3 --weaviate-cloud"
echo ""
echo "# Mixed text + images:"
echo "$WEAVE cols query TestCarDocs TestVintageCars \\"
echo "  \"1967 Mustang\" --agent rag-agent --top_k 5 --weaviate-cloud"
echo ""
echo "# Cross-VDB query:"
echo "$WEAVE cols query \\"
echo "  TestVintageCars:weaviate-cloud \\"
echo "  TestImagesDB:mongodb-cloud \\"
echo "  \"vintage cars\" --agent rag-agent --top_k 5"
echo ""
echo "# Verify image content extraction:"
echo "$WEAVE docs get TestVintageCars car-001 --weaviate-cloud"
echo ""
echo "🧹 Cleanup (when done testing):"
echo "$WEAVE cols delete TestVintageCars --weaviate-cloud"
echo "$WEAVE cols delete TestCarDocs --weaviate-cloud"
echo "$WEAVE cols delete TestImageMetadata --weaviate-local"
echo "$WEAVE cols delete TestImagesDB --mongodb-cloud"
echo ""
