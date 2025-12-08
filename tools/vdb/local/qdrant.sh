#!/bin/bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2025 dr.max

# Qdrant local management script
# Commands: start, stop, status, logs, clean
# Uses podman or docker via container abstraction

set -euo pipefail

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../../.." && pwd)"
CONTAINER_RUN="${SCRIPT_DIR}/../container/run.sh"
STORAGE_DIR="${PROJECT_ROOT}/local/storage/qdrant_storage"

# Qdrant container configuration
QDRANT_CONTAINER_NAME="qdrant"
QDRANT_IMAGE="qdrant/qdrant:latest"
QDRANT_HTTP_PORT="6333"
QDRANT_GRPC_PORT="6334"

usage() {
    cat <<EOF
Usage: $(basename "$0") <command>

Commands:
    start       Start Qdrant container
    stop        Stop Qdrant container
    status      Show Qdrant container status
    logs        Show Qdrant container logs
    clean       Stop and remove Qdrant container and volumes
    help        Show this help message

Environment:
    Uses podman or docker (auto-detected, podman preferred)
    Qdrant listens on:
        - HTTP: localhost:${QDRANT_HTTP_PORT}
        - gRPC: localhost:${QDRANT_GRPC_PORT}
    Storage: ${STORAGE_DIR}

Examples:
    ./qdrant.sh start
    ./qdrant.sh status
    ./qdrant.sh logs
    ./qdrant.sh stop
EOF
}

ensure_storage_dir() {
    if [ ! -d "$STORAGE_DIR" ]; then
        echo "📁 Creating Qdrant storage directory: $STORAGE_DIR"
        mkdir -p "$STORAGE_DIR"
    fi
}

start_qdrant() {
    echo "🚀 Starting Qdrant..."

    # Check if already running
    if "${CONTAINER_RUN}" ps --format "{{.Names}}" | grep -q "^${QDRANT_CONTAINER_NAME}$"; then
        echo "⚠️  Qdrant is already running"
        echo ""
        status_qdrant
        return 0
    fi

    ensure_storage_dir

    # Start Qdrant container
    "${CONTAINER_RUN}" run -d \
        --name "${QDRANT_CONTAINER_NAME}" \
        -p "${QDRANT_HTTP_PORT}:6333" \
        -p "${QDRANT_GRPC_PORT}:6334" \
        -v "${STORAGE_DIR}:/qdrant/storage:z" \
        "${QDRANT_IMAGE}"

    # Wait for Qdrant to be ready
    echo "⏳ Waiting for Qdrant to be ready..."
    sleep 2

    # Check health
    MAX_RETRIES=10
    RETRY_COUNT=0
    while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
        if curl -s "http://localhost:${QDRANT_HTTP_PORT}/healthz" > /dev/null 2>&1; then
            echo "✅ Qdrant started successfully"
            echo ""
            echo "Connection details:"
            echo "  HTTP endpoint: http://localhost:${QDRANT_HTTP_PORT}"
            echo "  gRPC endpoint: localhost:${QDRANT_GRPC_PORT}"
            echo "  Storage: ${STORAGE_DIR}"
            echo ""
            echo "Check status with: $(basename "$0") status"
            echo "View logs with: $(basename "$0") logs"
            return 0
        fi
        RETRY_COUNT=$((RETRY_COUNT + 1))
        sleep 1
    done

    echo "⚠️  Qdrant started but health check timed out"
    echo "   Check logs with: $(basename "$0") logs"
}

stop_qdrant() {
    echo "🛑 Stopping Qdrant..."

    if ! "${CONTAINER_RUN}" ps --format "{{.Names}}" | grep -q "^${QDRANT_CONTAINER_NAME}$"; then
        echo "⚠️  Qdrant is not running"
        return 0
    fi

    "${CONTAINER_RUN}" stop "${QDRANT_CONTAINER_NAME}"
    "${CONTAINER_RUN}" rm "${QDRANT_CONTAINER_NAME}"

    echo "✅ Qdrant stopped"
}

status_qdrant() {
    # Check if container exists and is running
    if "${CONTAINER_RUN}" ps --format "{{.Names}}" | grep -q "^${QDRANT_CONTAINER_NAME}$"; then
        echo "✅ Qdrant is running"
        echo ""
        "${CONTAINER_RUN}" ps --filter "name=${QDRANT_CONTAINER_NAME}" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
        echo ""

        # Try to get health status
        if curl -s "http://localhost:${QDRANT_HTTP_PORT}/healthz" > /dev/null 2>&1; then
            echo "🟢 Health check: OK"
        else
            echo "🔴 Health check: FAILED"
        fi
    else
        echo "❌ Qdrant is not running"
        echo ""
        echo "Start with: $(basename "$0") start"
        return 1
    fi
}

logs_qdrant() {
    if ! "${CONTAINER_RUN}" ps --format "{{.Names}}" | grep -q "^${QDRANT_CONTAINER_NAME}$"; then
        echo "❌ Qdrant is not running" >&2
        exit 1
    fi

    echo "📋 Qdrant logs (Ctrl+C to exit):"
    echo ""
    "${CONTAINER_RUN}" logs -f "${QDRANT_CONTAINER_NAME}"
}

clean_qdrant() {
    echo "🧹 Cleaning Qdrant (removing container and storage)..."

    # Stop and remove container
    if "${CONTAINER_RUN}" ps -a --format "{{.Names}}" | grep -q "^${QDRANT_CONTAINER_NAME}$"; then
        "${CONTAINER_RUN}" stop "${QDRANT_CONTAINER_NAME}" 2>/dev/null || true
        "${CONTAINER_RUN}" rm "${QDRANT_CONTAINER_NAME}" 2>/dev/null || true
    fi

    # Ask before removing storage
    if [ -d "$STORAGE_DIR" ]; then
        echo ""
        echo "⚠️  This will delete all Qdrant data in: $STORAGE_DIR"
        read -p "Continue? (y/N) " -n 1 -r
        echo ""
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            rm -rf "$STORAGE_DIR"
            echo "✅ Storage removed"
        else
            echo "ℹ️  Storage kept: $STORAGE_DIR"
        fi
    fi

    echo "✅ Qdrant cleaned"
}

# Main command dispatch
case "${1:-}" in
    start)
        start_qdrant
        ;;
    stop)
        stop_qdrant
        ;;
    status)
        status_qdrant
        ;;
    logs)
        logs_qdrant
        ;;
    clean)
        clean_qdrant
        ;;
    help|--help|-h)
        usage
        ;;
    *)
        echo "❌ Unknown command: ${1:-}" >&2
        echo "" >&2
        usage
        exit 1
        ;;
esac
