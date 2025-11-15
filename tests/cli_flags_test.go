// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package tests

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestMongoDBFlag tests the --mongodb flag on various commands
func TestMongoDBFlag(t *testing.T) {
	// Skip if no MongoDB credentials are provided
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		t.Skip("Skipping MongoDB flag test: MONGODB_URI environment variable not set")
	}

	binary := "../bin/weave"
	if _, err := os.Stat(binary); os.IsNotExist(err) {
		t.Skip("Skipping MongoDB flag test: weave binary not found (run ./build.sh first)")
	}

	tests := []struct {
		name    string
		command []string
		wantErr bool
	}{
		{
			name:    "collection list with --mongodb",
			command: []string{"collection", "list", "--mongodb"},
			wantErr: false,
		},
		{
			name:    "collection count with --mongodb",
			command: []string{"collection", "count", "--mongodb"},
			wantErr: false,
		},
		{
			name:    "docs ls with --mongodb",
			command: []string{"docs", "ls", "--mongodb"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binary, tt.command...)
			cmd.Env = os.Environ()
			output, err := cmd.CombinedOutput()

			if tt.wantErr && err == nil {
				t.Errorf("Expected error but got none. Output: %s", string(output))
			}

			if !tt.wantErr && err != nil {
				// Check if error is "unknown flag" - this would indicate flag wasn't added
				if strings.Contains(string(output), "unknown flag: --mongodb") {
					t.Errorf("Flag --mongodb not recognized. Output: %s", string(output))
				}
				// Other errors are OK for this test (e.g., connection errors)
			}
		})
	}
}

// TestSupabaseFlag tests the --supabase flag on various commands
func TestSupabaseFlag(t *testing.T) {
	// Skip if no Supabase credentials are provided
	supabaseURL := os.Getenv("SUPABASE_DATABASE_URL")
	if supabaseURL == "" {
		t.Skip("Skipping Supabase flag test: SUPABASE_DATABASE_URL environment variable not set")
	}

	binary := "../bin/weave"
	if _, err := os.Stat(binary); os.IsNotExist(err) {
		t.Skip("Skipping Supabase flag test: weave binary not found (run ./build.sh first)")
	}

	tests := []struct {
		name    string
		command []string
		wantErr bool
	}{
		{
			name:    "collection list with --supabase",
			command: []string{"collection", "list", "--supabase"},
			wantErr: false,
		},
		{
			name:    "collection count with --supabase",
			command: []string{"collection", "count", "--supabase"},
			wantErr: false,
		},
		{
			name:    "docs ls with --supabase",
			command: []string{"docs", "ls", "--supabase"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binary, tt.command...)
			cmd.Env = os.Environ()
			output, err := cmd.CombinedOutput()

			if tt.wantErr && err == nil {
				t.Errorf("Expected error but got none. Output: %s", string(output))
			}

			if !tt.wantErr && err != nil {
				// Check if error is "unknown flag" - this would indicate flag wasn't added
				if strings.Contains(string(output), "unknown flag: --supabase") {
					t.Errorf("Flag --supabase not recognized. Output: %s", string(output))
				}
				// Other errors are OK for this test (e.g., connection errors)
			}
		})
	}
}

// TestDatabaseFlagsExclusive tests that database flags work correctly
func TestDatabaseFlagsExclusive(t *testing.T) {
	binary := "../bin/weave"
	if _, err := os.Stat(binary); os.IsNotExist(err) {
		t.Skip("Skipping flag exclusivity test: weave binary not found (run ./build.sh first)")
	}

	// Test that --mongodb flag is recognized
	t.Run("mongodb flag recognized", func(t *testing.T) {
		cmd := exec.Command(binary, "collection", "list", "--mongodb", "--help")
		cmd.Env = os.Environ()
		output, _ := cmd.CombinedOutput()

		if strings.Contains(string(output), "unknown flag") {
			t.Errorf("--mongodb flag should be recognized. Output: %s", string(output))
		}
	})

	// Test that --supabase flag is recognized
	t.Run("supabase flag recognized", func(t *testing.T) {
		cmd := exec.Command(binary, "collection", "list", "--supabase", "--help")
		cmd.Env = os.Environ()
		output, _ := cmd.CombinedOutput()

		if strings.Contains(string(output), "unknown flag") {
			t.Errorf("--supabase flag should be recognized. Output: %s", string(output))
		}
	})

	// Test that --weaviate flag is recognized
	t.Run("weaviate flag recognized", func(t *testing.T) {
		cmd := exec.Command(binary, "collection", "list", "--weaviate", "--help")
		cmd.Env = os.Environ()
		output, _ := cmd.CombinedOutput()

		if strings.Contains(string(output), "unknown flag") {
			t.Errorf("--weaviate flag should be recognized. Output: %s", string(output))
		}
	})

	// Test that --mock flag is recognized
	t.Run("mock flag recognized", func(t *testing.T) {
		cmd := exec.Command(binary, "collection", "list", "--mock", "--help")
		cmd.Env = os.Environ()
		output, _ := cmd.CombinedOutput()

		if strings.Contains(string(output), "unknown flag") {
			t.Errorf("--mock flag should be recognized. Output: %s", string(output))
		}
	})
}
