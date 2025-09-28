# Weave CLI v0.1.9-rc1 Release Notes

**Release Date**: September 28, 2025  
**Type**: Release Candidate  
**Status**: ⚠️ Testing Required

## 🎉 New Features

### Pattern-Based Collection Deletion

- **`weave cols delete --pattern "WeaveDocs*"`** - Delete collections using shell
  glob patterns
- **`weave cols delete-schema --pattern "Test*"`** - Delete collection schemas
  using patterns
- **Auto-detection** of pattern types (shell glob vs regex)
- **Comprehensive validation** and safety prompts

### Enhanced PDF Processing

- **New PDF processor package** (`src/pkg/pdf/`)
- **Generic metadata structure** for better compatibility
- **Improved document creation** with proper field mapping

## 🔧 Technical Improvements

- Reused existing pattern matching logic from document deletion
- Enhanced validation to prevent mixing collection names with patterns
- Improved error handling for pattern-based operations
- Better Weaviate client integration

## ⚠️ Known Limitations & RagMeDocs Compatibility Issues

### Critical Issues for RagMeDocs Users

1. **PDF Metadata Extraction**: Not fully implemented
   - Missing: Title, Creator, Producer, Creation Date, Modification Date
   - Requires: pdfcpu library integration

2. **Virtual Document View**: Not fully compatible with RagMeDocs structure
   - Missing: Comprehensive AI summaries
   - Missing: Rich PDF metadata in virtual view
   - Different: Metadata structure and field names

3. **AI Summary Generation**: Needs enhancement
   - Current: Basic summaries based on document characteristics
   - Missing: Comprehensive content analysis like RagMeDocs
   - Missing: Rich document overviews

4. **Document Creation**: `weave docs create` commands may not produce fully
   compatible documents
   - Different metadata structure than RagMeDocs legacy system
   - Missing aggregate document with comprehensive metadata

## 🚨 Migration Warning

**If you're migrating from RagMeDocs legacy system:**

- ⚠️ **Test thoroughly** before bulk operations
- ⚠️ **Virtual document view** may not show expected information
- ⚠️ **Document creation** produces different metadata structure
- ⚠️ **AI summaries** are basic, not comprehensive

## 📋 Recommended Testing

### Before Production Use

1. **Test pattern deletion** with small collections first
2. **Verify virtual document view** meets your needs
3. **Test document creation** with sample files
4. **Check metadata structure** compatibility

### Test Commands

```bash
# Test pattern deletion (safe - shows matches first)
weave cols delete --pattern "Test*"

# Test document creation
weave docs create TestCollection sample.pdf

# Test virtual view
weave docs list TestCollection --virtual
```

## 🔄 Next Steps

### For Full RagMeDocs Compatibility

1. **Implement PDF metadata extraction** using pdfcpu
2. **Enhance AI summary generation** with content analysis
3. **Improve virtual document view** to match RagMeDocs structure
4. **Add aggregate document creation** with comprehensive metadata

### For Production Readiness

1. **Complete PDF metadata extraction**
2. **Implement comprehensive AI summaries**
3. **Fix virtual document view compatibility**
4. **Add full RagMeDocs metadata structure support**

## 📚 Documentation

- **CHANGELOG.md**: Complete change history
- **USER_GUIDE.md**: Updated usage examples
- **Pattern Examples**: See help for `weave cols delete --help`

## 🐛 Bug Reports

If you encounter issues with RagMeDocs compatibility:

1. Check the known limitations above
2. Test with small collections first
3. Report specific compatibility issues
4. Include sample documents and expected vs actual behavior

---

**This is a release candidate - please test thoroughly before using in production!**
