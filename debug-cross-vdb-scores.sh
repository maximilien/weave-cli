#!/bin/bash
# Debug script to see scores from all VDBs before deduplication

echo "=== Querying Milvus only ==="
weave cols query WeaveDocs:milvus-cloud "vdb" --top_k 3

echo ""
echo "=== Querying Weaviate only ==="
weave cols query WeaveDocs:weaviate-cloud "vdb" --top_k 3

echo ""
echo "=== Querying MongoDB only ==="
weave cols query WeaveDocs:mongodb-cloud "vdb" --top_k 3

echo ""
echo "=== Querying all 3 together ==="
weave cols query \
  WeaveDocs:milvus-cloud \
  WeaveDocs:weaviate-cloud \
  WeaveDocs:mongodb-cloud \
  "vdb" --top_k 3
