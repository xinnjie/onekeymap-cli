package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xinnjie/onekeymap-cli/internal"
	"github.com/xinnjie/onekeymap-cli/internal/cliconfig"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/internal/plugins"
	"github.com/xinnjie/onekeymap-cli/internal/plugins/helix"
	"github.com/xinnjie/onekeymap-cli/internal/plugins/intellij"
	"github.com/xinnjie/onekeymap-cli/internal/plugins/vscode"
	"github.com/xinnjie/onekeymap-cli/internal/plugins/zed"
	"github.com/xinnjie/onekeymap-cli/internal/updatecheck"

	"github.com/xinnjie/onekeymap-cli/pkg/exportapi"
	"github.com/xinnjie/onekeymap-cli/pkg/importapi"
	"github.com/xinnjie/onekeymap-cli/pkg/metrics"
)

var (
	// NOTE: use go build -ldflags "-X github.com/xinnjie/onekeymap-cli/internal/cmd.version=$(git describe)"
	version = "dev"
	//nolint:gochecknoglobals // use go build -ldflags "-X github.com/xinnjie/onekeymap-cli/internal/cmd.commit=$(git describe)"
	commit = "unknown"
	//nolint:gochecknoglobals // use go build -ldflags "-X github.com/xinnjie/onekeymap-cli/internal/cmd.dirty=$(git describe)"
	dirty = "unknown"
)

//nolint:gochecknoglobals // TODO(xinnjie): Stop using these global variables. But for now I can not think of a better way.
var (
	// Global shared state that needs to be accessed across commands
	cmdPluginRegistry *plugins.Registry
	cmdImportService  importapi.Importer
	cmdExportService  exportapi.Exporter
	cmdLogger         *slog.Logger
	cmdRecorder       metrics.Recorder
	cmdMappingConfig  *mappings.MappingConfig
)

type rootFlags struct {
	verbose         bool
	quiet           bool
	logJSON         bool
	backup          bool
	interactive     bool
	enableTelemetry bool
	sandbox         bool
	skipUpdateCheck bool
}

func NewCmdRoot() *cobra.Command {
	f := rootFlags{}

	cmd := &cobra.Command{
		Use:              "onekeymap-cli",
		Short:            "A tool to import, export, and synchronize keyboard shortcuts between editors.",
		Version:          buildVersionString(),
		PersistentPreRun: rootPersistentPreRun(&f),
		PersistentPostRun: func(cmd *cobra.Command, _ []string) {
			if err := cmdRecorder.Shutdown(cmd.Context()); err != nil {
				cmdLogger.Error("failed to shutdown telemetry", "error", err)
			}
		},
	}

	cmd.PersistentFlags().BoolVarP(&f.verbose, "verbose", "v", false, "Enable verbose output")
	cmd.PersistentFlags().BoolVarP(&f.quiet, "quiet", "q", false, "Suppress all output except for errors")
	cmd.PersistentFlags().BoolVar(&f.logJSON, "log-json", false, "Output logs in JSON format")
	cmd.PersistentFlags().BoolVarP(&f.backup, "backup", "b", true, "Create a backup of the target editor's keymap")
	cmd.PersistentFlags().BoolVarP(&f.interactive, "interactive", "i", true, "Run in interactive mode")
	cmd.PersistentFlags().
		BoolVar(&f.enableTelemetry, "telemetry", false, "Enable OpenTelemetry to help improve onekeymap")
	cmd.PersistentFlags().
		BoolVar(&f.sandbox, "sandbox", false, "Enable sandbox mode for macOS, restricting file access")
	cmd.PersistentFlags().
		BoolVar(&f.skipUpdateCheck, "skip-update-check", false, "Skip checking for updates")

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
	if err := viper.BindPFlag("sandbox", cmd.PersistentFlags().Lookup("sandbox")); err != nil {
		cmd.PrintErrf("Error binding sandbox flag: %v\n", err)
		os.Exit(1)
	}

	return cmd
}

func rootPersistentPreRun(f *rootFlags) func(cmd *cobra.Command, _ []string) {
	return func(cmd *cobra.Command, _ []string) {
		_, err := cliconfig.NewConfig(cmd)
		if err != nil {
			cmd.PrintErrf("Error initializing configuration: %v\n", err)
			os.Exit(1)
		}

		verbose := viper.GetBool("verbose")
		quiet := viper.GetBool("quiet")
		logJSON := viper.GetBool("log-json")
		sandbox := viper.GetBool("sandbox")
		ctx := cmd.Context()

		cmdMappingConfig, err = mappings.NewMappingConfig()
		if err != nil {
			cmd.PrintErrf("failed to initialize mapping config: %v\n", err)
			os.Exit(1)
		}

		cmdRecorder = metrics.NewNoop()
		if f.enableTelemetry {
			if viper.GetString("otel.exporter.otlp.endpoint") == "" {
				cmd.PrintErrln(
					"Warning: --telemetry is enabled, but otel.exporter.otlp.endpoint is not set. Telemetry data will not be sent.",
				)
			}

			cmdRecorder, err = metrics.New(cmd.Context(), version, cmdLogger, cmdMappingConfig)
			if err != nil {
				cmd.PrintErrf("failed to initialize telemetry: %v\n", err)
				os.Exit(1)
			}
		}

		// Set up logger based on the final configuration.
		var logLevel = slog.LevelWarn
		switch {
		case verbose:
			logLevel = slog.LevelDebug
		case quiet:
			logLevel = slog.LevelError
		}

		var output = os.Stdout
		if quiet {
			output = os.Stderr
		}

		var handler slog.Handler
		handlerOpts := &slog.HandlerOptions{
			Level: logLevel,
			ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
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

		cmdLogger = slog.New(handler)

		cmdPluginRegistry = plugins.NewRegistry()

		// VSCode family
		cmdPluginRegistry.Register(vscode.New(cmdMappingConfig, cmdLogger))
		cmdPluginRegistry.Register(vscode.NewWindsurf(cmdMappingConfig, cmdLogger))
		cmdPluginRegistry.Register(vscode.NewWindsurfNext(cmdMappingConfig, cmdLogger))
		cmdPluginRegistry.Register(vscode.NewCursor(cmdMappingConfig, cmdLogger))

		// IntelliJ family
		cmdPluginRegistry.Register(intellij.New(cmdMappingConfig, cmdLogger))
		cmdPluginRegistry.Register(intellij.NewPycharm(cmdMappingConfig, cmdLogger))
		cmdPluginRegistry.Register(intellij.NewIntelliJCommunity(cmdMappingConfig, cmdLogger))
		cmdPluginRegistry.Register(intellij.NewWebStorm(cmdMappingConfig, cmdLogger))
		cmdPluginRegistry.Register(intellij.NewClion(cmdMappingConfig, cmdLogger))
		cmdPluginRegistry.Register(intellij.NewPhpStorm(cmdMappingConfig, cmdLogger))
		cmdPluginRegistry.Register(intellij.NewRubyMine(cmdMappingConfig, cmdLogger))
		cmdPluginRegistry.Register(intellij.NewGoLand(cmdMappingConfig, cmdLogger))

		cmdPluginRegistry.Register(helix.New(cmdMappingConfig, cmdLogger))
		cmdPluginRegistry.Register(zed.New(cmdMappingConfig, cmdLogger))

		cmdImportService = internal.NewImportService(cmdPluginRegistry, cmdMappingConfig, cmdLogger, cmdRecorder)
		cmdExportService = internal.NewExportService(cmdPluginRegistry, cmdMappingConfig, cmdLogger)

		// Check for updates asynchronously
		if !sandbox && !f.skipUpdateCheck {
			checker := updatecheck.New(version, cmdLogger)
			checker.CheckForUpdate(ctx)
		}
	}
}

func Execute() {
	rootCmd := NewCmdRoot()

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
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func buildVersionString() string {
	commitID := commit
	if commitID == "" {
		commitID = "unknown"
	}

	status := dirty
	if status == "" {
		status = "unknown"
	}

	if version == "dev" {
		return fmt.Sprintf("%s (commit: %s, dirty: %s)", version, commitID, status)
	}

	return version
}
