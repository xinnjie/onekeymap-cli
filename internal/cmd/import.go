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

var (
	importFrom   *string
	importInput  *string
	importOutput *string
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import an editor's keymap to the universal format",
	RunE: func(cmd *cobra.Command, args []string) error {
		onekeymapConfig := viper.GetString("onekeymap")
		err := prepareImportInputFlags(cmd, onekeymapConfig)
		if err != nil {
			return err
		}

		var f *os.File
		if *importInput != "" {
			f, err = os.Open(*importInput)
			if err != nil {
				logger.Error("Failed to open input file", "path", *importInput, "error", err)
				return err
			}
			defer func() { _ = f.Close() }()
		}

		baseConfig := func() *keymapv1.KeymapSetting {
			if *importOutput != "" {
				baseConfigFile, err := os.Open(*importOutput)
				if err != nil {
					logger.Debug("Base config file not found, skip loading base config", "path", *importOutput)
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
			EditorType:  pluginapi.EditorType(*importFrom),
			InputStream: f,
			Base:        baseConfig,
		}

		result, err := importService.Import(cmd.Context(), opts)
		if err != nil {
			logger.Error("import failed", "error", err)
			return err
		}
		if *interactive {
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
		if *interactive && result != nil && result.Changes != nil {
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
		if err := os.MkdirAll(filepath.Dir(*importOutput), 0o750); err != nil {
			logger.Error("Failed to create output directory", "dir", filepath.Dir(*importOutput), "error", err)
			return err
		}
		file, err := os.OpenFile(*importOutput, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
		if err != nil {
			logger.Error("Failed to create output file", "error", err)
			return err
		}
		defer func() {
			_ = file.Close()
		}()

		if result == nil || result.Setting == nil {
			logger.Warn("No keymaps imported; nothing to save")
			return nil
		}

		if err := keymap.Save(file, result.Setting); err != nil {
			logger.Error("Failed to save config file", "error", err)
			return err
		}

		logger.Info("Successfully imported keymap", "output", *importOutput)
		if result != nil && result.Report != nil {
			logger.Debug("Import report", "report", result.Report)
		}
		return nil
	},
}

func prepareImportInputFlags(cmd *cobra.Command, onekeymapConfig string) error {
	if *interactive {
		needSelectEditor := !cmd.Flags().Changed("from") || *importFrom == ""
		needInput := !cmd.Flags().Changed("input") || *importInput == ""
		needOutput := !cmd.Flags().Changed("output") || *importOutput == ""

		if needSelectEditor || needInput || needOutput {
			if err := runImportForm(pluginRegistry, importFrom, importInput, importOutput, onekeymapConfig, needSelectEditor, needInput, needOutput); err != nil {
				return err
			}
		}
	} else {
		if *importFrom == "" {
			return errors.New("flag --from is required")
		}
		if onekeymapConfig != "" {
			*importOutput = onekeymapConfig
		}
	}

	p, ok := pluginRegistry.Get(pluginapi.EditorType(*importFrom))
	if !ok {
		logger.Error("Editor not found", "editor", *importFrom)
		return fmt.Errorf("editor %s not found", *importFrom)
	}

	if *importInput == "" {
		configPath := viper.GetString(fmt.Sprintf("editors.%s.keymap_path", *importFrom))
		if configPath != "" {
			*importInput = configPath
			logger.Info("Using keymap path from config", "editor", *importFrom, "path", configPath)
		} else {
			v, err := p.DefaultConfigPath()
			if err != nil {
				logger.Error("Failed to get default config path", "error", err)
				return err
			}
			*importInput = v[0]
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

func init() {
	rootCmd.AddCommand(importCmd)
	importFrom = importCmd.Flags().String("from", "", "Source editor to import from (e.g., vscode, zed)")
	importOutput = importCmd.Flags().String("output", "", "Path to save the generated onekeymap.json file")
	importInput = importCmd.Flags().
		String("input", "", "Optional: Path to the source editor's config file (overrides env vars)")

	// Add completion for 'from' flag
	_ = importCmd.RegisterFlagCompletionFunc(
		"from",
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return pluginRegistry.GetNames(), cobra.ShellCompDirectiveNoFileComp
		},
	)
}
