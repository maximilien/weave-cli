# Changelog

All notable changes to Weave CLI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.7.6] - 2025-12-10

### Added
- **Pinecone Support** - Full integration with Pinecone serverless vector database
  - Automatic embedding generation using OpenAI
  - All CRUD operations (create, read, update, delete) for collections and documents
  - Semantic search with vector similarity
  - Metadata filtering support
  - Batch document operations
  - Comprehensive integration tests
  - Complete documentation in VDB_SUPPORT.md

### Changed
- **PRESENTATION.md** - Updated and moved from archive to main docs
  - Now reflects v0.7.x features with all 9 vector databases
  - Updated database support section with complete list
  - Refreshed roadmap removing completed features
  - Updated examples and feature lists

### Technical Details
- Added Pinecone Go SDK v1.1.1 dependency
- Implemented structpb.Struct for Pinecone metadata handling
- Serverless index creation with sensible defaults (1536 dims, cosine metric)
- Integration with OpenAI text-embedding-3-small for embeddings
- Upsert-based update pattern following Pinecone best practices

## [0.7.5] - 2024-12-05

### Summary
Thursday Demo Release - 8 Working Vector Databases

### Added
- **Neo4j Integration** - Core client and collection operations
  - Vector index creation and management
  - Document CRUD operations with embeddings
  - Semantic search with vector similarity
  - Graph-based relationships with vector search

### Fixed
- Removed deprecated +build tags from Chroma files
- Disabled CGO for Linux/Windows release builds for better portability
- Added skip_qdrant usage to test.sh script

### Testing
- All integration tests passing for 8 vector databases
- 7/8 VDBs fully operational (Chroma quota limits expected)
- Comprehensive test coverage for Neo4j operations

## [0.7.4] - Earlier Releases

### Added
- Support for 8 vector databases (Weaviate, Qdrant, Chroma, Milvus, Neo4j, Supabase, MongoDB, Mock)
- Unified VectorDBClient interface across all databases
- Database factory pattern for easy switching
- Environment-based configuration
- Comprehensive integration test suite

### Features
- Collection management (create, list, delete, count)
- Document operations (CRUD, batch, metadata filtering)
- Search capabilities (semantic, BM25, hybrid, metadata)
- Cross-database compatibility layer
- Automatic embedding generation
- Schema validation and management

---

## Version History

- **v0.7.6** - Pinecone Support + Documentation Updates (2025-12-10)
- **v0.7.5** - Thursday Demo Release - 8 Working VDBs (2024-12-05)
- **v0.7.4** - Multi-VDB Support Foundation
- **v0.7.0** - Neo4j Integration
- **v0.6.8** - Supabase Support
- **v0.6.5** - Milvus Cloud & Local
- **v0.6.0** - Qdrant Integration
- **v0.5.5** - Chroma Support
- **v0.5.0** - Multi-VDB Architecture
- **v0.3.x** - AI Agents & Interactive REPL
- **v0.2.x** - Weaviate-only CLI Tool

---

## Supported Vector Databases

| Database | Status | Since Version | Type |
|----------|--------|---------------|------|
| Weaviate | ✅ Production | v0.1.0 | Cloud/Local |
| Pinecone | 🧪 Beta | v0.7.6 | Cloud |
| Qdrant | ✅ Production | v0.6.0 | Cloud/Local |
| Chroma | ✅ Production | v0.5.5 | Cloud/Local |
| Milvus | ✅ Production | v0.6.5 | Cloud/Local |
| Neo4j | 🧪 Beta | v0.7.0 | Cloud/Local |
| Supabase | ✅ Production | v0.6.8 | Cloud |
| MongoDB | ✅ Production | v0.7.2 | Cloud |
| Mock | ✅ Testing | v0.4.0 | In-Memory |

---

## Migration Guide

### Upgrading to v0.7.6

**New Pinecone Support:**
```bash
# Set environment variables
export VECTOR_DB_TYPE=pinecone
export PINECONE_API_KEY=your-api-key
export OPENAI_API_KEY=your-openai-key  # Required for embeddings

# Use standard commands
weave collection list
weave docs create MyCollection document.txt
weave collection query MyCollection "search query"
```

**Updated Documentation:**
- `docs/PRESENTATION.md` - Now in main docs (moved from archive)
- `docs/VDB_SUPPORT.md` - Added Pinecone feature matrix
- See [VDB_SUPPORT.md](docs/VDB_SUPPORT.md) for complete feature comparison

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct and the process for submitting pull requests.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
