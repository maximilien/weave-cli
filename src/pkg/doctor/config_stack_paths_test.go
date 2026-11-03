// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package doctor

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/config"
)

func TestConfigDiagnosticsWithLocalFixtures(t *testing.T) {
	t.Chdir(t.TempDir())
	if got := CheckConfig(nil, nil); len(got) != 1 || got[0].Status != StatusFail {
		t.Fatalf("missing config: %+v", got)
	}
	if err := os.WriteFile("config.yml", []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	if got := CheckConfig(nil, errors.New("invalid yaml")); len(got) != 2 || got[1].Status != StatusFail {
		t.Fatalf("invalid config: %+v", got)
	}
	cfg := &config.Config{}
	cfg.Databases.VectorDatabases = []config.VectorDBConfig{{Name: "broken", Type: config.VectorDBTypeMongoDB}}
	got := CheckConfig(cfg, nil)
	if len(got) < 3 || got[0].Status != StatusOK || got[1].Status != StatusOK || got[2].Status != StatusFail {
		t.Fatalf("validation errors: %+v", got)
	}
}

func TestStackAndEmbeddingDiagnosticsWithLocalFixtures(t *testing.T) {
	t.Chdir(t.TempDir())
	if got := CheckStack(); len(got) != 1 || got[0].Status != StatusSkip {
		t.Fatalf("missing stack: %+v", got)
	}
	if err := os.WriteFile("weave-stack.yaml", []byte(`version: "1"
name: fixture
infrastructure:
  vectordb:
    type: milvus
    version: "2"
collections:
  - name: mismatch
    schema:
      vector_dimensions: 384
    embedding:
      model: text-embedding-3-small
  - name: unknown
    embedding:
      model: definitely-not-a-model
`), 0600); err != nil {
		t.Fatal(err)
	}
	if got := CheckStack(); len(got) != 2 || got[1].Status != StatusWarn {
		t.Fatalf("missing cluster: %+v", got)
	}
	got := CheckEmbeddings(nil)
	if len(got) != 3 || got[0].Status != StatusOK || got[1].Status != StatusFail || got[2].Status != StatusWarn {
		t.Fatalf("model diagnostics: %+v", got)
	}
	if err := os.Mkdir(".weave-state", 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(".weave-state/cluster.json", []byte("[invalid"), 0600); err != nil {
		t.Fatal(err)
	}
	if got := CheckStack(); len(got) != 2 || got[1].Status != StatusFail {
		t.Fatalf("invalid cluster: %+v", got)
	}
	if err := os.WriteFile(".weave-state/cluster.json", []byte("name: fixture\nprovider: unknown\nstatus: stopped\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if got := CheckStack(); len(got) != 4 || got[2].Status != StatusWarn || got[3].Status != StatusFail {
		t.Fatalf("inactive cluster: %+v", got)
	}
}

func TestVDBFixSuggestions(t *testing.T) {
	for _, db := range []config.VectorDBConfig{{URL: "http://fixture"}, {DatabaseURL: "http://fixture"}, {Address: "http://fixture"}} {
		if got := suggestVDBFix(db); !strings.Contains(got, "http://fixture") {
			t.Fatalf("suggestVDBFix(%+v) = %q", db, got)
		}
	}
	if got := suggestVDBFix(config.VectorDBConfig{}); !strings.Contains(got, "URL/address") {
		t.Fatalf("empty fix = %q", got)
	}
}
