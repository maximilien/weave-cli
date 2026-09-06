// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package repl

import (
	"reflect"
	"sort"
	"testing"
)

func TestCompleterSuggestions(t *testing.T) {
	completer := NewCompleter()
	tests := []struct {
		name       string
		line       string
		want       []string
		wantLength int
	}{
		{name: "all commands", want: []string{"/clear", "/col", "/collection", "/examples", "/exit", "/help", "/history", "/mcp", "/quit", "/search", "/stats", "/status", "/use", "clear", "exit", "help", "quit"}},
		{name: "command prefix", line: "/sta", want: []string{"ts", "tus"}, wantLength: 4},
		{name: "subcommand prefix", line: "/mcp c", want: []string{"all"}, wantLength: 1},
		{name: "all subcommands", line: "/collection ", want: []string{"create", "delete", "help", "info", "list", "use"}},
		{name: "unknown command", line: "/missing ", want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRunes, gotLength := completer.Do([]rune(tt.line), len([]rune(tt.line)))
			got := runeStrings(gotRunes)
			sort.Strings(got)
			sort.Strings(tt.want)
			if !reflect.DeepEqual(got, tt.want) || gotLength != tt.wantLength {
				t.Fatalf("Do(%q) = %#v, %d; want %#v, %d", tt.line, got, gotLength, tt.want, tt.wantLength)
			}
		})
	}
}

func TestReadlineConfigAndPathExpansion(t *testing.T) {
	t.Setenv("HOME", "/tmp/weave-home")
	config := CreateReadlineConfig()
	if config.HistoryFile != "/tmp/weave-home/.weave_history" || config.AutoComplete == nil {
		t.Fatalf("CreateReadlineConfig() = %#v", config)
	}
	if got, ok := config.FuncFilterInputRune('x'); got != 'x' || !ok {
		t.Fatalf("input filter = %q, %t", got, ok)
	}
	if got := expandPath("/tmp/data"); got != "/tmp/data" {
		t.Fatalf("expandPath() = %q", got)
	}

	t.Setenv("HOME", "")
	if got := homeDir(); got != "." {
		t.Fatalf("homeDir() = %q", got)
	}
}

func runeStrings(values [][]rune) []string {
	if values == nil {
		return nil
	}
	result := make([]string, len(values))
	for i, value := range values {
		result[i] = string(value)
	}
	return result
}
