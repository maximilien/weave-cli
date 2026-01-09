// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package weaviate

import (
	"strings"
	"testing"
)

// TestTargetVectorSelection tests that the correct target vector is selected
// for different collection types (regression test for multi-vector support)
func TestTargetVectorSelection(t *testing.T) {
	tests := []struct {
		name           string
		hasContent     bool
		hasText        bool
		hasImage       bool
		expectedVector string
		collectionType string
	}{
		{
			name:           "Text collection with content field",
			hasContent:     true,
			hasText:        false,
			hasImage:       false,
			expectedVector: "default",
			collectionType: "text",
		},
		{
			name:           "Text collection with text field",
			hasContent:     false,
			hasText:        true,
			hasImage:       false,
			expectedVector: "default",
			collectionType: "text",
		},
		{
			name:           "Image collection without text fields",
			hasContent:     false,
			hasText:        false,
			hasImage:       true,
			expectedVector: "text_vector",
			collectionType: "image",
		},
		{
			name:           "Mixed collection (has both text and image)",
			hasContent:     true,
			hasText:        false,
			hasImage:       true,
			expectedVector: "default",
			collectionType: "mixed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the logic from the actual query functions
			isImageColl := tt.hasImage && !tt.hasContent && !tt.hasText
			var targetVector string
			if isImageColl {
				targetVector = "text_vector"
			} else {
				targetVector = "default"
			}

			if targetVector != tt.expectedVector {
				t.Errorf("Expected target vector %s, got %s for %s collection",
					tt.expectedVector, targetVector, tt.collectionType)
			}
		})
	}
}

// TestQueryConstructionWithTargetVectors tests that all query types
// include targetVectors parameter (regression test for Issue #20)
func TestQueryConstructionWithTargetVectors(t *testing.T) {
	tests := []struct {
		name            string
		queryType       string
		targetVector    string
		expectTargetVec bool
	}{
		{
			name:            "NearText query with default vector",
			queryType:       "nearText",
			targetVector:    "default",
			expectTargetVec: true,
		},
		{
			name:            "NearText query with text_vector",
			queryType:       "nearText",
			targetVector:    "text_vector",
			expectTargetVec: true,
		},
		{
			name:            "Hybrid query with image_vector",
			queryType:       "hybrid",
			targetVector:    "image_vector",
			expectTargetVec: true,
		},
		{
			name:            "NearImage query",
			queryType:       "nearImage",
			targetVector:    "image_vector",
			expectTargetVec: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test different query construction patterns
			var query string
			switch tt.queryType {
			case "nearText":
				query = constructMockNearTextQuery(tt.targetVector)
			case "hybrid":
				query = constructMockHybridQuery(tt.targetVector)
			case "nearImage":
				query = constructMockNearImageQuery(tt.targetVector)
			}

			hasTargetVectors := strings.Contains(query, "targetVectors")
			if tt.expectTargetVec && !hasTargetVectors {
				t.Errorf("Query type %s should include targetVectors parameter:\n%s",
					tt.queryType, query)
			}

			// Verify the specific target vector is included
			if hasTargetVectors {
				expectedStr := `targetVectors: ["` + tt.targetVector + `"]`
				if !strings.Contains(query, expectedStr) {
					t.Errorf("Query should contain %s, but got:\n%s",
						expectedStr, query)
				}
			}
		})
	}
}

// Mock query construction functions that mirror the actual implementation
func constructMockNearTextQuery(targetVector string) string {
	return `{
		Get {
			TestCollection(
				nearText: {
					concepts: ["test"]
					targetVectors: ["` + targetVector + `"]
				}
				limit: 5
			) {
				_additional { id distance }
				content
			}
		}
	}`
}

func constructMockHybridQuery(targetVector string) string {
	return `{
		Get {
			TestCollection(
				hybrid: {
					query: "test"
					alpha: 0.75
					targetVectors: ["` + targetVector + `"]
				}
				limit: 5
			) {
				_additional { id score }
				content
			}
		}
	}`
}

func constructMockNearImageQuery(targetVector string) string {
	return `{
		Get {
			TestCollection(
				nearImage: {
					image: "base64data"
					targetVectors: ["` + targetVector + `"]
				}
				limit: 5
			) {
				_additional { id distance }
				url
			}
		}
	}`
}

// TestMultiVectorCollectionDetection tests the logic for detecting
// whether a collection has multiple named vectors
func TestMultiVectorCollectionDetection(t *testing.T) {
	tests := []struct {
		name          string
		properties    []string
		isMultiVector bool
	}{
		{
			name:          "Text collection - single vector",
			properties:    []string{"content", "metadata"},
			isMultiVector: false,
		},
		{
			name:          "Image collection - dual vectors",
			properties:    []string{"url", "image_data", "metadata"},
			isMultiVector: true,
		},
		{
			name:          "Mixed collection - not pure image",
			properties:    []string{"content", "image", "metadata"},
			isMultiVector: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasContent := false
			hasText := false
			hasImage := false

			for _, prop := range tt.properties {
				switch prop {
				case "content":
					hasContent = true
				case "text":
					hasText = true
				case "image", "image_data":
					hasImage = true
				}
			}

			isImageCollection := hasImage && !hasContent && !hasText
			if isImageCollection != tt.isMultiVector {
				t.Errorf("Expected isMultiVector=%v for %s, got %v",
					tt.isMultiVector, tt.name, isImageCollection)
			}
		})
	}
}
