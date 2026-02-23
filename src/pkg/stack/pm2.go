// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package stack

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

// GeneratePM2Config generates ecosystem.config.js from template
func GeneratePM2Config(config *StackConfig, outputPath string) error {
	if config.Dashboard == nil {
		return fmt.Errorf("dashboard config not found in stack config")
	}

	if config.Dashboard.PM2 == nil {
		return fmt.Errorf("PM2 config not found in dashboard config")
	}

	pm2Config := config.Dashboard.PM2

	// Set defaults
	if pm2Config.Cwd == "" {
		pm2Config.Cwd = filepath.Dir(outputPath)
	}

	if pm2Config.LogDateFormat == "" {
		pm2Config.LogDateFormat = "YYYY-MM-DD HH:mm:ss Z"
	}

	// Load template
	tmplPath := "templates/pm2/ecosystem.config.js"
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("failed to load PM2 template: %w", err)
	}

	// Create output file
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create PM2 config file: %w", err)
	}
	defer f.Close()

	// Execute template
	if err := tmpl.Execute(f, pm2Config); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// PM2Start starts the dashboard with PM2
func PM2Start(appName, configPath string) error {
	// Check if PM2 is installed
	if !CommandExists("pm2") {
		return fmt.Errorf("pm2 not installed (install: npm install -g pm2)")
	}

	// Stop existing process (ignore errors)
	_ = PM2Stop(appName)

	// Start with PM2
	cmd := exec.Command("pm2", "start", configPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pm2 start failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// PM2Stop stops the PM2 process
func PM2Stop(appName string) error {
	if !CommandExists("pm2") {
		return fmt.Errorf("pm2 not installed")
	}

	// Stop process (ignore errors if not running)
	cmd := exec.Command("pm2", "stop", appName)
	_ = cmd.Run()

	// Delete process (ignore errors)
	cmd = exec.Command("pm2", "delete", appName)
	_ = cmd.Run()

	return nil
}

// PM2Restart restarts the PM2 process
func PM2Restart(appName string) error {
	if !CommandExists("pm2") {
		return fmt.Errorf("pm2 not installed")
	}

	cmd := exec.Command("pm2", "restart", appName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pm2 restart failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// PM2Status returns the PM2 process status
func PM2Status(appName string) (string, error) {
	if !CommandExists("pm2") {
		return "", fmt.Errorf("pm2 not installed")
	}

	cmd := exec.Command("pm2", "status", appName)
	output, err := cmd.Output()
	if err != nil {
		// PM2 returns error if process not found, but output is still useful
		return string(output), nil
	}

	return string(output), nil
}

// PM2Logs streams logs from PM2
func PM2Logs(appName string, lines int, follow bool) error {
	if !CommandExists("pm2") {
		return fmt.Errorf("pm2 not installed")
	}

	args := []string{"logs", appName, "--lines", fmt.Sprintf("%d", lines)}
	if !follow {
		args = append(args, "--nostream")
	}

	cmd := exec.Command("pm2", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// PM2List lists all PM2 processes
func PM2List() (string, error) {
	if !CommandExists("pm2") {
		return "", fmt.Errorf("pm2 not installed")
	}

	cmd := exec.Command("pm2", "list")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("pm2 list failed: %w", err)
	}

	return string(output), nil
}

// PM2Monit opens PM2 monitoring interface
func PM2Monit() error {
	if !CommandExists("pm2") {
		return fmt.Errorf("pm2 not installed")
	}

	cmd := exec.Command("pm2", "monit")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// CommandExists checks if a command is available in PATH
func CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
