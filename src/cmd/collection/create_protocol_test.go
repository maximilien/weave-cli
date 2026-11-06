// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package collection

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/weaviate/weaviate/entities/models"
)

func TestCollectionCreateWeaviateSchemas(t *testing.T) {
	for _, tt := range []struct {
		name       string
		flags      map[string]string
		properties []string
		model      string
	}{
		{
			name:       "default embedding",
			properties: []string{"text", "url", "metadata"},
			model:      "text-embedding-3-small",
		},
		{
			name:       "deprecated embedding alias",
			flags:      map[string]string{"embedding-model": "text-embedding-3-large"},
			properties: []string{"text", "url", "metadata"},
			model:      "text-embedding-3-large",
		},
		{
			name:       "explicit embedding takes precedence",
			flags:      map[string]string{"embedding": "text-embedding-ada-002", "embedding-model": "text-embedding-3-large"},
			properties: []string{"text", "url", "metadata"},
			model:      "text-embedding-ada-002",
		},
		{
			name:       "OSS embedding model",
			flags:      map[string]string{"embedding": "nomic-embed-text"},
			properties: []string{"text", "url", "metadata"},
			model:      "nomic-embed-text",
		},
		{
			name:       "text schema without configured schemas",
			flags:      map[string]string{"text": "true", "json-metadata": "true"},
			properties: []string{"text", "url", "metadata"},
			model:      "text-embedding-3-small",
		},
		{
			name:       "image schema without configured schemas",
			flags:      map[string]string{"image": "true", "flat-metadata": "true"},
			properties: []string{"url", "image", "image_data", "metadata", "text"},
			model:      "text-embedding-3-small",
		},
		{
			name:       "named schema",
			flags:      map[string]string{"schema": "Article"},
			properties: []string{"title"},
		},
		{
			name:       "schema file overrides class name",
			flags:      map[string]string{"schema-yaml-file": "schema.yaml"},
			properties: []string{"body"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			setupMockCollectionConfig(t)
			// Preserve the command's flag values and Changed bits across cases.
			CreateCmd.Flags().VisitAll(func(flag *pflag.Flag) {
				value, changed := flag.Value.String(), flag.Changed
				t.Cleanup(func() {
					if err := flag.Value.Set(value); err != nil {
						t.Errorf("restore %s: %v", flag.Name, err)
					}
					flag.Changed = changed
				})
			})
			for name, value := range tt.flags {
				if err := CreateCmd.Flags().Set(name, value); err != nil {
					t.Fatal(err)
				}
			}

			created := make(chan models.Class, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.Method == http.MethodGet && r.URL.Path == "/v1/meta":
					_, _ = fmt.Fprint(w, `{"version":"1.27.0","modules":{}}`)
				case r.Method == http.MethodGet && r.URL.Path == "/v1/schema":
					_, _ = fmt.Fprint(w, `{"classes":[]}`)
				case r.Method == http.MethodPost && r.URL.Path == "/v1/schema":
					var schema models.Class
					if err := json.NewDecoder(r.Body).Decode(&schema); err != nil {
						t.Errorf("decode schema: %v", err)
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					select {
					case created <- schema:
					default:
						t.Error("collection created more than once")
					}
					w.WriteHeader(http.StatusCreated)
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					http.NotFound(w, r)
				}
			}))
			t.Cleanup(server.Close)
			t.Setenv("VECTOR_DB_TYPE", "")
			t.Setenv("WEAVIATE_URL", server.URL)
			config := fmt.Sprintf(`databases:
  default: weaviate-local
  vector_databases:
    - name: fixture
      type: weaviate-local
      enabled: true
      url: %q
  schemas:
    - name: Article
      schema:
        class: OriginalArticle
        vectorizer: none
        properties:
          - name: title
            datatype: [text]
`, server.URL)
			if err := os.WriteFile(viper.GetString("config"), []byte(config), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile("schema.yaml", []byte("name: OriginalFile\nschema:\n  class: OriginalFile\n  vectorizer: none\n  properties:\n    - name: body\n      datatype: [text]\n"), 0o600); err != nil {
				t.Fatal(err)
			}

			runCollectionCreate(CreateCmd, []string{"RequestedCollection"})

			var schema models.Class
			select {
			case schema = <-created:
			default:
				t.Fatal("command did not create a collection")
			}
			if schema.Class != "RequestedCollection" {
				t.Fatalf("created class = %q", schema.Class)
			}
			var properties []string
			for _, property := range schema.Properties {
				properties = append(properties, property.Name)
			}
			slices.Sort(properties)
			slices.Sort(tt.properties)
			if !slices.Equal(properties, tt.properties) {
				t.Fatalf("properties = %v, want %v", properties, tt.properties)
			}
			if tt.model == "" {
				if schema.Vectorizer != "none" {
					t.Fatalf("schema vectorizer = %q, want none", schema.Vectorizer)
				}
				return
			}
			vector, ok := schema.VectorConfig["default"]
			if !ok {
				t.Fatal("missing default vector configuration")
			}
			vectorizer, ok := vector.Vectorizer.(map[string]interface{})
			if !ok {
				t.Fatalf("vectorizer = %#v", vector.Vectorizer)
			}
			openai, ok := vectorizer["text2vec-openai"].(map[string]interface{})
			if !ok || openai["model"] != tt.model {
				t.Fatalf("embedding configuration = %#v, want model %q", vectorizer, tt.model)
			}
		})
	}
}
