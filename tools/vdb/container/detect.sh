#!/bin/bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2025 dr.max

# Detect available container runtime (podman preferred over docker)
# Returns: podman, docker, or none

set -euo pipefail

detect_container_runtime() {
    # Check for podman first (preferred)
    if command -v podman &> /dev/null; then
        echo "podman"
        return 0
    fi

    # Fall back to docker
    if command -v docker &> /dev/null; then
        echo "docker"
        return 0
    fi

    # No container runtime found
    echo "none"
    return 1
}

# Main execution
RUNTIME=$(detect_container_runtime)

if [ "$RUNTIME" = "none" ]; then
    echo "❌ No container runtime found. Please install podman or docker." >&2
    echo "" >&2
    echo "Install podman (recommended):" >&2
    echo "  macOS:   brew install podman" >&2
    echo "  Linux:   sudo apt install podman   (Debian/Ubuntu)" >&2
    echo "           sudo dnf install podman   (Fedora/RHEL)" >&2
    echo "" >&2
    echo "Or install docker:" >&2
    echo "  macOS:   brew install --cask docker" >&2
    echo "  Linux:   https://docs.docker.com/engine/install/" >&2
    exit 1
fi

echo "$RUNTIME"
