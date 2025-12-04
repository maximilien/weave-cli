// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	configcmd "github.com/maximilien/weave-cli/src/cmd/config"
	"github.com/maximilien/weave-cli/src/pkg/config"
	"github.com/maximilien/weave-cli/src/pkg/repl"
	"github.com/maximilien/weave-cli/src/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile        string
	envFile        string
	noColor        bool
	noTruncate     bool
	noTips         bool
	jsonOutput     bool
	quietConfig    bool
	vectorDBType   string
	weaviateAPIKey string
	weaviateURL    string
	timeout        string
	noConfirm      bool
	queryStrings   string

	// Vector database type flags
	useWeaviate      bool
	useWeaviateLocal bool
	useWeaviateCloud bool
	useSupabase      bool
	useSupabaseLocal bool
	useSupabaseCloud bool
	useMongoDB       bool
	useMongoDBLocal  bool
	useMongoDBCloud  bool
	useMilvusLocal   bool
	useMilvusCloud   bool
	useChromaLocal   bool
	useChromaCloud   bool
	useQdrantLocal   bool
	useQdrantCloud   bool
	useNeo4jLocal    bool
	useNeo4jCloud    bool
	useMock          bool
	useAll           bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:                        "weave",
	Short:                      "Weave Vector Database Management Tool",
	SuggestionsMinimumDistance: 2,
	Run:                        runREPL,
	Long: `Weave is a command-line tool for managing vector databases.
Supports Weaviate (cloud/local), Milvus (local/cloud), MongoDB Atlas, Supabase PGVector, Chroma (local/cloud), Qdrant (local/cloud), Neo4j (local/cloud), and Mock databases.

📁 COLLECTION MANAGEMENT:
  weave cols ls                        # List all collections
  weave cols create COLLECTION         # Create a collection
  weave cols show COLLECTION           # Show collection details
  weave cols count                     # Count collections
  weave cols del COLLECTION            # Delete all documents in collection
  weave cols da                        # Delete all documents in all collections
  weave cols ds COLLECTION             # Delete collection schema (⚠️ destructive)

📄 DOCUMENT MANAGEMENT:
  weave docs ls COLLECTION             # List documents in collection
  weave docs show COLLECTION ID        # Show specific document
  weave docs show COLLECTION --name FILE # Show document by filename
  weave docs create COLLECTION FILE    # Create document from file
  weave docs count COLLECTION          # Count documents in collection
  weave docs del COLLECTION [ID...]    # Delete specific documents
  weave docs del COLLECTION --name FILE # Delete document by filename
  weave docs da COLLECTION             # Delete all documents in collection

⚙️ CONFIGURATION & HEALTH:
  weave config create --env            # Create new .env file interactively
  weave config update --env            # Update existing .env file
  weave config sync                    # Sync local config to ~/.weave-cli
  weave config update --weave-mcp      # Install weave-mcp binary for REPL
  weave config show                    # Show current configuration
  weave config list                    # List all configured databases
  weave health check                   # Check database health

📊 EMBEDDINGS:
  weave embeddings list                # List all available embedding models
  weave emb ls                         # Same as above (alias)
  weave embeddings list COLLECTION     # Show embeddings for specific collection

🔧 FILE SUPPORT:
  - Text files (.txt, .md, .json) → chunked text content
  - Image files (.jpg, .png, .gif) → base64 image data
  - PDF files (.pdf) → extracted text + images

🗄️ DATABASE SELECTION:
  --weaviate                           # Use Weaviate databases only
  --weaviate-local                     # Use Weaviate local database only
  --weaviate-cloud                     # Use Weaviate cloud database only
  --supabase                           # Use Supabase database only
  --supabase-local                     # Use Supabase local database only
  --supabase-cloud                     # Use Supabase cloud database only
  --mongodb                            # Use MongoDB database only
  --mongodb-local                      # Use MongoDB local database only
  --mongodb-cloud                      # Use MongoDB cloud (Atlas) database only
  --milvus-local                       # Use Milvus local database only
  --milvus-cloud                       # Use Milvus cloud (Zilliz) database only
  --chroma-local                       # Use Chroma local database only
  --chroma-cloud                       # Use Chroma cloud database only
  --qdrant-local                       # Use Qdrant local database only
  --qdrant-cloud                       # Use Qdrant cloud database only
  --neo4j-local                        # Use Neo4j local database only
  --neo4j-cloud                        # Use Neo4j cloud (Aura) database only
  --mock                               # Use mock database only
  --all                                # Use all configured databases (default)

The tool uses ./config.yaml and ./.env files by default, or you can specify
custom locations with --config and --env flags.

Priority order: command flags > --env file > .env file > shell environment.`,
	Version: version.Get().Version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig, initColor)

	// Define command groups for better organization
	rootCmd.AddGroup(&cobra.Group{ID: "data", Title: "Data Management:"})
	rootCmd.AddGroup(&cobra.Group{ID: "config", Title: "Configuration & Health:"})
	rootCmd.AddGroup(&cobra.Group{ID: "ai", Title: "AI & Search:"})

	// Register command packages
	rootCmd.AddCommand(configcmd.Cmd)
	configcmd.Cmd.GroupID = "config"

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")
	rootCmd.PersistentFlags().StringVar(&envFile, "env", "", "env file (default is ./.env)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "quiet output (minimal messages)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().BoolVar(&noTruncate, "no-truncate", false, "show all data without truncation")
	rootCmd.PersistentFlags().BoolVar(&noTips, "no-tips", false, "suppress helpful tips and suggestions")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output results in JSON format")
	rootCmd.PersistentFlags().BoolVar(&quietConfig, "quiet-config", false, "suppress config location information")
	rootCmd.PersistentFlags().BoolVar(&noConfirm, "no-confirm", false, "skip confirmation prompts for destructive operations")

	// REPL-specific flags
	rootCmd.Flags().StringVar(&queryStrings, "query-strings", "", "file with queries to execute (one per line, batch mode)")

	// Environment variable override flags (highest priority)
	rootCmd.PersistentFlags().StringVarP(&vectorDBType, "vector-db-type", "", "", "override VECTOR_DB_TYPE (weaviate-cloud|weaviate-local|milvus-local|milvus-cloud|mongodb|supabase|mock)")
	rootCmd.PersistentFlags().StringVar(&vectorDBType, "vdb", "", "alias for --vector-db-type")
	rootCmd.PersistentFlags().StringVar(&weaviateAPIKey, "weaviate-api-key", "", "override WEAVIATE_API_KEY")
	rootCmd.PersistentFlags().StringVar(&weaviateURL, "weaviate-url", "", "override WEAVIATE_URL")
	rootCmd.PersistentFlags().StringVar(&timeout, "timeout", "", "timeout for vector DB operations (e.g., 5s, 10s, 30s; default: 10s)")

	// Vector database type selection flags
	rootCmd.PersistentFlags().BoolVar(&useWeaviate, "weaviate", false, "use Weaviate vector database (weaviate-cloud or weaviate-local)")
	rootCmd.PersistentFlags().BoolVar(&useWeaviateLocal, "weaviate-local", false, "use Weaviate local vector database")
	rootCmd.PersistentFlags().BoolVar(&useWeaviateCloud, "weaviate-cloud", false, "use Weaviate cloud vector database")
	rootCmd.PersistentFlags().BoolVar(&useSupabase, "supabase", false, "use Supabase PGVector database (supabase-cloud or supabase-local)")
	rootCmd.PersistentFlags().BoolVar(&useSupabaseLocal, "supabase-local", false, "use Supabase local vector database")
	rootCmd.PersistentFlags().BoolVar(&useSupabaseCloud, "supabase-cloud", false, "use Supabase cloud vector database")
	rootCmd.PersistentFlags().BoolVar(&useMongoDB, "mongodb", false, "use MongoDB Atlas Vector Search database (mongodb-cloud or mongodb-local)")
	rootCmd.PersistentFlags().BoolVar(&useMongoDBLocal, "mongodb-local", false, "use MongoDB local vector database")
	rootCmd.PersistentFlags().BoolVar(&useMongoDBCloud, "mongodb-cloud", false, "use MongoDB cloud (Atlas) vector database")
	rootCmd.PersistentFlags().BoolVar(&useMilvusLocal, "milvus-local", false, "use Milvus local vector database")
	rootCmd.PersistentFlags().BoolVar(&useMilvusCloud, "milvus-cloud", false, "use Milvus cloud (Zilliz) vector database")
	rootCmd.PersistentFlags().BoolVar(&useChromaLocal, "chroma-local", false, "use Chroma local vector database")
	rootCmd.PersistentFlags().BoolVar(&useChromaCloud, "chroma-cloud", false, "use Chroma cloud vector database")
	rootCmd.PersistentFlags().BoolVar(&useQdrantLocal, "qdrant-local", false, "use Qdrant local vector database")
	rootCmd.PersistentFlags().BoolVar(&useQdrantCloud, "qdrant-cloud", false, "use Qdrant cloud vector database")
	rootCmd.PersistentFlags().BoolVar(&useNeo4jLocal, "neo4j-local", false, "use Neo4j local vector database")
	rootCmd.PersistentFlags().BoolVar(&useNeo4jCloud, "neo4j-cloud", false, "use Neo4j cloud (Aura) vector database")
	rootCmd.PersistentFlags().BoolVar(&useMock, "mock", false, "use mock vector database")
	rootCmd.PersistentFlags().BoolVar(&useAll, "all", false, "operate on all configured vector databases")

	// Set custom usage template with grouped flags
	rootCmd.SetUsageTemplate(getGroupedUsageTemplate())

	// Bind flags to viper
	_ = viper.BindPFlag("vector-db-type", rootCmd.PersistentFlags().Lookup("vector-db-type"))
	_ = viper.BindPFlag("vdb", rootCmd.PersistentFlags().Lookup("vdb"))
	_ = viper.BindPFlag("weaviate-api-key", rootCmd.PersistentFlags().Lookup("weaviate-api-key"))
	_ = viper.BindPFlag("weaviate-url", rootCmd.PersistentFlags().Lookup("weaviate-url"))
	_ = viper.BindPFlag("timeout", rootCmd.PersistentFlags().Lookup("timeout"))
	_ = viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	_ = viper.BindPFlag("env", rootCmd.PersistentFlags().Lookup("env"))
	_ = viper.BindPFlag("no-tips", rootCmd.PersistentFlags().Lookup("no-tips"))
	_ = viper.BindPFlag("json", rootCmd.PersistentFlags().Lookup("json"))
	_ = viper.BindPFlag("quiet-config", rootCmd.PersistentFlags().Lookup("quiet-config"))
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	_ = viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
	_ = viper.BindPFlag("no-color", rootCmd.PersistentFlags().Lookup("no-color"))
	_ = viper.BindPFlag("no-confirm", rootCmd.PersistentFlags().Lookup("no-confirm"))

	// Bind vector database type flags
	_ = viper.BindPFlag("weaviate", rootCmd.PersistentFlags().Lookup("weaviate"))
	_ = viper.BindPFlag("weaviate-local", rootCmd.PersistentFlags().Lookup("weaviate-local"))
	_ = viper.BindPFlag("weaviate-cloud", rootCmd.PersistentFlags().Lookup("weaviate-cloud"))
	_ = viper.BindPFlag("supabase", rootCmd.PersistentFlags().Lookup("supabase"))
	_ = viper.BindPFlag("supabase-local", rootCmd.PersistentFlags().Lookup("supabase-local"))
	_ = viper.BindPFlag("supabase-cloud", rootCmd.PersistentFlags().Lookup("supabase-cloud"))
	_ = viper.BindPFlag("mongodb", rootCmd.PersistentFlags().Lookup("mongodb"))
	_ = viper.BindPFlag("mongodb-local", rootCmd.PersistentFlags().Lookup("mongodb-local"))
	_ = viper.BindPFlag("mongodb-cloud", rootCmd.PersistentFlags().Lookup("mongodb-cloud"))
	_ = viper.BindPFlag("milvus-local", rootCmd.PersistentFlags().Lookup("milvus-local"))
	_ = viper.BindPFlag("milvus-cloud", rootCmd.PersistentFlags().Lookup("milvus-cloud"))
	_ = viper.BindPFlag("chroma-local", rootCmd.PersistentFlags().Lookup("chroma-local"))
	_ = viper.BindPFlag("chroma-cloud", rootCmd.PersistentFlags().Lookup("chroma-cloud"))
	_ = viper.BindPFlag("qdrant-local", rootCmd.PersistentFlags().Lookup("qdrant-local"))
	_ = viper.BindPFlag("qdrant-cloud", rootCmd.PersistentFlags().Lookup("qdrant-cloud"))
	_ = viper.BindPFlag("neo4j-local", rootCmd.PersistentFlags().Lookup("neo4j-local"))
	_ = viper.BindPFlag("neo4j-cloud", rootCmd.PersistentFlags().Lookup("neo4j-cloud"))
	_ = viper.BindPFlag("mock", rootCmd.PersistentFlags().Lookup("mock"))
	_ = viper.BindPFlag("all", rootCmd.PersistentFlags().Lookup("all"))

	// Add version flag with custom handler
	rootCmd.Flags().BoolP("version", "V", false, "show version information")

	// Override the default version template
	rootCmd.SetVersionTemplate(version.String())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Look for config.yaml in current directory
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		if viper.GetBool("verbose") {
			fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
		}
	} else {
		if viper.GetBool("verbose") {
			fmt.Fprintf(os.Stderr, "Warning: Could not read config file: %v\n", err)
		}
	}
}

// LoadConfigWithOverrides loads configuration with command-line overrides
func LoadConfigWithOverrides() (*config.Config, error) {
	return config.LoadConfigWithOptions(config.LoadConfigOptions{
		ConfigFile:     cfgFile,
		EnvFile:        envFile,
		VectorDBType:   vectorDBType,
		WeaviateAPIKey: weaviateAPIKey,
		WeaviateURL:    weaviateURL,
		Timeout:        timeout,
	})
}

// printHeader prints a colored header message
func printHeader(message string) {
	color.New(color.FgBlue, color.Bold).Printf("🔧 %s\n", message)
}

// printSuccess prints a colored success message
func printSuccess(message string) {
	color.New(color.FgGreen).Printf("✅ %s\n", message)
}

// printWarning prints a colored warning message
func printWarning(message string) {
	color.New(color.FgYellow).Printf("⚠️  %s\n", message)
}

// printError prints a colored error message
func printError(message string) {
	color.New(color.FgRed).Printf("❌ %s\n", message)
}

// initColor initializes color settings based on the no-color flag
func initColor() {
	if noColor {
		color.NoColor = true
	}
}

// ShouldShowTips returns true if tips should be displayed to the user
func ShouldShowTips() bool {
	return !noTips
}

// getGroupedUsageTemplate returns a custom usage template with logically grouped flags
func getGroupedUsageTemplate() string {
	return `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

Available Commands:{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

Additional Commands:{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}{{if eq .Name "weave"}}

Flags:

  Database Selection:
      --weaviate                  use Weaviate vector database (weaviate-cloud or weaviate-local)
      --weaviate-local            use Weaviate local vector database
      --weaviate-cloud            use Weaviate cloud vector database
      --supabase                  use Supabase PGVector database (supabase-cloud or supabase-local)
      --supabase-local            use Supabase local vector database
      --supabase-cloud            use Supabase cloud vector database
      --mongodb                   use MongoDB Atlas Vector Search database (mongodb-cloud or mongodb-local)
      --mongodb-local             use MongoDB local vector database
      --mongodb-cloud             use MongoDB cloud (Atlas) vector database
      --milvus-local              use Milvus local vector database
      --milvus-cloud              use Milvus cloud (Zilliz) vector database
      --chroma-local              use Chroma local vector database
      --chroma-cloud              use Chroma cloud vector database
      --qdrant-local              use Qdrant local vector database
      --qdrant-cloud              use Qdrant cloud vector database
      --neo4j-local               use Neo4j local vector database
      --neo4j-cloud               use Neo4j cloud (Aura) vector database
      --mock                      use mock vector database
      --all                       operate on all configured vector databases

  Database Override:
      --vector-db-type string     override VECTOR_DB_TYPE (weaviate-cloud|weaviate-local|milvus-local|milvus-cloud|mongodb|supabase|mock)
      --vdb string                alias for --vector-db-type
      --weaviate-api-key string   override WEAVIATE_API_KEY
      --weaviate-url string       override WEAVIATE_URL
      --timeout string            timeout for vector DB operations (e.g., 5s, 10s, 30s; default: 10s)

  Configuration:
      --config string             config file (default is ./config.yaml)
      --env string                env file (default is ./.env)

  Output Control:
  -v, --verbose                   verbose output
  -q, --quiet                     quiet output (minimal messages)
      --json                      output results in JSON format
      --no-color                  disable colored output
      --no-truncate               show all data without truncation
      --no-tips                   suppress helpful tips and suggestions
      --quiet-config              suppress config location information

  Behavior:
      --no-confirm                skip confirmation prompts for destructive operations

  REPL:
      --query-strings string      file with queries to execute (one per line, batch mode)

  Other:
  -h, --help                      help for weave
  -V, --version                   show version information
{{else}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:

  Database Selection:
      --weaviate                  use Weaviate vector database (weaviate-cloud or weaviate-local)
      --weaviate-local            use Weaviate local vector database
      --weaviate-cloud            use Weaviate cloud vector database
      --supabase                  use Supabase PGVector database (supabase-cloud or supabase-local)
      --supabase-local            use Supabase local vector database
      --supabase-cloud            use Supabase cloud vector database
      --mongodb                   use MongoDB Atlas Vector Search database (mongodb-cloud or mongodb-local)
      --mongodb-local             use MongoDB local vector database
      --mongodb-cloud             use MongoDB cloud (Atlas) vector database
      --milvus-local              use Milvus local vector database
      --milvus-cloud              use Milvus cloud (Zilliz) vector database
      --chroma-local              use Chroma local vector database
      --chroma-cloud              use Chroma cloud vector database
      --qdrant-local              use Qdrant local vector database
      --qdrant-cloud              use Qdrant cloud vector database
      --neo4j-local               use Neo4j local vector database
      --neo4j-cloud               use Neo4j cloud (Aura) vector database
      --mock                      use mock vector database
      --all                       operate on all configured vector databases

  Database Override:
      --vector-db-type string     override VECTOR_DB_TYPE (weaviate-cloud|weaviate-local|milvus-local|milvus-cloud|mongodb|supabase|mock)
      --vdb string                alias for --vector-db-type
      --weaviate-api-key string   override WEAVIATE_API_KEY
      --weaviate-url string       override WEAVIATE_URL
      --timeout string            timeout for vector DB operations (e.g., 5s, 10s, 30s; default: 10s)

  Configuration:
      --config string             config file (default is ./config.yaml)
      --env string                env file (default is ./.env)

  Output Control:
  -v, --verbose                   verbose output
  -q, --quiet                     quiet output (minimal messages)
      --json                      output results in JSON format
      --no-color                  disable colored output
      --no-truncate               show all data without truncation
      --no-tips                   suppress helpful tips and suggestions
      --quiet-config              suppress config location information

  Behavior:
      --no-confirm                skip confirmation prompts for destructive operations
{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
}

// IsJSONOutput returns true if JSON output is requested
func IsJSONOutput() bool {
	return jsonOutput
}

// runREPL starts the interactive REPL mode
func runREPL(cmd *cobra.Command, args []string) {
	// Create REPL with options
	opts := repl.Options{
		QueryStringsFile: queryStrings,
		NoConfirm:        noConfirm,
	}

	r, err := repl.NewWithOptions(opts)
	if err != nil {
		// Check if this is a configuration error
		formattedErr := config.FormatConfigError(err)
		if formattedErr != err.Error() {
			// Error was enhanced with configuration tips
			printError(formattedErr)

			// Check if we can prompt for fix - use REPL-specific check
			configErr := config.CheckREPLRequiredEnvVars()
			if configErr != nil {
				// For WEAVE_MCP_STDIO_PATH, we can't fix it interactively (it's a path to a binary)
				// So only prompt if other vars are missing
				hasMCPPath := true
				for _, v := range configErr.MissingVars {
					if v == "WEAVE_MCP_STDIO_PATH" {
						hasMCPPath = false
						break
					}
				}

				// Only prompt for interactive fix if we're not just missing MCP path
				if len(configErr.MissingVars) > 0 && (len(configErr.MissingVars) > 1 || hasMCPPath) {
					shouldFix, promptErr := config.PromptToFixConfig(configErr)
					if promptErr == nil && shouldFix {
						if fixErr := config.InteractiveConfigFix(configErr.EnvFileExists); fixErr != nil {
							printError(fmt.Sprintf("Failed to fix configuration: %v", fixErr))
							os.Exit(1)
						}

						// If WEAVE_MCP_STDIO_PATH is still missing, inform user
						if !hasMCPPath {
							printWarning("\nNote: WEAVE_MCP_STDIO_PATH still needs to be set manually.")
							printWarning("Please install weave-mcp and set the path to the binary.")
							os.Exit(1)
						}

						// After fixing, try creating REPL again
						r, err = repl.NewWithOptions(opts)
						if err != nil {
							printError(fmt.Sprintf("Failed to start REPL after configuration fix: %v", err))
							os.Exit(1)
						}
						// Continue to Run() below
					} else {
						os.Exit(1)
					}
				} else {
					// Only WEAVE_MCP_STDIO_PATH is missing, can't fix interactively
					os.Exit(1)
				}
			} else {
				os.Exit(1)
			}
		} else {
			printError(fmt.Sprintf("Failed to start REPL: %v", err))
			os.Exit(1)
		}
	}

	if err := r.Run(); err != nil {
		printError(fmt.Sprintf("REPL error: %v", err))
		os.Exit(1)
	}
}
