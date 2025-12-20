package pipeline

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/schollz/progressbar/v3"
)

// ProgressTracker handles progress reporting for pipeline operations
type ProgressTracker struct {
	enabled bool
	quiet   bool
	bar     *progressbar.ProgressBar
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(enabled bool, quiet bool) *ProgressTracker {
	return &ProgressTracker{
		enabled: enabled && !quiet,
		quiet:   quiet,
	}
}

// StartScanning shows scanning progress
func (p *ProgressTracker) StartScanning() {
	if !p.enabled {
		return
	}
	fmt.Fprintf(os.Stderr, "🔍 Scanning files...\n")
}

// FinishScanning completes scanning progress
func (p *ProgressTracker) FinishScanning(count int) {
	if !p.enabled {
		return
	}
	fmt.Fprintf(os.Stderr, "✅ Found %d files\n\n", count)
}

// StartProcessing creates a progress bar for file processing
func (p *ProgressTracker) StartProcessing(total int, batchSize int, workers int) {
	if !p.enabled {
		return
	}

	fmt.Fprintf(os.Stderr, "⚙️  Processing files (batch_size=%d, workers=%d)...\n", batchSize, workers)

	p.bar = progressbar.NewOptions(total,
		progressbar.OptionSetDescription("Processing"),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetItsString("files"),
		progressbar.OptionThrottle(100*time.Millisecond),
		progressbar.OptionSetWidth(30),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerPadding: "░",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetPredictTime(true),
	)
}

// UpdateProgress increments the progress bar
func (p *ProgressTracker) UpdateProgress(delta int) {
	if !p.enabled || p.bar == nil {
		return
	}
	p.bar.Add(delta)
}

// FinishProcessing completes the progress bar
func (p *ProgressTracker) FinishProcessing() {
	if !p.enabled || p.bar == nil {
		return
	}
	p.bar.Finish()
	fmt.Fprintf(os.Stderr, "\n")
}

// StartBatching shows batch creation progress
func (p *ProgressTracker) StartBatching(totalDocs int, batchSize int) {
	if !p.enabled {
		return
	}
	batches := (totalDocs + batchSize - 1) / batchSize
	fmt.Fprintf(os.Stderr, "📦 Creating documents in %d batches...\n", batches)
}

// UpdateBatch shows progress for batch operations
func (p *ProgressTracker) UpdateBatch(batchNum int, totalBatches int, docsInBatch int) {
	if !p.enabled {
		return
	}
	fmt.Fprintf(os.Stderr, "\r   Batch %d/%d (%d docs)...", batchNum, totalBatches, docsInBatch)
}

// FinishBatching completes batch creation
func (p *ProgressTracker) FinishBatching() {
	if !p.enabled {
		return
	}
	fmt.Fprintf(os.Stderr, "\r✅ Documents created successfully\n\n")
}

// ShowError displays an error message
func (p *ProgressTracker) ShowError(msg string) {
	if p.quiet {
		return
	}
	fmt.Fprintf(os.Stderr, "❌ %s\n", msg)
}

// ShowWarning displays a warning message
func (p *ProgressTracker) ShowWarning(msg string) {
	if p.quiet {
		return
	}
	fmt.Fprintf(os.Stderr, "⚠️  %s\n", msg)
}

// ShowInfo displays an info message
func (p *ProgressTracker) ShowInfo(msg string) {
	if p.quiet {
		return
	}
	fmt.Fprintf(os.Stderr, "ℹ️  %s\n", msg)
}

// GetWriter returns a writer for the progress output
func (p *ProgressTracker) GetWriter() io.Writer {
	if p.enabled {
		return os.Stderr
	}
	return io.Discard
}
