# Multi-Agent Orchestration: Key Decision Points

> **📚 Related Documents**: [Main Planning Doc](MULTI_AGENT_ORCHESTRATION.md) | [Examples & Use Cases](MULTI_AGENT_EXAMPLES.md) | [Planning Index](README.md)

Quick reference for reviewing the multi-agent orchestration proposal.

---

## 🎯 Quick Summary

**Goal**: Allow multiple agents to work in sequence (e.g., RAG → web search if no docs found)

**Proposed Approach**: Phased implementation
1. **Phase 1**: Simple agent chaining (`--agents rag-agent search-agent`)
2. **Phase 2**: Smart handoff conditions (`--handoff-if no_sources`)
3. **Phase 3**: Declarative configs (`--multiagent rag-with-fallback`)

---

## ❓ Key Questions for Review

### 1. Which UX approach do you prefer?

**Option A: Inline (Simple)**
```bash
weave cols query WeaveDocs "query" --agents rag-agent search-agent
```
- ✅ Easy to use, no new files
- ❌ Limited to sequential execution

**Option B: Config File (Declarative)**
```bash
weave cols query WeaveDocs "query" --multiagent rag-with-fallback
```
- ✅ Complex workflows, reusable
- ❌ More files to manage

**Option C: Both (Hybrid)**
```bash
# Simple: inline
weave cols query WeaveDocs "query" --agents rag-agent search-agent

# Complex: config
weave cols query WeaveDocs "query" --multiagent rag-with-fallback
```
- ✅ Best of both worlds
- ❌ Two patterns to learn

**👉 Recommendation**: Option C (Hybrid) - start with inline, add configs for advanced use cases

---

### 2. What handoff conditions do we need?

**Proposed Built-in Conditions**:
- `no_sources` - No sources found (empty result set)
- `low_scores` - All scores below threshold (e.g., < 0.5)
- `error` - Agent encountered error
- `always` - Always hand off (for pipelines)

**Questions**:
- [ ] Are these conditions sufficient for your use cases?
- [ ] Any other conditions needed? (e.g., `timeout`, `cost_limit`, `source_count_below`)
- [ ] Should handoff logic be in agent prompts or built-in conditions?

**👉 Recommendation**: Start with built-in conditions (more reliable than LLM-based handoff)

---

### 3. How should responses be handled?

**Option A: Last Agent Wins**
```
Agent 1: "No sources found"
Agent 2: "Here are web results"
→ Return only Agent 2 response
```

**Option B: Merge All Responses**
```
Agent 1: 2 sources (score 0.4, 0.3)
Agent 2: 3 sources (score 0.9, 0.8, 0.7)
→ Return merged 5 sources sorted by score
```

**Option C: User Selects**
```
Agent 1: "RAG response"
Agent 2: "Search response"
→ Show both, user picks
```

**Questions**:
- [ ] Default behavior for Phase 1?
- [ ] Should this be configurable per workflow?

**👉 Recommendation**: Start with "last agent wins" (simple), add merging in Phase 2

---

### 4. What observability do users need?

**Option A: Minimal (Silent)**
```
[Final response only, no indication multiple agents ran]
```

**Option B: Summary at End**
```
Response: [answer]

Agents used: rag-agent (no results), search-agent (success)
```

**Option C: Real-time Progress**
```
Running rag-agent...
  No relevant sources found (max score: 0.32)
Running search-agent...
  Found 5 web results
```

**Option D: Verbose Details**
```
[1/3] rag-agent (2.1s)
  - Queried: WeaveDocs
  - Sources: 3
  - Max score: 0.32
  - Decision: Handoff to search-agent (low scores)

[2/3] search-agent (1.8s)
  - Query: "how to configure timeouts"
  - Results: 5 web sources
  - Max score: 0.91
  - Decision: Success, returning response
```

**Questions**:
- [ ] Default verbosity level?
- [ ] Add `--verbose` flag for detailed output?
- [ ] Include timing information?

**👉 Recommendation**: Option B (summary) by default, Option C with `--verbose`

---

### 5. Should we support parallel execution?

**Use Case**: Query 3 RAG agents (different collections) simultaneously, merge results

**Pros**:
- ⚡ Faster (2s instead of 6s for 3 agents)
- 🎯 Better coverage (merge diverse sources)

**Cons**:
- 💰 Higher cost (all agents run, not just until success)
- 🔧 More complex implementation
- 🤔 Harder to debug

**Questions**:
- [ ] Is this needed for MVP?
- [ ] Defer to Phase 3?
- [ ] Specific use cases that need parallel?

**👉 Recommendation**: Defer to Phase 3 (focus on sequential first)

---

### 6. Error handling strategy?

**Scenario**: Agent 2 in chain fails

**Option A: Stop and Return Error**
```
Agent 1: Success
Agent 2: ERROR
→ Return error, show Agent 1 response
```

**Option B: Skip and Continue**
```
Agent 1: Success
Agent 2: ERROR (skip)
Agent 3: Success
→ Return Agent 3 response
```

**Option C: Configurable Per-Agent**
```yaml
agents:
  - name: rag-agent
    on_error: continue  # Skip and try next

  - name: search-agent
    on_error: stop  # Stop chain, return error
```

**Questions**:
- [ ] Default behavior?
- [ ] Should errors be visible to user?
- [ ] Retry logic needed?

**👉 Recommendation**: Option B (skip and continue) with `--strict` flag for stop-on-error

---

## 📋 Proposed Implementation Phases

### Phase 1: Basic Chaining (Week 1)
**UX**:
```bash
weave cols query WeaveDocs "query" --agents rag-agent search-agent
```

**Behavior**:
- Sequential execution (agent1 → agent2 → agent3)
- No automatic handoff (always try all agents)
- Last response wins

**Deliverables**:
- [ ] CLI flag: `--agents`
- [ ] Agent chain executor
- [ ] Basic tests
- [ ] Documentation

**Estimated Effort**: 2-3 days

---

### Phase 2: Smart Handoff (Week 2)
**UX**:
```bash
weave cols query WeaveDocs "query" \
  --agents rag-agent search-agent \
  --handoff-if low_scores \
  --handoff-threshold 0.5
```

**Behavior**:
- Early exit on success
- Built-in handoff conditions
- Progress indicators

**Deliverables**:
- [ ] Handoff condition engine
- [ ] CLI flags: `--handoff-if`, `--handoff-threshold`
- [ ] Progress/summary output
- [ ] Handoff tests

**Estimated Effort**: 3-4 days

---

### Phase 3: Config Files (Weeks 3-4)
**UX**:
```bash
weave cols query WeaveDocs "query" --multiagent rag-with-fallback
```

**Config**: `agents/rag-with-fallback.yaml`
```yaml
name: rag-with-fallback
type: multi-agent

agents:
  - name: rag-agent
    config: rag-agent.yaml
    handoff_conditions:
      - condition: low_scores
        threshold: 0.5
        next_agent: search-agent
      - condition: always
        next_agent: null

  - name: search-agent
    config: search-agent.yaml
```

**Deliverables**:
- [ ] Multi-agent config schema
- [ ] Config loader and validator
- [ ] Complex condition support
- [ ] Example configs

**Estimated Effort**: 5-7 days

---

## 🚀 Alternative Approaches (For Discussion)

### Alternative 1: Prompt-Based Handoff
Agents control handoff via prompts (not recommended):

```yaml
# rag-agent.yaml
system_prompt: |
  If no relevant sources found, respond: HANDOFF:search-agent
```

**Why not recommended**:
- ❌ Brittle (depends on LLM output parsing)
- ❌ Hard to test
- ❌ Non-deterministic

---

### Alternative 2: LLM Router Agent
Let LLM decide which agent to use:

```bash
weave cols query "query" \
  --router smart-router \
  --available-agents rag,search,code
```

**Pros**:
- ✅ Intelligent routing
- ✅ Adapts to query

**Cons**:
- ❌ Extra LLM call (cost + latency)
- ❌ Less predictable
- ❌ Harder to debug

**When to consider**: If query classification is critical (Phase 3+)

---

### Alternative 3: Middleware Pattern
Wrap agents with behavior:

```go
agent := NewRAGAgent(config)
agent = WithFallback(agent, searchAgent)
agent = WithRetry(agent, 3)
agent = WithTimeout(agent, 30*time.Second)
```

**Pros**:
- ✅ Composable
- ✅ Reusable

**Cons**:
- ❌ Code-based (not user-facing)
- ❌ No CLI support

**When to use**: Internal library, not user-facing feature

---

## ✅ Decisions Needed

Please review and provide feedback on:

1. **UX Approach**: Inline, Config, or Both?
2. **Handoff Conditions**: Sufficient? Any additions?
3. **Response Handling**: Last wins, merge, or user selects?
4. **Observability**: How much detail to show users?
5. **Parallel Execution**: MVP or Phase 3?
6. **Error Handling**: Skip or stop on error?
7. **Phase 1 Scope**: Is basic chaining sufficient to start?
8. **Timeline**: Start implementation this week or gather more feedback?

---

## 📚 Full Documentation

- **[Detailed Planning](MULTI_AGENT_ORCHESTRATION.md)**: Complete analysis of approaches, implementation plan, and technical details
- **[Examples & Use Cases](MULTI_AGENT_EXAMPLES.md)**: Real-world examples with expected flows and configurations
- **[This File](DECISION_POINTS.md)**: Quick reference for key decisions
- **[Planning Index](README.md)**: Overview of all planning documents

---

## 💬 Next Steps

1. Review planning docs
2. Gather user feedback (if applicable)
3. Make decisions on key questions
4. Start Phase 1 implementation (if approved)

**Estimated Total Implementation Time**:
- Phase 1: 2-3 days
- Phase 2: 3-4 days
- Phase 3: 5-7 days
- **Total**: 2-3 weeks for full implementation
