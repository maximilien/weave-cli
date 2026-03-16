#!/bin/bash
# Backup Performance Profiling
# Tests backup performance with different collections and settings

set -e

OUTPUT_DIR="/tmp/backup-perf-results"
mkdir -p "$OUTPUT_DIR"

echo "🔬 Backup Performance Profiling"
echo "================================"
echo ""

cd /Users/maximilien/github/maximilien/weave-cli

# Test 1: Small collection (38 docs) - Compressed
echo "Test 1: DemoDocs (38 docs) - Compressed"
echo "----------------------------------------"
time ./bin/weave backup create DemoDocs \
  --output "$OUTPUT_DIR/demodocs-compress.weavebak" \
  --compress \
  --quiet
echo ""

# Test 2: Small collection (38 docs) - Uncompressed
echo "Test 2: DemoDocs (38 docs) - Uncompressed"
echo "------------------------------------------"
time ./bin/weave backup create DemoDocs \
  --output "$OUTPUT_DIR/demodocs-nocompress.weavebak" \
  --no-compress \
  --quiet
echo ""

# Test 3: Medium collection (79 docs) - Compressed
echo "Test 3: WeaveDocs (79 docs) - Compressed"
echo "-----------------------------------------"
time ./bin/weave backup create WeaveDocs \
  --output "$OUTPUT_DIR/weavedocs-compress.weavebak" \
  --compress \
  --quiet
echo ""

# Test 4: Large collection (301 docs) - Compressed
echo "Test 4: AuctionsImages (301 docs) - Compressed"
echo "-----------------------------------------------"
time ./bin/weave backup create AuctionsImages \
  --output "$OUTPUT_DIR/auctionsimages-compress.weavebak" \
  --compress \
  --quiet
echo ""

# Test 5: Large collection (301 docs) - Uncompressed
echo "Test 5: AuctionsImages (301 docs) - Uncompressed"
echo "-------------------------------------------------"
time ./bin/weave backup create AuctionsImages \
  --output "$OUTPUT_DIR/auctionsimages-nocompress.weavebak" \
  --no-compress \
  --quiet
echo ""

# Test 6: Batch size variation (301 docs, batch=50)
echo "Test 6: AuctionsImages (301 docs) - Batch size 50"
echo "--------------------------------------------------"
time ./bin/weave backup create AuctionsImages \
  --output "$OUTPUT_DIR/auctionsimages-batch50.weavebak" \
  --compress \
  --batch-size 50 \
  --quiet
echo ""

# Test 7: Batch size variation (301 docs, batch=200)
echo "Test 7: AuctionsImages (301 docs) - Batch size 200"
echo "---------------------------------------------------"
time ./bin/weave backup create AuctionsImages \
  --output "$OUTPUT_DIR/auctionsimages-batch200.weavebak" \
  --compress \
  --batch-size 200 \
  --quiet
echo ""

# Show results summary
echo "📊 Results Summary"
echo "=================="
echo ""
echo "File sizes:"
ls -lh "$OUTPUT_DIR" | grep -E "\.weavebak" | awk '{print $9, "-", $5}'
echo ""
echo "✅ Profiling complete! Results saved to: $OUTPUT_DIR"
