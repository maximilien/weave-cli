# Issue Cleanup and Planning - Feb 27, 2026

**Status**: In Progress
**Last Updated**: 2026-02-27 16:30 PST

---

## Summary

Major cleanup needed after today's bug fixing session. Fixed 5+ critical bugs but haven't created corresponding GitHub issues yet.

## Issues Fixed Today (Need GitHub Issues)

### 1. Port-forward Service Discovery (#43 - CONFLICT!)
**Commit**: `87c4dd9`
**File**: `src/cmd/stack/port_forward.go`
**Issue**: Used cluster name instead of Helm release name
**Fix**: Load stack config, use `{helmReleaseName}-weave-stack-{service}`
**Status**: ✅ Fixed
**GitHub Issue**: Need to create (but #43 is taken by backup/restore!)

### 2. Status Command Pod Discovery (#44)
**Commit**: `3ce8295`
**File**: `src/cmd/stack/status.go`
**Issue**: Used cluster name in label selector
**Fix**: Use Helm release name for `app.kubernetes.io/instance`
**Status**: ✅ Fixed
**GitHub Issue**: Need to create

### 3. Logs Command Label Selector
**Commit**: `08713b9`
**File**: `src/cmd/stack/logs.go`
**Issue**: Used wrong label (`app.kubernetes.io/name=milvus`)
**Fix**: Changed to `app=milvus`
**Status**: ✅ Fixed
**GitHub Issue**: Need to create

### 4. Ingestion Milvus Address Field
**Commit**: `813db1f`
**File**: `src/pkg/stack/ingest.go`
**Issue**: Used `URL` field but Milvus expects `Address`
**Fix**: Changed to `Address` field
**Status**: ✅ Fixed
**GitHub Issue**: Need to create

### 5. Ingestion Port-forward Service Name (#45)
**Commit**: `291a056`
**File**: `src/pkg/stack/ingest.go`
**Issue**: StartMilvusPortForward() used cluster name
**Fix**: Load stack config, use Helm release name
**Status**: ✅ Fixed
**GitHub Issue**: Need to create

### 6. Vector Dimensions Warning Suppression
**Commit**: `6a79068`
**File**: `src/pkg/config/validation.go`
**Issue**: Warning shown even with correct weave-stack.yaml
**Fix**: Skip warning when weave-stack.yaml exists
**Status**: ✅ Fixed
**GitHub Issue**: Need to create

---

## Critical Bug Found (Need Issue)

### Stack Ingestion Pipeline Not Calling VDB (#45 candidate)
**Commits**: `ab38685` (debug logging)
**Files**: `src/pkg/stack/ingest.go`
**Symptom**: Ingestion reports success but no data in Milvus
**Root Cause**: `pipeline.ProcessFiles()` not calling VDB client methods
**Evidence**:
- Port-forward works ✅
- VDB client created (*milvus.Adapter) ✅
- Config correct (localhost:19530) ✅
- Processing reports "Documents created: 1" ✅
- **Milvus logs show ZERO activity** ❌

**Status**: 🔴 Critical - Debug logging added, need to fix pipeline
**GitHub Issue**: Need to create (#45)

---

## Open Issues Review

### High Priority (Client0 Requests)

#### Issue #43: Backup/Restore Commands ⭐
**Type**: Feature Request
**Status**: Open (just created by Client0)
**Priority**: HIGH - Client0 requested
**Description**: Add `weave cols backup` and `weave cols restore` commands
**Scope**:
- Backup collection to file (JSON/parquet)
- Restore from backup file
- Support all VDBs (Milvus, Qdrant, Weaviate, Chroma)

**Recommendation**: Add to Week 11 plan (Mar 3-7)

#### Issue #39: Ingestion Status Dashboard
**Type**: Feature Request
**Status**: Open
**Command**: `weave docs status`
**Priority**: MEDIUM
**Recommendation**: Defer to Phase 3 (monitoring)

#### Issue #38: Batch File Ingestion
**Type**: Feature Request
**Status**: Open
**Command**: `weave docs create-batch`
**Priority**: MEDIUM
**Recommendation**: Could combine with #39, defer to Phase 3

### Medium Priority

#### Issue #21: Image Ingestion Cross-VDB Testing
**Type**: Test
**Status**: Open
**Priority**: MEDIUM
**Recommendation**: Good task for stabilization week

#### Issue #16: Audit for v1.0
**Type**: Chore
**Status**: Open
**Priority**: HIGH (for v1.0)
**Recommendation**: Schedule for Phase 3 completion

#### Issue #15: Update Docs
**Type**: Documentation
**Status**: Open
**Priority**: MEDIUM
**Recommendation**: Ongoing, update with each release

### Low Priority

#### Issue #14: Different Agents with Easy Configs
**Type**: Feature
**Status**: Open
**Priority**: LOW
**Recommendation**: Phase 4 (advanced features)

#### Issue #12: Tips on -h Command
**Type**: Enhancement
**Status**: Open
**Priority**: LOW
**Recommendation**: Nice-to-have, Phase 3

#### Issue #11: Streamline Commands
**Type**: Chore
**Status**: Open
**Priority**: LOW
**Recommendation**: Part of v1.0 cleanup (#16)

#### Issue #8: PDF Extraction Testing
**Type**: Test
**Status**: Open
**Priority**: LOW
**Recommendation**: Defer to Phase 3

---

## Actions Required

### Immediate (Tonight/Tomorrow)

1. **Create GitHub Issues for Today's Fixes**
   - [ ] Port-forward service discovery (need new number, #43 taken)
   - [ ] Status command pod discovery
   - [ ] Logs command label selector
   - [ ] Ingestion address field
   - [ ] Ingestion port-forward
   - [ ] Vector dimensions warning

2. **Create Issue #45: Stack Ingestion Pipeline Bug** 🔴
   - Critical bug
   - Include debug findings
   - Include Milvus log evidence
   - Tag as `bug`, `critical`, `weave-stack`

3. **Review Issue #43: Backup/Restore**
   - Understand Client0's requirements
   - Create design doc if needed
   - Add to Week 11 plan

### This Week (Feb 27 - Mar 2)

4. **Fix Stack Ingestion Pipeline** (#45)
   - Debug why ProcessFiles doesn't call VDB
   - Fix pipeline VDB interaction
   - Test end-to-end workflow
   - Update quickstart guide

5. **Clean Up Open Issues**
   - Close or update stale issues
   - Add labels (priority, phase, etc.)
   - Link related issues

6. **Update Planning Docs**
   - Add backup/restore to roadmap
   - Update Phase 2 timeline
   - Document v0.10.3 scope

---

## Recommended Issue Numbering

Since #43 is taken by backup/restore, use these numbers:

- **#44**: ✅ Reserved (status pod discovery)
- **#45**: Stack ingestion pipeline bug (CRITICAL)
- **#46**: Port-forward service discovery fix
- **#47**: Logs command label selector fix
- **#48**: Ingestion Milvus address field fix
- **#49**: Ingestion port-forward fix
- **#50**: Vector dimensions warning suppression

---

## Week 11 Plan Update (Mar 3-7)

### Goals
1. Complete Phase 2 Week 1 (EKS schema + cluster creation)
2. Start backup/restore feature (#43)
3. Fix stack ingestion pipeline (#45)

### Tasks

#### Phase 2: Cloud Deployment
- [ ] EKS schema design (2 days)
- [ ] EKS cluster creation (2 days)
- [ ] Initial testing (1 day)

#### Client0 Features
- [ ] Fix stack ingestion pipeline (#45) - 1 day
- [ ] Design backup/restore (#43) - 1 day
- [ ] Implement backup command - 2 days
- [ ] Implement restore command - 2 days

#### Maintenance
- [ ] Create GitHub issues for fixed bugs
- [ ] Update documentation
- [ ] Test coverage for fixes

---

## Next Steps

1. **Create missing GitHub issues** (30 min)
2. **Review Issue #43 in detail** (15 min)
3. **Fix stack ingestion pipeline** (2-4 hours)
4. **Update planning docs** (30 min)
5. **Commit and push** (15 min)

---

## Notes

- Today was VERY productive - fixed 6 major bugs!
- Stack ingestion issue is critical but well-diagnosed
- Client0's backup/restore request is well-timed
- Need to balance Phase 2 work with Client0 features

---

**Updated**: 2026-02-27 16:30 PST
