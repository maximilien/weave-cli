// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package cmd

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestDoctorCommandSelectedSectionWithoutCredentials(t *testing.T) {
	for _, mode := range []string{"text", "JSON", "fix", "missing config"} {
		t.Run(mode, func(t *testing.T) {
			setupHealthConfig(t)
			t.Setenv("OPIK_API_KEY", "")
			previousJSON := jsonOutput
			jsonOutput = mode == "JSON"
			t.Cleanup(func() { jsonOutput = previousJSON })
			if mode == "missing config" {
				viper.Set("config", filepath.Join(t.TempDir(), "missing.yaml"))
			}
			command := &cobra.Command{}
			command.Flags().String("section", "opik", "")
			command.Flags().Bool("verbose", true, "")
			command.Flags().Bool("fix", mode == "fix", "")
			out := captureHealthOutput(t, func() { runDoctor(command, nil) })
			if mode == "JSON" {
				if !json.Valid([]byte(out)) {
					t.Fatalf("invalid diagnostic JSON: %s", out)
				}
			}
			if !strings.Contains(out, "OPIK_API_KEY not set") || strings.Contains(out, "OPENAI_API_KEY") {
				t.Fatalf("unexpected section output: %s", out)
			}
		})
	}
}
