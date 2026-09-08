// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package weaviate

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientSchemaWithFakeHTTPServer(t *testing.T) {
	var created map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/meta" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"hostname":"fake","version":"1.27.0","modules":{}}`)
			return
		}
		if r.URL.Path != "/v1/schema" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			if got := r.Header.Get("Authorization"); got != "Bearer api-key" {
				t.Errorf("Authorization = %q", got)
			}
			if err := json.NewDecoder(r.Body).Decode(&created); err != nil {
				t.Errorf("decode schema: %v", err)
			}
			w.WriteHeader(http.StatusCreated)
			return
		}
		_, _ = io.WriteString(w, `{
			"classes":[{
				"class":"Docs",
				"vectorizer":"none",
				"properties":[
					{"name":"content","dataType":["text"],"description":"body"},
					{"name":"metadata","dataType":["object"],"nestedProperties":[{"name":"author","dataType":["text"]}]}
				]
			}]
		}`)
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(&Config{URL: server.URL, APIKey: "api-key", Timeout: 1})
	if err != nil {
		t.Fatal(err)
	}
	properties, err := client.GetCollectionSchema(context.Background(), "Docs")
	if err != nil || strings.Join(properties, ",") != "content,metadata" {
		t.Fatalf("GetCollectionSchema() = %v, %v", properties, err)
	}
	missingProperties, err := client.GetCollectionSchema(context.Background(), "Missing")
	if err != nil || len(missingProperties) != 0 {
		t.Fatalf("GetCollectionSchema(missing) = %v, %v", missingProperties, err)
	}

	schema, err := client.GetFullCollectionSchema(context.Background(), "Docs")
	if err != nil {
		t.Fatalf("GetFullCollectionSchema() error = %v", err)
	}
	if schema.Class != "Docs" || schema.Vectorizer != "none" || schema.Properties[0].Description != "body" || schema.Properties[1].NestedProperties[0].Name != "author" {
		t.Fatalf("GetFullCollectionSchema() = %#v", schema)
	}
	if _, err := client.GetFullCollectionSchema(context.Background(), "Missing"); err == nil || !strings.Contains(err.Error(), "not found in schema") {
		t.Fatalf("GetFullCollectionSchema(missing) error = %v", err)
	}

	if err := client.CreateCollectionFromSchema(context.Background(), &CollectionSchema{Class: "Docs"}); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("CreateCollectionFromSchema(existing) error = %v", err)
	}
	newSchema := &CollectionSchema{
		Class: "NewDocs", Vectorizer: "none",
		Properties: []SchemaProperty{{
			Name: "metadata", DataType: []string{"object"}, Description: "metadata",
			NestedProperties: []SchemaProperty{{Name: "author", DataType: []string{"text"}, Description: "writer"}},
		}},
	}
	if err := client.CreateCollectionFromSchema(context.Background(), newSchema); err != nil {
		t.Fatalf("CreateCollectionFromSchema() error = %v", err)
	}
	if created["class"] != "NewDocs" || created["vectorizer"] != "none" {
		t.Fatalf("created schema = %#v", created)
	}
}

func TestClientSchemaHTTPErrors(t *testing.T) {
	t.Run("get schema", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = io.WriteString(w, "failed")
		}))
		t.Cleanup(server.Close)
		client, err := NewClient(&Config{URL: server.URL, Timeout: 1})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := client.GetCollectionSchema(context.Background(), "Docs"); err == nil || !strings.Contains(err.Error(), "failed to get schema") {
			t.Fatalf("GetCollectionSchema() error = %v", err)
		}
		if _, err := client.GetFullCollectionSchema(context.Background(), "Docs"); err == nil || !strings.Contains(err.Error(), "failed to get schema") {
			t.Fatalf("GetFullCollectionSchema() error = %v", err)
		}
	})

	t.Run("create schema", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if r.Method == http.MethodGet {
				_, _ = io.WriteString(w, `{"classes":[]}`)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, "invalid schema")
		}))
		t.Cleanup(server.Close)
		client, err := NewClient(&Config{URL: server.URL, Timeout: 1})
		if err != nil {
			t.Fatal(err)
		}
		err = client.CreateCollectionFromSchema(context.Background(), &CollectionSchema{Class: "Bad"})
		if err == nil || !strings.Contains(err.Error(), "status 400, body: invalid schema") {
			t.Fatalf("CreateCollectionFromSchema() error = %v", err)
		}
	})
}
