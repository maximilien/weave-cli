# This Week Plan - Feb 10-14, 2026

## 🎉 What We Accomplished Today (Feb 9)

### Morning: v0.9.17 Shipped
- ✅ Batch re-embedding feature complete
- ✅ CLI registration bug fixed (issue #28)
- ✅ 28 tests passing
- ✅ CHANGELOG updated
- ✅ v0.9.17 tagged and pushed

### Afternoon: v0.9.18 Shipped
- ✅ Issue #7: Ollama auto-discovery in agent wizard
- ✅ Issue #9: Embedding model comparison report generator
- ✅ Both features tested and working
- ✅ Client0 issues closed on GitHub
- ✅ v0.9.18 tagged and pushed

**Total**: 2 releases shipped in one day! 🚀

---

## 📋 This Week Priorities (Feb 10-14)

### Monday (Feb 10) - Client0 Support & Feedback

**Morning (2-3 hours):**
1. **Monitor Client0 Feedback**
   - Check GitHub issues for bug reports
   - Answer questions about v0.9.17/v0.9.18
   - Help with first re-embedding tests

2. **Documentation Updates**
   - Add comparison report examples to README
   - Update USER_GUIDE with new commands
   - Create Ollama setup guide

**Afternoon (2-3 hours):**
3. **CHANGELOG Update for v0.9.18**
   - Document Issue #7 and #9 features
   - Add usage examples
   - Update version history

4. **Quick Wins**
   - Fix any critical bugs reported by Client0
   - Performance profiling if needed

---

### Tuesday (Feb 11) - OSS Embedding Provider (High Priority)

**Goal**: Add sentence-transformers support for Client0's Week 2 testing

**Morning (2-3 hours):**
1. **sentence-transformers Provider Implementation**
   ```
   File: src/pkg/reembedding/providers/sentence_transformers.go
   ```

   **Approach Options:**
   - **Option A**: Python subprocess (simplest, but requires Python)
   - **Option B**: HTTP service (requires separate server)
   - **Option C**: gRPC bridge (complex, but cleanest)

   **Recommendation**: Start with Option A (subprocess) for speed

2. **Python Bridge**
   - Embed Python script in Go binary
   - Call sentence-transformers via subprocess
   - Handle batch embedding generation
   - Error handling for missing dependencies

**Afternoon (2-3 hours):**
3. **Provider Integration**
   - Update `EmbeddingPipeline.ProcessBatch()`
   - Detect sentence-transformers provider
   - Add provider tests
   - Integration test with re-embed workflow

4. **Models to Support**
   - `sentence-transformers/all-mpnet-base-v2` (768d)
   - `sentence-transformers/all-minilm-l6-v2` (384d)
   - `sentence-transformers/all-minilm-l12-v2` (384d)

**Deliverable**: Commit "feat: add sentence-transformers provider for re-embedding"

---

### Wednesday (Feb 12) - Ollama Provider

**Goal**: Add Ollama embedding support for Client0's Week 3 testing

**Morning (2-3 hours):**
1. **Ollama Provider Implementation**
   ```
   File: src/pkg/reembedding/providers/ollama.go
   ```

   - HTTP client for Ollama API (`POST /api/embeddings`)
   - Batch embedding support
   - Connection validation
   - Graceful error messages

2. **Models to Support**
   - `nomic-embed-text` (768d)
   - `mxbai-embed-large` (1024d)
   - `snowflake-arctic-embed` (1024d)

**Afternoon (2-3 hours):**
3. **Provider Tests**
   - Unit tests with mock HTTP server
   - Integration tests with re-embed workflow
   - Error handling (Ollama not running)

4. **Documentation**
   - Update `docs/BATCH_REEMBEDDING_SPEC.md`
   - Add Ollama setup instructions
   - Update CHANGELOG

**Deliverable**: Commit "feat: add Ollama provider for re-embedding"

---

### Thursday (Feb 13) - Testing & Polish

**Morning (2-3 hours):**
1. **End-to-End Testing**
   - Test sentence-transformers with mock data
   - Test Ollama with local instance
   - Test comparison reports with both providers
   - Performance benchmarks

2. **Bug Fixes**
   - Address any issues found in testing
   - Improve error messages
   - Add progress indicators

**Afternoon (2-3 hours):**
3. **Documentation Polish**
   - Create `docs/REEMBEDDING_GUIDE.md`
   - Add troubleshooting section
   - Update README with examples
   - Video walkthrough (optional)

4. **Client0 Validation Support**
   - Test with Client0's actual dataset (if available)
   - Generate example comparison reports
   - Document best practices

**Deliverable**: Commit "docs: comprehensive re-embedding guide with troubleshooting"

---

### Friday (Feb 14) - v0.9.19 Release

**Morning (2-3 hours):**
1. **Pre-Release Checklist**
   - All tests passing
   - Documentation complete
   - CHANGELOG updated
   - No known critical bugs

2. **Release Preparation**
   - Update CHANGELOG for v0.9.19
   - Tag release: `git tag -a v0.9.19`
   - Push to GitHub
   - Rebuild binary

**Afternoon (1-2 hours):**
3. **Client0 Handoff**
   - Update GitHub issues
   - Notify Client0 of v0.9.19 availability
   - Provide migration guide
   - Schedule follow-up for next week

4. **Week Review**
   - Document accomplishments
   - Plan next week priorities
   - Update roadmap

**Deliverable**: v0.9.19 shipped with full OSS support

---

## 🎯 Success Criteria

### Must Have (This Week)
- ✅ v0.9.18 shipped (DONE - Feb 9)
- ⏳ sentence-transformers provider working
- ⏳ Ollama provider working
- ⏳ Client0 can test OSS models
- ⏳ Comparison reports include all providers

### Nice to Have
- Performance optimization (if time permits)
- Checkpoint/resume for re-embedding
- Cost estimator in comparison reports
- Batch size auto-tuning

### Stretch Goals
- Cross-VDB re-embedding
- Multi-query aggregation in reports
- Automatic model recommendation

---

## 📊 Metrics to Track

### Code Quality
- Unit test coverage: >80% for new code
- All integration tests passing
- No linting errors
- Documentation complete

### Performance
- sentence-transformers: 150+ docs/min target
- Ollama: 200+ docs/min target
- Memory usage: <2GB peak
- API errors: <1% failure rate

### Client0 Success
- Can re-embed with OSS models
- Comparison reports generated successfully
- No blocking bugs
- Positive feedback on GitHub

---

## 🚧 Known Blockers & Mitigation

### Blocker 1: Python Dependencies
**Risk**: sentence-transformers requires Python + pip packages

**Mitigation**:
- Provide clear setup instructions
- Check for Python in provider initialization
- Graceful error messages
- Docker alternative (if needed)

### Blocker 2: Ollama Installation
**Risk**: Users need Ollama installed and running

**Mitigation**:
- Already have auto-discovery (Issue #7) ✅
- Setup instructions in error messages
- Test with mock Ollama server
- Document Docker deployment

### Blocker 3: Time Constraints
**Risk**: 5 days to ship 2 providers + testing

**Mitigation**:
- Prioritize sentence-transformers (most requested)
- Ollama can slip to next week if needed
- MVP approach: basic functionality first
- Polish in v0.9.20 if needed

---

## 📝 Daily Standups

### Monday Morning
- What shipped: v0.9.17, v0.9.18
- Today's goal: Support Client0, update docs
- Blockers: None

### Tuesday Morning
- Yesterday: Docs updated, Client0 supported
- Today's goal: sentence-transformers provider
- Blockers: TBD (Python setup complexity)

### Wednesday Morning
- Yesterday: sentence-transformers working
- Today's goal: Ollama provider
- Blockers: TBD

### Thursday Morning
- Yesterday: Ollama provider working
- Today's goal: Testing and polish
- Blockers: None expected

### Friday Morning
- Yesterday: Tests passing, docs complete
- Today's goal: Ship v0.9.19
- Blockers: None expected

---

## 🔄 Iteration Plan

### If Ahead of Schedule
1. Add cost estimator to comparison reports
2. Implement checkpoint/resume
3. Performance optimization (concurrency)
4. Cross-VDB re-embedding

### If Behind Schedule
**Cut Scope (Priority Order):**
1. Ollama provider → v0.9.20
2. Comparison report enhancements → v0.9.20
3. Documentation polish → incremental updates
4. Performance optimization → v0.9.20

**Keep in Scope:**
- sentence-transformers provider (highest priority)
- Basic testing
- Critical bug fixes
- Minimal documentation

---

## 📞 Communication Plan

### Client0
- **Monday**: Check for feedback on v0.9.18
- **Wednesday**: Notify when sentence-transformers ready
- **Friday**: Announce v0.9.19 release

### GitHub
- Close issues as features ship
- Update issue comments with progress
- Create new issues for bugs found

### Documentation
- Update CHANGELOG daily
- Keep README current
- Commit docs with code

---

## 🎯 Week Goals Summary

**Primary Goal**: Enable Client0 to test OSS embedding models

**Secondary Goals**:
1. sentence-transformers provider working
2. Ollama provider working
3. Comparison reports support all providers
4. Documentation complete

**Success Looks Like**:
- Client0 can test 4+ embedding models
- Comparison reports show all models
- No blocking bugs
- v0.9.19 shipped by Friday

**Time Budget**:
- Total: ~12-15 hours
- Tuesday-Wednesday: 8-10 hours (providers)
- Thursday: 3-4 hours (testing/polish)
- Friday: 2-3 hours (release)

---

## 📚 Reference

### Key Files
- `src/pkg/reembedding/pipeline.go` - Pipeline to enhance
- `src/pkg/ollama/client.go` - Ollama discovery (done)
- `src/cmd/collection/compare.go` - Comparison reports (done)
- `docs/planning/BATCH_REEMBEDDING_SPEC.md` - Technical spec

### Key Commands
```bash
# Test sentence-transformers
weave collection reembed Source \
  --new-embedding sentence-transformers/all-mpnet-base-v2 \
  --output Target

# Test Ollama
weave collection reembed Source \
  --new-embedding nomic-embed-text \
  --output Target

# Compare all models
weave collection compare \
  Original OpenAI OSS Ollama \
  --query "test" \
  --report comparison.md
```

### Related Issues
- Issue #7: Ollama auto-discovery (CLOSED)
- Issue #8: Batch re-embedding (CLOSED)
- Issue #9: Comparison reports (CLOSED)

---

**Week Start**: Monday, Feb 10, 2026
**Week End**: Friday, Feb 14, 2026
**Goal**: Ship v0.9.19 with OSS embedding support 🚀
