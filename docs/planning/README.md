# Planning Documents

**Last Updated**: 2026-02-20
**Current Version**: v0.9.29 (hotfix #42 pending)
**Status**: Awaiting Client0 feedback, prep for next week

---

## Current Status

### Week Feb 17-21 ✅ COMPLETE
- ✅ v0.9.27: Parallel processing (Issue #31) + PDF storage (Issue #33)
- ✅ v0.9.28: All 5 Client0 ingestion improvements shipped
- ✅ Issue #41: Image-page association bug fixed (hotfix)
- ✅ Issue #42: Image metadata fields discovered (hotfix pending Mon)

**Impact**: Client0's 255-line bash wrapper reduced to ~10 lines

### Week Feb 24-28 (Current)
- **Mon AM**: Issue #42 hotfix (CRITICAL - 15 min)
- **Mon PM-Thu**: Issue #38 `weave docs create-batch` implementation
- **Thu-Fri**: Issue #39 `weave docs status` dashboard
- **Target**: v0.9.29 release Friday

See archived: `docs/archive/planning/feb-2026/WEEK_FEB_24-28_PLAN.md`

---

## Active Plans

### Strategic Roadmap Documents

📄 **[V1_0_ROADMAP.md](./V1_0_ROADMAP.md)**
- Path to v1.0 production release
- Feature prioritization
- Timeline and milestones

📄 **[VECTOR_DB_INTEGRATIONS.md](./VECTOR_DB_INTEGRATIONS.md)**
- Status of all 10+ VDB integrations
- Capabilities and limitations
- Integration patterns

---

## Agent & Evaluation System

### Multi-Agent Orchestration
🤖 **Multi-Agent Feature Planning**
- **Status**: Design complete, implementation paused
- **Documents**:
  - **[DECISION_POINTS.md](./DECISION_POINTS.md)** - Quick reference ⭐ Start here
  - **[MULTI_AGENT_ORCHESTRATION.md](./MULTI_AGENT_ORCHESTRATION.md)** - Detailed design
  - **[MULTI_AGENT_EXAMPLES.md](./MULTI_AGENT_EXAMPLES.md)** - Real-world use cases

### Agent Evaluation System
📊 **Agent Evaluation & Quality**
- **Status**: Research phase
- **Documents**:
  - **[AGENT_EVALUATION_SYSTEM.md](./AGENT_EVALUATION_SYSTEM.md)** - System design
  - **[AGENT_EVAL_UPDATES.md](./AGENT_EVAL_UPDATES.md)** - Progress updates
  - **[OPIK_INTEGRATION_STRATEGY.md](./OPIK_INTEGRATION_STRATEGY.md)** - Opik integration
  - **[PLUGGABLE_EVALUATOR_DESIGN.md](./PLUGGABLE_EVALUATOR_DESIGN.md)** - Pluggable architecture
  - **[PLUGGABLE_EVALUATORS_IMPLEMENTATION.md](./PLUGGABLE_EVALUATORS_IMPLEMENTATION.md)**
  - **[OPIK_EVALUATOR_PORT_OPTION_B.md](./OPIK_EVALUATOR_PORT_OPTION_B.md)**
  - **[OPIK_API_INTEGRATION.md](./OPIK_API_INTEGRATION.md)**
  - **[CUSTOM_EVALUATORS_DESIGN.md](./CUSTOM_EVALUATORS_DESIGN.md)**

---

## Feature Planning

### VDB Lifecycle Management
📄 **[VDB_LIFECYCLE_MANAGEMENT.md](./VDB_LIFECYCLE_MANAGEMENT.md)**
- Collection backup/restore
- Migration between VDBs
- Disaster recovery

### Integration Tests
📄 **[INTEGRATION_TEST_PLAN.md](./INTEGRATION_TEST_PLAN.md)**
- Cross-VDB integration testing
- Test coverage improvements

📄 **[NEXT_SESSION_VDB_TESTS.md](./NEXT_SESSION_VDB_TESTS.md)**
- VDB-specific test plans

📄 **[TEST_AUDIT_2026-01-12.md](./TEST_AUDIT_2026-01-12.md)**
- Historical test coverage audit

### Feature Options (Strategic Planning)

📄 **[OPTION_1_NEW_FEATURES.md](./OPTION_1_NEW_FEATURES.md)** - New capabilities
📄 **[OPTION_2_VDB_EXPANSION.md](./OPTION_2_VDB_EXPANSION.md)** - Additional VDBs
📄 **[OPTION_3_PRODUCTION_HARDENING.md](./OPTION_3_PRODUCTION_HARDENING.md)** - Observability
📄 **[OPTION_4_TESTING_QUALITY.md](./OPTION_4_TESTING_QUALITY.md)** - Test coverage
📄 **[OPTION_5_COMMUNITY.md](./OPTION_5_COMMUNITY.md)** - Community building

### Recent Feature Plans (Feb 2026)

📄 **[NEW_FEATURES_2026-02-03.md](./NEW_FEATURES_2026-02-03.md)**
📄 **[TESTING_QUALITY_2026-02-03.md](./TESTING_QUALITY_2026-02-03.md)**
📄 **[PRODUCTION_HARDENING_2026-02-03.md](./PRODUCTION_HARDENING_2026-02-03.md)**
📄 **[CONFIG_FIX_FEATURE.md](./CONFIG_FIX_FEATURE.md)**

---

## Archived Plans

### February 2026 Week Plans
Moved to `docs/archive/planning/feb-2026/`:
- ✅ `WEEK_FEB_17-21_CONSOLIDATED.md` - v0.9.27, v0.9.28 releases
- ✅ `WEEK_FEB_24-28_PLAN.md` - Current week (Issue #38, #39)
- ✅ `CLIENT0_INGESTION_IMPROVEMENTS.md` - Pain point analysis
- ✅ `ISSUE_38_CREATE_BATCH_DESIGN.md` - Batch ingestion design
- ✅ `OSS_FEATURES_CLIENT0.md` - OSS stack features
- ✅ `OSS_EXECUTION_PLAN.md` - OSS implementation plan
- ✅ `BATCH_REEMBEDDING_SPEC.md` - Re-embedding specification
- ✅ `NEXT_STEPS.md` - Old next steps doc

### Historical Planning (Jan 2026 and earlier)
See `docs/archive/planning/`:
- Agent VDB support (v0.9.1)
- RAG agent system (v0.9.0)
- Embedding architecture fixes
- Test coverage plans
- Multi-modal RAG

---

## Session Summaries

📄 **[SESSION_SUMMARY_2026-01-12.md](./SESSION_SUMMARY_2026-01-12.md)**
- Latest session summary (Jan 12)

📄 **[ROADMAP_2026-01-14.md](./ROADMAP_2026-01-14.md)**
- Mid-January roadmap snapshot

📄 **[WEEKEND_AND_NEXT_WEEK.md](./WEEKEND_AND_NEXT_WEEK.md)**
- Weekend planning template

📄 **[THIS_WEEK_PLAN.md](./THIS_WEEK_PLAN.md)**
- Current week template

📄 **[AUDIT_AND_CLEANUP_2026-02-09.md](./AUDIT_AND_CLEANUP_2026-02-09.md)**
- Audit summary (Feb 9)

📄 **[DOC_AUDIT_2026-02-10.md](./DOC_AUDIT_2026-02-10.md)**
- Documentation audit (Feb 10)

---

## Integration Guides

📄 **[integrations/AIRFLOW.md](./integrations/AIRFLOW.md)**
📄 **[integrations/ARGO_WORKFLOWS.md](./integrations/ARGO_WORKFLOWS.md)**
📄 **[integrations/GITHUB_ACTIONS.md](./integrations/GITHUB_ACTIONS.md)**

---

## MCP Integration

📄 **[mcp/MCP_AI_TOOLS.md](./mcp/MCP_AI_TOOLS.md)**
- Model Context Protocol integration
- AI tool support

---

## Planning Process

### Document Lifecycle
```
Planning → Active → Completed → Archived
   ↓         ↓          ↓           ↓
OPTION_*.md  WEEK_*.md  ✅ Tag    docs/archive/
```

### Archive Structure
```
docs/archive/planning/
├── feb-2026/          # February 2026 week plans
├── README.md          # Archive index
└── [older plans]      # Pre-Jan 2026 plans
```

---

## Path to v1.0.0

**Current**: v0.9.29 (in progress)
**Next**: v0.9.30, v0.9.31, ...
**Target**: v1.0.0 (production ready)

**v1.0.0 Criteria**:
- ✅ 10+ VDBs fully supported
- ✅ Batch ingestion with checkpointing
- ✅ Image ingestion + external storage
- ⏳ Test coverage >70%
- ⏳ Production observability
- ⏳ Comprehensive documentation

---

**Maintainer**: @maximilien
**Repository**: https://github.com/maximilien/weave-cli
**Issues**: https://github.com/maximilien/weave-cli/issues
