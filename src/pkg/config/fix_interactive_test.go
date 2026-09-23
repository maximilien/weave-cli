// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package config

import (
	"bufio"
	"strings"
	"testing"
)

func newFixerWithInput(issues []ConfigIssue, input string) *InteractiveFixer {
	fixer := NewInteractiveFixer(issues)
	fixer.reader = bufio.NewReader(strings.NewReader(input))
	return fixer
}

func TestInteractiveFixerPresentationAndConfirmation(t *testing.T) {
	issues := []ConfigIssue{
		{Type: IssueTypeError, DatabaseName: "cloud", DatabaseIdx: 0, Field: "url", Message: "missing URL", Hint: "use HTTPS"},
		{Type: IssueTypeWarning, DatabaseName: "local", DatabaseIdx: 1, Field: "port", Message: "missing port"},
	}
	fixer := newFixerWithInput(issues, "yes\n")
	fixer.showSummary()
	fixer.showSeparator()
	fixer.showIssueHeader(1, len(issues), issues[0])
	fixer.showIssueHeader(2, len(issues), issues[1])
	if !fixer.confirmProceed() {
		t.Fatal("expected affirmative confirmation")
	}
	if newFixerWithInput(issues, "n\n").confirmProceed() {
		t.Fatal("expected negative confirmation")
	}
	if newFixerWithInput(issues, "").confirmProceed() {
		t.Fatal("expected read failure to decline confirmation")
	}
	if _, err := NewInteractiveFixer(issues).Run(); err == nil || !strings.Contains(err.Error(), "requires a terminal") {
		t.Fatalf("expected non-terminal error, got %v", err)
	}
}

func TestInteractiveFixerActions(t *testing.T) {
	issue := ConfigIssue{Type: IssueTypeError, DatabaseName: "cloud", DatabaseIdx: 0, Field: "url", Message: "missing URL"}
	tests := []struct {
		name   string
		input  string
		action FixAction
		value  string
	}{
		{name: "set value", input: "1\nhttps://example.test\n", action: FixActionSetValue, value: "https://example.test"},
		{name: "skip", input: "2\n", action: FixActionSkip},
		{name: "remove", input: "3\n", action: FixActionRemove},
		{name: "disable", input: "4\n", action: FixActionDisable},
		{name: "quit", input: "5\n", action: FixActionQuit},
		{name: "invalid", input: "invalid\n", action: FixActionSkip},
		{name: "out of range", input: "9\n", action: FixActionSkip},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := newFixerWithInput([]ConfigIssue{issue}, tt.input).promptFixAction(issue)
			if err != nil {
				t.Fatal(err)
			}
			if result.Action != tt.action || result.Value != tt.value {
				t.Fatalf("result = %#v, want action %s value %q", result, tt.action, tt.value)
			}
		})
	}
	if _, err := newFixerWithInput([]ConfigIssue{issue}, "").promptFixAction(issue); err == nil {
		t.Fatal("expected choice read error")
	}

	nonDatabaseIssue := ConfigIssue{Type: IssueTypeWarning, DatabaseIdx: -1, Field: "url", Message: "missing"}
	result, err := newFixerWithInput([]ConfigIssue{nonDatabaseIssue}, "3\n").promptFixAction(nonDatabaseIssue)
	if err != nil || result.Action != FixActionQuit {
		t.Fatalf("non-database quit = (%#v, %v)", result, err)
	}
}

func TestInteractiveFixerValueValidation(t *testing.T) {
	issue := ConfigIssue{Type: IssueTypeError, DatabaseName: "cloud", DatabaseIdx: 0, Field: "url", Message: "missing URL"}
	result, err := newFixerWithInput([]ConfigIssue{issue}, "\n").promptForValue(issue)
	if err != nil || result.Action != FixActionSkip {
		t.Fatalf("empty value = (%#v, %v)", result, err)
	}

	result, err = newFixerWithInput([]ConfigIssue{issue}, "not-a-url\nstill-bad\nnope\n").promptForValue(issue)
	if err != nil || result.Action != FixActionSkip {
		t.Fatalf("invalid values = (%#v, %v)", result, err)
	}

	result, err = newFixerWithInput([]ConfigIssue{issue}, "bad\nhttps://example.test\n").promptForValue(issue)
	if err != nil || result.Action != FixActionSetValue || result.Value != "https://example.test" {
		t.Fatalf("retried value = (%#v, %v)", result, err)
	}

	if _, err := newFixerWithInput([]ConfigIssue{issue}, "").promptForValue(issue); err == nil {
		t.Fatal("expected value read error")
	}
}

func TestInteractiveFixerSummaryResults(t *testing.T) {
	issues := []ConfigIssue{
		{Type: IssueTypeError, DatabaseName: "cloud", DatabaseIdx: 0, Field: "url"},
		{Type: IssueTypeWarning, DatabaseName: "cloud", DatabaseIdx: 0, Field: "api_key"},
	}
	results := []FixResult{
		{Issue: issues[0], Action: FixActionSetValue, Value: "https://example.test"},
		{Issue: issues[1], Action: FixActionSetValue, Value: "secret-value"},
		{Issue: issues[0], Action: FixActionSkip},
		{Issue: issues[0], Action: FixActionRemove},
		{Issue: issues[0], Action: FixActionDisable},
	}
	fixer := newFixerWithInput(issues, "y\n")
	fixer.results = results
	for _, result := range results {
		fixer.showActionResult(result)
	}
	got, err := fixer.showSummaryAndConfirm()
	if err != nil || len(got) != len(results) {
		t.Fatalf("confirmed summary = (%#v, %v)", got, err)
	}

	fixer = newFixerWithInput(issues, "no\n")
	fixer.results = results
	if _, err := fixer.showSummaryAndConfirm(); err == nil || !strings.Contains(err.Error(), "cancelled") {
		t.Fatalf("expected cancellation, got %v", err)
	}
	fixer = newFixerWithInput(issues, "")
	if _, err := fixer.showSummaryAndConfirm(); err == nil || !strings.Contains(err.Error(), "confirmation") {
		t.Fatalf("expected confirmation read error, got %v", err)
	}
}
