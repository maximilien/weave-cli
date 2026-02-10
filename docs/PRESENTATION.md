---
marp: true
theme: default
class: lead
paginate: true
backgroundColor: #fff
backgroundImage: url('https://marp.app/assets/hero-background.svg')
---

# Weave CLI
## Universal Vector Database Management Made Simple

**A powerful command-line tool for managing 9 vector databases with a unified interface**

**Maximilien.ai** • **v0.7.x+** 

---

# What is Weave CLI?

- 🤖 **AI-Powered** - Natural language queries with GPT-4o multi-agent system
- 🔄 **Interactive REPL** - Beautiful interactive mode with command history
- 🌐 **Fast & Lightweight** - Single binary deployment
- 🎨 **Beautiful CLI** - Colored output with emojis and clear formatting
- 🔧 **Easy Configuration** - Interactive setup with sensible defaults
- 📊 **Collection Management** - Full CRUD operations on collections
- 📄 **Document Management** - Create, list, show, and delete documents
- 📦 **Batch Processing** - Parallel document processing with auto-retry
- 📄 **PDF Support** - Extract text and images with intelligent chunking
- 🎭 **Mock Database** - Built-in testing and development support

---

# Key Features

## AI Agents (NEW in v0.3.0)
- **Natural Language Queries** - `weave query "show me all collections"`
- **Multi-Agent Architecture** - 7 specialized agents working together
- **Smart Execution** - Automatic planning and execution of complex tasks
- **Cost Tracking** - OpenTelemetry integration with Opik observability
- **Interactive REPL** - Type `weave` for AI-powered interactive mode

## Database Support (9 Vector Databases!)
- **Weaviate** - Cloud & Local with full feature support
- **Pinecone** - Serverless cloud vector database
- **Qdrant** - Cloud & Local high-performance VDB
- **Chroma** - Cloud & Local embedding database
- **Milvus** - Cloud & Local open-source VDB
- **Neo4j** - Graph database with vector search
- **Supabase** - Postgres-based vector storage
- **MongoDB** - Atlas vector search support
- **Mock** - Built-in testing database

## Document Processing
- **Batch Operations** - Parallel processing with configurable workers
- **PDF Processing** - Extract text and images with intelligent chunking
- **PDF Conversion** - CMYK to RGB conversion support
- **Smart Retry** - Automatic retry with progress tracking

## Smart Document Views
- **Regular View** - Individual document listing
- **Virtual View** - Aggregate chunked documents by original file
- **Cross-collection** - Automatically includes images from PDFs

---

# Installation & Setup

```bash
# Clone and build
git clone https://github.com/maximilien/weave-cli.git
cd weave-cli
./build.sh

# Configure interactively (RECOMMENDED)
./bin/weave config create --env

# Follow the prompts to enter:
# - WEAVIATE_URL
# - WEAVIATE_API_KEY
# - OPENAI_API_KEY (for AI features)

# Test your setup
./bin/weave health check

# Start interactive REPL mode
./bin/weave
```

---

# Configuration Management

## Interactive Configuration Commands

```bash
# Create new configuration files
weave config create --env              # Create .env file
weave config create --config-yaml      # Create config.yaml file

# Update existing configuration files
weave config update --env              # Update .env file
weave config update --config-yaml      # Update config.yaml file

# View configuration
weave config show                      # Show current configuration
weave config list                      # List all configured databases
```

## Configuration Priority (highest to lowest)

1. **Command-line flags** - `weave query --model gpt-4`
2. **Environment variables** - `export OPENAI_MODEL=gpt-4`
3. **config.yaml** (optional) - Advanced customization
4. **Built-in defaults** - Sensible defaults work out of box

---

# AI-Powered Natural Language Queries

```bash
# Start interactive REPL mode
weave

# Or use direct queries
weave query "show me all my collections"
weave q "check health"
weave query "create TestDocs collection"
weave q "add README.md to TestDocs"
weave query "find all empty collections"
weave q "convert CMYK PDFs in ./pdfs directory"

# With additional options
weave query "delete old collections" --dry-run
weave q "batch process docs" --no-confirm --verbose
```

---

# Basic Commands

```bash
# Health check
weave health check

# Collection management
weave cols ls                          # List collections
weave cols create MyCollection --text  # Create collection
weave cols show MyCollection           # Show details
weave cols count                       # Count collections
weave cols del MyCollection            # Delete documents
weave cols q MyCollection "query"      # Search collection

# Document management
weave docs ls MyCollection             # List documents
weave docs show MyCollection ID        # Show document
weave docs create MyCollection file.txt # Create document
weave docs count MyCollection          # Count documents
weave docs del MyCollection ID         # Delete document
weave docs da MyCollection             # Delete all documents

# Batch processing
weave docs batch --directory ./docs --collection MyDocs --parallel 3

# PDF processing
weave docs create MyDocs document.pdf
weave docs create MyDocs doc.pdf --skip-all-images --no-tips
weave docs pdf-convert document.pdf --rgb
```

---

# Command Structure

Weave follows a consistent pattern:
**`weave noun verb [arguments] [flags]`**

## Available Commands
- **query** (q) - AI-powered natural language queries
- **config** - Configuration management (create, update, show, list)
- **health** - Health and connectivity checks
- **collection** (cols) - Collection management (ls, create, show, del, q)
- **document** (docs) - Document management (ls, show, create, del, batch)

## Global Flags
- `--no-color` - Disable colored output
- `--no-truncate` - Show all data
- `--no-tips` - Suppress helpful tips
- `--no-confirm` - Skip confirmations (for automation)
- `--verbose` / `-v` - Detailed output
- `--quiet` / `-q` - Minimal output
- `--json` - JSON output format

---

# Virtual Document View

The `--virtual` flag provides intelligent aggregation:

```bash
$ weave document list MyCollection --virtual

✅ Found 3 virtual documents (aggregated from 15 total):

1. 📄 Document: research_paper.pdf
   📝 Chunks: 3/3
   🖼️ Images: 2
   📋 Metadata: original_filename: research_paper.pdf
   📝 Chunk Details: 
     1. ID: chunk-1 - Introduction to ML...
     2. ID: chunk-2 - Deep learning...
     3. ID: chunk-3 - Conclusion...
   🗂️ Stack Details: 
     1. ID: image-1 (from page 2)
     2. ID: image-2 (from page 5)
```

---

# Cross-Collection Features

## Smart Image Aggregation
- **Automatic mapping** - Maps document collections to image collections
- **PDF extraction** - Includes images extracted from PDFs
- **Performance optimized** - Excludes large base64 data for fast queries
- **Complete view** - Shows both text chunks and images in one view

## Collection Mapping
- `MyDocs` → `MyImages`
- `Documents` → `DocumentImages`
- Automatic detection based on naming patterns

---

# Configuration in Detail

## Interactive Setup (Recommended)
```bash
# Create new configuration
weave config create --env              # Interactive .env creation
weave config create --config-yaml      # Interactive config.yaml

# Update existing configuration
weave config update --env              # Update .env values
weave config update --config-yaml      # Update config.yaml

# View configuration
weave config show                      # Current config
weave config list                      # All databases
```

## Environment Variables (.env)
```bash
# Required - Only 3 credentials needed!
WEAVIATE_URL="https://your-cluster.weaviate.cloud"
WEAVIATE_API_KEY="your-api-key"
OPENAI_API_KEY="sk-proj-your-openai-key"

# Optional - Auto-detected from URL
VECTOR_DB_TYPE="weaviate-cloud"

# Optional - Sensible defaults provided
WEAVIATE_COLLECTION="MyCollection"
WEAVIATE_COLLECTION_IMAGES="MyImages"
```

## Config File (config.yaml) - Optional!
Most users only need `.env`! Use config.yaml for:
- Custom collection schemas
- Batch processing settings
- PDF processing options
- AI agent configuration

## Database Types
Supported via `--type` flag or `VECTOR_DB_TYPE` env var:
- **weaviate** - Weaviate Cloud/Local
- **pinecone** - Pinecone serverless
- **qdrant** - Qdrant Cloud/Local
- **chroma** - Chroma Cloud/Local
- **milvus** - Milvus Cloud/Local
- **neo4j** - Neo4j with vector index
- **supabase** - Supabase Weaviate
- **mongodb** - MongoDB Atlas
- **mock** - Built-in testing (no setup needed!)

---

# Development

## Project Structure
```
weave-cli/
├── src/                    # Source code
│   ├── cmd/               # CLI commands
│   ├── pkg/               # Public packages
│   │   ├── config/       # Configuration
│   │   ├── weaviate/     # Weaviate client
│   │   └── mock/         # Mock database
│   └── main.go           # Entry point
├── docs/                   # Documentation
├── tests/                 # Test files
└── bin/                   # Built binaries
```

---

# Development Workflow

```bash
# Build everything
./build.sh

# Run tests
./test.sh

# Run linter
./lint.sh

# Run security checks
govulncheck ./src/...
```

## Quality Assurance
- ✅ **Go linting** - golangci-lint
- ✅ **Security scanning** - govulncheck, gosec
- ✅ **YAML validation** - yamllint
- ✅ **Markdown linting** - markdownlint
- ✅ **Shell checking** - shellcheck

---

# Testing

## Test Coverage
- **Unit tests** - All components tested
- **Integration tests** - Weaviate connectivity
- **Mock database** - No external dependencies
- **Edge cases** - Error handling and validation

## Test Commands
```bash
# Run all tests
./test.sh

# Run specific test types
./test.sh unit
./test.sh integration
./test.sh coverage
```

---

# CI/CD Pipeline

## GitHub Actions
- **Multi-platform builds** - Linux, macOS, Windows
- **Automated testing** - Multiple Go versions
- **Security scanning** - Vulnerability checks
- **Automated releases** - Binary distribution

## Quality Gates
- All tests must pass
- All linting checks must pass
- Security scans must be clean
- Code coverage maintained

---

# Recent Additions (v0.7.x)

## ✅ Completed Features
- 🗄️ **9 Vector Databases** - Universal VDB support (v0.5.0+)
- 🎯 **Pinecone** - Serverless cloud vector database (v0.7.6)
- 🔍 **Qdrant** - High-performance similarity search (v0.6.0)
- 📊 **Chroma** - Embedding database support (v0.5.5)
- 🚀 **Milvus** - Cloud & Local open-source VDB (v0.6.5)
- 🔗 **Neo4j** - Graph database with vector search (v0.7.0)
- 💾 **Supabase** - Postgres-based vector storage (v0.6.8)
- 🍃 **MongoDB** - Atlas vector search (v0.7.2)
- 🤖 **AI Agents** - Natural language queries with GPT-4o (v0.3.0)
- 🔄 **Interactive REPL** - AI-powered interactive mode (v0.3.0)

---

# Future Roadmap

## High Priority
- 🧪 **Enhanced Testing** - Expand test coverage for all VDBs
- 🤖 **More LLM Models** - Claude, Gemini, Llama support
- 📊 **Advanced Schema** - Cross-VDB schema validation
- 🔍 **Unified Search** - Advanced filters across all databases

## Medium Priority
- 💾 **Backup/Restore** - Cross-VDB migration tools
- 📈 **Monitoring** - Real-time database metrics dashboard
- 🎯 **Query Templates** - Reusable query patterns
- 🔄 **Sync Operations** - Multi-VDB synchronization

---

# OSS Embedding Providers 🌟
## 100% Free, Local Embeddings (NEW in v0.9.19)

**3 Embedding Providers Available:**

| Provider | Type | Cost | Dimensions | Performance | Best For |
|----------|------|------|------------|-------------|----------|
| **OpenAI** | Cloud API | $0.02/1M | 1536, 3072 | Baseline (100%) | Production quality |
| **sentence-transformers** | Local Python | **FREE** | 384, 768 | 90-95% | Cost savings, privacy |
| **Ollama** | Local HTTP | **FREE** | 768, 1024 | 90-95% | Local LLMs, offline |

---

# Why OSS Embeddings?

## Benefits
- 💰 **100% Cost Savings** - No API fees, completely free
- 🔒 **Privacy** - All processing local, no data leaves your machine
- ⚡ **Performance** - Often faster due to no network latency
- 📴 **Offline** - Works without internet connection
- 🎯 **Quality** - 90%+ retention vs OpenAI in testing

## Use Cases
- **Development** - Free testing and prototyping
- **Production** - Cost-sensitive deployments
- **Privacy** - Sensitive data that can't use cloud APIs
- **Offline** - Air-gapped environments
- **Hybrid** - Mix OpenAI (critical) + OSS (bulk)

---

# Quick Start - OSS Embeddings

## sentence-transformers (Recommended)

```bash
# 1. Install (one time)
pip install sentence-transformers

# 2. Re-embed existing collection (20x faster than re-ingestion!)
weave collection reembed MyCollection \
  --new-embedding sentence-transformers/all-mpnet-base-v2 \
  --output MyCollection_OSS

# 3. Compare quality vs OpenAI
weave collection compare MyCollection MyCollection_OSS \
  --query "test query" \
  --report comparison.md

# 4. Review results and decide
cat comparison.md
```

**Time:** 5 minutes setup + ~1 minute per 1000 documents

---

# OSS Workflow Example

## 3-Way Comparison

```bash
# Re-embed with all 3 providers
weave collection reembed AuctionListings \
  --new-embedding text-embedding-3-small \
  --output AuctionListings_OpenAI

weave collection reembed AuctionListings \
  --new-embedding sentence-transformers/all-mpnet-base-v2 \
  --output AuctionListings_OSS

ollama pull nomic-embed-text
weave collection reembed AuctionListings \
  --new-embedding nomic-embed-text \
  --output AuctionListings_Ollama

# Compare all 3
weave collection compare \
  AuctionListings_OpenAI \
  AuctionListings_OSS \
  AuctionListings_Ollama \
  --query "vintage cameras" \
  --report comparison.md
```

---

# Performance Results

## Real-World Benchmarks (3,518 documents)

### Speed
- **OpenAI**: ~45 seconds (API calls)
- **sentence-transformers**: ~30 seconds (local)
- **Ollama**: ~35 seconds (local)

### Quality (Relevance Scores)
- **OpenAI**: 0.82 avg score (baseline)
- **sentence-transformers**: 0.78 avg score (95% retention) ✅
- **Ollama**: 0.77 avg score (94% retention) ✅

### Cost (Per Re-Embedding)
- **OpenAI**: $0.10
- **sentence-transformers**: $0.00 💰
- **Ollama**: $0.00 💰

**Annual Savings** (4 re-embeddings/year × 3 collections): **$1.20/year** per use case

---

# Supported Models

## sentence-transformers
- `all-mpnet-base-v2` (768d) - **Recommended** for quality
- `all-MiniLM-L6-v2` (384d) - Fastest, lightweight
- `all-MiniLM-L12-v2` (384d) - Balance of speed/quality
- `paraphrase-MiniLM-L6-v2` (384d) - Paraphrase detection

## Ollama
- `nomic-embed-text` (768d) - **Recommended** for embeddings
- `mxbai-embed-large` (1024d) - Higher dimensions
- `snowflake-arctic-embed` (1024d) - Alternative option

## OpenAI (Baseline)
- `text-embedding-3-small` (1536d) - Standard
- `text-embedding-3-large` (3072d) - High quality
- `text-embedding-ada-002` (1536d) - Legacy

---

# Architecture

## Provider Pattern

```
┌─────────────────────────────────────────┐
│         Embedding Pipeline              │
└────────────────┬────────────────────────┘
                 │
         ┌───────▼────────┐
         │     Factory     │
         │  (Auto-detect)  │
         └───────┬────────┘
                 │
     ┌───────────┼──────────┐
     │           │          │
┌────▼────┐ ┌───▼───┐ ┌───▼────┐
│ OpenAI  │ │ s-t   │ │ Ollama │
│Provider │ │Provider│ │Provider│
└────┬────┘ └───┬───┘ └───┬────┘
     │          │          │
     └──────────┼──────────┘
                │
         ┌──────▼──────┐
         │  Document   │
         │ (embedding) │
         └──────┬──────┘
                │
         ┌──────▼──────┐
         │  VDB Adapter│
         │(no regen!)  │
         └─────────────┘
```

**Key:** Pre-generated embeddings, no double-generation!

---

# Client0 Validation

## 3-Week Testing Workflow

**Week 1: OpenAI Baseline** ✅
- Re-embed all collections with OpenAI
- Establish quality/performance baseline
- Measure costs

**Week 2: OSS Testing** ✅
- Re-embed with sentence-transformers
- Compare quality metrics
- Validate performance

**Week 3: Ollama Testing** ✅
- Re-embed with Ollama
- Final 3-way comparison
- Make data-driven decision

**Result:** Successfully validated OSS embeddings with 90%+ quality retention!

---

# Getting Started

## Resources
- 📖 **[OSS Embedding Testing Guide](guides/OSS_EMBEDDING_TESTING_TIPS.md)** - Complete setup and troubleshooting
- 📊 **[Comparison Reports](https://github.com/maximilien/weave-cli/blob/main/CHANGELOG.md)** - Example results
- 🎬 **Demo Scripts** - `demos/oss-embeddings-demo.sh`

## Support
- Issues resolved: 4 of 5 Client0 issues ✅
- Critical Gap #1: Ollama reembed support ✅
- Production ready: All providers tested ✅

---

# Contributing

## Getting Started
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run test suite: `./test.sh`
6. Run linter: `./lint.sh`
7. Submit a pull request

## Code Quality
- Follow Go best practices
- Add comprehensive tests
- Update documentation
- Ensure all checks pass

---

# License & Acknowledgments

## License
This project is licensed under the **MIT License**

## Built With
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Weaviate Go Client](https://github.com/weaviate/weaviate-go-client) - Database operations
- Inspired by RAGme.io's tools/vdb.sh script

## Links
- 📖 **[Complete User Guide](USER_GUIDE.md)**
- 🤖 **[AI Agents Guide](WEAVE_CLI_AI.md)**
- 📦 **[Batch Processing Guide](BATCH_DOCS_CREATION.md)**
- 🎬 **[Demo Guide](DEMO.md)**
- 🐙 **[GitHub Repository](https://github.com/maximilien/weave-cli)**
- 📋 **[Issues & Discussions](https://github.com/maximilien/weave-cli/issues)**

---

# Questions?

## Get Help
- 📖 Read the [User Guide](USER_GUIDE.md)
- 🐛 Report issues on [GitHub](https://github.com/maximilien/weave-cli/issues)
- 💬 Join discussions in [GitHub Discussions](https://github.com/maximilien/weave-cli/discussions)

## Thank You!
**Weave CLI** - Making vector database management simple and powerful! 🚀