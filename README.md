# Weave CLI

A command-line tool for managing Weaviate vector databases, written in Go.
This tool provides a fast and easy way to manage content in text and image
collections of configured vector databases.

## 🚀 What's New in v0.3.0

### Latest Updates (Pending Release)

- **📦 MCP Integration Tests**: Automated test suite for weave-mcp compatibility
  - Comprehensive integration tests for all MCP operations
  - Collection and document operation testing
  - Error handling and edge case validation
  - Automated testing workflow for MCP releases
  - See [tests/README_MCP_TESTS.md](tests/README_MCP_TESTS.md) for details
- **🎨 Version Display in Banner**: REPL banner now shows version info
  - Displays version string in dimmed text below ASCII art
  - Consistent with `weave -V` output format

### Core Features (v0.3.0)

- **🔄 Interactive REPL Mode**: Run `weave` without arguments for an
  interactive session
  - Beautiful ASCII art banner with version and GitHub link
  - Natural language query support in interactive mode
  - Built-in help, examples, and history commands
  - CTRL-C to stop commands, twice to exit (like Claude CLI)
  - Command history saved to `~/.weave_history`
- **🤖 AI Agents Query Command**: New `weave query` (or `weave q`) command
  for natural language queries using GPT-4o
  - Automatically understands your intent and plans appropriate commands
  - Executes weave-cli and bash commands intelligently
  - Provides comprehensive reports with recommendations
  - Supports dry-run mode to preview execution plans
- **🧠 Multi-Agent Architecture**: 7 specialized agents working together
  - QueryAgent: Validates and fixes user queries with MCP tool awareness
  - PlanningAgent: Creates detailed execution plans with full tool schemas
  - WeaveAgent: Executes weave-cli commands via MCP protocol
  - BashAgent: Safely executes bash commands with user approval
  - OutputAgent: Beautiful, color-coded user-friendly output
  - ReportAgent: Comprehensive operation reports with
    LLM-generated recommendations
  - EvalAgent: Tracks metrics and evaluates success
- **📊 OpenTelemetry Integration**: Opik tracing for LLM observability
  - Automatic cost tracking for all LLM calls
  - Token usage metrics (prompt, completion, total)
  - Color-coded cost display (green/yellow/red)
  - Direct link to Opik dashboard for detailed traces
- **✨ Enhanced User Experience**:
  - Smart error detection in command output
  - Step progress with duration display
  - JSON output with syntax highlighting
  - Auto-approval for simple weave commands
  - Health check support via natural language

### Previous Updates (v0.2.14)

- **🔄 PDF Conversion Tool**: New `weave docs pdf-convert` command to convert CMYK
  PDFs to RGB format using Ghostscript or ImageMagick
- **🎯 Text-Only PDF Processing**: New `--skip-all-images` flag to extract only
  text from PDFs without image processing overhead
- **💬 Helpful Tips Control**: Global `--no-tips` flag to suppress helpful tips
  and suggestions when desired
- **🔧 CMYK PDF Support**: Graceful handling of CMYK PDFs with actionable tips
  for conversion using Ghostscript or ImageMagick

### Previous Updates (v0.2.11)

- **📦 Batch Document Creation**: Process entire directories with parallel
  processing and automatic retry
- **⚡ Parallel Processing**: Configure multiple workers for faster batch
  operations
- **📊 Progress Tracking**: Visual progress indicators with time estimation
- **🔄 Smart Retry**: Automatic retry on failures with `.processed` file
  tracking
- **📈 Comprehensive Reporting**: CSV reports with detailed processing statistics

### Previous Updates (v0.2.10)

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

- 🤖 **AI Agents** - Natural language queries with GPT-4o powered multi-agent system
- 🌐 **Weaviate Cloud Support** - Connect to Weaviate Cloud instances
- 🏠 **Weaviate Local Support** - Connect to local Weaviate instances
- 🎭 **Mock Database** - Built-in mock database for testing and development
- 📊 **Collection Management** - List, create, view, and delete collections
- 📄 **Document Management** - Create, update, list, show, and delete documents
- 📦 **Batch Processing** - Process entire directories with parallel workers and
  automatic retry
- 🔍 **Semantic Search** - Query collections with natural language
- 📄 **PDF Processing** - Extract text and images from PDFs with intelligent
  chunking and CMYK support
- 🔄 **PDF Conversion** - Convert CMYK PDFs to RGB format with Ghostscript or
  ImageMagick
- 🎯 **Flexible Processing** - Text-only mode with `--skip-all-images` for faster
  processing
- 💬 **User Experience** - Helpful tips and suggestions (can be disabled with
  `--no-tips`)
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

**Fastest Way to Get Started** (3 lines, no config files needed!):

```bash
# Copy .env.example and add your credentials
cp .env.example .env

# Edit .env with your 3 required credentials:
# - WEAVIATE_URL
# - WEAVIATE_API_KEY
# - OPENAI_API_KEY

# That's it! Start using weave:
weave health check
```

**Configuration Precedence** (highest to lowest):

1. **Command-line flags** - `weave query --model gpt-4`
2. **Environment variables** - `export OPENAI_MODEL=gpt-4`
3. **config.yaml** (optional) - For advanced customization
4. **Built-in defaults** - Sensible defaults work out of the box

**Advanced Configuration** (optional):

For fine-tuning collection names, batch sizes, PDF settings, etc.:

```bash
cp config.yaml.example config.yaml
# Edit config.yaml to customize defaults
```

### Interactive REPL Mode

Simply run `weave` without any arguments to start the interactive mode:

```bash
weave

# You'll see the ASCII art banner and prompt:
 __      __
/  \    /  \ ____ _____ ___  __ ____
\   \/\/   // __ \\__  \\  \/ // __ \
 \        /\  ___/ / __ \\   /\  ___/
  \__/\  /  \___  >____  /\_/  \___  >
       \/       \/     \/          \/

  Weave CLI - AI-Powered Vector Database Management
  https://github.com/maximilien/weave-cli

  Use natural language to manage your vector databases
  Type /help for commands, /exit to quit
  Press CTRL-C to stop current command, twice to exit

>
```

**REPL Commands:**

- `/help` - Show help and available commands
- `/examples` - Show example natural language queries
- `/history` - Show command history location
- `/clear` - Clear the screen
- `/exit` - Exit the REPL

**Keyboard Shortcuts:**

- `CTRL-C` - Stop current command
- `CTRL-C` (twice) - Exit the REPL
- `CTRL-D` - Exit the REPL
- `↑/↓` - Navigate command history

### Basic Usage

```bash
# Check database health
weave health check

# List collections
weave cols ls

# Create a collection
weave cols create MyCollection --text

# Add documents
weave docs create MyCollection document.txt
weave docs create MyCollection document.pdf

# Add PDF with text-only extraction (faster, no images)
weave docs create MyCollection document.pdf --skip-all-images

# Process with images in separate collection
weave docs create MyCollection document.pdf --image-collection MyImages

# Suppress helpful tips during processing
weave docs create MyCollection document.pdf --no-tips

# Convert CMYK PDF to RGB format
weave docs pdf-convert document.pdf --ghostscript
weave docs pdf-convert document.pdf --rgb  # auto-detect tool

# Batch process documents
weave docs batch --directory ./docs --collection MyCollection --parallel 3

# Search documents
weave cols q MyCollection "search query"

# List documents
weave docs ls MyCollection

# 🤖 AI Agents - Natural Language Queries
weave query "show me all my collections"
weave q "find all empty collections"
weave query "create TestDocs and TestImages collections"
weave q "check health"
weave query "add files in /tmp/docs to MyCollection" --dry-run
weave q "convert CMYK PDFs in ./pdfs directory" --verbose
```

## Documentation

- **[🤖 AI Agents Guide](docs/WEAVE_CLI_AI.md)** - Natural language query
  system with multi-agent architecture
- **[📖 User Guide](docs/USER_GUIDE.md)** - Comprehensive usage guide with
  examples
- **[📦 Batch Processing Guide](docs/BATCH_DOCS_CREATION.md)** - Batch document
  creation with parallel processing
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
