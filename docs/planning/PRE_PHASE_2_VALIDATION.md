# Pre-Phase 2 Validation Checklist

**Date**: Feb 24, 2026
**Goal**: Validate Phase 1 local runtime support before starting Phase 2 cloud deployments

---

## Local Runtime Support - Must Validate First

Before jumping into EKS/GKE (Phase 2), we need to ensure our local runtime support is solid.

### Kind + Podman (Primary OSS Stack)

**Status**: ✅ Tested and working (v0.10.0)

- [x] Kind cluster creation with podman
- [x] Helm deployment to Kind
- [x] Pod monitoring and health checks
- [x] Port forwarding
- [x] Log streaming
- [x] Ingestion pipeline
- [x] Integration tests passing
- [x] Documentation complete

**Evidence**: All Phase 1 features tested with `weave stack up --runtime kind`

### Kind + Docker (Fallback)

**Status**: ⚠️ Should test (not critical but good to verify)

- [ ] Kind cluster creation with docker
- [ ] Helm deployment to Kind
- [ ] Basic smoke test (deploy + query)

**Time**: 10 minutes

### Minikube + Podman

**Status**: ⚠️ NEEDS TESTING (implemented but not validated)

**Code exists**:
- ✅ `CreateMinikubeCluster()` in src/pkg/stack/cluster.go
- ✅ MinikubeConfig in types
- ✅ Template support for minikube

**Not validated**:
- [ ] Minikube cluster creation
- [ ] Minikube driver compatibility (docker/podman/kvm2)
- [ ] Helm deployment to Minikube
- [ ] Pod monitoring
- [ ] Basic smoke test

**Time**: 30 minutes

### Minikube + Docker

**Status**: ⚠️ NEEDS TESTING (implemented but not validated)

- [ ] Minikube cluster creation with docker driver
- [ ] Helm deployment
- [ ] Basic smoke test

**Time**: 10 minutes

---

## Recommendation

**Option A: Quick validation (40 min)**

Test minikube before starting Phase 2:

1. Test minikube + podman (30 min)
2. Test minikube + docker (10 min)

**Option B: Skip to Phase 2**

Assume minikube works (code looks correct), start EKS immediately.

**Risks**:
- Minikube might have issues we haven't discovered
- Users might try minikube and hit bugs
- Better to validate now than fix later

**Decision needed**: Which option?

---

## Testing Plan for Minikube

If we choose Option A:

### Test 1: Minikube + Podman (30 min)

```bash
# 1. Clean slate
weave stack down
minikube delete

# 2. Init with minikube
weave stack init --template quickstart --runtime minikube

# 3. Edit weave-stack.yaml
# Ensure minikube.driver = "podman" (or auto-detect)

# 4. Deploy
weave stack up --runtime minikube

# 5. Validate
weave stack status
weave stack ingest TestDocs data/
weave cols query TestDocs "search term"

# 6. Clean up
weave stack down
```

**Expected time**: 20-30 min

### Test 2: Minikube + Docker (10 min)

Same as above, but with `minikube.driver = "docker"`

---

## Current Status: Phase 1 Complete ✅

What we know works:

- Kind + Podman: **Fully tested** ✅
- Templates: All 4 templates working ✅
- Commands: init, validate, up, down, status, logs, kubectl, port-forward, ingest ✅
- Integration tests: Passing ✅
- Documentation: Complete ✅

What needs validation:

- Minikube: **Implemented but not tested** ⚠️

---

## Proposed Action Plan

**Tonight** (30-40 min):

1. Test minikube + podman (30 min)
2. Fix any issues found
3. Update test.sh to include minikube tests
4. Commit: "test: validate minikube runtime support"
5. Tag: v0.10.2 (if changes needed)

**Tomorrow**: Start Phase 2 with confidence that local runtimes are solid.

---

## Why This Matters

Phase 2 adds EKS and GKE. If we have bugs in our base runtime abstraction (cluster.go, runtime.go), they'll affect:

1. Kind (working)
2. Minikube (unknown)
3. EKS (Phase 2)
4. GKE (Phase 2)

Better to validate the abstraction now with minikube before adding cloud complexity.

---

## Decision

**Recommendation**: Option A (40 min validation)

**Reason**: Peace of mind, better UX, catches abstraction bugs early.

**If time is tight**: Option B (skip to Phase 2), but add minikube to "known issues" and test later.
