<!-- markdownlint-disable MD025 MD022 MD036 MD032 MD041 MD034 MD003 MD031 MD040 MD013 MD026 -->
---
marp: true
theme: default
class: lead
paginate: true
backgroundColor: #fff
backgroundImage: url('https://marp.app/assets/hero-background.svg')
---

# Weave CLI
## Multi-VDB Management Made Simple

**A powerful command-line tool for managing vector databases**

**Maximilien.ai** • v0.4.0

---

# What is Weave CLI?

- 🤖 **AI-Powered** - Natural language queries with GPT-4o multi-agent system
- 🗄️ **Multi-Database** - Weaviate, Supabase, MongoDB Atlas support
- 🔄 **Interactive REPL** - Beautiful interactive mode with command history
- 🌐 **Fast & Lightweight** - Single binary deployment (53MB)
- 🎨 **Beautiful CLI** - Colored output with emojis and clear formatting
- 🔧 **Easy Configuration** - Interactive setup with sensible defaults
- 📊 **Full CRUD** - Complete collection and document management
- 📦 **Batch Processing** - Parallel document processing with auto-retry
- 📄 **PDF Support** - Extract text and images with intelligent chunking
- 🎭 **Mock Database** - Built-in testing and development support

---

# Key Features

## AI Agents (v0.3.0+)
- **Natural Language Queries** - `weave query "show me all collections"`
- **Multi-Agent Architecture** - 7 specialized agents working together
- **Smart Execution** - Automatic planning and execution of complex tasks
- **Cost Tracking** - OpenTelemetry integration with Opik observability
- **Interactive REPL** - Type `weave` for AI-powered interactive mode

## Multi-Database Support (v0.4.0)
- **Weaviate Cloud/Local** - Production-ready (Stable)
- **Supabase PGVector** - Feature complete (Alpha)
- **MongoDB Atlas** - Vector search with automatic embeddings (Experimental)
- **Mock Database** - Built-in mock for testing (Stable)

---

# Database Support Matrix

| Database | Status | Maturity | Features |
|----------|--------|----------|----------|
| **Weaviate Cloud/Local** | ✅ | **Stable** | GraphQL API, native vectorizers |
| **Supabase PGVector** | ✅ | **Alpha** | PostgreSQL + pgvector, auto embeddings |
| **MongoDB Atlas** | ✅ | **Experimental** | Vector search, BM25, hybrid search |
| **Mock** | ✅ | **Stable** | Testing only, no dependencies |

**Maturity Levels:**
- **Stable**: Production-ready, well-tested
- **Alpha**: Feature complete, recommended for dev/test
- **Experimental**: Working, requires manual setup

---

# MongoDB Atlas Integration (NEW in v0.4.0)

## Key Features
- ✅ **Automatic Embeddings** - Documents get embeddings on creation
- ✅ **Vector Search** - Semantic search using $vectorSearch aggregation
- ✅ **BM25 Search** - Keyword search using MongoDB text indexes
- ✅ **Hybrid Search** - Combines vector + BM25 with RRF
- ✅ **Free Tier** - Works with MongoDB Atlas M0 (free)

## Requirements
- MongoDB Atlas cluster (M0 free tier available)
- Vector search index (created via Atlas UI)
- OPENAI_API_KEY environment variable

See [MongoDB Documentation](docs/mongodb/) for setup

---

# Document Processing

## Batch Operations
- **Parallel Processing** - Configurable workers for speed
- **Auto-Retry** - Smart retry with progress tracking
- **Directory Processing** - Process entire directories at once

## PDF Processing
- **Text Extraction** - Intelligent text chunking (default 5000 chars)
- **Image Extraction** - Extract images with OCR and EXIF data
- **CMYK Conversion** - Automatic CMYK to RGB conversion
- **Embedding Generation** - Automatic for all text chunks

## Smart Document Views
- **Regular View** - Individual document listing
- **Virtual View** (`-w`) - Aggregate chunked documents by original file
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
# - WEAVIATE_URL (or MONGODB_URI, SUPABASE_DATABASE_URL)
# - WEAVIATE_API_KEY (or database credentials)
# - OPENAI_API_KEY (for AI features and embeddings)

# Test your setup
./bin/weave health check

# Start interactive REPL mode
./bin/weave
```

---

# Quick Start Examples

```bash
# List collections (all configured VDBs)
weave cols ls

# List from specific database
weave cols ls --weaviate    # Weaviate only
weave cols ls --supabase    # Supabase only
weave cols ls --mongodb     # MongoDB Atlas only
weave cols ls --all         # All databases

# Create collection and add documents
weave cols create MyCollection --text
weave docs create MyCollection README.md
weave docs create MyCollection docs/USER_GUIDE.md

# Search with natural language
weave cols q MyCollection "what is weave?"
```

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

# Database-specific queries
weave q "list collections in mongodb"
weave query "create WeaveDocs in supabase"

# With additional options
weave query "delete old collections" --dry-run
weave q "batch process docs" --no-confirm --verbose
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

# Basic Commands

```bash
# Health check
weave health check

# Collection management
weave cols ls                          # List collections
weave cols create MyCollection --text  # Create collection
weave cols show MyCollection           # Show details
weave cols count                       # Count collections
weave cols del MyCollection            # Delete collection
weave cols q MyCollection "query"      # Search collection

# Document management
weave docs ls MyCollection             # List documents
weave docs show MyCollection ID        # Show document
weave docs create MyCollection file.txt # Create document
weave docs count MyCollection          # Count documents
weave docs del MyCollection ID         # Delete document
weave docs da MyCollection             # Delete all documents
```

---

# Advanced Features

## Embeddings Management (v0.3.12+)
```bash
# List available embedding models
weave embeddings list
weave emb ls --verbose

# Create collection with specific embedding
weave cols create MyDocs --embedding text-embedding-3-small
weave cols create MyDocs -e text-embedding-ada-002

# Create documents with custom embedding
weave docs create MyDocs file.txt --embedding text-embedding-3-large
```

## Batch Processing
```bash
# Batch create documents from directory
weave docs batch --directory ./docs --collection MyDocs --parallel 3

# PDF processing with custom settings
weave docs create MyDocs document.pdf --chunk-size 1000
weave docs create MyDocs doc.pdf --skip-all-images
```

---

# Virtual Document View

The `--virtual` (`-w`) flag provides intelligent aggregation:

```bash
$ weave docs list MyCollection --virtual -S

✅ Found 3 virtual documents (aggregated from 15 total):

1. 📄 Document: research_paper.pdf
   📝 Chunks: 3/3
   🖼️ Images: 2
   📋 Metadata
     original_filename: research_paper.pdf
     file_size: 524288
     embedding: text-embedding-3-small
   📝 Chunk Details:
     1. ID: chunk-1 - Introduction to ML...
     2. ID: chunk-2 - Deep learning...
     3. ID: chunk-3 - Conclusion...
   🗂️ Stack Details:
     1. ID: image-1 (from page 2)
     2. ID: image-2 (from page 5)
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
- **embeddings** (emb) - List and explore embedding models

## Global Flags
- `--no-color` - Disable colored output
- `--no-truncate` - Show all data
- `--no-tips` - Suppress helpful tips
- `--no-confirm` - Skip confirmations (for automation)
- `--verbose` / `-v` - Detailed output
- `--quiet` / `-q` - Minimal output
- `--json` - JSON output format

---

# Database-Specific Flags

All commands support database selection:

```bash
# Target specific database
weave cols ls --weaviate
weave cols ls --supabase
weave cols ls --mongodb

# Multiple databases
weave cols ls --weaviate --supabase
weave cols ls --all

# Create in specific database
weave docs create WeaveDocs README.md --mongodb
weave docs create MyDocs file.txt --supabase
```

**Auto-detection**: If only one database is configured, it's used automatically

---

# Environment Variables

## Weaviate
```bash
WEAVIATE_URL="https://your-cluster.weaviate.cloud"
WEAVIATE_API_KEY="your-api-key"
```

## Supabase (Alpha)
```bash
SUPABASE_DATABASE_URL="postgresql://postgres:[password]@db.[ref].supabase.co:5432/postgres"
SUPABASE_DATABASE_KEY="your-supabase-anon-key"
```

## MongoDB Atlas (Experimental)
```bash
MONGODB_URI="mongodb+srv://username:password@cluster.mongodb.net/?appName=weave-cli"
MONGODB_DATABASE="weave-cli"
```

## Required for All
```bash
OPENAI_API_KEY="sk-..."  # For AI features and embeddings
```

---

# Development

## Project Structure
```
weave-cli/
├── src/                    # Source code
│   ├── cmd/               # CLI commands
│   ├── pkg/               # Public packages
│   │   ├── config/       # Configuration
│   │   ├── vectordb/     # Vector DB abstraction
│   │   │   ├── weaviate/ # Weaviate client
│   │   │   ├── supabase/ # Supabase adapter
│   │   │   ├── mongodb/  # MongoDB adapter
│   │   │   └── mock/     # Mock database
│   │   ├── llm/          # LLM clients
│   │   └── agents/       # AI agents
│   └── main.go           # Entry point
├── docs/                  # Documentation
├── tests/                 # Test files
└── bin/                   # Built binaries
```

---

# Development Workflow

```bash
# Setup development environment
./setup.sh

# Build everything
./build.sh

# Run tests
./test.sh

# Run specific database tests
./test.sh --mongodb
./test.sh --supabase
./test.sh --weaviate

# Run linter
./lint.sh

# Run security checks
govulncheck ./src/...
```

---

# Quality Assurance

## Testing
- ✅ **Unit tests** - All components tested
- ✅ **Integration tests** - Weaviate, MongoDB connectivity
- ✅ **Mock database** - No external dependencies
- ✅ **Edge cases** - Error handling and validation

## Code Quality
- ✅ **Go linting** - golangci-lint
- ✅ **Security scanning** - govulncheck, gosec
- ✅ **YAML validation** - yamllint
- ✅ **Markdown linting** - markdownlint
- ✅ **Shell checking** - shellcheck

---

# Recent Additions (v0.4.0)

## ✅ MongoDB Atlas Integration
- 🗄️ **Vector Search** - Semantic search with automatic embeddings
- 🔍 **BM25 Search** - Keyword search using text indexes
- 🔀 **Hybrid Search** - Vector + BM25 combination
- 📊 **Full CRUD** - Complete document and collection management
- 🆓 **Free Tier** - Works with MongoDB Atlas M0

## ✅ Earlier Features (v0.3.x)
- 🤖 **AI Agents** - Natural language queries (v0.3.0)
- 📊 **Embeddings** - List and manage models (v0.3.12)
- 🔧 **Config Commands** - Interactive create/update (v0.3.5)
- 🗄️ **Supabase** - PostgreSQL + pgvector support (v0.3.11)

---

# Version History Highlights

## v0.4.0 (2025-11-15)
- 🗄️ MongoDB Atlas vector search integration
- ✅ Automatic embedding generation
- 🔍 BM25 and hybrid search support
- 📚 Comprehensive MongoDB documentation

## v0.3.14 (2025-11-14)
- 📊 Embedding model display in listings
- 🔧 Demo recording improvements
- 📝 Documentation updates

## v0.3.12 (2025-11-13)
- 📊 Embeddings list command
- 🔍 Show embedding models for collections
- 📝 Per-document embedding selection

---

# Roadmap

## Upcoming (v0.5.0+)
- 🧪 **Enhanced Testing** - More integration tests for all databases
- 🤖 **More LLM Models** - Claude, Gemini support
- 📊 **Advanced Search** - Multi-vector search capabilities
- 🔗 **More Databases** - Milvus, Qdrant, Redis, Pinecone

## Medium Priority
- 💾 **Backup/Restore** - Export/import collections
- 📈 **Monitoring** - Database statistics and metrics
- 🎯 **Query Templates** - Reusable query patterns
- 🔄 **Migration Tools** - Database-to-database migration

See [Vector DB Integrations Planning](docs/planning/VECTOR_DB_INTEGRATIONS.md) for details

---

# Documentation

## Core Guides
- 📖 **[User Guide](docs/USER_GUIDE.md)** - Complete feature documentation
- 🗂️ **[VDB Support Matrix](docs/VDB_SUPPORT.md)** - Database comparison
- 📋 **[Changelog](docs/CHANGELOG.md)** - Version history

## Database-Specific
- 🌐 **[Weaviate Documentation](docs/)** - Weaviate setup and usage
- 🐘 **[Supabase Documentation](docs/supabase/)** - Supabase integration
- 🗄️ **[MongoDB Documentation](docs/mongodb/)** - MongoDB Atlas setup

## Advanced Topics
- 🤖 **[AI Agents Guide](docs/guides/WEAVE_CLI_AI.md)** - REPL and natural language
- 📦 **[Batch Processing](docs/guides/BATCH_DOCS_CREATION.md)** - Directory processing
- 📚 **[VDB Abstraction](docs/guides/VECTOR_DB_ABSTRACTION.md)** - Architecture details

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

## Adding New Vector Databases
See [Vector DB Abstraction Guide](docs/guides/VECTOR_DB_ABSTRACTION.md) for implementation details

---

# License & Acknowledgments

## License
This project is licensed under the **MIT License**

## Built With
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Weaviate Go Client](https://github.com/weaviate/weaviate-go-client) - Weaviate operations
- [MongoDB Go Driver](https://go.mongodb.org/mongo-driver) - MongoDB operations
- [OpenAI Go SDK](https://github.com/sashabaranov/go-openai) - Embedding generation
- Inspired by RAGme.io's tools/vdb.sh script

## Links
- 🐙 **[GitHub Repository](https://github.com/maximilien/weave-cli)**
- 📋 **[Issues & Discussions](https://github.com/maximilien/weave-cli/issues)**

---

# Questions?

## Get Help
- 📖 Read the [User Guide](docs/USER_GUIDE.md)
- 🗄️ Check [Database Documentation](docs/)
- 🐛 Report issues on [GitHub](https://github.com/maximilien/weave-cli/issues)
- 💬 Join discussions in [GitHub Discussions](https://github.com/maximilien/weave-cli/discussions)

## Thank You!
**Weave CLI** - Making multi-database vector management simple and powerful! 🚀

**Current Version**: v0.4.0 • **Supported Databases**: Weaviate, Supabase, MongoDB Atlas
