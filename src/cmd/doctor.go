// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/doctor"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose your weave setup and report issues",
	Long: `Run a comprehensive diagnostic scan of your weave environment.

Checks system dependencies, configuration, environment variables,
VDB connectivity, LLM access, embedding models, stack status, and Opik.

Examples:
  weave doctor                    # Full diagnostic scan
  weave doctor --fix              # Scan + auto-fix fixable issues
  weave doctor --section config   # Only check config section
  weave doctor --json             # Machine-readable output
  weave doctor --verbose          # Show passing checks too`,
	Run: runDoctor,
}

func init() {
	doctorCmd.Flags().Bool("fix", false, "attempt to auto-fix fixable issues")
	doctorCmd.Flags().String("section", "", "run only the specified section (system, config, env, vdb, llm, embeddings, stack, opik)")
	doctorCmd.Flags().Bool("verbose", false, "show all checks including passing ones")
}

func runDoctor(cmd *cobra.Command, args []string) {
	fixFlag, _ := cmd.Flags().GetBool("fix")
	section, _ := cmd.Flags().GetString("section")
	verbose, _ := cmd.Flags().GetBool("verbose")
	jsonFlag := IsJSONOutput()

	// Validate section flag
	if section != "" {
		validSections := map[string]bool{
			"system": true, "config": true, "env": true, "vdb": true,
			"llm": true, "embeddings": true, "stack": true, "opik": true,
		}
		if !validSections[section] {
			printError(fmt.Sprintf("Unknown section: %s", section))
			fmt.Fprintln(os.Stderr, "Valid sections: system, config, env, vdb, llm, embeddings, stack, opik")
			os.Exit(1)
		}
	}

	// Load configuration (may fail – that's ok, CheckConfig will report it)
	cfg, loadErr := LoadConfigWithOverrides()

	dcfg := doctor.DoctorConfig{
		Fix:     fixFlag,
		Section: section,
		JSON:    jsonFlag,
		Verbose: verbose,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	d := doctor.New(cfg, loadErr, dcfg)
	results := d.RunAll(ctx)

	// Output
	if jsonFlag {
		if err := doctor.PrintJSON(results); err != nil {
			printError(fmt.Sprintf("Failed to write JSON: %v", err))
			os.Exit(1)
		}
	} else {
		doctor.PrintResults(results, verbose)
	}

	// --fix: show remediation guidance
	if fixFlag {
		runDoctorFix(cfg, results)
	}

	// Exit with non-zero if any failures
	if doctor.HasFailures(results) {
		os.Exit(1)
	}
}

// runDoctorFix provides auto-fix guidance for fixable issues.
func runDoctorFix(cfg *config.Config, results []doctor.CheckResult) {
	fixable := 0
	for _, r := range results {
		if r.Fix != "" && (r.Status == doctor.StatusFail || r.Status == doctor.StatusWarn) {
			fixable++
		}
	}

	if fixable == 0 {
		return
	}

	fmt.Fprintln(os.Stderr)
	printHeader("Auto-fix suggestions")
	fmt.Fprintln(os.Stderr)

	for _, r := range results {
		if r.Fix == "" || (r.Status != doctor.StatusFail && r.Status != doctor.StatusWarn) {
			continue
		}

		// For config issues, suggest interactive fixer
		if r.Section == doctor.SectionConfig {
			fmt.Fprintf(os.Stderr, "  %s → %s\n", r.Name, "weave config fix --errors-only")
			continue
		}

		// For env vars, print export commands
		if r.Section == doctor.SectionEnv {
			fmt.Fprintf(os.Stderr, "  %s → %s\n", r.Name, r.Fix)
			continue
		}

		// General fix suggestion
		fmt.Fprintf(os.Stderr, "  %s → %s\n", r.Name, r.Fix)
	}
	fmt.Fprintln(os.Stderr)
}
