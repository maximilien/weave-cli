#!/bin/bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2025 dr.max

# OpenSearch local management script
# Commands: start, stop, status, logs, clean
# Uses podman or docker via container abstraction

set -euo pipefail

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../../.." && pwd)"
CONTAINER_RUN="${SCRIPT_DIR}/../container/run.sh"
STORAGE_DIR="${PROJECT_ROOT}/local/storage/opensearch_storage"

# OpenSearch container configuration
OPENSEARCH_CONTAINER_NAME="opensearch"
OPENSEARCH_IMAGE="opensearchproject/opensearch:latest"
OPENSEARCH_HTTP_PORT="9200"
OPENSEARCH_PERF_PORT="9600"
OPENSEARCH_PASSWORD="${OPENSEARCH_PASSWORD:-weaveOpenSearch123!}"

usage() {
    cat <<EOF
Usage: $(basename "$0") <command>

Commands:
    start       Start OpenSearch container
    stop        Stop OpenSearch container
    status      Show OpenSearch container status
    logs        Show OpenSearch container logs
    clean       Stop and remove OpenSearch container and volumes
    help        Show this help message

Environment:
    Uses podman or docker (auto-detected, podman preferred)
    OpenSearch listens on:
        - HTTP/REST API: http://localhost:${OPENSEARCH_HTTP_PORT}
        - Performance Analyzer: http://localhost:${OPENSEARCH_PERF_PORT}
    Storage: ${STORAGE_DIR}
    Default credentials:
        - Username: admin
        - Password: ${OPENSEARCH_PASSWORD} (set via OPENSEARCH_PASSWORD env var)

Examples:
    ./opensearch.sh start
    ./opensearch.sh status
    ./opensearch.sh logs
    ./opensearch.sh stop

    # Set custom password
    OPENSEARCH_PASSWORD=mypassword ./opensearch.sh start
EOF
}

ensure_storage_dir() {
    if [ ! -d "$STORAGE_DIR" ]; then
        echo "📁 Creating OpenSearch storage directory: $STORAGE_DIR"
        mkdir -p "$STORAGE_DIR"
    fi
}

start_opensearch() {
    echo "🚀 Starting OpenSearch..."

    # Check if already running
    if "${CONTAINER_RUN}" ps --format "{{.Names}}" | grep -q "^${OPENSEARCH_CONTAINER_NAME}$"; then
        echo "⚠️  OpenSearch is already running"
        echo ""
        status_opensearch
        return 0
    fi

    ensure_storage_dir

    # Start OpenSearch container
    # Note: Memory settings tuned for systems with limited RAM (< 2GB)
    # Container limit: 1.5GB, JVM heap: 768MB (balanced for stability)
    # Removed bootstrap.memory_lock to avoid OOM on constrained systems
    "${CONTAINER_RUN}" run -d \
        --name "${OPENSEARCH_CONTAINER_NAME}" \
        --memory="1536m" \
        --memory-swap="1536m" \
        -p "${OPENSEARCH_HTTP_PORT}:9200" \
        -p "${OPENSEARCH_PERF_PORT}:9600" \
        -e "discovery.type=single-node" \
        -e "OPENSEARCH_INITIAL_ADMIN_PASSWORD=${OPENSEARCH_PASSWORD}" \
        -e "DISABLE_INSTALL_DEMO_CONFIG=true" \
        -e "DISABLE_SECURITY_PLUGIN=true" \
        -e "OPENSEARCH_JAVA_OPTS=-Xms768m -Xmx768m" \
        --ulimit nofile=65536:65536 \
        -v "${STORAGE_DIR}:/usr/share/opensearch/data:z" \
        "${OPENSEARCH_IMAGE}"

    # Wait for OpenSearch to be ready
    echo "⏳ Waiting for OpenSearch to be ready..."
    sleep 5

    # Check health
    MAX_RETRIES=30
    RETRY_COUNT=0
    while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
        if curl -s "http://localhost:${OPENSEARCH_HTTP_PORT}/_cluster/health" > /dev/null 2>&1; then
            echo "✅ OpenSearch started successfully"
            echo ""
            echo "Connection details:"
            echo "  HTTP endpoint: http://localhost:${OPENSEARCH_HTTP_PORT}"
            echo "  Performance Analyzer: http://localhost:${OPENSEARCH_PERF_PORT}"
            echo "  Storage: ${STORAGE_DIR}"
            echo ""
            echo "Security: Disabled (for local development)"
            echo ""
            echo "Environment variables for weave-cli:"
            echo "  export OPENSEARCH_URL=\"http://localhost:${OPENSEARCH_HTTP_PORT}\""
            echo ""
            echo "Check status with: $(basename "$0") status"
            echo "View logs with: $(basename "$0") logs"
            echo ""
            echo "Test with: weave health check --opensearch-local"
            return 0
        fi
        RETRY_COUNT=$((RETRY_COUNT + 1))
        sleep 1
    done

    echo "⚠️  OpenSearch started but health check timed out"
    echo "   Check logs with: $(basename "$0") logs"
}

stop_opensearch() {
    echo "🛑 Stopping OpenSearch..."

    if ! "${CONTAINER_RUN}" ps --format "{{.Names}}" | grep -q "^${OPENSEARCH_CONTAINER_NAME}$"; then
        echo "⚠️  OpenSearch is not running"
        return 0
    fi

    "${CONTAINER_RUN}" stop "${OPENSEARCH_CONTAINER_NAME}"
    "${CONTAINER_RUN}" rm "${OPENSEARCH_CONTAINER_NAME}"

    echo "✅ OpenSearch stopped"
}

status_opensearch() {
    # Check if container exists and is running
    if "${CONTAINER_RUN}" ps --format "{{.Names}}" | grep -q "^${OPENSEARCH_CONTAINER_NAME}$"; then
        echo "✅ OpenSearch is running"
        echo ""
        "${CONTAINER_RUN}" ps --filter "name=${OPENSEARCH_CONTAINER_NAME}" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
        echo ""

        # Try to get health status
        if curl -s "http://localhost:${OPENSEARCH_HTTP_PORT}/_cluster/health" > /dev/null 2>&1; then
            echo "🟢 Health check: OK"
            echo ""

            # Get cluster info
            CLUSTER_INFO=$(curl -s "http://localhost:${OPENSEARCH_HTTP_PORT}/_cluster/health" 2>/dev/null || echo "{}")
            CLUSTER_NAME=$(echo "$CLUSTER_INFO" | grep -o '"cluster_name":"[^"]*"' | cut -d'"' -f4 || echo "unknown")
            STATUS=$(echo "$CLUSTER_INFO" | grep -o '"status":"[^"]*"' | cut -d'"' -f4 || echo "unknown")
            NODES=$(echo "$CLUSTER_INFO" | grep -o '"number_of_nodes":[0-9]*' | cut -d':' -f2 || echo "0")

            echo "📦 Cluster: $CLUSTER_NAME"
            echo "📊 Status: $STATUS"
            echo "🖥️  Nodes: $NODES"

            # Get version
            VERSION=$(curl -s "http://localhost:${OPENSEARCH_HTTP_PORT}/" 2>/dev/null | grep -o '"number" : "[^"]*"' | cut -d'"' -f4 || echo "unknown")
            echo "📦 Version: $VERSION"
        else
            echo "🔴 Health check: FAILED (may still be starting up)"
        fi
    else
        echo "❌ OpenSearch is not running"
        echo ""
        echo "Start with: $(basename "$0") start"
        return 1
    fi
}

logs_opensearch() {
    if ! "${CONTAINER_RUN}" ps --format "{{.Names}}" | grep -q "^${OPENSEARCH_CONTAINER_NAME}$"; then
        echo "❌ OpenSearch is not running" >&2
        exit 1
    fi

    echo "📋 OpenSearch logs (Ctrl+C to exit):"
    echo ""
    "${CONTAINER_RUN}" logs -f "${OPENSEARCH_CONTAINER_NAME}"
}

clean_opensearch() {
    echo "🧹 Cleaning OpenSearch (removing container and storage)..."

    # Stop and remove container
    if "${CONTAINER_RUN}" ps -a --format "{{.Names}}" | grep -q "^${OPENSEARCH_CONTAINER_NAME}$"; then
        "${CONTAINER_RUN}" stop "${OPENSEARCH_CONTAINER_NAME}" 2>/dev/null || true
        "${CONTAINER_RUN}" rm "${OPENSEARCH_CONTAINER_NAME}" 2>/dev/null || true
    fi

    # Ask before removing storage
    if [ -d "$STORAGE_DIR" ]; then
        echo ""
        echo "⚠️  This will delete all OpenSearch data in: $STORAGE_DIR"
        read -p "Continue? (y/N) " -n 1 -r
        echo ""
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            rm -rf "$STORAGE_DIR"
            echo "✅ Storage removed"
        else
            echo "ℹ️  Storage kept: $STORAGE_DIR"
        fi
    fi

    echo "✅ OpenSearch cleaned"
}

# Main command dispatch
case "${1:-}" in
    start)
        start_opensearch
        ;;
    stop)
        stop_opensearch
        ;;
    status)
        status_opensearch
        ;;
    logs)
        logs_opensearch
        ;;
    clean)
        clean_opensearch
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
