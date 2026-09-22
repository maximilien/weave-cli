// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package stack

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/maximilien/weave-cli/src/cmd/utils"
	backuppkg "github.com/maximilien/weave-cli/src/pkg/backup"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

func newStackBackupClient(t *testing.T, collections ...string) vectordb.VectorDBClient {
	t.Helper()
	configured := make([]config.Collection, len(collections))
	for i, name := range collections {
		configured[i] = config.Collection{Name: name, Type: "text"}
	}
	client, err := utils.CreateVectorDBClient(&config.VectorDBConfig{
		Name:               "stack-mock",
		Type:               config.VectorDBTypeMock,
		Enabled:            true,
		SimulateEmbeddings: true,
		EmbeddingDimension: 3,
		Collections:        configured,
	})
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func TestBackupStackCollectionLifecycle(t *testing.T) {
	ctx := context.Background()
	client := newStackBackupClient(t, "Docs")
	documents := []*vectordb.Document{
		{ID: "one", Content: "first", Text: "first", Embedding: []float64{1, 2, 3}, Metadata: map[string]interface{}{"source": "a"}},
		{ID: "two", Content: "second", Image: "image.png", ImageData: "base64", URL: "https://example.invalid/image", ImageThumbnail: "thumb", ImageURL: "s3://bucket/image", ImageMetadata: map[string]interface{}{"width": 10}},
	}
	if err := client.CreateDocuments(ctx, "Docs", documents); err != nil {
		t.Fatal(err)
	}

	previousBatch, previousCompress := backupBatchSize, backupCompress
	backupBatchSize, backupCompress = 1, false
	t.Cleanup(func() { backupBatchSize, backupCompress = previousBatch, previousCompress })

	output := filepath.Join(t.TempDir(), "docs.weavebak")
	if err := backupStackCollection(ctx, client, "Docs", output); err != nil {
		t.Fatal(err)
	}
	backup, err := backuppkg.ReadBackup(output)
	if err != nil {
		t.Fatal(err)
	}
	if backup.Metadata.TotalDocuments != 2 || len(backup.Documents) != 2 {
		t.Fatalf("backup = %#v", backup)
	}
	if backup.Documents[1].Image != "image.png" || backup.Documents[1].URL != "https://example.invalid/image" {
		t.Fatalf("image backup = %#v", backup.Documents[1])
	}

	if err := backupStackCollection(ctx, client, "Missing", filepath.Join(t.TempDir(), "missing.weavebak")); err == nil {
		t.Fatal("missing collection backup returned no error")
	}
	fileParent := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(fileParent, []byte("not a directory"), 0o600); err != nil {
		t.Fatal(err)
	}
	badOutput := filepath.Join(fileParent, "docs.weavebak")
	if err := backupStackCollection(ctx, client, "Docs", badOutput); err == nil {
		t.Fatal("backup to missing directory returned no error")
	}
}

func TestBackupAllStackCollections(t *testing.T) {
	ctx := context.Background()
	client := newStackBackupClient(t, "Docs", "Images")
	if err := client.CreateDocument(ctx, "Docs", &vectordb.Document{ID: "one", Content: "content"}); err != nil {
		t.Fatal(err)
	}

	previousBatch, previousCompress := backupBatchSize, backupCompress
	backupBatchSize, backupCompress = 10, true
	t.Cleanup(func() { backupBatchSize, backupCompress = previousBatch, previousCompress })

	outputDir := filepath.Join(t.TempDir(), "backups")
	if err := backupAllStackCollections(ctx, client, outputDir); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"Docs.weavebak.gz", "Images.weavebak.gz"} {
		if _, err := os.Stat(filepath.Join(outputDir, name)); err != nil {
			t.Errorf("backup %s: %v", name, err)
		}
	}

	empty := newStackBackupClient(t)
	if err := backupAllStackCollections(ctx, empty, filepath.Join(t.TempDir(), "empty")); err != nil {
		t.Fatal(err)
	}
	fileTarget := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(fileTarget, []byte("not a directory"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := backupAllStackCollections(ctx, client, fileTarget); err == nil {
		t.Fatal("backup-all accepted a file output directory")
	}
}
