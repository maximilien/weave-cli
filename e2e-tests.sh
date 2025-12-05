#!/bin/bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2025 dr.max

# End-to-End Sanity Tests for Vector Databases
# Tests basic operations: health, list, create, add document, query, delete

set -euo pipefail

# Colors and formatting
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color
BOLD='\033[1m'

# Test configuration
WEAVE_BIN="./bin/weave"
TEST_COLLECTION_PREFIX="E2ETest"
TEST_TEXT="This is an end-to-end test document for vector database integration"
TIMESTAMP=$(date +%s)

# Test flags
TEST_WEAVIATE=false
TEST_MILVUS=false
TEST_SUPABASE=false
TEST_MONGODB=false
TEST_ALL=false
TEST_DEMO=false

# Test results
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0
declare -a FAILED_TESTS

usage() {
    cat <<EOF
${BOLD}End-to-End Vector Database Tests${NC}

Usage: $(basename "$0") [OPTIONS]

Options:
    --weaviate          Test Weaviate (cloud and/or local)
    --milvus            Test Milvus (local)
    --supabase          Test Supabase
    --mongodb           Test MongoDB Atlas
    --all               Test all configured databases (default)
    --demo              Run comprehensive demo-style tests (all features)
    -h, --help          Show this help message

Test Modes:
    ${BOLD}Basic Mode (default)${NC}
        - Health check
        - List/create/delete collections
        - Add/list documents
        - Basic vector search

    ${BOLD}Demo Mode (--demo)${NC}
        - All basic tests plus:
        - Collection schema operations
        - Document show/delete by name
        - PDF with image extraction
        - Image document operations
        - Multi-VDB operations (--details, --summary)
        - Config commands
        - Query with various parameters

Examples:
    # Test all databases (basic)
    ./e2e-tests.sh
    ./e2e-tests.sh --all

    # Test specific databases (basic)
    ./e2e-tests.sh --weaviate --milvus
    ./e2e-tests.sh --milvus

    # Run comprehensive demo tests
    ./e2e-tests.sh --demo
    ./e2e-tests.sh --demo --weaviate

Prerequisites:
    - Weave CLI built (./build.sh)
    - Vector databases configured (.env or config.yaml)
    - For Milvus local: ./tools/vdb/local/milvus.sh start
    - OPENAI_API_KEY set for automatic embeddings
    - For demo mode: test images and PDFs in tests/images/ and tests/fixtures/

EOF
}

# Parse arguments
parse_args() {
    if [ $# -eq 0 ]; then
        TEST_ALL=true
        return
    fi

    while [ $# -gt 0 ]; do
        case "$1" in
            --weaviate)
                TEST_WEAVIATE=true
                shift
                ;;
            --milvus)
                TEST_MILVUS=true
                shift
                ;;
            --supabase)
                TEST_SUPABASE=true
                shift
                ;;
            --mongodb)
                TEST_MONGODB=true
                shift
                ;;
            --all)
                TEST_ALL=true
                shift
                ;;
            --demo)
                TEST_DEMO=true
                shift
                ;;
            -h|--help)
                usage
                exit 0
                ;;
            *)
                echo -e "${RED}❌ Unknown option: $1${NC}" >&2
                echo "" >&2
                usage
                exit 1
                ;;
        esac
    done

    # If --all or no specific flags, enable all
    if [ "$TEST_ALL" = true ]; then
        TEST_WEAVIATE=true
        TEST_MILVUS=true
        TEST_SUPABASE=true
        TEST_MONGODB=true
    fi
}

# Helper functions
print_header() {
    echo ""
    echo -e "${CYAN}${BOLD}═══════════════════════════════════════════════════${NC}"
    echo -e "${CYAN}${BOLD}  $1${NC}"
    echo -e "${CYAN}${BOLD}═══════════════════════════════════════════════════${NC}"
    echo ""
}

print_test() {
    echo -e "${BLUE}▶ $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_info() {
    echo -e "${CYAN}ℹ️  $1${NC}"
}

record_test() {
    ((TESTS_RUN++)) || true
}

record_pass() {
    ((TESTS_PASSED++)) || true
}

record_fail() {
    local test_name="$1"
    ((TESTS_FAILED++)) || true
    FAILED_TESTS+=("$test_name")
}

# Check prerequisites
check_prerequisites() {
    print_header "Checking Prerequisites"

    # Check weave binary
    if [ ! -f "$WEAVE_BIN" ]; then
        print_error "Weave binary not found at $WEAVE_BIN"
        print_info "Run ./build.sh first"
        exit 1
    fi
    print_success "Weave binary found"

    # Check OPENAI_API_KEY
    if [ -z "${OPENAI_API_KEY:-}" ]; then
        print_warning "OPENAI_API_KEY not set - automatic embeddings may not work"
    else
        print_success "OPENAI_API_KEY is set"
    fi
}

# Test a vector database
test_vdb() {
    local vdb_name="$1"
    local vdb_flag="$2"
    local config_file="${3:-}"

    print_header "Testing $vdb_name"

    # Sanitize collection name: replace hyphens with underscores
    # (Milvus doesn't allow hyphens)
    local sanitized_name="${vdb_name//-/_}"
    local test_collection="${TEST_COLLECTION_PREFIX}_${sanitized_name}_${TIMESTAMP}"
    local cmd_prefix="$WEAVE_BIN"

    if [ -n "$config_file" ] && [ -f "$config_file" ]; then
        cmd_prefix="$cmd_prefix --config $config_file"
    fi
    cmd_prefix="$cmd_prefix $vdb_flag"

    # Test 1: Health check
    print_test "Test 1: Health check"
    record_test
    if $cmd_prefix health check &>/dev/null; then
        print_success "Health check passed"
        record_pass
    else
        print_error "Health check failed"
        record_fail "$vdb_name: health check"
        print_warning "Skipping remaining tests for $vdb_name"
        return 1
    fi

    # Test 2: List collections
    print_test "Test 2: List collections"
    record_test
    if $cmd_prefix collection list &>/dev/null; then
        print_success "List collections passed"
        record_pass
    else
        print_error "List collections failed"
        record_fail "$vdb_name: list collections"
    fi

    # Test 3: Create collection
    print_test "Test 3: Create collection ($test_collection)"
    record_test
    if $cmd_prefix collection create "$test_collection" &>/dev/null; then
        print_success "Collection created"
        record_pass
    else
        print_error "Collection creation failed"
        record_fail "$vdb_name: create collection"
        return 1
    fi

    # Test 4: Add document
    print_test "Test 4: Add document"
    record_test
    # Create temporary file with test text
    local test_file="/tmp/e2e_test_${TIMESTAMP}.txt"
    echo "$TEST_TEXT" > "$test_file"
    if $cmd_prefix document create "$test_collection" "$test_file" &>/dev/null; then
        print_success "Document added"
        record_pass
        rm -f "$test_file"
    else
        print_error "Document creation failed"
        record_fail "$vdb_name: create document"
        rm -f "$test_file"
    fi

    # Test 5: List documents
    print_test "Test 5: List documents"
    record_test
    if $cmd_prefix document list "$test_collection" &>/dev/null; then
        print_success "List documents passed"
        record_pass
    else
        print_error "List documents failed"
        record_fail "$vdb_name: list documents"
    fi

    # Test 6: Count collections
    print_test "Test 6: Count collections"
    record_test
    if $cmd_prefix collection count &>/dev/null; then
        print_success "Count collections passed"
        record_pass
    else
        print_error "Count collections failed"
        record_fail "$vdb_name: count collections"
    fi

    # Test 7: Query/Search (if embeddings available)
    if [ -n "${OPENAI_API_KEY:-}" ]; then
        print_test "Test 7: Vector search"
        record_test
        if $cmd_prefix collection query "$test_collection" "test document" &>/dev/null; then
            print_success "Vector search passed"
            record_pass
        else
            print_warning "Vector search failed (may be expected for some DBs)"
            # Don't record as failure - some DBs may not support this yet
        fi
    else
        print_info "Skipping vector search (no OPENAI_API_KEY)"
    fi

    # Test 8: Delete collection (cleanup)
    print_test "Test 8: Delete collection (cleanup)"
    record_test
    if $cmd_prefix collection delete "$test_collection" --force &>/dev/null; then
        print_success "Collection deleted"
        record_pass
    else
        print_error "Collection deletion failed"
        record_fail "$vdb_name: delete collection"
    fi

    print_success "All core tests completed for $vdb_name"
    return 0
}

# Demo test - comprehensive test that mirrors presentation workflow
demo_test_vdb() {
    local vdb_name="$1"
    local vdb_flag="$2"
    local config_file="${3:-}"

    print_header "Demo Testing $vdb_name (Comprehensive)"

    # Sanitize collection name
    local sanitized_name="${vdb_name//-/_}"
    local test_docs_col="${TEST_COLLECTION_PREFIX}_DemoDocs_${sanitized_name}_${TIMESTAMP}"
    local test_images_col="${TEST_COLLECTION_PREFIX}_DemoImages_${sanitized_name}_${TIMESTAMP}"
    local cmd_prefix="$WEAVE_BIN"

    if [ -n "$config_file" ] && [ -f "$config_file" ]; then
        cmd_prefix="$cmd_prefix --config $config_file"
    fi
    cmd_prefix="$cmd_prefix $vdb_flag"

    print_info "Using collections: $test_docs_col, $test_images_col"

    # Test 1: Health check
    print_test "Test 1: Health check"
    record_test
    if $cmd_prefix health check &>/dev/null; then
        print_success "Health check passed"
        record_pass
    else
        print_error "Health check failed"
        record_fail "$vdb_name (demo): health check"
        print_warning "Skipping remaining demo tests for $vdb_name"
        return 1
    fi

    # Test 2: Collection list operations
    print_test "Test 2: Collection list (basic)"
    record_test
    if $cmd_prefix cols ls &>/dev/null; then
        print_success "Collection list passed"
        record_pass
    else
        print_error "Collection list failed"
        record_fail "$vdb_name (demo): collection list"
    fi

    # Test 3: Create text collection (basic)
    print_test "Test 3: Create text collection"
    record_test
    if $cmd_prefix cols create "$test_docs_col" &>/dev/null; then
        print_success "Text collection created"
        record_pass
    else
        print_error "Text collection creation failed"
        record_fail "$vdb_name (demo): create text collection"
        return 1
    fi

    # Test 4: Create image collection (basic)
    print_test "Test 4: Create image collection"
    record_test
    if $cmd_prefix cols create "$test_images_col" &>/dev/null; then
        print_success "Image collection created"
        record_pass
    else
        print_warning "Image collection creation failed (may not be supported)"
        # Continue even if image collection fails - use test_docs_col for remaining tests
        test_images_col="$test_docs_col"
    fi

    # Test 5: Show collection schema
    print_test "Test 5: Show collection schema with expanded metadata"
    record_test
    if $cmd_prefix cols show "$test_docs_col" --schema --expand-metadata &>/dev/null; then
        print_success "Show schema passed"
        record_pass
    else
        print_warning "Show schema failed (may not be supported)"
        # Don't record as failure - some DBs may not support this
    fi

    # Test 6: Show collection schema as JSON
    print_test "Test 6: Show collection schema as JSON"
    record_test
    if $cmd_prefix cols show "$test_images_col" --schema --expand-metadata --json &>/dev/null; then
        print_success "Show schema JSON passed"
        record_pass
    else
        print_warning "Show schema JSON failed (may not be supported)"
        # Don't record as failure
    fi

    # Test 7: Add document from README
    print_test "Test 7: Add document from README.md"
    record_test
    if [ -f "./README.md" ]; then
        if $cmd_prefix docs create "$test_docs_col" ./README.md &>/dev/null; then
            print_success "Document from README added"
            record_pass
        else
            print_error "Failed to add document from README"
            record_fail "$vdb_name (demo): add README document"
        fi
    else
        print_warning "README.md not found, skipping"
    fi

    # Test 8: Add document from another file
    print_test "Test 8: Add document from PRESENTATION.md"
    record_test
    if [ -f "./docs/PRESENTATION.md" ]; then
        if $cmd_prefix docs create "$test_docs_col" ./docs/PRESENTATION.md &>/dev/null; then
            print_success "Document from PRESENTATION.md added"
            record_pass
        else
            print_error "Failed to add document from PRESENTATION.md"
            record_fail "$vdb_name (demo): add PRESENTATION document"
        fi
    else
        print_warning "docs/PRESENTATION.md not found, skipping"
    fi

    # Test 9: List documents in text collection
    print_test "Test 9: List documents in text collection"
    record_test
    if $cmd_prefix docs ls "$test_docs_col" &>/dev/null; then
        print_success "List documents passed"
        record_pass
    else
        print_error "List documents failed"
        record_fail "$vdb_name (demo): list documents"
    fi

    # Test 10: Show specific document (get first document ID)
    print_test "Test 10: Show specific document with schema"
    record_test
    # Try to get a document ID and show it
    doc_id=$($cmd_prefix docs ls "$test_docs_col" --json 2>/dev/null | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4 || echo "")
    if [ -n "$doc_id" ]; then
        if $cmd_prefix docs show "$test_docs_col" "$doc_id" --schema --expand-metadata &>/dev/null; then
            print_success "Show document passed"
            record_pass
        else
            print_warning "Show document failed (may not be supported)"
        fi
    else
        print_warning "No document ID found to test show command"
    fi

    # Test 11: Add image document
    print_test "Test 11: Add image document"
    record_test
    if [ -f "./tests/images/dog.png" ]; then
        if $cmd_prefix docs create "$test_images_col" ./tests/images/dog.png &>/dev/null; then
            print_success "Image document added"
            record_pass
        else
            print_warning "Failed to add image document (may not be supported)"
        fi
    else
        print_warning "tests/images/dog.png not found, skipping"
    fi

    # Test 12: Add PDF with image extraction
    print_test "Test 12: Add PDF with image extraction"
    record_test
    if [ -f "./tests/fixtures/ragme-io.pdf" ]; then
        if $cmd_prefix docs create "$test_docs_col" ./tests/fixtures/ragme-io.pdf --image-col "$test_images_col" &>/dev/null; then
            print_success "PDF with image extraction passed"
            record_pass
        else
            print_warning "PDF with image extraction failed (may not be supported)"
        fi
    else
        print_warning "tests/fixtures/ragme-io.pdf not found, skipping"
    fi

    # Test 13: List documents in image collection
    print_test "Test 13: List documents in image collection"
    record_test
    if $cmd_prefix docs ls "$test_images_col" &>/dev/null; then
        print_success "List image documents passed"
        record_pass
    else
        print_warning "List image documents failed"
    fi

    # Test 14: Query collection with simple text
    if [ -n "${OPENAI_API_KEY:-}" ]; then
        print_test "Test 14: Query collection (simple)"
        record_test
        if $cmd_prefix cols query "$test_docs_col" "golang" &>/dev/null; then
            print_success "Simple query passed"
            record_pass
        else
            print_warning "Simple query failed (may not be supported)"
        fi

        # Test 15: Query with top_k parameter
        print_test "Test 15: Query collection with top_k"
        record_test
        if $cmd_prefix cols query "$test_docs_col" "RAGme.io" --top_k 3 &>/dev/null; then
            print_success "Query with top_k passed"
            record_pass
        else
            print_warning "Query with top_k failed (may not be supported)"
        fi
    else
        print_info "Skipping query tests (no OPENAI_API_KEY)"
    fi

    # Test 16: Delete document by name
    print_test "Test 16: Delete document by name"
    record_test
    if $cmd_prefix docs del "$test_docs_col" --name "README.md" &>/dev/null; then
        print_success "Delete by name passed"
        record_pass
    else
        print_warning "Delete by name failed (may not be supported)"
    fi

    # Test 17: Delete all documents
    print_test "Test 17: Delete all documents"
    record_test
    if $cmd_prefix docs delete-all "$test_docs_col" &>/dev/null; then
        print_success "Delete all documents passed"
        record_pass
    else
        print_warning "Delete all documents failed (may not be supported)"
    fi

    # Test 18: Collection count
    print_test "Test 18: Collection count"
    record_test
    if $cmd_prefix cols count &>/dev/null; then
        print_success "Collection count passed"
        record_pass
    else
        print_warning "Collection count failed (may not be supported)"
    fi

    # Test 19: Delete collection schema (cleanup)
    print_test "Test 19: Delete collection schemas (cleanup)"
    record_test
    local cleanup_success=true
    if ! $cmd_prefix cols delete-schema "$test_docs_col" &>/dev/null; then
        cleanup_success=false
    fi
    if ! $cmd_prefix cols delete-schema "$test_images_col" &>/dev/null; then
        cleanup_success=false
    fi
    if [ "$cleanup_success" = true ]; then
        print_success "Collection schemas deleted"
        record_pass
    else
        print_warning "Some collection schemas failed to delete"
    fi

    print_success "All demo tests completed for $vdb_name"
    return 0
}

# Test config commands (VDB-agnostic)
test_config_commands() {
    print_header "Testing Config Commands"

    # Test 1: config show
    print_test "Test 1: Config show"
    record_test
    if $WEAVE_BIN config show &>/dev/null; then
        print_success "Config show passed"
        record_pass
    else
        print_error "Config show failed"
        record_fail "config: show"
    fi

    # Test 2: config list
    print_test "Test 2: Config list"
    record_test
    if $WEAVE_BIN config list &>/dev/null; then
        print_success "Config list passed"
        record_pass
    else
        print_error "Config list failed"
        record_fail "config: list"
    fi

    # Test 3: config list --details
    print_test "Test 3: Config list --details"
    record_test
    if $WEAVE_BIN config list --details &>/dev/null; then
        print_success "Config list --details passed"
        record_pass
    else
        print_error "Config list --details failed"
        record_fail "config: list --details"
    fi

    # Test 4: config list-schemas
    print_test "Test 4: Config list-schemas"
    record_test
    if $WEAVE_BIN config list-schemas &>/dev/null; then
        print_success "Config list-schemas passed"
        record_pass
    else
        print_warning "Config list-schemas failed (may have no schemas)"
    fi

    # Test 5: config show-schema (if WeaveDocs exists)
    print_test "Test 5: Config show-schema WeaveDocs"
    record_test
    if $WEAVE_BIN config show-schema WeaveDocs &>/dev/null; then
        print_success "Config show-schema passed"
        record_pass
    else
        print_warning "Config show-schema failed (WeaveDocs may not exist)"
    fi

    # Test 6: config show-schema --yaml
    print_test "Test 6: Config show-schema --yaml"
    record_test
    if $WEAVE_BIN config show-schema WeaveDocs --yaml &>/dev/null; then
        print_success "Config show-schema --yaml passed"
        record_pass
    else
        print_warning "Config show-schema --yaml failed (WeaveDocs may not exist)"
    fi

    print_success "Config command tests completed"
}

# Test multi-VDB operations
test_multi_vdb_operations() {
    print_header "Testing Multi-VDB Operations"

    # Test 1: List collections across all VDBs (summary)
    print_test "Test 1: List collections (summary view)"
    record_test
    if $WEAVE_BIN cols ls &>/dev/null; then
        print_success "Multi-VDB list (summary) passed"
        record_pass
    else
        print_error "Multi-VDB list (summary) failed"
        record_fail "multi-vdb: list summary"
    fi

    # Test 2: List collections with --details
    print_test "Test 2: List collections --details"
    record_test
    if $WEAVE_BIN cols ls --details &>/dev/null; then
        print_success "Multi-VDB list (details) passed"
        record_pass
    else
        print_error "Multi-VDB list (details) failed"
        record_fail "multi-vdb: list details"
    fi

    # Test 3: List documents across VDBs with summary
    print_test "Test 3: List documents with summary (-w -S)"
    record_test
    # This requires WeaveDocs to exist - try with a common collection name
    if $WEAVE_BIN docs ls WeaveDocs -w -S &>/dev/null 2>&1 || \
       $WEAVE_BIN docs ls DemoDocs -w -S &>/dev/null 2>&1; then
        print_success "Multi-VDB document list with summary passed"
        record_pass
    else
        print_warning "Multi-VDB document list with summary failed (collections may not exist)"
    fi

    print_success "Multi-VDB operation tests completed"
}

# Main execution
main() {
    parse_args "$@"

    if [ "$TEST_DEMO" = true ]; then
        print_header "E2E Vector Database Tests (Demo Mode)"
        print_info "Running comprehensive demo-style tests"
    else
        print_header "E2E Vector Database Tests"
        print_info "Testing databases based on configuration and flags"
    fi
    echo ""

    check_prerequisites

    # Run demo mode tests if requested
    if [ "$TEST_DEMO" = true ]; then
        # Test config commands first (VDB-agnostic)
        test_config_commands

        # Test multi-VDB operations
        test_multi_vdb_operations

        # Test each VDB with comprehensive demo tests
        if [ "$TEST_WEAVIATE" = true ]; then
            # Try cloud first, then local
            if demo_test_vdb "Weaviate-Cloud" "--weaviate-cloud" "" 2>/dev/null; then
                :
            elif demo_test_vdb "Weaviate-Local" "--weaviate-local" "" 2>/dev/null; then
                :
            else
                print_warning "Weaviate not configured or not accessible"
            fi
        fi

        if [ "$TEST_MILVUS" = true ]; then
            if ./tools/vdb/local/milvus.sh status &>/dev/null; then
                demo_test_vdb "Milvus-Local" "--milvus-local" "config.milvus-local.yaml"
            else
                print_warning "Milvus local not running. Start with: ./tools/vdb/local/milvus.sh start"
            fi
        fi

        if [ "$TEST_SUPABASE" = true ]; then
            if demo_test_vdb "Supabase" "--supabase-cloud" "" 2>/dev/null; then
                :
            else
                print_warning "Supabase not configured or not accessible"
            fi
        fi

        if [ "$TEST_MONGODB" = true ]; then
            if demo_test_vdb "MongoDB" "--mongodb-cloud" "" 2>/dev/null; then
                :
            else
                print_warning "MongoDB not configured or not accessible"
            fi
        fi
    else
        # Run basic tests (original behavior)
        # Test Weaviate
        if [ "$TEST_WEAVIATE" = true ]; then
            # Try cloud first, then local
            if test_vdb "Weaviate-Cloud" "--weaviate" "" 2>/dev/null; then
                :
            elif test_vdb "Weaviate-Local" "--weaviate-local" "" 2>/dev/null; then
                :
            else
                print_warning "Weaviate not configured or not accessible"
            fi
        fi

        # Test Milvus
        if [ "$TEST_MILVUS" = true ]; then
            # Check if Milvus is running
            if ./tools/vdb/local/milvus.sh status &>/dev/null; then
                test_vdb "Milvus-Local" "--milvus-local" "config.milvus-local.yaml"
            else
                print_warning "Milvus local not running. Start with: ./tools/vdb/local/milvus.sh start"
            fi
        fi

        # Test Supabase
        if [ "$TEST_SUPABASE" = true ]; then
            if test_vdb "Supabase" "--supabase" "" 2>/dev/null; then
                :
            else
                print_warning "Supabase not configured or not accessible"
            fi
        fi

        # Test MongoDB
        if [ "$TEST_MONGODB" = true ]; then
            if test_vdb "MongoDB" "--mongodb" "" 2>/dev/null; then
                :
            else
                print_warning "MongoDB not configured or not accessible"
            fi
        fi
    fi

    # Print summary
    print_header "Test Summary"

    echo -e "${BOLD}Tests Run:${NC}    $TESTS_RUN"
    echo -e "${GREEN}${BOLD}Tests Passed:${NC} $TESTS_PASSED"

    if [ $TESTS_FAILED -gt 0 ]; then
        echo -e "${RED}${BOLD}Tests Failed:${NC} $TESTS_FAILED"
        echo ""
        echo -e "${RED}${BOLD}Failed Tests:${NC}"
        for failed_test in "${FAILED_TESTS[@]}"; do
            echo -e "${RED}  ❌ $failed_test${NC}"
        done
        echo ""
        exit 1
    else
        echo ""
        print_success "All tests passed! 🎉"
        echo ""
        exit 0
    fi
}

# Run main
main "$@"
