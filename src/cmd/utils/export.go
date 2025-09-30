// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

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
	Type        string      `json:"type" yaml:"type"`
	Occurrences int         `json:"occurrences" yaml:"occurrences"`
	Sample      interface{} `json:"sample,omitempty" yaml:"sample,omitempty"`
}

// ExportCollectionSchemaAndMetadata exports collection schema and metadata
func ExportCollectionSchemaAndMetadata(ctx context.Context, client *weaviate.Client, collectionName string, includeMetadata bool, expandMetadata bool) (*CollectionExport, error) {
	export := &CollectionExport{
		Name: collectionName,
	}

	// Get schema
	schema, err := client.GetFullCollectionSchema(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection schema: %w", err)
	}
	export.Schema = schema

	// Get metadata if requested
	if includeMetadata {
		documents, err := client.ListDocuments(ctx, collectionName, 100)
		if err != nil {
			return nil, fmt.Errorf("failed to get documents for metadata analysis: %w", err)
		}

		if len(documents) > 0 {
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
						metadataTypes[key] = fmt.Sprintf("%T", value)
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
				export.Metadata[name] = info
			}
		}
	}

	return export, nil
}

// ExportAsYAML exports collection data as YAML
func ExportAsYAML(export *CollectionExport) (string, error) {
	data, err := yaml.Marshal(export)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to YAML: %w", err)
	}
	return string(data), nil
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
