# Backup & Restore Guide

Complete guide to backing up and restoring vector database collections with Weave CLI.

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Commands](#commands)
- [Remote Storage](#remote-storage)
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

## Remote Storage

Weave CLI supports uploading backups to and downloading backups from remote storage backends, enabling cloud-based disaster recovery and automated backup workflows.

### Supported Backends

- **AWS S3**: Amazon's object storage service
- **MinIO**: Self-hosted S3-compatible storage

### Key Features

✅ **Automatic Upload/Download**: Seamlessly integrated with backup create/restore
✅ **Environment Variable Support**: Use `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`
✅ **Path Prefixes**: Organize backups within buckets using prefixes (e.g., `backups/`)
✅ **SSL/TLS Support**: Configurable for both S3 (always SSL) and MinIO (optional)
✅ **Remote-Only Mode**: Upload without keeping local copy (`--remote-only`)
✅ **Flexible Cleanup**: Control whether downloaded files are kept or deleted

---

### S3 Configuration

#### Prerequisites

1. AWS account with S3 access
2. IAM user with S3 permissions:
   - `s3:PutObject` (for uploads)
   - `s3:GetObject` (for downloads)
   - `s3:ListBucket` (optional, for listing)
   - `s3:DeleteObject` (optional, for cleanup)
3. Access key ID and secret access key

#### Create Backup to S3

```bash
# Upload backup to S3 (keeps local copy)
weave backup create MyCollection --output backup.weavebak \
  --remote-storage s3 \
  --s3-bucket my-backups \
  --s3-region us-east-1 \
  --s3-access-key AKIAIOSFODNN7EXAMPLE \
  --s3-secret-key wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY

# Upload to S3 with path prefix
weave backup create MyCollection --output backup.weavebak \
  --remote-storage s3 \
  --s3-bucket my-backups \
  --s3-region us-west-2 \
  --s3-prefix backups/production

# Upload only (delete local file after)
weave backup create MyCollection --output backup.weavebak \
  --remote-storage s3 \
  --s3-bucket my-backups \
  --remote-only
```

#### Restore from S3

```bash
# Download and restore from S3
weave backup restore backup.weavebak.gz \
  --remote-storage s3 \
  --s3-bucket my-backups \
  --s3-region us-east-1

# Download and restore with path prefix
weave backup restore backup.weavebak.gz \
  --remote-storage s3 \
  --s3-bucket my-backups \
  --s3-region us-east-1 \
  --s3-prefix backups/production

# Keep downloaded file for inspection
weave backup restore backup.weavebak.gz \
  --remote-storage s3 \
  --s3-bucket my-backups \
  --keep-local
```

#### Using Environment Variables

Instead of passing credentials via flags, use environment variables:

```bash
# Set AWS credentials
export AWS_ACCESS_KEY_ID="AKIAIOSFODNN7EXAMPLE"
export AWS_SECRET_ACCESS_KEY="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"

# Create backup (credentials read from env)
weave backup create MyCollection --output backup.weavebak \
  --remote-storage s3 \
  --s3-bucket my-backups \
  --s3-region us-east-1

# Restore backup (credentials read from env)
weave backup restore backup.weavebak.gz \
  --remote-storage s3 \
  --s3-bucket my-backups
```

**Best Practice**: Use environment variables for credentials in production to avoid exposing secrets in command history or scripts.

---

### MinIO Configuration

MinIO is a high-performance, S3-compatible object storage system that you can self-host.

#### Prerequisites

1. MinIO server running (local or remote)
2. MinIO access key and secret key
3. Endpoint URL (e.g., `localhost:9000`)

#### Start MinIO (Docker)

```bash
# Run MinIO locally
docker run -d \
  -p 9000:9000 \
  -p 9001:9001 \
  --name minio \
  -e MINIO_ROOT_USER=minioadmin \
  -e MINIO_ROOT_PASSWORD=minioadmin \
  -v /data/minio:/data \
  minio/minio server /data --console-address ":9001"

# Create bucket (via MinIO console at http://localhost:9001)
# Or use mc client:
mc alias set local http://localhost:9000 minioadmin minioadmin
mc mb local/weave-backups
```

#### Create Backup to MinIO

```bash
# Upload backup to MinIO (local, no SSL)
weave backup create MyCollection --output backup.weavebak \
  --remote-storage minio \
  --s3-bucket weave-backups \
  --s3-endpoint localhost:9000 \
  --s3-access-key minioadmin \
  --s3-secret-key minioadmin \
  --s3-no-ssl

# Upload to MinIO with path prefix
weave backup create MyCollection --output backup.weavebak \
  --remote-storage minio \
  --s3-bucket weave-backups \
  --s3-endpoint minio.company.com:9000 \
  --s3-prefix backups/$(date +%Y%m%d) \
  --s3-no-ssl

# Upload only (delete local copy)
weave backup create MyCollection --output backup.weavebak \
  --remote-storage minio \
  --s3-bucket weave-backups \
  --s3-endpoint localhost:9000 \
  --s3-no-ssl \
  --remote-only
```

#### Restore from MinIO

```bash
# Download and restore from MinIO
weave backup restore backup.weavebak.gz \
  --remote-storage minio \
  --s3-bucket weave-backups \
  --s3-endpoint localhost:9000 \
  --s3-access-key minioadmin \
  --s3-secret-key minioadmin \
  --s3-no-ssl

# Download with path prefix
weave backup restore backup.weavebak.gz \
  --remote-storage minio \
  --s3-bucket weave-backups \
  --s3-endpoint localhost:9000 \
  --s3-prefix backups/20260310 \
  --s3-no-ssl

# Keep downloaded file
weave backup restore backup.weavebak.gz \
  --remote-storage minio \
  --s3-bucket weave-backups \
  --s3-endpoint localhost:9000 \
  --s3-no-ssl \
  --keep-local
```

#### MinIO with SSL/TLS

If your MinIO server has SSL/TLS enabled:

```bash
# Upload with SSL (omit --s3-no-ssl flag)
weave backup create MyCollection --output backup.weavebak \
  --remote-storage minio \
  --s3-bucket weave-backups \
  --s3-endpoint minio.company.com:9000 \
  --s3-access-key minioadmin \
  --s3-secret-key minioadmin
```

---

### Remote Storage Flags

#### Backup Create Flags

- `--remote-storage`: Storage type (`s3` or `minio`)
- `--s3-bucket`: Bucket name (required for remote storage)
- `--s3-region`: AWS region (default: `us-east-1`, S3 only)
- `--s3-endpoint`: MinIO endpoint (e.g., `localhost:9000`, MinIO only)
- `--s3-access-key`: Access key ID (or use `AWS_ACCESS_KEY_ID` env var)
- `--s3-secret-key`: Secret access key (or use `AWS_SECRET_ACCESS_KEY` env var)
- `--s3-prefix`: Path prefix within bucket (e.g., `backups/`)
- `--s3-no-ssl`: Disable SSL/TLS (MinIO only, default: SSL enabled)
- `--remote-only`: Upload to remote storage only (skip local file creation)
- `--remote-keep-local`: Keep local file after upload (default: true)

#### Backup Restore Flags

- `--remote-storage`: Storage type (`s3` or `minio`)
- `--s3-bucket`: Bucket name (required for remote storage)
- `--s3-region`: AWS region (default: `us-east-1`, S3 only)
- `--s3-endpoint`: MinIO endpoint (e.g., `localhost:9000`, MinIO only)
- `--s3-access-key`: Access key ID (or use `AWS_ACCESS_KEY_ID` env var)
- `--s3-secret-key`: Secret access key (or use `AWS_SECRET_ACCESS_KEY` env var)
- `--s3-prefix`: Path prefix within bucket (e.g., `backups/`)
- `--s3-no-ssl`: Disable SSL/TLS (MinIO only, default: SSL enabled)
- `--keep-local`: Keep downloaded file after restore (default: false)

---

### Automated Backup Workflows

#### Daily Backups to S3

```bash
#!/bin/bash
# /usr/local/bin/weave-backup-s3.sh

export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"

DATE=$(date +%Y%m%d)
COLLECTIONS=("ProductionDocs" "ProductionImages" "ProductionUsers")

for collection in "${COLLECTIONS[@]}"; do
  weave backup create "$collection" \
    --output "${collection}-${DATE}.weavebak" \
    --remote-storage s3 \
    --s3-bucket company-backups \
    --s3-region us-east-1 \
    --s3-prefix "weave-backups/${DATE}" \
    --remote-only \
    --quiet

  if [ $? -eq 0 ]; then
    echo "✅ Backed up $collection to S3"
  else
    echo "❌ Failed to backup $collection"
  fi
done
```

#### Disaster Recovery from S3

```bash
#!/bin/bash
# Restore all collections from specific date

DATE="20260310"
COLLECTIONS=("ProductionDocs" "ProductionImages" "ProductionUsers")

export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"

for collection in "${COLLECTIONS[@]}"; do
  weave backup restore "${collection}-${DATE}.weavebak.gz" \
    --remote-storage s3 \
    --s3-bucket company-backups \
    --s3-prefix "weave-backups/${DATE}" \
    --overwrite

  if [ $? -eq 0 ]; then
    echo "✅ Restored $collection from S3"
  else
    echo "❌ Failed to restore $collection"
  fi
done
```

#### MinIO Local Backups

```bash
#!/bin/bash
# Backup to local MinIO server

DATE=$(date +%Y%m%d)

weave backup create MyCollection \
  --output "mycollection-${DATE}.weavebak" \
  --remote-storage minio \
  --s3-bucket weave-backups \
  --s3-endpoint localhost:9000 \
  --s3-access-key minioadmin \
  --s3-secret-key minioadmin \
  --s3-prefix "backups/${DATE}" \
  --s3-no-ssl \
  --remote-only \
  --quiet

echo "✅ Backup uploaded to MinIO: backups/${DATE}/mycollection-${DATE}.weavebak.gz"
```

---

### Troubleshooting Remote Storage

#### S3 Issues

**Issue**: "Access Denied" when uploading

```bash
# Verify credentials
aws s3 ls s3://my-backups --profile myprofile

# Check IAM permissions (need s3:PutObject)
aws iam get-user-policy --user-name myuser --policy-name S3Access
```

**Solution**: Ensure IAM user has `s3:PutObject` permission for the bucket.

**Issue**: "Bucket not found"

```bash
# List available buckets
aws s3 ls

# Create bucket if needed
aws s3 mb s3://my-backups --region us-east-1
```

**Issue**: "Invalid region"

```bash
# Specify correct region
weave backup create MyCollection --output backup.weavebak \
  --remote-storage s3 \
  --s3-bucket my-backups \
  --s3-region us-west-2  # Match bucket's region
```

---

#### MinIO Issues

**Issue**: "Connection refused" to MinIO

```bash
# Check if MinIO is running
docker ps | grep minio

# Check endpoint is accessible
curl http://localhost:9000/minio/health/live

# Start MinIO if needed
docker start minio
```

**Issue**: "SSL error" with MinIO

```bash
# For local MinIO without SSL, use --s3-no-ssl
weave backup create MyCollection --output backup.weavebak \
  --remote-storage minio \
  --s3-endpoint localhost:9000 \
  --s3-no-ssl  # Disable SSL for local development
```

**Issue**: "Access Denied" with MinIO

```bash
# Verify credentials with mc client
mc alias set local http://localhost:9000 minioadmin minioadmin
mc ls local/weave-backups

# Check bucket policy
mc policy get local/weave-backups
```

---

#### General Issues

**Issue**: Downloaded file not cleaned up after restore

Check if `--keep-local` flag was used:

```bash
# Default behavior (cleanup)
weave backup restore backup.weavebak.gz \
  --remote-storage s3 \
  --s3-bucket my-backups

# Explicit cleanup (redundant but clear)
weave backup restore backup.weavebak.gz \
  --remote-storage s3 \
  --s3-bucket my-backups
# Downloaded file deleted automatically

# Keep file for debugging
weave backup restore backup.weavebak.gz \
  --remote-storage s3 \
  --s3-bucket my-backups \
  --keep-local
# Check: ls /tmp/backup.weavebak.gz
```

**Issue**: Large backups timing out

For large collections, increase batch size and monitor progress:

```bash
# Larger batch size for faster backup
weave backup create LargeCollection --output backup.weavebak \
  --batch-size 500 \
  --remote-storage s3 \
  --s3-bucket my-backups

# Upload may take time, but progress is shown
```

**Issue**: Environment variables not working

```bash
# Verify environment variables are set
echo $AWS_ACCESS_KEY_ID
echo $AWS_SECRET_ACCESS_KEY

# Export if not set
export AWS_ACCESS_KEY_ID="your-key"
export AWS_SECRET_ACCESS_KEY="your-secret"

# Re-run command
weave backup create MyCollection --output backup.weavebak \
  --remote-storage s3 \
  --s3-bucket my-backups
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

### Backup Performance (v0.11.3+ Profiling - March 2026)

**Latest profiling results** with default batch size (100):

| Collection | Documents | Vector Dims | Backup Time | Throughput | File Size (Compressed) |
|------------|-----------|-------------|-------------|------------|------------------------|
| DemoDocs | 38 | 1024 | 0.34s | 112 docs/sec | 254 KB |
| WeaveDocs | 79 | 1536 | 0.59s | 134 docs/sec | 552 KB |
| AuctionsImages | 301 | 1536 | 1.64s | 184 docs/sec | 27 KB |

**Batch Size Impact** (tested with 301 documents):

| Batch Size | Backup Time | Throughput | vs Batch=100 |
|------------|-------------|------------|--------------|
| 50 | 2.69s | 112 docs/sec | -51% ❌ |
| **100** (default) | **1.64s** | **184 docs/sec** | baseline |
| **200** | **0.80s** | **376 docs/sec** | **+104%** ✅ |

**Key Finding**: **Batch size 200 is 2x faster than batch size 100!**

### Restore Performance

| Documents | Restore Time | Throughput | Notes |
|-----------|--------------|------------|-------|
| 2 | <1s | N/A | Instant |
| 79 | Failed* | N/A | Metadata issue |
| **301** | **16.5s** | **18 docs/sec** | Full restore with images |
| 2,636 | ~120s | ~22 docs/sec | Projected |

*Failed due to invalid embedding model in source collection, not backup/restore bug

### Performance Tuning

#### 1. Batch Size Optimization ⚡ **HIGHEST IMPACT**

The `--batch-size` flag controls how many documents are fetched per VDB query. **This is the single biggest performance knob.**

**Recommendations:**
- **Small collections (<100 docs)**: Use default (100)
- **Medium collections (100-1000 docs)**: Use 200 for 2x speedup
- **Large collections (1000+ docs)**: Try 200-500 for best performance
- **Memory constrained**: Use 50-100

**Examples:**

```bash
# Fast backup for large collections (2x faster)
weave backup create MyCollection --output backup.weavebak --batch-size 200

# Extra fast for very large collections (test first)
weave backup create MyCollection --output backup.weavebak --batch-size 500

# Memory-efficient backup
weave backup create MyCollection --output backup.weavebak --batch-size 50
```

**Performance Impact** (301 docs, real-world test):
- Batch 50: 2.69s (112 docs/sec)
- Batch 100: 1.64s (184 docs/sec)
- **Batch 200: 0.80s (376 docs/sec)** ← **2x faster!**

#### 2. Compression Settings

**Compression is enabled by default and highly recommended:**

```bash
# Compressed (default, recommended)
weave backup create MyCollection --output backup.weavebak --compress

# Uncompressed (faster but larger files)
weave backup create MyCollection --output backup.weavebak --no-compress
```

**Trade-offs:**
- Compression saves 65-95% disk space
- Compression overhead: ~10% slower (0.16s for 301 docs)
- **Recommendation**: Always use compression unless disk space is unlimited

#### 3. Timing and Scheduling

**Backup During Off-Peak Hours:**
- Reduces load on VDB
- Faster network I/O
- Better for production systems

```bash
# Use cron for scheduled backups
0 2 * * * /usr/local/bin/weave backup create MyCollection --output /backups/daily.weavebak --batch-size 200 --quiet
```

#### 4. Progress Monitoring

```bash
# Default shows progress (recommended for interactive use)
weave backup create MyCollection --output backup.weavebak

# Quiet mode for scripts and cron jobs
weave backup create MyCollection --output backup.weavebak --quiet
```

### Performance Bottlenecks (v0.11.3)

**Primary Bottleneck**: VDB query latency
- Each batch requires a separate VDB query
- Larger batches = fewer queries = faster backups
- **Solution**: Use larger batch sizes (200+)

**Secondary Bottlenecks**:
- Startup overhead (~3-3.5s): Config loading, VDB connection
- Single-threaded processing: Batches fetched sequentially
- JSON serialization: CPU-bound for large documents

**Planned Optimizations (v0.12.0)**:
- Parallel batch processing (goroutines) → 2-3x improvement
- Connection pooling → 10-20% improvement
- Streaming JSON/compression → 5-10% improvement
- **Target**: 500+ docs/sec (current best: 376 docs/sec with batch=200)

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

**Last Updated**: March 10, 2026
**Weave CLI Version**: v0.11.3
