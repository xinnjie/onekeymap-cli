package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/cliconfig"
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

const (
	// Version information set by build system.
	version   = "dev"
	buildTime = "unknown"
	gitCommit = "unknown"
)

//nolint:gochecknoglobals // TODO(xinnjie): Stop using these global variables. But for now I can not think of a better way.
var (
	// Global shared state that needs to be accessed across commands
	pluginRegistry *plugins.Registry
	importService  importapi.Importer
	exportService  exportapi.Exporter
	logger         *slog.Logger
	recorder       metrics.Recorder
	mappingConfig  *mappings.MappingConfig
)

type rootFlags struct {
	verbose         bool
	quiet           bool
	logJSON         bool
	backup          bool
	interactive     bool
	enableTelemetry bool
}

func NewCmdRoot() *cobra.Command {
	f := rootFlags{}

	cmd := &cobra.Command{
		Use:              "onekeymap",
		Short:            "A tool to import, export, and synchronize keyboard shortcuts between editors.",
		Version:          fmt.Sprintf("%s (built %s, commit %s)", version, buildTime, gitCommit),
		PersistentPreRun: rootPersistentPreRun(&f),
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if recorder != nil {
				if err := recorder.Shutdown(context.Background()); err != nil {
					logger.Error("failed to shutdown telemetry", "error", err)
				}
			}
		},
	}

	// Define persistent flags
	cmd.PersistentFlags().BoolVarP(&f.verbose, "verbose", "v", false, "Enable verbose output")
	cmd.PersistentFlags().BoolVarP(&f.quiet, "quiet", "q", false, "Suppress all output except for errors")
	cmd.PersistentFlags().BoolVar(&f.logJSON, "log-json", false, "Output logs in JSON format")
	cmd.PersistentFlags().BoolVarP(&f.backup, "backup", "b", true, "Create a backup of the target editor's keymap")
	cmd.PersistentFlags().BoolVarP(&f.interactive, "interactive", "i", true, "Run in interactive mode")
	cmd.PersistentFlags().
		BoolVar(&f.enableTelemetry, "telemetry", false, "Enable OpenTelemetry to help improve onekeymap")

	// Bind cobra flags to viper.
	if err := viper.BindPFlag("verbose", cmd.PersistentFlags().Lookup("verbose")); err != nil {
		cmd.PrintErrf("Error binding verbose flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("quiet", cmd.PersistentFlags().Lookup("quiet")); err != nil {
		cmd.PrintErrf("Error binding quiet flag: %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("log-json", cmd.PersistentFlags().Lookup("log-json")); err != nil {
		cmd.PrintErrf("Error binding log-json flag: %v\n", err)
		os.Exit(1)
	}

	return cmd
}

func rootPersistentPreRun(f *rootFlags) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		// Initialize mapping config
		var err error
		mappingConfig, err = mappings.NewMappingConfig()
		if err != nil {
			cmd.PrintErrf("failed to initialize mapping config: %v\n", err)
			os.Exit(1)
		}

		recorder = metrics.NewNoop()
		if f.enableTelemetry {
			if viper.GetString("otel.exporter.otlp.endpoint") == "" {
				// Maybe print a warning that endpoint is not set
				fmt.Fprintln(
					os.Stderr,
					"Warning: --telemetry is enabled, but otel.exporter.otlp.endpoint is not set. Telemetry data will not be sent.",
				)
			}

			recorder, err = metrics.New(context.Background(), version, logger, mappingConfig)
			if err != nil {
				cmd.PrintErrf("failed to initialize telemetry: %v\n", err)
				os.Exit(1)
			}
		}

		verbose := viper.GetBool("verbose")
		quiet := viper.GetBool("quiet")

		// Set up logger based on the final configuration.
		var logLevel slog.Level
		switch {
		case verbose:
			logLevel = slog.LevelDebug
		case quiet:
			logLevel = slog.LevelError
		default:
			logLevel = slog.LevelWarn
		}

		var output *os.File
		if quiet {
			output = os.Stderr // Only errors go to stderr in quiet mode
		} else {
			output = os.Stdout
		}

		logJSON := viper.GetBool("log-json")
		var handler slog.Handler
		handlerOpts := &slog.HandlerOptions{
			Level: logLevel,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					return slog.Attr{}
				}
				return a
			},
		}

		if logJSON {
			handler = slog.NewJSONHandler(output, handlerOpts)
		} else {
			handler = slog.NewTextHandler(output, handlerOpts)
		}

		logger = slog.New(handler)

		// Initialize plugin registry and services
		pluginRegistry = plugins.NewRegistry()
		pluginRegistry.Register(vscode.New(mappingConfig, logger))
		pluginRegistry.Register(zed.New(mappingConfig, logger))
		pluginRegistry.Register(intellij.New(mappingConfig, logger))
		pluginRegistry.Register(helix.New(mappingConfig, logger))

		importService = onekeymap.NewImportService(pluginRegistry, mappingConfig, logger, recorder)
		exportService = onekeymap.NewExportService(pluginRegistry, mappingConfig, logger)
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd := NewCmdRoot()
	// Init viper config
	_, err := cliconfig.NewConfig(rootCmd)
	if err != nil {
		rootCmd.PrintErrf("Error initializing configuration: %v\n", err)
		os.Exit(1)
	}

	devCmd := NewCmdDev()
	rootCmd.AddCommand(devCmd)
	devCmd.AddCommand(NewCmdDevDocSupportActions())
	devCmd.AddCommand(NewCmdDevDoctor())
	devCmd.AddCommand(NewCmdDevListUnmappedActions())
	rootCmd.AddCommand(NewCmdView())
	rootCmd.AddCommand(NewCmdServe())
	rootCmd.AddCommand(NewCmdMigrate())
	rootCmd.AddCommand(NewCmdImport())
	rootCmd.AddCommand(NewCmdExport())
	err = rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
