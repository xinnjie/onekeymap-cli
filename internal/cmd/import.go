package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/keymap"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/plugins"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/internal/views"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/importapi"
	"github.com/xinnjie/watchbeats/onekeymap/onekeymap-cli/pkg/pluginapi"
	keymapv1 "github.com/xinnjie/watchbeats/protogen/keymap/v1"
)

type importFlags struct {
	from        string
	input       string
	output      string
	interactive bool
	backup      bool
}

// NewCmdImport is reported as duplicate of NewCmdExport, but it is necessary duplication
// nolint:dupl
func NewCmdImport() *cobra.Command {
	f := importFlags{}
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import an editor's keymap to the universal format",
		RunE:  importRun(&f),
		Args:  cobra.ExactArgs(0),
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
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return pluginRegistry.GetNames(), cobra.ShellCompDirectiveNoFileComp
		},
	)

	return cmd
}

func importRun(f *importFlags) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		onekeymapConfig := viper.GetString("onekeymap")
		err := prepareImportInputFlags(cmd, f, onekeymapConfig)
		if err != nil {
			return err
		}

		var file *os.File
		if f.input != "" {
			file, err = os.Open(f.input)
			if err != nil {
				logger.Error("Failed to open input file", "path", f.input, "error", err)
				return err
			}
			defer func() { _ = file.Close() }()
		}

		baseConfig := func() *keymapv1.KeymapSetting {
			if f.output != "" {
				baseConfigFile, err := os.Open(f.output)
				if err != nil {
					logger.Debug("Base config file not found, skip loading base config", "path", f.output)
					return nil
				}
				defer func() { _ = baseConfigFile.Close() }()
				baseConfig, err := keymap.Load(baseConfigFile)
				if err != nil {
					logger.Warn("Failed to load base keymap, treat as no base config", "error", err)
					return nil
				}
				return baseConfig
			}
			return nil
		}()

		opts := importapi.ImportOptions{
			EditorType:  pluginapi.EditorType(f.from),
			InputStream: file,
			Base:        baseConfig,
		}

		result, err := importService.Import(cmd.Context(), opts)
		if err != nil {
			logger.Error("import failed", "error", err)
			return err
		}
		if f.interactive {
			// Validation Report Display
			if result != nil && result.Report != nil &&
				(len(result.Report.GetIssues()) > 0 || len(result.Report.GetWarnings()) > 0) {
				logger.Info("Validation found issues. Displaying report...")
				if err := runValidationReportPreview(result.Report); err != nil {
					// Log the error but don't block the import process
					logger.Warn("Failed to display validation report", "error", err)
				}
			}
		}

		// If interactive, preview the calculated changes in three tables (Add/Remove/Update).
		if f.interactive && result != nil && result.Changes != nil {
			confirmed, err := runImportChangesPreview(result.Changes)
			if err != nil {
				logger.Warn("failed to render changes preview", "error", err)
			}
			if !confirmed {
				logger.Info("User cancelled applying changes; no file will be written")
				return nil
			}
		}

		// Ensure parent directory exists, then create/truncate output file
		if err := os.MkdirAll(filepath.Dir(f.output), 0o750); err != nil {
			logger.Error("Failed to create output directory", "dir", filepath.Dir(f.output), "error", err)
			return err
		}
		outputFile, err := os.OpenFile(f.output, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
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

		if err := keymap.Save(outputFile, result.Setting); err != nil {
			logger.Error("Failed to save config file", "error", err)
			return err
		}

		logger.Info("Successfully imported keymap", "output", f.output)
		if result.Report != nil {
			logger.Debug("Import report", "report", result.Report)
		}
		return nil
	}
}

func prepareImportInputFlags(cmd *cobra.Command, f *importFlags, onekeymapConfig string) error {
	if f.interactive {
		needSelectEditor := !cmd.Flags().Changed("from") || f.from == ""
		needInput := !cmd.Flags().Changed("input") || f.input == ""
		needOutput := !cmd.Flags().Changed("output") || f.output == ""

		if needSelectEditor || needInput || needOutput {
			if err := runImportForm(pluginRegistry, &f.from, &f.input, &f.output, onekeymapConfig, needSelectEditor, needInput, needOutput); err != nil {
				return err
			}
		}
	} else {
		if f.from == "" {
			return errors.New("flag --from is required")
		}
		if onekeymapConfig != "" {
			f.output = onekeymapConfig
		}
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
			v, err := p.DefaultConfigPath()
			if err != nil {
				logger.Error("Failed to get default config path", "error", err)
				return err
			}
			f.input = v[0]
		}
	}
	return nil
}

func runImportChangesPreview(changes *importapi.KeymapChanges) (bool, error) {
	confirmed := true
	m := views.NewKeymapChangesModel(changes, &confirmed)
	_, err := tea.NewProgram(m).Run()
	return confirmed, err
}

// runImportForm runs the interactive import form and returns the selected values.
// All TUI logic is encapsulated here to keep cmd/import.go simple.
func runImportForm(pluginRegistry *plugins.Registry, from, input, output *string, onekeymapConfigPlaceHolder string,
	needSelectEditor, needInput, needOutput bool) error {
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
