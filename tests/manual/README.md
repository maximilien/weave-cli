# Manual Test Collections

Scripts and instructions for manually testing Weave CLI features.

---

## Multi-Modal RAG Image Collections

**Script**: `test-image-collections.sh`

**Purpose**: Create test collections with image documents for manual testing of
multi-modal RAG support (Phase 1).

### Prerequisites

1. **Build Weave CLI**:

   ```bash
   cd /path/to/weave-cli
   ./build.sh
   ```

2. **Set up VDB credentials** (at least one required):

   ```bash
   # Weaviate Cloud
   export WEAVIATE_CLOUD_API_KEY="your-api-key"
   export WEAVIATE_CLOUD_URL="https://your-cluster.weaviate.network"

   # MongoDB Atlas
   export MONGODB_ATLAS_CONNECTION_STRING="mongodb+srv://..."

   # Or use local Weaviate (requires Docker/Podman)
   # No env vars needed - uses http://localhost:8080
   ```

### Usage

```bash
# Run from weave-cli root directory
./tests/manual/test-image-collections.sh
```

### What It Creates

#### 1. TestVintageCars (Weaviate Cloud)

Image collection with 3 vintage car photos:

- **car-001**: 1967 Ford Mustang
  - OCR: "1967 Ford Mustang"
  - Description: "Vintage red Mustang in excellent condition"
  - Tags: vintage, car, mustang, 1967, classic

- **car-002**: 1969 Chevrolet Camaro SS
  - OCR: "1969 Chevrolet Camaro SS"
  - Description: "Blue Camaro with original paint and chrome details"
  - Tags: vintage, car, camaro, 1969, muscle-car

- **car-003**: 1970 Dodge Challenger
  - OCR: "1970 Dodge Challenger"
  - Description: "Orange Dodge Challenger R/T with black stripes"
  - Tags: vintage, car, challenger, 1970, mopar

#### 2. TestCarDocs + TestVintageCars (Mixed Collections)

- **TestCarDocs**: Text collection with car history/specs
- **TestVintageCars**: Image collection with car photos

**Purpose**: Test multi-modal queries combining text and image sources.

#### 3. TestImageMetadata (Weaviate Local)

Tests various metadata levels:

- **full-001**: Complete metadata (OCR, description, tags)
- **ocr-only-001**: Only OCR text
- **desc-only-001**: Only description
- **no-meta-001**: No metadata (just image URL)

**Purpose**: Verify context builder handles all metadata scenarios.

#### 4. Cross-VDB Collections

- **TestVintageCars**: Weaviate Cloud
- **TestImagesDB**: MongoDB Cloud

**Purpose**: Test querying same collection type across different VDBs.

---

## Manual Test Scenarios

### Scenario 1: Basic Image Query

**Goal**: Verify images return results with rag-agent

```bash
./bin/weave cols query TestVintageCars "vintage muscle cars" \
  --agent rag-agent --top_k 3 --weaviate-cloud
```

**Expected Output**:

```text
[1] TestVintageCars (weaviate-cloud) - Score: 87.3%
Text in image: 1967 Ford Mustang
Description: Vintage red Mustang in excellent condition
Tags: vintage, car, mustang, 1967, classic
Image URL: https://cdn.example.com/1967-mustang.jpg

[2] TestVintageCars (weaviate-cloud) - Score: 82.1%
Text in image: 1969 Chevrolet Camaro SS
...
```

✅ **Success Criteria**: Results show OCR text, descriptions, tags, and image URLs

---

### Scenario 2: Mixed Text + Image Collections

**Goal**: Verify text and image results are properly mixed by score

```bash
./bin/weave cols query TestCarDocs TestVintageCars \
  "1967 Mustang specifications" \
  --agent rag-agent --top_k 5 --weaviate-cloud
```

**Expected Output**: Mix of text documents and image results, sorted by relevance

✅ **Success Criteria**:

- Results include both text (from TestCarDocs) and images (from TestVintageCars)
- Sorted by score regardless of source type
- Image results show metadata (OCR, description, etc.)

---

### Scenario 3: Image Content Extraction Levels

**Goal**: Verify all metadata levels handled correctly

```bash
./bin/weave cols query TestImageMetadata "test" \
  --agent rag-agent --top_k 5 --weaviate-local
```

**Expected Output**:

```text
[1] TestImageMetadata - Score: X%
Text in image: Complete OCR text from image
Description: Full description with all details
Tags: tag1, tag2, tag3
Image URL: https://cdn.example.com/full-metadata.jpg

[2] TestImageMetadata - Score: X%
Text in image: Only OCR text available
Image URL: https://cdn.example.com/ocr-only.jpg

[3] TestImageMetadata - Score: X%
Description: Only description provided
Image URL: https://cdn.example.com/desc-only.jpg

[4] TestImageMetadata - Score: X%
Image URL: https://cdn.example.com/no-metadata.jpg
```

✅ **Success Criteria**:

- Full metadata: Shows all fields
- OCR-only: Shows OCR + URL
- Description-only: Shows description + URL
- No metadata: Shows URL only
- NO "[Empty document]" for any image

---

### Scenario 4: Cross-VDB Image Query

**Goal**: Verify images work across multiple VDBs

```bash
./bin/weave cols query \
  TestVintageCars:weaviate-cloud \
  TestImagesDB:mongodb-cloud \
  "vintage cars" --agent rag-agent --top_k 5
```

**Expected Output**: Results from both Weaviate and MongoDB image collections

✅ **Success Criteria**:

- Results from both VDBs
- Citations show correct VDB names (weaviate-cloud, mongodb-cloud)
- Images from both sources properly formatted

---

### Scenario 5: Verify Document Structure

**Goal**: Inspect raw document to verify image fields

```bash
./bin/weave docs get TestVintageCars car-001 --weaviate-cloud
```

**Expected Output**:

```json
{
  "id": "car-001",
  "image": "https://cdn.example.com/1967-mustang.jpg",
  "metadata": {
    "ocr_text": "1967 Ford Mustang",
    "description": "Vintage red Mustang in excellent condition",
    "tags": ["vintage", "car", "mustang", "1967", "classic"],
    "_collection": "TestVintageCars",
    "_vdb": "weaviate-cloud"
  }
}
```

✅ **Success Criteria**:

- `image` field present
- Metadata includes ocr_text, description, tags
- Collection and VDB metadata auto-added

---

## Cleanup

When done testing, remove test collections:

```bash
# Weaviate Cloud
./bin/weave cols delete TestVintageCars --weaviate-cloud
./bin/weave cols delete TestCarDocs --weaviate-cloud

# Weaviate Local
./bin/weave cols delete TestImageMetadata --weaviate-local

# MongoDB Cloud
./bin/weave cols delete TestImagesDB --mongodb-cloud
```

---

## Troubleshooting

### No Results from Image Collection

**Symptom**: Query returns 0 results from image collection

**Check**:

1. Verify collection exists:

   ```bash
   ./bin/weave cols list --weaviate-cloud
   ```

2. Verify documents exist:

   ```bash
   ./bin/weave docs get TestVintageCars car-001 --weaviate-cloud
   ```

3. Check if image field is present in document

4. Try query without agent first:

   ```bash
   ./bin/weave cols query TestVintageCars "vintage" --top_k 3 --weaviate-cloud
   ```

### Images Show as "[Empty document]"

**This was the bug!** If you see this, the Phase 1 fix is not working.

**Verify**:

1. Check you're using the latest build:

   ```bash
   ./bin/weave --version
   ```

2. Rebuild:

   ```bash
   ./build.sh
   ```

3. Verify context_builder.go has extractImageContent() function

### Mixed Collections Return Only Text

**Check**:

1. Verify image collection has data:

   ```bash
   ./bin/weave cols list --weaviate-cloud
   ```

2. Query each collection separately to isolate issue

3. Check scores - text might just be ranking higher

---

## Expected Test Results

After running `./tests/manual/test-image-collections.sh`, you should be able to:

✅ Query image-only collections with rag-agent
✅ Query mixed text + image collections
✅ See OCR text, descriptions, and tags in results
✅ See image URLs in citations
✅ Get results from images with various metadata levels
✅ Query images across multiple VDBs

❌ Should NOT see:

- "[Empty document]" for images
- Zero results from image collections
- Missing metadata (OCR, descriptions) in output
- Images filtered out when mixed with text

---

## Phase 2 Preview

Once Phase 2 is implemented, you should see:

- 🖼️ Emoji prefix for image results
- Clickable markdown links for image URLs
- Better visual distinction between text and image sources

Example:

```text
[1] 🖼️ Image - TestVintageCars (weaviate-cloud) - Score: 87.3%
[View Image](https://cdn.example.com/1967-mustang.jpg)

Text in image: 1967 Ford Mustang
Description: Vintage red Mustang in excellent condition
Tags: vintage, car, mustang, 1967, classic
```

---

**Last Updated**: 2026-01-16 (Phase 1)
