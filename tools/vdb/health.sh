#!/bin/bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2025 dr.max

# Health check utility for all configured VDBs
# Checks connectivity and status of vector databases

set -euo pipefail

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

usage() {
    cat <<EOF
Usage: $(basename "$0") [options]

Options:
    --local     Check only local VDB instances
    --cloud     Check only cloud VDB instances
    --verbose   Show detailed health information
    -h, --help  Show this help message

Examples:
    ./health.sh
    ./health.sh --local
    ./health.sh --verbose
EOF
}

check_local_vdb() {
    local vdb="$1"
    local script="${SCRIPT_DIR}/local/${vdb}.sh"

    if [ ! -f "$script" ]; then
        echo -e "${YELLOW}⚠${NC}  ${vdb}: No management script found"
        return 1
    fi

    if "$script" status &>/dev/null; then
        echo -e "${GREEN}✓${NC} ${vdb}: Running"
        return 0
    else
        echo -e "${RED}✗${NC} ${vdb}: Not running"
        return 1
    fi
}

check_cloud_vdb() {
    local vdb_type="$1"
    local weave_bin="${PROJECT_ROOT}/bin/weave"

    if [ ! -f "$weave_bin" ]; then
        echo -e "${YELLOW}⚠${NC}  ${vdb_type}: weave binary not found (run ./build.sh)"
        return 1
    fi

    # Use weave health check command
    if "$weave_bin" health check --vector-db-type "$vdb_type" &>/dev/null; then
        echo -e "${GREEN}✓${NC} ${vdb_type}: Connected"
        return 0
    else
        echo -e "${YELLOW}⚠${NC}  ${vdb_type}: Not configured or unreachable"
        return 1
    fi
}

check_all_local() {
    echo "Checking local VDB instances..."
    echo ""

    local status=0

    # Check Milvus
    check_local_vdb "milvus" || status=1

    # Check Qdrant
    check_local_vdb "qdrant" || status=1

    return $status
}

check_all_cloud() {
    echo "Checking cloud VDB instances..."
    echo ""

    local status=0

    # Check Weaviate Cloud
    check_cloud_vdb "weaviate-cloud" || status=1

    # Check Supabase
    check_cloud_vdb "supabase" || status=1

    # Check MongoDB
    check_cloud_vdb "mongodb" || status=1

    # Check Milvus Cloud (Zilliz)
    check_cloud_vdb "milvus-cloud" || status=1

    # Check Qdrant Cloud
    check_cloud_vdb "qdrant-cloud" || status=1

    # Check Chroma Cloud
    check_cloud_vdb "chroma-cloud" || status=1

    return $status
}

check_all() {
    echo "=== Vector Database Health Check ==="
    echo ""

    local local_status=0
    local cloud_status=0

    check_all_local || local_status=1
    echo ""
    check_all_cloud || cloud_status=1

    echo ""
    if [ $local_status -eq 0 ] && [ $cloud_status -eq 0 ]; then
        echo -e "${GREEN}All VDBs are healthy${NC}"
        return 0
    else
        echo -e "${YELLOW}Some VDBs are unavailable${NC}"
        return 1
    fi
}

# Parse arguments
CHECK_LOCAL=false
CHECK_CLOUD=false
VERBOSE=false

while [ $# -gt 0 ]; do
    case "$1" in
        --local)
            CHECK_LOCAL=true
            shift
            ;;
        --cloud)
            CHECK_CLOUD=true
            shift
            ;;
        --verbose)
            export VERBOSE=true
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            echo "❌ Unknown option: $1" >&2
            echo "" >&2
            usage
            exit 1
            ;;
    esac
done

# Main execution
if [ "$CHECK_LOCAL" = true ] && [ "$CHECK_CLOUD" = false ]; then
    check_all_local
elif [ "$CHECK_CLOUD" = true ] && [ "$CHECK_LOCAL" = false ]; then
    check_all_cloud
else
    # Check all by default
    check_all
fi
