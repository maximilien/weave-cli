// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package stack

import "testing"

func TestGetEmbeddingDimensions(t *testing.T) {
	tests := map[string]int{
		"text-embedding-3-small":                  1536,
		"text-embedding-3-large":                  3072,
		"text-embedding-ada-002":                  1536,
		"sentence-transformers/all-mpnet-base-v2": 768,
		"sentence-transformers/all-MiniLM-L6-v2":  384,
		"nomic-embed-text":                        768,
		"mxbai-embed-large":                       1024,
		"unknown-model":                           1536,
	}
	for model, want := range tests {
		t.Run(model, func(t *testing.T) {
			if got := getEmbeddingDimensions(model); got != want {
				t.Fatalf("getEmbeddingDimensions(%q) = %d, want %d", model, got, want)
			}
		})
	}
}
