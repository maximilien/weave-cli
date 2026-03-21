// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package doctor

import (
	"context"
	"fmt"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// CheckVDB runs Health() on each configured VDB and reports connectivity, latency, and collection count.
func CheckVDB(ctx context.Context, cfg *config.Config) []CheckResult {
	if cfg == nil || len(cfg.Databases.VectorDatabases) == 0 {
		return []CheckResult{{
			Section: SectionVDB,
			Name:    "VDB connectivity",
			Status:  StatusSkip,
			Message: "no vector databases configured",
		}}
	}

	var results []CheckResult
	for _, db := range cfg.Databases.VectorDatabases {
		if !db.Enabled {
			results = append(results, CheckResult{
				Section: SectionVDB,
				Name:    fmt.Sprintf("%s (%s)", db.Name, db.Type),
				Status:  StatusSkip,
				Message: "disabled in config",
			})
			continue
		}

		results = append(results, checkSingleVDB(ctx, db))
	}
	return results
}

func checkSingleVDB(ctx context.Context, db config.VectorDBConfig) CheckResult {
	label := fmt.Sprintf("%s (%s)", db.Name, db.Type)

	client, err := vectordb.CreateClientFromVectorDBConfig(&db)
	if err != nil {
		return CheckResult{
			Section: SectionVDB,
			Name:    label,
			Status:  StatusFail,
			Message: fmt.Sprintf("client creation failed: %v", err),
			Fix:     "Check database configuration in config.yaml",
		}
	}

	// Health check with timeout
	checkCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	start := time.Now()
	if err := client.Health(checkCtx); err != nil {
		return CheckResult{
			Section: SectionVDB,
			Name:    label,
			Status:  StatusFail,
			Message: fmt.Sprintf("health check failed: %v", err),
			Fix:     suggestVDBFix(db),
		}
	}
	latency := time.Since(start)

	// Try to count collections
	colCount := -1
	if collections, err := client.ListCollections(checkCtx); err == nil {
		colCount = len(collections)
	}

	msg := fmt.Sprintf("connected in %s", latency.Round(time.Millisecond))
	if colCount >= 0 {
		msg += fmt.Sprintf(", %d collections", colCount)
	}

	return CheckResult{
		Section: SectionVDB,
		Name:    label,
		Status:  StatusOK,
		Message: msg,
		Latency: latency.Round(time.Millisecond).String(),
	}
}

func suggestVDBFix(db config.VectorDBConfig) string {
	url := db.URL
	if url == "" {
		url = db.DatabaseURL
	}
	if url == "" {
		url = db.Address
	}

	if url != "" {
		return fmt.Sprintf("Check connectivity to %s or run: weave stack up", url)
	}
	return "Check database URL/address in config.yaml or run: weave stack up"
}
