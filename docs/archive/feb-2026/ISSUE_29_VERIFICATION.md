# Issue #29 Verification - Milvus 65KB VARCHAR Limit

**Issue**: https://github.com/maximilien/weave-cli/issues/29
**Status**: OPEN (likely already fixed, needs verification)
**Created**: 2026-02-13

---

## Problem Statement

Milvus has a 65KB VARCHAR limit that blocks storing large image data directly in the database.

---

## Solution Already Implemented

### External Storage Feature (v0.9.21-23)
External storage was implemented in v0.9.21-23 with the following features:

1. **Storage Backends**:
   - S3 (AWS)
   - MinIO (OSS, Docker-based)
   - Local filesystem

2. **Image Handling**:
   - Thumbnails (<47KB) stored in VDB for fast preview
   - Full-resolution images stored in external storage
   - URLs stored in VDB metadata for retrieval

3. **CLI Flags** (in `src/cmd/document/create.go`):
   ```bash
   --image-storage {s3|minio|local}
   --s3-bucket <name>
   --s3-region <region>
   --minio-endpoint <host:port>
   --minio-bucket <name>
   --local-storage-path <path>
   ```

4. **Configuration** (in `src/pkg/config/config.go`):
   ```go
   ImageStorageType      string
   ImageStorageBucket    string
   ImageStorageRegion    string
   ImageStorageAccessKey string
   ImageStorageSecretKey string
   ImageStorageEndpoint  string
   ImageStorageUseSSL    bool
   ```

---

## Verification Steps

### 1. Check Implementation
```bash
# Verify flags exist
./bin/weave docs create --help | grep -A 20 "image-storage"

# Expected output:
#   --image-storage string          External storage backend for large images: s3, minio, or local
#   --s3-bucket string              S3 bucket name for image storage
#   --s3-region string              S3 region (default: us-east-1)
#   --minio-endpoint string         MinIO endpoint (default: localhost:9000)
#   --minio-bucket string           MinIO bucket name for image storage
#   --local-storage-path string     Local filesystem path (default: ./storage)
```

### 2. Test with Milvus + MinIO
```bash
# Start MinIO
docker-compose -f docker/docker-compose.minio.yml up -d

# Start Milvus
docker-compose -f docker/docker-compose.milvus.yml up -d

# Test ingestion with external storage
./bin/weave docs create MilvusImages test_image.jpg \
  --milvus-local \
  --image-storage minio \
  --minio-bucket test-images

# Verify:
# 1. Image ingested successfully
# 2. Thumbnail stored in Milvus (< 65KB)
# 3. Full image stored in MinIO
# 4. URL accessible from metadata
```

### 3. Test with Large Image
```bash
# Create a test image > 65KB
convert -size 1024x1024 xc:blue test_large.jpg

# Ingest with external storage
./bin/weave docs create MilvusImages test_large.jpg \
  --milvus-local \
  --image-storage minio \
  --minio-bucket test-images

# Should succeed (without external storage, this would fail)
```

---

## Expected Results

If external storage is working correctly:

1. ✅ Large images (>65KB) ingest successfully with Milvus
2. ✅ Thumbnails stored in Milvus VDB (< 47KB to stay under 65KB limit)
3. ✅ Full-resolution images accessible via URL
4. ✅ No VARCHAR limit errors

---

## Likely Outcome

**External storage already solves Issue #29**

The feature was implemented in:
- v0.9.21: External storage infrastructure
- v0.9.22: Auto-bucket creation for MinIO
- v0.9.23: External storage docs and examples

**Action Items**:
1. Run verification tests (above)
2. If tests pass, close Issue #29 with comment:
   ```
   Fixed in v0.9.21-23 with external storage feature.

   Use `--image-storage minio` (or s3/local) to store full-resolution
   images externally while keeping thumbnails (<47KB) in Milvus VDB.

   Example:
   ./bin/weave docs create MilvusImages images/ \
     --milvus-local \
     --image-storage minio \
     --minio-bucket my-images

   See README.md for full documentation.
   ```
3. Update VDB support matrix to note external storage support for Milvus

---

## Documentation References

- `README.md` - External storage examples
- `docs/VDB_SUPPORT_MATRIX.md` - Feature matrix
- `docs/guides/DEMO.md` - Demo with external storage
- `docs/planning/EXTERNAL_STORAGE_IMPLEMENTATION.md` - Implementation details
- `src/cmd/document/create.go` - CLI flags and help text

---

**Prepared**: 2026-02-15
**Status**: Needs verification testing
**Priority**: Medium (likely already fixed)
**Estimated**: 30 min to test and close
