# Release v0.9.1: Embedding Dimension Verification

## 🎉 Release Summary

**Version**: v0.9.1
**Date**: January 12, 2026
**Status**: ✅ Ready for Release
**Type**: Bug Fix Release - Vector Database Query Reliability

## ✨ Major Features

### 1. Embedding Dimension Verification

- **Automatic Dimension Checking**: All vector databases now verify embedding dimensions before queries
  - Retrieves actual collection/index dimensions from metadata at query time
  - Verifies query embedding dimensions match collection dimensions
  - Prevents cryptic dimension mismatch errors during searches
  - **Impact**: Eliminates common query failures when using different embedding models
  - **Location**: All VDB adapters in `src/pkg/vectordb/*/`

- **Helpful Error Messages**: Clear guidance when dimension mismatches occur
  - Explains the root cause (different embedding models)
  - Provides three actionable solutions:
    1. Use the same embedding model that created the index
    2. Recreate the index with current embedding model
    3. Configure dimensions in .env file
  - Shows actual vs expected dimensions in error message
  - **Impact**: Users can quickly diagnose and fix dimension issues

- **Graceful Fallback**: Smart dimension detection with fallback strategy
  - First: retrieves dimensions from collection/index metadata
  - Second: falls back to configured `VECTOR_DIMENSIONS` in .env
  - Third: uses sensible defaults (1536 for OpenAI, 384 for sentence-transformers)
  - **Impact**: Works reliably even when metadata unavailable

### 2. Multi-VDB Agent Support

- **Universal Agent Compatibility**: Agents now work with all 10+ vector databases
  - Generic `QueryCollectionWithAgent()` function for all VDBs
  - Supports: Weaviate, Chroma, Qdrant, Milvus, Neo4j, Supabase, MongoDB, Pinecone, Elasticsearch, OpenSearch
  - Maintains backward compatibility with existing Weaviate queries
  - Same agent behavior and output across all databases
  - **Location**: `src/pkg/vectordb/client.go`, agent implementations
  - **Impact**: RAG agents work consistently across all vector databases

### 3. Query Progress Indicator

- **Real-time Progress Updates**: Visual feedback during long-running queries
  - Added `--progress` flag to `weave cols query` command
  - Shows search, agent processing, and response generation phases
  - JSON output mode: progress as JSON Lines format (one object per line)
  - Text output mode: progress messages to stderr for clean piping
  - Works seamlessly with `--json`, `--output`, and `--verbose` flags
  - **Location**: `src/cmd/collection/query.go`
  - **Impact**: Better user experience for agent-based queries

## 🐛 Bug Fixes

### Vector Database Dimension Mismatch Errors

All vector databases now handle embedding dimensions correctly:

- **Qdrant**: Retrieves vector dimensions from collection config
  - Uses `DescribeCollection()` to get vector params
  - **Files**: `src/pkg/vectordb/qdrant/collection.go`, `src/pkg/vectordb/qdrant/adapter.go`

- **MongoDB**: Stores and retrieves embedding metadata
  - Creates `_weave_metadata` document in each collection
  - Stores dimensions and similarity metric at creation time
  - **Files**: `src/pkg/vectordb/mongodb/collection.go`, `src/pkg/vectordb/mongodb/query.go`

- **Supabase**: Creates metadata tracking table
  - `weave_collection_metadata` table stores collection dimensions
  - Tracks embedding model configuration per collection
  - **Files**: `src/pkg/vectordb/supabase/collections.go`, `src/pkg/vectordb/supabase/queries.go`

- **Pinecone**: Retrieves dimensions from index description
  - Uses `DescribeIndex()` API to get vector dimensions
  - Fixed SDK enum compatibility (removed DotProduct support)
  - **Files**: `src/pkg/vectordb/pinecone/collection.go`, `src/pkg/vectordb/pinecone/query.go`

- **Neo4j**: Extracts dimensions from vector index options
  - Parses `vector.dimensions` from index properties
  - Reads dimensions from graph database metadata
  - **Files**: `src/pkg/vectordb/neo4j/collection.go`, `src/pkg/vectordb/neo4j/adapter.go`

- **Elasticsearch**: Retrieves dimensions from index mappings
  - Reads `dense_vector` field configuration
  - Parses dimensions from index mapping properties
  - Fixed TypeMapping struct nil check issue
  - **Files**: `src/pkg/vectordb/elasticsearch/collection.go`, `src/pkg/vectordb/elasticsearch/query.go`

- **OpenSearch**: Retrieves dimensions from index mappings
  - Uses `Indices.Get()` API to retrieve full index info
  - Unmarshals JSON mappings to extract `knn_vector` dimensions
  - **Files**: `src/pkg/vectordb/opensearch/collection.go`, `src/pkg/vectordb/opensearch/query.go`

## 📋 Technical Details

### Implementation Pattern

All VDB adapters now follow a consistent pattern:

```go
// 1. Retrieve collection's actual dimensions
indexDims, err := getIndexDimensions(ctx, collectionName)
if err != nil {
    return nil, fmt.Errorf("failed to get index dimensions: %w", err)
}

// 2. Generate query embedding
embedding, err := llmClient.GenerateEmbedding(ctx, query, "")
if err != nil {
    return nil, fmt.Errorf("failed to generate embedding: %w", err)
}

// 3. Verify dimensions match
if len(embedding) != indexDims {
    return nil, fmt.Errorf(
        "embedding dimension mismatch: index has %d dimensions but "+
        "query embedding has %d dimensions\n\n"+
        "This usually means the index was created with a different embedding model.\n"+
        "Solutions:\n"+
        "  1. Use the same embedding model that was used to create the index\n"+
        "  2. Recreate the index with the current embedding model configuration\n"+
        "  3. Configure VECTOR_DIMENSIONS=%d in your .env file",
        indexDims, len(embedding), indexDims)
}
```

### Configuration Variables

Each VDB supports dimension configuration via environment variables:

- `WEAVIATE_VECTOR_DIMENSIONS`
- `QDRANT_VECTOR_DIMENSIONS`
- `MILVUS_VECTOR_DIMENSIONS`
- `CHROMA_VECTOR_DIMENSIONS`
- `MONGODB_VECTOR_DIMENSIONS`
- `SUPABASE_VECTOR_DIMENSIONS`
- `PINECONE_VECTOR_DIMENSIONS`
- `NEO4J_VECTOR_DIMENSIONS`
- `ELASTICSEARCH_VECTOR_DIMENSIONS`
- `OPENSEARCH_VECTOR_DIMENSIONS`

## 🚀 Upgrade Guide

### For Existing Users

No breaking changes! This release adds automatic dimension verification without requiring any configuration changes.

**Recommended Actions**:

1. **Update to v0.9.1**: Pull latest changes and rebuild
   ```bash
   git pull origin main
   ./build.sh
   ```

2. **Test Your Queries**: Existing queries continue to work
   ```bash
   weave cols query MyCollection "test query"
   ```

3. **If You See Dimension Errors**: Follow the suggested solutions
   - Option 1: Use the original embedding model
   - Option 2: Recreate collections with current model
   - Option 3: Configure dimensions in `.env` file

### For New Users

Dimension verification works automatically! Just ensure your `.env` file has the correct embedding model configuration for your chosen provider.

## 📦 Installation

```bash
# Clone or update repository
git clone https://github.com/maximilien/weave-cli.git
cd weave-cli
git checkout v0.9.1

# Build
./build.sh

# Verify installation
./bin/weave --version
```

## 🔗 Related Issues

- Fixes dimension mismatch errors across all vector databases
- Improves error messages for embedding model mismatches
- Adds agent support for non-Weaviate databases
- Enhances query progress visibility

## 🙏 Credits

This release focused on improving reliability and user experience across all supported vector databases.

---

**Full Changelog**: https://github.com/maximilien/weave-cli/blob/main/CHANGELOG.md
