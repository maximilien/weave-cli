// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package doctor

import (
	"fmt"
	"os"
	"strings"

	"github.com/maximilien/weave-cli/src/pkg/config"
)

// envVarCheck describes a single environment variable to audit.
type envVarCheck struct {
	name     string
	required bool   // true = FAIL if missing, false = WARN
	section  string // which VDB/service needs it (for context in messages)
}

// CheckEnv audits environment variables required by the current configuration.
func CheckEnv(cfg *config.Config) []CheckResult {
	var results []CheckResult

	// Always-relevant vars
	checks := []envVarCheck{
		{"OPENAI_API_KEY", false, "embeddings / LLM"},
	}

	// Derive additional checks from configured databases
	if cfg != nil {
		for _, db := range cfg.Databases.VectorDatabases {
			checks = append(checks, envVarsForDB(db)...)
		}
	}

	// Opik vars (optional – only warn)
	for _, v := range []string{"OPIK_API_KEY", "OPIK_WORKSPACE"} {
		checks = append(checks, envVarCheck{v, false, "Opik monitoring"})
	}

	// De-duplicate by var name
	seen := make(map[string]bool)
	for _, c := range checks {
		if seen[c.name] {
			continue
		}
		seen[c.name] = true

		val := os.Getenv(c.name)
		if val == "" {
			status := StatusWarn
			if c.required {
				status = StatusFail
			}
			results = append(results, CheckResult{
				Section: SectionEnv,
				Name:    c.name,
				Status:  status,
				Message: fmt.Sprintf("not set (needed for %s)", c.section),
				Fix:     fmt.Sprintf("export %s=<value>", c.name),
			})
			continue
		}

		// Check for placeholder values
		if isPlaceholder(val) {
			results = append(results, CheckResult{
				Section: SectionEnv,
				Name:    c.name,
				Status:  StatusWarn,
				Message: "looks like a placeholder value",
				Fix:     fmt.Sprintf("Set a real value: export %s=<value>", c.name),
			})
			continue
		}

		results = append(results, CheckResult{
			Section: SectionEnv,
			Name:    c.name,
			Status:  StatusOK,
			Message: "set",
		})
	}

	return results
}

func envVarsForDB(db config.VectorDBConfig) []envVarCheck {
	var checks []envVarCheck
	t := string(db.Type)

	switch {
	case strings.Contains(t, "weaviate"):
		checks = append(checks,
			envVarCheck{"WEAVIATE_URL", false, "Weaviate"},
			envVarCheck{"WEAVIATE_API_KEY", false, "Weaviate"},
		)
	case strings.Contains(t, "supabase"):
		checks = append(checks,
			envVarCheck{"SUPABASE_URL", false, "Supabase"},
			envVarCheck{"SUPABASE_KEY", false, "Supabase"},
		)
	case strings.Contains(t, "milvus"):
		checks = append(checks,
			envVarCheck{"MILVUS_ADDRESS", false, "Milvus"},
		)
	case strings.Contains(t, "mongodb"):
		checks = append(checks,
			envVarCheck{"MONGODB_URI", false, "MongoDB"},
		)
	case strings.Contains(t, "chroma"):
		checks = append(checks,
			envVarCheck{"CHROMA_URL", false, "Chroma"},
		)
	case strings.Contains(t, "qdrant"):
		checks = append(checks,
			envVarCheck{"QDRANT_URL", false, "Qdrant"},
		)
	case strings.Contains(t, "neo4j"):
		checks = append(checks,
			envVarCheck{"NEO4J_URI", false, "Neo4j"},
		)
	case strings.Contains(t, "opensearch"):
		checks = append(checks,
			envVarCheck{"OPENSEARCH_URL", false, "OpenSearch"},
		)
	case strings.Contains(t, "pinecone"):
		checks = append(checks,
			envVarCheck{"PINECONE_API_KEY", false, "Pinecone"},
		)
	}

	return checks
}

func isPlaceholder(val string) bool {
	lower := strings.ToLower(val)
	placeholders := []string{
		"your_", "changeme", "replace_me", "xxx", "todo",
		"<your", "placeholder", "example", "sk-xxx",
	}
	for _, p := range placeholders {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}
