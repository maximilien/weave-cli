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

# List configured databases
weave vdb list

# Show detailed database info
weave vdb info weaviate-cloud
```

For other databases, see their setup guides linked in the table above.

### Choose Your Embedding Provider

Weave CLI supports **3 embedding providers** - including 100% free, local options:

| Provider | Type | Cost | Dimensions | Setup Time | Best For |
|----------|------|------|------------|------------|----------|
| **OpenAI** | Cloud API | $0.02/1M tokens | 1536, 3072 | 2 min | Production, quality |
| **sentence-transformers** | Local Python | FREE | 384, 768 | 5 min | Cost savings, privacy |
| **Ollama** | Local HTTP | FREE | 768, 1024 | 10 min | Local LLMs, offline |

**Quick Start - OSS Embeddings:**

```bash
# Install sentence-transformers (one time)
pip install sentence-transformers

# Create collection with OSS embedding model
weave docs create MyDocs_OSS data/documents.pdf \
  --embedding sentence-transformers/all-mpnet-base-v2

# Re-embed existing collection (20x faster than re-ingestion!)
weave collection reembed MyCollection \
  --new-embedding sentence-transformers/all-mpnet-base-v2 \
  --output MyCollection_OSS

# Query automatically uses collection's embedding model
weave cols query MyDocs_OSS "search query" --top-k 5

# Compare OpenAI vs OSS embeddings
weave collection compare MyCollection MyCollection_OSS \
  --query "search query" \
  --report comparison.md
```

**Available Models:**

```bash
# List all available embedding models
weave embeddings list

# Output shows providers with dimensions and support status:
# OpenAI Models:
#   text-embedding-3-small     (1536 dims) ✅
#   text-embedding-3-large     (3072 dims) ✅
#   text-embedding-ada-002     (1536 dims) ✅
#
# sentence-transformers (OSS):
#   all-mpnet-base-v2          (768 dims)  ✅ Recommended
#   all-MiniLM-L6-v2           (384 dims)  ✅ Fast
#
# Ollama (Local):
#   nomic-embed-text           (768 dims)  ✅
#   mxbai-embed-large          (1024 dims) ✅
```

**Cost Savings:**

- **10M tokens/month**: Save **$240/year** using OSS vs OpenAI
- **Performance**: OSS models achieve 90%+ quality retention vs OpenAI
- **Privacy**: All processing happens locally - no data leaves your machine

**Works with ALL Vector Databases:**

OSS embeddings use pre-generated embeddings (via `doc.Embedding` field), making them compatible with all 10 supported vector databases - no VDB-specific configuration required!

📖 **See [OSS Embedding Testing Guide](docs/guides/OSS_EMBEDDING_TESTING_TIPS.md)
for setup, troubleshooting, and benchmarks**

### External Storage for Large Images (v0.10.0+)

Some vector databases have size limits for storing images directly. For example, **Milvus has a 65KB VARCHAR limit**, which can only safely store ~47KB images after base64 encoding.

Weave CLI automatically handles large images using **external storage** (S3, MinIO, or local filesystem):

| Storage | Type | Cost | Use Case |
|---------|------|------|----------|
| **MinIO** | Self-hosted S3 | FREE | Local development, testing |
| **AWS S3** | Cloud object storage | Pay-as-you-go | Production, CDN integration |
| **Local** | Filesystem | FREE | Testing, single-machine |

**How It Works:**

1. Images **≤47KB**: Stored directly in vector database (fast)
2. Images **>47KB**: Thumbnail (<47KB) stored in VDB, full image in external storage
3. Automatic: No code changes required - just add flags!

**Quick Start with MinIO (Local):**

```bash
# Start MinIO (one time)
./scripts/start-minio.sh

# Ingest images with automatic external storage
weave docs create AuctionImages data/images/ \
  --image-storage minio \
  --minio-bucket weave-images \
  --milvus-local

# Images >47KB automatically uploaded to MinIO
# Thumbnails stored in Milvus for fast preview
# Full-resolution URLs: http://localhost:9000/weave-images/...
```

**Production with AWS S3:**

```bash
# Set AWS credentials (or use --s3-access-key/--s3-secret-key)
export AWS_ACCESS_KEY_ID=your-key
export AWS_SECRET_ACCESS_KEY=your-secret

# Ingest with S3 storage
weave docs create ProductCatalog images/ \
  --image-storage s3 \
  --s3-bucket my-product-images \
  --s3-region us-west-2 \
  --milvus-cloud
```

**Supported Flags:**

```bash
--image-storage s3|minio|local  # Storage backend
--s3-bucket <name>              # S3 bucket name
--s3-region <region>            # AWS region (default: us-east-1)
--minio-bucket <name>           # MinIO bucket name
--minio-endpoint <host:port>    # MinIO endpoint (default: localhost:9000)
--local-storage-path <path>     # Local storage directory (default: ./storage)
--store-pdf                     # Also store PDFs in external storage
```

**What Gets Stored Where:**

| Field | Small Images (≤47KB) | Large Images (>47KB) |
|-------|---------------------|---------------------|
| `image_data` | Full base64 image | Empty (cleared) |
| `image_thumbnail` | Empty | Base64 thumbnail (<47KB) |
| `image_url` | Empty | S3/MinIO URL |
| `image_metadata` | Empty | Size, format, storage type |

**Benefits:**

- ✅ No VDB size limits - store unlimited image sizes
- ✅ Cost optimization - cheaper object storage for large files
- ✅ CDN integration - S3 URLs work with CloudFront
- ✅ Fast previews - thumbnails in VDB for instant display
- ✅ Full resolution - access original via URL when needed

📖 **See [MinIO Setup Guide](docker-compose.minio.yml) for detailed configuration**

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
- 📊 **Embeddings** - Multiple providers (OpenAI, sentence-transformers, Ollama)
  with 100% free local options. Batch re-embedding 20x faster than re-ingestion
- ⏱️ **Configurable Timeouts** - Smart operation-specific defaults (Health: 10s,
  Bulk: 120s), adjustable per command (`--timeout 30s`) or in config.yaml. See
  [Timeout Configuration Guide](docs/TIMEOUT_CONFIGURATION.md)
- 🛡️ **Production Hardening** - Agent input validation with typo suggestions,
  structured logging with file output (`--log-level`, `--log-file`), and rich
  error context for debugging (v0.9.12+)
- 📊 **Observability** - Production-ready monitoring with Prometheus metrics,
  health endpoints, and structured JSON logging for Kubernetes deployments
  (v0.9.15+). See [Observability Guide](docs/OBSERVABILITY.md)

## Documentation

### Core Documentation

- **[📖 User Guide](docs/USER_GUIDE.md)** - Complete feature documentation
- **[📋 Changelog](CHANGELOG.md)** - Version history and updates
- **[🗂️ VDB Support Matrix](docs/VDB_SUPPORT_MATRIX.md)** - Database feature
  comparison
- **[🏗️ Architecture](docs/ARCHITECTURE.md)** - System architecture including
  embedding provider patterns (v0.9.19+)

### Guides

- **[🤖 AI Agents](docs/guides/WEAVE_CLI_AI.md)** - REPL mode with natural
  language query system
- **[⚙️ Agent Management](docs/AGENT_MANAGEMENT.md)** - Create, customize, and
  manage RAG agents
- **[🔌 MCP AI Tools API](docs/mcp/MCP_AI_TOOLS.md)** - Using AI tools via MCP server
- **[🧬 OSS Embedding Testing](docs/guides/OSS_EMBEDDING_TESTING_TIPS.md)** -
  Setup, troubleshooting, and benchmarks for open-source embeddings (v0.9.19+)
- **[📊 Observability](docs/OBSERVABILITY.md)** - Production monitoring with
  Prometheus metrics, health endpoints, and structured logging (v0.9.15+)
- **[⏱️ Timeout Configuration](docs/TIMEOUT_CONFIGURATION.md)** - Configure and
  troubleshoot operation timeouts
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

### Vector Database Management

Manage your vector database configurations with dedicated commands:

```bash
# List all configured databases
weave vdb list                    # Show all databases
weave vdb ls                      # Alias for list
weave vdb list --cloud            # Show only cloud databases
weave vdb list --local            # Show only local databases

# Show detailed database information
weave vdb info weaviate-cloud     # Connection details, auth, collections
weave vdb info milvus-local       # Vector settings, endpoints

# Check database health
weave vdb health                  # Directs to 'weave health check'
weave vdb health weaviate-cloud   # Check specific database

# Aliases for discoverability
weave db list                     # Same as 'weave vdb list'
weave database info my-db         # Same as 'weave vdb info my-db'
```

**Features:**

- **List view** - Table showing name, type, endpoint, and default status
- **Info view** - Detailed configuration including auth status (masked
  secrets), vector dimensions, similarity metrics, and collections
- **Filtering** - `--cloud` and `--local` flags to filter by deployment
  type
- **Better UX** - More intuitive than `weave config list` for
  database-specific operations

### Timeout Configuration

Control operation timeouts for better performance and reliability:

```bash
# Quick timeout for fast connectivity test
weave health check --timeout 5s

# Longer timeout for cloud operations
weave cols query MyDocs "search" --timeout 30s

# Very long timeout for bulk imports
weave docs batch --dir ./data --collection Docs --timeout 600s

# Configure per-database in config.yaml
timeout: 30  # seconds
```

See **[Timeout Configuration Guide](docs/TIMEOUT_CONFIGURATION.md)** for
details on operation-specific defaults and troubleshooting.

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

**Multi-Agent Orchestration (New in v0.9.11):**

Chain multiple agents in sequence for advanced query processing. Each agent
receives the output from the previous agent, enabling sophisticated workflows
like comprehensive analysis followed by summarization.

```bash
# Chain agents: RAG analysis → QA extraction
weave cols query MyDocs "machine learning algorithms" --agents rag-agent,qa-agent

# Triple chain: Search → Comprehensive RAG → Summarization
weave cols query LargeDocs "annual report" --agents search-agent,rag-agent,summarize-agent

# Multi-collection + multi-agent
weave cols query Docs Images "product features" --agents rag-agent,summarize-agent --top_k 10

# With progress tracking
weave cols query MyDocs "complex query" --agents rag-agent,qa-agent --progress
```

**Chain Behavior:**

- Sequential execution: Agent1 → Agent2 → Agent3
- Context passing: Each agent receives previous agent's output
- Error handling: Continues on failure, returns last successful output
- All query types supported: single collection, multi-collection, cross-VDB

See [Multi-Agent Examples](docs/examples/MULTI_AGENT_EXAMPLES.md) for detailed
use cases and patterns.

**Requirements:**

- `OPENAI_API_KEY` environment variable must be set
- Works with: Qdrant, Milvus, Chroma, MongoDB, Neo4j, Weaviate, Supabase,
  and more

See [User Guide](docs/USER_GUIDE.md#rag-agents) for custom agent configuration.

### Agent Evaluation (New in v0.9.10)

Systematically evaluate and improve your RAG agents using test datasets and
multiple evaluation metrics. Perfect for iterating on agent configurations,
prompts, and parameters.

```bash
# Run evaluation with baseline dataset
weave eval run --agent rag-agent --dataset baseline

# Create custom test dataset from template
weave eval datasets create my-tests --template baseline

# List all available datasets
weave eval datasets list

# Show evaluation results
weave eval show eval-20260126-143022

# Compare multiple agents on same dataset
weave eval benchmark --agents rag-agent,qa-agent --dataset baseline
```

**Evaluator Providers:**

Choose between local or Opik-based evaluation:

```bash
# Default: Local evaluators (uses OpenAI API key)
weave eval run --agent rag-agent --dataset baseline

# Opik: Cloud-based with dashboards & cost tracking
weave eval run --agent rag-agent --dataset baseline --use-opik
```

**Local Evaluators (Default):**

- ✅ Works offline
- ✅ Fast iteration during development
- ✅ Direct OpenAI/Claude API calls
- ✅ No additional setup required

**Opik Evaluators (`--use-opik`):**

- ✅ Rich dashboard visualization
- ✅ Automatic cost tracking
- ✅ Historical trend analysis
- ✅ Team collaboration features
- 🔑 Requires: `OPIK_API_KEY` environment variable

**Evaluation Metrics:**

Weave CLI evaluates agents on 5 key dimensions:

1. **Accuracy** - Semantic similarity to expected answer (LLM-as-judge)
2. **Citation** - Proper source attribution (rule-based)
3. **Hallucination** - Detects unsupported claims (LLM-as-judge)
4. **Context Relevance** - Quality of retrieved chunks (LLM-as-judge)
5. **Faithfulness** - Answer supported by context (LLM-as-judge)

**Custom Evaluators:**

Define domain-specific evaluation metrics without writing code using YAML
configuration:

```bash
# List available custom evaluators
weave eval list-evaluators

# Create a new evaluator from template
weave eval create-evaluator citation_check --type regex

# Validate evaluator definition
weave eval validate-evaluator evals/evaluators/citation_check.yaml
```

**Supported Evaluator Types:**

- **llm_judge** - LLM-based evaluation with custom prompts
- **regex** - Pattern matching for specific formats
- **exact_match** - Exact string comparison
- **contains** - Substring/keyword checking

**Example Custom Evaluator (regex):**

```yaml
name: citation_check
description: Checks if response includes proper citations
version: 1.0.0

scoring:
  type: regex
  pattern: '\[[0-9]+\]'  # Matches [1], [2], etc.
  threshold: 0.9

tags:
  - citation
  - compliance
```

**Using in Datasets:**

Reference custom evaluators in your test datasets:

```yaml
name: my-dataset
custom_evaluators:
  - citation_check
  - technical_accuracy

test_cases:
  - id: test-001
    query: "What is a vector database?"
    expected_answer: "A vector database stores data..."
```

See [evals/evaluators/README.md](evals/evaluators/README.md) for complete
documentation.

**Example Output:**

```text
=== Evaluation Summary ===
Provider:      local
Total Tests:   25
Passed:        22 (88.0%)
Failed:        3 (12.0%)

Average Scores:
  Accuracy:       0.87
  Citation:       0.85
  Hallucination:  0.92 (higher is better)
  Context Rel:    0.89
  Faithfulness:   0.91

Custom Evaluators:
  citation_check:      0.92
  technical_accuracy:  0.84

Performance:
  Total time:     45.3s
  Avg per test:   1.8s

✅ Results saved to: evals/results/eval-20260126-143022.json
View results: weave eval show eval-20260126-143022
```

**With Opik Dashboard:**

When using `--use-opik`, get additional insights:

```text
=== Evaluation Summary ===
Provider:      opik
...
Cost: $0.45 (tracked by Opik)

📊 View detailed results in Opik dashboard:
   https://www.comet.com/your-workspace/weave-cli-evals

   Dashboard includes:
   - Detailed trace of each evaluation
   - Cost breakdown
   - Historical trends
   - Export to CSV/JSON
```

**Creating Custom Datasets:**

```bash
# Interactive creation
weave eval datasets create medical-qa --interactive

# From template
weave eval datasets create my-tests --template baseline

# Copy existing dataset
weave eval datasets create my-copy --from baseline

# Minimal dataset
weave eval datasets create quick-test --minimal
```

**Dataset Format (YAML):**

```yaml
name: medical-qa
version: "1.0.0"
description: Medical question answering tests

test_cases:
  - id: test-001
    query: "What are the symptoms of diabetes?"
    expected_answer: "Common symptoms include increased thirst..."
    required_concepts: ["diabetes", "symptoms"]
    must_cite: true
    min_relevance_score: 0.8

  - id: test-002
    query: "How is hypertension treated?"
    expected_answer: "Treatment includes lifestyle changes..."
    retrieved_context:
      - "Hypertension is treated with medications..."
      - "Lifestyle modifications are important..."
```

**Requirements:**

- `OPENAI_API_KEY` for local evaluators
- `OPIK_API_KEY` + `OPIK_WORKSPACE` for Opik evaluators (optional)

See [Evaluation Guide](docs/EVALUATION.md) for advanced usage.

### OSS Embedding Workflows

Comprehensive examples using open-source embedding models:

```bash
# 1. Setup OSS Embeddings (one-time)
pip install sentence-transformers

# For Ollama (alternative OSS provider):
# brew install ollama && ollama pull nomic-embed-text

# 2. Create Collection with OSS Model
weave docs create TechDocs documentation.pdf \
  --embedding sentence-transformers/all-mpnet-base-v2 \
  --milvus-local

# 3. Re-embed Existing Collections (20x faster!)
# Re-embed from OpenAI to OSS (saves $240/year on 10M tokens)
weave collection reembed MyDocs \
  --new-embedding sentence-transformers/all-mpnet-base-v2 \
  --output MyDocs_OSS

# Re-embed to test different OSS models
weave collection reembed MyDocs_OSS \
  --new-embedding sentence-transformers/all-MiniLM-L6-v2 \
  --output MyDocs_Fast

# 4. Query (automatically uses collection's embedding model)
weave cols query MyDocs_OSS "machine learning algorithms" --top-k 5

# Query with RAG agent
weave cols query MyDocs_OSS "explain neural networks" \
  --agent rag-agent \
  --top-k 10

# 5. Compare Embedding Models
# Compare OpenAI vs OSS quality
weave collection compare MyDocs MyDocs_OSS \
  --query "test query" \
  --report openai-vs-oss.md

# Compare two OSS models
weave collection compare MyDocs_OSS MyDocs_Fast \
  --query "performance test" \
  --report mpnet-vs-minilm.md

# 6. Multi-Collection with Mixed Embeddings
# Query collections using different embedding models
weave cols query OpenAIDocs OSSDocsw "search across both" \
  --top-k 5 \
  --agent rag-agent

# 7. List Available Embeddings
weave embeddings list              # Show all providers
weave emb ls --verbose             # Detailed info with dimensions
```

**Typical OSS Workflow:**

```bash
# Phase 1: Initial ingestion with OpenAI (fast development)
weave docs create QuickTest data.pdf --embedding text-embedding-3-small

# Phase 2: Re-embed to OSS for production (cost optimization)
weave collection reembed QuickTest \
  --new-embedding sentence-transformers/all-mpnet-base-v2 \
  --output Production

# Phase 3: Test quality vs original
weave collection compare QuickTest Production \
  --query "quality test" \
  --report quality-comparison.md

# Phase 4: Deploy OSS collection
weave cols query Production "production queries" --agent rag-agent
```

**Performance Benchmarks:**

- **Re-embedding speed**: 200+ documents/minute (vs 10 docs/min for full re-ingestion)
- **Quality retention**: 90%+ vs OpenAI for most use cases
- **Cost savings**: $240/year for 10M tokens/month workload
- **Compatibility**: Works with ALL 10 vector databases

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
