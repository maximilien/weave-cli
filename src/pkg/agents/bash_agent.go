// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package agents

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
)

// BashAgent executes bash commands safely
type BashAgent struct {
	allowedCommands   []string
	forbiddenCommands []string
	outputAgent       *OutputAgent
}

// NewBashAgent creates a new BashAgent
func NewBashAgent() *BashAgent {
	return &BashAgent{
		allowedCommands: []string{
			"ls", "file", "stat", "wc", "grep", "find",
			"cat", "head", "tail", "echo", "pwd", "which",
		},
		forbiddenCommands: []string{
			"rm", "sudo", "chmod", "chown", "dd", "mkfs",
			"fdisk", "kill", "killall", "reboot", "shutdown",
		},
	}
}

// SetOutputAgent sets the output agent for user interaction
func (a *BashAgent) SetOutputAgent(outputAgent *OutputAgent) {
	a.outputAgent = outputAgent
}

// Name returns the agent's name
func (a *BashAgent) Name() string {
	return "BashAgent"
}

// Execute executes a bash command
func (a *BashAgent) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	cmd, ok := input.(*BashCommand)
	if !ok {
		return nil, fmt.Errorf("invalid input type for BashAgent")
	}

	// Validate command safety
	if err := a.validateCommand(cmd.Command); err != nil {
		return nil, err
	}

	startTime := time.Now()

	// Set timeout if specified
	cmdCtx := ctx
	if cmd.Timeout > 0 {
		var cancel context.CancelFunc
		cmdCtx, cancel = context.WithTimeout(ctx, cmd.Timeout)
		defer cancel()
	} else {
		// Default timeout of 30 seconds
		var cancel context.CancelFunc
		cmdCtx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
	}

	// Build the command
	// If the command contains shell syntax (for, while, if, |, &&, etc.), execute it through sh -c
	var execCmd *exec.Cmd
	if a.isShellScript(cmd.Command) {
		// Execute as shell script
		script := cmd.Command
		if len(cmd.Args) > 0 {
			script = cmd.Command + " " + strings.Join(cmd.Args, " ")
		}
		execCmd = exec.CommandContext(cmdCtx, "sh", "-c", script)
	} else {
		execCmd = exec.CommandContext(cmdCtx, cmd.Command, cmd.Args...)
	}

	if cmd.WorkingDir != "" {
		execCmd.Dir = cmd.WorkingDir
	}

	// Set up environment
	env := os.Environ() // Start with current environment

	if cmd.Environment != nil {
		for k, v := range cmd.Environment {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	execCmd.Env = env

	var stdout, stderr bytes.Buffer
	execCmd.Stdout = &stdout
	execCmd.Stderr = &stderr

	// Execute the command
	err := execCmd.Run()
	duration := time.Since(startTime)

	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = -1
		}
	}

	result := &BashResult{
		Success:  err == nil,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
		Duration: duration,
	}

	return result, nil
}

// isSimpleWeaveCommand checks if the command is just a simple weave CLI call
// (possibly with PATH export prefix)
func isSimpleWeaveCommand(command string) bool {
	// Check if it's the PATH export + weave command pattern
	// Pattern: export PATH="..." && weave <subcommand> [args...]
	if strings.HasPrefix(command, "export PATH=") && strings.Contains(command, " && weave ") {
		// Extract the part after "&&"
		parts := strings.Split(command, " && ")
		if len(parts) == 2 {
			weaveCmd := strings.TrimSpace(parts[1])
			// Check if it's a simple weave command (no pipes, loops, etc.)
			if strings.HasPrefix(weaveCmd, "weave ") {
				// Make sure it doesn't contain complex shell constructs
				complexKeywords := []string{"|", ";", "for ", "while ", "if ", "$(", "`", "do "}
				for _, keyword := range complexKeywords {
					if strings.Contains(weaveCmd, keyword) {
						return false
					}
				}
				return true
			}
		}
	}

	// Also check for direct weave commands without PATH
	if strings.HasPrefix(strings.TrimSpace(command), "weave ") {
		complexKeywords := []string{"|", ";", "for ", "while ", "if ", "$(", "`", "do ", "&&"}
		for _, keyword := range complexKeywords {
			if strings.Contains(command, keyword) {
				return false
			}
		}
		return true
	}

	return false
}

// isShellScript checks if a command is a shell script (contains shell syntax)
func (a *BashAgent) isShellScript(command string) bool {
	// Check for common shell keywords and operators
	shellKeywords := []string{
		"for ", "while ", "if ", "case ", "until ",
		"do ", "done", "then", "else", "elif", "fi",
		"|", "&&", "||", ";", "$(", "`",
	}

	for _, keyword := range shellKeywords {
		if strings.Contains(command, keyword) {
			return true
		}
	}

	return false
}

// validateCommand validates that a command is safe to execute
func (a *BashAgent) validateCommand(command string) error {
	// Check if command contains forbidden commands
	for _, forbidden := range a.forbiddenCommands {
		if strings.Contains(command, forbidden) {
			return fmt.Errorf("command contains forbidden keyword '%s' for security reasons", forbidden)
		}
	}

	// Skip approval for simple weave commands with PATH export
	if isSimpleWeaveCommand(command) {
		return nil
	}

	// For shell scripts, validate the whole script
	if a.isShellScript(command) {
		// Ask user for permission to execute the script
		if a.outputAgent != nil {
			fmt.Println()
			a.outputAgent.PrintWarning("Script execution requires approval")
			fmt.Println()

			// Format the script nicely
			dim := color.New(color.Faint).SprintFunc()
			bold := color.New(color.Bold).SprintFunc()
			fmt.Println(bold("  Script:"))
			fmt.Println()

			// Split script into multiple lines for better readability
			script := command
			// Replace common separators with newlines for display
			script = strings.ReplaceAll(script, "; do ", ";\ndo ")
			script = strings.ReplaceAll(script, "; done", ";\ndone")
			script = strings.ReplaceAll(script, " | ", " |\n  ")
			script = strings.ReplaceAll(script, "$(", "$(\n    ")
			script = strings.ReplaceAll(script, "); ", "\n  );\n  ")

			// Show each line with indentation
			for _, line := range strings.Split(script, "\n") {
				if strings.TrimSpace(line) != "" {
					fmt.Printf("    %s\n", dim(line))
				}
			}
			fmt.Println()

			confirmed, err := a.outputAgent.AskConfirmation("Allow execution of this script?")
			if err != nil || !confirmed {
				return fmt.Errorf("user denied script execution")
			}
			return nil
		}

		// No output agent - deny by default
		return fmt.Errorf("script execution requires user approval")
	}

	// For regular commands, check if in allowed list
	isAllowed := false
	for _, allowed := range a.allowedCommands {
		if command == allowed {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		// Command not in allowed list - ask user for permission
		if a.outputAgent != nil {
			fmt.Println()
			a.outputAgent.PrintWarning(fmt.Sprintf("Command '%s' is not in the pre-approved list", command))
			fmt.Printf("  Allowed commands: %s\n", strings.Join(a.allowedCommands, ", "))
			fmt.Println()

			confirmed, err := a.outputAgent.AskConfirmation(fmt.Sprintf("Allow execution of '%s'?", command))
			if err != nil || !confirmed {
				return fmt.Errorf("user denied execution of command '%s'", command)
			}

			// User approved - temporarily add to allowed list for this session
			a.allowedCommands = append(a.allowedCommands, command)
			return nil
		}

		// No output agent available - deny by default
		return fmt.Errorf("command '%s' is not in the allowed list", command)
	}

	return nil
}

// CheckToolInstalled checks if a tool is installed
func (a *BashAgent) CheckToolInstalled(ctx context.Context, tool string) (bool, error) {
	cmd := &BashCommand{
		Command: "which",
		Args:    []string{tool},
		Timeout: 5 * time.Second,
	}

	result, err := a.Execute(ctx, cmd)
	if err != nil {
		return false, err
	}

	bashResult, ok := result.(*BashResult)
	if !ok {
		return false, fmt.Errorf("unexpected result type")
	}

	return bashResult.Success && strings.TrimSpace(bashResult.Stdout) != "", nil
}

// SuggestInstallation suggests how to install a missing tool
func (a *BashAgent) SuggestInstallation(tool string) string {
	os := runtime.GOOS

	suggestions := map[string]map[string]string{
		"pdftotext": {
			"darwin": "brew install poppler",
			"linux":  "sudo apt-get install poppler-utils",
		},
		"ghostscript": {
			"darwin": "brew install ghostscript",
			"linux":  "sudo apt-get install ghostscript",
		},
		"convert": {
			"darwin": "brew install imagemagick",
			"linux":  "sudo apt-get install imagemagick",
		},
		"tesseract": {
			"darwin": "brew install tesseract",
			"linux":  "sudo apt-get install tesseract-ocr",
		},
	}

	if toolSuggestions, ok := suggestions[tool]; ok {
		if suggestion, ok := toolSuggestions[os]; ok {
			return suggestion
		}
	}

	if os == "darwin" {
		return fmt.Sprintf("brew install %s", tool)
	}

	return fmt.Sprintf("Install %s using your system package manager", tool)
}

// ParseFileList parses output from ls command into file list
func ParseFileList(output string) []string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	files := []string{}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && line != "." && line != ".." {
			files = append(files, line)
		}
	}

	return files
}

// ClassifyFileType classifies a file based on its extension
func ClassifyFileType(filename string) string {
	lower := strings.ToLower(filename)

	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"}
	for _, ext := range imageExts {
		if strings.HasSuffix(lower, ext) {
			return "image"
		}
	}

	textExts := []string{".txt", ".md", ".json", ".yaml", ".yml", ".csv"}
	for _, ext := range textExts {
		if strings.HasSuffix(lower, ext) {
			return "text"
		}
	}

	if strings.HasSuffix(lower, ".pdf") {
		return "pdf"
	}

	return "unknown"
}
