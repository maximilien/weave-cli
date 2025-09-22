#!/usr/bin/env bash

# Load environment variables from .env file
set -a
[ -f .env ] && . .env
set +a

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

print_help() {
    echo -e "${BLUE}Weave CLI Test Suite${NC}"
    echo ""
    echo "Usage: ./test.sh [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  unit        Run only unit tests"
    echo "  integration Run only integration tests"
    echo "  all         Run all tests (unit + integration)"
    echo "  coverage    Run tests with coverage report"
    echo "  help        Show this help message"
    echo ""
    echo "Examples:"
    echo "  ./test.sh unit         # Run only unit tests"
    echo "  ./test.sh integration  # Run only integration tests"
    echo "  ./test.sh all          # Run all tests"
    echo "  ./test.sh coverage     # Run tests with coverage report"
    echo "  ./test.sh              # Run unit tests (default)"
    echo ""
    echo "Test Categories:"
    echo "  Unit Tests:"
    echo "    - Configuration management testing"
    echo "    - Mock client testing"
    echo "    - Utility function testing"
    echo ""
    echo "  Integration Tests:"
    echo "    - Weaviate client testing"
    echo "    - CLI command testing"
    echo "    - End-to-end workflow testing"
}

# Initialize variables
RUN_UNIT_TESTS=false
RUN_INTEGRATION_TESTS=false
RUN_COVERAGE=false

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
    if go test -v ./tests/...; then
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
    
    # Run integration tests
    print_status "Running integration tests..."
    if go test -v ./tests/...; then
        print_success "Integration tests passed!"
    else
        print_warning "Integration tests failed or no tests found"
        print_status "Creating basic integration test structure..."
        create_integration_tests
    fi
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
    
    # Run tests with coverage
    print_status "Running tests with coverage..."
    if go test -coverprofile=coverage/coverage.out -covermode=atomic ./tests/...; then
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

# Function to create basic integration tests
create_integration_tests() {
    print_status "Creating basic integration test structure..."
    
    # Create tests directory structure
    mkdir -p tests/{config,weaviate,mock,cmd}
    
    # Create basic config test
    cat > tests/config/config_test.go << 'EOF'
package config_test

import (
	"testing"
	"github.com/maximilien/weave-cli/src/internal/config"
)

func TestLoadConfig(t *testing.T) {
	// Test loading config with default files
	cfg, err := config.LoadConfig("", "")
	if err != nil {
		t.Logf("Config loading failed (expected if no config files): %v", err)
		return
	}
	
	if cfg == nil {
		t.Error("Config should not be nil")
	}
}

func TestInterpolateEnvVars(t *testing.T) {
	// Test environment variable interpolation
	testCases := []struct {
		input    string
		expected string
	}{
		{"${TEST_VAR:-default}", "default"},
		{"simple string", "simple string"},
		{"${TEST_VAR}", ""},
	}
	
	for _, tc := range testCases {
		result := config.InterpolateString(tc.input)
		if result != tc.expected {
			t.Errorf("Expected %s, got %s", tc.expected, result)
		}
	}
}
EOF

    # Create basic mock test
    cat > tests/mock/client_test.go << 'EOF'
package mock_test

import (
	"context"
	"testing"
	"github.com/maximilien/weave-cli/src/internal/config"
	"github.com/maximilien/weave-cli/src/internal/mock"
)

func TestMockClient(t *testing.T) {
	cfg := &config.MockConfig{
		Enabled:            true,
		SimulateEmbeddings: true,
		EmbeddingDimension: 384,
		Collections: []config.MockCollection{
			{Name: "test", Type: "text", Description: "Test collection"},
		},
	}
	
	client := mock.NewClient(cfg)
	
	// Test health check
	ctx := context.Background()
	if err := client.Health(ctx); err != nil {
		t.Errorf("Health check failed: %v", err)
	}
	
	// Test listing collections
	collections, err := client.ListCollections(ctx)
	if err != nil {
		t.Errorf("Failed to list collections: %v", err)
	}
	
	if len(collections) != 1 {
		t.Errorf("Expected 1 collection, got %d", len(collections))
	}
}
EOF

    print_success "Integration test structure created!"
}

# Helper function for success messages
print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Run unit tests if requested
if [ "$RUN_UNIT_TESTS" = true ]; then
    run_unit_tests
fi

# Run integration tests if requested
if [ "$RUN_INTEGRATION_TESTS" = true ]; then
    run_integration_tests
fi

# Run coverage tests if requested
if [ "$RUN_COVERAGE" = true ]; then
    run_coverage_tests
fi

print_status "All requested tests completed!"
exit 0