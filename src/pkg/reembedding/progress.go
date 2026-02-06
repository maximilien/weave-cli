// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package reembedding

import (
	"fmt"
	"time"
)

// ProgressTracker tracks and displays progress for re-embedding operations
type ProgressTracker struct {
	total      int64
	processed  int64
	startTime  time.Time
	lastUpdate time.Time
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(total int64) *ProgressTracker {
	now := time.Now()
	return &ProgressTracker{
		total:      total,
		processed:  0,
		startTime:  now,
		lastUpdate: now,
	}
}

// Update increments the progress counter
func (p *ProgressTracker) Update(count int) {
	p.processed += int64(count)
	p.lastUpdate = time.Now()
}

// Display shows the current progress with a progress bar and ETA
func (p *ProgressTracker) Display() {
	if p.total == 0 {
		return
	}

	percentage := float64(p.processed) / float64(p.total) * 100
	speed := p.GetSpeed()
	eta := p.GetETA()

	// Create progress bar
	barWidth := 40
	filledWidth := int(float64(barWidth) * percentage / 100)

	bar := "["
	for i := 0; i < barWidth; i++ {
		if i < filledWidth {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	bar += "]"

	// Format output
	statusMsg := fmt.Sprintf("%s %.0f%% (%d/%d docs) | Speed: %.0f docs/min | ETA: %s",
		bar, percentage, p.processed, p.total, speed, p.formatDuration(eta))

	fmt.Println("📊 " + statusMsg)
}

// GetStatusMessage returns formatted progress message without printing
func (p *ProgressTracker) GetStatusMessage() string {
	if p.total == 0 {
		return ""
	}

	percentage := float64(p.processed) / float64(p.total) * 100
	speed := p.GetSpeed()
	eta := p.GetETA()

	// Create progress bar
	barWidth := 40
	filledWidth := int(float64(barWidth) * percentage / 100)

	bar := "["
	for i := 0; i < barWidth; i++ {
		if i < filledWidth {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	bar += "]"

	// Format output
	return fmt.Sprintf("%s %.0f%% (%d/%d docs) | Speed: %.0f docs/min | ETA: %s",
		bar, percentage, p.processed, p.total, speed, p.formatDuration(eta))
}

// GetSpeed returns the current processing speed in documents per minute
func (p *ProgressTracker) GetSpeed() float64 {
	elapsed := time.Since(p.startTime).Seconds()
	if elapsed == 0 {
		return 0
	}
	return float64(p.processed) / elapsed * 60
}

// GetETA returns the estimated time remaining
func (p *ProgressTracker) GetETA() time.Duration {
	if p.processed == 0 {
		return 0
	}

	elapsed := time.Since(p.startTime)
	remaining := p.total - p.processed

	if remaining <= 0 {
		return 0
	}

	avgTimePerDoc := elapsed / time.Duration(p.processed)
	return avgTimePerDoc * time.Duration(remaining)
}

// IsComplete returns true if all documents have been processed
func (p *ProgressTracker) IsComplete() bool {
	return p.processed >= p.total
}

// GetPercentage returns the completion percentage
func (p *ProgressTracker) GetPercentage() float64 {
	if p.total == 0 {
		return 0
	}
	return float64(p.processed) / float64(p.total) * 100
}

// formatDuration formats a duration into a human-readable string
func (p *ProgressTracker) formatDuration(d time.Duration) string {
	if d == 0 {
		return "0s"
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}
