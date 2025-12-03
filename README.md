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

### Setup (Interactive - Recommended)

```bash
# Interactive configuration - fastest way to get started
# defaults to Weaviate vector database
weave config create --env

# Follow prompts to enter:
# - WEAVIATE_URL
# - WEAVIATE_API_KEY
# - OPENAI_API_KEY

# Verify setup
weave health check
```

### Supabase Setup (Alpha)

> **ℹ️ Alpha**: Supabase support is feature complete and functional.
> Recommended for development and testing. See the
> [Supabase Documentation](docs/supabase/) for comprehensive setup and usage
> guide.

To use Supabase as your vector database:

```bash
# Set Supabase configuration
export SUPABASE_DATABASE_URL="postgresql://postgres:[password]@db.[project-ref].supabase.co:5432/postgres"
export SUPABASE_DATABASE_KEY="your-supabase-anon-key"

# Configure weave to use Supabase
# Note: Use 'supabase-local' for consistency (v0.7.2+)
weave config create --database-type supabase

# Verify Supabase connection
weave health check supabase-local
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

### MongoDB Atlas Setup (Experimental)

> **🧪 Experimental**: MongoDB Atlas support is functional but requires manual
> vector search index setup. See the [MongoDB Documentation](docs/mongodb/)
> for complete setup guide.

To use MongoDB Atlas as your vector database:

```bash
# Set MongoDB configuration
export MONGODB_URI="mongodb+srv://username:password@cluster.mongodb.net/?appName=weave-cli"
export MONGODB_DATABASE="weave-cli"
export OPENAI_API_KEY="sk-..."  # Required for automatic embeddings

# Configure weave to use MongoDB
# Note: Use 'mongodb-local' for consistency (v0.7.2+)
weave config create --database-type mongodb

# Verify MongoDB connection
weave health check mongodb-local
```

**Important**: You must create a vector search index in the Atlas UI before
semantic search will work. See [ATLAS_SETUP.md](docs/mongodb/ATLAS_SETUP.md)
for detailed instructions.

### Chroma Setup

> **✅ Stable**: Chroma support is feature complete and fully tested. Supports
> both local development and cloud deployment. Uses Chroma Go SDK v2 API. See
> the [Chroma Documentation](docs/chroma/) for comprehensive setup guide.
>
> **⚠️ Platform Limitation**: Chroma is only supported on **macOS (AMD64/ARM64)**
> due to chroma-go v0.2.5 SDK's CGO dependency (libtokenizers). Linux and Windows
> are not supported. For other platforms, use Weaviate, Milvus, Qdrant, MongoDB,
> or Supabase.

To use Chroma locally:

```bash
# Start Chroma using podman (preferred) or docker
podman run -d --name chromadb -p 8000:8000 chromadb/chroma:0.6.2

# Set OpenAI API key for automatic embeddings
export OPENAI_API_KEY="sk-..."

# Configure weave to use local Chroma
weave config create --database-type chroma-local

# Verify Chroma connection
weave health check
```

To use Chroma Cloud:

```bash
# Set Chroma Cloud credentials
export CHROMA_CLOUD_URL="https://your-instance.chroma.cloud"
export CHROMA_CLOUD_API_KEY="your-api-key"
export OPENAI_API_KEY="sk-..."

# Configure weave to use Chroma Cloud
weave config create --database-type chroma-cloud

# Verify connection
weave health check
```

### Qdrant Setup (Experimental)

> **🧪 Experimental**: Qdrant support is newly added and functional. Supports
> both local development and cloud deployment. Requires testing with real
> Qdrant instances. See the [Qdrant Documentation](docs/qdrant/) for setup
> guide.

To use Qdrant locally:

```bash
# Start Qdrant using podman or docker
podman run -d --name qdrant -p 6333:6333 -p 6334:6334 \
  -v $(pwd)/qdrant_storage:/qdrant/storage:z \
  qdrant/qdrant

# Set OpenAI API key for automatic embeddings
export OPENAI_API_KEY="sk-..."

# Configure weave to use local Qdrant
weave config create --database-type qdrant-local

# Verify Qdrant connection
weave health check --qdrant-local
```

To use Qdrant Cloud:

```bash
# Set Qdrant Cloud credentials
export QDRANT_URL="https://your-cluster.cloud.qdrant.io:6334"
export QDRANT_API_KEY="your-api-key"
export OPENAI_API_KEY="sk-..."

# Configure weave to use Qdrant Cloud
weave config create --database-type qdrant-cloud

# Verify connection
weave health check --qdrant-cloud
```

### Neo4j Setup (Experimental)

> **🧪 Experimental**: Neo4j support is newly added and functional. Supports
> both local development and cloud deployment (Aura). Requires testing with real
> Neo4j instances. See the [Neo4j Documentation](docs/neo4j/) for setup guide.

To use Neo4j locally:

```bash
# Start Neo4j using podman or docker
podman run -d --name neo4j -p 7474:7474 -p 7687:7687 \
  -e NEO4J_AUTH=neo4j/yourpassword \
  neo4j:latest

# Set Neo4j credentials
export NEO4J_PASSWORD="yourpassword"
export OPENAI_API_KEY="sk-..."

# Configure weave to use local Neo4j
weave config create --database-type neo4j-local

# Verify Neo4j connection
weave health check --neo4j-local
```

To use Neo4j Cloud (Aura):

```bash
# Set Neo4j Aura credentials
export NEO4J_CLOUD_URL="neo4j+s://xxxxx.databases.neo4j.io"
export NEO4J_CLOUD_PASSWORD="your-password"
export OPENAI_API_KEY="sk-..."

# Configure weave to use Neo4j Aura
weave config create --database-type neo4j-cloud

# Verify connection
weave health check --neo4j-cloud
```

### Milvus Setup (Beta)

> **🧪 Beta**: Milvus support is feature complete and functional. Supports both
> local development and cloud deployment (Zilliz). See the
> [Milvus Documentation](docs/milvus/) for comprehensive setup guide.

To use Milvus locally:

```bash
# Start Milvus using podman (preferred) or docker
./tools/vdb/local/milvus.sh start

# Set OpenAI API key for automatic embeddings
export OPENAI_API_KEY="sk-..."

# Configure weave to use local Milvus
weave config create --database-type milvus-local

# Verify Milvus connection
weave health check
```

To use Milvus Cloud (Zilliz):

```bash
# Set Zilliz credentials
export MILVUS_CLOUD_ADDRESS="your-cluster.aws-us-west-2.vectordb.zillizcloud.com:19530"
export MILVUS_CLOUD_USERNAME="your-username"
export MILVUS_CLOUD_PASSWORD="your-password"
export OPENAI_API_KEY="sk-..."

# Configure weave to use Milvus Cloud
weave config create --database-type milvus-cloud

# Verify connection
weave health check
```

### Basic Usage

```bash
# List collections (all configured VDBs)
weave cols ls

# List collections from specific database types
weave cols ls --weaviate      # Weaviate only
weave cols ls --supabase      # Supabase only
weave cols ls --mongodb       # MongoDB Atlas only
weave cols ls --milvus-local  # Milvus local only
weave cols ls --milvus-cloud  # Milvus cloud (Zilliz) only
weave cols ls --chroma-local  # Chroma local only
weave cols ls --chroma-cloud  # Chroma cloud only
weave cols ls --qdrant-local  # Qdrant local only
weave cols ls --qdrant-cloud  # Qdrant cloud only
weave cols ls --neo4j-local   # Neo4j local only
weave cols ls --neo4j-cloud   # Neo4j cloud (Aura) only
weave cols ls --mock          # Mock database only
weave cols ls --all           # All configured databases

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
```

## Key Features

- 🤖 **AI-Powered** - AI Agent mode, natural language interface with GPT-4o
  multi-agent system
- ⚡ **Fast & Easy** - Written in Go with simple CLI and interactive REPL
  (AI Agent mode) with real-time progress feedback
- 🌐 **Flexible** - Weaviate Cloud, local instances, or built-in mock database
- 🔌 **Extensible** - Vector database abstraction layer supporting multiple
  backends (Weaviate, Milvus, Supabase PGVector, MongoDB Atlas, Chroma, Qdrant, Neo4j)
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

- **[🤖 AI Agents](docs/guides/WEAVE_CLI_AI.md)** - REPL mode with natural
  language query system
- **[📦 Batch Processing](docs/guides/BATCH_DOCS_CREATION.md)** - Directory
  processing guide
- **[📚 Vector DB Abstraction](docs/guides/VECTOR_DB_ABSTRACTION.md)** -
  Multi-database support architecture
- **[🎬 Demos](docs/guides/DEMO.md)** - Video demos and tutorials

### Database-Specific

- **[Chroma Documentation](docs/chroma/)** - Chroma integration guide (Beta)
- **[Milvus Documentation](docs/milvus/)** - Milvus integration guide (Beta)
- **[MongoDB Atlas Documentation](docs/mongodb/)** - MongoDB Atlas setup guide (Experimental)
- **[Neo4j Documentation](docs/neo4j/)** - Neo4j integration guide (Experimental)
- **[Qdrant Documentation](docs/qdrant/)** - Qdrant integration guide (Experimental)
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
|----------|------|--------|----------|------|
| **Weaviate Cloud** | `weaviate-cloud` | ✅ Production Ready | **Stable** | [Guide](docs/) |
| **Weaviate Local** | `weaviate-local` | ✅ Production Ready | **Stable** | [Guide](docs/) |
| **Milvus Local** | `milvus-local` | ✅ Functional | **Beta** - Feature complete, local testing ready | [Guide](docs/milvus/) |
| **Milvus Cloud** | `milvus-cloud` | ✅ Functional | **Beta** - Zilliz cloud integration ready | [Guide](docs/milvus/) |
| **Supabase** | `supabase` | ✅ Functional | **Alpha** - Feature complete, needs testing | [Guide](docs/supabase/) |
| **MongoDB Atlas** | `mongodb` | ✅ Functional | **Experimental** - Vector search requires index setup | [Guide](docs/mongodb/) |
| **Chroma Local** | `chroma-local` | ✅ Production Ready | **Stable** - Full CRUD, tested with v2 API | [Guide](docs/chroma/) |
| **Chroma Cloud** | `chroma-cloud` | ✅ Functional | **Beta** - Cloud integration (API key required) | [Guide](docs/chroma/) |
| **Mock** | `mock` | ✅ Testing Only | **Stable** | - |

### Maturity Levels

- **Stable**: Production-ready, well-tested, recommended for all use cases
- **Beta**: Feature complete, functional, ready for testing and feedback
- **Alpha**: Feature complete, functional, recommended for
  development/testing
- **Experimental**: Basic functionality working, may require manual setup,
  use with caution

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
