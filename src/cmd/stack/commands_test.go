// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package stack

import (
	"os"
	"strings"
	"testing"

	stackpkg "github.com/maximilien/weave-cli/src/pkg/stack"
)

func TestStackCommandContracts(t *testing.T) {
	if err := LogsCmd.Args(LogsCmd, []string{"milvus", "extra"}); err == nil {
		t.Fatal("LogsCmd accepted more than one service")
	}
	if err := LogsCmd.Args(LogsCmd, []string{"milvus"}); err != nil {
		t.Fatalf("LogsCmd rejected one service: %v", err)
	}
	if CollectionsCmd.Commands()[0] != collectionsResetCmd {
		t.Fatal("collections reset command is not registered")
	}
	found, _, err := CollectionsCmd.Find([]string{"reset"})
	if err != nil || found != collectionsResetCmd {
		t.Fatalf("find collections reset command = (%v, %v)", found, err)
	}
	if collectionsResetCmd.Flags().Lookup("force") == nil || collectionsResetCmd.Flags().Lookup("keep") == nil {
		t.Fatal("collections reset flags are not registered")
	}
}

func TestCommandsHandleMissingStack(t *testing.T) {
	inTempWorkingDir(t)

	oldIngestAll := ingestAll
	t.Cleanup(func() { ingestAll = oldIngestAll })
	ingestAll = false

	if err := runStatus(nil, nil); err != nil {
		t.Fatalf("runStatus() with no active stack error = %v", err)
	}
	if err := runLogs(nil, nil); err == nil || !strings.Contains(err.Error(), "no active stack") {
		t.Fatalf("runLogs() error = %v, want no-active-stack error", err)
	}
	if err := runCollectionsReset(nil, nil); err == nil || !strings.Contains(err.Error(), "no active stack") {
		t.Fatalf("runCollectionsReset() error = %v, want no-active-stack error", err)
	}
	if err := runDown(nil, nil); err != nil {
		t.Fatalf("runDown() with no active stack error = %v", err)
	}
	if err := runKubectl(nil, []string{"--", "get", "pods"}); err == nil || !strings.Contains(err.Error(), "no active stack") {
		t.Fatalf("runKubectl() error = %v, want no-active-stack error", err)
	}
	if err := runPortForward(nil, []string{"milvus", "19530:19530"}); err == nil || !strings.Contains(err.Error(), "no active stack") {
		t.Fatalf("runPortForward() error = %v, want no-active-stack error", err)
	}
	if err := runIngest(nil, []string{"Documents", "data"}); err == nil || !strings.Contains(err.Error(), "no active stack") {
		t.Fatalf("runIngest() error = %v, want no-active-stack error", err)
	}
}

func TestIngestValidationAndDefaults(t *testing.T) {
	inTempWorkingDir(t)

	oldIngestAll := ingestAll
	t.Cleanup(func() { ingestAll = oldIngestAll })

	ingestAll = false
	if err := runIngest(nil, []string{"Documents"}); err == nil || !strings.Contains(err.Error(), "requires collection name") {
		t.Fatalf("runIngest() error = %v, want argument error", err)
	}

	ingestAll = true
	if err := runIngest(nil, nil); err == nil || !strings.Contains(err.Error(), "failed to load stack config") {
		t.Fatalf("runIngest() --all error = %v, want missing-config error", err)
	}

	if got := getConfigValue("", "fallback"); got != "fallback" {
		t.Fatalf("getConfigValue(empty) = %q", got)
	}
	if got := getConfigValue("configured", "fallback"); got != "configured" {
		t.Fatalf("getConfigValue(configured) = %q", got)
	}
	if got := getConfigIntValue(0, 10); got != 10 {
		t.Fatalf("getConfigIntValue(0) = %d", got)
	}
	if got := getConfigIntValue(5, 10); got != 5 {
		t.Fatalf("getConfigIntValue(5) = %d", got)
	}
}

func TestCommandsReportMissingConfiguration(t *testing.T) {
	inTempWorkingDir(t)

	tests := []struct {
		name string
		run  func() error
	}{
		{name: "dashboard start", run: func() error { return runDashboardStart(nil, nil) }},
		{name: "dashboard stop", run: func() error { return runDashboardStop(nil, nil) }},
		{name: "dashboard restart", run: func() error { return runDashboardRestart(nil, nil) }},
		{name: "dashboard status", run: func() error { return runDashboardStatus(nil, nil) }},
		{name: "dashboard logs", run: func() error { return runDashboardLogs(nil, nil) }},
		{name: "stack up", run: func() error { return runUp(nil, nil) }},
		{name: "stack backup", run: func() error { return runStackBackup(nil, []string{"Documents"}) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.run(); err == nil || !strings.Contains(err.Error(), "failed to load stack config") {
				t.Fatalf("command error = %v, want missing-config error", err)
			}
		})
	}
}

func TestDashboardConfigurationValidation(t *testing.T) {
	inTempWorkingDir(t)

	config := generateQuickstartTemplate("kind")
	if err := stackpkg.SaveStackConfig(config, ""); err != nil {
		t.Fatalf("save stack config: %v", err)
	}
	if err := runDashboardStart(nil, nil); err == nil || !strings.Contains(err.Error(), "dashboard not enabled") {
		t.Fatalf("runDashboardStart() error = %v, want disabled error", err)
	}

	config.Dashboard = &stackpkg.DashboardConfig{Enabled: true, Runtime: "kubernetes"}
	if err := stackpkg.SaveStackConfig(config, ""); err != nil {
		t.Fatalf("save dashboard config: %v", err)
	}
	if err := runDashboardStart(nil, nil); err == nil || !strings.Contains(err.Error(), "not pm2") {
		t.Fatalf("runDashboardStart() error = %v, want runtime error", err)
	}

	config.Dashboard.Runtime = "pm2"
	if err := stackpkg.SaveStackConfig(config, ""); err != nil {
		t.Fatalf("save PM2 config: %v", err)
	}
	if err := runDashboardStart(nil, nil); err == nil || !strings.Contains(err.Error(), "PM2 configuration missing") {
		t.Fatalf("runDashboardStart() error = %v, want missing-PM2 error", err)
	}

	for name, run := range map[string]func() error{
		"stop":    func() error { return runDashboardStop(nil, nil) },
		"restart": func() error { return runDashboardRestart(nil, nil) },
		"status":  func() error { return runDashboardStatus(nil, nil) },
		"logs":    func() error { return runDashboardLogs(nil, nil) },
	} {
		t.Run(name, func(t *testing.T) {
			if err := run(); err == nil || !strings.Contains(err.Error(), "PM2 config not found") {
				t.Fatalf("dashboard %s error = %v, want missing-PM2 error", name, err)
			}
		})
	}
}

func TestRunValidate(t *testing.T) {
	inTempWorkingDir(t)

	if err := stackpkg.SaveStackConfig(generateQuickstartTemplate("kind"), ""); err != nil {
		t.Fatalf("save valid stack: %v", err)
	}
	if err := runValidate(nil, nil); err != nil {
		t.Fatalf("runValidate() valid config error = %v", err)
	}

	if err := os.WriteFile(stackpkg.DefaultStackFile, []byte("version: ["), 0o600); err != nil {
		t.Fatalf("write invalid stack: %v", err)
	}
	if err := runValidate(nil, nil); err == nil {
		t.Fatal("runValidate() accepted invalid YAML")
	}
}

func TestShowDashboardStatusWithoutExternalServices(t *testing.T) {
	t.Setenv("PATH", "")

	tests := []*stackpkg.StackConfig{
		{Dashboard: &stackpkg.DashboardConfig{Runtime: "kubernetes"}},
		{Dashboard: &stackpkg.DashboardConfig{Runtime: "docker"}},
		{Dashboard: &stackpkg.DashboardConfig{
			Runtime: "pm2",
			PM2:     &stackpkg.PM2Config{AppName: "weave-dashboard"},
		}},
	}
	for _, config := range tests {
		showDashboardStatus(config)
	}
}
