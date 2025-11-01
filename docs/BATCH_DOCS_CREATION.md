# Batch Document Creation

## Overview

The `weave docs batch` command enables efficient batch processing of multiple documents from a directory. It supports parallel processing, automatic retry on failures, progress tracking, and comprehensive reporting.

## Features

### Core Functionality

1. **Batch Processing**: Process all supported files in a directory automatically
2. **Parallel Processing**: Configure multiple workers with `--parallel` flag
3. **Smart Retry**: Automatic retry on failures with configurable attempts
4. **Progress Tracking**: Track already-processed files using `.processed` metadata files
5. **Visual Progress**: Emojis and colored output showing real-time progress
6. **Time Estimation**: Estimated time remaining based on processing speed
7. **Comprehensive Reporting**: CSV reports with detailed processing statistics

### Supported File Types

- **Text files**: `.txt`, `.md`, `.json`, `.yaml`, `.yml`
- **PDF files**: `.pdf` (with text extraction and image extraction)
- **Image files**: `.jpg`, `.jpeg`, `.png`, `.gif`, `.bmp`, `.webp`

## Command Syntax

```bash
weave docs batch --directory <dir> --collection <name> [options]
```

### Required Flags

- `--directory`, `-d`: Directory containing documents to process
- `--collection`, `-c`: Collection name for documents

### Optional Flags

#### Processing Options
- `--parallel`, `-p`: Number of parallel workers (default: 1)
- `--retry`: Number of retry attempts for failed documents (default: 2)
- `--chunk-size`, `-s`: Chunk size for text content in characters (default: 5000)
- `--force-reprocess`: Reprocess all files, ignore `.processed` files (default: false)

#### PDF-Specific Options
- `--image-collection`: Collection name for extracted PDF images (default: same as main collection)
- `--skip-small-images`: Skip small images when extracting from PDFs (default: true)
- `--min-image-size`: Minimum image size in bytes (default: 5120 = 5KB)
- `--batch-size`: Number of images to process before pausing for memory cleanup (default: 10)

#### Reporting Options
- `--create-report`: Create a new CSV report of processing results (default: batch-report.csv)

## Usage Examples

### Basic Usage

Process all documents in a directory:

```bash
weave docs batch --directory ./documents --collection MyDocs
```

### Parallel Processing

Use 5 parallel workers for faster processing:

```bash
weave docs batch --directory ./documents --collection MyDocs --parallel 5
```

### With Custom PDF Options

Process PDFs with custom image extraction settings:

```bash
weave docs batch \
  --directory ./pdfs \
  --collection MyPDFs \
  --image-collection MyImages \
  --min-image-size 10240 \
  --batch-size 5 \
  --parallel 3
```

### With Reporting

Generate a detailed CSV report:

```bash
weave docs batch \
  --directory ./documents \
  --collection MyDocs \
  --create-report processing-report.csv \
  --parallel 3
```

### Force Reprocess

Reprocess all files, even if already processed:

```bash
weave docs batch \
  --directory ./documents \
  --collection MyDocs \
  --force-reprocess
```

## How It Works

### 1. Directory Scanning

The command scans the specified directory recursively for all supported file types:

```
./documents/
├── file1.txt        ✓ Supported
├── file2.pdf        ✓ Supported
├── image1.jpg       ✓ Supported
├── data.csv         ✗ Not supported
└── subdirectory/
    └── file3.md     ✓ Supported
```

### 2. Processing State Tracking

For each file processed, a `.processed` metadata file is created:

```json
{
  "file_path": "/path/to/document.txt",
  "processed_at": "2025-10-31T10:30:00Z",
  "success": true,
  "text_chunks": 5,
  "images": 0,
  "processing_time_ms": 1234,
  "file_size": 25000,
  "retry_count": 0,
  "chunks_failed": 0,
  "images_failed": 0
}
```

This allows the batch processor to:
- Skip successfully processed files on subsequent runs
- Retry failed files automatically
- Track processing statistics per file

### 3. Parallel Processing

Workers process files concurrently based on `--parallel` setting:

```
Worker 1: file1.txt → file4.pdf → file7.jpg
Worker 2: file2.md  → file5.txt → file8.png
Worker 3: file3.pdf → file6.md  → file9.txt
```

### 4. Progress Display

Real-time progress with visual indicators:

```
Batch Processing Configuration:
  Directory:        ./documents
  Collection:       MyDocs
  Parallel workers: 3
  Retry attempts:   2
  Chunk size:       5000 chars

Found 15 file(s) to process

Processing 15 file(s)...

📝 ✓ [1/15] document1.txt (3 chunks) - 1.2s - ETA: 16s
📄 ✓ [2/15] report.pdf (10 chunks, 5 images) - 3.5s - ETA: 45s
🖼️  ✓ [3/15] photo.jpg - 0.8s - ETA: 9s
❌ ✗ [4/15] corrupted.pdf - Failed to read file (after 2 retries)
📝 ✓ [5/15] notes.md (2 chunks) - 0.5s - ETA: 5s
...
```

### 5. Summary Report

After processing, a summary is displayed:

```
============================================================
Batch Processing Summary
============================================================
Total files:      15
Successful:       13
Failed:           2
Text chunks:      125
Images extracted: 23
Total time:       1m 45s
Avg time/file:    7s

Failed files:
  - corrupted.pdf: Failed to read file
  - invalid.txt: Permission denied
============================================================
```

### 6. CSV Report

If `--create-report` is specified, a CSV file is generated:

```csv
Timestamp,File,Status,TextChunks,Images,ProcessingTimeMs,FileSize,RetryCount,ChunksFailed,ImagesFailed,Error
2025-10-31 10:30:15,document1.txt,success,3,0,1234,5000,0,0,0,
2025-10-31 10:30:18,report.pdf,success,10,5,3456,150000,1,0,0,
2025-10-31 10:30:19,photo.jpg,success,0,1,890,25000,0,0,0,
2025-10-31 10:30:22,corrupted.pdf,failed,0,0,567,0,2,0,0,Failed to read file
```

## Processing States

### Success State
- File processed completely
- All chunks/images created successfully
- `.processed` file created with `success: true`
- File skipped on subsequent batch runs (unless `--force-reprocess`)

### Partial Success State
- File processed but some chunks/images failed
- `.processed` file created with failure counts
- Considered successful overall but with warnings

### Failure State
- File failed to process after all retry attempts
- `.processed` file created with `success: false`
- File will be retried on subsequent batch runs

### Skipped State
- File has valid `.processed` file with `success: true`
- Automatically skipped to save time
- Can override with `--force-reprocess`

## Best Practices

### 1. Start with Small Parallel Count

Begin with 1-3 workers and increase based on system performance:

```bash
# Start conservative
weave docs batch --directory ./docs --collection MyDocs --parallel 3

# Increase if system can handle it
weave docs batch --directory ./docs --collection MyDocs --parallel 10
```

### 2. Use Appropriate Retry Count

- **Default (2)**: Good for network hiccups
- **Higher (5+)**: For unstable connections
- **Lower (1)**: For fast failure detection

```bash
# High retry for unreliable networks
weave docs batch --directory ./docs --collection MyDocs --retry 5
```

### 3. Monitor Resource Usage

For large batches with PDFs:

```bash
# Reduce batch size to prevent memory issues
weave docs batch \
  --directory ./pdfs \
  --collection Docs \
  --batch-size 5 \
  --parallel 2
```

### 4. Always Use Reports for Large Batches

Track processing details for auditing:

```bash
weave docs batch \
  --directory ./documents \
  --collection MyDocs \
  --create-report "$(date +%Y%m%d)-batch-report.csv"
```

### 5. Handle Failed Files

Review failed files in the report and retry with adjusted settings:

```bash
# First run
weave docs batch --directory ./docs --collection MyDocs

# Review failures, then reprocess failed files
weave docs batch \
  --directory ./docs \
  --collection MyDocs \
  --retry 5 \
  --parallel 1  # Slower but more reliable
```

## Advanced Scenarios

### Large PDF Collections

For directories with many large PDFs:

```bash
weave docs batch \
  --directory ./pdf-library \
  --collection PDFDocs \
  --image-collection PDFImages \
  --parallel 2 \
  --batch-size 5 \
  --min-image-size 10240 \
  --chunk-size 3000 \
  --retry 3 \
  --create-report pdf-batch-report.csv
```

**Rationale**:
- Low parallel count (2) prevents memory exhaustion
- Small batch size (5) for image processing
- Higher min image size (10KB) filters tiny images
- Smaller chunks (3000) for better granularity
- More retries (3) for large file reliability

### Mixed Content Directory

For directories with text, images, and PDFs:

```bash
weave docs batch \
  --directory ./mixed-content \
  --collection AllDocs \
  --image-collection AllImages \
  --parallel 5 \
  --chunk-size 5000 \
  --create-report mixed-batch-report.csv
```

### Incremental Processing

Regular batch runs will automatically skip processed files:

```bash
# Daily batch job
weave docs batch \
  --directory /data/daily-docs \
  --collection DailyDocs \
  --parallel 5 \
  --create-report "/reports/$(date +%Y%m%d)-batch.csv"
```

## Troubleshooting

### Issue: High Memory Usage

**Solution**: Reduce parallel workers and batch size

```bash
weave docs batch \
  --directory ./docs \
  --collection MyDocs \
  --parallel 1 \
  --batch-size 3
```

### Issue: Slow Processing

**Solution**: Increase parallel workers if system has capacity

```bash
weave docs batch \
  --directory ./docs \
  --collection MyDocs \
  --parallel 10
```

### Issue: Many Failed Files

**Solution**: Check error messages in report, increase retry count

```bash
weave docs batch \
  --directory ./docs \
  --collection MyDocs \
  --retry 5 \
  --parallel 1  # Single worker for debugging
```

### Issue: Files Not Being Skipped

**Solution**: Check for `.processed` files or use `--force-reprocess`

```bash
# Remove all .processed files to reprocess everything
find ./docs -name "*.processed" -delete

# Or use force-reprocess flag
weave docs batch \
  --directory ./docs \
  --collection MyDocs \
  --force-reprocess
```

### Issue: Time Estimation Inaccurate

**Note**: Time estimates improve as more files are processed. Initial estimates may be off.

## Performance Considerations

### CPU-Bound vs I/O-Bound

- **Text files**: CPU-bound, benefit from higher parallelism
- **PDFs with images**: I/O and memory-bound, use moderate parallelism
- **Images**: I/O-bound, moderate parallelism

### Recommended Settings by Workload

| Workload Type | Parallel | Batch Size | Retry |
|---------------|----------|------------|-------|
| Mostly text   | 5-10     | N/A        | 2     |
| Mostly images | 3-5      | N/A        | 2     |
| Mostly PDFs   | 2-3      | 5-10       | 3     |
| Mixed         | 3-5      | 10         | 2     |
| Large PDFs    | 1-2      | 3-5        | 3     |

## Integration with Workflows

### Automated Batch Processing

```bash
#!/bin/bash
# daily-batch.sh

DATE=$(date +%Y%m%d)
REPORT_DIR="/var/reports"
DOCS_DIR="/var/documents/incoming"

weave docs batch \
  --directory "$DOCS_DIR" \
  --collection DailyDocs \
  --parallel 5 \
  --retry 3 \
  --create-report "$REPORT_DIR/$DATE-batch-report.csv"

# Clean up successfully processed files if needed
# find "$DOCS_DIR" -name "*.processed" -mtime +7 -delete
```

### CI/CD Integration

```yaml
# .github/workflows/process-docs.yml
name: Process Documents

on:
  push:
    paths:
      - 'documents/**'

jobs:
  process:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Batch process documents
        run: |
          weave docs batch \
            --directory ./documents \
            --collection CI_Docs \
            --parallel 3 \
            --create-report batch-report.csv
      - name: Upload report
        uses: actions/upload-artifact@v2
        with:
          name: batch-report
          path: batch-report.csv
```

## File Format Specifications

### .processed File Format

```json
{
  "file_path": "string",           // Full path to original file
  "processed_at": "RFC3339",       // ISO 8601 timestamp
  "success": boolean,              // Overall success status
  "error": "string",               // Error message if failed
  "text_chunks": integer,          // Number of text chunks created
  "images": integer,               // Number of images extracted
  "processing_time_ms": integer,   // Processing time in milliseconds
  "file_size": integer,            // File size in bytes
  "retry_count": integer,          // Number of retry attempts
  "chunks_failed": integer,        // Number of chunks that failed
  "images_failed": integer         // Number of images that failed
}
```

### CSV Report Format

| Column            | Type    | Description                        |
|-------------------|---------|------------------------------------|
| Timestamp         | string  | Processing timestamp               |
| File              | string  | Filename (basename)                |
| Status            | string  | "success" or "failed"              |
| TextChunks        | integer | Number of text chunks              |
| Images            | integer | Number of images                   |
| ProcessingTimeMs  | integer | Processing time in milliseconds    |
| FileSize          | integer | File size in bytes                 |
| RetryCount        | integer | Number of retries                  |
| ChunksFailed      | integer | Number of failed chunks            |
| ImagesFailed      | integer | Number of failed images            |
| Error             | string  | Error message if failed            |

## See Also

- [User Guide](USER_GUIDE.md) - Complete user documentation
- [Weave vs RagMe](WEAVE_VS_RAGME.md) - Schema differences
- [Demo](DEMO.md) - Interactive examples
