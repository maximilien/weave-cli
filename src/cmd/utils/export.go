// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/maximilien/weave-cli/src/pkg/weaviate"
	"gopkg.in/yaml.v3"
)

// CollectionExport represents the combined schema and metadata for export
type CollectionExport struct {
	Name     string                       `json:"name" yaml:"name"`
	Schema   *weaviate.CollectionSchema   `json:"schema,omitempty" yaml:"schema,omitempty"`
	Metadata map[string]MetadataFieldInfo `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// MetadataFieldInfo represents metadata field information
type MetadataFieldInfo struct {
	Type        string                 `json:"type" yaml:"type"`
	Occurrences int                    `json:"occurrences,omitempty" yaml:"occurrences,omitempty"`
	Sample      interface{}            `json:"sample,omitempty" yaml:"sample,omitempty"`
	JSONSchema  map[string]interface{} `json:"json_schema,omitempty" yaml:"json_schema,omitempty"`
}

// ExportCollectionSchemaAndMetadata exports collection schema and metadata
func ExportCollectionSchemaAndMetadata(ctx context.Context, client *weaviate.Client, collectionName string, includeMetadata bool, expandMetadata bool, compact bool) (*CollectionExport, error) {
	export := &CollectionExport{
		Name: collectionName,
	}

	// Get schema
	schema, err := client.GetFullCollectionSchema(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection schema: %w", err)
	}

	// Apply compact mode if requested
	if compact {
		schema = compactSchema(schema)
	}

	export.Schema = schema

	// Get documents for JSON field analysis and metadata
	var jsonSchemas map[string]map[string]interface{}
	documents, err := client.ListDocuments(ctx, collectionName, 100)
	if err == nil && len(documents) > 0 {
		// Analyze JSON fields across all documents
		jsonSchemas = analyzeJSONFields(documents, ctx, client, collectionName)

		// Update schema properties with JSON structure info
		for i := range export.Schema.Properties {
			prop := &export.Schema.Properties[i]
			if jsonSchema, isJSON := jsonSchemas[prop.Name]; isJSON {
				prop.JSONSchema = jsonSchema
			}
		}
	}

	// Get metadata if requested
	if includeMetadata && len(documents) > 0 {
		metadataFields := make(map[string]int)
		metadataTypes := make(map[string]string)
		metadataSamples := make(map[string]interface{})

		for _, doc := range documents {
			fullDoc, err := client.GetDocument(ctx, collectionName, doc.ID)
			if err != nil {
				continue
			}

			for key, value := range fullDoc.Metadata {
				metadataFields[key]++
				if metadataTypes[key] == "" {
					// Check if this is a JSON field
					if _, isJSON := jsonSchemas[key]; isJSON {
						metadataTypes[key] = "json"
					} else {
						metadataTypes[key] = fmt.Sprintf("%T", value)
					}
					if expandMetadata {
						metadataSamples[key] = value
					}
				}
			}
		}

		export.Metadata = make(map[string]MetadataFieldInfo)
		for name, count := range metadataFields {
			info := MetadataFieldInfo{
				Type:        metadataTypes[name],
				Occurrences: count,
			}
			if expandMetadata {
				info.Sample = metadataSamples[name]
			}
			// Add JSON schema if this is a JSON field
			if jsonSchema, isJSON := jsonSchemas[name]; isJSON {
				info.JSONSchema = jsonSchema
			}
			export.Metadata[name] = info
		}
	}

	// In compact mode, keep metadata and JSON schema but remove occurrences and samples
	if compact && export.Metadata != nil {
		compactMetadata := make(map[string]MetadataFieldInfo)
		for name, info := range export.Metadata {
			compactMetadata[name] = MetadataFieldInfo{
				Type:       info.Type,
				JSONSchema: info.JSONSchema,
				// Omit Occurrences and Sample in compact mode
			}
		}
		export.Metadata = compactMetadata
	}

	return export, nil
}

// compactSchema removes empty nested properties from schema
func compactSchema(schema *weaviate.CollectionSchema) *weaviate.CollectionSchema {
	if schema == nil {
		return schema
	}

	// Create a new schema with compacted properties
	compactedSchema := &weaviate.CollectionSchema{
		Class:      schema.Class,
		Vectorizer: schema.Vectorizer,
		Properties: make([]weaviate.SchemaProperty, 0, len(schema.Properties)),
	}

	for _, prop := range schema.Properties {
		compactedProp := compactSchemaProperty(prop)
		compactedSchema.Properties = append(compactedSchema.Properties, compactedProp)
	}

	return compactedSchema
}

// compactSchemaProperty recursively removes empty nested properties
func compactSchemaProperty(prop weaviate.SchemaProperty) weaviate.SchemaProperty {
	compactedProp := weaviate.SchemaProperty{
		Name:        prop.Name,
		DataType:    prop.DataType,
		Description: prop.Description,
	}

	// Only include nested properties if they exist and are non-empty
	if len(prop.NestedProperties) > 0 {
		compactedNested := make([]weaviate.SchemaProperty, 0, len(prop.NestedProperties))
		for _, nested := range prop.NestedProperties {
			compactedNested = append(compactedNested, compactSchemaProperty(nested))
		}
		compactedProp.NestedProperties = compactedNested
	} else {
		// Explicitly set to nil so omitempty works in YAML
		compactedProp.NestedProperties = nil
	}

	return compactedProp
}

// ExportAsYAML exports collection data as YAML with compact formatting
func ExportAsYAML(export *CollectionExport) (string, error) {
	// Marshal to YAML first
	data, err := yaml.Marshal(export)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to YAML: %w", err)
	}

	// Post-process to make datatype arrays more compact
	result := compactYAMLArrays(string(data))
	return result, nil
}

// compactYAMLArrays converts multi-line YAML arrays to inline format and simplifies metadata
func compactYAMLArrays(yamlStr string) string {
	lines := strings.Split(yamlStr, "\n")
	result := make([]string, 0, len(lines))
	i := 0

	for i < len(lines) {
		line := lines[i]

		// Check if this line contains "datatype:" followed by array items
		if strings.Contains(line, "datatype:") && !strings.Contains(line, "[") {
			// Find the indentation level
			indent := len(line) - len(strings.TrimLeft(line, " "))

			// Collect array items
			arrayItems := []string{}
			j := i + 1
			for j < len(lines) {
				nextLine := lines[j]
				nextIndent := len(nextLine) - len(strings.TrimLeft(nextLine, " "))

				// Check if it's an array item at the right indentation
				trimmed := strings.TrimSpace(nextLine)
				if nextIndent > indent && strings.HasPrefix(trimmed, "- ") {
					// Extract the value after "- "
					value := strings.TrimSpace(trimmed[2:])
					arrayItems = append(arrayItems, value)
					j++
				} else {
					break
				}
			}

			// If we found array items, convert to inline format
			if len(arrayItems) > 0 {
				compactLine := strings.Repeat(" ", indent) + "datatype: [" + strings.Join(arrayItems, ", ") + "]"
				result = append(result, compactLine)
				i = j // Skip the array items we processed
				continue
			}
		}

		// Check if this is a simple metadata field (field_name: \n type: string)
		// Pattern: line with field name ending with ":", next line is "type: <value>", no other nested fields
		trimmed := strings.TrimSpace(line)
		if strings.HasSuffix(trimmed, ":") && !strings.Contains(line, "metadata:") &&
		   !strings.Contains(line, "schema:") && !strings.Contains(line, "properties:") &&
		   i+1 < len(lines) {
			indent := len(line) - len(strings.TrimLeft(line, " "))
			nextLine := lines[i+1]
			nextIndent := len(nextLine) - len(strings.TrimLeft(nextLine, " "))
			nextTrimmed := strings.TrimSpace(nextLine)

			// Check if next line is "type: <value>" and is indented more than current
			if nextIndent > indent && strings.HasPrefix(nextTrimmed, "type: ") {
				typeValue := strings.TrimSpace(strings.TrimPrefix(nextTrimmed, "type:"))

				// Look ahead to see if there are more nested fields (like json_schema)
				hasMoreFields := false
				if i+2 < len(lines) {
					thirdLine := lines[i+2]
					thirdIndent := len(thirdLine) - len(strings.TrimLeft(thirdLine, " "))
					thirdTrimmed := strings.TrimSpace(thirdLine)
					// If there's another field at the same or greater indentation, keep expanded
					if thirdIndent >= nextIndent && thirdTrimmed != "" {
						hasMoreFields = true
					}
				}

				// Only compact if it's just a simple "type: value" with no other fields
				if !hasMoreFields {
					compactLine := strings.Repeat(" ", indent) + trimmed + " " + typeValue
					result = append(result, compactLine)
					i += 2 // Skip both lines
					continue
				}
			}
		}

		result = append(result, line)
		i++
	}

	return strings.Join(result, "\n")
}

// ExportAsJSON exports collection data as JSON
func ExportAsJSON(export *CollectionExport) (string, error) {
	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal to JSON: %w", err)
	}
	return string(data), nil
}

// WriteToFile writes content to a file
func WriteToFile(filePath string, content string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

// LoadSchemaFromYAMLFile loads a collection schema from a YAML file
func LoadSchemaFromYAMLFile(filePath string) (*CollectionExport, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var export CollectionExport
	err = yaml.Unmarshal(data, &export)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &export, nil
}

// LoadSchemaFromJSONFile loads a collection schema from a JSON file
func LoadSchemaFromJSONFile(filePath string) (*CollectionExport, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var export CollectionExport
	err = json.Unmarshal(data, &export)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &export, nil
}

// inferJSONStructure attempts to infer the structure of a JSON string
func inferJSONStructure(value interface{}) (map[string]interface{}, bool) {
	// Try to parse as JSON string
	var jsonStr string
	switch v := value.(type) {
	case string:
		jsonStr = v
	default:
		return nil, false
	}

	// Try to unmarshal as JSON object
	var jsonObj map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &jsonObj); err == nil {
		// Successfully parsed as JSON object
		schema := make(map[string]interface{})
		for key, val := range jsonObj {
			schema[key] = inferValueType(val)
		}
		return schema, true
	}

	return nil, false
}

// inferValueType infers the type of a value
func inferValueType(value interface{}) string {
	if value == nil {
		return "null"
	}

	switch v := value.(type) {
	case bool:
		return "boolean"
	case float64:
		// Check if it's an integer
		if v == float64(int64(v)) {
			return "integer"
		}
		return "number"
	case string:
		return "string"
	case []interface{}:
		if len(v) > 0 {
			return fmt.Sprintf("array[%s]", inferValueType(v[0]))
		}
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		return fmt.Sprintf("%T", value)
	}
}

// analyzeJSONFields analyzes metadata fields across multiple documents to find JSON fields
func analyzeJSONFields(documents []weaviate.Document, ctx context.Context, client *weaviate.Client, collectionName string) map[string]map[string]interface{} {
	jsonSchemas := make(map[string]map[string]interface{})
	jsonFieldSamples := make(map[string][]interface{})

	// Collect samples from multiple documents
	for _, doc := range documents {
		fullDoc, err := client.GetDocument(ctx, collectionName, doc.ID)
		if err != nil {
			continue
		}

		for key, value := range fullDoc.Metadata {
			if schema, isJSON := inferJSONStructure(value); isJSON {
				// Collect samples for this field
				jsonFieldSamples[key] = append(jsonFieldSamples[key], value)

				// Merge schemas (union of fields)
				if existing, ok := jsonSchemas[key]; ok {
					// Merge the schemas
					for field, fieldType := range schema {
						if _, exists := existing[field]; !exists {
							existing[field] = fieldType
						}
					}
				} else {
					jsonSchemas[key] = schema
				}
			}
		}
	}

	return jsonSchemas
}
