// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package backup

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	backuppkg "github.com/maximilien/weave-cli/src/pkg/backup"
	"github.com/spf13/viper"
)

func setupMockBackupConfig(t *testing.T) string {
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
        - name: Existing
          type: text
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
	resetBackupCommandState(t)
	return root
}

func resetBackupCommandState(t *testing.T) {
	t.Helper()
	createOpts.Collection = ""
	createOpts.OutputFile = ""
	createOpts.VDBType = ""
	createOpts.BatchSize = 200
	createOpts.Compress = true
	createOpts.Quiet = false
	remoteStorage, s3Bucket, s3Endpoint, s3AccessKey, s3SecretKey, s3Prefix = "", "", "", "", "", ""
	s3Region, s3NoSSL, remoteOnly, remoteKeepLocal = "us-east-1", false, false, true
	restoreOpts.BackupFile = ""
	restoreOpts.Collection = ""
	restoreOpts.VDBType = ""
	restoreOpts.Overwrite = false
	restoreOpts.Quiet = false
	restoreRemoteStorage, restoreS3Bucket, restoreS3Endpoint = "", "", ""
	restoreS3AccessKey, restoreS3SecretKey, restoreS3Prefix = "", "", ""
	restoreS3Region, restoreS3NoSSL, restoreKeepLocal = "us-east-1", false, false
	t.Cleanup(func() {
		createOpts.OutputFile = ""
		createOpts.BatchSize = 200
		createOpts.Compress = true
		createOpts.Quiet = false
		remoteStorage = ""
		restoreOpts.BackupFile = ""
		restoreOpts.Collection = ""
		restoreOpts.Overwrite = false
		restoreOpts.Quiet = false
		restoreRemoteStorage = ""
	})
}

func TestRunBackupCreateWithConfiguredMockCollection(t *testing.T) {
	root := setupMockBackupConfig(t)
	createOpts.OutputFile = filepath.Join(root, "docs.weavebak")
	createOpts.BatchSize = 2

	output, err := captureBackupStdout(t, func() error {
		return runBackupCreate(CreateCmd, []string{"Docs"})
	})
	if err != nil {
		t.Fatal(err)
	}
	backupPath := createOpts.OutputFile
	if !strings.HasSuffix(backupPath, ".gz") {
		t.Fatalf("compressed backup path = %q", backupPath)
	}
	backup, err := backuppkg.ReadBackup(backupPath)
	if err != nil {
		t.Fatalf("read created backup: %v", err)
	}
	if backup.Metadata.Collection != "Docs" || backup.Metadata.VDBType != "mock" || len(backup.Documents) != 0 {
		t.Fatalf("created backup = %#v", backup.Metadata)
	}
	for _, want := range []string{"Creating backup", "Collection has 0 documents", "Backup created successfully", "Backup saved to"} {
		if !strings.Contains(output, want) {
			t.Errorf("create output missing %q:\n%s", want, output)
		}
	}

	createOpts.OutputFile = filepath.Join(root, "quiet.weavebak")
	createOpts.Compress = false
	createOpts.Quiet = true
	if err := runBackupCreate(CreateCmd, []string{"Docs"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(createOpts.OutputFile); err != nil {
		t.Fatalf("quiet backup file: %v", err)
	}

	createOpts.OutputFile = filepath.Join(root, "missing.weavebak")
	if err := runBackupCreate(CreateCmd, []string{"Missing"}); err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("missing collection error = %v", err)
	}
}

func TestRunBackupRestoreWithMockDatabase(t *testing.T) {
	root := setupMockBackupConfig(t)
	path := filepath.Join(root, "restore.weavebak")
	backup := backuppkg.NewBackupFormat("SourceDocs", "weaviate-local", "fixture-model", 3)
	for i := 0; i < 205; i++ {
		backup.Documents = append(backup.Documents, backuppkg.BackupDocument{
			ID:      "document-" + time.Unix(int64(i), 0).UTC().Format("150405"),
			Content: "fixture content", Text: "fixture text",
			Embedding: []float64{0.1, 0.2, 0.3},
			Metadata:  map[string]interface{}{"index": i},
			Image:     "image", ImageData: "data", URL: "https://fixture.invalid/image",
			ImageThumbnail: "thumb", ImageURL: "https://fixture.invalid/full",
			ImageMetadata: map[string]interface{}{"format": "png"},
		})
	}
	backup.Metadata.TotalDocuments = len(backup.Documents)
	if err := backuppkg.WriteBackup(backup, path, false); err != nil {
		t.Fatal(err)
	}

	restoreOpts.Collection = "Restored"
	output, err := captureBackupStdout(t, func() error {
		return runBackupRestore(RestoreCmd, []string{path})
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Backup file loaded", "205", "Cross-VDB migration", "Restore completed successfully", "Restored"} {
		if !strings.Contains(output, want) {
			t.Errorf("restore output missing %q:\n%s", want, output)
		}
	}

	restoreOpts.Collection = "Existing"
	restoreOpts.Overwrite = true
	restoreOpts.Quiet = true
	if err := runBackupRestore(RestoreCmd, []string{path}); err != nil {
		t.Fatalf("overwrite restore: %v", err)
	}

	restoreOpts.Collection = ""
	restoreOpts.Overwrite = false
	if err := runBackupRestore(RestoreCmd, []string{filepath.Join(root, "missing.weavebak")}); err == nil || !strings.Contains(err.Error(), "failed to read backup") {
		t.Fatalf("missing backup error = %v", err)
	}
}

func TestBackupRemoteStorageValidation(t *testing.T) {
	setupMockBackupConfig(t)
	t.Setenv("AWS_ACCESS_KEY_ID", "")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "")

	remoteStorage = "s3"
	if err := uploadToRemoteStorage(context.Background(), "missing.weavebak"); err == nil || !strings.Contains(err.Error(), "s3-bucket") {
		t.Fatalf("upload bucket error = %v", err)
	}
	s3Bucket = "fixtures"
	if err := uploadToRemoteStorage(context.Background(), "missing.weavebak"); err == nil || !strings.Contains(err.Error(), "credentials required") {
		t.Fatalf("upload credential error = %v", err)
	}
	s3AccessKey, s3SecretKey = "access", "secret"
	if err := uploadToRemoteStorage(context.Background(), "missing.weavebak"); err == nil || !strings.Contains(err.Error(), "failed to read backup file") {
		t.Fatalf("upload file error = %v", err)
	}
	remoteStorage, s3Endpoint = "minio", ""
	if err := uploadToRemoteStorage(context.Background(), "missing.weavebak"); err == nil || !strings.Contains(err.Error(), "endpoint is required") {
		t.Fatalf("upload endpoint error = %v", err)
	}

	restoreRemoteStorage = "s3"
	if _, err := downloadFromRemoteStorage(context.Background(), "missing.weavebak"); err == nil || !strings.Contains(err.Error(), "s3-bucket") {
		t.Fatalf("download bucket error = %v", err)
	}
	restoreS3Bucket = "fixtures"
	if _, err := downloadFromRemoteStorage(context.Background(), "missing.weavebak"); err == nil || !strings.Contains(err.Error(), "credentials required") {
		t.Fatalf("download credential error = %v", err)
	}
	restoreS3AccessKey, restoreS3SecretKey = "access", "secret"
	restoreRemoteStorage, restoreS3Endpoint = "minio", ""
	if _, err := downloadFromRemoteStorage(context.Background(), "missing.weavebak"); err == nil || !strings.Contains(err.Error(), "endpoint is required") {
		t.Fatalf("download endpoint error = %v", err)
	}
}

func TestBackupCommandContracts(t *testing.T) {
	if BackupCmd.Use != "backup" || len(BackupCmd.Commands()) != 4 {
		t.Fatalf("BackupCmd = %q with %d subcommands", BackupCmd.Use, len(BackupCmd.Commands()))
	}
	for name, command := range map[string]interface{ ValidateArgs([]string) error }{
		"create": CreateCmd, "list": ListCmd, "restore": RestoreCmd, "validate": ValidateCmd,
	} {
		t.Run(name, func(t *testing.T) { _ = command.ValidateArgs(nil) })
	}
}
