// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package config

import (
	"bufio"
	"strings"
	"testing"
)

func TestInteractiveFixChoicesAndValidation(t *testing.T) {
	issue := ConfigIssue{Type: IssueTypeError, DatabaseIdx: 0, DatabaseName: "fixture", Field: "url", Message: "missing URL", Hint: "include scheme"}
	for _, tc := range []struct {
		name, input string
		action      FixAction
		value       string
	}{
		{"valid after retry", "1\nnot-a-url\nhttps://example.test\n", FixActionSetValue, "https://example.test"},
		{"three invalid values", "1\nbad\nbad\nbad\n", FixActionSkip, ""},
		{"empty value", "1\n\n", FixActionSkip, ""},
		{"skip", "2\n", FixActionSkip, ""},
		{"remove", "3\n", FixActionRemove, ""},
		{"disable", "4\n", FixActionDisable, ""},
		{"quit", "5\n", FixActionQuit, ""},
		{"invalid choice", "wat\n", FixActionSkip, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixer := &InteractiveFixer{reader: bufio.NewReader(strings.NewReader(tc.input))}
			result, err := fixer.promptFixAction(issue)
			if err != nil || result.Action != tc.action || result.Value != tc.value {
				t.Fatalf("result = %+v, %v", result, err)
			}
		})
	}
	fixer := &InteractiveFixer{reader: bufio.NewReader(strings.NewReader(""))}
	if _, err := fixer.promptFixAction(issue); err == nil {
		t.Fatal("expected input error")
	}
}

func TestInteractiveFixReview(t *testing.T) {
	issue := ConfigIssue{Type: IssueTypeWarning, DatabaseIdx: 0, DatabaseName: "fixture", Field: "api_key"}
	results := []FixResult{
		{Issue: issue, Action: FixActionSetValue, Value: "fixture-secret"},
		{Issue: issue, Action: FixActionSkip},
		{Issue: issue, Action: FixActionRemove},
		{Issue: issue, Action: FixActionDisable},
	}
	fixer := &InteractiveFixer{issues: []ConfigIssue{issue, {Type: IssueTypeError}}, results: results}
	fixer.showSummary()
	fixer.showIssueHeader(1, 2, issue)
	for _, result := range results {
		fixer.showActionResult(result)
	}
	for _, tc := range []struct {
		input    string
		accepted bool
	}{{"yes\n", true}, {"n\n", false}, {"", false}} {
		fixer.reader = bufio.NewReader(strings.NewReader(tc.input))
		got, err := fixer.showSummaryAndConfirm()
		if (err == nil) != tc.accepted || (tc.accepted && len(got) != len(results)) {
			t.Fatalf("confirmation %q: %+v, %v", tc.input, got, err)
		}
		fixer.reader = bufio.NewReader(strings.NewReader(tc.input))
		if fixer.confirmProceed() != tc.accepted {
			t.Fatalf("proceed %q", tc.input)
		}
	}
}
