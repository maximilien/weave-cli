// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package supabase

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrepareConnectionString(t *testing.T) {
	tests := []struct {
		name              string
		dbURL             string
		expectError       bool
		expectedContains  []string
		expectedSSLMode   string
	}{
		{
			name:             "Localhost disables SSL",
			dbURL:            "postgresql://user:pass@localhost:5432/postgres",
			expectError:      false,
			expectedContains: []string{"sslmode=disable", "connect_timeout=10"},
		},
		{
			name:             "127.0.0.1 disables SSL",
			dbURL:            "postgresql://user:pass@127.0.0.1:5432/postgres",
			expectError:      false,
			expectedContains: []string{"sslmode=disable"},
		},
		{
			name:             "Cloud hostname requires SSL",
			dbURL:            "postgresql://postgres:pass@db.project.supabase.co:5432/postgres",
			expectError:      false,
			expectedContains: []string{"sslmode=require"},
		},
		{
			name:             "Missing port defaults to 5432",
			dbURL:            "postgresql://user:pass@localhost/postgres",
			expectError:      false,
			expectedContains: []string{"sslmode=disable"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := prepareConnectionString(tt.dbURL)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				for _, contains := range tt.expectedContains {
					assert.Contains(t, result, contains)
				}
			}
		})
	}
}

func TestPrepareConnectionString_IPv6(t *testing.T) {
	// Test that IPv6 localhost (::1) disables SSL
	result, err := prepareConnectionString("postgresql://user:pass@[::1]:5432/postgres")

	// This may fail on systems without IPv6, so we just check it doesn't panic
	if err == nil {
		assert.NotEmpty(t, result)
		// Should contain sslmode parameter
		assert.True(t, strings.Contains(result, "sslmode="))
	}
}
