// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	pkgconfig "github.com/maximilien/weave-cli/src/pkg/config"
)

func TestEnvironmentVariableFiltering(t *testing.T) {
	all := getEnvVariables("")
	if len(all) < 20 {
		t.Fatalf("getEnvVariables() returned only %d variables", len(all))
	}
	filtered := getEnvVariables("milvus-cloud")
	if len(filtered) == 0 {
		t.Fatal("getEnvVariables(milvus-cloud) returned no variables")
	}
	for _, variable := range filtered {
		if len(variable.VDBTypes) > 0 && !containsString(variable.VDBTypes, "milvus-cloud") {
			t.Fatalf("unrelated variable returned: %#v", variable)
		}
	}
	if !containsString([]string{"one", "two"}, "two") {
		t.Fatal("containsString() missed value")
	}
	if containsString([]string{"one"}, "missing") {
		t.Fatal("containsString() found missing value")
	}
}

func TestEnvironmentFileLifecycle(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	envPath := filepath.Join(root, ".env")
	initial := "# comment\n# API_KEY=placeholder\nUNCHANGED=original\nUPDATE=old\ninvalid line\n\n"
	if err := os.WriteFile(envPath, []byte(initial), 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	values := loadEnvFile(envPath)
	if values["UNCHANGED"] != "original" || values["UPDATE"] != "old" {
		t.Fatalf("loadEnvFile() = %#v", values)
	}
	if values := loadEnvFile(filepath.Join(root, "missing")); len(values) != 0 {
		t.Fatalf("loadEnvFile(missing) = %#v", values)
	}

	updates := map[string]string{
		"API_KEY": "secret",
		"UPDATE":  "new",
		"NEW_KEY": "new-value",
		"EMPTY":   "",
	}
	if err := saveEnvFile(envPath, updates); err != nil {
		t.Fatalf("saveEnvFile() error: %v", err)
	}
	data, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("read saved env: %v", err)
	}
	text := string(data)
	for _, expected := range []string{`API_KEY="secret"`, `UPDATE="new"`, `NEW_KEY="new-value"`, "UNCHANGED=original", "invalid line"} {
		if !strings.Contains(text, expected) {
			t.Errorf("saved env missing %q: %s", expected, text)
		}
	}

	example := "# FROM_EXAMPLE=value\n"
	if err := os.WriteFile(".env.example", []byte(example), 0o600); err != nil {
		t.Fatalf("write env example: %v", err)
	}
	secondPath := filepath.Join(root, "created.env")
	if err := saveEnvFile(secondPath, map[string]string{"FROM_EXAMPLE": "configured"}); err != nil {
		t.Fatalf("saveEnvFile(fallback) error: %v", err)
	}
	if got := loadEnvFile(secondPath)["FROM_EXAMPLE"]; got != "configured" {
		t.Fatalf("fallback env value = %q", got)
	}

	if err := os.Remove(".env.example"); err != nil {
		t.Fatalf("remove env example: %v", err)
	}
	if err := saveEnvFile(filepath.Join(root, "missing-source.env"), nil); err == nil {
		t.Fatal("saveEnvFile() succeeded without source")
	}
}

func TestCopyFilePreservesContentAndPermissions(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	destination := filepath.Join(root, "destination")
	if err := os.WriteFile(source, []byte("content"), 0o640); err != nil {
		t.Fatalf("write source: %v", err)
	}
	if err := copyFile(source, destination); err != nil {
		t.Fatalf("copyFile() error: %v", err)
	}
	data, err := os.ReadFile(destination)
	if err != nil || string(data) != "content" {
		t.Fatalf("copied content = %q, %v", data, err)
	}
	info, err := os.Stat(destination)
	if err != nil || info.Mode().Perm() != 0o640 {
		t.Fatalf("copied permissions = %v, %v", info.Mode().Perm(), err)
	}
	if err := copyFile(filepath.Join(root, "missing"), destination); err == nil {
		t.Fatal("copyFile() accepted missing source")
	}
}

func TestInputAndConfirmationHelpers(t *testing.T) {
	withConfigStdin(t, " value \n", func() {
		if got, err := readLine(); err != nil || got != "value" {
			t.Fatalf("readLine() = %q, %v", got, err)
		}
	})
	withConfigStdin(t, "secret\n", func() {
		if got, err := readSecret(); err != nil || got != "secret" {
			t.Fatalf("readSecret() = %q, %v", got, err)
		}
	})
	for _, test := range []struct {
		input string
		want  bool
	}{
		{input: "\n", want: true},
		{input: "YES\n", want: true},
		{input: "n\n", want: false},
	} {
		withConfigStdin(t, test.input, func() {
			if got := confirmSave("config.env"); got != test.want {
				t.Fatalf("confirmSave(%q) = %t, want %t", test.input, got, test.want)
			}
		})
	}
}

func TestDatabaseAndSchemaDisplayHelpers(t *testing.T) {
	collection := pkgconfig.Collection{Name: "docs", Type: "text", Description: "documents"}
	configs := []*pkgconfig.VectorDBConfig{
		{Type: pkgconfig.VectorDBTypeCloud, URL: "https://weaviate.example", APIKey: "secret", Collections: []pkgconfig.Collection{collection}},
		{Type: pkgconfig.VectorDBTypeCloud, URL: "https://weaviate.example"},
		{Type: pkgconfig.VectorDBTypeLocal, URL: "http://localhost:8080", Collections: []pkgconfig.Collection{collection}},
		{Type: pkgconfig.VectorDBTypeSupabaseCloud, DatabaseURL: "postgres://username:password@database.example:5432/postgres", Collections: []pkgconfig.Collection{collection}},
		{Type: pkgconfig.VectorDBTypeMongoDBCloud, URL: "mongodb+srv://username:password@cluster.example/database", Database: "weave", Collections: []pkgconfig.Collection{collection}},
		{Type: pkgconfig.VectorDBTypeMilvusLocal, Address: "localhost:19530", Database: "default", VectorDimensions: 384, SimilarityMetric: "cosine", Collections: []pkgconfig.Collection{collection}},
		{Type: pkgconfig.VectorDBTypeMilvusCloud, Address: "cloud.example", Username: "user", Password: "secret", Database: "default", VectorDimensions: 384, SimilarityMetric: "cosine", Collections: []pkgconfig.Collection{collection}},
		{Type: pkgconfig.VectorDBTypeMilvusCloud, Address: "cloud.example"},
		{Type: pkgconfig.VectorDBTypeMock, Enabled: true, SimulateEmbeddings: true, EmbeddingDimension: 384, Collections: []pkgconfig.Collection{collection}},
	}
	for _, cfg := range configs {
		displayDatabaseConfig("database", cfg)
	}

	if got := maskConnectionString("short"); got != "***hidden***" {
		t.Fatalf("maskConnectionString(short) = %q", got)
	}
	if got := maskConnectionString("postgres://username:password@database.example/db"); !strings.Contains(got, "...") || strings.Contains(got, "password") {
		t.Fatalf("maskConnectionString(long) = %q", got)
	}
	displayJSONSchemaFields(map[string]interface{}{"type": "string"}, "  ")
	printError("error")
	printWarning("warning")
	printHeader("header")

	schema := &pkgconfig.SchemaDefinition{
		Name: "docs",
		Schema: map[string]interface{}{
			"class": "Docs",
			"properties": []interface{}{map[string]interface{}{
				"name":        "metadata",
				"json_schema": map[string]interface{}{"type": "object"},
			}},
		},
	}
	outputSchemaAsYAML(schema)
	outputSchemaAsJSON(schema)
	fixed := fixSchemaJSONIndentation("property:\n  json_schema:\n    type: object\n  name: metadata\n")
	if !strings.Contains(fixed, "      type: object") {
		t.Fatalf("fixSchemaJSONIndentation() = %q", fixed)
	}
}

func withConfigStdin(t *testing.T, input string, run func()) {
	t.Helper()
	readEnd, writeEnd, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdin pipe: %v", err)
	}
	oldStdin := os.Stdin
	os.Stdin = readEnd
	t.Cleanup(func() {
		os.Stdin = oldStdin
		_ = readEnd.Close()
	})
	if _, err := writeEnd.WriteString(input); err != nil {
		t.Fatalf("write stdin: %v", err)
	}
	if err := writeEnd.Close(); err != nil {
		t.Fatalf("close stdin writer: %v", err)
	}
	run()
}
