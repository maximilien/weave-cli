# Weave CLI User Guide

A comprehensive guide to using the Weave CLI tool for managing Weaviate vector databases.

> **📖 Quick Reference**: For a quick overview, see the [README.md](../README.md)

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
- Tesseract OCR (optional, for OCR text extraction from images): `brew install tesseract`
- Poppler (optional, for PDF text extraction): `brew install poppler`

### Quick Start

**Method 1: Environment Variables Only (Recommended)**
```bash
# Set your environment variables
export VECTOR_DB_TYPE="weaviate-cloud"
export WEAVIATE_URL="https://your-cluster.weaviate.cloud"
export WEAVIATE_API_KEY="your-weaviate-api-key"
export OPENAI_API_KEY="sk-proj-your-openai-api-key"

# Test the connection
./bin/weave health check
```

**Method 2: Configuration Files**
```bash
# Copy the example configuration
cp config.yaml.example config.yaml
cp .env.example .env

# Edit with your Weaviate details
nano config.yaml
nano .env
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

Weave CLI supports flexible configuration through environment variables or YAML files.
**Configuration files are optional** - you can use environment variables alone!

### Configuration Methods

**Method 1: Environment Variables Only (Fastest)**
```bash
export VECTOR_DB_TYPE="weaviate-cloud"
export WEAVIATE_URL="https://your-cluster.weaviate.cloud"
export WEAVIATE_API_KEY="your-weaviate-api-key"
export OPENAI_API_KEY="sk-proj-your-openai-api-key"
```

**Method 2: Configuration Files**

Weave CLI can use two configuration files:

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

  # Collection Schemas - define reusable schema templates
  schemas:
    - name: RagMeDocs
      schema:
        class: RagMeDocs
        vectorizer: text2vec-weaviate
        properties:
          - name: url
            datatype: [text]
            description: the source URL of the webpage
          - name: text
            datatype: [text]
            description: the content of the webpage
          - name: metadata
            datatype: [text]
            description: additional metadata in JSON format
            json_schema:
              filename: string
              type: string
              date_added: string
      metadata:
        id: string
        url: string
        text: string
        metadata:
          type: json
          json_schema:
            filename: string
            type: string
            date_added: string

    - name: RagMeImages
      schema:
        class: RagMeImages
        vectorizer: text2vec-weaviate
        properties:
          - name: url
            datatype: [text]
            description: the source URL or filename of the image
          - name: image
            datatype: [text]
            description: the image reference (truncated base64 or URL)
          - name: metadata
            datatype: [text]
            description: additional metadata in JSON format
            json_schema:
              filename: string
              format: string
              source: string
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

### Schema Configuration

Weave CLI supports defining reusable schemas in two ways:
1. **Inline** in `config.yaml` under `databases.schemas`
2. **Directory-based** with individual YAML files in a schemas directory

**Note:** Inline schemas take precedence over directory schemas with the same name.

#### Schemas Directory

Configure a schemas directory to store schema files separately:

```yaml
# config.yaml
schemas_dir: ./schemas
```

Each schema file should contain a single schema definition:

```yaml
# schemas/testschema.yaml
---
name: TestSchema
schema:
  class: TestSchema
  vectorizer: text2vec-weaviate
  properties:
    - name: title
      datatype: [text]
      description: the title
metadata:
  id: string
  title: string
```

#### Defining Inline Schemas

Schemas can also be defined inline in the `databases.schemas` section of `config.yaml`:

```yaml
# config.yaml
schemas_dir: ./schemas  # Optional: load schemas from directory

databases:
  schemas:
    # Inline schemas (these take precedence over directory schemas)
    - name: RagMeDocs
      schema:
        class: RagMeDocs
        vectorizer: text2vec-weaviate
        properties:
          - name: url
            datatype: [text]
            description: the source URL of the webpage
          - name: text
            datatype: [text]
            description: the content of the webpage
          - name: metadata
            datatype: [text]
            description: additional metadata in JSON format
            json_schema:
              filename: string
              type: string
              date_added: string
```

**Schema Precedence:** If a schema named "RagMeDocs" exists both in `./schemas/ragmedocs.yaml` and inline in `config.yaml`, the inline version will be used.

#### Viewing Configured Schemas

```bash
# List all configured schemas
weave config list-schemas

# Show details of a specific schema
weave config show-schema RagMeDocs

# Output shows:
# - Schema class name
# - Vectorizer type
# - All properties with types and descriptions
# - JSON schema structures for complex fields
```

#### Using Schemas

```bash
# Create a collection using a named schema
weave collection create MyDocs --schema RagMeDocs

# Create an image collection using a named schema
weave collection create MyImages --schema RagMeImages
```

#### Schema Export and Import

You can export existing collection schemas and use them as templates:

```bash
# Export a collection's schema with JSON field inference
weave collection show MyCollection --schema --compact --yaml-file schema.yaml

# Create new collection from exported schema
weave collection create NewCollection --schema-yaml-file schema.yaml
```

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

# List all configured schemas
weave config list-schemas

# Show details of a specific schema
weave config show-schema RagMeDocs
weave config show-schema RagMeImages

# Output schema in YAML or JSON format
weave config show-schema RagMeDocs --yaml
weave config show-schema RagMeDocs --json

# Using aliases
weave config ls-schemas
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

# Create a new collection
weave collection create MyCollection

# Create collection with custom embedding model
weave collection create MyCollection --embedding text-embedding-3-small

# Create collection with custom fields
weave collection create MyCollection --field title:text,author:text,rating:float,published:bool

# Create collection with both custom embedding and fields
weave collection create MyCollection --embedding text-embedding-ada-002 --field title:text,content:text,metadata:object

# Clear a specific collection (delete all documents)
weave collection delete MyCollection

# Clear multiple collections (delete all documents)
weave collection delete Collection1 Collection2 Collection3

# Clear all collections (⚠️ DESTRUCTIVE)
weave collection delete-all
```

### Document Management

```bash
# Create documents with required schema flags
weave docs create MyTextCollection document.txt --text
weave docs create MyImageCollection image.jpg --image
weave docs create MyTextCollection document.pdf --text --chunk-size 500

# List documents in a collection
weave document list MyCollection

# List documents with virtual view
weave document list MyCollection --virtual

# Count documents in a single collection
weave document count MyCollection

# Count documents in multiple collections
weave document count RagMeDocs RagMeImages

# Show a specific document
weave document show MyCollection document-id

# Delete a specific document
weave document delete MyCollection document-id

# Delete multiple documents
weave document delete MyCollection doc1 doc2 doc3

# Delete documents by pattern (shell glob or regex)
weave document delete MyCollection --pattern "tmp*.png"
weave document delete MyCollection --pattern "*.jpg"
weave document delete MyCollection --pattern "file[0-9].txt"
weave document delete MyCollection --pattern "tmp.*\.png"

# Delete all documents in a collection (⚠️ DESTRUCTIVE)
weave document delete-all MyCollection
```

### Command Aliases

For convenience, shorter aliases are available:

```bash
# Collection commands
weave col list          # Same as: weave collection list
weave cols list         # Same as: weave collection list
weave col create MyCol  # Same as: weave collection create MyCol
weave cols create MyCol # Same as: weave collection create MyCol
weave cols c MyCol      # Same as: weave collection create MyCol
weave col delete MyCol  # Same as: weave collection delete MyCol
weave cols d Col1 Col2  # Same as: weave collection delete Col1 Col2

# Document commands  
weave doc list MyCol    # Same as: weave document list MyCol
weave docs list MyCol   # Same as: weave document list MyCol
weave doc C MyCol       # Same as: weave document count MyCol
weave docs C MyCol      # Same as: weave document count MyCol
weave docs C RagMeDocs RagMeImages  # Count multiple collections
weave doc show MyCol ID # Same as: weave document show MyCol ID
weave docs d MyCol doc1 doc2  # Same as: weave document delete MyCol doc1 doc2
```

## Collection Create Command

Collections can be created using named schemas from `config.yaml`, schema files, or with default/custom settings.

### Creating Collections with Named Schemas

The easiest way to create collections is to use a named schema defined in your `config.yaml`:

```bash
# Create collection using RagMeDocs schema from config.yaml
weave collection create MyDocsCol --schema RagMeDocs

# Create collection using RagMeImages schema from config.yaml
weave collection create MyImagesCol --schema RagMeImages

# Using aliases
weave cols c MyDocsCol --schema RagMeDocs
weave cols c MyImagesCol --schema RagMeImages
```

### Creating Collections from Schema Files

You can also create collections from exported schema YAML files:

```bash
# First, export an existing collection's schema
weave collection show ExistingCollection --yaml-file schema.yaml --compact

# Then create a new collection from that schema
weave collection create NewCollection --schema-yaml-file schema.yaml
```

### Default Collection Creation

**DEFAULT**: Collections are created with text schema (RagMeDocs format) unless otherwise specified.

```bash
# Create text collections (RagMeDocs schema) - DEFAULT
weave collection create MyTextCollection                    # Default: text schema
weave collection create MyTextCollection --text             # Explicit: text schema
weave collection create MyTextCollection --text --embedding text-embedding-3-small

# Create image collections (RagMeImages schema)
weave collection create MyImageCollection --image
weave collection create MyImageCollection --image --field title:text,content:text
```

### Schema Types

#### RagMeDocs Schema (Text Documents)
- **Properties**: `url`, `text`, `metadata`
- **Use case**: Text documents, PDF text chunks, web pages
- **Vectorization**: Enabled with text2vec-weaviate
- **Usage**: `--schema RagMeDocs` or default behavior

#### RagMeImages Schema (Image Documents)
- **Properties**: `url`, `image`, `metadata`, `image_data`
- **Use case**: Image documents, PDF extracted images
- **Vectorization**: Enabled with text2vec-weaviate
- **Usage**: `--schema RagMeImages` or `--image` flag
- **Enhanced Metadata**: Automatically extracts EXIF data, OCR text, timestamped storage paths, and processing metrics from images and PDF-extracted images

### Defining Custom Schemas in config.yaml

Add custom schemas to your `config.yaml` file under the `databases.schemas` section:

```yaml
databases:
  schemas:
    - name: MyCustomSchema
      schema:
        class: MyCustomSchema
        vectorizer: text2vec-weaviate
        properties:
          - name: title
            datatype:
              - text
            description: the document title
          - name: content
            datatype:
              - text
            description: the document content
```

Then use it:

```bash
weave collection create MyCollection --schema MyCustomSchema
```

### Schema Export with JSON Field Inference

When exporting collection schemas, the CLI automatically detects and infers
the structure of JSON-encoded string fields in metadata. This provides
accurate, detailed schema specifications for collections with structured data.

```bash
# Export schema to YAML with JSON structure inference
weave cols show RagMeDocs --schema --yaml --vector-db-type weaviate-cloud

# Export schema to JSON format
weave cols show RagMeDocs --schema --json --vector-db-type weaviate-cloud

# Export to file in compact mode (no samples/occurrences)
weave cols show RagMeDocs --schema --yaml-file schema.yaml --compact

# Export to JSON file
weave cols show RagMeDocs --schema --json-file schema.json
```

#### JSON Field Detection

The CLI analyzes metadata fields across multiple documents and automatically:
- Detects JSON-encoded string fields
- Infers field types (string, integer, number, boolean, array, object)
- Merges schemas across documents to capture all possible fields
- Adds `json_schema` property with field specifications

#### Example Output

For a collection with JSON metadata like `{"type": "pdf", "filename": "doc.pdf"}`:

```yaml
metadata:
  metadata:
    type: json
    json_schema:
      type: string
      filename: string
      date_added: string
      chunk_index: integer
      is_chunked: boolean
      total_chunks: integer
```

This makes it easy to:
- Understand the structure of your metadata
- Create accurate schema definitions for new collections
- Document your collection schemas comprehensively

## Document Create Command

Document creation works with existing collections (no schema flags required):

```bash
# Create documents in existing collections
weave docs create MyTextCollection document.txt
weave docs create MyTextCollection document.pdf --chunk-size 500
weave docs create MyImageCollection image.jpg

# PDF with both text and images
weave docs create MyTextCollection document.pdf --image-collection MyImageCollection

# PDF with image extraction (automatically filters images < 5KB)
weave docs create MyTextCollection document.pdf --image-collection MyImageCollection

# PDF with custom image size filter (e.g., 10KB minimum)
weave docs create MyTextCollection document.pdf --image-collection MyImageCollection --min-image-size 10240

# Using aliases
weave docs c MyTextCollection document.txt
weave docs c MyImageCollection image.jpg
```

## Collection Create Command (Legacy)

The `weave collection create` command (alias: `weave cols c`) allows you to
create new collections with custom fields and embedding models.

### Basic Collection Creation

```bash
# Create a basic collection with default fields
weave collection create MyCollection

# Using alias
weave cols c MyCollection

# Example output:
# ✅ Successfully created collection: MyCollection
# ℹ️  Embedding model: text-embedding-ada-002
```

### Custom Embedding Models

```bash
# Create collection with specific embedding model
weave collection create MyCollection --embedding text-embedding-3-small

# Using alias
weave cols c MyCollection --embedding text-embedding-ada-002

# Example output:
# ✅ Successfully created collection: MyCollection
# ℹ️  Embedding model: text-embedding-3-small
```

### Custom Fields

```bash
# Create collection with custom fields
weave collection create MyCollection --field title:text,author:text,rating:float

# Using alias
weave cols c MyCollection --field title:text,author:text,rating:float

# Example output:
# ✅ Successfully created collection: MyCollection
# 
# Custom fields:
#   - title: text
#   - author: text
#   - rating: float
# 
# ℹ️  Embedding model: text-embedding-ada-002
```

### Multiple Collection Creation

Create multiple collections at once:

```bash
# Create multiple collections with default settings
weave collection create WeaveDocs WeaveImages WeaveTest

# Using alias
weave cols c Col1 Col2 Col3 Col4

# Create multiple collections with custom embedding
weave collection create MyCol1 MyCol2 MyCol3 --embedding text-embedding-3-large

# Create multiple collections with custom fields
weave collection create DataCol1 DataCol2 --field title:text,author:text,tags:text

# Example output for multiple collections:
# 🔧 Create Collection(s)
# 
# Creating 3 collections in weaviate-cloud database...
# 
# ℹ️  Collections to create:
#   1. WeaveDocs
#   2. WeaveImages
#   3. WeaveTest
# 
# Creating collection 1/3: WeaveDocs
# ✅ Successfully created collection: WeaveDocs
# 
# Creating collection 2/3: WeaveImages
# ✅ Successfully created collection: WeaveImages
# 
# Creating collection 3/3: WeaveTest
# ✅ Successfully created collection: WeaveTest
# 
# ✅ All 3 collections created successfully!
```

### Combined Options

```bash
# Create collection with both custom embedding and fields
weave collection create MyCollection --embedding text-embedding-3-small --field title:text,content:text,metadata:object

# Using alias
weave cols c MyCollection --embedding text-embedding-ada-002 --field title:text,content:text,metadata:object
```

### Supported Field Types

- `text` - Text content
- `int` - Integer numbers
- `float` - Floating point numbers
- `bool` - Boolean values
- `date` - Date/time values
- `object` - JSON objects

### Error Handling

```bash
# Collection already exists
weave cols c ExistingCollection
# ❌ Failed to create collection 'ExistingCollection': collection 'ExistingCollection' already exists

# Invalid field type
weave cols c MyCollection --field title:invalid
# ❌ Invalid field definition: invalid field type 'invalid', supported types: text, int, float, bool, date, object

# Invalid field format
weave cols c MyCollection --field title
# ❌ Invalid field definition: field definition must be in format 'name:type', got 'title'
```

## Multi-Delete Commands

The `weave collection delete` and `weave document delete` commands now support
clearing multiple items at once with enhanced safety features.

**Note:** Collection deletion removes all documents from the collection but keeps
the collection schema intact. The collection will appear empty but still exist.

### Collection Schema Deletion

For complete removal of collections (including schema), use the `delete-schema` command:

```bash
# Delete collection schema completely
weave collection delete-schema WeaveDocs --force

# Using alias
weave cols ds WeaveImages --force

# Delete multiple collection schemas at once
weave collection delete-schema WeaveDocs WeaveImages WeaveTest --force
weave cols ds Col1 Col2 Col3 --force

# Example output for multiple schemas:
# 🔧 Delete Collection Schema(s)
# 
# ⚠️  WARNING: This will permanently delete the schemas for 3 collections!
# 
# ℹ️  Collections to delete:
#   1. WeaveDocs
#   2. WeaveImages
#   3. WeaveTest
# 
# Deleting schemas for 3 collections in weaviate-cloud database...
# 
# Deleting schema 1/3: WeaveDocs
# ✅ Successfully deleted schema for collection: WeaveDocs
# 
# Deleting schema 2/3: WeaveImages
# ✅ Successfully deleted schema for collection: WeaveImages
# 
# Deleting schema 3/3: WeaveTest
# ✅ Successfully deleted schema for collection: WeaveTest
# 
# ✅ All 3 collection schemas deleted successfully!
```

**Important:** Schema deletion completely removes the collection from the database.
Use this when you need to recreate a collection with a different schema.

### Collection Multi-Delete

```bash
# Clear multiple collections with confirmation
weave collection delete Collection1 Collection2 Collection3

# Using alias
weave cols d Collection1 Collection2 Collection3

# Skip confirmation with --force flag
weave cols d Collection1 Collection2 Collection3 --force

# Example output:
# 🔧 Delete Collection(s)
# 
# ⚠️  WARNING: This will permanently delete all documents from 3 collections!
# 
# ℹ️  Collections to delete:
#   1. Collection1
#   2. Collection2
#   3. Collection3
# 
# Are you sure you want to clear 3 collections? (y/N): y
# 
# Deleting 3 collections in weaviate-cloud database...
# 
# Deleting collection 1/3: Collection1
# ✅ Successfully deleted all documents from collection: Collection1
# 
# Deleting collection 2/3: Collection2
# ✅ Successfully deleted all documents from collection: Collection2
# 
# Deleting collection 3/3: Collection3
# ✅ Successfully deleted all documents from collection: Collection3
# 
# ✅ All 3 collections cleared successfully!
```

### Document Multi-Delete

```bash
# Delete multiple documents with confirmation
weave document delete MyCollection doc1 doc2 doc3

# Using alias
weave docs d MyCollection doc1 doc2 doc3

# Skip confirmation with --force flag
weave docs d MyCollection doc1 doc2 doc3 --force

# Example output:
# 🔧 Delete Document(s)
# 
# ⚠️  WARNING: This will permanently delete 3 documents from collection 'MyCollection'!
# 
# ℹ️  Documents to delete:
#   1. doc1
#   2. doc2
#   3. doc3
# 
# Are you sure you want to delete 3 documents? (y/N): y
# 
# Deleting 3 documents from weaviate-cloud database...
# 
# Deleting document 1/3: doc1
# ✅ Successfully deleted document: doc1
# 
# Deleting document 2/3: doc2
# ✅ Successfully deleted document: doc2
# 
# Deleting document 3/3: doc3
# ✅ Successfully deleted document: doc3
# 
# ✅ All 3 documents deleted successfully!
```

### Enhanced Safety Features

- **Itemized Lists**: Shows exactly what will be deleted before confirmation
- **Progress Tracking**: Displays "Deleting item X/Y" progress
- **Error Resilience**: Continues processing even if some deletions fail
- **Summary Reports**: Shows success/failure counts for multi-item operations
- **--force Flag**: Skip confirmation prompts for automated scripts

### Double Confirmation for Delete-All Commands

The most destructive operations (`weave cols da` and `weave docs da`) require
**double confirmation** for maximum safety:

```bash
# Collection delete-all with double confirmation
weave cols da

# Example output:
# 🔧 Delete All Collections
# 
# ⚠️  WARNING: This will permanently delete ALL collections and their data!
# 
# Are you sure you want to delete all collections? (y/N): y
# 
# 🚨 FINAL WARNING: This operation CANNOT be undone!
# All collections and their data will be permanently deleted.
# 
# Type 'yes' to confirm deletion: yes
# 
# Deleting all collections in weaviate-cloud database...
# ✅ All collections deleted successfully!
```

**Safety Features:**
- **First Confirmation**: Standard y/N prompt
- **Second Confirmation**: Red warning requiring exact "yes" input
- **Visual Warning**: 🚨 emoji and red text for maximum visibility
- **Exact Input Required**: Must type "yes" exactly (case-sensitive)
- **Clear Cancellation**: Shows "Operation cancelled" if confirmation not received

## Pattern-Based Document Deletion

The `weave document delete` command supports powerful pattern matching to delete
documents based on filename patterns. The CLI automatically detects whether you're
using shell glob patterns or regex patterns.

### Shell Glob Patterns (Default)

Shell glob patterns use familiar syntax that most users already know:

```bash
# Delete all PNG files starting with 'tmp'
weave document delete MyCollection --pattern "tmp*.png"

# Delete all JPG files
weave document delete MyCollection --pattern "*.jpg"

# Delete files with single character wildcard
weave document delete MyCollection --pattern "file?.txt"

# Delete files with character ranges
weave document delete MyCollection --pattern "doc[0-9].pdf"
weave document delete MyCollection --pattern "image_[a-z].png"

# Using aliases
weave docs d MyCollection --pattern "*.png"
weave docs d MyCollection --pattern "temp*.*"
```

### Regex Patterns (Auto-detected)

When the pattern contains regex-specific characters, it's automatically treated as regex:

```bash
# Delete files with regex patterns
weave document delete MyCollection --pattern "tmp.*\.png"
weave document delete MyCollection --pattern "^prefix.*\.jpg$"
weave document delete MyCollection --pattern ".*\.(png|jpg|gif)$"

# Complex regex patterns
weave document delete MyCollection --pattern "file_\d{4}\.txt"
weave document delete MyCollection --pattern "^(temp|tmp).*\.pdf$"
```

### Pattern Detection Logic

The CLI automatically detects pattern type based on content:

**Shell Glob Indicators:**
- Contains `*`, `?`, `[abc]` but no regex special characters
- Examples: `tmp*.png`, `file?.txt`, `doc[0-9].pdf`

**Regex Indicators:**
- Contains `^`, `$`, `\`, `.*`, `.+`, `.?`, `(`, `)`, `{`, `}`, `|`
- Examples: `tmp.*\.png`, `^file.*\.txt$`, `.*\.(png|jpg)$`

### Examples and Use Cases

```bash
# Clean up temporary files
weave docs d MyCollection --pattern "tmp*.png"
weave docs d MyCollection --pattern "temp*.*"

# Delete specific file types
weave docs d MyCollection --pattern "*.jpg"
weave docs d MyCollection --pattern "*.pdf"

# Delete files with specific naming patterns
weave docs d MyCollection --pattern "backup_*.txt"
weave docs d MyCollection --pattern "old_[0-9]*.log"

# Complex pattern matching
weave docs d MyCollection --pattern ".*\.(png|jpg|gif)$"  # All image files
weave docs d MyCollection --pattern "^temp.*\.pdf$"     # PDFs starting with temp
```

### Safety Features

- **Preview**: Shows all matching documents before deletion
- **Confirmation**: Requires user confirmation unless `--force` is used
- **Pattern Validation**: Validates pattern syntax before execution
- **Error Handling**: Clear error messages for invalid patterns

## Document Count Command

The `weave document count` command (alias: `weave docs C`) allows you to count
documents in one or more collections efficiently.

### Single Collection Count

```bash
# Count documents in a single collection
weave document count MyCollection

# Using alias
weave docs C MyCollection

# Example output:
# ✅ Found 150 documents in collection 'MyCollection'
```

### Multiple Collections Count

```bash
# Count documents in multiple collections
weave document count RagMeDocs RagMeImages

# Using alias
weave docs C RagMeDocs RagMeImages

# Example output:
# 📊 Document Count: 2 Collections
# 
# Counting documents in weaviate-cloud database...
# 
# 1. RagMeDocs: 150 documents
# 2. RagMeImages: 75 documents
# 
# ✅ Total documents across 2 collections: 225
```

### Error Handling

If a collection doesn't exist or there's an error accessing it, the command will
show an error for that specific collection but continue processing others:

```bash
# Example with one failing collection
weave docs C RagMeDocs NonExistentCollection RagMeImages

# Output:
# 📊 Document Count: 3 Collections
# 
# Counting documents in weaviate-cloud database...
# 
# 1. RagMeDocs: 150 documents
# 2. NonExistentCollection: ERROR - Collection 'NonExistentCollection' not found
# 3. RagMeImages: 75 documents
# 
# ✅ Total documents across 2 collections: 225
# ⚠️ Failed to count 1 collection(s)
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

## PDF and Image Processing

### PDF Document Processing

Weave CLI provides comprehensive PDF processing with intelligent text chunking and image extraction:

```bash
# Process PDF with text chunks only
weave docs create MyTextCollection document.pdf --chunk-size 500

# Process PDF with both text and images
weave docs create MyTextCollection document.pdf --image-collection MyImageCollection

# Process PDF with image size filtering (minimum 1KB)
weave docs create MyTextCollection document.pdf \
  --image-collection MyImageCollection \
  --min-image-size 1024
```

#### PDF Processing Features

- **📄 Text Chunking**: Automatically splits PDF text into manageable chunks
- **🖼️ Image Extraction**: Extracts embedded images from PDFs
- **📊 Progress Tracking**: Real-time progress for both text chunks and images
- **🔍 Smart Image Filtering**: Automatically skips decorative images < 5KB (configurable with `--min-image-size`)
- **📋 Metadata Preservation**: Maintains source PDF information for all extracted content

#### Example Output

```bash
$ weave docs create WeaveDocs ~/Desktop/ragme-io.pdf --image-collection WeaveImages

📄 Processing PDF: ragme-io.pdf
🔍 Extracting content from PDF...
✅ Found 3 text chunks and 4 images

📝 Creating text documents (3 chunks):
  [3/3] chunks created

🖼️  Processing extracted images (4 total):
  [1/4] Image 1: ragme-io_image_1.png (497.0 KB)
  [2/4] Image 2: ragme-io_image_2.png (624.8 KB)
  [3/4] Image 3: ragme-io_image_3.png (381.9 KB)
  [4/4] Image 4: ragme-io_image_4.png (697.3 KB)

✅ Successfully processed ragme-io.pdf
   Text chunks: 3 created
   Images: 4 extracted to WeaveImages (filtered out 20 small decorative images)
```

### Image Metadata Enhancement

All images (both directly uploaded and PDF-extracted) automatically include rich metadata:

#### EXIF Data Extraction
- **Camera Information**: Make, model, camera settings
- **Image Properties**: Width, height, orientation
- **Capture Details**: DateTime, exposure, ISO, focal length
- **GPS Location**: Latitude, longitude, altitude (when available)

#### OCR Text Extraction
- **Text Recognition**: Extracts readable text from images using Tesseract OCR
- **Confidence Scoring**: Provides OCR confidence metrics
- **Language Detection**: Identifies text language
- **Screenshot Support**: Excellent for extracting text from screenshots

#### Storage and Processing Metadata
- **Timestamped Paths**: Cloud-friendly format: `YYYYMMDD_HHMMSS_microseconds_filename.ext`
- **Processing Timestamps**: ISO 8601 format timestamps
- **Processing Duration**: Millisecond precision timing
- **Source Tracking**: PDF source information for extracted images

#### Example Metadata Fields

```yaml
# EXIF Metadata
exif_make: "Canon"
exif_model: "EOS R5"
exif_width: 5573
exif_height: 3715
exif_datetime: "2024:03:15 14:30:22"
exif_orientation: 1

# OCR Data
ocr_text: "PortfolioMax.ai - AI-Powered Portfolio Management"
ocr_word_count: 133
ocr_has_text: true
ocr_language: "eng"

# Storage and Processing
storage_path_relative: "images/20251015_102605_703162_screenshot.jpg"
processing_timestamp: "2025-10-15T10:26:05.703162Z"
processing_duration_ms: 1760

# PDF Source (for extracted images)
source_pdf: "/Users/max/Desktop/ragme-io.pdf"
pdf_image_index: 3
url: "pdf://ragme-io.pdf/image_4"
```

### Image Collection URLs

Images extracted from PDFs use a special URL format for grouping in RAGme-io:

```
pdf://filename.pdf/image_N
```

This format enables:
- **Document Grouping**: Images from the same PDF display together
- **RAGme-io Compatibility**: Proper image stacking and preview in RAGme-io application
- **Source Tracking**: Easy identification of image source

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

#### Pattern Matching Issues

**Issue**: Pattern not matching expected documents.

**Symptoms**:
- Pattern returns "No documents found" when documents exist
- Pattern matches unexpected documents

**Solutions**:

1. **Check pattern syntax**:
   ```bash
   # Shell glob (simple wildcards)
   weave docs d MyCollection --pattern "tmp*.png"
   
   # Regex (complex patterns)
   weave docs d MyCollection --pattern "tmp.*\.png"
   ```

2. **Verify filename field**:
   ```bash
   # Check what filename field looks like
   weave docs l MyCollection --limit 5
   ```

3. **Test pattern step by step**:
   ```bash
   # Start with simple patterns
   weave docs d MyCollection --pattern "*.png"
   weave docs d MyCollection --pattern "tmp*"
   ```

4. **Use regex for complex patterns**:
   ```bash
   # For complex matching, use regex
   weave docs d MyCollection --pattern ".*\.(png|jpg|gif)$"
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

# 3. List existing collections
weave collection list

# 4. Create a new collection (if needed)
weave collection create MyNewCollection

# 5. Count documents in collections
weave document count MyCollection

# 6. Count documents in multiple collections
weave document count RagMeDocs RagMeImages

# 7. List documents in a collection
weave document list MyCollection

# 8. View documents in virtual format
weave document list MyCollection --virtual

# 9. Delete documents by pattern
weave document delete MyCollection --pattern "tmp*.png"
weave document delete MyCollection --pattern "*.jpg"
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
    
    # Get document count (more efficient than listing all documents)
    count=$(weave document count "$collection" --no-color --quiet | grep -o '[0-9]\+' | tail -1)
    echo "Documents in $collection: $count"
done
```

### Monitoring and Logging

```bash
# Log collection status
weave collection list --no-color >> /var/log/weave-collections.log

# Monitor document counts (more efficient)
weave document count MyCollection --no-color --quiet >> /var/log/weave-documents.log

# Monitor multiple collections
weave document count RagMeDocs RagMeImages --no-color --quiet >> /var/log/weave-documents.log

# Health check monitoring
weave health check --no-color >> /var/log/weave-health.log
```

---

*For more information, see the [README.md](../README.md) or run `weave --help`.*
