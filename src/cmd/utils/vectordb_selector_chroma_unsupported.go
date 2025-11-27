//go:build !(linux && amd64) && !(darwin && amd64) && !(darwin && arm64)
// +build !linux !amd64
// +build !darwin !amd64
// +build !darwin !arm64

// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package utils

import (
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/config"
)

// getChromaLocalConfig returns an error on unsupported platforms
func getChromaLocalConfig(cfg *config.Config) (*config.VectorDBConfig, error) {
	return nil, fmt.Errorf("Chroma is only supported on Linux AMD64, macOS AMD64, and macOS ARM64")
}

// getChromaCloudConfig returns an error on unsupported platforms
func getChromaCloudConfig(cfg *config.Config) (*config.VectorDBConfig, error) {
	return nil, fmt.Errorf("Chroma is only supported on Linux AMD64, macOS AMD64, and macOS ARM64")
}

// tryCreateChromaLocalConfigFromEnv returns nil on unsupported platforms
func tryCreateChromaLocalConfigFromEnv() *config.VectorDBConfig {
	return nil
}

// tryCreateChromaCloudConfigFromEnv returns nil on unsupported platforms
func tryCreateChromaCloudConfigFromEnv() *config.VectorDBConfig {
	return nil
}
