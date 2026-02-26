#!/bin/bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2025 dr.max

# Install script for weave CLI
# Installs binary and templates to ~/.local/

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INSTALL_DIR="$HOME/.local"
BIN_DIR="$INSTALL_DIR/bin"
TEMPLATES_DIR="$INSTALL_DIR/templates"

echo -e "${BLUE}════════════════════════════════════════${NC}"
echo -e "${BLUE}  Weave CLI Installer${NC}"
echo -e "${BLUE}════════════════════════════════════════${NC}"
echo

# Check if binary exists
if [ ! -f "$SCRIPT_DIR/bin/weave" ]; then
    echo -e "${RED}Error: bin/weave not found${NC}"
    echo "Run './build.sh' first to build the binary"
    exit 1
fi

# Create directories
echo -e "${BLUE}▶ Creating directories...${NC}"
mkdir -p "$BIN_DIR"
mkdir -p "$TEMPLATES_DIR"

# Install binary
echo -e "${BLUE}▶ Installing binary to $BIN_DIR/weave...${NC}"
cp "$SCRIPT_DIR/bin/weave" "$BIN_DIR/weave"
chmod +x "$BIN_DIR/weave"
echo -e "${GREEN}✓ Binary installed${NC}"

# Install templates
echo -e "${BLUE}▶ Installing templates to $TEMPLATES_DIR...${NC}"
if [ -d "$SCRIPT_DIR/templates" ]; then
    # Remove old templates if they exist
    rm -rf "$TEMPLATES_DIR"
    # Copy new templates
    cp -r "$SCRIPT_DIR/templates" "$TEMPLATES_DIR"
    echo -e "${GREEN}✓ Templates installed${NC}"
else
    echo -e "${YELLOW}⚠️  Warning: templates directory not found${NC}"
    echo "   Stack commands may not work correctly"
fi

echo

# Check if ~/.local/bin is in PATH
if [[ ":$PATH:" != *":$BIN_DIR:"* ]]; then
    echo -e "${YELLOW}⚠️  $BIN_DIR is not in your PATH${NC}"
    echo
    echo "Add it to your shell profile:"
    echo
    if [ -f "$HOME/.zshrc" ]; then
        echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.zshrc"
        echo "  source ~/.zshrc"
    elif [ -f "$HOME/.bashrc" ]; then
        echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.bashrc"
        echo "  source ~/.bashrc"
    else
        echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
    fi
    echo
fi

# Show version
echo -e "${GREEN}✓ Installation complete!${NC}"
echo
echo "Installed:"
echo "  • Binary: $BIN_DIR/weave"
echo "  • Templates: $TEMPLATES_DIR"
echo
echo "Verify installation:"
echo "  weave --version"
echo

echo -e "${BLUE}════════════════════════════════════════${NC}"
