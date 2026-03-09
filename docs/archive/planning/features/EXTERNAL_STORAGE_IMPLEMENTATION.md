# External Storage Implementation Plan

**Issue**: #29 - Milvus 65KB VARCHAR Limit Blocks Image Storage
**Status**: Foundation Complete, Integration Pending
**Target**: v0.10.0 or v1.0
**Priority**: High (Client Production Blocker)

---

## Overview

Implement external storage (S3/MinIO) with thumbnail generation to solve Milvus 65KB VARCHAR limit for images. Also support storing original PDFs.

**Problem**:
- Milvus VARCHAR limit: 65,535 bytes (65KB)
- Base64 overhead: 1.37x
- Max safe image size: ~47KB
- Client images: 100-500KB (BLOCKED)

**Solution**:
- Store small thumbnail (<47KB) in Milvus
- Store full image in S3/MinIO
- Store URL in Milvus metadata

---

## ✅ Completed (Commits 492f1c7, 551ed4f)

### 1. Storage Package Foundation
```
src/pkg/storage/
├── interface.go    - Storage abstraction (Upload/Download/Delete/Exists)
├── minio.go        - S3 + MinIO implementation (S3-compatible)
├── local.go        - Local filesystem (testing)
├── thumbnail.go    - Thumbnail generation with Milvus limits
└── errors.go       - Storage-specific errors
```

**Features**:
- ✅ S3 storage backend
- ✅ MinIO storage backend (OSS, LGPL-safe)
- ✅ Local filesystem for testing
- ✅ Thumbnail generation (Lanczos, high-quality)
- ✅ Auto-fit to Milvus 65KB limit
- ✅ Pre-signed URLs for temporary access

### 2. Document Model Updates
```go
type Document struct {
    // Existing fields...

    // External storage fields (v0.10.0+)
    ImageThumbnail string                 `json:"image_thumbnail,omitempty"`
    ImageURL       string                 `json:"image_url,omitempty"`
    ImageMetadata  map[string]interface{} `json:"image_metadata,omitempty"`
}
```

---

## ⏳ Remaining Work

### Phase 1: Core Integration (2-3 hours)

#### 1.1 Update Milvus Adapter
**File**: `src/pkg/vectordb/milvus/document.go`

Add external storage to CreateDocument():
```go
func (a *Adapter) CreateDocument(ctx context.Context, collectionName string, document *vectordb.Document) error {
    // Existing code...

    // NEW: Check if image needs external storage
    if storage.IsThumbnailNeeded([]byte(document.ImageData)) {
        // 1. Upload full image to storage
        imageURL, err := a.imageStorage.Upload(ctx, []byte(document.ImageData), storage.ImageMetadata{
            Filename:       document.Image,
            ContentType:    "image/jpeg", // Detect from data
            CollectionName: collectionName,
            DocumentID:     document.ID,
        })

        // 2. Generate safe thumbnail
        thumbnail, err := storage.GenerateSafeThumbnail([]byte(document.ImageData))

        // 3. Update document
        document.ImageThumbnail = thumbnail
        document.ImageURL = imageURL
        document.ImageMetadata = map[string]interface{}{
            "original_size": len(document.ImageData),
            "thumbnail_size": len(thumbnail),
            "storage_backend": "s3", // or "minio"
        }

        // 4. Clear ImageData to save space
        document.ImageData = ""
    }

    // Continue with existing insertion logic...
}
```

**Files to Update**:
- `src/pkg/vectordb/milvus/adapter.go` - Add imageStorage field
- `src/pkg/vectordb/milvus/document.go` - Update CreateDocument() and CreateDocuments()
- `src/pkg/vectordb/milvus/client.go` - Add storage config

#### 1.2 Add Storage Configuration
**File**: `src/pkg/config/config.go`

Add storage config section:
```go
type Config struct {
    // Existing fields...

    // External storage configuration (v0.10.0+)
    ImageStorage *ImageStorageConfig `yaml:"image_storage,omitempty"`
    PDFStorage   *PDFStorageConfig   `yaml:"pdf_storage,omitempty"`
}

type ImageStorageConfig struct {
    Enabled      bool   `yaml:"enabled"`
    Type         string `yaml:"type"` // "s3", "minio", "local"
    Endpoint     string `yaml:"endpoint"`
    AccessKey    string `yaml:"access_key"`
    SecretKey    string `yaml:"secret_key"`
    Region       string `yaml:"region"`
    Bucket       string `yaml:"bucket"`
    PathPrefix   string `yaml:"path_prefix"`
    UseSSL       bool   `yaml:"use_ssl"`
}

type PDFStorageConfig struct {
    Enabled      bool   `yaml:"enabled"`
    Type         string `yaml:"type"`
    // Same fields as ImageStorageConfig
}
```

#### 1.3 Update Milvus Collection Schema
**File**: `src/pkg/vectordb/milvus/collection.go`

Add new fields to schema:
```go
// Add ImageThumbnail, ImageURL, ImageMetadata fields
schema.Fields = append(schema.Fields,
    &entity.Field{
        Name:       "image_thumbnail",
        DataType:   entity.FieldTypeVarChar,
        TypeParams: map[string]string{"max_length": "65535"},
    },
    &entity.Field{
        Name:     "image_url",
        DataType: entity.FieldTypeVarChar,
        TypeParams: map[string]string{"max_length": "512"},
    },
    &entity.Field{
        Name:     "image_metadata",
        DataType: entity.FieldTypeJSON,
    },
)
```

### Phase 2: CLI Integration (1-2 hours)

#### 2.1 Add Storage Flags
**File**: `src/cmd/document/create.go`

Add command-line flags:
```bash
weave docs create Collection document.pdf \
  --include-images \
  --image-storage s3 \
  --s3-bucket my-images \
  --s3-region us-west-2 \
  --s3-access-key $AWS_ACCESS_KEY_ID \
  --s3-secret-key $AWS_SECRET_ACCESS_KEY

# Or with MinIO (OSS)
weave docs create Collection document.pdf \
  --include-images \
  --image-storage minio \
  --minio-endpoint localhost:9000 \
  --minio-bucket weave-images \
  --minio-access-key minioadmin \
  --minio-secret-key minioadmin

# PDF storage
weave docs create Collection document.pdf \
  --store-pdf \
  --pdf-storage s3 \
  --s3-bucket my-pdfs
```

**New Flags**:
- `--image-storage <type>` - Storage type (s3, minio, local)
- `--s3-bucket <name>` - S3 bucket name
- `--s3-region <region>` - AWS region
- `--s3-access-key <key>` - AWS access key (or env var)
- `--s3-secret-key <key>` - AWS secret key (or env var)
- `--minio-endpoint <url>` - MinIO endpoint
- `--minio-bucket <name>` - MinIO bucket
- `--minio-access-key <key>` - MinIO access key
- `--minio-secret-key <key>` - MinIO secret key
- `--store-pdf` - Store original PDF in external storage
- `--pdf-storage <type>` - PDF storage type (s3, minio, local)

#### 2.2 Update Batch Processing
**File**: `src/cmd/document/batch.go`

Add storage flags to batch command:
```bash
weave docs batch --directory ./pdfs \
  --collection AuctionCatalogs \
  --include-images \
  --image-storage minio \
  --minio-endpoint localhost:9000 \
  --minio-bucket auction-images \
  --store-pdf \
  --pdf-storage minio \
  --minio-bucket auction-pdfs \
  --parallel 3
```

### Phase 3: MinIO Local Setup (30 min)

#### 3.1 Docker Compose for MinIO
**File**: `docker-compose.minio.yml`

```yaml
version: '3.8'

services:
  minio:
    image: minio/minio:latest
    container_name: weave-minio
    ports:
      - "9000:9000"      # API
      - "9001:9001"      # Console
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    command: server /data --console-address ":9001"
    volumes:
      - minio-data:/data
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9000/minio/health/live"]
      interval: 30s
      timeout: 20s
      retries: 3

  # Create buckets on startup
  minio-setup:
    image: minio/mc:latest
    depends_on:
      - minio
    entrypoint: >
      /bin/sh -c "
      sleep 5;
      mc alias set local http://minio:9000 minioadmin minioadmin;
      mc mb local/weave-images --ignore-existing;
      mc mb local/weave-pdfs --ignore-existing;
      mc anonymous set public local/weave-images;
      mc anonymous set public local/weave-pdfs;
      "

volumes:
  minio-data:
    driver: local
```

**Start MinIO**:
```bash
docker-compose -f docker-compose.minio.yml up -d

# Verify
curl http://localhost:9000/minio/health/live
# Open console: http://localhost:9001 (minioadmin/minioadmin)
```

#### 3.2 Update README with MinIO Setup
**File**: `README.md`

Add MinIO quick start section.

### Phase 4: PDF Storage Support (1 hour)

#### 4.1 Add PDF to Storage Package
**File**: `src/pkg/storage/interface.go`

Add PDF-specific methods:
```go
type PDFStorage interface {
    // UploadPDF stores a PDF and returns its URL
    UploadPDF(ctx context.Context, pdf []byte, metadata PDFMetadata) (string, error)

    // DownloadPDF retrieves a PDF by its URL
    DownloadPDF(ctx context.Context, url string) ([]byte, error)

    // Same interface as ImageStorage, reuse implementation
}

type PDFMetadata struct {
    Filename       string
    Size           int64
    CollectionName string
    DocumentID     string
    PageCount      int
    Custom         map[string]string
}
```

#### 4.2 Update PDF Processing
**File**: `src/pkg/pdf/pdf.go`

Add option to store PDF:
```go
type ProcessOptions struct {
    // Existing fields...

    // Storage options (v0.10.0+)
    StorePDF      bool
    PDFStorage    storage.ImageStorage // Reuse same interface
    ImageStorage  storage.ImageStorage
}

func ProcessPDFFile(file string, opts ProcessOptions) ([]*Document, error) {
    // Read PDF
    pdfBytes, err := os.ReadFile(file)

    // NEW: Store PDF if requested
    var pdfURL string
    if opts.StorePDF && opts.PDFStorage != nil {
        pdfURL, err = opts.PDFStorage.Upload(ctx, pdfBytes, storage.ImageMetadata{
            Filename:       filepath.Base(file),
            ContentType:    "application/pdf",
            Size:           int64(len(pdfBytes)),
            CollectionName: opts.CollectionName,
        })
    }

    // Process pages...
    // Store pdfURL in document metadata
}
```

### Phase 5: Testing (2-3 hours)

#### 5.1 Unit Tests
**File**: `src/pkg/storage/minio_test.go`

Test MinIO storage:
```go
func TestMinioStorage_Upload(t *testing.T) {
    // Setup MinIO client
    // Upload test image
    // Verify URL returned
    // Download and verify content matches
}

func TestMinioStorage_Thumbnail(t *testing.T) {
    // Upload large image (>47KB)
    // Generate thumbnail
    // Verify thumbnail <65KB
    // Verify can decode thumbnail
}
```

#### 5.2 Integration Tests
**File**: `src/pkg/vectordb/milvus/storage_test.go`

Test Milvus + Storage:
```go
func TestMilvus_CreateDocument_WithExternalStorage(t *testing.T) {
    // Setup MinIO + Milvus
    // Create document with large image
    // Verify thumbnail stored in Milvus
    // Verify full image in MinIO
    // Query and verify results
}
```

#### 5.3 End-to-End Test with Client Data
**File**: Manual testing script

```bash
# 1. Start MinIO
docker-compose -f docker-compose.minio.yml up -d

# 2. Ingest Client's auction catalog (250+ images, 100-500KB each)
weave docs create AuctionImages auction-catalog.pdf \
  --milvus-local \
  --include-images \
  --image-storage minio \
  --minio-endpoint localhost:9000 \
  --minio-bucket auction-images \
  --store-pdf \
  --pdf-storage minio

# 3. Verify all 250+ images uploaded successfully
weave cols show AuctionImages --milvus-local

# 4. Query with image results
weave cols query AuctionImages "vintage Leica camera" \
  --top-k 5 \
  --milvus-local

# 5. Verify thumbnail + URL returned
# 6. Download full image from MinIO URL
# 7. Verify quality
```

---

## Documentation Updates

### 1. Update VDB_SUPPORT_MATRIX.md
Change Milvus image storage from ⚠️ to ✅ with note:
```markdown
| **Image Storage** | ✅ | ✅ | ✅ | ...
- **Milvus Image Storage**: Supports external storage (S3/MinIO) for images >47KB. Use `--image-storage s3` flag. See [External Storage Guide](EXTERNAL_STORAGE.md)
```

### 2. Create EXTERNAL_STORAGE.md Guide
New guide document covering:
- Why external storage is needed
- Supported backends (S3, MinIO, local)
- Configuration options
- CLI usage examples
- MinIO local setup
- Production deployment (S3)
- Cost analysis
- Troubleshooting

### 3. Update README.md
Add external storage section:
- Quick start with MinIO
- S3 production setup
- Cost savings example

### 4. Update Issue #29
Close with resolution:
- Link to implementation
- Mention MinIO support
- Provide usage examples

---

## Timeline

| Phase | Task | Effort | Status |
|-------|------|--------|--------|
| 0 | Storage foundation | 3h | ✅ Complete |
| 1 | Core integration | 2-3h | ⏳ Pending |
| 2 | CLI integration | 1-2h | ⏳ Pending |
| 3 | MinIO setup | 30m | ⏳ Pending |
| 4 | PDF storage | 1h | ⏳ Pending |
| 5 | Testing | 2-3h | ⏳ Pending |
| 6 | Documentation | 1h | ⏳ Pending |

**Total**: ~10-13 hours
**Completed**: ~3 hours
**Remaining**: ~7-10 hours

---

## Success Criteria

✅ **Foundation**: Storage package implemented
⏳ **Integration**: Milvus adapter uses external storage
⏳ **CLI**: Storage flags work end-to-end
⏳ **Testing**: Client's 250+ images (100-500KB) work
⏳ **MinIO**: Local setup documented and tested
⏳ **S3**: Production deployment documented
⏳ **PDF**: Original PDFs stored in S3/MinIO
⏳ **Docs**: Complete user guide published

---

## Next Session Priorities

1. **Milvus Adapter Integration** (2-3h) - Highest priority
2. **CLI Flags** (1h) - Enable user testing
3. **MinIO Docker Compose** (30m) - Quick win
4. **Basic Integration Test** (1h) - Validate approach
5. **Client Testing** (1h) - Real-world validation

**Goal**: Working prototype with Client's data by next session.
