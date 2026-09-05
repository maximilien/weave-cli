// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package mcpinstaller

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

	"github.com/schollz/progressbar/v3"
)

const (
	// GitHub repository information
	repoOwner = "maximilien"
	repoName  = "weave-mcp"
	apiURL    = "https://api.github.com/repos/%s/%s/releases/latest"
)

// Release represents a GitHub release
type Release struct {
	TagName string  `json:"tag_name"`
	Name    string  `json:"name"`
	Assets  []Asset `json:"assets"`
}

// Asset represents a release asset
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// Platform represents the current platform information
type Platform struct {
	OS   string
	Arch string
}

// InstallOptions holds installation options
type InstallOptions struct {
	Version     string // Specific version to install (e.g., "v0.1.3"), empty for latest
	InstallPath string // Where to install the binary
	Force       bool   // Force reinstall even if already exists
}

// GetCurrentPlatform returns the current platform information
func GetCurrentPlatform() Platform {
	return Platform{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}
}

// GetDefaultInstallPath returns the default installation path for the current OS
func GetDefaultInstallPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	platform := GetCurrentPlatform()
	switch platform.OS {
	case "darwin", "linux":
		return filepath.Join(homeDir, ".local", "bin"), nil
	case "windows":
		return filepath.Join(homeDir, "AppData", "Local", "Programs"), nil
	default:
		return filepath.Join(homeDir, "bin"), nil
	}
}

// GetLatestRelease fetches the latest release information from GitHub
func GetLatestRelease() (*Release, error) {
	url := fmt.Sprintf(apiURL, repoOwner, repoName)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release information: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release information: %w", err)
	}

	return &release, nil
}

// GetBinaryAsset finds the appropriate binary asset for the current platform
func GetBinaryAsset(release *Release, platform Platform) (*Asset, error) {
	// Map Go OS/Arch to release naming convention
	osName := platform.OS
	archName := platform.Arch

	// weave-mcp uses linux/darwin/windows and amd64/arm64
	if archName == "amd64" {
		archName = "amd64"
	} else if archName == "arm64" {
		archName = "arm64"
	}

	// Look for binary asset matching platform
	// Binary names in releases: weave-mcp-stdio-darwin-arm64, weave-mcp-stdio-linux-amd64, etc.
	// Note: We need the stdio version, not the HTTP server version
	binaryName := fmt.Sprintf("weave-mcp-stdio-%s-%s", osName, archName)
	if platform.OS == "windows" {
		binaryName += ".exe"
	}

	for _, asset := range release.Assets {
		if asset.Name == binaryName {
			return &asset, nil
		}
	}

	return nil, fmt.Errorf("no binary found for platform %s-%s (looking for %s)", osName, archName, binaryName)
}

// GetChecksumsAsset finds the checksums.txt asset
func GetChecksumsAsset(release *Release) (*Asset, error) {
	for _, asset := range release.Assets {
		if asset.Name == "checksums.txt" {
			return &asset, nil
		}
	}
	return nil, fmt.Errorf("checksums.txt not found in release")
}

// DownloadFile downloads a file from the given URL with a progress bar
func DownloadFile(url string, filepath string, size int64) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("download failed with HTTP status %d", resp.StatusCode)
	}

	// Create progress bar
	bar := progressbar.DefaultBytes(
		size,
		"Downloading",
	)

	// Write the body to file with progress
	_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Println() // New line after progress bar
	return nil
}

// VerifyChecksum verifies the SHA256 checksum of a file
func VerifyChecksum(filepath string, expectedChecksum string) (bool, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return false, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return false, fmt.Errorf("failed to compute checksum: %w", err)
	}

	actualChecksum := hex.EncodeToString(hash.Sum(nil))
	return actualChecksum == expectedChecksum, nil
}

// GetExpectedChecksum fetches and parses the checksums.txt to get the expected checksum
func GetExpectedChecksum(checksumsAsset *Asset, binaryName string) (string, error) {
	// Download checksums.txt
	resp, err := http.Get(checksumsAsset.BrowserDownloadURL)
	if err != nil {
		return "", fmt.Errorf("failed to download checksums: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("checksum download failed with HTTP status %d", resp.StatusCode)
	}

	// Read and parse checksums
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read checksums: %w", err)
	}

	// Parse checksums.txt format: "checksum  filename"
	lines := strings.Split(string(body), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			if parts[1] == binaryName {
				return parts[0], nil
			}
		}
	}

	return "", fmt.Errorf("checksum not found for %s", binaryName)
}

// MakeExecutable makes a file executable on Unix-like systems
func MakeExecutable(filepath string) error {
	platform := GetCurrentPlatform()
	if platform.OS == "windows" {
		// Windows doesn't need chmod
		return nil
	}

	return os.Chmod(filepath, 0755)
}

// CheckIfExecutable checks if a file exists and is executable
func CheckIfExecutable(filepath string) bool {
	info, err := os.Stat(filepath)
	if err != nil {
		return false
	}

	platform := GetCurrentPlatform()
	if platform.OS == "windows" {
		// On Windows, just check if file exists
		return true
	}

	// On Unix-like systems, check executable bit
	return info.Mode()&0111 != 0
}
