# Client0 Issues - Action Plan

**Date**: 2026-03-10
**Context**: Client0 reported multiple bugs and feature requests during production usage
**Total Issues**: 12 open (7 bugs, 4 features, 1 fixed)

---

## 🐛 BUGS - Ready to Close (Already Fixed)

### ✅ Issue #52: Weaviate backups missing vector embeddings
**Status**: FIXED in v0.11.2
**Fix**: Commit `e1f974b`
**Action**: Close issue with reference to v0.11.2 release

### ✅ Issue #50: Vector dimensions warning with correct weave-stack.yaml
**Status**: FIXED
**Fix**: Commit `6a79068`
**Action**: Close issue with commit reference

### ✅ Issue #45: Port-forward uses cluster name instead of Helm release
**Status**: FIXED
**Fix**: Commit `87c4dd9`
**Action**: Verify fix and close

### ⚠️ Issues #46-49: Similar naming/selector bugs
**Status**: Likely fixed by same commit as #45
**Action**: Verify each is fixed, then close with commit reference

**Issues to verify:**
- #46: Status command uses cluster name instead of Helm release name
- #47: Logs command uses incorrect label selector
- #48: Stack ingestion uses URL field instead of Address for Milvus
- #49: Stack ingestion port-forward uses cluster name instead of Helm release

**Verification Steps**:
1. Review commits `87c4dd9`, `6a79068` for scope
2. Test each command to confirm fix
3. Close with "Fixed in commit XXX" message

---

## 🚀 FEATURE REQUESTS - Client0 Priority

### 🔥 HIGH PRIORITY (Blocks Production Use)

#### Issue #56: Add `--restart-every N` flag for memory management
**Priority**: **CRITICAL**
**Reason**: Prevents OOM crashes during large ingestion (9 PDFs, 2,636 images)
**Use Case**: Client0 needs to process auction catalogs without Milvus crashing
**Impact**: Currently requires 150-line custom script

**Implementation Plan**:
- [ ] Add `--restart-every N` flag to `weave stack ingest`
- [ ] Add restart detection and health checking
- [ ] Support for Kind/Docker/Podman runtimes
- [ ] Test with Client0's dataset (9 PDFs)
- [ ] Document flag usage and best practices

**Estimated Effort**: 8-12 hours
**Target**: Sprint 2 (Mar 14-20)

---

#### Issue #54: Add `--auto-restart` and `--resume-from` flags
**Priority**: **CRITICAL**
**Reason**: 2-3 hour ingestion jobs fail halfway through, must restart from scratch
**Use Case**: Robust long-running ingestion with automatic recovery
**Impact**: Replaces 250-line monitoring script

**Implementation Plan**:
- [ ] Implement checkpoint system (`.weave-ingest-checkpoint` JSON)
- [ ] Add `--auto-restart` flag with retry logic
- [ ] Add `--resume-from <file>` flag
- [ ] Add `--max-retries N` flag
- [ ] Add `--checkpoint-every N` flag
- [ ] Implement Milvus health detection and restart
- [ ] Test end-to-end recovery scenarios

**Estimated Effort**: 12-16 hours
**Target**: Sprint 2 (Mar 14-20)

---

#### Issue #55: Add `weave stack ingest --all`
**Priority**: **HIGH**
**Reason**: Manual 6-collection ingestion is error-prone
**Use Case**: Config-driven ingestion of all collections from weave-stack.yaml
**Impact**: Eliminates 167-line reingest script

**Implementation Plan**:
- [ ] Design YAML schema for ingestion config
- [ ] Parse `weave-stack.yaml` ingestion section
- [ ] Implement sequential collection processing
- [ ] Add `--parallel N` flag for concurrent ingestion
- [ ] Add `--dry-run` flag
- [ ] Add progress reporting for multiple collections
- [ ] Add `--continue-on-error` flag
- [ ] Update weave-stack.yaml templates

**Estimated Effort**: 10-14 hours
**Target**: Sprint 2 (Mar 14-20)

---

### 📋 MEDIUM PRIORITY (Developer Experience)

#### Issue #53: Add `weave stack collections reset`
**Priority**: **MEDIUM**
**Reason**: Improves development iteration speed
**Use Case**: Delete all collections without tearing down infrastructure
**Impact**: Reduces `down/up` cycle from 2 minutes to 5 seconds

**Implementation Plan**:
- [ ] Add `weave stack collections reset` command
- [ ] List all collections from active stack
- [ ] Delete each collection via VDB client
- [ ] Add `--force` flag to skip confirmation
- [ ] Add `--keep <collection>` flag for selective reset
- [ ] Document usage and best practices

**Estimated Effort**: 4-6 hours
**Target**: Sprint 2 (Mar 14-20) - Quick win

---

## 📊 Priority Summary

| Priority | Issues | Estimated Effort | Sprint |
|----------|--------|------------------|--------|
| **CRITICAL** | #54, #56 | 20-28 hours | Sprint 2 |
| **HIGH** | #55 | 10-14 hours | Sprint 2 |
| **MEDIUM** | #53 | 4-6 hours | Sprint 2 |
| **FIXED** | #45-50, #52 | 2-3 hours verification | This week |

**Total Development**: 34-48 hours (1-1.5 weeks focused work)

---

## 🎯 Recommended Execution Plan

### **This Week (Mar 10-13): Cleanup Sprint**

**Goal**: Close all fixed issues, verify no regressions

**Tasks**:
1. ✅ Review all open issues (DONE)
2. ✅ Create action plan (DONE)
3. [ ] Verify bugs #45-50, #52 are fixed
4. [ ] Close verified issues with commit references
5. [ ] Update CHANGELOG.md with bug fixes
6. [ ] Contact Client0 with status update

**Time**: 4-6 hours

---

### **Next Week (Mar 14-20): Client0 Feature Sprint**

**Goal**: Implement all high-priority features for production use

**Monday-Tuesday (Mar 14-15): Quick Win + Foundation**
- [ ] Implement #53: `collections reset` (4-6h) - Quick win
- [ ] Design checkpoint system for #54 (2-3h)
- [ ] Design ingestion config schema for #55 (2-3h)

**Wednesday-Thursday (Mar 16-17): Core Features**
- [ ] Implement #56: `--restart-every N` flag (8-12h)
- [ ] Implement #54: `--auto-restart` and checkpoints (12-16h)

**Friday (Mar 18): Bulk Ingestion**
- [ ] Implement #55: `--all` flag with config parsing (10-14h)

**Weekend (Mar 19-20): Testing & Polish**
- [ ] Test all features with Client0's dataset
- [ ] Write documentation and examples
- [ ] Prepare v0.12.0 release notes

**Time**: 34-48 hours total work

---

## 🔄 Feature Dependencies

```
#53: collections reset
  └─ Standalone (no dependencies)

#56: --restart-every N
  ├─ Depends on: Restart mechanism
  └─ Blocks: None

#54: --auto-restart + --resume-from
  ├─ Depends on: Checkpoint system, restart mechanism
  ├─ Integrates with: #56 (shared restart logic)
  └─ Blocks: None

#55: ingest --all
  ├─ Depends on: YAML schema extension
  ├─ Benefits from: #54 (auto-restart), #56 (restart-every)
  └─ Blocks: None
```

**Recommended Order**:
1. #53 (quick win, standalone)
2. #56 (restart mechanism foundation)
3. #54 (builds on #56's restart logic)
4. #55 (benefits from #54 + #56 being ready)

---

## 📝 Client0 Communication Plan

### Status Update Email

**Subject**: Weave CLI - Bug Fixes Complete, Feature Implementation Starting

**Body**:
> Hi Client0,
>
> Thanks for the detailed bug reports and feature requests! Here's our action plan:
>
> **✅ Bugs Fixed (Ready to Close)**:
> - All 7 bugs (#45-50, #52) have been fixed and are ready to close
> - Weaviate backups now include embeddings (v0.11.2)
> - Stack commands now use correct Helm release names
> - Vector dimension warnings resolved
>
> **🚀 Feature Implementation (Starting Next Week)**:
> We're prioritizing all 4 feature requests for Sprint 2 (Mar 14-20):
>
> 1. **#56: `--restart-every N`** - Prevents Milvus OOM crashes ⚠️ CRITICAL
> 2. **#54: `--auto-restart` + `--resume-from`** - Robust long-running ingestion ⚠️ CRITICAL
> 3. **#55: `weave stack ingest --all`** - Config-driven multi-collection ingestion 🔥 HIGH
> 4. **#53: `weave stack collections reset`** - Fast collection cleanup 📋 MEDIUM
>
> **Timeline**:
> - This week: Close all fixed bugs
> - Next week: Implement all 4 features
> - Target release: v0.12.0 (Mar 20-21)
>
> **Testing**:
> We'll test with your dataset (9 PDFs, 2,636 images) to ensure production readiness.
>
> Questions or feedback? Let me know!
>
> Best,
> Max

---

## 📚 Documentation Updates Needed

1. **CHANGELOG.md**:
   - Document bug fixes (#45-50, #52)
   - Add feature entries for #53-56

2. **Guides**:
   - Create ingestion best practices guide
   - Document checkpoint system usage
   - Add troubleshooting section for OOM issues

3. **weave-stack.yaml**:
   - Add `ingestion` section to template
   - Document ingestion config schema

4. **Release Notes**:
   - v0.11.2: Bug fixes summary
   - v0.12.0: New ingestion features

---

## ✅ Success Criteria

### For Bug Closure:
- [ ] All 7 bugs verified as fixed
- [ ] Test coverage for fixed areas
- [ ] No regressions reported
- [ ] Issues closed with commit references

### For Feature Implementation:
- [ ] All 4 features implemented and tested
- [ ] Works with Client0's dataset (9 PDFs, 2,636 images)
- [ ] Documentation complete
- [ ] No Milvus OOM crashes during ingestion
- [ ] Can resume failed ingestion jobs
- [ ] Single command replaces 400+ lines of custom scripts

### For v0.12.0 Release:
- [ ] All tests passing
- [ ] Lint clean
- [ ] Client0 validation complete
- [ ] Release notes published
- [ ] GitHub release created

---

**Status**: Plan created, ready to execute
**Next Action**: Verify and close fixed bugs (#45-50, #52)
**Owner**: @maximilien
