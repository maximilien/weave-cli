# Daily Plan - Friday, February 13, 2026

**Status**: In Progress
**Focus**: External Storage Integration (Issue #29)
**Goal**: Working prototype with Client's auction images

---

## 🎯 Today's Objectives

### Primary Goal
✅ Complete external storage integration for Milvus
✅ Enable Client to ingest 250+ auction images (100-500KB each)

### Stretch Goal
⚠️ Extend to other VDBs (Weaviate, Qdrant, etc.) for cost optimization

---

## ⏰ Time Allocation

**Total Available**: 6-8 hours
**Critical Path**: 4-5 hours (Phases 1-3)
**Testing**: 1-2 hours (Phase 4)
**Stretch**: 1-2 hours (Phase 5, if time)

---

## 📋 Phase 1: Milvus Adapter Integration (2-3 hours) 🔴 CRITICAL

### 1.1 Update Milvus Adapter Structure
**File**: `src/pkg/vectordb/milvus/adapter.go`

**Task**: Add ImageStorage field to Adapter

```go
type Adapter struct {
    *Client
    llmClient    *llm.OpenAIClient
    imageStorage storage.ImageStorage  // NEW
    pdfStorage   storage.ImageStorage  // NEW (reuse interface)
}

func NewAdapter(config *vectordb.Config) (*Adapter, error) {
    // ... existing code ...

    // NEW: Initialize storage if configured
    var imageStorage, pdfStorage storage.ImageStorage
    if config.ImageStorage != nil && config.ImageStorage.Enabled {
        imageStorage, err = storage.NewImageStorage(storage.Config{
            Type:       storage.StorageType(config.ImageStorage.Type),
            Endpoint:   config.ImageStorage.Endpoint,
            AccessKey:  config.ImageStorage.AccessKey,
            SecretKey:  config.ImageStorage.SecretKey,
            Region:     config.ImageStorage.Region,
            Bucket:     config.ImageStorage.Bucket,
            PathPrefix: config.ImageStorage.PathPrefix,
            UseSSL:     config.ImageStorage.UseSSL,
        })
        if err != nil {
            return nil, fmt.Errorf("failed to create image storage: %w", err)
        }
    }

    // Similar for PDF storage...

    return &Adapter{
        Client:       client,
        llmClient:    llmClient,
        imageStorage: imageStorage,
        pdfStorage:   pdfStorage,
    }, nil
}
```

**Time**: 30 min
**Checkpoint**: `go build` succeeds

---

### 1.2 Update CreateDocument for External Storage
**File**: `src/pkg/vectordb/milvus/document.go`

**Task**: Add external storage logic before existing image handling

```go
func (a *Adapter) CreateDocument(ctx context.Context, collectionName string, document *vectordb.Document) error {
    logger := logging.WithDocument("milvus", collectionName, document.ID)
    logger.Debug("Creating document")

    ctx, cancel := context.WithTimeout(ctx, a.getTimeout())
    defer cancel()

    // NEW: Handle external storage for large images
    if a.imageStorage != nil && document.ImageData != "" {
        imageBytes := []byte(document.ImageData)

        if storage.IsThumbnailNeeded(imageBytes) {
            logger.Info("Image exceeds Milvus limit, using external storage")

            // 1. Upload full image to S3/MinIO
            imageURL, err := a.imageStorage.Upload(ctx, imageBytes, storage.ImageMetadata{
                Filename:       document.Image,
                ContentType:    detectContentType(imageBytes),
                Size:           int64(len(imageBytes)),
                CollectionName: collectionName,
                DocumentID:     document.ID,
            })
            if err != nil {
                return fmt.Errorf("failed to upload image to storage: %w", err)
            }

            // 2. Generate safe thumbnail for Milvus
            thumbnail, width, height, err := storage.GenerateSafeThumbnail(imageBytes)
            if err != nil {
                return fmt.Errorf("failed to generate thumbnail: %w", err)
            }

            // 3. Update document with thumbnail + URL
            document.ImageThumbnail = thumbnail
            document.ImageURL = imageURL
            document.ImageMetadata = map[string]interface{}{
                "original_size": len(imageBytes),
                "thumbnail_size": len(thumbnail),
                "width": width,
                "height": height,
                "storage_backend": a.imageStorage.Type,
            }

            // 4. Clear ImageData to save space (already in S3)
            document.ImageData = ""

            logger.Info("Image uploaded to external storage",
                "url", imageURL,
                "thumbnail_size", len(thumbnail))
        }
    }

    // Continue with existing logic...
    mdoc := a.toMilvusDocument(document)

    // ... rest of existing CreateDocument code ...
}
```

**Helper function**:
```go
func detectContentType(data []byte) string {
    // Simple detection based on magic bytes
    if len(data) < 4 {
        return "application/octet-stream"
    }

    switch {
    case bytes.HasPrefix(data, []byte{0xFF, 0xD8, 0xFF}):
        return "image/jpeg"
    case bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47}):
        return "image/png"
    case bytes.HasPrefix(data, []byte{0x47, 0x49, 0x46}):
        return "image/gif"
    default:
        return "application/octet-stream"
    }
}
```

**Time**: 1 hour
**Checkpoint**: Compiles, logic in place

---

### 1.3 Update CreateDocuments (Batch)
**File**: `src/pkg/vectordb/milvus/document.go`

**Task**: Add same logic to batch creation

```go
func (a *Adapter) CreateDocuments(ctx context.Context, collectionName string, documents []*vectordb.Document) error {
    // ... existing setup ...

    for i, doc := range documents {
        // NEW: Handle external storage per document
        if a.imageStorage != nil && doc.ImageData != "" {
            imageBytes := []byte(doc.ImageData)

            if storage.IsThumbnailNeeded(imageBytes) {
                // Same logic as CreateDocument...
                imageURL, err := a.imageStorage.Upload(ctx, imageBytes, ...)
                thumbnail, _, _, err := storage.GenerateSafeThumbnail(imageBytes)

                doc.ImageThumbnail = thumbnail
                doc.ImageURL = imageURL
                doc.ImageMetadata = map[string]interface{}{...}
                doc.ImageData = ""
            }
        }

        // Continue with existing batch logic...
        mdoc := a.toMilvusDocument(doc)
        // ... existing code ...
    }
}
```

**Time**: 30 min
**Checkpoint**: Batch ingestion updated

---

### 1.4 Update Milvus Schema
**File**: `src/pkg/vectordb/milvus/collection.go`

**Task**: Add new fields to collection schema

```go
func (c *Client) CreateCollection(ctx context.Context, name string, schema *vectordb.CollectionSchema) error {
    // ... existing code ...

    // Add new fields for external storage
    fields := []*entity.Field{
        // ... existing fields ...

        // NEW: External storage fields
        {
            Name:     FieldImageThumbnail,
            DataType: entity.FieldTypeVarChar,
            TypeParams: map[string]string{
                "max_length": "65535", // Max safe for Milvus
            },
        },
        {
            Name:     FieldImageURL,
            DataType: entity.FieldTypeVarChar,
            TypeParams: map[string]string{
                "max_length": "512",
            },
        },
        {
            Name:     FieldImageMetadata,
            DataType: entity.FieldTypeJSON,
        },
    }

    // ... rest of existing code ...
}
```

**Add field constants**:
```go
const (
    // ... existing constants ...
    FieldImageThumbnail = "image_thumbnail"
    FieldImageURL       = "image_url"
    FieldImageMetadata  = "image_metadata"
)
```

**Time**: 20 min
**Checkpoint**: New collections have external storage fields

---

**Phase 1 Total Time**: 2-3 hours
**Phase 1 Checkpoint**: ✅ Milvus adapter compiles and has external storage logic

---

## 📋 Phase 2: Configuration & CLI Flags (1-2 hours) 🟡 HIGH PRIORITY

### 2.1 Add Storage Config
**File**: `src/pkg/vectordb/config.go`

```go
type Config struct {
    // ... existing fields ...

    // External storage configuration
    ImageStorage *StorageConfig `yaml:"image_storage,omitempty"`
    PDFStorage   *StorageConfig `yaml:"pdf_storage,omitempty"`
}

type StorageConfig struct {
    Enabled    bool   `yaml:"enabled"`
    Type       string `yaml:"type"` // "s3", "minio", "local"
    Endpoint   string `yaml:"endpoint,omitempty"`
    AccessKey  string `yaml:"access_key,omitempty"`
    SecretKey  string `yaml:"secret_key,omitempty"`
    Region     string `yaml:"region,omitempty"`
    Bucket     string `yaml:"bucket"`
    PathPrefix string `yaml:"path_prefix,omitempty"`
    UseSSL     bool   `yaml:"use_ssl"`
}
```

**Time**: 15 min

---

### 2.2 Add CLI Flags to docs create
**File**: `src/cmd/document/create.go`

```go
var (
    // ... existing flags ...

    // Image storage flags
    imageStorageType   string
    imageStorageBucket string
    s3Region           string
    s3AccessKey        string
    s3SecretKey        string
    minioEndpoint      string
    minioAccessKey     string
    minioSecretKey     string

    // PDF storage flags
    storePDF          bool
    pdfStorageType    string
    pdfStorageBucket  string
)

func init() {
    // ... existing flags ...

    // Image storage
    createCmd.Flags().StringVar(&imageStorageType, "image-storage", "", "Image storage type (s3, minio, local)")
    createCmd.Flags().StringVar(&imageStorageBucket, "image-bucket", "", "Storage bucket for images")

    // S3
    createCmd.Flags().StringVar(&s3Region, "s3-region", "us-east-1", "AWS S3 region")
    createCmd.Flags().StringVar(&s3AccessKey, "s3-access-key", "", "AWS access key (or use AWS_ACCESS_KEY_ID env)")
    createCmd.Flags().StringVar(&s3SecretKey, "s3-secret-key", "", "AWS secret key (or use AWS_SECRET_ACCESS_KEY env)")

    // MinIO
    createCmd.Flags().StringVar(&minioEndpoint, "minio-endpoint", "localhost:9000", "MinIO endpoint")
    createCmd.Flags().StringVar(&minioAccessKey, "minio-access-key", "minioadmin", "MinIO access key")
    createCmd.Flags().StringVar(&minioSecretKey, "minio-secret-key", "minioadmin", "MinIO secret key")

    // PDF storage
    createCmd.Flags().BoolVar(&storePDF, "store-pdf", false, "Store original PDF in external storage")
    createCmd.Flags().StringVar(&pdfStorageType, "pdf-storage", "", "PDF storage type (s3, minio, local)")
    createCmd.Flags().StringVar(&pdfStorageBucket, "pdf-bucket", "", "Storage bucket for PDFs")
}
```

**Time**: 30 min

---

### 2.3 Wire Flags to Config
**File**: `src/cmd/document/create.go`

```go
func runCreate(cmd *cobra.Command, args []string) error {
    // ... existing code ...

    // Build storage config from flags
    var imageStorageConfig *vectordb.StorageConfig
    if imageStorageType != "" {
        imageStorageConfig = &vectordb.StorageConfig{
            Enabled:    true,
            Type:       imageStorageType,
            Bucket:     imageStorageBucket,
            UseSSL:     imageStorageType == "s3",
        }

        switch imageStorageType {
        case "s3":
            imageStorageConfig.Region = s3Region
            imageStorageConfig.AccessKey = getEnvOrFlag(s3AccessKey, "AWS_ACCESS_KEY_ID")
            imageStorageConfig.SecretKey = getEnvOrFlag(s3SecretKey, "AWS_SECRET_ACCESS_KEY")
        case "minio":
            imageStorageConfig.Endpoint = minioEndpoint
            imageStorageConfig.AccessKey = minioAccessKey
            imageStorageConfig.SecretKey = minioSecretKey
            imageStorageConfig.UseSSL = false
        case "local":
            imageStorageConfig.PathPrefix = "./storage/images"
        }
    }

    // Add to VDB config
    vdbConfig.ImageStorage = imageStorageConfig

    // Similar for PDF storage...

    // Continue with existing create logic...
}

func getEnvOrFlag(flag, envVar string) string {
    if flag != "" {
        return flag
    }
    return os.Getenv(envVar)
}
```

**Time**: 30 min

---

**Phase 2 Total Time**: 1-2 hours
**Phase 2 Checkpoint**: ✅ CLI flags work end-to-end

---

## 📋 Phase 3: MinIO Local Setup (30 min) 🟢 QUICK WIN

### 3.1 Docker Compose
**File**: `docker-compose.minio.yml`

```yaml
version: '3.8'

services:
  minio:
    image: minio/minio:latest
    container_name: weave-minio
    ports:
      - "9000:9000"      # S3 API
      - "9001:9001"      # Web Console
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

  # Auto-create buckets
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
      mc anonymous set download local/weave-images;
      mc anonymous set download local/weave-pdfs;
      echo 'MinIO buckets ready: weave-images, weave-pdfs';
      "

volumes:
  minio-data:
    driver: local
```

**Time**: 15 min

---

### 3.2 Quick Start Script
**File**: `tools/start-minio.sh`

```bash
#!/bin/bash
set -e

echo "🚀 Starting MinIO for Weave CLI..."

# Start MinIO
docker-compose -f docker-compose.minio.yml up -d

# Wait for health
echo "⏳ Waiting for MinIO to be ready..."
sleep 5

# Verify
if curl -f http://localhost:9000/minio/health/live > /dev/null 2>&1; then
    echo "✅ MinIO is ready!"
    echo ""
    echo "📊 MinIO Console: http://localhost:9001"
    echo "   Username: minioadmin"
    echo "   Password: minioadmin"
    echo ""
    echo "🪣 Buckets created:"
    echo "   - weave-images (for images)"
    echo "   - weave-pdfs (for PDFs)"
    echo ""
    echo "🎯 Usage:"
    echo "   weave docs create Collection file.pdf \\"
    echo "     --include-images \\"
    echo "     --image-storage minio \\"
    echo "     --image-bucket weave-images \\"
    echo "     --milvus-local"
else
    echo "❌ MinIO failed to start"
    exit 1
fi
```

**Time**: 10 min

---

### 3.3 Update README
**File**: `README.md`

Add MinIO quick start section (after external storage was covered earlier).

**Time**: 5 min

---

**Phase 3 Total Time**: 30 min
**Phase 3 Checkpoint**: ✅ MinIO running locally with buckets

---

## 📋 Phase 4: Testing & Validation (1-2 hours) 🔵 CRITICAL

### 4.1 Unit Tests
**File**: `src/pkg/storage/minio_test.go`

Quick smoke tests:
```go
func TestMinioStorage_UploadDownload(t *testing.T) {
    // Test upload + download roundtrip
}

func TestThumbnail_SafeGeneration(t *testing.T) {
    // Verify thumbnail fits in 65KB
}
```

**Time**: 30 min

---

### 4.2 Integration Test
**File**: `test-external-storage.sh`

```bash
#!/bin/bash
set -e

echo "🧪 Testing External Storage Integration"

# 1. Start MinIO
./tools/start-minio.sh

# 2. Create test image (large, >47KB)
# Use existing auction catalog or create test image

# 3. Test single image
echo "Testing single image..."
weave docs create TestImages test-large-image.jpg \
  --milvus-local \
  --image-storage minio \
  --image-bucket weave-images \
  --minio-endpoint localhost:9000

# 4. Verify
weave cols show TestImages --milvus-local | grep "image_url"

# 5. Test PDF with images
echo "Testing PDF with images..."
weave docs create TestPDF test-catalog.pdf \
  --milvus-local \
  --include-images \
  --image-storage minio \
  --image-bucket weave-images

# 6. Query and verify results
weave cols query TestPDF "test" --top-k 1 --milvus-local

echo "✅ Tests passed!"
```

**Time**: 30 min

---

### 4.3 Client Data Test (CRITICAL)
**Manual testing with Client's auction catalog**

```bash
# Use Client's actual PDF (250+ images, 100-500KB each)
weave docs create AuctionImages_External auction-catalog.pdf \
  --milvus-local \
  --include-images \
  --image-storage minio \
  --image-bucket weave-images \
  --minio-endpoint localhost:9000

# Verify:
# 1. All 250+ images uploaded to MinIO
# 2. Thumbnails stored in Milvus
# 3. URLs returned in queries
# 4. Can download full images from MinIO URLs
```

**Time**: 30 min
**Checkpoint**: ✅ Client's 250+ images work!

---

**Phase 4 Total Time**: 1-2 hours
**Phase 4 Checkpoint**: ✅ End-to-end working with real data

---

## 🎯 Success Criteria for Tonight

### Minimum Viable (MUST HAVE)
- ✅ Milvus adapter uses external storage
- ✅ CLI flags work (--image-storage minio)
- ✅ MinIO docker-compose working
- ✅ Single large image (>47KB) ingests successfully
- ✅ Thumbnail stored in Milvus, URL accessible

### Target (SHOULD HAVE)
- ✅ Client's auction catalog (250+ images) works
- ✅ Batch ingestion tested
- ✅ Query returns thumbnails + URLs
- ✅ Can download full images from MinIO

### Stretch (NICE TO HAVE)
- ⚠️ PDF storage working
- ⚠️ S3 production setup documented
- ⚠️ Extended to other VDBs (Weaviate, Qdrant)

---

## 📋 Phase 5: Polish & Extend (1-2 hours) ⚪ STRETCH

### If Time Permits

#### 5.1 PDF Storage
Add `--store-pdf` support (similar to image storage)

**Time**: 1 hour

#### 5.2 Extend to Other VDBs
Make storage optional for Weaviate, Qdrant (cost optimization)

**Time**: 1 hour

#### 5.3 Documentation
- Update EXTERNAL_STORAGE_IMPLEMENTATION.md with actual vs planned
- Create user guide
- Add troubleshooting section

**Time**: 30 min

---

## 🚦 Decision Points

### Tonight (11pm PST cutoff)
**Can complete Phases 1-3**: ✅ Ship it!
**Cannot complete Phase 1**: ⏸️ Tag for tomorrow

### Tomorrow
**If tagged**: Complete Phases 1-4 (4-6 hours)
**If partially done**: Continue from checkpoint

---

## 📝 Commits Strategy

### Tonight
1. `feat: integrate external storage with Milvus adapter` (Phase 1)
2. `feat: add CLI flags for external storage configuration` (Phase 2)
3. `feat: add MinIO docker-compose for local development` (Phase 3)
4. `test: add external storage integration tests` (Phase 4)

### Tomorrow (if needed)
5. `feat: add PDF storage support`
6. `feat: extend external storage to all VDBs`
7. `docs: add external storage user guide`

---

## 🎯 Final Deliverable

**Working Command**:
```bash
# Client can run this and it WORKS
weave docs create AuctionImages auction-catalog.pdf \
  --milvus-local \
  --include-images \
  --image-storage minio \
  --image-bucket weave-images \
  --minio-endpoint localhost:9000

# Result: 250+ images (100-500KB each) successfully ingested
# - Thumbnails in Milvus (<47KB)
# - Full images in MinIO
# - URLs accessible
# - Queries work
```

---

## ⏰ Time Tracking

| Phase | Planned | Actual | Status |
|-------|---------|--------|--------|
| Phase 1 | 2-3h | ___ | ⏳ |
| Phase 2 | 1-2h | ___ | ⏳ |
| Phase 3 | 30m  | ___ | ⏳ |
| Phase 4 | 1-2h | ___ | ⏳ |
| Phase 5 | 1-2h | ___ | ⏳ Stretch |
| **Total** | **5-10h** | ___ | ⏳ |

---

## 🚀 Let's Go!

**Start Time**: ___
**Target**: Working prototype with Client's data
**Fallback**: Tag incomplete phases for tomorrow

**Ready? LFG! 🎯**
