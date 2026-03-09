# Planning Archive

This directory contains historical planning documents from the Weave CLI development process.

## Archive Structure

```
docs/archive/planning/
├── 2025-nov/              # November 2025 (v0.7.0-v0.7.1)
├── 2025-dec/              # December 2025 (v0.7.2)
├── 2026-jan/              # January 2026 (v0.9.0-v0.9.1)
├── 2026-feb/              # February 2026 (v0.9.27-v0.9.29)
├── feb-2026/              # February 2026 feature specs
├── features/              # Feature design specifications
├── strategic/             # Strategic planning & roadmaps
├── phase-1/               # Phase 1 development plans
├── weave-stack/           # Weave Stack specific plans
└── README.md              # This file
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

## 2025 Planning Documents (Nov-Dec)

### November 2025 (`2025-nov/`)
- Qdrant integration completion (v0.7.0)
- Neo4j integration (v0.7.1)
- Initial VDB expansion work
- TODOs for Pinecone, Redis, Neo4j

### December 2025 (`2025-dec/`)
- Chroma integration (v0.7.2)
- Work plans for Chroma and Neo4j
- Next steps planning

---

## 2026 Planning Documents

### January 2026 (`2026-jan/`)
- RAG agent system (v0.9.0-v0.9.1)
- Multi-VDB agent support
- Multi-modal RAG planning
- Test coverage focus
- Embedding architecture fixes

### February 2026 (`2026-feb/` and `feb-2026/`)
- Client0 ingestion improvements (v0.9.27-v0.9.29)
- Batch processing features
- External storage implementation
- Agent evaluation system
- Week plans for Feb 17-28

### Feature Specifications (`features/`)
- Agent evaluation system
- Pluggable evaluators
- Multi-agent orchestration
- External storage
- VDB lifecycle management
- Weave Stack PM2 dashboard
- Opik API integration

### Strategic Planning (`strategic/`)
- V1.0 roadmap
- Vector DB integrations matrix
- 5 strategic options analysis
- Testing & quality plans
- Production hardening
- Community building

---

## Current Planning

For active planning, see:
- **docs/PLAN.md** - Sprint planning & roadmap
- **docs/archive/mar-2026/** - March 2026 active work
- **CHANGELOG.md** - Release tracking

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

**Last Updated**: 2026-03-09
**Current Version**: v0.11.2
**Archive Reorganized**: March 9, 2026 - Consolidated 74 planning docs into organized structure
