// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max
package supabase

import (
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"testing"
)

func TestAdapterInitializationWithoutDatabaseConnection(t *testing.T) {
	for _, local := range []bool{false, true} {
		for _, key := range []string{"", "fixture-key"} {
			t.Run(key+map[bool]string{false: "cloud", true: "local"}[local], func(t *testing.T) {
				t.Setenv("OPENAI_API_KEY", key)
				kind := vectordb.VectorDBTypeSupabaseCloud
				if local {
					kind = vectordb.VectorDBTypeSupabaseLocal
				}
				cfg := &vectordb.Config{Type: kind, URL: "https://fixture.supabase.co", DatabaseKey: "fixture"}
				a, err := NewAdapter(cfg)
				if err != nil {
					t.Fatal(err)
				}
				if a.config != cfg || a.db != nil || (a.llmClient != nil) != (key != "") {
					t.Fatalf("unexpected initialization: %+v", a)
				}
				if (a.client == nil) != local {
					t.Fatal("unexpected HTTP client selection")
				}
				if err = a.Close(); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}
