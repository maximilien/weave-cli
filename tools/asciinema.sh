#!/usr/bin/env bash

# Weave CLI Asciinema Recording Tool
# Usage: ./tools/asciinema.sh [command]

# Load environment variables
if [ -f ".env" ]; then
    # shellcheck disable=SC1091
    source .env
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
    echo "  install     Install asciinema if not available"
    echo "  upload [FILE] Upload recording to asciinema.org (or latest if no file specified)"
    echo "  list        List available recordings"
    echo "  clean       Clean up old recordings"
    echo "  --help, -h  Show this help message"
    echo ""
    echo "Examples:"
    echo "  ./tools/asciinema.sh demo     # Record full demo"
    echo "  ./tools/asciinema.sh quick    # Record quick demo"
    echo "  ./tools/asciinema.sh upload   # Upload latest recording to asciinema.org"
    echo "  ./tools/asciinema.sh upload videos/weave-cli-quick-demo.cast  # Upload specific file"
    echo ""
    echo "Prerequisites:"
    echo "  • Weaviate Cloud instance configured"
    echo "  • Test collections available (WeaveDocs, WeaveImages)"
    echo "  • Demo documents in docs/ and images/ directories"
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
echo -e "${YELLOW}$ ./bin/weave cols delete-schema WeaveDocs --force 2>/dev/null || true${NC}"
sleep 1
./bin/weave cols delete-schema WeaveDocs --force 2>/dev/null || true
echo -e "${YELLOW}$ ./bin/weave cols delete-schema WeaveImages --force 2>/dev/null || true${NC}"
sleep 1
./bin/weave cols delete-schema WeaveImages --force 2>/dev/null || true
echo ""
sleep 2

# Page 1: Health Check & Configuration
page_break "1"
run_demo_cmd "./bin/weave health check" "Health Check"
run_demo_cmd "./bin/weave config show" "Configuration Display"
run_demo_cmd "./bin/weave --help | head -20" "Help Command"

# Page 2: Create Collections
page_break "2"
run_demo_cmd "./bin/weave cols create WeaveDocs --text --flat-metadata --embedding-model text-embedding-3-small || echo 'Collection already exists'" "Create Text Collection"
run_demo_cmd "./bin/weave cols create WeaveImages --image --flat-metadata --embedding-model text-embedding-3-small || echo 'Collection already exists'" "Create Image Collection"
run_demo_cmd "./bin/weave cols show WeaveDocs" "Show Collection Structure"
run_demo_cmd "./bin/weave cols show WeaveImages --schema" "Show Image Collection Schema"

# Page 3: List Collections
page_break "3"
run_demo_cmd "./bin/weave cols ls" "List All Collections"

# Page 4: Create Documents
page_break "4"
run_demo_cmd "if [ -f README.md ]; then ./bin/weave docs create WeaveDocs README.md; else echo 'README.md not found - creating sample document'; echo '# Sample Document\n\nThis is a sample document for the demo.' > README.md && ./bin/weave docs create WeaveDocs README.md; fi" "Create Text Document"
run_demo_cmd "if ./bin/weave docs create WeaveImages images/weave-cli_1.png >/dev/null 2>&1; then echo '✅ Image document created successfully'; else echo 'ℹ️ Image too large for embedding model - this is expected for large images'; fi" "Create Image Document"

# Page 5: Show Documents & Schema
page_break "5"
run_demo_cmd "./bin/weave docs show WeaveDocs --name README.md || echo 'Document not found - will show collection info instead'" "Show Document Details"
run_demo_cmd "./bin/weave cols show WeaveDocs --schema" "Show Collection Schema"
run_demo_cmd "./bin/weave cols show WeaveDocs --expand-metadata" "Show Collection Metadata Analysis"

# Page 6: List Documents
page_break "6"
run_demo_cmd "./bin/weave cols ls | grep WeaveDocs || echo 'WeaveDocs collection not found'" "Verify Collection Exists"
run_demo_cmd "./bin/weave docs ls WeaveDocs" "Simple Document List"
run_demo_cmd "./bin/weave docs ls WeaveDocs -w -S" "Virtual Document View with Summary"

# Page 7: Semantic Search & Query
page_break "7"
run_demo_cmd "./bin/weave cols q WeaveDocs 'weave-cli installation'" "Basic Semantic Search"
run_demo_cmd "./bin/weave cols q WeaveDocs 'machine learning' --top_k 3" "Search with Custom Result Limit"
run_demo_cmd "./bin/weave cols q WeaveDocs 'maximilien.org' --search-metadata" "Search with Metadata (NEW!)"
run_demo_cmd "./bin/weave cols q --help | head -15" "Query Help"

# Page 8: Delete Documents
page_break "8"
run_demo_cmd "./bin/weave docs delete WeaveDocs --name README.md --force" "Delete Document with Force"

# Page 9: Cleanup Operations
page_break "9"
run_demo_cmd "./bin/weave docs delete-all WeaveDocs --force" "Delete All Documents"
run_demo_cmd "./bin/weave cols delete-schema WeaveDocs --force" "Delete Collection Schema"

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
./bin/weave cols create DemoCollection --text --flat-metadata --embedding-model text-embedding-3-small
echo ""
sleep 2

echo -e "${BLUE}💻 Create Document${NC}"
echo -e "${YELLOW}$ ./bin/weave docs create DemoCollection README.md${NC}"
sleep 1
./bin/weave docs create DemoCollection README.md
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
    
    asciinema upload "$file_to_upload"
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

# Main script logic
case "${1:-help}" in
    "demo")
        record_demo "full"
        ;;
    "quick")
        record_quick_demo
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