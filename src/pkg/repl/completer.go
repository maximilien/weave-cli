// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package repl

import (
	"os"
	"strings"

	"github.com/chzyer/readline"
)

// Completer provides tab completion for REPL commands
type Completer struct {
	commands map[string][]string
}

// NewCompleter creates a new command completer
func NewCompleter() *Completer {
	return &Completer{
		commands: map[string][]string{
			// General commands
			"/help":     {},
			"/examples": {},
			"/history":  {},
			"/clear":    {},
			"/status":   {},
			"/exit":     {},
			"/quit":     {},
			"help":      {},
			"clear":     {},
			"exit":      {},
			"quit":      {},

			// MCP commands
			"/mcp": {"list", "call", "status", "help"},

			// Collection commands
			"/collection": {"list", "use", "create", "delete", "info", "help"},
			"/col":        {"list", "use", "create", "delete", "info", "help"},
			"/use":        {},

			// Search and stats
			"/search": {},
			"/stats":  {},
		},
	}
}

// Do implements readline.AutoCompleter interface
func (c *Completer) Do(line []rune, pos int) ([][]rune, int) {
	text := string(line[:pos])
	words := strings.Fields(text)

	// No words yet - suggest all commands
	if len(words) == 0 {
		return c.suggestCommands(""), len(line)
	}

	// First word - suggest commands
	if len(words) == 1 && !strings.HasSuffix(text, " ") {
		return c.suggestCommands(words[0]), len(words[0])
	}

	// Second word - suggest subcommands if available
	if len(words) >= 1 {
		cmd := words[0]
		if subcommands, ok := c.commands[cmd]; ok && len(subcommands) > 0 {
			prefix := ""
			if len(words) > 1 {
				prefix = words[1]
			}
			return c.suggestSubcommands(subcommands, prefix), len(prefix)
		}
	}

	return nil, 0
}

// suggestCommands returns command suggestions matching the prefix
func (c *Completer) suggestCommands(prefix string) [][]rune {
	var suggestions [][]rune

	for cmd := range c.commands {
		if strings.HasPrefix(cmd, prefix) {
			suggestions = append(suggestions, []rune(cmd[len(prefix):]))
		}
	}

	return suggestions
}

// suggestSubcommands returns subcommand suggestions matching the prefix
func (c *Completer) suggestSubcommands(subcommands []string, prefix string) [][]rune {
	var suggestions [][]rune

	for _, subcmd := range subcommands {
		if strings.HasPrefix(subcmd, prefix) {
			suggestions = append(suggestions, []rune(subcmd[len(prefix):]))
		}
	}

	return suggestions
}

// CreateReadlineConfig creates a readline config with completion
func CreateReadlineConfig() *readline.Config {
	return &readline.Config{
		Prompt:            "\033[36m>\033[0m ",
		HistoryFile:       expandPath("$HOME/.weave_history"),
		AutoComplete:      NewCompleter(),
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
		FuncFilterInputRune: func(r rune) (rune, bool) {
			// Allow all characters
			return r, true
		},
	}
}

// expandPath expands environment variables in path
func expandPath(path string) string {
	if strings.HasPrefix(path, "$HOME") {
		return strings.Replace(path, "$HOME", homeDir(), 1)
	}
	return path
}

func homeDir() string {
	// Get home directory from environment
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	// Fallback to current directory
	return "."
}
