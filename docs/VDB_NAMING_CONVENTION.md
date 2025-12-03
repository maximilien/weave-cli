# Vector Database Naming Convention

**Last Updated:** 2025-12-03
**Version:** 0.7.2

## Overview

Weave CLI uses a consistent naming convention for all vector databases to distinguish between local and cloud deployments.

## Naming Standard

All vector databases use the format: `<vdb-name>-<deployment>`

- **Local deployments:** `<vdb-name>-local`
- **Cloud deployments:** `<vdb-name>-cloud`

## Configured Databases

### Local Deployments

- `weaviate-local` - Weaviate running locally
- `milvus-local` - Milvus standalone or cluster (local)
- `chroma-local` - Chroma running locally
- `qdrant-local` - Qdrant running locally
- `neo4j-local` - Neo4j running locally

### Cloud Deployments

- `weaviate-cloud` - Weaviate Cloud Service (WCS)
- `milvus-cloud` - Zilliz Cloud (managed Milvus)
- `mongodb-cloud` - MongoDB Atlas (cloud-hosted)
- `supabase-cloud` - Supabase PGVector (cloud-hosted)
- `chroma-cloud` - Chroma Cloud
- `qdrant-cloud` - Qdrant Cloud
- `neo4j-cloud` - Neo4j Aura

## Shortcut Resolution

For convenience, bare VDB names automatically resolve to their `-cloud` variants:

| Shortcut   | Resolves To       |
|------------|-------------------|
| `weaviate` | `weaviate-cloud`  |
| `milvus`   | `milvus-cloud`    |
| `mongodb`  | `mongodb-cloud`   |
| `supabase` | `supabase-cloud`  |
| `chroma`   | `chroma-cloud`    |
| `neo4j`    | `neo4j-cloud`     |
| `qdrant`   | `qdrant-cloud`    |

### Examples

```bash
# These are equivalent:
weave health check weaviate
weave health check weaviate-cloud

# These are equivalent:
weave health check milvus
weave health check milvus-cloud

# To use local, specify explicitly:
weave health check weaviate-local
weave health check milvus-local
```

## Configuration

In `config.yaml`, all database configurations use the full `-local` or `-cloud` suffix:

```yaml
databases:
  default: weaviate-cloud

  vector_databases:
    # Cloud deployment
    - name: weaviate-cloud
      type: weaviate-cloud
      url: ${WEAVIATE_URL}
      api_key: ${WEAVIATE_API_KEY}

    # Local deployment
    - name: milvus-local
      type: milvus-local
      address: localhost:19530
```

## CLI Flags

CLI flags remain unchanged and use the provider name with deployment type:

- `--weaviate` - Selects weaviate-cloud or weaviate-local (if configured)
- `--milvus-local` - Selects only milvus-local
- `--milvus-cloud` - Selects only milvus-cloud
- `--mongodb` - Selects mongodb-local
- `--supabase` - Selects supabase-local
- `--chroma-local` - Selects only chroma-local
- `--chroma-cloud` - Selects only chroma-cloud
- `--qdrant-local` - Selects only qdrant-local
- `--qdrant-cloud` - Selects only qdrant-cloud
- `--neo4j-local` - Selects only neo4j-local
- `--neo4j-cloud` - Selects only neo4j-cloud

## Migration Notes

### For Existing Configurations

If you have an existing `config.yaml` with old naming, update:

```yaml
# OLD (before v0.7.2)
- name: milvus
  type: milvus-local

- name: mongodb
  type: mongodb

- name: supabase
  type: supabase

# NEW (v0.7.2+)
- name: milvus-local
  type: milvus-local

- name: mongodb-cloud
  type: mongodb-cloud

- name: supabase-cloud
  type: supabase-cloud
```

### Backward Compatibility

The shortcut resolution provides partial backward compatibility:

- Commands referencing bare names like `weaviate` will work and resolve to `weaviate-cloud`
- Commands referencing full names like `milvus-local` will continue to work
- Old config names (if not updated) may cause "database not found" errors

## Best Practices

1. **Always use full names in config.yaml** - Use `-local` or `-cloud` suffix
2. **Use shortcuts in CLI** - For convenience when referencing cloud deployments
3. **Be explicit when needed** - Use full names when you need to be clear about deployment type
4. **Update documentation** - When creating examples, use the standardized naming

## Related Documentation

- [Configuration Guide](../README.md#configuration)
- [VDB Support Matrix](VDB_SUPPORT.md)
- [Health Check Guide](../README.md#health-checks)
