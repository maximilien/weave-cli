#!/bin/bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2025 dr.max

# Milvus local management script
# Commands: start, stop, status, logs, clean
# Uses podman or docker via container abstraction

set -euo pipefail

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../../.." && pwd)"
CONTAINER_RUN="${SCRIPT_DIR}/../container/run.sh"
MILVUS_DIR="${PROJECT_ROOT}/local/milvus"

# Milvus container configuration
MILVUS_CONTAINER_NAME="milvus-standalone"
MILVUS_PORT="19530"
MILVUS_HTTP_PORT="9091"

usage() {
    cat <<EOF
Usage: $(basename "$0") <command>

Commands:
    start       Start Milvus standalone container
    stop        Stop Milvus standalone container
    status      Show Milvus container status
    logs        Show Milvus container logs
    clean       Stop and remove Milvus container and volumes
    help        Show this help message

Environment:
    Uses podman or docker (auto-detected, podman preferred)
    Milvus listens on:
        - gRPC: localhost:${MILVUS_PORT}
        - HTTP: localhost:${MILVUS_HTTP_PORT}

Examples:
    ./milvus.sh start
    ./milvus.sh status
    ./milvus.sh logs
    ./milvus.sh stop
EOF
}

check_compose_dir() {
    if [ ! -d "$MILVUS_DIR" ]; then
        echo "❌ Milvus compose directory not found: $MILVUS_DIR" >&2
        echo "" >&2
        echo "Expected directory structure:" >&2
        echo "  local/milvus/docker-compose.yml" >&2
        echo "  local/milvus/podman-compose.yml" >&2
        exit 1
    fi
}

start_milvus() {
    echo "🚀 Starting Milvus standalone..."
    check_compose_dir

    cd "$MILVUS_DIR"

    # Detect runtime and use appropriate compose file
    RUNTIME=$("${SCRIPT_DIR}/../container/detect.sh")

    if [ "$RUNTIME" = "podman" ]; then
        if [ -f "podman-compose.yml" ]; then
            # Use DOCKER_CONFIG=/dev/null to avoid docker-credential-desktop
            # errors when using podman
            DOCKER_CONFIG=/dev/null "${CONTAINER_RUN}" compose \
                -f podman-compose.yml up -d
        else
            DOCKER_CONFIG=/dev/null "${CONTAINER_RUN}" compose \
                -f docker-compose.yml up -d
        fi
    else
        "${CONTAINER_RUN}" compose -f docker-compose.yml up -d
    fi

    echo "✅ Milvus started successfully"
    echo ""
    echo "Connection details:"
    echo "  gRPC endpoint: localhost:${MILVUS_PORT}"
    echo "  HTTP endpoint: localhost:${MILVUS_HTTP_PORT}"
    echo ""
    echo "Check status with: $(basename "$0") status"
}

stop_milvus() {
    echo "🛑 Stopping Milvus standalone..."
    check_compose_dir

    cd "$MILVUS_DIR"

    # Detect runtime and use appropriate compose file
    RUNTIME=$("${SCRIPT_DIR}/../container/detect.sh")

    if [ "$RUNTIME" = "podman" ]; then
        if [ -f "podman-compose.yml" ]; then
            DOCKER_CONFIG=/dev/null "${CONTAINER_RUN}" compose \
                -f podman-compose.yml down
        else
            "${CONTAINER_RUN}" compose -f docker-compose.yml down
        fi
    else
        "${CONTAINER_RUN}" compose -f docker-compose.yml down
    fi

    echo "✅ Milvus stopped"
}

status_milvus() {
    check_compose_dir

    # Check if container is running
    if "${CONTAINER_RUN}" ps --format "{{.Names}}" | grep -q "^${MILVUS_CONTAINER_NAME}$"; then
        echo "✅ Milvus is running"
        echo ""
        "${CONTAINER_RUN}" ps --filter "name=${MILVUS_CONTAINER_NAME}" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
    else
        echo "❌ Milvus is not running"
        echo ""
        echo "Start with: $(basename "$0") start"
        return 1
    fi
}

logs_milvus() {
    check_compose_dir

    if ! "${CONTAINER_RUN}" ps --format "{{.Names}}" | grep -q "^${MILVUS_CONTAINER_NAME}$"; then
        echo "❌ Milvus is not running" >&2
        exit 1
    fi

    echo "📋 Milvus logs (Ctrl+C to exit):"
    echo ""
    "${CONTAINER_RUN}" logs -f "${MILVUS_CONTAINER_NAME}"
}

clean_milvus() {
    echo "🧹 Cleaning Milvus (removing container and volumes)..."
    check_compose_dir

    cd "$MILVUS_DIR"

    # Detect runtime and use appropriate compose file
    RUNTIME=$("${SCRIPT_DIR}/../container/detect.sh")

    if [ "$RUNTIME" = "podman" ]; then
        if [ -f "podman-compose.yml" ]; then
            "${CONTAINER_RUN}" compose -f podman-compose.yml down -v
        else
            "${CONTAINER_RUN}" compose -f docker-compose.yml down -v
        fi
    else
        "${CONTAINER_RUN}" compose -f docker-compose.yml down -v
    fi

    echo "✅ Milvus cleaned"
}

# Main command dispatch
case "${1:-}" in
    start)
        start_milvus
        ;;
    stop)
        stop_milvus
        ;;
    status)
        status_milvus
        ;;
    logs)
        logs_milvus
        ;;
    clean)
        clean_milvus
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
