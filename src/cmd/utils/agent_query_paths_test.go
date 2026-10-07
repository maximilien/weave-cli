// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package utils

import (
	"context"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	mockvdb "github.com/maximilien/weave-cli/src/pkg/vectordb/mock"
)

func TestIsImageCollectionBySchemaPaths(t *testing.T) {
	ctx := context.Background()
	if !isImageCollectionBySchema(ctx, nil, "PhotoArchive", false) {
		t.Fatal("image keyword was not detected")
	}
	if isImageCollectionBySchema(ctx, nil, "Documents", false) {
		t.Fatal("plain collection was incorrectly detected as image")
	}

	client, err := mockvdb.NewAdapter(&vectordb.Config{
		Type: vectordb.VectorDBTypeMock, Enabled: true,
		Collections: []string{"Docs"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := client.CreateDocument(ctx, "Docs", &vectordb.Document{ID: "image-1", ImageData: "base64"}); err != nil {
		t.Fatal(err)
	}
	if !isImageCollectionBySchema(ctx, client, "Docs", false) {
		t.Fatal("image document was not detected")
	}
}
