// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package image

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/rwcarlsen/goexif/exif"
)

// A minimal TIFF with EXIF fields exercises both byte and file extraction.
func exifFixture() []byte {
	data := make([]byte, 256)
	copy(data, "II")
	binary.LittleEndian.PutUint16(data[2:], 42)
	binary.LittleEndian.PutUint32(data[4:], 8)
	binary.LittleEndian.PutUint16(data[8:], 5)
	type entry struct {
		tag, kind    uint16
		count, value uint32
	}
	entries := []entry{
		{0x010f, 2, 6, 86},   // Make
		{0x0110, 2, 7, 92},   // Model
		{0x0112, 3, 1, 1},    // Orientation
		{0x0132, 2, 20, 100}, // DateTime
		{0x8769, 4, 1, 128},  // EXIF IFD
	}
	for i, e := range entries {
		off := 10 + i*12
		binary.LittleEndian.PutUint16(data[off:], e.tag)
		binary.LittleEndian.PutUint16(data[off+2:], e.kind)
		binary.LittleEndian.PutUint32(data[off+4:], e.count)
		binary.LittleEndian.PutUint32(data[off+8:], e.value)
	}
	copy(data[86:], "Canon\x00")
	copy(data[92:], "EOS 80\x00")
	copy(data[100:], "2026:10:22 10:00:00\x00")
	binary.LittleEndian.PutUint16(data[128:], 4)
	for i, e := range []entry{
		{0xa002, 4, 1, 6000}, // width
		{0xa003, 4, 1, 4000}, // height
		{0x829d, 5, 1, 190},  // aperture
		{0x8827, 3, 1, 200},  // ISO
	} {
		off := 130 + i*12
		binary.LittleEndian.PutUint16(data[off:], e.tag)
		binary.LittleEndian.PutUint16(data[off+2:], e.kind)
		binary.LittleEndian.PutUint32(data[off+4:], e.count)
		binary.LittleEndian.PutUint32(data[off+8:], e.value)
	}
	binary.LittleEndian.PutUint32(data[190:], 28)
	binary.LittleEndian.PutUint32(data[194:], 10)
	return data
}

func TestExtractEXIFMetadataFromBytesAndFile(t *testing.T) {
	fixture := exifFixture()
	if _, err := exif.Decode(bytes.NewReader(fixture)); err != nil {
		t.Fatalf("fixture decode: %v", err)
	}
	path := filepath.Join(t.TempDir(), "metadata.tiff")
	if err := os.WriteFile(path, fixture, 0o600); err != nil {
		t.Fatal(err)
	}
	for name, extract := range map[string]func() (*EXIFData, error){
		"bytes": func() (*EXIFData, error) { return ExtractEXIFFromBytes(fixture) },
		"file":  func() (*EXIFData, error) { return ExtractEXIF(path) },
	} {
		t.Run(name, func(t *testing.T) {
			got, err := extract()
			if err != nil {
				t.Fatal(err)
			}
			if got.Make != "Canon" || got.Model != "EOS 80" || got.Orientation != 1 ||
				got.Width != 6000 || got.Height != 4000 || got.ISO != 200 || got.FNumber != 2.8 {
				t.Fatalf("unexpected EXIF metadata: %+v", got)
			}
		})
	}
}
