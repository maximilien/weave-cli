#!/bin/bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2025 dr.max

# Weave CLI Supabase Integration Demo
# Demonstrates Supabase (PostgreSQL + pgvector) as vector database backend

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
echo "  Weave CLI - Supabase Integration Demo"
echo "  PostgreSQL + pgvector Backend"
echo "================================================"
echo ""

# Page 1: Introduction
echo -e "${BLUE}📋 Page 1: Supabase as Vector Database${NC}"
echo ""
echo "Weave CLI supports multiple vector database backends:"
echo "  • Weaviate Cloud"
echo "  • Weaviate Local"
echo "  • Supabase (PostgreSQL + pgvector) ⭐"
echo "  • Mock Database"
echo ""
echo "This demo shows Supabase integration (EXPERIMENTAL)"
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 2: Configuration
echo -e "${BLUE}📋 Page 2: Supabase Configuration${NC}"
echo ""
echo "Configure Supabase connection:"
echo ""
cat <<'EOF'
export SUPABASE_DATABASE_URL="postgres://postgres:[password]@db.[project].supabase.co:6543/postgres"
export SUPABASE_DATABASE_KEY="your-supabase-anon-key"
export VECTOR_DB_TYPE="supabase"

# Or use interactive setup
weave config create --database-type supabase
EOF
echo ""
echo -e "${YELLOW}Note: Use transaction pooler (port 6543) for IPv4 compatibility${NC}"
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 3: Show current config
echo -e "${BLUE}📋 Page 3: View Supabase Configuration${NC}"
echo ""
echo "$ weave config show"
"$WEAVE_BIN" config show
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 4: Health check
echo -e "${BLUE}📋 Page 4: Verify Supabase Connection${NC}"
echo ""
echo "$ weave health check"
"$WEAVE_BIN" health check
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 5: List collections
echo -e "${BLUE}📋 Page 5: List Supabase Collections${NC}"
echo ""
echo "$ weave cols ls --supabase"
"$WEAVE_BIN" cols ls --supabase
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 6: Create collection
echo -e "${BLUE}📋 Page 6: Create Collection in Supabase${NC}"
echo ""
echo "$ weave cols create SupabaseDemoCol --supabase"
"$WEAVE_BIN" cols create SupabaseDemoCol --supabase || true
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 7: Add documents
echo -e "${BLUE}📋 Page 7: Add Documents to Supabase${NC}"
echo ""
echo "$ weave docs create SupabaseDemoCol README.md --supabase"
"$WEAVE_BIN" docs create SupabaseDemoCol README.md --supabase || true
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 8: List documents
echo -e "${BLUE}📋 Page 8: List Documents in Supabase${NC}"
echo ""
echo "$ weave docs ls SupabaseDemoCol --supabase"
"$WEAVE_BIN" docs ls SupabaseDemoCol --supabase
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 9: Semantic search
echo -e "${BLUE}📋 Page 9: Semantic Search with Supabase${NC}"
echo ""
echo "$ weave cols q SupabaseDemoCol \"vector database\" --supabase"
"$WEAVE_BIN" cols q SupabaseDemoCol "vector database" --supabase
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 10: BM25 keyword search
echo -e "${BLUE}📋 Page 10: BM25 Keyword Search (Supabase Feature)${NC}"
echo ""
echo "Supabase supports full-text search with BM25:"
echo ""
echo "$ weave cols q SupabaseDemoCol \"installation\" --bm25 --supabase"
"$WEAVE_BIN" cols q SupabaseDemoCol "installation" --bm25 --supabase || echo -e "${YELLOW}BM25 requires text search setup${NC}"
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 11: Hybrid search
echo -e "${BLUE}📋 Page 11: Hybrid Search (Vector + BM25)${NC}"
echo ""
echo "Combine semantic and keyword search:"
echo ""
echo "$ weave cols q SupabaseDemoCol \"CLI tool\" --hybrid --supabase"
"$WEAVE_BIN" cols q SupabaseDemoCol "CLI tool" --hybrid --supabase || echo -e "${YELLOW}Hybrid search combines vector + BM25${NC}"
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 12: Multi-database operations
echo -e "${BLUE}📋 Page 12: Multi-Database Operations${NC}"
echo ""
echo "Query across both Weaviate and Supabase:"
echo ""
echo "$ weave cols ls --weaviate --supabase"
"$WEAVE_BIN" cols ls --weaviate --supabase || echo -e "${YELLOW}Shows collections from both databases${NC}"
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 13: Supabase features
echo -e "${BLUE}📋 Page 13: Supabase-Specific Features${NC}"
echo ""
echo "Supabase advantages:"
echo "  ✓ PostgreSQL reliability"
echo "  ✓ pgvector extension"
echo "  ✓ Full-text search (BM25)"
echo "  ✓ Hybrid search (vector + keyword)"
echo "  ✓ ACID transactions"
echo "  ✓ SQL interface"
echo "  ✓ Real-time subscriptions"
echo ""
echo "Current limitations (EXPERIMENTAL):"
echo "  ⚠ Only OpenAI embeddings supported"
echo "  ⚠ Some metadata features in development"
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 14: Cleanup
echo -e "${BLUE}📋 Page 14: Cleanup${NC}"
echo ""
echo "Clean up demo collection:"
echo ""
echo "$ weave cols delete SupabaseDemoCol --supabase --force"
read -r -p "Clean up demo collection? [y/N]: " -n 1
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    "$WEAVE_BIN" cols delete SupabaseDemoCol --supabase --force || true
fi
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 15: Summary
echo -e "${GREEN}✅ Supabase Integration Demo Complete!${NC}"
echo ""
echo "Summary:"
echo "  ✓ Configured Supabase connection"
echo "  ✓ Created collection in Supabase"
echo "  ✓ Added and queried documents"
echo "  ✓ Tried semantic search"
echo "  ✓ Tried BM25 keyword search"
echo "  ✓ Explored hybrid search"
echo "  ✓ Learned multi-database operations"
echo ""
echo "Next steps:"
echo "  • See: docs/supabase/README.md"
echo "  • Try: weave cols ls --all"
echo "  • Learn: docs/VDB_SUPPORT.md"
echo ""
echo "Documentation:"
echo "  • Supabase setup: docs/supabase/README.md"
echo "  • Testing guide: docs/supabase/TESTING.md"
echo "  • Feature matrix: docs/VDB_SUPPORT.md"
echo ""
echo "================================================"
