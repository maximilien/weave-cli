# Vector Database Support Matrix

This document tracks feature support and compatibility across different vector database implementations in weave-cli.

## Supported Vector Databases

| Database | Type | Status | Config Type | Target Version |
|----------|------|--------|-------------|----------------|
| Weaviate Cloud | Cloud | ✅ Production | `weaviate-cloud` | v0.3.x |
| Weaviate Local | Self-hosted | ✅ Production | `weaviate-local` | v0.3.x |
| Milvus Local | Self-hosted | 🧪 Beta | `milvus-local` | v0.3.16+ |
| Milvus Cloud (Zilliz) | Cloud | 🧪 Beta | `milvus-cloud` | v0.3.16+ |
| Chroma Local | Self-hosted | ✅ Production | `chroma-local` | v0.6.0+ (macOS only) |
| Chroma Cloud | Cloud | ✅ Production | `chroma-cloud` | v0.6.0+ (macOS only) |
| Qdrant Local | Self-hosted | 🧪 Experimental | `qdrant-local` | v0.7.0+ |
| Qdrant Cloud | Cloud | 🧪 Experimental | `qdrant-cloud` | v0.7.0+ |
| Neo4j Local | Self-hosted | ✅ Production | `neo4j-local` | v0.7.1+ |
| Neo4j Cloud (Aura) | Cloud | ⚠️ Untested | `neo4j-cloud` | v0.7.1+ |
| OpenSearch Local | Self-hosted | 🧪 Experimental | `opensearch-local` | v0.7.3+ |
| OpenSearch Cloud (AWS) | Cloud | 🧪 Experimental | `opensearch-cloud` | v0.7.3+ |
| Supabase | Cloud/Self-hosted | 🧪 Alpha | `supabase` | v0.3.x |
| MongoDB Atlas | Cloud | ✅ Functional | `mongodb` | v0.3.15+ |
| Mock | Testing | ✅ Production | `mock` | v0.3.x |
| **Redis** | **Cloud/Self-hosted** | **📋 Planned** | **`redis`** | **v0.6.0** |
| **Pinecone** | **Cloud** | **📋 Planned** | **`pinecone`** | **v0.8.0** |

## Feature Support Matrix

### Core Operations

| Feature | Weaviate Cloud | Weaviate Local | Milvus Local | Milvus Cloud | Chroma Local | Chroma Cloud | Qdrant Local | Qdrant Cloud | Neo4j Local | Neo4j Cloud | OpenSearch Local | OpenSearch Cloud | Supabase | MongoDB | Mock | Notes |
|---------|----------------|----------------|--------------|--------------|--------------|--------------|--------------|--------------|-------------|-------------|------------------|------------------|----------|---------|------|-------|
| Health Check | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🧪 | 🧪 | ✅ | ⚠️ | 🧪 | 🧪 | ✅ | ✅ | ✅ | - |
| List Collections | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🧪 | 🧪 | ✅ | ⚠️ | 🧪 | 🧪 | ✅ | ✅ | ✅ | - |
| Create Collection | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🧪 | 🧪 | ✅ | ⚠️ | 🧪 | 🧪 | ✅ | ✅ | ✅ | MongoDB: Vector index requires Atlas UI; Neo4j: Creates VECTOR INDEX; OpenSearch: kNN index |
| Delete Collection | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🧪 | 🧪 | ✅ | ⚠️ | 🧪 | 🧪 | ✅ | ✅ | ✅ | - |
| Collection Exists | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🧪 | 🧪 | ✅ | ⚠️ | 🧪 | 🧪 | ✅ | ✅ | ✅ | - |
| Get Collection Count | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🧪 | 🧪 | ✅ | ⚠️ | 🧪 | 🧪 | ✅ | ✅ | ✅ | - |
| Get Schema | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🧪 | 🧪 | ✅ | ⚠️ | 🧪 | 🧪 | ✅ | ✅ | ✅ | Milvus: Explicit schema; MongoDB: Schema-less; Neo4j: Flexible properties; OpenSearch: Dynamic mapping |
| Validate Schema | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🧪 | 🧪 | ✅ | ⚠️ | 🧪 | 🧪 | ✅ | ✅ | ✅ | Milvus: Schema immutable after creation; Neo4j: Validates dimensions; OpenSearch: Validates dimensions |

### Document Operations

| Feature | Weaviate Cloud | Weaviate Local | Milvus Local | Milvus Cloud | Chroma Local | Chroma Cloud | Qdrant Local | Qdrant Cloud | Neo4j Local | Neo4j Cloud | OpenSearch Local | OpenSearch Cloud | Supabase | MongoDB | Mock | Notes |
|---------|----------------|----------------|--------------|--------------|--------------|--------------|--------------|--------------|-------------|-------------|------------------|------------------|----------|---------|------|-------|
| Create Document | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🧪 | 🧪 | ✅ | ⚠️ | 🧪 | 🧪 | ✅ | ✅ | ✅ | Milvus/Qdrant/Neo4j/OpenSearch: Auto-embedding with OpenAI |
| Get Document | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🧪 | 🧪 | ✅ | ⚠️ | 🧪 | 🧪 | ✅ | ✅ | ✅ | - |
| Update Document | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🧪 | 🧪 | ✅ | ⚠️ | 🧪 | 🧪 | ✅ | ✅ | ✅ | Milvus/Qdrant/OpenSearch: Delete + insert; Neo4j: Native MERGE |
| Delete Document | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🧪 | 🧪 | ✅ | ⚠️ | 🧪 | 🧪 | ✅ | ✅ | ✅ | - |
| List Documents | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🧪 | 🧪 | ✅ | ⚠️ | 🧪 | 🧪 | ✅ | ✅ | ✅ | - |
| Batch Create | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🧪 | 🧪 | ✅ | ⚠️ | 🧪 | 🧪 | ✅ | ✅ | ✅ | - |
| Delete by Metadata | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🧪 | 🧪 | ✅ | ⚠️ | 🧪 | 🧪 | ✅ | ✅ | ✅ | - |

### Search Operations

| Feature | Weaviate Cloud | Weaviate Local | Milvus Local | Milvus Cloud | Chroma Local | Chroma Cloud | Qdrant Local | Qdrant Cloud | Neo4j Local | Neo4j Cloud | OpenSearch Local | OpenSearch Cloud | Supabase | MongoDB | Mock | Notes |
|---------|----------------|----------------|--------------|--------------|--------------|--------------|--------------|--------------|-------------|-------------|------------------|------------------|----------|---------|------|-------|
| Vector Search | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🧪 | 🧪 | ✅ | ⚠️ | 🧪 | 🧪 | ✅ | ✅ | ✅ | Milvus: IVF_FLAT; Qdrant/Neo4j/OpenSearch: HNSW |
| BM25 Search | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | 🧪 | 🧪 | ✅ | ✅ | ✅ | Milvus/OpenSearch: Native BM25; Chroma/Qdrant/Neo4j: Not supported |
| Hybrid Search | ✅ | ✅ | ✅ | ✅ | ⚠️ | ⚠️ | ⚠️ | ⚠️ | ❌ | ❌ | 🧪 | 🧪 | ✅ | ✅ | ✅ | Milvus/OpenSearch: RRF fusion; Chroma/Qdrant: Falls back to vector search; Neo4j: Not supported |
| Metadata Search | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🧪 | 🧪 | ✅ | ⚠️ | 🧪 | 🧪 | ✅ | ✅ | ✅ | Milvus/Qdrant/Neo4j/OpenSearch: JSON field filtering |

### Embedding Support

| Feature | Weaviate Cloud | Weaviate Local | Milvus Local | Milvus Cloud | Chroma Local | Chroma Cloud | Qdrant Local | Qdrant Cloud | Neo4j Local | Neo4j Cloud | OpenSearch Local | OpenSearch Cloud | Supabase | MongoDB | Mock | Notes |
|---------|----------------|----------------|--------------|--------------|--------------|--------------|--------------|--------------|-------------|-------------|------------------|------------------|----------|---------|------|-------|
| OpenAI Embeddings | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🧪 | 🧪 | ✅ | ⚠️ | 🧪 | 🧪 | ✅ | ✅ | ✅ | Milvus/Qdrant/Neo4j/OpenSearch: text-embedding-3-small default |
| Cohere Embeddings | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | Weaviate: `text2vec-cohere` |
| Hugging Face | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | Weaviate: `text2vec-huggingface` |
| No Vectorizer | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🧪 | 🧪 | ✅ | ⚠️ | 🧪 | 🧪 | ✅ | ✅ | ✅ | Manual embeddings |
| Custom Embeddings | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🧪 | 🧪 | ✅ | ⚠️ | 🧪 | 🧪 | ⚠️ | ✅ | ✅ | Supabase/OpenSearch: Limited |

### CLI Commands

| Command | Weaviate Cloud | Weaviate Local | Milvus Local | Milvus Cloud | Chroma Local | Chroma Cloud | Qdrant Local | Qdrant Cloud | Neo4j Local | Neo4j Cloud | OpenSearch Local | OpenSearch Cloud | Supabase | MongoDB | Mock | Notes |
|---------|----------------|----------------|--------------|--------------|--------------|--------------|--------------|--------------|-------------|-------------|------------------|------------------|----------|---------|------|-------|
| `weave health check` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🧪 | 🧪 | ✅ | ⚠️ | 🧪 | 🧪 | ✅ | ✅ | ✅ | - |
| `weave cols ls` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🧪 | 🧪 | ✅ | ⚠️ | 🧪 | 🧪 | ✅ | ✅ | ✅ | - |
| `weave cols create` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🧪 | 🧪 | ✅ | ⚠️ | 🧪 | 🧪 | ✅ | ✅ | ✅ | - |
| `weave cols delete` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🧪 | 🧪 | ✅ | ⚠️ | 🧪 | 🧪 | ✅ | ✅ | ✅ | - |
| `weave cols schema` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🧪 | 🧪 | ✅ | ⚠️ | 🧪 | 🧪 | ✅ | ✅ | ✅ | - |
| `weave docs create` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🧪 | 🧪 | ✅ | ⚠️ | 🧪 | 🧪 | ✅ | ✅ | ✅ | - |
| `weave docs get` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🧪 | 🧪 | ✅ | ⚠️ | 🧪 | 🧪 | ✅ | ✅ | ✅ | - |
| `weave docs update` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🧪 | 🧪 | ✅ | ⚠️ | 🧪 | 🧪 | ✅ | ✅ | ✅ | - |
| `weave docs delete` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🧪 | 🧪 | ✅ | ⚠️ | 🧪 | 🧪 | ✅ | ✅ | ✅ | - |
| `weave docs ls` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🧪 | 🧪 | ✅ | ⚠️ | 🧪 | 🧪 | ✅ | ✅ | ✅ | - |
| `weave search semantic` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🧪 | 🧪 | ✅ | ⚠️ | 🧪 | 🧪 | ✅ | ✅ | ✅ | - |
| `weave search bm25` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🧪 | 🧪 | ❌ | ❌ | 🧪 | 🧪 | ✅ | ✅ | ✅ | Neo4j: Not supported |
| `weave search hybrid` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🧪 | 🧪 | ❌ | ❌ | 🧪 | 🧪 | ✅ | ✅ | ✅ | Neo4j: Not supported |
| `weave search metadata` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🧪 | 🧪 | ✅ | ⚠️ | 🧪 | 🧪 | ✅ | ✅ | ✅ | - |

### Configuration

| Feature | Weaviate Cloud | Weaviate Local | Milvus Local | Milvus Cloud | Chroma Local | Chroma Cloud | Qdrant Local | Qdrant Cloud | Neo4j Local | Neo4j Cloud | OpenSearch Local | OpenSearch Cloud | Supabase | MongoDB | Mock | Notes |
|---------|----------------|----------------|--------------|--------------|--------------|--------------|--------------|--------------|-------------|-------------|------------------|------------------|----------|---------|------|-------|
| YAML Config | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🧪 | 🧪 | ✅ | ⚠️ | 🧪 | 🧪 | ✅ | ✅ | ✅ | - |
| Env Variables | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🧪 | 🧪 | ✅ | ⚠️ | 🧪 | 🧪 | ✅ | ✅ | ✅ | - |
| Global Config | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🧪 | 🧪 | ✅ | ⚠️ | 🧪 | 🧪 | ✅ | ✅ | ✅ | `~/.weave-cli` |
| Multiple Databases | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🧪 | 🧪 | ✅ | ⚠️ | 🧪 | 🧪 | ✅ | ✅ | ✅ | - |
| Schema Directory | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | 🧪 | 🧪 | ✅ | ⚠️ | 🧪 | 🧪 | ✅ | ✅ | ✅ | - |

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

**Platform Support:** ⚠️ **macOS only (AMD64/ARM64)** - Due to chroma-go v0.2.5
SDK's CGO dependency (libtokenizers), Chroma is not supported on Linux or Windows.
For other platforms, use Weaviate, Milvus, Qdrant, MongoDB, or Supabase.

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
- **macOS only** - Linux/Windows not supported due to SDK limitations

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

### Qdrant (🧪 Experimental)

**Status:** Experimental - Core functionality complete, testing in progress

**Strengths:**
- High-performance vector search with HNSW indexing
- gRPC and REST API support
- Advanced filtering capabilities
- Efficient payload storage and retrieval
- Both local (Docker/Podman) and cloud deployment
- Written in Rust for speed and safety
- Quantization support for memory efficiency

**Configuration (Local):**
```yaml
databases:
  default: qdrant-local
  vector_databases:
    - name: qdrant-local
      type: qdrant-local
      host: localhost
      port: 6334  # gRPC port
      vector_dimensions: 1536
      similarity_metric: Cosine
      timeout: 10
```

**Configuration (Cloud):**
```yaml
databases:
  default: qdrant-cloud
  vector_databases:
    - name: qdrant-cloud
      type: qdrant-cloud
      url: ${QDRANT_URL}
      api_key: ${QDRANT_API_KEY}
      vector_dimensions: 1536
      similarity_metric: Cosine
      timeout: 30
```

**What Works:**
- 🧪 Connection and health checks (gRPC)
- 🧪 Collection management (create, delete, list, exists, count)
- 🧪 Document CRUD operations
- 🧪 Automatic OpenAI embedding generation
- 🧪 Vector similarity search (HNSW index)
- 🧪 Metadata filtering (payload-based)
- 🧪 Batch operations

**Known Limitations:**
- **Experimental Status**: Not yet tested in production
- **No BM25**: Keyword search not supported natively
- **Hybrid Search**: Falls back to vector-only search
- **Float32 Vectors**: Embeddings converted from float64 to float32
- **Testing Needed**: Local and cloud instances need real-world validation

**Local Development:**
```bash
# Start Qdrant with Docker
docker run -p 6333:6333 -p 6334:6334 \
  -v $(pwd)/qdrant_storage:/qdrant/storage:z \
  qdrant/qdrant

# Or with Podman
podman run -p 6333:6333 -p 6334:6334 \
  -v $(pwd)/qdrant_storage:/qdrant/storage:z \
  qdrant/qdrant
```

**Key Features:**
- ✅ Open source (Apache 2.0)
- ✅ gRPC API for high performance
- ✅ HNSW indexing for fast similarity search
- ✅ Flexible payload filtering with JSON support
- ✅ Quantization for reduced memory usage
- ✅ Both local and cloud deployment options

**Performance:**
- HNSW index provides excellent speed and accuracy
- Efficient payload storage separate from vectors
- Configurable index parameters
- Support for multiple distance metrics (Cosine, Dot, Euclidean)

**Documentation:**
- [Qdrant Setup Guide](qdrant/SETUP.md)

### Neo4j (✅ Production - Local Only)

**Status:** Production (Local), Cloud Untested

**Strengths:**
- Graph database with vector search (best of both worlds)
- ACID transactions for data consistency
- Combines relationship queries with semantic search
- High-performance HNSW vector indexing
- Mature enterprise database
- Flexible property-based metadata
- Native Cypher query language for complex queries

**Configuration (Local):**
```yaml
databases:
  default: neo4j
  vector_databases:
    - name: neo4j
      type: neo4j-local
      url: ${NEO4J_URL:-bolt://localhost:7687}
      username: ${NEO4J_USERNAME:-neo4j}
      password: ${NEO4J_PASSWORD}
      database: ${NEO4J_DATABASE:-neo4j}
      timeout: 30
      vector_dimensions: 1536
      similarity_metric: cosine
```

**What Works:**
- ✅ Connection and health checks
- ✅ Collection management (create, list, delete, exists, count)
- ✅ Document CRUD operations
- ✅ Automatic OpenAI embedding generation
- ✅ Vector similarity search (HNSW index)
- ✅ Metadata filtering with Cypher WHERE clauses
- ✅ Batch operations
- ✅ Graph + vector combined queries

**Known Limitations:**
- **No BM25**: Keyword search not supported natively
- **No Hybrid Search**: Vector-only search
- **Local Only**: Cloud (Aura) support not yet tested
- **Vector Dimensions**: Must be consistent across collection
- **Requires Neo4j 5.11+**: Vector search added in 5.11

**Local Development:**
```bash
# Start Neo4j with Docker/Podman
./tools/vdb/local/neo4j.sh start

# Check status
./tools/vdb/local/neo4j.sh status

# View logs
./tools/vdb/local/neo4j.sh logs

# Stop Neo4j
./tools/vdb/local/neo4j.sh stop
```

**Key Features:**
- ✅ Open source (GPL/Commercial)
- ✅ Graph relationships + vector search
- ✅ HNSW indexing for fast similarity search
- ✅ ACID transactions
- ✅ Flexible metadata (any property)
- ✅ Cypher query language
- ✅ Enterprise-ready (clustering, backup, monitoring)

**Unique Capabilities:**
- **GraphRAG**: Combine graph traversal with vector search
- **Relationship-Aware Search**: Find similar documents connected by relationships
- **Complex Queries**: Use Cypher for sophisticated multi-hop queries

**Performance:**
- HNSW index provides excellent speed and accuracy
- Property-based storage for flexible metadata
- Configurable index parameters (M, ef_construction)
- Support for cosine and euclidean similarity

**Use Cases:**
- Knowledge graphs with semantic search
- Recommendation systems
- Fraud detection with similarity
- Complex relationship analysis + RAG

**Documentation:**
- [Neo4j Integration Guide](neo4j/README.md)
- [Neo4j Vector Search Docs](https://neo4j.com/docs/cypher-manual/current/indexes-for-vector-search/)

### OpenSearch (🧪 Experimental - Local & Cloud)

**Status:** Experimental - Core functionality complete, testing in progress

**Strengths:**
- Enterprise-grade search engine with vector capabilities
- High-performance k-NN vector search with HNSW algorithm
- Native BM25 full-text search
- Hybrid search combining vector and keyword approaches
- Both local (Docker/Podman) and cloud (AWS OpenSearch Service) deployment
- Elasticsearch API compatible
- Mature open-source project
- Advanced filtering capabilities

**Configuration (Local):**
```yaml
databases:
  default: opensearch-local
  vector_databases:
    - name: opensearch-local
      type: opensearch-local
      url: ${OPENSEARCH_URL:-https://localhost:9200}
      username: ${OPENSEARCH_USERNAME:-admin}
      password: ${OPENSEARCH_PASSWORD}
      vector_dimensions: 1536
      similarity_metric: cosinesimil
      timeout: 10
```

**Configuration (Cloud/AWS):**
```yaml
databases:
  default: opensearch-cloud
  vector_databases:
    - name: opensearch-cloud
      type: opensearch-cloud
      url: ${OPENSEARCH_CLOUD_URL}
      username: ${OPENSEARCH_CLOUD_USERNAME}
      password: ${OPENSEARCH_CLOUD_PASSWORD}
      vector_dimensions: 1536
      similarity_metric: cosinesimil
      timeout: 30
```

**What Works:**
- 🧪 Connection and health checks
- 🧪 Collection management (create, delete, list, exists, count)
- 🧪 Document CRUD operations
- 🧪 Automatic OpenAI embedding generation
- 🧪 Vector similarity search (k-NN with HNSW)
- 🧪 BM25 keyword search
- 🧪 Hybrid search with RRF
- 🧪 Metadata filtering (JSON field support)
- 🧪 Batch operations

**Known Limitations:**
- **Experimental Status**: New integration, limited production testing
- **Container Stability**: Local container may exit unexpectedly (code 137)
- **SSL Configuration**: Self-signed certificates for local setup
- **Float32 Vectors**: Embeddings converted from float64 to float32
- **Index Management**: Requires proper index mapping setup

**Local Development:**
```bash
# Start OpenSearch with Docker/Podman
./tools/vdb/local/opensearch.sh start

# Check status
./tools/vdb/local/opensearch.sh status

# View logs
./tools/vdb/local/opensearch.sh logs

# Stop OpenSearch
./tools/vdb/local/opensearch.sh stop
```

**Key Features:**
- ✅ Open source (Apache 2.0)
- ✅ k-NN vector search with HNSW indexing (lucene engine)
- ✅ Native BM25 full-text search
- ✅ Hybrid search with score combination
- ✅ Advanced JSON filtering
- ✅ Elasticsearch API compatibility
- ✅ Both local and cloud deployment

**Performance:**
- HNSW index provides excellent speed and accuracy
- Configurable ef_construction and M parameters
- Support for multiple similarity metrics (cosine, l2, inner product)
- Efficient filtering with OpenSearch query DSL

**Use Cases:**
- Enterprise search with semantic capabilities
- Log analytics with vector search
- E-commerce product search
- Document retrieval systems
- Knowledge base search

**Documentation:**
- [OpenSearch Integration Guide](opensearch/README.md)

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

| Test Type | Weaviate | Milvus | Chroma | Qdrant | Neo4j | OpenSearch | Supabase | MongoDB | Mock |
|-----------|----------|--------|--------|--------|-------|------------|----------|---------|------|
| Health Check | ✅ | ✅ | ✅ | 🧪 | ✅ | 🧪 | ✅ | ✅ | ✅ |
| Collection CRUD | ✅ | ✅ | ✅ | 🧪 | ✅ | 🧪 | ✅ | ✅ | ✅ |
| Document CRUD | ✅ | ✅ | ✅ | 🧪 | ✅ | 🧪 | ✅ | ✅ | ✅ |
| Batch Operations | ✅ | ✅ | ⚠️ | 🧪 | ✅ | 🧪 | ✅ | ✅ | ✅ |
| Semantic Search | ✅ | ✅ | ✅ | 🧪 | ✅ | 🧪 | ✅ | ✅ | ✅ |
| BM25 Search | ✅ | ✅ | ❌ | ❌ | ❌ | 🧪 | ✅ | ✅ | ✅ |
| Hybrid Search | ✅ | ✅ | ⚠️ | ⚠️ | ❌ | 🧪 | ✅ | ✅ | ✅ |
| Metadata Search | ✅ | ✅ | ✅ | 🧪 | ✅ | 🧪 | ✅ | ✅ | ✅ |
| Schema Operations | ✅ | ✅ | ✅ | 🧪 | ✅ | 🧪 | ✅ | ✅ | ✅ |
| OpenAI Embeddings | ✅ | ✅ | ✅ | 🧪 | ✅ | 🧪 | ✅ | ✅ | ✅ |
| No Vectorizer | ✅ | ✅ | ✅ | 🧪 | ✅ | 🧪 | ✅ | ✅ | ✅ |
| E2E Tests | ✅ | ✅ | ✅ | 🚧 | ✅ | 🧪 | ✅ | ⚠️ | ✅ |

**Legend:** ✅ Tested | 🧪 Experimental (implemented, needs real-world testing) | 🚧 Planned | ⚠️ Quota limits or known issues | ❌ Not supported

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

### v0.7.0 - Qdrant Integration (🧪 Experimental)

**Timeline**: Nov 28 - Dec 3, 2025

**Status**: 🧪 Experimental - Core implementation complete, testing in progress

**Completed**:
- ✅ Implement QdrantClient with VectorDB interface
- ✅ gRPC client integration (port 6334)
- ✅ Point-based data model mapping
- ✅ JSON payload filtering
- ✅ Collection operations (create, delete, list, exists, count)
- ✅ Document CRUD operations
- ✅ Automatic OpenAI embedding generation
- ✅ Vector similarity search (HNSW index)
- ✅ Batch operations
- ✅ Factory registration and config validation
- ✅ CLI flags (--qdrant-local, --qdrant-cloud)
- ✅ Integration test suite

**Remaining Tasks**:
- [ ] Test with local Qdrant instance (Docker/Podman)
- [ ] Test with Qdrant Cloud
- [ ] Documentation (SETUP.md)
- [ ] Demo script
- [ ] Production readiness validation

**Key Features**:
- ✅ Open source (Apache 2.0)
- ✅ High-performance gRPC API
- ✅ HNSW indexing for fast similarity search
- ✅ Flexible payload filtering
- ✅ Both local and cloud deployment
- ✅ Float32 vector support with auto-conversion

**Documentation**:
- [Qdrant Setup Guide](qdrant/SETUP.md)

### v0.7.1 - Neo4j Integration (✅ Completed - Local)

**Timeline**: Dec 1-2, 2025

**Status**: ✅ Production (Local Only) - Core functionality complete and tested

**Completed**:
- ✅ Implement Neo4jClient with VectorDB interface
- ✅ Collection operations (create, delete, list, exists, count)
- ✅ Document CRUD operations
- ✅ Automatic OpenAI embedding generation
- ✅ Vector similarity search (HNSW index)
- ✅ Metadata filtering with Cypher WHERE clauses
- ✅ Batch operations
- ✅ Factory registration and config validation
- ✅ CLI flags (--neo4j-local, --neo4j-cloud)
- ✅ Integration test suite (4 test suites passing)
- ✅ test.sh support with --neo4j flag
- ✅ Comprehensive documentation (docs/neo4j/README.md)
- ✅ Local management script (tools/vdb/local/neo4j.sh)

**Key Features**:
- ✅ Graph relationships + vector search
- ✅ HNSW indexing for fast similarity search
- ✅ Cypher query language support
- ✅ ACID transactions
- ✅ Flexible metadata (any property)
- ✅ GraphRAG support ready

**Known Limitations**:
- ⚠️ No BM25 keyword search
- ⚠️ No hybrid search
- ⚠️ Cloud (Aura) support untested
- ⚠️ Requires Neo4j 5.11+ for vector support

**Remaining Tasks**:
- [ ] Test with Neo4j Cloud (Aura)
- [ ] Test batch document creation (100+ documents)
- [ ] Decide on cloud support status
- [ ] Production validation at scale

**Documentation**:
- [Neo4j Integration Guide](neo4j/README.md)

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
