// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package utils

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

type recordingPDFClient struct {
	vectordb.VectorDBClient
	docs                  map[string][]*vectordb.Document
	textError, imageError string
	failTextAt            int
}

func (c *recordingPDFClient) CreateDocument(_ context.Context, collection string, doc *vectordb.Document) error {
	if c.docs == nil {
		c.docs = make(map[string][]*vectordb.Document)
	}
	c.docs[collection] = append(c.docs[collection], doc)
	if collection == "Docs" && c.textError != "" && (c.failTextAt == 0 || len(c.docs[collection]) == c.failTextAt) {
		return errors.New(c.textError)
	}
	if collection == "Images" && c.imageError != "" {
		return errors.New(c.imageError)
	}
	return nil
}

func TestGenericPDFIngestionTextImageAndFailures(t *testing.T) {
	root := t.TempDir()
	t.Setenv("HOME", root)
	t.Setenv("XDG_CONFIG_HOME", root)
	bin := filepath.Join(root, "bin")
	if err := os.MkdirAll(bin, 0700); err != nil {
		t.Fatal(err)
	}
	script := "#!/bin/sh\nprintf 'First paragraph for fixture\\n\\nSecond paragraph for fixture\\n\\nThird paragraph for fixture\\n' > \"$3\"\n"
	if err := os.WriteFile(filepath.Join(bin, "pdftotext"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
	path := filepath.Join(root, "catalogue.pdf")
	if err := os.WriteFile(path, pdfWithJPEG(t), 0600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, textErr, imageErr, wantErr, wantMessage string
		failTextAt                                    int
		images                                        bool
		wantText, wantImages                          int
	}{
		{name: "text and image", images: true, wantText: 3, wantImages: 1, wantMessage: "Successfully created 1/1 image documents"},
		{name: "partial text failure", textErr: "temporary failure", failTextAt: 2, images: true, wantText: 3, wantImages: 1, wantMessage: "Failed to create PDF text document"},
		{name: "all text fail", textErr: "temporary failure", wantText: 3, wantErr: "failed to create any PDF document chunks"},
		{name: "missing text collection", textErr: "collection not found", wantText: 1, wantErr: "Collection 'Docs' not found"},
		{name: "missing image collection", imageErr: "collection not found", images: true, wantText: 3, wantImages: 1, wantErr: "image collection 'Images' not found"},
		{name: "image rejected", imageErr: "temporary failure", images: true, wantText: 3, wantImages: 1, wantMessage: "Failed to create any image documents"},
		{name: "image not requested", wantText: 3, wantMessage: "Images will not be stored"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := &recordingPDFClient{textError: tc.textErr, imageError: tc.imageErr, failTextAt: tc.failTextAt}
			imageCollection := ""
			if tc.images {
				imageCollection = "Images"
			}
			var runErr error
			out, diagnostics := captureCollectionListing(t, func() {
				runErr = processPDFFileGeneric(context.Background(), &config.VectorDBConfig{}, client, "Docs", path, 28, imageCollection, false, 0, 1000)
			})
			if tc.wantErr != "" {
				if runErr == nil || !strings.Contains(runErr.Error(), tc.wantErr) {
					t.Fatalf("error = %v; want %s", runErr, tc.wantErr)
				}
			} else if runErr != nil {
				t.Fatal(runErr)
			}
			if len(client.docs["Docs"]) != tc.wantText || len(client.docs["Images"]) != tc.wantImages {
				t.Fatalf("documents = %#v", client.docs)
			}
			if tc.wantMessage != "" && !strings.Contains(out+diagnostics, tc.wantMessage) {
				t.Fatalf("output = %q %q; want %s", out, diagnostics, tc.wantMessage)
			}
			for _, doc := range client.docs["Docs"] {
				if doc.ID == "" || doc.Content == "" || doc.Text != doc.Content || doc.Metadata["filename"] != "catalogue.pdf" {
					t.Errorf("text document = %#v", doc)
				}
			}
			for _, doc := range client.docs["Images"] {
				if doc.ID == "" || doc.ImageData == "" || !strings.Contains(doc.Content, "paragraph") || doc.Metadata["source_pdf"] != path {
					t.Errorf("image document = %#v", doc)
				}
			}
		})
	}
}
