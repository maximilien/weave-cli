//go:build !linux || !amd64
// +build !linux !amd64

// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package utils

import (
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/config"
)

// getChromaLocalConfig returns an error on unsupported platforms
func getChromaLocalConfig(cfg *config.Config) (*config.VectorDBConfig, error) {
	return nil, fmt.Errorf("Chroma is only supported on Linux AMD64 due to libtokenizers dependency")
}

// getChromaCloudConfig returns an error on unsupported platforms
func getChromaCloudConfig(cfg *config.Config) (*config.VectorDBConfig, error) {
	return nil, fmt.Errorf("Chroma is only supported on Linux AMD64 due to libtokenizers dependency")
}

// tryCreateChromaLocalConfigFromEnv returns nil on unsupported platforms
func tryCreateChromaLocalConfigFromEnv() *config.VectorDBConfig {
	return nil
}

// tryCreateChromaCloudConfigFromEnv returns nil on unsupported platforms
func tryCreateChromaCloudConfigFromEnv() *config.VectorDBConfig {
	return nil
}
