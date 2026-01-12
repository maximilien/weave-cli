# Next Session: VDB Test Coverage Sprint

**Created:** 2026-01-12
**Status:** Ready to execute
**Duration:** 2-3 hours
**Goal:** Establish 25% minimum coverage baseline across 5+ VDBs

## Current State

### ✅ Completed (This Session)

| VDB | Coverage | Tests | Status |
|-----|----------|-------|--------|
| **Qdrant** | 21.9% | 33 tests (4 files) | ✅ DONE |
| **Weaviate** | 5.8% | 18 tests (3 files) | ✅ CLEAN |
| **MongoDB** | ~20% | 3 test files | ⚠️ Timeouts |

**Total:** 3 VDBs with test coverage established

### ❌ Pending

- Milvus: 0%
- Chroma: 0%
- Neo4j: 0%
- Supabase: 0%
- Pinecone: 0%
- Elasticsearch: 0%
- OpenSearch: 0%

**Total:** 7 VDBs with no coverage

## Next Session Goals

### Primary Target: 5 VDBs at 25%+ coverage

1. **Fix MongoDB** (5-10 mins)
   - Replace connection-based tests with pure unit tests
   - Target: 20% → 25%

2. **Milvus** (20-30 mins)
   - Apply Qdrant pattern
   - Target: 0% → 25%

3. **Chroma** (20-30 mins)
   - Apply Qdrant pattern
   - Target: 0% → 25%

4. **Neo4j** (20-30 mins)
   - Apply Qdrant pattern
   - Target: 0% → 25%

5. **Stretch: Supabase** (if time permits)
   - Apply Qdrant pattern
   - Target: 0% → 25%

## Testing Pattern (Proven)

### Files to Create (per VDB)

1. **`factory_test.go`** (~100 LOC)
   - NewFactory
   - GetSupportedTypes
   - ValidateConfig (URL, API keys, type validation)
   - CreateClient

2. **`helpers_test.go`** (~200 LOC)
   - Type mapping functions
   - Helper utilities (timeout, parsing, conversion)
   - Error wrapping
   - Configuration defaults

3. **`adapter_test.go`** (~100 LOC)
   - NewAdapter (success cases)
   - NewAdapter (error cases)
   - Document conversion functions
   - Schema conversion functions

### What Tests Cover (25% baseline)

✅ Factory and validation (no server required)
✅ Helper functions (pure logic)
✅ Adapter creation and config
✅ Type conversions
✅ Error handling
✅ Default values

❌ NOT covered (requires integration tests):
- Actual CRUD operations
- Live server connections
- Network I/O

## Execution Checklist

### For Each VDB:

- [ ] Check existing test files
- [ ] Identify helper functions to test
- [ ] Create factory_test.go
- [ ] Create helpers_test.go
- [ ] Create adapter_test.go
- [ ] Run tests: `go test ./src/pkg/vectordb/{vdb}`
- [ ] Check coverage: `go test -coverprofile=coverage.out ./src/pkg/vectordb/{vdb}`
- [ ] Verify ≥25% coverage
- [ ] Commit: `test: add {vdb} unit tests for 25% coverage baseline`

### Quick Commands

```bash
# Test specific VDB
go test -v ./src/pkg/vectordb/milvus

# Get coverage
go test -coverprofile=cov.out ./src/pkg/vectordb/milvus
go tool cover -func=cov.out | grep "^total:"

# View coverage by function
go tool cover -func=cov.out | head -30
```

## Success Criteria

By end of next session:

- ✅ **5+ VDBs** at 25% coverage minimum
- ✅ All tests passing
- ✅ Linting clean
- ✅ Pattern documented
- ✅ Commits for each VDB

## Post-Sprint Actions

1. **Run full coverage report**
   ```bash
   ./tools/test-coverage.sh
   ```

2. **Update TEST_AUDIT.md** with new coverage numbers

3. **Tag progress commit**
   - Consider v0.9.2-alpha with improved test coverage

4. **Document learnings**
   - Update TESTING.md with patterns
   - Note any VDB-specific quirks

## Notes for Next Developer

### Quick Start

```bash
# Start with Milvus (likely most similar to Qdrant)
cd src/pkg/vectordb/milvus

# Copy test structure from Qdrant
cp ../qdrant/*_test.go .

# Modify for Milvus specifics
# - Function signatures
# - Config fields
# - Type mappings

# Test and iterate
go test -v .
```

### Time Estimates

- MongoDB fix: 5-10 mins
- Milvus: 25 mins
- Chroma: 25 mins
- Neo4j: 30 mins (may have different patterns)
- Supabase: 25 mins

**Total:** ~2 hours for 5 VDBs

### Potential Blockers

1. **VDB-specific interfaces** - Some VDBs may have unique methods
2. **Config structures** - Different config fields need different tests
3. **Type systems** - Distance metrics, data types vary by VDB

### Mitigation

- Reference existing VDB implementation code
- Keep tests simple (25% baseline, not exhaustive)
- Skip complex integration scenarios
- Document gaps for future integration tests

## Long-Term Vision

This sprint sets foundation for:

- **v0.9.2:** 50%+ overall VDB coverage (5-6 VDBs at 25%+)
- **v0.9.3:** 60%+ overall VDB coverage (8+ VDBs at 25%+)
- **v0.9.4:** All 10 VDBs at 25%+ baseline
- **v1.0.0:** 70%+ overall coverage with integration tests

## References

- Established patterns: `src/pkg/vectordb/qdrant/*_test.go`
- Coverage script: `tools/test-coverage.sh`
- Mock infrastructure: `tests/mocks/`
- Current audit: `docs/planning/TEST_AUDIT_2026-01-12.md`
