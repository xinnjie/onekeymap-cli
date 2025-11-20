package cmd

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xinnjie/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/onekeymap-cli/internal/platform"
	"github.com/xinnjie/onekeymap-cli/internal/plugins"
	"github.com/xinnjie/onekeymap-cli/internal/views"
	"github.com/xinnjie/onekeymap-cli/pkg/api/importerapi"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
	keymapv1 "github.com/xinnjie/onekeymap-cli/protogen/keymap/v1"
)

type importFlags struct {
	from        string
	input       string
	output      string
	interactive bool
	backup      bool
}

//nolint:dupl // Import/Export command constructors are intentionally symmetrical; limited duplication keeps each isolated and clearer
func NewCmdImport() *cobra.Command {
	f := importFlags{}
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import an editor's keymap to the universal format",
		RunE: importRun(&f, func() (*slog.Logger, *plugins.Registry, importerapi.Importer) {
			return cmdLogger, cmdPluginRegistry, cmdImportService
		}),
		Args: cobra.ExactArgs(0),
	}

	cmd.Flags().StringVar(&f.from, "from", "", "Source editor to import from (e.g., vscode, zed)")
	cmd.Flags().StringVar(&f.output, "output", "", "Path to save the generated onekeymap.json file")
	cmd.Flags().
		StringVar(&f.input, "input", "", "Optional: Path to the source editor's config file (overrides env vars)")
	cmd.Flags().BoolVar(&f.interactive, "interactive", true, "Run in interactive mode")
	cmd.Flags().BoolVar(&f.backup, "backup", false, "Create a backup of the target editor's keymap")

	// Add completion for 'from' flag
	_ = cmd.RegisterFlagCompletionFunc(
		"from",
		func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return cmdPluginRegistry.GetNames(), cobra.ShellCompDirectiveNoFileComp
		},
	)

	return cmd
}

func importRun(
	f *importFlags,
	dependencies func() (*slog.Logger, *plugins.Registry, importerapi.Importer),
) func(cmd *cobra.Command, _ []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		logger, pluginRegistry, importService := dependencies()
		onekeymapConfig := viper.GetString("onekeymap")

		if f.interactive {
			return importRunInteractive(cmd, f, logger, pluginRegistry, importService, onekeymapConfig)
		}

		return importRunNonInteractive(cmd, f, logger, pluginRegistry, importService, onekeymapConfig)
	}
}

func importRunInteractive(
	cmd *cobra.Command,
	f *importFlags,
	logger *slog.Logger,
	pluginRegistry *plugins.Registry,
	importService importerapi.Importer,
	onekeymapConfig string,
) error {
	if err := prepareInteractiveImportFlags(cmd, f, onekeymapConfig, pluginRegistry, logger); err != nil {
		return err
	}

	return executeImportInteractive(cmd, f, logger, importService, onekeymapConfig)
}

func importRunNonInteractive(
	cmd *cobra.Command,
	f *importFlags,
	logger *slog.Logger,
	pluginRegistry *plugins.Registry,
	importService importerapi.Importer,
	onekeymapConfig string,
) error {
	if err := prepareNonInteractiveImportFlags(f, onekeymapConfig, pluginRegistry, logger); err != nil {
		return err
	}

	return executeImportNonInteractive(cmd, f, logger, importService, onekeymapConfig)
}

func executeImportInteractive(
	cmd *cobra.Command,
	f *importFlags,
	logger *slog.Logger,
	importService importerapi.Importer,
	onekeymapConfig string,
) error {
	var (
		file io.ReadCloser
		err  error
	)

	if f.input != "" {
		if f.from == "basekeymap" {
			// For basekeymap, f.input is the base keymap name, not a file path
			file = io.NopCloser(strings.NewReader(f.input))
		} else {
			file, err = os.Open(f.input)
			if err != nil {
				logger.Error("Failed to open input file", "path", f.input, "error", err)
				return err
			}
			defer func() { _ = file.Close() }()
		}
	}

	baseConfig := loadBaseConfig(f.output, onekeymapConfig, logger)

	opts := importerapi.ImportOptions{
		EditorType:  pluginapi.EditorType(f.from),
		InputStream: file,
		Base:        baseConfig,
	}

	result, err := importService.Import(cmd.Context(), opts)
	if err != nil {
		logger.Error("import failed", "error", err)
		return err
	}

	if result != nil && len(result.SkipReport.SkipActions) > 0 {
		vm := views.NewImportSkipReportViewModel(result.SkipReport)
		p := tea.NewProgram(vm)
		if _, err := p.Run(); err != nil {
			logger.Error("could not start program", "error", err)
		}
	}

	// Show validation report if there are issues
	if result != nil && result.Report != nil &&
		(len(result.Report.GetIssues()) > 0 || len(result.Report.GetWarnings()) > 0) {
		logger.Info("Validation found issues. Displaying report...")
		if err := runValidationReportPreview(result.Report); err != nil {
			logger.Warn("Failed to display validation report", "error", err)
		}
	}

	// Show changes preview and get user confirmation
	if result != nil && result.Changes != nil {
		confirmed, err := runImportChangesPreview(result.Changes)
		if err != nil {
			logger.Warn("failed to render changes preview", "error", err)
		}
		if !confirmed {
			logger.Info("User cancelled applying changes; no file will be written")
			return nil
		}
	}

	return saveImportResult(f.output, result, logger)
}

func executeImportNonInteractive(
	cmd *cobra.Command,
	f *importFlags,
	logger *slog.Logger,
	importService importerapi.Importer,
	onekeymapConfig string,
) error {
	var (
		file io.ReadCloser
		err  error
	)

	if f.input != "" {
		if f.from == "basekeymap" {
			// For basekeymap, f.input is the base keymap name, not a file path
			file = io.NopCloser(strings.NewReader(f.input))
		} else {
			file, err = os.Open(f.input)
			if err != nil {
				logger.Error("Failed to open input file", "path", f.input, "error", err)
				return err
			}
			defer func() { _ = file.Close() }()
		}
	}

	baseConfig := loadBaseConfig(f.output, onekeymapConfig, logger)

	opts := importerapi.ImportOptions{
		EditorType:  pluginapi.EditorType(f.from),
		InputStream: file,
		Base:        baseConfig,
	}

	result, err := importService.Import(cmd.Context(), opts)
	if err != nil {
		logger.Error("import failed", "error", err)
		return err
	}

	return saveImportResult(f.output, result, logger)
}

func loadBaseConfig(outputPath, onekeymapConfig string, logger *slog.Logger) *keymapv1.Keymap {
	basePath := outputPath
	if basePath == "" {
		basePath = onekeymapConfig
	}
	if basePath == "" {
		return nil
	}

	baseConfigFile, err := os.Open(basePath)
	if err != nil {
		logger.Debug("Base config file not found, skip loading base config", "path", basePath)
		return nil
	}
	defer func() { _ = baseConfigFile.Close() }()

	cfg, lerr := keymap.Load(baseConfigFile, keymap.LoadOptions{})
	if lerr != nil {
		logger.Warn("Failed to load base keymap, treat as no base config", "error", lerr)
		return nil
	}

	return cfg
}

func saveImportResult(outputPath string, result *importerapi.ImportResult, logger *slog.Logger) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o750); err != nil {
		logger.Error("Failed to create output directory", "dir", filepath.Dir(outputPath), "error", err)
		return err
	}
	outputFile, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		logger.Error("Failed to create output file", "error", err)
		return err
	}
	defer func() {
		_ = outputFile.Close()
	}()

	if result == nil || result.Setting == nil {
		logger.Warn("No keymaps imported; nothing to save")
		return nil
	}

	saveOpt := keymap.SaveOptions{Platform: platform.PlatformMacOS}
	if err := keymap.Save(outputFile, result.Setting, saveOpt); err != nil {
		logger.Error("Failed to save config file", "error", err)
		return err
	}

	logger.Info("Successfully imported keymap", "output", outputPath)
	if result.Report != nil {
		logger.Debug("Import report", "report", result.Report)
	}
	return nil
}

func handleInteractiveImportFlags(
	cmd *cobra.Command,
	f *importFlags,
	onekeymapConfig string,
	pluginRegistry *plugins.Registry,
	logger *slog.Logger,
) error {
	needSelectEditor := !cmd.Flags().Changed("from") || f.from == ""
	needInput := !cmd.Flags().Changed("input") || f.input == ""
	needOutput := !cmd.Flags().Changed("output") || f.output == ""

	if needSelectEditor || needInput || needOutput {
		if err := runImportForm(
			pluginRegistry,
			&f.from,
			&f.input,
			&f.output,
			onekeymapConfig,
			needSelectEditor,
			needInput,
			needOutput,
		); err != nil {
			return err
		}
	}

	logger.DebugContext(cmd.Context(), "after input propmt", "from", f.from, "input", f.input, "output", f.output)
	return nil
}

func prepareInteractiveImportFlags(
	cmd *cobra.Command,
	f *importFlags,
	onekeymapConfig string,
	pluginRegistry *plugins.Registry,
	logger *slog.Logger,
) error {
	if err := handleInteractiveImportFlags(cmd, f, onekeymapConfig, pluginRegistry, logger); err != nil {
		return err
	}

	p, ok := pluginRegistry.Get(pluginapi.EditorType(f.from))
	if !ok {
		logger.Error("Editor not found", "editor", f.from)
		return fmt.Errorf("editor %s not found", f.from)
	}

	if f.input == "" {
		configPath := viper.GetString(fmt.Sprintf("editors.%s.keymap_path", f.from))
		if configPath != "" {
			f.input = configPath
			logger.Info("Using keymap path from config", "editor", f.from, "path", configPath)
		} else {
			v, _, err := p.ConfigDetect(pluginapi.ConfigDetectOptions{})
			if err != nil {
				logger.Error("Failed to get default config path", "error", err)
				return err
			}
			f.input = v[0]
		}
	}
	return nil
}

func prepareNonInteractiveImportFlags(
	f *importFlags,
	onekeymapConfig string,
	pluginRegistry *plugins.Registry,
	logger *slog.Logger,
) error {
	if f.from == "" {
		return errors.New("flag --from is required")
	}
	if onekeymapConfig != "" {
		f.output = onekeymapConfig
	}

	p, ok := pluginRegistry.Get(pluginapi.EditorType(f.from))
	if !ok {
		logger.Error("Editor not found", "editor", f.from)
		return fmt.Errorf("editor %s not found", f.from)
	}

	if f.input == "" {
		configPath := viper.GetString(fmt.Sprintf("editors.%s.keymap_path", f.from))
		if configPath != "" {
			f.input = configPath
			logger.Info("Using keymap path from config", "editor", f.from, "path", configPath)
		} else {
			v, _, err := p.ConfigDetect(pluginapi.ConfigDetectOptions{})
			if err != nil {
				logger.Error("Failed to get default config path", "error", err)
				return err
			}
			f.input = v[0]
		}
	}
	return nil
}

func runImportChangesPreview(changes *importerapi.KeymapChanges) (bool, error) {
	confirmed := true
	m := views.NewKeymapChangesModel(changes, &confirmed)
	_, err := tea.NewProgram(m).Run()
	return confirmed, err
}

// runImportForm runs the interactive import form and returns the selected values.
// All TUI logic is encapsulated here to keep cmd/import.go simple.
func runImportForm(
	pluginRegistry *plugins.Registry,
	from, input, output *string,
	onekeymapConfigPlaceHolder string,
	needSelectEditor, needInput, needOutput bool,
) error {
	m, err := views.NewImportFormModel(
		pluginRegistry,
		needSelectEditor,
		needInput,
		needOutput,
		from,
		input,
		output,
		onekeymapConfigPlaceHolder,
	)
	if err != nil {
		return err
	}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		if errors.Is(err, tea.ErrProgramKilled) {
			os.Exit(0)
		}
		return err
	}
	return nil
}

// run the validation report TUI ---.
func runValidationReportPreview(report *keymapv1.ValidationReport) error {
	m := views.NewValidationReportModel(report)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
