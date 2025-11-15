# Vector Database Support Matrix

This document tracks feature support and compatibility across different vector database implementations in weave-cli.

## Supported Vector Databases

| Database | Type | Status | Config Type | Target Version |
|----------|------|--------|-------------|----------------|
| Weaviate Cloud | Cloud | ✅ Production | `weaviate-cloud` | v0.3.x |
| Weaviate Local | Self-hosted | ✅ Production | `weaviate-local` | v0.3.x |
| Supabase | Cloud/Self-hosted | 🧪 Alpha | `supabase` | v0.3.x |
| MongoDB Atlas | Cloud | 🚧 Experimental | `mongodb` | v0.3.15+ |
| Mock | Testing | ✅ Production | `mock` | v0.3.x |
| **Milvus** | **Cloud/Self-hosted** | **📋 Planned** | **`milvus`** | **v0.4.0** |
| **Qdrant** | **Cloud/Self-hosted** | **📋 Planned** | **`qdrant`** | **v0.5.0** |
| **Redis** | **Cloud/Self-hosted** | **📋 Planned** | **`redis`** | **v0.6.0** |
| **Pinecone** | **Cloud** | **📋 Planned** | **`pinecone`** | **v0.8.0** |

## Feature Support Matrix

### Core Operations

| Feature | Weaviate Cloud | Weaviate Local | Supabase | MongoDB | Mock | Notes |
|---------|----------------|----------------|----------|---------|------|-------|
| Health Check | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| List Collections | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| Create Collection | ✅ | ✅ | ✅ | ✅ | ✅ | MongoDB: Vector index requires Atlas UI |
| Delete Collection | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| Collection Exists | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| Get Collection Count | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| Get Schema | ✅ | ✅ | ✅ | ✅ | ✅ | MongoDB: Schema-less (returns default) |
| Validate Schema | ✅ | ✅ | ✅ | ✅ | ✅ | - |

### Document Operations

| Feature | Weaviate Cloud | Weaviate Local | Supabase | MongoDB | Mock | Notes |
|---------|----------------|----------------|----------|---------|------|-------|
| Create Document | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| Get Document | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| Update Document | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| Delete Document | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| List Documents | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| Batch Create | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| Delete by Metadata | ✅ | ✅ | ✅ | ✅ | ✅ | - |

### Search Operations

| Feature | Weaviate Cloud | Weaviate Local | Supabase | MongoDB | Mock | Notes |
|---------|----------------|----------------|----------|---------|------|-------|
| Semantic Search | ✅ | ✅ | ✅ | 🚧 | ✅ | MongoDB: Requires vector index + embeddings (not yet implemented) |
| BM25 Search | ✅ | ✅ | ✅ | ✅ | ✅ | MongoDB: Uses text indexes |
| Hybrid Search | ✅ | ✅ | ✅ | ⚠️ | ✅ | MongoDB: Falls back to BM25 only (vector search pending) |
| Metadata Search | ✅ | ✅ | ✅ | ✅ | ✅ | - |

### Embedding Support

| Feature | Weaviate Cloud | Weaviate Local | Supabase | MongoDB | Mock | Notes |
|---------|----------------|----------------|----------|---------|------|-------|
| OpenAI Embeddings | ✅ | ✅ | ✅ | 🚧 | ✅ | MongoDB: Planned |
| Cohere Embeddings | ✅ | ✅ | ❌ | ❌ | ❌ | `text2vec-cohere` |
| Hugging Face | ✅ | ✅ | ❌ | ❌ | ❌ | `text2vec-huggingface` |
| No Vectorizer | ✅ | ✅ | ✅ | ✅ | ✅ | Manual embeddings |
| Custom Embeddings | ✅ | ✅ | ⚠️ | ✅ | ✅ | Supabase: Limited |

### CLI Commands

| Command | Weaviate Cloud | Weaviate Local | Supabase | MongoDB | Mock | Notes |
|---------|----------------|----------------|----------|---------|------|-------|
| `weave health check` | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| `weave cols ls` | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| `weave cols create` | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| `weave cols delete` | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| `weave cols schema` | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| `weave docs create` | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| `weave docs get` | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| `weave docs update` | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| `weave docs delete` | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| `weave docs ls` | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| `weave search semantic` | ✅ | ✅ | ✅ | 🚧 | ✅ | MongoDB: Pending |
| `weave search bm25` | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| `weave search hybrid` | ✅ | ✅ | ✅ | ⚠️ | ✅ | MongoDB: BM25 only |
| `weave search metadata` | ✅ | ✅ | ✅ | ✅ | ✅ | - |

### Configuration

| Feature | Weaviate Cloud | Weaviate Local | Supabase | MongoDB | Mock | Notes |
|---------|----------------|----------------|----------|---------|------|-------|
| YAML Config | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| Env Variables | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| Global Config | ✅ | ✅ | ✅ | ✅ | ✅ | `~/.weave-cli` |
| Multiple Databases | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| Schema Directory | ✅ | ✅ | ✅ | ✅ | ✅ | - |

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

### Supabase (🧪 Alpha)

**Status:** Alpha - Core functionality working, production readiness being evaluated

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
- **Performance**: BM25 search works well but could be faster with pre-computed tsvector columns and GIN indexes (see [Supabase BM25 Improvement](supabase/BM25_IMPROVEMENT.md) for optimization options)
- **Maturity**: Alpha status - may have edge cases not yet covered

**BM25 Implementation:**
- Uses PostgreSQL's `ts_rank_cd()` with document length normalization
- Provides good ranking quality comparable to other systems
- For better performance on large datasets, consider adding GIN indexes (optional)

### MongoDB Atlas (🚧 Experimental)

**Status:** Experimental - Basic infrastructure complete, vector search embedding integration pending

**Strengths:**
- Popular document database (familiar to many developers)
- Managed Atlas service with free M0 tier (512MB)
- Native BM25 via text indexes (works immediately)
- Flexible document model
- Pre-filtering support for efficient queries

**Configuration:**
```yaml
databases:
  default: mongodb
  vector_databases:
    - name: mongodb
      type: mongodb
      url: ${MONGODB_URI}
      database: ${MONGODB_DATABASE}
      vector_dimensions: 1536
      similarity_metric: cosine
      timeout: 10
```

**Current Limitations:**
- **Vector Search**: Infrastructure ready, but embedding generation not yet integrated (returns error)
- **Hybrid Search**: Currently falls back to BM25-only (vector component pending)
- **Atlas Only**: Vector search requires MongoDB Atlas (not available in community edition)
- **Manual Index Setup**: Vector search indexes must be created via Atlas UI
- **Dimension Limit**: Max 8192 dimensions (lower than some other DBs)

**What Works:**
- ✅ Connection and health checks
- ✅ Collection management (create, list, delete)
- ✅ Document CRUD operations
- ✅ BM25 keyword search (fully functional)
- ✅ Metadata search and filtering
- ✅ Document storage with embedding fields

**What's Pending:**
- 🚧 Vector search (embedding generation integration needed)
- 🚧 True hybrid search (depends on vector search)
- 🚧 Integration tests with Atlas

**Documentation:**
- [MongoDB Integration Guide](mongodb/README.md)
- [Atlas Setup Instructions](mongodb/ATLAS_SETUP.md)

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

| Test Type | Weaviate | Supabase | MongoDB | Mock |
|-----------|----------|----------|---------|------|
| Health Check | ✅ | ✅ | 🚧 | ✅ |
| Collection CRUD | ✅ | ✅ | 🚧 | ✅ |
| Document CRUD | ✅ | ✅ | 🚧 | ✅ |
| Batch Operations | ✅ | ✅ | 🚧 | ✅ |
| Semantic Search | ✅ | ✅ | ❌ | ✅ |
| BM25 Search | ✅ | ✅ | 🚧 | ✅ |
| Hybrid Search | ✅ | ✅ | ❌ | ✅ |
| Metadata Search | ✅ | ✅ | 🚧 | ✅ |
| Schema Operations | ✅ | ✅ | 🚧 | ✅ |
| OpenAI Embeddings | ✅ | ✅ | ❌ | ✅ |
| No Vectorizer | ✅ | ✅ | 🚧 | ✅ |

**Legend:** ✅ Tested | 🚧 Planned | ❌ Not applicable/pending

## Roadmap

### v0.4.0 - Milvus Integration (Planned)

**Timeline**: ~2-3 weeks

- [ ] Implement MilvusClient with VectorDB interface
- [ ] Schema mapping (Weave → Milvus explicit schemas)
- [ ] Native BM25 + hybrid search support
- [ ] Geospatial data type support
- [ ] Multi-vector per document
- [ ] Docker-based integration tests
- [ ] Documentation and demo script

**Key Features**:
- ✅ Open source (Apache 2.0)
- ✅ Native BM25 full-text search
- ✅ Hybrid search (sparse + dense vectors)
- ✅ Geospatial support with GIS functions
- ✅ Distributed architecture for scale

See [Vector DB Integrations Planning](planning/VECTOR_DB_INTEGRATIONS.md) for
details.

### v0.5.0 - Qdrant Integration (Planned)

**Timeline**: ~1-2 weeks

- [ ] Implement QdrantClient with VectorDB interface
- [ ] gRPC client integration
- [ ] Point-based data model mapping
- [ ] JSON payload filtering
- [ ] HNSW + quantization support
- [ ] Docker-based integration tests
- [ ] Documentation and demo script

**Key Features**:
- ✅ Open source (Apache 2.0)
- ✅ High-performance gRPC API
- ✅ HNSW indexing + quantization
- ✅ Multiple vectors per point
- ✅ Flexible payload filtering

See [Vector DB Integrations Planning](planning/VECTOR_DB_INTEGRATIONS.md) for
details.

### v0.6.0 - Redis Integration (Planned)

**Timeline**: ~1-2 weeks

- [ ] Implement RedisClient with VectorDB interface
- [ ] RediSearch FT.CREATE/FT.SEARCH integration
- [ ] Hash-based document storage
- [ ] Hybrid search (vector + full-text)
- [ ] Geospatial search support
- [ ] Docker-based integration tests
- [ ] Documentation and demo script

**Key Features**:
- ✅ In-memory performance (extreme speed)
- ✅ Native BM25 full-text search via RediSearch
- ✅ Best-in-class hybrid search
- ✅ Geospatial support
- ✅ Familiar to many developers

**License Considerations**:
- RSALv2 (permissive, Redis 7.4+)
- SSPL v1 or AGPL v3 (older versions)
- Valkey fork available (BSD-3) if license is concern

See [Vector DB Integrations Planning](planning/VECTOR_DB_INTEGRATIONS.md) for
details.

### v0.3.15 - MongoDB Atlas Vector Search (Experimental - In Progress)

**Status**: 🚧 Experimental - Basic implementation complete, embedding integration pending

**Completed**:
- ✅ Implement MongoDBClient with VectorDB interface
- ✅ BM25 text search using MongoDB text indexes
- ✅ Document CRUD operations
- ✅ Collection management
- ✅ Configuration and factory integration
- ✅ Documentation and setup guides

**Pending**:
- [ ] OpenAI embedding generation integration
- [ ] Atlas Vector Search ($vectorSearch aggregation)
- [ ] Hybrid search (vector + BM25 combination)
- [ ] Atlas M0 free tier integration tests
- [ ] Demo script

**Key Features**:
- ✅ Popular document database (familiar to many)
- ✅ Managed Atlas service with free M0 tier (512MB)
- ✅ BM25 keyword search (functional)
- ✅ Pre-filtering for efficient queries
- ✅ Up to 8192 dimensions

**Current Limitations**:
- 🚧 Vector search pending embedding integration
- ⚠️ Atlas only (vector search not in community MongoDB)
- ⚠️ Manual index creation via Atlas UI required

See [MongoDB Documentation](mongodb/) for details.

### v0.7.0 - Pinecone Integration (Planned)

**Timeline**: ~1-2 weeks

- [ ] Implement PineconeClient with VectorDB interface
- [ ] API key authentication
- [ ] Namespace support for multi-tenancy
- [ ] Metadata filtering
- [ ] Hybrid search with keyword boosting
- [ ] Free tier CI/CD testing
- [ ] Pricing documentation
- [ ] Documentation and demo script

**Key Features**:
- ✅ Fully managed (zero infrastructure)
- ✅ Generous free tier (2GB storage, 2M writes/month)
- ✅ Serverless architecture
- ✅ Multi-cloud support (AWS, Azure, GCP)
- ✅ Hybrid search with keyword boosting

**Cost Considerations**:
- Free tier: 2GB storage, 2M writes/month
- Standard: $50/month minimum + pay-as-you-go

See [Vector DB Integrations Planning](planning/VECTOR_DB_INTEGRATIONS.md) for
details.

### Long Term (v0.9.0+)

- [ ] Multi-database queries
- [ ] Cross-database migration tools
- [ ] Unified embedding caching layer
- [ ] Additional vector DB support (Chroma, LanceDB, etc.)
- [ ] Path to v1.0.0 release

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

## Related Documentation

- [User Guide](USER_GUIDE.md) - Complete usage documentation
- [Vector DB Abstraction](guides/VECTOR_DB_ABSTRACTION.md) - Architecture details
- [Supabase Documentation](supabase/) - Supabase integration guide
- [Weaviate Documentation](weaviate/) - Weaviate integration status
- [Changelog](CHANGELOG.md) - Version history
