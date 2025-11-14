#!/bin/bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2025 dr.max

# Weave CLI REPL Demo
# AI-powered natural language interface demonstration

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
echo "  Weave CLI - AI-Powered REPL Demo"
echo "================================================"
echo ""

# Page 1: Introduction
echo -e "${BLUE}📋 Page 1: AI-Powered REPL${NC}"
echo ""
echo "Weave CLI includes an AI-powered REPL (Read-Eval-Print Loop) that"
echo "understands natural language commands using GPT-4o multi-agent system."
echo ""
echo "Prerequisites:"
echo "  • OPENAI_API_KEY set in environment"
echo "  • weave-mcp binary installed (run: weave config update --weave-mcp)"
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 2: Check Prerequisites
echo -e "${BLUE}📋 Page 2: Check Prerequisites${NC}"
echo ""
echo "1. Check if weave-mcp is installed:"
echo ""
echo "$ weave config show | grep WEAVE_MCP_STDIO_PATH"
"$WEAVE_BIN" config show | grep WEAVE_MCP_STDIO_PATH || echo -e "${YELLOW}weave-mcp not installed - run: weave config update --weave-mcp${NC}"
echo ""
echo "2. Verify OpenAI API key is set:"
echo ""
if [ -n "$OPENAI_API_KEY" ]; then
    echo -e "${GREEN}✓ OPENAI_API_KEY is set${NC}"
else
    echo -e "${YELLOW}⚠ OPENAI_API_KEY not set - REPL may not work${NC}"
fi
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 3: Simple Query Examples
echo -e "${BLUE}📋 Page 3: Simple Query Examples${NC}"
echo ""
echo "Try these natural language queries:"
echo ""
echo "$ weave query \"show me all my collections\""
echo "$ weave q \"create a collection called TestDocs\""
echo "$ weave q \"add README.md to TestDocs\""
echo "$ weave q \"search TestDocs for 'installation'\""
echo "$ weave q \"how many documents in TestDocs?\""
echo "$ weave q \"delete the TestDocs collection\""
echo ""
echo -e "${YELLOW}Note: Use 'weave query' or the short alias 'weave q'${NC}"
echo ""
read -r -p "Press Enter to see a live example..."
clear

# Page 4: Live Example - List Collections
echo -e "${BLUE}📋 Page 4: Example Query Output${NC}"
echo ""
echo "When you run:"
echo ""
echo "$ weave q \"show me all my collections\""
echo ""
echo "The AI REPL will:"
echo "  1. Understand your natural language query"
echo "  2. Plan the best weave command to execute"
echo "  3. Show you the plan and ask for approval"
echo "  4. Execute: weave cols ls"
echo "  5. Display results with real-time progress"
echo ""
echo -e "${YELLOW}(Skipping live execution in demo - requires interactive approval)${NC}"
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 5: Advanced Features
echo -e "${BLUE}📋 Page 5: Advanced Features${NC}"
echo ""
echo "The AI REPL supports:"
echo ""
echo "  🤖 Multi-Agent System"
echo "     • 7 specialized agents for different tasks"
echo "     • Planner, Executor, Validator agents"
echo ""
echo "  💰 Cost Tracking"
echo "     • Track token usage and costs"
echo "     • Color-coded cost display"
echo ""
echo "  🔍 Smart Error Detection"
echo "     • Detects errors in command output"
echo "     • Provides helpful suggestions"
echo ""
echo "  ⏱️ Progress Tracking"
echo "     • Real-time step progress"
echo "     • Duration tracking for each step"
echo ""
echo "  ✨ Auto-Approval"
echo "     • Simple weave commands auto-approved"
echo "     • Confirmation for destructive operations"
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 6: Dry-Run Mode
echo -e "${BLUE}📋 Page 6: Dry-Run Mode${NC}"
echo ""
echo "Test queries without making changes:"
echo ""
echo "$ weave q \"delete all my collections\" --dry-run"
echo ""
echo "Dry-run mode:"
echo "  • Shows what would be executed"
echo "  • No actual changes made"
echo "  • Safe for testing complex queries"
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 7: Interactive REPL Mode
echo -e "${BLUE}📋 Page 7: Interactive REPL Mode${NC}"
echo ""
echo "Start interactive REPL session:"
echo ""
echo "$ weave"
echo ""
echo "Then type natural language commands:"
echo ""
cat <<'EOF'
weave> show me all collections
weave> create TestDocs collection
weave> add README.md to TestDocs
weave> search TestDocs for "installation"
weave> delete TestDocs
weave> exit
EOF
echo ""
echo -e "${YELLOW}(Not running interactive mode in this demo)${NC}"
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 8: Observability
echo -e "${BLUE}📋 Page 8: Observability with Opik${NC}"
echo ""
echo "Weave CLI integrates with Opik for LLM observability:"
echo ""
echo "  📊 Metrics Tracking"
echo "     • Token usage per agent"
echo "     • Response times"
echo "     • Success/failure rates"
echo ""
echo "  🔍 Trace Analysis"
echo "     • Full conversation traces"
echo "     • Agent decision flow"
echo "     • Cost breakdown"
echo ""
echo "  📈 Performance Monitoring"
echo "     • Query performance over time"
echo "     • Model comparison"
echo "     • Cost optimization"
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 9: Summary
echo -e "${GREEN}✅ REPL Demo Complete!${NC}"
echo ""
echo "Summary:"
echo "  ✓ Learned about AI-powered REPL"
echo "  ✓ Saw natural language examples"
echo "  ✓ Discovered advanced features"
echo "  ✓ Learned about dry-run mode"
echo "  ✓ Explored observability features"
echo ""
echo "Key Features:"
echo "  🤖 GPT-4o multi-agent system"
echo "  💬 Natural language understanding"
echo "  🔒 Safe with confirmations & dry-run"
echo "  📊 Full observability with Opik"
echo "  ⚡ Real-time progress feedback"
echo ""
echo "Next steps:"
echo "  • Install: weave config update --weave-mcp"
echo "  • Try: weave q \"show me all collections\""
echo "  • Start REPL: weave"
echo "  • See: docs/guides/WEAVE_CLI_AI.md"
echo ""
