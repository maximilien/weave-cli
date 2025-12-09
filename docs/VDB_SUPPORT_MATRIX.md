# Vector Database Support Matrix

Weave CLI supports multiple vector databases with varying levels of maturity
and features.

## Supported Databases

| VDB | Status | Local | Cloud | Platforms | Setup Guide |
|-----|--------|-------|-------|-----------|-------------|
| **Weaviate** | ✅ Stable | ✅ | ✅ | All | [SETUP.md](weaviate/SETUP.md) |
| **Supabase** | 🟡 Alpha | ✅ | ✅ | All | [README.md](supabase/README.md) |
| **MongoDB Atlas** | 🧪 Experimental | ❌ | ✅ | All | [ATLAS_SETUP.md](mongodb/ATLAS_SETUP.md) |
| **Milvus** | 🟢 Beta | ✅ | ✅ | All | [LOCAL_SETUP.md](milvus/LOCAL_SETUP.md), [CLOUD_SETUP.md](milvus/CLOUD_SETUP.md) |
| **Chroma** | ✅ Stable | ✅ | ✅ | macOS only | [SETUP.md](chroma/SETUP.md) |
| **Qdrant** | 🧪 Experimental | ✅ | ✅ | All | [SETUP.md](qdrant/SETUP.md) |
| **Neo4j** | 🧪 Experimental | ✅ | ✅ | All | [README.md](neo4j/README.md) |
| **OpenSearch** | 🧪 Experimental | ✅ | ✅ | All | [README.md](opensearch/README.md) |

## Status Legend

- ✅ **Stable**: Production-ready, fully tested, recommended for all use cases
- 🟢 **Beta**: Feature complete, tested, recommended for development and testing
- 🟡 **Alpha**: Functional, some features may be incomplete, use with caution
- 🧪 **Experimental**: New, limited testing, may have breaking changes

## Feature Comparison

| Feature | Weaviate | Supabase | MongoDB | Milvus | Chroma | Qdrant | Neo4j | OpenSearch |
|---------|----------|----------|---------|--------|--------|--------|-------|------------|
| **Vector Search** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Hybrid Search** | ✅ | ✅ | ❌ | ✅ | ❌ | ✅ | ❌ | ✅ |
| **Image Vectors** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Metadata Filtering** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Batch Operations** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Collection Management** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Auto Embeddings** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |

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
