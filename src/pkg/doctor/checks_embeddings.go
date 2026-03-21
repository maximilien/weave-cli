// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package doctor

import (
	"fmt"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/embeddings"
	"github.com/maximilien/weave-cli/src/pkg/stack"
)

// CheckEmbeddings validates that configured embedding models are in the registry
// and that their dimensions match the VDB configuration.
func CheckEmbeddings(cfg *config.Config) []CheckResult {
	var results []CheckResult

	// Check stack config for embedding model info
	stackCfg, _ := stack.LoadStackConfig("")
	if stackCfg == nil {
		// No stack config – check if any VDB has vector_dimensions set
		if cfg != nil {
			for _, db := range cfg.Databases.VectorDatabases {
				if db.VectorDimensions > 0 {
					results = append(results, CheckResult{
						Section: SectionEmbeddings,
						Name:    fmt.Sprintf("%s dimensions", db.Name),
						Status:  StatusOK,
						Message: fmt.Sprintf("%d dimensions configured", db.VectorDimensions),
					})
				}
			}
		}

		if len(results) == 0 {
			results = append(results, CheckResult{
				Section: SectionEmbeddings,
				Name:    "Embedding config",
				Status:  StatusSkip,
				Message: "no stack config or explicit vector dimensions found",
			})
		}
		return results
	}

	// Walk collections in stack config looking for embedding model references
	for _, col := range stackCfg.Collections {
		if col.Embedding == nil || col.Embedding.Model == "" {
			continue
		}

		model := col.Embedding.Model
		info, err := embeddings.GetModelInfo(model)
		if err != nil {
			results = append(results, CheckResult{
				Section: SectionEmbeddings,
				Name:    fmt.Sprintf("Model: %s", model),
				Status:  StatusWarn,
				Message: "not found in embedding model registry",
				Fix:     "Run: weave embeddings list   to see known models",
			})
			continue
		}

		results = append(results, CheckResult{
			Section: SectionEmbeddings,
			Name:    fmt.Sprintf("Model: %s", model),
			Status:  StatusOK,
			Message: fmt.Sprintf("%s, %d dimensions, provider: %s", info.Name, info.Dimensions, info.Provider),
		})

		// Check dimension match with collection schema
		if col.Schema.VectorDimensions > 0 && col.Schema.VectorDimensions != info.Dimensions {
			results = append(results, CheckResult{
				Section: SectionEmbeddings,
				Name:    fmt.Sprintf("Dimension mismatch: %s", col.Name),
				Status:  StatusFail,
				Message: fmt.Sprintf("collection expects %d dims but model %s produces %d", col.Schema.VectorDimensions, model, info.Dimensions),
				Fix:     fmt.Sprintf("Update schema.vector_dimensions to %d or change embedding model", info.Dimensions),
			})
		}
	}

	if len(results) == 0 {
		results = append(results, CheckResult{
			Section: SectionEmbeddings,
			Name:    "Embedding config",
			Status:  StatusSkip,
			Message: "no embedding models configured in stack",
		})
	}

	return results
}
