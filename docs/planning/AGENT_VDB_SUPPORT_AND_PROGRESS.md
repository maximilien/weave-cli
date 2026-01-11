# Multi-VDB Agent Support and Query Progress

**Status:** Planning
**Target Version:** 0.9.1 or 0.10.0
**Date:** 2026-01-11
**Author:** Claude Code

## Overview

Two enhancements to improve the RAG agent user experience:

1. **Multi-VDB Agent Support** - Extend `--agent` flag to work with all
   vector databases, not just Weaviate
2. **Query Progress Indicator** - Add `--progress` flag to show real-time
   progress during long-running queries

## Problem Statement

### Issue 1: Agent Support Limited to Weaviate

**Current Behavior:**

```bash
# Works ✅
weave cols query MyDocs "query" --agent rag-agent --db weaviate

# Fails ❌
weave cols query MyDocs "query" --agent rag-agent --db chroma
# Error: Agent execution is currently only supported for Weaviate databases

# Fails ❌
weave cols query MyDocs "query" --agent summarize-agent --db qdrant
# Error: Agent execution is currently only supported for Weaviate databases
```

**Impact:**

- Users with Chroma, Qdrant, Milvus, Neo4j, etc. cannot use RAG agents
- Limits adoption of the agent feature
- Inconsistent user experience across VDBs

### Issue 2: No Progress Feedback During Queries

**Current Behavior:**

```bash
# User runs query
weave cols query MyDocs "complex query" --agent rag-agent --top_k 20

# Terminal appears frozen for 5-10 seconds
# User doesn't know if:
# - Query is running
# - System is stuck
# - Network is slow
# - LLM is processing

# Finally gets output
## Answer
...
```

**Impact:**

- Poor UX during long queries (vector search + LLM processing)
- Users may think the CLI is frozen or broken
- No visibility into what's happening

## Proposed Solutions

### Solution 1: Generic Agent Query Function

Create a generic `QueryCollectionWithAgent` function that works with
all VDBs using the `vectordb.VectorDBClient` interface.

**Architecture:**

```
┌─────────────────────────────────────────────────────────────┐
│ Collection Query Command                                     │
│ src/cmd/collection/query.go                                  │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ├─ --agent flag specified?
                 │
                 ├─ Weaviate? ──► QueryWeaviateCollectionWithAgent()
                 │                (specialized, keeps current behavior)
                 │
                 └─ Other VDB? ─► QueryCollectionWithAgent()
                                  (NEW: generic agent query)
                                  │
                                  ├─ CreateVectorDBClient()
                                  ├─ client.Search() → QueryResult[]
                                  └─ ExecuteQueryWithAgent()
```

**Code Changes:**

1. **New Function:** `QueryCollectionWithAgent()` in
   `src/cmd/utils/collection.go`

```go
// QueryCollectionWithAgent queries a collection using generic VDB client
// and processes results through an agent
func QueryCollectionWithAgent(
    ctx context.Context,
    cfg *config.VectorDBConfig,
    collectionName, queryText string,
    options weaviate.QueryOptions,
    agentName string,
    outputFormat string,
) {
    // Create generic VDB client
    client, err := CreateVectorDBClient(cfg)
    if err != nil {
        PrintError(fmt.Sprintf("Failed to create client: %v", err))
        return
    }

    // Convert to vectordb.QueryOptions
    vdbOptions := &vectordb.QueryOptions{
        TopK:           options.TopK,
        Distance:       options.Distance,
        SearchMetadata: options.SearchMetadata,
        NoTruncate:     options.NoTruncate,
        UseBM25:        options.UseBM25,
    }

    // Execute query
    var results []*vectordb.QueryResult
    if vdbOptions.UseBM25 {
        results, err = client.SearchBM25(ctx, collectionName, queryText,
                                         vdbOptions)
    } else {
        results, err = client.Search(ctx, collectionName, queryText,
                                     vdbOptions)
    }

    if err != nil {
        PrintError(fmt.Sprintf("Query failed: %v", err))
        return
    }

    // Convert to weaviate.QueryResult format for agent
    weaviateResults := make([]weaviate.QueryResult, len(results))
    for i, r := range results {
        weaviateResults[i] = weaviate.QueryResult{
            ID:       r.Document.ID,
            Content:  r.Document.Content,
            Metadata: r.Document.Metadata,
            Score:    r.Score,
        }
    }

    // Execute through agent
    ExecuteQueryWithAgent(ctx, agentName, queryText, weaviateResults,
                         outputFormat)
}
```

2. **Update Command:** Modify `src/cmd/collection/query.go`

```go
// If agent is specified, use agent-enabled query path
if agentName != "" {
    switch dbConfig.Type {
    case config.VectorDBTypeCloud, config.VectorDBTypeLocal:
        // Weaviate: use specialized function (keeps current behavior)
        utils.QueryWeaviateCollectionWithAgent(ctx, dbConfig, collectionName,
                                               queryText, options, agentName,
                                               outputFormat)
    case config.VectorDBTypeMock:
        // Mock: use specialized function
        utils.QueryMockCollectionWithAgent(ctx, dbConfig, collectionName,
                                           queryText, options, agentName,
                                           outputFormat)
    default:
        // All other VDBs: use generic function
        utils.QueryCollectionWithAgent(ctx, dbConfig, collectionName,
                                       queryText, options, agentName,
                                       outputFormat)
    }
    return
}
```

3. **Optional:** Create `QueryMockCollectionWithAgent()` for mock database

**Testing:**

- Test agents with each VDB: Chroma, Qdrant, Milvus, Neo4j, Supabase,
  MongoDB, Pinecone
- Verify same agent behavior across all VDBs
- Ensure output format consistency
- Test --json, --output flags with all VDBs

### Solution 2: Query Progress Indicator

Add `--progress` flag to show real-time progress during query execution.

**Architecture:**

```
Progress Phases:
┌─────────────────┐
│ 1. Querying VDB │  "🔍 Searching collection..."
└─────────────────┘  "Found 5 results"
        │
        ▼
┌─────────────────┐
│ 2. Processing   │  "🤖 Processing with rag-agent..."
│    with Agent   │  "Building context from 5 sources..."
└─────────────────┘  "Generating response..."
        │
        ▼
┌─────────────────┐
│ 3. Complete     │  "✅ Done in 3.2s"
└─────────────────┘
```

**Implementation Options:**

#### Option A: Simple Progress Messages (Recommended)

```go
type ProgressReporter struct {
    enabled   bool
    startTime time.Time
}

func (p *ProgressReporter) Start(msg string) {
    if !p.enabled {
        return
    }
    p.startTime = time.Now()
    fmt.Fprintf(os.Stderr, "🔍 %s\n", msg)
}

func (p *ProgressReporter) Update(msg string) {
    if !p.enabled {
        return
    }
    elapsed := time.Since(p.startTime).Truncate(100 * time.Millisecond)
    fmt.Fprintf(os.Stderr, "   %s (%.1fs)\n", msg, elapsed.Seconds())
}

func (p *ProgressReporter) Complete(msg string) {
    if !p.enabled {
        return
    }
    elapsed := time.Since(p.startTime)
    fmt.Fprintf(os.Stderr, "✅ %s (%.2fs)\n", msg, elapsed.Seconds())
}
```

**Usage:**

```go
func QueryCollectionWithAgent(..., progress bool) {
    reporter := &ProgressReporter{enabled: progress}

    reporter.Start("Searching collection...")

    // Execute query
    results, err := client.Search(...)
    reporter.Update(fmt.Sprintf("Found %d results", len(results)))

    // Process with agent
    reporter.Update("Processing with agent...")
    reporter.Update(fmt.Sprintf("Building context from %d sources...",
                                len(results)))

    // Generate response
    reporter.Update("Generating response...")
    output, err := ragAgent.Execute(...)

    reporter.Complete("Done")

    // Print actual output
    fmt.Println(formatted)
}
```

**Output Example:**

```bash
$ weave cols query MyDocs "what is AI?" --agent rag-agent --progress

🔍 Searching collection...
   Found 5 results (0.8s)
   Processing with agent... (0.9s)
   Building context from 5 sources... (1.0s)
   Generating response... (1.2s)
✅ Done (3.4s)

## Answer

Artificial Intelligence (AI) refers to...
[rest of agent output]
```

#### Option B: Spinner with Status (More Visual)

```go
import "github.com/briandowns/spinner"

type ProgressSpinner struct {
    enabled bool
    spinner *spinner.Spinner
}

func (p *ProgressSpinner) Start(msg string) {
    if !p.enabled {
        return
    }
    p.spinner = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
    p.spinner.Suffix = " " + msg
    p.spinner.Start()
}

func (p *ProgressSpinner) Update(msg string) {
    if !p.enabled {
        return
    }
    p.spinner.Suffix = " " + msg
}

func (p *ProgressSpinner) Stop() {
    if p.enabled && p.spinner != nil {
        p.spinner.Stop()
    }
}
```

**Output Example:**

```bash
$ weave cols query MyDocs "what is AI?" --agent rag-agent --progress

⠋ Generating response with rag-agent...
```

(Spinner animates)

**Recommendation:** Start with Option A (simple messages) - it's:
- Easier to implement
- Works in all terminals
- No external dependencies
- Better for CI/CD environments
- More informative than spinners

#### Progress Integration Points

1. **Vector Search:**

```go
reporter.Start("Searching collection...")
results, err := client.Search(...)
reporter.Update(fmt.Sprintf("Found %d results (scores: %.2f-%.2f)",
                            len(results), minScore, maxScore))
```

2. **Context Building:**

```go
reporter.Update("Building context from sources...")
context := contextBuilder.Build(...)
reporter.Update(fmt.Sprintf("Using %d sources (%.0f chars)",
                            len(context.Sources), totalChars))
```

3. **LLM Processing:**

```go
reporter.Update("Generating response with " + agentConfig.LLM.Model + "...")
output, err := llmClient.Complete(...)
reporter.Update(fmt.Sprintf("Generated %d tokens", tokenCount))
```

4. **Formatting:**

```go
reporter.Update("Formatting output...")
formatted, err := agent.FormatOutput(...)
reporter.Complete("Done")
```

**Code Changes:**

1. **Add Flag:**

```go
// In src/cmd/collection/query.go init()
QueryCmd.Flags().Bool("progress", false,
                      "Show progress during query execution")
```

2. **Pass to Functions:**

```go
progress, _ := cmd.Flags().GetBool("progress")

// Pass to query functions
utils.QueryCollectionWithAgent(..., agentName, outputFormat, progress)
utils.QueryWeaviateCollectionWithAgent(..., agentName, outputFormat,
                                       progress)
```

3. **Update Agent Query Functions:**

```go
func ExecuteQueryWithAgent(
    ctx context.Context,
    agentName string,
    query string,
    results []weaviate.QueryResult,
    outputFormat string,
    progress bool,  // NEW
) {
    reporter := &ProgressReporter{enabled: progress}

    reporter.Start(fmt.Sprintf("Processing %d results with %s...",
                               len(results), agentName))

    // Load agent
    reporter.Update("Loading agent configuration...")
    agentConfig, err := agents.LoadAgent(agentName)
    // ... error handling

    // Create LLM client
    reporter.Update("Initializing LLM client...")
    llmClient, err := llm.NewOpenAIClient(apiKey)
    // ... error handling

    // Create agent
    reporter.Update("Creating RAG agent...")
    ragAgent, err := agents.NewRAGAgent(agentConfig, llmClient)
    // ... error handling

    // Build context
    reporter.Update(fmt.Sprintf("Building context from %d sources...",
                                len(results)))
    // ... context building

    // Execute
    reporter.Update("Generating response...")
    output, err := ragAgent.Execute(ctx, input)
    // ... error handling

    // Format
    reporter.Update("Formatting output...")
    formatted, err := ragAgent.FormatOutput(output)
    // ... error handling

    reporter.Complete("Done")

    // Print output
    fmt.Println(formatted)
}
```

**Verbose vs Progress:**

- `--verbose`: Shows debug details (GraphQL queries, internal state)
- `--progress`: Shows user-friendly progress updates
- Can use both together: `--verbose --progress`

Example with both:

```bash
$ weave cols query MyDocs "AI" --agent rag-agent --verbose --progress

🔍 Searching collection...
2026/01/11 12:00:00 🔍 DEBUG: GraphQL Query (nearText):
{
  Get {
    MyDocs(nearText: {concepts: ["AI"]})
    ...
  }
}
   Found 5 results (0.8s)
   Processing with agent... (0.9s)
2026/01/11 12:00:01 🔍 DEBUG: Building context with 5 sources
   Building context from 5 sources... (1.0s)
   Generating response... (1.2s)
2026/01/11 12:00:02 🔍 DEBUG: LLM response: 1234 tokens
✅ Done (3.4s)

## Answer
...
```

**Progress with JSON Output:**

When `--json` or `--output json` is specified, progress messages go to
**stderr** while JSON output goes to **stdout**. This allows:

1. **Clean JSON piping:**

```bash
# Progress to terminal, JSON to file
$ weave cols query MyDocs "AI" --agent rag-agent --json --progress \
  > output.json

🔍 Searching collection...     # stderr (visible)
   Found 5 results (0.8s)      # stderr (visible)
   Generating response...      # stderr (visible)
✅ Done (3.4s)                  # stderr (visible)

# output.json contains clean JSON (from stdout)
$ cat output.json
{
  "answer": "Artificial Intelligence...",
  "sources": [...]
}
```

2. **Programmatic parsing with progress feedback:**

```bash
# Progress visible, JSON piped to jq
$ weave cols query MyDocs "AI" --agent rag-agent --json --progress \
  | jq '.answer'

🔍 Searching collection...     # stderr (visible)
   Found 5 results (0.8s)      # stderr (visible)
✅ Done (3.4s)                  # stderr (visible)
"Artificial Intelligence..."   # stdout (parsed by jq)
```

3. **Auto-disable when piping:**

```bash
# Progress auto-disabled when stdout is piped (unless --progress explicit)
$ weave cols query MyDocs "AI" --agent rag-agent --json | jq
# No progress (auto-disabled because stdout is piped)

$ weave cols query MyDocs "AI" --agent rag-agent --json --progress | jq
# Progress shown (explicit flag overrides auto-disable)
```

**Implementation:**

```go
func (p *ProgressReporter) Start(msg string) {
    if !p.enabled {
        return
    }
    // Always write to stderr, never stdout
    fmt.Fprintf(os.Stderr, "🔍 %s\n", msg)
}

func (p *ProgressReporter) Update(msg string) {
    if !p.enabled {
        return
    }
    elapsed := time.Since(p.startTime).Truncate(100 * time.Millisecond)
    // Write to stderr
    fmt.Fprintf(os.Stderr, "   %s (%.1fs)\n", msg, elapsed.Seconds())
}
```

**TTY Detection:**

```go
import "golang.org/x/term"

// Auto-disable progress when stdout is not a TTY (unless explicit flag)
func shouldEnableProgress(explicitFlag bool) bool {
    if explicitFlag {
        return true  // User explicitly requested progress
    }

    // Auto-enable only if stdout is a TTY
    return term.IsTerminal(int(os.Stdout.Fd()))
}
```

## Implementation Plan

### Phase 1: Multi-VDB Agent Support (Priority: High)

**Tasks:**

1. Create `QueryCollectionWithAgent()` in `src/cmd/utils/collection.go`
2. Create `QueryMockCollectionWithAgent()` in `src/cmd/utils/mock.go`
3. Update `src/cmd/collection/query.go` to route to correct function
4. Add unit tests for new functions
5. Integration tests with each VDB
6. Update documentation

**Files to Modify:**

- `src/cmd/utils/collection.go` - Add QueryCollectionWithAgent
- `src/cmd/utils/mock.go` - Add QueryMockCollectionWithAgent
- `src/cmd/collection/query.go` - Update routing logic
- `configs/agents/README.md` - Update "works with all VDBs" in docs
- `docs/planning/RAG_AGENT_FEATURE.md` - Update status

**Testing Checklist:**

- [ ] Test with Chroma (local + cloud)
- [ ] Test with Qdrant (local + cloud)
- [ ] Test with Milvus (local + cloud)
- [ ] Test with Neo4j (local + Aura)
- [ ] Test with Supabase (local + cloud)
- [ ] Test with MongoDB (local + Atlas)
- [ ] Test with Pinecone (cloud)
- [ ] Test with Mock database
- [ ] Verify --json output works
- [ ] Verify --output yaml works
- [ ] Test all three agents (rag, qa, summarize)

**Estimated Effort:** 2-3 hours

### Phase 2: Query Progress Indicator (Priority: Medium)

**Tasks:**

1. Create `ProgressReporter` struct in `src/pkg/progress/`
2. Add `--progress` flag to collection query command
3. Integrate into `ExecuteQueryWithAgent()`
4. Integrate into `QueryCollection()` (for non-agent queries too!)
5. Add progress to vector search phase
6. Add progress to agent processing phase
7. Update documentation

**Files to Create:**

- `src/pkg/progress/reporter.go` - ProgressReporter implementation
- `src/pkg/progress/reporter_test.go` - Unit tests

**Files to Modify:**

- `src/cmd/collection/query.go` - Add --progress flag
- `src/cmd/utils/agent_query.go` - Add progress reporting
- `src/cmd/utils/collection.go` - Add progress to standard queries
- `src/cmd/utils/weaviate.go` - Add progress to Weaviate queries
- `docs/USER_GUIDE.md` - Document --progress flag

**Testing Checklist:**

- [ ] Progress shows for agent queries
- [ ] Progress shows for non-agent queries
- [ ] Progress disabled by default
- [ ] Progress works with --json (goes to stderr, JSON to stdout)
- [ ] Progress + --verbose work together
- [ ] Progress timing accurate
- [ ] Progress messages clear and helpful
- [ ] No progress output when piping to file
- [ ] Progress respects --output json/yaml
- [ ] Progress auto-disabled when stdout is not a TTY

**Estimated Effort:** 3-4 hours

## User Experience Examples

### Multi-VDB Agent Support

**Before (v0.9.0):**

```bash
# Only works with Weaviate
$ weave cols query MyDocs "query" --agent rag-agent --db weaviate
✅ Works

# Fails with other VDBs
$ weave cols query MyDocs "query" --agent rag-agent --db chroma
❌ Error: Agent execution is currently only supported for Weaviate databases
```

**After (v0.9.1):**

```bash
# Works with all VDBs
$ weave cols query MyDocs "query" --agent rag-agent --db weaviate
✅ Works

$ weave cols query MyDocs "query" --agent rag-agent --db chroma
✅ Works

$ weave cols query MyDocs "query" --agent qa-agent --db qdrant
✅ Works

$ weave cols query MyDocs "query" --agent summarize-agent --db neo4j
✅ Works
```

### Query Progress Indicator

**Before (v0.9.0):**

```bash
$ weave cols query MyDocs "complex question" --agent rag-agent
[5 seconds of silence - user wondering if it's frozen]
## Answer
...
```

**After (v0.9.1):**

```bash
$ weave cols query MyDocs "complex question" --agent rag-agent --progress
🔍 Searching collection...
   Found 10 results (0.7s)
   Processing with rag-agent... (0.8s)
   Building context from 10 sources... (0.9s)
   Generating response... (1.5s)
✅ Done (3.2s)

## Answer
...
```

**With JSON output:**

```bash
# Progress to stderr, JSON to stdout
$ weave cols query MyDocs "question" --agent rag-agent --json --progress

🔍 Searching collection...     # stderr
   Found 5 results (0.8s)      # stderr
   Generating response...      # stderr
✅ Done (2.4s)                  # stderr

{                              # stdout (JSON)
  "answer": "...",
  "sources": [...]
}

# Clean piping works perfectly
$ weave cols query MyDocs "question" --agent rag-agent --json --progress \
  | jq '.answer'

🔍 Searching collection...     # stderr (visible)
✅ Done (2.4s)                  # stderr (visible)
"The answer is..."             # stdout (piped to jq)
```

## Backward Compatibility

Both features are fully backward compatible:

1. **Multi-VDB Support:**
   - Weaviate queries work exactly as before
   - No changes to existing functionality
   - Only adds support for other VDBs

2. **Progress Indicator:**
   - Disabled by default (opt-in with --progress)
   - No changes to output when not enabled
   - --json output unaffected

## Documentation Updates

1. **README.md:**
   - Update "Works with all 10+ vector databases"
   - Add --progress flag examples

2. **configs/agents/README.md:**
   - Update compatibility section
   - Add examples with different VDBs
   - Document --progress flag

3. **docs/USER_GUIDE.md:**
   - Add progress indicator section
   - Add multi-VDB agent examples

4. **CHANGELOG.md:**
   - Document both features
   - Include migration notes (none needed)

## Success Metrics

1. **Multi-VDB Support:**
   - ✅ All 10+ VDBs support --agent flag
   - ✅ Same agent behavior across all VDBs
   - ✅ Zero breaking changes

2. **Progress Indicator:**
   - ✅ Clear feedback during long queries
   - ✅ Accurate timing information
   - ✅ Works in all terminal environments
   - ✅ No performance impact when disabled

## Questions for Discussion

1. Should progress be enabled by default for queries > 2 seconds?
2. Should we add progress to document ingestion too?
3. Should we show token counts in progress messages?
4. Should we add ETA for multi-document operations?

## Next Steps

1. Review this plan
2. Decide on priority (0.9.1 or 0.10.0)
3. Implement Phase 1 (Multi-VDB support)
4. Implement Phase 2 (Progress indicator)
5. Update documentation
6. Release

## References

- Original RAG Agent Feature: `docs/planning/RAG_AGENT_FEATURE.md`
- VectorDB Interface: `src/pkg/vectordb/types.go`
- Agent Query Code: `src/cmd/utils/agent_query.go`
- Collection Query Code: `src/cmd/collection/query.go`
