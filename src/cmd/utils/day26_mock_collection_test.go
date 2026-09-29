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
}
