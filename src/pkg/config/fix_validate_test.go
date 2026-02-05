// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package config

import (
	"strings"
	"testing"
)

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
		errMsg  string
	}{
		{"valid https URL", "https://example.com", false, ""},
		{"valid http URL", "http://localhost:8080", false, ""},
		{"missing scheme", "example.com", true, "scheme"},
		{"invalid scheme", "ftp://example.com", true, "http or https"},
		{"empty URL", "", true, "scheme"},
		{"no host", "https://", true, "host"},
		{"valid with path", "https://api.example.com/v1", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateURL(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("validateURL() error = %v, want error containing %q", err, tt.errMsg)
			}
		})
	}
}

func TestValidateDatabaseURL(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
		errMsg  string
	}{
		{"valid mongodb+srv", "mongodb+srv://user:pass@cluster.mongodb.net/db", false, ""},
		{"valid mongodb", "mongodb://user:pass@localhost:27017/db", false, ""},
		{"valid postgresql", "postgresql://user:pass@localhost:5432/db", false, ""},
		{"missing credentials", "mongodb://localhost/db", true, "credentials"},
		{"invalid prefix", "http://localhost/db", true, "must start with"},
		{"empty URL", "", true, "must start with"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDatabaseURL(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDatabaseURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("validateDatabaseURL() error = %v, want error containing %q", err, tt.errMsg)
			}
		})
	}
}

func TestValidateAPIKey(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
		errMsg  string
	}{
		{"valid with prefix", "sk-proj-abcdefghijklmnopqrstuvwxyz1234567890", false, ""},
		{"valid long key", "1234567890abcdefghij1234567890abcdefghij", false, ""},
		{"too short", "short", true, "too short"},
		{"too long", strings.Repeat("a", 300), true, "too long"},
		{"placeholder", "your-api-key-here-with-more-chars", true, "placeholder"},
		{"example placeholder", "api_key_example_1234567890", true, "placeholder"},
		{"test placeholder", "test-key-1234567890abcdefghij", true, "placeholder"},
		{"empty", "", true, "short"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAPIKey(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAPIKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("validateAPIKey() error = %v, want error containing %q", err, tt.errMsg)
			}
		})
	}
}

func TestValidatePort(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
		errMsg  string
	}{
		{"valid port", "8080", false, ""},
		{"port 80", "80", false, ""},
		{"port 443", "443", false, ""},
		{"max port", "65535", false, ""},
		{"min port", "1", false, ""},
		{"zero port", "0", true, "between 1 and 65535"},
		{"too high", "65536", true, "between 1 and 65535"},
		{"negative", "-1", true, "between 1 and 65535"},
		{"not a number", "abc", true, "must be a number"},
		{"empty", "", true, "must be a number"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePort(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("validatePort() error = %v, want error containing %q", err, tt.errMsg)
			}
		})
	}
}

func TestValidateTenantID(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
		errMsg  string
	}{
		{"valid simple", "tenant123", false, ""},
		{"valid with hyphens", "my-org-123", false, ""},
		{"valid with underscores", "my_tenant_id", false, ""},
		{"valid alphanumeric", "TenantABC123", false, ""},
		{"too short", "ab", true, "too short"},
		{"invalid chars", "tenant@123", true, "only contain"},
		{"placeholder", "your-tenant-id", true, "placeholder"},
		{"example placeholder", "tenant-example", true, "placeholder"},
		{"empty", "", true, "too short"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTenantID(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTenantID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("validateTenantID() error = %v, want error containing %q", err, tt.errMsg)
			}
		})
	}
}

func TestValidateHost(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
		errMsg  string
	}{
		{"localhost", "localhost", false, ""},
		{"ipv4", "127.0.0.1", false, ""},
		{"ipv6", "::1", false, ""},
		{"valid hostname", "example.com", false, ""},
		{"valid subdomain", "api.example.com", false, ""},
		{"valid with many parts", "sub.api.example.com", false, ""},
		{"empty", "", true, "empty"},
		{"whitespace only", "   ", true, "empty"},
		{"invalid chars", "host@name", true, "invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateHost(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateHost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("validateHost() error = %v, want error containing %q", err, tt.errMsg)
			}
		})
	}
}

func TestValidateFieldValue(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		value   string
		wantErr bool
	}{
		{"url field valid", "url", "https://example.com", false},
		{"url field invalid", "url", "not-a-url", true},
		{"api_key valid", "api_key", "sk-proj-1234567890abcdefghij", false},
		{"api_key short", "api_key", "short", true},
		{"port valid", "port", "8080", false},
		{"port invalid", "port", "99999", true},
		{"tenant valid", "tenant", "my-tenant-123", false},
		{"tenant placeholder", "tenant", "your-tenant-id", true},
		{"unknown field", "custom_field", "any-value", false},
		{"empty value", "any_field", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFieldValue(tt.field, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFieldValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetValidationHint(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		err      error
		contains string
	}{
		{"url scheme error", "url", validateURL("example.com"), "http://"},
		{"database_url error", "database_url", validateDatabaseURL("http://wrong"), "mongodb"},
		{"api_key placeholder", "api_key", validateAPIKey("your-api-key-here-with-more-chars"), "actual"},
		{"port error", "port", validatePort("99999"), "port"},
		{"no error", "field", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hint := GetValidationHint(tt.field, tt.err)
			if tt.contains != "" && !strings.Contains(hint, tt.contains) {
				t.Errorf("GetValidationHint() = %q, want to contain %q", hint, tt.contains)
			}
			if tt.err == nil && hint != "" {
				t.Errorf("GetValidationHint() = %q, want empty for no error", hint)
			}
		})
	}
}

func TestGetFieldSuggestions(t *testing.T) {
	tests := []struct {
		name            string
		field           string
		value           string
		wantSuggestions bool
		contains        string
	}{
		{"url missing scheme", "url", "example.com", true, "Did you mean: https://example.com"},
		{"url with http", "url", "http://api.example.com", true, "Did you mean:"},
		{"url with https", "url", "https://example.com", false, ""},
		{"mongodb atlas without srv", "database_url", "mongodb.net/db", true, "mongodb+srv://"},
		{"api key too short", "api_key", "sk-short", true, "OpenAI"},
		{"api key good length", "api_key", "sk-proj-1234567890abcdefghijklmnopqrstuvwxyz12345", false, ""},
		{"port 80", "port", "80", true, "443"},
		{"port 8080", "port", "8080", true, "development"},
		{"empty value", "field", "", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := GetFieldSuggestions(tt.field, tt.value)
			hasSuggestions := len(suggestions) > 0

			if hasSuggestions != tt.wantSuggestions {
				t.Errorf("GetFieldSuggestions() returned %d suggestions, want suggestions=%v", len(suggestions), tt.wantSuggestions)
			}

			if tt.contains != "" {
				found := false
				for _, suggestion := range suggestions {
					if strings.Contains(suggestion, tt.contains) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("GetFieldSuggestions() = %v, want to contain %q", suggestions, tt.contains)
				}
			}
		})
	}
}
