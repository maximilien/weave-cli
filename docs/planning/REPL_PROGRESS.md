# REPL Script Execution Progress Improvement

**Status**: ✅ **COMPLETED** (All P0, P1, and P2 features implemented)

## Summary

Successfully implemented real-time progress feedback for REPL script execution, providing users with:
- **Real-time output streaming**: See results as they arrive (no more silent waits)
- **Progress spinner**: Animated spinner appears after 2s for slow operations
- **Elapsed time**: Shows actual time spent (e.g., `[8s]`)
- **Time estimation**: After 5s, estimates remaining time (e.g., `(~10s remaining)`)
- **Item counts**: Tracks and displays intermediate results (e.g., `11 items`)

**Before**: Users waited 10+ seconds with no feedback
**After**: Continuous visual feedback with spinner, timing, and progress indicators

## Problem

When executing bash scripts in REPL mode, users experience long waits (10+ seconds) with no feedback between "⏳ Step X" start and "✓ Step X completed".

**Example:**
```
⏳ Step 2: Identify and list empty collections
[10.8 seconds of silence - no feedback]
✓ Step 2 completed (10.8s)
```

## Solution: Real-Time Output Streaming

Stream script output to the user while it executes, providing immediate feedback.

### Inspiration: Claude Code

Claude Code provides excellent progress feedback during operations:
- **Real-time streaming** of command output
- **Progress spinners** for long-running tasks
- **Intermediate results** shown as they arrive
- **Clear visual hierarchy** with dimmed/faded text for less important info
- **Status indicators** (⏳, ✓, ⚠️) for each operation

We should follow similar UX patterns for consistency and quality.

### Implementation

**File:** `src/pkg/agents/bash_agent.go`

**Current Code** (lines 103-108):
```go
var stdout, stderr bytes.Buffer
execCmd.Stdout = &stdout
execCmd.Stderr = &stderr

// Execute the command
err := execCmd.Run()
```

**Improved Code:**
```go
// Create buffers to capture output
var stdout, stderr bytes.Buffer

// Create multi-writers to stream AND capture output
stdoutWriter := io.MultiWriter(&stdout, &progressWriter{prefix: "  "})
stderrWriter := io.MultiWriter(&stderr, &progressWriter{prefix: "  ⚠️  "})

execCmd.Stdout = stdoutWriter
execCmd.Stderr = stderrWriter

// Execute the command
err := execCmd.Run()
```

**Add Progress Writer:**
```go
// progressWriter writes output line-by-line with a prefix
type progressWriter struct {
	prefix string
	buffer []byte
}

func (pw *progressWriter) Write(p []byte) (n int, err error) {
	pw.buffer = append(pw.buffer, p...)

	// Process complete lines
	for {
		idx := bytes.IndexByte(pw.buffer, '\n')
		if idx == -1 {
			break
		}

		line := pw.buffer[:idx]
		pw.buffer = pw.buffer[idx+1:]

		// Print line with prefix (dimmed for less distraction)
		dim := color.New(color.Faint).SprintFunc()
		fmt.Printf("%s%s\n", pw.prefix, dim(string(line)))
	}

	return len(p), nil
}

func (pw *progressWriter) Flush() {
	if len(pw.buffer) > 0 {
		dim := color.New(color.Faint).SprintFunc()
		fmt.Printf("%s%s\n", pw.prefix, dim(string(pw.buffer)))
		pw.buffer = nil
	}
}
```

### User Experience Improvement

**Before:**
```
⏳ Step 2: Identify and list empty collections
[long silence]
✓ Step 2 completed (10.8s)
  CalendarMaxDocs_test
  DemoImage
  ...
```

**After:**
```
⏳ Step 2: Identify and list empty collections
  CalendarMaxDocs_test
  DemoImage
  NotesMaxDocs_test
  PortfolioMaxDocs
  ...
✓ Step 2 completed (10.8s)
```

### Additional Enhancements

#### 1. Progress Spinner for Long Operations

Add a spinner that updates every 500ms during execution:

```go
// Start progress indicator in goroutine
done := make(chan bool)
go func() {
	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	i := 0
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			fmt.Printf("\r  %s Processing...", spinner[i%len(spinner)])
			i++
		}
	}
}()

// Execute command
err := execCmd.Run()

// Stop spinner
done <- true
fmt.Print("\r  \r") // Clear spinner line
```

#### 2. Intermediate Results Display

For commands that produce multiple items (like listing collections), show count updates:

```
⏳ Step 2: Identify and list empty collections
  Checking: CalendarMaxDocs_test ✓ empty
  Checking: DemoImage ✓ empty
  Checking: WeaveDocs ✗ has 42 documents
  ...
  Found 11 empty collections
✓ Step 2 completed (10.8s)
```

#### 3. Time Estimation

After 2 seconds, show estimated completion:

```
⏳ Step 2: Identify and list empty collections (est. 8-12s)
  ...
```

## Implementation Priority

### P0 (Must Have) - Immediate Value ✅ COMPLETED
- ✅ Real-time output streaming
- ✅ Line-by-line display with dimmed text

### P1 (Should Have) - Significant UX improvement ✅ COMPLETED
- ✅ Progress spinner for operations >2 seconds
- ✅ Flush any remaining output at completion

### P2 (Nice to Have) - Polish ✅ COMPLETED
- ✅ Elapsed time display (shows actual time spent)
- ✅ Time estimation after 5s (estimates remaining time)
- ✅ Intermediate result counts (shows "N items" as they're processed)
- ⏭️ Progress percentage (deferred - requires knowing total count upfront)

## Files to Modify

1. **`src/pkg/agents/bash_agent.go`**
   - Add `progressWriter` type
   - Modify `Execute()` method to use MultiWriter
   - Add spinner support

2. **`src/pkg/agents/output_agent.go`**
   - Add `PrintProgress()` method
   - Add `StartSpinner()` / `StopSpinner()` methods

## Testing

```bash
# Test with a slow operation
weave
> show me all empty collections

# Should see:
# - Each collection name as it's checked
# - Real-time feedback
# - No long silent periods
```

## Benefits

1. **Better UX** - Users know the system is working
2. **Debugging** - See where scripts get stuck
3. **Transparency** - Understand what's happening
4. **Trust** - Reduces anxiety during long operations

## Estimated Effort

- P0 implementation: 1-2 hours
- P1 features: 2-3 hours
- P2 polish: 2-4 hours
- Testing: 1 hour

**Total: 6-10 hours for complete implementation**

## Related

- Similar to how `docker build` shows real-time output
- Similar to `npm install` progress indicators
- Similar to `pytest -v` verbose mode
