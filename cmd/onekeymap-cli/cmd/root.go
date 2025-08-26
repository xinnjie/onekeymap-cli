package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/cli"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/plugins"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/plugins/helix"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/plugins/intellij"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/plugins/vscode"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/plugins/zed"

	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/exportapi"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/importapi"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/metrics"
)

var (
	// Version information set by build system
	version   = "dev"
	buildTime = "unknown"
	gitCommit = "unknown"

	pluginRegistry  *plugins.Registry
	importService   importapi.Importer
	exportService   exportapi.Exporter
	logger          *slog.Logger
	interactive     *bool
	backup          *bool
	enableTelemetry *bool
	recorder        metrics.Recorder
	mappingConfig   *mappings.MappingConfig
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "onekeymap",
	Short:   "A tool to import, export, and synchronize keyboard shortcuts between editors.",
	Version: fmt.Sprintf("%s (built %s, commit %s)", version, buildTime, gitCommit),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize mapping config
		var err error
		mappingConfig, err = mappings.NewMappingConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to initialize mapping config: %v\n", err)
			os.Exit(1)
		}

		recorder = metrics.NewNoop()
		if *enableTelemetry {
			if viper.GetString("otel.exporter.otlp.endpoint") == "" {
				// Maybe print a warning that endpoint is not set
				fmt.Fprintln(os.Stderr, "Warning: --telemetry is enabled, but otel.exporter.otlp.endpoint is not set. Telemetry data will not be sent.")
			}

			recorder, err = metrics.New(context.Background(), version, logger, mappingConfig)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to initialize telemetry: %v\n", err)
				os.Exit(1)
			}
		}

		verbose := viper.GetBool("verbose")
		quiet := viper.GetBool("quiet")

		// Set up logger based on the final configuration.
		var logLevel slog.Level
		if verbose {
			logLevel = slog.LevelDebug
		} else if quiet {
			logLevel = slog.LevelError
		} else {
			logLevel = slog.LevelWarn
		}

		var output *os.File
		if quiet {
			output = os.Stderr // Only errors go to stderr in quiet mode
		} else {
			output = os.Stdout
		}

		logger = slog.New(slog.NewTextHandler(output, &slog.HandlerOptions{
			Level: logLevel,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					return slog.Attr{}
				}
				return a
			},
		}))

		// Initialize plugin registry and services
		pluginRegistry = plugins.NewRegistry()
		pluginRegistry.Register(vscode.New(mappingConfig, logger))
		pluginRegistry.Register(zed.New(mappingConfig, logger))
		pluginRegistry.Register(intellij.New(mappingConfig, logger))
		pluginRegistry.Register(helix.New(mappingConfig, logger))

		importService = onekeymap.NewImportService(pluginRegistry, mappingConfig, logger, recorder)
		exportService = onekeymap.NewExportService(pluginRegistry, mappingConfig, logger)
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if recorder != nil {
			if err := recorder.Shutdown(context.Background()); err != nil {
				logger.Error("failed to shutdown telemetry", "error", err)
			}
		}
	},
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
	// Define persistent flags for verbose and quiet modes.
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Suppress all output except for errors")
	backup = rootCmd.PersistentFlags().BoolP("backup", "b", true, "Create a backup of the target editor's keymap")
	interactive = rootCmd.PersistentFlags().BoolP("interactive", "i", true, "Run in interactive mode")
	enableTelemetry = rootCmd.PersistentFlags().Bool("telemetry", false, "Enable OpenTelemetry to help improve onekeymap")

	// Bind cobra flags to viper.
	if err := viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose")); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding verbose flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet")); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding quiet flag: %v\n", err)
		os.Exit(1)
	}

	// Init viper config
	var err error
	_, err = cli.NewConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing configuration: %v\n", err)
		os.Exit(1)
	}
}
