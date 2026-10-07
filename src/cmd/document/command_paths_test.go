// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package document

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func setupMockDocumentConfig(t *testing.T) string {
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
	return root
}

func TestDocumentReadCommandPathsWithMockDatabase(t *testing.T) {
	setupMockDocumentConfig(t)

	runDocumentList(ListCmd, []string{"Docs"})
	setDocumentFlags(t, ListCmd, map[string]string{
		"limit": "2", "offset": "1", "long": "true", "short": "2",
		"no-truncate": "true", "virtual": "true", "summary": "true", "output": "json",
	})
	runDocumentList(ListCmd, []string{"Docs"})
	resetDocumentFlags(ListCmd, map[string]string{
		"limit": "50", "offset": "0", "long": "false", "short": "5",
		"no-truncate": "false", "virtual": "false", "summary": "false", "output": "text",
	})

	runDocumentShow(ShowCmd, []string{"Docs"})
	setDocumentFlags(t, ShowCmd, map[string]string{
		"long": "true", "short": "2", "metadata": "filename=fixture.txt",
		"filename": "fixture.txt", "schema": "true", "expand-metadata": "true", "json": "true",
	})
	runDocumentShow(ShowCmd, []string{"Docs", "document-id"})
	resetDocumentFlags(ShowCmd, map[string]string{
		"long": "false", "short": "5", "metadata": "", "filename": "",
		"schema": "false", "expand-metadata": "false", "json": "false",
	})

	runDocumentCount(CountCmd, []string{"Docs", "Images"})
}

func TestBatchProcessingFailurePath(t *testing.T) {
	root := setupMockDocumentConfig(t)
	cfg, err := config.LoadConfig(filepath.Join(root, "config.yaml"), "")
	if err != nil {
		t.Fatal(err)
	}
	dbConfig := &cfg.Databases.VectorDatabases[0]
	dbConfig.Collections = []config.Collection{{Name: "Docs"}}
	progress := &BatchProgress{TotalFiles: 1, StartTime: time.Now()}
	status := processFileWithRetry(context.Background(), dbConfig, filepath.Join(root, "missing.txt"), "Docs", 0, 100, "", false, 0, 1, progress, &sync.Mutex{})
	if status.Success || status.Error == "" || progress.FailedFiles != 1 {
		t.Fatalf("unexpected batch failure status: %#v progress=%#v", status, progress)
	}
	validPath := filepath.Join(root, "valid.txt")
	if err := os.WriteFile(validPath, []byte("batch document content"), 0644); err != nil {
		t.Fatal(err)
	}
	validProgress := &BatchProgress{TotalFiles: 1, StartTime: time.Now()}
	valid := processFileWithRetry(context.Background(), dbConfig, validPath, "Docs", 0, 100, "", false, 0, 1, validProgress, &sync.Mutex{})
	if !valid.Success || valid.Error != "" || validProgress.SuccessFiles != 1 {
		t.Fatalf("unexpected batch success status: %#v progress=%#v", valid, validProgress)
	}
	parallelProgress := &BatchProgress{TotalFiles: 1, StartTime: time.Now()}
	results := processBatchFiles(context.Background(), dbConfig, []string{validPath}, "Docs", 2, 0, 100, "", false, 0, 1, parallelProgress)
	if len(results) != 1 || !results[0].Success {
		t.Fatalf("unexpected parallel batch results: %#v", results)
	}
}

func TestBatchCreateNoSupportedFiles(t *testing.T) {
	root := setupMockDocumentConfig(t)
	directory := filepath.Join(root, "batch")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "ignored.go"), []byte("package ignored"), 0o600); err != nil {
		t.Fatal(err)
	}
	setDocumentFlags(t, BatchCmd, map[string]string{
		"directory": directory, "collection": "Docs", "parallel": "2",
		"retry": "1", "chunk-size": "128", "since": "",
		"skip-existing": "true", "json": "false",
	})
	runBatchCreate(BatchCmd, nil)
	oldFile := filepath.Join(directory, "old.txt")
	if err := os.WriteFile(oldFile, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(oldFile, old, old); err != nil {
		t.Fatal(err)
	}
	setDocumentFlags(t, BatchCmd, map[string]string{"since": "1h"})
	runBatchCreate(BatchCmd, nil)
	status, err := json.Marshal(ProcessedFileStatus{FilePath: oldFile, Success: true, ProcessedAt: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(oldFile+".processed", status, 0o600); err != nil {
		t.Fatal(err)
	}
	setDocumentFlags(t, BatchCmd, map[string]string{"since": "", "skip-existing": "false"})
	runBatchCreate(BatchCmd, nil)
	resetDocumentFlags(BatchCmd, map[string]string{
		"directory": "./documents", "collection": "Documents", "parallel": "1",
		"retry": "2", "chunk-size": "5000", "since": "", "skip-existing": "false", "json": "false",
	})
}

func TestDocumentCreateCommandPathsWithMockDatabase(t *testing.T) {
	root := setupMockDocumentConfig(t)
	first := filepath.Join(root, "first.txt")
	second := filepath.Join(root, "second.md")
	image := filepath.Join(root, "image.png")
	for path, contents := range map[string]string{
		first: "first document content", second: "second document content", image: "image fixture",
	} {
		if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
			t.Fatalf("write fixture %s: %v", path, err)
		}
	}

	setDocumentFlags(t, CreateCmd, map[string]string{
		"workers": "0", "image-col": "Images", "create-report": filepath.Join(root, "report.csv"),
	})
	runDocumentCreate(CreateCmd, []string{"Docs", first})

	setDocumentFlags(t, CreateCmd, map[string]string{
		"workers": "20", "image-col": "", "image-cols": "Images", "create-report": "",
		"append-report": filepath.Join(root, "report.csv"), "image-storage": "s3",
		"s3-bucket": "fixtures", "store-pdf": "true", "timeout": "1m",
	})
	runDocumentCreate(CreateCmd, []string{"Docs", second})

	setDocumentFlags(t, CreateCmd, map[string]string{
		"workers": "2", "image-cols": "", "append-report": "", "image-storage": "minio",
		"minio-bucket": "fixtures", "timeout": "", "skip-all-images": "true",
	})
	runDocumentCreate(CreateCmd, []string{"Images", image})

	setDocumentFlags(t, CreateCmd, map[string]string{
		"image-storage": "local", "local-storage-path": filepath.Join(root, "storage"),
		"skip-all-images": "false", "image-collection": "Images",
	})
	runGlobDocumentCreate(CreateCmd, []string{"Docs", "*.txt"}, []string{first, second}, "Docs", 5000, "", false, true, 5120, 10, 2000, "", "", "", 2, time.Second, false)

	resetDocumentFlags(CreateCmd, map[string]string{
		"workers": "1", "image-col": "", "image-cols": "", "image-collection": "",
		"create-report": "", "append-report": "", "image-storage": "", "s3-bucket": "",
		"minio-bucket": "", "store-pdf": "false", "timeout": "", "skip-all-images": "false",
		"local-storage-path": "./storage",
	})
}

func TestDocumentDeleteCommandPathsWithMockDatabase(t *testing.T) {
	setupMockDocumentConfig(t)

	withDocumentStdin(t, "\n", func() { runDocumentDeleteAll(DeleteAllCmd, []string{"Docs"}) })
	for _, test := range []struct {
		args  []string
		flags map[string]string
	}{
		{args: []string{"Docs", "document-id"}},
		{args: []string{"Docs"}, flags: map[string]string{"pattern": "*.txt"}},
		{args: []string{"Docs"}, flags: map[string]string{"name": "fixture.txt"}},
		{args: []string{"Docs"}, flags: map[string]string{"metadata": "kind=fixture"}},
		{args: []string{"Docs"}},
	} {
		setDocumentFlags(t, DeleteCmd, test.flags)
		withDocumentStdin(t, "\n", func() { runDocumentDelete(DeleteCmd, test.args) })
		resetDocumentFlags(DeleteCmd, map[string]string{"pattern": "", "name": "", "metadata": ""})
	}

	setDocumentFlags(t, DeleteAllCmd, map[string]string{"force": "true"})
	runDocumentDeleteAll(DeleteAllCmd, []string{"Docs"})
	resetDocumentFlags(DeleteAllCmd, map[string]string{"force": "false"})

	setDocumentFlags(t, DeleteCmd, map[string]string{"force": "true"})
	for _, test := range []struct {
		args  []string
		flags map[string]string
	}{
		{args: []string{"Docs", "document-id"}},
		{args: []string{"Docs"}, flags: map[string]string{"metadata": "kind=fixture"}},
		{args: []string{"Docs"}, flags: map[string]string{"filename": "fixture.txt"}},
		{args: []string{"Docs"}, flags: map[string]string{"pattern": "*.txt"}},
		{args: []string{"Docs"}},
	} {
		setDocumentFlags(t, DeleteCmd, test.flags)
		runDocumentDelete(DeleteCmd, test.args)
		resetDocumentFlags(DeleteCmd, map[string]string{"pattern": "", "filename": "", "metadata": ""})
	}
	resetDocumentFlags(DeleteCmd, map[string]string{"force": "false"})
}

func TestDocumentCommandContracts(t *testing.T) {
	if DocumentCmd.Use != "document" || len(DocumentCmd.Commands()) != 8 {
		t.Fatalf("DocumentCmd = %q with %d subcommands", DocumentCmd.Use, len(DocumentCmd.Commands()))
	}
	for name, command := range map[string]interface{ ValidateArgs([]string) error }{
		"count": CountCmd, "create": CreateCmd, "delete": DeleteCmd,
		"delete all": DeleteAllCmd, "inspect": InspectCmd, "list": ListCmd,
		"pdf convert": PdfConvertCmd, "show": ShowCmd,
	} {
		t.Run(name, func(t *testing.T) { _ = command.ValidateArgs(nil) })
	}
	if got := formatCreateError("fixture.txt", "1s", os.ErrPermission); got == "" {
		t.Fatal("formatCreateError returned an empty message")
	}
	if got := formatCreateError("fixture.txt", "1s", context.DeadlineExceeded); got != "Timed out after 1s: fixture.txt" {
		t.Fatalf("formatCreateError deadline = %q", got)
	}
}

func setDocumentFlags(t *testing.T, command interface{ Flags() *pflag.FlagSet }, values map[string]string) {
	t.Helper()
	for name, value := range values {
		if flag := command.Flags().Lookup(name); flag != nil {
			if slice, ok := flag.Value.(pflag.SliceValue); ok {
				items := []string{}
				if value != "" {
					items = []string{value}
				}
				if err := slice.Replace(items); err != nil {
					t.Fatalf("set slice flag %s: %v", name, err)
				}
				continue
			}
		}
		if err := command.Flags().Set(name, value); err != nil {
			t.Fatalf("set flag %s: %v", name, err)
		}
	}
}

func resetDocumentFlags(command interface{ Flags() *pflag.FlagSet }, values map[string]string) {
	for name, value := range values {
		if flag := command.Flags().Lookup(name); flag != nil {
			if slice, ok := flag.Value.(pflag.SliceValue); ok {
				items := []string{}
				if value != "" {
					items = []string{value}
				}
				_ = slice.Replace(items)
				continue
			}
		}
		_ = command.Flags().Set(name, value)
	}
}

func withDocumentStdin(t *testing.T, input string, run func()) {
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
	run()
	os.Stdin = previous
	if err := reader.Close(); err != nil {
		t.Fatal(err)
	}
}
