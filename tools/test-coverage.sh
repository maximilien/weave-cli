#!/bin/bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2025 dr.max

# Test Coverage Analysis Script
# Generates coverage reports for Weave CLI

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Directories
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COVERAGE_DIR="${PROJECT_ROOT}/coverage"
COVERAGE_OUT="${COVERAGE_DIR}/coverage.out"
COVERAGE_HTML="${COVERAGE_DIR}/coverage.html"

# Create coverage directory
mkdir -p "${COVERAGE_DIR}"

echo -e "${BLUE}=== Weave CLI Test Coverage Analysis ===${NC}"
echo ""

# Run tests with coverage
echo -e "${YELLOW}Running tests with coverage...${NC}"
cd "${PROJECT_ROOT}"

# Run all tests with coverage
go test ./src/pkg/... \
    -coverprofile="${COVERAGE_OUT}" \
    -covermode=atomic \
    -v 2>&1 | grep -E "^(PASS|FAIL|ok|FAIL|coverage:)" || true

if [ ! -f "${COVERAGE_OUT}" ]; then
    echo -e "${RED}❌ Coverage file not generated${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}✅ Tests completed${NC}"
echo ""

# Generate HTML report
echo -e "${YELLOW}Generating HTML report...${NC}"
go tool cover -html="${COVERAGE_OUT}" -o "${COVERAGE_HTML}"
echo -e "${GREEN}✅ HTML report: ${COVERAGE_HTML}${NC}"
echo ""

# Display overall coverage
echo -e "${BLUE}=== Overall Coverage ===${NC}"
OVERALL=$(go tool cover -func="${COVERAGE_OUT}" | tail -1)
echo "${OVERALL}"
echo ""

# Extract coverage percentage
COVERAGE_PCT=$(echo "${OVERALL}" | awk '{print $3}' | sed 's/%//')

# Color code based on coverage
if (( $(echo "${COVERAGE_PCT} >= 70" | bc -l) )); then
    COLOR="${GREEN}"
    STATUS="✅ Excellent"
elif (( $(echo "${COVERAGE_PCT} >= 50" | bc -l) )); then
    COLOR="${YELLOW}"
    STATUS="⚠️  Good"
else
    COLOR="${RED}"
    STATUS="❌ Needs Work"
fi

echo -e "${COLOR}${STATUS}: ${COVERAGE_PCT}% coverage${NC}"
echo ""

# Coverage by package
echo -e "${BLUE}=== Coverage by Package ===${NC}"
go tool cover -func="${COVERAGE_OUT}" | \
    grep -E "github.com/maximilien/weave-cli/src/pkg/vectordb" | \
    grep -v "total:" | \
    awk '{
        pkg=$1
        pct=$3
        # Extract VDB name from path
        split(pkg, parts, "/")
        vdb=parts[length(parts)]
        printf "%-20s %s\n", vdb":", pct
    }' | \
    sort -t: -k2 -rn | \
    head -20

echo ""

# VDB-specific summary
echo -e "${BLUE}=== VDB Package Summary ===${NC}"
for vdb in qdrant weaviate mongodb milvus chroma neo4j supabase pinecone elasticsearch opensearch; do
    PKG_COVERAGE=$(go tool cover -func="${COVERAGE_OUT}" | \
        grep "github.com/maximilien/weave-cli/src/pkg/vectordb/${vdb}" | \
        grep "total:" | \
        awk '{print $3}' || echo "0.0%")

    if [ -n "${PKG_COVERAGE}" ] && [ "${PKG_COVERAGE}" != "0.0%" ]; then
        PCT=$(echo "${PKG_COVERAGE}" | sed 's/%//')
        if (( $(echo "${PCT} >= 60" | bc -l) )); then
            echo -e "  ${GREEN}✅${NC} ${vdb}: ${PKG_COVERAGE}"
        elif (( $(echo "${PCT} >= 40" | bc -l) )); then
            echo -e "  ${YELLOW}⚠️${NC}  ${vdb}: ${PKG_COVERAGE}"
        else
            echo -e "  ${RED}❌${NC} ${vdb}: ${PKG_COVERAGE}"
        fi
    else
        echo -e "  ${RED}❌${NC} ${vdb}: No coverage data"
    fi
done

echo ""

# Recommendations
echo -e "${BLUE}=== Recommendations ===${NC}"
if (( $(echo "${COVERAGE_PCT} < 70" | bc -l) )); then
    echo "• Add more unit tests to increase coverage"
    echo "• Focus on VDBs with <60% coverage"
    echo "• Test error paths and edge cases"
fi

if (( $(echo "${COVERAGE_PCT} >= 70" | bc -l) )); then
    echo "✅ Great coverage! Keep it up!"
fi

echo ""
echo -e "${BLUE}=== Files Generated ===${NC}"
echo "• Coverage profile: ${COVERAGE_OUT}"
echo "• HTML report: ${COVERAGE_HTML}"
echo ""
echo -e "${GREEN}Open HTML report: open ${COVERAGE_HTML}${NC}"
