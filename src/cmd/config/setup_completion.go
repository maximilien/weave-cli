// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package config

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var setupCompletionCmd = &cobra.Command{
	Use:   "setup-completion",
	Short: "Setup shell completion for weave CLI",
	Long: `Automatically detect your shell and setup tab completion for weave CLI.

This command will:
  1. Detect your current shell (bash, zsh, fish, powershell)
  2. Show you what will be done
  3. Ask for your permission
  4. Add completion to your shell configuration file

Supported shells: bash, zsh, fish, powershell`,
	Example: `  # Auto-detect and setup completion
  weave config setup-completion

  # Just show instructions without installing
  weave config setup-completion --dry-run

  # Update existing completion (after weave CLI update)
  weave config setup-completion --update`,
	Run: runSetupCompletion,
}

var (
	dryRun           bool
	updateCompletion bool
)

func init() {
	setupCompletionCmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be done without making changes")
	setupCompletionCmd.Flags().BoolVar(&updateCompletion, "update", false, "regenerate completion (useful after weave CLI updates)")
}

func runSetupCompletion(cmd *cobra.Command, args []string) {
	if updateCompletion {
		color.Cyan("🔄 Updating Weave CLI Shell Completion\n")
	} else {
		color.Cyan("🔧 Weave CLI Shell Completion Setup\n")
	}
	fmt.Println()

	// Detect shell
	shell := detectShell()
	if shell == "" {
		color.Red("❌ Could not detect your shell")
		fmt.Println()
		color.Yellow("💡 You can manually setup completion using:")
		fmt.Println("   weave completion --help")
		os.Exit(1)
	}

	color.Green("✓ Detected shell: %s", shell)
	fmt.Println()

	// Get shell-specific setup info
	setup := getShellSetup(shell)
	if setup == nil {
		color.Red("❌ Unsupported shell: %s", shell)
		fmt.Println()
		color.Yellow("💡 Supported shells: bash, zsh, fish, powershell")
		fmt.Println("   Run 'weave completion --help' for manual instructions")
		os.Exit(1)
	}

	// Show what will be done
	if updateCompletion {
		color.White("📋 Update Plan:")
	} else {
		color.White("📋 Setup Plan:")
	}
	fmt.Println()
	fmt.Printf("  Shell:        %s\n", shell)
	fmt.Printf("  Config file:  %s\n", setup.ConfigFile)
	fmt.Printf("  Method:       %s\n", setup.Description)
	fmt.Println()

	if updateCompletion {
		color.White("📝 This will regenerate completion for %s", shell)
		fmt.Println()
		color.Cyan("  %s", setup.CompletionLine)
		fmt.Println()
	} else {
		color.White("📝 This will add the following line to %s:", setup.ConfigFile)
		fmt.Println()
		color.Cyan("  %s", setup.CompletionLine)
		fmt.Println()
	}

	// Dry run - just show and exit
	if dryRun {
		color.Yellow("🔍 Dry run mode - no changes made")
		fmt.Println()
		color.White("To apply these changes, run:")
		if updateCompletion {
			fmt.Println("  weave config setup-completion --update")
		} else {
			fmt.Println("  weave config setup-completion")
		}
		return
	}

	// Ask for permission (skip if updating)
	if !updateCompletion && !askPermission() {
		color.Yellow("⏸️  Setup cancelled")
		return
	}

	// Perform setup
	if err := performSetup(setup); err != nil {
		color.Red("❌ Setup failed: %v", err)
		fmt.Println()
		color.Yellow("💡 You can manually add this line to %s:", setup.ConfigFile)
		fmt.Println("   " + setup.CompletionLine)
		os.Exit(1)
	}

	// Success!
	fmt.Println()
	color.Green("✅ Shell completion setup complete!")
	fmt.Println()
	color.White("🎯 Next steps:")
	fmt.Printf("  1. Reload your shell: %s\n", setup.ReloadCommand)
	fmt.Println("  2. Try it: weave co<TAB>")
	fmt.Println()
}

// ShellSetup contains shell-specific setup information
type ShellSetup struct {
	ConfigFile     string
	CompletionLine string
	Description    string
	ReloadCommand  string
}

// detectShell detects the current shell
func detectShell() string {
	// Try SHELL environment variable first
	shell := os.Getenv("SHELL")
	if shell != "" {
		return filepath.Base(shell)
	}

	// Windows PowerShell detection
	if runtime.GOOS == "windows" {
		return "powershell"
	}

	// Try to detect from parent process (works on Unix-like systems)
	cmd := exec.Command("ps", "-p", fmt.Sprintf("%d", os.Getppid()), "-o", "comm=")
	output, err := cmd.Output()
	if err == nil {
		shell = strings.TrimSpace(string(output))
		if shell != "" {
			return filepath.Base(shell)
		}
	}

	return ""
}

// getShellSetup returns setup information for a given shell
func getShellSetup(shell string) *ShellSetup {
	home, _ := os.UserHomeDir()

	switch shell {
	case "bash":
		// Prefer .bashrc, fallback to .bash_profile
		configFile := filepath.Join(home, ".bashrc")
		if runtime.GOOS == "darwin" {
			// macOS often uses .bash_profile
			if _, err := os.Stat(filepath.Join(home, ".bash_profile")); err == nil {
				configFile = filepath.Join(home, ".bash_profile")
			}
		}

		return &ShellSetup{
			ConfigFile:     configFile,
			CompletionLine: "source <(weave completion bash)",
			Description:    "Dynamic completion loading",
			ReloadCommand:  "source " + configFile,
		}

	case "zsh":
		return &ShellSetup{
			ConfigFile:     filepath.Join(home, ".zshrc"),
			CompletionLine: "source <(weave completion zsh)",
			Description:    "Dynamic completion loading",
			ReloadCommand:  "source ~/.zshrc",
		}

	case "fish":
		fishDir := filepath.Join(home, ".config", "fish", "completions")
		return &ShellSetup{
			ConfigFile:     filepath.Join(fishDir, "weave.fish"),
			CompletionLine: "# Generated by: weave completion fish",
			Description:    "Fish completion file",
			ReloadCommand:  "exec fish",
		}

	case "powershell", "pwsh":
		// Get PowerShell profile path
		cmd := exec.Command("powershell", "-NoProfile", "-Command", "echo $PROFILE")
		output, err := cmd.Output()
		profilePath := strings.TrimSpace(string(output))
		if err != nil || profilePath == "" {
			profilePath = filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
		}

		return &ShellSetup{
			ConfigFile:     profilePath,
			CompletionLine: "weave completion powershell | Out-String | Invoke-Expression",
			Description:    "PowerShell completion loading",
			ReloadCommand:  ". $PROFILE",
		}

	default:
		return nil
	}
}

// askPermission asks the user for permission to proceed
func askPermission() bool {
	color.Yellow("❓ Would you like to proceed? (Y/n): ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "" || response == "y" || response == "yes"
}

// performSetup performs the actual setup
func performSetup(setup *ShellSetup) error {
	// Special handling for fish - always regenerate on update
	if strings.HasSuffix(setup.ConfigFile, "weave.fish") {
		// Create directory if needed
		dir := filepath.Dir(setup.ConfigFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		// Generate fish completion
		cmd := exec.Command("weave", "completion", "fish")
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to generate fish completion: %w", err)
		}

		// Write to file (overwrites if exists)
		if err := os.WriteFile(setup.ConfigFile, output, 0644); err != nil {
			return fmt.Errorf("failed to write completion file: %w", err)
		}

		if updateCompletion {
			color.Green("✓ Fish completion regenerated")
		}
		return nil
	}

	// For bash, zsh, powershell - append to config file

	// Check if line already exists
	content, err := os.ReadFile(setup.ConfigFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	alreadyConfigured := strings.Contains(string(content), "weave completion")

	// If updating and already configured, show message
	if updateCompletion {
		if alreadyConfigured {
			color.Yellow("ℹ️  Completion already configured - no changes needed")
			color.White("   For %s, completion is loaded dynamically from 'weave completion %s'", setup.ConfigFile, detectShell())
			return nil
		}
		// If updating but not configured, fall through to add it
		color.Yellow("ℹ️  Completion not found in %s - adding it now", setup.ConfigFile)
	}

	// If not updating and already configured, skip
	if !updateCompletion && alreadyConfigured {
		color.Yellow("⚠️  Completion already configured in %s", setup.ConfigFile)
		return nil
	}

	// Open file for appending
	f, err := os.OpenFile(setup.ConfigFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	// Add comment and completion line
	comment := "\n# weave CLI completion (added by 'weave config setup-completion')\n"
	if _, err := f.WriteString(comment + setup.CompletionLine + "\n"); err != nil {
		return fmt.Errorf("failed to write to config file: %w", err)
	}

	return nil
}
