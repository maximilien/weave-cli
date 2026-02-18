# Client0 Ingestion Improvements Plan

**Prepared:** 2026-02-17
**Based on:** Client0 production analysis
**Client0 context:** 20 PDFs (~130MB catalogues + ~6MB results), ~2,600 images, 6 collections
**Key doc:** Client0's `INGESTION_LESSONS_LEARNED.md` and `WEAVE_CLI_ISSUES.md`

---

## Executive Summary

Client0 built a 255-line bash "disaster recovery" script (`reingest-pdf-monitored.sh`) to
work around weave-cli limitations. It implements: Milvus crash detection, 3-retry loops,
preventive restarts between every PDF, skip/resume flags, and chained run logic. This is
all complexity that belongs inside weave-cli itself. The improvements below eliminate the
need for that entire script and make bulk PDF ingestion reliable, observable, and resumable.

---

## Pain Points (Discovered From Client0)

| # | Severity | Pain Point | Client0 Workaround |
|---|---|---|---|
| P1 | 🔴 Critical | No idempotent ingestion — duplicates on re-run | Drop/recreate collection before each run |
| P2 | 🔴 Critical | No resume/checkpoint — full restart on failure | `START_FROM_PDF` env var + manual restart |
| P3 | 🔴 Critical | No per-file timeout — hangs on Milvus OOM | `ingest_file_monitored` with exit code watch |
| P4 | 🔴 Critical | stderr from sentence-transformers treated as error | Retry query (model cached on 2nd call) |
| P5 | 🟠 High | JSON output contaminated with progress messages | Scan stdout for first `{` line |
| P6 | 🟠 High | `--top-k` (hyphen) fails; `--top_k` works | Hardcoded `--top_k` in frontend |
| P7 | 🟠 High | No multi-file batch command — must script loops | 255-line bash wrapper |
| P8 | 🟠 High | Image collection must pre-exist for `--image-collection` | Pre-create step in pipeline |
| P9 | 🟡 Medium | Non-fatal flush timeouts look like failures | Manual doc count verification after |
| P10 | 🟡 Medium | No ingestion progress visibility across files | `ingestion-status.sh` polling script |
| P11 | 🟡 Medium | No ingestion state persistence / log file output | Custom logging via `tee -a logfile` |

---

## Proposed Features (Prioritized)

### Issue #34: Idempotent Document Ingestion (`--skip-existing`)

**Priority:** P0 — resolves P1, reduces need for P2
**Effort:** Medium (3-4 hours)
**Target:** v0.9.28

**Problem:** Every `weave docs create` on an already-ingested PDF adds duplicate documents.
Client0 must drop and recreate entire collections before re-ingesting a single updated PDF.

**Proposed Implementation:**

Add `--skip-existing` flag to `weave docs create`:

```bash
weave docs create AuctionListings 2022-catalogue.pdf \
  --milvus-local --skip-existing
```

Behavior:
- Before ingesting, check if a document with `source_document == basename(filePath)`
  already exists in the collection
- If exists: skip the file, print `⏭️ Skipping (already ingested): 2022-catalogue.pdf`
- If not: ingest normally
- Add `--overwrite` flag to replace existing (equivalent to today's default behavior)

**Metadata key to check:** `source_document` field (already populated in all text chunks)

**VDB Implementation:**
- Add `DocumentExistsBySource(ctx, collection, filename) (bool, error)` to
  `vectordb.VectorDBClient` interface
- Implement in each adapter using a metadata filter query

---

### Issue #35: Per-File Timeout (`--timeout`)

**Priority:** P0 — resolves P3 (the #1 operational pain: silent hangs)
**Effort:** Small (2 hours)
**Target:** v0.9.28

**Problem:** When Milvus runs OOM, `weave docs create` hangs indefinitely. Client0 wraps
every invocation with a 30-minute external timeout and Milvus health polling.

**Proposed Implementation:**

Add `--timeout` flag to `weave docs create`:

```bash
weave docs create AuctionImages catalogue.pdf \
  --milvus-local --timeout 30m
```

Behavior:
- Wrap the entire ingestion in a `context.WithTimeout`
- On timeout expiry: print clear error `❌ Timed out after 30m: catalogue.pdf`
- Exit with non-zero code so caller can detect and retry
- Support units: `s`, `m`, `h` (e.g., `30s`, `5m`, `2h`)

**Implementation:**
- `CreateDocument` already takes `ctx context.Context`
- Wire timeout from flag into context before calling

---

### Issue #36: Clean JSON Output (`--json` flag purity)

**Priority:** P1 — resolves P4, P5
**Effort:** Medium (3-4 hours)
**Target:** v0.9.28

**Problem A (stderr):** When sentence-transformers loads its model, tqdm progress bars go to
stderr. weave-cli captures stderr and treats it as an error. Queries fail on first call after
server startup even though the model loads successfully.

**Problem B (stdout contamination):** With `--json`, progress messages and config warnings
go to stdout, contaminating the JSON. Frontend must scan for first `{` line.

**Proposed Implementation:**

1. **Separate stderr from errors:** Only treat stderr as an error if the subprocess exits
   with non-zero status. Progress output on stderr with exit 0 should be ignored.

2. **JSON mode suppression:** When `--json` flag is set:
   - Route ALL non-JSON output to stderr (not stdout)
   - Suppress config warnings on stdout (or move to stderr)
   - stdout must contain ONLY valid JSON

3. **`--quiet` flag:** Suppress all progress/info output (useful for scripted pipelines):
   ```bash
   weave docs create AuctionListings file.pdf --milvus-local --quiet
   ```

**Code locations:**
- `src/cmd/utils/document.go` — `PrintError`, `PrintWarning`, `PrintSuccess` functions
- `src/pkg/output/` — output formatting package

---

### Issue #37: Flag Naming Consistency (`--top-k` vs `--top_k`)

**Priority:** P1 — resolves P6
**Effort:** Small (1 hour)
**Target:** v0.9.28

**Problem:** `--top-k` (hyphen, standard CLI convention) fails with "unknown flag".
`--top_k` (underscore) works. Client0 had to patch their TypeScript frontend.

**Proposed Fix:**

Add hyphen alias for `--top_k` in `cols query` command:
```go
// Support both for backward compatibility
cmd.Flags().IntVar(&topK, "top-k", 5, "Number of results to return")
cmd.Flags().IntVar(&topK, "top_k", 5, "Number of results to return (deprecated, use --top-k)")
```

Audit all flags for inconsistency — any `_` flag should have a `-` alias.

---

### Issue #38: Batch File Ingestion (`weave docs create-batch`)

**Priority:** P1 — resolves P7, P8, P10
**Effort:** Large (8-10 hours)
**Target:** v0.9.29

**Problem:** There is no native multi-file ingestion command. Client0 wrote a 255-line bash
script to process 20 PDFs sequentially with retry, restart, and skip logic.

**Proposed Implementation:**

New subcommand `weave docs create-batch`:

```bash
weave docs create-batch AuctionListings "data/tamarkin/*-catalogue.pdf" \
  --milvus-local \
  --embedding text-embedding-3-small \
  --skip-existing \
  --timeout 30m \
  --delay 10s \
  --max-retries 3 \
  --log-file logs/ingestion.log \
  --checkpoint-file .ingestion-checkpoint.json
```

**Key features:**

1. **Glob expansion** — already implemented in v0.9.27 (Issue #31)

2. **Sequential processing with configurable delay** (`--delay 10s`)
   - Prevents memory buildup in Milvus between files

3. **Retry with backoff** (`--max-retries 3`)
   - Auto-retry failed files up to N times
   - Configurable backoff: `--retry-delay 30s`

4. **Checkpoint/resume** (`--checkpoint-file`)
   - Save state after each successful file: `{"completed": ["2017-catalogue.pdf", ...], "failed": []}`
   - On re-run: skip already-completed files automatically
   - `--resume` flag: continue from checkpoint
   - `--ignore-checkpoint`: start fresh

5. **Structured log file** (`--log-file`)
   - Append-mode timestamped log
   - Replaces Client0's `tee -a logfile` pattern

6. **Summary report** — at end of batch:
   ```
   ✅ Batch complete: 9/9 succeeded, 0 failed, 2 skipped
   Duration: 1h 23m
   Log: logs/ingestion.log
   ```

7. **Auto-collection-create** — if `--image-collection` doesn't exist, create it
   automatically using the parent collection's schema

**Checkpoint file format:**
```json
{
  "collection": "AuctionListings",
  "started": "2026-02-17T10:00:00Z",
  "completed": [
    {"file": "2017-catalogue.pdf", "chunks": 28, "at": "2026-02-17T10:12:33Z"},
    {"file": "2018-catalogue.pdf", "chunks": 31, "at": "2026-02-17T10:25:11Z"}
  ],
  "failed": [],
  "skipped": []
}
```

---

### Issue #39: Ingestion Progress Dashboard (`weave docs status`)

**Priority:** P2 — resolves P10
**Effort:** Medium (4-5 hours)
**Target:** v0.9.29

**Problem:** Client0 wrote a separate `ingestion-status.sh` script that polls doc counts and
parses log files to show what's happening during a long ingestion run.

**Proposed Implementation:**

New command `weave docs status`:

```bash
weave docs status --milvus-local
```

Output:
```
📊 Collection Status (milvus-local)
════════════════════════════════════
Collection           Docs   Status   Last Updated
AuctionListings       267   ✅ ok    2026-02-17 14:32
AuctionListings_OSS   267   ✅ ok    2026-02-17 14:45
AuctionResults        127   ✅ ok    2026-02-17 15:01
AuctionResults_OSS    133   ✅ ok    2026-02-17 15:08
AuctionImages        6591   ✅ ok    2026-02-17 19:22
AuctionImages_OSS    6558   ✅ ok    2026-02-17 21:45

⚙️  Active ingestion: AuctionImages_OSS
   File: 2023-tamarkin-auction-catalogue.pdf [7/9]
   Chunks: 1,234 / ~2,000 estimated
   Images: 892 / ~250 this file
   Elapsed: 2h 14m
   ETA: ~45min remaining
```

**Live mode** (`--watch`): refresh every 5 seconds, like `top` for ingestion.

---

### Issue #40: Non-Fatal Error Classification

**Priority:** P2 — resolves P9
**Effort:** Small (2 hours)
**Target:** v0.9.29

**Problem:** Milvus flush timeout errors (`Failed to flush collection: DeadlineExceeded`)
appear identical to real failures but are non-fatal. Client0 learned this from trial and
error, not from weave-cli documentation or output.

**Proposed Implementation:**

Classify errors as fatal vs non-fatal in output:

```
⚠️  [non-fatal] Flush timeout (documents are stored): rpc error: DeadlineExceeded
```

vs:

```
❌ [fatal] Collection not found: AuctionImages — create it first
```

**Implementation:**
- In `processPDFFileGeneric` and Milvus adapter, detect flush timeout errors
- Log at `WARN` level (not `ERROR`) with explicit `[non-fatal]` tag
- Exit 0 if only non-fatal errors occurred

---

## Implementation Priority & Timeline

### v0.9.28 — "Reliability Release" (This Week, Tue-Wed)

| Issue | Feature | Hours | Impact |
|---|---|---|---|
| #35 | `--timeout` per-file | 2h | Eliminates silent hangs |
| #36 | JSON output purity + stderr fix | 4h | Fixes query failures + frontend parsing |
| #37 | `--top-k` flag alias | 1h | Fixes frontend compatibility |
| | **Total** | **7h** | |

These are fixes, not new features. Fast to ship, unblocks Client0 immediately.

### v0.9.29 — "Batch Ingestion Release" (Next Week)

| Issue | Feature | Hours | Impact |
|---|---|---|---|
| #34 | `--skip-existing` idempotent ingest | 3h | Eliminates duplicate docs |
| #38 | `weave docs create-batch` | 10h | Eliminates 255-line bash wrapper |
| #40 | Non-fatal error classification | 2h | Reduces operational confusion |
| | **Total** | **15h** | |

### v0.9.30 — "Observability Release" (Following Week)

| Issue | Feature | Hours | Impact |
|---|---|---|---|
| #39 | `weave docs status` dashboard | 5h | Real-time ingestion visibility |
| | **Total** | **5h** | |

---

## Quick Wins (Can Do Today)

Before building full features, these tiny fixes help Client0 immediately:

1. **`--top-k` alias** (Issue #37) — 1 hour, fixes frontend today
2. **JSON stdout purity** (part of Issue #36) — 2 hours, fixes frontend JSON parsing
3. **Warn on `--image-collection` pre-create** — 30 min, improve error message from
   "collection not found" to "create it first with: `weave cols create <name>`"

---

## What the 255-Line Script Does That weave-cli Should Own

```
Client0's reingest-pdf-monitored.sh    →   weave-cli feature
─────────────────────────────────────────────────────────────
3-retry loop per file                  →   --max-retries 3
Milvus crash detection (exit code)     →   --timeout + proper exit codes
Milvus restart between files           →   (VDB-level, out of scope)
Skip files before START_FROM_PDF       →   --checkpoint-file / --resume
SKIP_COLLECTION env vars               →   checkpoint knows what's done
tee -a logfile                         →   --log-file logs/ingestion.log
sleep between files                    →   --delay 10s
Milvus health polling after restart    →   (VDB-level, out of scope)
Summary: X/Y succeeded at end          →   Batch summary report
```

With Issues #34, #35, #38 implemented, Client0's bash script becomes:

```bash
weave docs create-batch AuctionListings "data/tamarkin/*-catalogue.pdf" \
  --milvus-local --embedding text-embedding-3-small \
  --skip-existing --timeout 30m --delay 10s --max-retries 3 \
  --log-file logs/ingestion.log --checkpoint-file .checkpoint.json

weave docs create-batch AuctionImages "data/tamarkin/*-catalogue.pdf" \
  --milvus-local --embedding text-embedding-3-small \
  --image-collection AuctionImages --image-storage minio --minio-bucket weave-images \
  --store-pdf --skip-existing --timeout 30m --delay 30s --max-retries 3 \
  --log-file logs/ingestion-images.log --checkpoint-file .checkpoint-images.json
```

**Two commands. No bash wrapper needed.**

---

## GitHub Issues Created

| Issue | Title |
|---|---|
| #34 | feat: per-file ingestion timeout (--timeout flag) |
| #35 | fix: JSON output purity and stderr not treated as error |
| #36 | fix: --top-k flag alias (hyphen) for --top_k (underscore) |
| #37 | feat: idempotent document ingestion (--skip-existing flag) |
| #38 | feat: batch file ingestion (weave docs create-batch) |
| #39 | feat: ingestion status dashboard (weave docs status) |
| #40 | fix: non-fatal flush timeout errors should not look like failures |

---

## Files Referenced (Client0 Codebase)

- `scripts/reingest-pdf-monitored.sh` — the 255-line workaround (Client0)
- `docs/INGESTION_LESSONS_LEARNED.md` — canonical pain point doc (Client0)
- `docs/issues/WEAVE_CLI_ISSUES.md` — bugs filed by Client0
- `docs/OSS_COLLECTIONS_BLOCKER.md` — embedding dim mismatch bug (Client0)
- `scripts/ingestion-status.sh` — monitoring workaround (Client0)
