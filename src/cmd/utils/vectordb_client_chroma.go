//go:build !windows
// +build !windows

// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package utils

import (
	// Import Chroma adapter to register its factory (non-Windows only)
	_ "github.com/maximilien/weave-cli/src/pkg/vectordb/chroma"
)
