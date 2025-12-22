// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package repl

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/executor"
	"github.com/maximilien/weave-cli/src/pkg/version"
	"github.com/spf13/viper"
)

// Options holds REPL configuration options
type Options struct {
	QueryStringsFile string
	NoConfirm        bool
}

// REPL represents the interactive Read-Eval-Print Loop
type REPL struct {
	executor         *executor.Executor
	rl               *readline.Instance
	interrupted      bool
	queryStringsFile string
	batchMode        bool
	queries          []string
	queryIndex       int
	currentVDB       string // Current VDB connection
	currentCollection string // Current active collection
}

// New creates a new REPL instance
func New() (*REPL, error) {
	return NewWithOptions(Options{})
}

// NewWithOptions creates a new REPL instance with options
func NewWithOptions(opts Options) (*REPL, error) {
	// Load .env file with proper precedence (local → global)
	configPaths, err := config.FindConfigPaths()
	if err == nil && configPaths.EnvPath != "" {
		_ = godotenv.Load(configPaths.EnvPath)
	} else {
		// Fallback to current directory
		_ = godotenv.Load()
	}

	// Create executor
	config := &executor.Config{
		DryRun:       false,
		NoConfirm:    opts.NoConfirm || viper.GetBool("no-confirm"),
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

	// Initialize REPL struct
	repl := &REPL{
		executor:         exec,
		interrupted:      false,
		queryStringsFile: opts.QueryStringsFile,
		batchMode:        opts.QueryStringsFile != "",
	}

	// Load queries from file if in batch mode
	if repl.batchMode {
		queries, err := loadQueriesFromFile(opts.QueryStringsFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load queries: %w", err)
		}
		repl.queries = queries
		repl.queryIndex = 0
	} else {
		// Create readline instance for interactive mode with tab completion
		rl, err := readline.NewEx(CreateReadlineConfig())
		if err != nil {
			return nil, fmt.Errorf("failed to create readline: %w", err)
		}
		repl.rl = rl
	}

	return repl, nil
}

// loadQueriesFromFile loads queries from a file (one per line)
func loadQueriesFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var queries []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments
		if line != "" && !strings.HasPrefix(line, "#") {
			queries = append(queries, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return queries, nil
}

// Run starts the REPL
func (r *REPL) Run() error {
	// Cleanup
	defer r.executor.Close()
	if r.rl != nil {
		defer r.rl.Close()
	}

	// Display welcome banner
	r.displayBanner()

	// Batch mode: execute all queries sequentially
	if r.batchMode {
		return r.runBatchMode()
	}

	// Interactive mode
	return r.runInteractiveMode()
}

// runBatchMode executes queries from file sequentially
func (r *REPL) runBatchMode() error {
	fmt.Printf("📝 Batch mode: executing %d queries\n\n", len(r.queries))

	for i, query := range r.queries {
		// Show query number and text
		fmt.Printf("Query %d/%d: %s\n", i+1, len(r.queries), query)

		// Execute the query
		r.executeQuery(query)

		// Add separator between queries
		if i < len(r.queries)-1 {
			fmt.Println("\n" + strings.Repeat("─", 60) + "\n")
		}
	}

	fmt.Printf("\n✅ Completed %d queries\n", len(r.queries))
	return nil
}

// runInteractiveMode runs the traditional interactive REPL
func (r *REPL) runInteractiveMode() error {
	// Track interrupt count for double CTRL-C to exit
	interruptCount := 0
	var interruptTimer *time.Timer

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
	fmt.Println(dim(fmt.Sprintf("  v%s", version.Version)))
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
	// Parse command and arguments
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return false
	}

	cmd := strings.ToLower(parts[0])
	args := parts[1:]

	switch cmd {
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

	case "/mcp":
		r.handleMCPCommand(args)
		return true

	case "/collection", "/col":
		r.handleCollectionCommand(args)
		return true

	case "/search":
		r.handleSearchCommand(args)
		return true

	case "/stats":
		r.handleStatsCommand(args)
		return true

	case "/use":
		r.handleUseCommand(args)
		return true

	case "/status":
		r.displayStatus()
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
	os.Stdout.Sync() // Ensure all output is flushed
}

// displayHelp shows help information
func (r *REPL) displayHelp() {
	cyan := color.New(color.FgCyan).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()

	fmt.Println()
	fmt.Println(bold("Available Commands:"))
	fmt.Println()
	fmt.Println(bold("  General:"))
	fmt.Println("    " + cyan("/help") + "      - Show this help message")
	fmt.Println("    " + cyan("/examples") + "  - Show example queries")
	fmt.Println("    " + cyan("/history") + "   - Show command history")
	fmt.Println("    " + cyan("/clear") + "     - Clear the screen")
	fmt.Println("    " + cyan("/status") + "    - Show current status")
	fmt.Println("    " + cyan("/exit") + "      - Exit the REPL")
	fmt.Println()
	fmt.Println(bold("  MCP Tools:"))
	fmt.Println("    " + cyan("/mcp list") + "              - List available MCP tools")
	fmt.Println("    " + cyan("/mcp call <tool> [args]") + " - Call an MCP tool")
	fmt.Println("    " + cyan("/mcp status") + "            - Show MCP connection status")
	fmt.Println()
	fmt.Println(bold("  Collections:"))
	fmt.Println("    " + cyan("/collection list") + "          - List all collections")
	fmt.Println("    " + cyan("/use <name>") + "              - Set active collection")
	fmt.Println("    " + cyan("/collection create <name>") + " - Create new collection")
	fmt.Println("    " + cyan("/collection delete <name>") + " - Delete collection")
	fmt.Println("    " + cyan("/collection info") + "         - Show current collection info")
	fmt.Println()
	fmt.Println(bold("  Search & Query:"))
	fmt.Println("    " + cyan("/search <query>") + "  - Semantic search")
	fmt.Println("    " + cyan("/stats") + "           - Show collection statistics")
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
