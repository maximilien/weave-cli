#!/bin/bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2025 dr.max

# Generic VDB local management script
# Manages multiple local vector database instances

set -euo pipefail

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Supported VDBs
SUPPORTED_VDBS=("milvus" "qdrant")

usage() {
    cat <<EOF
Usage: $(basename "$0") <vdb> <command>

Supported VDBs:
$(printf '    %s\n' "${SUPPORTED_VDBS[@]}")

Commands:
    start       Start the VDB instance
    stop        Stop the VDB instance
    status      Show VDB instance status
    logs        Show VDB instance logs
    clean       Stop and remove VDB instance and volumes
    help        Show help for specific VDB

Examples:
    ./manager.sh milvus start
    ./manager.sh milvus status
    ./manager.sh qdrant start
    ./manager.sh qdrant logs
    ./manager.sh milvus stop

    # Start all VDBs
    ./manager.sh all start

    # Check status of all VDBs
    ./manager.sh all status
EOF
}

is_supported_vdb() {
    local vdb="$1"
    for supported in "${SUPPORTED_VDBS[@]}"; do
        if [ "$vdb" = "$supported" ]; then
            return 0
        fi
    done
    return 1
}

manage_vdb() {
    local vdb="$1"
    local command="$2"

    if ! is_supported_vdb "$vdb"; then
        echo "❌ Unsupported VDB: $vdb" >&2
        echo "" >&2
        echo "Supported VDBs: ${SUPPORTED_VDBS[*]}" >&2
        exit 1
    fi

    local vdb_script="${SCRIPT_DIR}/${vdb}.sh"

    if [ ! -f "$vdb_script" ]; then
        echo "❌ VDB script not found: $vdb_script" >&2
        exit 1
    fi

    # Execute the VDB-specific script
    "$vdb_script" "$command"
}

manage_all() {
    local command="$1"

    case "$command" in
        start|stop|status|clean)
            echo "Running '$command' for all VDBs..."
            echo ""
            for vdb in "${SUPPORTED_VDBS[@]}"; do
                echo "==> $vdb"
                manage_vdb "$vdb" "$command" || true
                echo ""
            done
            ;;
        *)
            echo "❌ Command '$command' not supported for 'all'" >&2
            echo "Supported commands: start, stop, status, clean" >&2
            exit 1
            ;;
    esac
}

# Main command dispatch
if [ $# -lt 2 ]; then
    usage
    exit 1
fi

VDB="$1"
COMMAND="$2"

if [ "$VDB" = "all" ]; then
    manage_all "$COMMAND"
elif [ "$VDB" = "help" ] || [ "$VDB" = "--help" ] || [ "$VDB" = "-h" ]; then
    usage
else
    manage_vdb "$VDB" "$COMMAND"
fi
