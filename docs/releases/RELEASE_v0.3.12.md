# Release v0.3.12: Configuration Improvements, Supabase Parity, and Enhanced Documentation

## 🎉 Release Summary

**Version**: v0.3.12
**Date**: November 13, 2025
**Status**: ✅ Ready for Release (CI Passed)
**Type**: Point Release - Configuration, Supabase Improvements, and Documentation

## ✨ Major Features

### Global Configuration Support (`~/.weave-cli`)

- **Use Weave from Any Directory**: Configuration files can now be stored in `~/.weave-cli/` for global access
  - No need to have `.env` and `config.yaml` in each working directory
  - Run `weave` commands from any directory without local config files
  - **New Command**: `weave config sync` - Sync local config to `~/.weave-cli/`
  - **Impact**: Significantly improved developer experience and workflow flexibility

- **Configuration Path Resolution**: Enhanced config loading with proper precedence
  - Checks `~/.weave-cli/` first, then local directory
  - Properly tracks loaded env file paths
  - Improved `GetEnvFile()` to return actual loaded file path
  - **Impact**: More reliable configuration management

### Enhanced Embedding Compatibility Display

- **VDB Compatibility Filtering**: `weave embeds ls` now shows which vector databases support each embedding
  - **New Flag**: `--database` - Filter embeddings by VDB type (weaviate-cloud, weaviate-local, supabase, mock)
  - **New Flag**: `--show-compatibility` - Display VDB support for each embedding model
  - Added `SupportedDatabases` field to embedding model metadata
  - **Impact**: Easy to determine compatible embeddings for your chosen VDB

### JSON Output Support

- **Structured Output**: Added `--json` flag support to additional commands
  - `weave health check --json` - JSON health status
  - `weave docs show --json` - JSON document output
  - `weave cols query --json` - JSON collection query results
  - **Impact**: Better integration with scripts and automation

## 🐛 Bug Fixes

### Configuration Loading Fix

- **Fixed Config Loading from Any Directory**: Ensured `~/.weave-cli` config files are properly loaded
  - Properly tracks loaded env file path in package-level variable
  - Improved env file loading logic with file existence checks
  - Fixed `GetEnvFile()` to return actual loaded file path
  - **Impact**: Configuration now works reliably regardless of working directory

### Supabase Collection Name Normalization Fix

- **Preserve Original Collection Names**: Fixed automatic name normalization that caused confusion
  - **Before**: `My_Collection` → `my-collection` (lowercased, underscores → hyphens)
  - **After**: `My_Collection` → `My_Collection` (preserved exactly as specified)
  - Added `quoteIdentifier()` helper for PostgreSQL quoted identifiers
  - Updated all SQL queries to use quoted table names (preserves case)
  - **Impact**: Consistent behavior with Weaviate, matches user expectations

## 🚀 Supabase Parity Improvements

### Enhanced BM25 Search

- **Improved Ranking Quality**: Upgraded BM25 search implementation
  - Uses `ts_rank_cd()` with document length normalization (BM25-like)
  - Better ranking quality comparable to Weaviate's BM25
  - Removed limitations warnings for BM25 and Hybrid search
  - **Impact**: Full BM25 and Hybrid search support for Supabase

### Comprehensive Test Coverage

- **Enhanced Integration Tests**: Expanded Supabase test coverage
  - Collection name preservation tests covering 9 naming patterns
  - Comprehensive verification (11 operations per test case)
  - Added `TESTING_SUPABASE.md` documentation with setup guide
  - **Impact**: Better reliability and easier troubleshooting

## 📚 Documentation Improvements

### Vector Database Support Matrix

- **New Documentation**: `docs/VDB_SUPPORT.md` - Comprehensive feature comparison
  - Feature support matrix for all operations (core, documents, search, embeddings)
  - CLI command compatibility table
  - Database-specific notes and configuration examples
  - Known limitations and workarounds
  - Integration test coverage status
  - Roadmap for future improvements

### Documentation Reorganization

- **Streamlined Structure**: Reorganized documentation into logical directories
  - **New**: `docs/README.md` - Comprehensive documentation index
  - **guides/** - Feature guides (AI, batch processing, demos, etc.)
  - **supabase/** - Supabase-specific documentation
  - **weaviate/** - Weaviate-specific documentation
  - **releases/** - Release notes and checklists
  - **archive/** - Historical and reference documentation
  - **Impact**: Much easier to find relevant documentation

### Supabase Documentation

- **Complete Supabase Guide**: Comprehensive documentation suite
  - `supabase/README.md` - Overview and getting started
  - `supabase/TESTING.md` - Integration test setup and usage
  - `supabase/NAME_FIX.md` - Collection name preservation details
  - `supabase/BM25_IMPROVEMENT.md` - Full-text search optimization roadmap
  - `supabase/TODO.md` - Planned improvements

## 🧪 Testing Improvements

### Enhanced Integration Tests

- **More Embedding Tests**: Expanded embedding integration test coverage
  - Additional tests for both Supabase and Weaviate adapters
  - Better coverage of embedding functionality across VDB implementations

- **Comprehensive Name Preservation Tests**: 9 naming pattern test cases
  - Mixed case, underscores, hyphens, numbers, camel case, snake case
  - Each test case verifies 11 operations (create, exists, batch docs, retrieve, list, count, search, update, delete, etc.)
  - **Total**: 99 operations across all test cases

## 📊 Release Metrics

### Code Changes

- **Commits**: 11 commits since v0.3.11
- **Files Changed**: ~30+ files across features, fixes, tests, and docs
- **Documentation**: 5+ new documentation files, major reorganization

### Feature Summary

- **Major Features**: 3 (Global config, Embedding compatibility, JSON output)
- **Bug Fixes**: 2 (Config loading, Collection name normalization)
- **Improvements**: 2 (Supabase BM25, Test coverage)
- **Documentation**: Major reorganization and new guides

## 🎯 Usage Examples

### Global Configuration

```bash
# Sync local config to ~/.weave-cli for global access
weave config sync

# Now use weave from any directory
cd /any/directory
weave cols ls  # Works without local .env/config.yaml
```

### Embedding Compatibility

```bash
# List all embeddings
weave embeds ls

# Show only embeddings compatible with Supabase
weave embeds ls --database supabase

# Show compatibility for each embedding
weave embeds ls --show-compatibility
```

### JSON Output

```bash
# Get health status as JSON
weave health check --json

# Get document as JSON
weave docs show MyCollection doc-id --json

# Query collection with JSON output
weave cols query MyCollection --json
```

### Supabase Collection Names

```bash
# Create collection with mixed case - name is preserved!
weave cols create "MyTest_Collection"
# Creates: MyTest_Collection (not mytest-collection)

# Works with all operations
weave docs create "MyTest_Collection" --file doc.pdf
weave search semantic "MyTest_Collection" "query"
```

### Enhanced BM25 Search

```bash
# BM25 search now has improved ranking
weave search bm25 MyCollection "search query"

# Hybrid search combines semantic + improved BM25
weave search hybrid MyCollection "search query"
```

## 🚀 Deployment Ready

### Production Readiness Checklist

- ✅ Global configuration support working
- ✅ Supabase collection name preservation fixed
- ✅ Enhanced BM25 search implemented
- ✅ JSON output support added
- ✅ Comprehensive test coverage added
- ✅ Documentation reorganized and improved
- ✅ All tests passing
- ✅ Linting passes
- ✅ CI passed
- ✅ Backward compatible
- ✅ No breaking changes

### User Impact

**Major Improvements**:
- Use weave from any directory without local config files
- Easy to find compatible embeddings for your VDB
- JSON output for better automation
- Supabase collection names preserved as specified
- Better BM25 search quality for Supabase
- Much better documentation organization

**No Breaking Changes**:
- All existing functionality preserved
- Fully backward compatible
- Existing configs continue to work
- Weaviate functionality unchanged

## 🔄 Migration Guide

### For All Users

**No migration needed** - this release is fully backward compatible.

**New Capabilities**:
- Optionally use `weave config sync` to enable global config
- Use `--database` flag with `weave embeds ls` to filter by VDB
- Use `--json` flag for structured output
- Supabase users: Collection names now preserved exactly as specified

### For Supabase Users

**What's Fixed**:
- Collection names preserved (no more automatic normalization)
- Improved BM25 search ranking
- Full BM25 and Hybrid search support (no more limitations)

**What's New**:
- Enhanced integration test coverage
- Comprehensive Supabase documentation in `docs/supabase/`

### For Developers

**New Documentation Structure**:
- Check `docs/README.md` for navigation
- Database-specific docs in `docs/supabase/` and `docs/weaviate/`
- Feature guides in `docs/guides/`
- Release notes in `docs/releases/`

## 📞 Support Information

### For Users

- Global configuration: Use `weave config sync` to enable
- Embedding compatibility: Use `weave embeds ls --show-compatibility`
- JSON output: Use `--json` flag on supported commands
- Supabase: Collection names now work as expected
- Documentation: Check `docs/README.md` for navigation

### For Developers

- Configuration: See `src/pkg/config/paths.go` for path resolution
- Supabase: See `docs/supabase/` for complete documentation
- Testing: See `docs/supabase/TESTING.md` for test setup

## 🎯 Release Decision

### Approval Status

✅ **APPROVED FOR RELEASE** (CI Passed)

This release includes:
- Global configuration support (major UX improvement)
- JSON output support (better automation)
- Supabase parity improvements (BM25, name preservation)
- Enhanced embedding compatibility display
- Comprehensive documentation reorganization
- Production-ready implementation

**Ready for immediate deployment.**

---

## 📝 Changelog Entry

### v0.3.12 (2025-11-13)

#### Added
- Global configuration support via `~/.weave-cli/` directory
- `weave config sync` command to sync local config globally
- `--database` flag for `weave embeds ls` to filter by VDB type
- `--show-compatibility` flag for `weave embeds ls` to show VDB support
- `--json` flag support for `weave health check`, `weave docs show`, and `weave cols query`
- Vector database support matrix documentation (`docs/VDB_SUPPORT.md`)
- Comprehensive Supabase documentation suite (`docs/supabase/`)
- Documentation index and reorganization (`docs/README.md`)

#### Fixed
- Configuration loading from `~/.weave-cli` now works from any directory
- Supabase collection name normalization (names now preserved)
- `GetEnvFile()` now returns actual loaded file path

#### Changed
- Enhanced Supabase BM25 search with `ts_rank_cd()` and length normalization
- Improved embedding model metadata with `SupportedDatabases` field
- Reorganized documentation into logical directories (guides/, supabase/, weaviate/, releases/, archive/)
- Updated main README.md with new documentation structure

#### Improved
- Supabase BM25 and Hybrid search quality (removed limitations)
- Integration test coverage (9 naming patterns, 99 operations)
- Documentation discoverability and organization

---

## 🔗 Links

- **Repository**: https://github.com/maximilien/weave-cli
- **Issues**: https://github.com/maximilien/weave-cli/issues
- **CI Status**: https://github.com/maximilien/weave-cli/actions
- **Documentation**: https://github.com/maximilien/weave-cli/tree/main/docs

---

**Release Manager**: AI Assistant
**Review Status**: ✅ Approved (CI Passed)
**Deployment Status**: Ready for deployment
**Next Review**: Post-release user feedback
