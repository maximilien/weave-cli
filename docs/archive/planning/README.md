# Planning Archive

This directory contains historical planning documents from the Weave CLI development process.

## Archive Structure

```
docs/archive/planning/
├── feb-2026/              # February 2026 week plans
├── [legacy plans]         # Pre-2026 plans (listed below)
└── README.md             # This file
```

---

## February 2026 Week Plans

**Location**: `docs/archive/planning/feb-2026/`

### Week of Feb 17-21, 2026
- **WEEK_FEB_17-21_CONSOLIDATED.md** - v0.9.27, v0.9.28 releases
  - Issue #31: Parallel processing
  - Issue #33: PDF storage in MinIO
  - Issues #34-40: Client0 ingestion improvements

### Week of Feb 24-28, 2026
- **WEEK_FEB_24-28_PLAN.md** - v0.9.29 planning
  - Issue #38: `weave docs create-batch`
  - Issue #39: `weave docs status` dashboard
  - Issue #42: Image metadata fields hotfix

### Client0 Feature Planning
- **CLIENT0_INGESTION_IMPROVEMENTS.md** - Pain point analysis and feature requests
- **ISSUE_38_CREATE_BATCH_DESIGN.md** - Detailed batch ingestion design spec
- **OSS_FEATURES_CLIENT0.md** - OSS stack features (embeddings, re-embedding)
- **OSS_EXECUTION_PLAN.md** - OSS feature implementation plan
- **BATCH_REEMBEDDING_SPEC.md** - Batch re-embedding specification
- **NEXT_STEPS.md** - Re-embedding implementation plan

**Status**: All features shipped in v0.9.27-v0.9.29

---

## Legacy Planning Documents (2025)

### January 2026 Plans
See `docs/archive/planning/` (not in feb-2026/ subdirectory):
- **WEEK_OF_2026-01-13.md** - Test coverage focus
- **NEXT_STEPS_2026-01-20.md** - Multi-modal RAG planning
- **AGENT_VDB_SUPPORT_AND_PROGRESS.md** - Multi-VDB agent support (v0.9.1)
- **RAG_AGENT_FEATURE.md** - RAG agent system (v0.9.0)
- **VDB_AGENT_TESTING_PLAN.md** - Agent testing across VDBs
- **VDB_EMBEDDING_ARCHITECTURE_FIX.md** - Embedding dimension fixes

### December 2025 Plans
- **WORK_PLAN-2025-12-03.md** (v0.7.2) - Chroma integration
- **WORK_PLAN-chroma.md** - Chroma implementation details

### November 2025 Plans
- **WORK_PLAN-2025-11-29.md** (v0.7.0) - Qdrant integration completion
- **WORK_PLAN-neo4j.md** (v0.7.1) - Neo4j integration
- **SESSION_SUMMARY_2025-11-28.md** - Qdrant completion session

### Next Steps & TODOs
- **NEXT_STEPS.md** (v0.7.0) - Post-Qdrant planning
- **NEXT_STEPS_neo4j.md** - Neo4j planning
- **NEXT_STEPS_tomorrow.md** - Daily Neo4j plan
- **TODOs.md** - General project todos (superseded)
- **TODOs-neo4j.md** - Neo4j checklist (completed)
- **TODOs-pinecone.md** - Pinecone planning (future)
- **TODOs-redis.md** - Redis planning (future)

### Progress Tracking
- **REPL_PROGRESS.md** - REPL mode implementation progress

---

## Current Planning Documents

For active planning, see:
- **docs/planning/README.md** - Active planning index
- **docs/planning/V1_0_ROADMAP.md** - Path to v1.0
- **docs/planning/VECTOR_DB_INTEGRATIONS.md** - VDB status

---

## Archive Policy

Documents are moved here when:
1. ✅ All tasks are completed
2. ✅ The work has been released
3. ✅ The document is superseded by newer planning
4. ✅ Historical reference value only

**Active vs Archived**:
- Active: Current week plans, strategic roadmaps, ongoing feature designs
- Archived: Completed week plans, shipped feature specs, historical context

---

**Last Updated**: 2026-02-20
**Current Version**: v0.9.29 (in progress)
