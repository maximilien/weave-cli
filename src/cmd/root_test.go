// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package cmd

import (
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/maximilien/weave-cli/src/pkg/doctor"
	"github.com/spf13/cobra"
)

func TestRunDoctorFixSuggestions(t *testing.T) {
	results := []doctor.CheckResult{
		{Section: doctor.SectionConfig, Name: "Config file", Status: doctor.StatusFail, Fix: "replace config"},
		{Section: doctor.SectionEnv, Name: "API key", Status: doctor.StatusWarn, Fix: "export API_KEY=value"},
		{Section: doctor.SectionSystem, Name: "Tool", Status: doctor.StatusFail, Fix: "install tool"},
		{Section: doctor.SectionSystem, Name: "Healthy", Status: doctor.StatusOK, Fix: "ignored"},
		{Section: doctor.SectionSystem, Name: "No guidance", Status: doctor.StatusFail},
	}
	output := captureHealthOutput(t, func() { runDoctorFix(nil, results) })
	for _, want := range []string{
		"Auto-fix suggestions",
		"Config file → weave config fix --errors-only",
		"API key → export API_KEY=value",
		"Tool → install tool",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("doctor fix output missing %q:\n%s", want, output)
		}
	}
	for _, unwanted := range []string{"Healthy", "No guidance"} {
		if strings.Contains(output, unwanted) {
			t.Errorf("doctor fix output unexpectedly contains %q:\n%s", unwanted, output)
		}
	}

	if output := captureHealthOutput(t, func() {
		runDoctorFix(nil, []doctor.CheckResult{{Status: doctor.StatusOK, Fix: "ignored"}})
	}); output != "" {
		t.Fatalf("expected no suggestions, got %q", output)
	}
}

func TestStaticCompletionValues(t *testing.T) {
	tests := []struct {
		name    string
		fn      cobra.CompletionFunc
		minimum int
		prefix  string
	}{
		{name: "database types", fn: validDatabaseTypes, minimum: 17, prefix: "weaviate-cloud"},
		{name: "config fields", fn: validConfigFields, minimum: 7, prefix: "url"},
		{name: "log levels", fn: validLogLevels, minimum: 4, prefix: "debug"},
		{name: "log formats", fn: validLogFormats, minimum: 3, prefix: "json"},
		{name: "timeouts", fn: validTimeouts, minimum: 5, prefix: "5s"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, directive := tt.fn(nil, nil, "")
			if directive != cobra.ShellCompDirectiveNoFileComp || len(values) < tt.minimum {
				t.Fatalf("completion = %#v, %v", values, directive)
			}
			if !strings.HasPrefix(values[0], tt.prefix) {
				t.Fatalf("first completion = %q, want prefix %q", values[0], tt.prefix)
			}
		})
	}
}

func TestRegisterFlagCompletions(t *testing.T) {
	command := &cobra.Command{Use: "test"}
	for _, name := range []string{"vector-db-type", "vdb", "log-level", "log-format", "timeout"} {
		command.Flags().String(name, "", "test flag")
	}
	registerFlagCompletions(command)

	for _, name := range []string{"vector-db-type", "vdb", "log-level", "log-format", "timeout"} {
		completion, ok := command.GetFlagCompletionFunc(name)
		if !ok {
			t.Fatalf("completion not registered for %s", name)
		}
		values, directive := completion(command, nil, "")
		if len(values) == 0 || directive != cobra.ShellCompDirectiveNoFileComp {
			t.Fatalf("completion for %s = %#v, %v", name, values, directive)
		}
	}
}

func TestDatabaseNameCompletionWithoutConfig(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("HOME", t.TempDir())
	values, directive := validDatabaseNames(nil, nil, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("directive = %v", directive)
	}
	if values == nil {
		t.Fatal("validDatabaseNames() returned nil")
	}
}

func TestRootFlagAndPresentationHelpers(t *testing.T) {
	command := &cobra.Command{Use: "test"}
	command.Flags().Bool("version", false, "version")
	if err := handleVersionFlag(command, nil); err != nil {
		t.Fatalf("handleVersionFlag() error: %v", err)
	}
	if err := command.Flags().Set("version", "false"); err != nil {
		t.Fatalf("set version flag: %v", err)
	}
	if err := handleVersionFlag(command, nil); err != nil {
		t.Fatalf("handleVersionFlag(changed false) error: %v", err)
	}

	oldNoTips, oldJSON, oldNoColor := noTips, jsonOutput, noColor
	oldColorSetting := color.NoColor
	t.Cleanup(func() {
		noTips, jsonOutput, noColor = oldNoTips, oldJSON, oldNoColor
		color.NoColor = oldColorSetting
	})
	noTips = false
	if !ShouldShowTips() {
		t.Fatal("ShouldShowTips() = false")
	}
	noTips = true
	if ShouldShowTips() {
		t.Fatal("ShouldShowTips() = true")
	}
	jsonOutput = true
	if !IsJSONOutput() {
		t.Fatal("IsJSONOutput() = false")
	}
	noColor = true
	initColor()
	if !color.NoColor {
		t.Fatal("initColor() did not disable color")
	}

	template := getGroupedUsageTemplate()
	for _, expected := range []string{"Database Selection:", "Output Control:", "Global Flags:"} {
		if !strings.Contains(template, expected) {
			t.Errorf("usage template missing %q", expected)
		}
	}
	printHeader("header")
	printSuccess("success")
	printWarning("warning")
	printError("error")
}
