# Weave CLI Roadmap

**Last Updated**: 2026-02-09
**Current Version**: v0.9.18

## Overview

Weave CLI is a command-line tool for managing vector databases with a
unified interface. This roadmap outlines completed milestones and upcoming
features.

## Version History

### ✅ v0.9.18 (2026-02-09) - Client0 Features

**Status**: Released

**Features:**

- Ollama auto-discovery in agent wizard
- Embedding model comparison report generator
- Data-driven A/B testing for embedding models
- Client0 3-week validation workflow support

**Integration Status:**

- Ollama Client: ✅ Production Ready
- Comparison Reports: ✅ Production Ready

### ✅ v0.9.17 (2026-02-06) - Batch Re-Embedding

**Status**: Released

**Features:**

- Re-embed collections without re-ingestion (20x faster)
- Embedding pipeline with progress tracking
- Support for OpenAI models (OSS providers coming)
- 200+ documents/minute throughput

**Impact:**

- Time savings: 5+ hours → 15 minutes for 3,500 docs
- Perfect for testing different embedding models

### ✅ v0.9.16 (2026-02-05) - Auto-Dimension Detection

**Status**: Released

**Features:**

- Model registry with 17+ embedding models
- Auto-detect dimensions for sentence-transformers, OpenAI, Ollama, Cohere, Voyage AI
- OSS model flagging
- Reduces configuration errors by ~80%

### ✅ v0.9.15 (2026-01) - Production Hardening

**Status**: Released

**Features:**

- VDB lifecycle management improvements
- Enhanced error handling and logging
- Performance optimizations
- Security hardening

### ✅ v0.9.13 (2026-01) - VDB Lifecycle Management

**Status**: Released

**Features:**

- Health check system for all VDBs
- Connection validation and retry logic
- Graceful degradation
- Production monitoring support

### ✅ v0.9.1 (2026-01) - Agent Evaluation System

**Status**: Released

**Features:**

- Pluggable evaluator architecture
- Opik integration for agent evaluation
- Custom evaluator support
- Multi-agent orchestration

### ✅ v0.8.x (2025-12 - 2026-01) - OpenSearch, Pinecone, Elasticsearch

**Status**: Released

**Features:**

- OpenSearch k-NN integration
- Pinecone serverless support
- Elasticsearch vector search
- Hybrid search across multiple VDBs

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

## Current Focus (Q1 2026)

### 🚀 OSS Embedding Providers (v0.9.19)

**Priority**: High
**Status**: In Progress (Week of Feb 10-14, 2026)
**Target**: 2026-02-14

**Goal**: Enable Client0 to test OSS embedding models

**Tasks:**

- ✅ Ollama auto-discovery (shipped in v0.9.18)
- ✅ Comparison report generator (shipped in v0.9.18)
- 🔧 sentence-transformers provider (in progress)
- 🔧 Ollama embedding provider (in progress)
- ⏳ End-to-end testing with OSS models
- ⏳ Comprehensive documentation

**See**: `docs/planning/THIS_WEEK_PLAN.md`

### 🎯 Client0 Support

**Priority**: High
**Status**: Active
**Focus**: 3-week embedding model validation

**Workflow:**

- Week 1: OpenAI baseline testing (supported)
- Week 2: OSS models testing (in progress)
- Week 3: Local Ollama testing (in progress)

**Deliverables:**

- Data-driven model selection
- Cost vs performance comparison
- Production-ready OSS stack

---

## Short-term Roadmap (Q1 2026)

### v0.9.19 - OSS Embedding Providers

**Priority**: High (Current Sprint)
**Status**: In Progress
**Target**: 2026-02-14

**Features:**

- sentence-transformers provider (Python subprocess)
- Ollama embedding provider (HTTP API)
- Complete re-embedding workflow with OSS models
- Performance benchmarks and comparison reports

**Benefits:**

- No vendor lock-in (100% OSS stack possible)
- Cost savings (local/self-hosted models)
- Privacy (on-premise embedding generation)
- Perfect for Client0 validation workflow

### v1.0.0 - Production Ready

**Priority**: High
**Status**: Planned
**Target**: Q1 2026

**Features:**

- Stable API (semver guarantees)
- 12+ vector databases supported
- Comprehensive documentation
- Performance benchmarks
- Migration tools between VDBs

**Quality Gates:**

- <5 open bugs
- 99%+ test coverage
- <100ms average query latency
- Security audit complete

---

## Mid-term Roadmap (Q2 2026)

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

## Long-term Vision (2026+)

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

- ✅ Embedding model comparison (v0.9.18)
- ✅ A/B testing framework (v0.9.18)
- 🔧 OSS embedding providers (v0.9.19)
- Auto-embedding model selection
- Fine-tuning workflows

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

### Current Status (v0.9.18)

**Vector Databases:**

- Supported: 11 (production ready)
- Integration test coverage: 95%+
- Documentation coverage: 100%

**Code Quality:**

- Linting: ✅ Passing
- Security scans: ✅ Passing (Go 1.24.1)
- Build matrix: ✅ All platforms
- TODO audit: ✅ 13 TODOs (59% reduction from v0.8.0)

**Performance:**

- Batch operations: 200+ documents/minute
- Re-embedding: 20x faster than re-ingestion
- Parallel processing: Production tested
- Average operation time: <500ms per document

**Features:**

- ✅ Batch re-embedding (v0.9.17)
- ✅ Auto-dimension detection (v0.9.16)
- ✅ Ollama auto-discovery (v0.9.18)
- ✅ Model comparison reports (v0.9.18)
- 🔧 OSS embedding providers (v0.9.19 in progress)

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
**Repository**: <https://github.com/maximilien/weave-cli>
