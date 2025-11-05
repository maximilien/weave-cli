// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package repl

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/maximilien/weave-cli/src/pkg/executor"
	"github.com/spf13/viper"
)

// REPL represents the interactive Read-Eval-Print Loop
type REPL struct {
	executor    *executor.Executor
	rl          *readline.Instance
	interrupted bool
}

// New creates a new REPL instance
func New() (*REPL, error) {
	// Load .env file
	_ = godotenv.Load()

	// Create executor
	config := &executor.Config{
		DryRun:       false,
		NoConfirm:    false,
		Verbose:      viper.GetBool("verbose"),
		Quiet:        viper.GetBool("quiet"),
		NoColor:      viper.GetBool("no-color"),
		OutputFormat: "text",
		Model:        getModel(),
		StdioPath:    os.Getenv("WEAVE_MCP_STDIO_PATH"),
		MaxRetries:   3,
	}

	exec, err := executor.NewExecutor(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create executor: %w", err)
	}

	// Create readline instance
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "\033[36m>\033[0m ",
		HistoryFile:     os.ExpandEnv("$HOME/.weave_history"),
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create readline: %w", err)
	}

	return &REPL{
		executor:    exec,
		rl:          rl,
		interrupted: false,
	}, nil
}

// Run starts the REPL
func (r *REPL) Run() error {
	defer r.rl.Close()
	defer r.executor.Close()

	// Track interrupt count for double CTRL-C to exit
	interruptCount := 0
	var interruptTimer *time.Timer

	// Display welcome banner
	r.displayBanner()

	// Main REPL loop
	for {
		r.interrupted = false

		line, err := r.rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				// Handle CTRL-C from readline
				interruptCount++
				if interruptCount == 1 {
					r.interrupted = true
					fmt.Println("\nCommand interrupted. Press CTRL-C again to exit.")
					// Reset counter after 2 seconds
					if interruptTimer != nil {
						interruptTimer.Stop()
					}
					interruptTimer = time.AfterFunc(2*time.Second, func() {
						interruptCount = 0
					})
					continue
				} else {
					fmt.Println("\nExiting weave CLI...")
					return nil
				}
			} else if err == io.EOF {
				fmt.Println("\nExiting weave CLI...")
				return nil
			}
			return err
		}

		// Reset interrupt count on successful readline
		interruptCount = 0
		if interruptTimer != nil {
			interruptTimer.Stop()
		}

		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Handle special commands
		if r.handleSpecialCommand(line) {
			continue
		}

		// Execute query
		r.executeQuery(line)
	}
}

// displayBanner shows the welcome banner
func (r *REPL) displayBanner() {
	cyan := color.New(color.FgCyan).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()
	dim := color.New(color.Faint).SprintFunc()

	banner := `
 __      __
/  \    /  \ ____ _____ ___  __ ____
\   \/\/   // __ \\__  \\  \/ // __ \
 \        /\  ___/ / __ \\   /\  ___/
  \__/\  /  \___  >____  /\_/  \___  >
       \/       \/     \/          \/
`

	fmt.Println(cyan(banner))
	fmt.Println(bold("  Weave CLI - AI-Powered Vector Database Management"))
	fmt.Println(dim("  https://github.com/maximilien/weave-cli"))
	fmt.Println()
	fmt.Println("  Use " + cyan("natural language") + " to manage your vector databases")
	fmt.Println("  Type " + cyan("/help") + " for commands, " + cyan("/exit") + " to quit")
	fmt.Println("  Press " + cyan("CTRL-C") + " to stop current command, twice to exit")
	fmt.Println()
}

// handleSpecialCommand handles built-in commands
func (r *REPL) handleSpecialCommand(line string) bool {
	switch strings.ToLower(line) {
	case "/exit", "/quit", "exit", "quit":
		fmt.Println("Exiting weave CLI...")
		r.rl.Close()
		r.executor.Close()
		os.Exit(0)
		return true

	case "/help", "help":
		r.displayHelp()
		return true

	case "/clear", "clear":
		fmt.Print("\033[H\033[2J")
		r.displayBanner()
		return true

	case "/history":
		r.displayHistory()
		return true

	case "/examples":
		r.displayExamples()
		return true
	}

	return false
}

// executeQuery executes a natural language query
func (r *REPL) executeQuery(query string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Check for interruption
	done := make(chan bool)
	go func() {
		_, err := r.executor.Execute(ctx, query)
		if err != nil && !r.interrupted {
			color.Red("\n❌ Error: %v\n", err)
		}
		done <- true
	}()

	select {
	case <-done:
		// Execution completed
	case <-ctx.Done():
		if !r.interrupted {
			color.Yellow("\n⚠️  Query timed out\n")
		}
	}

	fmt.Println()
}

// displayHelp shows help information
func (r *REPL) displayHelp() {
	cyan := color.New(color.FgCyan).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()

	fmt.Println()
	fmt.Println(bold("Available Commands:"))
	fmt.Println()
	fmt.Println("  " + cyan("/help") + "      - Show this help message")
	fmt.Println("  " + cyan("/examples") + "  - Show example queries")
	fmt.Println("  " + cyan("/history") + "   - Show command history")
	fmt.Println("  " + cyan("/clear") + "     - Clear the screen")
	fmt.Println("  " + cyan("/exit") + "      - Exit the REPL")
	fmt.Println()
	fmt.Println(bold("Natural Language Queries:"))
	fmt.Println()
	fmt.Println("  Just type what you want to do:")
	fmt.Println("    • show me all my collections")
	fmt.Println("    • find empty collections")
	fmt.Println("    • add files from /tmp/docs to MyCollection")
	fmt.Println("    • count documents in TestDocs")
	fmt.Println()
	fmt.Println(bold("Keyboard Shortcuts:"))
	fmt.Println()
	fmt.Println("  " + cyan("CTRL-C") + "     - Stop current command")
	fmt.Println("  " + cyan("CTRL-C x2") + "  - Exit the REPL")
	fmt.Println("  " + cyan("CTRL-D") + "     - Exit the REPL")
	fmt.Println("  " + cyan("↑/↓") + "        - Navigate command history")
	fmt.Println()
}

// displayHistory shows command history
func (r *REPL) displayHistory() {
	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Println()
	fmt.Println(cyan("Command History:"))
	fmt.Println()
	fmt.Println("  History is stored in: " + os.ExpandEnv("$HOME/.weave_history"))
	fmt.Println("  Use ↑/↓ arrows to navigate history")
	fmt.Println()
}

// displayExamples shows example queries
func (r *REPL) displayExamples() {
	cyan := color.New(color.FgCyan).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()

	fmt.Println()
	fmt.Println(bold("Example Queries:"))
	fmt.Println()
	fmt.Println(cyan("  Collection Management:"))
	fmt.Println("    • show me all my collections")
	fmt.Println("    • create a text collection called MyDocs")
	fmt.Println("    • delete the collection TestCollection")
	fmt.Println()
	fmt.Println(cyan("  Document Operations:"))
	fmt.Println("    • list documents in MyCollection")
	fmt.Println("    • count documents in each collection")
	fmt.Println("    • add document.txt to MyCollection")
	fmt.Println("    • delete all documents in TestCollection")
	fmt.Println()
	fmt.Println(cyan("  Batch Operations:"))
	fmt.Println("    • add all files from /tmp/docs to MyCollection")
	fmt.Println("    • process PDFs in ./documents directory")
	fmt.Println("    • convert CMYK PDFs in ./pdfs to RGB")
	fmt.Println()
	fmt.Println(cyan("  Search & Query:"))
	fmt.Println("    • find empty collections")
	fmt.Println("    • search for documents about AI in MyDocs")
	fmt.Println("    • show me collections with more than 100 documents")
	fmt.Println()
}

// getModel returns the model to use
func getModel() string {
	if model := os.Getenv("OPENAI_MODEL"); model != "" {
		return model
	}
	return "gpt-4o"
}
