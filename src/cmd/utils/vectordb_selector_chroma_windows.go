//go:build windows
// +build windows

// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package utils

import (
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/config"
)

// getChromaLocalConfig returns an error on Windows (Chroma not supported)
func getChromaLocalConfig(cfg *config.Config) (*config.VectorDBConfig, error) {
	return nil, fmt.Errorf("Chroma is not supported on Windows due to libtokenizers dependency")
}

// getChromaCloudConfig returns an error on Windows (Chroma not supported)
func getChromaCloudConfig(cfg *config.Config) (*config.VectorDBConfig, error) {
	return nil, fmt.Errorf("Chroma is not supported on Windows due to libtokenizers dependency")
}

// tryCreateChromaLocalConfigFromEnv returns nil on Windows
func tryCreateChromaLocalConfigFromEnv() *config.VectorDBConfig {
	return nil
}

// tryCreateChromaCloudConfigFromEnv returns nil on Windows
func tryCreateChromaCloudConfigFromEnv() *config.VectorDBConfig {
	return nil
}
