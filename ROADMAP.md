# Weave CLI Roadmap

**Last Updated**: 2025-12-03
**Current Version**: v0.7.1

## Overview

Weave CLI is a command-line tool for managing vector databases with a unified interface. This roadmap outlines completed milestones and upcoming features.

## Version History

### ✅ v0.7.1 (2025-12-02) - Neo4j Integration

**Status**: Released

**Features:**
- Neo4j local vector database support
- Graph + vector hybrid capabilities
- HNSW vector indexing
- Cypher-based metadata filtering
- Batch document operations (tested with 100+ documents)
- Comprehensive integration tests (4 test suites passing)
- Full documentation and examples

**Integration Status:**
- Neo4j Local: ✅ Production Ready
- Neo4j Cloud (Aura): ⚠️ Experimental (untested)

### ✅ v0.7.0 (2025-11-29) - Qdrant Integration

**Status**: Released

**Features:**
- Qdrant local and cloud support
- gRPC-based high-performance communication
- HNSW vector indexing
- Metadata filtering via payload queries
- Comprehensive integration tests (14 tests passing)

**Integration Status:**
- Qdrant Local: ✅ Production Ready
- Qdrant Cloud: ✅ Production Ready

### ✅ v0.6.0 (2025-11-21) - Chroma Integration

**Status**: Released (macOS only)

**Features:**
- Chroma local and cloud support
- Client-side embedding generation
- Platform restriction: macOS only (SDK limitation)

**Known Issues:**
- 3 integration test failures (under investigation)

### ✅ v0.5.0 (2025-10-15) - Milvus Integration

**Status**: Released

**Features:**
- Milvus local and cloud (Zilliz) support
- Native BM25 and hybrid search
- High-performance vector operations

### ✅ Earlier Versions

- **v0.4.0** - Supabase PGVector (alpha)
- **v0.3.0** - Weaviate + MongoDB Atlas
- **v0.2.0** - Mock VDB for testing
- **v0.1.0** - Initial release

---

## Current Focus (v0.7.2)

### 🔧 Chroma Test Fixes

**Priority**: High
**Status**: In Progress
**Target**: 2025-12-03

**Issues:**
- 3 integration test failures
- Need root cause analysis
- Verify authentication and API compatibility

**See**: `WORK_PLAN-chroma.md`

### 🧪 Neo4j Cloud Support

**Priority**: Medium
**Status**: Experimental
**Target**: TBD

**Tasks:**
- Test with Neo4j Aura instances
- Add cloud configuration examples
- Update integration tests for cloud
- Add warning messages for experimental status

---

## Short-term Roadmap (Q1 2025)

### v0.8.0 - OpenSearch Integration

**Priority**: High (Next VDB)
**Status**: Planned

**Features:**
- OpenSearch k-NN plugin support
- Native vector search with HNSW/IVF algorithms
- Hybrid search (vector + BM25)
- Advanced filtering with query DSL
- Both local and cloud (AWS OpenSearch) support

**Benefits:**
- Open-source and AWS managed options
- Powerful full-text + vector hybrid search
- Rich ecosystem and tooling
- Excellent for production workloads
- Strong community support

### v0.8.1 - Pinecone Integration

**Priority**: High
**Status**: Planned

**Features:**
- Pinecone serverless support
- Metadata filtering
- Namespaces support
- Pod-based and serverless indexes

**Benefits:**
- Fully managed vector database
- Excellent production scalability
- Low-latency queries

### v0.8.2 - Redis Vector Search

**Priority**: Medium
**Status**: Planned

**Features:**
- Redis Stack vector search
- RediSearch integration
- Hybrid search (vector + full-text)
- In-memory performance

**Benefits:**
- Ultra-fast queries (in-memory)
- Existing Redis ecosystem
- Cost-effective for small datasets

---

## Mid-term Roadmap (Q2 2025)

### Enhanced Search Capabilities

- [ ] Hybrid search support across all VDBs
- [ ] Advanced metadata filtering (range queries, arrays)
- [ ] Multi-vector search
- [ ] Reranking strategies

### Performance & Scale

- [ ] Batch operations optimization
- [ ] Connection pooling improvements
- [ ] Query result caching
- [ ] Parallel batch processing (per-VDB optimization)

### Developer Experience

- [ ] Interactive REPL mode
- [ ] Query builder UI
- [ ] Performance profiling tools
- [ ] Migration tools between VDBs

---

## Long-term Vision (2025+)

### Vector Database Coverage

**Tier 1 (Production Ready):**
- ✅ Weaviate
- ✅ Qdrant
- ✅ Milvus
- ✅ Neo4j
- ✅ MongoDB Atlas
- ✅ Supabase PGVector
- 🔧 Chroma (fixing tests)

**Tier 2 (Planned):**
- ⏳ OpenSearch (Next - v0.8.0)
- ⏳ Pinecone
- ⏳ Redis Stack
- ⏳ Elasticsearch
- ⏳ Vespa

**Tier 3 (Consideration):**
- 🤔 Azure Cognitive Search
- 🤔 Google Vertex AI Matching Engine
- 🤔 AWS OpenSearch Serverless
- 🤔 Typesense

### Advanced Features

**Multi-modal Support:**
- Image embeddings and search
- Audio embeddings
- Video embeddings
- Cross-modal search

**RAG Enhancements:**
- Built-in chunking strategies
- Context window optimization
- Citation tracking
- Source attribution

**Enterprise Features:**
- Multi-tenancy support
- RBAC (Role-Based Access Control)
- Audit logging
- Cost tracking per VDB
- SLA monitoring

**AI/ML Integration:**
- Auto-embedding model selection
- Fine-tuning workflows
- Embedding model comparison
- A/B testing framework

---

## Community & Ecosystem

### Documentation

- [x] Quick start guides
- [x] Per-VDB integration guides
- [x] Troubleshooting documentation
- [ ] Video tutorials
- [ ] Best practices guide
- [ ] Migration guides

### Testing & Quality

- [x] Integration test framework
- [x] CI/CD pipeline
- [x] Automated releases
- [ ] Performance benchmarks
- [ ] Load testing suite
- [ ] Chaos engineering tests

### Community

- [ ] Contributing guide enhancements
- [ ] Issue templates
- [ ] PR templates
- [ ] Code of conduct
- [ ] Community forum/discussions

---

## Success Metrics

### Current Status (v0.7.1)

**Vector Databases:**
- Supported: 8 (7 production, 1 experimental)
- Integration test coverage: 90%+
- Documentation coverage: 100%

**Code Quality:**
- Linting: ✅ Passing
- Security scans: ✅ Passing (Go 1.24.11)
- Build matrix: ✅ All platforms

**Performance:**
- Batch operations: Tested with 100+ documents
- Parallel processing: Tested with 3 workers
- Average operation time: <1s per document

### Target Metrics (v1.0)

- 12+ vector databases supported
- <100ms average query latency
- 99.9% test coverage for core operations
- <5 open bugs at any time
- <24h response time for critical issues

---

## How to Contribute

See specific work plans for current tasks:
- `WORK_PLAN-chroma.md` - Fix Chroma integration tests
- `WORK_PLAN-current.md` - Today's priorities

For new features, please:
1. Check this roadmap for planned work
2. Open an issue for discussion
3. Wait for maintainer feedback
4. Submit a PR with tests and docs

---

## Legend

- ✅ Complete and released
- 🔧 In progress
- ⏳ Planned (next up)
- 🤔 Under consideration
- ⚠️ Experimental/untested
- 🧪 Beta/testing phase

---

**Maintained by**: @maximilien
**License**: MIT
**Repository**: https://github.com/maximilien/weave-cli
