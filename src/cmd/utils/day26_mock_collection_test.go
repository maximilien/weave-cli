package utils

import (
	"context"
	"testing"
)

func TestMockCollectionListingAndMaintenancePaths(t *testing.T) {
	cfg := mockOperationConfig()
	ctx := context.Background()
	ListMockCollections(ctx, cfg, 1, false, false)
	ListMockCollections(ctx, cfg, 0, false, true)
	if err := CreateMockCollection(ctx, cfg, "New", "", nil); err == nil {
		t.Fatal("expected unimplemented mock create error")
	}
	if _, err := CountMockCollections(ctx, cfg); err == nil {
		t.Fatal("expected unimplemented mock count error")
	}
	if err := DeleteMockCollections(ctx, cfg, []string{"AlphaDocs"}); err == nil {
		t.Fatal("expected unimplemented mock delete error")
	}
	if err := DeleteMockCollectionsByPattern(ctx, cfg, "Alpha*"); err == nil {
		t.Fatal("expected unimplemented mock pattern error")
	}
	ShowMockCollection(ctx, cfg, "AlphaDocs", 0, false, false, false, false, false, false, false, "", "", false)
	DeleteAllMockCollections(ctx, cfg)
	// Exercise the generic display paths against the deterministic mock backend.
	ShowGenericCollection(ctx, cfg, "AlphaDocs", 1, false, false, true, true, false, false, false, "", "", false)
	ShowGenericCollection(ctx, cfg, "AlphaDocs", 0, true, false, false, false, false, false, false, "", "", true)
	ShowGenericCollection(ctx, cfg, "MissingCollection", 1, false, false, false, false, false, false, false, "", "", false)
}

func TestImageCollectionDetectionPriorities(t *testing.T) {
	ctx := context.Background()
	cfg := mockOperationConfig()
	client, err := CreateVectorDBClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !isImageCollectionBySchema(ctx, client, "PhotoArchive", true) {
		t.Fatal("name keyword should identify image collection")
	}
	if isImageCollectionBySchema(ctx, client, "PlainDocs", true) {
		t.Fatal("plain mock collection should remain text")
	}
	if isImageCollectionBySchema(ctx, struct{}{}, "PlainDocs", true) {
		t.Fatal("unsupported client should remain text")
	}
}
