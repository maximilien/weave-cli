# Multi-Agent Query Examples

Multi-agent orchestration allows you to chain multiple agents in sequence to process query results through different specialized agents.

## Basic Usage

### Single Agent (Existing Behavior)
```bash
# Query with single agent using --agent flag
weave cols query MyDocs "What is machine learning?" --agent rag-agent

# Or using shorthand
weave cols q MyDocs "machine learning" --agent rag-agent
```

### Multi-Agent Chain (New Feature)
```bash
# Query with multiple agents using --agents flag
weave cols query MyDocs "What is machine learning?" --agents rag-agent,qa-agent

# Chain multiple agents in sequence
weave cols q MyDocs "complex query" --agents rag-agent,summarize-agent,qa-agent
```

## How It Works

When you specify multiple agents with `--agents`, they execute in sequence:

1. **Agent 1** receives the original query and search results
2. **Agent 2** receives the output from Agent 1 (via PassContext)
3. **Agent 3** receives the output from Agent 2
4. Final output is from the last successful agent

### Chain Configuration (Phase 1)

The default chain configuration is:
- **PassContext**: `true` - Output from each agent flows to the next
- **FailFast**: `false` - Continue executing if an agent fails
- **StopOnSuccess**: `false` - Execute all agents in the chain

## Multi-Collection + Multi-Agent

Combine multi-collection queries with multi-agent processing:

```bash
# Query multiple collections and process through agent chain
weave cols query WeaveDocs WeaveImages "screenshot" \
  --agents rag-agent,summarize-agent \
  --top_k 5 --top_k_images 2

# Cross-VDB multi-collection with multi-agent
weave cols query AuctionDocs:weaviate-local AuctionImages:milvus-cloud "vintage cars" \
  --agents rag-agent,qa-agent \
  --top_k 3
```

## Use Cases

### 1. RAG → Summarization
Process comprehensive RAG results and then summarize them:
```bash
weave cols query LargeDocs "annual report analysis" \
  --agents rag-agent,summarize-agent \
  --top_k 20
```

### 2. RAG → QA
Generate comprehensive answer first, then extract specific answers:
```bash
weave cols query TechnicalDocs "how to configure SSL" \
  --agents rag-agent,qa-agent
```

### 3. Search → RAG → Summarize
Chain multiple processing steps:
```bash
weave cols query MultiModal "product features" \
  --agents search-agent,rag-agent,summarize-agent
```

## Progress Tracking

Use `--progress` flag to see agent execution progress:

```bash
weave cols query MyDocs "query" \
  --agents rag-agent,qa-agent \
  --progress

# Output:
# ⠋ Processing 10 results with 2 agents...
# ⠙ Loading agent 1/2: rag-agent...
# ⠹ Loading agent 2/2: qa-agent...
# ⠸ Creating agent chain...
# ⠼ Executing chain: rag-agent → qa-agent...
# ⠴ Formatting final output...
# ✓ Executed 2 agents in 3.5s
```

## Error Handling

### FailFast: false (Default)
If an agent fails, the chain continues:
```bash
# If agent1 fails, agent2 and agent3 still execute
weave cols query MyDocs "query" --agents agent1,agent2,agent3
```

The final output will be from the last successful agent.

### Agent Not Found
```bash
weave cols query MyDocs "query" --agents invalid-agent

# Error: Agent 'invalid-agent' not found. Use 'weave agents list' to see available agents.
```

### All Agents Fail
```bash
# If all agents fail, you'll get an error:
# Error: Agent chain execution failed: no agent in chain produced valid output
```

## JSON Output

Combine multi-agent with JSON output:
```bash
weave cols query MyDocs "query" \
  --agents rag-agent,summarize-agent \
  --output json

# Or with progress (outputs JSON Lines format):
weave cols query MyDocs "query" \
  --agents rag-agent,summarize-agent \
  --json --progress
```

## Limitations

1. **Phase 1**: Only sequential execution (no branching or conditional logic)
2. **Phase 1**: All agents must be RAG-compatible (receive RAGInput, output RAGOutput)
3. **Phase 1**: No custom chain configuration (uses defaults)

## Future Enhancements (Phase 2 & 3)

Coming soon:
- Custom chain configurations (FailFast, StopOnSuccess)
- Smart handoff conditions
- Agent-specific routing
- Declarative YAML configs
- Non-RAG agent support

## Implementation Details

### Chain Execution Flow

```
Query Results → Agent1 → Agent2 → Agent3 → Final Output
                  ↓         ↓         ↓
              Output1   Output2   Output3
                        (used as input to Agent2)
                                  (used as input to Agent3)
```

### Code Structure

- `src/pkg/agents/chain.go`: AgentChain implementation
- `src/pkg/agents/chain_test.go`: Comprehensive test suite
- `src/cmd/utils/agent_query.go`: CLI integration
- `src/cmd/collection/query.go`: Query command setup

### Testing

Run the agent chain tests:
```bash
go test -v ./src/pkg/agents -run TestAgentChain
```

## See Also

- [Multi-Agent Orchestration Planning](../planning/MULTI_AGENT_ORCHESTRATION.md)
- [Agent Configuration](../agents/README.md)
- [RAG Agent Documentation](../agents/RAG_AGENT.md)
