// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package doctor

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/config"
)

// ---------------------------------------------------------------------------
// Status
// ---------------------------------------------------------------------------

func TestStatusString(t *testing.T) {
	tests := []struct {
		status Status
		want   string
	}{
		{StatusOK, "OK"},
		{StatusWarn, "WARN"},
		{StatusFail, "FAIL"},
		{StatusSkip, "SKIP"},
		{Status(99), "UNKNOWN"},
	}
	for _, tt := range tests {
		if got := tt.status.String(); got != tt.want {
			t.Errorf("Status(%d).String() = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestStatusMarshalJSON(t *testing.T) {
	tests := []struct {
		status Status
		want   string
	}{
		{StatusOK, `"OK"`},
		{StatusWarn, `"WARN"`},
		{StatusFail, `"FAIL"`},
		{StatusSkip, `"SKIP"`},
	}
	for _, tt := range tests {
		b, err := json.Marshal(tt.status)
		if err != nil {
			t.Fatalf("Marshal(%d): %v", tt.status, err)
		}
		if string(b) != tt.want {
			t.Errorf("Marshal(%d) = %s, want %s", tt.status, b, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// CheckResult JSON round-trip
// ---------------------------------------------------------------------------

func TestCheckResultJSON(t *testing.T) {
	r := CheckResult{
		Section: SectionVDB,
		Name:    "milvus-local",
		Status:  StatusFail,
		Message: "connection refused",
		Fix:     "run weave stack up",
	}

	b, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	// Verify status is a string in JSON
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("Unmarshal to map: %v", err)
	}
	if m["status"] != "FAIL" {
		t.Errorf("status = %v, want \"FAIL\"", m["status"])
	}
	if m["fix"] != "run weave stack up" {
		t.Errorf("fix = %v, want \"run weave stack up\"", m["fix"])
	}
}

// ---------------------------------------------------------------------------
// Output helpers
// ---------------------------------------------------------------------------

func TestHasFailures(t *testing.T) {
	tests := []struct {
		name    string
		results []CheckResult
		want    bool
	}{
		{"empty", nil, false},
		{"all ok", []CheckResult{{Status: StatusOK}, {Status: StatusWarn}}, false},
		{"has fail", []CheckResult{{Status: StatusOK}, {Status: StatusFail}}, true},
		{"only skip", []CheckResult{{Status: StatusSkip}}, false},
	}
	for _, tt := range tests {
		if got := HasFailures(tt.results); got != tt.want {
			t.Errorf("HasFailures(%s) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestCountByStatus(t *testing.T) {
	results := []CheckResult{
		{Status: StatusOK},
		{Status: StatusOK},
		{Status: StatusWarn},
		{Status: StatusFail},
		{Status: StatusSkip},
		{Status: StatusSkip},
	}
	ok, warn, fail, skip := countByStatus(results)
	if ok != 2 || warn != 1 || fail != 1 || skip != 2 {
		t.Errorf("countByStatus = (%d,%d,%d,%d), want (2,1,1,2)", ok, warn, fail, skip)
	}
}

func TestSectionTitle(t *testing.T) {
	tests := []struct {
		section string
		want    string
	}{
		{SectionSystem, "System"},
		{SectionConfig, "Config"},
		{SectionEnv, "Environment"},
		{SectionVDB, "VDB Connectivity"},
		{SectionLLM, "LLM"},
		{SectionEmbeddings, "Embeddings"},
		{SectionStack, "Stack"},
		{SectionOpik, "Opik"},
		{"custom", "custom"},
	}
	for _, tt := range tests {
		if got := sectionTitle(tt.section); got != tt.want {
			t.Errorf("sectionTitle(%q) = %q, want %q", tt.section, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// CheckSystem (unit-testable parts)
// ---------------------------------------------------------------------------

func TestCheckSystem_AlwaysHasCLIAndGoVersion(t *testing.T) {
	results := CheckSystem()
	if len(results) < 3 {
		t.Fatalf("CheckSystem returned %d results, want at least 3", len(results))
	}

	// First three are always CLI version, Go version, OS/Arch
	for _, r := range results[:3] {
		if r.Section != SectionSystem {
			t.Errorf("result %q section = %q, want %q", r.Name, r.Section, SectionSystem)
		}
		if r.Status != StatusOK {
			t.Errorf("result %q status = %v, want OK", r.Name, r.Status)
		}
	}

	if results[0].Name != "CLI version" {
		t.Errorf("results[0].Name = %q, want \"CLI version\"", results[0].Name)
	}
	if results[1].Name != "Go version" {
		t.Errorf("results[1].Name = %q, want \"Go version\"", results[1].Name)
	}
	if results[2].Name != "OS/Arch" {
		t.Errorf("results[2].Name = %q, want \"OS/Arch\"", results[2].Name)
	}
}

// ---------------------------------------------------------------------------
// CheckConfig
// ---------------------------------------------------------------------------

func TestCheckConfig_NilConfig(t *testing.T) {
	results := CheckConfig(nil, nil)
	// Should report config file not found (we're likely not in a dir with config.yaml)
	if len(results) == 0 {
		t.Fatal("CheckConfig(nil, nil) returned no results")
	}
	// First result should be about config file
	if results[0].Section != SectionConfig {
		t.Errorf("section = %q, want %q", results[0].Section, SectionConfig)
	}
}

func TestCheckConfig_WithValidConfig(t *testing.T) {
	cfg := &config.Config{
		Databases: config.DatabasesConfig{
			VectorDatabases: []config.VectorDBConfig{
				{
					Name:    "mock-db",
					Type:    config.VectorDBTypeMock,
					Enabled: true,
				},
			},
		},
	}
	// Will still fail on "config file not found" if not in project root,
	// but the validation portion should run if we pass no loadErr
	results := CheckConfig(cfg, nil)
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	for _, r := range results {
		if r.Section != SectionConfig {
			t.Errorf("section = %q, want %q", r.Section, SectionConfig)
		}
	}
}

// ---------------------------------------------------------------------------
// CheckEnv
// ---------------------------------------------------------------------------

func TestCheckEnv_NilConfig(t *testing.T) {
	results := CheckEnv(nil)
	// Should still check OPENAI_API_KEY and OPIK_* at minimum
	if len(results) < 1 {
		t.Fatal("CheckEnv(nil) returned no results")
	}
	for _, r := range results {
		if r.Section != SectionEnv {
			t.Errorf("section = %q, want %q", r.Section, SectionEnv)
		}
	}
}

func TestCheckEnv_WithWeaviateConfig(t *testing.T) {
	cfg := &config.Config{
		Databases: config.DatabasesConfig{
			VectorDatabases: []config.VectorDBConfig{
				{
					Name: "weaviate-cloud",
					Type: config.VectorDBTypeCloud,
				},
			},
		},
	}
	results := CheckEnv(cfg)
	// Should include WEAVIATE_URL and WEAVIATE_API_KEY checks
	found := map[string]bool{}
	for _, r := range results {
		found[r.Name] = true
	}
	for _, want := range []string{"WEAVIATE_URL", "WEAVIATE_API_KEY"} {
		if !found[want] {
			t.Errorf("expected check for %s, not found", want)
		}
	}
}

func TestIsPlaceholder(t *testing.T) {
	tests := []struct {
		val  string
		want bool
	}{
		{"sk-abc123realkey", false},
		{"your_api_key_here", true},
		{"changeme", true},
		{"sk-xxx", true},
		{"TODO", true},
		{"https://my-cluster.weaviate.cloud", false},
		{"placeholder-value", true},
		{"REPLACE_ME", true},
	}
	for _, tt := range tests {
		if got := isPlaceholder(tt.val); got != tt.want {
			t.Errorf("isPlaceholder(%q) = %v, want %v", tt.val, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// CheckVDB
// ---------------------------------------------------------------------------

func TestCheckVDB_NilConfig(t *testing.T) {
	results := CheckVDB(context.Background(), nil)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != StatusSkip {
		t.Errorf("status = %v, want SKIP", results[0].Status)
	}
}

func TestCheckVDB_MockDB(t *testing.T) {
	cfg := &config.Config{
		Databases: config.DatabasesConfig{
			VectorDatabases: []config.VectorDBConfig{
				{
					Name:    "test-db",
					Type:    "mock",
					Enabled: true,
				},
			},
		},
	}
	results := CheckVDB(context.Background(), cfg)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	// Mock DB should connect successfully
	if results[0].Section != SectionVDB {
		t.Errorf("section = %q, want %q", results[0].Section, SectionVDB)
	}
}

// ---------------------------------------------------------------------------
// CheckEmbeddings
// ---------------------------------------------------------------------------

func TestCheckEmbeddings_NilConfig(t *testing.T) {
	results := CheckEmbeddings(nil)
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	// All should be in embeddings section
	for _, r := range results {
		if r.Section != SectionEmbeddings {
			t.Errorf("section = %q, want %q", r.Section, SectionEmbeddings)
		}
	}
}

// ---------------------------------------------------------------------------
// CheckStack
// ---------------------------------------------------------------------------

func TestCheckStack_NoStackFile(t *testing.T) {
	// When run outside a project dir with weave-stack.yaml, should skip
	results := CheckStack()
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	// First result should be about stack config
	if results[0].Section != SectionStack {
		t.Errorf("section = %q, want %q", results[0].Section, SectionStack)
	}
}

// ---------------------------------------------------------------------------
// CheckLLM
// ---------------------------------------------------------------------------

func TestCheckLLM_NoAPIKey(t *testing.T) {
	// Save and clear env
	t.Setenv("OPENAI_API_KEY", "")

	results := CheckLLM(context.Background())
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != StatusSkip {
		t.Errorf("status = %v, want SKIP", results[0].Status)
	}
}

func TestCheckLLM_ShortKey(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "short")

	results := CheckLLM(context.Background())
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != StatusWarn {
		t.Errorf("status = %v, want WARN", results[0].Status)
	}
}

// ---------------------------------------------------------------------------
// CheckOpik
// ---------------------------------------------------------------------------

func TestCheckOpik_NoAPIKey(t *testing.T) {
	t.Setenv("OPIK_API_KEY", "")

	results := CheckOpik(context.Background())
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != StatusSkip {
		t.Errorf("status = %v, want SKIP", results[0].Status)
	}
}

// ---------------------------------------------------------------------------
// Doctor orchestrator
// ---------------------------------------------------------------------------

func TestDoctor_SectionFilter(t *testing.T) {
	dcfg := DoctorConfig{Section: SectionSystem}
	d := New(nil, nil, dcfg)
	results := d.RunAll(context.Background())

	for _, r := range results {
		if r.Section != SectionSystem {
			t.Errorf("with --section system, got result in section %q", r.Section)
		}
	}
}

func TestDoctor_AllSections(t *testing.T) {
	dcfg := DoctorConfig{}
	d := New(nil, nil, dcfg)
	results := d.RunAll(context.Background())

	sections := map[string]bool{}
	for _, r := range results {
		sections[r.Section] = true
	}

	// Should have results from at least system, config, env, vdb, llm, embeddings, stack, opik
	for _, want := range sectionOrder {
		if !sections[want] {
			t.Errorf("missing section %q in results", want)
		}
	}
}
