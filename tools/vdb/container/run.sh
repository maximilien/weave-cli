#!/bin/bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2025 dr.max

# Run container commands using detected runtime (podman or docker)
# Usage: ./run.sh <container-command> [args...]
# Examples:
#   ./run.sh ps -a
#   ./run.sh compose up -d
#   ./run.sh run -d --name milvus ...

set -euo pipefail

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Detect container runtime
RUNTIME=$("${SCRIPT_DIR}/detect.sh")

if [ "$RUNTIME" = "none" ]; then
    echo "❌ No container runtime found. Run detect.sh for installation instructions." >&2
    exit 1
fi

# Execute the command with the detected runtime
exec "$RUNTIME" "$@"
