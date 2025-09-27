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

- 🌐 **Fast & Lightweight** - Single binary deployment
- 🎨 **Beautiful CLI** - Colored output with emojis and clear formatting
- 🔧 **Easy Configuration** - YAML + Environment variables
- 📊 **Collection Management** - List, view, and delete collections
- 📄 **Document Management** - List, show, and delete documents
- 🎭 **Mock Database** - Built-in testing and development support

---

# Key Features

## Database Support
- **Weaviate Cloud** - Full support with API key authentication
- **Weaviate Local** - Support for local instances
- **Mock Database** - Built-in mock for testing

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

# Configure environment (multiple options)

# Option 1: Environment variables
export WEAVIATE_URL="your-weaviate-url.weaviate.cloud"
export WEAVIATE_API_KEY="your-api-key"
export VECTOR_DB_TYPE="weaviate-cloud"

# Option 2: Command-line flags (highest priority)
./bin/weave --vector-db-type weaviate-cloud \
            --weaviate-url "your-weaviate-url.weaviate.cloud" \
            --weaviate-api-key "your-api-key" \
            health check

# Option 3: .env file
echo "VECTOR_DB_TYPE=weaviate-cloud" > .env
echo "WEAVIATE_URL=your-weaviate-url.weaviate.cloud" >> .env
echo "WEAVIATE_API_KEY=your-api-key" >> .env
```

---

# Configuration Priority

**Priority Order** (highest to lowest):

1. **Command-line flags** (`--vector-db-type`, `--weaviate-url`, `--weaviate-api-key`)
2. **`--env` file** (specified with `--env` flag)
3. **`.env` file** (in current directory)
4. **Shell environment variables**

```bash
# Example: Override everything
weave --vector-db-type mock \
      --weaviate-url https://custom.weaviate.cloud \
      --weaviate-api-key custom-key \
      collection list
```

---

# Basic Commands

```bash
# Health check
weave health check

# List collections
weave collection list

# List documents
weave document list MyCollection

# Virtual document view
weave document list MyCollection --virtual

# Show specific document
weave document show MyCollection doc-id

# Show document by filename
weave document show MyCollection --name test_image.png

# Delete document by filename
weave document delete MyCollection --name test_image.png
```

---

# Command Structure

Weave follows a consistent pattern:
**`weave noun verb [arguments] [flags]`**

## Available Commands
- **config** - Configuration management
- **health** - Health and connectivity
- **collection** - Collection management
- **document** - Document management

## Global Flags
- `--no-color` - Disable colored output
- `--no-truncate` - Show all data
- `--verbose` - Detailed output
- `--quiet` - Minimal output

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

# Configuration

## Quick Setup
```bash
# Copy example files
cp config.yaml.example config.yaml
cp .env.example .env

# Edit with your values
nano .env
nano config.yaml
```

## Environment Variables (.env)
```bash
# Required for weaviate-cloud
VECTOR_DB_TYPE="weaviate-cloud"
WEAVIATE_URL="https://your-cluster.weaviate.cloud"
WEAVIATE_API_KEY="your-api-key"
OPENAI_API_KEY="sk-proj-your-openai-key"

# Optional collection names
WEAVIATE_COLLECTION="MyCollection"
WEAVIATE_COLLECTION_IMAGES="MyImages"
```

## Config File (config.yaml)
```yaml
databases:
  default: ${VECTOR_DB_TYPE:-weaviate-cloud}
  vector_databases:
    - name: weaviate-cloud
      type: weaviate-cloud
      url: ${WEAVIATE_URL}
      api_key: ${WEAVIATE_API_KEY}
      openai_api_key: ${OPENAI_API_KEY}
      collections:
        - name: ${WEAVIATE_COLLECTION:-WeaveDocs}
          type: text
```

## Database Types
- **weaviate-cloud**: Paid Weaviate Cloud service
- **weaviate-local**: Local Weaviate instance
- **mock**: Built-in testing database

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

# Future Roadmap

## High Priority
- 📄 **Document creation/upload** - Add new documents
- 🔍 **Search functionality** - Content and vector search
- 📊 **Collection schema management** - Create/modify schemas
- 🛠️ **Better error handling** - Improved UX

## Medium Priority
- 💾 **Backup/restore** - Export/import collections
- 📈 **Monitoring** - Database statistics and metrics
- 🔗 **More databases** - Pinecone, Qdrant, Chroma support

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