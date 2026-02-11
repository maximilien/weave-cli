# Vector Database Support Matrix

Weave CLI supports multiple vector databases with varying levels of maturity
and features.

## Supported Databases

| VDB | Status | Local | Cloud | Platforms | Setup Guide |
|-----|--------|-------|-------|-----------|-------------|
| **Weaviate** | ✅ Stable | ✅ | ✅ | All | [SETUP.md](weaviate/SETUP.md) |
| **Qdrant** | ✅ Stable | ✅ | ✅ | All | [SETUP.md](qdrant/SETUP.md) |
| **Milvus** | ✅ Stable | ✅ | ✅ | All | [LOCAL_SETUP.md](milvus/LOCAL_SETUP.md), [CLOUD_SETUP.md](milvus/CLOUD_SETUP.md) |
| **Chroma** | ✅ Stable | ✅ | ✅ | macOS only | [SETUP.md](chroma/SETUP.md) |
| **Supabase** | ✅ Stable | ✅ | ✅ | All | [README.md](supabase/README.md) |
| **Neo4j** | ✅ Stable | ✅ | ⚠️ Untested | All | [README.md](neo4j/README.md), [AURA_SETUP.md](neo4j/AURA_SETUP.md) |
| **MongoDB Atlas** | ✅ Stable | ❌ | ✅ | All | [ATLAS_SETUP.md](mongodb/ATLAS_SETUP.md) |
| **Pinecone** | 🟢 Beta | ❌ | ✅ | All | [SETUP.md](pinecone/SETUP.md) |
| **OpenSearch** | ✅ Stable | ✅ | ✅ | All (2GB+ RAM) | [README.md](opensearch/README.md) |
| **Elasticsearch** | 🟢 Beta | ✅ | ✅ | All (2GB+ RAM) | [README.md](elasticsearch/README.md), [SETUP.md](elasticsearch/SETUP.md) |

## Status Legend

- ✅ **Stable**: Production-ready, fully tested, recommended for all use cases
- 🟢 **Beta**: Feature complete, tested, recommended for development and testing
- 🚧 **In Progress**: Active development, core features working, documentation incomplete
- ⚠️ **Untested**: Implementation complete but cloud deployment not verified

## Feature Comparison

| Feature | Weaviate | Qdrant | Milvus | Chroma | Supabase | Neo4j | MongoDB | Pinecone | OpenSearch | Elasticsearch |
|---------|----------|--------|--------|--------|----------|-------|---------|----------|------------|---------------|
| **Vector Search** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **BM25 Full-Text** | ✅ | ❌ | ✅ | ❌ | ✅ | ❌ | ✅ | ❌ | ✅ | ✅ |
| **Hybrid Search** | ✅ | ⚠️ | ✅ | ⚠️ | ✅ | ❌ | ✅ | ⚠️ | ✅ | ✅ |
| **Metadata Filtering** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Batch Operations** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ⚠️ | ✅ |
| **Collection Management** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Schema Management** | ✅ | ⚠️ | ⚠️ | ✅ | ✅ | ❌ | ✅ | ⚠️ | ⚠️ | ✅ |
| **Auto Embeddings** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **OSS Embeddings** [v0.9.19+] | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Collection Re-embedding** [v0.9.19+] | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **AI Schema Suggestions** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **AI Chunking Suggestions** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |

**Legend:**
- ✅ Fully supported
- ⚠️ Partial support, workaround, or special behavior
- ❌ Not supported

**Feature Notes:**
- **Chroma Hybrid**: Falls back to semantic search (no error)
- **Qdrant Schema**: Immutable after collection creation
- **Milvus Schema**: Immutable after collection creation
- **Neo4j Schema**: Not supported - delete and recreate collection instead
- **OpenSearch Batch**: Sequential operations (no native bulk API yet)
- **Pinecone Hybrid**: Falls back to semantic search (vector-only database)
- **AI Schema Suggestions**: Use `weave schema suggest` to analyze documents and get AI-powered schema recommendations
- **AI Chunking Suggestions**: Use `weave chunking suggest` to analyze documents and get optimal chunk size recommendations
- **OSS Embeddings** [v0.9.19+]: Provider-independent open-source embedding support (sentence-transformers, Ollama)
  - Works with ALL vector databases (pre-generated embeddings via `doc.Embedding` field)
  - No VDB-specific configuration required
  - Supports OpenAI, sentence-transformers, Ollama out of the box
  - Use `weave embeddings list` to see available models
- **Collection Re-embedding** [v0.9.19+]: Fast model switching without document re-ingestion
  - Re-embed existing collections with different embedding models
  - 20x faster than full re-ingestion (200+ docs/min typical)
  - Perfect for testing model quality, switching providers, cost optimization
  - Use `weave collection re-embed SOURCE --new-embedding MODEL --output DEST`
  - Queries automatically use collection's embedding model (dimension auto-matching)

## Quick Start

### General Setup

1. **Install Weave CLI**:

   ```bash
   git clone https://github.com/maximilien/weave-cli.git
   cd weave-cli
   ./build.sh
   ```

2. **Choose your vector database** from the matrix above

3. **Follow the setup guide** for your chosen database

4. **Configure Weave** using interactive setup:

   ```bash
   weave config create --env
   ```

5. **Verify connection**:

   ```bash
   weave health check
   ```

## Platform Notes

- **Chroma**: macOS only due to CGO dependencies in chroma-go SDK
- **All others**: Cross-platform (Linux, macOS, Windows)

## Getting Help

- See individual setup guides for database-specific configuration
- Check [docs/](../) for additional documentation
- Report issues at [github.com/maximilien/weave-cli/issues](https://github.com/maximilien/weave-cli/issues)
