// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package mcpinstaller

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGetCurrentPlatform(t *testing.T) {
	platform := GetCurrentPlatform()
	if platform.OS != runtime.GOOS || platform.Arch != runtime.GOARCH {
		t.Fatalf("platform = %#v, want %s/%s", platform, runtime.GOOS, runtime.GOARCH)
	}
}

func TestGetDefaultInstallPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	installPath, err := GetDefaultInstallPath()
	if err != nil {
		t.Fatalf("GetDefaultInstallPath() error: %v", err)
	}

	var want string
	switch runtime.GOOS {
	case "darwin", "linux":
		want = filepath.Join(home, ".local", "bin")
	case "windows":
		want = filepath.Join(home, "AppData", "Local", "Programs")
	default:
		want = filepath.Join(home, "bin")
	}
	if installPath != want {
		t.Fatalf("install path = %q, want %q", installPath, want)
	}
}

func TestGetBinaryAsset(t *testing.T) {
	release := &Release{Assets: []Asset{
		{Name: "weave-mcp-http-linux-amd64"},
		{Name: "weave-mcp-stdio-linux-amd64", BrowserDownloadURL: "linux"},
		{Name: "weave-mcp-stdio-darwin-arm64", BrowserDownloadURL: "darwin"},
		{Name: "weave-mcp-stdio-windows-amd64.exe", BrowserDownloadURL: "windows"},
	}}

	tests := []struct {
		name     string
		platform Platform
		wantURL  string
	}{
		{name: "linux", platform: Platform{OS: "linux", Arch: "amd64"}, wantURL: "linux"},
		{name: "darwin", platform: Platform{OS: "darwin", Arch: "arm64"}, wantURL: "darwin"},
		{name: "windows", platform: Platform{OS: "windows", Arch: "amd64"}, wantURL: "windows"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asset, err := GetBinaryAsset(release, tt.platform)
			if err != nil {
				t.Fatalf("GetBinaryAsset() error: %v", err)
			}
			if asset.BrowserDownloadURL != tt.wantURL {
				t.Fatalf("download URL = %q, want %q", asset.BrowserDownloadURL, tt.wantURL)
			}
		})
	}

	_, err := GetBinaryAsset(release, Platform{OS: "plan9", Arch: "amd64"})
	if err == nil || !strings.Contains(err.Error(), "no binary found") {
		t.Fatalf("missing asset error = %v", err)
	}
}

func TestGetChecksumsAsset(t *testing.T) {
	release := &Release{Assets: []Asset{{Name: "binary"}, {Name: "checksums.txt", Size: 42}}}
	asset, err := GetChecksumsAsset(release)
	if err != nil {
		t.Fatalf("GetChecksumsAsset() error: %v", err)
	}
	if asset.Size != 42 {
		t.Fatalf("checksum asset size = %d, want 42", asset.Size)
	}

	if _, err := GetChecksumsAsset(&Release{}); err == nil {
		t.Fatal("GetChecksumsAsset() succeeded without a checksum asset")
	}
}

func TestDownloadFile(t *testing.T) {
	const body = "downloaded binary"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)

	destination := filepath.Join(t.TempDir(), "weave-mcp")
	if err := DownloadFile(server.URL, destination, int64(len(body))); err != nil {
		t.Fatalf("DownloadFile() error: %v", err)
	}
	data, err := os.ReadFile(destination)
	if err != nil {
		t.Fatalf("read download: %v", err)
	}
	if string(data) != body {
		t.Fatalf("download = %q, want %q", data, body)
	}

	if err := DownloadFile(server.URL, t.TempDir(), int64(len(body))); err == nil {
		t.Fatal("DownloadFile() succeeded when destination was a directory")
	}
	if err := DownloadFile("://invalid", filepath.Join(t.TempDir(), "bad"), 0); err == nil {
		t.Fatal("DownloadFile() succeeded with an invalid URL")
	}
}

func TestVerifyChecksum(t *testing.T) {
	content := []byte("checksum content")
	filePath := filepath.Join(t.TempDir(), "binary")
	if err := os.WriteFile(filePath, content, 0600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	digest := sha256.Sum256(content)
	expected := hex.EncodeToString(digest[:])

	valid, err := VerifyChecksum(filePath, expected)
	if err != nil || !valid {
		t.Fatalf("VerifyChecksum() = %v, %v; want true, nil", valid, err)
	}
	valid, err = VerifyChecksum(filePath, strings.Repeat("0", 64))
	if err != nil || valid {
		t.Fatalf("mismatched VerifyChecksum() = %v, %v; want false, nil", valid, err)
	}
	if _, err := VerifyChecksum(filepath.Join(t.TempDir(), "missing"), expected); err == nil {
		t.Fatal("VerifyChecksum() succeeded for a missing file")
	}
}

func TestGetExpectedChecksum(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("aaa  other\nbbb  weave-mcp-stdio-linux-amd64\nmalformed\n"))
	}))
	t.Cleanup(server.Close)
	asset := &Asset{BrowserDownloadURL: server.URL}

	checksum, err := GetExpectedChecksum(asset, "weave-mcp-stdio-linux-amd64")
	if err != nil || checksum != "bbb" {
		t.Fatalf("GetExpectedChecksum() = %q, %v; want bbb, nil", checksum, err)
	}
	if _, err := GetExpectedChecksum(asset, "missing"); err == nil {
		t.Fatal("GetExpectedChecksum() succeeded for a missing binary")
	}

	server.Close()
	if _, err := GetExpectedChecksum(asset, "binary"); err == nil {
		t.Fatal("GetExpectedChecksum() succeeded after the server closed")
	}
}

func TestExecutableHelpers(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "weave-mcp")
	if err := os.WriteFile(filePath, []byte("binary"), 0600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	if runtime.GOOS != "windows" && CheckIfExecutable(filePath) {
		t.Fatal("non-executable file reported as executable")
	}
	if err := MakeExecutable(filePath); err != nil {
		t.Fatalf("MakeExecutable() error: %v", err)
	}
	if !CheckIfExecutable(filePath) {
		t.Fatal("executable file reported as non-executable")
	}
	if CheckIfExecutable(filepath.Join(t.TempDir(), "missing")) {
		t.Fatal("missing file reported as executable")
	}
	if err := MakeExecutable(filepath.Join(t.TempDir(), "missing")); runtime.GOOS != "windows" && err == nil {
		t.Fatal("MakeExecutable() succeeded for a missing file")
	}
}
