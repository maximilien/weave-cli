// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package document

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestPDFConversionSelectsAvailableTool(t *testing.T) {
	for _, directory := range []bool{false, true} {
		for _, flag := range []string{"", "rgb", "imagemagick"} {
			for _, ghostscript := range []bool{false, true} {
				t.Run(fmt.Sprintf("directory=%t/flag=%s/gs=%t", directory, flag, ghostscript), func(t *testing.T) {
					bin := t.TempDir()
					writeExecutable(t, filepath.Join(bin, "convert"), "#!/bin/sh\nfor arg in \"$@\"; do out=\"$arg\"; done\nprintf 'imagemagick' > \"$out\"\n")
					if ghostscript {
						writeExecutable(t, filepath.Join(bin, "gs"), "#!/bin/sh\nfor arg in \"$@\"; do case \"$arg\" in -sOutputFile=*) out=\"${arg#*=}\" ;; esac; done\nprintf 'ghostscript' > \"$out\"\n")
					}
					t.Setenv("PATH", bin)
					root := t.TempDir()
					input := filepath.Join(root, "input.pdf")
					if err := os.WriteFile(input, []byte("source PDF"), 0600); err != nil {
						t.Fatal(err)
					}
					cmd := newPDFConvertTestCommand()
					if flag != "" {
						if err := cmd.Flags().Set(flag, "true"); err != nil {
							t.Fatal(err)
						}
					}
					args := []string{input}
					if directory {
						if err := cmd.Flags().Set("directory", root); err != nil {
							t.Fatal(err)
						}
						args = nil
					}
					runPdfConvert(cmd, args)
					tool := ToolImageMagick
					want := "imagemagick"
					if ghostscript && flag != "imagemagick" {
						tool = ToolGhostscript
						want = "ghostscript"
					}
					data, err := os.ReadFile(generateOutputFilename(input, tool))
					if err != nil || string(data) != want {
						t.Fatalf("selected tool output=%q, error=%v; want %q", data, err, want)
					}
					original, err := os.ReadFile(input)
					if err != nil || string(original) != "source PDF" {
						t.Fatalf("source changed: %q, %v", original, err)
					}
				})
			}
		}
	}
}
