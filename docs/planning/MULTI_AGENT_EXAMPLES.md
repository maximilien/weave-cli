# Multi-Agent Examples

> **📚 Related Documents**: [Main Planning Doc](MULTI_AGENT_ORCHESTRATION.md) | [Decision Points](DECISION_POINTS.md) | [Planning Index](README.md)

Real-world examples of multi-agent workflows and their expected behavior.

---

## Example 1: RAG with Web Search Fallback

### Scenario
User asks about a topic not in the vector database. RAG agent should gracefully hand off to web search.

### Command (Phase 1 - Simple)
```bash
weave cols query WeaveDocs "what is the latest Anthropic Claude release?" \
  --agents rag-agent search-agent \
  --top_k 5
```

### Expected Flow

**Step 1: RAG Agent Executes**
```
Running rag-agent...
  Querying collection: WeaveDocs
  Found 2 sources, max score: 0.32
  All sources below relevance threshold (0.5)

  Response: "I don't have recent information about Anthropic Claude releases in my knowledge base."
```

**Step 2: Search Agent Executes**
```
Running search-agent...
  Previous agent found no relevant sources
  Performing web search: "latest Anthropic Claude release"
  Found 5 web results

  Response: "The latest Anthropic Claude release is Claude 3.5 Sonnet, announced on June 20, 2024.
  It features improved coding abilities, vision capabilities, and a 200K context window. [1][2][3]"
```

**Final Output**
```
Sources:
[1] Search result: anthropic.com - Score: 95%
[2] Search result: techcrunch.com - Score: 88%
[3] Search result: theverge.com - Score: 82%

Response: The latest Anthropic Claude release is Claude 3.5 Sonnet...

Execution Summary:
  Agents used: rag-agent (no results), search-agent (success)
  Total time: 3.2s
```

### Command (Phase 2 - Smart Handoff)
```bash
weave cols query WeaveDocs "what is the latest Anthropic Claude release?" \
  --agents rag-agent search-agent \
  --handoff-if low_scores \
  --handoff-threshold 0.5
```

**Behavior**: Same as above, but RAG agent automatically triggers handoff without executing.

### Config (Phase 3 - Declarative)
```yaml
# agents/rag-or-search.yaml
name: rag-or-search
description: Try RAG first, fall back to web search
type: multi-agent

agents:
  - name: rag-agent
    config: rag-agent.yaml
    handoff_conditions:
      - condition: max_score_below
        threshold: 0.5
        next_agent: search-agent
      - condition: no_sources
        next_agent: search-agent
      - condition: always
        next_agent: null  # Return RAG result

  - name: search-agent
    config: search-agent.yaml
    handoff_conditions:
      - condition: always
        next_agent: null  # Final agent
```

```bash
weave cols query WeaveDocs "query" --multiagent rag-or-search
```

---

## Example 2: Multi-Collection Query Routing

### Scenario
User query could relate to code, docs, or API references. Classifier routes to appropriate specialist.

### Command
```bash
weave cols query "how to authenticate API requests" \
  --agents classifier-agent api-docs-agent code-agent \
  --collections APIDocs CodeExamples TechDocs
```

### Expected Flow

**Step 1: Classifier Agent**
```
Running classifier-agent...
  Analyzing query: "how to authenticate API requests"
  Query type: API Documentation
  Confidence: 92%

  Recommended next agent: api-docs-agent
```

**Step 2: API Docs Agent**
```
Running api-docs-agent...
  Querying collection: APIDocs
  Found 5 sources, max score: 0.87

  Response: "To authenticate API requests, include your API key in the Authorization header:

  Authorization: Bearer YOUR_API_KEY

  You can generate API keys in the dashboard under Settings > API Keys. [1][2]"
```

**Final Output**
```
Sources:
[1] APIDocs (weaviate-cloud) - Score: 87.3% - ID: api-auth-001
[2] APIDocs (mongodb-cloud) - Score: 82.1% - ID: api-auth-002

Response: To authenticate API requests, include your API key...

Execution Summary:
  Agents used: classifier-agent (routed to api-docs-agent), api-docs-agent (success)
  Collections queried: APIDocs
  Total time: 2.1s
```

---

## Example 3: Code Analysis → Fix Generation → Verification

### Scenario
User reports a bug. System should: analyze code, generate fix, verify the fix works.

### Command
```bash
weave cols query CodeBase "fix authentication timeout bug in auth.go" \
  --agents analysis-agent fix-agent verification-agent
```

### Expected Flow

**Step 1: Analysis Agent**
```
Running analysis-agent...
  Searching codebase for authentication logic
  Found 3 relevant files

  Analysis:
  - File: src/auth/auth.go:142
  - Issue: HTTP client created without timeout
  - Risk: Hanging connections on slow networks
  - Recommendation: Add timeout to http.Client initialization
```

**Step 2: Fix Agent**
```
Running fix-agent...
  Previous analysis identified timeout issue in auth.go:142

  Suggested Fix:

  // Before:
  client := &http.Client{}

  // After:
  client := &http.Client{
      Timeout: 30 * time.Second,
  }

  Explanation: This adds a 30-second timeout to prevent hanging connections.
```

**Step 3: Verification Agent**
```
Running verification-agent...
  Analyzing proposed fix...

  Verification Results:
  ✓ Fix addresses root cause (missing timeout)
  ✓ Follows Go best practices
  ✓ Timeout value (30s) is reasonable for API calls
  ✓ No side effects detected

  Recommendation: Apply fix
```

**Final Output**
```
Analysis:
  File: src/auth/auth.go:142
  Issue: HTTP client without timeout

Proposed Fix:
  client := &http.Client{
      Timeout: 30 * time.Second,
  }

Verification: ✓ APPROVED

Execution Summary:
  Agents used: analysis-agent, fix-agent, verification-agent (all succeeded)
  Total time: 5.7s
```

---

## Example 4: Error Recovery Chain

### Scenario
Query may fail for various reasons. Chain should try multiple strategies.

### Command
```bash
weave cols query TechDocs "kubernetes deployment strategies" \
  --agents semantic-search-agent keyword-search-agent fuzzy-search-agent \
  --handoff-if no_sources
```

### Expected Flow

**Step 1: Semantic Search Agent**
```
Running semantic-search-agent...
  Using embedding model: text-embedding-3-large
  ERROR: Embedding service unavailable (503)
  No results returned
```

**Step 2: Keyword Search Agent**
```
Running keyword-search-agent...
  Previous agent failed with embedding error
  Falling back to keyword search
  Query: "kubernetes deployment strategies"
  Found 4 matches
```

**Step 3: Fuzzy Search Agent** (Skipped - keyword search succeeded)

**Final Output**
```
Sources:
[1] TechDocs (weaviate-cloud) - Score: 76.2% (keyword match)
[2] TechDocs (mongodb-cloud) - Score: 71.8% (keyword match)

Response: Kubernetes supports several deployment strategies including
rolling updates, blue-green deployments, and canary releases...

Execution Summary:
  Agents used: semantic-search-agent (failed), keyword-search-agent (success)
  Fallback strategy: embedding → keyword search
  Total time: 2.8s
```

---

## Example 5: Parallel Agent Execution (Future)

### Scenario
Query multiple specialized agents simultaneously, merge best results.

### Command
```bash
weave cols query "best practices for API design" \
  --agents rag-agent,search-agent,code-agent \
  --mode parallel \
  --merge best
```

### Expected Flow

**Step 1: All Agents Execute Simultaneously**
```
Running 3 agents in parallel...

[rag-agent] Querying TechDocs...
  Found 5 sources, max score: 0.84

[search-agent] Searching web...
  Found 8 results from Stack Overflow, Medium, blogs

[code-agent] Searching code examples...
  Found 3 code samples with API implementations
```

**Step 2: Merge Results**
```
Merging responses from 3 agents...
  Strategy: best (highest-scored sources)

Selected sources:
  - 5 from rag-agent (scores: 0.84, 0.78, 0.76, 0.72, 0.69)
  - 3 from search-agent (scores: 0.91, 0.87, 0.82)
  - 2 from code-agent (scores: 0.79, 0.74)

Total: 10 sources sorted by score
```

**Final Output**
```
Sources (top 10):
[1] Search: api-design-patterns.com - Score: 91.2%
[2] Search: martinfowler.com - Score: 87.4%
[3] RAG: TechDocs (weaviate) - Score: 84.3%
[4] Search: stackoverflow.com - Score: 82.1%
[5] Code: examples/api/rest.go - Score: 79.5%
...

Response: Best practices for API design include:
1. Use RESTful conventions [1][3]
2. Version your APIs [2][4]
3. Implement proper authentication [5][7]
...

Execution Summary:
  Agents used: rag-agent, search-agent, code-agent (parallel)
  Sources merged: 10 from 3 agents
  Total time: 2.1s (parallel execution)
```

---

## Example 6: Conditional Branching (Advanced)

### Scenario
Different query types route to different agent paths.

### Config
```yaml
# agents/smart-router.yaml
name: smart-router
type: multi-agent

agents:
  - name: intent-classifier
    config: classifier-agent.yaml
    handoff_conditions:
      # Code questions → code path
      - condition: metadata.intent == "code_question"
        next_agent: code-agent

      # Documentation questions → RAG path
      - condition: metadata.intent == "docs_question"
        next_agent: rag-agent

      # General questions → search path
      - condition: metadata.intent == "general_question"
        next_agent: search-agent

      # Unknown → error
      - condition: always
        next_agent: error-agent

  - name: code-agent
    config: code-agent.yaml
    handoff_conditions:
      - condition: always
        next_agent: null

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
    handoff_conditions:
      - condition: always
        next_agent: null

  - name: error-agent
    config: error-agent.yaml
    handoff_conditions:
      - condition: always
        next_agent: null
```

### Command
```bash
weave cols query "how to implement rate limiting in Go" --multiagent smart-router
```

### Expected Flow
```
Running intent-classifier...
  Query: "how to implement rate limiting in Go"
  Intent detected: code_question
  Routing to: code-agent

Running code-agent...
  Searching code examples for: rate limiting Go
  Found 7 code samples

  Response: [code examples and implementation]

Execution Summary:
  Path: intent-classifier → code-agent
  Intent: code_question
  Total time: 3.4s
```

---

## Example 7: Summarization Pipeline

### Scenario
Query returns many results. Summarize progressively.

### Command
```bash
weave cols query TechDocs "vector database comparison" \
  --agents retrieval-agent extraction-agent summary-agent \
  --top_k 20
```

### Expected Flow

**Step 1: Retrieval Agent**
```
Running retrieval-agent...
  Found 20 sources about vector databases
  Passing to next agent for extraction
```

**Step 2: Extraction Agent**
```
Running extraction-agent...
  Received 20 sources from retrieval-agent
  Extracting key facts from each source...

  Extracted facts:
  - Weaviate: Cloud-native, GraphQL API, hybrid search
  - Milvus: High performance, GPU support, 10M+ vectors
  - Qdrant: Rust-based, filtering, payload support
  ...
```

**Step 3: Summary Agent**
```
Running summary-agent...
  Received extracted facts from 20 sources
  Generating comparative summary...

  Summary:
  Vector databases compared across 3 dimensions:

  Performance: Milvus > Qdrant > Weaviate
  Ease of use: Weaviate > Qdrant > Milvus
  Features: Qdrant > Weaviate > Milvus

  Recommendation: Choose based on your priority...
```

**Final Output**
```
Comparative Summary:
[Generated summary with key comparisons]

Sources: 20 documents analyzed
Agents: retrieval-agent → extraction-agent → summary-agent
Total time: 8.3s
```

---

## Configuration Examples

### Simple Fallback
```yaml
# agents/simple-fallback.yaml
name: simple-fallback
type: multi-agent

agents:
  - name: primary-agent
    config: rag-agent.yaml
    handoff_conditions:
      - condition: no_sources
        next_agent: fallback-agent
      - condition: always
        next_agent: null

  - name: fallback-agent
    config: search-agent.yaml
    handoff_conditions:
      - condition: always
        next_agent: null
```

### Error Recovery Chain
```yaml
# agents/error-recovery.yaml
name: error-recovery
type: multi-agent

agents:
  - name: semantic-search
    config: semantic-agent.yaml
    handoff_conditions:
      - condition: error
        next_agent: keyword-search
      - condition: always
        next_agent: null

  - name: keyword-search
    config: keyword-agent.yaml
    handoff_conditions:
      - condition: error
        next_agent: fuzzy-search
      - condition: always
        next_agent: null

  - name: fuzzy-search
    config: fuzzy-agent.yaml
    handoff_conditions:
      - condition: always
        next_agent: null
```

### Multi-Stage Pipeline
```yaml
# agents/analysis-pipeline.yaml
name: analysis-pipeline
type: multi-agent

agents:
  - name: retrieval
    config: retrieval-agent.yaml
    handoff_conditions:
      - condition: always
        next_agent: analysis

  - name: analysis
    config: analysis-agent.yaml
    handoff_conditions:
      - condition: always
        next_agent: summary

  - name: summary
    config: summary-agent.yaml
    handoff_conditions:
      - condition: always
        next_agent: null
```

---

## Testing Multi-Agent Workflows

### Unit Tests
```go
func TestAgentChain_RAGWithSearchFallback(t *testing.T) {
    // Setup: RAG agent returns no sources
    ragAgent := &MockAgent{
        Response: &AgentResponse{
            Sources: []SourceContext{},
            Content: "No relevant sources found",
        },
    }

    // Setup: Search agent returns web results
    searchAgent := &MockAgent{
        Response: &AgentResponse{
            Sources: []SourceContext{
                {Content: "Web result 1", Score: 0.89},
                {Content: "Web result 2", Score: 0.85},
            },
            Content: "Found web results",
        },
    }

    chain := NewAgentChain([]Agent{ragAgent, searchAgent})
    response, err := chain.Execute("test query", nil)

    require.NoError(t, err)
    assert.Equal(t, "Found web results", response.Content)
    assert.Equal(t, 2, len(response.Sources))
    assert.Equal(t, "search-agent", response.AgentName)
}
```

### Integration Tests
```bash
# Test RAG with search fallback
weave cols query WeaveDocs "nonexistent topic xyz123" \
  --agents rag-agent search-agent \
  --handoff-if no_sources

# Expected: Search agent provides results
# Verify: Output includes "Running search-agent..."
```

---

## Performance Considerations

### Sequential Execution
```
Agent 1: 2.1s
Agent 2: 1.8s (if handoff)
Agent 3: 1.5s (if handoff)
Total: Up to 5.4s
```

### Parallel Execution (Future)
```
Agent 1: 2.1s ─┐
Agent 2: 1.8s ─┼─ Merge: 0.3s
Agent 3: 1.5s ─┘
Total: 2.4s (max agent time + merge)
```

### Optimization Strategies
1. **Early exit**: Stop chain on first success
2. **Timeout per agent**: Prevent slow agents from blocking
3. **Caching**: Cache agent responses for repeated queries
4. **Lazy loading**: Only load agents when needed
5. **Parallel where possible**: Run independent agents simultaneously
