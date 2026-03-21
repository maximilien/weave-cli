// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package doctor

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fatih/color"
)

// statusTag returns a coloured bracketed tag like "[OK]  " for terminal output.
func statusTag(s Status) string {
	switch s {
	case StatusOK:
		return color.GreenString("[OK]  ")
	case StatusWarn:
		return color.YellowString("[WARN]")
	case StatusFail:
		return color.RedString("[FAIL]")
	case StatusSkip:
		return color.HiBlackString("[SKIP]")
	default:
		return "[????]"
	}
}

// PrintResults renders all check results to the terminal, grouped by section.
// When verbose is false only non-OK results and section headers with issues are shown.
func PrintResults(results []CheckResult, verbose bool) {
	if len(results) == 0 {
		fmt.Fprintln(os.Stderr, "No checks were executed.")
		return
	}

	currentSection := ""
	sectionHasOutput := false

	// Two-pass approach: first collect results per section so we can decide
	// whether to print the header.  For simplicity we do a single pass and
	// buffer the section header, printing it lazily.
	for _, r := range results {
		show := verbose || r.Status != StatusOK

		if r.Section != currentSection {
			currentSection = r.Section
			sectionHasOutput = false
		}

		if !show {
			continue
		}

		// Lazy section header
		if !sectionHasOutput {
			if currentSection != results[0].Section {
				fmt.Fprintln(os.Stderr)
			}
			color.New(color.FgCyan, color.Bold).Fprintf(os.Stderr, "%s\n", sectionTitle(currentSection))
			sectionHasOutput = true
		}

		// Check line
		fmt.Fprintf(os.Stderr, "  %s %s", statusTag(r.Status), r.Name)
		if r.Message != "" {
			fmt.Fprintf(os.Stderr, ": %s", r.Message)
		}
		fmt.Fprintln(os.Stderr)

		// Fix suggestion
		if r.Fix != "" && (r.Status == StatusFail || r.Status == StatusWarn) {
			color.New(color.FgHiBlack).Fprintf(os.Stderr, "         Fix: %s\n", r.Fix)
		}
	}

	// Summary footer
	fmt.Fprintln(os.Stderr)
	ok, warn, fail, skip := countByStatus(results)
	fmt.Fprintf(os.Stderr, "Summary: ")
	if ok > 0 {
		color.New(color.FgGreen).Fprintf(os.Stderr, "%d passed", ok)
	}
	if warn > 0 {
		if ok > 0 {
			fmt.Fprint(os.Stderr, ", ")
		}
		color.New(color.FgYellow).Fprintf(os.Stderr, "%d warnings", warn)
	}
	if fail > 0 {
		if ok > 0 || warn > 0 {
			fmt.Fprint(os.Stderr, ", ")
		}
		color.New(color.FgRed).Fprintf(os.Stderr, "%d failed", fail)
	}
	if skip > 0 {
		if ok > 0 || warn > 0 || fail > 0 {
			fmt.Fprint(os.Stderr, ", ")
		}
		color.New(color.FgHiBlack).Fprintf(os.Stderr, "%d skipped", skip)
	}
	fmt.Fprintln(os.Stderr)
}

// PrintJSON outputs results as a JSON object to stdout.
func PrintJSON(results []CheckResult) error {
	ok, warn, fail, skip := countByStatus(results)
	output := struct {
		Results []CheckResult `json:"results"`
		Summary summaryJSON   `json:"summary"`
	}{
		Results: results,
		Summary: summaryJSON{
			Total:    len(results),
			Passed:   ok,
			Warnings: warn,
			Failed:   fail,
			Skipped:  skip,
		},
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}

type summaryJSON struct {
	Total    int `json:"total"`
	Passed   int `json:"passed"`
	Warnings int `json:"warnings"`
	Failed   int `json:"failed"`
	Skipped  int `json:"skipped"`
}

// HasFailures returns true if any result is StatusFail.
func HasFailures(results []CheckResult) bool {
	for _, r := range results {
		if r.Status == StatusFail {
			return true
		}
	}
	return false
}

func countByStatus(results []CheckResult) (ok, warn, fail, skip int) {
	for _, r := range results {
		switch r.Status {
		case StatusOK:
			ok++
		case StatusWarn:
			warn++
		case StatusFail:
			fail++
		case StatusSkip:
			skip++
		}
	}
	return
}

func sectionTitle(section string) string {
	switch section {
	case SectionSystem:
		return "System"
	case SectionConfig:
		return "Config"
	case SectionEnv:
		return "Environment"
	case SectionVDB:
		return "VDB Connectivity"
	case SectionLLM:
		return "LLM"
	case SectionEmbeddings:
		return "Embeddings"
	case SectionStack:
		return "Stack"
	case SectionOpik:
		return "Opik"
	default:
		return section
	}
}
