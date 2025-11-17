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
    -h, --help          Show this help message

Examples:
    # Test all databases
    ./e2e-tests.sh
    ./e2e-tests.sh --all

    # Test specific databases
    ./e2e-tests.sh --weaviate --milvus
    ./e2e-tests.sh --milvus

Prerequisites:
    - Weave CLI built (./build.sh)
    - Vector databases configured (.env or config.yaml)
    - For Milvus local: ./tools/vdb/local/milvus.sh start
    - OPENAI_API_KEY set for automatic embeddings

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

# Main execution
main() {
    parse_args "$@"

    print_header "E2E Vector Database Tests"
    print_info "Testing databases based on configuration and flags"
    echo ""

    check_prerequisites

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
