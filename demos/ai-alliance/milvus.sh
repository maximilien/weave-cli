#!/bin/bash
# Delegates to shared milvus management script
exec "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/../shared/milvus.sh" "$@"
