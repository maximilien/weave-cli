#!/usr/bin/env bash

# Add SPDX license headers to all Go source files
# Usage: ./tools/add_license_headers.sh

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_header() {
    echo -e "${BLUE}[LICENSE]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# License header template
LICENSE_HEADER="// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

"

print_header "Adding SPDX license headers to all Go files..."

# Counter for processed files
count=0

# Find all Go files and add license headers
while IFS= read -r -d '' file; do
    # Skip if file already has license header
    if head -n 2 "$file" | grep -q "SPDX-License-Identifier"; then
        print_warning "Skipping $file (already has license header)"
        continue
    fi
    
    # Create temporary file with license header
    temp_file=$(mktemp)
    echo "$LICENSE_HEADER" > "$temp_file"
    cat "$file" >> "$temp_file"
    
    # Replace original file
    mv "$temp_file" "$file"
    
    count=$((count + 1))
    print_success "Added license header to: $file"
done < <(find . -name "*.go" -type f -print0)

print_success "Added license headers to $count Go files"
print_header "License header addition complete!"