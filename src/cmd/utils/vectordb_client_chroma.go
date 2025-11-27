//go:build (linux && amd64) || (darwin && amd64) || (darwin && arm64)
// +build linux,amd64 darwin,amd64 darwin,arm64

// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package utils

import (
	// Import Chroma adapter to register its factory (Linux AMD64 only)
	_ "github.com/maximilien/weave-cli/src/pkg/vectordb/chroma"
)
