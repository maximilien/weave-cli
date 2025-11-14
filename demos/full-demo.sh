#!/bin/bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2025 dr.max

# Weave CLI Full Demo (5 minutes)
# Complete walkthrough of all features

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
echo "  Weave CLI - Full Feature Demo (5 minutes)"
echo "================================================"
echo ""

# Cleanup: Delete demo collection if it exists (silent)
"$WEAVE_BIN" cols delete FullDemoCol --force >/dev/null 2>&1 || true
sleep 1  # Wait for deletion to complete

# Page 1: Configuration
echo -e "${BLUE}📋 Page 1: Configuration${NC}"
echo ""
echo "$ weave config show"
"$WEAVE_BIN" config show
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 2: Health Check
echo -e "${BLUE}📋 Page 2: Health Check${NC}"
echo ""
echo "$ weave health check"
"$WEAVE_BIN" health check
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 3: List Collections
echo -e "${BLUE}📋 Page 3: List Collections${NC}"
echo ""
echo "$ weave cols ls"
"$WEAVE_BIN" cols ls
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 4: Create Collection
echo -e "${BLUE}📋 Page 4: Create Collection${NC}"
echo ""
echo "$ weave cols create FullDemoCol --embedding text-embedding-3-small"
"$WEAVE_BIN" cols create FullDemoCol --embedding text-embedding-3-small || true
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 5: Show Collection Details
echo -e "${BLUE}📋 Page 5: Show Collection Details${NC}"
echo ""
echo "$ weave cols show FullDemoCol"
"$WEAVE_BIN" cols show FullDemoCol
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 6: Count Collections
echo -e "${BLUE}📋 Page 6: Count Collections${NC}"
echo ""
echo "$ weave cols count"
"$WEAVE_BIN" cols count
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 7: Add Documents
echo -e "${BLUE}📋 Page 7: Add Documents${NC}"
echo ""
echo "$ weave docs create FullDemoCol README.md"
"$WEAVE_BIN" docs create FullDemoCol README.md || true
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 8: List Documents
echo -e "${BLUE}📋 Page 8: List Documents${NC}"
echo ""
echo "$ weave docs ls FullDemoCol"
"$WEAVE_BIN" docs ls FullDemoCol
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 9: Count Documents
echo -e "${BLUE}📋 Page 9: Count Documents${NC}"
echo ""
echo "$ weave docs count FullDemoCol"
"$WEAVE_BIN" docs count FullDemoCol
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 10: Query Collection - Semantic Search
echo -e "${BLUE}📋 Page 10: Semantic Search${NC}"
echo ""
echo "$ weave cols q FullDemoCol \"vector database\""
"$WEAVE_BIN" cols q FullDemoCol "vector database"
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 11: List Available Embeddings
echo -e "${BLUE}📋 Page 11: Available Embeddings${NC}"
echo ""
echo "$ weave emb ls"
"$WEAVE_BIN" emb ls
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 12: Multi-Database Support
echo -e "${BLUE}📋 Page 12: Multi-Database Support${NC}"
echo ""
echo "List collections from all databases:"
echo ""
echo "$ weave cols ls --all"
"$WEAVE_BIN" cols ls --all || true
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 13: Batch Processing
echo -e "${BLUE}📋 Page 13: Batch Processing${NC}"
echo ""
echo "Batch process documents from a directory:"
echo ""
echo "$ weave docs batch --directory ./docs --collection MyCol --parallel 3"
echo ""
echo -e "${YELLOW}(Example command - not running in demo)${NC}"
echo ""
echo "Features:"
echo "  • Process entire directories"
echo "  • Parallel processing with configurable workers"
echo "  • Automatic retry on failures"
echo "  • Progress tracking"
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 14: Cleanup
echo -e "${BLUE}📋 Page 14: Cleanup${NC}"
echo ""
echo "To clean up the demo collection, run:"
echo ""
echo "$ weave cols delete FullDemoCol --force"
echo ""
echo -e "${YELLOW}(Skipping cleanup in demo - you can delete manually if needed)${NC}"
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 15: Summary
echo -e "${GREEN}✅ Full Demo Complete!${NC}"
echo ""
echo "Summary:"
echo "  ✓ Viewed configuration"
echo "  ✓ Checked database health"
echo "  ✓ Created and managed collections"
echo "  ✓ Added and queried documents"
echo "  ✓ Explored semantic search"
echo "  ✓ Learned about embeddings"
echo "  ✓ Discovered multi-database support"
echo "  ✓ Learned about batch processing"
echo ""
echo "Key Features:"
echo "  🤖 AI-Powered - Natural language REPL interface"
echo "  ⚡ Fast & Easy - Simple CLI with real-time feedback"
echo "  🌐 Flexible - Weaviate, Supabase, or mock database"
echo "  📦 Batch Processing - Parallel directory processing"
echo "  📄 PDF Support - Intelligent text & image extraction"
echo ""
echo "Next steps:"
echo "  • Try: weave --help"
echo "  • See: docs/USER_GUIDE.md"
echo "  • Run: demos/repl-demo.sh for AI interface demo"
echo "  • Run: demos/supabase-demo.sh for Supabase demo"
echo ""
