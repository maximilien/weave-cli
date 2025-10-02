# Weave CLI v0.2.7

A command-line tool for managing Weaviate vector databases, written in Go.
This tool provides a fast and easy way to manage content in text and image
collections of configured vector databases, independently of specific
applications.

## 🚀 What's New in v0.2.7

- **🔍 Semantic Search**: New `query` command for semantic search on collections
- **🎯 Smart Fallback**: Automatic fallback from semantic to text search when needed
- **📊 Query Results**: Beautiful formatted results with relevance scores
- **🧪 Comprehensive Testing**: 100% test coverage for query functionality
- **⚡ Performance**: Optimized query execution with configurable result limits
- **🎨 Enhanced Display**: Styled query results with emojis and clear formatting

## ⚠️ Known Limitations

**Weaviate Instance Requirements**: Some Weaviate instances may not support advanced search features:

- `nearText` semantic search requires vector search modules
- `bm25` keyword search requires BM25 module installation  
- `hybrid` search requires hybrid search module
- Fallback to simple text search works but may have accuracy limitations
- **Use `--vector-db-type mock` for full functionality testing**

For more details, see [WEAVIATE_INTEGRATION_STATUS.md](WEAVIATE_INTEGRATION_STATUS.md).

## 🚀 What's New in v0.2.2

- **📄 SPDX License Headers**: Added proper license headers to all 44 Go source files
- **⚖️ Legal Compliance**: Industry-standard SPDX license identification
- **🛠️ License Tool**: Automated script for managing license headers
- **✅ Quality Assurance**: All tests passing (34/34 E2E tests)

## 🚀 What's New in v0.2.1

- **🔧 CI/CD Fixes**: Resolved all linting issues and improved test reliability
- **🧹 Code Cleanup**: Removed unused functions from refactoring process
- **🧪 Enhanced E2E Testing**: More reliable cleanup with 33 comprehensive tests
- **✅ Quality Assurance**: All linting checks passing, improved error handling
- **🎬 Demo Videos**: Interactive asciinema recordings showcasing all features

## 🚀 What's New in v0.2.0

- **🧪 Complete E2E Testing**: 35 comprehensive tests against real Weaviate instances
- **📁 Major Refactoring**: Modular codebase with 98% file size reduction
- **🎨 Enhanced Visibility**: Bold white text for maximum contrast on dark terminals
- **⚙️ Complete Functionality**: All collection and document operations implemented
- **🔧 Smart Configuration**: Auto-detection of Weaviate availability
  with graceful fallback

## 🎬 Demo

Watch Weave CLI in action with our interactive demos:

- **📹 Full Demo** (5 minutes): Complete feature showcase
- **⚡ Quick Demo** (2 minutes): Rapid overview

```bash
# Record your own demo
./tools/asciinema.sh demo

# Upload to asciinema.org for sharing
./tools/asciinema.sh upload
```

See `docs/DEMO.md` for the complete demo script and `videos/README.md` for
recording details.

## Features

- 🌐 **Weaviate Cloud Support** - Connect to Weaviate Cloud instances
- 🏠 **Weaviate Local Support** - Connect to local Weaviate instances  
- 🎭 **Mock Database** - Built-in mock database for testing and development
- 📊 **Collection Management** - List, create, view, and delete collections
- 📄 **Document Management** - List, show, and delete individual documents
- 🔍 **Semantic Search** - Query collections with natural language using
  `weave collection query`
- 🔧 **Configuration Management** - YAML + Environment variable configuration
- 🎨 **Beautiful CLI** - Colored output with emojis and clear formatting
- 📋 **Virtual Document View** - Aggregate chunked documents by original file with
  cross-collection image support
- 🎯 **Smart Truncation** - Intelligent content truncation with `--no-truncate` option
- 🌈 **Color Control** - `--no-color` flag for terminal compatibility
- ⚡ **Fast & Lightweight** - Single binary deployment with optimized image collection
  queries

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

1. **Copy example files**:

   ```bash
   cp config.yaml.example config.yaml
   cp .env.example .env
   ```

2. **Configure your environment**:

   Edit `.env` with your values:

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

   **Alternative: Use command-line flags** (highest priority):

   ```bash
   # Override environment variables directly
   weave --vector-db-type weaviate-cloud \
         --weaviate-url https://your-cluster.weaviate.cloud \
         --weaviate-api-key your-api-key \
         health check
   ```

   **Priority order**: Command flags > `--env` file > `.env` file > Shell environment

3. **Test your connection**:

   ```bash
   ./bin/weave health check
   ```

4. **List your collections**:

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

# Create a new collection (defaults to text schema)
weave collection create MyCollection

# Create collection with specific schema
weave collection create MyTextCollection --text
weave collection create MyImageCollection --image

# Create documents (no schema flags needed - uses collection's schema)
weave docs create MyTextCollection document.txt
weave docs create MyImageCollection image.jpg

# List documents in a collection
weave document list MyCollection

# Virtual document view (aggregate chunks by original file)
weave document list MyCollection --virtual

# Show all data without truncation
weave document list MyCollection --no-truncate

# Disable colored output (useful for scripts)
weave document list MyCollection --no-color

# Show collection schema and metadata structure
weave collection show MyCollection --schema

# Show document details with schema information
weave document show MyCollection DOCUMENT_ID --schema

# Show document by name with schema
weave document show MyCollection --name filename.pdf --schema

# Show collection with expanded metadata analysis
weave collection show MyCollection --expand-metadata

# Show document with expanded metadata analysis
weave document show MyCollection DOCUMENT_ID --expand-metadata

# Query collection with semantic search
weave collection query MyCollection "machine learning algorithms"
weave cols q MyCollection "neural networks" --top_k 3

# Query with custom result limit
weave collection query MyCollection "artificial intelligence" --top_k 10
```

### Destructive Operations

Weave CLI includes safety features for destructive operations:

```bash
# Delete all documents from a collection (double confirmation required)
weave docs delete-all MyCollection

# Delete collection schema completely (double confirmation required)
weave collection delete-schema MyCollection

# Skip confirmations with --force flag
weave docs delete-all MyCollection --force
weave collection delete-schema MyCollection --force
```

**Safety Features:**

- **Double confirmation**: Destructive operations require two confirmations
  - First: Standard y/N prompt
  - Second: Red warning requiring exact "yes" input
- **Force flag**: Use `--force` to skip confirmations in scripts
- **Clear error messages**: Better feedback for empty collections and errors

## Detailed Configuration

Weave CLI uses a combination of YAML configuration files and environment
variables for flexible setup.

### Environment Variables (.env)

Create a `.env` file with your configuration:

```bash
# REQUIRED VALUES
VECTOR_DB_TYPE="weaviate-cloud"  # Options: weaviate-cloud, weaviate-local, mock

# Weaviate Cloud Configuration
WEAVIATE_URL="https://your-cluster.weaviate.cloud"
WEAVIATE_API_KEY="your-weaviate-api-key"
OPENAI_API_KEY="sk-proj-your-openai-api-key"  # Required for embeddings

# OPTIONAL VALUES
WEAVIATE_COLLECTION="MyCollection"
WEAVIATE_COLLECTION_IMAGES="MyImages"
WEAVIATE_COLLECTION_TEST="MyCollection_test"
```

### Command-Line Overrides

You can override any environment variable using command-line flags:

```bash
# Override database type
weave --vector-db-type mock collection list

# Override Weaviate connection details
weave --vector-db-type weaviate-cloud \
      --weaviate-url https://custom-cluster.weaviate.cloud \
      --weaviate-api-key custom-api-key \
      health check

# Use custom config and env files
weave --config /path/to/config.yaml --env /path/to/.env collection list
```

**Priority Order** (highest to lowest):

1. **Command-line flags** (`--vector-db-type`, `--weaviate-url`, `--weaviate-api-key`)
2. **`--env` file** (specified with `--env` flag)
3. **`.env` file** (in current directory)
4. **Shell environment variables**

### Configuration File (config.yaml)

The `config.yaml` file defines your database connections and collections:

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
          description: Main text documents collection
        - name: ${WEAVIATE_COLLECTION_IMAGES:-WeaveImages}
          type: image
          description: Image documents collection
```

### Database Types

#### Weaviate Cloud (weaviate-cloud)

- **Requirements**: `WEAVIATE_URL`, `WEAVIATE_API_KEY`, `OPENAI_API_KEY`
- **Use case**: Production environments with paid Weaviate Cloud service
- **Features**: Full vector search, embeddings, scalability

#### Weaviate Local (weaviate-local)

- **Requirements**: Local Weaviate instance running on `localhost:8080`
- **Use case**: Development and testing with local Weaviate
- **Features**: Full functionality without cloud dependencies

#### Mock Database (mock)

- **Requirements**: None
- **Use case**: Testing, development, demos
- **Features**: Simulated database with no external dependencies

### Quick Setup

```bash
# Copy example files
cp config.yaml.example config.yaml
cp .env.example .env

# Edit with your values
nano .env
nano config.yaml

# Test configuration
./bin/weave health check
```

### Security Notes

- **Never commit** `.env` or `config.yaml` files to version control
- **Use example files** (`config.yaml.example`, `.env.example`) as templates
- **Rotate API keys** regularly for production environments
- **Use different keys** for development and production

## Semantic Search

Query your collections using natural language with the new `query` command:

```bash
# Basic semantic search
weave collection query MyCollection "machine learning algorithms"
weave cols q MyCollection "neural networks"

# Limit results with top_k
weave collection query MyCollection "artificial intelligence" --top_k 5
weave cols q MyCollection "deep learning" --top_k 3

# Using aliases
weave cols q MyCollection "data science"
weave collection q MyCollection "computer vision"
```

**Query Features:**

- **🔍 Semantic Search**: Uses Weaviate's `nearText` for vector-based similarity
  search
- **🎯 Smart Fallback**: Automatically falls back to hybrid search if semantic
  search isn't supported
- **🔤 BM25 Override**: Use `--bm25` flag for keyword-based search instead of
  semantic search
- **📊 Real Scoring**: All search methods provide authentic Weaviate similarity
  scores (0.0 to 1.0)
- **🔍 Metadata Search**: Use `--search-metadata` flag to search in metadata
  fields
- **⚡ Configurable Limits**: Control number of results with `--top_k` flag
  (default: 5)
- **🎨 Beautiful Display**: Formatted results with emojis and clear structure
- **🔄 Cross-Database**: Works with Weaviate Cloud, Local, and Mock databases

**Query Examples:**

```bash
# Find documents about specific topics
weave cols q WeaveDocs "weave-cli installation"
weave cols q WeaveDocs "configuration setup"

# Search with different result limits
weave cols q WeaveDocs "machine learning" --top_k 10
weave cols q WeaveDocs "artificial intelligence" --top_k 3

# Case insensitive search
weave cols q WeaveDocs "MACHINE LEARNING"
weave cols q WeaveDocs "Artificial Intelligence"

# Search with metadata fields
weave cols q WeaveDocs "maximilien.org" --search-metadata

# Use BM25 keyword search
weave cols q WeaveDocs "exact keywords" --bm25

# Combine metadata search with BM25
weave cols q WeaveDocs "search term" --search-metadata --bm25
```

**Query Results Format:**

```bash
$ weave cols q WeaveDocs "weave-cli"

✅ Semantic search results for 'weave-cli' in collection 'WeaveDocs':

1. 🔍 Score: 1.000
   ID: c937af68-727e-4946-8df5-f26919df7645
   Content: # Weave CLI v0.2.6
   
   A command-line tool for managing Weaviate vector databases...
   📋 Metadata: {"filename": "README.md", "type": "text"}

2. 🔍 Score: 0.800
   ID: a0665b61-1558-4ac3-9b26-ecbf755e92b6
   Content: # Installation Guide
   
   Download and setup instructions for Weave CLI...
   📋 Metadata: {"filename": "INSTALL.md", "type": "text"}

📊 Summary: Found 2 results
```

## Collection Management

Create new collections with custom fields and embedding models:

```bash
# Create a basic collection
weave collection create MyCollection

# Create with custom embedding model
weave collection create MyCollection --embedding text-embedding-3-small

# Create with custom fields
weave collection create MyCollection --field title:text,author:text,rating:float

# Create with both custom embedding and fields
weave collection create MyCollection --embedding text-embedding-ada-002 --field title:text,content:text,metadata:object

# Using aliases
weave cols c MyCollection
weave cols create MyCollection --field title:text,author:text
```

## Collection Creation

**DEFAULT**: Collections are created with text schema (RagMeDocs format) unless
`--image` is specified:

```bash
# Create text collections (RagMeDocs schema) - DEFAULT
weave collection create MyTextCollection                    # Default: text schema
weave collection create MyTextCollection --text             # Explicit: text schema
weave collection create MyTextCollection --text --embedding text-embedding-3-small

# Create image collections (RagMeImages schema)  
weave collection create MyImageCollection --image
weave collection create MyImageCollection --image --field title:text,content:text

# Create multiple collections with same schema
weave collection create Col1 Col2 Col3                      # Default: text

# Using aliases
weave cols c MyTextCollection                               # Default: text schema
weave cols c MyImageCollection --image                      # Explicit: image schema
```

**Schema Types:**

- **Default**: Creates collection with text schema (RagMeDocs format) -
  Properties: `url`, `text`, `metadata`
- `--image`: Creates collection with image schema (RagMeImages format) -
  Properties: `url`, `image`, `metadata`, `image_data`

## Document Creation

Document creation works with existing collections (no schema flags required):

```bash
# Create documents in existing collections
weave docs create MyTextCollection document.txt
weave docs create MyTextCollection document.pdf --chunk-size 500
weave docs create MyImageCollection image.jpg

# PDF with both text and images
weave docs create MyTextCollection document.pdf --image-collection MyImageCollection

# Using aliases
weave docs c MyTextCollection document.txt
weave docs c MyImageCollection image.jpg
```

**Supported Field Types:**

- `text` - Text content
- `int` - Integer numbers
- `float` - Floating point numbers
- `bool` - Boolean values
- `date` - Date/time values
- `object` - JSON objects

## Pattern-Based Document Deletion

Delete documents using intuitive shell glob patterns or powerful regex expressions:

```bash
# Shell glob patterns (auto-detected)
weave document delete MyCollection --pattern "tmp*.png"
weave document delete MyCollection --pattern "*.jpg"
weave document delete MyCollection --pattern "file[0-9].txt"
weave document delete MyCollection --pattern "doc?.pdf"

# Regex patterns (auto-detected)
weave document delete MyCollection --pattern "tmp.*\.png"
weave document delete MyCollection --pattern "^prefix.*\.jpg$"
weave document delete MyCollection --pattern ".*\.(png|jpg|gif)$"

# Using aliases
weave docs d MyCollection --pattern "*.png"
weave docs d MyCollection --pattern "temp.*\.pdf"
```

**Pattern Types:**

- **Shell Glob**: `tmp*.png`, `file?.txt`, `doc[0-9].pdf` (familiar shell syntax)
- **Regex**: `tmp.*\.png`, `^file.*\.txt$`, `.*\.(png|jpg)$` (powerful pattern matching)

## Multi-Delete Operations

Delete multiple collections or documents efficiently with enhanced safety:

```bash
# Clear multiple collections (delete all documents)
weave collection delete Collection1 Collection2 Collection3

# Delete multiple documents
weave document delete MyCollection doc1 doc2 doc3

# Delete documents by pattern (shell glob or regex)
weave document delete MyCollection --pattern "tmp*.png"
weave document delete MyCollection --pattern "tmp.*\.png"

# Using aliases
weave cols d Col1 Col2 Col3
weave docs d MyCollection doc1 doc2 doc3
weave docs d MyCollection --pattern "*.jpg"

# Skip confirmation with --force flag
weave cols d Col1 Col2 Col3 --force
weave docs d MyCollection doc1 doc2 doc3 --force
```

**Safety Features:**

- Itemized lists showing exactly what will be deleted
- Progress tracking ("Deleting item X/Y")
- Error resilience (continues on individual failures)
- Summary reports with success/failure counts
- Double confirmation for delete-all operations
- Clear messaging: collection deletion removes documents but keeps schema

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
  - `weave collection create COLLECTION_NAME [COLLECTION_NAME...]` - Create collection(s)
  - `weave collection query COLLECTION_NAME "query text"` - Semantic search on collection
  - `weave collection list --virtual` - Show collections with virtual structure
  - `weave collection delete COLLECTION_NAME [COLLECTION_NAME...]` - Clear collections
  - `weave collection delete-schema COLLECTION_NAME [COLLECTION_NAME...]` -
    Delete collection schema(s) completely
  - `weave collection delete-all` - Clear all collections (double confirmation)

- **document** - Document management
  - `weave document list COLLECTION_NAME` - List documents in collection
  - `weave document list COLLECTION_NAME --virtual` - Virtual view
  - `weave document show COLLECTION_NAME DOCUMENT_ID` - Show document
  - `weave document delete COLLECTION_NAME [DOCUMENT_ID] [DOCUMENT_ID...]` -
    Delete docs
  - `weave document delete COLLECTION_NAME --pattern "PATTERN"` - Delete by pattern
  - `weave document delete-all COLLECTION_NAME` - Delete all docs (double confirmation)

### Command Aliases

For convenience, shorter aliases are available:

```bash
# Collection commands
weave col list          # Same as: weave collection list
weave cols list         # Same as: weave collection list
weave col create MyCol  # Same as: weave collection create MyCol
weave cols create MyCol # Same as: weave collection create MyCol
weave cols c MyCol      # Same as: weave collection create MyCol
weave cols c Col1 Col2 Col3  # Create multiple collections at once
weave cols q MyCol "query" # Same as: weave collection query MyCol "query"
weave col query MyCol "query" # Same as: weave collection query MyCol "query"
weave cols d Col1 Col2  # Same as: weave collection delete Col1 Col2
weave cols ds MyCol     # Same as: weave collection delete-schema MyCol
weave cols ds Col1 Col2 Col3  # Delete multiple collection schemas at once

# Document commands  
weave doc list MyCol    # Same as: weave document list MyCol
weave docs list MyCol   # Same as: weave document list MyCol
weave docs d MyCol doc1 doc2  # Same as: weave document delete MyCol doc1 doc2
```

## New Features

### Multiple Collection Creation

Create multiple collections at once with a single command:

```bash
# Create multiple collections
weave collection create WeaveDocs WeaveImages WeaveTest
weave cols c Col1 Col2 Col3 Col4

# With custom embedding model
weave collection create MyCol1 MyCol2 --embedding text-embedding-3-large

# With custom fields
weave collection create DataCol1 DataCol2 --field title:text,author:text,tags:text
```

### Collection Schema Management

Completely remove collection schemas (useful for schema updates):

```bash
# Delete collection schema completely
weave collection delete-schema WeaveDocs --force
weave cols ds WeaveImages --force

# Delete multiple collection schemas at once
weave collection delete-schema WeaveDocs WeaveImages WeaveTest --force
weave cols ds Col1 Col2 Col3 --force

# Then recreate with new schema
weave collection create WeaveDocs
```

**Note**: `delete-schema` removes the collection entirely, while `delete` only
clears documents.

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

✅ Found 3 virtual documents in collection 'MyCollection' (aggregated from 15 total
documents):

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

- **Cross-collection aggregation**: Automatically finds and includes images from
  corresponding image collections
- **Smart grouping**: Images are grouped with their source documents (PDFs,
  presentations, etc.)
- **Complete view**: Shows both text chunks and extracted images in one unified
  view
- **Collection mapping**: Maps document collections to image collections (e.g.,
  `MyDocs` → `MyImages`)
- **Performance optimized**: Excludes large base64 image data from queries for
  fast listing and counting

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

🎯 **[Presentation](docs/PRESENTATION.md)** - Marp presentation with overview,
features, and usage examples.

## Recent Improvements

- ✅ **Linting fixes** - All YAML, Markdown, and Go linting issues resolved
- ✅ **Security tools** - govulncheck and gosec installed and configured
- ✅ **CI/CD pipeline** - GitHub Actions for automated testing and building
- ✅ **Documentation** - Added Marp presentation and updated guides
- ✅ **Code quality** - Comprehensive test coverage and quality checks

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
│   ├── cmd/               # CLI commands (organized by functionality)
│   │   ├── document.go   # Main document command
│   │   ├── collection.go  # Main collection command
│   │   ├── list.go       # Document list command
│   │   ├── show.go       # Document show command
│   │   ├── count.go      # Document count command
│   │   ├── create.go     # Document create command
│   │   ├── delete.go     # Document delete command
│   │   ├── delete_all.go # Document delete-all command
│   │   ├── delete_schema.go # Collection delete-schema command
│   │   └── [other command files...]
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

**Code Organization:**

- Commands are organized into logical files by functionality
- Each subcommand has its own file (e.g., `list.go`, `create.go`)
- Main command files contain only the command definition
- Helper functions remain in the original files for shared access

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
- Uses [Weaviate Go Client](https://github.com/weaviate/weaviate-go-client) for database
  operations
- Inspired by RAGme.io's tools/vdb.sh script
