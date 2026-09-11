// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package utils

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb/weaviate"
)

type failingExportValue struct{}

func (failingExportValue) MarshalYAML() (interface{}, error) {
	return nil, errors.New("cannot marshal")
}

func (failingExportValue) MarshalJSON() ([]byte, error) {
	return nil, errors.New("cannot marshal")
}

func TestCompactSchema(t *testing.T) {
	if compactSchema(nil) != nil {
		t.Fatal("compactSchema(nil) was non-nil")
	}
	original := &weaviate.CollectionSchema{
		Class: "Docs", Vectorizer: "text2vec",
		Properties: []weaviate.SchemaProperty{{
			Name: "metadata", DataType: []string{"object"}, Description: "metadata",
			NestedProperties: []weaviate.SchemaProperty{{Name: "author", DataType: []string{"text"}}},
		}, {Name: "content", DataType: []string{"text"}}},
	}
	got := compactSchema(original)
	if got == original || got.Class != "Docs" || len(got.Properties) != 2 {
		t.Fatalf("compactSchema() = %#v", got)
	}
	if len(got.Properties[0].NestedProperties) != 1 || got.Properties[1].NestedProperties != nil {
		t.Fatalf("compacted properties = %#v", got.Properties)
	}
}

func TestSchemaExportRoundTrips(t *testing.T) {
	export := &CollectionExport{
		Name: "Docs",
		Schema: &weaviate.CollectionSchema{Class: "Docs", Properties: []weaviate.SchemaProperty{
			{Name: "values", DataType: []string{"number[]"}},
			{Name: "nested", DataType: []string{"object"}, NestedProperties: []weaviate.SchemaProperty{{Name: "flags", DataType: []string{"boolean[]"}}}},
		}},
	}
	yamlData, err := ExportAsYAML(export)
	if err != nil || !strings.HasPrefix(yamlData, "---\n") || !strings.Contains(yamlData, "datatype: [number[]]") {
		t.Fatalf("ExportAsYAML() = %q, %v", yamlData, err)
	}
	jsonData, err := ExportAsJSON(export)
	if err != nil || !strings.Contains(jsonData, `"name": "Docs"`) {
		t.Fatalf("ExportAsJSON() = %q, %v", jsonData, err)
	}

	root := t.TempDir()
	yamlPath := filepath.Join(root, "schema.yaml")
	if err := WriteToFile(yamlPath, yamlData); err != nil {
		t.Fatalf("WriteToFile(YAML): %v", err)
	}
	loadedYAML, err := LoadSchemaFromYAMLFile(yamlPath)
	if err != nil || loadedYAML.Name != "Docs" || loadedYAML.Schema.Properties[0].DataType[0] != "number[]" {
		t.Fatalf("LoadSchemaFromYAMLFile() = %#v, %v", loadedYAML, err)
	}
	jsonPath := filepath.Join(root, "schema.json")
	if err := WriteToFile(jsonPath, jsonData); err != nil {
		t.Fatalf("WriteToFile(JSON): %v", err)
	}
	loadedJSON, err := LoadSchemaFromJSONFile(jsonPath)
	if err != nil || loadedJSON.Name != "Docs" {
		t.Fatalf("LoadSchemaFromJSONFile() = %#v, %v", loadedJSON, err)
	}

	if err := WriteToFile(filepath.Join(root, "missing", "file"), "x"); err == nil {
		t.Fatal("WriteToFile() accepted a missing parent")
	}
	if _, err := LoadSchemaFromYAMLFile(filepath.Join(root, "missing.yaml")); err == nil {
		t.Fatal("LoadSchemaFromYAMLFile() accepted a missing file")
	}
	if _, err := LoadSchemaFromJSONFile(filepath.Join(root, "missing.json")); err == nil {
		t.Fatal("LoadSchemaFromJSONFile() accepted a missing file")
	}
	invalidYAML := filepath.Join(root, "invalid.yaml")
	if err := os.WriteFile(invalidYAML, []byte("name: ["), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadSchemaFromYAMLFile(invalidYAML); err == nil {
		t.Fatal("LoadSchemaFromYAMLFile() accepted invalid YAML")
	}
	invalidJSON := filepath.Join(root, "invalid.json")
	if err := os.WriteFile(invalidJSON, []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadSchemaFromJSONFile(invalidJSON); err == nil {
		t.Fatal("LoadSchemaFromJSONFile() accepted invalid JSON")
	}
	bad := &CollectionExport{Metadata: map[string]MetadataFieldInfo{"bad": {Sample: failingExportValue{}}}}
	if _, err := ExportAsYAML(bad); err == nil {
		t.Fatal("ExportAsYAML() accepted an unsupported value")
	}
	if _, err := ExportAsJSON(bad); err == nil {
		t.Fatal("ExportAsJSON() accepted an unsupported value")
	}
}

func TestYAMLFormattingHelpers(t *testing.T) {
	input := "properties:\n  - datatype:\n      - text\n      - number[]\nmetadata:\n  title:\n    type: string\n  data:\n    type: json\n    json_schema:\n      name: string\n      nested:\n        enabled: boolean\n"
	compacted := compactYAMLArrays(input)
	if !strings.Contains(compacted, "datatype: [text, number[]]") || !strings.Contains(compacted, "title: string") {
		t.Fatalf("compactYAMLArrays() = %q", compacted)
	}
	fixed := fixJSONSchemaIndentation(compacted)
	if !strings.Contains(fixed, "json_schema:\n        name: string") {
		t.Fatalf("fixJSONSchemaIndentation() = %q", fixed)
	}
	for _, arrayType := range []string{"number[]", "text[]", "int[]", "boolean[]"} {
		got := compactYAMLArrays("datatype: [" + arrayType + "]")
		if !strings.Contains(got, `"`+arrayType+`"`) {
			t.Errorf("compactYAMLArrays(%q) = %q", arrayType, got)
		}
	}
}

func TestJSONInference(t *testing.T) {
	schema, ok := inferJSONStructure(`{"null":null,"bool":true,"integer":2,"number":2.5,"string":"x","array":[1],"empty":[],"object":{"x":1}}`)
	if !ok {
		t.Fatal("inferJSONStructure() rejected JSON object")
	}
	for key, want := range map[string]string{
		"null": "null", "bool": "boolean", "integer": "integer", "number": "number", "string": "string",
		"array": "array[integer]", "empty": "array", "object": "object",
	} {
		if got := schema[key]; got != want {
			t.Errorf("schema[%q] = %#v, want %q", key, got, want)
		}
	}
	if _, ok := inferJSONStructure("not json"); ok {
		t.Fatal("inferJSONStructure() accepted invalid JSON")
	}
	if _, ok := inferJSONStructure(42); ok {
		t.Fatal("inferJSONStructure() accepted a non-string")
	}
	if got := inferValueType(int32(3)); got != "int32" {
		t.Fatalf("inferValueType(int32) = %q", got)
	}
}
