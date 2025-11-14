# WeaveImages Metadata Enhancement Plan

## Executive Summary

While the **CRITICAL** `image_data` field issue has been **RESOLVED** ✅, WeaveImages is missing significant metadata compared to RagMeImages. This document outlines the missing metadata and implementation plan.

## Missing Metadata Comparison

### WeaveImages Current Metadata
```json
{
  "type": "image",
  "filename": "Best-Small-Dog-Breeds-Yorkshire-Terrier.jpg",
  "original_filename": "Best-Small-Dog-Breeds-Yorkshire-Terrier.jpg",
  "storage_path": "/Users/maximilien/Desktop/Best-Small-Dog-Breeds-Yorkshire-Terrier.jpg",
  "date_added": "2025-10-15T09:05:48-07:00",
  "file_size": 2406719,
  "chunk_sizes": [2406719],
  "ai_summary": "",
  "content": "",
  "url": "file:///Users/maximilien/Desktop/Best-Small-Dog-Breeds-Yorkshire-Terrier.jpg"
}
```

### RagMeImages Complete Metadata
```json
{
  "type": "image",
  "filename": "Best-Small-Dog-Breeds-Yorkshire-Terrier.jpg",
  "storage_path": "images/20251015_083334_688551_Best-Small-Dog-Breeds-Yorkshire-Terrier.jpg",
  "date_added": "2025-10-15T08:34:28.830879",
  "file_size": 2406719,
  "url": "Best-Small-Dog-Breeds-Yorkshire-Terrier.jpg",
  "source": "file:///var/folders/.../tmp_rueigqj.jpg",
  "processing_timestamp": "2025-10-15T08:34:28.830881",

  // ⚠️ MISSING: EXIF Data
  "exif": {
    "orientation": "Orientation.TOP_LEFT",
    "x_resolution": "72.0",
    "y_resolution": "72.0",
    "resolution_unit": "ResolutionUnit.INCHES",
    "software": "Adobe Photoshop CS6 (Windows)",
    "date_time": "2017:05:17 15:23:42",
    "exif_version": "0221",
    "color_space": "sRGB"
  },

  // ⚠️ MISSING: AI Classification
  "ai_summary": "This is an image file named 'Best-Small-Dog-Breeds-Yorkshire-Terrier.jpg'. AI classification identifies it as 'Yorkshire terrier' with 95% confidence.",
  "classification": {
    "classifications": [
      {
        "confidence": 0.9508275985717773,
        "imagenet_id": 187,
        "label": "Yorkshire terrier",
        "rank": 1
      },
      {
        "confidence": 0.02614,
        "imagenet_id": 248,
        "label": "Eskimo dog",
        "rank": 2
      }
    ]
  },

  // ⚠️ MISSING: OCR Processing
  "ocr_content": {
    "block_count": 0,
    "confidence_threshold": 0.5,
    "engine": "easyocr",
    "extracted_text": "",
    "ocr_processing": true,
    "source_language": "en"
  }
}
```

## Missing Metadata Categories

### 1. EXIF Metadata (MISSING) ⚠️
**What**: Camera/photo metadata embedded in image files
**Examples**:
- Camera model, settings (ISO, aperture, shutter speed)
- GPS coordinates (location where photo was taken)
- Creation date/time
- Image dimensions, resolution
- Software used to create/edit image
- Orientation

**Benefit**: Rich contextual information about image origin

### 2. AI Image Classification (MISSING) ⚠️
**What**: Machine learning-based automatic image labeling
**Examples**:
- Object detection: "Yorkshire terrier" (95% confidence)
- Scene classification: "outdoor", "indoor"
- Multiple classifications with confidence scores
- ImageNet categories

**Benefit**: Searchability without manual tagging

### 3. OCR Text Extraction (MISSING) ⚠️
**What**: Extract text content from images
**Examples**:
- Text in photos (signs, documents, screenshots)
- Block count, confidence threshold
- Source language detection
- EasyOCR or Tesseract engine

**Benefit**: Makes text in images searchable

### 4. Storage Path Strategy (DIFFERENT) ⚠️
**WeaveImages**: Absolute desktop paths `/Users/.../Desktop/file.jpg`
- ❌ Not portable across systems
- ❌ Not cloud-friendly
- ❌ Collision risk with same filenames

**RagMeImages**: Timestamped relative paths `images/YYYYMMDD_HHMMSS_microsec_filename.jpg`
- ✅ Portable across systems
- ✅ Cloud storage friendly
- ✅ No filename collisions
- ✅ Organized by upload time

### 5. Processing Timestamps (MISSING) ⚠️
**What**: Detailed processing metadata
**Examples**:
- `processing_timestamp`: When image was processed
- `extraction_timestamp`: When metadata was extracted
- Processing duration

**Benefit**: Audit trail and troubleshooting

---

---

## Go Library Recommendations

### Proven Libraries for Each Feature

All libraries listed below are production-ready, actively maintained, and widely used in the Go community.

#### 1. EXIF Extraction

**✅ Recommended: `github.com/rwcarlsen/goexif/exif`**
- **Stars**: 2.1k
- **Status**: Mature, stable
- **Pros**: Pure Go, no C dependencies, simple API
- **Cons**: JPEG only (no HEIF/HEIC support)
- **Installation**: `go get github.com/rwcarlsen/goexif/exif`

**Alternative: `github.com/dsoprea/go-exif/v3`**
- **Stars**: 700+
- **Pros**: More comprehensive, supports TIFF/HEIF
- **Cons**: More complex API
- **Installation**: `go get github.com/dsoprea/go-exif/v3`

#### 2. AI Image Classification

**✅ Recommended: `gocv.io/x/gocv` (OpenCV Go bindings)**
- **Stars**: 6.7k
- **Status**: Very mature, active development
- **Pros**: Full OpenCV support, pre-trained model loading (ONNX, Caffe, TensorFlow), high performance
- **Cons**: Requires OpenCV system library
- **Installation**:
  ```bash
  go get -u gocv.io/x/gocv
  # System dependency:
  brew install opencv  # macOS
  apt install libopencv-dev  # Linux
  ```

**Alternative: `github.com/onnx/onnx-go`**
- **Stars**: 700+
- **Pros**: Pure Go ONNX runtime, no C dependencies
- **Cons**: Slower performance than gocv, less mature
- **Installation**: `go get github.com/onnx/onnx-go`

**Cloud API Options:**
- OpenAI Vision: `github.com/sashabaranov/go-openai` (5.3k stars)
- Google Cloud Vision: `cloud.google.com/go/vision/v2/apiv1`
- AWS Rekognition: `github.com/aws/aws-sdk-go-v2`

#### 3. OCR Text Extraction

**✅ Recommended: `github.com/otiai10/gosseract/v2`**
- **Stars**: 2.6k
- **Status**: Mature, actively maintained
- **Pros**: Wraps Tesseract OCR, supports 100+ languages, high accuracy
- **Cons**: Requires Tesseract system library
- **Installation**:
  ```bash
  go get github.com/otiai10/gosseract/v2
  # System dependency:
  brew install tesseract  # macOS
  apt install tesseract-ocr  # Linux
  ```

**Cloud API Options:**
- Google Cloud Vision OCR: `cloud.google.com/go/vision/v2/apiv1`
- AWS Textract: `github.com/aws/aws-sdk-go-v2/service/textract`
- Azure Computer Vision: `github.com/Azure/azure-sdk-for-go`

#### 4. Image Processing (General)

**`github.com/disintegration/imaging`**
- **Stars**: 5.3k
- **Pros**: Pure Go, comprehensive image manipulation (resize, crop, rotate, filters)
- **Cons**: None
- **Installation**: `go get github.com/disintegration/imaging`

---

## Quick Implementation Examples

### EXIF Extraction
```go
package image

import (
    "os"
    "github.com/rwcarlsen/goexif/exif"
)

func ExtractEXIF(imagePath string) (map[string]interface{}, error) {
    file, err := os.Open(imagePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    x, err := exif.Decode(file)
    if err != nil {
        // Not all images have EXIF
        return map[string]interface{}{}, nil
    }

    exifData := make(map[string]interface{})

    // Extract common fields
    if orientation, err := x.Get(exif.Orientation); err == nil {
        exifData["orientation"] = orientation.String()
    }
    if dateTime, err := x.Get(exif.DateTime); err == nil {
        exifData["date_time"] = dateTime.String()
    }
    if make, err := x.Get(exif.Make); err == nil {
        exifData["camera_make"] = make.String()
    }
    if model, err := x.Get(exif.Model); err == nil {
        exifData["camera_model"] = model.String()
    }

    // GPS coordinates
    lat, long, err := x.LatLong()
    if err == nil {
        exifData["gps_latitude"] = lat
        exifData["gps_longitude"] = long
    }

    return exifData, nil
}
```

### OCR Text Extraction
```go
package image

import (
    "strings"
    "github.com/otiai10/gosseract/v2"
)

type OCRResult struct {
    ExtractedText       string  `json:"extracted_text"`
    ConfidenceThreshold float64 `json:"confidence_threshold"`
    Engine              string  `json:"engine"`
    MeanConfidence      float64 `json:"mean_confidence"`
    BlockCount          int     `json:"block_count"`
}

func ExtractTextFromImage(imagePath string, language string) (OCRResult, error) {
    client := gosseract.NewClient()
    defer client.Close()

    client.SetImage(imagePath)
    client.SetLanguage(language) // "eng", "fra", "spa", etc.

    text, err := client.Text()
    if err != nil {
        return OCRResult{}, err
    }

    confidence, _ := client.GetMeanConfidence()

    return OCRResult{
        ExtractedText:       strings.TrimSpace(text),
        ConfidenceThreshold: 0.5,
        Engine:              "tesseract",
        MeanConfidence:      float64(confidence),
        BlockCount:          len(strings.Split(text, "\n\n")),
    }, nil
}
```

### AI Image Classification
```go
package image

import (
    "image"
    "gocv.io/x/gocv"
)

type Classification struct {
    Label      string  `json:"label"`
    Confidence float64 `json:"confidence"`
    ImageNetID int     `json:"imagenet_id"`
    Rank       int     `json:"rank"`
}

func ClassifyImage(imagePath string, modelPath string, labelsPath string) ([]Classification, error) {
    // Load image
    img := gocv.IMRead(imagePath, gocv.IMReadColor)
    if img.Empty() {
        return nil, fmt.Errorf("failed to read image")
    }
    defer img.Close()

    // Preprocess image for model
    blob := gocv.BlobFromImage(img, 1.0/255.0, image.Pt(224, 224),
        gocv.NewScalar(0, 0, 0, 0), true, false)
    defer blob.Close()

    // Load pre-trained model
    net := gocv.ReadNet(modelPath, "")
    if net.Empty() {
        return nil, fmt.Errorf("failed to load model")
    }
    defer net.Close()

    // Run inference
    net.SetInput(blob, "")
    predictions := net.Forward("")
    defer predictions.Close()

    // Parse top-5 predictions
    classifications := parseTopKPredictions(predictions, labelsPath, 5)

    return classifications, nil
}

func parseTopKPredictions(predictions gocv.Mat, labelsPath string, k int) []Classification {
    // Load ImageNet labels
    labels := loadLabels(labelsPath)

    // Get top-k predictions
    // ... implementation details ...

    return classifications
}
```

---

## Dependency Management

### Add to `go.mod`
```go
module github.com/maximilien/weave-cli

go 1.24

require (
    // Existing dependencies...

    // EXIF extraction
    github.com/rwcarlsen/goexif/exif v0.0.0-20201216231939-8d95c6ca79ce

    // OCR text extraction
    github.com/otiai10/gosseract/v2 v2.4.1

    // AI image classification (optional)
    gocv.io/x/gocv v0.35.0

    // Image processing utilities (optional)
    github.com/disintegration/imaging v1.6.2
)
```

### System Dependencies

#### macOS
```bash
# EXIF - no system dependencies (pure Go)

# OCR
brew install tesseract

# AI Classification (if using gocv)
brew install opencv

# Install all at once
brew install tesseract opencv
```

#### Linux (Ubuntu/Debian)
```bash
# EXIF - no system dependencies (pure Go)

# OCR
sudo apt-get update
sudo apt-get install -y tesseract-ocr tesseract-ocr-eng

# AI Classification (if using gocv)
sudo apt-get install -y libopencv-dev

# Install all at once
sudo apt-get install -y tesseract-ocr tesseract-ocr-eng libopencv-dev
```

#### Linux (RedHat/CentOS)
```bash
# OCR
sudo yum install -y tesseract tesseract-langpack-eng

# AI Classification
sudo yum install -y opencv-devel
```

### Model Files (for AI Classification)

Download pre-trained models:
```bash
# Create models directory
mkdir -p models

# Download MobileNetV2 (ImageNet) - ~14MB
wget https://github.com/onnx/models/raw/main/vision/classification/mobilenet/model/mobilenetv2-7.onnx \
     -O models/mobilenet_v2.onnx

# Download ImageNet labels - ~38KB
wget https://raw.githubusercontent.com/onnx/models/main/vision/classification/synset.txt \
     -O models/imagenet_labels.txt
```

---

## Implementation Plan

### Phase 1: EXIF Extraction (HIGH Priority)
**Estimated Time**: 2-3 hours
**Complexity**: Low-Medium
**Library**: `github.com/rwcarlsen/goexif/exif` ✅

#### Implementation Steps:
1. **Add EXIF library dependency**
   ```bash
   go get github.com/rwcarlsen/goexif/exif
   ```

2. **Create EXIF extractor module**
   ```
   src/pkg/image/exif_extractor.go
   ```

3. **Extract EXIF data**
   ```go
   func ExtractEXIF(imagePath string) (map[string]interface{}, error) {
       file, err := os.Open(imagePath)
       if err != nil {
           return nil, err
       }
       defer file.Close()

       x, err := exif.Decode(file)
       if err != nil {
           // Not all images have EXIF
           return map[string]interface{}{}, nil
       }

       exifData := make(map[string]interface{})

       // Extract common fields
       if orientation, err := x.Get(exif.Orientation); err == nil {
           exifData["orientation"] = orientation.String()
       }
       if dateTime, err := x.Get(exif.DateTime); err == nil {
           exifData["date_time"] = dateTime.String()
       }
       if make, err := x.Get(exif.Make); err == nil {
           exifData["camera_make"] = make.String()
       }
       if model, err := x.Get(exif.Model); err == nil {
           exifData["camera_model"] = model.String()
       }

       // GPS coordinates
       lat, long, err := x.LatLong()
       if err == nil {
           exifData["gps_latitude"] = lat
           exifData["gps_longitude"] = long
       }

       return exifData, nil
   }
   ```

4. **Integrate into image processing**
   - Modify `src/cmd/utils/document.go:processImageFile()`
   - Extract EXIF data after reading image file
   - Add to document metadata

5. **Test with various image types**
   - JPEG with EXIF
   - PNG (no EXIF)
   - HEIC/HEIF files

#### Files to Modify:
- `src/pkg/image/exif_extractor.go` (NEW)
- `src/cmd/utils/document.go:562-649`
- `go.mod` (add dependency)

---

### Phase 2: AI Image Classification (MEDIUM Priority)
**Estimated Time**: 4-6 hours
**Complexity**: Medium-High

#### Implementation Options:

**Option A: Local ML Model (Recommended)**
- Use TensorFlow Lite or ONNX Runtime
- Pre-trained ImageNet model
- No external API calls (faster, cheaper)
- Privacy-friendly (no data sent externally)

**Option B: Cloud API**
- OpenAI Vision API
- Google Cloud Vision API
- AWS Rekognition
- Easier but costs money per image

#### Implementation Steps (Option A):
1. **Add ML dependencies**
   ```bash
   # Use gocv (OpenCV Go bindings) with pre-trained model
   go get -u gocv.io/x/gocv
   ```

2. **Download pre-trained model**
   - MobileNetV2 or ResNet50 (ImageNet)
   - Convert to ONNX format if needed
   - Store in `models/` directory

3. **Create classifier module**
   ```
   src/pkg/image/classifier.go
   ```

4. **Implement classification**
   ```go
   func ClassifyImage(imagePath string) ([]Classification, error) {
       // Load image
       img := gocv.IMRead(imagePath, gocv.IMReadColor)
       defer img.Close()

       // Preprocess image
       blob := gocv.BlobFromImage(img, 1.0/255.0, image.Pt(224, 224),
           gocv.NewScalar(0, 0, 0, 0), true, false)
       defer blob.Close()

       // Load model and run inference
       net := gocv.ReadNet("models/mobilenet_v2.onnx", "")
       defer net.Close()

       net.SetInput(blob, "")
       predictions := net.Forward("")
       defer predictions.Close()

       // Parse top-5 predictions
       classifications := parseTopK(predictions, 5)

       return classifications, nil
   }

   type Classification struct {
       Label      string  `json:"label"`
       Confidence float64 `json:"confidence"`
       ImageNetID int     `json:"imagenet_id"`
       Rank       int     `json:"rank"`
   }
   ```

5. **Integrate into image processing**
   - Add `--classify` flag to enable classification
   - Make it optional (can be slow for large batches)
   - Cache results to avoid reprocessing

#### Files to Modify:
- `src/pkg/image/classifier.go` (NEW)
- `src/cmd/utils/document.go:562-649`
- `src/cmd/document/create.go` (add --classify flag)
- `go.mod` (add dependencies)
- `models/` directory (add pre-trained model)

---

### Phase 3: OCR Text Extraction (MEDIUM Priority)
**Estimated Time**: 3-4 hours
**Complexity**: Medium

#### Implementation Options:

**Option A: Tesseract OCR (Recommended)**
- Industry standard, high accuracy
- Good Go bindings: `github.com/otiai10/gosseract/v2`
- Supports 100+ languages
- Free and open source

**Option B: EasyOCR**
- Python-based, need to call via subprocess
- More complex integration
- Good for Asian languages

#### Implementation Steps (Option A):
1. **Install Tesseract system dependency**
   ```bash
   # macOS
   brew install tesseract

   # Linux
   apt-get install tesseract-ocr
   ```

2. **Add Go bindings**
   ```bash
   go get github.com/otiai10/gosseract/v2
   ```

3. **Create OCR module**
   ```
   src/pkg/image/ocr_extractor.go
   ```

4. **Implement OCR extraction**
   ```go
   func ExtractTextFromImage(imagePath string) (OCRResult, error) {
       client := gosseract.NewClient()
       defer client.Close()

       client.SetImage(imagePath)
       client.SetLanguage("eng") // Support multiple languages

       text, err := client.Text()
       if err != nil {
           return OCRResult{}, err
       }

       confidence, _ := client.GetMeanConfidence()

       return OCRResult{
           ExtractedText:       strings.TrimSpace(text),
           ConfidenceThreshold: 0.5,
           Engine:              "tesseract",
           OCRProcessing:       true,
           SourceLanguage:      "eng",
           BlockCount:          countBlocks(text),
           MeanConfidence:      confidence,
       }, nil
   }

   type OCRResult struct {
       ExtractedText       string  `json:"extracted_text"`
       ConfidenceThreshold float64 `json:"confidence_threshold"`
       Engine              string  `json:"engine"`
       OCRProcessing       bool    `json:"ocr_processing"`
       SourceLanguage      string  `json:"source_language"`
       BlockCount          int     `json:"block_count"`
       MeanConfidence      float64 `json:"mean_confidence"`
   }
   ```

5. **Integrate into image processing**
   - Add `--ocr` flag to enable OCR
   - Make it optional (can be slow)
   - Add `--ocr-lang` flag for language selection

6. **Update setup.sh**
   - Add tesseract installation to setup script

#### Files to Modify:
- `src/pkg/image/ocr_extractor.go` (NEW)
- `src/cmd/utils/document.go:562-649`
- `src/cmd/document/create.go` (add --ocr, --ocr-lang flags)
- `setup.sh` (add tesseract installation)
- `go.mod` (add dependency)

---

### Phase 4: Storage Path Strategy (LOW Priority)
**Estimated Time**: 2-3 hours
**Complexity**: Low

#### Implementation Steps:
1. **Create timestamped filename generator**
   ```go
   func GenerateTimestampedFilename(originalFilename string) string {
       now := time.Now()
       microsec := now.Nanosecond() / 1000
       timestamp := now.Format("20060102_150405")

       ext := filepath.Ext(originalFilename)
       baseName := strings.TrimSuffix(filepath.Base(originalFilename), ext)

       return fmt.Sprintf("%s_%06d_%s%s", timestamp, microsec, baseName, ext)
       // Result: 20251015_083334_688551_Best-Small-Dog-Breeds-Yorkshire-Terrier.jpg
   }
   ```

2. **Add storage configuration**
   - Add config option: `storage.use_timestamped_paths: true`
   - Add config option: `storage.relative_paths: true`
   - Add config option: `storage.base_path: "images/"`

3. **Modify image storage**
   - Generate timestamped filename when enabled
   - Use relative paths when configured
   - Update `storage_path` and `url` fields accordingly

#### Files to Modify:
- `src/pkg/storage/path_generator.go` (NEW)
- `src/cmd/utils/document.go:562-649`
- `config.yaml.example` (add storage options)

---

### Phase 5: Processing Timestamps (LOW Priority)
**Estimated Time**: 1 hour
**Complexity**: Low

#### Implementation:
```go
processingStart := time.Now()

// ... do image processing ...

processingEnd := time.Now()

metadata["processing_timestamp"] = processingEnd.Format(time.RFC3339Nano)
metadata["processing_duration_ms"] = processingEnd.Sub(processingStart).Milliseconds()
```

Simple addition to existing code.

---

## Priority Ranking

### Phase 1: EXIF Extraction (HIGH)
- **Time**: 2-3 hours
- **Benefit**: High (rich metadata with minimal effort)
- **Dependencies**: None
- **Risk**: Low

### Phase 4: Storage Path Strategy (HIGH)
- **Time**: 2-3 hours
- **Benefit**: High (portability, cloud-ready)
- **Dependencies**: None
- **Risk**: Low

### Phase 2: AI Classification (MEDIUM)
- **Time**: 4-6 hours
- **Benefit**: High (searchability)
- **Dependencies**: Model files, CV libraries
- **Risk**: Medium (model size, performance)

### Phase 3: OCR Extraction (MEDIUM)
- **Time**: 3-4 hours
- **Benefit**: Medium-High (text searchability)
- **Dependencies**: Tesseract system package
- **Risk**: Low-Medium

### Phase 5: Processing Timestamps (LOW)
- **Time**: 1 hour
- **Benefit**: Low (nice to have)
- **Dependencies**: None
- **Risk**: Low

---

## Recommended Implementation Order

### Week 1: Foundation
1. **EXIF Extraction** (2-3 hours) - Quick win
2. **Storage Path Strategy** (2-3 hours) - Important for portability

### Week 2: Intelligence
3. **OCR Text Extraction** (3-4 hours) - Adds searchability
4. **Processing Timestamps** (1 hour) - Easy addition

### Week 3: Advanced
5. **AI Classification** (4-6 hours) - Most complex, highest value

**Total Estimated Time**: 12-17 hours across 3 weeks

---

## Configuration Additions

### config.yaml
```yaml
image_processing:
  # EXIF extraction
  extract_exif: true

  # AI classification
  classify_images: false  # Optional, slower
  classification_model: "mobilenet_v2"
  classification_top_k: 5

  # OCR processing
  enable_ocr: false  # Optional, slower
  ocr_engine: "tesseract"
  ocr_languages: ["eng"]
  ocr_confidence_threshold: 0.5

  # Storage strategy
  use_timestamped_paths: true
  relative_paths: true
  storage_base_path: "images/"

  # Processing metadata
  add_processing_timestamps: true
```

### CLI Flags
```bash
# EXIF (always enabled by default)
weave docs create MyImages image.jpg --no-exif

# AI classification (opt-in)
weave docs create MyImages image.jpg --classify

# OCR (opt-in)
weave docs create MyImages image.jpg --ocr --ocr-lang eng

# Batch processing with all features
weave docs create MyImages *.jpg --classify --ocr --ocr-lang eng
```

---

## Dependencies Summary

### Go Packages
```
github.com/rwcarlsen/goexif/exif          # EXIF extraction
gocv.io/x/gocv                             # AI classification (OpenCV)
github.com/otiai10/gosseract/v2            # OCR text extraction
```

### System Dependencies
```bash
# macOS
brew install tesseract                     # OCR engine
brew install opencv                        # AI classification

# Linux
apt-get install tesseract-ocr              # OCR engine
apt-get install libopencv-dev              # AI classification
```

### Model Files
```
models/mobilenet_v2.onnx                   # ~14MB - ImageNet classification
models/imagenet_labels.txt                 # ~38KB - Class names
```

---

## Testing Strategy

### Unit Tests
- EXIF extraction with/without EXIF data
- Classification with known images
- OCR with text-heavy images
- Path generation with various filenames

### Integration Tests
- Upload image with all features enabled
- Compare metadata with RagMeImages format
- Verify storage paths are correct
- Test batch uploads

### Performance Tests
- Benchmark EXIF extraction (should be <100ms)
- Benchmark classification (target <2s per image)
- Benchmark OCR (target <3s per image)
- Test batch processing of 100 images

---

## Success Criteria

After implementation, WeaveImages metadata should:
1. ✅ Include EXIF data when available
2. ✅ Include AI classification (when enabled)
3. ✅ Include OCR text extraction (when enabled)
4. ✅ Use timestamped relative storage paths
5. ✅ Include processing timestamps
6. ✅ Match or exceed RagMeImages metadata richness
7. ✅ Maintain backward compatibility
8. ✅ Be configurable (opt-in for expensive operations)

---

## Cost-Benefit Analysis

| Feature | Implementation Time | Value | Maintenance |
|---------|-------------------|-------|-------------|
| EXIF | 2-3 hours | HIGH | LOW |
| Storage Paths | 2-3 hours | HIGH | LOW |
| OCR | 3-4 hours | MEDIUM-HIGH | LOW |
| Timestamps | 1 hour | LOW | NONE |
| AI Classification | 4-6 hours | HIGH | MEDIUM |

**Highest ROI**: EXIF + Storage Paths (4-6 hours for high value)
**Quick Wins**: EXIF, Storage Paths, Timestamps
**Advanced**: AI Classification, OCR

---

## Conclusion

Implementing all phases would bring WeaveImages to feature parity with RagMeImages, with total effort of ~12-17 hours. The recommended approach is to implement Phase 1 (EXIF) and Phase 4 (Storage Paths) first as they provide the highest value-to-effort ratio.
