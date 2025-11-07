// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"golang.org/x/term"
)

// ConfigError represents a configuration error with helpful information
type ConfigError struct {
	Message          string
	MissingVars      []string
	EnvFileExists    bool
	ConfigFileExists bool
	Tips             []string
}

// Error implements the error interface
func (e *ConfigError) Error() string {
	return e.Message
}

// CheckRequiredEnvVars checks if required environment variables are set
func CheckRequiredEnvVars() *ConfigError {
	requiredVars := []string{
		"WEAVIATE_URL",
		"WEAVIATE_API_KEY",
		"OPENAI_API_KEY",
	}

	var missing []string
	for _, v := range requiredVars {
		if os.Getenv(v) == "" {
			missing = append(missing, v)
		}
	}

	if len(missing) == 0 {
		return nil
	}

	// Check if .env file exists
	envExists := false
	if _, err := os.Stat(".env"); err == nil {
		envExists = true
	}

	// Check if config.yaml exists
	configExists := false
	if _, err := os.Stat("config.yaml"); err == nil {
		configExists = true
	}

	return &ConfigError{
		Message:          "Missing required configuration",
		MissingVars:      missing,
		EnvFileExists:    envExists,
		ConfigFileExists: configExists,
	}
}

// CheckREPLRequiredEnvVars checks if REPL-specific environment variables are set
func CheckREPLRequiredEnvVars() *ConfigError {
	requiredVars := []string{
		"WEAVIATE_URL",
		"WEAVIATE_API_KEY",
		"OPENAI_API_KEY",
		"WEAVE_MCP_STDIO_PATH",
	}

	var missing []string
	for _, v := range requiredVars {
		if os.Getenv(v) == "" {
			missing = append(missing, v)
		}
	}

	if len(missing) == 0 {
		return nil
	}

	// Check if .env file exists
	envExists := false
	if _, err := os.Stat(".env"); err == nil {
		envExists = true
	}

	// Check if config.yaml exists
	configExists := false
	if _, err := os.Stat("config.yaml"); err == nil {
		configExists = true
	}

	return &ConfigError{
		Message:          "Missing required REPL configuration",
		MissingVars:      missing,
		EnvFileExists:    envExists,
		ConfigFileExists: configExists,
	}
}

// FormatConfigError formats a configuration error with helpful tips
func FormatConfigError(err error) string {
	if err == nil {
		return ""
	}

	// Check if it's a ConfigError
	if configErr, ok := err.(*ConfigError); ok {
		return formatDetailedConfigError(configErr)
	}

	// Check if error message indicates missing configuration
	errMsg := err.Error()

	// Check for MCP/REPL-specific errors
	if strings.Contains(errMsg, "failed to create MCP client") ||
		strings.Contains(errMsg, "WEAVE_MCP_STDIO_PATH must be configured") ||
		strings.Contains(errMsg, "failed to initialize MCP connection") ||
		strings.Contains(errMsg, "failed to read initialize response") {
		// This is a REPL-specific configuration issue
		if configErr := CheckREPLRequiredEnvVars(); configErr != nil {
			return formatREPLConfigError(configErr)
		}
		// Even if env vars are set, provide helpful MCP error message
		return formatMCPConnectionError(errMsg)
	}

	// Check for general configuration errors
	if strings.Contains(errMsg, "no vector databases configured") ||
		strings.Contains(errMsg, "OPENAI_API_KEY environment variable is required") ||
		strings.Contains(errMsg, "WEAVIATE_API_KEY") && strings.Contains(errMsg, "required") ||
		strings.Contains(errMsg, "WEAVIATE_URL") && strings.Contains(errMsg, "required") ||
		strings.Contains(errMsg, "collection") && strings.Contains(errMsg, "not found") && strings.Contains(errMsg, "check database configuration") {
		// This might be a configuration issue, check env vars
		if configErr := CheckRequiredEnvVars(); configErr != nil {
			return formatDetailedConfigError(configErr)
		}
	}

	// Return original error if not a config issue
	return errMsg
}

// formatDetailedConfigError formats a detailed configuration error message
func formatDetailedConfigError(err *ConfigError) string {
	var sb strings.Builder

	// Title
	sb.WriteString("❌ Configuration Error: Missing required information\n\n")

	// Missing variables
	if len(err.MissingVars) > 0 {
		sb.WriteString("Missing environment variables:\n")
		for _, v := range err.MissingVars {
			sb.WriteString(fmt.Sprintf("  • %s\n", v))
		}
		sb.WriteString("\n")
	}

	// How to fix
	sb.WriteString("How to fix this:\n\n")
	sb.WriteString("Option 1: Use command-line flags\n")
	sb.WriteString("  weave docs ls COLLECTION --weaviate-url=\"https://your-cluster.weaviate.cloud\" \\\n")
	sb.WriteString("                           --weaviate-api-key=\"your-api-key\" \\\n")
	sb.WriteString("                           --vector-db-type=\"weaviate-cloud\"\n\n")

	sb.WriteString("Option 2: Set environment variables in your shell\n")
	for _, v := range err.MissingVars {
		example := getEnvVarExample(v)
		sb.WriteString(fmt.Sprintf("  export %s=\"%s\"\n", v, example))
	}
	sb.WriteString("\n")

	if err.EnvFileExists {
		sb.WriteString("Option 3: Update your existing .env file\n")
		sb.WriteString("  Run: weave config update --env\n\n")
	} else {
		sb.WriteString("Option 3: Create a .env file\n")
		sb.WriteString("  Run: weave config create --env\n\n")
	}

	// Config.yaml tips
	if !err.ConfigFileExists {
		sb.WriteString("💡 Tip: You can also create a config.yaml file for more control.\n")
		sb.WriteString("   See: https://github.com/maximilien/weave-cli for examples\n")
		sb.WriteString("   Or copy: config.yaml.example to config.yaml and customize it\n\n")
	}

	// Mock database tip for testing
	sb.WriteString("For testing without Weaviate:\n")
	sb.WriteString("  export VECTOR_DB_TYPE=\"mock\"\n\n")

	// Additional help
	sb.WriteString("For more help:\n")
	sb.WriteString("  weave config show    # Show current configuration\n")
	sb.WriteString("  weave --help         # Show all available commands\n")

	return sb.String()
}

// formatREPLConfigError formats a detailed configuration error message for REPL
func formatREPLConfigError(err *ConfigError) string {
	var sb strings.Builder

	// Title
	sb.WriteString("❌ REPL Configuration Error: Missing required information\n\n")

	// Missing variables
	if len(err.MissingVars) > 0 {
		sb.WriteString("Missing environment variables for REPL mode:\n")
		for _, v := range err.MissingVars {
			sb.WriteString(fmt.Sprintf("  • %s\n", v))
		}
		sb.WriteString("\n")
	}

	// Special note about WEAVE_MCP_STDIO_PATH
	if containsString(err.MissingVars, "WEAVE_MCP_STDIO_PATH") {
		sb.WriteString("⚠️  Note: REPL mode requires the weave-mcp server binary.\n\n")
		sb.WriteString("💡 Quick fix - Install weave-mcp with one command:\n")
		sb.WriteString("   weave config update --weave-mcp\n\n")
		sb.WriteString("   This will automatically download and install the binary for your platform.\n")
		sb.WriteString("   See: https://github.com/maximilien/weave-mcp for more information\n\n")
	}

	// How to fix
	sb.WriteString("How to fix this:\n\n")

	// Option 1: Install weave-mcp (if missing)
	if containsString(err.MissingVars, "WEAVE_MCP_STDIO_PATH") {
		sb.WriteString("Option 1: Install weave-mcp binary (recommended)\n")
		sb.WriteString("  Run: weave config update --weave-mcp\n\n")
	}

	// Option 2: Set environment variables
	optionNum := 1
	if containsString(err.MissingVars, "WEAVE_MCP_STDIO_PATH") {
		optionNum = 2
	}
	sb.WriteString(fmt.Sprintf("Option %d: Set environment variables in your shell\n", optionNum))
	for _, v := range err.MissingVars {
		example := getEnvVarExample(v)
		sb.WriteString(fmt.Sprintf("  export %s=\"%s\"\n", v, example))
	}
	sb.WriteString("\n")

	// Option 3: Update .env file
	optionNum++
	if err.EnvFileExists {
		sb.WriteString(fmt.Sprintf("Option %d: Update your existing .env file\n", optionNum))
		sb.WriteString("  Run: weave config update --env\n\n")
	} else {
		sb.WriteString(fmt.Sprintf("Option %d: Create a .env file\n", optionNum))
		sb.WriteString("  Run: weave config create --env\n\n")
	}

	// Additional help
	sb.WriteString("For more help:\n")
	sb.WriteString("  weave config show    # Show current configuration\n")
	sb.WriteString("  weave --help         # Show all available commands\n")

	return sb.String()
}

// formatMCPConnectionError formats an MCP connection error with helpful troubleshooting tips
func formatMCPConnectionError(errMsg string) string {
	var sb strings.Builder

	// Title
	sb.WriteString("❌ REPL MCP Connection Error\n\n")

	// Check if stderr contains config.yaml error
	if strings.Contains(errMsg, "config.yaml: no such file or directory") {
		sb.WriteString("⚠️  weave-mcp failed because it couldn't find config.yaml\n\n")
		sb.WriteString("The weave-mcp server requires a config.yaml file to run.\n")
		sb.WriteString("This is different from weave-cli which can work without config.yaml.\n\n")

		// Try to create config.yaml automatically
		if err := createMinimalConfigYAML(); err == nil {
			sb.WriteString("✅ Created minimal config.yaml for you!\n\n")
			sb.WriteString("Run 'weave' again to start the REPL.\n\n")
			sb.WriteString("Note: The config.yaml uses environment variables from your .env file.\n")
			sb.WriteString("You can customize it later. See config.yaml.example for all options.\n")
		} else {
			// Fallback to manual instructions if auto-creation fails
			sb.WriteString("Quick fix - Create a minimal config.yaml:\n\n")
			sb.WriteString("  cat > config.yaml << 'EOF'\n")
			sb.WriteString(getMinimalConfigYAMLContent())
			sb.WriteString("EOF\n\n")

			sb.WriteString("Then run 'weave' again.\n\n")
			sb.WriteString(fmt.Sprintf("Note: Auto-creation failed: %v\n", err))
		}

		sb.WriteString("\nFor more information:\n")
		sb.WriteString("  • Full example: https://github.com/maximilien/weave-cli/blob/main/config.yaml.example\n")
		sb.WriteString("  • weave-mcp repo: https://github.com/maximilien/weave-mcp\n")

		return sb.String()
	}

	// Generic MCP error handling
	sb.WriteString(fmt.Sprintf("Error: %s\n\n", errMsg))

	// Troubleshooting steps
	sb.WriteString("Troubleshooting steps:\n\n")

	sb.WriteString("1. Verify weave-mcp binary path:\n")
	mcpPath := os.Getenv("WEAVE_MCP_STDIO_PATH")
	if mcpPath != "" {
		sb.WriteString(fmt.Sprintf("   Current path: %s\n", mcpPath))
		if _, err := os.Stat(mcpPath); os.IsNotExist(err) {
			sb.WriteString("   ⚠️  Binary not found at this path!\n")
		} else if err != nil {
			sb.WriteString(fmt.Sprintf("   ⚠️  Cannot access binary: %v\n", err))
		} else {
			sb.WriteString("   ✓ Binary exists\n")
		}
	} else {
		sb.WriteString("   ⚠️  WEAVE_MCP_STDIO_PATH not set\n")
	}
	sb.WriteString("\n")

	sb.WriteString("2. Check if weave-mcp is installed:\n")
	sb.WriteString("   git clone https://github.com/maximilien/weave-mcp.git\n")
	sb.WriteString("   cd weave-mcp && go build -o bin/weave-mcp-stdio cmd/stdio/main.go\n\n")

	sb.WriteString("3. Verify the binary is executable:\n")
	if mcpPath != "" {
		sb.WriteString(fmt.Sprintf("   chmod +x %s\n", mcpPath))
		sb.WriteString(fmt.Sprintf("   %s --version\n\n", mcpPath))
	} else {
		sb.WriteString("   chmod +x /path/to/weave-mcp/bin/weave-mcp-stdio\n")
		sb.WriteString("   /path/to/weave-mcp/bin/weave-mcp-stdio --version\n\n")
	}

	sb.WriteString("4. Check environment variables are set:\n")
	sb.WriteString("   echo $WEAVE_MCP_STDIO_PATH\n")
	sb.WriteString("   echo $OPENAI_API_KEY\n")
	sb.WriteString("   echo $WEAVIATE_URL\n")
	sb.WriteString("   echo $WEAVIATE_API_KEY\n\n")

	sb.WriteString("For more information:\n")
	sb.WriteString("  • weave-mcp repo: https://github.com/maximilien/weave-mcp\n")
	sb.WriteString("  • weave config show    # Show current configuration\n")

	return sb.String()
}

// containsString checks if a string is in a slice
func containsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// getEnvVarExample returns an example value for an environment variable
func getEnvVarExample(varName string) string {
	examples := map[string]string{
		"WEAVIATE_URL":         "https://your-cluster-id.weaviate.cloud",
		"WEAVIATE_API_KEY":     "your-weaviate-api-key",
		"OPENAI_API_KEY":       "sk-proj-your-openai-api-key",
		"VECTOR_DB_TYPE":       "weaviate-cloud",
		"WEAVE_MCP_STDIO_PATH": "/path/to/weave-mcp/bin/weave-mcp-stdio",
	}
	if example, ok := examples[varName]; ok {
		return example
	}
	return "your-value-here"
}

// PromptToFixConfig prompts the user to fix configuration interactively
func PromptToFixConfig(err *ConfigError) (bool, error) {
	// Don't prompt if not in a terminal
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return false, nil
	}

	// Print error details first
	fmt.Println(formatDetailedConfigError(err))

	// Ask if user wants to fix now
	fmt.Println()
	if err.EnvFileExists {
		color.New(color.FgCyan, color.Bold).Print("Would you like to update your .env file now? (Y/n): ")
	} else {
		color.New(color.FgCyan, color.Bold).Print("Would you like to create a .env file now? (Y/n): ")
	}

	reader := bufio.NewReader(os.Stdin)
	response, readErr := reader.ReadString('\n')
	if readErr != nil {
		return false, readErr
	}

	response = strings.ToLower(strings.TrimSpace(response))
	if response == "" || response == "y" || response == "yes" {
		return true, nil
	}

	return false, nil
}

// InteractiveConfigFix runs the interactive configuration fix process
func InteractiveConfigFix(envFileExists bool) error {
	fmt.Println()
	color.New(color.FgGreen, color.Bold).Println("🔧 Interactive Configuration Setup")
	fmt.Println()

	// Collect values
	values := make(map[string]string)
	envVars := getRequiredEnvVars()

	for _, envVar := range envVars {
		// Load existing value if available
		existingValue := os.Getenv(envVar.Key)
		if existingValue == "" && envFileExists {
			// Try to load from .env file
			envValues := loadEnvFileSimple(".env")
			existingValue = envValues[envVar.Key]
		}

		// Prompt for value
		if envVar.IsSecret {
			fmt.Printf("%s:\n", envVar.Description)
			if existingValue != "" {
				fmt.Printf("  (Current: %s)\n", maskSecret(existingValue))
			}
			fmt.Printf("  Enter value (or press Enter to keep current): ")

			value, err := readSecretSimple()
			if err != nil {
				return fmt.Errorf("failed to read secret: %w", err)
			}

			if value == "" && existingValue != "" {
				value = existingValue
			}
			values[envVar.Key] = value
		} else {
			fmt.Printf("%s:\n", envVar.Description)
			if existingValue != "" {
				fmt.Printf("  (Current: %s)\n", existingValue)
			}
			fmt.Printf("  Example: %s\n", envVar.Example)
			fmt.Printf("  Enter value (or press Enter to keep current): ")

			reader := bufio.NewReader(os.Stdin)
			value, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}
			value = strings.TrimSpace(value)

			if value == "" && existingValue != "" {
				value = existingValue
			}
			values[envVar.Key] = value
		}
		fmt.Println()
	}

	// Show summary
	fmt.Println()
	color.New(color.FgBlue, color.Bold).Println("📋 Configuration Summary:")
	fmt.Println()
	for _, envVar := range envVars {
		value := values[envVar.Key]
		if value != "" {
			displayValue := value
			if envVar.IsSecret {
				displayValue = maskSecret(value)
			}
			fmt.Printf("  %s: %s\n", envVar.Key, displayValue)
		}
	}

	// Confirm save
	fmt.Println()
	color.New(color.FgYellow, color.Bold).Printf("💾 Save changes to .env? (Y/n): ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read confirmation: %w", err)
	}

	response = strings.ToLower(strings.TrimSpace(response))
	if response != "" && response != "y" && response != "yes" {
		fmt.Println("Configuration not saved.")
		return nil
	}

	// Save to .env file
	if err := saveEnvFileSimple(".env", values); err != nil {
		return fmt.Errorf("failed to save .env file: %w", err)
	}

	color.New(color.FgGreen, color.Bold).Println("\n✅ Configuration saved successfully!")
	fmt.Println("\nYou can now run your command again.")

	return nil
}

// Helper types and functions

type requiredEnvVar struct {
	Key         string
	Description string
	Example     string
	IsSecret    bool
}

func getRequiredEnvVars() []requiredEnvVar {
	return []requiredEnvVar{
		{
			Key:         "WEAVIATE_URL",
			Description: "Weaviate Cloud URL",
			Example:     "https://your-cluster-id.weaviate.cloud",
			IsSecret:    false,
		},
		{
			Key:         "WEAVIATE_API_KEY",
			Description: "Weaviate API Key",
			Example:     "your-weaviate-api-key",
			IsSecret:    true,
		},
		{
			Key:         "OPENAI_API_KEY",
			Description: "OpenAI API Key",
			Example:     "sk-proj-your-openai-api-key",
			IsSecret:    true,
		},
	}
}

func maskSecret(s string) string {
	if len(s) <= 8 {
		return "***"
	}
	return s[:4] + "..." + s[len(s)-4:]
}

func readSecretSimple() (string, error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(line), nil
	}

	password, err := term.ReadPassword(fd)
	if err != nil {
		return "", err
	}
	fmt.Println()
	return string(password), nil
}

func loadEnvFileSimple(filename string) map[string]string {
	values := make(map[string]string)
	file, err := os.Open(filename)
	if err != nil {
		return values
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			value = strings.Trim(value, "\"'")
			values[key] = value
		}
	}
	return values
}

func saveEnvFileSimple(filename string, values map[string]string) error {
	// Try to read .env.example first
	exampleFile := ".env.example"
	exampleContent, err := os.ReadFile(exampleFile)
	if err != nil {
		// If .env.example doesn't exist, create a simple .env file
		return createSimpleEnvFile(filename, values)
	}

	// Process the example file line by line
	lines := strings.Split(string(exampleContent), "\n")
	var output []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Keep comments and empty lines
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			output = append(output, line)
			continue
		}

		// Parse KEY=VALUE
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])

			// Use updated value if available
			if value, ok := values[key]; ok && value != "" {
				output = append(output, fmt.Sprintf("%s=\"%s\"", key, value))
			} else {
				// Keep original line if no value provided
				output = append(output, line)
			}
		} else {
			output = append(output, line)
		}
	}

	// Write to file
	content := strings.Join(output, "\n")
	if err := os.WriteFile(filename, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func createSimpleEnvFile(filename string, values map[string]string) error {
	var sb strings.Builder
	sb.WriteString("# Weave CLI Configuration\n")
	sb.WriteString("# Generated by weave config\n\n")

	envVars := getRequiredEnvVars()
	for _, envVar := range envVars {
		if value, ok := values[envVar.Key]; ok && value != "" {
			sb.WriteString(fmt.Sprintf("# %s\n", envVar.Description))
			sb.WriteString(fmt.Sprintf("%s=\"%s\"\n\n", envVar.Key, value))
		}
	}

	return os.WriteFile(filename, []byte(sb.String()), 0600)
}

// getMinimalConfigYAMLContent returns the content for a minimal config.yaml
func getMinimalConfigYAMLContent() string {
	return `databases:
  default: weaviate-cloud
  vector_databases:
    - name: weaviate-cloud
      type: weaviate-cloud
      url: ${WEAVIATE_URL}
      api_key: ${WEAVIATE_API_KEY}
      openai_api_key: ${OPENAI_API_KEY}
`
}

// createMinimalConfigYAML creates a minimal config.yaml file in the current directory
func createMinimalConfigYAML() error {
	// Check if config.yaml already exists
	if _, err := os.Stat("config.yaml"); err == nil {
		return fmt.Errorf("config.yaml already exists")
	}

	// Create the file
	content := getMinimalConfigYAMLContent()
	if err := os.WriteFile("config.yaml", []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write config.yaml: %w", err)
	}

	return nil
}
