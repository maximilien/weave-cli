# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.10] - 2025-01-27

### Changed

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

### Fixed

- **Empty collection document listing**: Fixed Weaviate client `ListDocuments` method
  to handle empty collections gracefully
  - Resolved confusing "chunk_index" error when listing documents from empty
    collections
  - Added fallback mechanisms using aggregation API and simple queries
  - Now shows clear "No documents found in collection 'X'" message instead of
    cryptic Weaviate errors
  - Maintains full backward compatibility with collections containing documents

### Added

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
