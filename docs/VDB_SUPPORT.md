# Vector Database Support Matrix

This document tracks feature support and compatibility across different vector database implementations in weave-cli.

## Supported Vector Databases

| Database | Type | Status | Config Type | Target Version |
|----------|------|--------|-------------|----------------|
| Weaviate Cloud | Cloud | ✅ Production | `weaviate-cloud` | v0.3.x |
| Weaviate Local | Self-hosted | ✅ Production | `weaviate-local` | v0.3.x |
| Milvus Local | Self-hosted | 🧪 Beta | `milvus-local` | v0.3.16+ |
| Milvus Cloud (Zilliz) | Cloud | 🧪 Beta | `milvus-cloud` | v0.3.16+ |
| Chroma Local | Self-hosted | ✅ Production | `chroma-local` | v0.6.0+ |
| Chroma Cloud | Cloud | ✅ Production | `chroma-cloud` | v0.6.0+ |
| Supabase | Cloud/Self-hosted | 🧪 Alpha | `supabase` | v0.3.x |
| MongoDB Atlas | Cloud | ✅ Functional | `mongodb` | v0.3.15+ |
| Mock | Testing | ✅ Production | `mock` | v0.3.x |
| **Qdrant** | **Cloud/Self-hosted** | **📋 Planned** | **`qdrant`** | **v0.5.0** |
| **Redis** | **Cloud/Self-hosted** | **📋 Planned** | **`redis`** | **v0.6.0** |
| **Pinecone** | **Cloud** | **📋 Planned** | **`pinecone`** | **v0.8.0** |

## Feature Support Matrix

### Core Operations

| Feature | Weaviate Cloud | Weaviate Local | Milvus Local | Milvus Cloud | Chroma Local | Chroma Cloud | Supabase | MongoDB | Mock | Notes |
|---------|----------------|----------------|--------------|--------------|--------------|--------------|----------|---------|------|-------|
| Health Check | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| List Collections | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| Create Collection | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | MongoDB: Vector index requires Atlas UI |
| Delete Collection | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| Collection Exists | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| Get Collection Count | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| Get Schema | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | Milvus: Explicit schema; MongoDB: Schema-less |
| Validate Schema | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | Milvus: Schema immutable after creation |

### Document Operations

| Feature | Weaviate Cloud | Weaviate Local | Milvus Local | Milvus Cloud | Chroma Local | Chroma Cloud | Supabase | MongoDB | Mock | Notes |
|---------|----------------|----------------|--------------|--------------|--------------|--------------|----------|---------|------|-------|
| Create Document | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | Milvus: Auto-embedding with OpenAI |
| Get Document | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| Update Document | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | Milvus: Delete + insert |
| Delete Document | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| List Documents | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| Batch Create | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| Delete by Metadata | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | - |

### Search Operations

| Feature | Weaviate Cloud | Weaviate Local | Milvus Local | Milvus Cloud | Chroma Local | Chroma Cloud | Supabase | MongoDB | Mock | Notes |
|---------|----------------|----------------|--------------|--------------|--------------|--------------|----------|---------|------|-------|
| Vector Search | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | Milvus: IVF_FLAT index |
| BM25 Search | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ✅ | ✅ | ✅ | Milvus: Native BM25 support; Chroma: Not supported |
| Hybrid Search | ✅ | ✅ | ✅ | ✅ | ⚠️ | ⚠️ | ✅ | ✅ | ✅ | Milvus: RRF fusion; Chroma: Falls back to vector search |
| Metadata Search | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | Milvus: JSON field filtering |

### Embedding Support

| Feature | Weaviate Cloud | Weaviate Local | Milvus Local | Milvus Cloud | Chroma Local | Chroma Cloud | Supabase | MongoDB | Mock | Notes |
|---------|----------------|----------------|--------------|--------------|--------------|--------------|----------|---------|------|-------|
| OpenAI Embeddings | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | Milvus: text-embedding-3-small default |
| Cohere Embeddings | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | Weaviate: `text2vec-cohere` |
| Hugging Face | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | Weaviate: `text2vec-huggingface` |
| No Vectorizer | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | Manual embeddings |
| Custom Embeddings | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ⚠️ | ✅ | ✅ | Supabase: Limited |

### CLI Commands

| Command | Weaviate Cloud | Weaviate Local | Milvus Local | Milvus Cloud | Chroma Local | Chroma Cloud | Supabase | MongoDB | Mock | Notes |
|---------|----------------|----------------|--------------|--------------|--------------|--------------|----------|---------|------|-------|
| `weave health check` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| `weave cols ls` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| `weave cols create` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| `weave cols delete` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| `weave cols schema` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| `weave docs create` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| `weave docs get` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| `weave docs update` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| `weave docs delete` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| `weave docs ls` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| `weave search semantic` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| `weave search bm25` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| `weave search hybrid` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| `weave search metadata` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | - |

### Configuration

| Feature | Weaviate Cloud | Weaviate Local | Milvus Local | Milvus Cloud | Chroma Local | Chroma Cloud | Supabase | MongoDB | Mock | Notes |
|---------|----------------|----------------|--------------|--------------|--------------|--------------|----------|---------|------|-------|
| YAML Config | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| Env Variables | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| Global Config | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | `~/.weave-cli` |
| Multiple Databases | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | - |
| Schema Directory | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | - |

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

### Chroma (Local & Cloud)

**Status:** Production - Fully functional with Chroma Go SDK v2 API

**Strengths:**
- Simple API, easy to get started
- Python-friendly ecosystem
- Good for development and prototyping
- Lightweight and fast for smaller datasets
- Automatic switching between local and cloud clients
- Free cloud tier available

**Considerations:**
- No native BM25 search (keyword search not supported)
- Simpler feature set than enterprise VDBs
- Hybrid search falls back to vector-only search
- Cloud free tier has quota limits (300 documents/request)

**Configuration (Local):**
```yaml
databases:
  default: chroma-local
  vector_databases:
    - name: chroma-local
      type: chroma-local
      url: ${CHROMA_URL}
      openai_api_key: ${OPENAI_API_KEY}
```

**Configuration (Cloud):**
```yaml
databases:
  default: chroma-cloud
  vector_databases:
    - name: chroma-cloud
      type: chroma-cloud
      url: ${CHROMA_CLOUD_URL}
      api_key: ${CHROMA_CLOUD_API_KEY}
      openai_api_key: ${OPENAI_API_KEY}
```

**Environment Variables:**
- `CHROMA_URL` - URL for local Chroma instance (e.g., `http://localhost:8000`)
- `CHROMA_CLOUD_URL` - URL for Chroma Cloud (defaults to
  `https://api.trychroma.com`)
- `CHROMA_CLOUD_API_KEY` or `CHROMA_API_KEY` - API key for Chroma Cloud
- `CHROMA_TENANT` - Team UUID for Chroma Cloud
- `CHROMA_DATABASE` - Database name for Chroma Cloud

**Known Limitations:**
- **Quota Limits**: Free tier limited to 300 documents per GET request (vs
  10,000 default)
- **No BM25**: Keyword search not supported natively
- Some integration tests may fail on free tier due to quota limits (expected
  behavior)

### Milvus (🧪 Beta)

**Status:** Beta - Core functionality complete, ready for testing and feedback

**Strengths:**
- High-performance vector database optimized for large-scale similarity search
- Native BM25 full-text search support
- Hybrid search with Reciprocal Rank Fusion (RRF)
- Both local (Docker/Podman) and cloud (Zilliz) deployment options
- Explicit schema with type safety
- Automatic embedding generation with OpenAI
- Multiple similarity metrics (L2, IP, COSINE)
- Advanced indexing (IVF_FLAT, HNSW, IVF_PQ, etc.)

**Configuration (Local):**
```yaml
databases:
  default: milvus-local
  vector_databases:
    - name: milvus-local
      type: milvus-local
      address: localhost:19530
      database: default
      vector_dimensions: 1536
      similarity_metric: L2
      timeout: 10
```

**Configuration (Cloud/Zilliz):**
```yaml
databases:
  default: milvus-cloud
  vector_databases:
    - name: milvus-cloud
      type: milvus-cloud
      address: ${MILVUS_CLOUD_ADDRESS}
      username: ${MILVUS_CLOUD_USERNAME}
      password: ${MILVUS_CLOUD_PASSWORD}
      database: default
      vector_dimensions: 1536
      similarity_metric: COSINE
      timeout: 30
```

**What Works:**
- ✅ Connection and health checks (local and cloud)
- ✅ Collection management with explicit schemas
- ✅ Document CRUD operations
- ✅ Automatic OpenAI embedding generation
- ✅ Vector similarity search (IVF_FLAT index)
- ✅ BM25 keyword search
- ✅ Hybrid search with RRF
- ✅ Metadata filtering (JSON field support)
- ✅ Batch operations
- ✅ Podman/Docker containerization

**Known Limitations:**
- **Schema Immutability**: Schemas cannot be updated after collection creation (Milvus design)
- **Update Operation**: Implemented as delete + insert (no native update)
- **Float32 Vectors**: Embeddings converted from float64 to float32
- **Local Setup**: Requires Docker or Podman for local development
- **Cloud Setup**: Requires Zilliz account for managed service

**Local Development:**
```bash
# Start Milvus with podman (preferred) or docker
./tools/vdb/local/milvus.sh start

# Check status
./tools/vdb/local/milvus.sh status

# View logs
./tools/vdb/local/milvus.sh logs

# Stop Milvus
./tools/vdb/local/milvus.sh stop
```

**Key Features:**
- **Multiple Index Types**: IVF_FLAT (default), HNSW, IVF_PQ for different use cases
- **Similarity Metrics**: L2 (Euclidean), IP (inner product), COSINE (angular)
- **Explicit Schema**: Strongly-typed fields with validation
- **Native BM25**: Built-in full-text search without additional setup
- **Hybrid Search**: Combines vector and BM25 with configurable RRF
- **JSON Metadata**: Flexible metadata storage and filtering
- **Production Ready**: ACID transactions, HA, horizontal scaling

**Performance:**
- IVF_FLAT index provides good balance of speed and accuracy
- Configurable nlist parameter (default: 128)
- Collections must be loaded into memory for search
- Automatic flush after write operations

**Container Runtime:**
- Podman preferred for better security and rootless operation
- Docker supported as fallback
- Automatic detection via `tools/vdb/container/detect.sh`
- SELinux-compatible volume mounts

**Documentation:**
- [Milvus Integration Guide](milvus/README.md)
- [Local Setup Instructions](milvus/LOCAL_SETUP.md)
- [Cloud Setup Guide (Zilliz)](milvus/CLOUD_SETUP.md)

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

| Test Type | Weaviate | Milvus | Chroma | Supabase | MongoDB | Mock |
|-----------|----------|--------|--------|----------|---------|------|
| Health Check | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Collection CRUD | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Document CRUD | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Batch Operations | ✅ | ✅ | ⚠️ | ✅ | ✅ | ✅ |
| Semantic Search | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| BM25 Search | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ |
| Hybrid Search | ✅ | ✅ | ⚠️ | ✅ | ✅ | ✅ |
| Metadata Search | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Schema Operations | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| OpenAI Embeddings | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| No Vectorizer | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| E2E Tests | ✅ | ✅ | ✅ | ✅ | ⚠️ | ✅ |

**Legend:** ✅ Tested | 🚧 Planned | ⚠️ Quota limits or known issues | ❌ Not
supported

## Roadmap

### v0.3.16 - Milvus Integration (✅ Completed)

**Status**: ✅ Beta - Core functionality complete, ready for testing

**Completed**:
- ✅ Implement Milvus Client with VectorDB interface
- ✅ Schema mapping (Weave → Milvus explicit schemas)
- ✅ Native BM25 + hybrid search with RRF
- ✅ Collection management (create, delete, list, exists, count, schema)
- ✅ Document CRUD operations
- ✅ Automatic OpenAI embedding generation
- ✅ Vector similarity search (IVF_FLAT index)
- ✅ Metadata filtering (JSON field support)
- ✅ Batch operations
- ✅ Local development setup (Docker/Podman)
- ✅ Cloud setup support (Zilliz)
- ✅ CLI flags (--milvus-local, --milvus-cloud)
- ✅ Factory registration and config validation
- ✅ Comprehensive documentation

**Remaining Tasks**:
- [ ] Integration tests with local Milvus
- [ ] Integration tests with Zilliz Cloud
- [ ] Demo script
- [ ] Performance benchmarking

**Key Features**:
- ✅ Open source (Apache 2.0)
- ✅ Native BM25 full-text search
- ✅ Hybrid search with RRF
- ✅ Both local and cloud deployment
- ✅ Explicit schema with type safety
- ✅ Multiple similarity metrics (L2, IP, COSINE)

**Documentation**:
- [Milvus Integration Guide](milvus/README.md)
- [Local Setup Instructions](milvus/LOCAL_SETUP.md)
- [Cloud Setup Guide](milvus/CLOUD_SETUP.md)

### v0.6.0 - Chroma Integration (✅ Completed)

**Timeline**: Nov 21-27, 2025

**Status**: ✅ Production - Fully functional with Chroma Go SDK v2 API

**Completed**:
- ✅ Add Chroma to VectorDBType enum
- ✅ Create chroma/ directory structure
- ✅ Add Chroma Go SDK v2 dependency
- ✅ Implement core client with automatic local/cloud switching
- ✅ Collection operations (create, delete, list, exists, count)
- ✅ Document CRUD operations
- ✅ Batch operations
- ✅ Search operations (semantic, metadata)
- ✅ Factory and CLI flags (--chroma-local, --chroma-cloud)
- ✅ Integration tests with quota limit handling
- ✅ Test script improvements (live output, summary reporting)
- ✅ Documentation updates

**Key Features**:
- ✅ Open source (Apache 2.0)
- ✅ Simple REST API
- ✅ SQLite (local) or ClickHouse (cloud) storage
- ✅ Automatic or manual embeddings
- ✅ Chroma Go SDK v2 API (NewCloudClient for cloud, NewHTTPClient for local)
- ✅ Free cloud tier with quota limits clearly documented

**Known Limitations**:
- ⚠️ No native BM25 support (keyword search not available)
- ⚠️ Cloud free tier limited to 300 documents per GET request
- ⚠️ Some integration tests fail on free tier due to quota (expected behavior)

### v0.7.0 - Qdrant Integration (Planned)

**Timeline**: Dec 2-6, 2025

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

### v0.8.0 - Neo4j Integration (Planned)

**Timeline**: Dec 9-13, 2025

- [ ] Implement Neo4jClient with VectorDB interface
- [ ] Graph + vector hybrid queries
- [ ] Cypher query integration
- [ ] Relationship-aware search
- [ ] Docker-based integration tests
- [ ] Documentation and demo script

**Key Features**:
- ✅ Graph relationships + vector search
- ✅ Enterprise adoption
- ✅ GraphRAG support

### v0.9.0 - OpenSearch Integration (Planned)

**Timeline**: ~1-2 weeks

- [ ] Implement OpenSearchClient with VectorDB interface
- [ ] kNN vector search
- [ ] BM25 + vector hybrid search
- [ ] Index management
- [ ] Docker-based integration tests
- [ ] Documentation and demo script

**Key Features**:
- ✅ Open source (Apache 2.0)
- ✅ Mature vector search with kNN
- ✅ Best-in-class hybrid search
- ✅ Elasticsearch API compatible
- ✅ Massive enterprise adoption

### v0.10.0 - Redis Integration (Planned)

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

### v0.3.15 - MongoDB Atlas Vector Search (✅ Fully Functional)

**Status**: ✅ Fully Functional - Complete with automatic embedding generation and vector search

**Completed**:
- ✅ Implement MongoDBClient with VectorDB interface
- ✅ Automatic OpenAI embedding generation on document creation
- ✅ Atlas Vector Search ($vectorSearch aggregation) - fully functional
- ✅ BM25 text search using MongoDB text indexes
- ✅ Hybrid search (vector + BM25 combination with RRF)
- ✅ Document CRUD operations with embedding support
- ✅ Document deletion by ID or filename
- ✅ Collection management
- ✅ Configuration and factory integration
- ✅ Documentation and setup guides

**Remaining Tasks**:
- [ ] Atlas M0 free tier integration tests
- [ ] Demo script

**Key Features**:
- ✅ Popular document database (familiar to many)
- ✅ Managed Atlas service with free M0 tier (512MB)
- ✅ Automatic embedding generation (requires `OPENAI_API_KEY`)
- ✅ Semantic vector search (requires Atlas vector index)
- ✅ BM25 keyword search (works without vector index)
- ✅ Hybrid search combining both approaches
- ✅ Pre-filtering for efficient queries
- ✅ Up to 8192 dimensions supported

**Requirements**:
- ✅ `OPENAI_API_KEY` environment variable for embeddings
- ✅ MongoDB Atlas cluster (free M0 tier available)
- ✅ Vector search index created via Atlas UI

**Limitations**:
- ⚠️ Atlas only (vector search not in community MongoDB)
- ⚠️ Manual index creation via Atlas UI required (not automated via API)

See [MongoDB Documentation](mongodb/) for complete setup and usage details.

### v1.0.0 - Pinecone Integration (Planned)

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

### Long Term (v1.1.0+)

- [ ] Multi-database queries
- [ ] Cross-database migration tools
- [ ] Unified embedding caching layer
- [ ] Additional vector DB support (LanceDB, etc.)

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
