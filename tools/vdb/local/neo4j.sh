#!/bin/bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2025 dr.max

# Neo4j local management script
# Commands: start, stop, status, logs, clean
# Uses podman or docker via container abstraction

set -euo pipefail

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../../.." && pwd)"
CONTAINER_RUN="${SCRIPT_DIR}/../container/run.sh"
STORAGE_DIR="${PROJECT_ROOT}/local/storage/neo4j_storage"

# Neo4j container configuration
NEO4J_CONTAINER_NAME="neo4j"
NEO4J_IMAGE="neo4j:latest"
NEO4J_HTTP_PORT="7474"
NEO4J_BOLT_PORT="7687"
NEO4J_PASSWORD="${NEO4J_PASSWORD:-weaveneo4j}"

usage() {
    cat <<EOF
Usage: $(basename "$0") <command>

Commands:
    start       Start Neo4j container
    stop        Stop Neo4j container
    status      Show Neo4j container status
    logs        Show Neo4j container logs
    clean       Stop and remove Neo4j container and volumes
    help        Show this help message

Environment:
    Uses podman or docker (auto-detected, podman preferred)
    Neo4j listens on:
        - Browser UI: http://localhost:${NEO4J_HTTP_PORT}
        - Bolt: bolt://localhost:${NEO4J_BOLT_PORT}
    Storage: ${STORAGE_DIR}
    Default credentials:
        - Username: neo4j
        - Password: ${NEO4J_PASSWORD} (set via NEO4J_PASSWORD env var)

Examples:
    ./neo4j.sh start
    ./neo4j.sh status
    ./neo4j.sh logs
    ./neo4j.sh stop

    # Set custom password
    NEO4J_PASSWORD=mypassword ./neo4j.sh start
EOF
}

ensure_storage_dirs() {
    for dir in "${STORAGE_DIR}/data" "${STORAGE_DIR}/logs" "${STORAGE_DIR}/import" "${STORAGE_DIR}/plugins"; do
        if [ ! -d "$dir" ]; then
            echo "📁 Creating Neo4j storage directory: $dir"
            mkdir -p "$dir"
        fi
    done
}

start_neo4j() {
    echo "🚀 Starting Neo4j..."

    # Check if already running
    if "${CONTAINER_RUN}" ps --format "{{.Names}}" | grep -q "^${NEO4J_CONTAINER_NAME}$"; then
        echo "⚠️  Neo4j is already running"
        echo ""
        status_neo4j
        return 0
    fi

    ensure_storage_dirs

    # Start Neo4j container
    "${CONTAINER_RUN}" run -d \
        --name "${NEO4J_CONTAINER_NAME}" \
        -p "${NEO4J_HTTP_PORT}:7474" \
        -p "${NEO4J_BOLT_PORT}:7687" \
        -e "NEO4J_AUTH=neo4j/${NEO4J_PASSWORD}" \
        -e "NEO4J_PLUGINS=[\"apoc\"]" \
        -v "${STORAGE_DIR}/data:/data:z" \
        -v "${STORAGE_DIR}/logs:/logs:z" \
        -v "${STORAGE_DIR}/import:/var/lib/neo4j/import:z" \
        -v "${STORAGE_DIR}/plugins:/plugins:z" \
        "${NEO4J_IMAGE}"

    # Wait for Neo4j to be ready
    echo "⏳ Waiting for Neo4j to be ready (this may take 30-60 seconds)..."
    sleep 5

    # Check health
    MAX_RETRIES=60
    RETRY_COUNT=0
    while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
        if curl -s "http://localhost:${NEO4J_HTTP_PORT}" > /dev/null 2>&1; then
            echo "✅ Neo4j started successfully"
            echo ""
            echo "Connection details:"
            echo "  Browser UI: http://localhost:${NEO4J_HTTP_PORT}"
            echo "  Bolt URL: bolt://localhost:${NEO4J_BOLT_PORT}"
            echo "  Username: neo4j"
            echo "  Password: ${NEO4J_PASSWORD}"
            echo "  Storage: ${STORAGE_DIR}"
            echo ""
            echo "Environment variables for weave-cli:"
            echo "  export NEO4J_URL=\"bolt://localhost:${NEO4J_BOLT_PORT}\""
            echo "  export NEO4J_USERNAME=\"neo4j\""
            echo "  export NEO4J_PASSWORD=\"${NEO4J_PASSWORD}\""
            echo ""
            echo "Check status with: $(basename "$0") status"
            echo "View logs with: $(basename "$0") logs"
            return 0
        fi
        RETRY_COUNT=$((RETRY_COUNT + 1))
        sleep 1
    done

    echo "⚠️  Neo4j started but health check timed out"
    echo "   Check logs with: $(basename "$0") logs"
}

stop_neo4j() {
    echo "🛑 Stopping Neo4j..."

    if ! "${CONTAINER_RUN}" ps --format "{{.Names}}" | grep -q "^${NEO4J_CONTAINER_NAME}$"; then
        echo "⚠️  Neo4j is not running"
        return 0
    fi

    "${CONTAINER_RUN}" stop "${NEO4J_CONTAINER_NAME}"
    "${CONTAINER_RUN}" rm "${NEO4J_CONTAINER_NAME}"

    echo "✅ Neo4j stopped"
}

status_neo4j() {
    # Check if container exists and is running
    if "${CONTAINER_RUN}" ps --format "{{.Names}}" | grep -q "^${NEO4J_CONTAINER_NAME}$"; then
        echo "✅ Neo4j is running"
        echo ""
        "${CONTAINER_RUN}" ps --filter "name=${NEO4J_CONTAINER_NAME}" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
        echo ""

        # Try to get health status
        if curl -s "http://localhost:${NEO4J_HTTP_PORT}" > /dev/null 2>&1; then
            echo "🟢 Health check: OK"
            echo ""
            echo "Access Neo4j:"
            echo "  Browser UI: http://localhost:${NEO4J_HTTP_PORT}"
            echo "  Bolt URL: bolt://localhost:${NEO4J_BOLT_PORT}"
        else
            echo "🔴 Health check: FAILED (may still be starting up)"
        fi
    else
        echo "❌ Neo4j is not running"
        echo ""
        echo "Start with: $(basename "$0") start"
        return 1
    fi
}

logs_neo4j() {
    if ! "${CONTAINER_RUN}" ps --format "{{.Names}}" | grep -q "^${NEO4J_CONTAINER_NAME}$"; then
        echo "❌ Neo4j is not running" >&2
        exit 1
    fi

    echo "📋 Neo4j logs (Ctrl+C to exit):"
    echo ""
    "${CONTAINER_RUN}" logs -f "${NEO4J_CONTAINER_NAME}"
}

clean_neo4j() {
    echo "🧹 Cleaning Neo4j (removing container and storage)..."

    # Stop and remove container
    if "${CONTAINER_RUN}" ps -a --format "{{.Names}}" | grep -q "^${NEO4J_CONTAINER_NAME}$"; then
        "${CONTAINER_RUN}" stop "${NEO4J_CONTAINER_NAME}" 2>/dev/null || true
        "${CONTAINER_RUN}" rm "${NEO4J_CONTAINER_NAME}" 2>/dev/null || true
    fi

    # Ask before removing storage
    if [ -d "$STORAGE_DIR" ]; then
        echo ""
        echo "⚠️  This will delete all Neo4j data in: $STORAGE_DIR"
        read -p "Continue? (y/N) " -n 1 -r
        echo ""
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            rm -rf "$STORAGE_DIR"
            echo "✅ Storage removed"
        else
            echo "ℹ️  Storage kept: $STORAGE_DIR"
        fi
    fi

    echo "✅ Neo4j cleaned"
}

# Main command dispatch
case "${1:-}" in
    start)
        start_neo4j
        ;;
    stop)
        stop_neo4j
        ;;
    status)
        status_neo4j
        ;;
    logs)
        logs_neo4j
        ;;
    clean)
        clean_neo4j
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
