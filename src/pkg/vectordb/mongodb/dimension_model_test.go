// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max
package mongodb

import "testing"

func TestDimensionModelInference(t *testing.T) {
	for dims, want := range map[int]string{768: "sentence-transformers/all-mpnet-base-v2", 384: "sentence-transformers/all-MiniLM-L6-v2", 1536: "text-embedding-3-small", 3072: "text-embedding-3-large", 1024: "nomic-embed-text", 17: "text-embedding-3-small"} {
		if got := inferEmbeddingModelFromDimensions(dims); got != want {
			t.Errorf("dimensions %d: got %s want %s", dims, got, want)
		}
	}
}
