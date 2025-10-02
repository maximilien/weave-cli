// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package utils

import (
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/spf13/viper"
)

// LoadConfigWithOverrides loads configuration with command-line overrides
func LoadConfigWithOverrides() (*config.Config, error) {
	// Get global flags from viper (set by root command)
	options := config.LoadConfigOptions{
		VectorDBType:   getStringFlag("vector-db-type"),
		WeaviateAPIKey: getStringFlag("weaviate-api-key"),
		WeaviateURL:    getStringFlag("weaviate-url"),
		ConfigFile:     getStringFlag("config"),
		EnvFile:        getStringFlag("env"),
	}

	cfg, err := config.LoadConfigWithOptions(options)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %v", err)
	}

	return cfg, nil
}

// getStringFlag gets a string flag value from viper
func getStringFlag(flagName string) string {
	if viper.IsSet(flagName) {
		return viper.GetString(flagName)
	}
	return ""
}
