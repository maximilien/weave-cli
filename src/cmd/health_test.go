// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package cmd

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/spf13/viper"
)

func captureHealthOutput(t *testing.T, run func()) string {
	t.Helper()
	previousOut, previousErr := os.Stdout, os.Stderr
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout, os.Stderr = writer, writer
	run()
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout, os.Stderr = previousOut, previousErr
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if err := reader.Close(); err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func setupHealthConfig(t *testing.T) {
	t.Helper()
	root := t.TempDir()
	t.Chdir(root)
	t.Setenv("HOME", root)
	t.Setenv("WEAVE_SKIP_CONFIG_VALIDATION", "true")
	t.Setenv("VECTOR_DB_TYPE", "mock")
	t.Setenv("WEAVIATE_URL", "https://fixture.invalid")
	t.Setenv("WEAVIATE_API_KEY", "fixture")
	t.Setenv("OPENAI_API_KEY", "fixture")
	path := filepath.Join(root, "config.yaml")
	contents := `databases:
  default: fixture
  vector_databases:
    - name: fixture
      type: mock
      enabled: true
      simulate_embeddings: true
      embedding_dimension: 3
      collections:
        - name: Zeta
          type: text
        - name: Alpha
          type: text
`
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	viper.Reset()
	viper.Set("config", path)
	viper.Set("env", "")
	viper.Set("quiet", true)
	viper.Set("no-color", true)
	t.Cleanup(viper.Reset)
}

func TestCheckSingleDatabase(t *testing.T) {
	db := &config.VectorDBConfig{
		Name: "fixture", Type: config.VectorDBTypeMock, Enabled: true,
		EmbeddingDimension: 3,
		Collections:        []config.Collection{{Name: "Zeta"}, {Name: "Alpha"}},
	}
	result := checkSingleDatabase(context.Background(), db.Name, db)
	if !result.Healthy || result.CollectionsCount != 2 {
		t.Fatalf("mock health result = %#v", result)
	}
	if strings.Join(result.Collections, ",") != "Alpha,Zeta" {
		t.Fatalf("sorted collections = %#v", result.Collections)
	}
	if result.URL != "configured location" || !strings.Contains(result.Message, "Successfully connected") {
		t.Fatalf("mock health details = %#v", result)
	}

	invalid := checkSingleDatabase(context.Background(), "broken", &config.VectorDBConfig{Type: config.VectorDBType("invalid")})
	if invalid.Healthy || !strings.Contains(invalid.Message, "Failed to create database client") {
		t.Fatalf("invalid health result = %#v", invalid)
	}

	for _, test := range []struct {
		config config.VectorDBConfig
		want   string
	}{
		{config: config.VectorDBConfig{URL: "https://vector.example"}, want: "https://vector.example"},
		{config: config.VectorDBConfig{DatabaseURL: "postgres://database"}, want: "postgres://database"},
		{config: config.VectorDBConfig{}, want: "configured location"},
	} {
		if got := getDisplayURL(&test.config); got != test.want {
			t.Errorf("getDisplayURL() = %q, want %q", got, test.want)
		}
	}
}

func TestHealthOutputPaths(t *testing.T) {
	healthy := HealthCheckResult{
		DatabaseName: "healthy", DatabaseType: "mock", Healthy: true,
		Message: "connected", URL: "https://example.test/" + strings.Repeat("long", 20),
		Collections: []string{"Docs", "Images"}, CollectionsCount: 2,
	}
	unhealthy := HealthCheckResult{
		DatabaseName: "broken", DatabaseType: "invalid", Healthy: false,
		Message: "unavailable", URL: "configured location",
	}

	output := captureHealthOutput(t, func() {
		displayHealthCheckResult(healthy)
		displayHealthCheckResult(HealthCheckResult{DatabaseName: "empty", DatabaseType: "mock", Healthy: true, Message: "connected"})
		displayHealthCheckResult(unhealthy)
		outputHealthCheckJSON([]HealthCheckResult{healthy, unhealthy})
	})
	for _, want := range []string{"Database connection is healthy", "Found 2 collections", "No collections found", "Database connection failed", `"total": 2`} {
		if !strings.Contains(output, want) {
			t.Errorf("detailed output missing %q:\n%s", want, output)
		}
	}

	results := []HealthCheckResult{healthy, unhealthy, {
		DatabaseName: "another", DatabaseType: "mock", Healthy: true, URL: "configured location",
	}}
	output = captureHealthOutput(t, func() {
		displayHealthCheckSummary(append([]HealthCheckResult(nil), results...), "type")
		displayHealthCheckSummary(append([]HealthCheckResult(nil), results...), "name")
	})
	for _, want := range []string{"Vector Database Health Check Summary", "Total: 3 databases", "..."} {
		if !strings.Contains(output, want) {
			t.Errorf("summary output missing %q:\n%s", want, output)
		}
	}

	configs := []config.VectorDBConfig{
		{Name: "zeta", Type: config.VectorDBTypeMock, EmbeddingDimension: 3},
		{Name: "broken", Type: config.VectorDBType("invalid")},
		{Name: "alpha", Type: config.VectorDBTypeMock, EmbeddingDimension: 3, URL: "https://example.test/" + strings.Repeat("long", 20)},
	}
	output = captureHealthOutput(t, func() {
		displayHealthCheckSummaryProgressive(context.Background(), append([]config.VectorDBConfig(nil), configs...), "type")
		displayHealthCheckSummaryProgressive(context.Background(), append([]config.VectorDBConfig(nil), configs...), "name")
	})
	if !strings.Contains(output, "Total: 3 databases") || !strings.Contains(output, "...") {
		t.Fatalf("progressive output missing totals:\n%s", output)
	}
}

func TestRunHealthCheckMockPaths(t *testing.T) {
	setupHealthConfig(t)

	output := captureHealthOutput(t, func() {
		runHealthCheck(healthCheckCmd, []string{"fixture"})
	})
	if !strings.Contains(output, "Database Health Check: fixture") || !strings.Contains(output, "Found 2 collections") {
		t.Fatalf("specific database output:\n%s", output)
	}

	if err := healthCheckCmd.Flags().Set("summary", "true"); err != nil {
		t.Fatal(err)
	}
	output = captureHealthOutput(t, func() {
		runHealthCheck(healthCheckCmd, []string{"fixture"})
	})
	if !strings.Contains(output, "Health Check Summary") {
		t.Fatalf("summary command output:\n%s", output)
	}
	_ = healthCheckCmd.Flags().Set("summary", "false")

	t.Cleanup(func() {
		_ = healthCheckCmd.Flags().Set("summary", "false")
	})
}
