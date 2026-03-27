#!/bin/bash
# Milvus local management for demos
# Commands: start, stop, status, clean

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MILVUS_DIR="${SCRIPT_DIR}/local/milvus"
CONTAINER_NAME="milvus-standalone"

# Auto-detect container runtime and compose command
if command -v podman &>/dev/null; then
    RUNTIME="podman"
    COMPOSE_FILE="podman-compose.yml"
    if command -v podman-compose &>/dev/null; then
        COMPOSE_CMD="podman-compose"
    else
        echo "Error: podman found but podman-compose is not installed" >&2
        echo "Install with: brew install podman-compose" >&2
        exit 1
    fi
elif command -v docker &>/dev/null; then
    RUNTIME="docker"
    COMPOSE_FILE="docker-compose.yml"
    COMPOSE_CMD="docker compose"
else
    echo "Error: neither podman nor docker found" >&2
    exit 1
fi

case "${1:-help}" in
    start)
        echo "Starting Milvus standalone ($RUNTIME)..."
        cd "$MILVUS_DIR"
        $COMPOSE_CMD -f "$COMPOSE_FILE" up -d
        echo ""
        echo "Milvus ready at localhost:19530 (gRPC) / localhost:9091 (HTTP)"
        ;;
    stop)
        echo "Stopping Milvus..."
        cd "$MILVUS_DIR"
        $COMPOSE_CMD -f "$COMPOSE_FILE" down
        ;;
    status)
        if $RUNTIME ps --format "{{.Names}}" 2>/dev/null | grep -q "^${CONTAINER_NAME}$"; then
            echo "Milvus is running"
            $RUNTIME ps --filter "name=${CONTAINER_NAME}"
        else
            echo "Milvus is not running"
        fi
        ;;
    clean)
        echo "Removing Milvus containers and volumes..."
        cd "$MILVUS_DIR"
        $COMPOSE_CMD -f "$COMPOSE_FILE" down -v
        ;;
    *)
        echo "Usage: $0 {start|stop|status|clean}"
        ;;
esac
