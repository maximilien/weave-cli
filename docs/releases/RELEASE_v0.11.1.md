# Release v0.11.1 - Critical Bug Fix

**Release Date**: 2026-03-06
**Git Tag**: `v0.11.1`
**Issue**: [#51](https://github.com/maximilien/weave-cli/issues/51)

---

## Overview

Critical patch release fixing backup/restore functionality - embeddings were not included in backups, making the feature unusable.

**This is a CRITICAL hotfix for v0.11.0. All users should upgrade immediately.**

---

## 🔥 Critical Bug Fix

### Milvus Backup Missing Vector Embeddings (Issue #51)

**Problem**: `weave backup create` exported documents WITHOUT vector embeddings.

**Impact**:
- ❌ All backups missing embeddings (dimension 0 instead of 1536)
- ❌ Validation failed for all backups
- ❌ Restore would create collections without vectors
- ❌ Feature completely unusable for Milvus users

**Root Cause**:
Milvus adapter's `ListDocuments()` did not include `FieldEmbedding` in `outputFields`.

**Key Finding**: Milvus wildcard `"*"` does NOT include vector fields - they must be specified explicitly.

**Fix**:
1. Added `FieldEmbedding` to outputFields explicitly
2. Extract embedding from FloatVector column
3. Convert `[]float32` to `[]float64` for Document type
4. Set embedding on document before return

**File Changed**: `src/pkg/vectordb/milvus/document.go`

**Code**:
```go
// Query all documents with pagination
// IMPORTANT: Must explicitly include FieldEmbedding (vectors not included in "*")
outputFields := []string{
    FieldDocumentID, FieldText, FieldContent,
    FieldImage, FieldImageData, FieldURL, FieldMetadata,
    FieldEmbedding, // Vector field - required for backup/restore
}

// ... later in the loop ...

// Extract embedding vector (CRITICAL for backup/restore)
var embedding []float64
if embeddingCol := result.GetColumn(FieldEmbedding); embeddingCol != nil {
    if floatVecCol, ok := embeddingCol.(*entity.ColumnFloatVector); ok {
        // Convert []float32 to []float64
        float32Vec := floatVecCol.Data()[i]
        embedding = make([]float64, len(float32Vec))
        for j, v := range float32Vec {
            embedding[j] = float64(v)
        }
    }
}

doc := c.fromMilvusDocument(...)
doc.Embedding = embedding // Set the embedding vector
```

---

## 📊 Testing Results

### Test Collection
- **Name**: AuctionResults
- **Documents**: 127
- **Model**: text-embedding-3-small
- **Dimensions**: 1536

### Before Fix (v0.11.0)
```bash
$ weave backup create AuctionResults --milvus-local --output test.weavebak
# Documents: 127

$ weave backup validate test.weavebak
# ❌ Error: ALL 127 documents missing embeddings
# ❌ Embedding dimensions: 0 (expected 1536)
# ❌ Backup validation failed
```

### After Fix (v0.11.1)
```bash
$ weave backup create AuctionResults --milvus-local --output test.weavebak
# Documents: 127
# Duration: 0.10s
# File size: 6.6MB (includes vectors)

$ weave backup validate test.weavebak
# ✅ Backup is valid and ready to restore!
# ✅ Collection: AuctionResults
# ✅ Documents: 127
# ✅ All embeddings present (1536 dimensions)
```

---

## 🆙 Upgrading from v0.11.0

### Breaking Changes

**None** - this is a bug fix only.

### Migration Guide

**Immediate action required for v0.11.0 users**:

1. **Recreate all existing backups**:
   ```bash
   # Old backups from v0.11.0 are missing embeddings
   # Delete them and recreate with v0.11.1
   rm old-backup.weavebak.gz

   # Create new backup with v0.11.1
   weave backup create MyCollection --output new-backup.weavebak
   ```

2. **Verify backups**:
   ```bash
   weave backup validate new-backup.weavebak.gz
   # Should show: "✅ Backup is valid and ready to restore!"
   ```

3. **Check embedding dimensions**:
   ```bash
   weave backup validate new-backup.weavebak.gz --json | jq '.vector_dimensions'
   # Should show: 1536 (or your model's dimension count)
   ```

### Affected Databases

**Only Milvus users affected**:
- ✅ Weaviate: Not affected (different implementation)
- ✅ Qdrant: Not affected
- ✅ Supabase: Not affected
- ✅ MongoDB: Not affected
- ❌ **Milvus Local**: AFFECTED - upgrade required
- ❌ **Milvus Cloud (Zilliz)**: AFFECTED - upgrade required

---

## 📋 Release Checklist

- ✅ Critical bug fixed
- ✅ Testing verified (127 documents with embeddings)
- ✅ CHANGELOG.md updated
- ✅ Release notes created
- ✅ Issue #51 closed
- ✅ Build passing
- ✅ Git tag created (`v0.11.1`)

---

## 🔮 What's Next

**v0.12.0** (planned):
- Incremental backups
- Backup encryption
- Remote storage (S3/MinIO/GCS)
- Backup comparison

See [ROADMAP.md](../ROADMAP.md) for full roadmap.

---

## 🙏 Contributors

- **@maximilien** - Bug fix implementation
- **Claude Code** - Development assistance

---

## 📚 Resources

- **Backup Guide**: [docs/guides/BACKUP_RESTORE.md](../guides/BACKUP_RESTORE.md)
- **v0.11.0 Release**: [docs/releases/RELEASE_v0.11.0.md](RELEASE_v0.11.0.md)
- **Issue #51**: https://github.com/maximilien/weave-cli/issues/51
- **Changelog**: [CHANGELOG.md](../../CHANGELOG.md)

---

## 🚀 Get Started

```bash
# Clone and build
git clone https://github.com/maximilien/weave-cli.git
cd weave-cli
git checkout v0.11.1
./build.sh

# Verify version
./bin/weave --version
# Should show: Weave CLI 0.11.1

# Test backup with embeddings
./bin/weave backup create MyCollection --milvus-local --output test.weavebak
./bin/weave backup validate test.weavebak.gz
# Should show: ✅ Backup is valid and ready to restore!
```

---

**Weave-cli v0.11.1 - Backup/Restore Now Works Correctly! 🎉**
