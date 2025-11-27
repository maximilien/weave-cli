//go:build !linux || !amd64
// +build !linux !amd64

// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

// Package chroma stub for unsupported platforms
// Chroma is only supported on Linux AMD64 due to libtokenizers CGO dependencies
package chroma

// This file exists to satisfy package imports on platforms where Chroma is not available.
// No actual functionality is provided.
