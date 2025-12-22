# Option 1: New Features - Detailed Implementation Plan

**Status**: In Progress (2/7 features complete)
**Priority**: ⭐⭐⭐ Highest
**Total Effort**: 19-26 hours (8-10 hours remaining)
**Target**: Weeks 1-2

---

## Overview

Add high-value user-facing features that leverage the solid v0.8.2 foundation. Focus on automation, interactivity, and AI integration to provide immediate value.

**Core Features:**
1. Pipeline Commands - Batch document ingestion
2. ✅ Interactive REPL Mode - Enhanced with structured commands (COMPLETED 2025-12-22)
3. MCP Client Integration - Call external MCP servers
4. Progress Bars - Visual feedback
5. JSON/YAML Output - Machine-readable formats
6. Collection Statistics - Analytics
7. ✅ AI Schema Suggestion - Generate schemas from samples (COMPLETED 2025-12-22)

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

## Feature 1.2: Interactive REPL Mode ✅ COMPLETED

**Status**: ✅ Completed 2025-12-22
**Effort**: 3-4 hours (actual)
**Commits**: 7024702, 887f22a (tests), 183ceef (formatting)

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


---

# Feature 1.2: MCP Client Integration

## Overview
Add MCP (Model Context Protocol) client capabilities to weave-cli, allowing it to call external MCP servers for document enrichment, metadata extraction, and enhanced query processing.

## Goals
1. Enable weave-cli to act as an MCP client (complement to weave-mcp server)
2. Support MCP stdio and HTTP transport protocols
3. Integrate MCP tools into pipeline and query workflows
4. Provide configuration-driven MCP server connections

## Use Cases

### 1. Document Enrichment During Ingestion
```bash
# Enrich documents with MCP tools before ingestion
weave pipeline ingest ./docs --collection knowledge \
  --mcp-enrich \
  --mcp-server localhost:8030 \
  --mcp-tools extract_entities,summarize
```

**Flow:**
- Scan documents
- For each document, call MCP tools (extract_entities, summarize)
- Add enriched metadata to document
- Generate embeddings
- Insert into VDB

### 2. Query Enhancement
```bash
# Use MCP to enhance queries before search
weave query --collection knowledge \
  --query "What are the key features?" \
  --mcp-enhance \
  --mcp-server localhost:8030 \
  --mcp-tool rewrite_query
```

**Flow:**
- User provides query
- Call MCP tool to enhance/rewrite query
- Execute enhanced query against VDB
- Optionally post-process results with MCP

### 3. Interactive REPL with MCP
```bash
# REPL with MCP tool access
weave repl --collection knowledge \
  --mcp-server localhost:8030
  
> /mcp list  # List available MCP tools
> /mcp call extract_entities --text "..."  # Call MCP tool
> /enrich last 5  # Enrich last 5 results with MCP
```

## Architecture

### Package Structure
```
src/pkg/mcp/
├── client.go           # MCP client interface
├── transport.go        # HTTP and stdio transports
├── tools.go            # MCP tool invocation
├── config.go           # MCP server configuration
└── integration.go      # Integration with pipeline/query

src/cmd/mcp/
├── mcp.go              # MCP command group
├── connect.go          # Connect to MCP server
├── list.go             # List MCP tools
├── call.go             # Call MCP tool
└── test.go             # Test MCP connection
```

### Core Types

```go
// MCPClient provides MCP protocol client functionality
type MCPClient interface {
    // Connection management
    Connect(ctx context.Context, config *MCPConfig) error
    Disconnect() error
    Ping(ctx context.Context) error
    
    // Tool discovery
    ListTools(ctx context.Context) ([]*MCPTool, error)
    GetTool(ctx context.Context, name string) (*MCPTool, error)
    
    // Tool invocation
    CallTool(ctx context.Context, name string, args map[string]interface{}) (*MCPResult, error)
    CallToolBatch(ctx context.Context, calls []*MCPToolCall) ([]*MCPResult, error)
}

// MCPConfig configures MCP server connection
type MCPConfig struct {
    ServerURL   string            // HTTP URL or stdio command
    Transport   MCPTransport      // http or stdio
    Auth        *MCPAuth          // Authentication config
    Timeout     time.Duration     // Request timeout
    MaxRetries  int               // Retry attempts
    Headers     map[string]string // Custom headers (HTTP only)
}

type MCPTransport string
const (
    TransportHTTP  MCPTransport = "http"
    TransportStdio MCPTransport = "stdio"
)

// MCPTool represents an available MCP tool
type MCPTool struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    InputSchema map[string]interface{} `json:"inputSchema"`
}

// MCPToolCall represents a tool invocation
type MCPToolCall struct {
    Tool string                 `json:"tool"`
    Args map[string]interface{} `json:"arguments"`
}

// MCPResult represents tool execution result
type MCPResult struct {
    Content interface{}            `json:"content"`
    Metadata map[string]interface{} `json:"metadata,omitempty"`
    Error   string                 `json:"error,omitempty"`
}
```

## Implementation Details

### 1. MCP Client (HTTP Transport)

```go
type HTTPMCPClient struct {
    baseURL    string
    httpClient *http.Client
    auth       *MCPAuth
}

func (c *HTTPMCPClient) CallTool(ctx context.Context, name string, args map[string]interface{}) (*MCPResult, error) {
    request := map[string]interface{}{
        "method": "tools/call",
        "params": map[string]interface{}{
            "name":      name,
            "arguments": args,
        },
    }
    
    resp, err := c.post(ctx, "/mcp/v1/rpc", request)
    if err != nil {
        return nil, err
    }
    
    var result MCPResult
    if err := json.Unmarshal(resp, &result); err != nil {
        return nil, err
    }
    
    return &result, nil
}
```

### 2. MCP Client (stdio Transport)

```go
type StdioMCPClient struct {
    cmd    *exec.Cmd
    stdin  io.WriteCloser
    stdout io.ReadCloser
    stderr io.ReadCloser
}

func (c *StdioMCPClient) Connect(ctx context.Context, config *MCPConfig) error {
    // Parse command from ServerURL
    cmdParts := strings.Split(config.ServerURL, " ")
    c.cmd = exec.CommandContext(ctx, cmdParts[0], cmdParts[1:]...)
    
    // Setup pipes
    c.stdin, _ = c.cmd.StdinPipe()
    c.stdout, _ = c.cmd.StdoutPipe()
    c.stderr, _ = c.cmd.StderrPipe()
    
    // Start process
    return c.cmd.Start()
}

func (c *StdioMCPClient) CallTool(ctx context.Context, name string, args map[string]interface{}) (*MCPResult, error) {
    // Write JSON-RPC request to stdin
    request := map[string]interface{}{
        "jsonrpc": "2.0",
        "id":      uuid.New().String(),
        "method":  "tools/call",
        "params": map[string]interface{}{
            "name":      name,
            "arguments": args,
        },
    }
    
    if err := json.NewEncoder(c.stdin).Encode(request); err != nil {
        return nil, err
    }
    
    // Read JSON-RPC response from stdout
    var response struct {
        Result *MCPResult `json:"result"`
        Error  *struct {
            Message string `json:"message"`
        } `json:"error"`
    }
    
    if err := json.NewDecoder(c.stdout).Decode(&response); err != nil {
        return nil, err
    }
    
    if response.Error != nil {
        return nil, fmt.Errorf("MCP error: %s", response.Error.Message)
    }
    
    return response.Result, nil
}
```

### 3. Pipeline Integration

```go
type MCPEnricher struct {
    client MCPClient
    tools  []string  // Tool names to apply
}

func (e *MCPEnricher) EnrichDocument(ctx context.Context, doc *vectordb.Document) error {
    for _, toolName := range e.tools {
        result, err := e.client.CallTool(ctx, toolName, map[string]interface{}{
            "text": doc.Text,
        })
        if err != nil {
            return fmt.Errorf("MCP tool %s failed: %w", toolName, err)
        }
        
        // Merge result metadata into document
        if result.Metadata != nil {
            if doc.Metadata == nil {
                doc.Metadata = make(map[string]interface{})
            }
            for k, v := range result.Metadata {
                doc.Metadata[fmt.Sprintf("mcp_%s_%s", toolName, k)] = v
            }
        }
    }
    
    return nil
}
```

## Configuration

### config.yaml
```yaml
mcp:
  servers:
    - name: local-enrichment
      url: http://localhost:8030
      transport: http
      enabled: true
      tools:
        - extract_entities
        - summarize
        - classify
        
    - name: weave-mcp-stdio
      url: /Users/max/weave-mcp/bin/weave-mcp-stdio
      transport: stdio
      enabled: true
      
  # Default MCP server for commands
  default_server: local-enrichment
  
  # Auto-enrichment settings
  auto_enrich:
    enabled: false
    tools:
      - extract_entities
```

## CLI Commands

### mcp connect
```bash
# Test connection to MCP server
weave mcp connect --server localhost:8030
weave mcp connect --server stdio:./bin/weave-mcp-stdio
```

### mcp list
```bash
# List available tools
weave mcp list --server localhost:8030

# Output:
# Available MCP Tools:
#   extract_entities    - Extract named entities from text
#   summarize           - Generate text summary
#   classify            - Classify document into categories
#   rewrite_query       - Enhance search queries
```

### mcp call
```bash
# Call MCP tool directly
weave mcp call extract_entities \
  --arg text="Apple Inc. was founded in Cupertino" \
  --server localhost:8030

# Output:
# {
#   "entities": [
#     {"text": "Apple Inc.", "type": "ORGANIZATION"},
#     {"text": "Cupertino", "type": "LOCATION"}
#   ]
# }
```

### pipeline ingest with MCP
```bash
# Enrich documents during ingestion
weave pipeline ingest ./docs \
  --collection knowledge \
  --mcp-enrich \
  --mcp-server localhost:8030 \
  --mcp-tools extract_entities,summarize \
  --output json
```

## Testing Strategy

### Unit Tests
```go
func TestMCPClient_CallTool_HTTP(t *testing.T) {
    // Mock HTTP MCP server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Respond with mock result
    }))
    defer server.Close()
    
    client := NewHTTPMCPClient(server.URL, nil)
    result, err := client.CallTool(context.Background(), "test_tool", map[string]interface{}{
        "arg1": "value1",
    })
    
    // Assert result
}

func TestMCPClient_CallTool_Stdio(t *testing.T) {
    // Test stdio transport with mock command
}

func TestMCPEnricher_EnrichDocument(t *testing.T) {
    // Test document enrichment flow
}
```

### Integration Tests
```go
func TestMCP_Integration_WithWeaveMCP(t *testing.T) {
    // Start weave-mcp server
    // Connect weave-cli as MCP client
    // Call tools
    // Verify results
}
```

## Integration Points

### 1. Pipeline Package
- Add MCPEnricher to processor pipeline
- Call MCP tools before/after document creation
- Enrich metadata automatically

### 2. Query Package  
- Add MCPQueryEnhancer
- Rewrite queries using MCP before search
- Post-process results with MCP

### 3. REPL Package
- Add `/mcp` commands
- Interactive MCP tool invocation
- Enrich search results on-the-fly

## Success Criteria
- ✅ Connect to weave-mcp server (HTTP and stdio)
- ✅ List available MCP tools
- ✅ Call MCP tools with arguments
- ✅ Integrate MCP into pipeline ingestion
- ✅ Integrate MCP into query enhancement
- ✅ Handle errors gracefully (timeout, connection issues)
- ✅ Configuration-driven MCP server management
- ✅ Tests cover HTTP and stdio transports

## Effort Estimate
- MCP client (HTTP transport): 1.5 hours
- MCP client (stdio transport): 1 hour
- Pipeline integration: 1 hour
- Query integration: 0.5 hour
- CLI commands: 1 hour
- Configuration: 0.5 hour
- Testing: 1.5 hours
- **Total: 5-6 hours**

## Dependencies
- weave-mcp server (companion project)
- JSON-RPC library for stdio transport
- HTTP client for HTTP transport

## Future Enhancements
- MCP tool chaining (pipe output of one tool to another)
- MCP result caching
- Async tool invocation for better performance
- MCP tool marketplace/discovery

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

## Feature 1.7: AI Schema Suggestion ✅ COMPLETED

**Status**: ✅ Completed 2025-12-22
**Effort**: 5-6 hours (actual)
**Commits**: cc07c41, 887f22a (tests), 183ceef (formatting), 6d8c23a (test fix)

### Goal
Enable users to generate optimized collection schemas by analyzing sample documents using an AI agent. The agent analyzes document structure, content patterns, and metadata to suggest an appropriate schema configuration.

### User Stories
- As a developer, I want to quickly create a schema without manually analyzing my documents
- As a data engineer, I want AI to suggest optimal field types and indexes for my use case
- As a new user, I want guidance on what schema works best for my document types

### CLI Interface

```bash
# Analyze files and suggest schema
weave schema suggest ./samples \
  --collection documents \
  --output schema.yaml

# Interactive mode with refinement
weave schema suggest ./samples \
  --collection documents \
  --interactive

# Analyze specific file types
weave schema suggest ./samples \
  --glob "**/*.{pdf,md}" \
  --collection documentation \
  --output docs-schema.yaml

# With custom requirements
weave schema suggest ./samples \
  --collection products \
  --requirements "Include price filtering, support multi-language search" \
  --output products-schema.yaml

# Apply suggested schema immediately
weave schema suggest ./samples \
  --collection docs \
  --vdb weaviate-cloud \
  --apply
```

### Technical Architecture

**Package Structure:**
```
src/pkg/agents/
├── schema_agent.go         # AI schema analysis agent

src/cmd/schema/
├── schema.go              # Schema command group
├── suggest.go             # Suggest subcommand
└── analyze.go             # Schema analysis helpers
```

**Core Types:**

```go
// SchemaAnalysisInput represents input for schema analysis
type SchemaAnalysisInput struct {
    SampleFiles      []string          // Paths to sample documents
    CollectionName   string            // Target collection name
    Requirements     string            // User requirements (optional)
    VDBType          string            // Target VDB type
    MaxSamples       int               // Max files to analyze (default: 50)
}

// SchemaAnalysisOutput represents AI-generated schema suggestion
type SchemaAnalysisOutput struct {
    Schema           SchemaConfig      // Suggested schema
    Reasoning        string            // Why this schema was chosen
    FieldAnalysis    []FieldSuggestion // Per-field recommendations
    Confidence       float64           // Confidence score (0-1)
    Warnings         []string          // Potential issues
    Alternatives     []SchemaConfig    // Alternative schemas
}

// FieldSuggestion represents analysis for a single field
type FieldSuggestion struct {
    Name         string   // Field name
    Type         string   // Data type (text, number, boolean, etc.)
    Indexed      bool     // Should be indexed
    Filterable   bool     // Should support filtering
    Required     bool     // Is required field
    Examples     []string // Example values found
    Frequency    float64  // How often field appears (0-1)
    Cardinality  int      // Unique value count
    Reasoning    string   // Why this configuration
}

// SchemaConfig represents a complete schema configuration
type SchemaConfig struct {
    CollectionName   string              `yaml:"collection_name"`
    VectorDimensions int                 `yaml:"vector_dimensions"`
    SimilarityMetric string              `yaml:"similarity_metric"`
    Fields           []FieldConfig       `yaml:"fields"`
    Indexes          []IndexConfig       `yaml:"indexes"`
    Metadata         map[string]string   `yaml:"metadata,omitempty"`
}

// FieldConfig represents a field configuration
type FieldConfig struct {
    Name        string `yaml:"name"`
    Type        string `yaml:"type"`
    Description string `yaml:"description,omitempty"`
    Indexed     bool   `yaml:"indexed"`
    Filterable  bool   `yaml:"filterable"`
    Required    bool   `yaml:"required"`
}

// IndexConfig represents an index configuration
type IndexConfig struct {
    Name   string   `yaml:"name"`
    Fields []string `yaml:"fields"`
    Type   string   `yaml:"type"` // "vector", "bm25", "composite"
}
```

### Implementation

**Schema Analysis Agent:**

```go
// SchemaAgent analyzes documents and suggests schemas
type SchemaAgent struct {
    llmClient *llm.OpenAIClient
}

func NewSchemaAgent(llmClient *llm.OpenAIClient) *SchemaAgent {
    return &SchemaAgent{llmClient: llmClient}
}

func (a *SchemaAgent) Name() string {
    return "schema-agent"
}

func (a *SchemaAgent) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    analysisInput := input.(*SchemaAnalysisInput)

    // Step 1: Sample and extract document content
    samples := a.extractSamples(analysisInput.SampleFiles, analysisInput.MaxSamples)

    // Step 2: Analyze document structure
    structure := a.analyzeStructure(samples)

    // Step 3: Use LLM to suggest schema
    prompt := a.buildAnalysisPrompt(structure, analysisInput)

    response, err := a.llmClient.CompleteStructured(ctx, prompt, &SchemaAnalysisOutput{})
    if err != nil {
        return nil, fmt.Errorf("failed to analyze schema: %w", err)
    }

    return response.(*SchemaAnalysisOutput), nil
}

func (a *SchemaAgent) buildAnalysisPrompt(structure DocumentStructure, input *SchemaAnalysisInput) string {
    return fmt.Sprintf(`Analyze the following document samples and suggest an optimal schema for a %s vector database collection.

Collection Name: %s
Number of Samples: %d
Requirements: %s

Document Structure Analysis:
- File Types: %v
- Common Fields: %v
- Field Types: %v
- Sample Content Lengths: %v

Please suggest a schema configuration that:
1. Optimizes for search and retrieval
2. Includes appropriate indexes
3. Handles the observed field types efficiently
4. Supports filtering on common metadata
5. Uses appropriate vector dimensions for the content type

Provide:
- Complete schema configuration (fields, indexes, settings)
- Reasoning for each field choice
- Confidence score (0-1)
- Any warnings or considerations
- Alternative schema options if applicable

Return the response in JSON format following the SchemaAnalysisOutput structure.`,
        input.VDBType,
        input.CollectionName,
        len(structure.Samples),
        input.Requirements,
        structure.FileTypes,
        structure.CommonFields,
        structure.FieldTypes,
        structure.ContentLengths,
    )
}
```

**Document Analysis:**

```go
type DocumentStructure struct {
    Samples        []DocumentSample
    FileTypes      []string
    CommonFields   map[string]int  // field -> frequency
    FieldTypes     map[string]string // field -> inferred type
    ContentLengths []int
    Languages      []string
}

type DocumentSample struct {
    Path     string
    Type     string
    Size     int64
    Fields   map[string]interface{}
    Preview  string // First 1000 chars
}

func (a *SchemaAgent) extractSamples(files []string, maxSamples int) []DocumentSample {
    var samples []DocumentSample
    count := 0

    for _, file := range files {
        if count >= maxSamples {
            break
        }

        sample := a.extractDocumentSample(file)
        if sample != nil {
            samples = append(samples, *sample)
            count++
        }
    }

    return samples
}

func (a *SchemaAgent) extractDocumentSample(filePath string) *DocumentSample {
    fileType := detectFileType(filePath)

    switch fileType {
    case FileTypePDF:
        return a.extractPDFSample(filePath)
    case FileTypeJSON:
        return a.extractJSONSample(filePath)
    case FileTypeMD, FileTypeTXT:
        return a.extractTextSample(filePath)
    default:
        return nil
    }
}

func (a *SchemaAgent) analyzeStructure(samples []DocumentSample) DocumentStructure {
    structure := DocumentStructure{
        Samples:      samples,
        CommonFields: make(map[string]int),
        FieldTypes:   make(map[string]string),
    }

    // Analyze file types
    typeMap := make(map[string]bool)
    for _, sample := range samples {
        typeMap[sample.Type] = true
        structure.ContentLengths = append(structure.ContentLengths, int(sample.Size))

        // Count field occurrences
        for fieldName, fieldValue := range sample.Fields {
            structure.CommonFields[fieldName]++

            // Infer field type
            if _, exists := structure.FieldTypes[fieldName]; !exists {
                structure.FieldTypes[fieldName] = inferType(fieldValue)
            }
        }
    }

    for fileType := range typeMap {
        structure.FileTypes = append(structure.FileTypes, fileType)
    }

    return structure
}

func inferType(value interface{}) string {
    switch v := value.(type) {
    case string:
        // Check if it's a date, URL, etc.
        if isDate(v) {
            return "datetime"
        }
        if isURL(v) {
            return "url"
        }
        return "text"
    case int, int64, float64:
        return "number"
    case bool:
        return "boolean"
    case []interface{}:
        return "array"
    case map[string]interface{}:
        return "object"
    default:
        return "unknown"
    }
}
```

**CLI Command:**

```go
var suggestCmd = &cobra.Command{
    Use:   "suggest SOURCE",
    Short: "Suggest collection schema by analyzing sample documents",
    Long: `Analyze sample documents and suggest an optimal collection schema using AI.

The AI agent examines document structure, field types, content patterns, and
metadata to recommend a schema configuration tailored to your data.

Examples:
  # Analyze samples and output suggested schema
  weave schema suggest ./samples --collection docs --output schema.yaml

  # Interactive mode with AI explanations
  weave schema suggest ./samples --collection products --interactive

  # Include custom requirements
  weave schema suggest ./samples --collection articles \
    --requirements "Support multi-language search, enable date filtering" \
    --output articles-schema.yaml

  # Apply schema immediately
  weave schema suggest ./samples --collection docs --vdb weaviate-cloud --apply`,
    Args: cobra.ExactArgs(1),
    RunE: runSchemaSuggest,
}

func runSchemaSuggest(cmd *cobra.Command, args []string) error {
    source := args[0]
    ctx := context.Background()

    // Create LLM client
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        return fmt.Errorf("OPENAI_API_KEY required for AI schema suggestion")
    }

    llmClient, err := llm.NewOpenAIClient(apiKey)
    if err != nil {
        return fmt.Errorf("failed to create LLM client: %w", err)
    }

    // Scan for sample files
    scanner := pipeline.NewFileScanner(source, glob, exclude, true)
    files, err := scanner.Scan(ctx)
    if err != nil {
        return fmt.Errorf("failed to scan files: %w", err)
    }

    fmt.Printf("🔍 Analyzing %d sample documents...\n", len(files))

    // Extract file paths
    var filePaths []string
    for _, file := range files {
        filePaths = append(filePaths, file.Path)
    }

    // Create schema agent
    agent := agents.NewSchemaAgent(llmClient)

    // Analyze and suggest schema
    input := &agents.SchemaAnalysisInput{
        SampleFiles:    filePaths,
        CollectionName: collection,
        Requirements:   requirements,
        VDBType:        vdbType,
        MaxSamples:     maxSamples,
    }

    result, err := agent.Execute(ctx, input)
    if err != nil {
        return fmt.Errorf("schema analysis failed: %w", err)
    }

    output := result.(*agents.SchemaAnalysisOutput)

    // Display results
    displaySchemaAnalysis(output, interactive)

    // Save to file if requested
    if outputFile != "" {
        if err := saveSchema(output.Schema, outputFile); err != nil {
            return fmt.Errorf("failed to save schema: %w", err)
        }
        fmt.Printf("✅ Schema saved to %s\n", outputFile)
    }

    // Apply schema if requested
    if apply {
        if err := applySchema(ctx, output.Schema, vdbType); err != nil {
            return fmt.Errorf("failed to apply schema: %w", err)
        }
        fmt.Printf("✅ Schema applied to %s collection\n", collection)
    }

    return nil
}

func displaySchemaAnalysis(output *agents.SchemaAnalysisOutput, interactive bool) {
    fmt.Printf("\n📋 Suggested Schema for '%s'\n", output.Schema.CollectionName)
    fmt.Printf("   Confidence: %.1f%%\n\n", output.Confidence*100)

    fmt.Println("🔧 Fields:")
    for _, field := range output.Schema.Fields {
        fmt.Printf("   • %s (%s)", field.Name, field.Type)
        var attrs []string
        if field.Indexed {
            attrs = append(attrs, "indexed")
        }
        if field.Filterable {
            attrs = append(attrs, "filterable")
        }
        if field.Required {
            attrs = append(attrs, "required")
        }
        if len(attrs) > 0 {
            fmt.Printf(" [%s]", strings.Join(attrs, ", "))
        }
        fmt.Println()
    }

    if len(output.Schema.Indexes) > 0 {
        fmt.Println("\n📊 Indexes:")
        for _, index := range output.Schema.Indexes {
            fmt.Printf("   • %s: %s (%v)\n", index.Name, index.Type, index.Fields)
        }
    }

    fmt.Printf("\n💡 Reasoning:\n%s\n", output.Reasoning)

    if len(output.Warnings) > 0 {
        fmt.Println("\n⚠️  Warnings:")
        for _, warning := range output.Warnings {
            fmt.Printf("   • %s\n", warning)
        }
    }

    if interactive && len(output.Alternatives) > 0 {
        fmt.Println("\n🔀 Alternative Schemas:")
        for i, alt := range output.Alternatives {
            fmt.Printf("   %d. %s (fields: %d, indexes: %d)\n",
                i+1, alt.CollectionName, len(alt.Fields), len(alt.Indexes))
        }
    }
}
```

### Integration with Existing Features

**1. Pipeline Integration:**
```bash
# Suggest schema, then ingest
weave schema suggest ./docs --collection documentation --output schema.yaml
weave cols create documentation --schema schema.yaml --vdb qdrant-cloud
weave pipeline ingest ./docs --collection documentation --vdb qdrant-cloud
```

**2. MCP Server Integration:**
```json
{
  "tool": "suggest_schema",
  "arguments": {
    "samples": "./docs",
    "collection": "knowledge-base",
    "requirements": "Support semantic search on technical documentation"
  }
}
```

**3. REPL Integration:**
```
weave> schema suggest ./samples --collection products
🔍 Analyzing 25 sample documents...
📋 Suggested schema ready
weave> schema show
[displays schema]
weave> schema apply
✅ Schema applied to products collection
```

### Success Metrics

- Schema suggestion accuracy: >85% user acceptance rate
- Analysis time: <30 seconds for 50 samples
- Field detection: >90% of common fields identified
- User satisfaction: Reduces schema creation time by >70%

### Effort: 3-4 hours

**Breakdown:**
- Schema agent implementation: 1.5 hours
- Document analysis logic: 1 hour
- CLI command and display: 1 hour
- Testing with various document types: 0.5 hour

---

## Feature 1.8: AI Chunking Strategy Suggestion

### Goal
Enable users to determine optimal document chunking strategies by analyzing sample documents using an AI agent. The agent examines document structure, content patterns, and use case requirements to suggest chunk size, overlap, and chunking method.

### User Stories
- As a RAG developer, I want to know the optimal chunk size for my documents without trial-and-error
- As a data engineer, I want AI to recommend chunking strategies that balance context and precision
- As a new user, I want to understand the trade-offs between different chunking approaches

### The Chunking Problem

**Too Small Chunks (e.g., 100 tokens):**
- ❌ Loss of context and coherence
- ❌ Too many embeddings → high cost
- ❌ Poor retrieval quality
- ✅ Fine-grained search

**Too Large Chunks (e.g., 2000 tokens):**
- ❌ Less precise retrieval
- ❌ Context window overflow
- ❌ Slower processing
- ✅ Better context preservation

**Optimal Chunking:**
- ✅ Balances context and precision
- ✅ Matches document structure
- ✅ Optimizes for use case
- ✅ Minimizes cost while maximizing quality

### CLI Interface

```bash
# Analyze and suggest chunking strategy
weave chunk suggest ./samples \
  --collection documents \
  --output chunking-config.yaml

# With specific use case
weave chunk suggest ./samples \
  --use-case "semantic search on technical documentation" \
  --output chunking-config.yaml

# Interactive mode with visual examples
weave chunk suggest ./samples \
  --interactive \
  --show-examples

# Test different strategies
weave chunk suggest ./samples \
  --test \
  --strategies "fixed,semantic,recursive"

# Compare chunk sizes
weave chunk suggest ./samples \
  --compare "256,512,1024" \
  --output comparison-report.json

# Apply to pipeline immediately
weave chunk suggest ./samples \
  --collection docs \
  --apply-to-pipeline
```

### Technical Architecture

**Package Structure:**
```
src/pkg/agents/
├── chunking_agent.go      # AI chunking analysis agent

src/pkg/chunking/
├── strategies.go          # Chunking strategy implementations
├── analyzer.go            # Document structure analysis
├── evaluator.go           # Chunk quality evaluation
└── types.go               # Chunking types and configs

src/cmd/chunk/
├── chunk.go               # Chunk command group
├── suggest.go             # Suggest subcommand
└── test.go                # Test subcommand
```

**Core Types:**

```go
// ChunkingAnalysisInput represents input for chunking analysis
type ChunkingAnalysisInput struct {
    SampleFiles      []string          // Paths to sample documents
    UseCase          string            // User's use case (optional)
    MaxSamples       int               // Max files to analyze (default: 30)
    TestStrategies   []string          // Strategies to test
}

// ChunkingAnalysisOutput represents AI-generated chunking suggestion
type ChunkingAnalysisOutput struct {
    Recommendation   ChunkingConfig    // Recommended configuration
    Reasoning        string            // Why this strategy was chosen
    Confidence       float64           // Confidence score (0-1)
    Alternatives     []ChunkingConfig  // Alternative configurations
    Analysis         DocumentAnalysis  // Detailed analysis
    Examples         []ChunkExample    // Example chunks
    Performance      PerformanceEst    // Performance estimates
    Warnings         []string          // Potential issues
}

// ChunkingConfig represents a chunking configuration
type ChunkingConfig struct {
    Strategy      string  `yaml:"strategy"`       // fixed, semantic, recursive, paragraph
    ChunkSize     int     `yaml:"chunk_size"`     // Tokens per chunk
    OverlapSize   int     `yaml:"overlap_size"`   // Token overlap
    OverlapPct    float64 `yaml:"overlap_pct"`    // Percentage overlap
    MinChunkSize  int     `yaml:"min_chunk_size"` // Minimum chunk size
    MaxChunkSize  int     `yaml:"max_chunk_size"` // Maximum chunk size
    Separators    []string `yaml:"separators,omitempty"` // For recursive
    Description   string  `yaml:"description"`
}

// ChunkingStrategy defines different chunking approaches
type ChunkingStrategy string

const (
    StrategyFixed      ChunkingStrategy = "fixed"      // Fixed-size chunks
    StrategySemantic   ChunkingStrategy = "semantic"   // Semantic boundaries
    StrategyRecursive  ChunkingStrategy = "recursive"  // Recursive splitting
    StrategyParagraph  ChunkingStrategy = "paragraph"  // Paragraph-based
    StrategySentence   ChunkingStrategy = "sentence"   // Sentence-based
)

// DocumentAnalysis represents document structure analysis
type DocumentAnalysis struct {
    AverageParagraphLength int              `json:"avg_paragraph_length"`
    AverageSentenceLength  int              `json:"avg_sentence_length"`
    ParagraphCount         int              `json:"paragraph_count"`
    SentenceCount          int              `json:"sentence_count"`
    HasSections            bool             `json:"has_sections"`
    SectionDepth           int              `json:"section_depth"`
    CodePercentage         float64          `json:"code_percentage"`
    ContentType            string           `json:"content_type"` // prose, technical, code, mixed
    StructureComplexity    string           `json:"structure_complexity"` // simple, moderate, complex
}

// ChunkExample shows an example chunk
type ChunkExample struct {
    Text         string  `json:"text"`
    TokenCount   int     `json:"token_count"`
    ChunkIndex   int     `json:"chunk_index"`
    SourceFile   string  `json:"source_file"`
    Quality      string  `json:"quality"` // good, acceptable, poor
    Explanation  string  `json:"explanation"`
}

// PerformanceEst estimates performance metrics
type PerformanceEst struct {
    EstimatedChunks     int     `json:"estimated_chunks"`
    EstimatedTokens     int     `json:"estimated_tokens"`
    EstimatedCost       float64 `json:"estimated_cost_usd"`
    RetrievalQuality    string  `json:"retrieval_quality"`  // excellent, good, fair, poor
    ContextPreservation string  `json:"context_preservation"` // excellent, good, fair, poor
}
```

### Implementation

**Chunking Analysis Agent:**

```go
// ChunkingAgent analyzes documents and suggests chunking strategies
type ChunkingAgent struct {
    llmClient *llm.OpenAIClient
}

func NewChunkingAgent(llmClient *llm.OpenAIClient) *ChunkingAgent {
    return &ChunkingAgent{llmClient: llmClient}
}

func (a *ChunkingAgent) Name() string {
    return "chunking-agent"
}

func (a *ChunkingAgent) Execute(ctx context.Context, input interface{}) (interface{}, error) {
    analysisInput := input.(*ChunkingAnalysisInput)

    // Step 1: Sample and analyze documents
    samples := a.extractSamples(analysisInput.SampleFiles, analysisInput.MaxSamples)

    // Step 2: Analyze document structure
    analysis := a.analyzeDocumentStructure(samples)

    // Step 3: Test different chunking strategies
    strategyResults := a.testStrategies(samples, analysis)

    // Step 4: Use LLM to recommend optimal strategy
    prompt := a.buildAnalysisPrompt(analysis, strategyResults, analysisInput)

    response, err := a.llmClient.CompleteStructured(ctx, prompt, &ChunkingAnalysisOutput{})
    if err != nil {
        return nil, fmt.Errorf("failed to analyze chunking strategy: %w", err)
    }

    output := response.(*ChunkingAnalysisOutput)
    output.Analysis = analysis

    // Step 5: Generate examples
    output.Examples = a.generateExamples(samples[0], output.Recommendation, 3)

    return output, nil
}

func (a *ChunkingAgent) buildAnalysisPrompt(analysis DocumentAnalysis, results map[string]StrategyResult, input *ChunkingAnalysisInput) string {
    return fmt.Sprintf(`Analyze the following document characteristics and recommend an optimal chunking strategy.

Use Case: %s

Document Analysis:
- Content Type: %s
- Average Paragraph Length: %d tokens
- Average Sentence Length: %d tokens
- Has Sections: %v
- Section Depth: %d
- Code Percentage: %.1f%%
- Structure Complexity: %s

Strategy Test Results:
%s

Please recommend:
1. Optimal chunking strategy (fixed, semantic, recursive, paragraph, sentence)
2. Specific chunk size in tokens
3. Overlap size/percentage
4. Reasoning for the choice
5. Confidence score (0-1)
6. Alternative configurations
7. Performance estimates (cost, quality)
8. Any warnings or considerations

Consider:
- Context preservation vs precision
- Token/cost efficiency
- Retrieval quality
- Document structure alignment
- Use case requirements

Return the response in JSON format following the ChunkingAnalysisOutput structure.`,
        input.UseCase,
        analysis.ContentType,
        analysis.AverageParagraphLength,
        analysis.AverageSentenceLength,
        analysis.HasSections,
        analysis.SectionDepth,
        analysis.CodePercentage,
        analysis.StructureComplexity,
        formatStrategyResults(results),
    )
}
```

**Document Structure Analysis:**

```go
func (a *ChunkingAgent) analyzeDocumentStructure(samples []string) DocumentAnalysis {
    var totalParagraphs, totalSentences int
    var totalParagraphTokens, totalSentenceTokens int
    var hasSections bool
    var maxSectionDepth int
    var codeLines, totalLines int

    for _, content := range samples {
        // Count paragraphs
        paragraphs := strings.Split(content, "\n\n")
        totalParagraphs += len(paragraphs)
        for _, p := range paragraphs {
            totalParagraphTokens += estimateTokens(p)
        }

        // Count sentences
        sentences := splitSentences(content)
        totalSentences += len(sentences)
        for _, s := range sentences {
            totalSentenceTokens += estimateTokens(s)
        }

        // Detect sections (markdown headers, etc.)
        if strings.Contains(content, "#") || strings.Contains(content, "==") {
            hasSections = true
            depth := detectSectionDepth(content)
            if depth > maxSectionDepth {
                maxSectionDepth = depth
            }
        }

        // Detect code blocks
        lines := strings.Split(content, "\n")
        totalLines += len(lines)
        for _, line := range lines {
            if isCodeLine(line) {
                codeLines++
            }
        }
    }

    avgParagraphLength := 0
    if totalParagraphs > 0 {
        avgParagraphLength = totalParagraphTokens / totalParagraphs
    }

    avgSentenceLength := 0
    if totalSentences > 0 {
        avgSentenceLength = totalSentenceTokens / totalSentences
    }

    codePercentage := 0.0
    if totalLines > 0 {
        codePercentage = float64(codeLines) / float64(totalLines) * 100
    }

    // Determine content type
    contentType := determineContentType(codePercentage, hasSections, avgParagraphLength)

    // Determine structure complexity
    complexity := determineComplexity(hasSections, maxSectionDepth, avgParagraphLength)

    return DocumentAnalysis{
        AverageParagraphLength: avgParagraphLength,
        AverageSentenceLength:  avgSentenceLength,
        ParagraphCount:         totalParagraphs,
        SentenceCount:          totalSentences,
        HasSections:            hasSections,
        SectionDepth:           maxSectionDepth,
        CodePercentage:         codePercentage,
        ContentType:            contentType,
        StructureComplexity:    complexity,
    }
}

func determineContentType(codePct float64, hasSections bool, avgParagraph int) string {
    if codePct > 50 {
        return "code"
    }
    if codePct > 20 {
        return "mixed"
    }
    if hasSections && avgParagraph > 300 {
        return "technical"
    }
    return "prose"
}

func determineComplexity(hasSections bool, depth int, avgParagraph int) string {
    if !hasSections && avgParagraph < 200 {
        return "simple"
    }
    if depth > 3 || avgParagraph > 500 {
        return "complex"
    }
    return "moderate"
}
```

**Strategy Testing:**

```go
type StrategyResult struct {
    Strategy        ChunkingStrategy
    AverageChunks   int
    AverageTokens   int
    ContextQuality  float64 // 0-1
    Examples        []string
}

func (a *ChunkingAgent) testStrategies(samples []string, analysis DocumentAnalysis) map[string]StrategyResult {
    strategies := []ChunkingStrategy{
        StrategyFixed,
        StrategySemantic,
        StrategyParagraph,
    }

    results := make(map[string]StrategyResult)

    for _, strategy := range strategies {
        result := a.testStrategy(samples[0], strategy, analysis)
        results[string(strategy)] = result
    }

    return results
}

func (a *ChunkingAgent) testStrategy(content string, strategy ChunkingStrategy, analysis DocumentAnalysis) StrategyResult {
    var chunks []string
    var avgTokens int

    switch strategy {
    case StrategyFixed:
        // Test with recommended size based on analysis
        size := a.recommendFixedSize(analysis)
        chunks = chunkFixed(content, size, size/10) // 10% overlap
    case StrategySemantic:
        chunks = chunkSemantic(content)
    case StrategyParagraph:
        chunks = chunkByParagraph(content)
    }

    totalTokens := 0
    for _, chunk := range chunks {
        totalTokens += estimateTokens(chunk)
    }
    if len(chunks) > 0 {
        avgTokens = totalTokens / len(chunks)
    }

    // Evaluate context quality
    quality := a.evaluateContextQuality(chunks, analysis)

    return StrategyResult{
        Strategy:       strategy,
        AverageChunks:  len(chunks),
        AverageTokens:  avgTokens,
        ContextQuality: quality,
        Examples:       chunks[:min(3, len(chunks))],
    }
}

func (a *ChunkingAgent) recommendFixedSize(analysis DocumentAnalysis) int {
    // Start with paragraph-based recommendation
    baseSize := analysis.AverageParagraphLength

    // Adjust based on content type
    switch analysis.ContentType {
    case "code":
        return clamp(baseSize*2, 256, 1024) // Code needs more context
    case "technical":
        return clamp(baseSize, 512, 1024)   // Technical docs medium chunks
    case "prose":
        return clamp(baseSize, 256, 512)    // Prose smaller chunks
    default:
        return 512
    }
}

func (a *ChunkingAgent) evaluateContextQuality(chunks []string, analysis DocumentAnalysis) float64 {
    score := 1.0

    // Penalize very small or very large chunks
    for _, chunk := range chunks {
        tokens := estimateTokens(chunk)
        if tokens < 100 {
            score -= 0.1
        }
        if tokens > 2000 {
            score -= 0.1
        }
    }

    // Reward chunks that align with structure
    if analysis.HasSections {
        sectionAligned := 0
        for _, chunk := range chunks {
            if strings.Contains(chunk, "#") || strings.Contains(chunk, "==") {
                sectionAligned++
            }
        }
        alignmentRatio := float64(sectionAligned) / float64(len(chunks))
        score += alignmentRatio * 0.2
    }

    return clamp(score, 0.0, 1.0)
}
```

**CLI Command:**

```go
var suggestCmd = &cobra.Command{
    Use:   "suggest SOURCE",
    Short: "Suggest optimal chunking strategy by analyzing sample documents",
    Long: `Analyze sample documents and suggest an optimal chunking strategy using AI.

The AI agent examines document structure, content patterns, and use case
requirements to recommend chunk size, overlap, and chunking method.

Examples:
  # Analyze samples and output chunking config
  weave chunk suggest ./samples --output chunking-config.yaml

  # With specific use case
  weave chunk suggest ./samples \
    --use-case "semantic search on code documentation" \
    --output config.yaml

  # Interactive mode with visual examples
  weave chunk suggest ./samples --interactive --show-examples

  # Test and compare strategies
  weave chunk suggest ./samples --test --compare "256,512,1024"

  # Apply to pipeline
  weave chunk suggest ./samples --apply-to-pipeline`,
    Args: cobra.ExactArgs(1),
    RunE: runChunkSuggest,
}

func runChunkSuggest(cmd *cobra.Command, args []string) error {
    source := args[0]
    ctx := context.Background()

    // Create LLM client
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        return fmt.Errorf("OPENAI_API_KEY required for AI chunking suggestion")
    }

    llmClient, err := llm.NewOpenAIClient(apiKey)
    if err != nil {
        return fmt.Errorf("failed to create LLM client: %w", err)
    }

    // Scan for sample files
    scanner := pipeline.NewFileScanner(source, glob, exclude, true)
    files, err := scanner.Scan(ctx)
    if err != nil {
        return fmt.Errorf("failed to scan files: %w", err)
    }

    fmt.Printf("🔍 Analyzing %d sample documents for chunking...\n", len(files))

    // Extract file paths
    var filePaths []string
    for _, file := range files {
        filePaths = append(filePaths, file.Path)
    }

    // Create chunking agent
    agent := agents.NewChunkingAgent(llmClient)

    // Analyze and suggest chunking strategy
    input := &agents.ChunkingAnalysisInput{
        SampleFiles:    filePaths,
        UseCase:        useCase,
        MaxSamples:     maxSamples,
        TestStrategies: testStrategies,
    }

    result, err := agent.Execute(ctx, input)
    if err != nil {
        return fmt.Errorf("chunking analysis failed: %w", err)
    }

    output := result.(*agents.ChunkingAnalysisOutput)

    // Display results
    displayChunkingAnalysis(output, interactive, showExamples)

    // Save to file if requested
    if outputFile != "" {
        if err := saveChunkingConfig(output.Recommendation, outputFile); err != nil {
            return fmt.Errorf("failed to save config: %w", err)
        }
        fmt.Printf("✅ Chunking config saved to %s\n", outputFile)
    }

    return nil
}

func displayChunkingAnalysis(output *agents.ChunkingAnalysisOutput, interactive bool, showExamples bool) {
    fmt.Printf("\n📋 Recommended Chunking Strategy\n")
    fmt.Printf("   Confidence: %.1f%%\n\n", output.Confidence*100)

    rec := output.Recommendation
    fmt.Printf("🔧 Configuration:\n")
    fmt.Printf("   Strategy: %s\n", rec.Strategy)
    fmt.Printf("   Chunk Size: %d tokens\n", rec.ChunkSize)
    fmt.Printf("   Overlap: %d tokens (%.0f%%)\n", rec.OverlapSize, rec.OverlapPct*100)
    if rec.MinChunkSize > 0 {
        fmt.Printf("   Min/Max Size: %d-%d tokens\n", rec.MinChunkSize, rec.MaxChunkSize)
    }

    fmt.Printf("\n📊 Document Analysis:\n")
    fmt.Printf("   Content Type: %s\n", output.Analysis.ContentType)
    fmt.Printf("   Avg Paragraph: %d tokens\n", output.Analysis.AverageParagraphLength)
    fmt.Printf("   Structure: %s\n", output.Analysis.StructureComplexity)
    if output.Analysis.CodePercentage > 0 {
        fmt.Printf("   Code: %.1f%%\n", output.Analysis.CodePercentage)
    }

    fmt.Printf("\n💡 Reasoning:\n%s\n", output.Reasoning)

    if showExamples && len(output.Examples) > 0 {
        fmt.Println("\n📝 Example Chunks:")
        for i, example := range output.Examples {
            fmt.Printf("\n   Example %d (%d tokens, %s quality):\n", i+1, example.TokenCount, example.Quality)
            preview := example.Text
            if len(preview) > 200 {
                preview = preview[:200] + "..."
            }
            fmt.Printf("   %s\n", preview)
            if example.Explanation != "" {
                fmt.Printf("   → %s\n", example.Explanation)
            }
        }
    }

    fmt.Printf("\n📈 Performance Estimates:\n")
    perf := output.Performance
    fmt.Printf("   Chunks per document: ~%d\n", perf.EstimatedChunks)
    fmt.Printf("   Total tokens: ~%d\n", perf.EstimatedTokens)
    fmt.Printf("   Embedding cost: ~$%.4f per document\n", perf.EstimatedCost)
    fmt.Printf("   Retrieval quality: %s\n", perf.RetrievalQuality)
    fmt.Printf("   Context preservation: %s\n", perf.ContextPreservation)

    if len(output.Warnings) > 0 {
        fmt.Println("\n⚠️  Warnings:")
        for _, warning := range output.Warnings {
            fmt.Printf("   • %s\n", warning)
        }
    }

    if interactive && len(output.Alternatives) > 0 {
        fmt.Println("\n🔀 Alternative Strategies:")
        for i, alt := range output.Alternatives {
            fmt.Printf("   %d. %s (%d tokens, %.0f%% overlap)\n",
                i+1, alt.Strategy, alt.ChunkSize, alt.OverlapPct*100)
        }
    }
}
```

### Integration with Existing Features

**1. Pipeline Integration:**
```bash
# Suggest chunking, then ingest with optimal size
weave chunk suggest ./docs --output chunking.yaml
weave pipeline ingest ./docs \
  --collection docs \
  --chunking-config chunking.yaml \
  --vdb qdrant-cloud
```

**2. Schema + Chunking Workflow:**
```bash
# Complete AI-assisted setup
weave schema suggest ./samples --collection docs --output schema.yaml
weave chunk suggest ./samples --output chunking.yaml
weave cols create docs --schema schema.yaml --vdb qdrant-cloud
weave pipeline ingest ./docs --chunking-config chunking.yaml --collection docs
```

**3. MCP Server Integration:**
```json
{
  "tool": "suggest_chunking",
  "arguments": {
    "samples": "./docs",
    "use_case": "technical documentation search"
  }
}
```

**4. REPL Integration:**
```
weave> chunk suggest ./samples
🔍 Analyzing documents...
📋 Recommended: 512 tokens, 10% overlap
weave> chunk test --show-examples
[displays chunked examples]
weave> chunk apply
✅ Chunking config saved
```

### Success Metrics

- Recommendation accuracy: >85% user acceptance
- Analysis time: <20 seconds for 30 samples
- Performance improvement: 30%+ better retrieval quality vs random chunk size
- Cost optimization: 20%+ token savings vs unoptimized chunking

### Effort: 3-4 hours

**Breakdown:**
- Chunking agent implementation: 1.5 hours
- Document structure analysis: 1 hour
- Strategy testing framework: 0.75 hour
- CLI command and display: 0.75 hour

---

## Total Timeline

**Week 1:**
- Day 1-2: Pipeline Commands (4-6h) ✅ DONE
- Day 3: Progress Bars + JSON/YAML Output (3-5h) ✅ DONE

**Week 2:**
- Day 1-2: REPL Mode (3-4h)
- Day 3-4: MCP Server (5-6h)
- Day 5: Collection Stats (2-3h)

**Week 3:**
- Day 1-2: AI Schema Suggestion (3-4h)
- Day 3: AI Chunking Suggestion (3-4h)

**Week 4 (Optional):**
- Day 1: CI/CD Integration (3-4h)

**Total: 28-37 hours across 3-4 weeks**

**Progress: 3/8 features completed (38%)**

**Completed Features:**
- ✅ 1.1 Pipeline Commands
- ✅ 1.4 Progress Bars
- ✅ 1.5 JSON/YAML Output

**Planned Features:**
- 📝 1.2 Interactive REPL
- 📝 1.3 MCP Server
- 📝 1.6 Collection Statistics
- 📝 1.7 AI Schema Suggestion (NEW!)
- 📝 1.8 AI Chunking Suggestion (NEW!)
- 📝 1.1a CI/CD Integration
