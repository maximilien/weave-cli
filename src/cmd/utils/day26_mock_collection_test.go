package utils

import (
	"context"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/pdf"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/weaviate"
	"github.com/spf13/viper"
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

func TestAgentQueryEmptyAndMissingConfigPaths(t *testing.T) {
	ctx := context.Background()
	cfg := mockOperationConfig()
	options := weaviate.QueryOptions{TopK: 2, Verbose: true}
	QueryMultipleCollectionsWithAgent(ctx, cfg, []string{"MissingCollection"}, "query", options, []string{"unused"}, "text", false)
	QueryMultipleCollectionsWithAgentCrossVDB(ctx,
		[]CollectionSpec{{Name: "MissingCollection", VDBKey: "missing"}},
		map[string]*config.VectorDBConfig{}, "query", options, []string{"unused"}, "text", false)
	QueryWeaviateCollectionWithAgent(ctx, &config.VectorDBConfig{Type: config.VectorDBTypeCloud}, "Docs", "query", options, []string{"unused"}, "text", false)
	QueryMockCollectionWithAgent(ctx, cfg, "MissingCollection", "query", options, []string{"unused"}, "text", false)
	options.UseBM25 = true
	QueryMultipleCollectionsWithAgent(ctx, cfg, []string{"MissingCollection"}, "query", options, []string{"unused"}, "text", true)
	QueryMultipleCollectionsWithAgentCrossVDB(ctx,
		[]CollectionSpec{{Name: "MissingCollection", VDBKey: "fixture"}},
		map[string]*config.VectorDBConfig{"MissingCollection": cfg}, "query", options, []string{"unused"}, "json", true)
	QueryMultipleCollectionsWithAgentCrossVDB(ctx,
		[]CollectionSpec{{Name: "InvalidCollection", VDBKey: "invalid"}},
		map[string]*config.VectorDBConfig{"InvalidCollection": &config.VectorDBConfig{Type: config.VectorDBType("invalid")}}, "query", options, []string{"unused"}, "text", false)
}

func TestConfigUtilityOverrideAndTipsPaths(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Cleanup(viper.Reset)
	viper.Reset()
	viper.Set("config", "missing-config.yaml")
	if _, err := LoadConfigWithOverrides(); err == nil {
		t.Fatal("expected missing config error")
	}
	viper.Set("no-tips", false)
	PrintConfigTips()
	if !HandleConfigError(context.Canceled, false) {
		t.Fatal("expected HandleConfigError to report an error")
	}
	viper.Set("no-tips", true)
	PrintConfigTips()
}

func TestPDFImageContentComposition(t *testing.T) {
	if got := buildCombinedImageContent(pdf.PDFImageData{}); got != "" {
		t.Fatalf("empty image data produced %q", got)
	}
	got := buildCombinedImageContent(pdf.PDFImageData{SectionHeading: "Heading", SurroundingText: "Context", OCRText: "OCR"})
	if got != "Heading\n\nContext\n\nOCR" {
		t.Fatalf("unexpected combined image content: %q", got)
	}
	if got := buildCombinedImageContent(pdf.PDFImageData{OCRText: "OCR only"}); got != "OCR only" {
		t.Fatalf("unexpected OCR-only content: %q", got)
	}
}
