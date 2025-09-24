# Weave CLI

A command-line tool for managing Weaviate vector databases, written in Go.
This tool provides a fast and easy way to manage content in text and image
collections of configured vector databases, independently of specific
applications.

## Features

- 🌐 **Weaviate Cloud Support** - Connect to Weaviate Cloud instances
- 🏠 **Weaviate Local Support** - Connect to local Weaviate instances  
- 🎭 **Mock Database** - Built-in mock database for testing and development
- 📊 **Collection Management** - List, view, and delete collections
- 📄 **Document Management** - List, show, and delete individual documents
- 🔧 **Configuration Management** - YAML + Environment variable configuration
- 🎨 **Beautiful CLI** - Colored output with emojis and clear formatting
- 📋 **Virtual Document View** - Aggregate chunked documents by original file with cross-collection image support
- 🎯 **Smart Truncation** - Intelligent content truncation with `--no-truncate` option
- 🌈 **Color Control** - `--no-color` flag for terminal compatibility
- ⚡ **Fast & Lightweight** - Single binary deployment

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/maximilien/weave-cli.git
cd weave-cli

# Build the CLI
./build.sh

# The binary will be available at bin/weave
```

### Configuration

1. **Set up your environment variables**:

   ```bash
   export WEAVIATE_URL="your-weaviate-url.weaviate.cloud"
   export WEAVIATE_API_KEY="your-api-key"
   export VECTOR_DB_TYPE="weaviate-cloud"
   ```

2. **Test your connection**:

   ```bash
   ./bin/weave health check
   ```

3. **List your collections**:

   ```bash
   ./bin/weave collection list
   ```

### Basic Usage

```bash
# Show help
weave --help

# Check database health
weave health check

# List collections
weave collection list

# List documents in a collection
weave document list MyCollection

# Virtual document view (aggregate chunks by original file)
weave document list MyCollection --virtual

# Show all data without truncation
weave document list MyCollection --no-truncate

# Disable colored output (useful for scripts)
weave document list MyCollection --no-color
```

## Command Structure

Weave follows a consistent command pattern:
`weave noun verb [arguments] [flags]`

### Available Commands

- **config** - Configuration management
  - `weave config show` - Show current configuration

- **health** - Health and connectivity management  
  - `weave health check` - Check database health

- **collection** - Collection management
  - `weave collection list` - List all collections
  - `weave collection list --virtual` - Show collections with virtual structure summary
  - `weave collection delete COLLECTION_NAME` - Delete a specific collection
  - `weave collection delete-all` - Delete all collections

- **document** - Document management
  - `weave document list COLLECTION_NAME` - List documents in collection
  - `weave document list COLLECTION_NAME --virtual` - Virtual document view
  - `weave document show COLLECTION_NAME DOCUMENT_ID` - Show specific document
  - `weave document delete COLLECTION_NAME DOCUMENT_ID` - Delete specific document
  - `weave document delete-all COLLECTION_NAME` - Delete all documents in collection

### Command Aliases

For convenience, shorter aliases are available:

```bash
# Collection commands
weave col list          # Same as: weave collection list
weave cols list         # Same as: weave collection list

# Document commands  
weave doc list MyCol    # Same as: weave document list MyCol
weave docs list MyCol   # Same as: weave document list MyCol
```

## Global Flags

- `--no-color` - Disable colored output (useful for scripts/logs)
- `--no-truncate` - Show all data without truncation
- `--verbose` - Provide detailed output for debugging
- `--quiet` - Minimal output for scripts

## Document Display

Both regular and virtual document views feature consistent visual styling for
better readability:

### Regular Document View

```bash
$ weave document list MyCollection

✅ Found 6 documents in collection 'MyCollection':

1. 📄 ID: doc1-chunk1
   Content: This is the first chunk of a document about machine learning...
   📋 Metadata: 
     metadata: {"original_filename": "ml_guide.pdf", "is_chunked": true...}
     author: Test Author
```

### Virtual Document View

The `--virtual` flag provides an intelligent view by aggregating chunked content
back into original documents. **NEW**: Cross-collection aggregation automatically
includes images extracted from PDFs:

```bash
$ weave document list MyCollection --virtual

✅ Found 3 virtual documents in collection 'MyCollection' (aggregated from 15 total documents):

1. 📄 Document: research_paper.pdf
   📝 Chunks: 3/3
   🖼️ Images: 2
   📋 Metadata: 
     original_filename: research_paper.pdf
   📝 Chunk Details: 
     1. ID: chunk-1
        Content: Introduction to machine learning concepts...
     2. ID: chunk-2  
        Content: Deep learning architectures and applications...
     3. ID: chunk-3
        Content: Conclusion and future work...
   🗂️ Stack Details: 
     1. ID: image-1 (from page 2)
     2. ID: image-2 (from page 5)

2. 📄 Document: presentation.pptx
   🖼️ Images: 5
   📋 Metadata: 
     original_filename: presentation.pptx
   🗂️ Stack Details: 
     1. ID: slide-1-image
     2. ID: slide-3-chart
     3. ID: slide-5-diagram
     4. ID: slide-7-graph
     5. ID: slide-9-logo
```

**Key Features:**
- **Cross-collection aggregation**: Automatically finds and includes images from corresponding image collections
- **Smart grouping**: Images are grouped with their source documents (PDFs, presentations, etc.)
- **Complete view**: Shows both text chunks and extracted images in one unified view
- **Collection mapping**: Maps document collections to image collections (e.g., `MyDocs` → `MyImages`)

### Visual Styling

Both views feature consistent visual hierarchy:

- **Top-level keys** (ID, Chunks, Images, Content) are prominent
- **Metadata keys** are dimmed for better hierarchy
- **Important values** (IDs, filenames, numbers) are highlighted
- **Emojis** provide visual structure (disabled with `--no-color`)

## Database Support

- **Weaviate Cloud** - Full support with API key authentication
- **Weaviate Local** - Support for local instances (no auth required)
- **Mock Database** - Built-in mock database for testing and development

## Prerequisites

- Go 1.21 or later
- Access to a Weaviate instance (cloud or local)

## Documentation

For comprehensive documentation, examples, and advanced usage:

📖 **[Complete User Guide](docs/USER_GUIDE.md)** - Detailed usage instructions,
configuration examples, troubleshooting, and more.

## Development

### Building

```bash
# Build everything
./build.sh

# Run tests
./test.sh

# Run linter
./lint.sh
```

### Project Structure

```text
weave-cli/
├── src/                    # Source code
│   ├── cmd/               # CLI commands
│   ├── pkg/               # Public packages
│   │   ├── config/       # Configuration management
│   │   ├── weaviate/     # Weaviate client
│   │   └── mock/         # Mock database client
│   └── main.go           # Main entry point
├── docs/                   # Documentation
├── tests/                 # Test files
├── bin/                   # Built binaries
└── README.md             # This file
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run the test suite: `./test.sh`
6. Run the linter: `./lint.sh`
7. Submit a pull request

## License

This project is licensed under the MIT License - see the
[LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) CLI framework
- Uses [Weaviate Go Client](https://github.com/weaviate/weaviate-go-client) for database operations
- Inspired by RAGme.io's tools/vdb.sh script
