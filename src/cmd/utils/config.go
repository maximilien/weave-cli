// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package utils

import (
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/config"
)

// LoadConfigWithOverrides loads configuration with command-line overrides
func LoadConfigWithOverrides() (*config.Config, error) {
	// Use LoadConfigWithOptions with empty options since global flags are handled by root.go
	options := config.LoadConfigOptions{}
	cfg, err := config.LoadConfigWithOptions(options)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %v", err)
	}

	return cfg, nil
}
