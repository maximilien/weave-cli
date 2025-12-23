// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package document

import (
	"testing"
	"time"
)

func TestDetermineExitCode(t *testing.T) {
	tests := []struct {
		name           string
		progress       *BatchProgress
		totalScanned   int
		expectedCode   int
		expectedReason string
	}{
		{
			name: "all success",
			progress: &BatchProgress{
				ProcessedFiles: 10,
				SuccessFiles:   10,
				FailedFiles:    0,
			},
			totalScanned:   10,
			expectedCode:   ExitSuccess,
			expectedReason: "all files processed successfully",
		},
		{
			name: "no files processed",
			progress: &BatchProgress{
				ProcessedFiles: 0,
				SuccessFiles:   0,
				FailedFiles:    0,
			},
			totalScanned:   10,
			expectedCode:   ExitCompleteFailure,
			expectedReason: "no files processed",
		},
		{
			name: "partial failure - 20% failed",
			progress: &BatchProgress{
				ProcessedFiles: 10,
				SuccessFiles:   8,
				FailedFiles:    2,
			},
			totalScanned:   10,
			expectedCode:   ExitPartialFailure,
			expectedReason: "20% failure rate",
		},
		{
			name: "complete failure - 60% failed",
			progress: &BatchProgress{
				ProcessedFiles: 10,
				SuccessFiles:   4,
				FailedFiles:    6,
			},
			totalScanned:   10,
			expectedCode:   ExitCompleteFailure,
			expectedReason: "60% failure rate (>50%)",
		},
		{
			name: "exactly 50% failed - should be partial",
			progress: &BatchProgress{
				ProcessedFiles: 10,
				SuccessFiles:   5,
				FailedFiles:    5,
			},
			totalScanned:   10,
			expectedCode:   ExitPartialFailure,
			expectedReason: "50% failure rate (=50%)",
		},
		{
			name: "with skipped files",
			progress: &BatchProgress{
				ProcessedFiles: 5,
				SuccessFiles:   5,
				FailedFiles:    0,
				SkippedFiles:   5,
			},
			totalScanned:   10,
			expectedCode:   ExitSuccess,
			expectedReason: "all processed files succeeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := DetermineExitCode(tt.progress, tt.totalScanned)
			if code != tt.expectedCode {
				t.Errorf("DetermineExitCode() = %d, want %d (reason: %s)",
					code, tt.expectedCode, tt.expectedReason)
			}
		})
	}
}

func TestBuildBatchReport(t *testing.T) {
	progress := &BatchProgress{
		TotalFiles:     10,
		ProcessedFiles: 8,
		SuccessFiles:   7,
		FailedFiles:    1,
		SkippedFiles:   2,
		TotalChunks:    100,
		TotalImages:    5,
		StartTime:      time.Now().Add(-5 * time.Minute),
	}

	statuses := []ProcessedFileStatus{
		{
			FilePath:         "/path/file1.pdf",
			Success:          true,
			TextChunks:       10,
			Images:           1,
			FileSize:         1024,
			ProcessingTimeMs: 1000,
		},
		{
			FilePath:         "/path/file2.pdf",
			Success:          false,
			Error:            "processing failed",
			FileSize:         2048,
			ProcessingTimeMs: 2000,
		},
	}

	report := BuildBatchReport(progress, statuses, 10, "test-collection", "qdrant-cloud", 100, 3)

	// Test configuration structure
	if report.Configuration.Collection != "test-collection" {
		t.Errorf("report.Configuration.Collection = %s, want test-collection", report.Configuration.Collection)
	}

	if report.Configuration.VDBType != "qdrant-cloud" {
		t.Errorf("report.Configuration.VDBType = %s, want qdrant-cloud", report.Configuration.VDBType)
	}

	// Test file counts
	if report.Files.Scanned != 10 {
		t.Errorf("report.Files.Scanned = %d, want 10", report.Files.Scanned)
	}

	if report.Files.Processed != 8 {
		t.Errorf("report.Files.Processed = %d, want 8", report.Files.Processed)
	}

	if report.Files.Failed != 1 {
		t.Errorf("report.Files.Failed = %d, want 1", report.Files.Failed)
	}

	if report.Files.Skipped != 2 {
		t.Errorf("report.Files.Skipped = %d, want 2", report.Files.Skipped)
	}

	// Test document counts (Created = TotalChunks from progress)
	// Note: Images are tracked separately, not included in Documents.Created
	if report.Documents.Created != progress.TotalChunks {
		t.Errorf("report.Documents.Created = %d, want %d", report.Documents.Created, progress.TotalChunks)
	}

	// Test configuration
	if report.Configuration.BatchSize != 100 {
		t.Errorf("report.Configuration.BatchSize = %d, want 100", report.Configuration.BatchSize)
	}

	if report.Configuration.Workers != 3 {
		t.Errorf("report.Configuration.Workers = %d, want 3", report.Configuration.Workers)
	}

	// Test exit code
	exitCode := DetermineExitCode(progress, 10)
	if report.ExitCode != exitCode {
		t.Errorf("report.ExitCode = %d, want %d", report.ExitCode, exitCode)
	}

	// Test status determination
	if report.Status != "partial" {
		t.Errorf("report.Status = %s, want partial", report.Status)
	}
}

func TestParseSinceDuration(t *testing.T) {
	tests := []struct {
		name        string
		since       string
		expectError bool
		expected    time.Duration
	}{
		{
			name:        "hours",
			since:       "24h",
			expectError: false,
			expected:    24 * time.Hour,
		},
		{
			name:        "days",
			since:       "7d",
			expectError: false,
			expected:    7 * 24 * time.Hour,
		},
		{
			name:        "weeks",
			since:       "2w",
			expectError: false,
			expected:    2 * 7 * 24 * time.Hour,
		},
		{
			name:        "minutes",
			since:       "30m",
			expectError: false,
			expected:    30 * time.Minute,
		},
		{
			name:        "invalid format",
			since:       "invalid",
			expectError: true,
		},
		{
			name:        "empty string",
			since:       "",
			expectError: false,
			expected:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration, err := parseSinceDuration(tt.since)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if duration != tt.expected {
				t.Errorf("parseSinceDuration(%s) = %v, want %v", tt.since, duration, tt.expected)
			}
		})
	}
}

func TestBatchProgressCalculations(t *testing.T) {
	progress := &BatchProgress{
		TotalFiles:     100,
		ProcessedFiles: 50,
		SuccessFiles:   45,
		FailedFiles:    5,
		SkippedFiles:   0,
		TotalChunks:    500,
		TotalImages:    10,
		StartTime:      time.Now().Add(-10 * time.Minute),
	}

	// Test failure rate calculation (implicit in exit code)
	exitCode := DetermineExitCode(progress, 100)
	if exitCode != ExitPartialFailure {
		t.Errorf("With 10%% failure rate, expected ExitPartialFailure, got %d", exitCode)
	}

	// Test with high failure rate
	progress.SuccessFiles = 20
	progress.FailedFiles = 30
	exitCode = DetermineExitCode(progress, 100)
	if exitCode != ExitCompleteFailure {
		t.Errorf("With 60%% failure rate, expected ExitCompleteFailure, got %d", exitCode)
	}
}

func TestExitCodeConstants(t *testing.T) {
	// Verify exit code values match expectations
	if ExitSuccess != 0 {
		t.Errorf("ExitSuccess = %d, want 0", ExitSuccess)
	}
	if ExitPartialFailure != 1 {
		t.Errorf("ExitPartialFailure = %d, want 1", ExitPartialFailure)
	}
	if ExitCompleteFailure != 2 {
		t.Errorf("ExitCompleteFailure = %d, want 2", ExitCompleteFailure)
	}
}
