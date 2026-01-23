# TODOs - Post v0.9.9 Release

**Last Updated**: 2026-01-23 13:15 PST
**Current Status**: v0.9.9 Released - Agent Evaluation System Phase 1 Complete

---

## 🎯 Critical Path (Next 1-2 Weeks)

### 1. Agent Evaluation System - Phase 2 (Advanced Evaluators & Benchmarking) 🚧

**Status**: Ready to start
**Owner**: Maintainer
**Priority**: HIGH

**Phase 2 Goals:**
- Advanced evaluators (context relevance, faithfulness)
- Benchmarking across different agents
- Performance metrics and analysis
- Comparative evaluation reports

**Tasks:**

- [ ] Implement Context Relevance Evaluator
  - Measures how relevant retrieved context is to the query
  - Uses LLM to score each retrieved chunk
  - Detects retrieval failures

- [ ] Implement Faithfulness Evaluator
  - Verifies answer is supported by retrieved context
  - Detects hallucinations more precisely than Phase 1
  - Uses entailment checking

- [ ] Add Benchmarking Command
  - `weave eval benchmark --agents agent1,agent2 --dataset baseline`
  - Compare multiple agents on same dataset
  - Generate comparison report (table, charts)

- [ ] Performance Metrics Collection
  - Response time per test case
  - Token usage tracking
  - Cost estimation
  - Memory usage

- [ ] Results Visualization
  - CLI tables with color coding
  - Export to CSV/JSON for external analysis
  - Generate markdown reports

**Estimated Time:** 4-5 days

---

### 2. Multi-Agent Support (Phase 0 - Planning) 🔮

**Status**: Planning phase
**Owner**: Maintainer
**Priority**: MEDIUM

**Research Questions:**

- What are the primary use cases for multi-agent systems in RAG?
  - Sequential workflows (e.g., query → retrieval → synthesis → citation)
  - Parallel workflows (e.g., query multiple domains simultaneously)
  - Hierarchical workflows (e.g., coordinator → specialist agents)

- How should agents communicate?
  - Direct message passing
  - Shared context/state
  - Event-driven

- What coordination patterns are needed?
  - Sequential chains
  - Parallel fan-out/fan-in
  - Conditional routing
  - Feedback loops

**Deliverables:**

- [ ] Create `docs/planning/MULTI_AGENT_DESIGN.md` with:
  - Use cases and requirements
  - Architecture proposals
  - API design
  - Implementation phases

- [ ] Prototype simple multi-agent workflow
  - Two-agent chain: retrieval agent → synthesis agent
  - Test with baseline dataset
  - Measure vs single-agent performance

**Estimated Time:** 3-4 days for planning + prototype

---

### 3. Test Coverage Improvement 📊

**Status**: Ongoing chore
**Owner**: Maintainer
**Priority**: MEDIUM

**Current Coverage:**
- Unit tests: Good coverage for core packages
- Integration tests: VDB-specific tests complete
- E2E tests: Basic scenarios covered

**Areas to Improve:**

- [ ] Evaluation System Coverage
  - Edge cases (empty datasets, malformed YAML)
  - Error handling (LLM API failures, timeouts)
  - Concurrent evaluation runs
  - Large datasets (100+ test cases)

- [ ] Agent Management Coverage
  - Agent validation edge cases
  - Template rendering errors
  - Concurrent agent operations

- [ ] Cross-VDB Consistency Tests
  - Same operations across all VDBs
  - Verify consistent behavior
  - Document VDB-specific quirks

- [ ] Performance Benchmarks
  - Query latency across VDBs
  - Scaling tests (1K, 10K, 100K docs)
  - Memory profiling

**Deliverables:**
- [ ] Add 20+ new integration tests
- [ ] Create benchmark suite
- [ ] Document coverage goals (target: 80%+)

**Estimated Time:** 2-3 days

---

## 🔨 Work for Next Week (January 27-31)

### 4. Agent Evaluation Phase 2 - Implementation ⏳

**Detailed Breakdown:**

**Day 1-2: Context Relevance Evaluator**
```go
// Pseudo-code structure
type ContextRelevanceEvaluator struct {
    llmClient LLMClient
}

func (e *ContextRelevanceEvaluator) Evaluate(
    query string,
    retrievedChunks []string,
) (score float64, details map[string]interface{}) {
    // Score each chunk for relevance to query
    // Return average score + per-chunk breakdown
}
```

**Day 3: Faithfulness Evaluator**
```go
type FaithfulnessEvaluator struct {
    llmClient LLMClient
}

func (e *FaithfulnessEvaluator) Evaluate(
    answer string,
    context []string,
) (score float64, unsupportedClaims []string) {
    // Extract claims from answer
    // Verify each claim is supported by context
    // Return score + list of unsupported claims
}
```

**Day 4: Benchmarking Command**
```bash
# CLI interface
weave eval benchmark \
  --agents rag-agent,qa-agent,summarize-agent \
  --dataset baseline \
  --output benchmark-results.json

# Output: Comparison table
AGENT         ACCURACY  CITATION  HALLUCINATION  AVG_TIME
rag-agent     0.85      0.90      0.95          150ms
qa-agent      0.88      0.85      0.92          120ms
summarize     0.80      0.70      0.90          200ms
```

**Day 5: Results Visualization**
- CLI table formatting with colors
- Export to CSV/JSON
- Generate markdown report

---

### 5. Multi-Agent Planning & Prototype ⏳

**Day 1-2: Design Document**

Create `docs/planning/MULTI_AGENT_DESIGN.md`:

```markdown
# Multi-Agent System Design

## Use Cases

### 1. Sequential RAG Pipeline
- Agent 1: Query understanding & expansion
- Agent 2: Retrieval from vector DB
- Agent 3: Answer synthesis
- Agent 4: Citation formatting

### 2. Parallel Domain Experts
- Medical agent queries medical docs
- Technical agent queries API docs
- Legal agent queries compliance docs
- Coordinator aggregates results

### 3. Iterative Refinement
- Agent generates initial answer
- Evaluator agent scores quality
- If score < threshold, refine and retry

## Architecture

### Option A: Sequential Chain
```go
type AgentChain struct {
    agents []Agent
}

func (c *AgentChain) Execute(query string) string {
    result := query
    for _, agent := range c.agents {
        result = agent.Process(result)
    }
    return result
}
```

### Option B: Coordinator Pattern
```go
type CoordinatorAgent struct {
    specialists map[string]Agent
}

func (c *CoordinatorAgent) Execute(query string) string {
    // Route to appropriate specialist
    // Or fan out to multiple specialists
    // Aggregate results
}
```

## Implementation Phases

Phase 0: Design (this doc)
Phase 1: Sequential chains
Phase 2: Parallel execution
Phase 3: Advanced coordination
```

**Day 3-4: Prototype**

Implement simple two-agent chain:
```go
// Example: Retrieval Agent → Synthesis Agent
chain := &AgentChain{
    agents: []Agent{
        NewRetrievalAgent("search-docs"),
        NewSynthesisAgent("rag-agent"),
    },
}

result := chain.Execute("What is a vector database?")
```

---

### 6. Test Coverage Improvements ⏳

**Priority Areas:**

**Week 1:**
- [ ] Evaluation system edge cases (15 tests)
- [ ] Agent validation corner cases (10 tests)
- [ ] Cross-VDB consistency (5 tests per VDB × 10 VDBs = 50 tests)

**Week 2:**
- [ ] Performance benchmarks (query latency, scaling)
- [ ] Memory profiling
- [ ] Concurrent operations tests

---

## 📊 Monitoring & Maintenance

### 7. Track GitHub Issues ⏳

**Status**: Ongoing
**Owner**: Maintainer
**Priority**: HIGH

**Actions:**
- [ ] Review open issues weekly
- [ ] Triage new issues within 48 hours
- [ ] Label issues with priorities
- [ ] Close stale issues

### 8. Community Engagement ⏳

**Status**: Backlog
**Owner**: Maintainer
**Priority**: LOW

**Ideas:**
- [ ] Blog post about agent evaluation system
- [ ] Twitter/social media updates
- [ ] Share on r/golang, r/MachineLearning
- [ ] Update project showcases

---

## 🔧 Enhancement Planning (Weeks 2-4)

### 9. Advanced Evaluation Features 🔮

**Ideas for Phase 3:**

- **Custom Evaluators**: Allow users to define their own evaluators
  ```yaml
  # custom-evaluator.yaml
  name: domain-specific-accuracy
  type: llm-as-judge
  prompt: |
    Evaluate the answer for medical accuracy...
  scoring: 0.0-1.0
  ```

- **Human-in-the-Loop**: Capture human feedback on evaluations
  ```bash
  weave eval review run-20260123-150405
  # Opens interactive review UI
  # Allows annotating test cases as correct/incorrect
  ```

- **Regression Testing**: Compare new agent versions vs baseline
  ```bash
  weave eval regression \
    --baseline v1-agent \
    --candidate v2-agent \
    --dataset baseline \
    --threshold 0.05  # Fail if accuracy drops > 5%
  ```

### 10. Multi-Agent Advanced Features 🔮

**Ideas:**

- **Agent Marketplace**: Share agent configs
- **Visual Workflow Editor**: Design multi-agent workflows
- **Conditional Routing**: Route queries based on type
- **Feedback Loops**: Iterative refinement
- **Agent Monitoring**: Track agent performance in production

---

## ✅ Completed Tasks (v0.9.9 - 2026-01-23)

### Agent Evaluation System - Phase 1
- ✅ Dataset management (YAML format)
- ✅ Three evaluators: Accuracy, Citation, Hallucination
- ✅ Results storage (JSON/YAML)
- ✅ CLI commands: datasets, run, results
- ✅ 5 example datasets (baseline, medical-qa, technical-docs, simple-qa, multi-collection)
- ✅ Dataset creation tools (templates, interactive, copy modes)
- ✅ Comprehensive integration tests (11 test cases)

### Code Quality
- ✅ Simplified database flag help output
- ✅ Fixed YAML linting issues
- ✅ Moved debug scripts to tools/dev
- ✅ Removed coverage files from root

---

## 📅 Timeline Summary

**Week of Jan 27** (Week 1):
- Complete Phase 2 evaluation features
- Multi-agent planning document
- 30+ new integration tests

**Week of Feb 3** (Week 2):
- Multi-agent prototype
- Performance benchmarks
- Advanced evaluator prototypes

**Week of Feb 10** (Week 3):
- Polish Phase 2 features
- Begin Phase 3 planning
- Community engagement

**Week of Feb 17** (Week 4):
- Release v0.10.0 with Phase 2 complete
- Start multi-agent Phase 1 implementation

---

## 🎯 Immediate Next Actions

**Today/Tomorrow (Jan 23-24):**

1. ✅ Release v0.9.9 with Phase 1 complete
2. ✅ Update planning documents
3. ⏳ Start Phase 2 design doc

**This Week (Jan 27-31):**

4. Implement Context Relevance Evaluator
5. Implement Faithfulness Evaluator
6. Add benchmarking command
7. Create multi-agent design document

**Next Week (Feb 3-7):**

8. Multi-agent prototype
9. Add 30+ integration tests
10. Performance benchmarks

---

**Questions?** Open a GitHub issue or discussion.
