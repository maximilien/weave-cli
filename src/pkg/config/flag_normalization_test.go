// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package config

import (
	"testing"

	"github.com/spf13/pflag"
)

// TestIssue2_FlagNormalization tests the flag normalization logic
// used in collection query command to support both --top-k and --top_k
// Regression test for AuctionsMax.ai Issue #2 (MEDIUM priority)
func TestIssue2_FlagNormalization(t *testing.T) {
	// This tests the normalization function used in query.go
	normalizeFunc := func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		switch name {
		case "top_k":
			name = "top-k"
		case "top_k_images":
			name = "top-k-images"
		}
		return pflag.NormalizedName(name)
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Underscore to hyphen: top_k",
			input:    "top_k",
			expected: "top-k",
		},
		{
			name:     "Hyphen stays: top-k",
			input:    "top-k",
			expected: "top-k",
		},
		{
			name:     "Underscore to hyphen: top_k_images",
			input:    "top_k_images",
			expected: "top-k-images",
		},
		{
			name:     "Hyphen stays: top-k-images",
			input:    "top-k-images",
			expected: "top-k-images",
		},
		{
			name:     "Unrelated flag unchanged",
			input:    "other-flag",
			expected: "other-flag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
			result := normalizeFunc(flags, tt.input)
			if string(result) != tt.expected {
				t.Errorf("normalizeFunc(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestIssue2_FlagSetBehavior tests actual pflag behavior with normalization
func TestIssue2_FlagSetBehavior(t *testing.T) {
	// Create flag set with normalization (simulates query command)
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		switch name {
		case "top_k":
			return "top-k"
		case "top_k_images":
			return "top-k-images"
		}
		return pflag.NormalizedName(name)
	})

	// Define flags with hyphen (preferred)
	flags.IntP("top-k", "k", 5, "Number of results")
	flags.Int("top-k-images", 0, "Number of image results")

	// Test 1: Parse with hyphen version (preferred)
	err := flags.Parse([]string{"--top-k", "10", "--top-k-images", "3"})
	if err != nil {
		t.Fatalf("Failed to parse with hyphen flags: %v", err)
	}

	topK, _ := flags.GetInt("top-k")
	topKImages, _ := flags.GetInt("top-k-images")

	if topK != 10 {
		t.Errorf("Expected top-k=10, got %d", topK)
	}
	if topKImages != 3 {
		t.Errorf("Expected top-k-images=3, got %d", topKImages)
	}

	// Test 2: Parse with underscore version (backward compat)
	flags2 := pflag.NewFlagSet("test2", pflag.ContinueOnError)
	flags2.SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		switch name {
		case "top_k":
			return "top-k"
		case "top_k_images":
			return "top-k-images"
		}
		return pflag.NormalizedName(name)
	})

	flags2.IntP("top-k", "k", 5, "Number of results")
	flags2.Int("top-k-images", 0, "Number of image results")

	err = flags2.Parse([]string{"--top_k", "15", "--top_k_images", "5"})
	if err != nil {
		t.Fatalf("Failed to parse with underscore flags (backward compat): %v", err)
	}

	topK2, _ := flags2.GetInt("top-k")
	topKImages2, _ := flags2.GetInt("top-k-images")

	if topK2 != 15 {
		t.Errorf("Expected top-k=15 (via top_k), got %d", topK2)
	}
	if topKImages2 != 5 {
		t.Errorf("Expected top-k-images=5 (via top_k_images), got %d", topKImages2)
	}

	// Test 3: Mixed usage (hyphen and underscore)
	flags3 := pflag.NewFlagSet("test3", pflag.ContinueOnError)
	flags3.SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		switch name {
		case "top_k":
			return "top-k"
		case "top_k_images":
			return "top-k-images"
		}
		return pflag.NormalizedName(name)
	})

	flags3.IntP("top-k", "k", 5, "Number of results")
	flags3.Int("top-k-images", 0, "Number of image results")

	err = flags3.Parse([]string{"--top-k", "20", "--top_k_images", "7"})
	if err != nil {
		t.Fatalf("Failed to parse with mixed flags: %v", err)
	}

	topK3, _ := flags3.GetInt("top-k")
	topKImages3, _ := flags3.GetInt("top-k-images")

	if topK3 != 20 {
		t.Errorf("Expected top-k=20, got %d", topK3)
	}
	if topKImages3 != 7 {
		t.Errorf("Expected top-k-images=7 (via top_k_images), got %d", topKImages3)
	}
}

// TestIssue2_ShortFlagAlias tests that -k shorthand still works
func TestIssue2_ShortFlagAlias(t *testing.T) {
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.IntP("top-k", "k", 5, "Number of results")

	// Parse with short flag
	err := flags.Parse([]string{"-k", "25"})
	if err != nil {
		t.Fatalf("Failed to parse with short flag -k: %v", err)
	}

	topK, _ := flags.GetInt("top-k")
	if topK != 25 {
		t.Errorf("Expected top-k=25 via -k, got %d", topK)
	}
}

// TestIssue2_DefaultValues tests that default values work correctly
func TestIssue2_DefaultValues(t *testing.T) {
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.IntP("top-k", "k", 5, "Number of results")
	flags.Int("top-k-images", 0, "Number of image results")

	// Parse with no flags set
	err := flags.Parse([]string{})
	if err != nil {
		t.Fatalf("Failed to parse with defaults: %v", err)
	}

	topK, _ := flags.GetInt("top-k")
	topKImages, _ := flags.GetInt("top-k-images")

	if topK != 5 {
		t.Errorf("Expected default top-k=5, got %d", topK)
	}
	if topKImages != 0 {
		t.Errorf("Expected default top-k-images=0, got %d", topKImages)
	}
}
