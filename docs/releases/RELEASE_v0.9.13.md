# Weave CLI v0.9.13 Release Notes

**Release Date**: January 30, 2026
**Release Type**: Production Hardening + Critical Bug Fix
**Previous Version**: v0.9.12

---

## 🚨 CRITICAL FIX - PDF Image Ingestion Panic

### What was broken?

A flag conflict caused the CLI to panic during PDF image ingestion workflows:

```text
panic: unable to redefine 't' shorthand in "stats" flagset: it's already used for "top" flag
```

### What we fixed

- Removed the `-t` shortcut for `--timeout` flag (conflicted with `stats --top`)
- The `stats` command retains `-t` for `--top` functionality
- All timeout operations now use `--timeout` (full flag name only)

### Migration Required

**Before** (v0.9.12 - broken):

```bash
weave health check -t 5s  # ❌ CAUSES PANIC
```

**After** (v0.9.13 - fixed):

```bash
weave health check --timeout 5s  # ✅ WORKS
weave stats MyCollection --top 10  # ✅ WORKS (-t still available here)
```

**Impact**: PDF image ingestion workflows are now fully functional.

---

## 🎯 New Features

### 1. Configuration Validation on Startup

Proactive validation of all vector database configurations before operations:

**Features**:

- Validates required fields (URLs, API keys, credentials)
- Database-specific validators for all 10+ supported VDBs
- Non-blocking warnings with actionable suggestions
- Can be disabled: `export WEAVE_SKIP_CONFIG_VALIDATION=true`

**Example Output**:

```text
⚠️  Configuration Warnings:
  • databases.vector_databases[4] (milvus-cloud).api_key: Milvus Cloud (Zilliz) requires an API key
    💡 Set 'api_key' field for authentication

  • databases.vector_databases[2] (mongodb-cloud).database_url: MongoDB database_url is required
    💡 Set 'database_url' field with MongoDB Atlas connection string
```

**Files**: `src/pkg/config/validation.go` (~500 lines)

### 2. `weave vdb` Command - Database Management

New dedicated command for managing vector database configurations:

**Commands**:

```bash
# List all configured databases
weave vdb list                    # All databases
weave vdb ls                      # Alias
weave vdb list --cloud            # Cloud databases only
weave vdb list --local            # Local databases only

# Show detailed database info
weave vdb info weaviate-cloud     # Connection, auth, collections
weave vdb info milvus-local       # Vector settings, endpoints

# Health checks (placeholder)
weave vdb health                  # Directs to 'weave health check'
```

**Aliases**: Use `weave db` or `weave database` for discoverability

**Output Example**:

```text
📊 Configured Vector Databases (14 total)

NAME                 TYPE                      ENDPOINT                                 DEFAULT
────────────────────────────────────────────────────────────────────────────────────────────
weaviate-cloud       weaviate-cloud            https://c4cclmintjcxv7wj7eetw.c0...     ✓
weaviate-local       weaviate-local            http://localhost:8080
mongodb-cloud        mongodb-cloud             mongodb+srv://weave-cli_db_user...
...
```

**Files**: `src/cmd/vdb/*.go` (~500 lines)

### 3. Rich Error Context for Qdrant, Milvus, Chroma

Extended v0.9.12's error context improvements to more VDBs:

**Features**:

- Structured logging with `WrapError` and `WrapErrorWithContext`
- Operation context: timeout, address, connection hints
- Debug logging on successful operations
- Consistent error handling across all VDBs

**Example Error**:

```text
connection refused [operation=Health vdb=qdrant endpoint=localhost:6334 hint=connection_refused]
```

**Files**:

- `src/pkg/vectordb/qdrant/client.go`
- `src/pkg/vectordb/milvus/client.go`
- `src/pkg/vectordb/chroma/client.go`

---

## 📚 Documentation

### 1. Timeout Configuration Guide

Comprehensive guide for timeout tuning and troubleshooting:

**Location**: `docs/TIMEOUT_CONFIGURATION.md`

**Covers**:

- Operation-specific timeout defaults (Health: 10s, Bulk: 120s)
- Cloud vs local timeout differences (2x multiplier for cloud)
- Configuration examples for common scenarios
- Troubleshooting guide with actionable steps
- Production and development best practices

**Examples**:

```bash
# Quick health check
weave health check --timeout 5s

# Large batch import
weave docs batch --dir ./data --collection Docs --timeout 600s

# Configure in config.yaml
timeout: 30  # seconds (per-database)
```

### 2. Enhanced README

**Updates**:

- Vector Database Management section with `weave vdb` examples
- Timeout configuration examples in Advanced Usage
- Updated Key Features section
- Links to comprehensive guides

---

## 🏗️ Production Hardening Summary

### Changes Overview

- **Config Validation**: Prevents confusing runtime errors
- **Error Context**: Makes debugging easier across all VDBs
- **Database Management**: Better UX for multi-database workflows
- **Documentation**: Comprehensive guides for timeout and deployment

### Total Changes

- ~1,500 lines of new code
- ~800 lines of documentation
- 8 commits with critical fixes
- All tests passing, no regressions

---

## 🚀 Installation & Upgrade

### From Source (Recommended)

```bash
# Clone/pull latest
git clone https://github.com/maximilien/weave-cli.git
cd weave-cli
git checkout v0.9.13

# Build
./build.sh

# Verify version
./bin/weave --version
# Expected: Weave CLI 0.9.13
```

### From Git Tag

```bash
cd weave-cli
git fetch --tags
git checkout v0.9.13
./build.sh
```

### Verify Installation

```bash
# Check version
./bin/weave --version

# Test new features
./bin/weave vdb list
./bin/weave vdb info <your-db-name>

# Test critical fix (should not panic)
./bin/weave health check --timeout 5s
```

---

## 🧪 Testing the Fix

### Original Failing Command (from customer report)

```bash
weave docs create "AuctionListings" \
  "data/tamarkin/2021-tamarkin-auction-catalogue.pdf" \
  --image-collection "AuctionImages" \
  --milvus-local --quiet-config --batch-size 5 \
  --skip-small-images --min-image-size 5120 \
  --max-metadata-length 2000 --verbose
```

**Expected**: No panic, successful image extraction and ingestion

### Additional Tests

```bash
# Config validation
weave vdb list
# Should show warnings for any misconfigured databases

# Database info
weave vdb info <database-name>
# Should show detailed configuration

# Stats command (verify -t still works)
weave stats MyCollection --top 10
# Should show top 10 metadata values
```

---

## 📋 Breaking Changes

### Timeout Flag Shortcut Removed

- **Old**: `weave health check -t 5s`
- **New**: `weave health check --timeout 5s`

**Reason**: Conflict with `stats --top` flag

**Impact**: Low - most users likely use `--timeout` already

**Migration**: Find/replace `-t` with `--timeout` in scripts

---

## 🐛 Known Issues

None reported for v0.9.13.

---

## 📞 Support & Feedback

### Reporting Issues

- GitHub Issues: [weave-cli/issues](https://github.com/maximilien/weave-cli/issues)
- Include version: `weave --version`
- Include logs: `weave <command> --log-level debug --log-file debug.log`

### Quick Support

For critical issues blocking production:

1. Check `docs/TIMEOUT_CONFIGURATION.md` for timeout issues
2. Run `weave vdb list` to verify configuration
3. Use `--log-level debug` for detailed error context
4. Report with full error message and context

---

## 🙏 Acknowledgments

**Special Thanks**:

- AuctionsMax.ai for rapid feedback on the flag conflict issue
- Early testers of v0.9.12 production hardening features

**Contributors**:

- Michael Maximilien (@maximilien)
- Claude Code (AI pair programming)

---

## 📈 What's Next?

**Planned for v0.9.14+**:

- Agent chain telemetry (token usage tracking, cost optimization)
- Retry logic with exponential backoff for VDB connections
- Circuit breaker pattern for cascade failure prevention
- Production deployment guide
- Logging configuration guide

**Roadmap**: See `docs/planning/WORK_PLAN_JAN_29_31.md`

---

## 📜 Full Changelog

See [CHANGELOG.md](../../CHANGELOG.md) for complete details.

---

**Release Tag**: v0.9.13
**Git Commit**: f84cf84
**Build Date**: January 30, 2026
**Status**: Production Ready ✅
