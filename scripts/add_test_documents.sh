#!/bin/bash

# Add Test Documents Script
# This script adds sample documents to the WEAVIATE_COLLECTION_TEST for testing purposes

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${BLUE}[SCRIPT]${NC} $1"
}

print_help() {
    echo -e "${BLUE}Add Test Documents Script${NC}"
    echo ""
    echo "Usage: ./scripts/add_test_documents.sh [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --help, -h     Show this help message"
    echo "  --dry-run      Show what would be added without actually adding"
    echo "  --collection   Specify collection name (default: from config)"
    echo ""
    echo "Environment Variables:"
    echo "  WEAVIATE_URL              Weaviate instance URL"
    echo "  WEAVIATE_API_KEY          Weaviate API key"
    echo "  WEAVIATE_COLLECTION_TEST  Test collection name"
    echo ""
    echo "Examples:"
    echo "  ./scripts/add_test_documents.sh"
    echo "  ./scripts/add_test_documents.sh --dry-run"
    echo "  ./scripts/add_test_documents.sh --collection MyTestCollection"
}

# Check if Go is installed
if ! command -v go >/dev/null 2>&1; then
    print_error "Go is not installed. Please install Go 1.21 or later."
    exit 1
fi

# Parse command line arguments
DRY_RUN=false
COLLECTION=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --help|-h)
            print_help
            exit 0
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --collection)
            COLLECTION="$2"
            shift 2
            ;;
        *)
            print_error "Unknown option: $1"
            echo ""
            print_help
            exit 1
            ;;
    esac
done

print_header "Adding Test Documents to Weaviate Collection"

# Check if Weaviate configuration is available
if [ -z "$WEAVIATE_URL" ] || [ -z "$WEAVIATE_API_KEY" ]; then
    print_warning "WEAVIATE_URL or WEAVIATE_API_KEY not set"
    print_status "Please set these environment variables:"
    echo "  export WEAVIATE_URL='your-weaviate-url.weaviate.cloud'"
    echo "  export WEAVIATE_API_KEY='your-api-key'"
    echo "  export WEAVIATE_COLLECTION_TEST='your-test-collection'"
    exit 1
fi

if [ -z "$WEAVIATE_COLLECTION_TEST" ]; then
    WEAVIATE_COLLECTION_TEST="WeaveDocs_test"
    print_warning "WEAVIATE_COLLECTION_TEST not set, using default: $WEAVIATE_COLLECTION_TEST"
fi

if [ -n "$COLLECTION" ]; then
    WEAVIATE_COLLECTION_TEST="$COLLECTION"
    print_status "Using specified collection: $WEAVIATE_COLLECTION_TEST"
fi

print_status "Configuration:"
echo "  URL: $WEAVIATE_URL"
echo "  Collection: $WEAVIATE_COLLECTION_TEST"
echo "  API Key: ${WEAVIATE_API_KEY:0:8}..."

if [ "$DRY_RUN" = true ]; then
    print_status "DRY RUN MODE - No documents will be added"
    print_status "Documents that would be added:"
    echo "  1. test-doc-1: This is a test document for Weave CLI testing..."
    echo "  2. test-doc-2: Another test document with different content..."
    echo "  3. test-doc-3: A third test document for comprehensive testing..."
    echo "  4. test-doc-4: Document with special characters..."
    echo "  5. test-doc-5: Large document content..."
    exit 0
fi

# Build and run the Go script
print_status "Building document addition script..."
if ! go build -o /tmp/add_test_documents ./scripts/add_test_documents.go; then
    print_error "Failed to build document addition script"
    exit 1
fi

print_status "Running document addition script..."
if ! /tmp/add_test_documents; then
    print_error "Failed to add test documents"
    exit 1
fi

# Clean up
rm -f /tmp/add_test_documents

print_status "Test documents added successfully!"
print_status "You can now run integration tests with:"
echo "  ./test.sh integration"
echo "  go test -v ./tests/... -run=\"TestWeaviate\""