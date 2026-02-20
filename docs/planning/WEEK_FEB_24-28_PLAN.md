# Week Plan: February 24-28, 2026

**Status**: 🚀 Starting fresh — v0.9.28 shipped Friday Feb 21, Issue #41 fixed Friday Feb 21
**Current Version**: v0.9.28 (released Wed Feb 18) + hotfix #41 (Feb 21) + hotfix #42 (pending)
**Last Updated**: 2026-02-21 (Friday PM)
**Work Schedule**: 4 hours/day, ~1 hour/day for client support

---

## 🎉 Last Week Victory (Feb 17-21) — v0.9.28 Released!

All 7 Client0 ingestion improvement issues shipped:

| Issue | Feature | Released |
|-------|---------|---------|
| #31 | `--workers` parallel processing | v0.9.27 |
| #33 | PDF storage in MinIO/S3 | v0.9.27 |
| #36 | `--top-k` alias (hyphen) | v0.9.28 |
| #35 | JSON stdout purity / stderr routing | v0.9.28 |
| #34 | `--timeout` per-file flag | v0.9.28 |
| #40 | Non-fatal Milvus flush timeout | v0.9.28 |
| #37 | `--skip-existing` idempotent ingestion | v0.9.28 |
| #29 | Milvus 65KB limit verified closed | v0.9.28 |
| #41 | Image-page association bug (wrong images) | v0.9.28-hotfix |

**Impact**: Client0 bash script (255 lines) is now ~10 lines. All pain points P1-P11 addressed.

---

## 🚨 BREAKING: Issue #42 Discovered Friday PM

**Issue #42**: Image documents missing `image_base64` and `image_url` fields after ingestion

**Severity**: 🔴 **CRITICAL** — images are ingested but inaccessible (no data/URL in metadata)

**Root cause**: Milvus `fromMilvusDocument` puts `ImageData` in `Document.ImageData` field but never copies it to `Metadata` map (Weaviate does this at `client_queries.go:600`). Client apps query metadata, not Document fields.

**Fix** (10 min): In `src/pkg/vectordb/milvus/client.go` `fromMilvusDocument()`, add:
```go
if imageData != "" {
    if metadata == nil {
        metadata = make(map[string]interface{})
    }
    metadata["image_base64"] = imageData
}
if image != "" && strings.HasPrefix(image, "http") {
    metadata["image_url"] = image
}
```

**Monday priority**: Fix #42 **before** starting #38. Client0 cannot use their image collections until this is fixed.

---

## 🎯 This Week Goal: v0.9.29

**Theme**: "Batch ingestion & observability"

**Target issues** (UPDATED):
- **Issue #42**: Image metadata fields missing — **10 min (Mon AM, URGENT)**
- **Issue #38**: `weave docs create-batch` — 10h (Mon PM-Thu)
- **Issue #39**: `weave docs status` dashboard — 5h (Thu-Fri)

**Deliverable**: v0.9.29-hotfix (Mon AM), then v0.9.29 (Fri) with both features shipped.

---

## 📊 Open Issues (Updated Feb 21 PM)

| # | Priority | Issue | Target | Est. |
|---|----------|-------|--------|------|
| #42 | P0 🔴🔴 | Image fields missing (`image_base64`, `image_url`) | v0.9.29-hotfix | 10 min |
| #38 | P1 🔴 | `weave docs create-batch` — glob + checkpoint + retry + delay | v0.9.29 | 10h |
| #39 | P2 🟠 | `weave docs status` dashboard — collection status + --watch | v0.9.29 | 5h |
| #21 | P3 🟡 | Image ingestion tests across all VDBs | v0.9.30 | 8h |
| #16 | P4 🟡 | Code audit + prep for v1.0 | v1.0-pre | 20h |
| #15 | P4 🟡 | Documentation updates | v1.0-pre | 10h |
| #14 | P5 🟢 | Agent configs (different agent types) | v1.0 | 15h |
| #12 | P5 🟢 | Tips on `command -h` | v1.0 | 3h |
| #11 | P5 🟢 | Streamline commands and shortcuts | v1.0 | 5h |
| #8  | P5 🟢 | Extract various PDF versions for testing | v0.9.x | 4h |

---

## 🗓️ Daily Breakdown (UPDATED for Issue #42)

### Monday, Feb 24 — Issue #42 hotfix + Issue #38 core (4 hours)

**URGENT: Issue #42 fix** (15 min):
- Fix `src/pkg/vectordb/milvus/client.go` `fromMilvusDocument()` — add `image_base64` and `image_url` to metadata map
- Build + lint
- Commit + push → v0.9.29-hotfix tag
- Close Issue #42

**Client check** (15 min):
- Notify Client0 of hotfix
- Check for additional feedback on v0.9.28

**Issue #38 — `create-batch` core** (3.5 hours):
1. Create `src/cmd/document/create_batch.go` — new `CreateBatchCmd`
2. Positional args: `COLLECTION_NAME GLOB_PATTERN [GLOB_PATTERN...]`
3. Implement flags: `--delay`, `--max-retries`, `--retry-delay`, `--checkpoint-file`, `--log-file`
4. Wire glob expansion (reuse from `create.go` `runGlobDocumentCreate`)
5. Checkpoint load/save (JSON format per design doc)
6. Per-file delay between ingestions

**Deliverable**: v0.9.29-hotfix released + basic `create-batch` working for sequential mode with checkpoint

---

### Tuesday, Feb 25 — Issue #38 retry + log (4 hours)

**Issue #38 continued** (4 hours):
1. Retry with exponential backoff (`--max-retries`, `--retry-delay`)
2. Structured log file (`--log-file`) — timestamped append-mode
3. Wire `--skip-existing` (reuse flag from `create.go`)
4. Wire `--timeout` per-file (reuse from `create.go`)
5. Batch summary output at completion

**Deliverable**: Retry + logging working end-to-end

---

### Wednesday, Feb 26 — Issue #38 polish + tests (4 hours)

**Issue #38 polish** (2 hours):
1. Auto-create image collection if `--image-collection` doesn't exist
2. `--json` output for CI/CD (reuse `BatchReport` from `batch.go`)
3. Update `--skip-existing` in `batch.go` to use `DocumentExistsByFilename` (currently TODO stub)
4. Lint + build

**Tests** (2 hours):
1. Unit tests for checkpoint load/save
2. Unit tests for retry logic
3. Integration test for glob + checkpoint scenario

**Deliverable**: Issue #38 complete, PR merged

---

### Thursday, Feb 27 — Release v0.9.29-pre + Issue #39 start (4 hours)

**Morning** (1 hour):
- Release v0.9.29-pre with Issue #38
- Client check on new feature

**Issue #39 — `weave docs status`** (3 hours):
1. Create `src/cmd/document/status.go` — new `StatusCmd`
2. List all collections for selected VDB
3. Per-collection doc count + last-updated timestamp
4. Active ingestion detection (read checkpoint files in CWD)
5. Formatted table output

**Deliverable**: Basic `docs status` showing collection counts

---

### Friday, Feb 28 — Issue #39 polish + release v0.9.29 (4 hours)

**Issue #39 polish** (2 hours):
1. `--watch` flag — refresh every 5 seconds (use `time.Ticker`)
2. Detect in-progress ingestion from checkpoint file
3. `--json` output mode
4. Lint + tests

**Release** (1 hour):
- Build + tests + lint
- Tag v0.9.29
- GitHub release notes
- Close issues #38 + #39

**Planning** (1 hour):
- Write v0.9.30 week plan (Issue #21 image tests + Issue #16 audit)
- GitHub issues triage

**Deliverable**: v0.9.29 released with `create-batch` + `docs status` 🚀

---

## 🎯 Weekly Goals

### Must Have
- [ ] **Issue #42 hotfix released** (Mon AM, 15 min)
- [ ] Issue #38: `weave docs create-batch` complete
- [ ] Issue #39: `weave docs status` complete
- [ ] v0.9.29 released by Friday
- [ ] All tests passing, lint clean
- [ ] Client0/Client1 responsive (<2h turnaround)

### Should Have
- [ ] `--skip-existing` wired in `batch.go` (currently stubbed)
- [ ] Unit tests for checkpoint logic
- [ ] Integration tests for `create-batch`

### Nice to Have
- [ ] `--watch` mode for `docs status`
- [ ] Video demo of `create-batch` replacing Client0's bash script
- [ ] Issue #8 quick wins (PDF test extraction)

---

## ⏰ Time Allocation

**Daily**:
- 09:00-09:15: Issue #42 hotfix (Mon only)
- 09:15-09:30: Client support window
- 09:30-12:30: Primary work block (3 hours)
- 13:00-14:00: Secondary work block (1 hour)

**Weekly Total**: ~20 hours (15h dev + 5h client support)

---

## 🔮 Issue #42 Details (CRITICAL)

**Symptom**: Client0 queries return image documents with `type: image` but **no `image_base64` or `image_url` fields** in metadata. Images are inaccessible.

**Root cause**: Milvus stores image data in dedicated `FieldImageData` column (not metadata). When querying, `fromMilvusDocument()` correctly extracts `ImageData` from the column and puts it in `Document.ImageData`, but this field is never copied to `Document.Metadata`. Client apps only see metadata.

**Weaviate comparison**: `client_queries.go:600` adds `image_base64` to metadata after query:
```go
if imageData, ok := resultItem["image_data"].(string); ok && imageData != "" {
    metadata["image_base64"] = imageData
}
```

**Fix location**: `src/pkg/vectordb/milvus/client.go` `fromMilvusDocument()` — add same logic.

**Impact**: **BLOCKS** Client0 image retrieval. They cannot use their 6,591-document image collection until fixed.

---

## 🔮 Issue #38 Design Summary

See `ISSUE_38_CREATE_BATCH_DESIGN.md` for full spec.

**Key decisions**:
- New command `weave docs create-batch` (not extending `weave docs batch`)
- Positional args match `create`: `COLLECTION_NAME GLOB [GLOB...]`
- Checkpoint JSON at `--checkpoint-file` (default: `.weave-checkpoint-COLLECTION.json`)
- `--delay 10s` between files (prevent Milvus OOM on rapid sequential ingestion)
- Retry with exponential backoff (1s, 2s, 4s...)
- Reuses all existing `create.go` flags (`--skip-existing`, `--timeout`, `--workers`, etc.)

---

## 📌 Key Files

### New Files (This Week)
- `src/cmd/document/create_batch.go` — Issue #38 command
- `src/cmd/document/status.go` — Issue #39 command
- `src/cmd/document/create_batch_test.go` — unit tests

### Modified Files (Hotfix #42)
- **`src/pkg/vectordb/milvus/client.go`** — `fromMilvusDocument()` add image fields to metadata

### Modified Files (Issue #38)
- `src/cmd/document/document.go` — register new commands
- `src/cmd/document/batch.go` — wire real `--skip-existing` (remove TODO stub)

### Planning Docs
- `docs/planning/WEEK_FEB_24-28_PLAN.md` (this file)
- `docs/planning/ISSUE_38_CREATE_BATCH_DESIGN.md` (detailed design)

---

**Prepared**: 2026-02-21 (Friday PM, updated for Issue #42)
**Previous week**: `docs/planning/WEEK_FEB_17-21_CONSOLIDATED.md`
