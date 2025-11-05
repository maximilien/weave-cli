# Weave CLI AI Agents - Design Document

## Overview

This document describes the design and implementation of AI Agents support in the Weave CLI through the `weave query` command. This feature enables users to interact with the CLI using natural language queries, which are automatically translated into appropriate weave-cli commands and bash operations.

## Recent Enhancements (v0.3.0+)

- **Opik/OpenTelemetry Integration**: Full LLM observability with automatic tracing
  - Cost tracking for all LLM calls with color-coded display
  - Token usage metrics (prompt, completion, total)
  - Direct link to Opik dashboard for detailed trace analysis
- **MCP Tool Schema Integration**: Planning agent receives full tool schemas
  - Better parameter validation and error messages
  - Required parameters automatically included in tool calls
  - Reduces execution errors from missing parameters
- **Enhanced Error Detection**: Smart detection of errors in command output
  - Checks for "error:", "failed to" patterns in output
  - Marks steps as failed even when MCP returns success
  - More accurate success rate reporting
- **Improved Step Progress Display**:
  - Shows duration on completion line instead of repeating description
  - Color-coded success/failure indicators
  - Real-time progress updates
- **JSON Syntax Highlighting**: Beautiful color-coded JSON output
  - Syntax highlighting for keys, strings, numbers, booleans
  - Better readability for structured data
- **Auto-Approval for Simple Commands**:
  - Simple `weave` commands auto-approved without user confirmation
  - Complex bash scripts still require approval for safety
  - Detects shell syntax (loops, pipes, conditionals)
- **Health Check Support**: Natural language health checks
  - "check health" automatically maps to `weave health check`
  - QueryAgent validates health queries as weave-related
  - PlanningAgent creates correct bash command execution

## Motivation

Currently, users need full knowledge of CLI commands, their structure, and parameters to use the Weave CLI effectively. The AI Agents feature simplifies this by:

1. **Natural Language Interface**: Users can describe what they want to do in plain English
2. **Intelligent Command Planning**: AI agents plan and execute the appropriate sequence of commands
3. **Error Handling**: Graceful handling of ambiguous queries with user feedback
4. **Comprehensive Reporting**: Clear, human-readable output of operations performed

## Command Interface

### Basic Usage

```bash
# Full command
weave query "show me all my collections"

# Short alias
weave q "show me all my empty collections"
```

### Command Flags

```bash
weave query [query-text] [flags]

Flags:
  --dry-run              Show planned commands without executing
  --verbose, -v          Show detailed execution information
  --no-confirm           Skip confirmation for destructive operations
  --output, -o string    Output format (text|json|yaml) (default "text")
  --model string         LLM model to use (default from config)
  --max-retries int      Maximum retry attempts for failed operations (default 3)
```

## Architecture

### Agent System Overview

The system uses the OpenAI Agent Framework golang SDK with the following specialized agents:

```
┌─────────────────┐
│   User Query    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  QueryAgent     │ ◄─── Validates and fixes query
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ PlanningAgent   │ ◄─── Plans command sequence
└────────┬────────┘
         │
         ├──────────┬──────────┐
         ▼          ▼          ▼
    ┌────────┐ ┌────────┐ ┌────────┐
    │ Weave  │ │  Bash  │ │ Output │
    │ Agent  │ │ Agent  │ │ Agent  │
    └───┬────┘ └───┬────┘ └───┬────┘
        │          │          │
        └──────────┴──────────┘
                   │
                   ▼
         ┌─────────────────┐
         │  ReportAgent    │ ◄─── Aggregates results
         └────────┬────────┘
                  │
                  ▼
         ┌─────────────────┐
         │   EvalAgent     │ ◄─── Evaluates success + metrics
         └─────────────────┘
```

### Agent Responsibilities

#### 1. QueryAgent

**Purpose**: Validates and normalizes user queries

**Responsibilities**:
- Parse user's natural language query
- Fix grammar, spelling, and ambiguities
- Determine if query is weave-cli related
- Reject non-weave queries with helpful message
- Extract intent and key parameters

**Input**: Raw user query string
**Output**: Fixed query + weave-cli relevance boolean

**Example**:
```go
type QueryAgentInput struct {
    Query string `json:"query"`
}

type QueryAgentOutput struct {
    IsWeaveQuery bool   `json:"is_weave_query"`
    FixedQuery   string `json:"fixed_query"`
    Intent       string `json:"intent"` // "list", "create", "query", "delete", etc.
    Confidence   float64 `json:"confidence"`
    Reason       string `json:"reason"` // Why it is/isn't a weave query
}
```

#### 2. PlanningAgent

**Purpose**: Plans the execution strategy

**Responsibilities**:
- Analyze fixed query and intent
- Identify required weave-cli commands
- Determine prerequisite bash operations
- Plan parameter gathering strategies
- Create execution DAG (Directed Acyclic Graph)
- Identify potential failure points
- Plan confirmation points for destructive operations

**Input**: QueryAgent output
**Output**: Execution plan with ordered steps

**Example**:
```go
type ExecutionStep struct {
    Type        string                 `json:"type"` // "bash", "weave", "confirm"
    Command     string                 `json:"command"`
    Description string                 `json:"description"`
    Args        []string              `json:"args,omitempty"`
    Params      map[string]interface{} `json:"params,omitempty"`
    DependsOn   []int                 `json:"depends_on,omitempty"` // Step indices
    Optional    bool                   `json:"optional"`
    Destructive bool                   `json:"destructive"`
}

type ExecutionPlan struct {
    Steps       []ExecutionStep `json:"steps"`
    Summary     string          `json:"summary"`
    Warnings    []string        `json:"warnings,omitempty"`
    Estimations struct {
        Duration string `json:"duration"`
        Risk     string `json:"risk"` // "low", "medium", "high"
    } `json:"estimations"`
}
```

#### 3. WeaveAgent

**Purpose**: Executes weave-cli commands via weave-mcp

**Responsibilities**:
- Interface with weave-mcp stdio binary
- Execute weave-cli commands
- Capture command output and errors
- Infer missing parameters when possible
- Validate required parameters
- Handle command failures gracefully

**Integration**:
- Uses weave-mcp stdio binary at `/Users/maximilien/github/maximilien/weave-mcp/bin/weave-mcp-stdio`
- Communicates via MCP protocol
- Supports all weave-mcp tools (list_collections, create_document, etc.)

**Example**:
```go
type WeaveAgentCommand struct {
    Tool       string                 `json:"tool"`
    Arguments  map[string]interface{} `json:"arguments"`
    Timeout    time.Duration         `json:"timeout,omitempty"`
}

type WeaveAgentResult struct {
    Success    bool                   `json:"success"`
    Output     interface{}           `json:"output"`
    Error      string                `json:"error,omitempty"`
    Duration   time.Duration         `json:"duration"`
    Retries    int                   `json:"retries"`
}
```

#### 4. BashAgent

**Purpose**: Executes bash commands and captures output

**Responsibilities**:
- Execute bash commands safely
- Capture stdout, stderr
- Handle command timeouts
- Validate command safety
- Suggest installations for missing tools
- Platform-specific handling (macOS suggestions with brew)

**Safety Features**:
- Command whitelist/blacklist
- Dry-run mode support
- Safe command validation
- Path traversal protection

**Example**:
```go
type BashCommand struct {
    Command     string        `json:"command"`
    Args        []string      `json:"args"`
    WorkingDir  string        `json:"working_dir,omitempty"`
    Timeout     time.Duration `json:"timeout,omitempty"`
    Environment map[string]string `json:"environment,omitempty"`
}

type BashResult struct {
    Success    bool          `json:"success"`
    Stdout     string        `json:"stdout"`
    Stderr     string        `json:"stderr"`
    ExitCode   int           `json:"exit_code"`
    Duration   time.Duration `json:"duration"`
}
```

#### 5. OutputAgent

**Purpose**: Formats and displays information to users

**Responsibilities**:
- Format plan for user review
- Display intermediate steps
- Show command execution progress
- Format command results
- Provide context for each step
- Handle different output formats (text/json/yaml)

**Features**:
- Progress indicators
- Color-coded output
- Emoji support (consistent with weave-cli style)
- Structured logging
- Real-time streaming for long operations

#### 6. ReportAgent

**Purpose**: Creates comprehensive operation reports

**Responsibilities**:
- Aggregate all command outputs
- Create human-readable summaries
- Generate detailed reports
- Track success/failure rates
- Provide actionable next steps
- Export reports in multiple formats

**Report Sections**:
1. Executive Summary
2. Operations Performed
3. Results by Command
4. Warnings and Errors
5. Recommendations
6. Next Steps

**Example**:
```go
type OperationReport struct {
    QueryIntent     string                 `json:"query_intent"`
    ExecutedSteps   int                   `json:"executed_steps"`
    SuccessfulSteps int                   `json:"successful_steps"`
    FailedSteps     int                   `json:"failed_steps"`
    StartTime       time.Time             `json:"start_time"`
    EndTime         time.Time             `json:"end_time"`
    Duration        time.Duration         `json:"duration"`
    Commands        []CommandReport       `json:"commands"`
    Summary         string                `json:"summary"`
    Recommendations []string              `json:"recommendations,omitempty"`
    NextSteps       []string              `json:"next_steps,omitempty"`
}

type CommandReport struct {
    Type        string                 `json:"type"` // "bash", "weave"
    Command     string                 `json:"command"`
    Success     bool                   `json:"success"`
    Output      string                 `json:"output"`
    Error       string                 `json:"error,omitempty"`
    Duration    time.Duration         `json:"duration"`
}
```

#### 7. EvalAgent

**Purpose**: Evaluates query execution and tracks metrics

**Responsibilities**:
- Verify output matches query intent
- Track LLM invocation metrics
- Generate evaluation statistics
- Integrate with Opik for monitoring
- Calculate success rates
- Track latencies and costs

**Integration with Opik**:
```go
type OpikConfig struct {
    APIKey     string `json:"api_key"`
    ProjectID  string `json:"project_id"`
    Endpoint   string `json:"endpoint,omitempty"`
    Enabled    bool   `json:"enabled"`
}

type EvaluationMetrics struct {
    QueryID          string        `json:"query_id"`
    Success          bool          `json:"success"`
    IntentMatched    bool          `json:"intent_matched"`
    LLMInvocations   int           `json:"llm_invocations"`
    TotalTokens      int           `json:"total_tokens"`
    PromptTokens     int           `json:"prompt_tokens"`
    CompletionTokens int           `json:"completion_tokens"`
    TotalCost        float64       `json:"total_cost"`
    Latency          time.Duration `json:"latency"`
    ErrorRate        float64       `json:"error_rate"`
    UserSatisfaction *float64      `json:"user_satisfaction,omitempty"`
}
```

## Implementation Structure

### Directory Layout

```
src/
├── cmd/
│   ├── query/
│   │   ├── query.go              # Main query command
│   │   ├── flags.go              # Command flags
│   │   └── output.go             # Output formatting
│   └── root.go                   # Register query command
├── pkg/
│   ├── agents/
│   │   ├── agent.go              # Base agent interface
│   │   ├── query_agent.go        # Query validation agent
│   │   ├── planning_agent.go     # Planning agent
│   │   ├── weave_agent.go        # Weave CLI execution agent
│   │   ├── bash_agent.go         # Bash execution agent
│   │   ├── output_agent.go       # Output formatting agent
│   │   ├── report_agent.go       # Report generation agent
│   │   └── eval_agent.go         # Evaluation agent
│   ├── mcp/
│   │   ├── client.go             # MCP client for weave-mcp
│   │   ├── stdio.go              # stdio transport
│   │   └── tools.go              # MCP tool definitions
│   ├── llm/
│   │   ├── client.go             # LLM client interface
│   │   ├── openai.go             # OpenAI implementation
│   │   └── types.go              # Common types
│   ├── opik/
│   │   ├── client.go             # Opik client
│   │   ├── metrics.go            # Metrics tracking
│   │   └── reporting.go          # Opik reporting
│   └── executor/
│       ├── executor.go           # Main execution engine
│       ├── plan.go               # Plan execution
│       └── context.go            # Execution context
tests/
├── query_test.go                 # Query command tests
├── agents_test.go                # Agent unit tests
└── integration/
    ├── query_integration_test.go # Integration tests
    └── examples_test.go          # Example-based tests
```

### Core Interfaces

```go
// Agent is the base interface for all agents
type Agent interface {
    Name() string
    Execute(ctx context.Context, input interface{}) (interface{}, error)
}

// MCPClient interface for weave-mcp interaction
type MCPClient interface {
    CallTool(ctx context.Context, tool string, args map[string]interface{}) (interface{}, error)
    ListTools(ctx context.Context) ([]MCPTool, error)
    Close() error
}

// LLMClient interface for LLM interactions
type LLMClient interface {
    Complete(ctx context.Context, prompt string, opts ...Option) (string, error)
    CompleteStructured(ctx context.Context, prompt string, schema interface{}, opts ...Option) (interface{}, error)
    StreamComplete(ctx context.Context, prompt string, opts ...Option) (<-chan string, error)
}

// Executor orchestrates agent execution
type Executor interface {
    Execute(ctx context.Context, query string) (*OperationReport, error)
    DryRun(ctx context.Context, query string) (*ExecutionPlan, error)
}
```

## User Interaction Flow

The query command follows a consistent template for user interaction:

### 1. Early Rejection (if applicable)

```
❌ Unable to process query

The query "how do I bake a cake" is not related to weave-cli operations.

Weave CLI helps you manage vector databases. Try queries like:
  • "show me all my collections"
  • "add documents from /path/to/docs to MyCollection"
  • "find empty collections"
```

### 2. Plan Display

```
📋 Query Plan

Intent: List all empty collections
Estimated time: 5-10 seconds
Risk level: low

Steps:
  1. [weave] List all collections
  2. [bash] Count documents in each collection
  3. [output] Display empty collections

⚠️  This operation will query all collections in your database.

Proceed? [Y/n]:
```

### 3. Execution Progress

```
⏳ Step 1: Create a new text collection named TestDocs
✓ Step 1 completed (292ms)
  {
    "description": "Collection for storing test documents",
    "name": "TestDocs",
    "status": "created",
    "type": "text"
  }

⏳ Step 2: Create a new image collection named TestImages
✓ Step 2 completed (288ms)
  {
    "description": "Collection for storing test images",
    "name": "TestImages",
    "status": "created",
    "type": "image"
  }
```

### 4. Results Display

```
📊 Results

Empty Collections (3):
  • TestDocs
  • Archive
  • Staging

Collections Summary:
  Total: 15 collections
  Empty: 3 collections
  With documents: 12 collections
  Total documents: 127,456
```

### 5. Summary Report (discrete)

```
✅ Query completed successfully

Operations: 17 commands executed
Duration: 8.4 seconds
Success rate: 100%

💡 Tip: Use `weave cols del COLLECTION` to delete empty collections
```

### 6. AI Metrics (always shown)

```
📋 AI Metrics
  LLM calls: 3
  Tokens: 3,353 (prompt: 2,956, completion: 397)
  Cost: $0.0114
  💡 View detailed traces in Opik dashboard: https://www.comet.com/opik
```

**Color-Coded Cost Display**:
- Green: < $0.01 (low cost)
- Yellow: $0.01 - $0.10 (medium cost)
- Red: > $0.10 (high cost)

## Example Query Workflows

### Example 1: List All Collections

**Query**: `weave query "show me all my collections"`

**Plan**:
```
1. [weave] list_collections
```

**Execution**:
```bash
$ weave q "show me all my collections"

📋 Query Plan
Intent: List collections
Steps: 1 weave command

✓ Listing collections...

📊 Collections (5):
  • WeaveDocs (text, 1,234 docs)
  • WeaveImages (image, 567 docs)
  • TestCollection (text, 0 docs)
  • Archive (text, 89 docs)
  • Staging (text, 0 docs)

✅ Query completed (1.2s)
```

### Example 2: Find Empty Collections

**Query**: `weave q "show me all my empty collections"`

**Plan**:
```
1. [weave] list_collections
2. [bash] For each collection, get document count
3. [output] Filter and display empty collections
```

**Execution**:
```bash
$ weave q "show me all my empty collections"

📋 Query Plan
Intent: Find empty collections
Steps: 1 weave + 5 bash commands

✓ Listing collections...
⏳ Checking sizes... [=====>] 5/5

📊 Empty Collections (2):
  • TestCollection
  • Staging

💡 Tip: Remove with `weave cols del COLLECTION`

✅ Query completed (3.4s)
```

### Example 3: Batch Add Files

**Query**: `weave query "add files in /tmp/test to my TestDocs and images to TestImages"`

**Plan**:
```
1. [bash] List files in /tmp/test
2. [bash] Classify files by type (text/image/PDF)
3. [weave] Create text documents in TestDocs
4. [weave] Create image documents in TestImages
5. [weave] Process PDFs (text → TestDocs, images → TestImages)
6. [output] Display summary report
```

**Execution**:
```bash
$ weave q "add files in /tmp/test to my TestDocs and images to TestImages"

📋 Query Plan
Intent: Batch add files to collections
Steps: 2 bash + 15 weave commands
Risk: medium (will create 23 documents)

Files to process:
  Text files: 8 (→ TestDocs)
  Images: 12 (→ TestImages)
  PDFs: 3 (text → TestDocs, images → TestImages)

Proceed? [Y/n]: y

⏳ Processing files...

✓ Text files [========] 8/8 (6.2s)
  ✓ doc1.txt → TestDocs
  ✓ doc2.md → TestDocs
  ...

✓ Images [========] 12/12 (4.8s)
  ✓ image1.jpg → TestImages
  ✓ image2.png → TestImages
  ...

✓ PDFs [========] 3/3 (12.3s)
  ✓ report.pdf → 5 text chunks (TestDocs) + 3 images (TestImages)
  ⚠️  presentation.pdf → CMYK format (converted)
  ✓ memo.pdf → 2 text chunks (TestDocs)

📊 Summary

Successfully added:
  TestDocs: 15 documents (8 text + 7 PDF chunks)
  TestImages: 15 documents (12 images + 3 PDF images)

Errors: 0
Duration: 23.7s

Collection Status:
  • TestDocs: 15 documents
  • TestImages: 15 documents

✅ Batch operation completed

💡 Use `weave docs ls TestDocs -S` for detailed summaries

✅ Query completed (24.1s)
```

## Configuration

### Quick Start (No config.yaml needed!)

Just set your credentials in `.env`:

```bash
# Required for AI agents
OPENAI_API_KEY="sk-proj-your-key"
WEAVIATE_URL="https://your-cluster.weaviate.cloud"
WEAVIATE_API_KEY="your-api-key"

# Optional for AI agents
WEAVE_MCP_STDIO_PATH="/path/to/weave-mcp/bin/weave-mcp-stdio"
OPIK_API_KEY="your-opik-key"  # For LLM observability
```

That's it! Defaults work great:

- Model: `gpt-4o`
- Max retries: 3
- Timeout: 300s
- Opik: Auto-enabled if `OPIK_API_KEY` is set

### Advanced Configuration (Optional)

Override defaults via environment variables:

```bash
export OPENAI_MODEL="gpt-4"              # Change default model
export WEAVE_MCP_STDIO_PATH="/path"     # MCP binary location
export OPIK_API_KEY="key"                # Enable Opik tracing
```

Or create `config.yaml` for persistent customization:

```yaml
ai:
  model: gpt-4o                   # LLM model to use
  max_retries: 3                  # Max retry attempts
  timeout: 300                    # Timeout in seconds
  enable_opik: true               # Enable Opik observability

# Note: Most settings have sensible defaults
# Only customize what you need!
    require_confirmation_for:
      - delete
      - update
      - batch_operations
```

### Environment Variables

```bash
# Required
OPENAI_API_KEY=sk-proj-...

# Optional
OPIK_API_KEY=opik-...
WEAVE_MCP_PATH=/path/to/weave-mcp-stdio
```

## Error Handling

### Query Not Weave-Related

```
❌ Cannot process query

Your query appears to be about [topic], which is outside the scope of weave-cli.

Weave CLI helps you:
  • Manage collections
  • Create and query documents
  • Process PDFs and images
  • Search vector databases

Try rephrasing your query or see `weave help` for available commands.
```

### Missing Dependencies

```
⚠️  Missing required tool: pdftotext

To process PDFs, please install poppler:
  macOS: brew install poppler
  Linux: sudo apt-get install poppler-utils

Alternatively, use --skip-all-images to extract text only.

Abort query? [Y/n]:
```

### Partial Failure

```
⚠️  Partial Success

Completed: 8/10 operations
Failed: 2 operations

Failed operations:
  1. Create document from corrupted.pdf
     Error: Unable to extract text from PDF

  2. Add large_image.jpg to TestImages
     Error: Image exceeds size limit (15MB > 10MB)

✓ Successfully processed 8 files

Would you like to:
  1. Retry failed operations
  2. Continue with successful operations
  3. Rollback all changes

Choice [1/2/3]:
```

## Testing Strategy

### Unit Tests

Test each agent independently with mocked dependencies:

```go
func TestQueryAgent_ValidWeaveQuery(t *testing.T) {
    agent := NewQueryAgent(mockLLM)

    input := &QueryAgentInput{
        Query: "show me all collections",
    }

    output, err := agent.Execute(context.Background(), input)
    require.NoError(t, err)

    result := output.(*QueryAgentOutput)
    assert.True(t, result.IsWeaveQuery)
    assert.Equal(t, "list", result.Intent)
    assert.Greater(t, result.Confidence, 0.9)
}
```

### Integration Tests

Test full query workflows using example queries:

```go
func TestQuery_ListEmptyCollections(t *testing.T) {
    // Setup test environment
    client := setupTestClient(t)
    defer client.Close()

    // Create test collections
    createTestCollection(t, client, "EmptyCol", 0)
    createTestCollection(t, client, "FullCol", 100)

    // Execute query
    cmd := exec.Command("weave", "q", "show me empty collections")
    output, err := cmd.CombinedOutput()
    require.NoError(t, err)

    // Verify output
    assert.Contains(t, string(output), "EmptyCol")
    assert.NotContains(t, string(output), "FullCol")
}
```

### Example-Based Tests

Use the provided examples as test cases:

```go
func TestQuery_Examples(t *testing.T) {
    examples := []struct {
        name     string
        query    string
        validate func(t *testing.T, output string, err error)
    }{
        {
            name:  "list_all_collections",
            query: "show me all my collections",
            validate: func(t *testing.T, output string, err error) {
                require.NoError(t, err)
                assert.Contains(t, output, "Collections")
            },
        },
        // More examples...
    }

    for _, ex := range examples {
        t.Run(ex.name, func(t *testing.T) {
            output, err := executeQuery(ex.query)
            ex.validate(t, output, err)
        })
    }
}
```

## Dependencies

### Go Modules

```go
require (
    github.com/openai/openai-go v0.x.x           // OpenAI client
    github.com/anthropics/anthropic-go v0.x.x    // Anthropic client (alternative)
    github.com/modelcontextprotocol/sdk-go v1.x.x // MCP SDK
    github.com/opik/opik-go v0.x.x              // Opik client
    github.com/spf13/cobra v1.8.0               // CLI framework
    github.com/spf13/viper v1.18.2              // Configuration
    github.com/fatih/color v1.16.0              // Color output
    github.com/schollz/progressbar/v3 v3.14.1   // Progress bars
    github.com/stretchr/testify v1.8.4          // Testing
)
```

### External Tools

- **Ghostscript**: For PDF conversion (optional)
- **ImageMagick**: For image processing (optional)
- **poppler**: For PDF text extraction (recommended)

## Implementation Phases

### Phase 1: Core Infrastructure (Week 1)

- [ ] Set up basic query command structure
- [ ] Implement base agent interface
- [ ] Create MCP client for weave-mcp integration
- [ ] Implement LLM client wrapper (OpenAI)
- [ ] Add configuration support

### Phase 2: Core Agents (Week 2)

- [ ] Implement QueryAgent
- [ ] Implement PlanningAgent
- [ ] Implement WeaveAgent
- [ ] Implement BashAgent
- [ ] Add unit tests for all agents

### Phase 3: Output & Reporting (Week 3)

- [ ] Implement OutputAgent
- [ ] Implement ReportAgent
- [ ] Add progress indicators
- [ ] Format output templates
- [ ] Add user confirmation flows

### Phase 4: Evaluation & Monitoring (Week 4)

- [ ] Implement EvalAgent
- [ ] Integrate Opik for metrics
- [ ] Add cost tracking
- [ ] Implement success rate monitoring
- [ ] Add verbose metrics output

### Phase 5: Testing & Documentation (Week 5)

- [ ] Write comprehensive unit tests
- [ ] Create integration tests
- [ ] Test all example workflows
- [ ] Write user documentation
- [ ] Update README and TODOs
- [ ] Create demo recordings

### Phase 6: Polish & Release (Week 6)

- [ ] Error handling improvements
- [ ] Performance optimization
- [ ] Security review
- [ ] Final testing
- [ ] Documentation review
- [ ] Release v0.3.0

## Security Considerations

1. **Command Injection Prevention**
   - Whitelist safe bash commands
   - Validate all user inputs
   - Escape special characters
   - Use parameterized command execution

2. **Path Traversal Protection**
   - Validate file paths
   - Restrict to allowed directories
   - Check for symbolic links
   - Sanitize user-provided paths

3. **Confirmation for Destructive Operations**
   - Require explicit user confirmation
   - Display clear warnings
   - Allow dry-run mode
   - Log all destructive operations

4. **API Key Security**
   - Store keys in environment variables
   - Never log API keys
   - Rotate keys regularly
   - Use least-privilege access

5. **Rate Limiting**
   - Limit LLM API calls
   - Throttle bash operations
   - Respect weave-mcp rate limits
   - Implement exponential backoff

## Performance Considerations

1. **Parallel Execution**
   - Execute independent steps in parallel
   - Use worker pools for batch operations
   - Limit concurrent operations
   - Monitor resource usage

2. **Caching**
   - Cache collection listings
   - Cache MCP tool schemas
   - Cache LLM responses for similar queries
   - Implement TTL for cached data

3. **Streaming**
   - Stream large outputs
   - Use progress indicators
   - Support cancellation
   - Handle timeouts gracefully

4. **Resource Limits**
   - Set operation timeouts
   - Limit batch sizes
   - Monitor memory usage
   - Implement circuit breakers

## Future Enhancements

1. **Multi-Model Support**
   - Support for Claude, Gemini, etc.
   - Model selection based on task
   - Fallback to alternative models
   - Cost optimization

2. **Query History**
   - Store successful queries
   - Learn from user patterns
   - Suggest similar queries
   - Export query history

3. **Query Templates**
   - Pre-defined query templates
   - Custom user templates
   - Template sharing
   - Template marketplace

4. **Interactive Mode**
   - Multi-turn conversations
   - Context-aware follow-ups
   - Clarification questions
   - Progressive refinement

5. **Advanced Planning**
   - Query optimization
   - Cost-based execution planning
   - Parallel execution optimization
   - Caching strategies

## Success Metrics

1. **User Adoption**
   - Number of queries executed
   - Unique users
   - Query success rate
   - User satisfaction scores

2. **Performance**
   - Average query latency
   - P95/P99 latencies
   - Cache hit rates
   - Error rates

3. **Cost**
   - Average cost per query
   - Token usage
   - API call counts
   - Cost per user

4. **Quality**
   - Intent matching accuracy
   - Command planning success
   - Output quality ratings
   - Error recovery rate

## References

- OpenAI Agent Framework: https://github.com/openai/swarm
- Model Context Protocol: https://modelcontextprotocol.io
- Opik Documentation: https://opik.ai/docs
- Weave MCP: /Users/maximilien/github/maximilien/weave-mcp
- Claude CLI: https://github.com/anthropics/claude-code

## Appendix: Agent Prompt Templates

### QueryAgent Prompt

```
You are a query validation agent for weave-cli, a vector database management tool.

Your task is to analyze user queries and determine if they are related to weave-cli operations.

Weave-cli supports:
- Collection management (list, create, delete, show)
- Document management (create, update, delete, list, query)
- PDF processing (text extraction, image extraction)
- Batch operations (process directories)
- Semantic search

User query: "{query}"

Respond with JSON:
{
  "is_weave_query": bool,
  "fixed_query": string,
  "intent": string,
  "confidence": float,
  "reason": string
}
```

### PlanningAgent Prompt

```
You are a planning agent for weave-cli operations.

Given a validated query, create a detailed execution plan.

Available tools:
- Weave-MCP tools: {list_mcp_tools}
- Bash commands: {list_safe_bash_commands}

Query: "{fixed_query}"
Intent: "{intent}"

Create an execution plan with steps including:
- Type (bash/weave)
- Command/tool name
- Arguments
- Dependencies
- Risk assessment

Respond with JSON ExecutionPlan.
```

## Conclusion

This design provides a comprehensive framework for implementing AI Agents support in weave-cli. The multi-agent architecture ensures separation of concerns, testability, and extensibility. The implementation prioritizes user experience, safety, and reliability while providing powerful natural language query capabilities.
