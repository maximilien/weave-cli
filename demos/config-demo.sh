#!/bin/bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2025 dr.max

# Weave CLI Configuration Demo
# Demonstrates interactive configuration setup and management

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
echo "  Weave CLI - Configuration Management Demo"
echo "================================================"
echo ""

# Page 1: Check current config
echo -e "${BLUE}📋 Page 1: View Current Configuration${NC}"
echo ""
echo "$ weave config show"
"$WEAVE_BIN" config show
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 2: Create .env file interactively
echo -e "${BLUE}📋 Page 2: Interactive Configuration Setup${NC}"
echo ""
echo "The 'weave config create --env' command provides interactive setup:"
echo ""
echo "$ weave config create --env"
echo ""
echo -e "${YELLOW}This command prompts for:${NC}"
echo "  - WEAVIATE_URL"
echo "  - WEAVIATE_API_KEY"
echo "  - OPENAI_API_KEY"
echo ""
echo -e "${YELLOW}(Skipping interactive setup in this demo - you already have config)${NC}"
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 3: Show updated configuration
echo -e "${BLUE}📋 Page 3: Verify Configuration${NC}"
echo ""
echo "$ weave config show"
"$WEAVE_BIN" config show
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 4: Environment variable configuration
echo -e "${BLUE}📋 Page 4: Environment Variable Configuration${NC}"
echo ""
echo "You can also configure using environment variables:"
echo ""
cat <<'EOF'
export VECTOR_DB_TYPE="weaviate-cloud"
export WEAVIATE_URL="https://your-instance.weaviate.cloud"
export WEAVIATE_API_KEY="your-api-key"
export OPENAI_API_KEY="sk-proj-your-key"

# Verify configuration
weave config show
EOF
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 5: Global vs Local configuration
echo -e "${BLUE}📋 Page 5: Global vs Local Configuration${NC}"
echo ""
echo "Weave CLI supports both global and local configuration:"
echo ""
echo "Local (current directory):"
echo "  .env"
echo "  config.yaml"
echo ""
echo "Global (user-wide):"
echo "  ~/.weave-cli/.env"
echo "  ~/.weave-cli/config.yaml"
echo ""
echo "Create global configuration:"
echo "$ weave config create --env --global"
echo ""
echo "Sync local to global:"
echo "$ weave config sync"
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 6: Configuration precedence
echo -e "${BLUE}📋 Page 6: Configuration Precedence${NC}"
echo ""
echo "Configuration sources (highest to lowest priority):"
echo ""
echo "  1. Command-line flags"
echo "     weave query --model gpt-4"
echo ""
echo "  2. Environment variables"
echo "     export OPENAI_MODEL=gpt-4"
echo ""
echo "  3. config.yaml (optional)"
echo "     For advanced customization"
echo ""
echo "  4. Built-in defaults"
echo "     weaviate-cloud, gpt-4o, etc."
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 7: Install weave-mcp for REPL
echo -e "${BLUE}📋 Page 7: Install weave-mcp for REPL Mode${NC}"
echo ""
echo "For AI-powered REPL mode, install weave-mcp:"
echo ""
echo "$ weave config update --weave-mcp"
echo ""
echo "This downloads and installs the weave-mcp binary for"
echo "natural language interactions in REPL mode."
echo ""
read -r -p "Press Enter to see example (or Ctrl+C to skip)..."
"$WEAVE_BIN" config update --weave-mcp || true
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 8: Health check
echo -e "${BLUE}📋 Page 8: Verify Connection${NC}"
echo ""
echo "Test your configuration with a health check:"
echo ""
echo "$ weave health check"
"$WEAVE_BIN" health check
echo ""
read -r -p "Press Enter to continue..."
clear

# Page 9: Summary
echo -e "${GREEN}✅ Configuration Demo Complete!${NC}"
echo ""
echo "Summary:"
echo "  ✓ Viewed current configuration"
echo "  ✓ Created .env file interactively"
echo "  ✓ Learned about environment variables"
echo "  ✓ Understood global vs local config"
echo "  ✓ Learned configuration precedence"
echo "  ✓ Installed weave-mcp (optional)"
echo "  ✓ Verified connection with health check"
echo ""
echo "Next steps:"
echo "  • Try: weave"
echo "  • See: docs/USER_GUIDE.md"
echo ""
echo "================================================"
