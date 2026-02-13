#!/bin/bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2025 dr.max

# Start MinIO for Weave CLI external storage
# Provides S3-compatible object storage for images exceeding VDB limits (e.g., Milvus 65KB)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "🚀 Starting MinIO for Weave CLI External Storage..."
echo ""

# Check if docker/podman is available
if command -v docker &> /dev/null; then
    DOCKER_CMD="docker"
    COMPOSE_CMD="docker-compose"
elif command -v podman &> /dev/null; then
    DOCKER_CMD="podman"
    COMPOSE_CMD="podman-compose"
else
    echo "❌ Error: Neither docker nor podman is installed"
    exit 1
fi

echo "Using: $DOCKER_CMD"
echo ""

# Start MinIO
cd "$PROJECT_ROOT"
$COMPOSE_CMD -f docker/docker-compose.minio.yml up -d

# Wait for MinIO to be healthy
echo ""
echo "⏳ Waiting for MinIO to be ready..."
sleep 5

# Check if MinIO is running
if $DOCKER_CMD ps | grep -q weave-minio; then
    echo ""
    echo "✅ MinIO Started Successfully!"
    echo ""
    echo "📊 Connection Details:"
    echo "   API Endpoint:  http://localhost:9000"
    echo "   Console:       http://localhost:9001"
    echo "   Username:      minioadmin"
    echo "   Password:      minioadmin"
    echo ""
    echo "📦 Default Buckets:"
    echo "   - weave-images (for image storage)"
    echo "   - weave-pdfs   (for PDF storage)"
    echo ""
    echo "💡 Usage Example:"
    echo "   weave docs create MyCollection images/ \\"
    echo "     --image-storage minio \\"
    echo "     --minio-bucket weave-images \\"
    echo "     --milvus-local"
    echo ""
    echo "🛑 To stop MinIO:"
    echo "   $COMPOSE_CMD -f docker/docker-compose.minio.yml down"
    echo ""
else
    echo "❌ Error: MinIO failed to start"
    echo ""
    echo "Check logs with:"
    echo "  $DOCKER_CMD logs weave-minio"
    exit 1
fi
