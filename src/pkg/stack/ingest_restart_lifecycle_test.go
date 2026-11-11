// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package stack

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/pipeline"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/milvus"
)

func TestIngestionRestartLifecycle(t *testing.T) {
	for _, mode := range []string{"scheduled", "automatic"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			t.Chdir(root)
			t.Setenv("HOME", root)
			stack := &StackConfig{Version: "1.0", Name: "fixture", Runtime: RuntimeConfig{Kubernetes: KubernetesConfig{Provider: "kind"}}, Infrastructure: InfrastructureConfig{VectorDB: VectorDBConfig{Type: "milvus", Version: "2.3.0"}}}
			if err := SaveStackConfig(stack, filepath.Join(root, DefaultStackFile)); err != nil {
				t.Fatal(err)
			}
			bin := t.TempDir()
			calls := filepath.Join(root, "kubectl.log")
			t.Setenv("KUBECTL_CALLS", calls)
			script := "#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$KUBECTL_CALLS\"\ncase \"$1\" in port-forward) exec /bin/sleep 30 ;; esac\n"
			if err := os.WriteFile(filepath.Join(bin, "kubectl"), []byte(script), 0700); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", bin)
			oldCancelled := false
			forward := &PortForwardContext{Cancel: func() { oldCancelled = true }}
			original := forward
			t.Cleanup(func() { forward.Stop() })
			factory := &ingestFixtureFactory{Factory: milvus.NewFactory(), client: &ingestFixtureClient{}}
			vectordb.RegisterFactory(vectordb.VectorDBTypeMilvusLocal, factory)
			t.Cleanup(func() { vectordb.RegisterFactory(vectordb.VectorDBTypeMilvusLocal, milvus.NewFactory()) })
			cfg := IngestConfig{CollectionName: "Target", RestartEvery: 1, MaxRetries: 1, SkipFailed: true, EmbeddingModel: "nomic-embed-text", ClusterInfo: &ClusterInfo{Context: "fixture-context"}, PortForwardCtx: &forward, CheckpointEvery: 1, CheckpointFile: filepath.Join(root, "checkpoint.json")}
			options := &pipeline.IngestOptions{Workers: 1, BatchSize: 1}
			progress := pipeline.NewProgressTracker(false, true)
			processor := pipeline.NewProcessor(nil, nil, options, progress)
			files := []pipeline.FileInfo{{Path: "first.fixture", Type: pipeline.FileType("unsupported")}, {Path: "second.fixture", Type: pipeline.FileType("unsupported")}}
			var report *pipeline.IngestReport
			var err error
			checkpoint := InitCheckpoint("Target", []string{files[0].Path}, nil)
			if mode == "scheduled" {
				report, err = processFilesWithRestarts(context.Background(), processor, files, cfg, nil, nil, options, progress)
			} else {
				files = files[:1]
				report, err = processFilesWithAutoRestart(context.Background(), processor, files, cfg, nil, nil, options, progress, checkpoint)
			}
			if err != nil {
				t.Fatal(err)
			}
			if report.FilesScanned != len(files) || report.FilesFailed != len(files) || report.FilesProcessed != 0 {
				t.Fatalf("aggregate report=%+v", report)
			}
			if !oldCancelled || forward == original || forward.LocalPort != 19530 || forward.Cmd == nil {
				t.Fatalf("port-forward was not replaced: old cancelled=%t new=%+v", oldCancelled, forward)
			}
			if factory.config == nil || factory.config.Address != "localhost:19530" || factory.config.VectorDimensions != 768 {
				t.Fatalf("replacement database config=%+v", factory.config)
			}
			data, err := os.ReadFile(calls)
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range []string{"rollout restart deployment/fixture-weave-stack-milvus --context fixture-context", "port-forward svc/fixture-weave-stack-milvus 19530:19530 --context fixture-context"} {
				if !strings.Contains(string(data), want) {
					t.Fatalf("kubectl calls missing %q: %s", want, data)
				}
			}
			if mode == "automatic" {
				saved, err := LoadCheckpoint(cfg.CheckpointFile)
				if err != nil || saved == nil || len(saved.FailedFiles) != 1 || saved.FailedFiles[0].Attempts != 2 {
					t.Fatalf("retry checkpoint=%+v error=%v", saved, err)
				}
			}
		})
	}
}
