# Backup & Restore Guide

Complete guide to backing up and restoring vector database collections with Weave CLI.

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Commands](#commands)
- [Use Cases](#use-cases)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [Performance](#performance)

---

## Overview

Weave CLI's backup/restore feature allows you to:

- **Prevent Data Loss**: Create snapshots of collections before infrastructure changes
- **Fast Recovery**: Restore collections in minutes, not hours
- **Cross-VDB Migration**: Export from one VDB type, restore to another
- **Space Efficient**: 65-95% compression with gzip
- **Complete Preservation**: Embeddings, metadata, images, and content

### Key Features

✅ Works with **all 15+ supported VDBs**
✅ **Portable format**: `.weavebak` files (JSON + optional gzip)
✅ **Batch processing**: Handles large collections efficiently
✅ **Progress tracking**: Real-time status during backup/restore
✅ **Validation**: Built-in integrity checks
✅ **Collection renaming**: Restore to different collection names

---

## Quick Start

### Backup a Collection

```bash
# Simple backup (compressed by default)
weave backup create MyCollection --output backup.weavebak

# Result: backup.weavebak.gz (auto-adds .gz extension)
```

### Validate Backup

```bash
weave backup validate backup.weavebak.gz
```

### Restore Collection

```bash
# Restore to original name
weave backup restore backup.weavebak.gz

# Restore to different name
weave backup restore backup.weavebak.gz --collection NewName
```

### List Backups

```bash
weave backup list /path/to/backups/
```

---

## Commands

### `weave backup create`

Export a collection to a portable `.weavebak` file.

```bash
weave backup create <collection> --output <file> [flags]
```

**Examples:**

```bash
# Basic backup (compressed)
weave backup create AuctionImages --output backup.weavebak
# Creates: backup.weavebak.gz

# Uncompressed backup
weave backup create AuctionImages --output backup.weavebak --no-compress

# Custom batch size (default: 100)
weave backup create AuctionImages --output backup.weavebak --batch-size 500

# Specific VDB
weave backup create AuctionImages --vdb milvus-local --output backup.weavebak

# Quiet mode (no progress)
weave backup create AuctionImages --output backup.weavebak --quiet
```

**Flags:**

- `--output, -o` (required): Output file path
- `--compress`: Enable gzip compression (default: true)
- `--no-compress`: Disable compression
- `--batch-size`: Documents per batch (default: 100)
- `--quiet, -q`: Suppress progress output
- `--vdb`: Select specific vector database

**What's Backed Up:**

- ✅ Document IDs
- ✅ Embeddings (vectors)
- ✅ Text content
- ✅ Metadata (all fields)
- ✅ Images (base64 data)
- ✅ Image URLs
- ✅ Image thumbnails
- ✅ Collection schema
- ✅ Embedding model name
- ✅ Vector dimensions

---

### `weave backup restore`

Restore a collection from a `.weavebak` file.

```bash
weave backup restore <backup-file> [flags]
```

**Examples:**

```bash
# Restore to original collection name
weave backup restore backup.weavebak.gz

# Restore to different name
weave backup restore backup.weavebak.gz --collection NewName

# Overwrite existing collection
weave backup restore backup.weavebak.gz --overwrite

# Restore to specific VDB
weave backup restore backup.weavebak.gz --vdb weaviate-cloud

# Quiet mode
weave backup restore backup.weavebak.gz --quiet
```

**Flags:**

- `--collection, -c`: Target collection name (default: name from backup)
- `--overwrite`: Delete existing collection before restore
- `--quiet, -q`: Suppress progress output
- `--vdb`: Select target vector database

**Behavior:**

- ✅ Auto-detects compressed vs uncompressed backups
- ✅ Creates collection if it doesn't exist
- ✅ Validates backup format before restore
- ✅ Batch inserts for performance
- ✅ Shows real-time progress
- ⚠️ Fails if collection exists (use `--overwrite` to replace)

---

### `weave backup validate`

Check backup file integrity.

```bash
weave backup validate <backup-file> [flags]
```

**Examples:**

```bash
# Validate backup
weave backup validate backup.weavebak.gz

# JSON output
weave backup validate backup.weavebak.gz --json
```

**Flags:**

- `--json`: Output results in JSON format

**Checks Performed:**

- ✅ File exists and is readable
- ✅ Valid JSON format
- ✅ Correct backup version
- ✅ Collection metadata present
- ✅ All documents have required fields (ID, embedding)
- ✅ Embedding dimensions match metadata
- ✅ No duplicate document IDs

**Exit Codes:**

- `0`: Backup is valid
- `1`: Backup has errors

**Use in CI/CD:**

```bash
#!/bin/bash
if weave backup validate backup.weavebak.gz; then
  echo "✅ Backup valid, proceeding with deployment"
else
  echo "❌ Backup validation failed"
  exit 1
fi
```

---

### `weave backup list`

List all backup files in a directory.

```bash
weave backup list [directory] [flags]
```

**Examples:**

```bash
# List backups in current directory
weave backup list

# List in specific directory
weave backup list /backups/

# JSON output
weave backup list /backups/ --json
```

**Flags:**

- `--json`: Output in JSON format

**Output:**

```
Found 6 backup file(s)

FILENAME                            COLLECTION               DOCS       SIZE   COMPRESSED CREATED
------------------------------------------------------------------------------------------------------------------------
auctionsimages-test.weavebak.gz     AuctionsImages            301   27.50 KB          Yes 2026-03-05
auctionsimages-301.weavebak         AuctionsImages            301  495.18 KB           No 2026-03-05
weavedocs-compressed.weavebak.gz    WeaveDocs                  79  115.38 KB          Yes 2026-03-05
```

---

## Use Cases

### 1. Before Infrastructure Changes

Backup before Docker/Kubernetes updates:

```bash
# Backup all important collections
weave backup create ProductionDocs --output prod-docs-$(date +%Y%m%d).weavebak
weave backup create ProductionImages --output prod-images-$(date +%Y%m%d).weavebak

# Validate backups
weave backup validate prod-docs-*.weavebak.gz
weave backup validate prod-images-*.weavebak.gz

# Now safe to update infrastructure
docker-compose down
docker-compose pull
docker-compose up -d

# Restore if needed
# weave backup restore prod-docs-20260305.weavebak.gz --overwrite
```

---

### 2. Regular Snapshots via Cron

Automate daily backups:

```bash
#!/bin/bash
# /usr/local/bin/weave-backup.sh

BACKUP_DIR="/backups/daily"
DATE=$(date +%Y%m%d)
RETENTION_DAYS=30

# Create backups
weave backup create ProductionDocs \
  --output "$BACKUP_DIR/docs-$DATE.weavebak" \
  --quiet

weave backup create ProductionImages \
  --output "$BACKUP_DIR/images-$DATE.weavebak" \
  --quiet

# Validate
weave backup validate "$BACKUP_DIR/docs-$DATE.weavebak.gz" --json > "$BACKUP_DIR/docs-$DATE.validation.json"
weave backup validate "$BACKUP_DIR/images-$DATE.weavebak.gz" --json > "$BACKUP_DIR/images-$DATE.validation.json"

# Cleanup old backups
find "$BACKUP_DIR" -name "*.weavebak.gz" -mtime +$RETENTION_DAYS -delete

echo "✅ Backup completed: $DATE"
```

**Crontab entry:**

```cron
# Daily backup at 2 AM
0 2 * * * /usr/local/bin/weave-backup.sh >> /var/log/weave-backup.log 2>&1
```

---

### 3. Disaster Recovery

Complete recovery workflow:

```bash
# 1. List available backups
weave backup list /backups/daily/

# 2. Choose most recent valid backup
weave backup validate /backups/daily/docs-20260305.weavebak.gz

# 3. Restore collection
weave backup restore /backups/daily/docs-20260305.weavebak.gz --overwrite

# 4. Verify document count
weave cols list | grep ProductionDocs
```

---

### 4. Cross-VDB Migration

Migrate from Weaviate Cloud to Milvus Local:

```bash
# 1. Backup from Weaviate Cloud
weave backup create MyCollection \
  --vdb weaviate-cloud \
  --output migration.weavebak

# 2. Validate backup
weave backup validate migration.weavebak.gz

# 3. Restore to Milvus Local
weave backup restore migration.weavebak.gz \
  --vdb milvus-local \
  --collection MyCollection

# 4. Verify
weave cols list --vdb milvus-local | grep MyCollection
```

**Supported Migrations:**

✅ Any VDB → Any VDB (all 15+ VDB types supported)

---

### 5. Development/Testing Workflows

Clone production data for testing:

```bash
# Backup production
weave backup create ProdCollection \
  --vdb weaviate-cloud \
  --output prod-snapshot.weavebak

# Restore to local dev environment
weave backup restore prod-snapshot.weavebak.gz \
  --vdb weaviate-local \
  --collection DevCollection

# Now safe to test without affecting production
```

---

## Best Practices

### 1. Always Validate After Backup

```bash
weave backup create MyCollection --output backup.weavebak
weave backup validate backup.weavebak.gz
```

### 2. Use Descriptive Filenames

```bash
# Good: includes collection, date, and purpose
weave backup create Users --output users-pre-migration-20260305.weavebak

# Bad: generic filename
weave backup create Users --output backup.weavebak
```

### 3. Store Backups Off-System

```bash
# Backup to network storage
weave backup create MyCollection --output /mnt/nas/backups/mycollection.weavebak

# Or upload to S3
weave backup create MyCollection --output backup.weavebak
aws s3 cp backup.weavebak.gz s3://my-backups/$(date +%Y%m%d)/
```

### 4. Test Restore Regularly

```bash
# Monthly restore test
weave backup restore latest.weavebak.gz --collection TestRestore
weave docs list --collection TestRestore --limit 10
weave cols delete TestRestore
```

### 5. Monitor Backup Size Growth

```bash
# Track backup sizes over time
weave backup list /backups/ --json | jq '.[] | {collection, docs, size_mb: (.backup_size_bytes / 1024 / 1024)}'
```

### 6. Use Compression for Large Collections

```bash
# Compression is enabled by default
weave backup create LargeCollection --output backup.weavebak
# Result: 65-95% size reduction

# Only disable for debugging
weave backup create LargeCollection --output backup.weavebak --no-compress
```

---

## Troubleshooting

### Backup Creation Issues

**Issue**: "Collection not found"

```bash
# Verify collection exists
weave cols list | grep MyCollection

# Check selected VDB
weave backup create MyCollection --output backup.weavebak --vdb weaviate-cloud
```

**Issue**: "Permission denied" writing backup file

```bash
# Check directory permissions
ls -la /path/to/backups/

# Use absolute path
weave backup create MyCollection --output $PWD/backup.weavebak
```

---

### Restore Issues

**Issue**: "Collection already exists"

```bash
# Option 1: Use --overwrite
weave backup restore backup.weavebak.gz --overwrite

# Option 2: Restore to different name
weave backup restore backup.weavebak.gz --collection NewName

# Option 3: Delete manually first
weave cols delete MyCollection
weave backup restore backup.weavebak.gz
```

**Issue**: "Invalid vectorizer" on restore

This occurs when backup metadata has invalid embedding model name. Check validation:

```bash
weave backup validate backup.weavebak.gz
```

**Workaround**: Edit backup file (advanced):

```bash
# Uncompress if needed
gunzip backup.weavebak.gz

# Edit metadata.embedding_model field
nano backup.weavebak

# Recompress
gzip backup.weavebak
```

---

### Validation Issues

**Issue**: "Missing embeddings"

This is a warning, not an error. Collections without embeddings can still be backed up for metadata/content preservation.

**Issue**: "Embedding dimension mismatch"

```bash
# Check backup metadata
weave backup validate backup.weavebak.gz --json | jq '.vector_dimensions'

# This indicates data inconsistency in source collection
```

---

## Performance

### Backup Performance

| Collection Size | Documents | Backup Time | Throughput | Uncompressed | Compressed | Ratio |
|-----------------|-----------|-------------|------------|--------------|------------|-------|
| Small | 2 | <1s | N/A | 1.92 KB | 668 B | 65% |
| Medium | 79 | 0.29s | 272 docs/sec | 796 KB | 115 KB | 86% |
| **Large** | **301** | **1.54s** | **195 docs/sec** | **495 KB** | **27 KB** | **95%** |
| Very Large* | 2,636 | ~120s | ~22 docs/sec | ~733 MB | ~50-60 MB | ~92% |

*Projected from Issue #43 requirements

### Restore Performance

| Documents | Restore Time | Throughput | Notes |
|-----------|--------------|------------|-------|
| 2 | <1s | N/A | Instant |
| 79 | Failed* | N/A | Metadata issue |
| **301** | **16.5s** | **18 docs/sec** | Full restore with images |
| 2,636 | ~120s | ~22 docs/sec | Projected |

*Failed due to invalid embedding model in source collection, not backup/restore bug

### Optimization Tips

1. **Increase Batch Size** for large collections:
   ```bash
   weave backup create LargeCollection --output backup.weavebak --batch-size 500
   ```

2. **Use Compression** (enabled by default):
   - Saves 65-95% disk space
   - Slightly slower but worth it

3. **Backup During Off-Peak Hours**:
   - Reduces load on VDB
   - Faster network I/O

4. **Monitor Progress**:
   ```bash
   # Default shows progress
   weave backup create MyCollection --output backup.weavebak

   # Quiet mode for scripts
   weave backup create MyCollection --output backup.weavebak --quiet
   ```

---

## File Format

### .weavebak Format

Portable JSON format with optional gzip compression:

```json
{
  "version": "1.0.0",
  "metadata": {
    "collection": "MyCollection",
    "vdb_type": "weaviate-cloud",
    "embedding_model": "text-embedding-3-small",
    "vector_dimensions": 1536,
    "created_at": "2026-03-05T07:00:47-08:00",
    "weave_version": "0.10.3",
    "total_documents": 301,
    "backup_size_bytes": 27648
  },
  "documents": [
    {
      "id": "doc-123",
      "content": "Document text content",
      "text": "Extracted text",
      "embedding": [0.123, 0.456, ...],
      "metadata": {
        "source": "file.pdf",
        "page": 1
      },
      "image": "base64-encoded-data",
      "image_url": "https://...",
      "image_thumbnail": "base64-thumbnail",
      "url": "https://source.com/doc",
      "image_metadata": {
        "width": 800,
        "height": 600
      }
    }
  ]
}
```

### Compression

- **Extension**: `.weavebak.gz` (auto-added when `--compress` used)
- **Algorithm**: gzip (standard)
- **Compression Ratio**: 65-95% depending on data
- **Auto-Detection**: Restore automatically detects compression

---

## FAQ

**Q: Can I backup multiple collections at once?**

Not directly, but use a script:

```bash
for collection in $(weave cols list --json | jq -r '.[].name'); do
  weave backup create "$collection" --output "backups/$collection.weavebak"
done
```

**Q: Are backups portable between VDB types?**

Yes! That's a key feature. Export from any VDB, restore to any other.

**Q: What happens if restore fails mid-way?**

The collection is left in a partial state. Use `--overwrite` to retry:

```bash
weave backup restore backup.weavebak.gz --overwrite
```

**Q: Can I edit backup files?**

Advanced users can edit the JSON, but:
1. Uncompress first: `gunzip backup.weavebak.gz`
2. Edit with care (JSON must remain valid)
3. Re-validate: `weave backup validate backup.weavebak`
4. Recompress: `gzip backup.weavebak`

**Q: How do I backup an entire VDB instance?**

Backup each collection individually (see script above).

**Q: What's the maximum collection size supported?**

No hard limit. Tested up to 301 documents, designed for 2,636+. Batch processing handles large datasets efficiently.

---

## Related Commands

- `weave cols list` - List collections
- `weave cols delete` - Delete collections
- `weave docs list` - List documents in collection
- `weave health check` - Verify VDB connectivity

---

## See Also

- [User Guide](../USER_GUIDE.md) - Complete CLI reference
- [Architecture](../ARCHITECTURE.md) - How backup/restore works internally
- [Weave Stack Guide](../WEAVE_STACK.md) - Stack backup integration
- [Test Guide](../TEST_GUIDE.md) - Testing backup/restore

---

**Last Updated**: March 5, 2026
**Weave CLI Version**: v0.11.0 (unreleased)
