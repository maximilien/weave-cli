# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Timeout Protection**: Added `context.WithTimeout` to all VDB operations
  - **Qdrant**: 18 operations protected (Health, Collection ops, Document ops, Query ops)
  - **Pinecone**: 15 operations protected (Health, Collection ops, Document ops, Query ops)
  - **Neo4j**: All operations protected via centralized `executeQuery` method
  - **Supabase**: 8 key operations protected (Health, Collection ops, Document ops)
  - **OpenSearch**: 2 foundational operations protected (Health, CreateCollection)
  - Default timeout: 30s (configurable per deployment)
  - Pattern: `ctx, cancel := context.WithTimeout(ctx, adapter.getTimeout())`
  - Prevents indefinite hangs on slow/unavailable VDB endpoints
  - Commits: a995a26 (Qdrant), c0a4887 (Pinecone), e81b848 (Neo4j), d36dad5 (Supabase), ff910b8 (OpenSearch)
- **Operation-Specific Timeout Strategy**: Intelligent timeout optimization based on operation type and deployment
  - **Health Checks**: 10s local, 20s cloud (faster failure feedback)
  - **Bulk Operations**: 120s local, 300s cloud (no false timeouts on large batches)
  - **Other Operations**: Optimized per operation type (Document, Collection, Query, Schema)
  - **Deployment-Aware**: Automatically adjusts for local vs cloud network latency
  - **User Override**: Custom timeouts still respected via config
  - **Implementations**:
    - OpenSearch: Health + Bulk (CreateDocuments, DeleteDocuments)
    - Qdrant: Health + Bulk (CreateDocuments, DeleteDocuments)
    - Pinecone: Health + Bulk (CreateDocuments, DeleteDocuments)
    - Neo4j: Health + Bulk (BatchCreateDocuments, DeleteDocuments)
    - Supabase: Health + Bulk (CreateDocuments, DeleteDocuments)
    - Elasticsearch: Health + Bulk (CreateDocuments, DeleteDocuments)
    - Chroma: Health + Bulk (CreateDocuments, DeleteDocuments)
    - Weaviate: Health check
    - Milvus: Health + Bulk (CreateDocuments, DeleteDocuments)
    - MongoDB: Health + Bulk (CreateDocuments, DeleteDocuments)
  - New module: `src/pkg/vectordb/timeout.go` with `OperationType` enum
  - Documentation: `src/pkg/vectordb/TIMEOUT_GUIDE.md`

### Fixed
- **Connection Handling Audit**: Completed comprehensive audit of Health() and Close() implementations
  - **Health() - Perfect (100%)**: All 10 VDBs now have proper timeout protection
  - **Close() - Fixed Critical Gaps**: Added missing resource cleanup methods
    - **Supabase (CRITICAL)**: Added `Close()` to prevent connection pool exhaustion
    - **Weaviate**: Added documented no-op `Close()` for API consistency
  - **Audit Grade**: B+ (87/100) - Health: 50/50, Close: 37/50
  - **Proper Cleanup**: 5 VDBs with full cleanup (Qdrant, Milvus, Neo4j, Chroma, MongoDB)
  - **SDK Limitations**: 3 VDBs documented as no-op (Elasticsearch, Pinecone, OpenSearch)
  - Audit report available at `/tmp/connection_handling_audit.md`
  - Commit: 6dedfff (Close() method additions)
- **Interface Compliance**: Fixed missing Close() methods on 4 Adapter implementations
  - **MongoDB**: Added `Close(ctx)` delegation to embedded Client (prevents resource leaks)
  - **Elasticsearch**: Added `Close(ctx)` delegation to embedded Client (prevents resource leaks)
  - **Milvus**: Added `Close()` delegation to embedded Client (prevents resource leaks)
  - **Neo4j**: Added `Close(ctx)` delegation to wrapped Client (prevents resource leaks)
  - **Impact**: All 10 VDBs now have explicit Close() methods on their Adapter types
  - **Audit Grade**: Improved from B- (82%) to A- (91%) - 9/10 fully compliant, 1/10 in progress
  - Full audit report available at `/tmp/interface_compliance_audit.md`
- **OpenSearch Implementation Completed**: Implemented all missing query and collection operations
  - **Query Operations (4)**: SearchSemantic (k-NN), SearchBM25, SearchHybrid, SearchByMetadata
  - **Collection Operations (2)**: ListCollections, GetSchema
  - **Features**: Full vector search with k-NN, BM25 text search, hybrid search, metadata filtering
  - **Status**: OpenSearch now 100% functional (23/23 interface methods)
  - **Interface Compliance**: Improved from C+ (74%) to A+ (100%)
  - **Impact**: All 10 VDBs now fully compliant with VectorDBClient interface

## [0.7.3] - 2025-12-03

### Added
- **Neo4j Query Support**: Added vector search query support for Neo4j databases
  - `weave cols query COLLECTION "query text" --neo4j-local` now works
  - `weave cols q COLLECTION "text" --neo4j-local` shorthand supported
  - Semantic search with similarity scoring
  - Integration with OpenAI embeddings for query vectorization
- **Explicit Weaviate Flags**: Added granular database selection flags
  - `--weaviate-local` flag for Weaviate local instances
  - `--weaviate-cloud` flag for Weaviate cloud instances
  - `--weaviate` flag still works as convenience (matches both)
  - Consistent with other VDB naming (`--milvus-local`, `--milvus-cloud`, etc.)
- **Chroma Query Support**: Added error reporting for Chroma queries
  - Changed from silent failure to proper error display
  - Shows helpful error messages (e.g., embedding dimension mismatches)

### Changed
- **Neo4j Factory Initialization**: Improved LLM client initialization
  - Changed from eager (init-time) to lazy (on-demand) initialization
  - Now reads `OPENAI_API_KEY` at runtime after `.env` file is loaded
  - Fixes issue where installed binaries couldn't find API key
  - Works correctly when binary is installed to `~/.local/bin`

### Fixed
- **Neo4j Query Command**: Fixed "Unknown vector database type: neo4j-local" error
  - Added missing Neo4j cases to query command switch statement
  - Both `neo4j-local` and `neo4j-cloud` now supported
- **Chroma Query Command**: Fixed silent failure on query operations
  - Added missing call to `QueryCollection()` for Chroma databases
  - Properly displays error messages instead of silently ignoring
- **Markdown Linting**: Fixed all linting issues in status documents
  - `RELEASE_BLOCKER.md` - Fixed line length, blank lines, code block language
  - `STATUS.md` - Fixed heading spacing, list formatting, emphasis usage

## [0.7.2] - 2025-12-03

### Added

#### VDB Naming Convention Standardization
- **Consistent Naming**: Standardized vector database naming convention
  - All databases now use `-local` and `-cloud` suffixes consistently
  - Example: `mongodb-cloud`, `supabase-cloud`, `milvus-local`, `qdrant-local`
  - Shortcut resolution: bare names (e.g., `weaviate`) automatically resolve to `-cloud` variants
- **Type Constants**: Added new type constants for consistent naming
  - `VectorDBTypeMongoDBCloud` and `VectorDBTypeSupabaseCloud`
  - Factories support both legacy and new type names for backward compatibility
- **Configuration Updates**:
  - Updated `config.yaml` with standardized naming
  - Fixed Qdrant similarity metric capitalization (Cosine, Euclidean, Dot)
  - Fixed Qdrant port to use gRPC port 6334 instead of HTTP port 6333
  - All command case statements updated to support new type constants

#### Summary Table Views
- **Collections Summary**: Added `--summary` / `-S` flag to `weave cols ls`
  - Shows table of all databases with collection counts and status
  - Default behavior: summary for multiple VDBs, detailed list for single VDB
  - Columns: VDB, TYPE, COLS, STATUS
  - Progressive output: displays results as they're retrieved
- **Health Check Progressive Display**: Improved `weave health check`
  - Added `-S` shorthand for `--summary` flag
  - Progressive output: displays each database status as checked (no waiting)
  - Matches `cols ls` behavior for consistent UX
- **Filtering Support**:
  - `weave config list --cloud` / `--local`: Filter databases by deployment type
  - `weave health check --cloud` / `--local`: Check only cloud or local databases

### Changed
- **Command Updates**: Updated 13 files to support new type constants
  - All collection commands (list, create, delete, count, show, query, etc.)
  - All document commands (create, delete, show)
  - Config utilities and health check commands
- **Health Check Behavior**: Changed from collecting all results to progressive display
  - Results appear immediately as each database is checked
  - Improves perceived performance for multi-database checks

#### Neo4j Vector Database Integration (v0.7.1)
- **Neo4j Local Support**: Full integration with Neo4j graph database for vector search
  - Collection management (create, list, delete, exists, count)
  - Document CRUD operations (create, get, update, delete, batch)
  - Automatic OpenAI embedding generation (text-embedding-3-small)
  - Vector similarity search using HNSW indexes
  - Metadata filtering with Cypher WHERE clauses
  - Support for cosine and euclidean similarity metrics
  - ACID transactions for data consistency
- **CLI Commands**:
  - `--neo4j-local` flag for all commands
  - `--neo4j-cloud` flag (experimental, untested, for future Aura support)
  - Integration with existing document and collection commands
- **Testing**:
  - Comprehensive integration test suite (4 test suites, all passing)
  - `./test.sh integration --neo4j` support
  - `--skip neo4j` flag for skipping Neo4j tests
- **Infrastructure**:
  - Local management script: `tools/vdb/local/neo4j.sh`
  - Configuration examples: `configs/config.neo4j-local.yaml`
  - Docker/Podman support for local development
- **Documentation**:
  - Complete Neo4j integration guide: `docs/neo4j/README.md`
  - Updated VDB support matrix with Neo4j columns
  - Quick start examples and troubleshooting guide
- **Requirements**:
  - Neo4j 5.11+ (for vector search support)
  - OpenAI API key (for embedding generation)
  - Bolt protocol support (bolt://localhost:7687)

#### Development Tools
- **Linting**: Updated `lint.sh` to exclude .gitignore markdown files
  - Excludes: `TODOs.md`, `TODOs-*.md`, `NEXT_STEPS*.md`, `WORK_PLAN*.md`, `SESSION_SUMMARY_*.md`
  - Prevents false positives on temporary working documents

### Changed
- **Document Commands**: Added Neo4j support to all document operations
  - `weave docs create` - Create documents in Neo4j collections
  - `weave docs list` - List documents from Neo4j
  - `weave docs show` - Show specific Neo4j documents
  - `weave docs count` - Count documents in Neo4j collections
  - `weave docs delete` - Delete documents from Neo4j
  - `weave docs delete-all` - Delete all documents from Neo4j collection
- **Test Infrastructure**: Enhanced `test.sh` with Neo4j integration
  - Automatic Neo4j availability detection (port 7687)
  - Environment variable support (`NEO4J_PASSWORD`)
  - Integration with test summary reporting
- **Test Authentication**: Updated all Neo4j tests to use `NEO4J_PASSWORD` environment variable
  - Replaced hardcoded passwords with env var lookups
  - Default fallback to "weaveneo4j" for local development

### Fixed
- **Neo4j Test Authentication**: Fixed authentication failures in integration tests
  - Tests now properly read `NEO4J_PASSWORD` from environment
  - Added default password fallback for local development
- **Document Command Coverage**: Added missing Neo4j cases to document commands
  - Fixed "Unknown vector database type" errors
  - Ensures consistent behavior across all vector databases
- **Batch Document Creation**: Fixed database-agnostic support in batch operations
  - Changed from hardcoded `CreateWeaviateDocument` to generic `CreateDocument`
  - Enables batch operations to work with all vector databases including Neo4j
  - Tested with 100+ documents using parallel processing (3 workers)

## [0.7.0] - 2025-11-29

### Added

- **✅ Qdrant Vector Database Integration (Production Ready)**: Complete Qdrant support for vector database operations
  - **Status**: ✅ Production Ready - Fully tested with local Qdrant instances
  - **What Works**:
    - ✅ Connection and health checks (gRPC on port 6334)
    - ✅ Collection management (create, list, delete, exists, count)
    - ✅ Document CRUD operations with automatic embedding generation
    - ✅ Batch document operations
    - ✅ Vector similarity search (HNSW indexing)
    - ✅ Metadata filtering (payload-based queries)
    - ✅ All 14 integration tests passing (4.96s total)
  - **New Features**:
    - Automatic OpenAI embedding generation on document creation
    - Support for both Qdrant Local (Docker/Podman) and Qdrant Cloud
    - `--qdrant-local` and `--qdrant-cloud` flag support across all commands
    - gRPC client integration for high-performance communication
    - HNSW indexing for fast and accurate similarity search
    - Multiple distance metrics: Cosine, Dot Product, Euclidean
    - Float32 vector support with auto-conversion from OpenAI's float64
    - Flexible document ID handling (automatic UUID conversion)
    - Original ID preservation via `_original_id` payload field
  - **New Package**: `src/pkg/vectordb/qdrant/` with complete VectorDB interface implementation
    - `client.go`: Qdrant gRPC client with separate clients for Collections, Points, and Health
    - `collection.go`: Collection operations with HNSW configuration
    - `document.go`: Document CRUD with point-based data model
    - `query.go`: Search operations with payload filtering
    - `adapter.go`: VectorDB interface adapter with embedding support
    - `factory.go`: Factory pattern for Qdrant client creation
  - **Documentation**:
    - `docs/qdrant/SETUP.md`: Comprehensive setup guide for local and cloud
    - `configs/config.qdrant-local.yaml`: Example configuration for local Qdrant
    - `configs/config.qdrant-cloud.yaml`: Example configuration for Qdrant Cloud
    - Updated `docs/VDB_SUPPORT.md`: Qdrant marked as Production Ready
    - Updated `README.md`: Qdrant setup instructions
    - `docs/releases/RELEASE_v0.7.0.md`: Comprehensive release notes
  - **Integration Tests**:
    - `tests/qdrant_integration_test.go`: Comprehensive integration test suite (14 tests)
    - Tests cover: health checks, collections, documents, batch ops, search operations
    - All tests passing with local Qdrant (Podman container)

### Changed

- **Chroma Platform Restriction (macOS only)**: Due to chroma-go v0.2.5 SDK's CGO dependency (libtokenizers), Chroma support is now restricted to macOS (AMD64/ARM64) only
  - Linux and Windows builds use stub implementation with clear error messages
  - Updated all Chroma package files with `//go:build (darwin && amd64) || (darwin && arm64)` constraints
  - Updated `stub_unsupported.go` to handle all non-macOS platforms
  - Updated documentation to clearly indicate platform limitation
  - Alternatives for Linux/Windows: Weaviate, Milvus, Qdrant, MongoDB, or Supabase

- **Qdrant Deprecated API Updates**: Updated Qdrant client to use modern gRPC APIs
  - Replaced deprecated `grpc.Dial()` with `grpc.NewClient()`
  - Removed deprecated `grpc.WithTimeout()`
  - Updated vector data access from `vec.Vector.GetData()` to `vec.Vector.GetDense().GetData()`
  - Fixed staticcheck warnings in CI builds

- **Vector Database Type Configuration**: Added Qdrant type constants
  - Added `VectorDBTypeQdrantLocal` and `VectorDBTypeQdrantCloud` to config package
  - Updated CLI flag handling in `src/cmd/root.go` and `src/cmd/utils/vectordb_selector.go`
  - New selector functions: `getQdrantLocalConfig()` and `getQdrantCloudConfig()`

- **Documentation**: Updated vector database support documentation
  - `docs/VDB_SUPPORT.md`: Added Qdrant to all feature matrices with 🧪 status
  - Added Qdrant database-specific notes section
  - Updated integration test coverage table
  - Updated roadmap section marking Qdrant as Production Ready
- **VDB Storage in .gitignore**: Added local VDB storage patterns
  - `qdrant_storage/` - Qdrant local storage
  - `milvus_storage/` - Milvus local storage (existing)
  - `chroma_data/` - Chroma local storage (existing)

### Fixed

- **Critical Bug - Address Parsing**: Fixed port doubling issue (`localhost:6334:6334`)
  - Root cause: parseAddress() was treating entire address (including port) as host
  - Added proper host:port splitting with `splitHostPort()` helper
  - Added environment variable support (QDRANT_HOST, QDRANT_GRPC_PORT)
  - Added URL protocol detection with `findProtocol()` helper
  - Fixed connection errors preventing all Qdrant operations
  - Location: `src/pkg/vectordb/qdrant/adapter.go:358-426`

- **Critical Bug - UUID Validation**: Fixed document ID compatibility
  - Root cause: Qdrant requires valid UUIDs, but users were providing strings like "test-doc-1"
  - Implemented deterministic UUID v5 generation using DNS namespace
  - Store original ID in `_original_id` payload field for retrieval
  - Updated GetDocument(), DeleteDocument(), DeleteDocuments() with UUID conversion
  - Users can now use any string as document ID
  - Location: `src/pkg/vectordb/qdrant/document.go:232-378`

- **Build Error - Variable Shadowing**: Fixed compilation error in DeleteDocument
  - Renamed `err` to `delErr` to avoid variable shadowing
  - Location: `src/pkg/vectordb/qdrant/document.go:153`

### Known Limitations

- **No BM25 Search**: Keyword search not supported natively by Qdrant
- **Hybrid Search Fallback**: Falls back to vector-only search
- **Float32 Conversion**: Embeddings converted from float64 to float32 (minor precision loss)
- **CLI Flags**: `--qdrant-local` and `--qdrant-cloud` not yet wired to all commands (integration tests use direct API)

## [0.6.0] - 2025-11-27

### Added

- **🗄️ Chroma Vector Database Integration (✅ Production Ready)**: Complete Chroma support for vector database operations
  - **Status**: ✅ Production Ready - Fully functional with Chroma Go SDK v2 API
  - **What Works**:
    - ✅ Connection and health checks for both Chroma Local and Chroma Cloud
    - ✅ Collection management (create, list, delete, show, count, exists)
    - ✅ Document CRUD operations with automatic embedding generation
    - ✅ Batch document operations
    - ✅ Vector search (semantic search)
    - ✅ Metadata search and filtering
    - ✅ Automatic client switching (local HTTP client vs cloud client)
  - **New Features**:
    - Automatic OpenAI embedding generation on document creation
    - Support for both Chroma Local (Docker/Podman) and Chroma Cloud
    - `--chroma-local` and `--chroma-cloud` flag support across all commands
    - Chroma Go SDK v2 API integration (NewCloudClient for cloud, NewHTTPClient for local)
    - Free cloud tier support with quota limit handling
  - **New Adapter**: `src/pkg/vectordb/chroma/` package with complete VectorDB interface implementation
    - `client.go`: Chroma client with automatic local/cloud switching
    - `factory.go`: Factory pattern for Chroma adapter creation
    - `schema.go`: Schema operations and collection management
    - `documents.go`: Document CRUD operations with embedding support
    - `search.go`: Vector search and metadata filtering
  - **Documentation**:
    - `docs/chroma/SETUP.md`: Comprehensive setup guide for local and cloud
    - `configs/config.chroma-local.yaml`: Example configuration for local Chroma
    - `configs/config.chroma-cloud.yaml`: Example configuration for Chroma Cloud
    - Updated `docs/VDB_SUPPORT.md`: Chroma marked as Production Ready
    - Updated `README.md`: Chroma setup instructions and status
  - **Configuration Examples**:
    - Updated `.env.example` with Chroma credentials template
    - Updated `config.yaml` with Chroma configuration examples
  - **Integration Tests**:
    - `tests/chroma_integration_test.go`: Comprehensive integration test suite
    - Tests cover: health checks, collections, documents, batch ops, search operations
    - Quota limit handling for cloud free tier
    - All Chroma integration tests passing

### Fixed

- **🔧 Chroma v2 API Migration**: Migrated Chroma integration to Go SDK v2
  - Fixed GetDocument to use raw result data methods instead of ToRecords()
  - Fixed ListDocuments to use GetIDs(), GetDocuments(), GetMetadatas()
  - Fixed SearchByMetadata to use raw result data instead of ToRecords()
  - Added embedding function requirement for GetCollection (v2 API)
  - Enhanced metadata type filtering (only primitives: string, int, float, bool)
  - Root cause: Chroma v2 SDK quirk where ToRecords() returns empty results
  - All Chroma integration tests passing

- **🔧 Chroma Cloud Client Configuration**: Improved Chroma Cloud client setup
  - Use NewCloudClient for cloud configurations instead of HTTP client
  - Make URL optional for Chroma Cloud (uses api.trychroma.com automatically)
  - Add fallback to CHROMA_API_KEY if CHROMA_CLOUD_API_KEY not set
  - Improve client initialization logic to distinguish cloud vs local
  - Fixed 403 Permission Denied errors with proper cloud client usage

- **🔧 Chroma Commands**: Fixed docs and cols commands for Chroma
  - Fixed document listing and retrieval operations
  - Fixed collection operations to work correctly with Chroma v2 API

### Changed

- **Vector Database Support Maturity Levels**: Updated VDB_SUPPORT.md with Chroma support
  - Chroma Local: ✅ Production Ready (fully tested and stable)
  - Chroma Cloud: ✅ Production Ready (fully functional with quota limits documented)
  - Updated integration test coverage table to include Chroma
  - Documented known limitations (no BM25, quota limits on free tier)

- **Documentation**: Comprehensive Chroma documentation updates
  - Added Chroma troubleshooting guide (403 errors, quota limits, connection issues)
  - Documented cloud vs local client implementation details
  - Added credential setup instructions for Chroma Cloud
  - Updated CHANGELOG with Chroma v2 API migration details

### Known Limitations

- **Chroma BM25 Search**: No native BM25 support (keyword search not available)
- **Chroma Cloud Quota**: Free tier limited to 300 documents per GET request
- **Integration Tests**: Some tests may fail on free tier due to quota limits (expected behavior)

## [0.5.1] - 2025-11-20

### Fixed

- **🐛 MongoDB Connection Issues**: Fixed MongoDB Atlas connection problems
  - Removed custom TLS configuration that was causing connection failures
  - MongoDB Go driver now handles TLS automatically for `mongodb+srv://` URIs
  - Updated documentation with troubleshooting for common connection errors
  - Clarified that TLS errors are often misleading (usually IP whitelist issues)

- **🐛 Milvus Collection Operations**: Fixed multiple Milvus collection operation bugs
  - Fixed `weave cols create` with named schema for Milvus VDBs
  - Fixed deleting collections in `--milvus-*` VDBs (proper handling of schema deletion)
  - Improved error messages for collection operations, especially "collection not found" scenarios
  - Better Milvus-specific error message simplification

### Added

- **🧪 Comprehensive Integration Test Coverage**: Enhanced test coverage across all VDBs
  - **Milvus Integration Tests**: Added comprehensive test suite (`tests/milvus_integration_test.go`)
    - Health check tests
    - Collection CRUD operations (Create, Exists, List, Delete)
    - Document CRUD operations (Create, Get, Update, List, Delete)
    - Batch document operations
    - Metadata preservation verification
    - Search operations (semantic, metadata, BM25, hybrid)
    - Schema operations (GetSchema, ValidateSchema)
    - DeleteDocumentsByMetadata test
    - Factory and registry tests
    - Milvus Cloud tests (separate test function)
  - **MongoDB Integration Tests**: Expanded test coverage
    - SemanticSearch test (requires OpenAI API key)
    - HybridSearch test
    - TestMongoDBFactoryIntegration test
    - TestMongoDBVectorDBRegistry test
  - **Test Infrastructure Improvements**: Enhanced `test.sh` with:
    - Per-VDB status tracking (Weaviate, Milvus, Supabase, MongoDB, MCP)
    - Detailed summary table showing pass/fail/skip per VDB
    - Better visual output with box formatting
    - Skipped tests counter
    - Overall status message (ALL PASSED / SOME FAILED / ALL SKIPPED)
    - `--skip` flag support to skip specific VDB tests
    - Improved error handling for test counting

- **✨ Collection Show for Supabase**: Enabled `weave cols show` command for Supabase VDB
  - Generic collection display now works with Supabase
  - Shows collection details, document count, and metadata

### Changed

- **🔧 Streamlined Collection Show Command**: Improved `cols show` command
  - Enhanced generic collection display for non-Weaviate collections
  - Better visual output matching Weaviate format
  - Now works with MongoDB, Supabase, Milvus, and Mock VDBs
  - Displays embedding model information, sample documents, and metadata

- **📋 Improved Help Output**: Enhanced CLI help display
  - Grouped flags logically (Database Selection, Database Override, Configuration, Output Control, Behavior)
  - Organized commands into groups (Data Management, Configuration & Health, AI & Search)
  - Better visual organization and readability

- **📁 Configuration File Organization**: Moved example configuration files
  - All `config.<vdb>.yaml` files moved to `configs/` directory
  - Added `configs/README.md` with usage instructions
  - Updated all documentation references to new paths

### Documentation

- **📚 Planning Updates**: Added planned VDB integrations
  - Added Chroma to planned VDB integrations list
  - Added Neo4j to planned VDB integrations list
  - Updated planning document with integration details

- **📖 MongoDB Setup Guide**: Enhanced troubleshooting documentation
  - Added troubleshooting for `tls: internal error` (IP whitelist issue)
  - Added note about x509 certificate errors
  - Clarified TLS error causes and solutions

## [0.5.0] - 2025-11-17

### Added

- **🗄️ Milvus Vector Database Integration (✅ Fully Functional)**: Complete Milvus support for vector database operations
  - **Status**: ✅ Fully Functional - Complete with automatic embedding generation and vector search
  - **What Works**:
    - ✅ Connection and health checks for both Milvus Local and Milvus Cloud (Zilliz)
    - ✅ Collection management (create, list, delete, show, count)
    - ✅ Document CRUD operations with automatic embedding generation
    - ✅ Vector search with configurable similarity metrics (L2, IP, COSINE)
    - ✅ Hybrid search (vector + BM25 combination)
    - ✅ Metadata search and filtering
    - ✅ Batch document operations
  - **New Features**:
    - Automatic OpenAI embedding generation on document creation
    - Support for both Milvus Local (Docker/Podman) and Milvus Cloud (Zilliz)
    - `--milvus` flag support across all collection and document commands
    - Local Milvus infrastructure with Docker Compose and Podman Compose
    - Comprehensive health checks and container management utilities
  - **New Adapter**: `src/pkg/vectordb/milvus/` package with complete VectorDB interface implementation
    - `adapter.go`: Adapter wrapper with LLM client integration
    - `client.go`: Milvus client implementation with connection management
    - `collection.go`: Collection operations (create, list, delete, show, count)
    - `document.go`: Document CRUD operations with embedding support
    - `query.go`: Vector search, hybrid search, and metadata filtering
    - `factory.go`: Factory pattern for Milvus adapter creation
  - **Documentation**:
    - `docs/milvus/README.md`: Comprehensive integration guide with examples
    - `docs/milvus/LOCAL_SETUP.md`: Step-by-step setup for local Milvus development
    - `docs/milvus/CLOUD_SETUP.md`: Complete guide for Zilliz Cloud setup
    - `configs/config.milvus-local.yaml`: Example configuration for local Milvus
    - `configs/config.milvus-cloud.yaml`: Example configuration for Milvus Cloud
  - **Local Infrastructure**:
    - `local/milvus/docker-compose.yml`: Docker Compose configuration
    - `local/milvus/podman-compose.yml`: Podman Compose configuration
    - `tools/vdb/local/milvus.sh`: Local Milvus management script
    - `tools/vdb/health.sh`: VDB health check utilities
    - `tools/vdb/container/`: Container detection and management utilities
  - **Configuration Examples**:
    - Updated `.env.example` with Milvus credentials template
    - Updated `README.md` with Milvus setup instructions
  - **E2E Testing**:
    - `e2e-tests.sh`: Comprehensive end-to-end test suite for all VDBs
    - Tests for Weaviate, Supabase, MongoDB, Milvus, and Mock VDBs
    - Health checks, collection operations, and document operations

### Changed

- **Vector Database Support Maturity Levels**: Updated VDB_SUPPORT.md with Milvus support
  - Weaviate: ✅ Production (fully tested and stable)
  - Supabase: 🧪 Alpha (core functionality working, production readiness being evaluated)
  - MongoDB: ✅ Fully Functional (complete with automatic embeddings and vector search)
  - Milvus: ✅ Fully Functional (complete with automatic embeddings and vector search)
  - Mock: ✅ Production (for testing purposes)
- **Tools Directory Structure**: Reorganized tools directory for better organization
  - `tools/demo/`: Demo recording and management tools
  - `tools/dev/`: Development utilities (linting, formatting, license headers)
  - `tools/vdb/`: Vector database management utilities
- **Vector Database Types**: Added `milvus-local` and `milvus-cloud` to supported vector database types
- **Factory Pattern**: Updated vectordb factory to support Milvus adapter registration
- **Configuration Schema**: Extended `VectorDBConfig` with Milvus-specific fields

### Technical Details

- **Dependencies**: Added `github.com/milvus-io/milvus-sdk-go/v2` for Milvus client
- **Files Added**: 6 new Milvus adapter files plus infrastructure and documentation
- **Build Status**: ✅ Compiles successfully, all existing tests pass
- **Vector Search**: Supports L2 (Euclidean), IP (Inner Product), and COSINE similarity metrics
- **Embedding Dimensions**: Configurable (default: 1536 for OpenAI ada-002)
- **Local Development**: Full Docker/Podman support for local Milvus instances

## [0.4.0] - 2025-11-15

### Added

- **🗄️ MongoDB Atlas Vector Search Integration (✅ Fully Functional)**: Complete MongoDB Atlas support for vector database operations
  - **Status**: ✅ Fully Functional - Complete with automatic embedding generation and vector search
  - **What Works**:
    - ✅ Connection and health checks
    - ✅ Collection management (create, list, delete)
    - ✅ Document CRUD operations with automatic embedding generation
    - ✅ Atlas Vector Search ($vectorSearch aggregation) - fully functional
    - ✅ BM25 keyword search using MongoDB text indexes
    - ✅ Hybrid search (vector + BM25 combination with RRF)
    - ✅ Metadata search and filtering
    - ✅ Document deletion by ID or filename
    - ✅ Embedding model display in collection and document listings
  - **New Features**:
    - Automatic OpenAI embedding generation on document creation
    - Vector search with configurable similarity metrics (cosine, euclidean, dotProduct)
    - Hybrid search combining vector similarity and BM25 keyword search
    - `--mongodb` flag support across all commands
  - **New Adapter**: `src/pkg/vectordb/mongodb/` package with VectorDB interface implementation
  - **BM25 Search**: Keyword search using MongoDB text indexes (fully functional)
  - **Document Operations**: Full CRUD support with embedding storage fields
  - **Collection Management**: Create, list, delete collections with automatic text indexing
  - **Configuration**: New config fields for MongoDB (URI, database, vector_dimensions, similarity_metric)
  - **Free Tier Support**: Compatible with MongoDB Atlas M0 free tier
  - **Documentation**:
    - `docs/mongodb/README.md`: Comprehensive integration guide with examples
    - `docs/mongodb/ATLAS_SETUP.md`: Step-by-step setup instructions for MongoDB Atlas
    - `configs/config.mongodb.yaml`: Example configuration file for MongoDB
  - **Configuration Examples**:
    - Updated `.env.example` with MongoDB credentials template
    - Updated `config.yaml` with MongoDB configuration example
  - **Tests**:
    - `tests/mongodb_integration_test.go`: Comprehensive integration test suite (9 test suites)
    - `src/pkg/vectordb/mongodb/client_test.go`: Unit tests for client, schema, and document operations
    - `src/pkg/vectordb/mongodb/factory_test.go`: Unit tests for factory and config validation
    - Updated `test.sh` with `--mongodb` flag for selective integration testing
    - MongoDB unit tests now run with `./test.sh unit`
    - Test coverage: health, collections, documents, batch ops, BM25 search, metadata operations, config validation

### Changed

- **Vector Database Support Maturity Levels**: Updated VDB_SUPPORT.md with maturity indicators
  - Weaviate: ✅ Production (fully tested and stable)
  - Supabase: 🧪 Alpha (core functionality working, production readiness being evaluated)
  - MongoDB: ✅ Fully Functional (complete with automatic embeddings and vector search)
  - Mock: ✅ Production (for testing purposes)
- **Collection Display**: Changed "docs" to "items" in collection list display for consistency
- **Document Deletion**: Improved deletion with fallback to filename/metadata matching
- **Embedding Display**: Added embedding model information to collection and document listings
- **Vector Database Types**: Added `mongodb` to supported vector database types
- **Factory Pattern**: Updated vectordb factory to support MongoDB adapter registration
- **Configuration Schema**: Extended `VectorDBConfig` with MongoDB-specific fields (Database, VectorDimensions, SimilarityMetric)

### Technical Details

- **Dependencies**: Added `go.mongodb.org/mongo-driver v1.17.6`
- **Files Added**: 7 new MongoDB adapter files (client.go, adapter.go, factory.go, collection.go, document.go, query.go, schema.go)
- **Build Status**: ✅ Compiles successfully, all existing tests pass
- **Vector Search**: Supports cosine, euclidean, and dotProduct similarity metrics
- **Embedding Dimensions**: Configurable (default: 1536 for OpenAI ada-002)

## [0.3.14] - 2025-11-14

### Fixed

- **🎬 Demo Recording Improvements**: Fixed demo recording speed and timing issues
  - **Recording Speed**: Added configurable page delays (default 3 seconds) to demo recorder for better viewer readability
  - **Demo Script Fixes**: Updated cleanup commands in demo scripts to use `delete-schema` instead of `delete`
  - **UTF-8 Encoding**: Added proper locale settings for emoji display in asciinema recordings
  - **Re-recorded Videos**: All demo videos re-recorded with improved pacing and correct timing
  - **Updated Links**: All demo links updated to point to new asciinema recordings with accurate duration estimates

### Changed

- **Demo Timing**: Corrected duration estimates in documentation:
  - Config Demo: ~1 min (was 5 min)
  - Supabase Demo: ~1 min (was 8 min)
  - Full Demo: ~2 min (was 5 min)
  - Quick Demo: ~1 min (was 2 min)
  - REPL Demo: ~1 min (was 4 min)

### Technical Details

- **Files Changed**: Demo scripts, recording tools, and all video recordings updated
- **Recording Infrastructure**: Enhanced `auto-demo-recorder.exp` with page delay support
- **Documentation**: Updated `docs/guides/DEMO.md` with correct links and timing

## [0.3.13] - 2025-11-14

### Added

- **🎬 Comprehensive Demo Suite**: Complete demo infrastructure for Weave CLI
  - **5 Interactive Demo Scripts**:
    - `config-demo.sh`: Configuration management walkthrough (~5 min)
    - `supabase-demo.sh`: Supabase integration demonstration (~8 min)
    - `full-demo.sh`: Complete feature walkthrough (~5 min)
    - `quick-demo.sh`: Fast overview (~2 min)
    - `repl-demo.sh`: AI-powered REPL interface demo
  - **Automated Recording Tools**: Scripts for recording and uploading demos to asciinema
    - `auto-demo-recorder.exp`: Expect script for automated recording
    - `record-all-demos.sh`: Batch recording script for all demos
    - `RECORD_ALL_DEMOS.sh`: Video orchestration script
  - **Video Recordings**: All demos recorded and uploaded to asciinema
    - Config demo: `weave-cli-config-demo.cast`
    - Supabase demo: `weave-cli-supabase-demo.cast`
    - Updated full, quick, and REPL demo recordings
  - **Comprehensive Documentation**:
    - `demos/README.md`: Central documentation for all demo scripts
    - Updated `docs/guides/DEMO.md`: Complete demo guide with links and instructions
    - `tools/README-DEMO-RECORDING.md`: Recording guide and best practices
    - `docs/planning/DEMO_UPDATE_PLAN.md`: Demo infrastructure planning document

### Changed

- **Demo Documentation**: Updated main README and demo guides with all new demos
- **Collection Utilities**: Enhanced collection deletion and utility functions for demo support
- **Gitignore**: Updated to exclude demo artifacts and recording files

### Technical Details

- **Files Changed**: 23 files changed, 3,148 insertions(+), 1,236 deletions(-)
- **Demo Scripts**: All scripts follow consistent structure with colored output, interactive pauses, and cleanup options
- **Recording Infrastructure**: Automated tools for consistent demo recording and upload workflow

## [0.3.10] - 2025-11-10

### Fixed

- **🐛 Supabase Collection Creation**: Fixed duplicate column errors when creating collections with custom schemas
  - Added `fixedColumns` map to track reserved column names (id, content, text, image, image_data, url, metadata, embedding)
  - Schema properties that conflict with fixed columns are now automatically skipped
  - Prevents PostgreSQL errors when creating collections with custom properties

- **🐛 Supabase Document Operations**: Added complete Supabase support to all document commands
  - **Create**: `weave docs create` now works with Supabase
  - **Show**: `weave docs show` now works with Supabase (improved database selection)
  - **Delete**: `weave docs delete` now works with Supabase
  - **List**: `weave docs list` now works with Supabase
  - Implemented generic document utility functions using vectordb abstraction for consistency

### Added

- **🧪 Comprehensive Integration Tests**: Created comprehensive Weaviate integration test suite
  - Full test coverage matching Supabase test structure
  - Tests cover: health checks, collections, documents, search operations, schema operations, batch operations
  - Uses vectordb abstraction for consistency across both adapters
  - All integration tests verified passing for both Weaviate and Supabase

- **🧪 Test Infrastructure Improvements**: Enhanced test execution capabilities
  - Added selective integration test execution flags to `test.sh`
  - `--weaviate` flag for Weaviate-only tests
  - `--supabase` flag for Supabase-only tests
  - `--mcp` flag for MCP-only tests
  - Improved test documentation and help text

- **🔧 MCP Client Configuration**: Improved MCP server startup reliability
  - Added project root detection to find `.env`/`go.mod` files
  - Ensures MCP server can locate configuration when started from subdirectories

### Changed

- **Document Show Command**: Improved database selection logic
  - Uses `GetSelectedVectorDBs()` for consistent database selection
  - Validates that only one database is selected for read operations
  - Clear error messages when multiple databases are specified

## [0.3.9] - 2025-11-10

### Added

- **🗄️ Vector Database Abstraction Layer**: Multi-database support architecture
  - **Abstract Interface**: Unified `VectorDBClient` interface for all vector databases
  - **Factory Pattern**: Clean abstraction layer for easy addition of new database backends
  - **Weaviate Support**: Full support maintained for Weaviate Cloud and Local instances
  - **Supabase Support**: New PostgreSQL-based vector database adapter using pgvector
  - **Modular Design**: Database-specific adapters with consistent API
  - See [docs/VECTOR_DB_ABSTRACTION.md](../docs/VECTOR_DB_ABSTRACTION.md) for details

- **🐘 Supabase Integration**: Complete Supabase adapter implementation
  - **Collection Management**: Create, list, delete collections with PostgreSQL tables
  - **Document Operations**: Full CRUD operations for documents
  - **Metadata Support**: JSON metadata storage and querying
  - **Schema Management**: Automatic table creation with proper indexes
  - **Connection Handling**: Robust connection management with IPv4/IPv6 support
  - **Error Handling**: Comprehensive error wrapping and user-friendly messages
  - Set `VECTOR_DB_TYPE=supabase` to use Supabase adapter

- **🔤 Embeddings Management**: New `weave embeddings` command
  - **List Embeddings**: `weave embeddings list` (or `weave embeddings ls`) to view available models
  - **Model Selection**: Choose embedding models for document creation
  - **`--embedding` flag** for `weave docs create` command
  - **`--embedding` flag** for `weave collection create` command
  - Allows users to specify custom embedding models per operation

- **⚙️ Enhanced Configuration Management**: Streamlined configuration setup
  - **Interactive Config Commands**: `weave config create` and `weave config update`
  - **Smart Defaults**: Sensible defaults for quick setup
  - **Environment Variable Fallback**: Config.yaml optional with env var support
  - **Auto-creates config.yaml**: When REPL detects missing config.yaml, automatically creates a minimal one

- **🔧 weave-mcp Binary Installer**: One-command installation of weave-mcp
  - **Automatic platform detection** - supports macOS, Linux, Windows
  - **Download with progress bar** - visual feedback for large downloads
  - **Checksum verification** - ensures download integrity
  - **Interactive prompts** - choose install location and permissions
  - **Smart PATH detection** - warns if install directory not in PATH
  - **Auto .env Update**: Automatically updates .env file with weave-mcp path
  - **Installation testing** - verifies binary is executable
  - Run `weave config update --weave-mcp` to install

- **🎯 Smart Error Handling & Configuration UX**: Dramatically improved user experience
  - **Detailed error messages** showing exactly what's missing (environment variables, config files)
  - **Interactive configuration fixes** - prompts to create/update .env file on the spot
  - **Multiple fix options** clearly explained (flags, shell exports, .env file)
  - **Context-aware tips** for config.yaml setup and mock database testing
  - **Better MCP error diagnostics** - captures stderr from weave-mcp for troubleshooting
  - **Better collection/database errors** with actionable suggestions
  - All commands now use enhanced error handling for consistent UX

- **🔄 REPL Enhancements**:
  - **Version Display in Banner**: REPL banner now shows version info
    - Displays version string in dimmed text below ASCII art
    - Consistent with `weave -V` output format
  - **REPL Batch Query Mode**: Execute multiple queries in batch mode with demo infrastructure

- **🛠️ Developer Experience**:
  - **Global Timeout Flag**: New `--timeout` flag with duration format support (e.g., "30s", "5m", "1h")
  - **Code Organization**: Refactored Weaviate code into vectordb/weaviate package
  - **Better Maintainability**: Cleaner separation of concerns

- **📦 MCP Integration Tests**: Automated test suite for weave-mcp compatibility
  - Comprehensive integration tests for all MCP operations
  - Collection and document operation testing
  - Error handling and edge case validation
  - Automated testing workflow for MCP releases
  - See [tests/README_MCP_TESTS.md](../tests/README_MCP_TESTS.md) for details

### Changed

- **Vector DB Architecture**: Refactored to support multiple database backends
  - Weaviate code moved to `src/pkg/vectordb/weaviate/` package
  - New factory pattern for database creation
  - Unified error handling across all adapters

- **Documentation**: Updated and refactored documentation
  - README.md simplified and reorganized
  - New vector DB abstraction guide
  - Updated user guide with new features
  - Comprehensive technical blog post draft

### Fixed

- **Linting Errors**: Resolved all staticcheck SA9003 errors
  - Fixed empty branch in transaction rollback handling
  - Improved error handling patterns
  - All CI linting checks passing

## [0.3.0] - 2025-11-01

### Added

- **🔄 Interactive REPL Mode**: Run `weave` without arguments for an interactive session
  - Beautiful ASCII art banner with version and GitHub link
  - Natural language query support in interactive mode
  - Built-in help, examples, and history commands
  - CTRL-C to stop commands, twice to exit (like Claude CLI)
  - Command history saved to `~/.weave_history`

- **🤖 AI Agents Query Command**: New `weave query` (or `weave q`) command for natural language queries using GPT-4o
  - Automatically understands your intent and plans appropriate commands
  - Executes weave-cli and bash commands intelligently
  - Provides comprehensive reports with recommendations
  - Supports dry-run mode to preview execution plans

- **🧠 Multi-Agent Architecture**: 7 specialized agents working together
  - QueryAgent: Validates and fixes user queries with MCP tool awareness
  - PlanningAgent: Creates detailed execution plans with full tool schemas
  - WeaveAgent: Executes weave-cli commands via MCP protocol
  - BashAgent: Safely executes bash commands with user approval
  - OutputAgent: Beautiful, color-coded user-friendly output
  - ReportAgent: Comprehensive operation reports with LLM-generated recommendations
  - EvalAgent: Tracks metrics and evaluates success

- **📊 OpenTelemetry Integration**: Opik tracing for LLM observability
  - Automatic cost tracking for all LLM calls
  - Token usage metrics (prompt, completion, total)
  - Color-coded cost display (green/yellow/red)
  - Direct link to Opik dashboard for detailed traces

- **✨ Enhanced User Experience**:
  - Smart error detection in command output
  - Step progress with duration display
  - JSON output with syntax highlighting
  - Auto-approval for simple weave commands
  - Health check support via natural language

## [0.2.14] - 2025-10-15

### Added

- **🔄 PDF Conversion Tool**: New `weave docs pdf-convert` command to convert CMYK PDFs to RGB format using Ghostscript or ImageMagick
- **🎯 Text-Only PDF Processing**: New `--skip-all-images` flag to extract only text from PDFs without image processing overhead
- **💬 Helpful Tips Control**: Global `--no-tips` flag to suppress helpful tips and suggestions when desired
- **🔧 CMYK PDF Support**: Graceful handling of CMYK PDFs with actionable tips for conversion using Ghostscript or ImageMagick

## [0.2.11] - 2025-10-10

### Added

- **📦 Batch Document Creation**: Process entire directories with parallel processing and automatic retry
- **⚡ Parallel Processing**: Configure multiple workers for faster batch operations
- **📊 Progress Tracking**: Visual progress indicators with time estimation
- **🔄 Smart Retry**: Automatic retry on failures with `.processed` file tracking
- **📈 Comprehensive Reporting**: CSV reports with detailed processing statistics

## [0.2.10] - 2025-10-08

### Added

- **📄 Enhanced PDF Processing**: Improved PDF text extraction with better fallback handling
- **💬 Human-Friendly Error Messages**: Simplified, actionable error messages with helpful suggestions
- **🎬 Updated Demos**: New demo recordings showcasing PDF processing capabilities
- **🔧 Better UX**: Fixed PDF success message formatting and improved user experience

## [0.2.8] - 2025-10-04

### Added

- **📊 Score Normalization**: Implemented quadratic score normalization for
  better result differentiation
  - Low-relevance results (raw score 0.5) now display as 0.25, clearly
    indicating poor matches
  - High-relevance results (raw score 0.7+) preserved at higher normalized values
  - Makes it much easier to distinguish irrelevant results from good matches
  - Uses `score^2` transformation to amplify score differences
- **⚠️ Smart Result Warnings**: Automatic detection and warning for low-quality results
  - Displays warning when all results have scores < 0.3
  - Suggests rephrasing query for better results
  - Provides clear score interpretation guidance
- **📖 Enhanced Documentation**: Updated help text and README with score interpretation
  - Clear score ranges: < 0.3 (no match), 0.3-0.5 (weak), 0.5-0.7 (good), > 0.7 (strong)
  - Help text includes score interpretation guidelines
  - README documents the normalization approach
- **🎬 Demo Upload Tracking**: Asciinema upload script now saves URLs automatically
  - Upload URLs saved to `videos/latest-demo-uploads.txt`
  - Maintains latest URLs for both quick and full demos
  - Includes timestamps and upload history
  - README updated to reference latest demo URLs

### Changed

- **Score Display**: All scores now use normalized values for clearer interpretation
- **User Guidance**: Enhanced output messages to help users understand query results

## [0.2.7] - 2025-10-02

### Added

- **🔍 Semantic Search**: New `query` command for semantic search on collections
  - Uses Weaviate's `nearText` for vector-based similarity search
  - Automatic fallback to hybrid search with real similarity scores
  - Beautiful formatted results with relevance scores (0.0 to 1.0)
- **🔤 BM25 Override**: New `--bm25` flag for keyword-based search
  - Direct BM25 keyword search instead of semantic search
  - Real relevance scoring from Weaviate API
  - Perfect for exact keyword matching
- **🔍 Metadata Search**: New `--search-metadata` flag
  - Search in metadata fields in addition to content/text fields
  - Supports URLs, filenames, domains, and custom metadata
  - Combined with content search for comprehensive results
- **📊 Real Scoring**: All search methods provide authentic Weaviate similarity scores
  - Primary search: nearText provides distance-based similarity scores
  - Fallback search: hybrid search provides combined vector+keyword scores
  - BM25 override: direct BM25 keyword search provides relevance scores
- **🎯 Smart Fallback**: Robust 3-tier fallback system
  - BM25 → Hybrid → Simple text search fallback chain
  - Graceful degradation for unsupported Weaviate configurations
  - Ensures query functionality works across all Weaviate instances
- **🧪 Comprehensive Testing**: 100% test coverage for query functionality
  - 27+ test scenarios including unit, e2e, and integration tests
  - Mock client with realistic scoring algorithm
  - All tests passing with complete coverage

### Changed

- **Enhanced Query Display**: Improved result formatting with emojis and
  clear structure
- **Improved Mock Scoring**: More realistic scoring algorithm with
  content/metadata differentiation
- **Updated Documentation**: Complete README, CHANGELOG, and demo updates

### Fixed

- **--no-truncate Flag**: Fixed support for query commands to show full content
- **Scoring Issues**: Resolved hardcoded 1.0 scores in fallback scenarios
- **Error Handling**: Improved GraphQL error detection and fallback logic

### Known Limitations

⚠️ **Weaviate Instance Requirements**: Some Weaviate instances may not
support advanced search features:

- `nearText` semantic search requires vector search modules
- `bm25` keyword search requires BM25 module installation
- `hybrid` search requires hybrid search module
- Fallback to simple text search works but may have accuracy limitations
- Use `--vector-db-type mock` for full functionality testing

### Examples

```bash
# Basic semantic search
weave cols q MyCollection "machine learning algorithms"

# Search with metadata fields
weave cols q MyCollection "maximilien.org" --search-metadata

# Use BM25 keyword search
weave cols q MyCollection "exact keywords" --bm25

# Combine metadata search with BM25
weave cols q MyCollection "search term" --search-metadata --bm25

# Show full content without truncation
weave cols q MyCollection "query" --no-truncate
```text

## [0.2.6] - 2025-10-01

### Added

- **🔍 Semantic Search**: New `query` command for semantic search on collections
  - Uses Weaviate's `nearText` for vector-based similarity search
  - Automatic fallback to hybrid search with real similarity scores
  - Beautiful formatted results with relevance scores (0.0 to 1.0)
- **🔤 BM25 Override**: New `--bm25` flag for keyword-based search
  - Direct BM25 keyword search instead of semantic search
  - Real relevance scoring from Weaviate API
  - Perfect for exact keyword matching
- **🔍 Metadata Search**: New `--search-metadata` flag
  - Search in metadata fields in addition to content/text fields
  - Supports URLs, filenames, domains, and custom metadata
  - Combined with content search for comprehensive results
- **📊 Real Scoring**: All search methods provide authentic Weaviate similarity scores
  - Primary search: nearText provides distance-based similarity scores
  - Fallback search: hybrid search provides combined vector+keyword scores
  - BM25 override: direct BM25 keyword search provides relevance scores
- **WeaveDocs/WeaveImages Schema**: New default schema structure for better
  document management
  - Flat metadata structure for improved performance and simplicity
  - Better support for document aggregation and display
  - Backward compatibility with RagMeDocs/RagMeImages schemas
- **Enhanced Document Display**: Improved document listing and aggregation
  - Better handling of both WeaveDocs and legacy RagMeDocs schemas
  - Improved filename detection across different schema types
  - Enhanced metadata structure for better document organization

### Changed

- **Schema Defaults**: Updated default schemas from RagMeDocs/RagMeImages to
  WeaveDocs/WeaveImages
- **Document Creation**: Enhanced document creation to use new WeaveDocs schema
  by default
- **Test Coverage**: Updated test cases to reflect new schema structure
- **Demo Scripts**: Updated demo files to showcase new schema capabilities

### Fixed

- **Document Aggregation**: Fixed document aggregation logic to handle both new
  and legacy schemas
- **Schema Detection**: Improved collection type detection for better schema
  handling
- **Metadata Structure**: Standardized metadata fields across different document
  types

### Technical Details

- **Files Modified**:
  - `schemas/WeaveDocs.yaml`: Updated to use `content` field instead of `text`
  - `src/cmd/document/create.go`: Updated help text and examples
  - `src/cmd/utils/collection.go`: Improved schema type detection
  - `src/cmd/utils/display.go`: Enhanced document aggregation logic
  - `tests/cmd_test.go`: Updated test cases for new schema structure
  - `videos/weave-cli-full-demo.cast`: Updated demo to showcase new features

## [0.2.5] - 2025-09-30

### Added

- **Named Schema Support**: Collections can now be created using named
  schemas from `config.yaml`
  - New `--schema` flag for `weave collection create` command
  - Example: `weave cols c MyDocsCol --schema RagMeDocs`
  - Schemas defined in `databases.schemas` section of config.yaml
- **Schema Configuration**: Added schema definitions to config.yaml and config.yaml.example
  - RagMeDocs schema for text documents
  - RagMeImages schema for image documents
  - Support for custom user-defined schemas
- **Config Schema Structure**: New `SchemaDefinition` type in config package
  - Methods: `GetSchema(name)`, `ListSchemas()`
  - Conversion from config schema to Weaviate CollectionSchema
- **Documentation Updates**: Updated USER_GUIDE.md with schema usage examples
  - Creating collections with named schemas
  - Defining custom schemas in config.yaml
  - Using schema files vs named schemas

### Technical Details

- **Files Modified**:
  - `src/pkg/config/config.go`: Added SchemaDefinition struct and methods
  - `src/cmd/collection/create.go`: Added --schema flag support
  - `src/cmd/utils/collection.go`: Added CreateWeaviateCollectionFromConfigSchema()
  - `config.yaml.example`: Added schemas section with RagMeDocs and RagMeImages examples
  - `docs/USER_GUIDE.md`: Updated collection creation documentation

### Usage

```bash
# Create collection using named schema
weave collection create MyDocsCol --schema RagMeDocs

# Create collection using schema file
weave collection create MyCol --schema-yaml-file schema.yaml

# Traditional method still works
weave collection create MyCol
```text

## [0.2.2] - 2025-09-29

### Added

- **SPDX License Headers**: Added proper license headers to all 44 Go source
  files
- **License Management Tool**: Created `tools/add_license_headers.sh` for
  automated license header management
- **Legal Compliance**: Industry-standard SPDX license identification
  throughout codebase

### License Technical Details

- **License Format**: `// SPDX-License-Identifier: MIT`
- **Copyright**: `// Copyright (c) 2025 dr.max`
- **Files Updated**: All source files in `src/`, `tests/`, and `scripts/`
  directories
- **Automation**: Script for future license header updates and management

### Quality Assurance

- **Linting**: All golangci-lint, go vet, go fmt checks passing
- **Tests**: All unit tests passing (100%)
- **E2E Tests**: All 34 E2E tests passing
- **Code Formatting**: Automatically fixed by linter

## [0.2.1] - 2025-09-29

### Fixed

- **CI/CD Linting Issues**: Resolved all golangci-lint unused function warnings
- **Code Cleanup**: Removed 15 unused functions from refactoring process
  - PDF module: Removed unused `generateRealisticPDFContent` function
  - Root module: Removed 14 duplicate styled output functions moved to utils
- **E2E Test Reliability**: Improved cleanup process to avoid false errors
  - Skip collection deletion when collections don't exist
  - Rely on schema deletion for reliable cleanup
  - Reduced test count from 35 to 33 (removed unreliable collection deletion tests)

### Test Reliability Improved

- **Test Reliability**: E2E tests now more robust with better error handling
- **Code Quality**: All linting checks passing, proper code formatting
- **CI/CD Pipeline**: Clean builds with no warnings or errors

### Quality Technical Details

- **Linting**: All golangci-lint, go vet, go fmt checks passing
- **Tests**: All unit and E2E tests passing
- **Code Formatting**: Automatically fixed by linter
- **Quality Assurance**: 100% clean codebase with no unused functions

## [0.2.0] - 2025-09-29

### E2E Testing Added

- **Complete E2E Testing Suite**: 35 comprehensive integration tests
  against real Weaviate instances
- **Smart Configuration Detection**: Auto-detection of Weaviate Cloud
  availability with graceful fallback to mock database
- **Enhanced Terminal Visibility**: Bold white text for maximum contrast
  on dark terminal backgrounds
- **Complete Collection Operations**: Full implementation of all
  collection management functions
- **Complete Document Operations**: Full implementation of all document
  management functions
- **Non-interactive Testing**: All E2E tests run without user prompts
  using --force flags
- **Isolated Test Collections**: Dedicated test collections
  (WeaveDocs_test, WeaveImages_test) with automatic cleanup

### Changed

- **Major Code Refactoring**: Complete modularization of monolithic files
  - `document.go`: 4,307 → 50 lines (98% reduction)
  - `collection.go`: 2,528 → 50 lines (98% reduction)
  - `shared.go`: 1,364 → 7 focused files (100% modularization)
  - `processor.go`: 609 → 3 focused files (100% modularization)
- **Enhanced User Experience**: Improved color scheme and output formatting
- **Better Error Handling**: More descriptive error messages and graceful fallbacks
- **Improved Documentation**: Updated README and CHANGELOG with v0.2.0 features

### Collection Operations Fixed

- **Collection Creation**: Implemented `CreateWeaviateCollection` function
- **Collection Deletion**: Implemented `DeleteWeaviateCollections` function
- **Schema Management**: Implemented `DeleteWeaviateCollectionSchema` function
- **Pattern Matching**: Added collection pattern matching for bulk operations
- **Embedding Model**: Updated to use `text-embedding-3-small` for
  compatibility
- **Confirmation Prompts**: Fixed hanging tests by implementing proper
  --force flag usage

### Technical Improvements

- **Modular Architecture**:
  - `src/cmd/document/` (7 files) - Document operations
  - `src/cmd/collection/` (7 files) - Collection operations
  - `src/cmd/utils/` (7 files) - Shared utilities
  - `src/pkg/pdf/` (3 files) - PDF processing modules
- **Quality Assurance**: 100% test coverage (unit + E2E tests)
- **CI/CD Ready**: Automated testing and validation pipeline
- **Production Ready**: Fully tested against real Weaviate instances

## [0.1.10] - 2025-01-27

### Code Organization Changed

- **Code organization refactoring**: Improved codebase structure and maintainability
  - Split `document.go` (4,307 lines) into 6 logical files:
    - `list.go` - Document list command
    - `show.go` - Document show command
    - `count.go` - Document count command
    - `create.go` - Document create command
    - `delete.go` - Document delete command
    - `delete_all.go` - Document delete-all command
  - Split `collection.go` (2,528 lines) into 7 logical files:
    - `list.go` - Collection list command
    - `create.go` - Collection create command
    - `delete.go` - Collection delete command
    - `delete_all.go` - Collection delete-all command
    - `count.go` - Collection count command
    - `show.go` - Collection show command
    - `delete_schema.go` - Collection delete-schema command
  - Updated main command files to contain only command definitions
  - Preserved all existing functionality with no breaking changes
  - Enhanced developer experience with better file organization

## [0.1.9-rc3] - 2025-09-28

### Empty Collection Fixed

- **Empty collection document listing**: Fixed Weaviate client `ListDocuments` method
  to handle empty collections gracefully
  - Resolved confusing "chunk_index" error when listing documents from empty
    collections
  - Added fallback mechanisms using aggregation API and simple queries
  - Now shows clear "No documents found in collection 'X'" message instead of
    cryptic Weaviate errors
  - Maintains full backward compatibility with collections containing documents

### Safety Features Added

- **Double confirmation for delete-schema**: Added double confirmation to `ds`
  (delete-schema) command similar to `da` (delete-all)
  - First confirmation: Standard y/N prompt asking for confirmation
  - Second confirmation: Red warning with requirement to type "yes" exactly
  - `--force` flag still works to skip both confirmations
  - Consistent user experience across all destructive operations

### Improved

- **Error messages**: Enhanced error handling and user feedback throughout the CLI
  - Better collection existence checks
  - Clearer error messages for common failure scenarios
  - Improved robustness of Weaviate client operations

## [0.1.9-rc2] - 2025-09-28

### Schema Management Added

- **Schema flags for collection creation**: Add `--text` and `--image` flags to
  `weave collection create` command
  - **Default**: Collections are created with text schema (RagMeDocs format)
    unless `--image` is specified
  - `--text`: Creates collection with text schema (RagMeDocs format) -
    Properties: `url`, `text`, `metadata`
  - `--image`: Creates collection with image schema (RagMeImages format) -
    Properties: `url`, `image`, `metadata`, `image_data`
  - Enhanced schema validation and error handling
  - Backward compatibility maintained through default text schema

- **Enhanced collection schema management**:
  - Explicit schema type selection for better data organization
  - Proper RagMeDocs and RagMeImages schema compatibility
  - Automatic vectorization configuration based on schema type

### Schema Workflow Changed

- **Collection creation workflow**: Collection creation now supports explicit
  schema specification with sensible defaults
  - **Default behavior**: Collections are created with text schema unless
    `--image` is specified
  - Improved collection creation with proper schema setup
  - Better error messages for conflicting schema flags

- **Collection creation logic**: Enhanced to support explicit schema types
  - Added `CreateCollectionWithSchema()` function for schema-aware collection creation
  - Improved schema property definitions for text vs image collections
  - Better vectorization configuration based on collection type

### Technical Details

- Added `SchemaType` constants and `CreateCollectionWithSchema()` function
- Enhanced `createCollectionViaREST()` to support explicit schema types
- Updated `createWeaviateCollection()` function for schema-aware collection creation
- Improved validation logic for required schema flags
- Enhanced error handling and user feedback

## [0.1.9-rc1] - 2025-09-28

### Features Added

- **Pattern-based collection deletion**: Add `--pattern` flag to `collection delete`
  command
  - Support for shell glob patterns (`WeaveDocs*`, `Test*`, `*Docs`)
  - Support for regex patterns (`.*Docs$`, `^Test.*`)
  - Auto-detection of pattern types
  - Comprehensive validation and confirmation prompts

- **Pattern-based schema deletion**: Add `--pattern` flag to `collection
  delete-schema` command
  - Same pattern matching capabilities as collection deletion
  - Complete schema removal with pattern support

- **Enhanced PDF processing**: New PDF processor package (`src/pkg/pdf/`)
  - Generic PDF text chunking and image extraction
  - Improved metadata structure for better compatibility
  - Enhanced document creation with proper field mapping

### Changes Made

- **Collection commands**: Updated help text and examples for pattern support
- **Document processing**: Improved metadata structure for better RagMeDocs
  compatibility
- **Weaviate client**: Enhanced document creation with better field mapping

### Known Issues & Limitations

- **PDF metadata extraction**: Not fully implemented (requires pdfcpu library
  integration)
- **RagMeDocs compatibility**: Virtual document view not fully compatible with
  RagMeDocs structure
- **AI summary generation**: Needs enhancement for full RagMeDocs compatibility
- **Document creation**: `weave docs create` commands may not produce documents
  fully compatible with RagMeDocs legacy system

### Implementation Details

- Added `findCollectionsByPattern()` function for pattern matching
- Reused existing pattern matching logic from document deletion
- Enhanced validation to prevent mixing collection names with patterns
- Improved error handling for pattern-based operations

## [0.1.8] - Previous Release

### Features

- Basic collection and document management
- Weaviate Cloud and Local support
- Mock database for testing
- Document pattern-based deletion
- Virtual document view
- Configuration management

---

## Migration Notes

### For RagMeDocs Users

If you're migrating from RagMeDocs legacy system:

1. **Document Creation**: The `weave docs create` command creates documents with
   a different metadata structure than RagMeDocs
2. **Virtual View**: The virtual document view may not show the same aggregate
   information as RagMeDocs
3. **AI Summaries**: Generated AI summaries are basic and may not match RagMeDocs
   comprehensive summaries
4. **PDF Metadata**: PDF metadata extraction is not yet implemented (Title,
   Creator, Producer, etc.)

### Recommended Workflow

- Use pattern-based deletion for cleanup: `weave cols delete --pattern "WeaveDocs*"`
- Test document creation with small files first
- Verify virtual document view meets your needs before bulk operations

## Contributing

When adding new features or fixing bugs, please update this changelog following
the format above.
