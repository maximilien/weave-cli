# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.9-rc1] - 2025-09-28

### Added
- **Pattern-based collection deletion**: Add `--pattern` flag to `collection delete` command
  - Support for shell glob patterns (`WeaveDocs*`, `Test*`, `*Docs`)
  - Support for regex patterns (`.*Docs$`, `^Test.*`)
  - Auto-detection of pattern types
  - Comprehensive validation and confirmation prompts
- **Pattern-based schema deletion**: Add `--pattern` flag to `collection delete-schema` command
  - Same pattern matching capabilities as collection deletion
  - Complete schema removal with pattern support
- **Enhanced PDF processing**: New PDF processor package (`src/pkg/pdf/`)
  - Generic PDF text chunking and image extraction
  - Improved metadata structure for better compatibility
  - Enhanced document creation with proper field mapping

### Changed
- **Collection commands**: Updated help text and examples for pattern support
- **Document processing**: Improved metadata structure for better RagMeDocs compatibility
- **Weaviate client**: Enhanced document creation with better field mapping

### Known Issues & Limitations
- **PDF metadata extraction**: Not fully implemented (requires pdfcpu library integration)
- **RagMeDocs compatibility**: Virtual document view not fully compatible with RagMeDocs structure
- **AI summary generation**: Needs enhancement for full RagMeDocs compatibility
- **Document creation**: `weave docs create` commands may not produce documents fully compatible with RagMeDocs legacy system

### Technical Details
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

1. **Document Creation**: The `weave docs create` command creates documents with a different metadata structure than RagMeDocs
2. **Virtual View**: The virtual document view may not show the same aggregate information as RagMeDocs
3. **AI Summaries**: Generated AI summaries are basic and may not match RagMeDocs comprehensive summaries
4. **PDF Metadata**: PDF metadata extraction is not yet implemented (Title, Creator, Producer, etc.)

### Recommended Workflow
- Use pattern-based deletion for cleanup: `weave cols delete --pattern "WeaveDocs*"`
- Test document creation with small files first
- Verify virtual document view meets your needs before bulk operations

## Contributing

When adding new features or fixing bugs, please update this changelog following the format above.