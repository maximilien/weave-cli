# Weave CLI Planning Documents

**Current Status**: v0.11.4 — preparing v0.12.0 (April 2026 launch)
**Last Updated**: March 21, 2026

---

## Current Priorities

### 1. Opik Video/Blog Collaboration (Mar 24–31)

**Checklist**: `docs/blogs/OPIK_VIDEO_BLOG_CHECKLIST.md`

Remaining work:

- [ ] Demo recordings: VDB+RAG (2 databases, multi-modal)
- [ ] Demo recordings: agent orchestration, REPL, custom agents
- [ ] Opik dashboard walkthrough (5+ span traces)
- [ ] Opik evaluation demo (single + benchmarking experiments)
- [ ] Architecture diagram review for video/blog

### 2. v0.12.0 — April 2026 Launch

**Completed:**

- ✅ `weave doctor` — unified diagnostic command
- ✅ Remote storage (S3/MinIO) for backups
- ✅ Document persistence verification
- ✅ 2x faster backups (batch size optimization)

**Remaining:**

- [ ] Fix `TestProviderFactory/CreateOpikProviderWithoutAPIKey`
- [ ] Performance: 500+ docs/sec backup target
- [ ] Any launch blockers identified during testing

### 3. Phase 2 — Cloud Deployments (Post-Launch)

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

**Last Updated**: March 21, 2026
