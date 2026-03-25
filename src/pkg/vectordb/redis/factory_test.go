// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package redis

import (
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

func TestFactory_GetSupportedTypes(t *testing.T) {
	f := NewFactory()
	types := f.GetSupportedTypes()
	if len(types) != 2 {
		t.Fatalf("expected 2 types, got %d", len(types))
	}
	found := map[vectordb.VectorDBType]bool{}
	for _, tp := range types {
		found[tp] = true
	}
	if !found[vectordb.VectorDBTypeRedisLocal] {
		t.Error("missing redis-local type")
	}
	if !found[vectordb.VectorDBTypeRedisCloud] {
		t.Error("missing redis-cloud type")
	}
}

func TestFactory_ValidateConfig(t *testing.T) {
	f := NewFactory()

	tests := []struct {
		name    string
		config  *vectordb.Config
		wantErr bool
	}{
		{
			name:    "Nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name: "Wrong database type",
			config: &vectordb.Config{
				Type: vectordb.VectorDBTypeQdrantLocal,
				URL:  "localhost:6379",
			},
			wantErr: true,
		},
		{
			name: "Missing URL and address",
			config: &vectordb.Config{
				Type: vectordb.VectorDBTypeRedisLocal,
			},
			wantErr: true,
		},
		{
			name: "Negative timeout",
			config: &vectordb.Config{
				Type:    vectordb.VectorDBTypeRedisLocal,
				URL:     "localhost:6379",
				Timeout: -1,
			},
			wantErr: true,
		},
		{
			name: "Negative vector dimensions",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeRedisLocal,
				URL:              "localhost:6379",
				VectorDimensions: -100,
			},
			wantErr: true,
		},
		{
			name: "Invalid similarity metric",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeRedisLocal,
				URL:              "localhost:6379",
				SimilarityMetric: "MANHATTAN",
			},
			wantErr: true,
		},
		{
			name: "Valid cosine metric",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeRedisLocal,
				URL:              "localhost:6379",
				SimilarityMetric: "COSINE",
			},
			wantErr: false,
		},
		{
			name: "Valid L2 metric",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeRedisLocal,
				URL:              "localhost:6379",
				SimilarityMetric: "L2",
			},
			wantErr: false,
		},
		{
			name: "Valid IP metric",
			config: &vectordb.Config{
				Type:             vectordb.VectorDBTypeRedisLocal,
				URL:              "localhost:6379",
				SimilarityMetric: "IP",
			},
			wantErr: false,
		},
		{
			name: "Valid minimal config",
			config: &vectordb.Config{
				Type: vectordb.VectorDBTypeRedisLocal,
				URL:  "localhost:6379",
			},
			wantErr: false,
		},
		{
			name: "Cloud missing auth",
			config: &vectordb.Config{
				Type: vectordb.VectorDBTypeRedisCloud,
				URL:  "redis.cloud.example.com:6380",
			},
			wantErr: true,
		},
		{
			name: "Cloud with password",
			config: &vectordb.Config{
				Type:     vectordb.VectorDBTypeRedisCloud,
				URL:      "redis.cloud.example.com:6380",
				Password: "secret",
			},
			wantErr: false,
		},
		{
			name: "Cloud with API key",
			config: &vectordb.Config{
				Type:   vectordb.VectorDBTypeRedisCloud,
				URL:    "redis.cloud.example.com:6380",
				APIKey: "api-key-123",
			},
			wantErr: false,
		},
		{
			name: "Address instead of URL",
			config: &vectordb.Config{
				Type:    vectordb.VectorDBTypeRedisLocal,
				Address: "localhost:6379",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := f.ValidateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v",
					err, tt.wantErr)
			}
		})
	}
}

func TestParseRedisAddress(t *testing.T) {
	tests := []struct {
		name string
		cfg  *vectordb.Config
		want string
	}{
		{
			name: "Address field",
			cfg:  &vectordb.Config{Address: "myhost:6380"},
			want: "myhost:6380",
		},
		{
			name: "URL with redis:// prefix",
			cfg:  &vectordb.Config{URL: "redis://myhost:6379"},
			want: "myhost:6379",
		},
		{
			name: "URL with rediss:// prefix",
			cfg:  &vectordb.Config{URL: "rediss://cloud.redis.io:6380"},
			want: "cloud.redis.io:6380",
		},
		{
			name: "URL with userinfo",
			cfg:  &vectordb.Config{URL: "redis://user:pass@myhost:6379/0"},
			want: "myhost:6379",
		},
		{
			name: "No URL or address",
			cfg:  &vectordb.Config{},
			want: "localhost:6379",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseRedisAddress(tt.cfg)
			if got != tt.want {
				t.Errorf("parseRedisAddress() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFloat64sToBytes(t *testing.T) {
	vec := []float64{1.0, 2.0, 3.0}
	b := float64sToBytes(vec)
	if len(b) != 12 { // 3 * 4 bytes
		t.Errorf("expected 12 bytes, got %d", len(b))
	}
}

func TestFloat32sToBytes(t *testing.T) {
	vec := []float32{1.0, 2.0, 3.0}
	b := float32sToBytes(vec)
	if len(b) != 12 {
		t.Errorf("expected 12 bytes, got %d", len(b))
	}
}

func TestFlatSliceToMap(t *testing.T) {
	s := []interface{}{"key1", "val1", "key2", "val2"}
	m := flatSliceToMap(s)
	if m["key1"] != "val1" || m["key2"] != "val2" {
		t.Errorf("unexpected map: %v", m)
	}
}

func TestEscapeRedisQuery(t *testing.T) {
	q := "hello.world@test"
	escaped := escapeRedisQuery(q)
	if escaped == q {
		t.Error("expected special chars to be escaped")
	}
	// . and @ should be escaped
	if escaped != "hello\\.world\\@test" {
		t.Errorf("unexpected escaped: %q", escaped)
	}
}

func TestInferEmbeddingModelFromDimensions(t *testing.T) {
	tests := []struct {
		dims int
		want string
	}{
		{1536, "text-embedding-3-small"},
		{3072, "text-embedding-3-large"},
		{768, "sentence-transformers/all-mpnet-base-v2"},
		{384, "sentence-transformers/all-MiniLM-L6-v2"},
		{1024, "nomic-embed-text"},
		{999, "text-embedding-3-small"}, // fallback
	}
	for _, tt := range tests {
		got := inferEmbeddingModelFromDimensions(tt.dims)
		if got != tt.want {
			t.Errorf("dims=%d: got %q, want %q", tt.dims, got, tt.want)
		}
	}
}
