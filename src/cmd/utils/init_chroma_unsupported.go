//go:build !(darwin && (amd64 || arm64))

// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package utils

import (
	// Import Chroma stub package to register stub factory
	_ "github.com/maximilien/weave-cli/src/pkg/vectordb/chroma"
)

// Chroma adapter only available on macOS AMD64 and macOS ARM64
// On other platforms, the stub factory is registered and returns platform errors
