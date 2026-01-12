# Planning Documents

**Last Updated**: 2026-01-12
**Current Version**: v0.9.1
**Next Release**: v0.9.2 (Test Coverage Focus)

---

## Active Plans

### This Week (Jan 13-17, 2026)

📋 **[WEEK_OF_2026-01-13.md](./WEEK_OF_2026-01-13.md)**
- **Focus**: Test Coverage Improvement
- **Goal**: v0.9.2 release with 60%+ unit test coverage
- **Key Tasks**:
  - Set up mock framework
  - Add unit tests for Qdrant, Weaviate, MongoDB
  - Establish CI coverage tracking
  - Create testing documentation

📊 **[TEST_AUDIT_2026-01-12.md](./TEST_AUDIT_2026-01-12.md)**
- **Focus**: Comprehensive test coverage analysis
- **Current State**: 50% (Qdrant only), minimal unit tests
- **Target**: 70%+ coverage across all VDBs
- **Priority**: #1 blocker for v1.0.0

---

## Strategic Options (Future Planning)

These documents outline potential focus areas for upcoming releases. Not all will be implemented - they represent options for discussion and prioritization.

### Option 1: New Features
📄 **[OPTION_1_NEW_FEATURES.md](./OPTION_1_NEW_FEATURES.md)**
- **Status**: Partially Complete (4/7 features)
- **Priority**: ⭐⭐⭐ High
- **Completed**:
  - ✅ Pipeline commands (batch ingestion)
  - ✅ Interactive REPL mode
  - ✅ MCP client integration
  - ✅ AI schema suggestion
- **Remaining**:
  - Progress bars
  - JSON/YAML output enhancements
  - Collection statistics

### Option 2: VDB Expansion
📄 **[OPTION_2_VDB_EXPANSION.md](./OPTION_2_VDB_EXPANSION.md)**
- **Status**: Planning
- **Priority**: ⭐ Medium
- **Target VDBs**:
  - LanceDB
  - Redis Stack
  - Vespa
  - Typesense
  - Marqo

### Option 3: Production Hardening
📄 **[OPTION_3_PRODUCTION_HARDENING.md](./OPTION_3_PRODUCTION_HARDENING.md)**
- **Status**: Planning
- **Priority**: ⭐⭐ High
- **Focus Areas**:
  - Structured logging (zerolog)
  - Prometheus metrics
  - OpenTelemetry tracing
  - Circuit breakers
  - Connection pooling
  - Security hardening

### Option 4: Testing & Quality
📄 **[OPTION_4_TESTING_QUALITY.md](./OPTION_4_TESTING_QUALITY.md)**
- **Status**: ⚡ **ACTIVE** (This Week's Focus)
- **Priority**: ⭐⭐⭐ Highest
- **See**: [TEST_AUDIT_2026-01-12.md](./TEST_AUDIT_2026-01-12.md)
- **Note**: This is our current top priority

### Option 5: Community
📄 **[OPTION_5_COMMUNITY.md](./OPTION_5_COMMUNITY.md)**
- **Status**: Planning
- **Priority**: ⭐ Medium
- **Focus Areas**:
  - Video tutorials
  - Blog posts
  - Conference talks
  - Community forum
  - Plugin system

---

## Reference Documents

### Vector DB Integrations
📄 **[VECTOR_DB_INTEGRATIONS.md](./VECTOR_DB_INTEGRATIONS.md)**
- Comprehensive overview of all VDB integrations
- Status, capabilities, limitations
- Integration patterns

---

## Completed Plans (Archived)

Completed planning documents have been moved to `docs/archive/planning/`:

- ✅ **AGENT_VDB_SUPPORT_AND_PROGRESS.md** (v0.9.1)
  - Multi-VDB agent support
  - Query progress indicators

- ✅ **RAG_AGENT_FEATURE.md** (v0.9.0)
  - RAG agent system implementation
  - Agent YAML configuration
  - Citation formats

- ✅ **VDB_EMBEDDING_ARCHITECTURE_FIX.md** (v0.9.1)
  - Embedding dimension verification
  - Dimension mismatch error handling
  - All 10 VDBs updated

- ✅ **VDB_AGENT_TESTING_PLAN.md** (v0.9.1)
  - Agent testing across all VDBs
  - Verification completed

See `docs/archive/planning/` for historical context.

---

## Planning Process

### How to Use These Documents

1. **Review Active Plans** - Check WEEK_OF_*.md for current week's priorities
2. **Reference Strategic Options** - OPTION_*.md provide context for future decisions
3. **Check Archived Plans** - See docs/archive/planning/ for completed work

### Creating New Plans

When creating new planning documents:

1. **Start with clear status** - Planning, Active, Completed
2. **Define success criteria** - What does "done" look like?
3. **Estimate effort** - Hours/days required
4. **Set priority** - ⭐ Low, ⭐⭐ Medium, ⭐⭐⭐ High
5. **Archive when complete** - Move to docs/archive/planning/

### Document Lifecycle

```
Planning → Active → Completed → Archived
   ↓         ↓          ↓           ↓
OPTION_*.md  WEEK_*.md  ✅ Tag    docs/archive/
```

---

## Path to v1.0.0

**Current**: v0.9.1
**Next**: v0.9.2 (Test Coverage)
**Then**: v0.9.3, v0.9.4, v0.9.5
**Release Candidate**: v1.0.0-rc1
**MAJOR RELEASE**: v1.0.0 🎉

**Target Date**: Mid-February 2026

**v1.0.0 Criteria**:
- ✅ 10+ VDBs fully supported
- ✅ Test coverage >70%
- ✅ Production hardening complete
- ✅ Comprehensive documentation
- ✅ Strong community adoption

See [WEEK_OF_2026-01-13.md](./WEEK_OF_2026-01-13.md) for detailed roadmap.

---

## Contact

**Maintainer**: @maximilien
**Repository**: https://github.com/maximilien/weave-cli
**Issues**: https://github.com/maximilien/weave-cli/issues
