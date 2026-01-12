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
}

// ProgressMessage represents a progress update in JSON format
type ProgressMessage struct {
	Type    string  `json:"type"`
	Phase   string  `json:"phase"`
	Message string  `json:"message"`
	Elapsed float64 `json:"elapsed"`
}

// NewReporter creates a new progress reporter
func NewReporter(enabled bool) *Reporter {
	return &Reporter{
		enabled: enabled,
	}
}

// NewJSONReporter creates a new progress reporter with JSON output
func NewJSONReporter(enabled bool) *Reporter {
	return &Reporter{
		enabled:    enabled,
		jsonOutput: true,
	}
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

	if p.jsonOutput {
		p.outputJSON("progress", msg, elapsed.Seconds())
	} else {
		// Write to stderr to keep stdout clean for actual output
		fmt.Fprintf(os.Stderr, "   %s (%.1fs)\n", msg, elapsed.Seconds())
	}
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
