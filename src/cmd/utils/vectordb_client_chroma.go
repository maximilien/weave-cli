//go:build linux && amd64
// +build linux,amd64

// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package utils

import (
	// Import Chroma adapter to register its factory (Linux AMD64 only)
	_ "github.com/maximilien/weave-cli/src/pkg/vectordb/chroma"
)
