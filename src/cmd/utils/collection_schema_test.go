// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package utils

import (
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/config"
)

func TestConvertSchemaDefinitionToCollectionSchema(t *testing.T) {
	definition := &config.SchemaDefinition{
		Schema: map[string]interface{}{
			"class":      "FromSchema",
			"vectorizer": "text2vec-openai",
			"properties": []interface{}{
				map[string]interface{}{
					"name":        "content",
					"datatype":    []interface{}{"text", 4},
					"description": "document content",
					"nestedProperties": []interface{}{
						map[string]interface{}{"name": "author", "datatype": []interface{}{"text"}, "description": "author name"},
						"invalid",
					},
				},
				map[string]interface{}{"name": "metadata", "datatype": []interface{}{"object"}},
				"invalid",
			},
		},
		Metadata: map[string]interface{}{
			"id":                "reserved",
			"content":           "already present",
			"added_date":        "2026-09-11",
			"creation_date":     "2026-09-10",
			"modified_date":     "2026-09-11",
			"creator":           "tester",
			"producer":          "weave",
			"title":             "Guide",
			"ai_summary":        "Summary",
			"filename":          "guide.txt",
			"is_chunked":        true,
			"total_chunks":      int64(3),
			"chunk_index":       1,
			"chunk_sizes":       []interface{}{100, 200},
			"original_filename": "original.txt",
			"storage_path":      "/tmp/guide.txt",
			"type":              "document",
			"string_array":      []interface{}{"a"},
			"boolean_array":     []interface{}{true},
			"unknown_array":     []interface{}{map[string]interface{}{}},
			"empty_array":       []interface{}{},
			"array_integer":     map[string]interface{}{"type": "array", "items": "integer"},
			"array_number":      map[string]interface{}{"type": "array", "items": "number"},
			"array_string":      map[string]interface{}{"type": "array", "items": "string"},
			"array_boolean":     map[string]interface{}{"type": "array", "items": "boolean"},
			"array_unknown":     map[string]interface{}{"type": "array", "items": "object"},
			"array_no_items":    map[string]interface{}{"type": "array"},
			"object":            map[string]interface{}{"type": "object"},
			"unknown":           struct{}{},
		},
	}

	schema, err := convertSchemaDefinitionToCollectionSchema(definition, "Override", "flat")
	if err != nil {
		t.Fatalf("convertSchemaDefinitionToCollectionSchema() error = %v", err)
	}
	if schema.Class != "Override" || schema.Vectorizer != "text2vec-openai" {
		t.Fatalf("schema identity = %#v", schema)
	}
	properties := make(map[string][]string, len(schema.Properties))
	for _, property := range schema.Properties {
		properties[property.Name] = property.DataType
	}
	if _, exists := properties["metadata"]; exists {
		t.Fatal("flat schema retained metadata object property")
	}
	if got := properties["content"]; len(got) != 2 || got[0] != "text" || got[1] != "" {
		t.Fatalf("content datatype = %#v", got)
	}
	for name, want := range map[string]string{
		"added_date": "string", "is_chunked": "boolean", "total_chunks": "number",
		"string_array": "string[]", "boolean_array": "boolean[]", "chunk_sizes": "number[]",
		"unknown_array": "string[]", "empty_array": "string[]", "array_integer": "number[]",
		"array_number": "number[]", "array_string": "string[]", "array_boolean": "boolean[]",
		"array_unknown": "string[]", "array_no_items": "string[]", "object": "string", "unknown": "string",
	} {
		if got := properties[name]; len(got) != 1 || got[0] != want {
			t.Errorf("property %q datatype = %#v, want %q", name, got, want)
		}
	}

	fromSchema, err := convertSchemaDefinitionToCollectionSchema(definition, "", "nested")
	if err != nil || fromSchema.Class != "FromSchema" {
		t.Fatalf("schema class fallback = %#v, %v", fromSchema, err)
	}
	if _, err := convertSchemaDefinitionToCollectionSchema(&config.SchemaDefinition{Schema: map[string]interface{}{}}, "", "nested"); err == nil {
		t.Fatal("missing collection name did not return an error")
	}
}
