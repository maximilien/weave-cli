// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package document

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPDFDirectoryConversionContinuesAfterFailures(t *testing.T) {
	bin := t.TempDir()
	root := t.TempDir()
	writeExecutable(t, filepath.Join(bin, "gs"), `#!/bin/sh
for arg in "$@"; do
 case "$arg" in -sOutputFile=*) out="${arg#*=}" ;; esac
 input="$arg"
done
case "$input" in *rejected.pdf) exit 2 ;; *missing.pdf) exit 0 ;; esac
printf 'converted' > "$out"
`)
	t.Setenv("PATH", bin)
	for _, name := range []string{"rejected.pdf", "missing.pdf", "success.pdf"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("source"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	output, err := os.CreateTemp(t.TempDir(), "conversion-*.log")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = output.Close() })
	old := os.Stdout
	os.Stdout = output
	t.Cleanup(func() { os.Stdout = old })
	runDirectoryConvert(newPDFConvertTestCommand(), root, true, false, false, false)
	os.Stdout = old
	data, err := os.ReadFile(output.Name())
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Total:   3 files", "Success: 1 files", "Failed:  2 files"} {
		if !strings.Contains(string(data), want) {
			t.Fatalf("missing %q: %s", want, data)
		}
	}
	converted, err := os.ReadFile(filepath.Join(root, "success-rgb-ghostscript.pdf"))
	if err != nil || string(converted) != "converted" {
		t.Fatalf("successful conversion=%q %v", converted, err)
	}
	for _, name := range []string{"rejected", "missing"} {
		if _, err := os.Stat(filepath.Join(root, name+"-rgb-ghostscript.pdf")); !os.IsNotExist(err) {
			t.Fatalf("failed conversion unexpectedly created output: %s %v", name, err)
		}
	}
}
