---
marp: true
theme: default
class: lead
paginate: true
backgroundColor: #fff
backgroundImage: url('https://marp.app/assets/hero-background.svg')
---

# Weave CLI
## Vector Database Management Made Simple

**A powerful command-line tool for managing Weaviate vector databases**

**Maximilien.ai** 

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

## Database Support
- **Weaviate Cloud** - Full support with API key authentication
- **Weaviate Local** - Support for local instances
- **Mock Database** - Built-in mock for testing

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
- **weaviate-cloud**: Weaviate Cloud service
- **weaviate-local**: Self-hosted instance
- **mock**: Built-in testing (no setup needed!)

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

# Recent Additions (v0.3.x)

## ✅ Completed Features
- 🤖 **AI Agents** - Natural language queries with GPT-4o (v0.3.0)
- 🔄 **Interactive REPL** - AI-powered interactive mode (v0.3.0)
- 📊 **Opik Integration** - LLM observability and cost tracking (v0.3.0)
- 🎨 **Enhanced UX** - Smart error detection, JSON highlighting (v0.3.0)
- 🔧 **Simplified Config** - 3-line setup with sensible defaults (v0.3.2)
- 📝 **Config Commands** - Interactive create/update commands (v0.3.5)
- 📦 **Batch Processing** - Parallel document creation (v0.2.11)
- 📄 **PDF Processing** - Text extraction and CMYK conversion (v0.2.14)

---

# Future Roadmap

## High Priority
- 🧪 **Enhanced Testing** - Unit tests for all agents
- 🤖 **More LLM Models** - Claude, Gemini support
- 📊 **Collection Schema Management** - Advanced schema operations
- 🔍 **Advanced Search** - Vector similarity with filters

## Medium Priority
- 💾 **Backup/restore** - Export/import collections
- 📈 **Monitoring** - Database statistics and metrics
- 🔗 **More Databases** - Pinecone, Qdrant, Chroma support
- 🎯 **Query Templates** - Reusable query patterns

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