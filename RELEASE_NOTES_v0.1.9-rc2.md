# Weave CLI v0.1.9-rc2 Release Notes

**Release Date**: September 28, 2025  
**Version**: 0.1.9-rc2  
**Type**: Release Candidate

## 🚀 Major New Features

### Required Schema Flags for Collection Creation

Weave CLI now supports explicit schema specification when creating collections,
ensuring proper collection setup and better data organization.

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
# Create text collections
weave collection create MyTextCollection --text
weave collection create MyTextCollection --text --embedding text-embedding-3-small

# Create image collections
weave collection create MyImageCollection --image
weave collection create MyImageCollection --image --field title:text,content:text

# Create multiple collections with same schema
weave collection create Col1 Col2 Col3 --text

# Using aliases
weave cols c MyTextCollection --text
weave cols c MyImageCollection --image
```

## 🔧 Enhanced Features

### Schema-Aware Collection Creation

- Collections are created with the appropriate schema based on the specified flag
- Proper schema validation ensures data consistency
- Enhanced error handling provides clear feedback

### Improved Schema Management

- Explicit schema type selection for better data organization
- Proper RagMeDocs and RagMeImages schema compatibility
- Automatic vectorization configuration based on schema type

## ⚠️ Breaking Changes

### Collection Creation Workflow

**IMPORTANT**: Collection creation now requires explicit schema specification.

**Before (v0.1.9-rc1)**:

```bash
weave collection create MyCollection  # Worked without flags
```

**After (v0.1.9-rc2)**:

```bash
weave collection create MyCollection --text   # Required --text flag
weave collection create MyCollection --image  # Required --image flag
```

### Error Handling

- Missing schema flags now show clear error messages
- Conflicting flags (both `--text` and `--image`) are properly validated
- Better user guidance for correct usage

## 🛠️ Technical Improvements

### New Functions

- `CreateCollectionWithSchema()`: Schema-aware collection creation
- Enhanced `createCollectionViaREST()`: Support for explicit schema types
- Updated `createWeaviateCollection()`: Schema-aware collection creation

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

1. **Update your scripts**: Add `--text` or `--image` flags to all
   `weave collection create` commands
2. **Review your workflows**: Ensure you're using the correct schema type for
   your data
3. **Test with small collections**: Verify the new behavior with test
   collections before bulk operations

### Example Migration

**Old workflow**:

```bash
weave collection create MyCollection
weave collection create MyImageCollection
```

**New workflow**:

```bash
weave collection create MyCollection --text
weave collection create MyImageCollection --image
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

**Documentation**: Updated [README.md](README.md) and
[docs/USER_GUIDE.md](docs/USER_GUIDE.md) with new usage examples.

**Support**: For questions or issues, please open a GitHub issue or check the
documentation.
