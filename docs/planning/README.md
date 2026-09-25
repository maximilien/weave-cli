# Weave CLI Planning Documents

**Current status**: 66.08% coverage; adapter floor complete; v0.15.0 released

**Last updated**: September 25, 2026

---

## Current Priorities

### 1. Continue toward 80% coverage

- The campaign raised statement coverage from the v0.14.0 baseline of 30.18%
  to 40.06% for v0.15.0 and 66.08% on Day 23.
- Preserve the 66% project ratchet and 80% patch target.
- Advance one percentage-point checkpoint per day; 67% is next on Day 24,
  and overshoot does not skip the following day's rung.
- Continue toward 80%, prioritizing command and core-package gaps.
- See [`../tests/COVERAGE_PLAN.md`](../tests/COVERAGE_PLAN.md).

### 2. Adapter consistency

- Preserve the completed 50% floor across every supported vector database
  adapter while deepening error and pagination coverage.

### 3. Phase 2 — Cloud deployments

Deferred to after April launch. Plans in this directory:

- [WEAVE_STACK_PHASE_2_PLAN.md](WEAVE_STACK_PHASE_2_PLAN.md)
- [PHASE_2_ITEMS.md](PHASE_2_ITEMS.md)

Scope: EKS/GKE support, TLS/SSL, secrets management,
monitoring/observability.

---

## Planning Documents

### Active

| Document | Purpose |
| --- | --- |
| [PHASE_2_ITEMS.md](PHASE_2_ITEMS.md) | Phase 2 task checklist |
| [WEAVE_STACK_PHASE_2_PLAN.md](WEAVE_STACK_PHASE_2_PLAN.md) | Phase 2 master plan |
| [MULTI_VDB_SUPPORT_PLAN.md](MULTI_VDB_SUPPORT_PLAN.md) | Multi-VDB stack support |
| [BACKUP_RESTORE_DESIGN.md](BACKUP_RESTORE_DESIGN.md) | Backup design (shipped) |

### Validation Results (Reference)

| Document | Purpose |
| --- | --- |
| [PRODUCTION_READINESS_AUDIT_RESULTS.md](PRODUCTION_READINESS_AUDIT_RESULTS.md) | v0.10.2 audit |
| [MINIKUBE_VALIDATION_RESULTS.md](MINIKUBE_VALIDATION_RESULTS.md) | Minikube constraints |
| [PRE_PHASE_2_VALIDATION.md](PRE_PHASE_2_VALIDATION.md) | Pre-Phase 2 checklist |

### Archive

Phase 1 plans and historical docs: `../archive/planning/`

---

## Quick Links

- **Roadmap**: [../ROADMAP.md](../ROADMAP.md)
- **Opik checklist**: `../blogs/OPIK_VIDEO_BLOG_CHECKLIST.md`
- **Getting started**: [../CLIENT0_GETTING_STARTED.md](../CLIENT0_GETTING_STARTED.md)
- **Stack guide**: [../guides/WEAVE_STACK_QUICKSTART.md](../guides/WEAVE_STACK_QUICKSTART.md)

---

**Last updated**: September 25, 2026
