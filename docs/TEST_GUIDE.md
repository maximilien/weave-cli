# Weave CLI Test Guide

This guide documents how to run and extend tests for weave-cli.

## Test Organization

### Unit Tests
Location: `src/pkg/*/`
Run: `go test ./src/...`

### Integration Tests
Location: `tests/`
Run: `go test -tags=integration ./tests`

### E2E Tests
Location: Root
Run: `./e2e-tests.sh`

## Image Ingestion Tests (Issue #21)

### Purpose
Systematically test image ingestion across all supported vector databases to identify size limits, schema issues, and compatibility problems.

### Test File
`tests/image_ingestion_test.go`

### What's Tested

#### TestImageIngestion_AllVDBs
Tests image ingestion for each VDB:
- ✅ Weaviate Local - No size limit
- ⚠️  Milvus Local - 64KB JSON field limit (known issue #19)
- ✅ Chroma Local
- ✅ Qdrant Local
- ✅ Neo4j Local
- ✅ OpenSearch Local
- ✅ Supabase
- ✅ MongoDB

**For each database:**
1. Creates test collection with image schema
2. Ingests small test images (<64KB)
3. Verifies collection count matches ingested count
4. Handles size limits appropriately
5. Cleans up after test

**Running:**
```bash
# All VDBs
go test -v -tags=integration ./tests -run TestImageIngestion_AllVDBs

# Specific VDB (e.g., Weaviate)
go test -v -tags=integration ./tests -run "TestImageIngestion_AllVDBs/Weaviate"

# Skip short mode to run full tests
go test -v -tags=integration ./tests -run TestImageIngestion_AllVDBs -short=false
```

#### TestImageIngestion_CMYK
Tests CMYK color space image handling.

**Status:** Placeholder (test data not yet available)

**To enable:**
1. Add CMYK test image to `tests/images/cmyk_image.png`
2. Remove skip in test

#### TestImageIngestion_LargeImages
Tests handling of large images (>64KB) to validate size limits.

**Status:** Placeholder (large test image not yet available)

**To enable:**
1. Add 81KB test image to `tests/images/large_81kb.png`
2. Remove skip in test

**Expected results:**
- Weaviate: Should handle without error
- Milvus: Should error due to 64KB limit

### Test Data

**Existing:**
- `tests/images/dog.png`
- `tests/images/test_*.png` (colored test images)

**Needed:**
- `tests/images/large_81kb.png` - For size limit testing
- `tests/images/cmyk_image.png` - For color space testing

### Success Criteria

For each VDB, tests verify:
- [ ] Collections create successfully
- [ ] All images within size limit ingest correctly
- [ ] Document counts match expected
- [ ] No silent failures or truncation
- [ ] Proper error handling for oversized images

### Known Issues

**Milvus 64KB Limit (Issue #19)**
- Milvus has a 64KB limit on JSON fields
- Images larger than ~64KB will fail to ingest
- Test skips large images for Milvus
- Marked as known limitation in test

## PDF Extraction Tests (Issue #8)

### Purpose
Validate PDF text and image extraction across different PDF versions (1.3 to 2.0) and content types (text-only, scanned, mixed, photo-heavy).

### Test File
`tests/pdf_integration_test.go`

### What's Tested

#### TestPDFVersionCompatibility
Tests extraction from PDFs of different versions:

| Version | Year | Features | Status |
|---------|------|----------|--------|
| PDF 1.3 | 1999 | Basic features | Placeholder |
| PDF 1.4 | 2001 | CMYK, transparency | Placeholder |
| PDF 1.7 | 2006 | Modern features | Placeholder |
| PDF 2.0 | 2017 | Latest spec | Placeholder |
| ragme-io.pdf | Current | Working baseline | ✅ Active |

**For each version:**
1. Extracts text and images
2. Verifies extraction succeeded
3. Validates metadata fields
4. Logs extraction statistics

**Running:**
```bash
# All PDF versions
go test -v ./tests -run TestPDFVersionCompatibility

# Specific version
go test -v ./tests -run "TestPDFVersionCompatibility/PDF_1.7"
```

#### TestPDFTypeExtraction
Tests extraction from different PDF content types:

| Type | Text Expected | Images Expected | Status |
|------|---------------|-----------------|--------|
| Text-only | ✅ | ❌ | Placeholder |
| Scanned (OCR) | ❌ | ✅ | Placeholder |
| Mixed | ✅ | ✅ | Placeholder |
| Photo-heavy | ✅ | ✅ (5+) | Placeholder |
| Auction Catalogue | ✅ | ✅ | ✅ Active |

**For each type:**
1. Extracts with images enabled
2. Verifies expected content types present
3. Validates minimum chunk/image counts
4. Logs warnings for unexpected results

**Running:**
```bash
# All PDF types
go test -v ./tests -run TestPDFTypeExtraction

# Specific type
go test -v ./tests -run "TestPDFTypeExtraction/Mixed"
```

#### TestPDFWithCMYKImages
Tests handling of PDFs with CMYK color space images.

**Status:** Placeholder (test PDF not yet available)

**To enable:**
1. Add CMYK PDF to `fixtures/pdf_types/cmyk.pdf`
2. Remove skip in test

### Test Data

**Existing:**
- `fixtures/ragme-io.pdf` - Working baseline (mixed content)

**Needed:**
- `fixtures/pdf_versions/pdf_1.3.pdf`
- `fixtures/pdf_versions/pdf_1.4.pdf`
- `fixtures/pdf_versions/pdf_1.7.pdf`
- `fixtures/pdf_versions/pdf_2.0.pdf`
- `fixtures/pdf_types/text_only.pdf`
- `fixtures/pdf_types/scanned.pdf`
- `fixtures/pdf_types/mixed.pdf`
- `fixtures/pdf_types/photo_heavy.pdf`

### Success Criteria

Tests verify:
- [ ] All PDF versions extract correctly
- [ ] Text-only PDFs extract text, no images
- [ ] Scanned PDFs detected (no text without OCR)
- [ ] Mixed PDFs extract both text and images
- [ ] Photo-heavy PDFs extract all images
- [ ] Metadata fields preserved across versions
- [ ] CMYK images handled correctly

## Creating Test PDFs

### Quick Test PDFs

**Text-only PDF:**
```bash
# macOS: Use Pages or TextEdit, export as PDF
# Linux: Use LibreOffice
echo "Test content" | ps2pdf - fixtures/pdf_types/text_only.pdf
```

**Image-only PDF (Scanned simulation):**
```bash
# Convert image to PDF
convert tests/images/dog.png fixtures/pdf_types/scanned.pdf
```

**Mixed PDF:**
Use any office suite to create a document with text and embedded images, then export to PDF.

**Photo-heavy PDF:**
```bash
# Combine multiple images into one PDF
convert tests/images/*.png fixtures/pdf_types/photo_heavy.pdf
```

### Version-specific PDFs

For testing specific PDF versions, use PDF creation tools that support version selection:
- **Adobe Acrobat** - Can save as specific PDF version
- **LibreOffice** - PDF export options include version
- **Ghostscript** - Can convert to specific versions

Example with Ghostscript:
```bash
gs -sDEVICE=pdfwrite -dCompatibilityLevel=1.4 \
   -o fixtures/pdf_versions/pdf_1.4.pdf input.pdf
```

## Running Full Test Suite

### Quick Tests (Short Mode)
```bash
# Run only fast tests
go test ./tests -short
```

### Integration Tests
```bash
# Run all integration tests
go test -v -tags=integration ./tests

# Run specific integration test
go test -v -tags=integration ./tests -run TestImageIngestion
```

### E2E Tests
```bash
# Full end-to-end test suite
./e2e-tests.sh

# With specific VDB
./e2e-tests.sh --weaviate-local
```

### CI/CD Tests
```bash
# Same as CI runs
./test.sh
```

## Test Coverage

### View Coverage Report
```bash
# Generate coverage
go test -coverprofile=coverage.out ./src/...

# View in browser
go tool cover -html=coverage.out
```

### Current Coverage
See `tools/test-coverage.sh` for detailed coverage analysis.

## Adding New Tests

### For New VDB Support

1. **Add to image_ingestion_test.go:**
```go
{
    name:    "MyNewVDB Local",
    vdbType: vectordb.VectorDBTypeMyNewVDBLocal,
    sizeLimit: 0, // or specific limit if known
},
```

2. **Add integration test file:**
Create `tests/mynewvdb_integration_test.go` following existing patterns.

3. **Update test.sh:**
Add VDB-specific test flags and logic.

### For New PDF Features

1. **Add test case to TestPDFTypeExtraction:**
```go
{
    name:           "My New PDF Type",
    pdfFile:        "fixtures/pdf_types/my_new_type.pdf",
    expectedText:   true,
    expectedImages: true,
    minTextChunks:  1,
    minImages:      1,
},
```

2. **Create test PDF:**
Add to `fixtures/pdf_types/`

3. **Update documentation:**
Add to this guide and relevant docs.

## Troubleshooting Tests

### OCR Tests Fail (gosseract)
```
tessbridge.cpp:5:10: fatal error: 'leptonica/allheaders.h' file not found
```

**Solution:** Install Tesseract dependencies
```bash
# macOS
brew install tesseract leptonica

# Ubuntu/Debian
sudo apt-get install tesseract-ocr libtesseract-dev libleptonica-dev
```

### VDB Connection Failures

**Problem:** Tests fail to connect to VDB

**Solutions:**
1. Ensure VDB is running: `./tools/vdb/local/<vdb>.sh start`
2. Check config: `weave config show`
3. Verify health: `weave health check --<vdb>-local`

### Test Data Not Found

**Problem:** Tests skip due to missing files

**Solution:** Check skip reasons and add required test data:
```bash
# See what's needed
go test -v ./tests -run TestImage 2>&1 | grep "Skip"
```

## Best Practices

1. **Always clean up:** Tests should delete collections after completion
2. **Use meaningful names:** Test names should describe what's being tested
3. **Log useful info:** Use `t.Logf()` for debugging information
4. **Skip gracefully:** Use `t.Skip()` with clear reasons for missing data
5. **Verify expectations:** Assert both success and failure cases
6. **Document assumptions:** Comment why specific values are expected

## Related Documentation

- [VDB Support Matrix](VDB_SUPPORT.md)
- [User Guide](USER_GUIDE.md)
- [Contributing Guide](../CONTRIBUTING.md)
- [Issue #21: Image Ingestion Testing](https://github.com/maximilien/weave-cli/issues/21)
- [Issue #8: PDF Version Testing](https://github.com/maximilien/weave-cli/issues/8)
