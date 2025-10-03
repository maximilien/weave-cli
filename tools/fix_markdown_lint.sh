#!/bin/bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2025 dr.max

# Automatic Markdown Linting Fixes
# This script automatically fixes common markdown linting issues

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "📝 Fixing markdown linting issues..."

# Fix all markdown files in project root
for file in "$PROJECT_ROOT"/*.md; do
    if [ -f "$file" ]; then
        echo "  Fixing: $(basename "$file")"

        # Remove trailing spaces
        sed -i '' 's/[[:space:]]*$//' "$file"

        # Ensure file ends with exactly one newline
        if [ -n "$(tail -c1 "$file" 2>/dev/null)" ]; then
            echo '' >> "$file"
        fi

        # Add 'text' language to code blocks that don't have a language
        sed -i '' 's/^```$/```text/' "$file" || true
    fi
done

echo "✅ Markdown linting fixes applied"
