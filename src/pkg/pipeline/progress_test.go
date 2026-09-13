// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package pipeline

import (
	"io"
	"testing"
)

func TestProgressTrackerEnabledLifecycle(t *testing.T) {
	tracker := NewProgressTracker(true, false)
	tracker.StartScanning()
	tracker.FinishScanning(3)
	tracker.StartProcessing(2, 1, 2)
	tracker.UpdateProgress(1)
	tracker.FinishProcessing()
	tracker.StartBatching(3, 2)
	tracker.UpdateBatch(1, 2, 2)
	tracker.FinishBatching()
	tracker.ShowError("error")
	tracker.ShowWarning("warning")
	tracker.ShowInfo("info")
	if tracker.GetWriter() == io.Discard {
		t.Fatal("enabled tracker returned io.Discard")
	}
}

func TestProgressTrackerDisabledAndQuietPaths(t *testing.T) {
	tracker := NewProgressTracker(false, false)
	tracker.StartScanning()
	tracker.FinishScanning(0)
	tracker.StartProcessing(1, 1, 1)
	tracker.UpdateProgress(1)
	tracker.FinishProcessing()
	tracker.StartBatching(1, 1)
	tracker.UpdateBatch(1, 1, 1)
	tracker.FinishBatching()
	if tracker.GetWriter() != io.Discard {
		t.Fatal("disabled tracker returned an output writer")
	}

	quiet := NewProgressTracker(true, true)
	quiet.ShowError("hidden")
	quiet.ShowWarning("hidden")
	quiet.ShowInfo("hidden")
	if quiet.GetWriter() != io.Discard {
		t.Fatal("quiet tracker returned an output writer")
	}
}
