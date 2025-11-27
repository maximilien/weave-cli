//go:build !(linux && amd64) && !(darwin && amd64) && !(darwin && arm64)
// +build !linux !amd64
// +build !darwin !amd64
// +build !darwin !arm64

// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package utils

import (
	// Import Chroma stub to register factory that returns platform errors
	_ "github.com/maximilien/weave-cli/src/pkg/vectordb/chroma"
)

// Chroma adapter only available on Linux AMD64, macOS AMD64, and macOS ARM64
