#!/bin/bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2025 dr.max

# Weaviate local management script
# Commands: start, stop, status, logs, clean
# Uses podman or docker via container abstraction

set -euo pipefail

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../../.." && pwd)"
CONTAINER_RUN="${SCRIPT_DIR}/../container/run.sh"
STORAGE_DIR="${PROJECT_ROOT}/local/storage/weaviate_storage"

# Weaviate container configuration
WEAVIATE_CONTAINER_NAME="weaviate"
WEAVIATE_IMAGE="semitechnologies/weaviate:latest"
WEAVIATE_HTTP_PORT="8080"
WEAVIATE_GRPC_PORT="50051"

# Check if .env exists for OPENAI_API_KEY
ENV_FILE="${PROJECT_ROOT}/.env"
OPENAI_API_KEY="${OPENAI_API_KEY:-}"

if [ -f "$ENV_FILE" ] && [ -z "$OPENAI_API_KEY" ]; then
    # Load OPENAI_API_KEY from .env if available
    OPENAI_API_KEY=$(grep "^OPENAI_API_KEY=" "$ENV_FILE" 2>/dev/null | cut -d '=' -f2 | tr -d '"' || echo "")
fi

usage() {
    cat <<EOF
Usage: $(basename "$0") <command>

Commands:
    start       Start Weaviate container
    stop        Stop Weaviate container
    status      Show Weaviate container status
    logs        Show Weaviate container logs
    clean       Stop and remove Weaviate container and volumes
    help        Show this help message

Environment:
    Uses podman or docker (auto-detected, podman preferred)
    Weaviate listens on:
        - HTTP: http://localhost:${WEAVIATE_HTTP_PORT}
        - gRPC: localhost:${WEAVIATE_GRPC_PORT}
    Storage: ${STORAGE_DIR}

    Optional: Set OPENAI_API_KEY in .env for automatic embeddings

Examples:
    ./weaviate.sh start
    ./weaviate.sh status
    ./weaviate.sh logs
    ./weaviate.sh stop
EOF
}

ensure_storage_dir() {
    if [ ! -d "$STORAGE_DIR" ]; then
        echo "📁 Creating Weaviate storage directory: $STORAGE_DIR"
        mkdir -p "$STORAGE_DIR"
    fi
}

start_weaviate() {
    echo "🚀 Starting Weaviate..."

    # Check if already running
    if "${CONTAINER_RUN}" ps --format "{{.Names}}" | grep -q "^${WEAVIATE_CONTAINER_NAME}$"; then
        echo "⚠️  Weaviate is already running"
        echo ""
        status_weaviate
        return 0
    fi

    ensure_storage_dir

    # Build environment variables for container
    ENV_ARGS=()
    ENV_ARGS+=(-e "AUTHENTICATION_ANONYMOUS_ACCESS_ENABLED=true")
    ENV_ARGS+=(-e "PERSISTENCE_DATA_PATH=/var/lib/weaviate")
    ENV_ARGS+=(-e "QUERY_DEFAULTS_LIMIT=25")
    ENV_ARGS+=(-e "DEFAULT_VECTORIZER_MODULE=text2vec-openai")
    ENV_ARGS+=(-e "ENABLE_MODULES=text2vec-openai")

    if [ -n "$OPENAI_API_KEY" ]; then
        ENV_ARGS+=(-e "OPENAI_APIKEY=${OPENAI_API_KEY}")
        echo "✅ Using OPENAI_API_KEY for automatic embeddings"
    else
        echo "⚠️  OPENAI_API_KEY not set - you'll need to provide embeddings manually"
        echo "   Set OPENAI_API_KEY in .env or export it before running"
    fi

    # Start Weaviate container
    "${CONTAINER_RUN}" run -d \
        --name "${WEAVIATE_CONTAINER_NAME}" \
        -p "${WEAVIATE_HTTP_PORT}:8080" \
        -p "${WEAVIATE_GRPC_PORT}:50051" \
        -v "${STORAGE_DIR}:/var/lib/weaviate:z" \
        "${ENV_ARGS[@]}" \
        "${WEAVIATE_IMAGE}"

    # Wait for Weaviate to be ready
    echo "⏳ Waiting for Weaviate to be ready..."
    sleep 3

    # Check health
    MAX_RETRIES=15
    RETRY_COUNT=0
    while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
        if curl -s "http://localhost:${WEAVIATE_HTTP_PORT}/v1/.well-known/ready" > /dev/null 2>&1; then
            echo "✅ Weaviate started successfully"
            echo ""
            echo "Connection details:"
            echo "  HTTP endpoint: http://localhost:${WEAVIATE_HTTP_PORT}"
            echo "  gRPC endpoint: localhost:${WEAVIATE_GRPC_PORT}"
            echo "  Storage: ${STORAGE_DIR}"
            echo ""
            echo "Check status with: $(basename "$0") status"
            echo "View logs with: $(basename "$0") logs"
            echo ""
            echo "Test with: weave health check --weaviate-local"
            return 0
        fi
        RETRY_COUNT=$((RETRY_COUNT + 1))
        sleep 1
    done

    echo "⚠️  Weaviate started but health check timed out"
    echo "   Check logs with: $(basename "$0") logs"
}

stop_weaviate() {
    echo "🛑 Stopping Weaviate..."

    if ! "${CONTAINER_RUN}" ps --format "{{.Names}}" | grep -q "^${WEAVIATE_CONTAINER_NAME}$"; then
        echo "⚠️  Weaviate is not running"
        return 0
    fi

    "${CONTAINER_RUN}" stop "${WEAVIATE_CONTAINER_NAME}"
    "${CONTAINER_RUN}" rm "${WEAVIATE_CONTAINER_NAME}"

    echo "✅ Weaviate stopped"
}

status_weaviate() {
    # Check if container exists and is running
    if "${CONTAINER_RUN}" ps --format "{{.Names}}" | grep -q "^${WEAVIATE_CONTAINER_NAME}$"; then
        echo "✅ Weaviate is running"
        echo ""
        "${CONTAINER_RUN}" ps --filter "name=${WEAVIATE_CONTAINER_NAME}" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
        echo ""

        # Try to get health status
        if curl -s "http://localhost:${WEAVIATE_HTTP_PORT}/v1/.well-known/ready" > /dev/null 2>&1; then
            echo "🟢 Health check: OK"

            # Get version info
            VERSION=$(curl -s "http://localhost:${WEAVIATE_HTTP_PORT}/v1/meta" 2>/dev/null | grep -o '"version":"[^"]*"' | cut -d'"' -f4 || echo "unknown")
            echo "📦 Version: $VERSION"
        else
            echo "🔴 Health check: FAILED"
        fi
    else
        echo "❌ Weaviate is not running"
        echo ""
        echo "Start with: $(basename "$0") start"
        return 1
    fi
}

logs_weaviate() {
    if ! "${CONTAINER_RUN}" ps --format "{{.Names}}" | grep -q "^${WEAVIATE_CONTAINER_NAME}$"; then
        echo "❌ Weaviate is not running" >&2
        exit 1
    fi

    echo "📋 Weaviate logs (Ctrl+C to exit):"
    echo ""
    "${CONTAINER_RUN}" logs -f "${WEAVIATE_CONTAINER_NAME}"
}

clean_weaviate() {
    echo "🧹 Cleaning Weaviate (removing container and storage)..."

    # Stop and remove container
    if "${CONTAINER_RUN}" ps -a --format "{{.Names}}" | grep -q "^${WEAVIATE_CONTAINER_NAME}$"; then
        "${CONTAINER_RUN}" stop "${WEAVIATE_CONTAINER_NAME}" 2>/dev/null || true
        "${CONTAINER_RUN}" rm "${WEAVIATE_CONTAINER_NAME}" 2>/dev/null || true
    fi

    # Ask before removing storage
    if [ -d "$STORAGE_DIR" ]; then
        echo ""
        echo "⚠️  This will delete all Weaviate data in: $STORAGE_DIR"
        read -p "Continue? (y/N) " -n 1 -r
        echo ""
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            rm -rf "$STORAGE_DIR"
            echo "✅ Storage removed"
        else
            echo "ℹ️  Storage kept: $STORAGE_DIR"
        fi
    fi

    echo "✅ Weaviate cleaned"
}

# Main command dispatch
case "${1:-}" in
    start)
        start_weaviate
        ;;
    stop)
        stop_weaviate
        ;;
    status)
        status_weaviate
        ;;
    logs)
        logs_weaviate
        ;;
    clean)
        clean_weaviate
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
