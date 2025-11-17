# Milvus Standalone - Local Development

This directory contains Docker and Podman compose files for running Milvus
standalone locally for development and testing.

## Quick Start

### Using Management Script (Recommended)

```bash
# Start Milvus
./tools/vdb/local/milvus.sh start

# Check status
./tools/vdb/local/milvus.sh status

# View logs
./tools/vdb/local/milvus.sh logs

# Stop Milvus
./tools/vdb/local/milvus.sh stop

# Clean up (removes volumes)
./tools/vdb/local/milvus.sh clean
```

### Using Container Commands Directly

#### Podman (Preferred)

```bash
cd local/milvus
podman-compose -f podman-compose.yml up -d
podman-compose -f podman-compose.yml down
```

#### Docker

```bash
cd local/milvus
docker compose -f docker-compose.yml up -d
docker compose -f docker-compose.yml down
```

## Connection Details

Once running, Milvus is accessible at:

- **gRPC endpoint**: `localhost:19530`
- **HTTP endpoint**: `localhost:9091`

## Configuration for weave-cli

Add to your `.env`:

```bash
# Milvus Local
MILVUS_LOCAL_ADDRESS=localhost:19530
MILVUS_LOCAL_DATABASE=default
```

Test connection:

```bash
weave health check --milvus-local
```

## Components

The Milvus standalone deployment includes three containers:

1. **milvus-standalone** - Main Milvus service (ports 19530, 9091)
2. **milvus-etcd** - Configuration and metadata storage
3. **milvus-minio** - Object storage for vector data

## Data Persistence

Data is stored in local volumes under `./volumes/`:

- `./volumes/milvus/` - Vector data and indexes
- `./volumes/etcd/` - Metadata
- `./volumes/minio/` - Object storage

To completely reset Milvus:

```bash
./tools/vdb/local/milvus.sh clean
rm -rf local/milvus/volumes/
```

## Podman vs Docker

### Podman-Specific Notes

- Uses `:Z` flag for SELinux labeling on volumes
- Doesn't require `seccomp:unconfined` in most cases
- Rootless by default (more secure)

### Docker-Specific Notes

- Uses `seccomp:unconfined` for compatibility
- May require Docker daemon to be running

The management script automatically detects which runtime is available and uses
the appropriate compose file.

## Troubleshooting

### Port Already in Use

If port 19530 or 9091 is already in use:

```bash
# Check what's using the port
lsof -i :19530
lsof -i :9091

# Or modify the port in compose file
```

### Permission Issues (Podman)

If you encounter permission errors with podman:

1. Uncomment `security_opt: - label=disable` in `podman-compose.yml`
2. Or run with `--privileged` flag

### Health Check Failures

Wait 90 seconds after startup for all services to be healthy:

```bash
# Check individual container health
podman ps
# or
docker ps
```

## References

- [Milvus Documentation](https://milvus.io/docs)
- [Milvus Standalone Installation](https://milvus.io/docs/install_standalone-docker.md)
- [Milvus Go SDK](https://github.com/milvus-io/milvus-sdk-go)
