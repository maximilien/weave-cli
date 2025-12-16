#!/bin/bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2025 dr.max

# Elasticsearch local management script
# Commands: start, stop, status, logs, clean
# Uses podman or docker via container abstraction

set -euo pipefail

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../../.." && pwd)"
CONTAINER_RUN="${SCRIPT_DIR}/../container/run.sh"
STORAGE_DIR="${PROJECT_ROOT}/local/storage/elasticsearch_storage"

# Elasticsearch container configuration
ELASTICSEARCH_CONTAINER_NAME="elasticsearch"
ELASTICSEARCH_IMAGE="docker.elastic.co/elasticsearch/elasticsearch:8.11.0"
ELASTICSEARCH_HTTP_PORT="9200"
ELASTICSEARCH_TRANSPORT_PORT="9300"

usage() {
    cat <<EOF
Usage: $(basename "$0") <command>

Commands:
    start       Start Elasticsearch container
    stop        Stop Elasticsearch container
    status      Show Elasticsearch container status
    logs        Show Elasticsearch container logs
    clean       Stop and remove Elasticsearch container and volumes
    help        Show this help message

Environment:
    Uses podman or docker (auto-detected, podman preferred)
    Elasticsearch listens on:
        - HTTP/REST API: http://localhost:${ELASTICSEARCH_HTTP_PORT}
        - Transport: ${ELASTICSEARCH_TRANSPORT_PORT}
    Storage: ${STORAGE_DIR}
    Security: Disabled (for local development)

Examples:
    ./elasticsearch.sh start
    ./elasticsearch.sh status
    ./elasticsearch.sh logs
    ./elasticsearch.sh stop

Notes:
    ⚠️  Port ${ELASTICSEARCH_HTTP_PORT} conflicts with OpenSearch
    Only one can run at a time. Stop OpenSearch first if needed:
      ./opensearch.sh stop
EOF
}

ensure_storage_dir() {
    if [ ! -d "$STORAGE_DIR" ]; then
        echo "📁 Creating Elasticsearch storage directory: $STORAGE_DIR"
        mkdir -p "$STORAGE_DIR"
    fi
}

start_elasticsearch() {
    echo "🚀 Starting Elasticsearch..."

    # Check if already running
    if "${CONTAINER_RUN}" ps --format "{{.Names}}" | grep -q "^${ELASTICSEARCH_CONTAINER_NAME}$"; then
        echo "⚠️  Elasticsearch is already running"
        echo ""
        status_elasticsearch
        return 0
    fi

    # Check if OpenSearch is using the same port
    if lsof -Pi :${ELASTICSEARCH_HTTP_PORT} -sTCP:LISTEN -t >/dev/null 2>&1; then
        echo "⚠️  Port ${ELASTICSEARCH_HTTP_PORT} is already in use (possibly OpenSearch)"
        echo "   Stop OpenSearch first: ./opensearch.sh stop"
        return 1
    fi

    ensure_storage_dir

    # Start Elasticsearch container
    # Note: Memory settings - Elasticsearch 8.x needs more RAM than most VDBs
    # Using minimal heap (512MB) for development
    # Security disabled for local development
    # Note: On macOS with limited resources, consider using Docker instead of podman
    "${CONTAINER_RUN}" run -d \
        --name "${ELASTICSEARCH_CONTAINER_NAME}" \
        -p "${ELASTICSEARCH_HTTP_PORT}:9200" \
        -p "${ELASTICSEARCH_TRANSPORT_PORT}:9300" \
        -e "discovery.type=single-node" \
        -e "xpack.security.enabled=false" \
        -e "ES_JAVA_OPTS=-Xms512m -Xmx512m" \
        -e "bootstrap.memory_lock=false" \
        --ulimit nofile=65536:65536 \
        -v "${STORAGE_DIR}:/usr/share/elasticsearch/data:z" \
        "${ELASTICSEARCH_IMAGE}"

    # Wait for Elasticsearch to be ready
    echo "⏳ Waiting for Elasticsearch to be ready..."
    sleep 5

    # Check health
    MAX_RETRIES=30
    RETRY_COUNT=0
    while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
        if curl -s "http://localhost:${ELASTICSEARCH_HTTP_PORT}/_cluster/health" > /dev/null 2>&1; then
            echo "✅ Elasticsearch started successfully"
            echo ""
            echo "Connection details:"
            echo "  HTTP endpoint: http://localhost:${ELASTICSEARCH_HTTP_PORT}"
            echo "  Storage: ${STORAGE_DIR}"
            echo ""
            echo "Security: Disabled (for local development)"
            echo ""
            echo "Environment variables for weave-cli:"
            echo "  export ELASTICSEARCH_LOCAL_ADDRESS=\"http://localhost:${ELASTICSEARCH_HTTP_PORT}\""
            echo ""
            echo "Check status with: $(basename "$0") status"
            echo "View logs with: $(basename "$0") logs"
            echo ""
            echo "Test with: weave health check"
            return 0
        fi
        RETRY_COUNT=$((RETRY_COUNT + 1))
        sleep 1
    done

    echo "⚠️  Elasticsearch started but health check timed out"
    echo "   Check logs with: $(basename "$0") logs"
}

stop_elasticsearch() {
    echo "🛑 Stopping Elasticsearch..."

    if ! "${CONTAINER_RUN}" ps --format "{{.Names}}" | grep -q "^${ELASTICSEARCH_CONTAINER_NAME}$"; then
        echo "⚠️  Elasticsearch is not running"
        return 0
    fi

    "${CONTAINER_RUN}" stop "${ELASTICSEARCH_CONTAINER_NAME}"
    "${CONTAINER_RUN}" rm "${ELASTICSEARCH_CONTAINER_NAME}"

    echo "✅ Elasticsearch stopped"
}

status_elasticsearch() {
    # Check if container exists and is running
    if "${CONTAINER_RUN}" ps --format "{{.Names}}" | grep -q "^${ELASTICSEARCH_CONTAINER_NAME}$"; then
        echo "✅ Elasticsearch is running"
        echo ""
        "${CONTAINER_RUN}" ps --filter "name=${ELASTICSEARCH_CONTAINER_NAME}" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
        echo ""

        # Try to get health status
        if curl -s "http://localhost:${ELASTICSEARCH_HTTP_PORT}/_cluster/health" > /dev/null 2>&1; then
            echo "🟢 Health check: OK"
            echo ""

            # Get cluster info
            CLUSTER_INFO=$(curl -s "http://localhost:${ELASTICSEARCH_HTTP_PORT}/_cluster/health" 2>/dev/null || echo "{}")
            CLUSTER_NAME=$(echo "$CLUSTER_INFO" | grep -o '"cluster_name":"[^"]*"' | cut -d'"' -f4 || echo "unknown")
            STATUS=$(echo "$CLUSTER_INFO" | grep -o '"status":"[^"]*"' | cut -d'"' -f4 || echo "unknown")
            NODES=$(echo "$CLUSTER_INFO" | grep -o '"number_of_nodes":[0-9]*' | cut -d':' -f2 || echo "0")

            echo "📦 Cluster: $CLUSTER_NAME"
            echo "📊 Status: $STATUS"
            echo "🖥️  Nodes: $NODES"

            # Get version
            VERSION_INFO=$(curl -s "http://localhost:${ELASTICSEARCH_HTTP_PORT}/" 2>/dev/null || echo "{}")
            VERSION=$(echo "$VERSION_INFO" | grep -o '"number" : "[^"]*"' | cut -d'"' -f4 || echo "unknown")
            echo "📦 Version: $VERSION"
        else
            echo "🔴 Health check: FAILED (may still be starting up)"
        fi
    else
        echo "❌ Elasticsearch is not running"
        echo ""
        echo "Start with: $(basename "$0") start"
        return 1
    fi
}

logs_elasticsearch() {
    if ! "${CONTAINER_RUN}" ps --format "{{.Names}}" | grep -q "^${ELASTICSEARCH_CONTAINER_NAME}$"; then
        echo "❌ Elasticsearch is not running" >&2
        exit 1
    fi

    echo "📋 Elasticsearch logs (Ctrl+C to exit):"
    echo ""
    "${CONTAINER_RUN}" logs -f "${ELASTICSEARCH_CONTAINER_NAME}"
}

clean_elasticsearch() {
    echo "🧹 Cleaning Elasticsearch (removing container and storage)..."

    # Stop and remove container
    if "${CONTAINER_RUN}" ps -a --format "{{.Names}}" | grep -q "^${ELASTICSEARCH_CONTAINER_NAME}$"; then
        "${CONTAINER_RUN}" stop "${ELASTICSEARCH_CONTAINER_NAME}" 2>/dev/null || true
        "${CONTAINER_RUN}" rm "${ELASTICSEARCH_CONTAINER_NAME}" 2>/dev/null || true
    fi

    # Ask before removing storage
    if [ -d "$STORAGE_DIR" ]; then
        echo ""
        echo "⚠️  This will delete all Elasticsearch data in: $STORAGE_DIR"
        read -p "Continue? (y/N) " -n 1 -r
        echo ""
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            rm -rf "$STORAGE_DIR"
            echo "✅ Storage removed"
        else
            echo "ℹ️  Storage kept: $STORAGE_DIR"
        fi
    fi

    echo "✅ Elasticsearch cleaned"
}

# Main command dispatch
case "${1:-}" in
    start)
        start_elasticsearch
        ;;
    stop)
        stop_elasticsearch
        ;;
    status)
        status_elasticsearch
        ;;
    logs)
        logs_elasticsearch
        ;;
    clean)
        clean_elasticsearch
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
