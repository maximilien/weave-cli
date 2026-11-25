// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max
package qdrant

import (
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"testing"
)

func TestAddressConfigurationPrecedence(t *testing.T) {
	t.Setenv("QDRANT_HOST", "env-host")
	t.Setenv("QDRANT_GRPC_PORT", "7444")
	for _, tc := range []struct {
		cfg  vectordb.Config
		host string
		port int
	}{{vectordb.Config{}, "env-host", 7444}, {vectordb.Config{URL: "https://url-host:8555"}, "url-host", 8555}, {vectordb.Config{Address: "address-host:9666", URL: "https://ignored:1"}, "address-host", 9666}} {
		host, port := parseAddress(&tc.cfg)
		if host != tc.host || port != tc.port {
			t.Fatalf("address=%s:%d want=%s:%d", host, port, tc.host, tc.port)
		}
	}
}
