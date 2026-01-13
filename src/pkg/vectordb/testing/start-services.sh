#!/usr/bin/env bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2025 dr.max

# Start vector database services for integration testing
# Supports both Docker and Podman

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPOSE_FILE="${SCRIPT_DIR}/docker-compose.test.yml"

# Detect container runtime
if command -v podman &> /dev/null && [ "${USE_PODMAN}" = "1" ]; then
    RUNTIME="podman"
    COMPOSE_CMD="podman-compose"
elif command -v docker &> /dev/null; then
    RUNTIME="docker"
    COMPOSE_CMD="docker-compose"
else
    echo "Error: Neither Docker nor Podman found. Please install one of them."
    exit 1
fi

echo "Using container runtime: ${RUNTIME}"

# Parse arguments
TIER="${1:-1}"
ACTION="${2:-up}"

case "${TIER}" in
    1|tier1)
        SERVICES="qdrant chroma mongodb"
        echo "Starting Tier 1 services: ${SERVICES}"
        ;;
    2|tier2)
        SERVICES="qdrant chroma mongodb milvus-etcd milvus-minio milvus weaviate neo4j"
        echo "Starting Tier 1 + 2 services"
        ;;
    3|tier3|all)
        SERVICES=""  # Empty means all services
        echo "Starting all services"
        ;;
    *)
        echo "Usage: $0 [tier] [action]"
        echo "  tier:   1 (default), 2, 3/all"
        echo "  action: up (default), down, restart, logs"
        echo ""
        echo "Environment:"
        echo "  USE_PODMAN=1  Force use of Podman instead of Docker"
        exit 1
        ;;
esac

# Execute action
case "${ACTION}" in
    up)
        ${COMPOSE_CMD} -f "${COMPOSE_FILE}" up -d ${SERVICES}
        echo ""
        echo "Services starting... waiting for health checks..."
        sleep 10
        ${COMPOSE_CMD} -f "${COMPOSE_FILE}" ps ${SERVICES}
        ;;
    down)
        ${COMPOSE_CMD} -f "${COMPOSE_FILE}" down
        ;;
    restart)
        ${COMPOSE_CMD} -f "${COMPOSE_FILE}" restart ${SERVICES}
        ;;
    logs)
        ${COMPOSE_CMD} -f "${COMPOSE_FILE}" logs -f ${SERVICES}
        ;;
    ps|status)
        ${COMPOSE_CMD} -f "${COMPOSE_FILE}" ps ${SERVICES}
        ;;
    *)
        echo "Unknown action: ${ACTION}"
        exit 1
        ;;
esac
