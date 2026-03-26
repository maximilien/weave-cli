// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package doctor

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// CheckLLM validates the OpenAI API key and tests a small embedding call.
func CheckLLM(ctx context.Context) []CheckResult {
	var results []CheckResult

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		results = append(results, CheckResult{
			Section: SectionLLM,
			Name:    "OpenAI API key",
			Status:  StatusSkip,
			Message: "OPENAI_API_KEY not set",
			Fix:     "export OPENAI_API_KEY=<your-key>",
		})
		return results
	}

	// Key format sanity check
	if len(apiKey) < 20 {
		results = append(results, CheckResult{
			Section: SectionLLM,
			Name:    "OpenAI API key",
			Status:  StatusWarn,
			Message: "key looks too short",
			Fix:     "Verify your OPENAI_API_KEY value",
		})
		return results
	}

	if !strings.HasPrefix(apiKey, "sk-") {
		results = append(results, CheckResult{
			Section: SectionLLM,
			Name:    "OpenAI API key format",
			Status:  StatusWarn,
			Message: fmt.Sprintf("key starts with %q — expected \"sk-\" prefix", apiKey[:5]),
			Fix:     "Check that OPENAI_API_KEY is correct; shell env overrides .env file",
		})
	}

	// Check if shell env might be shadowing .env file
	envFileKey := getEnvFileValue("OPENAI_API_KEY")
	if envFileKey != "" && envFileKey != apiKey {
		results = append(results, CheckResult{
			Section: SectionLLM,
			Name:    "OpenAI API key conflict",
			Status:  StatusWarn,
			Message: "shell env and .env file have different values — shell wins",
			Fix:     "Run: unset OPENAI_API_KEY   to use .env value, or update shell env",
		})
	}

	results = append(results, CheckResult{
		Section: SectionLLM,
		Name:    "OpenAI API key",
		Status:  StatusOK,
		Message: fmt.Sprintf("set (%s...)", apiKey[:8]),
	})

	// Test embedding generation
	results = append(results, testEmbedding(ctx, apiKey))

	return results
}

func testEmbedding(ctx context.Context, apiKey string) CheckResult {
	checkCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	client := openai.NewClient(option.WithAPIKey(apiKey))

	start := time.Now()
	_, err := client.Embeddings.New(checkCtx, openai.EmbeddingNewParams{
		Model: openai.EmbeddingModelTextEmbedding3Small,
		Input: openai.EmbeddingNewParamsInputUnion{
			OfString: openai.String("weave doctor test"),
		},
	})
	latency := time.Since(start)

	if err != nil {
		fix := "Verify OPENAI_API_KEY is valid and has embedding permissions"
		errMsg := err.Error()
		if strings.Contains(errMsg, "401") || strings.Contains(errMsg, "Unauthorized") || strings.Contains(errMsg, "Incorrect API key") {
			fix = "OPENAI_API_KEY is invalid. Shell env overrides .env — run: unset OPENAI_API_KEY   then retry"
		}
		return CheckResult{
			Section: SectionLLM,
			Name:    "Embedding API call",
			Status:  StatusFail,
			Message: fmt.Sprintf("failed: %v", err),
			Fix:     fix,
		}
	}

	return CheckResult{
		Section: SectionLLM,
		Name:    "Embedding API call",
		Status:  StatusOK,
		Message: fmt.Sprintf("text-embedding-3-small responded in %s", latency.Round(time.Millisecond)),
		Latency: latency.Round(time.Millisecond).String(),
	}
}

// getEnvFileValue reads a key from .env without affecting os env.
func getEnvFileValue(key string) string {
	for _, path := range []string{".env", "env"} {
		m, err := godotenv.Read(path)
		if err == nil {
			if v, ok := m[key]; ok {
				return v
			}
		}
	}
	return ""
}
