// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package doctor

import "encoding/json"

// Status represents the outcome of a single diagnostic check.
type Status int

const (
	StatusOK   Status = iota // Check passed
	StatusWarn               // Non-critical issue detected
	StatusFail               // Critical issue detected
	StatusSkip               // Check was skipped (precondition not met)
)

// String returns a human-readable label for the status.
func (s Status) String() string {
	switch s {
	case StatusOK:
		return "OK"
	case StatusWarn:
		return "WARN"
	case StatusFail:
		return "FAIL"
	case StatusSkip:
		return "SKIP"
	default:
		return "UNKNOWN"
	}
}

// MarshalJSON encodes the status as a JSON string (e.g. "OK", "FAIL").
func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// CheckResult holds the outcome of a single diagnostic check.
type CheckResult struct {
	Section string `json:"section"`           // e.g. "config", "vdb", "llm"
	Name    string `json:"name"`              // e.g. "Config file", "milvus-local connectivity"
	Status  Status `json:"status"`            // OK, WARN, FAIL, SKIP
	Message string `json:"message"`           // Human-readable detail
	Fix     string `json:"fix,omitempty"`     // Suggested remediation command or action
	Latency string `json:"latency,omitempty"` // Optional latency measurement
}

// DoctorConfig controls which sections to run and output behaviour.
type DoctorConfig struct {
	Fix     bool   // Attempt to auto-fix fixable issues
	Section string // Run only this section ("" = all)
	JSON    bool   // Machine-readable JSON output
	Verbose bool   // Show passing checks too
}

// SectionName constants used as keys across the codebase.
const (
	SectionSystem     = "system"
	SectionConfig     = "config"
	SectionEnv        = "env"
	SectionVDB        = "vdb"
	SectionLLM        = "llm"
	SectionEmbeddings = "embeddings"
	SectionStack      = "stack"
	SectionOpik       = "opik"
)
