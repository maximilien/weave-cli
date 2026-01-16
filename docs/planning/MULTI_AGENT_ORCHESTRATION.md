# Multi-Agent Orchestration Planning

> **📚 Related Documents**: [Decision Points](DECISION_POINTS.md) | [Examples & Use Cases](MULTI_AGENT_EXAMPLES.md) | [Planning Index](README.md)

---

## Problem Statement

When performing RAG queries, if no relevant documents are found, the user gets a "no information available" response. However, the question might still be answerable by other means (e.g., web search, API calls, general knowledge).

**Goal**: Allow multiple agents to work together in sequence or in a graph since the handoff can skip to an agent in different position to the agents listed, where one agent can hand off to another based on results or conditions.

**Example Use Case**:
```bash
weave cols query WeaveDocs "how to configure VDB timeouts" \
  --agents rag-agent search-agent \
  --top_k 5
```

Flow:
1. RAG agent queries vector DB
2. If no relevant docs (score < threshold), hand off to search-agent
3. Search agent performs web search
4. Return combined or best result

---

## Proposed Approaches

### Approach 1: Inline Agent List (Prompt-Based Handoff)

**UX**:
```bash
weave cols query WeaveDocs "query" --agents rag-agent search-agent fallback-agent
```

**Behavior**:
- Execute agents in order specified
- Each agent's prompt contains handoff logic
- Agent decides whether to hand off based on its own results and which agent it prefers to handoff too
- Since agents' names are unique the handoff can be done with names

**Handoff Logic in Prompt**:
```yaml
# rag-agent.yaml
system_prompt: |
  You are a RAG agent that answers questions using document context.

  If no relevant documents are found (all sources below 0.5 score),
  respond with: HANDOFF:search-agent

  Otherwise, provide a detailed answer using the sources.
```

**Implementation**:
```go
type AgentChain struct {
    agents []Agent
    query  string
}

func (ac *AgentChain) Execute() (*AgentResponse, error) {
    var lastResponse *AgentResponse

    for _, agent := range ac.agents {
        response, err := agent.Run(ac.query, lastResponse)
        if err != nil {
            return nil, err
        }

        // Check for handoff signal
        if strings.HasPrefix(response.Content, "HANDOFF:") {
            continue // Next agent in chain
        }

        // Agent produced final answer
        return response, nil
    }

    return lastResponse, nil
}
```

**Pros**:
- ✅ Simple UX: just list agents
- ✅ Flexible: each agent controls handoff logic
- ✅ Easy to understand execution order
- ✅ No new config files needed

**Cons**:
- ❌ Handoff logic in prompts (brittle, depends on LLM behavior)
- ❌ No conditional branching (only linear sequence)
- ❌ Hard to share handoff conditions across agents
- ❌ Parsing HANDOFF: signals from LLM output is fragile

---

### Approach 2: Multi-Agent Config File (Declarative)

**UX**:
```bash
weave cols query WeaveDocs "query" --multiagent rag-with-fallback
```

**Config**: `agents/rag-with-fallback.yaml`
```yaml
name: rag-with-fallback
description: RAG with web search fallback
type: multi-agent

agents:
  - name: rag-agent
    config: rag-agent.yaml
    handoff_conditions:
      - condition: "no_relevant_sources"
        threshold: 0.5
        next_agent: search-agent
      - condition: "always"
        next_agent: null  # Success, return result

  - name: search-agent
    config: search-agent.yaml
    handoff_conditions:
      - condition: "always"
        next_agent: null  # Final agent

# Optional: Define handoff conditions
conditions:
  no_relevant_sources:
    type: score_threshold
    max_score_below: 0.5
    min_sources: 1
```

NOTE: Should we have a schema for this YAML definition since we'd want to make sure we can parse and make sense of it? Perhaps also a command to generate it using AI?

**Implementation**:
```go
type MultiAgentConfig struct {
    Name        string         `yaml:"name"`
    Description string         `yaml:"description"`
    Type        string         `yaml:"type"` // "multi-agent"
    Agents      []AgentStep    `yaml:"agents"`
    Conditions  map[string]Condition `yaml:"conditions"`
}

type AgentStep struct {
    Name              string             `yaml:"name"`
    Config            string             `yaml:"config"`
    HandoffConditions []HandoffCondition `yaml:"handoff_conditions"`
}

type HandoffCondition struct {
    Condition string  `yaml:"condition"` // Reference to conditions map
    Threshold float64 `yaml:"threshold,omitempty"`
    NextAgent string  `yaml:"next_agent"` // null = return, or agent name
}

type Condition struct {
    Type           string  `yaml:"type"` // score_threshold, no_sources, error
    MaxScoreBelow  float64 `yaml:"max_score_below,omitempty"`
    MinSources     int     `yaml:"min_sources,omitempty"`
}
```

**Pros**:
- ✅ Declarative and explicit
- ✅ Conditions defined in config (not prompts)
- ✅ Reusable multi-agent configurations
- ✅ Easy to test and validate
- ✅ Version control for agent workflows
- ✅ Can support complex conditions

**Cons**:
- ❌ More complex config files
- ❌ New file type to manage
- ❌ Requires condition evaluation engine
- ❌ Less flexible than prompt-based (need to anticipate all conditions)

---

### Approach 3: Hybrid (Both Inline + Config)

**UX Option 1 (Inline)**:
```bash
weave cols query WeaveDocs "query" --agents rag-agent search-agent
```

**UX Option 2 (Config)**:
```bash
weave cols query WeaveDocs "query" --multiagent rag-with-fallback
```

**Behavior**:
- Inline: Simple sequential execution with prompt-based handoff
- Config: Declarative with explicit conditions

**Implementation**:
```go
type AgentExecutor interface {
    Execute(query string, context *QueryContext) (*AgentResponse, error)
}

type InlineAgentChain struct {
    agents []Agent
}

type MultiAgentOrchestrator struct {
    config *MultiAgentConfig
    agents map[string]Agent
}

func (e *Executor) RunAgents(query string, agentNames []string, multiAgent string) {
    if multiAgent != "" {
        orchestrator := LoadMultiAgentConfig(multiAgent)
        return orchestrator.Execute(query, context)
    } else {
        chain := NewInlineAgentChain(agentNames)
        return chain.Execute(query, context)
    }
}
```

**Pros**:
- ✅ Best of both worlds
- ✅ Simple cases use inline
- ✅ Complex workflows use config
- ✅ Progressive disclosure of complexity

**Cons**:
- ❌ Two UX patterns to learn
- ❌ More code to maintain
- ❌ Potential confusion about when to use which

---

### Approach 4: Conditional Handoff Engine

**UX**:
```bash
weave cols query WeaveDocs "query" --agents rag-agent --fallback search-agent
```

**Behavior**:
- Primary agent runs first
- If conditions met (no results, error, low scores), fallback executes
- Simple binary: primary vs fallback

**Config**: Built-in handoff conditions
```go
type AgentExecutionConfig struct {
    PrimaryAgent   string
    FallbackAgent  string
    HandoffTrigger HandoffTrigger
}

type HandoffTrigger struct {
    NoSources       bool
    MaxScoreBelow   float64
    MinSourcesBelow int
    OnError         bool
}
```

**Pros**:
- ✅ Simple UX for common case (primary + fallback)
- ✅ Built-in condition evaluation
- ✅ No config files for simple cases

**Cons**:
- ❌ Limited to 2 agents (primary + fallback)
- ❌ No support for complex workflows (3+ agents, branching)
- ❌ Conditions hardcoded in CLI flags

---

## Recommended Approach

**Hybrid Approach (Approach 3)** with phased implementation:

AGREED!

### Phase 1: Inline Agent Chain (Simple)
Start with simple sequential execution:

```bash
weave cols query WeaveDocs "query" --agents rag-agent search-agent
```

**Features**:
- Sequential execution (agent1 → agent2 → agent3)
- No automatic handoff (always execute all agents)
- Each agent sees previous agent's response
- User picks best response or last response wins

**Implementation**:
```go
func executeAgentChain(query string, agentNames []string, context *QueryContext) (*AgentResponse, error) {
    var responses []*AgentResponse
    var lastResponse *AgentResponse

    for i, agentName := range agentNames {
        agent, err := LoadAgent(agentName)
        if err != nil {
            return nil, err
        }

        // Pass previous response as context to next agent
        response, err := agent.Run(query, context, lastResponse)
        if err != nil {
            log.Warnf("Agent %s failed: %v", agentName, err)
            continue
        }

        responses = append(responses, response)
        lastResponse = response
    }

    // Return last successful response
    return lastResponse, nil
}
```

### Phase 2: Smart Handoff (Conditional)
Add built-in conditions to skip/continue:

```bash
weave cols query WeaveDocs "query" \
  --agents rag-agent search-agent \
  --handoff-if "no_sources" \
  --handoff-threshold 0.5
```

**Features**:
- Early exit if agent succeeds
- Handoff conditions built-in (not in prompts)
- Common conditions: `no_sources`, `low_scores`, `error`

**Implementation**:
```go
type HandoffConfig struct {
    Condition string  // "no_sources", "low_scores", "error", "always"
    Threshold float64 // For score-based conditions
}

func shouldHandoff(response *AgentResponse, config HandoffConfig) bool {
    switch config.Condition {
    case "no_sources":
        return len(response.Sources) == 0
    case "low_scores":
        return response.MaxScore() < config.Threshold
    case "error":
        return response.Error != nil
    case "always":
        return true
    default:
        return false
    }
}
```

### Phase 3: Multi-Agent Config (Advanced)
Add declarative config for complex workflows:

```bash
weave cols query WeaveDocs "query" --multiagent rag-with-fallback
```

**Features**:
- Declarative YAML configs
- Complex conditions and branching
- Reusable workflows
- Conditional logic beyond simple handoff

---

## Implementation Plan

### Step 1: Agent Interface Update
Add support for previous response in agent execution:

```go
type Agent interface {
    Run(query string, context *QueryContext, previousResponse *AgentResponse) (*AgentResponse, error)
}

type AgentResponse struct {
    Content     string
    Sources     []SourceContext
    Error       error
    Metadata    map[string]interface{}
    AgentName   string
    ExecutionTime time.Duration
}
```

### Step 2: CLI Flag Support
Add `--agents` flag:

```go
// cmd/collections/query.go
queryCmd.Flags().StringSlice("agents", []string{}, "List of agents to execute in sequence")
queryCmd.Flags().String("agent", "", "Single agent to use (deprecated in favor of --agents)")
```

### Step 3: Agent Chain Executor
Create chain executor:

```go
// src/pkg/agents/chain.go
type AgentChain struct {
    agents []Agent
    config *ChainConfig
}

type ChainConfig struct {
    StopOnSuccess    bool
    HandoffCondition HandoffConfig
}

func (ac *AgentChain) Execute(query string, context *QueryContext) (*AgentResponse, error) {
    // Implementation from Phase 1
}
```

### Step 4: Testing
Create comprehensive tests:

```
src/pkg/agents/
  chain_test.go          # Basic chain execution
  chain_handoff_test.go  # Handoff conditions
  multiagent_test.go     # Multi-agent configs (Phase 3)
```

### Step 5: Documentation
Document multi-agent usage:

```
docs/
  agents/
    multi-agent.md        # Guide to multi-agent orchestration
    rag-with-fallback.md  # Example: RAG + web search
    agent-chaining.md     # How to chain agents
```

---

## Example Workflows

### Example 1: RAG with Web Search Fallback

**Inline (Phase 1)**:
```bash
weave cols query WeaveDocs "how to configure timeouts" \
  --agents rag-agent search-agent
```

**Smart (Phase 2)**:
```bash
weave cols query WeaveDocs "how to configure timeouts" \
  --agents rag-agent search-agent \
  --handoff-if low_scores \
  --handoff-threshold 0.5
```

**Config (Phase 3)**:
```yaml
# agents/rag-with-search.yaml
name: rag-with-search
type: multi-agent

agents:
  - name: rag-agent
    config: rag-agent.yaml
    handoff_conditions:
      - condition: low_scores
        threshold: 0.5
        next_agent: search-agent
      - condition: no_sources
        next_agent: search-agent
      - condition: always
        next_agent: null

  - name: search-agent
    config: search-agent.yaml
    handoff_conditions:
      - condition: always
        next_agent: null
```

### Example 2: Multi-Stage Analysis

**Use Case**: Analyze code → Suggest fixes → Verify fixes

```bash
weave cols query CodeDocs "fix authentication bug" \
  --agents analysis-agent fix-agent verification-agent
```

**Flow**:
1. Analysis agent finds the bug
2. Fix agent generates code fix
3. Verification agent validates the fix

### Example 3: Routing Agent

**Use Case**: Classify query type, route to specialist agent

```yaml
# agents/query-router.yaml
name: query-router
type: multi-agent

agents:
  - name: classifier-agent
    config: classifier-agent.yaml
    handoff_conditions:
      - condition: metadata.query_type == "code"
        next_agent: code-agent
      - condition: metadata.query_type == "docs"
        next_agent: rag-agent
      - condition: metadata.query_type == "api"
        next_agent: api-agent

  - name: code-agent
    config: code-agent.yaml

  - name: rag-agent
    config: rag-agent.yaml

  - name: api-agent
    config: api-agent.yaml
```

---

## Open Questions

1. **Response Merging**: Should we support merging responses from multiple agents?
   - Example: RAG agent finds 2 sources, search agent finds 3 more → merge into 5 total
   - Or: Always use last agent's response?

2. **Parallel Execution**: Should we support running agents in parallel?
   - Example: Query 3 different RAG agents simultaneously, pick best result
   - Performance benefit but increased complexity

3. **Cost Management**: How to handle costs when chaining expensive LLM agents?
   - Option: Add `--max-agents` limit
   - Option: Add cost estimation before execution

4. **Error Handling**: What happens if an agent in the chain fails?
   - Skip and continue to next agent?
   - Stop and return error?
   - Configurable per-agent?

5. **Context Passing**: What context should be passed between agents?
   - Full previous response?
   - Just the answer?
   - Metadata only?

6. **Observability**: How to show user what's happening?
   - Progress indicator: "Running rag-agent... No results, trying search-agent..."
   - Final summary: "Used 2 agents: rag-agent (no results), search-agent (success)"

---

## Next Steps

### Immediate (This Week)
1. ✅ Review this planning doc with users
2. ⬜ Gather feedback on preferred approach
3. ⬜ Create initial prototype for Phase 1 (inline agents)

### Short-term (Next Week)
1. ⬜ Implement Phase 1: Basic agent chaining
2. ⬜ Add tests for sequential execution
3. ⬜ Update CLI with `--agents` flag
4. ⬜ Document basic usage

### Medium-term (2-3 Weeks)
1. ⬜ Implement Phase 2: Smart handoff with conditions
2. ⬜ Add handoff condition evaluation
3. ⬜ Create example workflows (RAG + search)
4. ⬜ Add observability (progress indicators)

### Long-term (1-2 Months)
1. ⬜ Implement Phase 3: Multi-agent config files
2. ⬜ Add complex condition support
3. ⬜ Add routing and branching
4. ⬜ Performance optimization (parallel execution?)

---

## Alternative Ideas

### Idea 1: Agent Middleware Pattern
Agents can wrap other agents, adding behavior:

```go
agent := NewRAGAgent(config)
agent = WithFallback(agent, searchAgent, condition)
agent = WithRetry(agent, maxRetries)
agent = WithTimeout(agent, 30*time.Second)
```

### Idea 2: Graph-Based Orchestration
Define agent flow as a DAG:

```yaml
graph:
  nodes:
    - id: rag
      agent: rag-agent
    - id: search
      agent: search-agent
    - id: merge
      agent: merge-agent

  edges:
    - from: rag
      to: merge
    - from: search
      to: merge
      condition: rag.no_sources
```

### Idea 3: LLM-Driven Routing
Let an LLM decide which agent to use:

```bash
weave cols query WeaveDocs "query" \
  --router-agent smart-router \
  --available-agents rag-agent,search-agent,code-agent
```

Router agent analyzes query and picks best agent to handle it.

NOTE: I like this idea and wondering if we can add it to the plan with a smart-router.yaml agent example. This can be added at the end after the config multiagent if it makes sense.

---

## Conclusion

**Recommended Path Forward**:
1. Start with **Phase 1** (inline agent list) for simplicity
2. Gather user feedback on real-world use cases
3. Implement **Phase 2** (smart handoff) based on common patterns
4. Consider **Phase 3** (config files) only if complexity demands it

This approach balances **immediate value** (simple chaining) with **future flexibility** (declarative configs), while keeping the UX simple for common cases.
