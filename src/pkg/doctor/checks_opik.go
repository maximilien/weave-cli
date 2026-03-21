// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package doctor

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"
)

// CheckOpik tests connectivity to the Opik workspace if OPIK_API_KEY is set.
func CheckOpik(ctx context.Context) []CheckResult {
	apiKey := os.Getenv("OPIK_API_KEY")
	if apiKey == "" {
		return []CheckResult{{
			Section: SectionOpik,
			Name:    "Opik API key",
			Status:  StatusSkip,
			Message: "OPIK_API_KEY not set",
		}}
	}

	var results []CheckResult
	results = append(results, CheckResult{
		Section: SectionOpik,
		Name:    "Opik API key",
		Status:  StatusOK,
		Message: "set",
	})

	workspace := os.Getenv("OPIK_WORKSPACE")
	if workspace == "" {
		results = append(results, CheckResult{
			Section: SectionOpik,
			Name:    "Opik workspace",
			Status:  StatusWarn,
			Message: "OPIK_WORKSPACE not set",
			Fix:     "export OPIK_WORKSPACE=<your-workspace>",
		})
	} else {
		results = append(results, CheckResult{
			Section: SectionOpik,
			Name:    "Opik workspace",
			Status:  StatusOK,
			Message: workspace,
		})
	}

	// Test API connectivity
	baseURL := os.Getenv("OPIK_BASE_URL")
	if baseURL == "" {
		baseURL = "https://www.comet.com/opik/api"
	}

	results = append(results, probeOpikAPI(ctx, baseURL, apiKey))

	return results
}

func probeOpikAPI(ctx context.Context, baseURL, apiKey string) CheckResult {
	checkCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(checkCtx, http.MethodGet, baseURL+"/v1/private/is-alive", nil)
	if err != nil {
		return CheckResult{
			Section: SectionOpik,
			Name:    "Opik API connectivity",
			Status:  StatusFail,
			Message: fmt.Sprintf("failed to create request: %v", err),
		}
	}
	req.Header.Set("Authorization", apiKey)

	start := time.Now()
	resp, err := http.DefaultClient.Do(req)
	latency := time.Since(start)

	if err != nil {
		return CheckResult{
			Section: SectionOpik,
			Name:    "Opik API connectivity",
			Status:  StatusFail,
			Message: fmt.Sprintf("connection failed: %v", err),
			Fix:     "Check OPIK_API_KEY and network connectivity",
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return CheckResult{
			Section: SectionOpik,
			Name:    "Opik API connectivity",
			Status:  StatusFail,
			Message: fmt.Sprintf("HTTP %d from %s", resp.StatusCode, baseURL),
			Fix:     "Verify OPIK_API_KEY is valid",
		}
	}

	return CheckResult{
		Section: SectionOpik,
		Name:    "Opik API connectivity",
		Status:  StatusOK,
		Message: fmt.Sprintf("reachable (%s)", latency.Round(time.Millisecond)),
		Latency: latency.Round(time.Millisecond).String(),
	}
}
