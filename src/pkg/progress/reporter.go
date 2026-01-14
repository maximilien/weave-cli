// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package progress

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Reporter provides progress reporting functionality for long-running operations
type Reporter struct {
	enabled    bool
	jsonOutput bool
	startTime  time.Time
	total      int
	current    int
	lastUpdate time.Time
}

// ProgressMessage represents a progress update in JSON format
type ProgressMessage struct {
	Type       string  `json:"type"`
	Phase      string  `json:"phase"`
	Message    string  `json:"message"`
	Elapsed    float64 `json:"elapsed"`
	Progress   float64 `json:"progress,omitempty"`   // 0-100 percentage
	Current    int     `json:"current,omitempty"`    // Current item count
	Total      int     `json:"total,omitempty"`      // Total item count
	ETA        float64 `json:"eta,omitempty"`        // Estimated seconds remaining
	Throughput float64 `json:"throughput,omitempty"` // Items per second
}

// NewReporter creates a new progress reporter
func NewReporter(enabled bool) *Reporter {
	return &Reporter{
		enabled:    enabled,
		lastUpdate: time.Now(),
	}
}

// NewJSONReporter creates a new progress reporter with JSON output
func NewJSONReporter(enabled bool) *Reporter {
	return &Reporter{
		enabled:    enabled,
		jsonOutput: true,
		lastUpdate: time.Now(),
	}
}

// SetTotal sets the total number of items to process (for percentage/ETA calculation)
func (p *Reporter) SetTotal(total int) {
	p.total = total
	p.current = 0
}

// Start begins a progress reporting session with an initial message
func (p *Reporter) Start(msg string) {
	if !p.enabled {
		return
	}
	p.startTime = time.Now()

	if p.jsonOutput {
		p.outputJSON("start", msg, 0.0)
	} else {
		// Write to stderr to keep stdout clean for actual output
		fmt.Fprintf(os.Stderr, "🔍 %s\n", msg)
	}
}

// Update reports progress with a message and elapsed time
func (p *Reporter) Update(msg string) {
	if !p.enabled {
		return
	}
	elapsed := time.Since(p.startTime).Truncate(100 * time.Millisecond)
	p.lastUpdate = time.Now()

	if p.jsonOutput {
		p.outputJSONWithProgress("progress", msg, elapsed.Seconds())
	} else {
		progressBar := p.formatProgressBar()
		if progressBar != "" {
			fmt.Fprintf(os.Stderr, "   %s %s (%.1fs)\n", progressBar, msg, elapsed.Seconds())
		} else {
			fmt.Fprintf(os.Stderr, "   %s (%.1fs)\n", msg, elapsed.Seconds())
		}
	}
}

// UpdateProgress reports progress with a specific count (increments current)
func (p *Reporter) UpdateProgress(msg string) {
	if !p.enabled {
		return
	}
	p.current++
	p.Update(msg)
}

// UpdateWithCount reports progress with an explicit current count
func (p *Reporter) UpdateWithCount(current int, msg string) {
	if !p.enabled {
		return
	}
	p.current = current
	p.Update(msg)
}

// Complete marks the operation as complete with a final message
func (p *Reporter) Complete(msg string) {
	if !p.enabled {
		return
	}
	elapsed := time.Since(p.startTime)

	if p.jsonOutput {
		p.outputJSON("complete", msg, elapsed.Seconds())
	} else {
		// Write to stderr to keep stdout clean for actual output
		fmt.Fprintf(os.Stderr, "✅ %s (%.2fs)\n\n", msg, elapsed.Seconds())
	}
}

// outputJSON outputs a progress message in JSON format to stdout
func (p *Reporter) outputJSON(phase, message string, elapsed float64) {
	msg := ProgressMessage{
		Type:    "progress",
		Phase:   phase,
		Message: message,
		Elapsed: elapsed,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	fmt.Println(string(data))
}

// outputJSONWithProgress outputs a progress message with percentage/ETA
func (p *Reporter) outputJSONWithProgress(phase, message string, elapsed float64) {
	msg := ProgressMessage{
		Type:    "progress",
		Phase:   phase,
		Message: message,
		Elapsed: elapsed,
	}

	// Add progress metrics if total is set
	if p.total > 0 {
		msg.Current = p.current
		msg.Total = p.total
		msg.Progress = float64(p.current) / float64(p.total) * 100.0

		// Calculate ETA if we have progress
		if p.current > 0 && elapsed > 0 {
			itemsPerSecond := float64(p.current) / elapsed
			remainingItems := p.total - p.current
			msg.ETA = float64(remainingItems) / itemsPerSecond
			msg.Throughput = itemsPerSecond
		}
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	fmt.Println(string(data))
}

// formatProgressBar creates a visual progress bar
func (p *Reporter) formatProgressBar() string {
	if p.total == 0 {
		return ""
	}

	percentage := float64(p.current) / float64(p.total) * 100.0
	barWidth := 20
	filled := int(percentage / 100.0 * float64(barWidth))

	bar := "["
	for i := 0; i < barWidth; i++ {
		if i < filled {
			bar += "="
		} else if i == filled {
			bar += ">"
		} else {
			bar += " "
		}
	}
	bar += "]"

	// Add ETA if available
	eta := p.calculateETA()
	if eta > 0 {
		return fmt.Sprintf("%s %3.0f%% (%d/%d, ETA: %s)", bar, percentage, p.current, p.total, p.formatDuration(eta))
	}

	return fmt.Sprintf("%s %3.0f%% (%d/%d)", bar, percentage, p.current, p.total)
}

// calculateETA estimates time remaining in seconds
func (p *Reporter) calculateETA() float64 {
	if p.current == 0 || p.total == 0 {
		return 0
	}

	elapsed := time.Since(p.startTime).Seconds()
	if elapsed == 0 {
		return 0
	}

	itemsPerSecond := float64(p.current) / elapsed
	remainingItems := p.total - p.current

	return float64(remainingItems) / itemsPerSecond
}

// formatDuration formats seconds into human-readable duration
func (p *Reporter) formatDuration(seconds float64) string {
	if seconds < 60 {
		return fmt.Sprintf("%.0fs", seconds)
	} else if seconds < 3600 {
		return fmt.Sprintf("%.0fm%.0fs", seconds/60, float64(int(seconds)%60))
	}
	hours := int(seconds / 3600)
	minutes := int(seconds/60) % 60
	return fmt.Sprintf("%dh%dm", hours, minutes)
}
