// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package doctor

import (
	"context"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/config"
)

// sectionOrder defines the order in which sections are executed.
var sectionOrder = []string{
	SectionSystem,
	SectionConfig,
	SectionEnv,
	SectionVDB,
	SectionLLM,
	SectionEmbeddings,
	SectionStack,
	SectionOpik,
}

// Doctor orchestrates all diagnostic checks.
type Doctor struct {
	cfg     *config.Config
	loadErr error
	dcfg    DoctorConfig
}

// New creates a Doctor instance.
// cfg may be nil when config loading itself failed; pass loadErr for that case.
func New(cfg *config.Config, loadErr error, dcfg DoctorConfig) *Doctor {
	return &Doctor{cfg: cfg, loadErr: loadErr, dcfg: dcfg}
}

// RunAll executes all enabled sections in order and returns the aggregated results.
func (d *Doctor) RunAll(ctx context.Context) []CheckResult {
	var results []CheckResult

	for _, section := range sectionOrder {
		if d.dcfg.Section != "" && d.dcfg.Section != section {
			continue
		}
		results = append(results, d.runSection(ctx, section)...)
	}

	return results
}

func (d *Doctor) runSection(ctx context.Context, section string) []CheckResult {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	switch section {
	case SectionSystem:
		return CheckSystem()
	case SectionConfig:
		return CheckConfig(d.cfg, d.loadErr)
	case SectionEnv:
		return CheckEnv(d.cfg)
	case SectionVDB:
		return CheckVDB(ctx, d.cfg)
	case SectionLLM:
		return CheckLLM(ctx)
	case SectionEmbeddings:
		return CheckEmbeddings(d.cfg)
	case SectionStack:
		return CheckStack()
	case SectionOpik:
		return CheckOpik(ctx)
	default:
		return nil
	}
}
