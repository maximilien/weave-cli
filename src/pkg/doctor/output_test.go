// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package doctor

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/fatih/color"
)

func captureDoctorOutput(t *testing.T, run func()) (string, string) {
	t.Helper()
	oldOut, oldErr, oldNoColor := os.Stdout, os.Stderr, color.NoColor
	outReader, outWriter, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	errReader, errWriter, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout, os.Stderr, color.NoColor = outWriter, errWriter, true
	run()
	_ = outWriter.Close()
	_ = errWriter.Close()
	os.Stdout, os.Stderr, color.NoColor = oldOut, oldErr, oldNoColor
	outData, readOutErr := io.ReadAll(outReader)
	errData, readErrErr := io.ReadAll(errReader)
	_ = outReader.Close()
	_ = errReader.Close()
	if readOutErr != nil || readErrErr != nil {
		t.Fatalf("read captured output: %v, %v", readOutErr, readErrErr)
	}
	return string(outData), string(errData)
}

func TestPrintResultsVerboseAndFiltered(t *testing.T) {
	results := []CheckResult{
		{Section: SectionSystem, Name: "Go", Status: StatusOK, Message: "available"},
		{Section: SectionConfig, Name: "Config", Status: StatusWarn, Message: "partial", Fix: "edit config"},
		{Section: SectionEnv, Name: "API key", Status: StatusFail, Message: "missing", Fix: "export API_KEY=x"},
		{Section: SectionStack, Name: "Runtime", Status: StatusSkip, Message: "disabled"},
	}
	_, verbose := captureDoctorOutput(t, func() { PrintResults(results, true) })
	for _, want := range []string{
		"System", "[OK]", "Config", "[WARN]", "Fix: edit config",
		"Environment", "[FAIL]", "Stack", "[SKIP]",
		"1 passed, 1 warnings, 1 failed, 1 skipped",
	} {
		if !strings.Contains(verbose, want) {
			t.Errorf("verbose output missing %q:\n%s", want, verbose)
		}
	}

	_, filtered := captureDoctorOutput(t, func() { PrintResults(results, false) })
	if strings.Contains(filtered, "Go: available") || !strings.Contains(filtered, "Config: partial") {
		t.Fatalf("unexpected filtered output:\n%s", filtered)
	}

	_, empty := captureDoctorOutput(t, func() { PrintResults(nil, false) })
	if !strings.Contains(empty, "No checks were executed") {
		t.Fatalf("unexpected empty output: %q", empty)
	}
}

func TestPrintJSONAndStatusTags(t *testing.T) {
	results := []CheckResult{
		{Section: SectionSystem, Name: "ok", Status: StatusOK},
		{Section: SectionEnv, Name: "warn", Status: StatusWarn},
		{Section: SectionVDB, Name: "fail", Status: StatusFail},
		{Section: SectionOpik, Name: "skip", Status: StatusSkip},
	}
	stdout, _ := captureDoctorOutput(t, func() {
		if err := PrintJSON(results); err != nil {
			t.Fatal(err)
		}
	})
	var decoded struct {
		Results []json.RawMessage `json:"results"`
		Summary summaryJSON       `json:"summary"`
	}
	if err := json.Unmarshal([]byte(stdout), &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Summary.Total != 4 || decoded.Summary.Passed != 1 || decoded.Summary.Warnings != 1 || decoded.Summary.Failed != 1 || decoded.Summary.Skipped != 1 {
		t.Fatalf("unexpected JSON summary: %#v", decoded.Summary)
	}

	for status, want := range map[Status]string{
		StatusOK: "[OK]", StatusWarn: "[WARN]", StatusFail: "[FAIL]", StatusSkip: "[SKIP]", Status(99): "[????]",
	} {
		if got := statusTag(status); !strings.Contains(got, want) {
			t.Errorf("statusTag(%v) = %q, want %q", status, got, want)
		}
	}
}
