// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

const nestedEditorFixture = `# Preserve this configuration comment
name: fixture
groups:
  - members:
      - name: first
        enabled: true
      - name: second
        enabled: true
`

func newNestedEditor(t *testing.T) (*YAMLEditor, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(nestedEditorFixture), 0600); err != nil {
		t.Fatal(err)
	}
	editor, err := NewYAMLEditor(path)
	if err != nil {
		t.Fatal(err)
	}
	return editor, path
}

func TestYAMLEditorInvalidPathsPreserveDocument(t *testing.T) {
	for _, operation := range []string{"set", "remove", "disable"} {
		for _, tc := range []struct{ path, want string }{
			{"groups[bad]", "invalid path"},
			{"missing.members", "not found"},
			{"name.members", "expected mapping"},
			{"name[0]", "expected sequence"},
			{"groups[9].members", "out of bounds"},
			{"groups[-1].members", "out of bounds"},
		} {
			t.Run(operation+"/"+tc.path, func(t *testing.T) {
				editor, path := newNestedEditor(t)
				before, err := yaml.Marshal(editor.root)
				if err != nil {
					t.Fatal(err)
				}
				switch operation {
				case "set":
					err = editor.SetValue(tc.path, "changed")
				case "remove":
					err = editor.RemoveArrayElement(tc.path, 0)
				case "disable":
					err = editor.DisableArrayElement(tc.path, 0)
				}
				if err == nil || !strings.Contains(err.Error(), tc.want) {
					t.Fatalf("error = %v; want %s", err, tc.want)
				}
				after, err := yaml.Marshal(editor.root)
				if err != nil {
					t.Fatal(err)
				}
				if string(before) != string(after) {
					t.Fatalf("invalid edit changed document: %s", after)
				}
				disk, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				if string(disk) != nestedEditorFixture {
					t.Fatal("invalid edit modified file")
				}
			})
		}
	}
}

func TestYAMLEditorNestedArrayEditsRoundTrip(t *testing.T) {
	editor, path := newNestedEditor(t)
	if err := editor.SetValue("groups[0].members[1].url", "https://example.invalid/db"); err != nil {
		t.Fatal(err)
	}
	if err := editor.DisableArrayElement("groups[0].members", 1); err != nil {
		t.Fatal(err)
	}
	if err := editor.RemoveArrayElement("groups[0].members", 0); err != nil {
		t.Fatal(err)
	}
	if err := editor.Save(); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "Preserve this configuration comment") || !strings.Contains(string(raw), "DISABLED") {
		t.Fatalf("comments lost: %s", raw)
	}
	var got struct {
		Name   string
		Groups []struct{ Members []map[string]interface{} }
	}
	if err := yaml.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	want := map[string]interface{}{"name": "second", "enabled": false, "url": "https://example.invalid/db"}
	if got.Name != "fixture" || len(got.Groups) != 1 || len(got.Groups[0].Members) != 1 || !reflect.DeepEqual(got.Groups[0].Members[0], want) {
		t.Fatalf("edited document = %#v", got)
	}
}

func TestYAMLEditorScalarAndSequenceBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name string
		edit func(*YAMLEditor) error
		want string
	}{
		{"set mapping", func(e *YAMLEditor) error { return e.SetValue("groups[0]", "bad") }, "non-scalar"},
		{"set missing index", func(e *YAMLEditor) error { return e.SetValue("groups[3]", "bad") }, "out of bounds"},
		{"remove scalar", func(e *YAMLEditor) error { return e.RemoveArrayElement("name", 0) }, "does not point to an array"},
		{"disable scalar", func(e *YAMLEditor) error { return e.DisableArrayElement("name", 0) }, "does not point to an array"},
		{"remove missing index", func(e *YAMLEditor) error { return e.RemoveArrayElement("groups[0].members", 3) }, "out of bounds"},
		{"disable negative index", func(e *YAMLEditor) error { return e.DisableArrayElement("groups[0].members", -1) }, "out of bounds"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			editor, _ := newNestedEditor(t)
			if err := tc.edit(editor); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v; want %s", err, tc.want)
			}
		})
	}
}
