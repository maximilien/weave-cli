# Option 1: New Features - Detailed Implementation Plan

**Status**: Planning
**Priority**: ⭐⭐⭐ Highest
**Total Effort**: 19-26 hours
**Target**: Weeks 1-2

---

## Overview

Add high-value user-facing features that leverage the solid v0.8.2 foundation. Focus on automation, interactivity, and AI integration to provide immediate value.

**Core Features:**
1. Pipeline Commands - Batch document ingestion
2. CI/CD Integration - GitHub Actions, Argo, Airflow
3. Interactive REPL Mode - Explore data interactively
4. MCP Server - Claude Desktop integration
5. Progress Bars - Visual feedback
6. JSON/YAML Output - Machine-readable formats
7. Collection Statistics - Analytics

---

## Feature 1.1: Pipeline Commands

### Goal
Enable batch document ingestion from files and directories with progress tracking, parallel processing, and error resilience.

### User Stories
- As a developer, I want to ingest entire documentation directories into my vector DB
- As a data engineer, I want to process PDFs, markdown, and text files in bulk
- As a DevOps engineer, I want to track progress and handle errors gracefully

### CLI Interface

```bash
# Basic usage
weave pipeline ingest ./docs --collection documentation --vdb qdrant-cloud

# Advanced usage with filters
weave pipeline ingest ./content \
  --collection knowledge-base \
  --vdb weaviate-cloud \
  --glob "**/*.{md,pdf,txt}" \
  --exclude "drafts/**" \
  --batch-size 100 \
  --workers 4 \
  --dry-run

# With metadata
weave pipeline ingest ./docs \
  --collection docs \
  --vdb qdrant-cloud \
  --metadata "project=weave-cli" \
  --metadata "version=v0.8.2" \
  --metadata "ingested_at=$(date -u +%Y-%m-%dT%H:%M:%SZ)"

# Resume from failure
weave pipeline ingest ./docs \
  --collection docs \
  --vdb qdrant-cloud \
  --resume-from ingest-state.json

# Output options
weave pipeline ingest ./docs \
  --collection docs \
  --vdb qdrant-cloud \
  --output json \
  --report ingest-report.json
```

### Technical Architecture

**Package Structure:**
```
src/pkg/pipeline/
├── ingest.go           # Main ingestion logic
├── scanner.go          # File discovery and filtering
├── processor.go        # Document processing pipeline
├── batch.go            # Batch grouping and optimization
├── worker.go           # Worker pool for parallel processing
├── progress.go         # Progress tracking and reporting
├── state.go            # State management for resume
└── report.go           # Report generation

src/cmd/pipeline/
├── pipeline.go         # Pipeline command group
├── ingest.go           # Ingest subcommand
└── flags.go            # Command flags and validation
```

**Core Types:**

```go
// IngestOptions configures the ingestion pipeline
type IngestOptions struct {
    // Source configuration
    Source      string   // File or directory path
    Glob        string   // Glob pattern (e.g., "**/*.pdf")
    Exclude     []string // Exclusion patterns
    Recursive   bool     // Recursive directory scan

    // Target configuration
    Collection  string   // Target collection name
    VDBType     string   // Vector database type

    // Processing configuration
    BatchSize   int      // Documents per batch (default: 100)
    Workers     int      // Concurrent workers (default: 4)

    // Document configuration
    Metadata    map[string]string // Additional metadata for all docs
    IncludeImages bool    // Extract images from PDFs

    // Operation modes
    DryRun      bool     // Preview without executing
    Resume      bool     // Resume from previous state
    StateFile   string   // State file path for resume

    // Output configuration
    OutputFormat string  // json, yaml, table
    ReportFile   string  // Report output file
    Quiet        bool    // Suppress progress output
}

// FileInfo represents a discovered file
type FileInfo struct {
    Path         string
    Size         int64
    ModTime      time.Time
    Type         FileType // PDF, TXT, MD, JSON, YAML, IMAGE
    Hash         string   // SHA256 for deduplication
}

// ProcessingResult represents the result of processing a file
type ProcessingResult struct {
    File            FileInfo
    Documents       []*vectordb.Document
    DocumentCount   int
    Error           error
    Duration        time.Duration
}

// IngestReport is the final summary
type IngestReport struct {
    StartTime        time.Time
    EndTime          time.Time
    Duration         time.Duration

    // File statistics
    FilesScanned     int
    FilesProcessed   int
    FilesSkipped     int
    FilesFailed      int

    // Document statistics
    DocumentsCreated int
    DocumentsFailed  int

    // Performance metrics
    ThroughputFiles  float64 // files per second
    ThroughputDocs   float64 // documents per second

    // Error details
    Errors           []ErrorDetail

    // Configuration
    Collection       string
    VDBType          string
    BatchSize        int
    Workers          int
}
```

**Implementation Flow:**

```go
// High-level pipeline flow
func RunIngestPipeline(ctx context.Context, opts *IngestOptions) (*IngestReport, error) {
    // 1. Initialize
    reporter := NewProgressReporter(opts)
    state := LoadOrCreateState(opts.StateFile)

    // 2. Discover files
    scanner := NewFileScanner(opts.Source, opts.Glob, opts.Exclude)
    files, err := scanner.Scan(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to scan files: %w", err)
    }
    reporter.SetTotal(len(files))

    // 3. Filter already processed (if resuming)
    if opts.Resume {
        files = state.FilterProcessed(files)
    }

    // 4. Dry run check
    if opts.DryRun {
        return PreviewIngestion(files, opts), nil
    }

    // 5. Create VDB client
    config := LoadVDBConfig(opts.VDBType)
    client, err := vectordb.CreateClient(config)
    if err != nil {
        return nil, fmt.Errorf("failed to create VDB client: %w", err)
    }
    defer client.Close()

    // 6. Ensure collection exists
    exists, err := client.CollectionExists(ctx, opts.Collection)
    if err != nil {
        return nil, fmt.Errorf("failed to check collection: %w", err)
    }
    if !exists {
        // Create with default schema
        schema := client.GetDefaultSchema(vectordb.SchemaTypeText, opts.Collection)
        if err := client.CreateCollection(ctx, opts.Collection, schema); err != nil {
            return nil, fmt.Errorf("failed to create collection: %w", err)
        }
    }

    // 7. Process files with worker pool
    processor := NewProcessor(client, opts)
    results := processor.ProcessFiles(ctx, files, reporter)

    // 8. Generate report
    report := GenerateReport(results, opts)

    // 9. Save report if requested
    if opts.ReportFile != "" {
        SaveReport(report, opts.ReportFile, opts.OutputFormat)
    }

    return report, nil
}

// FileScanner handles file discovery
type FileScanner struct {
    root    string
    pattern string
    exclude []string
}

func (s *FileScanner) Scan(ctx context.Context) ([]FileInfo, error) {
    var files []FileInfo

    // Use filepath.WalkDir for efficient traversal
    err := filepath.WalkDir(s.root, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }

        // Check context cancellation
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }

        // Skip directories
        if d.IsDir() {
            return nil
        }

        // Check exclusion patterns
        for _, pattern := range s.exclude {
            if matched, _ := filepath.Match(pattern, path); matched {
                return nil
            }
        }

        // Check glob pattern
        if s.pattern != "" {
            matched, err := filepath.Match(s.pattern, filepath.Base(path))
            if err != nil || !matched {
                return nil
            }
        }

        // Get file info
        info, err := d.Info()
        if err != nil {
            return nil
        }

        files = append(files, FileInfo{
            Path:    path,
            Size:    info.Size(),
            ModTime: info.ModTime(),
            Type:    DetectFileType(path),
        })

        return nil
    })

    return files, err
}

// Processor handles document processing with worker pool
type Processor struct {
    client  vectordb.VectorDBClient
    opts    *IngestOptions
    llm     *llm.OpenAIClient
}

func (p *Processor) ProcessFiles(ctx context.Context, files []FileInfo, reporter *ProgressReporter) []ProcessingResult {
    // Create worker pool
    jobs := make(chan FileInfo, len(files))
    results := make(chan ProcessingResult, len(files))

    // Start workers
    var wg sync.WaitGroup
    for i := 0; i < p.opts.Workers; i++ {
        wg.Add(1)
        go p.worker(ctx, jobs, results, &wg, reporter)
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
    var allResults []ProcessingResult
    for result := range results {
        allResults = append(allResults, result)
    }

    return allResults
}

func (p *Processor) worker(ctx context.Context, jobs <-chan FileInfo, results chan<- ProcessingResult, wg *sync.WaitGroup, reporter *ProgressReporter) {
    defer wg.Done()

    for file := range jobs {
        result := p.processFile(ctx, file)
        results <- result
        reporter.Increment()
    }
}

func (p *Processor) processFile(ctx context.Context, file FileInfo) ProcessingResult {
    start := time.Now()

    // Read and process file based on type
    var docs []*vectordb.Document
    var err error

    switch file.Type {
    case FileTypePDF:
        docs, err = p.processPDF(ctx, file)
    case FileTypeTXT, FileTypeMD:
        docs, err = p.processText(ctx, file)
    case FileTypeJSON:
        docs, err = p.processJSON(ctx, file)
    case FileTypeImage:
        docs, err = p.processImage(ctx, file)
    default:
        err = fmt.Errorf("unsupported file type: %s", file.Type)
    }

    if err != nil {
        return ProcessingResult{
            File:  file,
            Error: err,
            Duration: time.Since(start),
        }
    }

    // Add metadata to all documents
    for _, doc := range docs {
        if doc.Metadata == nil {
            doc.Metadata = make(map[string]interface{})
        }
        doc.Metadata["source_file"] = file.Path
        doc.Metadata["source_size"] = file.Size
        doc.Metadata["source_modified"] = file.ModTime.Format(time.RFC3339)

        // Add user-provided metadata
        for k, v := range p.opts.Metadata {
            doc.Metadata[k] = v
        }
    }

    // Batch insert documents
    if len(docs) > 0 {
        batchCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
        defer cancel()

        err = p.client.CreateDocuments(batchCtx, p.opts.Collection, docs)
        if err != nil {
            return ProcessingResult{
                File:          file,
                Documents:     docs,
                DocumentCount: len(docs),
                Error:         fmt.Errorf("failed to create documents: %w", err),
                Duration:      time.Since(start),
            }
        }
    }

    return ProcessingResult{
        File:          file,
        Documents:     docs,
        DocumentCount: len(docs),
        Duration:      time.Since(start),
    }
}

func (p *Processor) processPDF(ctx context.Context, file FileInfo) ([]*vectordb.Document, error) {
    // Use existing PDF processor
    extractor := pdf.NewPDFTextExtractor()
    text, err := extractor.ExtractText(file.Path)
    if err != nil {
        return nil, err
    }

    // Generate embedding
    embedding, err := p.llm.GenerateEmbedding(ctx, text, "")
    if err != nil {
        return nil, err
    }

    doc := &vectordb.Document{
        ID:       generateDocID(file.Path),
        Text:     text,
        Vector:   embedding,
        Metadata: map[string]interface{}{
            "type": "pdf",
        },
    }

    return []*vectordb.Document{doc}, nil
}
```

**Progress Reporting:**

```go
type ProgressReporter struct {
    total     int
    processed int
    errors    int
    mu        sync.Mutex
    bar       *progressbar.ProgressBar
    startTime time.Time
}

func NewProgressReporter(opts *IngestOptions) *ProgressReporter {
    if opts.Quiet {
        return &ProgressReporter{}
    }

    bar := progressbar.NewOptions(-1,
        progressbar.OptionSetDescription("Ingesting documents"),
        progressbar.OptionShowCount(),
        progressbar.OptionShowIts(),
        progressbar.OptionSetItsString("files"),
        progressbar.OptionSetWriter(os.Stderr),
        progressbar.OptionThrottle(100*time.Millisecond),
    )

    return &ProgressReporter{
        bar:       bar,
        startTime: time.Now(),
    }
}

func (r *ProgressReporter) SetTotal(total int) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.total = total
    if r.bar != nil {
        r.bar.ChangeMax(total)
    }
}

func (r *ProgressReporter) Increment() {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.processed++
    if r.bar != nil {
        r.bar.Add(1)
    }
}

func (r *ProgressReporter) IncrementError() {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.errors++
}
```

### Testing Strategy

**Unit Tests:**
```go
func TestFileScanner_Scan(t *testing.T) {
    // Test cases:
    // - Single file
    // - Directory with multiple files
    // - Glob pattern filtering
    // - Exclusion patterns
    // - Recursive vs non-recursive
}

func TestProcessor_ProcessFile(t *testing.T) {
    // Test different file types
    // - PDF
    // - Text
    // - Markdown
    // - JSON
    // - Images
}

func TestWorkerPool(t *testing.T) {
    // Test concurrent processing
    // - Correct number of workers
    // - All files processed
    // - Results collected correctly
}
```

**Integration Tests:**
```go
func TestPipeline_EndToEnd(t *testing.T) {
    // Create temporary directory with test files
    // Run ingestion
    // Verify documents created in VDB
    // Check report accuracy
}
```

### Success Criteria
- ✅ Can ingest 1000+ documents in < 5 minutes
- ✅ Progress bar updates in real-time
- ✅ Handles errors gracefully (logs and continues)
- ✅ Supports PDF, TXT, MD, JSON, YAML, images
- ✅ Resume capability works after interruption
- ✅ Dry-run mode shows accurate preview
- ✅ Report contains all relevant statistics

### Effort Estimate
- Core implementation: 3 hours
- File type handlers: 1 hour
- Progress reporting: 0.5 hour
- State management (resume): 1 hour
- Testing: 1.5 hours
- **Total: 4-6 hours**

---

## Feature 1.1a: CI/CD Integration

### Goal
Enable weave-cli pipeline ingestion in CI/CD platforms with proper error handling, reporting, and automation support.

### Supported Platforms
1. GitHub Actions
2. Argo Workflows
3. Apache Airflow
4. (Bonus) Jenkins

### Implementation Requirements

**Exit Codes:**
```go
const (
    ExitSuccess         = 0  // All documents ingested successfully
    ExitPartialFailure  = 1  // Some errors, but majority succeeded
    ExitCompleteFailure = 2  // Critical error, no documents ingested
)

func DetermineExitCode(report *IngestReport) int {
    if report.FilesFailed == 0 && report.DocumentsFailed == 0 {
        return ExitSuccess
    }

    totalFiles := report.FilesProcessed + report.FilesFailed
    failureRate := float64(report.FilesFailed) / float64(totalFiles)

    if failureRate > 0.5 {
        return ExitCompleteFailure
    }

    return ExitPartialFailure
}
```

**JSON Report Format:**
```json
{
  "status": "success|partial|failure",
  "start_time": "2025-12-19T13:00:00Z",
  "end_time": "2025-12-19T13:05:23Z",
  "duration_seconds": 323,
  "files": {
    "scanned": 127,
    "processed": 125,
    "skipped": 0,
    "failed": 2
  },
  "documents": {
    "created": 312,
    "failed": 0
  },
  "performance": {
    "throughput_files_per_sec": 0.387,
    "throughput_docs_per_sec": 0.967
  },
  "configuration": {
    "collection": "documentation",
    "vdb_type": "qdrant-cloud",
    "batch_size": 100,
    "workers": 4
  },
  "errors": [
    {
      "file": "docs/broken.pdf",
      "error": "PDF parsing failed: corrupted file",
      "timestamp": "2025-12-19T13:02:15Z"
    }
  ]
}
```

**Incremental Ingestion:**
```bash
# Only process files modified in last 24 hours
weave pipeline ingest ./docs \
  --collection docs \
  --vdb qdrant-cloud \
  --since "24h"

# Skip files already in VDB (based on metadata.source_file)
weave pipeline ingest ./docs \
  --collection docs \
  --vdb qdrant-cloud \
  --skip-existing
```

### Documentation Structure

```
docs/integrations/
├── GITHUB_ACTIONS.md       # Complete GH Actions guide
├── ARGO_WORKFLOWS.md       # Argo setup and examples
├── AIRFLOW.md              # Airflow DAG patterns
└── JENKINS.md              # Jenkins pipeline (bonus)

examples/ci-cd/
├── github-actions/
│   ├── basic-ingestion.yml
│   ├── multi-env.yml
│   ├── scheduled.yml
│   └── pr-preview.yml
├── argo/
│   ├── simple-workflow.yaml
│   ├── parallel-ingestion.yaml
│   └── configmap.yaml
└── airflow/
    ├── simple_dag.py
    ├── advanced_dag.py
    └── sensors_dag.py
```

### Success Criteria
- ✅ GitHub Actions workflow runs successfully
- ✅ Argo Workflows handles secrets properly
- ✅ Airflow DAG validates and reports errors
- ✅ Exit codes correctly reflect success/failure
- ✅ JSON reports are parseable and complete

### Effort Estimate
- Exit code implementation: 0.5 hour
- JSON output enhancement: 1 hour
- Documentation (3 platforms): 1.5 hours
- Example workflows: 1 hour
- **Total: 3-4 hours**

---

## Feature 1.2: Interactive REPL Mode

### Goal
Provide an interactive shell for exploring collections and documents without writing code.

### User Stories
- As a developer, I want to explore my vector DB interactively
- As a data scientist, I want to test queries without writing scripts
- As a DevOps engineer, I want to inspect collections and documents

### CLI Interface

```bash
weave repl
weave repl --vdb qdrant-cloud
weave repl --vdb qdrant-cloud --collection documents
```

### REPL Commands

**Connection Management:**
```
connect <vdb-type>          # Connect to vector database
disconnect                  # Disconnect current session
status                      # Show connection status
```

**Collection Operations:**
```
list collections            # List all collections (alias: ls)
use <collection>            # Set active collection (alias: select)
create collection <name>    # Create new collection
delete collection <name>    # Delete collection
info                        # Show current collection info
```

**Document Operations:**
```
search <query> [--top N]    # Semantic search
bm25 <query> [--top N]      # BM25 search (if supported)
hybrid <query> [--top N]    # Hybrid search
get <doc-id>                # Get document by ID
create <text>               # Create new document
delete <doc-id>             # Delete document
list [--limit N]            # List documents
```

**Utilities:**
```
stats                       # Show collection statistics
export <file>               # Export collection to JSON
import <file>               # Import documents from JSON
clear                       # Clear screen
history                     # Show command history
help [command]              # Show help
exit                        # Exit REPL (alias: quit)
```

### Technical Architecture

**Package Structure:**
```
src/cmd/repl/
├── repl.go          # Main REPL loop
├── commands.go      # Command parsing and dispatch
├── context.go       # Session state management
├── executor.go      # Command execution
├── formatter.go     # Output formatting
└── completer.go     # Tab completion
```

**Core Implementation:**

```go
type REPLSession struct {
    client     vectordb.VectorDBClient
    collection string
    vdbType    string
    history    []string
    rl         *readline.Instance
}

func NewREPLSession(vdbType string) (*REPLSession, error) {
    // Setup readline with history
    rl, err := readline.NewEx(&readline.Config{
        Prompt:          "weave> ",
        HistoryFile:     filepath.Join(os.Getenv("HOME"), ".weave", "history"),
        AutoComplete:    NewCompleter(),
        InterruptPrompt: "^C",
        EOFPrompt:       "exit",
    })
    if err != nil {
        return nil, err
    }

    return &REPLSession{
        vdbType: vdbType,
        rl:      rl,
    }, nil
}

func (s *REPLSession) Run(ctx context.Context) error {
    defer s.rl.Close()

    fmt.Println("Welcome to Weave CLI Interactive Mode")
    fmt.Println("Type 'help' for available commands")
    fmt.Println()

    // Auto-connect if VDB type provided
    if s.vdbType != "" {
        if err := s.connect(ctx, s.vdbType); err != nil {
            fmt.Printf("Failed to connect: %v\n", err)
        }
    }

    for {
        line, err := s.rl.Readline()
        if err != nil {
            if err == readline.ErrInterrupt {
                continue
            } else if err == io.EOF {
                break
            }
            return err
        }

        line = strings.TrimSpace(line)
        if line == "" {
            continue
        }

        // Parse and execute command
        if err := s.executeCommand(ctx, line); err != nil {
            if err == io.EOF {
                break
            }
            fmt.Printf("Error: %v\n", err)
        }
    }

    fmt.Println("Goodbye!")
    return nil
}

func (s *REPLSession) executeCommand(ctx context.Context, line string) error {
    parts := strings.Fields(line)
    if len(parts) == 0 {
        return nil
    }

    cmd := parts[0]
    args := parts[1:]

    switch cmd {
    case "connect":
        return s.cmdConnect(ctx, args)
    case "disconnect":
        return s.cmdDisconnect()
    case "status":
        return s.cmdStatus()
    case "list", "ls":
        return s.cmdList(ctx, args)
    case "use", "select":
        return s.cmdUse(args)
    case "search":
        return s.cmdSearch(ctx, args)
    case "get":
        return s.cmdGet(ctx, args)
    case "create":
        return s.cmdCreate(ctx, args)
    case "delete":
        return s.cmdDelete(ctx, args)
    case "stats":
        return s.cmdStats(ctx)
    case "export":
        return s.cmdExport(ctx, args)
    case "clear":
        fmt.Print("\033[H\033[2J")
        return nil
    case "history":
        return s.cmdHistory()
    case "help":
        return s.cmdHelp(args)
    case "exit", "quit":
        return io.EOF
    default:
        return fmt.Errorf("unknown command: %s (type 'help' for available commands)", cmd)
    }
}

func (s *REPLSession) cmdConnect(ctx context.Context, args []string) error {
    if len(args) == 0 {
        return fmt.Errorf("usage: connect <vdb-type>")
    }

    vdbType := args[0]

    // Load configuration
    config := LoadVDBConfig(vdbType)

    // Create client
    client, err := vectordb.CreateClient(config)
    if err != nil {
        return fmt.Errorf("failed to connect: %w", err)
    }

    // Test connection
    if err := client.Health(ctx); err != nil {
        client.Close()
        return fmt.Errorf("health check failed: %w", err)
    }

    s.client = client
    s.vdbType = vdbType

    fmt.Printf("✓ Connected to %s\n", vdbType)
    return nil
}

func (s *REPLSession) cmdSearch(ctx context.Context, args []string) error {
    if s.client == nil {
        return fmt.Errorf("not connected (use 'connect <vdb-type>')")
    }
    if s.collection == "" {
        return fmt.Errorf("no collection selected (use 'use <collection>')")
    }

    // Parse arguments
    var query string
    topK := 5

    for i := 0; i < len(args); i++ {
        if args[i] == "--top" && i+1 < len(args) {
            topK, _ = strconv.Atoi(args[i+1])
            i++
        } else {
            query += args[i] + " "
        }
    }
    query = strings.TrimSpace(query)

    if query == "" {
        return fmt.Errorf("usage: search <query> [--top N]")
    }

    // Execute search
    fmt.Printf("Searching...\n")
    results, err := s.client.SearchSemantic(ctx, s.collection, query, &vectordb.QueryOptions{
        TopK: topK,
    })
    if err != nil {
        return err
    }

    // Format results
    fmt.Printf("✓ Found %d results\n\n", len(results))
    for i, result := range results {
        fmt.Printf("%d. [%.3f] %s\n", i+1, result.Score, truncate(result.Text, 80))
        fmt.Printf("   ID: %s\n", result.ID)
        if len(result.Metadata) > 0 {
            fmt.Printf("   Metadata: %v\n", result.Metadata)
        }
        fmt.Println()
    }

    return nil
}
```

**Tab Completion:**

```go
type Completer struct {
    commands []string
}

func NewCompleter() *Completer {
    return &Completer{
        commands: []string{
            "connect", "disconnect", "status",
            "list", "ls", "use", "select",
            "search", "bm25", "hybrid",
            "get", "create", "delete",
            "stats", "export", "import",
            "clear", "history", "help", "exit", "quit",
        },
    }
}

func (c *Completer) Do(line []rune, pos int) ([][]rune, int) {
    lineStr := string(line[:pos])
    words := strings.Fields(lineStr)

    if len(words) == 0 {
        return c.completeCommand(""), 0
    }

    if len(words) == 1 {
        return c.completeCommand(words[0]), 0
    }

    // Command-specific completion
    switch words[0] {
    case "connect":
        return c.completeVDBType(words[len(words)-1]), len(lineStr) - len(words[len(words)-1])
    case "use", "select":
        return c.completeCollection(words[len(words)-1]), len(lineStr) - len(words[len(words)-1])
    }

    return nil, 0
}

func (c *Completer) completeCommand(prefix string) [][]rune {
    var matches [][]rune
    for _, cmd := range c.commands {
        if strings.HasPrefix(cmd, prefix) {
            matches = append(matches, []rune(cmd))
        }
    }
    return matches
}
```

### Success Criteria
- ✅ REPL starts and accepts commands
- ✅ Tab completion works for commands and arguments
- ✅ Command history persists across sessions
- ✅ Colored output for success/error/info
- ✅ All essential commands work (connect, search, get, create)
- ✅ Help system is comprehensive

### Effort Estimate
- REPL framework setup: 1.5 hours
- Command implementation (8-10 commands): 1.5 hours
- Tab completion and formatting: 1 hour
- **Total: 3-4 hours**

---

## Feature 1.3: MCP Server

### Goal
Expose weave-cli functionality as MCP (Model Context Protocol) tools for Claude Desktop and other MCP clients.

### MCP Tools Implementation

**Tool Definitions:**

```go
var mcpTools = []mcp.Tool{
    {
        Name:        "weave_health_check",
        Description: "Check vector database health and connectivity",
        InputSchema: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "vdb_type": {
                    "type": "string",
                    "enum": []string{
                        "qdrant-cloud", "qdrant-local",
                        "weaviate-cloud", "weaviate-local",
                        "milvus-cloud", "milvus-local",
                        // ... all 20 VDB types
                    },
                    "description": "Vector database type",
                },
            },
            "required": []string{"vdb_type"},
        },
    },
    {
        Name:        "weave_list_collections",
        Description: "List all collections in the vector database",
        InputSchema: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "vdb_type": {
                    "type":        "string",
                    "description": "Vector database type",
                },
            },
            "required": []string{"vdb_type"},
        },
    },
    {
        Name:        "weave_create_collection",
        Description: "Create a new collection for storing vector embeddings",
        InputSchema: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "vdb_type": {
                    "type":        "string",
                    "description": "Vector database type",
                },
                "collection_name": {
                    "type":        "string",
                    "description": "Name of the collection to create",
                },
                "schema_type": {
                    "type":        "string",
                    "enum":        []string{"text", "image", "custom"},
                    "description": "Type of schema to use",
                    "default":     "text",
                },
            },
            "required": []string{"vdb_type", "collection_name"},
        },
    },
    {
        Name:        "weave_search_semantic",
        Description: "Perform semantic search using vector similarity",
        InputSchema: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "vdb_type": {
                    "type":        "string",
                    "description": "Vector database type",
                },
                "collection_name": {
                    "type":        "string",
                    "description": "Collection to search in",
                },
                "query": {
                    "type":        "string",
                    "description": "Search query text",
                },
                "top_k": {
                    "type":        "number",
                    "description": "Number of results to return",
                    "default":     5,
                },
            },
            "required": []string{"vdb_type", "collection_name", "query"},
        },
    },
    {
        Name:        "weave_create_document",
        Description: "Add a new document to the vector database",
        InputSchema: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "vdb_type": {
                    "type":        "string",
                    "description": "Vector database type",
                },
                "collection_name": {
                    "type":        "string",
                    "description": "Collection to add document to",
                },
                "text": {
                    "type":        "string",
                    "description": "Document text content",
                },
                "metadata": {
                    "type":        "object",
                    "description": "Optional metadata for the document",
                },
            },
            "required": []string{"vdb_type", "collection_name", "text"},
        },
    },
    {
        Name:        "weave_get_document",
        Description: "Retrieve a specific document by its ID",
        InputSchema: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "vdb_type": {
                    "type":        "string",
                    "description": "Vector database type",
                },
                "collection_name": {
                    "type":        "string",
                    "description": "Collection containing the document",
                },
                "document_id": {
                    "type":        "string",
                    "description": "ID of the document to retrieve",
                },
            },
            "required": []string{"vdb_type", "collection_name", "document_id"},
        },
    },
    {
        Name:        "weave_delete_document",
        Description: "Delete a document from the vector database",
        InputSchema: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "vdb_type": {
                    "type":        "string",
                    "description": "Vector database type",
                },
                "collection_name": {
                    "type":        "string",
                    "description": "Collection containing the document",
                },
                "document_id": {
                    "type":        "string",
                    "description": "ID of the document to delete",
                },
            },
            "required": []string{"vdb_type", "collection_name", "document_id"},
        },
    },
    {
        Name:        "weave_collection_stats",
        Description: "Get statistics about a collection (document count, etc.)",
        InputSchema: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "vdb_type": {
                    "type":        "string",
                    "description": "Vector database type",
                },
                "collection_name": {
                    "type":        "string",
                    "description": "Collection to get stats for",
                },
            },
            "required": []string{"vdb_type", "collection_name"},
        },
    },
}
```

**MCP Server Implementation:**

```go
package mcpserver

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "os"

    "github.com/maximilien/weave-cli/src/pkg/vectordb"
)

type MCPServer struct {
    clients map[string]vectordb.VectorDBClient
}

func NewMCPServer() *MCPServer {
    return &MCPServer{
        clients: make(map[string]vectordb.VectorDBClient),
    }
}

func (s *MCPServer) Serve(ctx context.Context) error {
    // Use stdio transport
    decoder := json.NewDecoder(os.Stdin)
    encoder := json.NewEncoder(os.Stdout)

    for {
        var req MCPRequest
        if err := decoder.Decode(&req); err != nil {
            if err == io.EOF {
                return nil
            }
            return err
        }

        resp := s.handleRequest(ctx, &req)
        if err := encoder.Encode(resp); err != nil {
            return err
        }
    }
}

func (s *MCPServer) handleRequest(ctx context.Context, req *MCPRequest) *MCPResponse {
    switch req.Method {
    case "tools/list":
        return s.listTools()
    case "tools/call":
        return s.callTool(ctx, req.Params)
    default:
        return &MCPResponse{
            Error: &MCPError{
                Code:    -32601,
                Message: "Method not found",
            },
        }
    }
}

func (s *MCPServer) listTools() *MCPResponse {
    return &MCPResponse{
        Result: map[string]interface{}{
            "tools": mcpTools,
        },
    }
}

func (s *MCPServer) callTool(ctx context.Context, params map[string]interface{}) *MCPResponse {
    name, ok := params["name"].(string)
    if !ok {
        return errorResponse("missing tool name")
    }

    args, ok := params["arguments"].(map[string]interface{})
    if !ok {
        return errorResponse("missing tool arguments")
    }

    switch name {
    case "weave_health_check":
        return s.toolHealthCheck(ctx, args)
    case "weave_list_collections":
        return s.toolListCollections(ctx, args)
    case "weave_create_collection":
        return s.toolCreateCollection(ctx, args)
    case "weave_search_semantic":
        return s.toolSearchSemantic(ctx, args)
    case "weave_create_document":
        return s.toolCreateDocument(ctx, args)
    case "weave_get_document":
        return s.toolGetDocument(ctx, args)
    case "weave_delete_document":
        return s.toolDeleteDocument(ctx, args)
    case "weave_collection_stats":
        return s.toolCollectionStats(ctx, args)
    default:
        return errorResponse(fmt.Sprintf("unknown tool: %s", name))
    }
}

func (s *MCPServer) getOrCreateClient(vdbType string) (vectordb.VectorDBClient, error) {
    if client, exists := s.clients[vdbType]; exists {
        return client, nil
    }

    config := LoadVDBConfig(vdbType)
    client, err := vectordb.CreateClient(config)
    if err != nil {
        return nil, err
    }

    s.clients[vdbType] = client
    return client, nil
}

func (s *MCPServer) toolHealthCheck(ctx context.Context, args map[string]interface{}) *MCPResponse {
    vdbType, ok := args["vdb_type"].(string)
    if !ok {
        return errorResponse("vdb_type is required")
    }

    client, err := s.getOrCreateClient(vdbType)
    if err != nil {
        return errorResponse(fmt.Sprintf("failed to connect: %v", err))
    }

    if err := client.Health(ctx); err != nil {
        return &MCPResponse{
            Result: map[string]interface{}{
                "content": []map[string]interface{}{
                    {
                        "type": "text",
                        "text": fmt.Sprintf("❌ Health check failed: %v", err),
                    },
                },
            },
        }
    }

    return &MCPResponse{
        Result: map[string]interface{}{
            "content": []map[string]interface{}{
                {
                    "type": "text",
                    "text": fmt.Sprintf("✅ %s is healthy and accessible", vdbType),
                },
            },
        },
    }
}

func (s *MCPServer) toolSearchSemantic(ctx context.Context, args map[string]interface{}) *MCPResponse {
    vdbType, _ := args["vdb_type"].(string)
    collection, _ := args["collection_name"].(string)
    query, _ := args["query"].(string)
    topK := 5
    if k, ok := args["top_k"].(float64); ok {
        topK = int(k)
    }

    client, err := s.getOrCreateClient(vdbType)
    if err != nil {
        return errorResponse(err.Error())
    }

    results, err := client.SearchSemantic(ctx, collection, query, &vectordb.QueryOptions{
        TopK: topK,
    })
    if err != nil {
        return errorResponse(err.Error())
    }

    // Format results
    text := fmt.Sprintf("Found %d results:\n\n", len(results))
    for i, result := range results {
        text += fmt.Sprintf("%d. [Score: %.3f] %s\n", i+1, result.Score, truncate(result.Text, 200))
        text += fmt.Sprintf("   ID: %s\n", result.ID)
        if len(result.Metadata) > 0 {
            metaJSON, _ := json.MarshalIndent(result.Metadata, "   ", "  ")
            text += fmt.Sprintf("   Metadata: %s\n", string(metaJSON))
        }
        text += "\n"
    }

    return &MCPResponse{
        Result: map[string]interface{}{
            "content": []map[string]interface{}{
                {
                    "type": "text",
                    "text": text,
                },
            },
        },
    }
}
```

### Claude Desktop Configuration

**Setup Instructions:**

1. Build weave-cli binary
2. Add to Claude Desktop config:

```json
{
  "mcpServers": {
    "weave-cli": {
      "command": "/usr/local/bin/weave",
      "args": ["mcp", "serve"],
      "env": {
        "OPENAI_API_KEY": "sk-...",
        "QDRANT_API_KEY": "...",
        "QDRANT_URL": "https://xyz.qdrant.io",
        "WEAVIATE_API_KEY": "...",
        "WEAVIATE_URL": "https://xyz.weaviate.network"
      }
    }
  }
}
```

### Success Criteria
- ✅ MCP server starts and responds to requests
- ✅ All 8 tools are registered and callable
- ✅ Claude Desktop can use weave-cli tools
- ✅ Search results are properly formatted
- ✅ Error handling is robust

### Effort Estimate
- MCP server framework: 2 hours
- Tool implementations (8 tools): 2.5 hours
- Testing with Claude Desktop: 1 hour
- Documentation: 0.5 hour
- **Total: 5-6 hours**

---

## Feature 1.4: Progress Bars

### Implementation

```go
import "github.com/schollz/progressbar/v3"

func createProgressBar(total int, description string) *progressbar.ProgressBar {
    return progressbar.NewOptions(total,
        progressbar.OptionSetDescription(description),
        progressbar.OptionShowCount(),
        progressbar.OptionShowIts(),
        progressbar.OptionSetItsString("docs"),
        progressbar.OptionSetWriter(os.Stderr),
        progressbar.OptionThrottle(100*time.Millisecond),
        progressbar.OptionSetTheme(progressbar.Theme{
            Saucer:        "█",
            SaucerPadding: "░",
            BarStart:      "[",
            BarEnd:        "]",
        }),
    )
}
```

### Effort: 1-2 hours

---

## Feature 1.5: JSON/YAML Output

### Implementation

```go
type OutputFormatter interface {
    Format(data interface{}) ([]byte, error)
}

type JSONFormatter struct{}
type YAMLFormatter struct{}
type TableFormatter struct{}

func GetFormatter(format string) OutputFormatter {
    switch format {
    case "json":
        return &JSONFormatter{}
    case "yaml":
        return &YAMLFormatter{}
    default:
        return &TableFormatter{}
    }
}
```

### Effort: 2-3 hours

---

## Feature 1.6: Collection Statistics

### Implementation

```go
func GenerateCollectionStats(ctx context.Context, client vectordb.VectorDBClient, collection string) (*CollectionStats, error) {
    // Get document count
    count, _ := client.GetCollectionCount(ctx, collection)

    // Sample documents for analysis
    docs, _ := client.ListDocuments(ctx, collection, 1000, 0)

    // Analyze metadata distribution
    metadataStats := analyzeMetadata(docs)

    // Analyze document sizes
    sizeStats := analyzeSizes(docs)

    return &CollectionStats{
        DocumentCount:       count,
        VectorDimensions:    1536,
        MetadataDistribution: metadataStats,
        SizeStatistics:      sizeStats,
    }, nil
}
```

### Effort: 2-3 hours

---

## Total Timeline

**Week 1:**
- Day 1-2: Pipeline Commands (4-6h)
- Day 3: CI/CD Integration (3-4h)

**Week 2:**
- Day 1-2: REPL Mode (3-4h)
- Day 3-4: MCP Server (5-6h)
- Day 5: Progress + Output + Stats (5-8h)

**Total: 19-26 hours across 2 weeks**
