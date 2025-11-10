# Release v0.3.10: Supabase VDB Fixes and Integration Test Improvements

## 🎉 Release Summary

**Version**: v0.3.10
**Date**: November 10, 2025
**Status**: ✅ Ready for Release (Pending CI)
**Type**: Point Release - Critical Bug Fixes for Supabase Support

## 🐛 Critical Fixes

### Supabase Collection Creation Fix

- **Fixed Duplicate Column Errors**: Resolved issue where Supabase collection creation failed when schema properties conflicted with fixed columns
  - Added `fixedColumns` map to track reserved column names (id, content, text, image, image_data, url, metadata, embedding)
  - Schema properties that match fixed columns are now automatically skipped
  - Prevents PostgreSQL errors when creating collections with custom schemas
  - **Impact**: Supabase collections can now be created reliably with custom properties

### Supabase Document Operations

- **Complete Document Command Support**: Added Supabase support to all document operations
  - **Create**: `weave docs create` now works with Supabase
  - **Show**: `weave docs show` now works with Supabase
  - **Delete**: `weave docs delete` now works with Supabase
  - **List**: `weave docs list` now works with Supabase
  - **Impact**: All document management operations now fully functional with Supabase

- **Generic Document Utilities**: Implemented database-agnostic document functions
  - `CreateDocument()` - Works with all vector database types
  - `ShowDocument()` - Works with all vector database types
  - `DeleteDocument()` - Works with all vector database types
  - `ListDocuments()` - Works with all vector database types
  - **Impact**: Consistent behavior across all database adapters

### Document Show Command Improvements

- **Better Database Selection**: Improved `weave docs show` command
  - Uses `GetSelectedVectorDBs()` for consistent database selection
  - Validates that only one database is selected for read operations
  - Clear error messages when multiple databases are specified
  - **Impact**: More reliable and user-friendly document viewing

## 🧪 Testing Improvements

### Comprehensive Integration Tests

- **Weaviate Integration Tests**: Created comprehensive test suite matching Supabase structure
  - Full test coverage for all Weaviate operations
  - Health checks, collections, documents, search, schema, and batch operations
  - Uses vectordb abstraction for consistency
  - **Impact**: Ensures Weaviate adapter reliability and consistency

- **Consistent Test Structure**: Both Weaviate and Supabase tests now follow same pattern
  - Same test organization and coverage
  - Same error handling and validation
  - Same cleanup procedures
  - **Impact**: Easier to maintain and extend test coverage

- **Test Infrastructure Enhancements**: Improved test execution
  - Updated `test.sh` with selective test execution flags
  - `--weaviate` flag for Weaviate-only tests
  - `--supabase` flag for Supabase-only tests
  - `--mcp` flag for MCP-only tests
  - Better test documentation and help text
  - **Impact**: Faster development cycles with targeted testing

### Verified Test Coverage

- ✅ All Supabase integration tests passing
- ✅ All Weaviate integration tests passing
- ✅ Consistent test structure across both adapters
- ✅ All unit tests passing
- ✅ All linting checks passing

## 🔧 Technical Improvements

### MCP Client Configuration

- **Project Root Detection**: Improved MCP server startup
  - Automatically finds project root by looking for `.env` or `go.mod` files
  - Ensures MCP server can locate configuration files when started from subdirectories
  - **Impact**: More reliable MCP server operation in various directory contexts

### Code Quality

- **Better Error Handling**: Improved error messages and handling patterns
- **Code Reusability**: Generic functions reduce code duplication
- **Consistency**: Unified approach across all database adapters

## 📊 Release Metrics

### Code Changes

- **Files Modified**: 11 files
- **Lines Added**: +826 lines
- **Lines Removed**: -58 lines
- **New Files**: 1 file (`tests/weaviate_integration_test.go`)

### Bug Fixes

- **Critical**: 1 (Supabase collection creation)
- **Major**: 4 (Document operations for Supabase)
- **Enhancements**: 3 (Test infrastructure, MCP client, code organization)

## 🎯 Usage Examples

### Fixed Supabase Collection Creation

```bash
# This now works correctly - no duplicate column errors
weave collection create MyCollection --schema custom-schema.yaml

# Custom schema properties that conflict with fixed columns are automatically skipped
```

### Supabase Document Operations

```bash
# Create document in Supabase collection
weave docs create MyCollection --file document.pdf

# Show document from Supabase
weave docs show MyCollection doc-id-123

# Delete document from Supabase
weave docs delete MyCollection doc-id-123

# List documents in Supabase collection
weave docs list MyCollection
```

### Selective Integration Testing

```bash
# Run only Supabase tests
./test.sh integration --supabase

# Run only Weaviate tests
./test.sh integration --weaviate

# Run all integration tests
./test.sh integration
```

## 🚀 Deployment Ready

### Production Readiness Checklist

- ✅ Critical Supabase bugs fixed
- ✅ All document operations working with Supabase
- ✅ Comprehensive integration tests added
- ✅ All tests passing
- ✅ Linting passes
- ✅ Backward compatible
- ✅ No breaking changes

### User Impact

**Critical Fixes**:
- Supabase collections can now be created reliably
- All document operations work with Supabase
- Better error messages and user experience

**No Breaking Changes**:
- All existing functionality preserved
- Fully backward compatible
- Weaviate functionality unchanged

## 🔄 Migration Guide

### For Supabase Users

**No migration needed** - this release fixes bugs in v0.3.9. Simply upgrade to v0.3.10.

**What's Fixed**:
- Collection creation with custom schemas now works correctly
- Document create/show/delete/list operations now fully functional
- Better error messages for troubleshooting

### For Weaviate Users

**No changes required** - all Weaviate functionality remains the same.

**New Benefits**:
- Comprehensive integration tests ensure reliability
- Consistent behavior with Supabase adapter
- Better test infrastructure for development

## 📞 Support Information

### For Users

- Supabase collection creation issues are resolved
- All document operations now work with Supabase
- Use `./test.sh integration --supabase` to verify your setup

### For Developers

- Integration tests provide comprehensive coverage
- Test infrastructure supports selective test execution
- Generic document utilities reduce code duplication

## 🎯 Release Decision

### Approval Status

✅ **APPROVED FOR RELEASE** (Pending CI verification)

This release includes:
- Critical Supabase bug fixes
- Complete document operation support for Supabase
- Comprehensive integration tests
- Improved test infrastructure
- Production-ready implementation

**Ready for immediate deployment after CI verification.**

---

## 📝 Changelog Entry

See `CHANGELOG.md` for detailed v0.3.10 changes.

## 🔗 Links

- **Repository**: https://github.com/maximilien/weave-cli
- **Issues**: https://github.com/maximilien/weave-cli/issues
- **CI Status**: https://github.com/maximilien/weave-cli/actions

---

**Release Manager**: AI Assistant
**Review Status**: ✅ Approved (Pending CI)
**Deployment Status**: Ready after CI verification
**Next Review**: Post-release user feedback
