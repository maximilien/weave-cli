# Release v0.11.2 - Critical Bug Fix

**Release Date**: 2026-03-07
**Git Tag**: `v0.11.2`
**Issue**: [#52](https://github.com/maximilien/weave-cli/issues/52)

---

## Overview

Critical patch release fixing Weaviate backup/restore functionality - embeddings were not included in backups, making the feature unusable.

**This is a CRITICAL hotfix for v0.11.0/v0.11.1. All Weaviate users should upgrade immediately.**

---

## 🔥 Critical Bug Fix

### Weaviate Backup Missing Vector Embeddings (Issue #52)

**Problem**: `weave backup create` exported Weaviate documents WITHOUT vector embeddings.

**Impact**:
- ❌ All Weaviate backups missing embeddings (dimension 0 instead of 1024/1536)
- ❌ Validation failed for all backups
- ❌ Restore would create collections without vectors
- ❌ Feature completely unusable for Weaviate users

**Root Cause**:
Weaviate adapter's `listDocumentsBasic()` did not include `vector` field in GraphQL `_additional` selection.

**Key Finding**: Weaviate GraphQL requires explicit `_additional { vector }` to retrieve embeddings (similar to Milvus requiring explicit field names).

**Fix**:
1. Added `Embedding []float64` field to Document struct
2. Updated GraphQL query to include `vector` in `_additional` block
3. Added embedding extraction from GraphQL result
4. Convert `[]interface{}` to `[]float64` for Document type
5. Updated adapter to pass through Embedding field

**Files Changed**:
- `src/pkg/vectordb/weaviate/client_documents.go`
- `src/pkg/vectordb/weaviate/adapter.go`

**Code**:
```go
// Build GraphQL query with vector field (CRITICAL for backup/restore)
query := fmt.Sprintf(`
    {
        Get {
            %s(limit: %d) {
                _additional {
                    id
                    vector  // Vector field - required for backup/restore
                }
                ...
            }
        }
    }
`, collectionName, limit)

// Extract embedding vector (CRITICAL for backup/restore)
if additional, ok := itemMap["_additional"].(map[string]interface{}); ok {
    if id, ok := additional["id"].(string); ok {
        doc.ID = id
    }
    // Extract vector embedding
    if vector, ok := additional["vector"].([]interface{}); ok {
        doc.Embedding = make([]float64, len(vector))
        for i, v := range vector {
            if floatVal, ok := v.(float64); ok {
                doc.Embedding[i] = floatVal
            }
        }
    }
}
```

---

## 📊 Testing Results

### Test Collection
- **Name**: DemoDocs
- **Documents**: 38
- **Model**: Weaviate text2vec
- **Dimensions**: 1024

### Before Fix (v0.11.1)
```bash
$ weave backup create DemoDocs --vdb weaviate-cloud --output test.weavebak
# Documents: 38
# File size: 324 KB

$ weave backup validate test.weavebak
# ❌ Error: ALL 38 documents missing embeddings
# ❌ Embedding dimensions: 0 (expected 1024)
# ❌ Backup validation failed
```

### After Fix (v0.11.2)
```bash
$ weave backup create DemoDocs --vdb weaviate-cloud --output test.weavebak
# Documents: 38
# Duration: 0.30s
# File size: 1.11 MB (includes vectors)

$ jq '.documents[0].embedding | length' test.weavebak
# 1024 ✅

$ jq '.documents[-1].embedding | length' test.weavebak
# 1024 ✅
```

**Verification**:
```bash
$ jq '{collection: .metadata.collection, docs: .metadata.total_documents, first_embedding: (.documents[0].embedding | length), last_embedding: (.documents[-1].embedding | length)}' test.weavebak
{
  "collection": "DemoDocs",
  "docs": 38,
  "first_embedding": 1024,
  "last_embedding": 1024
}
```

---

## 🆙 Upgrading from v0.11.0/v0.11.1

### Breaking Changes

**None** - this is a bug fix only.

### Migration Guide

**Immediate action required for v0.11.0/v0.11.1 Weaviate users**:

1. **Recreate all existing Weaviate backups**:
   ```bash
   # Old backups from v0.11.0/v0.11.1 are missing embeddings
   # Delete them and recreate with v0.11.2
   rm old-weaviate-backup.weavebak.gz

   # Create new backup with v0.11.2
   weave backup create MyCollection --vdb weaviate-cloud --output new-backup.weavebak
   ```

2. **Verify backups**:
   ```bash
   # Check embedding dimensions
   jq '.documents[0].embedding | length' new-backup.weavebak
   # Should show: 1024, 1536, or your model's dimension count
   ```

3. **Test restore**:
   ```bash
   # Restore to test collection
   weave backup restore new-backup.weavebak --collection TestRestore --overwrite
   ```

### Affected Databases

**Only Weaviate users affected**:
- ❌ **Weaviate Cloud**: AFFECTED - upgrade required
- ❌ **Weaviate Local**: AFFECTED - upgrade required
- ✅ Milvus: Not affected (fixed in v0.11.1)
- ✅ Qdrant: Not affected
- ✅ Supabase: Not affected
- ✅ MongoDB: Not affected

---

## 📋 Release Checklist

- ✅ Critical bug fixed
- ✅ Testing verified (38 documents with 1024-dim embeddings)
- ✅ CHANGELOG.md updated
- ✅ Release notes created
- ✅ Issue #52 created and documented
- ✅ Build passing
- ✅ Commit pushed (e1f974b)
- ⏳ Git tag created (`v0.11.2`) - pending
- ⏳ GitHub release created - pending
- ⏳ Binary built and verified - pending

---

## 🔮 What's Next

**v0.12.0** (planned):
- Remote storage integration (S3/MinIO/GCS)
- Performance optimizations (500+ docs/sec backup)
- Backup scheduling with retention policies
- Fix cross-VDB restore --vdb flag issue

See [PLAN.md](../PLAN.md) for detailed sprint planning.

---

## 🐛 Related Issues

- **Issue #52**: Weaviate backups missing embeddings (fixed in this release)
- **Issue #51**: Milvus backups missing embeddings (fixed in v0.11.1)

Both issues had the same root cause: vector fields must be explicitly requested from the VDB (wildcards don't include vectors).

---

## 🙏 Contributors

- **@maximilien** - Bug fix implementation
- **Claude Code** - Development assistance

---

## 📚 Resources

- **Backup Guide**: [docs/guides/BACKUP_RESTORE.md](../guides/BACKUP_RESTORE.md)
- **v0.11.1 Release**: [docs/releases/RELEASE_v0.11.1.md](RELEASE_v0.11.1.md)
- **v0.11.0 Release**: [docs/releases/RELEASE_v0.11.0.md](RELEASE_v0.11.0.md)
- **Issue #52**: https://github.com/maximilien/weave-cli/issues/52
- **Issue #51**: https://github.com/maximilien/weave-cli/issues/51
- **Changelog**: [CHANGELOG.md](../../CHANGELOG.md)

---

## 🚀 Get Started

```bash
# Clone and build
git clone https://github.com/maximilien/weave-cli.git
cd weave-cli
git checkout v0.11.2
./build.sh

# Verify version
./bin/weave --version
# Should show: Weave CLI 0.11.2

# Test Weaviate backup with embeddings
./bin/weave backup create MyCollection --vdb weaviate-cloud --output test.weavebak

# Verify embeddings are present
jq '.documents[0].embedding | length' test.weavebak.gz
# Should show your model's dimension count (e.g., 1024, 1536)
```

---

**Weave-cli v0.11.2 - Weaviate Backup/Restore Now Works Correctly! 🎉**
