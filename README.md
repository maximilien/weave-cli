# Weave CLI

A command-line tool for managing Weaviate vector databases, written in Go. This tool provides a fast and easy way to manage content in text and image collections of configured vector databases, independently of specific applications.

## Features

- 🌐 **Weaviate Cloud Support** - Connect to Weaviate Cloud instances
- 🏠 **Weaviate Local Support** - Connect to local Weaviate instances  
- 🎭 **Mock Database** - Built-in mock database for testing and development
- 📊 **Collection Management** - List, view, and delete collections
- 📄 **Document Management** - List, show, and delete individual documents (shows document IDs and basic info)
- 🔧 **Configuration Management** - YAML + Environment variable configuration
- 🎨 **Beautiful CLI** - Colored output with emojis and clear formatting
- ⚡ **Fast & Lightweight** - Single binary deployment

## Installation

### Build from Source

```bash
# Clone the repository
git clone https://github.com/maximilien/weave-cli.git
cd weave-cli

# Build the CLI
./build.sh

# The binary will be available at bin/weave
```

### Prerequisites

- Go 1.21 or later
- Access to a Weaviate instance (cloud or local)

## Configuration

Weave CLI uses two configuration files:

### config.yaml

```yaml
# Database Configuration
database:
  # Vector Database settings
  vector_db:
    type: "${VECTOR_DB_TYPE:-weaviate-cloud}"  # Options: weaviate-cloud, weaviate-local, mock
    
    # Weaviate Cloud configuration
    weaviate_cloud:
      url: "${WEAVIATE_URL}"
      api_key: "${WEAVIATE_API_KEY}"
      collection_name: "${WEAVIATE_COLLECTION:-WeaveDocs}"
      collection_name_test: "${WEAVIATE_COLLECTION_TEST:-WeaveDocs_test}"
      
    # Weaviate Local configuration
    weaviate_local:
      url: "http://localhost:8080"
      collection_name: "${WEAVIATE_COLLECTION:-WeaveDocs}"
      collection_name_test: "${WEAVIATE_COLLECTION_TEST:-WeaveDocs_test}"
      
    # Mock Vector Database configuration (for development/testing)
    mock:
      enabled: true
      simulate_embeddings: true
      embedding_dimension: 384
      collections:
        - name: "WeaveDocs"
          type: "text"
          description: "Mock text documents collection"
        - name: "WeaveImages" 
          type: "image"
          description: "Mock image documents collection"
```

### .env

```bash
# Vector Database Configuration  
VECTOR_DB_TYPE="weaviate-cloud"
WEAVIATE_COLLECTION="WeaveDocs"
WEAVIATE_COLLECTION_TEST="WeaveDocs_test"

WEAVIATE_URL="your-weaviate-url.weaviate.cloud"
WEAVIATE_API_KEY="your-api-key"
```

## Usage

### Basic Commands

```bash
# Show help
weave help

# Show currently configured VDB
weave config show

# Check health of VDB connections
weave health check

# List all collections
weave collection list

# List documents in a collection (shows document IDs)
weave document list MyCollection

# Show a specific document
weave document show MyCollection document-id

# Delete a specific document
weave document delete MyCollection document-id

# Delete all documents in a collection (⚠️ DESTRUCTIVE)
weave document delete-all MyCollection

# Delete a specific collection
weave collection delete MyCollection

# Delete all collections (⚠️ DESTRUCTIVE)
weave collection delete-all
```

### Command Structure

Weave follows a consistent command pattern: `weave noun verb [arguments] [flags]`

- **config** - Configuration management
  - `weave config show` - Show current configuration

- **health** - Health and connectivity management  
  - `weave health check` - Check database health

- **collection** - Collection management
  - `weave collection list` - List all collections
  - `weave collection delete COLLECTION_NAME` - Delete a specific collection
  - `weave collection delete-all` - Delete all collections

- **document** - Document management
  - `weave document list COLLECTION_NAME` - List documents in collection
  - `weave document show COLLECTION_NAME DOCUMENT_ID` - Show specific document
  - `weave document delete COLLECTION_NAME DOCUMENT_ID` - Delete specific document
  - `weave document delete-all COLLECTION_NAME` - Delete all documents in collection

### Command Aliases

For convenience, shorter aliases are available:

```bash
# Collection commands
weave col list          # Same as: weave collection list
weave cols list         # Same as: weave collection list
weave col delete MyCol  # Same as: weave collection delete MyCol

# Document commands  
weave doc list MyCol    # Same as: weave document list MyCol
weave docs list MyCol   # Same as: weave document list MyCol
weave doc show MyCol ID # Same as: weave document show MyCol ID
```

### Enhanced Document Display

```bash
# Show only first 5 lines of content and metadata
weave doc list MyCollection --short 5

# Show full content and metadata
weave doc show MyCollection ID --long

# Limit number of documents shown
weave doc list MyCollection --limit 20
```

### Advanced Usage

```bash
# Use custom config and env files
weave config show --config /path/to/config.yaml --env /path/to/.env

# Verbose output
weave health check --verbose

# Quiet output
weave collection list --quiet

# Show full document content
weave document list MyCollection --long

# Limit number of documents shown
weave document list MyCollection --limit 5

# Get help for specific commands
weave collection --help
weave document list --help
```

## Development

### Project Structure

```
weave-cli/
├── src/                    # Source code
│   ├── cmd/               # CLI commands
│   ├── pkg/               # Public packages
│   │   ├── config/       # Configuration management
│   │   ├── weaviate/     # Weaviate client
│   │   └── mock/         # Mock database client
│   └── main.go           # Main entry point
├── tests/                 # Test files
├── bin/                   # Built binaries
├── build.sh              # Build script
├── test.sh               # Test script
├── lint.sh               # Lint script
├── config.yaml           # Configuration file
├── .env                  # Environment variables
└── README.md             # This file
```

### Building

```bash
# Build everything
./build.sh

# Clean build artifacts
./build.sh clean

# Build only (skip tests)
go build -o bin/weave ./src/main.go
```

### Testing

```bash
# Run all tests
./test.sh

# Run only unit tests
./test.sh unit

# Run only integration tests
./test.sh integration

# Run tests with coverage
./test.sh coverage

# Run Go tests directly
go test ./tests/...
```

### Linting

```bash
# Run all linters
./lint.sh

# Run Go linter directly
golangci-lint run ./src/...
```

## Database Support

### Weaviate Cloud

- Full support for Weaviate Cloud instances
- API key authentication
- HTTPS connections
- All CRUD operations

### Weaviate Local

- Support for local Weaviate instances
- No authentication required
- HTTP connections
- All CRUD operations

### Mock Database

- Built-in mock database for testing
- Simulates Weaviate behavior
- No external dependencies
- Perfect for development and testing

## Safety Features

- ⚠️ **Confirmation Prompts** - All destructive operations require confirmation
- 🔒 **API Key Masking** - API keys are never displayed in plain text
- 🛡️ **Error Handling** - Comprehensive error handling and reporting
- 📝 **Logging** - Detailed logging for debugging

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run the test suite: `./test.sh`
6. Run the linter: `./lint.sh`
7. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) CLI framework
- Uses [Weaviate Go Client](https://github.com/weaviate/weaviate-go-client) for database operations
- Inspired by RAGme.io's tools/vdb.sh script