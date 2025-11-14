#!/bin/bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2025 dr.max

# Weave CLI Quick Demo (2 minutes)
# Fast overview of core functionality

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
WEAVE_BIN="$PROJECT_ROOT/bin/weave"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "================================================"
echo "  Weave CLI - Quick Demo (2 minutes)"
echo "================================================"
echo ""

# Cleanup: Delete demo collection if it exists (silent)
"$WEAVE_BIN" cols delete QuickDemoCol --force >/dev/null 2>&1 || true
sleep 1  # Wait for deletion to complete

# Page 1: Health Check
echo -e "${BLUE}📋 Page 1: Health Check${NC}"
echo ""
echo "$ weave health check"
"$WEAVE_BIN" health check
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 2: List Collections
echo -e "${BLUE}📋 Page 2: List Collections${NC}"
echo ""
echo "$ weave cols ls"
"$WEAVE_BIN" cols ls
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 3: Create Collection
echo -e "${BLUE}📋 Page 3: Create Collection${NC}"
echo ""
echo "$ weave cols create QuickDemoCol --embedding text-embedding-3-small"
"$WEAVE_BIN" cols create QuickDemoCol --embedding text-embedding-3-small || true
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 4: Add Document
echo -e "${BLUE}📋 Page 4: Add Document${NC}"
echo ""
echo "$ weave docs create QuickDemoCol README.md"
"$WEAVE_BIN" docs create QuickDemoCol README.md || true
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 5: List Documents
echo -e "${BLUE}📋 Page 5: List Documents${NC}"
echo ""
echo "$ weave docs ls QuickDemoCol"
"$WEAVE_BIN" docs ls QuickDemoCol
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 6: Query Collection
echo -e "${BLUE}📋 Page 6: Query Collection${NC}"
echo ""
echo "$ weave cols q QuickDemoCol \"vector database\""
"$WEAVE_BIN" cols q QuickDemoCol "vector database"
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 7: Cleanup
echo -e "${BLUE}📋 Page 7: Cleanup${NC}"
echo ""
echo "To clean up the demo collection, run:"
echo ""
echo "$ weave cols delete QuickDemoCol --force"
echo ""
echo -e "${YELLOW}(Skipping cleanup in demo - you can delete manually if needed)${NC}"
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 8: Summary
echo -e "${GREEN}✅ Quick Demo Complete!${NC}"
echo ""
echo "Summary:"
echo "  ✓ Checked database health"
echo "  ✓ Listed collections"
echo "  ✓ Created a collection"
echo "  ✓ Added a document"
echo "  ✓ Queried the collection"
echo ""
echo "Next steps:"
echo "  • Try: weave --help"
echo "  • See: docs/USER_GUIDE.md"
echo "  • Run: demos/full-demo.sh for complete walkthrough"
echo ""
