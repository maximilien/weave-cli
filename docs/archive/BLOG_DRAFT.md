NOTE: this is actually not a bad start but needs update to account for recent features. In paticular thinking of `weave stack` commands to deploy entire stack including VBD and LLMs and day-2 devops. Also lacking `weave evals` which is very important. Let's add these and I'll do a second pass.

# Weave CLI: Simplifying Vector Database Management for RAG Developers

**TL;DR**: Weave CLI is a powerful command-line tool for managing Weaviate vector
databases, with support for batch document processing, PDF extraction, and
AI-powered natural language operations. Built for RAG developers and vector
database administrators who need fast, reliable database management.

---

## The Problem: Vector Database Management Complexity

If you're building RAG (Retrieval-Augmented Generation) applications, you know
the pain: managing vector databases shouldn't be harder than using them. Whether
you're a developer prototyping a new RAG pipeline or a database administrator
managing production collections, you need tools that are:

- **Fast** - No waiting for slow web UIs or complex API calls
- **Reliable** - Batch operations that don't fail halfway through
- **Scriptable** - Easy to automate and integrate into CI/CD pipelines
- **Developer-friendly** - Clear commands, helpful errors, beautiful output

Traditional approaches fall short:

- **Web UIs** are slow and don't scale for bulk operations
- **Direct API calls** require boilerplate code for every operation
- **Custom scripts** become maintenance nightmares
- **Existing CLI tools** lack features like PDF processing and batch operations

We built Weave CLI to solve these problems.

---

## What is Weave CLI?

Weave CLI is a production-ready command-line tool for Weaviate vector databases,
written in Go for maximum performance. It provides:

- 🤖 **AI-Powered Interface** - Natural language queries with GPT-4o
- 📦 **Batch Processing** - Parallel document creation with automatic retry
- 📄 **PDF Support** - Intelligent text and image extraction
- 🔄 **Interactive REPL** - Beautiful interactive mode for exploration
- 🎭 **Mock Database** - Test without external dependencies
- 🔧 **Simple Configuration** - 3-line setup with sensible defaults

**Single binary. Zero dependencies. Works anywhere.**

---

## Installation & Quick Start

```bash
# Clone and build
git clone https://github.com/maximilien/weave-cli.git
cd weave-cli
./build.sh

# Interactive configuration (recommended)
./bin/weave config create --env

# Test connection
./bin/weave health check

# Start exploring
./bin/weave
```

That's it. No Docker containers, no virtual environments, no package managers.
Just a 46MB binary that works.

---

## Core Features for RAG Developers

### 1. Fast Collection Operations

```bash
# List all collections
weave cols ls

# Create text collection
weave cols create WeaveDocs --text

# Create image collection
weave cols create WeaveImages --image

# Show collection details with schema
weave cols show WeaveDocs --schema

# Query collection with semantic search
weave cols q WeaveDocs "machine learning concepts"

# Count documents
weave cols count
```

### 2. Document Management

```bash
# Create document from text file
weave docs create WeaveDocs document.txt

# Create from PDF (text + images)
weave docs create WeaveDocs research.pdf

# Text-only PDF processing (faster)
weave docs create WeaveDocs paper.pdf --skip-all-images

# List documents
weave docs ls WeaveDocs

# Show specific document
weave docs show WeaveDocs doc-id

# Delete document
weave docs del WeaveDocs doc-id
```

### 3. Batch Processing for Production

Process entire directories with parallel workers and automatic retry:

```bash
# Process directory with 5 parallel workers
weave docs batch --directory ./documents \
                 --collection WeaveDocs \
                 --parallel 5

# Process with image extraction
weave docs batch --directory ./pdfs \
                 --collection WeaveDocs \
                 --image-collection WeaveImages \
                 --parallel 3

# With progress tracking and CSV reports
weave docs batch --directory ./corpus \
                 --collection WeaveDocs \
                 --parallel 10 \
                 --verbose
```

Features:

- **Parallel processing** - Configurable worker count
- **Smart retry** - Automatic retry with exponential backoff
- **Progress tracking** - Visual progress with `.processed` files
- **Comprehensive reports** - CSV reports with detailed statistics
- **Resume support** - Skip already processed files

Perfect for:

- Initial database seeding
- Incremental updates
- Data migrations
- CI/CD pipelines

---

## Features for Database Administrators

### Configuration Management

Interactive configuration with secret masking:

```bash
# Create new configuration
weave config create --env

# Update existing configuration
weave config update --env

# View current configuration
weave config show

# List all configured databases
weave config list
```

Configuration precedence (highest to lowest):

1. Command-line flags (`--weaviate-url`, `--weaviate-api-key`)
2. Environment variables (`export WEAVIATE_URL=...`)
3. config.yaml (optional for advanced settings)
4. Built-in defaults

### Health Monitoring

```bash
# Check database health
weave health check

# With verbose output
weave health check --verbose

# JSON output for monitoring
weave health check --json
```

### Bulk Operations

```bash
# Delete all documents in collection
weave cols del WeaveDocs

# Delete all documents in all collections
weave cols da

# Delete collection schema (destructive)
weave cols ds WeaveDocs
```

All destructive operations require confirmation (skip with `--no-confirm` for
automation).

---

## Running Example: Building a RAG Knowledge Base

Let's walk through a complete example of setting up collections and populating
them with documents.

### Step 1: Setup Collections

```bash
# Create text collection for documents
weave cols create WeaveDocs --text

# Create image collection for PDF images
weave cols create WeaveImages --image

# Verify collections
weave cols ls
```

### Step 2: Add Documents

```bash
# Add individual documents
weave docs create WeaveDocs README.md
weave docs create WeaveDocs user_guide.pdf

# Check what was created
weave docs ls WeaveDocs
weave docs ls WeaveImages  # Images from PDF
```

### Step 3: Batch Process Directory

```bash
# Process entire documentation directory
weave docs batch --directory ./docs \
                 --collection WeaveDocs \
                 --image-collection WeaveImages \
                 --parallel 3 \
                 --verbose

# Results:
# ✅ 45 documents processed successfully
# 📊 CSV report: ./docs/.batch-report.csv
# ⏱️  Total time: 2m 15s
```

### Step 4: Query and Explore

```bash
# Search for content
weave cols q WeaveDocs "configuration examples"

# Show document with full metadata
weave docs show WeaveDocs doc-id --expand-metadata

# Count total documents
weave docs count WeaveDocs
weave docs count WeaveImages
```

**For detailed walkthroughs, see our
[documentation](https://github.com/maximilien/weave-cli/blob/main/docs/USER_GUIDE.md)
and [demo recordings](https://github.com/maximilien/weave-cli/blob/main/docs/DEMO.md).**

---

## Advanced PDF Processing

Weave CLI includes sophisticated PDF handling:

### Text Extraction

```bash
# Full extraction (text + images)
weave docs create WeaveDocs document.pdf

# Text-only mode (3x faster)
weave docs create WeaveDocs document.pdf --skip-all-images

# With custom image collection
weave docs create WeaveDocs paper.pdf --image-collection PaperImages
```

Features:

- **Intelligent chunking** - Respects paragraph boundaries
- **Image extraction** - Separate collection for images
- **CMYK support** - Automatic detection with conversion tips
- **Fallback handling** - Multiple extraction methods

### PDF Conversion

```bash
# Convert CMYK to RGB
weave docs pdf-convert document.pdf --rgb

# Using specific tool
weave docs pdf-convert document.pdf --ghostscript
weave docs pdf-convert document.pdf --imagemagick

# Batch convert directory
weave docs pdf-convert --directory ./pdfs --rgb
```

---

## AI-Powered Natural Language Interface

**The game-changer**: No need to memorize commands anymore.

### Interactive REPL Mode

```bash
# Start interactive mode
weave

# Use natural language
> show me all my collections
> create a collection called TestDocs
> add README.md to TestDocs
> find all empty collections
> check health
```

Features:

- Beautiful ASCII art banner
- Command history (saved to `~/.weave_history`)
- Context-aware suggestions
- CTRL-C to stop commands, twice to exit
- Built-in help with `/help`, `/examples`

### Direct Queries

```bash
# One-shot queries
weave query "show me all my collections"
weave q "check health status"
weave query "create TestDocs and TestImages collections"

# With options
weave query "delete old collections" --dry-run
weave q "batch process docs" --no-confirm
```

### Multi-Agent Architecture

Behind the scenes, 7 specialized agents work together:

1. **QueryAgent** - Validates and fixes queries
2. **PlanningAgent** - Creates execution plans
3. **WeaveAgent** - Executes weave commands via MCP
4. **BashAgent** - Safely executes shell commands
5. **OutputAgent** - Formats beautiful output
6. **ReportAgent** - Generates comprehensive reports
7. **EvalAgent** - Tracks metrics and evaluates success

With OpenTelemetry integration via Opik for full observability:

- Token usage tracking (prompt, completion, total)
- Cost tracking with color-coded display
- Direct links to Opik dashboard
- Complete trace history

**Perfect for**:

- Database exploration without reading docs
- Quick prototyping and testing
- Demos and presentations
- Onboarding new team members
- Automation with natural language

See the [AI Agents Guide](https://github.com/maximilien/weave-cli/blob/main/docs/WEAVE_CLI_AI.md)
for detailed documentation.

---

## Use Cases

### RAG Application Development

```bash
# Quick prototype testing
weave query "create TestCollection and add sample docs"
weave cols q TestCollection "test query"

# Production deployment
weave docs batch --directory ./knowledge_base \
                 --collection ProductionDocs \
                 --parallel 10
```

### Database Administration

```bash
# Health monitoring
weave health check --json | jq '.status'

# Bulk operations
weave cols da  # Clear all collections
weave docs batch --directory ./new_corpus

# Configuration management
weave config show
weave config update --env
```

### CI/CD Integration

```bash
#!/bin/bash
# Deploy new document set
weave docs batch \
  --directory ./latest_docs \
  --collection Docs \
  --parallel 5 \
  --no-confirm \
  --json > deploy.log

# Check status
if [ $? -eq 0 ]; then
  echo "Deployment successful"
else
  echo "Deployment failed"
  exit 1
fi
```

### Data Migration

```bash
# Export from old system (custom script)
./export_old_db.sh > old_data/

# Import to Weaviate
weave docs batch \
  --directory old_data/ \
  --collection MigratedDocs \
  --parallel 20 \
  --verbose
```

---

## Architecture & Design

### Performance

- **Single binary** - 46MB, no dependencies
- **Parallel processing** - Configurable worker pools
- **Smart retry** - Exponential backoff with jitter
- **Connection pooling** - Efficient resource usage
- **Minimal memory** - Streaming where possible

### Reliability

- **Progress tracking** - `.processed` files prevent duplicates
- **Automatic retry** - Handles transient failures
- **Comprehensive logging** - CSV reports for audit trails
- **Graceful degradation** - Fallback strategies for PDF processing

### Developer Experience

- **Beautiful output** - Color-coded, emoji-rich display
- **Helpful errors** - Actionable suggestions
- **Smart defaults** - Works out of the box
- **Flexible configuration** - Override at any level

### Testing

- **Mock database** - Test without Weaviate instance
- **Unit tests** - Comprehensive test coverage
- **Integration tests** - Real Weaviate connectivity
- **MCP integration tests** - Automated MCP compatibility testing

---

## Future Directions

We're actively working on several exciting features:

### Additional Vector Database Support

- **Milvus** - High-performance open source
- **Redis Vector Search** - Leverage existing Redis infrastructure
- **Pinecone** - Managed vector database service
- **Qdrant** - High-performance vector search engine
- **Chroma** - Embedding database for LLM apps

Unified interface across all databases with database-specific optimizations.

### Enhanced Schema Management

```bash
# Define schemas with full control
weave schema create MyCollection \
  --text-properties title,content \
  --number-properties score,timestamp \
  --vectorizer text2vec-openai

# Import/export schemas
weave schema export MyCollection > schema.yaml
weave schema import NewCollection schema.yaml

# Schema validation and migration
weave schema validate schema.yaml
weave schema migrate OldCollection NewCollection
```

### Advanced Search Strategies

- **Hybrid search** - Combine vector and keyword search
- **Multi-vector search** - Multiple embeddings per document
- **Filtered search** - Complex metadata filtering
- **Reranking** - Integrate reranking models
- **Custom scoring** - User-defined relevance functions

### Flexible Chunking Strategies

```bash
# Semantic chunking
weave docs create Docs paper.pdf --chunking semantic

# Fixed-size with overlap
weave docs create Docs paper.pdf --chunking fixed \
  --chunk-size 512 --overlap 50

# Custom chunking
weave docs create Docs paper.pdf --chunking-strategy custom \
  --chunking-config ./chunking.yaml
```

Strategies:

- Semantic chunking (respects meaning boundaries)
- Sentence-based chunking
- Token-based chunking
- Paragraph-based chunking
- Custom regex patterns
- LLM-based chunking

### Improved PDF Processing

- **Faster extraction** - Parallel page processing
- **Better accuracy** - Enhanced OCR with Tesseract
- **Table extraction** - Preserve table structure
- **Formula recognition** - Math notation support
- **Layout preservation** - Maintain document structure
- **Multi-column support** - Handle complex layouts

### Enhanced Metadata Handling

```bash
# Rich metadata during creation
weave docs create Docs paper.pdf \
  --metadata author="Smith et al." \
  --metadata year=2024 \
  --metadata tags="ML,AI,RAG"

# Metadata templates
weave docs batch --directory ./papers \
  --metadata-template academic.yaml

# Automatic metadata extraction
weave docs create Docs paper.pdf --extract-metadata
```

### Additional LLM Support

- **Claude** - Anthropic's Claude models
- **Gemini** - Google's multimodal models
- **Local models** - Ollama integration
- **Custom endpoints** - OpenAI-compatible APIs

### Observability Enhancements

- Detailed performance metrics
- Cost optimization suggestions
- Query pattern analysis
- Automated health alerts
- Integration with monitoring tools (Prometheus, Grafana)

---

## Why Weave CLI?

### For RAG Developers

- **Fast iteration** - Test ideas quickly without web UIs
- **Batch operations** - Process thousands of documents reliably
- **PDF support** - No need for separate preprocessing
- **AI interface** - Use natural language when prototyping
- **Scriptable** - Integrate into your development workflow

### For Database Administrators

- **Production-ready** - Reliable batch operations with retry
- **Configuration management** - Interactive setup with validation
- **Health monitoring** - JSON output for integration
- **Audit trails** - CSV reports for compliance
- **Automation-friendly** - Skip confirmations for CI/CD

### For Teams

- **Low barrier to entry** - AI mode requires no training
- **Consistent tooling** - Single binary across all platforms
- **Documentation** - Comprehensive guides and examples
- **Open source** - MIT licensed, contributions welcome

---

## Get Started

```bash
# Clone and build
git clone https://github.com/maximilien/weave-cli.git
cd weave-cli
./build.sh

# Interactive setup
./bin/weave config create --env

# Start exploring with AI
./bin/weave
```

**Resources:**

- [📖 User Guide](https://github.com/maximilien/weave-cli/blob/main/docs/USER_GUIDE.md)
  Comprehensive usage guide
- [🤖 AI Agents Guide](https://github.com/maximilien/weave-cli/blob/main/docs/WEAVE_CLI_AI.md)
  Natural language interface
- [📦 Batch Processing](https://github.com/maximilien/weave-cli/blob/main/docs/BATCH_DOCS_CREATION.md)
  Parallel processing guide
- [🎬 Demo Videos](https://github.com/maximilien/weave-cli/blob/main/docs/DEMO.md)
  Interactive demos
- [🐙 GitHub](https://github.com/maximilien/weave-cli) - Source and issues

---

## Contributing

We welcome contributions! Whether it's:

- Bug reports and feature requests
- Documentation improvements
- Code contributions
- Testing and feedback

See our [Contributing Guide](https://github.com/maximilien/weave-cli#contributing)
to get started.

---

## Conclusion

Vector databases are powerful, but managing them shouldn't be complicated.
Weave CLI brings simplicity, speed, and reliability to vector database
management, with an AI-powered interface that makes it accessible to everyone.

Whether you're building your first RAG application or managing production
vector databases at scale, Weave CLI has you covered.

**Try it today** and let us know what you think!

---

**Built with ❤️ by [Maximilien.ai](https://github.com/maximilien)**

**License**: MIT

**Version**: 0.3.5
