// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package stack

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	stackpkg "github.com/maximilien/weave-cli/src/pkg/stack"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/maximilien/weave-cli/src/pkg/vectordb/milvus"
)

type resetFactory struct {
	vectordb.ClientFactory
	client *resetClient
	err    error
	config *vectordb.Config
}

func (f *resetFactory) ValidateConfig(*vectordb.Config) error { return nil }
func (f *resetFactory) CreateClient(cfg *vectordb.Config) (vectordb.VectorDBClient, error) {
	f.config = cfg
	return f.client, f.err
}

type resetClient struct {
	vectordb.VectorDBClient
	collections []vectordb.CollectionInfo
	listError   error
	fail        string
	deleted     []string
}

func (c *resetClient) ListCollections(context.Context) ([]vectordb.CollectionInfo, error) {
	return c.collections, c.listError
}
func (c *resetClient) DeleteCollection(_ context.Context, name string) error {
	c.deleted = append(c.deleted, name)
	if name == c.fail {
		return errors.New("fixture delete failed")
	}
	return nil
}

func TestCollectionsResetLifecycle(t *testing.T) {
	for _, mode := range []string{"force keep", "confirmed", "cancelled", "all kept", "empty", "list failure", "partial failure", "client failure"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			t.Chdir(root)
			t.Setenv("HOME", root)
			if err := stackpkg.SaveStackConfig(generateQuickstartTemplate("kind"), ""); err != nil {
				t.Fatal(err)
			}
			if err := stackpkg.SaveClusterInfo(&stackpkg.ClusterInfo{Name: "fixture", Context: "fixture-context"}); err != nil {
				t.Fatal(err)
			}
			bin := t.TempDir()
			if err := os.WriteFile(filepath.Join(bin, "kubectl"), []byte("#!/bin/sh\nexec /bin/sleep 30\n"), 0700); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", bin)
			previousForce, previousKeep := collectionsForce, collectionsKeep
			collectionsForce = true
			collectionsKeep = []string{"keep"}
			t.Cleanup(func() { collectionsForce, collectionsKeep = previousForce, previousKeep })
			client := &resetClient{collections: []vectordb.CollectionInfo{{Name: "keep"}, {Name: "drop"}, {Name: "bad"}}}
			factory := &resetFactory{client: client}
			vectordb.RegisterFactory(vectordb.VectorDBTypeMilvusLocal, factory)
			t.Cleanup(func() { vectordb.RegisterFactory(vectordb.VectorDBTypeMilvusLocal, milvus.NewFactory()) })
			var want []string
			wantError := ""
			switch mode {
			case "force keep":
				want = []string{"drop", "bad"}
			case "confirmed", "cancelled":
				collectionsForce = false
				input := "no\n"
				if mode == "confirmed" {
					input = "yes\n"
					want = []string{"drop", "bad"}
				}
				path := filepath.Join(root, "input")
				if err := os.WriteFile(path, []byte(input), 0600); err != nil {
					t.Fatal(err)
				}
				reader, err := os.Open(path)
				if err != nil {
					t.Fatal(err)
				}
				previous := os.Stdin
				os.Stdin = reader
				t.Cleanup(func() { os.Stdin = previous; _ = reader.Close() })
			case "all kept":
				collectionsKeep = []string{"keep", "drop", "bad"}
			case "empty":
				client.collections = nil
			case "list failure":
				client.listError = errors.New("fixture list failed")
				wantError = "failed to list collections"
			case "partial failure":
				client.fail = "drop"
				want = []string{"drop", "bad"}
				wantError = "some collections failed to delete"
			case "client failure":
				factory.err = errors.New("fixture client failed")
				wantError = "failed to create Milvus client"
			}
			err := runCollectionsReset(collectionsResetCmd, nil)
			if wantError == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), wantError) {
				t.Fatalf("reset error=%v; want %q", err, wantError)
			}
			if factory.config == nil || factory.config.Type != vectordb.VectorDBTypeMilvusLocal || factory.config.Address != "localhost:19530" || factory.config.Timeout != 30 {
				t.Fatalf("Milvus factory config=%+v", factory.config)
			}
			if !reflect.DeepEqual(client.deleted, want) {
				t.Fatalf("deletions=%v want=%v", client.deleted, want)
			}
		})
	}
}
