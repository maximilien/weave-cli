// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package collection

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// TestIssue2_FlagNamingConsistency tests that both hyphen and underscore
// versions of flags work correctly
// Regression test for AuctionsMax.ai Issue #2 (MEDIUM priority)
func TestIssue2_FlagNamingConsistency(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedTopK  int
		expectedImage int
		shouldFail    bool
	}{
		{
			name:          "Hyphen version (preferred)",
			args:          []string{"TestCol", "test query", "--top-k", "10", "--top-k-images", "3"},
			expectedTopK:  10,
			expectedImage: 3,
			shouldFail:    false,
		},
		{
			name:          "Underscore version (deprecated but supported)",
			args:          []string{"TestCol", "test query", "--top_k", "10", "--top_k_images", "3"},
			expectedTopK:  10,
			expectedImage: 3,
			shouldFail:    false,
		},
		{
			name:          "Mixed hyphen and underscore",
			args:          []string{"TestCol", "test query", "--top-k", "5", "--top_k_images", "2"},
			expectedTopK:  5,
			expectedImage: 2,
			shouldFail:    false,
		},
		{
			name:          "Short flag -k",
			args:          []string{"TestCol", "test query", "-k", "15"},
			expectedTopK:  15,
			expectedImage: 0, // default
			shouldFail:    false,
		},
		{
			name:          "Default values",
			args:          []string{"TestCol", "test query"},
			expectedTopK:  5, // default
			expectedImage: 0, // default
			shouldFail:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh command for each test
			cmd := &cobra.Command{
				Use: "query",
				Run: func(cmd *cobra.Command, args []string) {
					// Verify flag values
					topK, err := cmd.Flags().GetInt("top-k")
					if err != nil {
						if !tt.shouldFail {
							t.Errorf("Failed to get top-k flag: %v", err)
						}
						return
					}

					topKImages, err := cmd.Flags().GetInt("top-k-images")
					if err != nil {
						if !tt.shouldFail {
							t.Errorf("Failed to get top-k-images flag: %v", err)
						}
						return
					}

					if topK != tt.expectedTopK {
						t.Errorf("top-k: expected %d, got %d", tt.expectedTopK, topK)
					}

					if topKImages != tt.expectedImage {
						t.Errorf("top-k-images: expected %d, got %d", tt.expectedImage, topKImages)
					}
				},
			}

			// Initialize flags (same as QueryCmd init)
			cmd.Flags().IntP("top-k", "k", 5, "Number of top results to return (default: 5)")
			cmd.Flags().Int("top-k-images", 0, "Number of top results from image collections (0 = use top-k)")

			// Add normalization function (same as QueryCmd)
			cmd.Flags().SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
				switch name {
				case "top_k":
					name = "top-k"
				case "top_k_images":
					name = "top-k-images"
				}
				return pflag.NormalizedName(name)
			})

			// Set args and execute
			cmd.SetArgs(tt.args)
			err := cmd.Execute()

			if tt.shouldFail && err == nil {
				t.Error("Expected command to fail, but it succeeded")
			}
			if !tt.shouldFail && err != nil {
				t.Errorf("Expected command to succeed, but got error: %v", err)
			}
		})
	}
}

// TestIssue2_FlagNormalizationEdgeCases tests edge cases in flag normalization
func TestIssue2_FlagNormalizationEdgeCases(t *testing.T) {
	cmd := &cobra.Command{
		Use: "query",
		Run: func(cmd *cobra.Command, args []string) {},
	}

	cmd.Flags().IntP("top-k", "k", 5, "test")
	cmd.Flags().Int("top-k-images", 0, "test")

	cmd.Flags().SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		switch name {
		case "top_k":
			name = "top-k"
		case "top_k_images":
			name = "top-k-images"
		}
		return pflag.NormalizedName(name)
	})

	tests := []struct {
		name     string
		flagName string
		want     string
	}{
		{"Hyphen stays hyphen", "top-k", "top-k"},
		{"Underscore converts to hyphen", "top_k", "top-k"},
		{"Images hyphen stays hyphen", "top-k-images", "top-k-images"},
		{"Images underscore converts", "top_k_images", "top-k-images"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that the flag can be looked up with either version
			cmd.SetArgs([]string{"test", "query", "--" + tt.flagName, "10"})
			err := cmd.Execute()
			if err != nil {
				t.Errorf("Failed to execute with flag %s: %v", tt.flagName, err)
			}
		})
	}
}

// TestIssue2_BackwardCompatibility tests that old code using --top_k still works
func TestIssue2_BackwardCompatibility(t *testing.T) {
	// Simulate old Client0 frontend code using underscore version
	oldCommandArgs := []string{
		"AuctionImages_OSS",
		"Nikon F2",
		"--top_k", "5",
	}

	cmd := &cobra.Command{
		Use: "query",
		Run: func(cmd *cobra.Command, args []string) {
			topK, err := cmd.Flags().GetInt("top-k")
			if err != nil {
				t.Errorf("Failed to get top-k with old --top_k flag: %v", err)
				return
			}
			if topK != 5 {
				t.Errorf("Expected top-k=5 from old --top_k flag, got %d", topK)
			}
		},
	}

	cmd.Flags().IntP("top-k", "k", 5, "test")
	cmd.Flags().SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		if name == "top_k" {
			name = "top-k"
		}
		return pflag.NormalizedName(name)
	})

	cmd.SetArgs(oldCommandArgs)
	if err := cmd.Execute(); err != nil {
		t.Errorf("Backward compatibility broken: %v", err)
	}
}

// TestIssue2_HelpTextShowsHyphens verifies that help text uses preferred hyphen version
func TestIssue2_HelpTextShowsHyphens(t *testing.T) {
	// This test verifies that when users run --help, they see the preferred
	// hyphen version, not the deprecated underscore version

	cmd := &cobra.Command{
		Use: "query",
		Run: func(cmd *cobra.Command, args []string) {},
	}

	cmd.Flags().IntP("top-k", "k", 5, "Number of top results to return (default: 5)")
	cmd.Flags().Int("top-k-images", 0, "Number of top results from image collections (0 = use top-k)")

	// Get flag usage string
	usage := cmd.Flags().FlagUsages()

	// Verify hyphens appear in help text
	if !contains(usage, "--top-k") {
		t.Error("Help text should show --top-k flag")
	}
	if !contains(usage, "--top-k-images") {
		t.Error("Help text should show --top-k-images flag")
	}

	// Verify underscores don't appear (they're aliases, not in help)
	if contains(usage, "--top_k ") {
		t.Error("Help text should not show deprecated --top_k flag")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
