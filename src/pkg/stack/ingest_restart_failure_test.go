// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package stack

import (
	"context"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/pipeline"
)

func TestIngestionRestartFailureStopsPreviousPortForward(t *testing.T) {
	for _, mode := range []string{"scheduled", "automatic"} {
		t.Run(mode, func(t *testing.T) {
			// A missing stack config makes restart fail before invoking kubectl.
			t.Chdir(t.TempDir())
			cancelled := false
			portForward := &PortForwardContext{Cancel: func() { cancelled = true }}
			cfg := IngestConfig{
				RestartEvery: 1, MaxRetries: 1,
				ClusterInfo: &ClusterInfo{Context: "fixture"}, PortForwardCtx: &portForward,
			}
			options := &pipeline.IngestOptions{Workers: 1, BatchSize: 1}
			progress := pipeline.NewProgressTracker(false, true)
			processor := pipeline.NewProcessor(nil, nil, options, progress)
			files := []pipeline.FileInfo{
				{Path: "first.fixture", Type: pipeline.FileType("unsupported")},
				{Path: "second.fixture", Type: pipeline.FileType("unsupported")},
			}
			var report *pipeline.IngestReport
			var err error
			if mode == "scheduled" {
				report, err = processFilesWithRestarts(context.Background(), processor, files, cfg, nil, nil, options, progress)
			} else {
				report, err = processFilesWithAutoRestart(context.Background(), processor, files, cfg, nil, nil, options, progress, nil)
			}
			if err == nil || !strings.Contains(err.Error(), "failed to restart Milvus: failed to load stack config:") {
				t.Fatalf("restart error = %v", err)
			}
			if report != nil {
				t.Fatalf("report = %+v, want nil after restart failure", report)
			}
			if !cancelled {
				t.Fatal("old port-forward was not cancelled before restart failed")
			}
		})
	}
}
