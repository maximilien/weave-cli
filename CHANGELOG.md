# Changelog

All notable changes to Weave CLI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.8.2] - 2025-12-18

### Summary

UX & Test Quality Improvements - Enhanced error messages with troubleshooting
hints for all 10 VDBs (100% coverage) and strengthened batch document creation
test verification critical for production pipelining workflows.

### Added

- **Troubleshooting Hints** - Extended to 5 additional VDBs (100% coverage)
  - **Milvus**: Connection refused, timeout, and authentication errors
  - **Neo4j**: Bolt protocol errors, Aura cloud issues, credential problems
  - **MongoDB**: Atlas IP whitelist, TLS errors, authentication guidance
  - **Supabase**: PostgreSQL connection issues, password vs API key confusion
  - **OpenSearch**: Resource requirements (2GB+ RAM), AWS-specific guidance
  - Pattern: MongoDB-style error messages with "Common causes:" and action items
  - All 10 VDBs now provide helpful troubleshooting for connection/auth failures

### Changed

- **Batch Create Tests** - Enhanced verification for 3 VDBs
  - **Pinecone**: Added document retrieval verification after 5s indexing wait
  - **Supabase**: Added GetDocument verification for batch-doc-1
  - **Weaviate**: Added GetDocument verification for batch-doc-1
  - Impact: All 10 VDBs now verify batch operations work correctly
  - Critical for pipelining projects requiring batch document ingestion

### Fixed

- **VDB_SUPPORT.md** - Synced database statuses with VDB_SUPPORT_MATRIX.md
  - Updated 7 database statuses to reflect production readiness
  - Qdrant: 🧪 Experimental → ✅ Stable
  - OpenSearch: 🧪 Experimental → ✅ Stable
  - Supabase: 🧪 Alpha → ✅ Stable
  - Milvus: 🧪 Beta → ✅ Stable
  - Added Elasticsearch entries (Beta status)
  - Standardized terminology (Production → Stable)

### Documentation

- **Error Handling** - 100% VDB coverage for troubleshooting hints
  - 10/10 VDBs now have helpful error messages with actionable guidance
  - Docker startup commands included for local deployments
  - Cloud console links for Zilliz, Aura, Atlas, Supabase, AWS
  - Network/firewall troubleshooting steps for common scenarios

- **Test Quality** - Complete batch create verification audit
  - Documented test patterns in NEXT_STEPS.md (Task 4.4)
  - Tier 1 (thorough): Qdrant, MongoDB with individual retrieval
  - Tier 2 (enhanced): All remaining VDBs with count or retrieval checks
  - 100% batch operation test coverage ensures production readiness

### Technical Details

- **Troubleshooting Coverage**:
  - Previously: 5/10 VDBs (Pinecone, Qdrant, Weaviate, Elasticsearch, Chroma)
  - Now: 10/10 VDBs (100% coverage)
  - Pattern: Connection refused, timeout, authentication errors
  - Helpfulness: Docker commands, cloud console links, common fixes

- **Batch Create Test Verification**:
  - 7 VDBs: Already had count-based or retrieval verification
  - 3 VDBs: Enhanced with inline document retrieval (Pinecone, Supabase, Weaviate)
  - All tests verify documents actually created, not just error-free operation
  - Fast inline verification (single retrieval, no loops)

- **Database Status Updates**:
  - ✅ Stable: 7 VDBs (Qdrant, OpenSearch, Supabase, Milvus promoted)
  - 🟢 Beta: 3 VDBs (Pinecone, Elasticsearch, OpenSearch demoted from Experimental)
  - VDB_SUPPORT.md now matches VDB_SUPPORT_MATRIX.md (authoritative source)

## [0.8.0] - 2025-12-16

### Summary

Elasticsearch Beta Release & Test Coverage Completion - All 10 vector
databases now have comprehensive integration tests, and Elasticsearch has
been promoted to Beta status with full documentation.

### Added

- **Elasticsearch Integration** - Full Beta release with complete feature set
  - Dense vector search with HNSW indexing
  - Native BM25 full-text search
  - Hybrid search combining kNN + BM25
  - Metadata filtering with term and range queries
  - Batch operations using BulkIndexer
  - Complete documentation (README, SETUP, LOCAL_SETUP, CLOUD_SETUP)
  - 16/16 integration tests passing

- **OpenSearch Integration Tests** - Comprehensive test suite added
  - 16 test cases covering all operations
  - Health checks, collections, documents, search
  - Semantic, BM25, and hybrid search validation

- **VDB Management Scripts** - Local development tools
  - `tools/vdb/local/elasticsearch.sh` - Elasticsearch container management
  - Scripts for starting, stopping, and monitoring local VDB instances

### Changed

- **Elasticsearch** - Promoted from 🚧 In Progress → 🟢 Beta
  - All 7 phases complete (infrastructure, operations, tests, docs)
  - Ready for development and testing use
  - Updated status in README.md and VDB_SUPPORT_MATRIX.md

- **Test Coverage** - Achieved 100% VDB test coverage
  - All 10 vector databases now have integration tests
  - Standardized 16-test suite across Elasticsearch and OpenSearch
  - Neo4j tests fixed to use factory pattern

### Fixed

- **Neo4j Tests** - Corrected interface implementation
  - Fixed factory pattern instantiation
  - Updated SearchByContent → SearchSemantic
  - Corrected SearchByMetadata signature
  - Removed deprecated Close() method

- **Pinecone Error Messages** - Standardized capitalization
  - Capitalized "Pinecone" in all error messages (17 instances)
  - Consistent with other VDB implementations
  - Updated across adapter, collection, document, factory, and query files

- **README Markdown Linting** - Fixed line length violation
  - Split line 35 to comply with 80-character limit

- **.gitignore Pattern** - Fixed VDB script tracking
  - Changed `local/` → `/local/` to only ignore root-level directory
  - Allows tracking of `tools/vdb/local/*` scripts

### Documentation

- **VDB Support Matrix** - Updated with current status
  - Elasticsearch now listed as 🟢 Beta
  - All test coverage metrics updated to 100%
  - Platform compatibility notes refreshed

- **NEXT_STEPS.md** - Updated with session accomplishments
  - Documented test completion milestone
  - Updated VDB status summaries
  - Marked completed cleanup tasks

### Technical Details

- **Database Status Summary**:
  - ✅ Stable: Weaviate, Qdrant, Milvus, Chroma, Supabase, Neo4j, MongoDB (7)
  - 🟢 Beta: Pinecone, OpenSearch, Elasticsearch (3)
  - Total: 10 production-ready vector databases

- **Test Coverage**:
  - Weaviate: 10/10, Qdrant: 14/14, Milvus: 10/10
  - Chroma: 10/10, Supabase: 10/10, Neo4j: 10/10
  - MongoDB: 10/10, Pinecone: 8/8
  - OpenSearch: 16/16, Elasticsearch: 16/16
  - **Total: 10/10 databases (100% coverage)**

- **Integration Points**:
  - Official go-elasticsearch v9 TypedClient API
  - No CGO dependencies (pure Go)
  - OpenAI integration for automatic embeddings

## [0.7.7] - 2025-12-11

### Summary

VDB Audit & Stability Release - Post-Pinecone cleanup and production readiness improvements.

### Changed

- **Qdrant** - Promoted from 🧪 Experimental to ✅ Stable
  - All 14 integration tests passing
  - Validated HNSW vector search, full CRUD operations
  - Updated status in README.md and VDB_SUPPORT_MATRIX.md

### Added

- **Pinecone** - Added to VDB Support Matrix and test suite
  - Added Pinecone to VDB_SUPPORT_MATRIX.md with feature comparison
  - Added `--pinecone` flag to test.sh for selective testing
  - Documented status as 🧪 Beta, Cloud only

### Fixed

- **Markdown Linting** - Fixed MD060 table formatting in README.md
  - Corrected table separator spacing (compact style)
  - All markdown linting now passing

### Documentation

- **Chroma Platform Limitation** - Emphasized macOS-only constraint
  - Added ⚠️ warnings to Chroma entries in README.md
  - Created Platform Compatibility section
  - Documented alternatives (Weaviate, Milvus, Qdrant, Supabase)
  - Provided CI skip workaround (`--skip chroma`)

- **OpenSearch Memory Requirements** - Documented OOM issue
  - Added system requirements section (2GB min, 4GB recommended, 8GB production)
  - Documented exit code 137 (OOM killed) troubleshooting
  - Provided workarounds for systems with < 2GB RAM
  - Recommended cloud OpenSearch or alternative VDBs for constrained systems

### Technical Details

- Completed comprehensive audit of 9 vector databases
- Addressed all P0 (critical) issues blocking users
- Improved documentation clarity and completeness
- Enhanced platform compatibility documentation

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

- Support for 8 vector databases (Weaviate, Qdrant, Chroma, Milvus,
  Neo4j, Supabase, MongoDB, Mock)
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

- **v0.7.7** - VDB Audit & Stability Release (2025-12-11)
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

See [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct
and the process for submitting pull requests.

## License

This project is licensed under the MIT License - see the
[LICENSE](LICENSE) file for details.
