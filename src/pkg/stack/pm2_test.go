// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package stack

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratePM2Config(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "ecosystem.config.js")

	config := &StackConfig{
		Version: "1.0",
		Name:    "test-stack",
		Dashboard: &DashboardConfig{
			Enabled: true,
			Runtime: "pm2",
			PM2: &PM2Config{
				AppName:          "test-dashboard",
				Script:           "dist/index.js",
				Cwd:              tmpDir,
				Instances:        1,
				MaxMemoryRestart: "1G",
				Autorestart:      true,
				Watch:            false,
				ErrorLog:         "logs/pm2-error.log",
				OutLog:           "logs/pm2-out.log",
				MergeLogs:        true,
				MinUptime:        "10s",
				MaxRestarts:      10,
				KillTimeout:      5000,
				Env: map[string]string{
					"NODE_ENV":       "production",
					"DASHBOARD_PORT": "3000",
				},
			},
		},
	}

	// Skip if template doesn't exist (running from test dir)
	if _, err := os.Stat("templates/pm2/ecosystem.config.js"); os.IsNotExist(err) {
		t.Skip("Skipping: PM2 template not found (run from project root)")
	}

	err := GeneratePM2Config(config, outputPath)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(outputPath)
	require.NoError(t, err, "ecosystem.config.js should be created")

	// Read and verify content
	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	// Check for key elements
	assert.Contains(t, string(content), "test-dashboard", "should contain app name")
	assert.Contains(t, string(content), "dist/index.js", "should contain script path")
	assert.Contains(t, string(content), "NODE_ENV", "should contain env vars")
	assert.Contains(t, string(content), "production", "should contain env values")
	assert.Contains(t, string(content), "1G", "should contain memory limit")
	assert.Contains(t, string(content), "logs/pm2-error.log", "should contain error log path")
}

func TestGeneratePM2Config_NoDashboard(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "ecosystem.config.js")

	config := &StackConfig{
		Version: "1.0",
		Name:    "test-stack",
		// No dashboard config
	}

	err := GeneratePM2Config(config, outputPath)
	assert.Error(t, err, "should error when dashboard config missing")
	assert.Contains(t, err.Error(), "dashboard config not found")
}

func TestGeneratePM2Config_NoPM2Config(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "ecosystem.config.js")

	config := &StackConfig{
		Version: "1.0",
		Name:    "test-stack",
		Dashboard: &DashboardConfig{
			Enabled: true,
			Runtime: "pm2",
			// No PM2 config
		},
	}

	err := GeneratePM2Config(config, outputPath)
	assert.Error(t, err, "should error when PM2 config missing")
	assert.Contains(t, err.Error(), "PM2 config not found")
}

func TestGeneratePM2Config_WithDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "ecosystem.config.js")

	config := &StackConfig{
		Version: "1.0",
		Name:    "test-stack",
		Dashboard: &DashboardConfig{
			Enabled: true,
			Runtime: "pm2",
			PM2: &PM2Config{
				AppName:          "test-dashboard",
				Script:           "dist/index.js",
				Instances:        1,
				MaxMemoryRestart: "1G",
				Autorestart:      true,
				Watch:            false,
				ErrorLog:         "logs/pm2-error.log",
				OutLog:           "logs/pm2-out.log",
				MergeLogs:        true,
				MinUptime:        "10s",
				MaxRestarts:      10,
				KillTimeout:      5000,
				// No Cwd or LogDateFormat - should use defaults
			},
		},
	}

	// Skip if template doesn't exist
	if _, err := os.Stat("templates/pm2/ecosystem.config.js"); os.IsNotExist(err) {
		t.Skip("Skipping: PM2 template not found (run from project root)")
	}

	err := GeneratePM2Config(config, outputPath)
	require.NoError(t, err)

	// Verify defaults were set
	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	assert.Contains(t, string(content), "YYYY-MM-DD HH:mm:ss Z", "should have default log date format")
}

func TestCommandExists(t *testing.T) {
	// Test with a command that should exist
	assert.True(t, CommandExists("ls"), "ls should exist on system")

	// Test with a command that shouldn't exist
	assert.False(t, CommandExists("nonexistent-command-12345"), "nonexistent command should return false")
}

func TestPM2Start_NoPM2Installed(t *testing.T) {
	// This test validates error handling when PM2 is not installed
	// We can't actually test PM2 start without PM2 installed

	if CommandExists("pm2") {
		t.Skip("Skipping: PM2 is installed, can't test 'not installed' case")
	}

	err := PM2Start("test-app", "ecosystem.config.js")
	assert.Error(t, err, "should error when PM2 not installed")
	assert.Contains(t, err.Error(), "pm2 not installed")
}

func TestPM2Stop_NoPM2Installed(t *testing.T) {
	if CommandExists("pm2") {
		t.Skip("Skipping: PM2 is installed, can't test 'not installed' case")
	}

	err := PM2Stop("test-app")
	assert.Error(t, err, "should error when PM2 not installed")
	assert.Contains(t, err.Error(), "pm2 not installed")
}

func TestPM2ConfigStruct(t *testing.T) {
	// Test PM2Config struct can be created and has expected fields
	config := PM2Config{
		AppName:          "test-app",
		Script:           "dist/index.js",
		Cwd:              "/path/to/app",
		Instances:        2,
		MaxMemoryRestart: "512M",
		Autorestart:      true,
		Watch:            false,
		ErrorLog:         "logs/error.log",
		OutLog:           "logs/out.log",
		LogDateFormat:    "YYYY-MM-DD",
		MergeLogs:        true,
		MinUptime:        "5s",
		MaxRestarts:      5,
		KillTimeout:      3000,
		Env: map[string]string{
			"NODE_ENV": "development",
			"PORT":     "3000",
		},
	}

	assert.Equal(t, "test-app", config.AppName)
	assert.Equal(t, 2, config.Instances)
	assert.Equal(t, "512M", config.MaxMemoryRestart)
	assert.True(t, config.Autorestart)
	assert.False(t, config.Watch)
	assert.Equal(t, 2, len(config.Env))
	assert.Equal(t, "development", config.Env["NODE_ENV"])
}

func TestDashboardConfigWithPM2(t *testing.T) {
	// Test DashboardConfig can include PM2 config
	dashboard := DashboardConfig{
		Enabled: true,
		Type:    "web",
		Runtime: "pm2",
		PM2: &PM2Config{
			AppName:   "dashboard",
			Script:    "dist/index.js",
			Instances: 1,
		},
	}

	assert.True(t, dashboard.Enabled)
	assert.Equal(t, "pm2", dashboard.Runtime)
	assert.NotNil(t, dashboard.PM2)
	assert.Equal(t, "dashboard", dashboard.PM2.AppName)
}
