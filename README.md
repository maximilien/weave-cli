# Weave CLI

A command-line tool for managing Weaviate vector databases, written in Go.
This tool provides a fast and easy way to manage content in text and image
collections of configured vector databases.

## 🚀 What's New in v0.2.10

- **📄 Enhanced PDF Processing**: Improved PDF text extraction with better
  fallback handling
- **💬 Human-Friendly Error Messages**: Simplified, actionable error messages
  with helpful suggestions
- **🎬 Updated Demos**: New demo recordings showcasing PDF processing
  capabilities
- **🔧 Better UX**: Fixed PDF success message formatting and improved user
  experience

## 🎬 Demo

Watch Weave CLI in action with our interactive demos:

- **📹 [Full Demo](https://asciinema.org/a/feoupxMVzhHaNIzmWGTEuLpCR)**
  (5 minutes): Complete feature showcase with PDF processing
- **⚡ [Quick Demo](https://asciinema.org/a/qUgPvpBpqsJwlVrjVWVtnonKX)**
  (2 minutes): Rapid overview with environment variables

## Features

- 🌐 **Weaviate Cloud Support** - Connect to Weaviate Cloud instances
- 🏠 **Weaviate Local Support** - Connect to local Weaviate instances
- 🎭 **Mock Database** - Built-in mock database for testing and development
- 📊 **Collection Management** - List, create, view, and delete collections
- 📄 **Document Management** - Create, update, list, show, and delete documents
- 🔍 **Semantic Search** - Query collections with natural language
- 📄 **PDF Processing** - Extract text from PDF files with intelligent chunking
- 🔧 **Configuration Management** - YAML + Environment variable configuration
- 🎨 **Beautiful CLI** - Colored output with emojis and clear formatting

## Quick Start

### Installation

```bash
# Clone and build
git clone https://github.com/maximilien/weave-cli.git
cd weave-cli
./build.sh

# The binary will be available at bin/weave
```

### Configuration

#### Option 1: Environment Variables Only (Recommended)

```bash
export WEAVIATE_URL="https://your-cluster.weaviate.cloud"
export WEAVIATE_API_KEY="your-weaviate-api-key"
export OPENAI_API_KEY="sk-proj-your-openai-key"
```

#### Option 2: Configuration File

```bash
# Copy example config
cp config.yaml.example config.yaml
# Edit config.yaml with your settings
```

### Basic Usage

```bash
# Check database health
./bin/weave health check

# List collections
./bin/weave cols ls

# Create a collection
./bin/weave cols create MyCollection --text

# Add documents
./bin/weave docs create MyCollection document.txt
./bin/weave docs create MyCollection document.pdf

# Search documents
./bin/weave cols q MyCollection "search query"

# List documents
./bin/weave docs ls MyCollection
```

## Documentation

- **[📖 User Guide](docs/USER_GUIDE.md)** - Comprehensive usage guide with
  examples
- **[🎬 Demo Guide](docs/DEMO.md)** - Interactive demo scripts and recordings
- **[⚠️ Error Messages](docs/ERROR_MESSAGES.md)** - Human-friendly error message
  examples
- **[📋 Configuration Guide](docs/USER_GUIDE.md#configuration)** - Detailed
  configuration options
- **[🔍 Search Guide](docs/USER_GUIDE.md#semantic-search)** - Advanced search
  capabilities
- **[📄 PDF Processing](docs/USER_GUIDE.md#pdf-processing)** - PDF text
  extraction features

## Database Support

- **Weaviate Cloud** - Production-ready cloud instances
- **Weaviate Local** - Self-hosted Weaviate instances
- **Mock Database** - Built-in testing database (no external dependencies)

## Development

### Setup Development Environment

```bash
# Install all required development tools (linters, PDF tools, etc.)
./setup.sh
```

This will install:

- Go tools: `golangci-lint`, `goimports`, `govulncheck`, `gosec`
- PDF processing: `poppler` (provides `pdftotext` for better PDF text extraction)
- Shell linting: `shellcheck`
- YAML linting: `yamllint`
- Markdown linting: `markdownlint`
- Dependency checking: `go-mod-outdated`

### Build and Test

```bash
# Build the project
./build.sh

# Run tests
./test.sh

# Run linting
./lint.sh

# Record demos
./tools/asciinema.sh demo
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests and linting
5. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

Built with ❤️ by [github.com/maximilien](https://github.com/maximilien)
