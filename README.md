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

### Basic Usage

```bash
# List collections
weave cols ls

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
```

## Key Features

- 🤖 **AI-Powered** - Natural language interface with GPT-4o multi-agent system
- ⚡ **Fast & Easy** - Written in Go with simple CLI and interactive REPL
- 🌐 **Flexible** - Weaviate Cloud, local instances, or built-in mock database
- 📦 **Batch Processing** - Parallel processing of entire directories
- 📄 **PDF Support** - Intelligent text extraction and image processing
- 🔍 **Semantic Search** - Vector-based similarity search with natural language
- 📊 **Embeddings** - List and explore available embedding models
- ⏱️ **Configurable Timeouts** - Default 10s timeout, adjustable per command (e.g., `--timeout 5s`)

## Documentation

- **[📖 User Guide](docs/USER_GUIDE.md)** - Complete feature documentation
- **[🤖 AI Agents](docs/WEAVE_CLI_AI.md)** - Natural language query system
- **[📦 Batch Processing](docs/BATCH_DOCS_CREATION.md)** - Directory processing guide
- **[🎬 Demos](docs/DEMO.md)** - Video demos and tutorials
- **[📋 Changelog](docs/CHANGELOG.md)** - Version history and updates

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

See the [User Guide](docs/USER_GUIDE.md#configuration) for detailed
configuration options.

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
```

## Database Support

- **Weaviate Cloud** - Production-ready cloud instances
- **Weaviate Local** - Self-hosted Weaviate instances
- **Mock Database** - Built-in testing database (no external dependencies)

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

- **[GitHub Repository](https://github.com/maximilien/weave-cli)**
- **[Full Demo (5 min)](https://asciinema.org/a/LrKzmThBfDbTPISZzr8biP4dt)**
- **[Quick Demo (2 min)](https://asciinema.org/a/HiAU7h1iJvZ2QdJe70ae3Cc0b)**
- **[REPL Demo](https://asciinema.org/a/U504HN4FSeMsOA0qS0os0NWUE)**

## License

MIT License - see [LICENSE](LICENSE) file for details.
