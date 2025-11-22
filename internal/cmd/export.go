package cmd

import (
	"bytes"
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
	"github.com/xinnjie/onekeymap-cli/internal/plugins"
	"github.com/xinnjie/onekeymap-cli/internal/views"
	"github.com/xinnjie/onekeymap-cli/pkg/api/exporterapi"
	"github.com/xinnjie/onekeymap-cli/pkg/api/keymap"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
)

type exportFlags struct {
	to          string
	input       string
	output      string
	interactive bool
	backup      bool
}

//nolint:dupl // Import/Export command constructors are intentionally symmetrical; limited duplication keeps each isolated and clearer
func NewCmdExport() *cobra.Command {
	f := exportFlags{}
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export a universal keymap to an editor's format",
		RunE: exportRun(&f, func() (*slog.Logger, *plugins.Registry, exporterapi.Exporter) {
			return cmdLogger, cmdPluginRegistry, cmdExportService
		}),
		Args: cobra.ExactArgs(0),
	}

	cmd.Flags().StringVar(&f.to, "to", "", "Target editor to export to")
	cmd.Flags().StringVar(&f.input, "input", "", "Path to the source onekeymap.json file")
	cmd.Flags().StringVar(&f.output, "output", "", "Optional: Path to the target editor's config file")
	cmd.Flags().BoolVar(&f.interactive, "interactive", true, "Run in interactive mode")
	cmd.Flags().BoolVar(&f.backup, "backup", false, "Create a backup of the target editor's keymap")

	// Add completion for 'to' flag
	_ = cmd.RegisterFlagCompletionFunc(
		"to",
		func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return cmdPluginRegistry.GetNames(), cobra.ShellCompDirectiveNoFileComp
		},
	)

	return cmd
}

func exportRun(
	f *exportFlags,
	dependencies func() (*slog.Logger, *plugins.Registry, exporterapi.Exporter),
) func(cmd *cobra.Command, _ []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		logger, pluginRegistry, exportService := dependencies()
		onekeymapPlaceHolder := viper.GetString("onekeymap")
		err := prepareExportInputFlags(cmd, f, onekeymapPlaceHolder, pluginRegistry, logger)
		if err != nil {
			return err
		}
		logger.Info("Exporting config", "to", f.to, "from", f.input)

		inputFile, err := os.Open(f.input)
		if err != nil {
			logger.Error("Failed to open input file", "error", err)
			return err
		}
		defer func() {
			if err := inputFile.Close(); err != nil {
				logger.Error("Failed to close input file", "error", err)
			}
		}()

		setting, err := keymap.Load(inputFile, keymap.LoadOptions{})
		if err != nil {
			logger.Error("Failed to load config file", "error", err)
			return err
		}

		opts := exporterapi.ExportOptions{EditorType: pluginapi.EditorType(f.to)}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(f.output), 0o750); err != nil {
			logger.Error("Failed to create output directory", "dir", filepath.Dir(f.output), "error", err)
			return err
		}

		// Prepare base reader from existing output file if present for diff calculation
		var base io.Reader
		if outputFile, err := os.Open(f.output); err == nil {
			defer func() { _ = outputFile.Close() }()
			base = outputFile
		} else if !os.IsNotExist(err) {
			// Non-ENOENT error opening base file; log and continue without base
			logger.Warn("Failed to open existing output as base", "error", err)
		}
		opts.Base = base
		opts.DiffType = exporterapi.DiffTypeASCII
		opts.FilePath = f.output

		// Export to memory buffer first for preview, optional confirmation, and then write
		var mem bytes.Buffer
		report, err := exportService.Export(cmd.Context(), &mem, setting, opts)
		if err != nil {
			logger.Error("export failed", "error", err)
			return err
		}

		// Show skipped actions in interactive mode for visibility
		if f.interactive && report != nil && len(report.SkipActions) > 0 {
			vm := views.NewSkipReportViewModel(report.SkipActions)
			p := tea.NewProgram(vm)
			if _, err := p.Run(); err != nil {
				logger.Error("could not start program", "error", err)
			}
		}

		// Show diff preview
		cmd.Println("================ Export Diff Preview ================")
		if report != nil && strings.TrimSpace(report.Diff) != "" {
			cmd.Println(report.Diff)
		} else {
			cmd.Println("(no diff available)")
		}
		cmd.Println("=====================================================")

		// TODO(xinnjie): optimize skip writing when output is the same. Whether same or not should not rely on report.Diff

		// Confirm before writing only when interactive
		if f.interactive {
			if !confirm(cmd, f.output) {
				cmd.Println("Export canceled; no changes were written.")
				return nil
			}
		}

		// Backup existing file if requested
		if f.backup {
			if backupPath, err := backupIfExists(f.output); err != nil {
				logger.Warn("Failed to backup existing file", "path", f.output, "error", err)
			} else if backupPath != "" {
				logger.Info("Created backup of existing config", "backup", backupPath)
			}
		}

		// Write buffer to the target file
		outputFile, err := os.OpenFile(f.output, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
		if err != nil {
			logger.Error("Failed to create output file", "error", err)
			return err
		}
		defer func() {
			if err := outputFile.Close(); err != nil {
				logger.Error("Failed to close output file", "error", err)
			}
		}()
		if _, err := outputFile.Write(mem.Bytes()); err != nil {
			logger.Error("Failed to write to output file", "error", err)
			return err
		}

		logger.Info("Successfully exported keymap", "to", f.to, "output", f.output)
		return nil
	}
}

func handleInteractiveExportFlags(
	cmd *cobra.Command,
	f *exportFlags,
	onekeymapPlaceholder string,
	pluginRegistry *plugins.Registry,
) error {
	needSelectEditor := !cmd.Flags().Changed("to") || f.to == ""
	needInput := !cmd.Flags().Changed("input") || f.input == ""
	needOutput := !cmd.Flags().Changed("output") || f.output == ""

	if needSelectEditor || needInput || needOutput {
		if err := runExportForm(pluginRegistry, &f.to, &f.input, &f.output, onekeymapPlaceholder, needSelectEditor, needInput, needOutput); err != nil {
			return err
		}
	}
	return nil
}

func handleNonInteractiveExportFlags(
	cmd *cobra.Command,
	f *exportFlags,
	onekeymapPlaceholder string,
) error {
	if f.to == "" {
		return errors.New("flag --to is required")
	}
	if !cmd.Flags().Changed("input") && onekeymapPlaceholder != "" {
		f.input = onekeymapPlaceholder
	}
	return nil
}

func prepareExportInputFlags(
	cmd *cobra.Command,
	f *exportFlags,
	onekeymapPlaceholder string,
	pluginRegistry *plugins.Registry,
	logger *slog.Logger,
) error {
	if f.interactive {
		if err := handleInteractiveExportFlags(cmd, f, onekeymapPlaceholder, pluginRegistry); err != nil {
			return err
		}
	} else {
		if err := handleNonInteractiveExportFlags(cmd, f, onekeymapPlaceholder); err != nil {
			return err
		}
	}

	// Validate selected editor plugin exists
	p, ok := pluginRegistry.Get(pluginapi.EditorType(f.to))
	if !ok {
		logger.Error("Editor not found", "editor", f.to)
		return fmt.Errorf("editor %s not found", f.to)
	}

	// Determine input
	if f.input == "" {
		// Fallback to config default in case it wasn't provided
		f.input = onekeymapPlaceholder
	}
	// Determine output path
	if f.output == "" {
		configPath := viper.GetString(fmt.Sprintf("editors.%s.keymap_path", f.to))
		if configPath != "" {
			f.output = configPath
			logger.Info("Using keymap path from config", "editor", f.to, "path", configPath)
		} else {
			if v, _, err := p.ConfigDetect(pluginapi.ConfigDetectOptions{}); err == nil {
				f.output = v[0]
			}
		}
	}
	return nil
}

func runExportForm(pluginRegistry *plugins.Registry, to, input, output *string, onekeymapConfigPlaceHolder string,
	needSelectEditor, needInput, needOutput bool) error {
	m, err := views.NewOutputFormModel(
		pluginRegistry,
		needSelectEditor,
		needInput,
		needOutput,
		to,
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
