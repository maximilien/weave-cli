#!/bin/bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2025 dr.max

# PostgreSQL + pgvector local management script
# For local Supabase-compatible vector database
# Commands: start, stop, status, logs, clean
# Uses podman or docker via container abstraction

set -euo pipefail

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../../.." && pwd)"
CONTAINER_RUN="${SCRIPT_DIR}/../container/run.sh"
STORAGE_DIR="${PROJECT_ROOT}/local/storage/postgres_storage"

# PostgreSQL container configuration
POSTGRES_CONTAINER_NAME="postgres-pgvector"
POSTGRES_IMAGE="ankane/pgvector:latest"
POSTGRES_PORT="5432"
POSTGRES_USER="postgres"
POSTGRES_PASSWORD="postgres"
POSTGRES_DB="weave"

usage() {
    cat <<EOF
Usage: $(basename "$0") <command>

Commands:
    start       Start PostgreSQL + pgvector container
    stop        Stop PostgreSQL container
    status      Show PostgreSQL container status
    logs        Show PostgreSQL container logs
    clean       Stop and remove PostgreSQL container and volumes
    help        Show this help message

Environment:
    Uses podman or docker (auto-detected, podman preferred)
    PostgreSQL listens on:
        - Port: localhost:${POSTGRES_PORT}
    Connection string:
        postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:${POSTGRES_PORT}/${POSTGRES_DB}
    Storage: ${STORAGE_DIR}

Examples:
    ./postgres.sh start
    ./postgres.sh status
    ./postgres.sh logs
    ./postgres.sh stop
EOF
}

ensure_storage_dir() {
    if [ ! -d "$STORAGE_DIR" ]; then
        echo "📁 Creating PostgreSQL storage directory: $STORAGE_DIR"
        mkdir -p "$STORAGE_DIR"
    fi
}

start_postgres() {
    echo "🚀 Starting PostgreSQL + pgvector..."

    # Check if already running
    if "${CONTAINER_RUN}" ps --format "{{.Names}}" | grep -q "^${POSTGRES_CONTAINER_NAME}$"; then
        echo "⚠️  PostgreSQL is already running"
        echo ""
        status_postgres
        return 0
    fi

    ensure_storage_dir

    # Start PostgreSQL container
    "${CONTAINER_RUN}" run -d \
        --name "${POSTGRES_CONTAINER_NAME}" \
        -p "${POSTGRES_PORT}:5432" \
        -e "POSTGRES_USER=${POSTGRES_USER}" \
        -e "POSTGRES_PASSWORD=${POSTGRES_PASSWORD}" \
        -e "POSTGRES_DB=${POSTGRES_DB}" \
        -v "${STORAGE_DIR}:/var/lib/postgresql/data:z" \
        "${POSTGRES_IMAGE}"

    # Wait for PostgreSQL to be ready
    echo "⏳ Waiting for PostgreSQL to be ready..."
    sleep 3

    # Check health
    MAX_RETRIES=15
    RETRY_COUNT=0
    while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
        if "${CONTAINER_RUN}" exec "${POSTGRES_CONTAINER_NAME}" pg_isready -U "${POSTGRES_USER}" > /dev/null 2>&1; then
            echo "✅ PostgreSQL started successfully"
            echo ""
            echo "Connection details:"
            echo "  Host: localhost"
            echo "  Port: ${POSTGRES_PORT}"
            echo "  Database: ${POSTGRES_DB}"
            echo "  User: ${POSTGRES_USER}"
            echo "  Password: ${POSTGRES_PASSWORD}"
            echo ""
            echo "  Connection string:"
            echo "  postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:${POSTGRES_PORT}/${POSTGRES_DB}"
            echo ""
            echo "  pgvector extension: ✅ Enabled"
            echo "  Storage: ${STORAGE_DIR}"
            echo ""
            echo "Check status with: $(basename "$0") status"
            echo "View logs with: $(basename "$0") logs"
            echo ""
            echo "Test with: weave health check --supabase-local"
            return 0
        fi
        RETRY_COUNT=$((RETRY_COUNT + 1))
        sleep 1
    done

    echo "⚠️  PostgreSQL started but health check timed out"
    echo "   Check logs with: $(basename "$0") logs"
}

stop_postgres() {
    echo "🛑 Stopping PostgreSQL..."

    if ! "${CONTAINER_RUN}" ps --format "{{.Names}}" | grep -q "^${POSTGRES_CONTAINER_NAME}$"; then
        echo "⚠️  PostgreSQL is not running"
        return 0
    fi

    "${CONTAINER_RUN}" stop "${POSTGRES_CONTAINER_NAME}"
    "${CONTAINER_RUN}" rm "${POSTGRES_CONTAINER_NAME}"

    echo "✅ PostgreSQL stopped"
}

status_postgres() {
    # Check if container exists and is running
    if "${CONTAINER_RUN}" ps --format "{{.Names}}" | grep -q "^${POSTGRES_CONTAINER_NAME}$"; then
        echo "✅ PostgreSQL is running"
        echo ""
        "${CONTAINER_RUN}" ps --filter "name=${POSTGRES_CONTAINER_NAME}" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
        echo ""

        # Try to get health status
        if "${CONTAINER_RUN}" exec "${POSTGRES_CONTAINER_NAME}" pg_isready -U "${POSTGRES_USER}" > /dev/null 2>&1; then
            echo "🟢 Health check: OK"

            # Get version info
            VERSION=$("${CONTAINER_RUN}" exec "${POSTGRES_CONTAINER_NAME}" psql -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -t -c "SELECT version();" 2>/dev/null | head -1 | tr -d ' ' || echo "unknown")
            echo "📦 PostgreSQL: $VERSION"

            # Check pgvector extension
            PGVECTOR=$("${CONTAINER_RUN}" exec "${POSTGRES_CONTAINER_NAME}" psql -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -t -c "SELECT extversion FROM pg_extension WHERE extname='vector';" 2>/dev/null | tr -d ' ' || echo "not installed")
            if [ "$PGVECTOR" != "not installed" ] && [ -n "$PGVECTOR" ]; then
                echo "📦 pgvector: $PGVECTOR"
            else
                echo "⚠️  pgvector: Not installed (will be auto-created when needed)"
            fi
        else
            echo "🔴 Health check: FAILED"
        fi
    else
        echo "❌ PostgreSQL is not running"
        echo ""
        echo "Start with: $(basename "$0") start"
        return 1
    fi
}

logs_postgres() {
    if ! "${CONTAINER_RUN}" ps --format "{{.Names}}" | grep -q "^${POSTGRES_CONTAINER_NAME}$"; then
        echo "❌ PostgreSQL is not running" >&2
        exit 1
    fi

    echo "📋 PostgreSQL logs (Ctrl+C to exit):"
    echo ""
    "${CONTAINER_RUN}" logs -f "${POSTGRES_CONTAINER_NAME}"
}

clean_postgres() {
    echo "🧹 Cleaning PostgreSQL (removing container and storage)..."

    # Stop and remove container
    if "${CONTAINER_RUN}" ps -a --format "{{.Names}}" | grep -q "^${POSTGRES_CONTAINER_NAME}$"; then
        "${CONTAINER_RUN}" stop "${POSTGRES_CONTAINER_NAME}" 2>/dev/null || true
        "${CONTAINER_RUN}" rm "${POSTGRES_CONTAINER_NAME}" 2>/dev/null || true
    fi

    # Ask before removing storage
    if [ -d "$STORAGE_DIR" ]; then
        echo ""
        echo "⚠️  This will delete all PostgreSQL data in: $STORAGE_DIR"
        read -p "Continue? (y/N) " -n 1 -r
        echo ""
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            rm -rf "$STORAGE_DIR"
            echo "✅ Storage removed"
        else
            echo "ℹ️  Storage kept: $STORAGE_DIR"
        fi
    fi

    echo "✅ PostgreSQL cleaned"
}

# Main command dispatch
case "${1:-}" in
    start)
        start_postgres
        ;;
    stop)
        stop_postgres
        ;;
    status)
        status_postgres
        ;;
    logs)
        logs_postgres
        ;;
    clean)
        clean_postgres
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
