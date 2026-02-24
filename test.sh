#!/usr/bin/env bash

# Load environment variables from .env file
set -a
if [ -f .env ]; then
    # shellcheck disable=SC1091
    . .env
fi
set +a

# Set CGO environment variables for packages that need C dependencies (like gosseract)
export CGO_CPPFLAGS="-I/opt/homebrew/include -I/opt/homebrew/Cellar/leptonica/1.86.0/include -I/opt/homebrew/Cellar/tesseract/5.5.1/include"
export CGO_LDFLAGS="-L/opt/homebrew/lib"

# Weave CLI Test Suite

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
    echo -e "${BLUE}[TEST]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_help() {
    echo -e "${BLUE}Weave CLI Test Suite${NC}"
    echo ""
    echo "Usage: ./test.sh [COMMAND] [FLAGS]"
    echo ""
    echo "Commands:"
    echo "  unit        Run only unit tests"
    echo "  integration Run only integration tests"
    echo "  stack       Run weave stack integration tests"
    echo "  fast        Run fast tests (unit + mock integration)"
    echo "  all         Run all tests (unit + integration)"
    echo "  coverage    Run tests with coverage report"
    echo "  help        Show this help message"
    echo ""
    echo "Flags for 'integration' command:"
    echo "  --weaviate  Run only Weaviate integration tests"
    echo "  --milvus    Run only Milvus integration tests"
    echo "  --supabase  Run only Supabase integration tests"
    echo "  --mongodb   Run only MongoDB integration tests"
    echo "  --chroma    Run only Chroma integration tests"
    echo "  --qdrant    Run only Qdrant integration tests"
    echo "  --neo4j     Run only Neo4j integration tests"
    echo "  --pinecone  Run only Pinecone integration tests"
    echo "  --mcp       Run only MCP integration tests"
    echo "  --skip      Skip specified tests (e.g., --skip mcp,weaviate,neo4j,pinecone)"
    echo ""
    echo "Examples:"
    echo "  ./test.sh unit                    # Run only unit tests"
    echo "  ./test.sh integration             # Run all integration tests"
    echo "  ./test.sh stack                   # Run weave stack integration tests"
    echo "  ./test.sh integration --weaviate  # Run only Weaviate integration tests"
    echo "  ./test.sh integration --milvus    # Run only Milvus integration tests"
    echo "  ./test.sh integration --supabase  # Run only Supabase integration tests"
    echo "  ./test.sh integration --mongodb   # Run only MongoDB integration tests"
    echo "  ./test.sh integration --chroma    # Run only Chroma integration tests"
    echo "  ./test.sh integration --qdrant    # Run only Qdrant integration tests"
    echo "  ./test.sh integration --neo4j     # Run only Neo4j integration tests"
    echo "  ./test.sh integration --pinecone  # Run only Pinecone integration tests"
    echo "  ./test.sh integration --mcp       # Run only MCP integration tests"
    echo "  ./test.sh integration --skip mcp,weaviate,neo4j,pinecone  # Skip MCP, Weaviate, Neo4j, and Pinecone tests"
    echo "  ./test.sh fast                    # Run fast tests (unit + mock integration)"
    echo "  ./test.sh all                     # Run all tests"
    echo "  ./test.sh coverage                # Run tests with coverage report"
    echo "  ./test.sh                         # Run unit tests (default)"
    echo ""
    echo "Test Categories:"
    echo "  Unit Tests:"
    echo "    - Configuration management testing"
    echo "    - Mock client testing"
    echo "    - Utility function testing"
    echo ""
    echo "  Integration Tests:"
    echo "    - Weaviate client testing (requires WEAVIATE_URL, WEAVIATE_API_KEY, OPENAI_API_KEY)"
    echo "    - Milvus client testing (requires Milvus running locally at localhost:19530)"
    echo "    - Supabase client testing (requires SUPABASE_DATABASE_URL, SUPABASE_DATABASE_KEY)"
    echo "    - MongoDB client testing (requires MONGODB_URI, MONGODB_DATABASE)"
    echo "    - Chroma client testing (requires Chroma running locally at localhost:8000)"
    echo "    - Qdrant client testing (requires Qdrant running locally at localhost:6333)"
    echo "    - Neo4j client testing (requires Neo4j running locally at localhost:7687, NEO4J_PASSWORD)"
    echo "    - Pinecone client testing (requires PINECONE_API_KEY, OPENAI_API_KEY)"
    echo "    - MCP server testing (requires OPENAI_API_KEY, WEAVE_MCP_STDIO_PATH)"
    echo "    - CLI command testing"
    echo "    - End-to-end workflow testing"
}

# Initialize variables
RUN_UNIT_TESTS=false
RUN_INTEGRATION_TESTS=false
RUN_COVERAGE=false

# Global variables for integration test summary (must be global for trap handler)
declare -a vdb_names=()
declare -a vdb_status=()
declare -a vdb_passed=()
declare -a vdb_failed=()
total_suites=0
passed_suites=0
failed_suites=0
skipped_suites=0

# Check command line arguments
case "${1:-unit}" in
    "unit")
        RUN_UNIT_TESTS=true
        RUN_INTEGRATION_TESTS=false
        RUN_COVERAGE=false
        ;;
    "integration")
        RUN_UNIT_TESTS=false
        RUN_INTEGRATION_TESTS=true
        RUN_COVERAGE=false
        ;;
    "stack")
        # Stack tests handled separately below
        ;;
    "fast")
        RUN_UNIT_TESTS=true
        RUN_INTEGRATION_TESTS=true
        RUN_COVERAGE=false
        ;;
    "all")
        RUN_UNIT_TESTS=true
        RUN_INTEGRATION_TESTS=true
        RUN_COVERAGE=false
        ;;
    "coverage")
        RUN_UNIT_TESTS=false
        RUN_INTEGRATION_TESTS=false
        RUN_COVERAGE=true
        ;;
    "help"|"-h"|"--help")
        print_help
        exit 0
        ;;
    *)
        print_error "Unknown command: $1"
        echo ""
        print_help
        exit 1
        ;;
esac

# Function to run unit tests
run_unit_tests() {
    print_header "Running Unit Tests..."

    # Check if Go is installed
    if ! command -v go >/dev/null 2>&1; then
        print_error "Go is not installed. Please install Go 1.21 or later."
        exit 1
    fi

    # Run unit tests
    print_status "Running unit tests..."
    if go test -v -timeout=30s ./tests/... ./src/pkg/vectordb/mongodb/... -run="TestConfig|TestMock|TestWeaviateClient|TestFactory|TestDocumentConversion|TestGetTimeout|TestGetDefaultSchema|TestValidateSchema|TestMongoDBFactory|TestMongoDBClient"; then
        print_success "Unit tests passed!"
    else
        print_error "Unit tests failed!"
        exit 1
    fi
}

# Function to run integration tests
run_integration_tests() {
    print_header "Running Integration Tests..."

    # Check if Go is installed
    if ! command -v go >/dev/null 2>&1; then
        print_error "Go is not installed. Please install Go 1.21 or later."
        exit 1
    fi

    # Reset global counters and arrays for this test run
    total_suites=0
    passed_suites=0
    failed_suites=0
    skipped_suites=0
    vdb_names=()
    vdb_status=()
    vdb_passed=()
    vdb_failed=()

    # Trap to ensure summary is always printed (even on Ctrl+C or error)
    trap 'print_integration_summary; trap - EXIT' EXIT INT TERM

    # Parse integration test flags
    local run_weaviate=false
    local run_milvus=false
    local run_supabase=false
    local run_mongodb=false
    local run_chroma=false
    local run_qdrant=false
    local run_neo4j=false
    local run_pinecone=false
    local run_mcp=false
    local run_all=true

    # Skip flags
    local skip_weaviate=false
    local skip_milvus=false
    local skip_supabase=false
    local skip_mongodb=false
    local skip_chroma=false
    local skip_qdrant=false
    local skip_neo4j=false
    local skip_pinecone=false
    local skip_mcp=false
    local in_skip_mode=false

    # Check for specific flags
    for arg in "$@"; do
        # Check if we're in skip mode (collecting tests to skip)
        if [ "$in_skip_mode" = true ]; then
            # Handle comma-separated values
            IFS=',' read -ra SKIP_ITEMS <<< "$arg"
            for item in "${SKIP_ITEMS[@]}"; do
                case "$item" in
                    weaviate)
                        skip_weaviate=true
                        ;;
                    milvus)
                        skip_milvus=true
                        ;;
                    supabase)
                        skip_supabase=true
                        ;;
                    mongodb)
                        skip_mongodb=true
                        ;;
                    chroma)
                        skip_chroma=true
                        ;;
                    qdrant)
                        skip_qdrant=true
                        ;;
                    neo4j)
                        skip_neo4j=true
                        ;;
                    pinecone)
                        skip_pinecone=true
                        ;;
                    mcp)
                        skip_mcp=true
                        ;;
                esac
            done
            # Check if this arg starts with -- to exit skip mode
            if [[ "$arg" == --* ]]; then
                in_skip_mode=false
            else
                # Only process one skip argument then exit skip mode
                in_skip_mode=false
                continue
            fi
        fi

        case "$arg" in
            --weaviate)
                run_weaviate=true
                run_all=false
                ;;
            --milvus)
                run_milvus=true
                run_all=false
                ;;
            --supabase)
                run_supabase=true
                run_all=false
                ;;
            --mongodb)
                run_mongodb=true
                run_all=false
                ;;
            --chroma)
                run_chroma=true
                run_all=false
                ;;
            --qdrant)
                run_qdrant=true
                run_all=false
                ;;
            --neo4j)
                run_neo4j=true
                run_all=false
                ;;
            --pinecone)
                run_pinecone=true
                run_all=false
                ;;
            --mcp)
                run_mcp=true
                run_all=false
                ;;
            --skip)
                in_skip_mode=true
                ;;
        esac
    done

    # If run_all is true, enable all tests
    if [ "$run_all" = true ]; then
        run_weaviate=true
        run_milvus=true
        run_supabase=true
        run_mongodb=true
        run_chroma=true
        run_qdrant=true
        run_neo4j=true
        run_pinecone=true
        run_mcp=true

        # Run fast integration tests (mock only)
        print_status "Running fast integration tests (mock)..."
        total_suites=$((total_suites + 1))
        if go test -v -timeout=10s -count=1 ./tests/... -run="TestMock"; then
            print_success "Fast integration tests passed!"
            passed_suites=$((passed_suites + 1))
        else
            print_warning "Fast integration tests failed"
            failed_suites=$((failed_suites + 1))
        fi
    fi

    # Apply skip flags
    if [ "$skip_weaviate" = true ]; then
        run_weaviate=false
        print_status "Skipping Weaviate tests (--skip)"
    fi
    if [ "$skip_milvus" = true ]; then
        run_milvus=false
        print_status "Skipping Milvus tests (--skip)"
    fi
    if [ "$skip_supabase" = true ]; then
        run_supabase=false
        print_status "Skipping Supabase tests (--skip)"
    fi
    if [ "$skip_mongodb" = true ]; then
        run_mongodb=false
        print_status "Skipping MongoDB tests (--skip)"
    fi
    if [ "$skip_chroma" = true ]; then
        run_chroma=false
        print_status "Skipping Chroma tests (--skip)"
    fi
    if [ "$skip_qdrant" = true ]; then
        run_qdrant=false
        print_status "Skipping Qdrant tests (--skip)"
    fi
    if [ "$skip_neo4j" = true ]; then
        run_neo4j=false
        print_status "Skipping Neo4j tests (--skip)"
    fi
    if [ "$skip_pinecone" = true ]; then
        run_pinecone=false
        print_status "Skipping Pinecone tests (--skip)"
    fi
    if [ "$skip_mcp" = true ]; then
        run_mcp=false
        print_status "Skipping MCP tests (--skip)"
    fi

    # Run Weaviate integration tests if requested
    if [ "$run_weaviate" = true ]; then
        if [ -n "$WEAVIATE_URL" ] && [ -n "$WEAVIATE_API_KEY" ] && [ -n "$OPENAI_API_KEY" ]; then
            print_status "Running Weaviate integration tests..."
            total_suites=$((total_suites + 1))
            vdb_names+=("Weaviate")
            # Capture test output to count tests
            output=$(go test -v -timeout=2m -count=1 ./tests/... -run="TestWeaviate(Integration|FactoryIntegration|VectorDBRegistry|ConnectionSpeed|ErrorHandling)$" 2>&1)
            exit_code=$?
            pass_count=$(echo "$output" | grep -c "^--- PASS:" 2>/dev/null || true)
            fail_count=$(echo "$output" | grep -c "^--- FAIL:" 2>/dev/null || true)
            [ -z "$pass_count" ] && pass_count=0
            [ -z "$fail_count" ] && fail_count=0

            if [ $exit_code -eq 0 ]; then
                print_success "Weaviate integration tests passed!"
                passed_suites=$((passed_suites + 1))
                vdb_status+=("PASS")
            else
                print_warning "Weaviate integration tests failed"
                failed_suites=$((failed_suites + 1))
                vdb_status+=("FAIL")
            fi
            vdb_passed+=("$pass_count")
            vdb_failed+=("$fail_count")
            echo "$output"
        else
            print_warning "Skipping Weaviate integration tests - credentials not configured"
            print_status "Set WEAVIATE_URL, WEAVIATE_API_KEY, and OPENAI_API_KEY to run Weaviate tests"
            vdb_names+=("Weaviate")
            vdb_status+=("SKIP")
            vdb_passed+=("0")
            vdb_failed+=("0")
            skipped_suites=$((skipped_suites + 1))
        fi
    fi

    # Run Milvus integration tests if requested
    if [ "$run_milvus" = true ]; then
        # Check if Milvus is running locally (port 19530) or cloud credentials exist
        local milvus_available=false
        local milvus_host="${MILVUS_ADDRESS:-localhost}"
        local milvus_port="19530"

        # Parse host:port if address contains colon
        if [[ "$milvus_host" == *":"* ]]; then
            milvus_port="${milvus_host#*:}"
            milvus_host="${milvus_host%:*}"
        fi

        # Check if we can connect to Milvus port (timeout 2 seconds)
        if nc -z -w 2 "$milvus_host" "$milvus_port" 2>/dev/null; then
            milvus_available=true
        elif [ -f "./tools/vdb/local/milvus.sh" ] && ./tools/vdb/local/milvus.sh status &>/dev/null; then
            milvus_available=true
        fi

        if [ "$milvus_available" = true ] || [ -n "$MILVUS_CLOUD_ADDRESS" ]; then
            if [ "$milvus_available" = true ]; then
                print_status "Running Milvus integration tests (local)..."
            else
                print_status "Running Milvus integration tests (cloud)..."
            fi
            total_suites=$((total_suites + 1))
            vdb_names+=("Milvus")
            # Use shorter timeout - Milvus tests should complete quickly
            output=$(go test -v -timeout=3m -count=1 ./tests/... -run="TestMilvus" 2>&1)
            exit_code=$?
            pass_count=$(echo "$output" | grep -c "^--- PASS:" 2>/dev/null || true)
            fail_count=$(echo "$output" | grep -c "^--- FAIL:" 2>/dev/null || true)
            [ -z "$pass_count" ] && pass_count=0
            [ -z "$fail_count" ] && fail_count=0

            if [ $exit_code -eq 0 ]; then
                print_success "Milvus integration tests passed!"
                passed_suites=$((passed_suites + 1))
                vdb_status+=("PASS")
            else
                print_warning "Milvus integration tests failed"
                failed_suites=$((failed_suites + 1))
                vdb_status+=("FAIL")
            fi
            vdb_passed+=("$pass_count")
            vdb_failed+=("$fail_count")
            echo "$output"
        else
            print_warning "Skipping Milvus integration tests - Milvus not running"
            print_status "Start Milvus with: ./tools/vdb/local/milvus.sh start"
            vdb_names+=("Milvus")
            vdb_status+=("SKIP")
            vdb_passed+=("0")
            vdb_failed+=("0")
            skipped_suites=$((skipped_suites + 1))
        fi
    fi

    # Run Supabase integration tests if requested
    if [ "$run_supabase" = true ]; then
        if [ -n "$SUPABASE_DATABASE_URL" ] && [ -n "$SUPABASE_DATABASE_KEY" ]; then
            print_status "Running Supabase integration tests..."
            total_suites=$((total_suites + 1))
            vdb_names+=("Supabase")
            output=$(go test -v -timeout=2m -count=1 ./tests/... -run="TestSupabase" 2>&1)
            exit_code=$?
            pass_count=$(echo "$output" | grep -c "^--- PASS:" 2>/dev/null || true)
            fail_count=$(echo "$output" | grep -c "^--- FAIL:" 2>/dev/null || true)
            [ -z "$pass_count" ] && pass_count=0
            [ -z "$fail_count" ] && fail_count=0

            if [ $exit_code -eq 0 ]; then
                print_success "Supabase integration tests passed!"
                passed_suites=$((passed_suites + 1))
                vdb_status+=("PASS")
            else
                print_warning "Supabase integration tests failed"
                failed_suites=$((failed_suites + 1))
                vdb_status+=("FAIL")
            fi
            vdb_passed+=("$pass_count")
            vdb_failed+=("$fail_count")
            echo "$output"
        else
            print_warning "Skipping Supabase integration tests - credentials not configured"
            print_status "Set SUPABASE_DATABASE_URL and SUPABASE_DATABASE_KEY to run Supabase tests"
            vdb_names+=("Supabase")
            vdb_status+=("SKIP")
            vdb_passed+=("0")
            vdb_failed+=("0")
            skipped_suites=$((skipped_suites + 1))
        fi
    fi

    # Run MongoDB integration tests if requested
    if [ "$run_mongodb" = true ]; then
        if [ -n "$MONGODB_URI" ]; then
            print_status "Running MongoDB integration tests..."
            total_suites=$((total_suites + 1))
            vdb_names+=("MongoDB")
            output=$(go test -v -timeout=2m -count=1 ./tests/... -run="TestMongoDB" 2>&1)
            exit_code=$?
            pass_count=$(echo "$output" | grep -c "^--- PASS:" 2>/dev/null || true)
            fail_count=$(echo "$output" | grep -c "^--- FAIL:" 2>/dev/null || true)
            [ -z "$pass_count" ] && pass_count=0
            [ -z "$fail_count" ] && fail_count=0

            if [ $exit_code -eq 0 ]; then
                print_success "MongoDB integration tests passed!"
                passed_suites=$((passed_suites + 1))
                vdb_status+=("PASS")
            else
                print_warning "MongoDB integration tests failed"
                failed_suites=$((failed_suites + 1))
                vdb_status+=("FAIL")
            fi
            vdb_passed+=("$pass_count")
            vdb_failed+=("$fail_count")
            echo "$output"
        else
            print_warning "Skipping MongoDB integration tests - credentials not configured"
            print_status "Set MONGODB_URI and MONGODB_DATABASE to run MongoDB tests"
            vdb_names+=("MongoDB")
            vdb_status+=("SKIP")
            vdb_passed+=("0")
            vdb_failed+=("0")
            skipped_suites=$((skipped_suites + 1))
        fi
    fi

    # Run Chroma integration tests if requested
    if [ "$run_chroma" = true ]; then
        # Check if Chroma is running locally (port 8000)
        local chroma_available=false
        local chroma_url="${CHROMA_URL:-http://localhost:8000}"
        local chroma_host="localhost"
        local chroma_port="8000"

        # Parse host:port from URL
        if [[ "$chroma_url" =~ ://([^:/]+):([0-9]+) ]]; then
            chroma_host="${BASH_REMATCH[1]}"
            chroma_port="${BASH_REMATCH[2]}"
        elif [[ "$chroma_url" =~ ://([^:/]+) ]]; then
            chroma_host="${BASH_REMATCH[1]}"
        fi

        # Check if we can connect to Chroma port
        if nc -z -w 2 "$chroma_host" "$chroma_port" 2>/dev/null; then
            chroma_available=true
        fi

        if [ "$chroma_available" = true ] || [ -n "$CHROMA_API_KEY" ]; then
            if [ "$chroma_available" = true ]; then
                print_status "Running Chroma integration tests (local)..."
            else
                print_status "Running Chroma integration tests (cloud)..."
            fi
            total_suites=$((total_suites + 1))
            vdb_names+=("Chroma")
            # Run tests with live output to avoid appearance of hanging
            go test -v -timeout=2m -count=1 -tags=integration ./tests -run="TestChroma" 2>&1 | tee /tmp/chroma_test_output.txt
            exit_code=${PIPESTATUS[0]}
            output=$(cat /tmp/chroma_test_output.txt)
            pass_count=$(echo "$output" | grep -c "^--- PASS:" 2>/dev/null || true)
            fail_count=$(echo "$output" | grep -c "^--- FAIL:" 2>/dev/null || true)
            [ -z "$pass_count" ] && pass_count=0
            [ -z "$fail_count" ] && fail_count=0

            # Check if failures are due to quota limits
            quota_failures=$(echo "$output" | grep -c "Quota exceeded" 2>/dev/null || true)

            if [ $exit_code -eq 0 ]; then
                print_success "Chroma integration tests passed!"
                passed_suites=$((passed_suites + 1))
                vdb_status+=("PASS")
            else
                if [ "$quota_failures" -gt 0 ]; then
                    print_warning "Chroma integration tests failed"
                    print_status "Note: ${quota_failures} test(s) failed due to Chroma Cloud free tier quota limits (300 documents/request)"
                else
                    print_warning "Chroma integration tests failed"
                fi
                failed_suites=$((failed_suites + 1))
                vdb_status+=("FAIL")
            fi
            vdb_passed+=("$pass_count")
            vdb_failed+=("$fail_count")
            # Output already shown via tee, clean up temp file
            rm -f /tmp/chroma_test_output.txt
        else
            print_warning "Skipping Chroma integration tests - Chroma not running"
            print_status "Start Chroma with: podman run -d -p 8000:8000 chromadb/chroma:0.6.2"
            vdb_names+=("Chroma")
            vdb_status+=("SKIP")
            vdb_passed+=("0")
            vdb_failed+=("0")
            skipped_suites=$((skipped_suites + 1))
        fi
    fi

    # Run Qdrant integration tests if requested
    if [ "$run_qdrant" = true ]; then
        # Check if Qdrant is running locally (port 6333 HTTP, 6334 gRPC)
        local qdrant_available=false
        local qdrant_host="localhost"
        local qdrant_port="6333"

        # Check if we can connect to Qdrant HTTP port
        if nc -z -w 2 "$qdrant_host" "$qdrant_port" 2>/dev/null; then
            qdrant_available=true
        fi

        if [ "$qdrant_available" = true ] || [ -n "$QDRANT_API_KEY" ]; then
            if [ "$qdrant_available" = true ]; then
                print_status "Running Qdrant integration tests (local)..."
            else
                print_status "Running Qdrant integration tests (cloud)..."
            fi
            total_suites=$((total_suites + 1))
            vdb_names+=("Qdrant")
            # Run tests with live output
            go test -v -timeout=2m -count=1 -tags=integration ./tests -run="TestQdrant" 2>&1 | tee /tmp/qdrant_test_output.txt
            exit_code=${PIPESTATUS[0]}
            output=$(cat /tmp/qdrant_test_output.txt)
            pass_count=$(echo "$output" | grep -c "^--- PASS:" 2>/dev/null || true)
            fail_count=$(echo "$output" | grep -c "^--- FAIL:" 2>/dev/null || true)
            [ -z "$pass_count" ] && pass_count=0
            [ -z "$fail_count" ] && fail_count=0

            if [ $exit_code -eq 0 ]; then
                print_success "Qdrant integration tests passed!"
                passed_suites=$((passed_suites + 1))
                vdb_status+=("PASS")
            else
                print_warning "Qdrant integration tests failed"
                failed_suites=$((failed_suites + 1))
                vdb_status+=("FAIL")
            fi
            vdb_passed+=("$pass_count")
            vdb_failed+=("$fail_count")
            # Clean up temp file
            rm -f /tmp/qdrant_test_output.txt
        else
            print_warning "Skipping Qdrant integration tests - Qdrant not running"
            print_status "Start Qdrant with: podman run -d -p 6333:6333 -p 6334:6334 qdrant/qdrant:latest"
            vdb_names+=("Qdrant")
            vdb_status+=("SKIP")
            vdb_passed+=("0")
            vdb_failed+=("0")
            skipped_suites=$((skipped_suites + 1))
        fi
    fi

    # Run Neo4j integration tests if requested
    if [ "$run_neo4j" = true ]; then
        # Check if Neo4j is running locally (port 7687 Bolt)
        local neo4j_available=false
        local neo4j_host="localhost"
        local neo4j_port="7687"

        # Check if we can connect to Neo4j Bolt port
        if nc -z -w 2 "$neo4j_host" "$neo4j_port" 2>/dev/null; then
            neo4j_available=true
        fi

        # Check for required environment variables
        if [ "$neo4j_available" = true ] && [ -n "$NEO4J_PASSWORD" ]; then
            print_status "Running Neo4j integration tests (local)..."
            total_suites=$((total_suites + 1))
            vdb_names+=("Neo4j")
            # Run tests with live output
            go test -v -timeout=2m -count=1 -tags=integration ./tests -run="TestNeo4j" 2>&1 | tee /tmp/neo4j_test_output.txt
            exit_code=${PIPESTATUS[0]}
            output=$(cat /tmp/neo4j_test_output.txt)
            pass_count=$(echo "$output" | grep -c "^--- PASS:" 2>/dev/null || true)
            fail_count=$(echo "$output" | grep -c "^--- FAIL:" 2>/dev/null || true)
            [ -z "$pass_count" ] && pass_count=0
            [ -z "$fail_count" ] && fail_count=0

            if [ $exit_code -eq 0 ]; then
                print_success "Neo4j integration tests passed!"
                passed_suites=$((passed_suites + 1))
                vdb_status+=("PASS")
            else
                print_warning "Neo4j integration tests failed"
                failed_suites=$((failed_suites + 1))
                vdb_status+=("FAIL")
            fi
            vdb_passed+=("$pass_count")
            vdb_failed+=("$fail_count")
            # Clean up temp file
            rm -f /tmp/neo4j_test_output.txt
        else
            print_warning "Skipping Neo4j integration tests - Neo4j not running or credentials not configured"
            print_status "Start Neo4j with: ./tools/vdb/local/neo4j.sh start"
            print_status "Set NEO4J_PASSWORD environment variable"
            vdb_names+=("Neo4j")
            vdb_status+=("SKIP")
            vdb_passed+=("0")
            vdb_failed+=("0")
            skipped_suites=$((skipped_suites + 1))
        fi
    fi

    # Run Pinecone integration tests if requested
    if [ "$run_pinecone" = true ]; then
        if [ -n "$PINECONE_API_KEY" ] && [ -n "$OPENAI_API_KEY" ]; then
            print_status "Running Pinecone integration tests (cloud)..."
            total_suites=$((total_suites + 1))
            vdb_names+=("Pinecone")
            # Run tests with live output
            go test -v -timeout=5m -count=1 -tags=integration ./tests -run="TestPinecone" 2>&1 | tee /tmp/pinecone_test_output.txt
            exit_code=${PIPESTATUS[0]}
            output=$(cat /tmp/pinecone_test_output.txt)
            pass_count=$(echo "$output" | grep -c "^--- PASS:" 2>/dev/null || true)
            fail_count=$(echo "$output" | grep -c "^--- FAIL:" 2>/dev/null || true)
            [ -z "$pass_count" ] && pass_count=0
            [ -z "$fail_count" ] && fail_count=0

            if [ $exit_code -eq 0 ]; then
                print_success "Pinecone integration tests passed!"
                passed_suites=$((passed_suites + 1))
                vdb_status+=("PASS")
            else
                print_warning "Pinecone integration tests failed"
                failed_suites=$((failed_suites + 1))
                vdb_status+=("FAIL")
            fi
            vdb_passed+=("$pass_count")
            vdb_failed+=("$fail_count")
            # Clean up temp file
            rm -f /tmp/pinecone_test_output.txt
        else
            print_warning "Skipping Pinecone integration tests - credentials not configured"
            print_status "Set PINECONE_API_KEY and OPENAI_API_KEY to run Pinecone tests"
            vdb_names+=("Pinecone")
            vdb_status+=("SKIP")
            vdb_passed+=("0")
            vdb_failed+=("0")
            skipped_suites=$((skipped_suites + 1))
        fi
    fi

    # Run MCP integration tests if requested
    if [ "$run_mcp" = true ]; then
        if [ -n "$OPENAI_API_KEY" ] && [ -n "$WEAVE_MCP_STDIO_PATH" ]; then
            print_status "Running MCP integration tests..."
            total_suites=$((total_suites + 1))
            vdb_names+=("MCP")
            output=$(go test -v -tags=integration -timeout=5m -count=1 ./tests -run="TestMCP" 2>&1)
            exit_code=$?
            pass_count=$(echo "$output" | grep -c "^--- PASS:" 2>/dev/null || true)
            fail_count=$(echo "$output" | grep -c "^--- FAIL:" 2>/dev/null || true)
            [ -z "$pass_count" ] && pass_count=0
            [ -z "$fail_count" ] && fail_count=0

            if [ $exit_code -eq 0 ]; then
                print_success "MCP integration tests passed!"
                passed_suites=$((passed_suites + 1))
                vdb_status+=("PASS")
            else
                print_warning "MCP integration tests failed"
                failed_suites=$((failed_suites + 1))
                vdb_status+=("FAIL")
            fi
            vdb_passed+=("$pass_count")
            vdb_failed+=("$fail_count")
            echo "$output"
        else
            print_warning "Skipping MCP integration tests - configuration not provided"
            print_status "Set OPENAI_API_KEY and WEAVE_MCP_STDIO_PATH to run MCP tests"
            vdb_names+=("MCP")
            vdb_status+=("SKIP")
            vdb_passed+=("0")
            vdb_failed+=("0")
            skipped_suites=$((skipped_suites + 1))
        fi
    fi

    # Run CLI integration tests (--top_k_images, multi-collection, etc.)
    # These tests auto-detect available VDB and work with existing collections
    if [ "$run_all" = true ]; then
        print_status "Running CLI integration tests..."
        total_suites=$((total_suites + 1))
        vdb_names+=("CLI Tests")
        # Run with live output
        go test -v -tags=integration -timeout=10m -count=1 ./tests/integration/... 2>&1 | tee /tmp/cli_test_output.txt
        exit_code=${PIPESTATUS[0]}
        output=$(cat /tmp/cli_test_output.txt)
        pass_count=$(echo "$output" | grep -c "^--- PASS:" 2>/dev/null || true)
        fail_count=$(echo "$output" | grep -c "^--- FAIL:" 2>/dev/null || true)
        [ -z "$pass_count" ] && pass_count=0
        [ -z "$fail_count" ] && fail_count=0

        if [ $exit_code -eq 0 ]; then
            print_success "CLI integration tests passed!"
            passed_suites=$((passed_suites + 1))
            vdb_status+=("PASS")
        else
            # Check if all tests were skipped
            skip_count=$(echo "$output" | grep -c "Skipping.*No vector database" 2>/dev/null || true)
            if [ "$skip_count" -gt 0 ] && [ "$fail_count" -eq 0 ]; then
                print_warning "CLI integration tests skipped - no VDB credentials configured"
                skipped_suites=$((skipped_suites + 1))
                vdb_status+=("SKIP")
            else
                print_warning "CLI integration tests failed"
                failed_suites=$((failed_suites + 1))
                vdb_status+=("FAIL")
            fi
        fi
        vdb_passed+=("$pass_count")
        vdb_failed+=("$fail_count")
        # Clean up temp file
        rm -f /tmp/cli_test_output.txt
    fi

    # Call summary function
    print_integration_summary

    # Clear trap and return exit code based on results
    trap - EXIT
    if [ $failed_suites -gt 0 ]; then
        return 1
    fi
    return 0
}

# Function to print integration test summary
# This is a separate function so it can be called by trap handlers
print_integration_summary() {
    # Only print if we have vdb_names (i.e., tests have started running)
    if [ ${#vdb_names[@]} -eq 0 ]; then
        return
    fi

    # Print detailed summary
    echo ""
    echo ""
    print_header "═══════════════════════════════════════════════════════════════"
    print_header "                    INTEGRATION TEST SUMMARY                    "
    print_header "═══════════════════════════════════════════════════════════════"
    echo ""

    # Print table header
    printf "  %-15s %-12s %-20s\n" "VDB" "STATUS" "TESTS"
    echo "  ─────────────────────────────────────────────"

    # Print results for each VDB
    for i in "${!vdb_names[@]}"; do
        status="${vdb_status[$i]}"
        passed="${vdb_passed[$i]}"
        failed="${vdb_failed[$i]}"

        case "$status" in
            "PASS")
                status_color="${GREEN}✓ PASS${NC}"
                test_info="${GREEN}${passed} ✅${NC}"
                ;;
            "FAIL")
                status_color="${RED}✗ FAIL${NC}"
                if [ "$failed" -gt 0 ]; then
                    test_info="${GREEN}${passed} ✅${NC} ${RED}${failed} ❌${NC}"
                else
                    test_info="${GREEN}${passed} ✅${NC}"
                fi
                ;;
            "SKIP")
                status_color="${YELLOW}○ SKIP${NC}"
                test_info="${YELLOW}-${NC}"
                ;;
        esac
        printf "  %-15s " "${vdb_names[$i]}"
        echo -e "$status_color    $test_info"
    done

    echo "  ─────────────────────────────────────────"
    echo ""

    # Print totals
    echo "  Summary:"
    echo -e "    Total Suites:   $total_suites"
    echo -e "    ${GREEN}Passed:         $passed_suites${NC}"
    echo -e "    ${RED}Failed:         $failed_suites${NC}"
    echo -e "    ${YELLOW}Skipped:        $skipped_suites${NC}"
    echo ""

    if [ $failed_suites -eq 0 ] && [ $passed_suites -gt 0 ]; then
        echo -e "  ${GREEN}════════════════════════════════════════════════════════════${NC}"
        echo -e "  ${GREEN}  ✓ ALL INTEGRATION TESTS PASSED!                          ${NC}"
        echo -e "  ${GREEN}════════════════════════════════════════════════════════════${NC}"
    elif [ $failed_suites -gt 0 ]; then
        echo -e "  ${RED}════════════════════════════════════════════════════════════${NC}"
        echo -e "  ${RED}  ✗ SOME TESTS FAILED                                       ${NC}"
        echo -e "  ${RED}════════════════════════════════════════════════════════════${NC}"
        print_warning "Failures may be due to network issues, cloud configuration, quota limits, or test flakiness"
    else
        echo -e "  ${YELLOW}════════════════════════════════════════════════════════════${NC}"
        echo -e "  ${YELLOW}  ○ ALL TESTS SKIPPED - Check configuration                ${NC}"
        echo -e "  ${YELLOW}════════════════════════════════════════════════════════════${NC}"
    fi
    echo ""
}

# Function to run fast tests
run_fast_tests() {
    print_header "Running Fast Tests..."

    # Check if Go is installed
    if ! command -v go >/dev/null 2>&1; then
        print_error "Go is not installed. Please install Go 1.21 or later."
        exit 1
    fi

    # Run unit tests
    print_status "Running unit tests..."
    if go test -v -timeout=30s ./tests/... ./src/pkg/vectordb/mongodb/... -run="TestConfig|TestMock|TestWeaviateClient|TestFactory|TestDocumentConversion|TestGetTimeout|TestGetDefaultSchema|TestValidateSchema|TestMongoDBFactory|TestMongoDBClient"; then
        print_success "Unit tests passed!"
    else
        print_error "Unit tests failed!"
        exit 1
    fi

    # Run fast integration tests (mock only)
    print_status "Running fast integration tests (mock)..."
    if go test -v -timeout=10s ./tests/... -run="TestMock"; then
        print_success "Fast integration tests passed!"
    else
        print_warning "Fast integration tests failed"
    fi

    print_success "Fast tests completed!"
}

# Function to run coverage tests
run_coverage_tests() {
    print_header "Running Coverage Analysis..."
    
    # Check if Go is installed
    if ! command -v go >/dev/null 2>&1; then
        print_error "Go is not installed. Please install Go 1.21 or later."
        exit 1
    fi
    
    # Create coverage directory
    mkdir -p coverage
    
    # Run tests with coverage (only unit tests and mock integration tests)
    print_status "Running tests with coverage..."
    if go test -coverprofile=coverage/coverage.out -covermode=atomic ./tests/... -run="TestConfig|TestMock|TestCLI|TestFastMock|TestFastConfig"; then
        print_status "Generating coverage report..."
        
        # Generate HTML coverage report
        go tool cover -html=coverage/coverage.out -o coverage/coverage.html
        
        # Generate text coverage report
        go tool cover -func=coverage/coverage.out > coverage/coverage.txt
        
        print_success "Coverage analysis completed!"
        print_status "Coverage files available in:"
        echo "  - coverage/coverage.html (HTML report)"
        echo "  - coverage/coverage.txt (Text report)"
        echo "  - coverage/coverage.out (Raw coverage data)"
    else
        print_error "Coverage analysis failed!"
        exit 1
    fi
}

# Function to create basic integration tests (currently unused)
# create_integration_tests() {
#     print_status "Creating basic integration test structure..."
#     
#     # Create tests directory structure
#     mkdir -p tests/{config,weaviate,mock,cmd}
#     
#     # Create basic config test
#     cat > tests/config/config_test.go << 'EOF'
# package config_test
# 
# import (
# 	"testing"
# 	"github.com/maximilien/weave-cli/src/internal/config"
# )
# 
# func TestLoadConfig(t *testing.T) {
# 	// Test loading config with default files
# 	cfg, err := config.LoadConfig("", "")
# 	if err != nil {
# 		t.Logf("Config loading failed (expected if no config files): %v", err)
# 		return
# 	}
# 	
# 	if cfg == nil {
# 		t.Error("Config should not be nil")
# 	}
# }
# 
# func TestInterpolateEnvVars(t *testing.T) {
# 	// Test environment variable interpolation
# 	testCases := []struct {
# 		input    string
# 		expected string
# 	}{
# 		{"${TEST_VAR:-default}", "default"},
# 		{"simple string", "simple string"},
# 		{"${TEST_VAR}", ""},
# 	}
# 	
# 	for _, tc := range testCases {
# 		result := config.InterpolateString(tc.input)
# 		if result != tc.expected {
# 			t.Errorf("Expected %s, got %s", tc.expected, result)
# 		}
# 	}
# }
# EOF
# 
#     # Create basic mock test
#     cat > tests/mock/client_test.go << 'EOF'
# package mock_test
# 
# import (
# 	"context"
# 	"testing"
# 	"github.com/maximilien/weave-cli/src/internal/config"
# 	"github.com/maximilien/weave-cli/src/internal/mock"
# )
# 
# func TestMockClient(t *testing.T) {
# 	cfg := &config.MockConfig{
# 		Enabled:            true,
# 		SimulateEmbeddings: true,
# 		EmbeddingDimension: 384,
# 		Collections: []config.MockCollection{
# 			{Name: "test", Type: "text", Description: "Test collection"},
# 		},
# 	}
# 	
# 	client := mock.NewClient(cfg)
# 	
# 	// Test health check
# 	ctx := context.Background()
# 	if err := client.Health(ctx); err != nil {
# 		t.Errorf("Health check failed: %v", err)
# 	}
# 	
# 	// Test listing collections
# 	collections, err := client.ListCollections(ctx)
# 	if err != nil {
# 		t.Errorf("Failed to list collections: %v", err)
# 	}
# 	
# 	if len(collections) != 1 {
# 		t.Errorf("Expected 1 collection, got %d", len(collections))
# 	}
# }
# EOF
# 
#     print_success "Integration test structure created!"
# }


# Run weave stack integration tests
run_stack_tests() {
    print_header "Weave Stack Integration Tests"

    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    TEST_DIR="/tmp/weave-stack-test-$$"
    TESTS_PASSED=0
    TESTS_FAILED=0

    mkdir -p "$TEST_DIR"
    cd "$TEST_DIR"

    # Test 1: init
    print_status "Test 1: weave stack init"
    if "$SCRIPT_DIR/bin/weave" stack init --runtime kind --quiet-config > /tmp/init.log 2>&1; then
        if [ -f "weave-stack.yaml" ]; then
            print_success "✓ Test 1 passed: weave-stack.yaml created"
            ((TESTS_PASSED++))
        else
            print_error "✗ Test 1 failed: weave-stack.yaml not found"
            ((TESTS_FAILED++))
        fi
    else
        print_error "✗ Test 1 failed: init command failed"
        cat /tmp/init.log
        ((TESTS_FAILED++))
    fi

    # Test 2: validate
    print_status "Test 2: weave stack validate"
    if "$SCRIPT_DIR/bin/weave" stack validate --quiet-config > /dev/null 2>&1; then
        print_success "✓ Test 2 passed: configuration valid"
        ((TESTS_PASSED++))
    else
        print_error "✗ Test 2 failed: validation failed"
        ((TESTS_FAILED++))
    fi

    # Test 3: all templates
    for template in production multimodal oss; do
        print_status "Test: template $template"
        rm -rf *
        if "$SCRIPT_DIR/bin/weave" stack init --template "$template" --runtime kind --quiet-config > /dev/null 2>&1; then
            if "$SCRIPT_DIR/bin/weave" stack validate --quiet-config > /dev/null 2>&1; then
                print_success "✓ Template $template valid"
                ((TESTS_PASSED++))
            else
                print_error "✗ Template $template validation failed"
                ((TESTS_FAILED++))
            fi
        else
            print_error "✗ Template $template init failed"
            ((TESTS_FAILED++))
        fi
    done

    # Cleanup
    cd "$SCRIPT_DIR"
    rm -rf "$TEST_DIR"

    echo ""
    echo "Stack Test Summary:"
    echo "  Passed: $TESTS_PASSED"
    echo "  Failed: $TESTS_FAILED"

    if [ $TESTS_FAILED -gt 0 ]; then
        return 1
    fi
    return 0
}

# Run tests based on command
# Track overall exit code
overall_exit_code=0

if [ "${1:-unit}" = "stack" ]; then
    run_stack_tests
    overall_exit_code=$?
elif [ "$RUN_UNIT_TESTS" = true ] && [ "$RUN_INTEGRATION_TESTS" = true ]; then
    # Check if this is a fast test run
    if [ "${1:-unit}" = "fast" ]; then
        run_fast_tests
        overall_exit_code=$?
    else
        run_unit_tests
        overall_exit_code=$?
        # Run integration tests even if unit tests fail
        run_integration_tests "${@:2}" || overall_exit_code=1
    fi
elif [ "$RUN_UNIT_TESTS" = true ]; then
    run_unit_tests
    overall_exit_code=$?
elif [ "$RUN_INTEGRATION_TESTS" = true ]; then
    run_integration_tests "${@:2}"
    overall_exit_code=$?
fi

# Run coverage tests if requested
if [ "$RUN_COVERAGE" = true ]; then
    run_coverage_tests
    overall_exit_code=$?
fi

if [ $overall_exit_code -eq 0 ]; then
    print_status "All requested tests completed successfully!"
else
    print_warning "Tests completed with failures. See summary above for details."
fi

exit $overall_exit_code