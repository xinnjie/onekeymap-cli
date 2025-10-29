package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xinnjie/onekeymap-cli/internal"
	"github.com/xinnjie/onekeymap-cli/internal/cliconfig"
	"github.com/xinnjie/onekeymap-cli/internal/mappings"
	"github.com/xinnjie/onekeymap-cli/internal/plugins"
	"github.com/xinnjie/onekeymap-cli/internal/plugins/basekeymap"
	"github.com/xinnjie/onekeymap-cli/internal/plugins/helix"
	"github.com/xinnjie/onekeymap-cli/internal/plugins/intellij"
	"github.com/xinnjie/onekeymap-cli/internal/plugins/vscode"
	"github.com/xinnjie/onekeymap-cli/internal/plugins/zed"
	"github.com/xinnjie/onekeymap-cli/internal/updatecheck"
	"github.com/xinnjie/onekeymap-cli/internal/views"

	"github.com/xinnjie/onekeymap-cli/internal/metrics"
	"github.com/xinnjie/onekeymap-cli/pkg/exportapi"
	"github.com/xinnjie/onekeymap-cli/pkg/importapi"
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
	cmdUpdateMsgChan  <-chan string // Channel for async update check result
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
	onekeymap       string
}

func newCmdRoot() (*cobra.Command, *rootFlags) {
	f := rootFlags{}

	cmd := &cobra.Command{
		Use:              "onekeymap-cli",
		Short:            "A tool to import, export, and synchronize keyboard shortcuts between editors.",
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
		BoolVar(&f.enableTelemetry, "telemetry", false, "Enable telemetry to help improve onekeymap-cli")
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
	if err := viper.BindPFlag("telemetry.enabled", cmd.PersistentFlags().Lookup("telemetry")); err != nil {
		cmd.PrintErrf("Error binding telemetry flag: %v\n", err)
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

		// Show telemetry prompt in interactive mode if telemetry is not explicitly configured
		// and not explicitly enabled via flag
		if f.interactive && !cliconfig.IsTelemetryExplicitlySet() && !f.enableTelemetry {
			if err := showTelemetryPrompt(); err != nil {
				cmd.PrintErrf("Error showing telemetry prompt: %v\n", err)
				// Continue execution even if prompt fails
			}
		}

		verbose := f.verbose
		quiet := f.quiet
		logJSON := f.logJSON
		telemetryEnabled := viper.GetBool("telemetry.enabled")
		telemetryEndpoint := viper.GetString("telemetry.endpoint")
		telemetryHeaders := viper.GetString("telemetry.headers")
		ctx := cmd.Context()

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
		logger := cmdLogger

		cmdRecorder = metrics.NewNoop()
		if telemetryEnabled {
			logger.DebugContext(ctx, "Telemetry enabled")
			if telemetryEndpoint == "" {
				logger.WarnContext(
					ctx,
					"telemetry is enabled, but telemetry.endpoint is not set. Telemetry data will not be sent.",
				)
			}

			headers := parseHeaders(telemetryHeaders)

			cmdRecorder, err = metrics.New(ctx, version, logger, metrics.RecorderOption{
				Endpoint: telemetryEndpoint,
				Headers:  headers,
				UseDelta: true, // Use delta temporality for short-lived CLI application
			})
			if err != nil {
				logger.WarnContext(ctx, "failed to initialize telemetry", "error", err)
				os.Exit(1)
			}
		}

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
		cmdPluginRegistry.Register(intellij.NewRustRover(cmdMappingConfig, cmdLogger))

		cmdPluginRegistry.Register(helix.New(cmdMappingConfig, cmdLogger))
		cmdPluginRegistry.Register(zed.New(cmdMappingConfig, cmdLogger))

		cmdPluginRegistry.Register(basekeymap.New())

		cmdImportService = internal.NewImportService(cmdPluginRegistry, cmdMappingConfig, cmdLogger, cmdRecorder)
		cmdExportService = internal.NewExportService(cmdPluginRegistry, cmdMappingConfig, cmdLogger, cmdRecorder)

		// Start async update check only in interactive mode and when not in sandbox
		if f.interactive && !f.sandbox && !f.skipUpdateCheck {
			checker := updatecheck.New(version, cmdLogger)
			cmdUpdateMsgChan = checker.CheckForUpdateAsync(cmd.Context())
		}
	}
}

func Execute() {
	rootCmd, rootFlags := newCmdRoot()

	devCmd := NewCmdDev()
	rootCmd.AddCommand(devCmd)
	devCmd.AddCommand(NewCmdDevDocSupportActions())
	devCmd.AddCommand(NewCmdDevDoctor())
	devCmd.AddCommand(NewCmdDevListUnmappedActions())
	devCmd.AddCommand(NewCmdDevGenerateBase())
	rootCmd.AddCommand(NewCmdView())
	rootCmd.AddCommand(NewCmdServe(rootFlags))
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

const (
	headerKeyValueParts = 2
)

// parseHeaders parses a comma-separated string of key=value pairs into a map.
// Format: "key1=value1,key2=value2"
func parseHeaders(headersStr string) map[string]string {
	if headersStr == "" {
		return nil
	}

	headers := make(map[string]string)
	pairs := strings.Split(headersStr, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.SplitN(pair, "=", headerKeyValueParts)
		if len(parts) == headerKeyValueParts {
			headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return headers
}

// showTelemetryPrompt displays the telemetry consent prompt and updates config based on user choice
func showTelemetryPrompt() error {
	model := views.NewTelemetryPrompt()
	program := tea.NewProgram(model)

	finalModel, err := program.Run()
	if err != nil {
		return fmt.Errorf("failed to run telemetry prompt: %w", err)
	}

	// Cast back to our model type
	m, ok := finalModel.(views.TelemetryPromptModel)
	if !ok {
		return errors.New("unexpected model type")
	}

	// Only update config if user made a selection (didn't quit)
	if m.WasSelected() {
		enabled := m.GetChoice()
		if err := cliconfig.UpdateTelemetrySettings(enabled); err != nil {
			return fmt.Errorf("failed to update telemetry settings: %w", err)
		}

		// Update viper in-memory values for current session
		viper.Set("telemetry.enabled", enabled)
	}

	return nil
}
