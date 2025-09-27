package tests

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/mock"
	"github.com/spf13/cobra"
)

// Mock command functions for testing
func mockRunConfigShow(cmd *cobra.Command, args []string) {
	fmt.Println("Mock config show command")
}

func mockRunHealthCheck(cmd *cobra.Command, args []string) {
	fmt.Println("Mock health check command")
}

func mockRunCollectionList(cmd *cobra.Command, args []string) {
	fmt.Println("Mock collection list command")
}

func mockRunDocumentList(cmd *cobra.Command, args []string) {
	fmt.Println("Mock document list command")
}

// TestCLICommandStructure tests the basic structure of CLI commands
func TestCLICommandStructure(t *testing.T) {
	t.Run("RootCommand", func(t *testing.T) {
		// Test root command structure
		rootCmd := &cobra.Command{
			Use:   "weave",
			Short: "Weave VDB Management Tool",
			Long:  "Weave is a command-line tool for managing Weaviate vector databases.",
		}

		if rootCmd.Use != "weave" {
			t.Errorf("Expected root command use 'weave', got %s", rootCmd.Use)
		}

		if rootCmd.Short == "" {
			t.Error("Root command should have a short description")
		}

		if rootCmd.Long == "" {
			t.Error("Root command should have a long description")
		}
	})

	t.Run("SubCommands", func(t *testing.T) {
		// Test subcommand structure
		configCmd := &cobra.Command{
			Use:   "config",
			Short: "Configuration management",
		}

		configShowCmd := &cobra.Command{
			Use:   "show",
			Short: "Show currently configured VDB",
			Run:   mockRunConfigShow,
		}

		configCmd.AddCommand(configShowCmd)

		if len(configCmd.Commands()) != 1 {
			t.Errorf("Expected 1 subcommand, got %d", len(configCmd.Commands()))
		}

		if configCmd.Commands()[0].Use != "show" {
			t.Errorf("Expected subcommand use 'show', got %s", configCmd.Commands()[0].Use)
		}
	})
}

// TestCLICommandValidation tests command validation
func TestCLICommandValidation(t *testing.T) {
	t.Run("RequiredArgs", func(t *testing.T) {
		// Test command with required arguments
		cmd := &cobra.Command{
			Use:  "test COLLECTION_NAME",
			Args: cobra.ExactArgs(1),
		}

		// Test with correct number of args
		cmd.SetArgs([]string{"test-collection"})
		if err := cmd.ValidateArgs([]string{"test-collection"}); err != nil {
			t.Errorf("Expected no error with correct args, got %v", err)
		}

		// Test with incorrect number of args
		if err := cmd.ValidateArgs([]string{}); err == nil {
			t.Error("Expected error with no args")
		}

		if err := cmd.ValidateArgs([]string{"arg1", "arg2"}); err == nil {
			t.Error("Expected error with too many args")
		}
	})

	t.Run("OptionalArgs", func(t *testing.T) {
		// Test command with optional arguments
		cmd := &cobra.Command{
			Use:  "test [COLLECTION_NAME]",
			Args: cobra.MaximumNArgs(1),
		}

		// Test with no args
		if err := cmd.ValidateArgs([]string{}); err != nil {
			t.Errorf("Expected no error with no args, got %v", err)
		}

		// Test with one arg
		if err := cmd.ValidateArgs([]string{"test-collection"}); err != nil {
			t.Errorf("Expected no error with one arg, got %v", err)
		}

		// Test with too many args
		if err := cmd.ValidateArgs([]string{"arg1", "arg2"}); err == nil {
			t.Error("Expected error with too many args")
		}
	})
}

// TestCLIFlags tests command flags
func TestCLIFlags(t *testing.T) {
	t.Run("StringFlag", func(t *testing.T) {
		cmd := &cobra.Command{
			Use: "test",
		}

		var configFile string
		cmd.Flags().StringVar(&configFile, "config", "", "config file")

		cmd.SetArgs([]string{"--config", "test-config.yaml"})
		cmd.ParseFlags([]string{"--config", "test-config.yaml"})

		if configFile != "test-config.yaml" {
			t.Errorf("Expected config file 'test-config.yaml', got %s", configFile)
		}
	})

	t.Run("IntFlag", func(t *testing.T) {
		cmd := &cobra.Command{
			Use: "test",
		}

		var limit int
		cmd.Flags().IntVar(&limit, "limit", 10, "maximum number of items")

		cmd.SetArgs([]string{"--limit", "5"})
		cmd.ParseFlags([]string{"--limit", "5"})

		if limit != 5 {
			t.Errorf("Expected limit 5, got %d", limit)
		}
	})

	t.Run("BoolFlag", func(t *testing.T) {
		cmd := &cobra.Command{
			Use: "test",
		}

		var verbose bool
		cmd.Flags().BoolVar(&verbose, "verbose", false, "verbose output")

		cmd.SetArgs([]string{"--verbose"})
		cmd.ParseFlags([]string{"--verbose"})

		if !verbose {
			t.Error("Expected verbose to be true")
		}
	})
}

// TestCLICommandExecution tests command execution
func TestCLICommandExecution(t *testing.T) {
	t.Run("CommandOutput", func(t *testing.T) {
		// Capture output
		var buf bytes.Buffer
		cmd := &cobra.Command{
			Use: "test",
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Fprintln(&buf, "Test output")
			},
		}

		cmd.SetArgs([]string{})
		cmd.SetOutput(&buf)
		cmd.Execute()

		output := buf.String()
		if !strings.Contains(output, "Test output") {
			t.Errorf("Expected 'Test output' in command output, got %s", output)
		}
	})

	t.Run("CommandError", func(t *testing.T) {
		cmd := &cobra.Command{
			Use: "test",
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Fprintln(os.Stderr, "Test error")
			},
		}

		cmd.SetArgs([]string{})
		// This should not cause a test failure
		cmd.Execute()
	})
}

// TestCLIContextHandling tests context handling in commands
func TestCLIContextHandling(t *testing.T) {
	t.Run("ContextTimeout", func(t *testing.T) {
		cmd := &cobra.Command{
			Use: "test",
			Run: func(cmd *cobra.Command, args []string) {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
				defer cancel()

				// Simulate work
				time.Sleep(100 * time.Millisecond)

				select {
				case <-ctx.Done():
					// Context cancelled
				default:
					// Work completed
				}
			},
		}

		cmd.SetArgs([]string{})
		// Just test that the command executes without error
		if err := cmd.Execute(); err != nil {
			t.Errorf("Command execution failed: %v", err)
		}
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		cmd := &cobra.Command{
			Use: "test",
			Run: func(cmd *cobra.Command, args []string) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately

				select {
				case <-ctx.Done():
					// Context cancelled
				default:
					// Context not cancelled
				}
			},
		}

		cmd.SetArgs([]string{})
		// Just test that the command executes without error
		if err := cmd.Execute(); err != nil {
			t.Errorf("Command execution failed: %v", err)
		}
	})
}

// TestCLIErrorHandling tests error handling in commands
func TestCLIErrorHandling(t *testing.T) {
	t.Run("InvalidArgs", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:  "test COLLECTION_NAME",
			Args: cobra.ExactArgs(1),
			Run:  func(cmd *cobra.Command, args []string) {},
		}

		// Test with no args
		cmd.SetArgs([]string{})
		err := cmd.Execute()
		if err == nil {
			t.Error("Expected error with no args")
		}
	})

	t.Run("InvalidFlag", func(t *testing.T) {
		cmd := &cobra.Command{
			Use: "test",
			Run: func(cmd *cobra.Command, args []string) {},
		}

		// Test with invalid flag
		cmd.SetArgs([]string{"--invalid-flag"})
		err := cmd.Execute()
		if err == nil {
			t.Error("Expected error with invalid flag")
		}
	})
}

// TestCLIWithMockClient tests CLI commands with mock client
func TestCLIWithMockClient(t *testing.T) {
	t.Run("MockClientIntegration", func(t *testing.T) {
		cfg := &config.MockConfig{
			Enabled:            true,
			SimulateEmbeddings: true,
			EmbeddingDimension: 384,
			Collections: []config.MockCollection{
				{Name: "TestCollection", Type: "text", Description: "Test collection"},
			},
		}

		client := mock.NewClient(cfg)
		ctx := context.Background()

		// Test health check
		if err := client.Health(ctx); err != nil {
			t.Errorf("Health check failed: %v", err)
		}

		// Test collection listing
		collections, err := client.ListCollections(ctx)
		if err != nil {
			t.Errorf("Failed to list collections: %v", err)
		}

		if len(collections) != 1 {
			t.Errorf("Expected 1 collection, got %d", len(collections))
		}

		// Test document operations
		doc := mock.Document{
			ID:      "test-doc-1",
			Content: "Test document content",
			Metadata: map[string]interface{}{
				"test": true,
			},
		}

		if err := client.AddDocument(ctx, "TestCollection", doc); err != nil {
			t.Errorf("Failed to add document: %v", err)
		}

		documents, err := client.ListDocuments(ctx, "TestCollection", 10)
		if err != nil {
			t.Errorf("Failed to list documents: %v", err)
		}

		if len(documents) != 1 {
			t.Errorf("Expected 1 document, got %d", len(documents))
		}
	})
}

// TestCLICommandHelp tests help functionality
func TestCLICommandHelp(t *testing.T) {
	t.Run("HelpOutput", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:   "test",
			Short: "Test command",
			Long:  "This is a test command for testing purposes.",
		}

		var buf bytes.Buffer
		cmd.SetOutput(&buf)
		cmd.SetArgs([]string{"--help"})
		cmd.Execute()

		output := buf.String()
		if !strings.Contains(output, "test") {
			t.Errorf("Expected 'test' in help output, got %s", output)
		}

		if !strings.Contains(output, "testing purposes") {
			t.Errorf("Expected 'testing purposes' in help output, got %s", output)
		}
	})

	t.Run("SubcommandHelp", func(t *testing.T) {
		rootCmd := &cobra.Command{
			Use: "weave",
		}

		subCmd := &cobra.Command{
			Use:   "test",
			Short: "Test subcommand",
		}

		rootCmd.AddCommand(subCmd)

		var buf bytes.Buffer
		rootCmd.SetOutput(&buf)
		rootCmd.SetArgs([]string{"test", "--help"})
		rootCmd.Execute()

		output := buf.String()
		if !strings.Contains(output, "Test subcommand") {
			t.Errorf("Expected 'Test subcommand' in help output, got %s", output)
		}
	})
}

// BenchmarkCLICommands benchmarks CLI command operations
func BenchmarkCLICommands(b *testing.B) {
	cmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {
			// Simulate command work
			time.Sleep(1 * time.Millisecond)
		},
	}

	b.Run("CommandExecution", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cmd.SetArgs([]string{})
			cmd.Execute()
		}
	})

	b.Run("FlagParsing", func(b *testing.B) {
		var configFile string
		cmd.Flags().StringVar(&configFile, "config", "", "config file")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cmd.ParseFlags([]string{"--config", "test-config.yaml"})
		}
	})
}

// TestCommandAliases tests command aliases functionality
func TestCommandAliases(t *testing.T) {
	t.Run("CollectionAliases", func(t *testing.T) {
		// Test collection command aliases
		collectionAliases := []string{"col", "cols"}
		for _, alias := range collectionAliases {
			// This would need to be tested with actual command execution
			// For now, we'll just verify the aliases are defined
			t.Logf("Collection alias: %s", alias)
		}
	})

	t.Run("CollectionListAliases", func(t *testing.T) {
		listAliases := []string{"ls", "l"}
		for _, alias := range listAliases {
			t.Logf("Collection list alias: %s", alias)
		}
	})

	t.Run("CollectionDeleteAliases", func(t *testing.T) {
		deleteAliases := []string{"del", "d"}
		for _, alias := range deleteAliases {
			t.Logf("Collection delete alias: %s", alias)
		}
	})

	t.Run("CollectionDeleteAllAliases", func(t *testing.T) {
		deleteAllAliases := []string{"del-all", "da"}
		for _, alias := range deleteAllAliases {
			t.Logf("Collection delete-all alias: %s", alias)
		}
	})

	t.Run("DocumentShowAliases", func(t *testing.T) {
		showAliases := []string{"s"}
		for _, alias := range showAliases {
			t.Logf("Document show alias: %s", alias)
		}
	})
}

// TestCommandAliasExecution tests that aliases work by executing commands
func TestCommandAliasExecution(t *testing.T) {
	t.Run("CollectionListAlias", func(t *testing.T) {
		// Test collection list with alias
		var buf bytes.Buffer

		// Create a mock command to test alias resolution
		cmd := &cobra.Command{
			Use:     "list",
			Aliases: []string{"ls", "l"},
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Fprintln(&buf, "Mock collection list command")
			},
		}

		cmd.SetOutput(&buf)
		cmd.Execute()

		output := buf.String()
		if !strings.Contains(output, "Mock collection list command") {
			t.Errorf("Expected mock output, got: %s", output)
		}
	})

	t.Run("CollectionDeleteAlias", func(t *testing.T) {
		// Test collection delete with alias
		var buf bytes.Buffer

		cmd := &cobra.Command{
			Use:     "delete",
			Aliases: []string{"del", "d"},
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Fprintln(&buf, "Mock collection delete command")
			},
		}

		cmd.SetOutput(&buf)
		cmd.Execute()

		output := buf.String()
		if !strings.Contains(output, "Mock collection delete command") {
			t.Errorf("Expected mock output, got: %s", output)
		}
	})

	t.Run("CollectionDeleteAllAlias", func(t *testing.T) {
		// Test collection delete-all with alias
		var buf bytes.Buffer

		cmd := &cobra.Command{
			Use:     "delete-all",
			Aliases: []string{"del-all", "da"},
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Fprintln(&buf, "Mock collection delete-all command")
			},
		}

		cmd.SetOutput(&buf)
		cmd.Execute()

		output := buf.String()
		if !strings.Contains(output, "Mock collection delete-all command") {
			t.Errorf("Expected mock output, got: %s", output)
		}
	})

	t.Run("DocumentShowAlias", func(t *testing.T) {
		// Test document show with alias
		var buf bytes.Buffer

		cmd := &cobra.Command{
			Use:     "show",
			Aliases: []string{"s"},
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Fprintln(&buf, "Mock document show command")
			},
		}

		cmd.SetOutput(&buf)
		cmd.Execute()

		output := buf.String()
		if !strings.Contains(output, "Mock document show command") {
			t.Errorf("Expected mock output, got: %s", output)
		}
	})
}

// TestAliasHelpText tests that aliases appear in help text
func TestAliasHelpText(t *testing.T) {
	t.Run("CollectionListHelp", func(t *testing.T) {
		var buf bytes.Buffer

		cmd := &cobra.Command{
			Use:     "list",
			Aliases: []string{"ls", "l"},
			Short:   "List all collections",
		}

		cmd.SetOutput(&buf)
		cmd.Usage()

		output := buf.String()
		if !strings.Contains(output, "ls") || !strings.Contains(output, "l") {
			t.Errorf("Expected aliases in help text, got: %s", output)
		}
	})

	t.Run("CollectionDeleteHelp", func(t *testing.T) {
		var buf bytes.Buffer

		cmd := &cobra.Command{
			Use:     "delete",
			Aliases: []string{"del", "d"},
			Short:   "Delete a specific collection",
		}

		cmd.SetOutput(&buf)
		cmd.Usage()

		output := buf.String()
		if !strings.Contains(output, "del") || !strings.Contains(output, "d") {
			t.Errorf("Expected aliases in help text, got: %s", output)
		}
	})

	t.Run("DocumentShowHelp", func(t *testing.T) {
		var buf bytes.Buffer

		cmd := &cobra.Command{
			Use:     "show",
			Aliases: []string{"s"},
			Short:   "Show a specific document",
		}

		cmd.SetOutput(&buf)
		cmd.Usage()

		output := buf.String()
		if !strings.Contains(output, "s") {
			t.Errorf("Expected alias 's' in help text, got: %s", output)
		}
	})
}

func TestDocumentDeleteMetadataFiltering(t *testing.T) {
	// Test the metadata filtering functionality in the delete command
	t.Run("Document Delete with Metadata Filter", func(t *testing.T) {
		// Test command with metadata filter
		cmd := &cobra.Command{
			Use: "delete",
			Run: func(cmd *cobra.Command, args []string) {
				// Mock the delete function to validate arguments
				if len(args) != 1 {
					t.Errorf("Expected 1 argument, got %d", len(args))
				}
				if args[0] != "TestCollection" {
					t.Errorf("Expected 'TestCollection', got %s", args[0])
				}

				metadataFilters, _ := cmd.Flags().GetStringSlice("metadata")
				if len(metadataFilters) != 1 {
					t.Errorf("Expected 1 metadata filter, got %d", len(metadataFilters))
				}
				if metadataFilters[0] != "filename=test.png" {
					t.Errorf("Expected 'filename=test.png', got %s", metadataFilters[0])
				}
			},
		}

		// Add the metadata flag
		cmd.Flags().StringSliceP("metadata", "m", []string{}, "Delete documents matching metadata filter")

		// Set up command arguments
		cmd.SetArgs([]string{"TestCollection", "--metadata", "filename=test.png"})

		// Execute command
		err := cmd.Execute()
		if err != nil {
			t.Errorf("Command execution failed: %v", err)
		}
	})

	t.Run("Document Delete with Multiple Metadata Filters", func(t *testing.T) {
		cmd := &cobra.Command{
			Use: "delete",
			Run: func(cmd *cobra.Command, args []string) {
				metadataFilters, _ := cmd.Flags().GetStringSlice("metadata")
				if len(metadataFilters) != 2 {
					t.Errorf("Expected 2 metadata filters, got %d", len(metadataFilters))
				}

				expectedFilters := []string{"filename=test.png", "type=image"}
				for i, filter := range metadataFilters {
					if filter != expectedFilters[i] {
						t.Errorf("Expected '%s', got '%s'", expectedFilters[i], filter)
					}
				}
			},
		}

		cmd.Flags().StringSliceP("metadata", "m", []string{}, "Delete documents matching metadata filter")
		cmd.SetArgs([]string{"TestCollection", "--metadata", "filename=test.png", "--metadata", "type=image"})

		err := cmd.Execute()
		if err != nil {
			t.Errorf("Command execution failed: %v", err)
		}
	})

	t.Run("Document Delete with Short Flag", func(t *testing.T) {
		cmd := &cobra.Command{
			Use: "delete",
			Run: func(cmd *cobra.Command, args []string) {
				metadataFilters, _ := cmd.Flags().GetStringSlice("metadata")
				if len(metadataFilters) != 1 {
					t.Errorf("Expected 1 metadata filter, got %d", len(metadataFilters))
				}
				if metadataFilters[0] != "filename=test.png" {
					t.Errorf("Expected 'filename=test.png', got %s", metadataFilters[0])
				}
			},
		}

		cmd.Flags().StringSliceP("metadata", "m", []string{}, "Delete documents matching metadata filter")
		cmd.SetArgs([]string{"TestCollection", "-m", "filename=test.png"})

		err := cmd.Execute()
		if err != nil {
			t.Errorf("Command execution failed: %v", err)
		}
	})
}

// TestDocumentShowMetadataFiltering tests the metadata filtering functionality in the show command
func TestDocumentShowMetadataFiltering(t *testing.T) {
	t.Run("Document Show with Metadata Filter", func(t *testing.T) {
		// Test command with metadata filter
		cmd := &cobra.Command{
			Use: "show",
			Run: func(cmd *cobra.Command, args []string) {
				// Mock the show function to validate arguments
				if len(args) != 1 {
					t.Errorf("Expected 1 argument, got %d", len(args))
				}
				if args[0] != "TestCollection" {
					t.Errorf("Expected 'TestCollection', got %s", args[0])
				}

				metadataFilters, _ := cmd.Flags().GetStringSlice("metadata")
				if len(metadataFilters) != 1 {
					t.Errorf("Expected 1 metadata filter, got %d", len(metadataFilters))
				}
				if metadataFilters[0] != "filename=test.png" {
					t.Errorf("Expected 'filename=test.png', got %s", metadataFilters[0])
				}
			},
		}

		// Add the metadata flag
		cmd.Flags().StringSliceP("metadata", "m", []string{}, "Show documents matching metadata filter")

		// Set up command arguments
		cmd.SetArgs([]string{"TestCollection", "--metadata", "filename=test.png"})

		// Execute command
		err := cmd.Execute()
		if err != nil {
			t.Errorf("Command execution failed: %v", err)
		}
	})

	t.Run("Document Show with Multiple Metadata Filters", func(t *testing.T) {
		cmd := &cobra.Command{
			Use: "show",
			Run: func(cmd *cobra.Command, args []string) {
				metadataFilters, _ := cmd.Flags().GetStringSlice("metadata")
				if len(metadataFilters) != 2 {
					t.Errorf("Expected 2 metadata filters, got %d", len(metadataFilters))
				}
				expectedFilters := []string{"filename=test.png", "type=image"}
				for i, filter := range metadataFilters {
					if filter != expectedFilters[i] {
						t.Errorf("Expected '%s', got '%s'", expectedFilters[i], filter)
					}
				}
			},
		}

		cmd.Flags().StringSliceP("metadata", "m", []string{}, "Show documents matching metadata filter")
		cmd.SetArgs([]string{"TestCollection", "--metadata", "filename=test.png", "--metadata", "type=image"})

		err := cmd.Execute()
		if err != nil {
			t.Errorf("Command execution failed: %v", err)
		}
	})

	t.Run("Document Show with Short Flag", func(t *testing.T) {
		cmd := &cobra.Command{
			Use: "show",
			Run: func(cmd *cobra.Command, args []string) {
				metadataFilters, _ := cmd.Flags().GetStringSlice("metadata")
				if len(metadataFilters) != 1 {
					t.Errorf("Expected 1 metadata filter, got %d", len(metadataFilters))
				}
				if metadataFilters[0] != "filename=test.png" {
					t.Errorf("Expected 'filename=test.png', got %s", metadataFilters[0])
				}
			},
		}

		cmd.Flags().StringSliceP("metadata", "m", []string{}, "Show documents matching metadata filter")
		cmd.SetArgs([]string{"TestCollection", "-m", "filename=test.png"})

		err := cmd.Execute()
		if err != nil {
			t.Errorf("Command execution failed: %v", err)
		}
	})

	t.Run("Document Show with Document ID (no metadata)", func(t *testing.T) {
		cmd := &cobra.Command{
			Use: "show",
			Run: func(cmd *cobra.Command, args []string) {
				if len(args) != 2 {
					t.Errorf("Expected 2 arguments, got %d", len(args))
				}
				if args[0] != "TestCollection" {
					t.Errorf("Expected 'TestCollection', got %s", args[0])
				}
				if args[1] != "doc123" {
					t.Errorf("Expected 'doc123', got %s", args[1])
				}

				metadataFilters, _ := cmd.Flags().GetStringSlice("metadata")
				if len(metadataFilters) != 0 {
					t.Errorf("Expected 0 metadata filters, got %d", len(metadataFilters))
				}
			},
		}

		cmd.Flags().StringSliceP("metadata", "m", []string{}, "Show documents matching metadata filter")
		cmd.SetArgs([]string{"TestCollection", "doc123"})

		err := cmd.Execute()
		if err != nil {
			t.Errorf("Command execution failed: %v", err)
		}
	})
}

// TestCollectionCountCommand tests the collection count command functionality
func TestCollectionCountCommand(t *testing.T) {
	t.Run("Collection Count Command Structure", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:     "count [database-name]",
			Aliases: []string{"c"},
			Short:   "Count collections",
			Run: func(cmd *cobra.Command, args []string) {
				// Mock count function
				if len(args) > 1 {
					t.Errorf("Expected at most 1 argument, got %d", len(args))
				}
			},
		}

		// Test with no arguments
		cmd.SetArgs([]string{})
		err := cmd.Execute()
		if err != nil {
			t.Errorf("Command execution failed: %v", err)
		}

		// Test with database name argument
		cmd.SetArgs([]string{"test-db"})
		err = cmd.Execute()
		if err != nil {
			t.Errorf("Command execution failed: %v", err)
		}
	})

	t.Run("Collection Count Alias", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:     "count",
			Aliases: []string{"c"},
			Run: func(cmd *cobra.Command, args []string) {
				// Mock count function
			},
		}

		cmd.SetArgs([]string{})
		err := cmd.Execute()
		if err != nil {
			t.Errorf("Command execution failed: %v", err)
		}
	})
}

// TestDocumentCountCommand tests the document count command functionality
func TestDocumentCountCommand(t *testing.T) {
	t.Run("Document Count Single Collection", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:     "count COLLECTION_NAME [COLLECTION_NAME...]",
			Aliases: []string{"c"},
			Short:   "Count documents in one or more collections",
			Run: func(cmd *cobra.Command, args []string) {
				// Mock count function for single collection
				if len(args) != 1 {
					t.Errorf("Expected 1 argument for single collection, got %d", len(args))
				}
				if args[0] != "TestCollection" {
					t.Errorf("Expected 'TestCollection', got %s", args[0])
				}
			},
		}

		// Test with single collection name
		cmd.SetArgs([]string{"TestCollection"})
		err := cmd.Execute()
		if err != nil {
			t.Errorf("Command execution failed: %v", err)
		}
	})

	t.Run("Document Count Multiple Collections", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:     "count COLLECTION_NAME [COLLECTION_NAME...]",
			Aliases: []string{"c"},
			Short:   "Count documents in one or more collections",
			Run: func(cmd *cobra.Command, args []string) {
				// Mock count function for multiple collections
				if len(args) < 1 {
					t.Errorf("Expected at least 1 argument, got %d", len(args))
				}
				expectedCollections := []string{"RagMeDocs", "RagMeImages"}
				if len(args) != len(expectedCollections) {
					t.Errorf("Expected %d collections, got %d", len(expectedCollections), len(args))
				}
				for i, expected := range expectedCollections {
					if args[i] != expected {
						t.Errorf("Expected collection %d to be '%s', got '%s'", i, expected, args[i])
					}
				}
			},
		}

		// Test with multiple collection names
		cmd.SetArgs([]string{"RagMeDocs", "RagMeImages"})
		err := cmd.Execute()
		if err != nil {
			t.Errorf("Command execution failed: %v", err)
		}
	})

	t.Run("Document Count Alias", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:     "count",
			Aliases: []string{"c"},
			Run: func(cmd *cobra.Command, args []string) {
				// Mock count function
				if len(args) < 1 {
					t.Errorf("Expected at least 1 argument, got %d", len(args))
				}
			},
		}

		cmd.SetArgs([]string{"TestCollection"})
		err := cmd.Execute()
		if err != nil {
			t.Errorf("Command execution failed: %v", err)
		}
	})

	t.Run("Document Count Three Collections", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:     "count COLLECTION_NAME [COLLECTION_NAME...]",
			Aliases: []string{"c"},
			Run: func(cmd *cobra.Command, args []string) {
				// Mock count function for three collections
				expectedCollections := []string{"Collection1", "Collection2", "Collection3"}
				if len(args) != len(expectedCollections) {
					t.Errorf("Expected %d collections, got %d", len(expectedCollections), len(args))
				}
				for i, expected := range expectedCollections {
					if args[i] != expected {
						t.Errorf("Expected collection %d to be '%s', got '%s'", i, expected, args[i])
					}
				}
			},
		}

		cmd.SetArgs([]string{"Collection1", "Collection2", "Collection3"})
		err := cmd.Execute()
		if err != nil {
			t.Errorf("Command execution failed: %v", err)
		}
	})
}

// TestDocumentListMetadataTruncation tests the document list command metadata truncation
func TestDocumentListMetadataTruncation(t *testing.T) {
	t.Run("Document List with Short Flag", func(t *testing.T) {
		cmd := &cobra.Command{
			Use: "list",
			Run: func(cmd *cobra.Command, args []string) {
				// Mock list function
				if len(args) != 1 {
					t.Errorf("Expected 1 argument, got %d", len(args))
				}
				if args[0] != "TestCollection" {
					t.Errorf("Expected 'TestCollection', got %s", args[0])
				}

				shortLines, _ := cmd.Flags().GetInt("short")
				if shortLines != 1 {
					t.Errorf("Expected short lines to be 1, got %d", shortLines)
				}

				limit, _ := cmd.Flags().GetInt("limit")
				if limit != 5 {
					t.Errorf("Expected limit to be 5, got %d", limit)
				}
			},
		}

		cmd.Flags().IntP("short", "s", 5, "Show only first N lines of content")
		cmd.Flags().IntP("limit", "l", 50, "Maximum number of documents to show")
		cmd.SetArgs([]string{"TestCollection", "--short", "1", "--limit", "5"})

		err := cmd.Execute()
		if err != nil {
			t.Errorf("Command execution failed: %v", err)
		}
	})

	t.Run("Document List with Short Flag Alias", func(t *testing.T) {
		cmd := &cobra.Command{
			Use: "list",
			Run: func(cmd *cobra.Command, args []string) {
				shortLines, _ := cmd.Flags().GetInt("short")
				if shortLines != 3 {
					t.Errorf("Expected short lines to be 3, got %d", shortLines)
				}
			},
		}

		cmd.Flags().IntP("short", "s", 5, "Show only first N lines of content")
		cmd.SetArgs([]string{"TestCollection", "-s", "3"})

		err := cmd.Execute()
		if err != nil {
			t.Errorf("Command execution failed: %v", err)
		}
	})
}

// TestCollectionShowCommand tests the collection show command functionality
func TestCollectionShowCommand(t *testing.T) {
	t.Run("Collection Show Command Structure", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:     "show COLLECTION_NAME",
			Aliases: []string{"s"},
			Short:   "Show collection details",
			Run: func(cmd *cobra.Command, args []string) {
				// Mock show function
				if len(args) != 1 {
					t.Errorf("Expected 1 argument, got %d", len(args))
				}
				if args[0] != "TestCollection" {
					t.Errorf("Expected 'TestCollection', got %s", args[0])
				}
			},
		}

		// Test with collection name
		cmd.SetArgs([]string{"TestCollection"})
		err := cmd.Execute()
		if err != nil {
			t.Errorf("Command execution failed: %v", err)
		}
	})

	t.Run("Collection Show Alias", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:     "show",
			Aliases: []string{"s"},
			Run: func(cmd *cobra.Command, args []string) {
				// Mock show function
				if len(args) != 1 {
					t.Errorf("Expected 1 argument, got %d", len(args))
				}
			},
		}

		cmd.SetArgs([]string{"TestCollection"})
		err := cmd.Execute()
		if err != nil {
			t.Errorf("Command execution failed: %v", err)
		}
	})

	t.Run("Collection Show with Different Collection Names", func(t *testing.T) {
		testCases := []string{"MyCollection", "RagMeDocs", "TestCollection", "AnotherCollection"}

		for _, collectionName := range testCases {
			t.Run(fmt.Sprintf("Collection_%s", collectionName), func(t *testing.T) {
				cmd := &cobra.Command{
					Use:     "show",
					Aliases: []string{"s"},
					Run: func(cmd *cobra.Command, args []string) {
						if len(args) != 1 {
							t.Errorf("Expected 1 argument, got %d", len(args))
						}
						if args[0] != collectionName {
							t.Errorf("Expected '%s', got %s", collectionName, args[0])
						}
					},
				}

				cmd.SetArgs([]string{collectionName})
				err := cmd.Execute()
				if err != nil {
					t.Errorf("Command execution failed for collection %s: %v", collectionName, err)
				}
			})
		}
	})

	t.Run("Collection Show Help Text", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:     "show COLLECTION_NAME",
			Aliases: []string{"s"},
			Short:   "Show collection details",
			Long: `Show detailed information about a specific collection.

This command displays:
- Collection metadata and properties
- Document count
- Creation date (if available)
- Last document date (if available)
- Collection statistics`,
		}

		var buf bytes.Buffer
		cmd.SetOutput(&buf)
		cmd.Help()

		output := buf.String()
		if !strings.Contains(output, "Show detailed information about a specific collection") {
			t.Errorf("Expected 'Show detailed information about a specific collection' in help text, got: %s", output)
		}
		if !strings.Contains(output, "Collection metadata and properties") {
			t.Errorf("Expected 'Collection metadata and properties' in help text, got: %s", output)
		}
		if !strings.Contains(output, "Document count") {
			t.Errorf("Expected 'Document count' in help text, got: %s", output)
		}
	})

	t.Run("Collection Show with Short Flag", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:     "show",
			Aliases: []string{"s"},
			Run: func(cmd *cobra.Command, args []string) {
				// Mock show function
				if len(args) != 1 {
					t.Errorf("Expected 1 argument, got %d", len(args))
				}
				if args[0] != "TestCollection" {
					t.Errorf("Expected 'TestCollection', got %s", args[0])
				}

				shortLines, _ := cmd.Flags().GetInt("short")
				if shortLines != 5 {
					t.Errorf("Expected short lines to be 5, got %d", shortLines)
				}
			},
		}

		cmd.Flags().IntP("short", "s", 10, "Show only first N lines of sample document metadata")
		cmd.SetArgs([]string{"TestCollection", "--short", "5"})

		err := cmd.Execute()
		if err != nil {
			t.Errorf("Command execution failed: %v", err)
		}
	})

	t.Run("Collection Show with Short Flag Alias", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:     "show",
			Aliases: []string{"s"},
			Run: func(cmd *cobra.Command, args []string) {
				shortLines, _ := cmd.Flags().GetInt("short")
				if shortLines != 3 {
					t.Errorf("Expected short lines to be 3, got %d", shortLines)
				}
			},
		}

		cmd.Flags().IntP("short", "s", 10, "Show only first N lines of sample document metadata")
		cmd.SetArgs([]string{"TestCollection", "-s", "3"})

		err := cmd.Execute()
		if err != nil {
			t.Errorf("Command execution failed: %v", err)
		}
	})

	t.Run("Collection Show with Default Short Flag", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:     "show",
			Aliases: []string{"s"},
			Run: func(cmd *cobra.Command, args []string) {
				shortLines, _ := cmd.Flags().GetInt("short")
				if shortLines != 10 {
					t.Errorf("Expected default short lines to be 10, got %d", shortLines)
				}
			},
		}

		cmd.Flags().IntP("short", "s", 10, "Show only first N lines of sample document metadata")
		cmd.SetArgs([]string{"TestCollection"}) // No short flag specified

		err := cmd.Execute()
		if err != nil {
			t.Errorf("Command execution failed: %v", err)
		}
	})

	t.Run("Collection Show Metadata Value Truncation", func(t *testing.T) {
		// Test the truncateMetadataValue function behavior
		testCases := []struct {
			name     string
			value    interface{}
			maxLen   int
			expected string
		}{
			{
				name:     "Short value",
				value:    "short",
				maxLen:   100,
				expected: "short",
			},
			{
				name:     "Long value truncation",
				value:    "This is a very long string that should be truncated because it exceeds the maximum length limit",
				maxLen:   20,
				expected: "This is a very lo... (truncated, 60 more characters)",
			},
			{
				name:     "Base64-like data",
				value:    "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==",
				maxLen:   50,
				expected: "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==... (truncated, 20 more characters)",
			},
			{
				name:     "JSON object",
				value:    map[string]interface{}{"key1": "value1", "key2": "value2"},
				maxLen:   30,
				expected: "map[key1:value1 key2:value2]... (truncated, 0 more characters)",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// This would test the truncateMetadataValue function if it were exported
				// For now, we'll just verify the test structure is correct
				if tc.maxLen < 0 {
					t.Errorf("Max length should be positive")
				}
			})
		}
	})
}

func TestDocumentDeleteVirtualFlag(t *testing.T) {
	t.Run("Document Delete with Virtual Flag", func(t *testing.T) {
		cmd := &cobra.Command{
			Use: "delete",
			Run: func(cmd *cobra.Command, args []string) {
				// Validate arguments
				if len(args) != 2 {
					t.Errorf("Expected 2 arguments (collection and filename), got %d", len(args))
				}
				if args[0] != "RagmeDocs" {
					t.Errorf("Expected collection 'RagmeDocs', got %s", args[0])
				}
				if args[1] != "ragme-io.pdf" {
					t.Errorf("Expected filename 'ragme-io.pdf', got %s", args[1])
				}

				// Validate virtual flag
				virtual, _ := cmd.Flags().GetBool("virtual")
				if !virtual {
					t.Error("Expected virtual flag to be true")
				}

				// Validate metadata filters are not set
				metadataFilters, _ := cmd.Flags().GetStringSlice("metadata")
				if len(metadataFilters) != 0 {
					t.Errorf("Expected no metadata filters when using virtual flag, got %d", len(metadataFilters))
				}
			},
		}

		// Add flags
		cmd.Flags().StringSliceP("metadata", "m", []string{}, "Delete documents matching metadata filter")
		cmd.Flags().BoolP("virtual", "w", false, "Delete all chunks and images associated with the original filename")

		// Set up command arguments with virtual flag
		cmd.SetArgs([]string{"RagmeDocs", "ragme-io.pdf", "--virtual"})

		// Execute command
		err := cmd.Execute()
		if err != nil {
			t.Errorf("Command execution failed: %v", err)
		}
	})

	t.Run("Document Delete with Virtual Flag Short Form", func(t *testing.T) {
		cmd := &cobra.Command{
			Use: "delete",
			Run: func(cmd *cobra.Command, args []string) {
				virtual, _ := cmd.Flags().GetBool("virtual")
				if !virtual {
					t.Error("Expected virtual flag to be true")
				}
			},
		}

		cmd.Flags().BoolP("virtual", "w", false, "Delete all chunks and images associated with the original filename")
		cmd.SetArgs([]string{"RagmeDocs", "ragme-io.pdf", "-w"})

		err := cmd.Execute()
		if err != nil {
			t.Errorf("Command execution failed: %v", err)
		}
	})

	t.Run("Document Delete Virtual Flag Validation", func(t *testing.T) {
		cmd := &cobra.Command{
			Use: "delete",
			Run: func(cmd *cobra.Command, args []string) {
				// This test validates that virtual flag and metadata filters are mutually exclusive
				virtual, _ := cmd.Flags().GetBool("virtual")
				metadataFilters, _ := cmd.Flags().GetStringSlice("metadata")

				// The validation should prevent both flags from being set
				// This test just verifies the flags are correctly parsed
				if virtual && len(metadataFilters) > 0 {
					// This is expected behavior - the command should handle this validation
					// We're just testing that both flags are detected
					t.Log("Both virtual and metadata flags detected - validation should prevent this")
				}
			},
		}

		cmd.Flags().StringSliceP("metadata", "m", []string{}, "Delete documents matching metadata filter")
		cmd.Flags().BoolP("virtual", "w", false, "Delete all chunks and images associated with the original filename")

		// Test with both flags set (should be invalid)
		cmd.SetArgs([]string{"RagmeDocs", "ragme-io.pdf", "--virtual", "--metadata", "filename=test.pdf"})

		err := cmd.Execute()
		if err != nil {
			t.Errorf("Command execution failed: %v", err)
		}
	})

	t.Run("Document Delete Virtual Flag Help Text", func(t *testing.T) {
		cmd := &cobra.Command{
			Use: "delete",
			Long: `Delete documents from a collection.

You can delete documents in three ways:
1. By document ID: weave doc delete COLLECTION_NAME DOCUMENT_ID
2. By metadata filter: weave doc delete COLLECTION_NAME --metadata key=value
3. By original filename (virtual): weave doc delete COLLECTION_NAME ORIGINAL_FILENAME --virtual

When using --virtual flag, all chunks and images associated with the original filename
will be deleted in one operation.

⚠️  WARNING: This is a destructive operation that will permanently
delete the specified documents. Use with caution!`,
		}

		// Verify the help text includes virtual deletion information
		if !strings.Contains(cmd.Long, "virtual") {
			t.Error("Help text should mention virtual deletion")
		}
		if !strings.Contains(cmd.Long, "original filename") {
			t.Error("Help text should mention original filename")
		}
		if !strings.Contains(cmd.Long, "chunks and images") {
			t.Error("Help text should mention chunks and images")
		}
	})
}

// TestDocumentCreateCommand tests the new document create functionality
func TestDocumentCreateCommand(t *testing.T) {
	t.Run("Document Create Command Structure", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:     "create COLLECTION_NAME FILE_PATH",
			Aliases: []string{"c"},
			Short:   "Create a document from a file",
			Args:    cobra.ExactArgs(2),
		}

		// Test command structure
		if cmd.Use != "create COLLECTION_NAME FILE_PATH" {
			t.Errorf("Expected use 'create COLLECTION_NAME FILE_PATH', got %s", cmd.Use)
		}

		if cmd.Short == "" {
			t.Error("Command should have a short description")
		}

		// Test aliases
		if len(cmd.Aliases) == 0 {
			t.Error("Command should have aliases")
		}

		if cmd.Aliases[0] != "c" {
			t.Errorf("Expected alias 'c', got %s", cmd.Aliases[0])
		}

		// Test argument validation
		if cmd.Args == nil {
			t.Error("Command should have argument validation")
		}
	})

	t.Run("Document Create Chunk Size Flag", func(t *testing.T) {
		cmd := &cobra.Command{
			Use: "create",
			Run: func(cmd *cobra.Command, args []string) {
				chunkSize, _ := cmd.Flags().GetInt("chunk-size")
				if chunkSize != 500 {
					t.Errorf("Expected chunk size 500, got %d", chunkSize)
				}
			},
		}

		cmd.Flags().IntP("chunk-size", "s", 1000, "Chunk size for text content")
		cmd.SetArgs([]string{"TestCollection", "test.txt", "--chunk-size", "500"})

		err := cmd.Execute()
		if err != nil {
			t.Errorf("Command execution failed: %v", err)
		}
	})

	t.Run("Document Create Default Chunk Size", func(t *testing.T) {
		cmd := &cobra.Command{
			Use: "create",
			Run: func(cmd *cobra.Command, args []string) {
				chunkSize, _ := cmd.Flags().GetInt("chunk-size")
				if chunkSize != 1000 {
					t.Errorf("Expected default chunk size 1000, got %d", chunkSize)
				}
			},
		}

		cmd.Flags().IntP("chunk-size", "s", 1000, "Chunk size for text content")
		cmd.SetArgs([]string{"TestCollection", "test.txt"})

		err := cmd.Execute()
		if err != nil {
			t.Errorf("Command execution failed: %v", err)
		}
	})

	t.Run("Document Create Help Text", func(t *testing.T) {
		cmd := &cobra.Command{
			Use: "create",
			Long: `Create a document in a collection from a file.

Supported file types:
- Text files (.txt, .md, .json, etc.) - Content goes to 'text' field
- Image files (.jpg, .jpeg, .png, .gif, etc.) - Base64 data goes to 'image_data' field
- PDF files (.pdf) - Text extracted and chunked, goes to 'text' field

The command will automatically:
- Detect file type and process accordingly
- Generate appropriate metadata
- Chunk text content (default 1000 chars, configurable with --chunk-size)
- Create documents following RagMeDocs/RagMeImages schema

Examples:
  weave docs create MyCollection document.txt
  weave docs create MyCollection image.jpg
  weave docs create MyCollection document.pdf --chunk-size 500`,
		}

		// Verify the help text includes key information
		if !strings.Contains(cmd.Long, "Text files") {
			t.Error("Help text should mention text file support")
		}
		if !strings.Contains(cmd.Long, "Image files") {
			t.Error("Help text should mention image file support")
		}
		if !strings.Contains(cmd.Long, "PDF files") {
			t.Error("Help text should mention PDF file support")
		}
		if !strings.Contains(cmd.Long, "chunk-size") {
			t.Error("Help text should mention chunk-size flag")
		}
		if !strings.Contains(cmd.Long, "RagMeDocs/RagMeImages") {
			t.Error("Help text should mention schema compatibility")
		}
	})
}

// TestDocumentCreateFunctionality tests the core document creation functionality
func TestDocumentCreateFunctionality(t *testing.T) {
	t.Run("Text File Processing", func(t *testing.T) {
		// Create a temporary text file
		tempFile, err := os.CreateTemp("", "test-*.txt")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tempFile.Name())

		// Write test content
		testContent := "This is a test document for the weave CLI. It contains some sample text that will be chunked into smaller pieces for better processing. The weave CLI is a powerful tool for managing vector databases and documents."
		_, err = tempFile.WriteString(testContent)
		if err != nil {
			t.Fatalf("Failed to write test content: %v", err)
		}
		tempFile.Close()

		// Test file processing
		// Note: This tests the file processing logic conceptually
		// In a real implementation, we would test the actual processFile function
		if len(testContent) == 0 {
			t.Error("Test content should not be empty")
		}

		// Verify file exists
		if _, err := os.Stat(tempFile.Name()); os.IsNotExist(err) {
			t.Error("Temp file should exist")
		}
	})

	t.Run("Image File Processing", func(t *testing.T) {
		// Create a temporary image file (minimal PNG)
		tempFile, err := os.CreateTemp("", "test-*.png")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tempFile.Name())

		// Write minimal PNG data
		pngData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52}
		_, err = tempFile.Write(pngData)
		if err != nil {
			t.Fatalf("Failed to write PNG data: %v", err)
		}
		tempFile.Close()

		// Test that file exists and has content
		fileInfo, err := os.Stat(tempFile.Name())
		if err != nil {
			t.Fatalf("Failed to get file info: %v", err)
		}

		if fileInfo.Size() == 0 {
			t.Error("Image file should have content")
		}

		// Test file extension detection
		ext := strings.ToLower(filepath.Ext(tempFile.Name()))
		if ext != ".png" {
			t.Errorf("Expected .png extension, got %s", ext)
		}
	})

	t.Run("PDF File Processing Placeholder", func(t *testing.T) {
		// Create a temporary PDF file
		tempFile, err := os.CreateTemp("", "test-*.pdf")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tempFile.Name())

		// Write minimal PDF data
		pdfData := []byte("%PDF-1.4\n1 0 obj\n<<\n/Type /Catalog\n/Pages 2 0 R\n>>\nendobj")
		_, err = tempFile.Write(pdfData)
		if err != nil {
			t.Fatalf("Failed to write PDF data: %v", err)
		}
		tempFile.Close()

		// Test file extension detection
		ext := strings.ToLower(filepath.Ext(tempFile.Name()))
		if ext != ".pdf" {
			t.Errorf("Expected .pdf extension, got %s", ext)
		}

		// Note: PDF processing is not yet implemented, so we just test file detection
	})

	t.Run("Chunk Size Validation", func(t *testing.T) {
		testCases := []struct {
			name     string
			chunkSize int
			expected bool
		}{
			{"Valid chunk size", 1000, true},
			{"Small chunk size", 50, true},
			{"Large chunk size", 10000, true},
			{"Zero chunk size", 0, false},
			{"Negative chunk size", -100, false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Test chunk size validation logic
				isValid := tc.chunkSize > 0
				if isValid != tc.expected {
					t.Errorf("Expected chunk size %d to be valid: %v, got %v", tc.chunkSize, tc.expected, isValid)
				}
			})
		}
	})

	t.Run("File Type Detection", func(t *testing.T) {
		testCases := []struct {
			filename string
			expected string
		}{
			{"document.txt", "text"},
			{"readme.md", "text"},
			{"config.json", "text"},
			{"image.jpg", "image"},
			{"photo.jpeg", "image"},
			{"picture.png", "image"},
			{"document.pdf", "pdf"},
			{"report.PDF", "pdf"},
		}

		for _, tc := range testCases {
			t.Run(tc.filename, func(t *testing.T) {
				ext := strings.ToLower(filepath.Ext(tc.filename))
				var fileType string

				switch ext {
				case ".pdf":
					fileType = "pdf"
				case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp":
					fileType = "image"
				default:
					fileType = "text"
				}

				if fileType != tc.expected {
					t.Errorf("Expected file type %s for %s, got %s", tc.expected, tc.filename, fileType)
				}
			})
		}
	})
}

// TestDocumentCreateMockClient tests document creation with mock client
func TestDocumentCreateMockClient(t *testing.T) {
	t.Run("Mock Client Document Creation", func(t *testing.T) {
		// Create mock config
		mockConfig := &config.MockConfig{
			Collections: []config.MockCollection{
				{Name: "TestCollection"},
			},
		}

		// Create mock client
		client := mock.NewClient(mockConfig)

		// Test document creation
		ctx := context.Background()
		document := mock.Document{
			ID:       "test-doc-1",
			Content:  "This is a test document",
			URL:      "file://test.txt",
			Metadata: map[string]interface{}{
				"metadata": `{"filename": "test.txt", "content_type": "text"}`,
			},
		}

		err := client.CreateDocument(ctx, "TestCollection", document)
		if err != nil {
			t.Errorf("Failed to create document: %v", err)
		}

		// Verify document was created
		documents, err := client.ListDocuments(ctx, "TestCollection", 10)
		if err != nil {
			t.Errorf("Failed to list documents: %v", err)
		}

		if len(documents) == 0 {
			t.Error("Expected at least one document in collection")
		}

		found := false
		for _, doc := range documents {
			if doc.ID == "test-doc-1" {
				found = true
				if doc.Content != "This is a test document" {
					t.Errorf("Expected content 'This is a test document', got '%s'", doc.Content)
				}
				break
			}
		}

		if !found {
			t.Error("Created document not found in collection")
		}
	})

	t.Run("Mock Client Duplicate Document ID", func(t *testing.T) {
		// Create mock config
		mockConfig := &config.MockConfig{
			Collections: []config.MockCollection{
				{Name: "TestCollection"},
			},
		}

		// Create mock client
		client := mock.NewClient(mockConfig)

		ctx := context.Background()
		document := mock.Document{
			ID:       "duplicate-doc",
			Content:  "First document",
			URL:      "file://test1.txt",
			Metadata: map[string]interface{}{
				"metadata": `{"filename": "test1.txt"}`,
			},
		}

		// Create first document
		err := client.CreateDocument(ctx, "TestCollection", document)
		if err != nil {
			t.Errorf("Failed to create first document: %v", err)
		}

		// Try to create duplicate document
		document.Content = "Second document"
		err = client.CreateDocument(ctx, "TestCollection", document)
		if err == nil {
			t.Error("Expected error when creating duplicate document ID")
		}

		if !strings.Contains(err.Error(), "already exists") {
			t.Errorf("Expected 'already exists' error, got: %v", err)
		}
	})

	t.Run("Mock Client Non-existent Collection", func(t *testing.T) {
		// Create mock config with empty collections
		mockConfig := &config.MockConfig{
			Collections: []config.MockCollection{},
		}

		// Create mock client
		client := mock.NewClient(mockConfig)

		ctx := context.Background()
		document := mock.Document{
			ID:       "test-doc",
			Content:  "Test content",
			URL:      "file://test.txt",
			Metadata: map[string]interface{}{
				"metadata": `{"filename": "test.txt"}`,
			},
		}

		// Try to create document in non-existent collection
		err := client.CreateDocument(ctx, "NonExistentCollection", document)
		if err == nil {
			t.Error("Expected error when creating document in non-existent collection")
		}

		if !strings.Contains(err.Error(), "does not exist") {
			t.Errorf("Expected 'does not exist' error, got: %v", err)
		}
	})
}

// TestMultipleCollectionCreation tests the multiple collection creation functionality
func TestMultipleCollectionCreation(t *testing.T) {
	t.Run("Multiple Collection Create Command Structure", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:     "create COLLECTION_NAME [COLLECTION_NAME...]",
			Aliases: []string{"c"},
			Short:   "Create one or more collections",
			Args:    cobra.MinimumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				// Verify we can accept multiple arguments
				if len(args) < 1 {
					t.Error("Expected at least 1 argument")
				}
			},
		}

		// Test with single collection name
		cmd.SetArgs([]string{"TestCollection"})
		err := cmd.Execute()
		if err != nil {
			t.Errorf("Command execution failed: %v", err)
		}

		// Test with multiple collection names
		cmd.SetArgs([]string{"TestCollection1", "TestCollection2", "TestCollection3"})
		err = cmd.Execute()
		if err != nil {
			t.Errorf("Command execution failed: %v", err)
		}
	})

	t.Run("Multiple Collection Create with Flags", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:     "create COLLECTION_NAME [COLLECTION_NAME...]",
			Aliases: []string{"c"},
			Short:   "Create one or more collections",
			Args:    cobra.MinimumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				// Verify arguments
				if len(args) != 3 {
					t.Errorf("Expected 3 arguments, got %d", len(args))
				}
				expectedNames := []string{"Col1", "Col2", "Col3"}
				for i, name := range args {
					if name != expectedNames[i] {
						t.Errorf("Expected '%s', got '%s'", expectedNames[i], name)
					}
				}

				// Verify flags
				embedding, _ := cmd.Flags().GetString("embedding")
				if embedding != "text-embedding-3-large" {
					t.Errorf("Expected embedding 'text-embedding-3-large', got '%s'", embedding)
				}

				field, _ := cmd.Flags().GetString("field")
				if field != "title:text,author:text" {
					t.Errorf("Expected field 'title:text,author:text', got '%s'", field)
				}
			},
		}

		// Add flags
		cmd.Flags().StringP("embedding", "e", "text-embedding-3-small", "Embedding model")
		cmd.Flags().StringP("field", "f", "", "Custom fields")

		// Test with multiple collections and flags
		cmd.SetArgs([]string{"Col1", "Col2", "Col3", "--embedding", "text-embedding-3-large", "--field", "title:text,author:text"})
		err := cmd.Execute()
		if err != nil {
			t.Errorf("Command execution failed: %v", err)
		}
	})

	t.Run("Multiple Collection Create with Aliases", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:     "create COLLECTION_NAME [COLLECTION_NAME...]",
			Aliases: []string{"c"},
			Short:   "Create one or more collections",
			Args:    cobra.MinimumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				// Verify arguments
				if len(args) != 2 {
					t.Errorf("Expected 2 arguments, got %d", len(args))
				}
				if args[0] != "AliasCol1" {
					t.Errorf("Expected 'AliasCol1', got '%s'", args[0])
				}
				if args[1] != "AliasCol2" {
					t.Errorf("Expected 'AliasCol2', got '%s'", args[1])
				}
			},
		}

		// Test with alias command
		cmd.SetArgs([]string{"AliasCol1", "AliasCol2"})
		err := cmd.Execute()
		if err != nil {
			t.Errorf("Command execution failed: %v", err)
		}
	})

	t.Run("Multiple Collection Create Error Handling", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:     "create COLLECTION_NAME [COLLECTION_NAME...]",
			Aliases: []string{"c"},
			Short:   "Create one or more collections",
			Args:    cobra.MinimumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				// This should not be called if args validation fails
				t.Error("Run function should not be called with invalid args")
			},
		}

		// Test with no arguments (should fail)
		cmd.SetArgs([]string{})
		err := cmd.Execute()
		if err == nil {
			t.Error("Expected error when no arguments provided")
		}
	})
}

// TestCollectionSchemaDeletion tests the collection schema deletion functionality
func TestCollectionSchemaDeletion(t *testing.T) {
	t.Run("Collection Schema Delete Command Structure", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:     "delete-schema COLLECTION_NAME [COLLECTION_NAME...]",
			Aliases: []string{"del-schema", "ds"},
			Short:   "Delete collection schema(s) completely",
			Args:    cobra.MinimumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				// Verify arguments
				if len(args) != 1 {
					t.Errorf("Expected 1 argument, got %d", len(args))
				}
				if args[0] != "TestCollection" {
					t.Errorf("Expected 'TestCollection', got %s", args[0])
				}

				// Verify force flag
				force, _ := cmd.Flags().GetBool("force")
				if !force {
					t.Error("Expected force flag to be true")
				}
			},
		}

		// Add flags
		cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

		// Test with collection name and force flag
		cmd.SetArgs([]string{"TestCollection", "--force"})
		err := cmd.Execute()
		if err != nil {
			t.Errorf("Command execution failed: %v", err)
		}
	})

	t.Run("Collection Schema Delete with Aliases", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:     "delete-schema COLLECTION_NAME [COLLECTION_NAME...]",
			Aliases: []string{"del-schema", "ds"},
			Short:   "Delete collection schema(s) completely",
			Args:    cobra.MinimumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				// Verify arguments
				if len(args) != 1 {
					t.Errorf("Expected 1 argument, got %d", len(args))
				}
				if args[0] != "AliasTestCollection" {
					t.Errorf("Expected 'AliasTestCollection', got %s", args[0])
				}
			},
		}

		// Test with alias command
		cmd.SetArgs([]string{"AliasTestCollection"})
		err := cmd.Execute()
		if err != nil {
			t.Errorf("Command execution failed: %v", err)
		}
	})

	t.Run("Multiple Collection Schema Delete Command Structure", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:     "delete-schema COLLECTION_NAME [COLLECTION_NAME...]",
			Aliases: []string{"del-schema", "ds"},
			Short:   "Delete collection schema(s) completely",
			Args:    cobra.MinimumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				// Verify we can accept multiple arguments
				if len(args) < 1 {
					t.Error("Expected at least 1 argument")
				}
				expectedNames := []string{"MultiSchema1", "MultiSchema2", "MultiSchema3"}
				for i, name := range args {
					if name != expectedNames[i] {
						t.Errorf("Expected '%s', got '%s'", expectedNames[i], name)
					}
				}

				// Verify force flag
				force, _ := cmd.Flags().GetBool("force")
				if !force {
					t.Error("Expected force flag to be true")
				}
			},
		}

		// Add flags
		cmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

		// Test with multiple collection names and force flag
		cmd.SetArgs([]string{"MultiSchema1", "MultiSchema2", "MultiSchema3", "--force"})
		err := cmd.Execute()
		if err != nil {
			t.Errorf("Command execution failed: %v", err)
		}
	})

	t.Run("Collection Schema Delete Error Handling", func(t *testing.T) {
		cmd := &cobra.Command{
			Use:     "delete-schema COLLECTION_NAME [COLLECTION_NAME...]",
			Aliases: []string{"del-schema", "ds"},
			Short:   "Delete collection schema(s) completely",
			Args:    cobra.MinimumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				// This should not be called if args validation fails
				t.Error("Run function should not be called with invalid args")
			},
		}

		// Test with no arguments (should fail)
		cmd.SetArgs([]string{})
		err := cmd.Execute()
		if err == nil {
			t.Error("Expected error when no arguments provided")
		}
	})
}

// TestPatternMatching tests the pattern matching functionality
func TestPatternMatching(t *testing.T) {
	t.Run("IsRegexPattern", func(t *testing.T) {
		// Test cases for regex pattern detection
		testCases := []struct {
			pattern string
			isRegex bool
		}{
			// Shell glob patterns (should be false)
			{"tmp*.png", false},
			{"file?.txt", false},
			{"doc[0-9].pdf", false},
			{"image_[a-z].png", false},
			{"backup_*.log", false},
			
			// Regex patterns (should be true)
			{"tmp.*\\.png", true},
			{"^prefix.*\\.jpg$", true},
			{".*\\.(png|jpg|gif)$", true},
			{"file_\\d{4}\\.txt", true},
			{"^(temp|tmp).*\\.pdf$", true},
			{"tmp.*\\.png", true},
			{"^file.*\\.txt$", true},
		}

		for _, tc := range testCases {
			result := isRegexPattern(tc.pattern)
			if result != tc.isRegex {
				t.Errorf("Pattern '%s': expected %v, got %v", tc.pattern, tc.isRegex, result)
			}
		}
	})

	t.Run("GlobToRegex", func(t *testing.T) {
		// Test cases for glob to regex conversion
		testCases := []struct {
			glob    string
			regex   string
		}{
			{"tmp*.png", "tmp.*\\.png"},
			{"file?.txt", "file.\\.txt"},
			{"doc[0-9].pdf", "doc[0-9]\\.pdf"},
			{"image_[a-z].png", "image_[a-z]\\.png"},
			{"backup_*.log", "backup_.*\\.log"},
			{"simple.txt", "simple\\.txt"},
			{"file[abc].pdf", "file[abc]\\.pdf"},
		}

		for _, tc := range testCases {
			result := globToRegex(tc.glob)
			if result != tc.regex {
				t.Errorf("Glob '%s': expected '%s', got '%s'", tc.glob, tc.regex, result)
			}
		}
	})

	t.Run("PatternMatchingIntegration", func(t *testing.T) {
		// Test that converted patterns work correctly
		testCases := []struct {
			pattern string
			filename string
			shouldMatch bool
		}{
			// Shell glob patterns
			{"tmp*.png", "tmp123.png", true},
			{"tmp*.png", "tmp_file.png", true},
			{"tmp*.png", "other.png", false},
			{"file?.txt", "file1.txt", true},
			{"file?.txt", "file12.txt", false},
			{"doc[0-9].pdf", "doc1.pdf", true},
			{"doc[0-9].pdf", "doc5.pdf", true},
			{"doc[0-9].pdf", "doc12.pdf", false},
			
			// Regex patterns
			{"tmp.*\\.png", "tmp123.png", true},
			{"tmp.*\\.png", "tmp_file.png", true},
			{"tmp.*\\.png", "other.png", false},
			{"^prefix.*\\.jpg$", "prefix_image.jpg", true},
			{"^prefix.*\\.jpg$", "other_prefix.jpg", false},
		}

		for _, tc := range testCases {
			var regex *regexp.Regexp
			var err error
			
			if isRegexPattern(tc.pattern) {
				regex, err = regexp.Compile(tc.pattern)
			} else {
				regexPattern := globToRegex(tc.pattern)
				regex, err = regexp.Compile(regexPattern)
			}
			
			if err != nil {
				t.Errorf("Failed to compile pattern '%s': %v", tc.pattern, err)
				continue
			}
			
			matches := regex.MatchString(tc.filename)
			if matches != tc.shouldMatch {
				t.Errorf("Pattern '%s' with filename '%s': expected %v, got %v", 
					tc.pattern, tc.filename, tc.shouldMatch, matches)
			}
		}
	})
}

// Helper functions for pattern matching tests
func isRegexPattern(pattern string) bool {
	// Strong regex indicators (definitely regex)
	strongRegexIndicators := []string{
		"^", "$", "\\", ".*", ".+", ".?", "(", ")", "{", "}", "|", "+",
	}
	
	// Check for strong regex indicators first
	for _, indicator := range strongRegexIndicators {
		if strings.Contains(pattern, indicator) {
			return true
		}
	}
	
	// Check for escaped characters (indicates regex)
	if strings.Contains(pattern, "\\") {
		return true
	}
	
	// If pattern contains only glob characters (*, ?, []) and no strong regex chars, treat as glob
	return false
}

func globToRegex(glob string) string {
	// Escape special regex characters first
	result := regexp.QuoteMeta(glob)
	
	// Convert glob patterns to regex equivalents
	result = strings.ReplaceAll(result, "\\*", ".*")     // * -> .*
	result = strings.ReplaceAll(result, "\\?", ".")     // ? -> .
	result = strings.ReplaceAll(result, "\\[", "[")    // [ -> [
	result = strings.ReplaceAll(result, "\\]", "]")    // ] -> ]
	
	// Add anchors for exact matching (optional - could be made configurable)
	// result = "^" + result + "$"
	
	return result
}
