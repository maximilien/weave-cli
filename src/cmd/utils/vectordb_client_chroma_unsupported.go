//go:build !(darwin && (amd64 || arm64))

// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package utils

import (
	// Import Chroma stub to register factory that returns platform errors
	_ "github.com/maximilien/weave-cli/src/pkg/vectordb/chroma"
)

// Chroma adapter only available on macOS AMD64 and macOS ARM64 (not Linux due to CGO)
