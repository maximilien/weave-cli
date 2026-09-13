// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package agents

import "testing"

func TestAgentConfigResolvesPerAgentOverrides(t *testing.T) {
	config := defaultConfig()
	config.LLM.DefaultModel = "default-model"
	config.LLM.DefaultTemperature = 0.9
	config.LLM.DefaultMaxTokens = 999
	config.SchemaAgent.Model = "schema-model"
	config.ChunkingAgent.Model = "chunking-model"
	config.QueryAgent.Model = "query-model"
	config.PlanningAgent.Model = "planning-model"

	models := map[string]string{
		"schema": "schema-model", "chunking": "chunking-model",
		"query": "query-model", "planning": "planning-model", "unknown": "default-model",
	}
	for agentType, want := range models {
		if got := config.GetModel(agentType); got != want {
			t.Errorf("GetModel(%q) = %q, want %q", agentType, got, want)
		}
	}
	config.SchemaAgent.Model = ""
	if got := config.GetModel("schema"); got != "default-model" {
		t.Errorf("GetModel(schema fallback) = %q", got)
	}

	temperatures := map[string]float64{
		"schema": config.SchemaAgent.Temperature, "chunking": config.ChunkingAgent.Temperature,
		"query": config.QueryAgent.Temperature, "planning": config.PlanningAgent.Temperature,
		"report": config.ReportAgent.Temperature, "eval": config.EvalAgent.Temperature,
		"unknown": config.LLM.DefaultTemperature,
	}
	for agentType, want := range temperatures {
		if got := config.GetTemperature(agentType); got != want {
			t.Errorf("GetTemperature(%q) = %f, want %f", agentType, got, want)
		}
	}

	tokenLimits := map[string]int{
		"schema": config.SchemaAgent.MaxTokens, "chunking": config.ChunkingAgent.MaxTokens,
		"query": config.QueryAgent.MaxTokens, "planning": config.PlanningAgent.MaxTokens,
		"report": config.ReportAgent.MaxTokens, "eval": config.EvalAgent.MaxTokens,
		"unknown": config.LLM.DefaultMaxTokens,
	}
	for agentType, want := range tokenLimits {
		if got := config.GetMaxTokens(agentType); got != want {
			t.Errorf("GetMaxTokens(%q) = %d, want %d", agentType, got, want)
		}
	}
}
