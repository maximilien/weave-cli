// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package document

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
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

	// Test document counts (Created = TotalChunks + TotalImages from progress)
	// Note: Images are now included in Documents.Created for accurate total count
	expectedCreated := progress.TotalChunks + progress.TotalImages
	if report.Documents.Created != expectedCreated {
		t.Errorf("report.Documents.Created = %d, want %d", report.Documents.Created, expectedCreated)
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

func TestBatchFileDiscoveryAndFiltering(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "nested")
	if err := os.Mkdir(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	paths := []string{
		filepath.Join(root, "document.txt"), filepath.Join(root, "notes.MD"),
		filepath.Join(nested, "data.json"), filepath.Join(root, "image.PNG"),
		filepath.Join(root, "ignored.bin"),
	}
	for _, path := range paths {
		if err := os.WriteFile(path, []byte("content"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	files, err := scanDirectory(root)
	if err != nil || len(files) != 4 {
		t.Fatalf("scanDirectory() = %#v, %v", files, err)
	}
	if _, err := scanDirectory(filepath.Join(root, "missing")); err == nil {
		t.Fatal("scanDirectory() accepted a missing directory")
	}

	successPath := paths[0] + ".processed"
	if err := saveProcessedStatus(successPath, &ProcessedFileStatus{Success: true}); err != nil {
		t.Fatal(err)
	}
	failedPath := paths[1] + ".processed"
	if err := saveProcessedStatus(failedPath, &ProcessedFileStatus{Success: false}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(paths[2]+".processed", []byte("invalid"), 0o600); err != nil {
		t.Fatal(err)
	}
	unprocessed := filterProcessedFiles(paths[:4])
	if len(unprocessed) != 3 {
		t.Fatalf("filterProcessedFiles() = %#v", unprocessed)
	}

	old := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(paths[1], old, old); err != nil {
		t.Fatal(err)
	}
	if got := filterFilesBySince(paths[:2], 24*time.Hour); len(got) != 1 || got[0] != paths[0] {
		t.Fatalf("filterFilesBySince() = %#v", got)
	}
	withMissing := append(paths[:2], filepath.Join(root, "missing.txt"))
	if got := filterFilesBySince(withMissing, 0); len(got) != 3 {
		t.Fatalf("zero-duration filter = %#v", got)
	}
	if got := filterFilesBySince(withMissing, time.Hour); len(got) != 1 {
		t.Fatalf("missing-file filter = %#v", got)
	}
}

func TestBatchStatusPersistenceAndReport(t *testing.T) {
	root := t.TempDir()
	statusPath := filepath.Join(root, "doc.processed")
	status := &ProcessedFileStatus{
		FilePath: "doc.pdf", Success: true, TextChunks: 3, Images: 2,
		ProcessedAt: time.Date(2026, 9, 11, 10, 0, 0, 0, time.UTC), ProcessingTimeMs: 1250,
	}
	if err := saveProcessedStatus(statusPath, status); err != nil {
		t.Fatalf("saveProcessedStatus(): %v", err)
	}
	loaded, err := loadProcessedStatus(statusPath)
	if err != nil || !loaded.Success || loaded.TextChunks != 3 {
		t.Fatalf("loadProcessedStatus() = %#v, %v", loaded, err)
	}
	if _, err := loadProcessedStatus(filepath.Join(root, "missing")); err == nil {
		t.Fatal("loadProcessedStatus() accepted a missing file")
	}
	invalid := filepath.Join(root, "invalid.processed")
	if err := os.WriteFile(invalid, []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadProcessedStatus(invalid); err == nil {
		t.Fatal("loadProcessedStatus() accepted invalid JSON")
	}

	reportPath := filepath.Join(root, "report.csv")
	statuses := []ProcessedFileStatus{*status, {FilePath: "bad.txt", Error: "failed", RetryCount: 2}}
	if err := writeBatchReport(reportPath, statuses); err != nil {
		t.Fatalf("writeBatchReport(): %v", err)
	}
	report, err := os.ReadFile(reportPath)
	if err != nil || !strings.Contains(string(report), "doc.pdf,success") || !strings.Contains(string(report), "bad.txt,failed") {
		t.Fatalf("batch report = %q, %v", report, err)
	}
	if err := writeBatchReport(filepath.Join(root, "missing", "report.csv"), statuses); err == nil {
		t.Fatal("writeBatchReport() accepted a missing parent")
	}
}

func TestBatchProgressAndPresentationHelpers(t *testing.T) {
	progress := &BatchProgress{TotalFiles: 2, StartTime: time.Now().Add(-time.Second)}
	mu := &sync.Mutex{}
	success := &ProcessedFileStatus{Success: true, TextChunks: 3, Images: 2, ChunksFailed: 1, ProcessingTimeMs: 1250}
	updateProgress(progress, success, mu)
	if progress.ProcessedFiles != 1 || progress.SuccessFiles != 1 || progress.TotalChunks != 3 || progress.TotalImages != 2 {
		t.Fatalf("success progress = %#v", progress)
	}
	failure := &ProcessedFileStatus{Error: "boom", RetryCount: 2, ImagesFailed: 1}
	updateProgress(progress, failure, mu)
	if progress.ProcessedFiles != 2 || progress.FailedFiles != 1 || progress.FailedChunks != 1 || progress.FailedImages != 1 {
		t.Fatalf("failure progress = %#v", progress)
	}

	for path, want := range map[string]string{"doc.PDF": "pdf", "image.JPEG": "image", "image.webp": "image", "doc.txt": "text"} {
		if got := getFileType(path); got != want {
			t.Errorf("getFileType(%q) = %q, want %q", path, got, want)
		}
	}
	for duration, want := range map[time.Duration]string{30 * time.Second: "30s", 90 * time.Second: "1m 30s", 90 * time.Minute: "1h 30m"} {
		if got := formatDuration(duration); got != want {
			t.Errorf("formatDuration(%s) = %q, want %q", duration, got, want)
		}
	}

	printFileSuccess("doc.pdf", success, progress)
	printFileSuccess("image.png", &ProcessedFileStatus{Success: true}, progress)
	printFileSuccess("notes.txt", &ProcessedFileStatus{Success: true}, &BatchProgress{ProcessedFiles: 1, TotalFiles: 1})
	printFileFailure("bad.txt", failure, progress)
	printFileFailure("bad.txt", &ProcessedFileStatus{Error: "boom"}, progress)
	printBatchConfiguration("/tmp/docs", "Docs", 2, 1, 1000, "Images", true, 1024, 10, "report.csv")
	printBatchConfiguration("/tmp/docs", "Docs", 1, 0, 500, "", false, 0, 1, "")
	printBatchSummary(progress, []ProcessedFileStatus{{FilePath: "ok.txt", Success: true}, {FilePath: "bad.txt", Error: "boom"}})
	printBatchSummary(&BatchProgress{TotalFiles: 1, StartTime: time.Now()}, nil)
}
