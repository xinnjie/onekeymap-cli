package cmd

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/xinnjie/onekeymap-cli/internal/views"
	"github.com/xinnjie/onekeymap-cli/pkg/api/exporterapi"
	"github.com/xinnjie/onekeymap-cli/pkg/api/importerapi"
	"github.com/xinnjie/onekeymap-cli/pkg/api/pluginapi"
	"github.com/xinnjie/onekeymap-cli/pkg/registry"
)

type migrateFlags struct {
	from        string
	to          string
	input       string
	output      string
	interactive bool
	backup      bool
}

func NewCmdMigrate() *cobra.Command {
	f := migrateFlags{}
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate keymaps from one editor to another",
		RunE: migrateRun(&f, func() (*slog.Logger, importerapi.Importer, exporterapi.Exporter, *registry.Registry) {
			return cmdLogger, cmdImportService, cmdExportService, cmdPluginRegistry
		}),
		Args: cobra.ExactArgs(0),
	}

	cmd.Flags().StringVar(&f.from, "from", "", "Source editor, valid values: vscode, zed")
	cmd.Flags().StringVar(&f.to, "to", "", "Target editor, valid values: vscode, zed")
	cmd.Flags().StringVar(&f.input, "input", "", "Path to source editor config")
	cmd.Flags().StringVar(&f.output, "output", "", "Path to target editor config")
	cmd.Flags().BoolVar(&f.interactive, "interactive", true, "Run in interactive mode")
	cmd.Flags().BoolVar(&f.backup, "backup", false, "Create a backup of the target editor's keymap")

	return cmd
}

func migrateRun(
	f *migrateFlags,
	dependencies func() (*slog.Logger, importerapi.Importer, exporterapi.Exporter, *registry.Registry),
) func(cmd *cobra.Command, _ []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		logger, importService, exportService, pluginRegistry := dependencies()

		if f.interactive {
			// In interactive mode, we can use the form to get missing values.
			if f.from == "" || f.to == "" {
				model := views.NewMigrateFormModel(pluginRegistry, &f.from, &f.to, &f.input, &f.output)
				p := tea.NewProgram(model)
				if _, err := p.Run(); err != nil {
					logger.Error("failed to run interactive form", "error", err)
					return err
				}
			}
		}

		// After potentially running the form, `from` and `to` must be set.
		if f.from == "" || f.to == "" {
			return errors.New("required flags 'from' and 'to' not set")
		}

		logger.Info("Migrating keymaps", "from", f.from, "to", f.to)

		inputPlugin, ok := pluginRegistry.Get(pluginapi.EditorType(f.from))
		if !ok {
			logger.Error("failed to get input plugin", "from", f.from)
			return errors.New("failed to get input plugin")
		}
		outputPlugin, ok := pluginRegistry.Get(pluginapi.EditorType(f.to))
		if !ok {
			logger.Error("failed to get output plugin", "to", f.to)
			return errors.New("failed to get output plugin")
		}

		if f.input == "" {
			v, _, err := inputPlugin.ConfigDetect(pluginapi.ConfigDetectOptions{})
			if err != nil {
				logger.Error("failed to get default config path", "error", err)
				return err
			}
			f.input = v[0]
		}

		if f.output == "" {
			v, _, err := outputPlugin.ConfigDetect(pluginapi.ConfigDetectOptions{})
			if err != nil {
				logger.Error("failed to get default config path", "error", err)
				return err
			}
			f.output = v[0]
		}

		inputStream, err := os.Open(f.input)
		if err != nil {
			logger.Error("failed to open input file", "error", err)
			return err
		}
		defer func() { _ = inputStream.Close() }()

		importOpts := importerapi.ImportOptions{
			EditorType:  pluginapi.EditorType(f.from),
			InputStream: inputStream,
		}
		importResult, err := importService.Import(ctx, importOpts)
		if err != nil {
			logger.Error("migrate failed during import step", "error", err)
			return err
		}

		logger.Info("Import complete.")
		if importResult != nil {
			logger.Debug("Import Report", "report", importResult.Report)
		}

		if len(importResult.Setting.Actions) == 0 {
			logger.Warn("No imported keymaps to export; aborting migrate")
			return nil
		}

		// Prepare base reader from existing output file if present for diff calculation
		var base io.Reader
		if file, err := os.Open(f.output); err == nil {
			defer func() { _ = file.Close() }()
			base = file
		} else if !os.IsNotExist(err) {
			logger.Warn("Failed to open existing output as base", "error", err)
		}

		// Export to memory buffer first for preview, optional confirmation, and then write
		var mem bytes.Buffer
		exportOpts := exporterapi.ExportOptions{EditorType: pluginapi.EditorType(f.to), OriginalConfig: base}
		exportReport, err := exportService.Export(ctx, &mem, importResult.Setting, exportOpts)
		if err != nil {
			logger.Error("migrate failed during export step", "error", err)
			return err
		}

		// Show diff preview
		cmd.Println("================ Migrate Diff Preview ================")
		if exportReport != nil && strings.TrimSpace(exportReport.Diff) != "" {
			cmd.Println(exportReport.Diff)
		} else {
			cmd.Println("(no changes)")
		}
		cmd.Println("=====================================================")

		// Confirm before writing only when interactive
		if f.interactive {
			if !confirm(cmd, f.output) {
				logger.Info("Migration canceled; no changes were written.")
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
		outputStream, err := os.OpenFile(f.output, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
		if err != nil {
			logger.Error("failed to create output file", "error", err)
			return err
		}
		defer func() { _ = outputStream.Close() }()
		if _, err := outputStream.Write(mem.Bytes()); err != nil {
			logger.Error("failed to write to output file", "error", err)
			return err
		}

		logger.Info("Migration complete!")

		return nil
	}
}
