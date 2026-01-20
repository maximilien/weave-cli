# Multi-Modal RAG Support

> **✅ COMPLETED** - Client requirement for multi-modal RAG use case
>
> **📚 Related Documents**: [Planning Index](README.md)

**Status**: ✅ Complete (v0.9.4) - Ready for Production
**Issue**: ✅ RESOLVED - Image collections now work with rag-agent
**Impact**: Unblocks client deployment of multi-modal RAG to AuctionsMax.ai

---

## Problem Statement

### Current Issue

**User Report**: "They need to be able to query image collections using the
rag-agent (and other agents). Currently doing a multi-collection query where
one collection is an image collection is not getting any results."

**Example Query**:
```bash
weave cols query \
  ProductDocs:mongodb-cloud \
  ProductImages:weaviate-cloud \
  "vintage car" \
  --agent rag-agent \
  --top_k 3
```

**Expected**: Results from both text docs AND images
**Actual**: Only text docs, zero image results

---

## Root Cause Analysis

### Investigation

**File**: `src/pkg/agents/context_builder.go:217-229`

```go
// extractContent extracts the primary content from a document
func (cb *ContextBuilder) extractContent(doc vectordb.Document) string {
    // Priority: Content > Text > URL
    if doc.Content != "" {
        return doc.Content  // ✅ Text documents work
    }
    if doc.Text != "" {
        return doc.Text     // ✅ Text documents work
    }
    if doc.URL != "" {
        return fmt.Sprintf("[Document URL: %s]", doc.URL)  // ✅ URLs work
    }
    return "[Empty document]"  // ❌ Images treated as empty!
}
```

**Missing Fields**:
- ❌ `doc.Image` - Image URL/path
- ❌ `doc.ImageData` - Base64 encoded image data
- ❌ Image metadata (OCR text, descriptions, EXIF data)

**Result**: Image documents are treated as "empty" and filtered out or
shown as `[Empty document]`.

### Why This Happens

1. **Image Documents Structure**:
   ```json
   {
     "id": "auction-item-001",
     "content": "",       // Empty for images
     "text": "",          // Empty for images
     "image": "https://cdn.example.com/vintage-car.jpg",
     "image_data": "base64...",
     "metadata": {
       "ocr_text": "1967 Ford Mustang",
       "description": "Vintage red Mustang in excellent condition",
       "width": 1920,
       "height": 1080,
       "file_format": "JPEG",
       "_collection": "AuctionsImages",
       "_vdb": "weaviate-cloud"
     }
   }
   ```

2. **Current Logic**:
   - Checks `content` → empty
   - Checks `text` → empty
   - Checks `url` → empty
   - Returns `[Empty document]`
   - RAG agent filters it out or shows nothing

3. **What Should Happen**:
   - Check `image` field → has URL
   - Extract OCR text from `metadata.ocr_text`
   - Extract description from `metadata.description`
   - Build meaningful context for LLM

---

## Existing Infrastructure

### What We Already Have

✅ **Document Model** (`src/pkg/vectordb/interfaces.go:11-19`):
```go
type Document struct {
    ID        string
    Text      string
    Content   string
    Image     string                 // ✅ Image URL/path
    ImageData string                 // ✅ Base64 image data
    URL       string
    Metadata  map[string]interface{} // ✅ Can hold OCR, descriptions
}
```

✅ **Schema Types** (`src/pkg/vectordb/interfaces.go:62-64`):
```go
const (
    SchemaTypeText  SchemaType = "text"
    SchemaTypeImage SchemaType = "image"  // ✅ Image collection type exists
)
```

✅ **Image Processing** (Found in grep results):
- `src/pkg/image/ocr.go` - OCR text extraction
- `src/pkg/image/exif_extractor.go` - EXIF metadata extraction
- `src/pkg/pdf/image_extractor.go` - PDF image extraction

✅ **Image Data Handling** (Found in multiple VDBs):
- Milvus: `image_data` field support
- Weaviate: `image` field support
- MongoDB: Image metadata support
- OpenSearch/Elasticsearch: Image field support

---

## Solution Approaches

### Approach 1: Immediate Fix (Quick Win)

**Goal**: Make image collections work with rag-agent TODAY

**Changes**: Update `context_builder.go:extractContent()` to handle images

**Implementation**:
```go
// extractContent extracts the primary content from a document
func (cb *ContextBuilder) extractContent(doc vectordb.Document) string {
    // Priority: Content > Text > Image Metadata > Image URL

    // 1. Text content (existing)
    if doc.Content != "" {
        return doc.Content
    }
    if doc.Text != "" {
        return doc.Text
    }

    // 2. Image metadata (NEW)
    if doc.Image != "" || doc.ImageData != "" {
        return cb.extractImageContent(doc)
    }

    // 3. URL fallback (existing)
    if doc.URL != "" {
        return fmt.Sprintf("[Document URL: %s]", doc.URL)
    }

    return "[Empty document]"
}

// extractImageContent extracts content from image documents (NEW)
func (cb *ContextBuilder) extractImageContent(doc vectordb.Document) string {
    var parts []string

    // Check for OCR text in metadata
    if ocrText, ok := doc.Metadata["ocr_text"].(string); ok && ocrText != "" {
        parts = append(parts, fmt.Sprintf("Text in image: %s", ocrText))
    }

    // Check for description
    if desc, ok := doc.Metadata["description"].(string); ok && desc != "" {
        parts = append(parts, fmt.Sprintf("Description: %s", desc))
    }

    // Check for alt text
    if alt, ok := doc.Metadata["alt_text"].(string); ok && alt != "" {
        parts = append(parts, fmt.Sprintf("Alt text: %s", alt))
    }

    // Check for tags
    if tags, ok := doc.Metadata["tags"].([]interface{}); ok && len(tags) > 0 {
        tagStrs := make([]string, 0, len(tags))
        for _, tag := range tags {
            if tagStr, ok := tag.(string); ok {
                tagStrs = append(tagStrs, tagStr)
            }
        }
        if len(tagStrs) > 0 {
            parts = append(parts,
              fmt.Sprintf("Tags: %s", strings.Join(tagStrs, ", ")))
        }
    }

    // Include image URL
    if doc.Image != "" {
        parts = append(parts, fmt.Sprintf("Image URL: %s", doc.Image))
    }

    // If we have content, join it
    if len(parts) > 0 {
        return strings.Join(parts, "\n")
    }

    // Fallback: just indicate it's an image
    if doc.Image != "" {
        return fmt.Sprintf("[Image: %s]", doc.Image)
    }
    if doc.ImageData != "" {
        return "[Image data (base64)]"
    }

    return "[Image document without metadata]"
}
```

**Effort**: 2-3 hours
**Impact**: Immediate - AuctionsMax.ai unblocked
**Risk**: Low - additive change, no breaking changes

**Testing**:
```bash
# Test 1: Query image collection only
weave cols query AuctionsImages:weaviate-cloud "vintage car" \
  --agent rag-agent --top_k 3

# Expected: Image results with OCR text, descriptions

# Test 2: Query mixed text + image collections
weave cols query \
  AuctionsDocs:mongodb-cloud \
  AuctionsImages:weaviate-cloud \
  "vintage car" \
  --agent rag-agent --top_k 3

# Expected: Mix of text docs and image results
```

---

### Approach 2: Enhanced Image Citations

**Goal**: Show image results clearly in agent output

**Changes**: Update `rag_agent.go` to format image citations differently

**Implementation**:
```go
// In rag_agent.go - formatMarkdown()
func (ra *RAGAgent) formatMarkdown(context *QueryContext) string {
    for _, source := range context.Sources {
        // Detect if this is an image document
        isImage := source.Metadata["_is_image"] == true ||
                   source.DocID != "" && hasImageExtension(source.DocID)

        if isImage {
            // Format image citation
            builder.WriteString(fmt.Sprintf("**[%d] 🖼️ Image**", source.Index))

            // Add image URL if available
            if imageURL, ok := source.Metadata["image_url"].(string); ok {
                builder.WriteString(fmt.Sprintf(" - [View Image](%s)", imageURL))
            }

            // Add collection and VDB
            builder.WriteString(fmt.Sprintf(" - %s (%s)",
              collectionName, vdbName))

            // Add score
            builder.WriteString(fmt.Sprintf(" - Score: %.1f%%",
              source.Score*100))

            // Content (OCR, description, etc.)
            builder.WriteString(fmt.Sprintf("\n\n%s\n\n", source.Content))
        } else {
            // Normal text citation (existing code)
            // ...
        }
    }
}
```

**Example Output**:
```markdown
**[1] 🖼️ Image** - [View Image](https://cdn.../car.jpg) -
  AuctionsImages (weaviate-cloud) - Score: 87.3%

Text in image: 1967 Ford Mustang
Description: Vintage red Mustang in excellent condition
Tags: vintage, car, mustang, 1967

**[2]** AuctionsDocs (mongodb-cloud) - Score: 82.1%

Ford Mustang history and specifications...
```

**Effort**: 3-4 hours
**Impact**: High - better UX for image results
**Risk**: Low - additive formatting

---

### Approach 3: Multi-Modal Embeddings (Long-term)

**Goal**: Semantic image search with CLIP embeddings

**Use Case**: Find similar images visually, not just by text metadata

**Changes**:
- Integrate CLIP model for image embeddings
- Support visual similarity search
- Combine text + visual embeddings (hybrid search)

**Example**:
```bash
# Query with image input (future)
weave cols query AuctionsImages "vintage cars" \
  --image query-image.jpg \
  --agent multimodal-rag-agent

# System:
# 1. Generate CLIP embedding for query-image.jpg
# 2. Search for visually similar images
# 3. Also search text metadata for "vintage cars"
# 4. Combine results (hybrid)
```

**Implementation**:
- CLIP embedding generation (OpenAI CLIP or similar)
- VDB support for image vector search (most VDBs support this)
- Multi-modal agent that handles both text and image queries

**Effort**: 1-2 weeks
**Impact**: Game-changer for image-heavy applications
**Risk**: Medium - requires ML model integration

---

## Recommended Implementation Plan

### Phase 1: Emergency Fix ✅ COMPLETE (v0.9.4)

**Goal**: Unblock client deployment immediately

**Tasks**:
1. ✅ Identify root cause (found in `schema.go` - wrong model name)
2. ✅ Fix `GetDefaultSchema()` to use correct embedding model
3. ✅ Fix schema type detection in adapter
4. ✅ Test with production image collections across VDBs
5. ✅ Commit and tag as release (v0.9.4)

**Deliverables**:
- ✅ Updated `schema.go` with correct embedding model
- ✅ Updated `collections.go` with schema type detection
- ✅ Integration tests across all VDBs (Milvus, Weaviate, Chroma, Qdrant)
- ✅ CLI workflow tests for end-to-end verification
- ✅ Citation verification tests
- ✅ Release v0.9.4 tagged and documented

**Acceptance Criteria - MET**:
```bash
# This query returns BOTH text and image results
weave cols query \
  MilvusTextCol MilvusImageCol \
  "weave cli" \
  --agent rag-agent --top_k 5 --top_k_images 2 --milvus-cloud

# Output includes:
# ✅ 5 text documents from MilvusTextCol
# ✅ 2 image results from MilvusImageCol with descriptions
# ✅ RAG agent cites both collections
# ✅ Mix is sorted by relevance score
```

**Bugs Fixed**:
1. ✅ `schema.go` - Changed "text2vec-openai" to "text-embedding-3-small"
2. ✅ `collections.go` - Added schema type detection from properties
3. ✅ `client_collections.go` - Added debug logging for troubleshooting
4. ✅ Integration tests - Fixed UUID requirements
5. ✅ Integration tests - Simplified image document structure

---

### Phase 2: Enhanced Image Support ✅ COMPLETE (v0.9.4)

**Goal**: Better image result formatting and metadata handling

**Tasks**:
1. ✅ Image citations in RAG agent output (collection names shown)
2. ✅ Detection logic distinguishes text vs image collections
3. ✅ `--top_k_images` flag for guaranteed image results
4. ✅ Multi-VDB support (Milvus, Weaviate, Chroma, Qdrant)
5. ✅ Document image collection workflows and testing

**Deliverables**:
- ✅ RAG agent cites image collections with collection names
- ✅ `--top_k_images` flag implementation
- ✅ Name-based and schema-based collection type detection
- ✅ Comprehensive documentation in `tests/integration/README.md`
- ✅ Multi-VDB auto-detection in tests

**Additional Enhancements**:
- ✅ Verbose mode shows debug info (collection type, topK values)
- ✅ Citation format includes collection names for both text and images
- ✅ Integration tests verify citations work across all VDBs

---

### Phase 3: Multi-Modal Embeddings (Future - 2-3 weeks)

**Goal**: Full multi-modal RAG with visual similarity search

**Tasks**:
1. ⬜ Research CLIP integration options (OpenAI, Hugging Face)
2. ⬜ Add CLIP embedding support to VDB adapters
3. ⬜ Create `multimodal-rag-agent.yaml` config
4. ⬜ Support image query input (`--image` flag)
5. ⬜ Hybrid search (text + visual similarity)
6. ⬜ Image captioning for better context

**Deliverables**:
- CLIP embedding integration
- Visual similarity search
- Multi-modal agent
- Image query support

---

## Technical Details

### Image Metadata Standards

**Recommended Metadata Fields**:
```json
{
  "id": "unique-id",
  "image": "https://cdn.example.com/image.jpg",
  "image_data": "base64...",  // Optional: for embedding in VDB
  "metadata": {
    // Text extracted from image
    "ocr_text": "Text found in image via OCR",

    // Human-provided descriptions
    "description": "Human-readable description",
    "alt_text": "Accessibility alt text",
    "caption": "Image caption",

    // Visual metadata
    "tags": ["vintage", "car", "red", "1967"],
    "categories": ["vehicles", "classic-cars"],

    // Technical metadata (EXIF)
    "width": 1920,
    "height": 1080,
    "file_format": "JPEG",
    "file_size": 245678,
    "created_at": "2024-01-15T10:30:00Z",

    // Collection info (auto-added by weave)
    "_collection": "ProductImages",
    "_vdb": "weaviate-cloud",
    "_is_image": true
  }
}
```

### VDB Support Matrix

| VDB | Image Field | Image Data | Visual Search | CLIP Embeddings |
|-----|-------------|------------|---------------|-----------------|
| Weaviate | ✅ | ✅ | ✅ | ✅ (img2vec-neural) |
| Milvus | ✅ | ✅ | ✅ | ⚠️ (manual) |
| Qdrant | ✅ | ✅ | ✅ | ⚠️ (manual) |
| MongoDB | ✅ | ✅ | ⚠️ | ⚠️ (manual) |
| Pinecone | ✅ | ✅ | ✅ | ⚠️ (manual) |
| Chroma | ✅ | ✅ | ✅ | ⚠️ (manual) |

**Legend**:
- ✅ Full support
- ⚠️ Requires manual integration
- ❌ Not supported

---

## Real-World Use Cases

### Use Case 1: Mixed Text + Image Search

**Query**: "vintage cars from 1960s"

**Collections**:
- `ProductDocs` - Text descriptions, history, specifications
- `ProductImages` - Product photos with OCR and metadata

**Expected Results**:
```
[1] ProductDocs (mongodb-cloud) - Score: 89.2%
1967 Ford Mustang - Classic American muscle car from the golden age...

[2] 🖼️ Image - ProductImages (weaviate-cloud) - Score: 87.3%
Text in image: 1967 Ford Mustang
Description: Vintage red Mustang in excellent condition
Tags: vintage, car, mustang, 1967, classic
[View Image](https://cdn.example.com/items/car-001.jpg)

[3] ProductDocs (mongodb-cloud) - Score: 85.1%
1969 Chevrolet Camaro - Iconic muscle car with powerful V8...

[4] 🖼️ Image - ProductImages (weaviate-cloud) - Score: 82.7%
Text in image: 1969 Camaro SS
Description: Blue Camaro with original paint
Tags: vintage, car, camaro, 1969, muscle-car
[View Image](https://cdn.example.com/items/car-002.jpg)
```

### Use Case 2: Image-Only Query

**Query**: "show me furniture with ornate carvings"

**Collection**: `FurnitureImages` (image collection with OCR + descriptions)

**Expected Results**:
```
[1] 🖼️ Image - FurnitureImages (weaviate-cloud) - Score: 91.2%
Description: Victorian-era chair with intricate floral carvings
Tags: furniture, chair, victorian, carved, ornate
[View Image](https://cdn.example.com/furniture/chair-001.jpg)

[2] 🖼️ Image - FurnitureImages (weaviate-cloud) - Score: 88.5%
Description: Antique dresser with hand-carved details
Tags: furniture, dresser, antique, carved, ornate
[View Image](https://cdn.example.com/furniture/dresser-001.jpg)
```

---

## Testing Strategy

### Unit Tests

```go
// src/pkg/agents/context_builder_image_test.go

func TestExtractContent_ImageWithOCR(t *testing.T) {
    cb := NewContextBuilder(config)

    doc := vectordb.Document{
        ID: "img-001",
        Image: "https://example.com/image.jpg",
        Metadata: map[string]interface{}{
            "ocr_text": "1967 Ford Mustang",
            "description": "Vintage red Mustang",
            "tags": []interface{}{"vintage", "car"},
        },
    }

    content := cb.extractContent(doc)

    assert.Contains(t, content, "1967 Ford Mustang")
    assert.Contains(t, content, "Vintage red Mustang")
    assert.Contains(t, content, "vintage, car")
    assert.Contains(t, content, "https://example.com/image.jpg")
}

func TestExtractContent_ImageWithoutMetadata(t *testing.T) {
    cb := NewContextBuilder(config)

    doc := vectordb.Document{
        ID: "img-002",
        Image: "https://example.com/image2.jpg",
        Metadata: map[string]interface{}{},
    }

    content := cb.extractContent(doc)

    assert.Contains(t, content, "https://example.com/image2.jpg")
    assert.NotContains(t, content, "[Empty document]")
}
```

### Integration Tests

```bash
# Test 1: Create image collection
weave cols create ProductImages \
  --schema-type image \
  --weaviate-cloud

# Test 2: Add image with metadata
weave docs create ProductImages \
  --image https://example.com/car.jpg \
  --metadata '{"ocr_text":"1967 Mustang","description":"Vintage car"}' \
  --weaviate-cloud

# Test 3: Query with rag-agent
weave cols query ProductImages "vintage car" \
  --agent rag-agent \
  --top_k 3 \
  --weaviate-cloud

# Expected: Returns image with OCR text and description
```

---

## Risks and Mitigations

### Risk 1: Performance with Large Images

**Risk**: Base64 image data in responses could be very large

**Mitigation**:
- Don't include `image_data` in agent context (only `image` URL)
- Add `--include-image-data` flag for cases where needed
- Truncate large metadata fields

### Risk 2: Missing Metadata

**Risk**: Images without OCR or descriptions show as empty

**Mitigation**:
- Always show image URL as fallback
- Add default description: "[Image document without metadata]"
- Recommend OCR pipeline in documentation

### Risk 3: Mixed Result Quality

**Risk**: Text and image results mixed poorly in rankings

**Mitigation**:
- Fair score normalization across modalities
- Optional `--prefer-images` or `--prefer-text` flags
- Document collection ordering best practices

---

## Documentation Updates

### User Guide

**New Section**: "Working with Image Collections"

```markdown
## Querying Image Collections

Weave CLI supports multi-modal queries across text and image collections.

### Image Document Structure

Image documents should include:
- `image`: URL or path to image
- `metadata.ocr_text`: Text extracted from image via OCR
- `metadata.description`: Human-readable description
- `metadata.tags`: Searchable tags

### Example: Create Image Collection

\`\`\`bash
weave cols create ProductImages --schema-type image
\`\`\`

### Example: Add Image with Metadata

\`\`\`bash
weave docs create ProductImages \
  --image https://cdn.example.com/product.jpg \
  --metadata '{
    "ocr_text": "Product label text",
    "description": "Blue widget with chrome finish",
    "tags": ["widget", "blue", "chrome"]
  }'
\`\`\`

### Example: Query Mixed Text + Image Collections

\`\`\`bash
weave cols query \
  ProductDocs:mongodb-cloud \
  ProductImages:weaviate-cloud \
  "blue widgets" \
  --agent rag-agent \
  --top_k 5
\`\`\`
```

---

## Success Criteria

### Phase 1 (Emergency Fix)

- [x] AuctionsMax.ai can query image collections with rag-agent
- [x] Image results show OCR text and descriptions
- [x] Mixed text + image queries work correctly
- [x] No performance degradation

### Phase 2 (Enhanced Support)

- [ ] Image citations clearly distinguished from text
- [ ] Image URLs clickable in markdown output
- [ ] Documentation complete and tested

### Phase 3 (Multi-Modal)

- [ ] Visual similarity search working
- [ ] CLIP embeddings integrated
- [ ] Image query input supported
- [ ] Hybrid search (text + visual) functional

---

## ✅ Completed Actions (v0.9.4)

### Phase 1 & 2 Complete

1. **Emergency Fix Deployed**:
   - ✅ Fixed `schema.go` - correct embedding model
   - ✅ Fixed `collections.go` - schema type detection
   - ✅ Added debug logging for troubleshooting
   - ✅ Tested across all VDBs

2. **Production Testing Complete**:
   - ✅ Created test collections (text + image)
   - ✅ Queried with rag-agent
   - ✅ Verified results include both text and images
   - ✅ Verified `--top_k_images` works as expected

3. **Release v0.9.4**:
   - ✅ Committed changes (4 commits)
   - ✅ Tagged as v0.9.4
   - ✅ Updated CHANGELOG
   - ⏳ Ready to push and notify client

### Integration Testing Complete

1. ✅ API-level tests (`top_k_images_test.go`)
2. ✅ CLI workflow tests (`top_k_images_cli_test.go`)
3. ✅ Citation verification (`verify_citations_test.go`)
4. ✅ Documentation (`tests/integration/README.md`)

## Next Steps - Phase 3 (Future)

### Pending Client Feedback

**Immediate** (After AuctionsMax.ai Deployment):
1. Monitor production usage and performance
2. Gather feedback on `--top_k_images` feature
3. Document any edge cases discovered

**Short-term** (1-2 weeks):
1. Research CLIP integration options (OpenAI, Hugging Face)
2. Plan visual similarity search implementation
3. Design multi-modal agent architecture

**Long-term** (2-3 weeks):
1. Implement CLIP embeddings for visual search
2. Add image query input support (`--image` flag)
3. Build hybrid search (text + visual similarity)

---

## Questions for Client Implementation

1. **Metadata Fields**: What metadata is included with images?
   - OCR text?
   - Descriptions?
   - Tags/categories?
   - EXIF data?

2. **Image Storage**: Where are images hosted?
   - CDN URLs?
   - Base64 in VDB?
   - Local paths?

3. **Search Priorities**: What's most important?
   - Text in images (OCR)?
   - Image descriptions?
   - Visual similarity (future)?

4. **Collection Structure**: How many image collections?
   - One per category?
   - Mixed with text collections?

5. **Volume**: How many images?
   - Impacts indexing and search performance

---

**Last Updated**: 2026-01-16
**Status**: Planning - Ready for Implementation
**Priority**: 🚨 BLOCKER - Immediate Action Required
