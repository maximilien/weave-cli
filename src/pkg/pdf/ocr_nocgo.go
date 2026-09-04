//go:build !cgo

// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package pdf

// extractOCRText disables optional OCR when a binary is built without CGO.
func extractOCRText(_ string) string {
	return ""
}
