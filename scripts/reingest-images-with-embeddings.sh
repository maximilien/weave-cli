#!/bin/bash
# Re-ingest image collections with embeddings enabled
# This script exports existing image documents and re-creates collections with proper embeddings

set -e

WEAVE="./bin/weave"
COLLECTION="WeaveImages"
BACKUP_COLLECTION="${COLLECTION}_backup_$(date +%Y%m%d_%H%M%S)"
VDB_FLAG="--weaviate-cloud"

echo "🔄 Re-ingesting $COLLECTION with Embeddings"
echo "==========================================="
echo ""

# Step 1: Count existing documents
echo "📊 Step 1: Analyzing existing collection..."
echo "   Command: $WEAVE docs list $COLLECTION $VDB_FLAG"
DOC_COUNT=$($WEAVE docs list $COLLECTION $VDB_FLAG 2>&1 | grep "Found" | awk '{print $2}')
echo "   Found: $DOC_COUNT documents"
echo ""

# Step 2: Check embedding status
echo "📋 Step 2: Checking current embedding status..."
echo "   Command: $WEAVE docs list $COLLECTION --limit 1 $VDB_FLAG"
EMBEDDING_MODEL=$($WEAVE docs list $COLLECTION --limit 1 $VDB_FLAG 2>&1 | grep "Embedding model:" | awk '{print $3}')
echo "   Current embedding model: $EMBEDDING_MODEL"

if [ "$EMBEDDING_MODEL" != "none" ]; then
    echo "   ⚠️  Collection already has embeddings!"
    read -p "   Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "   Aborted."
        exit 1
    fi
fi
echo ""

# Step 3: Rename existing collection as backup
echo "📦 Step 3: Creating backup..."
echo "   Renaming $COLLECTION → $BACKUP_COLLECTION"
# Note: Weaviate doesn't support rename, so we'll just note this for manual backup
echo "   ⚠️  Manual step: You may want to export data first if needed"
echo "   (Collection will be deleted and recreated)"
read -p "   Continue? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "   Aborted."
    exit 1
fi
echo ""

# Step 4: Export all documents to JSON (for backup and re-ingestion)
echo "📤 Step 4: Exporting documents..."
EXPORT_FILE="weave_images_export_$(date +%Y%m%d_%H%M%S).jsonl"
echo "   Exporting to: $EXPORT_FILE"
echo "   Command: $WEAVE docs list $COLLECTION --limit 9999 $VDB_FLAG --json > $EXPORT_FILE"

# Export documents (we'll do this via script)
$WEAVE docs list $COLLECTION --limit 9999 $VDB_FLAG --json > "$EXPORT_FILE" 2>&1
echo "   ✅ Exported $DOC_COUNT documents"
echo ""

# Step 5: Delete old collection
echo "🗑️  Step 5: Deleting old collection..."
read -p "   DELETE $COLLECTION? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "   Command: $WEAVE cols delete $COLLECTION $VDB_FLAG"
    $WEAVE cols delete $COLLECTION $VDB_FLAG
    echo "   ✅ Deleted"
else
    echo "   Aborted."
    exit 1
fi
echo ""

# Step 6: Create new collection with embeddings
echo "🆕 Step 6: Creating new collection with embeddings..."
echo "   Command: $WEAVE cols create $COLLECTION --image --embedding text-embedding-3-small $VDB_FLAG"
$WEAVE cols create $COLLECTION \
    --image \
    --embedding text-embedding-3-small \
    $VDB_FLAG
echo "   ✅ Created with text2vec-openai embeddings"
echo ""

# Step 7: Re-ingest documents
echo "📥 Step 7: Re-ingesting documents with embeddings..."
echo "   This will take a while for $DOC_COUNT documents..."
echo ""
echo "   ⚠️  Manual re-ingestion required:"
echo "   Use: scripts/reingest-from-export.sh $EXPORT_FILE"
echo ""
echo "   Or manually re-ingest source images with:"
echo "   weave docs create $COLLECTION --image <path> --metadata {...} $VDB_FLAG"
echo ""

echo "✅ Collection recreated! Next: Re-ingest documents from export or source."
echo ""
echo "📋 Summary:"
echo "   - Old collection: DELETED"
echo "   - Backup export: $EXPORT_FILE"
echo "   - New collection: $COLLECTION (with text2vec-openai embeddings)"
echo "   - Documents to re-ingest: $DOC_COUNT"
echo ""
