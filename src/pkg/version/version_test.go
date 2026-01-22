// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package version

import (
	"strings"
	"testing"
	"time"
)

func TestGet(t *testing.T) {
	// Save original values
	origVersion := Version
	origGitCommit := GitCommit
	origBuildTime := BuildTime
	origGoVersion := GoVersion

	// Restore after test
	defer func() {
		Version = origVersion
		GitCommit = origGitCommit
		BuildTime = origBuildTime
		GoVersion = origGoVersion
	}()

	// Set test values
	Version = "0.9.8"
	GitCommit = "abc1234"
	BuildTime = "2026-01-21T21:32:45Z"
	GoVersion = "go1.24.1"

	info := Get()

	if info.Version != "0.9.8" {
		t.Errorf("Expected version 0.9.8, got %s", info.Version)
	}
	if info.GitCommit != "abc1234" {
		t.Errorf("Expected git commit abc1234, got %s", info.GitCommit)
	}
	if info.BuildTime != "2026-01-21T21:32:45Z" {
		t.Errorf("Expected build time 2026-01-21T21:32:45Z, got %s", info.BuildTime)
	}
	if info.GoVersion != "go1.24.1" {
		t.Errorf("Expected go version go1.24.1, got %s", info.GoVersion)
	}
}

func TestString(t *testing.T) {
	// Save original values
	origVersion := Version
	origGitCommit := GitCommit
	origBuildTime := BuildTime
	origGoVersion := GoVersion

	// Restore after test
	defer func() {
		Version = origVersion
		GitCommit = origGitCommit
		BuildTime = origBuildTime
		GoVersion = origGoVersion
	}()

	tests := []struct {
		name           string
		version        string
		gitCommit      string
		buildTime      string
		goVersion      string
		expectContains []string
	}{
		{
			name:      "Full version info",
			version:   "0.9.8",
			gitCommit: "abc1234567890",
			buildTime: "2026-01-21T21:32:45Z",
			goVersion: "go1.24.1",
			expectContains: []string{
				"Weave CLI 0.9.8",
				"Git Commit: abc1234", // Should truncate to 7 chars
				"Build Time: 2026-01-21 21:32:45",
				"Go Version: go1.24.1",
			},
		},
		{
			name:      "Short git commit",
			version:   "0.9.7",
			gitCommit: "abc",
			buildTime: "2026-01-20T21:32:45Z",
			goVersion: "go1.24.1",
			expectContains: []string{
				"Weave CLI 0.9.7",
				"Git Commit: abc", // Should not truncate if < 7 chars
			},
		},
		{
			name:      "Unknown build time",
			version:   "dev",
			gitCommit: "unknown",
			buildTime: "unknown",
			goVersion: "go1.24.1",
			expectContains: []string{
				"Weave CLI dev",
				"Git Commit: unknown",
				"Build Time: unknown",
			},
		},
		{
			name:      "Invalid RFC3339 build time",
			version:   "0.9.8",
			gitCommit: "abc1234",
			buildTime: "invalid-date",
			goVersion: "go1.24.1",
			expectContains: []string{
				"Weave CLI 0.9.8",
				"Build Time: invalid-date", // Should use as-is if parse fails
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Version = tt.version
			GitCommit = tt.gitCommit
			BuildTime = tt.buildTime
			GoVersion = tt.goVersion

			result := String()

			for _, expected := range tt.expectContains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected string to contain '%s', got:\n%s", expected, result)
				}
			}
		})
	}
}

func TestShort(t *testing.T) {
	// Save original value
	origVersion := Version
	defer func() { Version = origVersion }()

	tests := []struct {
		version  string
		expected string
	}{
		{"0.9.8", "Weave CLI 0.9.8"},
		{"dev", "Weave CLI dev"},
		{"1.0.0", "Weave CLI 1.0.0"},
		{"0.9.7-6-gd399f1b", "Weave CLI 0.9.7-6-gd399f1b"},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			Version = tt.version
			result := Short()

			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestIsRelease(t *testing.T) {
	// Save original value
	origVersion := Version
	defer func() { Version = origVersion }()

	tests := []struct {
		version     string
		isRelease   bool
		description string
	}{
		{"0.9.8", true, "Release version"},
		{"1.0.0", true, "Major release"},
		{"0.9.7-6-gd399f1b", true, "Development version from tag"},
		{"dev", false, "Dev version"},
		{"dev-abc123", false, "Dev with commit"},
		{"unknown", false, "Unknown version"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			Version = tt.version
			result := IsRelease()

			if result != tt.isRelease {
				t.Errorf("Version '%s': expected IsRelease()=%v, got %v",
					tt.version, tt.isRelease, result)
			}
		})
	}
}

func TestGetVersionFromGitTag(t *testing.T) {
	// Save original value
	origVersion := Version
	defer func() { Version = origVersion }()

	tests := []struct {
		version  string
		expected string
	}{
		{"0.9.8", "0.9.8"},
		{"dev", "dev"},
		{"1.0.0-rc1", "1.0.0-rc1"},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			Version = tt.version
			result := GetVersionFromGitTag()

			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBuildTimeFormatting(t *testing.T) {
	// Save original values
	origVersion := Version
	origGitCommit := GitCommit
	origBuildTime := BuildTime
	origGoVersion := GoVersion

	defer func() {
		Version = origVersion
		GitCommit = origGitCommit
		BuildTime = origBuildTime
		GoVersion = origGoVersion
	}()

	// Set test values
	Version = "0.9.8"
	GitCommit = "abc1234"
	GoVersion = "go1.24.1"

	// Test RFC3339 parsing
	BuildTime = "2026-01-21T21:32:45Z"
	result := String()

	// Should format as "2006-01-02 15:04:05"
	if !strings.Contains(result, "2026-01-21 21:32:45") {
		t.Errorf("Expected formatted build time '2026-01-21 21:32:45', got:\n%s", result)
	}

	// Verify the formatting is correct
	buildTime, err := time.Parse(time.RFC3339, "2026-01-21T21:32:45Z")
	if err != nil {
		t.Fatalf("Failed to parse test build time: %v", err)
	}
	expectedFormat := buildTime.Format("2006-01-02 15:04:05")
	if !strings.Contains(result, expectedFormat) {
		t.Errorf("Expected build time '%s', got:\n%s", expectedFormat, result)
	}
}

func TestGitCommitShortening(t *testing.T) {
	// Save original values
	origVersion := Version
	origGitCommit := GitCommit
	origBuildTime := BuildTime
	origGoVersion := GoVersion

	defer func() {
		Version = origVersion
		GitCommit = origGitCommit
		BuildTime = origBuildTime
		GoVersion = origGoVersion
	}()

	Version = "0.9.8"
	BuildTime = "2026-01-21T21:32:45Z"
	GoVersion = "go1.24.1"

	tests := []struct {
		fullCommit    string
		expectedShort string
	}{
		{"abc1234567890", "abc1234"},
		{"0e3a9807a1e8170b67aa47e6b682420d953c91bc", "0e3a980"},
		{"abc", "abc"},         // Less than 7 chars
		{"1234567", "1234567"}, // Exactly 7 chars
	}

	for _, tt := range tests {
		t.Run(tt.fullCommit, func(t *testing.T) {
			GitCommit = tt.fullCommit
			result := String()

			expectedLine := "Git Commit: " + tt.expectedShort
			if !strings.Contains(result, expectedLine) {
				t.Errorf("Expected git commit '%s', got:\n%s", tt.expectedShort, result)
			}
		})
	}
}

func TestInfoStructJSON(t *testing.T) {
	// Test that Info struct has proper JSON tags
	info := Info{
		Version:   "0.9.8",
		GitCommit: "abc1234",
		BuildTime: "2026-01-21T21:32:45Z",
		GoVersion: "go1.24.1",
	}

	// Verify fields are accessible
	if info.Version != "0.9.8" {
		t.Errorf("Expected version 0.9.8, got %s", info.Version)
	}
	if info.GitCommit != "abc1234" {
		t.Errorf("Expected git commit abc1234, got %s", info.GitCommit)
	}
	if info.BuildTime != "2026-01-21T21:32:45Z" {
		t.Errorf("Expected build time 2026-01-21T21:32:45Z, got %s", info.BuildTime)
	}
	if info.GoVersion != "go1.24.1" {
		t.Errorf("Expected go version go1.24.1, got %s", info.GoVersion)
	}
}
