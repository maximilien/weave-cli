package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

// Format represents an output format type
type Format string

const (
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
	FormatTable Format = "table"
)

// Formatter handles formatting data for output
type Formatter struct {
	format Format
	writer io.Writer
}

// NewFormatter creates a new formatter
func NewFormatter(format Format, writer io.Writer) *Formatter {
	return &Formatter{
		format: format,
		writer: writer,
	}
}

// Format formats the data according to the configured format
func (f *Formatter) Format(data interface{}) error {
	switch f.format {
	case FormatJSON:
		return f.formatJSON(data)
	case FormatYAML:
		return f.formatYAML(data)
	case FormatTable:
		return f.formatTable(data)
	default:
		return fmt.Errorf("unsupported format: %s", f.format)
	}
}

// formatJSON formats data as JSON
func (f *Formatter) formatJSON(data interface{}) error {
	encoder := json.NewEncoder(f.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// formatYAML formats data as YAML
func (f *Formatter) formatYAML(data interface{}) error {
	encoder := yaml.NewEncoder(f.writer)
	encoder.SetIndent(2)
	defer encoder.Close()
	return encoder.Encode(data)
}

// formatTable formats data as a human-readable table
func (f *Formatter) formatTable(data interface{}) error {
	// Type assertion to map for table formatting
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		// Try to convert via JSON
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal data: %w", err)
		}
		if err := json.Unmarshal(jsonBytes, &dataMap); err != nil {
			return fmt.Errorf("failed to unmarshal data: %w", err)
		}
	}

	// Build table output
	var output strings.Builder

	// Find max key length for alignment
	maxKeyLen := 0
	for key := range dataMap {
		if len(key) > maxKeyLen {
			maxKeyLen = len(key)
		}
	}

	// Write each key-value pair
	for key, value := range dataMap {
		// Format key with padding
		paddedKey := key + strings.Repeat(" ", maxKeyLen-len(key))

		// Format value based on type
		var formattedValue string
		switch v := value.(type) {
		case []interface{}:
			if len(v) == 0 {
				formattedValue = "[]"
			} else {
				// Format array as multi-line with indentation
				formattedValue = "\n"
				for i, item := range v {
					formattedValue += fmt.Sprintf("  [%d] %v\n", i, formatValue(item))
				}
				formattedValue = strings.TrimSuffix(formattedValue, "\n")
			}
		case map[string]interface{}:
			if len(v) == 0 {
				formattedValue = "{}"
			} else {
				// Format nested map as multi-line
				formattedValue = "\n"
				for k, val := range v {
					formattedValue += fmt.Sprintf("  %s: %v\n", k, formatValue(val))
				}
				formattedValue = strings.TrimSuffix(formattedValue, "\n")
			}
		default:
			formattedValue = fmt.Sprintf("%v", value)
		}

		output.WriteString(fmt.Sprintf("%s: %s\n", paddedKey, formattedValue))
	}

	_, err := f.writer.Write([]byte(output.String()))
	return err
}

// formatValue formats a value for table display
func formatValue(value interface{}) string {
	switch v := value.(type) {
	case map[string]interface{}:
		// Format nested map inline
		var parts []string
		for k, val := range v {
			parts = append(parts, fmt.Sprintf("%s=%v", k, val))
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case []interface{}:
		// Format array inline
		var parts []string
		for _, item := range v {
			parts = append(parts, fmt.Sprintf("%v", item))
		}
		return "[" + strings.Join(parts, ", ") + "]"
	default:
		return fmt.Sprintf("%v", value)
	}
}

// FormatIngestReport formats an ingestion report
func FormatIngestReport(report interface{}, format Format, writer io.Writer) error {
	formatter := NewFormatter(format, writer)
	return formatter.Format(report)
}
