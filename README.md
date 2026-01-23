# Weave CLI

A fast, AI-powered command-line (CLI) tool for managing your vector database (VDBs).

Built in Go for performance and ease of use (single binary).

## Quick Start

### Installation

```bash
git clone https://github.com/maximilien/weave-cli.git
cd weave-cli
./build.sh
# Binary available at bin/weave
```

### Choose Your Vector Database

Weave CLI supports **10 vector databases**. Choose the one that best fits your needs:

| VDB | Status | Local | Cloud | Best For |
|-----|--------|-------|-------|----------|
| **[Weaviate](docs/weaviate/SETUP.md)** | ✅ Stable | ✅ | ✅ | Production, all features, easiest setup |
| **[Qdrant](docs/qdrant/SETUP.md)** | ✅ Stable | ✅ | ✅ | Rust performance, HNSW index, filtering |
| **[Milvus](docs/milvus/SETUP.md)** | ✅ Stable | ✅ | ✅ | High performance, horizontal scaling |
| **[Chroma](docs/chroma/SETUP.md)** | ✅ Stable | ✅ | ✅ | macOS only, simple setup, embeddings |
| **[Supabase](docs/supabase/SETUP.md)** | ✅ Stable | ✅ | ✅ | PostgreSQL + pgvector, cost-effective |
| **[Neo4j](docs/neo4j/README.md)** | ✅ Stable | ✅ | ⚠️ Untested | Graph + vector search, Cypher queries |
| **[MongoDB](docs/mongodb/SETUP.md)** | ✅ Stable | ❌ | ✅ | Atlas Vector Search, existing MongoDB users |
| **[Pinecone](docs/pinecone/SETUP.md)** | 🟢 Beta | ❌ | ✅ | Serverless, auto-scaling, managed only |
| **[OpenSearch](docs/opensearch/README.md)** | ✅ Stable | ✅ | ✅ | AWS OpenSearch, k-NN + BM25 hybrid |
| **[Elasticsearch](docs/elasticsearch/)** | 🟢 Beta | ✅ | ✅ | Elastic Cloud, HNSW vector + BM25 hybrid |

📖 **See [Vector Database Support Matrix](docs/VDB_SUPPORT_MATRIX.md)
for detailed feature comparison**

### Quick Setup (Weaviate - Recommended)

```bash
# Interactive configuration - fastest way to get started
weave config create --env

# Follow prompts to enter:
# - WEAVIATE_URL
# - WEAVIATE_API_KEY
# - OPENAI_API_KEY

# Verify setup
weave health check
```

For other databases, see their setup guides linked in the table above.

### Basic Usage

```bash
# List collections (all configured VDBs)
weave cols ls

# List collections from specific database types
weave cols ls --weaviate            # Weaviate only
weave cols ls --qdrant-local        # Qdrant local only
weave cols ls --qdrant-cloud        # Qdrant cloud only
weave cols ls --milvus-local        # Milvus local only
weave cols ls --milvus-cloud        # Milvus cloud (Zilliz) only
weave cols ls --chroma-local        # Chroma local only
weave cols ls --chroma-cloud        # Chroma cloud only
weave cols ls --supabase            # Supabase only
weave cols ls --neo4j-local         # Neo4j local only
weave cols ls --neo4j-cloud         # Neo4j cloud (Aura) only
weave cols ls --mongodb             # MongoDB Atlas only
weave cols ls --pinecone            # Pinecone only
weave cols ls --opensearch-local    # OpenSearch local only
weave cols ls --opensearch-cloud    # OpenSearch cloud (AWS) only
weave cols ls --elasticsearch-local # Elasticsearch local only
weave cols ls --elasticsearch-cloud # Elasticsearch cloud (Elastic) only
weave cols ls --mock                # Mock database only
weave cols ls --all                 # All configured databases

# Create a collection
weave cols create MyCollection --text

# Add documents
weave docs create MyCollection document.txt
weave docs create MyCollection document.pdf

# Search with natural language
weave cols q MyCollection "search query"

# AI-powered Read, Evaluate, Print, Loop (REPL) mode or agent mode
weave
> show me all my collections
> create TestDocs collection
> add README.md to TestDocs

# Or doing one query at a time
weave query "show me all my collections"

# List available embeddings
weave embeddings list
weave emb ls --verbose

# Create collection with specific embedding (used as default for all documents)
weave cols create MyCollection --embedding text-embedding-3-small
weave cols create MyCollection -e text-embedding-ada-002

# Get AI-powered schema suggestions for your documents
weave schema suggest ./docs --collection MyDocs --output schema.yaml

# Get AI-powered chunking recommendations
weave chunking suggest ./docs --collection MyDocs --output chunking.yaml
```

## Key Features

- 🤖 **AI-Powered** - AI Agent mode, natural language interface with GPT-4o
  multi-agent system, schema suggestions, and chunking recommendations
- ⚡ **Fast & Easy** - Written in Go with simple CLI and interactive REPL
  (AI Agent mode) with real-time progress feedback
- 🌐 **Flexible** - Weaviate Cloud, local instances, or built-in mock database
- 🔌 **Extensible** - Vector database abstraction layer supporting multiple
  backends (Weaviate, Milvus, Supabase PGVector, MongoDB Atlas, Chroma, Qdrant,
  Neo4j, OpenSearch)
- 📦 **Batch Processing** - Parallel processing of entire directories
- 📄 **PDF Support** - Intelligent text extraction and image processing
- 🔍 **Semantic Search** - Vector-based similarity search with natural
  language, including multi-collection queries
- 🧠 **AI Schema & Chunking** - Analyze documents and get AI-powered schema and
  optimal chunk size recommendations
- 📊 **Embeddings** - List and explore available embedding models
- ⏱️ **Configurable Timeouts** - Default 10s timeout, adjustable per
  command

## Documentation

### Core Documentation

- **[📖 User Guide](docs/USER_GUIDE.md)** - Complete feature documentation
- **[📋 Changelog](docs/CHANGELOG.md)** - Version history and updates
- **[🗂️ VDB Support Matrix](docs/VDB_SUPPORT.md)** - Database feature
  comparison

### Guides

- **[🤖 AI Agents](docs/guides/WEAVE_CLI_AI.md)** - REPL mode with natural
  language query system
- **[⚙️ Agent Management](docs/AGENT_MANAGEMENT.md)** - Create, customize, and
  manage RAG agents
- **[🔌 MCP AI Tools API](docs/mcp/MCP_AI_TOOLS.md)** - Using AI tools via MCP server
- **[📦 Batch Processing](docs/guides/BATCH_DOCS_CREATION.md)** - Directory
  processing guide
- **[📚 Vector DB Abstraction](docs/guides/VECTOR_DB_ABSTRACTION.md)** -
  Multi-database support architecture
- **[🎬 Demos](docs/guides/DEMO.md)** - Video demos and tutorials

### Database-Specific

- **[Chroma Documentation](docs/chroma/)** - Chroma integration guide (Stable)
- **[Milvus Documentation](docs/milvus/)** - Milvus integration guide (Beta)
- **[MongoDB Atlas Documentation](docs/mongodb/)** - MongoDB Atlas setup guide (Stable)
- **[Neo4j Documentation](docs/neo4j/)** - Neo4j integration guide (Experimental)
- **[OpenSearch Documentation](docs/opensearch/)** - OpenSearch integration
  guide (Experimental)
- **[Pinecone Documentation](docs/pinecone/)** - Pinecone integration guide (Beta)
- **[Qdrant Documentation](docs/qdrant/)** - Qdrant integration guide (Stable)
- **[Supabase Documentation](docs/supabase/)** - Supabase integration guide (Alpha)
- **[Weaviate Documentation](docs/weaviate/)** - Weaviate integration status (Stable)

## Advanced Usage

### Configuration Options

#### Auto-Configuration

Weave CLI automatically detects missing configuration:

```bash
# Try any command - you'll get prompted to configure interactively
weave cols ls

# Or install latest release of weave-mcp for REPL mode
weave config update --weave-mcp
```

**Configuration Precedence** (highest to lowest):

1. Command-line flags - `weave query --model gpt-4`
2. Environment variables - `export OPENAI_MODEL=gpt-4`
3. config.yaml (optional) - For advanced customization
4. Built-in defaults

**Configuration Location** (precedence order):

1. Local directory (`.env`, `config.yaml`) - Project-specific
   configuration
2. Global directory (`~/.weave-cli/.env`, `~/.weave-cli/config.yaml`) -
   User-wide configuration

```bash
# Create configuration in global directory
weave config create --env --global

# Sync local configuration to global directory
weave config sync

# View which configuration location is being used
weave config show
```

See the [User Guide](docs/USER_GUIDE.md#configuration) for detailed
configuration options.

### Vector Database Selection

Control which vector database(s) to operate on with these flags:

**Important**: Database selection behavior depends on your configuration:

- **Single Database**: If only one DB is configured, it's used automatically
  (no flags needed!)
- **Multiple Databases**:
  - Read operations (ls, show, count) use all databases by default
  - Write/delete operations use smart selection:
    1. **Default Database**: Uses `VECTOR_DB_TYPE` from `.env` or config
    2. **Weaviate Collection Search**: For `--weaviate`, searches all
       Weaviate databases for the collection
    3. **Manual Selection**: Use `--vector-db-type` (or `--vdb`) to
       specify explicitly

```bash
# Single database setup - no flags needed!
weave docs create MyCollection doc.txt       # Uses your only configured DB

# Multiple databases with VECTOR_DB_TYPE set
export VECTOR_DB_TYPE=weaviate-cloud
weave docs create MyCollection doc.txt       # Uses weaviate-cloud (default)
weave docs delete MyCollection doc123        # Uses weaviate-cloud (default)

# Override default with --vdb (short) or --vector-db-type (long)
weave docs create MyCollection doc.txt --vdb weaviate-local
weave docs create MyCollection doc.txt --vector-db-type supabase

# --weaviate tries both weaviate-cloud and weaviate-local
weave docs ls MyCollection --weaviate        # Searches both for collection
weave cols delete MyCollection --weaviate    # Searches both for collection

# Read operations work with specific or all databases
weave cols ls --weaviate                     # All Weaviate databases
weave cols ls --supabase                     # Supabase only
weave cols ls --all                          # All configured databases (default)

# Query multiple databases at once
weave cols query MyCollection "search" --weaviate --supabase
```

**Database Selection Priority for Single-DB Operations**:

1. If only one database configured → use it
2. If `VECTOR_DB_TYPE` set → use as default
3. If `--weaviate` flag used → try all Weaviate databases for the collection
4. Otherwise → show error with available options

### Summary Views and Filtering (New in v0.7.2)

View database status and collections across multiple databases with summary
tables and progressive output:

```bash
# Collections summary across all databases (default for multiple VDBs)
weave cols ls                    # Shows summary table by default
weave cols ls -S                 # Explicit summary flag (shorthand)
weave cols ls --summary          # Explicit summary flag (long form)

# Health check with progressive output (new in v0.7.2)
weave health check               # Shows summary table, results appear
weave health check -S            # Same as above (shorthand)

# Filter databases by deployment type (new in v0.7.2)
weave config list --cloud        # Show only cloud databases
weave config list --local        # Show only local databases

weave health check --cloud       # Check only cloud databases
weave health check --local       # Check only local databases
weave health check --local -S    # Local databases summary

# Force detailed view for single database
weave cols ls --weaviate         # Detailed list (default for single VDB)
weave health check weaviate      # Detailed health check

# Collections summary also supports filtering
weave cols ls --cloud            # Collections from cloud databases only
weave cols ls --local -S         # Local collections summary
```

**Summary Table Features**:

- **Progressive Output**: Results appear immediately as they're
  retrieved/checked (no waiting!)
- **Status Indicators**: ✓ OK (green) or ✗ FAIL (red) with color
  coding
- **Footer Statistics**: Total count, collections/healthy count,
  failures
- **Auto-Selection**: Summary for multiple VDBs, detailed for single
  VDB
- **Cloud/Local Filtering**: Filter by deployment type with `--cloud`
  or `--local` flags
- **Consistent UX**: Same behavior across `cols ls`, `health check`,
  and `config list`

### RAG Agents (All Vector Databases)

Enhance query results with AI-powered RAG (Retrieval-Augmented Generation)
agents. Now works with **all supported vector databases**, not just Weaviate!

```bash
# Use RAG agent with any vector database
weave cols query MyDocs "What is machine learning?" --agent rag-agent
weave cols query MyDocs "Summarize main topics" --agent summarize-agent --db qdrant
weave cols query MyDocs "Answer this question" --agent qa-agent --db milvus

# Combine with database selection
weave cols q MyDocs "AI overview" --agent rag-agent --qdrant-local
weave cols q MyDocs "Quick summary" --agent summarize-agent --chroma-local
weave cols q MyDocs "Detailed analysis" --agent rag-agent --mongodb

# Show progress during agent execution
weave cols q MyDocs "complex query" --agent rag-agent --progress

# JSON output with progress (JSON Lines format)
weave cols q MyDocs "query" --agent rag-agent --json --progress

# Multi-collection queries (query multiple collections at once)
weave cols query WeaveDocs WeaveImages "weave cli" --agent rag-agent --top_k 3
weave cols query AuctionsDocs AuctionsImages AuctionResults "vintage cars" --agent rag-agent
```

**Multi-Collection Queries:**

Query multiple collections simultaneously and aggregate results. The command
returns top K results from EACH collection, which are then combined and passed
to the agent for processing. Each result includes `_collection` metadata to
track its source.

```bash
# Query 2 collections (returns top 3 from EACH)
weave cols query Collection1 Collection2 "search query" --agent rag-agent --top_k 3

# Query 3+ collections (useful for multi-modal data)
weave cols query Docs Images Audio "query" --agent summarize-agent --top_k 5
```

**Cross-VDB Queries:**

Query collections from different vector databases in a single command. Use the
`Collection:vdb-key` syntax to specify which VDB each collection resides in.
Results include both collection name and VDB information in citations.

```bash
# Query collections from different VDBs
weave cols query WeaveDocs:weaviate-local WeaveImages:milvus-local "weave cli" --agent rag-agent

# AuctionsMax.ai example: Query across MongoDB, Weaviate, and Milvus
weave cols query \
  AuctionListings:mongodb-cloud \
  AuctionResults:weaviate-cloud \
  AuctionImages:milvus-cloud \
  "vintage Leica cameras" \
  --agent rag-agent --top_k 3

# Mixed: explicit VDB + default from flags
weave cols query WeaveDocs WeaveImages:milvus-local "query" --weaviate-local --agent rag-agent
```

Supported VDB keys: `weaviate-local`, `weaviate-cloud`, `milvus-local`,
`milvus-cloud`, `mongodb-cloud`, `qdrant-local`, `qdrant-cloud`,
`neo4j-local`, `neo4j-cloud`, `chroma-local`, `chroma-cloud`,
`supabase-cloud`, etc.

**Available Agents:**

- `rag-agent` - Comprehensive answers with source citations
- `summarize-agent` - Concise summaries of retrieved content
- `qa-agent` - Precise question answering

**Requirements:**

- `OPENAI_API_KEY` environment variable must be set
- Works with: Qdrant, Milvus, Chroma, MongoDB, Neo4j, Weaviate, Supabase,
  and more

See [User Guide](docs/USER_GUIDE.md#rag-agents) for custom agent configuration.

### More Examples

```bash
# Batch process documents with parallel workers
weave docs batch --directory ./docs --collection MyCollection --parallel 3

# Convert CMYK PDFs to RGB
weave docs pdf-convert document.pdf --rgb

# Text-only PDF extraction (faster, no images)
weave docs create MyCollection document.pdf --skip-all-images

# Natural language queries with AI agents
weave q "find all empty collections"
weave query "create TestDocs and add README.md" --dry-run

# Configure timeout for slow connections
weave cols ls --timeout 30s
weave health check --timeout 60s

# Create collections and documents with specific embeddings
weave cols create MyCollection --embedding text-embedding-3-small
weave docs create MyCollection document.txt --embedding text-embedding-3-small
weave docs create MyCollection report.pdf --embedding text-embedding-ada-002
```

## Database Support

Weave CLI features a **pluggable vector database abstraction layer** that
allows seamless switching between different vector database backends.

### Support Matrix

| Database | Type | Status | Maturity | Docs |
| -------- | ---- | ------ | -------- | ---- |
| **Weaviate Cloud** | `weaviate-cloud` | ✅ Production Ready | **Stable** | [Guide](docs/) |
| **Weaviate Local** | `weaviate-local` | ✅ Production Ready | **Stable** | [Guide](docs/) |
| **Milvus Local** | `milvus-local` | ✅ Functional | **Beta** - Feature complete, local testing ready | [Guide](docs/milvus/) |
| **Milvus Cloud** | `milvus-cloud` | ✅ Functional | **Beta** - Zilliz cloud integration ready | [Guide](docs/milvus/) |
| **Supabase** | `supabase` | ✅ Functional | **Alpha** - Feature complete, needs testing | [Guide](docs/supabase/) |
| **MongoDB Atlas** | `mongodb` | ✅ Functional | **Experimental** - Vector search requires index setup | [Guide](docs/mongodb/) |
| **Chroma Local** | `chroma-local` | ✅ Production Ready | **Stable** - Full CRUD, tested with v2 API ⚠️ **macOS only** | [Guide](docs/chroma/) |
| **Chroma Cloud** | `chroma-cloud` | ✅ Functional | **Beta** - Cloud integration ⚠️ **macOS only** | [Guide](docs/chroma/) |
| **Qdrant Local** | `qdrant-local` | ✅ Production Ready | **Stable** - HNSW vector search, full CRUD | [Guide](docs/qdrant/) |
| **Neo4j Local** | `neo4j-local` | ✅ Functional | **Experimental** - Graph + vector search | [Guide](docs/neo4j/) |
| **OpenSearch Local** | `opensearch-local` | ✅ Functional | **Experimental** - k-NN with HNSW algorithm | [Guide](docs/opensearch/) |
| **OpenSearch Cloud** | `opensearch-cloud` | ✅ Functional | **Experimental** - AWS OpenSearch Service | [Guide](docs/opensearch/) |
| **Mock** | `mock` | ✅ Testing Only | **Stable** | - |

### Maturity Levels

- **Stable**: Production-ready, well-tested, recommended for all use cases
- **Beta**: Feature complete, functional, ready for testing and feedback
- **Alpha**: Feature complete, functional, recommended for
  development/testing
- **Experimental**: Basic functionality working, may require manual setup,
  use with caution

### Platform Compatibility

⚠️ **Important Platform Limitations:**

- **Chroma**: macOS only (CGO dependency on `libtokenizers`)
  - Linux/Windows users: Use Weaviate, Milvus, Qdrant, or Supabase instead
  - CI/CD on Linux: Skip Chroma tests with `--skip chroma`
- **MongoDB**: Cloud (Atlas) only - no local/self-hosted support
  - Offline deployments: Use Weaviate, Milvus, Qdrant, or Neo4j instead
- **All other VDBs**: Cross-platform (Linux, macOS, Windows)

### Additional Resources

- **[Vector DB Integrations Planning](docs/planning/VECTOR_DB_INTEGRATIONS.md)**
  \- Roadmap for upcoming database support
- **[Vector DB Abstraction Guide](docs/VECTOR_DB_ABSTRACTION.md)** -
  Architecture details and how to add new databases

### Abstraction Benefits

- **Unified Interface** - Same commands work across all database types
- **Easy Migration** - Switch databases without changing workflows
- **Extensible** - Add new vector databases with minimal code changes
- **Type Safety** - Compile-time validation of database operations
- **Error Handling** - Structured error types with context and recovery

See **[📚 Vector DB Abstraction Guide](docs/VECTOR_DB_ABSTRACTION.md)** for
implementation details and adding new database support.

## Development

```bash
# Setup development environment (installs linters, PDF tools, etc.)
./setup.sh

# Build, test, and lint
./build.sh
./test.sh
./lint.sh
```

See [User Guide](docs/USER_GUIDE.md) for detailed development instructions.

## Testing & Quality

### Test Coverage

Comprehensive integration tests ensure reliability across all supported vector databases:

| VectorDB | Coverage | Integration Tests | Status |
|----------|----------|-------------------|--------|
| **Milvus** | 51.5% | 15/15 ✅ | Production Ready |
| **Qdrant** | 45.8% | 14/14 ✅ | Production Ready |
| **MongoDB** | 59.1% | 13/13 ✅ | Production Ready |
| **Weaviate** | 23.6% | 13/13 ✅ | Production Ready |
| **Neo4j** | 37.9% | 13/13 ✅ | Production Ready |
| **Chroma** | 49.1% | 12/12 ✅ | Production Ready |
| **Supabase** | 30%+ | 11/11 ✅ | Production Ready |

**Total**: 91 integration tests covering CRUD operations, collections,
search functionality (semantic, BM25, hybrid, metadata), and end-to-end
workflows.

### Running Tests

```bash
# Unit tests only (fast)
go test ./...

# Integration tests (requires Docker)
go test -tags=integration ./src/pkg/vectordb/...

# Full test suite with coverage
./test.sh integration

# Specific VDB tests
go test -tags=integration -run TestIntegration_Milvus ./src/pkg/vectordb/milvus/
```

See [Integration Test Plan](docs/planning/INTEGRATION_TEST_PLAN.md) for
detailed testing strategy.

## Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes with tests
4. Run `./test.sh` and `./lint.sh`
5. Submit a pull request

## Links

### Video Demos

- **[Full Demo (5 min)](https://asciinema.org/a/LrKzmThBfDbTPISZzr8biP4dt)** -
  Complete feature walkthrough
- **[Quick Demo (2 min)](https://asciinema.org/a/HiAU7h1iJvZ2QdJe70ae3Cc0b)** -
  Quick overview
- **[REPL Demo](https://asciinema.org/a/U504HN4FSeMsOA0qS0os0NWUE)** -
  AI-powered natural language interface

### Interactive Demos

Run these scripts locally for hands-on demonstrations:

- **[Configuration Demo](demos/config-demo.sh)** - Interactive setup and
  configuration management
- **[Supabase Demo](demos/supabase-demo.sh)** - Supabase (PostgreSQL +
  pgvector) integration

See [demos/README.md](demos/README.md) for details.

### Resources

- **[GitHub Repository](https://github.com/maximilien/weave-cli)**
- **[Documentation](docs/)**
- **[User Guide](docs/USER_GUIDE.md)**

## Presentations

Recent presentations about Weave CLI:

- **[AI by the Bay (Nov 2025)](presentations/vdbs-ai-by-the-bay-11-25.pptx)** -
  Vector database management with AI agents
  - Event: [AI by the Bay Conference](https://ai.bythebay.io/)

- **[SV AI Demo 2 (Nov 2025)](presentations/vdbs-svai-demo-11-25.pptx)** -
  Multi-database vector management demo
  - Event: [SV AI Demo Meetup](https://luma.com/bs4fkbjs?tk=bNbVMV)

## License

MIT License - see [LICENSE](LICENSE) file for details.
