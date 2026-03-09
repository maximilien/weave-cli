# Weekend + Monday Plan: Feb 22-24, 2026

**Status**: ✅ v0.9.30 shipped, Issue #42 CLOSED, awaiting Client0 feedback
**Current Version**: v0.9.30 (released Thu Feb 20)
**Last Updated**: 2026-02-21 (Friday PM)
**Schedule**: Flexible 2-3 hours/day, responsive to Client0

---

## 🎉 This Week Victory (Feb 17-21)

**Shipped**:
- v0.9.27: Parallel processing + PDF storage (Mon)
- v0.9.28: 5 Client0 ingestion improvements (Tue-Wed)
- v0.9.30: Image metadata + local storage fixes (Thu-Fri)

**Impact**: Client0's 255-line bash wrapper → ~10 lines. All blockers resolved.

**Issue #42**: ✅ CLOSED with Client0 confirmation
- Fix 1: `image_url` field in Milvus query results
- Fix 2: Local storage initialization (`--local-storage-path`)

---

## 🎯 Weekend + Monday Goals

**Mode**: Maintenance + preparation while awaiting Client0 feedback

**Primary Goals**:
1. Keep responsive to Client0 (same-day turnaround if needed)
2. Small quality improvements (no new features)
3. Prep for next week (when Client0 feedback arrives)

**NOT Doing**:
- ❌ Issue #38 (`create-batch`) - wait for Client0 needs/feedback
- ❌ Issue #39 (`docs status`) - wait for Client0 needs/feedback
- ❌ Large refactors or new features

---

## 📅 Daily Breakdown (2-3 hours/day)

### Saturday, Feb 22 (2-3 hours)

**Focus**: Code quality & test coverage

**Options** (pick 1-2):
1. **Test coverage improvements** (2 hours)
   - Add unit tests for storage package (local.go, minio.go)
   - Cover the Issue #42 fix paths (LocalPath config chain)
   - Target: storage package 60%+ coverage

2. **Documentation polish** (1.5 hours)
   - Update VDB_SUPPORT_MATRIX.md with local storage status
   - Add troubleshooting section for image storage
   - Link Issue #42 fixes in relevant docs

3. **Minor bug fixes** (1 hour)
   - Review open issues for quick wins
   - Check for TODO comments in recent code
   - Audit error messages for clarity

**Deliverable**: Small PR with tests OR docs update

---

### Sunday, Feb 23 (2-3 hours)

**Focus**: Planning & organization

**Tasks**:
1. **Next week planning** (1.5 hours)
   - Create `WEEK_MAR_03-07_PLAN.md` (assuming Client0 feedback Mon-Tue)
   - Design Issue #38 checkpoint format (if still needed)
   - Research batch ingestion patterns

2. **Issue triage** (1 hour)
   - Review all open issues (#8, #11, #12, #14, #15, #16, #21)
   - Update priorities based on Client0 production use
   - Close stale issues or mark as backlog

3. **Dependency updates** (30 min)
   - Check `go.mod` for outdated deps
   - Security audit with `go mod audit`
   - Update if safe

**Deliverable**: Next week plan ready + clean issue backlog

---

### Monday, Feb 24 (2-3 hours)

**Focus**: Client0 response + prep

**Morning** (1 hour):
- Check for Client0 feedback on v0.9.30
- Check GitHub issues for new reports
- Respond to any questions

**Afternoon** (1-2 hours):

**If Client0 has new issues**:
- Debug + fix immediately (same-day turnaround)
- Hotfix release if needed

**If Client0 quiet**:
- Code cleanup (2 hours)
  - Remove debug logging from recent fixes
  - Simplify config chain (factory.go readability)
  - Add comments to Issue #42 fix locations

**Deliverable**: Ready to pivot based on Client0 needs

---

## 🔍 Optional Deep Dives (If Time)

### Image Storage Architecture Review (2-3 hours)

**Goal**: Document the full image storage flow for future maintainers

**Tasks**:
1. Create `docs/architecture/IMAGE_STORAGE.md`
2. Diagram: CLI flag → factory → adapter → storage
3. Document when to use local vs MinIO vs S3
4. Add decision tree for storage selection

**Why now**: Recent Issue #42 showed config chain complexity

---

### Test Coverage Deep Dive (3-4 hours)

**Goal**: Reach 60%+ coverage in critical paths

**Packages to target**:
- `src/pkg/storage/` (currently 0%)
- `src/pkg/vectordb/factory.go` (config chain)
- `src/pkg/vectordb/milvus/adapter.go` (storage init)

**Approach**:
- Add table-driven tests for config variations
- Mock storage backends for adapter tests
- Integration test for local storage end-to-end

---

## 📊 Success Metrics

**Client Support**:
- [ ] Response time <2 hours for all Client0 messages
- [ ] Same-day resolution if critical issue appears

**Code Quality**:
- [ ] All lint checks passing
- [ ] No new TODO comments without issue links
- [ ] Test coverage ≥50% (if working on tests)

**Planning**:
- [ ] Next week plan ready by Sunday EOD
- [ ] Issue backlog triaged and prioritized

---

## 🚨 Client0 Response Protocol

### If Client0 Reports Issue

**Critical (blocks production)**:
- Drop everything
- Debug immediately
- Hotfix within 4 hours
- Release v0.9.31

**Medium (workaround available)**:
- Acknowledge within 1 hour
- Fix within 24 hours
- Release v0.9.31 next day

**Low (enhancement/question)**:
- Respond within 2 hours
- Assess priority
- Add to next week plan if feature request

### If Client0 Requests Feature

**Small (<2 hours)**:
- Implement same day if time permits
- Release v0.9.31

**Medium (<1 day)**:
- Start over weekend if clear requirements
- Release early next week

**Large (>1 day)**:
- Create detailed design doc
- Get Client0 feedback on design
- Schedule for next week

---

## 📌 Key Context

**What Client0 is doing**:
- Full ingestion of 9 PDFs with local storage
- Testing 6,591-image collection retrieval
- Likely building UI/UX on top of weave-cli
- May discover edge cases or new needs

**What we're ready for**:
- Image ingestion issues (storage, metadata)
- Query performance issues
- Batch ingestion needs (Issue #38)
- Observability needs (Issue #39)

**What to watch**:
- GitHub Issue #42 (closed but may reopen)
- New GitHub issues from Client0
- Email/Slack if configured

---

## 🎯 Monday Decision Tree

```
Monday AM: Check Client0 status
│
├─ New critical issue?
│  └─ → Drop plans, fix immediately
│
├─ New feature request?
│  └─ → Assess scope, start if small
│
├─ Questions/clarifications?
│  └─ → Respond, continue with plan
│
└─ All quiet?
   └─ → Execute code cleanup plan
```

---

## 📝 Notes

**Why this approach**:
- Client0 is in production testing phase
- Feedback is most valuable now
- Responsive support builds trust
- Small improvements > big features right now

**What we're avoiding**:
- Starting Issue #38/39 without Client0 confirmation they need it
- Large refactors that could introduce bugs
- Breaking changes to CLI interface

**Flexibility**:
- Plans are guidelines, not commitments
- Client0 needs always take priority
- OK to do nothing if everything is stable

---

**Prepared**: 2026-02-21 (Friday PM)
**Mode**: Responsive + preparation
**Mindset**: Client success > feature velocity 🎯
