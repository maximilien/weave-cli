// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	// GlobalConfigDirName is the directory name for global config
	GlobalConfigDirName = ".weave-cli"
	// EnvFileName is the name of the environment file
	EnvFileName = ".env"
	// ConfigFileName is the name of the YAML config file
	ConfigFileName = "config.yaml"
)

// ConfigPaths holds the paths to configuration files
type ConfigPaths struct {
	EnvPath    string
	ConfigPath string
	Location   string // "local" or "global"
}

// GetGlobalConfigDir returns the path to ~/.weave-cli directory
func GetGlobalConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, GlobalConfigDirName), nil
}

// EnsureGlobalConfigDir creates ~/.weave-cli directory if it doesn't exist
// Returns true if directory was created, false if it already existed
func EnsureGlobalConfigDir() (bool, error) {
	globalDir, err := GetGlobalConfigDir()
	if err != nil {
		return false, err
	}

	// Check if directory already exists
	if _, err := os.Stat(globalDir); err == nil {
		return false, nil // Directory already exists
	}

	// Create directory
	if err := os.MkdirAll(globalDir, 0755); err != nil {
		return false, fmt.Errorf("failed to create global config directory: %w", err)
	}

	return true, nil
}

// FindConfigPaths finds configuration files with proper precedence:
// 1. Local directory (current working directory)
// 2. Global directory (~/.weave-cli)
func FindConfigPaths() (*ConfigPaths, error) {
	// Try local directory first
	localEnv := EnvFileName
	localConfig := ConfigFileName

	localEnvExists := fileExists(localEnv)
	localConfigExists := fileExists(localConfig)

	// If either local file exists, use local
	if localEnvExists || localConfigExists {
		return &ConfigPaths{
			EnvPath:    localEnv,
			ConfigPath: localConfig,
			Location:   "local",
		}, nil
	}

	// Try global directory
	globalDir, err := GetGlobalConfigDir()
	if err != nil {
		return nil, err
	}

	globalEnv := filepath.Join(globalDir, EnvFileName)
	globalConfig := filepath.Join(globalDir, ConfigFileName)

	globalEnvExists := fileExists(globalEnv)
	globalConfigExists := fileExists(globalConfig)

	// If either global file exists, use global
	if globalEnvExists || globalConfigExists {
		return &ConfigPaths{
			EnvPath:    globalEnv,
			ConfigPath: globalConfig,
			Location:   "global",
		}, nil
	}

	// No config found, return local paths as default
	return &ConfigPaths{
		EnvPath:    localEnv,
		ConfigPath: localConfig,
		Location:   "local",
	}, nil
}

// GetGlobalEnvPath returns the path to ~/.weave-cli/.env
func GetGlobalEnvPath() (string, error) {
	globalDir, err := GetGlobalConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(globalDir, EnvFileName), nil
}

// GetGlobalConfigPath returns the path to ~/.weave-cli/config.yaml
func GetGlobalConfigPath() (string, error) {
	globalDir, err := GetGlobalConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(globalDir, ConfigFileName), nil
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// HasLocalConfig returns true if local .env or config.yaml exists
func HasLocalConfig() bool {
	return fileExists(EnvFileName) || fileExists(ConfigFileName)
}

// HasGlobalConfig returns true if global .env or config.yaml exists
func HasGlobalConfig() (bool, error) {
	globalEnv, err := GetGlobalEnvPath()
	if err != nil {
		return false, err
	}
	globalConfig, err := GetGlobalConfigPath()
	if err != nil {
		return false, err
	}
	return fileExists(globalEnv) || fileExists(globalConfig), nil
}
