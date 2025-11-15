# Weave CLI

A fast, AI-powered command-line tool for managing Weaviate vector databases.
Built in Go for performance and ease of use.

## Quick Start

### Installation

```bash
git clone https://github.com/maximilien/weave-cli.git
cd weave-cli
./build.sh
# Binary available at bin/weave
```

### Setup (Interactive - Recommended)

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

### Supabase Setup (EXPERIMENTAL)

> **⚠️ EXPERIMENTAL**: Supabase support is currently experimental and under active
> development. Some features may not work as expected. Please report issues at
> <https://github.com/maximilien/weave-cli/issues>

To use Supabase as your vector database:

```bash
# Set Supabase configuration
export SUPABASE_DATABASE_URL="postgresql://postgres:[password]@db.[project-ref].supabase.co:5432/postgres"
export SUPABASE_DATABASE_KEY="your-supabase-anon-key"

# Configure weave to use Supabase
weave config create --database-type supabase

# Verify Supabase connection
weave health check
```

**Important Notes**:

1. **IPv6 Requirement**: Supabase database endpoints are IPv6-only. If your
   network doesn't support IPv6, use the connection pooler instead:

   ```bash
   # Get pooler URL from: Project Settings → Database → Connection Pooling
   export SUPABASE_DATABASE_URL="postgresql://postgres.[project].[string]:[password]@aws-0-[region].pooler.supabase.com:6543/postgres"
   ```

2. **pgvector Extension**: Ensure your Supabase project has pgvector enabled:

   ```sql
   CREATE EXTENSION IF NOT EXISTS vector;
   ```

### Basic Usage

```bash
# List collections
weave cols ls

# List collections from specific database types
weave cols ls --weaviate    # Weaviate only
weave cols ls --supabase    # Supabase only
weave cols ls --mock        # Mock database only
weave cols ls --all         # All configured databases

# Create a collection
weave cols create MyCollection --text

# Add documents
weave docs create MyCollection document.txt
weave docs create MyCollection document.pdf

# Search with natural language
weave cols q MyCollection "search query"

# AI-powered REPL mode
weave
> show me all my collections
> create TestDocs collection
> add README.md to TestDocs

# List available embeddings
weave embeddings list
weave emb ls --verbose

# Create collection with specific embedding (used as default for all documents)
weave cols create MyCollection --embedding text-embedding-3-small
weave cols create MyCollection -e text-embedding-ada-002
```

## Key Features

- 🤖 **AI-Powered** - Natural language interface with GPT-4o multi-agent system
- ⚡ **Fast & Easy** - Written in Go with simple CLI and interactive REPL with
  real-time progress feedback
- 🌐 **Flexible** - Weaviate Cloud, local instances, or built-in mock database
- 🔌 **Extensible** - Vector database abstraction layer supporting multiple
  backends (Supabase PGVector implemented, Milvus planned)
- 📦 **Batch Processing** - Parallel processing of entire directories
- 📄 **PDF Support** - Intelligent text extraction and image processing
- 🔍 **Semantic Search** - Vector-based similarity search with natural
  language
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

- **[🤖 AI Agents](docs/guides/WEAVE_CLI_AI.md)** - Natural language query
  system
- **[📦 Batch Processing](docs/guides/BATCH_DOCS_CREATION.md)** - Directory
  processing guide
- **[📚 Vector DB Abstraction](docs/guides/VECTOR_DB_ABSTRACTION.md)** -
  Multi-database support architecture
- **[🎬 Demos](docs/guides/DEMO.md)** - Video demos and tutorials

### Database-Specific

- **[Supabase Documentation](docs/supabase/)** - Supabase integration guide
- **[Weaviate Documentation](docs/weaviate/)** - Weaviate integration status

## Advanced Usage

### Configuration Options

#### Auto-Configuration

Weave CLI automatically detects missing configuration:

```bash
# Try any command - you'll get prompted to configure interactively
weave cols ls

# Or install weave-mcp for REPL mode
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
allows seamless switching between different vector database backends:

### Currently Supported

- **Weaviate Cloud** (`weaviate-cloud`) - Production-ready cloud instances ✅
- **Weaviate Local** (`weaviate-local`) - Self-hosted Weaviate instances ✅
- **Supabase PGVector** (`supabase`) - PostgreSQL with pgvector extension ⚠️
  EXPERIMENTAL
- **Mock Database** (`mock`) - Built-in testing database (no external
  dependencies) ✅

### Planned Support (Path to v1.0.0)

- **MongoDB** (`mongodb`) - ✅ v0.3.15 - Atlas Vector Search, automatic embeddings, hybrid search ([docs](docs/mongodb/))
- **Milvus** (`milvus`) - v0.4.0 - Open source, BM25 + hybrid search, geospatial
- **Qdrant** (`qdrant`) - v0.5.0 - High-performance gRPC, HNSW + quantization
- **Redis** (`redis`) - v0.6.0 - In-memory speed, RediSearch, best hybrid search
- **Pinecone** (`pinecone`) - v0.8.0 - Fully managed, serverless, generous free tier

See [Vector DB Integrations Planning](docs/planning/VECTOR_DB_INTEGRATIONS.md) for
detailed implementation plans and
[Vector DB Abstraction Documentation](docs/VECTOR_DB_ABSTRACTION.md) for
architecture details.

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

## License

MIT License - see [LICENSE](LICENSE) file for details.
