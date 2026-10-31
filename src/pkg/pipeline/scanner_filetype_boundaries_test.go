// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package pipeline

import (
	"path/filepath"
	"testing"
)

func TestDetectFileTypeHandlesCaseAndAliases(t *testing.T) {
	for _, tc := range []struct {
		path string
		want FileType
	}{
		{"paper.PDF", FileTypePDF}, {"notes.TXT", FileTypeTXT}, {"guide.markdown", FileTypeMD}, {"config.JSON", FileTypeJSON}, {"config.YML", FileTypeYAML}, {"photo.JPEG", FileTypeImage}, {"data.bin", FileTypeUnknown},
	} {
		t.Run(tc.path, func(t *testing.T) {
			if got := detectFileType(tc.path); got != tc.want {
				t.Fatalf("type=%q, want %q", got, tc.want)
			}
		})
	}
}

func TestCalculateFileHashReportsUnreadableInputs(t *testing.T) {
	dir := t.TempDir()
	for _, path := range []string{filepath.Join(dir, "missing"), dir} {
		if hash, err := calculateFileHash(path); err == nil || hash != "" {
			t.Fatalf("calculateFileHash(%q) = %q, %v", path, hash, err)
		}
	}
}
