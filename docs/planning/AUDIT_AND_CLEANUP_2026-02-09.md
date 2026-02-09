# Weave CLI Audit and Cleanup Plan - Feb 9, 2026

**Date**: February 9, 2026
**Version**: v0.9.18
**Status**: Ready for Cleanup

---

## Executive Summary

**Current State**:
- ✅ v0.9.18 shipped (Issues #7 & #9 complete)
- ✅ 28 tests passing (re-embedding feature)
- ✅ 8 open issues (1 stale, 7 legitimate)
- ⚠️ 152+ documentation files (many outdated)
- ⚠️ ROADMAP.md outdated (shows v0.7.1, we're at v0.9.18)
- ⚠️ CHANGELOG missing v0.9.18 entry

**Audit Findings**:
1. **Issues**: 1 stale issue (#17), others are valid
2. **Documentation**: 40+ archived files, 30+ planning docs, many outdated
3. **ROADMAP**: Last updated Dec 2025, missing v0.8.x-v0.9.x releases
4. **CHANGELOG**: Missing v0.9.18 entry (just shipped)

---

## Part 1: Issue Audit

### weave-cli Repository Issues

**Open Issues (8 total)**:

1. **#21: Test image ingestion across all VDBs** (enhancement)
   - Status: VALID - legitimate testing need
   - Priority: Medium
   - Note: Also exists as Client0 issue #1

2. **#17: Update videos and presentations** (documentation) ⚠️
   - Status: **STALE** - Created long ago, no recent activity
   - Recommendation: **CLOSE or update with specific requirements**
   - Videos/presentations may no longer be relevant

3. **#16: Audit code and tests, prep for v1.0** (chore)
   - Status: VALID - ongoing work
   - Priority: High
   - Action: Keep open, update with v0.9.18 progress

4. **#15: Update docs** (documentation)
   - Status: VALID - always relevant
   - Priority: Medium
   - Action: Keep open, consider specific sub-tasks

5. **#14: Add different agents with easy configs** (feature)
   - Status: VALID - future enhancement
   - Priority: Low (no immediate Client0 need)
   - Action: Keep open

6. **#12: Include tips on command -h** (enhancement)
   - Status: VALID - UX improvement
   - Priority: Low
   - Action: Keep open

7. **#11: Streamline commands and shortcuts** (chore)
   - Status: VALID - ongoing work
   - Priority: Medium
   - Action: Keep open

8. **#8: Test extract various PDFs** (test, enhancement)
   - Status: VALID - legitimate testing need
   - Priority: Medium
   - Action: Keep open

**Issue Recommendations**:
- ✅ Keep: #21, #16, #15, #14, #12, #11, #8
- ❌ Close or update: #17 (stale)

### Client0 (AuctionsMax.ai) Issues

**Status Summary**:
- ✅ **7 CLOSED** (Issues #2-9) - All resolved!
- ⏳ **2 OPEN** (Issues #1, #10)

**Open Issues**:

1. **#10: OSS stack quick-start template** (enhancement)
   - Status: VALID - lower priority
   - Priority: Low
   - Action: Keep open, address after v0.9.19

2. **#1: Test image ingestion across all VDBs** (enhancement)
   - Status: VALID - duplicates weave-cli #21
   - Priority: Medium
   - Action: Keep open, coordinate with weave-cli #21

**Recently Closed** (Feb 9, 2026):
- ✅ #9: Comparison reports (v0.9.18)
- ✅ #8: Batch re-embedding (v0.9.17)
- ✅ #7: Ollama auto-discovery (v0.9.18)
- ✅ #6: Auto-detect dimensions (v0.9.16)
- ✅ #5, #4, #3, #2: Various bugs (v0.9.16)

**Client0 Issue Action**: No cleanup needed, all closed issues are legitimate

---

## Part 2: Documentation Audit

### Documentation Structure (152 files)

**Category Breakdown**:

#### 1. Archive (40 files) ✅
**Location**: `docs/archive/`
**Status**: Well-organized
**Action**: No cleanup needed

Files properly archived:
- Old planning docs (2025-11-28 to 2026-01-23)
- Session summaries
- Completed work plans (Neo4j, Chroma, Qdrant)
- Historical bug fix plans

#### 2. Planning (30 files)
**Location**: `docs/planning/`
**Status**: ⚠️ Needs cleanup

**Active Planning Docs** (Keep):
- `THIS_WEEK_PLAN.md` (Feb 10-14, 2026) ✅
- `WEEKEND_AND_NEXT_WEEK.md` (Feb 2026) ✅
- `BATCH_REEMBEDDING_SPEC.md` (v0.9.17) ✅
- `VDB_LIFECYCLE_MANAGEMENT.md` (v0.9.12) ✅

**Outdated Planning Docs** (Archive or delete):
- `DAILY_PLAN_2026-02-03.md` ⚠️ (6 days old, superseded)
- `WEEK_PLAN_2026-02-03.md` ⚠️ (superseded by THIS_WEEK_PLAN.md)
- `WEEK_PLAN_2026-02-05.md` ⚠️ (superseded)
- `WEEK_OF_2026-01-13.md` ⚠️ (4 weeks old)
- `WORK_PLAN_JAN_29_31.md` ⚠️ (completed)
- `WEEKEND_PREP_JAN_29.md` ⚠️ (completed)

**Strategic Planning Docs** (Keep):
- `V1_0_ROADMAP.md` ✅
- `OPTION_1_NEW_FEATURES.md` through `OPTION_5_COMMUNITY.md` ✅
- `OSS_EXECUTION_PLAN.md` ✅
- `OSS_FEATURES_CLIENT0.md` ✅

**Specialized Planning** (Keep):
- Agent evaluation system docs (6 files) ✅
- Multi-agent orchestration ✅
- Integration test plan ✅

**Recommendation**:
```bash
# Archive old weekly/daily plans
mv docs/planning/DAILY_PLAN_2026-02-03.md docs/archive/planning/
mv docs/planning/WEEK_PLAN_2026-02-03.md docs/archive/planning/
mv docs/planning/WEEK_PLAN_2026-02-05.md docs/archive/planning/
mv docs/planning/WEEK_OF_2026-01-13.md docs/archive/planning/
mv docs/planning/WORK_PLAN_JAN_29_31.md docs/archive/planning/
mv docs/planning/WEEKEND_PREP_JAN_29.md docs/archive/planning/
mv docs/planning/SESSION_SUMMARY_2026-01-12.md docs/archive/planning/
mv docs/planning/NEXT_SESSION_VDB_TESTS.md docs/archive/planning/
```

#### 3. VDB-Specific Docs (48 files)
**Location**: `docs/{vdb_name}/`
**Status**: ✅ Well-organized

**VDBs with docs**:
- Weaviate (2 files) ✅
- Milvus (4 files) ✅
- MongoDB (3 files) ✅
- Neo4j (4 files) ✅
- Qdrant (2 files) ✅
- Chroma (2 files) ✅
- Supabase (6 files) ✅
- Elasticsearch (6 files) ✅
- OpenSearch (2 files) ✅
- Pinecone (2 files) ✅
- LanceDB (1 file) ✅

**Action**: No cleanup needed

#### 4. Releases (10 files)
**Location**: `docs/releases/`
**Status**: ⚠️ Incomplete

**Existing Release Docs**:
- RELEASE_CHECKLIST.md ✅
- RELEASE_v0.2.7.md, v0.2.8.md ✅
- RELEASE_v0.3.9.md through v0.3.12.md ✅
- RELEASE_v0.7.0.md, v0.7.1.md, v0.7.2.md ✅
- RELEASE_v0.9.1.md ✅
- RELEASE_v0.9.13.md ✅
- RELEASE_v0.9.15.md ✅

**Missing Release Docs**:
- RELEASE_v0.9.16.md ❌ (auto-dimension detection)
- RELEASE_v0.9.17.md ❌ (batch re-embedding)
- RELEASE_v0.9.18.md ❌ (Ollama discovery, comparison reports)

**Recommendation**: Create missing release docs (see Part 4)

#### 5. Guides (4 files)
**Location**: `docs/guides/`
**Status**: ✅ Good

Files:
- BATCH_DOCS_CREATION.md ✅
- DEMO.md ✅
- VECTOR_DB_ABSTRACTION.md ✅
- WEAVE_CLI_AI.md ✅

**Action**: No cleanup needed

#### 6. Integrations (3 files)
**Location**: `docs/integrations/`
**Status**: ✅ Good

Files:
- AIRFLOW.md ✅
- ARGO_WORKFLOWS.md ✅
- GITHUB_ACTIONS.md ✅

**Action**: No cleanup needed

#### 7. Root Docs (20 files)
**Location**: `docs/`
**Status**: ⚠️ Some outdated

**Current Files**:
- README.md ✅
- ROADMAP.md ⚠️ (Last updated Dec 2025, shows v0.7.1)
- TODO_AUDIT.md ✅ (Dec 2025, still accurate)
- ARCHITECTURE.md ✅
- USER_GUIDE.md ✅
- TEST_GUIDE.md ✅
- VDB_SUPPORT_MATRIX.md ✅
- VDB_SUPPORT.md ✅
- VDB_NAMING_CONVENTION.md ✅
- PRODUCTION_READY.md ✅
- PRESENTATION.md ⚠️ (may be outdated)
- AGENT_MANAGEMENT.md ✅
- CONSULTATION_GUIDE.md ✅
- DEMO_CONFIG_FIX.md ⚠️ (specific fix, could archive)
- ENHANCED_IMAGE_SEARCH.md ✅
- OBSERVABILITY.md ✅
- SHELL_COMPLETION.md ✅
- TIMEOUT_CONFIGURATION.md ✅
- WEAVE_MCP.md ✅

**Action**: Update ROADMAP.md, consider archiving DEMO_CONFIG_FIX.md

---

## Part 3: ROADMAP Update Needed

### Current Issues with ROADMAP.md

**Problems**:
1. Last updated: 2025-12-03 (2 months ago)
2. Shows "Current Version: v0.7.1" (we're at v0.9.18)
3. Missing versions v0.8.0 through v0.9.18
4. Doesn't reflect Client0 work

**Missing Version History**:
- v0.9.18: Ollama auto-discovery + comparison reports (Feb 9, 2026)
- v0.9.17: Batch re-embedding feature (Feb 6, 2026)
- v0.9.16: Auto-dimension detection (Feb 2026)
- v0.9.15: Production hardening (Jan 2026)
- v0.9.13: VDB lifecycle management (Jan 2026)
- v0.9.1: Previous features (Jan 2026)
- v0.8.x: Unknown (need to check git history)

**Action**: Update ROADMAP.md (see Part 4)

---

## Part 4: CHANGELOG Update Needed

### Missing Entry: v0.9.18

**Problem**: CHANGELOG.md shows v0.9.17 as latest, but we just shipped v0.9.18

**What's Missing**:
```markdown
## [0.9.18] - 2026-02-09

### Added - Client0 Features 🎉

- **Issue #7: Ollama Auto-Discovery**
  - Automatically discovers local Ollama embedding models
  - Integrates into `weave config agents` wizard
  - Shows setup instructions if Ollama not found
  - Lists available models with dimensions

- **Issue #9: Embedding Model Comparison Reports**
  - Compare search quality across 2+ collections
  - Generate markdown or JSON reports
  - Shows relevance scores, rankings, latency
  - Perfect for A/B testing embedding models

### Testing

- 5 Ollama client unit tests (all passing)
- Comparison report generation tested with multiple collections
- Integration with existing VDB commands

### Documentation

- Updated agent wizard with Ollama discovery flow
- Comparison report command help and examples
- Client0 update documents for PM meeting
```

**Action**: Add v0.9.18 entry to CHANGELOG.md

---

## Part 5: Cleanup Actions

### Priority 1: Critical Updates (Do Now)

1. **Update CHANGELOG.md**:
   - Add v0.9.18 entry (Ollama discovery + comparison reports)
   - Backfill v0.9.16 if missing

2. **Update ROADMAP.md**:
   - Change "Current Version: v0.9.18"
   - Add version history for v0.8.x-v0.9.18
   - Update "Current Focus" section
   - Reflect Client0 work and OSS embedding focus

3. **Close Stale Issue**:
   - Close weave-cli #17 (videos/presentations) OR
   - Update with specific requirements if still needed

### Priority 2: Documentation Cleanup (This Week)

4. **Archive Old Planning Docs** (6 files):
   ```bash
   mv docs/planning/DAILY_PLAN_2026-02-03.md docs/archive/planning/
   mv docs/planning/WEEK_PLAN_2026-02-03.md docs/archive/planning/
   mv docs/planning/WEEK_PLAN_2026-02-05.md docs/archive/planning/
   mv docs/planning/WEEK_OF_2026-01-13.md docs/archive/planning/
   mv docs/planning/WORK_PLAN_JAN_29_31.md docs/archive/planning/
   mv docs/planning/WEEKEND_PREP_JAN_29.md docs/archive/planning/
   ```

5. **Create Missing Release Docs**:
   - RELEASE_v0.9.16.md (auto-dimension detection)
   - RELEASE_v0.9.17.md (batch re-embedding)
   - RELEASE_v0.9.18.md (Ollama + comparison reports)

6. **Archive Specific Fix Docs**:
   ```bash
   mv docs/DEMO_CONFIG_FIX.md docs/archive/
   ```

### Priority 3: Ongoing Maintenance (As Needed)

7. **Update VDB_SUPPORT_MATRIX.md**:
   - Confirm all VDB status (production/beta)
   - Add any new features since last update

8. **Review PRESENTATION.md**:
   - Check if content is current
   - Archive if outdated

9. **Monitor docs/planning/**:
   - Archive old weekly plans after 2 weeks
   - Keep strategic planning docs

---

## Part 6: Cleanup Script

```bash
#!/bin/bash
# Weave CLI Cleanup Script - Feb 9, 2026

echo "🧹 Starting Weave CLI cleanup..."

# Archive old planning docs
echo "📦 Archiving old planning docs..."
mkdir -p docs/archive/planning
mv docs/planning/DAILY_PLAN_2026-02-03.md docs/archive/planning/
mv docs/planning/WEEK_PLAN_2026-02-03.md docs/archive/planning/
mv docs/planning/WEEK_PLAN_2026-02-05.md docs/archive/planning/
mv docs/planning/WEEK_OF_2026-01-13.md docs/archive/planning/
mv docs/planning/WORK_PLAN_JAN_29_31.md docs/archive/planning/
mv docs/planning/WEEKEND_PREP_JAN_29.md docs/archive/planning/

# Archive specific fix docs
echo "📦 Archiving specific fix docs..."
mv docs/DEMO_CONFIG_FIX.md docs/archive/

# Commit cleanup
echo "💾 Committing cleanup..."
git add docs/archive/planning/*.md
git add docs/archive/DEMO_CONFIG_FIX.md
git commit -m "chore: archive old planning docs and specific fix docs

- Move 6 old planning docs to archive/planning/
- Archive DEMO_CONFIG_FIX.md (specific fix, no longer needed in root)
- Cleanup part of Feb 9, 2026 audit"

echo "✅ Cleanup complete!"
echo ""
echo "📝 Next steps:"
echo "1. Update CHANGELOG.md with v0.9.18"
echo "2. Update ROADMAP.md (version + history)"
echo "3. Consider closing issue #17 (stale)"
```

---

## Part 7: Documentation Strategy Going Forward

### Weekly Planning Lifecycle

**Current Week**:
- Keep: `THIS_WEEK_PLAN.md`
- Delete after 2 weeks, archive if significant

**Weekly Archive Policy**:
- Week plans: Archive after 2 weeks
- Daily plans: Archive after 1 week
- Strategic plans: Never archive (keep indefinitely)

### Release Documentation

**For Every Release**:
1. Update CHANGELOG.md (immediately after tagging)
2. Create RELEASE_vX.X.X.md in `docs/releases/`
3. Update ROADMAP.md version number
4. Update VDB_SUPPORT_MATRIX.md if VDB changes

### Issue Hygiene

**Monthly Review**:
- Close stale issues (no activity in 60 days)
- Update issue labels
- Consolidate duplicate issues

**Weekly Review**:
- Triage new issues
- Update issue status
- Link PRs to issues

---

## Summary

**Audit Results**:
- ✅ Code quality: Excellent (28 tests passing, TODO audit clean)
- ✅ Issue tracking: Good (1 stale issue, rest valid)
- ⚠️ Documentation: Needs cleanup (old planning docs)
- ⚠️ CHANGELOG: Missing v0.9.18 entry
- ⚠️ ROADMAP: Outdated (shows v0.7.1, missing 8 versions)

**Immediate Actions** (Priority 1):
1. Update CHANGELOG.md (add v0.9.18)
2. Update ROADMAP.md (version + history)
3. Close or update issue #17

**This Week** (Priority 2):
4. Archive 6 old planning docs
5. Create 3 missing release docs
6. Archive DEMO_CONFIG_FIX.md

**Next Session** (Priority 3):
7. Review PRESENTATION.md
8. Update VDB_SUPPORT_MATRIX.md
9. Implement weekly cleanup policy

---

**Prepared by**: Claude Code
**Date**: February 9, 2026
**Next Review**: February 14, 2026 (end of week)
