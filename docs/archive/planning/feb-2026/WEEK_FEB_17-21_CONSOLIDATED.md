# Week Plan: February 17-21, 2026

**Status**: 🎉 MONDAY SUCCESS - v0.9.27 RELEASED! Issue #31 COMPLETE!
**Current Version**: v0.9.27 (released Mon Feb 17)
**Last Updated**: 2026-02-17 (Monday afternoon)
**Work Schedule**: 4 hours/day, ~1 hour/day for client feedback

---

## 🎉 Monday Victory (Feb 17) - v0.9.27 Released!

### Issue #31: Parallel Document Processing - COMPLETE! 🚀
- ✅ **Worker pool integration** - Parallel chunk processing (1-10 workers)
- ✅ **Glob pattern support** - Batch file selection: `"*.pdf"`, `"**/*.md"`
- ✅ **Progress aggregation** - Real-time ETA with progress bars
- ✅ **Integration tests** - 5 comprehensive test functions
- ✅ **Documentation** - Complete guide with performance benchmarks
- ✅ **All tests passing** - Build ✅ Lint ✅ Tests ✅

**Performance**: 3x speedup (45s → 15s for 10 chunks with 3 workers)

**Commits**: 8 total
- d068a28: Worker pool integration
- 9ff15b4: Glob pattern support
- 2ed719f: Progress aggregation
- a01d7e0: Integration tests
- b520084: Documentation
- 17ae738: Bug fixes (test config + markdown lint)

---

## 📊 Current State Check

### Versions
- **v0.9.26**: Document ingestion fix (Milvus) - COMPLETE ✅ (Feb 15)
- **v0.9.27**: Parallel document processing - COMPLETE ✅ (Feb 17)

### Open Issues (Updated)
1. ~~**Issue #31**: Parallel processing~~ - **COMPLETE** ✅ (v0.9.27)
2. ~~**Issue #33**: PDF storage in MinIO~~ - **COMPLETE** ✅ (v0.9.27, Mon Feb 17)
3. **Issue #36**: `--top-k` flag alias — 1h fix (v0.9.28)
4. **Issue #35**: JSON output purity + stderr fix — 4h (v0.9.28)
5. **Issue #34**: `--timeout` per-file flag — 2h (v0.9.28)
6. **Issue #37**: `--skip-existing` idempotent ingestion — 3h (v0.9.28)
7. **Issue #40**: Non-fatal flush timeout classification — 2h (v0.9.28)
8. **Issue #38**: `weave docs create-batch` — 10h (v0.9.29, next week)
9. **Issue #39**: `weave docs status` dashboard — 5h (v0.9.29, next week)
10. **Issue #29**: Milvus 65KB limit (likely fixed by external storage v0.9.21-23)
11. **Issue #21**: Image ingestion tests across all VDBs
12. **Issue #16**: Code audit for v1.0
13. **Issue #15**: Documentation updates

---

## 🗓️ Daily Breakdown (4 hrs/day + 1 hr client support)

### ✅ Monday, Feb 17 - COMPLETE!

**Actual Work Done** (~4 hours):
- ✅ Wire worker pool to document ingestion (1 hour)
- ✅ Add glob pattern support (45 min)
- ✅ Implement progress aggregation (1 hour)
- ✅ Create integration tests (45 min)
- ✅ Write documentation (30 min)
- ✅ Fix test failures + lint issues (30 min)
- ✅ Release v0.9.27 (15 min)

**Result**: Issue #31 100% COMPLETE, v0.9.27 released! 🎉

---

### Tuesday, Feb 18 (4 hours work) - v0.9.28 Reliability Fixes

**Goal**: Ship v0.9.28 — the "it just works" release for Client0

**Client Check** (30 min):
- Check Client0/Client1 feedback on v0.9.27 + Issue #33

**Issue #36** — `--top-k` flag alias (1 hour):
- Add hyphen alias for `--top_k` in `cols query`
- Audit all other flags for underscore/hyphen inconsistency
- Fixes Client0's TypeScript frontend compatibility

**Issue #35** — JSON output purity + stderr fix (3 hours):
- Route all non-JSON output to stderr when `--json` is set
- Stop treating sentence-transformers tqdm stderr as error
- Add `--quiet` flag to suppress all progress/info output
- Tests + lint

**Deliverable**: Issues #36 + #35 merged, v0.9.28-pre ready

---

### Wednesday, Feb 19 (4 hours work) - v0.9.28 continued

**Issue #34** — `--timeout` per-file flag (2 hours):
- Wire `context.WithTimeout` into `CreateDocument` from flag
- Support `30s`, `5m`, `2h` syntax
- Clear timeout error message + non-zero exit
- Tests

**Issue #40** — Non-fatal flush timeout classification (2 hours):
- Detect Milvus `DeadlineExceeded` flush errors in adapter
- Log at WARN with `[non-fatal]` tag instead of ERROR
- Exit 0 if only non-fatal errors occurred
- Tests

**Deliverable**: All 4 v0.9.28 issues complete → **release v0.9.28** 🚀

---

### Thursday, Feb 20 (4 hours work) - v0.9.28 + start #37

**Morning** (1 hour):
- Client check, respond to any feedback on v0.9.28
- Issue #29 quick verification (30 min) — close if resolved by external storage

**Issue #37** — `--skip-existing` idempotent ingestion (3 hours):
- Add `DocumentExistsBySource()` check before ingesting each file
- `--skip-existing` flag skips files already in collection by `source_document`
- `--overwrite` flag replaces existing (makes current behavior explicit)
- Tests across Milvus + mock adapters

**Deliverable**: Issue #37 complete, PR open

---

### Friday, Feb 21 (4 hours work) - v0.9.28 wrap-up + planning

**Morning** (2 hours):
- Merge Issue #37 → **release v0.9.28 final** with all 5 fixes
- Update week plan doc + next week plan

**Afternoon** (2 hours):
- Write next week plan (v0.9.29: Issue #38 `create-batch` + Issue #39 `docs status`)
- Issue #38 design doc — checkpoint format, flag spec, retry logic
- GitHub issues triage

**Deliverable**: v0.9.28 released, next week plan ready, Issue #38 designed

---

## 🎯 Weekly Goals - UPDATED

### Must Have (Primary Focus)
- ✅ Client0/Client1 support (responsive, same-day turnaround)
- ✅ **Issue #31 complete** - DONE! v0.9.27 released 🎉
- ✅ **Issue #33 complete** - PDF storage in MinIO, Mon Feb 17 🎉
- ✅ All tests passing (no regressions)
- [ ] **v0.9.28 released** — Issues #34 + #35 + #36 + #37 + #40 (Tue-Fri)

### Should Have (Client0 Reliability)
- [ ] Issue #36: `--top-k` alias (1h, Tue)
- [ ] Issue #35: JSON purity + stderr fix (4h, Tue)
- [ ] Issue #34: `--timeout` flag (2h, Wed)
- [ ] Issue #40: Non-fatal flush errors (2h, Wed)
- [ ] Issue #37: `--skip-existing` (3h, Thu)

### Nice to Have (Bonus)
- [ ] Issue #29 verified/closed (30 min, Thu)
- [ ] Issue #38 designed for next week
- [ ] Client feedback on v0.9.28 incorporated

### Monday Achievements ✅
- Issue #31: Parallel processing (100% complete)
- Issue #33: PDF storage in MinIO (100% complete)
- v0.9.27: Released with full test coverage
- Client0 ingestion analysis: 7 new GitHub issues (#34-40)
- Planning doc: CLIENT0_INGESTION_IMPROVEMENTS.md

---

## ⏰ Time Allocation Summary

**Daily Schedule**:
- 09:00-10:00: Client support window (check issues, respond)
- 10:00-13:00: Primary work block (3 hours)
- 13:00-14:00: Lunch + context switch
- 14:00-15:00: Secondary work block (1 hour)

**Weekly Total**:
- Client support: ~5 hours (1 hour/day)
- Development: ~15 hours (3 hours/day)
- **Total**: ~20 hours

**Allocation**:
- Issue #31 (if pursued): 8-10 hours
- Client support: 5 hours
- Stretch goals: 5-7 hours

---

## 📋 Issue #31: Parallel Processing (Optional This Week)

**Status**: 60% complete (infrastructure done)

**Remaining Work** (40%):
- Wire worker pool to document ingestion (~2-3 hours)
- Add glob pattern support (~1 hour)
- Progress aggregation (~30 min)
- Integration tests (~1 hour)
- Documentation (~30 min)

**Total**: 4-5 hours

**Decision Point**: Monday morning
- **If Client0/Client1 quiet**: Pursue Issue #31
- **If Client support needed**: Defer to next week

**Impact**: 2-3x speedup for bulk ingestion (2-3 hours → 45-60 min for 10 PDFs)

---

## 🎥 Stretch Goal: Video Demos

**Ready to Record**: 3 scripts complete in `videos/scripts/`

1. **oss-embeddings-basic.cast** (~2 min)
   - Setup sentence-transformers
   - Create collection with OSS model
   - Ingest documents
   - Query and compare results

2. **oss-embeddings-reembed.cast** (~3 min)
   - Existing OpenAI collection
   - Re-embed with sentence-transformers (20x faster)
   - Compare performance and costs

3. **oss-embeddings-compare.cast** (~3 min)
   - Side-by-side OpenAI vs OSS
   - Performance comparison
   - Cost analysis ($240/year savings)

**Prerequisites**:
- sentence-transformers installed
- Milvus running locally
- Test data prepared

**Time**: 3-4 hours total (recording + upload + README updates)

---

## 🧪 Stretch Goal: External Storage Tests

**Goal**: Real-world validation with actual images

**Test Cases**:
1. Small images (<10KB) - Direct VDB storage
2. Medium images (10-50KB) - Thumbnail decision point
3. Large images (50-100KB) - External storage required
4. Very large (>500KB) - Stress test
5. Multiple formats (JPEG, PNG, GIF, WebP)

**VDBs to Test**:
- ✅ Milvus (already implemented)
- [ ] Weaviate
- [ ] Qdrant
- [ ] Chroma

**Deliverable**: Integration test suite + performance data

**Time**: 2-3 hours

---

## 📝 Stretch Goal: Documentation Polish

**Tasks**:
- [ ] Link checker automation script
- [ ] Spell check pass (all markdown)
- [ ] Code example testing (copy-paste verification)
- [ ] Terminology consistency check
- [ ] Table of contents updates
- [ ] Cross-reference validation

**Tools**:
- markdownlint (already integrated)
- Link checker (create script)
- Spell checker (aspell/hunspell)

**Time**: 2-3 hours

---

## 📊 Stretch Goal: Performance Benchmarking

**Benchmarks to Run**:
- OSS re-embedding speed (docs/sec)
- External storage upload speed (images/sec)
- Thumbnail generation time (ms/image)
- Cost analysis (OpenAI vs OSS embeddings)
- Storage size comparison (VDB vs S3)

**Deliverable**: `BENCHMARKS.md` with detailed metrics

**Time**: 3-4 hours

---

## 🚨 Client Support Protocol

### Same-Day Response Guarantee

**Critical Issues** (Blocking production):
- Drop everything
- Debug immediately
- Hotfix release within 4 hours
- Stretch goals paused

**Medium Issues** (Not blocking):
- Respond within 2 hours
- Fix by end of day
- Point release next day
- Continue with planned work

**Feature Requests**:
- Assess scope and urgency
- Small (<2 hours): Implement same day
- Medium (<1 day): Plan for this week
- Large (>1 day): Create detailed plan, discuss timeline

**Questions/Clarifications**:
- Respond within 1 hour
- Provide examples and documentation
- Update docs if question reveals gap

---

## 📈 Success Metrics

### Code Quality
- [ ] Zero critical bugs
- [ ] All tests passing (no regressions)
- [ ] Linting clean
- [ ] No known production blockers

### Client Satisfaction
- [ ] Client0 happy with v0.9.26
- [ ] Client1 supported (if needed)
- [ ] Response time <2 hours for all inquiries
- [ ] Same-day resolution for urgent issues

### Progress
- [ ] At least 1 major goal completed (Issue #31 OR 2+ stretch goals)
- [ ] Documentation current and accurate
- [ ] Planning docs organized (archived old docs)

---

## 🔮 Scenarios & Responses

### Scenario 1: Both Clients Quiet ✅
**Action**:
- Pursue Issue #31 aggressively (Mon-Wed)
- Release v0.9.27 by Wednesday
- Stretch goals Thu-Fri

### Scenario 2: Client0 Needs Minor Support 🟡
**Action**:
- Morning: Client support (1-2 hours)
- Afternoon: Issue #31 OR stretch goals (2-3 hours)
- Flexibility to adapt daily

### Scenario 3: Client0 Critical Issue 🔴
**Action**:
- All hands on deck
- Issue #31 paused
- Stretch goals deferred
- Focus 100% on resolution

### Scenario 4: Client1 Enters Production 💡
**Action**:
- Split time between clients
- Issue #31 may be deferred
- Prioritize production stability
- Stretch goals paused

---

## 📌 Files & Locations

### Planning Docs (Active)
- `/docs/planning/WEEK_FEB_17-21_CONSOLIDATED.md` (this file)
- `/docs/planning/MONDAY_QUICK_START_FEB_17.md` (reference)
- `/docs/planning/NEXT_WEEK_FEB_17-21.md` (original plan)
- `/docs/planning/V1_0_ROADMAP.md` (long-term vision)

### Planning Docs (Archived)
- `/docs/archive/feb-2026/HOTFIX_FEB_15_ISSUE_32.md`
- `/docs/archive/feb-2026/WEEKEND_SUMMARY_FEB_14-15.md`
- `/docs/archive/feb-2026/STATUS_SNAPSHOT_FEB_15.md`
- `/docs/archive/feb-2026/WEEK_SUMMARY_FEB_10-13.md`
- (14 other historical docs archived)

### Key Code Files
- `src/pkg/vectordb/milvus/document.go` - Latest fix (v0.9.26)
- `src/pkg/vectordb/milvus/collection.go` - Collection creation (v0.9.25)
- `src/pkg/worker/pool.go` - Worker pool (ready for Issue #31)
- `src/pkg/ratelimit/ratelimit.go` - Rate limiting (ready for Issue #31)

---

## 🎯 Monday Morning Checklist

**Before Starting Work**:
- [ ] Check GitHub Issue #32 for Client0 feedback
- [ ] Check GitHub issues for Client1 activity
- [ ] Review email/Slack for direct messages
- [ ] Decide: Issue #31 OR stretch goals?

**If Issue #31**:
- [ ] Read `MONDAY_QUICK_START_FEB_17.md` for detailed plan
- [ ] Start with worker pool wiring
- [ ] Goal: Basic parallel ingestion by EOD

**If Stretch Goals**:
- [ ] Verify Issue #29 (30 min quick win)
- [ ] Start video demos OR external storage tests
- [ ] Update docs as you go

---

## 💡 Key Insights from Weekend

**What Worked**:
- Systematic debugging with debug output
- Comprehensive testing at each step
- Clear commit messages with context
- Detailed planning documents

**What to Continue**:
- Same rapid response for client issues
- Incremental commits with good messages
- Testing before release
- Documentation as we go

**Lessons Learned**:
- Don't assume fix is complete until full workflow tested
- Column dimension parameters are critical for Milvus
- OSS embeddings require provider routing at ALL ingestion points
- Client feedback is invaluable for catching bugs

---

## 🚀 Mindset for the Week

**Primary**: Client success > feature development
**Secondary**: Finish what we start (Issue #31 OR 2+ stretch goals)
**Tertiary**: Keep planning docs current
**Bonus**: Have fun building! 😎

**LFG!** 🚀

---

**Prepared**: 2026-02-15 (Sunday afternoon)
**Version**: v0.9.26
**Status**: ✅ Issue #32 COMPLETE, ready for productive week
**Next**: Monday client check-in → decide priority
