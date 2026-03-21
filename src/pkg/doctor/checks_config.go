// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package doctor

import (
	"fmt"
	"os"

	"github.com/maximilien/weave-cli/src/pkg/config"
)

// CheckConfig verifies that the config file exists, is parseable, and validates cleanly.
// cfg may be nil if loading failed; the caller should pass the error in loadErr.
func CheckConfig(cfg *config.Config, loadErr error) []CheckResult {
	var results []CheckResult

	// Check config file existence
	configPaths := []string{"config.yaml", "config.yml"}
	found := ""
	for _, p := range configPaths {
		if _, err := os.Stat(p); err == nil {
			found = p
			break
		}
	}

	if found == "" {
		results = append(results, CheckResult{
			Section: SectionConfig,
			Name:    "Config file",
			Status:  StatusFail,
			Message: "config.yaml not found in current directory",
			Fix:     "Run: weave config create --env",
		})
		return results
	}

	results = append(results, CheckResult{
		Section: SectionConfig,
		Name:    "Config file",
		Status:  StatusOK,
		Message: found,
	})

	// Check parseability
	if loadErr != nil {
		results = append(results, CheckResult{
			Section: SectionConfig,
			Name:    "Config parsing",
			Status:  StatusFail,
			Message: fmt.Sprintf("Failed to parse: %v", loadErr),
			Fix:     "Check config.yaml syntax or run: weave config create --env",
		})
		return results
	}

	results = append(results, CheckResult{
		Section: SectionConfig,
		Name:    "Config parsing",
		Status:  StatusOK,
		Message: "valid YAML",
	})

	// Run validation
	vResult := config.ValidateConfig(cfg)
	if vResult == nil {
		return results
	}

	if len(vResult.Errors) == 0 && len(vResult.Warnings) == 0 {
		results = append(results, CheckResult{
			Section: SectionConfig,
			Name:    "Config validation",
			Status:  StatusOK,
			Message: "no issues found",
		})
		return results
	}

	for _, e := range vResult.Errors {
		results = append(results, CheckResult{
			Section: SectionConfig,
			Name:    fmt.Sprintf("Config: %s", e.Field),
			Status:  StatusFail,
			Message: e.Message,
			Fix:     e.Suggestion,
		})
	}

	for _, w := range vResult.Warnings {
		results = append(results, CheckResult{
			Section: SectionConfig,
			Name:    fmt.Sprintf("Config: %s", w.Field),
			Status:  StatusWarn,
			Message: w.Message,
			Fix:     w.Suggestion,
		})
	}

	return results
}
