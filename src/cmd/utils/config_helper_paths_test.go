// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/spf13/viper"
)

func TestLoadConfigWithInteractiveHelpReturnsUnfixableLoadError(t *testing.T) {
	t.Chdir(t.TempDir())
	viper.Reset()
	t.Cleanup(viper.Reset)
	viper.Set("config", "missing-config.yaml")
	viper.Set("env", "")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("WEAVIATE_API_KEY", "")
	t.Setenv("WEAVIATE_URL", "")
	t.Setenv("QDRANT_API_KEY", "")

	var resultErr error
	withUtilsStdin(t, "n\n", func() {
		_, resultErr = LoadConfigWithInteractiveHelp()
	})
	if resultErr == nil || !strings.Contains(resultErr.Error(), "missing-config.yaml") {
		t.Fatalf("LoadConfigWithInteractiveHelp() error = %v", resultErr)
	}
}

func TestLoadConfigWithInteractiveHelpCanContinueWithMockDefaults(t *testing.T) {
	t.Chdir(t.TempDir())
	viper.Reset()
	t.Cleanup(viper.Reset)
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("WEAVIATE_API_KEY", "")
	t.Setenv("WEAVIATE_URL", "")
	t.Setenv("QDRANT_API_KEY", "")

	var cfgErr error
	var loaded interface{}
	withUtilsStdin(t, "n\n", func() {
		loaded, cfgErr = LoadConfigWithInteractiveHelp()
	})
	if cfgErr != nil || loaded == nil {
		t.Fatalf("LoadConfigWithInteractiveHelp() = (%v, %v), want mock defaults", loaded, cfgErr)
	}
}

func TestLoadConfigWithInteractiveHelpReturnsValidConfigWithoutPrompt(t *testing.T) {
	workDir := t.TempDir()
	t.Chdir(workDir)
	viper.Reset()
	t.Cleanup(viper.Reset)
	configPath := filepath.Join(workDir, "config.yaml")
	configYAML := `databases:
  default: primary
  vector_databases:
    - name: primary
      type: weaviate-cloud
      url: https://example.test
      api_key: test-key
`
	if err := os.WriteFile(configPath, []byte(configYAML), 0o600); err != nil {
		t.Fatal(err)
	}
	viper.Set("config", configPath)
	t.Setenv("WEAVIATE_URL", "https://example.test")
	t.Setenv("WEAVIATE_API_KEY", "test-key")
	t.Setenv("OPENAI_API_KEY", "test-key")

	cfg, err := LoadConfigWithInteractiveHelp()
	if err != nil || cfg == nil {
		t.Fatalf("LoadConfigWithInteractiveHelp() = (%v, %v)", cfg, err)
	}
}

func TestHandleConfigErrorCoversNilAndHelpfulError(t *testing.T) {
	if HandleConfigError(nil, false) {
		t.Fatal("HandleConfigError(nil) = true")
	}
	if !HandleConfigError(&config.ConfigError{
		Message:     "missing configuration",
		MissingVars: []string{"OPENAI_API_KEY"},
	}, false) {
		t.Fatal("HandleConfigError(config error) = false")
	}
}
