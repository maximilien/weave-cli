# New Features Plan - Option B

**Status**: Planned (Future)
**Priority**: Medium (User-facing value)
**Total Effort**: 15-20 hours
**Prerequisites**: None (can start anytime)

---

## Overview

Add powerful user-facing features to enhance weave-cli capabilities:
AI REPL improvements, advanced search, and robust pipeline processing.

---

## Feature 1: AI REPL Enhancements (3-4 hours)

### Current State
- Basic REPL with natural language queries
- Simple command execution
- Limited error handling

### Improvements

#### 1.1 Multi-Turn Conversations (1.5h)

**Goal**: Support context-aware conversations

```go
type REPLSession struct {
    history     []Message
    context     map[string]interface{}
    maxHistory  int
}

func (s *REPLSession) AddMessage(role, content string) {
    s.history = append(s.history, Message{
        Role:    role,
        Content: content,
        Time:    time.Now(),
    })

    // Keep only last N messages
    if len(s.history) > s.maxHistory {
        s.history = s.history[len(s.history)-s.maxHistory:]
    }
}

func (s *REPLSession) GetConversationContext() string {
    var sb strings.Builder
    for _, msg := range s.history {
        sb.WriteString(fmt.Sprintf("[%s] %s\n", msg.Role, msg.Content))
    }
    return sb.String()
}
```

**Example**:
```
> create a collection called TestDocs
✅ Created collection TestDocs

> add all files from ./docs to it
✅ Added 45 documents to TestDocs

> search it for "API documentation"
📊 Found 12 results in TestDocs
...
```

#### 1.2 Command History & Editing (1h)

**Goal**: Shell-like history navigation

```go
import "github.com/chzyer/readline"

func NewREPL() (*REPL, error) {
    rl, err := readline.NewEx(&readline.Config{
        Prompt:          "weave> ",
        HistoryFile:     "~/.weave-cli/history",
        InterruptPrompt: "^C",
        EOFPrompt:       "exit",
    })
    if err != nil {
        return nil, err
    }

    return &REPL{rl: rl}, nil
}
```

**Features**:
- ↑/↓ for history navigation
- Ctrl+R for reverse search
- History persistence across sessions
- Tab completion for commands

#### 1.3 Enhanced Error Recovery (0.5h)

**Goal**: Graceful error handling with suggestions

```go
func (r *REPL) handleError(err error) {
    log.Error().Err(err).Msg("Command failed")

    // Provide helpful suggestions
    if strings.Contains(err.Error(), "collection not found") {
        fmt.Println("\n💡 Tip: List collections with: show collections")
    } else if strings.Contains(err.Error(), "authentication") {
        fmt.Println("\n💡 Tip: Check your API keys with: weave config show")
    }

    fmt.Println("\n❌ Error:", err.Error())
}
```

---

## Feature 2: Advanced Search (4-5 hours)

### 2.1 Hybrid Search (2h)

**Goal**: Combine vector + BM25 search

```go
type HybridSearchOptions struct {
    VectorWeight float64 // 0.0-1.0
    BM25Weight   float64 // 0.0-1.0
    Alpha        float64 // Fusion parameter
}

func (a *Adapter) SearchHybrid(
    ctx context.Context,
    collection string,
    query string,
    opts *HybridSearchOptions,
) ([]*QueryResult, error) {
    // Get vector results
    vectorResults, err := a.SearchSemantic(ctx, collection, query, nil)
    if err != nil {
        return nil, err
    }

    // Get BM25 results
    bm25Results, err := a.SearchBM25(ctx, collection, query, nil)
    if err != nil {
        return nil, err
    }

    // Fuse results with weights
    return fuseResults(vectorResults, bm25Results, opts), nil
}
```

**CLI Usage**:
```bash
weave search MyDocs "API documentation" --hybrid --alpha 0.5
```

### 2.2 Multi-Query Search (1.5h)

**Goal**: Run multiple queries and combine results

```go
type MultiQueryOptions struct {
    Queries       []string
    FusionMethod  string // "rrf", "weighted", "union"
    K             int    // For RRF
}

func SearchMultiQuery(
    ctx context.Context,
    client VectorDBClient,
    collection string,
    opts *MultiQueryOptions,
) ([]*QueryResult, error) {
    var allResults [][]*QueryResult

    // Execute all queries in parallel
    var wg sync.WaitGroup
    resultsChan := make(chan []*QueryResult, len(opts.Queries))

    for _, query := range opts.Queries {
        wg.Add(1)
        go func(q string) {
            defer wg.Done()
            results, _ := client.SearchSemantic(ctx, collection, q, nil)
            resultsChan <- results
        }(query)
    }

    go func() {
        wg.Wait()
        close(resultsChan)
    }()

    for results := range resultsChan {
        allResults = append(allResults, results)
    }

    // Fuse results using selected method
    return fuseByMethod(allResults, opts), nil
}
```

**CLI Usage**:
```bash
weave search MyDocs \
  --queries "API docs" "REST endpoints" "authentication" \
  --fusion rrf
```

### 2.3 Reranking (1-1.5h)

**Goal**: Rerank results using cross-encoder

```go
import "github.com/cohere-ai/cohere-go"

type Reranker struct {
    client *cohere.Client
}

func (r *Reranker) Rerank(
    query string,
    results []*QueryResult,
    topK int,
) ([]*QueryResult, error) {
    // Prepare documents
    docs := make([]string, len(results))
    for i, r := range results {
        docs[i] = r.Text
    }

    // Call reranking API
    resp, err := r.client.Rerank(cohere.RerankRequest{
        Query:     query,
        Documents: docs,
        TopN:      topK,
        Model:     "rerank-english-v2.0",
    })
    if err != nil {
        return nil, err
    }

    // Reorder results
    reranked := make([]*QueryResult, len(resp.Results))
    for i, r := range resp.Results {
        reranked[i] = results[r.Index]
        reranked[i].Score = r.RelevanceScore
    }

    return reranked, nil
}
```

**CLI Usage**:
```bash
weave search MyDocs "API documentation" --rerank --rerank-top 50
```

---

## Feature 3: Pipeline Improvements (3-4 hours)

### 3.1 Parallel Processing (1.5h)

**Goal**: Process multiple files concurrently

```go
type PipelineConfig struct {
    Workers      int
    BatchSize    int
    RetryPolicy  *RetryPolicy
}

func ProcessDirectoryParallel(
    ctx context.Context,
    dir string,
    collection string,
    cfg *PipelineConfig,
) (*PipelineReport, error) {
    files, err := findFiles(dir)
    if err != nil {
        return nil, err
    }

    // Create worker pool
    jobs := make(chan string, len(files))
    results := make(chan *ProcessResult, len(files))

    var wg sync.WaitGroup
    for i := 0; i < cfg.Workers; i++ {
        wg.Add(1)
        go worker(ctx, jobs, results, collection, &wg)
    }

    // Send jobs
    for _, file := range files {
        jobs <- file
    }
    close(jobs)

    // Wait for completion
    go func() {
        wg.Wait()
        close(results)
    }()

    // Collect results
    return collectResults(results), nil
}
```

**CLI Usage**:
```bash
weave pipeline ingest ./docs \
  --collection MyDocs \
  --workers 10 \
  --batch-size 20
```

### 3.2 Resume Capability (1h)

**Goal**: Resume interrupted ingestion

```go
type PipelineState struct {
    JobID         string
    CollectionName string
    ProcessedFiles []string
    FailedFiles    []string
    StartTime      time.Time
    LastUpdate     time.Time
}

func (p *Pipeline) SaveState() error {
    data, _ := json.Marshal(p.state)
    return os.WriteFile(p.stateFile(), data, 0644)
}

func (p *Pipeline) Resume(jobID string) error {
    // Load previous state
    state, err := p.loadState(jobID)
    if err != nil {
        return err
    }

    // Skip already processed files
    remaining := difference(p.allFiles, state.ProcessedFiles)

    // Process remaining files
    return p.processFiles(remaining)
}
```

**CLI Usage**:
```bash
# Start ingestion
weave pipeline ingest ./docs --collection MyDocs --save-state

# If interrupted, resume
weave pipeline resume <job-id>
```

### 3.3 Better Error Recovery (0.5-1h)

**Goal**: Continue on errors, report at end

```go
type PipelineReport struct {
    TotalFiles      int
    ProcessedFiles  int
    FailedFiles     int
    SkippedFiles    int
    Errors          []FileError
    Duration        time.Duration
}

type FileError struct {
    File  string
    Error error
    Retry bool
}

func (p *Pipeline) Process() *PipelineReport {
    report := &PipelineReport{TotalFiles: len(p.files)}

    for _, file := range p.files {
        if err := p.processFile(file); err != nil {
            report.FailedFiles++
            report.Errors = append(report.Errors, FileError{
                File:  file,
                Error: err,
                Retry: isRetryable(err),
            })
            continue
        }
        report.ProcessedFiles++
    }

    return report
}
```

---

## Implementation Priority

**Week 1**: AI REPL Enhancements
- Multi-turn conversations
- Command history
- Error recovery

**Week 2**: Advanced Search
- Hybrid search
- Multi-query search
- Reranking

**Week 3**: Pipeline Improvements
- Parallel processing
- Resume capability
- Error recovery

---

## Success Metrics

- **REPL**: Users can have natural conversations with context
- **Search**: 20%+ improvement in relevance for hybrid search
- **Pipeline**: 3-5x faster ingestion with parallel processing
- **Reliability**: 95%+ success rate even with partial failures

---

## Dependencies

```bash
go get github.com/chzyer/readline
go get github.com/cohere-ai/cohere-go
```

---

## References

- [Option 1 Detailed Plan](OPTION_1_NEW_FEATURES.md)
- [Hybrid Search Research](https://www.pinecone.io/learn/hybrid-search/)
- [Reciprocal Rank Fusion](https://plg.uwaterloo.ca/~gvcormac/cormacksigir09-rrf.pdf)
