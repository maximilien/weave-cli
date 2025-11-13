# Vector Database Support Matrix

This document tracks feature support and compatibility across different vector database implementations in weave-cli.

## Supported Vector Databases

| Database | Type | Status | Config Type |
|----------|------|--------|-------------|
| Weaviate Cloud | Cloud | ✅ Production | `weaviate-cloud` |
| Weaviate Local | Self-hosted | ✅ Production | `weaviate-local` |
| Supabase | Cloud | ✅ Production | `supabase` |
| Mock | Testing | ✅ Production | `mock` |

## Feature Support Matrix

### Core Operations

| Feature | Weaviate Cloud | Weaviate Local | Supabase | Mock | Notes |
|---------|----------------|----------------|----------|------|-------|
| Health Check | ✅ | ✅ | ✅ | ✅ | - |
| List Collections | ✅ | ✅ | ✅ | ✅ | - |
| Create Collection | ✅ | ✅ | ✅ | ✅ | - |
| Delete Collection | ✅ | ✅ | ✅ | ✅ | - |
| Collection Exists | ✅ | ✅ | ✅ | ✅ | - |
| Get Collection Count | ✅ | ✅ | ✅ | ✅ | - |
| Get Schema | ✅ | ✅ | ✅ | ✅ | - |
| Validate Schema | ✅ | ✅ | ✅ | ✅ | - |

### Document Operations

| Feature | Weaviate Cloud | Weaviate Local | Supabase | Mock | Notes |
|---------|----------------|----------------|----------|------|-------|
| Create Document | ✅ | ✅ | ✅ | ✅ | - |
| Get Document | ✅ | ✅ | ✅ | ✅ | - |
| Update Document | ✅ | ✅ | ✅ | ✅ | - |
| Delete Document | ✅ | ✅ | ✅ | ✅ | - |
| List Documents | ✅ | ✅ | ✅ | ✅ | - |
| Batch Create | ✅ | ✅ | ✅ | ✅ | - |
| Delete by Metadata | ✅ | ✅ | ✅ | ✅ | - |

### Search Operations

| Feature | Weaviate Cloud | Weaviate Local | Supabase | Mock | Notes |
|---------|----------------|----------------|----------|------|-------|
| Semantic Search | ✅ | ✅ | ✅ | ✅ | Requires embeddings |
| BM25 Search | ✅ | ✅ | ✅ | ✅ | Supabase: Uses ts_rank_cd with length normalization |
| Hybrid Search | ✅ | ✅ | ✅ | ✅ | Combines semantic + BM25 |
| Metadata Search | ✅ | ✅ | ✅ | ✅ | - |

### Embedding Support

| Feature | Weaviate Cloud | Weaviate Local | Supabase | Mock | Notes |
|---------|----------------|----------------|----------|------|-------|
| OpenAI Embeddings | ✅ | ✅ | ✅ | ✅ | `text2vec-openai` |
| Cohere Embeddings | ✅ | ✅ | ❌ | ❌ | `text2vec-cohere` |
| Hugging Face | ✅ | ✅ | ❌ | ❌ | `text2vec-huggingface` |
| No Vectorizer | ✅ | ✅ | ✅ | ✅ | Manual embeddings |
| Custom Embeddings | ✅ | ✅ | ⚠️ | ✅ | Supabase: Limited |

### CLI Commands

| Command | Weaviate Cloud | Weaviate Local | Supabase | Mock | Notes |
|---------|----------------|----------------|----------|------|-------|
| `weave health check` | ✅ | ✅ | ✅ | ✅ | - |
| `weave cols ls` | ✅ | ✅ | ✅ | ✅ | - |
| `weave cols create` | ✅ | ✅ | ✅ | ✅ | - |
| `weave cols delete` | ✅ | ✅ | ✅ | ✅ | - |
| `weave cols schema` | ✅ | ✅ | ✅ | ✅ | - |
| `weave docs create` | ✅ | ✅ | ✅ | ✅ | - |
| `weave docs get` | ✅ | ✅ | ✅ | ✅ | - |
| `weave docs update` | ✅ | ✅ | ✅ | ✅ | - |
| `weave docs delete` | ✅ | ✅ | ✅ | ✅ | - |
| `weave docs ls` | ✅ | ✅ | ✅ | ✅ | - |
| `weave search semantic` | ✅ | ✅ | ✅ | ✅ | - |
| `weave search bm25` | ✅ | ✅ | ✅ | ✅ | - |
| `weave search hybrid` | ✅ | ✅ | ✅ | ✅ | - |
| `weave search metadata` | ✅ | ✅ | ✅ | ✅ | - |

### Configuration

| Feature | Weaviate Cloud | Weaviate Local | Supabase | Mock | Notes |
|---------|----------------|----------------|----------|------|-------|
| YAML Config | ✅ | ✅ | ✅ | ✅ | - |
| Env Variables | ✅ | ✅ | ✅ | ✅ | - |
| Global Config | ✅ | ✅ | ✅ | ✅ | `~/.weave-cli` |
| Multiple Databases | ✅ | ✅ | ✅ | ✅ | - |
| Schema Directory | ✅ | ✅ | ✅ | ✅ | - |

## Database-Specific Notes

### Weaviate (Cloud & Local)

**Strengths:**
- Full feature support across all operations
- Rich vectorizer ecosystem (OpenAI, Cohere, Hugging Face, etc.)
- Advanced hybrid search capabilities
- Native BM25 support
- Excellent schema validation

**Configuration:**
```yaml
databases:
  default: weaviate
  vector_databases:
    - name: weaviate
      type: weaviate-cloud  # or weaviate-local
      url: ${WEAVIATE_URL}
      api_key: ${WEAVIATE_API_KEY}
      openai_api_key: ${OPENAI_API_KEY}
```

**Known Limitations:**
- None identified

### Supabase

**Strengths:**
- PostgreSQL-based (familiar for many developers)
- Good semantic search support
- Standard metadata filtering
- Easy integration with existing PostgreSQL infrastructure

**Configuration:**
```yaml
databases:
  default: supabase
  vector_databases:
    - name: supabase
      type: supabase
      database_url: ${SUPABASE_DATABASE_URL}
      database_key: ${SUPABASE_DATABASE_KEY}
      openai_api_key: ${OPENAI_API_KEY}
      timeout: 30
```

**Known Limitations:**
- **Vectorizers**: Currently only supports OpenAI embeddings and manual embeddings
- **Collection Names**: Automatically normalizes names (underscores → hyphens)
- **Performance**: BM25 search works well but could be faster with pre-computed tsvector columns and GIN indexes (see docs/SUPABASE_BM25_IMPROVEMENT.md for optimization options)

**BM25 Implementation:**
- Uses PostgreSQL's `ts_rank_cd()` with document length normalization
- Provides good ranking quality comparable to other systems
- For better performance on large datasets, consider adding GIN indexes (optional)

### Mock Database

**Strengths:**
- Perfect for testing and development
- No external dependencies
- Simulates all operations
- Configurable embedding dimensions

**Configuration:**
```yaml
databases:
  default: mock
  vector_databases:
    - name: mock
      type: mock
      enabled: true
      simulate_embeddings: true
      embedding_dimension: 1536
```

**Known Limitations:**
- Not for production use
- Embeddings are simulated (random vectors)
- No persistence between runs
- Search results may not reflect real semantic similarity

## Integration Test Coverage

| Test Type | Weaviate | Supabase | Mock |
|-----------|----------|----------|------|
| Health Check | ✅ | ✅ | ✅ |
| Collection CRUD | ✅ | ✅ | ✅ |
| Document CRUD | ✅ | ✅ | ✅ |
| Batch Operations | ✅ | ✅ | ✅ |
| Semantic Search | ✅ | ✅ | ✅ |
| BM25 Search | ✅ | ✅ | ✅ |
| Hybrid Search | ✅ | ✅ | ✅ |
| Metadata Search | ✅ | ✅ | ✅ |
| Schema Operations | ✅ | ✅ | ✅ |
| OpenAI Embeddings | ✅ | ✅ | ✅ |
| No Vectorizer | ✅ | ✅ | ✅ |

## Roadmap

### Short Term
- [ ] Improve Supabase BM25 search implementation
- [ ] Add more embedding provider support for Supabase
- [ ] Document performance characteristics of each database

### Medium Term
- [ ] Add support for Pinecone
- [ ] Add support for Qdrant
- [ ] Implement database-specific optimization flags

### Long Term
- [ ] Multi-database queries
- [ ] Cross-database migration tools
- [ ] Unified embedding caching layer

## Contributing

When adding support for a new vector database or feature:

1. Update the feature support matrix in this document
2. Add integration tests in `tests/<database>_integration_test.go`
3. Ensure all core operations are implemented
4. Document any database-specific limitations
5. Update the CLI command documentation

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 0.3.11 | 2025-11-13 | Initial support matrix document |

## See Also

- [User Guide](USER_GUIDE.md) - General usage documentation
- [Configuration Guide](CONFIG.md) - Detailed configuration options
- [API Documentation](API.md) - Developer API reference
