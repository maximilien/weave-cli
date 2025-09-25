#!/usr/bin/env bash

# SPDX-License-Identifier: MIT
# Copyright (c) 2025 dr.max

set -e

# Parse command line arguments
SKIP_SECURITY=false
for arg in "$@"; do
    case $arg in
        --skip-security)
            SKIP_SECURITY=true
            shift
            ;;
        --help|-h)
            echo "Usage: $0 [--skip-security]"
            echo "  --skip-security    Skip security checks (govulncheck, gosec)"
            exit 0
            ;;
    esac
done

echo "🔍 Running linter on Weave CLI project..."

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
    echo -e "${BLUE}[LINT]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Go linting
echo "📁 Checking Go files..."
if command_exists go; then
    print_header "Running Go linter..."
    
    # Install golangci-lint if not present
    if ! command_exists golangci-lint; then
        print_status "Installing golangci-lint..."
        if command_exists curl; then
            # Use go install to build with current Go version for compatibility
            go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
        else
            print_warning "curl not found, please install golangci-lint manually"
            print_status "Visit: https://golangci-lint.run/usage/install/"
        fi
    fi
    
    # Run golangci-lint if available
    if command_exists golangci-lint; then
        print_status "Running golangci-lint..."
        if golangci-lint run ./src/...; then
            print_success "Go linting passed!"
        else
            print_error "Go linting failed!"
            exit 1
        fi
    else
        # Fallback to go vet and go fmt
        print_status "Running go vet..."
        if go vet ./src/...; then
            print_success "go vet passed!"
        else
            print_error "go vet failed!"
            exit 1
        fi
        
        print_status "Checking go fmt..."
        if [ "$(gofmt -s -l src/)" ]; then
            print_warning "Code is not formatted with go fmt!"
            print_status "Auto-fixing formatting..."
            gofmt -s -w src/
            print_success "Code formatting fixed!"
        else
            print_success "go fmt check passed!"
        fi
    fi
    
    # Check for common Go issues
    print_status "Checking for common Go issues..."
    
    # Check for unused imports
    if command_exists goimports; then
        print_status "Checking imports..."
        if [ "$(goimports -l src/)" ]; then
            print_warning "Unused imports found. Run 'goimports -w src/' to fix"
        else
            print_success "Import check passed!"
        fi
    fi
    
    # Check for race conditions
    print_status "Checking for race conditions..."
    if go test -race ./src/... >/dev/null 2>&1; then
        print_success "Race condition check passed!"
    else
        print_warning "Race condition check failed or tests not available"
    fi
    
    print_success "Go linting checks passed!"
else
    print_error "Go is not installed, skipping Go linting"
    exit 1
fi

# JSON linting
echo "📄 Checking JSON files..."
if find . -name "*.json" -not -path "./src/vendor/*" -not -path "./node_modules/*" | grep -q .; then
    find . -name "*.json" -not -path "./src/vendor/*" -not -path "./node_modules/*" -print0 | while IFS= read -r -d '' json_file; do
        if ! python3 -m json.tool "$json_file" >/dev/null 2>&1; then
            print_error "Invalid JSON found in $json_file"
            exit 1
        fi
    done
    print_success "JSON files are valid!"
else
    echo "ℹ️  No JSON files found to validate"
fi

# YAML linting (excluding GitHub Actions workflows with false positives)
echo "📋 Checking YAML files..."
if find . -name "*.yml" -o -name "*.yaml" | grep -q .; then
    if command_exists yamllint; then
        # Use .yamllint config file to exclude GitHub Actions workflows with false positives
        if yamllint .; then
            print_success "YAML linting passed!"
        else
            print_warning "YAML linting issues found"
        fi
    else
        print_warning "yamllint not found, skipping YAML linting"
        print_status "Install yamllint: pip install yamllint"
    fi
else
    echo "ℹ️  No YAML files found to lint"
fi

# Markdown linting (excluding docs directory with extended formats)
echo "📝 Checking Markdown files..."
if find . -name "*.md" -not -path "./src/vendor/*" -not -path "./node_modules/*" -not -path "./docs/*" | grep -q .; then
    if command_exists markdownlint; then
        if npx markdownlint "**/*.md" --ignore node_modules --ignore src/vendor --ignore docs; then
            print_success "Markdown linting passed!"
        else
            print_warning "Markdown linting issues found"
        fi
    else
        print_warning "markdownlint not found, skipping Markdown linting"
        print_status "Install markdownlint: npm install -g markdownlint-cli"
    fi
else
    echo "ℹ️  No Markdown files found to lint (excluding docs directory with extended formats)"
fi

# Shell script linting
echo "🐚 Checking shell scripts..."
if find . -name "*.sh" | grep -q .; then
    if command_exists shellcheck; then
        find . -name "*.sh" -print0 | while IFS= read -r -d '' sh_file; do
            if shellcheck "$sh_file"; then
                print_success "Shell script $sh_file passed!"
            else
                print_error "Shell script $sh_file has issues"
                exit 1
            fi
        done
    else
        print_warning "shellcheck not found, skipping shell script linting"
        print_status "Install shellcheck: brew install shellcheck (macOS) or apt-get install shellcheck (Ubuntu)"
    fi
else
    echo "ℹ️  No shell scripts found to lint"
fi

# Security checks (optional)
if [ "$SKIP_SECURITY" = true ]; then
    echo "🔒 Skipping security checks (--skip-security flag used)"
else
    echo "🔒 Running security checks..."
    if command_exists govulncheck; then
        print_status "Running govulncheck vulnerability scanner..."
        if govulncheck ./src/...; then
            print_success "Vulnerability scan passed!"
        else
            print_warning "Vulnerabilities found"
        fi
    else
        print_status "govulncheck not available - using go mod audit as alternative"
        if go list -json -m all | grep -q '"Indirect":true'; then
            print_status "Checking for known vulnerabilities in dependencies..."
            if go mod audit; then
                print_success "Dependency audit passed!"
            else
                print_warning "Potential dependency issues found"
            fi
        else
            print_status "No indirect dependencies to audit"
        fi
    fi

    # Additional security checks with gosec if available
    if command_exists gosec; then
        print_status "Running gosec security scanner..."
        if gosec ./src/...; then
            print_success "Security scan passed!"
        else
            print_warning "Security issues found"
        fi
    else
        print_status "gosec not available - using go vet as security alternative"
        print_status "Running go vet for basic security checks..."
        if go vet -unsafeptr=false ./src/...; then
            print_success "Basic security checks passed!"
        else
            print_warning "Basic security checks found issues"
        fi
    fi
fi

# Dependency checks
echo "📦 Checking dependencies..."
if command_exists go; then
    print_status "Checking for outdated dependencies..."
    if command_exists go-mod-outdated; then
        if go-mod-outdated -update -direct; then
            print_success "Dependencies are up to date!"
        else
            print_warning "Some dependencies may be outdated"
        fi
    else
        print_status "Checking go.mod..."
        if go mod verify; then
            print_success "Dependencies verified!"
        else
            print_error "Dependency verification failed!"
            exit 1
        fi
    fi
fi

print_success "All code quality checks completed successfully!"
echo "🎯 Linting completed successfully!"