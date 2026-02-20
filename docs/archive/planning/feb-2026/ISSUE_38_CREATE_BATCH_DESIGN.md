# Issue #38: `weave docs create-batch` Design Doc

**Issue**: #38
**Target**: v0.9.29
**Effort**: ~10 hours (Mon-Wed, Feb 24-26)
**Author**: dr.max
**Prepared**: 2026-02-21

---

## Problem

Client0 wrote a 255-line bash "disaster recovery" wrapper to handle batch PDF ingestion.
It implements retry, crash detection, skip/resume, and delay logic — all complexity that
belongs in weave-cli. Issue #38 ships this as a first-class command.

---

## Command Design

```bash
weave docs create-batch COLLECTION_NAME GLOB [GLOB...] [flags]
```

### Examples

```bash
# Basic: ingest all PDFs in data/pdfs/
weave docs create-batch AuctionListings "data/pdfs/*.pdf" --milvus-local

# Full Client0 replacement:
weave docs create-batch AuctionListings "data/pdfs/*-catalogue.pdf" \
  --milvus-local \
  --embedding text-embedding-3-small \
  --skip-existing \
  --timeout 30m \
  --delay 10s \
  --max-retries 3 \
  --retry-delay 30s \
  --log-file logs/ingestion.log \
  --checkpoint-file .weave-checkpoint.json \
  --image-storage minio \
  --minio-bucket my-bucket

# Resume after crash (checkpoint remembers what was done):
weave docs create-batch AuctionListings "data/pdfs/*.pdf" \
  --milvus-local \
  --checkpoint-file .weave-checkpoint.json

# CI/CD with JSON output:
weave docs create-batch AuctionListings "data/pdfs/*.pdf" \
  --milvus-local --json | jq '.status'
```

---

## Flags

### New flags (create-batch specific)

| Flag | Default | Description |
|------|---------|-------------|
| `--delay` | `0s` | Delay between files (e.g., `10s`, `1m`). Prevents Milvus OOM on rapid sequential ingestion. |
| `--max-retries` | `3` | Retry attempts per file on failure |
| `--retry-delay` | `30s` | Initial delay before retry (doubles each attempt: 30s, 60s, 120s) |
| `--checkpoint-file` | `.weave-checkpoint-COLLECTION.json` | JSON file tracking completed/failed files |
| `--log-file` | `` (none) | Append-mode timestamped log file |
| `--continue` | `false` | Resume from existing checkpoint (skip completed files) |
| `--reset-checkpoint` | `false` | Ignore existing checkpoint, start fresh |

### Inherited from `create` (all supported)

| Flag | Description |
|------|-------------|
| `--skip-existing` | Skip files already in VDB by filename |
| `--timeout` | Per-file timeout (e.g., `30m`) |
| `--workers` | Parallel workers (default: 1) |
| `--chunk-size` | Text chunk size |
| `--batch-size` | Image batch size |
| `--image-collection` | Collection for extracted images |
| `--image-storage` | `minio`, `s3`, or `local` |
| `--embedding` | Embedding model |
| `--skip-small-images` | Skip small images in PDFs |
| `--store-pdf` | Store PDF in external storage |
| `--json` | JSON output for CI/CD |

---

## Checkpoint File Format

Default path: `.weave-checkpoint-COLLECTIONNAME.json`

```json
{
  "version": 1,
  "collection": "AuctionListings",
  "started_at": "2026-02-24T10:00:00Z",
  "updated_at": "2026-02-24T12:34:56Z",
  "glob_patterns": ["data/pdfs/*.pdf"],
  "completed": [
    {
      "file": "data/pdfs/2017-catalogue.pdf",
      "chunks": 28,
      "images": 312,
      "completed_at": "2026-02-24T10:12:33Z",
      "duration_ms": 743210
    },
    {
      "file": "data/pdfs/2018-catalogue.pdf",
      "chunks": 31,
      "images": 287,
      "completed_at": "2026-02-24T10:24:11Z",
      "duration_ms": 691004
    }
  ],
  "failed": [
    {
      "file": "data/pdfs/2019-catalogue.pdf",
      "error": "context deadline exceeded (timeout: 30m)",
      "attempts": 3,
      "last_attempt_at": "2026-02-24T11:45:00Z"
    }
  ],
  "skipped": [
    {
      "file": "data/pdfs/2020-catalogue.pdf",
      "reason": "already ingested (--skip-existing)"
    }
  ]
}
```

### Resume logic

On startup with `--continue`:
1. Load checkpoint file
2. Build set of completed file paths
3. Filter glob results: skip completed files
4. Log: `"Resuming batch: 2 completed, 7 remaining"`

Without `--continue`: if checkpoint exists, warn and offer `--continue` or `--reset-checkpoint`.

---

## Log File Format

Append-mode, one line per event:

```
2026-02-24T10:00:00Z [INFO] Batch started: AuctionListings, 9 files
2026-02-24T10:00:00Z [INFO] Processing: data/pdfs/2017-catalogue.pdf (1/9)
2026-02-24T10:12:33Z [OK]   Completed: data/pdfs/2017-catalogue.pdf — 28 chunks, 312 images, 12m33s
2026-02-24T10:12:43Z [INFO] Delay 10s...
2026-02-24T10:12:53Z [INFO] Processing: data/pdfs/2018-catalogue.pdf (2/9)
2026-02-24T10:24:11Z [OK]   Completed: data/pdfs/2018-catalogue.pdf — 31 chunks, 287 images, 11m18s
2026-02-24T11:15:00Z [WARN] Failed: data/pdfs/2019-catalogue.pdf — context deadline exceeded (attempt 1/3)
2026-02-24T11:45:00Z [WARN] Failed: data/pdfs/2019-catalogue.pdf — context deadline exceeded (attempt 2/3)
2026-02-24T12:15:00Z [ERROR] Failed (max retries): data/pdfs/2019-catalogue.pdf — context deadline exceeded
2026-02-24T12:34:56Z [INFO] Batch complete: 8/9 succeeded, 1 failed, 0 skipped — 2h 34m 56s
```

---

## Retry Logic

```
attempt 1: immediate
attempt 2: wait retry-delay (default 30s)
attempt 3: wait retry-delay * 2 (default 60s)
attempt N: wait retry-delay * 2^(N-2)
```

Capped at `retry-delay * 8` (no infinite backoff).

Retryable errors:
- `context deadline exceeded` (timeout — retry with fresh context)
- `connection refused` / `dial tcp` (VDB restart)
- `server unavailable`
- Any error containing `EOF`

Non-retryable:
- File not found
- Permission denied
- `--skip-existing` skip (not an error)

---

## Batch Summary Output

### Text mode (stderr):

```
════════════════════════════════════════════════════════════
✅ Batch complete: AuctionListings
════════════════════════════════════════════════════════════
Files:      9 total — 8 succeeded, 1 failed, 0 skipped
Documents:  248 chunks, 2,891 images created
Duration:   2h 34m 56s | Avg: 17m 13s/file
Checkpoint: .weave-checkpoint-AuctionListings.json
Log:        logs/ingestion.log
════════════════════════════════════════════════════════════
Failed files:
  ❌ 2019-catalogue.pdf — context deadline exceeded (3 retries)
     → Retry with: weave docs create-batch AuctionListings "data/pdfs/2019-catalogue.pdf" --milvus-local --timeout 60m
════════════════════════════════════════════════════════════
```

### JSON mode (stdout, `--json`):

```json
{
  "status": "partial",
  "collection": "AuctionListings",
  "started_at": "2026-02-24T10:00:00Z",
  "completed_at": "2026-02-24T12:34:56Z",
  "duration_seconds": 9296,
  "files": {
    "total": 9,
    "succeeded": 8,
    "failed": 1,
    "skipped": 0
  },
  "documents": {
    "chunks_created": 248,
    "images_created": 2891
  },
  "failed_files": [
    {
      "file": "data/pdfs/2019-catalogue.pdf",
      "error": "context deadline exceeded",
      "attempts": 3
    }
  ],
  "checkpoint_file": ".weave-checkpoint-AuctionListings.json",
  "exit_code": 1
}
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | All files succeeded (or skipped) |
| 1 | Partial success — some files failed |
| 2 | Complete failure — no files succeeded |

---

## Implementation Plan

### Monday (3.5h) — Core

1. **`src/cmd/document/create_batch.go`** — new file
   - `CreateBatchCmd` cobra command
   - `init()` — all flags
   - `runCreateBatch()` — main handler
   - `loadCheckpoint()` / `saveCheckpoint()` — JSON checkpoint I/O
   - `appendLog()` — log file writer
   - `expandGlobs()` — reuse glob logic from create.go

2. **`src/cmd/document/document.go`** — register `CreateBatchCmd`

### Tuesday (4h) — Retry + Logging

3. `processFileWithRetry()` — retry loop with exponential backoff
4. `isRetryableError()` — error classification
5. Wire `--delay` between files
6. Wire inherited flags: `--skip-existing`, `--timeout`, `--workers`, `--embedding`, etc.
7. Batch summary output (text + JSON)

### Wednesday (4h) — Polish + Tests

8. Auto-create image collection if missing
9. `--continue` / `--reset-checkpoint` checkpoint resume logic
10. Update `batch.go` `--skip-existing` stub → wire `DocumentExistsByFilename`
11. Unit tests: `TestCheckpointLoadSave`, `TestRetryLogic`, `TestBatchSummary`
12. Build + lint + integration test

---

## Relationship to Existing `weave docs batch`

`weave docs batch` (existing) takes `--directory` and uses `.processed` sidecar files.
`weave docs create-batch` (new) takes glob patterns and uses a JSON checkpoint file.

Both remain. `create-batch` is the recommended command going forward — it:
- Matches `create`'s UX (glob as positional arg)
- Has a portable checkpoint (single JSON file, not scattered `.processed` files)
- Has structured retry with backoff
- Has `--delay` for Milvus stability

Eventually `batch` may be deprecated in favor of `create-batch`.

---

## Files Changed

| File | Change |
|------|--------|
| `src/cmd/document/create_batch.go` | NEW — create-batch command |
| `src/cmd/document/create_batch_test.go` | NEW — unit tests |
| `src/cmd/document/document.go` | register `CreateBatchCmd` |
| `src/cmd/document/batch.go` | wire real `--skip-existing` (remove TODO stub) |

---

**Design status**: APPROVED — ready to implement Monday Feb 24
