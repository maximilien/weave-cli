//go:build darwin && (amd64 || arm64)

// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package utils

import (
	// Import Chroma adapter to register its factory (macOS only)
	_ "github.com/maximilien/weave-cli/src/pkg/vectordb/chroma"
)
