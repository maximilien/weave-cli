#!/bin/bash
# Clean up demo collections before/after a run

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

if [ -x "$SCRIPT_DIR/bin/weave" ]; then
    WEAVE="$SCRIPT_DIR/bin/weave"
else
    WEAVE="weave"
fi

echo "Cleaning up OpikDemo collections..."
echo ""

# Milvus
echo "--- Milvus local ---"
"$WEAVE" cols del OpikDemo --force --milvus-local 2>&1 || echo "  (nothing to clean)"

# Weaviate
echo ""
echo "--- Weaviate cloud ---"
"$WEAVE" docs da OpikDemo --no-confirm --weaviate-cloud 2>&1 || echo "  (nothing to clean)"
"$WEAVE" docs da OpikDemoImages --no-confirm --weaviate-cloud 2>&1 || echo "  (nothing to clean)"
"$WEAVE" cols ds OpikDemo --no-confirm --weaviate-cloud 2>&1 || echo "  (nothing to clean)"
"$WEAVE" cols ds OpikDemoImages --no-confirm --weaviate-cloud 2>&1 || echo "  (nothing to clean)"

# Backup
rm -f /tmp/opik-demo.weavebak

echo ""
echo "Done. Ready for a fresh demo run."
