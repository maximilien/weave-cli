// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package pipeline

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/llm"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

type pipelineRoundTripper func(*http.Request) (*http.Response, error)

func (f pipelineRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newPipelineLLM(t *testing.T, status int) *llm.OpenAIClient {
	t.Helper()
	client, err := llm.NewOpenAIClientWithHTTP("test-key", &http.Client{Transport: pipelineRoundTripper(func(req *http.Request) (*http.Response, error) {
		body := `{"object":"list","data":[{"object":"embedding","embedding":[0.1,0.2,0.3],"index":0}],"model":"text-embedding-3-small","usage":{"prompt_tokens":1,"total_tokens":1}}`
		if status >= http.StatusBadRequest {
			body = `{"error":{"message":"embedding unavailable","type":"server_error"}}`
		}
		return &http.Response{
			StatusCode: status,
			Status:     http.StatusText(status),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    req,
		}, nil
	})})
	if err != nil {
		t.Fatalf("NewOpenAIClientWithHTTP(): %v", err)
	}
	return client
}

type recordingPipelineVDB struct {
	vectordb.VectorDBClient
	mu      sync.Mutex
	batches [][]*vectordb.Document
	err     error
}

func (v *recordingPipelineVDB) CreateDocuments(_ context.Context, _ string, documents []*vectordb.Document) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.batches = append(v.batches, append([]*vectordb.Document(nil), documents...))
	return v.err
}

func writePipelineFile(t *testing.T, dir, name, content string, fileType FileType) FileInfo {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile(%s): %v", name, err)
	}
	return FileInfo{Path: path, Size: int64(len(content)), Type: fileType, Hash: name + "-hash"}
}

func TestProcessorProcessesFilesAndBatchesDocuments(t *testing.T) {
	dir := t.TempDir()
	files := []FileInfo{
		writePipelineFile(t, dir, "one.txt", "plain text", FileTypeTXT),
		writePipelineFile(t, dir, "two.md", "# Markdown", FileTypeMD),
		writePipelineFile(t, dir, "three.json", `{"name":"three"}`, FileTypeJSON),
		writePipelineFile(t, dir, "four.yaml", "name: four", FileTypeYAML),
		{Path: filepath.Join(dir, "unknown.bin"), Type: FileTypeUnknown, Hash: "unknown-hash"},
	}
	vdb := &recordingPipelineVDB{}
	options := &IngestOptions{
		Collection: "docs", VDBType: "mock", BatchSize: 3, Workers: 2,
		Metadata: map[string]string{"campaign": "day-10"},
	}
	processor := NewProcessor(vdb, newPipelineLLM(t, http.StatusOK), options, NewProgressTracker(false, false))
	report, err := processor.ProcessFiles(context.Background(), files)
	if err != nil {
		t.Fatalf("ProcessFiles(): %v", err)
	}
	if report.Status != "partial" || report.FilesProcessed != 4 || report.FilesFailed != 1 || report.DocumentsCreated != 4 {
		t.Fatalf("report = %#v", report)
	}
	if report.FilesScanned != 5 || report.Collection != "docs" || report.VDBType != "mock" || report.ThroughputFiles <= 0 {
		t.Fatalf("report metadata = %#v", report)
	}
	vdb.mu.Lock()
	defer vdb.mu.Unlock()
	if len(vdb.batches) != 2 || len(vdb.batches[0]) != 3 || len(vdb.batches[1]) != 1 {
		t.Fatalf("batches = %#v", vdb.batches)
	}
	for _, batch := range vdb.batches {
		for _, doc := range batch {
			if len(doc.Embedding) != 3 || doc.Metadata["campaign"] != "day-10" || doc.Metadata["file_hash"] == "" {
				t.Errorf("document = %#v", doc)
			}
		}
	}
}

func TestProcessorResumeDryRunAndFailureStatus(t *testing.T) {
	dir := t.TempDir()
	stateFile := filepath.Join(dir, "state.json")
	skipped := writePipelineFile(t, dir, "skipped.txt", "skip", FileTypeTXT)
	processed := writePipelineFile(t, dir, "processed.txt", "process", FileTypeTXT)
	if err := saveState(stateFile, &IngestState{ProcessedFiles: map[string]bool{skipped.Hash: true}}); err != nil {
		t.Fatalf("saveState(): %v", err)
	}

	vdb := &recordingPipelineVDB{}
	processor := NewProcessor(vdb, newPipelineLLM(t, http.StatusOK), &IngestOptions{
		Collection: "docs", BatchSize: 10, Workers: 1, Resume: true, StateFile: stateFile, DryRun: true,
	}, NewProgressTracker(false, true))
	report, err := processor.ProcessFiles(context.Background(), []FileInfo{skipped, processed})
	if err != nil || report.Status != "success" || report.FilesSkipped != 1 || report.FilesProcessed != 1 {
		t.Fatalf("ProcessFiles(resume) = %#v, %v", report, err)
	}
	if len(vdb.batches) != 0 {
		t.Fatalf("dry run created batches: %#v", vdb.batches)
	}
	state, err := loadState(stateFile)
	if err != nil || !state.ProcessedFiles[processed.Hash] {
		t.Fatalf("loadState() = %#v, %v", state, err)
	}

	failure, err := processor.ProcessFiles(context.Background(), []FileInfo{{Path: "unsupported", Type: FileTypeImage, Hash: "image"}})
	if err != nil || failure.Status != "failure" || failure.FilesFailed != 1 || failure.FilesProcessed != 0 {
		t.Fatalf("ProcessFiles(failure) = %#v, %v", failure, err)
	}
}

func TestProcessorReportsEmbeddingAndBatchFailures(t *testing.T) {
	dir := t.TempDir()
	file := writePipelineFile(t, dir, "document.txt", "content", FileTypeTXT)
	progress := NewProgressTracker(false, false)

	embeddingFailure := NewProcessor(&recordingPipelineVDB{}, newPipelineLLM(t, http.StatusInternalServerError), &IngestOptions{
		Collection: "docs", BatchSize: 1, Workers: 1,
	}, progress)
	report, err := embeddingFailure.ProcessFiles(context.Background(), []FileInfo{file})
	if err != nil || report.Status != "failure" || !strings.Contains(report.Errors[0].Error, "embedding") {
		t.Fatalf("embedding failure report = %#v, %v", report, err)
	}

	batchVDB := &recordingPipelineVDB{err: errors.New("database offline")}
	batchFailure := NewProcessor(batchVDB, newPipelineLLM(t, http.StatusOK), &IngestOptions{
		Collection: "docs", BatchSize: 1, Workers: 1,
	}, progress)
	report, err = batchFailure.ProcessFiles(context.Background(), []FileInfo{file})
	if err != nil || report.Status != "partial" || !strings.Contains(report.Errors[0].Error, "database offline") {
		t.Fatalf("batch failure report = %#v, %v", report, err)
	}

	missing := FileInfo{Path: filepath.Join(dir, "missing.json"), Type: FileTypeJSON, Hash: "missing"}
	result := batchFailure.processFile(context.Background(), missing)
	if result.Error == nil || !strings.Contains(result.Error.Error(), "failed to read file") {
		t.Fatalf("processFile(missing) = %#v", result)
	}
}

func TestProcessorBatchDefaults(t *testing.T) {
	vdb := &recordingPipelineVDB{}
	processor := NewProcessor(vdb, nil, &IngestOptions{Collection: "docs", BatchSize: 0}, NewProgressTracker(false, false))
	documents := make([]*vectordb.Document, 101)
	for i := range documents {
		documents[i] = &vectordb.Document{ID: "document"}
	}
	if err := processor.createDocumentsInBatches(context.Background(), documents); err != nil {
		t.Fatalf("createDocumentsInBatches(): %v", err)
	}
	if len(vdb.batches) != 2 || len(vdb.batches[0]) != 100 || len(vdb.batches[1]) != 1 {
		t.Fatalf("default batches = %#v", vdb.batches)
	}
}
