// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package reembedding

import (
	"testing"
	"time"
)

func TestNewProgressTracker(t *testing.T) {
	total := int64(1000)
	tracker := NewProgressTracker(total)

	if tracker == nil {
		t.Fatal("Expected non-nil tracker")
	}

	if tracker.total != total {
		t.Errorf("Expected total %d, got %d", total, tracker.total)
	}

	if tracker.processed != 0 {
		t.Errorf("Expected processed 0, got %d", tracker.processed)
	}

	if tracker.startTime.IsZero() {
		t.Error("Expected non-zero start time")
	}
}

func TestProgressTracker_Update(t *testing.T) {
	tracker := NewProgressTracker(100)

	tracker.Update(10)
	if tracker.processed != 10 {
		t.Errorf("Expected processed 10, got %d", tracker.processed)
	}

	tracker.Update(20)
	if tracker.processed != 30 {
		t.Errorf("Expected processed 30, got %d", tracker.processed)
	}

	tracker.Update(70)
	if tracker.processed != 100 {
		t.Errorf("Expected processed 100, got %d", tracker.processed)
	}
}

func TestProgressTracker_GetPercentage(t *testing.T) {
	tests := []struct {
		name      string
		total     int64
		processed int64
		expected  float64
	}{
		{"0%", 100, 0, 0.0},
		{"25%", 100, 25, 25.0},
		{"50%", 100, 50, 50.0},
		{"75%", 100, 75, 75.0},
		{"100%", 100, 100, 100.0},
		{"empty collection", 0, 0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewProgressTracker(tt.total)
			tracker.processed = tt.processed

			percentage := tracker.GetPercentage()
			if percentage != tt.expected {
				t.Errorf("Expected %.1f%%, got %.1f%%", tt.expected, percentage)
			}
		})
	}
}

func TestProgressTracker_IsComplete(t *testing.T) {
	tracker := NewProgressTracker(100)

	if tracker.IsComplete() {
		t.Error("Expected incomplete at start")
	}

	tracker.Update(50)
	if tracker.IsComplete() {
		t.Error("Expected incomplete at 50%")
	}

	tracker.Update(50)
	if !tracker.IsComplete() {
		t.Error("Expected complete at 100%")
	}

	tracker.Update(10) // Beyond 100%
	if !tracker.IsComplete() {
		t.Error("Expected complete beyond 100%")
	}
}

func TestProgressTracker_GetSpeed(t *testing.T) {
	tracker := NewProgressTracker(1000)

	// Simulate processing with time delay
	tracker.startTime = time.Now().Add(-1 * time.Minute)
	tracker.processed = 100

	speed := tracker.GetSpeed()
	// Should be ~100 docs/min
	if speed < 90 || speed > 110 {
		t.Errorf("Expected speed ~100 docs/min, got %.2f", speed)
	}
}

func TestProgressTracker_GetETA(t *testing.T) {
	tracker := NewProgressTracker(1000)

	// Simulate 1 minute elapsed, 100 docs processed
	tracker.startTime = time.Now().Add(-1 * time.Minute)
	tracker.processed = 100

	eta := tracker.GetETA()

	// Should be ~9 minutes (900 docs remaining at 100 docs/min)
	expectedMin := 8 * time.Minute
	expectedMax := 10 * time.Minute

	if eta < expectedMin || eta > expectedMax {
		t.Errorf("Expected ETA ~9 minutes, got %v", eta)
	}
}

func TestProgressTracker_GetETA_ZeroProcessed(t *testing.T) {
	tracker := NewProgressTracker(1000)

	eta := tracker.GetETA()
	if eta != 0 {
		t.Errorf("Expected ETA 0 for zero processed, got %v", eta)
	}
}

func TestProgressTracker_GetETA_Complete(t *testing.T) {
	tracker := NewProgressTracker(100)
	tracker.processed = 100

	eta := tracker.GetETA()
	if eta != 0 {
		t.Errorf("Expected ETA 0 when complete, got %v", eta)
	}
}

func TestProgressTracker_formatDuration(t *testing.T) {
	tracker := NewProgressTracker(100)

	tests := []struct {
		duration time.Duration
		expected string
	}{
		{0, "0s"},
		{5 * time.Second, "5s"},
		{30 * time.Second, "30s"},
		{1 * time.Minute, "1m 0s"},
		{1*time.Minute + 30*time.Second, "1m 30s"},
		{1 * time.Hour, "1h 0m 0s"},
		{1*time.Hour + 30*time.Minute, "1h 30m 0s"},
		{2*time.Hour + 15*time.Minute + 45*time.Second, "2h 15m 45s"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tracker.formatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("formatDuration(%v) = %s, want %s", tt.duration, result, tt.expected)
			}
		})
	}
}

func TestProgressTracker_Display(t *testing.T) {
	// This test just ensures Display doesn't panic
	tracker := NewProgressTracker(100)

	tracker.Display() // 0%

	tracker.Update(25)
	tracker.Display() // 25%

	tracker.Update(75)
	tracker.Display() // 100%

	// Empty collection
	emptyTracker := NewProgressTracker(0)
	emptyTracker.Display() // Should not panic
}
