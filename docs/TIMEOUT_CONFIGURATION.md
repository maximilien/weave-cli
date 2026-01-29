# Timeout Configuration Guide

Weave CLI uses intelligent, operation-specific timeouts to prevent false failures while keeping operations responsive.

## Quick Reference

```bash
# Override timeout for a single command
weave cols query MyDocs "test" --timeout 30s

# Set timeout in config.yaml (per database)
databases:
  vector_databases:
    - name: weaviate-cloud
      type: weaviate-cloud
      timeout: 30  # seconds

# Common timeout values
--timeout 5s     # Fast local operations
--timeout 10s    # Default for most operations
--timeout 30s    # Cloud deployments
--timeout 60s    # Bulk operations
--timeout 300s   # Large batch imports (5 minutes)
```

## How Timeouts Work

### Operation-Specific Defaults

Weave automatically adjusts timeouts based on **operation type** and **deployment type** (local vs cloud):

| Operation Type | Local Default | Cloud Default | Use Case |
|---------------|---------------|---------------|----------|
| **Health Check** | 10s | 20s | Server connectivity test |
| **Document** (single) | 15s | 30s | Create/update/delete one doc |
| **Collection** | 20s | 40s | Create/list collections |
| **Query** (search) | 20s | 40s | Vector similarity search |
| **Schema** | 15s | 30s | Get/update schema |
| **Bulk** (batch) | 120s (2 min) | 300s (5 min) | Large batch operations |

**Cloud operations get 2x timeout** to account for network latency.

**Bulk operations get 5-10x timeout** to handle large batches without false timeouts.

### Priority Order

Timeouts are resolved in this order (highest to lowest priority):

1. **`--timeout` flag** - Command-line override (highest priority)
2. **config.yaml** - Per-database configuration
3. **Smart defaults** - Operation-specific defaults based on VDB type
4. **Fallback** - 30s local, 60s cloud

## Configuration Examples

### Per-Database Timeout (config.yaml)

```yaml
databases:
  vector_databases:
    # Local deployment - shorter timeouts
    - name: weaviate-local
      type: weaviate-local
      url: http://localhost:8080
      timeout: 10  # seconds

    # Cloud deployment - longer timeouts
    - name: weaviate-cloud
      type: weaviate-cloud
      url: https://cluster.weaviate.cloud
      timeout: 30  # seconds (network latency)

    # Bulk operations - very long timeout
    - name: milvus-cloud
      type: milvus-cloud
      address: cluster.zillizcloud.com:19530
      timeout: 120  # 2 minutes for large batches
```

### Command-Line Override

```bash
# Override for a single command
weave docs create MyDocs large-file.pdf --timeout 60s

# Batch operations with custom timeout
weave docs batch --directory ./data --collection Docs --timeout 300s

# Health check with short timeout
weave health check --timeout 5s
```

## Common Scenarios

### Slow Network (Cloud Deployments)

**Problem**: Operations timing out with cloud VDBs.

**Solution**: Increase timeout for cloud databases:

```yaml
# config.yaml
- name: qdrant-cloud
  type: qdrant-cloud
  url: https://cluster.qdrant.cloud
  timeout: 60  # Increase from default 30s
```

### Large Batch Imports

**Problem**: Bulk document imports failing with timeout errors.

**Solution**: Use longer timeout for bulk operations:

```bash
# Import 1000+ documents
weave docs batch --dir ./documents --collection Docs --timeout 600s
```

**Note**: Bulk operations automatically get longer timeouts (120s local, 300s cloud), but you can override if needed.

### Local Development (Fast Responses)

**Problem**: Want faster feedback during development.

**Solution**: Use shorter timeouts for local databases:

```yaml
# config.yaml
- name: chroma-local
  type: chroma-local
  url: http://localhost:8000
  timeout: 5  # Fast fail for quick iteration
```

### Connection Testing

**Problem**: Need to quickly test if VDB is reachable.

**Solution**: Use short timeout for health checks:

```bash
# Quick connectivity test (fail fast)
weave health check --timeout 3s
```

## Timeout Errors

When an operation times out, you'll see an error like:

```
Error: connection timeout [operation=Health vdb=weaviate endpoint=localhost:8080 timeout=10s]

Connection timeout. Common causes:
  1. Weaviate server not responding (check logs: docker logs <container>)
  2. Network/firewall issues blocking port 8080
  3. For Weaviate Cloud: verify cluster is running
  → Check Weaviate status and network connectivity
```

### Troubleshooting Timeout Errors

1. **Verify VDB is running**:
   ```bash
   # Check Docker containers
   docker ps

   # Check logs
   docker logs <container-name>
   ```

2. **Test connectivity**:
   ```bash
   # Test with short timeout
   weave health check --timeout 5s
   ```

3. **Increase timeout if needed**:
   ```bash
   # Try with longer timeout
   weave health check --timeout 30s
   ```

4. **Check network**:
   ```bash
   # For cloud deployments, test connectivity
   curl -I https://your-cluster.weaviate.cloud
   ```

## Advanced Configuration

### Different Timeouts for Different Operations

You can set a base timeout in config.yaml, and override it per command:

```yaml
# config.yaml - conservative default
- name: weaviate-cloud
  timeout: 30
```

```bash
# Quick health check
weave health check --weaviate-cloud --timeout 10s

# Long-running query
weave cols query Docs "complex search" --weaviate-cloud --timeout 60s

# Bulk import
weave docs batch --dir ./data --weaviate-cloud --timeout 300s
```

### Disabling Timeouts

**Not recommended**, but you can set very long timeouts:

```bash
# Effectively disable timeout (10 minutes)
weave docs batch --dir ./large-dataset --timeout 600s
```

## Best Practices

1. **Use smart defaults** - Let Weave handle timeouts automatically
2. **Configure per database** - Set base timeout in config.yaml
3. **Override when needed** - Use `--timeout` flag for special cases
4. **Cloud gets longer timeouts** - Account for network latency (30-60s)
5. **Bulk operations need patience** - Use 120-300s for large batches
6. **Test with short timeouts** - Use `--timeout 5s` for quick connectivity tests
7. **Monitor and adjust** - If operations consistently timeout, increase base timeout

## Related Documentation

- [Configuration Guide](USER_GUIDE.md#configuration)
- [VDB Support](VDB_SUPPORT.md)
- [Troubleshooting](USER_GUIDE.md#troubleshooting)
- [Architecture - Timeout Strategy](ARCHITECTURE.md)

## Examples

### Production Deployment

```yaml
# config.yaml - Production settings
databases:
  vector_databases:
    - name: weaviate-prod
      type: weaviate-cloud
      url: ${WEAVIATE_URL}
      api_key: ${WEAVIATE_API_KEY}
      timeout: 60  # Conservative for production

    - name: milvus-prod
      type: milvus-cloud
      address: ${MILVUS_ADDRESS}
      timeout: 90  # Account for network + processing
```

### Development Environment

```yaml
# config.yaml - Development settings
databases:
  vector_databases:
    - name: weaviate-dev
      type: weaviate-local
      url: http://localhost:8080
      timeout: 5  # Fast fail for quick iteration

    - name: chroma-dev
      type: chroma-local
      url: http://localhost:8000
      timeout: 5  # Fast feedback
```

### Data Migration

```bash
# Large dataset migration with generous timeout
weave docs batch \
  --directory ./archive/2024 \
  --collection Archive \
  --timeout 1800s \  # 30 minutes
  --parallel 5 \
  --retry 3
```

---

**Need help?**
- Run `weave --help` to see all timeout options
- Check `weave config show` to see current timeouts
- See [User Guide](USER_GUIDE.md) for more examples
