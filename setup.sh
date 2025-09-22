#!/usr/bin/env bash

# SPDX-License-Identifier: MIT
# Copyright (c) 2025 dr.max

set -e

echo "🚀 Setting up development environment for Weave CLI..."

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
    echo -e "${BLUE}[SETUP]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to detect OS
detect_os() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo "macos"
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        echo "linux"
    else
        echo "unknown"
    fi
}

# Function to detect CI environment
is_ci() {
    [[ -n "${CI:-}" ]] || [[ -n "${GITHUB_ACTIONS:-}" ]] || [[ -n "${GITLAB_CI:-}" ]] || [[ -n "${JENKINS_URL:-}" ]]
}

OS=$(detect_os)
CI_ENV=$(is_ci && echo "true" || echo "false")

print_header "Setting up linting tools for Weave CLI development..."

if [ "$CI_ENV" = "true" ]; then
    print_status "Running in CI environment - using automated setup"
fi

# Check if Go is installed
if ! command_exists go; then
    print_error "Go is not installed. Please install Go first:"
    print_status "Visit: https://golang.org/doc/install"
    exit 1
fi

print_status "Go version: $(go version)"

# Install golangci-lint
print_header "Installing golangci-lint..."
if ! command_exists golangci-lint; then
    print_status "Installing golangci-lint..."
    if command_exists curl; then
        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.54.2
        print_success "golangci-lint installed successfully!"
    else
        print_error "curl not found, please install golangci-lint manually"
        print_status "Visit: https://golangci-lint.run/usage/install/"
        exit 1
    fi
else
    print_success "golangci-lint is already installed"
fi

# Install goimports
print_header "Installing goimports..."
if ! command_exists goimports; then
    print_status "Installing goimports..."
    go install golang.org/x/tools/cmd/goimports@latest
    print_success "goimports installed successfully!"
else
    print_success "goimports is already installed"
fi

# Install govulncheck (official Go vulnerability scanner)
print_header "Installing govulncheck..."
if ! command_exists govulncheck; then
    print_status "Installing govulncheck..."
    go install golang.org/x/vuln/cmd/govulncheck@latest
    print_success "govulncheck installed successfully!"
else
    print_success "govulncheck is already installed"
fi

# Install gosec (optional, for additional security checks)
print_header "Installing gosec..."
if ! command_exists gosec; then
    print_status "Installing gosec..."
    print_status "CI_ENV: $CI_ENV"
    print_status "GITHUB_ACTIONS: ${GITHUB_ACTIONS:-not set}"
    
    # Try the install script method (most reliable)
    print_status "Using install script method..."
    if curl -sfL https://raw.githubusercontent.com/securecodewarrior/gosec/master/install.sh | sh -s -- -b $(go env GOPATH)/bin; then
        print_success "gosec installed via install script!"
    else
        print_warning "gosec installation failed - skipping (govulncheck provides security coverage)"
        print_status "Note: govulncheck is already installed and provides comprehensive security scanning"
    fi
else
    print_success "gosec is already installed"
fi

# Install shellcheck
print_header "Installing shellcheck..."
if ! command_exists shellcheck; then
    print_status "Installing shellcheck..."
    case $OS in
        "macos")
            if command_exists brew; then
                brew install shellcheck
                print_success "shellcheck installed successfully via Homebrew!"
            else
                print_warning "Homebrew not found. Please install shellcheck manually:"
                print_status "Visit: https://github.com/koalaman/shellcheck#installing"
            fi
            ;;
        "linux")
            if command_exists apt-get; then
                sudo apt-get update && sudo apt-get install -y shellcheck
                print_success "shellcheck installed successfully via apt-get!"
            elif command_exists yum; then
                sudo yum install -y shellcheck
                print_success "shellcheck installed successfully via yum!"
            else
                print_warning "Package manager not found. Please install shellcheck manually:"
                print_status "Visit: https://github.com/koalaman/shellcheck#installing"
            fi
            ;;
        *)
            print_warning "Unknown OS. Please install shellcheck manually:"
            print_status "Visit: https://github.com/koalaman/shellcheck#installing"
            ;;
    esac
else
    print_success "shellcheck is already installed"
fi

# Install yamllint
print_header "Installing yamllint..."
if ! command_exists yamllint; then
    print_status "Installing yamllint..."
    if command_exists pip3; then
        pip3 install yamllint
        print_success "yamllint installed successfully!"
    elif command_exists pip; then
        pip install yamllint
        print_success "yamllint installed successfully!"
    else
        print_warning "pip not found. Please install yamllint manually:"
        print_status "Visit: https://yamllint.readthedocs.io/en/stable/quickstart.html"
    fi
else
    print_success "yamllint is already installed"
fi

# Install markdownlint
print_header "Installing markdownlint..."
if ! command_exists markdownlint; then
    print_status "Installing markdownlint..."
    if command_exists npm; then
        # Try installing without sudo first, then with sudo if needed
        if npm install -g markdownlint-cli; then
            print_success "markdownlint installed successfully!"
        else
            print_warning "Global install failed, trying with sudo..."
            if sudo npm install -g markdownlint-cli; then
                print_success "markdownlint installed successfully with sudo!"
            else
                print_warning "markdownlint installation failed - skipping"
                print_status "You can install it manually: npm install -g markdownlint-cli"
            fi
        fi
    else
        print_warning "npm not found. Please install markdownlint manually:"
        print_status "Visit: https://github.com/igorshubovych/markdownlint-cli"
    fi
else
    print_success "markdownlint is already installed"
fi

# Install go-mod-outdated (optional)
print_header "Installing go-mod-outdated..."
if ! command_exists go-mod-outdated; then
    print_status "Installing go-mod-outdated..."
    go install github.com/psampaz/go-mod-outdated@latest
    print_success "go-mod-outdated installed successfully!"
else
    print_success "go-mod-outdated is already installed"
fi

# Verify installations
print_header "Verifying installations..."
echo ""

tools=("go" "golangci-lint" "goimports" "govulncheck" "gosec" "shellcheck" "yamllint" "markdownlint" "go-mod-outdated")

all_installed=true
for tool in "${tools[@]}"; do
    if command_exists "$tool"; then
        print_success "✓ $tool is installed"
    else
        print_warning "✗ $tool is not installed"
        all_installed=false
    fi
done

echo ""
if [ "$all_installed" = true ]; then
    print_success "🎉 All linting tools are installed and ready!"
    print_status "You can now run ./lint.sh without warnings"
else
    print_warning "⚠️  Some tools are missing. Please install them manually."
    print_status "Run this script again after installing missing tools."
fi

# Update PATH if needed
print_header "Checking PATH configuration..."
GOPATH=$(go env GOPATH)
if [[ ":$PATH:" != *":$GOPATH/bin:"* ]]; then
    print_warning "GOPATH/bin is not in your PATH"
    print_status "Add this to your shell profile (.bashrc, .zshrc, etc.):"
    echo "export PATH=\$PATH:$GOPATH/bin"
    echo ""
fi

print_success "🚀 Setup completed!"
print_status "Run './lint.sh' to test the linting setup"