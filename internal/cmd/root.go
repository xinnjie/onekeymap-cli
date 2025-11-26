package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xinnjie/onekeymap-cli/internal/cliconfig"
	"github.com/xinnjie/onekeymap-cli/internal/updatecheck"
	"github.com/xinnjie/onekeymap-cli/pkg/api/exporterapi"
	"github.com/xinnjie/onekeymap-cli/pkg/api/importerapi"
	"github.com/xinnjie/onekeymap-cli/pkg/exporter"
	"github.com/xinnjie/onekeymap-cli/pkg/importer"
	"github.com/xinnjie/onekeymap-cli/pkg/mappings"
	"github.com/xinnjie/onekeymap-cli/pkg/metrics"
	"github.com/xinnjie/onekeymap-cli/pkg/registry"
	"golang.org/x/term"
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
	cmdPluginRegistry *registry.Registry
	cmdImportService  importerapi.Importer
	cmdExportService  exporterapi.Exporter
	cmdLogger         *slog.Logger
	cmdRecorder       metrics.Recorder
	cmdMappingConfig  *mappings.MappingConfig
	cmdUpdateMsgChan  <-chan string // Channel for async update check result
)

type rootFlags struct {
	verbose         bool
	quiet           bool
	logJSON         bool
	backup          bool
	interactive     bool
	sandbox         bool
	skipUpdateCheck bool
	onekeymap       string
}

func newCmdRoot() (*cobra.Command, *rootFlags) {
	f := rootFlags{}

	cmd := &cobra.Command{
		Use:   "onekeymap-cli",
		Short: "A tool to import, export, and synchronize keyboard shortcuts between editors.",
		Long: "A tool to import, export, and synchronize keyboard shortcuts between editors.\n" +
			"If you encounter any issues, please open an issue at https://github.com/xinnjie/onekeymap-cli.",
		Version:          buildVersionString(),
		PersistentPreRun: rootPersistentPreRun(&f),
		PersistentPostRun: rootPersistentPostRun(
			&f,
			func() (*slog.Logger, metrics.Recorder) { return cmdLogger, cmdRecorder },
		),
	}

	cmd.PersistentFlags().BoolVarP(&f.verbose, "verbose", "v", false, "Enable verbose output")
	cmd.PersistentFlags().BoolVarP(&f.quiet, "quiet", "q", false, "Suppress all output except for errors")
	cmd.PersistentFlags().BoolVar(&f.logJSON, "log-json", false, "Output logs in JSON format")
	cmd.PersistentFlags().BoolVarP(&f.backup, "backup", "b", true, "Create a backup of the target editor's keymap")
	cmd.PersistentFlags().BoolVarP(&f.interactive, "interactive", "i", true, "Run in interactive mode")
	cmd.PersistentFlags().
		BoolVar(&f.sandbox, "sandbox", false, "Enable sandbox mode for macOS, restricting file access")
	cmd.PersistentFlags().
		BoolVar(&f.skipUpdateCheck, "skip-update-check", false, "Skip checking for updates")
	cmd.PersistentFlags().
		StringVar(&f.onekeymap, "onekeymap", "", "Path to onekeymap.json file (overrides config file setting)")

	if err := viper.BindPFlag("onekeymap", cmd.PersistentFlags().Lookup("onekeymap")); err != nil {
		cmd.PrintErrf("Error binding onekeymap flag: %v\n", err)
		os.Exit(1)
	}

	return cmd, &f
}

func rootPersistentPreRun(f *rootFlags) func(cmd *cobra.Command, _ []string) {
	return func(cmd *cobra.Command, _ []string) {
		_, err := cliconfig.NewConfig(f.sandbox)
		if err != nil {
			cmd.PrintErrf("Error initializing configuration: %v\n", err)
			os.Exit(1)
		}

		isTTY := term.IsTerminal(int(os.Stdin.Fd()))
		if !isTTY {
			f.interactive = false
		}

		verbose := f.verbose
		quiet := f.quiet
		logJSON := f.logJSON
		cmdRecorder = metrics.NewNoop()

		cmdMappingConfig, err = mappings.NewMappingConfig()
		if err != nil {
			cmd.PrintErrf("failed to initialize mapping config: %v\n", err)
			os.Exit(1)
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

		cmdPluginRegistry = registry.NewRegistryWithPlugins(cmdMappingConfig, cmdLogger, cmdRecorder)

		cmdImportService = importer.NewImporter(cmdPluginRegistry, cmdMappingConfig, cmdLogger, cmdRecorder)
		cmdExportService = exporter.NewExporter(cmdPluginRegistry, cmdMappingConfig, cmdLogger, cmdRecorder)

		// Start async update check only in interactive mode and when not in sandbox
		if f.interactive && !f.sandbox && !f.skipUpdateCheck {
			checker := updatecheck.New(version, cmdLogger)
			cmdUpdateMsgChan = checker.CheckForUpdateAsync(cmd.Context())
		}
	}
}

func Execute() {
	rootCmd, _ := newCmdRoot()

	devCmd := NewCmdDev()
	rootCmd.AddCommand(devCmd)
	devCmd.AddCommand(NewCmdDevDocSupportActions())
	devCmd.AddCommand(NewCmdDevDoctor())
	devCmd.AddCommand(NewCmdDevListUnmappedActions())
	devCmd.AddCommand(NewCmdDevGenerateBase())
	devCmd.AddCommand(NewCmdDevLookup())
	devCmd.AddCommand(NewCmdDevMapping())
	rootCmd.AddCommand(NewCmdView())
	rootCmd.AddCommand(NewCmdMigrate())
	rootCmd.AddCommand(NewCmdImport())
	rootCmd.AddCommand(NewCmdExport())
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func rootPersistentPostRun(
	_ *rootFlags,
	dependencies func() (*slog.Logger, metrics.Recorder),
) func(cmd *cobra.Command, _ []string) {
	return func(cmd *cobra.Command, _ []string) {
		logger, recorder := dependencies()
		ctx := cmd.Context()

		if err := recorder.Shutdown(ctx); err != nil {
			logger.ErrorContext(ctx, "failed to shutdown telemetry", "error", err)
		}

		// Try to get update message from async check (non-blocking)
		if cmdUpdateMsgChan != nil {
			select {
			case msg := <-cmdUpdateMsgChan:
				if msg != "" {
					cmd.Print(msg)
				}
			default:
				logger.Debug("update check not completed, skipping notification")
			}
		}
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
