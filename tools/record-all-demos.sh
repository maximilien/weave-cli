#!/bin/bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2025 dr.max

# Automated Demo Recording Script
# Records all Weave CLI demos in sequence

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
VIDEOS_DIR="$PROJECT_ROOT/videos"
DEMOS_DIR="$PROJECT_ROOT/demos"

# Parse command line arguments
OVERWRITE_FLAG=""
if [[ "$1" == "--overwrite" ]]; then
    OVERWRITE_FLAG="--overwrite"
fi

# Source .env file if it exists (to check for Supabase credentials)
if [ -f "$PROJECT_ROOT/.env" ]; then
    set -a
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/.env"
    set +a
fi

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}╔════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  Weave CLI - Automated Demo Recording         ║${NC}"
echo -e "${BLUE}╔════════════════════════════════════════════════╗${NC}"
echo ""

# Check for asciinema
if ! command -v asciinema &> /dev/null; then
    echo -e "${RED}❌ Error: asciinema not found${NC}"
    echo ""
    echo "Install asciinema:"
    echo "  macOS:   brew install asciinema"
    echo "  Linux:   apt-get install asciinema"
    echo ""
    exit 1
fi

# Check prerequisites
echo -e "${BLUE}📋 Checking prerequisites...${NC}"
echo ""

# Check if weave binary exists
if [ ! -f "$PROJECT_ROOT/bin/weave" ]; then
    echo -e "${RED}❌ Error: bin/weave not found${NC}"
    echo "Run: ./build.sh"
    exit 1
fi

echo -e "${GREEN}✓${NC} Weave binary found"

# Function to check configuration
check_config() {
    local config_type=$1
    echo -n "  Checking $config_type configuration... "

    if "$PROJECT_ROOT/bin/weave" health check &> /dev/null; then
        echo -e "${GREEN}✓${NC}"
        return 0
    else
        echo -e "${YELLOW}⚠${NC}"
        return 1
    fi
}

# Check configurations
WEAVIATE_OK=false
SUPABASE_OK=false

# Check if Weaviate is configured (either as default or available)
if check_config "current"; then
    if echo "${VECTOR_DB_TYPE:-weaviate-cloud}" | grep -q "weaviate"; then
        WEAVIATE_OK=true
    fi
fi

# Check if Supabase credentials are available (regardless of default DB)
if [ -n "$SUPABASE_DATABASE_URL" ] && [ -n "$SUPABASE_DATABASE_KEY" ]; then
    SUPABASE_OK=true
fi

# If VECTOR_DB_TYPE is explicitly supabase, ensure it's marked as OK
if echo "${VECTOR_DB_TYPE:-}" | grep -q "supabase"; then
    SUPABASE_OK=true
fi

echo ""

# Function to record a demo with asciinema
record_demo() {
    local demo_name=$1
    local demo_script=$2
    local output_file="$VIDEOS_DIR/weave-cli-${demo_name}.cast"

    echo -e "${BLUE}╔════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║  Recording: ${demo_name} Demo${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════╝${NC}"
    echo ""

    # Check if script exists
    if [ ! -f "$demo_script" ]; then
        echo -e "${RED}❌ Error: Demo script not found: $demo_script${NC}"
        return 1
    fi

    echo -e "${YELLOW}This will record the demo automatically.${NC}"
    echo -e "${YELLOW}The script will run in interactive mode.${NC}"
    echo ""
    read -r -p "Press Enter to start recording, or Ctrl+C to skip... "

    # Use expect to automate the demo and capture with asciinema
    local record_result=0
    if command -v expect &> /dev/null; then
        # Use expect for full automation
        echo -e "${GREEN}✓ Using expect for automated recording${NC}"
        # shellcheck disable=SC2086
        asciinema rec $OVERWRITE_FLAG "$output_file" -c "$SCRIPT_DIR/auto-demo-recorder.exp $demo_script"
        record_result=$?
    else
        # Fallback: manual pressing Enter
        echo -e "${YELLOW}⚠ Install 'expect' for fully automated recording${NC}"
        echo -e "${YELLOW}  macOS: brew install expect${NC}"
        echo -e "${YELLOW}  Linux: apt-get install expect${NC}"
        echo ""
        echo -e "${YELLOW}For now, you'll need to press Enter through the demo${NC}"
        echo ""
        read -r -p "Continue with manual mode? [y/N]: " -n 1
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            return 1
        fi
        # shellcheck disable=SC2086
        asciinema rec $OVERWRITE_FLAG "$output_file" -c "$demo_script"
        record_result=$?
    fi

    if [ $record_result -eq 0 ]; then
        echo ""
        echo -e "${GREEN}✓ Recording saved: $output_file${NC}"

        # Ask to review
        echo ""
        read -r -p "Review recording? [y/N]: " -n 1
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            asciinema play "$output_file"
        fi

        return 0
    else
        echo -e "${RED}❌ Recording failed${NC}"
        return 1
    fi
}

# Function to upload demo
upload_demo() {
    local demo_name=$1
    local cast_file="$VIDEOS_DIR/weave-cli-${demo_name}.cast"

    if [ ! -f "$cast_file" ]; then
        echo -e "${RED}❌ Error: Recording not found: $cast_file${NC}"
        return 1
    fi

    echo ""
    echo -e "${BLUE}Uploading ${demo_name} demo to asciinema.org...${NC}"

    # Upload and capture URL
    local upload_output
    upload_output=$(asciinema upload "$cast_file" 2>&1)
    local upload_result=$?

    if [ $upload_result -eq 0 ]; then
        # Extract URL from output
        local url
        url=$(echo "$upload_output" | grep -o 'https://asciinema.org/a/[^[:space:]]*' | head -1)

        if [ -n "$url" ]; then
            echo -e "${GREEN}✓ Uploaded successfully${NC}"
            echo -e "${GREEN}  URL: $url${NC}"

            # Save to tracking file
            {
                echo ""
                echo "# Upload on $(date '+%Y-%m-%d %H:%M:%S')"
                echo "# File: $cast_file"
                echo "# URL: $url"
                echo "${demo_name}: $url"
            } >> "$VIDEOS_DIR/latest-demo-uploads.txt"

            return 0
        fi
    fi

    echo -e "${RED}❌ Upload failed${NC}"
    return 1
}

# Main demo recording sequence
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo -e "${BLUE} Select demos to record:${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo ""
echo "1. Config Demo (always available)"
echo "2. Supabase Demo (requires Supabase configured)"
echo "3. Full Demo (5 min - requires Weaviate configured)"
echo "4. Quick Demo (2 min - requires Weaviate configured)"
echo "5. REPL Demo (requires OpenAI API key and weave-mcp)"
echo "6. All Available Demos"
echo ""
read -r -p "Select option [1-6]: " choice

case $choice in
    1)
        echo ""
        record_demo "config-demo" "$DEMOS_DIR/config-demo.sh"
        read -r -p "Upload to asciinema.org? [y/N]: " -n 1
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            upload_demo "config-demo"
        fi
        ;;
    2)
        if [ "$SUPABASE_OK" = true ]; then
            echo ""
            record_demo "supabase-demo" "$DEMOS_DIR/supabase-demo.sh"
            read -r -p "Upload to asciinema.org? [y/N]: " -n 1
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                upload_demo "supabase-demo"
            fi
        else
            echo -e "${RED}❌ Supabase not configured${NC}"
            echo "Set SUPABASE_DATABASE_URL and SUPABASE_DATABASE_KEY"
        fi
        ;;
    3)
        if [ "$WEAVIATE_OK" = true ]; then
            echo ""
            record_demo "full-demo" "$DEMOS_DIR/full-demo.sh"
            read -r -p "Upload to asciinema.org? [y/N]: " -n 1
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                upload_demo "full-demo"
            fi
        else
            echo -e "${RED}❌ Weaviate not configured${NC}"
            echo "Configure Weaviate or use mock database"
        fi
        ;;
    4)
        if [ "$WEAVIATE_OK" = true ]; then
            echo ""
            record_demo "quick-demo" "$DEMOS_DIR/quick-demo.sh"
            read -r -p "Upload to asciinema.org? [y/N]: " -n 1
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                upload_demo "quick-demo"
            fi
        else
            echo -e "${RED}❌ Weaviate not configured${NC}"
            echo "Configure Weaviate or use mock database"
        fi
        ;;
    5)
        # Check for OpenAI API key
        if [ -n "$OPENAI_API_KEY" ]; then
            echo ""
            record_demo "repl-demo" "$DEMOS_DIR/repl-demo.sh"
            read -r -p "Upload to asciinema.org? [y/N]: " -n 1
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                upload_demo "repl-demo"
            fi
        else
            echo -e "${RED}❌ OPENAI_API_KEY not set${NC}"
            echo "Set OPENAI_API_KEY environment variable"
        fi
        ;;
    6)
        # Record all available demos
        demos_recorded=0

        # Always record config demo
        echo ""
        if record_demo "config-demo" "$DEMOS_DIR/config-demo.sh"; then
            read -r -p "Upload config demo to asciinema.org? [y/N]: " -n 1
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                upload_demo "config-demo"
            fi
            ((demos_recorded++))
        fi

        # Record Supabase demo if available
        if [ "$SUPABASE_OK" = true ]; then
            echo ""
            if record_demo "supabase-demo" "$DEMOS_DIR/supabase-demo.sh"; then
                read -r -p "Upload Supabase demo to asciinema.org? [y/N]: " -n 1
                echo
                if [[ $REPLY =~ ^[Yy]$ ]]; then
                    upload_demo "supabase-demo"
                fi
                ((demos_recorded++))
            fi
        fi

        # Record Full and Quick demos if Weaviate is available
        if [ "$WEAVIATE_OK" = true ]; then
            echo ""
            if record_demo "full-demo" "$DEMOS_DIR/full-demo.sh"; then
                read -r -p "Upload Full demo to asciinema.org? [y/N]: " -n 1
                echo
                if [[ $REPLY =~ ^[Yy]$ ]]; then
                    upload_demo "full-demo"
                fi
                ((demos_recorded++))
            fi

            echo ""
            if record_demo "quick-demo" "$DEMOS_DIR/quick-demo.sh"; then
                read -r -p "Upload Quick demo to asciinema.org? [y/N]: " -n 1
                echo
                if [[ $REPLY =~ ^[Yy]$ ]]; then
                    upload_demo "quick-demo"
                fi
                ((demos_recorded++))
            fi
        fi

        # Record REPL demo if OpenAI API key is available
        if [ -n "$OPENAI_API_KEY" ]; then
            echo ""
            if record_demo "repl-demo" "$DEMOS_DIR/repl-demo.sh"; then
                read -r -p "Upload REPL demo to asciinema.org? [y/N]: " -n 1
                echo
                if [[ $REPLY =~ ^[Yy]$ ]]; then
                    upload_demo "repl-demo"
                fi
                ((demos_recorded++))
            fi
        fi

        echo ""
        echo -e "${GREEN}═══════════════════════════════════════════════════${NC}"
        echo -e "${GREEN}✓ Recorded $demos_recorded demos${NC}"
        echo -e "${GREEN}═══════════════════════════════════════════════════${NC}"
        ;;
    *)
        echo -e "${RED}Invalid choice${NC}"
        exit 1
        ;;
esac

echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo -e "${GREEN}✓ Demo recording complete!${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo ""
echo "📁 Recordings saved in: $VIDEOS_DIR"
echo ""
echo "Available .cast files:"
find "$VIDEOS_DIR" -name "*.cast" -type f -exec ls -lh {} \; 2>/dev/null | awk '{print "  • " $9 " (" $5 ")"}'
echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo -e "${BLUE} How to Review and Upload Recordings${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════${NC}"
echo ""
echo "📺 Review a recording:"
echo "  asciinema play videos/weave-cli-config-demo.cast"
echo ""
echo "🌐 Upload to asciinema.org:"
echo "  asciinema upload videos/weave-cli-config-demo.cast"
echo ""
echo "📋 Upload URLs are saved in:"
echo "  videos/latest-demo-uploads.txt"
echo ""
echo "📝 Update documentation with new URLs:"
echo "  1. Copy URLs from videos/latest-demo-uploads.txt"
echo "  2. Update docs/guides/DEMO.md"
echo "  3. Update README.md"
echo ""
echo "See docs/planning/DEMO_UPDATE_PLAN.md for complete details."
echo ""
