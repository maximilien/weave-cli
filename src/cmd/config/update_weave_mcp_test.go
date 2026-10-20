// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/mcpinstaller"
)

func TestWeaveMCPUpdateCanDeclineReinstall(t *testing.T) {
	installDir := t.TempDir()
	binary := filepath.Join(installDir, "weave-mcp-stdio")
	if err := os.WriteFile(binary, []byte("fixture"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("WEAVE_MCP_STDIO_PATH", binary)

	previousForce := weaveMCPForce
	weaveMCPForce = false
	t.Cleanup(func() { weaveMCPForce = previousForce })

	oldStdin := os.Stdin
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writer.WriteString("n\n"); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdin = reader
	t.Cleanup(func() {
		os.Stdin = oldStdin
		_ = reader.Close()
	})

	if err := runWeaveMCPUpdate(); err != nil {
		t.Fatalf("runWeaveMCPUpdate() after declining reinstall = %v", err)
	}
}

type mcpUpdateRoundTripper func(*http.Request) (*http.Response, error)

func (f mcpUpdateRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func configureMCPUpdateHTTP(t *testing.T, release mcpinstaller.Release, binary []byte, checksumBody string, binaryStatus int) {
	t.Helper()
	encodedRelease, err := json.Marshal(release)
	if err != nil {
		t.Fatal(err)
	}
	previousTransport := http.DefaultTransport
	http.DefaultTransport = mcpUpdateRoundTripper(func(request *http.Request) (*http.Response, error) {
		var body string
		status := http.StatusOK
		switch request.URL.Host + request.URL.Path {
		case "api.github.com/repos/maximilien/weave-mcp/releases/latest":
			body = string(encodedRelease)
		case "fixture.test/binary":
			body = string(binary)
			if binaryStatus != 0 {
				status = binaryStatus
			}
		case "fixture.test/checksums.txt":
			body = checksumBody
		default:
			return nil, fmt.Errorf("unexpected MCP installer request: %s", request.URL)
		}
		return &http.Response{
			StatusCode: status,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
			Request:    request,
		}, nil
	})
	t.Cleanup(func() { http.DefaultTransport = previousTransport })
}

func setMCPUpdateStdin(t *testing.T, input string) {
	t.Helper()
	oldStdin := os.Stdin
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(writer, input); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdin = reader
	t.Cleanup(func() {
		os.Stdin = oldStdin
		_ = reader.Close()
	})
}

func TestWeaveMCPUpdateInstallsAndUpdatesEnv(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	if err := os.MkdirAll(home, 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	t.Setenv("WEAVE_MCP_STDIO_PATH", "")
	t.Chdir(root)
	if err := os.WriteFile(".env", []byte("EXISTING=value\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	binaryName := fmt.Sprintf("weave-mcp-stdio-%s-%s", runtime.GOOS, runtime.GOARCH)
	binaryFileName := "weave-mcp-stdio"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
		binaryFileName += ".exe"
	}
	binary := []byte("fixture weave-mcp binary")
	digest := sha256.Sum256(binary)
	checksum := hex.EncodeToString(digest[:])
	release := mcpinstaller.Release{
		Name: "Fixture release", TagName: "v9.9.9",
		Assets: []mcpinstaller.Asset{
			{Name: binaryName, BrowserDownloadURL: "https://fixture.test/binary", Size: int64(len(binary))},
			{Name: "checksums.txt", BrowserDownloadURL: "https://fixture.test/checksums.txt"},
		},
	}
	configureMCPUpdateHTTP(t, release, binary, checksum+"  "+binaryName+"\n", http.StatusOK)

	previousForce := weaveMCPForce
	weaveMCPForce = true
	t.Cleanup(func() { weaveMCPForce = previousForce })
	setMCPUpdateStdin(t, "\n\ny\n")

	if err := runWeaveMCPUpdate(); err != nil {
		t.Fatalf("runWeaveMCPUpdate() = %v", err)
	}
	installedPath := filepath.Join(home, ".local", "bin", binaryFileName)
	installed, err := os.ReadFile(installedPath)
	if err != nil || string(installed) != string(binary) {
		t.Fatalf("installed binary = %q, %v", installed, err)
	}
	envFile, err := os.ReadFile(".env")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(envFile), `EXISTING="value"`) || !strings.Contains(string(envFile), "WEAVE_MCP_STDIO_PATH=") || !strings.Contains(string(envFile), installedPath) {
		t.Fatalf("updated .env = %q", envFile)
	}
}

func TestWeaveMCPUpdateReportsInstallerFailures(t *testing.T) {
	binary := []byte("fixture weave-mcp binary")
	digest := sha256.Sum256(binary)
	checksum := hex.EncodeToString(digest[:])
	binaryName := fmt.Sprintf("weave-mcp-stdio-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	validAssets := []mcpinstaller.Asset{
		{Name: binaryName, BrowserDownloadURL: "https://fixture.test/binary", Size: int64(len(binary))},
		{Name: "checksums.txt", BrowserDownloadURL: "https://fixture.test/checksums.txt"},
	}

	tests := []struct {
		name         string
		assets       []mcpinstaller.Asset
		checksumBody string
		binaryStatus int
		stdin        string
		clearHome    bool
		wantError    string
	}{
		{name: "missing binary asset", assets: validAssets[1:], wantError: "no binary found"},
		{name: "missing checksums asset", assets: validAssets[:1], wantError: "checksums.txt not found"},
		{name: "default path home unavailable", assets: validAssets, clearHome: true, wantError: "failed to get default path"},
		{name: "download failure", assets: validAssets, checksumBody: checksum + "  " + binaryName, binaryStatus: http.StatusBadGateway, stdin: "\n", wantError: "failed to download binary"},
		{name: "missing checksum entry", assets: validAssets, checksumBody: "other-checksum  unrelated-binary\n", stdin: "\n", wantError: "failed to get expected checksum"},
		{name: "checksum mismatch", assets: validAssets, checksumBody: strings.Repeat("0", 64) + "  " + binaryName, stdin: "\n", wantError: "checksum verification failed"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			home := filepath.Join(root, "home")
			if err := os.MkdirAll(home, 0o700); err != nil {
				t.Fatal(err)
			}
			t.Setenv("HOME", home)
			if test.clearHome {
				t.Setenv("HOME", "")
			}
			t.Setenv("WEAVE_MCP_STDIO_PATH", "")
			t.Chdir(root)
			configureMCPUpdateHTTP(t, mcpinstaller.Release{
				Name: "Fixture release", TagName: "v9.9.9", Assets: test.assets,
			}, binary, test.checksumBody, test.binaryStatus)
			setMCPUpdateStdin(t, test.stdin)

			previousForce := weaveMCPForce
			weaveMCPForce = true
			t.Cleanup(func() { weaveMCPForce = previousForce })
			if err := runWeaveMCPUpdate(); err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("runWeaveMCPUpdate() error = %v, want %q", err, test.wantError)
			}
		})
	}
}

func TestWeaveMCPUpdateReplacesAnExistingEnvPath(t *testing.T) {
	for _, test := range []struct {
		name  string
		input string
		want  bool
		force bool
	}{
		{name: "accept", input: "\n\ny\n", want: true, force: true},
		{name: "decline", input: "\n\nn\n", want: false, force: true},
		{name: "confirm reinstall", input: "y\n\n\ny\n", want: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			home := filepath.Join(root, "home")
			if err := os.MkdirAll(home, 0o700); err != nil {
				t.Fatal(err)
			}
			oldBinary := filepath.Join(root, "previous-install", "weave-mcp-stdio")
			if err := os.MkdirAll(filepath.Dir(oldBinary), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(oldBinary, []byte("previous"), 0o700); err != nil {
				t.Fatal(err)
			}
			t.Setenv("HOME", home)
			t.Setenv("WEAVE_MCP_STDIO_PATH", oldBinary)
			t.Chdir(root)
			if err := os.WriteFile(".env", []byte("EXISTING=value\nWEAVE_MCP_STDIO_PATH=\""+oldBinary+"\"\n"), 0o600); err != nil {
				t.Fatal(err)
			}

			binary := []byte("replacement weave-mcp binary")
			digest := sha256.Sum256(binary)
			checksum := hex.EncodeToString(digest[:])
			binaryName := fmt.Sprintf("weave-mcp-stdio-%s-%s", runtime.GOOS, runtime.GOARCH)
			binaryFileName := "weave-mcp-stdio"
			if runtime.GOOS == "windows" {
				binaryName += ".exe"
				binaryFileName += ".exe"
			}
			configureMCPUpdateHTTP(t, mcpinstaller.Release{
				Name: "Fixture release", TagName: "v9.9.9",
				Assets: []mcpinstaller.Asset{
					{Name: binaryName, BrowserDownloadURL: "https://fixture.test/binary", Size: int64(len(binary))},
					{Name: "checksums.txt", BrowserDownloadURL: "https://fixture.test/checksums.txt"},
				},
			}, binary, checksum+"  "+binaryName+"\n", http.StatusOK)
			setMCPUpdateStdin(t, test.input)
			previousForce := weaveMCPForce
			weaveMCPForce = test.force
			t.Cleanup(func() { weaveMCPForce = previousForce })

			if err := runWeaveMCPUpdate(); err != nil {
				t.Fatalf("runWeaveMCPUpdate() = %v", err)
			}
			envFile, err := os.ReadFile(".env")
			if err != nil {
				t.Fatal(err)
			}
			installedPath := filepath.Join(home, ".local", "bin", binaryFileName)
			if got := strings.Contains(string(envFile), installedPath); got != test.want {
				t.Fatalf(".env contains new install path = %t, want %t: %q", got, test.want, envFile)
			}
		})
	}
}

func TestWeaveMCPUpdateInstallPathAndPermissionOptions(t *testing.T) {
	binary := []byte("fixture weave-mcp binary")
	digest := sha256.Sum256(binary)
	checksum := hex.EncodeToString(digest[:])
	binaryName := fmt.Sprintf("weave-mcp-stdio-%s-%s", runtime.GOOS, runtime.GOARCH)
	binaryFileName := "weave-mcp-stdio"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
		binaryFileName += ".exe"
	}
	release := mcpinstaller.Release{
		Name: "Fixture release", TagName: "v9.9.9",
		Assets: []mcpinstaller.Asset{
			{Name: binaryName, BrowserDownloadURL: "https://fixture.test/binary", Size: int64(len(binary))},
			{Name: "checksums.txt", BrowserDownloadURL: "https://fixture.test/checksums.txt"},
		},
	}

	for _, test := range []struct {
		name       string
		installDir func(root, home string) string
		input      func(root, home string) string
		prepare    func(t *testing.T, installDir string)
		wantError  string
		makeNoExec bool
	}{
		{
			name:       "expand home shorthand",
			installDir: func(_, home string) string { return filepath.Join(home, ".custom", "bin") },
			input:      func(_, _ string) string { return "~/.custom/bin\ny\n" },
		},
		{
			name:       "create directory failure",
			installDir: func(root, _ string) string { return filepath.Join(root, "blocker", "child") },
			input:      func(root, _ string) string { return filepath.Join(root, "blocker", "child") + "\n" },
			prepare: func(t *testing.T, path string) {
				if err := os.WriteFile(filepath.Dir(path), []byte("blocker"), 0o600); err != nil {
					t.Fatal(err)
				}
			},
			wantError: "failed to create directory",
		},
		{
			name:       "rename failure",
			installDir: func(root, _ string) string { return filepath.Join(root, "install") },
			input:      func(root, _ string) string { return filepath.Join(root, "install") + "\n\ny\n" },
			prepare: func(t *testing.T, path string) {
				if err := os.MkdirAll(path, 0o700); err != nil {
					t.Fatal(err)
				}
			},
			wantError: "failed to install binary",
		},
		{
			name:       "leave installed binary non-executable",
			installDir: func(_, home string) string { return filepath.Join(home, ".local", "bin") },
			input:      func(_, _ string) string { return "\nn\n" },
			makeNoExec: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			home := filepath.Join(root, "home")
			if err := os.MkdirAll(home, 0o700); err != nil {
				t.Fatal(err)
			}
			t.Setenv("HOME", home)
			t.Setenv("WEAVE_MCP_STDIO_PATH", "")
			t.Chdir(root)
			installDir := test.installDir(root, home)
			if test.prepare != nil {
				if test.name == "rename failure" {
					if err := os.MkdirAll(installDir, 0o700); err != nil {
						t.Fatal(err)
					}
					installDir = filepath.Join(installDir, binaryFileName)
				}
				test.prepare(t, installDir)
			}
			configureMCPUpdateHTTP(t, release, binary, checksum+"  "+binaryName+"\n", http.StatusOK)
			setMCPUpdateStdin(t, test.input(root, home))
			previousForce := weaveMCPForce
			weaveMCPForce = true
			t.Cleanup(func() { weaveMCPForce = previousForce })

			err := runWeaveMCPUpdate()
			if test.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantError) {
					t.Fatalf("runWeaveMCPUpdate() error = %v, want %q", err, test.wantError)
				}
				return
			}
			if err != nil {
				t.Fatalf("runWeaveMCPUpdate() = %v", err)
			}
			if test.makeNoExec && runtime.GOOS != "windows" {
				path := filepath.Join(installDir, binaryFileName)
				if info, err := os.Stat(path); err != nil || info.Mode()&0o111 != 0 {
					t.Fatalf("installed binary mode = %v, %v; want non-executable", info, err)
				}
			}
		})
	}
}
