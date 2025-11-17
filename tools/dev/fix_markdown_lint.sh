#!/bin/bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2025 dr.max

# Automatic Markdown Linting Fixes
# This script automatically fixes common markdown linting issues

# Don't exit on error - we want to process all files
set +e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "📝 Fixing markdown linting issues..."

# Fix all markdown files in project root
for file in "$PROJECT_ROOT"/*.md; do
    if [ -f "$file" ]; then
        echo "  Fixing: $(basename "$file")"

        # Remove trailing spaces (macOS compatible)
        if [[ "$OSTYPE" == "darwin"* ]]; then
            sed -i '' 's/[[:space:]]*$//' "$file" 2>/dev/null || true
        else
            sed -i 's/[[:space:]]*$//' "$file" 2>/dev/null || true
        fi

        # Ensure file ends with exactly one newline
        if [ -n "$(tail -c1 "$file" 2>/dev/null)" ]; then
            echo '' >> "$file" 2>/dev/null || true
        fi

        # Add 'text' language to code blocks that don't have a language
        # Only replace opening ``` tags (followed by newline and content), not closing ones
        if [[ "$OSTYPE" == "darwin"* ]]; then
            sed -i '' '/^```$/{N;/^```\n[^`]/s/^```/```text/;}' "$file" 2>/dev/null || true
        else
            sed -i '/^```$/{N;/^```\n[^`]/s/^```/```text/;}' "$file" 2>/dev/null || true
        fi
    fi
done

echo "✅ Markdown linting fixes applied"

# Always exit successfully
exit 0
