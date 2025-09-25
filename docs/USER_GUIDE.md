# Weave CLI User Guide

A comprehensive guide to using the Weave CLI tool for managing Weaviate vector databases.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Configuration](#configuration)
3. [Basic Commands](#basic-commands)
4. [Virtual Document View](#virtual-document-view)
5. [Global Flags](#global-flags)
6. [Advanced Usage](#advanced-usage)
7. [Troubleshooting](#troubleshooting)
8. [Examples](#examples)

## Getting Started

### Installation

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

### Quick Start

1. **Configure your database**:

   ```bash
   # Copy the example configuration
   cp config.yaml.example config.yaml
   
   # Edit with your Weaviate details
   nano config.yaml
   ```

2. **Set environment variables**:

   ```bash
   export WEAVIATE_URL="your-weaviate-url.weaviate.cloud"
   export WEAVIATE_API_KEY="your-api-key"
   export VECTOR_DB_TYPE="weaviate-cloud"
   ```

3. **Test your connection**:

   ```bash
   ./bin/weave health check
   ```

4. **List your collections**:

   ```bash
   ./bin/weave collection list
   ```

## Configuration

### Configuration Files

Weave CLI uses two configuration files:

#### config.yaml

```yaml
# Databases Configuration
databases:
  # Default vector database to use
  default: "${VECTOR_DB_TYPE:-weaviate-cloud}"

  # Vector Databases settings
  vector_databases:
    # Weaviate Cloud configuration
    - name: "weaviate-cloud"
      type: "weaviate-cloud"
      url: "${WEAVIATE_URL}"
      api_key: "${WEAVIATE_API_KEY}"
      collections:
        - name: "${WEAVIATE_COLLECTION:-WeaveDocs}"
          type: "text"
        - name: "${WEAVIATE_COLLECTION_IMAGES:-WeaveImages}"
          type: "image"

    # Weaviate Local configuration
    - name: "weaviate-local"
      type: "weaviate-local"
      url: "http://localhost:8080"
      collections:
        - name: "${WEAVIATE_COLLECTION:-WeaveDocs}"
          type: "text"
        - name: "${WEAVIATE_COLLECTION_IMAGES:-WeaveImages}"
          type: "image"

    # Mock Vector Database configuration (for development/testing)
    - name: "mock"
      type: "mock"
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

#### .env

```bash
# Vector Database Configuration  
VECTOR_DB_TYPE="weaviate-cloud"
WEAVIATE_COLLECTION="WeaveDocs"
WEAVIATE_COLLECTION_TEST="WeaveDocs_test"

WEAVIATE_URL="your-weaviate-url.weaviate.cloud"
WEAVIATE_API_KEY="your-api-key"
```

### Database Types

#### Weaviate Cloud

- **Type**: `weaviate-cloud`
- **Authentication**: API key required
- **Connection**: HTTPS
- **Use Case**: Production environments

#### Weaviate Local

- **Type**: `weaviate-local`
- **Authentication**: None required
- **Connection**: HTTP
- **Use Case**: Development and testing

#### Mock Database

- **Type**: `mock`
- **Authentication**: None required
- **Connection**: In-memory
- **Use Case**: Testing, development, demonstrations

## Basic Commands

### Command Structure

Weave follows a consistent command pattern:
`weave noun verb [arguments] [flags]`

### Configuration Management

```bash
# Show current configuration
weave config show

# Show configuration with custom files
weave config show --config /path/to/config.yaml --env /path/to/.env
```

### Health Management

```bash
# Check database health
weave health check

# Check health with verbose output
weave health check --verbose
```

### Collection Management

```bash
# List all collections
weave collection list

# List collections with virtual structure summary
weave collection list --virtual

# List collections in specific database
weave collection list mock

# Delete a specific collection
weave collection delete MyCollection

# Delete all collections (⚠️ DESTRUCTIVE)
weave collection delete-all
```

### Document Management

```bash
# List documents in a collection
weave document list MyCollection

# List documents with virtual view
weave document list MyCollection --virtual

# Show a specific document
weave document show MyCollection document-id

# Delete a specific document
weave document delete MyCollection document-id

# Delete all documents in a collection (⚠️ DESTRUCTIVE)
weave document delete-all MyCollection
```

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

## Document Display

Both regular and virtual document views feature consistent visual styling for
better readability and user experience.

### Regular Document View

The standard document listing shows individual documents with enhanced styling:

```bash
# Basic document listing with improved styling
weave document list MyCollection

# Example output:
✅ Found 6 documents in collection 'MyCollection':

1. 📄 ID: doc1-chunk1
   Content: This is the first chunk of a document about machine learning...
   📋 Metadata: 
     metadata: {"original_filename": "ml_guide.pdf", "is_chunked": true...}
     author: Test Author
```

### Virtual Document View

The `--virtual` flag provides an intelligent view by aggregating chunked
content back into original documents.

#### Features

- **📄 Document Aggregation** - Groups chunks by original filename
- **📊 Smart Statistics** - Shows chunk counts and document structure
- **🎨 Visual Hierarchy** - Clear distinction between important and metadata information
- **🖼️ Image Support** - Handles image collections with page-based grouping
- **📋 Metadata Display** - Shows relevant metadata with proper formatting

### Example Usage

```bash
# Basic virtual view
weave document list MyCollection --virtual

# Virtual view with no truncation
weave document list MyCollection --virtual --no-truncate

# Virtual view without colors (for scripts)
weave document list MyCollection --virtual --no-color

# Collection virtual summary
weave collection list --virtual
```

### Example Output

```bash
$ weave document list MyCollection --virtual

✅ Found 3 virtual documents in collection 'MyCollection' (aggregated from
6 total documents):

1. 📄 Document: research_paper.pdf
   📝 Chunks: 3/3
   📋 Metadata: 
     original_filename: research_paper.pdf
   📝 Chunk Details: 
     1. ID: chunk-1
        Content: Introduction to machine learning concepts...
     2. ID: chunk-2  
        Content: Deep learning architectures and applications...
     3. ID: chunk-3
        Content: Conclusion and future research directions...

2. 📄 Document: presentation.pptx
   🖼️ Images: 2
   📋 Metadata: 
     original_filename: presentation.pptx
   🗂️ Stack Details: 
     1. ID: img-1
        Content: Slide 1: Overview diagram
     2. ID: img-2
        Content: Slide 2: Architecture diagram
```

### Visual Styling

- **Top-level keys** (ID, Chunks, Images, Content) are prominent
- **Metadata keys** are dimmed for better hierarchy
- **Important values** (IDs, filenames, numbers) are highlighted
- **Emojis** provide visual structure (disabled with `--no-color`)

## Global Flags

### --no-color

Disables colored output for better compatibility with scripts and logs.

```bash
# Disable colors
weave document list MyCollection --no-color

# Useful for logging
weave collection list --no-color >> output.log
```

### --no-truncate

Shows all data without truncation.

```bash
# Show full content
weave document list MyCollection --no-truncate

# Combine with virtual view
weave document list MyCollection --virtual --no-truncate
```

### --verbose

Provides detailed output for debugging.

```bash
# Verbose health check
weave health check --verbose

# Verbose configuration display
weave config show --verbose
```

### --quiet

Minimal output for scripts.

```bash
# Quiet collection listing
weave collection list --quiet

# Quiet document listing
weave document list MyCollection --quiet
```

## Advanced Usage

### Custom Configuration Files

```bash
# Use custom config file
weave config show --config /path/to/custom-config.yaml

# Use custom env file
weave config show --env /path/to/custom.env

# Combine both
weave config show --config /path/to/config.yaml --env /path/to/.env
```

### Document Display Options

```bash
# Show only first 5 lines of content
weave document list MyCollection --short 5

# Show full content
weave document list MyCollection --long

# Limit number of documents
weave document list MyCollection --limit 20

# Combine options
weave document list MyCollection --virtual --long --limit 10
```

### Database Selection

```bash
# Use mock database
VECTOR_DB_TYPE=mock weave collection list

# Use local Weaviate
VECTOR_DB_TYPE=weaviate-local weave document list MyCollection

# Use Weaviate Cloud (default)
VECTOR_DB_TYPE=weaviate-cloud weave health check
```

## Troubleshooting

### Common Issues

#### Connection Errors

```bash
# Check your configuration
weave config show

# Test connectivity
weave health check --verbose

# Verify environment variables
echo $WEAVIATE_URL
echo $WEAVIATE_API_KEY
```

#### Permission Errors

```bash
# Make sure the binary is executable
chmod +x bin/weave

# Check file permissions
ls -la bin/weave
```

#### Configuration Issues

```bash
# Validate configuration syntax
weave config show

# Check for missing environment variables
weave config show --verbose
```

#### Virtual Document Chunk Count Issues

**Issue**: Virtual document view (`-w` flag) shows incorrect chunk counts when
using limit parameter.

**Symptoms**:

- Commands like `weave docs l MyCollection -w -S -l 10` show wrong chunk
  counts
- Example: vectras.pdf shows "1 chunks" instead of "7 chunks"

**Solution**: This issue was fixed in v0.0.6. The virtual document view now
correctly retrieves all chunks for proper aggregation, regardless of the
limit parameter.

**Verification**:

```bash
# Test with your collection
weave docs l MyCollection -w -S -l 10

# Should show accurate chunk counts in summary
# Example output:
# 📋 Summary: 
#    1. document1.pdf - 10 chunks
#    2. document2.pdf - 25 chunks
```

### Debug Mode

```bash
# Enable verbose output for debugging
weave health check --verbose
weave collection list --verbose
weave document list MyCollection --verbose
```

### Getting Help

```bash
# General help
weave --help

# Command-specific help
weave collection --help
weave document list --help

# Version information
weave --version
```

## Examples

### Basic Workflow

```bash
# 1. Check configuration
weave config show

# 2. Test connection
weave health check

# 3. List collections
weave collection list

# 4. List documents in a collection
weave document list MyCollection

# 5. View documents in virtual format
weave document list MyCollection --virtual
```

### Development Workflow

```bash
# Use mock database for development
VECTOR_DB_TYPE=mock weave collection list

# Test with mock data
VECTOR_DB_TYPE=mock weave document list WeaveDocs --virtual

# Switch to real database
VECTOR_DB_TYPE=weaviate-cloud weave health check
```

### Script Integration

```bash
#!/bin/bash
# Example script using Weave CLI

# Check if Weave CLI is available
if ! command -v weave &> /dev/null; then
    echo "Weave CLI not found"
    exit 1
fi

# Get collection list (no colors for script output)
collections=$(weave collection list --no-color --quiet)

# Process each collection
for collection in $collections; do
    echo "Processing collection: $collection"
    
    # Get document count
    count=$(weave document list "$collection" --no-color --quiet | wc -l)
    echo "Documents in $collection: $count"
done
```

### Monitoring and Logging

```bash
# Log collection status
weave collection list --no-color >> /var/log/weave-collections.log

# Monitor document counts
weave document list MyCollection --no-color --quiet >> /var/log/weave-documents.log

# Health check monitoring
weave health check --no-color >> /var/log/weave-health.log
```

---

*For more information, see the [README.md](../README.md) or run `weave --help`.*
