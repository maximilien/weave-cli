// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package collection

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/spf13/viper"
)

func setupMockCollectionConfig(t *testing.T) {
	t.Helper()
	root := t.TempDir()
	t.Chdir(root)
	t.Setenv("HOME", root)
	t.Setenv("WEAVE_SKIP_CONFIG_VALIDATION", "true")
	t.Setenv("VECTOR_DB_TYPE", "mock")
	t.Setenv("WEAVIATE_URL", "https://fixture.invalid")
	t.Setenv("WEAVIATE_API_KEY", "fixture")
	t.Setenv("OPENAI_API_KEY", "fixture")
	path := filepath.Join(root, "config.yaml")
	contents := `databases:
  default: fixture
  vector_databases:
    - name: fixture
      type: mock
      enabled: true
      simulate_embeddings: true
      embedding_dimension: 3
      collections:
        - name: Docs
          type: text
        - name: Images
          type: image
`
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	viper.Reset()
	viper.Set("config", path)
	viper.Set("env", "")
	viper.Set("quiet", true)
	viper.Set("no-color", true)
	t.Cleanup(viper.Reset)
}

func TestCollectionReadCommandPathsWithMockDatabase(t *testing.T) {
	setupMockCollectionConfig(t)

	runCollectionList(ListCmd, nil)
	runCollectionList(ListCmd, []string{"fixture"})
	if err := ListCmd.Flags().Set("summary", "true"); err != nil {
		t.Fatal(err)
	}
	runCollectionList(ListCmd, nil)
	_ = ListCmd.Flags().Set("summary", "false")
	if err := ListCmd.Flags().Set("output", "json"); err != nil {
		t.Fatal(err)
	}
	runCollectionList(ListCmd, nil)
	_ = ListCmd.Flags().Set("output", "text")
	runCollectionShow(ShowCmd, []string{"Docs"})
	runCollectionQuery(QueryCmd, []string{"Docs", "search text"})
	runCollectionQuery(QueryCmd, []string{"Docs", "Images", "multi search"})
	runCollectionQuery(QueryCmd, []string{"Docs:fixture", "Images:fixture", "cross search"})

	imagePath := filepath.Join(t.TempDir(), "query.png")
	if err := os.WriteFile(imagePath, []byte("image fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	for name, value := range map[string]string{
		"search-type": "visual", "image": imagePath, "vector": "image_vector",
		"include-images": "true", "output": "json", "top-k": "2",
		"top-k-images": "1", "distance": "0.5", "search-metadata": "true",
		"bm25": "true", "verbose": "true",
	} {
		if err := QueryCmd.Flags().Set(name, value); err != nil {
			t.Fatalf("set query flag %s: %v", name, err)
		}
	}
	runCollectionQuery(QueryCmd, []string{"Docs", "visual search"})
	for name, value := range map[string]string{
		"search-type": "", "image": "", "vector": "", "include-images": "false",
		"output": "text", "top-k": "5", "top-k-images": "0", "distance": "0",
		"search-metadata": "false", "bm25": "false", "verbose": "false",
	} {
		_ = QueryCmd.Flags().Set(name, value)
	}
	t.Cleanup(func() {
		_ = ListCmd.Flags().Set("summary", "false")
		_ = ListCmd.Flags().Set("output", "text")
	})
	listCollectionsForDatabase(context.Background(), &config.VectorDBConfig{Type: config.VectorDBType("invalid")}, 10, false, false, false)

	displayCollectionsSummary(context.Background(), []config.VectorDBConfig{
		{Name: "fixture", Type: config.VectorDBTypeMock, Enabled: true},
		{Name: "invalid", Type: config.VectorDBType("invalid")},
	}, 10)
	if count, err := getCollectionCount(context.Background(), &config.VectorDBConfig{
		Name: "fixture", Type: config.VectorDBTypeMock, Enabled: true,
	}); err != nil || count != 0 {
		t.Fatalf("getCollectionCount(mock) = (%d, %v)", count, err)
	}
}

func TestCollectionMutationCommandPathsWithMockDatabase(t *testing.T) {
	setupMockCollectionConfig(t)

	withCollectionStdin(t, "\n", func() {
		runCollectionDeleteAll(DeleteAllCmd, nil)
	})
	withCollectionStdin(t, "\n", func() {
		runCollectionDelete(DeleteCmd, []string{"Docs"})
	})
	if err := DeleteCmd.Flags().Set("pattern", "Docs*"); err != nil {
		t.Fatal(err)
	}
	withCollectionStdin(t, "\n", func() {
		runCollectionDelete(DeleteCmd, nil)
	})
	_ = DeleteCmd.Flags().Set("pattern", "")

	if err := DeleteAllCmd.Flags().Set("force", "true"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = DeleteAllCmd.Flags().Set("force", "false") })
	runCollectionDeleteAll(DeleteAllCmd, nil)

	if err := DeleteSchemaCmd.Flags().Set("force", "true"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = DeleteSchemaCmd.Flags().Set("force", "false") })
	runCollectionDeleteSchema(DeleteSchemaCmd, []string{"Docs"})
	if err := DeleteSchemaCmd.Flags().Set("pattern", "Docs*"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = DeleteSchemaCmd.Flags().Set("pattern", "") })
	runCollectionDeleteSchema(DeleteSchemaCmd, nil)
}

func TestCollectionComparisonAndReembedPathsWithMockDatabase(t *testing.T) {
	setupMockCollectionConfig(t)
	report := filepath.Join(t.TempDir(), "comparison.md")
	for name, value := range map[string]string{
		"query":  "first query,second query",
		"top-k":  "2",
		"report": report,
		"format": "markdown",
	} {
		if err := CompareCmd.Flags().Set(name, value); err != nil {
			t.Fatalf("set compare flag %s: %v", name, err)
		}
	}
	runCompare(CompareCmd, []string{"Docs", "Images"})
	if data, err := os.ReadFile(report); err != nil || !strings.Contains(string(data), "Embedding Model Comparison") {
		t.Fatalf("comparison report = %q, %v", data, err)
	}
	if err := CompareCmd.Flags().Set("report", ""); err != nil {
		t.Fatal(err)
	}
	if err := CompareCmd.Flags().Set("format", "json"); err != nil {
		t.Fatal(err)
	}
	runCompare(CompareCmd, []string{"Docs", "Images"})

	for name, value := range map[string]string{
		"new-embedding": "text-embedding-3-small",
		"output":        "Docs",
		"batch-size":    "10",
		"skip-existing": "true",
	} {
		if err := ReEmbedCmd.Flags().Set(name, value); err != nil {
			t.Fatalf("set reembed flag %s: %v", name, err)
		}
	}
	runReEmbed(ReEmbedCmd, []string{"Docs"})
	t.Cleanup(func() {
		_ = CompareCmd.Flags().Set("query", "")
		_ = CompareCmd.Flags().Set("top-k", "5")
		_ = CompareCmd.Flags().Set("report", "")
		_ = CompareCmd.Flags().Set("format", "markdown")
		_ = ReEmbedCmd.Flags().Set("new-embedding", "")
		_ = ReEmbedCmd.Flags().Set("output", "")
		_ = ReEmbedCmd.Flags().Set("batch-size", "100")
		_ = ReEmbedCmd.Flags().Set("skip-existing", "false")
	})
}

func withCollectionStdin(t *testing.T, input string, run func()) {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writer.WriteString(input); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	previous := os.Stdin
	os.Stdin = reader
	t.Cleanup(func() { os.Stdin = previous })
	run()
	if err := reader.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdin = previous
}

func TestCollectionCommandContracts(t *testing.T) {
	if CollectionCmd.Use != "collection" || len(CollectionCmd.Commands()) != 8 {
		t.Fatalf("CollectionCmd = %q with %d subcommands", CollectionCmd.Use, len(CollectionCmd.Commands()))
	}
	for name, command := range map[string]interface {
		ValidateArgs([]string) error
	}{
		"count":         CountCmd,
		"create":        CreateCmd,
		"delete":        DeleteCmd,
		"delete all":    DeleteAllCmd,
		"delete schema": DeleteSchemaCmd,
		"show":          ShowCmd,
		"compare":       CompareCmd,
		"query":         QueryCmd,
		"reembed":       ReEmbedCmd,
	} {
		t.Run(name, func(t *testing.T) {
			_ = command.ValidateArgs(nil)
		})
	}
}
