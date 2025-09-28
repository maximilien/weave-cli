# Weave CLI v0.1.9-rc2 Release Notes

**Release Date**: September 28, 2025  
**Version**: 0.1.9-rc2  
**Type**: Release Candidate

## 🚀 Major New Features

### Required Schema Flags for Document Creation

Weave CLI now requires explicit schema specification when creating documents, ensuring proper collection setup and better data organization.

#### New Required Flags

- **`--text`**: Creates collection with text schema (RagMeDocs format)
  - Properties: `url`, `text`, `metadata`
  - Vectorization: Enabled with text2vec-openai
  - Use case: Text documents, PDF text chunks

- **`--image`**: Creates collection with image schema (RagMeImages format)
  - Properties: `url`, `image`, `metadata`, `image_data`
  - Vectorization: Disabled (none) to avoid issues with large base64 data
  - Use case: Image documents, PDF extracted images

#### Usage Examples

```bash
# Create text documents
weave docs create MyTextCollection document.txt --text
weave docs create MyTextCollection document.pdf --text --chunk-size 500

# Create image documents
weave docs create MyImageCollection image.jpg --image
weave docs create MyImageCollection image.png --image

# PDF with both text and images
weave docs create MyTextCollection document.pdf --text --image-collection MyImageCollection --image

# Using aliases
weave docs c MyTextCollection document.txt --text
weave docs c MyImageCollection image.jpg --image
```

## 🔧 Enhanced Features

### Automatic Collection Creation

- Collections are automatically created with the appropriate schema when documents are created
- Proper schema validation ensures data consistency
- Enhanced error handling provides clear feedback

### Improved Schema Management

- Explicit schema type selection for better data organization
- Proper RagMeDocs and RagMeImages schema compatibility
- Automatic vectorization configuration based on schema type

## ⚠️ Breaking Changes

### Document Creation Workflow

**IMPORTANT**: Document creation now requires explicit schema specification.

**Before (v0.1.9-rc1)**:
```bash
weave docs create MyCollection document.txt  # Worked without flags
```

**After (v0.1.9-rc2)**:
```bash
weave docs create MyCollection document.txt --text   # Required --text flag
weave docs create MyCollection image.jpg --image     # Required --image flag
```

### Error Handling

- Missing schema flags now show clear error messages
- Conflicting flags (both `--text` and `--image`) are properly validated
- Better user guidance for correct usage

## 🛠️ Technical Improvements

### New Functions

- `CreateCollectionWithSchema()`: Schema-aware collection creation
- `ensureCollectionExists()`: Automatic collection creation with proper schema
- Enhanced `createCollectionViaREST()`: Support for explicit schema types

### Schema Definitions

- **Text Schema**: Optimized for text content with proper vectorization
- **Image Schema**: Optimized for image content with disabled vectorization
- **Metadata Structure**: Enhanced metadata fields for better compatibility

## 📚 Documentation Updates

- Updated README.md with new document creation examples
- Enhanced USER_GUIDE.md with dedicated document creation section
- Added comprehensive error handling examples
- Updated CHANGELOG.md with detailed technical information

## 🔍 Migration Guide

### For Existing Users

1. **Update your scripts**: Add `--text` or `--image` flags to all `weave docs create` commands
2. **Review your workflows**: Ensure you're using the correct schema type for your data
3. **Test with small files**: Verify the new behavior with test documents before bulk operations

### Example Migration

**Old workflow**:
```bash
weave docs create MyCollection document.txt
weave docs create MyCollection image.jpg
```

**New workflow**:
```bash
weave docs create MyCollection document.txt --text
weave docs create MyCollection image.jpg --image
```

## 🧪 Testing

- All existing tests pass
- New validation tests added for schema flags
- Comprehensive error handling tests
- Integration tests for automatic collection creation

## 🐛 Bug Fixes

- Fixed collection creation with proper schema validation
- Improved error messages for missing required flags
- Enhanced validation logic for conflicting flags

## 📋 Known Issues

- PDF metadata extraction still requires pdfcpu library integration
- Some advanced PDF features may need additional testing

## 🚀 Next Steps

- Monitor user feedback on the new required flags
- Consider adding more schema types if needed
- Enhance PDF processing capabilities
- Improve error messages based on user experience

---

**Full Changelog**: See [CHANGELOG.md](CHANGELOG.md) for complete technical details.

**Documentation**: Updated [README.md](README.md) and [docs/USER_GUIDE.md](docs/USER_GUIDE.md) with new usage examples.

**Support**: For questions or issues, please open a GitHub issue or check the documentation.