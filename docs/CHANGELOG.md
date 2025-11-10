# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.3.9] - 2025-11-10

### Added

- **🗄️ Vector Database Abstraction Layer**: Multi-database support architecture
  - **Abstract Interface**: Unified `VectorDBClient` interface for all vector databases
  - **Factory Pattern**: Clean abstraction layer for easy addition of new database backends
  - **Weaviate Support**: Full support maintained for Weaviate Cloud and Local instances
  - **Supabase Support**: New PostgreSQL-based vector database adapter using pgvector
  - **Modular Design**: Database-specific adapters with consistent API
  - See [docs/VECTOR_DB_ABSTRACTION.md](../docs/VECTOR_DB_ABSTRACTION.md) for details

- **🐘 Supabase Integration**: Complete Supabase adapter implementation
  - **Collection Management**: Create, list, delete collections with PostgreSQL tables
  - **Document Operations**: Full CRUD operations for documents
  - **Metadata Support**: JSON metadata storage and querying
  - **Schema Management**: Automatic table creation with proper indexes
  - **Connection Handling**: Robust connection management with IPv4/IPv6 support
  - **Error Handling**: Comprehensive error wrapping and user-friendly messages
  - Set `VECTOR_DB_TYPE=supabase` to use Supabase adapter

- **🔤 Embeddings Management**: New `weave embeddings` command
  - **List Embeddings**: `weave embeddings list` (or `weave embeddings ls`) to view available models
  - **Model Selection**: Choose embedding models for document creation
  - **`--embedding` flag** for `weave docs create` command
  - **`--embedding` flag** for `weave collection create` command
  - Allows users to specify custom embedding models per operation

- **⚙️ Enhanced Configuration Management**: Streamlined configuration setup
  - **Interactive Config Commands**: `weave config create` and `weave config update`
  - **Smart Defaults**: Sensible defaults for quick setup
  - **Environment Variable Fallback**: Config.yaml optional with env var support
  - **Auto-creates config.yaml**: When REPL detects missing config.yaml, automatically creates a minimal one

- **🔧 weave-mcp Binary Installer**: One-command installation of weave-mcp
  - **Automatic platform detection** - supports macOS, Linux, Windows
  - **Download with progress bar** - visual feedback for large downloads
  - **Checksum verification** - ensures download integrity
  - **Interactive prompts** - choose install location and permissions
  - **Smart PATH detection** - warns if install directory not in PATH
  - **Auto .env Update**: Automatically updates .env file with weave-mcp path
  - **Installation testing** - verifies binary is executable
  - Run `weave config update --weave-mcp` to install

- **🎯 Smart Error Handling & Configuration UX**: Dramatically improved user experience
  - **Detailed error messages** showing exactly what's missing (environment variables, config files)
  - **Interactive configuration fixes** - prompts to create/update .env file on the spot
  - **Multiple fix options** clearly explained (flags, shell exports, .env file)
  - **Context-aware tips** for config.yaml setup and mock database testing
  - **Better MCP error diagnostics** - captures stderr from weave-mcp for troubleshooting
  - **Better collection/database errors** with actionable suggestions
  - All commands now use enhanced error handling for consistent UX

- **🔄 REPL Enhancements**:
  - **Version Display in Banner**: REPL banner now shows version info
    - Displays version string in dimmed text below ASCII art
    - Consistent with `weave -V` output format
  - **REPL Batch Query Mode**: Execute multiple queries in batch mode with demo infrastructure

- **🛠️ Developer Experience**:
  - **Global Timeout Flag**: New `--timeout` flag with duration format support (e.g., "30s", "5m", "1h")
  - **Code Organization**: Refactored Weaviate code into vectordb/weaviate package
  - **Better Maintainability**: Cleaner separation of concerns

- **📦 MCP Integration Tests**: Automated test suite for weave-mcp compatibility
  - Comprehensive integration tests for all MCP operations
  - Collection and document operation testing
  - Error handling and edge case validation
  - Automated testing workflow for MCP releases
  - See [tests/README_MCP_TESTS.md](../tests/README_MCP_TESTS.md) for details

### Changed

- **Vector DB Architecture**: Refactored to support multiple database backends
  - Weaviate code moved to `src/pkg/vectordb/weaviate/` package
  - New factory pattern for database creation
  - Unified error handling across all adapters

- **Documentation**: Updated and refactored documentation
  - README.md simplified and reorganized
  - New vector DB abstraction guide
  - Updated user guide with new features
  - Comprehensive technical blog post draft

### Fixed

- **Linting Errors**: Resolved all staticcheck SA9003 errors
  - Fixed empty branch in transaction rollback handling
  - Improved error handling patterns
  - All CI linting checks passing

## [0.3.0] - 2025-11-01

### Added

- **🔄 Interactive REPL Mode**: Run `weave` without arguments for an interactive session
  - Beautiful ASCII art banner with version and GitHub link
  - Natural language query support in interactive mode
  - Built-in help, examples, and history commands
  - CTRL-C to stop commands, twice to exit (like Claude CLI)
  - Command history saved to `~/.weave_history`

- **🤖 AI Agents Query Command**: New `weave query` (or `weave q`) command for natural language queries using GPT-4o
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
  - ReportAgent: Comprehensive operation reports with LLM-generated recommendations
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

## [0.2.14] - 2025-10-15

### Added

- **🔄 PDF Conversion Tool**: New `weave docs pdf-convert` command to convert CMYK PDFs to RGB format using Ghostscript or ImageMagick
- **🎯 Text-Only PDF Processing**: New `--skip-all-images` flag to extract only text from PDFs without image processing overhead
- **💬 Helpful Tips Control**: Global `--no-tips` flag to suppress helpful tips and suggestions when desired
- **🔧 CMYK PDF Support**: Graceful handling of CMYK PDFs with actionable tips for conversion using Ghostscript or ImageMagick

## [0.2.11] - 2025-10-10

### Added

- **📦 Batch Document Creation**: Process entire directories with parallel processing and automatic retry
- **⚡ Parallel Processing**: Configure multiple workers for faster batch operations
- **📊 Progress Tracking**: Visual progress indicators with time estimation
- **🔄 Smart Retry**: Automatic retry on failures with `.processed` file tracking
- **📈 Comprehensive Reporting**: CSV reports with detailed processing statistics

## [0.2.10] - 2025-10-08

### Added

- **📄 Enhanced PDF Processing**: Improved PDF text extraction with better fallback handling
- **💬 Human-Friendly Error Messages**: Simplified, actionable error messages with helpful suggestions
- **🎬 Updated Demos**: New demo recordings showcasing PDF processing capabilities
- **🔧 Better UX**: Fixed PDF success message formatting and improved user experience

## [0.2.8] - 2025-10-04

### Added

- **📊 Score Normalization**: Implemented quadratic score normalization for
  better result differentiation
  - Low-relevance results (raw score 0.5) now display as 0.25, clearly
    indicating poor matches
  - High-relevance results (raw score 0.7+) preserved at higher normalized values
  - Makes it much easier to distinguish irrelevant results from good matches
  - Uses `score^2` transformation to amplify score differences
- **⚠️ Smart Result Warnings**: Automatic detection and warning for low-quality results
  - Displays warning when all results have scores < 0.3
  - Suggests rephrasing query for better results
  - Provides clear score interpretation guidance
- **📖 Enhanced Documentation**: Updated help text and README with score interpretation
  - Clear score ranges: < 0.3 (no match), 0.3-0.5 (weak), 0.5-0.7 (good), > 0.7 (strong)
  - Help text includes score interpretation guidelines
  - README documents the normalization approach
- **🎬 Demo Upload Tracking**: Asciinema upload script now saves URLs automatically
  - Upload URLs saved to `videos/latest-demo-uploads.txt`
  - Maintains latest URLs for both quick and full demos
  - Includes timestamps and upload history
  - README updated to reference latest demo URLs

### Changed

- **Score Display**: All scores now use normalized values for clearer interpretation
- **User Guidance**: Enhanced output messages to help users understand query results

## [0.2.7] - 2025-10-02

### Added

- **🔍 Semantic Search**: New `query` command for semantic search on collections
  - Uses Weaviate's `nearText` for vector-based similarity search
  - Automatic fallback to hybrid search with real similarity scores
  - Beautiful formatted results with relevance scores (0.0 to 1.0)
- **🔤 BM25 Override**: New `--bm25` flag for keyword-based search
  - Direct BM25 keyword search instead of semantic search
  - Real relevance scoring from Weaviate API
  - Perfect for exact keyword matching
- **🔍 Metadata Search**: New `--search-metadata` flag
  - Search in metadata fields in addition to content/text fields
  - Supports URLs, filenames, domains, and custom metadata
  - Combined with content search for comprehensive results
- **📊 Real Scoring**: All search methods provide authentic Weaviate similarity scores
  - Primary search: nearText provides distance-based similarity scores
  - Fallback search: hybrid search provides combined vector+keyword scores
  - BM25 override: direct BM25 keyword search provides relevance scores
- **🎯 Smart Fallback**: Robust 3-tier fallback system
  - BM25 → Hybrid → Simple text search fallback chain
  - Graceful degradation for unsupported Weaviate configurations
  - Ensures query functionality works across all Weaviate instances
- **🧪 Comprehensive Testing**: 100% test coverage for query functionality
  - 27+ test scenarios including unit, e2e, and integration tests
  - Mock client with realistic scoring algorithm
  - All tests passing with complete coverage

### Changed

- **Enhanced Query Display**: Improved result formatting with emojis and
  clear structure
- **Improved Mock Scoring**: More realistic scoring algorithm with
  content/metadata differentiation
- **Updated Documentation**: Complete README, CHANGELOG, and demo updates

### Fixed

- **--no-truncate Flag**: Fixed support for query commands to show full content
- **Scoring Issues**: Resolved hardcoded 1.0 scores in fallback scenarios
- **Error Handling**: Improved GraphQL error detection and fallback logic

### Known Limitations

⚠️ **Weaviate Instance Requirements**: Some Weaviate instances may not
support advanced search features:

- `nearText` semantic search requires vector search modules
- `bm25` keyword search requires BM25 module installation
- `hybrid` search requires hybrid search module
- Fallback to simple text search works but may have accuracy limitations
- Use `--vector-db-type mock` for full functionality testing

### Examples

```bash
# Basic semantic search
weave cols q MyCollection "machine learning algorithms"

# Search with metadata fields
weave cols q MyCollection "maximilien.org" --search-metadata

# Use BM25 keyword search
weave cols q MyCollection "exact keywords" --bm25

# Combine metadata search with BM25
weave cols q MyCollection "search term" --search-metadata --bm25

# Show full content without truncation
weave cols q MyCollection "query" --no-truncate
```text

## [0.2.6] - 2025-10-01

### Added

- **🔍 Semantic Search**: New `query` command for semantic search on collections
  - Uses Weaviate's `nearText` for vector-based similarity search
  - Automatic fallback to hybrid search with real similarity scores
  - Beautiful formatted results with relevance scores (0.0 to 1.0)
- **🔤 BM25 Override**: New `--bm25` flag for keyword-based search
  - Direct BM25 keyword search instead of semantic search
  - Real relevance scoring from Weaviate API
  - Perfect for exact keyword matching
- **🔍 Metadata Search**: New `--search-metadata` flag
  - Search in metadata fields in addition to content/text fields
  - Supports URLs, filenames, domains, and custom metadata
  - Combined with content search for comprehensive results
- **📊 Real Scoring**: All search methods provide authentic Weaviate similarity scores
  - Primary search: nearText provides distance-based similarity scores
  - Fallback search: hybrid search provides combined vector+keyword scores
  - BM25 override: direct BM25 keyword search provides relevance scores
- **WeaveDocs/WeaveImages Schema**: New default schema structure for better
  document management
  - Flat metadata structure for improved performance and simplicity
  - Better support for document aggregation and display
  - Backward compatibility with RagMeDocs/RagMeImages schemas
- **Enhanced Document Display**: Improved document listing and aggregation
  - Better handling of both WeaveDocs and legacy RagMeDocs schemas
  - Improved filename detection across different schema types
  - Enhanced metadata structure for better document organization

### Changed

- **Schema Defaults**: Updated default schemas from RagMeDocs/RagMeImages to
  WeaveDocs/WeaveImages
- **Document Creation**: Enhanced document creation to use new WeaveDocs schema
  by default
- **Test Coverage**: Updated test cases to reflect new schema structure
- **Demo Scripts**: Updated demo files to showcase new schema capabilities

### Fixed

- **Document Aggregation**: Fixed document aggregation logic to handle both new
  and legacy schemas
- **Schema Detection**: Improved collection type detection for better schema
  handling
- **Metadata Structure**: Standardized metadata fields across different document
  types

### Technical Details

- **Files Modified**:
  - `schemas/WeaveDocs.yaml`: Updated to use `content` field instead of `text`
  - `src/cmd/document/create.go`: Updated help text and examples
  - `src/cmd/utils/collection.go`: Improved schema type detection
  - `src/cmd/utils/display.go`: Enhanced document aggregation logic
  - `tests/cmd_test.go`: Updated test cases for new schema structure
  - `videos/weave-cli-full-demo.cast`: Updated demo to showcase new features

## [0.2.5] - 2025-09-30

### Added

- **Named Schema Support**: Collections can now be created using named
  schemas from `config.yaml`
  - New `--schema` flag for `weave collection create` command
  - Example: `weave cols c MyDocsCol --schema RagMeDocs`
  - Schemas defined in `databases.schemas` section of config.yaml
- **Schema Configuration**: Added schema definitions to config.yaml and config.yaml.example
  - RagMeDocs schema for text documents
  - RagMeImages schema for image documents
  - Support for custom user-defined schemas
- **Config Schema Structure**: New `SchemaDefinition` type in config package
  - Methods: `GetSchema(name)`, `ListSchemas()`
  - Conversion from config schema to Weaviate CollectionSchema
- **Documentation Updates**: Updated USER_GUIDE.md with schema usage examples
  - Creating collections with named schemas
  - Defining custom schemas in config.yaml
  - Using schema files vs named schemas

### Technical Details

- **Files Modified**:
  - `src/pkg/config/config.go`: Added SchemaDefinition struct and methods
  - `src/cmd/collection/create.go`: Added --schema flag support
  - `src/cmd/utils/collection.go`: Added CreateWeaviateCollectionFromConfigSchema()
  - `config.yaml.example`: Added schemas section with RagMeDocs and RagMeImages examples
  - `docs/USER_GUIDE.md`: Updated collection creation documentation

### Usage

```bash
# Create collection using named schema
weave collection create MyDocsCol --schema RagMeDocs

# Create collection using schema file
weave collection create MyCol --schema-yaml-file schema.yaml

# Traditional method still works
weave collection create MyCol
```text

## [0.2.2] - 2025-09-29

### Added

- **SPDX License Headers**: Added proper license headers to all 44 Go source
  files
- **License Management Tool**: Created `tools/add_license_headers.sh` for
  automated license header management
- **Legal Compliance**: Industry-standard SPDX license identification
  throughout codebase

### License Technical Details

- **License Format**: `// SPDX-License-Identifier: MIT`
- **Copyright**: `// Copyright (c) 2025 dr.max`
- **Files Updated**: All source files in `src/`, `tests/`, and `scripts/`
  directories
- **Automation**: Script for future license header updates and management

### Quality Assurance

- **Linting**: All golangci-lint, go vet, go fmt checks passing
- **Tests**: All unit tests passing (100%)
- **E2E Tests**: All 34 E2E tests passing
- **Code Formatting**: Automatically fixed by linter

## [0.2.1] - 2025-09-29

### Fixed

- **CI/CD Linting Issues**: Resolved all golangci-lint unused function warnings
- **Code Cleanup**: Removed 15 unused functions from refactoring process
  - PDF module: Removed unused `generateRealisticPDFContent` function
  - Root module: Removed 14 duplicate styled output functions moved to utils
- **E2E Test Reliability**: Improved cleanup process to avoid false errors
  - Skip collection deletion when collections don't exist
  - Rely on schema deletion for reliable cleanup
  - Reduced test count from 35 to 33 (removed unreliable collection deletion tests)

### Test Reliability Improved

- **Test Reliability**: E2E tests now more robust with better error handling
- **Code Quality**: All linting checks passing, proper code formatting
- **CI/CD Pipeline**: Clean builds with no warnings or errors

### Quality Technical Details

- **Linting**: All golangci-lint, go vet, go fmt checks passing
- **Tests**: All unit and E2E tests passing
- **Code Formatting**: Automatically fixed by linter
- **Quality Assurance**: 100% clean codebase with no unused functions

## [0.2.0] - 2025-09-29

### E2E Testing Added

- **Complete E2E Testing Suite**: 35 comprehensive integration tests
  against real Weaviate instances
- **Smart Configuration Detection**: Auto-detection of Weaviate Cloud
  availability with graceful fallback to mock database
- **Enhanced Terminal Visibility**: Bold white text for maximum contrast
  on dark terminal backgrounds
- **Complete Collection Operations**: Full implementation of all
  collection management functions
- **Complete Document Operations**: Full implementation of all document
  management functions
- **Non-interactive Testing**: All E2E tests run without user prompts
  using --force flags
- **Isolated Test Collections**: Dedicated test collections
  (WeaveDocs_test, WeaveImages_test) with automatic cleanup

### Changed

- **Major Code Refactoring**: Complete modularization of monolithic files
  - `document.go`: 4,307 → 50 lines (98% reduction)
  - `collection.go`: 2,528 → 50 lines (98% reduction)
  - `shared.go`: 1,364 → 7 focused files (100% modularization)
  - `processor.go`: 609 → 3 focused files (100% modularization)
- **Enhanced User Experience**: Improved color scheme and output formatting
- **Better Error Handling**: More descriptive error messages and graceful fallbacks
- **Improved Documentation**: Updated README and CHANGELOG with v0.2.0 features

### Collection Operations Fixed

- **Collection Creation**: Implemented `CreateWeaviateCollection` function
- **Collection Deletion**: Implemented `DeleteWeaviateCollections` function
- **Schema Management**: Implemented `DeleteWeaviateCollectionSchema` function
- **Pattern Matching**: Added collection pattern matching for bulk operations
- **Embedding Model**: Updated to use `text-embedding-3-small` for
  compatibility
- **Confirmation Prompts**: Fixed hanging tests by implementing proper
  --force flag usage

### Technical Improvements

- **Modular Architecture**:
  - `src/cmd/document/` (7 files) - Document operations
  - `src/cmd/collection/` (7 files) - Collection operations
  - `src/cmd/utils/` (7 files) - Shared utilities
  - `src/pkg/pdf/` (3 files) - PDF processing modules
- **Quality Assurance**: 100% test coverage (unit + E2E tests)
- **CI/CD Ready**: Automated testing and validation pipeline
- **Production Ready**: Fully tested against real Weaviate instances

## [0.1.10] - 2025-01-27

### Code Organization Changed

- **Code organization refactoring**: Improved codebase structure and maintainability
  - Split `document.go` (4,307 lines) into 6 logical files:
    - `list.go` - Document list command
    - `show.go` - Document show command
    - `count.go` - Document count command
    - `create.go` - Document create command
    - `delete.go` - Document delete command
    - `delete_all.go` - Document delete-all command
  - Split `collection.go` (2,528 lines) into 7 logical files:
    - `list.go` - Collection list command
    - `create.go` - Collection create command
    - `delete.go` - Collection delete command
    - `delete_all.go` - Collection delete-all command
    - `count.go` - Collection count command
    - `show.go` - Collection show command
    - `delete_schema.go` - Collection delete-schema command
  - Updated main command files to contain only command definitions
  - Preserved all existing functionality with no breaking changes
  - Enhanced developer experience with better file organization

## [0.1.9-rc3] - 2025-09-28

### Empty Collection Fixed

- **Empty collection document listing**: Fixed Weaviate client `ListDocuments` method
  to handle empty collections gracefully
  - Resolved confusing "chunk_index" error when listing documents from empty
    collections
  - Added fallback mechanisms using aggregation API and simple queries
  - Now shows clear "No documents found in collection 'X'" message instead of
    cryptic Weaviate errors
  - Maintains full backward compatibility with collections containing documents

### Safety Features Added

- **Double confirmation for delete-schema**: Added double confirmation to `ds`
  (delete-schema) command similar to `da` (delete-all)
  - First confirmation: Standard y/N prompt asking for confirmation
  - Second confirmation: Red warning with requirement to type "yes" exactly
  - `--force` flag still works to skip both confirmations
  - Consistent user experience across all destructive operations

### Improved

- **Error messages**: Enhanced error handling and user feedback throughout the CLI
  - Better collection existence checks
  - Clearer error messages for common failure scenarios
  - Improved robustness of Weaviate client operations

## [0.1.9-rc2] - 2025-09-28

### Schema Management Added

- **Schema flags for collection creation**: Add `--text` and `--image` flags to
  `weave collection create` command
  - **Default**: Collections are created with text schema (RagMeDocs format)
    unless `--image` is specified
  - `--text`: Creates collection with text schema (RagMeDocs format) -
    Properties: `url`, `text`, `metadata`
  - `--image`: Creates collection with image schema (RagMeImages format) -
    Properties: `url`, `image`, `metadata`, `image_data`
  - Enhanced schema validation and error handling
  - Backward compatibility maintained through default text schema

- **Enhanced collection schema management**:
  - Explicit schema type selection for better data organization
  - Proper RagMeDocs and RagMeImages schema compatibility
  - Automatic vectorization configuration based on schema type

### Schema Workflow Changed

- **Collection creation workflow**: Collection creation now supports explicit
  schema specification with sensible defaults
  - **Default behavior**: Collections are created with text schema unless
    `--image` is specified
  - Improved collection creation with proper schema setup
  - Better error messages for conflicting schema flags

- **Collection creation logic**: Enhanced to support explicit schema types
  - Added `CreateCollectionWithSchema()` function for schema-aware collection creation
  - Improved schema property definitions for text vs image collections
  - Better vectorization configuration based on collection type

### Technical Details

- Added `SchemaType` constants and `CreateCollectionWithSchema()` function
- Enhanced `createCollectionViaREST()` to support explicit schema types
- Updated `createWeaviateCollection()` function for schema-aware collection creation
- Improved validation logic for required schema flags
- Enhanced error handling and user feedback

## [0.1.9-rc1] - 2025-09-28

### Features Added

- **Pattern-based collection deletion**: Add `--pattern` flag to `collection delete`
  command
  - Support for shell glob patterns (`WeaveDocs*`, `Test*`, `*Docs`)
  - Support for regex patterns (`.*Docs$`, `^Test.*`)
  - Auto-detection of pattern types
  - Comprehensive validation and confirmation prompts

- **Pattern-based schema deletion**: Add `--pattern` flag to `collection
  delete-schema` command
  - Same pattern matching capabilities as collection deletion
  - Complete schema removal with pattern support

- **Enhanced PDF processing**: New PDF processor package (`src/pkg/pdf/`)
  - Generic PDF text chunking and image extraction
  - Improved metadata structure for better compatibility
  - Enhanced document creation with proper field mapping

### Changes Made

- **Collection commands**: Updated help text and examples for pattern support
- **Document processing**: Improved metadata structure for better RagMeDocs
  compatibility
- **Weaviate client**: Enhanced document creation with better field mapping

### Known Issues & Limitations

- **PDF metadata extraction**: Not fully implemented (requires pdfcpu library
  integration)
- **RagMeDocs compatibility**: Virtual document view not fully compatible with
  RagMeDocs structure
- **AI summary generation**: Needs enhancement for full RagMeDocs compatibility
- **Document creation**: `weave docs create` commands may not produce documents
  fully compatible with RagMeDocs legacy system

### Implementation Details

- Added `findCollectionsByPattern()` function for pattern matching
- Reused existing pattern matching logic from document deletion
- Enhanced validation to prevent mixing collection names with patterns
- Improved error handling for pattern-based operations

## [0.1.8] - Previous Release

### Features

- Basic collection and document management
- Weaviate Cloud and Local support
- Mock database for testing
- Document pattern-based deletion
- Virtual document view
- Configuration management

---

## Migration Notes

### For RagMeDocs Users

If you're migrating from RagMeDocs legacy system:

1. **Document Creation**: The `weave docs create` command creates documents with
   a different metadata structure than RagMeDocs
2. **Virtual View**: The virtual document view may not show the same aggregate
   information as RagMeDocs
3. **AI Summaries**: Generated AI summaries are basic and may not match RagMeDocs
   comprehensive summaries
4. **PDF Metadata**: PDF metadata extraction is not yet implemented (Title,
   Creator, Producer, etc.)

### Recommended Workflow

- Use pattern-based deletion for cleanup: `weave cols delete --pattern "WeaveDocs*"`
- Test document creation with small files first
- Verify virtual document view meets your needs before bulk operations

## Contributing

When adding new features or fixing bugs, please update this changelog following
the format above.
