// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestParsePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected []pathSegment
		wantErr  bool
	}{
		{
			name: "simple key path",
			path: "databases.default",
			expected: []pathSegment{
				{Type: segmentTypeKey, Value: "databases"},
				{Type: segmentTypeKey, Value: "default"},
			},
			wantErr: false,
		},
		{
			name: "path with array index",
			path: "databases.vector_databases[2].database_url",
			expected: []pathSegment{
				{Type: segmentTypeKey, Value: "databases"},
				{Type: segmentTypeKey, Value: "vector_databases"},
				{Type: segmentTypeIndex, Index: 2},
				{Type: segmentTypeKey, Value: "database_url"},
			},
			wantErr: false,
		},
		{
			name: "path with database name annotation",
			path: "databases.vector_databases[4] (milvus-cloud).api_key",
			expected: []pathSegment{
				{Type: segmentTypeKey, Value: "databases"},
				{Type: segmentTypeKey, Value: "vector_databases"},
				{Type: segmentTypeIndex, Index: 4},
				{Type: segmentTypeKey, Value: "api_key"},
			},
			wantErr: false,
		},
		{
			name: "nested keys",
			path: "ai.model",
			expected: []pathSegment{
				{Type: segmentTypeKey, Value: "ai"},
				{Type: segmentTypeKey, Value: "model"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			segments, err := parsePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if len(segments) != len(tt.expected) {
				t.Errorf("Expected %d segments, got %d", len(tt.expected), len(segments))
				return
			}

			for i, seg := range segments {
				exp := tt.expected[i]
				if seg.Type != exp.Type {
					t.Errorf("Segment %d: type = %v, want %v", i, seg.Type, exp.Type)
				}
				if seg.Type == segmentTypeKey && seg.Value != exp.Value {
					t.Errorf("Segment %d: value = %v, want %v", i, seg.Value, exp.Value)
				}
				if seg.Type == segmentTypeIndex && seg.Index != exp.Index {
					t.Errorf("Segment %d: index = %v, want %v", i, seg.Index, exp.Index)
				}
			}
		})
	}
}

func TestYAMLEditor_SetValue(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		path        string
		value       string
		checkPath   string
		expected    string
	}{
		{
			name: "set simple key",
			yamlContent: `
databases:
  default: mock
`,
			path:      "databases.default",
			value:     "weaviate-cloud",
			checkPath: "databases.default",
			expected:  "weaviate-cloud",
		},
		{
			name: "set nested key in array",
			yamlContent: `
databases:
  vector_databases:
    - name: mongodb-cloud
      type: mongodb
      database_url: ""
`,
			path:      "databases.vector_databases[0].database_url",
			value:     "mongodb+srv://test@cluster.mongodb.net/db",
			checkPath: "databases.vector_databases[0].database_url",
			expected:  "mongodb+srv://test@cluster.mongodb.net/db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "config.yaml")
			if err := os.WriteFile(tmpFile, []byte(tt.yamlContent), 0644); err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}

			// Create editor
			editor, err := NewYAMLEditor(tmpFile)
			if err != nil {
				t.Fatalf("NewYAMLEditor() error = %v", err)
			}

			// Set value
			if err := editor.SetValue(tt.path, tt.value); err != nil {
				t.Fatalf("SetValue() error = %v", err)
			}

			// Save
			if err := editor.Save(); err != nil {
				t.Fatalf("Save() error = %v", err)
			}

			// Verify by reading back
			data, err := os.ReadFile(tmpFile)
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			}

			var result map[string]interface{}
			if err := yaml.Unmarshal(data, &result); err != nil {
				t.Fatalf("Failed to parse YAML: %v", err)
			}

			// Navigate to check path and verify value
			value := getValueAtPath(result, tt.checkPath)
			if value != tt.expected {
				t.Errorf("Value at %s = %v, want %v", tt.checkPath, value, tt.expected)
			}
		})
	}
}

func TestYAMLEditor_RemoveArrayElement(t *testing.T) {
	yamlContent := `
databases:
  vector_databases:
    - name: db1
      type: mock
    - name: db2
      type: mock
    - name: db3
      type: mock
`

	// Create temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(tmpFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Create editor
	editor, err := NewYAMLEditor(tmpFile)
	if err != nil {
		t.Fatalf("NewYAMLEditor() error = %v", err)
	}

	// Remove middle element
	if err := editor.RemoveArrayElement("databases.vector_databases", 1); err != nil {
		t.Fatalf("RemoveArrayElement() error = %v", err)
	}

	// Save
	if err := editor.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	var result map[string]interface{}
	if err := yaml.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	databases := result["databases"].(map[string]interface{})
	vdbs := databases["vector_databases"].([]interface{})

	if len(vdbs) != 2 {
		t.Errorf("Expected 2 databases, got %d", len(vdbs))
	}

	// Check remaining databases
	db0 := vdbs[0].(map[string]interface{})
	if db0["name"] != "db1" {
		t.Errorf("First db name = %v, want db1", db0["name"])
	}

	db1 := vdbs[1].(map[string]interface{})
	if db1["name"] != "db3" {
		t.Errorf("Second db name = %v, want db3", db1["name"])
	}
}

func TestYAMLEditor_DisableArrayElement(t *testing.T) {
	yamlContent := `
databases:
  vector_databases:
    - name: db1
      type: mock
    - name: db2
      type: mock
`

	// Create temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(tmpFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Create editor
	editor, err := NewYAMLEditor(tmpFile)
	if err != nil {
		t.Fatalf("NewYAMLEditor() error = %v", err)
	}

	// Disable first element
	if err := editor.DisableArrayElement("databases.vector_databases", 0); err != nil {
		t.Fatalf("DisableArrayElement() error = %v", err)
	}

	// Save
	if err := editor.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	var result map[string]interface{}
	if err := yaml.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	databases := result["databases"].(map[string]interface{})
	vdbs := databases["vector_databases"].([]interface{})

	// Check that enabled field is set to false
	db0 := vdbs[0].(map[string]interface{})
	if enabled, ok := db0["enabled"]; !ok || enabled != false {
		t.Errorf("Expected enabled=false, got %v", enabled)
	}

	// Check that file content has DISABLED comment
	fileContent := string(data)
	if !strings.Contains(fileContent, "DISABLED") {
		t.Error("Expected file to contain DISABLED comment")
	}
}

// Helper function to get value at a path in a map
func getValueAtPath(data map[string]interface{}, path string) interface{} {
	segments, err := parsePath(path)
	if err != nil {
		return nil
	}

	var current interface{} = data
	for _, seg := range segments {
		switch seg.Type {
		case segmentTypeKey:
			if m, ok := current.(map[string]interface{}); ok {
				current = m[seg.Value]
			} else {
				return nil
			}
		case segmentTypeIndex:
			if arr, ok := current.([]interface{}); ok {
				if seg.Index >= 0 && seg.Index < len(arr) {
					current = arr[seg.Index]
				} else {
					return nil
				}
			} else {
				return nil
			}
		}
	}

	return current
}
