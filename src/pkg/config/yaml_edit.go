// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// YAMLEditor provides utilities for editing YAML files while preserving structure
type YAMLEditor struct {
	filePath string
	root     *yaml.Node
}

// NewYAMLEditor creates a new YAML editor for the given file
func NewYAMLEditor(filePath string) (*YAMLEditor, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &YAMLEditor{
		filePath: filePath,
		root:     &root,
	}, nil
}

// SetValue sets a value at the given path (e.g., "databases.vector_databases[2].database_url")
func (e *YAMLEditor) SetValue(path string, value string) error {
	// Parse path into segments
	segments, err := parsePath(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Navigate to the target node
	node := e.root
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		node = node.Content[0]
	}

	for i, seg := range segments {
		isLast := i == len(segments)-1

		switch seg.Type {
		case segmentTypeKey:
			node, err = e.navigateToKey(node, seg.Value, isLast)
			if err != nil {
				return err
			}

		case segmentTypeIndex:
			idx := seg.Index
			node, err = e.navigateToIndex(node, idx, isLast)
			if err != nil {
				return err
			}
		}

		if isLast {
			// Set the value
			if node.Kind == yaml.ScalarNode {
				node.Value = value
			} else {
				return fmt.Errorf("cannot set value on non-scalar node")
			}
		}
	}

	return nil
}

// RemoveArrayElement removes an element from an array at the given path
func (e *YAMLEditor) RemoveArrayElement(path string, index int) error {
	// Parse path to get to the array
	segments, err := parsePath(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Navigate to the array node
	node := e.root
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		node = node.Content[0]
	}

	for _, seg := range segments {
		switch seg.Type {
		case segmentTypeKey:
			node, err = e.navigateToKey(node, seg.Value, false)
			if err != nil {
				return err
			}
		case segmentTypeIndex:
			node, err = e.navigateToIndex(node, seg.Index, false)
			if err != nil {
				return err
			}
		}
	}

	// node should now be a sequence
	if node.Kind != yaml.SequenceNode {
		return fmt.Errorf("path does not point to an array")
	}

	if index < 0 || index >= len(node.Content) {
		return fmt.Errorf("index %d out of bounds (array length: %d)", index, len(node.Content))
	}

	// Remove the element
	node.Content = append(node.Content[:index], node.Content[index+1:]...)

	return nil
}

// DisableArrayElement comments out an array element (marks it as disabled)
func (e *YAMLEditor) DisableArrayElement(path string, index int) error {
	// Parse path to get to the array
	segments, err := parsePath(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Navigate to the array node
	node := e.root
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		node = node.Content[0]
	}

	for _, seg := range segments {
		switch seg.Type {
		case segmentTypeKey:
			node, err = e.navigateToKey(node, seg.Value, false)
			if err != nil {
				return err
			}
		case segmentTypeIndex:
			node, err = e.navigateToIndex(node, seg.Index, false)
			if err != nil {
				return err
			}
		}
	}

	// node should now be a sequence
	if node.Kind != yaml.SequenceNode {
		return fmt.Errorf("path does not point to an array")
	}

	if index < 0 || index >= len(node.Content) {
		return fmt.Errorf("index %d out of bounds (array length: %d)", index, len(node.Content))
	}

	// Add "DISABLED: " comment to the element
	element := node.Content[index]
	if element.HeadComment != "" {
		element.HeadComment = "DISABLED: " + element.HeadComment
	} else {
		element.HeadComment = "DISABLED by weave config fix"
	}

	// Also add enabled: false field if it's a mapping
	if element.Kind == yaml.MappingNode {
		// Find or add "enabled" field
		foundEnabled := false
		for i := 0; i < len(element.Content); i += 2 {
			if element.Content[i].Value == "enabled" {
				element.Content[i+1].Value = "false"
				foundEnabled = true
				break
			}
		}
		if !foundEnabled {
			// Add enabled: false
			keyNode := &yaml.Node{
				Kind:  yaml.ScalarNode,
				Value: "enabled",
			}
			valueNode := &yaml.Node{
				Kind:  yaml.ScalarNode,
				Value: "false",
			}
			element.Content = append(element.Content, keyNode, valueNode)
		}
	}

	return nil
}

// Save writes the modified YAML back to the file
func (e *YAMLEditor) Save() error {
	data, err := yaml.Marshal(e.root)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	if err := os.WriteFile(e.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// SaveToFile saves the modified YAML to a different file
func (e *YAMLEditor) SaveToFile(filePath string) error {
	data, err := yaml.Marshal(e.root)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// navigateToKey navigates to a key in a mapping node
func (e *YAMLEditor) navigateToKey(node *yaml.Node, key string, createIfMissing bool) (*yaml.Node, error) {
	if node.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("expected mapping node, got %v", node.Kind)
	}

	// Search for the key
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1], nil
		}
	}

	if createIfMissing {
		// Create the key-value pair
		keyNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: key,
		}
		valueNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "",
		}
		node.Content = append(node.Content, keyNode, valueNode)
		return valueNode, nil
	}

	return nil, fmt.Errorf("key %q not found", key)
}

// navigateToIndex navigates to an index in a sequence node
func (e *YAMLEditor) navigateToIndex(node *yaml.Node, index int, allowMissing bool) (*yaml.Node, error) {
	if node.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("expected sequence node, got %v", node.Kind)
	}

	if index < 0 || index >= len(node.Content) {
		if allowMissing {
			return nil, fmt.Errorf("index %d out of bounds (array length: %d)", index, len(node.Content))
		}
		return nil, fmt.Errorf("index %d out of bounds (array length: %d)", index, len(node.Content))
	}

	return node.Content[index], nil
}

// Path segment types
type segmentType int

const (
	segmentTypeKey segmentType = iota
	segmentTypeIndex
)

type pathSegment struct {
	Type  segmentType
	Value string // For keys
	Index int    // For array indices
}

// parsePath parses a path string like "databases.vector_databases[2].database_url"
func parsePath(path string) ([]pathSegment, error) {
	var segments []pathSegment

	// Split by dots first
	parts := strings.Split(path, ".")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Check if part contains array index
		if strings.Contains(part, "[") {
			// Extract key and index
			bracketIdx := strings.Index(part, "[")
			key := part[:bracketIdx]
			indexPart := part[bracketIdx:]

			// Add key segment
			if key != "" {
				segments = append(segments, pathSegment{
					Type:  segmentTypeKey,
					Value: key,
				})
			}

			// Parse index: [2] or [2] (mongodb-cloud)
			var index int
			if _, err := fmt.Sscanf(indexPart, "[%d]", &index); err == nil {
				segments = append(segments, pathSegment{
					Type:  segmentTypeIndex,
					Index: index,
				})
			} else {
				return nil, fmt.Errorf("invalid array index in %q", part)
			}
		} else {
			// Simple key
			// Remove (database-name) suffix if present
			if parenIdx := strings.Index(part, "("); parenIdx != -1 {
				part = strings.TrimSpace(part[:parenIdx])
			}
			segments = append(segments, pathSegment{
				Type:  segmentTypeKey,
				Value: part,
			})
		}
	}

	return segments, nil
}
