# Weave CLI Planning Documents

**Current status**: preparing v0.13.0; next milestone is v0.14.0 at 30% coverage

**Last updated**: September 8, 2026

---

## Current Priorities

### 1. v0.13.0 release

- Publish the security, dependency, and 25% coverage milestone.
- Verify cross-platform binaries and checksums from the release workflow.

### 2. v0.14.0 coverage campaign

- Raise statement coverage from 25.28% to 30%.
- Focus on command boundaries, Weaviate protocols, and low-coverage adapters.
- See [`../tests/COVERAGE_PLAN.md`](../tests/COVERAGE_PLAN.md).

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

**Last updated**: September 8, 2026
