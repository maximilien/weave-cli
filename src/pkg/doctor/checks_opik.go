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

	// Test API connectivity using the OTEL endpoint (same as the real tracing path)
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "https://www.comet.com/opik/api/v1/private/otel/v1/traces"
	}

	results = append(results, probeOpikAPI(ctx, endpoint, apiKey, workspace))

	return results
}

func probeOpikAPI(ctx context.Context, endpoint, apiKey, workspace string) CheckResult {
	checkCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// POST to the OTEL traces endpoint with an empty body.
	// A 200 or 415 (unsupported media type) means the server is reachable
	// and authenticated. Only network errors or 401/403 indicate a problem.
	req, err := http.NewRequestWithContext(checkCtx, http.MethodPost, endpoint, nil)
	if err != nil {
		return CheckResult{
			Section: SectionOpik,
			Name:    "Opik API connectivity",
			Status:  StatusFail,
			Message: fmt.Sprintf("failed to create request: %v", err),
		}
	}
	req.Header.Set("Authorization", apiKey)
	if workspace != "" {
		req.Header.Set("Comet-Workspace", workspace)
	}

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

	// 401/403 = bad auth, other 4xx/5xx we treat as reachable
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return CheckResult{
			Section: SectionOpik,
			Name:    "Opik API connectivity",
			Status:  StatusFail,
			Message: fmt.Sprintf("HTTP %d — authentication failed", resp.StatusCode),
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
