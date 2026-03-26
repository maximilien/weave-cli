// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

// PrintError prints an error message in red to stderr.
// Errors are diagnostic output and must not contaminate stdout JSON.
func PrintError(message string) {
	color.New(color.FgRed).Fprintln(os.Stderr, "❌ "+message)
}

// PrintSuccess prints a success message in green to stderr.
// Progress/status output goes to stderr so stdout stays clean for JSON.
func PrintSuccess(message string) {
	color.New(color.FgGreen).Fprintln(os.Stderr, "✅ "+message)
}

// PrintWarning prints a warning message in yellow to stderr.
func PrintWarning(message string) {
	color.New(color.FgYellow).Fprintln(os.Stderr, "⚠️  "+message)
}

// PrintInfo prints an info message in blue to stderr.
func PrintInfo(message string) {
	color.New(color.FgBlue).Fprintln(os.Stderr, "ℹ️  "+message)
}

// PrintHeader prints a header message in bold cyan to stderr.
func PrintHeader(message string) {
	color.New(color.FgCyan, color.Bold).Fprintln(os.Stderr, "📋 "+message)
}

// ConfirmAction prompts the user for confirmation
func ConfirmAction(message string) bool {
	fmt.Printf("%s (y/N): ", message)
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

// ConfirmActionStrict requires the user to type "yes" (not just "y")
func ConfirmActionStrict(message string) bool {
	fmt.Printf("%s (yes/N): ", message)
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "yes"
}

// PrintStyledKey prints a styled key
func PrintStyledKey(key string) {
	color.New(color.FgCyan).Print(key)
}

// PrintStyledValue prints a styled value
func PrintStyledValue(value string) {
	color.New(color.FgWhite).Print(value)
}

// PrintStyledValueDimmed prints a dimmed styled value
func PrintStyledValueDimmed(value string) {
	color.New(color.FgHiBlack).Print(value)
}

// PrintStyledID prints a styled ID
func PrintStyledID(id string) {
	color.New(color.FgGreen).Print(id)
}

// PrintStyledFilename prints a styled filename
func PrintStyledFilename(filename string) {
	color.New(color.FgWhite).Print(filename)
}

// PrintStyledNumber prints a styled number
func PrintStyledNumber(num int) {
	color.New(color.FgYellow).Print(num)
}

// PrintStyledEmoji prints a styled emoji
func PrintStyledEmoji(emoji string) {
	color.New(color.FgWhite, color.Bold).Print(emoji)
}

// PrintStyledKeyValueDimmed prints a dimmed key-value pair
func PrintStyledKeyValueDimmed(key, value string) {
	PrintStyledKey(key)
	fmt.Printf(": ")
	PrintStyledValueDimmed(value)
}

// PrintStyledKeyProminent prints a prominent styled key
func PrintStyledKeyProminent(key string) {
	color.New(color.FgWhite, color.Bold).Print(key)
}

// PrintStyledKeyValueProminentWithEmoji prints a prominent key-value pair with emoji
func PrintStyledKeyValueProminentWithEmoji(key, value, emoji string) {
	PrintStyledEmoji(emoji)
	fmt.Printf(" ")
	PrintStyledKeyProminent(key)
	if value != "" {
		fmt.Printf(": %s", value)
	}
}

// PrintStyledKeyNumberProminentWithEmoji prints a prominent key-number pair with emoji
func PrintStyledKeyNumberProminentWithEmoji(key string, num int, emoji string) {
	PrintStyledEmoji(emoji)
	fmt.Printf(" ")
	PrintStyledKeyProminent(key)
	fmt.Printf(": ")
	PrintStyledNumber(num)
}

// GetStyledKeyProminent returns a styled key string
func GetStyledKeyProminent(key string) string {
	return color.New(color.FgWhite, color.Bold).Sprint(key)
}

// GetStyledKeyDimmed returns a dimmed styled key string
func GetStyledKeyDimmed(key string) string {
	return color.New(color.FgHiBlack).Sprint(key)
}

// GetStyledNumber returns a styled number string
func GetStyledNumber(num interface{}) string {
	return color.New(color.FgYellow).Sprint(num)
}

// GetStyledValueDimmed returns a dimmed styled value string
func GetStyledValueDimmed(value string) string {
	return color.New(color.FgHiBlack).Sprint(value)
}

// GetStyledEmoji returns a styled emoji string
func GetStyledEmoji(emoji string) string {
	return color.New(color.FgWhite, color.Bold).Sprint(emoji)
}
