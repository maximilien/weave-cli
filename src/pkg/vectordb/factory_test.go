// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package vectordb

import (
	"errors"
	"testing"

	legacyconfig "github.com/maximilien/weave-cli/src/pkg/config"
)

type stubFactory struct {
	validationErr error
	createErr     error
	createdConfig *Config
}

func (f *stubFactory) CreateClient(config *Config) (VectorDBClient, error) {
	f.createdConfig = config
	return nil, f.createErr
}

func (f *stubFactory) GetSupportedTypes() []VectorDBType { return []VectorDBType{"stub"} }

func (f *stubFactory) ValidateConfig(*Config) error { return f.validationErr }

func TestRegistry(t *testing.T) {
	registry := NewRegistry()
	factory := &stubFactory{}
	registry.Register("stub", factory)

	if !registry.IsSupported("stub") || registry.IsSupported("missing") {
		t.Fatal("registry support lookup returned an unexpected result")
	}
	types := registry.GetSupportedTypes()
	if len(types) != 1 || types[0] != "stub" {
		t.Fatalf("supported types = %#v, want [stub]", types)
	}
	config := &Config{Type: "stub", URL: "http://example.test"}
	client, err := registry.CreateClient(config)
	if err != nil || client != nil || factory.createdConfig != config {
		t.Fatalf("CreateClient() = %T, %v; captured=%p", client, err, factory.createdConfig)
	}

	if _, err := registry.CreateClient(&Config{Type: "missing"}); err == nil {
		t.Fatal("CreateClient() accepted an unsupported type")
	}
	factory.validationErr = errors.New("invalid")
	if _, err := registry.CreateClient(config); err == nil || !errors.Is(err, factory.validationErr) {
		t.Fatalf("validation error = %v", err)
	}
	factory.validationErr = nil
	factory.createErr = errors.New("create failed")
	if _, err := registry.CreateClient(config); !errors.Is(err, factory.createErr) {
		t.Fatalf("create error = %v", err)
	}
}

func TestGlobalRegistryAndLegacyConfigConversion(t *testing.T) {
	const dbType VectorDBType = "day1-test"
	factory := &stubFactory{}
	RegisterFactory(dbType, factory)
	t.Cleanup(func() { delete(globalRegistry.factories, dbType) })

	if !IsSupported(dbType) {
		t.Fatal("global registry did not register factory")
	}
	if _, err := CreateClient(&Config{Type: dbType}); err != nil {
		t.Fatalf("global CreateClient() error: %v", err)
	}
	if len(GetSupportedTypes()) == 0 {
		t.Fatal("global supported types is empty")
	}

	legacy := &legacyconfig.VectorDBConfig{
		Type:                   legacyconfig.VectorDBType(dbType),
		URL:                    "http://database.test",
		APIKey:                 "database-key",
		OpenAIAPIKey:           "openai-key",
		Timeout:                17,
		Enabled:                true,
		SimulateEmbeddings:     true,
		EmbeddingDimension:     384,
		DatabaseURL:            "postgres://database",
		DatabaseKey:            "postgres-key",
		Database:               "database",
		Tenant:                 "tenant",
		VectorDimensions:       1536,
		SimilarityMetric:       "cosine",
		Address:                "localhost:19530",
		Username:               "user",
		Password:               "password",
		ImageStorageType:       "minio",
		ImageStorageEndpoint:   "localhost:9000",
		ImageStorageAccessKey:  "access",
		ImageStorageSecretKey:  "secret",
		ImageStorageRegion:     "us-west-2",
		ImageStorageBucket:     "images",
		ImageStoragePathPrefix: "prefix",
		ImageStorageUseSSL:     true,
		PDFStorageEnabled:      true,
	}
	if _, err := CreateClientFromVectorDBConfig(legacy); err != nil {
		t.Fatalf("CreateClientFromVectorDBConfig() error: %v", err)
	}
	converted := factory.createdConfig
	if converted == nil || converted.URL != legacy.URL || converted.VectorDimensions != 1536 {
		t.Fatalf("converted config = %#v", converted)
	}
	if converted.ImageStorage == nil || converted.ImageStorage.PathPrefix != "prefix" {
		t.Fatalf("converted image storage = %#v", converted.ImageStorage)
	}
	if converted.PDFStorage == nil || converted.PDFStorage.PathPrefix != "pdfs" {
		t.Fatalf("converted PDF storage = %#v", converted.PDFStorage)
	}
}
