// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/maximilien/weave-cli/src/pkg/mcpinstaller"
)

var (
	weaveMCPForce bool
)

// runWeaveMCPUpdate handles the weave-mcp installation
func runWeaveMCPUpdate() error {
	printHeader("Weave MCP Binary Installer")
	fmt.Println()

	// Get current platform
	platform := mcpinstaller.GetCurrentPlatform()
	color.New(color.FgCyan).Printf("📋 Platform: %s-%s\n", platform.OS, platform.Arch)
	fmt.Println()

	// Check if already installed
	currentPath := os.Getenv("WEAVE_MCP_STDIO_PATH")
	if currentPath != "" {
		exists := mcpinstaller.CheckIfExecutable(currentPath)
		if exists && !weaveMCPForce {
			color.New(color.FgYellow).Printf("⚠️  weave-mcp is already installed at: %s\n", currentPath)
			fmt.Println()
			fmt.Print("Do you want to reinstall? (y/N): ")
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.ToLower(strings.TrimSpace(response))
			if response != "y" && response != "yes" {
				fmt.Println("Installation cancelled.")
				return nil
			}
			fmt.Println()
		}
	}

	// Fetch latest release
	color.New(color.FgCyan).Println("Fetching latest release information from GitHub...")
	release, err := mcpinstaller.GetLatestRelease()
	if err != nil {
		return fmt.Errorf("failed to fetch release: %w", err)
	}

	color.New(color.FgGreen).Printf("✅ Found release: %s (%s)\n", release.Name, release.TagName)
	fmt.Println()

	// Get binary asset for current platform
	binaryAsset, err := mcpinstaller.GetBinaryAsset(release, platform)
	if err != nil {
		return fmt.Errorf("failed to find binary: %w", err)
	}

	color.New(color.FgCyan).Printf("📦 Binary: %s (%.2f MB)\n", binaryAsset.Name, float64(binaryAsset.Size)/(1024*1024))
	fmt.Println()

	// Get checksums asset
	checksumsAsset, err := mcpinstaller.GetChecksumsAsset(release)
	if err != nil {
		return fmt.Errorf("failed to find checksums: %w", err)
	}

	// Ask for installation directory
	defaultPath, err := mcpinstaller.GetDefaultInstallPath()
	if err != nil {
		return fmt.Errorf("failed to get default path: %w", err)
	}

	fmt.Printf("📂 Where would you like to install weave-mcp?\n")
	fmt.Printf("   Default: %s\n", defaultPath)
	fmt.Printf("   Note: This directory should be in your PATH\n")
	fmt.Print("\nInstall directory (press Enter for default): ")

	reader := bufio.NewReader(os.Stdin)
	installDir, _ := reader.ReadString('\n')
	installDir = strings.TrimSpace(installDir)
	if installDir == "" {
		installDir = defaultPath
	}

	// Expand ~ to home directory
	if strings.HasPrefix(installDir, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		installDir = filepath.Join(homeDir, installDir[1:])
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Set binary name based on platform
	binaryName := "weave-mcp-stdio"
	if platform.OS == "windows" {
		binaryName += ".exe"
	}
	binaryPath := filepath.Join(installDir, binaryName)

	fmt.Println()
	color.New(color.FgCyan).Printf("Installing to: %s\n", binaryPath)
	fmt.Println()

	// Download binary
	tempFile := binaryPath + ".tmp"
	if err := mcpinstaller.DownloadFile(binaryAsset.BrowserDownloadURL, tempFile, binaryAsset.Size); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("failed to download binary: %w", err)
	}

	// Verify checksum
	color.New(color.FgCyan).Println("Verifying checksum...")
	expectedChecksum, err := mcpinstaller.GetExpectedChecksum(checksumsAsset, binaryAsset.Name)
	if err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("failed to get expected checksum: %w", err)
	}

	valid, err := mcpinstaller.VerifyChecksum(tempFile, expectedChecksum)
	if err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("failed to verify checksum: %w", err)
	}

	if !valid {
		os.Remove(tempFile)
		return fmt.Errorf("checksum verification failed - download may be corrupted")
	}

	color.New(color.FgGreen).Println("✅ Checksum verified")
	fmt.Println()

	// Ask before making executable
	if platform.OS != "windows" {
		fmt.Print("Make binary executable? (Y/n): ")
		response, _ := reader.ReadString('\n')
		response = strings.ToLower(strings.TrimSpace(response))
		if response == "" || response == "y" || response == "yes" {
			if err := mcpinstaller.MakeExecutable(tempFile); err != nil {
				os.Remove(tempFile)
				return fmt.Errorf("failed to make executable: %w", err)
			}
			color.New(color.FgGreen).Println("✅ Binary made executable")
			fmt.Println()
		}
	}

	// Move temp file to final location
	if err := os.Rename(tempFile, binaryPath); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("failed to install binary: %w", err)
	}

	// Success!
	color.New(color.FgGreen, color.Bold).Println("✅ weave-mcp installed successfully!")
	fmt.Println()

	// Test binary
	color.New(color.FgCyan).Println("Testing installation...")
	if mcpinstaller.CheckIfExecutable(binaryPath) {
		color.New(color.FgGreen).Println("✅ Binary is executable")
	} else {
		color.New(color.FgYellow).Println("⚠️  Binary may not be executable")
	}
	fmt.Println()

	// Provide next steps
	printHeader("Next Steps")
	fmt.Println()

	// Check if install directory is in PATH
	pathEnv := os.Getenv("PATH")
	inPath := strings.Contains(pathEnv, installDir)

	if !inPath {
		color.New(color.FgYellow).Printf("⚠️  %s is not in your PATH\n\n", installDir)
		fmt.Println("Add it to your PATH by adding this to your shell profile:")
		fmt.Printf("  export PATH=\"%s:$PATH\"\n\n", installDir)
	}

	fmt.Println("Set the WEAVE_MCP_STDIO_PATH environment variable:")
	fmt.Printf("  export WEAVE_MCP_STDIO_PATH=\"%s\"\n\n", binaryPath)

	fmt.Println("Or add it to your .env file:")
	fmt.Printf("  echo 'WEAVE_MCP_STDIO_PATH=\"%s\"' >> .env\n\n", binaryPath)

	fmt.Println("Test the installation:")
	fmt.Printf("  %s --version\n\n", binaryPath)

	fmt.Println("Start using REPL mode:")
	fmt.Println("  weave")
	fmt.Println()

	return nil
}
