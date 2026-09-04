// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package config

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// ValidateFieldValue validates a field value based on its name and type
func ValidateFieldValue(field, value string) error {
	if value == "" {
		return fmt.Errorf("value cannot be empty")
	}

	fieldLower := strings.ToLower(field)

	// URL validation
	if strings.Contains(fieldLower, "url") {
		return validateURL(value)
	}

	// Database URL validation (MongoDB specific)
	if strings.Contains(fieldLower, "database_url") {
		return validateDatabaseURL(value)
	}

	// API key validation
	if strings.Contains(fieldLower, "api_key") || strings.Contains(fieldLower, "apikey") {
		return validateAPIKey(value)
	}

	// Token validation
	if strings.Contains(fieldLower, "token") {
		return validateToken(value)
	}

	// Port validation
	if strings.Contains(fieldLower, "port") {
		return validatePort(value)
	}

	// Tenant ID validation
	if strings.Contains(fieldLower, "tenant") {
		return validateTenantID(value)
	}

	// Host validation
	if strings.Contains(fieldLower, "host") {
		return validateHost(value)
	}

	return nil // No specific validation needed
}

// validateURL validates a URL format
func validateURL(value string) error {
	parsed, err := url.Parse(value)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Check scheme
	if parsed.Scheme == "" {
		return fmt.Errorf("URL must include scheme (http:// or https://)")
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("URL scheme must be http or https, got: %s", parsed.Scheme)
	}

	// Check host
	if parsed.Host == "" {
		return fmt.Errorf("URL must include a host")
	}

	return nil
}

// validateDatabaseURL validates database connection URLs (MongoDB, PostgreSQL, etc.)
func validateDatabaseURL(value string) error {
	// Support various database URL formats
	validPrefixes := []string{
		"mongodb://",
		"mongodb+srv://",
		"postgresql://",
		"postgres://",
		"mysql://",
	}

	hasValidPrefix := false
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(value, prefix) {
			hasValidPrefix = true
			break
		}
	}

	if !hasValidPrefix {
		return fmt.Errorf("database URL must start with: %s", strings.Join(validPrefixes, ", "))
	}

	// Check for credentials (most DB URLs need them)
	if !strings.Contains(value, "@") {
		return fmt.Errorf("database URL should include credentials (user:pass@host)")
	}

	return nil
}

// validateAPIKey validates API key format
func validateAPIKey(value string) error {
	// Check minimum length
	if len(value) < 20 {
		return fmt.Errorf("API key seems too short (expected at least 20 characters)")
	}

	// Check maximum length (reasonable upper bound)
	if len(value) > 256 {
		return fmt.Errorf("API key seems too long (expected at most 256 characters)")
	}

	// Check it's not a placeholder
	placeholders := []string{
		"your-api-key",
		"your_api_key",
		"api-key-here",
		"replace-me",
		"changeme",
		"example",
		"test",
	}
	valueLower := strings.ToLower(value)
	for _, placeholder := range placeholders {
		if strings.Contains(valueLower, placeholder) {
			return fmt.Errorf("API key appears to be a placeholder ('%s')", placeholder)
		}
	}

	return nil
}

// validateToken validates token format
func validateToken(value string) error {
	// Similar to API key but more lenient
	if len(value) < 16 {
		return fmt.Errorf("token seems too short (expected at least 16 characters)")
	}

	if len(value) > 512 {
		return fmt.Errorf("token seems too long (expected at most 512 characters)")
	}

	// Check it's not a placeholder
	placeholders := []string{
		"your-token",
		"token-here",
		"replace",
		"example",
	}
	valueLower := strings.ToLower(value)
	for _, placeholder := range placeholders {
		if strings.Contains(valueLower, placeholder) {
			return fmt.Errorf("token appears to be a placeholder")
		}
	}

	return nil
}

// validatePort validates port number
func validatePort(value string) error {
	port, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("port must be a number")
	}

	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got: %d", port)
	}

	return nil
}

// validateTenantID validates tenant ID format
func validateTenantID(value string) error {
	if len(value) < 3 {
		return fmt.Errorf("tenant ID seems too short (expected at least 3 characters)")
	}

	// Check for valid characters (alphanumeric, hyphens, underscores)
	validTenantID := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validTenantID.MatchString(value) {
		return fmt.Errorf("tenant ID can only contain letters, numbers, hyphens, and underscores")
	}

	// Check it's not a placeholder
	placeholders := []string{
		"your-tenant",
		"tenant-id",
		"example",
		"test",
	}
	valueLower := strings.ToLower(value)
	for _, placeholder := range placeholders {
		if strings.Contains(valueLower, placeholder) {
			return fmt.Errorf("tenant ID appears to be a placeholder")
		}
	}

	return nil
}

// validateHost validates hostname format
func validateHost(value string) error {
	// Check it's not empty
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("host cannot be empty")
	}

	// Allow localhost
	if value == "localhost" || value == "127.0.0.1" || value == "::1" {
		return nil
	}

	// Check for valid hostname pattern
	validHost := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)
	if !validHost.MatchString(value) {
		return fmt.Errorf("invalid hostname format")
	}

	return nil
}

// GetValidationHint returns a helpful hint for validation errors
func GetValidationHint(field string, err error) string {
	if err == nil {
		return ""
	}

	fieldLower := strings.ToLower(field)
	errMsg := err.Error()

	// URL hints
	if strings.Contains(fieldLower, "url") {
		if strings.Contains(errMsg, "scheme") {
			return "💡 URLs must start with http:// or https://\n   Try: https://your-instance.cloud-provider.com"
		}
		if strings.Contains(errMsg, "host") {
			return "💡 Example: https://your-instance.cloud-provider.com"
		}
	}

	// Database URL hints
	if strings.Contains(fieldLower, "database_url") {
		return "💡 MongoDB Atlas example:\n   mongodb+srv://username:password@cluster.mongodb.net/database\n   💡 Get connection string from Atlas dashboard → Database → Connect"
	}

	// API key hints
	if strings.Contains(fieldLower, "api_key") {
		if strings.Contains(errMsg, "placeholder") {
			return "💡 Use your actual API key from the provider's dashboard\n   ⚠️  Don't use example values like 'your-api-key-here'"
		}
		if strings.Contains(errMsg, "short") {
			return "💡 Make sure you copied the complete API key\n   🔍 Check for truncation - API keys are usually 32+ characters"
		}
	}

	// Port hints
	if strings.Contains(fieldLower, "port") {
		return "💡 Common ports:\n   • 8080 (HTTP dev server)\n   • 443 (HTTPS)\n   • 27017 (MongoDB)\n   • 6333 (Qdrant)"
	}

	// Tenant hints
	if strings.Contains(fieldLower, "tenant") {
		if strings.Contains(errMsg, "placeholder") {
			return "💡 Use your actual tenant ID from the provider\n   📍 Usually found in Settings → Organization or Account"
		}
	}

	return ""
}

// GetFieldSuggestions returns smart suggestions based on common mistakes
func GetFieldSuggestions(field, value string) []string {
	if value == "" {
		return nil
	}

	var suggestions []string
	fieldLower := strings.ToLower(field)
	valueLower := strings.ToLower(value)

	// URL suggestions
	if strings.Contains(fieldLower, "url") {
		// Missing scheme
		if !strings.HasPrefix(value, "http://") && !strings.HasPrefix(value, "https://") {
			if strings.Contains(value, ".") {
				suggestions = append(suggestions, fmt.Sprintf("Did you mean: https://%s", value))
			}
		}
		// Wrong scheme
		if strings.HasPrefix(value, "http://") && !strings.Contains(value, "localhost") {
			httpsVersion := strings.Replace(value, "http://", "https://", 1)
			suggestions = append(suggestions, fmt.Sprintf("Did you mean: %s (HTTPS is more secure)", httpsVersion))
		}
	}

	// Database URL suggestions
	if strings.Contains(fieldLower, "database_url") {
		if strings.Contains(value, "mongodb.net") && !strings.HasPrefix(value, "mongodb+srv://") {
			suggestions = append(suggestions, "💡 MongoDB Atlas requires mongodb+srv:// prefix")
		}
		if strings.Contains(value, "mongodb://") && !strings.Contains(value, "@") {
			suggestions = append(suggestions, "💡 MongoDB URLs need credentials: mongodb://user:pass@host/db")
		}
	}

	// API key suggestions
	if strings.Contains(fieldLower, "api_key") {
		if len(value) > 0 && len(value) < 20 {
			suggestions = append(suggestions, "⚠️  This looks too short for an API key - make sure you copied it completely")
		}
		if strings.Contains(valueLower, "sk-") && len(value) < 32 {
			suggestions = append(suggestions, "⚠️  OpenAI API keys starting with sk- are usually 51+ characters")
		}
	}

	// Port suggestions
	if strings.Contains(fieldLower, "port") {
		if value == "80" {
			suggestions = append(suggestions, "💡 Port 80 is HTTP - use 443 for HTTPS in production")
		}
		if value == "8080" {
			suggestions = append(suggestions, "✓ Port 8080 is common for development servers")
		}
	}

	return suggestions
}
