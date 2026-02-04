// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
)

// FixApplier applies fixes to a config file
type FixApplier struct {
	configPath string
	backupPath string
	editor     *YAMLEditor
}

// NewFixApplier creates a new fix applier
func NewFixApplier(configPath string) (*FixApplier, error) {
	editor, err := NewYAMLEditor(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return &FixApplier{
		configPath: configPath,
		editor:     editor,
	}, nil
}

// ApplyFixes applies a list of fix results to the config file
func (a *FixApplier) ApplyFixes(results []FixResult, createBackup bool) error {
	// Create backup if requested
	if createBackup {
		if err := a.createBackup(); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}

	// Group fixes by action type for efficient processing
	var setValueFixes []FixResult
	var removeFixes []FixResult
	var disableFixes []FixResult

	for _, result := range results {
		switch result.Action {
		case FixActionSetValue:
			setValueFixes = append(setValueFixes, result)
		case FixActionRemove:
			removeFixes = append(removeFixes, result)
		case FixActionDisable:
			disableFixes = append(disableFixes, result)
		case FixActionSkip:
			// Nothing to do
		}
	}

	// Apply set value fixes
	for _, fix := range setValueFixes {
		if err := a.applySetValue(fix); err != nil {
			return fmt.Errorf("failed to set %s.%s: %w", fix.Issue.DatabaseName, fix.Issue.Field, err)
		}
	}

	// Apply disable fixes
	for _, fix := range disableFixes {
		if err := a.applyDisable(fix); err != nil {
			return fmt.Errorf("failed to disable %s: %w", fix.Issue.DatabaseName, err)
		}
	}

	// Apply remove fixes (do these last, in reverse order to avoid index issues)
	// Sort by index in descending order
	for i := len(removeFixes) - 1; i >= 0; i-- {
		fix := removeFixes[i]
		if err := a.applyRemove(fix); err != nil {
			return fmt.Errorf("failed to remove %s: %w", fix.Issue.DatabaseName, err)
		}
	}

	// Save the modified config
	if err := a.editor.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// applySetValue applies a set value fix
func (a *FixApplier) applySetValue(fix FixResult) error {
	return a.editor.SetValue(fix.Issue.Path, fix.Value)
}

// applyRemove applies a remove fix
func (a *FixApplier) applyRemove(fix FixResult) error {
	if fix.Issue.DatabaseIdx < 0 {
		return fmt.Errorf("cannot remove non-database field")
	}

	return a.editor.RemoveArrayElement("databases.vector_databases", fix.Issue.DatabaseIdx)
}

// applyDisable applies a disable fix
func (a *FixApplier) applyDisable(fix FixResult) error {
	if fix.Issue.DatabaseIdx < 0 {
		return fmt.Errorf("cannot disable non-database field")
	}

	return a.editor.DisableArrayElement("databases.vector_databases", fix.Issue.DatabaseIdx)
}

// createBackup creates a backup of the config file
func (a *FixApplier) createBackup() error {
	// Generate backup filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	dir := filepath.Dir(a.configPath)
	base := filepath.Base(a.configPath)
	ext := filepath.Ext(base)
	nameWithoutExt := base[:len(base)-len(ext)]

	backupName := fmt.Sprintf("%s.backup-%s%s", nameWithoutExt, timestamp, ext)
	a.backupPath = filepath.Join(dir, backupName)

	// Copy file
	data, err := os.ReadFile(a.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	if err := os.WriteFile(a.backupPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write backup: %w", err)
	}

	color.Green("💾 Backup created: %s", a.backupPath)
	return nil
}

// GetBackupPath returns the path to the backup file
func (a *FixApplier) GetBackupPath() string {
	return a.backupPath
}

// DryRun shows what would be changed without actually changing anything
func DryRun(results []FixResult) {
	color.New(color.FgCyan, color.Bold).Println("\n🔍 Dry Run - No changes will be made")
	fmt.Println()

	color.White("The following changes would be applied:")
	fmt.Println()

	for _, result := range results {
		issue := result.Issue
		switch result.Action {
		case FixActionSetValue:
			displayValue := result.Value
			if issue.IsSecretField() {
				displayValue = maskSecret(displayValue)
			}
			color.Green("  ✅ SET: %s.%s = %s", issue.DatabaseName, issue.Field, displayValue)
		case FixActionSkip:
			color.Yellow("  ⏭️  SKIP: %s.%s", issue.DatabaseName, issue.Field)
		case FixActionRemove:
			color.Red("  🗑️  REMOVE: database %s (index %d)", issue.DatabaseName, issue.DatabaseIdx)
		case FixActionDisable:
			color.Yellow("  🚫 DISABLE: database %s (index %d)", issue.DatabaseName, issue.DatabaseIdx)
		}
	}

	fmt.Println()
	color.Cyan("To apply these changes, run without --dry-run")
}
