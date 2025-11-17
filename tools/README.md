# Tools Directory

Utility scripts for development, testing, and demo recording.

## Directory Structure

```text
tools/
├── vdb/                          # Vector database management
│   ├── local/                    # Local VDB management
│   │   ├── milvus.sh            # Milvus start/stop/status
│   │   └── manager.sh           # Generic VDB local manager
│   ├── container/                # Container runtime management
│   │   ├── detect.sh            # Detect podman/docker
│   │   └── run.sh               # Container command abstraction
│   └── health.sh                # Health check utility
├── demo/                         # Demo recording and playback
│   ├── record.sh                # Record demos (asciinema)
│   ├── record-all-demos.sh      # Record all demos
│   ├── auto-demo-recorder.exp   # Expect script for automation
│   └── README-DEMO-RECORDING.md # Demo recording guide
├── dev/                          # Development utilities
│   ├── add_license_headers.sh   # Add MIT license headers
│   ├── fix_markdown_lint.sh     # Fix markdown linting issues
│   └── fix-markdown-code-blocks.sh # Fix markdown code blocks
└── README.md                     # This file
```

## Vector Database Tools (`vdb/`)

### Local VDB Management

Start, stop, and manage local vector database instances using Docker or Podman.

```bash
# Start Milvus locally
./tools/vdb/local/milvus.sh start

# Check status
./tools/vdb/local/milvus.sh status

# Stop Milvus
./tools/vdb/local/milvus.sh stop

# View logs
./tools/vdb/local/milvus.sh logs
```

### Container Management

Abstraction layer for Docker/Podman operations.

```bash
# Detect available container runtime
./tools/vdb/container/detect.sh

# Run container command (auto-detects runtime)
./tools/vdb/container/run.sh ps
```

### Health Checks

```bash
# Check health of all configured VDBs
./tools/vdb/health.sh
```

## Demo Tools (`demo/`)

Record and manage demonstration videos using asciinema.

```bash
# Record a single demo
./tools/demo/record.sh demos/quick-demo.sh

# Record all demos
./tools/demo/record-all-demos.sh
```

See [Demo Recording Guide](demo/README-DEMO-RECORDING.md) for details.

## Development Tools (`dev/`)

### License Headers

Add MIT license headers to source files:

```bash
./tools/dev/add_license_headers.sh
```

### Markdown Linting

Fix markdown linting issues:

```bash
./tools/dev/fix_markdown_lint.sh
./tools/dev/fix-markdown-code-blocks.sh
```

## Adding New Tools

When adding new tools:

1. Choose the appropriate subdirectory (`vdb/`, `demo/`, or `dev/`)
2. Make the script executable: `chmod +x tools/category/script.sh`
3. Add documentation to this README
4. Include usage examples

## Quick Links

- **VDB Local Setup**: See `local/` directory in project root
- **Testing**: See `test.sh` in project root
- **CI/CD**: See `.github/workflows/` directory
