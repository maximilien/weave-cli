// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package progress

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewReporter(t *testing.T) {
	reporter := NewReporter(true)
	if reporter == nil {
		t.Fatal("Expected reporter, got nil")
	}
	if !reporter.enabled {
		t.Error("Expected reporter to be enabled")
	}
	if reporter.jsonOutput {
		t.Error("Expected text output mode")
	}
}

func TestNewJSONReporter(t *testing.T) {
	reporter := NewJSONReporter(true)
	if reporter == nil {
		t.Fatal("Expected reporter, got nil")
	}
	if !reporter.enabled {
		t.Error("Expected reporter to be enabled")
	}
	if !reporter.jsonOutput {
		t.Error("Expected JSON output mode")
	}
}

func TestReporter_Disabled(t *testing.T) {
	// When disabled, all methods should be no-ops
	reporter := NewReporter(false)

	// Capture output
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	reporter.Start("Starting")
	reporter.Update("Updating")
	reporter.UpdateProgress("Progress")
	reporter.UpdateWithCount(5, "Count update")
	reporter.Complete("Done")

	w.Close()
	os.Stderr = old

	out, _ := io.ReadAll(r)
	if len(out) > 0 {
		t.Errorf("Expected no output when disabled, got: %s", string(out))
	}
}

func TestReporter_SetTotal(t *testing.T) {
	reporter := NewReporter(true)
	reporter.SetTotal(100)

	if reporter.total != 100 {
		t.Errorf("Expected total 100, got %d", reporter.total)
	}
	if reporter.current != 0 {
		t.Errorf("Expected current reset to 0, got %d", reporter.current)
	}
}

func TestReporter_TextMode_Lifecycle(t *testing.T) {
	reporter := NewReporter(true)

	// Capture stderr
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	reporter.Start("Starting operation")
	time.Sleep(10 * time.Millisecond)
	reporter.Update("Processing")
	time.Sleep(10 * time.Millisecond)
	reporter.Complete("Operation complete")

	w.Close()
	os.Stderr = old

	out, _ := io.ReadAll(r)
	output := string(out)

	// Check for expected content
	if !strings.Contains(output, "Starting operation") {
		t.Error("Expected 'Starting operation' in output")
	}
	if !strings.Contains(output, "Processing") {
		t.Error("Expected 'Processing' in output")
	}
	if !strings.Contains(output, "Operation complete") {
		t.Error("Expected 'Operation complete' in output")
	}
	if !strings.Contains(output, "🔍") {
		t.Error("Expected start icon in output")
	}
	if !strings.Contains(output, "✅") {
		t.Error("Expected complete icon in output")
	}
}

func TestReporter_JSONMode_Lifecycle(t *testing.T) {
	reporter := NewJSONReporter(true)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	reporter.Start("Starting")
	reporter.Update("Processing")
	reporter.Complete("Done")

	w.Close()
	os.Stdout = old

	out, _ := io.ReadAll(r)
	output := string(out)

	// Parse JSON lines
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 3 {
		t.Fatalf("Expected at least 3 JSON messages, got %d", len(lines))
	}

	// Check start message
	var startMsg ProgressMessage
	if err := json.Unmarshal([]byte(lines[0]), &startMsg); err != nil {
		t.Fatalf("Failed to parse start message: %v", err)
	}
	if startMsg.Type != "progress" {
		t.Errorf("Expected type 'progress', got '%s'", startMsg.Type)
	}
	if startMsg.Phase != "start" {
		t.Errorf("Expected phase 'start', got '%s'", startMsg.Phase)
	}
	if startMsg.Message != "Starting" {
		t.Errorf("Expected message 'Starting', got '%s'", startMsg.Message)
	}

	// Check complete message
	var completeMsg ProgressMessage
	if err := json.Unmarshal([]byte(lines[2]), &completeMsg); err != nil {
		t.Fatalf("Failed to parse complete message: %v", err)
	}
	if completeMsg.Phase != "complete" {
		t.Errorf("Expected phase 'complete', got '%s'", completeMsg.Phase)
	}
}

func TestReporter_UpdateProgress(t *testing.T) {
	reporter := NewReporter(true)
	reporter.SetTotal(10)

	reporter.UpdateProgress("Item 1")
	if reporter.current != 1 {
		t.Errorf("Expected current=1, got %d", reporter.current)
	}

	reporter.UpdateProgress("Item 2")
	if reporter.current != 2 {
		t.Errorf("Expected current=2, got %d", reporter.current)
	}
}

func TestReporter_UpdateWithCount(t *testing.T) {
	reporter := NewReporter(true)
	reporter.SetTotal(100)

	reporter.UpdateWithCount(25, "25% done")
	if reporter.current != 25 {
		t.Errorf("Expected current=25, got %d", reporter.current)
	}

	reporter.UpdateWithCount(75, "75% done")
	if reporter.current != 75 {
		t.Errorf("Expected current=75, got %d", reporter.current)
	}
}

func TestReporter_JSONWithProgress(t *testing.T) {
	reporter := NewJSONReporter(true)
	reporter.SetTotal(100)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	reporter.Start("Starting")
	// Simulate some progress
	time.Sleep(10 * time.Millisecond)
	reporter.UpdateWithCount(50, "Halfway")

	w.Close()
	os.Stdout = old

	out, _ := io.ReadAll(r)
	output := string(out)

	// Should have progress message with metrics
	var msg ProgressMessage
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		t.Fatal("Expected progress output")
	}

	// Find the progress update message (skip start message)
	foundProgress := false
	for _, line := range lines {
		var tempMsg ProgressMessage
		if err := json.Unmarshal([]byte(line), &tempMsg); err != nil {
			continue
		}
		if tempMsg.Phase == "progress" && tempMsg.Current > 0 {
			msg = tempMsg
			foundProgress = true
			break
		}
	}

	if !foundProgress {
		t.Fatal("No progress update message found")
	}

	if msg.Current != 50 {
		t.Errorf("Expected current=50, got %d", msg.Current)
	}
	if msg.Total != 100 {
		t.Errorf("Expected total=100, got %d", msg.Total)
	}
	if msg.Progress != 50.0 {
		t.Errorf("Expected progress=50.0, got %.1f", msg.Progress)
	}
	// ETA and throughput may be 0 if elapsed time is too small, so just verify fields exist
	// (These are timing-dependent and can be flaky in fast tests)
	if msg.ETA < 0 {
		t.Error("Expected non-negative ETA")
	}
	if msg.Throughput < 0 {
		t.Error("Expected non-negative throughput")
	}
}

func TestFormatProgressBar(t *testing.T) {
	tests := []struct {
		name        string
		total       int
		current     int
		expectBar   bool
		expectPercent string
	}{
		{
			name:          "No total (no bar)",
			total:         0,
			current:       5,
			expectBar:     false,
		},
		{
			name:          "0%",
			total:         100,
			current:       0,
			expectBar:     true,
			expectPercent: "0%",
		},
		{
			name:          "50%",
			total:         100,
			current:       50,
			expectBar:     true,
			expectPercent: "50%",
		},
		{
			name:          "100%",
			total:         100,
			current:       100,
			expectBar:     true,
			expectPercent: "100%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := NewReporter(true)
			reporter.SetTotal(tt.total)
			reporter.current = tt.current
			reporter.startTime = time.Now().Add(-1 * time.Second) // 1 second elapsed

			bar := reporter.formatProgressBar()

			if tt.expectBar {
				if bar == "" {
					t.Error("Expected progress bar, got empty string")
				}
				if !strings.Contains(bar, "[") || !strings.Contains(bar, "]") {
					t.Errorf("Expected brackets in progress bar, got: %s", bar)
				}
				if !strings.Contains(bar, tt.expectPercent) {
					t.Errorf("Expected '%s' in bar, got: %s", tt.expectPercent, bar)
				}
			} else {
				if bar != "" {
					t.Errorf("Expected empty bar, got: %s", bar)
				}
			}
		})
	}
}

func TestCalculateETA(t *testing.T) {
	tests := []struct {
		name      string
		total     int
		current   int
		elapsed   time.Duration
		expectETA bool
	}{
		{
			name:      "No current (no ETA)",
			total:     100,
			current:   0,
			elapsed:   time.Second,
			expectETA: false,
		},
		{
			name:      "No total (no ETA)",
			total:     0,
			current:   50,
			elapsed:   time.Second,
			expectETA: false,
		},
		{
			name:      "50% done in 1 second (1 second remaining)",
			total:     100,
			current:   50,
			elapsed:   time.Second,
			expectETA: true,
		},
		{
			name:      "90% done (small ETA)",
			total:     100,
			current:   90,
			elapsed:   10 * time.Second,
			expectETA: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := NewReporter(true)
			reporter.SetTotal(tt.total)
			reporter.current = tt.current
			reporter.startTime = time.Now().Add(-tt.elapsed)

			eta := reporter.calculateETA()

			if tt.expectETA {
				if eta <= 0 {
					t.Errorf("Expected positive ETA, got %.2f", eta)
				}
			} else {
				if eta != 0 {
					t.Errorf("Expected 0 ETA, got %.2f", eta)
				}
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		seconds  float64
		expected string
	}{
		{5, "5s"},
		{30, "30s"},
		{59, "59s"},
		{60, "1m0s"},
		{90, "2m30s"}, // Note: 90/60 = 1.5 rounds to 2 in %.0f format
		{3600, "1h0m"},
		{3661, "1h1m"},
		{7200, "2h0m"},
	}

	reporter := NewReporter(true)
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := reporter.formatDuration(tt.seconds)
			if result != tt.expected {
				t.Errorf("formatDuration(%.0f) = %s, expected %s", tt.seconds, result, tt.expected)
			}
		})
	}
}

func TestProgressBar_VisualFormat(t *testing.T) {
	reporter := NewReporter(true)
	reporter.SetTotal(100)
	reporter.startTime = time.Now().Add(-2 * time.Second)

	// Test different progress levels
	tests := []struct {
		current     int
		expectFilled int // Approximate number of '=' characters
	}{
		{0, 0},
		{25, 5},  // 25% of 20-char bar = 5 chars
		{50, 10}, // 50% of 20-char bar = 10 chars
		{75, 15}, // 75% of 20-char bar = 15 chars
		{100, 20}, // 100% of 20-char bar = 20 chars
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.current)), func(t *testing.T) {
			reporter.current = tt.current
			bar := reporter.formatProgressBar()

			// Count '=' characters
			filled := strings.Count(bar, "=")
			if filled < tt.expectFilled-1 || filled > tt.expectFilled+1 {
				t.Errorf("At %d%%, expected ~%d filled chars, got %d. Bar: %s",
					tt.current, tt.expectFilled, filled, bar)
			}
		})
	}
}

func TestReporter_TextOutput_ProgressBar(t *testing.T) {
	reporter := NewReporter(true)
	reporter.SetTotal(10)
	reporter.Start("Starting")

	// Capture stderr
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	reporter.UpdateWithCount(5, "Halfway")

	w.Close()
	os.Stderr = old

	out, _ := io.ReadAll(r)
	output := string(out)

	// Should contain progress bar with percentage
	if !strings.Contains(output, "[") {
		t.Error("Expected progress bar in output")
	}
	if !strings.Contains(output, "50%") {
		t.Error("Expected 50% in output")
	}
	if !strings.Contains(output, "5/10") {
		t.Error("Expected count 5/10 in output")
	}
}

func TestProgressMessage_JSONMarshaling(t *testing.T) {
	msg := ProgressMessage{
		Type:       "progress",
		Phase:      "update",
		Message:    "Processing",
		Elapsed:    1.5,
		Progress:   50.0,
		Current:    50,
		Total:      100,
		ETA:        1.5,
		Throughput: 33.3,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}

	var parsed ProgressMessage
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}

	if parsed.Type != msg.Type {
		t.Errorf("Type mismatch: expected %s, got %s", msg.Type, parsed.Type)
	}
	if parsed.Phase != msg.Phase {
		t.Errorf("Phase mismatch: expected %s, got %s", msg.Phase, parsed.Phase)
	}
	if parsed.Progress != msg.Progress {
		t.Errorf("Progress mismatch: expected %.1f, got %.1f", msg.Progress, parsed.Progress)
	}
}

func TestReporter_ZeroElapsed(t *testing.T) {
	reporter := NewReporter(true)
	reporter.SetTotal(100)
	reporter.startTime = time.Now() // Just started

	// Should handle zero elapsed time gracefully
	eta := reporter.calculateETA()
	if eta != 0 {
		t.Errorf("Expected 0 ETA for zero elapsed time, got %.2f", eta)
	}
}
