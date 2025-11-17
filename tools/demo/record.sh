#!/usr/bin/env bash

# Weave CLI Asciinema Recording Tool
# Usage: ./tools/asciinema.sh [command]

# Load environment variables
if [ -f ".env" ]; then
    # shellcheck disable=SC1091
    source .env
fi

# Set up environment variables for demo (if not already set)
if [ -z "$VECTOR_DB_TYPE" ]; then
    export VECTOR_DB_TYPE="weaviate-cloud"
fi

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_header() {
    echo -e "${BLUE}[ASCII]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_help() {
    echo -e "${BLUE}Weave CLI Asciinema Recording Tool${NC}"
    echo ""
    echo "Usage: ./tools/asciinema.sh [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  demo        Record the full demo (5 minutes)"
    echo "  quick       Record a quick 2-minute demo"
    echo "  repl        Record REPL AI-assisted demo (NEW!)"
    echo "  install     Install asciinema if not available"
    echo "  upload [FILE] Upload recording to asciinema.org (or latest if no file specified)"
    echo "  list        List available recordings"
    echo "  clean       Clean up old recordings"
    echo "  --help, -h  Show this help message"
    echo ""
    echo "Examples:"
    echo "  ./tools/asciinema.sh demo     # Record full demo"
    echo "  ./tools/asciinema.sh quick    # Record quick demo"
    echo "  ./tools/asciinema.sh repl     # Record REPL AI demo"
    echo "  ./tools/asciinema.sh upload   # Upload latest recording to asciinema.org"
    echo "  ./tools/asciinema.sh upload videos/weave-cli-quick-demo.cast  # Upload specific file"
    echo ""
    echo "Prerequisites:"
    echo "  • Weaviate Cloud instance configured"
    echo "  • Test collections available (DemoCollection, DemoCollectionImages)"
    echo "  • Demo documents in docs/ and images/ directories"
    echo "  • PDF files in tests/fixtures/ directory"
}

# Function to check if asciinema is installed
check_asciinema() {
    if ! command -v asciinema &> /dev/null; then
        print_warning "asciinema is not installed"
        echo "Install it with: ./tools/asciinema.sh install"
        return 1
    fi
    return 0
}

# Function to install asciinema
install_asciinema() {
    print_header "Installing asciinema..."
    
    if command -v brew &> /dev/null; then
        print_header "Installing via Homebrew..."
        brew install asciinema
    elif command -v pip3 &> /dev/null; then
        print_header "Installing via pip3..."
        pip3 install asciinema
    elif command -v pip &> /dev/null; then
        print_header "Installing via pip..."
        pip install asciinema
    else
        print_error "No package manager found (brew, pip, pip3)"
        echo "Please install asciinema manually: https://asciinema.org/docs/installation"
        return 1
    fi
    
    if check_asciinema; then
        print_success "asciinema installed successfully!"
    else
        print_error "Failed to install asciinema"
        return 1
    fi
}

# Function to create demo script
create_demo_script() {
    local script_type="$1"
    local script_file="/tmp/weave_demo_${script_type}.sh"
    
    cat > "$script_file" << 'EOF'
#!/usr/bin/env bash

# Weave CLI Demo Script for Asciinema Recording
# This script runs the demo commands with proper timing

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Function to run command with timing
run_demo_cmd() {
    local cmd="$1"
    local description="$2"
    local delay="${3:-2}"
    
    echo -e "${BLUE}💻 ${description}${NC}"
    echo -e "${YELLOW}$ ${cmd}${NC}"
    sleep 1
    eval "$cmd"
    echo ""
    sleep "$delay"
}

# Function to add page break
page_break() {
    local page="$1"
    echo ""
    echo -e "${GREEN}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${GREEN}                    PAGE ${page}${NC}"
    echo -e "${GREEN}═══════════════════════════════════════════════════════════════${NC}"
    echo ""
    sleep 2
}

# Start demo
echo -e "${GREEN}🚀 Weave CLI Demo Starting...${NC}"
echo ""
sleep 2

# Pre-demo cleanup
echo -e "${BLUE}💻 Pre-demo cleanup${NC}"
echo -e "${YELLOW}$ ./bin/weave cols delete-schema DemoCollection --force 2>/dev/null || true${NC}"
sleep 1
./bin/weave cols delete-schema DemoCollection --force 2>/dev/null || true
echo -e "${YELLOW}$ ./bin/weave cols delete-schema DemoCollectionImages --force 2>/dev/null || true${NC}"
sleep 1
./bin/weave cols delete-schema DemoCollectionImages --force 2>/dev/null || true
echo ""
sleep 2

# Page 1: Configuration & Health Check
page_break "1"
run_demo_cmd "./bin/weave config show" "Configuration Display (Environment Variables)"
run_demo_cmd "./bin/weave health check" "Health Check"
run_demo_cmd "./bin/weave --help | head -20" "Help Command"

# Page 2: Create Collections
page_break "2"
run_demo_cmd "./bin/weave cols create DemoCollection --text --flat-metadata --embedding-model text-embedding-3-small || echo 'Collection already exists - continuing demo'" "Create Text Collection"
run_demo_cmd "./bin/weave cols create DemoCollectionImages --image --flat-metadata --embedding-model text-embedding-3-small || echo 'Collection already exists - continuing demo'" "Create Image Collection"
run_demo_cmd "./bin/weave cols show DemoCollection" "Show Collection Structure"
run_demo_cmd "./bin/weave cols show DemoCollectionImages --schema" "Show Image Collection Schema"

# Page 3: List Collections
page_break "3"
run_demo_cmd "./bin/weave cols ls" "List All Collections"

# Page 4: Create Documents
page_break "4"
run_demo_cmd "if [ -f README.md ]; then ./bin/weave docs create DemoCollection README.md || echo 'Document creation failed - continuing demo'; else echo 'README.md not found - creating sample document'; echo '# Sample Document\n\nThis is a sample document for the demo.\n\n## Features\n- Vector embeddings\n- Semantic search\n- Document management' > README.md && ./bin/weave docs create DemoCollection README.md || echo 'Document creation failed - continuing demo'; fi" "Create Text Document"
run_demo_cmd "if [ -f tests/fixtures/ragme-io.pdf ]; then ./bin/weave docs create DemoCollection tests/fixtures/ragme-io.pdf || echo 'PDF document creation failed - continuing demo'; else echo 'ℹ️ No PDF file found - skipping PDF document creation'; fi" "Create PDF Document (NEW!)"
run_demo_cmd "if [ -f images/weave-cli_1.png ]; then if ./bin/weave docs create DemoCollectionImages images/weave-cli_1.png >/dev/null 2>&1; then echo '✅ Image document created successfully'; else echo 'ℹ️ Image too large for embedding model - this is expected for large images'; fi; else echo 'ℹ️ No image file found - skipping image document creation'; fi" "Create Image Document"

# Page 5: Show Documents & Schema
page_break "5"
run_demo_cmd "./bin/weave docs show DemoCollection --name README.md || echo 'Document not found - will show collection info instead'" "Show Document Details"
run_demo_cmd "./bin/weave cols show DemoCollection --schema" "Show Collection Schema"
run_demo_cmd "./bin/weave cols show DemoCollection --expand-metadata" "Show Collection Metadata Analysis"

# Page 6: List Documents
page_break "6"
run_demo_cmd "./bin/weave cols ls | grep DemoCollection || echo 'DemoCollection collection not found'" "Verify Collection Exists"
run_demo_cmd "./bin/weave docs ls DemoCollection" "Simple Document List"
run_demo_cmd "./bin/weave docs ls DemoCollection -w -S" "Virtual Document View with Summary"

# Page 7: Semantic Search & Query
page_break "7"
run_demo_cmd "./bin/weave cols q DemoCollection 'weave-cli installation'" "Basic Semantic Search"
run_demo_cmd "./bin/weave cols q DemoCollection 'machine learning' --top_k 3" "Search with Custom Result Limit"
run_demo_cmd "./bin/weave cols q DemoCollection 'ragme.io' --search-metadata" "Search PDF Content (NEW!)"
run_demo_cmd "./bin/weave cols q DemoCollection 'maximilien.org' --search-metadata" "Search with Metadata"
run_demo_cmd "./bin/weave cols q --help | head -15" "Query Help"

# Page 8: Delete Documents
page_break "8"
run_demo_cmd "./bin/weave docs delete DemoCollection --name README.md --force" "Delete Document with Force"

# Page 9: Cleanup Operations
page_break "9"
run_demo_cmd "./bin/weave docs delete-all DemoCollection --force" "Delete All Documents"
run_demo_cmd "./bin/weave cols delete-schema DemoCollection --force" "Delete Collection Schema"

# Page 10: Getting Weave CLI
page_break "10"
echo -e "${BLUE}💻 Getting Weave CLI${NC}"
echo -e "${YELLOW}# Download from GitHub releases${NC}"
echo -e "${YELLOW}# Build from source: git clone && ./build.sh${NC}"
echo -e "${YELLOW}# MIT License - Free for commercial use${NC}"
echo -e "${YELLOW}# Built with ❤️ by github.com/maximilien${NC}"
echo ""

# Page 11: Thank You
page_break "11"
run_demo_cmd "echo '🎉 Demo completed successfully!'" "Demo Complete"
run_demo_cmd "./bin/weave --version" "Version Information"

echo -e "${GREEN}🎉 Thank you for watching!${NC}"
echo -e "${BLUE}Repository: https://github.com/maximilien/weave-cli${NC}"
EOF

    chmod +x "$script_file"
    echo "$script_file"
}

# Function to record full demo
record_demo() {
    local demo_type="$1"
    local output_file="videos/weave-cli-${demo_type}-demo.cast"
    
    print_header "Recording ${demo_type} demo..."
    
    # Check if we have the required environment variables for Weaviate Cloud
    if [ -z "$WEAVIATE_URL" ] || [ -z "$WEAVIATE_API_KEY" ] || [ -z "$OPENAI_API_KEY" ]; then
        print_error "❌ Required environment variables not set for Weaviate Cloud demo"
        print_error "Please set the following environment variables:"
        echo "  export WEAVIATE_URL=\"https://your-cluster.weaviate.cloud\""
        echo "  export WEAVIATE_API_KEY=\"your-weaviate-api-key\""
        echo "  export OPENAI_API_KEY=\"sk-proj-your-openai-api-key\""
        echo ""
        print_error "Demo cannot proceed without Weaviate credentials."
        print_error "For testing without Weaviate, use: VECTOR_DB_TYPE=\"mock\" ./tools/asciinema.sh quick"
        return 1
    fi
    
    if ! check_asciinema; then
        return 1
    fi
    
    # Create videos directory if it doesn't exist
    mkdir -p videos
    
    # Create demo script
    local script_file
    script_file=$(create_demo_script "$demo_type")
    
    # Record the demo
    print_header "Starting recording... (Press Ctrl+C to stop)"
    echo "Recording will be saved to: $output_file"
    echo ""
    
    asciinema rec "$output_file" --command "$script_file"
    
    # Clean up script
    rm -f "$script_file"
    
    if [ -f "$output_file" ]; then
        print_success "Demo recorded successfully: $output_file"
        print_header "To play the recording:"
        echo "  asciinema play $output_file"
        echo ""
        print_header "To upload to asciinema.org:"
        echo "  asciinema upload $output_file"
    else
        print_error "Recording failed"
        return 1
    fi
}

# Function to record quick demo
record_quick_demo() {
    local output_file="videos/weave-cli-quick-demo.cast"
    
    print_header "Recording quick demo..."
    
    # Check if we have the required environment variables for Weaviate Cloud
    if [ -z "$WEAVIATE_URL" ] || [ -z "$WEAVIATE_API_KEY" ] || [ -z "$OPENAI_API_KEY" ]; then
        print_error "❌ Required environment variables not set for Weaviate Cloud demo"
        print_error "Please set the following environment variables:"
        echo "  export WEAVIATE_URL=\"https://your-cluster.weaviate.cloud\""
        echo "  export WEAVIATE_API_KEY=\"your-weaviate-api-key\""
        echo "  export OPENAI_API_KEY=\"sk-proj-your-openai-api-key\""
        echo ""
        print_error "Demo cannot proceed without Weaviate credentials."
        print_error "For testing without Weaviate, use: VECTOR_DB_TYPE=\"mock\" ./tools/asciinema.sh quick"
        return 1
    fi
    
    if ! check_asciinema; then
        return 1
    fi
    
    mkdir -p videos
    
    # Create quick demo script
    local script_file="/tmp/weave_quick_demo.sh"
    
    cat > "$script_file" << 'EOF'
#!/usr/bin/env bash

# Quick Weave CLI Demo (2 minutes)
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}🚀 Weave CLI Quick Demo${NC}"
echo ""
sleep 2

echo -e "${BLUE}💻 Configuration (Environment Variables)${NC}"
echo -e "${YELLOW}$ ./bin/weave config show${NC}"
sleep 1
./bin/weave config show
echo ""
sleep 2

echo -e "${BLUE}💻 Health Check${NC}"
echo -e "${YELLOW}$ ./bin/weave health check${NC}"
sleep 1
./bin/weave health check
echo ""
sleep 2

echo -e "${BLUE}💻 List Collections${NC}"
echo -e "${YELLOW}$ ./bin/weave cols ls${NC}"
sleep 1
./bin/weave cols ls
echo ""
sleep 2

echo -e "${BLUE}🧹 Cleanup${NC}"
echo -e "${YELLOW}$ ./bin/weave cols delete-schema DemoCollection --force${NC}"
sleep 1
# Check if collection exists before trying to delete
if ./bin/weave cols show DemoCollection --vector-db-type weaviate-cloud >/dev/null 2>&1; then
    ./bin/weave cols delete-schema DemoCollection --force 2>/dev/null || true
else
    echo "Collection DemoCollection does not exist, skipping cleanup"
fi
echo ""
sleep 2

echo -e "${BLUE}💻 Create Collection${NC}"
echo -e "${YELLOW}$ ./bin/weave cols create DemoCollection --text --flat-metadata${NC}"
sleep 1
./bin/weave cols create DemoCollection --text --flat-metadata --embedding-model text-embedding-3-small || echo "Collection may already exist"
echo ""
sleep 2

echo -e "${BLUE}💻 Create Document${NC}"
echo -e "${YELLOW}$ ./bin/weave docs create DemoCollection README.md${NC}"
sleep 1
if [ -f README.md ]; then
    ./bin/weave docs create DemoCollection README.md || echo "Document creation may have failed - continuing demo"
else
    echo "Creating sample document for demo..."
    echo "# Sample Document

This is a sample document for the Weave CLI demo.
It contains some text that we can search through using semantic search.

## Features
- Vector embeddings
- Semantic search
- Document management
- Collection management" > README.md
    ./bin/weave docs create DemoCollection README.md || echo "Document creation may have failed - continuing demo"
fi
echo ""
sleep 2

echo -e "${BLUE}💻 Create PDF Document (NEW!)${NC}"
echo -e "${YELLOW}$ ./bin/weave docs create DemoCollection tests/fixtures/ragme-io.pdf${NC}"
sleep 1
if [ -f tests/fixtures/ragme-io.pdf ]; then
    ./bin/weave docs create DemoCollection tests/fixtures/ragme-io.pdf || echo "PDF document creation may have failed - continuing demo"
else
    echo "ℹ️ No PDF file found - skipping PDF document creation"
fi
echo ""
sleep 2

echo -e "${BLUE}💻 List Documents${NC}"
echo -e "${YELLOW}$ ./bin/weave docs ls DemoCollection${NC}"
sleep 1
./bin/weave docs ls DemoCollection
echo ""
sleep 2

echo -e "${BLUE}💻 Semantic Search${NC}"
echo -e "${YELLOW}$ ./bin/weave cols q DemoCollection 'sample document'${NC}"
sleep 1
./bin/weave cols q DemoCollection "sample document"
echo ""
sleep 2

echo -e "${BLUE}💻 Search PDF Content (NEW!)${NC}"
echo -e "${YELLOW}$ ./bin/weave cols q DemoCollection 'ragme.io' --search-metadata${NC}"
sleep 1
./bin/weave cols q DemoCollection "ragme.io" --search-metadata
echo ""
sleep 2

echo -e "${BLUE}💻 Search with Metadata${NC}"
echo -e "${YELLOW}$ ./bin/weave cols q DemoCollection 'README' --search-metadata${NC}"
sleep 1
./bin/weave cols q DemoCollection "README" --search-metadata
echo ""
sleep 2

echo -e "${BLUE}💻 BM25 Keyword Search${NC}"
echo -e "${YELLOW}$ ./bin/weave cols q DemoCollection 'sample' --bm25${NC}"
sleep 1
./bin/weave cols q DemoCollection "sample" --bm25
echo ""
sleep 2

echo -e "${BLUE}💻 Cleanup${NC}"
echo -e "${YELLOW}$ ./bin/weave cols delete-schema DemoCollection --force${NC}"
sleep 1
./bin/weave cols delete-schema DemoCollection --force
echo ""
sleep 2

echo -e "${GREEN}🎉 Quick demo completed!${NC}"
echo -e "${BLUE}Repository: https://github.com/maximilien/weave-cli${NC}"
echo -e "${BLUE}License: MIT - Free for commercial use${NC}"
EOF

    chmod +x "$script_file"
    
    # Record the quick demo
    print_header "Starting quick demo recording..."
    asciinema rec "$output_file" --command "$script_file"
    
    # Clean up
    rm -f "$script_file"
    
    if [ -f "$output_file" ]; then
        print_success "Quick demo recorded: $output_file"
    else
        print_error "Quick demo recording failed"
        return 1
    fi
}

# Function to upload recording
upload_recording() {
    local file_to_upload="$1"

    # If no file specified, find the latest recording
    if [ -z "$file_to_upload" ]; then
        # Use macOS-compatible approach to find latest .cast file
        # Find the most recent .cast file (macOS compatible)
        file_to_upload=$(find videos -name "*.cast" -type f -exec stat -f "%m %N" {} \; 2>/dev/null | sort -nr | head -1 | cut -d' ' -f2-)

        if [ -z "$file_to_upload" ]; then
            print_error "No recordings found in videos/ directory"
            return 1
        fi

        print_header "Uploading latest recording: $file_to_upload"
    else
        # Check if the specified file exists
        if [ ! -f "$file_to_upload" ]; then
            print_error "File not found: $file_to_upload"
            return 1
        fi

        print_header "Uploading specified recording: $file_to_upload"
    fi

    if ! check_asciinema; then
        return 1
    fi

    # Create videos directory if it doesn't exist
    mkdir -p videos

    # Upload and capture the URL
    print_header "Uploading to asciinema.org..."
    local upload_output
    upload_output=$(asciinema upload "$file_to_upload" 2>&1)
    local upload_status=$?

    # Display the upload output
    echo "$upload_output"

    if [ $upload_status -eq 0 ]; then
        # Extract the URL from the output
        local upload_url
        upload_url=$(echo "$upload_output" | grep -o 'https://asciinema.org/a/[^[:space:]]*')

        if [ -n "$upload_url" ]; then
            # Determine the demo type from filename
            local demo_type="unknown"
            if [[ "$file_to_upload" == *"quick"* ]]; then
                demo_type="quick"
            elif [[ "$file_to_upload" == *"full"* ]]; then
                demo_type="full"
            fi

            # Save to latest-demo-uploads.txt
            local upload_file="videos/latest-demo-uploads.txt"
            local template_file="videos/latest-demo-uploads.txt.template"
            local timestamp
            timestamp=$(date "+%Y-%m-%d %H:%M:%S")

            # Create from template if it doesn't exist
            if [ ! -f "$upload_file" ]; then
                if [ -f "$template_file" ]; then
                    cp "$template_file" "$upload_file"
                else
                    echo "# Weave CLI Demo Uploads" > "$upload_file"
                    echo "# Latest uploads for quick and full demos" >> "$upload_file"
                    echo "" >> "$upload_file"
                fi
            fi

            # Update or add the entry for this demo type
            if grep -q "^${demo_type}:" "$upload_file" 2>/dev/null; then
                # Update existing entry (macOS compatible)
                if [[ "$OSTYPE" == "darwin"* ]]; then
                    sed -i '' "s|^${demo_type}:.*|${demo_type}: ${upload_url}|" "$upload_file"
                else
                    sed -i "s|^${demo_type}:.*|${demo_type}: ${upload_url}|" "$upload_file"
                fi
            else
                # Add new entry
                echo "${demo_type}: ${upload_url}" >> "$upload_file"
            fi

            # Also add a timestamped entry
            echo "" >> "$upload_file"
            echo "# Upload on ${timestamp}" >> "$upload_file"
            echo "# File: ${file_to_upload}" >> "$upload_file"
            echo "# URL: ${upload_url}" >> "$upload_file"

            print_success "Upload URL saved to: $upload_file"
            print_success "Demo type: $demo_type"
            print_success "URL: $upload_url"
        fi
    else
        print_error "Upload failed"
        return 1
    fi
}

# Function to list recordings
list_recordings() {
    print_header "Available recordings:"
    
    if [ ! -d "videos" ] || [ -z "$(ls -A videos/*.cast 2>/dev/null)" ]; then
        print_warning "No recordings found"
        return 0
    fi
    
    for file in videos/*.cast; do
        if [ -f "$file" ]; then
            local size
            size=$(du -h "$file" | cut -f1)
            local date
            date=$(stat -f "%Sm" -t "%Y-%m-%d %H:%M" "$file" 2>/dev/null || stat -c "%y" "$file" | cut -d' ' -f1-2)
            echo "  📹 $(basename "$file") (${size}, ${date})"
        fi
    done
}

# Function to clean old recordings
clean_recordings() {
    print_header "Cleaning old recordings..."
    
    if [ ! -d "videos" ]; then
        print_warning "No videos directory found"
        return 0
    fi
    
    local count
    count=$(find videos -name "*.cast" -type f | wc -l)
    
    if [ "$count" -eq 0 ]; then
        print_warning "No recordings to clean"
        return 0
    fi
    
    print_header "Found $count recordings"
    echo "This will remove all .cast files from videos/ directory"
    read -p "Are you sure? (y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -f videos/*.cast
        print_success "Cleaned $count recordings"
    else
        print_warning "Cleanup cancelled"
    fi
}

# Function to record REPL demo
record_repl_demo() {
    local output_file="videos/weave-cli-repl-demo.cast"

    print_header "Recording REPL AI-assisted demo..."

    # Check if we have the required environment variables
    if [ -z "$WEAVIATE_URL" ] || [ -z "$WEAVIATE_API_KEY" ] || [ -z "$OPENAI_API_KEY" ]; then
        print_error "❌ Required environment variables not set"
        print_error "Please set: WEAVIATE_URL, WEAVIATE_API_KEY, OPENAI_API_KEY"
        return 1
    fi

    # Check for WEAVE_MCP_STDIO_PATH
    if [ -z "$WEAVE_MCP_STDIO_PATH" ]; then
        print_error "❌ WEAVE_MCP_STDIO_PATH not set"
        print_error "Please set the path to your weave-mcp stdio binary"
        return 1
    fi

    if ! check_asciinema; then
        return 1
    fi

    mkdir -p videos

    # Record using the batch query mode
    print_header "Starting REPL demo recording..."
    print_header "This will execute queries from videos/weave-repl-demo-queries.txt"
    echo ""

    asciinema rec "$output_file" --command "./bin/weave --query-strings videos/weave-repl-demo-queries.txt --no-confirm"

    if [ -f "$output_file" ]; then
        print_success "REPL demo recorded: $output_file"
        print_header "To play the recording:"
        echo "  asciinema play $output_file"
        echo ""
        print_header "To upload to asciinema.org:"
        echo "  asciinema upload $output_file"
    else
        print_error "REPL demo recording failed"
        return 1
    fi
}

# Main script logic
case "${1:-help}" in
    "demo")
        record_demo "full"
        ;;
    "quick")
        record_quick_demo
        ;;
    "repl")
        record_repl_demo
        ;;
    "install")
        install_asciinema
        ;;
    "upload")
        upload_recording "$2"
        ;;
    "list")
        list_recordings
        ;;
    "clean")
        clean_recordings
        ;;
    "--help"|"-h"|"help")
        print_help
        ;;
    *)
        print_error "Unknown command: $1"
        print_help
        exit 1
        ;;
esac